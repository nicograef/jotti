package application

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/nicograef/jotti/backend/repository/tse_repo"
)

type tseSignaturauftragQueryRepo interface {
	GetTSESignaturauftraege(ctx context.Context) ([]tse_repo.Signaturauftrag, error)
	GetTSESignaturQueueZustand(ctx context.Context) (tse_repo.SignaturQueueZustand, error)
	GetAlleTSEStoerungen(ctx context.Context) ([]tse_repo.Stoerungszeitraum, error)
}

type Query struct {
	TSERepo tseSignaturauftragQueryRepo
}

// GetTSESignaturauftraege liefert die Signaturauftraege fuer die
// Signaturauftrags-Verwaltung.
func (q Query) GetTSESignaturauftraege(ctx context.Context) ([]tse_repo.Signaturauftrag, error) {
	log := zerolog.Ctx(ctx)

	auftraege, err := q.TSERepo.GetTSESignaturauftraege(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve tse signaturauftraege")
		return nil, ErrDatabase
	}

	log.Info().Int("count", len(auftraege)).Msg("Retrieved tse signaturauftraege")
	return auftraege, nil
}

// GetTSESignaturQueueZustand liefert den Zustand der Signatur-Queue fuer das
// Admin-Monitoring (Rueckstand und Leistung ueber ein 15-Minuten-Fenster).
func (q Query) GetTSESignaturQueueZustand(ctx context.Context) (tse_repo.SignaturQueueZustand, error) {
	log := zerolog.Ctx(ctx)

	zustand, err := q.TSERepo.GetTSESignaturQueueZustand(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve tse signatur queue zustand")
		return tse_repo.SignaturQueueZustand{}, ErrDatabase
	}

	return zustand, nil
}

// GetTSEStoerungen liefert das Stoerungsprotokoll (Ausfalldokumentation):
// die Stoerungszeitraeume mit Beginn, Ende und Grund.
func (q Query) GetTSEStoerungen(ctx context.Context) ([]tse_repo.Stoerungszeitraum, error) {
	log := zerolog.Ctx(ctx)

	stoerungen, err := q.TSERepo.GetAlleTSEStoerungen(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve tse stoerungen")
		return nil, ErrDatabase
	}

	log.Info().Int("count", len(stoerungen)).Msg("Retrieved tse stoerungen")
	return stoerungen, nil
}
