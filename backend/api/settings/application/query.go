package application

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/settings"
	"github.com/rs/zerolog"
)

type settingsQueryRepo interface {
	GetSystemConfig(ctx context.Context) (settings.SystemConfig, error)
	GetBetreiber(ctx context.Context) (settings.Betreiber, error)
}

type Query struct {
	SettingsRepo settingsQueryRepo
}

func (q Query) GetSeriennummer(ctx context.Context) (uuid.UUID, error) {
	log := zerolog.Ctx(ctx)

	cfg, err := q.SettingsRepo.GetSystemConfig(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve system config")
		return uuid.UUID{}, ErrDatabase
	}
	return cfg.Seriennummer, nil
}

func (q Query) GetBetreiber(ctx context.Context) (settings.Betreiber, error) {
	log := zerolog.Ctx(ctx)

	b, err := q.SettingsRepo.GetBetreiber(ctx)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return settings.Betreiber{}, ErrNotFound
		}
		log.Error().Err(err).Msg("Failed to retrieve betreiber")
		return settings.Betreiber{}, ErrDatabase
	}
	return b, nil
}
