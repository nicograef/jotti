package http

import (
	"context"
	"errors"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api/direktverkauf/application"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/middleware"
	"github.com/nicograef/jotti/backend/domain/product"
)

type command interface {
	DirektverkaufTaetigen(ctx context.Context, userID int, userName string, positionen []application.VerkaufPositionInput, kommentar string) error
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
	Positionen []verkaufPositionInput `json:"positionen"`
	Kommentar  string                 `json:"kommentar"`
}

var verkaufPositionInputSchema = z.Struct(z.Shape{
	"ProduktID":  product.IDSchema.Required(),
	"VarianteID": product.IDSchema.Required(),
	"Menge":      z.Int().GTE(1).Required(),
})

var direktverkaufTaetigenSchema = z.Struct(z.Shape{
	"Positionen": z.Slice(verkaufPositionInputSchema).Min(1).Required(),
	"Kommentar":  z.String().Max(100),
})

func toVerkaufPositionInputs(positionen []verkaufPositionInput) []application.VerkaufPositionInput {
	out := make([]application.VerkaufPositionInput, len(positionen))
	for i, p := range positionen {
		out[i] = application.VerkaufPositionInput{
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

		userID, ok := r.Context().Value(middleware.UserIDKey).(int)
		if !ok {
			helper.SendServerError(w)
			return
		}
		userName, _ := r.Context().Value(middleware.UserNameKey).(string)

		err := h.Command.DirektverkaufTaetigen(r.Context(), userID, userName, toVerkaufPositionInputs(body.Positionen), body.Kommentar)
		if err != nil {
			switch {
			case errors.Is(err, application.ErrConflict):
				helper.SendConflictError(w)
			case errors.Is(err, application.ErrKasseNichtGeoeffnet):
				helper.SendConflict(w, "kasse_nicht_geoeffnet")
			default:
				helper.MapError(w, err, map[error]string{
					application.ErrProduktNotFound: "produkt_not_found",
				})
			}
			return
		}

		helper.SendEmptyResponse(w)
	}
}
