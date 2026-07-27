package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
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

// dialTimeout ist der kurze TCP-Timeout für den einen Verbindungsaufbau pro
// Ziel-IP und Zyklus. Ein nicht erreichbarer Drucker verzögert seine eigene
// IP-Gruppe nur um diese Spanne; andere Gruppen laufen parallel weiter.
const dialTimeout = 2 * time.Second

// writeTimeout begrenzt das Senden der Druckdaten an den Drucker.
const writeTimeout = 10 * time.Second

// maxBonsProZyklus begrenzt, wie viele Bons eine Ziel-IP je Zyklus erhält
// (Begründung bei bildeZustellgruppen).
const maxBonsProZyklus = 6

// zustellTimeouts bündelt die Lese-Timeouts einer Gruppen-Zustellung. Sie sind
// injizierbar, damit Tests nicht auf echte Drucker- und Netzzeiten warten.
type zustellTimeouts struct {
	papier         time.Duration
	spuelen        time.Duration
	quittungBasis  time.Duration
	quittungProBon time.Duration
}

// produktionsTimeouts sind die Werte für den echten Betrieb:
//   - papier: Warten auf die Papierstatus-Antwort (Echtzeit-Kommando DLE EOT).
//   - spuelen: Fenster, in dem vor dem Quittungsumlauf noch eintreffende Reste
//     der Papierstatus-Runde verworfen werden. Kurz, weil seit der Abfrage
//     bereits das Papier-Timeout und das Schreiben aller Bons vergangen sind.
//   - quittungBasis/quittungProBon: Lese-Timeout der Zustellquittung. GS r ist
//     ein gepuffertes Kommando: der Drucker antwortet erst, nachdem er die
//     vorher empfangenen Bons gedruckt und geschnitten hat — erfahrungsgemäß
//     1–2 s je Bon. Der Zuschlag von 2 s je Bon deckt das ab, die Basis von 3 s
//     die Verarbeitung des Kommandos und die Netz-Latenz. Weil maxBonsProZyklus
//     eine Gruppe auf sechs Bons begrenzt, wartet sie höchstens 15 s auf die
//     Quittung.
var produktionsTimeouts = zustellTimeouts{
	papier:         2 * time.Second,
	spuelen:        100 * time.Millisecond,
	quittungBasis:  3 * time.Second,
	quittungProBon: 2 * time.Second,
}

// Ausgang der Zustellquittung einer Gruppe — nur für die Logzeile je Gruppe.
const (
	ausgangBestaetigt    = "bestaetigt"
	ausgangUnbeantwortet = "unbeantwortet"
	ausgangAbgebrochen   = "abgebrochen"
)

// dlePapierstatus fragt den Rollenpapier-Sensor ab (DLE EOT n=4). Echtzeit-
// Kommando: der Drucker antwortet sofort, unabhängig von seinem Druckpuffer.
var dlePapierstatus = []byte{0x10, 0x04, 0x04}

// gsStatusabfrage ist GS r 1 = Statusabfrage (Papiersensor); gepuffert. Der
// Drucker führt das Kommando erst aus, wenn er die davor empfangenen Druckdaten
// verarbeitet hat. Eine Antwort beweist deshalb, dass alle Bons der Gruppe
// konsumiert wurden — anders als das Echtzeit-Kommando DLE EOT.
var gsStatusabfrage = []byte{0x1D, 0x72, 0x01}

// verbindeFunc öffnet eine Verbindung zum Drucker mit der gegebenen Ziel-IP.
// Injizierbar, damit die Zustellung in Tests gegen einen lokalen TCP-Listener
// laufen kann.
type verbindeFunc func(zielIP string) (net.Conn, error)

// gruppenDruckFunc stellt alle Aufträge einer Ziel-IP zu und meldet die bestätigt
// zugestellten Auftrags-IDs sowie die Fehlversuche. Injizierbar, damit die
// Zyklus-Logik ohne echte Drucker testbar ist.
type gruppenDruckFunc func(zielIP string, auftraege []DruckAuftrag) ([]int, []fehlversuch)

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

	// Lokaler Standard (Einzelgerät hinter selbstsigniertem TLS auf localhost).
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
		fmt.Println(envHinweis())
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

	zustelle := zustelleGruppe(verbindeMitDrucker, produktionsTimeouts)
	lastStatusLog := time.Now()

	for {
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
			ergebnis, meldeErr := fuehreZyklusAus(auftraege, zustelle, melde)

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

		// Wartezeit bis zum nächsten Poll, unterbrechbar durch das Shutdown-Signal.
		select {
		case <-quit:
			log.Printf("Shutdown-Signal empfangen. Beende.")
			return
		case <-time.After(time.Duration(config.PollSeconds) * time.Second):
		}
	}
}

