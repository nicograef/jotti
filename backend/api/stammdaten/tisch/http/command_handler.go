package http

import (
	"context"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/middleware"
	"github.com/nicograef/jotti/backend/api/stammdaten/tisch/application"
	tischDom "github.com/nicograef/jotti/backend/domain/tisch"
)

type command interface {
	TischErstellen(ctx context.Context, name string) (int, error)
	TischAktualisieren(ctx context.Context, id int, name string) error
	TischAktivieren(ctx context.Context, id int) error
	TischDeaktivieren(ctx context.Context, id int) error
	TischLoeschen(ctx context.Context, id int) error
	FavoritHinzufuegen(ctx context.Context, userID, tischID int) error
	FavoritEntfernen(ctx context.Context, userID, tischID int) error
}

type CommandHandler struct {
	Command command
}

type createTischRequest struct {
	Name string `json:"name"`
}

var createTischSchema = z.Struct(z.Shape{
	"Name": tischDom.TischNameSchema.Required(),
})

type createTischResponse struct {
	ID int `json:"id"`
}

func (h *CommandHandler) TischErstellenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := createTischRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, createTischSchema) {
			return
		}

		id, err := h.Command.TischErstellen(r.Context(), body.Name)
		if err != nil {
			helper.MapError(w, err, []helper.ErrorCode{
				{Err: application.ErrTischAlreadyExists, Code: "tisch_already_exists"},
				{Err: application.ErrInvalidTischData, Code: "invalid_tisch_data"},
			})
			return
		}

		helper.SendResponse(w, createTischResponse{ID: id})
	}
}

type updateTischRequest struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var updateTischSchema = z.Struct(z.Shape{
	"ID":   tischDom.TischIDSchema.Required(),
	"Name": tischDom.TischNameSchema.Required(),
})

func (h *CommandHandler) TischAktualisierenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := updateTischRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, updateTischSchema) {
			return
		}

		err := h.Command.TischAktualisieren(r.Context(), body.ID, body.Name)
		if err != nil {
			helper.MapError(w, err, []helper.ErrorCode{
				{Err: application.ErrTischNotFound, Code: "tisch_not_found"},
				{Err: application.ErrTischAlreadyExists, Code: "tisch_already_exists"},
				{Err: application.ErrInvalidTischData, Code: "invalid_tisch_data"},
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}

type favoritRequest struct {
	TischID int `json:"tischId"`
}

var favoritSchema = z.Struct(z.Shape{
	"TischID": tischDom.TischIDSchema.Required(),
})

func (h *CommandHandler) FavoritHinzufuegenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := favoritRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, favoritSchema) {
			return
		}

		userID, _, ok := middleware.UserFromContext(r.Context())
		if !ok {
			helper.SendServerError(w)
			return
		}

		err := h.Command.FavoritHinzufuegen(r.Context(), userID, body.TischID)
		if err != nil {
			helper.MapError(w, err, []helper.ErrorCode{
				{Err: application.ErrTischNotFound, Code: "tisch_not_found"},
				{Err: application.ErrTischNotActive, Code: "tisch_not_active"},
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}

func (h *CommandHandler) FavoritEntfernenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := favoritRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, favoritSchema) {
			return
		}

		userID, _, ok := middleware.UserFromContext(r.Context())
		if !ok {
			helper.SendServerError(w)
			return
		}

		err := h.Command.FavoritEntfernen(r.Context(), userID, body.TischID)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendEmptyResponse(w)
	}
}

// idRequest is the body of every endpoint that addresses one entity by its ID.
type idRequest struct {
	ID int `json:"id"`
}

var idSchema = z.Struct(z.Shape{
	"ID": tischDom.TischIDSchema.Required(),
})

func (h *CommandHandler) TischAktivierenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := idRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, idSchema) {
			return
		}

		err := h.Command.TischAktivieren(r.Context(), body.ID)
		if err != nil {
			helper.MapError(w, err, []helper.ErrorCode{
				{Err: application.ErrTischNotFound, Code: "tisch_not_found"},
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}

func (h *CommandHandler) TischDeaktivierenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := idRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, idSchema) {
			return
		}

		err := h.Command.TischDeaktivieren(r.Context(), body.ID)
		if err != nil {
			helper.MapError(w, err, []helper.ErrorCode{
				{Err: application.ErrTischNotFound, Code: "tisch_not_found"},
				{Err: application.ErrTischSaldoOffen, Code: "tisch_saldo_offen"},
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}

func (h *CommandHandler) TischLoeschenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := idRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, idSchema) {
			return
		}

		err := h.Command.TischLoeschen(r.Context(), body.ID)
		if err != nil {
			helper.MapError(w, err, []helper.ErrorCode{
				{Err: application.ErrTischNotFound, Code: "tisch_not_found"},
				{Err: application.ErrTischSaldoOffen, Code: "tisch_saldo_offen"},
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}
