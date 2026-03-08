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
	BestellungAufgeben(ctx context.Context, userID int, tischID int, positionen []table.Position, comment string) error
	ZahlungRegistrieren(ctx context.Context, userID int, tischID int, positionen []table.Position, comment string) error
	ProdukteStornieren(ctx context.Context, userID int, tischID int, positionen []table.Position, comment string) error
	ProdukteLiefern(ctx context.Context, userID int, tischID int, positionen []table.Position, comment string) error
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

type bestellungAufgeben struct {
	TischID    int              `json:"tischId"`
	Positionen []table.Position `json:"positionen"`
	Comment    string           `json:"comment"`
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
		err := h.Command.BestellungAufgeben(r.Context(), userID, body.TischID, body.Positionen, body.Comment)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendEmptyResponse(w)
	}
}

type zahlungRegistrieren struct {
	TischID    int              `json:"tischId"`
	Positionen []table.Position `json:"positionen"`
	Comment    string           `json:"comment"`
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
		err := h.Command.ZahlungRegistrieren(r.Context(), userID, body.TischID, body.Positionen, body.Comment)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendEmptyResponse(w)
	}
}

type produkteStornieren struct {
	TischID    int              `json:"tischId"`
	Positionen []table.Position `json:"positionen"`
	Comment    string           `json:"comment"`
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
		err := h.Command.ProdukteStornieren(r.Context(), userID, body.TischID, body.Positionen, body.Comment)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendEmptyResponse(w)
	}
}

type produkteLiefern struct {
	TischID    int              `json:"tischId"`
	Positionen []table.Position `json:"positionen"`
	Comment    string           `json:"comment"`
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
		err := h.Command.ProdukteLiefern(r.Context(), userID, body.TischID, body.Positionen, body.Comment)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendEmptyResponse(w)
	}
}
