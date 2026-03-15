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
	GetReporting(ctx context.Context, zeitraum reporting.Zeitraum) (reporting.ReportingData, error)
	GetEigeneUebersicht(ctx context.Context, userID int) (reporting.EigeneUebersicht, error)
}

type QueryHandler struct {
	Query query
}

type getReportingRequest struct {
	Von time.Time `json:"von"`
	Bis time.Time `json:"bis"`
}

type reportingResponse struct {
	Zeitraum      zeitraum            `json:"zeitraum"`
	Summary       summaryResponse     `json:"summary"`
	Breakdowns    breakdownsResponse  `json:"breakdowns"`
	Stornierungen []stornierungDetail `json:"stornierungen"`
}

func (h QueryHandler) GetReportingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := getReportingRequest{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		if body.Von.IsZero() || body.Bis.IsZero() {
			helper.SendClientError(w, "invalid_zeitraum", "von_und_bis_sind_erforderlich")
			return
		}

		if !isUTC(body.Von) || !isUTC(body.Bis) {
			helper.SendClientError(w, "invalid_timezone", "von_und_bis_muessen_utc_sein")
			return
		}

		if !body.Von.Before(body.Bis) {
			helper.SendClientError(w, "invalid_zeitraum", "von_muss_kleiner_als_bis_sein")
			return
		}

		zeitraum := reporting.Zeitraum{
			Von: body.Von,
			Bis: body.Bis,
		}

		data, err := h.Query.GetReporting(r.Context(), zeitraum)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, toReportingResponse(data))
	}
}

func isUTC(ts time.Time) bool {
	_, offset := ts.Zone()
	return offset == 0
}

type zeitraum struct {
	Von time.Time `json:"von"`
	Bis time.Time `json:"bis"`
}

type summaryResponse struct {
	GesamtUmsatzCents           int `json:"gesamtUmsatzCents"`
	GesamtAuszahlungenCents     int `json:"gesamtAuszahlungenCents"`
	GesamtBestellungenCents     int `json:"gesamtBestellungenCents"`
	GesamtStornierungenCents    int `json:"gesamtStornierungenCents"`
	OffeneSaldiCents            int `json:"offeneSaldiCents"`
	AusstehendAuszahlungenCents int `json:"ausstehendAuszahlungenCents"`
	AnzahlOffeneTische          int `json:"anzahlOffeneTische"`
	AnzahlBestellungen          int `json:"anzahlBestellungen"`
	AnzahlStornierungen         int `json:"anzahlStornierungen"`
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

func toZeitraum(z reporting.Zeitraum) zeitraum {
	return zeitraum{
		Von: z.Von,
		Bis: z.Bis,
	}
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
		Zeitraum: toZeitraum(d.Zeitraum),
		Summary: summaryResponse{
			GesamtUmsatzCents:           d.Summary.GesamtUmsatzCents,
			GesamtAuszahlungenCents:     d.Summary.GesamtAuszahlungenCents,
			GesamtBestellungenCents:     d.Summary.GesamtBestellungenCents,
			GesamtStornierungenCents:    d.Summary.GesamtStornierungenCents,
			OffeneSaldiCents:            d.Summary.OffeneSaldiCents,
			AusstehendAuszahlungenCents: d.Summary.AusstehendAuszahlungenCents,
			AnzahlOffeneTische:          d.Summary.AnzahlOffeneTische,
			AnzahlBestellungen:          d.Summary.AnzahlBestellungen,
			AnzahlStornierungen:         d.Summary.AnzahlStornierungen,
		},
		Breakdowns: breakdownsResponse{
			UmsatzProServicekraft: toUmsatzServicekraftList(d.Breakdowns.UmsatzProServicekraft),
			UmsatzProTisch:        toUmsatzTischList(d.Breakdowns.UmsatzProTisch),
		},
		Stornierungen: toStornierungDetails(d.Stornierungen),
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
