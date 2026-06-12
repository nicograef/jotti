//go:build integration

package seed

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	dbpkg "github.com/nicograef/jotti/backend/db"
)

func cleanSeedDB(t *testing.T, db *sql.DB) {
	t.Helper()
	stmts := []string{
		"DELETE FROM tse_signaturen",
		"DELETE FROM tse_nachsignier_auftraege",
		"DELETE FROM druckauftraege",
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
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("cleanSeedDB %q: %v", stmt, err)
		}
	}
}

func TestSeedRun_ErstlaufUndGuard(t *testing.T) {
	db := dbpkg.OpenTestDatabase()
	cleanSeedDB(t, db)
	t.Cleanup(func() { cleanSeedDB(t, db) })

	ctx := context.Background()

	// --- Erstlauf: erfolgreiches Seeding ---
	if err := Run(ctx, db); err != nil {
		t.Fatalf("Erstlauf Run: %v", err)
	}

	// Demo-Login: Benutzer hat den jotti123-Passwort-Hash gesetzt.
	var hash sql.NullString
	if err := db.QueryRow("SELECT password_hash FROM users WHERE username = $1", "maria").Scan(&hash); err != nil {
		t.Fatalf("Benutzer maria abfragen: %v", err)
	}
	if !hash.Valid || hash.String != demoArgon2idHash {
		t.Error("Benutzer maria: password_hash nicht korrekt gesetzt")
	}

	// Betreiber sichtbar.
	var vereinsname string
	if err := db.QueryRow("SELECT vereinsname FROM betreiber WHERE id = 1").Scan(&vereinsname); err != nil {
		t.Fatalf("Betreiber abfragen: %v", err)
	}
	if vereinsname != "TSV Musterstadt e.V." {
		t.Errorf("Betreiber vereinsname = %q, erwartet %q", vereinsname, "TSV Musterstadt e.V.")
	}

	// Offene Kassensitzung mit z_nr=3.
	var status string
	if err := db.QueryRow("SELECT status FROM kassensitzungen WHERE z_nr = 3").Scan(&status); err != nil {
		t.Fatalf("Kassensitzung 3 abfragen: %v", err)
	}
	if status != "offen" {
		t.Errorf("Kassensitzung 3 Status = %q, erwartet offen", status)
	}

	// Kassenjournal enthält Events.
	var eventCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM kassenjournal").Scan(&eventCount); err != nil {
		t.Fatalf("Kassenjournal zählen: %v", err)
	}
	if eventCount == 0 {
		t.Fatal("Kassenjournal ist nach dem Seeding leer")
	}

	// Tisch-Session-Projektion: Tisch 1 ist vollständig bezahlt (Saldo 0).
	var saldo int
	if err := db.QueryRow("SELECT saldo_cents FROM tisch_sessions WHERE kassensitzung_nr = 3 AND tisch_id = 1").Scan(&saldo); err != nil {
		t.Fatalf("Tisch-Session Tisch 1 abfragen: %v", err)
	}
	if saldo != 0 {
		t.Errorf("Tisch 1 Saldo = %d, erwartet 0 (bezahlt)", saldo)
	}

	// --- Zweiter Lauf: Guard greift, ohne etwas zu schreiben ---
	if err := Run(ctx, db); err == nil {
		t.Fatal("zweiter Run sollte am Guard scheitern, lieferte aber keinen Fehler")
	}

	var eventCountNachher int
	if err := db.QueryRow("SELECT COUNT(*) FROM kassenjournal").Scan(&eventCountNachher); err != nil {
		t.Fatalf("Kassenjournal erneut zählen: %v", err)
	}
	if eventCountNachher != eventCount {
		t.Errorf("Guard hat geschrieben: %d Events vorher, %d nachher", eventCount, eventCountNachher)
	}
}
