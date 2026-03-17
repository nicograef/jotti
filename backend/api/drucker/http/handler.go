package http

import (
	"context"
	"net/http"
	"regexp"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/repository/drucker_repo"
)

// --- Query Handler ---

type druckerQuery interface {
	GetAlleKategorieDrucker(ctx context.Context) ([]drucker_repo.DruckerKonfig, error)
}

type QueryHandler struct {
	Query druckerQuery
}

type druckerKonfigDTO struct {
	Kategorie string `json:"kategorie"`
	DruckerIP string `json:"druckerIp"`
	Bonmodus  string `json:"bonmodus"`
}

type getDruckerConfigResponse struct {
	Drucker []druckerKonfigDTO `json:"drucker"`
}

// POST /admin/get-drucker-config
func (h *QueryHandler) GetDruckerConfigHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		konfigs, err := h.Query.GetAlleKategorieDrucker(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}

		dtos := make([]druckerKonfigDTO, 0, len(konfigs))
		for _, k := range konfigs {
			dtos = append(dtos, druckerKonfigDTO{
				Kategorie: k.Kategorie,
				DruckerIP: k.DruckerIP,
				Bonmodus:  k.Bonmodus,
			})
		}

		helper.SendResponse(w, getDruckerConfigResponse{Drucker: dtos})
	}
}

// --- Command Handler ---

type druckerCommand interface {
	UpsertKategorieDrucker(ctx context.Context, kategorie, druckerIP, bonmodus string) error
}

type CommandHandler struct {
	Command druckerCommand
}

type updateDruckerConfigRequest struct {
	Kategorie string `json:"kategorie"`
	DruckerIP string `json:"druckerIp"`
	Bonmodus  string `json:"bonmodus"`
}

var ipv4Regex = regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)

var updateDruckerConfigSchema = z.Struct(z.Shape{
	"Kategorie": z.String().OneOf(
		[]string{"essen", "getraenk", "sonstiges"},
		z.Message("Ungültige Kategorie"),
	).Required(),
	"DruckerIP": z.String().Match(ipv4Regex, z.Message("Ungültige IPv4-Adresse")).Optional(),
	"Bonmodus": z.String().OneOf(
		[]string{"pro_position", "pro_bestellung"},
		z.Message("Ungültiger Bonmodus"),
	).Required(),
})

// POST /admin/update-drucker-config
func (h *CommandHandler) UpdateDruckerConfigHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body updateDruckerConfigRequest
		if !helper.ReadAndValidateBody(w, r, &body, updateDruckerConfigSchema) {
			return
		}

		err := h.Command.UpsertKategorieDrucker(r.Context(), body.Kategorie, body.DruckerIP, body.Bonmodus)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendEmptyResponse(w)
	}
}
