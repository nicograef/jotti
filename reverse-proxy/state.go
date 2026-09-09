package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"regexp"
)

// subdomainPattern begrenzt die von acme-dns vergebene Subdomain auf ein
// einzelnes DNS-Label (acme-dns vergibt eine UUID). Die Subdomain landet
// ungequotet in der Site-Adresse `*.<subdomain>.<zone>` des gerenderten
// Caddyfiles (wildcardSite): ein Leerzeichen oder eine geschweifte Klammer
// darin wäre eine zusätzliche Caddy-Direktive. Ein Site-Adress-Token lässt
// sich nicht quoten, also muss der Wert selbst eng sein.
var subdomainPattern = regexp.MustCompile(`^[a-z0-9-]{1,63}$`)

// InstallState ist der persistente Zustand einer Installation: die bei acme-dns
// registrierten Credentials. Die Install-ID ist die von acme-dns vergebene
// Subdomain — sie taucht zugleich im Hostnamen (`*.<subdomain>.lokal…`) und in
// der Challenge-Delegation auf. Keine personenbezogenen Daten.
type InstallState struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Subdomain string `json:"subdomain"`
}

// validate meldet, warum der State unbrauchbar ist, oder nil. Der Grund wandert
// in die Fehlermeldung: eine abgelehnte Subdomain ist etwas anderes als ein
// fehlendes Passwort, und der Betreiber liest nur die Meldung.
func (s InstallState) validate() error {
	if s.Username == "" || s.Password == "" {
		return errors.New("unvollständige Credentials")
	}
	if !subdomainPattern.MatchString(s.Subdomain) {
		return fmt.Errorf("die Subdomain %q ist kein einzelnes DNS-Label", s.Subdomain)
	}
	return nil
}

// stateDeps bündelt die injizierbaren Abhängigkeiten von ensureState, damit die
// Idempotenz ohne echten Dateizugriff und ohne echtes acme-dns testbar ist.
type stateDeps struct {
	path      string
	readFile  func(string) ([]byte, error)
	writeFile func(string, []byte, fs.FileMode) error
	register  func() (InstallState, error)
}

// ensureState lädt den Installations-State oder registriert — falls noch keiner
// existiert — genau einmal bei acme-dns und persistiert das Ergebnis. Ein
// vorhandener, gültiger State wird nie überschrieben und nie neu registriert
// (Idempotenz über Neustarts). Ein vorhandener, aber beschädigter State ist ein
// Fehler statt eines stillen Überschreibens — sonst gingen gültige Credentials
// (und das daran hängende Zertifikat) verloren.
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

	// Kein State vorhanden: einmalig registrieren und persistieren.
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

// parseState liest und validiert den persistierten State.
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
