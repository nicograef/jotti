package kasse

import (
	"fmt"
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
	"GesamtbetragCents": z.Int().GTE(0).Required(),
	"Kommentar":         z.String().Max(100),
})

// DirektverkaufStorniertV1Data stores the cancelled positions as fat positions — self-contained
// for reporting, consistent with the Tisch-Storno (stornierung-erteilt:v1).
// The json-keys are stable and must not be changed (immutable events).
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
	"GesamtStornierungCents": z.Int().GTE(0).Required(),
	"Kommentar":              z.String().Min(3).Max(100).Required(),
})

// NewDirektverkaufGetaetigtEvent creates the single event for a completed Direktverkauf.
// PositionIDs are generated server-side and GesamtbetragCents is derived from the positions.
func NewDirektverkaufGetaetigtEvent(subject string, verkaufID string, userID int, userName string, positionen []Position, kommentar string) (e.Event, error) {
	// Generate PositionIDs for each position (on a copy, so the caller's slice stays untouched)
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

	if err := direktverkaufGetaetigtV1DataSchema.Validate(&data); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return e.Event{}, fmt.Errorf("direktverkauf getaetigt data validation failed: %v", issues)
	}

	return e.New(userID, userName, string(EventTypeDirektverkaufGetaetigtV1), subject, data)
}

// NewDirektverkaufStorniertEvent creates a position-precise cancellation event for a Direktverkauf.
// It stores the cancelled positions as fat positions (self-contained, like the Tisch-Storno); the
// monetary impact is carried by gesamtStornierungCents, which the command computes from the
// not-yet-cancelled positions. StornierungID is generated server-side.
func NewDirektverkaufStorniertEvent(subject string, verkaufID string, userID int, userName string, positionen []Position, gesamtStornierungCents int, kommentar string) (e.Event, error) {
	data := DirektverkaufStorniertV1Data{
		StornierungID:          uuid.New().String(),
		VerkaufID:              verkaufID,
		Positionen:             toPositionenEventData(positionen),
		GesamtStornierungCents: gesamtStornierungCents,
		Kommentar:              kommentar,
	}

	if err := direktverkaufStorniertV1DataSchema.Validate(&data); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return e.Event{}, fmt.Errorf("direktverkauf storniert data validation failed: %v", issues)
	}

	return e.New(userID, userName, string(EventTypeDirektverkaufStorniertV1), subject, data)
}
