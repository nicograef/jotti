package core

import (
	"errors"
	"path/filepath"
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
	res := ResolveEnv("POSTGRES_PASSWORD=ausvolume\n", []string{"POSTGRES_PASSWORD=lokal\n"}, true)
	if res.Seed {
		t.Fatal("Seed: got true, want false (Volume-Secret darf nicht neu geschrieben werden)")
	}
	if res.Abort {
		t.Fatal("Abort: got true, want false (Volume hat ein Secret)")
	}
	if res.Content != "POSTGRES_PASSWORD=ausvolume\n" {
		t.Fatalf("Volume-Inhalt muss unveraendert uebernommen werden: got %q", res.Content)
	}
}

func TestResolveEnvAdoptsFirstNonEmptyCandidate(t *testing.T) {
	// Volume leer, erster Kandidat (Host-Spiegel) leer → der zweite (ordnerlokal
	// neben der Exe) gewinnt und muss ins Volume.
	res := ResolveEnv("   \n", []string{"  \n", "POSTGRES_PASSWORD=lokal\n"}, true)
	if !res.Seed {
		t.Fatal("Seed: got false, want true (adoptierter Inhalt muss ins Volume)")
	}
	if res.Abort {
		t.Fatal("Abort: got true, want false (ein Kandidat traegt ein Secret)")
	}
	if res.Content != "POSTGRES_PASSWORD=lokal\n" {
		t.Fatalf("erster nicht-leerer Kandidat muss adoptiert werden: got %q", res.Content)
	}
}

func TestResolveEnvAbortsWhenDataButNoSecret(t *testing.T) {
	// Daten vorhanden, aber nirgends ein Secret → abbrechen, nicht neu erzeugen.
	res := ResolveEnv("", []string{"", "   "}, true)
	if !res.Abort {
		t.Fatal("Abort: got false, want true (Fail-Safe: Daten ohne Secret)")
	}
	if res.Seed || res.Content != "" {
		t.Fatalf("Abort darf keinen Inhalt erzeugen: got Content=%q Seed=%v", res.Content, res.Seed)
	}
}

func TestResolveEnvGeneratesFreshOnFirstInstall(t *testing.T) {
	// Keine Daten, kein Secret → echte Erstinstallation, frische Secrets ins Volume.
	res := ResolveEnv("", nil, false)
	if !res.Seed {
		t.Fatal("Seed: got false, want true (frische Secrets muessen ins Volume)")
	}
	if res.Abort {
		t.Fatal("Abort: got true, want false (ohne Daten ist es eine Erstinstallation)")
	}
	for _, key := range []string{"POSTGRES_USER=admin", "POSTGRES_PASSWORD=", "JWT_SECRET=", "RELAY_AUTH_TOKEN="} {
		if !strings.Contains(res.Content, key) {
			t.Fatalf("frischer Inhalt enthaelt %q nicht:\n%s", key, res.Content)
		}
	}
}

func TestStateDirWindowsUsesProgramData(t *testing.T) {
	got := StateDir("windows", `C:\ProgramData`, `D:\jotti-entpackt`)
	want := filepath.Join(`C:\ProgramData`, "jotti")
	if got != want {
		t.Fatalf("Windows-Zustandsverzeichnis: got %q, want %q", got, want)
	}
}

func TestStateDirFallsBackWhenNoProgramData(t *testing.T) {
	// Linux-Dev (kein PROGRAMDATA) sowie Windows ohne gesetztes PROGRAMDATA
	// bleiben ordnerlokal.
	for _, goos := range []string{"linux", "windows"} {
		if got := StateDir(goos, "", "/opt/jotti"); got != "/opt/jotti" {
			t.Fatalf("%s ohne PROGRAMDATA: got %q, want /opt/jotti", goos, got)
		}
	}
}
