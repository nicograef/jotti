package table

import e "github.com/nicograef/jotti/backend/domain/event"

type EventType string

const (
	EventTypeOrderPlacedV1       EventType = "table.order-placed:v1"
	EventTypePaymentRegisteredV1 EventType = "table.payment-registered:v1"
	EventTypeProductsCanceledV1  EventType = "table.products-canceled:v1"
	EventTypeProductsDeliveredV1 EventType = "table.products-delivered:v1"
	EventTypeSnapshotV1          EventType = "table.snapshot:v1"
)

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
		case string(EventTypeProductsCanceledV1):
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
		case string(EventTypeProductsCanceledV1):
			cancelation, err := buildCancelationFromEvent(event)
			if err != nil {
				return []any{}, err
			}
			history = append(history, cancelation)

		case string(EventTypeProductsDeliveredV1):
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

// accumulateProducts adds products to a list, merging quantities for matching products
func accumulateProducts(list []OrderProduct, products []OrderProduct) []OrderProduct {
	for _, product := range products {
		found := false
		for i, existing := range list {
			if existing.ID == product.ID && existing.NetPriceCents == product.NetPriceCents {
				list[i].Quantity += product.Quantity
				found = true
				break
			}
		}
		if !found {
			list = append(list, product)
		}
	}
	return list
}

// reduceProducts subtracts products from a list, removing entries when quantity reaches zero
func reduceProducts(list []OrderProduct, products []OrderProduct) []OrderProduct {
	for _, product := range products {
		for i := 0; i < len(list); i++ {
			if list[i].ID == product.ID && list[i].NetPriceCents == product.NetPriceCents {
				if list[i].Quantity > product.Quantity {
					list[i].Quantity -= product.Quantity
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

func GetUnpaidProductsFromEvents(events []e.Event) ([]OrderProduct, error) {
	unpaidProducts := []OrderProduct{}

	for _, event := range events {
		switch event.Type {
		case string(EventTypeSnapshotV1):
			snapshot, err := buildSnapshotFromEvent(event)
			if err != nil {
				return nil, err
			}
			unpaidProducts = snapshot.UnpaidProducts

		case string(EventTypeOrderPlacedV1):
			order, err := buildOrderFromEvent(event)
			if err != nil {
				return nil, err
			}
			unpaidProducts = accumulateProducts(unpaidProducts, order.Products)

		case string(EventTypePaymentRegisteredV1):
			payment, err := buildPaymentFromEvent(event)
			if err != nil {
				return nil, err
			}
			unpaidProducts = reduceProducts(unpaidProducts, payment.Products)

		case string(EventTypeProductsCanceledV1):
			cancelation, err := buildCancelationFromEvent(event)
			if err != nil {
				return nil, err
			}
			unpaidProducts = reduceProducts(unpaidProducts, cancelation.Products)
		}
	}

	return unpaidProducts, nil
}

func GetUndeliveredProductsFromEvents(events []e.Event) ([]OrderProduct, error) {
	undeliveredProducts := []OrderProduct{}

	for _, event := range events {
		switch event.Type {
		case string(EventTypeSnapshotV1):
			snapshot, err := buildSnapshotFromEvent(event)
			if err != nil {
				return nil, err
			}
			undeliveredProducts = snapshot.UndeliveredProducts

		case string(EventTypeOrderPlacedV1):
			order, err := buildOrderFromEvent(event)
			if err != nil {
				return nil, err
			}
			undeliveredProducts = accumulateProducts(undeliveredProducts, order.Products)

		case string(EventTypeProductsDeliveredV1):
			delivery, err := buildDeliveryFromEvent(event)
			if err != nil {
				return nil, err
			}
			undeliveredProducts = reduceProducts(undeliveredProducts, delivery.Products)

		case string(EventTypeProductsCanceledV1):
			cancelation, err := buildCancelationFromEvent(event)
			if err != nil {
				return nil, err
			}
			undeliveredProducts = reduceProducts(undeliveredProducts, cancelation.Products)
		}
	}

	return undeliveredProducts, nil
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
