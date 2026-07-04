package http

import (
	"context"
	"net/http"
	"time"

	"github.com/nicograef/jotti/backend/api/helper"
	t "github.com/nicograef/jotti/backend/domain/table"
)

type query interface {
	GetAllTische(ctx context.Context) ([]t.Tisch, error)
}

type QueryHandler struct {
	Query query
}

type tisch struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type getAllTischeResponse struct {
	Tische []tisch `json:"tische"`
}

func toTisch(src t.Tisch) tisch {
	return tisch{
		ID:        src.ID,
		Name:      src.Name,
		Status:    string(src.Status),
		CreatedAt: src.CreatedAt,
		UpdatedAt: src.UpdatedAt,
	}
}

func toTische(tische []t.Tisch) []tisch {
	tischeResponse := make([]tisch, 0, len(tische))
	for _, tisch := range tische {
		tischeResponse = append(tischeResponse, toTisch(tisch))
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
