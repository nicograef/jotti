package http

import (
	"context"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/domain/betreiber"
)

type betreiberCommand interface {
	UpdateBetreiber(ctx context.Context, b betreiber.Betreiber) error
	SetzeElsterMeldung(ctx context.Context) error
	NimmElsterMeldungZurueck(ctx context.Context) error
}

type CommandHandler struct {
	Command betreiberCommand
}

type updateBetreiberRequest struct {
	Vereinsname  string  `json:"vereinsname"`
	Strasse      string  `json:"strasse"`
	Plz          string  `json:"plz"`
	Ort          string  `json:"ort"`
	Steuernummer *string `json:"steuernummer"`
	UstID        *string `json:"ustId"`
}

var updateBetreiberSchema = z.Struct(z.Shape{
	"Vereinsname":  z.String().Min(1, z.Message("Vereinsname ist erforderlich")).Required(),
	"Strasse":      z.String().Min(1, z.Message("Straße ist erforderlich")).Required(),
	"Plz":          z.String().Min(1, z.Message("PLZ ist erforderlich")).Required(),
	"Ort":          z.String().Min(1, z.Message("Ort ist erforderlich")).Required(),
	"Steuernummer": z.Ptr(z.String()),
	"UstID":        z.Ptr(z.String()),
})

func (h *CommandHandler) UpdateBetreiberHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body updateBetreiberRequest
		if !helper.ReadAndValidateBody(w, r, &body, updateBetreiberSchema) {
			return
		}

		b, err := betreiber.NewBetreiber(body.Vereinsname, body.Strasse, body.Plz, body.Ort, body.Steuernummer, body.UstID)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		if err := h.Command.UpdateBetreiber(r.Context(), b); err != nil {
			helper.SendServerError(w)
			return
		}
		helper.SendEmptyResponse(w)
	}
}

// SetzeElsterMeldungHandler markiert die ELSTER-Kassenmeldung als erledigt
// (serverseitig auf das aktuelle Datum). Kein Request-Body.
func (h *CommandHandler) SetzeElsterMeldungHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h.Command.SetzeElsterMeldung(r.Context()); err != nil {
			helper.SendServerError(w)
			return
		}
		helper.SendEmptyResponse(w)
	}
}

// NimmElsterMeldungZurueckHandler setzt die ELSTER-Kassenmeldung zurück (NULL),
// damit ein Fehlklick korrigierbar bleibt. Kein Request-Body.
func (h *CommandHandler) NimmElsterMeldungZurueckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h.Command.NimmElsterMeldungZurueck(r.Context()); err != nil {
			helper.SendServerError(w)
			return
		}
		helper.SendEmptyResponse(w)
	}
}
