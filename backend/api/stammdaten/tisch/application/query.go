package application

import (
	"context"

	t "github.com/nicograef/jotti/backend/domain/tisch"
	"github.com/rs/zerolog"
)

type Query struct {
	TableRepo tableRepo
}

func (q Query) GetAllTische(ctx context.Context) ([]t.Tisch, error) {
	log := zerolog.Ctx(ctx)

	tische, err := q.TableRepo.GetAllTables(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve all tische")
		return nil, ErrDatabase
	}

	log.Debug().Int("count", len(tische)).Msg("Retrieved all tische")
	return tische, nil
}
