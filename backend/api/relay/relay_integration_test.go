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

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nicograef/jotti/backend/app"
	"github.com/nicograef/jotti/backend/config"
	jottiDB "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/jwt"
)

const (
	testJWTSecret  = "test-secret"
	testRelayToken = "test-relay-token-abc123"
)

// pollRequest mirrors the relay poll request DTO.
type pollRequest struct {
	Token       string `json:"token"`
	LastEventID int    `json:"lastEventId"`
}

// pollResponse mirrors the relay poll response DTO.
type pollResponse struct {
	Auftraege []druckAuftragDTO `json:"auftraege"`
	Cursor    int               `json:"cursor"`
}

type druckAuftragDTO struct {
	EventID   int    `json:"eventId"`
	DruckerIP string `json:"druckerIp"`
	Payload   string `json:"payload"`
}

// bestellungRequest mirrors the service bestellung-aufnehmen request DTO.
type bestellungRequest struct {
	TischID    int                    `json:"tischId"`
	Positionen []bestellPositionInput `json:"positionen"`
	Kommentar  string                 `json:"kommentar"`
}

type bestellPositionInput struct {
	ProduktID  int `json:"produktId"`
	VarianteID int `json:"varianteId"`
	Menge      int `json:"menge"`
}

// updateDruckstationenRequest mirrors the admin update-drucker-config request.
type updateDruckstationenRequest struct {
	Kategorie string `json:"kategorie"`
	DruckerIP string `json:"druckerIp"`
	Bonmodus  string `json:"bonmodus"`
}

type kassensitzungEroeffnenRequest struct {
	Bezeichnung string `json:"bezeichnung"`
	BetragCents int    `json:"betragCents"`
}

// seedTestData creates the minimal test data needed for integration tests.
// The migration only creates user ID 1 (admin "nico"), so we need to insert
// betreiber master data, a service user, products with variants, and active tische.
func seedTestData(t *testing.T, db *sql.DB) (serviceUserID, produktID, varianteID, tischID int) {
	t.Helper()

	// Betreiber master data is a precondition for opening a Kassensitzung.
	_, err := db.Exec(`
		INSERT INTO betreiber (id, vereinsname, strasse, plz, ort, updated_at)
		VALUES (1, 'Test-Verein e.V.', 'Teststraße 1', '12345', 'Teststadt', now())
		ON CONFLICT (id) DO UPDATE SET vereinsname = EXCLUDED.vereinsname
	`)
	if err != nil {
		t.Fatalf("Failed to seed betreiber: %v", err)
	}

	// Create a service user for placing orders
	var userID int
	err = db.QueryRow(`
		INSERT INTO users (name, username, password_hash, role, status, created_at, updated_at)
		VALUES ('Test Service', 'test-service', 'unused', 'service', 'active', now(), now())
		ON CONFLICT (username) DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`).Scan(&userID)
	if err != nil {
		t.Fatalf("Failed to create test service user: %v", err)
	}

	// Create a product with category "essen"
	var prodID int
	err = db.QueryRow(`
		INSERT INTO produkte (name, kategorie, status, created_at, updated_at)
		VALUES ('Test-Bratwurst', 'essen', 'active', now(), now())
		ON CONFLICT (name) DO UPDATE SET status = 'active'
		RETURNING id
	`).Scan(&prodID)
	if err != nil {
		t.Fatalf("Failed to create test product: %v", err)
	}

	// Create a variant for the product
	var varID int
	err = db.QueryRow(`
		INSERT INTO produkt_varianten (produkt_id, name, preis_cents, status, created_at, updated_at)
		VALUES ($1, 'Normal', 350, 'active', now(), now())
		RETURNING id
	`, prodID).Scan(&varID)
	if err != nil {
		t.Fatalf("Failed to create test variant: %v", err)
	}

	// Create active tische
	var tID int
	err = db.QueryRow(`
		INSERT INTO tische (name, status, created_at, updated_at)
		VALUES ('Test-Tisch 1', 'active', now(), now())
		ON CONFLICT (name) DO UPDATE SET status = 'active'
		RETURNING id
	`).Scan(&tID)
	if err != nil {
		t.Fatalf("Failed to create test tisch: %v", err)
	}

	// Ensure a second tisch exists for other tests
	_, err = db.Exec(`
		INSERT INTO tische (name, status, created_at, updated_at)
		VALUES ('Test-Tisch 2', 'active', now(), now())
		ON CONFLICT (name) DO UPDATE SET status = 'active'
	`)
	if err != nil {
		t.Fatalf("Failed to create second test tisch: %v", err)
	}

	return userID, prodID, varID, tID
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

	os.Setenv("JWT_SECRET", testJWTSecret)
	os.Setenv("RELAY_AUTH_TOKEN", testRelayToken)

	db := jottiDB.OpenTestDatabase()
	t.Cleanup(func() { db.Close() })

	serviceUserID, produktID, varianteID, tischID := seedTestData(t, db)

	cfg := config.Load()
	handler := app.SetupRoutes(cfg, db)
	ts := httptest.NewServer(handler)
	t.Cleanup(func() { ts.Close() })

	adminTkn, err := jwt.GenerateJWTTokenForUser(1, "Nico Gräf", "admin", testJWTSecret)
	if err != nil {
		t.Fatalf("Failed to generate admin JWT: %v", err)
	}
	svcTkn, err := jwt.GenerateJWTTokenForUser(serviceUserID, "Test Service", "service", testJWTSecret)
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
	defer resp.Body.Close()
	// 200 = newly opened, 400+kasse_bereits_geoeffnet = already open — both are fine for tests
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
	defer resp.Body.Close()
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
	defer resp.Body.Close()
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
		TischID: env.tischID,
		Positionen: []bestellPositionInput{
			{ProduktID: env.produktID, VarianteID: env.varianteID, Menge: 2},
		},
		Kommentar: kommentar,
	}, env.svcToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var errResp map[string]any
		json.NewDecoder(resp.Body).Decode(&errResp)
		t.Fatalf("Failed to create bestellung: status %d, body %v", resp.StatusCode, errResp)
	}
}

