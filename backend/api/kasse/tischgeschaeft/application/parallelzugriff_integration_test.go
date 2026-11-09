//go:build integration

package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
)

// Two Servicekräfte order and pay at the same Tisch over several rounds, each on its own variant; OCC
// conflicts are expected and retried. Afterwards replay equals the projection, Saldo is 0, and per
// variant ordered equals paid quantity (no lost position, no double booking).
func TestParallelzugriff_ZweiClientsSelberTisch(t *testing.T) {
	ctx, cmd, db, userID, ksNr, tischID, produktID, varianteA := setupBestellungIntegration(t)

	// Zweite Variante desselben Produkts für die zweite Servicekraft.
	var varianteB int
	if err := db.QueryRow(
		"INSERT INTO produkt_varianten (produkt_id, name, preis_cents, status, created_at, updated_at) VALUES ($1, '0.3L', 250, 'active', now(), now()) RETURNING id",
		produktID,
	).Scan(&varianteB); err != nil {
		t.Fatalf("create variante B: %v", err)
	}

	const runden = 6
	const mengeProRunde = 2
	subject := kasse.TischSessionSubject(ksNr, tischID)

	clients := []struct {
		userName   string
		varianteID int
	}{
		{"Anna", varianteA},
		{"Bernd", varianteB},
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(clients))
	for _, cl := range clients {
		wg.Add(1)
		go func(userName string, varianteID int) {
			defer wg.Done()
			if err := bedieneTisch(ctx, cmd, subject, userID, userName, produktID, tischID, varianteID, runden, mengeProRunde); err != nil {
				errCh <- err
			}
		}(cl.userName, cl.varianteID)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Fatalf("Client-Fehler: %v", err)
	}

	events, err := cmd.EventRepo.ReadEventsBySubject(ctx, subject)
	if err != nil {
		t.Fatalf("ReadEventsBySubject: %v", err)
	}

	// Journal-Replay: Zustand aus den Events falten.
	replay := kasse.TischSession{Subject: subject}
	for _, ev := range events {
		replay, err = kasse.ApplyEvent(replay, ev)
		if err != nil {
			t.Fatalf("ApplyEvent (event %d, type %s): %v", ev.ID, ev.Type, err)
		}
	}

	// Projektion aus der Datenbank.
	projektion, err := cmd.EventRepo.ReadTischSession(ctx, subject)
	if err != nil {
		t.Fatalf("ReadTischSession: %v", err)
	}

	// 1) Journal-Replay muss der persistierten Projektion entsprechen.
	if replay.SaldoCents != projektion.SaldoCents {
		t.Errorf("Saldo-Drift Journal↔Projektion: replay=%d, projektion=%d", replay.SaldoCents, projektion.SaldoCents)
	}
	if replay.GesamtZahlungenCents != projektion.GesamtZahlungenCents {
		t.Errorf("Zahlungssummen-Drift Journal↔Projektion: replay=%d, projektion=%d", replay.GesamtZahlungenCents, projektion.GesamtZahlungenCents)
	}
	if len(replay.UnbezahltePositionen) != len(projektion.UnbezahltePositionen) {
		t.Errorf("Unbezahlt-Positionsanzahl-Drift Journal↔Projektion: replay=%d, projektion=%d", len(replay.UnbezahltePositionen), len(projektion.UnbezahltePositionen))
	}

	// 2) Saldo 0: jede bestellte Position wurde genau einmal bezahlt.
	if projektion.SaldoCents != 0 {
		t.Errorf("Saldo nach vollständigem Kassieren erwartet 0, ist %d (verlorene oder doppelte Buchung)", projektion.SaldoCents)
	}
	if len(projektion.UnbezahltePositionen) != 0 {
		t.Errorf("Es dürfen keine unbezahlten Positionen übrig sein, sind %d", len(projektion.UnbezahltePositionen))
	}

	// 3) Je Variante: bestellte == bezahlte Menge (aus dem Journal aggregiert).
	bestellt := map[int]int{}
	bezahlt := map[int]int{}
	for _, ev := range events {
		switch ev.Type {
		case string(kasse.EventTypeBestellungAufgenommenV1):
			addMengen(t, ev.Data, bestellt)
		case string(kasse.EventTypeZahlungKassiertV1), string(kasse.EventTypeZahlungKassiertV2):
			addMengen(t, ev.Data, bezahlt)
		}
	}

	const erwarteteMenge = runden * mengeProRunde
	for _, v := range []int{varianteA, varianteB} {
		if bestellt[v] != erwarteteMenge {
			t.Errorf("Variante %d: erwartet %d bestellte Stück, Journal zeigt %d", v, erwarteteMenge, bestellt[v])
		}
		if bezahlt[v] != bestellt[v] {
			t.Errorf("Variante %d: bestellt %d != bezahlt %d (verlorene Position oder Doppelbuchung)", v, bestellt[v], bezahlt[v])
		}
	}
}

