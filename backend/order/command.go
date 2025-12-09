package order

import (
	"context"

	"github.com/google/uuid"
	e "github.com/nicograef/jotti/backend/event"
	"github.com/rs/zerolog"
)

type commandPersistence interface {
	WriteEvent(ctx context.Context, event e.Event) (int, error)
}

type Command struct {
	Persistence commandPersistence
}

// PlaceOrder places a new order by writing an event to the database.
func (s *Command) PlaceOrder(ctx context.Context, userID, tableID int, products []orderProduct) error {
	log := zerolog.Ctx(ctx)

	orderID := uuid.New()
	event, err := newOrderPlacedV1Event(userID, tableID, orderID.String(), products)
	if err != nil {
		log.Error().Err(err).Int("table_id", tableID).Msg("Failed to create order placed event")
		return err
	}

	_, err = s.Persistence.WriteEvent(ctx, *event)
	if err != nil {
		log.Error().Err(err).Int("table_id", tableID).Msg("Failed to write order placed event to database")
		return ErrDatabase
	}

	log.Info().Int("table_id", tableID).Msg("Order placed")
	return nil
}
