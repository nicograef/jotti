//go:build unit

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nicograef/jotti/backend/repository/druckstation_repo"
)

type mockDruckstationCommand struct {
	err error
}

func (m *mockDruckstationCommand) UpsertDruckstation(ctx context.Context, kategorie, druckerIP, bonmodus string) error {
	return m.err
}

type mockDruckstationQuery struct {
	result []druckstation_repo.Druckstation
	err    error
}

func (m *mockDruckstationQuery) GetAlleDruckstationen(ctx context.Context) ([]druckstation_repo.Druckstation, error) {
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

	body := `{"kategorie":"essen","druckerIp":"not-an-ip","bonmodus":"pro_position"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/update-druckstationen", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateDruckstationenHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid IP, got %d", rec.Code)
	}
}

func TestGetDruckstationenHandler_Success(t *testing.T) {
	konfigs := []druckstation_repo.Druckstation{
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
