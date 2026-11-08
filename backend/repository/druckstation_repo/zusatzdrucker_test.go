//go:build integration

package druckstation_repo

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nicograef/jotti/backend/db/dbtest"
	"github.com/nicograef/jotti/backend/domain/druckstation"
)

type zusatzdruckerFixture struct {
	repo       Repository
	db         *sql.DB
	produktIDs []int
}

// setupZusatzdrucker removes every Zusatzdrucker and every Produkt the test creates: all
// integration packages share one database.
func setupZusatzdrucker(t *testing.T) *zusatzdruckerFixture {
	t.Helper()
	database := dbtest.Open()
	f := &zusatzdruckerFixture{repo: NewRepository(database), db: database}

	t.Cleanup(func() {
		if _, err := database.Exec("DELETE FROM zusatzdrucker"); err != nil {
			t.Errorf("Failed to delete zusatzdrucker: %v", err)
		}
		for _, id := range f.produktIDs {
			if _, err := database.Exec("DELETE FROM produkt_varianten WHERE produkt_id = $1", id); err != nil {
				t.Errorf("Failed to delete varianten: %v", err)
			}
			if _, err := database.Exec("DELETE FROM produkte WHERE id = $1", id); err != nil {
				t.Errorf("Failed to delete produkt: %v", err)
			}
		}
		_ = database.Close()
	})

	return f
}

// createProdukt inserts an active Produkt with the given Varianten; it returns the Produkt-ID, then the Variante-IDs.
func (f *zusatzdruckerFixture) createProdukt(t *testing.T, name string, varianten ...string) []int {
	t.Helper()
	var produktID int
	if err := f.db.QueryRow(
		"INSERT INTO produkte (name, kategorie, steuersatz, status, created_at, updated_at) VALUES ($1, 'essen', 'regel', 'active', now(), now()) RETURNING id",
		name,
	).Scan(&produktID); err != nil {
		t.Fatalf("Failed to insert produkt: %v", err)
	}
	f.produktIDs = append(f.produktIDs, produktID)

	ids := []int{produktID}
	for _, v := range varianten {
		var varianteID int
		if err := f.db.QueryRow(
			"INSERT INTO produkt_varianten (produkt_id, name, preis_cents, status, created_at, updated_at) VALUES ($1, $2, 300, 'active', now(), now()) RETURNING id",
			produktID, v,
		).Scan(&varianteID); err != nil {
			t.Fatalf("Failed to insert variante: %v", err)
		}
		ids = append(ids, varianteID)
	}
	return ids
}

func TestGetZusatzdruckerJeVariante_Zusatzdrucker(t *testing.T) {
	f := setupZusatzdrucker(t)
	repo := f.repo
	ctx := context.Background()

	kuchen := f.createProdukt(t, "Zusatzdrucker-Testkuchen", "Stück", "Blech")
	wurst := f.createProdukt(t, "Zusatzdrucker-Testwurst", "Rot")

	id, err := repo.CreateZusatzdrucker(ctx, druckstation.Zusatzdrucker{Name: "Kaffee & Kuchen", Bonmodus: druckstation.BonmodusProBestellung})
	if err != nil {
		t.Fatalf("CreateZusatzdrucker: %v", err)
	}
	if err := repo.AssignProdukt(ctx, kuchen[0], id); err != nil {
		t.Fatalf("AssignProdukt: %v", err)
	}

	got, err := repo.GetZusatzdruckerJeVariante(ctx, []int{kuchen[1], kuchen[2], wurst[1]})
	if err != nil {
		t.Fatalf("GetZusatzdruckerJeVariante: %v", err)
	}
	// Ohne IP trotzdem enthalten: Ein Zusatzdrucker ohne IP druckt nichts statt an der Kategorie.
	want := druckstation.Zusatzdrucker{ID: id, Name: "Kaffee & Kuchen", DruckerIP: "", Bonmodus: druckstation.BonmodusProBestellung}
	if len(got) != 2 {
		t.Fatalf("GetZusatzdruckerJeVariante returned %d entries, want 2: %+v", len(got), got)
	}
	for _, varianteID := range kuchen[1:] {
		if got[varianteID] != want {
			t.Errorf("Variante %d: got %+v, want %+v", varianteID, got[varianteID], want)
		}
	}
	if _, ok := got[wurst[1]]; ok {
		t.Errorf("Variante %d of an unassigned Produkt has an entry", wurst[1])
	}

	if err := repo.DeleteZusatzdrucker(ctx, id); err != nil {
		t.Fatalf("DeleteZusatzdrucker: %v", err)
	}
	got, err = repo.GetZusatzdruckerJeVariante(ctx, []int{kuchen[1], kuchen[2], wurst[1]})
	if err != nil {
		t.Fatalf("GetZusatzdruckerJeVariante after delete: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("GetZusatzdruckerJeVariante after delete = %+v, want no entry", got)
	}
}
