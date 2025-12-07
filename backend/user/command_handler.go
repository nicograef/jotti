package user

import (
	"context"
	"errors"
	"net/http"

	z "github.com/Oudwins/zog"
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

var createUserSchema = z.Struct(z.Shape{
	"Name":     NameSchema.Required(),
	"Username": UsernameSchema.Required(),
	"Role":     RoleSchema.Required(),
})

type createUserResponse struct {
	ID              int    `json:"id"`
	OnetimePassword string `json:"onetimePassword"`
}

// CreateUserHandler handles requests to create a new user.
func (h *CommandHandler) CreateUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := createUser{}
		if !api.ReadAndValidateBody(w, r, &body, createUserSchema) {
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

var updateUserSchema = z.Struct(z.Shape{
	"ID":       IDSchema.Required(),
	"Name":     NameSchema.Required(),
	"Username": UsernameSchema.Required(),
	"Role":     RoleSchema.Required(),
})

// UpdateUserHandler handles requests to update a user.
func (h *CommandHandler) UpdateUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := updateUser{}
		if !api.ReadAndValidateBody(w, r, &body, updateUserSchema) {
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
	UserID int `json:"userID"`
}

var resetPasswordSchema = z.Struct(z.Shape{
	"UserID": IDSchema.Required(),
})

type resetPasswordResponse struct {
	OnetimePassword string `json:"onetimePassword"`
}

// ResetPasswordHandler handles requests to reset a user's password.
func (h *CommandHandler) ResetPasswordHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := resetPassword{}
		if !api.ReadAndValidateBody(w, r, &body, resetPasswordSchema) {
			return
		}

		ctx := r.Context()
		onetimePassword, err := h.Command.ResetPassword(ctx, body.UserID)
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

var activateUserSchema = z.Struct(z.Shape{
	"ID": IDSchema.Required(),
})

// ActivateUserHandler handles requests to activate a user.
func (h *CommandHandler) ActivateUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := activateUser{}
		if !api.ReadAndValidateBody(w, r, &body, activateUserSchema) {
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

var deactivateUserSchema = z.Struct(z.Shape{
	"ID": IDSchema.Required(),
})

// DeactivateUserHandler handles requests to deactivate a user.
func (h *CommandHandler) DeactivateUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := deactivateUser{}
		if !api.ReadAndValidateBody(w, r, &body, deactivateUserSchema) {
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
