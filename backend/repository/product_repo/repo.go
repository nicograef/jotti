package product_repo

import (
	"context"
	"fmt"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/product"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

func (r Repository) GetProduct(ctx context.Context, id int) (product.Produkt, error) {
	row, err := r.q.GetProduct(ctx, id)
	if err != nil {
		return product.Produkt{}, db.Error(err)
	}

	variants, err := parseVariantsJSON(row.Variants)
	if err != nil {
		return product.Produkt{}, fmt.Errorf("failed to unmarshal variants: %w", err)
	}

	return product.Produkt{
		ID:        row.ID,
		Name:      row.Name,
		Kategorie: product.Kategorie(row.Category),
		Status:    product.Status(row.Status),
		Variants:  variants,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r Repository) GetAllProducts(ctx context.Context) ([]product.Produkt, error) {
	rows, err := r.q.GetAllProducts(ctx)
	if err != nil {
		return nil, db.Error(err)
	}

	products := make([]product.Produkt, 0, len(rows))
	for i := range rows {
		variants, err := parseVariantsJSON(rows[i].Variants)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal variants: %w", err)
		}

		products = append(products, product.Produkt{
			ID:        rows[i].ID,
			Name:      rows[i].Name,
			Kategorie: product.Kategorie(rows[i].Category),
			Status:    product.Status(rows[i].Status),
			Variants:  variants,
			CreatedAt: rows[i].CreatedAt,
			UpdatedAt: rows[i].UpdatedAt,
		})
	}

	return products, nil
}

func (r Repository) GetActiveProducts(ctx context.Context) ([]product.Produkt, error) {
	rows, err := r.q.GetActiveProducts(ctx)
	if err != nil {
		return nil, db.Error(err)
	}

	products := make([]product.Produkt, 0, len(rows))
	for i := range rows {
		variants, err := parseVariantsJSON(rows[i].Variants)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal variants: %w", err)
		}

		products = append(products, product.Produkt{
			ID:        rows[i].ID,
			Name:      rows[i].Name,
			Kategorie: product.Kategorie(rows[i].Category),
			Status:    product.Status(rows[i].Status),
			Variants:  variants,
			CreatedAt: rows[i].CreatedAt,
			UpdatedAt: rows[i].UpdatedAt,
		})
	}

	return products, nil
}

func (r Repository) CreateProduct(ctx context.Context, p product.Produkt) (int, error) {
	id, err := r.q.CreateProduct(ctx, dbgen.CreateProductParams{
		Name:      p.Name,
		Category:  dbgen.Productcategory(p.Kategorie),
		Status:    dbgen.Entitystatus(p.Status),
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	})
	if err != nil {
		return 0, db.Error(err)
	}

	return id, nil
}

func (r Repository) UpdateProduct(ctx context.Context, p product.Produkt) error {
	result, err := r.q.UpdateProduct(ctx, dbgen.UpdateProductParams{
		Name:      p.Name,
		Category:  dbgen.Productcategory(p.Kategorie),
		Status:    dbgen.Entitystatus(p.Status),
		UpdatedAt: p.UpdatedAt,
		ID:        p.ID,
	})
	if err != nil {
		return db.Error(err)
	}

	return db.ResultError(result)
}
