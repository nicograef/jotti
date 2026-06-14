package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"jotti-starter/core"
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
// aufsetzt. POSTGRES_USER ist fest "admin" (siehe core.EnvContent).
func dumpDatabase(name string) error {
	out, err := exec.Command("docker", "exec", postgresContainer,
		"pg_dump", "--clean", "--if-exists", "-U", "admin", "-d", "jotti",
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
