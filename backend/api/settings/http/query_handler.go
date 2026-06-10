package http

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/settings/application"
	"github.com/nicograef/jotti/backend/domain/settings"
	"github.com/nicograef/jotti/backend/domain/tse"
)

type settingsQuery interface {
	GetKassenidentitaet(ctx context.Context) (settings.Kassenidentitaet, error)
	GetBetreiber(ctx context.Context) (settings.Betreiber, error)
	GetBondruckEinstellungen(ctx context.Context) (settings.BondruckEinstellungen, error)
	GetTSEKonfiguration(ctx context.Context) (settings.TSEKonfiguration, error)
	TestTSEVerbindung(ctx context.Context) (tse.VerbindungStatus, error)
}

type QueryHandler struct {
	Query settingsQuery
}

type kassenidentitaetResponse struct {
	Seriennummer string    `json:"seriennummer"`
	AngelegtAm   time.Time `json:"angelegtAm"`
}

type betreiberResponse struct {
	Vereinsname  string  `json:"vereinsname"`
	Strasse      string  `json:"strasse"`
	Plz          string  `json:"plz"`
	Ort          string  `json:"ort"`
	Steuernummer *string `json:"steuernummer"`
	UstID        *string `json:"ustId"`
}

type bondruckEinstellungenResponse struct {
	KassenbelegDruckerIP string `json:"kassenbelegDruckerIp"`
	DirektverkaufModus   string `json:"direktverkaufModus"`
	AbholbonDruckerIP    string `json:"abholbonDruckerIp"`
}

type tseKonfigurationResponse struct {
	ApiKeyGesetzt    bool   `json:"apiKeyGesetzt"`
	ApiSecretGesetzt bool   `json:"apiSecretGesetzt"`
	TssID            string `json:"tssId"`
	ClientID         string `json:"clientId"`
	IstKonfiguriert  bool   `json:"istKonfiguriert"`
}

type tseVerbindungResponse struct {
	Umgebung string `json:"umgebung"`
	TSSState string `json:"tssState"`
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

func (h *QueryHandler) GetBetreiberHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := h.Query.GetBetreiber(r.Context())
		if err != nil {
			if errors.Is(err, application.ErrNotFound) {
				helper.SendResponse(w, betreiberResponse{})
				return
			}
			helper.SendServerError(w)
			return
		}
		helper.SendResponse(w, betreiberResponse{
			Vereinsname:  b.Vereinsname,
			Strasse:      b.Strasse,
			Plz:          b.Plz,
			Ort:          b.Ort,
			Steuernummer: b.Steuernummer,
			UstID:        b.UstID,
		})
	}
}

func (h *QueryHandler) GetBondruckEinstellungenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := h.Query.GetBondruckEinstellungen(r.Context())
		if err != nil {
			if errors.Is(err, application.ErrNotFound) {
				helper.SendResponse(w, bondruckEinstellungenResponse{
					DirektverkaufModus: string(settings.DirektverkaufModusKeinBon),
				})
				return
			}
			helper.SendServerError(w)
			return
		}
		helper.SendResponse(w, bondruckEinstellungenResponse{
			KassenbelegDruckerIP: b.KassenbelegDruckerIP,
			DirektverkaufModus:   string(b.DirektverkaufModus),
			AbholbonDruckerIP:    b.AbholbonDruckerIP,
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
			Umgebung: string(status.Umgebung),
			TSSState: status.TSSState,
		})
	}
}
