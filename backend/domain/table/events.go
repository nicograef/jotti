package table

import e "github.com/nicograef/jotti/backend/domain/event"

type EventType string

const (
	EventTypeOrderPlacedV1       EventType = "table.order-placed:v1"
	EventTypePaymentRegisteredV1 EventType = "table.payment-registered:v1"
)

func GetBalanceFromEvents(events []e.Event) (int, error) {
	balanceCents := 0

	for _, event := range events {
		if event.Type == string(EventTypeOrderPlacedV1) {
			order, err := buildOrderFromEvent(event)
			if err != nil {
				return 0, err
			}
			balanceCents += order.TotalNetPriceCents
		} else if event.Type == string(EventTypePaymentRegisteredV1) {
			payment, err := buildPaymentFromEvent(event)
			if err != nil {
				return 0, err
			}
			balanceCents -= payment.TotalPaymentCents
		}
	}

	return balanceCents, nil
}

func GetOrdersFromEvents(events []e.Event) ([]Order, error) {
	orders := []Order{}

	for _, event := range events {
		if event.Type == string(EventTypeOrderPlacedV1) {
			order, err := buildOrderFromEvent(event)
			if err != nil {
				return []Order{}, err
			}
			orders = append(orders, order)
		}
	}

	return orders, nil
}

func GetPaymentsFromEvents(events []e.Event) ([]Payment, error) {
	payments := []Payment{}

	for _, event := range events {
		if event.Type == string(EventTypePaymentRegisteredV1) {
			payment, err := buildPaymentFromEvent(event)
			if err != nil {
				return []Payment{}, err
			}
			payments = append(payments, payment)
		}
	}

	return payments, nil
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
		}
	}

	return unpaidProducts, nil
}
