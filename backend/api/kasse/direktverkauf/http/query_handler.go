package http

import (
	"context"
	"net/http"
	"time"

	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/domain/kasse"
)

type query interface {
	GetDirektverkaufHistorie(ctx context.Context) ([]kasse.DirektverkaufHistorieEintrag, error)
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

type direktverkaufStornierung struct {
	StornierungID          string    `json:"stornierungId"`
	StorniertAm            time.Time `json:"storniertAm"`
	GesamtStornierungCents int       `json:"gesamtStornierungCents"`
}

type direktverkaufHistorieEintrag struct {
	VerkaufID            string                     `json:"verkaufId"`
	UserName             string                     `json:"userName"`
	GetaetigtAm          time.Time                  `json:"getaetigtAm"`
	Positionen           []position                 `json:"positionen"`
	GesamtbetragCents    int                        `json:"gesamtbetragCents"`
	Kommentar            string                     `json:"kommentar"`
	OffenePositionen     []position                 `json:"offenePositionen"`
	GesamtStorniertCents int                        `json:"gesamtStorniertCents"`
	Stornierungen        []direktverkaufStornierung `json:"stornierungen"`
}

type getDirektverkaufHistorieResponse struct {
	Historie []direktverkaufHistorieEintrag `json:"historie"`
}

func toStornierungen(stornierungen []kasse.DirektverkaufStornierung) []direktverkaufStornierung {
	out := make([]direktverkaufStornierung, 0, len(stornierungen))
	for _, s := range stornierungen {
		out = append(out, direktverkaufStornierung{
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
