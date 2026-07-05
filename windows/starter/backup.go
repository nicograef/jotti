package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/nicograef/jotti/windows/starter/core"
)

// postgresContainer ist der feste container_name des postgres-Service (siehe
// docker-compose.release.yml). pg_dump laeuft per `docker exec` direkt in diesem
// Container: lokale Socket-Verbindungen sind dort trust-authentifiziert, sodass
// fuer das Backup kein Passwort durchgereicht werden muss.
const postgresContainer = "jotti-postgres-local"

// postgresDataVolume ist das (projektpraefigierte) Daten-Volume. Sein Vorhandensein
// entscheidet mit, ob es ueberhaupt Daten zu sichern gibt.
const postgresDataVolume = "jotti-local_postgres-data"

// backupDir ist der Mountpunkt des jotti-backups-Volumes im postgres-Container.
const backupDir = "/jotti-backups"

// keptBackups ist die Anzahl der vorgehaltenen Pre-Update-Dumps; aeltere werden
// nach jedem neuen Dump rotiert, damit das Volume nicht unbegrenzt waechst.
const keptBackups = 5

// hostBackupDirName ist der Unterordner im Zustandsverzeichnis, in den jeder
// Pre-Update-Dump zusaetzlich gespiegelt wird (unter Windows
// %PROGRAMDATA%\jotti\backups). Anders als das Docker-Volume ueberlebt dieser
// Ordner ein `docker compose down -v` — er ist die zweite, unabhaengige Kopie.
const hostBackupDirName = "backups"

// dumpPrefix und dumpSuffix umrahmen die zeitgestempelten Backup-Dateinamen
// (jotti-YYYYMMDD-HHMMSS.sql). Erzeugung und Rotations-Filter teilen sie sich,
// damit der Filter nicht still aufhoert zu greifen, sollte sich das Namensschema
// einmal aendern.
const (
	dumpPrefix = "jotti-"
	dumpSuffix = ".sql"
)

// maybeBackupBeforeUpdate zieht vor dem vollen `up` (inkl. migrate) automatisch
// einen pg_dump, sobald bereits Daten existieren und die laufende Version von der
// zuletzt gesund gestarteten abweicht oder noch kein Marker vorliegt (Erst-Upgrade
// von einer Vor-Phase-3-Version — siehe core.ShouldBackup). Es faehrt dafuer nur
// postgres hoch, wartet auf gesund, dumpt zeitgestempelt ins jotti-backups-Volume
// und rotiert auf die neuesten keptBackups. Ein Fehler hier ist fatal fuer den
// Start — lieber nicht migrieren als ohne Sicherungspunkt migrieren. Bei gleicher
// Version, echter Erstinstallation oder Dev-Build kehrt die Funktion sofort zurueck.
func maybeBackupBeforeUpdate(composePath, envPath, stateDir string) error {
	lastVersion := readLastVersion(stateDir)
	dataExists, err := volumeExists(postgresDataVolume)
	if err != nil {
		return err
	}
	if !core.ShouldBackup(lastVersion, version, dataExists) {
		return nil
	}

	if lastVersion == "" {
		fmt.Printf("Erstes Upgrade erkannt (auf %s) — sichere die Daten vor dem Update ...\n", version)
	} else {
		fmt.Printf("Versionswechsel erkannt (%s → %s) — sichere die Daten vor dem Update ...\n", lastVersion, version)
	}
	if err := runCompose(os.Environ(), composePath, envPath, "up", "-d", "--wait", "postgres"); err != nil {
		return fmt.Errorf("postgres fuer das Backup hochfahren fehlgeschlagen: %w", err)
	}

	name := dumpPrefix + time.Now().Format("20060102-150405") + dumpSuffix
	if err := dumpDatabase(name); err != nil {
		return err
	}
	fmt.Printf("Backup erstellt: %s (im jotti-backups-Volume).\n", name)

	if err := rotateBackups(keptBackups); err != nil {
		// Rotation ist Hygiene, kein Grund den Start abzubrechen.
		fmt.Printf("Hinweis: alte Backups konnten nicht rotiert werden (%v).\n", err)
	}

	// Den Dump zusaetzlich auf den Host spiegeln. Der Dump im Volume existiert
	// bereits, daher ist ein Fehlschlag hier nur ein Hinweis, kein Startabbruch:
	// so vernichtet ein spaeteres `docker compose down -v` nicht Daten und
	// Backups zugleich.
	hostDir := filepath.Join(stateDir, hostBackupDirName)
	if err := mirrorBackupToHost(name, hostDir); err != nil {
		fmt.Printf("Hinweis: Backup konnte nicht nach %s gespiegelt werden (%v).\n", hostDir, err)
	} else {
		fmt.Printf("Backup zusaetzlich gesichert in: %s\n", hostDir)
	}
	return nil
}

