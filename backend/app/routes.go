package app

import (
	"database/sql"
	"net/http"

	"github.com/nicograef/jotti/backend/api"
	"github.com/nicograef/jotti/backend/api/middleware"
	testapi "github.com/nicograef/jotti/backend/api/test"
	"github.com/nicograef/jotti/backend/config"
)

// Area deklariert die Zugriffsregeln eines Routen-Bereichs. SetupRoutes
// registriert daraus alle Routen; die Berechtigungs-Matrix
// (matrix_integration_test.go) liest dieselbe Tabelle. Jeder Bereich muss Rollen
// oder bewusst kein JWT deklarieren — keine Route ohne Rollenentscheidung.
type Area struct {
	// Prefix ist das URL-Präfix (z. B. "/admin"); beim Mounten per StripPrefix entfernt.
	Prefix string
	// AllowedRoles ist leer nur bei RequiresAuth == false (auth, relay).
	AllowedRoles []string
	// RequiresAuth false ⇒ bewusst ohne JWT (auth/relay); die Prüfung liegt dann
	// im Handler (Relay-Token) bzw. entfällt (Login).
	RequiresAuth bool
	// RateLimited ⇒ zusätzlich IP-Rate-Limit (Login/Relay gegen Brute-Force, siehe mountArea).
	RateLimited bool
	// build liefert den Bereichs-Handler und seine Pfade — die Zeilen der Berechtigungs-Matrix.
	build func(cfg config.Config, deps api.Deps) (http.Handler, []string)
}

var (
	rolesAdmin          = []string{"admin"}
	rolesService        = []string{"admin", "serviceleitung", "service"}
	rolesServiceleitung = []string{"admin", "serviceleitung"}
)

func Areas() []Area {
	return []Area{
		{
			Prefix:       "/auth",
			RequiresAuth: false,
			RateLimited:  true,
			build:        api.NewAuthApi,
		},
		{
			Prefix:       "/admin",
			AllowedRoles: rolesAdmin,
			RequiresAuth: true,
			build: func(_ config.Config, deps api.Deps) (http.Handler, []string) {
				return api.NewAdminApi(deps)
			},
		},
		{
			Prefix:       "/service",
			AllowedRoles: rolesService,
			RequiresAuth: true,
			build: func(_ config.Config, deps api.Deps) (http.Handler, []string) {
				return api.NewServiceApi(deps)
			},
		},
		{
			Prefix:       "/serviceleitung",
			AllowedRoles: rolesServiceleitung,
			RequiresAuth: true,
			build: func(_ config.Config, deps api.Deps) (http.Handler, []string) {
				return api.NewServiceleitungApi(deps)
			},
		},
		{
			Prefix:       "/relay",
			RequiresAuth: false,
			RateLimited:  true,
			build: func(cfg config.Config, deps api.Deps) (http.Handler, []string) {
				return api.NewRelayApi(deps, cfg.RelayToken)
			},
		},
	}
}

func mountArea(r *http.ServeMux, area Area, cfg config.Config, deps api.Deps) {
	handler, _ := area.build(cfg, deps)

	if area.RequiresAuth {
		// Der Benutzer-Lookup pro Request stellt sicher, dass deaktivierte
		// Benutzer sofort ausgesperrt sind, nicht erst beim Token-Ablauf.
		jwt := middleware.NewJwtMiddleware(cfg.JWTSecret, area.AllowedRoles, deps.UserRepo)
		handler = jwt(http.StripPrefix(area.Prefix, handler))
	} else {
		handler = http.StripPrefix(area.Prefix, handler)
	}

	if area.RateLimited {
		handler = middleware.RateLimitMiddleware(5)(handler)
	}

	r.Handle(area.Prefix+"/", handler)
}

// testResetArea ist der Test-Bereich (POST /test/reset-and-seed): nur bei
// JOTTI_ENABLE_TEST_API=1 an Areas angehängt, bewusst ohne JWT wie auth/relay
// und rate-limitet, damit Truncate + Reseed kein DoS-Vektor wird.
func testResetArea(db *sql.DB) Area {
	return Area{
		Prefix:       "/test",
		RequiresAuth: false,
		RateLimited:  true,
		build: func(_ config.Config, _ api.Deps) (http.Handler, []string) {
			handler := testapi.NewHandler(db)
			mux := http.NewServeMux()
			mux.HandleFunc("/reset-and-seed", handler.ResetAndSeedHandler())
			return mux, []string{"/reset-and-seed"}
		},
	}
}
