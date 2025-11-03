package api

import (
	"net/http"

	"github.com/nicograef/jotti/backend/domain/user"
)

type createUserRequest struct {
	Name     string    `json:"name"`
	Username string    `json:"username"`
	Role     user.Role `json:"role"`
}

type createUserResponse = user.User

// CreateUserHandler handles requests to create a new user.
func CreateUserHandler(us *user.Service) http.HandlerFunc {
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
