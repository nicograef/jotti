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
	CreateProduct(ctx context.Context, name string, category product.Category) (int, error)
	UpdateProduct(ctx context.Context, id int, name string, category product.Category) error
	CreateVariant(ctx context.Context, productID int, name string, priceCents int) (int, error)
	UpdateVariant(ctx context.Context, variantID int, name string, priceCents int) error
	ActivateVariant(ctx context.Context, variantID int) error
	DeactivateVariant(ctx context.Context, variantID int) error
}

type CommandHandler struct {
	Command command
}

// Product handlers

type createProduct struct {
	Name     string           `json:"name"`
	Category product.Category `json:"category"`
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

		id, err := h.Command.CreateProduct(r.Context(), body.Name, body.Category)
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
	ID       int              `json:"id"`
	Name     string           `json:"name"`
	Category product.Category `json:"category"`
}

func (h *CommandHandler) UpdateProductHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := updateProduct{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		err := h.Command.UpdateProduct(r.Context(), body.ID, body.Name, body.Category)
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

// Variant handlers

type createVariant struct {
	ProductID  int    `json:"productId"`
	Name       string `json:"name"`
	PriceCents int    `json:"priceCents"`
}

type createVariantResponse struct {
	ID int `json:"id"`
}

func (h *CommandHandler) CreateVariantHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := createVariant{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		id, err := h.Command.CreateVariant(r.Context(), body.ProductID, body.Name, body.PriceCents)
		if err != nil {
			if errors.Is(err, application.ErrProductNotFound) {
				helper.SendClientError(w, "product_not_found", nil)
				return
			} else if errors.Is(err, application.ErrInvalidVariantData) {
				helper.SendClientError(w, "invalid_variant_data", nil)
				return
			} else {
				helper.SendServerError(w)
				return
			}
		}

		helper.SendResponse(w, createVariantResponse{ID: id})
	}
}

type updateVariant struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	PriceCents int    `json:"priceCents"`
}

func (h *CommandHandler) UpdateVariantHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := updateVariant{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		err := h.Command.UpdateVariant(r.Context(), body.ID, body.Name, body.PriceCents)
		if err != nil {
			if errors.Is(err, application.ErrVariantNotFound) {
				helper.SendClientError(w, "variant_not_found", nil)
				return
			} else if errors.Is(err, application.ErrInvalidVariantData) {
				helper.SendClientError(w, "invalid_variant_data", nil)
				return
			} else {
				helper.SendServerError(w)
				return
			}
		}

		helper.SendEmptyResponse(w)
	}
}

type activateVariant struct {
	ID int `json:"id"`
}

func (h *CommandHandler) ActivateVariantHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := activateVariant{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		err := h.Command.ActivateVariant(r.Context(), body.ID)
		if err != nil {
			if errors.Is(err, application.ErrVariantNotFound) {
				helper.SendClientError(w, "variant_not_found", nil)
				return
			} else {
				helper.SendServerError(w)
				return
			}
		}

		helper.SendEmptyResponse(w)
	}
}

type deactivateVariant struct {
	ID int `json:"id"`
}

func (h *CommandHandler) DeactivateVariantHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := deactivateVariant{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		err := h.Command.DeactivateVariant(r.Context(), body.ID)
		if err != nil {
			if errors.Is(err, application.ErrVariantNotFound) {
				helper.SendClientError(w, "variant_not_found", nil)
				return
			} else {
				helper.SendServerError(w)
				return
			}
		}

		helper.SendEmptyResponse(w)
	}
}
