package application

import (
	"errors"
	"log/slog"

	"github.com/nicograef/jotti/backend/db"
)

var ErrUserNotFound = errors.New("user not found")

var ErrUsernameAlreadyExists = errors.New("username already exists")

var ErrInvalidUserData = errors.New("invalid user data")

// The self-protection errors keep an admin from locking the instance out of its own account.
var (
	ErrCannotDemoteSelf     = errors.New("cannot demote own account")
	ErrCannotDeactivateSelf = errors.New("cannot deactivate own account")
	ErrCannotDeleteSelf     = errors.New("cannot delete own account")
)

var ErrDatabase = db.ErrDatabase

// fromRepositoryError maps a repository error on a user; vorgang and logArgs identify the call site.
func fromRepositoryError(err error, log *slog.Logger, vorgang string, logArgs ...any) error {
	logArgs = append([]any{"vorgang", vorgang}, logArgs...)
	switch {
	case errors.Is(err, db.ErrNotFound):
		log.Warn("User not found", logArgs...)
		return ErrUserNotFound
	case errors.Is(err, db.ErrAlreadyExists):
		log.Warn("Username already exists", append(logArgs, "error", err)...)
		return ErrUsernameAlreadyExists
	default:
		log.Error("User database error", append(logArgs, "error", err)...)
		return ErrDatabase
	}
}
