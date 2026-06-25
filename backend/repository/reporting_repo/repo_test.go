//go:build integration

package reporting_repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	dbpkg "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/reporting"
)

func cleanDB(t *testing.T, db *sql.DB) {
	t.Helper()
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

func createUser(t *testing.T, db *sql.DB, name, username, status string) int {
	t.Helper()
	var id int
	err := db.QueryRow(
		"INSERT INTO users (name, username, role, status, password_hash, onetime_password_hash, created_at, updated_at) VALUES ($1, $2, 'service', $3, 'hash', 'hash', now(), now()) RETURNING id",
		name, username, status,
	).Scan(&id)
	if err != nil {
		t.Fatalf("Failed to insert user %q: %v", username, err)
	}
	return id
}

func createKassensitzung(t *testing.T, db *sql.DB) int {
	t.Helper()
	var zNr int
	err := db.QueryRow(
		"INSERT INTO kassensitzungen (datum, bezeichnung, status, created_at, updated_at) VALUES ((NOW() AT TIME ZONE 'Europe/Berlin')::date, 'Test-Sitzung', 'offen', NOW(), NOW()) RETURNING z_nr",
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

func zahlungData(gesamtCents int) map[string]any {
	return map[string]any{
		"zahlungId": "z0000000-0000-0000-0000-000000000001",
		"positionen": []map[string]any{{
			"positionId":   "p0000000-0000-0000-0000-000000000001",
			"produktName":  "Bier",
			"varianteName": "0.5L",
			"steuersatz":   "regel",
			"einzelpreis":  gesamtCents,
			"menge":        1,
		}},
		"gesamtZahlungCents": gesamtCents,
		"kommentar":          "",
	}
}

func stornierungData(betragCents int, kommentar string) map[string]any {
	return map[string]any{
		"stornierungId":          "s0000000-0000-0000-0000-000000000001",
		"zahlungId":              "z0000000-0000-0000-0000-000000000001",
		"gesamtStornierungCents": betragCents,
		"kommentar":              kommentar,
		"positionen": []map[string]any{{
			"produktName":  "Bier",
			"varianteName": "0.5L",
			"menge":        1,
			"einzelpreis":  betragCents,
		}},
	}
}

func korrekturData(betragCents int, kommentar string) map[string]any {
	return map[string]any{
		"korrekturId": "k0000000-0000-0000-0000-000000000001",
		"gesamtCents": betragCents,
		"kommentar":   kommentar,
		"positionen": []map[string]any{{
			"produktName":  "Limo",
			"varianteName": "0.3L",
			"menge":        1,
			"einzelpreis":  betragCents,
		}},
	}
}

// TestGetReporting_ResolvesKlarnameIncludingSoftDeleted verifies that the live LEFT JOIN
// resolves the current Klarname for both active and soft-deleted users, while the frozen
// username stays the maßgebliche identity in the event rows.
func TestGetReporting_ResolvesKlarnameIncludingSoftDeleted(t *testing.T) {
	db := dbpkg.OpenTestDatabase()
	defer db.Close()
	cleanDB(t, db)
	defer cleanDB(t, db)

	repo := NewRepository(db)
	ctx := context.Background()

	annaID := createUser(t, db, "Anna Müller", "anna", "active")
	bobID := createUser(t, db, "Bob Schmidt", "bob", "deleted")
	ksNr := createKassensitzung(t, db)

	// Anna gets more revenue so she sorts first (ORDER BY zahlungen_cents DESC).
	insertEvent(t, db, annaID, "anna", "zahlung-kassiert:v1", "kassensitzung-1/tisch-1", 1, zahlungData(2000), ksNr)
	insertEvent(t, db, bobID, "bob", "zahlung-kassiert:v1", "kassensitzung-1/tisch-2", 1, zahlungData(1000), ksNr)

	insertEvent(t, db, annaID, "anna", "stornierung-erteilt:v1", "kassensitzung-1/tisch-1", 2, stornierungData(500, "Anna storniert"), ksNr)
	insertEvent(t, db, bobID, "bob", "stornierung-erteilt:v1", "kassensitzung-1/tisch-2", 2, stornierungData(300, "Bob storniert"), ksNr)

	data, err := repo.GetReporting(ctx, ksNr)
	if err != nil {
		t.Fatalf("GetReporting failed: %v", err)
	}

	// Umsatz pro Servicekraft: username frozen in the event, Klarname resolved live.
	klarnameByUsername := map[string]string{}
	for _, sk := range data.Breakdowns.UmsatzProServicekraft {
		klarnameByUsername[sk.UserName] = sk.Name
	}
	if got := klarnameByUsername["anna"]; got != "Anna Müller" {
		t.Errorf("expected anna Klarname 'Anna Müller', got %q", got)
	}
	if got := klarnameByUsername["bob"]; got != "Bob Schmidt" {
		t.Errorf("expected soft-deleted bob Klarname 'Bob Schmidt', got %q", got)
	}

	// Stornierungen: same resolution including the soft-deleted user.
	stornoKlarnameByUsername := map[string]string{}
	for _, s := range data.Stornierungen {
		stornoKlarnameByUsername[s.UserName] = s.Name
	}
	if got := stornoKlarnameByUsername["anna"]; got != "Anna Müller" {
		t.Errorf("expected anna storno Klarname 'Anna Müller', got %q", got)
	}
	if got := stornoKlarnameByUsername["bob"]; got != "Bob Schmidt" {
		t.Errorf("expected soft-deleted bob storno Klarname 'Bob Schmidt', got %q", got)
	}
}

// TestGetReporting_IncludesBeideStornoArten verifies that the Stornierungsliste and the
// Stornoquote count both storno kinds: the cash-relevant Warenrücknahme (stornierung-erteilt,
// marked as Bar-Rückgabe) and the geldneutral Korrektur (bestellung-korrigiert).
func TestGetReporting_IncludesBeideStornoArten(t *testing.T) {
	db := dbpkg.OpenTestDatabase()
	defer db.Close()
	cleanDB(t, db)
	defer cleanDB(t, db)

	repo := NewRepository(db)
	ctx := context.Background()

	userID := createUser(t, db, "Anna Müller", "anna", "active")
	ksNr := createKassensitzung(t, db)

	insertEvent(t, db, userID, "anna", "zahlung-kassiert:v1", "kassensitzung-1/tisch-1", 1, zahlungData(2000), ksNr)
	insertEvent(t, db, userID, "anna", "stornierung-erteilt:v1", "kassensitzung-1/tisch-1", 2, stornierungData(500, "Warenruecknahme"), ksNr)
	insertEvent(t, db, userID, "anna", "bestellung-korrigiert:v1", "kassensitzung-1/tisch-1", 3, korrekturData(300, "Korrektur"), ksNr)

	data, err := repo.GetReporting(ctx, ksNr)
	if err != nil {
		t.Fatalf("GetReporting failed: %v", err)
	}

	if len(data.Stornierungen) != 2 {
		t.Fatalf("expected 2 Stornierungen (Warenrücknahme + Korrektur), got %d", len(data.Stornierungen))
	}

	byKommentar := map[string]reporting.StornierungDetail{}
	for _, s := range data.Stornierungen {
		byKommentar[s.Kommentar] = s
	}

	warenruecknahme, ok := byKommentar["Warenruecknahme"]
	if !ok {
		t.Fatal("expected a Warenrücknahme entry")
	}
	if !warenruecknahme.BarRueckgabe {
		t.Error("expected Warenrücknahme to be marked as Bar-Rückgabe")
	}
	if warenruecknahme.BetragCents != 500 {
		t.Errorf("expected Warenrücknahme betrag 500, got %d", warenruecknahme.BetragCents)
	}

	korrektur, ok := byKommentar["Korrektur"]
	if !ok {
		t.Fatal("expected a Korrektur entry")
	}
	if korrektur.BarRueckgabe {
		t.Error("expected geldneutrale Korrektur to NOT be marked as Bar-Rückgabe")
	}
	if korrektur.BetragCents != 300 {
		t.Errorf("expected Korrektur betrag 300 (aus gesamtCents), got %d", korrektur.BetragCents)
	}

	// Die Stornoquote/Summe zählt beide Arten (500 + 300) und beide Events.
	if data.Summary.GesamtStornierungenCents != 800 {
		t.Errorf("expected gesamt stornierungen 800 (beide Arten), got %d", data.Summary.GesamtStornierungenCents)
	}
	if data.Summary.AnzahlStornierungen != 2 {
		t.Errorf("expected anzahl stornierungen 2, got %d", data.Summary.AnzahlStornierungen)
	}
}
