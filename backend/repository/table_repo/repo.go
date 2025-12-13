package table_repo

import (
	"context"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/table"
)

func (r Repository) GetTable(ctx context.Context, id int) (table.Table, error) {
	var dbTable dbtable
	err := r.DB.QueryRowContext(ctx, "SELECT id, name, status, created_at FROM tables WHERE id = $1 AND status != 'deleted'", id).
		Scan(&dbTable.ID, &dbTable.Name, &dbTable.Status, &dbTable.CreatedAt)
	if err != nil {
		return table.Table{}, db.Error(err)
	}

	return dbTable.toDomain(), nil
}

func (r Repository) GetAllTables(ctx context.Context) ([]table.Table, error) {
	rows, err := r.DB.QueryContext(ctx, "SELECT id, name, status, created_at FROM tables WHERE status != 'deleted' ORDER BY id ASC")
	if err != nil {
		return nil, db.Error(err)
	}
	defer db.Close(rows, "tables")

	tables := []table.Table{}
	for rows.Next() {
		var dbTable dbtable
		if err := rows.Scan(&dbTable.ID, &dbTable.Name, &dbTable.Status, &dbTable.CreatedAt); err != nil {
			return nil, db.Error(err)
		}

		tables = append(tables, dbTable.toDomain())
	}

	if err := rows.Err(); err != nil {
		return nil, db.Error(err)
	}

	return tables, nil
}

func (r Repository) GetActiveTables(ctx context.Context) ([]table.Table, error) {
	rows, err := r.DB.QueryContext(ctx, "SELECT id, name, status, created_at FROM tables WHERE status = 'active' ORDER BY id ASC")
	if err != nil {
		return nil, db.Error(err)
	}
	defer db.Close(rows, "tables")

	tables := []table.Table{}
	for rows.Next() {
		var dbTable dbtable
		if err := rows.Scan(&dbTable.ID, &dbTable.Name, &dbTable.Status, &dbTable.CreatedAt); err != nil {
			return nil, db.Error(err)
		}

		tables = append(tables, dbTable.toDomain())
	}

	if err := rows.Err(); err != nil {
		return nil, db.Error(err)
	}

	return tables, nil
}

func (r Repository) CreateTable(ctx context.Context, t table.Table) (int, error) {
	var id int
	err := r.DB.QueryRowContext(ctx, "INSERT INTO tables (name, status, created_at) VALUES ($1, $2, $3) RETURNING id", t.Name, t.Status, t.CreatedAt).Scan(&id)
	if err != nil {
		return 0, db.Error(err)
	}

	return id, nil
}

func (r Repository) UpdateTable(ctx context.Context, t table.Table) error {
	result, err := r.DB.ExecContext(ctx, "UPDATE tables SET name = $1, status = $2 WHERE id = $3", t.Name, t.Status, t.ID)
	if err != nil {
		return db.Error(err)
	}

	return db.ResultError(result)
}
