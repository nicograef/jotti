package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/nicograef/jotti/backend/config"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/repository/tse_repo"
	"github.com/rs/zerolog/log"
)

const (
	// tseSignaturPollInterval ist der Polling-Fallback des Signatur-Workers:
	// Der Sofort-Trigger nach jedem Commit ist der Regelweg, der Tick faengt
	// verlorene Trigger (z. B. nach einem Crash zwischen Commit und Trigger)
	// und stellt Backoff-Wiedervorlagen zu.
	tseSignaturPollInterval = 5 * time.Second
	tseSignaturBatchSize    = 20
	// tseSignaturDurchlaufDeadline begrenzt jeden Durchlauf: Ein haengender
	// Durchlauf wuerde den seriellen Worker sonst unbegrenzt blockieren. Ein
	// Deadline-Abbruch gilt als TSE-weiter Fehler.
	tseSignaturDurchlaufDeadline = 2 * time.Minute
	// tseStoerungBackoffBasis/-Deckel spannen den Backoff des Worker-
	// Stoerungszustands nach TSE-weiten Fehlern auf: 5 s, verdoppelt je
	// Fehlerserie bis zum Deckel von 2 Minuten — die Erholung wird binnen
	// Minuten erkannt, fiskaly waehrend der Stoerung nicht mit dem Rueckstand
	// bombardiert. Bewusst ohne Jitter: Ein einzelner serieller Worker hat
	// nichts zu desynchronisieren, die Tests bleiben deterministisch.
	tseStoerungBackoffBasis  = 5 * time.Second
	tseStoerungBackoffDeckel = 2 * time.Minute
	// tseSignaturWorkerLockKey ist der frei gewaehlte Schluessel des Postgres
	// Advisory Locks, der die Single-Prozess-Annahme absichert: Nur der
	// Lock-Halter spricht mit der TSE.
	tseSignaturWorkerLockKey = 823914502
)

type tseSettingsReader interface {
	GetTSEKonfiguration(ctx context.Context) (tse.Konfiguration, error)
}

type tseSignaturStore interface {
	GetOffeneTSESignaturauftraege(ctx context.Context, limit int) ([]tse_repo.OffenerSignaturauftrag, error)
	QuittiereTSESignaturauftrag(ctx context.Context, auftragID int, signatur tse.Signatur) error
	TSESignaturauftragFehlversuch(ctx context.Context, auftragID int, fehler string) error
	MarkiereOffeneAlsNichtKonfiguriert(ctx context.Context) (int64, error)
	OeffneTSEStoerung(ctx context.Context, grundArt string, fehlertext string) error
	SchliesseTSEStoerung(ctx context.Context, grundArt string) error
}

// tseWorkerClient beschreibt, was der Signatur-Worker von der TSE braucht:
// signieren und den Ist-Zustand einer Transaktion abfragen.
type tseWorkerClient interface {
	tse.TSEClient
	tse.TransactionRetriever
}

type tseClientFactory func(credentials tse.Credentials) (tseWorkerClient, error)

