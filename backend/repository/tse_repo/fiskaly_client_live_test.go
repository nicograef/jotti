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

// TestFiskalyClient_LiveSigniertTransaktion signiert eine echte Transaktion
// gegen die fiskaly-TEST-Umgebung. Der Test läuft nur, wenn die
// FISKALY_TEST_*-Umgebungsvariablen gesetzt sind, und wird sonst übersprungen.
//
//	FISKALY_TEST_API_KEY=...    FISKALY_TEST_API_SECRET=... \
//	FISKALY_TEST_TSS_ID=...     FISKALY_TEST_CLIENT_ID=... \
//	go test -tags=integration -run LiveSigniert ./repository/tse_repo/
func TestFiskalyClient_LiveSigniertTransaktion(t *testing.T) {
	credentials := tse.Credentials{
		ApiKey:    os.Getenv("FISKALY_TEST_API_KEY"),
		ApiSecret: os.Getenv("FISKALY_TEST_API_SECRET"),
		TssID:     os.Getenv("FISKALY_TEST_TSS_ID"),
		ClientID:  os.Getenv("FISKALY_TEST_CLIENT_ID"),
	}
	if credentials.Validate() != nil {
		t.Skip("FISKALY_TEST_* nicht gesetzt — Live-Test übersprungen")
	}

	baseURL := os.Getenv("FISKALY_BASE_URL")
	if baseURL == "" {
		baseURL = "https://kassensichv-middleware.fiskaly.com"
	}

	client, err := NewFiskalyTSEClient(baseURL, credentials, nil)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	status, err := client.TestConnection(ctx)
	if err != nil {
		t.Fatalf("test connection failed: %v", err)
	}
	if status.Umgebung != tse.UmgebungTest {
		t.Fatalf("Live-Test nur gegen die TEST-Umgebung erlaubt, Credentials zeigen auf %s", status.Umgebung)
	}

	txID := uuid.NewString()

	start, err := client.StartTransaction(ctx, txID)
	if err != nil {
		t.Fatalf("start transaction failed: %v", err)
	}
	if start.TransactionNumber == 0 {
		t.Fatalf("expected a transaction number, got 0")
	}

	finish, err := client.FinishTransaction(ctx, txID, "Kassenbeleg-V1", "Beleg^0.00_2.55_0.00_0.00_0.00^2.55:Bar")
	if err != nil {
		t.Fatalf("finish transaction failed: %v", err)
	}
	if finish.Signature == "" {
		t.Fatalf("expected a signature, got empty string")
	}
	if !strings.HasPrefix(finish.QRCodeData, "V0;") {
		t.Fatalf("expected qr_code_data with prefix V0;, got %q", finish.QRCodeData)
	}
}
