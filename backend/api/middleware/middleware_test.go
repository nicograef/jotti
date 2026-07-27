//go:build unit

package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/jwt"
	"github.com/nicograef/jotti/backend/domain/user"
)

// stubUsers satisfies UserGetter with a fixed user and error.
type stubUsers struct {
	user user.User
	err  error
}

func (s stubUsers) GetUser(context.Context, int) (user.User, error) {
	return s.user, s.err
}

// activeUser returns a UserGetter whose user is active and has the given role.
// The middleware authorizes against this DB role, not the token claim.
func activeUser(role user.Role) stubUsers {
	return stubUsers{user: user.User{ID: 1, Role: role, Status: user.ActiveStatus}}
}

func TestCorrelationIDMiddleware_GeneratesID(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID, ok := r.Context().Value(CorrelationIDKey).(string)
		if !ok || correlationID == "" {
			t.Error("expected correlation ID in context")
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := CorrelationIDMiddleware(handler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	correlationID := rec.Header().Get("X-Correlation-ID")
	if correlationID == "" {
		t.Error("expected X-Correlation-ID header to be set")
	}
}

func TestCorrelationIDMiddleware_UsesExisting(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := CorrelationIDMiddleware(handler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Correlation-ID", "test-correlation-id")
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	correlationID := rec.Header().Get("X-Correlation-ID")
	if correlationID != "test-correlation-id" {
		t.Errorf("expected correlation ID 'test-correlation-id', got '%s'", correlationID)
	}
}

// Der reale 500-Fall aus dem PRD: Ein Panic im Handler muss trotzdem eine
// X-Correlation-ID-Antwortheader tragen, damit der Verein die Fehler-Referenz
// im Toast melden kann und der Betreiber sie im Server-Log wiederfindet.
func TestCorrelationIDMiddleware_PanicResponseHatCorrelationIDHeader(t *testing.T) {
	handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	})

	middleware := CorrelationIDMiddleware(RecoveryMiddleware(handler))
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
	if correlationID := rec.Header().Get("X-Correlation-ID"); correlationID == "" {
		t.Error("expected X-Correlation-ID header to be set on 500 response")
	}
}

func TestRateLimitMiddleware_AllowsWithinLimit(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RateLimitMiddleware(10)(handler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestRateLimitMiddleware_BlocksExceedingLimit(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RateLimitMiddleware(1)(handler)

	// Fill the limiter
	for range 10 {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		middleware.ServeHTTP(rec, req)
	}

	// This request should be rate limited
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", rec.Code)
	}
}

// Die Limit-Antwort muss dem kanonischen JSON-Fehlerformat der API folgen
// (`{"code":"rate_limited"}`), nicht Plain Text — sonst kann der Backend-Client
// den Code nicht parsen und meldet dem Helfer fälschlich einen Serverfehler.
func TestRateLimitMiddleware_AntwortFolgtJSONFehlerformat(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RateLimitMiddleware(1)(handler)

	for range 10 {
		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		rec := httptest.NewRecorder()
		middleware.ServeHTTP(rec, req)
	}

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	rec := httptest.NewRecorder()
	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status 429, got %d", rec.Code)
	}
	if contentType := rec.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", contentType)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != `{"code":"rate_limited"}` {
		t.Errorf("expected body %q, got %q", `{"code":"rate_limited"}`, body)
	}
}

// Der Limiter-Key darf sich nicht über Client-kontrollierte X-Forwarded-For-Einträge
// variieren lassen: Nur der LETZTE Eintrag (vom eigenen Reverse-Proxy angehängt) zählt.
func TestRateLimitMiddleware_XForwardedForSpoofingUmgehtLimitNicht(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RateLimitMiddleware(1)(handler)

	// Der Angreifer variiert vorangestellte Einträge, die echte IP (letzter Eintrag,
	// vom Proxy gesetzt) bleibt gleich → alle Requests treffen denselben Limiter.
	blocked := false
	for i := range 10 {
		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		req.Header.Set("X-Forwarded-For", fmt.Sprintf("10.0.0.%d, 203.0.113.7", i))
		rec := httptest.NewRecorder()
		middleware.ServeHTTP(rec, req)
		if rec.Code == http.StatusTooManyRequests {
			blocked = true
		}
	}

	if !blocked {
		t.Error("expected rate limit to trigger despite varying X-Forwarded-For prefixes")
	}
}

