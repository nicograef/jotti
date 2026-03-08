package table

import (
	"fmt"
	"strconv"

	z "github.com/Oudwins/zog"
	"github.com/google/uuid"
	e "github.com/nicograef/jotti/backend/domain/event"
)

type bestellungAufgegebenV1Data struct {
	BestellungID     string     `json:"bestellungId"`
	Positionen       []Position `json:"positionen"`
	GesamtPreisCents int        `json:"gesamtPreisCents"`
	Comment          string     `json:"comment"`
}

var bestellungAufgegebenV1DataSchema = z.Struct(z.Shape{
	"BestellungID":     z.String().UUID().Required(),
	"Positionen":       z.Slice(positionSchema).Min(1).Required(),
	"GesamtPreisCents": z.Int().GTE(0).Required(),
	"Comment":          z.String().Max(100),
})

func NewBestellungAufgegebenEvent(userID, tischID int, positionen []Position, comment string) (e.Event, error) {
	gesamtPreisCents := 0
	for _, pos := range positionen {
		gesamtPreisCents += pos.PreisCents * pos.Quantity
	}

	data := bestellungAufgegebenV1Data{
		BestellungID:     uuid.New().String(),
		Positionen:       positionen,
		GesamtPreisCents: gesamtPreisCents,
		Comment:          comment,
	}

	if err := bestellungAufgegebenV1DataSchema.Validate(&data); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return e.Event{}, fmt.Errorf("bestellung aufgegeben data validation failed: %v", issues)
	}

	event, err := e.New(userID, string(EventTypeBestellungAufgegebenV1), "tisch:"+strconv.Itoa(tischID), data)
	if err != nil {
		return e.Event{}, err
	}

	return event, nil
}

func buildBestellungFromEvent(event e.Event) (Bestellung, error) {
	if event.Type != string(EventTypeBestellungAufgegebenV1) {
		return Bestellung{}, fmt.Errorf("unsupported event type: %s", event.Type)
	}

	tischID, err := parseTischIDFromSubject(event.Subject)
	if err != nil {
		return Bestellung{}, err
	}

	data := bestellungAufgegebenV1Data{}
	err = e.ParseData(event, &data, bestellungAufgegebenV1DataSchema)
	if err != nil {
		return Bestellung{}, err
	}

	bestellung := Bestellung{
		ID:               data.BestellungID,
		UserID:           event.UserID,
		TischID:          tischID,
		Positionen:       data.Positionen,
		GesamtPreisCents: data.GesamtPreisCents,
		Comment:          data.Comment,
		AufgegebenAm:     event.Time,
	}

	if err := bestellungSchema.Validate(&bestellung); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return Bestellung{}, fmt.Errorf("bestellung validation failed: %v", issues)
	}

	return bestellung, nil
}
