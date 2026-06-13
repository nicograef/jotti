// Command jotti-start ist der klickbare Windows-Starter fuer den lokalen
// jotti-Betrieb. Ein Doppelklick (mit Administratorrechten, requireAdministrator-
// Manifest) raeumt die Docker-Voraussetzungen aus (Daemon-Start, Linux-Engine),
// holt das Install-Secret aus dem jotti-config-Volume (oder erzeugt es beim
// Erststart) und spiegelt es in die .env, prueft die Ports, gibt die Firewall
// frei, faehrt den Compose-Stack hoch, wartet, bis jotti unter /api/health bereit
// ist, und weist danach (non-fatal, online) auf eine neuere Version hin. Host-
// Zustand (.env-Spiegel, last-version-Marker) liegt unter
// Windows kanonisch in %PROGRAMDATA%\jotti — unabhaengig vom Entpack-Ort. Die
// reine Logik liegt in cmd/starter/core; diese Datei verbindet sie mit den echten
// Seiteneffekten. Alle Windows-spezifischen Schritte laufen nur unter
// runtime.GOOS == "windows" — der Repo-Dev-Lauf unter Linux ueberspringt sie und
// haelt den Zustand weiterhin ordnerlokal.
package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"jotti-starter/core"
)

// version wird beim Release per -ldflags "-X main.version=vX.Y.Z" gesetzt.
var version = "dev"

const (
	releaseComposeFile = "docker-compose.release.yml"
	localComposeFile   = "docker-compose.local.yml"
)

func main() {
	code := run()
	// Ein per Doppelklick gestartetes Konsolenfenster schliesst beim Exit sofort
	// — bei Erfolg wie Fehler auf Enter warten, damit die Ausgabe lesbar bleibt.
	waitForEnter()
	os.Exit(code)
}

// run fuehrt den gesamten Startablauf aus und liefert den Exit-Code (0 = Erfolg,
// 1 = Preflight- oder Health-Fehler).
func run() int {
	fmt.Printf("jotti Starter %s\n\n", version)

	composePath, err := resolveComposeFile()
	if err != nil {
		fmt.Println(err)
		return 1
	}
	fmt.Printf("Compose-Datei: %s\n", composePath)

	stateDir, err := resolveStateDir(filepath.Dir(composePath))
	if err != nil {
		fmt.Printf("Zustandsverzeichnis konnte nicht angelegt werden: %v\n", err)
		return 1
	}
	envPath := filepath.Join(stateDir, ".env")

	if runtime.GOOS == "windows" {
		// Reihenfolge bewusst: ensureDocker zuerst, weil der Secret-Read aus dem
		// Volume einen laufenden Daemon braucht; danach die Host-.env spiegeln,
		// damit `compose --env-file` sie interpolieren kann.
		if msg := ensureDocker(); msg != "" {
			fmt.Println(msg)
			return 1
		}
		if err := materializeEnvFromVolume(envPath, envCandidateDirs(stateDir, filepath.Dir(composePath))); err != nil {
			// Den Fail-Safe-Abbruch hat materializeEnvFromVolume bereits ausfuehrlich
			// gemeldet; nur echte Fehler bekommen hier ein Praefix.
			if !errors.Is(err, errSecretFehltMitDaten) {
				fmt.Printf("Zugangsdaten konnten nicht bereitgestellt werden: %v\n", err)
			}
			return 1
		}
		if msg := checkPorts(composePath); msg != "" {
			fmt.Println(msg)
			return 1
		}
		ensureFirewall()
	} else {
		// Linux-Dev-Lauf: ohne Docker-Daemon-Garantie und ohne Volume bleibt die
		// .env ordnerlokal und wird nur erzeugt, wenn sie fehlt (wie bisher).
		created, err := core.MaterializeEnv(envPath, fileExists, writeEnvFile)
		if err != nil {
			fmt.Printf("Konfiguration (.env) konnte nicht erstellt werden: %v\n", err)
			return 1
		}
		if created {
			fmt.Println("Konfiguration (.env) mit frischen Zugangsdaten erstellt.")
		}
	}

	// Vor dem vollen `up` (das die Migrationen anstoesst) bei einem
	// Versionswechsel automatisch die Daten sichern — der Sicherungspunkt
	// entsteht so vor jeder schemaveraendernden Migration.
	if err := maybeBackupBeforeUpdate(composePath, envPath, stateDir); err != nil {
		fmt.Printf("Automatisches Pre-Update-Backup fehlgeschlagen: %v\n", err)
		return 1
	}

	lanIP := detectLANIP()
	if err := composeUp(composePath, envPath, lanIP); err != nil {
		fmt.Printf("Der jotti-Stack konnte nicht gestartet werden: %v\n", err)
		return 1
	}

	if err := waitForHealth(); err != nil {
		fmt.Printf("\n%v\n", err)
		return 1
	}

	// Erst nach gesundem Stack festhalten, welche Version zuletzt lief — der
	// Marker steuert in Phase 3 das automatische Pre-Update-Backup.
	if err := writeLastVersion(stateDir); err != nil {
		fmt.Printf("Hinweis: Versionsmarker konnte nicht geschrieben werden (%v).\n", err)
	}

	printSuccess()

	// Nach gesundem Start kurz online pruefen, ob eine neuere Version vorliegt, und
	// nur darauf hinweisen. Non-fatal: offline/Timeout/Fehler ueberspringen still.
	notifyIfUpdateAvailable()
	return 0
}

