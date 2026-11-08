package http

import (
	"context"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api/druck/zusatzdrucker/application"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/domain/druckstation"
	"github.com/nicograef/jotti/backend/domain/produkt"
)

type zusatzdruckerQuery interface {
	GetAlle(ctx context.Context) (application.Uebersicht, error)
}

type QueryHandler struct {
	Query zusatzdruckerQuery
}

type zugeordnetesProduktDTO struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type zusatzdruckerDTO struct {
	ID        int                      `json:"id"`
	Name      string                   `json:"name"`
	DruckerIP string                   `json:"druckerIp"`
	Bonmodus  string                   `json:"bonmodus"`
	Produkte  []zugeordnetesProduktDTO `json:"produkte"`
}

// produktAuswahlDTO carries an empty zusatzdruckerName for a Produkt without Zuordnung.
type produktAuswahlDTO struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	Kategorie         string `json:"kategorie"`
	ZusatzdruckerName string `json:"zusatzdruckerName"`
}

type getZusatzdruckerResponse struct {
	Zusatzdrucker  []zusatzdruckerDTO  `json:"zusatzdrucker"`
	ProduktAuswahl []produktAuswahlDTO `json:"produktAuswahl"`
}

func (h *QueryHandler) GetZusatzdruckerHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uebersicht, err := h.Query.GetAlle(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}

		dtos := make([]zusatzdruckerDTO, 0, len(uebersicht.Zusatzdrucker))
		for _, zd := range uebersicht.Zusatzdrucker {
			produkte := make([]zugeordnetesProduktDTO, 0, len(zd.Produkte))
			for _, p := range zd.Produkte {
				produkte = append(produkte, zugeordnetesProduktDTO{ID: p.ProduktID, Name: p.ProduktName})
			}
			dtos = append(dtos, zusatzdruckerDTO{
				ID:        zd.ID,
				Name:      zd.Name,
				DruckerIP: zd.DruckerIP,
				Bonmodus:  string(zd.Bonmodus),
				Produkte:  produkte,
			})
		}

		auswahl := make([]produktAuswahlDTO, 0, len(uebersicht.ProduktAuswahl))
		for _, p := range uebersicht.ProduktAuswahl {
			auswahl = append(auswahl, produktAuswahlDTO{
				ID:                p.ProduktID,
				Name:              p.ProduktName,
				Kategorie:         string(p.Kategorie),
				ZusatzdruckerName: p.ZusatzdruckerName,
			})
		}

		helper.SendResponse(w, getZusatzdruckerResponse{Zusatzdrucker: dtos, ProduktAuswahl: auswahl})
	}
}

type zusatzdruckerCommand interface {
	Anlegen(ctx context.Context, name string) (int, error)
	Aendern(ctx context.Context, id int, name, druckerIP, bonmodus string) error
	Loeschen(ctx context.Context, id int) error
	TestbonDrucken(ctx context.Context, id int) error
	ProduktZuordnen(ctx context.Context, produktID, zusatzdruckerID int) error
	ZuordnungEntfernen(ctx context.Context, produktID int) error
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

type assignProduktRequest struct {
	ProduktID       int `json:"produktId"`
	ZusatzdruckerID int `json:"zusatzdruckerId"`
}

var assignProduktSchema = z.Struct(z.Shape{
	"ProduktID":       produkt.IDSchema.Required(),
	"ZusatzdruckerID": druckstation.ZusatzdruckerIDSchema.Required(),
})

func (h *CommandHandler) AssignProduktHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body assignProduktRequest
		if !helper.ReadAndValidateBody(w, r, &body, assignProduktSchema) {
			return
		}

		err := h.Command.ProduktZuordnen(r.Context(), body.ProduktID, body.ZusatzdruckerID)
		if err != nil {
			helper.MapError(w, err, []helper.ErrorCode{
				{Err: application.ErrZuordnungNichtMoeglich, Code: "zuordnung_nicht_moeglich"},
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}

type unassignProduktRequest struct {
	ProduktID int `json:"produktId"`
}

var unassignProduktSchema = z.Struct(z.Shape{
	"ProduktID": produkt.IDSchema.Required(),
})

func (h *CommandHandler) UnassignProduktHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body unassignProduktRequest
		if !helper.ReadAndValidateBody(w, r, &body, unassignProduktSchema) {
			return
		}

		if err := h.Command.ZuordnungEntfernen(r.Context(), body.ProduktID); err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendEmptyResponse(w)
	}
}
