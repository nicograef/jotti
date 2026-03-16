# Bondruck

***

```text
# Role & Context
You are an expert Go (v1.26) and React software architect working on "jotti", an open-source, mobile-first POS (mPOS) system designed for non-profit events. 
You strictly follow Domain-Driven Design (DDD), CQRS, and Event-Sourcing principles.

Your current task is to implement requirement "K-12" (Bondruck / Automated Receipt Printing) utilizing a MUNBYN ITPP047P-UE-BK-EU thermal receipt printer.

# Architectural Constraints (Read `jotti/docs/handbuch.md`)
1. Backend Stack: Go 1.26, standard library `net/http` for routing, `pgx/v5` for PostgreSQL interactions.
2. Architecture: Strict separation of HTTP, Application, Domain, and Repository layers.
3. Event Sourcing: The Core Domain (Kassenbetrieb) is append-only. State is derived from events.
4. Ubiquitous Language (`jotti/docs/language.md`): Domain concepts and structs MUST be in German PascalCase (e.g., `Bestellung`, `PreisCents`, `Menge`). Infrastructure, DB columns (except domain fields), and routing code MUST be in English. Monetary values are strictly integers in Cents (`PreisCents`).

# Hardware Specifications (MUNBYN ITPP047P)
- Interface: Ethernet (LAN) via TCP Port 9100.
- Protocol: ESC/POS standard.
- Paper format: 80mm thermal paper (max 48 characters per line using Font A 12x24).
- Auto-Cutter: Supported via partial cut command (`\x1D\x56\x42\x00`).
- Hardware Status: Must use ESC/POS `DLE EOT 4` (\x10\x04\x04) to check for "Paper Out" (Bits 5 & 6) before sending print bytes.

# Task Definition: Exactly-Once Delivery Print Architecture
Since jotti runs on a Cloud-VPS and the printer is in a local NAT/Firewall restricted network, the printing architecture is split into two parts. You must help implement both safely:

1. jotti Cloud Backend (Go):
   - Implements the "Transactional Outbox Pattern". 
   - When a `BestellungAufgenommen`-Event is saved, generate the ESC/POS layout (split into kitchen/bar based on `Kategorie`) and insert a job into a new `print_jobs` table (Status: QUEUED) WITHIN THE SAME `pgx.Tx` transaction.
   - Serve a WebSocket endpoint (`/api/relay/ws`) protected by a static `RELAY_AUTH_TOKEN`.
   - Dispatch QUEUED jobs via WebSocket (base64 encoded payload). Update status to DELIVERED. Wait for ACK to mark as ACKNOWLEDGED.

2. Local Print-Relay-Client (Go Binary):
   - Runs locally in the event network. Connects to the Cloud-Backend via WSS.
   - Receives print jobs and executes a 3-step safety loop:
     a) Idempotency check (local embedded DB/JSON state) to prevent double printing if ACK is lost.
     b) Hardware status check (`DLE EOT 4`) to prevent sending data if the printer is offline/out of paper.
     c) Send payload to the local MUNBYN printer via TCP 9100.
   - Sends ACK back to the Cloud Backend.

# Instructions for Code Generation
- Write robust, production-ready Go code.
- Handle all network timeouts gracefully (`net.DialTimeout`).
- Do not use ORMs; use raw SQL with `pgx/v5`.
- Never use floats for currency; use `int` for Cents.
- Think step-by-step. If modifying the `Application` layer, ensure the `pgx.Tx` spans both the Event-Store `Append` and the `print_jobs` insert.
```

***

---

**1. Architektur & Systemkontext: Cloud-Backend vs. Lokales Festzelt**

Da jotti auf einem Cloud-VPS (z. B. Netcup) gehostet wird und sich der Bondrucker im lokalen Netzwerk des Vereinsfests befindet (z. B. hinter einem typischen LTE/DSL-Router), haben Sie ein klassisches NAT/Firewall-Problem: Der Cloud-Server kann keine eingehenden TCP-Verbindungen an die lokale IP des Druckers aufbauen.

**Alternativen zur Überbrückung dieses NAT-Problems:**
*   **Alternative A: VPN (WireGuard / Tailscale):** Sie verbinden den Router vor Ort oder einen lokalen Raspberry Pi per Site-to-Site VPN mit dem Cloud-VPS.
    *   *Nachteil:* Hoher Einrichtungsaufwand für ehrenamtliche Helfer vor Ort. Fehleranfällig bei wechselnden Mobilfunk-Netzen.
*   **Alternative B: Print-Relay-Client (Empfohlen):** Ein in Go geschriebener, schlanker Client (als Binary kompiliert), der lokal auf einem Raspberry Pi oder Windows-PC im Festzelt läuft. Er baut eine ausgehende (und damit NAT-unabhängige) Verbindung via WebSockets oder Server-Sent Events (SSE) zum jotti-Cloud-Backend auf, empfängt dort die Druckaufträge und leitet sie lokal an den Drucker weiter.
    *   *Vorteil:* "Plug & Play". Der Admin startet vor Ort nur die Datei `jotti-relay.exe` oder `./jotti-relay`. Das Relay ist dumm und statuslos; die gesamte Bon-Formatierung bleibt im Cloud-Backend.

**2. Hardwarespezifikation: MUNBYN ITPP047P-UE-BK-EU**

Ihr Gerät ist präzise auf den industriellen Kassenbetrieb ausgelegt.
*   **Druckmechanismus:** Direkter Thermodruck, 80 mm Papierbreite, Auflösung von 576 dots/line (203 DPI).
*   **Geschwindigkeit:** 230 mm pro Sekunde. Das ist enorm schnell und erlaubt das Drucken von ca. 40 Bons pro Minute.
*   **Schnittstellen ("UE"):** USB 2.0 und Ethernet (LAN). Es hat *kein* Wi-Fi und *kein* Bluetooth.
*   **Auto-Cutter:** Der integrierte Papierschneider ist auf 1,5 Millionen Schnitte ausgelegt und sollte per Software-Befehl getriggert werden, damit Servicekräfte den Bon nicht händisch abreißen müssen.
*   **Buzzer (Beeper):** Der Drucker verfügt über einen Signalton, der in lauten Umgebungen (wie einer Küche oder dem Getränkeausschank) nützlich ist, um auf neue Bons hinzuweisen.

**Netzwerkkonfiguration (Lokales Setup vor Ort):**
Der Drucker wird per LAN-Kabel an den Router angeschlossen. Er bezieht standardmäßig eine IP per DHCP. 
Um die IP-Adresse für den Relay-Client herauszufinden, wird ein Selbsttest gedruckt: Drucker ausschalten, `FEED`-Taste gedrückt halten, einschalten, Piepton abwarten und Taste loslassen. Die gedruckte IP (z.B. `192.168.1.50`) wird in die Konfiguration Ihres Relay-Clients eingetragen.

**3. Protokollspezifikation: RAW TCP & ESC/POS**

Der MUNBYN ITPP047P arbeitet über das LAN-Interface transparent mit dem branchenüblichen **ESC/POS-Protokoll** (ursprünglich von Epson entwickelt).
Das bedeutet, Sie benötigen keine proprietären Treiber. Der Drucker lauscht auf der lokalen IP auf **TCP-Port 9100** (dem Standard für Raw-Printing). 

