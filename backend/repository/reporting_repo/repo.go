package reporting_repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"golang.org/x/sync/errgroup"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/kasse"
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

// stornierungEventData holds the shared JSONB fields of the storno events. The query normalizes the total
// (gesamtStornierungCents or gesamtCents) into its own column.
type stornierungEventData struct {
	Kommentar  string                    `json:"kommentar"`
	Positionen []kasse.PositionEventData `json:"positionen"`
}

// servicekraftRefJSON decodes one entry of the betroffene column of GetStornierungen.
type servicekraftRefJSON struct {
	UserID   int    `json:"userId"`
	UserName string `json:"userName"`
	Name     string `json:"name"`
}

// gemeinsameAbfragen holds the rows the daily report and the live dashboard both read.
type gemeinsameAbfragen struct {
	stats        dbgen.GetReportingStatsRow
	kassiertRows []dbgen.GetKassiertProServicekraftRow
	stornoRows   []dbgen.GetStornierungenRow
	pfandRows    []dbgen.GetPfandProServicekraftRow
}

// starte registers the shared queries on g; the rows are valid once g.Wait returns nil.
func (a *gemeinsameAbfragen) starte(ctx context.Context, g *errgroup.Group, q *dbgen.Queries, kassensitzungNr int) {
	g.Go(func() error {
		var err error
		a.stats, err = q.GetReportingStats(ctx, kassensitzungNr)
		return err
	})
	g.Go(func() error {
		var err error
		a.kassiertRows, err = q.GetKassiertProServicekraft(ctx, kassensitzungNr)
		return err
	})
	g.Go(func() error {
		var err error
		a.stornoRows, err = q.GetStornierungen(ctx, kassensitzungNr)
		return err
	})
	g.Go(func() error {
		var err error
		a.pfandRows, err = q.GetPfandProServicekraft(ctx, kassensitzungNr)
		return err
	})
}

func (r Repository) GetReporting(ctx context.Context, kassensitzungNr int) (reporting.ReportingData, error) {
	var (
		gemeinsam    gemeinsameAbfragen
		zeilenRows   []dbgen.GetUmsatzPositionszeilenRow
		metadatenRow dbgen.GetKassensitzungMetadatenRow
	)

	g, ctx := errgroup.WithContext(ctx)

	gemeinsam.starte(ctx, g, r.q, kassensitzungNr)
	g.Go(func() error {
		var err error
		zeilenRows, err = r.q.GetUmsatzPositionszeilen(ctx, kassensitzungNr)
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

	stornierungen, err := toStornierungen(gemeinsam.stornoRows)
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
		Summary:         toSummary(gemeinsam.stats),
		Breakdowns: reporting.Breakdowns{
			AbrechnungProServicekraft: toAbrechnungServicekraft(gemeinsam.kassiertRows),
			PfandProServicekraft:      toPfandServicekraft(gemeinsam.pfandRows),
		},
		UmsatzProSteuersatz: umsatzProSteuersatz,
		Stornierungen:       stornierungen,
	}, nil
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
	var data *kasse.KassensturzDurchgefuehrtV1Data
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
		gemeinsam        gemeinsameAbfragen
		offeneTischeRows []dbgen.GetOffeneTischeDetailsRow
	)

	g, ctx := errgroup.WithContext(ctx)

	gemeinsam.starte(ctx, g, r.q, kassensitzungNr)
	g.Go(func() error {
		var err error
		offeneTischeRows, err = r.q.GetOffeneTischeDetails(ctx, kassensitzungNr)
		return err
	})

	if err := g.Wait(); err != nil {
		return reporting.LiveReportingData{}, db.Error(err)
	}

	offeneTische := make([]reporting.OffenerTisch, len(offeneTischeRows))
	offeneSaldi := 0
	for i, row := range offeneTischeRows {
		offeneTische[i] = reporting.OffenerTisch{
			TischID:    row.TischID,
			TischName:  row.TischName,
			SaldoCents: row.SaldoCents,
		}
		offeneSaldi += row.SaldoCents
	}

	stornierungen, err := toStornierungen(gemeinsam.stornoRows)
	if err != nil {
		return reporting.LiveReportingData{}, err
	}

	return reporting.LiveReportingData{
		KassensitzungNr:  kassensitzungNr,
		OffeneTische:     offeneTische,
		OffeneSaldiCents: offeneSaldi,
		Summary:          toSummary(gemeinsam.stats),
		Breakdowns: reporting.Breakdowns{
			AbrechnungProServicekraft: toAbrechnungServicekraft(gemeinsam.kassiertRows),
			PfandProServicekraft:      toPfandServicekraft(gemeinsam.pfandRows),
		},
		Stornierungen: stornierungen,
	}, nil
}

func toStornierungPosition(p kasse.PositionEventData) reporting.StornierungPosition {
	return reporting.StornierungPosition{
		ProduktName:      p.ProduktName,
		VarianteName:     p.VarianteName,
		Menge:            p.Menge,
		EinzelpreisCents: p.EinzelpreisCents,
	}
}

func toStornierungPositionen(positionen []kasse.PositionEventData) []reporting.StornierungPosition {
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
		AnzahlPfandrueckgaben:    stats.AnzahlPfandrueckgaben,
		PfandrueckgabenCents:     stats.PfandrueckgabenCents,
	}
}

func toPfandServicekraft(rows []dbgen.GetPfandProServicekraftRow) []reporting.PfandServicekraft {
	pfand := make([]reporting.PfandServicekraft, len(rows))
	for i, row := range rows {
		pfand[i] = reporting.PfandServicekraft{
			UserID:     row.UserID,
			UserName:   row.UserName,
			Name:       row.Name,
			PfandCents: row.PfandCents,
		}
	}
	return pfand
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

// GetEigeneUebersicht takes the user's Pfand row from GetPfandProServicekraft, so the overview and the settlement row
// attribute Pfand identically.
func (r Repository) GetEigeneUebersicht(ctx context.Context, userID int, kassensitzungNr int) (reporting.EigeneUebersicht, error) {
	var (
		row       dbgen.GetEigeneUebersichtRow
		pfandRows []dbgen.GetPfandProServicekraftRow
	)

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		var err error
		row, err = r.q.GetEigeneUebersicht(gctx, dbgen.GetEigeneUebersichtParams{
			UserID:          userID,
			KassensitzungNr: kassensitzungNr,
		})
		return err
	})
	g.Go(func() error {
		var err error
		pfandRows, err = r.q.GetPfandProServicekraft(gctx, kassensitzungNr)
		return err
	})
	if err := g.Wait(); err != nil {
		return reporting.EigeneUebersicht{}, db.Error(err)
	}

	pfandCents := 0
	for _, p := range pfandRows {
		if p.UserID == userID {
			pfandCents = p.PfandCents
			break
		}
	}

	return reporting.EigeneUebersicht{
		AnzahlBestellungen: row.AnzahlBestellungen,
		BestellungenCents:  row.BestellungenCents,
		AnzahlZahlungen:    row.AnzahlZahlungen,
		ZahlungenCents:     row.ZahlungenCents,
		AnzahlRuecknahmen:  row.AnzahlRuecknahmen,
		RuecknahmenCents:   row.RuecknahmenCents,
		PfandCents:         pfandCents,
		AbzugebenCents:     row.ZahlungenCents - row.RuecknahmenCents - pfandCents,
	}, nil
}
