package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/settings/application"
	"github.com/nicograef/jotti/backend/domain/settings"
)

type settingsQuery interface {
	GetKassenidentitaet(ctx context.Context) (settings.Kassenidentitaet, error)
	GetBetreiber(ctx context.Context) (settings.Betreiber, error)
	GetBondruckEinstellungen(ctx context.Context) (settings.BondruckEinstellungen, error)
}

type QueryHandler struct {
	Query settingsQuery
}

type kassenidentitaetResponse struct {
	Seriennummer string    `json:"seriennummer"`
	AngelegtAm   time.Time `json:"angelegtAm"`
}

type betreiberResponse struct {
	Vereinsname  string  `json:"vereinsname"`
	Strasse      string  `json:"strasse"`
	Plz          string  `json:"plz"`
	Ort          string  `json:"ort"`
	Steuernummer *string `json:"steuernummer"`
	UstID        *string `json:"ustId"`
}

type bondruckEinstellungenResponse struct {
	KassenbelegDruckerIP string `json:"kassenbelegDruckerIp"`
}

func (h *QueryHandler) GetKassenidentitaetHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identitaet, err := h.Query.GetKassenidentitaet(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}
		helper.SendResponse(w, kassenidentitaetResponse{
			Seriennummer: identitaet.Seriennummer.String(),
			AngelegtAm:   identitaet.AngelegtAm,
		})
	}
}

func (h *QueryHandler) GetBetreiberHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := h.Query.GetBetreiber(r.Context())
		if err != nil {
			if errors.Is(err, application.ErrNotFound) {
				helper.SendResponse(w, betreiberResponse{})
				return
			}
			helper.SendServerError(w)
			return
		}
		helper.SendResponse(w, betreiberResponse{
			Vereinsname:  b.Vereinsname,
			Strasse:      b.Strasse,
			Plz:          b.Plz,
			Ort:          b.Ort,
			Steuernummer: b.Steuernummer,
			UstID:        b.UstID,
		})
	}
}

func (h *QueryHandler) GetBondruckEinstellungenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := h.Query.GetBondruckEinstellungen(r.Context())
		if err != nil {
			if errors.Is(err, application.ErrNotFound) {
				helper.SendResponse(w, bondruckEinstellungenResponse{})
				return
			}
			helper.SendServerError(w)
			return
		}
		helper.SendResponse(w, bondruckEinstellungenResponse{
			KassenbelegDruckerIP: b.KassenbelegDruckerIP,
		})
	}
}
