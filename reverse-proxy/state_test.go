package main

import (
	"encoding/json"
	"io/fs"
	"testing"
)

// fakeFS ist ein In-Memory-Dateisystem für die injizierten Dateizugriffe von
// ensureState.
type fakeFS struct {
	files map[string][]byte
}

func newFakeFS() *fakeFS { return &fakeFS{files: map[string][]byte{}} }

func (f *fakeFS) read(path string) ([]byte, error) {
	data, ok := f.files[path]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return data, nil
}

func (f *fakeFS) write(path string, data []byte, _ fs.FileMode) error {
	f.files[path] = data
	return nil
}

// failWrite liefert eine writeFile-Funktion, die den Test fehlschlagen lässt,
// sobald sie aufgerufen wird — für Fälle, in denen nichts geschrieben werden darf.
func failWrite(t *testing.T) func(string, []byte, fs.FileMode) error {
	t.Helper()
	return func(string, []byte, fs.FileMode) error {
		t.Fatal("writeFile darf nicht aufgerufen werden")
		return nil
	}
}

func TestEnsureStateRegistersOnceAndPersists(t *testing.T) {
	const path = "/state/install.json"
	fsys := newFakeFS()
	calls := 0
	deps := stateDeps{
		path:      path,
		readFile:  fsys.read,
		writeFile: fsys.write,
		register: func() (InstallState, error) {
			calls++
			return InstallState{Username: "u", Password: "p", Subdomain: "sub-id"}, nil
		},
	}

	first, err := ensureState(deps)
	if err != nil {
		t.Fatalf("erster Lauf: unerwarteter Fehler: %v", err)
	}
	if calls != 1 {
		t.Fatalf("Registrierungen nach erstem Lauf: got %d, want 1", calls)
	}
	if first.Subdomain != "sub-id" {
		t.Fatalf("Subdomain: got %q, want %q", first.Subdomain, "sub-id")
	}
	if _, ok := fsys.files[path]; !ok {
		t.Fatalf("State wurde nicht persistiert")
	}

	// Zweiter Lauf: State liegt vor → keine erneute Registrierung, unverändert.
	second, err := ensureState(deps)
	if err != nil {
		t.Fatalf("zweiter Lauf: unerwarteter Fehler: %v", err)
	}
	if calls != 1 {
		t.Fatalf("Registrierung wurde erneut aufgerufen: %d", calls)
	}
	if second != first {
		t.Fatalf("State änderte sich zwischen den Läufen: %+v vs %+v", second, first)
	}
}

func TestEnsureStateNeverOverwritesExisting(t *testing.T) {
	const path = "/state/install.json"
	existing := InstallState{Username: "keep-u", Password: "keep-p", Subdomain: "keep-id"}
	data, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}
	fsys := &fakeFS{files: map[string][]byte{path: data}}

	got, err := ensureState(stateDeps{
		path:      path,
		readFile:  fsys.read,
		writeFile: failWrite(t),
		register: func() (InstallState, error) {
			t.Fatal("register darf bei vorhandenem State nicht aufgerufen werden")
			return InstallState{}, nil
		},
	})
	if err != nil {
		t.Fatalf("unerwarteter Fehler: %v", err)
	}
	if got != existing {
		t.Fatalf("State: got %+v, want %+v", got, existing)
	}
}

func TestEnsureStateRejectsCorruptState(t *testing.T) {
	const path = "/state/install.json"
	fsys := &fakeFS{files: map[string][]byte{path: []byte("{ kaputt")}}

	_, err := ensureState(stateDeps{
		path:      path,
		readFile:  fsys.read,
		writeFile: failWrite(t),
		register: func() (InstallState, error) {
			t.Fatal("register darf bei beschädigtem State nicht aufgerufen werden")
			return InstallState{}, nil
		},
	})
	if err == nil {
		t.Fatal("erwartete Fehler bei beschädigtem State, bekam nil")
	}
}

func TestEnsureStateRejectsIncompleteRegistration(t *testing.T) {
	fsys := newFakeFS()
	_, err := ensureState(stateDeps{
		path:      "/state/install.json",
		readFile:  fsys.read,
		writeFile: failWrite(t),
		register: func() (InstallState, error) {
			return InstallState{Username: "u"}, nil // Passwort/Subdomain fehlen
		},
	})
	if err == nil {
		t.Fatal("erwartete Fehler bei unvollständigen Credentials, bekam nil")
	}
}
