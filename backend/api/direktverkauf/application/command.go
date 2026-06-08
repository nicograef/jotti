package application

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/product"
	"github.com/rs/zerolog"
)

type eventRepo interface {
	WriteEvent(ctx context.Context, e event.Event, streamType kasse.StreamType, kassensitzungNr int) (int, error)
	GetMaxVersion(ctx context.Context, subject string) (int, error)
	ReadEventsBySubject(ctx context.Context, subject string) ([]event.Event, error)
}

type kassensitzungenRepo interface {
	GetOffeneKassensitzung(ctx context.Context) (*kasse.Kassensitzung, error)
}

type productRepo interface {
	GetVariantsByIDs(ctx context.Context, ids []int) (map[int]product.Variante, error)
	GetProductsByIDs(ctx context.Context, ids []int) (map[int]product.Produkt, error)
}

// VerkaufPositionInput represents a single position of a Direktverkauf.
// The application layer enriches it with product/variant details (fat events).
type VerkaufPositionInput struct {
	ProduktID  int
	VarianteID int
	Menge      int
}

type Command struct {
	EventRepo           eventRepo
	ProductRepo         productRepo
	KassensitzungenRepo kassensitzungenRepo
}

// DirektverkaufTaetigen records a Direktverkauf as a single immutable event in its own stream
// (kassensitzung-{nr}/direktverkauf-{uuid}). It requires an open Kassensitzung and writes nothing
// to any projection. Returns ErrKasseNichtGeoeffnet (HTTP 409) when no Kassensitzung is open.
func (c Command) DirektverkaufTaetigen(ctx context.Context, userID int, userName string, inputs []VerkaufPositionInput, kommentar string) error {
	log := zerolog.Ctx(ctx)

	ks, err := c.KassensitzungenRepo.GetOffeneKassensitzung(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load open Kassensitzung")
		return ErrDatabase
	}
	if ks == nil {
		return ErrKasseNichtGeoeffnet
	}

	positionen, err := c.enrichPositionen(ctx, inputs)
	if err != nil {
		return err
	}

	verkaufID := uuid.New().String()
	subject := kasse.DirektverkaufSubject(ks.ZNr, verkaufID)

	evt, err := kasse.NewDirektverkaufGetaetigtEvent(subject, verkaufID, userID, userName, positionen, kommentar)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create direktverkauf getaetigt event")
		return err
	}

	if err := c.writeEvent(ctx, evt, subject, ks.ZNr); err != nil {
		if errors.Is(err, ErrConflict) {
			return ErrConflict
		}
		log.Error().Err(err).Msg("Failed to write direktverkauf getaetigt event")
		return ErrDatabase
	}

	log.Info().Str("verkauf_id", verkaufID).Msg("Direktverkauf getaetigt")
	return nil
}

// DirektverkaufStornieren records a position-precise cancellation of a Direktverkauf as an immutable
// event appended to that verkauf's own stream (version = maxVersion + 1, OCC). The returned cash
// reduces the Soll-Kassenbestand directly — there is no separate Auszahlung, because a Direktverkauf
// has no open Saldo. Requires an open Kassensitzung (ErrKasseNichtGeoeffnet otherwise). Returns
// ErrVerkaufNichtGefunden when the verkauf does not exist and ErrPositionNichtStornierbar when a
// requested position is not (or no longer) cancellable.
func (c Command) DirektverkaufStornieren(ctx context.Context, userID int, userName string, verkaufID string, positionen []kasse.PositionRef, kommentar string) error {
	log := zerolog.Ctx(ctx)

	ks, err := c.KassensitzungenRepo.GetOffeneKassensitzung(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load open Kassensitzung")
		return ErrDatabase
	}
	if ks == nil {
		return ErrKasseNichtGeoeffnet
	}

	subject := kasse.DirektverkaufSubject(ks.ZNr, verkaufID)

	events, err := c.EventRepo.ReadEventsBySubject(ctx, subject)
	if err != nil {
		log.Error().Err(err).Str("verkauf_id", verkaufID).Msg("Failed to read direktverkauf events")
		return ErrDatabase
	}
	if len(events) == 0 {
		return ErrVerkaufNichtGefunden
	}

	nichtStorniert, err := kasse.ComputeNichtStornierteVerkaufPositionen(events)
	if err != nil {
		log.Error().Err(err).Str("verkauf_id", verkaufID).Msg("Failed to compute nicht-stornierte Positionen")
		return ErrDatabase
	}

	if !validatePositionRefs(nichtStorniert, positionen) {
		log.Warn().Str("verkauf_id", verkaufID).Msg("Storno-Invariante verletzt: angeforderte Positionen nicht stornierbar")
		return ErrPositionNichtStornierbar
	}

	resolvedPositionen, gesamtStornierungCents := resolvePositionen(nichtStorniert, positionen)

	evt, err := kasse.NewDirektverkaufStorniertEvent(subject, verkaufID, userID, userName, resolvedPositionen, gesamtStornierungCents, kommentar)
	if err != nil {
		log.Error().Err(err).Str("verkauf_id", verkaufID).Msg("Failed to create direktverkauf storniert event")
		return err
	}

	if err := c.writeEvent(ctx, evt, subject, ks.ZNr); err != nil {
		if errors.Is(err, ErrConflict) {
			return ErrConflict
		}
		log.Error().Err(err).Str("verkauf_id", verkaufID).Msg("Failed to write direktverkauf storniert event")
		return ErrDatabase
	}

	log.Info().Str("verkauf_id", verkaufID).Int("gesamt_stornierung_cents", gesamtStornierungCents).Msg("Direktverkauf storniert")
	return nil
}

