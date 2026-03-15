package application

import (
	"context"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/reporting"
	"github.com/rs/zerolog"
)

var ErrDatabase = db.ErrDatabase

type reportingRepo interface {
	GetReporting(ctx context.Context, zeitraum reporting.Zeitraum) (reporting.ReportingData, error)
	GetEigeneUebersicht(ctx context.Context, userID int) (reporting.EigeneUebersicht, error)
}

type Query struct {
	ReportingRepo reportingRepo
}

func (q Query) GetReporting(ctx context.Context, zeitraum reporting.Zeitraum) (reporting.ReportingData, error) {
	log := zerolog.Ctx(ctx)

	data, err := q.ReportingRepo.GetReporting(ctx, zeitraum)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get reporting")
		return reporting.ReportingData{}, ErrDatabase
	}

	log.Info().Msg("Retrieved reporting")
	return data, nil
}

func (q Query) GetEigeneUebersicht(ctx context.Context, userID int) (reporting.EigeneUebersicht, error) {
	log := zerolog.Ctx(ctx)

	data, err := q.ReportingRepo.GetEigeneUebersicht(ctx, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get eigene uebersicht")
		return reporting.EigeneUebersicht{}, ErrDatabase
	}

	log.Info().Msg("Retrieved eigene uebersicht")
	return data, nil
}
