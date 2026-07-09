//go:build integration

package druckstation_repo

import (
	"context"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	dbpkg "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/druckstation"
)

func setup(t *testing.T) (Repository, func(t *testing.T)) {
	t.Helper()
	db := dbpkg.OpenTestDatabase()

	reset := func() {
		// Reset to default state: leere drucker_ip; Bonmodus pro_position für
		// Produktkategorien, pro_bestellung für abholbon, NULL für kassenbeleg
		// (CHECK-Constraint).
		if _, err := db.Exec("UPDATE druckstationen SET drucker_ip = '', bonmodus = 'pro_position' WHERE kategorie IN ('essen', 'getraenk', 'sonstiges')"); err != nil {
			t.Fatalf("Failed to reset produktstationen: %v", err)
		}
		if _, err := db.Exec("UPDATE druckstationen SET drucker_ip = '', bonmodus = 'pro_bestellung' WHERE kategorie = 'abholbon'"); err != nil {
			t.Fatalf("Failed to reset abholbon: %v", err)
		}
		if _, err := db.Exec("UPDATE druckstationen SET drucker_ip = '', bonmodus = NULL WHERE kategorie = 'kassenbeleg'"); err != nil {
			t.Fatalf("Failed to reset kassenbeleg: %v", err)
		}
	}

	reset()

	return NewRepository(db), func(t *testing.T) {
		reset()
		_ = db.Close()
	}
}

func TestGetAlleDruckstationen(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	konfigs, err := repo.GetAlleDruckstationen(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(konfigs) != 5 {
		t.Fatalf("Expected 5 Kategorien, got %d", len(konfigs))
	}
}

func TestGetKonfigurierteDruckstationen_Leer(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	result, err := repo.GetKonfigurierteDruckstationen(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("Expected 0 konfigurierte Drucker, got %d", len(result))
	}
}

func TestGetAlleDruckstationen_StabileReihenfolge(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()

	// Upsert auf eine mittlere Kategorie darf die Reihenfolge nicht verschieben.
	err := repo.UpsertDruckstation(ctx, druckstation.Druckstation{
		Kategorie: druckstation.KategorieSonstiges,
		DruckerIP: "192.168.1.50",
		Bonmodus:  druckstation.BonmodusProPosition,
	})
	if err != nil {
		t.Fatalf("Expected no error on upsert, got %v", err)
	}

	result, err := repo.GetAlleDruckstationen(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// ORDER BY kategorie folgt der Enum-Deklaration, nicht dem Alphabet.
	want := []druckstation.Kategorie{
		druckstation.KategorieEssen,
		druckstation.KategorieGetraenk,
		druckstation.KategorieSonstiges,
		druckstation.KategorieKassenbeleg,
		druckstation.KategorieAbholbon,
	}
	if len(result) != len(want) {
		t.Fatalf("Expected %d Kategorien, got %d", len(want), len(result))
	}
	for i, kategorie := range want {
		if result[i].Kategorie != kategorie {
			t.Errorf("Position %d: expected %q, got %q", i, kategorie, result[i].Kategorie)
		}
	}
}

func TestUpsertDruckstation(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()

	err := repo.UpsertDruckstation(ctx, druckstation.Druckstation{
		Kategorie: druckstation.KategorieEssen,
		DruckerIP: "192.168.1.51",
		Bonmodus:  druckstation.BonmodusProPosition,
	})
	if err != nil {
		t.Fatalf("Expected no error on upsert, got %v", err)
	}

	result, err := repo.GetKonfigurierteDruckstationen(ctx)
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
	if konfig.Bonmodus != druckstation.BonmodusProPosition {
		t.Errorf("Expected Bonmodus 'pro_position', got %q", konfig.Bonmodus)
	}
}

func TestUpsertDruckstation_Update(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()

	_ = repo.UpsertDruckstation(ctx, druckstation.Druckstation{
		Kategorie: druckstation.KategorieEssen,
		DruckerIP: "192.168.1.51",
		Bonmodus:  druckstation.BonmodusProPosition,
	})
	err := repo.UpsertDruckstation(ctx, druckstation.Druckstation{
		Kategorie: druckstation.KategorieEssen,
		DruckerIP: "192.168.1.99",
		Bonmodus:  druckstation.BonmodusProBestellung,
	})
	if err != nil {
		t.Fatalf("Expected no error on second upsert, got %v", err)
	}

	result, err := repo.GetKonfigurierteDruckstationen(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	konfig := result["essen"]
	if konfig.DruckerIP != "192.168.1.99" {
		t.Errorf("Expected updated DruckerIP '192.168.1.99', got %q", konfig.DruckerIP)
	}
	if konfig.Bonmodus != druckstation.BonmodusProBestellung {
		t.Errorf("Expected updated Bonmodus 'pro_bestellung', got %q", konfig.Bonmodus)
	}
}

func TestUpsertDruckstation_Deaktivieren(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()

	_ = repo.UpsertDruckstation(ctx, druckstation.Druckstation{
		Kategorie: druckstation.KategorieGetraenk,
		DruckerIP: "192.168.1.50",
		Bonmodus:  druckstation.BonmodusProPosition,
	})
	err := repo.UpsertDruckstation(ctx, druckstation.Druckstation{
		Kategorie: druckstation.KategorieGetraenk,
		DruckerIP: "",
		Bonmodus:  druckstation.BonmodusProPosition,
	})
	if err != nil {
		t.Fatalf("Expected no error when deactivating, got %v", err)
	}

	result, err := repo.GetKonfigurierteDruckstationen(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if _, ok := result["getraenk"]; ok {
		t.Fatal("Expected 'getraenk' to NOT be in konfigurierte result after deactivation")
	}
}

func TestUpsertDruckstation_Kassenbeleg(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()

	// kassenbeleg trägt keinen Bonmodus (NULL); das muss sauber durchlaufen.
	err := repo.UpsertDruckstation(ctx, druckstation.Druckstation{
		Kategorie: druckstation.KategorieKassenbeleg,
		DruckerIP: "192.168.1.60",
		Bonmodus:  "",
	})
	if err != nil {
		t.Fatalf("Expected no error on kassenbeleg upsert, got %v", err)
	}

	result, err := repo.GetKonfigurierteDruckstationen(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	konfig, ok := result["kassenbeleg"]
	if !ok {
		t.Fatal("Expected 'kassenbeleg' to be in result")
	}
	if konfig.DruckerIP != "192.168.1.60" {
		t.Errorf("Expected DruckerIP '192.168.1.60', got %q", konfig.DruckerIP)
	}
	if konfig.Bonmodus != "" {
		t.Errorf("Expected empty Bonmodus for kassenbeleg, got %q", konfig.Bonmodus)
	}
}

func TestUpsertDruckstation_Abholbon(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()

	// abholbon trägt einen Bonmodus wie eine Produktstation.
	err := repo.UpsertDruckstation(ctx, druckstation.Druckstation{
		Kategorie: druckstation.KategorieAbholbon,
		DruckerIP: "192.168.1.70",
		Bonmodus:  druckstation.BonmodusProPosition,
	})
	if err != nil {
		t.Fatalf("Expected no error on abholbon upsert, got %v", err)
	}

	result, err := repo.GetKonfigurierteDruckstationen(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	konfig, ok := result["abholbon"]
	if !ok {
		t.Fatal("Expected 'abholbon' to be in result")
	}
	if konfig.Bonmodus != druckstation.BonmodusProPosition {
		t.Errorf("Expected Bonmodus 'pro_position' for abholbon, got %q", konfig.Bonmodus)
	}
}
