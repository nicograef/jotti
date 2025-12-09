package order

import (
	"context"
	"strconv"

	e "github.com/nicograef/jotti/backend/event"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type queryPersistence interface {
	ReadEventsBySubject(ctx context.Context, subject string, eventTypes []string) ([]e.Event, error)
}

type Query struct {
	Persistence queryPersistence
}

// GetOrders retrieves all orders for a given table by reading events from the database.
func (q *Query) GetOrders(ctx context.Context, tableID int) ([]Order, error) {
	logger := zerolog.Ctx(ctx)

	events, err := q.Persistence.ReadEventsBySubject(ctx, "table:"+strconv.Itoa(tableID), []string{string(eventTypeOrderPlacedV1)})
	if err != nil {
		logger.Error().Int("table_id", tableID).Msg("Failed to read order events for table")
		return nil, ErrDatabase
	}

	orders := []Order{}
	for _, event := range events {
		order, err := buildOrderFromEvent(event)
		if err != nil {
			logger.Error().Int("table_id", tableID).Err(err).Msg("Failed to build order from event")
			return nil, err
		}
		orders = append(orders, *order)
	}

	log.Info().Int("table_id", tableID).Int("order_count", len(orders)).Msg("Retrieved orders for table")
	return orders, nil
}

func (q *Query) GetTableBalance(ctx context.Context, tableID int) (int, error) {
	logger := zerolog.Ctx(ctx)

	events, err := q.Persistence.ReadEventsBySubject(ctx, "table:"+strconv.Itoa(tableID), []string{string(eventTypeOrderPlacedV1)})
	if err != nil {
		logger.Error().Int("table_id", tableID).Msg("Failed to read order events for table")
		return 0, ErrDatabase
	}

	totalBalanceCents := 0
	for _, event := range events {
		if event.Type == string(eventTypeOrderPlacedV1) {
			order, err := buildOrderFromEvent(event)
			if err != nil {
				logger.Error().Int("table_id", tableID).Err(err).Msg("Failed to build order from event")
				return 0, err
			}
			totalBalanceCents += order.TotalNetPriceCents
		}
	}

	log.Info().Int("table_id", tableID).Int("total_balance_cents", totalBalanceCents).Msg("Calculated table balance")
	return totalBalanceCents, nil
}

func (q *Query) GetTableUnpaidProducts(ctx context.Context, tableID int) ([]orderProduct, error) {
	logger := zerolog.Ctx(ctx)

	events, err := q.Persistence.ReadEventsBySubject(ctx, "table:"+strconv.Itoa(tableID), []string{string(eventTypeOrderPlacedV1)})
	if err != nil {
		logger.Error().Int("table_id", tableID).Msg("Failed to read order events for table")
		return nil, ErrDatabase
	}

	unpaidProducts := []orderProduct{}
	for _, event := range events {
		if event.Type == string(eventTypeOrderPlacedV1) {
			order, err := buildOrderFromEvent(event)
			if err != nil {
				logger.Error().Int("table_id", tableID).Err(err).Msg("Failed to build order from event")
				return nil, err
			}
			// accumulate quantities of unpaid products without duplicate product entries
			for _, prod := range order.Products {
				found := false
				for i, unpaidProd := range unpaidProducts {
					if unpaidProd.ID == prod.ID && unpaidProd.NetPriceCents == prod.NetPriceCents {
						unpaidProducts[i].Quantity += prod.Quantity
						found = true
						break
					}
				}
				if !found {
					unpaidProducts = append(unpaidProducts, prod)
				}
			}
		}
	}

	log.Info().Int("table_id", tableID).Int("unpaid_product_count", len(unpaidProducts)).Msg("Retrieved unpaid products for table")
	return unpaidProducts, nil
}
