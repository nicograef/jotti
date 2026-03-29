package product_repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/product"
)

// GetVariantsByIDs fetches multiple variants in a single query.
// Returns a map keyed by variant ID for O(1) lookup during Bestellung enrichment.
// Uses a dynamically built parameterized IN clause; all arguments are integer IDs from
// validated application inputs (not user-supplied strings), so this is safe.
func (r Repository) GetVariantsByIDs(ctx context.Context, ids []int) (map[int]product.Variante, error) {
	if len(ids) == 0 {
		return make(map[int]product.Variante), nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := `SELECT id, name, preis_cents, status, created_at, updated_at
              FROM produkt_varianten
              WHERE id IN (` + strings.Join(placeholders, ",") + `) AND status != 'deleted'`

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, db.Error(err)
	}
	defer rows.Close()

	result := make(map[int]product.Variante, len(ids))
	for rows.Next() {
		var (
			id         int
			name       string
			preisCents int
			status     string
			createdAt  time.Time
			updatedAt  time.Time
		)
		if err := rows.Scan(&id, &name, &preisCents, &status, &createdAt, &updatedAt); err != nil {
			return nil, db.Error(err)
		}
		result[id] = product.Variante{
			ID:         id,
			Name:       name,
			PreisCents: preisCents,
			Status:     product.Status(status),
			CreatedAt:  createdAt,
			UpdatedAt:  updatedAt,
		}
	}
	if err := rows.Close(); err != nil {
		return nil, db.Error(err)
	}
	if err := rows.Err(); err != nil {
		return nil, db.Error(err)
	}

	return result, nil
}

// GetProductsByIDs fetches multiple products in a single query.
// Returns a map keyed by product ID for O(1) lookup during Bestellung enrichment.
// Only retrieves fields needed for fat-event enrichment (Name, Kategorie).
// Uses a dynamically built parameterized IN clause; see GetVariantsByIDs for rationale.
func (r Repository) GetProductsByIDs(ctx context.Context, ids []int) (map[int]product.Produkt, error) {
	if len(ids) == 0 {
		return make(map[int]product.Produkt), nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := `SELECT id, name, kategorie
              FROM produkte
              WHERE id IN (` + strings.Join(placeholders, ",") + `) AND status != 'deleted'`

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, db.Error(err)
	}
	defer rows.Close()

	result := make(map[int]product.Produkt, len(ids))
	for rows.Next() {
		var (
			id        int
			name      string
			kategorie string
		)
		if err := rows.Scan(&id, &name, &kategorie); err != nil {
			return nil, db.Error(err)
		}
		result[id] = product.Produkt{
			ID:        id,
			Name:      name,
			Kategorie: product.Kategorie(kategorie),
		}
	}
	if err := rows.Close(); err != nil {
		return nil, db.Error(err)
	}
	if err := rows.Err(); err != nil {
		return nil, db.Error(err)
	}

	return result, nil
}
