package http

import (
	"context"
	"errors"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/kasse/tischgeschaeft/application"
	"github.com/nicograef/jotti/backend/api/middleware"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/domain/tisch"
)

type command interface {
	BestellungAufnehmen(ctx context.Context, userID int, userName string, bestellungID string, tischID int, positionen []application.BestellPositionInput, kommentar string) error
	BestellungUmbuchen(ctx context.Context, userID int, userName string, quellTischID int, zielTischID int, positionen []kasse.PositionRef, benutzerKommentar string) error
	ZahlungKassieren(ctx context.Context, userID int, userName string, tischID int, positionen []kasse.PositionRef, kommentar string) error
	StornierungErteilen(ctx context.Context, userID int, userName string, tischID int, positionen []kasse.PositionRef, kommentar string) error
}

type CommandHandler struct {
	Command command
}

type bestellPositionInput struct {
	ProduktID  int `json:"produktId"`
	VarianteID int `json:"varianteId"`
	Menge      int `json:"menge"`
}

func toBestellPositionInput(p bestellPositionInput) application.BestellPositionInput {
	return application.BestellPositionInput{
		ProduktID:  p.ProduktID,
		VarianteID: p.VarianteID,
		Menge:      p.Menge,
	}
}

func toBestellPositionInputs(positionen []bestellPositionInput) []application.BestellPositionInput {
	out := make([]application.BestellPositionInput, len(positionen))
	for i, p := range positionen {
		out[i] = toBestellPositionInput(p)
	}
	return out
}

type bestellungAufnehmenRequest struct {
	BestellungID string                 `json:"bestellungId"`
	TischID      int                    `json:"tischId"`
	Positionen   []bestellPositionInput `json:"positionen"`
	Kommentar    string                 `json:"kommentar"`
}

type positionRefRequest struct {
	PositionID string `json:"positionId"`
	Menge      int    `json:"menge"`
}

func toPositionRef(p positionRefRequest) kasse.PositionRef {
	return kasse.PositionRef{
		PositionID: p.PositionID,
		Menge:      p.Menge,
	}
}

func toPositionRefs(refs []positionRefRequest) []kasse.PositionRef {
	out := make([]kasse.PositionRef, len(refs))
	for i, ref := range refs {
		out[i] = toPositionRef(ref)
	}
	return out
}

var bestellPositionInputSchema = z.Struct(z.Shape{
	"ProduktID":  produkt.IDSchema.Required(),
	"VarianteID": produkt.IDSchema.Required(),
	"Menge":      z.Int().GTE(1).Required(),
})

var positionRefRequestSchema = z.Struct(z.Shape{
	"PositionID": z.String().UUID().Required(),
	"Menge":      z.Int().GTE(1).Required(),
})

var bestellungAufnehmenSchema = z.Struct(z.Shape{
	"BestellungID": z.String().UUID().Required(),
	"TischID":      tisch.TischIDSchema.Required(),
	"Positionen":   z.Slice(bestellPositionInputSchema).Min(1).Required(),
	"Kommentar":    z.String().Max(100),
})

func (h *CommandHandler) BestellungAufnehmenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := bestellungAufnehmenRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, bestellungAufnehmenSchema) {
			return
		}

		userID, userName, ok := middleware.UserFromContext(r.Context())
		if !ok {
			helper.SendServerError(w)
			return
		}
		err := h.Command.BestellungAufnehmen(r.Context(), userID, userName, body.BestellungID, body.TischID, toBestellPositionInputs(body.Positionen), body.Kommentar)
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
					application.ErrTischNotFound:      "tisch_not_found",
					application.ErrTischNotActive:     "tisch_not_active",
					application.ErrProduktNotFound:    "produkt_not_found",
					application.ErrVarianteNichtAktiv: "variante_nicht_aktiv",
				})
			}
			return
		}

		helper.SendEmptyResponse(w)
	}
}

