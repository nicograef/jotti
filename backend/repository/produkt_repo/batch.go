package produkt_repo

import (
	"context"
	"time"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/domain/steuer"
)

// GetVariantsByIDs fetches multiple variants in a single query.
// Returns a map keyed by variant ID for O(1) lookup during Bestellung enrichment.
// Uses ANY($1) with a []int32 parameter; pgx v5 encodes Go slices as PostgreSQL arrays
// natively, so no dynamic SQL building is required.
func (r Repository) GetVariantsByIDs(ctx context.Context, ids []int) (map[int]produkt.Variante, error) {
	if len(ids) == 0 {
		return make(map[int]produkt.Variante), nil
	}

	ids32 := toInt32Slice(ids)

	const query = `SELECT id, name, preis_cents, status, created_at, updated_at
		FROM produkt_varianten
		WHERE id = ANY($1) AND status != 'deleted'`

	rows, err := r.db.QueryContext(ctx, query, ids32)
	if err != nil {
		return nil, db.Error(err)
	}
	defer rows.Close() //nolint:errcheck // explicit Close with error check below

	result := make(map[int]produkt.Variante, len(ids))
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
		result[id] = produkt.Variante{
			ID:         id,
			Name:       name,
			PreisCents: preisCents,
			Status:     produkt.Status(status),
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
// Only retrieves fields needed for fat-event enrichment (Name, Kategorie, Steuersatz).
// Uses ANY($1) with a []int32 parameter; see GetVariantsByIDs for rationale.
func (r Repository) GetProductsByIDs(ctx context.Context, ids []int) (map[int]produkt.Produkt, error) {
	if len(ids) == 0 {
		return make(map[int]produkt.Produkt), nil
	}

	ids32 := toInt32Slice(ids)

	const query = `SELECT id, name, kategorie, steuersatz
		FROM produkte
		WHERE id = ANY($1) AND status != 'deleted'`

	rows, err := r.db.QueryContext(ctx, query, ids32)
	if err != nil {
		return nil, db.Error(err)
	}
	defer rows.Close() //nolint:errcheck // explicit Close with error check below

	result := make(map[int]produkt.Produkt, len(ids))
	for rows.Next() {
		var (
			id         int
			name       string
			kategorie  string
			steuersatz string
		)
		if err := rows.Scan(&id, &name, &kategorie, &steuersatz); err != nil {
			return nil, db.Error(err)
		}
		result[id] = produkt.Produkt{
			ID:         id,
			Name:       name,
			Kategorie:  produkt.Kategorie(kategorie),
			Steuersatz: steuer.Steuersatz(steuersatz),
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

// toInt32Slice converts []int to []int32 for use as a PostgreSQL int4[] array parameter.
// pgx v5 encodes []int32 as int4[] natively via TryWrapSliceEncodePlan.
func toInt32Slice(ids []int) []int32 {
	out := make([]int32, len(ids))
	for i, id := range ids {
		out[i] = int32(id) //nolint:gosec // IDs are positive entity IDs, not user-supplied bit-width-sensitive values
	}
	return out
}
