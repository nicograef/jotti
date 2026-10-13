package application

import (
	"context"

	"github.com/nicograef/jotti/backend/internal/logging"

	"github.com/nicograef/jotti/backend/domain/druckstation"
)

type druckstationQueryRepo interface {
	GetAlleDruckstationen(ctx context.Context) ([]druckstation.Druckstation, error)
}

type Query struct {
	DruckstationRepo druckstationQueryRepo
}

func (q Query) GetAlleDruckstationen(ctx context.Context) ([]druckstation.Druckstation, error) {
	log := logging.FromContext(ctx)

	konfigs, err := q.DruckstationRepo.GetAlleDruckstationen(ctx)
	if err != nil {
		log.Error("Failed to retrieve druckstationen", "error", err)
		return nil, ErrDatabase
	}

	log.Info("Retrieved druckstationen", "count", len(konfigs))
	return konfigs, nil
}