// validatePositionRefs checks that every requested PositionRef exists in the available positions
// and that the requested Menge does not exceed the available Menge.
func validatePositionRefs(available []kasse.Position, requested []kasse.PositionRef) bool {
	for _, ref := range requested {
		found := false
		for _, pos := range available {
			if pos.PositionID == ref.PositionID {
				if ref.Menge > pos.Menge {
					return false
				}
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// resolvePositionen resolves the requested PositionRefs to fat Positions using the available
// (not-yet-cancelled) positions, returning the resolved positions and their total in cents.
// Mirrors the Tisch-Storno so the storno event is self-contained (fat).
func resolvePositionen(available []kasse.Position, requested []kasse.PositionRef) ([]kasse.Position, int) {
	resolved := make([]kasse.Position, 0, len(requested))
	totalCents := 0
	for _, ref := range requested {
		for _, pos := range available {
			if pos.PositionID == ref.PositionID {
				resolved = append(resolved, kasse.Position{
					PositionID:   pos.PositionID,
					VarianteID:   pos.VarianteID,
					ProduktName:  pos.ProduktName,
					VarianteName: pos.VarianteName,
					Kategorie:    pos.Kategorie,
					Einzelpreis:  pos.Einzelpreis,
					Menge:        ref.Menge,
				})
				totalCents += pos.Einzelpreis * ref.Menge
				break
			}
		}
	}
	return resolved, totalCents
}

// enrichPositionen batch-fetches the referenced variants and products and turns the inputs
// into fat Positions carrying name, category and price for the event store.
func (c Command) enrichPositionen(ctx context.Context, inputs []VerkaufPositionInput) ([]kasse.Position, error) {
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

	variantenByID, err := c.ProductRepo.GetVariantsByIDs(ctx, varianteIDs)
	if err != nil {
		log.Error().Err(err).Msg("Failed to batch-fetch variants for position enrichment")
		return nil, ErrProduktNotFound
	}
	produkteByID, err := c.ProductRepo.GetProductsByIDs(ctx, produktIDs)
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

		positionen = append(positionen, kasse.Position{
			VarianteID:   input.VarianteID,
			ProduktName:  prod.Name,
			VarianteName: variant.Name,
			Kategorie:    string(prod.Kategorie),
			Einzelpreis:  variant.PreisCents,
			Menge:        input.Menge,
		})
	}

	return positionen, nil
}

// writeEvent writes an event with optimistic concurrency control.
// It reads the current max version for the subject, sets event.Version = maxVersion + 1
// (1 for a fresh Direktverkauf stream), and writes the event into the kassenjournal.
// Returns ErrConflict on a UNIQUE(subject, version) violation.
func (c Command) writeEvent(ctx context.Context, e event.Event, subject string, kassensitzungNr int) error {
	maxVersion, err := c.EventRepo.GetMaxVersion(ctx, subject)
	if err != nil {
		return err
	}

	e.Version = maxVersion + 1

	if _, err := c.EventRepo.WriteEvent(ctx, e, kasse.StreamTypeDirektverkauf, kassensitzungNr); err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			zerolog.Ctx(ctx).Warn().Int("version", e.Version).Str("subject", subject).Msg("OCC conflict")
			return ErrConflict
		}
		return err
	}

	return nil
}