// waitForEnter haelt das Doppelklick-Fenster offen, bis der Nutzer Enter drueckt
// — sonst verschwindet eine Konfigurationsmeldung sofort beim Exit.
func waitForEnter() {
	fmt.Print("\nEnter druecken zum Schliessen ...")
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
}

// fuehreZyklusAus verarbeitet alle Aufträge eines Polls und meldet das Ergebnis.
// Zustellung und Meldung sind über zustelle/melde injizierbar.
func fuehreZyklusAus(auftraege []DruckAuftrag, zustelle gruppenDruckFunc, melde meldeFunc) (zyklusErgebnis, error) {
	ergebnis := verarbeiteZyklus(auftraege, zustelle)
	if len(ergebnis.gedruckteIDs) == 0 && len(ergebnis.fehlversuche) == 0 {
		return ergebnis, nil
	}
	return ergebnis, melde(ergebnis)
}

// verarbeiteZyklus verarbeitet die Aufträge eines Polls: gruppiert nach Ziel-IP,
// Gruppen laufen parallel (ein toter Drucker blockiert keinen anderen). Jede
// Gruppe wird genau einmal zugestellt. Die Resultate werden nach ID sortiert
// zurückgegeben.
func verarbeiteZyklus(auftraege []DruckAuftrag, zustelle gruppenDruckFunc) zyklusErgebnis {
	var (
		mu       sync.Mutex
		wg       sync.WaitGroup
		ergebnis zyklusErgebnis
	)

	for zielIP, gruppe := range bildeZustellgruppen(auftraege) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			gedruckte, fehler := zustelle(zielIP, gruppe)

			mu.Lock()
			defer mu.Unlock()
			ergebnis.gedruckteIDs = append(ergebnis.gedruckteIDs, gedruckte...)
			ergebnis.fehlversuche = append(ergebnis.fehlversuche, fehler...)
		}()
	}
	wg.Wait()

	sort.Ints(ergebnis.gedruckteIDs)
	sort.Slice(ergebnis.fehlversuche, func(i, j int) bool {
		return ergebnis.fehlversuche[i].ID < ergebnis.fehlversuche[j].ID
	})
	return ergebnis
}

// bildeZustellgruppen bildet je Ziel-IP die Gruppe, die dieser Zyklus zustellt:
// höchstens maxBonsProZyklus Aufträge, in der Eingabe-Reihenfolge (älteste ID
// zuerst). Die übrigen Aufträge bleiben offen und folgen im nächsten Zyklus.
//
// Die Obergrenze deckelt zugleich drei Größen, die sonst allein an der Länge der
// Warteschlange hängen: das Quittungsfenster einer Gruppe, die Wartezeit der
// übrigen Drucker (der nächste Poll startet erst, wenn die langsamste Gruppe
// fertig ist) und die Zahl der Bons, die ein Abbruch doppelt drucken lässt.
func bildeZustellgruppen(auftraege []DruckAuftrag) map[string][]DruckAuftrag {
	gruppen := make(map[string][]DruckAuftrag)
	for _, a := range auftraege {
		if len(gruppen[a.ZielIP]) == maxBonsProZyklus {
			continue
		}
		gruppen[a.ZielIP] = append(gruppen[a.ZielIP], a)
	}
	return gruppen
}

// verbindeMitDrucker öffnet die TCP-Verbindung zum Rohdaten-Port 9100 des
// Druckers — die eine Verbindung, über die eine ganze Gruppe zugestellt wird.
func verbindeMitDrucker(zielIP string) (net.Conn, error) {
	return net.DialTimeout("tcp", net.JoinHostPort(zielIP, "9100"), dialTimeout)
}

