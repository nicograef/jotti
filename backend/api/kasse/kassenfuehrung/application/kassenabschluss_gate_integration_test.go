//go:build integration

package application

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/repository/kassensitzungen_repo"
	"github.com/nicograef/jotti/backend/repository/tse_repo"
)

func gateCommand(db *sql.DB) Command {
	return Command{
		KassenjournalRepo:   kassenjournal_repo.NewRepository(db),
		KassensitzungenRepo: kassensitzungen_repo.NewRepository(db),
		TSERepo:             tse_repo.NewRepository(db),
		DruckauftragRepo:    &fakeDruckauftraege{},
	}
}

// einlageBuchen books an Einlage; its Signaturauftrag stays offen because no worker runs.
func einlageBuchen(ctx context.Context, t *testing.T, cmd Command, userID, betragCents int) {
	t.Helper()
	if err := cmd.GeldtransitBuchen(ctx, userID, "test", uuid.NewString(), "einlage", betragCents, "Wechselgeld"); err != nil {
		t.Fatalf("Einlage: %v", err)
	}
}

func kassensitzungStatus(t *testing.T, db *sql.DB) string {
	t.Helper()
	var status string
	if err := db.QueryRow("SELECT status FROM kassensitzungen").Scan(&status); err != nil {
		t.Fatalf("kassensitzung status lesen: %v", err)
	}
	return status
}

// A fresh offen Signaturauftrag without Störung blocks the Abschluss before the barrier: the session
// stays offen and no closing event is written.
func TestKasseAbschliessen_GateBlocktBeiAusstehend(t *testing.T) {
	ctx, _, db, userID := setupKassenfuehrungIntegration(t)
	cmd := gateCommand(db)
	einlageBuchen(ctx, t, cmd, userID, 1000)

	_, err := cmd.KasseAbschliessen(ctx, userID, "test", 1000)
	var ausstehend *SignaturenAusstehendError
	if !errors.As(err, &ausstehend) {
		t.Fatalf("expected SignaturenAusstehendError, got %v", err)
	}
	if ausstehend.Anzahl != 1 {
		t.Errorf("expected Anzahl 1, got %d", ausstehend.Anzahl)
	}
	if status := kassensitzungStatus(t, db); status != string(kasse.KassensitzungOffen) {
		t.Errorf("expected status %q, got %q", kasse.KassensitzungOffen, status)
	}
	for _, typ := range []kasse.EventType{kasse.EventTypeKassensturzDurchgefuehrtV1, kasse.EventTypeTagesabschlussErstelltV1} {
		if n := countJournalEvents(t, db, string(typ)); n != 0 {
			t.Errorf("expected no %s when the gate blocks, got %d", typ, n)
		}
	}
}

// During a logged tse_fehler outage the offen order counts as Ausfall-Rest, like the finally
// fehlgeschlagen one after three job-specific failures; the Abschluss passes and reports both.
func TestKasseAbschliessen_GateLaesstAusfallResteDurch(t *testing.T) {
	ctx, _, db, userID := setupKassenfuehrungIntegration(t)
	cmd := gateCommand(db)
	tseRepo := tse_repo.NewRepository(db)
	einlageBuchen(ctx, t, cmd, userID, 1000)
	einlageBuchen(ctx, t, cmd, userID, 500)

	var giftID int
	if err := db.QueryRow("SELECT id FROM tse_signaturauftraege ORDER BY id LIMIT 1").Scan(&giftID); err != nil {
		t.Fatalf("auftrag lesen: %v", err)
	}
	for range tse_repo.MaxSignaturVersuche {
		if err := tseRepo.TSESignaturauftragFehlversuch(ctx, giftID, "E_FAILED_SCHEMA_VALIDATION"); err != nil {
			t.Fatalf("Fehlversuch: %v", err)
		}
	}
	if err := tseRepo.OpenTSEStoerung(ctx, tse.StoerungGrundTSEFehler, "503 service unavailable"); err != nil {
		t.Fatalf("Stoerung oeffnen: %v", err)
	}

	ergebnis, err := cmd.KasseAbschliessen(ctx, userID, "test", 1500)
	if err != nil {
		t.Fatalf("expected the Abschluss to pass with Ausfall-Reste, got %v", err)
	}
	if ergebnis.AusfallResteAnzahl != 2 {
		t.Errorf("expected AusfallResteAnzahl 2 (fehlgeschlagen + offen bei Stoerung), got %d", ergebnis.AusfallResteAnzahl)
	}
	if ergebnis.OhneKonfigurationAnzahl != 0 {
		t.Errorf("expected OhneKonfigurationAnzahl 0, got %d", ergebnis.OhneKonfigurationAnzahl)
	}
	if status := kassensitzungStatus(t, db); status != string(kasse.KassensitzungAbgeschlossen) {
		t.Errorf("expected status %q, got %q", kasse.KassensitzungAbgeschlossen, status)
	}
}
