package application

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/settings"
	"github.com/rs/zerolog"
)

type settingsCommandRepo interface {
	UpsertBetreiber(ctx context.Context, b settings.Betreiber) error
	UpsertBondruckEinstellungen(ctx context.Context, b settings.BondruckEinstellungen) error
	UpsertTSEKonfiguration(ctx context.Context, b settings.TSEKonfiguration) error
}

type Command struct {
	SettingsRepo settingsCommandRepo
}

func (c Command) UpdateBetreiber(ctx context.Context, b settings.Betreiber) error {
	log := zerolog.Ctx(ctx)

	if err := c.SettingsRepo.UpsertBetreiber(ctx, b); err != nil {
		log.Error().Err(err).Msg("Failed to save betreiber")
		return ErrDatabase
	}
	log.Info().Str("vereinsname", b.Vereinsname).Msg("Betreiber saved")
	return nil
}

func (c Command) UpdateBondruckEinstellungen(ctx context.Context, b settings.BondruckEinstellungen) error {
	log := zerolog.Ctx(ctx)

	if err := c.SettingsRepo.UpsertBondruckEinstellungen(ctx, b); err != nil {
		log.Error().Err(err).Msg("Failed to save bondruck_einstellungen")
		return ErrDatabase
	}
	log.Info().
		Str("kassenbeleg_drucker_ip", b.KassenbelegDruckerIP).
		Str("direktverkauf_modus", string(b.DirektverkaufModus)).
		Str("abholbon_drucker_ip", b.AbholbonDruckerIP).
		Msg("Bondruck-Einstellungen saved")
	return nil
}

func (c Command) UpdateTSEKonfiguration(ctx context.Context, conf settings.TSEKonfiguration) error {
	log := zerolog.Ctx(ctx)

	if err := c.SettingsRepo.UpsertTSEKonfiguration(ctx, conf); err != nil {
		log.Error().Err(err).Msg("Failed to save tse_konfiguration")
		return ErrDatabase
	}

	log.Info().
		Bool("ist_konfiguriert", conf.IstKonfiguriert()).
		Msg("TSE-Konfiguration saved")

	return nil
}
