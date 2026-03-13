package http

import (
	"context"
	"errors"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/product/application"
	"github.com/nicograef/jotti/backend/domain/product"
)

type command interface {
	CreateProduct(ctx context.Context, name string, kategorie product.Kategorie) (int, error)
	UpdateProduct(ctx context.Context, id int, name string, kategorie product.Kategorie) error
	ActivateProduct(ctx context.Context, productID int) error
	DeactivateProduct(ctx context.Context, productID int) error
	DeleteProdukt(ctx context.Context, productID int) error
	CreateVariant(ctx context.Context, productID int, name string, preisCents int) (int, error)
	UpdateVariant(ctx context.Context, variantID int, name string, preisCents int) error
	ActivateVariant(ctx context.Context, variantID int) error
	DeactivateVariant(ctx context.Context, variantID int) error
	DeleteVariante(ctx context.Context, produktID int, variantID int) error
}

type CommandHandler struct {
	Command command
}

// Product handlers

type createProduct struct {
	Name      string            `json:"name"`
	Kategorie product.Kategorie `json:"kategorie"`
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

		id, err := h.Command.CreateProduct(r.Context(), body.Name, body.Kategorie)
		if err != nil {
			switch {
			case errors.Is(err, application.ErrProduktAlreadyExists):
				helper.SendClientError(w, "produkt_already_exists", nil)
			case errors.Is(err, application.ErrInvalidProduktData):
				helper.SendClientError(w, "invalid_produkt_data", nil)
			default:
				helper.SendServerError(w)
			}
			return
		}

		helper.SendResponse(w, createProductResponse{ID: id})
	}
}

type updateProduct struct {
	ID        int               `json:"id"`
	Name      string            `json:"name"`
	Kategorie product.Kategorie `json:"kategorie"`
}

func (h *CommandHandler) UpdateProductHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := updateProduct{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		err := h.Command.UpdateProduct(r.Context(), body.ID, body.Name, body.Kategorie)
		if err != nil {
			switch {
			case errors.Is(err, application.ErrProduktNotFound):
				helper.SendClientError(w, "produkt_not_found", nil)
			case errors.Is(err, application.ErrInvalidProduktData):
				helper.SendClientError(w, "invalid_produkt_data", nil)
			default:
				helper.SendServerError(w)
			}
			return
		}

		helper.SendEmptyResponse(w)
	}
}

type activateProduct struct {
	ID int `json:"id"`
}

var activateProductSchema = z.Struct(z.Shape{
	"ID": product.IDSchema.Required(),
})

func (h *CommandHandler) ActivateProductHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := activateProduct{}
		if !helper.ReadAndValidateBody(w, r, &body, activateProductSchema) {
			return
		}

		err := h.Command.ActivateProduct(r.Context(), body.ID)
		if err != nil {
			if errors.Is(err, application.ErrProduktNotFound) {
				helper.SendClientError(w, "produkt_not_found", nil)
				return
			} else {
				helper.SendServerError(w)
				return
			}
		}

		helper.SendEmptyResponse(w)
	}
}

type deactivateProduct struct {
	ID int `json:"id"`
}

var deactivateProductSchema = z.Struct(z.Shape{
	"ID": product.IDSchema.Required(),
})

func (h *CommandHandler) DeactivateProductHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := deactivateProduct{}
		if !helper.ReadAndValidateBody(w, r, &body, deactivateProductSchema) {
			return
		}

		err := h.Command.DeactivateProduct(r.Context(), body.ID)
		if err != nil {
			if errors.Is(err, application.ErrProduktNotFound) {
				helper.SendClientError(w, "produkt_not_found", nil)
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
	ProductID  int    `json:"produktId"`
	Name       string `json:"name"`
	PreisCents int    `json:"preisCents"`
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

		id, err := h.Command.CreateVariant(r.Context(), body.ProductID, body.Name, body.PreisCents)
		if err != nil {
			switch {
			case errors.Is(err, application.ErrProduktNotFound):
				helper.SendClientError(w, "produkt_not_found", nil)
			case errors.Is(err, application.ErrInvalidVarianteData):
				helper.SendClientError(w, "invalid_variante_data", nil)
			default:
				helper.SendServerError(w)
			}
			return
		}

		helper.SendResponse(w, createVariantResponse{ID: id})
	}
}

