package application

import (
	"errors"

	"github.com/nicograef/jotti/backend/db"
)

var ErrUserNotFound = errors.New("user not found")

var ErrNotActive = errors.New("user not active")

var ErrNoPassword = errors.New("no password set")

var ErrInvalidPassword = errors.New("invalid password")

var ErrNoOnetimePassword = errors.New("no onetime password set")

// ErrOnetimePasswordLocked: das Einmalpasswort wurde nach zu vielen Fehlversuchen
// ungueltig; der Admin muss ein neues erzeugen.
var ErrOnetimePasswordLocked = errors.New("onetime password locked")

var ErrPasswordTooWeak = errors.New("password too weak")

var ErrTokenGeneration = errors.New("token generation failed")

// ErrLoginThrottled: für dieses Konto sind zu viele Fehlanmeldungen aufgelaufen;
// der nächste Versuch ist kurz gedrosselt (Soft-Throttle, läuft von selbst ab).
// Bewusst getrennt von ErrInvalidPassword, damit der Handler eine klare Meldung
// (HTTP 429) statt "ungültige Zugangsdaten" liefern kann.
var ErrLoginThrottled = errors.New("login throttled")

var ErrDatabase = db.ErrDatabase
