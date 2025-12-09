package table_admin

import (
	"context"

	"github.com/rs/zerolog"
)

type queryPersistence interface {
	GetAllTables(ctx context.Context) ([]Table, error)
}

type Query struct {
	Persistence queryPersistence
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
