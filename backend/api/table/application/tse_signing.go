package application

import (
	"context"

	tseApp "github.com/nicograef/jotti/backend/api/tse/application"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/rs/zerolog"
)

func (c Command) signZahlungKassiertEvent(ctx context.Context, evt event.Event, positionen []kasse.Position, zahlbetragCents int) (tseApp.Signierung, error) {
	processData, err := tseApp.BuildKassenbelegProcessData(positionen, zahlbetragCents)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("Failed to build TSE process_data for zahlung")
		return tseApp.Signierung{}, ErrDatabase
	}

	return c.TSESignierer.SignEvent(ctx, evt, tse.ProcessTypeKassenbelegV1, processData, kasse.EmbedTSEInZahlungKassiert)
}

func (c Command) signBestellungAufgenommenEvent(ctx context.Context, evt event.Event, positionen []kasse.Position) (tseApp.Signierung, error) {
	processData, err := tseApp.BuildBestellungProcessData(positionen)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("Failed to build TSE process_data for bestellung")
		return tseApp.Signierung{}, ErrDatabase
	}

	return c.TSESignierer.SignEvent(ctx, evt, tse.ProcessTypeBestellungV1, processData, kasse.EmbedTSEInBestellungAufgenommen)
}

func (c Command) signStornierungErteiltEvent(ctx context.Context, evt event.Event, positionen []kasse.Position, stornoBetragCents int) (tseApp.Signierung, error) {
	processData, err := tseApp.BuildKassenbelegProcessDataWithFaktor(positionen, -stornoBetragCents, -1)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("Failed to build TSE process_data for stornierung")
		return tseApp.Signierung{}, ErrDatabase
	}

	return c.TSESignierer.SignEvent(ctx, evt, tse.ProcessTypeKassenbelegV1, processData, kasse.EmbedTSEInStornierungErteilt)
}

func (c Command) signAuszahlungGeleistetEvent(ctx context.Context, evt event.Event, betragCents int) (tseApp.Signierung, error) {
	// Eine Auszahlung ist ein Eigenbeleg ohne Umsatz (AEAO 2.2.3.6.1): alle
	// Steuerbetraege 0.00, nur der negative Zahlbetrag ist gefuellt — wie der
	// Geldtransit.
	processData := tseApp.BuildEigenbelegProcessData(-betragCents)
	return c.TSESignierer.SignEvent(ctx, evt, tse.ProcessTypeKassenbelegV1, processData, kasse.EmbedTSEInAuszahlungGeleistet)
}