func pollRelay(t *testing.T, serverURL string, relayToken string, lastEventID int) *http.Response {
	t.Helper()
	return postJSON(t, serverURL+"/relay/poll", pollRequest{
		Token:       relayToken,
		LastEventID: lastEventID,
	}, "")
}

// TestRelayPoll_BestellungToAuftraege tests the full flow:
// configure druckstation → create bestellung → poll relay → receive correct Druck-Aufträge
func TestRelayPoll_BestellungToAuftraege(t *testing.T) {
	env := setupTestEnv(t)

	// Reset druckstationen to clean state
	resetDruckstationen(t, env.server.URL, env.adminToken)

	// Configure a druckstation for "essen" category
	configureDruckstation(t, env.server.URL, env.adminToken, "essen", "192.168.1.51", "pro_position")

	// Get the current cursor (before our bestellung)
	resp := pollRelay(t, env.server.URL, testRelayToken, 0)
	before := decodePollResponse(t, resp)
	cursorBefore := before.Cursor

	// Create a bestellung with an "essen" product
	createBestellung(t, env, "ohne Senf")

	// Poll relay from before cursor → should receive Druck-Aufträge
	resp = pollRelay(t, env.server.URL, testRelayToken, cursorBefore)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}
	result := decodePollResponse(t, resp)

	if len(result.Auftraege) == 0 {
		t.Fatal("Expected at least 1 Druck-Auftrag, got 0")
	}

	// Verify the Druck-Auftrag structure
	found := false
	for _, a := range result.Auftraege {
		if a.DruckerIP == "192.168.1.51" {
			found = true
			if a.EventID == 0 {
				t.Error("EventID should not be 0")
			}
			if a.Payload == "" {
				t.Error("Payload should not be empty")
			}
		}
	}
	if !found {
		t.Error("Expected at least one Auftrag with DruckerIP 192.168.1.51")
	}

	// Cursor should have advanced
	if result.Cursor <= cursorBefore {
		t.Errorf("Cursor should advance: was %d, now %d", cursorBefore, result.Cursor)
	}
}

