package order

import (
	"context"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api"
	"github.com/nicograef/jotti/backend/table"
	"github.com/nicograef/jotti/backend/user"
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

var placeOrderSchema = z.Struct(z.Shape{
	"TableID":  table.IDSchema.Required(),
	"Products": z.Slice(orderProductSchema).Min(1).Required(),
})

// PlaceOrderHandler handles requests to place a new order.
func (h *CommandHandler) PlaceOrderHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := placeOrder{}
		if !api.ReadAndValidateBody(w, r, &body, placeOrderSchema) {
			return
		}

		ctx := r.Context()
		userID := ctx.Value(user.UserIDKey).(int)
		err := h.Command.PlaceOrder(ctx, userID, body.TableID, body.Products)
		if err != nil {
			api.SendServerError(w)
			return
		}

		api.SendEmptyResponse(w)
	}
}
