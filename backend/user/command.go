package user

import (
	"context"
	"errors"
	"strings"

	"github.com/rs/zerolog"
)

type commandPersistence interface {
	GetUserID(ctx context.Context, username string) (int, error)
	GetUser(ctx context.Context, id int) (*User, error)
	CreateUser(ctx context.Context, name, username, onetimePasswordHash string, role Role) (int, error)
	UpdateUser(ctx context.Context, id int, name, username string, role Role) error
	ActivateUser(ctx context.Context, id int) error
	DeactivateUser(ctx context.Context, id int) error
	SetPasswordHash(ctx context.Context, id int, passwordHash string) error
	SetOnetimePasswordHash(ctx context.Context, id int, onetimePasswordHash string) error
}

// Command provides user-related operations.
type Command struct {
	Persistence commandPersistence
}

// CreateUser creates a new user in the database without setting a password.
func (s *Command) CreateUser(ctx context.Context, name, username string, role Role) (int, string, error) {
	log := zerolog.Ctx(ctx)

	onetimePassword, err := generateOnetimePassword()
	if err != nil {
		log.Error().Err(err).Msg("Failed to create one-time password")
		return 0, "", ErrPasswordHashing
	}

	onetimePasswordHash, err := createArgon2idHash(onetimePassword)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash one-time password")
		return 0, "", ErrPasswordHashing
	}

	lowerCaseUsername := strings.ToLower(username) // usernames are always lowercase in jotti
	id, err := s.Persistence.CreateUser(ctx, name, lowerCaseUsername, onetimePasswordHash, role)
	if err != nil {
		log.Error().Err(err).Str("username", lowerCaseUsername).Msg("Failed to create user")
		return 0, "", ErrDatabase
	}

	return id, onetimePassword, nil
}

// VerifyPasswordAndGetUser logs in a user by validating the provided password against the stored password hash.
// If the user has no password set, it sets the provided password as the new password.
func (s *Command) VerifyPasswordAndGetUser(ctx context.Context, username, password string) (*User, error) {
	log := zerolog.Ctx(ctx)

	id, err := s.Persistence.GetUserID(ctx, username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			log.Warn().Str("username", username).Msg("User not found during login")
			return nil, ErrUserNotFound
		}
		log.Error().Err(err).Str("username", username).Msg("Failed to retrieve user")
		return nil, ErrDatabase
	}

	user, err := s.Persistence.GetUser(ctx, id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			log.Warn().Str("username", username).Msg("User not found during login")
			return nil, ErrUserNotFound
		}
		log.Error().Err(err).Str("username", username).Msg("Failed to retrieve password hash")
		return nil, ErrDatabase
	}
	if user.PasswordHash == "" {
		log.Warn().Str("username", username).Msg("No password set for user")
		return nil, ErrNoPassword
	}

	if err := verifyPassword(user.PasswordHash, password); err != nil {
		log.Warn().Err(err).Str("username", username).Msg("Password validation failed")
		return nil, ErrInvalidPassword
	}

	return user, nil
}

// SetNewPassword logs in a user by validating the provided one-time password against the stored password hash.
// If the user has no password set, it sets the provided password as the new password.
func (s *Command) SetNewPassword(ctx context.Context, username, newPassword, onetimePassword string) (*User, error) {
	log := zerolog.Ctx(ctx)

	id, err := s.Persistence.GetUserID(ctx, username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			log.Warn().Str("username", username).Msg("User not found during password validation")
			return nil, ErrUserNotFound
		}
		log.Error().Err(err).Str("username", username).Msg("Failed to retrieve user")
		return nil, ErrDatabase
	}

	user, err := s.Persistence.GetUser(ctx, id)
	if err != nil {
		log.Error().Err(err).Str("username", username).Msg("Failed to retrieve one-time password hash")
		return nil, ErrDatabase
	}
	if user.OnetimePasswordHash == "" {
		log.Warn().Str("username", username).Msg("No one-time password set for user")
		return nil, ErrNoOnetimePassword
	}
	if err := verifyPassword(user.OnetimePasswordHash, onetimePassword); err != nil {
		log.Warn().Err(err).Str("username", username).Msg("One-time password validation failed")
		return nil, ErrInvalidPassword
	}

	hashedPassword, err := createArgon2idHash(newPassword)
	if err != nil {
		log.Error().Err(err).Str("username", username).Msg("Failed to hash password")
		return nil, ErrPasswordHashing
	}

	if err := s.Persistence.SetPasswordHash(ctx, user.ID, hashedPassword); err != nil {
		log.Error().Err(err).Str("username", username).Msg("Failed to set password hash in persistence")
		return nil, ErrDatabase
	}

	log.Info().Str("username", username).Msg("Password set successfully")

	return user, nil
}

// ResetPassword resets the password for the user with the given user ID and returns a new one-time password.
func (s *Command) ResetPassword(ctx context.Context, userID int) (string, error) {
	log := zerolog.Ctx(ctx)

	onetimePassword, err := generateOnetimePassword()
	if err != nil {
		log.Error().Err(err).Msg("Failed to create one-time password")
		return "", ErrPasswordHashing
	}

	onetimePasswordHash, err := createArgon2idHash(onetimePassword)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash one-time password")
		return "", ErrPasswordHashing
	}

	err = s.Persistence.SetOnetimePasswordHash(ctx, userID, onetimePasswordHash)
	if err != nil && errors.Is(err, ErrUserNotFound) {
		log.Warn().Int("user_id", userID).Msg("User not found for password reset")
		return "", ErrUserNotFound
	} else if err != nil {
		log.Error().Err(err).Int("user_id", userID).Msg("Failed to set one-time password hash in persistence")
		return "", ErrDatabase
	}

	return onetimePassword, nil
}

// UpdateUser updates the user's details in the database.
func (s *Command) UpdateUser(ctx context.Context, id int, name, username string, role Role) error {
	log := zerolog.Ctx(ctx)

	err := s.Persistence.UpdateUser(ctx, id, name, username, role)
	if err != nil && errors.Is(err, ErrUserNotFound) {
		log.Warn().Int("user_id", id).Msg("User not found for update")
		return ErrUserNotFound
	} else if err != nil {
		log.Error().Err(err).Int("user_id", id).Msg("Failed to update user")
		return ErrDatabase
	}

	return nil
}

// ActivateUser sets the status of the user with the given user ID to 'active'.
func (s *Command) ActivateUser(ctx context.Context, id int) error {
	log := zerolog.Ctx(ctx)

	err := s.Persistence.ActivateUser(ctx, id)
	if err != nil && errors.Is(err, ErrUserNotFound) {
		log.Warn().Int("user_id", id).Msg("User not found for activation")
		return ErrUserNotFound
	} else if err != nil {
		log.Error().Err(err).Int("user_id", id).Msg("Failed to activate user")
		return ErrDatabase
	}

	return nil
}

// DeactivateUser sets the status of the user with the given user ID to 'inactive'.
func (s *Command) DeactivateUser(ctx context.Context, id int) error {
	log := zerolog.Ctx(ctx)

	err := s.Persistence.DeactivateUser(ctx, id)
	if err != nil && errors.Is(err, ErrUserNotFound) {
		log.Warn().Int("user_id", id).Msg("User not found for deactivation")
		return ErrUserNotFound
	} else if err != nil {
		log.Error().Err(err).Int("user_id", id).Msg("Failed to deactivate user")
		return ErrDatabase
	}

	return nil
}
