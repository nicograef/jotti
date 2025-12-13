package auth

import (
	"context"
	"errors"

	pwd "github.com/nicograef/jotti/backend/auth/password"
	"github.com/nicograef/jotti/backend/db"
	"github.com/rs/zerolog"
)

type commandPersistence interface {
	GetUserID(ctx context.Context, username string) (int, error)
	GetUser(ctx context.Context, id int) (*User, error)
	SetPasswordHash(ctx context.Context, id int, passwordHash string) error
}

// Command provides user-related operations.
type Command struct {
	Persistence commandPersistence
}

// VerifyPasswordAndGetUser logs in a user by validating the provided password against the stored password hash.
// If the user has no password set, it sets the provided password as the new password.
func (s *Command) VerifyPasswordAndGetUser(ctx context.Context, username, password string) (*User, error) {
	log := zerolog.Ctx(ctx)

	id, err := s.Persistence.GetUserID(ctx, username)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Str("username", username).Msg("User not found during login")
			return nil, ErrUserNotFound
		} else {
			log.Error().Str("username", username).Msg("Failed to retrieve user ID")
			return nil, ErrDatabase
		}
	}

	user, err := s.Persistence.GetUser(ctx, id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Str("username", username).Msg("User not found during login")
			return nil, ErrUserNotFound
		} else {
			log.Error().Str("username", username).Msg("Failed to retrieve password hash")
			return nil, ErrDatabase
		}
	}

	if user.PasswordHash == "" {
		log.Warn().Str("username", username).Msg("No password set for user")
		return nil, ErrNoPassword
	}

	if err := pwd.VerifyPassword(user.PasswordHash, password); err != nil {
		log.Warn().Err(err).Str("username", username).Msg("Password validation failed")
		return nil, ErrInvalidPassword
	}

	log.Info().Str("username", username).Msg("User logged in successfully")
	return user, nil
}

// SetNewPassword logs in a user by validating the provided one-time password against the stored password hash.
// If the user has no password set, it sets the provided password as the new password.
func (s *Command) SetNewPassword(ctx context.Context, username, newPassword, onetimePassword string) error {
	log := zerolog.Ctx(ctx)

	id, err := s.Persistence.GetUserID(ctx, username)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Str("username", username).Msg("User not found during password validation")
			return ErrUserNotFound
		} else {
			log.Error().Str("username", username).Msg("Failed to retrieve user")
			return ErrDatabase
		}
	}

	user, err := s.Persistence.GetUser(ctx, id)
	if err != nil {
		log.Error().Str("username", username).Msg("Failed to retrieve one-time password hash")
		return ErrDatabase
	}

	if user.OnetimePasswordHash == "" {
		log.Warn().Str("username", username).Msg("No one-time password set for user")
		return ErrNoOnetimePassword
	}

	if err := pwd.VerifyPassword(user.OnetimePasswordHash, onetimePassword); err != nil {
		log.Warn().Err(err).Str("username", username).Msg("One-time password validation failed")
		return ErrInvalidPassword
	}

	hashedPassword, err := pwd.CreateArgon2idHash(newPassword)
	if err != nil {
		log.Error().Err(err).Str("username", username).Msg("Failed to hash password")
		return ErrPasswordHashing
	}

	if err := s.Persistence.SetPasswordHash(ctx, user.ID, hashedPassword); err != nil {
		log.Error().Str("username", username).Msg("Failed to set password hash in persistence")
		return ErrDatabase
	}

	log.Info().Str("username", username).Msg("New password set successfully")
	return nil
}
