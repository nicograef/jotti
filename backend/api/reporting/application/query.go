package application

import (
	"context"
	"sort"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/kasse"
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
}

type kassensitzungenRepo interface {
	GetAbgeschlosseneKassensitzungen(ctx context.Context) ([]kasse.Kassensitzung, error)
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

func (q Query) GetAbgeschlosseneKassensitzungen(ctx context.Context) ([]kasse.Kassensitzung, error) {
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

	sessions, err := q.TischSessionRepo.GetTischSessionsByKassensitzungNr(ctx, kassensitzungNr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get tisch sessions for eigene uebersicht")
		return reporting.EigeneUebersicht{}, ErrDatabase
	}

	tische, err := q.TischRepo.GetAllTables(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get tische for eigene uebersicht")
		return reporting.EigeneUebersicht{}, ErrDatabase
	}
	nameByTischID := make(map[int]string, len(tische))
	for _, t := range tische {
		nameByTischID[t.ID] = t.Name
	}

	rollup := kasse.ComputeOffeneArbeitRollup(sessions, userID)
	data.AlleErledigt = rollup.Erledigt
	data.OffeneTische = make([]reporting.OffeneArbeitTisch, len(rollup.OffeneTische))
	for i, tisch := range rollup.OffeneTische {
		data.OffeneTische[i] = reporting.OffeneArbeitTisch{
			TischID:         tisch.TischID,
			TischName:       nameByTischID[tisch.TischID],
			AnzahlUnbezahlt: tisch.AnzahlUnbezahlt,
			AnzahlOffen:     tisch.AnzahlOffen,
			OffenCents:      tisch.OffenCents,
		}
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
		for i, tisch := range arbeit.OffeneTische {
			offeneTische[i] = reporting.OffeneArbeitTisch{
				TischID:         tisch.TischID,
				TischName:       nameByTischID[tisch.TischID],
				AnzahlUnbezahlt: tisch.AnzahlUnbezahlt,
				AnzahlOffen:     tisch.AnzahlOffen,
				OffenCents:      tisch.OffenCents,
			}
		}

		if idx, ok := indexByUserID[arbeit.UserID]; ok {
			servicekraefte[idx].OffeneTische = offeneTische
			servicekraefte[idx].Erledigt = false
			continue
		}

		servicekraefte = append(servicekraefte, reporting.ServicekraftLive{
			UserID:       arbeit.UserID,
			UserName:     arbeit.UserName,
			OffeneTische: offeneTische,
			Erledigt:     false,
		})
	}

	return servicekraefte
}
