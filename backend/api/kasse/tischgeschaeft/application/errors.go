package application

import (
	"errors"

	"github.com/nicograef/jotti/backend/db"
	"github.com/rs/zerolog"
)

var ErrKasseNichtGeoeffnet = errors.New("kasse nicht geoeffnet")

// ErrKasseWirdAbgeschlossen signals the transient status 'wird_abgeschlossen' — the Kassenabschluss barrier is active.
var ErrKasseWirdAbgeschlossen = errors.New("kasse wird gerade abgeschlossen")

var ErrTischNotFound = errors.New("tisch not found")

var ErrDatabase = db.ErrDatabase

// ErrConflict is deliberately per-context, not a shared kernel: the HTTP layer maps 409 via
// errors.Is against this exact sentinel, and one sentinel shared across bounded contexts would
// couple them and risk a silent 409-to-500 regression.
var ErrConflict = errors.New("conflict")

var ErrTischNotActive = errors.New("tisch not active")

var ErrPositionNichtBezahlbar = errors.New("position nicht bezahlbar")

var ErrPositionNichtStornierbar = errors.New("position nicht stornierbar")

var ErrPositionNichtUmbuchbar = errors.New("position nicht umbuchbar")

var ErrUmbuchungGleicherTisch = errors.New("umbuchung gleicher tisch")

func fromRepositoryError(err error, log *zerolog.Logger, id int) error {
	if errors.Is(err, db.ErrNotFound) {
		log.Warn().Err(err).Int("tisch_id", id).Msg("Tisch not found")
		return ErrTischNotFound
	}

	log.Error().Err(err).Int("tisch_id", id).Msg("Database error")
	return ErrDatabase
}
