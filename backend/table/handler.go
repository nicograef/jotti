package table

import (
	"context"
	"errors"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api"
)

type service interface {
	CreateTable(ctx context.Context, name string) (*Table, error)
	UpdateTable(ctx context.Context, id int, name string) (*Table, error)
	ActivateTable(ctx context.Context, id int) error
	DeactivateTable(ctx context.Context, id int) error
	GetTable(ctx context.Context, id int) (*Table, error)
	GetAllTables(ctx context.Context) ([]*Table, error)
	GetActiveTables(ctx context.Context) ([]*TablePublic, error)
}

type Handler struct {
	Service service
}

type createTableRequest struct {
	Name string `json:"name"`
}

var createTableRequestSchema = z.Struct(z.Shape{
	"Name": NameSchema.Required(),
})

type createTableResponse struct {
	Table Table `json:"table"`
}

// CreateTableHandler handles requests to create a new table.
func (h *Handler) CreateTableHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := createTableRequest{}
		if !api.ReadAndValidateBody(w, r, &body, createTableRequestSchema) {
			return
		}

		ctx := r.Context()
		table, err := h.Service.CreateTable(ctx, body.Name)
		if err != nil {
			api.SendInternalServerError(w)
			return
		}

		api.SendResponse(w, createTableResponse{
			Table: *table,
		})
	}
}

type updateTableRequest struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var updateTableRequestSchema = z.Struct(z.Shape{
	"ID":   IDSchema.Required(),
	"Name": NameSchema.Required(),
})

type updateTableResponse struct {
	Table Table `json:"table"`
}

// UpdateTableHandler handles requests to update an existing table.
func (h *Handler) UpdateTableHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := updateTableRequest{}
		if !api.ReadAndValidateBody(w, r, &body, updateTableRequestSchema) {
			return
		}

		ctx := r.Context()
		table, err := h.Service.UpdateTable(ctx, body.ID, body.Name)
		if err != nil {
			if errors.Is(err, ErrTableNotFound) {
				api.SendNotFoundError(w, api.ErrorResponse{
					Message: "Table not found",
					Code:    "table_not_found",
				})
				return
			}
			api.SendInternalServerError(w)
			return
		}

		api.SendResponse(w, updateTableResponse{
			Table: *table,
		})
	}
}

type getAllTablesResponse struct {
	Tables []*Table `json:"tables"`
}

// GetAllTablesHandler handles requests to retrieve all tables.
func (h *Handler) GetAllTablesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tables, err := h.Service.GetAllTables(ctx)
		if err != nil {
			api.SendInternalServerError(w)
			return
		}

		if tables == nil {
			tables = []*Table{}
		}

		api.SendResponse(w, getAllTablesResponse{
			Tables: tables,
		})
	}
}

type getActiveTablesResponse struct {
	Tables []*TablePublic `json:"tables"`
}

// GetActiveTablesHandler handles requests to retrieve all active tables.
func (h *Handler) GetActiveTablesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tables, err := h.Service.GetActiveTables(ctx)
		if err != nil {
			api.SendInternalServerError(w)
			return
		}

		if tables == nil {
			tables = []*TablePublic{}
		}

		api.SendResponse(w, getActiveTablesResponse{
			Tables: tables,
		})
	}
}

type getTableRequest struct {
	ID int `json:"id"`
}

var getTableRequestSchema = z.Struct(z.Shape{
	"ID": IDSchema.Required(),
})

type getTableResponse struct {
	Table Table `json:"table"`
}

// GetTableHandler handles requests to retrieve a table by its ID.
func (h *Handler) GetTableHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := getTableRequest{}
		if !api.ReadAndValidateBody(w, r, &body, getTableRequestSchema) {
			return
		}

		ctx := r.Context()
		table, err := h.Service.GetTable(ctx, body.ID)
		if err != nil {
			if errors.Is(err, ErrTableNotFound) {
				api.SendNotFoundError(w, api.ErrorResponse{
					Message: "Table not found",
					Code:    "table_not_found",
				})
				return
			}
			api.SendInternalServerError(w)
			return
		}

		api.SendResponse(w, getTableResponse{
			Table: *table,
		})
	}
}

type activateTableRequest struct {
	ID int `json:"id"`
}

var activateTableRequestSchema = z.Struct(z.Shape{
	"ID": IDSchema.Required(),
})

// ActivateTableHandler handles requests to activate a table.
func (h *Handler) ActivateTableHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := activateTableRequest{}
		if !api.ReadAndValidateBody(w, r, &body, activateTableRequestSchema) {
			return
		}

		ctx := r.Context()
		err := h.Service.ActivateTable(ctx, body.ID)
		if err != nil {
			if errors.Is(err, ErrTableNotFound) {
				api.SendNotFoundError(w, api.ErrorResponse{
					Message: "Table not found",
					Code:    "table_not_found",
				})
				return
			}
			api.SendInternalServerError(w)
			return
		}

		api.SendEmptyResponse(w)
	}
}

type deactivateTableRequest struct {
	ID int `json:"id"`
}

var deactivateTableRequestSchema = z.Struct(z.Shape{
	"ID": IDSchema.Required(),
})

// DeactivateTableHandler handles requests to deactivate a table.
func (h *Handler) DeactivateTableHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := deactivateTableRequest{}
		if !api.ReadAndValidateBody(w, r, &body, deactivateTableRequestSchema) {
			return
		}

		ctx := r.Context()
		err := h.Service.DeactivateTable(ctx, body.ID)
		if err != nil {
			if errors.Is(err, ErrTableNotFound) {
				api.SendNotFoundError(w, api.ErrorResponse{
					Message: "Table not found",
					Code:    "table_not_found",
				})
				return
			}
			api.SendInternalServerError(w)
			return
		}

		api.SendEmptyResponse(w)
	}
}
