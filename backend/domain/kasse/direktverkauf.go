package kasse

import (
	"encoding/json"
	"fmt"
	"slices"
	"time"

	e "github.com/nicograef/jotti/backend/domain/event"
)

// VorgangHistorieEintrag holds what Direktverkauf and Pfandrückgabe entries share.
// OffenePositionen are the not-yet-cancelled positions — the candidates for a stornierung.
type VorgangHistorieEintrag struct {
	UserID               int
	UserName             string
	GetaetigtAm          time.Time
	Positionen           []Position
	GesamtbetragCents    int
	Kommentar            string
	OffenePositionen     []Position
	GesamtStorniertCents int
	Stornierungen        []VorgangStornierung
}

// VorgangStornierung is one cancellation within a Vorgang; its StornierungID identifies the
// Stornobeleg that can be printed for it.
type VorgangStornierung struct {
	StornierungID          string
	StorniertAm            time.Time
	GesamtStornierungCents int
}

// DirektverkaufHistorieEintrag is one Direktverkauf replayed from its stream.
type DirektverkaufHistorieEintrag struct {
	VorgangHistorieEintrag
	VerkaufID string
}

type vorgangKern struct {
	Positionen        []PositionEventData
	GesamtbetragCents int
	Kommentar         string
}

type stornoKern struct {
	StornierungID          string
	Positionen             []PositionEventData
	GesamtStornierungCents int
}

// replayVorgang replays one stream of a Vorgang (V, any of vorgangTypen) and its cancellations (S);
// it returns the Vorgang's data so the caller reads its own ID fields.
func replayVorgang[V any, S any](events []e.Event, vorgangTypen []EventType, stornoTyp EventType, vorgang func(V) vorgangKern, storno func(S) stornoKern) (VorgangHistorieEintrag, V, error) {
	eintrag := VorgangHistorieEintrag{}
	var vorgangData V

	for _, evt := range events {
		switch {
		case slices.Contains(vorgangTypen, EventType(evt.Type)):
			if err := json.Unmarshal(evt.Data, &vorgangData); err != nil {
				return VorgangHistorieEintrag{}, vorgangData, fmt.Errorf("unmarshal %s data: %w", evt.Type, err)
			}
			kern := vorgang(vorgangData)
			positionen := fromPositionenEventData(kern.Positionen)
			eintrag.UserID = evt.UserID
			eintrag.UserName = evt.UserName
			eintrag.GetaetigtAm = evt.Time
			eintrag.Positionen = positionen
			eintrag.GesamtbetragCents = kern.GesamtbetragCents
			eintrag.Kommentar = kern.Kommentar
			eintrag.OffenePositionen = accumulatePositionen(nil, positionen)

		case evt.Type == string(stornoTyp):
			var data S
			if err := json.Unmarshal(evt.Data, &data); err != nil {
				return VorgangHistorieEintrag{}, vorgangData, fmt.Errorf("unmarshal %s data: %w", stornoTyp, err)
			}
			kern := storno(data)
			eintrag.OffenePositionen = reduceByPosition(eintrag.OffenePositionen, fromPositionenEventData(kern.Positionen))
			eintrag.GesamtStorniertCents += kern.GesamtStornierungCents
			eintrag.Stornierungen = append(eintrag.Stornierungen, VorgangStornierung{
				StornierungID:          kern.StornierungID,
				StorniertAm:            evt.Time,
				GesamtStornierungCents: kern.GesamtStornierungCents,
			})

		default:
			return VorgangHistorieEintrag{}, vorgangData, fmt.Errorf("unknown event type: %s", evt.Type)
		}
	}

	return eintrag, vorgangData, nil
}

func BuildDirektverkaufHistorieEintrag(events []e.Event) (DirektverkaufHistorieEintrag, error) {
	eintrag, data, err := replayVorgang(events, []EventType{EventTypeDirektverkaufGetaetigtV1, EventTypeDirektverkaufGetaetigtV2}, EventTypeDirektverkaufStorniertV1,
		func(d DirektverkaufGetaetigtV1Data) vorgangKern {
			return vorgangKern{Positionen: d.Positionen, GesamtbetragCents: d.GesamtbetragCents, Kommentar: d.Kommentar}
		},
		func(d DirektverkaufStorniertV1Data) stornoKern {
			return stornoKern{StornierungID: d.StornierungID, Positionen: d.Positionen, GesamtStornierungCents: d.GesamtStornierungCents}
		},
	)
	if err != nil {
		return DirektverkaufHistorieEintrag{}, err
	}
	return DirektverkaufHistorieEintrag{VorgangHistorieEintrag: eintrag, VerkaufID: data.VerkaufID}, nil
}
