package application

import (
	"context"

	"github.com/rs/zerolog"
)

type Query struct {
	DruckauftragRepo druckauftragRepo
}

type druckauftragRepo interface {
	GetOffeneDruckauftraege(ctx context.Context) ([]DruckAuftrag, error)
}

type DruckAuftrag struct {
	ID      int
	ZielIP  string
	Payload string
}

func (q Query) GetOffeneDruckauftraege(ctx context.Context) ([]DruckAuftrag, error) {
	log := zerolog.Ctx(ctx)

	auftraege, err := q.DruckauftragRepo.GetOffeneDruckauftraege(ctx)
	if err != nil {
		return nil, err
	}

	log.Debug().Int("offene_auftraege", len(auftraege)).Msg("Relay poll")
	return auftraege, nil
}
