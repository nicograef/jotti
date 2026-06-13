package core

import (
	"strings"
	"testing"
)

func TestPortBelegtDiagnoseNamesOwner(t *testing.T) {
	msg := PortBelegtDiagnose(80, []PortOwner{{LocalPort: 80, PID: 4, ProcessName: "System"}})
	for _, want := range []string{"Port 80", "System", "PID 4", "belegt", "beenden"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("Diagnose enthaelt %q nicht: %q", want, msg)
		}
	}
}

func TestPortBelegtDiagnoseGenericFallback(t *testing.T) {
	msg := PortBelegtDiagnose(443, nil)
	if !strings.Contains(msg, "Port 443") {
		t.Fatalf("Port fehlt: %q", msg)
	}
	if !strings.Contains(msg, "nicht ermittelt werden") {
		t.Fatalf("generischer Fallback-Hinweis fehlt: %q", msg)
	}
}

func TestPortBelegtDiagnoseUnknownProcessName(t *testing.T) {
	msg := PortBelegtDiagnose(80, []PortOwner{{LocalPort: 80, PID: 9999, ProcessName: ""}})
	if !strings.Contains(msg, "PID 9999") {
		t.Fatalf("PID fehlt: %q", msg)
	}
	if !strings.Contains(msg, "unbekanntes Programm") {
		t.Fatalf("Platzhalter fuer unbekannten Prozess fehlt: %q", msg)
	}
}

func TestStaticDiagnosesHaveActionHints(t *testing.T) {
	diagnosen := map[string]string{
		"DockerCLIFehlt":             DiagnoseDockerCLIFehlt,
		"DockerNichtInstalliert":     DiagnoseDockerNichtInstalliert,
		"DockerStartFehlgeschlagen":  DiagnoseDockerStartFehlgeschlagen,
		"EngineSwitchFehlgeschlagen": DiagnoseEngineSwitchFehlgeschlagen,
	}
	for name, msg := range diagnosen {
		if !strings.Contains(msg, "Bitte") {
			t.Fatalf("%s nennt keinen Handlungsschritt: %q", name, msg)
		}
	}
	if !strings.Contains(DiagnoseDockerCLIFehlt, DockerBinPath) {
		t.Fatalf("Docker-CLI-Diagnose nennt den Bin-Pfad nicht: %q", DiagnoseDockerCLIFehlt)
	}
	if !strings.Contains(DiagnoseDockerNichtInstalliert, DockerDesktopPath) {
		t.Fatalf("Installations-Diagnose nennt den Desktop-Pfad nicht: %q", DiagnoseDockerNichtInstalliert)
	}
}
