//go:build unit

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nicograef/jotti/backend/repository/drucker_repo"
)

type mockDruckerCommand struct {
	err error
}

func (m *mockDruckerCommand) UpsertKategorieDrucker(ctx context.Context, kategorie, druckerIP, bonmodus string) error {
	return m.err
}

type mockDruckerQuery struct {
	result []drucker_repo.DruckerKonfig
	err    error
}

func (m *mockDruckerQuery) GetAlleKategorieDrucker(ctx context.Context) ([]drucker_repo.DruckerKonfig, error) {
	return m.result, m.err
}

func TestUpdateDruckerConfigHandler_Success(t *testing.T) {
	handler := &CommandHandler{Command: &mockDruckerCommand{}}

	body := `{"kategorie":"essen","druckerIp":"192.168.1.51","bonmodus":"pro_position"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/update-drucker-config", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateDruckerConfigHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateDruckerConfigHandler_EmptyIPAllowed(t *testing.T) {
	handler := &CommandHandler{Command: &mockDruckerCommand{}}

	body := `{"kategorie":"getraenk","druckerIp":"","bonmodus":"pro_bestellung"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/update-drucker-config", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateDruckerConfigHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200 for empty IP, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateDruckerConfigHandler_InvalidKategorie(t *testing.T) {
	handler := &CommandHandler{Command: &mockDruckerCommand{}}

	body := `{"kategorie":"invalid","druckerIp":"192.168.1.51","bonmodus":"pro_position"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/update-drucker-config", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateDruckerConfigHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid kategorie, got %d", rec.Code)
	}
}

func TestUpdateDruckerConfigHandler_InvalidBonmodus(t *testing.T) {
	handler := &CommandHandler{Command: &mockDruckerCommand{}}

	body := `{"kategorie":"essen","druckerIp":"192.168.1.51","bonmodus":"invalid"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/update-drucker-config", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateDruckerConfigHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid bonmodus, got %d", rec.Code)
	}
}

func TestUpdateDruckerConfigHandler_InvalidIP(t *testing.T) {
	handler := &CommandHandler{Command: &mockDruckerCommand{}}

	body := `{"kategorie":"essen","druckerIp":"not-an-ip","bonmodus":"pro_position"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/update-drucker-config", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateDruckerConfigHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid IP, got %d", rec.Code)
	}
}

func TestGetDruckerConfigHandler_Success(t *testing.T) {
	konfigs := []drucker_repo.DruckerKonfig{
		{Kategorie: "essen", DruckerIP: "192.168.1.51", Bonmodus: "pro_position"},
		{Kategorie: "getraenk", DruckerIP: "", Bonmodus: "pro_position"},
		{Kategorie: "sonstiges", DruckerIP: "", Bonmodus: "pro_position"},
	}
	handler := &QueryHandler{Query: &mockDruckerQuery{result: konfigs}}

	req := httptest.NewRequest(http.MethodPost, "/admin/get-drucker-config", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.GetDruckerConfigHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}
