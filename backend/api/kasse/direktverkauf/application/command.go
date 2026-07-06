package application

import (
	"context"
	"errors"

	"github.com/google/uuid"
	bondruckApp "github.com/nicograef/jotti/backend/api/druck/bondruck/application"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/druckstation"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/rs/zerolog"
)

type eventRepo interface {
	WriteEvent(ctx context.Context, e event.Event, streamType kasse.StreamType, kassensitzungNr int) (int, error)
	WriteEventWithDruckauftraege(ctx context.Context, e event.Event, streamType kasse.StreamType, kassensitzungNr int, buildAuftraege func(event.Event) []druckauftrag_repo.NeuerDruckauftrag) (int, error)
	GetMaxVersion(ctx context.Context, subject string) (int, error)
	ReadEventsBySubject(ctx context.Context, subject string) ([]event.Event, error)
}

type kassensitzungenRepo interface {
	GetOffeneKassensitzung(ctx context.Context) (*kasse.Kassensitzung, error)
	GetAktiveKassensitzung(ctx context.Context) (*kasse.Kassensitzung, error)
}

type produktRepo interface {
	GetVariantsByIDs(ctx context.Context, ids []int) (map[int]produkt.Variante, error)
	GetProductsByIDs(ctx context.Context, ids []int) (map[int]produkt.Produkt, error)
}

type druckstationRepo interface {
	GetKonfigurierteDruckstationen(ctx context.Context) (map[string]druckstation.Druckstation, error)
}

// VerkaufPositionInput is the input for a single Position of a Direktverkauf.
// The application layer enriches it with Produkt/Variante details (fat events).
type VerkaufPositionInput struct {
	ProduktID  int
	VarianteID int
	Menge      int
}

type Command struct {
	EventRepo           eventRepo
	ProduktRepo         produktRepo
	KassensitzungenRepo kassensitzungenRepo
	DruckstationRepo    druckstationRepo
}

// getOffeneKassensitzungOderFehler retrieves the open Kassensitzung for a Direktverkauf. It returns
// ErrKasseNichtGeoeffnet when none is active and ErrKasseWirdAbgeschlossen while the Kassensitzung is
// being closed (barrier active), rejecting the Direktverkauf before any TSE roundtrip.
func (c Command) getOffeneKassensitzungOderFehler(ctx context.Context) (*kasse.Kassensitzung, error) {
	ks, err := c.KassensitzungenRepo.GetAktiveKassensitzung(ctx)
	if err != nil {
		return nil, ErrDatabase
	}
	if ks == nil {
		return nil, ErrKasseNichtGeoeffnet
	}
	if ks.Status == kasse.KassensitzungWirdAbgeschlossen {
		return nil, ErrKasseWirdAbgeschlossen
	}
	return ks, nil
}

