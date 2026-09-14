package application

import (
	"errors"

	"github.com/nicograef/jotti/backend/db"
)

var ErrKasseNichtGeoeffnet = errors.New("kasse nicht geoeffnet")

// ErrKasseWirdAbgeschlossen signals the transient status 'wird_abgeschlossen' — the Kassenabschluss barrier is active.
var ErrKasseWirdAbgeschlossen = errors.New("kasse wird gerade abgeschlossen")

var ErrVerkaufNichtGefunden = errors.New("verkauf nicht gefunden")

var ErrPositionNichtStornierbar = errors.New("position nicht stornierbar")

// ErrConflict is deliberately per-context, not a shared kernel: the HTTP layer maps 409 via
// errors.Is against this exact sentinel, and one sentinel shared across bounded contexts would
// couple them and risk a silent 409-to-500 regression.
var ErrConflict = errors.New("conflict")

var ErrDatabase = db.ErrDatabase
