//go:build unit

package tse_repo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/domain/tse"
)

const (
	testTxID = "8c0f9c4e-3a52-4f5d-9e6b-2d1c7a8b4f01"

	// Beispiel aus DSFinV-K Anhang I / fiskaly-Spec: Klartext und sein Base64.
	specProcessData       = "Beleg^0.00_2.55_0.00_0.00_0.00^2.55:Bar"
	specProcessDataBase64 = "QmVsZWdeMC4wMF8yLjU1XzAuMDBfMC4wMF8wLjAwXjIuNTU6QmFy"
)

// TestFiskalyClient_StartAndFinishContract bildet den API-Kontrakt der fiskaly
// SIGN-DE-Spec 2.2.2 ab: Start ohne Schema (DSFinV-K: processType/processData
// bei Start immer leer), Finish mit Base64-codiertem process_data,
// UUID-Transaktionspfade und Revisionsfolge 1→2.
func TestFiskalyClient_StartAndFinishContract(t *testing.T) {
	var authCalls int32
	var mu sync.Mutex
	var revisions []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v2/auth":
			atomic.AddInt32(&authCalls, 1)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":            "token-1",
				"access_token_expires_at": time.Now().Add(1 * time.Hour).Unix(),
				"access_token_claims": map[string]any{
					"env": "TEST",
				},
			})
		case r.Method == http.MethodPut && r.URL.Path == "/api/v2/tss/tss-1/tx/"+testTxID:
			revision := r.URL.Query().Get("tx_revision")
			mu.Lock()
			revisions = append(revisions, revision)
			mu.Unlock()

			body := map[string]any{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("failed to decode request body: %v", err)
			}
			if body["client_id"] != "client-1" {
				t.Errorf("expected client_id client-1, got %v", body["client_id"])
			}

			switch revision {
			case "1":
				if body["state"] != "ACTIVE" {
					t.Errorf("expected state ACTIVE on start, got %v", body["state"])
				}
				if _, hasSchema := body["schema"]; hasSchema {
					t.Errorf("start request must not contain a schema, got %v", body["schema"])
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"number":            42,
					"tss_serial_number": "TSS-SN-1",
					"time_start":        1700000000,
					"log": map[string]any{
						"timestamp": 1700000000,
					},
					"signature": map[string]any{
						"counter": "7",
					},
				})
			case "2":
				if body["state"] != "FINISHED" {
					t.Errorf("expected state FINISHED on finish, got %v", body["state"])
				}
				schema, _ := body["schema"].(map[string]any)
				raw, ok := schema["raw"].(map[string]any)
				if !ok {
					t.Errorf("finish request must contain schema.raw, got %v", body["schema"])
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				if raw["process_type"] != "Kassenbeleg-V1" {
					t.Errorf("expected process_type Kassenbeleg-V1, got %v", raw["process_type"])
				}
				if raw["process_data"] != specProcessDataBase64 {
					t.Errorf("expected base64 process_data %q, got %v", specProcessDataBase64, raw["process_data"])
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"number":            42,
					"tss_serial_number": "TSS-SN-1",
					"time_start":        1700000000,
					"time_end":          1700000600,
					"qr_code_data":      "V0;...",
					"log": map[string]any{
						"timestamp": 1700000600,
					},
					"signature": map[string]any{
						"counter": 8,
						"value":   "sig-abc",
					},
				})
			default:
				t.Errorf("unexpected tx_revision %q", revision)
				w.WriteHeader(http.StatusBadRequest)
			}
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := NewFiskalyTSEClient(server.URL, tse.Credentials{
		ApiKey:    "api-key",
		ApiSecret: "api-secret",
		TssID:     "tss-1",
		ClientID:  "client-1",
	}, nil)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	start, err := client.StartTransaction(context.Background(), testTxID, "Kassenbeleg-V1", "")
	if err != nil {
		t.Fatalf("start transaction failed: %v", err)
	}
	if start.TransactionNumber != 42 {
		t.Fatalf("expected transaction number 42, got %d", start.TransactionNumber)
	}
	if start.SignatureCounter != 7 {
		t.Fatalf("expected start signature counter 7, got %d", start.SignatureCounter)
	}

	finish, err := client.FinishTransaction(context.Background(), testTxID, start.TransactionNumber, "Kassenbeleg-V1", specProcessData)
	if err != nil {
		t.Fatalf("finish transaction failed: %v", err)
	}
	if finish.SignatureCounter != 8 {
		t.Fatalf("expected finish signature counter 8, got %d", finish.SignatureCounter)
	}
	if finish.Signature != "sig-abc" {
		t.Fatalf("expected signature sig-abc, got %q", finish.Signature)
	}
	if finish.QRCodeData != "V0;..." {
		t.Fatalf("expected qr_code_data to be mapped")
	}

	expectedRevisions := []string{"1", "2"}
	if len(revisions) != len(expectedRevisions) || revisions[0] != "1" || revisions[1] != "2" {
		t.Fatalf("expected revision sequence %v, got %v", expectedRevisions, revisions)
	}
	if atomic.LoadInt32(&authCalls) != 1 {
		t.Fatalf("expected exactly one auth call, got %d", authCalls)
	}
}

