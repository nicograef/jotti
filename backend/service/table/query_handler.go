package table

import (
	"context"
	"errors"
	"net/http"

	"github.com/nicograef/jotti/backend/api"
)

type query interface {
	GetTable(ctx context.Context, id int) (*Table, error)
	GetAllTables(ctx context.Context) ([]Table, error)
	GetOrders(ctx context.Context, tableID int) ([]Order, error)
	GetTableBalance(ctx context.Context, tableID int) (int, error)
	GetTableUnpaidProducts(ctx context.Context, tableID int) ([]orderProduct, error)
}

type QueryHandler struct {
	Query query
}

type getTable struct {
	ID int `json:"id"`
}

type getTableResponse struct {
	Table Table `json:"table"`
}

// GetTableHandler handles requests to retrieve a table by its ID.
func (h *QueryHandler) GetTableHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := getTable{}
		if !api.ReadBody(w, r, &body) {
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

type getOrders struct {
	TableID int `json:"tableId"`
}

type getOrdersResponse struct {
	Orders []Order `json:"orders"`
}

// GetOrdersHandler handles requests to retrieve orders for a specific table.
func (h *QueryHandler) GetOrdersHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := getOrders{}
		if !api.ReadBody(w, r, &body) {
			return
		}

		ctx := r.Context()
		orders, err := h.Query.GetOrders(ctx, body.TableID)
		if err != nil {
			api.SendServerError(w)
			return
		}

		api.SendResponse(w, getOrdersResponse{Orders: orders})
	}
}

type getTableBalance struct {
	TableID int `json:"tableId"`
}

type getTableBalanceResponse struct {
	TotalBalanceCents int `json:"totalBalanceCents"`
}

// GetTableBalanceHandler handles requests to retrieve the total balance for a specific table.
func (h *QueryHandler) GetTableBalanceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := getTableBalance{}
		if !api.ReadBody(w, r, &body) {
			return
		}

		ctx := r.Context()
		totalBalanceCents, err := h.Query.GetTableBalance(ctx, body.TableID)
		if err != nil {
			api.SendServerError(w)
			return
		}

		api.SendResponse(w, getTableBalanceResponse{TotalBalanceCents: totalBalanceCents})
	}
}

type getTableUnpaidProducts struct {
	TableID int `json:"tableId"`
}

type getTableUnpaidProductsResponse struct {
	Products []orderProduct `json:"products"`
}

// GetTableUnpaidProductsHandler handles requests to retrieve unpaid products for a specific table.
func (h *QueryHandler) GetTableUnpaidProductsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := getTableUnpaidProducts{}
		if !api.ReadBody(w, r, &body) {
			return
		}

		ctx := r.Context()
		products, err := h.Query.GetTableUnpaidProducts(ctx, body.TableID)
		if err != nil {
			api.SendServerError(w)
			return
		}

		api.SendResponse(w, getTableUnpaidProductsResponse{Products: products})
	}
}
