package table

import (
	"fmt"
	"strconv"

	z "github.com/Oudwins/zog"
	"github.com/google/uuid"
	e "github.com/nicograef/jotti/backend/domain/event"
)

type stornierungErteiltV1Data struct {
	StornierungID          string     `json:"stornierungId"`
	Positionen             []Position `json:"positionen"`
	GesamtStornierungCents int        `json:"gesamtStornierungCents"`
	Kommentar              string     `json:"kommentar"`
}

var stornierungErteiltV1DataSchema = z.Struct(z.Shape{
	"StornierungID":          z.String().UUID().Required(),
	"Positionen":             z.Slice(positionSchema).Min(1).Required(),
	"GesamtStornierungCents": z.Int().GTE(0).Required(),
	"Kommentar":              z.String().Max(100),
})

func NewStornierungErteiltEvent(userID int, userName string, tischID int, positionen []Position, gesamtStornierungCents int, kommentar string) (e.Event, error) {
	data := stornierungErteiltV1Data{
		StornierungID:          uuid.New().String(),
		Positionen:             positionen,
		GesamtStornierungCents: gesamtStornierungCents,
		Kommentar:              kommentar,
	}

	if err := stornierungErteiltV1DataSchema.Validate(&data); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return e.Event{}, fmt.Errorf("stornierung erteilt data validation failed: %v", issues)
	}

	event, err := e.New(userID, userName, string(EventTypeStornierungErteiltV1), "tisch:"+strconv.Itoa(tischID), data)
	if err != nil {
		return e.Event{}, err
	}

	return event, nil
}

func buildStornierungFromEvent(event e.Event) (Stornierung, error) {
	if event.Type != string(EventTypeStornierungErteiltV1) {
		return Stornierung{}, fmt.Errorf("unsupported event type: %s", event.Type)
	}

	tischID, err := parseTischIDFromSubject(event.Subject)
	if err != nil {
		return Stornierung{}, err
	}

	data := stornierungErteiltV1Data{}
	err = e.ParseData(event, &data, stornierungErteiltV1DataSchema)
	if err != nil {
		return Stornierung{}, err
	}

	stornierung := Stornierung{
		ID:                     data.StornierungID,
		UserID:                 event.UserID,
		TischID:                tischID,
		Positionen:             data.Positionen,
		GesamtStornierungCents: data.GesamtStornierungCents,
		Kommentar:              data.Kommentar,
		StorniertAm:            event.Time,
	}

	if err := stornierungSchema.Validate(&stornierung); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return Stornierung{}, fmt.Errorf("stornierung validation failed: %v", issues)
	}

	return stornierung, nil
}
