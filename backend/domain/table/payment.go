package table

import (
	"time"

	z "github.com/Oudwins/zog"
)

type Payment struct {
	ID                string         `json:"id"`
	UserID            int            `json:"userId"`
	TableID           int            `json:"tableId"`
	Variants          []OrderVariant `json:"variants"`
	TotalPaymentCents int            `json:"totalPaymentCents"`
	Comment           string         `json:"comment"`
	RegisteredAt      time.Time      `json:"registeredAt"`
}

var paymentSchema = z.Struct(z.Shape{
	"ID":                z.String().UUID().Required(),
	"UserID":            z.Int().GTE(1).Required(),
	"TableID":           z.Int().GTE(1).Required(),
	"Variants":          z.Slice(orderVariantSchema).Min(1).Required(),
	"TotalPaymentCents": z.Int().GTE(0).Required(),
	"Comment":           z.String().Max(100),
	"RegisteredAt":      z.Time().Required(),
})
