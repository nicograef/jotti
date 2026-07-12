package application

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/rs/zerolog"
)

type Query struct {
	KassenjournalRepo   kassenjournalRepo
	KassensitzungenRepo kassensitzungenRepo
}

// GetOffeneKassensitzung returns the currently open Kassensitzung or nil if none exists.
func (q Query) GetOffeneKassensitzung(ctx context.Context) (*kasse.Kassensitzung, error) {
	log := zerolog.Ctx(ctx)

	ks, err := q.KassensitzungenRepo.GetOffeneKassensitzung(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get offene Kassensitzung")
		return nil, ErrDatabase
	}

	log.Debug().Msg("Retrieved offene Kassensitzung")
	return ks, nil
}

// GetKassenbestand returns the Soll-Kassenbestand for the given Kassensitzung
// together with its four components (Anfangsbestand, Bareinnahmen, Einlagen, Entnahmen).
func (q Query) GetKassenbestand(ctx context.Context, kassensitzungNr int) (kasse.Kassenbestand, error) {
	log := zerolog.Ctx(ctx)

	bestand, err := q.KassenjournalRepo.GetKassenbestand(ctx, kassensitzungNr)
	if err != nil {
		log.Error().Err(err).Int("z_nr", kassensitzungNr).Msg("Failed to get Kassenbestand")
		return kasse.Kassenbestand{}, ErrDatabase
	}

	log.Debug().Int("z_nr", kassensitzungNr).Int("bestand_cents", bestand.SollBestandCents).Msg("Retrieved Kassenbestand")
	return bestand, nil
}

// GetGeldtransitListe returns all Geldbewegungen (Einlagen/Entnahmen) of the given
// Kassensitzung, newest first — a pure projection of the geldtransit-gebucht:v1 events.
func (q Query) GetGeldtransitListe(ctx context.Context, kassensitzungNr int) ([]kasse.Geldtransit, error) {
	log := zerolog.Ctx(ctx)

	buchungen, err := q.KassenjournalRepo.GetGeldtransitListe(ctx, kassensitzungNr)
	if err != nil {
		log.Error().Err(err).Int("z_nr", kassensitzungNr).Msg("Failed to get Geldtransit-Liste")
		return nil, ErrDatabase
	}

	log.Debug().Int("z_nr", kassensitzungNr).Int("anzahl", len(buchungen)).Msg("Retrieved Geldtransit-Liste")
	return buchungen, nil
}
