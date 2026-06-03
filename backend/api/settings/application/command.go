package application

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/settings"
	"github.com/rs/zerolog"
)

type settingsCommandRepo interface {
	UpsertBetreiber(ctx context.Context, b settings.Betreiber) error
}

type Command struct {
	SettingsRepo settingsCommandRepo
}

func (c Command) UpdateBetreiber(ctx context.Context, b settings.Betreiber) error {
	log := zerolog.Ctx(ctx)

	if err := c.SettingsRepo.UpsertBetreiber(ctx, b); err != nil {
		log.Error().Err(err).Msg("Failed to save betreiber")
		return ErrDatabase
	}
	log.Info().Str("vereinsname", b.Vereinsname).Msg("Betreiber saved")
	return nil
}
