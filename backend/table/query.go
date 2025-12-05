package table

import (
	"context"
	"errors"

	"github.com/nicograef/jotti/backend/db"
	"github.com/rs/zerolog"
)

type queryPersistence interface {
	GetTable(ctx context.Context, id int) (*Table, error)
	GetAllTables(ctx context.Context) ([]Table, error)
	GetActiveTables(ctx context.Context) ([]TablePublic, error)
}

type Query struct {
	Persistence queryPersistence
}

// GetTable retrieves a table by its ID.
func (q *Query) GetTable(ctx context.Context, id int) (*Table, error) {
	log := zerolog.Ctx(ctx)

	table, err := q.Persistence.GetTable(ctx, id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("table_id", id).Msg("Table not found")
			return nil, ErrTableNotFound
		} else {
			log.Error().Int("table_id", id).Msg("Failed to retrieve table")
			return nil, ErrDatabase
		}
	}

	log.Debug().Int("table_id", id).Msg("Table retrieved")
	return table, nil
}

// GetAllTables retrieves all tables.
func (q *Query) GetAllTables(ctx context.Context) ([]Table, error) {
	log := zerolog.Ctx(ctx)

	tables, err := q.Persistence.GetAllTables(ctx)
	if err != nil {
		log.Error().Msg("Failed to retrieve all tables")
		return nil, ErrDatabase
	}

	log.Debug().Int("count", len(tables)).Msg("Retrieved all tables")
	return tables, nil
}

// GetActiveTables retrieves all active tables.
func (q *Query) GetActiveTables(ctx context.Context) ([]TablePublic, error) {
	log := zerolog.Ctx(ctx)

	tables, err := q.Persistence.GetActiveTables(ctx)
	if err != nil {
		log.Error().Msg("Failed to retrieve active tables")
		return nil, ErrDatabase
	}

	log.Debug().Int("count", len(tables)).Msg("Retrieved active tables")
	return tables, nil
}
