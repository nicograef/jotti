package main

import (
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func TestLoadConfigFromEnv(t *testing.T) {
	tests := []struct {
		name        string
		env         map[string]string
		wantConfig  RelayConfig
		wantErr     bool
		errContains string
	}{
		{
			name: "uses defaults when optional values are missing",
			env: map[string]string{
				"RELAY_AUTH_TOKEN": "devrelay",
			},
			wantConfig: RelayConfig{
				Token:         "devrelay",
				BackendURL:    defaultBackendURL,
				PollSeconds:   defaultPollSeconds,
				TLSSkipVerify: true,
			},
		},
		{
			name: "uses overridden backend url and poll interval",
			env: map[string]string{
				"RELAY_AUTH_TOKEN":   "devrelay",
				"RELAY_BACKEND_URL":  "https://example.org/api",
				"RELAY_POLL_SECONDS": "5",
			},
			wantConfig: RelayConfig{
				Token:         "devrelay",
				BackendURL:    "https://example.org/api",
				PollSeconds:   5,
				TLSSkipVerify: false,
			},
		},
		{
			name: "trims trailing slash from backend url",
			env: map[string]string{
				"RELAY_AUTH_TOKEN":  "devrelay",
				"RELAY_BACKEND_URL": "https://example.org/api/",
			},
			wantConfig: RelayConfig{
				Token:         "devrelay",
				BackendURL:    "https://example.org/api",
				PollSeconds:   defaultPollSeconds,
				TLSSkipVerify: false,
			},
		},
		{
			name: "accepts explicit TLS skip true",
			env: map[string]string{
				"RELAY_AUTH_TOKEN":      "devrelay",
				"RELAY_BACKEND_URL":     "https://example.org/api",
				"RELAY_TLS_SKIP_VERIFY": "true",
			},
			wantConfig: RelayConfig{
				Token:         "devrelay",
				BackendURL:    "https://example.org/api",
				PollSeconds:   defaultPollSeconds,
				TLSSkipVerify: true,
			},
		},
		{
			name: "accepts explicit TLS skip false on localhost",
			env: map[string]string{
				"RELAY_AUTH_TOKEN":      "devrelay",
				"RELAY_TLS_SKIP_VERIFY": "0",
			},
			wantConfig: RelayConfig{
				Token:         "devrelay",
				BackendURL:    defaultBackendURL,
				PollSeconds:   defaultPollSeconds,
				TLSSkipVerify: false,
			},
		},
		{
			name: "fails on invalid TLS skip value",
			env: map[string]string{
				"RELAY_AUTH_TOKEN":      "devrelay",
				"RELAY_TLS_SKIP_VERIFY": "ja",
			},
			wantErr:     true,
			errContains: "RELAY_TLS_SKIP_VERIFY",
		},
		{
			name: "fails when token is missing",
			env: map[string]string{
				"RELAY_BACKEND_URL": "https://example.org/api",
			},
			wantErr:     true,
			errContains: "RELAY_AUTH_TOKEN",
		},
		{
			name: "fails when poll is not a positive integer",
			env: map[string]string{
				"RELAY_AUTH_TOKEN":   "devrelay",
				"RELAY_POLL_SECONDS": "0",
			},
			wantErr:     true,
			errContains: "RELAY_POLL_SECONDS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := loadConfigFromEnv(func(key string) string {
				return tt.env[key]
			})

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("expected error to contain %q, got %q", tt.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if config.Token != tt.wantConfig.Token {
				t.Fatalf("token mismatch: got %q, want %q", config.Token, tt.wantConfig.Token)
			}
			if config.BackendURL != tt.wantConfig.BackendURL {
				t.Fatalf("backend url mismatch: got %q, want %q", config.BackendURL, tt.wantConfig.BackendURL)
			}
			if config.PollSeconds != tt.wantConfig.PollSeconds {
				t.Fatalf("poll seconds mismatch: got %d, want %d", config.PollSeconds, tt.wantConfig.PollSeconds)
			}
			if config.TLSSkipVerify != tt.wantConfig.TLSSkipVerify {
				t.Fatalf("tls skip verify mismatch: got %t, want %t", config.TLSSkipVerify, tt.wantConfig.TLSSkipVerify)
			}
		})
	}
}

