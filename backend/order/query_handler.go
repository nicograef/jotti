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
