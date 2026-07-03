package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/jwt"
	"github.com/nicograef/jotti/backend/domain/user"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
)

type ContextKey string

const (
	UserIDKey        ContextKey = "userid"
	UserNameKey      ContextKey = "username"
	CorrelationIDKey ContextKey = "correlation_id"
)

// UserFromContext returns the authenticated user's ID and name from the request
// context, as populated by NewJwtMiddleware. ok is false when no user ID is present.
func UserFromContext(ctx context.Context) (userID int, userName string, ok bool) {
	userID, ok = ctx.Value(UserIDKey).(int)
	if !ok {
		return 0, "", false
	}
	userName, _ = ctx.Value(UserNameKey).(string)
	return userID, userName, true
}

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
		start := time.Now().UTC()

		correlationID, _ := r.Context().Value(CorrelationIDKey).(string)
		logger := log.With().Str("correlation", correlationID).Logger()
		r = r.WithContext(logger.WithContext(r.Context()))

		ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(ww, r)

		logger.Info().
			Str("path", r.URL.Path).
			Int("status", ww.statusCode).
			Int64("duration_ms", time.Since(start).Milliseconds()).
			Msg("Request completed")
	})
}

// limiterEntry wraps a rate limiter with a last-seen timestamp for cleanup.
type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimitMiddleware limits requests per IP address
func RateLimitMiddleware(requestsPerSecond int) func(http.Handler) http.Handler {
	var mu sync.Mutex
	limiters := make(map[string]*limiterEntry)

	// Cleanup goroutine: remove entries not seen for 10+ minutes
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			mu.Lock()
			for ip, entry := range limiters {
				if time.Since(entry.lastSeen) > 10*time.Minute {
					delete(limiters, ip)
				}
			}
			mu.Unlock()
		}
	}()

	getLimiter := func(ip string) *rate.Limiter {
		mu.Lock()
		defer mu.Unlock()

		if entry, exists := limiters[ip]; exists {
			entry.lastSeen = time.Now().UTC()
			return entry.limiter
		}

		limiter := rate.NewLimiter(rate.Limit(requestsPerSecond), requestsPerSecond*2)
		limiters[ip] = &limiterEntry{limiter: limiter, lastSeen: time.Now().UTC()}
		return limiter
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := zerolog.Ctx(r.Context())

			ip := clientIP(r)
			limiter := getLimiter(ip)
			if !limiter.Allow() {
				logger.Warn().Str("ip", ip).Msg("Rate limit exceeded")
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// clientIP bestimmt die Client-IP für das Rate-Limiting. jotti läuft produktiv
// hinter dem eigenen Reverse-Proxy (Caddy), der die echte Client-IP als LETZTEN
// Eintrag an X-Forwarded-For anhängt. Der ganze Header ist als Limiter-Key
// ungeeignet: Ein Client kann beliebige eigene Einträge voranstellen und so für
// jeden Request einen frischen Key erzeugen — nur der letzte (vom eigenen Proxy
// gesetzte) Eintrag ist vertrauenswürdig. Ohne Header zählt RemoteAddr.
func clientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff == "" {
		return r.RemoteAddr
	}
	parts := strings.Split(xff, ",")
	return strings.TrimSpace(parts[len(parts)-1])
}

// PostMethodOnlyMiddleware middleware ensures the request method is POST
func PostMethodOnlyMiddleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := zerolog.Ctx(r.Context())

		// Ops exception: /health must be probeable via GET for container orchestrators.
		if r.Method == http.MethodGet && r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

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

// UserGetter loads a user by ID; the JWT middleware uses it to verify that the
// account behind a valid token is still active.
type UserGetter interface {
	GetUser(ctx context.Context, id int) (user.User, error)
}

// NewJwtMiddleware validates the JWT Token in the Authorization header.
// If valid, it verifies the user is still active (deactivated users lose
// access immediately, not at token expiry) and adds the user information
// to the request context.
func NewJwtMiddleware(jwtSecret string, allowedRoles []string, users UserGetter) func(http.Handler) http.HandlerFunc {
	return func(h http.Handler) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := zerolog.Ctx(r.Context())

			token := r.Header.Get("Authorization")
			if token == "" {
				logger.Error().Msg("Missing Authorization header")
				helper.SendClientError(w, "missing_authorization", nil)
				return
			}

			const bearerPrefix = "Bearer "
			if len(token) <= len(bearerPrefix) || token[:len(bearerPrefix)] != bearerPrefix {
				logger.Error().Msg("Invalid Authorization header format")
				helper.SendClientError(w, "invalid_authorization_format", nil)
				return
			}
			token = token[len(bearerPrefix):]
			userID, userName, userRole, err := jwt.ParseAndValidateJWTToken(token, jwtSecret)
			if err != nil {
				logger.Error().Err(err).Msg("Invalid JWT token")
				helper.SendClientError(w, "invalid_jwt", nil)
				return
			}

			roleAllowed := slices.Contains(allowedRoles, userRole)
			if !roleAllowed {
				logger.Warn().Str("role", userRole).Msg("Insufficient permissions")
				helper.SendClientError(w, "insufficient_permissions", fmt.Sprintf("Insufficient permissions for role %s", userRole))
				return
			}

			u, err := users.GetUser(r.Context(), userID)
			if err != nil {
				if errors.Is(err, db.ErrNotFound) {
					logger.Warn().Int("user_id", userID).Msg("User from token no longer exists")
					helper.SendUnauthorized(w, "user_inactive")
					return
				}
				logger.Error().Err(err).Int("user_id", userID).Msg("Failed to load user for active-status check")
				helper.SendServerError(w)
				return
			}
			if u.Status != user.ActiveStatus {
				logger.Warn().Int("user_id", userID).Str("status", string(u.Status)).Msg("User is not active")
				helper.SendUnauthorized(w, "user_inactive")
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, UserIDKey, userID)
			ctx = context.WithValue(ctx, UserNameKey, userName)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
