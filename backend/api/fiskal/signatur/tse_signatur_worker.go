package signatur

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/repository/tse_repo"
	"github.com/rs/zerolog/log"
)

const (
	// tseSignaturPollInterval catches triggers lost between commit and trigger and delivers backoff retries.
	tseSignaturPollInterval = 5 * time.Second
	tseSignaturBatchSize    = 20
	// tseSignaturDurchlaufDeadline keeps a hanging run from blocking the serial worker forever.
	// Hitting it counts as a TSE-wide error.
	tseSignaturDurchlaufDeadline = 2 * time.Minute
	// Outage backoff doubles per TSE-wide error series up to the cap, sparing fiskaly the backlog.
	// No jitter: a single serial worker has nothing to desynchronise, and tests stay deterministic.
	tseStoerungBackoffBasis  = 5 * time.Second
	tseStoerungBackoffDeckel = 2 * time.Minute
	// tseSignaturWorkerLockKey is an arbitrary Postgres advisory lock key; only its holder talks to the TSE.
	tseSignaturWorkerLockKey = 823914502
)

type tseSettingsReader interface {
	GetTSEKonfiguration(ctx context.Context) (tse.Konfiguration, error)
}

type tseSignaturStore interface {
	GetOffeneTSESignaturauftraege(ctx context.Context, limit int) ([]tse_repo.OffenerSignaturauftrag, error)
	QuittiereTSESignaturauftrag(ctx context.Context, auftragID int, signatur tse.Signatur) error
	TSESignaturauftragFehlversuch(ctx context.Context, auftragID int, fehler string) error
	MarkOffeneAlsNichtKonfiguriert(ctx context.Context) (int64, error)
	OpenTSEStoerung(ctx context.Context, grundArt string, fehlertext string) error
	CloseTSEStoerung(ctx context.Context, grundArt string) error
}

type tseWorkerClient interface {
	tse.TSEClient
	tse.TransactionRetriever
}

type tseClientFactory func(credentials tse.Credentials) (tseWorkerClient, error)

// tseSignaturWorker is the only speaker for TSE signature transactions.
// See docs/handbuch.md §3.13 (Signatur-Worker).
type tseSignaturWorker struct {
	// lockDB nil (unit tests) skips the advisory lock.
	lockDB       *sql.DB
	settingsRepo tseSettingsReader
	store        tseSignaturStore
	newTSEClient tseClientFactory
	trigger      <-chan struct{}
	// pollInterval 0 falls back to tseSignaturPollInterval.
	pollInterval time.Duration
	// durchlaufDeadline 0 falls back to tseSignaturDurchlaufDeadline.
	durchlaufDeadline time.Duration
	now               func() time.Time

	lockConn *sql.Conn
	lockHeld bool

	// Outage state: no TSE calls before stoerungNaechsterVersuch; the next run's first job is the half-open probe.
	stoerungNaechsterVersuch time.Time
	stoerungSerie            int

	// client keeps its auth token across runs and is rebuilt only when the credentials change.
	client      tseWorkerClient
	clientCreds tse.Credentials
}

type Runner interface {
	Run(ctx context.Context)
}

// recoverPanic, deferred in each tick, logs a panic so the run loop survives to the next tick.
func recoverPanic(worker string) {
	if r := recover(); r != nil {
		log.Error().Interface("panic", r).Bytes("stack", debug.Stack()).Msg(worker + ": Panic im Durchlauf abgefangen; Loop laeuft weiter")
	}
}

// NewTSESignaturWorker takes fiskalyBaseURL as a parameter so this package does not import config.
func NewTSESignaturWorker(fiskalyBaseURL string, database *sql.DB) Runner {
	return &tseSignaturWorker{
		lockDB:       database,
		settingsRepo: tse_repo.NewRepository(database),
		store:        tse_repo.NewRepository(database),
		newTSEClient: func(credentials tse.Credentials) (tseWorkerClient, error) {
			return tse_repo.NewFiskalyTSEClient(fiskalyBaseURL, credentials, nil)
		},
		trigger: tse_repo.SignaturWorkerTrigger(),
		now:     time.Now,
	}
}

// Run blocks until ctx is cancelled.
func (w *tseSignaturWorker) Run(ctx context.Context) {
	defer w.releaseLock()

	interval := w.pollInterval
	if interval <= 0 {
		interval = tseSignaturPollInterval
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.trigger:
			// Immediate trigger after a commit that enqueued a Signaturauftrag.
		case <-ticker.C:
			// Fallback for lost triggers and backoff retries.
		}

		w.tick(ctx)
	}
}