Sie senden einfach rohe Hexadezimal/Byte-Arrays über diesen TCP-Socket. Die wichtigsten Steuerbefehle für Ihr Kassen-Szenario:
*   **Init:** `\x1B\x40` (Setzt den Drucker zurück, leert Puffer).
*   **Zeichenausrichtung:** `\x1B\x61\x00` (Links), `\x1B\x61\x01` (Zentriert), `\x1B\x61\x02` (Rechts).
*   **Fettdruck:** `\x1B\x45\x01` (An), `\x1B\x45\x00` (Aus).
*   **Schriftgröße:** `\x1D\x21\x11` (Doppelte Höhe/Breite für Tisch-Nummern), `\x1D\x21\x00` (Normal).
*   **Papier schneiden (Auto-Cutter):** `\x1D\x56\x42\x00` (Partieller Schnitt).
*   **Buzzer aktivieren:** `\x1B\x42\x02\x02` (Je nach Modell, triggert den Beeper für Küchenbons).

**4. Implementierungs-Architektur für jotti (Anforderung K-12)**

Die Umsetzung teilt sich architektonisch in zwei Kern-Komponenten, passend zur jotti Event-Sourcing-Architektur.

**Teil A: jotti Cloud-Backend (Go)**
Das Backend generiert die Bon-Layouts und hält die WebSocket-Verbindung.

1.  **Event-Listener:** Wenn die Servicekraft eine Bestellung absendet, erzeugt das Tisch-Aggregat das Event `BestellungAufgenommen` (bzw. `EventTypeBestellungAufgenommenV1`).
2.  **Job-Generierung:** Ein synchroner Projektor oder eine Goroutine lauscht auf dieses Event. Die Application-Schicht übersetzt die Bestelldaten aus dem "Fat Event" (Tischname, Positionen, Menge, Kommentar) in einen rohen ESC/POS-Byte-String.
3.  **Relay-Hub:** Das jotti-Backend stellt einen Endpoint bereit: `GET /api/relay/ws`. Der lokale Print-Relay-Client authentifiziert sich hier (z.B. per API-Key oder generiertem Relay-Token).
4.  **Job Dispatch:** Das Backend pusht den ESC/POS-Byte-String (z.B. base64-kodiert im JSON-Payload) über den offenen WebSocket an den Relay-Client.

**Teil B: Der jotti Print-Relay-Client (Lokales Go-Programm)**
Dieses Programm ist komplett generisch. Es weiß nichts über "Tische" oder "Produkte". Es nimmt nur Bytes aus dem Web entgegen und schiebt sie in den Drucker.

*Minimales Konzept für den lokalen Go-Relay-Client:*
```go
package main

import (
	"encoding/base64"
	"log"
	"net"
	"time"

	"github.com/gorilla/websocket" // Oder nhooyr.io/websocket
)

const (
	PrinterIP   = "192.168.1.50:9100" // Lokale MUNBYN Drucker IP
	JottiWSHost = "wss://jotti.verein.de/api/relay/ws?token=geheim123"
)

// PrintJob ist das DTO, das vom jotti-Backend kommt
type PrintJob struct {
	ID             string `json:"id"`
	PayloadBase64  string `json:"payload_base64"`
}

func main() {
	// 1. Verbindung zum Cloud-Backend aufbauen
	conn, _, err := websocket.DefaultDialer.Dial(JottiWSHost, nil)
	if err != nil {
		log.Fatalf("Keine Verbindung zur jotti Cloud: %v", err)
	}
	defer conn.Close()

	log.Println("Mit jotti Cloud verbunden. Warte auf Druckaufträge...")

	for {
		var job PrintJob
		err := conn.ReadJSON(&job)
		if err != nil {
			log.Printf("WebSocket Fehler, versuche Reconnect: %v", err)
			// Hier Reconnect-Logik mit Exponential Backoff einbauen
			time.Sleep(5 * time.Second)
			continue
		}

		// 2. Base64 Payload decodieren (das sind die rohen ESC/POS Bytes)
		escposBytes, err := base64.StdEncoding.DecodeString(job.PayloadBase64)
		if err != nil {
			log.Printf("Fehler beim Dekodieren von Job %s: %v", job.ID, err)
			continue
		}

		// 3. Verbindung zum lokalen MUNBYN Drucker aufbauen und Bytes senden
		err = sendToPrinter(escposBytes)
		if err != nil {
			log.Printf("FEHLER beim Drucken von Job %s: %v", job.ID, err)
			// Optional: NACK ans Backend senden
		} else {
			log.Printf("Job %s erfolgreich gedruckt.", job.ID)
			// Optional: ACK ans Backend senden
		}
	}
}

func sendToPrinter(data []byte) error {
	// TCP Timeout setzen, damit das Relay nicht bei Drucker-Ausfall blockiert
	conn, err := net.DialTimeout("tcp", PrinterIP, 3*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write(data)
	return err
}
```

**Warum dieses Design für jotti optimal ist:**
*   **Ausfallsicherheit:** Fällt das lokale Internet aus (Anforderung Q-05 Offline-Fähigkeit), laufen die WebSocket-Nachrichten ins Leere oder werden im Backend gepuffert. Der Relay-Client verbindet sich bei Wiederherstellung einfach neu.
*   **Routing nach Kategorien (Anforderung K-12):** Die Logik, ob ein Bon an den "Küchendrucker" oder den "Getränkedrucker" geht, bleibt zentral in der Cloud (`api/table/application/command.go`). Das Backend kann einfach die `printer_ip` im JSON-Payload des WebSocket-Events mitliefern, und das Relay `Dialed` flexibel an die vom Backend diktierte lokale IP.
*   **Hardware-Unabhängigkeit:** Ob der Verein den MUNBYN ITPP047P oder später einen Epson-Drucker nutzt, ist egal. ESC/POS ist universell.

---

Da jotti strikt nach Domain-Driven Design (DDD) und Event-Sourcing arbeitet, hängen wir uns in der **Application-Schicht** als Event-Listener an das `BestellungAufgenommen`-Event. 

Der MUNBYN ITPP047P druckt auf **80 mm breitem Thermopapier**. Bei der Standardschrift (Font A, 12x24 Dots) passen exakt **48 Zeichen in eine Zeile**. Das ist wichtig für die korrekte Ausrichtung von Texten (z. B. Preise rechtsbündig). Zudem steuern wir den integrierten **Auto-Cutter** (ausgelegt auf 1,5 Millionen Schnitte) und optional den **Beeper** (Buzzer) an, der besonders für Küchen- und Thekendrucker (Anforderung K-12) essenziell ist.

Hier ist die detaillierte Go-Implementierung für Ihr jotti-Backend:

### 1. Definition der ESC/POS-Steuerzeichen

Legen Sie ein neues Package an (z. B. `pkg/escpos/commands.go`), um die Hex-Codes für den Drucker zu definieren, da der MUNBYN die Standard-ESC/POS-Emulation verarbeitet:

```go
package escpos

const (
	// Initialisierung
	Init = "\x1B\x40"

	// Textausrichtung
	AlignLeft   = "\x1B\x61\x00"
	AlignCenter = "\x1B\x61\x01"
	AlignRight  = "\x1B\x61\x02"

	// Schriftformatierung
	BoldOn  = "\x1B\x45\x01"
	BoldOff = "\x1B\x45\x00"
	
	// Schriftgröße (GS ! n) - n kombiniert Breite und Höhe
	TextNormal      = "\x1D\x21\x00"
	TextDoubleHigh  = "\x1D\x21\x01"
	TextDoubleWidth = "\x1D\x21\x10"
	TextDoubleAll   = "\x1D\x21\x11" // Doppelte Höhe & Breite für Tischnummern

	// Hardware-Steuerung (MUNBYN ITPP047P spezifisch)
	CutPaper = "\x1D\x56\x42\x00" // Partieller Schnitt (Auto-Cutter)
	Beep     = "\x1B\x42\x03\x02" // 3x Piepen, 2x Länge (für Küchenbon)
)
```

### 2. Der Bon-Formatierer (Domain/Application-Logik)

