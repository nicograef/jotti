package application

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/internal/logging"
)

type produktQueryRepo interface {
	GetAllProdukte(ctx context.Context) ([]produkt.Produkt, error)
	GetActiveProdukte(ctx context.Context) ([]produkt.Produkt, error)
}

type Query struct {
	ProduktRepo produktQueryRepo
}

func (q Query) GetAllProdukte(ctx context.Context) ([]produkt.Produkt, error) {
	log := logging.FromContext(ctx)

	produkte, err := q.ProduktRepo.GetAllProdukte(ctx)
	if err != nil {
		log.Error("Failed to retrieve all produkte", "error", err)
		return nil, ErrDatabase
	}

	log.Info("Retrieved all produkte", "count", len(produkte))
	return produkte, nil
}

// AktiveProdukte: PfandProdukte is the subset of Produkte with the Pfand flag, offered by the Pfandrückgabe.
type AktiveProdukte struct {
	Produkte      []produkt.Produkt
	PfandProdukte []produkt.Produkt
}

func (q Query) GetActiveProdukte(ctx context.Context) (AktiveProdukte, error) {
	log := logging.FromContext(ctx)

	produkte, err := q.ProduktRepo.GetActiveProdukte(ctx)
	if err != nil {
		log.Error("Failed to retrieve active produkte", "error", err)
		return AktiveProdukte{}, ErrDatabase
	}

	pfandProdukte := []produkt.Produkt{}
	for _, p := range produkte {
		if p.Pfand {
			pfandProdukte = append(pfandProdukte, p)
		}
	}

	log.Info("Retrieved active produkte", "count", len(produkte), "pfand_count", len(pfandProdukte))
	return AktiveProdukte{Produkte: produkte, PfandProdukte: pfandProdukte}, nil
}
