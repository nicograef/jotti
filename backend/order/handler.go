package order

import (
	"context"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api"
	"github.com/nicograef/jotti/backend/auth"
	"github.com/nicograef/jotti/backend/table"
)

type service interface {
	PlaceOrder(ctx context.Context, userID int, tableID int, products []OrderProduct) (*Order, error)
}

type Handler struct {
	Service service
}

type placeOrderRequest struct {
	TableID  int            `json:"tableId"`
	Products []OrderProduct `json:"products"`
}

var placeOrderRequestSchema = z.Struct(z.Shape{
	"TableID":  table.IDSchema.Required(),
	"Products": z.Slice(OrderProductSchema).Min(1).Required(),
})

type placeOrderResponse struct {
	Order Order `json:"order"`
}

// PlaceOrderHandler handles requests to place a new order.
func (h *Handler) PlaceOrderHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !api.ValidateMethod(w, r, http.MethodPost) {
			return
		}

		body := placeOrderRequest{}
		if !api.ReadJSONRequest(w, r, &body) {
			return
		}

		if !api.ValidateBody(w, &body, placeOrderRequestSchema) {
			return
		}

		ctx := r.Context()
		userID := ctx.Value(auth.UserIDKey).(int)
		order, err := h.Service.PlaceOrder(ctx, userID, body.TableID, body.Products)
		if err != nil {
			api.SendInternalServerError(w)
			return
		}

		api.SendResponse(w, placeOrderResponse{
			Order: *order,
		})
	}
}
