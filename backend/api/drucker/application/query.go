package application

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/nicograef/jotti/backend/repository/drucker_repo"
)

type druckerQueryRepo interface {
	GetAlleKategorieDrucker(ctx context.Context) ([]drucker_repo.DruckerKonfig, error)
}

type Query struct {
	DruckerRepo druckerQueryRepo
}

func (q Query) GetAlleKategorieDrucker(ctx context.Context) ([]drucker_repo.DruckerKonfig, error) {
	log := zerolog.Ctx(ctx)

	konfigs, err := q.DruckerRepo.GetAlleKategorieDrucker(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve drucker configuration")
		return nil, ErrDatabase
	}

	log.Info().Int("count", len(konfigs)).Msg("Retrieved drucker configuration")
	return konfigs, nil
}
