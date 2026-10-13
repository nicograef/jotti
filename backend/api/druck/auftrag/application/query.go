package application

import (
	"context"

	"github.com/nicograef/jotti/backend/internal/logging"

	"github.com/nicograef/jotti/backend/domain/druckstation"
)

type druckauftragQueryRepo interface {
	GetFehlgeschlageneDruckauftraege(ctx context.Context) ([]druckstation.FehlgeschlagenerDruckauftrag, error)
}

type Query struct {
	DruckauftragRepo druckauftragQueryRepo
}

func (q Query) GetFehlgeschlageneDruckauftraege(ctx context.Context) ([]druckstation.FehlgeschlagenerDruckauftrag, error) {
	log := logging.FromContext(ctx)

	auftraege, err := q.DruckauftragRepo.GetFehlgeschlageneDruckauftraege(ctx)
	if err != nil {
		log.Error("Failed to retrieve fehlgeschlagene druckauftraege", "error", err)
		return nil, ErrDatabase
	}

	log.Info("Retrieved fehlgeschlagene druckauftraege", "count", len(auftraege))
	return auftraege, nil
}
