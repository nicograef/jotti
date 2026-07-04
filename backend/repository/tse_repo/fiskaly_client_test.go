//go:build unit

package tse_repo

import (
	"context"
	"encoding/json"
	"errors"
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

	start, err := client.StartTransaction(context.Background(), testTxID)
	if err != nil {
		t.Fatalf("start transaction failed: %v", err)
	}
	if start.TransactionNumber != 42 {
		t.Fatalf("expected transaction number 42, got %d", start.TransactionNumber)
	}
	if start.SignatureCounter != 7 {
		t.Fatalf("expected start signature counter 7, got %d", start.SignatureCounter)
	}

	finish, err := client.FinishTransaction(context.Background(), testTxID, "Kassenbeleg-V1", specProcessData)
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

	_, err = client.StartTransaction(context.Background(), testTxID)
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

	_, err = client.StartTransaction(context.Background(), testTxID)
	if err != nil {
		t.Fatalf("expected retry to recover, got error: %v", err)
	}
	if atomic.LoadInt32(&txCalls) != 3 {
		t.Fatalf("expected 3 transaction calls due to retries, got %d", txCalls)
	}
}

// TestFiskalyClient_RetrieveTransaction bildet den Kontrakt von "Retrieve a
// transaction" ab: GET auf den UUID-Transaktionspfad, Antwort enthaelt state
// und dieselben Signaturfelder wie ein Finish-Response.
func TestFiskalyClient_RetrieveTransaction(t *testing.T) {
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
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/tss/tss-1/tx/"+testTxID:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"number":            42,
				"state":             "FINISHED",
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

	result, err := client.RetrieveTransaction(context.Background(), testTxID)
	if err != nil {
		t.Fatalf("retrieve transaction failed: %v", err)
	}
	if result.State != tse.TransactionStateFinished {
		t.Fatalf("expected state FINISHED, got %q", result.State)
	}
	if result.TransactionNumber != 42 {
		t.Fatalf("expected transaction number 42, got %d", result.TransactionNumber)
	}
	if result.Signature != "sig-abc" || result.SignatureCounter != 8 {
		t.Fatalf("expected signature sig-abc/8, got %q/%d", result.Signature, result.SignatureCounter)
	}
	if result.QRCodeData != "V0;..." {
		t.Fatalf("expected qr_code_data to be mapped")
	}
	if result.LogTimeStart.Unix() != 1700000000 || result.LogTimeEnd.Unix() != 1700000600 {
		t.Fatalf("expected mapped unix times, got %v / %v", result.LogTimeStart, result.LogTimeEnd)
	}
}

func TestFiskalyClient_RetrieveTransaction_NotFound(t *testing.T) {
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
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/api/v2/tss/tss-1/tx/"):
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code":        "E_TX_NOT_FOUND",
				"message":     "transaction not found",
				"status_code": 404,
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

	_, err = client.RetrieveTransaction(context.Background(), testTxID)
	if !errors.Is(err, tse.ErrTransactionNichtGefunden) {
		t.Fatalf("expected ErrTransactionNichtGefunden, got %v", err)
	}
}

