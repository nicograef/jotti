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
	Locked    bool         `db:"locked"`
	CreatedAt sql.NullTime `db:"created_at"`
}

// GetTable retrieves a table from the database by its ID.
func (p *TablePersistence) GetTable(id int) (*table.Table, error) {
	row := p.DB.QueryRow("SELECT id, name, locked, created_at FROM tables WHERE id = $1", id)

	var dbTable dbtable
	if err := row.Scan(&dbTable.ID, &dbTable.Name, &dbTable.Locked, &dbTable.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, table.ErrTableNotFound
		}
		return nil, err
	}

	return &table.Table{
		ID:        dbTable.ID,
		Name:      dbTable.Name,
		Locked:    dbTable.Locked,
		CreatedAt: dbTable.CreatedAt.Time,
	}, nil
}

// GetAllTables retrieves all tables from the database.
func (p *TablePersistence) GetAllTables() ([]*table.Table, error) {
	rows, err := p.DB.Query("SELECT id, name, locked, created_at FROM tables")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []*table.Table
	for rows.Next() {
		var dbTable dbtable
		if err := rows.Scan(&dbTable.ID, &dbTable.Name, &dbTable.Locked, &dbTable.CreatedAt); err != nil {
			return nil, err
		}

		tables = append(tables, &table.Table{
			ID:        dbTable.ID,
			Name:      dbTable.Name,
			Locked:    dbTable.Locked,
			CreatedAt: dbTable.CreatedAt.Time,
		})
	}

	return tables, nil
}

// CreateTable creates a new table in the database.
func (p *TablePersistence) CreateTable(name string) (int, error) {
	var id int
	err := p.DB.QueryRow("INSERT INTO tables (name, locked, created_at) VALUES ($1, $2, NOW()) RETURNING id", name, false).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// UpdateTable updates an existing table in the database.
func (p *TablePersistence) UpdateTable(id int, name string, locked bool) error {
	result, err := p.DB.Exec("UPDATE tables SET name = $1, locked = $2 WHERE id = $3", name, locked, id)
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
