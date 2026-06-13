package core

import (
	"path/filepath"
	"strings"
)

// StateDir liefert das Host-Zustandsverzeichnis fuer den .env-Spiegel, den
// last-version-Marker und exportierte Backups. Unter Windows ist das kanonisch
// %PROGRAMDATA%\jotti — unabhaengig davon, wohin der Nutzer das ZIP entpackt.
// Damit entfaellt die fehleranfaellige "ueber denselben Ordner entpacken"-Regel:
// der Zustand lebt an einem festen Ort, die Programmdateien duerfen irgendwo
// liegen. Sonst (Linux-Dev oder Windows ohne gesetztes PROGRAMDATA) bleibt der
// Zustand ordnerlokal im uebergebenen fallback — wie bisher.
func StateDir(goos, programData, fallback string) string {
	if goos == "windows" && programData != "" {
		return filepath.Join(programData, "jotti")
	}
	return fallback
}

// EnvContent erzeugt den vollstaendigen .env-Inhalt mit frisch erzeugten
// Secrets. POSTGRES_USER ist fest "admin" (wie .env.example); die drei Secrets
// stammen aus GenerateSecret. Der Kommentar-Header haelt die erste Zeile frei
// von einem Key: Schreibt Notepad spaeter ein UTF-8-BOM in die Datei, landet es
// so vor einem Kommentar statt vor POSTGRES_USER und beschaedigt keinen Key.
func EnvContent() string {
	lines := []string{
		"# Diese Datei wurde automatisch von jotti erzeugt. Hier muss nichts geaendert werden.",
		"POSTGRES_USER=admin",
		"POSTGRES_PASSWORD=" + GenerateSecret(),
		"JWT_SECRET=" + GenerateSecret(),
		"RELAY_AUTH_TOKEN=" + GenerateSecret(),
		"",
	}
	return strings.Join(lines, "\n")
}

// MaterializeEnv schreibt die .env nach path, falls sie noch nicht existiert,
// und meldet ueber created, ob geschrieben wurde. Eine vorhandene Datei wird
// nie ueberschrieben (idempotent wie scripts/init-env.sh) — die Secrets werden
// dann gar nicht erst erzeugt. exists und write kapseln die Dateizugriffe, damit
// die Funktion ohne echtes Dateisystem testbar ist.
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

// EnvResolution ist das Ergebnis der Secret-Discovery: welchen .env-Inhalt der
// Start verwenden soll und wie damit zu verfahren ist. Abort schliesst die anderen
// Felder aus — ist es gesetzt, wurde kein Secret gefunden, obwohl Daten existieren,
// und der Start muss abbrechen, ohne irgendetwas zu veraendern.
type EnvResolution struct {
	Content string // zu verwendender .env-Inhalt (leer, wenn Abort)
	Seed    bool   // Content muss noch ins jotti-config-Volume geschrieben werden
	Abort   bool   // kein Secret gefunden, aber postgres-data vorhanden → abbrechen
}

// ResolveEnv waehlt das Install-Secret aus der ersten nicht-leeren Quelle in fester
// Reihenfolge: zuerst das jotti-config-Volume (volumeContent) — die Vorwaerts-Quelle
// der Wahrheit, deren Inhalt unveraendert uebernommen wird (Seed false), damit der
// Schluessel nie von den Daten abweicht, die er entsperrt. Ist das Volume leer,
// gewinnt der erste nicht-leere ordnerlokale Kandidat (localCandidates, in
// Prioritaetsreihenfolge: Host-Spiegel unter %PROGRAMDATA%\jotti, dann .env neben
// Compose/Exe, wie eine Version vor dem Volume sie hinterlassen hat); er wird
// adoptiert und ins Volume geschrieben (Seed true).
//
// Findet sich nirgends ein Secret, entscheidet postgresDataExists ueber den
// Fail-Safe: existieren bereits Daten, wird abgebrochen (Abort), statt frische
// Secrets neben vorhandene Daten zu erzeugen und damit das alte Passwort
// auszusperren. Nur bei echter Erstinstallation (keine Daten) werden frische
// Secrets erzeugt (Seed true) — wie bisher.
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
