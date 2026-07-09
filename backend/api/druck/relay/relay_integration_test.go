//go:build integration

package relay_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nicograef/jotti/backend/app"
	"github.com/nicograef/jotti/backend/config"
	jottiDB "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/jwt"
)

const (
	testJWTSecret  = "test-jwt-secret-abc123456789"
	testRelayToken = "test-relay-token-abc123"
	testDBPassword = "test-postgres-password-1234" //nolint:gosec // Test-Platzhalter, kein echtes Secret
)

type pollRequest struct {
	Token string `json:"token"`
}

type pollResponse struct {
	Auftraege []druckAuftragDTO `json:"auftraege"`
}

type druckAuftragDTO struct {
	ID      int    `json:"id"`
	ZielIP  string `json:"zielIp"`
	Payload string `json:"payload"`
}

type ergebnisRequest struct {
	Token        string `json:"token"`
	GedruckteIDs []int  `json:"gedruckteIds"`
}

type bestellungRequest struct {
	BestellungID string                 `json:"bestellungId"`
	TischID      int                    `json:"tischId"`
	Positionen   []bestellPositionInput `json:"positionen"`
	Kommentar    string                 `json:"kommentar"`
}

type bestellPositionInput struct {
	ProduktID  int `json:"produktId"`
	VarianteID int `json:"varianteId"`
	Menge      int `json:"menge"`
}

type updateDruckstationenRequest struct {
	Kategorie string `json:"kategorie"`
	DruckerIP string `json:"druckerIp"`
	Bonmodus  string `json:"bonmodus"`
}

type kassensitzungEroeffnenRequest struct {
	Bezeichnung string `json:"bezeichnung"`
	BetragCents int    `json:"betragCents"`
}

func seedTestData(t *testing.T, db *sql.DB) (adminUserID, serviceUserID, produktID, varianteID, tischID int) {
	t.Helper()

	_, err := db.Exec(`
		INSERT INTO betreiber (id, vereinsname, strasse, plz, ort, updated_at)
		VALUES (1, 'Test-Verein e.V.', 'Teststrasse 1', '12345', 'Teststadt', now())
		ON CONFLICT (id) DO UPDATE SET vereinsname = EXCLUDED.vereinsname
	`)
	if err != nil {
		t.Fatalf("Failed to seed betreiber: %v", err)
	}

	_, err = db.Exec("DELETE FROM druckauftraege")
	if err != nil {
		t.Fatalf("Failed to reset druckauftraege: %v", err)
	}

	// Der Admin muss real in der users-Tabelle existieren: Der Event-Write hat einen
	// FK auf users(id), ein JWT mit hartkodierter ID schlägt fehl, sobald andere
	// Test-Suiten die users-Tabelle geleert und die ID-Sequenz weitergedreht haben.
	var adminID int
	err = db.QueryRow(`
		INSERT INTO users (name, username, password_hash, role, status, created_at, updated_at)
		VALUES ('Test Admin', 'test-admin', 'unused', 'admin', 'active', now(), now())
		ON CONFLICT (username) WHERE status != 'deleted' DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`).Scan(&adminID)
	if err != nil {
		t.Fatalf("Failed to create test admin user: %v", err)
	}

	var userID int
	err = db.QueryRow(`
		INSERT INTO users (name, username, password_hash, role, status, created_at, updated_at)
		VALUES ('Test Service', 'test-service', 'unused', 'service', 'active', now(), now())
		ON CONFLICT (username) WHERE status != 'deleted' DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`).Scan(&userID)
	if err != nil {
		t.Fatalf("Failed to create test service user: %v", err)
	}

	var prodID int
	err = db.QueryRow(`
		INSERT INTO produkte (name, kategorie, steuersatz, status, created_at, updated_at)
		VALUES ('Test-Bratwurst', 'essen', 'regel', 'active', now(), now())
		ON CONFLICT (name) WHERE status != 'deleted' DO UPDATE SET status = 'active', steuersatz = 'regel'
		RETURNING id
	`).Scan(&prodID)
	if err != nil {
		t.Fatalf("Failed to create test product: %v", err)
	}

	var varID int
	err = db.QueryRow(`
		INSERT INTO produkt_varianten (produkt_id, name, preis_cents, status, created_at, updated_at)
		VALUES ($1, 'Normal', 350, 'active', now(), now())
		RETURNING id
	`, prodID).Scan(&varID)
	if err != nil {
		t.Fatalf("Failed to create test variant: %v", err)
	}

	var tID int
	err = db.QueryRow(`
		INSERT INTO tische (name, status, created_at, updated_at)
		VALUES ('Test-Tisch 1', 'active', now(), now())
		ON CONFLICT (name) WHERE status != 'deleted' DO UPDATE SET status = 'active'
		RETURNING id
	`).Scan(&tID)
	if err != nil {
		t.Fatalf("Failed to create test tisch: %v", err)
	}

	return adminID, userID, prodID, varID, tID
}

