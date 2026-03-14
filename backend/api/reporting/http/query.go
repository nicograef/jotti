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

type zeitraumDTO struct {
	Von time.Time `json:"von"`
	Bis time.Time `json:"bis"`
}

type umsatzServicekraftDTO struct {
	UserID          int    `json:"userId"`
	UserName        string `json:"userName"`
	ZahlungenCents  int    `json:"zahlungenCents"`
	AnzahlZahlungen int    `json:"anzahlZahlungen"`
}

type umsatzTischDTO struct {
	TischID         int    `json:"tischId"`
	TischName       string `json:"tischName"`
	ZahlungenCents  int    `json:"zahlungenCents"`
	AnzahlZahlungen int    `json:"anzahlZahlungen"`
}

type stornierungPositionDTO struct {
	ProduktName  string `json:"produktName"`
	VarianteName string `json:"varianteName"`
	Menge        int    `json:"menge"`
	Einzelpreis  int    `json:"einzelpreis"`
}

type stornierungDetailDTO struct {
	Zeitpunkt   time.Time                `json:"zeitpunkt"`
	TischID     int                      `json:"tischId"`
	TischName   string                   `json:"tischName"`
	UserID      int                      `json:"userId"`
	UserName    string                   `json:"userName"`
	BetragCents int                      `json:"betragCents"`
	Kommentar   string                   `json:"kommentar"`
	Positionen  []stornierungPositionDTO `json:"positionen"`
}

type dashboardDataDTO struct {
	GesamtUmsatzCents        int `json:"gesamtUmsatzCents"`
	AnzahlOffeneTische       int `json:"anzahlOffeneTische"`
	AnzahlBestellungen       int `json:"anzahlBestellungen"`
	AnzahlStornierungen      int `json:"anzahlStornierungen"`
	GesamtBestellungenCents  int `json:"gesamtBestellungenCents"`
	GesamtStornierungenCents int `json:"gesamtStornierungenCents"`
}

type tagesabrechnungDataDTO struct {
	Zeitraum                 zeitraumDTO             `json:"zeitraum"`
	GesamtUmsatzCents        int                     `json:"gesamtUmsatzCents"`
	GesamtBestellungenCents  int                     `json:"gesamtBestellungenCents"`
	GesamtStornierungenCents int                     `json:"gesamtStornierungenCents"`
	OffeneSaldiCents         int                     `json:"offeneSaldiCents"`
	AnzahlBestellungen       int                     `json:"anzahlBestellungen"`
	AnzahlStornierungen      int                     `json:"anzahlStornierungen"`
	UmsatzProServicekraft    []umsatzServicekraftDTO `json:"umsatzProServicekraft"`
	UmsatzProTisch           []umsatzTischDTO        `json:"umsatzProTisch"`
	Stornierungen            []stornierungDetailDTO  `json:"stornierungen"`
}

func toZeitraumDTO(z reporting.Zeitraum) zeitraumDTO {
	return zeitraumDTO{
		Von: z.Von,
		Bis: z.Bis,
	}
}

func toUmsatzServicekraftDTO(u reporting.UmsatzServicekraft) umsatzServicekraftDTO {
	return umsatzServicekraftDTO{
		UserID:          u.UserID,
		UserName:        u.UserName,
		ZahlungenCents:  u.ZahlungenCents,
		AnzahlZahlungen: u.AnzahlZahlungen,
	}
}

func toUmsatzServicekraftDTOs(umsatz []reporting.UmsatzServicekraft) []umsatzServicekraftDTO {
	out := make([]umsatzServicekraftDTO, len(umsatz))
	for i := range umsatz {
		out[i] = toUmsatzServicekraftDTO(umsatz[i])
	}
	return out
}

func toUmsatzTischDTO(u reporting.UmsatzTisch) umsatzTischDTO {
	return umsatzTischDTO{
		TischID:         u.TischID,
		TischName:       u.TischName,
		ZahlungenCents:  u.ZahlungenCents,
		AnzahlZahlungen: u.AnzahlZahlungen,
	}
}

func toUmsatzTischDTOs(tische []reporting.UmsatzTisch) []umsatzTischDTO {
	out := make([]umsatzTischDTO, len(tische))
	for i := range tische {
		out[i] = toUmsatzTischDTO(tische[i])
	}
	return out
}

func toStornierungPositionDTO(p reporting.StornierungPosition) stornierungPositionDTO {
	return stornierungPositionDTO{
		ProduktName:  p.ProduktName,
		VarianteName: p.VarianteName,
		Menge:        p.Menge,
		Einzelpreis:  p.Einzelpreis,
	}
}

func toStornierungPositionDTOs(positionen []reporting.StornierungPosition) []stornierungPositionDTO {
	out := make([]stornierungPositionDTO, len(positionen))
	for i := range positionen {
		out[i] = toStornierungPositionDTO(positionen[i])
	}
	return out
}

func toStornierungDetailDTO(d reporting.StornierungDetail) stornierungDetailDTO {
	return stornierungDetailDTO{
		Zeitpunkt:   d.Zeitpunkt,
		TischID:     d.TischID,
		TischName:   d.TischName,
		UserID:      d.UserID,
		UserName:    d.UserName,
		BetragCents: d.BetragCents,
		Kommentar:   d.Kommentar,
		Positionen:  toStornierungPositionDTOs(d.Positionen),
	}
}

func toStornierungDetailDTOs(details []reporting.StornierungDetail) []stornierungDetailDTO {
	out := make([]stornierungDetailDTO, len(details))
	for i := range details {
		out[i] = toStornierungDetailDTO(details[i])
	}
	return out
}

func toDashboardDataDTO(d reporting.DashboardData) dashboardDataDTO {
	return dashboardDataDTO{
		GesamtUmsatzCents:        d.GesamtUmsatzCents,
		AnzahlOffeneTische:       d.AnzahlOffeneTische,
		AnzahlBestellungen:       d.AnzahlBestellungen,
		AnzahlStornierungen:      d.AnzahlStornierungen,
		GesamtBestellungenCents:  d.GesamtBestellungenCents,
		GesamtStornierungenCents: d.GesamtStornierungenCents,
	}
}

func toTagesabrechnungDataDTO(d reporting.TagesabrechnungData) tagesabrechnungDataDTO {
	return tagesabrechnungDataDTO{
		Zeitraum:                 toZeitraumDTO(d.Zeitraum),
		GesamtUmsatzCents:        d.GesamtUmsatzCents,
		GesamtBestellungenCents:  d.GesamtBestellungenCents,
		GesamtStornierungenCents: d.GesamtStornierungenCents,
		OffeneSaldiCents:         d.OffeneSaldiCents,
		AnzahlBestellungen:       d.AnzahlBestellungen,
		AnzahlStornierungen:      d.AnzahlStornierungen,
		UmsatzProServicekraft:    toUmsatzServicekraftDTOs(d.UmsatzProServicekraft),
		UmsatzProTisch:           toUmsatzTischDTOs(d.UmsatzProTisch),
		Stornierungen:            toStornierungDetailDTOs(d.Stornierungen),
	}
}

func (h QueryHandler) GetDashboardHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := h.Query.GetDashboardData(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}
		helper.SendResponse(w, toDashboardDataDTO(data))
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
		helper.SendResponse(w, toTagesabrechnungDataDTO(data))
	}
}
