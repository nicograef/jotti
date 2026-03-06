package product_repo

import (
	"context"
	"fmt"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/product"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

func (r Repository) GetProduct(ctx context.Context, id int) (product.Product, error) {
	row, err := r.q.GetProduct(ctx, id)
	if err != nil {
		return product.Product{}, db.Error(err)
	}

	variants, err := parseVariantsJSON(row.Variants)
	if err != nil {
		return product.Product{}, fmt.Errorf("failed to unmarshal variants: %w", err)
	}

	return product.Product{
		ID:        row.ID,
		Name:      row.Name,
		Category:  product.Category(row.Category),
		Variants:  variants,
		CreatedAt: row.CreatedAt,
	}, nil
}

func (r Repository) GetAllProducts(ctx context.Context) ([]product.Product, error) {
	rows, err := r.q.GetAllProducts(ctx)
	if err != nil {
		return nil, db.Error(err)
	}

	products := make([]product.Product, 0, len(rows))
	for _, row := range rows {
		variants, err := parseVariantsJSON(row.Variants)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal variants: %w", err)
		}

		products = append(products, product.Product{
			ID:        row.ID,
			Name:      row.Name,
			Category:  product.Category(row.Category),
			Variants:  variants,
			CreatedAt: row.CreatedAt,
		})
	}

	return products, nil
}

func (r Repository) GetActiveProducts(ctx context.Context) ([]product.Product, error) {
	rows, err := r.q.GetActiveProducts(ctx)
	if err != nil {
		return nil, db.Error(err)
	}

	products := make([]product.Product, 0, len(rows))
	for _, row := range rows {
		variants, err := parseVariantsJSON(row.Variants)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal variants: %w", err)
		}

		products = append(products, product.Product{
			ID:        row.ID,
			Name:      row.Name,
			Category:  product.Category(row.Category),
			Variants:  variants,
			CreatedAt: row.CreatedAt,
		})
	}

	return products, nil
}

func (r Repository) CreateProduct(ctx context.Context, p product.Product) (int, error) {
	id, err := r.q.CreateProduct(ctx, dbgen.CreateProductParams{
		Name:      p.Name,
		Category:  dbgen.Productcategory(p.Category),
		CreatedAt: p.CreatedAt,
	})
	if err != nil {
		return 0, db.Error(err)
	}

	return id, nil
}

func (r Repository) UpdateProduct(ctx context.Context, p product.Product) error {
	result, err := r.q.UpdateProduct(ctx, dbgen.UpdateProductParams{
		Name:     p.Name,
		Category: dbgen.Productcategory(p.Category),
		ID:       p.ID,
	})
	if err != nil {
		return db.Error(err)
	}

	return db.ResultError(result)
}
