package api

import (
	"errors"
	"net/http"

	z "github.com/Oudwins/zog"

	tbl "github.com/nicograef/jotti/backend/domain/table"
)

type createTableRequest struct {
	Name string `json:"name"`
}

var createTableRequestSchema = z.Struct(z.Shape{
	"Name": tbl.NameSchema.Required(),
})

type createTableResponse struct {
	Table tbl.Table `json:"table"`
}

// CreateTableHandler handles requests to create a new table.
func CreateTableHandler(ts *tbl.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !validateMethod(w, r, http.MethodPost) {
			return
		}

		body := createTableRequest{}
		if !readJSONRequest(w, r, &body) {
			return
		}

		if !validateBody(w, &body, createTableRequestSchema) {
			return
		}

		table, err := ts.CreateTable(body.Name)
		if err != nil {
			sendInternalServerError(w)
			return
		}

		sendResponse(w, createTableResponse{
			Table: *table,
		})
	}
}

type updateTableRequest struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var updateTableRequestSchema = z.Struct(z.Shape{
	"ID":   tbl.IDSchema.Required(),
	"Name": tbl.NameSchema.Required(),
})

// UpdateTableHandler handles requests to update an existing table.
func UpdateTableHandler(ts *tbl.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !validateMethod(w, r, http.MethodPost) {
			return
		}

		body := updateTableRequest{}
		if !readJSONRequest(w, r, &body) {
			return
		}

		if !validateBody(w, &body, updateTableRequestSchema) {
			return
		}

		table, err := ts.UpdateTable(body.ID, body.Name)
		if err != nil {
			if errors.Is(err, tbl.ErrTableNotFound) {
				sendNotFoundError(w, errorResponse{
					Message: "Table not found",
					Code:    "table_not_found",
				})
				return
			}
			sendInternalServerError(w)
			return
		}

		sendResponse(w, createTableResponse{
			Table: *table,
		})
	}
}

type getTablesResponse struct {
	Tables []*tbl.Table `json:"tables"`
}

// GetTablesHandler handles requests to retrieve all tables.
func GetTablesHandler(ts *tbl.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !validateMethod(w, r, http.MethodPost) {
			return
		}

		tables, err := ts.GetAllTables()
		if err != nil {
			sendInternalServerError(w)
			return
		}

		if tables == nil {
			tables = []*tbl.Table{}
		}

		sendResponse(w, getTablesResponse{
			Tables: tables,
		})
	}
}
