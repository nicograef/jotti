package http

import (
	"context"
	"net/http"
	"time"

	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/stammdaten/tisch/application"
)

type query interface {
	GetAllTische(ctx context.Context) ([]application.TischMitSaldo, error)
}

type QueryHandler struct {
	Query query
}

type tisch struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	SaldoCents int       `json:"saldoCents"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type getAllTischeResponse struct {
	Tische []tisch `json:"tische"`
}

func toTisch(src application.TischMitSaldo) tisch {
	return tisch{
		ID:         src.Tisch.ID,
		Name:       src.Tisch.Name,
		Status:     string(src.Tisch.Status),
		SaldoCents: src.SaldoCents,
		CreatedAt:  src.Tisch.CreatedAt,
		UpdatedAt:  src.Tisch.UpdatedAt,
	}
}

func toTische(tische []application.TischMitSaldo) []tisch {
	tischeResponse := make([]tisch, 0, len(tische))
	for i := range tische {
		tischeResponse = append(tischeResponse, toTisch(tische[i]))
	}

	return tischeResponse
}

func (h *QueryHandler) GetAllTischeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tische, err := h.Query.GetAllTische(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, getAllTischeResponse{Tische: toTische(tische)})
	}
}
