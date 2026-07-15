package http

import (
	"context"
	"net/http"
	"time"

	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/middleware"
	"github.com/nicograef/jotti/backend/domain/reporting"
)

type query interface {
	GetReporting(ctx context.Context, kassensitzungNr int) (reporting.ReportingData, error)
	GetEigeneUebersicht(ctx context.Context, userID int) (reporting.EigeneUebersicht, error)
	GetAbgeschlosseneKassensitzungen(ctx context.Context) ([]reporting.AbgeschlosseneSitzung, error)
	GetLiveReporting(ctx context.Context) (*reporting.LiveReportingData, error)
}

type QueryHandler struct {
	Query query
}

type getReportingRequest struct {
	KassensitzungNr int `json:"kassensitzungNr"`
}

type metadatenResponse struct {
	EroeffnetAm               *time.Time `json:"eroeffnetAm"`
	AbgeschlossenAm           *time.Time `json:"abgeschlossenAm"`
	AbgeschlossenVon          string     `json:"abgeschlossenVon"`
	KassensturzDifferenzCents *int       `json:"kassensturzDifferenzCents"`
}

type reportingResponse struct {
	KassensitzungNr     int                        `json:"kassensitzungNr"`
	Metadaten           metadatenResponse          `json:"metadaten"`
	Summary             summaryResponse            `json:"summary"`
	Breakdowns          breakdownsResponse         `json:"breakdowns"`
	UmsatzProSteuersatz []umsatzSteuersatzResponse `json:"umsatzProSteuersatz"`
	Stornierungen       []stornierungDetail        `json:"stornierungen"`
	ProduktStatistik    []produktStatistikResponse `json:"produktStatistik"`
}

func (h *QueryHandler) GetReportingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := getReportingRequest{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		if body.KassensitzungNr < 1 {
			helper.SendClientError(w, "invalid_kassensitzung_nr", "kassensitzungNr must be at least 1")
			return
		}

		data, err := h.Query.GetReporting(r.Context(), body.KassensitzungNr)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, toReportingResponse(data))
	}
}

type summaryResponse struct {
	GesamtUmsatzCents        int `json:"gesamtUmsatzCents"`
	GesamtBestellungenCents  int `json:"gesamtBestellungenCents"`
	GesamtStornierungenCents int `json:"gesamtStornierungenCents"`
	GeldtransitCents         int `json:"geldtransitCents"`
	AnzahlBestellungen       int `json:"anzahlBestellungen"`
	AnzahlStornierungen      int `json:"anzahlStornierungen"`
	AnzahlDirektverkaeufe    int `json:"anzahlDirektverkaeufe"`
	DirektverkaufUmsatzCents int `json:"direktverkaufUmsatzCents"`
}

type breakdownsResponse struct {
	UmsatzProServicekraft        []umsatzServicekraft      `json:"umsatzProServicekraft"`
	StornierungenProServicekraft []stornierungServicekraft `json:"stornierungenProServicekraft"`
}

type umsatzServicekraft struct {
	UserID         int    `json:"userId"`
	UserName       string `json:"userName"`
	Name           string `json:"name"`
	ZahlungenCents int    `json:"zahlungenCents"`
}

// stornierungServicekraft ist das Storno-Aggregat pro Servicekraft (Anzahl und
// Betrag) — identisch in Live- und Reporting-Response.
type stornierungServicekraft struct {
	UserID              int    `json:"userId"`
	UserName            string `json:"userName"`
	Name                string `json:"name"`
	AnzahlStornierungen int    `json:"anzahlStornierungen"`
	StornierungenCents  int    `json:"stornierungenCents"`
}

type umsatzSteuersatzResponse struct {
	Satz        string `json:"satz"`
	BruttoCents int    `json:"bruttoCents"`
	NettoCents  int    `json:"nettoCents"`
	SteuerCents int    `json:"steuerCents"`
}