// gruppenErgebnis ist das vollständige Ergebnis einer Gruppen-Zustellung.
// ausgang und gesendet speisen die Logzeile — und machen den Quittungs-Ausgang
// prüfbar, ohne die Logausgabe abfangen zu müssen.
type gruppenErgebnis struct {
	gedruckteIDs []int
	fehler       []fehlversuch
	ausgang      string
	gesendet     int
}

// zustelleGruppe liefert die Zustellfunktion für eine Auftragsgruppe und
// protokolliert je Gruppe eine Zeile. Die Zustellung selbst macht stelleGruppeZu.
func zustelleGruppe(verbinde verbindeFunc, timeouts zustellTimeouts) gruppenDruckFunc {
	return func(zielIP string, auftraege []DruckAuftrag) ([]int, []fehlversuch) {
		if len(auftraege) == 0 {
			return nil, nil
		}

		start := time.Now()
		ergebnis := stelleGruppeZu(verbinde, timeouts, zielIP, auftraege)
		log.Printf("Drucker %s: %d/%d Bons gesendet, Quittung %s, Dauer %s",
			zielIP, ergebnis.gesendet, len(auftraege), ergebnis.ausgang, time.Since(start).Round(time.Millisecond))
		return ergebnis.gedruckteIDs, ergebnis.fehler
	}
}

// stelleGruppeZu stellt alle Aufträge einer Ziel-IP zu: einmal verbinden,
// Papierstatus prüfen, alle Bons in ID-Reihenfolge über dieselbe Verbindung
// schreiben, danach die Quittung einholen. Eine Verbindung je Ziel-IP und Zyklus
// statt zwei je Bon — sonst weist ein Bondrucker, der nur eine Verbindung
// gleichzeitig annimmt, die Folgeaufträge stillschweigend ab.
//
// Bricht die Gruppe vor der Quittung ab, gilt nichts als zugestellt — auch nicht
// die bereits geschriebenen Bons. Sie werden im nächsten Zyklus erneut zugestellt
// (bewusster Doppeldruck: ein doppelter Arbeitsbon kostet Papier, ein fehlender
// ein Getränk).
//
// verbinde ist der Injektionspunkt für Tests; auftraege ist nie leer (das prüft
// zustelleGruppe).
func stelleGruppeZu(verbinde verbindeFunc, timeouts zustellTimeouts, zielIP string, auftraege []DruckAuftrag) gruppenErgebnis {
	ergebnis := gruppenErgebnis{ausgang: ausgangAbgebrochen}

	conn, err := verbinde(zielIP)
	if err != nil {
		ergebnis.fehler = gruppenFehlversuche(auftraege, fmt.Errorf("drucker %s: nicht erreichbar: %w", zielIP, err))
		return ergebnis
	}
	defer func() { _ = conn.Close() }()

	if err := pruefePapier(conn, zielIP, timeouts.papier); err != nil {
		ergebnis.fehler = gruppenFehlversuche(auftraege, fmt.Errorf("drucker %s: %w", zielIP, err))
		return ergebnis
	}

	for _, a := range auftraege {
		if err := sendeBon(conn, a); err != nil {
			// Nur der betroffene Auftrag: ein Sendefehler hängt am einzelnen Bon
			// (etwa eine ungültige Payload) und darf keine Versuche der übrigen
			// Aufträge verbrauchen.
			ergebnis.fehler = []fehlversuch{{ID: a.ID, Fehler: fmt.Sprintf("drucker %s: %v", zielIP, err)}}
			return ergebnis
		}
		ergebnis.gesendet++
	}

	ergebnis.ausgang, err = holeQuittung(conn, timeouts, zielIP, len(auftraege))
	if err != nil {
		ergebnis.fehler = gruppenFehlversuche(auftraege, fmt.Errorf("drucker %s: quittung fehlgeschlagen: %w", zielIP, err))
		return ergebnis
	}
	ergebnis.gedruckteIDs = auftragsIDs(auftraege)
	return ergebnis
}