// DirektverkaufTaetigen records a Direktverkauf as a single immutable event in its own stream
// (kassensitzung-{nr}/direktverkauf-{uuid}). It requires an open Kassensitzung and writes nothing
// to any projection. Returns ErrKasseNichtGeoeffnet (HTTP 409) when no Kassensitzung is open.
func (c Command) DirektverkaufTaetigen(ctx context.Context, userID int, userName string, inputs []VerkaufPositionInput, kommentar string) error {
	log := zerolog.Ctx(ctx)

	ks, err := c.getOffeneKassensitzungOderFehler(ctx)
	if err != nil {
		return err
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

	druckstationen, err := c.konfigurierteDruckstationen(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load druckstationen for direktverkauf")
		return ErrDatabase
	}

	buildAuftraege := func(stored event.Event) []druckauftrag_repo.NeuerDruckauftrag {
		return bondruckApp.CreateArbeitsbonAuftraegeFromEvent(stored, druckstationen)
	}

	// Frischer Stream (neue UUID): erwartete Version 0, das Event ist immer version = 1.
	if err := c.persistVerkaufEvent(ctx, evt, subject, 0, ks.ZNr, buildAuftraege); err != nil {
		if errors.Is(err, ErrConflict) {
			return ErrConflict
		}
		log.Error().Err(err).Msg("Failed to write direktverkauf getaetigt event")
		return ErrDatabase
	}

	log.Info().Str("verkauf_id", verkaufID).Msg("Direktverkauf getaetigt")
	return nil
}

// konfigurierteDruckstationen returns the configured Druckstationen, or an empty
// map when no DruckstationRepo is wired (e.g. in tests). Without configured stations
// the policy derives no print jobs.
func (c Command) konfigurierteDruckstationen(ctx context.Context) (map[string]druckstation.Druckstation, error) {
	if c.DruckstationRepo == nil {
		return nil, nil
	}
	return c.DruckstationRepo.GetKonfigurierteDruckstationen(ctx)
}

// DirektverkaufStornieren records a position-precise cancellation of a Direktverkauf as an immutable
// event appended to that verkauf's own stream (version = maxVersion + 1, OCC). The returned cash
// reduces the Soll-Kassenbestand directly — there is no separate Auszahlung, because a Direktverkauf
// has no open Saldo. Requires an open Kassensitzung (ErrKasseNichtGeoeffnet otherwise). Returns
// ErrVerkaufNichtGefunden when the verkauf does not exist and ErrPositionNichtStornierbar when a
// requested position is not (or no longer) cancellable.
func (c Command) DirektverkaufStornieren(ctx context.Context, userID int, userName string, verkaufID string, positionen []kasse.PositionRef, kommentar string) error {
	log := zerolog.Ctx(ctx)

	ks, err := c.getOffeneKassensitzungOderFehler(ctx)
	if err != nil {
		return err
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

	// OCC gegen den validierten Zustand: Basis ist die höchste Version des Replays,
	// gegen den die Storno-Invariante geprüft wurde.
	if err := c.persistVerkaufEvent(ctx, evt, subject, events[len(events)-1].Version, ks.ZNr, nil); err != nil {
		if errors.Is(err, ErrConflict) {
			return ErrConflict
		}
		log.Error().Err(err).Str("verkauf_id", verkaufID).Msg("Failed to write direktverkauf storniert event")
		return ErrDatabase
	}

	log.Info().Str("verkauf_id", verkaufID).Int("gesamt_stornierung_cents", gesamtStornierungCents).Msg("Direktverkauf storniert")
	return nil
}

// validatePositionRefs checks that every requested PositionRef exists in the available positions,
// that no PositionID is referenced more than once (duplicates would add up unnoticed), and that
// the requested Menge does not exceed the available Menge.
func validatePositionRefs(available []kasse.Position, requested []kasse.PositionRef) bool {
	seen := make(map[string]bool, len(requested))
	for _, ref := range requested {
		if seen[ref.PositionID] {
			return false
		}
		seen[ref.PositionID] = true
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
					PositionID:       pos.PositionID,
					VarianteID:       pos.VarianteID,
					ProduktName:      pos.ProduktName,
					VarianteName:     pos.VarianteName,
					Kategorie:        pos.Kategorie,
					Steuersatz:       pos.Steuersatz,
					EinzelpreisCents: pos.EinzelpreisCents,
					Menge:            ref.Menge,
				})
				totalCents += pos.EinzelpreisCents * ref.Menge
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

	variantenByID, err := c.ProduktRepo.GetVariantsByIDs(ctx, varianteIDs)
	if err != nil {
		log.Error().Err(err).Msg("Failed to batch-fetch variants for position enrichment")
		return nil, ErrProduktNotFound
	}
	produkteByID, err := c.ProduktRepo.GetProductsByIDs(ctx, produktIDs)
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

// writeVersionedEvent writes the event with version expectedVersion+1 via write.
// expectedVersion ist die Version des Zustands, gegen den der Command validiert hat
// (1. Event eines frischen Streams: 0; Storno: höchste Version des Replays). Ein
// UNIQUE(subject, version)-Konflikt — der Stream hat sich seit dem Lesen geändert —
// wird zu ErrConflict.
func writeVersionedEvent(ctx context.Context, e event.Event, subject string, expectedVersion int, write func(event.Event) (int, error)) error {
	e.Version = expectedVersion + 1

	if _, err := write(e); err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			zerolog.Ctx(ctx).Warn().Int("version", e.Version).Str("subject", subject).Msg("OCC conflict")
			return ErrConflict
		}
		if errors.Is(err, db.ErrConflict) {
			zerolog.Ctx(ctx).Warn().Str("subject", subject).Msg("Deadlock on event write")
			return ErrConflict
		}
		if errors.Is(err, kassenjournal_repo.ErrKassensitzungNichtOffen) {
			zerolog.Ctx(ctx).Warn().Str("subject", subject).Msg("Kassensitzung nicht mehr offen")
			return ErrKasseNichtGeoeffnet
		}
		return err
	}

	return nil
}

// persistVerkaufEvent writes a Direktverkauf event with OCC against expectedVersion.
// When buildAuftraege is non-nil the derived print jobs are written in the same
// transaction; der Signaturauftrag des Events entsteht in jedem Fall im selben
// Commit (fiskalische Projektion). Returns ErrConflict on a version conflict.
func (c Command) persistVerkaufEvent(ctx context.Context, evt event.Event, subject string, expectedVersion int, kassensitzungNr int, buildAuftraege func(event.Event) []druckauftrag_repo.NeuerDruckauftrag) error {
	if buildAuftraege != nil {
		return writeVersionedEvent(ctx, evt, subject, expectedVersion, func(versioned event.Event) (int, error) {
			return c.EventRepo.WriteEventWithDruckauftraege(ctx, versioned, kasse.StreamTypeDirektverkauf, kassensitzungNr, buildAuftraege)
		})
	}
	return writeVersionedEvent(ctx, evt, subject, expectedVersion, func(versioned event.Event) (int, error) {
		return c.EventRepo.WriteEvent(ctx, versioned, kasse.StreamTypeDirektverkauf, kassensitzungNr)
	})
}
