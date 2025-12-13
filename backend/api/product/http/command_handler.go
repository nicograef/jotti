package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/product/application"
	"github.com/nicograef/jotti/backend/domain/product"
)

type command interface {
	CreateProduct(ctx context.Context, name, description string, netPriceCents int, category product.Category) (int, error)
	UpdateProduct(ctx context.Context, id int, name, description string, netPriceCents int, category product.Category) error
	ActivateProduct(ctx context.Context, id int) error
	DeactivateProduct(ctx context.Context, id int) error
}

type CommandHandler struct {
	Command command
}

type createProduct struct {
	Name          string           `json:"name"`
	Description   string           `json:"description"`
	NetPriceCents int              `json:"netPriceCents"`
	Category      product.Category `json:"category"`
}

type createProductResponse struct {
	ID int `json:"id"`
}

func (h *CommandHandler) CreateProductHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := createProduct{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		id, err := h.Command.CreateProduct(r.Context(), body.Name, body.Description, body.NetPriceCents, body.Category)
		if err != nil {
			if errors.Is(err, application.ErrProductAlreadyExists) {
				helper.SendClientError(w, "product_already_exists", nil)
				return
			} else if errors.Is(err, application.ErrInvalidProductData) {
				helper.SendClientError(w, "invalid_product_data", nil)
				return
			} else {
				helper.SendServerError(w)
				return
			}
		}

		helper.SendResponse(w, createProductResponse{ID: id})
	}
}

type updateProduct struct {
	ID            int              `json:"id"`
	Name          string           `json:"name"`
	Description   string           `json:"description"`
	NetPriceCents int              `json:"netPriceCents"`
	Category      product.Category `json:"category"`
}

func (h *CommandHandler) UpdateProductHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := updateProduct{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		err := h.Command.UpdateProduct(r.Context(), body.ID, body.Name, body.Description, body.NetPriceCents, body.Category)
		if err != nil {
			if errors.Is(err, application.ErrProductNotFound) {
				helper.SendClientError(w, "product_not_found", nil)
				return
			} else if errors.Is(err, application.ErrInvalidProductData) {
				helper.SendClientError(w, "invalid_product_data", nil)
				return
			} else {
				helper.SendServerError(w)
				return
			}
		}

		helper.SendEmptyResponse(w)
	}
}

type activateProduct struct {
	ID int `json:"id"`
}

func (h *CommandHandler) ActivateProductHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := activateProduct{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		err := h.Command.ActivateProduct(r.Context(), body.ID)
		if err != nil {
			if errors.Is(err, application.ErrProductNotFound) {
				helper.SendClientError(w, "product_not_found", nil)
				return
			} else {
				helper.SendServerError(w)
				return
			}
		}

		helper.SendEmptyResponse(w)
	}
}

type deactivateTable struct {
	ID int `json:"id"`
}

func (h *CommandHandler) DeactivateProductHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := deactivateTable{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		err := h.Command.DeactivateProduct(r.Context(), body.ID)
		if err != nil {
			if errors.Is(err, application.ErrProductNotFound) {
				helper.SendClientError(w, "product_not_found", nil)
				return
			} else {
				helper.SendServerError(w)
				return
			}
		}

		helper.SendEmptyResponse(w)
	}
}
