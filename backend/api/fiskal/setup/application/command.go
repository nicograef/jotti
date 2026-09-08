package application

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/rs/zerolog"
)

type tseCommandRepo interface {
	SaveEinrichtung(ctx context.Context, c tse.Konfiguration) error
	UpsertTSEStammdaten(ctx context.Context, s tse.Stammdaten) error
	GetKassenidentitaet(ctx context.Context) (tse.Kassenidentitaet, error)
}

// kassensitzungReader meldet, ob gerade eine Kassensitzung offen ist. Änderungen
// der TSE-Konfiguration sind nur ohne offene Kassensitzung erlaubt: Das
// Signaturgeraet darf nicht mitten in einem laufenden Kassentag wechseln.
type kassensitzungReader interface {
	GetOffeneKassensitzung(ctx context.Context) (*kasse.Kassensitzung, error)
}

type Command struct {
	TSERepo             tseCommandRepo
	KassensitzungenRepo kassensitzungReader
	NewTSESetupClient   NewTSESetupClient
}

// ensureKeineOffeneKassensitzung lehnt eine TSE-Konfigurationsänderung ab,
// solange eine Kassensitzung offen ist (gemeinsamer Guard aller drei
// Änderungspfade: Neuanlage, Übernahme, Zugangsdaten-Wechsel).
func (c Command) ensureKeineOffeneKassensitzung(ctx context.Context) error {
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

// UpdateTSEKonfiguration speichert eine von Hand eingetragene TSE-Konfiguration.
// Sie nimmt dasselbe Schloss wie Neuanlage und Übernahme (einrichtungLaeuft in
// setup.go): Alle drei schreiben über SaveEinrichtung dieselbe Konfiguration,
// und in der Oberfläche liegt dieser Pfad direkt unter dem Einrichtungs-Wizard.
// Ohne das Schloss gewänne der letzte Schreiber, und die Instanz signierte
// danach gegen eine TSS/Client-Kombination, die nicht die eingerichtete ist.
func (c Command) UpdateTSEKonfiguration(ctx context.Context, conf tse.Konfiguration) error {
	log := zerolog.Ctx(ctx)

	freigeben, err := acquireEinrichtung()
	if err != nil {
		return err
	}
	defer freigeben()

	if err := c.ensureKeineOffeneKassensitzung(ctx); err != nil {
		return err
	}

	// Auch der direkte Zugangsdaten-Pfad speichert über SaveEinrichtung:
	// Führt er den Übergang zu konfiguriert aus, laufen Einrichtungs-Sweep und
	// das Schließen des keine_konfiguration-Störungszeitraums in derselben
	// Transaktion — sonst bliebe der Zeitraum für immer offen.
	if err := c.TSERepo.SaveEinrichtung(ctx, conf); err != nil {
		log.Error().Err(err).Msg("Failed to save tse_konfiguration")
		return ErrDatabase
	}

	log.Info().
		Bool("ist_konfiguriert", conf.IstKonfiguriert()).
		Msg("TSE-Konfiguration saved")

	return nil
}
