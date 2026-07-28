package application

import (
	"errors"

	"github.com/nicograef/jotti/backend/db"
)

// ErrKasseNichtGeoeffnet is returned when no open Kassensitzung exists.
var ErrKasseNichtGeoeffnet = errors.New("kasse nicht geoeffnet")

// ErrKasseWirdAbgeschlossen is returned when a Direktverkauf is attempted while the Kassensitzung is
// in the transient 'wird_abgeschlossen' status (the Kassenabschluss barrier is active).
var ErrKasseWirdAbgeschlossen = errors.New("kasse wird gerade abgeschlossen")

// ErrVerkaufNichtGefunden is returned when no Direktverkauf exists for the given verkaufId.
var ErrVerkaufNichtGefunden = errors.New("verkauf nicht gefunden")

// ErrPositionNichtStornierbar is returned when requested positions are not (or no longer) cancellable.
var ErrPositionNichtStornierbar = errors.New("position nicht stornierbar")

// ErrVorgangDatenAbweichend is returned when a known vorgangId is submitted again with
// different Nutzdaten. Beide stillen Ausgänge wären hier falsch: Eine zweite Buchung bucht
// doppelt, eine Erfolgsantwort verschluckt die geänderte Einreichung. Die HTTP-Schicht
// bildet den Fehler deshalb auf 409 vorgang_daten_abweichend ab.
var ErrVorgangDatenAbweichend = errors.New("vorgang daten abweichend")

// ErrConflict is returned when a concurrent write conflicts with this operation.
// Deliberately per-context, not a shared kernel: errors.Is against this exact sentinel is what the
// HTTP layer relies on to map the error to 409; a shared sentinel across bounded contexts would
// couple them and risk a silent 409-to-500 regression (2026-07-17 multi-expert review).
var ErrConflict = errors.New("conflict")

// ErrDatabase is returned when there is a database error.
var ErrDatabase = db.ErrDatabase
