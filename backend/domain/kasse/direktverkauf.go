package kasse

import (
	"encoding/json"
	"fmt"
	"time"

	e "github.com/nicograef/jotti/backend/domain/event"
)

// ComputeNichtStornierteVerkaufPositionen replays a single Direktverkauf stream to compute all
// positions that were sold but not yet cancelled. Used on-demand for stornierung validation
// (there is no projection for Direktverkauf).
func ComputeNichtStornierteVerkaufPositionen(events []e.Event) ([]Position, error) {
	var nichtStorniert []Position

	for _, evt := range events {
		switch evt.Type {
		case string(EventTypeDirektverkaufGetaetigtV1):
			var data direktverkaufGetaetigtV1Data
			if err := json.Unmarshal(evt.Data, &data); err != nil {
				return nil, fmt.Errorf("unmarshal direktverkauf getaetigt data: %w", err)
			}
			nichtStorniert = accumulatePositionen(nichtStorniert, fromPositionenEventData(data.Positionen))

		case string(EventTypeDirektverkaufStorniertV1):
			var data direktverkaufStorniertV1Data
			if err := json.Unmarshal(evt.Data, &data); err != nil {
				return nil, fmt.Errorf("unmarshal direktverkauf storniert data: %w", err)
			}
			nichtStorniert = reduceByPosition(nichtStorniert, fromPositionenEventData(data.Positionen))

		default:
			return nil, fmt.Errorf("unknown event type: %s", evt.Type)
		}
	}

	return nichtStorniert, nil
}

// DirektverkaufHistorieEintrag is the compact history of a single Direktverkauf (one row per sale),
// derived by replaying the verkauf stream. OffenePositionen are the not-yet-cancelled positions
// (the candidates for a stornierung); GesamtStorniertCents is the sum of all cancellations.
type DirektverkaufHistorieEintrag struct {
	VerkaufID            string
	UserID               int
	UserName             string
	GetaetigtAm          time.Time
	Positionen           []Position
	GesamtbetragCents    int
	Kommentar            string
	OffenePositionen     []Position
	GesamtStorniertCents int
}

// BuildDirektverkaufHistorieEintrag replays a single Direktverkauf stream (getaetigt + stornos)
// into a compact history entry.
func BuildDirektverkaufHistorieEintrag(events []e.Event) (DirektverkaufHistorieEintrag, error) {
	eintrag := DirektverkaufHistorieEintrag{}

	for _, evt := range events {
		switch evt.Type {
		case string(EventTypeDirektverkaufGetaetigtV1):
			var data direktverkaufGetaetigtV1Data
			if err := json.Unmarshal(evt.Data, &data); err != nil {
				return DirektverkaufHistorieEintrag{}, fmt.Errorf("unmarshal direktverkauf getaetigt data: %w", err)
			}
			positionen := fromPositionenEventData(data.Positionen)
			eintrag.VerkaufID = data.VerkaufID
			eintrag.UserID = evt.UserID
			eintrag.UserName = evt.UserName
			eintrag.GetaetigtAm = evt.Time
			eintrag.Positionen = positionen
			eintrag.GesamtbetragCents = data.GesamtbetragCents
			eintrag.Kommentar = data.Kommentar
			eintrag.OffenePositionen = accumulatePositionen(nil, positionen)

		case string(EventTypeDirektverkaufStorniertV1):
			var data direktverkaufStorniertV1Data
			if err := json.Unmarshal(evt.Data, &data); err != nil {
				return DirektverkaufHistorieEintrag{}, fmt.Errorf("unmarshal direktverkauf storniert data: %w", err)
			}
			eintrag.OffenePositionen = reduceByPosition(eintrag.OffenePositionen, fromPositionenEventData(data.Positionen))
			eintrag.GesamtStorniertCents += data.GesamtStornierungCents

		default:
			return DirektverkaufHistorieEintrag{}, fmt.Errorf("unknown event type: %s", evt.Type)
		}
	}

	return eintrag, nil
}
