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

func (m *mockQuery) GetAllProducts(ctx context.Context) ([]application.ProduktMitVerkauf, error) {
	return []application.ProduktMitVerkauf{
		{
			Produkt: dom.Produkt{
				ID:        1,
				Name:      "French Fries",
				Kategorie: dom.EssenKategorie,
				Varianten: []dom.Variante{
					{ID: 1, Name: "Regular", PreisCents: 499, Status: dom.ActiveStatus},
					{ID: 2, Name: "Large", PreisCents: 699, Status: dom.ActiveStatus},
				},
			},
			HatVerkaeufe: true,
		},
		{
			Produkt: dom.Produkt{
				ID:        2,
				Name:      "Cola",
				Kategorie: dom.GetraenkKategorie,
				Varianten: []dom.Variante{
					{ID: 3, Name: "0.5L", PreisCents: 250, Status: dom.ActiveStatus},
				},
			},
			HatVerkaeufe: false,
		},
	}, m.err
}

func (m *mockQuery) GetActiveProducts(ctx context.Context) ([]dom.Produkt, error) {
	return []dom.Produkt{{
		ID:        1,
		Name:      "French Fries",
		Kategorie: dom.EssenKategorie,
		Varianten: []dom.Variante{
			{ID: 1, Name: "Regular", PreisCents: 499, Status: dom.ActiveStatus},
		},
	}}, m.err
}

func TestGetAllProductsHandler_Success(t *testing.T) {
	handler := &QueryHandler{Query: &mockQuery{}}

	req := httptest.NewRequest(http.MethodGet, "/get-all-produkte", nil)
	rec := httptest.NewRecorder()

	handler.GetAllProductsHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	// hatVerkaeufe muss pro Produkt im JSON serialisiert werden — fängt eine
	// Regression, die das Feld beim DTO-Mapping wieder verliert.
	var body struct {
		Produkte []struct {
			ID           int  `json:"id"`
			HatVerkaeufe bool `json:"hatVerkaeufe"`
		} `json:"produkte"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if len(body.Produkte) != 2 {
		t.Fatalf("expected 2 produkte, got %d", len(body.Produkte))
	}
	if !body.Produkte[0].HatVerkaeufe {
		t.Errorf("expected produkt %d hatVerkaeufe=true, got false", body.Produkte[0].ID)
	}
	if body.Produkte[1].HatVerkaeufe {
		t.Errorf("expected produkt %d hatVerkaeufe=false, got true", body.Produkte[1].ID)
	}
}

func TestGetAllProductsHandler_Failure(t *testing.T) {
	handler := &QueryHandler{Query: &mockQuery{err: application.ErrDatabase}}

	req := httptest.NewRequest(http.MethodGet, "/get-all-produkte", nil)
	rec := httptest.NewRecorder()

	handler.GetAllProductsHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}
