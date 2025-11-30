package order

import (
	"context"
	"errors"

	"github.com/google/uuid"
	e "github.com/nicograef/jotti/backend/event"
)

type EventType string

const (
	// EventTypeOrderPlaced represents an order placed event.
	EventTypeOrderPlacedV1 EventType = "jotti.order.placed:v1"
)

type OrderPlacedData struct {
	Products        []OrderProduct `json:"products"`
	TotalPriceCents int            `json:"totalPriceCents"`
}

// OrderProduct represents a product within an order.
type OrderProduct struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	NetPriceCents int    `json:"netPriceCents"`
	Quantity      int    `json:"quantity"`
}

// Order represents an order aggregate model.
type Order struct {
	ID              uuid.UUID      `json:"id"`
	UserID          int            `json:"userId"`
	TableID         string         `json:"tableId"`
	Products        []OrderProduct `json:"products"`
	TotalPriceCents int            `json:"totalPriceCents"`
}

// ErrDatabase is returned when there is a database error.
var ErrDatabase = errors.New("database error")

type persistence interface {
	WriteEvent(ctx context.Context, event e.Event) (uuid.UUID, error)
	ReadEventsBySubject(ctx context.Context, subject string, eventTypes []string) ([]e.Event, error)
}

// Service provides order-related operations.
type Service struct {
	Persistence persistence
}

func (s *Service) PlaceOrder(ctx context.Context, userID int, tableID string, products []OrderProduct) (*Order, error) {
	totalPriceCents := 0
	for _, product := range products {
		totalPriceCents += product.NetPriceCents * product.Quantity
	}

	event, err := e.New(e.Candidate{
		UserID:  userID,
		Type:    string(EventTypeOrderPlacedV1),
		Subject: "table:" + tableID,
		Data:    OrderPlacedData{Products: products, TotalPriceCents: totalPriceCents},
	})
	if err != nil {
		return nil, err
	}

	id, err := s.Persistence.WriteEvent(ctx, *event)
	if err != nil {
		return nil, ErrDatabase
	}

	order := &Order{
		ID:              id,
		UserID:          userID,
		TableID:         tableID,
		Products:        products,
		TotalPriceCents: totalPriceCents,
	}

	return order, nil
}

func (s *Service) GetOrders(ctx context.Context, tableID string) (*[]Order, error) {
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

		order := Order{
			ID:              event.ID,
			UserID:          event.UserID,
			TableID:         tableID,
			Products:        products,
			TotalPriceCents: totalPriceCents,
		}
		orders = append(orders, order)
	}

	return &orders, nil
}
