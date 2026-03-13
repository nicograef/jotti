package reporting_repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/nicograef/jotti/backend/domain/reporting"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

type Repository struct {
	q *dbgen.Queries
}

func NewRepository(db *sql.DB) Repository {
	return Repository{q: dbgen.New(db)}
}

func (r Repository) GetDashboardData(ctx context.Context) (reporting.DashboardData, error) {
	stats, err := r.q.GetDashboardStats(ctx)
	if err != nil {
		return reporting.DashboardData{}, err
	}

	offeneTische, err := r.q.GetOffeneTische(ctx)
	if err != nil {
		return reporting.DashboardData{}, err
	}

	return reporting.DashboardData{
		GesamtUmsatzCents:        stats.GesamtUmsatzCents,
		AnzahlOffeneTische:       offeneTische,
		AnzahlBestellungen:       stats.AnzahlBestellungen,
		AnzahlStornierungen:      stats.AnzahlStornierungen,
		GesamtBestellungenCents:  stats.GesamtBestellungenCents,
		GesamtStornierungenCents: stats.GesamtStornierungenCents,
	}, nil
}

// stornierungEventData represents the JSONB structure of a fat stornierung event.
type stornierungEventData struct {
	GesamtStornierungCents int                             `json:"gesamtStornierungCents"`
	Kommentar              string                          `json:"kommentar"`
	Positionen             []reporting.StornierungPosition `json:"positionen"`
}

func (r Repository) GetTagesabrechnung(ctx context.Context, von, bis time.Time) (reporting.TagesabrechnungData, error) {
	stats, err := r.q.GetAbrechnungStats(ctx, dbgen.GetAbrechnungStatsParams{Von: von, Bis: bis})
	if err != nil {
		return reporting.TagesabrechnungData{}, err
	}

	offeneSaldi, err := r.q.GetOffeneSaldi(ctx)
	if err != nil {
		return reporting.TagesabrechnungData{}, err
	}

	umsatzRows, err := r.q.GetUmsatzProServicekraft(ctx, dbgen.GetUmsatzProServicekraftParams{Von: von, Bis: bis})
	if err != nil {
		return reporting.TagesabrechnungData{}, err
	}
	umsatz := make([]reporting.UmsatzServicekraft, len(umsatzRows))
	for i, row := range umsatzRows {
		umsatz[i] = reporting.UmsatzServicekraft{
			UserID:          row.UserID,
			UserName:        row.UserName,
			ZahlungenCents:  row.ZahlungenCents,
			AnzahlZahlungen: row.AnzahlZahlungen,
		}
	}

	stornoRows, err := r.q.GetStornierungen(ctx, dbgen.GetStornierungenParams{Von: von, Bis: bis})
	if err != nil {
		return reporting.TagesabrechnungData{}, err
	}
	stornierungen := make([]reporting.StornierungDetail, len(stornoRows))
	for i, row := range stornoRows {
		var data stornierungEventData
		if err := json.Unmarshal(row.Data, &data); err != nil {
			return reporting.TagesabrechnungData{}, err
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

	return reporting.TagesabrechnungData{
		Zeitraum:                 reporting.Zeitraum{Von: von, Bis: bis},
		GesamtUmsatzCents:        stats.GesamtUmsatzCents,
		GesamtBestellungenCents:  stats.GesamtBestellungenCents,
		GesamtStornierungenCents: stats.GesamtStornierungenCents,
		OffeneSaldiCents:         offeneSaldi,
		AnzahlBestellungen:       stats.AnzahlBestellungen,
		AnzahlStornierungen:      stats.AnzahlStornierungen,
		UmsatzProServicekraft:    umsatz,
		Stornierungen:            stornierungen,
	}, nil
}
