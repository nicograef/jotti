package http

import (
	"context"
	"net/http"
	"slices"

	z "github.com/Oudwins/zog"
	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/kasse/eingabe"
	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
	enrichmentHTTP "github.com/nicograef/jotti/backend/api/kasse/enrichment/http"
	"github.com/nicograef/jotti/backend/api/kasse/tischgeschaeft/application"
	"github.com/nicograef/jotti/backend/api/middleware"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tisch"
	"github.com/nicograef/jotti/backend/internal/zeichen"
)

type command interface {
	BestellungAufnehmen(ctx context.Context, userID int, userName string, bestellungID string, tischID int, positionen []enrichment.PositionInput, kommentar string) error
	BestellungUmbuchen(ctx context.Context, userID int, userName string, umbuchungID string, quellTischID int, zielTischID int, positionen []kasse.PositionRef, benutzerKommentar string) error
	ZahlungKassieren(ctx context.Context, userID int, userName string, zahlungID string, tischID int, positionen []kasse.PositionRef, kommentar string, pfand *enrichment.PfandEingabe) error
	StornierungErteilen(ctx context.Context, userID int, userName string, stornoID uuid.UUID, tischID int, positionen []kasse.PositionRef, kommentar string) error
}

type CommandHandler struct {
	Command command
}

// kasseFehler opens every Kasse handler's table: a stale write or a closed Kasse is a 409.
var kasseFehler = []helper.ErrorCode{
	{Err: application.ErrConflict, Code: "conflict", Status: http.StatusConflict},
	{Err: application.ErrKasseWirdAbgeschlossen, Code: "kasse_wird_abgeschlossen", Status: http.StatusConflict},
	{Err: application.ErrKasseNichtGeoeffnet, Code: "kasse_nicht_geoeffnet", Status: http.StatusConflict},
	{Err: application.ErrVorgangAbweichend, Code: "vorgang_abweichend", Status: http.StatusConflict},
}

type bestellungAufnehmenRequest struct {
	BestellungID string                              `json:"bestellungId"`
	TischID      int                                 `json:"tischId"`
	Positionen   []eingabe.KommentiertePositionInput `json:"positionen"`
	Kommentar    string                              `json:"kommentar"`
}

var bestellungAufnehmenSchema = z.Struct(z.Shape{
	"BestellungID": z.String().UUID().Required(),
	"TischID":      tisch.TischIDSchema.Required(),
	"Positionen":   z.Slice(eingabe.KommentiertePositionInputSchema).Min(1).Required(),
	"Kommentar":    z.String().TestFunc(zeichen.Max(100)),
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
		err := h.Command.BestellungAufnehmen(r.Context(), userID, userName, body.BestellungID, body.TischID, eingabe.ToKommentiertePositionInputs(body.Positionen), body.Kommentar)
		if err != nil {
			helper.MapError(w, err, slices.Concat(kasseFehler, []helper.ErrorCode{
				{Err: application.ErrTischNotFound, Code: "tisch_not_found"},
				{Err: application.ErrTischNotActive, Code: "tisch_not_active"},
			}, enrichmentHTTP.Fehler))
			return
		}

		helper.SendEmptyResponse(w)
	}
}

// ZahlungID is the client's idempotency key, kept per selection.
type zahlungKassierenRequest struct {
	ZahlungID  string                `json:"zahlungId"`
	TischID    int                   `json:"tischId"`
	Positionen []eingabe.PositionRef `json:"positionen"`
	Kommentar  string                `json:"kommentar"`
	Pfand      *eingabe.PfandRequest `json:"pfand"`
}

