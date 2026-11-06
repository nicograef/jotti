package http

import (
	"context"
	"errors"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api/druck/beleg/application"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/domain/tisch"
)

type command interface {
	KassenbelegDrucken(ctx context.Context, cmd application.KassenbelegDruckenCommand) (application.BelegStatus, error)
}

type CommandHandler struct {
	Command command
}

type belegDruckenRequest struct {
	TischID       *int    `json:"tischId"`
	ZahlungID     *string `json:"zahlungId"`
	VerkaufID     *string `json:"verkaufId"`
	RueckgabeID   *string `json:"rueckgabeId"`
	StornierungID *string `json:"stornierungId"`
}

type belegDruckenZahlungRequest struct {
	TischID   int    `json:"tischId"`
	ZahlungID string `json:"zahlungId"`
}

type belegDruckenTischStornoRequest struct {
	TischID       int    `json:"tischId"`
	StornierungID string `json:"stornierungId"`
}

type belegDruckenVerkaufRequest struct {
	VerkaufID string `json:"verkaufId"`
}

type belegDruckenStornoRequest struct {
	VerkaufID     string `json:"verkaufId"`
	StornierungID string `json:"stornierungId"`
}

type belegDruckenPfandRequest struct {
	RueckgabeID string `json:"rueckgabeId"`
}

type belegDruckenPfandStornoRequest struct {
	RueckgabeID   string `json:"rueckgabeId"`
	StornierungID string `json:"stornierungId"`
}

var belegDruckenZahlungSchema = z.Struct(z.Shape{
	"TischID":   tisch.TischIDSchema.Required(),
	"ZahlungID": z.String().UUID().Required(),
})

var belegDruckenTischStornoSchema = z.Struct(z.Shape{
	"TischID":       tisch.TischIDSchema.Required(),
	"StornierungID": z.String().UUID().Required(),
})

var belegDruckenVerkaufSchema = z.Struct(z.Shape{
	"VerkaufID": z.String().UUID().Required(),
})

var belegDruckenStornoSchema = z.Struct(z.Shape{
	"VerkaufID":     z.String().UUID().Required(),
	"StornierungID": z.String().UUID().Required(),
})

var belegDruckenPfandSchema = z.Struct(z.Shape{
	"RueckgabeID": z.String().UUID().Required(),
})

var belegDruckenPfandStornoSchema = z.Struct(z.Shape{
	"RueckgabeID":   z.String().UUID().Required(),
	"StornierungID": z.String().UUID().Required(),
})

const belegDruckenValidationMessage = "entweder tischId+zahlungId, tischId+stornierungId, verkaufId, verkaufId+stornierungId, rueckgabeId oder rueckgabeId+stornierungId senden"

type belegDruckenResponse struct {
	Status application.BelegStatus `json:"status"`
}

func (h *CommandHandler) KassenbelegDruckenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cmd, ok := readKassenbelegCommand(w, r)
		if !ok {
			return
		}

		status, err := h.Command.KassenbelegDrucken(r.Context(), cmd)
		if err != nil {
			switch {
			case errors.Is(err, application.ErrKasseWirdAbgeschlossen):
				helper.SendConflict(w, "kasse_wird_abgeschlossen")
			case errors.Is(err, application.ErrKasseNichtGeoeffnet):
				helper.SendConflict(w, "kasse_nicht_geoeffnet")
			case errors.Is(err, application.ErrRueckgabeNichtGefunden):
				helper.SendNotFound(w, "rueckgabe_not_found")
			case errors.Is(err, application.ErrStornierungNichtGefunden):
				helper.SendNotFound(w, "stornierung_not_found")
			default:
				// Ein gemeinsames Mapping für alle Formen: jede Form kann nur ihre
				// Teilmenge dieser Fehler liefern, und die Codes sind formunabhängig gleich.
				helper.MapError(w, err, []helper.ErrorCode{
					{Err: application.ErrVerkaufNichtGefunden, Code: "verkauf_not_found"},
					{Err: application.ErrZahlungNichtGefunden, Code: "zahlung_not_found"},
					{Err: application.ErrKassenbelegDruckerNichtKonfiguriert, Code: "kassenbeleg_drucker_nicht_konfiguriert"},
				})
			}
			return
		}

		helper.SendResponse(w, belegDruckenResponse{Status: status})
	}
}

