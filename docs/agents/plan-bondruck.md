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
- `backend/sqlc/dbgen/` — generierte Query-Funktionen

---

## Phase 2: ESC/POS-Formatter (Domain-nah, reine Funktionen)

### 2.1 · ESC/POS-Konstanten und Bon-Formatierung

- [ ] Package `backend/api/relay/application/escpos/` erstellen
- [ ] `constants.go` mit ESC/POS-Befehlen (Init, Alignment, Bold, Cut, Beep, Schriftgrößen)
- [ ] `formatter.go` mit `FormatPositionBon()` und `FormatSammelBon()`
- [ ] `helpers.go` mit `truncate()` und `wrapLine()`
- [ ] Unit-Tests für Formatter-Funktionen

**Kontext:**
- Bon-Format-Spezifikation: siehe unten
- Hardware: MUNBYN ITPP047P-UE, 48 Zeichen pro Zeile (Font A, 80mm), ESC/POS via TCP:9100

**ESC/POS-Konstanten:**

```go
const Init           = "\x1B\x40"
const AlignLeft      = "\x1B\x61\x00"
const AlignCenter    = "\x1B\x61\x01"
const BoldOn         = "\x1B\x45\x01"
const BoldOff        = "\x1B\x45\x00"
const TextNormal     = "\x1D\x21\x00"
const TextDoubleHigh = "\x1D\x21\x01"
const TextDoubleAll  = "\x1D\x21\x11"
const CutPaper       = "\x1D\x56\x42\x00"  // Partial Cut
const Beep           = "\x1B\x42\x03\x02"  // 3 Piepser
```

**Bon-Struktur (pro Position — Standard):**

```
══ Tisch 7 ══           (doppelte Größe, fett, zentriert)

3x Pommes (groß)        (doppelte Höhe, fett, zentriert)

ohne Ketchup, extra Salz (normal, fett — nur wenn Kommentar)

------------------------------------------------
  19:34  Bedienung: Maria

                         (5 Leerzeilen vor Cut)
      ✂ Partial Cut
```

**Bon-Struktur (pro Bestellung — optionaler Modus):**

```
══ Tisch 7 ══           (doppelte Größe, fett, zentriert)

3x Pommes (groß)        (doppelte Höhe, fett)
1x Bratwurst (mit Brot) (doppelte Höhe, fett)

ohne Ketchup            (normal, fett — nur wenn Kommentar)

------------------------------------------------
  19:34  Bedienung: Maria

      ✂ Partial Cut
```

**Designentscheidungen:**
- Kein Gesamtpreis — die Küche braucht keine Preise, nur Arbeitsaufträge
- Kein Header — spart Papier und Lesezeit
- Tisch dominant — doppelte Größe, fett, sofort erkennbar beim Aufhängen
- Kommentar prominent — falls vorhanden, fett und direkt unter der Position
- Metadaten unten — Zeitstempel und Servicekraft sind sekundäre Information
- Vor dem Schnitt 5 Leerzeilen — das Messer sitzt mechanisch ~3mm über dem Druckkopf

---

## Phase 3: Backend-Relay-Endpunkt

### 3.1 · Application-Schicht: Relay-Query-Service

- [ ] Package `backend/api/relay/application/` erstellen
- [ ] `query.go` mit `GetDruckAuftraege(ctx, lastEventID)` — liest Events, löst Drucker-IPs auf, generiert ESC/POS
- [ ] `print.go` mit `createDruckAuftraegeFromEvent()` — Logik für Positions-/Sammelbon-Erzeugung
- [ ] Unit-Tests für Query-Service und Druck-Auftrags-Generator

**Kontext:**
- `backend/api/table/application/` — als Referenz für Application-Service-Struktur
- `backend/domain/table/events.go` — Event-Datenstrukturen (BestellungAufgenommenV1)
- `backend/repository/event_repo/` — Event-Repository-Interfaces

**Interfaces:**

```go
type eventRepo interface {
    GetBestellungEventsSinceCursor(ctx context.Context, cursor int) ([]event.Event, error)
}

type druckerRepo interface {
    GetKategorieDrucker(ctx context.Context) (map[string]DruckerKonfig, error)
}
```

**DruckerKonfig:**

```go
type DruckerKonfig struct {
    IP       string // z.B. "192.168.1.51", leer = kein Drucker
    Bonmodus string // "pro_position" oder "pro_bestellung"
}
```

**DruckAuftrag (Application-DTO):**

```go
type DruckAuftrag struct {
    EventID   int
    DruckerIP string
    Payload   string // Base64-kodierter ESC/POS-Byte-String
}
```

### 3.2 · HTTP-Handler: Relay-Poll-Endpunkt

