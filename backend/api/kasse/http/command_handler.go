package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/nicograef/jotti/backend/api/helper"
	kasseApp "github.com/nicograef/jotti/backend/api/kasse/application"
	"github.com/nicograef/jotti/backend/api/middleware"
)

type command interface {
	KassensitzungEroeffnen(ctx context.Context, userID int, userName string, bezeichnung string, betragCents int) (int, error)
	GeldtransitBuchen(ctx context.Context, userID int, userName string, richtung string, betragCents int, kommentar string) error
	KassensturzDurchfuehren(ctx context.Context, userID int, userName string, istBestandCents int) error
	TagesabschlussErstellen(ctx context.Context, userID int, userName string) error
}

type CommandHandler struct {
	Command command
}

// --- Request / Response DTOs ---

type kassensitzungEroeffnenRequest struct {
	Bezeichnung string `json:"bezeichnung"`
	BetragCents int    `json:"betragCents"`
}

type kassensitzungEroeffnenResponse struct {
	ZNr int `json:"zNr"`
}

type geldtransitBuchenRequest struct {
	Richtung    string `json:"richtung"`
	BetragCents int    `json:"betragCents"`
	Kommentar   string `json:"kommentar"`
}

type kassensturzDurchfuehrenRequest struct {
	IstBestandCents int `json:"istBestandCents"`
}

// --- Handlers ---

func (h *CommandHandler) KassensitzungEroeffnenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := kassensitzungEroeffnenRequest{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		userID, ok := r.Context().Value(middleware.UserIDKey).(int)
		if !ok {
			helper.SendServerError(w)
			return
		}
		userName, _ := r.Context().Value(middleware.UserNameKey).(string)

		zNr, err := h.Command.KassensitzungEroeffnen(r.Context(), userID, userName, body.Bezeichnung, body.BetragCents)
		if err != nil {
			helper.MapError(w, err, map[error]string{
				kasseApp.ErrKasseAlreadyOpen:           "kasse_bereits_geoeffnet",
				kasseApp.ErrBetreiberNichtKonfiguriert: "betreiber_nicht_konfiguriert",
			})
			return
		}

		helper.SendResponse(w, kassensitzungEroeffnenResponse{ZNr: zNr})
	}
}

func (h *CommandHandler) GeldtransitBuchenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := geldtransitBuchenRequest{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		if body.Richtung == "" {
			helper.SendClientError(w, "richtung_erforderlich", nil)
			return
		}

		userID, ok := r.Context().Value(middleware.UserIDKey).(int)
		if !ok {
			helper.SendServerError(w)
			return
		}
		userName, _ := r.Context().Value(middleware.UserNameKey).(string)

		err := h.Command.GeldtransitBuchen(r.Context(), userID, userName, body.Richtung, body.BetragCents, body.Kommentar)
		if err != nil {
			if errors.Is(err, kasseApp.ErrKonflikt) {
				helper.SendConflictError(w)
			} else {
				helper.MapError(w, err, map[error]string{
					kasseApp.ErrKasseNichtGeoeffnet: "kasse_nicht_geoeffnet",
				})
			}
			return
		}

		helper.SendEmptyResponse(w)
	}
}

func (h *CommandHandler) KassensturzDurchfuehrenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := kassensturzDurchfuehrenRequest{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		userID, ok := r.Context().Value(middleware.UserIDKey).(int)
		if !ok {
			helper.SendServerError(w)
			return
		}
		userName, _ := r.Context().Value(middleware.UserNameKey).(string)

		err := h.Command.KassensturzDurchfuehren(r.Context(), userID, userName, body.IstBestandCents)
		if err != nil {
			if errors.Is(err, kasseApp.ErrKonflikt) {
				helper.SendConflictError(w)
			} else {
				helper.MapError(w, err, map[error]string{
					kasseApp.ErrKasseNichtGeoeffnet: "kasse_nicht_geoeffnet",
				})
			}
			return
		}

		helper.SendEmptyResponse(w)
	}
}

func (h *CommandHandler) TagesabschlussErstellenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(middleware.UserIDKey).(int)
		if !ok {
			helper.SendServerError(w)
			return
		}
		userName, _ := r.Context().Value(middleware.UserNameKey).(string)

		err := h.Command.TagesabschlussErstellen(r.Context(), userID, userName)
		if err != nil {
			if errors.Is(err, kasseApp.ErrKonflikt) {
				helper.SendConflictError(w)
			} else {
				helper.MapError(w, err, map[error]string{
					kasseApp.ErrKasseNichtGeoeffnet:     "kasse_nicht_geoeffnet",
					kasseApp.ErrKassensturzErforderlich: "kassensturz_erforderlich",
					kasseApp.ErrTischeSaldoOffen:        "tische_saldo_offen",
				})
			}
			return
		}

		helper.SendEmptyResponse(w)
	}
}
