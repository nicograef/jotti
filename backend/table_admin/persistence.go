package table_admin

import (
	"context"
	"database/sql"

	"github.com/nicograef/jotti/backend/db"
	"github.com/rs/zerolog"
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

// GetAllTables retrieves all tables from the database.
func (p *Persistence) GetAllTables(ctx context.Context) ([]Table, error) {
	log := zerolog.Ctx(ctx)

	rows, err := p.DB.QueryContext(ctx, "SELECT id, name, status, created_at FROM tables WHERE status != 'deleted' ORDER BY id ASC")
	if err != nil {
		log.Error().Err(err).Msg("DB Error querying all tables")
		return nil, db.Error(err)
	}
	defer db.Close(rows, "tables", log)

	tables := []Table{}
	for rows.Next() {
		var dbTable dbtable
		if err := rows.Scan(&dbTable.ID, &dbTable.Name, &dbTable.Status, &dbTable.CreatedAt); err != nil {
			log.Error().Err(err).Msg("DB Error scanning table row")
			return nil, db.Error(err)
		}

		tables = append(tables, Table{
			ID:        dbTable.ID,
			Name:      dbTable.Name,
			Status:    Status(dbTable.Status),
			CreatedAt: dbTable.CreatedAt.Time,
		})
	}

	if err := rows.Err(); err != nil {
		log.Error().Err(err).Msg("DB Error iterating over table rows")
		return nil, db.Error(err)
	}

	return tables, nil
}

// CreateTable creates a new table in the database.
func (p *Persistence) CreateTable(ctx context.Context, name string) (int, error) {
	log := zerolog.Ctx(ctx)

	var id int
	err := p.DB.QueryRowContext(ctx, "INSERT INTO tables (name) VALUES ($1) RETURNING id", name).Scan(&id)

	if err != nil {
		log.Error().Err(err).Str("table_name", name).Msg("DB Error creating table")
		return 0, db.Error(err)
	}

	return id, nil
}

// UpdateTable updates an existing table in the database.
func (p *Persistence) UpdateTable(ctx context.Context, id int, name string) error {
	log := zerolog.Ctx(ctx)

	result, err := p.DB.ExecContext(ctx, "UPDATE tables SET name = $1 WHERE id = $2", name, id)
	if err != nil {
		log.Error().Err(err).Int("table_id", id).Msg("DB Error updating table")
		return db.Error(err)
	}

	return db.ResultError(result)
}

// ActivateTable sets the status of a table to active.
func (p *Persistence) ActivateTable(ctx context.Context, id int) error {
	log := zerolog.Ctx(ctx)

	result, err := p.DB.ExecContext(ctx, "UPDATE tables SET status = 'active' WHERE id = $1", id)
	if err != nil {
		log.Error().Err(err).Int("table_id", id).Msg("DB Error activating table")
		return db.Error(err)
	}

	return db.ResultError(result)
}

// DeactivateTable sets the status of a table to inactive.
func (p *Persistence) DeactivateTable(ctx context.Context, id int) error {
	log := zerolog.Ctx(ctx)

	result, err := p.DB.ExecContext(ctx, "UPDATE tables SET status = 'inactive' WHERE id = $1", id)
	if err != nil {
		log.Error().Err(err).Int("table_id", id).Msg("DB Error deactivating table")
		return db.Error(err)
	}

	return db.ResultError(result)
}
