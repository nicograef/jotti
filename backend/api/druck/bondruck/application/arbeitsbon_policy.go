package application

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nicograef/jotti/backend/api/druck/bondruck/application/escpos"
	"github.com/nicograef/jotti/backend/domain/druckstation"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
)

// positionenMitKommentarData spiegelt die benoetigten Felder von
// bestellung-aufgenommen:v1 und direktverkauf-getaetigt:v1.
// Keine Schema-Validierung noetig, da die Daten beim Event-Write validiert wurden.
type positionenMitKommentarData struct {
	Positionen []kasse.Position `json:"positionen"`
	Kommentar  string           `json:"kommentar"`
}

// CreateArbeitsbonAuftraegeFromEvent erzeugt Druckauftraege aus einem Bestell- oder
// Direktverkauf-Event anhand der konfigurierten Druckstationen.
//   - bestellung-aufgenommen: Arbeitsbons an die Produktstationen je Kategorie.
//   - direktverkauf-getaetigt (Ableitungsregel): ist die Abholbon-Station konfiguriert,
//     entstehen Abholbon(s) an dieser Station gemaess ihrem Bonmodus; sonst Arbeitsbons
//     an die Produktstationen; ohne konfigurierte Stationen entstehen keine Auftraege.
//
// Bonmodus pro_position (Standard) erzeugt einen Bon je Position, pro_bestellung einen
// Sammelbon je Kategorie bzw. einen Sammel-Abholbon.
func CreateArbeitsbonAuftraegeFromEvent(
	evt event.Event,
	druckstationen map[string]druckstation.Druckstation,
) []druckauftrag_repo.NeuerDruckauftrag {
	switch evt.Type {
	case string(kasse.EventTypeBestellungAufgenommenV1):
		return createStationsAuftraege(evt, druckstationen, parseTischName(evt.Subject), fmt.Sprintf("bestellung-aufgenommen:%d", evt.ID))
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

	// Ableitungsregel: Abholbon-Station konfiguriert -> Abholbon(s), sonst Produktstationen.
	if abholbon, ok := druckstationen["abholbon"]; ok && abholbon.DruckerIP != "" {
		return createAbholbonAuftraege(evt, data, abholbon, referenz)
	}

	return createStationsAuftraegeFromData(evt, data, druckstationen, "Direktverkauf", referenz)
}

// createAbholbonAuftraege erzeugt Abholbons fuer einen Direktverkauf gemaess Bonmodus:
// pro_bestellung = ein Sammel-Abholbon, pro_position = ein Abholbon je Position.
func createAbholbonAuftraege(
	evt event.Event,
	data positionenMitKommentarData,
	station druckstation.Druckstation,
	referenz string,
) []druckauftrag_repo.NeuerDruckauftrag {
	if station.Bonmodus == "pro_bestellung" {
		payload := escpos.FormatDirektverkaufAbholbon(data.Positionen, evt.UserName, evt.Time, data.Kommentar)
		return []druckauftrag_repo.NeuerDruckauftrag{{
			ZielIP:   station.DruckerIP,
			Payload:  base64.StdEncoding.EncodeToString(payload),
			BonArt:   "arbeitsbon",
			Referenz: referenz,
		}}
	}

	auftraege := make([]druckauftrag_repo.NeuerDruckauftrag, 0, len(data.Positionen))
	for _, pos := range data.Positionen {
		payload := escpos.FormatDirektverkaufAbholbon([]kasse.Position{pos}, evt.UserName, evt.Time, data.Kommentar)
		auftraege = append(auftraege, druckauftrag_repo.NeuerDruckauftrag{
			ZielIP:   station.DruckerIP,
			Payload:  base64.StdEncoding.EncodeToString(payload),
			BonArt:   "arbeitsbon",
			Referenz: referenz,
		})
	}
	return auftraege
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
	data positionenMitKommentarData,
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

		withBeep := kategorie == "essen"

		if konfig.Bonmodus == "pro_bestellung" {
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

func unmarshalPositionenMitKommentar(evt event.Event) (positionenMitKommentarData, bool) {
	var data positionenMitKommentarData
	if err := json.Unmarshal(evt.Data, &data); err != nil {
		return positionenMitKommentarData{}, false
	}

	return data, true
}

// parseTischName converts an Event Subject to a human-readable table name.
// Format: "kassensitzung-{nr}/tisch-{id}" -> "Tisch {id}".
// If "/tisch-" is missing, the subject is returned unchanged.
func parseTischName(subject string) string {
	if idx := strings.LastIndex(subject, "/tisch-"); idx != -1 {
		return "Tisch " + subject[idx+len("/tisch-"):]
	}
	return subject
}
