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
	CreateTable(ctx context.Context, name string) (int, error)
	UpdateTable(ctx context.Context, id int, name string) error
	ActivateTable(ctx context.Context, id int) error
	DeactivateTable(ctx context.Context, id int) error
	PlaceTableOrder(ctx context.Context, userID int, tableID int, products []table.OrderProduct) error
}

type CommandHandler struct {
	Command command
}

type createTable struct {
	Name string `json:"name"`
}

type createTableResponse struct {
	ID int `json:"id"`
}

func (h *CommandHandler) CreateTableHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := createTable{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		id, err := h.Command.CreateTable(r.Context(), body.Name)
		if err != nil {
			if errors.Is(err, application.ErrTableAlreadyExists) {
				helper.SendClientError(w, "table_already_exists", nil)
				return
			} else {
				helper.SendServerError(w)
				return
			}
		}

		helper.SendResponse(w, createTableResponse{ID: id})
	}
}

type updateTable struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (h *CommandHandler) UpdateTableHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := updateTable{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		err := h.Command.UpdateTable(r.Context(), body.ID, body.Name)
		if err != nil {
			if errors.Is(err, application.ErrTableNotFound) {
				helper.SendClientError(w, "table_not_found", nil)
				return
			} else {
				helper.SendServerError(w)
				return
			}
		}

		helper.SendEmptyResponse(w)
	}
}

type activateTable struct {
	ID int `json:"id"`
}

func (h *CommandHandler) ActivateTableHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := activateTable{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		err := h.Command.ActivateTable(r.Context(), body.ID)
		if err != nil {
			if errors.Is(err, application.ErrTableNotFound) {
				helper.SendClientError(w, "table_not_found", nil)
				return
			} else {
				helper.SendServerError(w)
				return
			}
		}

		helper.SendEmptyResponse(w)
	}
}

type deactivateTable struct {
	ID int `json:"id"`
}

func (h *CommandHandler) DeactivateTableHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := deactivateTable{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		err := h.Command.DeactivateTable(r.Context(), body.ID)
		if err != nil {
			if errors.Is(err, application.ErrTableNotFound) {
				helper.SendClientError(w, "table_not_found", nil)
				return
			} else {
				helper.SendServerError(w)
				return
			}
		}

		helper.SendEmptyResponse(w)
	}
}

type placeTableOrder struct {
	TableID  int                  `json:"tableId"`
	Products []table.OrderProduct `json:"products"`
}

func (h *CommandHandler) PlaceTableOrderHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := placeTableOrder{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		userID := r.Context().Value(middleware.UserIDKey).(int)
		err := h.Command.PlaceTableOrder(r.Context(), userID, body.TableID, body.Products)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendEmptyResponse(w)
	}
}
