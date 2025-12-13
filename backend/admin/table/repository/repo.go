package repository

import (
	"context"

	"github.com/nicograef/jotti/backend/admin/table/domain"
	"github.com/nicograef/jotti/backend/db"
)

// GetTable retrieves a table by its ID from the database.
func (r Repository) GetTable(ctx context.Context, id int) (domain.Table, error) {
	var dbTable dbtable
	err := r.DB.QueryRowContext(ctx, "SELECT id, name, status, created_at FROM tables WHERE id = $1 AND status != 'deleted'", id).
		Scan(&dbTable.ID, &dbTable.Name, &dbTable.Status, &dbTable.CreatedAt)
	if err != nil {
		return domain.Table{}, db.Error(err)
	}

	return dbTable.toDomain(), nil
}

// GetAllTables retrieves all tables from the database.
func (r Repository) GetAllTables(ctx context.Context) ([]domain.Table, error) {
	rows, err := r.DB.QueryContext(ctx, "SELECT id, name, status, created_at FROM tables WHERE status != 'deleted' ORDER BY id ASC")
	if err != nil {
		return nil, db.Error(err)
	}
	defer db.Close(rows, "tables")

	tables := []domain.Table{}
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

// CreateTable creates a new table in the database.
func (r Repository) CreateTable(ctx context.Context, t domain.Table) (int, error) {
	var id int
	err := r.DB.QueryRowContext(ctx, "INSERT INTO tables (name, status, created_at) VALUES ($1, $2, $3) RETURNING id", t.Name, t.Status, t.CreatedAt).Scan(&id)
	if err != nil {
		return 0, db.Error(err)
	}

	return id, nil
}

// UpdateTable updates an existing table in the database.
func (r Repository) UpdateTable(ctx context.Context, t domain.Table) error {
	result, err := r.DB.ExecContext(ctx, "UPDATE tables SET name = $1, status = $2 WHERE id = $3", t.Name, t.Status, t.ID)
	if err != nil {
		return db.Error(err)
	}

	return db.ResultError(result)
}
