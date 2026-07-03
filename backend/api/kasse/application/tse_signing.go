package application

import (
	"context"
	"time"

	tseApp "github.com/nicograef/jotti/backend/api/tse/application"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/rs/zerolog"
)

// signKassensitzungEroeffnetEvent signiert den Anfangsbestand als Eigenbeleg
// (Bareinlage zu Sitzungsbeginn, AEAO 2.2.3.6.1) — wie Geldtransit und
// Kassendifferenz ist er ein abzusichernder Geschäftsvorfall.
func (c Command) signKassensitzungEroeffnetEvent(ctx context.Context, evt event.Event, betragCents int) (tseApp.Signierung, error) {
	processData := tseApp.BuildEigenbelegProcessData(betragCents)
	return c.TSESignierer.SignEvent(ctx, evt, tse.ProcessTypeKassenbelegV1, processData, kasse.EmbedTSEInKassensitzungEroeffnet)
}

func (c Command) signGeldtransitGebuchtEvent(ctx context.Context, evt event.Event, richtung string, betragCents int) (tseApp.Signierung, error) {
	processData, err := tseApp.BuildGeldtransitProcessData(richtung, betragCents)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("Failed to build TSE process_data for geldtransit")
		return tseApp.Signierung{}, ErrDatabase
	}

	return c.TSESignierer.SignEvent(ctx, evt, tse.ProcessTypeKassenbelegV1, processData, kasse.EmbedTSEInGeldtransitGebucht)
}

func (c Command) signDifferenzSollIstGebuchtEvent(ctx context.Context, evt event.Event, differenzCents int) (tseApp.Signierung, error) {
	// differenzCents = Soll − Ist. Die tatsächliche Bargeldbewegung ist Ist − Soll:
	// ein Fehlbetrag (Soll > Ist) mindert den Bestand, ein Überschuss mehrt ihn.
	processData := tseApp.BuildEigenbelegProcessData(-differenzCents)
	return c.TSESignierer.SignEvent(ctx, evt, tse.ProcessTypeKassenbelegV1, processData, kasse.EmbedTSEInDifferenzSollIstGebucht)
}

func (c Command) signTagesabschlussErstelltEvent(ctx context.Context, evt event.Event, zNr int, zeitraumVon time.Time, zeitraumBis time.Time) (tseApp.Signierung, error) {
	processData := tseApp.BuildTagesabschlussProcessData(zNr, zeitraumVon, zeitraumBis)
	return c.TSESignierer.SignEvent(ctx, evt, tse.ProcessTypeSonstigerVorgang, processData, kasse.EmbedTSEInTagesabschlussErstellt)
}
