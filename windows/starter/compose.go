package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"
)

// Grosszuegig fuer die Erst-Migrationen auf Altgeraeten.
const healthTimeout = 120 * time.Second

// composeUp faehrt den Stack hoch und reicht die Ausgabe live durch. Der Proxy wird
// danach neu erzeugt, weil das Caddyfile nur im Entrypoint aus LAN_IP gerendert wird.
func composeUp(composePath, envPath, lanIP string) error {
	env := os.Environ()
	if lanIP != "" {
		env = append(env, "LAN_IP="+lanIP)
	}
	if err := runCompose(env, composePath, envPath, "up", "-d", "--build"); err != nil {
		return err
	}
	return runCompose(env, composePath, envPath, "up", "-d", "--no-deps", "--force-recreate", "reverse-proxy")
}

// runCompose benennt die .env-Quelle explizit per --env-file: nach der
// UAC-Elevation ist das Arbeitsverzeichnis System32, nicht das Projektverzeichnis.
func runCompose(env []string, composePath, envPath string, args ...string) error {
	full := append([]string{"compose", "-f", composePath, "--env-file", envPath}, args...)
	cmd := exec.Command("docker", full...)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// waitForHealth pollt https://localhost/api/health (TLS-Verify aus — localhost
// trifft Caddys interne CA), bis HTTP 200 kommt oder healthTimeout ablaeuft. Nur
// 200 gilt als bereit; das Backend liefert 503, solange die DB nicht antwortet.
func waitForHealth() error {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	fmt.Print("Warte darauf, dass jotti bereit ist ")
	deadline := time.Now().Add(healthTimeout)
	for time.Now().Before(deadline) {
		resp, err := client.Get("https://localhost/api/health")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				fmt.Println(" bereit.")
				return nil
			}
		}
		fmt.Print(".")
		time.Sleep(2 * time.Second)
	}
	fmt.Println()
	return fmt.Errorf("jotti wurde nicht innerhalb von %s bereit. Bitte die Logs pruefen "+
		"(docker compose logs) und jotti erneut starten", healthTimeout)
}
