package kasse

import (
	"time"

	z "github.com/Oudwins/zog"
)

type Stornierung struct {
	ID     string
	UserID int
	// UserName ist der eingefrorene Username des Akteurs aus dem Event-Umschlag,
	// der die Historie beschriftet.
	UserName               string
	TischID                int
	Positionen             []Position
	GesamtStornierungCents int
	Kommentar              string
	StorniertAm            time.Time
	// BarRueckgabe unterscheidet die beiden Storno-Arten und wird zur Lesezeit aus
	// dem Event-Typ abgeleitet (nie im Event gespeichert): true bei der
	// kassenwirksamen Warenrücknahme (`stornierung-erteilt:v1`), false bei der
	// geldneutralen Korrektur (`bestellung-korrigiert:v1`). Spiegelt Namen und
	// Semantik des gleichnamigen Reporting-Felds.
	BarRueckgabe bool
}

var stornierungSchema = z.Struct(z.Shape{
	"ID":                     z.String().UUID().Required(),
	"UserID":                 z.Int().GTE(1).Required(),
	"UserName":               z.String().Min(1).Required(),
	"TischID":                z.Int().GTE(1).Required(),
	"Positionen":             z.Slice(positionSchema).Min(1).Required(),
	"GesamtStornierungCents": z.Int().GTE(0).Required(),
	"Kommentar":              z.String().Min(3).Max(100).Required(),
	"StorniertAm":            z.Time().Required(),
})
