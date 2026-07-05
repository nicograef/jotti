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

const belegDruckenValidationMessage = "entweder tischId+zahlungId, tischId+stornierungId, verkaufId oder verkaufId+stornierungId senden"

// belegDruckenResponse meldet den Beleg-Status: "eingereiht" (Druckauftrag
// angelegt) oder "ausstehend" (TSE-Signatur liegt noch nicht vor; die UI ruft
// denselben Endpunkt erneut auf).
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
			default:
				// Ein gemeinsames Mapping für alle vier Formen: jede Form kann nur ihre
				// Teilmenge dieser Fehler liefern, und die Codes sind formunabhängig gleich.
				helper.MapError(w, err, map[error]string{
					application.ErrVerkaufNichtGefunden:                "verkauf_not_found",
					application.ErrStornierungNichtGefunden:            "stornierung_not_found",
					application.ErrZahlungNichtGefunden:                "zahlung_not_found",
					application.ErrKassenbelegDruckerNichtKonfiguriert: "kassenbeleg_drucker_nicht_konfiguriert",
				})
			}
			return
		}

		helper.SendResponse(w, belegDruckenResponse{Status: status})
	}
}

// readKassenbelegCommand liest die Beleg-Anfrage, bestimmt anhand der gesetzten
// Felder eine der vier gültigen Body-Formen und validiert deren Pflichtfelder.
// Die eigentliche Auswahl, welcher Beleg daraus entsteht, liegt in der
// Application-Schicht (KassenbelegDrucken). Bei ungültiger Kombination oder
// ungültigen Feldern sendet die Funktion die Client-Fehlerantwort und liefert
// ok=false.
func readKassenbelegCommand(w http.ResponseWriter, r *http.Request) (application.KassenbelegDruckenCommand, bool) {
	body := belegDruckenRequest{}
	if !helper.ReadBody(w, r, &body) {
		return application.KassenbelegDruckenCommand{}, false
	}

	hasTisch := body.TischID != nil
	hasZahlung := body.ZahlungID != nil
	hasVerkauf := body.VerkaufID != nil
	hasStornierung := body.StornierungID != nil

	switch {
	case hasVerkauf && !hasTisch && !hasZahlung:
		if hasStornierung {
			req := belegDruckenStornoRequest{VerkaufID: *body.VerkaufID, StornierungID: *body.StornierungID}
			if issues := belegDruckenStornoSchema.Validate(&req); issues != nil {
				helper.SendClientError(w, "validation_error", z.Issues.FlattenAndCollect(issues))
				return application.KassenbelegDruckenCommand{}, false
			}
			return application.KassenbelegDruckenCommand{VerkaufID: req.VerkaufID, StornierungID: req.StornierungID}, true
		}

		req := belegDruckenVerkaufRequest{VerkaufID: *body.VerkaufID}
		if issues := belegDruckenVerkaufSchema.Validate(&req); issues != nil {
			helper.SendClientError(w, "validation_error", z.Issues.FlattenAndCollect(issues))
			return application.KassenbelegDruckenCommand{}, false
		}
		return application.KassenbelegDruckenCommand{VerkaufID: req.VerkaufID}, true

	case hasTisch && hasStornierung && !hasZahlung && !hasVerkauf:
		req := belegDruckenTischStornoRequest{TischID: *body.TischID, StornierungID: *body.StornierungID}
		if issues := belegDruckenTischStornoSchema.Validate(&req); issues != nil {
			helper.SendClientError(w, "validation_error", z.Issues.FlattenAndCollect(issues))
			return application.KassenbelegDruckenCommand{}, false
		}
		return application.KassenbelegDruckenCommand{TischID: req.TischID, StornierungID: req.StornierungID}, true

	case hasTisch && hasZahlung && !hasStornierung && !hasVerkauf:
		req := belegDruckenZahlungRequest{TischID: *body.TischID, ZahlungID: *body.ZahlungID}
		if issues := belegDruckenZahlungSchema.Validate(&req); issues != nil {
			helper.SendClientError(w, "validation_error", z.Issues.FlattenAndCollect(issues))
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
