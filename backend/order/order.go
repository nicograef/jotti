package order

import (
	"errors"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/product"
)

type orderProduct struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	NetPriceCents int    `json:"netPriceCents"`
	Quantity      int    `json:"quantity"`
}

var orderProductSchema = z.Struct(z.Shape{
	"ID":            product.IDSchema.Required(),
	"Name":          product.NameSchema.Required(),
	"NetPriceCents": product.NetPriceCentsSchema.Required(),
	"Quantity":      z.Int().GTE(1, z.Message("Quantity must be at least 1")).Required(),
})

type Order struct {
	ID                 int            `json:"id"`
	UserID             int            `json:"userId"`
	TableID            int            `json:"tableId"`
	Products           []orderProduct `json:"products"`
	TotalNetPriceCents int            `json:"totalNetPriceCents"`
	PlacedAt           time.Time      `json:"placedAt"`
}

// ErrDatabase is returned when there is a database error.
var ErrDatabase = errors.New("database error")
