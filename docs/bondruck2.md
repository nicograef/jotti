# Bondruck (K-12) — Architekturentscheidung

## Inhaltsverzeichnis

1. [Überblick](#1-überblick)
2. [Hardware: MUNBYN ITPP047P-UE](#2-hardware-munbyn-itpp047p-ue)
3. [Systemarchitektur](#3-systemarchitektur)
4. [Cloud-Backend-Komponente](#4-cloud-backend-komponente)
5. [Print-Relay-Komponente (lokal)](#5-print-relay-komponente-lokal)
6. [ESC/POS Bon-Format](#6-escpos-bon-format)
7. [Kategorie-Drucker-Konfiguration](#7-kategorie-drucker-konfiguration)
8. [Verworfene Alternativen](#8-verworfene-alternativen)

---

## 1. Überblick

Anforderung **K-12** erfordert, dass beim Aufnehmen einer Bestellung automatisch Bons an die zuständige Ausgabestation gedruckt werden (Essen → Küchenbon, Getränke → Thekenbon). Da jotti auf einem Cloud-VPS läuft, sich der Drucker aber im lokalen Netzwerk des Vereinsfests befindet (NAT/Firewall), ist eine Zwei-Komponenten-Architektur notwendig.

**Kernentscheidungen im Überblick:**

| Aspekt | Entscheidung | Begründung |
|---|---|---|
| Netzwerk-Brücke | Print-Relay-Client (lokales Binary) | Einfachste Lösung für NAT-Problem, kein VPN-Setup nötig |
| Transport | HTTP-Polling (POST) | Passt zur POST-only-Architektur, keine zusätzlichen Protokolle |
| Zustellgarantie | Transactional Outbox + lokale Idempotenz | Kein Bonverlust und kein Doppeldruck |
| Protokoll | ESC/POS via TCP 9100 | Herstellerstandard, treiberfrei, universell |
| Bon-Aufteilung | Ein Bon pro Kategorie | Küche und Theke erhalten nur ihre relevanten Positionen |

---

## 2. Hardware: MUNBYN ITPP047P-UE

Das „UE" im Modellnamen steht für **USB + Ethernet**. Dieses konkrete Modell hat kein WLAN und kein Bluetooth — Verbindung ausschließlich per LAN-Kabel.

**Relevante technische Eckdaten:**

| Eigenschaft | Wert |
|---|---|
| Schnittstelle | Ethernet (LAN), TCP Port 9100 |
| Protokoll | ESC/POS (Epson-Standard) |
| Papierbreite | 80 mm (Druckbreite 72 mm) |
| Auflösung | 576 dots/line (203 DPI) |
| Druckgeschwindigkeit | 230 mm/s (~40 Bons/min) |
| Font A | 12×24 Dots → **48 Zeichen pro Zeile** |
| Auto-Cutter | bis 1,5 Mio. Schnitte, Partial-Cut via `GS V B 0` |
| Buzzer | Ja, ansteuerbar via `ESC B n1 n2` |
| Druckkopf-Lebensdauer | 150 km |
| Eingangspuffer | 256 kByte |
| Statussensoren | Papier leer, Papier fast leer, Deckel offen, Messerklemme |

**IP-Adresse ermitteln:** Drucker ausschalten → FEED-Taste gedrückt halten → einschalten → Piepton abwarten → Taste loslassen. Der Selbsttest druckt die aktuelle IP (DHCP). Diese IP wird fest in der Relay-Konfiguration eingetragen. Empfehlung: statische IP per DHCP-Reservierung am Router des Vereinsfests.

**Wichtige ESC/POS-Befehle:**

```go
package escpos

// Initialisierung
const Init = "\x1B\x40"

// Ausrichtung
const AlignLeft   = "\x1B\x61\x00"
const AlignCenter = "\x1B\x61\x01"
const AlignRight  = "\x1B\x61\x02"

// Schrift
const BoldOn  = "\x1B\x45\x01"
const BoldOff = "\x1B\x45\x00"

// Schriftgröße (GS ! n)
const TextNormal      = "\x1D\x21\x00"
const TextDoubleHigh  = "\x1D\x21\x01" // Doppelte Höhe
const TextDoubleWidth = "\x1D\x21\x10" // Doppelte Breite
const TextDoubleAll   = "\x1D\x21\x11" // Doppelte Höhe und Breite (für Tischnummer)

// Hardware
const CutPaper = "\x1D\x56\x42\x00" // Partial Cut (GS V B 0)
const Beep     = "\x1B\x42\x03\x02" // 3 Piepser, Dauer 2 (ESC B n1 n2)

// Hardware-Statusabfrage
const StatusPaper = "\x10\x04\x04" // DLE EOT 4 — liefert 1 Byte zurück
// Antwortbyte: Bit 5 (0x20) = Papier fast leer, Bit 6 (0x40) = Papier leer
// Drucker "bereit" wenn: (antwort & 0x60) == 0
```

**Hinweis Auto-Cutter:** Vor dem Schnitt müssen mindestens 4–5 Leerzeilen gedruckt werden, da das Messer mechanisch ca. 3 mm über dem Druckkopf sitzt — sonst schneidet es in den Text.

---

## 3. Systemarchitektur

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Cloud-VPS (jotti)                             │
│                                                                      │
│  ┌───────────────┐   ┌─────────────────┐   ┌─────────────────────┐ │
│  │   Frontend    │   │  Go Backend     │   │    PostgreSQL       │ │
│  │  (Browser,    │──▶│  POST /service/ │──▶│  events (append)   │ │
│  │   Smartphone) │   │  bestellung-    │   │  table_state       │ │
│  │               │   │  aufnehmen      │   │  print_jobs        │ │
│  └───────────────┘   │                 │   │  kategorie_drucker │ │
│                      │  POST /relay/   │   └─────────────────────┘ │
│                      │  get-jobs       │                            │
│                      │  POST /relay/   │                            │
│                      │  ack-job        │                            │
│                      └────────┬────────┘                            │
└───────────────────────────────┼─────────────────────────────────────┘
                                │ HTTPS (ausgehende Verbindung vom
                                │ Festzelt → NAT-freundlich)
┌───────────────────────────────┼─────────────────────────────────────┐
│              Vereinsfest (lokales LAN)                               │
│                                                                      │
│                      ┌────────▼────────┐                            │
│                      │  Print-Relay    │                            │
│                      │  (Go Binary)    │                            │
│                      │  Raspberry Pi   │                            │
│                      │  oder Windows-  │                            │
│                      │  PC             │                            │
│                      └────────┬────────┘                            │
│                               │                                      │
│              ┌────────────────┴─────────────────┐                   │
│              │                                  │                   │
│     ┌────────▼──────┐                  ┌────────▼──────┐           │
│     │  MUNBYN       │                  │  MUNBYN       │           │
│     │ ITPP047P      │                  │ ITPP047P      │           │
│     │ Getränke      │                  │ Küche/Essen   │           │
│     │ 192.168.1.50  │                  │ 192.168.1.51  │           │
│     │ TCP:9100      │                  │ TCP:9100      │           │
│     └───────────────┘                  └───────────────┘           │
└──────────────────────────────────────────────────────────────────────┘
```

**Ablauf Ende zu Ende:**

```
Servicekraft tippt Bestellung ab
        │
        ▼
POST /service/bestellung-aufnehmen
        │
        ▼ (eine Datenbanktransaktion)
┌───────────────────────────────────────┐
│ 1. events INSERT                      │
│    (BestellungAufgenommenV1)          │
│ 2. table_state UPSERT (Projektion)    │
│ 3. print_jobs INSERT                  │
│    (ein Eintrag pro Kategorie,        │
│     Status = QUEUED)                  │
└───────────────────────────────────────┘
        │ COMMIT
        ▼
Relay pollt POST /relay/get-jobs alle ~2 Sekunden
        │
        ▼
Relay empfängt QUEUED-Jobs
        │
        ▼
Pro Job: Idempotenzcheck → Hardware-Check → TCP 9100 → ACK
        │
        ▼
POST /relay/ack-job → print_jobs Status = ACKNOWLEDGED
```

---

## 4. Cloud-Backend-Komponente

### 4.1 Datenbankschema

Die `print_jobs`-Tabelle wird **in derselben Transaktion** wie das Event geschrieben (Outbox Pattern). Damit ist garantiert: Es gibt kein Event ohne Druckauftrag und keinen Druckauftrag ohne Event.

```sql
-- In database/migrations/01_initial.up.sql ergänzen

CREATE TYPE print_job_status AS ENUM ('QUEUED', 'ACKNOWLEDGED');

CREATE TABLE print_jobs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id    INT NOT NULL REFERENCES events(id),
    tisch_id    INT NOT NULL,
    kategorie   TEXT NOT NULL CHECK (kategorie IN ('essen', 'getraenk', 'sonstiges')),
    drucker_ip  VARCHAR(50) NOT NULL,
    payload     TEXT NOT NULL,         -- Base64-kodierter ESC/POS-Byte-String
    status      print_job_status NOT NULL DEFAULT 'QUEUED',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index für schnelles Polling der offenen Jobs
CREATE INDEX idx_print_jobs_queued ON print_jobs(created_at)
    WHERE status = 'QUEUED';

CREATE TABLE kategorie_drucker (
    kategorie   TEXT PRIMARY KEY CHECK (kategorie IN ('essen', 'getraenk', 'sonstiges')),
    drucker_ip  VARCHAR(50) NOT NULL DEFAULT '',
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO kategorie_drucker (kategorie) VALUES ('essen'), ('getraenk'), ('sonstiges');
```

**Hinweis zu DELIVERED:** Die bondruck.md schlägt einen zusätzlichen Status `DELIVERED` vor (zwischen QUEUED und ACKNOWLEDGED). Dieser Zwischenstatus bringt bei HTTP-Polling keinen Mehrwert: Der Relay-Client weiß durch seine lokale Idempotenzprüfung selbst, ob ein Job bereits bearbeitet wurde. `DELIVERED` entfällt damit.

### 4.2 Outbox-Integration in WriteEvent

Die bestehende `WriteEvent`-Methode im `event_repo` startet bereits eine Transaktion für Event + Projektion. Sie wird um einen optionalen Hook erweitert, damit die Application-Schicht innerhalb derselben Transaktion weitere Writes ausführen kann:

```go
// backend/repository/event_repo/repo.go

// TxHook ist ein optionaler Callback, der innerhalb der WriteEvent-Transaktion aufgerufen wird.
// Wird für das Outbox-Pattern genutzt (z.B. print_jobs INSERT).
type TxHook func(ctx context.Context, tx *sql.Tx, eventID int) error

// WriteEvent speichert ein Event und aktualisiert synchron die table_state-Projektion.
// Optionale TxHooks laufen in derselben Transaktion (Outbox Pattern).
func (r Repository) WriteEvent(ctx context.Context, e event.Event, hooks ...TxHook) (int, error) {
    tx, err := r.DB.BeginTx(ctx, nil)
    // ... (bestehende Logik unverändert)

    // NEU: Hooks in derselben Transaktion ausführen
    for _, hook := range hooks {
        if err := hook(ctx, tx, id); err != nil {
            return 0, err
        }
    }

    if err := tx.Commit(); err != nil {
        return 0, db.Error(err)
    }
    return id, nil
}
```

**Warum Hooks statt direkte Kopplung:** Das `event_repo` kennt weiterhin nichts über Bondruck. Die Application-Schicht entscheidet, welche Hooks (falls überhaupt) übergeben werden. Events die kein Printing auslösen (z.B. `ZahlungKassiert`) erhalten keine Hooks.

### 4.3 Application-Schicht: BestellungAufnehmen

```go
// backend/api/table/application/command.go (Auszug)

func (c Command) BestellungAufnehmen(ctx context.Context, userID int, userName string,
    tischID int, inputs []BestellPositionInput, kommentar string) error {

    // ... (bestehende Logik: Positionen anreichern, Event erstellen, OCC) ...

    // Outbox-Hook: print_jobs in derselben Transaktion schreiben
    printHook := c.PrintJobRepo.CreateJobsHook(ctx, tischName, event)

    subject := "tisch:" + strconv.Itoa(tischID)
    if err := writeEvent(ctx, c.EventRepo, event, subject, printHook); err != nil {
        return err
    }

    return nil
}
```

```go
// backend/api/table/application/print_job_repo.go

type PrintJobRepo interface {
    // CreateJobsHook gibt einen TxHook zurück, der für jede Kategorie in der Bestellung
    // einen print_jobs-Eintrag schreibt (sofern ein Drucker konfiguriert ist).
    CreateJobsHook(ctx context.Context, tischName string, event event.Event) event_repo.TxHook
    // GetQueuedJobs gibt alle offenen Druckaufträge zurück.
    GetQueuedJobs(ctx context.Context) ([]PrintJob, error)
    // AckJob markiert einen Job als ACKNOWLEDGED.
    AckJob(ctx context.Context, jobID string) error
}

type PrintJob struct {
    ID        string
    TischID   int
    Kategorie string
    DruckerIP string
    Payload   string // Base64 ESC/POS
    CreatedAt time.Time
}
```

### 4.4 HTTP-Endpunkte für den Relay-Client

Alle Endpunkte sind POST-only und werden durch einen statischen `RELAY_AUTH_TOKEN` (`.env`) gesichert — unabhängig von der regulären JWT-Authentifizierung der Servicekräfte.

**`POST /relay/get-jobs`**
- Request: `{"token": "..."}` 
- Response: `{"jobs": [{"id": "...", "drucker_ip": "192.168.1.50", "payload": "<base64>"}]}`
- Gibt alle Jobs mit `status = QUEUED` zurück (maximal 50 pro Anfrage)

**`POST /relay/ack-job`**
- Request: `{"token": "...", "job_id": "..."}`
- Response: `{"ok": true}`
- Setzt `status = ACKNOWLEDGED`, `updated_at = NOW()`

**Keine explizite NACK-Logik nötig:** Jobs in `QUEUED` werden beim nächsten Poll automatisch erneut ausgeliefert, bis ein ACK erfolgt. Das Relay entscheidet lokal über Idempotenz.

---

## 5. Print-Relay-Komponente (lokal)

### 5.1 Design-Prinzipien

Das Relay ist ein schlankes Go-Binary ohne externe Abhängigkeiten. Es:
- ist **nicht stateless** (entgegen der ursprünglichen Annahme) — es führt eine lokale Idempotenzliste
- kennt **keine Domänenbegriffe** (kein "Tisch", kein "Bestellung") — es verarbeitet nur Bytes
- läuft **ohne Konfigurationsdatei** — alle Parameter kommen als Flags oder Umgebungsvariablen
- überlebt **Neustarts** — der lokale State ist persistent (JSON-Datei)

### 5.2 Dreistufige Sicherheitsschleife (Exactly-Once Delivery)

Pro empfangenem Druckauftrag:

```
Job erhalten
    │
    ├─ 1. Idempotenzcheck: Ist job.ID bereits in relay_state.json?
    │      JA  → Nur ACK senden, nicht drucken (Doppeldruck-Schutz)
    │      NEIN → weiter
    │
    ├─ 2. Hardware-Check: Sende DLE EOT 4 (\x10\x04\x04) an TCP 9100
    │      Kein TCP-Connect → warten (5s), erneut prüfen
    │      Drucker antwortet mit "Papier leer" (Bit 5+6 gesetzt) → warten (5s)
    │      Drucker bereit → weiter
    │
    └─ 3. Drucken: Base64 dekodieren, ESC/POS-Bytes an TCP 9100 senden
           Erfolg → Job-ID in relay_state.json speichern → ACK senden
           Fehler → nicht ACKen (Backend liefert beim nächsten Poll erneut)
```

### 5.3 Implementierung (main.go)

```go
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
    "time"
)

// RelayState speichert IDs bereits gedruckter Jobs (Idempotenz)
type RelayState struct {
    ProcessedIDs []string `json:"processed_ids"` // Rolling window, max 2000
}

// PrintJob ist das DTO vom jotti-Backend
type PrintJob struct {
    ID        string `json:"id"`
    DruckerIP string `json:"drucker_ip"`
    Payload   string `json:"payload"` // Base64 ESC/POS
}

var (
    backendURL  = flag.String("backend", "https://jotti.meinverein.de", "jotti Backend URL")
    token       = flag.String("token", "", "RELAY_AUTH_TOKEN aus .env")
    stateFile   = flag.String("state", "relay_state.json", "Pfad zur lokalen State-Datei")
    pollSeconds = flag.Int("poll", 2, "Poll-Intervall in Sekunden")
)

func main() {
    flag.Parse()
    if *token == "" {
        log.Fatal("--token ist erforderlich")
    }

    log.Printf("jotti Print-Relay gestartet | Backend: %s | Poll: %ds", *backendURL, *pollSeconds)

    state := loadState(*stateFile)
    client := &http.Client{Timeout: 10 * time.Second}

    for {
        jobs, err := fetchJobs(client)
        if err != nil {
            log.Printf("Fehler beim Abrufen der Jobs: %v", err)
        } else {
            for _, job := range jobs {
                processJob(client, state, job)
                saveState(*stateFile, state)
            }
        }
        time.Sleep(time.Duration(*pollSeconds) * time.Second)
    }
}

func processJob(client *http.Client, state *RelayState, job PrintJob) {
    // 1. Idempotenzcheck
    if hasProcessed(state, job.ID) {
        log.Printf("Job %s bereits gedruckt — sende nur ACK", job.ID)
        ackJob(client, job.ID)
        return
    }

    // 2. Hardware-Check (Wiederholungsschleife)
    for {
        if err := checkPrinter(job.DruckerIP); err != nil {
            log.Printf("Drucker %s nicht bereit: %v — warte 5s", job.DruckerIP, err)
            time.Sleep(5 * time.Second)
            continue
        }
        break
    }

    // 3. Drucken
    escpos, err := base64.StdEncoding.DecodeString(job.Payload)
    if err != nil {
        log.Printf("Job %s: ungültiges Base64: %v", job.ID, err)
        return
    }

    if err := sendToPrinter(job.DruckerIP, escpos); err != nil {
        log.Printf("Job %s: Druckfehler: %v", job.ID, err)
        return // Kein ACK → Backend liefert erneut
    }

    // 4. Als erfolgreich markieren
    markProcessed(state, job.ID)
    ackJob(client, job.ID)
    log.Printf("Job %s erfolgreich gedruckt auf %s", job.ID, job.DruckerIP)
}

func checkPrinter(ip string) error {
    conn, err := net.DialTimeout("tcp", ip+":9100", 3*time.Second)
    if err != nil {
        return fmt.Errorf("nicht erreichbar: %w", err)
    }
    defer conn.Close()

    // DLE EOT 4: Papiersensor-Status abfragen
    if _, err := conn.Write([]byte{0x10, 0x04, 0x04}); err != nil {
        return fmt.Errorf("status-abfrage fehlgeschlagen: %w", err)
    }

    if err := conn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
        return fmt.Errorf("set read deadline: %w", err)
    }
    reply := make([]byte, 1)
    if _, err := conn.Read(reply); err != nil {
        // Manche Drucker antworten nicht auf Status-Abfragen — kein Fehler
        return nil
    }

    if reply[0]&0x60 == 0x60 {
        return fmt.Errorf("papier leer (status=0x%02X)", reply[0])
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

func fetchJobs(client *http.Client) ([]PrintJob, error) {
    body, _ := json.Marshal(map[string]string{"token": *token})
    resp, err := client.Post(*backendURL+"/relay/get-jobs", "application/json", bytes.NewReader(body))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result struct {
        Jobs []PrintJob `json:"jobs"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    return result.Jobs, nil
}

func ackJob(client *http.Client, jobID string) {
    body, _ := json.Marshal(map[string]string{"token": *token, "job_id": jobID})
    resp, err := client.Post(*backendURL+"/relay/ack-job", "application/json", bytes.NewReader(body))
    if err != nil {
        log.Printf("ACK für %s fehlgeschlagen: %v", jobID, err)
        return
    }
    resp.Body.Close()
}

// --- State-Verwaltung ---

func loadState(path string) *RelayState {
    data, err := os.ReadFile(path)
    if err != nil {
        return &RelayState{}
    }
    var s RelayState
    if err := json.Unmarshal(data, &s); err != nil {
        return &RelayState{}
    }
    return &s
}

func saveState(path string, s *RelayState) {
    data, _ := json.Marshal(s)
    if err := os.WriteFile(path, data, 0600); err != nil {
        log.Printf("WARNUNG: Relay-State konnte nicht gespeichert werden: %v (Doppeldruck möglich nach Neustart)", err)
    }
}

func hasProcessed(s *RelayState, id string) bool {
    for _, v := range s.ProcessedIDs {
        if v == id {
            return true
        }
    }
    return false
}

func markProcessed(s *RelayState, id string) {
    s.ProcessedIDs = append(s.ProcessedIDs, id)
    // Rolling window: maximal 2000 IDs vorhalten (reicht für ein Vereinsfest)
    if len(s.ProcessedIDs) > 2000 {
        s.ProcessedIDs = s.ProcessedIDs[len(s.ProcessedIDs)-2000:]
    }
}
```

### 5.4 Deployment (Cross-Compilation)

Das Relay hat **keine externen Abhängigkeiten** (nur Go-Stdlib). Ein einzelnes Binary:

```bash
# Für Windows-PC am Ausschank:
GOOS=windows GOARCH=amd64 go build -o jotti-relay.exe ./cmd/relay/

# Für Raspberry Pi (Linux ARM64):
GOOS=linux GOARCH=arm64 go build -o jotti-relay ./cmd/relay/

# Starten:
./jotti-relay \
  --backend="https://jotti.meinverein.de" \
  --token="<RELAY_AUTH_TOKEN>" \
  --poll=2
```

Das Binary ist vollständig portabel — keine Runtime, keine Konfigurationsdatei nötig. Es kann beim nächsten Vereinsfest direkt wiederverwendet werden (nur `--backend` und `--token` müssen aktualisiert werden).

---

## 6. ESC/POS Bon-Format

### 6.1 Bon-Struktur

```
┌────────────────────────────────────────────────┐ ← 48 Zeichen (Font A, 80mm)
│              VEREINSFEST 2026                  │ fett, zentriert
│================================================│ 48 ×  '='
│                  Tisch 7                       │ doppelte Größe, zentriert
│  02.08.2026 19:34      Bedienung: Maria        │ normal
│------------------------------------------------│ 48 × '-'
│  2 x Hefeweizen (0,5l)           5.00 EUR     │ linksbündig + rechtsbündig Preis
│  1 x Bratwurst (mit Brot)        2.50 EUR     │
│------------------------------------------------│
│                            GESAMT:  7.50 EUR   │ doppelte Höhe, rechtsbündig
│                                                │
│ HINWEIS:                                       │ (nur wenn Kommentar vorhanden)
│ Bitte scharf für Tisch 7                       │
│                                                │
│           Vielen Dank!                         │ zentriert
│                                                │
│                                                │ ← 5 Leerzeilen vor Cut
└────────────────────────────────────────────────┘
      ✂ (Partial Cut)
```

### 6.2 Go-Implementierung

```go
// backend/api/table/application/escpos/formatter.go
package escpos

import (
    "bytes"
    "fmt"
    "strings"
    "time"

    "github.com/nicograef/jotti/backend/domain/table"
)

const lineWidth = 48 // Font A, 12×24 Dots bei 576 dots/line → 48 Zeichen

// FormatBestellungBon generiert den ESC/POS-Byte-Payload für eine Kategorie.
// Es werden nur die Positionen der übergebenen Kategorie gedruckt.
func FormatBestellungBon(
    positionen []table.Position,
    tischName string,
    userName string,
    aufgenommenAm time.Time,
    kommentar string,
    withBeep bool,
) []byte {
    var buf bytes.Buffer

    if withBeep {
        buf.WriteString(Beep)
    }

    buf.WriteString(Init)
    buf.WriteString(AlignCenter)
    buf.WriteString(BoldOn)
    buf.WriteString("VEREINSFEST 2026\n")
    buf.WriteString(strings.Repeat("=", lineWidth) + "\n")
    buf.WriteString(BoldOff)

    // Tischnummer groß
    buf.WriteString(TextDoubleAll)
    buf.WriteString(fmt.Sprintf("%s\n", tischName))
    buf.WriteString(TextNormal)

    buf.WriteString(fmt.Sprintf(
        "%-24s%24s\n",
        aufgenommenAm.Format("02.01.2006 15:04"),
        "Bedienung: "+truncate(userName, 14),
    ))
    buf.WriteString(strings.Repeat("-", lineWidth) + "\n")

    // Positionen
    buf.WriteString(AlignLeft)
    gesamtCents := 0

    for _, pos := range positionen {
        gesamtCents += pos.Einzelpreis * pos.Menge

        // Artikelzeile: "2 x Hefeweizen (0,5l)"
        artikel := fmt.Sprintf("%d x %s (%s)", pos.Menge, pos.ProduktName, pos.VarianteName)
        // Preisanzeige ohne float: Integer-Arithmetik (Q-04 konform)
        preis := formatCents(pos.Einzelpreis * pos.Menge)

        // Padding damit Preis rechtsbündig steht
        padding := lineWidth - len(artikel) - len(preis)
        if padding < 1 {
            artikel = truncate(artikel, lineWidth-len(preis)-1)
            padding = 1
        }
        buf.WriteString(artikel + strings.Repeat(" ", padding) + preis + "\n")
    }

    buf.WriteString(strings.Repeat("-", lineWidth) + "\n")

    // Gesamtpreis rechtsbündig, doppelte Höhe
    gesamt := "GESAMT: " + formatCents(gesamtCents)
    buf.WriteString(TextDoubleHigh)
    buf.WriteString(AlignRight)
    buf.WriteString(gesamt + "\n")
    buf.WriteString(TextNormal)
    buf.WriteString(AlignLeft)

    // Kommentar (optional)
    if kommentar != "" {
        buf.WriteString("\n" + BoldOn + "HINWEIS:\n" + BoldOff)
        buf.WriteString(wrapLine(kommentar, lineWidth) + "\n")
    }

    // Abschluss
    buf.WriteString(AlignCenter)
    buf.WriteString("\n    Vielen Dank!    \n")
    buf.WriteString(strings.Repeat("\n", 5)) // 5 Leerzeilen vor dem Schnitt
    buf.WriteString(CutPaper)

    return buf.Bytes()
}

// formatCents formatiert Cent-Beträge als "X.YY EUR" ohne float-Arithmetik.
// Beispiel: 550 → "5.50 EUR"
func formatCents(cents int) string {
    return fmt.Sprintf("%d.%02d EUR", cents/100, cents%100)
}

// truncate kürzt einen String auf maxLen Zeichen.
func truncate(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    return s[:maxLen-1] + "…"
}

// wrapLine bricht einen langen String an Wortgrenzen um.
func wrapLine(s string, width int) string {
    if len(s) <= width {
        return s
    }
    var result strings.Builder
    words := strings.Fields(s)
    line := ""
    for _, w := range words {
        if len(line)+1+len(w) > width && line != "" {
            result.WriteString(line + "\n")
            line = w
        } else {
            if line == "" {
                line = w
            } else {
                line += " " + w
            }
        }
    }
    if line != "" {
        result.WriteString(line)
    }
    return result.String()
}
```

**Wichtig:** `formatCents` arbeitet ausschließlich mit Integer-Arithmetik (`cents/100` und `cents%100`), um die jotti-Regel „keine Floats für Geldbeträge" auch bei der Anzeige konsequent einzuhalten.

### 6.3 Bon-Aufteiler (Kategorie-Routing)

```go
// backend/api/table/application/print.go

// CreatePrintJobsForBestellung teilt die Positionen nach Kategorie auf
// und erstellt für jede Kategorie einen ESC/POS-Payload.
func CreatePrintJobsForBestellung(
    ctx context.Context,
    bestellung table.Bestellung,
    tischName string,
    druckerConfig map[string]string, // kategorie → drucker_ip
) []PrintJobDraft {
    // Positionen nach Kategorie gruppieren
    byKategorie := map[string][]table.Position{}
    for _, pos := range bestellung.Positionen {
        byKategorie[pos.Kategorie] = append(byKategorie[pos.Kategorie], pos)
    }

    var jobs []PrintJobDraft
    for kategorie, positionen := range byKategorie {
        ip, ok := druckerConfig[kategorie]
        if !ok || ip == "" {
            continue // Kein Drucker für diese Kategorie konfiguriert
        }

        withBeep := kategorie == "essen" // Küche: Piepser an
        payload := escpos.FormatBestellungBon(
            positionen,
            tischName,
            bestellung.UserName,
            bestellung.AufgenommenAm,
            bestellung.Kommentar,
            withBeep,
        )

        jobs = append(jobs, PrintJobDraft{
            Kategorie: kategorie,
            DruckerIP: ip,
            Payload:   base64.StdEncoding.EncodeToString(payload),
        })
    }
    return jobs
}
```

---

## 7. Kategorie-Drucker-Konfiguration

Drucker werden pro Kategorie konfiguriert. Der Admin trägt im Frontend die IP-Adresse des jeweiligen Druckers ein.

**Backend-Endpunkte (Admin):**

- `POST /admin/get-drucker-config` → gibt aktuelle Konfiguration zurück
- `POST /admin/update-drucker-config` → speichert neue IP per UPSERT

**Validierung:**
- Backend (zog): IPv4-Regex oder leer (leer = kein Drucker für diese Kategorie)
- Frontend (Zod): Gleiche Regel, live beim Eingeben

**Verhalten bei unkonfiguriertem Drucker:**
- Ist `drucker_ip` leer oder nicht gesetzt, wird für diese Kategorie kein `print_job` erstellt
- Keine Fehlermeldung an die Servicekraft — Bondruck ist ein "Best Effort"-Feature im laufenden Betrieb
- Der Admin sieht im Druckerkonfigurationsbereich den Status

**Datenmodell:**

```
┌─────────────────────────────────────────────┐
│             kategorie_drucker               │
├─────────────┬──────────────┬────────────────┤
│ kategorie   │ drucker_ip   │ updated_at     │
├─────────────┼──────────────┼────────────────┤
│ essen       │ 192.168.1.51 │ 2026-08-02 ... │
│ getraenk    │ 192.168.1.50 │ 2026-08-02 ... │
│ sonstiges   │              │ 2026-08-02 ... │
└─────────────┴──────────────┴────────────────┘
```

---

## 8. Verworfene Alternativen

### 8.1 WebSocket statt HTTP-Polling

**bondruck.md empfiehlt WebSocket (gorilla/websocket).**

Gründe dagegen:
- **Architektur-Mismatch:** jotti verwendet ausschließlich POST-Endpunkte; WebSocket erfordert einen HTTP-Upgrade-Handler und bricht dieses Prinzip auf.
- **Komplexität:** WebSocket-Hub mit `sync.Mutex`, Goroutinen, Ping/Pong, Reconnect-Logik im Backend sind erheblich mehr Code als zwei POST-Endpunkte.
- **gorilla/websocket ist seit 2022 archiviert** und wird nicht mehr aktiv gepflegt. Alternativen (nhooyr.io/websocket, gobwas/ws) haben keinen klaren Community-Konsens.
- **Latenz ist nicht kritisch:** Für Küche und Theke ist ein 2-Sekunden-Versatz beim Bondruck vollkommen akzeptabel.

HTTP-Polling mit 2-Sekunden-Intervall ist zuverlässiger, einfacher zu debuggen und vollständig jotti-konform.

### 8.2 VPN (WireGuard / Tailscale) statt Print-Relay

Würde eine direkte TCP-Verbindung vom Cloud-Backend zum Drucker ermöglichen. Abgelehnt, weil:
- Hoher Setup-Aufwand für ehrenamtliche Helfer
- Nicht für wechselnde Netze (LTE) geeignet
- Erfordert Zugriff auf den Router des Vereinsfests

### 8.3 Browser-Druck (Print CSS / HTML)

Servicekraft druckt manuell aus dem Browser via Ctrl+P. Abgelehnt, weil:
- Nicht automatisch (K-12 fordert automatischen Druck)
- Schlechte Kontrolle über Layout und Papiervorschub
- Kein ESC/POS-Zugriff aus dem Browser möglich

### 8.4 Relay mit externer Datenbank (SQLite/bbolt) statt JSON-State

bondruck.md schlägt SQLite oder bbolt für die lokale Idempotenzliste vor. Für maximal ~2.000 Jobs pro Vereinsfest ist eine einfache JSON-Datei vollkommen ausreichend, deutlich einfacher und erfordert keine zusätzliche Dependency.

### 8.5 DELIVERED-Status in print_jobs

bondruck.md sieht drei Status vor: `QUEUED → DELIVERED → ACKNOWLEDGED`. Bei HTTP-Polling ist `DELIVERED` nicht sinnvoll: Es gibt keinen Moment, an dem das Backend "zugestellt hat" ohne das der Relay-Client gleichzeitig die Kontrolle übernommen hat. Die lokale Idempotenzliste des Relay-Clients übernimmt diese Funktion vollständig. Zwei Status (`QUEUED` und `ACKNOWLEDGED`) reichen.

### 8.6 Float-Formatierung für Centbeträge

bondruck.md nutzt `float64(cents) / 100.0` für die Anzeige. Obwohl dies technisch korrekt wäre (display-only), widerspricht es dem Spirit der jotti-Regel. `fmt.Sprintf("%d.%02d EUR", cents/100, cents%100)` ist äquivalent und ausnahmslos integer-basiert.