func TestClientIP(t *testing.T) {
	cases := []struct {
		name       string
		remoteAddr string
		xff        string
		want       string
	}{
		// Ohne Proxy zählt RemoteAddr — aber OHNE den ephemeren Port, sonst
		// bekäme jede Verbindung desselben Clients einen frischen Limiter-Key.
		{"ohne Header: RemoteAddr ohne Port", "192.0.2.1:1234", "", "192.0.2.1"},
		{"ohne Header: anderer Port derselben IP", "192.0.2.1:56789", "", "192.0.2.1"},
		{"ohne Header: IPv6 RemoteAddr ohne Port", "[2001:db8::1]:1234", "", "2001:db8::1"},
		{"ein Eintrag", "192.0.2.1:1234", "203.0.113.7", "203.0.113.7"},
		{"Client-Spoofing: letzter Eintrag zählt", "192.0.2.1:1234", "1.2.3.4, 5.6.7.8, 203.0.113.7", "203.0.113.7"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/test", nil)
			req.RemoteAddr = tc.remoteAddr
			if tc.xff != "" {
				req.Header.Set("X-Forwarded-For", tc.xff)
			}
			if got := clientIP(req); got != tc.want {
				t.Errorf("clientIP() = %q, want %q", got, tc.want)
			}
		})
	}
}

// Ohne vertrauenswürdigen Proxy (kein X-Forwarded-For) müssen zwei Requests
// desselben Clients mit WECHSELNDEM ephemerem Port denselben Limiter treffen —
// sonst greift das Limit nie und die Limiter-Map wächst je Verbindung.
func TestRateLimitMiddleware_GleicheIPWechselndePortsTeilenLimiter(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := RateLimitMiddleware(1)(handler)

	blocked := false
	for port := 40000; port < 40010; port++ {
		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		req.RemoteAddr = fmt.Sprintf("192.0.2.1:%d", port)
		rec := httptest.NewRecorder()
		mw.ServeHTTP(rec, req)
		if rec.Code == http.StatusTooManyRequests {
			blocked = true
		}
	}

	if !blocked {
		t.Error("expected rate limit to trigger for the same IP despite changing ephemeral ports")
	}
}

func TestJwtMiddleware_ValidToken(t *testing.T) {
	secret := "test-secret"
	token, err := jwt.GenerateJWTTokenForUser(1, "admin", "admin", secret)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := NewJwtMiddleware(secret, []string{"admin"}, activeUser(user.AdminRole))(handler)
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

// Fehlende Tokens sind ein Authentifizierungsfehler: 401, damit das Frontend
// den Auto-Logout auslöst.
func TestJwtMiddleware_NoToken(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := NewJwtMiddleware("test-secret", []string{"admin"}, activeUser(user.AdminRole))(handler)
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

// Ungültige Tokens (falsches Secret, abgelaufen) sind ein
// Authentifizierungsfehler: 401, damit das Frontend den Auto-Logout auslöst.
func TestJwtMiddleware_InvalidToken(t *testing.T) {
	cases := []struct {
		name  string
		token func(t *testing.T) string
	}{
		{"wrong secret", func(t *testing.T) string {
			token, err := jwt.GenerateJWTTokenForUser(1, "admin", "admin", "other-secret")
			if err != nil {
				t.Fatalf("failed to generate token: %v", err)
			}
			return token
		}},
		{"expired token", func(t *testing.T) string {
			return expiredToken(t, "test-secret")
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Error("handler must not be called")
			})
			middleware := NewJwtMiddleware("test-secret", []string{"admin"}, activeUser(user.AdminRole))(handler)
			req := httptest.NewRequest(http.MethodPost, "/admin", nil)
			req.Header.Set("Authorization", "Bearer "+tc.token(t))
			rec := httptest.NewRecorder()

			middleware.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("expected status 401, got %d", rec.Code)
			}
		})
	}
}

