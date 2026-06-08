package http

import (
	"context"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/repository/druckstation_repo"
)

// --- Query Handler ---

type druckstationQuery interface {
	GetAlleDruckstationen(ctx context.Context) ([]druckstation_repo.Druckstation, error)
}

type QueryHandler struct {
	Query druckstationQuery
}

type druckstationDTO struct {
	Kategorie string `json:"kategorie"`
	DruckerIP string `json:"druckerIp"`
	Bonmodus  string `json:"bonmodus"`
}

type getDruckstationenResponse struct {
	Druckstationen []druckstationDTO `json:"druckstationen"`
}

// POST /admin/get-druckstationen
func (h *QueryHandler) GetDruckstationenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		konfigs, err := h.Query.GetAlleDruckstationen(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}

		dtos := make([]druckstationDTO, 0, len(konfigs))
		for _, k := range konfigs {
			dtos = append(dtos, druckstationDTO{
				Kategorie: k.Kategorie,
				DruckerIP: k.DruckerIP,
				Bonmodus:  k.Bonmodus,
			})
		}

		helper.SendResponse(w, getDruckstationenResponse{Druckstationen: dtos})
	}
}

// --- Command Handler ---

type druckstationCommand interface {
	UpsertDruckstation(ctx context.Context, kategorie, druckerIP, bonmodus string) error
}

type CommandHandler struct {
	Command druckstationCommand
}

type updateDruckstationenRequest struct {
	Kategorie string `json:"kategorie"`
	DruckerIP string `json:"druckerIp"`
	Bonmodus  string `json:"bonmodus"`
}

var updateDruckstationenSchema = z.Struct(z.Shape{
	"Kategorie": z.String().OneOf(
		[]string{"essen", "getraenk", "sonstiges"},
		z.Message("Ungültige Kategorie"),
	).Required(),
	"DruckerIP": z.String().IPv4(z.Message("Ungültige IPv4-Adresse")).Optional(),
	"Bonmodus": z.String().OneOf(
		[]string{"pro_position", "pro_bestellung"},
		z.Message("Ungültiger Bonmodus"),
	).Required(),
})

// POST /admin/update-druckstationen
func (h *CommandHandler) UpdateDruckstationenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body updateDruckstationenRequest
		if !helper.ReadAndValidateBody(w, r, &body, updateDruckstationenSchema) {
			return
		}

		err := h.Command.UpsertDruckstation(r.Context(), body.Kategorie, body.DruckerIP, body.Bonmodus)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendEmptyResponse(w)
	}
}
