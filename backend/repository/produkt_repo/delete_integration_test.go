//go:build integration

package produkt_repo

import (
	"context"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nicograef/jotti/backend/domain/produkt"
)

// applyDelete marks the produkt and all its varianten as deleted, mirroring what
// the application layer does before persisting the atomic soft-delete.
func applyDelete(p *produkt.Produkt) {
	for i := range p.Varianten {
		p.Varianten[i].Delete()
	}
	p.Delete()
}

func rawProduktStatus(t *testing.T, repo Repository, id int) string {
	t.Helper()
	var status string
	err := repo.db.QueryRowContext(context.Background(), "SELECT status FROM produkte WHERE id = $1", id).Scan(&status)
	if err != nil {
		t.Fatalf("failed to read produkt status: %v", err)
	}
	return status
}

func rawVarianteStatus(t *testing.T, repo Repository, id int) string {
	t.Helper()
	var status string
	err := repo.db.QueryRowContext(context.Background(), "SELECT status FROM produkt_varianten WHERE id = $1", id).Scan(&status)
	if err != nil {
		t.Fatalf("failed to read variante status: %v", err)
	}
	return status
}

func TestDeleteProduktMitVarianten_HappyPath(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	produktID, _ := repo.CreateProdukt(ctx, newProdukt("Pizza", produkt.EssenKategorie))
	smallID, _ := repo.CreateVariante(ctx, produktID, newVariante("Small", 899, produkt.ActiveStatus))
	largeID, _ := repo.CreateVariante(ctx, produktID, newVariante("Large", 1299, produkt.ActiveStatus))

	p, err := repo.GetProdukt(ctx, produktID)
	if err != nil {
		t.Fatalf("failed to load produkt: %v", err)
	}
	applyDelete(&p)

	if err := repo.DeleteProduktMitVarianten(ctx, p); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := rawProduktStatus(t, repo, produktID); got != string(produkt.DeletedStatus) {
		t.Errorf("expected produkt status deleted, got %q", got)
	}
	if got := rawVarianteStatus(t, repo, smallID); got != string(produkt.DeletedStatus) {
		t.Errorf("expected variante %d status deleted, got %q", smallID, got)
	}
	if got := rawVarianteStatus(t, repo, largeID); got != string(produkt.DeletedStatus) {
		t.Errorf("expected variante %d status deleted, got %q", largeID, got)
	}
}

func TestDeleteProduktMitVarianten_MidFailureRollsBack(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	produktID, _ := repo.CreateProdukt(ctx, newProdukt("Pizza", produkt.EssenKategorie))
	varianteID, _ := repo.CreateVariante(ctx, produktID, newVariante("Small", 899, produkt.ActiveStatus))

	p, err := repo.GetProdukt(ctx, produktID)
	if err != nil {
		t.Fatalf("failed to load produkt: %v", err)
	}
	applyDelete(&p)

	// Inject a mid-transaction failure: append a phantom variante whose UPDATE
	// affects zero rows (no such id), so the transaction aborts after the real
	// variante has already been updated inside the same tx.
	p.Varianten = append(p.Varianten, produkt.Variante{
		ID:         999999,
		Name:       "ghost",
		PreisCents: 100,
		Status:     produkt.DeletedStatus,
		UpdatedAt:  time.Now().UTC(),
	})

	if err := repo.DeleteProduktMitVarianten(ctx, p); err == nil {
		t.Fatal("expected an error from the injected mid-transaction failure, got nil")
	}

	// Atomicity: nothing was partially deleted — produkt and real variante remain active.
	if got := rawProduktStatus(t, repo, produktID); got != string(produkt.ActiveStatus) {
		t.Errorf("expected produkt to remain active after rollback, got %q", got)
	}
	if got := rawVarianteStatus(t, repo, varianteID); got != string(produkt.ActiveStatus) {
		t.Errorf("expected variante to remain active after rollback, got %q", got)
	}
}

// A position ordered before the delete stays payable, so its Pfand flag must stay readable.
func TestGetPfandVarianten_LiestGeloeschteProdukte(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	pfand := newProdukt("Pfand", produkt.GetraenkKategorie)
	pfand.Pfand = true
	pfandID, _ := repo.CreateProdukt(ctx, pfand)
	glasID, _ := repo.CreateVariante(ctx, pfandID, newVariante("Glas", 200, produkt.ActiveStatus))
	bierID, _ := repo.CreateProdukt(ctx, newProdukt("Bier", produkt.GetraenkKategorie))
	pilsID, _ := repo.CreateVariante(ctx, bierID, newVariante("Pils", 400, produkt.ActiveStatus))

	p, err := repo.GetProdukt(ctx, pfandID)
	if err != nil {
		t.Fatalf("failed to load produkt: %v", err)
	}
	applyDelete(&p)
	if err := repo.DeleteProduktMitVarianten(ctx, p); err != nil {
		t.Fatalf("delete: %v", err)
	}

	got, err := repo.GetPfandVarianten(ctx, []int{glasID, pilsID})
	if err != nil {
		t.Fatalf("GetPfandVarianten: %v", err)
	}
	if len(got) != 1 || !got[glasID] {
		t.Errorf("GetPfandVarianten = %v, want only %d", got, glasID)
	}
}
