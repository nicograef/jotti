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
	"github.com/nicograef/jotti/backend/domain/kasse"
)

type pfandCommand interface {
	PfandZurueckgeben(ctx context.Context, userID int, userName string, rueckgabeID string, tischID int, positionen []enrichment.PositionInput, kommentar string) error
	PfandrueckgabeStornieren(ctx context.Context, userID int, userName string, stornierungID string, rueckgabeID string, positionen []kasse.PositionRef, kommentar string) error
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
			case errors.Is(err, application.ErrPfandrueckgabeAbweichend):
				helper.SendConflict(w, "pfandrueckgabe_abweichend")
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

type pfandrueckgabeStornierenRequest struct {
	StornierungID string               `json:"stornierungId"`
	RueckgabeID   string               `json:"rueckgabeId"`
	Positionen    []positionRefRequest `json:"positionen"`
	Kommentar     string               `json:"kommentar"`
}

var pfandrueckgabeStornierenSchema = z.Struct(z.Shape{
	"StornierungID": z.String().UUID().Required(),
	"RueckgabeID":   z.String().UUID().Required(),
	"Positionen":    z.Slice(positionRefRequestSchema).Min(1).Required(),
	"Kommentar":     z.String().Min(3).Max(100).Required(),
}).TestFunc(func(val any, _ z.Ctx) bool {
	body, ok := val.(*pfandrueckgabeStornierenRequest)
	// Beide IDs werden BON_ID im DSFinV-K-Export desselben Kassenabschlusses.
	return ok && body.StornierungID != body.RueckgabeID
}, z.Message("Stornierungs-ID muss sich von der Rückgabe-ID unterscheiden"))

func (h *PfandCommandHandler) PfandrueckgabeStornierenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := pfandrueckgabeStornierenRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, pfandrueckgabeStornierenSchema) {
			return
		}

		userID, userName, ok := middleware.UserFromContext(r.Context())
		if !ok {
			helper.SendServerError(w)
			return
		}

		err := h.Command.PfandrueckgabeStornieren(r.Context(), userID, userName, body.StornierungID, body.RueckgabeID, toPositionRefs(body.Positionen), body.Kommentar)
		if err != nil {
			switch {
			case errors.Is(err, application.ErrConflict):
				helper.SendConflictError(w)
			case errors.Is(err, application.ErrKasseWirdAbgeschlossen):
				helper.SendConflict(w, "kasse_wird_abgeschlossen")
			case errors.Is(err, application.ErrKasseNichtGeoeffnet):
				helper.SendConflict(w, "kasse_nicht_geoeffnet")
			case errors.Is(err, application.ErrRueckgabeNichtGefunden):
				helper.SendNotFound(w, "rueckgabe_not_found")
			case errors.Is(err, application.ErrPositionNichtStornierbar):
				helper.SendConflict(w, "position_nicht_stornierbar")
			case errors.Is(err, application.ErrPfandrueckgabeAbweichend):
				helper.SendConflict(w, "pfandrueckgabe_abweichend")
			default:
				helper.SendServerError(w)
			}
			return
		}

		helper.SendEmptyResponse(w)
	}
}
