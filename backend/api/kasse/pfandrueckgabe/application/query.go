package application

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/internal/logging"
)

type historieEventRepo interface {
	ReadPfandrueckgabeEvents(ctx context.Context, kassensitzungNr int, tischID int) ([]event.Event, error)
}

type Query struct {
	EventRepo           historieEventRepo
	KassensitzungenRepo kassensitzungenRepo
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

	historie, err := kasse.BuildVorgangHistorie(events, kasse.BuildPfandrueckgabeHistorieEintrag)
	if err != nil {
		log.Error("Failed to build pfandrueckgabe historie", "error", err, "tisch_id", tischID)
		return nil, err
	}

	log.Info("Retrieved pfandrueckgabe historie", "tisch_id", tischID, "historie_count", len(historie))
	return historie, nil
}
