package application

import (
	"context"
	"errors"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/betreiber"
	"github.com/rs/zerolog"
)

type betreiberRepo interface {
	GetBetreiber(ctx context.Context) (betreiber.Betreiber, error)
}

type Query struct {
	BetreiberRepo betreiberRepo
}

func (q Query) GetBetreiber(ctx context.Context) (betreiber.Betreiber, error) {
	log := zerolog.Ctx(ctx)

	b, err := q.BetreiberRepo.GetBetreiber(ctx)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return betreiber.Betreiber{}, ErrNotFound
		}
		log.Error().Err(err).Msg("Failed to retrieve betreiber")
		return betreiber.Betreiber{}, ErrDatabase
	}
	return b, nil
}
