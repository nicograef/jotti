//go:build integration

package application

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
	dbpkg "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/repository/druckstation_repo"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/repository/kassensitzungen_repo"
	"github.com/nicograef/jotti/backend/repository/produkt_repo"
	"github.com/nicograef/jotti/backend/repository/tisch_repo"
)

func cleanTischDB(t *testing.T, db *sql.DB) {
	t.Helper()
	stmts := []string{
		"DELETE FROM tse_signaturauftraege",
		"DELETE FROM druckauftraege",
		"DELETE FROM tisch_sessions",
		"ALTER TABLE kassenjournal DISABLE TRIGGER kassenjournal_no_delete",
		"DELETE FROM kassenjournal",
		"ALTER TABLE kassenjournal ENABLE TRIGGER kassenjournal_no_delete",
		"DELETE FROM kassensitzungen",
		"DELETE FROM tische",
		"DELETE FROM produkt_varianten",
		"DELETE FROM produkte",
		"DELETE FROM users",
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("cleanTischDB %q: %v", stmt, err)
		}
	}
}

func setupBestellungIntegration(t *testing.T) (ctx context.Context, cmd Command, db *sql.DB, userID, ksNr, tischID, produktID, varianteID int) {
	t.Helper()
	db = dbpkg.OpenTestDatabase()
	cleanTischDB(t, db)
	t.Cleanup(func() {
		cleanTischDB(t, db)
		_ = db.Close()
	})

	ctx = context.Background()

	if err := db.QueryRow(
		"INSERT INTO users (name, username, role, status, password_hash, onetime_password_hash, created_at, updated_at) VALUES ('Test', 'test', 'admin', 'active', 'hash', 'hash', now(), now()) RETURNING id",
	).Scan(&userID); err != nil {
		t.Fatalf("create user: %v", err)
	}

	if err := db.QueryRow(
		"INSERT INTO kassensitzungen (datum, bezeichnung, status, created_at, updated_at) VALUES ((NOW() AT TIME ZONE 'Europe/Berlin')::date, 'Test', 'offen', now(), now()) RETURNING z_nr",
	).Scan(&ksNr); err != nil {
		t.Fatalf("create kassensitzung: %v", err)
	}

	if err := db.QueryRow(
		"INSERT INTO tische (name, status, created_at, updated_at) VALUES ('Tisch 1', 'active', now(), now()) RETURNING id",
	).Scan(&tischID); err != nil {
		t.Fatalf("create tisch: %v", err)
	}

	if err := db.QueryRow(
		"INSERT INTO produkte (name, kategorie, steuersatz, status, created_at, updated_at) VALUES ('Bier', 'getraenk', 'regel', 'active', now(), now()) RETURNING id",
	).Scan(&produktID); err != nil {
		t.Fatalf("create produkt: %v", err)
	}

	if err := db.QueryRow(
		"INSERT INTO produkt_varianten (produkt_id, name, preis_cents, status, created_at, updated_at) VALUES ($1, '0.5L', 350, 'active', now(), now()) RETURNING id",
		produktID,
	).Scan(&varianteID); err != nil {
		t.Fatalf("create variante: %v", err)
	}

	cmd = Command{
		TischRepo:           tisch_repo.NewRepository(db),
		EventRepo:           kassenjournal_repo.NewRepository(db),
		ProduktRepo:         produkt_repo.NewRepository(db),
		KassensitzungenRepo: kassensitzungen_repo.NewRepository(db),
		DruckstationRepo:    druckstation_repo.NewRepository(db),
	}
	return
}

// TestBestellungAufnehmen_DuplikatBestellungId_IdempotenterErfolg: Zwei identische Aufrufe mit
// derselben bestellungId erzeugen genau ein kassenjournal-Event und einen Signaturauftrag; der
// zweite Aufruf gibt nil zurück (idempotenter Erfolg).
func TestBestellungAufnehmen_DuplikatBestellungId_IdempotenterErfolg(t *testing.T) {
	ctx, cmd, db, userID, _, tischID, produktID, varianteID := setupBestellungIntegration(t)

	bestellungID := uuid.New().String()
	inputs := []enrichment.PositionInput{{ProduktID: produktID, VarianteID: varianteID, Menge: 1}}

	if err := cmd.BestellungAufnehmen(ctx, userID, "test", bestellungID, tischID, inputs, ""); err != nil {
		t.Fatalf("erster Aufruf: %v", err)
	}

	if err := cmd.BestellungAufnehmen(ctx, userID, "test", bestellungID, tischID, inputs, ""); err != nil {
		t.Fatalf("zweiter Aufruf (Duplikat) erwartet nil, bekam: %v", err)
	}

	var eventCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM kassenjournal WHERE type = 'bestellung-aufgenommen:v1'").Scan(&eventCount); err != nil {
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
