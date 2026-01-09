package table

import e "github.com/nicograef/jotti/backend/domain/event"

type EventType string

const (
	EventTypeOrderPlacedV1       EventType = "table.order-placed:v1"
	EventTypePaymentRegisteredV1 EventType = "table.payment-registered:v1"
	EventTypeProductsCanceledV1  EventType = "table.products-canceled:v1"
)

func GetBalanceFromEvents(events []e.Event) (int, error) {
	balanceCents := 0

	for _, event := range events {
		if event.Type == string(EventTypeOrderPlacedV1) {
			order, err := buildOrderFromEvent(event)
			if err != nil {
				return 0, err
			}
			balanceCents += order.TotalPriceCents
		} else if event.Type == string(EventTypePaymentRegisteredV1) {
			payment, err := buildPaymentFromEvent(event)
			if err != nil {
				return 0, err
			}
			balanceCents -= payment.TotalPaymentCents

		} else if event.Type == string(EventTypeProductsCanceledV1) {
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
		}
	}

	// reverse the order of the array so that the most recent event is first
	for i, j := 0, len(history)-1; i < j; i, j = i+1, j-1 {
		history[i], history[j] = history[j], history[i]
	}

	return history, nil
}

func GetUnpaidProductsFromEvents(events []e.Event) ([]OrderProduct, error) {
	unpaidProducts := []OrderProduct{}

	for _, event := range events {
		if event.Type == string(EventTypeOrderPlacedV1) {
			order, err := buildOrderFromEvent(event)
			if err != nil {
				return []OrderProduct{}, err
			}

			// accumulate quantities of unpaid products without duplicate product entries
			for _, orderProduct := range order.Products {
				found := false
				for i, unpaidProd := range unpaidProducts {
					if unpaidProd.ID == orderProduct.ID && unpaidProd.NetPriceCents == orderProduct.NetPriceCents {
						unpaidProducts[i].Quantity += orderProduct.Quantity
						found = true
						break
					}
				}
				if !found {
					unpaidProducts = append(unpaidProducts, orderProduct)
				}
			}
		} else if event.Type == string(EventTypePaymentRegisteredV1) {
			payment, err := buildPaymentFromEvent(event)
			if err != nil {
				return []OrderProduct{}, err
			}

			// reduce quantities of paid products from unpaidProducts
			for _, paidProduct := range payment.Products {
				for i := 0; i < len(unpaidProducts); i++ {
					if unpaidProducts[i].ID == paidProduct.ID && unpaidProducts[i].NetPriceCents == paidProduct.NetPriceCents {
						if unpaidProducts[i].Quantity > paidProduct.Quantity {
							unpaidProducts[i].Quantity -= paidProduct.Quantity
						} else {
							// remove product from unpaidProducts if fully paid
							unpaidProducts = append(unpaidProducts[:i], unpaidProducts[i+1:]...)
							i-- // adjust index after removal
						}
						break
					}
				}
			}
		} else if event.Type == string(EventTypeProductsCanceledV1) {
			cancelation, err := buildCancelationFromEvent(event)
			if err != nil {
				return []OrderProduct{}, err
			}

			// reduce quantities of canceled products from unpaidProducts
			for _, canceledProduct := range cancelation.Products {
				for i := 0; i < len(unpaidProducts); i++ {
					if unpaidProducts[i].ID == canceledProduct.ID && unpaidProducts[i].NetPriceCents == canceledProduct.NetPriceCents {
						if unpaidProducts[i].Quantity > canceledProduct.Quantity {
							unpaidProducts[i].Quantity -= canceledProduct.Quantity
						} else {
							// remove product from unpaidProducts if fully paid
							unpaidProducts = append(unpaidProducts[:i], unpaidProducts[i+1:]...)
							i-- // adjust index after removal
						}
						break
					}
				}
			}
		}
	}

	return unpaidProducts, nil
}