// readLastVersion liest den last-version-Marker. Fehlt er oder ist er unlesbar,
// gilt das als "keine bekannte Vorversion" (leerer String) — dann wird nicht
// gesichert (Erststart).
func readLastVersion(stateDir string) string {
	data, err := os.ReadFile(filepath.Join(stateDir, lastVersionFile))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// volumeExists meldet, ob ein benanntes Docker-Volume existiert. Nur ein echter
// Docker-Fehler (Daemon nicht erreichbar o. Ae.) wird durchgereicht; ein fehlendes
// Volume ist der regulaere "nein"-Fall.
func volumeExists(name string) (bool, error) {
	err := exec.Command("docker", "volume", "inspect", name).Run()
	if err == nil {
		return true, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return false, nil
	}
	return false, fmt.Errorf("docker volume inspect fehlgeschlagen: %w", err)
}

// dumpDatabase schreibt einen vollstaendigen pg_dump per `docker exec` in das im
// postgres-Container gemountete jotti-backups-Volume. --clean --if-exists setzt
// DROP-Anweisungen voran, damit ein spaeterer Restore die Objekte sauber neu
// aufsetzt. Die Rolle stammt aus core.PostgresUser — derselben Quelle, aus der
// EnvContent POSTGRES_USER in die .env schreibt, damit beide nie auseinanderlaufen.
func dumpDatabase(name string) error {
	out, err := exec.Command("docker", "exec", postgresContainer,
		"pg_dump", "--clean", "--if-exists", "-U", core.PostgresUser, "-d", "jotti",
		"-f", backupDir+"/"+name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("pg_dump fehlgeschlagen: %v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// rotateBackups loescht alle bis auf die neuesten keep Dumps im
// jotti-backups-Volume. Gefiltert wird auf die zeitgestempelten jotti-*.sql-Dumps,
// damit nichts anderes im Volume angetastet wird.
func rotateBackups(keep int) error {
	out, err := exec.Command("docker", "exec", postgresContainer, "ls", "-1", backupDir).Output()
	if err != nil {
		return fmt.Errorf("auflisten der Backups fehlgeschlagen: %w", err)
	}
	var names []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		name := strings.TrimSpace(line)
		if strings.HasPrefix(name, dumpPrefix) && strings.HasSuffix(name, dumpSuffix) {
			names = append(names, name)
		}
	}
	for _, name := range core.DumpsToDelete(names, keep) {
		if err := exec.Command("docker", "exec", postgresContainer, "rm", "-f", backupDir+"/"+name).Run(); err != nil {
			return fmt.Errorf("altes Backup %s loeschen fehlgeschlagen: %w", name, err)
		}
	}
	return nil
}

// mirrorBackupToHost kopiert den frisch erstellten Dump per `docker cp` aus dem
// postgres-Container in hostDir und rotiert dort auf die neuesten keptBackups
// Dateien. core.PlanBackupMirror entscheidet rein, was zu kopieren und zu
// loeschen ist; diese Funktion fuehrt nur die Seiteneffekte aus. Der Aufrufer
// behandelt einen Fehler hier als Hinweis, nicht als Startabbruch.
func mirrorBackupToHost(name, hostDir string) error {
	if err := os.MkdirAll(hostDir, 0o755); err != nil {
		return fmt.Errorf("Backup-Ordner %s anlegen fehlgeschlagen: %w", hostDir, err)
	}
	existing, err := listHostBackups(hostDir)
	if err != nil {
		return err
	}
	plan := core.PlanBackupMirror(name, existing, keptBackups)

	if plan.Copy != "" {
		out, err := exec.Command("docker", "cp",
			postgresContainer+":"+backupDir+"/"+plan.Copy, filepath.Join(hostDir, plan.Copy)).CombinedOutput()
		if err != nil {
			return fmt.Errorf("docker cp des Backups auf den Host fehlgeschlagen: %v: %s", err, strings.TrimSpace(string(out)))
		}
	}
	for _, del := range plan.Delete {
		if err := os.Remove(filepath.Join(hostDir, del)); err != nil {
			return fmt.Errorf("altes Host-Backup %s loeschen fehlgeschlagen: %w", del, err)
		}
	}
	return nil
}

// listHostBackups liefert die zeitgestempelten jotti-*.sql-Dumps in dir. Ein noch
// fehlender Ordner gilt als leer (kein Fehler); alles ausserhalb des
// Namensschemas bleibt unangetastet.
func listHostBackups(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("Host-Backups auflisten fehlgeschlagen: %w", err)
	}
	var names []string
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, dumpPrefix) && strings.HasSuffix(name, dumpSuffix) {
			names = append(names, name)
		}
	}
	return names, nil
}
