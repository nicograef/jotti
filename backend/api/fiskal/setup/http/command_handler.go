package http

import (
	"context"
	"net/http"
	"slices"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api/fiskal/setup/application"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/domain/tse"
)

// tseSetupWriteTimeout replaces the server's 10 s write deadline and budgets only the response write.
// The handlers reset it after the lifecycle, so the one-time PUK and admin PIN still reach the client.
const tseSetupWriteTimeout = 2 * time.Minute

// tseSetupLebenszyklusTimeout only keeps a hung fiskaly connection from holding the detached
// lifecycle open; it is no response budget. It must exceed the worst case of about 10 min: up to eleven
// calls of four 10 s attempts plus three retry delays of at most 5 s each.
const tseSetupLebenszyklusTimeout = 15 * time.Minute

// lebenszyklusKontext detaches the lifecycle from client cancellation but keeps the context values;
// see docs/handbuch.md §3.13. A client abort would otherwise leave a paid, half-built TSS whose
// admin PIN existed only in the lost response.
func lebenszyklusKontext(r *http.Request) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(r.Context()), tseSetupLebenszyklusTimeout)
}

type tseSetupCommand interface {
	UpdateTSEKonfiguration(ctx context.Context, b tse.Konfiguration) error
	RichteTSEEin(ctx context.Context, credentials tse.SetupCredentials, bestaetigteUmgebung tse.Umgebung, neuAnlegenTrotzVorhandener bool) (application.TSESetupErgebnis, error)
	UebernimmTSE(ctx context.Context, credentials tse.SetupCredentials, bestaetigteUmgebung tse.Umgebung, tssID, pin, puk string) (application.TSESetupErgebnis, error)
}

type CommandHandler struct {
	Command tseSetupCommand
}

// setupFehlerKopf opens the Einrichtung and Übernahme tables. 409, not 400: the request is valid,
// another path just holds the setup lock.
var setupFehlerKopf = []helper.ErrorCode{
	{Err: application.ErrTSESetupLaeuftBereits, Code: "tse_setup_laeuft_bereits", Status: http.StatusConflict},
	{Err: application.ErrTSESetupZugangsdaten, Code: "tse_setup_zugangsdaten_ungueltig"},
	{Err: application.ErrTSESetupUmgebungAbweichung, Code: "tse_setup_umgebung_abweichung"},
	{Err: application.ErrTSEKonfigurationKassensitzungOffen, Code: "tse_konfiguration_kassensitzung_offen"},
}

// setupFehlerSchluss closes them: any other failure of the fiskaly lifecycle.
var setupFehlerSchluss = []helper.ErrorCode{
	{Err: application.ErrTSEEinrichtung, Code: "tse_einrichtung_fehlgeschlagen"},
	{Err: application.ErrTSEVerbindungFehlgeschlagen, Code: "tse_einrichtung_fehlgeschlagen"},
}

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

// The credential and environment fields shared by the Einrichtung and the Übernahme.
var (
	setupApiKeySchema    = z.String().Min(1, z.Message("API-Key ist erforderlich")).Max(500, z.Message("API-Key darf höchstens 500 Zeichen lang sein")).Required()
	setupApiSecretSchema = z.String().Min(1, z.Message("API-Secret ist erforderlich")).Max(500, z.Message("API-Secret darf höchstens 500 Zeichen lang sein")).Required()
	setupUmgebungSchema  = z.String().OneOf([]string{string(tse.UmgebungTest), string(tse.UmgebungLive)}, z.Message("Ungültige Umgebung")).Required()
)

type tseEinrichtenRequest struct {
	ApiKey                     string `json:"apiKey"`
	ApiSecret                  string `json:"apiSecret"`
	Umgebung                   string `json:"umgebung"`
	NeuAnlegenTrotzVorhandener bool   `json:"neuAnlegenTrotzVorhandener"`
}

// NeuAnlegenTrotzVorhandener bypasses the existing-TSS lock in TEST only; LIVE ignores it.
var tseEinrichtenSchema = z.Struct(z.Shape{
	"ApiKey":                     setupApiKeySchema,
	"ApiSecret":                  setupApiSecretSchema,
	"Umgebung":                   setupUmgebungSchema,
	"NeuAnlegenTrotzVorhandener": z.Bool().Optional(),
})

type tseUebernehmenRequest struct {
	ApiKey    string `json:"apiKey"`
	ApiSecret string `json:"apiSecret"`
	Umgebung  string `json:"umgebung"`
	TssID     string `json:"tssId"`
	Pin       string `json:"pin"`
	Puk       string `json:"puk"`
}

