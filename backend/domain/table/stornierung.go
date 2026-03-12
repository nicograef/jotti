package table

import (
	"time"

	z "github.com/Oudwins/zog"
)

type Stornierung struct {
	ID                     string        `json:"id"`
	UserID                 int           `json:"userId"`
	TischID                int           `json:"tischId"`
	Positionen             []PositionRef `json:"positionen"`
	GesamtStornierungCents int           `json:"gesamtStornierungCents"`
	Kommentar              string        `json:"kommentar"`
	StorniertAm            time.Time     `json:"storniertAm"`
}

var stornierungSchema = z.Struct(z.Shape{
	"ID":                     z.String().UUID().Required(),
	"UserID":                 z.Int().GTE(1).Required(),
	"TischID":                z.Int().GTE(1).Required(),
	"Positionen":             z.Slice(positionRefSchema).Min(1).Required(),
	"GesamtStornierungCents": z.Int().GTE(0).Required(),
	"Kommentar":              z.String().Max(100),
	"StorniertAm":            z.Time().Required(),
})
