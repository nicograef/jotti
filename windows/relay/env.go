package main

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// parseEnvFile liest eine .env im Key=Value-Format; CR (CRLF von Notepad),
// optionale Anfuehrungszeichen und ein fuehrendes UTF-8-BOM werden toleriert.
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

// loadEnvFile liest die erste .env aus envSearchDirs; fehlt sie, kommt eine leere
// Map zurueck — der Datei-Fallback ist optional.
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

// envSearchDirs liefert die .env-Suchverzeichnisse in Prioritaetsreihenfolge. Unter
// Windows zuerst %PROGRAMDATA%\jotti — dorthin schreibt jotti-start.exe den
// .env-Spiegel, den das nicht-elevierte Relay aus einem anderen Verzeichnis sonst
// nicht faende; danach Programm- und Arbeitsverzeichnis.
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

func envHinweis() string {
	if runtime.GOOS == "windows" {
		return "Bitte zuerst jotti-start.exe ausfuehren - sie erzeugt die Zugangsdaten."
	}
	return "Bitte RELAY_AUTH_TOKEN in der .env-Datei neben jotti-relay.exe setzen."
}

func envWithFileFallback(fileValues map[string]string) func(string) string {
	return func(key string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		return fileValues[key]
	}
}
