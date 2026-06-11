package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/nicograef/jotti/backend/config"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/settings"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/repository/settings_repo"
	"github.com/nicograef/jotti/backend/repository/tse_repo"
	"github.com/rs/zerolog/log"
)

const (
	tseNachsignierPollInterval = 5 * time.Second
	tseNachsignierBatchSize    = 20
)

type tseSettingsReader interface {
	GetTSEKonfiguration(ctx context.Context) (settings.TSEKonfiguration, error)
}

type tseNachsignierStore interface {
	GetOffeneTSENachsignierAuftraege(ctx context.Context, limit int) ([]tse_repo.OffenerNachsignierAuftrag, error)
	QuittiereTSENachsignierAuftrag(ctx context.Context, auftragID int, signatur tse_repo.Signatur) error
	TSENachsignierAuftragFehlversuch(ctx context.Context, auftragID int, fehler string) error
}

// tseWorkerClient beschreibt, was der Nachsignier-Worker von der TSE braucht:
// signieren und den Ist-Zustand einer Transaktion abfragen.
type tseWorkerClient interface {
	tse.TSEClient
	tse.TransactionRetriever
}

type tseClientFactory func(credentials tse.Credentials) (tseWorkerClient, error)

type tseNachsignierWorker struct {
	settingsRepo tseSettingsReader
	store        tseNachsignierStore
	newTSEClient tseClientFactory
	now          func() time.Time
}

func newTSENachsignierWorker(cfg config.Config, database *sql.DB) tseNachsignierWorker {
	settingsRepo := settings_repo.NewRepository(database)
	store := tse_repo.NewStore(database)

	return tseNachsignierWorker{
		settingsRepo: settingsRepo,
		store:        store,
		newTSEClient: func(credentials tse.Credentials) (tseWorkerClient, error) {
			return tse_repo.NewFiskalyTSEClient(cfg.FiskalyBaseURL, credentials, nil)
		},
		now: time.Now,
	}
}

func (w tseNachsignierWorker) run(ctx context.Context) {
	ticker := time.NewTicker(tseNachsignierPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.processOnce(ctx); err != nil {
				log.Error().Err(err).Msg("TSE-Nachsignier-Worker tick failed")
			}
		}
	}
}

func (w tseNachsignierWorker) processOnce(ctx context.Context) error {
	if w.settingsRepo == nil || w.store == nil || w.newTSEClient == nil {
		return nil
	}

	conf, err := w.settingsRepo.GetTSEKonfiguration(ctx)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil
		}
		return err
	}
	if !conf.IstKonfiguriert() {
		return nil
	}

	client, err := w.newTSEClient(tse.Credentials{
		ApiKey:    conf.ApiKey,
		ApiSecret: conf.ApiSecret,
		TssID:     conf.TssID,
		ClientID:  conf.ClientID,
	})
	if err != nil {
		log.Warn().Err(err).Msg("TSE-Nachsignier-Worker could not create TSE client")
		return nil
	}

	auftraege, err := w.store.GetOffeneTSENachsignierAuftraege(ctx, tseNachsignierBatchSize)
	if err != nil {
		return err
	}

	for _, auftrag := range auftraege {
		if err := w.processAuftrag(ctx, client, auftrag); err != nil {
			log.Warn().Err(err).Str("tx_id", auftrag.TxID).Int("auftrag_id", auftrag.ID).Msg("TSE-Nachsignierung fehlgeschlagen")
			if err := w.store.TSENachsignierAuftragFehlversuch(ctx, auftrag.ID, err.Error()); err != nil {
				log.Error().Err(err).Int("auftrag_id", auftrag.ID).Msg("Failed to record TSE-Nachsignier-Fehlversuch")
			}
		}
	}

	return nil
}

func (w tseNachsignierWorker) processAuftrag(ctx context.Context, client tseWorkerClient, auftrag tse_repo.OffenerNachsignierAuftrag) error {
	finishResult, startLogTime, err := w.beschaffeSignatur(ctx, client, auftrag)
	if err != nil {
		return err
	}

	logTimeStart := nonZeroWorkerTime(startLogTime, finishResult.LogTimeStart)
	if logTimeStart.IsZero() {
		logTimeStart = w.now().UTC()
	}
	logTimeEnd := nonZeroWorkerTime(finishResult.LogTime, finishResult.LogTimeEnd)
	if logTimeEnd.IsZero() {
		logTimeEnd = logTimeStart
	}

	return w.store.QuittiereTSENachsignierAuftrag(ctx, auftrag.ID, tse_repo.Signatur{
		TxID:              auftrag.TxID,
		TransaktionNummer: finishResult.TransactionNumber,
		SignaturZaehler:   finishResult.SignatureCounter,
		TSESeriennummer:   finishResult.SerialNumberTSE,
		LogTimeStart:      logTimeStart,
		LogTimeEnd:        logTimeEnd,
		Signatur:          finishResult.Signature,
		QRCodeData:        finishResult.QRCodeData,
	})
}

// beschaffeSignatur liefert die Signaturdaten fuer den Auftrag. Vor einem
// neuen Signierversuch wird der Ist-Zustand bei fiskaly abgefragt: Eine dort
// bereits abgeschlossene Transaktion wird direkt uebernommen statt erneut
// signiert (heilt das 409-Szenario nach Abbruch zwischen Signierung und
// Quittierung), eine noch aktive Transaktion wird nur noch abgeschlossen.
func (w tseNachsignierWorker) beschaffeSignatur(ctx context.Context, client tseWorkerClient, auftrag tse_repo.OffenerNachsignierAuftrag) (tse.FinishResult, time.Time, error) {
	vorhanden, err := client.RetrieveTransaction(ctx, auftrag.TxID)
	if errors.Is(err, tse.ErrTransactionNichtGefunden) {
		startResult, err := client.StartTransaction(ctx, auftrag.TxID)
		if err != nil {
			return tse.FinishResult{}, time.Time{}, err
		}
		finishResult, err := client.FinishTransaction(ctx, auftrag.TxID, auftrag.ProcessType, auftrag.ProcessData)
		if err != nil {
			return tse.FinishResult{}, time.Time{}, err
		}
		return finishResult, startResult.LogTime, nil
	}
	if err != nil {
		return tse.FinishResult{}, time.Time{}, err
	}

	switch vorhanden.State {
	case tse.TransactionStateFinished:
		return vorhanden.FinishResult, vorhanden.LogTimeStart, nil
	case tse.TransactionStateActive:
		finishResult, err := client.FinishTransaction(ctx, auftrag.TxID, auftrag.ProcessType, auftrag.ProcessData)
		if err != nil {
			return tse.FinishResult{}, time.Time{}, err
		}
		return finishResult, vorhanden.LogTimeStart, nil
	default:
		return tse.FinishResult{}, time.Time{}, fmt.Errorf("transaktion %s hat unerwarteten Zustand %q bei fiskaly", auftrag.TxID, vorhanden.State)
	}
}

func nonZeroWorkerTime(primary time.Time, fallback time.Time) time.Time {
	if !primary.IsZero() {
		return primary
	}
	return fallback
}
