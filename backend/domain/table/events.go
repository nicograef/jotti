package table

import (
	"fmt"
	"slices"
	"strconv"

	e "github.com/nicograef/jotti/backend/domain/event"
)

type EventType string

const (
	EventTypeBestellungAufgegebenV1 EventType = "tisch.bestellung-aufgegeben:v1"
	EventTypeZahlungRegistriertV1   EventType = "tisch.zahlung-registriert:v1"
	EventTypeProdukteStorniertV1    EventType = "tisch.produkte-storniert:v1"
	EventTypeProdukteGeliefertV1    EventType = "tisch.produkte-geliefert:v1"
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

func GetHistoryFromEvents(events []e.Event) ([]any, error) {
	history := []any{}

	for _, event := range events {
		switch event.Type {
		case string(EventTypeBestellungAufgegebenV1):
			bestellung, err := buildBestellungFromEvent(event)
			if err != nil {
				return []any{}, err
			}
			history = append(history, bestellung)
		case string(EventTypeZahlungRegistriertV1):
			zahlung, err := buildZahlungFromEvent(event)
			if err != nil {
				return []any{}, err
			}
			history = append(history, zahlung)
		case string(EventTypeProdukteStorniertV1):
			stornierung, err := buildStornierungFromEvent(event)
			if err != nil {
				return []any{}, err
			}
			history = append(history, stornierung)

		case string(EventTypeProdukteGeliefertV1):
			lieferung, err := buildLieferungFromEvent(event)
			if err != nil {
				return []any{}, err
			}
			history = append(history, lieferung)
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
