package produkt_repo

import (
	"context"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

func (r Repository) GetVariante(ctx context.Context, id int) (produkt.Variante, error) {
	row, err := r.q.GetVariante(ctx, id)
	if err != nil {
		return produkt.Variante{}, db.Error(err)
	}

	return varianteRowToDomain(row), nil
}

// UpdateVarianten writes all given Varianten in one transaction: either every one changes or none.
func (r Repository) UpdateVarianten(ctx context.Context, varianten []produkt.Variante) error {
	return db.WithTx(ctx, r.db, func(qtx *dbgen.Queries) error {
		for i := range varianten {
			result, err := qtx.UpdateVariante(ctx, updateVarianteParams(varianten[i]))
			if err != nil {
				return db.Error(err)
			}
			if err := db.ResultError(result); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r Repository) CreateVariante(ctx context.Context, produktID int, v produkt.Variante) (int, error) {
	id, err := r.q.CreateVariante(ctx, dbgen.CreateVarianteParams{
		ProduktID:  produktID,
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

func (r Repository) UpdateVariante(ctx context.Context, v produkt.Variante) error {
	result, err := r.q.UpdateVariante(ctx, updateVarianteParams(v))
	if err != nil {
		return db.Error(err)
	}

	return db.ResultError(result)
}
