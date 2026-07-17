//go:build unit

package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nicograef/jotti/backend/api/stammdaten/produkt/application"
	dom "github.com/nicograef/jotti/backend/domain/produkt"
)

type mockQuery struct {
	err error
}

func (m *mockQuery) GetAllProdukte(ctx context.Context) ([]dom.Produkt, error) {
	return []dom.Produkt{
		{
			ID:        1,
			Name:      "French Fries",
			Kategorie: dom.EssenKategorie,
			Varianten: []dom.Variante{
				{ID: 1, Name: "Regular", PreisCents: 499, Status: dom.ActiveStatus},
				{ID: 2, Name: "Large", PreisCents: 699, Status: dom.ActiveStatus},
			},
		},
		{
			ID:        2,
			Name:      "Cola",
			Kategorie: dom.GetraenkKategorie,
			Varianten: []dom.Variante{
				{ID: 3, Name: "0.5L", PreisCents: 250, Status: dom.ActiveStatus},
			},
		},
	}, m.err
}

func (m *mockQuery) GetActiveProdukte(ctx context.Context) ([]dom.Produkt, error) {
	return []dom.Produkt{{
		ID:        1,
		Name:      "French Fries",
		Kategorie: dom.EssenKategorie,
		Varianten: []dom.Variante{
			{ID: 1, Name: "Regular", PreisCents: 499, Status: dom.ActiveStatus},
		},
	}}, m.err
}

func TestGetAllProdukteHandler_Success(t *testing.T) {
	handler := &QueryHandler{Query: &mockQuery{}}

	req := httptest.NewRequest(http.MethodGet, "/get-all-produkte", nil)
	rec := httptest.NewRecorder()

	handler.GetAllProdukteHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var body struct {
		Produkte []struct {
			ID        int    `json:"id"`
			Name      string `json:"name"`
			Varianten []struct {
				ID int `json:"id"`
			} `json:"varianten"`
		} `json:"produkte"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if len(body.Produkte) != 2 {
		t.Fatalf("expected 2 produkte, got %d", len(body.Produkte))
	}
	if body.Produkte[0].Name != "French Fries" {
		t.Errorf("expected first produkt 'French Fries', got %q", body.Produkte[0].Name)
	}
	if len(body.Produkte[0].Varianten) != 2 {
		t.Errorf("expected produkt %d to have 2 varianten, got %d", body.Produkte[0].ID, len(body.Produkte[0].Varianten))
	}
}

func TestGetAllProdukteHandler_Failure(t *testing.T) {
	handler := &QueryHandler{Query: &mockQuery{err: application.ErrDatabase}}

	req := httptest.NewRequest(http.MethodGet, "/get-all-produkte", nil)
	rec := httptest.NewRecorder()

	handler.GetAllProdukteHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}
