# ADR: Bondruck (K-12) — Cursor-basiertes Event-Polling mit Print-Relay

## Status

**Accepted** — Cursor-basiertes Event-Polling statt Transactional Outbox Pattern. Print-Relay als lokale Netzwerk-Brücke.

## Kontext

### Ausgangslage

Anforderung K-12 erfordert automatischen Bondruck an Ausgabestationen (Küche, Getränketheke) beim Aufnehmen einer Bestellung. Da jotti auf einem Cloud-VPS läuft, sich die Bondrucker aber im lokalen Netzwerk des Vereinsfests befinden (NAT/Firewall), ist eine Zwei-Komponenten-Architektur notwendig: Cloud-Backend + lokaler Print-Client.

### Lastprofil

| Kennzahl                       | Wert         |
| ------------------------------ | ------------ |
| Bestellungen pro Veranstaltung | < 2.000      |
| Bons pro Veranstaltung         | < 5.000      |
| Latenzanforderung              | < 5 Sekunden |
| Gleichzeitige Drucker          | 2–3          |

### Randbedingungen

- jotti verwendet ausschließlich POST-Endpunkte
- Bestellungs-Events sind Fat Events (enthalten Positionen, Kategorie, Tischname, Servicekraft, Zeitstempel, Kommentar)
- `WriteEvent()` darf nicht für Bondruck modifiziert werden (Isolation der Core Domain)
- Ehrenamtliche Helfer müssen das System ohne IT-Kenntnisse betreiben können
- Hardware: ESC/POS-Bondrucker via Ethernet (TCP:9100), kein WLAN/Bluetooth

## Bewertete Alternativen

### A: Transactional Outbox Pattern mit `print_jobs`-Tabelle (verworfen)

Eine `print_jobs`-Tabelle wird in derselben Transaktion wie das Event beschrieben. Das Relay pollt diese Tabelle, ein ACK-Endpunkt markiert Jobs als erledigt.

| Vorteil                                      | Nachteil                                                                      |
| -------------------------------------------- | ----------------------------------------------------------------------------- |
| Atomare Konsistenz (Event + Job in einer TX) | Kopplung: `WriteEvent()` braucht TxHook-Mechanismus                           |
| Bekanntes Enterprise-Pattern                 | Unnötige Duplikation — Fat Events enthalten bereits alle Daten                |
|                                              | Drucker-IP zum Schreibzeitpunkt festgelegt (stale bei Konfigurationsänderung) |
|                                              | ESC/POS-Payload pre-rendered (Formatfehler nicht korrigierbar)                |
|                                              | Nicht isolierbar — Bondruck in Core-Domain-Write-Pfad eingebettet             |
|                                              | Zusätzlicher ACK-Endpunkt nötig                                               |

### B: Cursor-basiertes Event-Polling (gewählt)

Die `kassenjournal`-Tabelle IST die Outbox. Das Relay sendet seinen Cursor (`lastEventId`) und erhält neue Bestellungs-Events. Drucker-IPs werden zur Lesezeit aufgelöst, ESC/POS wird on-the-fly generiert.

| Vorteil                                                               | Nachteil                                                           |
| --------------------------------------------------------------------- | ------------------------------------------------------------------ |
| Keine Änderung an `WriteEvent()`                                      | At-Least-Once statt Exactly-Once (mitigiert durch Idempotenzliste) |
| Keine zusätzliche Tabelle                                             | Polling-Latenz (~2 Sekunden)                                       |
| Drucker-IP zur Lesezeit aufgelöst (immer aktuell)                     |                                                                    |
| ESC/POS on-the-fly generiert (Format jederzeit änderbar)              |                                                                    |
| Vollständig isolierbar (Relay-Code hat keinen Import aus Core Domain) |                                                                    |
| Feature abschaltbar ohne Code-Änderung am Event-Store                 |                                                                    |

### C: WebSocket statt HTTP-Polling (verworfen)

| Vorteil                    | Nachteil                                              |
| -------------------------- | ----------------------------------------------------- |
| Geringere Latenz (< 100ms) | Architektur-Mismatch (jotti = POST-only)              |
| Kein unnötiges Polling     | WebSocket-Hub, Goroutinen, Ping/Pong, Reconnect-Logik |
|                            | gorilla/websocket seit 2022 archiviert                |

**Verworfen:** 2-Sekunden-Latenz ist für Küche/Theke vollkommen akzeptabel. Die zusätzliche Komplexität eines WebSocket-Hubs ist unverhältnismäßig.

### D: VPN (WireGuard/Tailscale) statt Print-Relay (verworfen)

Direkte TCP-Verbindung vom Cloud-Backend zum Drucker via VPN-Tunnel.

**Verworfen:** Hoher Setup-Aufwand für ehrenamtliche Helfer, nicht für wechselnde Netze (LTE) geeignet, erfordert Router-Zugriff am Veranstaltungsort.

### E: Browser-Druck (Print CSS / HTML) (verworfen)

Servicekraft druckt manuell aus dem Browser via Ctrl+P.

**Verworfen:** Nicht automatisch (K-12 fordert automatischen Druck), keine ESC/POS-Kontrolle, schlechte Layout-Kontrolle.

## Entscheidung

**Cursor-basiertes Event-Polling mit Print-Relay.** Zwei Komponenten:

