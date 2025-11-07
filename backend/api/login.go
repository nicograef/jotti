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

// LoginHandler handles user login requests by validating the password hash against the database and returns a jwt token if successful.
func LoginHandler(us *usr.Service, as *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !validateMethod(w, r, http.MethodPost) {
			return
		}

		body := loginRequest{}
		if !readJSONRequest(w, r, &body) {
			return
		}

		user, err := us.VerifyPasswordAndGetUser(body.Username, body.Password)
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

		if user.Locked {
			sendUnauthorizedError(w, errorResponse{
				Message: "User account is locked",
				Code:    "user_locked",
			})
			return
		}

		stringToken, err := as.GenerateJWTTokenForUser(*user)
		if err != nil {
			sendInternalServerError(w)
			return
		}

		sendResponse(w, loginResponse{
			Token: stringToken,
		})
	}
}

type setPasswordRequest struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	OnetimePassword string `json:"onetimePassword"`
}

type setPasswordResponse struct {
	Token string `json:"token"`
}

// SetPasswordHandler handles setting a new password for a user using a one-time password and returns a jwt token if successful.
func SetPasswordHandler(us *usr.Service, as *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !validateMethod(w, r, http.MethodPost) {
			return
		}

		body := setPasswordRequest{}
		if !readJSONRequest(w, r, &body) {
			return
		}

		user, err := us.SetNewPassword(body.Username, body.Password, body.OnetimePassword)
		if err != nil {
			if errors.Is(err, usr.ErrUserNotFound) || errors.Is(err, usr.ErrInvalidPassword) {
				sendUnauthorizedError(w, errorResponse{
					Message: "Invalid username or password",
					Code:    "invalid_credentials",
				})
			} else if errors.Is(err, usr.ErrNoOnetimePassword) {
				sendBadRequestError(w, errorResponse{
					Message: "No one-time password set for user. User probably already has a password.",
					Code:    "already_has_password",
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

		sendResponse(w, setPasswordResponse{
			Token: stringToken,
		})
	}
}
