package table

import (
	"time"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/domain/product"
)

type Position struct {
	PositionID   string `json:"positionId"`
	VarianteID   int    `json:"varianteId"`
	ProduktName  string `json:"produktName"`
	VarianteName string `json:"varianteName"`
	Kategorie    string `json:"kategorie"`
	Einzelpreis  int    `json:"einzelpreis"`
	Menge        int    `json:"menge"`
}

var positionSchema = z.Struct(z.Shape{
	"PositionID":   z.String().UUID().Required(),
	"VarianteID":   product.IDSchema.Required(),
	"ProduktName":  product.NameSchema.Required(),
	"VarianteName": product.NameSchema.Required(),
	"Kategorie":    z.String().OneOf([]string{"food", "beverage", "other"}, z.Message("Invalid category")).Required(),
	"Einzelpreis":  product.PreisCentsSchema.Required(),
	"Menge":        z.Int().GTE(1, z.Message("Menge must be at least 1")).Required(),
})

// PositionRef is a lightweight reference to a position, used in delivery/payment/cancellation events.
type PositionRef struct {
	PositionID string `json:"positionId"`
	Menge      int    `json:"menge"`
}

var positionRefSchema = z.Struct(z.Shape{
	"PositionID": z.String().UUID().Required(),
	"Menge":      z.Int().GTE(1, z.Message("Menge must be at least 1")).Required(),
})

type Bestellung struct {
	ID               string     `json:"id"`
	UserID           int        `json:"userId"`
	TischID          int        `json:"tischId"`
	Positionen       []Position `json:"positionen"`
	GesamtPreisCents int        `json:"gesamtPreisCents"`
	Kommentar        string     `json:"kommentar"`
	AufgegebenAm     time.Time  `json:"aufgegebenAm"`
}

var bestellungSchema = z.Struct(z.Shape{
	"ID":               z.String().UUID().Required(),
	"UserID":           z.Int().GTE(1).Required(),
	"TischID":          z.Int().GTE(1).Required(),
	"Positionen":       z.Slice(positionSchema).Min(1).Required(),
	"GesamtPreisCents": z.Int().GTE(0).Required(),
	"Kommentar":        z.String().Max(100),
	"AufgegebenAm":     z.Time().Required(),
})
