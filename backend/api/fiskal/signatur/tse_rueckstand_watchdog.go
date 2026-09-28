package signatur

import (
	"context"
	"database/sql"
	"time"

	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/repository/tse_repo"
	"github.com/rs/zerolog/log"
)

const rueckstandFehlertext = "Signaturaufträge im Rückstand: der älteste offene Auftrag wartet länger als die Rückstands-Schwelle auf die TSE-Signatur"

type rueckstandStore interface {
	GetAeltesterOffenerTSESignaturauftrag(ctx context.Context) (*time.Time, error)
	OpenTSEStoerung(ctx context.Context, grundArt string, fehlertext string) error
	CloseTSEStoerung(ctx context.Context, grundArt string) error
}

// tseRueckstandWatchdog records signing backlogs in the Störungsprotokoll.
// Its own ticker lets it record a hung worker too, independent of reader traffic.
type tseRueckstandWatchdog struct {
	store rueckstandStore
	// tickInterval 0 falls back to tse.WatchdogTickIntervall.
	tickInterval time.Duration
	now          func() time.Time
}

func NewTSERueckstandWatchdog(database *sql.DB) Runner {
	return &tseRueckstandWatchdog{
		store: tse_repo.NewRepository(database),
		now:   time.Now,
	}
}

// Run blocks until ctx is cancelled.
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

func (w *tseRueckstandWatchdog) tick(ctx context.Context) {
	defer recoverPanic("TSE-Rückstands-Watchdog")

	if err := w.checkRueckstand(ctx); err != nil {
		log.Error().Err(err).Msg("TSE-Rückstands-Watchdog Durchlauf fehlgeschlagen")
	}
}

// checkRueckstand is idempotent and closes only rueckstand periods.
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
