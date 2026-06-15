package http

import (
	"context"
	"errors"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/settings/application"
	"github.com/nicograef/jotti/backend/domain/settings"
	"github.com/nicograef/jotti/backend/domain/tse"
)

type settingsCommand interface {
	UpdateBetreiber(ctx context.Context, b settings.Betreiber) error
	UpdateTSEKonfiguration(ctx context.Context, b settings.TSEKonfiguration) error
	RichteTSEEin(ctx context.Context, credentials tse.SetupCredentials, bestaetigteUmgebung tse.Umgebung) (application.TSESetupErgebnis, error)
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
	"Steuernummer": z.Ptr(z.String()),
	"UstID":        z.Ptr(z.String()),
})

type updateTSEKonfigurationRequest struct {
	ApiKey    string `json:"apiKey"`
	ApiSecret string `json:"apiSecret"`
	TssID     string `json:"tssId"`
	ClientID  string `json:"clientId"`
}

var updateTSEKonfigurationSchema = z.Struct(z.Shape{
	"ApiKey":    z.String().Max(500, z.Message("API-Key darf höchstens 500 Zeichen lang sein")).Optional(),
	"ApiSecret": z.String().Max(500, z.Message("API-Secret darf höchstens 500 Zeichen lang sein")).Optional(),
	"TssID":     z.String().Max(255, z.Message("TSS-ID darf höchstens 255 Zeichen lang sein")).Optional(),
	"ClientID":  z.String().Max(255, z.Message("Client-ID darf höchstens 255 Zeichen lang sein")).Optional(),
})

type tseEinrichtenRequest struct {
	ApiKey    string `json:"apiKey"`
	ApiSecret string `json:"apiSecret"`
	Umgebung  string `json:"umgebung"`
}

var tseEinrichtenSchema = z.Struct(z.Shape{
	"ApiKey":    z.String().Min(1, z.Message("API-Key ist erforderlich")).Max(500, z.Message("API-Key darf höchstens 500 Zeichen lang sein")).Required(),
	"ApiSecret": z.String().Min(1, z.Message("API-Secret ist erforderlich")).Max(500, z.Message("API-Secret darf höchstens 500 Zeichen lang sein")).Required(),
	"Umgebung":  z.String().OneOf([]string{string(tse.UmgebungTest), string(tse.UmgebungLive)}, z.Message("Ungültige Umgebung")).Required(),
})

// tseEinrichtenResponse uebergibt PUK und Admin-PIN genau einmal an die UI. Sie
// werden nie persistiert und nie geloggt; der Admin verwahrt sie extern.
type tseEinrichtenResponse struct {
	TssID    string `json:"tssId"`
	ClientID string `json:"clientId"`
	Puk      string `json:"puk"`
	AdminPin string `json:"adminPin"`
	Umgebung string `json:"umgebung"`
}

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

func (h *CommandHandler) UpdateTSEKonfigurationHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body updateTSEKonfigurationRequest
		if !helper.ReadAndValidateBody(w, r, &body, updateTSEKonfigurationSchema) {
			return
		}

		conf, err := settings.NewTSEKonfiguration(body.ApiKey, body.ApiSecret, body.TssID, body.ClientID)
		if err != nil {
			helper.SendClientError(w, "validation_error", nil)
			return
		}

		if err := h.Command.UpdateTSEKonfiguration(r.Context(), conf); err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendEmptyResponse(w)
	}
}

func (h *CommandHandler) RichteTSEEinHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body tseEinrichtenRequest
		if !helper.ReadAndValidateBody(w, r, &body, tseEinrichtenSchema) {
			return
		}

		ergebnis, err := h.Command.RichteTSEEin(
			r.Context(),
			tse.SetupCredentials{ApiKey: body.ApiKey, ApiSecret: body.ApiSecret},
			tse.Umgebung(body.Umgebung),
		)
		if err != nil {
			switch {
			case errors.Is(err, application.ErrTSESetupZugangsdaten):
				helper.SendClientError(w, "tse_setup_zugangsdaten_ungueltig", nil)
			case errors.Is(err, application.ErrTSESetupUmgebungAbweichung):
				helper.SendClientError(w, "tse_setup_umgebung_abweichung", nil)
			case errors.Is(err, application.ErrTSEBereitsEingerichtet):
				helper.SendClientError(w, "tse_bereits_eingerichtet", nil)
			case errors.Is(err, application.ErrTSEEinrichtung),
				errors.Is(err, application.ErrTSEVerbindungFehlgeschlagen):
				helper.SendClientError(w, "tse_einrichtung_fehlgeschlagen", nil)
			default:
				helper.SendServerError(w)
			}
			return
		}

		helper.SendResponse(w, tseEinrichtenResponse{
			TssID:    ergebnis.TssID,
			ClientID: ergebnis.ClientID,
			Puk:      ergebnis.PUK,
			AdminPin: ergebnis.AdminPIN,
			Umgebung: ergebnis.Umgebung,
		})
	}
}
