package application

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/internal/logging"
)

type historieEventRepo interface {
	ReadDirektverkaufEvents(ctx context.Context, kassensitzungNr int) ([]event.Event, error)
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

	historie, err := kasse.BuildVorgangHistorie(events, kasse.BuildDirektverkaufHistorieEintrag)
	if err != nil {
		log.Error("Failed to build direktverkauf historie", "error", err)
		return nil, err
	}

	log.Info("Retrieved direktverkauf historie", "historie_count", len(historie))
	return historie, nil
}
