//go:build unit

package test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nicograef/jotti/backend/seed"
)

type fakeReseter struct {
	called bool
	err    error
}

func (f *fakeReseter) ResetAndSeed(_ context.Context) error {
	f.called = true
	return f.err
}

func TestResetAndSeedHandler_Erfolg(t *testing.T) {
	fake := &fakeReseter{}
	h := Handler{Reseter: fake}

	req := httptest.NewRequest(http.MethodPost, "/test/reset-and-seed", nil)
	w := httptest.NewRecorder()
	h.ResetAndSeedHandler()(w, req)

	if !fake.called {
		t.Fatal("ResetAndSeed wurde nicht aufgerufen")
	}
	if w.Code != http.StatusOK {
		t.Fatalf("erwartet 200, bekam %d", w.Code)
	}

	var resp ResetResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Antwort dekodieren: %v", err)
	}

	if resp.Service.Username != seed.DemoServiceUsername || resp.Service.Password != seed.DemoPassword {
		t.Fatalf("unerwartete Service-Zugangsdaten: %+v", resp.Service)
	}
	if resp.Admin.Username != seed.DemoAdminUsername || resp.Admin.Password != seed.DemoPassword {
		t.Fatalf("unerwartete Admin-Zugangsdaten: %+v", resp.Admin)
	}
	if resp.Serviceleitung.Username != seed.DemoServiceleitungUsername {
		t.Fatalf("unerwartete Serviceleitung-Zugangsdaten: %+v", resp.Serviceleitung)
	}
}

func TestResetAndSeedHandler_Fehler(t *testing.T) {
	fake := &fakeReseter{err: errors.New("boom")}
	h := Handler{Reseter: fake}

	req := httptest.NewRequest(http.MethodPost, "/test/reset-and-seed", nil)
	w := httptest.NewRecorder()
	h.ResetAndSeedHandler()(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("erwartet 500, bekam %d", w.Code)
	}
}
