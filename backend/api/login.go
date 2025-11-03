package api

import (
	"errors"
	"net/http"

	"github.com/nicograef/jotti/backend/domain/auth"
	usr "github.com/nicograef/jotti/backend/domain/user"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

// LoginHandler handles user login requests by validating the password hash against the database
// and returns a jwt token if successful.
// If this is the first time the user logs in (no password hash set), it sets the provided password as the new password.
func LoginHandler(us *usr.Service, as *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !validateMethod(w, r, http.MethodPost) {
			return
		}

		body := loginRequest{}
		if !readJSONRequest(w, r, &body) {
			return
		}

		user, err := us.LoginUserViaPassword(body.Username, body.Password)
		if err != nil {
			if errors.Is(err, usr.ErrUserNotFound) || errors.Is(err, usr.ErrInvalidPassword) {
				sendUnauthorizedError(w, errorResponse{
					Message: "Invalid username or password",
					Code:    "invalid_credentials",
				})
			} else {
				sendInternalServerError(w)
			}
			return
		}

		stringToken, err := as.GenerateJWTTokenForUser(*user)
		if err != nil {
			sendInternalServerError(w)
			return
		}

		sendJSONResponse(w, loginResponse{
			Token: stringToken,
		})
	}
}
