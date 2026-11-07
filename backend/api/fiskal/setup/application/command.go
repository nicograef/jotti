package application

import (
	"context"
	"errors"

	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/internal/logging"
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
	log := logging.FromContext(ctx)

	aktive, err := c.KassensitzungenRepo.GetAktiveKassensitzung(ctx)
	if err != nil {
		log.Error("Failed to check for aktive Kassensitzung before TSE config change", "error", err)
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
	log := logging.FromContext(ctx)

	freigeben, err := acquireEinrichtung()
	if err != nil {
		return err
	}
	defer freigeben()

	if err := c.ensureKeineAktiveKassensitzung(ctx); err != nil {
		return err
	}

	// tse.csv of the DSFinV-K export names the configured TSS, so its Stammdaten are stored first;
	// they are keyed per TSS, so a later failed save leaves nothing wrong behind.
	if conf.IstKonfiguriert() {
		if err := c.speichereStammdatenVonHand(ctx, conf); err != nil {
			return err
		}
	}

	// SaveEinrichtung, because this path can also make the transition to configured;
	// otherwise the keine_konfiguration outage would stay open forever.
	if err := c.TSERepo.SaveEinrichtung(ctx, conf); err != nil {
		log.Error("Failed to save tse_konfiguration", "error", err)
		return ErrDatabase
	}

	log.Info("TSE-Konfiguration saved", "ist_konfiguriert", conf.IstKonfiguriert())

	return nil
}

func (c Command) speichereStammdatenVonHand(ctx context.Context, conf tse.Konfiguration) error {
	log := logging.FromContext(ctx)

	client, err := c.NewTSESetupClient(tse.SetupCredentials{ApiKey: conf.ApiKey, ApiSecret: conf.ApiSecret})
	if err != nil {
		log.Error("Failed to create TSE setup client for manual configuration", "error", err)
		return ErrTSEVerbindungFehlgeschlagen
	}
	stammdaten, err := client.RetrieveTSSStammdaten(ctx, conf.TssID)
	if err != nil {
		if errors.Is(err, tse.ErrSetupAuthFehlgeschlagen) {
			return ErrTSESetupZugangsdaten
		}
		log.Warn("Failed to retrieve TSE Stammdaten for manual configuration", "error", err, "tss_id", conf.TssID)
		return ErrTSEVerbindungFehlgeschlagen
	}
	if err := c.TSERepo.UpsertTSEStammdaten(ctx, stammdaten); err != nil {
		log.Error("Failed to save TSE Stammdaten for manual configuration", "error", err, "tss_id", conf.TssID)
		return ErrDatabase
	}
	return nil
}
