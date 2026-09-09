package produkt_repo

import (
	"context"
	"time"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/domain/steuer"
)

// GetVariantenByIDs fetches multiple varianten in a single query.
// Returns a map keyed by variante ID for O(1) lookup during Bestellung enrichment.
// Each entry carries its produkt_id, so enrichment can verify the Produkt/Variante
// pairing sent by the client instead of trusting it.
// Uses ANY($1) with a []int32 parameter; pgx v5 encodes Go slices as PostgreSQL arrays
// natively, so no dynamic SQL building is required.
func (r Repository) GetVariantenByIDs(ctx context.Context, ids []int) (map[int]produkt.VarianteMitProdukt, error) {
	if len(ids) == 0 {
		return make(map[int]produkt.VarianteMitProdukt), nil
	}

	ids32 := toInt32Slice(ids)

	const query = `SELECT id, produkt_id, name, preis_cents, status, created_at, updated_at
		FROM produkt_varianten
		WHERE id = ANY($1) AND status != 'deleted'`

	rows, err := r.db.QueryContext(ctx, query, ids32)
	if err != nil {
		return nil, db.Error(err)
	}
	defer rows.Close() //nolint:errcheck // explicit Close with error check below

	result := make(map[int]produkt.VarianteMitProdukt, len(ids))
	for rows.Next() {
		var (
			id         int
			produktID  int
			name       string
			preisCents int
			status     string
			createdAt  time.Time
			updatedAt  time.Time
		)
		if err := rows.Scan(&id, &produktID, &name, &preisCents, &status, &createdAt, &updatedAt); err != nil {
			return nil, db.Error(err)
		}
		result[id] = produkt.VarianteMitProdukt{
			Variante: produkt.Variante{
				ID:         id,
				Name:       name,
				PreisCents: preisCents,
				Status:     produkt.Status(status),
				CreatedAt:  createdAt,
				UpdatedAt:  updatedAt,
			},
			ProduktID: produktID,
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

// GetProdukteByIDs fetches multiple produkte in a single query.
// Returns a map keyed by produkt ID for O(1) lookup during Bestellung enrichment.
// Retrieves the fields needed for fat-event enrichment (Name, Kategorie, Steuersatz)
// plus Status, so the sales path can reject deactivated produkte server-side.
// Uses ANY($1) with a []int32 parameter; see GetVariantenByIDs for rationale.
func (r Repository) GetProdukteByIDs(ctx context.Context, ids []int) (map[int]produkt.Produkt, error) {
	if len(ids) == 0 {
		return make(map[int]produkt.Produkt), nil
	}

	ids32 := toInt32Slice(ids)

	const query = `SELECT id, name, kategorie, steuersatz, status
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
			status     string
		)
		if err := rows.Scan(&id, &name, &kategorie, &steuersatz, &status); err != nil {
			return nil, db.Error(err)
		}
		result[id] = produkt.Produkt{
			ID:         id,
			Name:       name,
			Kategorie:  produkt.Kategorie(kategorie),
			Steuersatz: steuer.Steuersatz(steuersatz),
			Status:     produkt.Status(status),
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
