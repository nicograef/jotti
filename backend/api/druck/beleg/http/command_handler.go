package http

import (
	"context"
	"net/http"
	"slices"

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

// belegDruckenSchema checks the format of the fields a request carries; which fields may be combined
// is belegFormen's job.
var belegDruckenSchema = z.Struct(z.Shape{
	"TischID":       z.Ptr(tisch.TischIDSchema.Required()),
	"ZahlungID":     z.Ptr(z.String().UUID().Required()),
	"VerkaufID":     z.Ptr(z.String().UUID().Required()),
	"RueckgabeID":   z.Ptr(z.String().UUID().Required()),
	"StornierungID": z.Ptr(z.String().UUID().Required()),
})

// belegForm is the set of fields a request carries.
type belegForm struct {
	tisch, zahlung, verkauf, rueckgabe, stornierung bool
}

// belegFormen are the body forms of docs/handbuch.md §4.5.
var belegFormen = []belegForm{
	{tisch: true, zahlung: true},
	{tisch: true, stornierung: true},
	{verkauf: true},
	{verkauf: true, stornierung: true},
	{rueckgabe: true},
	{rueckgabe: true, stornierung: true},
}

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
			// Ein gemeinsames Mapping für alle Formen: jede Form kann nur ihre
			// Teilmenge dieser Fehler liefern, und die Codes sind formunabhängig gleich.
			helper.MapError(w, err, []helper.ErrorCode{
				{Err: application.ErrKasseWirdAbgeschlossen, Code: "kasse_wird_abgeschlossen", Status: http.StatusConflict},
				{Err: application.ErrKasseNichtGeoeffnet, Code: "kasse_nicht_geoeffnet", Status: http.StatusConflict},
				{Err: application.ErrRueckgabeNichtGefunden, Code: "rueckgabe_not_found"},
				{Err: application.ErrStornierungNichtGefunden, Code: "stornierung_not_found"},
				{Err: application.ErrVerkaufNichtGefunden, Code: "verkauf_not_found"},
				{Err: application.ErrZahlungNichtGefunden, Code: "zahlung_not_found"},
				{Err: application.ErrKassenbelegDruckerNichtKonfiguriert, Code: "kassenbeleg_drucker_nicht_konfiguriert"},
			})
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

	form := belegForm{
		tisch:       body.TischID != nil,
		zahlung:     body.ZahlungID != nil,
		verkauf:     body.VerkaufID != nil,
		rueckgabe:   body.RueckgabeID != nil,
		stornierung: body.StornierungID != nil,
	}
	if !slices.Contains(belegFormen, form) {
		helper.SendClientError(w, "validation_error", map[string][]string{
			"body": {belegDruckenValidationMessage},
		})
		return application.KassenbelegDruckenCommand{}, false
	}

	if issues := belegDruckenSchema.Validate(&body); issues != nil {
		helper.SendClientError(w, "validation_error", z.Issues.FlattenAndCollect(issues))
		return application.KassenbelegDruckenCommand{}, false
	}

	return application.KassenbelegDruckenCommand{
		TischID:       wert(body.TischID),
		ZahlungID:     wert(body.ZahlungID),
		VerkaufID:     wert(body.VerkaufID),
		RueckgabeID:   wert(body.RueckgabeID),
		StornierungID: wert(body.StornierungID),
	}, true
}

// wert returns the pointed-to value, the zero value for an absent field.
func wert[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
}
