package persistence

import (
	"database/sql"

	"github.com/nicograef/jotti/backend/domain/table"
)

// UserPersistence implements user persistence layer using a SQL database.
type TablePersistence struct {
	DB *sql.DB
}

type dbtable struct {
	ID        int          `db:"id"`
	Name      string       `db:"name"`
	Status    string       `db:"status"`
	CreatedAt sql.NullTime `db:"created_at"`
}

// GetTable retrieves a table from the database by its ID.
func (p *TablePersistence) GetTable(id int) (*table.Table, error) {
	row := p.DB.QueryRow("SELECT id, name, status, created_at FROM tables WHERE id = $1 AND status != 'deleted'", id)

	var dbTable dbtable
	if err := row.Scan(&dbTable.ID, &dbTable.Name, &dbTable.Status, &dbTable.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, table.ErrTableNotFound
		}
		return nil, err
	}

	return &table.Table{
		ID:        dbTable.ID,
		Name:      dbTable.Name,
		Status:    table.Status(dbTable.Status),
		CreatedAt: dbTable.CreatedAt.Time,
	}, nil
}

// GetAllTables retrieves all tables from the database.
func (p *TablePersistence) GetAllTables() ([]*table.Table, error) {
	rows, err := p.DB.Query("SELECT id, name, status, created_at FROM tables WHERE status != 'deleted' ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			err = closeErr
		}
	}()

	var tables []*table.Table
	for rows.Next() {
		var dbTable dbtable
		if err := rows.Scan(&dbTable.ID, &dbTable.Name, &dbTable.Status, &dbTable.CreatedAt); err != nil {
			return nil, err
		}

		tables = append(tables, &table.Table{
			ID:        dbTable.ID,
			Name:      dbTable.Name,
			Status:    table.Status(dbTable.Status),
			CreatedAt: dbTable.CreatedAt.Time,
		})
	}

	return tables, nil
}

// CreateTable creates a new table in the database.
func (p *TablePersistence) CreateTable(name string) (int, error) {
	var id int
	err := p.DB.QueryRow("INSERT INTO tables (name) VALUES ($1) RETURNING id", name).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// UpdateTable updates an existing table in the database.
func (p *TablePersistence) UpdateTable(id int, name string) error {
	result, err := p.DB.Exec("UPDATE tables SET name = $1 WHERE id = $2", name, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return table.ErrTableNotFound
	}

	return nil
}
