package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/nicograef/jotti/backend/admin/product/application"
	"github.com/nicograef/jotti/backend/admin/product/domain"
	"github.com/nicograef/jotti/backend/api"
)

type command interface {
	CreateProduct(ctx context.Context, name, description string, netPriceCents int, category domain.Category) (int, error)
	UpdateProduct(ctx context.Context, id int, name, description string, netPriceCents int, category domain.Category) error
	ActivateProduct(ctx context.Context, id int) error
	DeactivateProduct(ctx context.Context, id int) error
}

type CommandHandler struct {
	Command command
}

type createProduct struct {
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	NetPriceCents int             `json:"netPriceCents"`
	Category      domain.Category `json:"category"`
}

type createProductResponse struct {
	ID int `json:"id"`
}

// CreateProductHandler handles requests to create a new product.
func (h *CommandHandler) CreateProductHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := createProduct{}
		if !api.ReadBody(w, r, &body) {
			return
		}

		ctx := r.Context()
		id, err := h.Command.CreateProduct(ctx, body.Name, body.Description, body.NetPriceCents, body.Category)
		if err != nil {
			if errors.Is(err, application.ErrProductAlreadyExists) {
				api.SendClientError(w, "product_already_exists", nil)
				return
			} else if errors.Is(err, application.ErrInvalidProductData) {
				api.SendClientError(w, "invalid_product_data", nil)
				return
			} else {
				api.SendServerError(w)
				return
			}
		}

		api.SendResponse(w, createProductResponse{ID: id})
	}
}

type updateProduct struct {
	ID            int             `json:"id"`
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	NetPriceCents int             `json:"netPriceCents"`
	Category      domain.Category `json:"category"`
}

// UpdateProductHandler handles requests to update an existing product.
func (h *CommandHandler) UpdateProductHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := updateProduct{}
		if !api.ReadBody(w, r, &body) {
			return
		}

		ctx := r.Context()
		err := h.Command.UpdateProduct(ctx, body.ID, body.Name, body.Description, body.NetPriceCents, body.Category)
		if err != nil {
			if errors.Is(err, application.ErrProductNotFound) {
				api.SendClientError(w, "product_not_found", nil)
				return
			} else if errors.Is(err, application.ErrInvalidProductData) {
				api.SendClientError(w, "invalid_product_data", nil)
				return
			} else {
				api.SendServerError(w)
				return
			}
		}

		api.SendEmptyResponse(w)
	}
}

type activateProduct struct {
	ID int `json:"id"`
}

// ActivateProductHandler handles requests to activate a product.
func (h *CommandHandler) ActivateProductHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := activateProduct{}
		if !api.ReadBody(w, r, &body) {
			return
		}

		ctx := r.Context()
		err := h.Command.ActivateProduct(ctx, body.ID)
		if err != nil {
			if errors.Is(err, application.ErrProductNotFound) {
				api.SendClientError(w, "product_not_found", nil)
				return
			} else {
				api.SendServerError(w)
				return
			}
		}

		api.SendEmptyResponse(w)
	}
}

type deactivateTable struct {
	ID int `json:"id"`
}

// DeactivateProductHandler handles requests to deactivate a product.
func (h *CommandHandler) DeactivateProductHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := deactivateTable{}
		if !api.ReadBody(w, r, &body) {
			return
		}

		ctx := r.Context()
		err := h.Command.DeactivateProduct(ctx, body.ID)
		if err != nil {
			if errors.Is(err, application.ErrProductNotFound) {
				api.SendClientError(w, "product_not_found", nil)
				return
			} else {
				api.SendServerError(w)
				return
			}
		}

		api.SendEmptyResponse(w)
	}
}
