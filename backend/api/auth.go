package api

import (
	"context"
	"net/http"

	"github.com/nicograef/jotti/backend/domain/auth"
)

// Context key types to avoid collisions
type contextKey string

const (
	userIDKey contextKey = "UserID"
	roleKey   contextKey = "Role"
)

// NewJWTMiddleware creates a new JWT middleware instance.
// It validates JWT tokens and adds user info to the request context.
func NewJWTMiddleware(a *auth.Service) func(http.Handler) http.HandlerFunc {
	return func(h http.Handler) http.HandlerFunc {
		return jwtMiddleware(a, h)
	}
}

// jwtMiddleware validates the JWT Token in the Authorization header.
// If valid, it adds the user information to the request context.
func jwtMiddleware(a *auth.Service, h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			sendUnauthorizedError(w, errorResponse{
				Message: "Missing Authorization header",
				Code:    "missing_authorization",
			})
			return
		}

		// get jwt token, remove "Bearer " prefix
		token = token[len("Bearer "):]

		payload, err := a.ParseAndValidateJWTToken(token)
		if err != nil {
			sendUnauthorizedError(w, errorResponse{
				Message: "Invalid JWT token",
				Code:    "invalid_jwt",
			})
			return
		}

		// Add user info to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, userIDKey, payload.UserID)
		ctx = context.WithValue(ctx, roleKey, payload.Role)
		r = r.WithContext(ctx)

		h.ServeHTTP(w, r)
	}
}

// AdminMiddleware ensures that the request is made by an admin user
// by checking the "Role" value in the request context.
// It should therefore be used after the JWT middleware.
func AdminMiddleware(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value(roleKey).(string)
		if !ok || role != "admin" {
			sendForbiddenError(w, errorResponse{
				Message: "Admin access required",
				Code:    "admin_access_required",
			})
			return
		}
		h.ServeHTTP(w, r)
	}
}
