package application

import (
	"errors"

	"github.com/nicograef/jotti/backend/db"
)

var ErrKasseNichtGeoeffnet = errors.New("kasse nicht geoeffnet")

// ErrKasseWirdAbgeschlossen signals the transient status 'wird_abgeschlossen' — the Kassenabschluss barrier is active.
var ErrKasseWirdAbgeschlossen = errors.New("kasse wird gerade abgeschlossen")

var ErrTischNichtGefunden = errors.New("tisch nicht gefunden")

var ErrTischNichtAktiv = errors.New("tisch nicht aktiv")

var ErrRueckgabeNichtGefunden = errors.New("pfandrueckgabe nicht gefunden")

// ErrPfandrueckgabeAbweichend: the rueckgabeId or the stornierungId is booked, but not with the
// request's content, or the Rückgabe was settled with a Zahlung or Verkauf.
var ErrPfandrueckgabeAbweichend = errors.New("pfandrueckgabe abweichend")

var ErrPositionNichtStornierbar = errors.New("position nicht stornierbar")

// ErrConflict is deliberately per-context, not a shared kernel: the HTTP layer maps 409 via
// errors.Is against this exact sentinel, and one sentinel shared across bounded contexts would
// couple them and risk a silent 409-to-500 regression.
var ErrConflict = errors.New("conflict")

var ErrDatabase = db.ErrDatabase