// bedieneTisch runs runden × (order → pay) on the Servicekraft's own variant, then a final pay pass
// for positions left open by interleaving.
func bedieneTisch(ctx context.Context, cmd Command, subject string, userID int, userName string, produktID, tischID, varianteID, runden, menge int) error {
	for range runden {
		bestellungID := uuid.New().String()
		inputs := []enrichment.PositionInput{{ProduktID: produktID, VarianteID: varianteID, Menge: menge}}
		if err := retryConflict(func() error {
			return cmd.BestellungAufnehmen(ctx, userID, userName, bestellungID, tischID, inputs, "")
		}); err != nil {
			return err
		}
		if err := kassiere(ctx, cmd, subject, userID, userName, tischID, varianteID); err != nil {
			return err
		}
	}
	// Abschließender Durchlauf: Positionen, die in einer Runde wegen Interleaving
	// noch offen waren, werden hier kassiert.
	return kassiere(ctx, cmd, subject, userID, userName, tischID, varianteID)
}

// kassiere kassiert die noch unbezahlten Positionen der Variante aus dem
// aktuellen Sessionzustand.
func kassiere(ctx context.Context, cmd Command, subject string, userID int, userName string, tischID, varianteID int) error {
	return retryConflict(func() error {
		refs, err := offeneRefsFuerVariante(ctx, cmd, subject, varianteID)
		if err != nil || len(refs) == 0 {
			return err
		}
		return cmd.ZahlungKassieren(ctx, userID, userName, uuid.NewString(), tischID, refs, "", nil)
	})
}

// offeneRefsFuerVariante returns the variant's unpaid positions, so each Servicekraft pays only its
// own; the server-side Bezahl-Invariante also prevents paying a position twice.
func offeneRefsFuerVariante(ctx context.Context, cmd Command, subject string, varianteID int) ([]kasse.PositionRef, error) {
	state, err := cmd.EventRepo.ReadTischSession(ctx, subject)
	if err != nil {
		return nil, err
	}
	var refs []kasse.PositionRef
	for _, p := range state.UnbezahltePositionen {
		if p.VarianteID == varianteID {
			refs = append(refs, kasse.PositionRef{PositionID: p.PositionID, Menge: p.Menge})
		}
	}
	return refs, nil
}

// retryConflict repeats op while it returns ErrConflict, each try reading fresh state. The cap keeps a
// persistent error from looping forever.
func retryConflict(op func() error) error {
	const maxVersuche = 500
	for range maxVersuche {
		err := op()
		if err == nil {
			return nil
		}
		if errors.Is(err, ErrConflict) {
			continue
		}
		return err
	}
	return errors.New("retryConflict: zu viele OCC-Konflikte")
}

// addMengen summiert die Positionsmengen der Event-Payload je Variante auf.
func addMengen(t *testing.T, data []byte, ziel map[int]int) {
	t.Helper()
	var d struct {
		Positionen []kasse.PositionEventData `json:"positionen"`
	}
	if err := json.Unmarshal(data, &d); err != nil {
		t.Fatalf("unmarshal event positionen: %v", err)
	}
	for _, p := range d.Positionen {
		ziel[p.VarianteID] += p.Menge
	}
}

