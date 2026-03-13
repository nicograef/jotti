package table

import (
	"fmt"
	"strconv"

	z "github.com/Oudwins/zog"
	"github.com/google/uuid"
	e "github.com/nicograef/jotti/backend/domain/event"
)

type zahlungRegistriertV1Data struct {
	ZahlungID          string     `json:"zahlungId"`
	Positionen         []Position `json:"positionen"`
	GesamtZahlungCents int        `json:"gesamtZahlungCents"`
	Kommentar          string     `json:"kommentar"`
}

var zahlungRegistriertV1DataSchema = z.Struct(z.Shape{
	"ZahlungID":          z.String().UUID().Required(),
	"Positionen":         z.Slice(positionSchema).Min(1).Required(),
	"GesamtZahlungCents": z.Int().GTE(0).Required(),
	"Kommentar":          z.String().Max(100),
})

func NewZahlungRegistriertEvent(userID int, userName string, tischID int, positionen []Position, gesamtZahlungCents int, kommentar string) (e.Event, error) {
	data := zahlungRegistriertV1Data{
		ZahlungID:          uuid.New().String(),
		Positionen:         positionen,
		GesamtZahlungCents: gesamtZahlungCents,
		Kommentar:          kommentar,
	}

	if err := zahlungRegistriertV1DataSchema.Validate(&data); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return e.Event{}, fmt.Errorf("zahlung registriert data validation failed: %v", issues)
	}

	event, err := e.New(userID, userName, string(EventTypeZahlungRegistriertV1), "tisch:"+strconv.Itoa(tischID), data)
	if err != nil {
		return e.Event{}, err
	}

	return event, nil
}

func buildZahlungFromEvent(event e.Event) (Zahlung, error) {
	if event.Type != string(EventTypeZahlungRegistriertV1) {
		return Zahlung{}, fmt.Errorf("unsupported event type: %s", event.Type)
	}

	tischID, err := parseTischIDFromSubject(event.Subject)
	if err != nil {
		return Zahlung{}, err
	}

	data := zahlungRegistriertV1Data{}
	err = e.ParseData(event, &data, zahlungRegistriertV1DataSchema)
	if err != nil {
		return Zahlung{}, err
	}

	zahlung := Zahlung{
		ID:                 data.ZahlungID,
		UserID:             event.UserID,
		TischID:            tischID,
		Positionen:         data.Positionen,
		GesamtZahlungCents: data.GesamtZahlungCents,
		Kommentar:          data.Kommentar,
		RegistriertAm:      event.Time,
	}

	if err := zahlungSchema.Validate(&zahlung); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return Zahlung{}, fmt.Errorf("zahlung validation failed: %v", issues)
	}

	return zahlung, nil
}
