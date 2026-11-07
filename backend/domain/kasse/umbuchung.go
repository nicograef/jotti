package kasse

import (
	"time"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/internal/zeichen"
)

// Umbuchung ist der Historien-Eintrag einer geldneutralen Umbuchung, wie er auf
// Quell- und Zieltisch erscheint.
type Umbuchung struct {
	ID     string
	UserID int
	// UserName ist der eingefrorene Username des Akteurs aus dem Event-Umschlag.
	UserName     string
	TischID      int
	QuellTischID int
	ZielTischID  int
	Positionen   []Position
	GesamtCents  int
	// Kommentar ist der Richtungs-Autotext, BenutzerKommentar der optionale freie Text.
	Kommentar         string
	BenutzerKommentar string
	UmgebuchtAm       time.Time
}

func (u Umbuchung) IstZugang() bool {
	return u.TischID == u.ZielTischID
}

var umbuchungSchema = z.Struct(z.Shape{
	"ID":                z.String().UUID().Required(),
	"UserID":            z.Int().GTE(1).Required(),
	"UserName":          z.String().TestFunc(zeichen.Min(1)).Required(),
	"TischID":           z.Int().GTE(1).Required(),
	"QuellTischID":      z.Int().GTE(1).Required(),
	"ZielTischID":       z.Int().GTE(1).Required(),
	"Positionen":        z.Slice(positionSchema).Min(1).Required(),
	"GesamtCents":       z.Int().GTE(1).Required(),
	"Kommentar":         z.String().TestFunc(zeichen.Max(100)),
	"BenutzerKommentar": z.String().TestFunc(zeichen.Max(100)),
})
