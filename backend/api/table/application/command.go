package application

import (
	"context"
	"strconv"

	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/table"
	"github.com/rs/zerolog"
)

type tableRepoCommand interface {
	GetTable(ctx context.Context, id int) (table.Table, error)
	CreateTable(ctx context.Context, t table.Table) (int, error)
	UpdateTable(ctx context.Context, t table.Table) error
}

type eventRepoCommand interface {
	WriteEvent(ctx context.Context, event event.Event) (int, error)
	ReadEventsWithSnapshot(ctx context.Context, subject string, snapshotEventType string) ([]event.Event, error)
}

type Command struct {
	TableRepo tableRepoCommand
	EventRepo eventRepoCommand
}

func (c Command) CreateTable(ctx context.Context, name string) (int, error) {
	log := zerolog.Ctx(ctx)

	table, err := table.NewTable(name)
	if err != nil {
		log.Warn().Err(err).Str("table_name", name).Msg("Invalid table data")
		return 0, ErrInvalidTableData
	}

	id, err := c.TableRepo.CreateTable(ctx, table)
	if err != nil {
		return 0, fromRepositoryError(err, log, 0)
	}

	log.Info().Int("table_id", id).Msg("Table created")
	return id, nil
}

func (c Command) UpdateTable(ctx context.Context, id int, name string) error {
	log := zerolog.Ctx(ctx)

	table, err := c.TableRepo.GetTable(ctx, id)
	if err != nil {
		return fromRepositoryError(err, log, id)
	}

	err = table.Rename(name)
	if err != nil {
		log.Warn().Err(err).Int("table_id", id).Msg("Invalid table data for update")
		return ErrInvalidTableData
	}

	err = c.TableRepo.UpdateTable(ctx, table)
	if err != nil {
		return fromRepositoryError(err, log, id)
	}

	log.Info().Int("table_id", id).Msg("Table updated")
	return nil
}

func (c Command) ActivateTable(ctx context.Context, id int) error {
	log := zerolog.Ctx(ctx)

	table, err := c.TableRepo.GetTable(ctx, id)
	if err != nil {
		return fromRepositoryError(err, log, id)
	}

	table.Activate()

	err = c.TableRepo.UpdateTable(ctx, table)
	if err != nil {
		return fromRepositoryError(err, log, id)
	}

	log.Info().Int("table_id", id).Msg("Table activated")
	return nil
}

func (c Command) DeactivateTable(ctx context.Context, id int) error {
	log := zerolog.Ctx(ctx)
	table, err := c.TableRepo.GetTable(ctx, id)
	if err != nil {
		return fromRepositoryError(err, log, id)
	}

	table.Deactivate()

	err = c.TableRepo.UpdateTable(ctx, table)
	if err != nil {
		return fromRepositoryError(err, log, id)
	}

	log.Info().Int("table_id", id).Msg("Table deactivated")
	return nil
}

func (c Command) PlaceTableOrder(ctx context.Context, userID, tableID int, variants []table.LineItem, comment string) error {
	log := zerolog.Ctx(ctx)

	event, err := table.NewOrderPlacedEvent(userID, tableID, variants, comment)
	if err != nil {
		log.Error().Err(err).Int("table_id", tableID).Msg("Failed to create order placed event")
		return err
	}

	_, err = c.EventRepo.WriteEvent(ctx, event)
	if err != nil {
		log.Error().Int("table_id", tableID).Msg("Failed to write order placed event to database")
		return ErrDatabase
	}

	log.Info().Int("table_id", tableID).Msg("Order placed")
	return nil
}

func (c Command) RegisterTablePayment(ctx context.Context, userID, tableID int, variants []table.LineItem, comment string) error {
	log := zerolog.Ctx(ctx)

	event, err := table.NewPaymentRegisteredEvent(userID, tableID, variants, comment)
	if err != nil {
		log.Error().Err(err).Int("table_id", tableID).Msg("Failed to create payment registered event")
		return err
	}

	_, err = c.EventRepo.WriteEvent(ctx, event)
	if err != nil {
		log.Error().Int("table_id", tableID).Msg("Failed to write payment registered event to database")
		return ErrDatabase
	}

	log.Info().Int("table_id", tableID).Msg("Payment registered")
	return nil
}

func (c Command) CancelTableVariants(ctx context.Context, userID, tableID int, variants []table.LineItem, comment string) error {
	log := zerolog.Ctx(ctx)

	event, err := table.NewVariantsCanceledEvent(userID, tableID, variants, comment)
	if err != nil {
		log.Error().Err(err).Int("table_id", tableID).Msg("Failed to create variants canceled event")
		return err
	}

	_, err = c.EventRepo.WriteEvent(ctx, event)
	if err != nil {
		log.Error().Int("table_id", tableID).Msg("Failed to write variants canceled event to database")
		return ErrDatabase
	}

	log.Info().Int("table_id", tableID).Msg("Variants canceled")
	return nil
}

func (c Command) DeliverTableVariants(ctx context.Context, userID, tableID int, variants []table.LineItem, comment string) error {
	log := zerolog.Ctx(ctx)

	event, err := table.NewVariantsDeliveredEvent(userID, tableID, variants, comment)
	if err != nil {
		log.Error().Err(err).Int("table_id", tableID).Msg("Failed to create variants delivered event")
		return err
	}

	_, err = c.EventRepo.WriteEvent(ctx, event)
	if err != nil {
		log.Error().Int("table_id", tableID).Msg("Failed to write variants delivered event to database")
		return ErrDatabase
	}

	log.Info().Int("table_id", tableID).Msg("Variants delivered")
	return nil
}

func (c Command) CreateTableSnapshot(ctx context.Context, userID, tableID int) error {
	log := zerolog.Ctx(ctx)

	subject := "table:" + strconv.Itoa(tableID)
	events, err := c.EventRepo.ReadEventsWithSnapshot(ctx, subject, string(table.EventTypeSnapshotV1))
	if err != nil {
		log.Error().Err(err).Int("table_id", tableID).Msg("Failed to read events for snapshot")
		return ErrDatabase
	}

	balance, err := table.GetBalanceFromEvents(events)
	if err != nil {
		return err
	}
	unpaid, err := table.GetUnpaidVariantsFromEvents(events)
	if err != nil {
		return err
	}
	undelivered, err := table.GetUndeliveredVariantsFromEvents(events)
	if err != nil {
		return err
	}
	totalPayment, err := table.GetTotalPaymentsFromEvents(events)
	if err != nil {
		return err
	}

	log.Debug().
		Int("table_id", tableID).
		Int("balance", balance).
		Int("unpaid_count", len(unpaid)).
		Int("undelivered_count", len(undelivered)).
		Int("total_payments", totalPayment).
		Msg("Creating snapshot with computed state")

	snapshotEvent, err := table.NewSnapshotEvent(userID, tableID, balance, unpaid, undelivered, totalPayment)
	if err != nil {
		log.Error().Err(err).Int("table_id", tableID).Msg("Failed to create snapshot event")
		return err
	}

	_, err = c.EventRepo.WriteEvent(ctx, snapshotEvent)
	if err != nil {
		log.Error().Int("table_id", tableID).Msg("Failed to write snapshot event")
		return ErrDatabase
	}

	log.Info().Int("table_id", tableID).Int("balance", balance).Msg("Snapshot created")
	return nil
}
