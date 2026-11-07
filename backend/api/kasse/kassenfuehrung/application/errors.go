package application

import (
	"errors"

	"github.com/nicograef/jotti/backend/api/kasse/kassenfehler"
	"github.com/nicograef/jotti/backend/db"
)

var ErrDatabase = db.ErrDatabase

var ErrKasseAlreadyOpen = errors.New("kasse bereits geoeffnet")

var ErrKasseNichtGeoeffnet = errors.New("kasse nicht geoeffnet")

// ErrKasseWirdAbgeschlossen: barrier status, see kasse.KassensitzungWirdAbgeschlossen.
var ErrKasseWirdAbgeschlossen = errors.New("kasse wird gerade abgeschlossen")

// ErrConflict is a per-context sentinel, see tischgeschaeft/application/errors.go.
var ErrConflict = errors.New("conflict")

var ErrTischeSaldoOffen = errors.New("tische mit offenem saldo")

var ErrBetreiberNichtKonfiguriert = errors.New("betreiber nicht konfiguriert")

// ErrBuchungenNachKassensturz: ein Kassenabschluss-Wiederanlauf fand Buchungen nach dem bereits
// persistierten Kassensturz — der veraltete Ist-Bestand würde sie als Soll-Ist-Differenz verbuchen.
var ErrBuchungenNachKassensturz = errors.New("buchungen nach kassensturz")

var sentinels = kassenfehler.Sentinels{KasseNichtGeoeffnet: ErrKasseNichtGeoeffnet, KasseWirdAbgeschlossen: ErrKasseWirdAbgeschlossen, Conflict: ErrConflict}
