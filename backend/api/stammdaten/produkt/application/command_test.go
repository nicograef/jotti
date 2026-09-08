//go:build unit

package application

import (
	"context"
	"errors"
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
	if !errors.Is(err, ErrProduktAlreadyExists) {
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
	if !errors.Is(err, ErrProduktNotFound) {
		t.Fatalf("expected ErrProduktNotFound, got %v", err)
	}
}

func TestUpdateProdukt_AlreadyExists(t *testing.T) {
	repo := produkt_repo.NewMock([]produkt.Produkt{testProdukt}, nil)
	repo.SetUpdateProduktError(db.ErrAlreadyExists)
	cmd := Command{ProduktRepo: repo}

	err := cmd.UpdateProdukt(context.Background(), 1, "Fanta", produkt.GetraenkKategorie, steuer.RegelSteuersatz)
	if !errors.Is(err, ErrProduktAlreadyExists) {
		t.Fatalf("expected ErrProduktAlreadyExists, got %v", err)
	}
}

func TestVerschiebeProdukt_NotFound(t *testing.T) {
	repo := produkt_repo.NewMock(nil, db.ErrNotFound)
	cmd := Command{ProduktRepo: repo}

	err := cmd.VerschiebeProdukt(context.Background(), 999, produkt.RichtungHoch)
	if !errors.Is(err, ErrProduktNotFound) {
		t.Fatalf("expected ErrProduktNotFound, got %v", err)
	}
}

func TestVerschiebeVariante_NotFound(t *testing.T) {
	repo := produkt_repo.NewMock(nil, db.ErrNotFound)
	cmd := Command{ProduktRepo: repo}

	err := cmd.VerschiebeVariante(context.Background(), 999, produkt.RichtungRunter)
	if !errors.Is(err, ErrVarianteNotFound) {
		t.Fatalf("expected ErrVarianteNotFound, got %v", err)
	}
}

// testVarianteVon baut ein Produkt mit genau einer Variante — so, wie GetProdukt
// es liefert: die Variantenliste des Produkts und die Variante selbst.
func testVarianteVon(id int) (produkt.Produkt, produkt.Variante) {
	variante := produkt.Variante{
		ID:         1,
		Name:       "0,5l",
		PreisCents: 350,
		Status:     produkt.ActiveStatus,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	p := testProdukt
	p.ID = id
	p.Varianten = []produkt.Variante{variante}

	return p, variante
}

func TestDeleteVariante(t *testing.T) {
	eigenes, variante := testVarianteVon(1)
	repo := produkt_repo.NewMock([]produkt.Produkt{eigenes}, nil)
	repo.AddVariante(eigenes.ID, variante)
	cmd := Command{ProduktRepo: repo}

	if err := cmd.DeleteVariante(context.Background(), eigenes.ID, variante.ID); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	geloescht, err := repo.GetVariante(context.Background(), variante.ID)
	if err != nil {
		t.Fatalf("failed to get variante after delete: %v", err)
	}
	if geloescht.Status != produkt.DeletedStatus {
		t.Errorf("expected status deleted, got %q", geloescht.Status)
	}
}

// Eine Variante gehört genau einem Produkt. Nennt der Aufruf ein fremdes
// Produkt, wird nichts gelöscht.
func TestDeleteVariante_FremdeVariante(t *testing.T) {
	eigenes, variante := testVarianteVon(1)
	fremdes := testProdukt
	fremdes.ID = 2
	fremdes.Name = "Wasser"
	fremdes.Varianten = []produkt.Variante{}

	repo := produkt_repo.NewMock([]produkt.Produkt{eigenes, fremdes}, nil)
	repo.AddVariante(eigenes.ID, variante)
	cmd := Command{ProduktRepo: repo}

	err := cmd.DeleteVariante(context.Background(), fremdes.ID, variante.ID)
	if !errors.Is(err, ErrVarianteNotFound) {
		t.Fatalf("expected ErrVarianteNotFound, got %v", err)
	}

	unberuehrt, err := repo.GetVariante(context.Background(), variante.ID)
	if err != nil {
		t.Fatalf("failed to get variante after rejected delete: %v", err)
	}
	if unberuehrt.Status != produkt.ActiveStatus {
		t.Errorf("expected status active, got %q", unberuehrt.Status)
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
