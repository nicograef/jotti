package table

import (
	"time"

	z "github.com/Oudwins/zog"
)

type Cancelation struct {
	ID                    string     `json:"id"`
	UserID                int        `json:"userId"`
	TableID               int        `json:"tableId"`
	Variants              []LineItem `json:"variants"`
	TotalCancelationCents int        `json:"totalCancelationCents"`
	Comment               string     `json:"comment"`
	CanceledAt            time.Time  `json:"canceledAt"`
}

var cancelationSchema = z.Struct(z.Shape{
	"ID":                    z.String().UUID().Required(),
	"UserID":                z.Int().GTE(1).Required(),
	"TableID":               z.Int().GTE(1).Required(),
	"Variants":              z.Slice(lineItemSchema).Min(1).Required(),
	"TotalCancelationCents": z.Int().GTE(0).Required(),
	"Comment":               z.String().Max(100),
	"CanceledAt":            z.Time().Required(),
})
