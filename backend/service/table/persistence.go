package table

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

// GetTable retrieves a table from the database by its ID.
func (p *Persistence) GetTable(ctx context.Context, id int) (*Table, error) {
	log := zerolog.Ctx(ctx)

	row := p.DB.QueryRowContext(ctx, "SELECT id, name FROM tables WHERE status = 'active' AND id = $1", id)

	var table Table
	if err := row.Scan(&table.ID, &table.Name); err != nil {
		log.Error().Err(err).Int("table_id", id).Msg("DB Error scanning table row")
		return nil, db.Error(err)
	}

	return &table, nil
}

// GetAllTables retrieves all active tables from the database.
func (p *Persistence) GetAllTables(ctx context.Context) ([]Table, error) {
	log := zerolog.Ctx(ctx)

	rows, err := p.DB.QueryContext(ctx, "SELECT id, name FROM tables WHERE status = 'active'")
	if err != nil {
		log.Error().Err(err).Msg("DB Error querying active tables")
		return nil, db.Error(err)
	}
	defer db.Close(rows, "tables", log)

	tables := []Table{}
	for rows.Next() {
		var table Table
		if err := rows.Scan(&table.ID, &table.Name); err != nil {
			log.Error().Err(err).Msg("DB Error scanning active table row")
			return nil, db.Error(err)
		}

		tables = append(tables, table)
	}

	if err := rows.Err(); err != nil {
		log.Error().Err(err).Msg("DB Error iterating over table rows")
		return nil, db.Error(err)
	}

	return tables, nil
}
