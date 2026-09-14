package core

import (
	"path/filepath"
	"strings"
)

// StateDir liefert das Host-Zustandsverzeichnis fuer .env-Spiegel,
// last-version-Marker und exportierte Backups: unter Windows kanonisch
// %PROGRAMDATA%\jotti, unabhaengig vom Entpack-Ort; sonst der uebergebene fallback.
func StateDir(goos, programData, fallback string) string {
	if goos == "windows" && programData != "" {
		return filepath.Join(programData, "jotti")
	}
	return fallback
}

// PostgresUser ist der Postgres-Rollenname und die einzige Quelle der Wahrheit
// dafuer: EnvContent schreibt ihn in die .env, das Pre-Update-Backup dumpt als
// dieser Rolle. Wert wie .env.example / scripts/init-env.sh.
const PostgresUser = "admin"

// EnvContent erzeugt den .env-Inhalt mit frisch erzeugten Secrets. Die erste Zeile
// ist bewusst ein Kommentar: schreibt Notepad spaeter ein UTF-8-BOM, landet es so
// vor dem Kommentar statt vor einem Key.
func EnvContent() string {
	lines := []string{
		"# Diese Datei wurde automatisch von jotti erzeugt. Hier muss nichts geaendert werden.",
		"POSTGRES_USER=" + PostgresUser,
		"POSTGRES_PASSWORD=" + GenerateSecret(),
		"JWT_SECRET=" + GenerateSecret(),
		"RELAY_AUTH_TOKEN=" + GenerateSecret(),
		"",
	}
	return strings.Join(lines, "\n")
}

// MaterializeEnv schreibt die .env nach path, falls sie noch nicht existiert. Eine
// vorhandene Datei wird nie ueberschrieben — die Secrets werden dann gar nicht erst
// erzeugt.
func MaterializeEnv(path string, exists func(string) (bool, error), write func(string, []byte) error) (created bool, err error) {
	present, err := exists(path)
	if err != nil {
		return false, err
	}
	if present {
		return false, nil
	}
	if err := write(path, []byte(EnvContent())); err != nil {
		return false, err
	}
	return true, nil
}

// EnvResolution ist das Ergebnis der Secret-Discovery. Abort schliesst die anderen
// Felder aus: kein Secret gefunden, obwohl Daten existieren — der Start muss
// abbrechen, ohne etwas zu veraendern.
type EnvResolution struct {
	Content string // zu verwendender .env-Inhalt (leer, wenn Abort)
	Seed    bool   // Content muss noch ins jotti-config-Volume geschrieben werden
	Abort   bool   // kein Secret gefunden, aber postgres-data vorhanden → abbrechen
}

// ResolveEnv waehlt das Install-Secret aus der ersten nicht-leeren Quelle: zuerst
// das jotti-config-Volume, dessen Inhalt unveraendert uebernommen wird (Seed false),
// damit der Schluessel nie von den Daten abweicht, die er entsperrt. Sonst gewinnt
// der erste nicht-leere Kandidat aus localCandidates (Prioritaet: Host-Spiegel unter
// %PROGRAMDATA%\jotti, dann .env neben Compose/Exe); er wird adoptiert und ins
// Volume geschrieben (Seed true). Findet sich nirgends ein Secret, entscheidet
// postgresDataExists: mit Daten Abort, statt sie mit frischen Secrets auszusperren;
// ohne Daten frische Secrets (Seed true).
func ResolveEnv(volumeContent string, localCandidates []string, postgresDataExists bool) EnvResolution {
	if strings.TrimSpace(volumeContent) != "" {
		return EnvResolution{Content: volumeContent, Seed: false}
	}
	for _, candidate := range localCandidates {
		if strings.TrimSpace(candidate) != "" {
			return EnvResolution{Content: candidate, Seed: true}
		}
	}
	if postgresDataExists {
		return EnvResolution{Abort: true}
	}
	return EnvResolution{Content: EnvContent(), Seed: true}
}
