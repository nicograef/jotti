package signatur

import (
	"context"
	"database/sql"
	"time"

	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/repository/tse_repo"
	"github.com/rs/zerolog/log"
)

// rueckstandFehlertext beschreibt den Rueckstands-Stoerungszeitraum im
// Stoerungsprotokoll.
const rueckstandFehlertext = "Signaturaufträge im Rückstand: der älteste offene Auftrag wartet länger als die Rückstands-Schwelle auf die TSE-Signatur"

type rueckstandStore interface {
	GetAeltesterOffenerTSESignaturauftrag(ctx context.Context) (*time.Time, error)
	OeffneTSEStoerung(ctx context.Context, grundArt string, fehlertext string) error
	SchliesseTSEStoerung(ctx context.Context, grundArt string) error
}

// tseRueckstandWatchdog dokumentiert Signatur-Rueckstaende im
// Stoerungsprotokoll: Er prueft im Tick-Intervall das Alter des aeltesten
// offenen Signaturauftrags, oeffnet ab der Rueckstands-Schwelle einen
// Rueckstands-Zeitraum und schliesst ihn beim Unterschreiten. Als eigener
// Ticker neben dem Signatur-Worker dokumentiert er auch einen haengenden
// Worker und haengt nicht am Leser-Traffic.
type tseRueckstandWatchdog struct {
	store rueckstandStore
	// tickInterval ist der Pruef-Takt; 0 (Zero Value in Tests) faellt auf
	// tse.WatchdogTickIntervall zurueck.
	tickInterval time.Duration
	now          func() time.Time
}

// NewTSERueckstandWatchdog erstellt den Rueckstands-Watchdog.
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

		if err := w.pruefeRueckstand(ctx); err != nil {
			log.Error().Err(err).Msg("TSE-Rückstands-Watchdog Durchlauf fehlgeschlagen")
		}
	}
}

// pruefeRueckstand oeffnet den Rueckstands-Zeitraum, sobald der aelteste
// offene Auftrag die Rueckstands-Schwelle erreicht, und schliesst ihn, sobald
// der Rueckstand abgebaut ist. Beide Schritte sind idempotent; der Watchdog
// schliesst nur Zeitraeume seiner Grund-Art.
func (w *tseRueckstandWatchdog) pruefeRueckstand(ctx context.Context) error {
	aeltester, err := w.store.GetAeltesterOffenerTSESignaturauftrag(ctx)
	if err != nil {
		return err
	}

	if aeltester != nil && w.now().Sub(*aeltester) >= tse.RueckstandSchwelle {
		return w.store.OeffneTSEStoerung(ctx, tse.StoerungGrundRueckstand, rueckstandFehlertext)
	}
	return w.store.SchliesseTSEStoerung(ctx, tse.StoerungGrundRueckstand)
}
