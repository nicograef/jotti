package kasse

import (
	"time"

	z "github.com/Oudwins/zog"
)

type Ausgabe struct {
	ID           string
	UserID       int
	TischID      int
	Positionen   []Position
	Kommentar    string
	AusgegebenAm time.Time
}

var ausgabeSchema = z.Struct(z.Shape{
	"ID":           z.String().UUID().Required(),
	"UserID":       z.Int().GTE(1).Required(),
	"TischID":      z.Int().GTE(1).Required(),
	"Positionen":   z.Slice(positionSchema).Min(1).Required(),
	"Kommentar":    z.String().Max(100),
	"AusgegebenAm": z.Time().Required(),
})
