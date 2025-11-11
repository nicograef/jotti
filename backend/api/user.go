package api

import (
	"errors"
	"net/http"

	z "github.com/Oudwins/zog"

	usr "github.com/nicograef/jotti/backend/domain/user"
)

type createUserRequest struct {
	Name     string   `json:"name"`
	Username string   `json:"username"`
	Role     usr.Role `json:"role"`
}

var createUserRequestSchema = z.Struct(z.Shape{
	"Name":     usr.NameSchema.Required(),
	"Username": usr.UsernameSchema.Required(),
	"Role":     usr.RoleSchema.Required(),
})

type createUserResponse struct {
	User            usr.User `json:"user"`
	OnetimePassword string   `json:"onetimePassword"`
}

// CreateUserHandler handles requests to create a new user.
func CreateUserHandler(us *usr.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !validateMethod(w, r, http.MethodPost) {
			return
		}

		body := createUserRequest{}
		if !readJSONRequest(w, r, &body) {
			return
		}

		if !validateBody(w, &body, createUserRequestSchema) {
			return
		}

		user, onetimePassword, err := us.CreateUser(body.Name, body.Username, body.Role)
		if err != nil {
			sendInternalServerError(w)
			return
		}

		sendResponse(w, createUserResponse{
			User:            *user,
			OnetimePassword: onetimePassword,
		})
	}
}

type updateUserRequest struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Username string   `json:"username"`
	Role     usr.Role `json:"role"`
	Locked   bool     `json:"locked"`
}

var updateUserRequestSchema = z.Struct(z.Shape{
	"ID":       usr.IDSchema.Required(),
	"Name":     usr.NameSchema.Required(),
	"Username": usr.UsernameSchema.Required(),
	"Role":     usr.RoleSchema.Required(),
	"Locked":   z.Bool().Required(),
})

type updateUserResponse = struct {
	User usr.User `json:"user"`
}

// UpdateUserHandler handles requests to update an existing user.
func UpdateUserHandler(us *usr.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !validateMethod(w, r, http.MethodPost) {
			return
		}

		body := updateUserRequest{}
		if !readJSONRequest(w, r, &body) {
			return
		}

		if !validateBody(w, &body, updateUserRequestSchema) {
			return
		}

		user, err := us.UpdateUser(body.ID, body.Name, body.Username, body.Role, body.Locked)
		if err != nil && errors.Is(err, usr.ErrUserNotFound) {
			sendNotFoundError(w, errorResponse{
				Message: "User not found",
				Code:    "user_not_found",
			})
			return
		} else if err != nil {
			sendInternalServerError(w)
			return
		}

		sendResponse(w, updateUserResponse{
			User: *user,
		})
	}
}

type getUsersResponse = struct {
	Users []*usr.User `json:"users"`
}

// GetUsersHandler handles requests to retrieve all users.
func GetUsersHandler(us *usr.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !validateMethod(w, r, http.MethodPost) {
			return
		}

		users, err := us.GetAllUsers()
		if err != nil {
			sendInternalServerError(w)
			return
		}

		sendResponse(w, getUsersResponse{
			Users: users,
		})
	}
}

type resetPasswordRequest struct {
	UserID int `json:"userID"`
}

var resetPasswordRequestSchema = z.Struct(z.Shape{
	"UserID": usr.IDSchema.Required(),
})

type resetPasswordResponse struct {
	OnetimePassword string `json:"onetimePassword"`
}

// ResetPasswordHandler handles requests to reset a user's password.
func ResetPasswordHandler(us *usr.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !validateMethod(w, r, http.MethodPost) {
			return
		}

		body := resetPasswordRequest{}
		if !readJSONRequest(w, r, &body) {
			return
		}

		if !validateBody(w, &body, resetPasswordRequestSchema) {
			return
		}

		onetimePassword, err := us.ResetPassword(body.UserID)
		if err != nil && errors.Is(err, usr.ErrUserNotFound) {
			sendNotFoundError(w, errorResponse{
				Message: "User not found",
				Code:    "user_not_found",
			})
			return
		} else if err != nil {
			sendInternalServerError(w)
			return
		}

		sendResponse(w, resetPasswordResponse{
			OnetimePassword: onetimePassword,
		})
	}
}
