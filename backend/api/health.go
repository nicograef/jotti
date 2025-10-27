package api

import (
	"net/http"
)

func NewHealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !validateMethod(w, r, http.MethodGet) {
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}
