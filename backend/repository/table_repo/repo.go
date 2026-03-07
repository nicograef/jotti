package table_repo

import (
	"context"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/table"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

func (r Repository) GetTable(ctx context.Context, id int) (table.Table, error) {
	row, err := r.q.GetTable(ctx, id)
	if err != nil {
		return table.Table{}, db.Error(err)
	}

	return tableRowToDomain(row), nil
}

func (r Repository) GetAllTables(ctx context.Context) ([]table.Table, error) {
	rows, err := r.q.GetAllTables(ctx)
	if err != nil {
		return nil, db.Error(err)
	}

	tables := make([]table.Table, 0, len(rows))
	for _, row := range rows {
		tables = append(tables, tableRowToDomain(row))
	}

	return tables, nil
}

func (r Repository) GetActiveTables(ctx context.Context) ([]table.Table, error) {
	rows, err := r.q.GetActiveTables(ctx)
	if err != nil {
		return nil, db.Error(err)
	}

	tables := make([]table.Table, 0, len(rows))
	for _, row := range rows {
		tables = append(tables, tableRowToDomain(row))
	}

	return tables, nil
}

func (r Repository) CreateTable(ctx context.Context, t table.Table) (int, error) {
	id, err := r.q.CreateTable(ctx, dbgen.CreateTableParams{
		Name:      t.Name,
		Status:    dbgen.Entitystatus(t.Status),
		CreatedAt: t.CreatedAt,
	})
	if err != nil {
		return 0, db.Error(err)
	}

	return id, nil
}

func (r Repository) UpdateTable(ctx context.Context, t table.Table) error {
	result, err := r.q.UpdateTable(ctx, dbgen.UpdateTableParams{
		Name:   t.Name,
		Status: dbgen.Entitystatus(t.Status),
		ID:     t.ID,
	})
	if err != nil {
		return db.Error(err)
	}

	return db.ResultError(result)
}
