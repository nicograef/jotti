//go:build integration

package application

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
	tischgeschaeftApp "github.com/nicograef/jotti/backend/api/kasse/tischgeschaeft/application"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/repository/druckstation_repo"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/repository/kassensitzungen_repo"
	"github.com/nicograef/jotti/backend/repository/produkt_repo"
	"github.com/nicograef/jotti/backend/repository/tisch_repo"
)

// TestPfandZurueckgeben_DuplikatRueckgabeId_IdempotenterErfolg: two calls with the same rueckgabeId store one event
// and one Signaturauftrag; the second call answers success.
func TestPfandZurueckgeben_DuplikatRueckgabeId_IdempotenterErfolg(t *testing.T) {
	ctx, cmd, db, userID, _, produktID, varianteID := setupDVIntegration(t)

	if _, err := db.Exec("UPDATE produkte SET pfand = true WHERE id = $1", produktID); err != nil {
		t.Fatalf("Pfand-Kennzeichen setzen: %v", err)
	}

	rueckgabeID := uuid.New().String()
	inputs := []enrichment.PositionInput{{ProduktID: produktID, VarianteID: varianteID, Menge: 2}}

	if err := cmd.PfandZurueckgeben(ctx, userID, "test", rueckgabeID, 0, inputs, ""); err != nil {
		t.Fatalf("erster Aufruf: %v", err)
	}
	if err := cmd.PfandZurueckgeben(ctx, userID, "test", rueckgabeID, 0, inputs, ""); err != nil {
		t.Errorf("zweiter Aufruf (Duplikat) erwartet nil, bekam: %v", err)
	}

	var eventCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM kassenjournal WHERE type = 'pfand-zurueckgegeben:v1'").Scan(&eventCount); err != nil {
		t.Fatalf("kassenjournal zählen: %v", err)
	}
	if eventCount != 1 {
		t.Errorf("erwartet 1 Event im kassenjournal, gespeichert: %d", eventCount)
	}

	var auftraegeCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM tse_signaturauftraege").Scan(&auftraegeCount); err != nil {
		t.Fatalf("tse_signaturauftraege zählen: %v", err)
	}
	if auftraegeCount != 1 {
		t.Errorf("erwartet 1 tse_signaturauftrag, vorhanden: %d", auftraegeCount)
	}
}

// TestPfandZurueckgeben_AbweichendeRueckgabe_KeinEvent: a pure-return retry under a booked
// rueckgabeId with other quantities, or whose Rückgabe was offset against a Verkauf, is refused with
// ErrPfandrueckgabeAbweichend and books nothing.
func TestPfandZurueckgeben_AbweichendeRueckgabe_KeinEvent(t *testing.T) {
	// Each case books a first Vorgang; then follows a pure return of Pfand Menge 2 at tischId 0.
	tests := []struct {
		name string
		erst func(f verrechnungsFall) error
	}{
		{"andere Menge", func(f verrechnungsFall) error {
			return f.cmd.PfandZurueckgeben(f.ctx, f.userID, "test", f.rueckgabeID, 0, f.pfand(3), "")
		}},
		{"mit Verkauf verrechnet", func(f verrechnungsFall) error {
			return f.cmd.DirektverkaufTaetigen(f.ctx, f.userID, "test", f.verkaufID, f.verkauf(2), "", &enrichment.PfandEingabe{RueckgabeID: f.rueckgabeID, Positionen: f.pfand(2)})
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cmd, db, userID, _, produktID, varianteID := setupDVIntegration(t)
			pfandProduktID, pfandVarianteID := insertPfandVariante(t, db)
			f := verrechnungsFall{
				ctx: ctx, cmd: cmd, userID: userID,
				verkaufID: uuid.New().String(), rueckgabeID: uuid.New().String(),
				verkauf: func(menge int) []enrichment.PositionInput {
					return []enrichment.PositionInput{{ProduktID: produktID, VarianteID: varianteID, Menge: menge}}
				},
				pfand: func(menge int) []enrichment.PositionInput {
					return []enrichment.PositionInput{{ProduktID: pfandProduktID, VarianteID: pfandVarianteID, Menge: menge}}
				},
			}

			if err := tt.erst(f); err != nil {
				t.Fatalf("erster Vorgang: %v", err)
			}
			var vorher int
			if err := db.QueryRow("SELECT COUNT(*) FROM kassenjournal").Scan(&vorher); err != nil {
				t.Fatalf("kassenjournal zählen: %v", err)
			}

			if err := cmd.PfandZurueckgeben(ctx, userID, "test", f.rueckgabeID, 0, f.pfand(2), ""); !errors.Is(err, ErrPfandrueckgabeAbweichend) {
				t.Errorf("erwartet ErrPfandrueckgabeAbweichend, bekam: %v", err)
			}

			var nachher int
			if err := db.QueryRow("SELECT COUNT(*) FROM kassenjournal").Scan(&nachher); err != nil {
				t.Fatalf("kassenjournal zählen: %v", err)
			}
			if nachher != vorher {
				t.Errorf("erwartet kein neues Event, kassenjournal wuchs von %d auf %d", vorher, nachher)
			}
		})
	}
}