// varianteStatistikResponse und produktStatistikResponse tragen die gruppierte
// Verkaufsstatistik der Response; die flachen Repo-Zeilen erscheinen nie hier.
type varianteStatistikResponse struct {
	VarianteID       int    `json:"varianteId"`
	VarianteName     string `json:"varianteName"`
	AusgegebeneMenge int    `json:"ausgegebeneMenge"`
	UmsatzCents      int    `json:"umsatzCents"`
}

type produktStatistikResponse struct {
	Kategorie        string                      `json:"kategorie"`
	ProduktName      string                      `json:"produktName"`
	AusgegebeneMenge int                         `json:"ausgegebeneMenge"`
	UmsatzCents      int                         `json:"umsatzCents"`
	Varianten        []varianteStatistikResponse `json:"varianten"`
}

type stornierungPosition struct {
	ProduktName      string `json:"produktName"`
	VarianteName     string `json:"varianteName"`
	Menge            int    `json:"menge"`
	EinzelpreisCents int    `json:"einzelpreisCents"`
}

type stornierungDetail struct {
	Zeitpunkt    time.Time             `json:"zeitpunkt"`
	Quelle       string                `json:"quelle"`
	BarRueckgabe bool                  `json:"barRueckgabe"`
	TischID      int                   `json:"tischId"`
	TischName    string                `json:"tischName"`
	UserID       int                   `json:"userId"`
	UserName     string                `json:"userName"`
	Name         string                `json:"name"`
	BetragCents  int                   `json:"betragCents"`
	Kommentar    string                `json:"kommentar"`
	Positionen   []stornierungPosition `json:"positionen"`
}

func toUmsatzServicekraft(u reporting.UmsatzServicekraft) umsatzServicekraft {
	return umsatzServicekraft{
		UserID:         u.UserID,
		UserName:       u.UserName,
		Name:           u.Name,
		ZahlungenCents: u.ZahlungenCents,
	}
}

func toUmsatzServicekraftList(umsatz []reporting.UmsatzServicekraft) []umsatzServicekraft {
	out := make([]umsatzServicekraft, len(umsatz))
	for i := range umsatz {
		out[i] = toUmsatzServicekraft(umsatz[i])
	}
	return out
}

func toStornierungenProServicekraft(werte []reporting.StornierungServicekraft) []stornierungServicekraft {
	out := make([]stornierungServicekraft, len(werte))
	for i, w := range werte {
		out[i] = stornierungServicekraft{
			UserID:              w.UserID,
			UserName:            w.UserName,
			Name:                w.Name,
			AnzahlStornierungen: w.AnzahlStornierungen,
			StornierungenCents:  w.StornierungenCents,
		}
	}
	return out
}

func toUmsatzSteuersatz(u reporting.UmsatzSteuersatz) umsatzSteuersatzResponse {
	return umsatzSteuersatzResponse{
		Satz:        string(u.Satz),
		BruttoCents: u.BruttoCents,
		NettoCents:  u.NettoCents,
		SteuerCents: u.SteuerCents,
	}
}

func toUmsatzSteuersatzList(werte []reporting.UmsatzSteuersatz) []umsatzSteuersatzResponse {
	if len(werte) == 0 {
		return []umsatzSteuersatzResponse{}
	}

	out := make([]umsatzSteuersatzResponse, len(werte))
	for i := range werte {
		out[i] = toUmsatzSteuersatz(werte[i])
	}
	return out
}

func toStornierungPosition(p reporting.StornierungPosition) stornierungPosition {
	return stornierungPosition{
		ProduktName:      p.ProduktName,
		VarianteName:     p.VarianteName,
		Menge:            p.Menge,
		EinzelpreisCents: p.EinzelpreisCents,
	}
}

func toStornierungPositionen(positionen []reporting.StornierungPosition) []stornierungPosition {
	out := make([]stornierungPosition, len(positionen))
	for i := range positionen {
		out[i] = toStornierungPosition(positionen[i])
	}
	return out
}

