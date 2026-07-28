package application

import (
	"errors"

	"github.com/nicograef/jotti/backend/db"
)

var ErrDatabase = db.ErrDatabase

// ErrKasseAlreadyOpen is returned when trying to open a Kassensitzung but one is already open.
var ErrKasseAlreadyOpen = errors.New("kasse bereits geoeffnet")

// ErrKasseNichtGeoeffnet is returned when an operation requires an open Kassensitzung but none exists.
var ErrKasseNichtGeoeffnet = errors.New("kasse nicht geoeffnet")

// ErrKasseWirdAbgeschlossen is returned when a booking is attempted while the Kassensitzung is in
// the transient 'wird_abgeschlossen' status (the Kassenabschluss barrier is active).
var ErrKasseWirdAbgeschlossen = errors.New("kasse wird gerade abgeschlossen")

// ErrKasseAlreadyAbgeschlossen is returned when a Kassensitzung is already closed.
var ErrKasseAlreadyAbgeschlossen = errors.New("kasse bereits abgeschlossen")

// ErrConflict is returned on a concurrent write conflict.
// Deliberately per-context, not a shared kernel: errors.Is against this exact sentinel is what the
// HTTP layer relies on to map the error to 409; a shared sentinel across bounded contexts would
// couple them and risk a silent 409-to-500 regression (2026-07-17 multi-expert review).
var ErrConflict = errors.New("conflict")

// ErrVorgangDatenAbweichend is returned when a known geldtransitId is submitted again with
// different Nutzdaten. Weder eine zweite Buchung (Doppelbuchung) noch eine stille
// Erfolgsantwort (verschluckte Einreichung) wäre richtig; die HTTP-Schicht bildet den
// Fehler deshalb auf 409 vorgang_daten_abweichend ab.
var ErrVorgangDatenAbweichend = errors.New("vorgang daten abweichend")

// ErrTischeSaldoOffen is returned when a Kassenabschluss is attempted but tisch sessions have non-zero saldi.
var ErrTischeSaldoOffen = errors.New("tische mit offenem saldo")

// ErrBetreiberNichtKonfiguriert is returned when a Kassensitzung is opened but betreiber data is incomplete.
var ErrBetreiberNichtKonfiguriert = errors.New("betreiber nicht konfiguriert")

// ErrBuchungenNachKassensturz is returned when a Kassenabschluss retry finds bookings that were
// recorded after the already-persisted Kassensturz. Reusing the stale Ist-Bestand would book those
// legitimate turnovers as a fake Soll-Ist-Differenz, so the retry aborts instead.
var ErrBuchungenNachKassensturz = errors.New("buchungen nach kassensturz")
