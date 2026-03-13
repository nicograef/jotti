package table

import (
	"time"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/domain/product"
)

type Position struct {
	PositionID   string
	VarianteID   int
	ProduktName  string
	VarianteName string
	Kategorie    string
	Einzelpreis  int
	Menge        int
}

var positionSchema = z.Struct(z.Shape{
	"PositionID":   z.String().UUID().Required(),
	"VarianteID":   product.IDSchema.Required(),
	"ProduktName":  product.NameSchema.Required(),
	"VarianteName": product.NameSchema.Required(),
	"Kategorie":    z.String().OneOf([]string{"essen", "getraenk", "sonstiges"}, z.Message("Invalid category")).Required(),
	"Einzelpreis":  product.PreisCentsSchema.Required(),
	"Menge":        z.Int().GTE(1, z.Message("Menge must be at least 1")).Required(),
})

// PositionRef is a lightweight reference used in API request commands for payment/delivery/cancellation.
// Enriched to Position (fat) in the command layer before being stored in events.
type PositionRef struct {
	PositionID string
	Menge      int
}

type Bestellung struct {
	ID               string
	UserID           int
	TischID          int
	Positionen       []Position
	GesamtPreisCents int
	Kommentar        string
	AufgegebenAm     time.Time
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
