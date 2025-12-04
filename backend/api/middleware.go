package api

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
)

type contextKey string

const CorrelationIDKey contextKey = "correlation_id"

// CorrelationIDMiddleware adds a correlation ID to each request for tracing
func CorrelationIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if correlation ID already exists in header
		correlationID := r.Header.Get("X-Correlation-ID")
		if correlationID == "" {
			// Generate new correlation ID
			correlationID = uuid.New().String()
		}

		// Add to response header
		w.Header().Set("X-Correlation-ID", correlationID)

		// Add to context
		ctx := context.WithValue(r.Context(), CorrelationIDKey, correlationID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoggingMiddleware logs HTTP requests with correlation ID
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Extract correlation ID from context
		correlationID, _ := r.Context().Value(CorrelationIDKey).(string)

		// Create a response writer wrapper to capture status code
		ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(ww, r)

		log.Info().
			Str("correlation_id", correlationID).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", ww.statusCode).
			Dur("duration_ms", time.Since(start)).
			Msg("HTTP request")
	})
}

// RateLimitMiddleware limits requests per IP
func RateLimitMiddleware(requestsPerSecond int) func(http.Handler) http.Handler {
	limiter := rate.NewLimiter(rate.Limit(requestsPerSecond), requestsPerSecond*2)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				log.Warn().Str("path", r.URL.Path).Msg("Rate limit exceeded")
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
		if r.Method != http.MethodPost {
			log.Error().
				Str("method", r.Method).
				Str("expected", http.MethodPost).
				Msg("Invalid method")
			SendMethodNotAllowedError(w, ErrorResponse{
				Message: "Method not allowed",
				Code:    "method_not_allowed",
			})
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
