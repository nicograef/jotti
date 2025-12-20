package table

import (
	"time"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/domain/product"
)

type PaymentProduct struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	NetPriceCents int    `json:"netPriceCents"`
	Quantity      int    `json:"quantity"`
}

var paymentProductSchema = z.Struct(z.Shape{
	"ID":            product.IDSchema.Required(),
	"Name":          product.NameSchema.Required(),
	"NetPriceCents": product.NetPriceCentsSchema.Required(),
	"Quantity":      z.Int().GTE(1, z.Message("Quantity must be at least 1")).Required(),
})

type Payment struct {
	ID                string           `json:"id"`
	UserID            int              `json:"userId"`
	TableID           int              `json:"tableId"`
	Products          []PaymentProduct `json:"products"`
	TotalPaymentCents int              `json:"totalPaymentCents"`
	RegisteredAt      time.Time        `json:"registeredAt"`
}

var paymentSchema = z.Struct(z.Shape{
	"ID":                z.String().UUID().Required(),
	"UserID":            z.Int().GTE(1).Required(),
	"TableID":           z.Int().GTE(1).Required(),
	"Products":          z.Slice(paymentProductSchema).Min(1).Required(),
	"TotalPaymentCents": z.Int().GTE(0).Required(),
	"RegisteredAt":      z.Time().Required(),
})
