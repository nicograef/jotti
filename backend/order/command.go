package order

import (
	"context"
	"strconv"

	"github.com/google/uuid"
	e "github.com/nicograef/jotti/backend/event"
)

type commandPersistence interface {
	WriteEvent(ctx context.Context, event e.Event) (uuid.UUID, error)
	ReadEventsBySubject(ctx context.Context, subject string, eventTypes []string) ([]e.Event, error)
}

type commandService struct {
	persistence commandPersistence
}

// PlaceOrder places a new order by writing an event to the database.
func (s *commandService) PlaceOrder(ctx context.Context, userID, tableID int, products []OrderProduct) (*Order, error) {
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

	id, err := s.persistence.WriteEvent(ctx, *event)
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
