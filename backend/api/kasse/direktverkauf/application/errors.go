package application

import (
	"errors"

	"github.com/nicograef/jotti/backend/api/kasse/kassenfehler"
	"github.com/nicograef/jotti/backend/db"
)

var ErrKasseNichtGeoeffnet = errors.New("kasse nicht geoeffnet")

// ErrKasseWirdAbgeschlossen: barrier status, see kasse.KassensitzungWirdAbgeschlossen.
var ErrKasseWirdAbgeschlossen = errors.New("kasse wird gerade abgeschlossen")

var ErrVerkaufNichtGefunden = errors.New("verkauf nicht gefunden")

// ErrPfandrueckgabeAbweichend: the rueckgabeId or the verkaufId is booked, but not with the
// request's content or without its settled partner.
var ErrPfandrueckgabeAbweichend = errors.New("pfandrueckgabe abweichend")

// ErrVorgangAbweichend: the client's stornierungId is booked, but with other content.
var ErrVorgangAbweichend = errors.New("vorgang abweichend")

var ErrPositionNichtStornierbar = errors.New("position nicht stornierbar")

// ErrConflict is a per-context sentinel, see tischgeschaeft/application/errors.go.
var ErrConflict = errors.New("conflict")

var ErrDatabase = db.ErrDatabase

var sentinels = kassenfehler.Sentinels{KasseNichtGeoeffnet: ErrKasseNichtGeoeffnet, KasseWirdAbgeschlossen: ErrKasseWirdAbgeschlossen, Conflict: ErrConflict}
