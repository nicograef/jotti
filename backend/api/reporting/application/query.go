package application

import (
	"context"
	"sort"

	"golang.org/x/sync/errgroup"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/domain/reporting"
	"github.com/nicograef/jotti/backend/domain/steuer"
	"github.com/nicograef/jotti/backend/internal/logging"
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
	GetAktiveKassensitzung(ctx context.Context) (*kasse.Kassensitzung, error)
}

type tischSessionRepo interface {
	GetTischSessionsByKassensitzungNr(ctx context.Context, kassensitzungNr int) ([]kasse.TischSession, error)
}

type tischRepo interface {
	GetAlleTischNamen(ctx context.Context) (map[int]string, error)
}

type Query struct {
	ReportingRepo       reportingRepo
	KassensitzungenRepo kassensitzungenRepo
	TischSessionRepo    tischSessionRepo
	TischRepo           tischRepo
}

func (q Query) GetReporting(ctx context.Context, kassensitzungNr int) (reporting.ReportingData, error) {
	log := logging.FromContext(ctx)

	var (
		data   reporting.ReportingData
		zeilen []reporting.ProduktStatistikZeile
	)
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		var err error
		if data, err = q.ReportingRepo.GetReporting(gctx, kassensitzungNr); err != nil {
			log.Error("Failed to get reporting", "error", err)
			return ErrDatabase
		}
		return nil
	})
	g.Go(func() error {
		var err error
		if zeilen, err = q.ReportingRepo.GetProduktStatistik(gctx, kassensitzungNr); err != nil {
			log.Error("Failed to get produkt statistik", "error", err)
			return ErrDatabase
		}
		return nil
	})
	if err := g.Wait(); err != nil {
		return reporting.ReportingData{}, err
	}

	data.UmsatzProSteuersatz = computeUmsatzProSteuersatz(data.UmsatzProSteuersatz)
	data.Breakdowns.AbrechnungProServicekraft = aggregateAbrechnungProServicekraft(
		data.Breakdowns.AbrechnungProServicekraft, data.Stornierungen, data.Breakdowns.PfandProServicekraft,
	)
	data.ProduktStatistik = gruppiereProduktStatistik(zeilen)

	log.Info("Retrieved reporting")
	return data, nil
}

// aggregateAbrechnungProServicekraft merges table payments (by actor), storno rows (by storno attribution) and
// table Pfand returns (by the return's actor). The rules are in docs/handbuch.md §7.2.
func aggregateAbrechnungProServicekraft(
	kassiert []reporting.AbrechnungServicekraft,
	stornierungen []reporting.StornierungDetail,
	pfand []reporting.PfandServicekraft,
) []reporting.AbrechnungServicekraft {
	out := make([]reporting.AbrechnungServicekraft, len(kassiert))
	copy(out, kassiert)
	indexByUserID := make(map[int]int, len(kassiert))
	for i, k := range out {
		indexByUserID[k.UserID] = i
	}
	zeileFuer := func(ref reporting.ServicekraftRef) int {
		idx, ok := indexByUserID[ref.UserID]
		if !ok {
			idx = len(out)
			indexByUserID[ref.UserID] = idx
			out = append(out, reporting.AbrechnungServicekraft{
				UserID:   ref.UserID,
				UserName: ref.UserName,
				Name:     ref.Name,
			})
		}
		return idx
	}

	for _, s := range stornierungen {
		if s.Quelle == reporting.QuelleDirektverkauf {
			continue
		}
		for _, betroffen := range s.Betroffene {
			idx := zeileFuer(betroffen)
			out[idx].AnzahlStornierungen++
			// Only the cash-relevant return carries an amount; its zahlungId attributes it to one cashier.
			if s.BarRueckgabe {
				out[idx].RuecknahmenCents += s.BetragCents
			}
		}
	}

	for _, p := range pfand {
		idx := zeileFuer(reporting.ServicekraftRef{UserID: p.UserID, UserName: p.UserName, Name: p.Name})
		out[idx].PfandCents += p.PfandCents
	}

	for i := range out {
		out[i].AbzugebenCents = out[i].KassiertCents - out[i].RuecknahmenCents - out[i].PfandCents
	}

	sort.SliceStable(out, func(a, b int) bool {
		return out[a].AbzugebenCents > out[b].AbzugebenCents
	})

	return out
}

// computeUmsatzProSteuersatz splits each gross row with steuer.Aufteilen before aggregating per rate, the same basis as
// Beleg, TSE processData and DSFinV-K export. steuer.Aufteilen ignores negatives, so returns are split by magnitude.
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
		}
	}

	return out
}

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

