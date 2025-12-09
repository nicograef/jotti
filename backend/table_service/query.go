package table_service

import (
	"context"
	"errors"
	"strconv"

	"github.com/nicograef/jotti/backend/db"
	e "github.com/nicograef/jotti/backend/event"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type eventQueryPersistence interface {
	ReadEventsBySubject(ctx context.Context, subject string, eventTypes []string) ([]e.Event, error)
}

type tableQueryPersistence interface {
	GetTable(ctx context.Context, id int) (*Table, error)
	GetAllTables(ctx context.Context) ([]Table, error)
}

type Query struct {
	EventPersistence eventQueryPersistence
	TablePersistence tableQueryPersistence
}

// GetTable retrieves a table by its ID.
func (q *Query) GetTable(ctx context.Context, id int) (*Table, error) {
	log := zerolog.Ctx(ctx)

	table, err := q.TablePersistence.GetTable(ctx, id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("table_id", id).Msg("Table not found")
			return nil, ErrTableNotFound
		} else {
			log.Error().Int("table_id", id).Msg("Failed to retrieve table")
			return nil, ErrDatabase
		}
	}

	log.Debug().Int("table_id", id).Msg("Table retrieved")
	return table, nil
}

// GetAllTables retrieves all tables.
func (q *Query) GetAllTables(ctx context.Context) ([]Table, error) {
	log := zerolog.Ctx(ctx)

	tables, err := q.TablePersistence.GetAllTables(ctx)
	if err != nil {
		log.Error().Msg("Failed to retrieve all tables")
		return nil, ErrDatabase
	}

	log.Debug().Int("count", len(tables)).Msg("Retrieved all tables")
	return tables, nil
}

// GetOrders retrieves all orders for a given table by reading events from the database.
func (q *Query) GetOrders(ctx context.Context, tableID int) ([]Order, error) {
	logger := zerolog.Ctx(ctx)

	subject := "table:" + strconv.Itoa(tableID)
	types := []string{string(eventTypeOrderPlacedV1)}
	events, err := q.EventPersistence.ReadEventsBySubject(ctx, subject, types)
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

	subject := "table:" + strconv.Itoa(tableID)
	types := []string{string(eventTypeOrderPlacedV1)}
	events, err := q.EventPersistence.ReadEventsBySubject(ctx, subject, types)
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

	subject := "table:" + strconv.Itoa(tableID)
	types := []string{string(eventTypeOrderPlacedV1)}
	events, err := q.EventPersistence.ReadEventsBySubject(ctx, subject, types)
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
