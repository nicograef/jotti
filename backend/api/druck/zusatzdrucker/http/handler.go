package http

import (
	"context"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api/druck/zusatzdrucker/application"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/domain/druckstation"
)

type zusatzdruckerQuery interface {
	GetAlle(ctx context.Context) ([]druckstation.Zusatzdrucker, error)
}

type QueryHandler struct {
	Query zusatzdruckerQuery
}

type zusatzdruckerDTO struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	DruckerIP string `json:"druckerIp"`
	Bonmodus  string `json:"bonmodus"`
}

type getZusatzdruckerResponse struct {
	Zusatzdrucker []zusatzdruckerDTO `json:"zusatzdrucker"`
}

func (h *QueryHandler) GetZusatzdruckerHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		alle, err := h.Query.GetAlle(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}

		dtos := make([]zusatzdruckerDTO, 0, len(alle))
		for _, zd := range alle {
			dtos = append(dtos, zusatzdruckerDTO{
				ID:        zd.ID,
				Name:      zd.Name,
				DruckerIP: zd.DruckerIP,
				Bonmodus:  string(zd.Bonmodus),
			})
		}

		helper.SendResponse(w, getZusatzdruckerResponse{Zusatzdrucker: dtos})
	}
}

type zusatzdruckerCommand interface {
	Anlegen(ctx context.Context, name string) (int, error)
	Aendern(ctx context.Context, id int, name, druckerIP, bonmodus string) error
	Loeschen(ctx context.Context, id int) error
	TestbonDrucken(ctx context.Context, id int) error
}

type CommandHandler struct {
	Command zusatzdruckerCommand
}

type createZusatzdruckerRequest struct {
	Name string `json:"name"`
}

var createZusatzdruckerSchema = z.Struct(z.Shape{
	"Name": druckstation.ZusatzdruckerNameSchema.Required(),
})

type createZusatzdruckerResponse struct {
	ID int `json:"id"`
}

func (h *CommandHandler) CreateZusatzdruckerHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body createZusatzdruckerRequest
		if !helper.ReadAndValidateBody(w, r, &body, createZusatzdruckerSchema) {
			return
		}

		id, err := h.Command.Anlegen(r.Context(), body.Name)
		if err != nil {
			helper.MapError(w, err, []helper.ErrorCode{
				{Err: application.ErrUngueltigerZusatzdrucker, Code: "validation_error"},
			})
			return
		}

		helper.SendResponse(w, createZusatzdruckerResponse{ID: id})
	}
}

type updateZusatzdruckerRequest struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	DruckerIP string `json:"druckerIp"`
	Bonmodus  string `json:"bonmodus"`
}

var updateZusatzdruckerSchema = z.Struct(z.Shape{
	"ID":        druckstation.ZusatzdruckerIDSchema.Required(),
	"Name":      druckstation.ZusatzdruckerNameSchema.Required(),
	"DruckerIP": z.String().IPv4(z.Message("Ungültige IPv4-Adresse")).Optional(),
	"Bonmodus": z.String().OneOf(
		druckstation.ZusatzdruckerBonmodi(),
		z.Message("Ungültiger Bonmodus"),
	).Required(),
})

func (h *CommandHandler) UpdateZusatzdruckerHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body updateZusatzdruckerRequest
		if !helper.ReadAndValidateBody(w, r, &body, updateZusatzdruckerSchema) {
			return
		}

		err := h.Command.Aendern(r.Context(), body.ID, body.Name, body.DruckerIP, body.Bonmodus)
		if err != nil {
			helper.MapError(w, err, []helper.ErrorCode{
				{Err: application.ErrZusatzdruckerNichtGefunden, Code: "zusatzdrucker_not_found"},
				{Err: application.ErrUngueltigerZusatzdrucker, Code: "validation_error"},
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}

// idRequest is the body of every endpoint that addresses one Zusatzdrucker.
type idRequest struct {
	ID int `json:"id"`
}

var idSchema = z.Struct(z.Shape{
	"ID": druckstation.ZusatzdruckerIDSchema.Required(),
})

func (h *CommandHandler) DeleteZusatzdruckerHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body idRequest
		if !helper.ReadAndValidateBody(w, r, &body, idSchema) {
			return
		}

		err := h.Command.Loeschen(r.Context(), body.ID)
		if err != nil {
			helper.MapError(w, err, []helper.ErrorCode{
				{Err: application.ErrZusatzdruckerNichtGefunden, Code: "zusatzdrucker_not_found"},
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}

func (h *CommandHandler) TestbonDruckenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body idRequest
		if !helper.ReadAndValidateBody(w, r, &body, idSchema) {
			return
		}

		err := h.Command.TestbonDrucken(r.Context(), body.ID)
		if err != nil {
			helper.MapError(w, err, []helper.ErrorCode{
				{Err: application.ErrZusatzdruckerNichtGefunden, Code: "zusatzdrucker_not_found"},
				{Err: application.ErrZusatzdruckerNichtKonfiguriert, Code: "zusatzdrucker_nicht_konfiguriert"},
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}
