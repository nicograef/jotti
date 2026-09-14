// Package test stellt Testhilfen als HTTP-Endpunkte bereit; die Routen werden
// ausschließlich in der E2E-Umgebung registriert (JOTTI_ENABLE_TEST_API=1), nie
// in Produktion. POST /test/reset-and-seed setzt die Datenbank auf den
// deterministischen Demo-Zustand zurück und liefert die Zugangsdaten.
package test

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/seed"
	"github.com/rs/zerolog"
)

type Zugangsdaten struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ResetResponse struct {
	Admin          Zugangsdaten `json:"admin"`
	Serviceleitung Zugangsdaten `json:"serviceleitung"`
	Service        Zugangsdaten `json:"service"`
}

type reseter interface {
	ResetAndSeed(ctx context.Context) error
}

type dbReseter struct {
	db *sql.DB
}

func (r dbReseter) ResetAndSeed(ctx context.Context) error {
	return seed.ResetAndSeed(ctx, r.db)
}

type Handler struct {
	Reseter reseter
}

func NewHandler(db *sql.DB) Handler {
	return Handler{Reseter: dbReseter{db: db}}
}

func (h Handler) ResetAndSeedHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := zerolog.Ctx(r.Context())

		if err := h.Reseter.ResetAndSeed(r.Context()); err != nil {
			log.Error().Err(err).Msg("reset-and-seed failed")
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, ResetResponse{
			Admin:          Zugangsdaten{Username: seed.DemoAdminUsername, Password: seed.DemoPassword},
			Serviceleitung: Zugangsdaten{Username: seed.DemoServiceleitungUsername, Password: seed.DemoPassword},
			Service:        Zugangsdaten{Username: seed.DemoServiceUsername, Password: seed.DemoPassword},
		})
	}
}
