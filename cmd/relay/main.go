package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// RelayState speichert den Cursor und die Idempotenzliste.
type RelayState struct {
	LastEventID     int   `json:"last_event_id"`
	PrintedEventIDs []int `json:"printed_event_ids"`
}

// DruckAuftrag ist das DTO vom jotti-Backend.
type DruckAuftrag struct {
	EventID   int    `json:"eventId"`
	DruckerIP string `json:"druckerIp"`
	Payload   string `json:"payload"`
}

var (
	backendURL  = flag.String("backend", "https://jotti.meinverein.de", "jotti Backend URL")
	token       = flag.String("token", "", "RELAY_AUTH_TOKEN aus .env")
	stateFile   = flag.String("state", "relay_state.json", "Pfad zur lokalen State-Datei")
	pollSeconds = flag.Int("poll", 2, "Poll-Intervall in Sekunden")
)

const maxPrintedIDs = 2000
const maxRetries = 60

func main() {
	flag.Parse()
	if *token == "" {
		log.Fatal("--token ist erforderlich")
	}

	log.Printf("jotti Print-Relay gestartet | Backend: %s | Poll: %ds", *backendURL, *pollSeconds)

	state := loadState(*stateFile)
	client := &http.Client{Timeout: 10 * time.Second}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	idempotencySet := buildIdempotencySet(state.PrintedEventIDs)
	lastStatusLog := time.Now()

	for {
		select {
		case <-quit:
			log.Printf("Shutdown-Signal empfangen. Cursor bei %d. Beende.", state.LastEventID)
			saveState(*stateFile, state)
			return
		default:
		}

		auftraege, err := poll(client, state.LastEventID)
		if err != nil {
			log.Printf("Fehler beim Poll: %v", err)
		} else if len(auftraege) == 0 {
			if time.Since(lastStatusLog) > 5*time.Minute {
				log.Printf("Relay aktiv, Cursor bei %d, keine neuen Auftraege", state.LastEventID)
				lastStatusLog = time.Now()
			}
		} else {
			for _, a := range auftraege {
				if idempotencySet[a.EventID] {
					log.Printf("Event %d bereits gedruckt (Idempotenz) -- ueberspringe", a.EventID)
					state.LastEventID = a.EventID
					continue
				}

				if err := printAuftragWithRetry(a); err != nil {
					log.Printf("Druckfehler nach max. Versuchen (Event %d): %v -- ueberspringe", a.EventID, err)
				} else {
					log.Printf("Event %d erfolgreich gedruckt auf %s", a.EventID, a.DruckerIP)
				}

				state.LastEventID = a.EventID
				state.PrintedEventIDs = appendWithLimit(state.PrintedEventIDs, a.EventID, maxPrintedIDs)
				idempotencySet[a.EventID] = true
				saveState(*stateFile, state)
			}
			lastStatusLog = time.Now()
		}

		time.Sleep(time.Duration(*pollSeconds) * time.Second)
	}
}

func printAuftragWithRetry(a DruckAuftrag) error {
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if err := checkPrinter(a.DruckerIP); err != nil {
			log.Printf("Drucker %s nicht bereit (Versuch %d/%d): %v -- warte 5s",
				a.DruckerIP, attempt, maxRetries, err)
			time.Sleep(5 * time.Second)
			continue
		}

		escposData, err := base64.StdEncoding.DecodeString(a.Payload)
		if err != nil {
			return fmt.Errorf("ungueltiges Base64: %w", err)
		}
		if err := sendToPrinter(a.DruckerIP, escposData); err != nil {
			log.Printf("Sendefehler (Versuch %d/%d): %v", attempt, maxRetries, err)
			time.Sleep(5 * time.Second)
			continue
		}
		return nil
	}
	return fmt.Errorf("max. Versuche (%d) erreicht fuer Drucker %s", maxRetries, a.DruckerIP)
}

func poll(client *http.Client, lastEventID int) ([]DruckAuftrag, error) {
	reqBody, _ := json.Marshal(map[string]any{
		"token":       *token,
		"lastEventId": lastEventID,
	})
	resp, err := client.Post(*backendURL+"/relay/poll", "application/json", bytes.NewReader(reqBody))
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
		Cursor    int            `json:"cursor"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("JSON-Decode fehlgeschlagen: %w", err)
	}
	return result.Auftraege, nil
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

func loadState(path string) *RelayState {
	data, err := os.ReadFile(path)
	if err != nil {
		return &RelayState{}
	}
	var s RelayState
	if err := json.Unmarshal(data, &s); err != nil {
		log.Printf("WARNUNG: State-Datei korrupt, starte neu: %v", err)
		return &RelayState{}
	}
	return &s
}

func saveState(path string, s *RelayState) {
	data, _ := json.Marshal(s)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		log.Printf("WARNUNG: Relay-State konnte nicht geschrieben werden: %v", err)
		return
	}
	if err := os.Rename(tmp, path); err != nil {
		log.Printf("WARNUNG: Relay-State-Rename fehlgeschlagen: %v", err)
	}
}

func buildIdempotencySet(ids []int) map[int]bool {
	set := make(map[int]bool, len(ids))
	for _, id := range ids {
		set[id] = true
	}
	return set
}

func appendWithLimit(ids []int, id int, limit int) []int {
	ids = append(ids, id)
	if len(ids) > limit {
		ids = ids[len(ids)-limit:]
	}
	return ids
}
