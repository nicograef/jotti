package kasse

import (
	z "github.com/Oudwins/zog"
	e "github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/internal/zeichen"
)

const (
	EventTypeDirektverkaufGetaetigtV1 EventType = "direktverkauf-getaetigt:v1"
	EventTypeDirektverkaufGetaetigtV2 EventType = "direktverkauf-getaetigt:v2"
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
	"Kommentar":         z.String().TestFunc(zeichen.Max(100)),
})

// DirektverkaufGetaetigtV2Data is v1 plus the Pfand flag per position, for the DSFinV-K GV_TYP. A
// reader that ignores the flag reads both versions as DirektverkaufGetaetigtV1Data.
type DirektverkaufGetaetigtV2Data struct {
	VerkaufID         string                `json:"verkaufId"`
	Positionen        []PositionV2EventData `json:"positionen"`
	GesamtbetragCents int                   `json:"gesamtbetragCents"`
	Kommentar         string                `json:"kommentar"`
}

func (d DirektverkaufGetaetigtV2Data) v1() DirektverkaufGetaetigtV1Data {
	return DirektverkaufGetaetigtV1Data{
		VerkaufID:         d.VerkaufID,
		Positionen:        positionenV1(d.Positionen),
		GesamtbetragCents: d.GesamtbetragCents,
		Kommentar:         d.Kommentar,
	}
}

// IstDirektverkaufGetaetigt reports direktverkauf-getaetigt in any version.
func IstDirektverkaufGetaetigt(eventType string) bool {
	return eventType == string(EventTypeDirektverkaufGetaetigtV1) || eventType == string(EventTypeDirektverkaufGetaetigtV2)
}

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
	"Kommentar":              z.String().TestFunc(zeichen.Min(3)).TestFunc(zeichen.Max(100)).Required(),
})

func NewDirektverkaufGetaetigtEvent(subject string, verkaufID string, userID int, userName string, positionen []Position, kommentar string) (e.Event, error) {
	positionen = mitNeuenPositionIDs(positionen)
	gesamtbetragCents := summeCents(positionen)

	data := DirektverkaufGetaetigtV2Data{
		VerkaufID:         verkaufID,
		Positionen:        toPositionenV2EventData(positionen),
		GesamtbetragCents: gesamtbetragCents,
		Kommentar:         kommentar,
	}

	basis := data.v1()
	if err := validateEventData(direktverkaufGetaetigtV1DataSchema, &basis, "direktverkauf getaetigt"); err != nil {
		return e.Event{}, err
	}

	return e.New(userID, userName, string(EventTypeDirektverkaufGetaetigtV2), subject, data)
}

// NewDirektverkaufStorniertEvent: gesamtStornierungCents is computed by the caller from
// the not-yet-cancelled positions.
// NewDirektverkaufStorniertEvent takes stornierungID from the client: it is the idempotency key.
func NewDirektverkaufStorniertEvent(subject string, stornierungID string, verkaufID string, userID int, userName string, positionen []Position, gesamtStornierungCents int, kommentar string) (e.Event, error) {
	data := DirektverkaufStorniertV1Data{
		StornierungID:          stornierungID,
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
