# Implementierungsplan: Bondruck (K-12)

## Überblick

Bondruck ermöglicht das automatische Drucken von Bestell-Bons an Ausgabestationen (Küche, Getränketheke) beim Aufnehmen einer Bestellung. Die Architektur besteht aus zwei Komponenten:

1. **Cloud-Backend**: Neuer `POST /relay/poll`-Endpunkt, der Bestellungs-Events als Druck-Aufträge bereitstellt (Cursor-basiertes Event-Polling). ESC/POS-Payloads werden on-the-fly generiert, Drucker-IPs zur Lesezeit per JOIN aufgelöst.
2. **Print-Relay**: Eigenständiges Go-Binary im lokalen Netzwerk des Vereinsfests. Pollt das Backend, druckt ESC/POS-Bytes an Bondrucker via TCP:9100.

**Zentrale Designentscheidungen:**

- Kein Outbox-Pattern, keine `print_jobs`-Tabelle — die `events`-Tabelle IST die Outbox
- `WriteEvent()` bleibt unverändert — vollständige Isolation vom Kassenbetrieb
- Cursor-basiertes Polling mit Per-Event-Cursor-Inkrement und lokaler Idempotenzliste für Exactly-Once Delivery
- Statischer `RELAY_AUTH_TOKEN` statt JWT (das Relay ist kein Benutzer)
- Ein einziger Endpunkt (`POST /relay/poll`) — der Cursor selbst ist die Bestätigung

**ADR:** [docs/adr/bondruck.md](../adr/bondruck.md)

**Systemarchitektur:**

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Cloud-VPS (jotti)                             │
│  ┌───────────────┐   ┌─────────────────┐   ┌─────────────────────┐ │
│  │   Frontend    │──▶│  Go Backend     │──▶│    PostgreSQL       │ │
│  │  (Smartphone) │   │  POST /service/ │   │  events (append)   │ │
│  └───────────────┘   │  bestellung-    │   │  table_state       │ │
│                      │  aufnehmen      │   │  kategorie_drucker │ │
│                      │  POST /relay/   │   └─────────────────────┘ │
│                      │  poll           │ ← einziger Relay-Endpunkt │
│                      └────────┬────────┘                            │
└───────────────────────────────┼─────────────────────────────────────┘
                                │ HTTPS (ausgehend vom Festzelt → NAT)
