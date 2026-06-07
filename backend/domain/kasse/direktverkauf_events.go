package kasse

import (
	"fmt"

	z "github.com/Oudwins/zog"
	"github.com/google/uuid"
	e "github.com/nicograef/jotti/backend/domain/event"
)

const (
	EventTypeDirektverkaufGetaetigtV1 EventType = "direktverkauf-getaetigt:v1"
)

type direktverkaufGetaetigtV1Data struct {
	VerkaufID         string              `json:"verkaufId"`
	Positionen        []positionEventData `json:"positionen"`
	GesamtbetragCents int                 `json:"gesamtbetragCents"`
	Kommentar         string              `json:"kommentar"`
}

var direktverkaufGetaetigtV1DataSchema = z.Struct(z.Shape{
	"VerkaufID":         z.String().UUID().Required(),
	"Positionen":        z.Slice(positionSchema).Min(1).Required(),
	"GesamtbetragCents": z.Int().GTE(0).Required(),
	"Kommentar":         z.String().Max(100),
})

// NewDirektverkaufGetaetigtEvent creates the single event for a completed Direktverkauf.
// PositionIDs are generated server-side and GesamtbetragCents is derived from the positions.
func NewDirektverkaufGetaetigtEvent(subject string, verkaufID string, userID int, userName string, positionen []Position, kommentar string) (e.Event, error) {
	for i := range positionen {
		positionen[i].PositionID = uuid.New().String()
	}

	gesamtbetragCents := 0
	for _, pos := range positionen {
		gesamtbetragCents += pos.Einzelpreis * pos.Menge
	}

	data := direktverkaufGetaetigtV1Data{
		VerkaufID:         verkaufID,
		Positionen:        toPositionenEventData(positionen),
		GesamtbetragCents: gesamtbetragCents,
		Kommentar:         kommentar,
	}

	if err := direktverkaufGetaetigtV1DataSchema.Validate(&data); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return e.Event{}, fmt.Errorf("direktverkauf getaetigt data validation failed: %v", issues)
	}

	event, err := e.New(userID, userName, string(EventTypeDirektverkaufGetaetigtV1), subject, data)
	if err != nil {
		return e.Event{}, err
	}

	return event, nil
}
