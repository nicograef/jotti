package http

import (
	"context"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/middleware"
	"github.com/nicograef/jotti/backend/api/stammdaten/user/application"
	"github.com/nicograef/jotti/backend/domain/user"
)

type command interface {
	CreateUser(ctx context.Context, name, username string, role user.Role) (int, string, error)
	UpdateUser(ctx context.Context, actorID, id int, name, username string, role user.Role) error
	ActivateUser(ctx context.Context, id int) error
	DeactivateUser(ctx context.Context, actorID, id int) error
	DeleteUser(ctx context.Context, actorID, id int) error
	ResetPassword(ctx context.Context, userID int) (string, error)
}

type CommandHandler struct {
	Command command
}

type createUserRequest struct {
	Name     string    `json:"name"`
	Username string    `json:"username"`
	Role     user.Role `json:"role"`
}

var createUserSchema = z.Struct(z.Shape{
	"Name":     user.NameSchema.Required(),
	"Username": user.UsernameSchema.Required(),
	"Role":     user.RoleSchema.Required(),
})

type createUserResponse struct {
	ID              int    `json:"id"`
	OnetimePassword string `json:"onetimePassword"`
}

func (h *CommandHandler) CreateUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := createUserRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, createUserSchema) {
			return
		}

		userID, onetimePassword, err := h.Command.CreateUser(r.Context(), body.Name, body.Username, body.Role)
		if err != nil {
			helper.MapError(w, err, []helper.ErrorCode{
				{Err: application.ErrUsernameAlreadyExists, Code: "username_already_exists"},
			})
			return
		}

		helper.SendResponse(w, createUserResponse{ID: userID, OnetimePassword: onetimePassword})
	}
}

type updateUserRequest struct {
	ID       int       `json:"id"`
	Name     string    `json:"name"`
	Username string    `json:"username"`
	Role     user.Role `json:"role"`
}

var updateUserSchema = z.Struct(z.Shape{
	"ID":       user.IDSchema.Required(),
	"Name":     user.NameSchema.Required(),
	"Username": user.UsernameSchema.Required(),
	"Role":     user.RoleSchema.Required(),
})

func (h *CommandHandler) UpdateUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := updateUserRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, updateUserSchema) {
			return
		}

		actorID, ok := actor(w, r)
		if !ok {
			return
		}

		err := h.Command.UpdateUser(r.Context(), actorID, body.ID, body.Name, body.Username, body.Role)
		if err != nil {
			helper.MapError(w, err, []helper.ErrorCode{
				{Err: application.ErrCannotDemoteSelf, Code: "cannot_demote_self"},
				{Err: application.ErrUserNotFound, Code: "user_not_found"},
				{Err: application.ErrUsernameAlreadyExists, Code: "username_already_exists"},
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}

// idRequest is the body of every endpoint that addresses one entity by its ID.
type idRequest struct {
	ID int `json:"id"`
}

var idSchema = z.Struct(z.Shape{
	"ID": user.IDSchema.Required(),
})

type resetPasswordResponse struct {
	OnetimePassword string `json:"onetimePassword"`
}

func (h *CommandHandler) ResetPasswordHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := idRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, idSchema) {
			return
		}

		onetimePassword, err := h.Command.ResetPassword(r.Context(), body.ID)
		if err != nil {
			helper.MapError(w, err, []helper.ErrorCode{
				{Err: application.ErrUserNotFound, Code: "user_not_found"},
			})
			return
		}

		helper.SendResponse(w, resetPasswordResponse{OnetimePassword: onetimePassword})
	}
}

func (h *CommandHandler) ActivateUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := idRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, idSchema) {
			return
		}

		err := h.Command.ActivateUser(r.Context(), body.ID)
		if err != nil {
			helper.MapError(w, err, []helper.ErrorCode{
				{Err: application.ErrUserNotFound, Code: "user_not_found"},
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}

func (h *CommandHandler) DeactivateUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := idRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, idSchema) {
			return
		}

		actorID, ok := actor(w, r)
		if !ok {
			return
		}

		err := h.Command.DeactivateUser(r.Context(), actorID, body.ID)
		if err != nil {
			helper.MapError(w, err, []helper.ErrorCode{
				{Err: application.ErrCannotDeactivateSelf, Code: "cannot_deactivate_self"},
				{Err: application.ErrUserNotFound, Code: "user_not_found"},
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}

func (h *CommandHandler) DeleteUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := idRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, idSchema) {
			return
		}

		actorID, ok := actor(w, r)
		if !ok {
			return
		}

		err := h.Command.DeleteUser(r.Context(), actorID, body.ID)
		if err != nil {
			helper.MapError(w, err, []helper.ErrorCode{
				{Err: application.ErrCannotDeleteSelf, Code: "cannot_delete_self"},
				{Err: application.ErrUserNotFound, Code: "user_not_found"},
			})
			return
		}

		helper.SendEmptyResponse(w)
	}
}

// actor is the authenticated admin; the JWT middleware always sets it, so its absence is a server fault.
func actor(w http.ResponseWriter, r *http.Request) (int, bool) {
	actorID, _, ok := middleware.UserFromContext(r.Context())
	if !ok {
		helper.SendServerError(w)
	}
	return actorID, ok
}
