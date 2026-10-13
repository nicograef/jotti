package core

import (
	"strings"
	"testing"
)

func TestParseAdminOTPPlainLine(t *testing.T) {
	code, found := ParseAdminOTP("10:22AM INF ADMIN-EINMALPASSWORT benutzer=admin code=327547\n")
	if !found || code != "327547" {
		t.Fatalf("ParseAdminOTP = (%q, %v), want (\"327547\", true)", code, found)
	}
}

func TestParseAdminOTPJSONLine(t *testing.T) {
	// So schreibt der JSON-Handler des Backends die Zeile in `docker logs`.
	logs := `{"time":"2026-10-13T10:22:00.000000000Z","level":"INFO","msg":"ADMIN-EINMALPASSWORT benutzer=admin code=654321"}` + "\n"
	code, found := ParseAdminOTP(logs)
	if !found || code != "654321" {
		t.Fatalf("ParseAdminOTP (JSON) = (%q, %v), want (\"654321\", true)", code, found)
	}
}

func TestParseAdminOTPNewestWins(t *testing.T) {
	logs := strings.Join([]string{
		"INF ADMIN-EINMALPASSWORT benutzer=admin code=111111",
		"INF irgendeine andere Zeile",
		"INF ADMIN-EINMALPASSWORT benutzer=admin code=222222",
	}, "\n")
	code, found := ParseAdminOTP(logs)
	if !found || code != "222222" {
		t.Fatalf("ParseAdminOTP (newest) = (%q, %v), want (\"222222\", true)", code, found)
	}
}

func TestParseAdminOTPNonePresent(t *testing.T) {
	code, found := ParseAdminOTP("INF backend gestartet\nINF DB verbunden\n")
	if found || code != "" {
		t.Fatalf("ParseAdminOTP (keine) = (%q, %v), want (\"\", false)", code, found)
	}
}

func TestAdminCodeHinweisFound(t *testing.T) {
	msg := AdminCodeHinweis("327547", true)
	for _, want := range []string{AdminUsername, "327547", "Passwort festlegen"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("gefunden-Hinweis enthaelt %q nicht: %q", want, msg)
		}
	}
}

func TestAdminCodeHinweisNotFound(t *testing.T) {
	msg := AdminCodeHinweis("", false)
	if msg != "Einrichtung ist abgeschlossen oder es liegt kein Code vor. jotti neu starten, dann wird ein neuer Code angezeigt." {
		t.Fatalf("unerwartete nicht-gefunden-Meldung: %q", msg)
	}
	lower := strings.ToLower(msg)
	for _, verboten := range []string{"log", "docker"} {
		if strings.Contains(lower, verboten) {
			t.Fatalf("nicht-gefunden-Meldung darf %q nicht enthalten: %q", verboten, msg)
		}
	}
	if strings.ContainsAny(msg, "0123456789") {
		t.Fatalf("nicht-gefunden-Meldung darf keinen Code enthalten: %q", msg)
	}
}
