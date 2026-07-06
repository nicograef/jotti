//go:build unit

package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/domain/kasse"
)

type mockQuery struct {
	historie []kasse.DirektverkaufHistorieEintrag
	err      error
}

func (m *mockQuery) GetDirektverkaufHistorie(_ context.Context) ([]kasse.DirektverkaufHistorieEintrag, error) {
	return m.historie, m.err
}

func TestGetDirektverkaufHistorieHandler_ReturnsHistorie(t *testing.T) {
	const posID = "6f9619ff-8b86-d011-b42d-00cf4fc964ff"
	handler := &QueryHandler{Query: &mockQuery{historie: []kasse.DirektverkaufHistorieEintrag{
		{
			VerkaufID:            "11111111-1111-1111-1111-111111111111",
			UserName:             "Leitung",
			GetaetigtAm:          time.Now(),
			Positionen:           []kasse.Position{{PositionID: posID, VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 500, Menge: 2}},
			GesamtbetragCents:    1000,
			OffenePositionen:     []kasse.Position{{PositionID: posID, VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 500, Menge: 1}},
			GesamtStorniertCents: 500,
		},
	}}}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/get-direktverkauf-historie", nil)
	handler.GetDirektverkaufHistorieHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp struct {
		Historie []struct {
			VerkaufID            string `json:"verkaufId"`
			GesamtbetragCents    int    `json:"gesamtbetragCents"`
			GesamtStorniertCents int    `json:"gesamtStorniertCents"`
			Positionen           []struct {
				Steuersatz string `json:"steuersatz"`
			} `json:"positionen"`
			OffenePositionen []struct {
				Menge      int    `json:"menge"`
				Steuersatz string `json:"steuersatz"`
			} `json:"offenePositionen"`
		} `json:"historie"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(resp.Historie) != 1 {
		t.Fatalf("expected 1 historie entry, got %d", len(resp.Historie))
	}
	if resp.Historie[0].GesamtbetragCents != 1000 {
		t.Errorf("expected gesamtbetragCents 1000, got %d", resp.Historie[0].GesamtbetragCents)
	}
	if resp.Historie[0].GesamtStorniertCents != 500 {
		t.Errorf("expected gesamtStorniertCents 500, got %d", resp.Historie[0].GesamtStorniertCents)
	}
	if len(resp.Historie[0].Positionen) != 1 || resp.Historie[0].Positionen[0].Steuersatz == "" {
		t.Errorf("expected one position with steuersatz, got %+v", resp.Historie[0].Positionen)
	}
	if len(resp.Historie[0].OffenePositionen) != 1 || resp.Historie[0].OffenePositionen[0].Menge != 1 {
		t.Errorf("expected one offene Position with menge 1, got %+v", resp.Historie[0].OffenePositionen)
	}
	if resp.Historie[0].OffenePositionen[0].Steuersatz == "" {
		t.Errorf("expected offene position steuersatz to be set")
	}
}

func TestGetDirektverkaufHistorieHandler_Error(t *testing.T) {
	handler := &QueryHandler{Query: &mockQuery{err: errors.New("boom")}}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/get-direktverkauf-historie", nil)
	handler.GetDirektverkaufHistorieHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}