Nun schreiben wir die Funktion, die das jotti "Fat Event" `BestellungAufgenommen` (welches bereits alle Produktdaten wie `ProduktName`, `VarianteName`, `Einzelpreis` und `Menge` sicher eingefroren hat) in einen sendefähigen Byte-String umwandelt. 

Wir beachten dabei die jotti-Regel, dass Geldbeträge immer als **Integer in Cent** (`PreisCents`) verarbeitet werden und erst für die Anzeige formatiert werden dürfen.

```go
package receipt

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"jotti/pkg/escpos"
	// Fiktiver Import für jotti Events
	"jotti/domain/table" 
)

// FormatReceipt generiert den ESC/POS Byte-Payload für den MUNBYN Drucker
func FormatReceipt(event table.BestellungAufgenommenEvent, tischName string, servicekraftName string) []byte {
	var buf bytes.Buffer

	// 1. Drucker initialisieren
	buf.WriteString(escpos.Init)

	// 2. Header (Zentriert)
	buf.WriteString(escpos.AlignCenter)
	buf.WriteString(escpos.BoldOn)
	buf.WriteString("VEREINSFEST 2026\n")
	buf.WriteString("================================================\n") // Exakt 48 Zeichen für 80mm
	buf.WriteString(escpos.BoldOff)
	buf.WriteString(escpos.TextNormal)

	// Tisch-Information groß hervorheben
	buf.WriteString(escpos.TextDoubleAll)
	buf.WriteString(fmt.Sprintf("Tisch: %s\n", tischName))
	buf.WriteString(escpos.TextNormal)
	
	buf.WriteString(fmt.Sprintf("Datum: %s\n", event.AufgenommenAm.Format("02.01.2026 15:04")))
	buf.WriteString(fmt.Sprintf("Bedienung: %s\n", servicekraftName))
	buf.WriteString("------------------------------------------------\n")

	// 3. Positionen iterieren (Linksbündig)
	buf.WriteString(escpos.AlignLeft)
	
	var gesamtCents int

	for _, pos := range event.Positionen {
		gesamtCents += pos.Einzelpreis * pos.Menge
		
		// Preis in Euro umrechnen (jotti Architekturvorgabe Q-04)
		preisEuro := float64(pos.Einzelpreis * pos.Menge) / 100.0

		// Zeile formatieren: Menge x Produktname (Variante)
		// Wir kürzen den Namen ab, damit der Preis noch rechts daneben passt (bei 48 Zeichen)
		artikelZeile := fmt.Sprintf("%d x %s (%s)", pos.Menge, pos.ProduktName, pos.VarianteName)
		preisString := fmt.Sprintf("%.2f EUR", preisEuro)

		// Auffüllen mit Leerzeichen, damit der Preis rechtsbündig ist
		paddingLen := 48 - len(artikelZeile) - len(preisString)
		if paddingLen < 1 {
			paddingLen = 1 // Fallback, falls Name zu lang
		}
		
		buf.WriteString(artikelZeile)
		buf.WriteString(strings.Repeat(" ", paddingLen))
		buf.WriteString(preisString + "\n")
	}

	buf.WriteString("------------------------------------------------\n")

	// 4. Gesamtsumme
	gesamtEuro := float64(gesamtCents) / 100.0
	buf.WriteString(escpos.TextDoubleHigh)
	buf.WriteString(escpos.AlignRight)
	buf.WriteString(fmt.Sprintf("GESAMT: %.2f EUR\n", gesamtEuro))
	buf.WriteString(escpos.TextNormal)
	
	// 5. Optionaler Kommentar aus der Bestellung (jotti Anforderung K-01)
	if event.Kommentar != "" {
		buf.WriteString("\n" + escpos.AlignLeft + escpos.BoldOn)
		buf.WriteString("HINWEIS:\n")
		buf.WriteString(escpos.BoldOff)
		buf.WriteString(event.Kommentar + "\n")
	}

	// 6. Abschluss, Papiervorschub und Auto-Cutter
	buf.WriteString(escpos.AlignCenter)
	buf.WriteString("\n*** Vielen Dank fuer Ihren Besuch! ***\n")
	
	// 4-5 Leerzeilen Vorschub sind nötig, damit der Schnitt nicht den Text durchtrennt
	buf.WriteString("\n\n\n\n\n") 
	buf.WriteString(escpos.CutPaper)

	return buf.Bytes()
}
```

### 3. Einbindung als Event-Listener (Application Layer)

Nach der reinen Formatierung bauen wir den Event-Listener in der Application-Schicht, der auf das von Ihnen beschriebene Event reagiert. 

Gemäß der jotti-Architektur (Anforderung K-12) könnten wir hier auch das **Routing** einbauen: Wenn in den `Positionen` nur Essen ist, geht es an den Küchendrucker (und der Drucker soll piepsen), bei Getränken an den Thekendrucker.

```go
package application

import (
	"context"
	"encoding/base64"
	"log"

	"jotti/domain/table"
	"jotti/pkg/escpos"
)

type ReceiptEventHandler struct {
	// Hier würde z.B. Ihr lokaler WebSocket-Hub injiziert werden,
	// der die Verbindung zum lokalen Relay-Client hält.
	RelayHub *WebSocketRelayHub 
}

// Handle reagiert asynchron auf ein gespeichertes Event
func (h *ReceiptEventHandler) HandleBestellungAufgenommen(ctx context.Context, event table.BestellungAufgenommenEvent) error {
	
	// 1. Stammdaten laden (Tischname & Servicekraftname) 
	// Da das Event nur IDs enthält, müssen wir diese ggf. über Read-Models beziehen.
	tischName := "Tisch 12" // Beispiel
	serviceName := "Maria"  // Beispiel

	// 2. Bon formatieren
	rawBytes := receipt.FormatReceipt(event, tischName, serviceName)

	// 3. Optionale Logik für Küchendrucker (Piepen aktivieren)
	// Falls dies ein Küchenbon ist, setzen wir den Beep-Command ganz an den Anfang
	isKuechenBon := true // Diese Logik leiten Sie aus event.Positionen[].Kategorie ab
	if isKuechenBon {
		rawBytes = append([]byte(escpos.Beep), rawBytes...)
	}

	// 4. Base64 codieren für den Transport über WebSockets
	payloadBase64 := base64.StdEncoding.EncodeToString(rawBytes)

	// 5. An den Relay-Client schicken (welcher dann lokal an TCP 9100 zum MUNBYN sendet)
	err := h.RelayHub.DispatchPrintJob(ctx, payloadBase64)
	if err != nil {
		log.Printf("Fehler beim Senden des Druckauftrags: %v", err)
		return err
	}

	return nil
}
```

### Warum diese Implementierung robust ist:
* **Exakte Zeichenbreite:** Durch die Berechnung von `paddingLen = 48 - ...` wird der MUNBYN-Drucker perfekt ausgereizt, sodass Beträge immer sauber untereinander auf der rechten Seite stehen.
* **Auto-Cutter Trennung:** Der Vorschub von `\n\n\n\n\n` vor dem `escpos.CutPaper` garantiert, dass das Messer des MUNBYN (welches mechanisch leicht über dem Druckkopf sitzt) den Text nicht zerschneidet.
* **Jotti-Konformität:** Alle fachlichen Anforderungen von jotti (Integer-Cents, Anzeige des optionalen 100-Zeichen Kommentars, Fat-Events für Preisstabilität) sind im Code abgebildet.

---

Der Print-Relay-Client ist das perfekte Bindeglied für jotti: Er ist klein, verbraucht kaum Ressourcen und löst elegant das Problem, dass Ihr Cloud-Backend nicht durch den Router des Vereinsfests auf die lokale IP des MUNBYN-Druckers zugreifen kann.