// tseSignaturWorker ist der einzige Sprecher fuer TSE-Signaturtransaktionen:
// Er arbeitet die Signaturauftraege FIFO ab, heilt per Ist-Abfrage und
// quittiert die Signatur mit einem einzelnen Update am Auftrag.
type tseSignaturWorker struct {
	// lockDB liefert die dedizierte, fuer die Worker-Lebenszeit gepinnte
	// Connection des Advisory Locks; nil (Unit-Tests) ueberspringt den Lock.
	lockDB       *sql.DB
	settingsRepo tseSettingsReader
	store        tseSignaturStore
	newTSEClient tseClientFactory
	trigger      <-chan struct{}
	// pollInterval ist der Polling-Fallback-Takt; 0 (Zero Value in Tests)
	// faellt auf tseSignaturPollInterval zurueck.
	pollInterval time.Duration
	// durchlaufDeadline begrenzt einen Durchlauf; 0 (Zero Value in Tests)
	// faellt auf tseSignaturDurchlaufDeadline zurueck.
	durchlaufDeadline time.Duration
	now               func() time.Time

	lockConn *sql.Conn
	lockHeld bool

	// Stoerungszustand nach einem TSE-weiten Fehler — kein Zustandsautomat,
	// nur zwei Werte: Bis stoerungNaechsterVersuch laesst der Worker fiskaly
	// in Ruhe, stoerungSerie zaehlt die Fehlerserie fuer den wachsenden
	// Backoff. Die Half-Open-Probe ist schlicht der erste Auftrag des
	// naechsten Durchlaufs: Scheitert er TSE-weit, bricht der Durchlauf
	// erneut ab und der Backoff waechst; gelingt er, laeuft die volle
	// Aufarbeitung und die erste erfolgreiche Signatur beendet die Stoerung.
	stoerungNaechsterVersuch time.Time
	stoerungSerie            int

	// client wird ueber Durchlaeufe hinweg wiederverwendet (samt Auth-Token)
	// und nur bei geaenderten Zugangsdaten neu gebaut.
	client      tseWorkerClient
	clientCreds tse.Credentials
}

func newTSESignaturWorker(cfg config.Config, database *sql.DB) *tseSignaturWorker {
	return &tseSignaturWorker{
		lockDB:       database,
		settingsRepo: tse_repo.NewRepository(database),
		store:        tse_repo.NewRepository(database),
		newTSEClient: func(credentials tse.Credentials) (tseWorkerClient, error) {
			return tse_repo.NewFiskalyTSEClient(cfg.FiskalyBaseURL, credentials, nil)
		},
		trigger: tse_repo.SignaturWorkerTrigger(),
		now:     time.Now,
	}
}

func (w *tseSignaturWorker) run(ctx context.Context) {
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
			// Sofort-Trigger nach einem Commit mit neuem Signaturauftrag.
		case <-ticker.C:
			// Polling-Fallback fuer verlorene Trigger und Backoff-Wiedervorlagen.
		}

		if !w.ensureLock(ctx) {
			continue
		}
		if err := w.processOnce(ctx); err != nil {
			log.Error().Err(err).Msg("TSE-Signatur-Worker Durchlauf fehlgeschlagen")
		}
	}
}

