package api

import (
	"net/http"

	"github.com/nicograef/jotti/backend/domain/user"
)

type CreateUserRequest struct {
	Name     string        `json:"name"`
	Username string        `json:"username"`
	Role     user.UserRole `json:"role"`
}

type CreateUserResponse = user.User

type UserService interface {
	CreateUserWithoutPassword(name, username string, role user.UserRole) (*user.User, error)
}

func CreateUserHandler(us UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !validateMethod(w, r, http.MethodPost) {
			return
		}

		body := CreateUserRequest{}
		if !readJSONRequest(w, r, &body) {
			return
		}

		user, err := us.CreateUserWithoutPassword(body.Name, body.Username, body.Role)
		if err != nil {
			sendInternalServerError(w)
			return
		}

		sendJSONResponse(w, CreateUserResponse{
			ID:       user.ID,
			Name:     user.Name,
			Username: user.Username,
			Role:     user.Role,
		})
	}
}
