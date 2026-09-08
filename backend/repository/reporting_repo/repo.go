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

// servicekraftRefJSON deserialisiert eine Servicekraft-Referenz aus der
// betroffene-Spalte der GetStornierungen-Query (jsonb_build_object).
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
		return reporting.ReportingData{}, err
	}

	stornierungen, err := toStornierungen(stornoRows)
	if err != nil {
		return reporting.ReportingData{}, err
	}

	metadaten, err := toMetadaten(metadatenRow)
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
		Metadaten:       metadaten,
		Summary:         toSummary(stats),
		Breakdowns: reporting.Breakdowns{
			AbrechnungProServicekraft: toAbrechnungServicekraft(kassiertRows),
		},
		UmsatzProSteuersatz: umsatzProSteuersatz,
		Stornierungen:       stornierungen,
	}, nil
}

// kassensturzDataJSON deserialisiert die für den Berichtskopf benötigte
// Kassensturz-Differenz aus dem kassensturz-durchgefuehrt:v1-Event.
type kassensturzDataJSON struct {
	DifferenzCents int `json:"differenzCents"`
}

// toMetadaten übersetzt die Sitzungs-Metadaten-Zeile in das Domänenmodell:
// nullable Zeitpunkte/Benutzer werden zu optionalen Feldern, die
// Kassensturz-Differenz wird aus dem JSONB-Event geparst (nil ohne Kassensturz).
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
	// Ohne Kassensturz liefert die Query das JSON-Literal 'null'; in ein
	// Pointer-Ziel deserialisiert das zu nil und lässt das Feld sauber leer.
	var data *kassensturzDataJSON
	if err := json.Unmarshal(row.KassensturzData, &data); err != nil {
		return reporting.Metadaten{}, err
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

// toAbrechnungServicekraft übersetzt die Kassiert-Zeilen der Query in die
// Abrechnung pro Servicekraft. Rücknahmen, Storno-Zähler und der
// Abzugeben-Saldo entstehen erst in der Anwendungsschicht aus den
// Storno-Detailzeilen (Storno-Zuordnung) und bleiben hier null.
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

// toBetroffene übersetzt die von der Query aufgelöste Storno-Zuordnung in
// Domänen-Referenzen. Die Query garantiert eine nicht-leere Liste (Rückfall auf
// den Akteur), sodass hier keine Ersatzlogik nötig ist.
func toBetroffene(raw json.RawMessage) ([]reporting.ServicekraftRef, error) {
	var refs []servicekraftRefJSON
	if err := json.Unmarshal(raw, &refs); err != nil {
		return nil, err
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
			return nil, err
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

// GetProduktStatistik liefert die flachen Verkaufszeilen je Variante einer
// Kassensitzung (ausgegebene Menge und Umsatz); die Gruppierung/Sortierung zu
// Kategorie-Abschnitten übernimmt die Anwendungsschicht. Bewusst als eigene
// Methode statt in der GetReporting-errgroup, damit derselbe Code den
// Abrechnungs- und den Live-Pfad speist.
func (r Repository) GetProduktStatistik(ctx context.Context, kassensitzungNr int) ([]reporting.ProduktStatistikZeile, error) {
	rows, err := r.q.GetProduktStatistik(ctx, kassensitzungNr)
	if err != nil {
		return nil, err
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
		return reporting.EigeneUebersicht{}, err
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
