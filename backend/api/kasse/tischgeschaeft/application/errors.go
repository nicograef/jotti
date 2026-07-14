package application

import (
	"errors"

	"github.com/nicograef/jotti/backend/db"
	"github.com/rs/zerolog"
)

// ErrKasseNichtGeoeffnet is returned when an operation requires an open Kassensitzung but none exists.
var ErrKasseNichtGeoeffnet = errors.New("kasse nicht geoeffnet")

// ErrKasseWirdAbgeschlossen is returned when a booking is attempted while the Kassensitzung is in
// the transient 'wird_abgeschlossen' status (the Kassenabschluss barrier is active).
var ErrKasseWirdAbgeschlossen = errors.New("kasse wird gerade abgeschlossen")

// ErrTischNotFound is returned when a tisch is not found.
var ErrTischNotFound = errors.New("tisch not found")

// ErrDatabase is returned when there is a database error.
var ErrDatabase = db.ErrDatabase

// ErrConflict is returned when a concurrent write conflicts with this operation.
var ErrConflict = errors.New("conflict")

// ErrProduktNotFound is returned when a product or variant is not found during enrichment.
var ErrProduktNotFound = errors.New("produkt not found")

// ErrVarianteNichtAktiv is returned when a referenced variant or its product is
// deactivated (inactive). Kept separate from ErrProduktNotFound, which covers
// deleted or non-existent IDs.
var ErrVarianteNichtAktiv = errors.New("variante nicht aktiv")

// ErrTischNotActive is returned when an operation is attempted on an inactive or deleted tisch.
var ErrTischNotActive = errors.New("tisch not active")

// ErrPositionNichtBezahlbar is returned when a position cannot be paid (not in unbezahlt list).
var ErrPositionNichtBezahlbar = errors.New("position nicht bezahlbar")

// ErrPositionNichtStornierbar is returned when a position cannot be cancelled (not in unbezahlt list).
var ErrPositionNichtStornierbar = errors.New("position nicht stornierbar")

// ErrPositionNichtUmbuchbar is returned when a position cannot be moved (not in unbezahlt list).
var ErrPositionNichtUmbuchbar = errors.New("position nicht umbuchbar")

// ErrUmbuchungGleicherTisch is returned when source and target tisch are identical.
var ErrUmbuchungGleicherTisch = errors.New("umbuchung gleicher tisch")

// fromRepositoryError maps repository errors to application-layer errors with structured logging.
// It consolidates tisch-ID context and error mapping at a single location, avoiding duplicated
// error-handling code across the many command methods that load or update tisch records.
func fromRepositoryError(err error, log *zerolog.Logger, id int) error {
	if errors.Is(err, db.ErrNotFound) {
		log.Warn().Err(err).Int("tisch_id", id).Msg("Tisch not found")
		return ErrTischNotFound
	}

	log.Error().Err(err).Int("tisch_id", id).Msg("Database error")
	return ErrDatabase
}
