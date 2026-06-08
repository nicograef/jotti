package http

import (
	"context"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/domain/settings"
)

type settingsCommand interface {
	UpdateBetreiber(ctx context.Context, b settings.Betreiber) error
	UpdateBondruckEinstellungen(ctx context.Context, b settings.BondruckEinstellungen) error
}

type CommandHandler struct {
	Command settingsCommand
}

type updateBetreiberRequest struct {
	Vereinsname  string  `json:"vereinsname"`
	Strasse      string  `json:"strasse"`
	Plz          string  `json:"plz"`
	Ort          string  `json:"ort"`
	Steuernummer *string `json:"steuernummer"`
	UstID        *string `json:"ustId"`
}

var updateBetreiberSchema = z.Struct(z.Shape{
	"Vereinsname":  z.String().Min(1, z.Message("Vereinsname ist erforderlich")).Required(),
	"Strasse":      z.String().Min(1, z.Message("Straße ist erforderlich")).Required(),
	"Plz":          z.String().Min(1, z.Message("PLZ ist erforderlich")).Required(),
	"Ort":          z.String().Min(1, z.Message("Ort ist erforderlich")).Required(),
	"Steuernummer": z.String().Optional(),
	"UstID":        z.String().Optional(),
})

type updateBondruckEinstellungenRequest struct {
	KassenbelegDruckerIP string `json:"kassenbelegDruckerIp"`
	DirektverkaufModus   string `json:"direktverkaufModus"`
	AbholbonDruckerIP    string `json:"abholbonDruckerIp"`
}

var updateBondruckEinstellungenSchema = z.Struct(z.Shape{
	"KassenbelegDruckerIP": z.String().IPv4(z.Message("Ungültige IPv4-Adresse")).Optional(),
	"DirektverkaufModus": z.String().OneOf(
		[]string{
			string(settings.DirektverkaufModusKeinBon),
			string(settings.DirektverkaufModusAbholbon),
			string(settings.DirektverkaufModusAnStationen),
		},
		z.Message("Ungültiger Direktverkauf-Modus"),
	).Required(),
	"AbholbonDruckerIP": z.String().IPv4(z.Message("Ungültige IPv4-Adresse")).Optional(),
})

func (h *CommandHandler) UpdateBetreiberHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body updateBetreiberRequest
		if !helper.ReadAndValidateBody(w, r, &body, updateBetreiberSchema) {
			return
		}

		b, err := settings.NewBetreiber(body.Vereinsname, body.Strasse, body.Plz, body.Ort, body.Steuernummer, body.UstID)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		if err := h.Command.UpdateBetreiber(r.Context(), b); err != nil {
			helper.SendServerError(w)
			return
		}
		helper.SendEmptyResponse(w)
	}
}

func (h *CommandHandler) UpdateBondruckEinstellungenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body updateBondruckEinstellungenRequest
		if !helper.ReadAndValidateBody(w, r, &body, updateBondruckEinstellungenSchema) {
			return
		}

		b, err := settings.NewBondruckEinstellungen(
			body.KassenbelegDruckerIP,
			settings.DirektverkaufModus(body.DirektverkaufModus),
			body.AbholbonDruckerIP,
		)
		if err != nil {
			helper.SendClientError(w, "validation_error", nil)
			return
		}

		if err := h.Command.UpdateBondruckEinstellungen(r.Context(), b); err != nil {
			helper.SendServerError(w)
			return
		}
		helper.SendEmptyResponse(w)
	}
}
