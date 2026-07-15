package application

import (
	"context"
	"sort"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/domain/reporting"
	"github.com/nicograef/jotti/backend/domain/steuer"
	"github.com/nicograef/jotti/backend/domain/tisch"
	"github.com/rs/zerolog"
)

var ErrDatabase = db.ErrDatabase

type reportingRepo interface {
	GetReporting(ctx context.Context, kassensitzungNr int) (reporting.ReportingData, error)
	GetEigeneUebersicht(ctx context.Context, userID int, kassensitzungNr int) (reporting.EigeneUebersicht, error)
	GetLiveReporting(ctx context.Context, kassensitzungNr int) (reporting.LiveReportingData, error)
	GetProduktStatistik(ctx context.Context, kassensitzungNr int) ([]reporting.ProduktStatistikZeile, error)
}

type kassensitzungenRepo interface {
	GetAbgeschlosseneKassensitzungen(ctx context.Context) ([]reporting.AbgeschlosseneSitzung, error)
	GetOffeneKassensitzungNr(ctx context.Context) (int, error)
	GetOffeneKassensitzung(ctx context.Context) (*kasse.Kassensitzung, error)
}

type tischSessionRepo interface {
	GetTischSessionsByKassensitzungNr(ctx context.Context, kassensitzungNr int) ([]kasse.TischSession, error)
}

type tischRepo interface {
	GetAllTables(ctx context.Context) ([]tisch.Tisch, error)
}

type Query struct {
	ReportingRepo       reportingRepo
	KassensitzungenRepo kassensitzungenRepo
	TischSessionRepo    tischSessionRepo
	TischRepo           tischRepo
}

func (q Query) GetReporting(ctx context.Context, kassensitzungNr int) (reporting.ReportingData, error) {
	log := zerolog.Ctx(ctx)

	data, err := q.ReportingRepo.GetReporting(ctx, kassensitzungNr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get reporting")
		return reporting.ReportingData{}, ErrDatabase
	}

	data.UmsatzProSteuersatz = computeUmsatzProSteuersatz(data.UmsatzProSteuersatz)
	data.Breakdowns.StornierungenProServicekraft = aggregateStornierungenProServicekraft(data.Stornierungen)

	zeilen, err := q.ReportingRepo.GetProduktStatistik(ctx, kassensitzungNr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get produkt statistik")
		return reporting.ReportingData{}, ErrDatabase
	}
	data.ProduktStatistik = gruppiereProduktStatistik(zeilen)

	log.Info().Msg("Retrieved reporting")
	return data, nil
}

// aggregateStornierungenProServicekraft fasst die Storno-Detailzeilen pro
// Servicekraft zusammen (Anzahl und Betrag) — als Kontroll-Signal fürs
// Admin-Dashboard. Die Reihenfolge folgt dem ersten Auftreten in der
// Detail-Liste (stabil). Die Summen entsprechen den Storno-Kennzahlen der
// Summary, weil beide dieselben Storno-Events auswerten.
func aggregateStornierungenProServicekraft(stornierungen []reporting.StornierungDetail) []reporting.StornierungServicekraft {
	out := []reporting.StornierungServicekraft{}
	indexByUserID := make(map[int]int, len(stornierungen))
	for _, s := range stornierungen {
		if idx, ok := indexByUserID[s.UserID]; ok {
			out[idx].AnzahlStornierungen++
			out[idx].StornierungenCents += s.BetragCents
			continue
		}
		indexByUserID[s.UserID] = len(out)
		out = append(out, reporting.StornierungServicekraft{
			UserID:              s.UserID,
			UserName:            s.UserName,
			Name:                s.Name,
			AnzahlStornierungen: 1,
			StornierungenCents:  s.BetragCents,
		})
	}
	return out
}

