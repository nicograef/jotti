package main

import (
	"bufio"
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
	"sort"
	"strconv"
	"strings"
	"sync"
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

const defaultBackendURL = "https://localhost/api"
const defaultPollSeconds = 2

// version wird beim Release per -ldflags "-X main.version=vX.Y.Z" gesetzt.
var version = "dev"

// dialTimeout ist der kurze TCP-Timeout für genau einen Zustellversuch pro
// Auftrag und Zyklus. Ein nicht erreichbarer Drucker verzögert seine eigene
// IP-Gruppe nur um diese Spanne; andere Gruppen laufen parallel weiter.
const dialTimeout = 2 * time.Second

// druckFunc stellt einen einzelnen Auftrag zu und meldet einen Fehler, wenn der
// Versuch scheitert. Injizierbar, damit die Zyklus-Logik ohne echte Drucker
// testbar ist.
type druckFunc func(a DruckAuftrag) error

// meldeFunc meldet das Ergebnis eines Zyklus ans Backend. Injizierbar, damit die
// Zyklus-Logik ohne HTTP-Backend testbar ist.
type meldeFunc func(ergebnis zyklusErgebnis) error

// fehlversuch beschreibt einen gescheiterten Zustellversuch eines Auftrags.
type fehlversuch struct {
	ID     int
	Fehler string
}

// zyklusErgebnis fasst zusammen, was ein Poll-Zyklus zugestellt und welche
// Aufträge dabei gescheitert sind.
type zyklusErgebnis struct {
	gedruckteIDs []int
	fehlversuche []fehlversuch
}

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
	config, err := loadConfigFromEnv(envWithFileFallback(loadEnvFile()))
	if err != nil {
		// Beim Doppelklick schliesst das Konsolenfenster beim Exit sofort — die
		// Konfigurationsmeldung waere unlesbar. Deshalb auf Enter warten.
		fmt.Printf("jotti Print-Relay %s\n\n", version)
		fmt.Printf("Konfigurationsfehler: %v\n", err)
		fmt.Println("Bitte RELAY_AUTH_TOKEN in der .env-Datei neben jotti-relay.exe setzen.")
		waitForEnter()
		os.Exit(1)
	}

	log.Printf("jotti Print-Relay %s gestartet | Backend: %s | Poll: %ds", version, config.BackendURL, config.PollSeconds)

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
			melde := func(ergebnis zyklusErgebnis) error {
				return meldeErgebnis(client, ergebnis, config)
			}
			ergebnis, meldeErr := fuehreZyklusAus(auftraege, druckeAuftrag, melde)

			for _, id := range ergebnis.gedruckteIDs {
				log.Printf("Auftrag %d erfolgreich gedruckt", id)
			}
			for _, f := range ergebnis.fehlversuche {
				log.Printf("Auftrag %d fehlgeschlagen: %s", f.ID, f.Fehler)
			}
			if meldeErr != nil {
				log.Printf("Ergebnis-Meldung fehlgeschlagen: %v", meldeErr)
			} else if len(ergebnis.gedruckteIDs) > 0 || len(ergebnis.fehlversuche) > 0 {
				log.Printf("Zyklus gemeldet: %d gedruckt, %d Fehlversuche",
					len(ergebnis.gedruckteIDs), len(ergebnis.fehlversuche))
			}
			lastStatusLog = time.Now()
		}

		time.Sleep(time.Duration(config.PollSeconds) * time.Second)
	}
}

// waitForEnter haelt das Doppelklick-Fenster offen, bis der Nutzer Enter drueckt
// — sonst verschwindet eine Konfigurationsmeldung sofort beim Exit.
func waitForEnter() {
	fmt.Print("\nEnter druecken zum Schliessen ...")
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
}

// fuehreZyklusAus verarbeitet alle Aufträge eines Polls und meldet das Ergebnis.
// Verarbeitung und Meldung sind über druck/melde injizierbar.
func fuehreZyklusAus(auftraege []DruckAuftrag, druck druckFunc, melde meldeFunc) (zyklusErgebnis, error) {
	ergebnis := verarbeiteZyklus(auftraege, druck)
	if len(ergebnis.gedruckteIDs) == 0 && len(ergebnis.fehlversuche) == 0 {
		return ergebnis, nil
	}
	return ergebnis, melde(ergebnis)
}

// verarbeiteZyklus verarbeitet die Aufträge eines Polls: gruppiert nach Ziel-IP,
// Gruppen laufen parallel (ein toter Drucker blockiert keinen anderen).
// Innerhalb einer IP bleibt die ID-Reihenfolge erhalten und der erste Fehler
// bricht die Gruppe ab. Die Resultate werden nach ID sortiert zurückgegeben.
func verarbeiteZyklus(auftraege []DruckAuftrag, druck druckFunc) zyklusErgebnis {
	var (
		mu       sync.Mutex
		wg       sync.WaitGroup
		ergebnis zyklusErgebnis
	)

	for _, gruppe := range gruppiereNachIP(auftraege) {
		wg.Add(1)
		go func(gruppe []DruckAuftrag) {
			defer wg.Done()
			gedruckte, fehler := verarbeiteGruppe(gruppe, druck)

			mu.Lock()
			defer mu.Unlock()
			ergebnis.gedruckteIDs = append(ergebnis.gedruckteIDs, gedruckte...)
			if fehler != nil {
				ergebnis.fehlversuche = append(ergebnis.fehlversuche, *fehler)
			}
		}(gruppe)
	}
	wg.Wait()

	sort.Ints(ergebnis.gedruckteIDs)
	sort.Slice(ergebnis.fehlversuche, func(i, j int) bool {
		return ergebnis.fehlversuche[i].ID < ergebnis.fehlversuche[j].ID
	})
	return ergebnis
}

