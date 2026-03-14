package table

import (
	"fmt"
	"slices"
	"strconv"

	e "github.com/nicograef/jotti/backend/domain/event"
)

type EventType string

type HistorieEintragArt string

const (
	HistorieEintragBestellung  HistorieEintragArt = "bestellung"
	HistorieEintragZahlung     HistorieEintragArt = "zahlung"
	HistorieEintragStornierung HistorieEintragArt = "stornierung"
	HistorieEintragAusgabe     HistorieEintragArt = "ausgabe"
)

type HistorieEintrag struct {
	Art         HistorieEintragArt
	Bestellung  *Bestellung
	Zahlung     *Zahlung
	Stornierung *Stornierung
	Ausgabe     *Ausgabe
}

const (
	EventTypeBestellungAufgenommenV1 EventType = "tisch.bestellung-aufgenommen:v1"
	EventTypeZahlungKassiertV1       EventType = "tisch.zahlung-kassiert:v1"
	EventTypeStornierungErteiltV1    EventType = "tisch.stornierung-erteilt:v1"
	EventTypeAusgabeBestaetigtV1     EventType = "tisch.ausgabe-bestaetigt:v1"
)

// parseTischIDFromSubject extracts the table ID from an event subject like "tisch:42".
func parseTischIDFromSubject(subject string) (int, error) {
	const prefix = "tisch:"
	if len(subject) <= len(prefix) || subject[:len(prefix)] != prefix {
		return 0, fmt.Errorf("invalid event subject format: %s", subject)
	}
	id, err := strconv.Atoi(subject[len(prefix):])
	if err != nil {
		return 0, fmt.Errorf("invalid tisch ID in event subject: %w", err)
	}
	return id, nil
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
		}
	}

	// reverse the order of the array so that the most recent event is first
	slices.Reverse(history)

	return history, nil
}

// accumulatePositionen adds positions to a list, merging quantities for matching positions (by PositionID)
func accumulatePositionen(list []Position, positionen []Position) []Position {
	for _, pos := range positionen {
		found := false
		for i, existing := range list {
			if existing.PositionID == pos.PositionID {
				list[i].Menge += pos.Menge
				found = true
				break
			}
		}
		if !found {
			list = append(list, pos)
		}
	}
	return list
}

// reduceByPosition subtracts positions from a list, removing entries when quantity reaches zero
func reduceByPosition(list []Position, reductions []Position) []Position {
	for _, red := range reductions {
		for i := 0; i < len(list); i++ {
			if list[i].PositionID == red.PositionID {
				if list[i].Menge > red.Menge {
					list[i].Menge -= red.Menge
				} else {
					list = append(list[:i], list[i+1:]...)
					i--
				}
				break
			}
		}
	}
	return list
}