// Pin is the stored admin PIN, needed from UNINITIALIZED on.
// Puk is set only for a PIN reset after the PIN was lost or locked.
var tseUebernehmenSchema = z.Struct(z.Shape{
	"ApiKey":    setupApiKeySchema,
	"ApiSecret": setupApiSecretSchema,
	"Umgebung":  setupUmgebungSchema,
	"TssID":     z.String().Min(1, z.Message("TSS-ID ist erforderlich")).Max(255, z.Message("TSS-ID darf höchstens 255 Zeichen lang sein")).Required(),
	"Pin":       z.String().Max(50, z.Message("Admin-PIN darf höchstens 50 Zeichen lang sein")).Optional(),
	"Puk":       z.String().Max(100, z.Message("Admin-PUK darf höchstens 100 Zeichen lang sein")).Optional(),
})

// tseEinrichtenResponse is the only delivery of PUK and admin PIN: never persisted or logged.
type tseEinrichtenResponse struct {
	TssID    string `json:"tssId"`
	ClientID string `json:"clientId"`
	Puk      string `json:"puk"`
	AdminPin string `json:"adminPin"`
	Umgebung string `json:"umgebung"`
}

func (h *CommandHandler) UpdateTSEKonfigurationHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body updateTSEKonfigurationRequest
		if !helper.ReadAndValidateBody(w, r, &body, updateTSEKonfigurationSchema) {
			return
		}

		conf, err := tse.NewKonfiguration(body.ApiKey, body.ApiSecret, body.TssID, body.ClientID)
		if err != nil {
			helper.SendClientError(w, "validation_error", nil)
			return
		}

		if err := h.Command.UpdateTSEKonfiguration(r.Context(), conf); err != nil {
			// 409: this path shares the setup lock with RichteTSEEin and UebernimmTSE.
			helper.MapError(w, err, []helper.ErrorCode{
				{Err: application.ErrTSESetupLaeuftBereits, Code: "tse_setup_laeuft_bereits", Status: http.StatusConflict},
				{Err: application.ErrTSEKonfigurationKassensitzungOffen, Code: "tse_konfiguration_kassensitzung_offen"},
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}

// RichteTSEEinHandler runs the lifecycle under lebenszyklusKontext, whether or not the client still
// listens.
func (h *CommandHandler) RichteTSEEinHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		helper.ExtendWriteDeadline(w, r, tseSetupWriteTimeout)

		var body tseEinrichtenRequest
		if !helper.ReadAndValidateBody(w, r, &body, tseEinrichtenSchema) {
			return
		}

		ctx, cancel := lebenszyklusKontext(r)
		defer cancel()

		ergebnis, err := h.Command.RichteTSEEin(
			ctx,
			tse.SetupCredentials{ApiKey: body.ApiKey, ApiSecret: body.ApiSecret},
			tse.Umgebung(body.Umgebung),
			body.NeuAnlegenTrotzVorhandener,
		)

		// The deadline set on entry is absolute and may have expired during a long lifecycle.
		// Resetting it here covers both the error and the success branch.
		helper.ExtendWriteDeadline(w, r, tseSetupWriteTimeout)

		if err != nil {
			helper.MapError(w, err, slices.Concat(setupFehlerKopf, []helper.ErrorCode{
				{Err: application.ErrTSEBereitsEingerichtet, Code: "tse_bereits_eingerichtet"},
				{Err: application.ErrTSESetupTSSLimitErreicht, Code: "tse_setup_tss_limit_erreicht"},
			}, setupFehlerSchluss))
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

// UebernimmTSEHandler also runs under lebenszyklusKontext: a takeover can create a PIN that exists
// only in this response.
func (h *CommandHandler) UebernimmTSEHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		helper.ExtendWriteDeadline(w, r, tseSetupWriteTimeout)

		var body tseUebernehmenRequest
		if !helper.ReadAndValidateBody(w, r, &body, tseUebernehmenSchema) {
			return
		}

		ctx, cancel := lebenszyklusKontext(r)
		defer cancel()

		ergebnis, err := h.Command.UebernimmTSE(
			ctx,
			tse.SetupCredentials{ApiKey: body.ApiKey, ApiSecret: body.ApiSecret},
			tse.Umgebung(body.Umgebung),
			body.TssID,
			body.Pin,
			body.Puk,
		)

		// Reset the absolute write deadline, as in RichteTSEEinHandler.
		helper.ExtendWriteDeadline(w, r, tseSetupWriteTimeout)

		if err != nil {
			helper.MapError(w, err, slices.Concat(setupFehlerKopf, []helper.ErrorCode{
				{Err: application.ErrTSESetupTSSNichtGefunden, Code: "tse_setup_tss_nicht_gefunden"},
				{Err: application.ErrTSESetupPINErforderlich, Code: "tse_setup_pin_erforderlich"},
				{Err: application.ErrTSESetupPINUnbekannt, Code: "tse_setup_pin_unbekannt"},
				{Err: application.ErrTSESetupPUKUnbekannt, Code: "tse_setup_puk_unbekannt"},
				{Err: application.ErrTSESetupUebernahmeNichtMoeglich, Code: "tse_setup_uebernahme_nicht_moeglich"},
			}, setupFehlerSchluss))
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
