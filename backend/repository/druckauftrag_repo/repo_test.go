//go:build integration

package druckauftrag_repo

import (
	"context"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	dbpkg "github.com/nicograef/jotti/backend/db"
)

func setup(t *testing.T) (Repository, func(t *testing.T)) {
	t.Helper()
	database := dbpkg.OpenTestDatabase()

	_, err := database.Exec("DELETE FROM druckauftraege")
	if err != nil {
		t.Fatalf("Failed to reset druckauftraege: %v", err)
	}

	return NewRepository(database), func(t *testing.T) {
		_, err := database.Exec("DELETE FROM druckauftraege")
		if err != nil {
			t.Fatalf("Failed to reset druckauftraege: %v", err)
		}
		database.Close()
	}
}

func TestEnqueueAndGetOffeneDruckauftraege(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	err := repo.EnqueueDruckauftraege(context.Background(), []NeuerDruckauftrag{
		{
			ZielIP:   "192.168.1.51",
			Payload:  "AAA=",
			BonArt:   "arbeitsbon",
			Referenz: "bestellung-aufgenommen:1",
		},
		{
			ZielIP:   "192.168.1.52",
			Payload:  "BBB=",
			BonArt:   "arbeitsbon",
			Referenz: "bestellung-aufgenommen:2",
		},
	})
	if err != nil {
		t.Fatalf("Expected no enqueue error, got %v", err)
	}

	offene, err := repo.GetOffeneDruckauftraege(context.Background())
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(offene) != 2 {
		t.Fatalf("Expected 2 offene auftraege, got %d", len(offene))
	}
	if offene[0].ID == 0 || offene[1].ID == 0 {
		t.Fatalf("Expected generated IDs, got %+v", offene)
	}
	if offene[0].ZielIP != "192.168.1.51" || offene[1].ZielIP != "192.168.1.52" {
		t.Fatalf("Unexpected order or ziel_ip: %+v", offene)
	}
}

func TestQuittiereGedruckteAuftraege(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	err := repo.EnqueueDruckauftraege(context.Background(), []NeuerDruckauftrag{
		{
			ZielIP:   "192.168.1.51",
			Payload:  "AAA=",
			BonArt:   "arbeitsbon",
			Referenz: "bestellung-aufgenommen:1",
		},
	})
	if err != nil {
		t.Fatalf("Expected no enqueue error, got %v", err)
	}

	offene, err := repo.GetOffeneDruckauftraege(context.Background())
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(offene) != 1 {
		t.Fatalf("Expected 1 offener auftrag, got %d", len(offene))
	}

	err = repo.QuittiereGedruckteAuftraege(context.Background(), []int{offene[0].ID})
	if err != nil {
		t.Fatalf("Expected no quittieren error, got %v", err)
	}

	offene, err = repo.GetOffeneDruckauftraege(context.Background())
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(offene) != 0 {
		t.Fatalf("Expected no offene auftraege after quittieren, got %d", len(offene))
	}
}
