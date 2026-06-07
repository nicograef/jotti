package application

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/nicograef/jotti/backend/repository/druckstation_repo"
)

type druckstationQueryRepo interface {
	GetAlleDruckstationen(ctx context.Context) ([]druckstation_repo.Druckstation, error)
}

type Query struct {
	DruckstationRepo druckstationQueryRepo
}

func (q Query) GetAlleDruckstationen(ctx context.Context) ([]druckstation_repo.Druckstation, error) {
	log := zerolog.Ctx(ctx)

	konfigs, err := q.DruckstationRepo.GetAlleDruckstationen(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve druckstationen")
		return nil, ErrDatabase
	}

	log.Info().Int("count", len(konfigs)).Msg("Retrieved druckstationen")
	return konfigs, nil
}
