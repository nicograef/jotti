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

// stornierungPositionJSON is used for deserializing position data from the stornierung event JSONB.
type stornierungPositionJSON struct {
	ProduktName  string `json:"produktName"`
	VarianteName string `json:"varianteName"`
	Menge        int    `json:"menge"`
	Einzelpreis  int    `json:"einzelpreis"`
}

// stornierungEventData represents the JSONB structure of a fat stornierung event.
type stornierungEventData struct {
	GesamtStornierungCents int                       `json:"gesamtStornierungCents"`
	Kommentar              string                    `json:"kommentar"`
	Positionen             []stornierungPositionJSON `json:"positionen"`
}

func (r Repository) GetReporting(ctx context.Context, kassensitzungNr int) (reporting.ReportingData, error) {
	var (
		stats      dbgen.GetReportingStatsRow
		umsatzRows []dbgen.GetUmsatzProServicekraftRow
		tischRows  []dbgen.GetUmsatzProTischRow
		stornoRows []dbgen.GetStornierungenRow
	)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var err error
		stats, err = r.q.GetReportingStats(ctx, kassensitzungNr)
		return err
	})
	g.Go(func() error {
		var err error
		umsatzRows, err = r.q.GetUmsatzProServicekraft(ctx, kassensitzungNr)
		return err
	})
	g.Go(func() error {
		var err error
		tischRows, err = r.q.GetUmsatzProTisch(ctx, kassensitzungNr)
		return err
	})
	g.Go(func() error {
		var err error
		stornoRows, err = r.q.GetStornierungen(ctx, kassensitzungNr)
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
		positionen := toStornierungPositionen(data.Positionen)
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
		KassensitzungNr: kassensitzungNr,
		Summary: reporting.Summary{
			GesamtUmsatzCents:        int(stats.GesamtUmsatzCents),
			GesamtAuszahlungenCents:  stats.GesamtAuszahlungenCents,
			GesamtBestellungenCents:  stats.GesamtBestellungenCents,
			GesamtStornierungenCents: stats.GesamtStornierungenCents,
			AnzahlBestellungen:       stats.AnzahlBestellungen,
			AnzahlStornierungen:      stats.AnzahlStornierungen,
		},
		Breakdowns: reporting.Breakdowns{
			UmsatzProServicekraft: umsatz,
			UmsatzProTisch:        tische,
		},
		Stornierungen: stornierungen,
	}, nil
}

func (r Repository) GetLiveReporting(ctx context.Context, kassensitzungNr int) (reporting.LiveReportingData, error) {
	var (
		stats                  dbgen.GetReportingStatsRow
		offeneSaldi            int
		ausstehendAuszahlungen int
		offeneTischeRows       []dbgen.GetOffeneTischeDetailsRow
		umsatzRows             []dbgen.GetUmsatzProServicekraftRow
		tischRows              []dbgen.GetUmsatzProTischRow
		stornoRows             []dbgen.GetStornierungenRow
	)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var err error
		stats, err = r.q.GetReportingStats(ctx, kassensitzungNr)
		return err
	})
	g.Go(func() error {
		var err error
		offeneSaldi, err = r.q.GetOffeneSaldi(ctx)
		return err
	})
	g.Go(func() error {
		var err error
		ausstehendAuszahlungen, err = r.q.GetAusstehendAuszahlungen(ctx)
		return err
	})
	g.Go(func() error {
		var err error
		offeneTischeRows, err = r.q.GetOffeneTischeDetails(ctx)
		return err
	})
	g.Go(func() error {
		var err error
		umsatzRows, err = r.q.GetUmsatzProServicekraft(ctx, kassensitzungNr)
		return err
	})
	g.Go(func() error {
		var err error
		tischRows, err = r.q.GetUmsatzProTisch(ctx, kassensitzungNr)
		return err
	})
	g.Go(func() error {
		var err error
		stornoRows, err = r.q.GetStornierungen(ctx, kassensitzungNr)
		return err
	})

	if err := g.Wait(); err != nil {
		return reporting.LiveReportingData{}, err
	}

	offeneTische := make([]reporting.OffenerTisch, len(offeneTischeRows))
	for i, row := range offeneTischeRows {
		offeneTische[i] = reporting.OffenerTisch{
			TischID:    row.TischID,
			TischName:  row.TischName,
			SaldoCents: row.SaldoCents,
		}
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
			return reporting.LiveReportingData{}, err
		}
		positionen := toStornierungPositionen(data.Positionen)
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

	return reporting.LiveReportingData{
		KassensitzungNr:             kassensitzungNr,
		OffeneTische:                offeneTische,
		OffeneSaldiCents:            offeneSaldi,
		AusstehendAuszahlungenCents: ausstehendAuszahlungen,
		Summary: reporting.Summary{
			GesamtUmsatzCents:        int(stats.GesamtUmsatzCents),
			GesamtAuszahlungenCents:  stats.GesamtAuszahlungenCents,
			GesamtBestellungenCents:  stats.GesamtBestellungenCents,
			GesamtStornierungenCents: stats.GesamtStornierungenCents,
			AnzahlBestellungen:       stats.AnzahlBestellungen,
			AnzahlStornierungen:      stats.AnzahlStornierungen,
		},
		Breakdowns: reporting.Breakdowns{
			UmsatzProServicekraft: umsatz,
			UmsatzProTisch:        tische,
		},
		Stornierungen: stornierungen,
	}, nil
}

func toStornierungPosition(p stornierungPositionJSON) reporting.StornierungPosition {
	return reporting.StornierungPosition{
		ProduktName:  p.ProduktName,
		VarianteName: p.VarianteName,
		Menge:        p.Menge,
		Einzelpreis:  p.Einzelpreis,
	}
}

func toStornierungPositionen(positionen []stornierungPositionJSON) []reporting.StornierungPosition {
	out := make([]reporting.StornierungPosition, len(positionen))
	for i, p := range positionen {
		out[i] = toStornierungPosition(p)
	}
	return out
}

func (r Repository) GetEigeneUebersicht(ctx context.Context, userID int, kassensitzungNr int) (reporting.EigeneUebersicht, error) {
	row, err := r.q.GetEigeneUebersicht(ctx, dbgen.GetEigeneUebersichtParams{
		UserID:          userID,
		KassensitzungNr: kassensitzungNr,
	})
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
