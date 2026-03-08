package table

import (
	"time"

	z "github.com/Oudwins/zog"
)

type Zahlung struct {
	ID                 string     `json:"id"`
	UserID             int        `json:"userId"`
	TischID            int        `json:"tischId"`
	Positionen         []Position `json:"positionen"`
	GesamtZahlungCents int        `json:"gesamtZahlungCents"`
	Comment            string     `json:"comment"`
	RegistriertAm      time.Time  `json:"registriertAm"`
}

var zahlungSchema = z.Struct(z.Shape{
	"ID":                 z.String().UUID().Required(),
	"UserID":             z.Int().GTE(1).Required(),
	"TischID":            z.Int().GTE(1).Required(),
	"Positionen":         z.Slice(positionSchema).Min(1).Required(),
	"GesamtZahlungCents": z.Int().GTE(0).Required(),
	"Comment":            z.String().Max(100),
	"RegistriertAm":      z.Time().Required(),
})
