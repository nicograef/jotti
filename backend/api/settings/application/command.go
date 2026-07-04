package application

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/settings"
	"github.com/rs/zerolog"
)

type settingsCommandRepo interface {
	UpsertBetreiber(ctx context.Context, b settings.Betreiber) error
	UpsertTSEKonfiguration(ctx context.Context, b settings.TSEKonfiguration) error
	SpeichereEinrichtung(ctx context.Context, c settings.TSEKonfiguration) error
	UpsertTSEStammdaten(ctx context.Context, s settings.TSEStammdaten) error
	GetKassenidentitaet(ctx context.Context) (settings.Kassenidentitaet, error)
}

// kassensitzungReader meldet, ob gerade eine Kassensitzung offen ist. Aenderungen
// der TSE-Konfiguration sind nur ohne offene Kassensitzung erlaubt: Das
// Signaturgeraet darf nicht mitten in einem laufenden Kassentag wechseln.
type kassensitzungReader interface {
	GetOffeneKassensitzung(ctx context.Context) (*kasse.Kassensitzung, error)
}

type Command struct {
	SettingsRepo        settingsCommandRepo
	KassensitzungenRepo kassensitzungReader
	NewTSESetupClient   NewTSESetupClient
}

// pruefeKeineOffeneKassensitzung lehnt eine TSE-Konfigurationsaenderung ab,
// solange eine Kassensitzung offen ist (gemeinsamer Guard aller drei
// Aenderungspfade: Neuanlage, Uebernahme, Zugangsdaten-Wechsel).
func (c Command) pruefeKeineOffeneKassensitzung(ctx context.Context) error {
	log := zerolog.Ctx(ctx)

	offene, err := c.KassensitzungenRepo.GetOffeneKassensitzung(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check for offene Kassensitzung before TSE config change")
		return ErrDatabase
	}
	if offene != nil {
		return ErrTSEKonfigurationKassensitzungOffen
	}
	return nil
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

func (c Command) UpdateTSEKonfiguration(ctx context.Context, conf settings.TSEKonfiguration) error {
	log := zerolog.Ctx(ctx)

	if err := c.pruefeKeineOffeneKassensitzung(ctx); err != nil {
		return err
	}

	if err := c.SettingsRepo.UpsertTSEKonfiguration(ctx, conf); err != nil {
		log.Error().Err(err).Msg("Failed to save tse_konfiguration")
		return ErrDatabase
	}

	log.Info().
		Bool("ist_konfiguriert", conf.IstKonfiguriert()).
		Msg("TSE-Konfiguration saved")

	return nil
}
