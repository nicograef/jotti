package table

import (
	"time"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/domain/product"
)

type OrderVariant struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	PriceCents int    `json:"priceCents"`
	Quantity   int    `json:"quantity"`
}

var orderVariantSchema = z.Struct(z.Shape{
	"ID":         product.IDSchema.Required(),
	"Name":       product.NameSchema.Required(),
	"PriceCents": product.PriceCentsSchema.Required(),
	"Quantity":   z.Int().GTE(1, z.Message("Quantity must be at least 1")).Required(),
})

type Order struct {
	ID              string         `json:"id"`
	UserID          int            `json:"userId"`
	TableID         int            `json:"tableId"`
	Variants        []OrderVariant `json:"variants"`
	TotalPriceCents int            `json:"totalPriceCents"`
	Comment         string         `json:"comment"`
	PlacedAt        time.Time      `json:"placedAt"`
}

var orderSchema = z.Struct(z.Shape{
	"ID":              z.String().UUID().Required(),
	"UserID":          z.Int().GTE(1).Required(),
	"TableID":         z.Int().GTE(1).Required(),
	"Variants":        z.Slice(orderVariantSchema).Min(1).Required(),
	"TotalPriceCents": z.Int().GTE(0).Required(),
	"Comment":         z.String().Max(100),
	"PlacedAt":        z.Time().Required(),
})
