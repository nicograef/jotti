package kasse

import (
	"time"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/domain/product"
	"github.com/nicograef/jotti/backend/domain/steuer"
)

type Position struct {
	PositionID   string
	VarianteID   int
	ProduktName  string
	VarianteName string
	Kategorie    string
	Steuersatz   string
	Einzelpreis  int
	Menge        int
}

// positionEventData is the serialization-friendly representation of Position for the event store.
// The json-keys are stable and must not be changed (immutable events).
type positionEventData struct {
	PositionID   string `json:"positionId"`
	VarianteID   int    `json:"varianteId"`
	ProduktName  string `json:"produktName"`
	VarianteName string `json:"varianteName"`
	Kategorie    string `json:"kategorie"`
	Steuersatz   string `json:"steuersatz"`
	Einzelpreis  int    `json:"einzelpreis"`
	Menge        int    `json:"menge"`
}

func toPositionenEventData(positionen []Position) []positionEventData {
	out := make([]positionEventData, len(positionen))
	for i, p := range positionen {
		out[i] = positionEventData(p)
	}
	return out
}

func fromPositionenEventData(positionen []positionEventData) []Position {
	out := make([]Position, len(positionen))
	for i, p := range positionen {
		out[i] = Position(p)
	}
	return out
}

var positionSchema = z.Struct(z.Shape{
	"PositionID":   z.String().UUID().Required(),
	"VarianteID":   product.IDSchema.Required(),
	"ProduktName":  product.NameSchema.Required(),
	"VarianteName": product.NameSchema.Required(),
	"Kategorie":    z.String().OneOf([]string{string(product.EssenKategorie), string(product.GetraenkKategorie), string(product.SonstigesKategorie)}, z.Message("Ungültige Kategorie")).Required(),
	"Steuersatz":   z.String().OneOf([]string{string(steuer.RegelSteuersatz), string(steuer.ErmaessigtSteuersatz), string(steuer.BefreitSteuersatz), string(steuer.KombiSteuersatz)}, z.Message("Ungültiger Steuersatz")).Required(),
	"Einzelpreis":  product.PreisCentsSchema.Required(),
	"Menge":        z.Int().GTE(1, z.Message("Menge muss mindestens 1 betragen")).Required(),
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
	AufgenommenAm    time.Time
}

var bestellungSchema = z.Struct(z.Shape{
	"ID":               z.String().UUID().Required(),
	"UserID":           z.Int().GTE(1).Required(),
	"TischID":          z.Int().GTE(1).Required(),
	"Positionen":       z.Slice(positionSchema).Min(1).Required(),
	"GesamtPreisCents": z.Int().GTE(0).Required(),
	"Kommentar":        z.String().Max(100),
	"AufgenommenAm":    z.Time().Required(),
})
