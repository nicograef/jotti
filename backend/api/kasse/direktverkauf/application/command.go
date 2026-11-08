package application

import (
	"context"
	"errors"

	bondruckApp "github.com/nicograef/jotti/backend/api/druck/bondruck/application"
	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
	"github.com/nicograef/jotti/backend/domain/druckstation"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/internal/logging"
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
	GetZusatzdruckerJeVariante(ctx context.Context, varianteIDs []int) (map[int]druckstation.Zusatzdrucker, error)
}

type Command struct {
	EventRepo           eventRepo
	ProduktRepo         produktRepo
	KassensitzungenRepo kassensitzungenRepo
	DruckstationRepo    druckstationRepo
}

// DirektverkaufTaetigen writes the sale into its own stream, without projection; with pfand the settled
// Rückgabe (tischId 0) goes into its own stream in the same transaction. verkaufID is the client's
// idempotency key: on an OCC conflict a stored Verkauf with that verkaufId means idempotent success,
// none means a real 409. Without pfand the stored Verkauf must be unsettled (payload not compared),
// with pfand the stored pair must match the request, see pruefeVerrechnetenVerkauf.
func (c Command) DirektverkaufTaetigen(ctx context.Context, userID int, userName string, verkaufID string, inputs []enrichment.PositionInput, kommentar string, pfand *enrichment.PfandEingabe) error {
	log := logging.FromContext(ctx)

	ks, err := sentinels.OffeneKassensitzung(ctx, c.KassensitzungenRepo)
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

	druckstationen, err := c.DruckstationRepo.GetKonfigurierteDruckstationen(ctx)
	if err != nil {
		log.Error("Failed to load druckstationen for direktverkauf", "error", err)
		return ErrDatabase
	}

	zusatzdrucker, err := c.DruckstationRepo.GetZusatzdruckerJeVariante(ctx, kasse.VarianteIDs(positionen))
	if err != nil {
		log.Error("Failed to load zusatzdrucker for direktverkauf", "error", err)
		return ErrDatabase
	}

	buildAuftraege := func(stored event.Event) []druckstation.NeuerDruckauftrag {
		return bondruckApp.CreateArbeitsbonAuftraegeFromEvent(stored, druckstationen, zusatzdrucker, "")
	}

	events := []kassenjournal_repo.NeuesEvent{{Event: evt, StreamType: kasse.StreamTypeDirektverkauf, BuildAuftraege: buildAuftraege}}

	if pfand != nil {
		rueckgabe, err := enrichment.VerrechneteRueckgabe(ctx, c.ProduktRepo, ks.ZNr, *pfand, 0, evt)
		if err != nil {
			return err
		}
		events = append(events, kassenjournal_repo.NeuesEvent{Event: rueckgabe, StreamType: kasse.StreamTypePfandrueckgabe})
	}

	if err := c.EventRepo.WriteEventsAtomic(ctx, ks.ZNr, events); err != nil {
		err = sentinels.Schreibfehler(ctx, err, "subject", subject)
		if errors.Is(err, ErrConflict) {
			if pfand != nil {
				return c.pruefeVerrechnetenVerkauf(ctx, verkaufID, inputs, pfand)
			}
			return c.pruefeUnverrechnetenVerkauf(ctx, verkaufID)
		}
		return err
	}

	log.Info("Direktverkauf getaetigt", "verkauf_id", verkaufID, "mit_pfand", pfand != nil)
	return nil
}

// DirektverkaufStornieren appends an immutable cancellation event to the verkauf's own stream. The
// returned cash reduces the Soll-Kassenbestand directly — no separate Auszahlung, because a
// Direktverkauf has no open Saldo. stornierungID is the client's idempotency key, see
// pruefeGebuchteStornierung.
func (c Command) DirektverkaufStornieren(ctx context.Context, userID int, userName string, stornierungID string, verkaufID string, positionen []kasse.PositionRef, kommentar string) error {
	log := logging.FromContext(ctx)

	ks, err := sentinels.OffeneKassensitzung(ctx, c.KassensitzungenRepo)
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
	if gebucht, err := pruefeGebuchteStornierung(ctx, events, stornierungID, positionen); gebucht || err != nil {
		return err
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

	evt, err := kasse.NewDirektverkaufStorniertEvent(subject, stornierungID, verkaufID, userID, userName, resolvedPositionen, gesamtStornierungCents, kommentar)
	if err != nil {
		log.Error("Failed to create direktverkauf storniert event", "error", err, "verkauf_id", verkaufID)
		return err
	}

	evt.Version = events[len(events)-1].Version + 1
	if _, err := c.EventRepo.WriteEvent(ctx, evt, kasse.StreamTypeDirektverkauf, ks.ZNr); err != nil {
		err = sentinels.Schreibfehler(ctx, err, "subject", subject, "version", evt.Version, "type", evt.Type)
		if errors.Is(err, ErrConflict) {
			events, readErr := c.EventRepo.ReadEventsBySubject(ctx, subject)
			if readErr != nil {
				log.Error("Failed to read direktverkauf events after conflict", "error", readErr, "verkauf_id", verkaufID)
				return ErrDatabase
			}
			if gebucht, err := pruefeGebuchteStornierung(ctx, events, stornierungID, positionen); gebucht || err != nil {
				return err
			}
		}
		return err
	}

	log.Info("Direktverkauf storniert", "verkauf_id", verkaufID, "gesamt_stornierung_cents", gesamtStornierungCents)
	return nil
}

// pruefeGebuchteStornierung reports whether stornierungID is booked in the Verkauf stream. A booked
// one answers success only if it cancels the same quantities.
func pruefeGebuchteStornierung(ctx context.Context, events []event.Event, stornierungID string, positionen []kasse.PositionRef) (bool, error) {
	log := logging.FromContext(ctx)

	gespeichert, found, err := kasse.GebuchteDirektverkaufStornierung(events, stornierungID)
	if err != nil {
		log.Error("Failed to search booked direktverkauf stornierung", "error", err, "stornierung_id", stornierungID)
		return false, ErrDatabase
	}
	if !found {
		return false, nil
	}
	gleich, err := kasse.GleicheDirektverkaufStornierung(gespeichert, positionen)
	if err != nil {
		log.Error("Failed to compare booked direktverkauf stornierung", "error", err, "stornierung_id", stornierungID)
		return true, ErrDatabase
	}
	if !gleich {
		log.Warn("Direktverkauf-Stornierung mit anderem Inhalt bereits gebucht", "stornierung_id", stornierungID)
		return true, ErrVorgangAbweichend
	}
	log.Info("Idempotente Direktverkauf-Stornierung: stornierungId bereits vorhanden", "stornierung_id", stornierungID)
	return true, nil
}
