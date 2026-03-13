package http

import (
	"context"
	"errors"
	"net/http"

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
	TischSnapshotErstellen(ctx context.Context, userID int, userName string, tischID int) error
	BestellungAufgeben(ctx context.Context, userID int, userName string, tischID int, positionen []application.BestellPositionInput, kommentar string) error
	ZahlungRegistrieren(ctx context.Context, userID int, userName string, tischID int, positionen []table.PositionRef, gesamtZahlungCents int, kommentar string) error
	ProdukteStornieren(ctx context.Context, userID int, userName string, tischID int, positionen []table.PositionRef, gesamtStornierungCents int, kommentar string) error
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
			if errors.Is(err, application.ErrTischAlreadyExists) {
				helper.SendClientError(w, "tisch_already_exists", nil)
				return
			} else if errors.Is(err, application.ErrInvalidTischData) {
				helper.SendClientError(w, "invalid_tisch_data", nil)
				return
			} else {
				helper.SendServerError(w)
				return
			}
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
			if errors.Is(err, application.ErrTischNotFound) {
				helper.SendClientError(w, "tisch_not_found", nil)
				return
			} else if errors.Is(err, application.ErrInvalidTischData) {
				helper.SendClientError(w, "invalid_tisch_data", nil)
				return
			} else {
				helper.SendServerError(w)
				return
			}
		}

		helper.SendEmptyResponse(w)
	}
}

type activateTisch struct {
	ID int `json:"id"`
}

func (h *CommandHandler) TischAktivierenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := activateTisch{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		err := h.Command.TischAktivieren(r.Context(), body.ID)
		if err != nil {
			if errors.Is(err, application.ErrTischNotFound) {
				helper.SendClientError(w, "tisch_not_found", nil)
				return
			} else {
				helper.SendServerError(w)
				return
			}
		}

		helper.SendEmptyResponse(w)
	}
}

type deactivateTisch struct {
	ID int `json:"id"`
}

func (h *CommandHandler) TischDeaktivierenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := deactivateTisch{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		err := h.Command.TischDeaktivieren(r.Context(), body.ID)
		if err != nil {
			if errors.Is(err, application.ErrTischNotFound) {
				helper.SendClientError(w, "tisch_not_found", nil)
				return
			} else {
				helper.SendServerError(w)
				return
			}
		}

		helper.SendEmptyResponse(w)
	}
}

type deleteTisch struct {
	ID int `json:"id"`
}

func (h *CommandHandler) TischLoeschenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := deleteTisch{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		err := h.Command.TischLoeschen(r.Context(), body.ID)
		if err != nil {
			if errors.Is(err, application.ErrTischNotFound) {
				helper.SendClientError(w, "tisch_not_found", nil)
				return
			} else {
				helper.SendServerError(w)
				return
			}
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
			switch {
			case errors.Is(err, application.ErrTischNotFound):
				helper.SendClientError(w, "tisch_not_found", nil)
			case errors.Is(err, application.ErrTischNotActive):
				helper.SendClientError(w, "tisch_not_active", nil)
			case errors.Is(err, application.ErrConflict):
				helper.SendConflictError(w)
			default:
				helper.SendServerError(w)
			}
			return
		}

		helper.SendEmptyResponse(w)
	}
}

type zahlungRegistrieren struct {
	TischID            int                 `json:"tischId"`
	Positionen         []table.PositionRef `json:"positionen"`
	GesamtZahlungCents int                 `json:"gesamtZahlungCents"`
	Kommentar          string              `json:"kommentar"`
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
		err := h.Command.ZahlungRegistrieren(r.Context(), userID, userName, body.TischID, body.Positionen, body.GesamtZahlungCents, body.Kommentar)
		if err != nil {
			switch {
			case errors.Is(err, application.ErrTischNotFound):
				helper.SendClientError(w, "tisch_not_found", nil)
			case errors.Is(err, application.ErrTischNotActive):
				helper.SendClientError(w, "tisch_not_active", nil)
			case errors.Is(err, application.ErrPositionNichtBezahlbar):
				helper.SendClientError(w, "position_nicht_bezahlbar", nil)
			case errors.Is(err, application.ErrConflict):
				helper.SendConflictError(w)
			default:
				helper.SendServerError(w)
			}
			return
		}

		helper.SendEmptyResponse(w)
	}
}

type produkteStornieren struct {
	TischID                int                 `json:"tischId"`
	Positionen             []table.PositionRef `json:"positionen"`
	GesamtStornierungCents int                 `json:"gesamtStornierungCents"`
	Kommentar              string              `json:"kommentar"`
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
		err := h.Command.ProdukteStornieren(r.Context(), userID, userName, body.TischID, body.Positionen, body.GesamtStornierungCents, body.Kommentar)
		if err != nil {
			switch {
			case errors.Is(err, application.ErrTischNotFound):
				helper.SendClientError(w, "tisch_not_found", nil)
			case errors.Is(err, application.ErrTischNotActive):
				helper.SendClientError(w, "tisch_not_active", nil)
			case errors.Is(err, application.ErrPositionNichtStornierbar):
				helper.SendClientError(w, "position_nicht_stornierbar", nil)
			case errors.Is(err, application.ErrConflict):
				helper.SendConflictError(w)
			default:
				helper.SendServerError(w)
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
			switch {
			case errors.Is(err, application.ErrTischNotFound):
				helper.SendClientError(w, "tisch_not_found", nil)
			case errors.Is(err, application.ErrTischNotActive):
				helper.SendClientError(w, "tisch_not_active", nil)
			case errors.Is(err, application.ErrPositionNichtLieferbar):
				helper.SendClientError(w, "position_nicht_lieferbar", nil)
			case errors.Is(err, application.ErrConflict):
				helper.SendConflictError(w)
			default:
				helper.SendServerError(w)
			}
			return
		}

		helper.SendEmptyResponse(w)
	}
}

type createTischSnapshot struct {
	TischID int `json:"tischId"`
}

func (h *CommandHandler) TischSnapshotErstellenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := createTischSnapshot{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		userID, ok := r.Context().Value(middleware.UserIDKey).(int)
		if !ok {
			helper.SendServerError(w)
			return
		}
		userName, _ := r.Context().Value(middleware.UserNameKey).(string)

		err := h.Command.TischSnapshotErstellen(r.Context(), userID, userName, body.TischID)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendEmptyResponse(w)
	}
}
