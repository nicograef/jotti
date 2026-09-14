package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"regexp"
)

// subdomainPattern begrenzt die acme-dns-Subdomain auf ein einzelnes DNS-Label. Sie
// landet ungequotet in der Site-Adresse `*.<subdomain>.<zone>` (wildcardSite) — ein
// Leerzeichen oder eine geschweifte Klammer wäre dort eine zusätzliche
// Caddy-Direktive, und ein Site-Adress-Token lässt sich nicht quoten.
var subdomainPattern = regexp.MustCompile(`^[a-z0-9-]{1,63}$`)

// InstallState ist der persistente Zustand einer Installation: die acme-dns-
// Credentials. Die Subdomain ist zugleich die Install-ID im Hostnamen und in der
// Challenge-Delegation. Keine personenbezogenen Daten.
type InstallState struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Subdomain string `json:"subdomain"`
}

func (s InstallState) validate() error {
	if s.Username == "" || s.Password == "" {
		return errors.New("unvollständige Credentials")
	}
	if !subdomainPattern.MatchString(s.Subdomain) {
		return fmt.Errorf("die Subdomain %q ist kein einzelnes DNS-Label", s.Subdomain)
	}
	return nil
}

// stateDeps bündelt die injizierbaren Abhängigkeiten von ensureState (Tests ohne
// Dateisystem und ohne acme-dns).
type stateDeps struct {
	path      string
	readFile  func(string) ([]byte, error)
	writeFile func(string, []byte, fs.FileMode) error
	register  func() (InstallState, error)
}

// ensureState lädt den Installations-State oder registriert genau einmal bei
// acme-dns. Ein gültiger State wird nie überschrieben (Idempotenz über Neustarts);
// ein beschädigter ist ein Fehler statt eines stillen Überschreibens — sonst gingen
// gültige Credentials und das daran hängende Zertifikat verloren.
func ensureState(deps stateDeps) (InstallState, error) {
	data, err := deps.readFile(deps.path)
	switch {
	case err == nil:
		state, perr := parseState(data)
		if perr != nil {
			return InstallState{}, fmt.Errorf("vorhandener State unter %s ist ungültig und wird nicht überschrieben: %w", deps.path, perr)
		}
		return state, nil
	case !errors.Is(err, fs.ErrNotExist):
		return InstallState{}, fmt.Errorf("state lesen: %w", err)
	}

	state, err := deps.register()
	if err != nil {
		return InstallState{}, fmt.Errorf("acme-dns-Registrierung: %w", err)
	}
	if err := state.validate(); err != nil {
		return InstallState{}, fmt.Errorf("acme-dns lieferte einen unbrauchbaren State: %w", err)
	}

	encoded, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return InstallState{}, fmt.Errorf("state serialisieren: %w", err)
	}
	if err := deps.writeFile(deps.path, encoded, 0o600); err != nil {
		return InstallState{}, fmt.Errorf("state schreiben: %w", err)
	}
	return state, nil
}

func parseState(data []byte) (InstallState, error) {
	var state InstallState
	if err := json.Unmarshal(data, &state); err != nil {
		return InstallState{}, fmt.Errorf("JSON-Decode: %w", err)
	}
	if err := state.validate(); err != nil {
		return InstallState{}, err
	}
	return state, nil
}