// ensureLock haelt den session-gebundenen Advisory Lock auf einer dedizierten,
// fuer die Worker-Lebenszeit gepinnten Connection (nicht auf dem Pool). Ein
// Verbindungsabriss gibt den Lock still frei; danach wird er auf einer
// frischen Connection neu erworben. Bekommt eine zweite Instanz den Lock
// nicht, laeuft die App weiter und der Worker versucht es am naechsten Tick
// erneut — mit deutlicher Error-Log-Warnung, kein Fail-Fast.
func (w *tseSignaturWorker) ensureLock(ctx context.Context) bool {
	if w.lockDB == nil {
		return true
	}

	if w.lockConn != nil {
		if err := w.lockConn.PingContext(ctx); err == nil {
			if w.lockHeld {
				return true
			}
		} else {
			w.lockConn.Close() //nolint:errcheck,gosec // Connection ist bereits abgerissen
			w.lockConn = nil
			w.lockHeld = false
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
		w.lockConn.Close() //nolint:errcheck,gosec // Connection wird verworfen
		w.lockConn = nil
		w.lockHeld = false
		return false
	}
	if !w.lockHeld {
		log.Error().Msg("TSE-Signatur-Worker: Advisory Lock nicht erhalten — laeuft eine zweite Instanz? Neuer Versuch am naechsten Tick")
	}
	return w.lockHeld
}

// releaseLock schliesst die gepinnte Lock-Connection; der session-gebundene
// Advisory Lock wird damit freigegeben.
func (w *tseSignaturWorker) releaseLock() {
	if w.lockConn != nil {
		w.lockConn.Close() //nolint:errcheck,gosec // Shutdown
		w.lockConn = nil
		w.lockHeld = false
	}
}

func (w *tseSignaturWorker) processOnce(ctx context.Context) error {
	// Stoerungszustand: Bis zum naechsten Versuch laesst der Worker fiskaly
	// in Ruhe, statt es mit dem Rueckstand zu bombardieren. Trigger und Ticks
	// laufen weiter; der erste Durchlauf nach Ablauf ist die Half-Open-Probe.
	if w.now().Before(w.stoerungNaechsterVersuch) {
		return nil
	}

	conf, err := w.settingsRepo.GetTSEKonfiguration(ctx)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			// Fehlende Konfiguration: offene Auftraege endgueltig markieren.
			return w.markiereNichtKonfiguriert(ctx)
		}
		// Nicht lesbare Konfiguration (echter DB-Fehler): nichts markieren, es
		// koennte gleich wieder gehen — ein Lesefehler ist kein Dauerzustand.
		return err
	}
	if !conf.IstKonfiguriert() {
		// Vorhandene, aber leere Konfiguration = keine TSE eingerichtet.
		return w.markiereNichtKonfiguriert(ctx)
	}

	client, err := w.clientFuer(conf.Credentials())
	if err != nil {
		log.Warn().Err(err).Msg("TSE-Signatur-Worker could not create TSE client")
		return nil
	}

	// Jeder Durchlauf hat eine Deadline; die Buchhaltung (Fehlversuch,
	// Stoerungsprotokoll) laeuft auf dem Eltern-Kontext, damit sie auch nach
	// aufgebrauchtem Budget noch schreiben kann.
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
			// Auftragsspezifischer Fehler: Fehlversuch am Auftrag verbuchen
			// und den Auftrag ueberspringen — ein Gift-Auftrag staut nie die
			// Queue und schlaegt nach MaxSignaturVersuche endgueltig fehl.
			log.Warn().Err(err).Str("tx_id", auftrag.TxID).Int("auftrag_id", auftrag.ID).Msg("TSE-Signierung fuer Auftrag abgelehnt")
			if err := w.store.TSESignaturauftragFehlversuch(ctx, auftrag.ID, err.Error()); err != nil {
				log.Error().Err(err).Int("auftrag_id", auftrag.ID).Msg("Failed to record TSE-Signatur-Fehlversuch")
			}
			continue
		}

		// TSE-weiter Fehler: Durchlauf abbrechen, ohne Fehlversuche an den
		// Auftraegen — ein mehrstuendiger Ausfall laesst keine Auftraege
		// endgueltig fehlschlagen.
		w.beginneStoerung(ctx, err)
		return fmt.Errorf("TSE-weiter Fehler bei Auftrag %d (Fehlerserie %d, naechster Versuch %s): %w",
			auftrag.ID, w.stoerungSerie, w.stoerungNaechsterVersuch.Format(time.RFC3339), err)
	}

	return nil
}

// markiereNichtKonfiguriert markiert alle offenen Auftraege endgueltig als
// tse_nicht_konfiguriert, solange keine TSE konfiguriert ist. Der Dauerzustand
// ohne Konfiguration ist die dritte Stoerungsquelle: Sind Auftraege betroffen,
// oeffnet der Worker den keine_konfiguration-Zeitraum (No-Op, solange bereits
// ein Zeitraum aktiv ist), damit auch das kurze Fenster zwischen Einreihen und
// Markieren als Ausfall belegt ist. Der Zeitraum endet erst mit der Einrichtung.
func (w *tseSignaturWorker) markiereNichtKonfiguriert(ctx context.Context) error {
	markiert, err := w.store.MarkiereOffeneAlsNichtKonfiguriert(ctx)
	if err != nil {
		return err
	}
	if markiert == 0 {
		return nil
	}

	log.Warn().Int64("anzahl", markiert).Msg("TSE-Signatur-Worker: offene Auftraege ohne TSE-Konfiguration endgueltig markiert")
	if err := w.store.OeffneTSEStoerung(ctx, tse.StoerungGrundKeineKonfiguration, "keine TSE-Konfiguration"); err != nil {
		log.Error().Err(err).Msg("TSE-Stoerungszeitraum keine_konfiguration nicht geoeffnet")
	}
	return nil
}

