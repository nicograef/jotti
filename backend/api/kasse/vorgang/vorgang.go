// Package vorgang holds the response DTOs that the Direktverkauf and Pfandrückgabe histories share.
package vorgang

import (
	"time"

	"github.com/nicograef/jotti/backend/domain/kasse"
)

type Position struct {
	PositionID       string `json:"positionId"`
	VarianteID       int    `json:"varianteId"`
	ProduktName      string `json:"produktName"`
	VarianteName     string `json:"varianteName"`
	Kategorie        string `json:"kategorie"`
	Steuersatz       string `json:"steuersatz"`
	EinzelpreisCents int    `json:"einzelpreisCents"`
	Menge            int    `json:"menge"`
}

type Stornierung struct {
	StornierungID          string    `json:"stornierungId"`
	StorniertAm            time.Time `json:"storniertAm"`
	GesamtStornierungCents int       `json:"gesamtStornierungCents"`
}

// HistorieEintrag carries kasse.VorgangHistorieEintrag; each history entry DTO embeds it, and
// encoding/json flattens the embedded fields into the entry.
type HistorieEintrag struct {
	UserName             string        `json:"userName"`
	GetaetigtAm          time.Time     `json:"getaetigtAm"`
	Positionen           []Position    `json:"positionen"`
	GesamtbetragCents    int           `json:"gesamtbetragCents"`
	Kommentar            string        `json:"kommentar"`
	OffenePositionen     []Position    `json:"offenePositionen"`
	GesamtStorniertCents int           `json:"gesamtStorniertCents"`
	Stornierungen        []Stornierung `json:"stornierungen"`
}

func ToHistorieEintrag(e kasse.VorgangHistorieEintrag) HistorieEintrag {
	return HistorieEintrag{
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

func toPositionen(positionen []kasse.Position) []Position {
	out := make([]Position, 0, len(positionen))
	for _, p := range positionen {
		out = append(out, Position{
			PositionID:       p.PositionID,
			VarianteID:       p.VarianteID,
			ProduktName:      p.ProduktName,
			VarianteName:     p.VarianteName,
			Kategorie:        p.Kategorie,
			Steuersatz:       p.Steuersatz,
			EinzelpreisCents: p.EinzelpreisCents,
			Menge:            p.Menge,
		})
	}
	return out
}

func toStornierungen(stornierungen []kasse.VorgangStornierung) []Stornierung {
	out := make([]Stornierung, 0, len(stornierungen))
	for _, s := range stornierungen {
		out = append(out, Stornierung{
			StornierungID:          s.StornierungID,
			StorniertAm:            s.StorniertAm,
			GesamtStornierungCents: s.GesamtStornierungCents,
		})
	}
	return out
}
