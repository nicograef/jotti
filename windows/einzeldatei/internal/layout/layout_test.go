package layout

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDockerInstallation(t *testing.T) {
	cases := []struct {
		name        string
		lastVersion bool
		pgdata      bool
		want        bool
	}{
		{"leer", false, false, false},
		{"nur last-version: Docker", true, false, true},
		{"nur pgdata: abgebrochener erster Start", false, true, false},
		{"beides: Einzeldatei", true, true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			// Eine .env bleibt auch nach einem abgebrochenen Start zurück und entscheidet nichts.
			if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("X=1\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			if tc.lastVersion {
				if err := os.WriteFile(filepath.Join(dir, "last-version"), []byte("v1.0.0\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if tc.pgdata {
				if err := os.Mkdir(filepath.Join(dir, "pgdata"), 0o700); err != nil {
					t.Fatal(err)
				}
			}
			got, err := DockerInstallation(dir)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Fatalf("DockerInstallation = %v, want %v", got, tc.want)
			}
		})
	}
}
