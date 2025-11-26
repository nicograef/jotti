package table

import (
	"context"
	"database/sql"

	"github.com/rs/zerolog/log"
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
func (p *Persistence) GetTable(ctx context.Context, id int) (*Table, error) {
	correlationID, _ := ctx.Value("correlation_id").(string)
	row := p.DB.QueryRowContext(ctx, "SELECT id, name, status, created_at FROM tables WHERE id = $1 AND status != 'deleted'", id)

	var dbTable dbtable
	if err := row.Scan(&dbTable.ID, &dbTable.Name, &dbTable.Status, &dbTable.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			log.Warn().Str("correlation_id", correlationID).Int("table_id", id).Msg("Table not found")
			return nil, ErrTableNotFound
		}
		log.Error().Str("correlation_id", correlationID).Err(err).Int("table_id", id).Msg("Failed to scan table")
		return nil, err
	}

	log.Debug().Str("correlation_id", correlationID).Int("table_id", id).Msg("Table retrieved")
	return &Table{
		ID:        dbTable.ID,
		Name:      dbTable.Name,
		Status:    Status(dbTable.Status),
		CreatedAt: dbTable.CreatedAt.Time,
	}, nil
}

// GetAllTables retrieves all tables from the database.
func (p *Persistence) GetAllTables(ctx context.Context) ([]*Table, error) {
	correlationID, _ := ctx.Value("correlation_id").(string)
	rows, err := p.DB.QueryContext(ctx, "SELECT id, name, status, created_at FROM tables WHERE status != 'deleted' ORDER BY id ASC")
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

	log.Debug().Str("correlation_id", correlationID).Int("count", len(tables)).Msg("Retrieved all tables")
	return tables, nil
}

// GetActiveTables retrieves all active tables from the database.
func (p *Persistence) GetActiveTables(ctx context.Context) ([]*TablePublic, error) {
	correlationID, _ := ctx.Value("correlation_id").(string)
	rows, err := p.DB.QueryContext(ctx, "SELECT id, name FROM tables WHERE status = 'active' ORDER BY name ASC")
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

	log.Debug().Str("correlation_id", correlationID).Int("count", len(tables)).Msg("Retrieved active tables")
	return tables, nil
}

// CreateTable creates a new table in the database.
func (p *Persistence) CreateTable(ctx context.Context, name string) (int, error) {
	correlationID, _ := ctx.Value("correlation_id").(string)
	var id int
	err := p.DB.QueryRowContext(ctx, "INSERT INTO tables (name) VALUES ($1) RETURNING id", name).Scan(&id)
	if err != nil {
		log.Error().Str("correlation_id", correlationID).Err(err).Str("name", name).Msg("Failed to create table")
		return 0, err
	}
	log.Info().Str("correlation_id", correlationID).Int("table_id", id).Str("name", name).Msg("Table created")
	return id, nil
}

// UpdateTable updates an existing table in the database.
func (p *Persistence) UpdateTable(ctx context.Context, id int, name string) error {
	correlationID, _ := ctx.Value("correlation_id").(string)
	result, err := p.DB.ExecContext(ctx, "UPDATE tables SET name = $1 WHERE id = $2", name, id)
	if err != nil {
		log.Error().Str("correlation_id", correlationID).Err(err).Int("table_id", id).Msg("Failed to update table")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		log.Warn().Str("correlation_id", correlationID).Int("table_id", id).Msg("Table not found for update")
		return ErrTableNotFound
	}

	log.Info().Str("correlation_id", correlationID).Int("table_id", id).Msg("Table updated")
	return nil
}

// ActivateTable sets the status of a table to active.
func (p *Persistence) ActivateTable(ctx context.Context, id int) error {
	correlationID, _ := ctx.Value("correlation_id").(string)
	result, err := p.DB.ExecContext(ctx, "UPDATE tables SET status = 'active' WHERE id = $1", id)
	if err != nil {
		log.Error().Str("correlation_id", correlationID).Err(err).Int("table_id", id).Msg("Failed to activate table")
		return err
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		log.Warn().Str("correlation_id", correlationID).Int("table_id", id).Msg("Table not found for activation")
		return ErrTableNotFound
	}

	log.Info().Str("correlation_id", correlationID).Int("table_id", id).Msg("Table activated")
	return nil
}

// DeactivateTable sets the status of a table to inactive.
func (p *Persistence) DeactivateTable(ctx context.Context, id int) error {
	correlationID, _ := ctx.Value("correlation_id").(string)
	result, err := p.DB.ExecContext(ctx, "UPDATE tables SET status = 'inactive' WHERE id = $1", id)
	if err != nil {
		log.Error().Str("correlation_id", correlationID).Err(err).Int("table_id", id).Msg("Failed to deactivate table")
		return err
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		log.Debug().Str("correlation_id", correlationID).Int("table_id", id).Msg("Table not found for deactivation")
		return ErrTableNotFound
	}

	log.Info().Str("correlation_id", correlationID).Int("table_id", id).Msg("Table deactivated")
	return nil
}