// readKassenbelegCommand accepts the body forms of docs/handbuch.md §4.5.
// On invalid input it sends the client error response and returns ok=false.
func readKassenbelegCommand(w http.ResponseWriter, r *http.Request) (application.KassenbelegDruckenCommand, bool) {
	body := belegDruckenRequest{}
	if !helper.ReadBody(w, r, &body) {
		return application.KassenbelegDruckenCommand{}, false
	}

	hasTisch := body.TischID != nil
	hasZahlung := body.ZahlungID != nil
	hasVerkauf := body.VerkaufID != nil
	hasRueckgabe := body.RueckgabeID != nil
	hasStornierung := body.StornierungID != nil

	switch {
	case hasRueckgabe && !hasTisch && !hasZahlung && !hasVerkauf:
		if hasStornierung {
			req := belegDruckenPfandStornoRequest{RueckgabeID: *body.RueckgabeID, StornierungID: *body.StornierungID}
			if !validateBody(w, belegDruckenPfandStornoSchema, &req) {
				return application.KassenbelegDruckenCommand{}, false
			}
			return application.KassenbelegDruckenCommand{RueckgabeID: req.RueckgabeID, StornierungID: req.StornierungID}, true
		}

		req := belegDruckenPfandRequest{RueckgabeID: *body.RueckgabeID}
		if !validateBody(w, belegDruckenPfandSchema, &req) {
			return application.KassenbelegDruckenCommand{}, false
		}
		return application.KassenbelegDruckenCommand{RueckgabeID: req.RueckgabeID}, true

	case hasVerkauf && !hasTisch && !hasZahlung && !hasRueckgabe:
		if hasStornierung {
			req := belegDruckenStornoRequest{VerkaufID: *body.VerkaufID, StornierungID: *body.StornierungID}
			if !validateBody(w, belegDruckenStornoSchema, &req) {
				return application.KassenbelegDruckenCommand{}, false
			}
			return application.KassenbelegDruckenCommand{VerkaufID: req.VerkaufID, StornierungID: req.StornierungID}, true
		}

		req := belegDruckenVerkaufRequest{VerkaufID: *body.VerkaufID}
		if !validateBody(w, belegDruckenVerkaufSchema, &req) {
			return application.KassenbelegDruckenCommand{}, false
		}
		return application.KassenbelegDruckenCommand{VerkaufID: req.VerkaufID}, true

	case hasTisch && hasStornierung && !hasZahlung && !hasVerkauf && !hasRueckgabe:
		req := belegDruckenTischStornoRequest{TischID: *body.TischID, StornierungID: *body.StornierungID}
		if !validateBody(w, belegDruckenTischStornoSchema, &req) {
			return application.KassenbelegDruckenCommand{}, false
		}
		return application.KassenbelegDruckenCommand{TischID: req.TischID, StornierungID: req.StornierungID}, true

	case hasTisch && hasZahlung && !hasStornierung && !hasVerkauf && !hasRueckgabe:
		req := belegDruckenZahlungRequest{TischID: *body.TischID, ZahlungID: *body.ZahlungID}
		if !validateBody(w, belegDruckenZahlungSchema, &req) {
			return application.KassenbelegDruckenCommand{}, false
		}
		return application.KassenbelegDruckenCommand{TischID: req.TischID, ZahlungID: req.ZahlungID}, true

	default:
		helper.SendClientError(w, "validation_error", map[string][]string{
			"body": {belegDruckenValidationMessage},
		})
		return application.KassenbelegDruckenCommand{}, false
	}
}

func validateBody(w http.ResponseWriter, schema *z.StructSchema, body any) bool {
	if issues := schema.Validate(body); issues != nil {
		helper.SendClientError(w, "validation_error", z.Issues.FlattenAndCollect(issues))
		return false
	}
	return true
}
