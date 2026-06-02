package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/nicograef/jotti/backend/api/helper"
	kasseApp "github.com/nicograef/jotti/backend/api/kasse/application"
	"github.com/nicograef/jotti/backend/api/middleware"
	"github.com/nicograef/jotti/backend/domain/kasse"
)

type command interface {
	KassensitzungEroeffnen(ctx context.Context, userID int, userName string, datum time.Time, bezeichnung string) (int, error)
	AnfangsbestandSetzen(ctx context.Context, userID int, userName string, betragCents int) error
	KassenbewegungBuchen(ctx context.Context, userID int, userName string, art string, betragCents int, kommentar string) error
	KassensturzDurchfuehren(ctx context.Context, userID int, userName string, istBestandCents int) error
	TagesabschlussErstellen(ctx context.Context, userID int, userName string) error
}

type query interface {
	GetOffeneKassensitzung(ctx context.Context) (*kasse.Kassensitzung, error)
	GetKassenbestand(ctx context.Context, kassensitzungNr int) (int, error)
}

type Handler struct {
	Command command
	Query   query
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

type offeneKassensitzungResponse struct {
	ZNr         int    `json:"zNr"`
	Datum       string `json:"datum"`
	Bezeichnung string `json:"bezeichnung"`
	Status      string `json:"status"`
}

type kassenbestandRequest struct {
	KassensitzungNr int `json:"kassensitzungNr"`
}

type kassenbestandResponse struct {
	SollBestandCents int `json:"sollBestandCents"`
}

// --- Handlers ---

func (h *Handler) KassensitzungEroeffnenHandler() http.HandlerFunc {
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

func (h *Handler) AnfangsbestandSetzenHandler() http.HandlerFunc {
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

func (h *Handler) KassenbewegungBuchenHandler() http.HandlerFunc {
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

func (h *Handler) KassensturzDurchfuehrenHandler() http.HandlerFunc {
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

func (h *Handler) TagesabschlussErstellenHandler() http.HandlerFunc {
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

func (h *Handler) GetOffeneKassensitzungHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ks, err := h.Query.GetOffeneKassensitzung(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}

		if ks == nil {
			helper.SendResponse(w, nil)
			return
		}

		helper.SendResponse(w, offeneKassensitzungResponse{
			ZNr:         ks.ZNr,
			Datum:       ks.Datum.Format("2006-01-02"),
			Bezeichnung: ks.Bezeichnung,
			Status:      ks.Status,
		})
	}
}

func (h *Handler) GetKassenbestandHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := kassenbestandRequest{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		if body.KassensitzungNr < 1 {
			helper.SendClientError(w, "invalid_kassensitzung_nr", nil)
			return
		}

		bestand, err := h.Query.GetKassenbestand(r.Context(), body.KassensitzungNr)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, kassenbestandResponse{SollBestandCents: bestand})
	}
}
