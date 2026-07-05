package application

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/tisch"
	"github.com/rs/zerolog"
)

type tableRepo interface {
	GetTable(ctx context.Context, id int) (tisch.Tisch, error)
	CreateTable(ctx context.Context, t tisch.Tisch) (int, error)
	UpdateTable(ctx context.Context, t tisch.Tisch) error
	GetAllTables(ctx context.Context) ([]tisch.Tisch, error)
}

type favoritRepo interface {
	Add(ctx context.Context, userID, tischID int) error
	Remove(ctx context.Context, userID, tischID int) error
}

type Command struct {
	TableRepo   tableRepo
	FavoritRepo favoritRepo
}

func (c Command) FavoritHinzufuegen(ctx context.Context, userID, tischID int) error {
	log := zerolog.Ctx(ctx)

	t, err := c.TableRepo.GetTable(ctx, tischID)
	if err != nil {
		return fromRepositoryError(err, log, tischID)
	}

	if t.Status != tisch.ActiveStatus {
		log.Warn().Int("tisch_id", tischID).Str("status", string(t.Status)).Msg("Tisch is not active")
		return ErrTischNotActive
	}

	if err := c.FavoritRepo.Add(ctx, userID, tischID); err != nil {
		log.Error().Err(err).Int("user_id", userID).Int("tisch_id", tischID).Msg("Failed to add favorit")
		return ErrDatabase
	}

	log.Info().Int("user_id", userID).Int("tisch_id", tischID).Msg("Favorit added")
	return nil
}

func (c Command) FavoritEntfernen(ctx context.Context, userID, tischID int) error {
	log := zerolog.Ctx(ctx)

	if err := c.FavoritRepo.Remove(ctx, userID, tischID); err != nil {
		log.Error().Err(err).Int("user_id", userID).Int("tisch_id", tischID).Msg("Failed to remove favorit")
		return ErrDatabase
	}

	log.Info().Int("user_id", userID).Int("tisch_id", tischID).Msg("Favorit removed")
	return nil
}

func (c Command) TischErstellen(ctx context.Context, name string) (int, error) {
	log := zerolog.Ctx(ctx)

	tisch, err := tisch.NewTisch(name)
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
	return c.applyTischStatusChange(ctx, id, "Tisch activated", func(t *tisch.Tisch) { t.Activate() })
}

func (c Command) TischDeaktivieren(ctx context.Context, id int) error {
	return c.applyTischStatusChange(ctx, id, "Tisch deactivated", func(t *tisch.Tisch) { t.Deactivate() })
}

func (c Command) TischLoeschen(ctx context.Context, id int) error {
	return c.applyTischStatusChange(ctx, id, "Tisch deleted", func(t *tisch.Tisch) { t.Delete() })
}

func (c Command) applyTischStatusChange(ctx context.Context, id int, successMsg string, action func(*tisch.Tisch)) error {
	log := zerolog.Ctx(ctx)

	tisch, err := c.TableRepo.GetTable(ctx, id)
	if err != nil {
		return fromRepositoryError(err, log, id)
	}
	action(&tisch)
	if err := c.TableRepo.UpdateTable(ctx, tisch); err != nil {
		return fromRepositoryError(err, log, id)
	}
	log.Info().Int("tisch_id", id).Msg(successMsg)
	return nil
}
