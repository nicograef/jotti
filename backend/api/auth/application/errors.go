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

var ErrPasswordTooWeak = errors.New("password too weak")

var ErrTokenGeneration = errors.New("token generation failed")

var ErrDatabase = db.ErrDatabase
