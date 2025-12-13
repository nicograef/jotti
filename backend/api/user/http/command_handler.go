package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/user/application"
	"github.com/nicograef/jotti/backend/domain/user"
)

type command interface {
	CreateUser(ctx context.Context, name, username string, role user.Role) (int, string, error)
	UpdateUser(ctx context.Context, id int, name, username string, role user.Role) error
	ActivateUser(ctx context.Context, id int) error
	DeactivateUser(ctx context.Context, id int) error
	ResetPassword(ctx context.Context, userID int) (string, error)
}

type CommandHandler struct {
	Command command
}

type createUser struct {
	Name     string    `json:"name"`
	Username string    `json:"username"`
	Role     user.Role `json:"role"`
}

type createUserResponse struct {
	ID              int    `json:"id"`
	OnetimePassword string `json:"onetimePassword"`
}

func (h CommandHandler) CreateUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := createUser{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		userID, onetimePassword, err := h.Command.CreateUser(r.Context(), body.Name, body.Username, body.Role)
		if err != nil {
			if errors.Is(err, application.ErrUsernameAlreadyExists) {
				helper.SendClientError(w, "username_already_exists", nil)
				return
			} else {
				helper.SendServerError(w)
				return
			}
		}

		helper.SendResponse(w, createUserResponse{ID: userID, OnetimePassword: onetimePassword})
	}
}

type updateUser struct {
	ID       int       `json:"id"`
	Name     string    `json:"name"`
	Username string    `json:"username"`
	Role     user.Role `json:"role"`
}

func (h CommandHandler) UpdateUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := updateUser{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		err := h.Command.UpdateUser(r.Context(), body.ID, body.Name, body.Username, body.Role)
		if err != nil {
			if errors.Is(err, application.ErrUserNotFound) {
				helper.SendClientError(w, "user_not_found", nil)
				return
			} else if errors.Is(err, application.ErrUsernameAlreadyExists) {
				helper.SendClientError(w, "username_already_exists", nil)
				return
			} else {
				helper.SendServerError(w)
				return
			}
		}

		helper.SendEmptyResponse(w)
	}
}

type resetPassword struct {
	ID int `json:"id"`
}

type resetPasswordResponse struct {
	OnetimePassword string `json:"onetimePassword"`
}

// ResetPasswordHandler handles requests to reset a user's password.
func (h CommandHandler) ResetPasswordHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := resetPassword{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		onetimePassword, err := h.Command.ResetPassword(r.Context(), body.ID)
		if err != nil {
			if errors.Is(err, application.ErrUserNotFound) {
				helper.SendClientError(w, "user_not_found", nil)
				return
			} else {
				helper.SendServerError(w)
				return
			}
		}

		helper.SendResponse(w, resetPasswordResponse{OnetimePassword: onetimePassword})
	}
}

type activateUser struct {
	ID int `json:"id"`
}

// ActivateUserHandler handles requests to activate a user.
func (h CommandHandler) ActivateUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := activateUser{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		err := h.Command.ActivateUser(r.Context(), body.ID)
		if err != nil {
			if errors.Is(err, application.ErrUserNotFound) {
				helper.SendClientError(w, "user_not_found", nil)
				return
			} else {
				helper.SendServerError(w)
				return
			}
		}

		helper.SendEmptyResponse(w)
	}
}

type deactivateUser struct {
	ID int `json:"id"`
}

// DeactivateUserHandler handles requests to deactivate a user.
func (h CommandHandler) DeactivateUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := deactivateUser{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		err := h.Command.DeactivateUser(r.Context(), body.ID)
		if err != nil {
			if errors.Is(err, application.ErrUserNotFound) {
				helper.SendClientError(w, "user_not_found", nil)
				return
			} else {
				helper.SendServerError(w)
				return
			}
		}

		helper.SendEmptyResponse(w)
	}
}
