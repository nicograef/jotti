package table

import (
	"fmt"
	"strconv"

	z "github.com/Oudwins/zog"
	"github.com/google/uuid"
	e "github.com/nicograef/jotti/backend/domain/event"
)

type variantsDeliveredV1Data struct {
	DeliveryID string     `json:"deliveryId"` // UUID string
	Variants   []LineItem `json:"variants"`
	Comment    string     `json:"comment"`
}

var variantsDeliveredV1DataSchema = z.Struct(z.Shape{
	"DeliveryID": z.String().UUID().Required(),
	"Variants":   z.Slice(lineItemSchema).Min(1).Required(),
	"Comment":    z.String().Max(100),
})

func NewVariantsDeliveredEvent(userID, tableID int, variants []LineItem, comment string) (e.Event, error) {
	data := variantsDeliveredV1Data{
		DeliveryID: uuid.New().String(),
		Variants:   variants,
		Comment:    comment,
	}

	if err := variantsDeliveredV1DataSchema.Validate(&data); err != nil {
		issues := z.Issues.SanitizeMapAndCollect(err)
		return e.Event{}, fmt.Errorf("variants delivered data validation failed: %v", issues)
	}

	event, err := e.New(userID, string(EventTypeVariantsDeliveredV1), "table:"+strconv.Itoa(tableID), data)
	if err != nil {
		return e.Event{}, err
	}

	return event, nil
}

func buildDeliveryFromEvent(event e.Event) (Delivery, error) {
	if event.Type != string(EventTypeVariantsDeliveredV1) {
		return Delivery{}, fmt.Errorf("unsupported event type: %s", event.Type)
	}

	tableID, err := parseTableIDFromSubject(event.Subject)
	if err != nil {
		return Delivery{}, err
	}

	data := variantsDeliveredV1Data{}
	err = e.ParseData(event, &data, variantsDeliveredV1DataSchema)
	if err != nil {
		return Delivery{}, err
	}

	delivery := Delivery{
		ID:          data.DeliveryID,
		UserID:      event.UserID,
		TableID:     tableID,
		Variants:    data.Variants,
		Comment:     data.Comment,
		DeliveredAt: event.Time,
	}

	if err := deliverySchema.Validate(&delivery); err != nil {
		issues := z.Issues.SanitizeMapAndCollect(err)
		return Delivery{}, fmt.Errorf("delivery validation failed: %v", issues)
	}

	return delivery, nil
}
