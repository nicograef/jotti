package application

import (
	"context"
	"errors"
	"strconv"

	"github.com/nicograef/jotti/backend/db"
	t "github.com/nicograef/jotti/backend/domain/table"
	"github.com/rs/zerolog"
)

type Query struct {
	TableRepo   tableRepo
	EventRepo   eventRepo
	FavoritRepo favoritRepo
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

func (q Query) GetAktiveTische(ctx context.Context) ([]t.AktiverTisch, error) {
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

	tisch, err := q.TableRepo.GetTable(ctx, tischID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("tisch_id", tischID).Msg("Tisch not found")
			return t.TischState{}, ErrTischNotFound
		}
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to retrieve tisch")
		return t.TischState{}, ErrDatabase
	}

	state, err := q.EventRepo.ReadTableState(ctx, tischID)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to read table state")
		return t.TischState{}, ErrDatabase
	}

	state.TischID = tisch.ID
	state.TischName = tisch.Name

	log.Info().Int("tisch_id", tischID).Int("saldo_cents", state.SaldoCents).Msg("Retrieved tisch state")
	return state, nil
}

func (q Query) GetAktiveTischeMitFavoriten(ctx context.Context, userID int) ([]t.AktiverTischMitFavorit, error) {
	log := zerolog.Ctx(ctx)

	tische, err := q.TableRepo.GetActiveTablesWithFavorites(ctx, userID)
	if err != nil {
		log.Error().Err(err).Int("user_id", userID).Msg("Failed to retrieve active tische mit favoriten")
		return nil, ErrDatabase
	}

	log.Info().Int("user_id", userID).Int("count", len(tische)).Msg("Retrieved active tische mit favoriten")
	return tische, nil
}

func (q Query) GetMeineTischeState(ctx context.Context, userID int) ([]t.TischState, error) {
	log := zerolog.Ctx(ctx)

	favoritIDs, err := q.FavoritRepo.GetByUser(ctx, userID)
	if err != nil {
		log.Error().Err(err).Int("user_id", userID).Msg("Failed to retrieve favorit IDs")
		return nil, ErrDatabase
	}

	if len(favoritIDs) == 0 {
		log.Debug().Int("user_id", userID).Msg("No favoriten for user")
		return []t.TischState{}, nil
	}

	states, err := q.TableRepo.GetTableStatesByIDs(ctx, favoritIDs)
	if err != nil {
		log.Error().Err(err).Int("user_id", userID).Msg("Failed to retrieve tisch states by IDs")
		return nil, ErrDatabase
	}

	for i, state := range states {
		tisch, err := q.TableRepo.GetTable(ctx, state.TischID)
		if err != nil {
			log.Error().Err(err).Int("tisch_id", state.TischID).Msg("Failed to resolve tisch name")
			return nil, ErrDatabase
		}
		states[i].TischName = tisch.Name
	}

	log.Info().Int("user_id", userID).Int("count", len(states)).Msg("Retrieved meine tische state")
	return states, nil
}

func (q Query) GetTischHistorie(ctx context.Context, tischID int) ([]t.HistorieEintrag, error) {
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
