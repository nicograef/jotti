//go:build integration

package produkt_repo

import (
	"context"
	"database/sql"
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

func rawPfandVerweis(t *testing.T, repo Repository, id int) *int {
	t.Helper()
	var verweis sql.NullInt32
	err := repo.db.QueryRowContext(context.Background(), "SELECT pfand_variante_id FROM produkt_varianten WHERE id = $1", id).Scan(&verweis)
	if err != nil {
		t.Fatalf("failed to read pfand_variante_id: %v", err)
	}
	if !verweis.Valid {
		return nil
	}
	v := int(verweis.Int32)
	return &v
}

// pfandFixture creates a Pfand-Produkt with the given Varianten and returns their IDs in order.
func pfandFixture(t *testing.T, repo Repository, name string, varianten ...string) (int, []int) {
	t.Helper()
	ctx := context.Background()
	p := newProdukt(name, produkt.GetraenkKategorie)
	p.Pfand = true
	produktID, err := repo.CreateProdukt(ctx, p)
	if err != nil {
		t.Fatalf("failed to create %s: %v", name, err)
	}
	ids := make([]int, 0, len(varianten))
	for _, v := range varianten {
		id, err := repo.CreateVariante(ctx, produktID, newVariante(v, 200, produkt.ActiveStatus))
		if err != nil {
			t.Fatalf("failed to create %s %s: %v", name, v, err)
		}
		ids = append(ids, id)
	}
	return produktID, ids
}

// verknuepfteVariante creates a Variante of a new Produkt that links pfandID as its Pfand.
func verknuepfteVariante(t *testing.T, repo Repository, name string, pfandID int) int {
	t.Helper()
	ctx := context.Background()
	produktID, err := repo.CreateProdukt(ctx, newProdukt(name, produkt.GetraenkKategorie))
	if err != nil {
		t.Fatalf("failed to create %s: %v", name, err)
	}
	v := newVariante("0,5 l", 400, produkt.ActiveStatus)
	v.PfandVarianteID = &pfandID
	id, err := repo.CreateVariante(ctx, produktID, v)
	if err != nil {
		t.Fatalf("failed to create variante of %s: %v", name, err)
	}
	return id
}

func assertRawPfandVerweis(t *testing.T, repo Repository, varianteID int, want *int) {
	t.Helper()
	got := rawPfandVerweis(t, repo, varianteID)
	switch {
	case want == nil && got != nil:
		t.Errorf("variante %d: expected no Pfand, got %d", varianteID, *got)
	case want != nil && (got == nil || *got != *want):
		t.Errorf("variante %d: expected Pfand %d, got %v", varianteID, *want, got)
	}
}

func TestDeleteVariante_EntferntPfandVerweise(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	_, pfand := pfandFixture(t, repo, "Pfand", "Glas", "Becher")
	glasID, becherID := pfand[0], pfand[1]
	bierID := verknuepfteVariante(t, repo, "Bier", glasID)
	radlerID := verknuepfteVariante(t, repo, "Radler", glasID)
	weinID := verknuepfteVariante(t, repo, "Wein", becherID)

	glas, err := repo.GetVariante(ctx, glasID)
	if err != nil {
		t.Fatalf("failed to load variante: %v", err)
	}
	glas.Delete()
	if err := repo.DeleteVariante(ctx, glas); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := rawVarianteStatus(t, repo, glasID); got != string(produkt.DeletedStatus) {
		t.Errorf("expected variante %d status deleted, got %q", glasID, got)
	}
	assertRawPfandVerweis(t, repo, bierID, nil)
	assertRawPfandVerweis(t, repo, radlerID, nil)
	assertRawPfandVerweis(t, repo, weinID, &becherID)
}

func TestDeleteProduktMitVarianten_EntferntPfandVerweise(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	glasPfandID, glaeser := pfandFixture(t, repo, "Glaspfand", "Glas 0,3 l", "Glas 0,5 l")
	_, teller := pfandFixture(t, repo, "Tellerpfand", "Teller")
	kleinID := verknuepfteVariante(t, repo, "Schorle", glaeser[0])
	grossID := verknuepfteVariante(t, repo, "Bier", glaeser[1])
	wurstID := verknuepfteVariante(t, repo, "Wurst", teller[0])

	p, err := repo.GetProdukt(ctx, glasPfandID)
	if err != nil {
		t.Fatalf("failed to load produkt: %v", err)
	}
	applyDelete(&p)
	if err := repo.DeleteProduktMitVarianten(ctx, p); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := rawProduktStatus(t, repo, glasPfandID); got != string(produkt.DeletedStatus) {
		t.Errorf("expected produkt status deleted, got %q", got)
	}
	assertRawPfandVerweis(t, repo, kleinID, nil)
	assertRawPfandVerweis(t, repo, grossID, nil)
	assertRawPfandVerweis(t, repo, wurstID, &teller[0])
}
