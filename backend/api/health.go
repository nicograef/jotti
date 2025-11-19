package api

import (
	"net/http"
)

// NewHealthHandler returns an HTTP handler for the health check endpoint.
func NewHealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !ValidateMethod(w, r, http.MethodGet) {
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
