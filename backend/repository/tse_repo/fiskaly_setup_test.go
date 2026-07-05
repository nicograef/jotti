//go:build unit

package tse_repo

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/domain/tse"
)

// TestFiskalySetupClient_OnlyAuthAndReads bildet den Kontrakt des Prüf-Schritts
// ab: die Setup-Operationen senden ausschließlich die Auth-POST und GET-Requests
// — niemals schreibende Methoden. Außerdem werden Umgebung, TSS-Zustände und die
// Client-serial_number korrekt aus den fiskaly-Antworten gelesen.
func TestFiskalySetupClient_OnlyAuthAndReads(t *testing.T) {
	type call struct {
		method string
		path   string
	}
	var mu sync.Mutex
	var calls []call

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		calls = append(calls, call{method: r.Method, path: r.URL.Path})
		mu.Unlock()

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v2/auth":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":            "token-1",
				"access_token_expires_at": time.Now().Add(1 * time.Hour).Unix(),
				"access_token_claims":     map[string]any{"env": "TEST"},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/tss":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"_env": "TEST",
				"data": []map[string]any{
					{"_id": "tss-1", "state": "INITIALIZED"},
					{"_id": "tss-2", "state": "CREATED"},
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/tss/tss-1/client":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{"_id": "client-1", "serial_number": "kasse-serial-1", "state": "REGISTERED"},
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/tss/tss-2":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"_id": "tss-2", "state": "CREATED", "admin_puk": "puk-refetch",
			})
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := NewFiskalyTSESetupClient(server.URL, tse.SetupCredentials{
		ApiKey:    "api-key",
		ApiSecret: "api-secret",
	}, nil)
	if err != nil {
		t.Fatalf("failed to create setup client: %v", err)
	}

	env, tssList, err := client.ListTSS(context.Background())
	if err != nil {
		t.Fatalf("list tss failed: %v", err)
	}
	if env != tse.UmgebungTest {
		t.Fatalf("expected TEST environment, got %s", env)
	}
	if len(tssList) != 2 || tssList[0].ID != "tss-1" || tssList[0].State != "INITIALIZED" || tssList[1].State != "CREATED" {
		t.Fatalf("unexpected tss list: %+v", tssList)
	}

	clients, err := client.ListClients(context.Background(), "tss-1")
	if err != nil {
		t.Fatalf("list clients failed: %v", err)
	}
	if len(clients) != 1 || clients[0].ID != "client-1" || clients[0].SerialNumber != "kasse-serial-1" || clients[0].State != "REGISTERED" {
		t.Fatalf("unexpected client list: %+v", clients)
	}

	puk, err := client.GetAdminPUK(context.Background(), "tss-2")
	if err != nil {
		t.Fatalf("hole admin puk failed: %v", err)
	}
	if puk != "puk-refetch" {
		t.Fatalf("expected refetched puk, got %q", puk)
	}

	mu.Lock()
	defer mu.Unlock()
	for _, c := range calls {
		isAuth := c.method == http.MethodPost && c.path == "/api/v2/auth"
		isRead := c.method == http.MethodGet
		if !isAuth && !isRead {
			t.Fatalf("setup must only send auth and GET requests, got %s %s", c.method, c.path)
		}
	}
}

