package application

import (
	"context"
	"errors"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/settings"
	"github.com/rs/zerolog"
)

type settingsQueryRepo interface {
	GetKassenidentitaet(ctx context.Context) (settings.Kassenidentitaet, error)
	GetBetreiber(ctx context.Context) (settings.Betreiber, error)
	GetBondruckEinstellungen(ctx context.Context) (settings.BondruckEinstellungen, error)
	GetTSEKonfiguration(ctx context.Context) (settings.TSEKonfiguration, error)
}

type Query struct {
	SettingsRepo settingsQueryRepo
}

func (q Query) GetKassenidentitaet(ctx context.Context) (settings.Kassenidentitaet, error) {
	log := zerolog.Ctx(ctx)

	identitaet, err := q.SettingsRepo.GetKassenidentitaet(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve kassenidentitaet")
		return settings.Kassenidentitaet{}, ErrDatabase
	}
	return identitaet, nil
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

func (q Query) GetBondruckEinstellungen(ctx context.Context) (settings.BondruckEinstellungen, error) {
	log := zerolog.Ctx(ctx)

	b, err := q.SettingsRepo.GetBondruckEinstellungen(ctx)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return settings.BondruckEinstellungen{}, ErrNotFound
		}
		log.Error().Err(err).Msg("Failed to retrieve bondruck_einstellungen")
		return settings.BondruckEinstellungen{}, ErrDatabase
	}
	return b, nil
}

func (q Query) GetTSEKonfiguration(ctx context.Context) (settings.TSEKonfiguration, error) {
	log := zerolog.Ctx(ctx)

	c, err := q.SettingsRepo.GetTSEKonfiguration(ctx)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return settings.TSEKonfiguration{}, ErrNotFound
		}
		log.Error().Err(err).Msg("Failed to retrieve tse_konfiguration")
		return settings.TSEKonfiguration{}, ErrDatabase
	}

	return c, nil
}
