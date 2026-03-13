package http

import (
	"context"
	"net/http"

	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/product/application"
	"github.com/nicograef/jotti/backend/domain/product"
)

type productReader interface {
	GetAllProducts(ctx context.Context) ([]product.Produkt, error)
	GetActiveProducts(ctx context.Context) ([]product.Produkt, error)
}

type QueryHandler struct {
	ProductRepo productReader
}

type getAllProductsResponse struct {
	Produkte []product.Produkt `json:"produkte"`
}

func (h *QueryHandler) GetAllProductsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		products, err := application.GetAllProducts(ctx, h.ProductRepo)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, getAllProductsResponse{Produkte: products})
	}
}

type getActiveProductsResponse struct {
	Produkte []product.Produkt `json:"produkte"`
}

func (h *QueryHandler) GetActiveProductsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		products, err := application.GetActiveProducts(ctx, h.ProductRepo)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, getActiveProductsResponse{Produkte: products})
	}
}
