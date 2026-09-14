package kasse

import (
	"encoding/json"
	"fmt"
	"time"

	e "github.com/nicograef/jotti/backend/domain/event"
)

// ComputeNichtStornierteVerkaufPositionen replays a single Direktverkauf stream on demand
// (there is no projection for Direktverkauf) to the positions sold but not yet cancelled.
func ComputeNichtStornierteVerkaufPositionen(events []e.Event) ([]Position, error) {
	var nichtStorniert []Position

	for _, evt := range events {
		switch evt.Type {
		case string(EventTypeDirektverkaufGetaetigtV1):
			var data DirektverkaufGetaetigtV1Data
			if err := json.Unmarshal(evt.Data, &data); err != nil {
				return nil, fmt.Errorf("unmarshal direktverkauf getaetigt data: %w", err)
			}
			nichtStorniert = accumulatePositionen(nichtStorniert, fromPositionenEventData(data.Positionen))

		case string(EventTypeDirektverkaufStorniertV1):
			var data DirektverkaufStorniertV1Data
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

// DirektverkaufHistorieEintrag is one Direktverkauf replayed from its stream.
// OffenePositionen are the not-yet-cancelled positions — the candidates for a stornierung.
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
	Stornierungen        []DirektverkaufStornierung
}

// DirektverkaufStornierung is one cancellation within a Direktverkauf; its StornierungID
// identifies the Stornobeleg that can be printed for it.
type DirektverkaufStornierung struct {
	StornierungID          string
	StorniertAm            time.Time
	GesamtStornierungCents int
}

func BuildDirektverkaufHistorieEintrag(events []e.Event) (DirektverkaufHistorieEintrag, error) {
	eintrag := DirektverkaufHistorieEintrag{}

	for _, evt := range events {
		switch evt.Type {
		case string(EventTypeDirektverkaufGetaetigtV1):
			var data DirektverkaufGetaetigtV1Data
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
			var data DirektverkaufStorniertV1Data
			if err := json.Unmarshal(evt.Data, &data); err != nil {
				return DirektverkaufHistorieEintrag{}, fmt.Errorf("unmarshal direktverkauf storniert data: %w", err)
			}
			eintrag.OffenePositionen = reduceByPosition(eintrag.OffenePositionen, fromPositionenEventData(data.Positionen))
			eintrag.GesamtStorniertCents += data.GesamtStornierungCents
			eintrag.Stornierungen = append(eintrag.Stornierungen, DirektverkaufStornierung{
				StornierungID:          data.StornierungID,
				StorniertAm:            evt.Time,
				GesamtStornierungCents: data.GesamtStornierungCents,
			})

		default:
			return DirektverkaufHistorieEintrag{}, fmt.Errorf("unknown event type: %s", evt.Type)
		}
	}

	return eintrag, nil
}
