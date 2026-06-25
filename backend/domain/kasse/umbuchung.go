package kasse

import (
	"time"

	z "github.com/Oudwins/zog"
)

// Umbuchung ist ein Historien-Eintrag einer geldneutralen Umbuchung, wie er auf
// Quell- oder Zieltisch erscheint. Richtung (Abgang/Zugang) ergibt sich aus dem
// Verhältnis von TischID zu QuellTischID/ZielTischID.
type Umbuchung struct {
	ID     string
	UserID int
	// UserName ist der eingefrorene Username des Akteurs aus dem Event-Umschlag,
	// der die Historie beschriftet.
	UserName     string
	TischID      int
	QuellTischID int
	ZielTischID  int
	Positionen   []Position
	GesamtCents  int
	Kommentar    string
	UmgebuchtAm  time.Time
}

// IstZugang meldet, ob dieser Eintrag den Zugang auf dem Zieltisch beschreibt
// (Positionen kommen hinzu) statt den Abgang auf dem Quelltisch.
func (u Umbuchung) IstZugang() bool {
	return u.TischID == u.ZielTischID
}

var umbuchungSchema = z.Struct(z.Shape{
	"ID":           z.String().UUID().Required(),
	"UserID":       z.Int().GTE(1).Required(),
	"UserName":     z.String().Min(1).Required(),
	"TischID":      z.Int().GTE(1).Required(),
	"QuellTischID": z.Int().GTE(1).Required(),
	"ZielTischID":  z.Int().GTE(1).Required(),
	"Positionen":   z.Slice(positionSchema).Min(1).Required(),
	"GesamtCents":  z.Int().GTE(0).Required(),
	"Kommentar":    z.String().Max(100),
})
