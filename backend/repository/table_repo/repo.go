package table_repo

import (
	"context"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/table"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

func (r Repository) GetTable(ctx context.Context, id int) (table.Tisch, error) {
	row, err := r.q.GetTisch(ctx, id)
	if err != nil {
		return table.Tisch{}, db.Error(err)
	}

	return tischRowToDomain(row), nil
}

func (r Repository) GetAllTables(ctx context.Context) ([]table.Tisch, error) {
	rows, err := r.q.GetAlleTische(ctx)
	if err != nil {
		return nil, db.Error(err)
	}

	tables := make([]table.Tisch, 0, len(rows))
	for _, row := range rows {
		tables = append(tables, tischRowToDomain(row))
	}

	return tables, nil
}

func (r Repository) GetActiveTables(ctx context.Context) ([]table.AktiverTisch, error) {
	rows, err := r.q.GetAktiveTische(ctx)
	if err != nil {
		return nil, db.Error(err)
	}

	tables := make([]table.AktiverTisch, 0, len(rows))
	for _, row := range rows {
		tables = append(tables, table.AktiverTisch{
			ID:         row.ID,
			Name:       row.Name,
			SaldoCents: row.SaldoCents,
		})
	}

	return tables, nil
}

func (r Repository) CreateTable(ctx context.Context, t table.Tisch) (int, error) {
	id, err := r.q.CreateTisch(ctx, dbgen.CreateTischParams{
		Name:      t.Name,
		Status:    dbgen.Entitystatus(t.Status),
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	})
	if err != nil {
		return 0, db.Error(err)
	}

	return id, nil
}

func (r Repository) UpdateTable(ctx context.Context, t table.Tisch) error {
	result, err := r.q.UpdateTisch(ctx, dbgen.UpdateTischParams{
		Name:      t.Name,
		Status:    dbgen.Entitystatus(t.Status),
		UpdatedAt: t.UpdatedAt,
		ID:        t.ID,
	})
	if err != nil {
		return db.Error(err)
	}

	return db.ResultError(result)
}
