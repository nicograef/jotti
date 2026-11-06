package http

import (
	"context"
	"net/http"
	"slices"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/kasse/direktverkauf/application"
	"github.com/nicograef/jotti/backend/api/kasse/eingabe"
	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
	"github.com/nicograef/jotti/backend/api/middleware"
	"github.com/nicograef/jotti/backend/domain/kasse"
)

type command interface {
	DirektverkaufTaetigen(ctx context.Context, userID int, userName string, verkaufID string, positionen []enrichment.PositionInput, kommentar string, pfand *enrichment.PfandEingabe) error
	DirektverkaufStornieren(ctx context.Context, userID int, userName string, verkaufID string, positionen []kasse.PositionRef, kommentar string) error
}

type CommandHandler struct {
	Command command
}

// kasseFehler opens every Kasse handler's table: a stale write or a closed Kasse is a 409.
var kasseFehler = []helper.ErrorCode{
	{Err: application.ErrConflict, Code: "conflict", Status: http.StatusConflict},
	{Err: application.ErrKasseWirdAbgeschlossen, Code: "kasse_wird_abgeschlossen", Status: http.StatusConflict},
	{Err: application.ErrKasseNichtGeoeffnet, Code: "kasse_nicht_geoeffnet", Status: http.StatusConflict},
}

// pfandEingabeFehler covers the Pfand and product checks of the position enrichment.
var pfandEingabeFehler = []helper.ErrorCode{
	{Err: enrichment.ErrKeinPfandProdukt, Code: "kein_pfand_produkt"},
	{Err: enrichment.ErrProduktNotFound, Code: "produkt_not_found"},
	{Err: enrichment.ErrVarianteNichtAktiv, Code: "variante_nicht_aktiv"},
}

type direktverkaufTaetigenRequest struct {
	VerkaufID  string                  `json:"verkaufId"`
	Positionen []eingabe.PositionInput `json:"positionen"`
	Kommentar  string                  `json:"kommentar"`
	Pfand      *eingabe.PfandRequest   `json:"pfand"`
}

var direktverkaufTaetigenSchema = z.Struct(z.Shape{
	"VerkaufID":  z.String().UUID().Required(),
	"Positionen": z.Slice(eingabe.PositionInputSchema).Min(1).Required(),
	"Kommentar":  z.String().Max(100),
	"Pfand":      z.Ptr(eingabe.PfandSchema),
}).TestFunc(func(val any, _ z.Ctx) bool {
	body, ok := val.(*direktverkaufTaetigenRequest)
	// Beide IDs werden BON_ID im DSFinV-K-Export desselben Kassenabschlusses.
	return ok && (body.Pfand == nil || body.Pfand.RueckgabeID != body.VerkaufID)
}, z.Message("Rückgabe-ID muss sich von der Verkaufs-ID unterscheiden"))

func (h *CommandHandler) DirektverkaufTaetigenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := direktverkaufTaetigenRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, direktverkaufTaetigenSchema) {
			return
		}

		userID, userName, ok := middleware.UserFromContext(r.Context())
		if !ok {
			helper.SendServerError(w)
			return
		}

		err := h.Command.DirektverkaufTaetigen(r.Context(), userID, userName, body.VerkaufID, eingabe.ToPositionInputs(body.Positionen), body.Kommentar, eingabe.ToPfandEingabe(body.Pfand))
		if err != nil {
			helper.MapError(w, err, slices.Concat(kasseFehler, []helper.ErrorCode{
				{Err: application.ErrPfandrueckgabeAbweichend, Code: "pfandrueckgabe_abweichend", Status: http.StatusConflict},
			}, pfandEingabeFehler))
			return
		}

		helper.SendEmptyResponse(w)
	}
}

type direktverkaufStornierenRequest struct {
	VerkaufID  string                `json:"verkaufId"`
	Positionen []eingabe.PositionRef `json:"positionen"`
	Kommentar  string                `json:"kommentar"`
}

var direktverkaufStornierenSchema = z.Struct(z.Shape{
	"VerkaufID":  z.String().UUID().Required(),
	"Positionen": z.Slice(eingabe.PositionRefSchema).Min(1).Required(),
	"Kommentar":  z.String().Min(3).Max(100).Required(),
})

func (h *CommandHandler) DirektverkaufStornierenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := direktverkaufStornierenRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, direktverkaufStornierenSchema) {
			return
		}

		userID, userName, ok := middleware.UserFromContext(r.Context())
		if !ok {
			helper.SendServerError(w)
			return
		}

		err := h.Command.DirektverkaufStornieren(r.Context(), userID, userName, body.VerkaufID, eingabe.ToPositionRefs(body.Positionen), body.Kommentar)
		if err != nil {
			helper.MapError(w, err, slices.Concat(kasseFehler, []helper.ErrorCode{
				{Err: application.ErrVerkaufNichtGefunden, Code: "verkauf_not_found"},
				{Err: application.ErrPositionNichtStornierbar, Code: "position_nicht_stornierbar"},
			}))
			return
		}

		helper.SendEmptyResponse(w)
	}
}
