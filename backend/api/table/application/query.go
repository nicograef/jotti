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
	GetTable(ctx context.Context, id int) (t.Tisch, error)
	GetAllTables(ctx context.Context) ([]t.Tisch, error)
	GetActiveTables(ctx context.Context) ([]t.Tisch, error)
}

type eventRepoQuery interface {
	ReadEventsBySubject(ctx context.Context, subject string) ([]e.Event, error)
	ReadTableState(ctx context.Context, tischID int) (t.TischState, error)
}

type Query struct {
	TableRepo tableRepoQuery
	EventRepo eventRepoQuery
}

func (q Query) GetTisch(ctx context.Context, id int) (t.Tisch, error) {
	log := zerolog.Ctx(ctx)

	tisch, err := q.TableRepo.GetTable(ctx, id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("tisch_id", id).Msg("Tisch not found")
			return t.Tisch{}, ErrTischNotFound
		} else {
			log.Error().Err(err).Int("tisch_id", id).Msg("Failed to retrieve tisch")
			return t.Tisch{}, ErrDatabase
		}
	}

	log.Debug().Int("tisch_id", id).Msg("Tisch retrieved")
	return tisch, nil
}

func (q Query) GetAllTische(ctx context.Context) ([]t.Tisch, error) {
	log := zerolog.Ctx(ctx)

	tische, err := q.TableRepo.GetAllTables(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve all tische")
		return nil, ErrDatabase
	}

	log.Debug().Int("count", len(tische)).Msg("Retrieved all tische")
	return tische, nil
}

func (q Query) GetAktiveTische(ctx context.Context) ([]t.Tisch, error) {
	log := zerolog.Ctx(ctx)

	tische, err := q.TableRepo.GetActiveTables(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve active tische")
		return nil, ErrDatabase
	}

	log.Info().Int("count", len(tische)).Msg("Retrieved active tische")
	return tische, nil
}

func (q Query) GetTischState(ctx context.Context, tischID int) (t.TischState, error) {
	log := zerolog.Ctx(ctx)

	state, err := q.EventRepo.ReadTableState(ctx, tischID)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to read table state")
		return t.TischState{}, ErrDatabase
	}

	log.Info().Int("tisch_id", tischID).Int("saldo_cents", state.SaldoCents).Msg("Retrieved tisch state")
	return state, nil
}

func (q Query) GetTischHistorie(ctx context.Context, tischID int) ([]any, error) {
	log := zerolog.Ctx(ctx)

	// Note: History needs all events, not just since snapshot, to show full timeline
	subject := "tisch:" + strconv.Itoa(tischID)
	events, err := q.EventRepo.ReadEventsBySubject(ctx, subject)
	if err != nil {
		log.Error().Int("tisch_id", tischID).Msg("Failed to read events for tisch")
		return nil, ErrDatabase
	}

	historie, err := t.GetHistoryFromEvents(events)
	if err != nil {
		log.Error().Int("tisch_id", tischID).Err(err).Msg("Failed to build historie from events")
		return nil, err
	}

	log.Info().Int("tisch_id", tischID).Int("historie_count", len(historie)).Msg("Retrieved historie for tisch")
	return historie, nil
}
