package application

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nicograef/jotti/backend/api/bondruck/application/escpos"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
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

// bestellungEventData spiegelt die relevanten Felder von BestellungAufgenommenV1.
// Keine Schema-Validierung noetig, da die Daten beim Event-Write validiert wurden.
type bestellungEventData struct {
	Positionen []kasse.Position `json:"positionen"`
	Kommentar  string           `json:"kommentar"`
}

// CreateArbeitsbonAuftraegeFromEvent erzeugt Druckauftraege aus einem BestellungAufgenommen-Event.
// - pro_position (Standard): 1 Bon pro Position
// - pro_bestellung: 1 Sammelbon pro Kategorie
func CreateArbeitsbonAuftraegeFromEvent(
	evt event.Event,
	druckstationen map[string]Druckstation,
) []Druckauftrag {
	var data bestellungEventData
	if err := json.Unmarshal(evt.Data, &data); err != nil {
		return nil
	}

	tischName := parseTischName(evt.Subject)

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
		referenz := fmt.Sprintf("bestellung-aufgenommen:%d", evt.ID)

		if konfig.Bonmodus == "pro_bestellung" {
			payload := escpos.FormatSammelBon(
				positionen,
				tischName,
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
				tischName,
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

// parseTischName converts an Event Subject to a human-readable table name.
// New format: "kassensitzung-{nr}/tisch-{id}" -> "Tisch {id}"
// Legacy format: "tisch:{id}" -> "Tisch {id}"
func parseTischName(subject string) string {
	if idx := strings.LastIndex(subject, "/tisch-"); idx != -1 {
		return "Tisch " + subject[idx+len("/tisch-"):]
	}

	parts := strings.SplitN(subject, ":", 2)
	if len(parts) == 2 {
		return "Tisch " + parts[1]
	}
	return subject
}
