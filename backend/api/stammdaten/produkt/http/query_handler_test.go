//go:build unit

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nicograef/jotti/backend/api/stammdaten/produkt/application"
	dom "github.com/nicograef/jotti/backend/domain/produkt"
)

type mockQuery struct {
	err error
}

func (m *mockQuery) GetAllProducts(ctx context.Context) ([]dom.Produkt, error) {
	return []dom.Produkt{{
		ID:        1,
		Name:      "French Fries",
		Kategorie: dom.EssenKategorie,
		Varianten: []dom.Variante{
			{ID: 1, Name: "Regular", PreisCents: 499, Status: dom.ActiveStatus},
			{ID: 2, Name: "Large", PreisCents: 699, Status: dom.ActiveStatus},
		},
	}}, m.err
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
