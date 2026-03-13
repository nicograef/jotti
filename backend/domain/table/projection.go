package table

import (
	"encoding/json"
	"fmt"

	e "github.com/nicograef/jotti/backend/domain/event"
)

// TischState represents the projected state of a table, derived from applying events.
// Zero-value represents a table with no events (Saldo 0, empty lists).
type TischState struct {
	TischID                int
	TischName              string
	SaldoCents             int
	UnbezahltePositionen   []Position
	UngeliefertePositionen []Position
	GesamtZahlungenCents   int
	LastEventID            int
	LastEventVersion       int
}

// ApplyEvent applies a single domain event to the current TischState and returns the new state.
func ApplyEvent(state TischState, evt e.Event) (TischState, error) {
	switch evt.Type {
	case string(EventTypeBestellungAufgegebenV1):
		var data bestellungAufgegebenV1Data
		if err := json.Unmarshal(evt.Data, &data); err != nil {
			return state, fmt.Errorf("unmarshal bestellung data: %w", err)
		}
		state.SaldoCents += data.GesamtPreisCents
		state.UnbezahltePositionen = accumulatePositionen(state.UnbezahltePositionen, data.Positionen)
		state.UngeliefertePositionen = accumulatePositionen(state.UngeliefertePositionen, data.Positionen)

	case string(EventTypeZahlungRegistriertV1):
		var data zahlungRegistriertV1Data
		if err := json.Unmarshal(evt.Data, &data); err != nil {
			return state, fmt.Errorf("unmarshal zahlung data: %w", err)
		}
		state.SaldoCents -= data.GesamtZahlungCents
		state.GesamtZahlungenCents += data.GesamtZahlungCents
		state.UnbezahltePositionen = reduceByPosition(state.UnbezahltePositionen, data.Positionen)

	case string(EventTypeProdukteStorniertV1):
		var data produkteStorniertV1Data
		if err := json.Unmarshal(evt.Data, &data); err != nil {
			return state, fmt.Errorf("unmarshal stornierung data: %w", err)
		}
		state.SaldoCents -= data.GesamtStornierungCents
		state.UnbezahltePositionen = reduceByPosition(state.UnbezahltePositionen, data.Positionen)
		state.UngeliefertePositionen = reduceByPosition(state.UngeliefertePositionen, data.Positionen)

	case string(EventTypeProdukteGeliefertV1):
		var data produkteGeliefertV1Data
		if err := json.Unmarshal(evt.Data, &data); err != nil {
			return state, fmt.Errorf("unmarshal lieferung data: %w", err)
		}
		state.UngeliefertePositionen = reduceByPosition(state.UngeliefertePositionen, data.Positionen)

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
		case string(EventTypeBestellungAufgegebenV1):
			var data bestellungAufgegebenV1Data
			if err := json.Unmarshal(evt.Data, &data); err != nil {
				return nil, fmt.Errorf("unmarshal bestellung data: %w", err)
			}
			nichtStorniert = accumulatePositionen(nichtStorniert, data.Positionen)

		case string(EventTypeProdukteStorniertV1):
			var data produkteStorniertV1Data
			if err := json.Unmarshal(evt.Data, &data); err != nil {
				return nil, fmt.Errorf("unmarshal stornierung data: %w", err)
			}
			nichtStorniert = reduceByPosition(nichtStorniert, data.Positionen)

		case string(EventTypeZahlungRegistriertV1), string(EventTypeProdukteGeliefertV1):
			continue

		default:
			return nil, fmt.Errorf("unknown event type: %s", evt.Type)
		}
	}

	return nichtStorniert, nil
}
