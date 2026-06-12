package http

import (
	"context"
	"net/http"
	"time"

	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/middleware"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/reporting"
)

type query interface {
	GetReporting(ctx context.Context, kassensitzungNr int) (reporting.ReportingData, error)
	GetEigeneUebersicht(ctx context.Context, userID int) (reporting.EigeneUebersicht, error)
	GetAllKassensitzungen(ctx context.Context) ([]kasse.Kassensitzung, error)
	GetLiveReporting(ctx context.Context) (*reporting.LiveReportingData, error)
}

type QueryHandler struct {
	Query query
}

type getReportingRequest struct {
	KassensitzungNr int `json:"kassensitzungNr"`
}

type reportingResponse struct {
	KassensitzungNr     int                        `json:"kassensitzungNr"`
	Summary             summaryResponse            `json:"summary"`
	Breakdowns          breakdownsResponse         `json:"breakdowns"`
	UmsatzProSteuersatz []umsatzSteuersatzResponse `json:"umsatzProSteuersatz"`
	Stornierungen       []stornierungDetail        `json:"stornierungen"`
}

func (h QueryHandler) GetReportingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := getReportingRequest{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		if body.KassensitzungNr < 1 {
			helper.SendClientError(w, "invalid_kassensitzung_nr", "kassensitzung_nr_ist_erforderlich")
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
	GesamtAuszahlungenCents  int `json:"gesamtAuszahlungenCents"`
	GesamtBestellungenCents  int `json:"gesamtBestellungenCents"`
	GesamtStornierungenCents int `json:"gesamtStornierungenCents"`
	AnzahlBestellungen       int `json:"anzahlBestellungen"`
	AnzahlStornierungen      int `json:"anzahlStornierungen"`
	AnzahlDirektverkaeufe    int `json:"anzahlDirektverkaeufe"`
	DirektverkaufUmsatzCents int `json:"direktverkaufUmsatzCents"`
}

type breakdownsResponse struct {
	UmsatzProServicekraft []umsatzServicekraft `json:"umsatzProServicekraft"`
	UmsatzProTisch        []umsatzTisch        `json:"umsatzProTisch"`
}

type umsatzServicekraft struct {
	UserID            int    `json:"userId"`
	UserName          string `json:"userName"`
	ZahlungenCents    int    `json:"zahlungenCents"`
	AuszahlungenCents int    `json:"auszahlungenCents"`
	AnzahlZahlungen   int    `json:"anzahlZahlungen"`
}

type umsatzTisch struct {
	TischID           int    `json:"tischId"`
	TischName         string `json:"tischName"`
	ZahlungenCents    int    `json:"zahlungenCents"`
	AuszahlungenCents int    `json:"auszahlungenCents"`
	AnzahlZahlungen   int    `json:"anzahlZahlungen"`
}

type umsatzSteuersatzResponse struct {
	Satz        string `json:"satz"`
	BruttoCents int    `json:"bruttoCents"`
	NettoCents  int    `json:"nettoCents"`
	SteuerCents int    `json:"steuerCents"`
}

type stornierungPosition struct {
	ProduktName  string `json:"produktName"`
	VarianteName string `json:"varianteName"`
	Menge        int    `json:"menge"`
	Einzelpreis  int    `json:"einzelpreis"`
}

type stornierungDetail struct {
	Zeitpunkt   time.Time             `json:"zeitpunkt"`
	TischID     int                   `json:"tischId"`
	TischName   string                `json:"tischName"`
	UserID      int                   `json:"userId"`
	UserName    string                `json:"userName"`
	BetragCents int                   `json:"betragCents"`
	Kommentar   string                `json:"kommentar"`
	Positionen  []stornierungPosition `json:"positionen"`
}

func toUmsatzServicekraft(u reporting.UmsatzServicekraft) umsatzServicekraft {
	return umsatzServicekraft{
		UserID:            u.UserID,
		UserName:          u.UserName,
		ZahlungenCents:    u.ZahlungenCents,
		AuszahlungenCents: u.AuszahlungenCents,
		AnzahlZahlungen:   u.AnzahlZahlungen,
	}
}

func toUmsatzServicekraftList(umsatz []reporting.UmsatzServicekraft) []umsatzServicekraft {
	out := make([]umsatzServicekraft, len(umsatz))
	for i := range umsatz {
		out[i] = toUmsatzServicekraft(umsatz[i])
	}
	return out
}

func toUmsatzTisch(u reporting.UmsatzTisch) umsatzTisch {
	return umsatzTisch{
		TischID:           u.TischID,
		TischName:         u.TischName,
		ZahlungenCents:    u.ZahlungenCents,
		AuszahlungenCents: u.AuszahlungenCents,
		AnzahlZahlungen:   u.AnzahlZahlungen,
	}
}

func toUmsatzTischList(tische []reporting.UmsatzTisch) []umsatzTisch {
	out := make([]umsatzTisch, len(tische))
	for i := range tische {
		out[i] = toUmsatzTisch(tische[i])
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
		ProduktName:  p.ProduktName,
		VarianteName: p.VarianteName,
		Menge:        p.Menge,
		Einzelpreis:  p.Einzelpreis,
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
		Zeitpunkt:   d.Zeitpunkt,
		TischID:     d.TischID,
		TischName:   d.TischName,
		UserID:      d.UserID,
		UserName:    d.UserName,
		BetragCents: d.BetragCents,
		Kommentar:   d.Kommentar,
		Positionen:  toStornierungPositionen(d.Positionen),
	}
}

func toStornierungDetails(details []reporting.StornierungDetail) []stornierungDetail {
	out := make([]stornierungDetail, len(details))
	for i := range details {
		out[i] = toStornierungDetail(details[i])
	}
	return out
}

func toReportingResponse(d reporting.ReportingData) reportingResponse {
	return reportingResponse{
		KassensitzungNr: d.KassensitzungNr,
		Summary: summaryResponse{
			GesamtUmsatzCents:        d.Summary.GesamtUmsatzCents,
			GesamtAuszahlungenCents:  d.Summary.GesamtAuszahlungenCents,
			GesamtBestellungenCents:  d.Summary.GesamtBestellungenCents,
			GesamtStornierungenCents: d.Summary.GesamtStornierungenCents,
			AnzahlBestellungen:       d.Summary.AnzahlBestellungen,
			AnzahlStornierungen:      d.Summary.AnzahlStornierungen,
			AnzahlDirektverkaeufe:    d.Summary.AnzahlDirektverkaeufe,
			DirektverkaufUmsatzCents: d.Summary.DirektverkaufUmsatzCents,
		},
		Breakdowns: breakdownsResponse{
			UmsatzProServicekraft: toUmsatzServicekraftList(d.Breakdowns.UmsatzProServicekraft),
			UmsatzProTisch:        toUmsatzTischList(d.Breakdowns.UmsatzProTisch),
		},
		UmsatzProSteuersatz: toUmsatzSteuersatzList(d.UmsatzProSteuersatz),
		Stornierungen:       toStornierungDetails(d.Stornierungen),
	}
}

type kassensitzungItem struct {
	ZNr         int    `json:"zNr"`
	Datum       string `json:"datum"`
	Bezeichnung string `json:"bezeichnung"`
	Status      string `json:"status"`
}

type getAllKassensitzungenResponse struct {
	Kassensitzungen []kassensitzungItem `json:"kassensitzungen"`
}

func (h QueryHandler) GetAllKassensitzungenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := h.Query.GetAllKassensitzungen(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}

		items := make([]kassensitzungItem, len(data))
		for i, k := range data {
			items[i] = kassensitzungItem{
				ZNr:         k.ZNr,
				Datum:       k.Datum.Format("2006-01-02"),
				Bezeichnung: k.Bezeichnung,
				Status:      string(k.Status),
			}
		}

		helper.SendResponse(w, getAllKassensitzungenResponse{Kassensitzungen: items})
	}
}

