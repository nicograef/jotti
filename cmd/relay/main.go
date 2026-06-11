package main

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// DruckAuftrag ist das DTO vom jotti-Backend.
type DruckAuftrag struct {
	ID      int    `json:"id"`
	ZielIP  string `json:"zielIp"`
	Payload string `json:"payload"`
}

type RelayConfig struct {
	BackendURL    string
	Token         string
	PollSeconds   int
	TLSSkipVerify bool
}

const maxRetries = 60
const defaultBackendURL = "https://localhost/api"
const defaultPollSeconds = 2

func loadConfigFromEnv(getenv func(string) string) (RelayConfig, error) {
	token := strings.TrimSpace(getenv("RELAY_AUTH_TOKEN"))
	if token == "" {
		return RelayConfig{}, fmt.Errorf("RELAY_AUTH_TOKEN ist erforderlich")
	}

	backendURL := strings.TrimRight(strings.TrimSpace(getenv("RELAY_BACKEND_URL")), "/")
	if backendURL == "" {
		backendURL = defaultBackendURL
	}

	pollSeconds := defaultPollSeconds
	pollSecondsRaw := strings.TrimSpace(getenv("RELAY_POLL_SECONDS"))
	if pollSecondsRaw != "" {
		parsedPollSeconds, err := strconv.Atoi(pollSecondsRaw)
		if err != nil || parsedPollSeconds < 1 {
			return RelayConfig{}, fmt.Errorf("RELAY_POLL_SECONDS muss eine positive Ganzzahl sein")
		}
		pollSeconds = parsedPollSeconds
	}

	tlsSkipRaw := strings.TrimSpace(getenv("RELAY_TLS_SKIP_VERIFY"))
	tlsSkipVerify, err := parseTLSSkipVerify(tlsSkipRaw)
	if err != nil {
		return RelayConfig{}, err
	}

	// Local default (single-device setup behind self-signed TLS on localhost).
	if tlsSkipRaw == "" && backendURL == defaultBackendURL {
		tlsSkipVerify = true
	}

	return RelayConfig{
		BackendURL:    backendURL,
		Token:         token,
		PollSeconds:   pollSeconds,
		TLSSkipVerify: tlsSkipVerify,
	}, nil
}

func parseTLSSkipVerify(raw string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "":
		return false, nil
	case "1", "true":
		return true, nil
	case "0", "false":
		return false, nil
	default:
		return false, fmt.Errorf("RELAY_TLS_SKIP_VERIFY muss 1/true oder 0/false sein")
	}
}

func main() {
	config, err := loadConfigFromEnv(os.Getenv)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("jotti Print-Relay gestartet | Backend: %s | Poll: %ds", config.BackendURL, config.PollSeconds)

	client := &http.Client{Timeout: 10 * time.Second}
	if config.TLSSkipVerify {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		log.Printf("TLS-Zertifikatsprüfung deaktiviert (selbstsigniert)")
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	lastStatusLog := time.Now()

	for {
		select {
		case <-quit:
			log.Printf("Shutdown-Signal empfangen. Beende.")
			return
		default:
		}

		auftraege, err := poll(client, config)
		if err != nil {
			log.Printf("Fehler beim Poll: %v", err)
		} else if len(auftraege) == 0 {
			if time.Since(lastStatusLog) > 5*time.Minute {
				log.Printf("Relay aktiv, keine offenen Auftraege")
				lastStatusLog = time.Now()
			}
		} else {
			gedruckteIDs := make([]int, 0, len(auftraege))
			for _, a := range auftraege {
				if err := printAuftragWithRetry(a); err != nil {
					log.Printf("Druckfehler nach max. Versuchen (Auftrag %d): %v -- bleibt offen", a.ID, err)
				} else {
					log.Printf("Auftrag %d erfolgreich gedruckt auf %s", a.ID, a.ZielIP)
					gedruckteIDs = append(gedruckteIDs, a.ID)
				}
			}

			if len(gedruckteIDs) > 0 {
				if err := quittieren(client, gedruckteIDs, config); err != nil {
					log.Printf("Quittieren fehlgeschlagen (%d Auftraege): %v", len(gedruckteIDs), err)
				} else {
					log.Printf("%d Auftraege quittiert", len(gedruckteIDs))
				}
			}
			lastStatusLog = time.Now()
		}

		time.Sleep(time.Duration(config.PollSeconds) * time.Second)
	}
}

func printAuftragWithRetry(a DruckAuftrag) error {
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if err := checkPrinter(a.ZielIP); err != nil {
			log.Printf("Drucker %s nicht bereit (Versuch %d/%d): %v -- warte 5s",
				a.ZielIP, attempt, maxRetries, err)
			time.Sleep(5 * time.Second)
			continue
		}

		escposData, err := base64.StdEncoding.DecodeString(a.Payload)
		if err != nil {
			return fmt.Errorf("ungueltiges Base64: %w", err)
		}
		if err := sendToPrinter(a.ZielIP, escposData); err != nil {
			log.Printf("Sendefehler (Versuch %d/%d): %v", attempt, maxRetries, err)
			time.Sleep(5 * time.Second)
			continue
		}
		return nil
	}
	return fmt.Errorf("max. Versuche (%d) erreicht fuer Drucker %s", maxRetries, a.ZielIP)
}

func poll(client *http.Client, config RelayConfig) ([]DruckAuftrag, error) {
	reqBody, _ := json.Marshal(map[string]any{
		"token": config.Token,
	})
	resp, err := client.Post(config.BackendURL+"/relay/poll", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("ungueltiger Token (401) -- Relay-Token pruefen")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unerwarteter HTTP-Status: %d", resp.StatusCode)
	}

	var result struct {
		Auftraege []DruckAuftrag `json:"auftraege"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("JSON-Decode fehlgeschlagen: %w", err)
	}
	return result.Auftraege, nil
}

func quittieren(client *http.Client, ids []int, config RelayConfig) error {
	reqBody, _ := json.Marshal(map[string]any{
		"token":        config.Token,
		"gedruckteIds": ids,
	})
	resp, err := client.Post(config.BackendURL+"/relay/quittieren", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("ungueltiger Token (401) -- Relay-Token pruefen")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unerwarteter HTTP-Status: %d", resp.StatusCode)
	}

	return nil
}

func checkPrinter(ip string) error {
	conn, err := net.DialTimeout("tcp", ip+":9100", 3*time.Second)
	if err != nil {
		return fmt.Errorf("nicht erreichbar: %w", err)
	}
	defer conn.Close()

	if _, err := conn.Write([]byte{0x10, 0x04, 0x04}); err != nil {
		return fmt.Errorf("status-abfrage fehlgeschlagen: %w", err)
	}

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	reply := make([]byte, 1)
	if _, err := conn.Read(reply); err != nil {
		return nil
	}

	if reply[0]&0x40 != 0 {
		return fmt.Errorf("papier leer (status=0x%02X)", reply[0])
	}
	if reply[0]&0x20 != 0 {
		log.Printf("WARNUNG: Drucker %s meldet Papier fast leer", ip)
	}
	return nil
}

func sendToPrinter(ip string, data []byte) error {
	conn, err := net.DialTimeout("tcp", ip+":9100", 3*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
	if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
		return fmt.Errorf("set write deadline: %w", err)
	}
	_, err = conn.Write(data)
	return err
}
