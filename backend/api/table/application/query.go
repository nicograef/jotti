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
	ReadEventsWithSnapshot(ctx context.Context, subject string, snapshotEventType string) ([]e.Event, error)
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

func (q Query) GetTischSaldo(ctx context.Context, tischID int) (int, error) {
	log := zerolog.Ctx(ctx)

	subject := "tisch:" + strconv.Itoa(tischID)
	events, err := q.EventRepo.ReadEventsWithSnapshot(ctx, subject, string(t.EventTypeSnapshotV1))
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to read events for tisch")
		return 0, ErrDatabase
	}

	saldoCents, err := t.GetSaldoFromEvents(events)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to calculate saldo from events")
		return 0, err
	}

	log.Info().Int("tisch_id", tischID).Int("saldo_cents", saldoCents).Msg("Calculated tisch saldo")
	return saldoCents, nil
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

func (q Query) GetTischUnbezahlt(ctx context.Context, tischID int) ([]t.Position, error) {
	log := zerolog.Ctx(ctx)

	subject := "tisch:" + strconv.Itoa(tischID)
	events, err := q.EventRepo.ReadEventsWithSnapshot(ctx, subject, string(t.EventTypeSnapshotV1))
	if err != nil {
		log.Error().Int("tisch_id", tischID).Msg("Failed to read events for tisch")
		return nil, ErrDatabase
	}

	unbezahltePositionen, err := t.GetUnbezahltePositionenFromEvents(events)
	if err != nil {
		log.Error().Int("tisch_id", tischID).Err(err).Msg("Failed to build unbezahlte positionen from events")
		return nil, err
	}

	log.Info().Int("tisch_id", tischID).Int("unbezahlt_count", len(unbezahltePositionen)).Msg("Retrieved unbezahlte positionen for tisch")
	return unbezahltePositionen, nil
}

func (q Query) GetTischUngeliefert(ctx context.Context, tischID int) ([]t.Position, error) {
	log := zerolog.Ctx(ctx)

	subject := "tisch:" + strconv.Itoa(tischID)
	events, err := q.EventRepo.ReadEventsWithSnapshot(ctx, subject, string(t.EventTypeSnapshotV1))
	if err != nil {
		log.Error().Int("tisch_id", tischID).Msg("Failed to read events for tisch")
		return nil, ErrDatabase
	}

	ungeliefertePositionen, err := t.GetUngeliefertePositionenFromEvents(events)
	if err != nil {
		log.Error().Int("tisch_id", tischID).Err(err).Msg("Failed to build ungelieferte positionen from events")
		return nil, err
	}

	log.Info().Int("tisch_id", tischID).Int("ungeliefert_count", len(ungeliefertePositionen)).Msg("Retrieved ungelieferte positionen for tisch")
	return ungeliefertePositionen, nil
}
