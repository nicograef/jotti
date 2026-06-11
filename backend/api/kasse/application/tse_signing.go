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

type geldtransitGebuchtV1Data struct {
	BewegungID  string         `json:"bewegungId"`
	Richtung    string         `json:"richtung"`
	BetragCents int            `json:"betragCents"`
	Kommentar   string         `json:"kommentar"`
	GebuchtVon  int            `json:"gebuchtVon"`
	TSETxID     string         `json:"tseTxId,omitempty"`
	TSEData     *kasse.TSEData `json:"tseData,omitempty"`
}

type differenzSollIstGebuchtV1Data struct {
	BetragCents int            `json:"betragCents"`
	GebuchtVon  int            `json:"gebuchtVon"`
	TSETxID     string         `json:"tseTxId,omitempty"`
	TSEData     *kasse.TSEData `json:"tseData,omitempty"`
}

type tagesabschlussErstelltV1Data struct {
	ZNr               int            `json:"zNr"`
	ZeitraumVon       time.Time      `json:"zeitraumVon"`
	ZeitraumBis       time.Time      `json:"zeitraumBis"`
	UmsatzGesamtCents int            `json:"umsatzGesamtCents"`
	StornierungCents  int            `json:"stornierungCents"`
	AuszahlungenCents int            `json:"auszahlungenCents"`
	GeldtransitCents  int            `json:"geldtransitCents"`
	ErstelltVon       int            `json:"erstelltVon"`
	TSETxID           string         `json:"tseTxId,omitempty"`
	TSEData           *kasse.TSEData `json:"tseData,omitempty"`
}

func (c Command) signGeldtransitGebuchtEvent(ctx context.Context, evt event.Event, richtung string, betragCents int) (tseApp.Signierung, error) {
	processData, err := tseApp.BuildGeldtransitProcessData(richtung, betragCents)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("Failed to build TSE process_data for geldtransit")
		return tseApp.Signierung{}, ErrDatabase
	}

	return c.TSESignierer.SignEvent(ctx, evt, tse.ProcessTypeKassenbelegV1, processData, withGeldtransitEventTSE)
}

func (c Command) signDifferenzSollIstGebuchtEvent(ctx context.Context, evt event.Event, differenzCents int) (tseApp.Signierung, error) {
	processData := tseApp.BuildEigenbelegProcessData(differenzCents)
	return c.TSESignierer.SignEvent(ctx, evt, tse.ProcessTypeKassenbelegV1, processData, withDifferenzEventTSE)
}

func (c Command) signTagesabschlussErstelltEvent(ctx context.Context, evt event.Event, zNr int, zeitraumVon time.Time, zeitraumBis time.Time) (tseApp.Signierung, error) {
	processData := tseApp.BuildTagesabschlussProcessData(zNr, zeitraumVon, zeitraumBis)
	return c.TSESignierer.SignEvent(ctx, evt, tse.ProcessTypeSonstigerVorgang, processData, withTagesabschlussEventTSE)
}

var withGeldtransitEventTSE = tseApp.EmbedTSEInData(kasse.EventTypeGeldtransitGebuchtV1, func(data *geldtransitGebuchtV1Data, txID string, tseData *kasse.TSEData) {
	data.TSETxID = txID
	data.TSEData = tseData
})

var withDifferenzEventTSE = tseApp.EmbedTSEInData(kasse.EventTypeDifferenzSollIstGebuchtV1, func(data *differenzSollIstGebuchtV1Data, txID string, tseData *kasse.TSEData) {
	data.TSETxID = txID
	data.TSEData = tseData
})

var withTagesabschlussEventTSE = tseApp.EmbedTSEInData(kasse.EventTypeTagesabschlussErstelltV1, func(data *tagesabschlussErstelltV1Data, txID string, tseData *kasse.TSEData) {
	data.TSETxID = txID
	data.TSEData = tseData
})
