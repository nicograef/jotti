package produkt_repo

import (
	"context"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

func (r Repository) GetVariant(ctx context.Context, id int) (produkt.Variante, error) {
	row, err := r.q.GetVariante(ctx, id)
	if err != nil {
		return produkt.Variante{}, db.Error(err)
	}

	return variantRowToDomain(row), nil
}

func (r Repository) CreateVariant(ctx context.Context, productID int, v produkt.Variante) (int, error) {
	id, err := r.q.CreateVariante(ctx, dbgen.CreateVarianteParams{
		ProduktID:  productID,
		Name:       v.Name,
		PreisCents: v.PreisCents,
		Status:     dbgen.Entitystatus(v.Status),
		CreatedAt:  v.CreatedAt,
		UpdatedAt:  v.UpdatedAt,
	})
	if err != nil {
		return 0, db.Error(err)
	}

	return id, nil
}

func (r Repository) UpdateVariant(ctx context.Context, v produkt.Variante) error {
	result, err := r.q.UpdateVariante(ctx, dbgen.UpdateVarianteParams{
		Name:       v.Name,
		PreisCents: v.PreisCents,
		Status:     dbgen.Entitystatus(v.Status),
		UpdatedAt:  v.UpdatedAt,
		ID:         v.ID,
	})
	if err != nil {
		return db.Error(err)
	}

	return db.ResultError(result)
}
