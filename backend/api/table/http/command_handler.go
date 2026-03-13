package http

import (
	"context"
	"errors"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/middleware"
	"github.com/nicograef/jotti/backend/api/table/application"
	"github.com/nicograef/jotti/backend/domain/table"
)

type command interface {
	TischErstellen(ctx context.Context, name string) (int, error)
	TischAktualisieren(ctx context.Context, id int, name string) error
	TischAktivieren(ctx context.Context, id int) error
	TischDeaktivieren(ctx context.Context, id int) error
	TischLoeschen(ctx context.Context, id int) error
	BestellungAufgeben(ctx context.Context, userID int, userName string, tischID int, positionen []application.BestellPositionInput, kommentar string) error
	ZahlungRegistrieren(ctx context.Context, userID int, userName string, tischID int, positionen []table.PositionRef, kommentar string) error
	ProdukteStornieren(ctx context.Context, userID int, userName string, tischID int, positionen []table.PositionRef, kommentar string) error
	ProdukteLiefern(ctx context.Context, userID int, userName string, tischID int, positionen []table.PositionRef, kommentar string) error
}

type CommandHandler struct {
	Command command
}

type createTisch struct {
	Name string `json:"name"`
}

type createTischResponse struct {
	ID int `json:"id"`
}

func (h *CommandHandler) TischErstellenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := createTisch{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		id, err := h.Command.TischErstellen(r.Context(), body.Name)
		if err != nil {
			helper.MapError(w, err, map[error]string{
				application.ErrTischAlreadyExists: "tisch_already_exists",
				application.ErrInvalidTischData:   "invalid_tisch_data",
			})
			return
		}

		helper.SendResponse(w, createTischResponse{ID: id})
	}
}

