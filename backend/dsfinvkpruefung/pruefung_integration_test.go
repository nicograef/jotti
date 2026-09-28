//go:build integration

package dsfinvkpruefung_test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	exportApp "github.com/nicograef/jotti/backend/api/fiskal/export/application"
	"github.com/nicograef/jotti/backend/db/dbtest"
	"github.com/nicograef/jotti/backend/dsfinvkpruefung"
	"github.com/nicograef/jotti/backend/repository/betreiber_repo"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/repository/kassensitzungen_repo"
	"github.com/nicograef/jotti/backend/repository/tisch_repo"
	"github.com/nicograef/jotti/backend/repository/tse_repo"
	"github.com/nicograef/jotti/backend/seed"
)

// cleanSeedDB empties every seeded table so the seeder guard does not trip.
// The order mirrors the seeder integration test (foreign keys).
func cleanSeedDB(t *testing.T, db *sql.DB) {
	t.Helper()
	stmts := []string{
		"DELETE FROM tse_signaturauftraege",
		"DELETE FROM tse_stoerungen",
		"DELETE FROM druckauftraege",
		"UPDATE druckstationen SET drucker_ip = '', bonmodus = 'pro_position' WHERE kategorie IN ('essen', 'getraenk', 'sonstiges')",
		"UPDATE druckstationen SET drucker_ip = '', bonmodus = 'pro_bestellung' WHERE kategorie = 'abholbon'",
		"UPDATE druckstationen SET drucker_ip = '', bonmodus = NULL WHERE kategorie = 'kassenbeleg'",
		"DELETE FROM tisch_favoriten",
		"DELETE FROM tisch_sessions",
		"ALTER TABLE kassenjournal DISABLE TRIGGER kassenjournal_no_delete",
		"DELETE FROM kassenjournal",
		"ALTER TABLE kassenjournal ENABLE TRIGGER kassenjournal_no_delete",
		"DELETE FROM kassensitzungen",
		"DELETE FROM produkt_varianten",
		"DELETE FROM produkte",
		"DELETE FROM tische",
		"DELETE FROM betreiber",
		"DELETE FROM users",
		// Reset tse_konfiguration to the migration's empty singleton row, whether or not a reset ran before.
		"INSERT INTO tse_konfiguration (id, api_key, api_secret, tss_id, client_id, updated_at) VALUES (1, '', '', '', '', NOW()) ON CONFLICT (id) DO UPDATE SET api_key = '', api_secret = '', tss_id = '', client_id = ''",
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("cleanSeedDB %q: %v", stmt, err)
		}
	}
}

// TestSeedExportBefundfrei runs seed, real DSFinV-K export and check: every seeded Kassensitzung's
// archive must be free of findings. The seeder's fake TSE supplies realistic signature data.
func TestSeedExportBefundfrei(t *testing.T) {
	db := dbtest.Open()
	cleanSeedDB(t, db)
	t.Cleanup(func() { cleanSeedDB(t, db) })

	ctx := context.Background()
	if err := seed.Run(ctx, db); err != nil {
		t.Fatalf("seed.Run: %v", err)
	}

	export := exportApp.Export{
		KassenjournalRepo:   kassenjournal_repo.NewRepository(db),
		KassensitzungenRepo: kassensitzungen_repo.NewRepository(db),
		BetreiberRepo:       betreiber_repo.NewRepository(db),
		TSERepo:             tse_repo.NewRepository(db),
		TischRepo:           tisch_repo.NewRepository(db),
		Version:             "integration-test",
	}

	// The seeder creates three Kassensitzungen (two closed, one open).
	for _, znr := range []int{1, 2, 3} {
		archiv, err := export.Erstellen(ctx, znr)
		if err != nil {
			t.Fatalf("Export Kassensitzung %d: %v", znr, err)
		}

		befunde, err := dsfinvkpruefung.PruefenBytes(archiv.Inhalt)
		if err != nil {
			t.Fatalf("Prüfung Kassensitzung %d: %v", znr, err)
		}
		if len(befunde) != 0 {
			for _, b := range befunde {
				t.Errorf("Kassensitzung %d Befund: %s", znr, b.String())
			}
		}
	}
}
