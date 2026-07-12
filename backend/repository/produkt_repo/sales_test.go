//go:build integration

package produkt_repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	dbpkg "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/produkt"
)

// setupSales prepares an isolated environment for the hatVerkaeufe projection
// tests: it clears products, variants and the journal (plus its FK dependencies)
// and returns the repo together with a freshly created user and Kassensitzung so
// journal events can be inserted.
func setupSales(t *testing.T) (Repository, *sql.DB, int, int, func()) {
	t.Helper()
	db := dbpkg.OpenTestDatabase()

	cleanSales(t, db)

	var userID int
	if err := db.QueryRow(
		"INSERT INTO users (name, username, role, status, password_hash, onetime_password_hash, created_at, updated_at) VALUES ('nico', 'nico', 'admin', 'active', 'hash', 'hash', now(), now()) RETURNING id",
	).Scan(&userID); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	var zNr int
	if err := db.QueryRow(
		"INSERT INTO kassensitzungen (datum, bezeichnung, status, created_at, updated_at) VALUES ((NOW() AT TIME ZONE 'Europe/Berlin')::date, 'Test-Sitzung', 'offen', NOW(), NOW()) RETURNING z_nr",
	).Scan(&zNr); err != nil {
		t.Fatalf("Failed to create kassensitzung: %v", err)
	}

	return NewRepository(db), db, userID, zNr, func() {
		cleanSales(t, db)
		_ = db.Close()
	}
}

func cleanSales(t *testing.T, db *sql.DB) {
	t.Helper()
	if _, err := db.Exec("DELETE FROM tisch_sessions"); err != nil {
		t.Fatalf("Failed to clean tisch_sessions: %v", err)
	}
	if _, err := db.Exec("ALTER TABLE kassenjournal DISABLE TRIGGER kassenjournal_no_delete"); err != nil {
		t.Fatalf("Failed to disable kassenjournal trigger: %v", err)
	}
	if _, err := db.Exec("DELETE FROM kassenjournal"); err != nil {
		t.Fatalf("Failed to clean kassenjournal: %v", err)
	}
	if _, err := db.Exec("ALTER TABLE kassenjournal ENABLE TRIGGER kassenjournal_no_delete"); err != nil {
		t.Fatalf("Failed to enable kassenjournal trigger: %v", err)
	}
	if _, err := db.Exec("DELETE FROM kassensitzungen"); err != nil {
		t.Fatalf("Failed to clean kassensitzungen: %v", err)
	}
	if _, err := db.Exec("DELETE FROM produkt_varianten"); err != nil {
		t.Fatalf("Failed to clean produkt_varianten: %v", err)
	}
	if _, err := db.Exec("DELETE FROM produkte"); err != nil {
		t.Fatalf("Failed to clean produkte: %v", err)
	}
	if _, err := db.Exec("DELETE FROM users"); err != nil {
		t.Fatalf("Failed to clean users: %v", err)
	}
}

func insertJournalEvent(t *testing.T, db *sql.DB, userID int, eventType, subject string, data map[string]any, ksNr int) {
	t.Helper()
	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal event data: %v", err)
	}
	_, err = db.Exec(
		"INSERT INTO kassenjournal (user_id, user_name, type, subject, version, data, timestamp, kassensitzung_nr) VALUES ($1, 'nico', $2, $3, 1, $4, $5, $6)",
		userID, eventType, subject, raw, time.Now().UTC(), ksNr,
	)
	if err != nil {
		t.Fatalf("Failed to insert %s event: %v", eventType, err)
	}
}

func bestellungWithVariante(varianteID int) map[string]any {
	return map[string]any{
		"bestellungId": "b0000000-0000-0000-0000-000000000001",
		"positionen": []map[string]any{{
			"positionId":       "p0000000-0000-0000-0000-000000000001",
			"varianteId":       varianteID,
			"produktName":      "Bier",
			"varianteName":     "0.5L",
			"kategorie":        "getraenk",
			"einzelpreisCents": 350,
			"menge":            1,
		}},
		"gesamtPreisCents": 350,
		"kommentar":        "",
	}
}

