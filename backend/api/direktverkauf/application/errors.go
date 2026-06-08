package application

import (
	"errors"

	"github.com/nicograef/jotti/backend/db"
)

// ErrKasseNichtGeoeffnet is returned when no open Kassensitzung exists.
var ErrKasseNichtGeoeffnet = errors.New("kasse nicht geoeffnet")

// ErrProduktNotFound is returned when a product or variant is not found during enrichment.
var ErrProduktNotFound = errors.New("produkt not found")

// ErrVerkaufNichtGefunden is returned when no Direktverkauf exists for the given verkaufId.
var ErrVerkaufNichtGefunden = errors.New("verkauf nicht gefunden")

// ErrPositionNichtStornierbar is returned when requested positions are not (or no longer) cancellable.
var ErrPositionNichtStornierbar = errors.New("position nicht stornierbar")

// ErrConflict is returned when a concurrent write conflicts with this operation.
var ErrConflict = errors.New("conflict")

// ErrDatabase is returned when there is a database error.
var ErrDatabase = db.ErrDatabase
