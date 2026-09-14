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

var ErrOnetimePasswordLocked = errors.New("onetime password locked")

var ErrPasswordTooWeak = errors.New("password too weak")

var ErrTokenGeneration = errors.New("token generation failed")

// ErrLoginThrottled ist bewusst von ErrInvalidPassword getrennt, damit der
// Handler HTTP 429 statt "ungültige Zugangsdaten" liefern kann.
var ErrLoginThrottled = errors.New("login throttled")

var ErrDatabase = db.ErrDatabase
