package kasse

import (
	"slices"

	z "github.com/Oudwins/zog"
	"github.com/google/uuid"
	e "github.com/nicograef/jotti/backend/domain/event"
)

const (
	EventTypeDirektverkaufGetaetigtV1 EventType = "direktverkauf-getaetigt:v1"
	EventTypeDirektverkaufStorniertV1 EventType = "direktverkauf-storniert:v1"
)

type DirektverkaufGetaetigtV1Data struct {
	VerkaufID         string              `json:"verkaufId"`
	Positionen        []PositionEventData `json:"positionen"`
	GesamtbetragCents int                 `json:"gesamtbetragCents"`
	Kommentar         string              `json:"kommentar"`
}

var direktverkaufGetaetigtV1DataSchema = z.Struct(z.Shape{
	"VerkaufID":         z.String().UUID().Required(),
	"Positionen":        z.Slice(positionSchema).Min(1).Required(),
	"GesamtbetragCents": z.Int().GTE(1).Required(),
	"Kommentar":         z.String().Max(100),
})

// DirektverkaufStorniertV1Data stores the cancelled positions as fat positions
// (self-contained for reporting). The json keys are frozen — immutable events.
type DirektverkaufStorniertV1Data struct {
	StornierungID          string              `json:"stornierungId"`
	VerkaufID              string              `json:"verkaufId"`
	Positionen             []PositionEventData `json:"positionen"`
	GesamtStornierungCents int                 `json:"gesamtStornierungCents"`
	Kommentar              string              `json:"kommentar"`
}

var direktverkaufStorniertV1DataSchema = z.Struct(z.Shape{
	"StornierungID":          z.String().UUID().Required(),
	"VerkaufID":              z.String().UUID().Required(),
	"Positionen":             z.Slice(positionSchema).Min(1).Required(),
	"GesamtStornierungCents": z.Int().GTE(1).Required(),
	"Kommentar":              z.String().Min(3).Max(100).Required(),
})

func NewDirektverkaufGetaetigtEvent(subject string, verkaufID string, userID int, userName string, positionen []Position, kommentar string) (e.Event, error) {
	// On a copy, so the caller's slice stays untouched.
	positionen = slices.Clone(positionen)
	for i := range positionen {
		positionen[i].PositionID = uuid.New().String()
	}

	gesamtbetragCents := 0
	for _, pos := range positionen {
		gesamtbetragCents += pos.EinzelpreisCents * pos.Menge
	}

	data := DirektverkaufGetaetigtV1Data{
		VerkaufID:         verkaufID,
		Positionen:        toPositionenEventData(positionen),
		GesamtbetragCents: gesamtbetragCents,
		Kommentar:         kommentar,
	}

	if err := validateEventData(direktverkaufGetaetigtV1DataSchema, &data, "direktverkauf getaetigt"); err != nil {
		return e.Event{}, err
	}

	return e.New(userID, userName, string(EventTypeDirektverkaufGetaetigtV1), subject, data)
}

// NewDirektverkaufStorniertEvent: gesamtStornierungCents is computed by the caller from
// the not-yet-cancelled positions.
func NewDirektverkaufStorniertEvent(subject string, verkaufID string, userID int, userName string, positionen []Position, gesamtStornierungCents int, kommentar string) (e.Event, error) {
	data := DirektverkaufStorniertV1Data{
		StornierungID:          uuid.New().String(),
		VerkaufID:              verkaufID,
		Positionen:             toPositionenEventData(positionen),
		GesamtStornierungCents: gesamtStornierungCents,
		Kommentar:              kommentar,
	}

	if err := validateEventData(direktverkaufStorniertV1DataSchema, &data, "direktverkauf storniert"); err != nil {
		return e.Event{}, err
	}

	return e.New(userID, userName, string(EventTypeDirektverkaufStorniertV1), subject, data)
}
