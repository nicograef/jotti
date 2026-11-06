// Package eingabe holds the request DTOs and zog schemas the tischgeschaeft and direktverkauf
// handlers share, so both validate one wire shape the same way.
package eingabe

import (
	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/produkt"
)

// PositionInput is one ordered, sold or returned product variant.
type PositionInput struct {
	ProduktID  int `json:"produktId"`
	VarianteID int `json:"varianteId"`
	Menge      int `json:"menge"`
}

var PositionInputSchema = z.Struct(z.Shape{
	"ProduktID":  produkt.IDSchema.Required(),
	"VarianteID": produkt.IDSchema.Required(),
	"Menge":      kasse.PositionEingabeSchema,
})

// PositionRef addresses a quantity of an existing position.
type PositionRef struct {
	PositionID string `json:"positionId"`
	Menge      int    `json:"menge"`
}

var PositionRefSchema = z.Struct(z.Shape{
	"PositionID": z.String().UUID().Required(),
	"Menge":      z.Int().GTE(1).Required(),
})

// PfandRequest is the Pfand return settled within a Zahlung or a Direktverkauf.
type PfandRequest struct {
	RueckgabeID string          `json:"rueckgabeId"`
	Positionen  []PositionInput `json:"positionen"`
}

var PfandSchema = z.Struct(z.Shape{
	"RueckgabeID": z.String().UUID().Required(),
	"Positionen":  z.Slice(PositionInputSchema).Min(1).Required(),
})

func ToPositionInputs(positionen []PositionInput) []enrichment.PositionInput {
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

func ToPositionRefs(refs []PositionRef) []kasse.PositionRef {
	out := make([]kasse.PositionRef, len(refs))
	for i, ref := range refs {
		out[i] = kasse.PositionRef{PositionID: ref.PositionID, Menge: ref.Menge}
	}
	return out
}

func ToPfandEingabe(pfand *PfandRequest) *enrichment.PfandEingabe {
	if pfand == nil {
		return nil
	}
	return &enrichment.PfandEingabe{RueckgabeID: pfand.RueckgabeID, Positionen: ToPositionInputs(pfand.Positionen)}
}
