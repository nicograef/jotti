//go:build integration

package application

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
	"github.com/nicograef/jotti/backend/domain/kasse"
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
