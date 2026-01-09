package table

import (
	"time"

	z "github.com/Oudwins/zog"
)

type Delivery struct {
	ID          string         `json:"id"`
	UserID      int            `json:"userId"`
	TableID     int            `json:"tableId"`
	Products    []OrderProduct `json:"products"`
	DeliveredAt time.Time      `json:"deliveredAt"`
}

var deliverySchema = z.Struct(z.Shape{
	"ID":          z.String().UUID().Required(),
	"UserID":      z.Int().GTE(1).Required(),
	"TableID":     z.Int().GTE(1).Required(),
	"Products":    z.Slice(orderProductSchema).Min(1).Required(),
	"DeliveredAt": z.Time().Required(),
})
