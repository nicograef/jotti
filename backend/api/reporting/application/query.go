package application

import (
	"context"
	"sort"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/reporting"
	"github.com/nicograef/jotti/backend/domain/steuer"
	"github.com/rs/zerolog"
)

var ErrDatabase = db.ErrDatabase

type reportingRepo interface {
	GetReporting(ctx context.Context, kassensitzungNr int) (reporting.ReportingData, error)
	GetEigeneUebersicht(ctx context.Context, userID int, kassensitzungNr int) (reporting.EigeneUebersicht, error)
	GetLiveReporting(ctx context.Context, kassensitzungNr int) (reporting.LiveReportingData, error)
}

type kassensitzungenRepo interface {
	GetAllKassensitzungen(ctx context.Context) ([]kasse.Kassensitzung, error)
	GetOffeneKassensitzungNr(ctx context.Context) (int, error)
	GetOffeneKassensitzung(ctx context.Context) (*kasse.Kassensitzung, error)
}

type Query struct {
	ReportingRepo       reportingRepo
	KassensitzungenRepo kassensitzungenRepo
}

func (q Query) GetReporting(ctx context.Context, kassensitzungNr int) (reporting.ReportingData, error) {
	log := zerolog.Ctx(ctx)

	data, err := q.ReportingRepo.GetReporting(ctx, kassensitzungNr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get reporting")
		return reporting.ReportingData{}, ErrDatabase
	}

	data.UmsatzProSteuersatz = berechneUmsatzProSteuersatz(data.UmsatzProSteuersatz)

	log.Info().Msg("Retrieved reporting")
	return data, nil
}

func berechneUmsatzProSteuersatz(bruttoProSatz []reporting.UmsatzSteuersatz) []reporting.UmsatzSteuersatz {
	if len(bruttoProSatz) == 0 {
		return []reporting.UmsatzSteuersatz{}
	}

	aggregiert := make(map[steuer.Steuersatz]reporting.UmsatzSteuersatz, len(bruttoProSatz))
	for _, eintrag := range bruttoProSatz {
		aufteilungen := steuer.Aufteilen(eintrag.BruttoCents, eintrag.Satz)
		for _, aufteilung := range aufteilungen {
			current := aggregiert[aufteilung.Satz]
			current.Satz = aufteilung.Satz
			current.BruttoCents += aufteilung.Brutto
			current.NettoCents += aufteilung.Netto
			current.SteuerCents += aufteilung.Steuer
			aggregiert[aufteilung.Satz] = current
		}
	}

	if len(aggregiert) == 0 {
		return []reporting.UmsatzSteuersatz{}
	}

	orderedSaetze := []steuer.Steuersatz{
		steuer.RegelSteuersatz,
		steuer.ErmaessigtSteuersatz,
		steuer.BefreitSteuersatz,
	}

	out := make([]reporting.UmsatzSteuersatz, 0, len(aggregiert))
	for _, satz := range orderedSaetze {
		if eintrag, ok := aggregiert[satz]; ok {
			out = append(out, eintrag)
			delete(aggregiert, satz)
		}
	}

	if len(aggregiert) == 0 {
		return out
	}

	restSaetze := make([]steuer.Steuersatz, 0, len(aggregiert))
	for satz := range aggregiert {
		restSaetze = append(restSaetze, satz)
	}
	sort.Slice(restSaetze, func(i, j int) bool {
		return restSaetze[i] < restSaetze[j]
	})
	for _, satz := range restSaetze {
		out = append(out, aggregiert[satz])
	}

	return out
}

func (q Query) GetAllKassensitzungen(ctx context.Context) ([]kasse.Kassensitzung, error) {
	log := zerolog.Ctx(ctx)

	data, err := q.KassensitzungenRepo.GetAllKassensitzungen(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get all kassensitzungen")
		return nil, ErrDatabase
	}

	log.Info().Msg("Retrieved all kassensitzungen")
	return data, nil
}

func (q Query) GetEigeneUebersicht(ctx context.Context, userID int) (reporting.EigeneUebersicht, error) {
	log := zerolog.Ctx(ctx)

	kassensitzungNr, err := q.KassensitzungenRepo.GetOffeneKassensitzungNr(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get offene kassensitzung nr")
		return reporting.EigeneUebersicht{}, ErrDatabase
	}

	data, err := q.ReportingRepo.GetEigeneUebersicht(ctx, userID, kassensitzungNr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get eigene uebersicht")
		return reporting.EigeneUebersicht{}, ErrDatabase
	}

	log.Info().Msg("Retrieved eigene uebersicht")
	return data, nil
}

func (q Query) GetLiveReporting(ctx context.Context) (*reporting.LiveReportingData, error) {
	log := zerolog.Ctx(ctx)

	ks, err := q.KassensitzungenRepo.GetOffeneKassensitzung(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get offene kassensitzung")
		return nil, ErrDatabase
	}
	if ks == nil {
		return nil, nil
	}

	data, err := q.ReportingRepo.GetLiveReporting(ctx, ks.ZNr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get live reporting")
		return nil, ErrDatabase
	}

	data.Bezeichnung = ks.Bezeichnung
	data.Datum = ks.Datum

	log.Info().Msg("Retrieved live reporting")
	return &data, nil
}
