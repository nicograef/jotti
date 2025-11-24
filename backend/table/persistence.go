package table

import (
	"database/sql"
)

// Persistence implements table persistence layer using a SQL database.
type Persistence struct {
	DB *sql.DB
}

type dbtable struct {
	ID        int          `db:"id"`
	Name      string       `db:"name"`
	Status    string       `db:"status"`
	CreatedAt sql.NullTime `db:"created_at"`
}

// GetTable retrieves a table from the database by its ID.
func (p *Persistence) GetTable(id int) (*Table, error) {
	row := p.DB.QueryRow("SELECT id, name, status, created_at FROM tables WHERE id = $1 AND status != 'deleted'", id)

	var dbTable dbtable
	if err := row.Scan(&dbTable.ID, &dbTable.Name, &dbTable.Status, &dbTable.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrTableNotFound
		}
		return nil, err
	}

	return &Table{
		ID:        dbTable.ID,
		Name:      dbTable.Name,
		Status:    Status(dbTable.Status),
		CreatedAt: dbTable.CreatedAt.Time,
	}, nil
}

// GetAllTables retrieves all tables from the database.
func (p *Persistence) GetAllTables() ([]*Table, error) {
	rows, err := p.DB.Query("SELECT id, name, status, created_at FROM tables WHERE status != 'deleted' ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			err = closeErr
		}
	}()

	var tables []*Table
	for rows.Next() {
		var dbTable dbtable
		if err := rows.Scan(&dbTable.ID, &dbTable.Name, &dbTable.Status, &dbTable.CreatedAt); err != nil {
			return nil, err
		}

		tables = append(tables, &Table{
			ID:        dbTable.ID,
			Name:      dbTable.Name,
			Status:    Status(dbTable.Status),
			CreatedAt: dbTable.CreatedAt.Time,
		})
	}

	return tables, nil
}

// GetActiveTables retrieves all active tables from the database.
func (p *Persistence) GetActiveTables() ([]*TablePublic, error) {
	rows, err := p.DB.Query("SELECT id, name FROM tables WHERE status = 'active' ORDER BY name ASC")
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			err = closeErr
		}
	}()

	var tables []*TablePublic
	for rows.Next() {
		var dbTable dbtable
		if err := rows.Scan(&dbTable.ID, &dbTable.Name); err != nil {
			return nil, err
		}

		tables = append(tables, &TablePublic{
			ID:   dbTable.ID,
			Name: dbTable.Name,
		})
	}

	return tables, nil
}

// CreateTable creates a new table in the database.
func (p *Persistence) CreateTable(name string) (int, error) {
	var id int
	err := p.DB.QueryRow("INSERT INTO tables (name) VALUES ($1) RETURNING id", name).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// UpdateTable updates an existing table in the database.
func (p *Persistence) UpdateTable(id int, name string) error {
	result, err := p.DB.Exec("UPDATE tables SET name = $1 WHERE id = $2", name, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrTableNotFound
	}

	return nil
}

// ActivateTable sets the status of a table to active.
func (p *Persistence) ActivateTable(id int) error {
	result, err := p.DB.Exec("UPDATE tables SET status = 'active' WHERE id = $1", id)
	if err != nil {
		return err
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		return ErrTableNotFound
	}

	return nil
}

// DeactivateTable sets the status of a table to inactive.
func (p *Persistence) DeactivateTable(id int) error {
	result, err := p.DB.Exec("UPDATE tables SET status = 'inactive' WHERE id = $1", id)
	if err != nil {
		return err
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		return ErrTableNotFound
	}

	return nil
}