// gruppiereProduktStatistik returns products with subtotals, fully sorted; the frontend only collapses
// single-variant products.
func gruppiereProduktStatistik(zeilen []reporting.ProduktStatistikZeile) []reporting.ProduktStatistik {
	// Within a session a product has exactly one category and a unique name.
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
	log := logging.FromContext(ctx)

	data, err := q.KassensitzungenRepo.GetAbgeschlosseneKassensitzungen(ctx)
	if err != nil {
		log.Error("Failed to get abgeschlossene kassensitzungen", "error", err)
		return nil, ErrDatabase
	}

	log.Info("Retrieved abgeschlossene kassensitzungen")
	return data, nil
}

func (q Query) GetEigeneUebersicht(ctx context.Context, userID int) (reporting.EigeneUebersicht, error) {
	log := logging.FromContext(ctx)

	kassensitzungNr, err := q.KassensitzungenRepo.GetOffeneKassensitzungNr(ctx)
	if err != nil {
		log.Error("Failed to get offene kassensitzung nr", "error", err)
		return reporting.EigeneUebersicht{}, ErrDatabase
	}

	data, err := q.ReportingRepo.GetEigeneUebersicht(ctx, userID, kassensitzungNr)
	if err != nil {
		log.Error("Failed to get eigene uebersicht", "error", err)
		return reporting.EigeneUebersicht{}, ErrDatabase
	}

	log.Info("Retrieved eigene uebersicht")
	return data, nil
}

func (q Query) GetLiveReporting(ctx context.Context) (*reporting.LiveReportingData, error) {
	log := logging.FromContext(ctx)

	ks, err := q.KassensitzungenRepo.GetAktiveKassensitzung(ctx)
	if err != nil {
		log.Error("Failed to get aktive kassensitzung", "error", err)
		return nil, ErrDatabase
	}
	if ks == nil {
		return nil, nil
	}

	var (
		data          reporting.LiveReportingData
		sessions      []kasse.TischSession
		nameByTischID map[int]string
		zeilen        []reporting.ProduktStatistikZeile
	)
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		var err error
		if data, err = q.ReportingRepo.GetLiveReporting(gctx, ks.ZNr); err != nil {
			log.Error("Failed to get live reporting", "error", err)
			return ErrDatabase
		}
		return nil
	})
	g.Go(func() error {
		var err error
		if sessions, err = q.TischSessionRepo.GetTischSessionsByKassensitzungNr(gctx, ks.ZNr); err != nil {
			log.Error("Failed to get tisch sessions for live reporting", "error", err)
			return ErrDatabase
		}
		return nil
	})
	g.Go(func() error {
		var err error
		if nameByTischID, err = q.TischRepo.GetAlleTischNamen(gctx); err != nil {
			log.Error("Failed to get tische for live reporting", "error", err)
			return ErrDatabase
		}
		return nil
	})
	g.Go(func() error {
		var err error
		if zeilen, err = q.ReportingRepo.GetProduktStatistik(gctx, ks.ZNr); err != nil {
			log.Error("Failed to get produkt statistik for live reporting", "error", err)
			return ErrDatabase
		}
		return nil
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}

	data.Bezeichnung = ks.Bezeichnung
	data.Datum = ks.Datum

	data.Breakdowns.AbrechnungProServicekraft = aggregateAbrechnungProServicekraft(
		data.Breakdowns.AbrechnungProServicekraft, data.Stornierungen, data.Breakdowns.PfandProServicekraft,
	)
	data.Servicekraefte = mergeServicekraefteLive(data.Breakdowns.AbrechnungProServicekraft, sessions, nameByTischID)
	data.ProduktStatistik = gruppiereProduktStatistik(zeilen)

	log.Info("Retrieved live reporting")
	return &data, nil
}

// mergeServicekraefteLive keeps the settlement order first, then appends staff with only open work by ascending UserID.
func mergeServicekraefteLive(
	abrechnung []reporting.AbrechnungServicekraft,
	sessions []kasse.TischSession,
	nameByTischID map[int]string,
) []reporting.ServicekraftLive {
	servicekraefte := make([]reporting.ServicekraftLive, len(abrechnung))
	indexByUserID := make(map[int]int, len(abrechnung))
	for i, a := range abrechnung {
		servicekraefte[i] = reporting.ServicekraftLive{
			UserID:              a.UserID,
			UserName:            a.UserName,
			Name:                a.Name,
			KassiertCents:       a.KassiertCents,
			RuecknahmenCents:    a.RuecknahmenCents,
			AnzahlStornierungen: a.AnzahlStornierungen,
			PfandCents:          a.PfandCents,
			AbzugebenCents:      a.AbzugebenCents,
			OffeneTische:        []reporting.OffeneArbeitTisch{},
			Erledigt:            true,
		}
		indexByUserID[a.UserID] = i
	}

	for _, arbeit := range kasse.ComputeOffeneArbeitProServicekraft(sessions) {
		offeneTische := make([]reporting.OffeneArbeitTisch, len(arbeit.OffeneTische))
		offenCents := 0
		for i, tisch := range arbeit.OffeneTische {
			offeneTische[i] = reporting.OffeneArbeitTisch{
				TischID:   tisch.TischID,
				TischName: nameByTischID[tisch.TischID],
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
