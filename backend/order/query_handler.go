package order

import (
	"context"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api"
	"github.com/nicograef/jotti/backend/table"
)

type query interface {
	GetOrders(ctx context.Context, tableID int) ([]Order, error)
	GetTableBalance(ctx context.Context, tableID int) (int, error)
	GetTableUnpaidProducts(ctx context.Context, tableID int) ([]orderProduct, error)
}

type QueryHandler struct {
	Query query
}

type getOrders struct {
	TableID int `json:"tableId"`
}

var getOrdersSchema = z.Struct(z.Shape{
	"TableID": table.IDSchema.Required(),
})

type getOrdersResponse struct {
	Orders []Order `json:"orders"`
}

// GetOrdersHandler handles requests to retrieve orders for a specific table.
func (h *QueryHandler) GetOrdersHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := getOrders{}
		if !api.ReadAndValidateBody(w, r, &body, getOrdersSchema) {
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

var getTableBalanceSchema = z.Struct(z.Shape{
	"TableID": table.IDSchema.Required(),
})

type getTableBalanceResponse struct {
	TotalBalanceCents int `json:"totalBalanceCents"`
}

// GetTableBalanceHandler handles requests to retrieve the total balance for a specific table.
func (h *QueryHandler) GetTableBalanceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := getTableBalance{}
		if !api.ReadAndValidateBody(w, r, &body, getTableBalanceSchema) {
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

var getTableUnpaidProductsSchema = z.Struct(z.Shape{
	"TableID": table.IDSchema.Required(),
})

type getTableUnpaidProductsResponse struct {
	Products []orderProduct `json:"products"`
}

// GetTableUnpaidProductsHandler handles requests to retrieve unpaid products for a specific table.
func (h *QueryHandler) GetTableUnpaidProductsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := getTableUnpaidProducts{}
		if !api.ReadAndValidateBody(w, r, &body, getTableUnpaidProductsSchema) {
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
