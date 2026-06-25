package http

import (
	"context"
	"errors"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/middleware"
	"github.com/nicograef/jotti/backend/api/table/application"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/product"
	"github.com/nicograef/jotti/backend/domain/table"
)

type command interface {
	TischErstellen(ctx context.Context, name string) (int, error)
	TischAktualisieren(ctx context.Context, id int, name string) error
	TischAktivieren(ctx context.Context, id int) error
	TischDeaktivieren(ctx context.Context, id int) error
	TischLoeschen(ctx context.Context, id int) error
	BestellungAufnehmen(ctx context.Context, userID int, userName string, tischID int, positionen []application.BestellPositionInput, kommentar string) error
	BestellungUmbuchen(ctx context.Context, userID int, userName string, quellTischID int, zielTischID int, positionen []kasse.PositionRef) error
	ZahlungKassieren(ctx context.Context, userID int, userName string, tischID int, positionen []kasse.PositionRef, kommentar string) error
	StornierungErteilen(ctx context.Context, userID int, userName string, tischID int, positionen []kasse.PositionRef, kommentar string) error
	AusgabeBestaetigen(ctx context.Context, userID int, userName string, tischID int, positionen []kasse.PositionRef, kommentar string) error
	AuszahlungLeisten(ctx context.Context, userID int, userName string, tischID int, betragCents int, kommentar string) error
	KassenbelegDrucken(ctx context.Context, tischID int, zahlungID string, verkaufID string, stornierungID string) error
	FavoritHinzufuegen(ctx context.Context, userID, tischID int) error
	FavoritEntfernen(ctx context.Context, userID, tischID int) error
}

type CommandHandler struct {
	Command command
}

type createTischRequest struct {
	Name string `json:"name"`
}

var createTischSchema = z.Struct(z.Shape{
	"Name": table.TischNameSchema.Required(),
})

type createTischResponse struct {
	ID int `json:"id"`
}

func (h *CommandHandler) TischErstellenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := createTischRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, createTischSchema) {
			return
		}

		id, err := h.Command.TischErstellen(r.Context(), body.Name)
		if err != nil {
			helper.MapError(w, err, map[error]string{
				application.ErrTischAlreadyExists: "tisch_already_exists",
				application.ErrInvalidTischData:   "invalid_tisch_data",
			})
			return
		}

		helper.SendResponse(w, createTischResponse{ID: id})
	}
}

type updateTischRequest struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var updateTischSchema = z.Struct(z.Shape{
	"ID":   table.TischIDSchema.Required(),
	"Name": table.TischNameSchema.Required(),
})

func (h *CommandHandler) TischAktualisierenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := updateTischRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, updateTischSchema) {
			return
		}

		err := h.Command.TischAktualisieren(r.Context(), body.ID, body.Name)
		if err != nil {
			helper.MapError(w, err, map[error]string{
				application.ErrTischNotFound:    "tisch_not_found",
				application.ErrInvalidTischData: "invalid_tisch_data",
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}

type favoritRequest struct {
	TischID int `json:"tischId"`
}

var favoritSchema = z.Struct(z.Shape{
	"TischID": table.TischIDSchema.Required(),
})

func (h *CommandHandler) FavoritHinzufuegenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := favoritRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, favoritSchema) {
			return
		}

		userID, _, ok := middleware.UserFromContext(r.Context())
		if !ok {
			helper.SendServerError(w)
			return
		}

		err := h.Command.FavoritHinzufuegen(r.Context(), userID, body.TischID)
		if err != nil {
			helper.MapError(w, err, map[error]string{
				application.ErrTischNotFound:  "tisch_not_found",
				application.ErrTischNotActive: "tisch_not_active",
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}

func (h *CommandHandler) FavoritEntfernenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := favoritRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, favoritSchema) {
			return
		}

		userID, _, ok := middleware.UserFromContext(r.Context())
		if !ok {
			helper.SendServerError(w)
			return
		}

		err := h.Command.FavoritEntfernen(r.Context(), userID, body.TischID)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendEmptyResponse(w)
	}
}

type activateTischRequest struct {
	ID int `json:"id"`
}

var activateTischSchema = z.Struct(z.Shape{
	"ID": table.TischIDSchema.Required(),
})

