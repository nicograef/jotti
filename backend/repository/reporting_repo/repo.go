package reporting_repo

import (
	"context"
	"database/sql"
	"encoding/json"

	"golang.org/x/sync/errgroup"

	"github.com/nicograef/jotti/backend/domain/reporting"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

type Repository struct {
	q *dbgen.Queries
}

func NewRepository(db *sql.DB) Repository {
	return Repository{q: dbgen.New(db)}
}

// stornierungEventData represents the JSONB structure of a fat stornierung event.
type stornierungEventData struct {
	GesamtStornierungCents int                             `json:"gesamtStornierungCents"`
	Kommentar              string                          `json:"kommentar"`
	Positionen             []reporting.StornierungPosition `json:"positionen"`
}

func (r Repository) GetReporting(ctx context.Context, zeitraum reporting.Zeitraum) (reporting.ReportingData, error) {
	var (
		stats                  dbgen.GetReportingStatsRow
		offeneSaldi            int
		offeneTische           int
		ausstehendAuszahlungen int
		umsatzRows             []dbgen.GetUmsatzProServicekraftRow
		tischRows              []dbgen.GetUmsatzProTischRow
		stornoRows             []dbgen.GetStornierungenRow
	)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var err error
		stats, err = r.q.GetReportingStats(ctx, dbgen.GetReportingStatsParams{Von: zeitraum.Von, Bis: zeitraum.Bis})
		return err
	})
	g.Go(func() error {
		var err error
		offeneSaldi, err = r.q.GetOffeneSaldi(ctx)
		return err
	})
	g.Go(func() error {
		var err error
		offeneTische, err = r.q.GetOffeneTische(ctx)
		return err
	})
	g.Go(func() error {
		var err error
		ausstehendAuszahlungen, err = r.q.GetAusstehendAuszahlungen(ctx)
		return err
	})
	g.Go(func() error {
		var err error
		umsatzRows, err = r.q.GetUmsatzProServicekraft(ctx, dbgen.GetUmsatzProServicekraftParams{Von: zeitraum.Von, Bis: zeitraum.Bis})
		return err
	})
	g.Go(func() error {
		var err error
		tischRows, err = r.q.GetUmsatzProTisch(ctx, dbgen.GetUmsatzProTischParams{Von: zeitraum.Von, Bis: zeitraum.Bis})
		return err
	})
	g.Go(func() error {
		var err error
		stornoRows, err = r.q.GetStornierungen(ctx, dbgen.GetStornierungenParams{Von: zeitraum.Von, Bis: zeitraum.Bis})
		return err
	})

	if err := g.Wait(); err != nil {
		return reporting.ReportingData{}, err
	}

	umsatz := make([]reporting.UmsatzServicekraft, len(umsatzRows))
	for i, row := range umsatzRows {
		userName, _ := row.UserName.(string)
		umsatz[i] = reporting.UmsatzServicekraft{
			UserID:            row.UserID,
			UserName:          userName,
			ZahlungenCents:    row.ZahlungenCents,
			AuszahlungenCents: row.AuszahlungenCents,
			AnzahlZahlungen:   row.AnzahlZahlungen,
		}
	}

	tische := make([]reporting.UmsatzTisch, len(tischRows))
	for i, row := range tischRows {
		tische[i] = reporting.UmsatzTisch{
			TischID:           row.TischID,
			TischName:         row.TischName,
			ZahlungenCents:    row.ZahlungenCents,
			AuszahlungenCents: row.AuszahlungenCents,
			AnzahlZahlungen:   row.AnzahlZahlungen,
		}
	}

	stornierungen := make([]reporting.StornierungDetail, len(stornoRows))
	for i, row := range stornoRows {
		var data stornierungEventData
		if err := json.Unmarshal(row.Data, &data); err != nil {
			return reporting.ReportingData{}, err
		}
		positionen := data.Positionen
		if positionen == nil {
			positionen = []reporting.StornierungPosition{}
		}
		stornierungen[i] = reporting.StornierungDetail{
			Zeitpunkt:   row.Timestamp,
			TischID:     row.TischID,
			TischName:   row.TischName,
			UserID:      row.UserID,
			UserName:    row.UserName,
			BetragCents: data.GesamtStornierungCents,
			Kommentar:   data.Kommentar,
			Positionen:  positionen,
		}
	}

	return reporting.ReportingData{
		Zeitraum: zeitraum,
		Summary: reporting.Summary{
			GesamtUmsatzCents:           int(stats.GesamtUmsatzCents),
			GesamtAuszahlungenCents:     stats.GesamtAuszahlungenCents,
			GesamtBestellungenCents:     stats.GesamtBestellungenCents,
			GesamtStornierungenCents:    stats.GesamtStornierungenCents,
			OffeneSaldiCents:            offeneSaldi,
			AusstehendAuszahlungenCents: ausstehendAuszahlungen,
			AnzahlOffeneTische:          offeneTische,
			AnzahlBestellungen:          stats.AnzahlBestellungen,
			AnzahlStornierungen:         stats.AnzahlStornierungen,
		},
		Breakdowns: reporting.Breakdowns{
			UmsatzProServicekraft: umsatz,
			UmsatzProTisch:        tische,
		},
		Stornierungen: stornierungen,
	}, nil
}

func (r Repository) GetEigeneUebersicht(ctx context.Context, userID int) (reporting.EigeneUebersicht, error) {
	row, err := r.q.GetEigeneUebersicht(ctx, userID)
	if err != nil {
		return reporting.EigeneUebersicht{}, err
	}

	return reporting.EigeneUebersicht{
		AnzahlBestellungen: row.AnzahlBestellungen,
		BestellungenCents:  row.BestellungenCents,
		AnzahlZahlungen:    row.AnzahlZahlungen,
		ZahlungenCents:     row.ZahlungenCents,
	}, nil
}