1. **Cloud-Backend:** Ein neuer `POST /relay/poll`-Endpunkt liest `bestellung-aufgenommen`-Events ab einem Cursor, löst Drucker-IPs zur Lesezeit per JOIN auf `kategorie_drucker` auf, generiert ESC/POS-Payloads on-the-fly und gibt Druck-Aufträge zurück.
2. **Print-Relay:** Ein eigenständiges Go-Binary im lokalen Netzwerk pollt den Backend-Endpunkt, druckt ESC/POS-Bytes an Bondrucker via TCP:9100 und verwaltet seinen Cursor lokal.

### Delivery-Garantie

**Ziel: Exactly-Once Delivery (Best Effort)**

Zwei Mechanismen arbeiten zusammen:

1. **Per-Event Cursor-Inkrement:** Der Cursor wird nach jedem einzelnen erfolgreich gedruckten Auftrag persistiert (nicht nach dem gesamten Batch). Das minimiert das Fenster für Doppeldrucke auf maximal einen Auftrag.
2. **Lokale Idempotenzliste:** Das Relay führt eine begrenzte Liste der zuletzt gedruckten Event-IDs. Vor jedem Druck wird geprüft, ob die Event-ID bereits gedruckt wurde. Das fängt den Edge-Case ab, dass der Cursor nach Druck, aber vor Persistierung, verloren geht.

**Verbleibende Edge-Cases:** Ein Doppeldruck ist nur möglich, wenn das Relay exakt zwischen „Bytes an Drucker gesendet" und „Cursor-Datei geschrieben" abstürzt UND die Idempotenzliste ebenfalls verloren geht (z.B. bei Stromausfall). In diesem Fall druckt die Küche maximal einen Bon doppelt — erkennbar am identischen Zeitstempel. Für ein Vereinsfest ist das akzeptabel.

### Robustheit des Print-Relays

- **Atomare State-Datei-Schreibvorgänge:** Write-to-Temp + Rename, um Korruption bei Absturz zu vermeiden.
- **Drucker-Retry mit Maximum:** Max. 60 Versuche (5 Minuten) pro Drucker, dann Auftrag überspringen und loggen. Verhindert, dass ein dauerhaft offline Drucker das gesamte Relay blockiert.
- **Graceful Shutdown:** SIGTERM/SIGINT-Handler schließt laufenden Druckjob ab, speichert State und beendet sauber.
- **HTTP-Fehlerbehandlung:** Statuscode prüfen vor JSON-Decode (401 → Token falsch, 500 → Backend-Fehler).

### Authentifizierung

Statischer `RELAY_AUTH_TOKEN` (kein JWT), da das Relay kein Benutzer ist. Der Token wird im Request-Body mitgesendet und im Handler geprüft (nicht in Middleware). Der Token wird als Umgebungsvariable konfiguriert.

### Bonmodus

Zwei Modi, konfigurierbar pro Kategorie in der `kategorie_drucker`-Tabelle:

- **Pro Position (Standard):** Jede Position einer Bestellung erzeugt einen eigenen Bon — entspricht dem Küchen-Workflow (1 Bon = 1 Arbeitsauftrag).
- **Pro Bestellung:** Alle Positionen einer Kategorie werden auf einem Sammelbon zusammengefasst.

### Bon-Design

- Kein Gesamtpreis auf dem Bon — die Küche braucht Arbeitsaufträge, keine Preise
- Tisch dominant (doppelte Schriftgröße) — sofort erkennbar beim Aufhängen
- Kommentar prominent — Sonderwünsche direkt unter der Position
- Metadaten (Uhrzeit, Servicekraft) sekundär unten

## Konsequenzen

### Positiv

- **Vollständige Isolation:** Kein Import aus der Core Domain, kein Einfluss auf `WriteEvent()`. Bondruck kann entfernt werden, ohne den Kassenbetrieb zu berühren.
- **Einfachheit:** Ein einziger Endpunkt (`POST /relay/poll`), keine `print_jobs`-Tabelle, kein ACK-Endpunkt, kein TxHook.
- **Aktuelle Konfiguration:** Drucker-IPs und ESC/POS-Format werden zur Lesezeit aufgelöst — Änderungen wirken sofort.
- **Portables Relay:** Ein einziges Go-Binary ohne Abhängigkeiten, cross-compilierbar für Windows, Linux (x86/ARM).
- **NAT-freundlich:** Ausgehende HTTPS-Verbindung vom Festzelt zum Cloud-VPS — kein VPN, kein Router-Zugriff nötig.

### Negativ

- **Polling-Latenz:** ~2 Sekunden Verzögerung zwischen Bestellung und Bondruck (akzeptabel für Küche/Theke).
- **Kein Exactly-Once auf Infrastrukturebene:** Delivery-Garantie basiert auf Application-Level-Logik (Cursor + Idempotenzliste), nicht auf transaktionaler Atomarität.
- **Zusätzliche Komponente:** Das Print-Relay muss auf einem Gerät im lokalen Netzwerk laufen (Raspberry Pi, Windows-PC, Laptop).

## Referenzen

- [ADR: Event-Sourcing für Kasse-Operationen](event-sourcing.md) — Fat Events als Datengrundlage
- [ADR: CQRS-Projektionen für Kasse-Kontext](cqrs.md) — Projektionsarchitektur, Read-Model-Strategie
- [Anforderungen K-12](../anforderungen.md) — Akzeptanzkriterien für Bondruck
