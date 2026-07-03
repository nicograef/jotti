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

func (c Command) signBestellungUmgebuchtEvent(ctx context.Context, evt event.Event, positionen []kasse.Position, faktor int) (tseApp.Signierung, error) {
	// Eine Umbuchung ist geldneutral (keine :Bar-Zahlungszeile). Der Abgang vom
	// Quelltisch wird mit negativen Mengen signiert (faktor -1), der Zugang auf dem
	// Zieltisch mit positiven — sonst erschiene die Ware TSE-seitig doppelt bestellt.
	processData, err := tseApp.BuildBestellungProcessDataWithFaktor(positionen, faktor)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("Failed to build TSE process_data for umbuchung")
		return tseApp.Signierung{}, ErrDatabase
	}

	return c.TSESignierer.SignEvent(ctx, evt, tse.ProcessTypeBestellungV1, processData, kasse.EmbedTSEInBestellungUmgebucht)
}

func (c Command) signStornierungErteiltEvent(ctx context.Context, evt event.Event, positionen []kasse.Position, stornoBetragCents int) (tseApp.Signierung, error) {
	processData, err := tseApp.BuildKassenbelegProcessDataWithFaktor(positionen, -stornoBetragCents, -1)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("Failed to build TSE process_data for stornierung")
		return tseApp.Signierung{}, ErrDatabase
	}

	return c.TSESignierer.SignEvent(ctx, evt, tse.ProcessTypeKassenbelegV1, processData, kasse.EmbedTSEInStornierungErteilt)
}

func (c Command) signBestellungKorrigiertEvent(ctx context.Context, evt event.Event, positionen []kasse.Position) (tseApp.Signierung, error) {
	// Die geldneutrale Korrektur ist kassenneutral (keine :Bar-Zahlungszeile) und
	// nimmt Positionen zurück: negative Mengen (Anhang I), damit die Rücknahme
	// TSE-seitig von einer Neubestellung unterscheidbar ist.
	processData, err := tseApp.BuildBestellungProcessDataWithFaktor(positionen, -1)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("Failed to build TSE process_data for korrektur")
		return tseApp.Signierung{}, ErrDatabase
	}

	return c.TSESignierer.SignEvent(ctx, evt, tse.ProcessTypeBestellungV1, processData, kasse.EmbedTSEInBestellungKorrigiert)
}
