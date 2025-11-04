package api

import (
	"errors"
	"net/http"

	usr "github.com/nicograef/jotti/backend/domain/user"
)

type createUserRequest struct {
	Name     string   `json:"name"`
	Username string   `json:"username"`
	Role     usr.Role `json:"role"`
}

type createUserResponse = usr.User

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

		user, err := us.CreateUserWithoutPassword(body.Name, body.Username, body.Role)
		if err != nil {
			sendInternalServerError(w)
			return
		}

		sendJSONResponse(w, createUserResponse{
			ID:       user.ID,
			Name:     user.Name,
			Username: user.Username,
			Role:     user.Role,
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

type updateUserResponse = usr.User

// CreateUserHandler handles requests to create a new user.
func UpdateUserHandler(us *usr.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !validateMethod(w, r, http.MethodPost) {
			return
		}

		body := updateUserRequest{}
		if !readJSONRequest(w, r, &body) {
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

		sendJSONResponse(w, updateUserResponse{
			ID:       user.ID,
			Name:     user.Name,
			Username: user.Username,
			Role:     user.Role,
			Locked:   user.Locked,
		})
	}
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

		sendJSONResponse(w, users)
	}
}
