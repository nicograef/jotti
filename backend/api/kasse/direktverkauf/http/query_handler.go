package http

import (
	"context"
	"math"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/kasse/vorgang"
	"github.com/nicograef/jotti/backend/domain/kasse"
)

type query interface {
	GetDirektverkaufHistorie(ctx context.Context) ([]kasse.DirektverkaufHistorieEintrag, error)
	GetPfandrueckgabeHistorie(ctx context.Context, tischID int) ([]kasse.PfandrueckgabeHistorieEintrag, error)
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

type pfandrueckgabeHistorieEintrag struct {
	RueckgabeID string `json:"rueckgabeId"`
	TischID     int    `json:"tischId"`
	vorgang.HistorieEintrag
}

type getPfandrueckgabeHistorieRequest struct {
	TischID *int `json:"tischId"`
}

var getPfandrueckgabeHistorieSchema = z.Struct(z.Shape{
	// Pointer: tischId 0 (Direktverkauf) is valid, a missing tischId is not.
	"TischID": z.Ptr(z.Int().GTE(0).LTE(math.MaxInt32)).NotNil(z.Message("Tisch-ID ist erforderlich")),
})

type getPfandrueckgabeHistorieResponse struct {
	Historie []pfandrueckgabeHistorieEintrag `json:"historie"`
}

func toPfandrueckgabeHistorieEintrag(e kasse.PfandrueckgabeHistorieEintrag) pfandrueckgabeHistorieEintrag {
	return pfandrueckgabeHistorieEintrag{RueckgabeID: e.RueckgabeID, TischID: e.TischID, HistorieEintrag: vorgang.ToHistorieEintrag(e.VorgangHistorieEintrag)}
}

func (h *QueryHandler) GetPfandrueckgabeHistorieHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := getPfandrueckgabeHistorieRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, getPfandrueckgabeHistorieSchema) {
			return
		}

		historie, err := h.Query.GetPfandrueckgabeHistorie(r.Context(), *body.TischID)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		eintraege := make([]pfandrueckgabeHistorieEintrag, 0, len(historie))
		for i := range historie {
			eintraege = append(eintraege, toPfandrueckgabeHistorieEintrag(historie[i]))
		}

		helper.SendResponse(w, getPfandrueckgabeHistorieResponse{Historie: eintraege})
	}
}
