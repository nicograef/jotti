//go:build integration

package application

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nicograef/jotti/backend/domain/kasse"
)

// TestParallelzugriff_ZweiClientsSelberTisch prüft die Konsistenz von
// Kassenjournal, seiner Projektion (tisch_sessions) und den Salden, wenn zwei
// nebenläufige Servicekräfte DENSELBEN Tisch bedienen (bestellen, kassieren) —
// über mehrere Runden hinweg.
//
// Jede Servicekraft arbeitet auf einer eigenen Produktvariante (A vs. B), sodass
// die erwarteten Endsummen je Variante eindeutig sind. Der geteilte Event-Stream
// des Tisches serialisiert nebenläufige Schreibzugriffe über die optimistische
// Nebenläufigkeitskontrolle (OCC); ein Konflikt (ErrConflict) ist erwartet und
// wird per Retry aufgelöst. Am Ende muss gelten:
//   - Journal-Replay == projizierte tisch_sessions-Zeile (kein Projektions-Drift);
//   - Saldo 0 (jede bestellte Position wurde genau einmal bezahlt);
//   - je Variante: bestellte Menge == bezahlte Menge (keine verlorene Position,
//     keine Doppelbuchung).
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
		case string(kasse.EventTypeZahlungKassiertV1):
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

// bedieneTisch führt für eine Servicekraft runden × (bestellen → kassieren) auf
// ihrer eigenen Variante aus, plus einen abschließenden Durchlauf für etwaige
// durch Interleaving übrig gebliebene Positionen. Alle Schreibzugriffe treffen
// den geteilten Tisch-Stream; OCC-Konflikte werden per Retry aufgelöst.
func bedieneTisch(ctx context.Context, cmd Command, subject string, userID int, userName string, produktID, tischID, varianteID, runden, menge int) error {
	for r := 0; r < runden; r++ {
		bestellungID := uuid.New().String()
		inputs := []BestellPositionInput{{ProduktID: produktID, VarianteID: varianteID, Menge: menge}}
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
		return cmd.ZahlungKassieren(ctx, userID, userName, tischID, refs, "")
	})
}

// offeneRefsFuerVariante liest die aktuelle Tisch-Session und liefert die noch
// unbezahlten Positionen der gegebenen Variante als PositionRefs. So bearbeitet
// jede Servicekraft nur ihre eigenen Positionen; eine Doppelverarbeitung
// derselben Position verhindert zusätzlich die serverseitige Bezahl-Invariante.
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

// retryConflict wiederholt op, solange ein OCC-Konflikt (ErrConflict) auftritt.
// Nebenläufige Schreibzugriffe auf denselben Stream lösen einen Konflikt aus; der
// Retry liest den frischen Zustand und versucht erneut. Begrenzt, damit ein echter
// Dauerfehler nicht zur Endlosschleife wird.
func retryConflict(op func() error) error {
	const maxVersuche = 500
	for i := 0; i < maxVersuche; i++ {
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
