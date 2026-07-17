//go:build integration

package produkt_repo

import (
	"context"
	"errors"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	dbpkg "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/domain/steuer"
)

func setup(t *testing.T) (Repository, func(t *testing.T)) {
	db := dbpkg.OpenTestDatabase()

	// Clean up in correct order due to foreign key constraints
	_, err := db.Exec("DELETE FROM produkt_varianten")
	if err != nil {
		t.Fatalf("Failed to clean produkt_varianten table: %v", err)
	}
	_, err = db.Exec("DELETE FROM produkte")
	if err != nil {
		t.Fatalf("Failed to clean produkte table: %v", err)
	}

	return NewRepository(db), func(t *testing.T) {
		_, err = db.Exec("DELETE FROM produkt_varianten")
		if err != nil {
			t.Fatalf("Failed to clean produkt_varianten table: %v", err)
		}
		_, err = db.Exec("DELETE FROM produkte")
		if err != nil {
			t.Fatalf("Failed to clean produkte table: %v", err)
		}

		_ = db.Close()
	}
}

func newProdukt(name string, kategorie produkt.Kategorie) produkt.Produkt {
	now := time.Now().UTC()
	steuersatz := steuer.RegelSteuersatz
	if kategorie == produkt.EssenKategorie {
		steuersatz = steuer.ErmaessigtSteuersatz
	}
	if kategorie == produkt.SonstigesKategorie {
		steuersatz = steuer.BefreitSteuersatz
	}

	return produkt.Produkt{
		Name:       name,
		Kategorie:  kategorie,
		Steuersatz: steuersatz,
		Status:     produkt.ActiveStatus,
		Varianten:  []produkt.Variante{},
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func newVariante(name string, preisCents int, status produkt.Status) produkt.Variante {
	now := time.Now().UTC()
	return produkt.Variante{
		Name:       name,
		PreisCents: preisCents,
		Status:     status,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func TestGetAllProdukte(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	_, _ = repo.CreateProdukt(ctx, newProdukt("Produkt 1", produkt.EssenKategorie))
	_, _ = repo.CreateProdukt(ctx, newProdukt("Produkt 2", produkt.GetraenkKategorie))

	produkte, err := repo.GetAllProdukte(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(produkte) != 2 {
		t.Fatalf("Expected 2 produkte, got %d", len(produkte))
	}
}

func TestGetAllProdukte_WithVarianten(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	produktID, _ := repo.CreateProdukt(ctx, newProdukt("Pizza", produkt.EssenKategorie))
	_, _ = repo.CreateVariante(ctx, produktID, newVariante("Small", 899, produkt.ActiveStatus))
	_, _ = repo.CreateVariante(ctx, produktID, newVariante("Large", 1299, produkt.ActiveStatus))

	produkte, err := repo.GetAllProdukte(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(produkte) != 1 {
		t.Fatalf("Expected 1 produkt, got %d", len(produkte))
	}
	if len(produkte[0].Varianten) != 2 {
		t.Fatalf("Expected 2 varianten, got %d", len(produkte[0].Varianten))
	}
}

func TestGetActiveProdukte(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()

	// Produkt with active variante
	produkt1ID, _ := repo.CreateProdukt(ctx, newProdukt("Produkt 1", produkt.EssenKategorie))
	_, _ = repo.CreateVariante(ctx, produkt1ID, newVariante("Regular", 999, produkt.ActiveStatus))

	// Produkt with only inactive variante
	produkt2ID, _ := repo.CreateProdukt(ctx, newProdukt("Produkt 2", produkt.EssenKategorie))
	_, _ = repo.CreateVariante(ctx, produkt2ID, newVariante("Regular", 999, produkt.InactiveStatus))

	// Produkt with no varianten
	_, _ = repo.CreateProdukt(ctx, newProdukt("Produkt 3", produkt.EssenKategorie))

	produkte, err := repo.GetActiveProdukte(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(produkte) != 1 {
		t.Fatalf("Expected 1 produkt with active varianten, got %d", len(produkte))
	}
	if produkte[0].Name != "Produkt 1" {
		t.Fatalf("Expected 'Produkt 1', got %s", produkte[0].Name)
	}
}

func TestGetProdukt(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	produktID, _ := repo.CreateProdukt(ctx, newProdukt("Burger", produkt.EssenKategorie))
	_, _ = repo.CreateVariante(ctx, produktID, newVariante("Single", 599, produkt.ActiveStatus))
	_, _ = repo.CreateVariante(ctx, produktID, newVariante("Double", 899, produkt.ActiveStatus))

	p, err := repo.GetProdukt(ctx, produktID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if p.Name != "Burger" {
		t.Fatalf("Expected 'Burger', got %s", p.Name)
	}
	if len(p.Varianten) != 2 {
		t.Fatalf("Expected 2 varianten, got %d", len(p.Varianten))
	}
}

func TestGetProdukt_NotFound(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	_, err := repo.GetProdukt(ctx, 999999)

	if !errors.Is(err, dbpkg.ErrNotFound) {
		t.Fatalf("Expected ErrNotFound, got %v", err)
	}
}

func TestCreateProdukt(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	produktID, err := repo.CreateProdukt(ctx, newProdukt("French Fries", produkt.EssenKategorie))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if produktID < 1 {
		t.Fatalf("Expected valid produkt ID, got %d", produktID)
	}
}

func TestUpdateProdukt(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	p := newProdukt("Original Produkt", produkt.EssenKategorie)
	produktID, _ := repo.CreateProdukt(ctx, p)

	p.ID = produktID
	p.Name = "Updated Name"
	p.Kategorie = produkt.GetraenkKategorie
	err := repo.UpdateProdukt(ctx, p)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	updated, err := repo.GetProdukt(ctx, produktID)
	if err != nil {
		t.Fatalf("Expected no error retrieving produkt, got %v", err)
	}
	if updated.Name != "Updated Name" {
		t.Fatalf("Expected produkt name 'Updated Name', got %s", updated.Name)
	}
	if updated.Kategorie != produkt.GetraenkKategorie {
		t.Fatalf("Expected produkt category 'getraenk', got %s", updated.Kategorie)
	}
}

func TestUpdateProdukt_NotFound(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	err := repo.UpdateProdukt(ctx, produkt.Produkt{ID: 999999, Name: "Updated Name", Kategorie: produkt.GetraenkKategorie, Steuersatz: steuer.RegelSteuersatz, Status: produkt.ActiveStatus, UpdatedAt: time.Now().UTC()})

	if !errors.Is(err, dbpkg.ErrNotFound) {
		t.Fatalf("Expected ErrNotFound, got %v", err)
	}
}

// Variante tests

func TestCreateVariante(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	produktID, _ := repo.CreateProdukt(ctx, newProdukt("Cola", produkt.GetraenkKategorie))

	varianteID, err := repo.CreateVariante(ctx, produktID, newVariante("0.5L", 299, produkt.ActiveStatus))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if varianteID < 1 {
		t.Fatalf("Expected valid variante ID, got %d", varianteID)
	}
}

func TestGetVariante(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	produktID, _ := repo.CreateProdukt(ctx, newProdukt("Cola", produkt.GetraenkKategorie))
	varianteID, _ := repo.CreateVariante(ctx, produktID, newVariante("0.5L", 299, produkt.ActiveStatus))

	v, err := repo.GetVariante(ctx, varianteID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if v.Name != "0.5L" {
		t.Fatalf("Expected '0.5L', got %s", v.Name)
	}
	if v.PreisCents != 299 {
		t.Fatalf("Expected price 299, got %d", v.PreisCents)
	}
	if v.Status != produkt.ActiveStatus {
		t.Fatalf("Expected status 'active', got %s", v.Status)
	}
}

func TestGetVariante_NotFound(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	_, err := repo.GetVariante(ctx, 999999)

	if !errors.Is(err, dbpkg.ErrNotFound) {
		t.Fatalf("Expected ErrNotFound, got %v", err)
	}
}

func TestUpdateVariante(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	produktID, _ := repo.CreateProdukt(ctx, newProdukt("Cola", produkt.GetraenkKategorie))
	varianteID, _ := repo.CreateVariante(ctx, produktID, newVariante("0.5L", 299, produkt.ActiveStatus))

	v := produkt.Variante{
		ID:         varianteID,
		Name:       "1.0L",
		PreisCents: 499,
		Status:     produkt.InactiveStatus,
	}
	err := repo.UpdateVariante(ctx, v)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	updated, _ := repo.GetVariante(ctx, varianteID)
	if updated.Name != "1.0L" {
		t.Fatalf("Expected '1.0L', got %s", updated.Name)
	}
	if updated.PreisCents != 499 {
		t.Fatalf("Expected price 499, got %d", updated.PreisCents)
	}
	if updated.Status != produkt.InactiveStatus {
		t.Fatalf("Expected status 'inactive', got %s", updated.Status)
	}
}

func TestUpdateVariante_NotFound(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	err := repo.UpdateVariante(ctx, produkt.Variante{ID: 999999, Name: "Test", PreisCents: 100, Status: produkt.ActiveStatus})

	if !errors.Is(err, dbpkg.ErrNotFound) {
		t.Fatalf("Expected ErrNotFound, got %v", err)
	}
}

func TestDeletedVariantenNotReturned(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	produktID, _ := repo.CreateProdukt(ctx, newProdukt("Pizza", produkt.EssenKategorie))
	_, _ = repo.CreateVariante(ctx, produktID, newVariante("Small", 899, produkt.ActiveStatus))
	deletedVarianteID, _ := repo.CreateVariante(ctx, produktID, newVariante("Large", 1299, produkt.ActiveStatus))

	// Simulate soft delete by updating status to 'deleted'
	_, _ = repo.db.ExecContext(ctx, "UPDATE produkt_varianten SET status = 'deleted' WHERE id = $1", deletedVarianteID)

	p, err := repo.GetProdukt(ctx, produktID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(p.Varianten) != 1 {
		t.Fatalf("Expected 1 variante (deleted should be excluded), got %d", len(p.Varianten))
	}
	if p.Varianten[0].Name != "Small" {
		t.Fatalf("Expected 'Small' variante, got %s", p.Varianten[0].Name)
	}
}

func TestVarianteOrderStableAfterUpdate(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	produktID, _ := repo.CreateProdukt(ctx, newProdukt("Pizza", produkt.EssenKategorie))

	firstID, _ := repo.CreateVariante(ctx, produktID, newVariante("Klein", 899, produkt.ActiveStatus))
	middleID, _ := repo.CreateVariante(ctx, produktID, newVariante("Mittel", 1099, produkt.ActiveStatus))
	lastID, _ := repo.CreateVariante(ctx, produktID, newVariante("Groß", 1299, produkt.ActiveStatus))

	// Updating the middle variante must not change its position in the aggregated
	// variante list — without ORDER BY in the json_agg, Postgres may reorder rows.
	err := repo.UpdateVariante(ctx, produkt.Variante{
		ID:         middleID,
		Name:       "Mittel neu",
		PreisCents: 1150,
		Status:     produkt.ActiveStatus,
	})
	if err != nil {
		t.Fatalf("Expected no error updating variante, got %v", err)
	}

	wantOrder := []int{firstID, middleID, lastID}

	p, err := repo.GetProdukt(ctx, produktID)
	if err != nil {
		t.Fatalf("Expected no error from GetProdukt, got %v", err)
	}
	assertVarianteOrder(t, "GetProdukt", p.Varianten, wantOrder)

	all, err := repo.GetAllProdukte(ctx)
	if err != nil {
		t.Fatalf("Expected no error from GetAllProdukte, got %v", err)
	}
	assertVarianteOrder(t, "GetAllProdukte", all[0].Varianten, wantOrder)

	active, err := repo.GetActiveProdukte(ctx)
	if err != nil {
		t.Fatalf("Expected no error from GetActiveProdukte, got %v", err)
	}
	assertVarianteOrder(t, "GetActiveProdukte", active[0].Varianten, wantOrder)
}

func assertVarianteOrder(t *testing.T, query string, varianten []produkt.Variante, wantIDs []int) {
	t.Helper()
	if len(varianten) != len(wantIDs) {
		t.Fatalf("%s: expected %d varianten, got %d", query, len(wantIDs), len(varianten))
	}
	for i, wantID := range wantIDs {
		if varianten[i].ID != wantID {
			t.Fatalf("%s: expected variante ID %d at position %d, got %d", query, wantID, i, varianten[i].ID)
		}
	}
}
