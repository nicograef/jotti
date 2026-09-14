package kasse

import (
	"encoding/json"
	"fmt"
	"time"

	e "github.com/nicograef/jotti/backend/domain/event"
)

// TischSession is the projected state of a table session; the zero value is a session
// without events (Saldo 0, empty lists).
type TischSession struct {
	Subject                string
	TischID                int
	KassensitzungNr        int
	SaldoCents             int
	UnbezahltePositionen   []Position
	GesamtZahlungenCents   int
	ErsteBestellungLogTime *time.Time
	LastEventID            int
	LastEventVersion       int
}

func ApplyEvent(state TischSession, evt e.Event) (TischSession, error) {
	switch evt.Type {
	case string(EventTypeBestellungAufgenommenV1):
		var data BestellungAufgenommenV1Data
		if err := json.Unmarshal(evt.Data, &data); err != nil {
			return state, fmt.Errorf("unmarshal bestellung data: %w", err)
		}
		neuePositionen := tagBesteller(fromPositionenEventData(data.Positionen), evt.UserID, evt.UserName)
		state.UnbezahltePositionen = accumulatePositionen(state.UnbezahltePositionen, neuePositionen)

		setErsteBestellungLogTime(&state, evt.Time)

	case string(EventTypeZahlungKassiertV1):
		var data ZahlungKassiertV1Data
		if err := json.Unmarshal(evt.Data, &data); err != nil {
			return state, fmt.Errorf("unmarshal zahlung data: %w", err)
		}
		state.GesamtZahlungenCents += data.GesamtZahlungCents
		unbezahlt, err := reduceByPositionStrict(state.UnbezahltePositionen, fromPositionenEventData(data.Positionen))
		if err != nil {
			return state, fmt.Errorf("zahlung %s: %w", evt.Subject, err)
		}
		state.UnbezahltePositionen = unbezahlt

	case string(EventTypeStornierungErteiltV1):
		// Warenrücknahme bezahlter Positionen: der offene Betrag bleibt unverändert (sie
		// waren bereits bezahlt), die Bar-Rückgabe mindert die vereinnahmten Zahlungen.
		var data StornierungErteiltV1Data
		if err := json.Unmarshal(evt.Data, &data); err != nil {
			return state, fmt.Errorf("unmarshal stornierung data: %w", err)
		}
		state.GesamtZahlungenCents -= data.GesamtStornierungCents

	case string(EventTypeBestellungKorrigiertV1):
		var data BestellungKorrigiertV1Data
		if err := json.Unmarshal(evt.Data, &data); err != nil {
			return state, fmt.Errorf("unmarshal korrektur data: %w", err)
		}
		unbezahlt, err := reduceByPositionStrict(state.UnbezahltePositionen, fromPositionenEventData(data.Positionen))
		if err != nil {
			return state, fmt.Errorf("korrektur %s: %w", evt.Subject, err)
		}
		state.UnbezahltePositionen = unbezahlt

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
			// Abgang: die Positionen verlassen den Quelltisch.
			unbezahlt, err := reduceByPositionStrict(state.UnbezahltePositionen, positionen)
			if err != nil {
				return state, fmt.Errorf("umbuchung %s: %w", evt.Subject, err)
			}
			state.UnbezahltePositionen = unbezahlt
		} else {
			// Zugang: wie eine frische Bestellung auf dem Zieltisch.
			neuePositionen := tagBesteller(positionen, evt.UserID, evt.UserName)
			state.UnbezahltePositionen = accumulatePositionen(state.UnbezahltePositionen, neuePositionen)

			setErsteBestellungLogTime(&state, evt.Time)
		}

	default:
		return state, fmt.Errorf("unknown event type: %s", evt.Type)
	}

	// SaldoCents ist vollständig aus UnbezahltePositionen abgeleitet (Σ EinzelpreisCents ×
	// Menge) und wird deshalb hier einmal berechnet statt in jedem Arm fortgeschrieben.
	// GesamtZahlungenCents ist ein echter Akkumulator (nicht ableitbar) und wird oben je
	// Arm fortgeschrieben.
	state.SaldoCents = saldoAusPositionen(state.UnbezahltePositionen)

	state.LastEventID = evt.ID
	state.LastEventVersion = evt.Version

	return state, nil
}

