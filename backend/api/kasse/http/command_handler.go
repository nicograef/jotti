package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/nicograef/jotti/backend/api/helper"
	kasseApp "github.com/nicograef/jotti/backend/api/kasse/application"
	"github.com/nicograef/jotti/backend/api/middleware"
)

type command interface {
	KassensitzungEroeffnen(ctx context.Context, userID int, userName string, datum time.Time, bezeichnung string) (int, error)
	AnfangsbestandSetzen(ctx context.Context, userID int, userName string, betragCents int) error
	KassenbewegungBuchen(ctx context.Context, userID int, userName string, art string, betragCents int, kommentar string) error
	KassensturzDurchfuehren(ctx context.Context, userID int, userName string, istBestandCents int) error
	TagesabschlussErstellen(ctx context.Context, userID int, userName string) error
}

type CommandHandler struct {
	Command command
}

// --- Request / Response DTOs ---

type kassensitzungEroeffnenRequest struct {
	Datum       string `json:"datum"`
	Bezeichnung string `json:"bezeichnung"`
}

type kassensitzungEroeffnenResponse struct {
	ZNr int `json:"zNr"`
}

type anfangsbestandSetzenRequest struct {
	BetragCents int `json:"betragCents"`
}

type kassenbewegungBuchenRequest struct {
	Art         string `json:"art"`
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

		if body.Datum == "" {
			helper.SendClientError(w, "datum_erforderlich", nil)
			return
		}

		datum, err := time.Parse("2006-01-02", body.Datum)
		if err != nil {
			helper.SendClientError(w, "invalid_datum", nil)
			return
		}

		userID, ok := r.Context().Value(middleware.UserIDKey).(int)
		if !ok {
			helper.SendServerError(w)
			return
		}
		userName, _ := r.Context().Value(middleware.UserNameKey).(string)

		zNr, err := h.Command.KassensitzungEroeffnen(r.Context(), userID, userName, datum, body.Bezeichnung)
		if err != nil {
			helper.MapError(w, err, map[error]string{
				kasseApp.ErrKasseAlreadyOpen: "kasse_bereits_geoeffnet",
			})
			return
		}

		helper.SendResponse(w, kassensitzungEroeffnenResponse{ZNr: zNr})
	}
}

func (h *CommandHandler) AnfangsbestandSetzenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := anfangsbestandSetzenRequest{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		userID, ok := r.Context().Value(middleware.UserIDKey).(int)
		if !ok {
			helper.SendServerError(w)
			return
		}
		userName, _ := r.Context().Value(middleware.UserNameKey).(string)

		err := h.Command.AnfangsbestandSetzen(r.Context(), userID, userName, body.BetragCents)
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

func (h *CommandHandler) KassenbewegungBuchenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := kassenbewegungBuchenRequest{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		if body.Art == "" {
			helper.SendClientError(w, "art_erforderlich", nil)
			return
		}

		userID, ok := r.Context().Value(middleware.UserIDKey).(int)
		if !ok {
			helper.SendServerError(w)
			return
		}
		userName, _ := r.Context().Value(middleware.UserNameKey).(string)

		err := h.Command.KassenbewegungBuchen(r.Context(), userID, userName, body.Art, body.BetragCents, body.Kommentar)
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
