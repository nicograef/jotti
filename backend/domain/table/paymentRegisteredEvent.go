package table

import (
	"fmt"
	"strconv"

	z "github.com/Oudwins/zog"
	"github.com/google/uuid"
	e "github.com/nicograef/jotti/backend/domain/event"
)

type paymentRegisteredV1Data struct {
	PaymentID string           `json:"paymentId"` // UUID string
	Products  []PaymentProduct `json:"products"`
}

var paymentRegisteredV1DataSchema = z.Struct(z.Shape{
	"PaymentID": z.String().UUID().Required(),
	"Products":  z.Slice(paymentProductSchema).Min(1).Required(),
})

func NewPaymentRegisteredEvent(userID, tableID int, products []PaymentProduct) (e.Event, error) {
	data := paymentRegisteredV1Data{
		PaymentID: uuid.New().String(),
		Products:  products,
	}

	if err := paymentRegisteredV1DataSchema.Validate(&data); err != nil {
		issues := z.Issues.SanitizeMapAndCollect(err)
		return e.Event{}, fmt.Errorf("payment registered data validation failed: %v", issues)
	}

	event, err := e.New(userID, string(EventTypePaymentRegisteredV1), "table:"+strconv.Itoa(tableID), data)
	if err != nil {
		return e.Event{}, err
	}

	return event, nil
}

func buildPaymentFromEvent(event e.Event) (Payment, error) {
	if event.Type != string(EventTypePaymentRegisteredV1) {
		return Payment{}, fmt.Errorf("unsupported event type: %s", event.Type)
	}

	tableID, err := strconv.Atoi(event.Subject[len("table:"):])
	if err != nil {
		return Payment{}, fmt.Errorf("invalid table ID in event subject: %v", err)
	}

	data := paymentRegisteredV1Data{}
	err = e.ParseData(event, &data, paymentRegisteredV1DataSchema)
	if err != nil {
		return Payment{}, err
	}

	totalPaymentCents := 0
	for _, product := range data.Products {
		totalPaymentCents += product.NetPriceCents * product.Quantity
	}

	payment := Payment{
		ID:                data.PaymentID,
		UserID:            event.UserID,
		TableID:           tableID,
		Products:          data.Products,
		TotalPaymentCents: totalPaymentCents,
		RegisteredAt:      event.Time,
	}

	if err := paymentSchema.Validate(&payment); err != nil {
		issues := z.Issues.SanitizeMapAndCollect(err)
		return Payment{}, fmt.Errorf("payment validation failed: %v", issues)
	}

	return payment, nil
}