func (h *CommandHandler) TischAktivierenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := activateTischRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, activateTischSchema) {
			return
		}

		err := h.Command.TischAktivieren(r.Context(), body.ID)
		if err != nil {
			helper.MapError(w, err, map[error]string{
				application.ErrTischNotFound: "tisch_not_found",
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}

type deactivateTischRequest struct {
	ID int `json:"id"`
}

var deactivateTischSchema = z.Struct(z.Shape{
	"ID": table.TischIDSchema.Required(),
})

func (h *CommandHandler) TischDeaktivierenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := deactivateTischRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, deactivateTischSchema) {
			return
		}

		err := h.Command.TischDeaktivieren(r.Context(), body.ID)
		if err != nil {
			helper.MapError(w, err, map[error]string{
				application.ErrTischNotFound: "tisch_not_found",
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}

type deleteTischRequest struct {
	ID int `json:"id"`
}

var deleteTischSchema = z.Struct(z.Shape{
	"ID": table.TischIDSchema.Required(),
})

func (h *CommandHandler) TischLoeschenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := deleteTischRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, deleteTischSchema) {
			return
		}

		err := h.Command.TischLoeschen(r.Context(), body.ID)
		if err != nil {
			helper.MapError(w, err, map[error]string{
				application.ErrTischNotFound: "tisch_not_found",
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
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
	TischID    int                    `json:"tischId"`
	Positionen []bestellPositionInput `json:"positionen"`
	Kommentar  string                 `json:"kommentar"`
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
	"ProduktID":  product.IDSchema.Required(),
	"VarianteID": product.IDSchema.Required(),
	"Menge":      z.Int().GTE(1).Required(),
})

var positionRefRequestSchema = z.Struct(z.Shape{
	"PositionID": z.String().UUID().Required(),
	"Menge":      z.Int().GTE(1).Required(),
})

var bestellungAufnehmenSchema = z.Struct(z.Shape{
	"TischID":    table.TischIDSchema.Required(),
	"Positionen": z.Slice(bestellPositionInputSchema).Min(1).Required(),
	"Kommentar":  z.String().Max(100),
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
		err := h.Command.BestellungAufnehmen(r.Context(), userID, userName, body.TischID, toBestellPositionInputs(body.Positionen), body.Kommentar)
		if err != nil {
			switch {
			case errors.Is(err, application.ErrConflict):
				helper.SendConflictError(w)
			case errors.Is(err, application.ErrKasseNichtGeoeffnet):
				helper.SendConflict(w, "kasse_nicht_geoeffnet")
			default:
				helper.MapError(w, err, map[error]string{
					application.ErrTischNotFound:   "tisch_not_found",
					application.ErrTischNotActive:  "tisch_not_active",
					application.ErrProduktNotFound: "produkt_not_found",
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
	"TischID":    table.TischIDSchema.Required(),
	"Positionen": z.Slice(positionRefRequestSchema).Min(1).Required(),
	"Kommentar":  z.String().Max(100),
})

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
	"TischID":   table.TischIDSchema.Required(),
	"ZahlungID": z.String().UUID().Required(),
})

var belegDruckenTischStornoSchema = z.Struct(z.Shape{
	"TischID":       table.TischIDSchema.Required(),
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

			err := h.Command.KassenbelegDrucken(r.Context(), 0, "", verkaufID, stornierungID)
			if err != nil {
				if errors.Is(err, application.ErrKasseNichtGeoeffnet) {
					helper.SendConflict(w, "kasse_nicht_geoeffnet")
				} else {
					helper.MapError(w, err, map[error]string{
						application.ErrVerkaufNichtGefunden:                "verkauf_not_found",
						application.ErrStornierungNichtGefunden:            "stornierung_not_found",
						application.ErrKassenbelegDruckerNichtKonfiguriert: "kassenbeleg_drucker_nicht_konfiguriert",
					})
				}
				return
			}

			helper.SendEmptyResponse(w)

		case hasTisch && hasStornierung && !hasZahlung && !hasVerkauf:
			cmd := belegDruckenTischStornoRequest{TischID: *body.TischID, StornierungID: *body.StornierungID}
			if issues := belegDruckenTischStornoSchema.Validate(&cmd); issues != nil {
				helper.SendClientError(w, "validation_error", z.Issues.FlattenAndCollect(issues))
				return
			}

			err := h.Command.KassenbelegDrucken(r.Context(), cmd.TischID, "", "", cmd.StornierungID)
			if err != nil {
				if errors.Is(err, application.ErrKasseNichtGeoeffnet) {
					helper.SendConflict(w, "kasse_nicht_geoeffnet")
				} else {
					helper.MapError(w, err, map[error]string{
						application.ErrStornierungNichtGefunden:            "stornierung_not_found",
						application.ErrZahlungNichtGefunden:                "zahlung_not_found",
						application.ErrKassenbelegDruckerNichtKonfiguriert: "kassenbeleg_drucker_nicht_konfiguriert",
					})
				}
				return
			}

			helper.SendEmptyResponse(w)

		case hasTisch && hasZahlung && !hasStornierung && !hasVerkauf:
			cmd := belegDruckenZahlungRequest{TischID: *body.TischID, ZahlungID: *body.ZahlungID}
			if issues := belegDruckenZahlungSchema.Validate(&cmd); issues != nil {
				helper.SendClientError(w, "validation_error", z.Issues.FlattenAndCollect(issues))
				return
			}

			err := h.Command.KassenbelegDrucken(r.Context(), cmd.TischID, cmd.ZahlungID, "", "")
			if err != nil {
				if errors.Is(err, application.ErrKasseNichtGeoeffnet) {
					helper.SendConflict(w, "kasse_nicht_geoeffnet")
				} else {
					helper.MapError(w, err, map[error]string{
						application.ErrZahlungNichtGefunden:                "zahlung_not_found",
						application.ErrKassenbelegDruckerNichtKonfiguriert: "kassenbeleg_drucker_nicht_konfiguriert",
					})
				}
				return
			}

			helper.SendEmptyResponse(w)

		default:
			helper.SendClientError(w, "validation_error", map[string][]string{
				"body": {belegDruckenValidationMessage},
			})
		}
	}
}

type stornierungErteilenRequest struct {
	TischID    int                  `json:"tischId"`
	Positionen []positionRefRequest `json:"positionen"`
	Kommentar  string               `json:"kommentar"`
}

type bestellungUmbuchenRequest struct {
	QuellTischID int                  `json:"quellTischId"`
	ZielTischID  int                  `json:"zielTischId"`
	Positionen   []positionRefRequest `json:"positionen"`
}

var stornierungErteilenSchema = z.Struct(z.Shape{
	"TischID":    table.TischIDSchema.Required(),
	"Positionen": z.Slice(positionRefRequestSchema).Min(1).Required(),
	"Kommentar":  z.String().Min(3).Max(100).Required(),
})

var bestellungUmbuchenSchema = z.Struct(z.Shape{
	"QuellTischID": table.TischIDSchema.Required(),
	"ZielTischID":  table.TischIDSchema.Required(),
	"Positionen":   z.Slice(positionRefRequestSchema).Min(1).Required(),
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

		err := h.Command.BestellungUmbuchen(r.Context(), userID, userName, body.QuellTischID, body.ZielTischID, toPositionRefs(body.Positionen))
		if err != nil {
			switch {
			case errors.Is(err, application.ErrConflict):
				helper.SendConflictError(w)
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

type ausgabeBestaetigenRequest struct {
	TischID    int                  `json:"tischId"`
	Positionen []positionRefRequest `json:"positionen"`
	Kommentar  string               `json:"kommentar"`
}

type auszahlungLeistenRequest struct {
	TischID     int    `json:"tischId"`
	BetragCents int    `json:"betragCents"`
	Kommentar   string `json:"kommentar"`
}

var ausgabeBestaetigenSchema = z.Struct(z.Shape{
	"TischID":    table.TischIDSchema.Required(),
	"Positionen": z.Slice(positionRefRequestSchema).Min(1).Required(),
	"Kommentar":  z.String().Max(100),
})

func (h *CommandHandler) AusgabeBestaetigenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := ausgabeBestaetigenRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, ausgabeBestaetigenSchema) {
			return
		}

		userID, userName, ok := middleware.UserFromContext(r.Context())
		if !ok {
			helper.SendServerError(w)
			return
		}
		err := h.Command.AusgabeBestaetigen(r.Context(), userID, userName, body.TischID, toPositionRefs(body.Positionen), body.Kommentar)
		if err != nil {
			switch {
			case errors.Is(err, application.ErrConflict):
				helper.SendConflictError(w)
			case errors.Is(err, application.ErrKasseNichtGeoeffnet):
				helper.SendConflict(w, "kasse_nicht_geoeffnet")
			default:
				helper.MapError(w, err, map[error]string{
					application.ErrTischNotFound:          "tisch_not_found",
					application.ErrTischNotActive:         "tisch_not_active",
					application.ErrPositionNichtAusgebbar: "position_nicht_ausgebbar",
				})
			}
			return
		}

		helper.SendEmptyResponse(w)
	}
}

var auszahlungLeistenSchema = z.Struct(z.Shape{
	"TischID":     table.TischIDSchema.Required(),
	"BetragCents": z.Int().GTE(1).Required(),
	"Kommentar":   z.String().Min(3).Max(100).Required(),
})

func (h *CommandHandler) AuszahlungLeistenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := auszahlungLeistenRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, auszahlungLeistenSchema) {
			return
		}

		userID, userName, ok := middleware.UserFromContext(r.Context())
		if !ok {
			helper.SendServerError(w)
			return
		}
		err := h.Command.AuszahlungLeisten(r.Context(), userID, userName, body.TischID, body.BetragCents, body.Kommentar)
		if err != nil {
			switch {
			case errors.Is(err, application.ErrConflict):
				helper.SendConflictError(w)
			case errors.Is(err, application.ErrKasseNichtGeoeffnet):
				helper.SendConflict(w, "kasse_nicht_geoeffnet")
			default:
				helper.MapError(w, err, map[error]string{
					application.ErrTischNotFound:  "tisch_not_found",
					application.ErrTischNotActive: "tisch_not_active",
				})
			}
			return
		}

		helper.SendEmptyResponse(w)
	}
}
