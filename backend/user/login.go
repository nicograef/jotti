package user

import (
	"context"
	"errors"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api"
	"github.com/rs/zerolog"
)

type authCommand interface {
	VerifyPasswordAndGetUser(ctx context.Context, username, password string) (*User, error)
	SetNewPassword(ctx context.Context, username, password, onetimePassword string) (*User, error)
}

type AuthHandler struct {
	Command   authCommand
	JWTSecret string
}

type login struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var loginSchema = z.Struct(z.Shape{
	"Username": UsernameSchema.Required(),
	"Password": PasswordSchema.Required(),
})

type loginResponse struct {
	Token string `json:"token"`
}

// LoginHandler handles user login requests by validating the password hash against the database and returns a jwt token if successful.
func (h *AuthHandler) LoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := zerolog.Ctx(ctx)

		body := login{}
		if !api.ReadAndValidateBody(w, r, &body, loginSchema) {
			return
		}

		user, err := h.Command.VerifyPasswordAndGetUser(ctx, body.Username, body.Password)
		if err != nil {
			if errors.Is(err, ErrUserNotFound) || errors.Is(err, ErrInvalidPassword) {
				api.SendClientError(w, "invalid_credentials", nil)
				return
			} else if errors.Is(err, ErrNoPassword) {
				api.SendClientError(w, "no_password_set", "No password set for user. Please set a password first.")
				return
			} else {
				api.SendServerError(w)
				return
			}
		}

		if user.Status != ActiveStatus {
			log.Warn().Str("username", body.Username).Msg("Inactive user attempted to log in")
			api.SendClientError(w, "user_inactive", nil)
			return
		}

		stringToken, err := generateJWTTokenForUser(*user, h.JWTSecret)
		if err != nil {
			log.Error().Err(err).Str("username", body.Username).Msg("Failed to generate JWT token")
			api.SendServerError(w)
			return
		}

		api.SendResponse(w, loginResponse{Token: stringToken})
	}
}

type setPassword struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	OnetimePassword string `json:"onetimePassword"`
}

var setPasswordSchema = z.Struct(z.Shape{
	"Username":        UsernameSchema.Required(),
	"Password":        PasswordSchema.Required(),
	"OnetimePassword": OnetimePasswordSchema.Required(),
})

type setPasswordResponse struct {
	Token string `json:"token"`
}

// SetPasswordHandler handles setting a new password for a user using a one-time password and returns a jwt token if successful.
func (h *AuthHandler) SetPasswordHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := zerolog.Ctx(ctx)

		body := setPassword{}
		if !api.ReadAndValidateBody(w, r, &body, setPasswordSchema) {
			return
		}

		user, err := h.Command.SetNewPassword(ctx, body.Username, body.Password, body.OnetimePassword)
		if err != nil {
			if errors.Is(err, ErrUserNotFound) {
				api.SendClientError(w, "invalid_credentials", nil)
				return
			} else if errors.Is(err, ErrInvalidPassword) {
				api.SendClientError(w, "invalid_credentials", nil)
				return
			} else if errors.Is(err, ErrNoOnetimePassword) {
				api.SendClientError(w, "already_has_password", "No one-time password set for user. User probably already has a password.")
				return
			} else {
				api.SendServerError(w)
				return
			}
		}

		stringToken, err := generateJWTTokenForUser(*user, h.JWTSecret)
		if err != nil {
			log.Error().Err(err).Str("username", body.Username).Msg("Failed to generate JWT token")
			api.SendServerError(w)
			return
		}

		api.SendResponse(w, setPasswordResponse{Token: stringToken})
	}
}