// beginneStoerung betritt den Stoerungszustand nach einem TSE-weiten Fehler:
// Backoff fuer den naechsten Versuch setzen und den Stoerungszeitraum im
// Stoerungsprotokoll oeffnen (No-Op, solange bereits ein Zeitraum aktiv ist).
func (w *tseSignaturWorker) beginneStoerung(ctx context.Context, cause error) {
	w.stoerungSerie++
	w.stoerungNaechsterVersuch = w.now().Add(tseStoerungBackoff(w.stoerungSerie))
	if err := w.store.OeffneTSEStoerung(ctx, tse.StoerungGrundTSEFehler, cause.Error()); err != nil {
		log.Error().Err(err).Msg("TSE-Stoerungszeitraum nicht geoeffnet")
	}
}

// beendeStoerung verlaesst den Stoerungszustand: Die erste erfolgreiche
// Signatur eines Durchlaufs schliesst den TSE-Fehler-Stoerungszeitraum
// (idempotent, auch nach einem Worker-Neustart mit offenem Zeitraum) und
// setzt die Fehlerserie zurueck.
func (w *tseSignaturWorker) beendeStoerung(ctx context.Context) {
	w.stoerungSerie = 0
	w.stoerungNaechsterVersuch = time.Time{}
	if err := w.store.SchliesseTSEStoerung(ctx, tse.StoerungGrundTSEFehler); err != nil {
		log.Error().Err(err).Msg("TSE-Stoerungszeitraum nicht geschlossen")
	}
}

// tseStoerungBackoff liefert die Wartezeit des Stoerungszustands fuer die
// n-te TSE-weite Fehlerserie: Basis verdoppelt je Serie, gedeckelt —
// deterministisch, ohne Jitter.
func tseStoerungBackoff(serie int) time.Duration {
	backoff := tseStoerungBackoffBasis
	for i := 1; i < serie && backoff < tseStoerungBackoffDeckel; i++ {
		backoff *= 2
	}
	return min(backoff, tseStoerungBackoffDeckel)
}

// clientFuer liefert den ueber Durchlaeufe hinweg wiederverwendeten TSE-Client
// (samt Auth-Token); neu gebaut wird nur bei geaenderten Zugangsdaten.
func (w *tseSignaturWorker) clientFuer(creds tse.Credentials) (tseWorkerClient, error) {
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

// beschaffeSignatur liefert die Signaturdaten fuer den Auftrag. Vor einem
// neuen Signierversuch wird der Ist-Zustand bei fiskaly abgefragt: Eine dort
// bereits abgeschlossene Transaktion wird direkt uebernommen statt erneut
// signiert (heilt das 409-Szenario nach Abbruch zwischen Signierung und
// Quittierung), eine noch aktive Transaktion wird nur noch abgeschlossen.
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
		// Ein unerwarteter Zustand (etwa CANCELLED) haengt an dieser einen
		// Transaktion — auftragsspezifisch, kein Grund fuer einen
		// Durchlauf-Abbruch.
		return tse.FinishResult{}, time.Time{}, tse.AuftragsFehler{Err: fmt.Errorf("transaktion %s hat unerwarteten Zustand %q bei fiskaly", auftrag.TxID, vorhanden.State)}
	}
}

func nonZeroTime(primary time.Time, fallback time.Time) time.Time {
	if !primary.IsZero() {
		return primary
	}
	return fallback
}