// expiredToken signiert ein bereits abgelaufenes jotti-Token mit dem gegebenen
// Secret (GenerateJWTTokenForUser erzeugt nur gültige Tokens).
func expiredToken(t *testing.T, secret string) string {
	t.Helper()
	claims := gojwt.MapClaims{
		"iss":      "jotti",
		"iat":      gojwt.NewNumericDate(time.Now().UTC().Add(-13 * time.Hour)),
		"exp":      gojwt.NewNumericDate(time.Now().UTC().Add(-1 * time.Hour)),
		"sub":      1,
		"username": "admin",
		"role":     "admin",
	}
	token, err := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign expired token: %v", err)
	}
	return token
}

func TestJwtMiddleware_InvalidBearerFormat(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	testCases := []struct {
		name  string
		value string
	}{
		{"short header", "Bear"},
		{"wrong prefix", "Basic xyz123"},
		{"just Bearer", "Bearer"},
		{"bearer lowercase", "bearer token123"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			middleware := NewJwtMiddleware("test-secret", []string{"admin"}, activeUser(user.AdminRole))(handler)
			req := httptest.NewRequest(http.MethodGet, "/admin", nil)
			req.Header.Set("Authorization", tc.value)
			rec := httptest.NewRecorder()

			middleware.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("expected status 401 for %q, got %d", tc.value, rec.Code)
			}
		})
	}
}

// Ein authentifizierter Benutzer ohne ausreichende Rolle bekommt 403 —
// kein Auto-Logout, die Sitzung bleibt bestehen.
func TestJwtMiddleware_ServiceRole(t *testing.T) {
	secret := "test-secret"
	token, err := jwt.GenerateJWTTokenForUser(2, "service", "service", secret)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := NewJwtMiddleware(secret, []string{"admin"}, activeUser(user.ServiceRole))(handler)
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

// Die Autorisierung prüft die Rolle aus der Datenbank, nicht aus dem
// Token-Claim: Ein Rollenwechsel wirkt beim nächsten Request, nicht erst
// nach Token-Ablauf.
func TestJwtMiddleware_RollenwechselWirktSofort(t *testing.T) {
	secret := "test-secret"

	cases := []struct {
		name      string
		tokenRole string
		dbRole    user.Role
		wantCode  int
	}{
		{"Herabstufung: Token admin, DB service", "admin", user.ServiceRole, http.StatusForbidden},
		{"Heraufstufung: Token service, DB admin", "service", user.AdminRole, http.StatusOK},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := jwt.GenerateJWTTokenForUser(1, "someone", tc.tokenRole, secret)
			if err != nil {
				t.Fatalf("failed to generate token: %v", err)
			}

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			middleware := NewJwtMiddleware(secret, []string{"admin"}, activeUser(tc.dbRole))(handler)
			req := httptest.NewRequest(http.MethodPost, "/admin", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()

			middleware.ServeHTTP(rec, req)

			if rec.Code != tc.wantCode {
				t.Errorf("expected status %d, got %d", tc.wantCode, rec.Code)
			}
		})
	}
}

