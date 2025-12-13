package http

import (
	"context"
	"net/http"

	"github.com/nicograef/jotti/backend/admin/table/domain"
	"github.com/nicograef/jotti/backend/api"
)

type query interface {
	GetAllTables(ctx context.Context) ([]domain.Table, error)
}

type QueryHandler struct {
	Query query
}

type getAllTablesResponse struct {
	Tables []domain.Table `json:"tables"`
}

// GetAllTablesHandler handles requests to retrieve all tables.
func (h *QueryHandler) GetAllTablesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tables, err := h.Query.GetAllTables(ctx)
		if err != nil {
			api.SendServerError(w)
			return
		}

		api.SendResponse(w, getAllTablesResponse{Tables: tables})
	}
}
