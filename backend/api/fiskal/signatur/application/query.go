package application

import (
	"context"

	"github.com/nicograef/jotti/backend/internal/logging"

	"github.com/nicograef/jotti/backend/domain/tse"
)

type tseSignaturauftragQueryRepo interface {
	GetTSESignaturQueueZustand(ctx context.Context) (tse.SignaturQueueZustand, error)
	GetAlleTSEStoerungen(ctx context.Context) ([]tse.Stoerungszeitraum, error)
}

type Query struct {
	TSERepo tseSignaturauftragQueryRepo
}

// GetTSESignaturQueueZustand reports backlog and throughput over a 15-minute window for admin monitoring.
func (q Query) GetTSESignaturQueueZustand(ctx context.Context) (tse.SignaturQueueZustand, error) {
	log := logging.FromContext(ctx)

	zustand, err := q.TSERepo.GetTSESignaturQueueZustand(ctx)
	if err != nil {
		log.Error("Failed to retrieve tse signatur queue zustand", "error", err)
		return tse.SignaturQueueZustand{}, ErrDatabase
	}

	return zustand, nil
}

// GetTSEStoerungen returns the Störungsprotokoll, the outage record.
func (q Query) GetTSEStoerungen(ctx context.Context) ([]tse.Stoerungszeitraum, error) {
	log := logging.FromContext(ctx)

	stoerungen, err := q.TSERepo.GetAlleTSEStoerungen(ctx)
	if err != nil {
		log.Error("Failed to retrieve tse stoerungen", "error", err)
		return nil, ErrDatabase
	}

	log.Info("Retrieved tse stoerungen", "count", len(stoerungen))
	return stoerungen, nil
}
