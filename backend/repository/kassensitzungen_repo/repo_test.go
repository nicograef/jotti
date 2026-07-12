//go:build integration

package kassensitzungen_repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	dbpkg "github.com/nicograef/jotti/backend/db"
)

func cleanDB(t *testing.T, db *sql.DB) {
	t.Helper()
	// tisch_sessions referenziert kassenjournal (last_event_id) und muss zuerst weg —
	// auch Hinterlassenschaften anderer Testpakete auf der geteilten Test-DB.
	if _, err := db.Exec("DELETE FROM tisch_sessions"); err != nil {
		t.Fatalf("Failed to clean tisch_sessions table: %v", err)
	}
	if _, err := db.Exec("ALTER TABLE kassenjournal DISABLE TRIGGER kassenjournal_no_delete"); err != nil {
		t.Fatalf("Failed to disable kassenjournal_no_delete trigger: %v", err)
	}
	if _, err := db.Exec("DELETE FROM kassenjournal"); err != nil {
		t.Fatalf("Failed to clean kassenjournal table: %v", err)
	}
	if _, err := db.Exec("ALTER TABLE kassenjournal ENABLE TRIGGER kassenjournal_no_delete"); err != nil {
		t.Fatalf("Failed to enable kassenjournal_no_delete trigger: %v", err)
	}
	if _, err := db.Exec("DELETE FROM kassensitzungen"); err != nil {
		t.Fatalf("Failed to clean kassensitzungen table: %v", err)
	}
	if _, err := db.Exec("DELETE FROM users"); err != nil {
		t.Fatalf("Failed to clean users table: %v", err)
	}
}

func createUser(t *testing.T, db *sql.DB, name, username string) int {
	t.Helper()
	var id int
	err := db.QueryRow(
		"INSERT INTO users (name, username, role, status, password_hash, onetime_password_hash, created_at, updated_at) VALUES ($1, $2, 'admin', 'active', 'hash', 'hash', now(), now()) RETURNING id",
		name, username,
	).Scan(&id)
	if err != nil {
		t.Fatalf("Failed to insert user %q: %v", username, err)
	}
	return id
}

func createKassensitzung(t *testing.T, db *sql.DB, datum string, bezeichnung, status string) int {
	t.Helper()
	var zNr int
	err := db.QueryRow(
		"INSERT INTO kassensitzungen (datum, bezeichnung, status, created_at, updated_at) VALUES ($1::date, $2, $3, NOW(), NOW()) RETURNING z_nr",
		datum, bezeichnung, status,
	).Scan(&zNr)
	if err != nil {
		t.Fatalf("Failed to create kassensitzung: %v", err)
	}
	return zNr
}

func insertEvent(t *testing.T, db *sql.DB, userID int, userName, eventType, subject string, version int, data map[string]any, ksNr int) {
	t.Helper()
	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal event data: %v", err)
	}
	_, err = db.Exec(
		"INSERT INTO kassenjournal (user_id, user_name, type, subject, version, data, timestamp, kassensitzung_nr) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		userID, userName, eventType, subject, version, raw, time.Now().UTC(), ksNr,
	)
	if err != nil {
		t.Fatalf("Failed to insert %s event: %v", eventType, err)
	}
}

// TestGetAbgeschlosseneKassensitzungen_MitUmsatzUndAbschlusszeit verifiziert, dass
// jeder Sitzungslisten-Eintrag den Gesamtumsatz und den Abschlusszeitpunkt aus dem
// tagesabschluss-erstellt:v1-Event projiziert und die offene Sitzung ausblendet.
func TestGetAbgeschlosseneKassensitzungen_MitUmsatzUndAbschlusszeit(t *testing.T) {
	db := dbpkg.OpenTestDatabase()
	defer func() { _ = db.Close() }()
	cleanDB(t, db)
	defer cleanDB(t, db)

	repo := NewRepository(db)
	ctx := context.Background()

	userID := createUser(t, db, "Nico Gräf", "nico")

	// Zwei abgeschlossene Sitzungen mit Tagesabschluss-Event, eine offene ohne.
	ks10 := createKassensitzung(t, db, "2026-05-01", "Maihock", "abgeschlossen")
	insertEvent(t, db, userID, "nico", "tagesabschluss-erstellt:v1", "kassensitzung-10", 1, map[string]any{
		"zNr": ks10, "umsatzGesamtCents": 210850, "erstelltVon": userID,
	}, ks10)

	ks11 := createKassensitzung(t, db, "2026-07-05", "Sommerfest Tag 1", "abgeschlossen")
	insertEvent(t, db, userID, "nico", "tagesabschluss-erstellt:v1", "kassensitzung-11", 1, map[string]any{
		"zNr": ks11, "umsatzGesamtCents": 341200, "erstelltVon": userID,
	}, ks11)

	ksOffen := createKassensitzung(t, db, "2026-07-06", "Sommerfest Tag 2", "offen")
	insertEvent(t, db, userID, "nico", "kassensitzung-eroeffnet:v1", "kassensitzung-12", 1, map[string]any{
		"datum": "2026-07-06", "bezeichnung": "Sommerfest Tag 2", "betragCents": 10000, "eroeffnetVon": userID,
	}, ksOffen)

	sitzungen, err := repo.GetAbgeschlosseneKassensitzungen(ctx)
	if err != nil {
		t.Fatalf("GetAbgeschlosseneKassensitzungen failed: %v", err)
	}

	if len(sitzungen) != 2 {
		t.Fatalf("expected 2 abgeschlossene Sitzungen (offene ausgeblendet), got %d", len(sitzungen))
	}

	// Sortierung: Datum DESC, also Sommerfest (07-05) vor Maihock (05-01).
	if sitzungen[0].ZNr != ks11 || sitzungen[0].UmsatzGesamtCents != 341200 {
		t.Errorf("unexpected first entry: %+v", sitzungen[0])
	}
	if sitzungen[0].AbgeschlossenAm == nil {
		t.Error("expected abgeschlossenAm to be set for a closed session")
	}
	if sitzungen[1].ZNr != ks10 || sitzungen[1].UmsatzGesamtCents != 210850 {
		t.Errorf("unexpected second entry: %+v", sitzungen[1])
	}
}
