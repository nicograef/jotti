package http

import (
	"context"
	"errors"
	"net/http"

	z "github.com/Oudwins/zog"
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

var kassensitzungEroeffnenSchema = z.Struct(z.Shape{
	"Bezeichnung": z.String().Min(1, z.Message("Bezeichnung ist erforderlich")).Max(200, z.Message("Bezeichnung darf höchstens 200 Zeichen lang sein")).Required(),
	"BetragCents": z.Int().GTE(0, z.Message("Anfangsbestand darf nicht negativ sein")).Required(),
})

type kassensitzungEroeffnenResponse struct {
	ZNr int `json:"zNr"`
}

type geldtransitBuchenRequest struct {
	Richtung    string `json:"richtung"`
	BetragCents int    `json:"betragCents"`
	Kommentar   string `json:"kommentar"`
}

var geldtransitBuchenSchema = z.Struct(z.Shape{
	"Richtung": z.String().OneOf(
		[]string{"einlage", "entnahme"},
		z.Message("Ungültige Richtung"),
	).Required(),
	"BetragCents": z.Int().GTE(1, z.Message("Betrag muss mindestens 1 Cent sein")).Required(),
	"Kommentar":   z.String().Min(3, z.Message("Kommentar muss mindestens 3 Zeichen lang sein")).Max(200, z.Message("Kommentar darf höchstens 200 Zeichen lang sein")).Required(),
})

type kassensturzDurchfuehrenRequest struct {
	IstBestandCents int `json:"istBestandCents"`
}

var kassensturzDurchfuehrenSchema = z.Struct(z.Shape{
	"IstBestandCents": z.Int().GTE(0, z.Message("Ist-Bestand darf nicht negativ sein")).Required(),
})

// --- Handlers ---

func (h *CommandHandler) KassensitzungEroeffnenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := kassensitzungEroeffnenRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, kassensitzungEroeffnenSchema) {
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
		if !helper.ReadAndValidateBody(w, r, &body, geldtransitBuchenSchema) {
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
			switch {
			case errors.Is(err, kasseApp.ErrKonflikt):
				helper.SendConflictError(w)
			case errors.Is(err, kasseApp.ErrKasseNichtGeoeffnet):
				helper.SendConflict(w, "kasse_nicht_geoeffnet")
			default:
				helper.SendServerError(w)
			}
			return
		}

		helper.SendEmptyResponse(w)
	}
}

func (h *CommandHandler) KassensturzDurchfuehrenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := kassensturzDurchfuehrenRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, kassensturzDurchfuehrenSchema) {
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
			switch {
			case errors.Is(err, kasseApp.ErrKonflikt):
				helper.SendConflictError(w)
			case errors.Is(err, kasseApp.ErrKasseNichtGeoeffnet):
				helper.SendConflict(w, "kasse_nicht_geoeffnet")
			default:
				helper.SendServerError(w)
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
			switch {
			case errors.Is(err, kasseApp.ErrKonflikt):
				helper.SendConflictError(w)
			case errors.Is(err, kasseApp.ErrKasseNichtGeoeffnet):
				helper.SendConflict(w, "kasse_nicht_geoeffnet")
			default:
				helper.MapError(w, err, map[error]string{
					kasseApp.ErrKassensturzErforderlich: "kassensturz_erforderlich",
					kasseApp.ErrTischeSaldoOffen:        "tische_saldo_offen",
				})
			}
			return
		}

		helper.SendEmptyResponse(w)
	}
}
