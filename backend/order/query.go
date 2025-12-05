package order

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	z "github.com/Oudwins/zog"
	e "github.com/nicograef/jotti/backend/event"
	"github.com/rs/zerolog"
)

type queryPersistence interface {
	ReadEventsBySubject(ctx context.Context, subject string, eventTypes []string) ([]e.Event, error)
}

type queryService struct {
	persistence queryPersistence
}

// GetOrders retrieves all orders for a given table by reading events from the database.
func (s *queryService) GetOrders(ctx context.Context, tableID int) ([]Order, error) {
	logger := zerolog.Ctx(ctx)
	events, err := s.persistence.ReadEventsBySubject(ctx, "table:"+strconv.Itoa(tableID), []string{string(eventTypeOrderPlacedV1)})
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
	data := &orderPlacedV1Data{}
	parseErr := orderPlacedV1DataSchema.Parse(event.Data, data)
	if parseErr != nil {
		issues := z.Issues.SanitizeMapAndCollect(parseErr)
		return nil, fmt.Errorf("validation failed: %v", issues)
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
