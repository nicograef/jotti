package middleware

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"runtime/debug"
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

// UserFromContext reads what NewJwtMiddleware put into the request context.
func UserFromContext(ctx context.Context) (userID int, userName string, ok bool) {
	userID, ok = ctx.Value(UserIDKey).(int)
	if !ok {
		return 0, "", false
	}
	userName, _ = ctx.Value(UserNameKey).(string)
	return userID, userName, true
}

func CorrelationIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := r.Header.Get("X-Correlation-ID")
		if correlationID == "" {
			correlationID = uuid.NewString()[:8]
		}

		w.Header().Set("X-Correlation-ID", correlationID)

		ctx := context.WithValue(r.Context(), CorrelationIDKey, correlationID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

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

// RecoveryMiddleware beendet einen Handler-Panic mit 500 im bestehenden
// Fehler-Response-Format statt mit einer abgerissenen Verbindung (net/http würde
// nur schließen). http.ErrAbortHandler wird durchgereicht — das ist net/https
// idiomatisches Signal, eine Response bewusst abzubrechen.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}
			if err, ok := rec.(error); ok && errors.Is(err, http.ErrAbortHandler) {
				panic(rec)
			}
			zerolog.Ctx(r.Context()).Error().
				Interface("panic", rec).
				Bytes("stack", debug.Stack()).
				Str("path", r.URL.Path).
				Msg("Panic in HTTP handler")
			helper.SendServerError(w)
		}()

		next.ServeHTTP(w, r)
	})
}

type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func RateLimitMiddleware(requestsPerSecond int) func(http.Handler) http.Handler {
	var mu sync.Mutex
	limiters := make(map[string]*limiterEntry)

	// A panic here must not tear down the process; the loop continues at the next
	// interval.
	cleanup := func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error().Interface("panic", r).Bytes("stack", debug.Stack()).Msg("Rate-Limiter-Cleanup: Panic abgefangen; Loop laeuft weiter")
			}
		}()
		mu.Lock()
		defer mu.Unlock()
		for ip, entry := range limiters {
			if time.Since(entry.lastSeen) > 10*time.Minute {
				delete(limiters, ip)
			}
		}
	}
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			cleanup()
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
				helper.SendTooManyRequests(w, "rate_limited")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// clientIP liefert den Limiter-Key. jotti läuft hinter dem eigenen
// Reverse-Proxy (Caddy), der die echte Client-IP als LETZTEN X-Forwarded-For-
// Eintrag anhängt; nur dieser ist vertrauenswürdig — ein Client kann eigene
// Einträge voranstellen und sich so je Request einen frischen Key erzeugen.
// Ohne Header zählt RemoteAddr.
func clientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff == "" {
		// RemoteAddr ist "IP:Port": Der ephemere Port wechselt je Verbindung und
		// gehört NICHT in den Limiter-Key — sonst greift das Limit nie und die Map
		// wächst unbegrenzt.
		if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
			return host
		}
		return r.RemoteAddr
	}
	parts := strings.Split(xff, ",")
	return strings.TrimSpace(parts[len(parts)-1])
}

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

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Unwrap gibt den umschlossenen ResponseWriter frei: http.ResponseController
// sucht genau diese Methode, um an SetWriteDeadline, SetReadDeadline und Flush
// des echten Writers zu kommen. Ohne Unwrap scheitert hinter dieser Middleware
// jeder Controller-Aufruf mit "feature not supported" — und LoggingMiddleware
// umschließt die gesamte Routenkette (backend/app/app.go).
func (rw *responseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

type UserGetter interface {
	GetUser(ctx context.Context, id int) (user.User, error)
}

// NewJwtMiddleware validates the JWT and checks status and role against a fresh
// database record: deactivated users lose access immediately, role changes take
// effect on the next request instead of at token expiry. Authentication failures
// yield 401, insufficient privileges 403.
func NewJwtMiddleware(jwtSecret string, allowedRoles []string, users UserGetter) func(http.Handler) http.HandlerFunc {
	return func(h http.Handler) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := zerolog.Ctx(r.Context())

			token := r.Header.Get("Authorization")
			if token == "" {
				logger.Error().Msg("Missing Authorization header")
				helper.SendUnauthorized(w, "missing_authorization")
				return
			}

			const bearerPrefix = "Bearer "
			if len(token) <= len(bearerPrefix) || token[:len(bearerPrefix)] != bearerPrefix {
				logger.Error().Msg("Invalid Authorization header format")
				helper.SendUnauthorized(w, "invalid_authorization_format")
				return
			}
			token = token[len(bearerPrefix):]
			userID, _, err := jwt.ParseAndValidateJWTToken(token, jwtSecret)
			if err != nil {
				logger.Error().Err(err).Msg("Invalid JWT token")
				helper.SendUnauthorized(w, "invalid_jwt")
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

			if !slices.Contains(allowedRoles, string(u.Role)) {
				logger.Warn().Str("role", string(u.Role)).Msg("Insufficient permissions")
				helper.SendForbidden(w, "insufficient_permissions", fmt.Sprintf("role %s is not allowed for this endpoint", u.Role))
				return
			}

			// Der Name im Context stammt aus dem Datensatz, nicht aus dem Token-Claim:
			// ein umbenannter Benutzer erscheint im Kassenjournal unter dem heutigen Namen.
			ctx := r.Context()
			ctx = context.WithValue(ctx, UserIDKey, userID)
			ctx = context.WithValue(ctx, UserNameKey, u.Username)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