// computeUmsatzProSteuersatz aggregiert die USt-Aufschlüsselung aus den
// unaggregierten Brutto-Positionszeilen — auf derselben Basis wie Beleg,
// TSE-processData und DSFinV-K-Export: steuer.Aufteilen je Positionszeile,
// danach Aggregation je Steuersatz. Warenrücknahmen kommen als negative
// Zeilen; wie beim Stornobeleg wird die positive Magnitude aufgeteilt und
// das Vorzeichen danach angewendet (steuer.Aufteilen ignoriert Negatives).
func computeUmsatzProSteuersatz(bruttoZeilen []reporting.UmsatzSteuersatz) []reporting.UmsatzSteuersatz {
	aggregiert := make(map[steuer.Steuersatz]reporting.UmsatzSteuersatz, 3)
	for _, zeile := range bruttoZeilen {
		vorzeichen := 1
		brutto := zeile.BruttoCents
		if brutto < 0 {
			vorzeichen = -1
			brutto = -brutto
		}
		for _, aufteilung := range steuer.Aufteilen(brutto, zeile.Satz) {
			current := aggregiert[aufteilung.Satz]
			current.Satz = aufteilung.Satz
			current.BruttoCents += vorzeichen * aufteilung.Brutto
			current.NettoCents += vorzeichen * aufteilung.Netto
			current.SteuerCents += vorzeichen * aufteilung.Steuer
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

// kategorieRang legt die feste Reihenfolge der Kategorie-Abschnitte fest:
// Essen → Getränke → Sonstiges. Unbekannte Kategorien (theoretischer Randfall)
// sortieren dahinter und werden untereinander alphabetisch geordnet.
func kategorieRang(kategorie string) int {
	switch kategorie {
	case string(produkt.EssenKategorie):
		return 0
	case string(produkt.GetraenkKategorie):
		return 1
	case string(produkt.SonstigesKategorie):
		return 2
	default:
		return 3
	}
}

// gruppiereProduktStatistik baut aus den flachen Varianten-Zeilen die
// Produkt-Hierarchie für den Report: je Produkt eine Gruppe mit Zwischensumme
// über ihre Varianten, in Kategorie-Abschnitte gegliedert. Sortierung:
// Kategorien fest (Essen → Getränke → Sonstiges), Produkte je Kategorie und
// Varianten je Produkt nach ausgegebener Menge absteigend, Name als stabiler
// Tiebreaker. Reine Funktion — das Backend liefert die Liste fertig sortiert,
// die Ein-Varianten-Zusammenfassung bleibt reine Präsentation im Frontend.
func gruppiereProduktStatistik(zeilen []reporting.ProduktStatistikZeile) []reporting.ProduktStatistik {
	// Produkte per (Kategorie, ProduktName) sammeln; ein Produkt liegt je Sitzung
	// in genau einer Kategorie, der Produktname ist innerhalb der Sitzung eindeutig.
	type produktKey struct {
		kategorie   string
		produktName string
	}
	indexByKey := make(map[produktKey]int, len(zeilen))
	produkte := []reporting.ProduktStatistik{}
	for _, z := range zeilen {
		key := produktKey{kategorie: z.Kategorie, produktName: z.ProduktName}
		idx, ok := indexByKey[key]
		if !ok {
			idx = len(produkte)
			indexByKey[key] = idx
			produkte = append(produkte, reporting.ProduktStatistik{
				Kategorie:   z.Kategorie,
				ProduktName: z.ProduktName,
				Varianten:   []reporting.VarianteStatistik{},
			})
		}
		produkte[idx].Varianten = append(produkte[idx].Varianten, reporting.VarianteStatistik{
			VarianteID:       z.VarianteID,
			VarianteName:     z.VarianteName,
			AusgegebeneMenge: z.AusgegebeneMenge,
			UmsatzCents:      z.UmsatzCents,
		})
		produkte[idx].AusgegebeneMenge += z.AusgegebeneMenge
		produkte[idx].UmsatzCents += z.UmsatzCents
	}

	for i := range produkte {
		varianten := produkte[i].Varianten
		sort.SliceStable(varianten, func(a, b int) bool {
			if varianten[a].AusgegebeneMenge != varianten[b].AusgegebeneMenge {
				return varianten[a].AusgegebeneMenge > varianten[b].AusgegebeneMenge
			}
			return varianten[a].VarianteName < varianten[b].VarianteName
		})
	}

	sort.SliceStable(produkte, func(a, b int) bool {
		rangA, rangB := kategorieRang(produkte[a].Kategorie), kategorieRang(produkte[b].Kategorie)
		if rangA != rangB {
			return rangA < rangB
		}
		if produkte[a].Kategorie != produkte[b].Kategorie {
			return produkte[a].Kategorie < produkte[b].Kategorie
		}
		if produkte[a].AusgegebeneMenge != produkte[b].AusgegebeneMenge {
			return produkte[a].AusgegebeneMenge > produkte[b].AusgegebeneMenge
		}
		return produkte[a].ProduktName < produkte[b].ProduktName
	})

	return produkte
}

func (q Query) GetAbgeschlosseneKassensitzungen(ctx context.Context) ([]reporting.AbgeschlosseneSitzung, error) {
	log := zerolog.Ctx(ctx)

	data, err := q.KassensitzungenRepo.GetAbgeschlosseneKassensitzungen(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get abgeschlossene kassensitzungen")
		return nil, ErrDatabase
	}

	log.Info().Msg("Retrieved abgeschlossene kassensitzungen")
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

	sessions, err := q.TischSessionRepo.GetTischSessionsByKassensitzungNr(ctx, ks.ZNr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get tisch sessions for live reporting")
		return nil, ErrDatabase
	}

	tische, err := q.TischRepo.GetAllTables(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get tische for live reporting")
		return nil, ErrDatabase
	}
	nameByTischID := make(map[int]string, len(tische))
	for _, t := range tische {
		nameByTischID[t.ID] = t.Name
	}

	data.Servicekraefte = mergeServicekraefteLive(data.Breakdowns.UmsatzProServicekraft, sessions, nameByTischID)
	data.Breakdowns.StornierungenProServicekraft = aggregateStornierungenProServicekraft(data.Stornierungen)

	zeilen, err := q.ReportingRepo.GetProduktStatistik(ctx, ks.ZNr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get produkt statistik for live reporting")
		return nil, ErrDatabase
	}
	data.ProduktStatistik = gruppiereProduktStatistik(zeilen)

	log.Info().Msg("Retrieved live reporting")
	return &data, nil
}

// mergeServicekraefteLive führt den kassierten Umsatz pro Servicekraft mit der
// offenen eigenen Arbeit aus den Tisch-Sessions per user_id zusammen.
// Servicekräfte mit kassiertem Umsatz erscheinen zuerst (in Umsatz-Reihenfolge),
// danach Personen mit ausschließlich offener Arbeit (aufsteigend nach UserID).
func mergeServicekraefteLive(
	umsatz []reporting.UmsatzServicekraft,
	sessions []kasse.TischSession,
	nameByTischID map[int]string,
) []reporting.ServicekraftLive {
	servicekraefte := make([]reporting.ServicekraftLive, len(umsatz))
	indexByUserID := make(map[int]int, len(umsatz))
	for i, u := range umsatz {
		servicekraefte[i] = reporting.ServicekraftLive{
			UserID:          u.UserID,
			UserName:        u.UserName,
			Name:            u.Name,
			ZahlungenCents:  u.ZahlungenCents,
			AnzahlZahlungen: u.AnzahlZahlungen,
			OffeneTische:    []reporting.OffeneArbeitTisch{},
			Erledigt:        true,
		}
		indexByUserID[u.UserID] = i
	}

	for _, arbeit := range kasse.ComputeOffeneArbeitProServicekraft(sessions) {
		offeneTische := make([]reporting.OffeneArbeitTisch, len(arbeit.OffeneTische))
		offenCents := 0
		for i, tisch := range arbeit.OffeneTische {
			offeneTische[i] = reporting.OffeneArbeitTisch{
				TischID:         tisch.TischID,
				TischName:       nameByTischID[tisch.TischID],
				AnzahlUnbezahlt: tisch.AnzahlUnbezahlt,
				AnzahlOffen:     tisch.AnzahlOffen,
				OffenCents:      tisch.OffenCents,
			}
			offenCents += tisch.OffenCents
		}

		if idx, ok := indexByUserID[arbeit.UserID]; ok {
			servicekraefte[idx].OffeneTische = offeneTische
			servicekraefte[idx].OffenCents = offenCents
			servicekraefte[idx].Erledigt = false
			continue
		}

		servicekraefte = append(servicekraefte, reporting.ServicekraftLive{
			UserID:       arbeit.UserID,
			UserName:     arbeit.UserName,
			OffenCents:   offenCents,
			OffeneTische: offeneTische,
			Erledigt:     false,
		})
	}

	return servicekraefte
}
