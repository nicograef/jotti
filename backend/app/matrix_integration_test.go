//go:build integration

package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nicograef/jotti/backend/api"
	"github.com/nicograef/jotti/backend/config"
	dbpkg "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/jwt"
	"github.com/nicograef/jotti/backend/domain/user"
	"github.com/nicograef/jotti/backend/repository/user_repo"
)

const jwtSecret = "matrix-test-jwt-secret"

var alleRollen = []user.Role{user.AdminRole, user.ServiceleitungRole, user.ServiceRole}

func setupMatrix(t *testing.T) (http.Handler, map[user.Role]string, func()) {
	t.Helper()
	db := dbpkg.OpenTestDatabase()

	if _, err := db.Exec("DELETE FROM users"); err != nil {
		t.Fatalf("users bereinigen: %v", err)
	}

	repo := user_repo.NewRepository(db)
	tokens := make(map[user.Role]string, len(alleRollen))
	for _, role := range alleRollen {
		u, _, err := user.NewUser("Matrix "+string(role), "matrix"+strings.ReplaceAll(string(role), "-", ""), role)
		if err != nil {
			t.Fatalf("NewUser(%s): %v", role, err)
		}
		u.Status = user.ActiveStatus
		id, err := repo.CreateUser(context.Background(), u)
		if err != nil {
			t.Fatalf("CreateUser(%s): %v", role, err)
		}
		token, err := jwt.GenerateJWTTokenForUser(id, string(role), jwtSecret)
		if err != nil {
			t.Fatalf("Token(%s): %v", role, err)
		}
		tokens[role] = token
	}

	handler := SetupRoutes(testConfig(), db, "test")

	return handler, tokens, func() {
		_, _ = db.Exec("DELETE FROM users")
		_ = db.Close()
	}
}

// testConfig umgeht config.Load, damit der Test nicht von Umgebungsvariablen abhängt.
func testConfig() config.Config {
	return config.Config{
		Port:       3000,
		JWTSecret:  jwtSecret,
		RelayToken: "matrix-test-relay-token",
	}
}

func doRequest(t *testing.T, handler http.Handler, path, token string) (int, string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	code := ""
	var body struct {
		Code string `json:"code"`
	}
	if json.NewDecoder(w.Body).Decode(&body) == nil {
		code = body.Code
	}
	return w.Code, code
}

// TestBerechtigungsMatrix prüft jede geschützte Route aus der Areas-Tabelle gegen
// jede Rolle sowie "kein Token" und "ungültiger Token". Bei erlaubter Rolle wird
// nur auf kein 401/403 geprüft — der leere Body erzeugt danach oft einen
// fachlichen Fehler.
func TestBerechtigungsMatrix(t *testing.T) {
	handler, tokens, teardown := setupMatrix(t)
	defer teardown()

	for _, area := range Areas() {
		if !area.RequiresAuth {
			continue // auth/relay: kein JWT — eigener Test unten
		}
		_, paths := area.build(testConfig(), api.Deps{})
		allowed := make(map[user.Role]bool, len(area.AllowedRoles))
		for _, r := range area.AllowedRoles {
			allowed[user.Role(r)] = true
		}

		for _, p := range paths {
			fullPath := area.Prefix + p

			if code, ec := doRequest(t, handler, fullPath, ""); code != http.StatusUnauthorized {
				t.Errorf("%s ohne Token: Status %d (%s), erwartet 401", fullPath, code, ec)
			}

			if code, ec := doRequest(t, handler, fullPath, "kaputt.token.wert"); code != http.StatusUnauthorized {
				t.Errorf("%s mit ungültigem Token: Status %d (%s), erwartet 401", fullPath, code, ec)
			}

			for _, role := range alleRollen {
				code, ec := doRequest(t, handler, fullPath, tokens[role])
				if allowed[role] {
					if code == http.StatusUnauthorized || code == http.StatusForbidden {
						t.Errorf("%s als %s (erlaubt): Status %d (%s), darf nicht 401/403 sein", fullPath, role, code, ec)
					}
				} else {
					if code != http.StatusForbidden {
						t.Errorf("%s als %s (verboten): Status %d (%s), erwartet 403", fullPath, role, code, ec)
					}
				}
			}
		}
	}
}

