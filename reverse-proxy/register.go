package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// registerWithACMEDNS registriert eine Installation bei acme-dns
// (POST <baseURL>/register). Die vergebene UUID-Subdomain wird zur Install-ID; der
// Aufruf ist offen (kein Auth).
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
	if err := state.validate(); err != nil {
		return InstallState{}, fmt.Errorf("acme-dns-Antwort unbrauchbar: %w", err)
	}
	return state, nil
}
