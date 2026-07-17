package application

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/nicograef/jotti/backend/domain/tse"
)

type tseSignaturauftragQueryRepo interface {
	GetTSESignaturQueueZustand(ctx context.Context) (tse.SignaturQueueZustand, error)
	GetAlleTSEStoerungen(ctx context.Context) ([]tse.Stoerungszeitraum, error)
}

type Query struct {
	TSERepo tseSignaturauftragQueryRepo
}

// GetTSESignaturQueueZustand liefert den Zustand der Signatur-Queue fuer das
// Admin-Monitoring (Rueckstand und Leistung ueber ein 15-Minuten-Fenster).
func (q Query) GetTSESignaturQueueZustand(ctx context.Context) (tse.SignaturQueueZustand, error) {
	log := zerolog.Ctx(ctx)

	zustand, err := q.TSERepo.GetTSESignaturQueueZustand(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve tse signatur queue zustand")
		return tse.SignaturQueueZustand{}, ErrDatabase
	}

	return zustand, nil
}

// GetTSEStoerungen liefert das Stoerungsprotokoll (Ausfalldokumentation):
// die Stoerungszeitraeume mit Beginn, Ende und Grund.
func (q Query) GetTSEStoerungen(ctx context.Context) ([]tse.Stoerungszeitraum, error) {
	log := zerolog.Ctx(ctx)

	stoerungen, err := q.TSERepo.GetAlleTSEStoerungen(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve tse stoerungen")
		return nil, ErrDatabase
	}

	log.Info().Int("count", len(stoerungen)).Msg("Retrieved tse stoerungen")
	return stoerungen, nil
}