type zahlungKassierenRequest struct {
	TischID    int                  `json:"tischId"`
	Positionen []positionRefRequest `json:"positionen"`
	Kommentar  string               `json:"kommentar"`
}

var zahlungKassierenSchema = z.Struct(z.Shape{
	"TischID":    tisch.TischIDSchema.Required(),
	"Positionen": z.Slice(positionRefRequestSchema).Min(1).Required(),
	"Kommentar":  z.String().Max(100),
})

func (h *CommandHandler) ZahlungKassierenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := zahlungKassierenRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, zahlungKassierenSchema) {
			return
		}

		userID, userName, ok := middleware.UserFromContext(r.Context())
		if !ok {
			helper.SendServerError(w)
			return
		}
		err := h.Command.ZahlungKassieren(r.Context(), userID, userName, body.TischID, toPositionRefs(body.Positionen), body.Kommentar)
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
					application.ErrTischNotFound:          "tisch_not_found",
					application.ErrTischNotActive:         "tisch_not_active",
					application.ErrPositionNichtBezahlbar: "position_nicht_bezahlbar",
				})
			}
			return
		}

		helper.SendEmptyResponse(w)
	}
}

type stornierungErteilenRequest struct {
	TischID    int                  `json:"tischId"`
	Positionen []positionRefRequest `json:"positionen"`
	Kommentar  string               `json:"kommentar"`
}

type bestellungUmbuchenRequest struct {
	QuellTischID      int                  `json:"quellTischId"`
	ZielTischID       int                  `json:"zielTischId"`
	Positionen        []positionRefRequest `json:"positionen"`
	BenutzerKommentar string               `json:"benutzerKommentar"`
}

var stornierungErteilenSchema = z.Struct(z.Shape{
	"TischID":    tisch.TischIDSchema.Required(),
	"Positionen": z.Slice(positionRefRequestSchema).Min(1).Required(),
	"Kommentar":  z.String().Min(3).Max(100).Required(),
})

var bestellungUmbuchenSchema = z.Struct(z.Shape{
	"QuellTischID":      tisch.TischIDSchema.Required(),
	"ZielTischID":       tisch.TischIDSchema.Required(),
	"Positionen":        z.Slice(positionRefRequestSchema).Min(1).Required(),
	"BenutzerKommentar": z.String().Max(100),
})

func (h *CommandHandler) StornierungErteilenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := stornierungErteilenRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, stornierungErteilenSchema) {
			return
		}

		userID, userName, ok := middleware.UserFromContext(r.Context())
		if !ok {
			helper.SendServerError(w)
			return
		}
		err := h.Command.StornierungErteilen(r.Context(), userID, userName, body.TischID, toPositionRefs(body.Positionen), body.Kommentar)
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
					application.ErrTischNotFound:            "tisch_not_found",
					application.ErrTischNotActive:           "tisch_not_active",
					application.ErrPositionNichtStornierbar: "position_nicht_stornierbar",
				})
			}
			return
		}

		helper.SendEmptyResponse(w)
	}
}

func (h *CommandHandler) BestellungUmbuchenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := bestellungUmbuchenRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, bestellungUmbuchenSchema) {
			return
		}

		userID, userName, ok := middleware.UserFromContext(r.Context())
		if !ok {
			helper.SendServerError(w)
			return
		}

		err := h.Command.BestellungUmbuchen(r.Context(), userID, userName, body.QuellTischID, body.ZielTischID, toPositionRefs(body.Positionen), body.BenutzerKommentar)
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
					application.ErrTischNotFound:          "tisch_not_found",
					application.ErrTischNotActive:         "tisch_not_active",
					application.ErrPositionNichtUmbuchbar: "position_nicht_umbuchbar",
					application.ErrUmbuchungGleicherTisch: "umbuchung_gleicher_tisch",
				})
			}
			return
		}

		helper.SendEmptyResponse(w)
	}
}
