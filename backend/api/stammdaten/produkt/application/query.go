package application

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/rs/zerolog"
)

type produktQueryRepo interface {
	GetAllProducts(ctx context.Context) ([]produkt.Produkt, error)
	GetActiveProducts(ctx context.Context) ([]produkt.Produkt, error)
	GetProduktIDsMitVerkaeufen(ctx context.Context) (map[int]bool, error)
}

type Query struct {
	ProduktRepo produktQueryRepo
}

// ProduktMitVerkauf ergänzt ein Produkt um das Projektionsflag hatVerkaeufe
// (mindestens eine Variante wurde bereits verkauft). Das Flag ist keine
// Domäneneigenschaft, sondern eine Journal-Projektion, und lebt deshalb hier
// statt am Domain-Modell.
type ProduktMitVerkauf struct {
	Produkt      produkt.Produkt
	HatVerkaeufe bool
}

func (q Query) GetAllProducts(ctx context.Context) ([]ProduktMitVerkauf, error) {
	log := zerolog.Ctx(ctx)

	products, err := q.ProduktRepo.GetAllProducts(ctx)
	if err != nil {
		log.Error().Msg("Failed to retrieve all products")
		return nil, ErrDatabase
	}

	verkaufteIDs, err := q.ProduktRepo.GetProduktIDsMitVerkaeufen(ctx)
	if err != nil {
		log.Error().Msg("Failed to retrieve products with sales")
		return nil, ErrDatabase
	}

	result := make([]ProduktMitVerkauf, 0, len(products))
	for i := range products {
		result = append(result, ProduktMitVerkauf{
			Produkt:      products[i],
			HatVerkaeufe: verkaufteIDs[products[i].ID],
		})
	}

	log.Info().Int("count", len(result)).Msg("Retrieved all products")
	return result, nil
}

func (q Query) GetActiveProducts(ctx context.Context) ([]produkt.Produkt, error) {
	log := zerolog.Ctx(ctx)

	products, err := q.ProduktRepo.GetActiveProducts(ctx)
	if err != nil {
		log.Error().Msg("Failed to retrieve active products")
		return nil, ErrDatabase
	}

	log.Info().Int("count", len(products)).Msg("Retrieved active products")
	return products, nil
}
