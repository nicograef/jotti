package application

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/betreiber"
	"github.com/rs/zerolog"
)

type betreiberCommandRepo interface {
	UpsertBetreiber(ctx context.Context, b betreiber.Betreiber) error
	SetElsterGemeldetAm(ctx context.Context) error
	ClearElsterGemeldetAm(ctx context.Context) error
}

type Command struct {
	BetreiberRepo betreiberCommandRepo
}

func (c Command) UpdateBetreiber(ctx context.Context, b betreiber.Betreiber) error {
	log := zerolog.Ctx(ctx)

	if err := c.BetreiberRepo.UpsertBetreiber(ctx, b); err != nil {
		log.Error().Err(err).Msg("Failed to save betreiber")
		return ErrDatabase
	}
	log.Info().Str("vereinsname", b.Vereinsname).Msg("Betreiber saved")
	return nil
}

// SetzeElsterMeldung markiert die ELSTER-Kassenmeldung als erledigt (serverseitig
// auf das aktuelle Datum, § 146a Abs. 4 AO).
func (c Command) SetzeElsterMeldung(ctx context.Context) error {
	log := zerolog.Ctx(ctx)

	if err := c.BetreiberRepo.SetElsterGemeldetAm(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to set elster meldung")
		return ErrDatabase
	}
	log.Info().Msg("Elster meldung marked as done")
	return nil
}

// NimmElsterMeldungZurueck setzt die ELSTER-Kassenmeldung auf NULL zurück, damit
// ein Fehlklick korrigierbar bleibt.
func (c Command) NimmElsterMeldungZurueck(ctx context.Context) error {
	log := zerolog.Ctx(ctx)

	if err := c.BetreiberRepo.ClearElsterGemeldetAm(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to clear elster meldung")
		return ErrDatabase
	}
	log.Info().Msg("Elster meldung reset")
	return nil
}
