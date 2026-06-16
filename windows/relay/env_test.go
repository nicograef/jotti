package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseEnvFile(t *testing.T) {
	tests := []struct {
		name string
		data string
		want map[string]string
	}{
		{
			name: "key=value mit Kommentar und Leerzeile",
			data: "# Kommentar\n\nRELAY_AUTH_TOKEN=abc123\n",
			want: map[string]string{"RELAY_AUTH_TOKEN": "abc123"},
		},
		{
			name: "Whitespace um Key und Value wird getrimmt",
			data: "  RELAY_AUTH_TOKEN  =  abc123  \n",
			want: map[string]string{"RELAY_AUTH_TOKEN": "abc123"},
		},
		{
			name: "CRLF-Zeilenenden (Notepad)",
			data: "RELAY_AUTH_TOKEN=abc123\r\nRELAY_BACKEND_URL=https://example.org/api\r\n",
			want: map[string]string{
				"RELAY_AUTH_TOKEN":  "abc123",
				"RELAY_BACKEND_URL": "https://example.org/api",
			},
		},
		{
			name: "optionale Anfuehrungszeichen werden entfernt",
			data: "RELAY_AUTH_TOKEN=\"abc123\"\nRELAY_BACKEND_URL='https://example.org/api'\n",
			want: map[string]string{
				"RELAY_AUTH_TOKEN":  "abc123",
				"RELAY_BACKEND_URL": "https://example.org/api",
			},
		},
		{
			name: "UTF-8-BOM auf der ersten Zeile wird toleriert",
			data: "\xef\xbb\xbf# Kommentar\nRELAY_AUTH_TOKEN=abc123\n",
			want: map[string]string{"RELAY_AUTH_TOKEN": "abc123"},
		},
		{
			name: "Zeilen ohne Gleichheitszeichen werden uebersprungen",
			data: "kaputt\nRELAY_AUTH_TOKEN=abc123\n",
			want: map[string]string{"RELAY_AUTH_TOKEN": "abc123"},
		},
		{
			name: "leere Datei ergibt leere Map",
			data: "",
			want: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseEnvFile([]byte(tt.data))
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("parseEnvFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnvWithFileFallbackPrefersProcessEnv(t *testing.T) {
	t.Setenv("RELAY_AUTH_TOKEN", "aus-env")
	getenv := envWithFileFallback(map[string]string{
		"RELAY_AUTH_TOKEN":  "aus-datei",
		"RELAY_BACKEND_URL": "https://datei.example/api",
	})

	if got := getenv("RELAY_AUTH_TOKEN"); got != "aus-env" {
		t.Fatalf("gesetzte Env-Variable muss Vorrang haben: got %q, want %q", got, "aus-env")
	}
	if got := getenv("RELAY_BACKEND_URL"); got != "https://datei.example/api" {
		t.Fatalf("fehlende Env-Variable muss aus der Datei kommen: got %q", got)
	}
	if got := getenv("RELAY_POLL_SECONDS"); got != "" {
		t.Fatalf("unbekannter Key muss leer sein: got %q", got)
	}
}

func TestEnvSearchDirsWindowsPrefersProgramData(t *testing.T) {
	got := envSearchDirs("windows", `C:\ProgramData`, `D:\relay`, `D:\wd`)
	want := []string{filepath.Join(`C:\ProgramData`, "jotti"), `D:\relay`, `D:\wd`}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Windows-Suchreihenfolge: got %v, want %v", got, want)
	}
}

func TestEnvSearchDirsLinuxSkipsProgramData(t *testing.T) {
	// Unter Linux (Server/Dev) wird PROGRAMDATA ignoriert, selbst wenn gesetzt.
	got := envSearchDirs("linux", "/ignored", "/opt/relay", "/work")
	want := []string{"/opt/relay", "/work"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Linux-Suchreihenfolge: got %v, want %v", got, want)
	}
}

func TestLoadEnvFileFromWorkingDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(dir+"/.env", []byte("RELAY_AUTH_TOKEN=aus-datei\n"), 0o600); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Chdir(dir)

	values := loadEnvFile()
	if got := values["RELAY_AUTH_TOKEN"]; got != "aus-datei" {
		t.Fatalf("loadEnvFile() = %v, want RELAY_AUTH_TOKEN=aus-datei", values)
	}
}
