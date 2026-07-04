package application

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/nicograef/jotti/backend/domain/druckstation"
)

type druckstationCommandRepo interface {
	UpsertDruckstation(ctx context.Context, station druckstation.Druckstation) error
}

type Command struct {
	DruckstationRepo druckstationCommandRepo
}

func (c Command) UpsertDruckstation(ctx context.Context, kategorie, druckerIP, bonmodus string) error {
	log := zerolog.Ctx(ctx)

	station, err := druckstation.NewDruckstation(
		druckstation.Kategorie(kategorie),
		druckerIP,
		druckstation.Bonmodus(bonmodus),
	)
	if err != nil {
		log.Warn().Err(err).Str("kategorie", kategorie).Msg("Invalid druckstation")
		return ErrUngueltigeDruckstation
	}

	if err := c.DruckstationRepo.UpsertDruckstation(ctx, station); err != nil {
		log.Error().Err(err).Str("kategorie", kategorie).Msg("Failed to upsert druckstation")
		return ErrDatabase
	}

	log.Info().Str("kategorie", kategorie).Str("druckerIP", druckerIP).Msg("Druckstation updated")
	return nil
}
