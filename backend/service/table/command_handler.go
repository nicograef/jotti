package table

import (
	"context"
	"net/http"

	"github.com/nicograef/jotti/backend/api"
)

type command interface {
	PlaceOrder(ctx context.Context, userID int, tableID int, products []orderProduct) error
}

type CommandHandler struct {
	Command command
}

type placeOrder struct {
	TableID  int            `json:"tableId"`
	Products []orderProduct `json:"products"`
}

// PlaceOrderHandler handles requests to place a new order.
func (h *CommandHandler) PlaceOrderHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := placeOrder{}
		if !api.ReadBody(w, r, &body) {
			return
		}

		ctx := r.Context()
		userID := ctx.Value(api.UserIDKey).(int)
		err := h.Command.PlaceOrder(ctx, userID, body.TableID, body.Products)
		if err != nil {
			api.SendServerError(w)
			return
		}

		api.SendEmptyResponse(w)
	}
}
