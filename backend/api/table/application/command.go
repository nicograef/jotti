package application

import (
	"context"

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

func (c Command) PlaceOrder(ctx context.Context, userID, tableID int, products []table.OrderProduct) error {
	log := zerolog.Ctx(ctx)

	event, err := table.NewOrderPlacedEvent(userID, tableID, products)
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
