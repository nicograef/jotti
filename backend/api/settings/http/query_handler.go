package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/settings/application"
	"github.com/nicograef/jotti/backend/domain/settings"
)

type settingsQuery interface {
	GetSeriennummer(ctx context.Context) (uuid.UUID, error)
	GetBetreiber(ctx context.Context) (settings.Betreiber, error)
}

type QueryHandler struct {
	Query settingsQuery
}

type seriennummerResponse struct {
	Seriennummer string `json:"seriennummer"`
}

type betreiberResponse struct {
	Vereinsname  string  `json:"vereinsname"`
	Strasse      string  `json:"strasse"`
	Plz          string  `json:"plz"`
	Ort          string  `json:"ort"`
	Steuernummer *string `json:"steuernummer"`
	UstID        *string `json:"ustId"`
}

func (h *QueryHandler) GetSeriennummerHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		seriennummer, err := h.Query.GetSeriennummer(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}
		helper.SendResponse(w, seriennummerResponse{Seriennummer: seriennummer.String()})
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
