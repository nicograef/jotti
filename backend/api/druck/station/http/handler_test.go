package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nicograef/jotti/backend/api/druck/station/application"
)

type mockDruckstationCommand struct {
	testbonErr        error
	upserts           []string
	testbonKategorien []string
}

func (m *mockDruckstationCommand) UpsertDruckstation(_ context.Context, kategorie, _, bonmodus string) error {
	m.upserts = append(m.upserts, kategorie+"/"+bonmodus)
	return nil
}

func (m *mockDruckstationCommand) TestbonDrucken(_ context.Context, kategorie string) error {
	m.testbonKategorien = append(m.testbonKategorien, kategorie)
	return m.testbonErr
}

// Which Bonmodus a station category takes is the handler's decision: Produktstationen
// and Abholbon need one, the Kassenbeleg none, pro_stueck only the Abholbon. An
// accepted row must reach the command, a rejected one must not.
func TestUpdateDruckstationenHandler_KategorieUndBonmodus(t *testing.T) {
	tests := []struct {
		kategorie, druckerIP, bonmodus string
		wantStatus                     int
	}{
		{"essen", "192.168.1.51", "pro_position", http.StatusOK},
		{"getraenk", "", "pro_bestellung", http.StatusOK},
		{"kassenbeleg", "192.168.1.60", "", http.StatusOK},
		{"abholbon", "192.168.1.70", "pro_bestellung", http.StatusOK},
		{"abholbon", "192.168.1.70", "pro_stueck", http.StatusOK},
		{"invalid", "192.168.1.51", "pro_position", http.StatusBadRequest},
		{"essen", "192.168.1.51", "invalid", http.StatusBadRequest},
		{"essen", "999.999.999.999", "pro_position", http.StatusBadRequest},
		{"essen", "192.168.1.51", "pro_stueck", http.StatusBadRequest},
		{"getraenk", "192.168.1.51", "pro_stueck", http.StatusBadRequest},
		{"sonstiges", "192.168.1.51", "pro_stueck", http.StatusBadRequest},
		{"kassenbeleg", "", "pro_position", http.StatusBadRequest},
		{"essen", "192.168.1.51", "", http.StatusBadRequest},
		{"abholbon", "192.168.1.51", "", http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.kategorie+"/"+tt.druckerIP+"/"+tt.bonmodus, func(t *testing.T) {
			cmd := &mockDruckstationCommand{}
			body := `{"kategorie":"` + tt.kategorie + `","druckerIp":"` + tt.druckerIP + `","bonmodus":"` + tt.bonmodus + `"}`
			req := httptest.NewRequest(http.MethodPost, "/admin/update-druckstationen", strings.NewReader(body))
			rec := httptest.NewRecorder()

			(&CommandHandler{Command: cmd}).UpdateDruckstationenHandler().ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}
			wantUpserts := 0
			if tt.wantStatus == http.StatusOK {
				wantUpserts = 1
			}
			if len(cmd.upserts) != wantUpserts {
				t.Errorf("expected %d upserts, got %v", wantUpserts, cmd.upserts)
			}
		})
	}
}

func TestTestbonDruckenHandler_Success(t *testing.T) {
	cmd := &mockDruckstationCommand{}
	handler := &CommandHandler{Command: cmd}

	body := `{"kategorie":"essen"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/testbon-drucken", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.TestbonDruckenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(cmd.testbonKategorien) != 1 || cmd.testbonKategorien[0] != "essen" {
		t.Errorf("expected TestbonDrucken called with 'essen', got %v", cmd.testbonKategorien)
	}
}

func TestTestbonDruckenHandler_InvalidKategorie(t *testing.T) {
	cmd := &mockDruckstationCommand{}
	handler := &CommandHandler{Command: cmd}

	body := `{"kategorie":"quatsch"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/testbon-drucken", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.TestbonDruckenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid kategorie, got %d", rec.Code)
	}
	if len(cmd.testbonKategorien) != 0 {
		t.Errorf("expected command not called on validation error, got %v", cmd.testbonKategorien)
	}
}

func TestTestbonDruckenHandler_NichtKonfiguriert(t *testing.T) {
	cmd := &mockDruckstationCommand{testbonErr: application.ErrDruckstationNichtKonfiguriert}
	handler := &CommandHandler{Command: cmd}

	body := `{"kategorie":"essen"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/testbon-drucken", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.TestbonDruckenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for nicht konfiguriert, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "druckstation_nicht_konfiguriert") {
		t.Errorf("expected error code druckstation_nicht_konfiguriert, got %s", rec.Body.String())
	}
}
