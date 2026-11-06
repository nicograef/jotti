package http

import (
	"context"
	"math"
	"net/http"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/domain/kasse"
)

type query interface {
	GetDirektverkaufHistorie(ctx context.Context) ([]kasse.DirektverkaufHistorieEintrag, error)
	GetPfandrueckgabeHistorie(ctx context.Context, tischID int) ([]kasse.PfandrueckgabeHistorieEintrag, error)
}

type QueryHandler struct {
	Query query
}

type position struct {
	PositionID       string `json:"positionId"`
	VarianteID       int    `json:"varianteId"`
	ProduktName      string `json:"produktName"`
	VarianteName     string `json:"varianteName"`
	Kategorie        string `json:"kategorie"`
	Steuersatz       string `json:"steuersatz"`
	EinzelpreisCents int    `json:"einzelpreisCents"`
	Menge            int    `json:"menge"`
}

func toPosition(p kasse.Position) position {
	return position{
		PositionID:       p.PositionID,
		VarianteID:       p.VarianteID,
		ProduktName:      p.ProduktName,
		VarianteName:     p.VarianteName,
		Kategorie:        p.Kategorie,
		Steuersatz:       p.Steuersatz,
		EinzelpreisCents: p.EinzelpreisCents,
		Menge:            p.Menge,
	}
}

func toPositionen(positionen []kasse.Position) []position {
	out := make([]position, 0, len(positionen))
	for _, p := range positionen {
		out = append(out, toPosition(p))
	}
	return out
}

type vorgangStornierung struct {
	StornierungID          string    `json:"stornierungId"`
	StorniertAm            time.Time `json:"storniertAm"`
	GesamtStornierungCents int       `json:"gesamtStornierungCents"`
}

type direktverkaufHistorieEintrag struct {
	VerkaufID            string               `json:"verkaufId"`
	UserName             string               `json:"userName"`
	GetaetigtAm          time.Time            `json:"getaetigtAm"`
	Positionen           []position           `json:"positionen"`
	GesamtbetragCents    int                  `json:"gesamtbetragCents"`
	Kommentar            string               `json:"kommentar"`
	OffenePositionen     []position           `json:"offenePositionen"`
	GesamtStorniertCents int                  `json:"gesamtStorniertCents"`
	Stornierungen        []vorgangStornierung `json:"stornierungen"`
}

type getDirektverkaufHistorieResponse struct {
	Historie []direktverkaufHistorieEintrag `json:"historie"`
}

func toStornierungen(stornierungen []kasse.VorgangStornierung) []vorgangStornierung {
	out := make([]vorgangStornierung, 0, len(stornierungen))
	for _, s := range stornierungen {
		out = append(out, vorgangStornierung{
			StornierungID:          s.StornierungID,
			StorniertAm:            s.StorniertAm,
			GesamtStornierungCents: s.GesamtStornierungCents,
		})
	}
	return out
}

func toHistorieEintrag(e kasse.DirektverkaufHistorieEintrag) direktverkaufHistorieEintrag {
	return direktverkaufHistorieEintrag{
		VerkaufID:            e.VerkaufID,
		UserName:             e.UserName,
		GetaetigtAm:          e.GetaetigtAm,
		Positionen:           toPositionen(e.Positionen),
		GesamtbetragCents:    e.GesamtbetragCents,
		Kommentar:            e.Kommentar,
		OffenePositionen:     toPositionen(e.OffenePositionen),
		GesamtStorniertCents: e.GesamtStorniertCents,
		Stornierungen:        toStornierungen(e.Stornierungen),
	}
}

func (h *QueryHandler) GetDirektverkaufHistorieHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		historie, err := h.Query.GetDirektverkaufHistorie(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}

		eintraege := make([]direktverkaufHistorieEintrag, 0, len(historie))
		for i := range historie {
			eintraege = append(eintraege, toHistorieEintrag(historie[i]))
		}

		helper.SendResponse(w, getDirektverkaufHistorieResponse{Historie: eintraege})
	}
}

type pfandrueckgabeHistorieEintrag struct {
	RueckgabeID          string               `json:"rueckgabeId"`
	TischID              int                  `json:"tischId"`
	UserName             string               `json:"userName"`
	GetaetigtAm          time.Time            `json:"getaetigtAm"`
	Positionen           []position           `json:"positionen"`
	GesamtbetragCents    int                  `json:"gesamtbetragCents"`
	Kommentar            string               `json:"kommentar"`
	OffenePositionen     []position           `json:"offenePositionen"`
	GesamtStorniertCents int                  `json:"gesamtStorniertCents"`
	Stornierungen        []vorgangStornierung `json:"stornierungen"`
}

type getPfandrueckgabeHistorieRequest struct {
	TischID *int `json:"tischId"`
}

var getPfandrueckgabeHistorieSchema = z.Struct(z.Shape{
	// Pointer: tischId 0 (Direktverkauf) is valid, a missing tischId is not.
	"TischID": z.Ptr(z.Int().GTE(0).LTE(math.MaxInt32)).NotNil(z.Message("Tisch-ID ist erforderlich")),
})

type getPfandrueckgabeHistorieResponse struct {
	Historie []pfandrueckgabeHistorieEintrag `json:"historie"`
}

func toPfandrueckgabeHistorieEintrag(e kasse.PfandrueckgabeHistorieEintrag) pfandrueckgabeHistorieEintrag {
	return pfandrueckgabeHistorieEintrag{
		RueckgabeID:          e.RueckgabeID,
		TischID:              e.TischID,
		UserName:             e.UserName,
		GetaetigtAm:          e.GetaetigtAm,
		Positionen:           toPositionen(e.Positionen),
		GesamtbetragCents:    e.GesamtbetragCents,
		Kommentar:            e.Kommentar,
		OffenePositionen:     toPositionen(e.OffenePositionen),
		GesamtStorniertCents: e.GesamtStorniertCents,
		Stornierungen:        toStornierungen(e.Stornierungen),
	}
}

func (h *QueryHandler) GetPfandrueckgabeHistorieHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := getPfandrueckgabeHistorieRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, getPfandrueckgabeHistorieSchema) {
			return
		}

		historie, err := h.Query.GetPfandrueckgabeHistorie(r.Context(), *body.TischID)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		eintraege := make([]pfandrueckgabeHistorieEintrag, 0, len(historie))
		for i := range historie {
			eintraege = append(eintraege, toPfandrueckgabeHistorieEintrag(historie[i]))
		}

		helper.SendResponse(w, getPfandrueckgabeHistorieResponse{Historie: eintraege})
	}
}