// TestRelayPoll_CursorFortschritt tests that a second poll with the returned cursor
// does not return duplicate events.
func TestRelayPoll_CursorFortschritt(t *testing.T) {
	env := setupTestEnv(t)

	resetDruckstationen(t, env.server.URL, env.adminToken)
	configureDruckstation(t, env.server.URL, env.adminToken, "essen", "192.168.1.51", "pro_position")

	// Poll to get current cursor
	resp := pollRelay(t, env.server.URL, testRelayToken, 0)
	initial := decodePollResponse(t, resp)

	// Create a bestellung
	createBestellung(t, env, "")

	// First poll: should get new auftraege
	resp = pollRelay(t, env.server.URL, testRelayToken, initial.Cursor)
	first := decodePollResponse(t, resp)
	if len(first.Auftraege) == 0 {
		t.Fatal("First poll should return auftraege")
	}

	// Second poll with advanced cursor: no new events → empty
	resp = pollRelay(t, env.server.URL, testRelayToken, first.Cursor)
	second := decodePollResponse(t, resp)

	if len(second.Auftraege) != 0 {
		t.Errorf("Second poll with advanced cursor should return 0 auftraege, got %d", len(second.Auftraege))
	}

	// Cursor should remain the same
	if second.Cursor != first.Cursor {
		t.Errorf("Cursor should stay at %d, got %d", first.Cursor, second.Cursor)
	}
}

// TestRelayPoll_DruckerConfigAenderung tests that changing drucker config
// affects subsequent polls (IPs resolved at read time).
func TestRelayPoll_DruckerConfigAenderung(t *testing.T) {
	env := setupTestEnv(t)

	resetDruckstationen(t, env.server.URL, env.adminToken)

	// Configure essen drucker with IP .51
	configureDruckstation(t, env.server.URL, env.adminToken, "essen", "192.168.1.51", "pro_position")

	// Record cursor before bestellung
	resp := pollRelay(t, env.server.URL, testRelayToken, 0)
	before := decodePollResponse(t, resp)

	// Create a bestellung
	createBestellung(t, env, "")

	// Change drucker IP to .52 BEFORE polling
	configureDruckstation(t, env.server.URL, env.adminToken, "essen", "192.168.1.52", "pro_position")

	// Poll → should get auftraege with the NEW IP (.52)
	resp = pollRelay(t, env.server.URL, testRelayToken, before.Cursor)
	result := decodePollResponse(t, resp)

	if len(result.Auftraege) == 0 {
		t.Fatal("Expected auftraege after config change")
	}

	for _, a := range result.Auftraege {
		if a.DruckerIP != "192.168.1.52" {
			t.Errorf("Expected DruckerIP 192.168.1.52, got %s", a.DruckerIP)
		}
	}
}

// TestRelayPoll_KeinDruckerKonfiguriert tests that when no drucker is configured
// (empty drucker_ip), the poll returns an empty auftraege list.
func TestRelayPoll_KeinDruckerKonfiguriert(t *testing.T) {
	env := setupTestEnv(t)

	// Reset all drucker configs to empty
	resetDruckstationen(t, env.server.URL, env.adminToken)

	// Record cursor
	resp := pollRelay(t, env.server.URL, testRelayToken, 0)
	before := decodePollResponse(t, resp)

	// Create a bestellung
	createBestellung(t, env, "")

	// Poll → no drucker configured → empty auftraege, but cursor should still advance
	resp = pollRelay(t, env.server.URL, testRelayToken, before.Cursor)
	result := decodePollResponse(t, resp)

	if len(result.Auftraege) != 0 {
		t.Errorf("Expected 0 auftraege when no drucker configured, got %d", len(result.Auftraege))
	}

	// Cursor should still advance (events were processed, just no drucker matched)
	if result.Cursor <= before.Cursor {
		t.Errorf("Cursor should advance even without drucker config: was %d, now %d", before.Cursor, result.Cursor)
	}
}

// TestRelayPoll_FalscherToken tests that a wrong relay token returns an error.
func TestRelayPoll_FalscherToken(t *testing.T) {
	env := setupTestEnv(t)

	resp := pollRelay(t, env.server.URL, "wrong-token", 0)
	defer resp.Body.Close()

	// The handler returns 400 with code "unauthorized"
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400 for wrong token, got %d", resp.StatusCode)
	}

	var errResp struct {
		Code string `json:"code"`
	}
	json.NewDecoder(resp.Body).Decode(&errResp)
	if errResp.Code != "unauthorized" {
		t.Errorf("Expected error code 'unauthorized', got %q", errResp.Code)
	}
}
