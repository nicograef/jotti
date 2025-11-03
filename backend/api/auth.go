package api

import (
	"context"
	"net/http"

	"github.com/nicograef/jotti/backend/domain/auth"
)

func NewJWTMiddleware(a *auth.Service) func(http.Handler) http.HandlerFunc {
	return func(h http.Handler) http.HandlerFunc {
		return JWTMiddleware(a, h)
	}
}

// JWTMiddleware validates the JWT Token in the Authorization header.
// If valid, it adds the user information to the request context.
func JWTMiddleware(a *auth.Service, h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		// get jwt token, remove "Bearer " prefix
		token = token[len("Bearer "):]

		payload, err := a.ParseAndValidateJWTToken(token)
		if err != nil {
			http.Error(w, "Invalid JWT token", http.StatusUnauthorized)
			return
		}

		// Add user info to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, "UserID", payload.UserID)
		ctx = context.WithValue(ctx, "Role", payload.Role)
		r = r.WithContext(ctx)

		h.ServeHTTP(w, r)
	}
}

// AdminMiddleware ensures that the request is made by an admin user.
func AdminMiddleware(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value("Role").(string)
		if !ok || role != "admin" {
			http.Error(w, "Forbidden: Admins only", http.StatusForbidden)
			return
		}
		h.ServeHTTP(w, r)
	}
}
