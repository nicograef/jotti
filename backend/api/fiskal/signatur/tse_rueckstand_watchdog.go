package signatur

import (
	"context"
	"database/sql"
	"time"

	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/repository/tse_repo"
	"github.com/rs/zerolog/log"
)

// rueckstandFehlertext beschreibt den Rückstands-Störungszeitraum im
// Störungsprotokoll.
const rueckstandFehlertext = "Signaturaufträge im Rückstand: der älteste offene Auftrag wartet länger als die Rückstands-Schwelle auf die TSE-Signatur"

type rueckstandStore interface {
	GetAeltesterOffenerTSESignaturauftrag(ctx context.Context) (*time.Time, error)
	OpenTSEStoerung(ctx context.Context, grundArt string, fehlertext string) error
	CloseTSEStoerung(ctx context.Context, grundArt string) error
}

// tseRueckstandWatchdog dokumentiert Signatur-Rückstände im
// Störungsprotokoll: Er prüft im Tick-Intervall das Alter des ältesten
// offenen Signaturauftrags, öffnet ab der Rückstands-Schwelle einen
// Rückstands-Zeitraum und schließt ihn beim Unterschreiten. Als eigener
// Ticker neben dem Signatur-Worker dokumentiert er auch einen hängenden
// Worker und hängt nicht am Leser-Traffic.
type tseRueckstandWatchdog struct {
	store rueckstandStore
	// tickInterval ist der Prüf-Takt; 0 (Zero Value in Tests) fällt auf
	// tse.WatchdogTickIntervall zurück.
	tickInterval time.Duration
	now          func() time.Time
}

// NewTSERueckstandWatchdog erstellt den Rückstands-Watchdog.
func NewTSERueckstandWatchdog(database *sql.DB) Runner {
	return &tseRueckstandWatchdog{
		store: tse_repo.NewRepository(database),
		now:   time.Now,
	}
}

// Run startet den Watchdog und blockiert bis ctx abgebrochen wird.
func (w *tseRueckstandWatchdog) Run(ctx context.Context) {
	interval := w.tickInterval
	if interval <= 0 {
		interval = tse.WatchdogTickIntervall
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		w.tick(ctx)
	}
}

// tick führt eine Loop-Iteration aus. Ein Panic wird abgefangen und geloggt
// statt den Run-Loop zu beenden — die Überwachung läuft am nächsten Tick
// weiter.
func (w *tseRueckstandWatchdog) tick(ctx context.Context) {
	defer recoverPanic("TSE-Rückstands-Watchdog")

	if err := w.checkRueckstand(ctx); err != nil {
		log.Error().Err(err).Msg("TSE-Rückstands-Watchdog Durchlauf fehlgeschlagen")
	}
}

// checkRueckstand öffnet den Rückstands-Zeitraum, sobald der älteste
// offene Auftrag die Rückstands-Schwelle erreicht, und schließt ihn, sobald
// der Rückstand abgebaut ist. Beide Schritte sind idempotent; der Watchdog
// schließt nur Zeiträume seiner Grund-Art.
func (w *tseRueckstandWatchdog) checkRueckstand(ctx context.Context) error {
	aeltester, err := w.store.GetAeltesterOffenerTSESignaturauftrag(ctx)
	if err != nil {
		return err
	}

	if aeltester != nil && w.now().Sub(*aeltester) >= tse.RueckstandSchwelle {
		return w.store.OpenTSEStoerung(ctx, tse.StoerungGrundRueckstand, rueckstandFehlertext)
	}
	return w.store.CloseTSEStoerung(ctx, tse.StoerungGrundRueckstand)
}
