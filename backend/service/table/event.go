package table

import (
	"fmt"
	"strconv"
	"strings"

	z "github.com/Oudwins/zog"
	e "github.com/nicograef/jotti/backend/event"
)

type eventType string

const (
	eventTypeOrderPlacedV1 eventType = "jotti.order.placed:v1"
)

type orderPlacedV1Data struct {
	OrderID  string         `json:"orderId"` // UUID string
	Products []orderProduct `json:"products"`
}

var orderPlacedV1DataSchema = z.Struct(z.Shape{
	"OrderID":  z.String().UUID().Required(),
	"Products": z.Slice(orderProductSchema).Min(1).Required(),
})

func newOrderPlacedV1Event(userID, tableID int, orderID string, products []orderProduct) (*e.Event, error) {
	data := orderPlacedV1Data{
		OrderID:  orderID,
		Products: products,
	}

	if err := orderPlacedV1DataSchema.Validate(&data); err != nil {
		issues := z.Issues.SanitizeMapAndCollect(err)
		return nil, fmt.Errorf("order placed data validation failed: %v", issues)
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

func buildOrderFromEvent(event e.Event) (*Order, error) {
	data := orderPlacedV1Data{}
	err := e.ParseData(event, &data, orderPlacedV1DataSchema)
	if err != nil {
		return nil, err
	}

	tableIDStr := strings.TrimPrefix(event.Subject, "table:")
	tableID, err := strconv.Atoi(tableIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid table ID: %w", err)
	}

	totalPriceCents := 0
	for _, product := range data.Products {
		totalPriceCents += product.NetPriceCents * product.Quantity
	}

	order := Order{
		ID:                 data.OrderID,
		UserID:             event.UserID,
		TableID:            tableID,
		Products:           data.Products,
		TotalNetPriceCents: totalPriceCents,
		PlacedAt:           event.Time,
	}

	if err := orderSchema.Validate(&order); err != nil {
		issues := z.Issues.SanitizeMapAndCollect(err)
		return nil, fmt.Errorf("order validation failed: %v", issues)
	}

	return &order, nil
}
