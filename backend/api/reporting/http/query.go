package http

import (
	"context"
	"net/http"
	"time"

	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/domain/reporting"
)

type query interface {
	GetDashboardData(ctx context.Context) (reporting.DashboardData, error)
	GetTagesabrechnung(ctx context.Context, von, bis time.Time) (reporting.TagesabrechnungData, error)
}

type QueryHandler struct {
	Query query
}

func (h QueryHandler) GetDashboardHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := h.Query.GetDashboardData(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}
		helper.SendResponse(w, data)
	}
}

type getTagesabrechnungRequest struct {
	Von time.Time `json:"von"`
	Bis time.Time `json:"bis"`
}

func (h QueryHandler) GetTagesabrechnungHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := getTagesabrechnungRequest{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		if body.Von.IsZero() || body.Bis.IsZero() || !body.Von.Before(body.Bis) {
			helper.SendClientError(w, "invalid_zeitraum", nil)
			return
		}

		data, err := h.Query.GetTagesabrechnung(r.Context(), body.Von, body.Bis)
		if err != nil {
			helper.SendServerError(w)
			return
		}
		helper.SendResponse(w, data)
	}
}