func (w *tseSignaturWorker) tick(ctx context.Context) {
	defer recoverPanic("TSE-Signatur-Worker")

	if !w.ensureLock(ctx) {
		return
	}
	if err := w.processOnce(ctx); err != nil {
		log.Error().Err(err).Msg("TSE-Signatur-Worker Durchlauf fehlgeschlagen")
	}
}

// ensureLock holds the session-scoped advisory lock on a pinned connection, since a dropped connection frees it silently.
// A second instance logs an error and retries each tick instead of failing fast.
func (w *tseSignaturWorker) ensureLock(ctx context.Context) bool {
	if w.lockDB == nil {
		return true
	}

	if w.lockConn != nil {
		if err := w.lockConn.PingContext(ctx); err != nil {
			w.lockConn.Close() //nolint:errcheck,gosec // connection is already broken
			w.lockConn = nil
			w.lockHeld = false
		} else if w.lockHeld {
			return true
		}
	}

	if w.lockConn == nil {
		conn, err := w.lockDB.Conn(ctx)
		if err != nil {
			log.Error().Err(err).Msg("TSE-Signatur-Worker: keine Connection fuer den Advisory Lock")
			return false
		}
		w.lockConn = conn
	}

	if err := w.lockConn.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", tseSignaturWorkerLockKey).Scan(&w.lockHeld); err != nil {
		log.Error().Err(err).Msg("TSE-Signatur-Worker: Advisory Lock nicht pruefbar")
		w.lockConn.Close() //nolint:errcheck,gosec // connection is discarded
		w.lockConn = nil
		w.lockHeld = false
		return false
	}
	if !w.lockHeld {
		log.Error().Msg("TSE-Signatur-Worker: Advisory Lock nicht erhalten — laeuft eine zweite Instanz? Neuer Versuch am naechsten Tick")
	}
	return w.lockHeld
}

// releaseLock frees the session-scoped advisory lock by closing its connection.
func (w *tseSignaturWorker) releaseLock() {
	if w.lockConn != nil {
		w.lockConn.Close() //nolint:errcheck,gosec // Shutdown
		w.lockConn = nil
		w.lockHeld = false
	}
}

func (w *tseSignaturWorker) processOnce(ctx context.Context) error {
	// Outage backoff; the first run after it expires is the half-open probe.
	if w.now().Before(w.stoerungNaechsterVersuch) {
		return nil
	}

	conf, err := w.settingsRepo.GetTSEKonfiguration(ctx)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return w.markiereNichtKonfiguriert(ctx)
		}
		// A read error is transient, so nothing gets marked as not configured.
		return err
	}
	if !conf.IstKonfiguriert() {
		// An empty configuration row means no TSE is set up.
		return w.markiereNichtKonfiguriert(ctx)
	}

	client, err := w.clientFor(conf.Credentials())
	if err != nil {
		log.Warn().Err(err).Msg("TSE-Signatur-Worker could not create TSE client")
		return nil
	}

	// Bookkeeping (failed attempts, Störungsprotokoll) uses the parent ctx so it still writes after the deadline.
	deadline := w.durchlaufDeadline
	if deadline <= 0 {
		deadline = tseSignaturDurchlaufDeadline
	}
	durchlaufCtx, cancel := context.WithTimeout(ctx, deadline)
	defer cancel()

	auftraege, err := w.store.GetOffeneTSESignaturauftraege(durchlaufCtx, tseSignaturBatchSize)
	if err != nil {
		return err
	}

	erfolgVermerkt := false
	for _, auftrag := range auftraege {
		err := w.processAuftrag(durchlaufCtx, client, auftrag)
		if err == nil {
			if !erfolgVermerkt {
				erfolgVermerkt = true
				w.beendeStoerung(ctx)
			}
			continue
		}

		if tse.IstAuftragsFehler(err) {
			// Job-specific error: count the attempt and skip, so a poison job never blocks the queue.
			log.Warn().Err(err).Str("tx_id", auftrag.TxID).Int("auftrag_id", auftrag.ID).Msg("TSE-Signierung fuer Auftrag abgelehnt")
			if err := w.store.TSESignaturauftragFehlversuch(ctx, auftrag.ID, err.Error()); err != nil {
				log.Error().Err(err).Int("auftrag_id", auftrag.ID).Msg("Failed to record TSE-Signatur-Fehlversuch")
			}
			continue
		}

		// TSE-wide error: abort without counting attempts, so a long outage fails no job for good.
		w.beginneStoerung(ctx, err)
		return fmt.Errorf("TSE-weiter Fehler bei Auftrag %d (Fehlerserie %d, naechster Versuch %s): %w",
			auftrag.ID, w.stoerungSerie, w.stoerungNaechsterVersuch.Format(time.RFC3339), err)
	}

	return nil
}

