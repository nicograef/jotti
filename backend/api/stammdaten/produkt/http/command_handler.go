package http

import (
	"context"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/stammdaten/produkt/application"
	dom "github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/domain/steuer"
)

type command interface {
	CreateProdukt(ctx context.Context, name string, kategorie dom.Kategorie, steuersatz steuer.Steuersatz) (int, error)
	UpdateProdukt(ctx context.Context, id int, name string, kategorie dom.Kategorie, steuersatz steuer.Steuersatz) error
	DeleteProdukt(ctx context.Context, produktID int) error
	CreateVariante(ctx context.Context, produktID int, name string, preisCents int) (int, error)
	UpdateVariante(ctx context.Context, varianteID int, name string, preisCents int) error
	ActivateVariante(ctx context.Context, varianteID int) error
	DeactivateVariante(ctx context.Context, varianteID int) error
	DeleteVariante(ctx context.Context, produktID int, varianteID int) error
}

type CommandHandler struct {
	Command command
}

// Produkt handlers

type createProduktRequest struct {
	Name       string            `json:"name"`
	Kategorie  dom.Kategorie     `json:"kategorie"`
	Steuersatz steuer.Steuersatz `json:"steuersatz"`
}

var createProduktSchema = z.Struct(z.Shape{
	"Name":       dom.NameSchema.Required(),
	"Kategorie":  dom.KategorieSchema.Required(),
	"Steuersatz": steuer.SteuersatzSchema.Required(),
})

type createProduktResponse struct {
	ID int `json:"id"`
}

func (h *CommandHandler) CreateProduktHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := createProduktRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, createProduktSchema) {
			return
		}

		id, err := h.Command.CreateProdukt(r.Context(), body.Name, body.Kategorie, body.Steuersatz)
		if err != nil {
			helper.MapError(w, err, map[error]string{
				application.ErrProduktAlreadyExists: "produkt_already_exists",
				application.ErrInvalidProduktData:   "invalid_produkt_data",
			})
			return
		}

		helper.SendResponse(w, createProduktResponse{ID: id})
	}
}

type updateProduktRequest struct {
	ID         int               `json:"id"`
	Name       string            `json:"name"`
	Kategorie  dom.Kategorie     `json:"kategorie"`
	Steuersatz steuer.Steuersatz `json:"steuersatz"`
}

var updateProduktSchema = z.Struct(z.Shape{
	"ID":         dom.IDSchema.Required(),
	"Name":       dom.NameSchema.Required(),
	"Kategorie":  dom.KategorieSchema.Required(),
	"Steuersatz": steuer.SteuersatzSchema.Required(),
})

func (h *CommandHandler) UpdateProduktHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := updateProduktRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, updateProduktSchema) {
			return
		}

		err := h.Command.UpdateProdukt(r.Context(), body.ID, body.Name, body.Kategorie, body.Steuersatz)
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

// Variante handlers

type createVarianteRequest struct {
	ProduktID  int    `json:"produktId"`
	Name       string `json:"name"`
	PreisCents int    `json:"preisCents"`
}

var createVarianteSchema = z.Struct(z.Shape{
	"ProduktID":  dom.IDSchema.Required(),
	"Name":       dom.NameSchema.Required(),
	"PreisCents": dom.PreisCentsSchema.Required(),
})

type createVarianteResponse struct {
	ID int `json:"id"`
}

func (h *CommandHandler) CreateVarianteHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := createVarianteRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, createVarianteSchema) {
			return
		}

		id, err := h.Command.CreateVariante(r.Context(), body.ProduktID, body.Name, body.PreisCents)
		if err != nil {
			helper.MapError(w, err, map[error]string{
				application.ErrProduktNotFound:     "produkt_not_found",
				application.ErrInvalidVarianteData: "invalid_variante_data",
			})
			return
		}

		helper.SendResponse(w, createVarianteResponse{ID: id})
	}
}

type updateVarianteRequest struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	PreisCents int    `json:"preisCents"`
}

var updateVarianteSchema = z.Struct(z.Shape{
	"ID":         dom.IDSchema.Required(),
	"Name":       dom.NameSchema.Required(),
	"PreisCents": dom.PreisCentsSchema.Required(),
})

func (h *CommandHandler) UpdateVarianteHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := updateVarianteRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, updateVarianteSchema) {
			return
		}

		err := h.Command.UpdateVariante(r.Context(), body.ID, body.Name, body.PreisCents)
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

type activateVarianteRequest struct {
	ID int `json:"id"`
}

var activateVarianteSchema = z.Struct(z.Shape{
	"ID": dom.IDSchema.Required(),
})

func (h *CommandHandler) ActivateVarianteHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := activateVarianteRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, activateVarianteSchema) {
			return
		}

		err := h.Command.ActivateVariante(r.Context(), body.ID)
		if err != nil {
			helper.MapError(w, err, map[error]string{
				application.ErrVarianteNotFound: "variante_not_found",
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}

type deactivateVarianteRequest struct {
	ID int `json:"id"`
}

var deactivateVarianteSchema = z.Struct(z.Shape{
	"ID": dom.IDSchema.Required(),
})

func (h *CommandHandler) DeactivateVarianteHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := deactivateVarianteRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, deactivateVarianteSchema) {
			return
		}

		err := h.Command.DeactivateVariante(r.Context(), body.ID)
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
	"ID": dom.IDSchema.Required(),
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
	"ProduktID": dom.IDSchema.Required(),
	"ID":        dom.IDSchema.Required(),
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
