package reporting_repo

import (
	"context"
	"database/sql"
	"encoding/json"

	"golang.org/x/sync/errgroup"

	"github.com/nicograef/jotti/backend/domain/reporting"
	"github.com/nicograef/jotti/backend/domain/steuer"
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
	ProduktName      string `json:"produktName"`
	VarianteName     string `json:"varianteName"`
	Menge            int    `json:"menge"`
	EinzelpreisCents int    `json:"einzelpreisCents"`
}

// stornierungEventData represents the shared JSONB fields of the storno events.
// Der Gesamtbetrag wird in der Query normalisiert (gesamtStornierungCents bzw. gesamtCents)
// und kommt als eigene Spalte; hier nur Kommentar und Positionen.
type stornierungEventData struct {
	Kommentar  string                    `json:"kommentar"`
	Positionen []stornierungPositionJSON `json:"positionen"`
}

func (r Repository) GetReporting(ctx context.Context, kassensitzungNr int) (reporting.ReportingData, error) {
	var (
		stats      dbgen.GetReportingStatsRow
		umsatzRows []dbgen.GetUmsatzProServicekraftRow
		zeilenRows []dbgen.GetUmsatzPositionszeilenRow
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
		zeilenRows, err = r.q.GetUmsatzPositionszeilen(ctx, kassensitzungNr)
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

	stornierungen, err := toStornierungen(stornoRows)
	if err != nil {
		return reporting.ReportingData{}, err
	}

	// Unaggregierte Brutto-Positionszeilen; die Anwendungsschicht ersetzt sie
	// durch die aggregierte USt-Aufschlüsselung (steuer.Aufteilen je Zeile).
	umsatzProSteuersatz := make([]reporting.UmsatzSteuersatz, len(zeilenRows))
	for i, row := range zeilenRows {
		umsatzProSteuersatz[i] = reporting.UmsatzSteuersatz{
			Satz:        steuer.Steuersatz(row.Steuersatz),
			BruttoCents: row.BruttoCents,
		}
	}

	return reporting.ReportingData{
		KassensitzungNr: kassensitzungNr,
		Summary:         toSummary(stats),
		Breakdowns: reporting.Breakdowns{
			UmsatzProServicekraft: toUmsatzServicekraft(umsatzRows),
		},
		UmsatzProSteuersatz: umsatzProSteuersatz,
		Stornierungen:       stornierungen,
	}, nil
}

func (r Repository) GetLiveReporting(ctx context.Context, kassensitzungNr int) (reporting.LiveReportingData, error) {
	var (
		stats            dbgen.GetReportingStatsRow
		offeneSaldi      int
		offeneTischeRows []dbgen.GetOffeneTischeDetailsRow
		umsatzRows       []dbgen.GetUmsatzProServicekraftRow
		stornoRows       []dbgen.GetStornierungenRow
	)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var err error
		stats, err = r.q.GetReportingStats(ctx, kassensitzungNr)
		return err
	})
	g.Go(func() error {
		var err error
		offeneSaldi, err = r.q.GetOffeneSaldi(ctx, kassensitzungNr)
		return err
	})
	g.Go(func() error {
		var err error
		offeneTischeRows, err = r.q.GetOffeneTischeDetails(ctx, kassensitzungNr)
		return err
	})
	g.Go(func() error {
		var err error
		umsatzRows, err = r.q.GetUmsatzProServicekraft(ctx, kassensitzungNr)
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

	stornierungen, err := toStornierungen(stornoRows)
	if err != nil {
		return reporting.LiveReportingData{}, err
	}

	return reporting.LiveReportingData{
		KassensitzungNr:  kassensitzungNr,
		OffeneTische:     offeneTische,
		OffeneSaldiCents: offeneSaldi,
		Summary:          toSummary(stats),
		Breakdowns: reporting.Breakdowns{
			UmsatzProServicekraft: toUmsatzServicekraft(umsatzRows),
		},
		Stornierungen: stornierungen,
	}, nil
}

func toStornierungPosition(p stornierungPositionJSON) reporting.StornierungPosition {
	return reporting.StornierungPosition{
		ProduktName:      p.ProduktName,
		VarianteName:     p.VarianteName,
		Menge:            p.Menge,
		EinzelpreisCents: p.EinzelpreisCents,
	}
}

func toStornierungPositionen(positionen []stornierungPositionJSON) []reporting.StornierungPosition {
	out := make([]reporting.StornierungPosition, len(positionen))
	for i, p := range positionen {
		out[i] = toStornierungPosition(p)
	}
	return out
}

func toSummary(stats dbgen.GetReportingStatsRow) reporting.Summary {
	return reporting.Summary{
		GesamtUmsatzCents:        stats.GesamtUmsatzCents,
		GesamtBestellungenCents:  stats.GesamtBestellungenCents,
		GesamtStornierungenCents: stats.GesamtStornierungenCents,
		GeldtransitCents:         stats.GesamtGeldtransitCents,
		AnzahlBestellungen:       stats.AnzahlBestellungen,
		AnzahlStornierungen:      stats.AnzahlStornierungen,
		AnzahlDirektverkaeufe:    stats.AnzahlDirektverkaeufe,
		DirektverkaufUmsatzCents: stats.DirektverkaufUmsatzCents,
	}
}

func toUmsatzServicekraft(rows []dbgen.GetUmsatzProServicekraftRow) []reporting.UmsatzServicekraft {
	umsatz := make([]reporting.UmsatzServicekraft, len(rows))
	for i, row := range rows {
		umsatz[i] = reporting.UmsatzServicekraft{
			UserID:          row.UserID,
			UserName:        row.UserName,
			Name:            row.Name,
			ZahlungenCents:  row.ZahlungenCents,
			AnzahlZahlungen: row.AnzahlZahlungen,
		}
	}
	return umsatz
}

func toStornierungen(rows []dbgen.GetStornierungenRow) ([]reporting.StornierungDetail, error) {
	stornierungen := make([]reporting.StornierungDetail, len(rows))
	for i, row := range rows {
		var data stornierungEventData
		if err := json.Unmarshal(row.Data, &data); err != nil {
			return nil, err
		}
		positionen := toStornierungPositionen(data.Positionen)
		if positionen == nil {
			positionen = []reporting.StornierungPosition{}
		}
		stornierungen[i] = reporting.StornierungDetail{
			Zeitpunkt:    row.Timestamp,
			Quelle:       row.Quelle,
			BarRueckgabe: row.BarRueckgabe,
			TischID:      row.TischID,
			TischName:    row.TischName,
			UserID:       row.UserID,
			UserName:     row.UserName,
			Name:         row.Name,
			BetragCents:  row.BetragCents,
			Kommentar:    data.Kommentar,
			Positionen:   positionen,
		}
	}
	return stornierungen, nil
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