func TestFiskalyClient_RefreshesTokenOn401(t *testing.T) {
	var authCalls int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v2/auth":
			current := atomic.AddInt32(&authCalls, 1)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":            "token-" + strconv.Itoa(int(current)),
				"access_token_expires_at": time.Now().Add(1 * time.Hour).Unix(),
				"access_token_claims": map[string]any{
					"env": "TEST",
				},
			})
		case r.Method == http.MethodPut && strings.Contains(r.URL.Path, "/api/v2/tss/tss-1/tx/"):
			authHeader := r.Header.Get("Authorization")
			if authHeader == "Bearer token-1" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"number":            1,
				"tss_serial_number": "TSS-1",
				"time_start":        1700000000,
				"log": map[string]any{
					"timestamp": 1700000000,
				},
				"signature": map[string]any{
					"counter": 1,
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := NewFiskalyTSEClient(server.URL, tse.Credentials{
		ApiKey:    "api-key",
		ApiSecret: "api-secret",
		TssID:     "tss-1",
		ClientID:  "client-1",
	}, nil)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	_, err = client.StartTransaction(context.Background(), testTxID, "Kassenbeleg-V1", "")
	if err != nil {
		t.Fatalf("start transaction failed: %v", err)
	}

	if atomic.LoadInt32(&authCalls) != 2 {
		t.Fatalf("expected token refresh after 401, auth calls=%d", authCalls)
	}
}

func TestFiskalyClient_RetriesOnRetryableErrors(t *testing.T) {
	var txCalls int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v2/auth":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":            "token-1",
				"access_token_expires_at": time.Now().Add(1 * time.Hour).Unix(),
				"access_token_claims": map[string]any{
					"env": "TEST",
				},
			})
		case r.Method == http.MethodPut && strings.Contains(r.URL.Path, "/api/v2/tss/tss-1/tx/"):
			call := atomic.AddInt32(&txCalls, 1)
			if call == 1 {
				w.Header().Set("Retry-After", "0")
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			if call == 2 {
				w.Header().Set("Retry-After", "0")
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"number":            9,
				"tss_serial_number": "TSS-1",
				"time_start":        1700000000,
				"log": map[string]any{
					"timestamp": 1700000000,
				},
				"signature": map[string]any{
					"counter": 4,
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := NewFiskalyTSEClient(server.URL, tse.Credentials{
		ApiKey:    "api-key",
		ApiSecret: "api-secret",
		TssID:     "tss-1",
		ClientID:  "client-1",
	}, nil)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	_, err = client.StartTransaction(context.Background(), testTxID, "Kassenbeleg-V1", "")
	if err != nil {
		t.Fatalf("expected retry to recover, got error: %v", err)
	}
	if atomic.LoadInt32(&txCalls) != 3 {
		t.Fatalf("expected 3 transaction calls due to retries, got %d", txCalls)
	}
}

func TestFiskalyClient_TestConnection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v2/auth":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":            "token-1",
				"access_token_expires_at": time.Now().Add(1 * time.Hour).Unix(),
				"access_token_claims": map[string]any{
					"env": "LIVE",
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/tss/tss-1":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"state": "INITIALIZED",
				"_env":  "TEST",
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := NewFiskalyTSEClient(server.URL, tse.Credentials{
		ApiKey:    "api-key",
		ApiSecret: "api-secret",
		TssID:     "tss-1",
		ClientID:  "client-1",
	}, nil)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	status, err := client.TestConnection(context.Background())
	if err != nil {
		t.Fatalf("test connection failed: %v", err)
	}
	if status.Umgebung != tse.UmgebungLive {
		t.Fatalf("expected LIVE environment from token claim, got %s", status.Umgebung)
	}
	if status.TSSState != "INITIALIZED" {
		t.Fatalf("expected state INITIALIZED, got %s", status.TSSState)
	}
}