// lastVersionFile haelt im Zustandsverzeichnis fest, welche jotti-Version zuletzt
// gesund gestartet ist.
const lastVersionFile = "last-version"

// resolveStateDir bestimmt das Host-Zustandsverzeichnis und legt es an. Unter
// Windows ist das %PROGRAMDATA%\jotti, das nicht zwangslaeufig existiert; unter
// Linux-Dev ist es der bereits vorhandene ordnerlokale fallback (MkdirAll ist
// dann ein No-op).
func resolveStateDir(fallback string) (string, error) {
	dir := core.StateDir(runtime.GOOS, os.Getenv("PROGRAMDATA"), fallback)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// writeLastVersion schreibt die aktuelle Version in den last-version-Marker. Phase
// 3 vergleicht ihn beim naechsten Start mit der eigenen Version und zieht bei
// einem Wechsel vor den Migrationen automatisch ein Backup. Ein Schreibfehler ist
// nicht fatal: er kostet hoechstens dieses Backup, nie den laufenden Start.
func writeLastVersion(stateDir string) error {
	return os.WriteFile(filepath.Join(stateDir, lastVersionFile), []byte(version+"\n"), 0o644)
}

// envCandidateDirs liefert die Verzeichnisse, in denen materializeEnvFromVolume nach
// einer bestehenden .env sucht (Reihenfolge: kanonisches Zustandsverzeichnis, Ordner
// der Compose-Datei, Ordner der Programmdatei). Bewusst NICHT das Arbeitsverzeichnis:
// nach der UAC-Elevation ist das C:\Windows\System32 (vgl. resolveComposeFile) — die
// alte ordnerlokale .env liegt neben Compose/Exe, nicht dort.
func envCandidateDirs(stateDir, composeDir string) []string {
	dirs := []string{stateDir, composeDir}
	if exe, err := os.Executable(); err == nil {
		dirs = append(dirs, filepath.Dir(exe))
	}
	return dirs
}

// resolveComposeFile sucht die Compose-Datei immer relativ zur Programmdatei —
// nach der UAC-Elevation ist das Arbeitsverzeichnis C:\Windows\System32, nicht
// der Programmordner. Im Release-ZIP liegt docker-compose.release.yml neben der
// Exe; im Repo-Dev-Lauf (go run) faellt die Suche aufs Arbeitsverzeichnis zurueck.
func resolveComposeFile() (string, error) {
	var dirs []string
	if exe, err := os.Executable(); err == nil {
		dirs = append(dirs, filepath.Dir(exe))
	}
	if wd, err := os.Getwd(); err == nil {
		dirs = append(dirs, wd)
	}

	for _, dir := range dirs {
		for _, name := range []string{releaseComposeFile, localComposeFile} {
			path := filepath.Join(dir, name)
			if ok, _ := fileExists(path); ok {
				return path, nil
			}
		}
	}

	return "", fmt.Errorf("keine Compose-Datei gefunden (%s oder %s neben der Programmdatei erwartet). "+
		"Bitte das vollstaendige jotti-ZIP entpacken und jotti-start.exe daraus starten",
		releaseComposeFile, localComposeFile)
}

// writeEnvFile passt os.WriteFile an die von core.MaterializeEnv erwartete
// Signatur an und schreibt die Secrets nur fuer den Eigentuemer lesbar.
func writeEnvFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0o600)
}

// fileExists meldet, ob path existiert; ein echter Stat-Fehler (z. B. fehlende
// Rechte) wird durchgereicht, damit MaterializeEnv ihn nicht als "fehlt" wertet.
func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

// printSuccess gibt die Erfolgsmeldung aus: Version, Verweis auf die Status-Seite
// (dort stehen die gruene Adresse und der QR-Code — der Starter kennt die
// Install-ID nicht und baut keine eigene URL), Firewall-Bestaetigung und die
// Sicherheitswarnung.
func printSuccess() {
	fmt.Println()
	fmt.Printf("jotti Starter %s — jotti laeuft.\n\n", version)
	fmt.Println("Status & Zugangsadresse: http://localhost:8484")
	fmt.Println("  Dort stehen die Zugangsadresse fuers WLAN und ein QR-Code fuer die Helfer-Handys.")
	if runtime.GOOS == "windows" {
		fmt.Println("Firewall-Freigabe fuers lokale Netzwerk ist eingerichtet.")
	}
	fmt.Println()
	fmt.Println("SICHERHEIT: jotti niemals ins Internet oeffnen (keine Port-Weiterleitung im Router).")
}

// waitForEnter haelt das Doppelklick-Fenster offen, bis der Nutzer Enter drueckt.
func waitForEnter() {
	fmt.Print("\nEnter druecken zum Schliessen ...")
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
}
