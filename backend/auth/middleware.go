package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/nicograef/jotti/backend/api"
	"github.com/nicograef/jotti/backend/user"
	"github.com/rs/zerolog/log"
)

// Context key types to avoid collisions
type ContextKey string

const (
	UserIDKey ContextKey = "userid"
)

// NewAdminMiddleware ensures that the request is made by an admin user.
func NewAdminMiddleware(secret string) func(http.Handler) http.HandlerFunc {
	return func(h http.Handler) http.HandlerFunc {
		return jwtMiddleware(secret, []user.Role{user.AdminRole}, h)
	}
}

// NewServiceMiddleware ensures that the request is made by a service or admin user.
func NewServiceMiddleware(secret string) func(http.Handler) http.HandlerFunc {
	return func(h http.Handler) http.HandlerFunc {
		return jwtMiddleware(secret, []user.Role{user.AdminRole, user.ServiceRole}, h)
	}
}

// jwtMiddleware validates the JWT Token in the Authorization header.
// If valid, it adds the user information to the request context.
func jwtMiddleware(jwtSecret string, allowedRoles []user.Role, h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			log.Error().Msg("Missing Authorization header")
			api.SendClientError(w, "missing_authorization", nil)
			return
		}

		// get jwt token, remove "Bearer " prefix
		token = token[len("Bearer "):]
		payload, err := parseAndValidateJWTToken(token, jwtSecret)
		if err != nil {
			log.Error().Err(err).Msg("Invalid JWT token")
			api.SendClientError(w, "invalid_jwt", nil)
			return
		}

		// check if role is allowed
		roleAllowed := false
		for _, role := range allowedRoles {
			if payload.Role == role {
				roleAllowed = true
				break
			}
		}
		if !roleAllowed {
			api.SendClientError(w, "insufficient_permissions", fmt.Sprintf("Insufficient permissions for role %s", payload.Role))
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, UserIDKey, payload.UserID)

		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
