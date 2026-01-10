package table

import (
	"fmt"
	"strconv"

	z "github.com/Oudwins/zog"
	"github.com/google/uuid"
	e "github.com/nicograef/jotti/backend/domain/event"
)

type productsCanceledV1Data struct {
	CancelationID         string         `json:"cancelationId"` // UUID string
	Products              []OrderProduct `json:"products"`
	TotalCancelationCents int            `json:"totalCancelationCents"`
	Comment               string         `json:"comment"`
}

var productsCanceledV1DataSchema = z.Struct(z.Shape{
	"CancelationID":         z.String().UUID().Required(),
	"Products":              z.Slice(orderProductSchema).Min(1).Required(),
	"TotalCancelationCents": z.Int().GTE(0).Required(),
	"Comment":               z.String().Optional(),
})

func NewProductsCanceledEvent(userID, tableID int, products []OrderProduct, comment string) (e.Event, error) {
	totalCancelationCents := 0
	for _, product := range products {
		totalCancelationCents += product.NetPriceCents * product.Quantity
	}

	data := productsCanceledV1Data{
		CancelationID:         uuid.New().String(),
		Products:              products,
		TotalCancelationCents: totalCancelationCents,
		Comment:               comment,
	}

	if err := productsCanceledV1DataSchema.Validate(&data); err != nil {
		issues := z.Issues.SanitizeMapAndCollect(err)
		return e.Event{}, fmt.Errorf("products canceled data validation failed: %v", issues)
	}

	event, err := e.New(userID, string(EventTypeProductsCanceledV1), "table:"+strconv.Itoa(tableID), data)
	if err != nil {
		return e.Event{}, err
	}

	return event, nil
}

func buildCancelationFromEvent(event e.Event) (Cancelation, error) {
	if event.Type != string(EventTypeProductsCanceledV1) {
		return Cancelation{}, fmt.Errorf("unsupported event type: %s", event.Type)
	}

	tableID, err := strconv.Atoi(event.Subject[len("table:"):])
	if err != nil {
		return Cancelation{}, fmt.Errorf("invalid table ID in event subject: %v", err)
	}

	data := productsCanceledV1Data{}
	err = e.ParseData(event, &data, productsCanceledV1DataSchema)
	if err != nil {
		return Cancelation{}, err
	}

	cancelation := Cancelation{
		ID:                    data.CancelationID,
		UserID:                event.UserID,
		TableID:               tableID,
		Products:              data.Products,
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
