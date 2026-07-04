package kasse

import (
	"encoding/json"

	e "github.com/nicograef/jotti/backend/domain/event"
)

// AbschlussSummen enthält die drei Geldbetragsfelder des
// tagesabschluss-erstellt:v1-Events.
type AbschlussSummen struct {
	UmsatzCents      int
	StornierungCents int
	GeldtransitCents int
}

// BerechneAbschlussSummen aggregiert alle Events einer Kassensitzung zu den
// drei Z-Bon-Summen gemäß reporting.sql:10-43:
//
//	Umsatz        = Zahlungen + Direktverkäufe − Direktverkauf-Storni − Warenrücknahmen
//	Stornierungen = Warenrücknahmen + Korrekturen + Direktverkauf-Storni
//	Geldtransit   = Einlagen − Entnahmen
//
// Summen-wirksam: zahlung-kassiert, stornierung-erteilt, bestellung-korrigiert,
// direktverkauf-getaetigt, direktverkauf-storniert, geldtransit-gebucht.
// Alle übrigen Typen (Bestellung, Umbuchung, Kassensturz, Differenzbuchung,
// Eröffnung, …) sind summen-neutral. Events mit nicht parsebaren Daten werden
// stillschweigend übersprungen.
// Die Funktion hat keine Repository- oder Kontext-Abhängigkeiten.
func BerechneAbschlussSummen(events []e.Event) AbschlussSummen {
	var s AbschlussSummen
	for _, evt := range events {
		switch EventType(evt.Type) {
		case EventTypeZahlungKassiertV1:
			var data ZahlungKassiertV1Data
			if err := json.Unmarshal(evt.Data, &data); err == nil {
				s.UmsatzCents += data.GesamtZahlungCents
			}
		case EventTypeStornierungErteiltV1:
			var data StornierungErteiltV1Data
			if err := json.Unmarshal(evt.Data, &data); err == nil {
				s.UmsatzCents -= data.GesamtStornierungCents
				s.StornierungCents += data.GesamtStornierungCents
			}
		case EventTypeBestellungKorrigiertV1:
			var data BestellungKorrigiertV1Data
			if err := json.Unmarshal(evt.Data, &data); err == nil {
				s.StornierungCents += data.GesamtCents
			}
		case EventTypeDirektverkaufGetaetigtV1:
			var data DirektverkaufGetaetigtV1Data
			if err := json.Unmarshal(evt.Data, &data); err == nil {
				s.UmsatzCents += data.GesamtbetragCents
			}
		case EventTypeDirektverkaufStorniertV1:
			var data DirektverkaufStorniertV1Data
			if err := json.Unmarshal(evt.Data, &data); err == nil {
				s.UmsatzCents -= data.GesamtStornierungCents
				s.StornierungCents += data.GesamtStornierungCents
			}
		case EventTypeGeldtransitGebuchtV1:
			var data GeldtransitGebuchtV1Data
			if err := json.Unmarshal(evt.Data, &data); err == nil {
				if data.Richtung == "einlage" {
					s.GeldtransitCents += data.BetragCents
				} else {
					s.GeldtransitCents -= data.BetragCents
				}
			}
		}
	}
	return s
}