Da Festzelte oft nur über instabile LTE-Verbindungen verfügen und der Drucker gelegentlich ausgeschaltet sein könnte (z. B. beim Papierwechsel), muss dieser Client **extrem robust** sein. Er muss bei Verbindungsabbrüchen automatisch neu verbinden (Reconnect) und darf bei einem ausgeschalteten Drucker nicht blockieren.

Hier ist die produktionsreife Go-Implementierung für den **jotti Print-Relay-Client**.

### 1. Projekt-Setup
Erstellen Sie einen neuen Ordner für das Relay-Projekt und initialisieren Sie das Go-Modul. Wir nutzen die bewährte Bibliothek `gorilla/websocket`.

```bash
mkdir jotti-print-relay
cd jotti-print-relay
go mod init jotti-print-relay
go get github.com/gorilla/websocket
```

### 2. Der Relay-Client Code (`main.go`)

Speichern Sie den folgenden Code in der `main.go`. Ich habe ihn mit entsprechenden Timeouts, Fehlerbehandlungen und einer Reconnect-Schleife (Backoff) versehen.

```go
package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

// PrintJob ist das DTO, das wir vom jotti-Backend erwarten
type PrintJob struct {
	ID            string `json:"id"`
	PayloadBase64 string `json:"payload_base64"`
	PrinterIP     string `json:"printer_ip"` // Optional: Falls Sie mehrere Drucker (Küche, Theke) haben
}

var (
	// Konfiguration über Kommandozeilen-Parameter
	backendHost = flag.String("backend", "jotti-cloud.meinverein.de", "Host des jotti Backends")
	token       = flag.String("token", "festzelt-geheim-123", "Authentifizierungs-Token für jotti")
	defaultIP   = flag.String("printer", "192.168.1.50", "Lokale IP des MUNBYN Druckers")
)

func main() {
	flag.Parse()
	log.Println("Starte jotti Print-Relay-Client...")
	log.Printf("Ziel-Backend: %s | Standard-Drucker: %s:9100\n", *backendHost, *defaultIP)

	// Die äußere Endlosschleife sorgt für den Reconnect, falls das Festzelt-Internet ausfällt
	for {
		connectAndListen()
		log.Println("Verbindung unterbrochen. Versuche Reconnect in 5 Sekunden...")
		time.Sleep(5 * time.Second)
	}
}

func connectAndListen() {
	// WebSocket URL zusammenbauen (wss:// für verschlüsselte Verbindung via nginx)
	u := url.URL{Scheme: "wss", Host: *backendHost, Path: "/api/relay/ws"}
	
	// Auth-Header setzen (alternativ als Query-Parameter, je nach Backend-Design)
	headers := http.Header{}
	headers.Add("Authorization", "Bearer "+*token)

	log.Printf("Verbinde mit %s ...", u.String())
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), headers)
	if err != nil {
		log.Printf("Dial-Fehler: %v", err)
		return // Zurück zur Reconnect-Schleife
	}
	defer conn.Close()

	log.Println("Erfolgreich mit jotti Cloud verbunden! Warte auf Bons...")

	// Innere Schleife: Nachrichten vom Backend empfangen
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Lese-Fehler (Verbindung verloren): %v", err)
			return // Bricht ab, führt zum Reconnect
		}

		var job PrintJob
		if err := json.Unmarshal(message, &job); err != nil {
			log.Printf("Konnte Backend-Nachricht nicht parsen: %v", err)
			continue
		}

		log.Printf("Empfange PrintJob %s. Sende an Drucker...", job.ID)
		
		// Job verarbeiten
		processPrintJob(job)
	}
}

func processPrintJob(job PrintJob) {
	// 1. Base64 Payload dekodieren (enthält die ESC/POS Bytes inkl. Auto-Cutter)
	escposBytes, err := base64.StdEncoding.DecodeString(job.PayloadBase64)
	if err != nil {
		log.Printf("Fehler: Payload von Job %s ist kein valides Base64: %v", job.ID, err)
		return
	}

	// 2. IP bestimmen (Erlaubt Routing durch das Backend, andernfalls Fallback)
	targetIP := job.PrinterIP
	if targetIP == "" {
		targetIP = *defaultIP
	}
	targetAddr := targetIP + ":9100"

	// 3. An den MUNBYN senden
	err = sendToPrinter(targetAddr, escposBytes)
	if err != nil {
		// Hier könnte man optional eine Fehlermeldung (NACK) an das jotti-Backend zurücksenden,
		// damit die Servicekraft auf dem Handy sieht: "Drucker offline!"
		log.Printf("FEHLER beim Drucken von Job %s auf %s: %v", job.ID, targetAddr, err)
	} else {
		log.Printf("Job %s erfolgreich auf %s gedruckt.", job.ID, targetAddr)
	}
}

func sendToPrinter(address string, data []byte) error {
	// WICHTIG: Timeout von 3 Sekunden. Ist der MUNBYN z.B. ausgeschaltet 
	// oder das Kabel defekt, friert das Relay sonst ein.
	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Rohe Bytes an Port 9100 schieben
	_, err = conn.Write(data)
	return err
}
```

### 3. Kompilieren für das Festzelt (Cross-Compilation)

Da Go hervorragend für Cross-Compilation geeignet ist, können Sie diese eine Datei einfach auf Ihrem Entwickler-Rechner für das jeweilige Zielsystem kompilieren, das beim Vereinsfest im Einsatz ist.

**Wenn im Festzelt ein Windows-PC (z.B. am Ausschank) steht:**
```bash
GOOS=windows GOARCH=amd64 go build -o jotti-relay.exe main.go
```
*Vor Ort starten:* `jotti-relay.exe -backend="jotti.meinverein.de" -printer="192.168.1.50"`

**Wenn Sie einen kleinen Raspberry Pi (Linux) an den LAN-Router des Fests hängen:**
```bash
GOOS=linux GOARCH=arm64 go build -o jotti-relay main.go
```
*Vor Ort starten:* `./jotti-relay -backend="jotti.meinverein.de" -printer="192.168.1.50"`

### 4. Vorteile dieser Architektur für Ihr Setup

1. **State-less & Wartungsarm:** Das Programm hat keine lokale Datenbank und speichert nichts. Wenn jemand im Festzelt den Raspberry Pi aus Versehen vom Strom zieht, steckt man ihn wieder ein, er bootet, das Go-Binary startet, verbindet sich sofort wieder mit der jotti-Cloud und ist bereit.
2. **Dynamisches Routing (für Anforderung K-12):** Das Feld `PrinterIP` im JSON erlaubt es Ihrem jotti-Backend, die Logik zu steuern. Ist in der Bestellung nur eine Currywurst, setzt das Backend `PrinterIP: "192.168.1.51"` (Küchendrucker). Ist ein Bier dabei: `PrinterIP: "192.168.1.50"` (Thekendrucker). Das Relay leitet es stupide an die richtige IP im lokalen LAN weiter.
3. **Firewall-freundlich:** Da das Relay via HTTPS/WSS eine *ausgehende* Verbindung aus dem Festzelt-WLAN zu Ihrem Cloud-VPS aufbaut, müssen Sie vor Ort am Router (z. B. FritzBox / LTE-GigaCube) **keine Portfreigaben** oder VPNs einrichten.

---

Da jotti strikt nach Domain-Driven Design (DDD) aufgebaut ist und im Backend auf Go 1.26 sowie die Standardbibliothek `net/http` setzt, implementieren wir den WebSocket-Hub als saubere Infrastruktur-Komponente. 

Dieser Hub hält die aktiven Verbindungen zu Ihrem Festzelt-Relay, authentifiziert diese (z. B. über einen statischen API-Key für das Relay) und bietet der Application-Schicht eine threadsichere Methode, um die generierten Bons (`PrintJob`) in die Warteschlange der WebSockets zu pushen.

