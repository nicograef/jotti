package application

import (
	"errors"

	"github.com/nicograef/jotti/backend/db"
)

var ErrUserNotFound = errors.New("user not found")

var ErrUsernameAlreadyExists = errors.New("username already exists")

var ErrInvalidUserData = errors.New("invalid user data")

var ErrDatabase = db.ErrDatabase
