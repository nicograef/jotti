package application

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nicograef/jotti/backend/api/bondruck/application/escpos"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/settings"
)

type Druckstation struct {
	IP       string // z. B. "192.168.1.51", leer = kein Drucker
	Bonmodus string // "pro_position" (Standard) oder "pro_bestellung"
}

type Druckauftrag struct {
	ZielIP   string
	Payload  string
	BonArt   string
	Referenz string
}

type DirektverkaufBondruckKonfiguration struct {
	Modus             settings.DirektverkaufModus
	AbholbonDruckerIP string
}

// positionenMitKommentarData spiegelt die benoetigten Felder von
// bestellung-aufgenommen:v1 und direktverkauf-getaetigt:v1.
// Keine Schema-Validierung noetig, da die Daten beim Event-Write validiert wurden.
type positionenMitKommentarData struct {
	Positionen []kasse.Position `json:"positionen"`
	Kommentar  string           `json:"kommentar"`
}

// CreateArbeitsbonAuftraegeFromEvent erzeugt Druckauftraege aus einem BestellungAufgenommen-Event.
// - pro_position (Standard): 1 Bon pro Position
// - pro_bestellung: 1 Sammelbon pro Kategorie
func CreateArbeitsbonAuftraegeFromEvent(
	evt event.Event,
	druckstationen map[string]Druckstation,
	direktverkaufKonfig DirektverkaufBondruckKonfiguration,
) []Druckauftrag {
	switch evt.Type {
	case string(kasse.EventTypeBestellungAufgenommenV1):
		return createStationsAuftraege(evt, druckstationen, parseTischName(evt.Subject), fmt.Sprintf("bestellung-aufgenommen:%d", evt.ID))
	case string(kasse.EventTypeDirektverkaufGetaetigtV1):
		return createDirektverkaufAuftraege(evt, druckstationen, direktverkaufKonfig)
	default:
		return nil
	}
}

func createDirektverkaufAuftraege(
	evt event.Event,
	druckstationen map[string]Druckstation,
	direktverkaufKonfig DirektverkaufBondruckKonfiguration,
) []Druckauftrag {
	if direktverkaufKonfig.Modus == settings.DirektverkaufModusKeinBon {
		return nil
	}

	data, ok := unmarshalPositionenMitKommentar(evt)
	if !ok {
		return nil
	}

	referenz := fmt.Sprintf("direktverkauf-getaetigt:%d", evt.ID)

	switch direktverkaufKonfig.Modus {
	case settings.DirektverkaufModusAnStationen:
		return createStationsAuftraegeFromData(evt, data, druckstationen, "Direktverkauf", referenz)
	case settings.DirektverkaufModusAbholbon:
		if direktverkaufKonfig.AbholbonDruckerIP == "" {
			return nil
		}

		payload := escpos.FormatDirektverkaufAbholbon(data.Positionen, evt.UserName, evt.Time, data.Kommentar)
		return []Druckauftrag{{
			ZielIP:   direktverkaufKonfig.AbholbonDruckerIP,
			Payload:  base64.StdEncoding.EncodeToString(payload),
			BonArt:   "arbeitsbon",
			Referenz: referenz,
		}}
	default:
		return nil
	}
}

func createStationsAuftraege(
	evt event.Event,
	druckstationen map[string]Druckstation,
	kontextName string,
	referenz string,
) []Druckauftrag {
	data, ok := unmarshalPositionenMitKommentar(evt)
	if !ok {
		return nil
	}

	return createStationsAuftraegeFromData(evt, data, druckstationen, kontextName, referenz)
}

func createStationsAuftraegeFromData(
	evt event.Event,
	data positionenMitKommentarData,
	druckstationen map[string]Druckstation,
	kontextName string,
	referenz string,
) []Druckauftrag {

	byKategorie := map[string][]kasse.Position{}
	for _, pos := range data.Positionen {
		byKategorie[pos.Kategorie] = append(byKategorie[pos.Kategorie], pos)
	}

	var auftraege []Druckauftrag
	for kategorie, positionen := range byKategorie {
		konfig, ok := druckstationen[kategorie]
		if !ok || konfig.IP == "" {
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
			auftraege = append(auftraege, Druckauftrag{
				ZielIP:   konfig.IP,
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
			auftraege = append(auftraege, Druckauftrag{
				ZielIP:   konfig.IP,
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
