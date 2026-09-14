package application

import (
	"errors"

	"github.com/nicograef/jotti/backend/db"
)

var ErrDatabase = db.ErrDatabase

var ErrKasseNichtGeoeffnet = errors.New("kasse nicht geoeffnet")

// ErrKasseWirdAbgeschlossen: the Kassensitzung is in the transient status
// 'wird_abgeschlossen' — the Kassenabschluss barrier is active.
var ErrKasseWirdAbgeschlossen = errors.New("kasse wird gerade abgeschlossen")

var ErrZahlungNichtGefunden = errors.New("zahlung nicht gefunden")

var ErrVerkaufNichtGefunden = errors.New("verkauf nicht gefunden")

var ErrStornierungNichtGefunden = errors.New("stornierung nicht gefunden")

var ErrKassenbelegDruckerNichtKonfiguriert = errors.New("kassenbeleg drucker nicht konfiguriert")
