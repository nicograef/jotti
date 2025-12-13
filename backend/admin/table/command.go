package table

import (
	"context"
	"errors"

	"github.com/nicograef/jotti/backend/db"
	"github.com/rs/zerolog"
)

type commandPersistence interface {
	CreateTable(ctx context.Context, name string) (int, error)
	UpdateTable(ctx context.Context, id int, name string) error
	ActivateTable(ctx context.Context, id int) error
	DeactivateTable(ctx context.Context, id int) error
}

type Command struct {
	Persistence commandPersistence
}

// CreateTable creates a new table in the database.
func (s *Command) CreateTable(ctx context.Context, name string) (int, error) {
	log := zerolog.Ctx(ctx)

	id, err := s.Persistence.CreateTable(ctx, name)
	if err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			log.Warn().Str("table_name", name).Msg("Table name already exists")
			return 0, ErrTableAlreadyExists
		} else {
			log.Error().Str("table_name", name).Msg("Failed to create table")
			return 0, ErrDatabase
		}
	}

	log.Info().Int("table_id", id).Msg("Table created")
	return id, nil
}

// UpdateTable updates an existing table in the database.
func (s *Command) UpdateTable(ctx context.Context, id int, name string) error {
	log := zerolog.Ctx(ctx)

	err := s.Persistence.UpdateTable(ctx, id, name)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("table_id", id).Msg("Table not found for update")
			return ErrTableNotFound
		} else {
			log.Error().Int("table_id", id).Msg("Failed to update table")
			return ErrDatabase
		}
	}

	log.Info().Int("table_id", id).Msg("Table updated")
	return nil
}

// ActivateTable sets the status of a table to active.
func (s *Command) ActivateTable(ctx context.Context, id int) error {
	log := zerolog.Ctx(ctx)

	err := s.Persistence.ActivateTable(ctx, id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("table_id", id).Msg("Table not found for activation")
			return ErrTableNotFound
		} else {
			log.Error().Int("table_id", id).Msg("Failed to activate table")
			return ErrDatabase
		}
	}

	log.Info().Int("table_id", id).Msg("Table activated")
	return nil
}

// DeactivateTable sets the status of a table to inactive.
func (s *Command) DeactivateTable(ctx context.Context, id int) error {
	log := zerolog.Ctx(ctx)

	err := s.Persistence.DeactivateTable(ctx, id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("table_id", id).Msg("Table not found for deactivation")
			return ErrTableNotFound
		} else {
			log.Error().Int("table_id", id).Msg("Failed to deactivate table")
			return ErrDatabase
		}
	}

	log.Info().Int("table_id", id).Msg("Table deactivated")
	return nil
}
