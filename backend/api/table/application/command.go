package application

import (
	"context"
	"strconv"

	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/table"
	"github.com/rs/zerolog"
)

type tableRepoCommand interface {
	GetTable(ctx context.Context, id int) (table.Tisch, error)
	CreateTable(ctx context.Context, t table.Tisch) (int, error)
	UpdateTable(ctx context.Context, t table.Tisch) error
}

type eventRepoCommand interface {
	WriteEvent(ctx context.Context, event event.Event) (int, error)
	ReadEventsWithSnapshot(ctx context.Context, subject string, snapshotEventType string) ([]event.Event, error)
}

type Command struct {
	TableRepo tableRepoCommand
	EventRepo eventRepoCommand
}

func (c Command) TischErstellen(ctx context.Context, name string) (int, error) {
	log := zerolog.Ctx(ctx)

	tisch, err := table.NewTisch(name)
	if err != nil {
		log.Warn().Err(err).Str("tisch_name", name).Msg("Invalid tisch data")
		return 0, ErrInvalidTischData
	}

	id, err := c.TableRepo.CreateTable(ctx, tisch)
	if err != nil {
		return 0, fromRepositoryError(err, log, 0)
	}

	log.Info().Int("tisch_id", id).Msg("Tisch created")
	return id, nil
}

func (c Command) TischAktualisieren(ctx context.Context, id int, name string) error {
	log := zerolog.Ctx(ctx)

	tisch, err := c.TableRepo.GetTable(ctx, id)
	if err != nil {
		return fromRepositoryError(err, log, id)
	}

	err = tisch.Rename(name)
	if err != nil {
		log.Warn().Err(err).Int("tisch_id", id).Msg("Invalid tisch data for update")
		return ErrInvalidTischData
	}

	err = c.TableRepo.UpdateTable(ctx, tisch)
	if err != nil {
		return fromRepositoryError(err, log, id)
	}

	log.Info().Int("tisch_id", id).Msg("Tisch updated")
	return nil
}

func (c Command) TischAktivieren(ctx context.Context, id int) error {
	log := zerolog.Ctx(ctx)

	tisch, err := c.TableRepo.GetTable(ctx, id)
	if err != nil {
		return fromRepositoryError(err, log, id)
	}

	tisch.Activate()

	err = c.TableRepo.UpdateTable(ctx, tisch)
	if err != nil {
		return fromRepositoryError(err, log, id)
	}

	log.Info().Int("tisch_id", id).Msg("Tisch activated")
	return nil
}

func (c Command) TischDeaktivieren(ctx context.Context, id int) error {
	log := zerolog.Ctx(ctx)
	tisch, err := c.TableRepo.GetTable(ctx, id)
	if err != nil {
		return fromRepositoryError(err, log, id)
	}

	tisch.Deactivate()

	err = c.TableRepo.UpdateTable(ctx, tisch)
	if err != nil {
		return fromRepositoryError(err, log, id)
	}

	log.Info().Int("tisch_id", id).Msg("Tisch deactivated")
	return nil
}

func (c Command) BestellungAufgeben(ctx context.Context, userID, tischID int, positionen []table.Position, comment string) error {
	log := zerolog.Ctx(ctx)

	event, err := table.NewBestellungAufgegebenEvent(userID, tischID, positionen, comment)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to create bestellung aufgegeben event")
		return err
	}

	_, err = c.EventRepo.WriteEvent(ctx, event)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to write bestellung aufgegeben event to database")
		return ErrDatabase
	}

	log.Info().Int("tisch_id", tischID).Msg("Bestellung aufgegeben")
	return nil
}

func (c Command) ZahlungRegistrieren(ctx context.Context, userID, tischID int, positionen []table.Position, comment string) error {
	log := zerolog.Ctx(ctx)

	event, err := table.NewZahlungRegistriertEvent(userID, tischID, positionen, comment)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to create zahlung registriert event")
		return err
	}

	_, err = c.EventRepo.WriteEvent(ctx, event)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to write zahlung registriert event to database")
		return ErrDatabase
	}

	log.Info().Int("tisch_id", tischID).Msg("Zahlung registriert")
	return nil
}

func (c Command) ProdukteStornieren(ctx context.Context, userID, tischID int, positionen []table.Position, comment string) error {
	log := zerolog.Ctx(ctx)

	event, err := table.NewProdukteStorniertEvent(userID, tischID, positionen, comment)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to create produkte storniert event")
		return err
	}

	_, err = c.EventRepo.WriteEvent(ctx, event)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to write produkte storniert event to database")
		return ErrDatabase
	}

	log.Info().Int("tisch_id", tischID).Msg("Produkte storniert")
	return nil
}

func (c Command) ProdukteLiefern(ctx context.Context, userID, tischID int, positionen []table.Position, comment string) error {
	log := zerolog.Ctx(ctx)

	event, err := table.NewProdukteGeliefertEvent(userID, tischID, positionen, comment)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to create produkte geliefert event")
		return err
	}

	_, err = c.EventRepo.WriteEvent(ctx, event)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to write produkte geliefert event to database")
		return ErrDatabase
	}

	log.Info().Int("tisch_id", tischID).Msg("Produkte geliefert")
	return nil
}

func (c Command) TischSnapshotErstellen(ctx context.Context, userID, tischID int) error {
	log := zerolog.Ctx(ctx)

	subject := "tisch:" + strconv.Itoa(tischID)
	events, err := c.EventRepo.ReadEventsWithSnapshot(ctx, subject, string(table.EventTypeSnapshotV1))
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to read events for snapshot")
		return ErrDatabase
	}

	saldo, err := table.GetSaldoFromEvents(events)
	if err != nil {
		return err
	}
	unbezahlt, err := table.GetUnbezahltePositionenFromEvents(events)
	if err != nil {
		return err
	}
	ungeliefert, err := table.GetUngeliefertePositionenFromEvents(events)
	if err != nil {
		return err
	}
	gesamtZahlungen, err := table.GetGesamtZahlungenFromEvents(events)
	if err != nil {
		return err
	}

	log.Debug().
		Int("tisch_id", tischID).
		Int("saldo", saldo).
		Int("unbezahlt_count", len(unbezahlt)).
		Int("ungeliefert_count", len(ungeliefert)).
		Int("gesamt_zahlungen", gesamtZahlungen).
		Msg("Creating snapshot with computed state")

	snapshotEvent, err := table.NewSnapshotEvent(userID, tischID, saldo, unbezahlt, ungeliefert, gesamtZahlungen)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to create snapshot event")
		return err
	}

	_, err = c.EventRepo.WriteEvent(ctx, snapshotEvent)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to write snapshot event")
		return ErrDatabase
	}

	log.Info().Int("tisch_id", tischID).Int("saldo", saldo).Msg("Snapshot created")
	return nil
}