// fakePrinter ist eine injizierbare Druck-Funktion, die je Ziel-IP die
// Aufruf-Reihenfolge protokolliert und für vorgegebene IDs einen Fehler liefert.
type fakePrinter struct {
	mu        sync.Mutex
	callsByIP map[string][]int
	failIDs   map[int]string
}

func newFakePrinter(failIDs map[int]string) *fakePrinter {
	return &fakePrinter{
		callsByIP: make(map[string][]int),
		failIDs:   failIDs,
	}
}

func (f *fakePrinter) druck(a DruckAuftrag) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.callsByIP[a.ZielIP] = append(f.callsByIP[a.ZielIP], a.ID)
	if msg, ok := f.failIDs[a.ID]; ok {
		return errors.New(msg)
	}
	return nil
}

func TestVerarbeiteZyklusGruppiertUndHaeltReihenfolge(t *testing.T) {
	auftraege := []DruckAuftrag{
		{ID: 1, ZielIP: "10.0.0.1"},
		{ID: 2, ZielIP: "10.0.0.2"},
		{ID: 3, ZielIP: "10.0.0.1"},
		{ID: 4, ZielIP: "10.0.0.2"},
		{ID: 5, ZielIP: "10.0.0.1"},
	}
	printer := newFakePrinter(nil)

	ergebnis := verarbeiteZyklus(auftraege, printer.druck)

	if want := []int{1, 2, 3, 4, 5}; !reflect.DeepEqual(ergebnis.gedruckteIDs, want) {
		t.Fatalf("gedruckteIDs: got %v, want %v", ergebnis.gedruckteIDs, want)
	}
	if len(ergebnis.fehlversuche) != 0 {
		t.Fatalf("fehlversuche: got %v, want none", ergebnis.fehlversuche)
	}
	if want := []int{1, 3, 5}; !reflect.DeepEqual(printer.callsByIP["10.0.0.1"], want) {
		t.Fatalf("Reihenfolge IP1: got %v, want %v", printer.callsByIP["10.0.0.1"], want)
	}
	if want := []int{2, 4}; !reflect.DeepEqual(printer.callsByIP["10.0.0.2"], want) {
		t.Fatalf("Reihenfolge IP2: got %v, want %v", printer.callsByIP["10.0.0.2"], want)
	}
}

func TestVerarbeiteZyklusSkipNachErstfehler(t *testing.T) {
	auftraege := []DruckAuftrag{
		{ID: 1, ZielIP: "10.0.0.1"},
		{ID: 2, ZielIP: "10.0.0.1"},
		{ID: 3, ZielIP: "10.0.0.1"},
		{ID: 4, ZielIP: "10.0.0.2"},
		{ID: 5, ZielIP: "10.0.0.2"},
	}
	printer := newFakePrinter(map[int]string{1: "nicht erreichbar"})

	ergebnis := verarbeiteZyklus(auftraege, printer.druck)

	if want := []int{4, 5}; !reflect.DeepEqual(ergebnis.gedruckteIDs, want) {
		t.Fatalf("gedruckteIDs: got %v, want %v", ergebnis.gedruckteIDs, want)
	}
	if want := []fehlversuch{{ID: 1, Fehler: "nicht erreichbar"}}; !reflect.DeepEqual(ergebnis.fehlversuche, want) {
		t.Fatalf("fehlversuche: got %v, want %v", ergebnis.fehlversuche, want)
	}
	// Nach dem Erstfehler von IP1 darf nur Auftrag 1 versucht worden sein.
	if want := []int{1}; !reflect.DeepEqual(printer.callsByIP["10.0.0.1"], want) {
		t.Fatalf("IP1-Versuche: got %v, want %v", printer.callsByIP["10.0.0.1"], want)
	}
	// Der tote Drucker blockiert IP2 nicht.
	if want := []int{4, 5}; !reflect.DeepEqual(printer.callsByIP["10.0.0.2"], want) {
		t.Fatalf("IP2-Versuche: got %v, want %v", printer.callsByIP["10.0.0.2"], want)
	}
}

