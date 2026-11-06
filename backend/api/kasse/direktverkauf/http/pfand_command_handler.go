package http

import (
	"context"
	"math"
	"net/http"
	"slices"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/kasse/direktverkauf/application"
	"github.com/nicograef/jotti/backend/api/kasse/eingabe"
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
	RueckgabeID string                  `json:"rueckgabeId"`
	TischID     *int                    `json:"tischId"`
	Positionen  []eingabe.PositionInput `json:"positionen"`
	Kommentar   string                  `json:"kommentar"`
}

var pfandZurueckgebenSchema = z.Struct(z.Shape{
	"RueckgabeID": z.String().UUID().Required(),
	// Pointer: tischId 0 (Direktverkauf) is valid, a missing tischId is not.
	"TischID":    z.Ptr(z.Int().GTE(0).LTE(math.MaxInt32)).NotNil(z.Message("Tisch-ID ist erforderlich")),
	"Positionen": z.Slice(eingabe.PositionInputSchema).Min(1).Required(),
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

		err := h.Command.PfandZurueckgeben(r.Context(), userID, userName, body.RueckgabeID, *body.TischID, eingabe.ToPositionInputs(body.Positionen), body.Kommentar)
		if err != nil {
			helper.MapError(w, err, slices.Concat(kasseFehler, []helper.ErrorCode{
				{Err: application.ErrPfandrueckgabeAbweichend, Code: "pfandrueckgabe_abweichend", Status: http.StatusConflict},
				{Err: application.ErrTischNichtGefunden, Code: "tisch_not_found", Status: http.StatusNotFound},
				{Err: application.ErrTischNichtAktiv, Code: "tisch_not_active", Status: http.StatusConflict},
			}, pfandEingabeFehler))
			return
		}

		helper.SendEmptyResponse(w)
	}
}

type pfandrueckgabeStornierenRequest struct {
	StornierungID string                `json:"stornierungId"`
	RueckgabeID   string                `json:"rueckgabeId"`
	Positionen    []eingabe.PositionRef `json:"positionen"`
	Kommentar     string                `json:"kommentar"`
}

var pfandrueckgabeStornierenSchema = z.Struct(z.Shape{
	"StornierungID": z.String().UUID().Required(),
	"RueckgabeID":   z.String().UUID().Required(),
	"Positionen":    z.Slice(eingabe.PositionRefSchema).Min(1).Required(),
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

		err := h.Command.PfandrueckgabeStornieren(r.Context(), userID, userName, body.StornierungID, body.RueckgabeID, eingabe.ToPositionRefs(body.Positionen), body.Kommentar)
		if err != nil {
			helper.MapError(w, err, slices.Concat(kasseFehler, []helper.ErrorCode{
				{Err: application.ErrRueckgabeNichtGefunden, Code: "rueckgabe_not_found", Status: http.StatusNotFound},
				{Err: application.ErrPositionNichtStornierbar, Code: "position_nicht_stornierbar", Status: http.StatusConflict},
				{Err: application.ErrPfandrueckgabeAbweichend, Code: "pfandrueckgabe_abweichend", Status: http.StatusConflict},
			}))
			return
		}

		helper.SendEmptyResponse(w)
	}
}
