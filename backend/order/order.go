package order

import (
	"errors"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/product"
)

type EventType string

const (
	// EventTypeOrderPlaced represents an order placed event.
	EventTypeOrderPlacedV1 EventType = "jotti.order.placed:v1"
)

type OrderPlacedData struct {
	Products        []OrderProduct `json:"products"`
	TotalPriceCents int            `json:"totalPriceCents"`
}

// OrderProduct represents a product within an order.
type OrderProduct struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	NetPriceCents int    `json:"netPriceCents"`
	Quantity      int    `json:"quantity"`
}

var OrderProductSchema = z.Struct(z.Shape{
	"ID":            product.IDSchema.Required(),
	"Name":          product.NameSchema.Required(),
	"NetPriceCents": product.NetPriceCentsSchema.Required(),
	"Quantity":      z.Int().GTE(1, z.Message("Quantity must be at least 1")).Required(),
})

// Order represents an order aggregate model.
type Order struct {
	ID                 uuid.UUID      `json:"id"`
	UserID             int            `json:"userId"`
	TableID            int            `json:"tableId"`
	Products           []OrderProduct `json:"products"`
	TotalNetPriceCents int            `json:"totalNetPriceCents"`
	PlacedAt           time.Time      `json:"placedAt"`
}

// ErrDatabase is returned when there is a database error.
var ErrDatabase = errors.New("database error")
