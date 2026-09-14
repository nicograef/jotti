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

// GetAktiveKassensitzung returns the Kassensitzung in status 'offen' or 'wird_abgeschlossen', nil if
// none — the barrier included, so an interrupted Kassenabschluss does not read as a closed Kasse.
func (q Query) GetAktiveKassensitzung(ctx context.Context) (*kasse.Kassensitzung, error) {
	log := zerolog.Ctx(ctx)

	ks, err := q.KassensitzungenRepo.GetAktiveKassensitzung(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get aktive Kassensitzung")
		return nil, ErrDatabase
	}

	log.Debug().Msg("Retrieved aktive Kassensitzung")
	return ks, nil
}

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

// GetGeldtransitListe returns the Einlagen/Entnahmen of the Kassensitzung, newest first.
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
