// Package enrichment turns thin position inputs into fat kasse.Position values for the Kasse commands
// and builds the Pfandrückgabe settled with a Zahlung or a Direktverkauf.
package enrichment

import (
	"context"
	"errors"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/internal/logging"
)

type PositionInput struct {
	ProduktID  int
	VarianteID int
	Menge      int
}

// PfandEingabe is the Pfand return settled within a payment or a Direktverkauf; RueckgabeID is the
// client's idempotency key of the Rückgabe.
type PfandEingabe struct {
	RueckgabeID string
	Positionen  []PositionInput
}

// MengenJeVariante sums the requested quantity per variant, the form a retry compares against the
// stored Vorgang.
func MengenJeVariante(inputs []PositionInput) map[int]int {
	mengen := make(map[int]int, len(inputs))
	for _, input := range inputs {
		mengen[input.VarianteID] += input.Menge
	}
	return mengen
}

type produktRepo interface {
	GetVariantenByIDs(ctx context.Context, ids []int) (map[int]produkt.VarianteMitProdukt, error)
	GetProdukteByIDs(ctx context.Context, ids []int) (map[int]produkt.Produkt, error)
}

var ErrProduktNotFound = errors.New("produkt not found")

// ErrVarianteNichtAktiv covers a deactivated variant or product; ErrProduktNotFound covers deleted or unknown IDs.
var ErrVarianteNichtAktiv = errors.New("variante nicht aktiv")

// ErrKeinPfandProdukt: a Pfand position references a product without the Pfand flag.
var ErrKeinPfandProdukt = errors.New("kein pfand produkt")

// EnrichPfandPositionen enriches like EnrichPositionen and accepts only variants of Pfand products.
func EnrichPfandPositionen(ctx context.Context, repo produktRepo, inputs []PositionInput) ([]kasse.Position, error) {
	return enrich(ctx, repo, inputs, true)
}

// VerrechneteRueckgabe builds the Rückgabe settled with partner, the Zahlung or Verkauf it is written
// with, as version 1 of its own stream; tischID 0 is the Direktverkauf. It takes the partner's user
// and timestamp, the pairing kasse.VerrechnungsPartner finds it by, and no Kommentar: that belongs
// to the partner.
func VerrechneteRueckgabe(ctx context.Context, repo produktRepo, kassensitzungNr int, pfand PfandEingabe, tischID int, partner event.Event) (event.Event, error) {
	positionen, err := EnrichPfandPositionen(ctx, repo, pfand.Positionen)
	if err != nil {
		return event.Event{}, err
	}
	rueckgabe, err := kasse.NewPfandZurueckgegebenEvent(kasse.PfandrueckgabeSubject(kassensitzungNr, pfand.RueckgabeID), pfand.RueckgabeID, tischID, partner.UserID, partner.UserName, positionen, "")
	if err != nil {
		logging.FromContext(ctx).Error("Failed to create pfand zurueckgegeben event", "error", err, "tisch_id", tischID)
		return event.Event{}, err
	}
	rueckgabe = kasse.Verrechne(rueckgabe, partner)
	rueckgabe.Version = 1
	return rueckgabe, nil
}

func EnrichPositionen(ctx context.Context, repo produktRepo, inputs []PositionInput) ([]kasse.Position, error) {
	return enrich(ctx, repo, inputs, false)
}

func enrich(ctx context.Context, repo produktRepo, inputs []PositionInput, nurPfand bool) ([]kasse.Position, error) {
	log := logging.FromContext(ctx)

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
		log.Error("Failed to batch-fetch variants for position enrichment", "error", err)
		return nil, db.ErrDatabase
	}
	produkteByID, err := repo.GetProdukteByIDs(ctx, produktIDs)
	if err != nil {
		log.Error("Failed to batch-fetch products for position enrichment", "error", err)
		return nil, db.ErrDatabase
	}

	positionen := make([]kasse.Position, 0, len(inputs))
	for _, input := range inputs {
		position, err := positionAus(ctx, input, variantenByID, produkteByID, nurPfand)
		if err != nil {
			return nil, err
		}
		positionen = append(positionen, position)
	}

	return positionen, nil
}

// positionAus validates one input against the batch-loaded variants and products and builds its position.
func positionAus(ctx context.Context, input PositionInput, variantenByID map[int]produkt.VarianteMitProdukt, produkteByID map[int]produkt.Produkt, nurPfand bool) (kasse.Position, error) {
	log := logging.FromContext(ctx)

	varianteMitProdukt, ok := variantenByID[input.VarianteID]
	if !ok {
		log.Error("Variant not found in batch result", "variante_id", input.VarianteID)
		return kasse.Position{}, ErrProduktNotFound
	}
	prod, ok := produkteByID[input.ProduktID]
	if !ok {
		log.Error("Product not found in batch result", "produkt_id", input.ProduktID)
		return kasse.Position{}, ErrProduktNotFound
	}

	// Die Paarung Produkt/Variante kommt vom Client: ohne diese Prüfung erbte die Position
	// Kategorie und Steuersatz eines fremden Produkts, den Preis aber von der Variante.
	if varianteMitProdukt.ProduktID != input.ProduktID {
		log.Error("Variant does not belong to the referenced product", "variante_id", input.VarianteID, "produkt_id", input.ProduktID)
		return kasse.Position{}, ErrProduktNotFound
	}
	variante := varianteMitProdukt.Variante

	// Defense-in-Depth: deaktivierte (inactive) Varianten/Produkte tauchen im
	// Menü nicht auf, könnten aber per direktem POST referenziert werden.
	if variante.Status != produkt.ActiveStatus || prod.Status != produkt.ActiveStatus {
		log.Warn("Variant or product not active", "variante_id", input.VarianteID, "produkt_id", input.ProduktID)
		return kasse.Position{}, ErrVarianteNichtAktiv
	}

	if nurPfand && !prod.Pfand {
		log.Warn("Product is no Pfand product", "produkt_id", input.ProduktID)
		return kasse.Position{}, ErrKeinPfandProdukt
	}

	return kasse.Position{
		VarianteID:       input.VarianteID,
		ProduktName:      prod.Name,
		VarianteName:     variante.Name,
		Kategorie:        string(prod.Kategorie),
		Steuersatz:       string(prod.Steuersatz),
		EinzelpreisCents: variante.PreisCents,
		Menge:            input.Menge,
		Pfand:            prod.Pfand,
	}, nil
}
