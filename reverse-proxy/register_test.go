package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterWithACMEDNS(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/register" {
			http.Error(w, "unerwarteter Request", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		// acme-dns liefert zusätzlich fulldomain/allowfrom — werden ignoriert.
		_, _ = w.Write([]byte(`{"username":"u1","password":"p1","fulldomain":"sub1.auth.jotti.rocks","subdomain":"sub1","allowfrom":[]}`))
	}))
	defer srv.Close()

	state, err := registerWithACMEDNS(srv.Client(), srv.URL)
	if err != nil {
		t.Fatalf("unerwarteter Fehler: %v", err)
	}
	if state.Username != "u1" || state.Password != "p1" || state.Subdomain != "sub1" {
		t.Fatalf("Credentials: got %+v, want {u1 p1 sub1}", state)
	}
}

func TestRegisterWithACMEDNSRejectsErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Rate-Limit auf /register (siehe nginx-Konfiguration der rocks-Infra).
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	if _, err := registerWithACMEDNS(srv.Client(), srv.URL); err == nil {
		t.Fatal("erwartete Fehler bei HTTP 429, bekam nil")
	}
}

func TestRegisterWithACMEDNSRejectsIncompleteBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"username":"u1"}`)) // password/subdomain fehlen
	}))
	defer srv.Close()

	if _, err := registerWithACMEDNS(srv.Client(), srv.URL); err == nil {
		t.Fatal("erwartete Fehler bei unvollständiger Antwort, bekam nil")
	}
}
