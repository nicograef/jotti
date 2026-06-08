package application

import (
	"errors"

	"github.com/nicograef/jotti/backend/db"
	"github.com/rs/zerolog"
)

// ErrKasseNichtGeoeffnet is returned when an operation requires an open Kassensitzung but none exists.
var ErrKasseNichtGeoeffnet = errors.New("kasse nicht geoeffnet")

// ErrTischNotFound is returned when a tisch is not found.
var ErrTischNotFound = errors.New("tisch not found")

// ErrTischAlreadyExists is returned when a tisch already exists.
var ErrTischAlreadyExists = errors.New("tisch already exists")

// ErrDatabase is returned when there is a database error.
var ErrDatabase = db.ErrDatabase

// ErrInvalidTischData is returned when the provided tisch data is invalid.
var ErrInvalidTischData = errors.New("invalid tisch data")

// ErrConflict is returned when a concurrent write conflicts with this operation.
var ErrConflict = errors.New("conflict")

// ErrProduktNotFound is returned when a product or variant is not found during enrichment.
var ErrProduktNotFound = errors.New("produkt not found")

// ErrUserNotFound is returned when the user cannot be found for event creation.
var ErrUserNotFound = errors.New("user not found")

// ErrTischNotActive is returned when an operation is attempted on an inactive or deleted tisch.
var ErrTischNotActive = errors.New("tisch not active")

// ErrInvalidPositionen is returned when the provided positions are not valid.
var ErrInvalidPositionen = errors.New("invalid positionen")

// ErrPositionNichtBezahlbar is returned when a position cannot be paid (not in unbezahlt list).
var ErrPositionNichtBezahlbar = errors.New("position nicht bezahlbar")

// ErrPositionNichtAusgebbar is returned when a position cannot be issued (not in ausstehend list).
var ErrPositionNichtAusgebbar = errors.New("position nicht ausgebbar")

// ErrPositionNichtStornierbar is returned when a position cannot be cancelled (not in unbezahlt list).
var ErrPositionNichtStornierbar = errors.New("position nicht stornierbar")

// ErrZahlungNichtGefunden is returned when a requested payment reference does not exist.
var ErrZahlungNichtGefunden = errors.New("zahlung nicht gefunden")

// ErrVerkaufNichtGefunden is returned when a requested Direktverkauf reference does not exist.
var ErrVerkaufNichtGefunden = errors.New("verkauf nicht gefunden")

// ErrKassenbelegDruckerNichtKonfiguriert is returned when no receipt printer IP is configured.
var ErrKassenbelegDruckerNichtKonfiguriert = errors.New("kassenbeleg drucker nicht konfiguriert")

// fromRepositoryError maps repository errors to application-layer errors with structured logging.
// It consolidates tisch-ID context and error mapping at a single location, avoiding duplicated
// error-handling code across the many command methods that load or update tisch records.
func fromRepositoryError(err error, log *zerolog.Logger, id int) error {
	if errors.Is(err, db.ErrNotFound) {
		log.Warn().Err(err).Int("tisch_id", id).Msg("Tisch not found")
		return ErrTischNotFound
	}

	if errors.Is(err, db.ErrAlreadyExists) {
		log.Warn().Err(err).Msg("Tisch already exists")
		return ErrTischAlreadyExists
	}

	log.Error().Err(err).Int("tisch_id", id).Msg("Database error")
	return ErrDatabase
}
