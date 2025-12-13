package application

import (
	"context"

	"github.com/nicograef/jotti/backend/admin/table/domain"
	"github.com/rs/zerolog"
)

type tableRepository interface {
	GetTable(ctx context.Context, id int) (domain.Table, error)
	CreateTable(ctx context.Context, t domain.Table) (int, error)
	UpdateTable(ctx context.Context, t domain.Table) error
}

type Command struct {
	TableRepo tableRepository
}

// CreateTable creates a new table in the database.
func (c Command) CreateTable(ctx context.Context, name string) (int, error) {
	log := zerolog.Ctx(ctx)

	table, err := domain.NewTable(name)
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

// UpdateTable updates an existing table in the database.
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

// ActivateTable sets the status of a table to active.
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

// DeactivateTable sets the status of a table to inactive.
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
