package http

import (
	"context"
	"net/http"

	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/kasse/vorgang"
	"github.com/nicograef/jotti/backend/domain/kasse"
)

type query interface {
	GetDirektverkaufHistorie(ctx context.Context) ([]kasse.DirektverkaufHistorieEintrag, error)
}

type QueryHandler struct {
	Query query
}

type direktverkaufHistorieEintrag struct {
	VerkaufID string `json:"verkaufId"`
	vorgang.HistorieEintrag
}

type getDirektverkaufHistorieResponse struct {
	Historie []direktverkaufHistorieEintrag `json:"historie"`
}

func toHistorieEintrag(e kasse.DirektverkaufHistorieEintrag) direktverkaufHistorieEintrag {
	return direktverkaufHistorieEintrag{VerkaufID: e.VerkaufID, HistorieEintrag: vorgang.ToHistorieEintrag(e.VorgangHistorieEintrag)}
}

func (h *QueryHandler) GetDirektverkaufHistorieHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		historie, err := h.Query.GetDirektverkaufHistorie(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}

		eintraege := make([]direktverkaufHistorieEintrag, 0, len(historie))
		for i := range historie {
			eintraege = append(eintraege, toHistorieEintrag(historie[i]))
		}

		helper.SendResponse(w, getDirektverkaufHistorieResponse{Historie: eintraege})
	}
}
