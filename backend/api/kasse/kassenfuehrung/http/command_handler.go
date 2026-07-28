package http

import (
	"context"
	"errors"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api/helper"
	kasseApp "github.com/nicograef/jotti/backend/api/kasse/kassenfuehrung/application"
	"github.com/nicograef/jotti/backend/api/middleware"
)

type command interface {
	KassensitzungEroeffnen(ctx context.Context, userID int, userName string, bezeichnung string, betragCents int) (int, error)
	GeldtransitBuchen(ctx context.Context, userID int, userName string, geldtransitID string, richtung string, betragCents int, kommentar string) error
	KasseAbschliessen(ctx context.Context, userID int, userName string, istBestandCents int) (kasseApp.KassenabschlussErgebnis, error)
}

type CommandHandler struct {
	Command command
}

// --- Request / Response DTOs ---

type kassensitzungEroeffnenRequest struct {
	Bezeichnung string `json:"bezeichnung"`
	BetragCents *int   `json:"betragCents"`
}

var kassensitzungEroeffnenSchema = z.Struct(z.Shape{
	"Bezeichnung": z.String().Min(1, z.Message("Bezeichnung ist erforderlich")).Max(200, z.Message("Bezeichnung darf höchstens 200 Zeichen lang sein")).Required(),
	"BetragCents": z.Ptr(z.Int().GTE(0, z.Message("Anfangsbestand darf nicht negativ sein"))).NotNil(z.Message("Anfangsbestand ist erforderlich")),
})

type kassensitzungEroeffnenResponse struct {
	ZNr int `json:"zNr"`
}

type geldtransitBuchenRequest struct {
	GeldtransitID string `json:"geldtransitId"`
	Richtung      string `json:"richtung"`
	BetragCents   int    `json:"betragCents"`
	Kommentar     string `json:"kommentar"`
}

var geldtransitBuchenSchema = z.Struct(z.Shape{
	"GeldtransitID": z.String().UUID().Required(),
	"Richtung": z.String().OneOf(
		[]string{"einlage", "entnahme"},
		z.Message("Ungültige Richtung"),
	).Required(),
	"BetragCents": z.Int().GTE(1, z.Message("Betrag muss mindestens 1 Cent sein")).Required(),
	"Kommentar":   z.String().Min(3, z.Message("Kommentar muss mindestens 3 Zeichen lang sein")).Max(200, z.Message("Kommentar darf höchstens 200 Zeichen lang sein")).Required(),
})

type kasseAbschliessenRequest struct {
	IstBestandCents *int `json:"istBestandCents"`
}

var kasseAbschliessenSchema = z.Struct(z.Shape{
	"IstBestandCents": z.Ptr(z.Int().GTE(0, z.Message("Ist-Bestand darf nicht negativ sein"))).NotNil(z.Message("Ist-Bestand ist erforderlich")),
})

// kasseAbschliessenResponse weist die beim Abschluss verbliebenen Ausfall-Reste
// aus: Vorgänge, die die TSE noch nachsigniert (AusfallResteAnzahl) und
// Vorgänge ohne Signatur mangels TSE-Konfiguration (OhneKonfigurationAnzahl).
type kasseAbschliessenResponse struct {
	AusfallResteAnzahl      int `json:"ausfallResteAnzahl"`
	OhneKonfigurationAnzahl int `json:"ohneKonfigurationAnzahl"`
}

// signaturenAusstehendDetails sind die strukturierten 409-Details des Gates:
// wie viele Signaturen noch ausstehen.
type signaturenAusstehendDetails struct {
	Anzahl int `json:"anzahl"`
}

// --- Handlers ---

func (h *CommandHandler) KassensitzungEroeffnenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := kassensitzungEroeffnenRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, kassensitzungEroeffnenSchema) {
			return
		}

		userID, userName, ok := middleware.UserFromContext(r.Context())
		if !ok {
			helper.SendServerError(w)
			return
		}

		zNr, err := h.Command.KassensitzungEroeffnen(r.Context(), userID, userName, body.Bezeichnung, *body.BetragCents)
		if err != nil {
			helper.MapError(w, err, map[error]string{
				kasseApp.ErrKasseAlreadyOpen:           "kasse_bereits_geoeffnet",
				kasseApp.ErrBetreiberNichtKonfiguriert: "betreiber_nicht_konfiguriert",
			})
			return
		}

		helper.SendResponse(w, kassensitzungEroeffnenResponse{ZNr: zNr})
	}
}

func (h *CommandHandler) GeldtransitBuchenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := geldtransitBuchenRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, geldtransitBuchenSchema) {
			return
		}

		userID, userName, ok := middleware.UserFromContext(r.Context())
		if !ok {
			helper.SendServerError(w)
			return
		}

		err := h.Command.GeldtransitBuchen(r.Context(), userID, userName, body.GeldtransitID, body.Richtung, body.BetragCents, body.Kommentar)
		if err != nil {
			switch {
			case errors.Is(err, kasseApp.ErrConflict):
				helper.SendConflictError(w)
			case errors.Is(err, kasseApp.ErrVorgangDatenAbweichend):
				helper.SendConflict(w, "vorgang_daten_abweichend")
			case errors.Is(err, kasseApp.ErrKasseWirdAbgeschlossen):
				helper.SendConflict(w, "kasse_wird_abgeschlossen")
			case errors.Is(err, kasseApp.ErrKasseNichtGeoeffnet):
				helper.SendConflict(w, "kasse_nicht_geoeffnet")
			default:
				helper.SendServerError(w)
			}
			return
		}

		helper.SendEmptyResponse(w)
	}
}

func (h *CommandHandler) KasseAbschliessenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := kasseAbschliessenRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, kasseAbschliessenSchema) {
			return
		}

		userID, userName, ok := middleware.UserFromContext(r.Context())
		if !ok {
			helper.SendServerError(w)
			return
		}

		ergebnis, err := h.Command.KasseAbschliessen(r.Context(), userID, userName, *body.IstBestandCents)
		if err != nil {
			var ausstehend *kasseApp.SignaturenAusstehendError
			switch {
			case errors.As(err, &ausstehend):
				helper.SendConflictDetails(w, "signaturen_ausstehend", signaturenAusstehendDetails{
					Anzahl: ausstehend.Anzahl,
				})
			case errors.Is(err, kasseApp.ErrConflict):
				helper.SendConflictError(w)
			case errors.Is(err, kasseApp.ErrKasseNichtGeoeffnet):
				helper.SendConflict(w, "kasse_nicht_geoeffnet")
			default:
				helper.MapError(w, err, map[error]string{
					kasseApp.ErrTischeSaldoOffen:         "tische_saldo_offen",
					kasseApp.ErrBuchungenNachKassensturz: "buchungen_nach_kassensturz",
				})
			}
			return
		}

		helper.SendResponse(w, kasseAbschliessenResponse{
			AusfallResteAnzahl:      ergebnis.AusfallResteAnzahl,
			OhneKonfigurationAnzahl: ergebnis.OhneKonfigurationAnzahl,
		})
	}
}
