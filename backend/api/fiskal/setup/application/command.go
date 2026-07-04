package application

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/rs/zerolog"
)

type settingsCommandRepo interface {
	SpeichereEinrichtung(ctx context.Context, c tse.Konfiguration) error
	UpsertTSEStammdaten(ctx context.Context, s tse.Stammdaten) error
	GetKassenidentitaet(ctx context.Context) (tse.Kassenidentitaet, error)
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

func (c Command) UpdateTSEKonfiguration(ctx context.Context, conf tse.Konfiguration) error {
	log := zerolog.Ctx(ctx)

	if err := c.pruefeKeineOffeneKassensitzung(ctx); err != nil {
		return err
	}

	// Auch der direkte Zugangsdaten-Pfad speichert ueber SpeichereEinrichtung:
	// Fuehrt er den Uebergang zu konfiguriert aus, laufen Einrichtungs-Sweep und
	// das Schliessen des keine_konfiguration-Stoerungszeitraums in derselben
	// Transaktion — sonst bliebe der Zeitraum fuer immer offen.
	if err := c.SettingsRepo.SpeichereEinrichtung(ctx, conf); err != nil {
		log.Error().Err(err).Msg("Failed to save tse_konfiguration")
		return ErrDatabase
	}

	log.Info().
		Bool("ist_konfiguriert", conf.IstKonfiguriert()).
		Msg("TSE-Konfiguration saved")

	return nil
}
