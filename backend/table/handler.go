package table

import (
	"errors"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api"
)

type service interface {
	CreateTable(name string) (*Table, error)
	UpdateTable(id int, name string) (*Table, error)
	ActivateTable(id int) error
	DeactivateTable(id int) error
	GetAllTables() ([]*Table, error)
	GetActiveTables() ([]*TablePublic, error)
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
		if !api.ValidateMethod(w, r, http.MethodPost) {
			return
		}

		body := createTableRequest{}
		if !api.ReadJSONRequest(w, r, &body) {
			return
		}

		if !api.ValidateBody(w, &body, createTableRequestSchema) {
			return
		}

		table, err := h.Service.CreateTable(body.Name)
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
		if !api.ValidateMethod(w, r, http.MethodPost) {
			return
		}

		body := updateTableRequest{}
		if !api.ReadJSONRequest(w, r, &body) {
			return
		}

		if !api.ValidateBody(w, &body, updateTableRequestSchema) {
			return
		}

		table, err := h.Service.UpdateTable(body.ID, body.Name)
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
		if !api.ValidateMethod(w, r, http.MethodPost) {
			return
		}

		tables, err := h.Service.GetAllTables()
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
		if !api.ValidateMethod(w, r, http.MethodPost) {
			return
		}

		tables, err := h.Service.GetActiveTables()
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

type activateTableRequest struct {
	ID int `json:"id"`
}

var activateTableRequestSchema = z.Struct(z.Shape{
	"ID": IDSchema.Required(),
})

// ActivateTableHandler handles requests to activate a table.
func (h *Handler) ActivateTableHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !api.ValidateMethod(w, r, http.MethodPost) {
			return
		}

		body := activateTableRequest{}
		if !api.ReadJSONRequest(w, r, &body) {
			return
		}

		if !api.ValidateBody(w, &body, activateTableRequestSchema) {
			return
		}

		err := h.Service.ActivateTable(body.ID)
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
		if !api.ValidateMethod(w, r, http.MethodPost) {
			return
		}

		body := deactivateTableRequest{}
		if !api.ReadJSONRequest(w, r, &body) {
			return
		}

		if !api.ValidateBody(w, &body, deactivateTableRequestSchema) {
			return
		}

		err := h.Service.DeactivateTable(body.ID)
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
