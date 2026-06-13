// Command jotti-local-proxy ist der Entrypoint des lokalen Caddy-Containers
// (docker-compose.local.yml). Ablauf beim Start: Installations-State
// sicherstellen (Install-ID + acme-dns-Credentials, einmalige Registrierung) →
// LAN-IP bestimmen → Caddyfile rendern → Caddy als Kindprozess starten. Caddy
// holt und erneuert das vertrauenswürdige Wildcard-Zertifikat asynchron; bis
// dahin (und offline) trägt die Fallback-Site mit interner CA. prod/rocks bleiben
// auf nginx und nutzen dieses Programm nicht.
package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const (
	defaultZone          = "lokal.jotti.rocks"
	defaultACMEDNSURL    = "https://auth.jotti.rocks"
	defaultStatePath     = "/state/install.json"
	defaultCaddyfilePath = "/etc/caddy/Caddyfile"
	defaultCaddyBin      = "caddy"
)

// registerTimeout begrenzt den einmaligen Registrierungs-Request bei acme-dns.
const registerTimeout = 30 * time.Second

type config struct {
	lanIPEnv      string
	zone          string
	acmeDNSURL    string
	statePath     string
	caddyfilePath string
	caddyBin      string
	leStaging     bool
}

func loadConfig(getenv func(string) string) config {
	return config{
		lanIPEnv:      getenv("LAN_IP"),
		zone:          valueOrDefault(getenv("PROXY_ZONE"), defaultZone),
		acmeDNSURL:    valueOrDefault(getenv("ACMEDNS_BASE_URL"), defaultACMEDNSURL),
		statePath:     valueOrDefault(getenv("PROXY_STATE_PATH"), defaultStatePath),
		caddyfilePath: valueOrDefault(getenv("PROXY_CADDYFILE_PATH"), defaultCaddyfilePath),
		caddyBin:      valueOrDefault(getenv("PROXY_CADDY_BIN"), defaultCaddyBin),
		leStaging:     parseBool(getenv("PROXY_LE_STAGING")),
	}
}

func main() {
	cfg := loadConfig(os.Getenv)

	state, err := ensureState(stateDeps{
		path:      cfg.statePath,
		readFile:  os.ReadFile,
		writeFile: os.WriteFile,
		register: func() (InstallState, error) {
			return registerWithACMEDNS(&http.Client{Timeout: registerTimeout}, cfg.acmeDNSURL)
		},
	})
	hasState := err == nil
	if hasState {
		log.Printf("Installations-State geladen | Install-ID: %s", state.Subdomain)
	} else {
		log.Printf("Kein nutzbarer Installations-State (%v) — Start nur mit der Fallback-Adresse; die Registrierung wird beim nächsten Start erneut versucht", err)
	}

	lanIP, lanOK := resolveLANIP(cfg.lanIPEnv)
	if lanOK {
		if hasState {
			log.Printf("Vertrauenswürdige Adresse: https://%s", deriveHostname(lanIP, state.Subdomain, cfg.zone))
		}
		log.Printf("Fallback-Adresse: https://%s", lanIP)
	} else {
		log.Printf("LAN-IP unbekannt (LAN_IP nicht gesetzt) — Zugangsadresse erst sichtbar, sobald die IP übergeben wird")
	}

	if cfg.leStaging {
		log.Printf("ACME-CA: Let's-Encrypt-STAGING (Testmodus, kein vertrauenswürdiges Zertifikat)")
	}

	caddyfile := renderCaddyfile(caddyfileInput{
		state:      state,
		hasState:   hasState,
		zone:       cfg.zone,
		acmeDNSURL: cfg.acmeDNSURL,
		leStaging:  cfg.leStaging,
	})
	if err := os.WriteFile(cfg.caddyfilePath, []byte(caddyfile), 0o644); err != nil {
		log.Fatalf("Caddyfile schreiben: %v", err)
	}

	// Status-Seite parallel zu Caddy bereitstellen (im Compose nur an 127.0.0.1
	// gemappt). Sie probt laufend Zertifikat und Rebind und wechselt von der
	// Fallback- auf die grüne Adresse, sobald Caddy ausgestellt hat.
	status := newStatusServer(statusConfig{
		zone:      cfg.zone,
		state:     state,
		hasState:  hasState,
		lanIP:     lanIP,
		lanOK:     lanOK,
		leStaging: cfg.leStaging,
	})
	go func() {
		if err := status.listenAndServe(); err != nil {
			log.Printf("Status-Seite beendet: %v", err)
		}
	}()
	log.Printf("Status & Zugangsadresse: http://localhost:8484")

	if err := runCaddy(cfg.caddyBin, cfg.caddyfilePath); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		log.Fatalf("Caddy konnte nicht gestartet werden: %v", err)
	}
}

// runCaddy startet Caddy als Kindprozess, reicht Terminationssignale durch und
// blockiert bis Caddy endet — Caddy ist der lang laufende Vordergrundprozess des
// Containers. Der Exit-Status wird vom Aufrufer gespiegelt.
func runCaddy(bin, configPath string) error {
	cmd := exec.Command(bin, "run", "--config", configPath, "--adapter", "caddyfile")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return err
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		s := <-signals
		log.Printf("Signal %s empfangen, leite an Caddy weiter", s)
		_ = cmd.Process.Signal(s)
	}()

	return cmd.Wait()
}

func valueOrDefault(raw, fallback string) string {
	if v := strings.TrimSpace(raw); v != "" {
		return v
	}
	return fallback
}

func parseBool(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes":
		return true
	default:
		return false
	}
}
