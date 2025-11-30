package order

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/google/uuid"
	e "github.com/nicograef/jotti/backend/event"
	"github.com/nicograef/jotti/backend/product"
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

var OrderProductSchema = z.Struct(z.Shape{
	"ID":            product.IDSchema.Required(),
	"Name":          product.NameSchema.Required(),
	"NetPriceCents": product.NetPriceCentsSchema.Required(),
	"Quantity":      z.Int().GTE(1, z.Message("Quantity must be at least 1")).Required(),
})

// Order represents an order aggregate model.
type Order struct {
	ID                 uuid.UUID      `json:"id"`
	UserID             int            `json:"userId"`
	TableID            int            `json:"tableId"`
	Products           []OrderProduct `json:"products"`
	TotalNetPriceCents int            `json:"totalNetPriceCents"`
	PlacedAt           time.Time      `json:"placedAt"`
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

// PlaceOrder places a new order by writing an event to the database.
func (s *Service) PlaceOrder(ctx context.Context, userID, tableID int, products []OrderProduct) (*Order, error) {
	totalPriceCents := 0
	for _, product := range products {
		totalPriceCents += product.NetPriceCents * product.Quantity
	}

	event, err := e.New(e.Candidate{
		UserID:  userID,
		Type:    string(EventTypeOrderPlacedV1),
		Subject: "table:" + strconv.Itoa(tableID),
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
		ID:                 id,
		UserID:             userID,
		TableID:            tableID,
		Products:           products,
		TotalNetPriceCents: totalPriceCents,
		PlacedAt:           event.Time,
	}

	return order, nil
}

// GetOrders retrieves all orders for a given table by reading events from the database.
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
