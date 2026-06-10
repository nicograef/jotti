//go:build unit

package tse_repo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/domain/tse"
)

func TestFiskalyClient_StartAndFinishMapping(t *testing.T) {
	var authCalls int32
	var startCalls int32
	var finishCalls int32

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
		case r.Method == http.MethodPut && strings.Contains(r.URL.Path, "/api/v2/tss/tss-1/tx/"):
			if r.URL.Query().Get("tx_revision") == "1" {
				atomic.AddInt32(&startCalls, 1)
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
				return
			}
			if r.URL.Query().Get("tx_revision") == "2" {
				atomic.AddInt32(&finishCalls, 1)
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
				return
			}
			w.WriteHeader(http.StatusBadRequest)
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

	start, err := client.StartTransaction(context.Background(), "kasse-1", "Kassenbeleg-V1", "Beleg^1")
	if err != nil {
		t.Fatalf("start transaction failed: %v", err)
	}
	if start.TransactionNumber != 42 {
		t.Fatalf("expected transaction number 42, got %d", start.TransactionNumber)
	}
	if start.SignatureCounter != 7 {
		t.Fatalf("expected start signature counter 7, got %d", start.SignatureCounter)
	}

	finish, err := client.FinishTransaction(context.Background(), "kasse-1", start.TransactionNumber, "Kassenbeleg-V1", "Beleg^1")
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

	if atomic.LoadInt32(&authCalls) != 1 {
		t.Fatalf("expected exactly one auth call, got %d", authCalls)
	}
	if atomic.LoadInt32(&startCalls) != 1 {
		t.Fatalf("expected exactly one start call, got %d", startCalls)
	}
	if atomic.LoadInt32(&finishCalls) != 1 {
		t.Fatalf("expected exactly one finish call, got %d", finishCalls)
	}
}

func TestFiskalyClient_UsesProvidedTxIDForStartAndFinish(t *testing.T) {
	const expectedTxID = "deterministic-tx-id"

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
		case r.Method == http.MethodPut && r.URL.Path == "/api/v2/tss/tss-1/tx/"+expectedTxID && r.URL.Query().Get("tx_revision") == "1":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"number":            42,
				"tss_serial_number": "TSS-SN-1",
				"time_start":        1700000000,
				"log": map[string]any{
					"timestamp": 1700000000,
				},
				"signature": map[string]any{
					"counter": 1,
				},
			})
		case r.Method == http.MethodPut && r.URL.Path == "/api/v2/tss/tss-1/tx/"+expectedTxID && r.URL.Query().Get("tx_revision") == "2":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"number":            42,
				"tss_serial_number": "TSS-SN-1",
				"time_start":        1700000000,
				"time_end":          1700000600,
				"log": map[string]any{
					"timestamp": 1700000600,
				},
				"signature": map[string]any{
					"counter": 2,
					"value":   "sig-abc",
				},
			})
		default:
			t.Fatalf("unexpected request %s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)
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

	start, err := client.StartTransaction(context.Background(), expectedTxID, "Kassenbeleg-V1", "Beleg^1")
	if err != nil {
		t.Fatalf("start transaction failed: %v", err)
	}

	if _, err := client.FinishTransaction(context.Background(), expectedTxID, start.TransactionNumber, "Kassenbeleg-V1", "Beleg^1"); err != nil {
		t.Fatalf("finish transaction failed: %v", err)
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

	_, err = client.StartTransaction(context.Background(), "kasse-1", "Kassenbeleg-V1", "Beleg^1")
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

	_, err = client.StartTransaction(context.Background(), "kasse-1", "Kassenbeleg-V1", "Beleg^1")
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