// gleichzeitig runs op once per ID behind a shared gate, so both requests read the same stream state.
func gleichzeitig(ids []string, op func(id string) error) []error {
	start := make(chan struct{})
	ergebnisse := make([]error, len(ids))
	var wg sync.WaitGroup
	for i, id := range ids {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			ergebnisse[i] = op(id)
		}()
	}
	close(start)
	wg.Wait()
	return ergebnisse
}

// genauEinGewinner fails unless exactly one request succeeded and the other was refused with one
// of the allowed errors.
func genauEinGewinner(t *testing.T, ergebnisse []error, erlaubt ...error) {
	t.Helper()
	erfolge := 0
	for _, err := range ergebnisse {
		if err == nil {
			erfolge++
			continue
		}
		if !slices.ContainsFunc(erlaubt, func(ziel error) bool { return errors.Is(err, ziel) }) {
			t.Errorf("unerwarteter Fehler des Verlierers: %v", err)
		}
	}
	if erfolge != 1 {
		t.Errorf("erwartet genau einen erfolgreichen Request, bekam %d (%v)", erfolge, ergebnisse)
	}
}

func zaehleEvents(events []event.Event, typ kasse.EventType) int {
	anzahl := 0
	for _, ev := range events {
		if ev.Type == string(typ) {
			anzahl++
		}
	}
	return anzahl
}

// bestelleEinBier books one Bier at 350 and returns its open position.
func bestelleEinBier(ctx context.Context, t *testing.T, cmd Command, subject string, userID, produktID, tischID, varianteID int) kasse.PositionRef {
	t.Helper()
	inputs := []enrichment.PositionInput{{ProduktID: produktID, VarianteID: varianteID, Menge: 1}}
	if err := cmd.BestellungAufnehmen(ctx, userID, "Anna", uuid.NewString(), tischID, inputs, ""); err != nil {
		t.Fatalf("BestellungAufnehmen: %v", err)
	}
	state, err := cmd.EventRepo.ReadTischSession(ctx, subject)
	if err != nil {
		t.Fatalf("ReadTischSession: %v", err)
	}
	if len(state.UnbezahltePositionen) != 1 {
		t.Fatalf("erwartet 1 offene Position, sind %d", len(state.UnbezahltePositionen))
	}
	return kasse.PositionRef{PositionID: state.UnbezahltePositionen[0].PositionID, Menge: 1}
}

// Two devices pay the same position at the same moment with different zahlungIds: exactly one
// payment lands, the other is refused, and the Tisch carries 350 paid and Saldo 0.
func TestParallelzugriff_GleichePositionZweimalKassiert(t *testing.T) {
	ctx, cmd, _, userID, ksNr, tischID, produktID, varianteID := setupBestellungIntegration(t)
	subject := kasse.TischSessionSubject(ksNr, tischID)
	ref := bestelleEinBier(ctx, t, cmd, subject, userID, produktID, tischID, varianteID)

	ergebnisse := gleichzeitig([]string{uuid.NewString(), uuid.NewString()}, func(zahlungID string) error {
		return cmd.ZahlungKassieren(ctx, userID, "Anna", zahlungID, tischID, []kasse.PositionRef{ref}, "", nil)
	})
	genauEinGewinner(t, ergebnisse, ErrConflict, ErrPositionNichtBezahlbar)

	events, err := cmd.EventRepo.ReadEventsBySubject(ctx, subject)
	if err != nil {
		t.Fatalf("ReadEventsBySubject: %v", err)
	}
	if n := zaehleEvents(events, kasse.EventTypeZahlungKassiertV1) + zaehleEvents(events, kasse.EventTypeZahlungKassiertV2); n != 1 {
		t.Errorf("erwartet 1 zahlung-kassiert-Event, sind %d", n)
	}
	state, err := cmd.EventRepo.ReadTischSession(ctx, subject)
	if err != nil {
		t.Fatalf("ReadTischSession: %v", err)
	}
	if state.GesamtZahlungenCents != 350 || state.SaldoCents != 0 {
		t.Errorf("gesamtZahlungen/saldo = %d/%d, erwartet 350/0", state.GesamtZahlungenCents, state.SaldoCents)
	}
}

