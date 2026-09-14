package application

import (
	"errors"

	"github.com/nicograef/jotti/backend/db"
)

var ErrDatabase = db.ErrDatabase

var ErrKasseAlreadyOpen = errors.New("kasse bereits geoeffnet")

var ErrKasseNichtGeoeffnet = errors.New("kasse nicht geoeffnet")

// ErrKasseWirdAbgeschlossen signals the transient status 'wird_abgeschlossen' — the Kassenabschluss barrier is active.
var ErrKasseWirdAbgeschlossen = errors.New("kasse wird gerade abgeschlossen")

// ErrConflict is deliberately per-context, not a shared kernel: the HTTP layer maps 409 via
// errors.Is against this exact sentinel, and one sentinel shared across bounded contexts would
// couple them and risk a silent 409-to-500 regression.
var ErrConflict = errors.New("conflict")

var ErrTischeSaldoOffen = errors.New("tische mit offenem saldo")

var ErrBetreiberNichtKonfiguriert = errors.New("betreiber nicht konfiguriert")

// ErrBuchungenNachKassensturz: ein Kassenabschluss-Wiederanlauf fand Buchungen nach dem bereits
// persistierten Kassensturz — der veraltete Ist-Bestand würde sie als Soll-Ist-Differenz verbuchen.
var ErrBuchungenNachKassensturz = errors.New("buchungen nach kassensturz")
