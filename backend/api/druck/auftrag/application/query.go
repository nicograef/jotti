package application

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
)

type druckauftragQueryRepo interface {
	GetFehlgeschlageneDruckauftraege(ctx context.Context) ([]druckauftrag_repo.FehlgeschlagenerDruckauftrag, error)
}

type Query struct {
	DruckauftragRepo druckauftragQueryRepo
}

func (q Query) GetFehlgeschlageneDruckauftraege(ctx context.Context) ([]druckauftrag_repo.FehlgeschlagenerDruckauftrag, error) {
	log := zerolog.Ctx(ctx)

	auftraege, err := q.DruckauftragRepo.GetFehlgeschlageneDruckauftraege(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve fehlgeschlagene druckauftraege")
		return nil, ErrDatabase
	}

	log.Info().Int("count", len(auftraege)).Msg("Retrieved fehlgeschlagene druckauftraege")
	return auftraege, nil
}