func toStornierungDetail(d reporting.StornierungDetail) stornierungDetail {
	return stornierungDetail{
		Zeitpunkt:    d.Zeitpunkt,
		Quelle:       d.Quelle,
		BarRueckgabe: d.BarRueckgabe,
		TischID:      d.TischID,
		TischName:    d.TischName,
		UserID:       d.UserID,
		UserName:     d.UserName,
		Name:         d.Name,
		BetragCents:  d.BetragCents,
		Kommentar:    d.Kommentar,
		Positionen:   toStornierungPositionen(d.Positionen),
	}
}

func toStornierungDetails(details []reporting.StornierungDetail) []stornierungDetail {
	out := make([]stornierungDetail, len(details))
	for i := range details {
		out[i] = toStornierungDetail(details[i])
	}
	return out
}

func toProduktStatistik(p reporting.ProduktStatistik) produktStatistikResponse {
	varianten := make([]varianteStatistikResponse, len(p.Varianten))
	for i, v := range p.Varianten {
		varianten[i] = varianteStatistikResponse{
			VarianteID:       v.VarianteID,
			VarianteName:     v.VarianteName,
			AusgegebeneMenge: v.AusgegebeneMenge,
			UmsatzCents:      v.UmsatzCents,
		}
	}
	return produktStatistikResponse{
		Kategorie:        p.Kategorie,
		ProduktName:      p.ProduktName,
		AusgegebeneMenge: p.AusgegebeneMenge,
		UmsatzCents:      p.UmsatzCents,
		Varianten:        varianten,
	}
}

func toProduktStatistikList(werte []reporting.ProduktStatistik) []produktStatistikResponse {
	out := make([]produktStatistikResponse, len(werte))
	for i := range werte {
		out[i] = toProduktStatistik(werte[i])
	}
	return out
}

func toReportingResponse(d reporting.ReportingData) reportingResponse {
	return reportingResponse{
		KassensitzungNr: d.KassensitzungNr,
		Metadaten: metadatenResponse{
			EroeffnetAm:               d.Metadaten.EroeffnetAm,
			AbgeschlossenAm:           d.Metadaten.AbgeschlossenAm,
			AbgeschlossenVon:          d.Metadaten.AbgeschlossenVon,
			KassensturzDifferenzCents: d.Metadaten.KassensturzDifferenzCents,
		},
		Summary: summaryResponse{
			GesamtUmsatzCents:        d.Summary.GesamtUmsatzCents,
			GesamtBestellungenCents:  d.Summary.GesamtBestellungenCents,
			GesamtStornierungenCents: d.Summary.GesamtStornierungenCents,
			GeldtransitCents:         d.Summary.GeldtransitCents,
			AnzahlBestellungen:       d.Summary.AnzahlBestellungen,
			AnzahlStornierungen:      d.Summary.AnzahlStornierungen,
			AnzahlDirektverkaeufe:    d.Summary.AnzahlDirektverkaeufe,
			DirektverkaufUmsatzCents: d.Summary.DirektverkaufUmsatzCents,
		},
		Breakdowns: breakdownsResponse{
			UmsatzProServicekraft:        toUmsatzServicekraftList(d.Breakdowns.UmsatzProServicekraft),
			StornierungenProServicekraft: toStornierungenProServicekraft(d.Breakdowns.StornierungenProServicekraft),
		},
		UmsatzProSteuersatz: toUmsatzSteuersatzList(d.UmsatzProSteuersatz),
		Stornierungen:       toStornierungDetails(d.Stornierungen),
		ProduktStatistik:    toProduktStatistikList(d.ProduktStatistik),
	}
}

type kassensitzungItem struct {
	ZNr               int        `json:"zNr"`
	Datum             string     `json:"datum"`
	Bezeichnung       string     `json:"bezeichnung"`
	UmsatzGesamtCents int        `json:"umsatzGesamtCents"`
	AbgeschlossenAm   *time.Time `json:"abgeschlossenAm"`
}

