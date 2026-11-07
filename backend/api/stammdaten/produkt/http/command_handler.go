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
	CreateProdukt(ctx context.Context, name string, kategorie dom.Kategorie, steuersatz steuer.Steuersatz, pfand bool) (int, error)
	UpdateProdukt(ctx context.Context, id int, name string, kategorie dom.Kategorie, steuersatz steuer.Steuersatz, pfand bool) error
	DeleteProdukt(ctx context.Context, produktID int) error
	VerschiebeProdukt(ctx context.Context, produktID int, richtung dom.Richtung) error
	CreateVariante(ctx context.Context, produktID int, name string, preisCents int, pfandVarianteID *int) (int, error)
	UpdateVariante(ctx context.Context, varianteID int, name string, preisCents int, pfandVarianteID *int) error
	ActivateVariante(ctx context.Context, varianteID int) error
	DeactivateVariante(ctx context.Context, varianteID int) error
	DeactivateAlleVarianten(ctx context.Context, produktID int) error
	DeleteVariante(ctx context.Context, produktID int, varianteID int) error
	VerschiebeVariante(ctx context.Context, varianteID int, richtung dom.Richtung) error
	SortiereVariantenAlphabetisch(ctx context.Context, produktID int) error
}

type CommandHandler struct {
	Command command
}

type createProduktRequest struct {
	Name       string            `json:"name"`
	Kategorie  dom.Kategorie     `json:"kategorie"`
	Steuersatz steuer.Steuersatz `json:"steuersatz"`
	Pfand      bool              `json:"pfand"`
}

