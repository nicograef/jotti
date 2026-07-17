package http

import (
	"context"
	"net/http"
	"time"

	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/domain/tse"
)

// --- Query Handler ---

type tseSignaturauftragQuery interface {
	GetTSESignaturQueueZustand(ctx context.Context) (tse.SignaturQueueZustand, error)
	GetTSEStoerungen(ctx context.Context) ([]tse.Stoerungszeitraum, error)
}

type QueryHandler struct {
	Query tseSignaturauftragQuery
}

type signaturQueueResponse struct {
	OffeneAuftraege          int     `json:"offeneAuftraege"`
	FehlgeschlageneAuftraege int     `json:"fehlgeschlageneAuftraege"`
	LetzterFehler            string  `json:"letzterFehler"`
	RueckstandSekunden       int     `json:"rueckstandSekunden"`
	SignaturenProMinute      float64 `json:"signaturenProMinute"`
	SignierdauerP95Sekunden  float64 `json:"signierdauerP95Sekunden"`
}

// POST /admin/get-tse-signatur-queue
func (h *QueryHandler) GetTSESignaturQueueHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		zustand, err := h.Query.GetTSESignaturQueueZustand(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, signaturQueueResponse{
			OffeneAuftraege:          zustand.OffeneAuftraege,
			FehlgeschlageneAuftraege: zustand.FehlgeschlageneAuftraege,
			LetzterFehler:            zustand.LetzterFehler,
			RueckstandSekunden:       zustand.RueckstandSekunden,
			SignaturenProMinute:      zustand.SignaturenProMinute,
			SignierdauerP95Sekunden:  zustand.SignierdauerP95Sekunden,
		})
	}
}

type stoerungDTO struct {
	ID         int        `json:"id"`
	Beginn     time.Time  `json:"beginn"`
	Ende       *time.Time `json:"ende"`
	GrundArt   string     `json:"grundArt"`
	Fehlertext string     `json:"fehlertext"`
}

type getStoerungenResponse struct {
	Stoerungen []stoerungDTO `json:"stoerungen"`
}

// POST /admin/get-tse-stoerungen (Stoerungsprotokoll / Ausfalldokumentation)
func (h *QueryHandler) GetTSEStoerungenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stoerungen, err := h.Query.GetTSEStoerungen(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}

		dtos := make([]stoerungDTO, 0, len(stoerungen))
		for _, s := range stoerungen {
			dtos = append(dtos, stoerungDTO{
				ID:         s.ID,
				Beginn:     s.Beginn,
				Ende:       s.Ende,
				GrundArt:   s.GrundArt,
				Fehlertext: s.Fehlertext,
			})
		}

		helper.SendResponse(w, getStoerungenResponse{Stoerungen: dtos})
	}
}
