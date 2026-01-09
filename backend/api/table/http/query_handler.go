package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/table/application"
	t "github.com/nicograef/jotti/backend/domain/table"
)

type query interface {
	GetTable(ctx context.Context, id int) (t.Table, error)
	GetAllTables(ctx context.Context) ([]t.Table, error)
	GetActiveTables(ctx context.Context) ([]t.Table, error)
	GetTableHistory(ctx context.Context, tableID int) ([]any, error)
	GetTableBalance(ctx context.Context, tableID int) (int, error)
	GetTableUnpaidProducts(ctx context.Context, tableID int) ([]t.OrderProduct, error)
	GetTableUndeliveredProducts(ctx context.Context, tableID int) ([]t.OrderProduct, error)
}

type QueryHandler struct {
	Query query
}

type getTable struct {
	ID int `json:"id"`
}

type getTableResponse struct {
	Table t.Table `json:"table"`
}

func (h QueryHandler) GetTableHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := getTable{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		table, err := h.Query.GetTable(r.Context(), body.ID)
		if err != nil {
			if errors.Is(err, application.ErrTableNotFound) {
				helper.SendClientError(w, "table_not_found", nil)
				return
			} else {
				helper.SendServerError(w)
				return
			}
		}

		helper.SendResponse(w, getTableResponse{Table: table})
	}
}

type getAllTablesResponse struct {
	Tables []t.Table `json:"tables"`
}

func (h QueryHandler) GetAllTablesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tables, err := h.Query.GetAllTables(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, getAllTablesResponse{Tables: tables})
	}
}

type activeTable struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type getActiveTablesResponse struct {
	Tables []activeTable `json:"tables"`
}

func (h QueryHandler) GetActiveTablesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tables, err := h.Query.GetActiveTables(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}

		activeTables := make([]activeTable, len(tables))
		for i, table := range tables {
			activeTables[i] = activeTable{
				ID:   table.ID,
				Name: table.Name,
			}
		}

		helper.SendResponse(w, getActiveTablesResponse{Tables: activeTables})
	}
}

type getTableHistory struct {
	TableID int `json:"tableId"`
}

type getTableHistoryResponse struct {
	History []any `json:"history"`
}

func (h QueryHandler) GetTableHistoryHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := getTableHistory{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		history, err := h.Query.GetTableHistory(r.Context(), body.TableID)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, getTableHistoryResponse{History: history})
	}
}

type getTableBalance struct {
	TableID int `json:"tableId"`
}

type getTableBalanceResponse struct {
	BalanceCents int `json:"balanceCents"`
}

func (h QueryHandler) GetTableBalanceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := getTableBalance{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		balanceCents, err := h.Query.GetTableBalance(r.Context(), body.TableID)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, getTableBalanceResponse{BalanceCents: balanceCents})
	}
}

type getTableUnpaidProducts struct {
	TableID int `json:"tableId"`
}

type getTableUnpaidProductsResponse struct {
	Products []t.OrderProduct `json:"products"`
}

func (h QueryHandler) GetTableUnpaidProductsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := getTableUnpaidProducts{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		products, err := h.Query.GetTableUnpaidProducts(r.Context(), body.TableID)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, getTableUnpaidProductsResponse{Products: products})
	}
}

type getTableUndeliveredProducts struct {
	TableID int `json:"tableId"`
}

type getTableUndeliveredProductsResponse struct {
	Products []t.OrderProduct `json:"products"`
}

func (h QueryHandler) GetTableUndeliveredProductsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := getTableUndeliveredProducts{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		products, err := h.Query.GetTableUndeliveredProducts(r.Context(), body.TableID)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, getTableUndeliveredProductsResponse{Products: products})
	}
}