var zahlungKassierenSchema = z.Struct(z.Shape{
	"ZahlungID":  z.String().UUID().Required(),
	"TischID":    tisch.TischIDSchema.Required(),
	"Positionen": z.Slice(eingabe.PositionRefSchema).Min(1).Required(),
	"Kommentar":  z.String().TestFunc(zeichen.Max(100)),
	"Pfand":      z.Ptr(eingabe.PfandSchema),
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
		err := h.Command.ZahlungKassieren(r.Context(), userID, userName, body.ZahlungID, body.TischID, eingabe.ToPositionRefs(body.Positionen), body.Kommentar, eingabe.ToPfandEingabe(body.Pfand))
		if err != nil {
			helper.MapError(w, err, slices.Concat(kasseFehler, []helper.ErrorCode{
				{Err: application.ErrPfandrueckgabeAbweichend, Code: "pfandrueckgabe_abweichend", Status: http.StatusConflict},
				{Err: application.ErrTischNotFound, Code: "tisch_not_found"},
				{Err: application.ErrTischNotActive, Code: "tisch_not_active"},
				{Err: application.ErrPositionNichtBezahlbar, Code: "position_nicht_bezahlbar"},
			}, enrichmentHTTP.Fehler))
			return
		}

		helper.SendEmptyResponse(w)
	}
}

// StornoID is the client's idempotency key, kept per attempt; every event ID derives from it.
type stornierungErteilenRequest struct {
	StornoID   string                `json:"stornoId"`
	TischID    int                   `json:"tischId"`
	Positionen []eingabe.PositionRef `json:"positionen"`
	Kommentar  string                `json:"kommentar"`
}

// UmbuchungID is the client's idempotency key, kept per attempt.
type bestellungUmbuchenRequest struct {
	UmbuchungID       string                `json:"umbuchungId"`
	QuellTischID      int                   `json:"quellTischId"`
	ZielTischID       int                   `json:"zielTischId"`
	Positionen        []eingabe.PositionRef `json:"positionen"`
	BenutzerKommentar string                `json:"benutzerKommentar"`
}

var stornierungErteilenSchema = z.Struct(z.Shape{
	"StornoID":   z.String().UUID().Required(),
	"TischID":    tisch.TischIDSchema.Required(),
	"Positionen": z.Slice(eingabe.PositionRefSchema).Min(1).Required(),
	"Kommentar":  z.String().TestFunc(zeichen.Min(3)).TestFunc(zeichen.Max(100)).Required(),
})

var bestellungUmbuchenSchema = z.Struct(z.Shape{
	"UmbuchungID":       z.String().UUID().Required(),
	"QuellTischID":      tisch.TischIDSchema.Required(),
	"ZielTischID":       tisch.TischIDSchema.Required(),
	"Positionen":        z.Slice(eingabe.PositionRefSchema).Min(1).Required(),
	"BenutzerKommentar": z.String().TestFunc(zeichen.Max(100)),
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
		err := h.Command.StornierungErteilen(r.Context(), userID, userName, uuid.MustParse(body.StornoID), body.TischID, eingabe.ToPositionRefs(body.Positionen), body.Kommentar)
		if err != nil {
			helper.MapError(w, err, slices.Concat(kasseFehler, []helper.ErrorCode{
				{Err: application.ErrTischNotFound, Code: "tisch_not_found"},
				{Err: application.ErrTischNotActive, Code: "tisch_not_active"},
				{Err: application.ErrPositionNichtStornierbar, Code: "position_nicht_stornierbar"},
			}))
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

		err := h.Command.BestellungUmbuchen(r.Context(), userID, userName, body.UmbuchungID, body.QuellTischID, body.ZielTischID, eingabe.ToPositionRefs(body.Positionen), body.BenutzerKommentar)
		if err != nil {
			helper.MapError(w, err, slices.Concat(kasseFehler, []helper.ErrorCode{
				{Err: application.ErrTischNotFound, Code: "tisch_not_found"},
				{Err: application.ErrTischNotActive, Code: "tisch_not_active"},
				{Err: application.ErrPositionNichtUmbuchbar, Code: "position_nicht_umbuchbar"},
				{Err: application.ErrUmbuchungGleicherTisch, Code: "umbuchung_gleicher_tisch"},
			}))
			return
		}

		helper.SendEmptyResponse(w)
	}
}