type getAbgeschlosseneKassensitzungenResponse struct {
	Kassensitzungen []kassensitzungItem `json:"kassensitzungen"`
}

func (h *QueryHandler) GetAbgeschlosseneKassensitzungenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := h.Query.GetAbgeschlosseneKassensitzungen(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}

		items := make([]kassensitzungItem, len(data))
		for i, k := range data {
			items[i] = kassensitzungItem{
				ZNr:               k.ZNr,
				Datum:             k.Datum.Format("2006-01-02"),
				Bezeichnung:       k.Bezeichnung,
				UmsatzGesamtCents: k.UmsatzGesamtCents,
				AbgeschlossenAm:   k.AbgeschlossenAm,
			}
		}

		helper.SendResponse(w, getAbgeschlosseneKassensitzungenResponse{Kassensitzungen: items})
	}
}

type eigeneUebersichtResponse struct {
	AnzahlBestellungen int `json:"anzahlBestellungen"`
	BestellungenCents  int `json:"bestellungenCents"`
	AnzahlZahlungen    int `json:"anzahlZahlungen"`
	ZahlungenCents     int `json:"zahlungenCents"`
}

func (h *QueryHandler) GetEigeneUebersichtHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _, ok := middleware.UserFromContext(r.Context())
		if !ok {
			helper.SendServerError(w)
			return
		}

		data, err := h.Query.GetEigeneUebersicht(r.Context(), userID)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, eigeneUebersichtResponse{
			AnzahlBestellungen: data.AnzahlBestellungen,
			BestellungenCents:  data.BestellungenCents,
			AnzahlZahlungen:    data.AnzahlZahlungen,
			ZahlungenCents:     data.ZahlungenCents,
		})
	}
}

type offenerTischResponse struct {
	TischID    int    `json:"tischId"`
	TischName  string `json:"tischName"`
	SaldoCents int    `json:"saldoCents"`
}

type liveSummaryResponse struct {
	GesamtUmsatzCents        int `json:"gesamtUmsatzCents"`
	GesamtBestellungenCents  int `json:"gesamtBestellungenCents"`
	GesamtStornierungenCents int `json:"gesamtStornierungenCents"`
	GeldtransitCents         int `json:"geldtransitCents"`
	AnzahlBestellungen       int `json:"anzahlBestellungen"`
	AnzahlStornierungen      int `json:"anzahlStornierungen"`
	AnzahlDirektverkaeufe    int `json:"anzahlDirektverkaeufe"`
	DirektverkaufUmsatzCents int `json:"direktverkaufUmsatzCents"`
}

// offeneArbeitTischLiveResponse ist die schlanke Tisch-Zeile der Live-Sicht: nur
// der Tisch-Name für die Inline-Anzeige. Der offene Betrag wird auf
// Servicekraft-Ebene (servicekraftLiveResponse.OffenCents) aggregiert.
type offeneArbeitTischLiveResponse struct {
	TischID   int    `json:"tischId"`
	TischName string `json:"tischName"`
}

// servicekraftLiveResponse ist die Live-Sicht pro Servicekraft: kassierter
// Umsatz zusammengeführt mit der offenen eigenen Arbeit.
type servicekraftLiveResponse struct {
	UserID         int                             `json:"userId"`
	UserName       string                          `json:"userName"`
	Name           string                          `json:"name"`
	ZahlungenCents int                             `json:"zahlungenCents"`
	OffenCents     int                             `json:"offenCents"`
	OffeneTische   []offeneArbeitTischLiveResponse `json:"offeneTische"`
	Erledigt       bool                            `json:"erledigt"`
}

// liveBreakdownsResponse trägt im Live-Dashboard die zusammengeführte
// Servicekraft-Sicht statt des reinen kassierten Umsatzes; das Storno-Aggregat
// pro Servicekraft ist identisch zur Reporting-Response.
type liveBreakdownsResponse struct {
	Servicekraefte               []servicekraftLiveResponse `json:"servicekraefte"`
	StornierungenProServicekraft []stornierungServicekraft  `json:"stornierungenProServicekraft"`
}

