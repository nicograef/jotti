//go:build integration

package drucker_repo

import (
	"context"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	dbpkg "github.com/nicograef/jotti/backend/db"
)

func setup(t *testing.T) (Repository, func(t *testing.T)) {
	t.Helper()
	db := dbpkg.OpenTestDatabase()

	// Reset to default state (leer drucker_ip, pro_position)
	_, err := db.Exec("UPDATE kategorie_drucker SET drucker_ip = '', bonmodus = 'pro_position'")
	if err != nil {
		t.Fatalf("Failed to reset kategorie_drucker: %v", err)
	}

	return NewRepository(db), func(t *testing.T) {
		_, err := db.Exec("UPDATE kategorie_drucker SET drucker_ip = '', bonmodus = 'pro_position'")
		if err != nil {
			t.Fatalf("Failed to reset kategorie_drucker: %v", err)
		}
		db.Close()
	}
}

func TestGetAlleKategorieDrucker(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	konfigs, err := repo.GetAlleKategorieDrucker(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(konfigs) != 3 {
		t.Fatalf("Expected 3 Kategorien, got %d", len(konfigs))
	}
}

func TestGetKonfigurierteKategorieDrucker_Leer(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	result, err := repo.GetKonfigurierteKategorieDrucker(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("Expected 0 konfigurierte Drucker, got %d", len(result))
	}
}

func TestUpsertKategorieDrucker(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()

	err := repo.UpsertKategorieDrucker(ctx, "essen", "192.168.1.51", "pro_position")
	if err != nil {
		t.Fatalf("Expected no error on upsert, got %v", err)
	}

	result, err := repo.GetKonfigurierteKategorieDrucker(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("Expected 1 konfigurierter Drucker, got %d", len(result))
	}
	konfig, ok := result["essen"]
	if !ok {
		t.Fatal("Expected 'essen' to be in result")
	}
	if konfig.DruckerIP != "192.168.1.51" {
		t.Errorf("Expected DruckerIP '192.168.1.51', got %q", konfig.DruckerIP)
	}
	if konfig.Bonmodus != "pro_position" {
		t.Errorf("Expected Bonmodus 'pro_position', got %q", konfig.Bonmodus)
	}
}

func TestUpsertKategorieDrucker_Update(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()

	_ = repo.UpsertKategorieDrucker(ctx, "essen", "192.168.1.51", "pro_position")
	err := repo.UpsertKategorieDrucker(ctx, "essen", "192.168.1.99", "pro_bestellung")
	if err != nil {
		t.Fatalf("Expected no error on second upsert, got %v", err)
	}

	result, err := repo.GetKonfigurierteKategorieDrucker(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	konfig := result["essen"]
	if konfig.DruckerIP != "192.168.1.99" {
		t.Errorf("Expected updated DruckerIP '192.168.1.99', got %q", konfig.DruckerIP)
	}
	if konfig.Bonmodus != "pro_bestellung" {
		t.Errorf("Expected updated Bonmodus 'pro_bestellung', got %q", konfig.Bonmodus)
	}
}

func TestUpsertKategorieDrucker_Deaktivieren(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()

	_ = repo.UpsertKategorieDrucker(ctx, "getraenk", "192.168.1.50", "pro_position")
	err := repo.UpsertKategorieDrucker(ctx, "getraenk", "", "pro_position")
	if err != nil {
		t.Fatalf("Expected no error when deactivating, got %v", err)
	}

	result, err := repo.GetKonfigurierteKategorieDrucker(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if _, ok := result["getraenk"]; ok {
		t.Fatal("Expected 'getraenk' to NOT be in konfigurierte result after deactivation")
	}
}
