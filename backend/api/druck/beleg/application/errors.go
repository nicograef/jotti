package application

import (
	"errors"

	"github.com/nicograef/jotti/backend/db"
)

// ErrDatabase is returned when there is a database error.
var ErrDatabase = db.ErrDatabase

// ErrKasseNichtGeoeffnet is returned when an operation requires an open Kassensitzung but none exists.
var ErrKasseNichtGeoeffnet = errors.New("kasse nicht geoeffnet")

// ErrKasseWirdAbgeschlossen is returned when a beleg is requested while the Kassensitzung is in
// the transient 'wird_abgeschlossen' status (the Kassenabschluss barrier is active).
var ErrKasseWirdAbgeschlossen = errors.New("kasse wird gerade abgeschlossen")

// ErrZahlungNichtGefunden is returned when a requested payment reference does not exist.
var ErrZahlungNichtGefunden = errors.New("zahlung nicht gefunden")

// ErrVerkaufNichtGefunden is returned when a requested Direktverkauf reference does not exist.
var ErrVerkaufNichtGefunden = errors.New("verkauf nicht gefunden")

// ErrStornierungNichtGefunden is returned when a requested Direktverkauf-Stornierung does not exist.
var ErrStornierungNichtGefunden = errors.New("stornierung nicht gefunden")

// ErrKassenbelegDruckerNichtKonfiguriert is returned when no receipt printer IP is configured.
var ErrKassenbelegDruckerNichtKonfiguriert = errors.New("kassenbeleg drucker nicht konfiguriert")
