package core

import (
	"errors"
	"strings"
	"testing"
)

func TestMaterializeEnvCreatesWhenMissing(t *testing.T) {
	var gotPath string
	var gotContent []byte
	created, err := MaterializeEnv("/some/.env",
		func(string) (bool, error) { return false, nil },
		func(p string, b []byte) error { gotPath = p; gotContent = b; return nil },
	)
	if err != nil {
		t.Fatalf("unerwarteter Fehler: %v", err)
	}
	if !created {
		t.Fatal("created: got false, want true")
	}
	if gotPath != "/some/.env" {
		t.Fatalf("Pfad: got %q, want /some/.env", gotPath)
	}

	content := string(gotContent)
	if !strings.HasPrefix(content, "#") {
		t.Fatalf("Inhalt beginnt nicht mit Kommentar-Header:\n%s", content)
	}
	// Die vier Keys aus .env.example muessen vorhanden sein.
	for _, key := range []string{"POSTGRES_USER=admin", "POSTGRES_PASSWORD=", "JWT_SECRET=", "RELAY_AUTH_TOKEN="} {
		if !strings.Contains(content, key) {
			t.Fatalf("Inhalt enthaelt %q nicht:\n%s", key, content)
		}
	}
}

func TestMaterializeEnvNeverOverwrites(t *testing.T) {
	written := false
	created, err := MaterializeEnv("/some/.env",
		func(string) (bool, error) { return true, nil },
		func(string, []byte) error { written = true; return nil },
	)
	if err != nil {
		t.Fatalf("unerwarteter Fehler: %v", err)
	}
	if created {
		t.Fatal("created: got true, want false (vorhandene Datei darf nicht ueberschrieben werden)")
	}
	if written {
		t.Fatal("write wurde aufgerufen, obwohl die Datei bereits existiert")
	}
}

func TestMaterializeEnvPropagatesExistsError(t *testing.T) {
	wantErr := errors.New("stat fehlgeschlagen")
	_, err := MaterializeEnv("/some/.env",
		func(string) (bool, error) { return false, wantErr },
		func(string, []byte) error { return nil },
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("exists-Fehler nicht durchgereicht: got %v", err)
	}
}

func TestResolveEnvVolumeWins(t *testing.T) {
	content, seed := ResolveEnv("POSTGRES_PASSWORD=ausvolume\n", "POSTGRES_PASSWORD=lokal\n")
	if seed {
		t.Fatal("seed: got true, want false (Volume-Secret darf nicht neu geschrieben werden)")
	}
	if content != "POSTGRES_PASSWORD=ausvolume\n" {
		t.Fatalf("Volume-Inhalt muss unveraendert uebernommen werden: got %q", content)
	}
}

func TestResolveEnvAdoptsLocalWhenVolumeEmpty(t *testing.T) {
	content, seed := ResolveEnv("   \n", "POSTGRES_PASSWORD=lokal\n")
	if !seed {
		t.Fatal("seed: got false, want true (adoptierter Inhalt muss ins Volume)")
	}
	if content != "POSTGRES_PASSWORD=lokal\n" {
		t.Fatalf("ordnerlokaler Inhalt muss adoptiert werden: got %q", content)
	}
}

func TestResolveEnvGeneratesFreshWhenBothEmpty(t *testing.T) {
	content, seed := ResolveEnv("", "")
	if !seed {
		t.Fatal("seed: got false, want true (frische Secrets muessen ins Volume)")
	}
	for _, key := range []string{"POSTGRES_USER=admin", "POSTGRES_PASSWORD=", "JWT_SECRET=", "RELAY_AUTH_TOKEN="} {
		if !strings.Contains(content, key) {
			t.Fatalf("frischer Inhalt enthaelt %q nicht:\n%s", key, content)
		}
	}
}
