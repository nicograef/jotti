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

// TestFiskalySetup_LiveVollerDurchlauf richtet aus einem fiskaly-TEST-Konto eine
// vollständige, signierfähige TSS samt Client ein und signiert anschließend eine
// Transaktion über die frisch angelegte Konfiguration. Der Test läuft nur, wenn
// FISKALY_TEST_API_KEY/SECRET gesetzt sind.
//
// ACHTUNG: Jeder Lauf legt im TEST-Konto eine nicht löschbare TSS an. Bewusst
// sparsam ausführen (siehe Plan, „TEST-Konto füllt sich").
//
//	FISKALY_TEST_API_KEY=... FISKALY_TEST_API_SECRET=... \
//	go test -tags=integration -run LiveVollerDurchlauf ./repository/tse_repo/
func TestFiskalySetup_LiveVollerDurchlauf(t *testing.T) {
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
	if err := setupClient.SetzeAdminPIN(ctx, erstellt.ID, erstellt.PUK, pin); err != nil {
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

	// Die frisch eingerichtete TSS muss signierfähig sein.
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
		t.Fatalf("expected a REGISTERED client, got %q", status.ClientState)
	}
	if status.ClientSerialNumber != seriennummer {
		t.Fatalf("expected client serial %q, got %q", seriennummer, status.ClientSerialNumber)
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
		t.Fatal("expected a signature from the freshly set up TSS")
	}
	if !strings.HasPrefix(finish.QRCodeData, "V0;") {
		t.Fatalf("expected qr_code_data with prefix V0;, got %q", finish.QRCodeData)
	}
}
