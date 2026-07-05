package application

import (
	"context"
	"errors"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/jwt"
	"github.com/nicograef/jotti/backend/domain/user"
	"github.com/rs/zerolog"
)

type commandUserRepo interface {
	GetUserByUsername(ctx context.Context, username string) (user.User, error)
	// SetPasswordTx lädt den Benutzer mit Zeilensperre, führt apply aus und
	// persistiert das Ergebnis in EINER Transaktion (siehe Repository).
	SetPasswordTx(ctx context.Context, username string, apply func(*user.User) error) error
}

// loginThrottle drosselt Fehlanmeldungen pro Konto (In-Memory-Infrastruktur, kein
// Domain-State). Injiziert wie das Repository; die konkrete Implementierung ist
// throttle.LoginThrottle.
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
	log := zerolog.Ctx(ctx)

	// Kontobezogener Soft-Throttle VOR der Passwortprüfung: zu viele Fehlversuche
	// drosseln den nächsten Versuch kurz (läuft von selbst ab).
	if !c.Throttle.Allow(username) {
		log.Warn().Str("username", username).Msg("Login throttled after repeated failures")
		return "", ErrLoginThrottled
	}

	u, err := c.UserRepo.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			c.Throttle.RecordFailure(username)
			log.Warn().Str("username", username).Msg("User not found during login")
			return "", ErrUserNotFound
		}
		log.Error().Str("username", username).Msg("Failed to retrieve user ID")
		return "", ErrDatabase
	}

	if err := u.VerifyPassword(password); err != nil {
		c.Throttle.RecordFailure(username)
		switch {
		case errors.Is(err, user.ErrNotActive):
			log.Warn().Str("username", username).Msg("Inactive user attempted to log in")
			return "", ErrNotActive
		case errors.Is(err, user.ErrNoPassword):
			log.Warn().Str("username", username).Msg("No password set for user during login")
			return "", ErrNoPassword
		case errors.Is(err, user.ErrInvalidPassword):
			log.Warn().Err(err).Str("username", username).Msg("Password validation failed")
			return "", ErrInvalidPassword
		default:
			log.Error().Err(err).Str("username", username).Msg("Failed to verify password")
			return "", ErrTokenGeneration
		}
	}

	token, err := jwt.GenerateJWTTokenForUser(u.ID, u.Username, string(u.Role), c.JWTSecret)
	if err != nil {
		log.Error().Err(err).Str("username", username).Msg("Failed to generate JWT token")
		return "", ErrTokenGeneration
	}

	c.Throttle.Reset(username)
	log.Info().Str("username", username).Msg("User logged in successfully")
	return token, nil
}

func (c Command) SetNewPassword(ctx context.Context, username, newPassword, onetimePassword string) error {
	log := zerolog.Ctx(ctx)

	// Read-modify-write in EINER Transaktion mit Zeilensperre: SetPasswordTx sperrt
	// die User-Zeile (FOR UPDATE), führt SetPassword aus (das den Fehlversuchszähler
	// hochzählt) und persistiert das Ergebnis im selben Commit. So können sich
	// nebenläufige Set-Password-Versuche nicht überholen und der Zähler unterzählen.
	err := c.UserRepo.SetPasswordTx(ctx, username, func(u *user.User) error {
		return u.SetPassword(onetimePassword, newPassword)
	})
	if err != nil {
		switch {
		case errors.Is(err, db.ErrNotFound):
			log.Warn().Str("username", username).Msg("User not found during password reset")
			return ErrUserNotFound
		case errors.Is(err, user.ErrNoPassword):
			log.Warn().Str("username", username).Msg("No one-time password set for user during password reset")
			return ErrNoOnetimePassword
		case errors.Is(err, user.ErrPasswordTooWeak):
			log.Warn().Str("username", username).Msg("Password too weak")
			return ErrPasswordTooWeak
		case errors.Is(err, user.ErrOnetimePasswordLocked):
			log.Warn().Str("username", username).Msg("One-time password locked after too many failed attempts")
			return ErrOnetimePasswordLocked
		case errors.Is(err, user.ErrInvalidPassword):
			log.Warn().Err(err).Str("username", username).Msg("One-time password validation failed")
			return ErrInvalidPassword
		default:
			log.Error().Err(err).Str("username", username).Msg("Failed to set new password")
			return ErrDatabase
		}
	}

	log.Info().Str("username", username).Msg("New password set successfully")
	return nil
}
