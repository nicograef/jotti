package application

import (
	"context"

	"github.com/nicograef/jotti/backend/admin/table/domain"
	"github.com/rs/zerolog"
)

type tableRepositoryQuery interface {
	GetAllTables(ctx context.Context) ([]domain.Table, error)
}

type Query struct {
	TableRepo tableRepositoryQuery
}

// GetAllTables retrieves all tables.
func (q Query) GetAllTables(ctx context.Context) ([]domain.Table, error) {
	log := zerolog.Ctx(ctx)

	tables, err := q.TableRepo.GetAllTables(ctx)
	if err != nil {
		log.Error().Msg("Failed to retrieve all tables")
		return nil, ErrDatabase
	}

	log.Debug().Int("count", len(tables)).Msg("Retrieved all tables")
	return tables, nil
}
