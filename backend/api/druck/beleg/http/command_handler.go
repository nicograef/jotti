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
	KassenbelegDrucken(ctx context.Context, tischID int, zahlungID string, verkaufID string, stornierungID string) (application.BelegStatus, error)
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
		body := belegDruckenRequest{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		hasTisch := body.TischID != nil
		hasZahlung := body.ZahlungID != nil
		hasVerkauf := body.VerkaufID != nil
		hasStornierung := body.StornierungID != nil

		switch {
		case hasVerkauf && !hasTisch && !hasZahlung:
			verkaufID := *body.VerkaufID
			stornierungID := ""

			if hasStornierung {
				cmd := belegDruckenStornoRequest{VerkaufID: verkaufID, StornierungID: *body.StornierungID}
				if issues := belegDruckenStornoSchema.Validate(&cmd); issues != nil {
					helper.SendClientError(w, "validation_error", z.Issues.FlattenAndCollect(issues))
					return
				}
				stornierungID = cmd.StornierungID
			} else {
				cmd := belegDruckenVerkaufRequest{VerkaufID: verkaufID}
				if issues := belegDruckenVerkaufSchema.Validate(&cmd); issues != nil {
					helper.SendClientError(w, "validation_error", z.Issues.FlattenAndCollect(issues))
					return
				}
			}

			status, err := h.Command.KassenbelegDrucken(r.Context(), 0, "", verkaufID, stornierungID)
			if err != nil {
				switch {
				case errors.Is(err, application.ErrKasseWirdAbgeschlossen):
					helper.SendConflict(w, "kasse_wird_abgeschlossen")
				case errors.Is(err, application.ErrKasseNichtGeoeffnet):
					helper.SendConflict(w, "kasse_nicht_geoeffnet")
				default:
					helper.MapError(w, err, map[error]string{
						application.ErrVerkaufNichtGefunden:                "verkauf_not_found",
						application.ErrStornierungNichtGefunden:            "stornierung_not_found",
						application.ErrKassenbelegDruckerNichtKonfiguriert: "kassenbeleg_drucker_nicht_konfiguriert",
					})
				}
				return
			}

			helper.SendResponse(w, belegDruckenResponse{Status: status})

		case hasTisch && hasStornierung && !hasZahlung && !hasVerkauf:
			cmd := belegDruckenTischStornoRequest{TischID: *body.TischID, StornierungID: *body.StornierungID}
			if issues := belegDruckenTischStornoSchema.Validate(&cmd); issues != nil {
				helper.SendClientError(w, "validation_error", z.Issues.FlattenAndCollect(issues))
				return
			}

			status, err := h.Command.KassenbelegDrucken(r.Context(), cmd.TischID, "", "", cmd.StornierungID)
			if err != nil {
				switch {
				case errors.Is(err, application.ErrKasseWirdAbgeschlossen):
					helper.SendConflict(w, "kasse_wird_abgeschlossen")
				case errors.Is(err, application.ErrKasseNichtGeoeffnet):
					helper.SendConflict(w, "kasse_nicht_geoeffnet")
				default:
					helper.MapError(w, err, map[error]string{
						application.ErrStornierungNichtGefunden:            "stornierung_not_found",
						application.ErrZahlungNichtGefunden:                "zahlung_not_found",
						application.ErrKassenbelegDruckerNichtKonfiguriert: "kassenbeleg_drucker_nicht_konfiguriert",
					})
				}
				return
			}

			helper.SendResponse(w, belegDruckenResponse{Status: status})

		case hasTisch && hasZahlung && !hasStornierung && !hasVerkauf:
			cmd := belegDruckenZahlungRequest{TischID: *body.TischID, ZahlungID: *body.ZahlungID}
			if issues := belegDruckenZahlungSchema.Validate(&cmd); issues != nil {
				helper.SendClientError(w, "validation_error", z.Issues.FlattenAndCollect(issues))
				return
			}

			status, err := h.Command.KassenbelegDrucken(r.Context(), cmd.TischID, cmd.ZahlungID, "", "")
			if err != nil {
				switch {
				case errors.Is(err, application.ErrKasseWirdAbgeschlossen):
					helper.SendConflict(w, "kasse_wird_abgeschlossen")
				case errors.Is(err, application.ErrKasseNichtGeoeffnet):
					helper.SendConflict(w, "kasse_nicht_geoeffnet")
				default:
					helper.MapError(w, err, map[error]string{
						application.ErrZahlungNichtGefunden:                "zahlung_not_found",
						application.ErrKassenbelegDruckerNichtKonfiguriert: "kassenbeleg_drucker_nicht_konfiguriert",
					})
				}
				return
			}

			helper.SendResponse(w, belegDruckenResponse{Status: status})

		default:
			helper.SendClientError(w, "validation_error", map[string][]string{
				"body": {belegDruckenValidationMessage},
			})
		}
	}
}
