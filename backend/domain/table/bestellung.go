package table

import (
	"time"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/domain/product"
)

type Position struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	PreisCents int    `json:"preisCents"`
	Quantity   int    `json:"quantity"`
}

var positionSchema = z.Struct(z.Shape{
	"ID":         product.IDSchema.Required(),
	"Name":       product.NameSchema.Required(),
	"PreisCents": product.PriceCentsSchema.Required(),
	"Quantity":   z.Int().GTE(1, z.Message("Quantity must be at least 1")).Required(),
})

type Bestellung struct {
	ID               string     `json:"id"`
	UserID           int        `json:"userId"`
	TischID          int        `json:"tischId"`
	Positionen       []Position `json:"positionen"`
	GesamtPreisCents int        `json:"gesamtPreisCents"`
	Comment          string     `json:"comment"`
	AufgegebenAm     time.Time  `json:"aufgegebenAm"`
}

var bestellungSchema = z.Struct(z.Shape{
	"ID":               z.String().UUID().Required(),
	"UserID":           z.Int().GTE(1).Required(),
	"TischID":          z.Int().GTE(1).Required(),
	"Positionen":       z.Slice(positionSchema).Min(1).Required(),
	"GesamtPreisCents": z.Int().GTE(0).Required(),
	"Comment":          z.String().Max(100),
	"AufgegebenAm":     z.Time().Required(),
})
