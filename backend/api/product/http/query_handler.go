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
	Produkte []product.Product `json:"produkte"`
}

func (h *QueryHandler) GetAllProductsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		products, err := h.Query.GetAllProducts(ctx)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, getAllProductsResponse{Produkte: products})
	}
}

type getActiveProductsResponse struct {
	Produkte []product.Product `json:"produkte"`
}

func (h *QueryHandler) GetActiveProductsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		products, err := h.Query.GetActiveProducts(ctx)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, getActiveProductsResponse{Produkte: products})
	}
}
