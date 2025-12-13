package user

import (
	"context"
	"errors"
	"net/http"

	"github.com/nicograef/jotti/backend/api"
)

type command interface {
	CreateUser(ctx context.Context, name, username string, role Role) (int, string, error)
	UpdateUser(ctx context.Context, id int, name, username string, role Role) error
	ActivateUser(ctx context.Context, id int) error
	DeactivateUser(ctx context.Context, id int) error
	ResetPassword(ctx context.Context, userID int) (string, error)
}

type CommandHandler struct {
	Command command
}

type createUser struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Role     Role   `json:"role"`
}

type createUserResponse struct {
	ID              int    `json:"id"`
	OnetimePassword string `json:"onetimePassword"`
}

// CreateUserHandler handles requests to create a new user.
func (h *CommandHandler) CreateUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := createUser{}
		if !api.ReadBody(w, r, &body) {
			return
		}

		ctx := r.Context()
		userID, onetimePassword, err := h.Command.CreateUser(ctx, body.Name, body.Username, body.Role)
		if err != nil {
			if errors.Is(err, ErrUsernameAlreadyExists) {
				api.SendClientError(w, "username_already_exists", nil)
				return
			} else {
				api.SendServerError(w)
				return
			}
		}

		api.SendResponse(w, createUserResponse{ID: userID, OnetimePassword: onetimePassword})
	}
}

type updateUser struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Role     Role   `json:"role"`
}

// UpdateUserHandler handles requests to update a user.
func (h *CommandHandler) UpdateUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := updateUser{}
		if !api.ReadBody(w, r, &body) {
			return
		}

		ctx := r.Context()
		err := h.Command.UpdateUser(ctx, body.ID, body.Name, body.Username, body.Role)
		if err != nil {
			if errors.Is(err, ErrUserNotFound) {
				api.SendClientError(w, "user_not_found", nil)
				return
			} else if errors.Is(err, ErrUsernameAlreadyExists) {
				api.SendClientError(w, "username_already_exists", nil)
				return
			} else {
				api.SendServerError(w)
				return
			}
		}

		api.SendEmptyResponse(w)
	}
}

type resetPassword struct {
	ID int `json:"id"`
}

type resetPasswordResponse struct {
	OnetimePassword string `json:"onetimePassword"`
}

// ResetPasswordHandler handles requests to reset a user's password.
func (h *CommandHandler) ResetPasswordHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := resetPassword{}
		if !api.ReadBody(w, r, &body) {
			return
		}

		ctx := r.Context()
		onetimePassword, err := h.Command.ResetPassword(ctx, body.ID)
		if err != nil {
			if errors.Is(err, ErrUserNotFound) {
				api.SendClientError(w, "user_not_found", nil)
				return
			} else {
				api.SendServerError(w)
				return
			}
		}

		api.SendResponse(w, resetPasswordResponse{OnetimePassword: onetimePassword})
	}
}

type activateUser struct {
	ID int `json:"id"`
}

// ActivateUserHandler handles requests to activate a user.
func (h *CommandHandler) ActivateUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := activateUser{}
		if !api.ReadBody(w, r, &body) {
			return
		}

		ctx := r.Context()
		err := h.Command.ActivateUser(ctx, body.ID)
		if err != nil {
			if errors.Is(err, ErrUserNotFound) {
				api.SendClientError(w, "user_not_found", nil)
				return
			} else {
				api.SendServerError(w)
				return
			}
		}

		api.SendEmptyResponse(w)
	}
}

type deactivateUser struct {
	ID int `json:"id"`
}

// DeactivateUserHandler handles requests to deactivate a user.
func (h *CommandHandler) DeactivateUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := deactivateUser{}
		if !api.ReadBody(w, r, &body) {
			return
		}

		ctx := r.Context()
		err := h.Command.DeactivateUser(ctx, body.ID)
		if err != nil {
			if errors.Is(err, ErrUserNotFound) {
				api.SendClientError(w, "user_not_found", nil)
				return
			} else {
				api.SendServerError(w)
				return
			}
		}

		api.SendEmptyResponse(w)
	}
}
