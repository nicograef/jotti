//go:build unit

package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nicograef/jotti/backend/api/middleware"
	"github.com/nicograef/jotti/backend/api/table/application"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/table"
)

type mockQuery struct {
	tisch    table.Tisch
	order    kasse.Bestellung
	position kasse.Position
	balance  int
	err      error
}

func (m mockQuery) GetAllTische(ctx context.Context) ([]table.Tisch, error) {
	return []table.Tisch{m.tisch}, m.err
}

func (m mockQuery) GetAktiveTische(ctx context.Context) ([]table.AktiverTisch, error) {
	return []table.AktiverTisch{{ID: m.tisch.ID, Name: m.tisch.Name, SaldoCents: 0}}, m.err
}

func (m mockQuery) GetTischHistorie(ctx context.Context, tischID int) ([]kasse.HistorieEintrag, error) {
	return []kasse.HistorieEintrag{{Art: kasse.HistorieEintragBestellung, Bestellung: &m.order}}, m.err
}

func (m mockQuery) GetTischState(ctx context.Context, tischID int) (application.TischStateView, error) {
	return application.TischStateView{
		SaldoCents:            m.balance,
		UnbezahltePositionen:  []kasse.Position{m.position},
		AusstehendePositionen: []kasse.Position{m.position},
	}, m.err
}

func (m mockQuery) GetAktiveTischeMitFavoriten(_ context.Context, _ int) ([]table.AktiverTischMitFavorit, error) {
	return nil, m.err
}

func (m mockQuery) GetMeineTischeState(_ context.Context, _ int) ([]application.TischStateView, error) {
	return []application.TischStateView{{
		TischID:               m.tisch.ID,
		TischName:             m.tisch.Name,
		SaldoCents:            m.balance,
		UnbezahltePositionen:  []kasse.Position{m.position},
		AusstehendePositionen: []kasse.Position{m.position},
	}}, m.err
}

func TestGetAllTischeHandler_Success(t *testing.T) {
	handler := &QueryHandler{Query: mockQuery{}}

	req := httptest.NewRequest(http.MethodPost, "/get-all-tische", nil)
	rec := httptest.NewRecorder()

	handler.GetAllTischeHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestGetAllTischeHandler_Failure(t *testing.T) {
	handler := &QueryHandler{Query: mockQuery{err: application.ErrDatabase}}

	req := httptest.NewRequest(http.MethodPost, "/get-all-tische", nil)
	rec := httptest.NewRecorder()

	handler.GetAllTischeHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}

func TestPositionResponsesIncludeSteuersatz(t *testing.T) {
	t.Parallel()

	testPosition := kasse.Position{
		PositionID:   "123e4567-e89b-42d3-a456-426614174000",
		VarianteID:   10,
		ProduktName:  "Apfelschorle",
		VarianteName: "0,5l",
		Kategorie:    "getraenk",
		Steuersatz:   "regel",
		Einzelpreis:  350,
		Menge:        2,
	}

	handler := &QueryHandler{Query: mockQuery{
		tisch:    table.Tisch{ID: 1, Name: "Tisch 1"},
		order:    kasse.Bestellung{Positionen: []kasse.Position{testPosition}},
		position: testPosition,
		balance:  700,
	}}

	t.Run("get-tisch-state", func(t *testing.T) {
		body := []byte(`{"tischId":1}`)
		req := httptest.NewRequest(http.MethodPost, "/get-tisch-state", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		handler.GetTischStateHandler().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var resp getTischStateResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("expected valid response body, got %v", err)
		}

		if len(resp.UnbezahltePositionen) == 0 {
			t.Fatal("expected unbezahltePositionen to contain at least one position")
		}
		if resp.UnbezahltePositionen[0].Steuersatz == "" {
			t.Fatal("expected steuersatz in unbezahltePositionen[0] to be present")
		}

		if len(resp.AusstehendePositionen) == 0 {
			t.Fatal("expected ausstehendePositionen to contain at least one position")
		}
		if resp.AusstehendePositionen[0].Steuersatz == "" {
			t.Fatal("expected steuersatz in ausstehendePositionen[0] to be present")
		}
	})

	t.Run("get-tisch-historie", func(t *testing.T) {
		body := []byte(`{"tischId":1}`)
		req := httptest.NewRequest(http.MethodPost, "/get-tisch-historie", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		handler.GetTischHistorieHandler().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var resp struct {
			Historie []struct {
				Positionen []position `json:"positionen"`
			} `json:"historie"`
		}
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("expected valid response body, got %v", err)
		}

		if len(resp.Historie) == 0 || len(resp.Historie[0].Positionen) == 0 {
			t.Fatal("expected historie entry with at least one position")
		}
		if resp.Historie[0].Positionen[0].Steuersatz == "" {
			t.Fatal("expected steuersatz in historie[0].positionen[0] to be present")
		}
	})

	t.Run("get-meine-tische-state", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/get-meine-tische-state", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, 7))
		rec := httptest.NewRecorder()

		handler.GetMeineTischeStateHandler().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var resp getMeineTischeStateResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("expected valid response body, got %v", err)
		}

		if len(resp.Tische) == 0 || len(resp.Tische[0].UnbezahltePositionen) == 0 {
			t.Fatal("expected at least one tisch with unbezahltePositionen")
		}
		if resp.Tische[0].UnbezahltePositionen[0].Steuersatz == "" {
			t.Fatal("expected steuersatz in tische[0].unbezahltePositionen[0] to be present")
		}

		if len(resp.Tische[0].AusstehendePositionen) == 0 {
			t.Fatal("expected at least one ausstehende position")
		}
		if resp.Tische[0].AusstehendePositionen[0].Steuersatz == "" {
			t.Fatal("expected steuersatz in tische[0].ausstehendePositionen[0] to be present")
		}
	})
}