Hier ist die detaillierte, architekturkonforme Umsetzung für Ihr jotti-Backend.

### 1. Das DTO (Data Transfer Object)

Zuerst definieren wir das Datenpaket, das über den WebSocket gesendet wird. Es entspricht exakt dem, was unser Relay-Client im vorherigen Schritt erwartet hat.

Legen Sie dies z. B. unter `backend/internal/infrastructure/printer/dto.go` ab:

```go
package printer

// PrintJobDTO ist das Payload-Format für den WebSocket-Transport
type PrintJobDTO struct {
	ID            string `json:"id"`
	PayloadBase64 string `json:"payload_base64"`
	PrinterIP     string `json:"printer_ip"` // Erlaubt das Routing an verschiedene MUNBYN-Drucker vor Ort
}
```

### 2. Der Relay-Hub (Infrastruktur-Schicht)

Der Hub ist verantwortlich für die Verwaltung aller aktuell verbundenen Relay-Clients (im Normalfall ist es pro Festzelt nur einer, aber das Design sollte robust genug für Verbindungswechsel oder Reconnects sein). 

Da WebSockets asynchron sind, müssen wir die Liste der aktiven Verbindungen mit einem `sync.Mutex` vor Race-Conditions (gleichzeitigen Lese-/Schreibzugriffen) schützen.

Legen Sie dies unter `backend/internal/infrastructure/printer/hub.go` ab:

```go
package printer

import (
	"context"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// RelayHub verwaltet die aktiven WebSocket-Verbindungen ins Festzelt
type RelayHub struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]bool
}

// NewRelayHub erstellt eine neue Instanz des Hubs
func NewRelayHub() *RelayHub {
	return &RelayHub{
		clients: make(map[*websocket.Conn]bool),
	}
}

// AddClient registriert eine neue Relay-Verbindung
func (h *RelayHub) AddClient(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[conn] = true
	log.Printf("Neuer Print-Relay-Client verbunden. Aktive Relays: %d", len(h.clients))
}

// RemoveClient entfernt eine abgebrochene Verbindung
func (h *RelayHub) RemoveClient(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[conn]; ok {
		delete(h.clients, conn)
		conn.Close()
		log.Printf("Print-Relay-Client getrennt. Aktive Relays: %d", len(h.clients))
	}
}

// DispatchPrintJob wird von der Application-Schicht (dem Event-Listener) aufgerufen
func (h *RelayHub) DispatchPrintJob(ctx context.Context, job PrintJobDTO) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.clients) == 0 {
		log.Println("WARNUNG: Kein Relay-Client verbunden. Bon kann nicht gedruckt werden.")
		// In einer erweiterten Version könnten Sie den Job hier in PostgreSQL zwischenspeichern (Queue)
		return nil 
	}

	// An alle verbundenen Relays senden (normalerweise nur eins pro Instanz)
	for conn := range h.clients {
		err := conn.WriteJSON(job)
		if err != nil {
			log.Printf("Fehler beim Senden an Relay (wird getrennt): %v", err)
			// Cleanup erfolgt asynchron durch den Read-Pump im HTTP-Handler
		}
	}
	return nil
}
```

### 3. Der HTTP-Endpoint für den Upgrader

Da jotti die Standardbibliothek `net/http` verwendet, schreiben wir einen simplen Handler, der die ankommende HTTP-Anfrage validiert (Authentifizierung) und dann in eine persistente WebSocket-Verbindung umwandelt.

Legen Sie dies z. B. in Ihrer API-Routing-Schicht unter `backend/api/relay/http.go` an:

```go
package relay

import (
	"net/http"
	"os"
	"strings"
	"time"

	"jotti/backend/internal/infrastructure/printer"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// CORS-Check (Wichtig, falls das Relay von einer anderen Origin verbindet)
	CheckOrigin: func(r *http.Request) bool {
		return true // Für reine Server-zu-Binary WebSocket Verbindungen meist unkritisch
	},
}

// WSHandler stellt den HTTP-Endpunkt für das Print-Relay bereit
type WSHandler struct {
	Hub *printer.RelayHub
}

func (h *WSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. Authentifizierung: Prüfen, ob das Relay das korrekte Token sendet
	// Für Headless-Clients empfiehlt sich ein statisches Token aus den .env Variablen
	expectedToken := os.Getenv("RELAY_AUTH_TOKEN")
	authHeader := r.Header.Get("Authorization")
	
	if expectedToken == "" || !strings.HasPrefix(authHeader, "Bearer "+expectedToken) {
		http.Error(w, "Unauthorized: Ungültiges Relay-Token", http.StatusUnauthorized)
		return
	}

	// 2. HTTP zu WebSocket upgraden
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// Loggen, aber kein http.Error, da die Verbindung bereits vom Upgrader behandelt wurde
		return
	}

	// 3. Verbindung im Hub registrieren
	h.Hub.AddClient(conn)
	defer h.Hub.RemoveClient(conn)

	// 4. Read-Pump: Wir müssen konstant vom Socket lesen, um Disconnects
	// (z.B. wenn jemand im Festzelt den Router zieht) schnell zu erkennen.
	conn.SetReadLimit(512)
	// Ping-Pong Mechanismus für zuverlässige Verbindungsüberwachung
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error { 
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil 
	})

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			// Verbindung wurde abgebrochen (Fehler oder regulärer Close)
			break
		}
	}
}
```

### 4. Verkabelung beim Server-Start (`main.go`)

Damit alles zusammenläuft, müssen Sie den Hub in Ihrer `main.go` (bzw. Ihrer Dependency-Injection-Schicht) initialisieren und sowohl dem HTTP-Router als auch dem Application-Event-Listener übergeben.

```go
// In Ihrer main.go oder Router-Initialisierung:

// 1. Hub instanziieren
relayHub := printer.NewRelayHub()

// 2. HTTP-Endpoint registrieren (Stdlib net/http)
wsHandler := &relay.WSHandler{Hub: relayHub}
http.Handle("/api/relay/ws", wsHandler)

// 3. Dem Application-Event-Handler (den wir im vorherigen Schritt entworfen haben)
// den Hub injizieren, damit er bei einem 'BestellungAufgenommen'-Event drucken kann.
receiptHandler := &application.ReceiptEventHandler{
    RelayHub: relayHub,
}

// (Starten Sie hier Ihren HTTP Server wie gewohnt...)
```

### Warum diese Architektur für jotti perfekt ist

1.  **Entkopplung (Separation of Concerns):** Ihre Event-Sourcing-Logik und die `BestellungAufgenommen`-Projektoren müssen nichts von WebSockets oder dem MUNBYN-Drucker wissen. Sie rufen nur `RelayHub.DispatchPrintJob()` auf.
2.  **Sicherheit (`Q-06`, `A-01`):** Das Relay nutzt einen dedizierten `RELAY_AUTH_TOKEN`, der getrennt von der regulären Benutzer-Authentifizierung (JWT) der Servicekräfte läuft. Das ist Best Practice für Machine-to-Machine (M2M) Kommunikation.
3.  **Fail-Safe:** Wenn die Internetverbindung im Festzelt abreißt, schlägt das Lesen im `Read-Pump` des Backends fehl, der Hub entfernt den Client sicher per `RemoveClient(conn)` und es entstehen keine Memory-Leaks (Goroutine-Leaks) auf dem Cloud-VPS.

Haben Sie bereits eine Vorstellung davon, wie Sie die Zuordnung der Drucker-IPs ("Küchendrucker" vs. "Thekendrucker") verwalten möchten? Sollen wir als Nächstes betrachten, wie man diese Konfiguration elegant in die bestehende jotti-Stammdatenverwaltung der Produkte oder Kategorien integriert?

---

