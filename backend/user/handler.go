package user

import (
	"context"
	"errors"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api"
)

type service interface {
	CreateUser(ctx context.Context, name, username string, role Role) (*User, string, error)
	UpdateUser(ctx context.Context, id int, name, username string, role Role) (*User, error)
	ActivateUser(ctx context.Context, id int) error
	DeactivateUser(ctx context.Context, id int) error
	GetAllUsers(ctx context.Context) ([]*User, error)
	ResetPassword(ctx context.Context, userID int) (string, error)
}

type Handler struct {
	Service service
}

type createUserRequest struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Role     Role   `json:"role"`
}

var createUserRequestSchema = z.Struct(z.Shape{
	"Name":     NameSchema.Required(),
	"Username": UsernameSchema.Required(),
	"Role":     RoleSchema.Required(),
})

type createUserResponse struct {
	User            User   `json:"user"`
	OnetimePassword string `json:"onetimePassword"`
}

// CreateUserHandler handles requests to create a new user.
func (h *Handler) CreateUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := createUserRequest{}
		if !api.ReadAndValidateBody(w, r, &body, createUserRequestSchema) {
			return
		}

		ctx := r.Context()
		user, onetimePassword, err := h.Service.CreateUser(ctx, body.Name, body.Username, body.Role)
		if err != nil {
			api.SendInternalServerError(w)
			return
		}

		api.SendResponse(w, createUserResponse{
			User:            *user,
			OnetimePassword: onetimePassword,
		})
	}
}

type updateUserRequest struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Role     Role   `json:"role"`
}

var updateUserRequestSchema = z.Struct(z.Shape{
	"ID":       IDSchema.Required(),
	"Name":     NameSchema.Required(),
	"Username": UsernameSchema.Required(),
	"Role":     RoleSchema.Required(),
})

type updateUserResponse = struct {
	User User `json:"user"`
}

// UpdateUserHandler handles requests to update a user.
func (h *Handler) UpdateUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := updateUserRequest{}
		if !api.ReadAndValidateBody(w, r, &body, updateUserRequestSchema) {
			return
		}

		ctx := r.Context()
		user, err := h.Service.UpdateUser(ctx, body.ID, body.Name, body.Username, body.Role)
		if err != nil {
			if errors.Is(err, ErrUserNotFound) {
				api.SendNotFoundError(w, api.ErrorResponse{
					Message: "User not found",
					Code:    "user_not_found",
				})
			} else {
				api.SendInternalServerError(w)
			}
			return
		}

		api.SendResponse(w, updateUserResponse{
			User: *user,
		})
	}
}

type getUsersResponse = struct {
	Users []*User `json:"users"`
}

// GetAllUsersHandler handles requests to retrieve all users.
func (h *Handler) GetAllUsersHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		users, err := h.Service.GetAllUsers(ctx)
		if err != nil {
			api.SendInternalServerError(w)
			return
		}

		api.SendResponse(w, getUsersResponse{
			Users: users,
		})
	}
}

type resetPasswordRequest struct {
	UserID int `json:"userID"`
}

var resetPasswordRequestSchema = z.Struct(z.Shape{
	"UserID": IDSchema.Required(),
})

type resetPasswordResponse struct {
	OnetimePassword string `json:"onetimePassword"`
}

// ResetPasswordHandler handles requests to reset a user's password.
func (h *Handler) ResetPasswordHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := resetPasswordRequest{}
		if !api.ReadAndValidateBody(w, r, &body, resetPasswordRequestSchema) {
			return
		}

		ctx := r.Context()
		onetimePassword, err := h.Service.ResetPassword(ctx, body.UserID)
		if err != nil {
			if errors.Is(err, ErrUserNotFound) {
				api.SendNotFoundError(w, api.ErrorResponse{
					Message: "User not found",
					Code:    "user_not_found",
				})
			} else {
				api.SendInternalServerError(w)
			}
			return
		}

		api.SendResponse(w, resetPasswordResponse{
			OnetimePassword: onetimePassword,
		})
	}
}

type activateUserRequest struct {
	ID int `json:"id"`
}

var activateUserRequestSchema = z.Struct(z.Shape{
	"ID": IDSchema.Required(),
})

// ActivateUserHandler handles requests to activate a user.
func (h *Handler) ActivateUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		body := activateUserRequest{}
		if !api.ReadAndValidateBody(w, r, &body, activateUserRequestSchema) {
			return
		}

		err := h.Service.ActivateUser(ctx, body.ID)
		if err != nil {
			if errors.Is(err, ErrUserNotFound) {
				api.SendNotFoundError(w, api.ErrorResponse{
					Message: "User not found",
					Code:    "user_not_found",
				})
			} else {
				api.SendInternalServerError(w)
			}
			return
		}

		api.SendEmptyResponse(w)
	}
}

type deactivateUserRequest struct {
	ID int `json:"id"`
}

var deactivateUserRequestSchema = z.Struct(z.Shape{
	"ID": IDSchema.Required(),
})

// DeactivateUserHandler handles requests to deactivate a user.
func (h *Handler) DeactivateUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := deactivateUserRequest{}
		if !api.ReadAndValidateBody(w, r, &body, deactivateUserRequestSchema) {
			return
		}

		ctx := r.Context()
		err := h.Service.DeactivateUser(ctx, body.ID)
		if err != nil {
			if errors.Is(err, ErrUserNotFound) {
				api.SendNotFoundError(w, api.ErrorResponse{
					Message: "User not found",
					Code:    "user_not_found",
				})
			} else {
				api.SendInternalServerError(w)
			}
			return
		}

		api.SendEmptyResponse(w)
	}
}