// TestPfandZurueckgeben_AmTisch: a pure return at an active table books the tischId; an inactive or
// unknown table books nothing; a retry under a rueckgabeId settled with a Zahlung at that table is
// refused, since answering success would pay out cash the journal never booked.
func TestPfandZurueckgeben_AmTisch(t *testing.T) {
	const unbekannterTisch = 999999
	tests := []struct {
		name        string
		tischStatus string // leer: kein Tisch angelegt
		verrechnet  bool
		wantErr     error
	}{
		{"aktiver Tisch", "active", false, nil},
		{"inaktiver Tisch", "inactive", false, ErrTischNichtAktiv},
		{"unbekannter Tisch", "", false, ErrTischNichtGefunden},
		{"mit Zahlung verrechnet", "active", true, ErrPfandrueckgabeAbweichend},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cmd, db, userID, ksNr, produktID, varianteID := setupDVIntegration(t)
			pfandProduktID, pfandVarianteID := insertPfandVariante(t, db)
			pfand := []enrichment.PositionInput{{ProduktID: pfandProduktID, VarianteID: pfandVarianteID, Menge: 2}}
			rueckgabeID := uuid.New().String()

			tischID := unbekannterTisch
			if tt.tischStatus != "" {
				if err := db.QueryRow(
					"INSERT INTO tische (name, status, created_at, updated_at) VALUES ('Tisch 1', $1, now(), now()) RETURNING id", tt.tischStatus,
				).Scan(&tischID); err != nil {
					t.Fatalf("create tisch: %v", err)
				}
			}
			if tt.verrechnet {
				verrechneMitZahlung(t, db, userID, ksNr, tischID, enrichment.PositionInput{ProduktID: produktID, VarianteID: varianteID, Menge: 1}, &enrichment.PfandEingabe{RueckgabeID: rueckgabeID, Positionen: pfand})
			}
			var vorher int
			if err := db.QueryRow("SELECT COUNT(*) FROM kassenjournal").Scan(&vorher); err != nil {
				t.Fatalf("kassenjournal zählen: %v", err)
			}

			err := cmd.PfandZurueckgeben(ctx, userID, "test", rueckgabeID, tischID, pfand, "")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("erwartet %v, bekam: %v", tt.wantErr, err)
			}

			var nachher int
			if err := db.QueryRow("SELECT COUNT(*) FROM kassenjournal").Scan(&nachher); err != nil {
				t.Fatalf("kassenjournal zählen: %v", err)
			}
			if tt.wantErr != nil {
				if nachher != vorher {
					t.Errorf("erwartet kein neues Event, kassenjournal wuchs von %d auf %d", vorher, nachher)
				}
				return
			}
			var gespeicherterTisch int
			if err := db.QueryRow("SELECT (data->>'tischId')::int FROM kassenjournal WHERE type = 'pfand-zurueckgegeben:v1' AND data->>'rueckgabeId' = $1", rueckgabeID).Scan(&gespeicherterTisch); err != nil {
				t.Fatalf("Rückgabe lesen: %v", err)
			}
			if gespeicherterTisch != tischID {
				t.Errorf("erwartet tischId %d an der Rückgabe, bekam %d", tischID, gespeicherterTisch)
			}
		})
	}
}

