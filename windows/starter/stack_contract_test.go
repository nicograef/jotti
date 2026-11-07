package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nicograef/jotti/windows/starter/core"
)

// TestStackIdentitaetPasstZuComposeUndSkripten haelt die Namen, die der Starter und
// die Windows-Skripte ansprechen, an den Compose-Dateien fest: Ein umbenanntes
// Projekt oder Volume liesse den Starter sonst frische Secrets ueber bestehende
// Daten legen.
func TestStackIdentitaetPasstZuComposeUndSkripten(t *testing.T) {
	for _, name := range []string{"docker-compose.local.yml", "docker-compose.release.yml"} {
		t.Run(name, func(t *testing.T) {
			compose := readRepoFile(t, name)
			for _, want := range []string{
				"name: " + composeProject + "\n",
				"container_name: " + postgresContainer + "\n",
				"- " + postgresDataVolumeKey + ":/var/lib/postgresql/data\n",
				"- " + configVolumeKey + ":/jotti-config:ro\n",
				"- jotti-backups:" + backupDir + "\n",
				"\n  " + postgresDataVolumeKey + ":\n",
				"\n  " + configVolumeKey + ":\n",
			} {
				if !strings.Contains(compose, want) {
					t.Errorf("%s enthaelt %q nicht", name, want)
				}
			}
		})
	}

	scripts := map[string][]string{
		"jotti-repair.cmd":  {"docker exec " + postgresContainer + " psql -U " + core.PostgresUser + " ", "ALTER USER " + core.PostgresUser + " "},
		"jotti-restore.cmd": {"docker exec " + postgresContainer + " ", backupDir + "/jotti-*.sql", "psql -U " + core.PostgresUser + " "},
	}
	for name, wants := range scripts {
		t.Run(name, func(t *testing.T) {
			script := readRepoFile(t, filepath.Join("packaging", "windows", name))
			for _, want := range wants {
				if !strings.Contains(script, want) {
					t.Errorf("%s enthaelt %q nicht", name, want)
				}
			}
		})
	}
}

func readRepoFile(t *testing.T, rel string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", rel))
	if err != nil {
		t.Fatalf("%s lesen: %v", rel, err)
	}
	return strings.ReplaceAll(string(data), "\r\n", "\n")
}
