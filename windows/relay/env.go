package main

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"runtime"
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

// loadEnvFile sucht die .env in den Verzeichnissen aus envSearchDirs und gibt die
// geparsten Schluessel zurueck. Fehlt die Datei (oder ist sie nicht lesbar), kommt
// eine leere Map zurueck — der Doppelklick-Fallback ist optional, im Server-Betrieb
// zaehlen ohnehin die echten Env-Variablen.
func loadEnvFile() map[string]string {
	exeDir := ""
	if exe, err := os.Executable(); err == nil {
		exeDir = filepath.Dir(exe)
	}
	wd, _ := os.Getwd()
	for _, dir := range envSearchDirs(runtime.GOOS, os.Getenv("PROGRAMDATA"), exeDir, wd) {
		if data, err := os.ReadFile(filepath.Join(dir, ".env")); err == nil {
			return parseEnvFile(data)
		}
	}
	return map[string]string{}
}

// envSearchDirs liefert die .env-Suchverzeichnisse in Prioritaetsreihenfolge.
// Unter Windows steht das kanonische %PROGRAMDATA%\jotti zuerst — dorthin schreibt
// jotti-start.exe den .env-Spiegel, sodass das Relay die Zugangsdaten unabhaengig
// vom eigenen Ordner findet (es laeuft nicht-eleviert und evtl. aus einem anderen
// Verzeichnis). Danach (und unter Linux ausschliesslich) der eigene Programmordner
// und das Arbeitsverzeichnis. Reine Funktion: die echten OS-Werte reicht
// loadEnvFile ein, damit die Reihenfolge testbar bleibt.
func envSearchDirs(goos, programData, exeDir, wd string) []string {
	var dirs []string
	if goos == "windows" && programData != "" {
		dirs = append(dirs, filepath.Join(programData, "jotti"))
	}
	if exeDir != "" {
		dirs = append(dirs, exeDir)
	}
	if wd != "" {
		dirs = append(dirs, wd)
	}
	return dirs
}

// envHinweis liefert den OS-spezifischen Hinweis, woher die Zugangsdaten kommen,
// wenn die Konfiguration fehlt. Unter Windows schreibt jotti-start.exe die .env
// nach %PROGRAMDATA%\jotti; fehlt sie, lief der Starter schlicht noch nicht. Unter
// Linux (Server/Dev) zaehlen echte Env-Variablen bzw. eine .env neben der
// Programmdatei.
func envHinweis() string {
	if runtime.GOOS == "windows" {
		return "Bitte zuerst jotti-start.exe ausfuehren — sie erzeugt die Zugangsdaten."
	}
	return "Bitte RELAY_AUTH_TOKEN in der .env-Datei neben jotti-relay.exe setzen."
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
