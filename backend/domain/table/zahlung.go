package table

import (
	"time"

	z "github.com/Oudwins/zog"
)

type Zahlung struct {
	ID                 string
	UserID             int
	TischID            int
	Positionen         []Position
	GesamtZahlungCents int
	Kommentar          string
	KassiertAm         time.Time
}

var zahlungSchema = z.Struct(z.Shape{
	"ID":                 z.String().UUID().Required(),
	"UserID":             z.Int().GTE(1).Required(),
	"TischID":            z.Int().GTE(1).Required(),
	"Positionen":         z.Slice(positionSchema).Min(1).Required(),
	"GesamtZahlungCents": z.Int().GTE(0).Required(),
	"Kommentar":          z.String().Max(100),
	"KassiertAm":         z.Time().Required(),
})