// verrechneMitZahlung books an order at tischID and pays it with pfand offset, through the
// Tischgeschäft commands.
func verrechneMitZahlung(t *testing.T, db *sql.DB, userID, ksNr, tischID int, bestellt enrichment.PositionInput, pfand *enrichment.PfandEingabe) {
	t.Helper()
	ctx := context.Background()
	tg := tischgeschaeftApp.Command{
		TischRepo:           tisch_repo.NewRepository(db),
		EventRepo:           kassenjournal_repo.NewRepository(db),
		ProduktRepo:         produkt_repo.NewRepository(db),
		KassensitzungenRepo: kassensitzungen_repo.NewRepository(db),
		DruckstationRepo:    druckstation_repo.NewRepository(db),
	}
	if err := tg.BestellungAufnehmen(ctx, userID, "test", uuid.New().String(), tischID, []enrichment.PositionInput{bestellt}, ""); err != nil {
		t.Fatalf("Bestellung aufnehmen: %v", err)
	}
	session, err := kassenjournal_repo.NewRepository(db).ReadTischSession(ctx, kasse.TischSessionSubject(ksNr, tischID))
	if err != nil {
		t.Fatalf("Tisch-Session lesen: %v", err)
	}
	refs := []kasse.PositionRef{{PositionID: session.UnbezahltePositionen[0].PositionID, Menge: bestellt.Menge}}
	if err := tg.ZahlungKassieren(ctx, userID, "test", tischID, refs, "", pfand); err != nil {
		t.Fatalf("Zahlung mit Pfand kassieren: %v", err)
	}
}

// TestPfandrueckgabeStornieren_DuplikatStornierungId_IdempotenterErfolg: two cancellations with the
// same stornierungId store one event and one Signaturauftrag; the second call answers success.
func TestPfandrueckgabeStornieren_DuplikatStornierungId_IdempotenterErfolg(t *testing.T) {
	ctx, cmd, db, userID, ksNr, produktID, varianteID := setupDVIntegration(t)

	if _, err := db.Exec("UPDATE produkte SET pfand = true WHERE id = $1", produktID); err != nil {
		t.Fatalf("Pfand-Kennzeichen setzen: %v", err)
	}

	rueckgabeID := uuid.New().String()
	inputs := []enrichment.PositionInput{{ProduktID: produktID, VarianteID: varianteID, Menge: 2}}
	if err := cmd.PfandZurueckgeben(ctx, userID, "test", rueckgabeID, 0, inputs, ""); err != nil {
		t.Fatalf("Rückgabe: %v", err)
	}

	events, err := cmd.EventRepo.ReadEventsBySubject(ctx, kasse.PfandrueckgabeSubject(ksNr, rueckgabeID))
	if err != nil {
		t.Fatalf("Stream lesen: %v", err)
	}
	eintrag, err := kasse.BuildPfandrueckgabeHistorieEintrag(events)
	if err != nil {
		t.Fatalf("Replay: %v", err)
	}
	refs := []kasse.PositionRef{{PositionID: eintrag.OffenePositionen[0].PositionID, Menge: 1}}

	stornierungID := uuid.New().String()
	if err := cmd.PfandrueckgabeStornieren(ctx, userID, "test", stornierungID, rueckgabeID, refs, "Falsch gezählt"); err != nil {
		t.Fatalf("erste Stornierung: %v", err)
	}
	if err := cmd.PfandrueckgabeStornieren(ctx, userID, "test", stornierungID, rueckgabeID, refs, "Falsch gezählt"); err != nil {
		t.Errorf("zweite Stornierung (Duplikat) erwartet nil, bekam: %v", err)
	}

	var stornoCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM kassenjournal WHERE type = 'pfandrueckgabe-storniert:v1'").Scan(&stornoCount); err != nil {
		t.Fatalf("kassenjournal zählen: %v", err)
	}
	if stornoCount != 1 {
		t.Errorf("erwartet 1 Stornierung im kassenjournal, gespeichert: %d", stornoCount)
	}

	var auftraegeCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM tse_signaturauftraege").Scan(&auftraegeCount); err != nil {
		t.Fatalf("tse_signaturauftraege zählen: %v", err)
	}
	if auftraegeCount != 2 {
		t.Errorf("erwartet 2 tse_signaturauftraege (Rückgabe und Stornierung), vorhanden: %d", auftraegeCount)
	}
}