// TestKlassifiziereSignierFehler bildet die Fehlertaxonomie ab:
// auftragsspezifische Ablehnungen (400/409/422) werden als tse.AuftragsFehler
// gekennzeichnet, TSS-Zustandscodes und alle uebrigen Fehler bleiben TSE-weit.
func TestKlassifiziereSignierFehler(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		auftragsFehler bool
	}{
		{"400 abgelehnte processData", apiError{StatusCode: 400, Code: "E_FAILED_SCHEMA_VALIDATION"}, true},
		{"400 ohne Code", apiError{StatusCode: 400}, true},
		{"409 Transaktionszustand", apiError{StatusCode: 409, Code: "E_TX_NO_TYPE_DEFINED"}, true},
		{"422 Unprocessable", apiError{StatusCode: 422}, true},
		{"400 TSS nicht initialisiert", apiError{StatusCode: 400, Code: "E_TSS_NOT_INITIALIZED"}, false},
		{"400 TSS deaktiviert", apiError{StatusCode: 400, Code: "E_TSS_DISABLED"}, false},
		{"400 Client deregistriert", apiError{StatusCode: 400, Code: "E_CLIENT_DEREGISTERED"}, false},
		{"400 Transaktionslimit", apiError{StatusCode: 400, Code: "E_TX_LIMIT_REACHED"}, false},
		{"401 Unauthorized", apiError{StatusCode: 401}, false},
		{"403 Forbidden", apiError{StatusCode: 403}, false},
		{"404 TSS nicht gefunden", apiError{StatusCode: 404, Code: "E_TSS_NOT_FOUND"}, false},
		{"423 TSS defekt", apiError{StatusCode: 423, Code: "E_TSS_DEFECTIVE"}, false},
		{"429 Rate-Limit", apiError{StatusCode: 429}, false},
		{"503 Service Unavailable", apiError{StatusCode: 503}, false},
		{"Verbindungsfehler", errors.New("connection refused"), false},
		{"Context-Deadline", context.DeadlineExceeded, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			klassifiziert := klassifiziereSignierFehler(tt.err)
			if got := tse.IstAuftragsFehler(klassifiziert); got != tt.auftragsFehler {
				t.Fatalf("expected IstAuftragsFehler=%v, got %v", tt.auftragsFehler, got)
			}
			// Der Original-Fehler bleibt per errors.As/Is erreichbar.
			var apiErr apiError
			if errors.As(tt.err, &apiErr) && !errors.As(klassifiziert, &apiErr) {
				t.Fatal("expected wrapped apiError to stay reachable via errors.As")
			}
		})
	}
}

// Die Klassifizierung ist in Start/Finish verdrahtet: Eine 400-Ablehnung durch
// fiskaly kommt als tse.AuftragsFehler beim Aufrufer an.
func TestFiskalyClient_FinishTransaction_AblehnungAlsAuftragsFehler(t *testing.T) {
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
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code":        "E_FAILED_SCHEMA_VALIDATION",
				"message":     "process_data does not match the schema",
				"status_code": 400,
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

	_, err = client.FinishTransaction(context.Background(), testTxID, "Kassenbeleg-V1", specProcessData)
	if err == nil {
		t.Fatal("expected error from rejected finish")
	}
	if !tse.IstAuftragsFehler(err) {
		t.Fatalf("expected auftragsspezifischen Fehler, got %v", err)
	}
}

// TestFiskalyClient_TestConnection bildet den Kontrakt des Verbindungstests ab:
// neben dem TSS-Zustand wird auch der Client abgefragt, und dessen state sowie
// serial_number landen aufgeschluesselt im VerbindungStatus.
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
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/tss/tss-1/client/client-1":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"state":         "REGISTERED",
				"serial_number": "kasse-serial-1",
			})
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
	if status.ClientState != "REGISTERED" {
		t.Fatalf("expected client state REGISTERED, got %s", status.ClientState)
	}
	if status.ClientSerialNumber != "kasse-serial-1" {
		t.Fatalf("expected client serial kasse-serial-1, got %s", status.ClientSerialNumber)
	}
}

// TestFiskalyClient_TestConnection_DeregisteredClient sichert, dass ein
// nicht-REGISTERED-Client kein Transportfehler ist, sondern als Befund im
// VerbindungStatus transportiert wird — die UI meldet ihn dann als Fehler.
func TestFiskalyClient_TestConnection_DeregisteredClient(t *testing.T) {
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
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/tss/tss-1":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"state": "INITIALIZED",
				"_env":  "TEST",
			})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/tss/tss-1/client/client-1":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"state":         "DEREGISTERED",
				"serial_number": "kasse-serial-1",
			})
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

	status, err := client.TestConnection(context.Background())
	if err != nil {
		t.Fatalf("test connection should not fail for a deregistered client: %v", err)
	}
	if status.ClientState != "DEREGISTERED" {
		t.Fatalf("expected client state DEREGISTERED to be reported, got %s", status.ClientState)
	}
}