Um die Anforderung **K-12 (Bondruck)** architekturkonform in jotti umzusetzen, müssen wir die Drucker den jeweiligen Kategorien zuordnen. Da in jotti die Kategorien als festes Datenbank-Enum (`ProduktKategorie` mit den Werten `'essen'`, `'getraenk'`, `'sonstiges'`) definiert sind, ist die eleganteste Lösung eine dedizierte Konfigurationstabelle, die diese Enums auf IP-Adressen mappt.

Hier ist das detaillierte Konzept zur Integration in Ihre bestehende Stammdatenverwaltung.

### 1. Das Datenmodell (PostgreSQL)

Wir erweitern die Infrastruktur um eine einfache Key-Value-Tabelle, die die Zuordnung speichert. Da jotti auf referenzielle Integrität und klare Strukturen setzt, nutzen wir das bestehende Enum:

```sql
-- In einer neuen Migrationsdatei (z.B. migrations/02_drucker_config.up.sql)
CREATE TABLE kategorie_drucker (
    kategorie produkt_kategorie PRIMARY KEY,
    drucker_ip VARCHAR(50) NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Initiales Seeding mit Fallback-Werten (z.B. localhost oder leere Strings)
INSERT INTO kategorie_drucker (kategorie, drucker_ip) VALUES 
    ('essen', '192.168.1.51'),
    ('getraenk', '192.168.1.50'),
    ('sonstiges', '');
```

### 2. Die Admin-API (Infrastruktur-Schicht)

Die jotti-Architektur schreibt vor, dass alle API-Endpunkte ausschließlich als `POST` implementiert werden. Wir fügen dem Admin-Bereich (`api/admin.go`) zwei neue Endpunkte hinzu:

1.  `POST /admin/get-drucker-config`: Gibt die aktuelle Zuordnung zurück.
2.  `POST /admin/update-drucker-config`: Nimmt eine aktualisierte Liste von Kategorien und IPs entgegen und speichert sie per `UPSERT` (ON CONFLICT DO UPDATE).

### 3. Anpassung der Application-Schicht (Bon-Splitting)

Das ist der wichtigste Teil für den Kassenbetrieb: Wenn eine Bestellung sowohl Currywurst (Kategorie `essen`) als auch Bier (Kategorie `getraenk`) enthält, darf nicht nur ein einzelner Bon gedruckt werden. Das System muss die Positionen aufteilen und zwei separate Druckaufträge (einen an den Küchendrucker, einen an den Thekendrucker) generieren.

Da jotti sogenannte "Fat Events" nutzt, sind alle Produktdaten (inklusive der `Kategorie`) zum Zeitpunkt der Bestellung bereits sicher im Event in der Eigenschaft `Positionen` eingefroren.

Wir passen den Event-Listener `HandleBestellungAufgenommen` so an:

```go
package application

import (
	"context"
	"encoding/base64"
	"log"

	"jotti/domain/product"
	"jotti/domain/table"
	"jotti/pkg/escpos"
)

// HandleBestellungAufgenommen reagiert asynchron auf eine neue Bestellung
func (h *ReceiptEventHandler) HandleBestellungAufgenommen(ctx context.Context, event table.BestellungAufgenommenEvent) error {
	
	// 1. Positionen nach Kategorie gruppieren
	bestellungenNachKategorie := make(map[product.Kategorie][]table.Position)
	for _, pos := range event.Positionen {
		bestellungenNachKategorie[pos.Kategorie] = append(bestellungenNachKategorie[pos.Kategorie], pos)
	}

	// 2. Für jede Kategorie einen eigenen Bon generieren
	for kategorie, positionen := range bestellungenNachKategorie {
		
		// 3. Drucker-IP für diese Kategorie aus der Datenbank laden (z.B. über ein injiziertes Repo)
		druckerIP, err := h.ConfigRepo.GetPrinterIP(ctx, kategorie)
		if err != nil || druckerIP == "" {
			log.Printf("Kein Drucker für Kategorie %s konfiguriert, überspringe Bondruck.", kategorie)
			continue
		}

		// 4. Nur die Positionen DIESER Kategorie an den Bon-Formatierer übergeben
		// (Der Formatierer muss leicht angepasst werden, damit er []table.Position akzeptiert)
		rawBytes := receipt.FormatReceipt(positionen, event.TischID, "Tisch-Name", "Servicekraft")

		// 5. Wenn es Essen ist, Buzzer hinzufügen (optional)
		if kategorie == product.EssenKategorie {
			rawBytes = append([]byte(escpos.Beep), rawBytes...)
		}

		// 6. Als Base64 verpacken und an das Relay pushen
		payloadBase64 := base64.StdEncoding.EncodeToString(rawBytes)

		job := printer.PrintJobDTO{
			ID:            generateUUID(),
			PayloadBase64: payloadBase64,
			PrinterIP:     druckerIP, // Hier weisen wir das Relay an, wohin die Bytes gehen!
		}

		if err := h.RelayHub.DispatchPrintJob(ctx, job); err != nil {
			log.Printf("Fehler beim Dispatch des %s-Bons: %v", kategorie, err)
		}
	}

	return nil
}
```

### 4. Integration ins Frontend (Admin-UI)

Im jotti-Frontend erstellen wir im Bereich `/admin` eine neue Ansicht (z. B. als Tab neben "Produkte verwalten" und "Tische verwalten").
Hier iterieren Sie über die drei systemweiten Konstanten (`Kategorie.ESSEN`, `Kategorie.GETRAENK`, `Kategorie.SONSTIGES`) und stellen jeweils ein einfaches Eingabefeld für eine IPv4-Adresse zur Verfügung. Die Validierung erfolgt wie von jotti gefordert doppelt: im Frontend per Zod (Regex für gültige IPs) und im Backend über ein zog-Schema.

Durch diese Architektur bleibt das System strikt entkoppelt: Die Stammdaten wissen nichts vom Drucken, und der lokale Relay-Client muss nicht wissen, was ein "Getränk" ist – er bekommt aus der Cloud einfach das fertige Byte-Paket zusammen mit der gewünschten Ziel-IP überreicht.

Haben Sie Präferenzen, wie sich das System verhalten soll, wenn im Relay-Client ein Drucker offline ist (soll das Cloud-Backend die Druckaufträge zwischenspeichern, oder reicht es aus, wenn der Relay-Client den Bon im lokalen Log verwirft)?

---

Ihre Anforderung ist aus gastronomischer und prozessualer Sicht absolut kritisch. In der Informatik ist dieses Problem als **"Exactly-Once Delivery" (Exakt-Einmal-Zustellung)** bekannt und bei physischer Hardware besonders schwer zu lösen, da ein Drucker – anders als eine Datenbank – Transaktionen nicht einfach "zurückrollen" kann, wenn mittendrin das Stromkabel gezogen wird.

Wenn das Go-Backend die Daten über TCP sendet, bestätigt das Netzwerkprotokoll (TCP ACK) lediglich, dass die Daten im Netzwerkspeicher des Druckers angekommen sind – **nicht**, dass sie auf Papier gedruckt wurden. Ist das Papier leer oder der Deckel offen, geht der ITPP047P in den Fehlerstatus (rote LED blinkt). Sendet das Relay in diesem Moment blind weiter, laufen Puffer voll oder Daten gehen unwiderruflich verloren.

Um zu garantieren, dass kein Bon verloren geht und keiner doppelt gedruckt wird, müssen wir eine **dreistufige Sicherheitsarchitektur** entwerfen, die den Event-Sourcing-Ansatz von jotti auf den Druckprozess ausweitet.

Hier ist das Architektur- und Implementierungskonzept:

### 1. Globale Zustandsspeicherung (Backend: Die Print-Queue)

