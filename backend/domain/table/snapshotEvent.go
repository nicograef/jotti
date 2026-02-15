package table

import (
	"fmt"
	"strconv"
	"time"

	z "github.com/Oudwins/zog"
	e "github.com/nicograef/jotti/backend/domain/event"
)

type snapshotV1Data struct {
	BalanceCents        int        `json:"balanceCents"`
	UnpaidVariants      []LineItem `json:"unpaidVariants"`
	UndeliveredVariants []LineItem `json:"undeliveredVariants"`
	TotalPaymentsCents  int        `json:"totalPaymentsCents"`
}

var snapshotV1DataSchema = z.Struct(z.Shape{
	"BalanceCents":        z.Int(),                 // Can be 0 (no outstanding balance)
	"UnpaidVariants":      z.Slice(lineItemSchema), // Can be empty slice
	"UndeliveredVariants": z.Slice(lineItemSchema), // Can be empty slice
	"TotalPaymentsCents":  z.Int(),                 // Can be 0 (no payments yet)
})

// Snapshot represents the materialized state at a point in time
type Snapshot struct {
	TableID             int
	BalanceCents        int
	UnpaidVariants      []LineItem
	UndeliveredVariants []LineItem
	TotalPaymentsCents  int
	CreatedAt           time.Time
}

func NewSnapshotEvent(userID, tableID int, balance int, unpaid, undelivered []LineItem, totalPayments int) (e.Event, error) {
	data := snapshotV1Data{
		BalanceCents:        balance,
		UnpaidVariants:      unpaid,
		UndeliveredVariants: undelivered,
		TotalPaymentsCents:  totalPayments,
	}

	if err := snapshotV1DataSchema.Validate(&data); err != nil {
		issues := z.Issues.SanitizeMapAndCollect(err)
		return e.Event{}, fmt.Errorf("snapshot data validation failed: %v", issues)
	}

	return e.New(userID, string(EventTypeSnapshotV1), "table:"+strconv.Itoa(tableID), data)
}

func buildSnapshotFromEvent(event e.Event) (Snapshot, error) {
	if event.Type != string(EventTypeSnapshotV1) {
		return Snapshot{}, fmt.Errorf("unsupported event type: %s", event.Type)
	}

	tableID, err := parseTableIDFromSubject(event.Subject)
	if err != nil {
		return Snapshot{}, err
	}

	var data snapshotV1Data
	if err := e.ParseData(event, &data, snapshotV1DataSchema); err != nil {
		return Snapshot{}, err
	}

	return Snapshot{
		TableID:             tableID,
		BalanceCents:        data.BalanceCents,
		UnpaidVariants:      data.UnpaidVariants,
		UndeliveredVariants: data.UndeliveredVariants,
		TotalPaymentsCents:  data.TotalPaymentsCents,
		CreatedAt:           event.Time,
	}, nil
}
