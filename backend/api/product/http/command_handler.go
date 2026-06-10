package http

import (
	"context"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/product/application"
	"github.com/nicograef/jotti/backend/domain/product"
	"github.com/nicograef/jotti/backend/domain/steuer"
)

type command interface {
	CreateProduct(ctx context.Context, name string, kategorie product.Kategorie, steuersatz steuer.Steuersatz) (int, error)
	UpdateProduct(ctx context.Context, id int, name string, kategorie product.Kategorie, steuersatz steuer.Steuersatz) error
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

type createProductRequest struct {
	Name       string            `json:"name"`
	Kategorie  product.Kategorie `json:"kategorie"`
	Steuersatz steuer.Steuersatz `json:"steuersatz"`
}

var createProductSchema = z.Struct(z.Shape{
	"Name":       product.NameSchema.Required(),
	"Kategorie":  product.KategorieSchema.Required(),
	"Steuersatz": steuer.SteuersatzSchema.Required(),
})

type createProductResponse struct {
	ID int `json:"id"`
}

func (h *CommandHandler) CreateProductHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := createProductRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, createProductSchema) {
			return
		}

		id, err := h.Command.CreateProduct(r.Context(), body.Name, body.Kategorie, body.Steuersatz)
		if err != nil {
			helper.MapError(w, err, map[error]string{
				application.ErrProduktAlreadyExists: "produkt_already_exists",
				application.ErrInvalidProduktData:   "invalid_produkt_data",
			})
			return
		}

		helper.SendResponse(w, createProductResponse{ID: id})
	}
}

type updateProductRequest struct {
	ID         int               `json:"id"`
	Name       string            `json:"name"`
	Kategorie  product.Kategorie `json:"kategorie"`
	Steuersatz steuer.Steuersatz `json:"steuersatz"`
}

var updateProductSchema = z.Struct(z.Shape{
	"ID":         product.IDSchema.Required(),
	"Name":       product.NameSchema.Required(),
	"Kategorie":  product.KategorieSchema.Required(),
	"Steuersatz": steuer.SteuersatzSchema.Required(),
})

func (h *CommandHandler) UpdateProductHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := updateProductRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, updateProductSchema) {
			return
		}

		err := h.Command.UpdateProduct(r.Context(), body.ID, body.Name, body.Kategorie, body.Steuersatz)
		if err != nil {
			helper.MapError(w, err, map[error]string{
				application.ErrProduktNotFound:    "produkt_not_found",
				application.ErrInvalidProduktData: "invalid_produkt_data",
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}

// Variant handlers

type createVariantRequest struct {
	ProductID  int    `json:"produktId"`
	Name       string `json:"name"`
	PreisCents int    `json:"preisCents"`
}

var createVariantSchema = z.Struct(z.Shape{
	"ProductID":  product.IDSchema.Required(),
	"Name":       product.NameSchema.Required(),
	"PreisCents": product.PreisCentsSchema.Required(),
})

type createVariantResponse struct {
	ID int `json:"id"`
}

func (h *CommandHandler) CreateVariantHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := createVariantRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, createVariantSchema) {
			return
		}

		id, err := h.Command.CreateVariant(r.Context(), body.ProductID, body.Name, body.PreisCents)
		if err != nil {
			helper.MapError(w, err, map[error]string{
				application.ErrProduktNotFound:     "produkt_not_found",
				application.ErrInvalidVarianteData: "invalid_variante_data",
			})
			return
		}

		helper.SendResponse(w, createVariantResponse{ID: id})
	}
}

type updateVariantRequest struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	PreisCents int    `json:"preisCents"`
}

var updateVariantSchema = z.Struct(z.Shape{
	"ID":         product.IDSchema.Required(),
	"Name":       product.NameSchema.Required(),
	"PreisCents": product.PreisCentsSchema.Required(),
})

func (h *CommandHandler) UpdateVariantHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := updateVariantRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, updateVariantSchema) {
			return
		}

		err := h.Command.UpdateVariant(r.Context(), body.ID, body.Name, body.PreisCents)
		if err != nil {
			helper.MapError(w, err, map[error]string{
				application.ErrVarianteNotFound:    "variante_not_found",
				application.ErrInvalidVarianteData: "invalid_variante_data",
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}

type activateVariantRequest struct {
	ID int `json:"id"`
}

var activateVariantSchema = z.Struct(z.Shape{
	"ID": product.IDSchema.Required(),
})

func (h *CommandHandler) ActivateVariantHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := activateVariantRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, activateVariantSchema) {
			return
		}

		err := h.Command.ActivateVariant(r.Context(), body.ID)
		if err != nil {
			helper.MapError(w, err, map[error]string{
				application.ErrVarianteNotFound: "variante_not_found",
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}

type deactivateVariantRequest struct {
	ID int `json:"id"`
}

var deactivateVariantSchema = z.Struct(z.Shape{
	"ID": product.IDSchema.Required(),
})

func (h *CommandHandler) DeactivateVariantHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := deactivateVariantRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, deactivateVariantSchema) {
			return
		}

		err := h.Command.DeactivateVariant(r.Context(), body.ID)
		if err != nil {
			helper.MapError(w, err, map[error]string{
				application.ErrVarianteNotFound: "variante_not_found",
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}

type deleteProduktRequest struct {
	ID int `json:"id"`
}

var deleteProduktSchema = z.Struct(z.Shape{
	"ID": product.IDSchema.Required(),
})

func (h *CommandHandler) DeleteProduktHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := deleteProduktRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, deleteProduktSchema) {
			return
		}

		err := h.Command.DeleteProdukt(r.Context(), body.ID)
		if err != nil {
			helper.MapError(w, err, map[error]string{
				application.ErrProduktNotFound: "produkt_not_found",
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}

type deleteVarianteRequest struct {
	ProduktID int `json:"produktId"`
	ID        int `json:"id"`
}

var deleteVarianteSchema = z.Struct(z.Shape{
	"ProduktID": product.IDSchema.Required(),
	"ID":        product.IDSchema.Required(),
})

func (h *CommandHandler) DeleteVarianteHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := deleteVarianteRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, deleteVarianteSchema) {
			return
		}

		err := h.Command.DeleteVariante(r.Context(), body.ProduktID, body.ID)
		if err != nil {
			helper.MapError(w, err, map[error]string{
				application.ErrProduktNotFound:  "produkt_not_found",
				application.ErrVarianteNotFound: "variante_not_found",
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}
