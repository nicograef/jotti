package application

import (
	"context"
	"slices"

	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/rs/zerolog"
)

type historieEventRepo interface {
	ReadDirektverkaufEvents(ctx context.Context, kassensitzungNr int) ([]event.Event, error)
}

type Query struct {
	EventRepo           historieEventRepo
	KassensitzungenRepo kassensitzungenRepo
}

// GetDirektverkaufHistorie returns the compact Direktverkauf history of the open Kassensitzung —
// one entry per Verkauf, most recent first. Returns an empty slice when no Kassensitzung is open.
func (q Query) GetDirektverkaufHistorie(ctx context.Context) ([]kasse.DirektverkaufHistorieEintrag, error) {
	log := zerolog.Ctx(ctx)

	ks, err := q.KassensitzungenRepo.GetOffeneKassensitzung(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load open Kassensitzung for direktverkauf historie")
		return nil, ErrDatabase
	}
	if ks == nil {
		return []kasse.DirektverkaufHistorieEintrag{}, nil
	}

	events, err := q.EventRepo.ReadDirektverkaufEvents(ctx, ks.ZNr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read direktverkauf events")
		return nil, ErrDatabase
	}

	historie, err := buildHistorie(events)
	if err != nil {
		log.Error().Err(err).Msg("Failed to build direktverkauf historie")
		return nil, err
	}

	log.Info().Int("historie_count", len(historie)).Msg("Retrieved direktverkauf historie")
	return historie, nil
}

// buildHistorie groups the flat event list by verkauf stream (subject) and builds one compact
// entry per Verkauf, ordered most recent first.
func buildHistorie(events []event.Event) ([]kasse.DirektverkaufHistorieEintrag, error) {
	subjectOrder := []string{}
	eventsBySubject := map[string][]event.Event{}
	for _, evt := range events {
		if _, seen := eventsBySubject[evt.Subject]; !seen {
			subjectOrder = append(subjectOrder, evt.Subject)
		}
		eventsBySubject[evt.Subject] = append(eventsBySubject[evt.Subject], evt)
	}

	historie := make([]kasse.DirektverkaufHistorieEintrag, 0, len(subjectOrder))
	for _, subject := range subjectOrder {
		eintrag, err := kasse.BuildDirektverkaufHistorieEintrag(eventsBySubject[subject])
		if err != nil {
			return nil, err
		}
		historie = append(historie, eintrag)
	}

	slices.Reverse(historie)

	return historie, nil
}
