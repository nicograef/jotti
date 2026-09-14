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

// postgresContainer ist der feste container_name des postgres-Service. pg_dump
// laeuft per `docker exec` darin: lokale Socket-Verbindungen sind dort
// trust-authentifiziert, also braucht das Backup kein Passwort.
const postgresContainer = "jotti-postgres-local"

const postgresDataVolume = "jotti-local_postgres-data"

// backupDir ist der Mountpunkt des jotti-backups-Volumes im postgres-Container.
const backupDir = "/jotti-backups"

const keptBackups = 5

// hostBackupDirName ist der Unterordner im Zustandsverzeichnis (unter Windows
// %PROGRAMDATA%\jotti\backups), in den jeder Pre-Update-Dump gespiegelt wird. Er
// ueberlebt ein `docker compose down -v` — die zweite, unabhaengige Kopie.
const hostBackupDirName = "backups"

// dumpPrefix und dumpSuffix umrahmen die zeitgestempelten Dateinamen
// (jotti-YYYYMMDD-HHMMSS.sql). Erzeugung und Rotations-Filter teilen sie sich,
// damit der Filter nie still aufhoert zu greifen.
const (
	dumpPrefix = "jotti-"
	dumpSuffix = ".sql"
)

// maybeBackupBeforeUpdate zieht vor dem vollen `up` (inkl. migrate) einen pg_dump,
// sobald core.ShouldBackup es verlangt: postgres hochfahren, zeitgestempelt ins
// jotti-backups-Volume dumpen, auf keptBackups rotieren, auf den Host spiegeln.
// Ein Fehler ist fatal — lieber nicht migrieren als ohne Sicherungspunkt migrieren.
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
		fmt.Printf("Erstes Upgrade erkannt (auf %s) - sichere die Daten vor dem Update ...\n", version)
	} else {
		fmt.Printf("Versionswechsel erkannt (%s -> %s) - sichere die Daten vor dem Update ...\n", lastVersion, version)
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

	// Fehlschlag ist nur ein Hinweis, kein Startabbruch: der Dump im Volume existiert
	// bereits.
	hostDir := filepath.Join(stateDir, hostBackupDirName)
	if err := mirrorBackupToHost(name, hostDir); err != nil {
		fmt.Printf("Hinweis: Backup konnte nicht nach %s gespiegelt werden (%v).\n", hostDir, err)
	} else {
		fmt.Printf("Backup zusaetzlich gesichert in: %s\n", hostDir)
	}
	return nil
}

// readLastVersion liest den last-version-Marker; fehlt er oder ist er unlesbar,
// gilt das als "keine bekannte Vorversion" (leerer String).
func readLastVersion(stateDir string) string {
	data, err := os.ReadFile(filepath.Join(stateDir, lastVersionFile))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// volumeExists meldet, ob ein benanntes Docker-Volume existiert. Ein fehlendes
// Volume ist der regulaere "nein"-Fall; nur ein echter Docker-Fehler wird
// durchgereicht.
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

// dumpDatabase schreibt einen pg_dump per `docker exec` ins jotti-backups-Volume.
// --clean --if-exists setzt DROP-Anweisungen voran, damit ein Restore die Objekte
// sauber neu aufsetzt; die Rolle kommt aus core.PostgresUser — derselben Quelle wie
// POSTGRES_USER in der .env.
func dumpDatabase(name string) error {
	out, err := exec.Command("docker", "exec", postgresContainer,
		"pg_dump", "--clean", "--if-exists", "-U", core.PostgresUser, "-d", "jotti",
		"-f", backupDir+"/"+name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("pg_dump fehlgeschlagen: %v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// rotateBackups loescht alle bis auf die neuesten keep Dumps. Gefiltert wird auf
// jotti-*.sql, damit nichts anderes im Volume angetastet wird.
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

// mirrorBackupToHost kopiert den frischen Dump per `docker cp` nach hostDir und
// rotiert dort auf keptBackups Dateien.
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

// listHostBackups liefert die jotti-*.sql-Dumps in dir; ein fehlender Ordner gilt
// als leer.
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
