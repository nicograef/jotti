package kasse

import (
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
	switch EventType(evt.Type) {
	case EventTypeBestellungAufgenommenV1:
		data, err := e.Decode[BestellungAufgenommenV1Data](evt)
		if err != nil {
			return state, err
		}
		neuePositionen := tagBesteller(PositionenFromEventData(data.Positionen), evt.UserID, evt.UserName)
		state.UnbezahltePositionen = accumulatePositionen(state.UnbezahltePositionen, neuePositionen)

		setErsteBestellungLogTime(&state, evt.Time)

	case EventTypeZahlungKassiertV1, EventTypeZahlungKassiertV2:
		data, err := e.Decode[ZahlungKassiertV1Data](evt)
		if err != nil {
			return state, err
		}
		state.GesamtZahlungenCents += data.GesamtZahlungCents
		unbezahlt, err := reduceByPositionStrict(state.UnbezahltePositionen, PositionenFromEventData(data.Positionen))
		if err != nil {
			return state, fmt.Errorf("zahlung %s: %w", evt.Subject, err)
		}
		state.UnbezahltePositionen = unbezahlt

	case EventTypeStornierungErteiltV1:
		// Warenrücknahme bezahlter Positionen: der offene Betrag bleibt unverändert (sie
		// waren bereits bezahlt), die Bar-Rückgabe mindert die vereinnahmten Zahlungen.
		data, err := e.Decode[StornierungErteiltV1Data](evt)
		if err != nil {
			return state, err
		}
		state.GesamtZahlungenCents -= data.GesamtStornierungCents

	case EventTypeBestellungKorrigiertV1:
		data, err := e.Decode[BestellungKorrigiertV1Data](evt)
		if err != nil {
			return state, err
		}
		unbezahlt, err := reduceByPositionStrict(state.UnbezahltePositionen, PositionenFromEventData(data.Positionen))
		if err != nil {
			return state, fmt.Errorf("korrektur %s: %w", evt.Subject, err)
		}
		state.UnbezahltePositionen = unbezahlt

	case EventTypeBestellungUmgebuchtV1:
		data, err := e.Decode[BestellungUmgebuchtV1Data](evt)
		if err != nil {
			return state, err
		}
		tischID, err := ParseTischIDFromSubject(evt.Subject)
		if err != nil {
			return state, err
		}
		positionen := PositionenFromEventData(data.Positionen)
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

	case EventTypeKassensitzungEroeffnetV1, EventTypeGeldtransitGebuchtV1, EventTypeKassensturzDurchgefuehrtV1,
		EventTypeDifferenzSollIstGebuchtV1, EventTypeTagesabschlussErstelltV1,
		EventTypeDirektverkaufGetaetigtV1, EventTypeDirektverkaufGetaetigtV2, EventTypeDirektverkaufStorniertV1,
		EventTypePfandZurueckgegebenV1, EventTypePfandrueckgabeStorniertV1:
		return state, fmt.Errorf("event type %s does not belong to a tisch session", evt.Type)

	default:
		return state, fmt.Errorf("unknown event type: %s", evt.Type)
	}

	// SaldoCents derives from UnbezahltePositionen, so it is computed once here; GesamtZahlungenCents
	// is a true accumulator, updated per arm above (docs/handbuch.md §3.8).
	state.SaldoCents = summeCents(state.UnbezahltePositionen)

	state.LastEventID = evt.ID
	state.LastEventVersion = evt.Version

	return state, nil
}

// ErsteBestellung findet das Event, das in der Projektion ErsteBestellungLogTime setzt: die erste
// Bestellung des Streams oder einen Umbuchungs-Zugang davor. Der Zahlungsbeleg druckt den TSE-Start
// seiner Signatur (DSFinV-K Tz. 2.7.2).
func ErsteBestellung(events []e.Event) (e.Event, bool, error) {
	for _, evt := range events {
		if EventType(evt.Type) == EventTypeBestellungAufgenommenV1 {
			return evt, true, nil
		}
		if EventType(evt.Type) != EventTypeBestellungUmgebuchtV1 {
			continue
		}
		data, err := e.Decode[BestellungUmgebuchtV1Data](evt)
		if err != nil {
			return e.Event{}, false, err
		}
		tischID, err := ParseTischIDFromSubject(evt.Subject)
		if err != nil {
			return e.Event{}, false, err
		}
		if tischID == data.ZielTischID {
			return evt, true, nil
		}
	}
	return e.Event{}, false, nil
}

// setErsteBestellungLogTime stempelt die Event-Zeit der ersten Bestellung, sofern noch nicht gesetzt; eine
// Umbuchung auf einen leeren Zieltisch zählt als erste Bestellung (siehe ErsteBestellung).
func setErsteBestellungLogTime(state *TischSession, eventTime time.Time) {
	if state.ErsteBestellungLogTime != nil {
		return
	}

	logTime := eventTime.UTC()
	state.ErsteBestellungLogTime = &logTime
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

// reduceByPosition subtracts positions from a clone of list, dropping entries at quantity zero.
// It tolerates missing positions and over-reductions, so use it only where that is legitimate.
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

// reduceByPositionStrict subtracts positions from a clone of list and fails on inconsistencies:
// a reduction that hits no position or exceeds the available Menge is the symptom of a slipped
// double write (OCC violation) and must not silently falsify the projection.
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
