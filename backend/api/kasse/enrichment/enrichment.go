// Package enrichment turns thin position inputs (Produkt/Variante IDs + Menge)
// into fat kasse.Position values by batch-loading the referenced Produkte and
// Varianten. It is the single shared implementation used by both the
// tischgeschaeft (BestellungAufnehmen) and direktverkauf (DirektverkaufTaetigen)
// command paths (extracted per the 2026-07-17 review, go-code-quality-1).
package enrichment

import (
	"context"
	"errors"

	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/rs/zerolog"
)

// PositionInput is the thin input for a single Position: which Variante of which
// Produkt, and how many. EnrichPositionen enriches it with Produkt/Variante
// details (fat events).
type PositionInput struct {
	ProduktID  int
	VarianteID int
	Menge      int
}

// produktRepo is the narrow read side enrichment needs: the two batch lookups.
type produktRepo interface {
	GetVariantsByIDs(ctx context.Context, ids []int) (map[int]produkt.Variante, error)
	GetProductsByIDs(ctx context.Context, ids []int) (map[int]produkt.Produkt, error)
}

// ErrProduktNotFound is returned when a product or variant is not found during enrichment.
var ErrProduktNotFound = errors.New("produkt not found")

// ErrVarianteNichtAktiv is returned when a referenced variant or its product is
// deactivated (inactive). Kept separate from ErrProduktNotFound, which covers
// deleted or non-existent IDs.
var ErrVarianteNichtAktiv = errors.New("variante nicht aktiv")

// EnrichPositionen batch-fetches the referenced variants and products and turns the inputs
// into fat Positions carrying name, category and price for the event store.
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

	variantenByID, err := repo.GetVariantsByIDs(ctx, varianteIDs)
	if err != nil {
		log.Error().Err(err).Msg("Failed to batch-fetch variants for position enrichment")
		return nil, ErrProduktNotFound
	}
	produkteByID, err := repo.GetProductsByIDs(ctx, produktIDs)
	if err != nil {
		log.Error().Err(err).Msg("Failed to batch-fetch products for position enrichment")
		return nil, ErrProduktNotFound
	}

	positionen := make([]kasse.Position, 0, len(inputs))
	for _, input := range inputs {
		variant, ok := variantenByID[input.VarianteID]
		if !ok {
			log.Error().Int("variante_id", input.VarianteID).Msg("Variant not found in batch result")
			return nil, ErrProduktNotFound
		}
		prod, ok := produkteByID[input.ProduktID]
		if !ok {
			log.Error().Int("produkt_id", input.ProduktID).Msg("Product not found in batch result")
			return nil, ErrProduktNotFound
		}

		// Defense-in-Depth: deaktivierte (inactive) Varianten/Produkte tauchen im
		// Menü nicht auf, könnten aber per direktem POST referenziert werden.
		if variant.Status != produkt.ActiveStatus || prod.Status != produkt.ActiveStatus {
			log.Warn().Int("variante_id", input.VarianteID).Int("produkt_id", input.ProduktID).Msg("Variant or product not active")
			return nil, ErrVarianteNichtAktiv
		}

		positionen = append(positionen, kasse.Position{
			VarianteID:       input.VarianteID,
			ProduktName:      prod.Name,
			VarianteName:     variant.Name,
			Kategorie:        string(prod.Kategorie),
			Steuersatz:       string(prod.Steuersatz),
			EinzelpreisCents: variant.PreisCents,
			Menge:            input.Menge,
		})
	}

	return positionen, nil
}
