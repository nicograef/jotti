package produkt_repo

import (
	"context"
	"fmt"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/domain/steuer"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

func (r Repository) GetProduct(ctx context.Context, id int) (produkt.Produkt, error) {
	row, err := r.q.GetProdukt(ctx, id)
	if err != nil {
		return produkt.Produkt{}, db.Error(err)
	}

	varianten, err := parseVariantsJSON(row.Varianten)
	if err != nil {
		return produkt.Produkt{}, fmt.Errorf("unmarshal variants: %w", err)
	}

	return produkt.Produkt{
		ID:         row.ID,
		Name:       row.Name,
		Kategorie:  produkt.Kategorie(row.Kategorie),
		Steuersatz: steuer.Steuersatz(row.Steuersatz),
		Status:     produkt.Status(row.Status),
		Varianten:  varianten,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}, nil
}

func (r Repository) GetAllProducts(ctx context.Context) ([]produkt.Produkt, error) {
	rows, err := r.q.GetAlleProdukte(ctx)
	if err != nil {
		return nil, db.Error(err)
	}

	products := make([]produkt.Produkt, 0, len(rows))
	for i := range rows {
		varianten, err := parseVariantsJSON(rows[i].Varianten)
		if err != nil {
			return nil, fmt.Errorf("unmarshal variants: %w", err)
		}

		products = append(products, produkt.Produkt{
			ID:         rows[i].ID,
			Name:       rows[i].Name,
			Kategorie:  produkt.Kategorie(rows[i].Kategorie),
			Steuersatz: steuer.Steuersatz(rows[i].Steuersatz),
			Status:     produkt.Status(rows[i].Status),
			Varianten:  varianten,
			CreatedAt:  rows[i].CreatedAt,
			UpdatedAt:  rows[i].UpdatedAt,
		})
	}

	return products, nil
}

func (r Repository) GetActiveProducts(ctx context.Context) ([]produkt.Produkt, error) {
	rows, err := r.q.GetAktiveProdukte(ctx)
	if err != nil {
		return nil, db.Error(err)
	}

	products := make([]produkt.Produkt, 0, len(rows))
	for i := range rows {
		varianten, err := parseVariantsJSON(rows[i].Varianten)
		if err != nil {
			return nil, fmt.Errorf("unmarshal variants: %w", err)
		}

		products = append(products, produkt.Produkt{
			ID:         rows[i].ID,
			Name:       rows[i].Name,
			Kategorie:  produkt.Kategorie(rows[i].Kategorie),
			Steuersatz: steuer.Steuersatz(rows[i].Steuersatz),
			Status:     produkt.Status(rows[i].Status),
			Varianten:  varianten,
			CreatedAt:  rows[i].CreatedAt,
			UpdatedAt:  rows[i].UpdatedAt,
		})
	}

	return products, nil
}

func (r Repository) CreateProduct(ctx context.Context, p produkt.Produkt) (int, error) {
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

func (r Repository) UpdateProduct(ctx context.Context, p produkt.Produkt) error {
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

// DeleteProduktMitVarianten persists the soft-delete of a product together with
// all its variants in a single transaction. The caller passes the product with
// Delete() already applied to it and each variant; this method only writes the
// status transitions. Because all writes share one db.WithTx, a mid-operation
// failure rolls the whole delete back — the product and every variant stay in
// their pre-delete state, never a partial delete.
func (r Repository) DeleteProduktMitVarianten(ctx context.Context, p produkt.Produkt) error {
	return db.WithTx(ctx, r.db, func(qtx *dbgen.Queries) error {
		for i := range p.Varianten {
			v := p.Varianten[i]
			result, err := qtx.UpdateVariante(ctx, dbgen.UpdateVarianteParams{
				Name:       v.Name,
				PreisCents: v.PreisCents,
				Status:     dbgen.Entitystatus(v.Status),
				UpdatedAt:  v.UpdatedAt,
				ID:         v.ID,
			})
			if err != nil {
				return db.Error(err)
			}
			if err := db.ResultError(result); err != nil {
				return err
			}
		}

		result, err := qtx.UpdateProdukt(ctx, dbgen.UpdateProduktParams{
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
	})
}
