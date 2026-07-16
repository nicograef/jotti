package http

import (
	"context"
	"errors"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api/auth/application"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/domain/user"
)

type authCommand interface {
	GenerateJWTToken(ctx context.Context, username, password string) (string, error)
	SetNewPassword(ctx context.Context, username, password, onetimePassword string) error
}

type CommandHandler struct {
	Command authCommand
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var loginSchema = z.Struct(z.Shape{
	"Username": z.String().Trim().Min(1, z.Message("Benutzername ist erforderlich")).Required(),
	"Password": z.String().Min(1, z.Message("Passwort ist erforderlich")).Required(),
})

type loginResponse struct {
	Token string `json:"token"`
}

func (h *CommandHandler) LoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		body := loginRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, loginSchema) {
			return
		}

		token, err := h.Command.GenerateJWTToken(ctx, body.Username, body.Password)
		if err != nil {
			switch {
			case errors.Is(err, application.ErrLoginThrottled):
				helper.SendTooManyRequests(w, "login_throttled")
			case errors.Is(err, application.ErrNotActive):
				helper.SendClientError(w, "user_inactive", nil)
			case errors.Is(err, application.ErrUserNotFound), errors.Is(err, application.ErrInvalidPassword):
				helper.SendClientError(w, "invalid_credentials", nil)
			case errors.Is(err, application.ErrNoPassword):
				helper.SendClientError(w, "no_password_set", "user has no password yet, set one via one-time password first")
			default:
				helper.SendServerError(w)
			}
			return
		}

		helper.SendResponse(w, loginResponse{Token: token})
	}
}

type setPasswordRequest struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	OnetimePassword string `json:"onetimePassword"`
}

var setPasswordSchema = z.Struct(z.Shape{
	"Username":        user.UsernameSchema.Required(),
	"Password":        user.PasswordSchema.Required(),
	"OnetimePassword": user.OnetimePasswordSchema.Required(),
})

func (h *CommandHandler) SetPasswordHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		body := setPasswordRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, setPasswordSchema) {
			return
		}

		err := h.Command.SetNewPassword(ctx, body.Username, body.Password, body.OnetimePassword)
		if err != nil {
			switch {
			case errors.Is(err, application.ErrUserNotFound), errors.Is(err, application.ErrInvalidPassword):
				helper.SendClientError(w, "invalid_credentials", nil)
			case errors.Is(err, application.ErrNoOnetimePassword):
				helper.SendClientError(w, "already_has_password", "no one-time password set, user already has a password")
			case errors.Is(err, application.ErrPasswordTooWeak):
				helper.SendClientError(w, "password_too_weak", nil)
			case errors.Is(err, application.ErrOnetimePasswordLocked):
				helper.SendClientError(w, "onetime_password_locked", "one-time password locked after too many failed attempts, an admin must generate a new one")
			default:
				helper.SendServerError(w)
			}
			return
		}

		helper.SendEmptyResponse(w)
	}
}
