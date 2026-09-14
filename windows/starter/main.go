// Command jotti-start ist der klickbare Windows-Starter fuer den lokalen
// jotti-Betrieb (requireAdministrator-Manifest, eine UAC-Abfrage pro Start).
// Host-Zustand (.env-Spiegel, last-version-Marker) liegt unter Windows kanonisch in
// %PROGRAMDATA%\jotti — unabhaengig vom Entpack-Ort. Die reine Logik liegt in
// windows/starter/core; die Windows-Schritte laufen nur unter GOOS == "windows",
// der Repo-Dev-Lauf unter Linux ueberspringt sie und bleibt ordnerlokal.
package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/nicograef/jotti/windows/starter/core"
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
			if errors.Is(err, errSecretFehltMitDaten) {
				fmt.Println(core.DiagnoseSecretFehltMitDaten)
			} else {
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
		// Linux-Dev-Lauf: ohne Daemon-Garantie und ohne Volume bleibt die .env ordnerlokal.
		created, err := core.MaterializeEnv(envPath, fileExists, writeEnvFile)
		if err != nil {
			fmt.Printf("Konfiguration (.env) konnte nicht erstellt werden: %v\n", err)
			return 1
		}
		if created {
			fmt.Println("Konfiguration (.env) mit frischen Zugangsdaten erstellt.")
		}
	}

	// Downgrade verweigern: eine aeltere Exe darf nicht gegen neuere Daten starten.
	if lv := readLastVersion(stateDir); core.IsDowngrade(version, lv) {
		fmt.Printf("Start verweigert: Diese Version (%s) ist aelter als die zuletzt gestartete (%s).\n", version, lv)
		fmt.Println("  Neuere Daten lassen sich nicht auf aeltere Versionen zurueckrollen.")
		fmt.Println("  Neue Version herunterladen: https://github.com/nicograef/jotti/releases/latest")
		return 1
	}

	// Vor dem vollen `up` sichern: der Sicherungspunkt muss vor jeder
	// schemaveraendernden Migration entstehen.
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

	// Erst nach gesundem Stack festhalten — der Marker steuert das Pre-Update-Backup
	// beim naechsten Start (siehe core.ShouldBackup).
	if err := writeLastVersion(stateDir); err != nil {
		fmt.Printf("Hinweis: Versionsmarker konnte nicht geschrieben werden (%v).\n", err)
	}

	printSuccess()

	printAdminCode(composePath)

	notifyIfUpdateAvailable()
	return 0
}

const lastVersionFile = "last-version"

func resolveStateDir(fallback string) (string, error) {
	dir := core.StateDir(runtime.GOOS, os.Getenv("PROGRAMDATA"), fallback)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func writeLastVersion(stateDir string) error {
	return os.WriteFile(filepath.Join(stateDir, lastVersionFile), []byte(version+"\n"), 0o644)
}

// envCandidateDirs liefert die .env-Suchverzeichnisse: Zustandsverzeichnis, Ordner
// der Compose-Datei, Ordner der Programmdatei. Bewusst NICHT das Arbeitsverzeichnis
// — nach der UAC-Elevation ist das C:\Windows\System32.
func envCandidateDirs(stateDir, composeDir string) []string {
	dirs := []string{stateDir, composeDir}
	if exe, err := os.Executable(); err == nil {
		dirs = append(dirs, filepath.Dir(exe))
	}
	return dirs
}

// resolveComposeFile sucht die Compose-Datei relativ zur Programmdatei — nach der
// UAC-Elevation ist das Arbeitsverzeichnis C:\Windows\System32. Im Repo-Dev-Lauf
// faellt die Suche aufs Arbeitsverzeichnis zurueck.
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

func writeEnvFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0o600)
}

// fileExists meldet, ob path existiert; ein echter Stat-Fehler wird durchgereicht,
// damit er nicht als "fehlt" gewertet wird.
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

// printSuccess verweist auf die Status-Seite statt auf eine eigene URL: der Starter
// kennt die Install-ID nicht.
func printSuccess() {
	fmt.Println()
	fmt.Printf("jotti Starter %s - jotti laeuft.\n\n", version)
	fmt.Println("Status & Zugangsadresse: http://localhost:8484")
	fmt.Println("  Dort stehen die Zugangsadresse fuers WLAN und ein QR-Code fuer die Helfer-Handys.")
	if runtime.GOOS == "windows" {
		fmt.Println("Firewall-Freigabe fuers lokale Netzwerk ist eingerichtet.")
	}
	fmt.Println()
	fmt.Println("SICHERHEIT: jotti niemals ins Internet oeffnen (keine Port-Weiterleitung im Router).")
}

func waitForEnter() {
	fmt.Print("\nEnter druecken zum Schliessen ...")
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
}
