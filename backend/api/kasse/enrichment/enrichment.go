// Package enrichment turns thin position inputs into fat kasse.Position values, shared by the
// tischgeschaeft and direktverkauf command paths.
package enrichment

import (
	"context"
	"errors"

	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/rs/zerolog"
)

type PositionInput struct {
	ProduktID  int
	VarianteID int
	Menge      int
}

type produktRepo interface {
	GetVariantenByIDs(ctx context.Context, ids []int) (map[int]produkt.VarianteMitProdukt, error)
	GetProdukteByIDs(ctx context.Context, ids []int) (map[int]produkt.Produkt, error)
}

var ErrProduktNotFound = errors.New("produkt not found")

// ErrVarianteNichtAktiv covers a deactivated variant or product; ErrProduktNotFound covers deleted or unknown IDs.
var ErrVarianteNichtAktiv = errors.New("variante nicht aktiv")

func EnrichPositionen(ctx context.Context, repo produktRepo, inputs []PositionInput) ([]kasse.Position, error) {
	log := zerolog.Ctx(ctx)

	varianteIDs := make([]int, 0, len(inputs))
	produktIDs := make([]int, 0, len(inputs))
	seenVarianten := make(map[int]bool, len(inputs))
	seenProdukte := make(map[int]bool, len(inputs))
	for _, input := range inputs {
		if !seenVarianten[input.VarianteID] {
			varianteIDs = append(varianteIDs, input.VarianteID)
			seenVarianten[input.VarianteID] = true
		}
		if !seenProdukte[input.ProduktID] {
			produktIDs = append(produktIDs, input.ProduktID)
			seenProdukte[input.ProduktID] = true
		}
	}

	variantenByID, err := repo.GetVariantenByIDs(ctx, varianteIDs)
	if err != nil {
		log.Error().Err(err).Msg("Failed to batch-fetch variants for position enrichment")
		return nil, ErrProduktNotFound
	}
	produkteByID, err := repo.GetProdukteByIDs(ctx, produktIDs)
	if err != nil {
		log.Error().Err(err).Msg("Failed to batch-fetch products for position enrichment")
		return nil, ErrProduktNotFound
	}

	positionen := make([]kasse.Position, 0, len(inputs))
	for _, input := range inputs {
		varianteMitProdukt, ok := variantenByID[input.VarianteID]
		if !ok {
			log.Error().Int("variante_id", input.VarianteID).Msg("Variant not found in batch result")
			return nil, ErrProduktNotFound
		}
		prod, ok := produkteByID[input.ProduktID]
		if !ok {
			log.Error().Int("produkt_id", input.ProduktID).Msg("Product not found in batch result")
			return nil, ErrProduktNotFound
		}

		// Die Paarung Produkt/Variante kommt vom Client: ohne diese Prüfung erbte die Position
		// Kategorie und Steuersatz eines fremden Produkts, den Preis aber von der Variante.
		if varianteMitProdukt.ProduktID != input.ProduktID {
			log.Error().Int("variante_id", input.VarianteID).Int("produkt_id", input.ProduktID).Msg("Variant does not belong to the referenced product")
			return nil, ErrProduktNotFound
		}
		variante := varianteMitProdukt.Variante

		// Defense-in-Depth: deaktivierte (inactive) Varianten/Produkte tauchen im
		// Menü nicht auf, könnten aber per direktem POST referenziert werden.
		if variante.Status != produkt.ActiveStatus || prod.Status != produkt.ActiveStatus {
			log.Warn().Int("variante_id", input.VarianteID).Int("produkt_id", input.ProduktID).Msg("Variant or product not active")
			return nil, ErrVarianteNichtAktiv
		}

		positionen = append(positionen, kasse.Position{
			VarianteID:       input.VarianteID,
			ProduktName:      prod.Name,
			VarianteName:     variante.Name,
			Kategorie:        string(prod.Kategorie),
			Steuersatz:       string(prod.Steuersatz),
			EinzelpreisCents: variante.PreisCents,
			Menge:            input.Menge,
		})
	}

	return positionen, nil
}
