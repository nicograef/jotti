package main

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"
)

func newTestStatusServer(t *testing.T) *statusServer {
	t.Helper()
	return newStatusServer(statusConfig{
		zone:     "lokal.jotti.rocks",
		state:    InstallState{Username: "u", Password: "p", Subdomain: "sub-1"},
		hasState: true,
		lanIP:    netip.MustParseAddr("192.168.1.50"),
		lanOK:    true,
	})
}

func render(t *testing.T, s *statusServer) string {
	t.Helper()
	rec := httptest.NewRecorder()
	s.handle(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("Status-Code = %d, want 200", rec.Code)
	}
	return rec.Body.String()
}

func TestStatusPageGreen(t *testing.T) {
	s := newTestStatusServer(t)
	s.probeCert = func() certState { return certValid }
	s.checkRebind = func() bool { return true }

	body := render(t, s)

	if !strings.Contains(body, "https://192-168-1-50.sub-1.lokal.jotti.rocks") {
		t.Error("grüne Adresse fehlt auf der Seite")
	}
	if !strings.Contains(body, "data:image/png;base64,") {
		t.Error("QR-Code fehlt im grünen Zustand")
	}
	if strings.Contains(body, "http-equiv=\"refresh\"") {
		t.Error("grüner Zustand soll sich nicht selbst aktualisieren")
	}
	if !strings.Contains(body, "Vereins-WLAN") {
		t.Error("WLAN-Hinweis fehlt")
	}
}

func TestStatusPageShowsSetupHint(t *testing.T) {
	s := newTestStatusServer(t)
	s.probeCert = func() certState { return certValid }
	s.checkRebind = func() bool { return true }

	body := render(t, s)

	if !strings.Contains(body, "Startkonsole") {
		t.Error("Ersteinrichtungs-Hinweis (Startkonsole) fehlt auf der Status-Seite")
	}
}

func TestStatusPageFallbackWhileIssuing(t *testing.T) {
	s := newTestStatusServer(t)
	s.probeCert = func() certState { return certNone }
	s.checkRebind = func() bool { return true }

	body := render(t, s)

	if !strings.Contains(body, "https://192.168.1.50") {
		t.Error("Fallback-Adresse fehlt")
	}
	if strings.Contains(body, "data:image/png;base64,") {
		t.Error("ohne gültiges Zertifikat darf kein QR-Code erscheinen")
	}
	if !strings.Contains(body, "http-equiv=\"refresh\"") {
		t.Error("Ausstellungs-Zustand soll sich selbst aktualisieren")
	}
}

func TestStatusPageRebindShowsGuideLink(t *testing.T) {
	s := newTestStatusServer(t)
	s.probeCert = func() certState { return certValid }
	s.checkRebind = func() bool { return false }

	body := render(t, s)

	if !strings.Contains(body, rebindGuideURL) {
		t.Error("Rebind-Anleitung-Link fehlt")
	}
	if !strings.Contains(body, "https://192.168.1.50") {
		t.Error("Fallback-Adresse fehlt bei Rebind-Blockade")
	}
}

func TestQRDataURI(t *testing.T) {
	uri := string(qrDataURI("https://192-168-1-50.sub-1.lokal.jotti.rocks"))
	if !strings.HasPrefix(uri, "data:image/png;base64,") {
		t.Fatalf("unerwartetes Präfix: %.40q", uri)
	}
	if len(uri) < 100 {
		t.Errorf("QR-Daten-URI verdächtig kurz: %d Zeichen", len(uri))
	}
}