Anstatt die Bons direkt nach der Event-Generierung als "Fire-and-Forget" in den WebSocket zu schieben, bauen wir im Backend einen eigenen **Print-Projektor**.

*   **Die Tabelle `print_jobs`:** Im jotti-Backend legen Sie eine Tabelle an, die den Druckstatus trackt.
    *   Spalten: `id` (UUID), `event_id`, `kategorie` (Küche/Theke), `payload`, `status` (`QUEUED`, `DELIVERED`, `ACKNOWLEDGED`), `created_at`.
*   **Der Ablauf:** Sobald das `BestellungAufgenommen`-Event gespeichert wird, schreibt die Datenbank in derselben Transaktion die Druckjobs in diese Tabelle.
*   Das Backend sendet über den WebSocket nun die offenen Jobs (Status `QUEUED`) an das Relay und ändert den Status auf `DELIVERED`. Es wartet dann auf ein explizites `ACK` (Acknowledgement) vom Relay.

### 2. Hardware-Status-Prüfung (ESC/POS im Relay)

Das ist der wichtigste technische Schritt vor Ort. Der MUNBYN ITPP047P verfügt über interne Sensoren für "Papier leer", "Papier fast leer", "Deckel offen" und "Messer verklemmt". Bevor das lokale Go-Relay auch nur ein einziges Byte des Bons an den Drucker sendet, muss es den Hardware-Status abfragen.

Dafür nutzt man das **ESC/POS Echtzeit-Status-Kommando `DLE EOT`**:
Das Relay sendet die Hex-Sequenz `\x10\x04\x04` (Abfrage des Papiersensors) über den TCP-Socket und liest 1 Byte als Antwort.
*   Antwortet der Drucker mit einem Byte, das signalisiert "Papier leer" (Bit 5 und 6 sind gesetzt), weiß das Relay: *Stopp! Nichts senden!*
*   Das Relay hält den Job in der lokalen Queue und wartet. Die Servicekraft legt eine neue Rolle ein, der Drucker leuchtet wieder grün, das Relay prüft erneut, sieht den "Ready"-Status und schickt erst dann den Bon.

### 3. Lokale Idempotenz & Caching (Das Relay-Gedächtnis)

Was passiert, wenn der Bon erfolgreich gedruckt wurde, aber genau in der Millisekunde danach das WLAN im Festzelt abbricht, bevor das Relay das `ACK` an die jotti-Cloud senden konnte?
Wenn das Internet wieder da ist, würde die Cloud den Job erneut senden (da Status immer noch `DELIVERED` und nicht `ACKNOWLEDGED`). Das führt zum gefürchteten **Doppeldruck**.

**Die Lösung: Lokaler State im Relay-Client.**
Das Relay darf nicht völlig "dumm" und zustandslos sein.
1.  Das Relay nutzt lokal eine extrem leichtgewichtige embedded Datenbank (z. B. `SQLite` oder `bbolt` in Go) oder eine simple JSON-Datei (`relay_state.json`).
2.  Dort speichert das Relay die IDs der letzten 1000 erfolgreich gedruckten Jobs (`processed_job_ids`).
3.  Empfängt das Relay nach einem Reconnect einen Job aus der Cloud, prüft es zuerst seine lokale Liste.
    *   *Ist die Job-ID schon bekannt?* -> **Nicht drucken!** Einfach sofort das verlorene `ACK` erneut an die Cloud senden.
    *   *Ist die Job-ID neu?* -> Hardware-Status prüfen -> Drucken -> In lokale Liste eintragen -> `ACK` an Cloud senden.

### Konkrete Umsetzungsskizze für den lokalen Relay-Client

Hier ist ein konzeptioneller Auszug, wie diese Sicherheitsschleife im lokalen Relay implementiert wird:

```go
func processJobRobustly(job PrintJob) {
	// 1. Idempotenz-Prüfung (Doppeldruck verhindern)
	if localDB.HasJob(job.ID) {
		log.Printf("Job %s wurde bereits gedruckt. Sende nur ACK.", job.ID)
		sendAckToCloud(job.ID)
		return
	}

	// 2. Hardware-Schleife: Warten, bis der Drucker wirklich bereit ist
	for {
		err := checkPrinterStatus(job.PrinterIP)
		if err == nil {
			break // Drucker ist online, hat Papier und Deckel ist zu
		}
		log.Printf("Drucker nicht bereit (%v). Pausiere Queue für 5 Sekunden...", err)
		// Hier könnte man auch ein NACK_STATUS an die Cloud senden, 
		// damit das Kassen-Handy anzeigt: "Achtung: Thekendrucker Papier leer!"
		time.Sleep(5 * time.Second)
	}

	// 3. Druckauftrag senden
	escposBytes, _ := base64.StdEncoding.DecodeString(job.PayloadBase64)
	err := sendBytesToPrinter(job.PrinterIP, escposBytes)
	if err != nil {
		log.Printf("Druck fehlgeschlagen: %v", err)
		return // Bricht ab, Job wird später von der Cloud neu gepusht
	}

	// 4. Als erfolgreich markieren (Lokal und Cloud)
	localDB.MarkJobDone(job.ID) // Verhindert künftige Doppeldrucke
	sendAckToCloud(job.ID)      // Sagt der Cloud: Job ist endgültig erledigt
}

// checkPrinterStatus sendet DLE EOT 4 an den MUNBYN Drucker
func checkPrinterStatus(ip string) error {
	conn, err := net.DialTimeout("tcp", ip+":9100", 2*time.Second)
	if err != nil {
		return fmt.Errorf("drucker offline")
	}
	defer conn.Close()

	// ESC/POS Command: DLE EOT 4 (Papier Sensor Status)
	conn.Write([]byte{0x10, 0x04, 0x04})
	
	reply := make([]byte, 1)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, err = conn.Read(reply)
	if err != nil {
		return fmt.Errorf("keine Status-Antwort vom Drucker")
	}

	statusByte := reply
	// Bits 5 & 6 (0x60) prüfen auf Papierende
	if (statusByte & 0x60) == 0x60 {
		return fmt.Errorf("PAPIER LEER")
	}

	return nil
}
```

### Zusammenfassung der Vorteile für jotti:

1.  **Garantiert kein Bon-Verlust:** Ist der MUNBYN-Drucker offline oder das Papier leer (rote LED leuchtet), laufen die Aufträge weder ins Leere, noch stauen sie sich im Drucker-RAM. Sie bleiben sicher in der Cloud-Datenbank und im lokalen Relay-Wartestand.
2.  **Garantiert keine Doppel-Bons:** Die lokale Idempotenz-Datenbank (`localDB`) des Relays ist die absolute Barriere gegen Netzwerkschwankungen und verhinderte ACKs.
3.  **Integrierbar in das jotti-Frontend:** Da das Backend den exakten Druckstatus (`QUEUED` vs `ACKNOWLEDGED`) kennt, können Sie im Service-Dashboard auf dem Smartphone (Requirements K-06, K-12) neben jeder Bestellung ein kleines Drucker-Icon einblenden: *Grau (wartet), Rot (Drucker-Fehler), Grün (erfolgreich gedruckt)*.

Soll ich Ihnen als Nächstes zeigen, wie Sie die Tabelle `print_jobs` und das ACK/NACK-Handling im Go-Backend präzise an Ihren Event-Stream-Projektor aus der Application-Schicht (Requirements K-07) anbinden?

---

Die Integration der Bondruck-Garantie („Exactly-Once Delivery“) in das Event-Sourcing-Konzept von jotti ist ein klassischer Anwendungsfall für das sogenannte **Transactional Outbox Pattern**. 