type updateTisch struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (h *CommandHandler) TischAktualisierenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := updateTisch{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		err := h.Command.TischAktualisieren(r.Context(), body.ID, body.Name)
		if err != nil {
			helper.MapError(w, err, map[error]string{
				application.ErrTischNotFound:    "tisch_not_found",
				application.ErrInvalidTischData: "invalid_tisch_data",
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}

type activateTisch struct {
	ID int `json:"id"`
}

var activateTischSchema = z.Struct(z.Shape{
	"ID": table.TischIDSchema.Required(),
})

func (h *CommandHandler) TischAktivierenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := activateTisch{}
		if !helper.ReadAndValidateBody(w, r, &body, activateTischSchema) {
			return
		}

		err := h.Command.TischAktivieren(r.Context(), body.ID)
		if err != nil {
			helper.MapError(w, err, map[error]string{
				application.ErrTischNotFound: "tisch_not_found",
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}

type deactivateTisch struct {
	ID int `json:"id"`
}

var deactivateTischSchema = z.Struct(z.Shape{
	"ID": table.TischIDSchema.Required(),
})

func (h *CommandHandler) TischDeaktivierenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := deactivateTisch{}
		if !helper.ReadAndValidateBody(w, r, &body, deactivateTischSchema) {
			return
		}

		err := h.Command.TischDeaktivieren(r.Context(), body.ID)
		if err != nil {
			helper.MapError(w, err, map[error]string{
				application.ErrTischNotFound: "tisch_not_found",
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}

type deleteTisch struct {
	ID int `json:"id"`
}

var deleteTischSchema = z.Struct(z.Shape{
	"ID": table.TischIDSchema.Required(),
})

func (h *CommandHandler) TischLoeschenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := deleteTisch{}
		if !helper.ReadAndValidateBody(w, r, &body, deleteTischSchema) {
			return
		}

		err := h.Command.TischLoeschen(r.Context(), body.ID)
		if err != nil {
			helper.MapError(w, err, map[error]string{
				application.ErrTischNotFound: "tisch_not_found",
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}

type bestellungAufgeben struct {
	TischID    int                                `json:"tischId"`
	Positionen []application.BestellPositionInput `json:"positionen"`
	Kommentar  string                             `json:"kommentar"`
}

func (h *CommandHandler) BestellungAufgebenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := bestellungAufgeben{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		userID, ok := r.Context().Value(middleware.UserIDKey).(int)
		if !ok {
			helper.SendServerError(w)
			return
		}
		userName, _ := r.Context().Value(middleware.UserNameKey).(string)
		err := h.Command.BestellungAufgeben(r.Context(), userID, userName, body.TischID, body.Positionen, body.Kommentar)
		if err != nil {
			if errors.Is(err, application.ErrConflict) {
				helper.SendConflictError(w)
			} else {
				helper.MapError(w, err, map[error]string{
					application.ErrTischNotFound:   "tisch_not_found",
					application.ErrTischNotActive:  "tisch_not_active",
					application.ErrProduktNotFound: "produkt_not_found",
				})
			}
			return
		}

		helper.SendEmptyResponse(w)
	}
}

type zahlungRegistrieren struct {
	TischID    int                 `json:"tischId"`
	Positionen []table.PositionRef `json:"positionen"`
	Kommentar  string              `json:"kommentar"`
}

func (h *CommandHandler) ZahlungRegistrierenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := zahlungRegistrieren{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		userID, ok := r.Context().Value(middleware.UserIDKey).(int)
		if !ok {
			helper.SendServerError(w)
			return
		}
		userName, _ := r.Context().Value(middleware.UserNameKey).(string)
		err := h.Command.ZahlungRegistrieren(r.Context(), userID, userName, body.TischID, body.Positionen, body.Kommentar)
		if err != nil {
			if errors.Is(err, application.ErrConflict) {
				helper.SendConflictError(w)
			} else {
				helper.MapError(w, err, map[error]string{
					application.ErrTischNotFound:          "tisch_not_found",
					application.ErrTischNotActive:         "tisch_not_active",
					application.ErrPositionNichtBezahlbar: "position_nicht_bezahlbar",
				})
			}
			return
		}

		helper.SendEmptyResponse(w)
	}
}

type produkteStornieren struct {
	TischID    int                 `json:"tischId"`
	Positionen []table.PositionRef `json:"positionen"`
	Kommentar  string              `json:"kommentar"`
}

func (h *CommandHandler) ProdukteStornierenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := produkteStornieren{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		userID, ok := r.Context().Value(middleware.UserIDKey).(int)
		if !ok {
			helper.SendServerError(w)
			return
		}
		userName, _ := r.Context().Value(middleware.UserNameKey).(string)
		err := h.Command.ProdukteStornieren(r.Context(), userID, userName, body.TischID, body.Positionen, body.Kommentar)
		if err != nil {
			if errors.Is(err, application.ErrConflict) {
				helper.SendConflictError(w)
			} else {
				helper.MapError(w, err, map[error]string{
					application.ErrTischNotFound:            "tisch_not_found",
					application.ErrTischNotActive:           "tisch_not_active",
					application.ErrPositionNichtStornierbar: "position_nicht_stornierbar",
				})
			}
			return
		}

		helper.SendEmptyResponse(w)
	}
}

type produkteLiefern struct {
	TischID    int                 `json:"tischId"`
	Positionen []table.PositionRef `json:"positionen"`
	Kommentar  string              `json:"kommentar"`
}

func (h *CommandHandler) ProdukteLiefernHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := produkteLiefern{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		userID, ok := r.Context().Value(middleware.UserIDKey).(int)
		if !ok {
			helper.SendServerError(w)
			return
		}
		userName, _ := r.Context().Value(middleware.UserNameKey).(string)
		err := h.Command.ProdukteLiefern(r.Context(), userID, userName, body.TischID, body.Positionen, body.Kommentar)
		if err != nil {
			if errors.Is(err, application.ErrConflict) {
				helper.SendConflictError(w)
			} else {
				helper.MapError(w, err, map[error]string{
					application.ErrTischNotFound:          "tisch_not_found",
					application.ErrTischNotActive:         "tisch_not_active",
					application.ErrPositionNichtLieferbar: "position_nicht_lieferbar",
				})
			}
			return
		}

		helper.SendEmptyResponse(w)
	}
}
