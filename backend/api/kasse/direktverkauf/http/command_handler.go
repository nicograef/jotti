package http

import (
	"context"
	"errors"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/kasse/direktverkauf/application"
	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
	"github.com/nicograef/jotti/backend/api/middleware"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/produkt"
)

type command interface {
	DirektverkaufTaetigen(ctx context.Context, userID int, userName string, verkaufID string, positionen []enrichment.PositionInput, kommentar string) error
	DirektverkaufStornieren(ctx context.Context, userID int, userName string, verkaufID string, positionen []kasse.PositionRef, kommentar string) error
}

type CommandHandler struct {
	Command command
}

type verkaufPositionInput struct {
	ProduktID  int `json:"produktId"`
	VarianteID int `json:"varianteId"`
	Menge      int `json:"menge"`
}

type direktverkaufTaetigenRequest struct {
	VerkaufID  string                 `json:"verkaufId"`
	Positionen []verkaufPositionInput `json:"positionen"`
	Kommentar  string                 `json:"kommentar"`
}

var verkaufPositionInputSchema = z.Struct(z.Shape{
	"ProduktID":  produkt.IDSchema.Required(),
	"VarianteID": produkt.IDSchema.Required(),
	"Menge":      z.Int().GTE(1).Required(),
})

var direktverkaufTaetigenSchema = z.Struct(z.Shape{
	"VerkaufID":  z.String().UUID().Required(),
	"Positionen": z.Slice(verkaufPositionInputSchema).Min(1).Required(),
	"Kommentar":  z.String().Max(100),
})

func toVerkaufPositionInputs(positionen []verkaufPositionInput) []enrichment.PositionInput {
	out := make([]enrichment.PositionInput, len(positionen))
	for i, p := range positionen {
		out[i] = enrichment.PositionInput{
			ProduktID:  p.ProduktID,
			VarianteID: p.VarianteID,
			Menge:      p.Menge,
		}
	}
	return out
}

func (h *CommandHandler) DirektverkaufTaetigenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := direktverkaufTaetigenRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, direktverkaufTaetigenSchema) {
			return
		}

		userID, userName, ok := middleware.UserFromContext(r.Context())
		if !ok {
			helper.SendServerError(w)
			return
		}

		err := h.Command.DirektverkaufTaetigen(r.Context(), userID, userName, body.VerkaufID, toVerkaufPositionInputs(body.Positionen), body.Kommentar)
		if err != nil {
			switch {
			case errors.Is(err, application.ErrConflict):
				helper.SendConflictError(w)
			case errors.Is(err, application.ErrKasseWirdAbgeschlossen):
				helper.SendConflict(w, "kasse_wird_abgeschlossen")
			case errors.Is(err, application.ErrKasseNichtGeoeffnet):
				helper.SendConflict(w, "kasse_nicht_geoeffnet")
			default:
				helper.MapError(w, err, map[error]string{
					enrichment.ErrProduktNotFound:    "produkt_not_found",
					enrichment.ErrVarianteNichtAktiv: "variante_nicht_aktiv",
				})
			}
			return
		}

		helper.SendEmptyResponse(w)
	}
}

type positionRefRequest struct {
	PositionID string `json:"positionId"`
	Menge      int    `json:"menge"`
}

var positionRefRequestSchema = z.Struct(z.Shape{
	"PositionID": z.String().UUID().Required(),
	"Menge":      z.Int().GTE(1).Required(),
})

func toPositionRefs(refs []positionRefRequest) []kasse.PositionRef {
	out := make([]kasse.PositionRef, len(refs))
	for i, ref := range refs {
		out[i] = kasse.PositionRef{PositionID: ref.PositionID, Menge: ref.Menge}
	}
	return out
}

type direktverkaufStornierenRequest struct {
	VerkaufID  string               `json:"verkaufId"`
	Positionen []positionRefRequest `json:"positionen"`
	Kommentar  string               `json:"kommentar"`
}

var direktverkaufStornierenSchema = z.Struct(z.Shape{
	"VerkaufID":  z.String().UUID().Required(),
	"Positionen": z.Slice(positionRefRequestSchema).Min(1).Required(),
	"Kommentar":  z.String().Min(3).Max(100).Required(),
})

func (h *CommandHandler) DirektverkaufStornierenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := direktverkaufStornierenRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, direktverkaufStornierenSchema) {
			return
		}

		userID, userName, ok := middleware.UserFromContext(r.Context())
		if !ok {
			helper.SendServerError(w)
			return
		}

		err := h.Command.DirektverkaufStornieren(r.Context(), userID, userName, body.VerkaufID, toPositionRefs(body.Positionen), body.Kommentar)
		if err != nil {
			switch {
			case errors.Is(err, application.ErrConflict):
				helper.SendConflictError(w)
			case errors.Is(err, application.ErrKasseWirdAbgeschlossen):
				helper.SendConflict(w, "kasse_wird_abgeschlossen")
			case errors.Is(err, application.ErrKasseNichtGeoeffnet):
				helper.SendConflict(w, "kasse_nicht_geoeffnet")
			default:
				helper.MapError(w, err, map[error]string{
					application.ErrVerkaufNichtGefunden:     "verkauf_not_found",
					application.ErrPositionNichtStornierbar: "position_nicht_stornierbar",
				})
			}
			return
		}

		helper.SendEmptyResponse(w)
	}
}
