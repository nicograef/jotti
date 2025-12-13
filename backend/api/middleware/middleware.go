package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/domain/jwt"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
)

type ContextKey string

const (
	UserIDKey        ContextKey = "userid"
	CorrelationIDKey ContextKey = "correlation_id"
)

// CorrelationIDMiddleware adds a correlation ID to each request for tracing
func CorrelationIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := r.Header.Get("X-Correlation-ID")
		if correlationID == "" {
			correlationID = uuid.NewString()[:8] // Shorten UUID for brevity
		}

		w.Header().Set("X-Correlation-ID", correlationID)

		ctx := context.WithValue(r.Context(), CorrelationIDKey, correlationID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoggingMiddleware logs HTTP requests with correlation ID
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		correlationID, _ := r.Context().Value(CorrelationIDKey).(string)
		logger := log.With().Str("correlation", correlationID).Logger()
		r = r.WithContext(logger.WithContext(r.Context()))

		// Create a response writer wrapper to capture status code
		ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(ww, r)

		logger.Info().
			Str("path", r.URL.Path).
			Int("status", ww.statusCode).
			Int64("duration_ms", time.Since(start).Milliseconds()).
			Msg("Request completed")
	})
}

// RateLimitMiddleware limits requests per IP
func RateLimitMiddleware(requestsPerSecond int) func(http.Handler) http.Handler {
	limiter := rate.NewLimiter(rate.Limit(requestsPerSecond), requestsPerSecond*2)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := zerolog.Ctx(r.Context())

			if !limiter.Allow() {
				logger.Warn().Msg("Rate limit exceeded")
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// PostMethodOnlyMiddleware middleware ensures the request method is POST
func PostMethodOnlyMiddleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := zerolog.Ctx(r.Context())

		if r.Method != http.MethodPost {
			logger.Error().Str("method", r.Method).Msg("Invalid method.")
			helper.SendClientError(w, "method_not_allowed", nil)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// NewJwtMiddleware validates the JWT Token in the Authorization header.
// If valid, it adds the user information to the request context.
func NewJwtMiddleware(jwtSecret string, allowedRoles []string) func(http.Handler) http.HandlerFunc {
	return func(h http.Handler) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				log.Error().Msg("Missing Authorization header")
				helper.SendClientError(w, "missing_authorization", nil)
				return
			}

			// get jwt token, remove "Bearer " prefix
			token = token[len("Bearer "):]
			userID, userRole, err := jwt.ParseAndValidateJWTToken(token, jwtSecret)
			if err != nil {
				log.Error().Err(err).Msg("Invalid JWT token")
				helper.SendClientError(w, "invalid_jwt", nil)
				return
			}

			// check if role is allowed
			roleAllowed := false
			for _, role := range allowedRoles {
				if userRole == role {
					roleAllowed = true
					break
				}
			}
			if !roleAllowed {
				helper.SendClientError(w, "insufficient_permissions", fmt.Sprintf("Insufficient permissions for role %s", userRole))
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, UserIDKey, userID)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