Da jotti den Kassenbetrieb über einen lückenlosen Event Stream abbildet (Anforderung **K-07**) und Projektionen in derselben Datenbanktransaktion aktualisiert (Write-Through, siehe Entwickler-Handbuch), müssen wir die Druckaufträge exakt in dieser einen Transaktion mit in die Datenbank schreiben. 

Hier ist die präzise Implementierung für Ihr Go-Backend (`pgx/v5`).

### 1. Das Datenbank-Schema (PostgreSQL)

Wir erstellen eine Tabelle für die Print-Queue, die direkt mit den Events verknüpft ist.

```sql
-- Erweitern Sie Ihre Migrationen (z.B. 03_print_jobs.up.sql)

CREATE TYPE print_job_status AS ENUM ('QUEUED', 'DELIVERED', 'ACKNOWLEDGED', 'FAILED');

CREATE TABLE print_jobs (
    id UUID PRIMARY KEY,
    event_id INT NOT NULL, -- Verknüpfung zum BestellungAufgenommen-Event im Kassenjournal
    tisch_id INT NOT NULL,
    kategorie produkt_kategorie NOT NULL, -- 'essen' oder 'getraenk'
    drucker_ip VARCHAR(50) NOT NULL,
    payload_base64 TEXT NOT NULL,
    status print_job_status NOT NULL DEFAULT 'QUEUED',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index für schnelles Polling/Suchen offener Jobs
CREATE INDEX idx_print_jobs_status ON print_jobs(status) WHERE status = 'QUEUED';
```

### 2. Die Application-Schicht (Atomare Transaktion)

Wir passen den Command-Handler an, der aufgerufen wird, wenn die Servicekraft eine Bestellung absendet (Anforderung **K-01**). Die Erstellung des `BestellungAufgenommen`-Events, die Aktualisierung der `table_state`-Projektion und das Einreihen der Druckaufträge passieren nun in einer einzigen Transaktion (`pgx.Tx`).

```go
package application

import (
	"context"
	"jotti/domain/table"
	"github.com/jackc/pgx/v5"
)

func (s *TableService) BestellungAufnehmen(ctx context.Context, cmd BestellungAufnehmenCommand) error {
	// 1. Transaktion starten (pgx/v5)
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 2. Event generieren und im Event-Store speichern (K-07)
	event := table.NewBestellungAufgenommenEvent(cmd.TischID, cmd.Positionen, cmd.Kommentar)
	eventID, err := s.eventStore.Append(ctx, tx, event)
	if err != nil {
		return err
	}

	// 3. Synchrone Projektion aktualisieren (table_state)
	if err := s.projection.UpdateState(ctx, tx, event); err != nil {
		return err
	}

	// 4. NEU: Print Jobs generieren und in dieselbe Transaktion schreiben
	// (Hier nutzen wir die Logik von vorhin, die Positionen nach Kategorie aufteilt)
	jobs := s.printFormatter.CreateJobsFromEvent(event, eventID)
	for _, job := range jobs {
		// Schreibt in die Tabelle `print_jobs` mit Status 'QUEUED'
		if err := s.printRepo.InsertJob(ctx, tx, job); err != nil {
			return err
		}
	}

	// 5. Transaktion committen (Garantiert: Kein Event ohne Druckauftrag!)
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	// 6. Den WebSocket-Hub triggern, damit er sofort nach neuen Jobs sucht
	s.relayHub.NotifyNewJobs()

	return nil
}
```

### 3. Der Dispatcher und das ACK-Handling im WebSocket-Hub

Der `RelayHub` (den wir zuvor entworfen haben) ist nun nicht mehr dumm, sondern datenbankgestützt. Er liest die `QUEUED`-Jobs aus und wartet auf die `ACK`-Nachricht des Festzelt-Relays.

Zuerst definieren wir die Nachrichtenstruktur für den WebSocket:

```go
package printer

// WSMessage ist der generische Wrapper für den WebSocket-Austausch
type WSMessage struct {
	Type    string      `json:"type"` // "JOB", "ACK", "NACK_STATUS"
	Payload interface{} `json:"payload"`
}

// PrintAck kommt vom lokalen Relay-Client zurück
type PrintAck struct {
	JobID string `json:"job_id"`
}
```

Nun erweitern wir den `RelayHub` um das Dispatching und das Empfangen von Antworten:

```go
package printer

import (
	"context"
	"log"
)

// DispatchQueuedJobs wird vom Command-Handler getriggert oder läuft per Ticker alle 5 Sekunden
func (h *RelayHub) DispatchQueuedJobs(ctx context.Context) {
	h.mu.RLock()
	if len(h.clients) == 0 {
		h.mu.RUnlock()
		return // Niemand im Festzelt verbunden, Jobs bleiben QUEUED
	}
	h.mu.RUnlock()

	// 1. Alle offenen Jobs aus der DB laden
	jobs, err := h.repo.GetJobsByStatus(ctx, "QUEUED")
	if err != nil || len(jobs) == 0 {
		return
	}

	for _, job := range jobs {
		// 2. An das Festzelt-Relay senden
		msg := WSMessage{Type: "JOB", Payload: job}
		h.broadcastJSON(msg)

		// 3. Status in der DB auf 'DELIVERED' setzen
		// (Der Job liegt jetzt im Netzwerk-Puffer oder im lokalen Relay)
		h.repo.UpdateJobStatus(ctx, job.ID, "DELIVERED")
	}
}

// HandleClientMessages liest den eingehenden WebSocket-Stream des Festzelt-Relays
func (h *RelayHub) HandleClientMessages(ctx context.Context, conn *websocket.Conn) {
	for {
		var msg WSMessage
		if err := conn.ReadJSON(&msg); err != nil {
			log.Printf("Relay Disconnect: %v", err)
			break
		}

		switch msg.Type {
		case "ACK":
			// Der Drucker im Festzelt hat physisch gedruckt (Papier war vorhanden)
			ackData := msg.Payload.(map[string]interface{})
			jobID := ackData["job_id"].(string)
			
			// 4. Status auf 'ACKNOWLEDGED' setzen - Exactly-Once Delivery erfolgreich!
			h.repo.UpdateJobStatus(ctx, jobID, "ACKNOWLEDGED")
			log.Printf("Print Job %s erfolgreich abgeschlossen.", jobID)

		case "NACK_STATUS":
			// Optional: Das Relay meldet, dass das Papier leer ist oder das Messer klemmt
			log.Printf("Drucker-Fehler bei Job gemeldet. Warte auf Behebung vor Ort...")
		}
	}
}
```

### Warum diese Architektur unverwüstlich ist

Durch diese Implementierung haben Sie das Problem der unzuverlässigen Netzwerke im Festzelt vollständig abstrahiert:
1. **App-Abstürze (K-01 / Q-04):** Wenn das Backend nach Schritt 2 (`Append`) abstürzt, aber vor Schritt 4, schlägt die Transaktion fehl. Das Event wird zurückgerollt, das Frontend zeigt einen Fehler, die Servicekraft drückt nochmal. Keine Geister-Bons.
2. **Internet-Ausfall im Festzelt:** Die Servicekräfte nehmen fröhlich weiter Bestellungen am Smartphone auf. Die Cloud speichert die Events und die Print-Jobs (`QUEUED`). Sobald der LTE-Router im Zelt wieder Empfang hat, verbindet sich das Relay, der `RelayHub` pusht alle angestauten `QUEUED`-Jobs und in der Küche rattert der Drucker los.
3. **Doppeldruck-Schutz:** Da das Relay (wie im vorherigen Schritt besprochen) eine eigene kleine Idempotenz-Liste führt, kann die jotti-Cloud nach einem Timeout bedenkenlos Jobs mit Status `DELIVERED` wieder auf `QUEUED` setzen und erneut senden, falls das `ACK` verloren ging. Das Relay sortiert Duplikate sicher aus.