- [ ] Package `backend/api/relay/http/` erstellen
- [ ] `handler.go` mit `PollHandler()` — `POST /relay/poll`
- [ ] Token-Prüfung gegen `RELAY_AUTH_TOKEN` aus Konfiguration
- [ ] Request/Response-DTOs mit `json`-Tags

**Kontext:**
- `backend/api/table/http/` — als Referenz für Handler-Struktur
- `backend/api/helper/` — HTTP-Hilfsfunktionen (ReadBody, SendResponse, SendClientError)
- `backend/api/middleware/` — bestehende Middleware-Referenz

**Endpunkt:**

```
POST /relay/poll
Request:  {"token": "...", "lastEventId": 42}
Response: {"auftraege": [...], "cursor": 55}
```

Jeder Auftrag in der Response:

```json
{"eventId": 43, "druckerIp": "192.168.1.51", "payload": "<base64>"}
```

**Sicherheit:**
- Kein JWT — statischer `RELAY_AUTH_TOKEN` wird im Request-Body mitgesendet
- Token-Vergleich im Handler (nicht in Middleware), da das Relay kein Benutzer ist

### 3.3 · Routing-Integration

- [ ] `backend/api/relay.go` erstellen — `NewRelayApi()` Factory-Funktion
- [ ] In `main.go` unter `/relay/` mounten — kein JWT-Middleware
- [ ] `RELAY_AUTH_TOKEN` als neue Konfigurationsvariable in `backend/config/` ergänzen

**Kontext:**
- `backend/api/service.go`, `backend/api/admin.go` — als Referenz für API-Registrierung
- `backend/main.go` — Routing-Setup
- `backend/config/` — Konfiguration

**Routing:**

```go
mux.Handle("/relay/", http.StripPrefix("/relay", relayApi))
```

**Konfiguration (.env):**

```
RELAY_AUTH_TOKEN=<zufälliger Token, z.B. openssl rand -hex 32>
```

---

## Phase 4: Admin-Druckerkonfiguration

### 4.1 · Backend: Admin-Endpunkte für Druckerkonfiguration

- [ ] `POST /admin/get-drucker-config` — gibt aktuelle Konfiguration aller Kategorien zurück
- [ ] `POST /admin/update-drucker-config` — speichert Drucker-IP und Bonmodus per UPSERT
- [ ] zog-Validierung: IPv4-Format oder leer, Bonmodus `pro_position`/`pro_bestellung`
- [ ] Unit-Tests für Validierung und Fehler-Mapping

**Kontext:**
- `backend/api/admin.go` — Admin-Routen registrieren
- `backend/api/product/http/` — als Referenz für Admin-Handler-Struktur
- `backend/api/product/application/` — als Referenz für Application-Service-Struktur

### 4.2 · Frontend: Druckerkonfiguration im Admin-Bereich

- [ ] Neue Admin-Seite/Komponente für Druckerkonfiguration
- [ ] Pro Kategorie: Input-Feld für IP-Adresse, Dropdown für Bonmodus
- [ ] Zod-Validierung: IPv4-Format oder leer, Bonmodus `pro_position`/`pro_bestellung`
- [ ] Backend-Klasse `DruckerBackend` mit `BackendClient`
- [ ] In Admin-Navigation integrieren

**Kontext:**
- `frontend/src/admin/` — bestehende Admin-Seiten als Referenz
- `frontend/src/lib/Backend.ts` — BackendClient-Interface
- `frontend/src/routes.ts` — Routen-Konfiguration

**UI-Entwurf:**
- Tabelle mit 3 Zeilen (Essen, Getränke, Sonstiges)
- Pro Zeile: Kategorie-Name (readonly), IP-Adresse (Input), Bonmodus (Select)
- Speichern-Button pro Zeile oder als Gesamtformular
- Leere IP = kein Drucker für diese Kategorie (visueller Hinweis)
- Deutsche Labels: „Drucker-IP", „Bonmodus", „Pro Position", „Pro Bestellung"

---

## Phase 5: Print-Relay (eigenständiges Go-Binary)

### 5.1 · Relay-Binary: Grundstruktur

- [ ] `cmd/relay/main.go` erstellen
- [ ] CLI-Flags: `--backend`, `--token`, `--state`, `--poll`
- [ ] Poll-Loop: HTTP-POST an `/relay/poll`, Druck-Aufträge verarbeiten

**Kontext:**
- Kein Import aus `backend/` — das Relay ist ein eigenständiges Binary
- Nur Go-Stdlib (kein pgx, kein zerolog)

### 5.2 · Relay: Exactly-Once Delivery

Der Relay implementiert robustes Exactly-Once Delivery durch zwei Mechanismen:

**Mechanismus 1 — Per-Event Cursor-Inkrement:**
Nach jedem einzelnen erfolgreich gedruckten Auftrag wird der Cursor sofort lokal persistiert (nicht erst nach dem gesamten Batch). Das minimiert das Fenster für Doppeldrucke bei Absturz auf maximal einen einzelnen Auftrag.