// gruppenFehlversuche macht jeden Auftrag der Gruppe zum Fehlversuch. Das ist der
// Fall bei Problemen, die die ganze Gruppe betreffen (Verbindung, Papier,
// Quittung) und sich keinem einzelnen Bon zuordnen lassen: gescheitert ist die
// Zustellung aller dieser Aufträge, also zählt der Versuch auch für alle. Nur so
// erreichen sie die Höchstzahl an Versuchen und werden im Admin als
// fehlgeschlagen sichtbar, statt unbemerkt offen zu bleiben.
func gruppenFehlversuche(auftraege []DruckAuftrag, err error) []fehlversuch {
	fehlversuche := make([]fehlversuch, 0, len(auftraege))
	for _, a := range auftraege {
		fehlversuche = append(fehlversuche, fehlversuch{ID: a.ID, Fehler: err.Error()})
	}
	return fehlversuche
}

func auftragsIDs(auftraege []DruckAuftrag) []int {
	ids := make([]int, 0, len(auftraege))
	for _, a := range auftraege {
		ids = append(ids, a.ID)
	}
	return ids
}

// pruefePapier fragt den Rollenpapier-Sensor auf der bestehenden Verbindung ab.
func pruefePapier(conn net.Conn, zielIP string, timeout time.Duration) error {
	if err := conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
		return fmt.Errorf("set write deadline: %w", err)
	}
	if _, err := conn.Write(dlePapierstatus); err != nil {
		return fmt.Errorf("status-abfrage fehlgeschlagen: %w", err)
	}

	if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return fmt.Errorf("set read deadline: %w", err)
	}
	antwort := make([]byte, 1)
	if _, err := conn.Read(antwort); err != nil {
		// Nicht jeder Drucker beantwortet die DLE-EOT-Statusabfrage (manche
		// ESC/POS-Modelle unterstützen sie schlicht nicht). Eine ausbleibende
		// Antwort gilt daher als erreichbar-und-OK, nicht als Fehler: die
		// Verbindung steht bereits, mehr lässt sich ohne Antwort nicht prüfen.
		return nil
	}

	// DLE EOT n=4 liefert den Rollenpapier-Sensorstatus (ESC/POS). Bits 2,3
	// (Maske 0x0C) melden den Near-End-Sensor (Papier fast leer), Bits 5,6
	// (Maske 0x60) den End-Sensor (Papier leer). Die übrigen Bits sind fest.
	// Quelle: Epson ESC/POS DLE EOT, bestätigt via escpos.readthedocs.io.
	if antwort[0]&0x60 != 0 {
		return fmt.Errorf("papier leer (status=0x%02X)", antwort[0])
	}
	if antwort[0]&0x0C != 0 {
		log.Printf("WARNUNG: Drucker %s meldet Papier fast leer", zielIP)
	}
	return nil
}

// sendeBon dekodiert die Payload eines Auftrags und schreibt sie auf die offene
// Verbindung. Weil alle Bons einer Gruppe dieselbe Verbindung nutzen, wirkt
// TCP-Backpressure: ist der Empfangspuffer des Druckers voll, blockiert Write bis
// zum writeTimeout und meldet einen echten Fehler, statt Daten still zu verlieren.
func sendeBon(conn net.Conn, a DruckAuftrag) error {
	escposData, err := base64.StdEncoding.DecodeString(a.Payload)
	if err != nil {
		return fmt.Errorf("ungueltiges Base64: %w", err)
	}
	if err := conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
		return fmt.Errorf("set write deadline: %w", err)
	}
	if _, err := conn.Write(escposData); err != nil {
		return fmt.Errorf("senden fehlgeschlagen: %w", err)
	}
	return nil
}