type testEnv struct {
	server     *httptest.Server
	adminToken string
	svcToken   string
	produktID  int
	varianteID int
	tischID    int
}

func setupTestEnv(t *testing.T) testEnv {
	t.Helper()

	t.Setenv("JWT_SECRET", testJWTSecret)
	t.Setenv("RELAY_AUTH_TOKEN", testRelayToken)

	db := jottiDB.OpenTestDatabase()
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("Testdatenbank schließen: %v", err)
		}
	})

	adminUserID, serviceUserID, produktID, varianteID, tischID := seedTestData(t, db)

	// The throwaway DB uses password "admin", which config.Load now rejects as a
	// known placeholder. The DB handle is already open (and passed to SetupRoutes),
	// so override the env with a valid value purely to satisfy config validation,
	// then restore it so later tests' OpenTestDatabase still reaches the DB.
	origPW := os.Getenv("POSTGRES_PASSWORD")
	if err := os.Setenv("POSTGRES_PASSWORD", testDBPassword); err != nil {
		t.Fatalf("POSTGRES_PASSWORD setzen: %v", err)
	}
	cfg := config.Load()
	if err := os.Setenv("POSTGRES_PASSWORD", origPW); err != nil {
		t.Fatalf("POSTGRES_PASSWORD zurücksetzen: %v", err)
	}
	handler := app.SetupRoutes(cfg, db, "dev")
	ts := httptest.NewServer(handler)
	t.Cleanup(func() { ts.Close() })

	adminTkn, err := jwt.GenerateJWTTokenForUser(adminUserID, "test-admin", "admin", testJWTSecret)
	if err != nil {
		t.Fatalf("Failed to generate admin JWT: %v", err)
	}
	svcTkn, err := jwt.GenerateJWTTokenForUser(serviceUserID, "test-service", "service", testJWTSecret)
	if err != nil {
		t.Fatalf("Failed to generate service JWT: %v", err)
	}

	env := testEnv{
		server:     ts,
		adminToken: adminTkn,
		svcToken:   svcTkn,
		produktID:  produktID,
		varianteID: varianteID,
		tischID:    tischID,
	}

	openKassensitzungIfNeeded(t, env)
	return env
}

