package table_admin

import (
	"context"
	"errors"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api"
)

type command interface {
	CreateTable(ctx context.Context, name string) (int, error)
	UpdateTable(ctx context.Context, id int, name string) error
	ActivateTable(ctx context.Context, id int) error
	DeactivateTable(ctx context.Context, id int) error
}

type CommandHandler struct {
	Command command
}

type createTable struct {
	Name string `json:"name"`
}

var createTableSchema = z.Struct(z.Shape{
	"Name": NameSchema.Required(),
})

type createTableResponse struct {
	ID int `json:"id"`
}

// CreateTableHandler handles requests to create a new table.
func (h *CommandHandler) CreateTableHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := createTable{}
		if !api.ReadAndValidateBody(w, r, &body, createTableSchema) {
			return
		}

		ctx := r.Context()
		id, err := h.Command.CreateTable(ctx, body.Name)
		if err != nil {
			if errors.Is(err, ErrTableAlreadyExists) {
				api.SendClientError(w, "table_already_exists", nil)
				return
			} else {
				api.SendServerError(w)
				return
			}
		}

		api.SendResponse(w, createTableResponse{ID: id})
	}
}

type updateTable struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var updateTableSchema = z.Struct(z.Shape{
	"ID":   IDSchema.Required(),
	"Name": NameSchema.Required(),
})

// UpdateTableHandler handles requests to update an existing table.
func (h *CommandHandler) UpdateTableHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := updateTable{}
		if !api.ReadAndValidateBody(w, r, &body, updateTableSchema) {
			return
		}

		ctx := r.Context()
		err := h.Command.UpdateTable(ctx, body.ID, body.Name)
		if err != nil {
			if errors.Is(err, ErrTableNotFound) {
				api.SendClientError(w, "table_not_found", nil)
				return
			} else {
				api.SendServerError(w)
				return
			}
		}

		api.SendEmptyResponse(w)
	}
}

type activateTable struct {
	ID int `json:"id"`
}

var activateTableSchema = z.Struct(z.Shape{
	"ID": IDSchema.Required(),
})

// ActivateTableHandler handles requests to activate a table.
func (h *CommandHandler) ActivateTableHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := activateTable{}
		if !api.ReadAndValidateBody(w, r, &body, activateTableSchema) {
			return
		}

		ctx := r.Context()
		err := h.Command.ActivateTable(ctx, body.ID)
		if err != nil {
			if errors.Is(err, ErrTableNotFound) {
				api.SendClientError(w, "table_not_found", nil)
				return
			} else {
				api.SendServerError(w)
				return
			}
		}

		api.SendEmptyResponse(w)
	}
}

type deactivateTable struct {
	ID int `json:"id"`
}

var deactivateTableSchema = z.Struct(z.Shape{
	"ID": IDSchema.Required(),
})

// DeactivateTableHandler handles requests to deactivate a table.
func (h *CommandHandler) DeactivateTableHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := deactivateTable{}
		if !api.ReadAndValidateBody(w, r, &body, deactivateTableSchema) {
			return
		}

		ctx := r.Context()
		err := h.Command.DeactivateTable(ctx, body.ID)
		if err != nil {
			if errors.Is(err, ErrTableNotFound) {
				api.SendClientError(w, "table_not_found", nil)
				return
			} else {
				api.SendServerError(w)
				return
			}
		}

		api.SendEmptyResponse(w)
	}
}