┌───────────────────────────────┼─────────────────────────────────────┐
│              Vereinsfest (lokales LAN)                               │
│                      ┌────────▼────────┐                            │
│                      │  Print-Relay    │                            │
│                      │  (Go Binary)    │                            │
│                      └────────┬────────┘                            │
│              ┌────────────────┴──────────────────┐                  │
│     ┌────────▼──────┐                  ┌─────────▼─────┐           │
│     │  MUNBYN       │                  │  MUNBYN       │           │
│     │ ITPP047P      │                  │ ITPP047P      │           │
│     │ Getränke      │                  │ Küche/Essen   │           │
│     │ 192.168.1.50  │                  │ 192.168.1.51  │           │
│     │ TCP:9100      │                  │ TCP:9100      │           │
│     └───────────────┘                  └───────────────┘           │
└──────────────────────────────────────────────────────────────────────┘
```

---

## Phase 1: Datenbank + Repository

### 1.1 · Schema: `kategorie_drucker`-Tabelle

- [ ] `kategorie_drucker`-Tabelle in `database/migrations/01_initial.up.sql` ergänzen
- [ ] Drei Kategorien vorbelegen: `essen`, `getraenk`, `sonstiges`

**Kontext:**

- `database/migrations/01_initial.up.sql` — am Ende ergänzen
- `docs/adr/bondruck.md` — Schema-Entscheidungen
- Bestehende Tabellen im selben Migrationsskript als Referenz für Stil/Konventionen

**Schema:**

```sql
CREATE TABLE kategorie_drucker (
    kategorie   ProduktKategorie PRIMARY KEY,
    drucker_ip  VARCHAR(50) NOT NULL DEFAULT '',
    bonmodus    TEXT NOT NULL DEFAULT 'pro_position'
                CHECK (bonmodus IN ('pro_position', 'pro_bestellung')),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO kategorie_drucker (kategorie) VALUES ('essen'), ('getraenk'), ('sonstiges');
```

### 1.2 · sqlc-Queries für Relay und Admin

- [ ] Neue Query-Datei `backend/sqlc/queries/relay.sql` erstellen
- [ ] Neue Query-Datei `backend/sqlc/queries/drucker.sql` erstellen (Admin-Queries)
- [ ] `make sqlc` ausführen und generierten Code prüfen

**Kontext:**

- `backend/sqlc/queries/` — bestehende Query-Dateien als Referenz für Stil
- `backend/sqlc.yaml` — sqlc-Konfiguration

**Queries (relay.sql):**

```sql
-- name: GetBestellungEventsSinceCursor :many
SELECT id, user_name, subject, data, timestamp
FROM events
WHERE type = 'tisch.bestellung-aufgenommen:v1'
  AND id > $1
ORDER BY id ASC
LIMIT 50;
```

**Queries (drucker.sql):**

```sql
-- name: GetKategorieDrucker :many
SELECT kategorie, drucker_ip, bonmodus
FROM kategorie_drucker;

-- name: GetKonfigurierteKategorieDrucker :many
SELECT kategorie, drucker_ip, bonmodus
FROM kategorie_drucker
WHERE drucker_ip != '';

-- name: UpsertKategorieDrucker :exec
INSERT INTO kategorie_drucker (kategorie, drucker_ip, bonmodus, updated_at)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (kategorie) DO UPDATE SET
    drucker_ip = EXCLUDED.drucker_ip,
    bonmodus = EXCLUDED.bonmodus,
    updated_at = NOW();
```

### 1.3 · Repository: Drucker-Konfiguration

- [ ] `backend/repository/drucker_repo/repo.go` erstellen
- [ ] Interface-Definition und Implementierung basierend auf sqlc-generierten Queries
- [ ] Unit-Tests für Repository-Methoden

**Kontext:**

- `backend/repository/event_repo/` — als Referenz für Repository-Struktur
- `backend/sqlc/dbgen/` — generierte Query-Funktionen (nach `make sqlc`)

**Implementierungsvorlage `backend/repository/drucker_repo/repo.go`:**

```go
package drucker_repo

import (
    "context"
    "database/sql"

    "github.com/nicograef/jotti/backend/sqlc/dbgen"
)

type DruckerKonfig struct {
    Kategorie string
    DruckerIP string
    Bonmodus  string // "pro_position" | "pro_bestellung"
}

type Repository struct {
    q *dbgen.Queries
}

func NewRepository(db *sql.DB) Repository {
    return Repository{q: dbgen.New(db)}
}

// GetAlleKategorieDrucker gibt die Konfiguration aller drei Kategorien zurück
// (auch unkonfigurierte, mit leerem drucker_ip).
func (r Repository) GetAlleKategorieDrucker(ctx context.Context) ([]DruckerKonfig, error) {
    rows, err := r.q.GetKategorieDrucker(ctx)
    if err != nil {
        return nil, err
    }
    result := make([]DruckerKonfig, 0, len(rows))
    for _, row := range rows {
        result = append(result, DruckerKonfig{
            Kategorie: string(row.Kategorie),
            DruckerIP: row.DruckerIp,
            Bonmodus:  row.Bonmodus,
        })
    }
    return result, nil
}

// GetKonfigurierteKategorieDrucker gibt nur Kategorien mit konfiguriertem Drucker zurück.
// Wird vom Relay-Service verwendet.
func (r Repository) GetKonfigurierteKategorieDrucker(ctx context.Context) (map[string]DruckerKonfig, error) {
    rows, err := r.q.GetKonfigurierteKategorieDrucker(ctx)
    if err != nil {
        return nil, err
    }
    result := make(map[string]DruckerKonfig, len(rows))
    for _, row := range rows {
        result[string(row.Kategorie)] = DruckerKonfig{
            Kategorie: string(row.Kategorie),
            DruckerIP: row.DruckerIp,
            Bonmodus:  row.Bonmodus,
        }
    }
    return result, nil
}

// UpsertKategorieDrucker speichert die Drucker-IP und den Bonmodus für eine Kategorie.
func (r Repository) UpsertKategorieDrucker(ctx context.Context, kategorie, druckerIP, bonmodus string) error {
    return r.q.UpsertKategorieDrucker(ctx, dbgen.UpsertKategorieDruckerParams{
        Kategorie: dbgen.ProduktKategorie(kategorie),
        DruckerIp: druckerIP,
        Bonmodus:  bonmodus,
    })
}
```

### 1.4 · Repository: Event-Repo um GetBestellungEventsSinceCursor erweitern

- [ ] Neue Methode `GetBestellungEventsSinceCursor` in `backend/repository/event_repo/repo.go` ergänzen

**Implementierungsvorlage (an bestehendes `event_repo/repo.go` anfügen):**

```go
// GetBestellungEventsSinceCursor liest BestellungAufgenommenV1-Events ab dem
// angegebenen Cursor (exklusive). Wird vom Relay-Service für das Bondruck-
// Polling verwendet.
func (r Repository) GetBestellungEventsSinceCursor(ctx context.Context, cursor int) ([]event.Event, error) {
    rows, err := r.q.GetBestellungEventsSinceCursor(ctx, int64(cursor))
    if err != nil {
        return nil, err
    }
    events := make([]event.Event, 0, len(rows))
    for _, row := range rows {
        events = append(events, event.Event{
            ID:       int(row.ID),
            UserName: row.UserName,
            Subject:  row.Subject,
            Data:     row.Data,
            Time:     row.Timestamp,
        })
    }
    return events, nil
}
```

---

## Phase 2: ESC/POS-Formatter (Domain-nah, reine Funktionen)

### 2.1 · ESC/POS-Konstanten und Bon-Formatierung

- [ ] Package `backend/api/relay/application/escpos/` erstellen
- [ ] `constants.go` mit allen ESC/POS-Befehlen
- [ ] `formatter.go` mit `FormatPositionBon()` und `FormatSammelBon()`
- [ ] Unit-Tests für Formatter-Funktionen

**Kontext:**

- Hardware: MUNBYN ITPP047P-UE, 48 Zeichen pro Zeile (Font A, 80mm), ESC/POS via TCP:9100
- Hinweis Auto-Cutter: Messer sitzt mechanisch ~3mm über dem Druckkopf → **5 Leerzeilen vor `CutPaper`** obligatorisch

**`backend/api/relay/application/escpos/constants.go`:**

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

// Hardware-Statusabfrage (wird im Relay verwendet, nicht im Backend)
const StatusPaper = "\x10\x04\x04" // DLE EOT 4 — liefert 1 Byte zurück
// Antwortbyte: Bit 5 (0x20) = Papier fast leer, Bit 6 (0x40) = Papier leer
// Drucker "bereit" wenn: (antwort & 0x60) == 0
```

**Bon-Struktur (pro Position — Standard):**

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
│  19:34  Bedienung: Maria                       │ normal, links
│                                                │
│                                                │ ← 5 Leerzeilen vor Cut
└────────────────────────────────────────────────┘
      ✂ (Partial Cut)
```

**Bon-Struktur (pro Bestellung — optionaler Modus):**

```
┌────────────────────────────────────────────────┐
│                                                │
│               ══ Tisch 7 ══                    │ doppelte Größe, fett, zentriert
│                                                │
│  3x Pommes (groß)                              │ doppelte Höhe, fett, links
│  1x Bratwurst (mit Brot)                       │ doppelte Höhe, fett, links
│                                                │
│  ohne Ketchup für die Pommes                   │ normal, fett (Kommentar)
│                                                │
│------------------------------------------------│
│  19:34  Bedienung: Maria                       │
│                                                │ ← 5 Leerzeilen vor Cut
└────────────────────────────────────────────────┘
      ✂ (Partial Cut)
```

**Designentscheidungen:**

- Kein Gesamtpreis — die Küche braucht keine Preise, nur Arbeitsaufträge
- Kein Header — spart Papier und Lesezeit
- Tisch dominant — doppelte Größe, fett, sofort erkennbar beim Aufhängen
- Kommentar prominent — falls vorhanden, fett und direkt unter der Position
- Metadaten unten — Zeitstempel und Servicekraft sind sekundäre Information

**`backend/api/relay/application/escpos/formatter.go`:**

```go
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

    // Tisch — groß und fett, zentriert
    buf.WriteString(AlignCenter)
    buf.WriteString(TextDoubleAll)
    buf.WriteString(BoldOn)
    buf.WriteString(fmt.Sprintf("== %s ==\n", tischName))
    buf.WriteString(BoldOff)
    buf.WriteString(TextNormal)
    buf.WriteString("\n")

    // Position — doppelte Höhe, fett, zentriert
    buf.WriteString(TextDoubleHigh)
    buf.WriteString(BoldOn)
    artikel := fmt.Sprintf("%dx %s (%s)", pos.Menge, pos.ProduktName, pos.VarianteName)
    buf.WriteString(artikel + "\n")
    buf.WriteString(BoldOff)
    buf.WriteString(TextNormal)

    // Kommentar (optional) — fett, linksbündig
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

    // 5 Leerzeilen vor dem Schnitt (Messer sitzt ~3mm über dem Druckkopf)
    buf.WriteString(strings.Repeat("\n", 5))
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

    // Tisch — groß und fett, zentriert
    buf.WriteString(AlignCenter)
    buf.WriteString(TextDoubleAll)
    buf.WriteString(BoldOn)
    buf.WriteString(fmt.Sprintf("== %s ==\n", tischName))
    buf.WriteString(BoldOff)
    buf.WriteString(TextNormal)
    buf.WriteString("\n")

    // Positionen — doppelte Höhe, fett, linksbündig
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

    // 5 Leerzeilen vor dem Schnitt
    buf.WriteString(strings.Repeat("\n", 5))
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

---

## Phase 3: Backend-Relay-Endpunkt

### 3.1 · Application-Schicht: Relay-Query-Service

- [ ] Package `backend/api/relay/application/` erstellen
- [ ] `query.go` mit `Query`-Struct, Interfaces und `GetDruckAuftraege(ctx, lastEventID)`
- [ ] `print.go` mit `DruckerKonfig`, `createDruckAuftraegeFromEvent()` und `parseTischName()`
- [ ] Unit-Tests für Query-Service und Druck-Auftrags-Generator

**Kontext:**

- `backend/api/table/application/` — als Referenz für Application-Service-Struktur
- `backend/domain/table/bestellung.go` — `Position`-Struct
- `backend/domain/event/event.go` — `Event`-Struct
- `backend/repository/event_repo/` — Muster für Repository-Interface

**`backend/api/relay/application/query.go`:**

```go
package application

import (
    "context"

    "github.com/rs/zerolog"

    "github.com/nicograef/jotti/backend/domain/event"
)

type Query struct {
    EventRepo   eventRepo
    DruckerRepo druckerRepo
}

type eventRepo interface {
    GetBestellungEventsSinceCursor(ctx context.Context, cursor int) ([]event.Event, error)
}

type druckerRepo interface {
    // Gibt nur Kategorien zurück, für die eine drucker_ip konfiguriert ist.
    GetKonfigurierteKategorieDrucker(ctx context.Context) (map[string]DruckerKonfig, error)
}

// DruckAuftrag ist das Application-DTO, das an den HTTP-Handler weitergegeben wird.
type DruckAuftrag struct {
    EventID   int    // Für Cursor-Tracking
    DruckerIP string // Zur Lesezeit aufgelöst (immer aktuell)
    Payload   string // Base64-kodierter ESC/POS-Byte-String
}

// GetDruckAuftraege liest neue BestellungAufgenommenV1-Events seit dem Cursor
// und erzeugt daraus Druck-Aufträge (1 pro Position oder 1 pro Bestellung je nach Bonmodus).
// Gibt nil, nil zurück wenn keine neuen Events vorhanden.
func (q Query) GetDruckAuftraege(ctx context.Context, lastEventID int) ([]DruckAuftrag, error) {
    log := zerolog.Ctx(ctx)

    // 1. Neue BestellungAufgenommenV1-Events lesen
    events, err := q.EventRepo.GetBestellungEventsSinceCursor(ctx, lastEventID)
    if err != nil {
        return nil, err
    }

    if len(events) == 0 {
        return nil, nil
    }

    // 2. Druckerkonfiguration zur Lesezeit holen (immer aktuell)
    druckerConfig, err := q.DruckerRepo.GetKonfigurierteKategorieDrucker(ctx)
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

**`backend/api/relay/application/print.go`:**

```go
package application

import (
    "encoding/base64"
    "encoding/json"
    "strings"

    "github.com/nicograef/jotti/backend/api/relay/application/escpos"
    "github.com/nicograef/jotti/backend/domain/event"
    "github.com/nicograef/jotti/backend/domain/table"
)

// DruckerKonfig enthält die Konfiguration eines Druckers für eine Kategorie.
type DruckerKonfig struct {
    IP       string // z.B. "192.168.1.51", leer = kein Drucker
    Bonmodus string // "pro_position" (Standard) oder "pro_bestellung"
}

// bestellungEventData spiegelt die relevanten Felder von BestellungAufgenommenV1.
// Kein Schema-Validierung nötig — die Daten wurden beim Schreiben bereits validiert.
type bestellungEventData struct {
    Positionen []table.Position `json:"positionen"`
    Kommentar  string           `json:"kommentar"`
}

// createDruckAuftraegeFromEvent erzeugt Druck-Aufträge aus einem BestellungAufgenommen-Event.
//   - pro_position (Standard): 1 Bon pro Position (jede Position = eigener Arbeitsauftrag)
//   - pro_bestellung: 1 Sammelbon pro Kategorie (alle Positionen einer Kategorie auf einem Bon)
//
// Kategorien ohne konfigurierte drucker_ip werden übersprungen (kein Fehler).
func createDruckAuftraegeFromEvent(
    evt event.Event,
    druckerConfig map[string]DruckerKonfig, // kategorie → DruckerKonfig
) []DruckAuftrag {

    var data bestellungEventData
    if err := json.Unmarshal(evt.Data, &data); err != nil {
        return nil
    }

    tischName := parseTischName(evt.Subject) // "tisch:7" → "Tisch 7"

    // Positionen nach Kategorie gruppieren
    byKategorie := map[string][]table.Position{}
    for _, pos := range data.Positionen {
        byKategorie[pos.Kategorie] = append(byKategorie[pos.Kategorie], pos)
    }

    var auftraege []DruckAuftrag

    for kategorie, positionen := range byKategorie {
        konfig, ok := druckerConfig[kategorie]
        if !ok || konfig.IP == "" {
            continue // Kein Drucker konfiguriert → überspringen
        }

        withBeep := kategorie == "essen" // Küche: Piepser aktivieren

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
                withBeep = false // Nur beim ersten Bon einer Kategorie piepsen
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

// parseTischName konvertiert ein Event-Subject ("tisch:7") in einen lesbaren Namen ("Tisch 7").
func parseTischName(subject string) string {
    parts := strings.SplitN(subject, ":", 2)
    if len(parts) != 2 {
        return subject
    }
    return "Tisch " + parts[1]
}
```

### 3.2 · HTTP-Handler: Relay-Poll-Endpunkt

- [ ] Package `backend/api/relay/http/` erstellen
- [ ] `handler.go` mit `Handler`-Struct und `PollHandler()` — `POST /relay/poll`
- [ ] Token-Prüfung gegen `RelayToken` aus Konfiguration (kein JWT)
- [ ] Request/Response-DTOs mit `json`-Tags

**Kontext:**

- `backend/api/table/http/command_handler.go` — als Referenz für Handler-Struct-Muster
- `backend/api/helper/http.go` — `ReadBody`, `SendResponse`, `SendClientError`, `SendServerError`

**`backend/api/relay/http/handler.go`:**

```go
package http

import (
    "net/http"

    "github.com/nicograef/jotti/backend/api/helper"
    relayApp "github.com/nicograef/jotti/backend/api/relay/application"
)

type Handler struct {
    Query      relayApp.Query
    RelayToken string
}

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
        var body pollRequest
        if !helper.ReadBody(w, r, &body) {
            return
        }

        // Statischer Token-Vergleich — das Relay ist kein Benutzer, kein JWT
        if body.Token != h.RelayToken {
            helper.SendClientError(w, "unauthorized", nil)
            return
        }

        auftraege, err := h.Query.GetDruckAuftraege(r.Context(), body.LastEventID)
        if err != nil {
            helper.SendServerError(w)
            return
        }

        // Cursor = höchste verarbeitete Event-ID (oder unverändert wenn keine neuen Events)
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

### 3.3 · Routing-Integration

- [ ] `backend/api/relay.go` erstellen — `NewRelayApi()` Factory-Funktion
- [ ] In `backend/app/app.go` `SetupRoutes()` unter `/relay/` mounten — **kein** JWT-Middleware
- [ ] `RelayToken` als neue Konfigurationsvariable in `backend/config/config.go` ergänzen

**Kontext:**

- `backend/api/admin.go`, `backend/api/service.go` — Referenz für API-Factory-Muster
- `backend/app/app.go` `SetupRoutes()` — hier wird der Relay-Mount ergänzt
- `backend/config/config.go` — `Config`-Struct um `RelayToken string` erweitern

**`backend/api/relay.go`:**

```go
package api

import (
    "database/sql"
    "net/http"

    relayApp "github.com/nicograef/jotti/backend/api/relay/application"
    relayHTTP "github.com/nicograef/jotti/backend/api/relay/http"
    "github.com/nicograef/jotti/backend/repository/drucker_repo"
    "github.com/nicograef/jotti/backend/repository/event_repo"
)

func NewRelayApi(db *sql.DB, relayToken string) http.Handler {
    r := http.NewServeMux()

    eventRepo := event_repo.NewRepository(db)
    druckerRepo := drucker_repo.NewRepository(db)

    handler := relayHTTP.Handler{
        Query: relayApp.Query{
            EventRepo:   eventRepo,
            DruckerRepo: druckerRepo,
        },
        RelayToken: relayToken,
    }

    r.HandleFunc("/poll", handler.PollHandler())

    return r
}
```

**Änderungen an `backend/app/app.go` in `SetupRoutes()`:**

```go
// Relay — kein JWT, Token-Prüfung im Handler
relayApi := api.NewRelayApi(db, cfg.RelayToken)
r.Handle("/relay/", http.StripPrefix("/relay", relayApi))
```

**Änderungen an `backend/config/config.go`:**

```go
// Config struct: RelayToken ergänzen
type Config struct {
    Port       int
    Postgres   postgresConfig
    JWTSecret  string
    RelayToken string // Statischer Token für das Print-Relay
}

// In Load(): RelayToken lesen
RelayToken: parseEnvString("RELAY_AUTH_TOKEN", ""),
```

Hinweis: `RELAY_AUTH_TOKEN` kann leer sein (Bondruck nicht aktiviert). Wenn leer, lehnt der Handler jeden Request mit 401 ab (leerer Token ≠ valider Token, da Request-Token nicht leer ist). Alternativ kann im Handler explizit geprüft werden ob `RelayToken == ""`.

**`.env` ergänzen:**

```
RELAY_AUTH_TOKEN=<zufälliger Token, z.B. openssl rand -hex 32>
```

---

## Phase 4: Admin-Druckerkonfiguration

### 4.1 · Backend: Admin-Endpunkte für Druckerkonfiguration

- [ ] Package `backend/api/drucker/http/` erstellen — `handler.go` mit `CommandHandler` und `QueryHandler`
- [ ] Package `backend/api/drucker/application/` erstellen — `command.go` und `query.go`
- [ ] `POST /admin/get-drucker-config` — gibt aktuelle Konfiguration aller Kategorien zurück
- [ ] `POST /admin/update-drucker-config` — speichert Drucker-IP und Bonmodus per UPSERT
- [ ] zog-Validierung: IPv4-Regex oder leer, Bonmodus `pro_position`/`pro_bestellung`
- [ ] In `backend/api/admin.go` `NewAdminApi()` registrieren
- [ ] Unit-Tests für Validierung

**Kontext:**

- `backend/api/product/http/` und `backend/api/product/application/` — Referenz für Struktur
- `backend/api/admin.go` — hier werden die neuen Handler-Routen eingehängt
- `backend/repository/drucker_repo/repo.go` — Repository aus Phase 1.3

**Endpunkte:**

```
POST /admin/get-drucker-config
Request:  {} (kein Body nötig)
Response: {"drucker": [{"kategorie":"essen","druckerIp":"192.168.1.51","bonmodus":"pro_position"}, ...]}

POST /admin/update-drucker-config
Request:  {"kategorie": "essen", "druckerIp": "192.168.1.51", "bonmodus": "pro_position"}
Response: {} (leeres Success-Objekt)
```

**Validierung `update-drucker-config` (zog):**

```go
// IPv4-Regex oder leer (leer = Drucker deaktivieren)
var ipv4Regex = regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)

var updateDruckerSchema = z.Struct(z.Shape{
    "Kategorie": z.String().OneOf([]string{"essen", "getraenk", "sonstiges"}, ...).Required(),
    "DruckerIP": z.String().Match(ipv4Regex, z.Message("Ungültige IPv4-Adresse")).Optional(),
    "Bonmodus":  z.String().OneOf([]string{"pro_position", "pro_bestellung"}, ...).Required(),
})
```

**Einbindung in `backend/api/admin.go`:**

```go
druckerRepo := drucker_repo.NewRepository(db)
dc := druckerHTTP.CommandHandler{}
dc.Command = druckerApp.Command{DruckerRepo: druckerRepo}
dq := druckerHTTP.QueryHandler{}
dq.Query = druckerApp.Query{DruckerRepo: druckerRepo}
r.HandleFunc("/get-drucker-config", dq.GetDruckerConfigHandler())
r.HandleFunc("/update-drucker-config", dc.UpdateDruckerConfigHandler())
```

### 4.2 · Frontend: Druckerkonfiguration im Admin-Bereich

- [ ] `frontend/src/admin/DruckerConfigPage.tsx` erstellen
- [ ] Pro Kategorie: Input-Feld für IP-Adresse, Dropdown für Bonmodus
- [ ] Zod-Validierung: IPv4-Format oder leer, Bonmodus `pro_position`/`pro_bestellung`
- [ ] `frontend/src/lib/DruckerBackend.ts` — Backend-Klasse mit `BackendClient`
- [ ] Route in `frontend/src/routes.ts` ergänzen
- [ ] In Admin-Navigation (`frontend/src/admin/`) verlinken

**Kontext:**

- `frontend/src/admin/` — bestehende Admin-Seiten als Referenz für Struktur und Styling
- `frontend/src/lib/Backend.ts` — `BackendClient`-Interface (nie direkt `fetch()` verwenden)
- `frontend/src/routes.ts` — Routen-Konfiguration

**UI-Entwurf:**

- Tabelle mit 3 Zeilen (Essen, Getränke, Sonstiges)
- Pro Zeile: Kategorie-Name (readonly), IP-Adresse (Input), Bonmodus (Select)
- Speichern-Button pro Zeile (sofortiges Feedback, kein Gesamt-Submit)
- Leere IP = kein Drucker für diese Kategorie (visueller Hinweis: grauer Text „kein Drucker")
- Deutsche Labels: „Drucker-IP", „Bonmodus", „Pro Position", „Pro Bestellung"

**Zod-Schema:**

```ts
const druckerConfigSchema = z.object({
  kategorie: z.enum(["essen", "getraenk", "sonstiges"]),
  druckerIp: z
    .string()
    .regex(/^(\d{1,3}\.){3}\d{1,3}$/, "Ungültige IPv4-Adresse")
    .or(z.literal("")),
  bonmodus: z.enum(["pro_position", "pro_bestellung"]),
});
```

---

## Phase 5: Print-Relay (eigenständiges Go-Binary)

### 5.1 · Relay-Binary: Grundstruktur

- [ ] `cmd/relay/main.go` erstellen (im Workspace-Root, nicht unter `backend/`)
- [ ] CLI-Flags: `--backend`, `--token`, `--state`, `--poll`
- [ ] Poll-Loop: HTTP-POST an `/relay/poll`, Druck-Aufträge verarbeiten
- [ ] Graceful Shutdown (SIGTERM/SIGINT)

**Kontext:**

- Eigenständiges Binary — **kein** Import aus `backend/`
- Nur Go-Stdlib (`net/http`, `net`, `encoding/json`, `encoding/base64`, `flag`, `os`, `log`, `time`)
- Kein `go.mod` unter `cmd/relay/` — eigenes Modul `jotti-relay` anlegen

### 5.2 · Relay: Exactly-Once Delivery

Der Relay implementiert robustes Exactly-Once Delivery durch **zwei** Mechanismen:

**Mechanismus 1 — Per-Event Cursor-Inkrement:**
Nach jedem einzelnen erfolgreich gedruckten Auftrag wird der Cursor sofort lokal persistiert. Minimiert das Fenster für Doppeldrucke bei Absturz auf maximal einen einzelnen Auftrag.

**Mechanismus 2 — Lokale Idempotenzliste:**
Das Relay führt eine lokale Liste der zuletzt gedruckten Event-IDs (max. 2000). Vor jedem Druckvorgang wird geprüft, ob die Event-ID bereits bekannt ist. Fängt den Edge-Case ab, dass der Cursor nach erfolgreichem Druck aber vor dem Persistieren nicht gespeichert wurde.

- [ ] Per-Event Cursor-Inkrement (Cursor nach jedem erfolgreichen Druck speichern, nicht nach Batch)
- [ ] Lokale Idempotenzliste (`printed_event_ids` im State, max. 2000 Einträge, älteste rausrotieren)
- [ ] Atomare State-Datei-Schreibvorgänge (Write-to-Temp + `os.Rename`, verhindert Korruption bei Absturz)

**State-Datei `relay_state.json`:**

```json
{
  "last_event_id": 42,
  "printed_event_ids": [38, 39, 40, 41, 42]
}
```

### 5.3 · Relay: Drucker-Kommunikation

- [ ] `checkPrinter(ip string) error` — DLE EOT 4 Statusabfrage via TCP:9100
- [ ] `sendToPrinter(ip string, data []byte) error` — ESC/POS-Bytes an Drucker senden
- [ ] Retry-Schleife: Max. 60 Versuche (~5 Min.) bei Drucker nicht erreichbar, dann überspringen und loggen
- [ ] Write-Deadline auf TCP-Verbindung setzen (10s)

**Statusabfrage-Protokoll:**

- Befehl: `\x10\x04\x04` (DLE EOT 4) senden
- Antwort: 1 Byte; Bit 5 (0x20) = Papier fast leer, Bit 6 (0x40) = Papier leer
- Manche Drucker antworten nicht → `Read`-Timeout akzeptieren, kein Fehler

### 5.4 · Relay: Robustheit

- [ ] Graceful Shutdown (SIGTERM/SIGINT): laufenden Druckjob abschließen, dann State speichern und beenden
- [ ] HTTP-Statuscode vor JSON-Decode prüfen (401, 500, etc. → loggen, Retry-Backoff)
- [ ] Poll-Fehler loggen, aber weiterlaufen (kein `log.Fatal` bei transienten Netzwerkfehlern)
- [ ] Periodischer Statuslog (alle 5 Min. `„Relay aktiv, Cursor bei X, keine neuen Aufträge"`)

### 5.5 · Implementierungsvorlage `cmd/relay/main.go`

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
    "os/signal"
    "syscall"
    "time"
)

// RelayState speichert den Cursor und die Idempotenzliste.
type RelayState struct {
    LastEventID     int   `json:"last_event_id"`
    PrintedEventIDs []int `json:"printed_event_ids"` // max. 2000 Einträge
}

// DruckAuftrag ist das DTO vom jotti-Backend.
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

const maxPrintedIDs = 2000
const maxRetries    = 60 // ~5 Minuten bei 5s Retry-Intervall

func main() {
    flag.Parse()
    if *token == "" {
        log.Fatal("--token ist erforderlich")
    }

    log.Printf("jotti Print-Relay gestartet | Backend: %s | Poll: %ds", *backendURL, *pollSeconds)

    state := loadState(*stateFile)
    client := &http.Client{Timeout: 10 * time.Second}

    // Graceful Shutdown
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
            // Periodischer Statuslog (alle 5 Minuten)
            if time.Since(lastStatusLog) > 5*time.Minute {
                log.Printf("Relay aktiv, Cursor bei %d, keine neuen Aufträge", state.LastEventID)
                lastStatusLog = time.Now()
            }
        } else {
            for _, a := range auftraege {
                // Idempotenz-Check
                if idempotencySet[a.EventID] {
                    log.Printf("Event %d bereits gedruckt (Idempotenz) — überspringe", a.EventID)
                    state.LastEventID = a.EventID
                    continue
                }

                if err := printAuftragWithRetry(a); err != nil {
                    log.Printf("Druckfehler nach max. Versuchen (Event %d): %v — überspringe", a.EventID, err)
                } else {
                    log.Printf("Event %d erfolgreich gedruckt auf %s", a.EventID, a.DruckerIP)
                }

                // Per-Event Cursor-Inkrement: sofort nach jedem Auftrag speichern
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
        // 1. Hardware-Check
        if err := checkPrinter(a.DruckerIP); err != nil {
            log.Printf("Drucker %s nicht bereit (Versuch %d/%d): %v — warte 5s",
                a.DruckerIP, attempt, maxRetries, err)
            time.Sleep(5 * time.Second)
            continue
        }

        // 2. Drucken
        escpos, err := base64.StdEncoding.DecodeString(a.Payload)
        if err != nil {
            return fmt.Errorf("ungültiges Base64: %w", err)
        }
        if err := sendToPrinter(a.DruckerIP, escpos); err != nil {
            log.Printf("Sendefehler (Versuch %d/%d): %v", attempt, maxRetries, err)
            time.Sleep(5 * time.Second)
            continue
        }
        return nil
    }
    return fmt.Errorf("max. Versuche (%d) erreicht für Drucker %s", maxRetries, a.DruckerIP)
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
        return nil, fmt.Errorf("ungültiger Token (401) — Relay-Token prüfen")
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

    // DLE EOT 4: Papiersensor-Status abfragen
    if _, err := conn.Write([]byte{0x10, 0x04, 0x04}); err != nil {
        return fmt.Errorf("status-abfrage fehlgeschlagen: %w", err)
    }

    _ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
    reply := make([]byte, 1)
    if _, err := conn.Read(reply); err != nil {
        // Manche Drucker antworten nicht auf Status-Abfragen → kein Fehler
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

// --- State-Verwaltung ---

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

// saveState schreibt die State-Datei atomar (Write-to-Temp + Rename).
// Verhindert Datei-Korruption bei Absturz während des Schreibens.
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
        ids = ids[len(ids)-limit:] // Älteste entfernen
    }
    return ids
}
```

### 5.6 · Relay: Cross-Compilation und Deployment

- [ ] `go.mod` für `cmd/relay/` anlegen: `go mod init jotti-relay`
- [ ] Makefile-Target für Relay-Build ergänzen (optional)

**Build-Befehle:**

```bash
cd cmd/relay

# Windows-PC am Ausschank:
GOOS=windows GOARCH=amd64 go build -o jotti-relay.exe .

# Raspberry Pi (Linux ARM64):
GOOS=linux GOARCH=arm64 go build -o jotti-relay .

# Starten:
./jotti-relay --backend="https://jotti.meinverein.de" --token="<RELAY_AUTH_TOKEN>" --poll=2
```

---

## Phase 6: Integration + Tests

### 6.1 · Integrationstests

- [ ] Integrationstest: Bestellung aufnehmen → Relay-Endpoint pollt → erhält korrekte Druck-Aufträge
- [ ] Integrationstest: Cursor-Fortschritt (zweiter Poll mit gleichem Cursor liefert keine Duplikate)
- [ ] Integrationstest: Druckerkonfiguration ändern → nächster Poll nutzt neue IPs
- [ ] Integrationstest: Kein Drucker konfiguriert (leere `drucker_ip`) → `auftraege: []`
- [ ] Integrationstest: `401` bei falschem Token

**Kontext:**

- `backend/api/health/health_integration_test.go` — als Referenz für Integrationstest-Muster
- `test-integration.sh` — Skript das Integrationstests startet

### 6.2 · End-to-End-Validierung

- [ ] `make sqlc` — nach Query-Änderungen (sqlc-Code regenerieren)
- [ ] `make check` läuft durch (lint + unit tests)
- [ ] `make lint` läuft durch
- [ ] Manueller Test: Dev-Stack starten, Bestellung aufnehmen, `POST /relay/poll` mit `curl` testen
- [ ] `docs/bondruck.md` aktualisiert

**Manueller curl-Test:**

```bash
curl -X POST http://localhost:3000/relay/poll \
  -H "Content-Type: application/json" \
  -d '{"token":"<RELAY_AUTH_TOKEN>","lastEventId":0}'
```

---

## Abhängigkeiten

```
Phase 1 ─── unabhängig (DB-Schema + Queries + Repository)
Phase 2 ─── unabhängig (reine Funktionen, kein DB-Zugriff)
Phase 3 ─── abhängig von Phase 1 + Phase 2
Phase 4 ─── Backend abhängig von Phase 1; Frontend unabhängig von Phase 3 startbar
Phase 5 ─── abhängig von Phase 3 (braucht funktionierenden Endpunkt zum Testen)
Phase 6 ─── abhängig von allen vorherigen Phasen
```

**Parallelisierbar:** Phase 1 und Phase 2 können parallel bearbeitet werden.
**Parallelisierbar:** Phase 4 (Frontend) kann parallel zu Phase 5 (Relay) bearbeitet werden, sobald Phase 3 abgeschlossen ist.

## Datei-Übersicht (neue Dateien)

```
backend/
  api/
    relay.go                              ← NewRelayApi() Factory
    relay/
      application/
        query.go                          ← Query-Struct + GetDruckAuftraege()
        print.go                          ← DruckerKonfig, createDruckAuftraege...()
        escpos/
          constants.go                    ← ESC/POS-Befehle
          formatter.go                    ← FormatPositionBon(), FormatSammelBon()
      http/
        handler.go                        ← Handler-Struct + PollHandler()
    drucker/
      application/
        command.go                        ← UpsertKategorieDrucker
        query.go                          ← GetAlleKategorieDrucker
      http/
        handler.go                        ← CommandHandler + QueryHandler
  repository/
    drucker_repo/
      repo.go                             ← Repository + DruckerKonfig
sqlc/queries/
  relay.sql                               ← GetBestellungEventsSinceCursor
  drucker.sql                             ← Get/Upsert KategorieDrucker

frontend/src/
  admin/
    DruckerConfigPage.tsx                 ← Admin-UI für Druckerkonfiguration
  lib/
    DruckerBackend.ts                     ← Backend-Klasse

cmd/relay/
  main.go                                 ← Eigenständiges Print-Relay-Binary
  go.mod                                  ← Eigenes Modul: jotti-relay
```

**Geänderte Dateien:**

```
database/migrations/01_initial.up.sql    ← kategorie_drucker Tabelle + INSERT
backend/config/config.go                 ← RelayToken Feld + RELAY_AUTH_TOKEN
backend/app/app.go                       ← /relay/ Route mounten
backend/api/admin.go                     ← Drucker-Handler registrieren
backend/repository/event_repo/repo.go   ← GetBestellungEventsSinceCursor Methode
frontend/src/routes.ts                   ← /admin/drucker Route
```