var createProduktSchema = z.Struct(z.Shape{
	"Name":       dom.NameSchema.Required(),
	"Kategorie":  dom.KategorieSchema.Required(),
	"Steuersatz": steuer.SteuersatzSchema.Required(),
	"Pfand":      z.Bool(),
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

		id, err := h.Command.CreateProdukt(r.Context(), body.Name, body.Kategorie, body.Steuersatz, body.Pfand)
		if err != nil {
			helper.MapError(w, err, []helper.ErrorCode{
				{Err: application.ErrProduktAlreadyExists, Code: "produkt_already_exists"},
				{Err: application.ErrInvalidProduktData, Code: "invalid_produkt_data"},
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
	Pfand      bool              `json:"pfand"`
}

var updateProduktSchema = z.Struct(z.Shape{
	"ID":         dom.IDSchema.Required(),
	"Name":       dom.NameSchema.Required(),
	"Kategorie":  dom.KategorieSchema.Required(),
	"Steuersatz": steuer.SteuersatzSchema.Required(),
	"Pfand":      z.Bool(),
})

func (h *CommandHandler) UpdateProduktHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := updateProduktRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, updateProduktSchema) {
			return
		}

		err := h.Command.UpdateProdukt(r.Context(), body.ID, body.Name, body.Kategorie, body.Steuersatz, body.Pfand)
		if err != nil {
			helper.MapError(w, err, []helper.ErrorCode{
				{Err: application.ErrProduktNotFound, Code: "produkt_not_found"},
				{Err: application.ErrProduktAlreadyExists, Code: "produkt_already_exists"},
				{Err: application.ErrInvalidProduktData, Code: "invalid_produkt_data"},
				{Err: application.ErrPfandZielUngueltig, Code: "pfand_ziel_ungueltig"},
				{Err: application.ErrPfandVerkettet, Code: "pfand_verkettet"},
				{Err: application.ErrPfandSteuersatzAbweichend, Code: "pfand_steuersatz_abweichend"},
				{Err: application.ErrPfandNochVerknuepft, Code: "pfand_noch_verknuepft"},
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}

// verschiebeRequest moves a Produkt or a Variante one place up or down.
type verschiebeRequest struct {
	ID       int          `json:"id"`
	Richtung dom.Richtung `json:"richtung"`
}

var verschiebeSchema = z.Struct(z.Shape{
	"ID":       dom.IDSchema.Required(),
	"Richtung": dom.RichtungSchema.Required(),
})

func (h *CommandHandler) VerschiebeProduktHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := verschiebeRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, verschiebeSchema) {
			return
		}

		err := h.Command.VerschiebeProdukt(r.Context(), body.ID, body.Richtung)
		if err != nil {
			helper.MapError(w, err, []helper.ErrorCode{
				{Err: application.ErrProduktNotFound, Code: "produkt_not_found"},
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}

// createVarianteRequest: PfandVarianteID null or missing means no Pfand.
type createVarianteRequest struct {
	ProduktID       int    `json:"produktId"`
	Name            string `json:"name"`
	PreisCents      int    `json:"preisCents"`
	PfandVarianteID *int   `json:"pfandVarianteId"`
}

var createVarianteSchema = z.Struct(z.Shape{
	"ProduktID":       dom.IDSchema.Required(),
	"Name":            dom.NameSchema.Required(),
	"PreisCents":      dom.PreisCentsSchema,
	"PfandVarianteID": z.Ptr(dom.IDSchema),
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

		id, err := h.Command.CreateVariante(r.Context(), body.ProduktID, body.Name, body.PreisCents, body.PfandVarianteID)
		if err != nil {
			helper.MapError(w, err, []helper.ErrorCode{
				{Err: application.ErrProduktNotFound, Code: "produkt_not_found"},
				{Err: application.ErrInvalidVarianteData, Code: "invalid_variante_data"},
				{Err: application.ErrPfandZielUngueltig, Code: "pfand_ziel_ungueltig"},
				{Err: application.ErrPfandVerkettet, Code: "pfand_verkettet"},
				{Err: application.ErrPfandSteuersatzAbweichend, Code: "pfand_steuersatz_abweichend"},
			})
			return
		}

		helper.SendResponse(w, createVarianteResponse{ID: id})
	}
}

// updateVarianteRequest: PfandVarianteID null or missing removes the Pfand.
type updateVarianteRequest struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	PreisCents      int    `json:"preisCents"`
	PfandVarianteID *int   `json:"pfandVarianteId"`
}

var updateVarianteSchema = z.Struct(z.Shape{
	"ID":              dom.IDSchema.Required(),
	"Name":            dom.NameSchema.Required(),
	"PreisCents":      dom.PreisCentsSchema,
	"PfandVarianteID": z.Ptr(dom.IDSchema),
})

func (h *CommandHandler) UpdateVarianteHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := updateVarianteRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, updateVarianteSchema) {
			return
		}

		err := h.Command.UpdateVariante(r.Context(), body.ID, body.Name, body.PreisCents, body.PfandVarianteID)
		if err != nil {
			helper.MapError(w, err, []helper.ErrorCode{
				{Err: application.ErrVarianteNotFound, Code: "variante_not_found"},
				{Err: application.ErrInvalidVarianteData, Code: "invalid_variante_data"},
				{Err: application.ErrPfandZielUngueltig, Code: "pfand_ziel_ungueltig"},
				{Err: application.ErrPfandVerkettet, Code: "pfand_verkettet"},
				{Err: application.ErrPfandSteuersatzAbweichend, Code: "pfand_steuersatz_abweichend"},
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}

func (h *CommandHandler) VerschiebeVarianteHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := verschiebeRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, verschiebeSchema) {
			return
		}

		err := h.Command.VerschiebeVariante(r.Context(), body.ID, body.Richtung)
		if err != nil {
			helper.MapError(w, err, []helper.ErrorCode{
				{Err: application.ErrVarianteNotFound, Code: "variante_not_found"},
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}

// idRequest is the body of every endpoint that addresses one entity by its ID.
type idRequest struct {
	ID int `json:"id"`
}

var idSchema = z.Struct(z.Shape{
	"ID": dom.IDSchema.Required(),
})

func (h *CommandHandler) ActivateVarianteHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := idRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, idSchema) {
			return
		}

		err := h.Command.ActivateVariante(r.Context(), body.ID)
		if err != nil {
			helper.MapError(w, err, []helper.ErrorCode{
				{Err: application.ErrVarianteNotFound, Code: "variante_not_found"},
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}

func (h *CommandHandler) DeactivateVarianteHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := idRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, idSchema) {
			return
		}

		err := h.Command.DeactivateVariante(r.Context(), body.ID)
		if err != nil {
			helper.MapError(w, err, []helper.ErrorCode{
				{Err: application.ErrVarianteNotFound, Code: "variante_not_found"},
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}

func (h *CommandHandler) DeleteProduktHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := idRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, idSchema) {
			return
		}

		err := h.Command.DeleteProdukt(r.Context(), body.ID)
		if err != nil {
			helper.MapError(w, err, []helper.ErrorCode{
				{Err: application.ErrProduktNotFound, Code: "produkt_not_found"},
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
			helper.MapError(w, err, []helper.ErrorCode{
				{Err: application.ErrProduktNotFound, Code: "produkt_not_found"},
				{Err: application.ErrVarianteNotFound, Code: "variante_not_found"},
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}

type produktIDRequest struct {
	ProduktID int `json:"produktId"`
}

var produktIDSchema = z.Struct(z.Shape{
	"ProduktID": dom.IDSchema.Required(),
})

func (h *CommandHandler) SortiereVariantenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := produktIDRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, produktIDSchema) {
			return
		}

		err := h.Command.SortiereVariantenAlphabetisch(r.Context(), body.ProduktID)
		if err != nil {
			helper.MapError(w, err, []helper.ErrorCode{
				{Err: application.ErrProduktNotFound, Code: "produkt_not_found"},
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}

func (h *CommandHandler) DeactivateAlleVariantenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := produktIDRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, produktIDSchema) {
			return
		}

		err := h.Command.DeactivateAlleVarianten(r.Context(), body.ProduktID)
		if err != nil {
			helper.MapError(w, err, []helper.ErrorCode{
				{Err: application.ErrProduktNotFound, Code: "produkt_not_found"},
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}
