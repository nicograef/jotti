//go:build unit

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nicograef/jotti/backend/api/druck/station/application"
	"github.com/nicograef/jotti/backend/domain/druckstation"
)

type mockDruckstationCommand struct {
	err               error
	testbonErr        error
	testbonKategorien []string
}

func (m *mockDruckstationCommand) UpsertDruckstation(ctx context.Context, kategorie, druckerIP, bonmodus string) error {
	return m.err
}

func (m *mockDruckstationCommand) TestbonDrucken(ctx context.Context, kategorie string) error {
	m.testbonKategorien = append(m.testbonKategorien, kategorie)
	return m.testbonErr
}

type mockDruckstationQuery struct {
	result []druckstation.Druckstation
	err    error
}

func (m *mockDruckstationQuery) GetAlleDruckstationen(ctx context.Context) ([]druckstation.Druckstation, error) {
	return m.result, m.err
}

func TestUpdateDruckstationenHandler_Success(t *testing.T) {
	handler := &CommandHandler{Command: &mockDruckstationCommand{}}

	body := `{"kategorie":"essen","druckerIp":"192.168.1.51","bonmodus":"pro_position"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/update-druckstationen", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateDruckstationenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateDruckstationenHandler_EmptyIPAllowed(t *testing.T) {
	handler := &CommandHandler{Command: &mockDruckstationCommand{}}

	body := `{"kategorie":"getraenk","druckerIp":"","bonmodus":"pro_bestellung"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/update-druckstationen", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateDruckstationenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200 for empty IP, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateDruckstationenHandler_InvalidKategorie(t *testing.T) {
	handler := &CommandHandler{Command: &mockDruckstationCommand{}}

	body := `{"kategorie":"invalid","druckerIp":"192.168.1.51","bonmodus":"pro_position"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/update-druckstationen", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateDruckstationenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid kategorie, got %d", rec.Code)
	}
}

func TestUpdateDruckstationenHandler_InvalidBonmodus(t *testing.T) {
	handler := &CommandHandler{Command: &mockDruckstationCommand{}}

	body := `{"kategorie":"essen","druckerIp":"192.168.1.51","bonmodus":"invalid"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/update-druckstationen", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateDruckstationenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid bonmodus, got %d", rec.Code)
	}
}

func TestUpdateDruckstationenHandler_InvalidIP(t *testing.T) {
	handler := &CommandHandler{Command: &mockDruckstationCommand{}}

	body := `{"kategorie":"essen","druckerIp":"999.999.999.999","bonmodus":"pro_position"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/update-druckstationen", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateDruckstationenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid IP, got %d", rec.Code)
	}
}

func TestUpdateDruckstationenHandler_Kassenbeleg(t *testing.T) {
	handler := &CommandHandler{Command: &mockDruckstationCommand{}}

	body := `{"kategorie":"kassenbeleg","druckerIp":"192.168.1.60","bonmodus":""}`
	req := httptest.NewRequest(http.MethodPost, "/admin/update-druckstationen", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateDruckstationenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200 for kassenbeleg, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateDruckstationenHandler_Abholbon(t *testing.T) {
	handler := &CommandHandler{Command: &mockDruckstationCommand{}}

	body := `{"kategorie":"abholbon","druckerIp":"192.168.1.70","bonmodus":"pro_bestellung"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/update-druckstationen", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateDruckstationenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200 for abholbon with bonmodus, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateDruckstationenHandler_BonmodusFuerKassenbelegAbgelehnt(t *testing.T) {
	handler := &CommandHandler{Command: &mockDruckstationCommand{}}

	body := `{"kategorie":"kassenbeleg","druckerIp":"","bonmodus":"pro_position"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/update-druckstationen", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateDruckstationenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for bonmodus on kassenbeleg, got %d", rec.Code)
	}
}

func TestUpdateDruckstationenHandler_StationOhneBonmodusAbgelehnt(t *testing.T) {
	handler := &CommandHandler{Command: &mockDruckstationCommand{}}

	for _, kategorie := range []string{"essen", "abholbon"} {
		body := `{"kategorie":"` + kategorie + `","druckerIp":"192.168.1.51","bonmodus":""}`
		req := httptest.NewRequest(http.MethodPost, "/admin/update-druckstationen", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.UpdateDruckstationenHandler().ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400 for %s without bonmodus, got %d", kategorie, rec.Code)
		}
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

func TestGetDruckstationenHandler_Success(t *testing.T) {
	konfigs := []druckstation.Druckstation{
		{Kategorie: "essen", DruckerIP: "192.168.1.51", Bonmodus: "pro_position"},
		{Kategorie: "getraenk", DruckerIP: "", Bonmodus: "pro_position"},
		{Kategorie: "sonstiges", DruckerIP: "", Bonmodus: "pro_position"},
	}
	handler := &QueryHandler{Query: &mockDruckstationQuery{result: konfigs}}

	req := httptest.NewRequest(http.MethodPost, "/admin/get-druckstationen", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.GetDruckstationenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}
