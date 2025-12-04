package auth

import (
	"context"
	"errors"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api"
	usr "github.com/nicograef/jotti/backend/user"
)

type userService interface {
	VerifyPasswordAndGetUser(ctx context.Context, username, password string) (*usr.User, error)
	SetNewPassword(ctx context.Context, username, password, onetimePassword string) (*usr.User, error)
}

type Handler struct {
	UserService userService
	JWTSecret   string
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var loginRequestSchema = z.Struct(z.Shape{
	"Username": usr.UsernameSchema.Required(),
	"Password": usr.PasswordSchema.Required(),
})

type loginResponse struct {
	Token string `json:"token"`
}

// LoginHandler handles user login requests by validating the password hash against the database and returns a jwt token if successful.
func (h *Handler) LoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := loginRequest{}
		if !api.ReadAndValidateBody(w, r, &body, loginRequestSchema) {
			return
		}

		ctx := r.Context()
		user, err := h.UserService.VerifyPasswordAndGetUser(ctx, body.Username, body.Password)
		if err != nil {
			if errors.Is(err, usr.ErrUserNotFound) || errors.Is(err, usr.ErrInvalidPassword) {
				api.SendUnauthorizedError(w, api.ErrorResponse{
					Message: "Invalid username or password",
					Code:    "invalid_credentials",
				})
			} else {
				api.SendInternalServerError(w)
			}
			return
		}

		if user.Status != usr.ActiveStatus {
			api.SendUnauthorizedError(w, api.ErrorResponse{
				Message: "User account is not active",
				Code:    "user_inactive",
			})
			return
		}

		stringToken, err := generateJWTTokenForUser(*user, h.JWTSecret)
		if err != nil {
			api.SendInternalServerError(w)
			return
		}

		api.SendResponse(w, loginResponse{
			Token: stringToken,
		})
	}
}

type setPasswordRequest struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	OnetimePassword string `json:"onetimePassword"`
}

var setPasswordRequestSchema = z.Struct(z.Shape{
	"Username":        usr.UsernameSchema.Required(),
	"Password":        usr.PasswordSchema.Required(),
	"OnetimePassword": usr.OnetimePasswordSchema.Required(),
})

type setPasswordResponse struct {
	Token string `json:"token"`
}

// SetPasswordHandler handles setting a new password for a user using a one-time password and returns a jwt token if successful.
func (h *Handler) SetPasswordHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := setPasswordRequest{}
		if !api.ReadAndValidateBody(w, r, &body, setPasswordRequestSchema) {
			return
		}

		ctx := r.Context()
		user, err := h.UserService.SetNewPassword(ctx, body.Username, body.Password, body.OnetimePassword)
		if err != nil {
			if errors.Is(err, usr.ErrUserNotFound) || errors.Is(err, usr.ErrInvalidPassword) {
				api.SendUnauthorizedError(w, api.ErrorResponse{
					Message: "Invalid username or password",
					Code:    "invalid_credentials",
				})
			} else if errors.Is(err, usr.ErrNoOnetimePassword) {
				api.SendBadRequestError(w, api.ErrorResponse{
					Message: "No one-time password set for user. User probably already has a password.",
					Code:    "already_has_password",
				})
			} else {
				api.SendInternalServerError(w)
			}
			return
		}

		stringToken, err := generateJWTTokenForUser(*user, h.JWTSecret)
		if err != nil {
			api.SendInternalServerError(w)
			return
		}

		api.SendResponse(w, setPasswordResponse{
			Token: stringToken,
		})
	}
}