type updateVariant struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	PreisCents int    `json:"preisCents"`
}

func (h *CommandHandler) UpdateVariantHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := updateVariant{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		err := h.Command.UpdateVariant(r.Context(), body.ID, body.Name, body.PreisCents)
		if err != nil {
			switch {
			case errors.Is(err, application.ErrVarianteNotFound):
				helper.SendClientError(w, "variante_not_found", nil)
			case errors.Is(err, application.ErrInvalidVarianteData):
				helper.SendClientError(w, "invalid_variante_data", nil)
			default:
				helper.SendServerError(w)
			}
			return
		}

		helper.SendEmptyResponse(w)
	}
}

type activateVariant struct {
	ID int `json:"id"`
}

var activateVariantSchema = z.Struct(z.Shape{
	"ID": product.IDSchema.Required(),
})

func (h *CommandHandler) ActivateVariantHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := activateVariant{}
		if !helper.ReadAndValidateBody(w, r, &body, activateVariantSchema) {
			return
		}

		err := h.Command.ActivateVariant(r.Context(), body.ID)
		if err != nil {
			if errors.Is(err, application.ErrVarianteNotFound) {
				helper.SendClientError(w, "variante_not_found", nil)
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

var deactivateVariantSchema = z.Struct(z.Shape{
	"ID": product.IDSchema.Required(),
})

func (h *CommandHandler) DeactivateVariantHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := deactivateVariant{}
		if !helper.ReadAndValidateBody(w, r, &body, deactivateVariantSchema) {
			return
		}

		err := h.Command.DeactivateVariant(r.Context(), body.ID)
		if err != nil {
			if errors.Is(err, application.ErrVarianteNotFound) {
				helper.SendClientError(w, "variante_not_found", nil)
				return
			} else {
				helper.SendServerError(w)
				return
			}
		}

		helper.SendEmptyResponse(w)
	}
}

type deleteProdukt struct {
	ID int `json:"id"`
}

var deleteProduktSchema = z.Struct(z.Shape{
	"ID": product.IDSchema.Required(),
})

func (h *CommandHandler) DeleteProduktHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := deleteProdukt{}
		if !helper.ReadAndValidateBody(w, r, &body, deleteProduktSchema) {
			return
		}

		err := h.Command.DeleteProdukt(r.Context(), body.ID)
		if err != nil {
			if errors.Is(err, application.ErrProduktNotFound) {
				helper.SendClientError(w, "produkt_not_found", nil)
				return
			} else {
				helper.SendServerError(w)
				return
			}
		}

		helper.SendEmptyResponse(w)
	}
}

type deleteVariante struct {
	ProduktID int `json:"produktId"`
	ID        int `json:"id"`
}

var deleteVarianteSchema = z.Struct(z.Shape{
	"ProduktID": product.IDSchema.Required(),
	"ID":        product.IDSchema.Required(),
})

func (h *CommandHandler) DeleteVarianteHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := deleteVariante{}
		if !helper.ReadAndValidateBody(w, r, &body, deleteVarianteSchema) {
			return
		}

		err := h.Command.DeleteVariante(r.Context(), body.ProduktID, body.ID)
		if err != nil {
			switch {
			case errors.Is(err, application.ErrProduktNotFound):
				helper.SendClientError(w, "produkt_not_found", nil)
			case errors.Is(err, application.ErrVarianteNotFound):
				helper.SendClientError(w, "variante_not_found", nil)
			default:
				helper.SendServerError(w)
			}
			return
		}

		helper.SendEmptyResponse(w)
	}
}
