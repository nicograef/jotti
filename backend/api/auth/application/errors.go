package application

import (
	"errors"
)

var ErrUserNotFound = errors.New("user not found")

var ErrNotActive = errors.New("user not active")

var ErrNoPassword = errors.New("no password set")

var ErrInvalidPassword = errors.New("invalid password")

var ErrNoOnetimePassword = errors.New("no onetime password set")

var ErrTokenGeneration = errors.New("token generation failed")

var ErrDatabase = errors.New("database error")
