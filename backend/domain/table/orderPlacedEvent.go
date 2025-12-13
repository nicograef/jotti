package table

import (
	"fmt"
	"strconv"

	z "github.com/Oudwins/zog"
	"github.com/google/uuid"
	e "github.com/nicograef/jotti/backend/domain/event"
)

type orderPlacedV1Data struct {
	OrderID  string         `json:"orderId"` // UUID string
	Products []OrderProduct `json:"products"`
}

var orderPlacedV1DataSchema = z.Struct(z.Shape{
	"OrderID":  z.String().UUID().Required(),
	"Products": z.Slice(orderProductSchema).Min(1).Required(),
})

func NewOrderPlacedEvent(userID, tableID int, products []OrderProduct) (e.Event, error) {
	data := orderPlacedV1Data{
		OrderID:  uuid.New().String(),
		Products: products,
	}

	if err := orderPlacedV1DataSchema.Validate(&data); err != nil {
		issues := z.Issues.SanitizeMapAndCollect(err)
		return e.Event{}, fmt.Errorf("order placed data validation failed: %v", issues)
	}

	event, err := e.New(userID, string(EventTypeOrderPlacedV1), "table:"+strconv.Itoa(tableID), data)
	if err != nil {
		return e.Event{}, err
	}

	return event, nil
}

func buildOrderFromEvent(event e.Event) (Order, error) {
	if event.Type != string(EventTypeOrderPlacedV1) {
		return Order{}, fmt.Errorf("unsupported event type: %s", event.Type)
	}

	tableID, err := strconv.Atoi(event.Subject[len("table:"):])
	if err != nil {
		return Order{}, fmt.Errorf("invalid table ID in event subject: %v", err)
	}

	data := orderPlacedV1Data{}
	err = e.ParseData(event, &data, orderPlacedV1DataSchema)
	if err != nil {
		return Order{}, err
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
		return Order{}, fmt.Errorf("order validation failed: %v", issues)
	}

	return order, nil
}
