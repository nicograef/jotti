package kasse

import (
	"time"

	z "github.com/Oudwins/zog"
)

type Zahlung struct {
	ID     string
	UserID int
	// UserName ist der eingefrorene Username des Akteurs aus dem Event-Umschlag,
	// der die Historie beschriftet.
	UserName           string
	TischID            int
	Positionen         []Position
	GesamtZahlungCents int
	Kommentar          string
	KassiertAm         time.Time
}

var zahlungSchema = z.Struct(z.Shape{
	"ID":                 z.String().UUID().Required(),
	"UserID":             z.Int().GTE(1).Required(),
	"UserName":           z.String().Min(1).Required(),
	"TischID":            z.Int().GTE(1).Required(),
	"Positionen":         z.Slice(positionSchema).Min(1).Required(),
	"GesamtZahlungCents": z.Int().GTE(0).Required(),
	"Kommentar":          z.String().Max(100),
	"KassiertAm":         z.Time().Required(),
})