func direktverkaufWithVariante(varianteID int) map[string]any {
	return map[string]any{
		"verkaufId":         "v0000000-0000-0000-0000-000000000001",
		"gesamtbetragCents": 350,
		"positionen": []map[string]any{{
			"positionId":       "d0000000-0000-0000-0000-000000000001",
			"varianteId":       varianteID,
			"produktName":      "Bier",
			"varianteName":     "0.5L",
			"kategorie":        "getraenk",
			"steuersatz":       "regel",
			"einzelpreisCents": 350,
			"menge":            1,
		}},
		"kommentar": "",
	}
}

func TestGetProduktIDsMitVerkaeufen(t *testing.T) {
	repo, db, userID, zNr, teardown := setupSales(t)
	defer teardown()
	ctx := context.Background()

	// Produkt A: verkauft über eine Bestellung.
	produktA, _ := repo.CreateProduct(ctx, newProduct("Bier", produkt.GetraenkKategorie))
	varianteA, _ := repo.CreateVariant(ctx, produktA, newVariant("0.5L", 350, produkt.ActiveStatus))
	insertJournalEvent(t, db, userID, "bestellung-aufgenommen:v1", "b/tisch-1/bestellung-1", bestellungWithVariante(varianteA), zNr)

	// Produkt B: verkauft über einen Direktverkauf.
	produktB, _ := repo.CreateProduct(ctx, newProduct("Wein", produkt.GetraenkKategorie))
	varianteB, _ := repo.CreateVariant(ctx, produktB, newVariant("0.2L", 500, produkt.ActiveStatus))
	insertJournalEvent(t, db, userID, "direktverkauf-getaetigt:v1", "d/verkauf-1", direktverkaufWithVariante(varianteB), zNr)

	// Produkt C: angelegt, aber nie verkauft.
	produktC, _ := repo.CreateProduct(ctx, newProduct("Brezel", produkt.EssenKategorie))
	_, _ = repo.CreateVariant(ctx, produktC, newVariant("Standard", 200, produkt.ActiveStatus))

	verkaufte, err := repo.GetProduktIDsMitVerkaeufen(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !verkaufte[produktA] {
		t.Errorf("Produkt A (Bestellung) should be marked as sold")
	}
	if !verkaufte[produktB] {
		t.Errorf("Produkt B (Direktverkauf) should be marked as sold")
	}
	if verkaufte[produktC] {
		t.Errorf("Produkt C (no sales) must not be marked as sold")
	}
}

func TestProduktHatVerkaeufe(t *testing.T) {
	repo, db, userID, zNr, teardown := setupSales(t)
	defer teardown()
	ctx := context.Background()

	verkauft, _ := repo.CreateProduct(ctx, newProduct("Bier", produkt.GetraenkKategorie))
	variante, _ := repo.CreateVariant(ctx, verkauft, newVariant("0.5L", 350, produkt.ActiveStatus))
	insertJournalEvent(t, db, userID, "bestellung-aufgenommen:v1", "b/tisch-1/bestellung-1", bestellungWithVariante(variante), zNr)

	nichtVerkauft, _ := repo.CreateProduct(ctx, newProduct("Brezel", produkt.EssenKategorie))
	_, _ = repo.CreateVariant(ctx, nichtVerkauft, newVariant("Standard", 200, produkt.ActiveStatus))

	hat, err := repo.ProduktHatVerkaeufe(ctx, verkauft)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !hat {
		t.Errorf("Product with a Bestellung should report hatVerkaeufe=true")
	}

	hat, err = repo.ProduktHatVerkaeufe(ctx, nichtVerkauft)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if hat {
		t.Errorf("Product without sales should report hatVerkaeufe=false")
	}
}
