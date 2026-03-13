package application

import (
	"context"
	"time"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/reporting"
	"github.com/rs/zerolog"
)

var ErrDatabase = db.ErrDatabase

type reportingRepo interface {
	GetDashboardData(ctx context.Context) (reporting.DashboardData, error)
	GetTagesabrechnung(ctx context.Context, von, bis time.Time) (reporting.TagesabrechnungData, error)
}

type Query struct {
	ReportingRepo reportingRepo
}

func (q Query) GetDashboardData(ctx context.Context) (reporting.DashboardData, error) {
	log := zerolog.Ctx(ctx)

	data, err := q.ReportingRepo.GetDashboardData(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get dashboard data")
		return reporting.DashboardData{}, ErrDatabase
	}

	log.Info().Msg("Retrieved dashboard data")
	return data, nil
}

func (q Query) GetTagesabrechnung(ctx context.Context, von, bis time.Time) (reporting.TagesabrechnungData, error) {
	log := zerolog.Ctx(ctx)

	data, err := q.ReportingRepo.GetTagesabrechnung(ctx, von, bis)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get tagesabrechnung")
		return reporting.TagesabrechnungData{}, ErrDatabase
	}

	log.Info().Msg("Retrieved tagesabrechnung")
	return data, nil
}
