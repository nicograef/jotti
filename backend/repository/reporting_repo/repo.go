package reporting_repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"golang.org/x/sync/errgroup"

	"github.com/nicograef/jotti/backend/db"
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

// stornierungEventData holds the shared JSONB fields of the storno events. The query normalizes the total
// (gesamtStornierungCents or gesamtCents) into its own column.
type stornierungEventData struct {
	Kommentar  string                    `json:"kommentar"`
	Positionen []stornierungPositionJSON `json:"positionen"`
}

// servicekraftRefJSON decodes one entry of the betroffene column of GetStornierungen.
type servicekraftRefJSON struct {
	UserID   int    `json:"userId"`
	UserName string `json:"userName"`
	Name     string `json:"name"`
}

func (r Repository) GetReporting(ctx context.Context, kassensitzungNr int) (reporting.ReportingData, error) {
	var (
		stats        dbgen.GetReportingStatsRow
		kassiertRows []dbgen.GetKassiertProServicekraftRow
		zeilenRows   []dbgen.GetUmsatzPositionszeilenRow
		stornoRows   []dbgen.GetStornierungenRow
		metadatenRow dbgen.GetKassensitzungMetadatenRow
	)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var err error
		stats, err = r.q.GetReportingStats(ctx, kassensitzungNr)
		return err
	})
	g.Go(func() error {
		var err error
		kassiertRows, err = r.q.GetKassiertProServicekraft(ctx, kassensitzungNr)
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
	g.Go(func() error {
		var err error
		metadatenRow, err = r.q.GetKassensitzungMetadaten(ctx, kassensitzungNr)
		return err
	})

	if err := g.Wait(); err != nil {
		return reporting.ReportingData{}, db.Error(err)
	}

	stornierungen, err := toStornierungen(stornoRows)
	if err != nil {
		return reporting.ReportingData{}, err
	}

	metadaten, err := toMetadaten(metadatenRow)
	if err != nil {
		return reporting.ReportingData{}, err
	}

	// Raw gross position rows; the application layer replaces them with the aggregated VAT breakdown.
	umsatzProSteuersatz := make([]reporting.UmsatzSteuersatz, len(zeilenRows))
	for i, row := range zeilenRows {
		umsatzProSteuersatz[i] = reporting.UmsatzSteuersatz{
			Satz:        steuer.Steuersatz(row.Steuersatz),
			BruttoCents: row.BruttoCents,
		}
	}

	return reporting.ReportingData{
		KassensitzungNr: kassensitzungNr,
		Metadaten:       metadaten,
		Summary:         toSummary(stats),
		Breakdowns: reporting.Breakdowns{
			AbrechnungProServicekraft: toAbrechnungServicekraft(kassiertRows),
		},
		UmsatzProSteuersatz: umsatzProSteuersatz,
		Stornierungen:       stornierungen,
	}, nil
}

// kassensturzDataJSON decodes the count difference of the kassensturz-durchgefuehrt:v1 event.
type kassensturzDataJSON struct {
	DifferenzCents int `json:"differenzCents"`
}

func toMetadaten(row dbgen.GetKassensitzungMetadatenRow) (reporting.Metadaten, error) {
	metadaten := reporting.Metadaten{}

	if row.EroeffnetAm.Valid {
		eroeffnetAm := row.EroeffnetAm.Time
		metadaten.EroeffnetAm = &eroeffnetAm
	}
	if row.AbgeschlossenAm.Valid {
		abgeschlossenAm := row.AbgeschlossenAm.Time
		metadaten.AbgeschlossenAm = &abgeschlossenAm
	}
	if row.AbgeschlossenVon.Valid {
		metadaten.AbgeschlossenVon = row.AbgeschlossenVon.String
	}
	// Without a Kassensturz the query yields the JSON literal 'null', which decodes to a nil pointer.
	var data *kassensturzDataJSON
	if err := json.Unmarshal(row.KassensturzData, &data); err != nil {
		return reporting.Metadaten{}, fmt.Errorf("unmarshal kassensturz data: %w", err)
	}
	if data != nil {
		differenzCents := data.DifferenzCents
		metadaten.KassensturzDifferenzCents = &differenzCents
	}

	return metadaten, nil
}

func (r Repository) GetLiveReporting(ctx context.Context, kassensitzungNr int) (reporting.LiveReportingData, error) {
	var (
		stats            dbgen.GetReportingStatsRow
		offeneSaldi      int
		offeneTischeRows []dbgen.GetOffeneTischeDetailsRow
		kassiertRows     []dbgen.GetKassiertProServicekraftRow
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
		kassiertRows, err = r.q.GetKassiertProServicekraft(ctx, kassensitzungNr)
		return err
	})
	g.Go(func() error {
		var err error
		stornoRows, err = r.q.GetStornierungen(ctx, kassensitzungNr)
		return err
	})

	if err := g.Wait(); err != nil {
		return reporting.LiveReportingData{}, db.Error(err)
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
			AbrechnungProServicekraft: toAbrechnungServicekraft(kassiertRows),
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

// toAbrechnungServicekraft leaves returns, storno count and Abzugeben at zero; the application layer derives them
// from the storno rows.
func toAbrechnungServicekraft(rows []dbgen.GetKassiertProServicekraftRow) []reporting.AbrechnungServicekraft {
	abrechnung := make([]reporting.AbrechnungServicekraft, len(rows))
	for i, row := range rows {
		abrechnung[i] = reporting.AbrechnungServicekraft{
			UserID:          row.UserID,
			UserName:        row.UserName,
			Name:            row.Name,
			KassiertCents:   row.KassiertCents,
			AnzahlZahlungen: row.AnzahlZahlungen,
		}
	}
	return abrechnung
}

// toBetroffene needs no fallback: the query guarantees a non-empty list by falling back to the actor.
func toBetroffene(raw json.RawMessage) ([]reporting.ServicekraftRef, error) {
	var refs []servicekraftRefJSON
	if err := json.Unmarshal(raw, &refs); err != nil {
		return nil, fmt.Errorf("unmarshal betroffene: %w", err)
	}
	out := make([]reporting.ServicekraftRef, len(refs))
	for i, ref := range refs {
		out[i] = reporting.ServicekraftRef{
			UserID:   ref.UserID,
			UserName: ref.UserName,
			Name:     ref.Name,
		}
	}
	return out, nil
}

func toStornierungen(rows []dbgen.GetStornierungenRow) ([]reporting.StornierungDetail, error) {
	stornierungen := make([]reporting.StornierungDetail, len(rows))
	for i, row := range rows {
		var data stornierungEventData
		if err := json.Unmarshal(row.Data, &data); err != nil {
			return nil, fmt.Errorf("unmarshal stornierung data: %w", err)
		}
		betroffene, err := toBetroffene(row.Betroffene)
		if err != nil {
			return nil, err
		}
		positionen := toStornierungPositionen(data.Positionen)
		stornierungen[i] = reporting.StornierungDetail{
			Zeitpunkt:    row.Timestamp,
			Quelle:       row.Quelle,
			BarRueckgabe: row.BarRueckgabe,
			TischID:      row.TischID,
			TischName:    row.TischName,
			Akteur: reporting.ServicekraftRef{
				UserID:   row.UserID,
				UserName: row.UserName,
				Name:     row.Name,
			},
			Betroffene:  betroffene,
			BetragCents: row.BetragCents,
			Kommentar:   data.Kommentar,
			Positionen:  positionen,
		}
	}
	return stornierungen, nil
}

// GetProduktStatistik returns flat per-variant rows for the application layer to group. It stays outside the
// GetReporting errgroup so the settlement and the live path share it.
func (r Repository) GetProduktStatistik(ctx context.Context, kassensitzungNr int) ([]reporting.ProduktStatistikZeile, error) {
	rows, err := r.q.GetProduktStatistik(ctx, kassensitzungNr)
	if err != nil {
		return nil, db.Error(err)
	}

	zeilen := make([]reporting.ProduktStatistikZeile, len(rows))
	for i, row := range rows {
		zeilen[i] = reporting.ProduktStatistikZeile{
			Kategorie:        row.Kategorie,
			ProduktName:      row.ProduktName,
			VarianteID:       row.VarianteID,
			VarianteName:     row.VarianteName,
			AusgegebeneMenge: row.AusgegebeneMenge,
			UmsatzCents:      row.UmsatzCents,
		}
	}
	return zeilen, nil
}

func (r Repository) GetEigeneUebersicht(ctx context.Context, userID int, kassensitzungNr int) (reporting.EigeneUebersicht, error) {
	row, err := r.q.GetEigeneUebersicht(ctx, dbgen.GetEigeneUebersichtParams{
		UserID:          userID,
		KassensitzungNr: kassensitzungNr,
	})
	if err != nil {
		return reporting.EigeneUebersicht{}, db.Error(err)
	}

	return reporting.EigeneUebersicht{
		AnzahlBestellungen: row.AnzahlBestellungen,
		BestellungenCents:  row.BestellungenCents,
		AnzahlZahlungen:    row.AnzahlZahlungen,
		ZahlungenCents:     row.ZahlungenCents,
		AnzahlRuecknahmen:  row.AnzahlRuecknahmen,
		RuecknahmenCents:   row.RuecknahmenCents,
		AbzugebenCents:     row.AbzugebenCents,
	}, nil
}
