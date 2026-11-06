//go:build integration

package application

import (
	"testing"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
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
