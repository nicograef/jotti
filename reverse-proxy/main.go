// Command jotti-reverse-proxy ist der Caddy-Container-Entrypoint für drei Modi, die
// die Umgebung wählt: PROXY_HTTP_ONLY (E2E, Klartext-HTTP auf :80) vor JOTTI_DOMAIN
// (Public, eine Site mit Let's-Encrypt-Zertifikat), sonst LAN-Mode (Install-State,
// LAN-IP, Wildcard- plus Fallback-Site, Status-Seite).
// Die jotti.rocks-Demo bleibt auf nginx und nutzt dieses Programm nicht.
package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
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
	domain        string // JOTTI_DOMAIN gesetzt ⇒ Public-Mode statt LAN-Mode
	httpOnly      bool   // PROXY_HTTP_ONLY gesetzt ⇒ Klartext-HTTP-Mode (nur E2E)
	email         string // LETSENCRYPT_EMAIL (Public-Mode)
	wwwRedirect   bool   // JOTTI_WWW_REDIRECT (Public-Mode)
	lanIPEnv      string
	zone          string
	acmeDNSURL    string
	statePath     string
	caddyfilePath string
	caddyBin      string
	leStaging     bool
}

// loadConfig legt den Modus fest. dirExists prüft, ob das State-Verzeichnis
// gemountet ist (nur die LAN-Stacks mounten proxy-state:/state); fehlt es und ist
// JOTTI_DOMAIN leer, bleibt kein gültiger Modus übrig — sonst fiele der
// Public-Stack still in den LAN-Modus und stellte für jede SNI ein Zertifikat der
// internen CA aus, während der Healthcheck grün bleibt.
func loadConfig(getenv func(string) string, dirExists func(string) bool) (config, error) {
	cfg := config{
		domain:        strings.TrimSpace(getenv("JOTTI_DOMAIN")),
		httpOnly:      parseBool(getenv("PROXY_HTTP_ONLY")),
		email:         strings.TrimSpace(getenv("LETSENCRYPT_EMAIL")),
		wwwRedirect:   parseBool(getenv("JOTTI_WWW_REDIRECT")),
		lanIPEnv:      strings.TrimSpace(getenv("LAN_IP")),
		zone:          valueOrDefault(getenv("PROXY_ZONE"), defaultZone),
		acmeDNSURL:    valueOrDefault(getenv("ACMEDNS_BASE_URL"), defaultACMEDNSURL),
		statePath:     valueOrDefault(getenv("PROXY_STATE_PATH"), defaultStatePath),
		caddyfilePath: valueOrDefault(getenv("PROXY_CADDYFILE_PATH"), defaultCaddyfilePath),
		caddyBin:      valueOrDefault(getenv("PROXY_CADDY_BIN"), defaultCaddyBin),
		leStaging:     parseBool(getenv("PROXY_LE_STAGING")),
	}

	if !cfg.httpOnly && cfg.domain == "" {
		if stateDir := filepath.Dir(cfg.statePath); !dirExists(stateDir) {
			return config{}, fmt.Errorf("kein Modus bestimmbar: JOTTI_DOMAIN ist leer und das State-Verzeichnis %s des LAN-Modus fehlt", stateDir)
		}
	}

	return cfg, nil
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func main() {
	cfg, err := loadConfig(os.Getenv, dirExists)
	if err != nil {
		log.Fatalf("Konfiguration: %v", err)
	}

	if cfg.httpOnly {
		runHTTPOnlyMode(cfg)
		return
	}

	if cfg.domain != "" {
		runPublicMode(cfg)
		return
	}

	runLANMode(cfg)
}

// runPublicMode rendert und startet Caddy für den öffentlichen Self-Hoster-Stack:
// eine Site für JOTTI_DOMAIN mit automatischem Let's-Encrypt-Zertifikat.
func runPublicMode(cfg config) {
	if cfg.email == "" {
		log.Fatalf("Public-Mode (JOTTI_DOMAIN=%s): LETSENCRYPT_EMAIL muss gesetzt sein", cfg.domain)
	}

	log.Printf("Public-Mode aktiv | Zugangsadresse: https://%s", cfg.domain)
	if cfg.wwwRedirect {
		log.Printf("www.%s leitet dauerhaft auf %s um", cfg.domain, cfg.domain)
	}
	if cfg.leStaging {
		log.Printf("ACME-CA: Let's-Encrypt-STAGING (Testmodus, kein vertrauenswürdiges Zertifikat)")
	}

	caddyfile := renderPublicCaddyfile(publicInput{
		domain:      cfg.domain,
		email:       cfg.email,
		wwwRedirect: cfg.wwwRedirect,
		leStaging:   cfg.leStaging,
	})
	writeCaddyfileOrExit(cfg.caddyfilePath, caddyfile)

	runCaddyOrExit(cfg)
}

// runHTTPOnlyMode rendert und startet Caddy für die E2E-Testumgebung: Klartext-HTTP
// auf :80, ohne TLS, ACME oder Status-Seite.
func runHTTPOnlyMode(cfg config) {
	log.Printf("HTTP-Only-Mode aktiv (nur E2E) | Zugangsadresse: http://<host>")

	caddyfile := renderHTTPOnlyCaddyfile()
	writeCaddyfileOrExit(cfg.caddyfilePath, caddyfile)

	runCaddyOrExit(cfg)
}

func runLANMode(cfg config) {
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
		log.Printf("Kein nutzbarer Installations-State (%v) — Start nur mit der Fallback-Adresse; die grüne Adresse entsteht erst bei einem Neustart mit gültigem State", err)
	}

	lanIP, lanOK := resolveLANIP(cfg.lanIPEnv)
	if lanOK {
		if hasState {
			log.Printf("Vertrauenswürdige Adresse: https://%s", deriveHostname(lanIP, state.Subdomain, cfg.zone))
		}
		log.Printf("Fallback-Adresse: https://%s", lanIP)
	} else {
		log.Printf("LAN-IP unbekannt (LAN_IP nicht gesetzt) — eine Zugangsadresse entsteht erst bei einem Neustart mit gesetztem LAN_IP")
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
	writeCaddyfileOrExit(cfg.caddyfilePath, caddyfile)

	// Status-Seite parallel zu Caddy (im Compose nur an 127.0.0.1 gemappt).
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

	runCaddyOrExit(cfg)
}

// caddyfileMode: nur für den Eigentümer lesbar — die LAN-Caddyfile trägt die
// acme-dns-Zugangsdaten im Klartext.
const caddyfileMode = 0o600

// writeCaddyfileOrExit bricht bei einem Fehler ab — ohne Konfiguration hat der
// Start keinen Sinn.
func writeCaddyfileOrExit(path, caddyfile string) {
	if err := writeCaddyfile(path, caddyfile); err != nil {
		log.Fatalf("Caddyfile schreiben: %v", err)
	}
}

// writeCaddyfile chmodded explizit: os.WriteFile setzt den Modus nur beim Anlegen,
// eine bereits vorhandene Datei behielte ihren.
func writeCaddyfile(path, caddyfile string) error {
	if err := os.WriteFile(path, []byte(caddyfile), caddyfileMode); err != nil {
		return err
	}
	return os.Chmod(path, caddyfileMode)
}

func runCaddyOrExit(cfg config) {
	if err := runCaddy(cfg.caddyBin, cfg.caddyfilePath); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		log.Fatalf("Caddy konnte nicht gestartet werden: %v", err)
	}
}

// runCaddy startet Caddy als Kindprozess und reicht Terminationssignale durch;
// Caddy ist der lang laufende Vordergrundprozess des Containers.
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