// verarbeiteGruppe stellt die Aufträge einer einzelnen Ziel-IP in ID-Reihenfolge
// zu — genau ein Versuch pro Auftrag. Beim ersten Fehler bricht die Gruppe ab:
// Dieser Auftrag wird als Fehlversuch gemeldet, die übrigen Aufträge dieser IP
// bleiben offen und werden im nächsten Zyklus erneut versucht.
func verarbeiteGruppe(gruppe []DruckAuftrag, druck druckFunc) ([]int, *fehlversuch) {
	var gedruckteIDs []int
	for _, a := range gruppe {
		if err := druck(a); err != nil {
			return gedruckteIDs, &fehlversuch{ID: a.ID, Fehler: err.Error()}
		}
		gedruckteIDs = append(gedruckteIDs, a.ID)
	}
	return gedruckteIDs, nil
}

// gruppiereNachIP gruppiert Aufträge nach Ziel-IP. Innerhalb jeder Gruppe bleibt
// die Eingabe-Reihenfolge (älteste ID zuerst) erhalten.
func gruppiereNachIP(auftraege []DruckAuftrag) map[string][]DruckAuftrag {
	gruppen := make(map[string][]DruckAuftrag)
	for _, a := range auftraege {
		gruppen[a.ZielIP] = append(gruppen[a.ZielIP], a)
	}
	return gruppen
}

// druckeAuftrag stellt einen Auftrag mit genau einem Versuch zu: Payload
// dekodieren, Drucker prüfen, Daten senden.
func druckeAuftrag(a DruckAuftrag) error {
	escposData, err := base64.StdEncoding.DecodeString(a.Payload)
	if err != nil {
		return fmt.Errorf("ungueltiges Base64: %w", err)
	}
	if err := checkPrinter(a.ZielIP); err != nil {
		return fmt.Errorf("drucker %s: %w", a.ZielIP, err)
	}
	if err := sendToPrinter(a.ZielIP, escposData); err != nil {
		return fmt.Errorf("drucker %s: senden fehlgeschlagen: %w", a.ZielIP, err)
	}
	return nil
}

func poll(client *http.Client, config RelayConfig) ([]DruckAuftrag, error) {
	reqBody, _ := json.Marshal(map[string]any{
		"token": config.Token,
	})
	resp, err := client.Post(config.BackendURL+"/relay/poll", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

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

// meldeErgebnis meldet Erfolge und Fehlversuche eines Zyklus gesammelt in einem
// Request an das Backend. Das Backend besitzt die Fehlversuchs-Logik (zählt
// hoch und markiert nach drei Versuchen als fehlgeschlagen).
func meldeErgebnis(client *http.Client, ergebnis zyklusErgebnis, config RelayConfig) error {
	fehlversuche := make([]map[string]any, 0, len(ergebnis.fehlversuche))
	for _, f := range ergebnis.fehlversuche {
		fehlversuche = append(fehlversuche, map[string]any{
			"id":     f.ID,
			"fehler": f.Fehler,
		})
	}

	gedruckteIDs := ergebnis.gedruckteIDs
	if gedruckteIDs == nil {
		gedruckteIDs = []int{}
	}

	reqBody, _ := json.Marshal(map[string]any{
		"token":        config.Token,
		"gedruckteIds": gedruckteIDs,
		"fehlversuche": fehlversuche,
	})
	resp, err := client.Post(config.BackendURL+"/relay/ergebnis", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("ungueltiger Token (401) -- Relay-Token pruefen")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unerwarteter HTTP-Status: %d", resp.StatusCode)
	}

	return nil
}

func checkPrinter(ip string) error {
	conn, err := net.DialTimeout("tcp", ip+":9100", dialTimeout)
	if err != nil {
		return fmt.Errorf("nicht erreichbar: %w", err)
	}
	defer func() { _ = conn.Close() }()

	if _, err := conn.Write([]byte{0x10, 0x04, 0x04}); err != nil {
		return fmt.Errorf("status-abfrage fehlgeschlagen: %w", err)
	}

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	reply := make([]byte, 1)
	if _, err := conn.Read(reply); err != nil {
		return nil
	}

	// DLE EOT n=4 liefert den Rollenpapier-Sensorstatus (ESC/POS). Bits 2,3
	// (Maske 0x0C) melden den Near-End-Sensor (Papier fast leer), Bits 5,6
	// (Maske 0x60) den End-Sensor (Papier leer). Die übrigen Bits sind fest.
	// Quelle: Epson ESC/POS DLE EOT, bestätigt via escpos.readthedocs.io.
	if reply[0]&0x60 != 0 {
		return fmt.Errorf("papier leer (status=0x%02X)", reply[0])
	}
	if reply[0]&0x0C != 0 {
		log.Printf("WARNUNG: Drucker %s meldet Papier fast leer", ip)
	}
	return nil
}

func sendToPrinter(ip string, data []byte) error {
	conn, err := net.DialTimeout("tcp", ip+":9100", dialTimeout)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()
	if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
		return fmt.Errorf("set write deadline: %w", err)
	}
	_, err = conn.Write(data)
	return err
}