// TestPfandrueckgabeStornieren_AbweichendeStornierungId_KeinEvent: a stornierungId booked for one
// Rückgabe and reused for another Rückgabe, or for other quantities, is refused with
// ErrPfandrueckgabeAbweichend and books nothing; it would duplicate a BON_ID or skip a Storno.
func TestPfandrueckgabeStornieren_AbweichendeStornierungId_KeinEvent(t *testing.T) {
	tests := []struct {
		name            string
		andereRueckgabe bool
		menge           int
	}{
		{"andere Rückgabe", true, 1},
		{"andere Menge", false, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cmd, db, userID, ksNr, _, _ := setupDVIntegration(t)
			pfandProduktID, pfandVarianteID := insertPfandVariante(t, db)
			inputs := []enrichment.PositionInput{{ProduktID: pfandProduktID, VarianteID: pfandVarianteID, Menge: 2}}

			zurueckgeben := func() (string, string) {
				rueckgabeID := uuid.New().String()
				if err := cmd.PfandZurueckgeben(ctx, userID, "test", rueckgabeID, 0, inputs, ""); err != nil {
					t.Fatalf("Rückgabe: %v", err)
				}
				events, err := cmd.EventRepo.ReadEventsBySubject(ctx, kasse.PfandrueckgabeSubject(ksNr, rueckgabeID))
				if err != nil {
					t.Fatalf("Stream lesen: %v", err)
				}
				eintrag, err := kasse.BuildPfandrueckgabeHistorieEintrag(events)
				if err != nil {
					t.Fatalf("Replay: %v", err)
				}
				return rueckgabeID, eintrag.OffenePositionen[0].PositionID
			}

			rueckgabeID, positionID := zurueckgeben()
			stornierungID := uuid.New().String()
			if err := cmd.PfandrueckgabeStornieren(ctx, userID, "test", stornierungID, rueckgabeID, []kasse.PositionRef{{PositionID: positionID, Menge: 1}}, "Falsch gezählt"); err != nil {
				t.Fatalf("erste Stornierung: %v", err)
			}
			if tt.andereRueckgabe {
				rueckgabeID, positionID = zurueckgeben()
			}
			var vorher int
			if err := db.QueryRow("SELECT COUNT(*) FROM kassenjournal").Scan(&vorher); err != nil {
				t.Fatalf("kassenjournal zählen: %v", err)
			}

			err := cmd.PfandrueckgabeStornieren(ctx, userID, "test", stornierungID, rueckgabeID, []kasse.PositionRef{{PositionID: positionID, Menge: tt.menge}}, "Falsch gezählt")
			if !errors.Is(err, ErrPfandrueckgabeAbweichend) {
				t.Errorf("erwartet ErrPfandrueckgabeAbweichend, bekam: %v", err)
			}

			var nachher int
			if err := db.QueryRow("SELECT COUNT(*) FROM kassenjournal").Scan(&nachher); err != nil {
				t.Fatalf("kassenjournal zählen: %v", err)
			}
			if nachher != vorher {
				t.Errorf("erwartet kein neues Event, kassenjournal wuchs von %d auf %d", vorher, nachher)
			}
		})
	}
}

// A stornierungId equal to the rueckgabeId of another Rückgabe would export two Belege with one
// BON_ID; the cross-type BON_ID index refuses it and nothing is booked.
func TestPfandrueckgabeStornieren_StornierungIdGleichFremderRueckgabeId_KeinEvent(t *testing.T) {
	ctx, cmd, db, userID, ksNr, _, _ := setupDVIntegration(t)
	pfandProduktID, pfandVarianteID := insertPfandVariante(t, db)
	inputs := []enrichment.PositionInput{{ProduktID: pfandProduktID, VarianteID: pfandVarianteID, Menge: 1}}

	rueckgabeA, rueckgabeB := uuid.New().String(), uuid.New().String()
	for _, id := range []string{rueckgabeA, rueckgabeB} {
		if err := cmd.PfandZurueckgeben(ctx, userID, "test", id, 0, inputs, ""); err != nil {
			t.Fatalf("Rückgabe %s: %v", id, err)
		}
	}
	events, err := cmd.EventRepo.ReadEventsBySubject(ctx, kasse.PfandrueckgabeSubject(ksNr, rueckgabeA))
	if err != nil {
		t.Fatalf("Stream lesen: %v", err)
	}
	eintrag, err := kasse.BuildPfandrueckgabeHistorieEintrag(events)
	if err != nil {
		t.Fatalf("Replay: %v", err)
	}
	var vorher int
	if err := db.QueryRow("SELECT COUNT(*) FROM kassenjournal").Scan(&vorher); err != nil {
		t.Fatalf("kassenjournal zählen: %v", err)
	}

	err = cmd.PfandrueckgabeStornieren(ctx, userID, "test", rueckgabeB, rueckgabeA, []kasse.PositionRef{{PositionID: eintrag.OffenePositionen[0].PositionID, Menge: 1}}, "Falsch gezählt")
	if !errors.Is(err, ErrConflict) {
		t.Errorf("erwartet ErrConflict, bekam: %v", err)
	}

	var nachher int
	if err := db.QueryRow("SELECT COUNT(*) FROM kassenjournal").Scan(&nachher); err != nil {
		t.Fatalf("kassenjournal zählen: %v", err)
	}
	if nachher != vorher {
		t.Errorf("erwartet kein neues Event, kassenjournal wuchs von %d auf %d", vorher, nachher)
	}
}
