//go:build integration

package produkt_repo

import (
	"context"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nicograef/jotti/backend/domain/produkt"
)

// applyDelete marks the product and all its variants as deleted, mirroring what
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
	productID, _ := repo.CreateProduct(ctx, newProduct("Pizza", produkt.EssenKategorie))
	smallID, _ := repo.CreateVariant(ctx, productID, newVariant("Small", 899, produkt.ActiveStatus))
	largeID, _ := repo.CreateVariant(ctx, productID, newVariant("Large", 1299, produkt.ActiveStatus))

	p, err := repo.GetProduct(ctx, productID)
	if err != nil {
		t.Fatalf("failed to load product: %v", err)
	}
	applyDelete(&p)

	if err := repo.DeleteProduktMitVarianten(ctx, p); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := rawProduktStatus(t, repo, productID); got != string(produkt.DeletedStatus) {
		t.Errorf("expected product status deleted, got %q", got)
	}
	if got := rawVarianteStatus(t, repo, smallID); got != string(produkt.DeletedStatus) {
		t.Errorf("expected variant %d status deleted, got %q", smallID, got)
	}
	if got := rawVarianteStatus(t, repo, largeID); got != string(produkt.DeletedStatus) {
		t.Errorf("expected variant %d status deleted, got %q", largeID, got)
	}
}

func TestDeleteProduktMitVarianten_MidFailureRollsBack(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	productID, _ := repo.CreateProduct(ctx, newProduct("Pizza", produkt.EssenKategorie))
	variantID, _ := repo.CreateVariant(ctx, productID, newVariant("Small", 899, produkt.ActiveStatus))

	p, err := repo.GetProduct(ctx, productID)
	if err != nil {
		t.Fatalf("failed to load product: %v", err)
	}
	applyDelete(&p)

	// Inject a mid-transaction failure: append a phantom variant whose UPDATE
	// affects zero rows (no such id), so the transaction aborts after the real
	// variant has already been updated inside the same tx.
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

	// Atomicity: nothing was partially deleted — product and real variant remain active.
	if got := rawProduktStatus(t, repo, productID); got != string(produkt.ActiveStatus) {
		t.Errorf("expected product to remain active after rollback, got %q", got)
	}
	if got := rawVarianteStatus(t, repo, variantID); got != string(produkt.ActiveStatus) {
		t.Errorf("expected variant to remain active after rollback, got %q", got)
	}
}
