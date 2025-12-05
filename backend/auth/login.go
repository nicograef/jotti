package auth

import (
	"context"
	"errors"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api"
	usr "github.com/nicograef/jotti/backend/user"
	"github.com/rs/zerolog"
)

type userService interface {
	VerifyPasswordAndGetUser(ctx context.Context, username, password string) (*usr.User, error)
	SetNewPassword(ctx context.Context, username, password, onetimePassword string) (*usr.User, error)
}

type Handler struct {
	UserService userService
	JWTSecret   string
}

type login struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var loginSchema = z.Struct(z.Shape{
	"Username": usr.UsernameSchema.Required(),
	"Password": usr.PasswordSchema.Required(),
})

type loginResponse struct {
	Token string `json:"token"`
}

// LoginHandler handles user login requests by validating the password hash against the database and returns a jwt token if successful.
func (h *Handler) LoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := zerolog.Ctx(ctx)

		body := login{}
		if !api.ReadAndValidateBody(w, r, &body, loginSchema) {
			return
		}

		user, err := h.UserService.VerifyPasswordAndGetUser(ctx, body.Username, body.Password)
		if err != nil {
			if errors.Is(err, usr.ErrUserNotFound) || errors.Is(err, usr.ErrInvalidPassword) {
				api.SendClientError(w, "invalid_credentials", nil)
				return
			} else {
				api.SendServerError(w)
				return
			}
		}

		if user.Status != usr.ActiveStatus {
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
	"Username":        usr.UsernameSchema.Required(),
	"Password":        usr.PasswordSchema.Required(),
	"OnetimePassword": usr.OnetimePasswordSchema.Required(),
})

type setPasswordResponse struct {
	Token string `json:"token"`
}

// SetPasswordHandler handles setting a new password for a user using a one-time password and returns a jwt token if successful.
func (h *Handler) SetPasswordHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := zerolog.Ctx(ctx)

		body := setPassword{}
		if !api.ReadAndValidateBody(w, r, &body, setPasswordSchema) {
			return
		}

		user, err := h.UserService.SetNewPassword(ctx, body.Username, body.Password, body.OnetimePassword)
		if err != nil {
			if errors.Is(err, usr.ErrUserNotFound) {
				api.SendClientError(w, "invalid_credentials", nil)
				return
			} else if errors.Is(err, usr.ErrInvalidPassword) {
				api.SendClientError(w, "invalid_credentials", nil)
				return
			} else if errors.Is(err, usr.ErrNoOnetimePassword) {
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
