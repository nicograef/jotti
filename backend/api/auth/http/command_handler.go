package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/nicograef/jotti/backend/api/auth/application"
	"github.com/nicograef/jotti/backend/api/helper"
)

type authCommand interface {
	GenerateJWTToken(ctx context.Context, username, password string) (string, error)
	SetNewPassword(ctx context.Context, username, password, onetimePassword string) error
}

type CommandHandler struct {
	Command authCommand
}

type login struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

func (h *CommandHandler) LoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		body := login{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		token, err := h.Command.GenerateJWTToken(ctx, body.Username, body.Password)
		if err != nil {
			if errors.Is(err, application.ErrNotActive) {
				helper.SendClientError(w, "user_inactive", nil)
				return
			} else if errors.Is(err, application.ErrUserNotFound) || errors.Is(err, application.ErrInvalidPassword) {
				helper.SendClientError(w, "invalid_credentials", nil)
				return
			} else if errors.Is(err, application.ErrNoPassword) {
				helper.SendClientError(w, "no_password_set", "No password set for user. Please set a password first.")
				return
			} else {
				helper.SendServerError(w)
				return
			}
		}

		helper.SendResponse(w, loginResponse{Token: token})
	}
}

type setPassword struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	OnetimePassword string `json:"onetimePassword"`
}

func (h *CommandHandler) SetPasswordHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		body := setPassword{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		err := h.Command.SetNewPassword(ctx, body.Username, body.Password, body.OnetimePassword)
		if err != nil {
			if errors.Is(err, application.ErrUserNotFound) {
				helper.SendClientError(w, "invalid_credentials", nil)
				return
			} else if errors.Is(err, application.ErrInvalidPassword) {
				helper.SendClientError(w, "invalid_credentials", nil)
				return
			} else if errors.Is(err, application.ErrNoOnetimePassword) {
				helper.SendClientError(w, "already_has_password", "No one-time password set for user. User probably already has a password.")
				return
			} else {
				helper.SendServerError(w)
				return
			}
		}

		helper.SendEmptyResponse(w)
	}
}
