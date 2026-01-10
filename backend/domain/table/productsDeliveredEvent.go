package table

import (
	"fmt"
	"strconv"

	z "github.com/Oudwins/zog"
	"github.com/google/uuid"
	e "github.com/nicograef/jotti/backend/domain/event"
)

type productsDeliveredV1Data struct {
	DeliveryID string         `json:"deliveryId"` // UUID string
	Products   []OrderProduct `json:"products"`
	Comment    string         `json:"comment"`
}

var productsDeliveredV1DataSchema = z.Struct(z.Shape{
	"DeliveryID": z.String().UUID().Required(),
	"Products":   z.Slice(orderProductSchema).Min(1).Required(),
	"Comment":    z.String().Max(100).Optional(),
})

func NewProductsDeliveredEvent(userID, tableID int, products []OrderProduct, comment string) (e.Event, error) {
	data := productsDeliveredV1Data{
		DeliveryID: uuid.New().String(),
		Products:   products,
		Comment:    comment,
	}

	if err := productsDeliveredV1DataSchema.Validate(&data); err != nil {
		issues := z.Issues.SanitizeMapAndCollect(err)
		return e.Event{}, fmt.Errorf("products delivered data validation failed: %v", issues)
	}

	event, err := e.New(userID, string(EventTypeProductsDeliveredV1), "table:"+strconv.Itoa(tableID), data)
	if err != nil {
		return e.Event{}, err
	}

	return event, nil
}

func buildDeliveryFromEvent(event e.Event) (Delivery, error) {
	if event.Type != string(EventTypeProductsDeliveredV1) {
		return Delivery{}, fmt.Errorf("unsupported event type: %s", event.Type)
	}

	tableID, err := strconv.Atoi(event.Subject[len("table:"):])
	if err != nil {
		return Delivery{}, fmt.Errorf("invalid table ID in event subject: %v", err)
	}

	data := productsDeliveredV1Data{}
	err = e.ParseData(event, &data, productsDeliveredV1DataSchema)
	if err != nil {
		return Delivery{}, err
	}

	delivery := Delivery{
		ID:          data.DeliveryID,
		UserID:      event.UserID,
		TableID:     tableID,
		Products:    data.Products,
		Comment:     data.Comment,
		DeliveredAt: event.Time,
	}

	if err := deliverySchema.Validate(&delivery); err != nil {
		issues := z.Issues.SanitizeMapAndCollect(err)
		return Delivery{}, fmt.Errorf("delivery validation failed: %v", issues)
	}

	return delivery, nil
}
