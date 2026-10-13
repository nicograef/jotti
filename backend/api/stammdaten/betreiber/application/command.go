package application

import (
	"context"
	"time"

	"github.com/nicograef/jotti/backend/domain/betreiber"
	"github.com/nicograef/jotti/backend/internal/logging"
	"github.com/nicograef/jotti/backend/internal/zeit"
)

type betreiberCommandRepo interface {
	UpsertBetreiber(ctx context.Context, b betreiber.Betreiber) error
	SetElsterGemeldetAm(ctx context.Context, gemeldetAm time.Time) error
	ClearElsterGemeldetAm(ctx context.Context) error
}

type Command struct {
	BetreiberRepo betreiberCommandRepo
	// now ist die Uhr des Meldedatums: nil steht für time.Now (Produktivpfad),
	// Tests setzen sie.
	now func() time.Time
}

func (c Command) UpdateBetreiber(ctx context.Context, b betreiber.Betreiber) error {
	log := logging.FromContext(ctx)

	if err := c.BetreiberRepo.UpsertBetreiber(ctx, b); err != nil {
		log.Error("Failed to save betreiber", "error", err)
		return ErrDatabase
	}
	log.Info("Betreiber saved", "vereinsname", b.Vereinsname)
	return nil
}

// meldedatum liefert den Berliner Kalendertag des Zeitpunkts als Mitternacht UTC,
// passend zur DATE-Spalte.
func meldedatum(zeitpunkt time.Time) time.Time {
	jahr, monat, tag := zeitpunkt.In(zeit.Berlin).Date()
	return time.Date(jahr, monat, tag, 0, 0, 0, 0, time.UTC)
}

// SetzeElsterMeldung markiert die ELSTER-Kassenmeldung als erledigt (serverseitig
// auf das aktuelle Datum, § 146a Abs. 4 AO). Das Datum ist der Berliner
// Kalendertag: die Container laufen in UTC und lägen abends einen Tag zurück.
func (c Command) SetzeElsterMeldung(ctx context.Context) error {
	log := logging.FromContext(ctx)

	now := c.now
	if now == nil {
		now = time.Now
	}

	gemeldetAm := meldedatum(now())
	if err := c.BetreiberRepo.SetElsterGemeldetAm(ctx, gemeldetAm); err != nil {
		log.Error("Failed to set elster meldung", "error", err)
		return ErrDatabase
	}
	log.Info("Elster meldung marked as done", "gemeldet_am", gemeldetAm.Format("2006-01-02"))
	return nil
}

// NimmElsterMeldungZurueck setzt die ELSTER-Kassenmeldung auf NULL zurück, damit
// ein Fehlklick korrigierbar bleibt.
func (c Command) NimmElsterMeldungZurueck(ctx context.Context) error {
	log := logging.FromContext(ctx)

	if err := c.BetreiberRepo.ClearElsterGemeldetAm(ctx); err != nil {
		log.Error("Failed to clear elster meldung", "error", err)
		return ErrDatabase
	}
	log.Info("Elster meldung reset")
	return nil
}
