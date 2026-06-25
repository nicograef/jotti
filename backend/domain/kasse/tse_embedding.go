package kasse

import (
	"encoding/json"
	"fmt"

	e "github.com/nicograef/jotti/backend/domain/event"
)

// embedTSEInData baut eine Embed-Funktion fuer einen Event-Data-Typ T: Typ-Check,
// optionale TSEData-Validierung und JSON-Roundtrip sind fuer alle Event-Typen
// identisch; nur das Setzen der TSE-Felder (apply) ist typspezifisch. Die tx-ID
// wird immer gesetzt, die TSE-Signaturdaten nur bei Erfolg (tseData != nil).
func embedTSEInData[T any](eventType EventType, apply func(data *T, txID string, tseData *TSEData)) func(e.Event, string, *TSEData) (e.Event, error) {
	return func(evt e.Event, txID string, tseData *TSEData) (e.Event, error) {
		if evt.Type != string(eventType) {
			return e.Event{}, fmt.Errorf("unsupported event type for TSE data: %s", evt.Type)
		}
		if tseData != nil {
			if err := tseData.Validate(); err != nil {
				return e.Event{}, err
			}
		}

		var data T
		if err := json.Unmarshal(evt.Data, &data); err != nil {
			return e.Event{}, err
		}

		apply(&data, txID, tseData)

		encoded, err := json.Marshal(data)
		if err != nil {
			return e.Event{}, err
		}

		evt.Data = encoded
		return evt, nil
	}
}

// Pro signierbarem Event-Typ existiert eine Embed-Funktion, die das Ergebnis eines
// TSE-Signierversuchs in die Event-Daten schreibt (Signatur kompatibel zu
// tseApp.EmbedTSE). tseData == nil markiert den Ausfall; Event-Typen mit
// Ausfallvermerk setzen dann zusaetzlich ihr TSEAusfall-Flag.
var (
	EmbedTSEInBestellungAufgenommen = embedTSEInData(EventTypeBestellungAufgenommenV1, func(data *BestellungAufgenommenV1Data, txID string, tseData *TSEData) {
		data.TSETxID = txID
		data.TSEData = tseData
	})

	EmbedTSEInZahlungKassiert = embedTSEInData(EventTypeZahlungKassiertV1, func(data *ZahlungKassiertV1Data, txID string, tseData *TSEData) {
		data.TSETxID = txID
		data.TSEData = tseData
		data.TSEAusfall = tseData == nil
	})

	EmbedTSEInStornierungErteilt = embedTSEInData(EventTypeStornierungErteiltV1, func(data *StornierungErteiltV1Data, txID string, tseData *TSEData) {
		data.TSETxID = txID
		data.TSEData = tseData
	})

	EmbedTSEInBestellungUmgebucht = embedTSEInData(EventTypeBestellungUmgebuchtV1, func(data *BestellungUmgebuchtV1Data, txID string, tseData *TSEData) {
		data.TSETxID = txID
		data.TSEData = tseData
	})

	EmbedTSEInAuszahlungGeleistet = embedTSEInData(EventTypeAuszahlungGeleistetV1, func(data *AuszahlungGeleistetV1Data, txID string, tseData *TSEData) {
		data.TSETxID = txID
		data.TSEData = tseData
	})

	EmbedTSEInDirektverkaufGetaetigt = embedTSEInData(EventTypeDirektverkaufGetaetigtV1, func(data *DirektverkaufGetaetigtV1Data, txID string, tseData *TSEData) {
		data.TSETxID = txID
		data.TSEData = tseData
		data.TSEAusfall = tseData == nil
	})

	EmbedTSEInDirektverkaufStorniert = embedTSEInData(EventTypeDirektverkaufStorniertV1, func(data *DirektverkaufStorniertV1Data, txID string, tseData *TSEData) {
		data.TSETxID = txID
		data.TSEData = tseData
	})

	EmbedTSEInGeldtransitGebucht = embedTSEInData(EventTypeGeldtransitGebuchtV1, func(data *GeldtransitGebuchtV1Data, txID string, tseData *TSEData) {
		data.TSETxID = txID
		data.TSEData = tseData
	})

	EmbedTSEInDifferenzSollIstGebucht = embedTSEInData(EventTypeDifferenzSollIstGebuchtV1, func(data *DifferenzSollIstGebuchtV1Data, txID string, tseData *TSEData) {
		data.TSETxID = txID
		data.TSEData = tseData
	})

	EmbedTSEInTagesabschlussErstellt = embedTSEInData(EventTypeTagesabschlussErstelltV1, func(data *TagesabschlussErstelltV1Data, txID string, tseData *TSEData) {
		data.TSETxID = txID
		data.TSEData = tseData
	})
)
