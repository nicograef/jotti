package api

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type database interface {
	PingContext(ctx context.Context) error
}

// HealthCheck provides health check functionality with database connectivity testing.
type HealthCheck struct {
	DB database
}

// HealthResponse represents the health check response structure.
type HealthResponse struct {
	Status    string `json:"status"`
	Database  string `json:"database"`
	Timestamp string `json:"timestamp"`
}

// Handler returns an HTTP handler for the enhanced health check endpoint with database ping.
func (h *HealthCheck) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := zerolog.Ctx(r.Context())

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		dbStatus := "ok"
		overallStatus := "ok"
		statusCode := http.StatusOK

		if err := h.DB.PingContext(ctx); err != nil {
			dbStatus = "error"
			overallStatus = "degraded"
			statusCode = http.StatusServiceUnavailable
			log.Error().Err(err).Msg("Database ping failed")
		}

		response := HealthResponse{
			Status:    overallStatus,
			Database:  dbStatus,
			Timestamp: time.Now().Format(time.RFC3339),
		}

		sendJSONResponse(w, response, statusCode)
	}
}
