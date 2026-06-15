//go:build unit

package tse_repo

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
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
