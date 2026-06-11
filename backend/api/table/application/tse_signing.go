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

	return c.TSESignierer.SignEvent(ctx, evt, tse.ProcessTypeKassenbelegV1, processData, withZahlungEventTSE)
}

func (c Command) signBestellungAufgenommenEvent(ctx context.Context, evt event.Event, positionen []kasse.Position) (tseApp.Signierung, error) {
	processData, err := tseApp.BuildBestellungProcessData(positionen)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("Failed to build TSE process_data for bestellung")
		return tseApp.Signierung{}, ErrDatabase
	}

	return c.TSESignierer.SignEvent(ctx, evt, tse.ProcessTypeBestellungV1, processData, withBestellungEventTSE)
}

func (c Command) signStornierungErteiltEvent(ctx context.Context, evt event.Event, positionen []kasse.Position, stornoBetragCents int) (tseApp.Signierung, error) {
	processData, err := tseApp.BuildKassenbelegProcessDataWithFaktor(positionen, -stornoBetragCents, -1)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("Failed to build TSE process_data for stornierung")
		return tseApp.Signierung{}, ErrDatabase
	}

	return c.TSESignierer.SignEvent(ctx, evt, tse.ProcessTypeKassenbelegV1, processData, withStornierungEventTSE)
}

func (c Command) signAuszahlungGeleistetEvent(ctx context.Context, evt event.Event, betragCents int) (tseApp.Signierung, error) {
	processData, err := tseApp.BuildKassenbelegProcessData(nil, -betragCents)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("Failed to build TSE process_data for auszahlung")
		return tseApp.Signierung{}, ErrDatabase
	}

	return c.TSESignierer.SignEvent(ctx, evt, tse.ProcessTypeKassenbelegV1, processData, withAuszahlungEventTSE)
}

var withZahlungEventTSE = tseApp.EmbedTSEInData(kasse.EventTypeZahlungKassiertV1, func(data *zahlungKassiertV1Data, txID string, tseData *kasse.TSEData) {
	data.TSETxID = txID
	data.TSEData = tseData
	data.TSEAusfall = tseData == nil
})

var withBestellungEventTSE = tseApp.EmbedTSEInData(kasse.EventTypeBestellungAufgenommenV1, func(data *bestellungAufgenommenV1Data, txID string, tseData *kasse.TSEData) {
	data.TSETxID = txID
	data.TSEData = tseData
})

var withStornierungEventTSE = tseApp.EmbedTSEInData(kasse.EventTypeStornierungErteiltV1, func(data *stornierungErteiltV1Data, txID string, tseData *kasse.TSEData) {
	data.TSETxID = txID
	data.TSEData = tseData
})

var withAuszahlungEventTSE = tseApp.EmbedTSEInData(kasse.EventTypeAuszahlungGeleistetV1, func(data *auszahlungGeleistetV1Data, txID string, tseData *kasse.TSEData) {
	data.TSETxID = txID
	data.TSEData = tseData
})
