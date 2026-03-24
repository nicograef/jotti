package application

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/rs/zerolog"
)

type kassenQueryRepo interface {
	GetOffeneKassensitzung(ctx context.Context) (*kasse.KassensitzungState, error)
	GetKassenbestand(ctx context.Context, kassensitzungNr int) (int, error)
}

type Query struct {
	KassenRepo kassenQueryRepo
}

// GetOffeneKassensitzung returns the currently open Kassensitzung or nil if none exists.
func (q Query) GetOffeneKassensitzung(ctx context.Context) (*kasse.KassensitzungState, error) {
	log := zerolog.Ctx(ctx)

	ks, err := q.KassenRepo.GetOffeneKassensitzung(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get offene Kassensitzung")
		return nil, ErrDatabase
	}

	log.Debug().Msg("Retrieved offene Kassensitzung")
	return ks, nil
}

// GetKassenbestand returns the Soll-Kassenbestand for the given Kassensitzung.
func (q Query) GetKassenbestand(ctx context.Context, kassensitzungNr int) (int, error) {
	log := zerolog.Ctx(ctx)

	bestand, err := q.KassenRepo.GetKassenbestand(ctx, kassensitzungNr)
	if err != nil {
		log.Error().Err(err).Int("z_nr", kassensitzungNr).Msg("Failed to get Kassenbestand")
		return 0, ErrDatabase
	}

	log.Debug().Int("z_nr", kassensitzungNr).Int("bestand_cents", bestand).Msg("Retrieved Kassenbestand")
	return bestand, nil
}
