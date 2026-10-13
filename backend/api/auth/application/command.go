package application

import (
	"context"
	"errors"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/jwt"
	"github.com/nicograef/jotti/backend/domain/user"
	"github.com/nicograef/jotti/backend/internal/logging"
)

type commandUserRepo interface {
	GetUserByUsername(ctx context.Context, username string) (user.User, error)
	// SetPasswordTx sperrt die User-Zeile, führt apply aus und persistiert in EINER Transaktion.
	SetPasswordTx(ctx context.Context, username string, apply func(*user.User) error) error
}

type loginThrottle interface {
	Allow(username string) bool
	RecordFailure(username string)
	Reset(username string)
}

type Command struct {
	JWTSecret string
	UserRepo  commandUserRepo
	Throttle  loginThrottle
}

func (c Command) GenerateJWTToken(ctx context.Context, username, password string) (string, error) {
	log := logging.FromContext(ctx)

	if !c.Throttle.Allow(username) {
		log.Warn("Login throttled after repeated failures", "username", username)
		return "", ErrLoginThrottled
	}

	u, err := c.UserRepo.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			c.Throttle.RecordFailure(username)
			log.Warn("User not found during login", "username", username)
			return "", ErrUserNotFound
		}
		log.Error("Failed to retrieve user ID", "error", err, "username", username)
		return "", ErrDatabase
	}

	if err := u.VerifyPassword(password); err != nil {
		c.Throttle.RecordFailure(username)
		switch {
		case errors.Is(err, user.ErrNotActive):
			log.Warn("Inactive user attempted to log in", "username", username)
			return "", ErrNotActive
		case errors.Is(err, user.ErrNoPassword):
			log.Warn("No password set for user during login", "username", username)
			return "", ErrNoPassword
		case errors.Is(err, user.ErrInvalidPassword):
			log.Warn("Password validation failed", "error", err, "username", username)
			return "", ErrInvalidPassword
		default:
			log.Error("Failed to verify password", "error", err, "username", username)
			return "", ErrTokenGeneration
		}
	}

	token, err := jwt.GenerateJWTTokenForUser(u.ID, string(u.Role), c.JWTSecret)
	if err != nil {
		log.Error("Failed to generate JWT token", "error", err, "username", username)
		return "", ErrTokenGeneration
	}

	c.Throttle.Reset(username)
	log.Info("User logged in successfully", "username", username)
	return token, nil
}

func (c Command) SetNewPassword(ctx context.Context, username, newPassword, onetimePassword string) error {
	log := logging.FromContext(ctx)

	// SetPasswordTx sperrt die User-Zeile (FOR UPDATE) und persistiert das Ergebnis
	// im selben Commit; sonst könnten nebenläufige Versuche den Fehlversuchszähler
	// unterzählen.
	err := c.UserRepo.SetPasswordTx(ctx, username, func(u *user.User) error {
		return u.SetPassword(onetimePassword, newPassword)
	})
	if err != nil {
		switch {
		case errors.Is(err, db.ErrNotFound):
			log.Warn("User not found during password reset", "username", username)
			return ErrUserNotFound
		case errors.Is(err, user.ErrNoPassword):
			log.Warn("No one-time password set for user during password reset", "username", username)
			return ErrNoOnetimePassword
		case errors.Is(err, user.ErrPasswordTooWeak):
			log.Warn("Password too weak", "username", username)
			return ErrPasswordTooWeak
		case errors.Is(err, user.ErrOnetimePasswordLocked):
			log.Warn("One-time password locked after too many failed attempts", "username", username)
			return ErrOnetimePasswordLocked
		case errors.Is(err, user.ErrInvalidPassword):
			log.Warn("One-time password validation failed", "error", err, "username", username)
			return ErrInvalidPassword
		default:
			log.Error("Failed to set new password", "error", err, "username", username)
			return ErrDatabase
		}
	}

	log.Info("New password set successfully", "username", username)
	return nil
}
