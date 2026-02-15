package table

import (
	"fmt"
	"strconv"

	e "github.com/nicograef/jotti/backend/domain/event"
)

type EventType string

const (
	EventTypeOrderPlacedV1       EventType = "table.order-placed:v1"
	EventTypePaymentRegisteredV1 EventType = "table.payment-registered:v1"
	EventTypeVariantsCanceledV1  EventType = "table.variants-canceled:v1"
	EventTypeVariantsDeliveredV1 EventType = "table.variants-delivered:v1"
	EventTypeSnapshotV1          EventType = "table.snapshot:v1"
)

// parseTableIDFromSubject extracts the table ID from an event subject like "table:42".
func parseTableIDFromSubject(subject string) (int, error) {
	const prefix = "table:"
	if len(subject) <= len(prefix) || subject[:len(prefix)] != prefix {
		return 0, fmt.Errorf("invalid event subject format: %s", subject)
	}
	id, err := strconv.Atoi(subject[len(prefix):])
	if err != nil {
		return 0, fmt.Errorf("invalid table ID in event subject: %v", err)
	}
	return id, nil
}

func GetBalanceFromEvents(events []e.Event) (int, error) {
	balanceCents := 0

	for _, event := range events {
		switch event.Type {
		case string(EventTypeSnapshotV1):
			snapshot, err := buildSnapshotFromEvent(event)
			if err != nil {
				return 0, err
			}
			balanceCents = snapshot.BalanceCents

		case string(EventTypeOrderPlacedV1):
			order, err := buildOrderFromEvent(event)
			if err != nil {
				return 0, err
			}
			balanceCents += order.TotalPriceCents
		case string(EventTypePaymentRegisteredV1):
			payment, err := buildPaymentFromEvent(event)
			if err != nil {
				return 0, err
			}
			balanceCents -= payment.TotalPaymentCents
		case string(EventTypeVariantsCanceledV1):
			cancelation, err := buildCancelationFromEvent(event)
			if err != nil {
				return 0, err
			}
			balanceCents -= cancelation.TotalCancelationCents
		}
	}

	return balanceCents, nil
}

func GetHistoryFromEvents(events []e.Event) ([]any, error) {
	history := []any{}

	for _, event := range events {
		switch event.Type {
		case string(EventTypeOrderPlacedV1):
			order, err := buildOrderFromEvent(event)
			if err != nil {
				return []any{}, err
			}
			history = append(history, order)
		case string(EventTypePaymentRegisteredV1):
			payment, err := buildPaymentFromEvent(event)
			if err != nil {
				return []any{}, err
			}
			history = append(history, payment)
		case string(EventTypeVariantsCanceledV1):
			cancelation, err := buildCancelationFromEvent(event)
			if err != nil {
				return []any{}, err
			}
			history = append(history, cancelation)

		case string(EventTypeVariantsDeliveredV1):
			delivery, err := buildDeliveryFromEvent(event)
			if err != nil {
				return []any{}, err
			}
			history = append(history, delivery)
		}
	}

	// reverse the order of the array so that the most recent event is first
	for i, j := 0, len(history)-1; i < j; i, j = i+1, j-1 {
		history[i], history[j] = history[j], history[i]
	}

	return history, nil
}

// accumulateVariants adds variants to a list, merging quantities for matching variants
func accumulateVariants(list []LineItem, variants []LineItem) []LineItem {
	for _, variant := range variants {
		found := false
		for i, existing := range list {
			if existing.ID == variant.ID && existing.PriceCents == variant.PriceCents {
				list[i].Quantity += variant.Quantity
				found = true
				break
			}
		}
		if !found {
			list = append(list, variant)
		}
	}
	return list
}

// reduceVariants subtracts variants from a list, removing entries when quantity reaches zero
func reduceVariants(list []LineItem, variants []LineItem) []LineItem {
	for _, variant := range variants {
		for i := 0; i < len(list); i++ {
			if list[i].ID == variant.ID && list[i].PriceCents == variant.PriceCents {
				if list[i].Quantity > variant.Quantity {
					list[i].Quantity -= variant.Quantity
				} else {
					list = append(list[:i], list[i+1:]...)
					i--
				}
				break
			}
		}
	}
	return list
}

func GetUnpaidVariantsFromEvents(events []e.Event) ([]LineItem, error) {
	unpaidVariants := []LineItem{}

	for _, event := range events {
		switch event.Type {
		case string(EventTypeSnapshotV1):
			snapshot, err := buildSnapshotFromEvent(event)
			if err != nil {
				return nil, err
			}
			unpaidVariants = snapshot.UnpaidVariants

		case string(EventTypeOrderPlacedV1):
			order, err := buildOrderFromEvent(event)
			if err != nil {
				return nil, err
			}
			unpaidVariants = accumulateVariants(unpaidVariants, order.Variants)

		case string(EventTypePaymentRegisteredV1):
			payment, err := buildPaymentFromEvent(event)
			if err != nil {
				return nil, err
			}
			unpaidVariants = reduceVariants(unpaidVariants, payment.Variants)

		case string(EventTypeVariantsCanceledV1):
			cancelation, err := buildCancelationFromEvent(event)
			if err != nil {
				return nil, err
			}
			unpaidVariants = reduceVariants(unpaidVariants, cancelation.Variants)
		}
	}

	return unpaidVariants, nil
}

func GetUndeliveredVariantsFromEvents(events []e.Event) ([]LineItem, error) {
	undeliveredVariants := []LineItem{}

	for _, event := range events {
		switch event.Type {
		case string(EventTypeSnapshotV1):
			snapshot, err := buildSnapshotFromEvent(event)
			if err != nil {
				return nil, err
			}
			undeliveredVariants = snapshot.UndeliveredVariants

		case string(EventTypeOrderPlacedV1):
			order, err := buildOrderFromEvent(event)
			if err != nil {
				return nil, err
			}
			undeliveredVariants = accumulateVariants(undeliveredVariants, order.Variants)

		case string(EventTypeVariantsDeliveredV1):
			delivery, err := buildDeliveryFromEvent(event)
			if err != nil {
				return nil, err
			}
			undeliveredVariants = reduceVariants(undeliveredVariants, delivery.Variants)

		case string(EventTypeVariantsCanceledV1):
			cancelation, err := buildCancelationFromEvent(event)
			if err != nil {
				return nil, err
			}
			undeliveredVariants = reduceVariants(undeliveredVariants, cancelation.Variants)
		}
	}

	return undeliveredVariants, nil
}

func GetTotalPaymentsFromEvents(events []e.Event) (int, error) {
	totalPaymentsCents := 0

	for _, event := range events {
		switch event.Type {
		case string(EventTypeSnapshotV1):
			snapshot, err := buildSnapshotFromEvent(event)
			if err != nil {
				return 0, err
			}
			totalPaymentsCents = snapshot.TotalPaymentsCents

		case string(EventTypePaymentRegisteredV1):
			payment, err := buildPaymentFromEvent(event)
			if err != nil {
				return 0, err
			}
			totalPaymentsCents += payment.TotalPaymentCents
		}
	}

	return totalPaymentsCents, nil
}
