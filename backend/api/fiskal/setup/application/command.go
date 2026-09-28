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

// kassensitzungReader reports an aktive Kassensitzung, i.e. offen or wird_abgeschlossen.
type kassensitzungReader interface {
	GetAktiveKassensitzung(ctx context.Context) (*kasse.Kassensitzung, error)
}

type Command struct {
	TSERepo             tseCommandRepo
	KassensitzungenRepo kassensitzungReader
	NewTSESetupClient   NewTSESetupClient
}

// ensureKeineAktiveKassensitzung keeps the signing device from changing mid-Kassentag
// (docs/handbuch.md §3.13). wird_abgeschlossen counts too: a closing that still signs belongs to
// the old TSS.
func (c Command) ensureKeineAktiveKassensitzung(ctx context.Context) error {
	log := zerolog.Ctx(ctx)

	aktive, err := c.KassensitzungenRepo.GetAktiveKassensitzung(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check for aktive Kassensitzung before TSE config change")
		return ErrDatabase
	}
	if aktive != nil {
		return ErrTSEKonfigurationKassensitzungOffen
	}
	return nil
}

// UpdateTSEKonfiguration saves a hand-entered configuration under the setup lock (einrichtungLaeuft).
// Without it the last writer would win and the instance would sign against a TSS it was not set up with.
func (c Command) UpdateTSEKonfiguration(ctx context.Context, conf tse.Konfiguration) error {
	log := zerolog.Ctx(ctx)

	freigeben, err := acquireEinrichtung()
	if err != nil {
		return err
	}
	defer freigeben()

	if err := c.ensureKeineAktiveKassensitzung(ctx); err != nil {
		return err
	}

	// SaveEinrichtung, because this path can also make the transition to configured;
	// otherwise the keine_konfiguration outage would stay open forever.
	if err := c.TSERepo.SaveEinrichtung(ctx, conf); err != nil {
		log.Error().Err(err).Msg("Failed to save tse_konfiguration")
		return ErrDatabase
	}

	log.Info().
		Bool("ist_konfiguriert", conf.IstKonfiguriert()).
		Msg("TSE-Konfiguration saved")

	return nil
}
