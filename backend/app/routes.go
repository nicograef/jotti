package app

import (
	"database/sql"
	"net/http"

	"github.com/nicograef/jotti/backend/api"
	"github.com/nicograef/jotti/backend/api/middleware"
	testapi "github.com/nicograef/jotti/backend/api/test"
	"github.com/nicograef/jotti/backend/config"
)

// Area beschreibt einen zusammenhängenden Routen-Bereich als einzige Quelle
// seiner Zugriffsregeln. Aus dieser Tabelle registriert SetupRoutes alle Routen;
// die Berechtigungs-Matrix (matrix_integration_test.go) liest dieselbe Tabelle.
// Weil jeder Bereich seine erlaubten Rollen (oder bewusst kein JWT) deklarieren
// MUSS, kann keine Route ohne Rollenentscheidung existieren.
type Area struct {
	// Name ist ein sprechender Bezeichner für Logs und Tests.
	Name string
	// Prefix ist das URL-Präfix des Bereichs (z. B. "/admin"). Beim Mounten
	// wird es per http.StripPrefix entfernt, die Bereichs-Handler sehen den
	// Restpfad (z. B. "/create-user").
	Prefix string
	// AllowedRoles sind die Rollen, die diesen Bereich aufrufen dürfen. Leer nur
	// bei RequiresAuth == false (öffentliche Bereiche: auth, relay).
	AllowedRoles []string
	// RequiresAuth == true ⇒ Bereich wird mit der JWT-Middleware geschützt.
	// false ⇒ bewusst ohne JWT (health/auth/relay); die Prüfung liegt dann im
	// Handler (Relay-Token) bzw. entfällt (Login).
	RequiresAuth bool
	// RateLimited == true ⇒ zusätzlich IP-Rate-Limit (Login/Relay gegen
	// Brute-Force). Der Wert 5 bildet das bisherige Verhalten ab.
	RateLimited bool
	// build konstruiert den Bereichs-Handler und liefert dessen registrierte
	// Pfade zurück; die Pfade sind die Zeilen der Berechtigungs-Matrix.
	build func(cfg config.Config, deps api.Deps) (http.Handler, []string)
}

// Rollen-Mengen als Konstanten, damit Tabelle und Matrix-Test denselben Bezug
// haben (bislang als String-Literale in SetupRoutes verstreut).
var (
	rolesAdmin          = []string{"admin"}
	rolesService        = []string{"admin", "serviceleitung", "service"}
	rolesServiceleitung = []string{"admin", "serviceleitung"}
)

// Areas ist die deklarative Routentabelle — die einzige Registrierungsquelle.
// Verhalten identisch zur früheren imperativen Registrierung in SetupRoutes.
func Areas() []Area {
	return []Area{
		{
			Name:         "auth",
			Prefix:       "/auth",
			RequiresAuth: false,
			RateLimited:  true,
			build:        api.NewAuthApi,
		},
		{
			Name:         "admin",
			Prefix:       "/admin",
			AllowedRoles: rolesAdmin,
			RequiresAuth: true,
			build: func(_ config.Config, deps api.Deps) (http.Handler, []string) {
				return api.NewAdminApi(deps)
			},
		},
		{
			Name:         "service",
			Prefix:       "/service",
			AllowedRoles: rolesService,
			RequiresAuth: true,
			build: func(_ config.Config, deps api.Deps) (http.Handler, []string) {
				return api.NewServiceApi(deps)
			},
		},
		{
			Name:         "serviceleitung",
			Prefix:       "/serviceleitung",
			AllowedRoles: rolesServiceleitung,
			RequiresAuth: true,
			build: func(_ config.Config, deps api.Deps) (http.Handler, []string) {
				return api.NewServiceleitungApi(deps)
			},
		},
		{
			Name:         "relay",
			Prefix:       "/relay",
			RequiresAuth: false,
			RateLimited:  true,
			build: func(cfg config.Config, deps api.Deps) (http.Handler, []string) {
				return api.NewRelayApi(deps, cfg.RelayToken)
			},
		},
	}
}

// mountArea registriert einen Bereich am Router gemäß seiner Deklaration:
// JWT-Middleware (falls RequiresAuth), Rate-Limit (falls RateLimited),
// Prefix-Strip. Gibt die absoluten Pfade des Bereichs zurück.
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

// testResetArea liefert den Test-Bereich (POST /test/reset-and-seed) als
// deklarative Area. Er wird nur bei JOTTI_ENABLE_TEST_API=1 an Areas angehängt und
// läuft — wie auth/relay — bewusst ohne JWT: der Endpunkt setzt die Datenbank
// auf den Test-Zustand zurück und ist ausschließlich in der E2E-Umgebung
// registriert. Er ist wie auth/relay rate-limitet, damit ein voller Truncate +
// Reseed nicht als DoS-Vektor missbraucht werden kann, und wird über dieselbe
// mountArea-Verdrahtung gemountet.
func testResetArea(db *sql.DB) Area {
	return Area{
		Name:         "test",
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
