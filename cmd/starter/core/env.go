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

// ResolveEnv entscheidet, welchen .env-Inhalt der Start verwenden soll. Quelle
// der Wahrheit ist das jotti-config-Volume: liegt dort bereits ein Secret
// (volumeContent), wird es unveraendert uebernommen — so kann der Schluessel nie
// von den Daten abweichen, die er entsperrt. Beim Erststart ist das Volume leer;
// dann wird ein vorhandener ordnerlokaler .env-Inhalt (localContent) adoptiert
// (Upgrade einer Installation aus der Zeit vor dem Volume, deren Secret weiter
// zur bestehenden postgres-data passen muss), sonst werden frische Secrets
// erzeugt. seed meldet, ob der gewaehlte Inhalt noch ins Volume geschrieben
// werden muss (false nur, wenn er von dort stammt).
func ResolveEnv(volumeContent, localContent string) (content string, seed bool) {
	if strings.TrimSpace(volumeContent) != "" {
		return volumeContent, false
	}
	if strings.TrimSpace(localContent) != "" {
		return localContent, true
	}
	return EnvContent(), true
}
