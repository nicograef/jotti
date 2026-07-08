// Package test stellt Testhilfen als HTTP-Endpunkte bereit, die ausschließlich
// in Test- und Demo-Umgebungen registriert werden dürfen (JOTTI_ALLOW_SEED=1).
// Der einzige Endpunkt, POST /test/reset-and-seed, setzt die Datenbank auf den
// deterministischen Demo-Zustand zurück und gibt der aufrufenden E2E-Suite die
// benötigten Zugangsdaten zurück. In Produktion wird die Route nie registriert.
package test

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/seed"
	"github.com/rs/zerolog"
)

// Zugangsdaten beschreibt einen Seed-Benutzer, mit dem sich die Suite anmeldet.
type Zugangsdaten struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ResetResponse ist die Antwort von POST /test/reset-and-seed: die
// deterministischen Zugangsdaten der wichtigsten Rollen aus dem Seed-Zustand.
type ResetResponse struct {
	Admin          Zugangsdaten `json:"admin"`
	Serviceleitung Zugangsdaten `json:"serviceleitung"`
	Service        Zugangsdaten `json:"service"`
}

// reseter kapselt das Zurücksetzen und Neu-Seeden der Datenbank, damit der
// Handler in Unit-Tests ohne echte Datenbank geprüft werden kann.
type reseter interface {
	ResetAndSeed(ctx context.Context) error
}

// dbReseter setzt über die Seed-Engine auf einer echten Datenbank zurück.
type dbReseter struct {
	db *sql.DB
}

func (r dbReseter) ResetAndSeed(ctx context.Context) error {
	return seed.ResetAndSeed(ctx, r.db)
}

// Handler bündelt die Abhängigkeiten des Test-Endpunkts.
type Handler struct {
	Reseter reseter
}

// NewHandler konstruiert den Test-Handler für die echte Datenbank.
func NewHandler(db *sql.DB) Handler {
	return Handler{Reseter: dbReseter{db: db}}
}

// ResetAndSeedHandler leert die Datenbank und seedet den Demo-Zustand neu, dann
// liefert er die Zugangsdaten der Seed-Rollen zurück.
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
