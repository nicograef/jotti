package product

import (
	"context"
	"net/http"

	"github.com/nicograef/jotti/backend/api"
)

type query interface {
	GetAllProducts(ctx context.Context) ([]Product, error)
	GetActiveProducts(ctx context.Context) ([]ProductPublic, error)
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

type getActiveProductsResponse struct {
	Products []ProductPublic `json:"products"`
}

// GetActiveProductsHandler handles requests to retrieve all active products.
func (h *QueryHandler) GetActiveProductsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		products, err := h.Query.GetActiveProducts(ctx)
		if err != nil {
			api.SendServerError(w)
			return
		}

		api.SendResponse(w, getActiveProductsResponse{Products: products})
	}
}
