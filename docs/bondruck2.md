# Bondruck (K-12) — Architekturentscheidung

## Inhaltsverzeichnis

1. [Überblick](#1-überblick)
2. [Evaluation des Vorentwurfs](#2-evaluation-des-vorentwurfs)
3. [Hardware: MUNBYN ITPP047P-UE](#3-hardware-munbyn-itpp047p-ue)
4. [Systemarchitektur](#4-systemarchitektur)
5. [Cloud-Backend-Komponente](#5-cloud-backend-komponente)
6. [Print-Relay-Komponente (lokal)](#6-print-relay-komponente-lokal)
7. [ESC/POS Bon-Format](#7-escpos-bon-format)
8. [Kategorie-Drucker-Konfiguration](#8-kategorie-drucker-konfiguration)
9. [Verworfene Alternativen](#9-verworfene-alternativen)

---

## 1. Überblick

Anforderung **K-12** erfordert, dass beim Aufnehmen einer Bestellung automatisch Bons an die zuständige Ausgabestation gedruckt werden (Essen → Küchenbon, Getränke → Thekenbon). Da jotti auf einem Cloud-VPS läuft, sich der Drucker aber im lokalen Netzwerk des Vereinsfests befindet (NAT/Firewall), ist eine Zwei-Komponenten-Architektur notwendig.

**Kernentscheidungen im Überblick:**

| Aspekt | Entscheidung | Begründung |
|---|---|---|
| Netzwerk-Brücke | Print-Relay-Client (lokales Binary) | Einfachste Lösung für NAT-Problem, kein VPN-Setup nötig |
| Transport | HTTP-Polling (POST) | Passt zur POST-only-Architektur, keine zusätzlichen Protokolle |
| Datenquelle | Cursor-basiertes Event-Polling | Events sind die einzige Source of Truth — kein Outbox-Pattern, keine zusätzliche Tabelle |
| Protokoll | ESC/POS via TCP 9100 | Herstellerstandard, treiberfrei, universell |
| Bon-Aufteilung | Ein Bon pro Position (Standard) | Jede Position ist ein eigener Arbeitsauftrag für die Ausgabestation |
| Bonmodus-Option | Ein Bon pro Bestellung (Admin-Einstellung) | Für Vereine, die lieber einen Sammelbon pro Bestellung drucken |

---

## 2. Evaluation des Vorentwurfs

Der ursprüngliche Entwurf sah ein **Transactional Outbox Pattern** vor: Eine `print_jobs`-Tabelle wird in derselben Transaktion wie das Event beschrieben, das Relay pollt diese Tabelle, und ein ACK-Endpunkt markiert Jobs als erledigt. Diese Architektur wurde nach kritischer Evaluation durch einen einfacheren Ansatz ersetzt.

### 2.1 Probleme des Outbox-Ansatzes

| Problem | Beschreibung |
|---|---|
| **Kopplung an WriteEvent** | `WriteEvent()` müsste einen TxHook-Mechanismus erhalten, damit die Application-Schicht innerhalb derselben Transaktion weitere Writes ausführen kann. Das verschmutzt die Signatur der zentralen Event-Store-Methode und koppelt Infrastruktur (Drucken) an die Core Domain. |
| **Unnötige `print_jobs`-Tabelle** | Die Tabelle dupliziert Informationen, die bereits vollständig in den Fat Events vorhanden sind (Positionen, Kategorie, Tischname, Servicekraft, Zeitstempel, Kommentar). |
| **IP-Adresse beim Schreiben festgelegt** | Die Drucker-IP wird zum Zeitpunkt der Bestellung in `print_jobs.drucker_ip` gespeichert. Ändert der Admin die Druckerkonfiguration, zeigen bereits erstellte (aber noch nicht gedruckte) Jobs auf die alte IP. |
| **Pre-Rendered Payload** | Der ESC/POS-Payload wird beim Schreiben als Base64 generiert und gespeichert. Formatfehler können nicht nachträglich korrigiert werden, ohne die Tabelle zu manipulieren. |
| **Zusätzlicher ACK-Endpunkt** | Ein separater `POST /relay/ack-job` ist nötig, um Jobs als erledigt zu markieren — zusätzliche Komplexität ohne Mehrwert. |
| **Nicht isolierbar** | Durch den TxHook ist Bondruck in den Write-Pfad der Core Domain eingebettet. Das Feature kann nicht entfernt oder deaktiviert werden, ohne WriteEvent anzufassen. |

### 2.2 Verbesserter Ansatz: Cursor-basiertes Event-Polling

**Kernidee:** Die `events`-Tabelle IST bereits die Outbox. Bestellungs-Events enthalten als Fat Events alle Informationen, die ein Bon benötigt. Das Relay muss nur wissen: „Welche Bestellungen sind seit meinem letzten Poll neu?"

| Eigenschaft | Outbox-Ansatz (alt) | Cursor-Polling (neu) |
|---|---|---|
| Änderungen an `WriteEvent` | TxHook-Mechanismus nötig | Keine |
| Zusätzliche Tabellen | `print_jobs` (6+ Spalten) | Keine |
| Kopplung an Core Domain | Ja (TxHook in Transaktion) | Nein (reines Read Model) |
| Drucker-IP aufgelöst | Beim Schreiben (stale möglich) | Beim Lesen (immer aktuell) |
| ESC/POS generiert | Beim Schreiben (nicht korrigierbar) | Beim Lesen (Format jederzeit änderbar) |
| Feature abschaltbar | Nein (in Transaktion eingebettet) | Ja (Relay-Endpunkte sind komplett isoliert) |
| Doppeldruck-Schutz | `print_jobs.status` + lokale Idempotenz | Cursor-Position + lokale Idempotenz |

**Ablauf:** Das Relay sendet bei jedem Poll seinen Cursor (`lastEventId`) mit. Das Backend liest neue `BestellungAufgenommen`-Events ab diesem Cursor, löst die Drucker-IP per JOIN auf `kategorie_drucker` zur Lesezeit auf, generiert den ESC/POS-Payload on-the-fly und gibt die Druck-Aufträge zurück. Das Relay druckt, speichert den neuen Cursor lokal und sendet ihn beim nächsten Poll.

---

## 3. Hardware: MUNBYN ITPP047P-UE

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

## 4. Systemarchitektur

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Cloud-VPS (jotti)                             │
│                                                                      │
│  ┌───────────────┐   ┌─────────────────┐   ┌─────────────────────┐ │
│  │   Frontend    │   │  Go Backend     │   │    PostgreSQL       │ │
│  │  (Browser,    │──▶│  POST /service/ │──▶│  events (append)   │ │
│  │   Smartphone) │   │  bestellung-    │   │  table_state       │ │
│  │               │   │  aufnehmen      │   │  kategorie_drucker │ │
│  └───────────────┘   │                 │   └─────────────────────┘ │
│                      │  POST /relay/   │                            │
│                      │  poll           │ ← einziger Relay-Endpunkt │
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
        ▼ (bestehende Transaktion — keine Änderung)
┌───────────────────────────────────────┐
│ 1. events INSERT                      │
│    (BestellungAufgenommenV1)          │
│ 2. table_state UPSERT (Projektion)    │
└───────────────────────────────────────┘
        │ COMMIT
        ▼
Relay pollt POST /relay/poll alle ~2 Sekunden
  → sendet lastEventId (Cursor-Position)
        │
        ▼
Backend: SELECT neue Bestellungs-Events seit lastEventId
  → JOIN kategorie_drucker für Drucker-IPs
  → Generiert ESC/POS on-the-fly
  → Gibt Druck-Aufträge + neuen Cursor zurück
        │
        ▼
Relay druckt Bons → speichert neuen Cursor lokal
```

**Kernunterschied zum Vorentwurf:** Die Bestellungstransaktion bleibt unverändert. Es gibt kein Outbox-Pattern, keine `print_jobs`-Tabelle und keinen ACK-Endpunkt. Das Backend erzeugt Druck-Aufträge dynamisch als Read Model beim Polling.

---

## 5. Cloud-Backend-Komponente

### 5.1 Datenbankschema

Es wird **keine `print_jobs`-Tabelle** benötigt. Die bestehende `events`-Tabelle enthält als Fat Events bereits alle Informationen für den Bondruck. Es wird lediglich eine Konfigurationstabelle für die Drucker-Zuordnung ergänzt:

```sql
-- In database/migrations/01_initial.up.sql ergänzen

CREATE TABLE kategorie_drucker (
    kategorie   ProduktKategorie PRIMARY KEY,
    drucker_ip  VARCHAR(50) NOT NULL DEFAULT '',
    bonmodus    TEXT NOT NULL DEFAULT 'pro_position'
                CHECK (bonmodus IN ('pro_position', 'pro_bestellung')),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO kategorie_drucker (kategorie) VALUES ('essen'), ('getraenk'), ('sonstiges');
```

Die `events`-Tabelle hat bereits den benötigten Index (`idx_events_type` auf `type`) und auto-increment `id` für effiziente Cursor-Abfragen.

**Warum keine `print_jobs`-Tabelle:**
- Die `events`-Tabelle IST bereits die Outbox — `BestellungAufgenommenV1`-Events enthalten alle Daten, die ein Bon braucht (Positionen, Kategorie, Tischname, Servicekraft, Zeitstempel, Kommentar).
- Drucker-IPs werden beim Lesen per JOIN aufgelöst → Konfigurationsänderungen wirken sofort.
- ESC/POS-Payloads werden on-the-fly generiert → Formatänderungen wirken sofort.

### 5.2 Keine Änderungen an WriteEvent

`WriteEvent()` im `event_repo` bleibt **vollständig unverändert**. Es gibt keinen TxHook, keinen zusätzlichen Parameter und keine Kopplung zum Bondruck. Der Bestellungs-Write-Pfad kennt keinen Bondruck.

```go
// backend/repository/event_repo/repo.go — UNVERÄNDERT
func (r Repository) WriteEvent(ctx context.Context, e event.Event) (int, error) {
    // bestehende Logik: BEGIN TX → INSERT event → APPLY to state → UPSERT projection → COMMIT
    // Keine Änderung nötig.
}
```

### 5.3 Relay-Query: Neue Events seit Cursor

Eine neue Query liest `BestellungAufgenommenV1`-Events ab einem gegebenen Cursor:

```sql
-- backend/sqlc/queries/relay.sql

-- name: GetBestellungEventsSinceCursor :many
SELECT id, user_name, subject, data, timestamp
FROM events
WHERE type = 'tisch.bestellung-aufgenommen:v1'
  AND id > $1
ORDER BY id ASC
LIMIT 50;

-- name: GetKategorieDrucker :many
SELECT kategorie, drucker_ip
FROM kategorie_drucker
WHERE drucker_ip != '';

-- name: UpsertKategorieDrucker :exec
INSERT INTO kategorie_drucker (kategorie, drucker_ip, updated_at)
VALUES ($1, $2, NOW())
ON CONFLICT (kategorie) DO UPDATE SET
    drucker_ip = EXCLUDED.drucker_ip,
    updated_at = NOW();
```

### 5.4 Application-Schicht: Relay-Service

Der Relay-Service ist ein reines Read Model — er liest Events und erzeugt daraus Druck-Aufträge. Er ist vollständig isoliert vom Kassenbetrieb.

```go
// backend/api/relay/application/query.go

type Query struct {
    EventRepo   eventRepo
    DruckerRepo druckerRepo
}

type eventRepo interface {
    GetBestellungEventsSinceCursor(ctx context.Context, cursor int) ([]event.Event, error)
}

type druckerRepo interface {
    GetKategorieDrucker(ctx context.Context) (map[string]string, error) // kategorie → drucker_ip
}

// DruckAuftrag ist das DTO, das an das Relay gesendet wird.
type DruckAuftrag struct {
    EventID   int    // Für Cursor-Tracking
    DruckerIP string // Zur Lesezeit aufgelöst
    Payload   string // Base64-kodierter ESC/POS-Byte-String
}

// GetDruckAuftraege liest neue Bestellungs-Events seit dem Cursor
// und erzeugt daraus Druck-Aufträge (1 pro Position oder 1 pro Bestellung je nach Bonmodus).
func (q Query) GetDruckAuftraege(ctx context.Context, lastEventID int) ([]DruckAuftrag, error) {
    log := zerolog.Ctx(ctx)

    // 1. Neue Events lesen
    events, err := q.EventRepo.GetBestellungEventsSinceCursor(ctx, lastEventID)
    if err != nil {
        return nil, err
    }

    if len(events) == 0 {
        return nil, nil
    }

    // 2. Druckerkonfiguration lesen (zur Lesezeit — immer aktuell)
    druckerConfig, err := q.DruckerRepo.GetKategorieDrucker(ctx)
    if err != nil {
        return nil, err
    }

    // 3. Pro Event Druck-Aufträge erzeugen
    var auftraege []DruckAuftrag
    for _, evt := range events {
        jobs := createDruckAuftraegeFromEvent(evt, druckerConfig)
        auftraege = append(auftraege, jobs...)
    }

    log.Debug().Int("cursor", lastEventID).Int("new_events", len(events)).
        Int("auftraege", len(auftraege)).Msg("Relay poll")

    return auftraege, nil
}
```

### 5.5 HTTP-Endpunkt für das Relay

Es gibt **einen einzigen Endpunkt** für das Relay — kein separater ACK-Endpunkt nötig, da der Cursor selbst die Bestätigung ist.

```go
// backend/api/relay/http/handler.go

// POST /relay/poll
// Request:  {"token": "...", "lastEventId": 42}
// Response: {"auftraege": [...], "cursor": 55}
type pollRequest struct {
    Token       string `json:"token"`
    LastEventID int    `json:"lastEventId"`
}

type pollResponse struct {
    Auftraege []druckAuftragDTO `json:"auftraege"`
    Cursor    int               `json:"cursor"`
}

type druckAuftragDTO struct {
    EventID   int    `json:"eventId"`
    DruckerIP string `json:"druckerIp"`
    Payload   string `json:"payload"` // Base64 ESC/POS
}

func (h *Handler) PollHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        body := pollRequest{}
        if !helper.ReadBody(w, r, &body) {
            return
        }

        // Token-Prüfung (statischer RELAY_AUTH_TOKEN aus .env)
        if body.Token != h.RelayToken {
            helper.SendClientError(w, "unauthorized", nil)
            return
        }

        auftraege, err := h.Query.GetDruckAuftraege(r.Context(), body.LastEventID)
        if err != nil {
            helper.SendServerError(w)
            return
        }

        // Cursor = höchste Event-ID in der Response (oder lastEventId wenn keine neuen)
        cursor := body.LastEventID
        if len(auftraege) > 0 {
            cursor = auftraege[len(auftraege)-1].EventID
        }

        dtos := make([]druckAuftragDTO, 0, len(auftraege))
        for _, a := range auftraege {
            dtos = append(dtos, druckAuftragDTO{
                EventID:   a.EventID,
                DruckerIP: a.DruckerIP,
                Payload:   a.Payload,
            })
        }

        helper.SendResponse(w, pollResponse{
            Auftraege: dtos,
            Cursor:    cursor,
        })
    }
}
```

**Warum kein ACK-Endpunkt:** Der Cursor IST die Bestätigung. Beim nächsten Poll sendet das Relay den Cursor der letzten erfolgreich verarbeiteten Event-ID. Das Backend liefert nur Events ab diesem Cursor. Wenn das Relay abstürzt, startet es mit dem letzten lokal gespeicherten Cursor — nicht verarbeitete Events werden automatisch erneut geliefert.

### 5.6 Routing-Integration

Der Relay-Endpunkt wird als eigenes Modul registriert — vollständig isoliert von Admin-, Service- und Auth-Routen:

```go
// backend/api/relay.go

func NewRelayApi(relayToken string, eventRepo ..., druckerRepo ...) http.Handler {
    r := http.NewServeMux()

    query := relay_app.Query{EventRepo: eventRepo, DruckerRepo: druckerRepo}
    handler := relay_http.Handler{Query: query, RelayToken: relayToken}

    r.HandleFunc("POST /poll", handler.PollHandler())

    return r
}
```

```go
// main.go — Registrierung
mux.Handle("/relay/", http.StripPrefix("/relay", relayApi))
// Kein JWT-Middleware — der Relay-Token wird im Handler geprüft
```

**Isolation:** Der gesamte Bondruck-Code lebt in `api/relay/` und `api/relay/application/`. Er liest nur aus der bestehenden `events`-Tabelle und `kategorie_drucker`. Kein Import aus `api/table/`, kein Einfluss auf den Bestellungs-Write-Pfad.

---

## 6. Print-Relay-Komponente (lokal)

### 6.1 Design-Prinzipien

Das Relay ist ein schlankes Go-Binary ohne externe Abhängigkeiten. Es:
- ist **stateful** — es speichert nur einen einzigen Integer (Cursor-Position) und optional eine Idempotenzliste
- kennt **keine Domänenbegriffe** (kein "Tisch", kein "Bestellung") — es verarbeitet nur Bytes
- läuft **ohne Konfigurationsdatei** — alle Parameter kommen als Flags oder Umgebungsvariablen
- überlebt **Neustarts** — der Cursor ist persistent (JSON-Datei)

### 6.2 Sicherheitsschleife (At-Least-Once Delivery)

Pro empfangenem Druckauftrag:

```
Aufträge erhalten (via POST /relay/poll mit lastEventId)
    │
    ├─ Pro Auftrag:
    │   │
    │   ├─ 1. Hardware-Check: Sende DLE EOT 4 (\x10\x04\x04) an TCP 9100
    │   │      Kein TCP-Connect → warten (5s), erneut prüfen
    │   │      Drucker antwortet mit "Papier leer" → warten (5s)
    │   │      Drucker bereit → weiter
    │   │
    │   └─ 2. Drucken: Base64 dekodieren, ESC/POS-Bytes an TCP 9100 senden
    │          Erfolg → weiter zum nächsten Auftrag
    │          Fehler → Abbruch, Cursor NICHT vorrücken
    │
    └─ Alle Aufträge gedruckt → Cursor lokal speichern
       Nächster Poll mit neuem Cursor
```

**Doppeldruck-Schutz:** Da das Relay seinen Cursor erst nach erfolgreichem Druck aller Aufträge eines Polls vorrückt, ist Doppeldruck bei einem Absturz während des Druckens möglich. Das ist akzeptabel:
- Bei einem Vereinsfest ist ein gelegentlicher Doppelbon kein Problem (die Küche erkennt Duplikate am Zeitstempel).
- Falls gewünscht, kann eine optionale lokale Idempotenzliste (Event-IDs) hinzugefügt werden.

### 6.3 Implementierung (main.go)

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

// RelayState speichert den Cursor (letzte verarbeitete Event-ID)
type RelayState struct {
    LastEventID int `json:"last_event_id"`
}

// DruckAuftrag ist das DTO vom jotti-Backend
type DruckAuftrag struct {
    EventID   int    `json:"eventId"`
    DruckerIP string `json:"druckerIp"`
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
        auftraege, newCursor, err := poll(client, state.LastEventID)
        if err != nil {
            log.Printf("Fehler beim Poll: %v", err)
        } else if len(auftraege) > 0 {
            allOK := true
            for _, a := range auftraege {
                if err := printAuftrag(a); err != nil {
                    log.Printf("Druckfehler (Event %d): %v — stoppe Poll-Verarbeitung", a.EventID, err)
                    allOK = false
                    break
                }
                log.Printf("Event %d erfolgreich gedruckt auf %s", a.EventID, a.DruckerIP)
            }
            if allOK {
                state.LastEventID = newCursor
                saveState(*stateFile, state)
            }
        }
        time.Sleep(time.Duration(*pollSeconds) * time.Second)
    }
}

func printAuftrag(a DruckAuftrag) error {
    // 1. Hardware-Check (Wiederholungsschleife)
    for {
        if err := checkPrinter(a.DruckerIP); err != nil {
            log.Printf("Drucker %s nicht bereit: %v — warte 5s", a.DruckerIP, err)
            time.Sleep(5 * time.Second)
            continue
        }
        break
    }

    // 2. Drucken
    escpos, err := base64.StdEncoding.DecodeString(a.Payload)
    if err != nil {
        return fmt.Errorf("ungültiges Base64: %w", err)
    }

    return sendToPrinter(a.DruckerIP, escpos)
}

func poll(client *http.Client, lastEventID int) ([]DruckAuftrag, int, error) {
    reqBody, _ := json.Marshal(map[string]any{
        "token":       *token,
        "lastEventId": lastEventID,
    })
    resp, err := client.Post(*backendURL+"/relay/poll", "application/json", bytes.NewReader(reqBody))
    if err != nil {
        return nil, lastEventID, err
    }
    defer resp.Body.Close()

    var result struct {
        Auftraege []DruckAuftrag `json:"auftraege"`
        Cursor    int            `json:"cursor"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, lastEventID, err
    }
    return result.Auftraege, result.Cursor, nil
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
        log.Printf("WARNUNG: Relay-State konnte nicht gespeichert werden: %v", err)
    }
}
```

### 6.4 Deployment (Cross-Compilation)

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

## 7. ESC/POS Bon-Format

### 7.1 Bonmodus

**Standard: 1 Bon pro Position.** Jede Position einer Bestellung erzeugt einen eigenen Bon. Das entspricht dem Arbeitsablauf in der Küche: Jeder Bon ist ein Arbeitsauftrag.

Beispiel: Eine Bestellung mit „3x Pommes, 2x Hefeweizen, 1x Bratwurst" erzeugt:
- **Küchendrucker:** 1 Bon „3x Pommes" + 1 Bon „1x Bratwurst"
- **Getränkedrucker:** 1 Bon „2x Hefeweizen"

**Optionaler Modus: 1 Bon pro Bestellung** (Admin-Einstellung). Alle Positionen einer Kategorie werden auf einem Sammelbon zusammengefasst:
- **Küchendrucker:** 1 Bon „3x Pommes, 1x Bratwurst"
- **Getränkedrucker:** 1 Bon „2x Hefeweizen"

Die Einstellung wird in der `kategorie_drucker`-Tabelle als Spalte `bonmodus` gespeichert (siehe Schema in §5.1):

```
kategorie_drucker.bonmodus = 'pro_position'   → Standard (1 Bon pro Position)
kategorie_drucker.bonmodus = 'pro_bestellung' → Sammelbon (1 Bon pro Kategorie)
```

### 7.2 Bon-Struktur (1 Bon pro Position — Standard)

Der Bon ist auf maximale Lesbarkeit in der Küche optimiert. Auf den ersten Blick muss erkennbar sein: **Was, wie oft, für welchen Tisch, von wem, und gibt es einen Sonderwunsch?**

```
┌────────────────────────────────────────────────┐ ← 48 Zeichen (Font A, 80mm)
│                                                │
│               ══ Tisch 7 ══                    │ doppelte Größe, fett, zentriert
│                                                │
│              3x Pommes (groß)                  │ doppelte Höhe, fett, zentriert
│                                                │
│  ohne Ketchup, extra Salz                      │ normal, fett (nur wenn Kommentar)
│                                                │
│------------------------------------------------│ 48 × '-'
│  19:34  Bedienung: Maria                       │ normal, klein
│                                                │
│                                                │ ← 5 Leerzeilen vor Cut
└────────────────────────────────────────────────┘
      ✂ (Partial Cut)
```

**Designentscheidungen:**
- **Kein Gesamtpreis** — die Küche braucht keine Preise, nur Arbeitsaufträge.
- **Kein Header „VEREINSFEST"** — spart Papier und Lesezeit.
- **Tisch dominant** — doppelte Größe, fett, sofort erkennbar beim Aufhängen.
- **Position dominant** — Menge + Artikel + Variante in doppelter Höhe.
- **Kommentar prominent** — falls vorhanden, fett und direkt unter der Position.
- **Metadaten unten** — Zeitstempel und Servicekraft sind sekundäre Information.

### 7.3 Bon-Struktur (1 Bon pro Bestellung — optionaler Modus)

```
┌────────────────────────────────────────────────┐ ← 48 Zeichen (Font A, 80mm)
│                                                │
│               ══ Tisch 7 ══                    │ doppelte Größe, fett, zentriert
│                                                │
│  3x Pommes (groß)                              │ doppelte Höhe, fett
│  1x Bratwurst (mit Brot)                       │ doppelte Höhe, fett
│                                                │
│  ohne Ketchup für die Pommes                   │ normal, fett (Kommentar)
│                                                │
│------------------------------------------------│
│  19:34  Bedienung: Maria                       │ normal, klein
│                                                │
│                                                │ ← 5 Leerzeilen vor Cut
└────────────────────────────────────────────────┘
      ✂ (Partial Cut)
```

### 7.4 Go-Implementierung

```go
// backend/api/relay/application/escpos/formatter.go
package escpos

import (
    "bytes"
    "fmt"
    "strings"
    "time"

    "github.com/nicograef/jotti/backend/domain/table"
)

const lineWidth = 48 // Font A, 12×24 Dots bei 576 dots/line → 48 Zeichen

// FormatPositionBon generiert einen Bon für eine einzelne Position (Standard-Bonmodus).
func FormatPositionBon(
    pos table.Position,
    tischName string,
    userName string,
    zeitpunkt time.Time,
    kommentar string,
    withBeep bool,
) []byte {
    var buf bytes.Buffer

    if withBeep {
        buf.WriteString(Beep)
    }
    buf.WriteString(Init)

    // Tisch — groß und fett
    buf.WriteString(AlignCenter)
    buf.WriteString(TextDoubleAll)
    buf.WriteString(BoldOn)
    buf.WriteString(fmt.Sprintf("== %s ==\n", tischName))
    buf.WriteString(BoldOff)
    buf.WriteString(TextNormal)
    buf.WriteString("\n")

    // Position — doppelte Höhe, fett
    buf.WriteString(TextDoubleHigh)
    buf.WriteString(BoldOn)
    artikel := fmt.Sprintf("%dx %s (%s)", pos.Menge, pos.ProduktName, pos.VarianteName)
    buf.WriteString(artikel + "\n")
    buf.WriteString(BoldOff)
    buf.WriteString(TextNormal)

    // Kommentar (optional) — fett
    if kommentar != "" {
        buf.WriteString("\n")
        buf.WriteString(AlignLeft)
        buf.WriteString(BoldOn)
        buf.WriteString(wrapLine(kommentar, lineWidth) + "\n")
        buf.WriteString(BoldOff)
    }

    // Trennlinie + Metadaten
    buf.WriteString(AlignLeft)
    buf.WriteString(strings.Repeat("-", lineWidth) + "\n")
    buf.WriteString(fmt.Sprintf("  %s  Bedienung: %s\n",
        zeitpunkt.Format("15:04"),
        truncate(userName, 24),
    ))

    // Abschluss
    buf.WriteString(strings.Repeat("\n", 5)) // 5 Leerzeilen vor dem Schnitt
    buf.WriteString(CutPaper)

    return buf.Bytes()
}

// FormatSammelBon generiert einen Bon für alle Positionen einer Kategorie (optionaler Bonmodus).
func FormatSammelBon(
    positionen []table.Position,
    tischName string,
    userName string,
    zeitpunkt time.Time,
    kommentar string,
    withBeep bool,
) []byte {
    var buf bytes.Buffer

    if withBeep {
        buf.WriteString(Beep)
    }
    buf.WriteString(Init)

    // Tisch — groß und fett
    buf.WriteString(AlignCenter)
    buf.WriteString(TextDoubleAll)
    buf.WriteString(BoldOn)
    buf.WriteString(fmt.Sprintf("== %s ==\n", tischName))
    buf.WriteString(BoldOff)
    buf.WriteString(TextNormal)
    buf.WriteString("\n")

    // Positionen — doppelte Höhe, fett
    buf.WriteString(AlignLeft)
    buf.WriteString(TextDoubleHigh)
    buf.WriteString(BoldOn)
    for _, pos := range positionen {
        artikel := fmt.Sprintf("%dx %s (%s)", pos.Menge, pos.ProduktName, pos.VarianteName)
        buf.WriteString(artikel + "\n")
    }
    buf.WriteString(BoldOff)
    buf.WriteString(TextNormal)

    // Kommentar (optional) — fett
    if kommentar != "" {
        buf.WriteString("\n")
        buf.WriteString(BoldOn)
        buf.WriteString(wrapLine(kommentar, lineWidth) + "\n")
        buf.WriteString(BoldOff)
    }

    // Trennlinie + Metadaten
    buf.WriteString(strings.Repeat("-", lineWidth) + "\n")
    buf.WriteString(fmt.Sprintf("  %s  Bedienung: %s\n",
        zeitpunkt.Format("15:04"),
        truncate(userName, 24),
    ))

    // Abschluss
    buf.WriteString(strings.Repeat("\n", 5)) // 5 Leerzeilen vor dem Schnitt
    buf.WriteString(CutPaper)

    return buf.Bytes()
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

### 7.5 Druck-Auftrags-Generator

```go
// backend/api/relay/application/print.go

// DruckerKonfig enthält die Konfiguration eines Druckers für eine Kategorie.
type DruckerKonfig struct {
    IP       string // z.B. "192.168.1.51", leer = kein Drucker
    Bonmodus string // "pro_position" (Standard) oder "pro_bestellung"
}

// bestellungEventData spiegelt die Event-Data-Struktur von BestellungAufgenommenV1.
type bestellungEventData struct {
    Positionen []table.Position `json:"positionen"`
    Kommentar  string           `json:"kommentar"`
}

// createDruckAuftraegeFromEvent erzeugt Druck-Aufträge aus einem BestellungAufgenommen-Event.
// Im Standard-Bonmodus (pro_position) wird pro Position ein eigener Bon erzeugt.
// Im Sammel-Bonmodus (pro_bestellung) wird pro Kategorie ein Sammelbon erzeugt.
func createDruckAuftraegeFromEvent(
    evt event.Event,
    druckerConfig map[string]DruckerKonfig, // kategorie → DruckerKonfig
) []DruckAuftrag {

    data, err := event.ParseData[bestellungEventData](evt)
    if err != nil {
        return nil
    }

    tischName := parseTischName(evt.Subject) // "tisch:7" → "Tisch 7"

    var auftraege []DruckAuftrag

    // Positionen nach Kategorie gruppieren
    byKategorie := map[string][]table.Position{}
    for _, pos := range data.Positionen {
        byKategorie[pos.Kategorie] = append(byKategorie[pos.Kategorie], pos)
    }

    for kategorie, positionen := range byKategorie {
        konfig, ok := druckerConfig[kategorie]
        if !ok || konfig.IP == "" {
            continue // Kein Drucker für diese Kategorie
        }

        withBeep := kategorie == "essen" // Küche: Piepser an

        if konfig.Bonmodus == "pro_position" {
            // Standard: 1 Bon pro Position
            for _, pos := range positionen {
                payload := escpos.FormatPositionBon(
                    pos, tischName, evt.UserName, evt.Time,
                    data.Kommentar, withBeep,
                )
                auftraege = append(auftraege, DruckAuftrag{
                    EventID:   evt.ID,
                    DruckerIP: konfig.IP,
                    Payload:   base64.StdEncoding.EncodeToString(payload),
                })
                withBeep = false // Nur beim ersten Bon piepsen
            }
        } else {
            // Sammelbon: 1 Bon pro Bestellung (pro Kategorie)
            payload := escpos.FormatSammelBon(
                positionen, tischName, evt.UserName, evt.Time,
                data.Kommentar, withBeep,
            )
            auftraege = append(auftraege, DruckAuftrag{
                EventID:   evt.ID,
                DruckerIP: konfig.IP,
                Payload:   base64.StdEncoding.EncodeToString(payload),
            })
        }
    }

    return auftraege
}
```

---

## 8. Kategorie-Drucker-Konfiguration

Drucker werden pro Kategorie konfiguriert. Der Admin trägt im Frontend die IP-Adresse des jeweiligen Druckers ein und wählt den Bonmodus.

**Backend-Endpunkte (Admin):**

- `POST /admin/get-drucker-config` → gibt aktuelle Konfiguration zurück
- `POST /admin/update-drucker-config` → speichert neue Konfiguration per UPSERT

**Validierung:**
- Backend (zog): IPv4-Regex oder leer (leer = kein Drucker für diese Kategorie). Bonmodus: `pro_position` oder `pro_bestellung`.
- Frontend (Zod): Gleiche Regel, live beim Eingeben.

**Verhalten bei unkonfiguriertem Drucker:**
- Ist `drucker_ip` leer, wird für diese Kategorie kein Druck-Auftrag erzeugt.
- Keine Fehlermeldung an die Servicekraft — Bondruck ist ein "Best Effort"-Feature im laufenden Betrieb.
- Der Admin sieht im Druckerkonfigurationsbereich den Status.

**Datenmodell:**

```
┌───────────────────────────────────────────────────────────────┐
│                     kategorie_drucker                         │
├─────────────┬──────────────┬────────────────┬─────────────────┤
│ kategorie   │ drucker_ip   │ bonmodus       │ updated_at      │
├─────────────┼──────────────┼────────────────┼─────────────────┤
│ essen       │ 192.168.1.51 │ pro_position   │ 2026-08-02 ...  │
│ getraenk    │ 192.168.1.50 │ pro_position   │ 2026-08-02 ...  │
│ sonstiges   │              │ pro_position   │ 2026-08-02 ...  │
└─────────────┴──────────────┴────────────────┴─────────────────┘
```

---

## 9. Verworfene Alternativen

### 9.1 Transactional Outbox Pattern mit `print_jobs`-Tabelle

Der Vorentwurf (und die erste bondruck.md) schlagen vor, `print_jobs` in derselben Transaktion wie das Event zu schreiben. Abgelehnt, weil:

- **Unnötige Kopplung:** `WriteEvent()` müsste einen Hook-Mechanismus erhalten, der die Signatur verschmutzt und die Core Domain mit dem Bondruck-Concern koppelt.
- **Unnötige Duplikation:** Die `events`-Tabelle enthält bereits alle benötigten Daten (Fat Events). Eine zweite Tabelle dupliziert Information.
- **Stale Drucker-IPs:** Die Drucker-IP wird zum Schreibzeitpunkt festgelegt. Konfigurationsänderungen wirken erst auf zukünftige Bestellungen.
- **Stale Payloads:** Der ESC/POS-Payload wird pre-rendered. Formatfehler sind nicht korrigierbar.
- **Nicht isolierbar:** Durch den TxHook ist Bondruck in den Write-Pfad eingebettet und kann nicht ohne Code-Änderung am Event-Store entfernt werden.

Das Cursor-basierte Event-Polling nutzt die `events`-Tabelle als natürliche Outbox und vermeidet all diese Probleme.

### 9.2 WebSocket statt HTTP-Polling

Gründe dagegen:
- **Architektur-Mismatch:** jotti verwendet ausschließlich POST-Endpunkte; WebSocket erfordert einen HTTP-Upgrade-Handler.
- **Komplexität:** WebSocket-Hub mit `sync.Mutex`, Goroutinen, Ping/Pong, Reconnect-Logik.
- **gorilla/websocket ist seit 2022 archiviert** und wird nicht mehr aktiv gepflegt.
- **Latenz ist nicht kritisch:** 2-Sekunden-Versatz beim Bondruck ist für Küche und Theke vollkommen akzeptabel.

### 9.3 VPN (WireGuard / Tailscale) statt Print-Relay

Würde eine direkte TCP-Verbindung vom Cloud-Backend zum Drucker ermöglichen. Abgelehnt, weil:
- Hoher Setup-Aufwand für ehrenamtliche Helfer
- Nicht für wechselnde Netze (LTE) geeignet
- Erfordert Zugriff auf den Router des Vereinsfests

### 9.4 Browser-Druck (Print CSS / HTML)

Servicekraft druckt manuell aus dem Browser via Ctrl+P. Abgelehnt, weil:
- Nicht automatisch (K-12 fordert automatischen Druck)
- Schlechte Kontrolle über Layout und Papiervorschub
- Kein ESC/POS-Zugriff aus dem Browser möglich

### 9.5 Relay mit externer Datenbank (SQLite/bbolt)

Für einen einzigen Integer (Cursor-Position) und optional ~2.000 Event-IDs pro Vereinsfest ist eine JSON-Datei vollkommen ausreichend, deutlich einfacher und erfordert keine zusätzliche Dependency.

### 9.6 Float-Formatierung für Centbeträge

bondruck.md nutzt `float64(cents) / 100.0` für die Anzeige. Obwohl dies technisch korrekt wäre (display-only), widerspricht es dem Spirit der jotti-Regel. `fmt.Sprintf("%d.%02d EUR", cents/100, cents%100)` ist äquivalent und ausnahmslos integer-basiert. Der neue Entwurf zeigt keine Preise auf dem Bon (Küche braucht keine Preise), daher entfällt die Frage.
