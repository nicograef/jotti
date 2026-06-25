package kasse

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	e "github.com/nicograef/jotti/backend/domain/event"
)

// TischSession represents the projected state of a table session, derived from applying events.
// Zero-value represents a table session with no events (Saldo 0, empty lists).
type TischSession struct {
	Subject                string
	TischID                int
	KassensitzungNr        int
	SaldoCents             int
	UnbezahltePositionen   []Position
	AusstehendePositionen  []Position
	GesamtZahlungenCents   int
	ErsteBestellungLogTime *time.Time
	LastEventID            int
	LastEventVersion       int
}

// ApplyEvent applies a single domain event to the current TischSession and returns the new state.
func ApplyEvent(state TischSession, evt e.Event) (TischSession, error) {
	switch evt.Type {
	case string(EventTypeBestellungAufgenommenV1):
		var data BestellungAufgenommenV1Data
		if err := json.Unmarshal(evt.Data, &data); err != nil {
			return state, fmt.Errorf("unmarshal bestellung data: %w", err)
		}
		state.SaldoCents += data.GesamtPreisCents
		neuePositionen := tagBesteller(fromPositionenEventData(data.Positionen), evt.UserID, evt.UserName)
		state.UnbezahltePositionen = accumulatePositionen(state.UnbezahltePositionen, neuePositionen)
		state.AusstehendePositionen = accumulatePositionen(state.AusstehendePositionen, neuePositionen)

		if err := setErsteBestellungLogTime(&state, data.TSEData, evt.Time); err != nil {
			return state, err
		}

	case string(EventTypeZahlungKassiertV1):
		var data ZahlungKassiertV1Data
		if err := json.Unmarshal(evt.Data, &data); err != nil {
			return state, fmt.Errorf("unmarshal zahlung data: %w", err)
		}
		state.SaldoCents -= data.GesamtZahlungCents
		state.GesamtZahlungenCents += data.GesamtZahlungCents
		state.UnbezahltePositionen = reduceByPosition(state.UnbezahltePositionen, fromPositionenEventData(data.Positionen))

	case string(EventTypeStornierungErteiltV1):
		var data StornierungErteiltV1Data
		if err := json.Unmarshal(evt.Data, &data); err != nil {
			return state, fmt.Errorf("unmarshal stornierung data: %w", err)
		}
		state.SaldoCents -= data.GesamtStornierungCents
		state.UnbezahltePositionen = reduceByPosition(state.UnbezahltePositionen, fromPositionenEventData(data.Positionen))
		state.AusstehendePositionen = reduceByPosition(state.AusstehendePositionen, fromPositionenEventData(data.Positionen))

	case string(EventTypeBestellungUmgebuchtV1):
		var data BestellungUmgebuchtV1Data
		if err := json.Unmarshal(evt.Data, &data); err != nil {
			return state, fmt.Errorf("unmarshal umbuchung data: %w", err)
		}
		tischID, err := ParseTischIDFromSubject(evt.Subject)
		if err != nil {
			return state, err
		}
		positionen := fromPositionenEventData(data.Positionen)
		if tischID == data.QuellTischID {
			// Abgang: die Positionen verlassen den Quelltisch (geldneutral je System,
			// der Quell-Saldo sinkt).
			state.SaldoCents -= data.GesamtCents
			state.UnbezahltePositionen = reduceByPosition(state.UnbezahltePositionen, positionen)
			state.AusstehendePositionen = reduceByPosition(state.AusstehendePositionen, positionen)
		} else {
			// Zugang: die Positionen kommen auf den Zieltisch, wie eine frische
			// Bestellung (Saldo steigt, Positionen sind ausstehend und unbezahlt).
			neuePositionen := tagBesteller(positionen, evt.UserID, evt.UserName)
			state.SaldoCents += data.GesamtCents
			state.UnbezahltePositionen = accumulatePositionen(state.UnbezahltePositionen, neuePositionen)
			state.AusstehendePositionen = accumulatePositionen(state.AusstehendePositionen, neuePositionen)

			if err := setErsteBestellungLogTime(&state, data.TSEData, evt.Time); err != nil {
				return state, err
			}
		}

	case string(EventTypeAusgabeBestaetigtV1):
		var data AusgabeBestaetigtV1Data
		if err := json.Unmarshal(evt.Data, &data); err != nil {
			return state, fmt.Errorf("unmarshal ausgabe data: %w", err)
		}
		state.AusstehendePositionen = reduceByPosition(state.AusstehendePositionen, fromPositionenEventData(data.Positionen))

	case string(EventTypeAuszahlungGeleistetV1):
		var data AuszahlungGeleistetV1Data
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

// setErsteBestellungLogTime stempelt den Zeitpunkt der ersten Bestellung auf den
// Tisch, sofern noch nicht gesetzt. AEAO 1.14.3: Liegt keine TSE-logTime vor (z. B.
// TSE-Ausfall bei der ersten Bestellung), stellt das Aufzeichnungssystem den
// Zeitpunkt — Fallback auf die Event-Zeit. Eine Umbuchung auf einen leeren Zieltisch
// zählt dabei wie eine erste Bestellung.
func setErsteBestellungLogTime(state *TischSession, tseData *TSEData, eventTime time.Time) error {
	if state.ErsteBestellungLogTime != nil {
		return nil
	}

	logTime := eventTime.UTC()
	if tseData != nil && strings.TrimSpace(tseData.LogTimeStart) != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(tseData.LogTimeStart))
		if err != nil {
			return fmt.Errorf("parse erste bestellung log time: %w", err)
		}
		logTime = parsed.UTC()
	}
	state.ErsteBestellungLogTime = &logTime
	return nil
}

