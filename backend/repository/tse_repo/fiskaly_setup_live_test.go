//go:build integration

package tse_repo

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/domain/tse"
)

// TestFiskalySetup_LiveVollerDurchlauf sets up a signing-ready TSS with client in a fiskaly TEST account and signs one transaction.
// Each run creates an undeletable TSS and the TEST account allows only five active ones, so run it sparingly.
// Run: JOTTI_TSE_LIVE=1 with FISKALY_TEST_API_KEY/SECRET via make test-tse-live-setup (loads .env.fiskaly-test).
func TestFiskalySetup_LiveVollerDurchlauf(t *testing.T) {
	if os.Getenv("JOTTI_TSE_LIVE") != "1" {
		t.Skip("JOTTI_TSE_LIVE != 1 — Setup-Live-Test übersprungen (Opt-in via make test-tse-live-setup)")
	}
	apiKey := os.Getenv("FISKALY_TEST_API_KEY")
	apiSecret := os.Getenv("FISKALY_TEST_API_SECRET")
	if apiKey == "" || apiSecret == "" {
		t.Skip("FISKALY_TEST_API_KEY/SECRET nicht gesetzt — Setup-Live-Test übersprungen")
	}

	baseURL := os.Getenv("FISKALY_BASE_URL")
	if baseURL == "" {
		baseURL = "https://kassensichv-middleware.fiskaly.com"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	setupClient, err := NewFiskalyTSESetupClient(baseURL, tse.SetupCredentials{ApiKey: apiKey, ApiSecret: apiSecret}, nil)
	if err != nil {
		t.Fatalf("failed to create setup client: %v", err)
	}

	umgebung, _, err := setupClient.ListTSS(ctx)
	if err != nil {
		t.Fatalf("list tss failed: %v", err)
	}
	if umgebung != tse.UmgebungTest {
		t.Fatalf("Live-Test nur gegen die TEST-Umgebung erlaubt, Konto zeigt auf %s", umgebung)
	}

	seriennummer := uuid.NewString()
	clientID := uuid.NewString()
	pin := "1234567890"

	erstellt, err := setupClient.CreateTSS(ctx)
	if err != nil {
		t.Fatalf("create tss failed: %v", err)
	}
	if erstellt.ID == "" || erstellt.PUK == "" {
		t.Fatalf("expected tss id and puk, got %+v", erstellt)
	}
	if err := setupClient.PersonalisiereTSS(ctx, erstellt.ID); err != nil {
		t.Fatalf("personalize failed: %v", err)
	}
	if err := setupClient.SetAdminPIN(ctx, erstellt.ID, erstellt.PUK, pin); err != nil {
		t.Fatalf("set admin pin failed: %v", err)
	}
	if err := setupClient.AuthentifiziereAdmin(ctx, erstellt.ID, pin); err != nil {
		t.Fatalf("admin auth (init) failed: %v", err)
	}
	if err := setupClient.InitialisiereTSS(ctx, erstellt.ID); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}
	if err := setupClient.AuthentifiziereAdmin(ctx, erstellt.ID, pin); err != nil {
		t.Fatalf("admin auth (client) failed: %v", err)
	}
	if err := setupClient.RegistriereClient(ctx, erstellt.ID, clientID, seriennummer); err != nil {
		t.Fatalf("register client failed: %v", err)
	}

	// The TSS master data for the DSFinV-K export must be readable.
	stammdaten, err := setupClient.RetrieveTSSStammdaten(ctx, erstellt.ID)
	if err != nil {
		t.Fatalf("retrieve tss stammdaten failed: %v", err)
	}
	if stammdaten.SignaturAlgorithmus == "" || stammdaten.PublicKey == "" || stammdaten.Zertifikat == "" {
		t.Errorf("expected non-empty algorithm, public key and certificate, got %+v", stammdaten)
	}
	if stammdaten.LogTimeFormat == "" {
		t.Errorf("expected a log time format, got %+v", stammdaten)
	}
	if stammdaten.Seriennummer == "" {
		t.Errorf("expected the tss serial_number, got %+v", stammdaten)
	}

	// The freshly set up TSS must be able to sign.
	signClient, err := NewFiskalyTSEClient(baseURL, tse.Credentials{
		ApiKey:    apiKey,
		ApiSecret: apiSecret,
		TssID:     erstellt.ID,
		ClientID:  clientID,
	}, nil)
	if err != nil {
		t.Fatalf("failed to create sign client: %v", err)
	}

	status, err := signClient.TestConnection(ctx)
	if err != nil {
		t.Fatalf("test connection on fresh TSS failed: %v", err)
	}
	if status.ClientState != "REGISTERED" {
		t.Errorf("expected a REGISTERED client, got %q", status.ClientState)
	}
	if status.ClientSerialNumber != seriennummer {
		t.Errorf("expected client serial %q, got %q", seriennummer, status.ClientSerialNumber)
	}

	txID := uuid.NewString()
	if _, err := signClient.StartTransaction(ctx, txID); err != nil {
		t.Fatalf("start transaction failed: %v", err)
	}
	finish, err := signClient.FinishTransaction(ctx, txID, "Kassenbeleg-V1", "Beleg^0.00_2.55_0.00_0.00_0.00^2.55:Bar")
	if err != nil {
		t.Fatalf("finish transaction failed: %v", err)
	}
	if finish.Signature == "" {
		t.Error("expected a signature from the freshly set up TSS")
	}
	if !strings.HasPrefix(finish.QRCodeData, "V0;") {
		t.Errorf("expected qr_code_data with prefix V0;, got %q", finish.QRCodeData)
	}
}