func openKassensitzungIfNeeded(t *testing.T, env testEnv) {
	t.Helper()
	resp := postJSON(t, env.server.URL+"/admin/kassensitzung-eroeffnen", kassensitzungEroeffnenRequest{
		Bezeichnung: "Integration Test",
		BetragCents: 10000,
	}, env.adminToken)
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Code string `json:"code"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil || errResp.Code != "kasse_bereits_geoeffnet" {
			t.Fatalf("openKassensitzungIfNeeded: unexpected status %d (code=%q)", resp.StatusCode, errResp.Code)
		}
	}
}

func postJSON(t *testing.T, url string, body any, authToken string) *http.Response {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	return resp
}

func decodePollResponse(t *testing.T, resp *http.Response) pollResponse {
	t.Helper()
	defer resp.Body.Close() //nolint:errcheck // Testpfad: Body-Close-Fehler ohne Belang
	var result pollResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode poll response: %v", err)
	}
	return result
}

func configureDruckstation(t *testing.T, serverURL, token, kategorie, ip, bonmodus string) {
	t.Helper()
	resp := postJSON(t, serverURL+"/admin/update-druckstationen", updateDruckstationenRequest{
		Kategorie: kategorie,
		DruckerIP: ip,
		Bonmodus:  bonmodus,
	}, token)
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to configure druckstation for %s: status %d", kategorie, resp.StatusCode)
	}
}

func resetDruckstationen(t *testing.T, serverURL, token string) {
	t.Helper()
	for _, kat := range []string{"essen", "getraenk", "sonstiges"} {
		configureDruckstation(t, serverURL, token, kat, "", "pro_position")
	}
}

func createBestellung(t *testing.T, env testEnv, kommentar string) {
	t.Helper()
	resp := postJSON(t, env.server.URL+"/service/bestellung-aufnehmen", bestellungRequest{
		BestellungID: uuid.NewString(),
		TischID:      env.tischID,
		Positionen: []bestellPositionInput{
			{ProduktID: env.produktID, VarianteID: env.varianteID, Menge: 2},
		},
		Kommentar: kommentar,
	}, env.svcToken)
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		var errResp map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		t.Fatalf("Failed to create bestellung: status %d, body %v", resp.StatusCode, errResp)
	}
}

func pollRelay(t *testing.T, serverURL string, relayToken string) *http.Response {
	t.Helper()
	return postJSON(t, serverURL+"/relay/poll", pollRequest{Token: relayToken}, "")
}

func reportErgebnis(t *testing.T, serverURL string, relayToken string, ids []int) *http.Response {
	t.Helper()
	return postJSON(t, serverURL+"/relay/ergebnis", ergebnisRequest{
		Token:        relayToken,
		GedruckteIDs: ids,
	}, "")
}

func TestRelayPollErgebnisFlow(t *testing.T) {
	env := setupTestEnv(t)

	resetDruckstationen(t, env.server.URL, env.adminToken)
	configureDruckstation(t, env.server.URL, env.adminToken, "essen", "192.168.1.51", "pro_position")

	createBestellung(t, env, "ohne Senf")

	resp := pollRelay(t, env.server.URL, testRelayToken)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}
	result := decodePollResponse(t, resp)

	if len(result.Auftraege) == 0 {
		t.Fatal("Expected at least 1 Druck-Auftrag, got 0")
	}

	ids := make([]int, 0, len(result.Auftraege))
	for _, auftrag := range result.Auftraege {
		if auftrag.ZielIP != "192.168.1.51" {
			t.Fatalf("Expected ZielIP 192.168.1.51, got %s", auftrag.ZielIP)
		}
		if auftrag.Payload == "" {
			t.Fatal("Payload should not be empty")
		}
		if auftrag.ID == 0 {
			t.Fatal("Auftrag ID should not be 0")
		}
		ids = append(ids, auftrag.ID)
	}

	resp = reportErgebnis(t, env.server.URL, testRelayToken, ids)
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected ergebnis status 200, got %d", resp.StatusCode)
	}

	resp = pollRelay(t, env.server.URL, testRelayToken)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected second poll status 200, got %d", resp.StatusCode)
	}
	after := decodePollResponse(t, resp)
	if len(after.Auftraege) != 0 {
		t.Fatalf("Expected no offene Auftraege after ergebnis, got %d", len(after.Auftraege))
	}
}

func TestRelayPollKeinDruckerKonfiguriert(t *testing.T) {
	env := setupTestEnv(t)

	resetDruckstationen(t, env.server.URL, env.adminToken)
	createBestellung(t, env, "")

	resp := pollRelay(t, env.server.URL, testRelayToken)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}
	result := decodePollResponse(t, resp)

	if len(result.Auftraege) != 0 {
		t.Fatalf("Expected 0 auftraege when no drucker configured, got %d", len(result.Auftraege))
	}
}

func TestRelayPollFalscherToken(t *testing.T) {
	env := setupTestEnv(t)

	resp := pollRelay(t, env.server.URL, "wrong-token")
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected 400 for wrong token, got %d", resp.StatusCode)
	}

	var errResp struct {
		Code string `json:"code"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&errResp)
	if errResp.Code != "unauthorized" {
		t.Fatalf("Expected error code 'unauthorized', got %q", errResp.Code)
	}
}
