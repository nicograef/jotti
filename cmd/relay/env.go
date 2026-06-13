package main

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
)

// parseEnvFile liest eine .env im Key=Value-Format. Kommentare (# ...) und
// Leerzeilen werden ignoriert; Whitespace, CR (CRLF von Notepad) und optionale
// Anfuehrungszeichen um den Wert werden getrimmt; ein fuehrendes UTF-8-BOM wird
// toleriert. Zeilen ohne '=' werden uebersprungen.
func parseEnvFile(data []byte) map[string]string {
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
	values := make(map[string]string)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key != "" {
			values[key] = value
		}
	}
	return values
}

// loadEnvFile sucht eine .env neben der Programmdatei und im Arbeitsverzeichnis
// und gibt die geparsten Schluessel zurueck. Fehlt die Datei (oder ist sie nicht
// lesbar), kommt eine leere Map zurueck — der Doppelklick-Fallback ist optional,
// im Server-Betrieb zaehlen ohnehin die echten Env-Variablen.
func loadEnvFile() map[string]string {
	var dirs []string
	if exe, err := os.Executable(); err == nil {
		dirs = append(dirs, filepath.Dir(exe))
	}
	if wd, err := os.Getwd(); err == nil {
		dirs = append(dirs, wd)
	}

	for _, dir := range dirs {
		if data, err := os.ReadFile(filepath.Join(dir, ".env")); err == nil {
			return parseEnvFile(data)
		}
	}
	return map[string]string{}
}

// envWithFileFallback liefert eine getenv-Funktion, die echte Umgebungsvariablen
// bevorzugt und nur fehlende Werte aus der .env-Datei nachreicht. So behalten
// gesetzte Env-Variablen (Server-Betrieb via systemd o. Ae.) Vorrang vor der
// Datei.
func envWithFileFallback(fileValues map[string]string) func(string) string {
	return func(key string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		return fileValues[key]
	}
}
