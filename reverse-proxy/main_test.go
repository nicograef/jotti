package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadConfigModes pinnt die Modus-Entscheidung aus der Umgebung, samt der
// Zusage, dass der LAN-Modus ohne gemountetes State-Verzeichnis kein Modus mehr
// ist: ein leeres (oder nur aus Leerzeichen bestehendes) JOTTI_DOMAIN im
// Public-Stack bricht damit ab, statt still auf die interne CA auszuweichen.
func TestLoadConfigModes(t *testing.T) {
	tests := []struct {
		name        string
		env         map[string]string
		stateDir    bool
		wantChecked string
		wantErr     bool
	}{
		{
			name:        "LAN-Modus mit gemountetem State-Verzeichnis",
			env:         map[string]string{},
			stateDir:    true,
			wantChecked: "/state",
		},
		{
			name:        "LAN-Modus ohne State-Verzeichnis",
			env:         map[string]string{},
			stateDir:    false,
			wantChecked: "/state",
			wantErr:     true,
		},
		{
			name:        "JOTTI_DOMAIN nur aus Leerzeichen ist leer",
			env:         map[string]string{"JOTTI_DOMAIN": "   "},
			stateDir:    false,
			wantChecked: "/state",
			wantErr:     true,
		},
		{
			name:     "Public-Modus braucht kein State-Verzeichnis",
			env:      map[string]string{"JOTTI_DOMAIN": "jotti.meinverein.de"},
			stateDir: false,
		},
		{
			name:     "HTTP-Only-Modus braucht kein State-Verzeichnis",
			env:      map[string]string{"PROXY_HTTP_ONLY": "1"},
			stateDir: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checked := ""
			cfg, err := loadConfig(
				func(key string) string { return tt.env[key] },
				func(path string) bool { checked = path; return tt.stateDir },
			)

			if (err != nil) != tt.wantErr {
				t.Fatalf("loadConfig-Fehler = %v, wantErr %v", err, tt.wantErr)
			}
			if checked != tt.wantChecked {
				t.Errorf("geprüftes Verzeichnis = %q, want %q", checked, tt.wantChecked)
			}
			if tt.wantErr {
				return
			}
			if cfg.zone != defaultZone {
				t.Errorf("zone = %q, want %q", cfg.zone, defaultZone)
			}
			if cfg.statePath != defaultStatePath {
				t.Errorf("statePath = %q, want %q", cfg.statePath, defaultStatePath)
			}
		})
	}
}

// TestWriteCaddyfileIsOwnerOnly prüft den Modus der erzeugten Datei — beide
// Fälle: neu angelegt und über eine bereits vorhandene Datei geschrieben. Die
// LAN-Caddyfile trägt die acme-dns-Zugangsdaten, wie install.json.
func TestWriteCaddyfileIsOwnerOnly(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(t *testing.T, path string)
	}{
		{name: "neue Datei", prepare: func(*testing.T, string) {}},
		{
			name: "vorhandene Datei mit weiterem Modus",
			prepare: func(t *testing.T, path string) {
				if err := os.WriteFile(path, []byte("# alt"), 0o644); err != nil {
					t.Fatalf("Vorbereitung: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "Caddyfile")
			tt.prepare(t, path)

			if err := writeCaddyfile(path, "# gerendert"); err != nil {
				t.Fatalf("writeCaddyfile: %v", err)
			}

			info, err := os.Stat(path)
			if err != nil {
				t.Fatalf("Stat: %v", err)
			}
			// Bewusst der Literalwert statt caddyfileMode: ein Vergleich der
			// Konstante mit sich selbst würde jede Änderung durchlassen.
			if got := info.Mode().Perm(); got != 0o600 {
				t.Errorf("Modus = %04o, want 0600", got)
			}

			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("ReadFile: %v", err)
			}
			if string(content) != "# gerendert" {
				t.Errorf("Inhalt = %q, want %q", content, "# gerendert")
			}
		})
	}
}
