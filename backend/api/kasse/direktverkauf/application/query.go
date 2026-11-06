package application

import (
	"context"
	"slices"

	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/internal/logging"
)

type historieEventRepo interface {
	ReadDirektverkaufEvents(ctx context.Context, kassensitzungNr int) ([]event.Event, error)
	ReadPfandrueckgabeEvents(ctx context.Context, kassensitzungNr int, tischID int) ([]event.Event, error)
}

type Query struct {
	EventRepo           historieEventRepo
	KassensitzungenRepo kassensitzungenRepo
}

// GetDirektverkaufHistorie returns one entry per Verkauf of the active Kassensitzung, most recent
// first; an empty slice when no Kassensitzung is active.
func (q Query) GetDirektverkaufHistorie(ctx context.Context) ([]kasse.DirektverkaufHistorieEintrag, error) {
	log := logging.FromContext(ctx)

	ks, err := q.KassensitzungenRepo.GetAktiveKassensitzung(ctx)
	if err != nil {
		log.Error("Failed to load aktive Kassensitzung for direktverkauf historie", "error", err)
		return nil, ErrDatabase
	}
	if ks == nil {
		return []kasse.DirektverkaufHistorieEintrag{}, nil
	}

	events, err := q.EventRepo.ReadDirektverkaufEvents(ctx, ks.ZNr)
	if err != nil {
		log.Error("Failed to read direktverkauf events", "error", err)
		return nil, ErrDatabase
	}

	historie, err := buildHistorie(events, kasse.BuildDirektverkaufHistorieEintrag)
	if err != nil {
		log.Error("Failed to build direktverkauf historie", "error", err)
		return nil, err
	}

	log.Info("Retrieved direktverkauf historie", "historie_count", len(historie))
	return historie, nil
}

// GetPfandrueckgabeHistorie returns one entry per Pfandrückgabe at the table (0 = Direktverkauf) in
// the active Kassensitzung, most recent first; an empty slice when no Kassensitzung is active.
func (q Query) GetPfandrueckgabeHistorie(ctx context.Context, tischID int) ([]kasse.PfandrueckgabeHistorieEintrag, error) {
	log := logging.FromContext(ctx)

	ks, err := q.KassensitzungenRepo.GetAktiveKassensitzung(ctx)
	if err != nil {
		log.Error("Failed to load aktive Kassensitzung for pfandrueckgabe historie", "error", err)
		return nil, ErrDatabase
	}
	if ks == nil {
		return []kasse.PfandrueckgabeHistorieEintrag{}, nil
	}

	events, err := q.EventRepo.ReadPfandrueckgabeEvents(ctx, ks.ZNr, tischID)
	if err != nil {
		log.Error("Failed to read pfandrueckgabe events", "error", err, "tisch_id", tischID)
		return nil, ErrDatabase
	}

	historie, err := buildHistorie(events, kasse.BuildPfandrueckgabeHistorieEintrag)
	if err != nil {
		log.Error("Failed to build pfandrueckgabe historie", "error", err, "tisch_id", tischID)
		return nil, err
	}

	log.Info("Retrieved pfandrueckgabe historie", "tisch_id", tischID, "historie_count", len(historie))
	return historie, nil
}

// buildHistorie groups the events by subject, one entry per stream, most recent first.
func buildHistorie[T any](events []event.Event, buildEintrag func([]event.Event) (T, error)) ([]T, error) {
	subjectOrder := []string{}
	eventsBySubject := map[string][]event.Event{}
	for _, evt := range events {
		if _, seen := eventsBySubject[evt.Subject]; !seen {
			subjectOrder = append(subjectOrder, evt.Subject)
		}
		eventsBySubject[evt.Subject] = append(eventsBySubject[evt.Subject], evt)
	}

	historie := make([]T, 0, len(subjectOrder))
	for _, subject := range subjectOrder {
		eintrag, err := buildEintrag(eventsBySubject[subject])
		if err != nil {
			return nil, err
		}
		historie = append(historie, eintrag)
	}

	slices.Reverse(historie)

	return historie, nil
}
