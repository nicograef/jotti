package http

import (
	"context"
	"errors"
	"math"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/kasse/direktverkauf/application"
	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
	"github.com/nicograef/jotti/backend/api/middleware"
)

type pfandCommand interface {
	PfandZurueckgeben(ctx context.Context, userID int, userName string, rueckgabeID string, tischID int, positionen []enrichment.PositionInput, kommentar string) error
}

type PfandCommandHandler struct {
	Command pfandCommand
}

type pfandZurueckgebenRequest struct {
	RueckgabeID string                 `json:"rueckgabeId"`
	TischID     *int                   `json:"tischId"`
	Positionen  []verkaufPositionInput `json:"positionen"`
	Kommentar   string                 `json:"kommentar"`
}

var pfandZurueckgebenSchema = z.Struct(z.Shape{
	"RueckgabeID": z.String().UUID().Required(),
	// Pointer: tischId 0 (Direktverkauf) is valid, a missing tischId is not.
	"TischID":    z.Ptr(z.Int().GTE(0).LTE(math.MaxInt32)).NotNil(z.Message("Tisch-ID ist erforderlich")),
	"Positionen": z.Slice(verkaufPositionInputSchema).Min(1).Required(),
	"Kommentar":  z.String().Max(100),
})

func (h *PfandCommandHandler) PfandZurueckgebenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := pfandZurueckgebenRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, pfandZurueckgebenSchema) {
			return
		}

		userID, userName, ok := middleware.UserFromContext(r.Context())
		if !ok {
			helper.SendServerError(w)
			return
		}

		err := h.Command.PfandZurueckgeben(r.Context(), userID, userName, body.RueckgabeID, *body.TischID, toVerkaufPositionInputs(body.Positionen), body.Kommentar)
		if err != nil {
			switch {
			case errors.Is(err, application.ErrConflict):
				helper.SendConflictError(w)
			case errors.Is(err, application.ErrKasseWirdAbgeschlossen):
				helper.SendConflict(w, "kasse_wird_abgeschlossen")
			case errors.Is(err, application.ErrKasseNichtGeoeffnet):
				helper.SendConflict(w, "kasse_nicht_geoeffnet")
			case errors.Is(err, application.ErrTischNichtGefunden):
				helper.SendNotFound(w, "tisch_not_found")
			case errors.Is(err, application.ErrTischNichtAktiv):
				helper.SendConflict(w, "tisch_not_active")
			default:
				helper.MapError(w, err, []helper.ErrorCode{
					{Err: enrichment.ErrKeinPfandProdukt, Code: "kein_pfand_produkt"},
					{Err: enrichment.ErrProduktNotFound, Code: "produkt_not_found"},
					{Err: enrichment.ErrVarianteNichtAktiv, Code: "variante_nicht_aktiv"},
				})
			}
			return
		}

		helper.SendEmptyResponse(w)
	}
}
