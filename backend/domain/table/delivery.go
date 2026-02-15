package table

import (
	"time"

	z "github.com/Oudwins/zog"
)

type Delivery struct {
	ID          string         `json:"id"`
	UserID      int            `json:"userId"`
	TableID     int            `json:"tableId"`
	Variants    []LineItem `json:"variants"`
	Comment     string         `json:"comment"`
	DeliveredAt time.Time      `json:"deliveredAt"`
}

var deliverySchema = z.Struct(z.Shape{
	"ID":          z.String().UUID().Required(),
	"UserID":      z.Int().GTE(1).Required(),
	"TableID":     z.Int().GTE(1).Required(),
	"Variants":    z.Slice(lineItemSchema).Min(1).Required(),
	"Comment":     z.String().Max(100),
	"DeliveredAt": z.Time().Required(),
})
