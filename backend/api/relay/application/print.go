package application

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/nicograef/jotti/backend/api/relay/application/escpos"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
)

// DruckerKonfig enthält die Konfiguration eines Druckers für eine Kategorie.
type DruckerKonfig struct {
	IP       string // z.B. "192.168.1.51", leer = kein Drucker
	Bonmodus string // "pro_position" (Standard) oder "pro_bestellung"
}

// bestellungEventData spiegelt die relevanten Felder von BestellungAufgenommenV1.
// Keine Schema-Validierung nötig — die Daten wurden beim Schreiben bereits validiert.
type bestellungEventData struct {
	Positionen []kasse.Position `json:"positionen"`
	Kommentar  string           `json:"kommentar"`
}

// createDruckAuftraegeFromEvent erzeugt Druck-Aufträge aus einem BestellungAufgenommen-Event.
//   - pro_position (Standard): 1 Bon pro Position (jede Position = eigener Arbeitsauftrag)
//   - pro_bestellung: 1 Sammelbon pro Kategorie (alle Positionen einer Kategorie auf einem Bon)
//
// Kategorien ohne konfigurierte drucker_ip werden übersprungen (kein Fehler).
func createDruckAuftraegeFromEvent(
	evt event.Event,
	druckerConfig map[string]DruckerKonfig, // kategorie → DruckerKonfig
) []DruckAuftrag {
	var data bestellungEventData
	if err := json.Unmarshal(evt.Data, &data); err != nil {
		return nil
	}

	tischName := parseTischName(evt.Subject) // "tisch:7" → "Tisch 7"

	// Positionen nach Kategorie gruppieren
	byKategorie := map[string][]kasse.Position{}
	for _, pos := range data.Positionen {
		byKategorie[pos.Kategorie] = append(byKategorie[pos.Kategorie], pos)
	}

	var auftraege []DruckAuftrag

	for kategorie, positionen := range byKategorie {
		konfig, ok := druckerConfig[kategorie]
		if !ok || konfig.IP == "" {
			continue // Kein Drucker konfiguriert → überspringen
		}

		withBeep := kategorie == "essen" // Küche: Piepser aktivieren

		if konfig.Bonmodus == "pro_bestellung" {
			// Sammelbon: 1 Bon pro Bestellung (pro Kategorie)
			payload := escpos.FormatSammelBon(
				positionen, tischName, evt.UserName, evt.Time,
				data.Kommentar, withBeep,
			)
			auftraege = append(auftraege, DruckAuftrag{
				EventID:   evt.ID,
				DruckerIP: konfig.IP,
				Payload:   base64.StdEncoding.EncodeToString(payload),
			})
		} else {
			// Standard: 1 Bon pro Position
			for _, pos := range positionen {
				payload := escpos.FormatPositionBon(
					pos, tischName, evt.UserName, evt.Time,
					data.Kommentar, withBeep,
				)
				auftraege = append(auftraege, DruckAuftrag{
					EventID:   evt.ID,
					DruckerIP: konfig.IP,
					Payload:   base64.StdEncoding.EncodeToString(payload),
				})
				withBeep = false // Nur beim ersten Bon einer Kategorie piepsen
			}
		}
	}

	return auftraege
}

// parseTischName converts an Event Subject to a human-readable table name.
// New format: "kassensitzung-{nr}/tisch-{id}" → "Tisch {id}"
// Legacy format: "tisch:{id}" → "Tisch {id}"
func parseTischName(subject string) string {
	// New format: "kassensitzung-1/tisch-42"
	if idx := strings.LastIndex(subject, "/tisch-"); idx != -1 {
		return "Tisch " + subject[idx+len("/tisch-"):]
	}
	// Legacy format: "tisch:42"
	parts := strings.SplitN(subject, ":", 2)
	if len(parts) == 2 {
		return "Tisch " + parts[1]
	}
	return subject
}