func TestVerarbeiteGruppeFehlerInMitteUeberspringtRest(t *testing.T) {
	gruppe := []DruckAuftrag{
		{ID: 1, ZielIP: "10.0.0.1"},
		{ID: 2, ZielIP: "10.0.0.1"},
		{ID: 3, ZielIP: "10.0.0.1"},
	}
	printer := newFakePrinter(map[int]string{2: "senden fehlgeschlagen"})

	gedruckte, fehler := verarbeiteGruppe(gruppe, printer.druck)

	if want := []int{1}; !reflect.DeepEqual(gedruckte, want) {
		t.Fatalf("gedruckte: got %v, want %v", gedruckte, want)
	}
	if fehler == nil || fehler.ID != 2 || fehler.Fehler != "senden fehlgeschlagen" {
		t.Fatalf("fehler: got %+v, want {2 senden fehlgeschlagen}", fehler)
	}
	// Auftrag 3 darf nach dem Fehler bei 2 nicht mehr versucht worden sein.
	if want := []int{1, 2}; !reflect.DeepEqual(printer.callsByIP["10.0.0.1"], want) {
		t.Fatalf("Versuche: got %v, want %v", printer.callsByIP["10.0.0.1"], want)
	}
}

func TestFuehreZyklusAusMeldetErfolgeUndFehlversuche(t *testing.T) {
	auftraege := []DruckAuftrag{
		{ID: 1, ZielIP: "10.0.0.1"},
		{ID: 2, ZielIP: "10.0.0.1"},
		{ID: 3, ZielIP: "10.0.0.2"},
		{ID: 4, ZielIP: "10.0.0.2"},
	}
	printer := newFakePrinter(map[int]string{3: "kaputt"})

	var gemeldet []zyklusErgebnis
	melde := func(ergebnis zyklusErgebnis) error {
		gemeldet = append(gemeldet, ergebnis)
		return nil
	}

	ergebnis, err := fuehreZyklusAus(auftraege, printer.druck, melde)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(gemeldet) != 1 {
		t.Fatalf("melde calls: got %d, want 1", len(gemeldet))
	}
	if !reflect.DeepEqual(gemeldet[0], ergebnis) {
		t.Fatalf("gemeldetes Ergebnis weicht ab: got %+v, want %+v", gemeldet[0], ergebnis)
	}
	if want := []int{1, 2}; !reflect.DeepEqual(ergebnis.gedruckteIDs, want) {
		t.Fatalf("gedruckteIDs: got %v, want %v", ergebnis.gedruckteIDs, want)
	}
	if want := []fehlversuch{{ID: 3, Fehler: "kaputt"}}; !reflect.DeepEqual(ergebnis.fehlversuche, want) {
		t.Fatalf("fehlversuche: got %v, want %v", ergebnis.fehlversuche, want)
	}
}

func TestFuehreZyklusAusOhneAuftraegeMeldetNicht(t *testing.T) {
	printer := newFakePrinter(nil)

	called := false
	melde := func(zyklusErgebnis) error {
		called = true
		return nil
	}

	ergebnis, err := fuehreZyklusAus(nil, printer.druck, melde)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatalf("melde wurde ohne Aufträge aufgerufen")
	}
	if len(ergebnis.gedruckteIDs) != 0 || len(ergebnis.fehlversuche) != 0 {
		t.Fatalf("Ergebnis nicht leer: %+v", ergebnis)
	}
}

func TestFuehreZyklusAusGibtMeldefehlerWeiter(t *testing.T) {
	auftraege := []DruckAuftrag{{ID: 1, ZielIP: "10.0.0.1"}}
	printer := newFakePrinter(nil)

	melde := func(zyklusErgebnis) error {
		return errors.New("backend nicht erreichbar")
	}

	_, err := fuehreZyklusAus(auftraege, printer.druck, melde)
	if err == nil || !strings.Contains(err.Error(), "backend nicht erreichbar") {
		t.Fatalf("expected melde error, got %v", err)
	}
}
