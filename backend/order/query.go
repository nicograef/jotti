package order

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	e "github.com/nicograef/jotti/backend/event"
	"github.com/rs/zerolog"
)

type queryPersistence interface {
	ReadEventsBySubject(ctx context.Context, subject string, eventTypes []string) ([]e.Event, error)
}

type Query struct {
	Persistence queryPersistence
}

// GetOrders retrieves all orders for a given table by reading events from the database.
func (s *Query) GetOrders(ctx context.Context, tableID int) ([]Order, error) {
	logger := zerolog.Ctx(ctx)
	events, err := s.Persistence.ReadEventsBySubject(ctx, "table:"+strconv.Itoa(tableID), []string{string(eventTypeOrderPlacedV1)})
	if err != nil {
		logger.Error().Int("table_id", tableID).Err(err).Msg("Failed to read order events for table")
		return nil, ErrDatabase
	}

	var orders []Order
	for _, event := range events {
		order, err := buildOrderFromEvent(event)
		if err != nil {
			logger.Error().Int("table_id", tableID).Err(err).Msg("Failed to build order from event")
			return nil, err
		}
		orders = append(orders, *order)
	}

	return orders, nil
}

func buildOrderFromEvent(event e.Event) (*Order, error) {
	data := orderPlacedV1Data{}
	err := e.ParseData(event, &data, orderPlacedV1DataSchema)
	if err != nil {
		return nil, err
	}

	tableIDStr := strings.TrimPrefix(event.Subject, "table:")
	tableID, err := strconv.Atoi(tableIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid table ID: %w", err)
	}

	order := Order{
		ID:                 event.ID,
		UserID:             event.UserID,
		TableID:            tableID,
		Products:           data.Products,
		TotalNetPriceCents: data.TotalPriceCents,
		PlacedAt:           event.Time,
	}

	return &order, nil
}
