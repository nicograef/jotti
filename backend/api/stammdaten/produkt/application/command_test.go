//go:build unit

package application

import (
	"context"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/domain/steuer"
	"github.com/nicograef/jotti/backend/repository/produkt_repo"
)

var testProdukt = produkt.Produkt{
	ID:         1,
	Name:       "Cola",
	Kategorie:  produkt.GetraenkKategorie,
	Steuersatz: steuer.RegelSteuersatz,
	Status:     produkt.ActiveStatus,
	Varianten:  []produkt.Variante{},
	CreatedAt:  time.Now().UTC(),
	UpdatedAt:  time.Now().UTC(),
}

func TestCreateProdukt(t *testing.T) {
	repo := produkt_repo.NewMock(nil, nil)
	cmd := Command{ProduktRepo: repo}

	id, err := cmd.CreateProdukt(context.Background(), "Bier", produkt.GetraenkKategorie, steuer.RegelSteuersatz)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != 1 {
		t.Errorf("expected produkt ID 1, got %d", id)
	}
}

func TestCreateProdukt_AlreadyExists(t *testing.T) {
	repo := produkt_repo.NewMock(nil, db.ErrAlreadyExists)
	cmd := Command{ProduktRepo: repo}

	_, err := cmd.CreateProdukt(context.Background(), "Bier", produkt.GetraenkKategorie, steuer.RegelSteuersatz)
	if err != ErrProduktAlreadyExists {
		t.Fatalf("expected ErrProduktAlreadyExists, got %v", err)
	}
}

func TestUpdateProdukt(t *testing.T) {
	repo := produkt_repo.NewMock([]produkt.Produkt{testProdukt}, nil)
	cmd := Command{ProduktRepo: repo}

	err := cmd.UpdateProdukt(context.Background(), 1, "Fanta", produkt.GetraenkKategorie, steuer.RegelSteuersatz)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	updated, err := repo.GetProdukt(context.Background(), 1)
	if err != nil {
		t.Fatalf("failed to get updated produkt: %v", err)
	}
	if updated.Name != "Fanta" {
		t.Errorf("expected produkt name 'Fanta', got '%s'", updated.Name)
	}
}

func TestUpdateProdukt_NotFound(t *testing.T) {
	repo := produkt_repo.NewMock(nil, db.ErrNotFound)
	cmd := Command{ProduktRepo: repo}

	err := cmd.UpdateProdukt(context.Background(), 999, "Fanta", produkt.GetraenkKategorie, steuer.RegelSteuersatz)
	if err != ErrProduktNotFound {
		t.Fatalf("expected ErrProduktNotFound, got %v", err)
	}
}

func TestDeleteProdukt(t *testing.T) {
	repo := produkt_repo.NewMock([]produkt.Produkt{testProdukt}, nil)
	cmd := Command{ProduktRepo: repo}

	err := cmd.DeleteProdukt(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	deleted, err := repo.GetProdukt(context.Background(), 1)
	if err != nil {
		t.Fatalf("failed to get produkt after delete: %v", err)
	}
	if deleted.Status != produkt.DeletedStatus {
		t.Errorf("expected status deleted, got %q", deleted.Status)
	}
}