func TestServiceMiddleware_ValidToken(t *testing.T) {
	secret := "test-secret"
	token, err := jwt.GenerateJWTTokenForUser(2, "service", "service", secret)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := NewJwtMiddleware(secret, []string{"service"}, activeUser(user.ServiceRole))(handler)
	req := httptest.NewRequest(http.MethodGet, "/service", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestServiceleitungRole_AllowedForServiceEndpoints(t *testing.T) {
	secret := "test-secret"
	token, err := jwt.GenerateJWTTokenForUser(3, "serviceleitung", "serviceleitung", secret)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := NewJwtMiddleware(secret, []string{"admin", "serviceleitung", "service"}, activeUser(user.ServiceleitungRole))(handler)
	req := httptest.NewRequest(http.MethodGet, "/service", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestServiceleitungRole_AllowedForCancelEndpoint(t *testing.T) {
	secret := "test-secret"
	token, err := jwt.GenerateJWTTokenForUser(3, "serviceleitung", "serviceleitung", secret)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := NewJwtMiddleware(secret, []string{"admin", "serviceleitung"}, activeUser(user.ServiceleitungRole))(handler)
	req := httptest.NewRequest(http.MethodGet, "/service/cancel", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestJwtMiddleware_SetsUserNameInContext(t *testing.T) {
	secret := "test-secret"
	token, err := jwt.GenerateJWTTokenForUser(1, "admin", "admin", secret)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, ok := r.Context().Value(UserNameKey).(string)
		if !ok || username != "admin" {
			t.Errorf("expected UserNameKey 'admin' in context, got '%s'", username)
		}
		userID, ok := r.Context().Value(UserIDKey).(int)
		if !ok || userID != 1 {
			t.Errorf("expected UserIDKey 1 in context, got %d", userID)
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := NewJwtMiddleware(secret, []string{"admin"}, activeUser(user.AdminRole))(handler)
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

// Deaktivierte oder gelöschte Benutzer verlieren den Zugriff sofort (401),
// nicht erst beim Ablauf ihres Tokens.
func TestJwtMiddleware_UserStatusCheck(t *testing.T) {
	secret := "test-secret"
	token, err := jwt.GenerateJWTTokenForUser(1, "admin", "admin", secret)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	cases := []struct {
		name     string
		users    stubUsers
		wantCode int
	}{
		{"inactive user", stubUsers{user: user.User{ID: 1, Status: user.InactiveStatus}}, http.StatusUnauthorized},
		{"deleted user", stubUsers{user: user.User{ID: 1, Status: user.DeletedStatus}}, http.StatusUnauthorized},
		{"user no longer exists", stubUsers{err: db.ErrNotFound}, http.StatusUnauthorized},
		{"lookup error", stubUsers{err: db.ErrDatabase}, http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Error("handler must not be called")
			})
			middleware := NewJwtMiddleware(secret, []string{"admin"}, tc.users)(handler)
			req := httptest.NewRequest(http.MethodPost, "/admin", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()

			middleware.ServeHTTP(rec, req)

			if rec.Code != tc.wantCode {
				t.Errorf("expected status %d, got %d", tc.wantCode, rec.Code)
			}
		})
	}
}

func TestServiceRole_DeniedForCancelEndpoint(t *testing.T) {
	secret := "test-secret"
	token, err := jwt.GenerateJWTTokenForUser(2, "service", "service", secret)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := NewJwtMiddleware(secret, []string{"admin", "serviceleitung"}, activeUser(user.ServiceRole))(handler)
	req := httptest.NewRequest(http.MethodGet, "/service/cancel", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

// Ein Panic in einem Handler ergibt eine 500-Antwort im bestehenden
// Fehler-Response-Format; der Prozess lebt weiter und bedient den naechsten
// Request regulaer.
func TestRecoveryMiddleware_PanicErgibt500UndNaechsterRequestFunktioniert(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/panic", func(http.ResponseWriter, *http.Request) {
		panic("kaputter handler")
	})
	mux.HandleFunc("/ok", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := RecoveryMiddleware(mux)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/panic", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != `{"code":"internal_server_error"}` {
		t.Fatalf("expected error response format, got %q", body)
	}

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/ok", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected next request to succeed with 200, got %d", rec.Code)
	}
}