// TestFiskalySetupClient_Lebenszyklus bildet den Kontrakt der schreibenden
// Setup-Operationen ab: jede Operation trifft den richtigen fiskaly-Endpunkt mit
// der richtigen Methode und dem richtigen Body (Zustandsübergänge, PUK/PIN,
// serial_number). Der Server gibt die TSS-ID und den PUK aus CreateTSS zurück.
func TestFiskalySetupClient_Lebenszyklus(t *testing.T) {
	type call struct {
		method string
		path   string
		body   map[string]any
	}
	var mu sync.Mutex
	var calls []call

	record := func(r *http.Request) map[string]any {
		body := map[string]any{}
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&body)
		}
		mu.Lock()
		calls = append(calls, call{method: r.Method, path: r.URL.Path, body: body})
		mu.Unlock()
		return body
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := record(r)
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v2/auth":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":            "token-1",
				"access_token_expires_at": time.Now().Add(1 * time.Hour).Unix(),
				"access_token_claims":     map[string]any{"env": "TEST"},
			})
		case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/api/v2/tss/") && strings.Contains(r.URL.Path, "/client/"):
			_ = json.NewEncoder(w).Encode(map[string]any{"_id": "client-1", "serial_number": body["serial_number"], "state": "REGISTERED"})
		case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/api/v2/tss/"):
			// fiskaly waehlt die TSS-ID nicht selbst — der Client erzeugt sie als
			// UUID und PUTtet sie. Der Server spiegelt sie in _id zurueck.
			id := strings.TrimPrefix(r.URL.Path, "/api/v2/tss/")
			_ = json.NewEncoder(w).Encode(map[string]any{"_id": id, "admin_puk": "puk-xyz", "state": "CREATED"})
		case r.Method == http.MethodPatch, r.Method == http.MethodPost:
			_ = json.NewEncoder(w).Encode(map[string]any{})
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := NewFiskalyTSESetupClient(server.URL, tse.SetupCredentials{ApiKey: "api-key", ApiSecret: "api-secret"}, nil)
	if err != nil {
		t.Fatalf("failed to create setup client: %v", err)
	}
	ctx := context.Background()

	erstellt, err := client.CreateTSS(ctx)
	if err != nil {
		t.Fatalf("create tss failed: %v", err)
	}
	if erstellt.ID == "" || erstellt.PUK != "puk-xyz" || erstellt.State != "CREATED" {
		t.Fatalf("unexpected create result: %+v", erstellt)
	}

	if err := client.PersonalisiereTSS(ctx, erstellt.ID); err != nil {
		t.Fatalf("personalize failed: %v", err)
	}
	if err := client.SetAdminPIN(ctx, erstellt.ID, erstellt.PUK, "1234567890"); err != nil {
		t.Fatalf("set admin pin failed: %v", err)
	}
	if err := client.AuthentifiziereAdmin(ctx, erstellt.ID, "1234567890"); err != nil {
		t.Fatalf("admin auth failed: %v", err)
	}
	if err := client.InitialisiereTSS(ctx, erstellt.ID); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}
	if err := client.RegistriereClient(ctx, erstellt.ID, "client-uuid", "kasse-serial"); err != nil {
		t.Fatalf("register client failed: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	// assertCall sucht einen Aufruf, der Methode, Pfad und alle erwarteten
	// Body-Felder erfuellt. Mehrere PATCH-Aufrufe treffen denselben TSS-Pfad
	// (UNINITIALIZED, INITIALIZED) — daher muss der Body mitgeprueft werden.
	assertCall := func(method, pathSuffix string, wantBody map[string]any) {
		for _, c := range calls {
			if c.method != method || !strings.HasSuffix(c.path, pathSuffix) {
				continue
			}
			if bodyMatcht(c.body, wantBody) {
				return
			}
		}
		t.Fatalf("expected a %s request to %s with body %v, none found", method, pathSuffix, wantBody)
	}

	assertCall(http.MethodPut, "/tss/"+erstellt.ID, nil)
	assertCall(http.MethodPatch, "/tss/"+erstellt.ID, map[string]any{"state": "INITIALIZED"})
	assertCall(http.MethodPatch, "/tss/"+erstellt.ID+"/admin", map[string]any{"admin_puk": "puk-xyz", "new_admin_pin": "1234567890"})
	assertCall(http.MethodPost, "/tss/"+erstellt.ID+"/admin/auth", map[string]any{"admin_pin": "1234567890"})
	assertCall(http.MethodPut, "/tss/"+erstellt.ID+"/client/client-uuid", map[string]any{"serial_number": "kasse-serial"})
}

// TestFiskalySetupClient_ReaktiviereClient bildet den Kontrakt der
// Client-Reaktivierung (F7) ab: ein DEREGISTERED Client wird per PATCH mit
// state=REGISTERED auf demselben Client-Pfad reaktiviert (kein neuer Client),
// mit anliegendem Admin-Token (Bearer).
func TestFiskalySetupClient_ReaktiviereClient(t *testing.T) {
	var (
		mu      sync.Mutex
		method  string
		path    string
		body    map[string]any
		authHdr string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v2/auth":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":            "token-1",
				"access_token_expires_at": time.Now().Add(1 * time.Hour).Unix(),
				"access_token_claims":     map[string]any{"env": "TEST"},
			})
		case r.Method == http.MethodPatch:
			b := map[string]any{}
			_ = json.NewDecoder(r.Body).Decode(&b)
			mu.Lock()
			method, path, body = r.Method, r.URL.Path, b
			authHdr = r.Header.Get("Authorization")
			mu.Unlock()
			_ = json.NewEncoder(w).Encode(map[string]any{})
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := NewFiskalyTSESetupClient(server.URL, tse.SetupCredentials{ApiKey: "api-key", ApiSecret: "api-secret"}, nil)
	if err != nil {
		t.Fatalf("failed to create setup client: %v", err)
	}

	if err := client.ReaktiviereClient(context.Background(), "tss-1", "client-1"); err != nil {
		t.Fatalf("reaktiviere client failed: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if method != http.MethodPatch || path != "/api/v2/tss/tss-1/client/client-1" {
		t.Fatalf("expected PATCH to the client path, got %s %s", method, path)
	}
	if body["state"] != "REGISTERED" {
		t.Fatalf("expected body state=REGISTERED, got %v", body)
	}
	if !strings.HasPrefix(authHdr, "Bearer ") {
		t.Fatalf("expected an admin bearer token, got %q", authHdr)
	}
}

// TestFiskalySetupClient_RetrieveTSSStammdaten bildet den Kontrakt der
// Stammdaten-Leseoperation fuer den DSFinV-K-Export ab: ein GET auf die
// TSS-Ressource liest signature_algorithm, public_key, certificate und
// signature_timestamp_format (Log-Time-Format) — und sendet ausschliesslich
// Auth- und GET-Requests.
func TestFiskalySetupClient_RetrieveTSSStammdaten(t *testing.T) {
	var (
		mu    sync.Mutex
		calls []string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		calls = append(calls, r.Method)
		mu.Unlock()
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v2/auth":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":            "token-1",
				"access_token_expires_at": time.Now().Add(1 * time.Hour).Unix(),
				"access_token_claims":     map[string]any{"env": "TEST"},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/tss/tss-1":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"_id":                        "tss-1",
				"state":                      "INITIALIZED",
				"signature_algorithm":        "ecdsa-plain-SHA256",
				"public_key":                 "public-key-b64",
				"certificate":                "certificate-b64",
				"signature_timestamp_format": "unixTime",
			})
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := NewFiskalyTSESetupClient(server.URL, tse.SetupCredentials{ApiKey: "api-key", ApiSecret: "api-secret"}, nil)
	if err != nil {
		t.Fatalf("failed to create setup client: %v", err)
	}

	stammdaten, err := client.RetrieveTSSStammdaten(context.Background(), "tss-1")
	if err != nil {
		t.Fatalf("retrieve tss stammdaten failed: %v", err)
	}
	want := tse.TSSStammdaten{
		SignaturAlgorithmus: "ecdsa-plain-SHA256",
		PublicKey:           "public-key-b64",
		Zertifikat:          "certificate-b64",
		LogTimeFormat:       "unixTime",
	}
	if stammdaten != want {
		t.Fatalf("unexpected stammdaten, got %+v want %+v", stammdaten, want)
	}

	mu.Lock()
	defer mu.Unlock()
	for _, m := range calls {
		if m != http.MethodPost && m != http.MethodGet {
			t.Fatalf("stammdaten read must only send auth and GET requests, got %s", m)
		}
	}
}

