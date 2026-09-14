package application

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/nicograef/jotti/backend/api/druck/bondruck/application/escpos"
	"github.com/nicograef/jotti/backend/domain/druckstation"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
)

// positionenMitKommentarEventData spiegelt die benötigten Felder von
// bestellung-aufgenommen:v1 und direktverkauf-getaetigt:v1 in ihrer Event-Form.
// Keine Schema-Validierung nötig, da die Daten beim Event-Write validiert wurden.
type positionenMitKommentarEventData struct {
	Positionen []kasse.PositionEventData `json:"positionen"`
	Kommentar  string                    `json:"kommentar"`
}

type arbeitsbonDaten struct {
	Positionen []kasse.Position
	Kommentar  string
}

// CreateArbeitsbonAuftraegeFromEvent erzeugt Druckaufträge anhand der
// konfigurierten Druckstationen: bestellung-aufgenommen geht als Arbeitsbon je
// Kategorie an die Produktstationen; direktverkauf-getaetigt an die
// Abholbon-Station, wenn sie konfiguriert ist, sonst an die Produktstationen —
// ohne konfigurierte Station entsteht kein Auftrag. Bonmodus: pro_position
// (Standard) ein Bon je Position, pro_bestellung ein Sammelbon, am Abholbon
// zusätzlich pro_stueck je Einheit.
func CreateArbeitsbonAuftraegeFromEvent(
	evt event.Event,
	druckstationen map[string]druckstation.Druckstation,
	tischName string,
) []druckauftrag_repo.NeuerDruckauftrag {
	switch evt.Type {
	case string(kasse.EventTypeBestellungAufgenommenV1):
		return createStationsAuftraege(evt, druckstationen, tischName, fmt.Sprintf("bestellung-aufgenommen:%d", evt.ID))
	case string(kasse.EventTypeDirektverkaufGetaetigtV1):
		return createDirektverkaufAuftraege(evt, druckstationen)
	default:
		return nil
	}
}

func createDirektverkaufAuftraege(
	evt event.Event,
	druckstationen map[string]druckstation.Druckstation,
) []druckauftrag_repo.NeuerDruckauftrag {
	data, ok := unmarshalPositionenMitKommentar(evt)
	if !ok {
		return nil
	}

	referenz := fmt.Sprintf("direktverkauf-getaetigt:%d", evt.ID)

	if abholbon, ok := druckstationen[string(druckstation.KategorieAbholbon)]; ok && abholbon.DruckerIP != "" {
		return createAbholbonAuftraege(evt, data, abholbon, referenz)
	}

	return createStationsAuftraegeFromData(evt, data, druckstationen, "Direktverkauf", referenz)
}

func createAbholbonAuftraege(
	evt event.Event,
	data arbeitsbonDaten,
	station druckstation.Druckstation,
	referenz string,
) []druckauftrag_repo.NeuerDruckauftrag {
	abholbon := func(positionen []kasse.Position) druckauftrag_repo.NeuerDruckauftrag {
		payload := escpos.FormatDirektverkaufAbholbon(positionen, evt.UserName, evt.Time, data.Kommentar)
		return druckauftrag_repo.NeuerDruckauftrag{
			ZielIP:   station.DruckerIP,
			Payload:  base64.StdEncoding.EncodeToString(payload),
			BonArt:   "arbeitsbon",
			Referenz: referenz,
		}
	}

	switch station.Bonmodus {
	case druckstation.BonmodusProBestellung:
		return []druckauftrag_repo.NeuerDruckauftrag{abholbon(data.Positionen)}

	case druckstation.BonmodusProStueck:
		var auftraege []druckauftrag_repo.NeuerDruckauftrag
		for _, pos := range data.Positionen {
			einheit := pos
			einheit.Menge = 1
			for range pos.Menge {
				auftraege = append(auftraege, abholbon([]kasse.Position{einheit}))
			}
		}
		return auftraege

	default: // BonmodusProPosition, zugleich Rückfall für unbekannte Werte
		auftraege := make([]druckauftrag_repo.NeuerDruckauftrag, 0, len(data.Positionen))
		for _, pos := range data.Positionen {
			auftraege = append(auftraege, abholbon([]kasse.Position{pos}))
		}
		return auftraege
	}
}

func createStationsAuftraege(
	evt event.Event,
	druckstationen map[string]druckstation.Druckstation,
	kontextName string,
	referenz string,
) []druckauftrag_repo.NeuerDruckauftrag {
	data, ok := unmarshalPositionenMitKommentar(evt)
	if !ok {
		return nil
	}

	return createStationsAuftraegeFromData(evt, data, druckstationen, kontextName, referenz)
}

func createStationsAuftraegeFromData(
	evt event.Event,
	data arbeitsbonDaten,
	druckstationen map[string]druckstation.Druckstation,
	kontextName string,
	referenz string,
) []druckauftrag_repo.NeuerDruckauftrag {

	byKategorie := map[string][]kasse.Position{}
	for _, pos := range data.Positionen {
		byKategorie[pos.Kategorie] = append(byKategorie[pos.Kategorie], pos)
	}

	var auftraege []druckauftrag_repo.NeuerDruckauftrag
	for kategorie, positionen := range byKategorie {
		konfig, ok := druckstationen[kategorie]
		if !ok || konfig.DruckerIP == "" {
			continue
		}

		withBeep := kategorie == string(druckstation.KategorieEssen)

		if konfig.Bonmodus == druckstation.BonmodusProBestellung {
			payload := escpos.FormatSammelBon(
				positionen,
				kontextName,
				evt.UserName,
				evt.Time,
				data.Kommentar,
				withBeep,
			)
			auftraege = append(auftraege, druckauftrag_repo.NeuerDruckauftrag{
				ZielIP:   konfig.DruckerIP,
				Payload:  base64.StdEncoding.EncodeToString(payload),
				BonArt:   "arbeitsbon",
				Referenz: referenz,
			})
			continue
		}

		for _, pos := range positionen {
			payload := escpos.FormatPositionBon(
				pos,
				kontextName,
				evt.UserName,
				evt.Time,
				data.Kommentar,
				withBeep,
			)
			auftraege = append(auftraege, druckauftrag_repo.NeuerDruckauftrag{
				ZielIP:   konfig.DruckerIP,
				Payload:  base64.StdEncoding.EncodeToString(payload),
				BonArt:   "arbeitsbon",
				Referenz: referenz,
			})
			withBeep = false
		}
	}

	return auftraege
}

func unmarshalPositionenMitKommentar(evt event.Event) (arbeitsbonDaten, bool) {
	var data positionenMitKommentarEventData
	if err := json.Unmarshal(evt.Data, &data); err != nil {
		return arbeitsbonDaten{}, false
	}

	positionen := make([]kasse.Position, 0, len(data.Positionen))
	for _, pos := range data.Positionen {
		positionen = append(positionen, kasse.PositionFromEventData(pos))
	}

	return arbeitsbonDaten{Positionen: positionen, Kommentar: data.Kommentar}, true
}
