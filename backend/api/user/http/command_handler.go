package http

import (
	"context"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/middleware"
	"github.com/nicograef/jotti/backend/api/user/application"
	"github.com/nicograef/jotti/backend/domain/user"
)

type command interface {
	CreateUser(ctx context.Context, name, username string, role user.Role) (int, string, error)
	UpdateUser(ctx context.Context, id int, name, username string, role user.Role) error
	ActivateUser(ctx context.Context, id int) error
	DeactivateUser(ctx context.Context, id int) error
	DeleteUser(ctx context.Context, id int) error
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
			helper.MapError(w, err, map[error]string{
				application.ErrUsernameAlreadyExists: "username_already_exists",
			})
			return
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
			helper.MapError(w, err, map[error]string{
				application.ErrUserNotFound:          "user_not_found",
				application.ErrUsernameAlreadyExists: "username_already_exists",
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}

type resetPassword struct {
	ID int `json:"id"`
}

var resetPasswordSchema = z.Struct(z.Shape{
	"ID": user.IDSchema.Required(),
})

type resetPasswordResponse struct {
	OnetimePassword string `json:"onetimePassword"`
}

// ResetPasswordHandler handles requests to reset a user's password.
func (h CommandHandler) ResetPasswordHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := resetPassword{}
		if !helper.ReadAndValidateBody(w, r, &body, resetPasswordSchema) {
			return
		}

		onetimePassword, err := h.Command.ResetPassword(r.Context(), body.ID)
		if err != nil {
			helper.MapError(w, err, map[error]string{
				application.ErrUserNotFound: "user_not_found",
			})
			return
		}

		helper.SendResponse(w, resetPasswordResponse{OnetimePassword: onetimePassword})
	}
}

type activateUser struct {
	ID int `json:"id"`
}

var activateUserSchema = z.Struct(z.Shape{
	"ID": user.IDSchema.Required(),
})

// ActivateUserHandler handles requests to activate a user.
func (h CommandHandler) ActivateUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := activateUser{}
		if !helper.ReadAndValidateBody(w, r, &body, activateUserSchema) {
			return
		}

		err := h.Command.ActivateUser(r.Context(), body.ID)
		if err != nil {
			helper.MapError(w, err, map[error]string{
				application.ErrUserNotFound: "user_not_found",
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}

type deactivateUser struct {
	ID int `json:"id"`
}

var deactivateUserSchema = z.Struct(z.Shape{
	"ID": user.IDSchema.Required(),
})

// DeactivateUserHandler handles requests to deactivate a user.
func (h CommandHandler) DeactivateUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := deactivateUser{}
		if !helper.ReadAndValidateBody(w, r, &body, deactivateUserSchema) {
			return
		}

		err := h.Command.DeactivateUser(r.Context(), body.ID)
		if err != nil {
			helper.MapError(w, err, map[error]string{
				application.ErrUserNotFound: "user_not_found",
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}

type deleteUser struct {
	ID int `json:"id"`
}

var deleteUserSchema = z.Struct(z.Shape{
	"ID": user.IDSchema.Required(),
})

func (h CommandHandler) DeleteUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := deleteUser{}
		if !helper.ReadAndValidateBody(w, r, &body, deleteUserSchema) {
			return
		}

		currentUserID, ok := r.Context().Value(middleware.UserIDKey).(int)
		if !ok {
			helper.SendServerError(w)
			return
		}
		if body.ID == currentUserID {
			helper.SendClientError(w, "cannot_delete_self", nil)
			return
		}

		err := h.Command.DeleteUser(r.Context(), body.ID)
		if err != nil {
			helper.MapError(w, err, map[error]string{
				application.ErrUserNotFound: "user_not_found",
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}