// bodyMatcht meldet, ob jedes erwartete Feld im tatsaechlichen Body steht.
func bodyMatcht(body, want map[string]any) bool {
	for k, v := range want {
		if body[k] != v {
			return false
		}
	}
	return true
}

// TestFiskalySetupClient_AuthFailure sichert, dass falsche Zugangsdaten als
// ErrSetupAuthFehlgeschlagen gemeldet werden — die Grundlage für eine
// verständliche Fehlermeldung im Wizard.
func TestFiskalySetupClient_AuthFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/v2/auth" {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": "E_UNAUTHORIZED", "message": "invalid credentials"})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, err := NewFiskalyTSESetupClient(server.URL, tse.SetupCredentials{
		ApiKey:    "wrong",
		ApiSecret: "wrong",
	}, nil)
	if err != nil {
		t.Fatalf("failed to create setup client: %v", err)
	}

	_, _, err = client.ListTSS(context.Background())
	if !errors.Is(err, tse.ErrSetupAuthFehlgeschlagen) {
		t.Fatalf("expected ErrSetupAuthFehlgeschlagen, got %v", err)
	}
}

// TestFiskalySetupClient_AdminPINBlocked sichert, dass eine nach fuenf
// Fehlversuchen gesperrte Admin-PIN (fiskaly: Status 423, Code E_ADMIN_PIN_BLOCKED)
// als ErrSetupAuthFehlgeschlagen gemeldet wird. So laeuft die Uebernahme in die
// PIN-Sackgasse (mit PUK-Reset als Ausweg) statt in einen technischen Fehler.
func TestFiskalySetupClient_AdminPINBlocked(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v2/auth":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":            "token-1",
				"access_token_expires_at": time.Now().Add(1 * time.Hour).Unix(),
				"access_token_claims":     map[string]any{"env": "TEST"},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v2/tss/tss-1/admin/auth":
			w.WriteHeader(http.StatusLocked)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": "E_ADMIN_PIN_BLOCKED", "message": "admin pin is blocked"})
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := NewFiskalyTSESetupClient(server.URL, tse.SetupCredentials{ApiKey: "api-key", ApiSecret: "api-secret"}, nil)
	if err != nil {
		t.Fatalf("failed to create setup client: %v", err)
	}

	err = client.AuthentifiziereAdmin(context.Background(), "tss-1", "0000000000")
	if !errors.Is(err, tse.ErrSetupAuthFehlgeschlagen) {
		t.Fatalf("expected ErrSetupAuthFehlgeschlagen for a blocked admin pin, got %v", err)
	}
}
