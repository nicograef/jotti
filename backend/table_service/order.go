package table_service

import (
	"time"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/product_admin"
)

type orderProduct struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	NetPriceCents int    `json:"netPriceCents"`
	Quantity      int    `json:"quantity"`
}

var orderProductSchema = z.Struct(z.Shape{
	"ID":            product_admin.IDSchema.Required(),
	"Name":          product_admin.NameSchema.Required(),
	"NetPriceCents": product_admin.NetPriceCentsSchema.Required(),
	"Quantity":      z.Int().GTE(1, z.Message("Quantity must be at least 1")).Required(),
})

type Order struct {
	ID                 string         `json:"id"`
	UserID             int            `json:"userId"`
	TableID            int            `json:"tableId"`
	Products           []orderProduct `json:"products"`
	TotalNetPriceCents int            `json:"totalNetPriceCents"`
	PlacedAt           time.Time      `json:"placedAt"`
}

var orderSchema = z.Struct(z.Shape{
	"ID":                 z.String().UUID().Required(),
	"UserID":             z.Int().GTE(1).Required(),
	"TableID":            z.Int().GTE(1).Required(),
	"Products":           z.Slice(orderProductSchema).Min(1).Required(),
	"TotalNetPriceCents": z.Int().GTE(0).Required(),
	"PlacedAt":           z.Time().Required(),
})
