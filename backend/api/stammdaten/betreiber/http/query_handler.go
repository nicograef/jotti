package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/stammdaten/betreiber/application"
	"github.com/nicograef/jotti/backend/domain/betreiber"
)

type betreiberQuery interface {
	GetBetreiber(ctx context.Context) (betreiber.Betreiber, error)
}

type QueryHandler struct {
	Query betreiberQuery
}

type betreiberResponse struct {
	Vereinsname  string  `json:"vereinsname"`
	Strasse      string  `json:"strasse"`
	Plz          string  `json:"plz"`
	Ort          string  `json:"ort"`
	Steuernummer *string `json:"steuernummer"`
	UstID        *string `json:"ustId"`
	// ElsterGemeldetAm ist das Datum der ELSTER-Kassenmeldung als YYYY-MM-DD
	// oder null, solange die Kasse noch nicht gemeldet wurde.
	ElsterGemeldetAm *string `json:"elsterGemeldetAm"`
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
		var elsterGemeldetAm *string
		if b.ElsterGemeldetAm != nil {
			datum := b.ElsterGemeldetAm.Format("2006-01-02")
			elsterGemeldetAm = &datum
		}
		helper.SendResponse(w, betreiberResponse{
			Vereinsname:      b.Vereinsname,
			Strasse:          b.Strasse,
			Plz:              b.Plz,
			Ort:              b.Ort,
			Steuernummer:     b.Steuernummer,
			UstID:            b.UstID,
			ElsterGemeldetAm: elsterGemeldetAm,
		})
	}
}