type liveReportingResponse struct {
	KassensitzungNr  int                        `json:"kassensitzungNr"`
	Bezeichnung      string                     `json:"bezeichnung"`
	Datum            string                     `json:"datum"` // Kalendertag YYYY-MM-DD
	OffeneTische     []offenerTischResponse     `json:"offeneTische"`
	OffeneSaldiCents int                        `json:"offeneSaldiCents"`
	Summary          liveSummaryResponse        `json:"summary"`
	Breakdowns       liveBreakdownsResponse     `json:"breakdowns"`
	Stornierungen    []stornierungDetail        `json:"stornierungen"`
	ProduktStatistik []produktStatistikResponse `json:"produktStatistik"`
}

func toServicekraftLive(s reporting.ServicekraftLive) servicekraftLiveResponse {
	offeneTische := make([]offeneArbeitTischLiveResponse, len(s.OffeneTische))
	for i, t := range s.OffeneTische {
		offeneTische[i] = offeneArbeitTischLiveResponse{
			TischID:   t.TischID,
			TischName: t.TischName,
		}
	}
	return servicekraftLiveResponse{
		UserID:         s.UserID,
		UserName:       s.UserName,
		Name:           s.Name,
		ZahlungenCents: s.ZahlungenCents,
		OffenCents:     s.OffenCents,
		OffeneTische:   offeneTische,
		Erledigt:       s.Erledigt,
	}
}

func toServicekraefteLive(servicekraefte []reporting.ServicekraftLive) []servicekraftLiveResponse {
	out := make([]servicekraftLiveResponse, len(servicekraefte))
	for i := range servicekraefte {
		out[i] = toServicekraftLive(servicekraefte[i])
	}
	return out
}

func toLiveReportingResponse(d reporting.LiveReportingData) liveReportingResponse {
	offeneTische := make([]offenerTischResponse, len(d.OffeneTische))
	for i, t := range d.OffeneTische {
		offeneTische[i] = offenerTischResponse{
			TischID:    t.TischID,
			TischName:  t.TischName,
			SaldoCents: t.SaldoCents,
		}
	}
	return liveReportingResponse{
		KassensitzungNr:  d.KassensitzungNr,
		Bezeichnung:      d.Bezeichnung,
		Datum:            d.Datum.Format("2006-01-02"),
		OffeneTische:     offeneTische,
		OffeneSaldiCents: d.OffeneSaldiCents,
		Summary: liveSummaryResponse{
			GesamtUmsatzCents:        d.Summary.GesamtUmsatzCents,
			GesamtBestellungenCents:  d.Summary.GesamtBestellungenCents,
			GesamtStornierungenCents: d.Summary.GesamtStornierungenCents,
			GeldtransitCents:         d.Summary.GeldtransitCents,
			AnzahlBestellungen:       d.Summary.AnzahlBestellungen,
			AnzahlStornierungen:      d.Summary.AnzahlStornierungen,
			AnzahlDirektverkaeufe:    d.Summary.AnzahlDirektverkaeufe,
			DirektverkaufUmsatzCents: d.Summary.DirektverkaufUmsatzCents,
		},
		Breakdowns: liveBreakdownsResponse{
			Servicekraefte:               toServicekraefteLive(d.Servicekraefte),
			StornierungenProServicekraft: toStornierungenProServicekraft(d.Breakdowns.StornierungenProServicekraft),
		},
		Stornierungen:    toStornierungDetails(d.Stornierungen),
		ProduktStatistik: toProduktStatistikList(d.ProduktStatistik),
	}
}

func (h *QueryHandler) GetLiveReportingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := h.Query.GetLiveReporting(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}
		if data == nil {
			helper.SendResponse(w, nil)
			return
		}
		helper.SendResponse(w, toLiveReportingResponse(*data))
	}
}
