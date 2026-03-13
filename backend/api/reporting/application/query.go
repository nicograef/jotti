package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/nicograef/jotti/backend/sqlc/dbgen"
	"github.com/rs/zerolog"
)

type Query struct {
	db *sql.DB
	q  *dbgen.Queries
}

func NewQuery(db *sql.DB) Query {
	return Query{db: db, q: dbgen.New(db)}
}

// Domain types for responses

type DashboardData struct {
	GesamtUmsatzCents        int `json:"gesamtUmsatzCents"`
	AnzahlOffeneTische       int `json:"anzahlOffeneTische"`
	AnzahlBestellungen       int `json:"anzahlBestellungen"`
	AnzahlStornierungen      int `json:"anzahlStornierungen"`
	GesamtBestellungenCents  int `json:"gesamtBestellungenCents"`
	GesamtStornierungenCents int `json:"gesamtStornierungenCents"`
}

type Zeitraum struct {
	Von time.Time `json:"von"`
	Bis time.Time `json:"bis"`
}

type UmsatzServicekraft struct {
	UserID          int    `json:"userId"`
	UserName        string `json:"userName"`
	ZahlungenCents  int    `json:"zahlungenCents"`
	AnzahlZahlungen int    `json:"anzahlZahlungen"`
}

type StornierungPosition struct {
	ProduktName  string `json:"produktName"`
	VarianteName string `json:"varianteName"`
	Menge        int    `json:"menge"`
	Einzelpreis  int    `json:"einzelpreis"`
}

type StornierungDetail struct {
	Zeitpunkt   time.Time             `json:"zeitpunkt"`
	TischID     int                   `json:"tischId"`
	TischName   string                `json:"tischName"`
	UserID      int                   `json:"userId"`
	UserName    string                `json:"userName"`
	BetragCents int                   `json:"betragCents"`
	Kommentar   string                `json:"kommentar"`
	Positionen  []StornierungPosition `json:"positionen"`
}

type TagesabrechnungData struct {
	Zeitraum                 Zeitraum             `json:"zeitraum"`
	GesamtUmsatzCents        int                  `json:"gesamtUmsatzCents"`
	GesamtBestellungenCents  int                  `json:"gesamtBestellungenCents"`
	GesamtStornierungenCents int                  `json:"gesamtStornierungenCents"`
	OffeneSaldiCents         int                  `json:"offeneSaldiCents"`
	AnzahlBestellungen       int                  `json:"anzahlBestellungen"`
	AnzahlStornierungen      int                  `json:"anzahlStornierungen"`
	UmsatzProServicekraft    []UmsatzServicekraft `json:"umsatzProServicekraft"`
	Stornierungen            []StornierungDetail  `json:"stornierungen"`
}

func (q Query) GetDashboardData(ctx context.Context) (DashboardData, error) {
	log := zerolog.Ctx(ctx)

	stats, err := q.q.GetDashboardStats(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get dashboard stats")
		return DashboardData{}, err
	}

	offeneTische, err := q.q.GetOffeneTische(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get offene tische")
		return DashboardData{}, err
	}

	return DashboardData{
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
	GesamtStornierungCents int                   `json:"gesamtStornierungCents"`
	Kommentar              string                `json:"kommentar"`
	Positionen             []StornierungPosition `json:"positionen"`
}

func (q Query) GetTagesabrechnung(ctx context.Context, von, bis time.Time) (TagesabrechnungData, error) {
	log := zerolog.Ctx(ctx)

	stats, err := q.q.GetAbrechnungStats(ctx, dbgen.GetAbrechnungStatsParams{Von: von, Bis: bis})
	if err != nil {
		log.Error().Err(err).Msg("Failed to get abrechnung stats")
		return TagesabrechnungData{}, err
	}

	offeneSaldi, err := q.q.GetOffeneSaldi(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get offene saldi")
		return TagesabrechnungData{}, err
	}

	umsatzRows, err := q.q.GetUmsatzProServicekraft(ctx, dbgen.GetUmsatzProServicekraftParams{Von: von, Bis: bis})
	if err != nil {
		log.Error().Err(err).Msg("Failed to get umsatz pro servicekraft")
		return TagesabrechnungData{}, err
	}
	umsatz := make([]UmsatzServicekraft, len(umsatzRows))
	for i, row := range umsatzRows {
		umsatz[i] = UmsatzServicekraft{
			UserID:          row.UserID,
			UserName:        row.UserName,
			ZahlungenCents:  row.ZahlungenCents,
			AnzahlZahlungen: row.AnzahlZahlungen,
		}
	}

	stornoRows, err := q.q.GetStornierungen(ctx, dbgen.GetStornierungenParams{Von: von, Bis: bis})
	if err != nil {
		log.Error().Err(err).Msg("Failed to get stornierungen")
		return TagesabrechnungData{}, err
	}
	stornierungen := make([]StornierungDetail, len(stornoRows))
	for i, row := range stornoRows {
		var data stornierungEventData
		if err := json.Unmarshal(row.Data, &data); err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal stornierung event data")
			return TagesabrechnungData{}, err
		}
		positionen := data.Positionen
		if positionen == nil {
			positionen = []StornierungPosition{}
		}
		stornierungen[i] = StornierungDetail{
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

	return TagesabrechnungData{
		Zeitraum:                 Zeitraum{Von: von, Bis: bis},
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