// markiereNichtKonfiguriert opens keine_konfiguration so the enqueue-to-mark window counts as an outage; only the TSE setup closes it.
// See docs/handbuch.md §3.13 (Störungsprotokoll).
func (w *tseSignaturWorker) markiereNichtKonfiguriert(ctx context.Context) error {
	markiert, err := w.store.MarkOffeneAlsNichtKonfiguriert(ctx)
	if err != nil {
		return err
	}
	if markiert == 0 {
		return nil
	}

	log.Warn().Int64("anzahl", markiert).Msg("TSE-Signatur-Worker: offene Auftraege ohne TSE-Konfiguration endgueltig markiert")
	if err := w.store.OpenTSEStoerung(ctx, tse.StoerungGrundKeineKonfiguration, "keine TSE-Konfiguration"); err != nil {
		log.Error().Err(err).Msg("TSE-Stoerungszeitraum keine_konfiguration nicht geoeffnet")
	}
	return nil
}

// beginneStoerung opens the Störungszeitraum, a no-op while one is active.
func (w *tseSignaturWorker) beginneStoerung(ctx context.Context, cause error) {
	w.stoerungSerie++
	w.stoerungNaechsterVersuch = w.now().Add(tseStoerungBackoff(w.stoerungSerie))
	if err := w.store.OpenTSEStoerung(ctx, tse.StoerungGrundTSEFehler, cause.Error()); err != nil {
		log.Error().Err(err).Msg("TSE-Stoerungszeitraum nicht geoeffnet")
	}
}

// beendeStoerung runs on a run's first successful signature and is idempotent, covering a restart with an open period.
func (w *tseSignaturWorker) beendeStoerung(ctx context.Context) {
	w.stoerungSerie = 0
	w.stoerungNaechsterVersuch = time.Time{}
	if err := w.store.CloseTSEStoerung(ctx, tse.StoerungGrundTSEFehler); err != nil {
		log.Error().Err(err).Msg("TSE-Stoerungszeitraum nicht geschlossen")
	}
}

func tseStoerungBackoff(serie int) time.Duration {
	backoff := tseStoerungBackoffBasis
	for i := 1; i < serie && backoff < tseStoerungBackoffDeckel; i++ {
		backoff *= 2
	}
	return min(backoff, tseStoerungBackoffDeckel)
}

func (w *tseSignaturWorker) clientFor(creds tse.Credentials) (tseWorkerClient, error) {
	if w.client != nil && w.clientCreds == creds {
		return w.client, nil
	}

	client, err := w.newTSEClient(creds)
	if err != nil {
		return nil, err
	}
	w.client = client
	w.clientCreds = creds
	return client, nil
}

func (w *tseSignaturWorker) processAuftrag(ctx context.Context, client tseWorkerClient, auftrag tse_repo.OffenerSignaturauftrag) error {
	finishResult, startLogTime, err := w.beschaffeSignatur(ctx, client, auftrag)
	if err != nil {
		return err
	}

	logTimeStart := nonZeroTime(startLogTime, finishResult.LogTimeStart)
	if logTimeStart.IsZero() {
		logTimeStart = w.now().UTC()
	}
	logTimeEnd := nonZeroTime(finishResult.LogTime, finishResult.LogTimeEnd)
	if logTimeEnd.IsZero() {
		logTimeEnd = logTimeStart
	}

	return w.store.QuittiereTSESignaturauftrag(ctx, auftrag.ID, tse.Signatur{
		TransaktionNummer: finishResult.TransactionNumber,
		SignaturZaehler:   finishResult.SignatureCounter,
		TSESeriennummer:   finishResult.SerialNumberTSE,
		LogTimeStart:      logTimeStart,
		LogTimeEnd:        logTimeEnd,
		Signatur:          finishResult.Signature,
		QRCodeData:        finishResult.QRCodeData,
	})
}

// beschaffeSignatur queries fiskaly first: a finished transaction is adopted and an active one only finished.
// This heals the 409 after a crash between signing and acknowledging.
func (w *tseSignaturWorker) beschaffeSignatur(ctx context.Context, client tseWorkerClient, auftrag tse_repo.OffenerSignaturauftrag) (tse.FinishResult, time.Time, error) {
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
		// An unexpected state such as CANCELLED concerns this one transaction only.
		return tse.FinishResult{}, time.Time{}, tse.AuftragsFehler{Err: fmt.Errorf("transaktion %s hat unerwarteten Zustand %q bei fiskaly", auftrag.TxID, vorhanden.State)}
	}
}

func nonZeroTime(primary time.Time, fallback time.Time) time.Time {
	if !primary.IsZero() {
		return primary
	}
	return fallback
}
