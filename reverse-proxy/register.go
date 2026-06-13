package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// registerWithACMEDNS registriert eine neue Installation bei acme-dns
// (POST <baseURL>/register) und liefert die ausgegebenen Credentials. acme-dns
// vergibt dabei eine zufällige UUID-Subdomain — sie wird zur Install-ID. Der
// Aufruf ist offen (kein Auth); die Antwort wird einmalig persistiert.
func registerWithACMEDNS(client *http.Client, baseURL string) (InstallState, error) {
	url := strings.TrimRight(strings.TrimSpace(baseURL), "/") + "/register"
	resp, err := client.Post(url, "application/json", nil)
	if err != nil {
		return InstallState{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return InstallState{}, fmt.Errorf("unerwarteter HTTP-Status bei /register: %d", resp.StatusCode)
	}

	var state InstallState
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return InstallState{}, fmt.Errorf("JSON-Decode der Registrierung: %w", err)
	}
	if !state.valid() {
		return InstallState{}, errors.New("acme-dns-Antwort ohne vollständige Credentials")
	}
	return state, nil
}
