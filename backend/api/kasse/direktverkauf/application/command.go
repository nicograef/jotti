package application

import (
	"context"
	"errors"

	bondruckApp "github.com/nicograef/jotti/backend/api/druck/bondruck/application"
	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
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
	EventExistsByTypeAndVorgangsID(ctx context.Context, eventType, vorgangsID, jsonKey string) (bool, error)
}

type kassensitzungenRepo interface {
	GetAktiveKassensitzung(ctx context.Context) (*kasse.Kassensitzung, error)
}

type produktRepo interface {
	GetVariantenByIDs(ctx context.Context, ids []int) (map[int]produkt.VarianteMitProdukt, error)
	GetProdukteByIDs(ctx context.Context, ids []int) (map[int]produkt.Produkt, error)
}

type druckstationRepo interface {
	GetKonfigurierteDruckstationen(ctx context.Context) (map[string]druckstation.Druckstation, error)
}

type Command struct {
	EventRepo           eventRepo
	ProduktRepo         produktRepo
	KassensitzungenRepo kassensitzungenRepo
	DruckstationRepo    druckstationRepo
}

// getOffeneKassensitzungOderFehler rejects the Direktverkauf before any TSE roundtrip when no Kassensitzung is open or the barrier is active.
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

// DirektverkaufTaetigen schreibt ein einziges unveränderliches Event in den eigenen Stream des
// Verkaufs und aktualisiert keine Projektion. verkaufID ist ein client-seitig erzeugter
// Idempotenz-Schlüssel (UUID): Bei OCC-Konflikt entscheidet die Suche nach der verkaufId — Treffer
// = idempotente Erfolgsantwort, kein Treffer = echter Konflikt (409). Gleiche ID bedeutet denselben
// Vorgang, der Payload wird nicht verglichen.
func (c Command) DirektverkaufTaetigen(ctx context.Context, userID int, userName string, verkaufID string, inputs []enrichment.PositionInput, kommentar string) error {
	log := zerolog.Ctx(ctx)

	ks, err := c.getOffeneKassensitzungOderFehler(ctx)
	if err != nil {
		return err
	}

	positionen, err := enrichment.EnrichPositionen(ctx, c.ProduktRepo, inputs)
	if err != nil {
		return err
	}

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
		return bondruckApp.CreateArbeitsbonAuftraegeFromEvent(stored, druckstationen, "")
	}

	// Frischer Stream (client-UUID): erwartete Version 0, das Event ist immer version = 1.
	if err := c.persistVerkaufEvent(ctx, evt, subject, 0, ks.ZNr, buildAuftraege); err != nil {
		if errors.Is(err, ErrConflict) {
			exists, lookupErr := c.EventRepo.EventExistsByTypeAndVorgangsID(ctx, string(kasse.EventTypeDirektverkaufGetaetigtV1), verkaufID, "verkaufId")
			if lookupErr != nil {
				log.Error().Err(lookupErr).Str("verkauf_id", verkaufID).Msg("Failed to lookup direktverkauf idempotency")
				return ErrDatabase
			}
			if exists {
				log.Info().Str("verkauf_id", verkaufID).Msg("Idempotenter Direktverkauf: verkaufId bereits vorhanden")
				return nil
			}
			return ErrConflict
		}
		log.Error().Err(err).Msg("Failed to write direktverkauf getaetigt event")
		return ErrDatabase
	}

	log.Info().Str("verkauf_id", verkaufID).Msg("Direktverkauf getaetigt")
	return nil
}

// konfigurierteDruckstationen returns nil when no DruckstationRepo is wired (tests); without stations no print jobs are derived.
func (c Command) konfigurierteDruckstationen(ctx context.Context) (map[string]druckstation.Druckstation, error) {
	if c.DruckstationRepo == nil {
		return nil, nil
	}
	return c.DruckstationRepo.GetKonfigurierteDruckstationen(ctx)
}

// DirektverkaufStornieren appends an immutable cancellation event to the verkauf's own stream. The
// returned cash reduces the Soll-Kassenbestand directly — no separate Auszahlung, because a
// Direktverkauf has no open Saldo.
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

	if !kasse.ValidatePositionRefs(nichtStorniert, positionen) {
		log.Warn().Str("verkauf_id", verkaufID).Msg("Storno-Invariante verletzt: angeforderte Positionen nicht stornierbar")
		return ErrPositionNichtStornierbar
	}

	resolvedPositionen, gesamtStornierungCents := kasse.ResolvePositionen(nichtStorniert, positionen)

	evt, err := kasse.NewDirektverkaufStorniertEvent(subject, verkaufID, userID, userName, resolvedPositionen, gesamtStornierungCents, kommentar)
	if err != nil {
		log.Error().Err(err).Str("verkauf_id", verkaufID).Msg("Failed to create direktverkauf storniert event")
		return err
	}

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

// persistVerkaufEvent: mit nicht-nil buildAuftraege entstehen die Druckaufträge in derselben
// Transaktion; der Signaturauftrag des Events entsteht in jedem Fall im selben Commit.
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
