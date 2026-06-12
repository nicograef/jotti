package application

import (
	"context"

	tseApp "github.com/nicograef/jotti/backend/api/tse/application"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/rs/zerolog"
)

func (c Command) signDirektverkaufGetaetigtEvent(ctx context.Context, evt event.Event, positionen []kasse.Position) (tseApp.Signierung, error) {
	processData, err := tseApp.BuildKassenbelegProcessData(positionen, sumPositionenCents(positionen))
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("Failed to build TSE process_data for direktverkauf")
		return tseApp.Signierung{}, ErrDatabase
	}

	return c.TSESignierer.SignEvent(ctx, evt, tse.ProcessTypeKassenbelegV1, processData, kasse.EmbedTSEInDirektverkaufGetaetigt)
}

func (c Command) signDirektverkaufStorniertEvent(ctx context.Context, evt event.Event, positionen []kasse.Position, stornoBetragCents int) (tseApp.Signierung, error) {
	processData, err := tseApp.BuildKassenbelegProcessDataWithFaktor(positionen, -stornoBetragCents, -1)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("Failed to build TSE process_data for direktverkauf storno")
		return tseApp.Signierung{}, ErrDatabase
	}

	return c.TSESignierer.SignEvent(ctx, evt, tse.ProcessTypeKassenbelegV1, processData, kasse.EmbedTSEInDirektverkaufStorniert)
}

func sumPositionenCents(positionen []kasse.Position) int {
	sum := 0
	for _, pos := range positionen {
		sum += pos.Einzelpreis * pos.Menge
	}
	return sum
}
