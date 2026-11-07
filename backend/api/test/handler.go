// Package test serves test helpers as HTTP endpoints, registered only in the E2E environment (JOTTI_ENABLE_TEST_API=1).
// POST /test/reset-and-seed resets the database to the deterministic demo state and returns the credentials.
package test

import (
	"database/sql"
	"net/http"

	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/internal/logging"
	"github.com/nicograef/jotti/backend/seed"
)

type zugangsdaten struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type resetResponse struct {
	Admin          zugangsdaten `json:"admin"`
	Serviceleitung zugangsdaten `json:"serviceleitung"`
	Service        zugangsdaten `json:"service"`
}

type Handler struct {
	DB *sql.DB
}

func NewHandler(db *sql.DB) Handler {
	return Handler{DB: db}
}

func (h Handler) ResetAndSeedHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logging.FromContext(r.Context())

		if err := seed.ResetAndSeed(r.Context(), h.DB); err != nil {
			log.Error("reset-and-seed failed", "error", err)
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, resetResponse{
			Admin:          zugangsdaten{Username: seed.DemoAdminUsername, Password: seed.DemoPassword},
			Serviceleitung: zugangsdaten{Username: seed.DemoServiceleitungUsername, Password: seed.DemoPassword},
			Service:        zugangsdaten{Username: seed.DemoServiceUsername, Password: seed.DemoPassword},
		})
	}
}
