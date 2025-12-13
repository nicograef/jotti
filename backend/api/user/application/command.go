package application

import (
	"context"
	"errors"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/user"
	"github.com/rs/zerolog"
)

type commandUserRepo interface {
	GetUser(ctx context.Context, userID int) (user.User, error)
	CreateUser(ctx context.Context, u user.User) (int, error)
	UpdateUser(ctx context.Context, u user.User) error
}

type Command struct {
	UserRepo commandUserRepo
}

func (c Command) CreateUser(ctx context.Context, name, username string, role user.Role) (int, string, error) {
	log := zerolog.Ctx(ctx)

	user, onetimePassword, err := user.NewUser(name, username, role)
	if err != nil {
		log.Warn().Err(err).Str("username", username).Msg("Invalid user data")
		return 0, "", ErrInvalidUserData
	}

	userID, err := c.UserRepo.CreateUser(ctx, user)
	if err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			log.Warn().Err(err).Str("username", user.Username).Msg("Username already exists")
			return 0, "", ErrUsernameAlreadyExists
		} else {
			log.Error().Str("username", user.Username).Msg("Failed to create user")
			return 0, "", ErrDatabase
		}
	}

	log.Info().Str("username", user.Username).Msg("User created successfully")
	return userID, onetimePassword, nil
}

func (c Command) UpdateUser(ctx context.Context, userID int, name, username string, role user.Role) error {
	log := zerolog.Ctx(ctx)

	user, err := c.UserRepo.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("user_id", userID).Msg("User not found for update")
			return ErrUserNotFound
		} else {
			log.Error().Int("user_id", userID).Msg("Failed to retrieve user for update")
			return ErrDatabase
		}
	}

	err = user.UpdateDetails(name, username, role)
	if err != nil {
		log.Warn().Err(err).Int("user_id", userID).Msg("Invalid user data for update")
		return ErrInvalidUserData
	}

	err = c.UserRepo.UpdateUser(ctx, user)
	if err != nil {
		log.Error().Err(err).Int("user_id", userID).Msg("Failed to update user")
		return ErrDatabase

	}

	log.Info().Int("user_id", userID).Msg("User updated successfully")
	return nil
}

func (c Command) ActivateUser(ctx context.Context, userID int) error {
	log := zerolog.Ctx(ctx)

	user, err := c.UserRepo.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("user_id", userID).Msg("User not found for activation")
			return ErrUserNotFound
		} else {
			log.Error().Int("user_id", userID).Msg("Failed to retrieve user for activation")
			return ErrDatabase
		}
	}

	user.Activate()

	err = c.UserRepo.UpdateUser(ctx, user)
	if err != nil {
		log.Error().Err(err).Int("user_id", userID).Msg("Failed to update user")
		return ErrDatabase
	}

	log.Info().Int("user_id", userID).Msg("User activated successfully")
	return nil
}

func (c Command) DeactivateUser(ctx context.Context, userID int) error {
	log := zerolog.Ctx(ctx)

	user, err := c.UserRepo.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("user_id", userID).Msg("User not found for deactivation")
			return ErrUserNotFound
		} else {
			log.Error().Int("user_id", userID).Msg("Failed to retrieve user for deactivation")
			return ErrDatabase
		}
	}

	user.Deactivate()

	err = c.UserRepo.UpdateUser(ctx, user)
	if err != nil {
		log.Error().Err(err).Int("user_id", userID).Msg("Failed to update user")
		return ErrDatabase
	}

	log.Info().Int("user_id", userID).Msg("User deactivated successfully")
	return nil
}

func (c Command) ResetPassword(ctx context.Context, userID int) (string, error) {
	log := zerolog.Ctx(ctx)

	user, err := c.UserRepo.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("user_id", userID).Msg("User not found for password reset")
			return "", ErrUserNotFound
		} else {
			log.Error().Int("user_id", userID).Msg("Failed to retrieve user for password reset")
			return "", ErrDatabase
		}
	}

	onetimePassword, err := user.ResetPassword()
	if err != nil {
		log.Error().Err(err).Int("user_id", userID).Msg("Failed to reset password")
		return "", err
	}

	err = c.UserRepo.UpdateUser(ctx, user)
	if err != nil {
		log.Error().Int("user_id", userID).Msg("Failed to update user in persistence")
		return "", ErrDatabase
	}

	log.Info().Int("user_id", userID).Msg("Password reset successfully")
	return onetimePassword, nil
}
