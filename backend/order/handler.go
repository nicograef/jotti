package order

import (
	"context"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api"
	"github.com/nicograef/jotti/backend/event"
	"github.com/nicograef/jotti/backend/table"
	"github.com/nicograef/jotti/backend/user"
)

func NewHandler(p *event.Persistence) *handler {
	return &handler{&commandService{persistence: p}, &queryService{persistence: p}}
}

type command interface {
	PlaceOrder(ctx context.Context, userID int, tableID int, products []orderProduct) (*Order, error)
}

type query interface {
	GetOrders(ctx context.Context, tableID int) ([]Order, error)
}

type handler struct {
	command command
	query   query
}

type placeOrder struct {
	TableID  int            `json:"tableId"`
	Products []orderProduct `json:"products"`
}

var placeOrderSchema = z.Struct(z.Shape{
	"TableID":  table.IDSchema.Required(),
	"Products": z.Slice(orderProductSchema).Min(1).Required(),
})

type placeOrderResponse struct {
	Order Order `json:"order"`
}

// PlaceOrderHandler handles requests to place a new order.
func (h *handler) PlaceOrderHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := placeOrder{}
		if !api.ReadAndValidateBody(w, r, &body, placeOrderSchema) {
			return
		}

		ctx := r.Context()
		userID := ctx.Value(user.UserIDKey).(int)
		order, err := h.command.PlaceOrder(ctx, userID, body.TableID, body.Products)
		if err != nil {
			api.SendServerError(w)
			return
		}

		api.SendResponse(w, placeOrderResponse{
			Order: *order,
		})
	}
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
func (h *handler) GetOrdersHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := getOrders{}
		if !api.ReadAndValidateBody(w, r, &body, getOrdersSchema) {
			return
		}

		ctx := r.Context()
		orders, err := h.query.GetOrders(ctx, body.TableID)
		if err != nil {
			api.SendServerError(w)
			return
		}

		api.SendResponse(w, getOrdersResponse{
			Orders: orders,
		})
	}
}
