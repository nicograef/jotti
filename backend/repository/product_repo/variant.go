package product_repo

import (
	"context"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/product"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

func (r Repository) GetVariant(ctx context.Context, id int) (product.Variant, error) {
	row, err := r.q.GetVariant(ctx, id)
	if err != nil {
		return product.Variant{}, db.Error(err)
	}

	return variantRowToDomain(row), nil
}

func (r Repository) CreateVariant(ctx context.Context, productID int, v product.Variant) (int, error) {
	id, err := r.q.CreateVariant(ctx, dbgen.CreateVariantParams{
		ProductID:  productID,
		Name:       v.Name,
		PriceCents: v.PriceCents,
		Status:     dbgen.Entitystatus(v.Status),
		CreatedAt:  v.CreatedAt,
	})
	if err != nil {
		return 0, db.Error(err)
	}

	return id, nil
}

func (r Repository) UpdateVariant(ctx context.Context, v product.Variant) error {
	result, err := r.q.UpdateVariant(ctx, dbgen.UpdateVariantParams{
		Name:       v.Name,
		PriceCents: v.PriceCents,
		Status:     dbgen.Entitystatus(v.Status),
		ID:         v.ID,
	})
	if err != nil {
		return db.Error(err)
	}

	return db.ResultError(result)
}
