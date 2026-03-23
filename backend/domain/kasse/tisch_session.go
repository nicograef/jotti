package kasse

import (
	"encoding/json"
	"fmt"

	e "github.com/nicograef/jotti/backend/domain/event"
)

// TischSession represents the projected state of a table session, derived from applying events.
// Zero-value represents a table session with no events (Saldo 0, empty lists).
type TischSession struct {
	Subject               string
	TischID               int
	KassensitzungNr       int
	SaldoCents            int
	UnbezahltePositionen  []Position
	AusstehendePositionen []Position
	GesamtZahlungenCents  int
	LastEventID           int
	LastEventVersion      int
}

// ApplyEvent applies a single domain event to the current TischSession and returns the new state.
func ApplyEvent(state TischSession, evt e.Event) (TischSession, error) {
	switch evt.Type {
	case string(EventTypeBestellungAufgenommenV1):
		var data bestellungAufgenommenV1Data
		if err := json.Unmarshal(evt.Data, &data); err != nil {
			return state, fmt.Errorf("unmarshal bestellung data: %w", err)
		}
		state.SaldoCents += data.GesamtPreisCents
		state.UnbezahltePositionen = accumulatePositionen(state.UnbezahltePositionen, fromPositionenEventData(data.Positionen))
		state.AusstehendePositionen = accumulatePositionen(state.AusstehendePositionen, fromPositionenEventData(data.Positionen))

	case string(EventTypeZahlungKassiertV1):
		var data zahlungKassiertV1Data
		if err := json.Unmarshal(evt.Data, &data); err != nil {
			return state, fmt.Errorf("unmarshal zahlung data: %w", err)
		}
		state.SaldoCents -= data.GesamtZahlungCents
		state.GesamtZahlungenCents += data.GesamtZahlungCents
		state.UnbezahltePositionen = reduceByPosition(state.UnbezahltePositionen, fromPositionenEventData(data.Positionen))

	case string(EventTypeStornierungErteiltV1):
		var data stornierungErteiltV1Data
		if err := json.Unmarshal(evt.Data, &data); err != nil {
			return state, fmt.Errorf("unmarshal stornierung data: %w", err)
		}
		state.SaldoCents -= data.GesamtStornierungCents
		state.UnbezahltePositionen = reduceByPosition(state.UnbezahltePositionen, fromPositionenEventData(data.Positionen))
		state.AusstehendePositionen = reduceByPosition(state.AusstehendePositionen, fromPositionenEventData(data.Positionen))

	case string(EventTypeAusgabeBestaetigtV1):
		var data ausgabeBestaetigtV1Data
		if err := json.Unmarshal(evt.Data, &data); err != nil {
			return state, fmt.Errorf("unmarshal ausgabe data: %w", err)
		}
		state.AusstehendePositionen = reduceByPosition(state.AusstehendePositionen, fromPositionenEventData(data.Positionen))

	case string(EventTypeAuszahlungGeleistetV1):
		var data auszahlungGeleistetV1Data
		if err := json.Unmarshal(evt.Data, &data); err != nil {
			return state, fmt.Errorf("unmarshal auszahlung data: %w", err)
		}
		state.SaldoCents += data.BetragCents

	default:
		return state, fmt.Errorf("unknown event type: %s", evt.Type)
	}

	state.LastEventID = evt.ID
	state.LastEventVersion = evt.Version

	return state, nil
}

// ComputeNichtStorniertePositionen replays events to compute all positions that were ordered
// but not yet cancelled. Used on-demand for stornierung validation.
func ComputeNichtStorniertePositionen(events []e.Event) ([]Position, error) {
	var nichtStorniert []Position

	for _, evt := range events {
		switch evt.Type {
		case string(EventTypeBestellungAufgenommenV1):
			var data bestellungAufgenommenV1Data
			if err := json.Unmarshal(evt.Data, &data); err != nil {
				return nil, fmt.Errorf("unmarshal bestellung data: %w", err)
			}
			nichtStorniert = accumulatePositionen(nichtStorniert, fromPositionenEventData(data.Positionen))

		case string(EventTypeStornierungErteiltV1):
			var data stornierungErteiltV1Data
			if err := json.Unmarshal(evt.Data, &data); err != nil {
				return nil, fmt.Errorf("unmarshal stornierung data: %w", err)
			}
			nichtStorniert = reduceByPosition(nichtStorniert, fromPositionenEventData(data.Positionen))

		case string(EventTypeZahlungKassiertV1), string(EventTypeAusgabeBestaetigtV1), string(EventTypeAuszahlungGeleistetV1):
			continue

		default:
			return nil, fmt.Errorf("unknown event type: %s", evt.Type)
		}
	}

	return nichtStorniert, nil
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
