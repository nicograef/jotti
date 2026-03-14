package http

import (
	"context"
	"net/http"
	"time"

	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/domain/reporting"
)

type query interface {
	GetDashboardData(ctx context.Context) (reporting.DashboardData, error)
	GetTagesabrechnung(ctx context.Context, von, bis time.Time) (reporting.TagesabrechnungData, error)
}

type QueryHandler struct {
	Query query
}

type zeitraum struct {
	Von time.Time `json:"von"`
	Bis time.Time `json:"bis"`
}

type umsatzServicekraft struct {
	UserID          int    `json:"userId"`
	UserName        string `json:"userName"`
	ZahlungenCents  int    `json:"zahlungenCents"`
	AnzahlZahlungen int    `json:"anzahlZahlungen"`
}

type umsatzTisch struct {
	TischID         int    `json:"tischId"`
	TischName       string `json:"tischName"`
	ZahlungenCents  int    `json:"zahlungenCents"`
	AnzahlZahlungen int    `json:"anzahlZahlungen"`
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

type dashboardResponse struct {
	GesamtUmsatzCents        int `json:"gesamtUmsatzCents"`
	AnzahlOffeneTische       int `json:"anzahlOffeneTische"`
	AnzahlBestellungen       int `json:"anzahlBestellungen"`
	AnzahlStornierungen      int `json:"anzahlStornierungen"`
	GesamtBestellungenCents  int `json:"gesamtBestellungenCents"`
	GesamtStornierungenCents int `json:"gesamtStornierungenCents"`
}

type tagesabrechnungResponse struct {
	Zeitraum                 zeitraum             `json:"zeitraum"`
	GesamtUmsatzCents        int                  `json:"gesamtUmsatzCents"`
	GesamtBestellungenCents  int                  `json:"gesamtBestellungenCents"`
	GesamtStornierungenCents int                  `json:"gesamtStornierungenCents"`
	OffeneSaldiCents         int                  `json:"offeneSaldiCents"`
	AnzahlBestellungen       int                  `json:"anzahlBestellungen"`
	AnzahlStornierungen      int                  `json:"anzahlStornierungen"`
	UmsatzProServicekraft    []umsatzServicekraft `json:"umsatzProServicekraft"`
	UmsatzProTisch           []umsatzTisch        `json:"umsatzProTisch"`
	Stornierungen            []stornierungDetail  `json:"stornierungen"`
}

func toZeitraum(z reporting.Zeitraum) zeitraum {
	return zeitraum{
		Von: z.Von,
		Bis: z.Bis,
	}
}

func toUmsatzServicekraft(u reporting.UmsatzServicekraft) umsatzServicekraft {
	return umsatzServicekraft{
		UserID:          u.UserID,
		UserName:        u.UserName,
		ZahlungenCents:  u.ZahlungenCents,
		AnzahlZahlungen: u.AnzahlZahlungen,
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
		TischID:         u.TischID,
		TischName:       u.TischName,
		ZahlungenCents:  u.ZahlungenCents,
		AnzahlZahlungen: u.AnzahlZahlungen,
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

func toDashboardResponse(d reporting.DashboardData) dashboardResponse {
	return dashboardResponse{
		GesamtUmsatzCents:        d.GesamtUmsatzCents,
		AnzahlOffeneTische:       d.AnzahlOffeneTische,
		AnzahlBestellungen:       d.AnzahlBestellungen,
		AnzahlStornierungen:      d.AnzahlStornierungen,
		GesamtBestellungenCents:  d.GesamtBestellungenCents,
		GesamtStornierungenCents: d.GesamtStornierungenCents,
	}
}

func toTagesabrechnungResponse(d reporting.TagesabrechnungData) tagesabrechnungResponse {
	return tagesabrechnungResponse{
		Zeitraum:                 toZeitraum(d.Zeitraum),
		GesamtUmsatzCents:        d.GesamtUmsatzCents,
		GesamtBestellungenCents:  d.GesamtBestellungenCents,
		GesamtStornierungenCents: d.GesamtStornierungenCents,
		OffeneSaldiCents:         d.OffeneSaldiCents,
		AnzahlBestellungen:       d.AnzahlBestellungen,
		AnzahlStornierungen:      d.AnzahlStornierungen,
		UmsatzProServicekraft:    toUmsatzServicekraftList(d.UmsatzProServicekraft),
		UmsatzProTisch:           toUmsatzTischList(d.UmsatzProTisch),
		Stornierungen:            toStornierungDetails(d.Stornierungen),
	}
}

func (h QueryHandler) GetDashboardHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := h.Query.GetDashboardData(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}
		helper.SendResponse(w, toDashboardResponse(data))
	}
}

type getTagesabrechnungRequest struct {
	Von time.Time `json:"von"`
	Bis time.Time `json:"bis"`
}

func (h QueryHandler) GetTagesabrechnungHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := getTagesabrechnungRequest{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		if body.Von.IsZero() || body.Bis.IsZero() || !body.Von.Before(body.Bis) {
			helper.SendClientError(w, "invalid_zeitraum", nil)
			return
		}

		data, err := h.Query.GetTagesabrechnung(r.Context(), body.Von, body.Bis)
		if err != nil {
			helper.SendServerError(w)
			return
		}
		helper.SendResponse(w, toTagesabrechnungResponse(data))
	}
}
