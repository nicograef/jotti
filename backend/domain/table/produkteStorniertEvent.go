package table

import (
	"fmt"
	"strconv"

	z "github.com/Oudwins/zog"
	"github.com/google/uuid"
	e "github.com/nicograef/jotti/backend/domain/event"
)

type produkteStorniertV1Data struct {
	StornierungID          string     `json:"stornierungId"`
	Positionen             []Position `json:"positionen"`
	GesamtStornierungCents int        `json:"gesamtStornierungCents"`
	Comment                string     `json:"comment"`
}

var produkteStorniertV1DataSchema = z.Struct(z.Shape{
	"StornierungID":          z.String().UUID().Required(),
	"Positionen":             z.Slice(positionSchema).Min(1).Required(),
	"GesamtStornierungCents": z.Int().GTE(0).Required(),
	"Comment":                z.String().Max(100),
})

func NewProdukteStorniertEvent(userID, tischID int, positionen []Position, comment string) (e.Event, error) {
	gesamtStornierungCents := 0
	for _, pos := range positionen {
		gesamtStornierungCents += pos.PreisCents * pos.Quantity
	}

	data := produkteStorniertV1Data{
		StornierungID:          uuid.New().String(),
		Positionen:             positionen,
		GesamtStornierungCents: gesamtStornierungCents,
		Comment:                comment,
	}

	if err := produkteStorniertV1DataSchema.Validate(&data); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return e.Event{}, fmt.Errorf("produkte storniert data validation failed: %v", issues)
	}

	event, err := e.New(userID, string(EventTypeProdukteStorniertV1), "tisch:"+strconv.Itoa(tischID), data)
	if err != nil {
		return e.Event{}, err
	}

	return event, nil
}

func buildStornierungFromEvent(event e.Event) (Stornierung, error) {
	if event.Type != string(EventTypeProdukteStorniertV1) {
		return Stornierung{}, fmt.Errorf("unsupported event type: %s", event.Type)
	}

	tischID, err := parseTischIDFromSubject(event.Subject)
	if err != nil {
		return Stornierung{}, err
	}

	data := produkteStorniertV1Data{}
	err = e.ParseData(event, &data, produkteStorniertV1DataSchema)
	if err != nil {
		return Stornierung{}, err
	}

	stornierung := Stornierung{
		ID:                     data.StornierungID,
		UserID:                 event.UserID,
		TischID:                tischID,
		Positionen:             data.Positionen,
		GesamtStornierungCents: data.GesamtStornierungCents,
		Comment:                data.Comment,
		StorniertAm:            event.Time,
	}

	if err := stornierungSchema.Validate(&stornierung); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return Stornierung{}, fmt.Errorf("stornierung validation failed: %v", issues)
	}

	return stornierung, nil
}
