//go:build integration

package application

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	dbpkg "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/repository/kassensitzungen_repo"
)

func cleanGeldtransitDB(t *testing.T, db *sql.DB) {
	t.Helper()
	stmts := []string{
		"DELETE FROM tse_signaturauftraege",
		"DELETE FROM druckauftraege",
		"DELETE FROM tisch_sessions",
		"ALTER TABLE kassenjournal DISABLE TRIGGER kassenjournal_no_delete",
		"DELETE FROM kassenjournal",
		"ALTER TABLE kassenjournal ENABLE TRIGGER kassenjournal_no_delete",
		"DELETE FROM kassensitzungen",
		"DELETE FROM betreiber",
		"DELETE FROM users",
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("cleanGeldtransitDB %q: %v", stmt, err)
		}
	}
}

func setupGeldtransitIntegration(t *testing.T) (ctx context.Context, cmd Command, db *sql.DB, userID int) {
	t.Helper()
	db = dbpkg.OpenTestDatabase()
	cleanGeldtransitDB(t, db)
	t.Cleanup(func() {
		cleanGeldtransitDB(t, db)
		db.Close()
	})

	ctx = context.Background()

	if err := db.QueryRow(
		"INSERT INTO users (name, username, role, status, password_hash, onetime_password_hash, created_at, updated_at) VALUES ('Test', 'test', 'admin', 'active', 'hash', 'hash', now(), now()) RETURNING id",
	).Scan(&userID); err != nil {
		t.Fatalf("create user: %v", err)
	}

	if _, err := db.Exec(
		"INSERT INTO kassensitzungen (datum, bezeichnung, status, created_at, updated_at) VALUES ((NOW() AT TIME ZONE 'Europe/Berlin')::date, 'Test', 'offen', now(), now())",
	); err != nil {
		t.Fatalf("create kassensitzung: %v", err)
	}

	kjRepo := kassenjournal_repo.NewRepository(db)
	cmd = Command{
		KassenjournalRepo:   kjRepo,
		KassensitzungenRepo: kassensitzungen_repo.NewRepository(db),
	}
	return
}

// TestGeldtransitBuchen_DuplikatGeldtransitId_IdempotenterErfolg: Zwei identische Aufrufe mit
// derselben geldtransitId erzeugen genau ein kassenjournal-Event und einen Signaturauftrag; der
// zweite Aufruf gibt nil zurück (idempotenter Erfolg).
func TestGeldtransitBuchen_DuplikatGeldtransitId_IdempotenterErfolg(t *testing.T) {
	ctx, cmd, db, userID := setupGeldtransitIntegration(t)

	geldtransitID := uuid.New().String()

	if err := cmd.GeldtransitBuchen(ctx, userID, "test", geldtransitID, "einlage", 1000, "Test"); err != nil {
		t.Fatalf("erster Aufruf: %v", err)
	}

	if err := cmd.GeldtransitBuchen(ctx, userID, "test", geldtransitID, "einlage", 1000, "Test"); err != nil {
		t.Fatalf("zweiter Aufruf (Duplikat) erwartet nil, bekam: %v", err)
	}

	var eventCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM kassenjournal WHERE type = 'geldtransit-gebucht:v1'").Scan(&eventCount); err != nil {
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
