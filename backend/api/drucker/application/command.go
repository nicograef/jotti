package application

import (
	"context"

	"github.com/rs/zerolog"
)

type druckerCommandRepo interface {
	UpsertKategorieDrucker(ctx context.Context, kategorie, druckerIP, bonmodus string) error
}

type Command struct {
	DruckerRepo druckerCommandRepo
}

func (c Command) UpsertKategorieDrucker(ctx context.Context, kategorie, druckerIP, bonmodus string) error {
	log := zerolog.Ctx(ctx)

	err := c.DruckerRepo.UpsertKategorieDrucker(ctx, kategorie, druckerIP, bonmodus)
	if err != nil {
		log.Error().Err(err).Str("kategorie", kategorie).Msg("Failed to upsert drucker configuration")
		return ErrDatabase
	}

	log.Info().Str("kategorie", kategorie).Str("druckerIP", druckerIP).Msg("Drucker configuration updated")
	return nil
}
