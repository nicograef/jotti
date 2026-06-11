package application

import (
	"context"

	tseApp "github.com/nicograef/jotti/backend/api/tse/application"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/rs/zerolog"
)

type direktverkaufGetaetigtV1Data struct {
	VerkaufID         string           `json:"verkaufId"`
	GesamtbetragCents int              `json:"gesamtbetragCents"`
	Positionen        []kasse.Position `json:"positionen"`
	Kommentar         string           `json:"kommentar"`
	TSETxID           string           `json:"tseTxId,omitempty"`
	TSEData           *kasse.TSEData   `json:"tseData,omitempty"`
	TSEAusfall        bool             `json:"tseAusfall,omitempty"`
}

type direktverkaufStorniertV1Data struct {
	StornierungID          string           `json:"stornierungId"`
	VerkaufID              string           `json:"verkaufId"`
	Positionen             []kasse.Position `json:"positionen"`
	GesamtStornierungCents int              `json:"gesamtStornierungCents"`
	Kommentar              string           `json:"kommentar"`
	TSETxID                string           `json:"tseTxId,omitempty"`
	TSEData                *kasse.TSEData   `json:"tseData,omitempty"`
}

func (c Command) signDirektverkaufGetaetigtEvent(ctx context.Context, evt event.Event, positionen []kasse.Position) (tseApp.Signierung, error) {
	processData, err := tseApp.BuildKassenbelegProcessData(positionen, sumPositionenCents(positionen))
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("Failed to build TSE process_data for direktverkauf")
		return tseApp.Signierung{}, ErrDatabase
	}

	return c.TSESignierer.SignEvent(ctx, evt, tse.ProcessTypeKassenbelegV1, processData, withDirektverkaufGetaetigtEventTSE)
}

func (c Command) signDirektverkaufStorniertEvent(ctx context.Context, evt event.Event, positionen []kasse.Position, stornoBetragCents int) (tseApp.Signierung, error) {
	processData, err := tseApp.BuildKassenbelegProcessDataWithFaktor(positionen, -stornoBetragCents, -1)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("Failed to build TSE process_data for direktverkauf storno")
		return tseApp.Signierung{}, ErrDatabase
	}

	return c.TSESignierer.SignEvent(ctx, evt, tse.ProcessTypeKassenbelegV1, processData, withDirektverkaufStorniertEventTSE)
}

var withDirektverkaufGetaetigtEventTSE = tseApp.EmbedTSEInData(kasse.EventTypeDirektverkaufGetaetigtV1, func(data *direktverkaufGetaetigtV1Data, txID string, tseData *kasse.TSEData) {
	data.TSETxID = txID
	data.TSEData = tseData
	data.TSEAusfall = tseData == nil
})

var withDirektverkaufStorniertEventTSE = tseApp.EmbedTSEInData(kasse.EventTypeDirektverkaufStorniertV1, func(data *direktverkaufStorniertV1Data, txID string, tseData *kasse.TSEData) {
	data.TSETxID = txID
	data.TSEData = tseData
})

func sumPositionenCents(positionen []kasse.Position) int {
	sum := 0
	for _, pos := range positionen {
		sum += pos.Einzelpreis * pos.Menge
	}
	return sum
}