type eigeneUebersichtResponse struct {
	AnzahlBestellungen int `json:"anzahlBestellungen"`
	BestellungenCents  int `json:"bestellungenCents"`
	AnzahlZahlungen    int `json:"anzahlZahlungen"`
	ZahlungenCents     int `json:"zahlungenCents"`
}

func (h QueryHandler) GetEigeneUebersichtHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(middleware.UserIDKey).(int)
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
	GesamtAuszahlungenCents  int `json:"gesamtAuszahlungenCents"`
	GesamtBestellungenCents  int `json:"gesamtBestellungenCents"`
	GesamtStornierungenCents int `json:"gesamtStornierungenCents"`
	AnzahlBestellungen       int `json:"anzahlBestellungen"`
	AnzahlStornierungen      int `json:"anzahlStornierungen"`
	AnzahlDirektverkaeufe    int `json:"anzahlDirektverkaeufe"`
	DirektverkaufUmsatzCents int `json:"direktverkaufUmsatzCents"`
}

type liveReportingResponse struct {
	KassensitzungNr             int                    `json:"kassensitzungNr"`
	Bezeichnung                 string                 `json:"bezeichnung"`
	Datum                       time.Time              `json:"datum"`
	OffeneTische                []offenerTischResponse `json:"offeneTische"`
	OffeneSaldiCents            int                    `json:"offeneSaldiCents"`
	AusstehendAuszahlungenCents int                    `json:"ausstehendAuszahlungenCents"`
	Summary                     liveSummaryResponse    `json:"summary"`
	Breakdowns                  breakdownsResponse     `json:"breakdowns"`
	Stornierungen               []stornierungDetail    `json:"stornierungen"`
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
		KassensitzungNr:             d.KassensitzungNr,
		Bezeichnung:                 d.Bezeichnung,
		Datum:                       d.Datum,
		OffeneTische:                offeneTische,
		OffeneSaldiCents:            d.OffeneSaldiCents,
		AusstehendAuszahlungenCents: d.AusstehendAuszahlungenCents,
		Summary: liveSummaryResponse{
			GesamtUmsatzCents:        d.Summary.GesamtUmsatzCents,
			GesamtAuszahlungenCents:  d.Summary.GesamtAuszahlungenCents,
			GesamtBestellungenCents:  d.Summary.GesamtBestellungenCents,
			GesamtStornierungenCents: d.Summary.GesamtStornierungenCents,
			AnzahlBestellungen:       d.Summary.AnzahlBestellungen,
			AnzahlStornierungen:      d.Summary.AnzahlStornierungen,
			AnzahlDirektverkaeufe:    d.Summary.AnzahlDirektverkaeufe,
			DirektverkaufUmsatzCents: d.Summary.DirektverkaufUmsatzCents,
		},
		Breakdowns: breakdownsResponse{
			UmsatzProServicekraft: toUmsatzServicekraftList(d.Breakdowns.UmsatzProServicekraft),
			UmsatzProTisch:        toUmsatzTischList(d.Breakdowns.UmsatzProTisch),
		},
		Stornierungen: toStornierungDetails(d.Stornierungen),
	}
}

func (h QueryHandler) GetLiveReportingHandler() http.HandlerFunc {
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
