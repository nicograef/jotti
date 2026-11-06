//go:build integration

package application

import (
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
