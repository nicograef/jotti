package kasse

import (
	"time"

	z "github.com/Oudwins/zog"
)

type Stornierung struct {
	ID                     string
	UserID                 int
	TischID                int
	Positionen             []Position
	GesamtStornierungCents int
	Kommentar              string
	StorniertAm            time.Time
}

var stornierungSchema = z.Struct(z.Shape{
	"ID":                     z.String().UUID().Required(),
	"UserID":                 z.Int().GTE(1).Required(),
	"TischID":                z.Int().GTE(1).Required(),
	"Positionen":             z.Slice(positionSchema).Min(1).Required(),
	"GesamtStornierungCents": z.Int().GTE(0).Required(),
	"Kommentar":              z.String().Min(3).Max(100).Required(),
	"StorniertAm":            z.Time().Required(),
})
