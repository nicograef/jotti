package order

import (
	"context"
	"strconv"
	"strings"

	e "github.com/nicograef/jotti/backend/event"
)

type queryPersistence interface {
	ReadEventsBySubject(ctx context.Context, subject string, eventTypes []string) ([]e.Event, error)
}

type queryService struct {
	Persistence queryPersistence
}

// GetOrders retrieves all orders for a given table by reading events from the database.
func (s *queryService) GetOrders(ctx context.Context, tableID string) (*[]Order, error) {
	events, err := s.Persistence.ReadEventsBySubject(ctx, "table:"+tableID, []string{string(EventTypeOrderPlacedV1)})
	if err != nil {
		return nil, ErrDatabase
	}

	var orders []Order
	for _, event := range events {
		dataMap, ok := event.Data.(map[string]any)
		if !ok {
			continue
		}

		productsInterface, ok := dataMap["products"].([]any)
		if !ok {
			continue
		}

		var products []OrderProduct
		for _, p := range productsInterface {
			pMap, ok := p.(map[string]any)
			if !ok {
				continue
			}
			product := OrderProduct{
				ID:            int(pMap["id"].(float64)),
				Name:          pMap["name"].(string),
				NetPriceCents: int(pMap["netPriceCents"].(float64)),
				Quantity:      int(pMap["quantity"].(float64)),
			}
			products = append(products, product)
		}
		totalPriceCents := int(dataMap["totalPriceCents"].(float64))

		tableIDStr := strings.TrimPrefix(event.Subject, "table:")
		tableID, err := strconv.Atoi(tableIDStr)
		if err != nil {
			continue
		}

		order := Order{
			ID:                 event.ID,
			UserID:             event.UserID,
			TableID:            tableID,
			Products:           products,
			TotalNetPriceCents: totalPriceCents,
			PlacedAt:           event.Time,
		}
		orders = append(orders, order)
		orders = append(orders, order)
	}

	return &orders, nil
}
