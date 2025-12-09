package product_service

import (
	"context"
	"net/http"

	"github.com/nicograef/jotti/backend/api"
)

type query interface {
	GetAllProducts(ctx context.Context) ([]Product, error)
}

type QueryHandler struct {
	Query query
}

type getAllProductsResponse struct {
	Products []Product `json:"products"`
}

// GetAllProductsHandler handles requests to retrieve all products.
func (h *QueryHandler) GetAllProductsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		products, err := h.Query.GetAllProducts(ctx)
		if err != nil {
			api.SendServerError(w)
			return
		}

		api.SendResponse(w, getAllProductsResponse{Products: products})
	}
}
