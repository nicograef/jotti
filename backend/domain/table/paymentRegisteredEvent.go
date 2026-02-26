package table

import (
	"fmt"
	"strconv"

	z "github.com/Oudwins/zog"
	"github.com/google/uuid"
	e "github.com/nicograef/jotti/backend/domain/event"
)

type paymentRegisteredV1Data struct {
	PaymentID         string     `json:"paymentId"` // UUID string
	Variants          []LineItem `json:"variants"`
	TotalPaymentCents int        `json:"totalPaymentCents"`
	Comment           string     `json:"comment"`
}

var paymentRegisteredV1DataSchema = z.Struct(z.Shape{
	"PaymentID":         z.String().UUID().Required(),
	"Variants":          z.Slice(lineItemSchema).Min(1).Required(),
	"TotalPaymentCents": z.Int().GTE(0).Required(),
	"Comment":           z.String().Max(100),
})

func NewPaymentRegisteredEvent(userID, tableID int, variants []LineItem, comment string) (e.Event, error) {
	totalPaymentCents := 0
	for _, variant := range variants {
		totalPaymentCents += variant.PriceCents * variant.Quantity
	}

	data := paymentRegisteredV1Data{
		PaymentID:         uuid.New().String(),
		Variants:          variants,
		TotalPaymentCents: totalPaymentCents,
		Comment:           comment,
	}

	if err := paymentRegisteredV1DataSchema.Validate(&data); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
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

	tableID, err := parseTableIDFromSubject(event.Subject)
	if err != nil {
		return Payment{}, err
	}

	data := paymentRegisteredV1Data{}
	err = e.ParseData(event, &data, paymentRegisteredV1DataSchema)
	if err != nil {
		return Payment{}, err
	}

	payment := Payment{
		ID:                data.PaymentID,
		UserID:            event.UserID,
		TableID:           tableID,
		Variants:          data.Variants,
		TotalPaymentCents: data.TotalPaymentCents,
		Comment:           data.Comment,
		RegisteredAt:      event.Time,
	}

	if err := paymentSchema.Validate(&payment); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return Payment{}, fmt.Errorf("payment validation failed: %v", issues)
	}

	return payment, nil
}
