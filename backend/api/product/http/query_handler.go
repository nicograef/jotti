package http

import (
	"context"
	"net/http"

	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/domain/product"
)

type query interface {
	GetAllProducts(ctx context.Context) ([]product.Product, error)
	GetActiveProducts(ctx context.Context) ([]product.Product, error)
}

type QueryHandler struct {
	Query query
}

type getAllProductsResponse struct {
	Products []product.Product `json:"products"`
}

func (h *QueryHandler) GetAllProductsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		products, err := h.Query.GetAllProducts(ctx)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, getAllProductsResponse{Products: products})
	}
}

type activeProduct struct {
	ID            int              `json:"id"`
	Name          string           `json:"name"`
	Description   string           `json:"description"`
	NetPriceCents int              `json:"netPriceCents"`
	Category      product.Category `json:"category"`
}

type getActiveProductsResponse struct {
	Products []activeProduct `json:"products"`
}

func (h *QueryHandler) GetActiveProductsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		products, err := h.Query.GetActiveProducts(ctx)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		convertedProducts := make([]activeProduct, len(products))
		for i, p := range products {
			convertedProducts[i] = activeProduct{
				ID:            p.ID,
				Name:          p.Name,
				Description:   p.Description,
				NetPriceCents: p.NetPriceCents,
				Category:      p.Category,
			}
		}

		helper.SendResponse(w, getActiveProductsResponse{Products: convertedProducts})
	}
}
