package table

import (
	"context"
	"errors"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api"
)

type query interface {
	GetTable(ctx context.Context, id int) (*Table, error)
	GetAllTables(ctx context.Context) ([]Table, error)
	GetActiveTables(ctx context.Context) ([]TablePublic, error)
}

type QueryHandler struct {
	Query query
}

type getAllTablesResponse struct {
	Tables []Table `json:"tables"`
}

// GetAllTablesHandler handles requests to retrieve all tables.
func (h *QueryHandler) GetAllTablesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tables, err := h.Query.GetAllTables(ctx)
		if err != nil {
			api.SendServerError(w)
			return
		}

		api.SendResponse(w, getAllTablesResponse{Tables: tables})
	}
}

type getActiveTablesResponse struct {
	Tables []TablePublic `json:"tables"`
}

// GetActiveTablesHandler handles requests to retrieve all active tables.
func (h *QueryHandler) GetActiveTablesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tables, err := h.Query.GetActiveTables(ctx)
		if err != nil {
			api.SendServerError(w)
			return
		}

		api.SendResponse(w, getActiveTablesResponse{Tables: tables})
	}
}

type getTable struct {
	ID int `json:"id"`
}

var getTableSchema = z.Struct(z.Shape{
	"ID": IDSchema.Required(),
})

type getTableResponse struct {
	Table Table `json:"table"`
}

// GetTableHandler handles requests to retrieve a table by its ID.
func (h *QueryHandler) GetTableHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := getTable{}
		if !api.ReadAndValidateBody(w, r, &body, getTableSchema) {
			return
		}

		ctx := r.Context()
		table, err := h.Query.GetTable(ctx, body.ID)
		if err != nil {
			if errors.Is(err, ErrTableNotFound) {
				api.SendClientError(w, "table_not_found", nil)
				return
			} else {
				api.SendServerError(w)
				return
			}
		}

		api.SendResponse(w, getTableResponse{Table: *table})
	}
}
