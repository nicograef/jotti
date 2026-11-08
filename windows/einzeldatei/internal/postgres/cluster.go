// Package postgres steuert den eingebetteten Postgres-Cluster über initdb und pg_ctl.
// pg_ctl startet den Postmaster unter Administratorrechten mit eingeschränktem Token.
package postgres

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

// Cluster ist ein Datenverzeichnis samt Binaries; Postgres lauscht nur auf 127.0.0.1:Port.
type Cluster struct {
	Bin  string // Ordner mit initdb, pg_ctl und postgres
	Data string // Datenverzeichnis (pgdata)
	Port int
}

// Initialized meldet, ob Data schon einen Cluster trägt.
func (c Cluster) Initialized() bool {
	_, err := os.Stat(filepath.Join(c.Data, "PG_VERSION"))
	return err == nil
}

// Init legt den Cluster mit ICU-Sortierung en-US und SCRAM-Anmeldung an.
func (c Cluster) Init(user, password string) error {
	pwfile, err := os.CreateTemp("", "jotti-pw-*")
	if err != nil {
		return err
	}
	defer os.Remove(pwfile.Name()) //nolint:errcheck // Aufräumen nach initdb
	if _, err := pwfile.WriteString(password); err != nil {
		_ = pwfile.Close()
		return err
	}
	if err := pwfile.Close(); err != nil {
		return err
	}
	return c.run("initdb",
		"-D", c.Data,
		"-U", user,
		"--pwfile="+pwfile.Name(),
		"-E", "UTF8",
		"--locale=C",
		"--locale-provider=icu",
		"--icu-locale=en-US",
		"--auth=scram-sha-256",
	)
}

// Start startet den Postmaster und wartet, bis er Verbindungen annimmt.
// Einen verwaisten Postmaster, etwa nach einem abgestürzten Lauf, stoppt Start vorher mit -m fast.
func (c Cluster) Start() error {
	if c.Running() {
		if err := c.Stop(); err != nil {
			return fmt.Errorf("verwaisten Postmaster stoppen: %w", err)
		}
	}
	opts := "-p " + strconv.Itoa(c.Port) + " -c listen_addresses=127.0.0.1"
	return c.run("pg_ctl", "start", "-D", c.Data, "-w", "-t", "120", "-l", c.logFile(), "-o", opts)
}

// Stop fährt den Postmaster mit -m fast herunter und wartet auf das Ende.
func (c Cluster) Stop() error {
	return c.run("pg_ctl", "stop", "-D", c.Data, "-m", "fast", "-w", "-t", "120")
}

// Running meldet, ob pg_ctl einen laufenden Postmaster für Data findet.
func (c Cluster) Running() bool {
	cmd := exec.Command(c.exe("pg_ctl"), "status", "-D", c.Data) //nolint:gosec // G204: Pfade der eigenen Ablage
	return cmd.Run() == nil
}

func (c Cluster) logFile() string {
	return filepath.Join(c.Data, "postgres.log")
}

func (c Cluster) exe(name string) string {
	return filepath.Join(c.Bin, name+exeSuffix)
}

func (c Cluster) run(name string, args ...string) error {
	cmd := exec.Command(c.exe(name), args...) //nolint:gosec // G204: Pfade der eigenen Ablage
	out, err := cmd.CombinedOutput()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("%s: %w\n%s", name, err, out)
		}
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}
