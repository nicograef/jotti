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
	"github.com/nicograef/jotti/backend/domain/tisch"
	"github.com/nicograef/jotti/backend/internal/logging"
	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
)

type eventRepo interface {
	WriteEvent(ctx context.Context, e event.Event, streamType kasse.StreamType, kassensitzungNr int) (int, error)
	WriteEventsAtomic(ctx context.Context, kassensitzungNr int, events []kassenjournal_repo.NeuesEvent) error
	ReadEventsBySubject(ctx context.Context, subject string) ([]event.Event, error)
	ReadDirektverkaufEvents(ctx context.Context, kassensitzungNr int) ([]event.Event, error)
	ReadPfandrueckgabeEvents(ctx context.Context, kassensitzungNr int, tischID int) ([]event.Event, error)
	GetDirektverkauf(ctx context.Context, verkaufID string) (event.Event, bool, error)
	GetPfandrueckgabe(ctx context.Context, rueckgabeID string) (event.Event, bool, error)
	GetPfandrueckgabeStornierung(ctx context.Context, stornierungID string) (event.Event, bool, error)
}

type kassensitzungenRepo interface {
	GetAktiveKassensitzung(ctx context.Context) (*kasse.Kassensitzung, error)
}

type produktRepo interface {
	GetVariantenByIDs(ctx context.Context, ids []int) (map[int]produkt.VarianteMitProdukt, error)
	GetProdukteByIDs(ctx context.Context, ids []int) (map[int]produkt.Produkt, error)
}

type tischRepo interface {
	GetTisch(ctx context.Context, id int) (tisch.Tisch, error)
}

type druckstationRepo interface {
	GetKonfigurierteDruckstationen(ctx context.Context) (map[string]druckstation.Druckstation, error)
}

type Command struct {
	EventRepo           eventRepo
	ProduktRepo         produktRepo
	KassensitzungenRepo kassensitzungenRepo
	DruckstationRepo    druckstationRepo
	TischRepo           tischRepo
}

