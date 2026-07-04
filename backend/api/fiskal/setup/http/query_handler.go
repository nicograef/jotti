package http

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api/fiskal/setup/application"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/domain/tse"
)

type settingsQuery interface {
	GetKassenidentitaet(ctx context.Context) (tse.Kassenidentitaet, error)
	GetTSEKonfiguration(ctx context.Context) (tse.Konfiguration, error)
	TestTSEVerbindung(ctx context.Context) (tse.VerbindungStatus, error)
	PruefeTSESetup(ctx context.Context, credentials tse.SetupCredentials) (application.TSESetupBefund, error)
	GetTSEStatus(ctx context.Context) (application.TSEStatus, error)
}

type QueryHandler struct {
	Query settingsQuery
}

type kassenidentitaetResponse struct {
	Seriennummer string    `json:"seriennummer"`
	AngelegtAm   time.Time `json:"angelegtAm"`
}

type tseKonfigurationResponse struct {
	ApiKeyGesetzt    bool   `json:"apiKeyGesetzt"`
	ApiSecretGesetzt bool   `json:"apiSecretGesetzt"`
	TssID            string `json:"tssId"`
	ClientID         string `json:"clientId"`
	IstKonfiguriert  bool   `json:"istKonfiguriert"`
}

type tseVerbindungResponse struct {
	Umgebung            string `json:"umgebung"`
	TSSState            string `json:"tssState"`
	ClientState         string `json:"clientState"`
	ClientSerialNumber  string `json:"clientSerialNumber"`
	SeriennummerKorrekt bool   `json:"seriennummerKorrekt"`
}

type tseStatusResponse struct {
	Umgebung        string `json:"umgebung"`
	IstKonfiguriert bool   `json:"istKonfiguriert"`
}

type pruefeTSESetupRequest struct {
	ApiKey    string `json:"apiKey"`
	ApiSecret string `json:"apiSecret"`
}

var pruefeTSESetupSchema = z.Struct(z.Shape{
	"ApiKey":    z.String().Min(1, z.Message("API-Key ist erforderlich")).Max(500, z.Message("API-Key darf höchstens 500 Zeichen lang sein")).Required(),
	"ApiSecret": z.String().Min(1, z.Message("API-Secret ist erforderlich")).Max(500, z.Message("API-Secret darf höchstens 500 Zeichen lang sein")).Required(),
})

type tseSetupBefundResponse struct {
	Umgebung      string              `json:"umgebung"`
	VorhandeneTSS []tssBefundResponse `json:"vorhandeneTss"`
}

type tssBefundResponse struct {
	ID              string                `json:"id"`
	State           string                `json:"state"`
	PassenderClient *clientBefundResponse `json:"passenderClient"`
}

type clientBefundResponse struct {
	ID           string `json:"id"`
	SerialNumber string `json:"serialNumber"`
	State        string `json:"state"`
}

func (h *QueryHandler) GetKassenidentitaetHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identitaet, err := h.Query.GetKassenidentitaet(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}
		helper.SendResponse(w, kassenidentitaetResponse{
			Seriennummer: identitaet.Seriennummer.String(),
			AngelegtAm:   identitaet.AngelegtAm,
		})
	}
}

func (h *QueryHandler) GetTSEKonfigurationHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := h.Query.GetTSEKonfiguration(r.Context())
		if err != nil {
			if errors.Is(err, application.ErrNotFound) {
				helper.SendResponse(w, tseKonfigurationResponse{})
				return
			}
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, tseKonfigurationResponse{
			ApiKeyGesetzt:    strings.TrimSpace(c.ApiKey) != "",
			ApiSecretGesetzt: strings.TrimSpace(c.ApiSecret) != "",
			TssID:            c.TssID,
			ClientID:         c.ClientID,
			IstKonfiguriert:  c.IstKonfiguriert(),
		})
	}
}

func (h *QueryHandler) TestTSEVerbindungHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, err := h.Query.TestTSEVerbindung(r.Context())
		if err != nil {
			switch {
			case errors.Is(err, application.ErrTSENichtKonfiguriert):
				helper.SendClientError(w, "tse_nicht_konfiguriert", nil)
			case errors.Is(err, application.ErrTSEVerbindungFehlgeschlagen):
				helper.SendClientError(w, "tse_verbindung_fehlgeschlagen", nil)
			default:
				helper.SendServerError(w)
			}
			return
		}

		helper.SendResponse(w, tseVerbindungResponse{
			Umgebung:            string(status.Umgebung),
			TSSState:            status.TSSState,
			ClientState:         status.ClientState,
			ClientSerialNumber:  status.ClientSerialNumber,
			SeriennummerKorrekt: status.SeriennummerKorrekt,
		})
	}
}

func (h *QueryHandler) PruefeTSESetupHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body pruefeTSESetupRequest
		if !helper.ReadAndValidateBody(w, r, &body, pruefeTSESetupSchema) {
			return
		}

		befund, err := h.Query.PruefeTSESetup(r.Context(), tse.SetupCredentials{
			ApiKey:    body.ApiKey,
			ApiSecret: body.ApiSecret,
		})
		if err != nil {
			switch {
			case errors.Is(err, application.ErrTSESetupZugangsdaten):
				helper.SendClientError(w, "tse_setup_zugangsdaten_ungueltig", nil)
			case errors.Is(err, application.ErrTSEVerbindungFehlgeschlagen):
				helper.SendClientError(w, "tse_verbindung_fehlgeschlagen", nil)
			default:
				helper.SendServerError(w)
			}
			return
		}

		tssListe := make([]tssBefundResponse, 0, len(befund.VorhandeneTSS))
		for _, t := range befund.VorhandeneTSS {
			var client *clientBefundResponse
			if t.PassenderClient != nil {
				client = &clientBefundResponse{
					ID:           t.PassenderClient.ID,
					SerialNumber: t.PassenderClient.SerialNumber,
					State:        t.PassenderClient.State,
				}
			}
			tssListe = append(tssListe, tssBefundResponse{
				ID:              t.ID,
				State:           t.State,
				PassenderClient: client,
			})
		}

		helper.SendResponse(w, tseSetupBefundResponse{
			Umgebung:      befund.Umgebung,
			VorhandeneTSS: tssListe,
		})
	}
}

func (h *QueryHandler) GetTSEStatusHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, err := h.Query.GetTSEStatus(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, tseStatusResponse{
			Umgebung:        status.Umgebung,
			IstKonfiguriert: status.IstKonfiguriert,
		})
	}
}
