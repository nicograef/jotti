package product_repo

import (
	"context"
	"fmt"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/product"
	"github.com/nicograef/jotti/backend/domain/steuer"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

func (r Repository) GetProduct(ctx context.Context, id int) (product.Produkt, error) {
	row, err := r.q.GetProdukt(ctx, id)
	if err != nil {
		return product.Produkt{}, db.Error(err)
	}

	varianten, err := parseVariantsJSON(row.Varianten)
	if err != nil {
		return product.Produkt{}, fmt.Errorf("unmarshal variants: %w", err)
	}

	return product.Produkt{
		ID:         row.ID,
		Name:       row.Name,
		Kategorie:  product.Kategorie(row.Kategorie),
		Steuersatz: steuer.Steuersatz(row.Steuersatz),
		Status:     product.Status(row.Status),
		Varianten:  varianten,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}, nil
}

func (r Repository) GetAllProducts(ctx context.Context) ([]product.Produkt, error) {
	rows, err := r.q.GetAlleProdukte(ctx)
	if err != nil {
		return nil, db.Error(err)
	}

	products := make([]product.Produkt, 0, len(rows))
	for i := range rows {
		varianten, err := parseVariantsJSON(rows[i].Varianten)
		if err != nil {
			return nil, fmt.Errorf("unmarshal variants: %w", err)
		}

		products = append(products, product.Produkt{
			ID:         rows[i].ID,
			Name:       rows[i].Name,
			Kategorie:  product.Kategorie(rows[i].Kategorie),
			Steuersatz: steuer.Steuersatz(rows[i].Steuersatz),
			Status:     product.Status(rows[i].Status),
			Varianten:  varianten,
			CreatedAt:  rows[i].CreatedAt,
			UpdatedAt:  rows[i].UpdatedAt,
		})
	}

	return products, nil
}

func (r Repository) GetActiveProducts(ctx context.Context) ([]product.Produkt, error) {
	rows, err := r.q.GetAktiveProdukte(ctx)
	if err != nil {
		return nil, db.Error(err)
	}

	products := make([]product.Produkt, 0, len(rows))
	for i := range rows {
		varianten, err := parseVariantsJSON(rows[i].Varianten)
		if err != nil {
			return nil, fmt.Errorf("unmarshal variants: %w", err)
		}

		products = append(products, product.Produkt{
			ID:         rows[i].ID,
			Name:       rows[i].Name,
			Kategorie:  product.Kategorie(rows[i].Kategorie),
			Steuersatz: steuer.Steuersatz(rows[i].Steuersatz),
			Status:     product.Status(rows[i].Status),
			Varianten:  varianten,
			CreatedAt:  rows[i].CreatedAt,
			UpdatedAt:  rows[i].UpdatedAt,
		})
	}

	return products, nil
}

func (r Repository) CreateProduct(ctx context.Context, p product.Produkt) (int, error) {
	id, err := r.q.CreateProdukt(ctx, dbgen.CreateProduktParams{
		Name:       p.Name,
		Kategorie:  dbgen.Produktkategorie(p.Kategorie),
		Steuersatz: dbgen.Steuersatz(p.Steuersatz),
		Status:     dbgen.Entitystatus(p.Status),
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
	})
	if err != nil {
		return 0, db.Error(err)
	}

	return id, nil
}

func (r Repository) UpdateProduct(ctx context.Context, p product.Produkt) error {
	result, err := r.q.UpdateProdukt(ctx, dbgen.UpdateProduktParams{
		Name:       p.Name,
		Kategorie:  dbgen.Produktkategorie(p.Kategorie),
		Steuersatz: dbgen.Steuersatz(p.Steuersatz),
		Status:     dbgen.Entitystatus(p.Status),
		UpdatedAt:  p.UpdatedAt,
		ID:         p.ID,
	})
	if err != nil {
		return db.Error(err)
	}

	return db.ResultError(result)
}
