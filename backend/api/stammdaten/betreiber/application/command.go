package application

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/betreiber"
	"github.com/rs/zerolog"
)

type betreiberCommandRepo interface {
	UpsertBetreiber(ctx context.Context, b betreiber.Betreiber) error
}

type Command struct {
	BetreiberRepo betreiberCommandRepo
}

func (c Command) UpdateBetreiber(ctx context.Context, b betreiber.Betreiber) error {
	log := zerolog.Ctx(ctx)

	if err := c.BetreiberRepo.UpsertBetreiber(ctx, b); err != nil {
		log.Error().Err(err).Msg("Failed to save betreiber")
		return ErrDatabase
	}
	log.Info().Str("vereinsname", b.Vereinsname).Msg("Betreiber saved")
	return nil
}
