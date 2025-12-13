package user

import (
	"context"
	"errors"
	"strings"

	pwd "github.com/nicograef/jotti/backend/auth/password"
	"github.com/nicograef/jotti/backend/db"
	"github.com/rs/zerolog"
)

type commandPersistence interface {
	CreateUser(ctx context.Context, name, username, onetimePasswordHash string, role Role) (int, error)
	UpdateUser(ctx context.Context, id int, name, username string, role Role) error
	ActivateUser(ctx context.Context, id int) error
	DeactivateUser(ctx context.Context, id int) error
	SetOnetimePasswordHash(ctx context.Context, id int, onetimePasswordHash string) error
}

// Command provides user-related operations.
type Command struct {
	Persistence commandPersistence
}

// CreateUser creates a new user in the database without setting a password.
func (s *Command) CreateUser(ctx context.Context, name, username string, role Role) (int, string, error) {
	log := zerolog.Ctx(ctx)

	onetimePassword, err := pwd.GenerateOnetimePassword()
	if err != nil {
		log.Error().Err(err).Msg("Failed to create one-time password")
		return 0, "", ErrPasswordHashing
	}

	onetimePasswordHash, err := pwd.CreateArgon2idHash(onetimePassword)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash one-time password")
		return 0, "", ErrPasswordHashing
	}

	lowerCaseUsername := strings.ToLower(username) // usernames are always lowercase in jotti
	id, err := s.Persistence.CreateUser(ctx, name, lowerCaseUsername, onetimePasswordHash, role)
	if err != nil {
		log.Error().Str("username", lowerCaseUsername).Msg("Failed to create user")
		return 0, "", ErrDatabase
	}

	log.Info().Str("username", lowerCaseUsername).Msg("User created successfully")
	return id, onetimePassword, nil
}

// ResetPassword resets the password for the user with the given user ID and returns a new one-time password.
func (s *Command) ResetPassword(ctx context.Context, userID int) (string, error) {
	log := zerolog.Ctx(ctx)

	onetimePassword, err := pwd.GenerateOnetimePassword()
	if err != nil {
		log.Error().Err(err).Msg("Failed to create one-time password")
		return "", ErrPasswordHashing
	}

	onetimePasswordHash, err := pwd.CreateArgon2idHash(onetimePassword)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash one-time password")
		return "", ErrPasswordHashing
	}

	err = s.Persistence.SetOnetimePasswordHash(ctx, userID, onetimePasswordHash)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("user_id", userID).Msg("User not found for password reset")
			return "", ErrUserNotFound
		} else {
			log.Error().Int("user_id", userID).Msg("Failed to set one-time password hash in persistence")
			return "", ErrDatabase
		}
	}

	log.Info().Int("user_id", userID).Msg("Password reset successfully")
	return onetimePassword, nil
}

// UpdateUser updates the user's details in the database.
func (s *Command) UpdateUser(ctx context.Context, id int, name, username string, role Role) error {
	log := zerolog.Ctx(ctx)

	err := s.Persistence.UpdateUser(ctx, id, name, username, role)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("user_id", id).Msg("User not found for update")
			return ErrUserNotFound
		} else {
			log.Error().Err(err).Int("user_id", id).Msg("Failed to update user")
			return ErrDatabase
		}
	}

	log.Info().Int("user_id", id).Msg("User updated successfully")
	return nil
}

// ActivateUser sets the status of the user with the given user ID to 'active'.
func (s *Command) ActivateUser(ctx context.Context, id int) error {
	log := zerolog.Ctx(ctx)

	err := s.Persistence.ActivateUser(ctx, id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("user_id", id).Msg("User not found for activation")
			return ErrUserNotFound
		} else {
			log.Error().Int("user_id", id).Msg("Failed to activate user")
			return ErrDatabase
		}
	}

	log.Info().Int("user_id", id).Msg("User activated successfully")
	return nil
}

// DeactivateUser sets the status of the user with the given user ID to 'inactive'.
func (s *Command) DeactivateUser(ctx context.Context, id int) error {
	log := zerolog.Ctx(ctx)

	err := s.Persistence.DeactivateUser(ctx, id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("user_id", id).Msg("User not found for deactivation")
			return ErrUserNotFound
		} else {
			log.Error().Int("user_id", id).Msg("Failed to deactivate user")
			return ErrDatabase
		}
	}

	log.Info().Int("user_id", id).Msg("User deactivated successfully")
	return nil
}
