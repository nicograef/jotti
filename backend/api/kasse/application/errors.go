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

// ErrKonflikt is returned on a concurrent write conflict.
var ErrKonflikt = errors.New("konflikt")

// ErrTischeSaldoOffen is returned when a Kassenabschluss is attempted but tisch sessions have non-zero saldi.
var ErrTischeSaldoOffen = errors.New("tische mit offenem saldo")

// ErrBetreiberNichtKonfiguriert is returned when a Kassensitzung is opened but betreiber data is incomplete.
var ErrBetreiberNichtKonfiguriert = errors.New("betreiber nicht konfiguriert")
