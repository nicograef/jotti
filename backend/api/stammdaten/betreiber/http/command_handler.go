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
	"Vereinsname":  betreiber.VereinsnameSchema,
	"Strasse":      betreiber.StrasseSchema,
	"Plz":          betreiber.PlzSchema,
	"Ort":          betreiber.OrtSchema,
	"Steuernummer": z.Ptr(betreiber.SteuernummerSchema),
	"UstID":        z.Ptr(betreiber.UstIDSchema),
})

func (h *CommandHandler) UpdateBetreiberHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body updateBetreiberRequest
		if !helper.ReadAndValidateBody(w, r, &body, updateBetreiberSchema) {
			return
		}

		b, err := betreiber.NewBetreiber(body.Vereinsname, body.Strasse, body.Plz, body.Ort, body.Steuernummer, body.UstID)
		if err != nil {
			// Der Zweig ist defensiv: updateBetreiberSchema prüft dieselben
			// Feld-Schemas, die der Konstruktor erneut prüft, also lehnt er einen
			// angenommenen Body nicht ab. Lehnt er doch ab, liegt es an der
			// Eingabe und nicht am Server — deshalb 400.
			helper.SendClientError(w, "validation_error", nil)
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
