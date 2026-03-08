package table

import (
	"fmt"
	"strconv"

	z "github.com/Oudwins/zog"
	"github.com/google/uuid"
	e "github.com/nicograef/jotti/backend/domain/event"
)

type produkteGeliefertV1Data struct {
	LieferungID string     `json:"lieferungId"`
	Positionen  []Position `json:"positionen"`
	Comment     string     `json:"comment"`
}

var produkteGeliefertV1DataSchema = z.Struct(z.Shape{
	"LieferungID": z.String().UUID().Required(),
	"Positionen":  z.Slice(positionSchema).Min(1).Required(),
	"Comment":     z.String().Max(100),
})

func NewProdukteGeliefertEvent(userID, tischID int, positionen []Position, comment string) (e.Event, error) {
	data := produkteGeliefertV1Data{
		LieferungID: uuid.New().String(),
		Positionen:  positionen,
		Comment:     comment,
	}

	if err := produkteGeliefertV1DataSchema.Validate(&data); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return e.Event{}, fmt.Errorf("produkte geliefert data validation failed: %v", issues)
	}

	event, err := e.New(userID, string(EventTypeProdukteGeliefertV1), "tisch:"+strconv.Itoa(tischID), data)
	if err != nil {
		return e.Event{}, err
	}

	return event, nil
}

func buildLieferungFromEvent(event e.Event) (Lieferung, error) {
	if event.Type != string(EventTypeProdukteGeliefertV1) {
		return Lieferung{}, fmt.Errorf("unsupported event type: %s", event.Type)
	}

	tischID, err := parseTischIDFromSubject(event.Subject)
	if err != nil {
		return Lieferung{}, err
	}

	data := produkteGeliefertV1Data{}
	err = e.ParseData(event, &data, produkteGeliefertV1DataSchema)
	if err != nil {
		return Lieferung{}, err
	}

	lieferung := Lieferung{
		ID:          data.LieferungID,
		UserID:      event.UserID,
		TischID:     tischID,
		Positionen:  data.Positionen,
		Comment:     data.Comment,
		GeliefertAm: event.Time,
	}

	if err := lieferungSchema.Validate(&lieferung); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return Lieferung{}, fmt.Errorf("lieferung validation failed: %v", issues)
	}

	return lieferung, nil
}
