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

// jwtSecret ist das Signaturgeheimnis, mit dem Handler und Test-Tokens gebaut werden.
const jwtSecret = "matrix-test-jwt-secret"

// alleRollen sind die drei Rollen des Systems; die Matrix testet jede Route
// gegen jede davon (plus die Fälle "kein Token" und "ungültiger Token").
var alleRollen = []user.Role{user.AdminRole, user.ServiceleitungRole, user.ServiceRole}

// testUser bündelt eine angelegte Rolle mit ihrem gültigen Bearer-Token.
type testUser struct {
	role  user.Role
	token string
}

// setupMatrix legt für jede Rolle einen aktiven Benutzer an und baut den echten
// Router (SetupRoutes) mit einer festen JWT-Secret-Config auf. Rückgabe: Handler,
// Rolle→Token-Map, Teardown.
func setupMatrix(t *testing.T) (http.Handler, map[user.Role]testUser, func()) {
	t.Helper()
	db := dbpkg.OpenTestDatabase()

	if _, err := db.Exec("DELETE FROM users"); err != nil {
		t.Fatalf("users bereinigen: %v", err)
	}

	repo := user_repo.NewRepository(db)
	users := make(map[user.Role]testUser, len(alleRollen))
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
		token, err := jwt.GenerateJWTTokenForUser(id, u.Username, string(role), jwtSecret)
		if err != nil {
			t.Fatalf("Token(%s): %v", role, err)
		}
		users[role] = testUser{role: role, token: token}
	}

	handler := SetupRoutes(testConfig(), db, "test")

	return handler, users, func() {
		_, _ = db.Exec("DELETE FROM users")
		_ = db.Close()
	}
}

// testConfig baut eine minimale Config mit dem Test-JWT-Secret. Sie umgeht
// config.Load, damit der Test nicht von Umgebungsvariablen abhängt.
func testConfig() config.Config {
	return config.Config{
		Port:       3000,
		JWTSecret:  jwtSecret,
		RelayToken: "matrix-test-relay-token",
	}
}

// doRequest schickt einen POST an path mit optionalem Bearer-Token und gibt den
// Statuscode und den Fehlercode (falls JSON mit "code") zurück.
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

// TestBerechtigungsMatrix prüft für JEDE geschützte Route (aus der deklarativen
// Areas-Tabelle) JEDE Rolle sowie die Fälle "kein Token" und "ungültiger Token":
//   - erlaubte Rolle  ⇒ weder 401 noch 403 (Autorisierung bestanden; danach ggf.
//     fachlicher Fehler wegen leerem Body)
//   - verbotene Rolle ⇒ 403 insufficient_permissions
//   - kein Token      ⇒ 401 missing_authorization
//   - ungültiger Token⇒ 401 invalid_jwt
func TestBerechtigungsMatrix(t *testing.T) {
	handler, users, teardown := setupMatrix(t)
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

			// Kein Token ⇒ 401.
			if code, ec := doRequest(t, handler, fullPath, ""); code != http.StatusUnauthorized {
				t.Errorf("%s ohne Token: Status %d (%s), erwartet 401", fullPath, code, ec)
			}

			// Ungültiger Token ⇒ 401.
			if code, ec := doRequest(t, handler, fullPath, "kaputt.token.wert"); code != http.StatusUnauthorized {
				t.Errorf("%s mit ungültigem Token: Status %d (%s), erwartet 401", fullPath, code, ec)
			}

			// Jede Rolle.
			for _, role := range alleRollen {
				code, ec := doRequest(t, handler, fullPath, users[role].token)
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

// TestBerechtigungsMatrix_OeffentlicheBereiche prüft, dass die JWT-freien
// Bereiche (auth, relay) ohne Token NICHT mit 401 der JWT-Middleware antworten,
// sondern den Request bis zum Handler durchreichen (dort greift Body-/Token-
// Validierung mit eigenem Fehlercode).
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

// TestBerechtigungsMatrix_TestResetOeffentlich behandelt den Sonderfall des
// bedingten Test-Reset-Bereichs (POST /test/reset-and-seed) explizit: er wird
// nur bei JOTTI_ENABLE_TEST_API=1 registriert und läuft — wie auth/relay — bewusst
// ohne JWT. Der Test baut den Bereich über dieselbe Fabrik wie SetupRoutes und
// prüft die Deklaration (RequiresAuth == false ⇒ keine JWT-Middleware im
// mountArea-Pfad) samt Pfad. Der Endpunkt selbst wird bewusst NICHT aufgerufen:
// ResetAndSeed würde die von setupMatrix geteilte Datenbank neu seeden und die
// nachfolgenden Tests stören. Die Env-Registrierung (404 ohne Flag, nicht 404
// mit Flag) deckt der Unit-Test TestSetupRoutes_ResetSeedRouteGuardedByEnv ab.
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

// TestBerechtigungsMatrix_Objektbezug deckt den bereichsübergreifenden
// Objektzugriff explizit ab: die Service-Rolle darf die Serviceleitungs-Route
// /serviceleitung/stornierung-erteilen nicht aufrufen (403), obwohl sie im
// benachbarten Service-Bereich privilegiert ist. Ein-Mandanten-System: es gibt
// keine mandantenfremden Objekte, die fachliche Abgrenzung ist die Rolle.
func TestBerechtigungsMatrix_Objektbezug(t *testing.T) {
	handler, users, teardown := setupMatrix(t)
	defer teardown()

	code, ec := doRequest(t, handler, "/serviceleitung/stornierung-erteilen", users[user.ServiceRole].token)
	if code != http.StatusForbidden {
		t.Fatalf("Service-Rolle auf Serviceleitungs-Storno: Status %d (%s), erwartet 403", code, ec)
	}

	// Serviceleitung darf; Autorisierung muss passieren (kein 401/403).
	code, ec = doRequest(t, handler, "/serviceleitung/stornierung-erteilen", users[user.ServiceleitungRole].token)
	if code == http.StatusForbidden || code == http.StatusUnauthorized {
		t.Fatalf("Serviceleitung auf Serviceleitungs-Storno: Status %d (%s), darf nicht 401/403 sein", code, ec)
	}
}

// TestLoginRateLimit prüft, dass der /auth-Bereich per RateLimitMiddleware(5)
// gedrosselt wird: Nach dem Burst (5 r/s ⇒ Burst 10) liefert der Router für
// weitere schnelle Requests derselben IP 429. Der Test geht durch den echten
// Router (SetupRoutes), damit die Verdrahtung aus der Areas-Tabelle mitgeprüft
// wird — nicht nur die Middleware isoliert.
func TestLoginRateLimit(t *testing.T) {
	handler, _, teardown := setupMatrix(t)
	defer teardown()

	// Burst ist requestsPerSecond*2 = 10; der 11. Request innerhalb einer
	// Sekunde muss abgewiesen werden. httptest setzt eine feste RemoteAddr,
	// alle Requests teilen sich also denselben Limiter-Key.
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
