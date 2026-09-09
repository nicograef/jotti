package application

import (
	"errors"

	"github.com/nicograef/jotti/backend/db"
)

// ErrUserNotFound is returned when a user is not found.
var ErrUserNotFound = errors.New("user not found")

// ErrUsernameAlreadyExists is returned when trying to create a user with a username that already exists.
var ErrUsernameAlreadyExists = errors.New("username already exists")

var ErrInvalidUserData = errors.New("invalid user data")

// ErrDatabase is returned when there is a database error.
var ErrDatabase = db.ErrDatabase