// getOffeneKassensitzungOderFehler rejects the Direktverkauf or Pfandrückgabe before any TSE roundtrip when no Kassensitzung is open or the barrier is active.
func (c Command) getOffeneKassensitzungOderFehler(ctx context.Context) (*kasse.Kassensitzung, error) {
	ks, err := c.KassensitzungenRepo.GetAktiveKassensitzung(ctx)
	if err != nil {
		logging.FromContext(ctx).Error("Failed to load aktive Kassensitzung", "error", err)
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

// DirektverkaufTaetigen writes the sale into its own stream, without projection; with pfand the settled
// Rückgabe (tischId 0) goes into its own stream in the same transaction. verkaufID is the client's
// idempotency key: on an OCC conflict a stored Verkauf with that verkaufId means idempotent success,
// none means a real 409. Without pfand the stored Verkauf must be unsettled (payload not compared),
// with pfand the stored pair must match the request, see pruefeVerrechnetenVerkauf.
func (c Command) DirektverkaufTaetigen(ctx context.Context, userID int, userName string, verkaufID string, inputs []enrichment.PositionInput, kommentar string, pfand *enrichment.PfandEingabe) error {
	log := logging.FromContext(ctx)

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
		log.Error("Failed to create direktverkauf getaetigt event", "error", err)
		return err
	}
	// Frischer Stream (client-UUID): das Event ist immer version = 1.
	evt.Version = 1

	druckstationen, err := c.konfigurierteDruckstationen(ctx)
	if err != nil {
		log.Error("Failed to load druckstationen for direktverkauf", "error", err)
		return ErrDatabase
	}

	buildAuftraege := func(stored event.Event) []druckauftrag_repo.NeuerDruckauftrag {
		return bondruckApp.CreateArbeitsbonAuftraegeFromEvent(stored, druckstationen, "")
	}

	events := []kassenjournal_repo.NeuesEvent{{Event: evt, StreamType: kasse.StreamTypeDirektverkauf, BuildAuftraege: buildAuftraege}}

	if pfand != nil {
		pfandPositionen, err := enrichment.EnrichPfandPositionen(ctx, c.ProduktRepo, pfand.Positionen)
		if err != nil {
			return err
		}
		// Der Kommentar gehört zum Verkauf; die verrechnete Rückgabe trägt keinen.
		rueckgabe, err := kasse.NewPfandZurueckgegebenEvent(kasse.PfandrueckgabeSubject(ks.ZNr, pfand.RueckgabeID), pfand.RueckgabeID, 0, userID, userName, pfandPositionen, "")
		if err != nil {
			log.Error("Failed to create pfand zurueckgegeben event", "error", err)
			return err
		}
		rueckgabe = kasse.Verrechne(rueckgabe, evt)
		rueckgabe.Version = 1
		events = append(events, kassenjournal_repo.NeuesEvent{Event: rueckgabe, StreamType: kasse.StreamTypePfandrueckgabe})
	}

	if err := c.EventRepo.WriteEventsAtomic(ctx, ks.ZNr, events); err != nil {
		if errors.Is(err, db.ErrAlreadyExists) || errors.Is(err, db.ErrConflict) {
			log.Warn("OCC conflict bei Direktverkauf", "subject", subject)
			if pfand != nil {
				return c.pruefeVerrechnetenVerkauf(ctx, verkaufID, inputs, pfand)
			}
			return c.pruefeUnverrechnetenVerkauf(ctx, verkaufID)
		}
		if errors.Is(err, kassenjournal_repo.ErrKassensitzungNichtOffen) {
			log.Warn("Kassensitzung nicht mehr offen", "verkauf_id", verkaufID)
			return ErrKasseNichtGeoeffnet
		}
		log.Error("Failed to write direktverkauf getaetigt event", "error", err)
		return ErrDatabase
	}

	log.Info("Direktverkauf getaetigt", "verkauf_id", verkaufID, "mit_pfand", pfand != nil)
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
	log := logging.FromContext(ctx)

	ks, err := c.getOffeneKassensitzungOderFehler(ctx)
	if err != nil {
		return err
	}

	subject := kasse.DirektverkaufSubject(ks.ZNr, verkaufID)

	events, err := c.EventRepo.ReadEventsBySubject(ctx, subject)
	if err != nil {
		log.Error("Failed to read direktverkauf events", "error", err, "verkauf_id", verkaufID)
		return ErrDatabase
	}
	if len(events) == 0 {
		return ErrVerkaufNichtGefunden
	}

	eintrag, err := kasse.BuildDirektverkaufHistorieEintrag(events)
	if err != nil {
		log.Error("Failed to replay direktverkauf", "error", err, "verkauf_id", verkaufID)
		return ErrDatabase
	}
	nichtStorniert := eintrag.OffenePositionen

	resolvedPositionen, gesamtStornierungCents, ok := kasse.ResolvePositionen(nichtStorniert, positionen)
	if !ok {
		log.Warn("Storno-Invariante verletzt: angeforderte Positionen nicht stornierbar", "verkauf_id", verkaufID)
		return ErrPositionNichtStornierbar
	}

	evt, err := kasse.NewDirektverkaufStorniertEvent(subject, verkaufID, userID, userName, resolvedPositionen, gesamtStornierungCents, kommentar)
	if err != nil {
		log.Error("Failed to create direktverkauf storniert event", "error", err, "verkauf_id", verkaufID)
		return err
	}

	if err := c.writeEventOCC(ctx, evt, kasse.StreamTypeDirektverkauf, subject, events[len(events)-1].Version, ks.ZNr); err != nil {
		return err
	}

	log.Info("Direktverkauf storniert", "verkauf_id", verkaufID, "gesamt_stornierung_cents", gesamtStornierungCents)
	return nil
}

// writeEventOCC writes at expectedVersion+1: 0 for a fresh stream, the replay's highest version for
// a Storno. A UNIQUE(subject, version) conflict or a deadlock means the stream changed since reading
// and becomes ErrConflict.
func (c Command) writeEventOCC(ctx context.Context, e event.Event, streamType kasse.StreamType, subject string, expectedVersion int, kassensitzungNr int) error {
	log := logging.FromContext(ctx)
	e.Version = expectedVersion + 1

	_, err := c.EventRepo.WriteEvent(ctx, e, streamType, kassensitzungNr)
	switch {
	case err == nil:
		return nil
	case errors.Is(err, db.ErrAlreadyExists):
		log.Warn("OCC conflict", "version", e.Version, "subject", subject)
		return ErrConflict
	case errors.Is(err, db.ErrConflict):
		log.Warn("Deadlock on event write", "subject", subject)
		return ErrConflict
	case errors.Is(err, kassenjournal_repo.ErrKassensitzungNichtOffen):
		log.Warn("Kassensitzung nicht mehr offen", "subject", subject)
		return ErrKasseNichtGeoeffnet
	default:
		log.Error("Failed to write event", "error", err, "subject", subject, "type", e.Type)
		return ErrDatabase
	}
}
