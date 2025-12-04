package order

import (
	"fmt"
	"strconv"

	z "github.com/Oudwins/zog"
	e "github.com/nicograef/jotti/backend/event"
)

type eventType string

const (
	eventTypeOrderPlacedV1 eventType = "jotti.order.placed:v1"
)

type orderPlacedV1Data struct {
	Products        []orderProduct `json:"products"`
	TotalPriceCents int            `json:"totalPriceCents"`
}

var orderPlacedV1DataSchema = z.Struct(z.Shape{
	"Products":        z.Slice(orderProductSchema).Min(1).Required(),
	"TotalPriceCents": z.Int().GTE(0, z.Message("TotalPriceCents must be non-negative")).Required(),
})

func newOrderPlacedV1Event(userID, tableID int, products []orderProduct, totalPriceCents int) (*e.Event, error) {
	data := orderPlacedV1Data{
		Products:        products,
		TotalPriceCents: totalPriceCents,
	}

	if err := orderPlacedV1DataSchema.Validate(data); err != nil {
		issues := z.Issues.SanitizeMapAndCollect(err)
		return nil, fmt.Errorf("validation failed: %v", issues)
	}

	event, err := e.New(e.Candidate{
		UserID:  userID,
		Type:    string(eventTypeOrderPlacedV1),
		Subject: "table:" + strconv.Itoa(tableID),
		Data:    data,
	})
	if err != nil {
		return nil, err
	}

	return event, nil
}