// holeQuittung verwirft Reste der Papierstatus-Runde, sendet GS r 1 und wartet
// auf die Antwort des Druckers. Sie liefert den Ausgang für das Log und einen
// Fehler genau dann, wenn die Zustellung unbestätigt bleibt:
//   - Antwort erhalten: der Drucker hat alle Bons der Gruppe verarbeitet.
//   - Lese-Timeout ohne Antwort: entweder ist der Drucker offline — bei Papierende
//     geht er laut ESC/POS-Referenz offline und führt GS r nicht mehr aus — oder er
//     unterstützt GS r schlicht nicht. Beides trennt eine erneute Papierprüfung per
//     Echtzeit-Kommando, das auch ein Offline-Drucker beantwortet: meldet sie
//     Papierende, gilt nichts als zugestellt, sonst die ganze Gruppe (sonst wäre ein
//     Drucker ohne GS r dauerhaft unbenutzbar). Andere Offline-Zustände (offener
//     Deckel, Fehlerzustand) erkennt DLE EOT n=4 nicht; sie bleiben unbeantwortet.
//   - Verbindungsabbruch (EOF, Reset, Schreibfehler): kein Timeout und kein
//     Nachweis. Die Gruppe bleibt offen.
func holeQuittung(conn net.Conn, timeouts zustellTimeouts, zielIP string, anzahlBons int) (string, error) {
	// Papierstatus-Abfrage und Quittung teilen sich denselben Lesestrom. Liegt
	// aus der Papierstatus-Runde noch ein Byte im Empfangspuffer — weil der
	// Drucker mehr als ein Byte geantwortet hat oder erst nach dem Papier-Timeout
	// antwortet —, würde es hier als Quittung gelten und die ganze Gruppe
	// bestätigen, ohne dass der Drucker irgendetwas verarbeitet hat.
	if err := spueleEmpfangspuffer(conn, timeouts.spuelen); err != nil {
		return ausgangAbgebrochen, fmt.Errorf("empfangspuffer spuelen: %w", err)
	}

	if err := conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
		return ausgangAbgebrochen, fmt.Errorf("set write deadline: %w", err)
	}
	if _, err := conn.Write(gsStatusabfrage); err != nil {
		return ausgangAbgebrochen, err
	}

	quittungTimeout := timeouts.quittungBasis + time.Duration(anzahlBons)*timeouts.quittungProBon
	if err := conn.SetReadDeadline(time.Now().Add(quittungTimeout)); err != nil {
		return ausgangAbgebrochen, fmt.Errorf("set read deadline: %w", err)
	}
	antwort := make([]byte, 1)
	if _, err := conn.Read(antwort); err != nil {
		if istTimeout(err) {
			if papierErr := pruefePapier(conn, zielIP, timeouts.papier); papierErr != nil {
				return ausgangAbgebrochen, papierErr
			}
			return ausgangUnbeantwortet, nil
		}
		return ausgangAbgebrochen, err
	}
	return ausgangBestaetigt, nil
}

// spueleEmpfangspuffer liest alles, was der Drucker bis zum Ablauf des Fensters
// noch sendet, und verwirft es. Das Fenster muss echt sein: Go liefert bei einer
// bereits abgelaufenen Lese-Deadline sofort ein i/o-Timeout, ohne gepufferte
// Bytes herauszugeben — ein Leseversuch ohne Wartezeit würde also nichts leeren.
// Lesefehler sind hier belanglos (Abbrüche fallen beim Quittungsumlauf erneut
// an); nur eine nicht setzbare Deadline wird gemeldet. Ein Byte, das erst nach
// dem Fenster eintrifft, gilt weiterhin als Quittung — DLE EOT und GS r 1
// antworten beide mit einem Statusbyte und sind am Inhalt nicht zu trennen.
func spueleEmpfangspuffer(conn net.Conn, fenster time.Duration) error {
	if err := conn.SetReadDeadline(time.Now().Add(fenster)); err != nil {
		return err
	}
	verworfen := make([]byte, 64)
	for {
		if _, err := conn.Read(verworfen); err != nil {
			return nil
		}
	}
}

// istTimeout unterscheidet ein abgelaufenes Lese-Timeout von einem echten
// Verbindungsabbruch — über die Fehler-Semantik, nicht über den Fehlertext.
func istTimeout(err error) bool {
	if errors.Is(err, os.ErrDeadlineExceeded) {
		return true
	}
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
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

	if err := pruefeRelayStatus(resp); err != nil {
		return nil, err
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
// hoch und markiert nach sechs Versuchen als fehlgeschlagen).
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

	return pruefeRelayStatus(resp)
}

// pruefeRelayStatus übersetzt den HTTP-Status einer Relay-Antwort in einen Fehler.
// Einen falschen Token meldet das Backend als 400 mit {"code":"unauthorized"} —
// daraus wird ein klarer Hinweis, weil das die häufigste Fehlkonfiguration vor
// Ort ist.
func pruefeRelayStatus(resp *http.Response) error {
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	var errResp struct {
		Code string `json:"code"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&errResp)
	if errResp.Code == "unauthorized" {
		return fmt.Errorf("ungueltiger Token -- Relay-Token pruefen")
	}
	return fmt.Errorf("unerwarteter HTTP-Status: %d", resp.StatusCode)
}
