package kasse

import (
	"time"

	z "github.com/Oudwins/zog"
)

type Auszahlung struct {
	ID     string
	UserID int
	// UserName ist der eingefrorene Username des Akteurs aus dem Event-Umschlag,
	// der die Historie beschriftet.
	UserName    string
	TischID     int
	BetragCents int
	Kommentar   string
	GeleistetAm time.Time
}

var auszahlungSchema = z.Struct(z.Shape{
	"ID":          z.String().UUID().Required(),
	"UserID":      z.Int().GTE(1).Required(),
	"UserName":    z.String().Min(1).Required(),
	"TischID":     z.Int().GTE(1).Required(),
	"BetragCents": z.Int().GTE(1).Required(),
	"Kommentar":   z.String().Min(3).Max(100).Required(),
	"GeleistetAm": z.Time().Required(),
})
