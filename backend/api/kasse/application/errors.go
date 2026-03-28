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

// ErrKasseAlreadyAbgeschlossen is returned when a Kassensitzung is already closed.
var ErrKasseAlreadyAbgeschlossen = errors.New("kasse bereits abgeschlossen")

// ErrKonflikt is returned on a concurrent write conflict.
var ErrKonflikt = errors.New("konflikt")

// ErrKassensturzErforderlich is returned when a Tagesabschluss is attempted without prior Kassensturz.
var ErrKassensturzErforderlich = errors.New("kassensturz erforderlich")

// ErrTischeSaldoOffen is returned when a Tagesabschluss is attempted but tisch sessions have non-zero saldi.
var ErrTischeSaldoOffen = errors.New("tische mit offenem saldo")
