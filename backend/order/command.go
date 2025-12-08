package order

import (
	"context"

	e "github.com/nicograef/jotti/backend/event"
)

type commandPersistence interface {
	WriteEvent(ctx context.Context, event e.Event) (int, error)
}

type Command struct {
	Persistence commandPersistence
}

// PlaceOrder places a new order by writing an event to the database.
func (s *Command) PlaceOrder(ctx context.Context, userID, tableID int, products []orderProduct) (int, error) {
	totalPriceCents := 0
	for _, product := range products {
		totalPriceCents += product.NetPriceCents * product.Quantity
	}

	event, err := newOrderPlacedV1Event(userID, tableID, products, totalPriceCents)
	if err != nil {
		return 0, err
	}

	eventID, err := s.Persistence.WriteEvent(ctx, *event)
	if err != nil {
		return 0, ErrDatabase
	}

	return eventID, nil
}