// Die JWT-freien Bereiche (auth, relay) dürfen ohne Token nicht mit 401 der
// JWT-Middleware antworten, sondern reichen bis zum Handler durch (eigener
// Fehlercode aus Body-/Token-Validierung).
func TestBerechtigungsMatrix_OeffentlicheBereiche(t *testing.T) {
	handler, _, teardown := setupMatrix(t)
	defer teardown()

	for _, area := range Areas() {
		if area.RequiresAuth {
			continue
		}
		_, paths := area.build(testConfig(), api.Deps{})
		for _, p := range paths {
			fullPath := area.Prefix + p
			code, ec := doRequest(t, handler, fullPath, "")
			if code == http.StatusUnauthorized && (ec == "missing_authorization" || ec == "invalid_jwt") {
				t.Errorf("%s (öffentlich) darf nicht von der JWT-Middleware mit %q abgewiesen werden", fullPath, ec)
			}
		}
	}
}

// Prüft die Deklaration des bedingten Test-Reset-Bereichs (RequiresAuth == false
// ⇒ keine JWT-Middleware) samt Pfad. Der Endpunkt wird bewusst NICHT aufgerufen:
// ResetAndSeed würde die von setupMatrix geteilte Datenbank neu seeden und die
// folgenden Tests stören. Die Env-Registrierung deckt
// TestSetupRoutes_ResetSeedRouteGuardedByEnv ab.
func TestBerechtigungsMatrix_TestResetOeffentlich(t *testing.T) {
	area := testResetArea(nil)

	if area.RequiresAuth {
		t.Fatalf("Test-Reset-Bereich muss ohne JWT laufen (RequiresAuth == false), ist aber true")
	}

	_, paths := area.build(testConfig(), api.Deps{})
	if len(paths) != 1 || paths[0] != "/reset-and-seed" {
		t.Fatalf("Test-Reset-Bereich muss genau /reset-and-seed exponieren, hat aber %v", paths)
	}
	if area.Prefix != "/test" {
		t.Fatalf("Test-Reset-Bereich muss Präfix /test haben, hat aber %q", area.Prefix)
	}
}

// Die Service-Rolle darf die Serviceleitungs-Route nicht aufrufen (403), obwohl
// sie im benachbarten Service-Bereich privilegiert ist. Ein-Mandanten-System —
// die fachliche Abgrenzung ist die Rolle.
func TestBerechtigungsMatrix_Objektbezug(t *testing.T) {
	handler, tokens, teardown := setupMatrix(t)
	defer teardown()

	code, ec := doRequest(t, handler, "/serviceleitung/stornierung-erteilen", tokens[user.ServiceRole])
	if code != http.StatusForbidden {
		t.Fatalf("Service-Rolle auf Serviceleitungs-Storno: Status %d (%s), erwartet 403", code, ec)
	}

	code, ec = doRequest(t, handler, "/serviceleitung/stornierung-erteilen", tokens[user.ServiceleitungRole])
	if code == http.StatusForbidden || code == http.StatusUnauthorized {
		t.Fatalf("Serviceleitung auf Serviceleitungs-Storno: Status %d (%s), darf nicht 401/403 sein", code, ec)
	}
}

// /auth läuft über RateLimitMiddleware(5), Burst 10 — der 11. schnelle Request
// derselben IP bekommt 429. Der Test geht durch den echten Router, damit die
// Verdrahtung aus der Areas-Tabelle mitgeprüft wird.
func TestLoginRateLimit(t *testing.T) {
	handler, _, teardown := setupMatrix(t)
	defer teardown()

	// httptest setzt eine feste RemoteAddr: alle Requests teilen denselben Limiter-Key.
	var gotTooMany bool
	for i := 0; i < 20; i++ {
		req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader("{}"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code == http.StatusTooManyRequests {
			gotTooMany = true
			break
		}
	}

	if !gotTooMany {
		t.Fatal("Login-Rate-Limit hat innerhalb von 20 schnellen Requests kein 429 geliefert")
	}
}
