package table

import (
	"time"

	z "github.com/Oudwins/zog"
)

type Lieferung struct {
	ID          string     `json:"id"`
	UserID      int        `json:"userId"`
	TischID     int        `json:"tischId"`
	Positionen  []Position `json:"positionen"`
	Comment     string     `json:"comment"`
	GeliefertAm time.Time  `json:"geliefertAm"`
}

var lieferungSchema = z.Struct(z.Shape{
	"ID":          z.String().UUID().Required(),
	"UserID":      z.Int().GTE(1).Required(),
	"TischID":     z.Int().GTE(1).Required(),
	"Positionen":  z.Slice(positionSchema).Min(1).Required(),
	"Comment":     z.String().Max(100),
	"GeliefertAm": z.Time().Required(),
})