**Mechanismus 2 — Lokale Idempotenzliste:**
Das Relay führt eine lokale Liste der zuletzt gedruckten Event-IDs (z.B. letzte 2000 IDs). Vor jedem Druckvorgang wird geprüft, ob die Event-ID bereits gedruckt wurde. Das fängt den Edge-Case ab, dass der Cursor nach erfolgreichem Druck, aber vor dem Persistieren des neuen Cursors, nicht aktualisiert wurde.

- [ ] Per-Event Cursor-Inkrement implementieren (Cursor nach jedem erfolgreichen Druck speichern, nicht nach Batch)
- [ ] Lokale Idempotenzliste implementieren (Set der zuletzt gedruckten Event-IDs, max. 2000 Einträge)
- [ ] Atomare State-Datei-Schreibvorgänge (Write-to-Temp + Rename, um Korruption bei Absturz zu vermeiden)

**State-Datei Struktur:**

```json
{
  "last_event_id": 42,
  "printed_event_ids": [38, 39, 40, 41, 42]
}
```

### 5.3 · Relay: Drucker-Kommunikation

- [ ] `checkPrinter(ip)` — DLE EOT 4 Statusabfrage via TCP:9100
- [ ] `sendToPrinter(ip, data)` — ESC/POS-Bytes an Drucker senden
- [ ] Retry mit Maximum: Max. 60 Versuche (5 Min.) pro Drucker, dann Auftrag überspringen und loggen
- [ ] Write-Deadline auf TCP-Verbindung setzen

**Kontext:**
- Hardware: MUNBYN ITPP047P-UE, TCP:9100, ESC/POS
- DLE EOT 4 (`\x10\x04\x04`): Papiersensor-Status abfragen
- Antwortbyte: Bit 5 (0x20) = Papier fast leer, Bit 6 (0x40) = Papier leer

### 5.4 · Relay: Robustheit

- [ ] Graceful Shutdown (SIGTERM/SIGINT Handler) — laufenden Druckjob abschließen, dann State speichern und beenden
- [ ] HTTP Response-Statuscode prüfen vor JSON-Decode (401, 500, etc.)
- [ ] Poll-Fehler loggen, aber weiterlaufen (kein Fatal bei transienten Netzwerkfehlern)
- [ ] Periodischer Statuslog (z.B. alle 5 Minuten „Relay aktiv, Cursor bei X, keine neuen Aufträge")

### 5.5 · Relay: Cross-Compilation und Deployment

- [ ] Makefile-Target für Relay-Build ergänzen (optional)
- [ ] Dokumentation: Build-Befehle für Windows (amd64), Linux (arm64/amd64)
- [ ] Beispiel-Startbefehl in Benutzerdokumentation

**Build-Befehle:**

```bash
# Windows-PC:
GOOS=windows GOARCH=amd64 go build -o jotti-relay.exe ./cmd/relay/

# Raspberry Pi (Linux ARM64):
GOOS=linux GOARCH=arm64 go build -o jotti-relay ./cmd/relay/

# Starten:
./jotti-relay --backend="https://jotti.meinverein.de" --token="<TOKEN>" --poll=2
```

---

## Phase 6: Integration + Tests

### 6.1 · Integrationstests

- [ ] Integrationstest: Bestellung aufnehmen → Relay pollt → erhält korrekte Druck-Aufträge
- [ ] Integrationstest: Cursor-Fortschritt (zweiter Poll liefert keine Duplikate)
- [ ] Integrationstest: Druckerkonfiguration ändern → nächster Poll nutzt neue IPs
- [ ] Integrationstest: Kein Drucker konfiguriert → leere Aufträge

### 6.2 · End-to-End-Validierung

- [ ] `make check` läuft durch
- [ ] `make lint` läuft durch
- [ ] Manueller Test: Dev-Stack starten, Bestellung aufnehmen, Relay-Endpunkt manuell testen
- [ ] Dokumentation: `docs/bondruck.md` aktualisiert

---

## Abhängigkeiten

```
Phase 1 ─── unabhängig
Phase 2 ─── unabhängig (reine Funktionen, kein DB-Zugriff)
Phase 3 ─── abhängig von Phase 1 + Phase 2
Phase 4 ─── abhängig von Phase 1 (Backend) + Phase 3 (für End-to-End)
Phase 5 ─── abhängig von Phase 3 (braucht funktionierenden Endpunkt zum Testen)
Phase 6 ─── abhängig von allen vorherigen Phasen
```

**Parallelisierbar:** Phase 1 und Phase 2 können parallel bearbeitet werden.
**Parallelisierbar:** Phase 4 (Frontend) kann parallel zu Phase 5 (Relay) bearbeitet werden, sobald Phase 3 abgeschlossen ist.