// ComputeNichtStorniertePositionen replays events to compute all positions that were ordered
// but not yet cancelled. Used on-demand for stornierung validation.
func ComputeNichtStorniertePositionen(events []e.Event) ([]Position, error) {
	var nichtStorniert []Position

	for _, evt := range events {
		switch evt.Type {
		case string(EventTypeBestellungAufgenommenV1):
			var data BestellungAufgenommenV1Data
			if err := json.Unmarshal(evt.Data, &data); err != nil {
				return nil, fmt.Errorf("unmarshal bestellung data: %w", err)
			}
			nichtStorniert = accumulatePositionen(nichtStorniert, fromPositionenEventData(data.Positionen))

		case string(EventTypeStornierungErteiltV1):
			var data StornierungErteiltV1Data
			if err := json.Unmarshal(evt.Data, &data); err != nil {
				return nil, fmt.Errorf("unmarshal stornierung data: %w", err)
			}
			nichtStorniert = reduceByPosition(nichtStorniert, fromPositionenEventData(data.Positionen))

		case string(EventTypeBestellungUmgebuchtV1):
			// Abgang entfernt die umgebuchten Positionen vom Quelltisch, Zugang fügt
			// sie dem Zieltisch hinzu — wie eine Bestellung. Welche Seite hier vorliegt,
			// folgt aus dem Tisch des Subjects.
			var data BestellungUmgebuchtV1Data
			if err := json.Unmarshal(evt.Data, &data); err != nil {
				return nil, fmt.Errorf("unmarshal umbuchung data: %w", err)
			}
			tischID, err := ParseTischIDFromSubject(evt.Subject)
			if err != nil {
				return nil, err
			}
			positionen := fromPositionenEventData(data.Positionen)
			if tischID == data.QuellTischID {
				nichtStorniert = reduceByPosition(nichtStorniert, positionen)
			} else {
				nichtStorniert = accumulatePositionen(nichtStorniert, positionen)
			}

		case string(EventTypeZahlungKassiertV1), string(EventTypeAusgabeBestaetigtV1), string(EventTypeAuszahlungGeleistetV1):
			continue

		default:
			return nil, fmt.Errorf("unknown event type: %s", evt.Type)
		}
	}

	return nichtStorniert, nil
}

// tagBesteller stamps the ordering Servicekraft (from the event envelope) onto
// each freshly ordered position. Payment/cancellation/delivery keep the tag via
// the position ID in reduceByPosition.
func tagBesteller(positionen []Position, userID int, userName string) []Position {
	for i := range positionen {
		positionen[i].BestellerUserID = userID
		positionen[i].BestellerName = userName
	}
	return positionen
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
				}
				break
			}
		}
	}
	return list
}
