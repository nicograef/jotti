package application

import (
	"context"
	"errors"
	"strconv"

	"github.com/nicograef/jotti/backend/db"
	e "github.com/nicograef/jotti/backend/domain/event"
	t "github.com/nicograef/jotti/backend/domain/table"
	"github.com/rs/zerolog"
)

type tableRepoQuery interface {
	GetTable(ctx context.Context, id int) (t.Table, error)
	GetAllTables(ctx context.Context) ([]t.Table, error)
	GetActiveTables(ctx context.Context) ([]t.Table, error)
}

type eventRepoQuery interface {
	ReadEventsBySubject(ctx context.Context, subject string) ([]e.Event, error)
	ReadEventsWithSnapshot(ctx context.Context, subject string, snapshotEventType string) ([]e.Event, error)
}

type Query struct {
	TableRepo tableRepoQuery
	EventRepo eventRepoQuery
}

func (q Query) GetTable(ctx context.Context, id int) (t.Table, error) {
	log := zerolog.Ctx(ctx)

	table, err := q.TableRepo.GetTable(ctx, id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("table_id", id).Msg("Table not found")
			return t.Table{}, ErrTableNotFound
		} else {
			log.Error().Err(err).Int("table_id", id).Msg("Failed to retrieve table")
			return t.Table{}, ErrDatabase
		}
	}

	log.Debug().Int("table_id", id).Msg("Table retrieved")
	return table, nil
}

func (q Query) GetAllTables(ctx context.Context) ([]t.Table, error) {
	log := zerolog.Ctx(ctx)

	tables, err := q.TableRepo.GetAllTables(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve all tables")
		return nil, ErrDatabase
	}

	log.Debug().Int("count", len(tables)).Msg("Retrieved all tables")
	return tables, nil
}

func (q Query) GetActiveTables(ctx context.Context) ([]t.Table, error) {
	log := zerolog.Ctx(ctx)

	tables, err := q.TableRepo.GetActiveTables(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve active tables")
		return nil, ErrDatabase
	}

	log.Info().Int("count", len(tables)).Msg("Retrieved active tables")
	return tables, nil
}

func (q Query) GetTableBalance(ctx context.Context, tableID int) (int, error) {
	log := zerolog.Ctx(ctx)

	subject := "table:" + strconv.Itoa(tableID)
	events, err := q.EventRepo.ReadEventsWithSnapshot(ctx, subject, string(t.EventTypeSnapshotV1))
	if err != nil {
		log.Error().Err(err).Int("table_id", tableID).Msg("Failed to read order events for table")
		return 0, ErrDatabase
	}

	balanceCents, err := t.GetBalanceFromEvents(events)
	if err != nil {
		log.Error().Err(err).Int("table_id", tableID).Msg("Failed to calculate balance from events")
		return 0, err
	}

	log.Info().Int("table_id", tableID).Int("total_balance_cents", balanceCents).Msg("Calculated table balance")
	return balanceCents, nil
}

func (q Query) GetTableHistory(ctx context.Context, tableID int) ([]any, error) {
	log := zerolog.Ctx(ctx)

	// Note: History needs all events, not just since snapshot, to show full timeline
	subject := "table:" + strconv.Itoa(tableID)
	events, err := q.EventRepo.ReadEventsBySubject(ctx, subject)
	if err != nil {
		log.Error().Int("table_id", tableID).Msg("Failed to read events for table")
		return nil, ErrDatabase
	}

	history, err := t.GetHistoryFromEvents(events)
	if err != nil {
		log.Error().Int("table_id", tableID).Err(err).Msg("Failed to build history from events")
		return nil, err
	}

	log.Info().Int("table_id", tableID).Int("history_count", len(history)).Msg("Retrieved history for table")
	return history, nil
}

func (q Query) GetTableUnpaidProducts(ctx context.Context, tableID int) ([]t.OrderProduct, error) {
	log := zerolog.Ctx(ctx)

	subject := "table:" + strconv.Itoa(tableID)
	events, err := q.EventRepo.ReadEventsWithSnapshot(ctx, subject, string(t.EventTypeSnapshotV1))
	if err != nil {
		log.Error().Int("table_id", tableID).Msg("Failed to read events for table")
		return nil, ErrDatabase
	}

	unpaidProducts, err := t.GetUnpaidProductsFromEvents(events)
	if err != nil {
		log.Error().Int("table_id", tableID).Err(err).Msg("Failed to build unpaid products from events")
		return nil, err
	}

	log.Info().Int("table_id", tableID).Int("unpaid_product_count", len(unpaidProducts)).Msg("Retrieved unpaid products for table")
	return unpaidProducts, nil
}

func (q Query) GetTableUndeliveredProducts(ctx context.Context, tableID int) ([]t.OrderProduct, error) {
	log := zerolog.Ctx(ctx)

	subject := "table:" + strconv.Itoa(tableID)
	events, err := q.EventRepo.ReadEventsWithSnapshot(ctx, subject, string(t.EventTypeSnapshotV1))
	if err != nil {
		log.Error().Int("table_id", tableID).Msg("Failed to read events for table")
		return nil, ErrDatabase
	}

	undeliveredProducts, err := t.GetUndeliveredProductsFromEvents(events)
	if err != nil {
		log.Error().Int("table_id", tableID).Err(err).Msg("Failed to build undelivered products from events")
		return nil, err
	}

	log.Info().Int("table_id", tableID).Int("undelivered_product_count", len(undeliveredProducts)).Msg("Retrieved undelivered products for table")
	return undeliveredProducts, nil
}
