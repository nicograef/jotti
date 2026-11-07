package application

import (
	"errors"
	"log/slog"

	"github.com/nicograef/jotti/backend/api/kasse/kassenfehler"
	"github.com/nicograef/jotti/backend/db"
)

var ErrKasseNichtGeoeffnet = errors.New("kasse nicht geoeffnet")

// ErrKasseWirdAbgeschlossen: barrier status, see kasse.KassensitzungWirdAbgeschlossen.
var ErrKasseWirdAbgeschlossen = errors.New("kasse wird gerade abgeschlossen")

var ErrTischNotFound = errors.New("tisch not found")

var ErrDatabase = db.ErrDatabase

// ErrConflict is deliberately per-context, not a shared kernel: the HTTP layer maps 409 via
// errors.Is against this exact sentinel, and one sentinel shared across bounded contexts would
// couple them and risk a silent 409-to-500 regression.
var ErrConflict = errors.New("conflict")

var ErrTischNotActive = errors.New("tisch not active")

var ErrPositionNichtBezahlbar = errors.New("position nicht bezahlbar")

// ErrPfandrueckgabeAbweichend: the rueckgabeId is booked, but not with the request's content or
// without the payment it was settled with.
var ErrPfandrueckgabeAbweichend = errors.New("pfandrueckgabe abweichend")

// ErrVorgangAbweichend: the client's key is booked, but with other content. Answering success would
// hide that the requested Vorgang never happened; booking it would reuse the key.
var ErrVorgangAbweichend = errors.New("vorgang abweichend")

var ErrPositionNichtStornierbar = errors.New("position nicht stornierbar")

var ErrPositionNichtUmbuchbar = errors.New("position nicht umbuchbar")

var ErrUmbuchungGleicherTisch = errors.New("umbuchung gleicher tisch")

func fromRepositoryError(err error, log *slog.Logger, id int) error {
	if errors.Is(err, db.ErrNotFound) {
		log.Warn("Tisch not found", "error", err, "tisch_id", id)
		return ErrTischNotFound
	}

	log.Error("Database error", "error", err, "tisch_id", id)
	return ErrDatabase
}

var sentinels = kassenfehler.Sentinels{KasseNichtGeoeffnet: ErrKasseNichtGeoeffnet, KasseWirdAbgeschlossen: ErrKasseWirdAbgeschlossen, Conflict: ErrConflict, TischNichtAktiv: ErrTischNotActive}
