package http

import (
	"context"
	"net/http"
	"slices"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/kasse/direktverkauf/application"
	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
	"github.com/nicograef/jotti/backend/api/middleware"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/produkt"
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

type verkaufPositionInput struct {
	ProduktID  int `json:"produktId"`
	VarianteID int `json:"varianteId"`
	Menge      int `json:"menge"`
}

type pfandRequest struct {
	RueckgabeID string                 `json:"rueckgabeId"`
	Positionen  []verkaufPositionInput `json:"positionen"`
}

type direktverkaufTaetigenRequest struct {
	VerkaufID  string                 `json:"verkaufId"`
	Positionen []verkaufPositionInput `json:"positionen"`
	Kommentar  string                 `json:"kommentar"`
	Pfand      *pfandRequest          `json:"pfand"`
}

var verkaufPositionInputSchema = z.Struct(z.Shape{
	"ProduktID":  produkt.IDSchema.Required(),
	"VarianteID": produkt.IDSchema.Required(),
	"Menge":      kasse.PositionEingabeSchema,
})

var pfandRequestSchema = z.Struct(z.Shape{
	"RueckgabeID": z.String().UUID().Required(),
	"Positionen":  z.Slice(verkaufPositionInputSchema).Min(1).Required(),
})

var direktverkaufTaetigenSchema = z.Struct(z.Shape{
	"VerkaufID":  z.String().UUID().Required(),
	"Positionen": z.Slice(verkaufPositionInputSchema).Min(1).Required(),
	"Kommentar":  z.String().Max(100),
	"Pfand":      z.Ptr(pfandRequestSchema),
}).TestFunc(func(val any, _ z.Ctx) bool {
	body, ok := val.(*direktverkaufTaetigenRequest)
	// Beide IDs werden BON_ID im DSFinV-K-Export desselben Kassenabschlusses.
	return ok && (body.Pfand == nil || body.Pfand.RueckgabeID != body.VerkaufID)
}, z.Message("Rückgabe-ID muss sich von der Verkaufs-ID unterscheiden"))

func toPfandEingabe(pfand *pfandRequest) *enrichment.PfandEingabe {
	if pfand == nil {
		return nil
	}
	return &enrichment.PfandEingabe{RueckgabeID: pfand.RueckgabeID, Positionen: toVerkaufPositionInputs(pfand.Positionen)}
}

func toVerkaufPositionInputs(positionen []verkaufPositionInput) []enrichment.PositionInput {
	out := make([]enrichment.PositionInput, len(positionen))
	for i, p := range positionen {
		out[i] = enrichment.PositionInput{
			ProduktID:  p.ProduktID,
			VarianteID: p.VarianteID,
			Menge:      p.Menge,
		}
	}
	return out
}

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

		err := h.Command.DirektverkaufTaetigen(r.Context(), userID, userName, body.VerkaufID, toVerkaufPositionInputs(body.Positionen), body.Kommentar, toPfandEingabe(body.Pfand))
		if err != nil {
			helper.MapError(w, err, slices.Concat(kasseFehler, []helper.ErrorCode{
				{Err: application.ErrPfandrueckgabeAbweichend, Code: "pfandrueckgabe_abweichend", Status: http.StatusConflict},
			}, pfandEingabeFehler))
			return
		}

		helper.SendEmptyResponse(w)
	}
}

type positionRefRequest struct {
	PositionID string `json:"positionId"`
	Menge      int    `json:"menge"`
}

var positionRefRequestSchema = z.Struct(z.Shape{
	"PositionID": z.String().UUID().Required(),
	"Menge":      z.Int().GTE(1).Required(),
})

func toPositionRefs(refs []positionRefRequest) []kasse.PositionRef {
	out := make([]kasse.PositionRef, len(refs))
	for i, ref := range refs {
		out[i] = kasse.PositionRef{PositionID: ref.PositionID, Menge: ref.Menge}
	}
	return out
}

type direktverkaufStornierenRequest struct {
	VerkaufID  string               `json:"verkaufId"`
	Positionen []positionRefRequest `json:"positionen"`
	Kommentar  string               `json:"kommentar"`
}

var direktverkaufStornierenSchema = z.Struct(z.Shape{
	"VerkaufID":  z.String().UUID().Required(),
	"Positionen": z.Slice(positionRefRequestSchema).Min(1).Required(),
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

		err := h.Command.DirektverkaufStornieren(r.Context(), userID, userName, body.VerkaufID, toPositionRefs(body.Positionen), body.Kommentar)
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
