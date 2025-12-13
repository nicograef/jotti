package application

import (
	"errors"
)

// ErrUserNotFound is returned when a user is not found.
var ErrUserNotFound = errors.New("user not found")

// ErrUserNotFound is returned when a user is not found.
var ErrUsernameAlreadyExists = errors.New("username already exists")

// ErrNoPassword is returned when there is no password set for the user.
var ErrNoPassword = errors.New("no password set")

var ErrInvalidUserData = errors.New("invalid user data")

// ErrNoOnetimePassword is returned when there is no one-time password set for the user.
var ErrNoOnetimePassword = errors.New("no onetime password set")

// ErrDatabase is returned when there is a database error.
var ErrDatabase = errors.New("database error")
