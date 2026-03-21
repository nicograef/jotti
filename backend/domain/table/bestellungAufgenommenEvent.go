package table

import (
	"fmt"
	"strconv"

	z "github.com/Oudwins/zog"
	"github.com/google/uuid"
	e "github.com/nicograef/jotti/backend/domain/event"
)

type bestellungAufgenommenV1Data struct {
	BestellungID     string              `json:"bestellungId"`
	Positionen       []positionEventData `json:"positionen"`
	GesamtPreisCents int                 `json:"gesamtPreisCents"`
	Kommentar        string              `json:"kommentar"`
}

var bestellungAufgenommenV1DataSchema = z.Struct(z.Shape{
	"BestellungID":     z.String().UUID().Required(),
	"Positionen":       z.Slice(positionSchema).Min(1).Required(),
	"GesamtPreisCents": z.Int().GTE(0).Required(),
	"Kommentar":        z.String().Max(100),
})

func NewBestellungAufgenommenEvent(userID int, userName string, tischID int, positionen []Position, kommentar string) (e.Event, error) {
	// Generate PositionIDs for each position
	for i := range positionen {
		positionen[i].PositionID = uuid.New().String()
	}

	gesamtPreisCents := 0
	for _, pos := range positionen {
		gesamtPreisCents += pos.Einzelpreis * pos.Menge
	}

	data := bestellungAufgenommenV1Data{
		BestellungID:     uuid.New().String(),
		Positionen:       toPositionenEventData(positionen),
		GesamtPreisCents: gesamtPreisCents,
		Kommentar:        kommentar,
	}

	if err := bestellungAufgenommenV1DataSchema.Validate(&data); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return e.Event{}, fmt.Errorf("bestellung aufgenommen data validation failed: %v", issues)
	}

	event, err := e.New(userID, userName, string(EventTypeBestellungAufgenommenV1), "tisch:"+strconv.Itoa(tischID), data)
	if err != nil {
		return e.Event{}, err
	}

	return event, nil
}

func buildBestellungFromEvent(event e.Event) (Bestellung, error) {
	if event.Type != string(EventTypeBestellungAufgenommenV1) {
		return Bestellung{}, fmt.Errorf("unsupported event type: %s", event.Type)
	}

	tischID, err := parseTischIDFromSubject(event.Subject)
	if err != nil {
		return Bestellung{}, err
	}

	data := bestellungAufgenommenV1Data{}
	err = e.ParseData(event, &data, bestellungAufgenommenV1DataSchema)
	if err != nil {
		return Bestellung{}, err
	}

	bestellung := Bestellung{
		ID:               data.BestellungID,
		UserID:           event.UserID,
		TischID:          tischID,
		Positionen:       fromPositionenEventData(data.Positionen),
		GesamtPreisCents: data.GesamtPreisCents,
		Kommentar:        data.Kommentar,
		AufgenommenAm:    event.Time,
	}

	if err := bestellungSchema.Validate(&bestellung); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return Bestellung{}, fmt.Errorf("bestellung validation failed: %v", issues)
	}

	return bestellung, nil
}
