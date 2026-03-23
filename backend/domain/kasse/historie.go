package kasse

import (
	"slices"

	e "github.com/nicograef/jotti/backend/domain/event"
)

type HistorieEintragArt string

const (
	HistorieEintragBestellung  HistorieEintragArt = "bestellung"
	HistorieEintragZahlung     HistorieEintragArt = "zahlung"
	HistorieEintragStornierung HistorieEintragArt = "stornierung"
	HistorieEintragAusgabe     HistorieEintragArt = "ausgabe"
	HistorieEintragAuszahlung  HistorieEintragArt = "auszahlung"
)

type HistorieEintrag struct {
	Art         HistorieEintragArt
	Bestellung  *Bestellung
	Zahlung     *Zahlung
	Stornierung *Stornierung
	Ausgabe     *Ausgabe
	Auszahlung  *Auszahlung
}

func GetHistoryFromEvents(events []e.Event) ([]HistorieEintrag, error) {
	history := []HistorieEintrag{}

	for _, event := range events {
		switch event.Type {
		case string(EventTypeBestellungAufgenommenV1):
			bestellung, err := buildBestellungFromEvent(event)
			if err != nil {
				return []HistorieEintrag{}, err
			}
			history = append(history, HistorieEintrag{Art: HistorieEintragBestellung, Bestellung: &bestellung})
		case string(EventTypeZahlungKassiertV1):
			zahlung, err := buildZahlungFromEvent(event)
			if err != nil {
				return []HistorieEintrag{}, err
			}
			history = append(history, HistorieEintrag{Art: HistorieEintragZahlung, Zahlung: &zahlung})
		case string(EventTypeStornierungErteiltV1):
			stornierung, err := buildStornierungFromEvent(event)
			if err != nil {
				return []HistorieEintrag{}, err
			}
			history = append(history, HistorieEintrag{Art: HistorieEintragStornierung, Stornierung: &stornierung})

		case string(EventTypeAusgabeBestaetigtV1):
			ausgabe, err := buildAusgabeFromEvent(event)
			if err != nil {
				return []HistorieEintrag{}, err
			}
			history = append(history, HistorieEintrag{Art: HistorieEintragAusgabe, Ausgabe: &ausgabe})

		case string(EventTypeAuszahlungGeleistetV1):
			auszahlung, err := buildAuszahlungFromEvent(event)
			if err != nil {
				return []HistorieEintrag{}, err
			}
			history = append(history, HistorieEintrag{Art: HistorieEintragAuszahlung, Auszahlung: &auszahlung})
		}
	}

	// reverse the order of the array so that the most recent event is first
	slices.Reverse(history)

	return history, nil
}