// Two devices cancel the same paid position at the same moment with different stornoIds: exactly
// one stornierung-erteilt:v1 lands, the other request is refused.
func TestParallelzugriff_GleichePositionZweimalStorniert(t *testing.T) {
	ctx, cmd, _, userID, ksNr, tischID, produktID, varianteID := setupBestellungIntegration(t)
	subject := kasse.TischSessionSubject(ksNr, tischID)
	ref := bestelleEinBier(ctx, t, cmd, subject, userID, produktID, tischID, varianteID)
	if err := cmd.ZahlungKassieren(ctx, userID, "Anna", uuid.NewString(), tischID, []kasse.PositionRef{ref}, "", nil); err != nil {
		t.Fatalf("ZahlungKassieren: %v", err)
	}

	ergebnisse := gleichzeitig([]string{uuid.NewString(), uuid.NewString()}, func(stornoID string) error {
		return cmd.StornierungErteilen(ctx, userID, "Leitung", uuid.MustParse(stornoID), tischID, []kasse.PositionRef{ref}, "Reklamation")
	})
	genauEinGewinner(t, ergebnisse, ErrConflict, ErrPositionNichtStornierbar)

	events, err := cmd.EventRepo.ReadEventsBySubject(ctx, subject)
	if err != nil {
		t.Fatalf("ReadEventsBySubject: %v", err)
	}
	if n := zaehleEvents(events, kasse.EventTypeStornierungErteiltV1); n != 1 {
		t.Errorf("erwartet 1 stornierung-erteilt:v1-Event, sind %d", n)
	}
	if n := zaehleEvents(events, kasse.EventTypeBestellungKorrigiertV1); n != 0 {
		t.Errorf("erwartet kein bestellung-korrigiert:v1-Event, sind %d", n)
	}
}

// A deletion that commits while a booking waits on the Tisch row wins: the booking sees the deleted
// Tisch and fails, so no saldo lands on a Tisch that left service. The raw transaction holds the lock
// a status change takes.
func TestParallelzugriff_LoeschenWaehrendBestellung(t *testing.T) {
	ctx, cmd, db, userID, ksNr, tischID, produktID, varianteID := setupBestellungIntegration(t)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op
	if _, err := tx.ExecContext(ctx, "SELECT status FROM tische WHERE id = $1 FOR UPDATE", tischID); err != nil {
		t.Fatalf("lock tisch: %v", err)
	}

	ergebnis := make(chan error, 1)
	go func() {
		inputs := []enrichment.PositionInput{{ProduktID: produktID, VarianteID: varianteID, Menge: 1}}
		ergebnis <- cmd.BestellungAufnehmen(ctx, userID, "Anna", uuid.New().String(), tischID, inputs, "")
	}()
	wartetAufZeilensperre(t, db)

	if _, err := tx.ExecContext(ctx, "UPDATE tische SET status = 'deleted' WHERE id = $1", tischID); err != nil {
		t.Fatalf("delete tisch: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}

	if err := <-ergebnis; !errors.Is(err, ErrTischNotActive) {
		t.Fatalf("expected ErrTischNotActive, got %v", err)
	}
	state, err := cmd.EventRepo.ReadTischSession(ctx, kasse.TischSessionSubject(ksNr, tischID))
	if err != nil {
		t.Fatalf("ReadTischSession: %v", err)
	}
	if state.SaldoCents != 0 {
		t.Errorf("deleted tisch carries saldo %d, want 0", state.SaldoCents)
	}
}

// wartetAufZeilensperre returns once a session waits for a lock, so the test commits only after the
// booking reached the Tisch row.
func wartetAufZeilensperre(t *testing.T, db *sql.DB) {
	t.Helper()
	for range 500 {
		var wartend int
		if err := db.QueryRow("SELECT count(*) FROM pg_locks WHERE NOT granted").Scan(&wartend); err != nil {
			t.Fatalf("read pg_locks: %v", err)
		}
		if wartend > 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("the booking never waited for the Tisch row lock")
}
