package application

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/rs/zerolog"
)

type produktQueryRepo interface {
	GetAllProdukte(ctx context.Context) ([]produkt.Produkt, error)
	GetActiveProdukte(ctx context.Context) ([]produkt.Produkt, error)
}

type Query struct {
	ProduktRepo produktQueryRepo
}

func (q Query) GetAllProdukte(ctx context.Context) ([]produkt.Produkt, error) {
	log := zerolog.Ctx(ctx)

	produkte, err := q.ProduktRepo.GetAllProdukte(ctx)
	if err != nil {
		log.Error().Msg("Failed to retrieve all produkte")
		return nil, ErrDatabase
	}

	log.Info().Int("count", len(produkte)).Msg("Retrieved all produkte")
	return produkte, nil
}

func (q Query) GetActiveProdukte(ctx context.Context) ([]produkt.Produkt, error) {
	log := zerolog.Ctx(ctx)

	produkte, err := q.ProduktRepo.GetActiveProdukte(ctx)
	if err != nil {
		log.Error().Msg("Failed to retrieve active produkte")
		return nil, ErrDatabase
	}

	log.Info().Int("count", len(produkte)).Msg("Retrieved active produkte")
	return produkte, nil
}
