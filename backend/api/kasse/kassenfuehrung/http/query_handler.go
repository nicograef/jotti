package http

import (
	"context"
	"net/http"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/domain/kasse"
)

type query interface {
	GetAktiveKassensitzung(ctx context.Context) (*kasse.Kassensitzung, error)
	GetKassenbestand(ctx context.Context, kassensitzungNr int) (kasse.Kassenbestand, error)
	GetGeldtransitListe(ctx context.Context, kassensitzungNr int) ([]kasse.Geldtransit, error)
}

type QueryHandler struct {
	Query query
}

type aktiveKassensitzungResponse struct {
	ZNr         int    `json:"zNr"`
	Datum       string `json:"datum"`
	Bezeichnung string `json:"bezeichnung"`
	Status      string `json:"status"`
	EroeffnetAm string `json:"eroeffnetAm"`
}

type kassensitzungNrRequest struct {
	KassensitzungNr int `json:"kassensitzungNr"`
}

var kassensitzungNrSchema = z.Struct(z.Shape{
	"KassensitzungNr": kasse.KassensitzungNrSchema.Required(),
})

// Invariante (vor Kassensturz): anfangsbestand + bareinnahmen + einlagen − entnahmen = sollBestand.
type kassenbestandResponse struct {
	SollBestandCents    int `json:"sollBestandCents"`
	AnfangsbestandCents int `json:"anfangsbestandCents"`
	BareinnahmenCents   int `json:"bareinnahmenCents"`
	EinlagenCents       int `json:"einlagenCents"`
	EntnahmenCents      int `json:"entnahmenCents"`
}

type geldtransitItemResponse struct {
	Zeitpunkt   string `json:"zeitpunkt"`
	Richtung    string `json:"richtung"`
	BetragCents int    `json:"betragCents"`
	Kommentar   string `json:"kommentar"`
	GebuchtVon  string `json:"gebuchtVon"`
}

func (h *QueryHandler) GetAktiveKassensitzungHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ks, err := h.Query.GetAktiveKassensitzung(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}

		if ks == nil {
			helper.SendResponse(w, nil)
			return
		}

		helper.SendResponse(w, aktiveKassensitzungResponse{
			ZNr:         ks.ZNr,
			Datum:       ks.Datum.Format("2006-01-02"),
			Bezeichnung: ks.Bezeichnung,
			Status:      string(ks.Status),
			EroeffnetAm: ks.CreatedAt.Format(time.RFC3339),
		})
	}
}

func (h *QueryHandler) GetKassenbestandHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := kassensitzungNrRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, kassensitzungNrSchema) {
			return
		}

		bestand, err := h.Query.GetKassenbestand(r.Context(), body.KassensitzungNr)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, kassenbestandResponse{
			SollBestandCents:    bestand.SollBestandCents,
			AnfangsbestandCents: bestand.AnfangsbestandCents,
			BareinnahmenCents:   bestand.BareinnahmenCents,
			EinlagenCents:       bestand.EinlagenCents,
			EntnahmenCents:      bestand.EntnahmenCents,
		})
	}
}

func (h *QueryHandler) GetGeldtransitListeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := kassensitzungNrRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, kassensitzungNrSchema) {
			return
		}

		buchungen, err := h.Query.GetGeldtransitListe(r.Context(), body.KassensitzungNr)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		items := make([]geldtransitItemResponse, 0, len(buchungen))
		for _, b := range buchungen {
			items = append(items, geldtransitItemResponse{
				Zeitpunkt:   b.Zeitpunkt.Format(time.RFC3339),
				Richtung:    b.Richtung,
				BetragCents: b.BetragCents,
				Kommentar:   b.Kommentar,
				GebuchtVon:  b.GebuchtVon,
			})
		}

		helper.SendResponse(w, items)
	}
}
