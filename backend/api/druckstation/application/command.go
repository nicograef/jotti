package application

import (
	"context"

	"github.com/rs/zerolog"
)

type druckstationCommandRepo interface {
	UpsertDruckstation(ctx context.Context, kategorie, druckerIP, bonmodus string) error
}

type Command struct {
	DruckstationRepo druckstationCommandRepo
}

func (c Command) UpsertDruckstation(ctx context.Context, kategorie, druckerIP, bonmodus string) error {
	log := zerolog.Ctx(ctx)

	err := c.DruckstationRepo.UpsertDruckstation(ctx, kategorie, druckerIP, bonmodus)
	if err != nil {
		log.Error().Err(err).Str("kategorie", kategorie).Msg("Failed to upsert druckstation")
		return ErrDatabase
	}

	log.Info().Str("kategorie", kategorie).Str("druckerIP", druckerIP).Msg("Druckstation updated")
	return nil
}
