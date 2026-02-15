package table

import (
	"fmt"
	"strconv"

	z "github.com/Oudwins/zog"
	"github.com/google/uuid"
	e "github.com/nicograef/jotti/backend/domain/event"
)

type variantsCanceledV1Data struct {
	CancelationID         string         `json:"cancelationId"` // UUID string
	Variants              []LineItem `json:"variants"`
	TotalCancelationCents int            `json:"totalCancelationCents"`
	Comment               string         `json:"comment"`
}

var variantsCanceledV1DataSchema = z.Struct(z.Shape{
	"CancelationID":         z.String().UUID().Required(),
	"Variants":              z.Slice(lineItemSchema).Min(1).Required(),
	"TotalCancelationCents": z.Int().GTE(0).Required(),
	"Comment":               z.String().Max(100),
})

func NewVariantsCanceledEvent(userID, tableID int, variants []LineItem, comment string) (e.Event, error) {
	totalCancelationCents := 0
	for _, variant := range variants {
		totalCancelationCents += variant.PriceCents * variant.Quantity
	}

	data := variantsCanceledV1Data{
		CancelationID:         uuid.New().String(),
		Variants:              variants,
		TotalCancelationCents: totalCancelationCents,
		Comment:               comment,
	}

	if err := variantsCanceledV1DataSchema.Validate(&data); err != nil {
		issues := z.Issues.SanitizeMapAndCollect(err)
		return e.Event{}, fmt.Errorf("variants canceled data validation failed: %v", issues)
	}

	event, err := e.New(userID, string(EventTypeVariantsCanceledV1), "table:"+strconv.Itoa(tableID), data)
	if err != nil {
		return e.Event{}, err
	}

	return event, nil
}

func buildCancelationFromEvent(event e.Event) (Cancelation, error) {
	if event.Type != string(EventTypeVariantsCanceledV1) {
		return Cancelation{}, fmt.Errorf("unsupported event type: %s", event.Type)
	}

	tableID, err := strconv.Atoi(event.Subject[len("table:"):])
	if err != nil {
		return Cancelation{}, fmt.Errorf("invalid table ID in event subject: %v", err)
	}

	data := variantsCanceledV1Data{}
	err = e.ParseData(event, &data, variantsCanceledV1DataSchema)
	if err != nil {
		return Cancelation{}, err
	}

	cancelation := Cancelation{
		ID:                    data.CancelationID,
		UserID:                event.UserID,
		TableID:               tableID,
		Variants:              data.Variants,
		TotalCancelationCents: data.TotalCancelationCents,
		Comment:               data.Comment,
		CanceledAt:            event.Time,
	}

	if err := cancelationSchema.Validate(&cancelation); err != nil {
		issues := z.Issues.SanitizeMapAndCollect(err)
		return Cancelation{}, fmt.Errorf("cancelation validation failed: %v", issues)
	}

	return cancelation, nil
}
