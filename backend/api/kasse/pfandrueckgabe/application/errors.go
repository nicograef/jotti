package application

import (
	"errors"

	"github.com/nicograef/jotti/backend/api/kasse/kassenfehler"
	"github.com/nicograef/jotti/backend/db"
)

var ErrKasseNichtGeoeffnet = errors.New("kasse nicht geoeffnet")

// ErrKasseWirdAbgeschlossen: barrier status, see kasse.KassensitzungWirdAbgeschlossen.
var ErrKasseWirdAbgeschlossen = errors.New("kasse wird gerade abgeschlossen")

var ErrTischNichtGefunden = errors.New("tisch nicht gefunden")

var ErrTischNichtAktiv = errors.New("tisch nicht aktiv")

var ErrRueckgabeNichtGefunden = errors.New("pfandrueckgabe nicht gefunden")

// ErrPfandrueckgabeAbweichend: the rueckgabeId or the stornierungId is booked, but not with the
// request's content, or the Rückgabe was settled with a Zahlung or Verkauf.
var ErrPfandrueckgabeAbweichend = errors.New("pfandrueckgabe abweichend")

var ErrPositionNichtStornierbar = errors.New("position nicht stornierbar")

// ErrConflict is a per-context sentinel, see tischgeschaeft/application/errors.go.
var ErrConflict = errors.New("conflict")

var ErrDatabase = db.ErrDatabase

var sentinels = kassenfehler.Sentinels{KasseNichtGeoeffnet: ErrKasseNichtGeoeffnet, KasseWirdAbgeschlossen: ErrKasseWirdAbgeschlossen, Conflict: ErrConflict}
