package table

import e "github.com/nicograef/jotti/backend/domain/event"

type EventType string

const (
	EventTypeOrderPlacedV1 EventType = "table.order-placed:v1"
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
		}
	}

	return unpaidProducts, nil
}