func saldoAusPositionen(positionen []Position) int {
	saldo := 0
	for _, pos := range positionen {
		saldo += pos.EinzelpreisCents * pos.Menge
	}
	return saldo
}

// setErsteBestellungLogTime stempelt den Zeitpunkt der ersten Bestellung, sofern noch nicht
// gesetzt (AEAO 1.14.3: Das Aufzeichnungssystem stellt den Zeitpunkt, die TSE-Signatur
// entsteht asynchron). Eine Umbuchung auf einen leeren Zieltisch zählt als erste Bestellung.
func setErsteBestellungLogTime(state *TischSession, eventTime time.Time) {
	if state.ErsteBestellungLogTime != nil {
		return
	}

	logTime := eventTime.UTC()
	state.ErsteBestellungLogTime = &logTime
}

// ComputeNichtStorniertePositionen replays events to the positions that were ordered but
// not yet cancelled. Seed helper: its only non-test caller is the seed engine.
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

		case string(EventTypeBestellungKorrigiertV1):
			var data BestellungKorrigiertV1Data
			if err := json.Unmarshal(evt.Data, &data); err != nil {
				return nil, fmt.Errorf("unmarshal korrektur data: %w", err)
			}
			nichtStorniert = reduceByPosition(nichtStorniert, fromPositionenEventData(data.Positionen))

		case string(EventTypeBestellungUmgebuchtV1):
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

		case string(EventTypeZahlungKassiertV1):
			continue

		default:
			return nil, fmt.Errorf("unknown event type: %s", evt.Type)
		}
	}

	return nichtStorniert, nil
}

// tagBesteller stamps the ordering Servicekraft (from the event envelope) onto each freshly
// ordered position. Returns a copy — the caller's slice is not modified.
func tagBesteller(positionen []Position, userID int, userName string) []Position {
	out := make([]Position, len(positionen))
	copy(out, positionen)
	for i := range out {
		out[i].BestellerUserID = userID
		out[i].BestellerName = userName
	}
	return out
}

// accumulatePositionen adds positions to a list, merging quantities for matching positions (by PositionID).
// Works on a clone of list so the caller's backing array is never modified.
func accumulatePositionen(list []Position, positionen []Position) []Position {
	out := make([]Position, len(list))
	copy(out, list)
	for _, pos := range positionen {
		found := false
		for i, existing := range out {
			if existing.PositionID == pos.PositionID {
				out[i].Menge += pos.Menge
				found = true
				break
			}
		}
		if !found {
			out = append(out, pos)
		}
	}
	return out
}

// reduceByPosition subtracts positions from a list, removing entries at quantity zero.
// Missing positions and over-reductions are tolerated — only use it where that is legitimate
// (ComputeNichtStorniertePositionen: a position may already have been moved away).
// Works on a clone of list so the caller's backing array is never modified.
func reduceByPosition(list []Position, reductions []Position) []Position {
	out := make([]Position, len(list))
	copy(out, list)
	for _, red := range reductions {
		for i := 0; i < len(out); i++ {
			if out[i].PositionID == red.PositionID {
				if out[i].Menge > red.Menge {
					out[i].Menge -= red.Menge
				} else {
					out = append(out[:i], out[i+1:]...)
				}
				break
			}
		}
	}
	return out
}

// reduceByPositionStrict subtracts positions and fails on inconsistencies: a reduction that
// hits no position or exceeds the available Menge is the symptom of a slipped double write
// (OCC violation) and must not silently falsify the projection.
// Works on a clone of list so the caller's backing array is never modified.
func reduceByPositionStrict(list []Position, reductions []Position) ([]Position, error) {
	out := make([]Position, len(list))
	copy(out, list)
	for _, red := range reductions {
		found := false
		for i := 0; i < len(out); i++ {
			if out[i].PositionID == red.PositionID {
				if red.Menge > out[i].Menge {
					return nil, fmt.Errorf("überreduktion für position %s: %d > %d", red.PositionID, red.Menge, out[i].Menge)
				}
				if out[i].Menge > red.Menge {
					out[i].Menge -= red.Menge
				} else {
					out = append(out[:i], out[i+1:]...)
				}
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("position %s nicht in der liste", red.PositionID)
		}
	}
	return out, nil
}
