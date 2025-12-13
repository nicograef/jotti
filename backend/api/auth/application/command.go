package application

import (
	"context"
	"errors"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/user"
	"github.com/rs/zerolog"
)

type commandUserRepo interface {
	GetUserByUsername(ctx context.Context, username string) (user.User, error)
	UpdateUser(ctx context.Context, u user.User) error
}

type Command struct {
	JWTSecret string
	UserRepo  commandUserRepo
}

func (c Command) GenerateJWTToken(ctx context.Context, username, password string) (string, error) {
	log := zerolog.Ctx(ctx)

	u, err := c.UserRepo.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Str("username", username).Msg("User not found during login")
			return "", ErrUserNotFound
		} else {
			log.Error().Str("username", username).Msg("Failed to retrieve user ID")
			return "", ErrDatabase
		}
	}

	token, err := u.GenerateJWTToken(password, c.JWTSecret)
	if err != nil {
		if errors.Is(err, user.ErrNotActive) {
			log.Warn().Str("username", username).Msg("Inactive user attempted to log in")
			return "", ErrNotActive
		} else if errors.Is(err, user.ErrNoPassword) {
			log.Warn().Str("username", username).Msg("No password set for user during login")
			return "", ErrNoPassword
		} else if errors.Is(err, user.ErrInvalidPassword) {
			log.Warn().Err(err).Str("username", username).Msg("Password validation failed")
			return "", ErrInvalidPassword
		} else {
			log.Error().Err(err).Str("username", username).Msg("Failed to generate JWT token")
			return "", ErrTokenGeneration
		}
	}

	log.Info().Str("username", username).Msg("User logged in successfully")
	return token, nil
}

func (c Command) SetNewPassword(ctx context.Context, username, newPassword, onetimePassword string) error {
	log := zerolog.Ctx(ctx)

	u, err := c.UserRepo.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Str("username", username).Msg("User not found during login")
			return ErrUserNotFound
		} else {
			log.Error().Str("username", username).Msg("Failed to retrieve user ID")
			return ErrDatabase
		}
	}

	err = u.SetPassword(onetimePassword, newPassword)
	if err != nil {
		if errors.Is(err, user.ErrNoPassword) {
			log.Warn().Str("username", username).Msg("No one-time password set for user during password reset")
			return ErrNoOnetimePassword
		} else {
			log.Warn().Err(err).Str("username", username).Msg("One-time password validation failed")
			return ErrInvalidPassword
		}
	}

	err = c.UserRepo.UpdateUser(ctx, u)
	if err != nil {
		log.Error().Str("username", username).Msg("Failed to set password hash in persistence")
		return ErrDatabase
	}

	log.Info().Str("username", username).Msg("New password set successfully")
	return nil
}
