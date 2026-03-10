# ADR-Backlog — Fehlende Architecture Decision Records

> Dieses Dokument listet Architekturentscheidungen, die noch als eigenständige ADRs ausgearbeitet werden sollten.
> Das [System Design](../design.md) dokumentiert seit Kurzem viele Entscheidungen inline (D1–D15, §1–§15).
> Dieses Backlog trackt, welche davon die Tiefe eines eigenständigen ADRs benötigen (Alternativen, Trade-offs,
> Revisionskriterien) und welche durch design.md bereits ausreichend abgedeckt sind.
>
> **Vorhandene ADRs:** [event-sourcing.md](event-sourcing.md), [orm.md](orm.md)
> **Designentscheidungen in design.md:** D1–D15 (§15.2), offene Fragen Q1–Q5 (§15.3)

---

## Entfallen — durch design.md ausreichend abgedeckt

Die folgenden Entscheidungen waren bisher als fehlende ADRs gelistet, sind aber durch das System Design Dokument nun hinreichend dokumentiert (Entscheidung + Begründung + Kontext). Ein eigenständiges ADR wäre Duplikation.

| Thema                                            | Abgedeckt in design.md                                                 | Designentscheidung |
| ------------------------------------------------ | ---------------------------------------------------------------------- | ------------------ |
| Modularer Monolith                               | §5.1 (Warum kein Microservice), §6 (Modulstruktur, Inter-Modul-Regeln) | D1                 |
| Backend als Single Source of Truth               | §12.2 (Architekturprinzipien), §10.1 (Validierung beidseitig)          | D12                |
| Frontend State Management (Singleton)            | §12.2 („Kein globaler State-Store"), §12.4 (Daten-Fluss-Pattern)       | D11                |
| Soft-Delete via EntityStatus                     | §9.5 (Soft-Delete-Strategie + Begründung: FK-Constraints, Audit-Trail) | D10                |
| Feature-basierte Ordnerstruktur                  | §6.1 (Modulstruktur), §12.6 (Frontend-Verzeichnisstruktur)             | D13                |
| Deutsche Ubiquitous Language                     | §3 (Sprachregeln + kanonisches Glossar)                                | D6                 |
| Duale Validierung (Zod + zog)                    | §13.1 (Validierungsdiagramm), §10.1 (Grundprinzipien)                  | D12                |
| Fat Events mit eingebetteten Daten               | §7.3 (Event-Design-Prinzipien: self-contained, kein Nachladen)         | D4                 |
| PostgreSQL als Event Store                       | §7.2 (Event Store DDL + Immutabilitäts-Garantie)                       | D7                 |
| Append-only mit DB-Triggern                      | §7.2 (4 Defense-in-Depth-Maßnahmen aufgelistet)                        | D14                |
| Snapshot als Event → Ablösung durch Projektionen | §7.5 (Mechanismus + geplante Evolution), §8 (CQRS Stufe 2)             | D15                |

---

## Priorität: Hoch

Sicherheits- und vertragskritische Entscheidungen. design.md dokumentiert das _Was_, aber nicht die Tiefe an Alternativen und Trade-offs, die ein ADR erfordert.

### [ ] ADR: JWT-Konfiguration und Session-Management

**Designentscheidung:** D8 — „JWT mit HS256, 12h Gültigkeit — simpel, passend für Session-Dauer einer Veranstaltung"
**Abgedeckt in design.md:** §11.1 (Konfigurationstabelle), §11.2 (Passwort-Flow), §11.4 (401-Interceptor)
**Betroffene Dateien:**

- `backend/domain/jwt/jwt.go` — HS256, Issuer "jotti", 12h Gültigkeit
- `backend/api/middleware/middleware.go` — Token-Extraktion aus Authorization-Header
- `frontend/src/lib/Auth.ts` — Token in localStorage, automatischer Logout bei 401

**Was design.md NICHT abdeckt — im ADR zu klären:**

- **HS256 vs. RS256:** Warum symmetrisch? Trade-off: Einfachheit vs. Verifizierung durch Dritte. Revisionskriterium: Wird relevant bei Multi-Service-Architektur.
- **12h ohne Refresh-Token:** Risiko bei gestohlenem Token (kein Revocation-Mechanismus). Trade-off: Einfachheit vs. Session-Hijacking-Risiko. Ist 12h zu lang für ein öffentlich genutztes Gerät?
- **`localStorage` vs. `httpOnly`-Cookie:** XSS-Risiko bei localStorage. POST-only API eliminiert CSRF, daher ist Cookie weniger nötig — aber die Abwägung fehlt.
- **Token-Revocation-Strategie:** Fehlt komplett. Aktuell unmöglich ohne Server-State. Akzeptiertes Risiko oder Lücke?

---

### [ ] ADR: Argon2id-Parameter für Passwort-Hashing

**Designentscheidung:** D9 — „Argon2id — State-of-the-Art, resistent gegen GPU/ASIC-Angriffe"
**Abgedeckt in design.md:** §11.2 (Passwort-Flow mit Argon2id-Hash)
**Betroffene Dateien:**

- `backend/domain/user/password.go` — Argon2id-Konfiguration

**Was design.md NICHT abdeckt — im ADR zu klären:**

- **Konkrete Parameter-Wahl:** TimeCost=2, MemoryCost=64MB, Threads=2, Salt=16B, Key=32B. Warum genau diese Werte? OWASP-Empfehlung? Benchmark auf Ziel-Hardware?
- **Argon2id vs. bcrypt vs. scrypt:** Vergleich Memory-Hardness, Verbreitung, Library-Reife
- **Login-Latenz:** 64MB Memory-Allokation pro Login-Versuch. Auswirkung bei gleichzeitigen Logins zu Veranstaltungsbeginn (z.B. 30 Servicekräfte loggen sich ein → 30 × 64MB)
- **Upgrade-Pfad:** Was passiert bei Parameterverschärfung? Re-Hash bei nächstem Login? Marker für veraltete Hashes?

---

### [ ] ADR: Security Headers ausschließlich im Reverse Proxy

**Abgedeckt in design.md:** §14.3 (nginx: TLS-Terminierung, Routing)
**Betroffene Dateien:**

- `reverse-proxy/nginx.conf` — CSP, HSTS (63 Tage), X-Frame-Options DENY, X-Content-Type-Options, Permissions-Policy
- `backend/app/app.go` — Backend setzt keine Security Headers

**Was design.md NICHT abdeckt — im ADR zu klären:**

- design.md §14.3 erwähnt nginx-Funktionen, aber **nicht die Security-Header-Strategie** (CSP, HSTS etc.)
- Bewusste Entscheidung oder Lücke? Defense-in-Depth spricht für Header in beiden Schichten
- Backend ohne nginx ist unsicher — ist das akzeptiert? Auswirkung auf `docker-compose.dev.yml`?
- Vergleich: Header nur in nginx vs. zusätzlich im Backend vs. nur im Backend

---

## Priorität: Mittel

Architektur- und evolutionsrelevant. design.md dokumentiert die Entscheidung, aber die Tiefe eines ADRs (Alternativen, Revisionskriterien) fehlt.

### [ ] ADR: POST-only API-Design

**Designentscheidung:** D3 — „Vereinfachung, konsistentes Verhalten, keine Cache-Probleme"
**Abgedeckt in design.md:** §10.1 (Grundprinzip: POST-only), §10.2 (vollständige Endpoint-Übersicht)
**Betroffene Dateien:**

- `backend/api/middleware/middleware.go` — `PostMethodOnlyMiddleware`
- `backend/api/service.go`, `backend/api/admin.go`, `backend/api/auth.go` — alle Routen

**Was design.md NICHT abdeckt — im ADR zu klären:**

- D3 nennt die Begründung einzeilig. **Alternativen fehlen:** RESTful Verbs, GraphQL, gRPC — warum verworfen?
- **Caching-Trade-off:** GET ist cacheable (Browser, CDN, Proxy), POST nicht. Bei einem Self-hosted-System mit wenigen Clients relevant?
- **Konventions-Bruch:** Jeder neue Entwickler erwartet REST. Ist der Vereinfachungsgewinn diesen Bruch wert?
- Passt POST-only zur Event-Sourcing-Philosophie? (Alles ist ein Command/Query, nicht CRUD)

**Priorität herabgestuft** (von Hoch auf Mittel): design.md §10 dokumentiert das Was und Warum. Ein ADR würde Tiefe bei Alternativen ergänzen, ist aber nicht sicherheitskritisch.

---

### [ ] ADR: Event-Typ-Versionierung per Suffix-Konvention

**Abgedeckt in design.md:** §7.3 (Event-Design-Prinzipien: „Versioniert — Event-Typen mit `:v1`-Suffix für Schema-Evolution"), §7.4 (Event-Typen-Tabelle)
**Betroffene Dateien:**

- `backend/domain/table/events.go` — Event-Typ-Konstanten (`tisch.bestellung-aufgegeben:v1`, etc.)

**Was design.md NICHT abdeckt — im ADR zu klären:**

- **Migrations-Strategie bei `:v2`:** Parallele Handler für v1+v2? Upcasting (v1 → v2 beim Lesen)? Einmalige Migration bestehender Events? Nichts davon ist dokumentiert.
- **Version im Type vs. in Metadaten:** CloudEvents hat ein `schemaversion`-Attribut. Warum stattdessen im Type-String?
- **Revisionskriterium:** Wird kritisch, sobald Events tatsächlich migriert werden müssen (erster Schema-Change)

---

### [ ] ADR: Manuelles Dependency Wiring statt DI-Container

**Abgedeckt in design.md:** §5.3 („Kein DI-Framework — nur Go-Konstruktoren")
**Betroffene Dateien:**

- `backend/app/app.go` — `NewApp()` verdrahtet alle Dependencies manuell

**Was design.md NICHT abdeckt — im ADR zu klären:**

- Warum kein Wire/Fx/dig? Go-Idiome bevorzugen explizites Wiring — aber ist das dokumentierte Absicht oder Default?
- Ab welcher Modulzahl wird manuelles Wiring zum Problem? (Aktuell 4 Module — handhabbar)
- Auswirkung auf Testbarkeit: Mocks müssen manuell injiziert werden, keine Auto-Mocking-Option

---

### [ ] ADR: Sentinel Errors für Datenbank-Fehler-Mapping

**Abgedeckt in design.md:** §13.2 (Fehler-Mapping-Diagramm: PostgreSQL → Sentinel → HTTP Status)
**Betroffene Dateien:**

- `backend/db/db.go` — `Error()` Funktion, `ErrNotFound`, `ErrAlreadyExists`, `ErrDatabase`
- HTTP-Handler mappen Sentinel Errors auf HTTP-Status-Codes

**Was design.md NICHT abdeckt — im ADR zu klären:**

- §13.2 dokumentiert das Mapping, aber nicht die **Designentscheidung** (warum nur 3 Fehlertypen?)
- Reicht das? (z.B. kein `ErrConflict` für Race Conditions, kein `ErrTimeout`)
- Vergleich mit Error-Wrapping (`%w`) oder Custom-Error-Typen mit strukturiertem Kontext

---

## Priorität: Niedrig

Operativ relevante Entscheidungen. Teilweise durch design.md abgedeckt, Rest hat begrenzten Wirkungsbereich.

### [ ] ADR: Rate Limiting — doppelt auf zwei Ebenen

**Abgedeckt in design.md:** §13.4 (Rate-Limiting erwähnt, aber nur für `/auth/login`)
**Betroffene Dateien:**

- `backend/api/middleware/middleware.go` — `RateLimitMiddleware` (10 req/s, Burst 20, per IP, in-memory)
- `reverse-proxy/nginx.conf` — `limit_req_zone` (10 req/s, Burst 20)

**Kommentar:** §13.4 beschreibt Rate-Limiting nur auf sensiblen Endpoints, aber im Code gibt es Rate-Limiting auf **zwei Ebenen** (nginx + Go). Bewusst doppelt oder Redundanz? Klären, ob unterschiedliche Limits sinnvoll wären.

---

### [ ] ADR: Docker-Netzwerksegmentierung

**Abgedeckt in design.md:** §14.1 (Container-Diagramm zeigt Topologie)
**Betroffene Dateien:**

- `docker-compose.yml` — `app-network`, `db-network`

**Kommentar:** §14.1 zeigt die Topologie, aber nicht die bewusste Netzwerktrennung (Frontend/nginx haben keinen DB-Zugriff). Defense-in-Depth-Begründung fehlt.

---

### [ ] ADR: Graceful Shutdown mit 30-Sekunden-Timeout

**Nicht in design.md abgedeckt.**
**Betroffene Dateien:**

- `backend/app/app.go` — Signal-Handling, `context.WithTimeout(30s)`

**Kommentar:** Begründung für 30s fehlt (maximale Request-Dauer? DB-Transaction-Timeout?). Relevant für Rolling Deployments.

---

### [ ] ADR: HTTP-Server-Timeouts und Connection-Pool-Konfiguration

**Nicht in design.md abgedeckt.**
**Betroffene Dateien:**

- `backend/app/app.go` — ReadTimeout 5s, WriteTimeout 10s, IdleTimeout 120s
- `backend/main.go` — ConnMaxLifetime 5min, MaxOpenConns 50, MaxIdleConns 10

**Kommentar:** Hardcoded Werte ohne dokumentierte Herleitung. Besonders relevant bei WebSocket oder Streaming-Endpoints.

---

### [ ] ADR: Onetime-Password für Benutzer-Onboarding

**Abgedeckt in design.md:** §11.2 (Passwort-Flow-Diagramm: Einmalpasswort → eigenes Passwort)
**Betroffene Dateien:**

- `database/migrations/01_initial.up.sql` — `onetime_password_hash`-Spalte
- `backend/domain/user/` — Passwort-Logik

**Kommentar:** §11.2 dokumentiert den Flow. Alternativenbewertung (Magic-Link, E-Mail-Einladung) fehlt — aber pragmatisch begründbar (kein E-Mail-Versand). Niedrige Priorität.

---

### [ ] ADR: 12-Factor-Konfiguration via Umgebungsvariablen

**Abgedeckt in design.md:** §14.2 (Umgebungen-Tabelle)
**Betroffene Dateien:**

- `backend/config/config.go` — alle Werte aus `os.Getenv()`
- `docker-compose.dev.yml`, `docker-compose.yml` — Environment-Variablen

**Kommentar:** design.md listet Umgebungen, aber nicht die einzelnen Variablen, Defaults oder Secrets-Rotation-Strategie.

---

### [x] ~~ADR: Tailwind CSS 4 via Vite-Plugin~~ — Entfällt

**Begründung:** design.md §12.1 dokumentiert den Frontend-Stack inkl. Tailwind CSS 4. Reine Tooling-Wahl ohne architektonische Trade-offs, die über den Tech-Stack-Eintrag hinausgehen.

---

## Neu: Offene Designfragen (Q1–Q5) → potenzielle ADRs

design.md §15.3 listet 5 offene Designfragen, die bei Entscheidung als ADR dokumentiert werden sollten:

### [ ] ADR: Tischumbuchung (Q1)

**Offene Frage:** Wie wird die Tischumbuchung atomar umgesetzt?
**Kontext:** Zwei Events (Storno am Quell-Tisch + Neubestellung am Ziel-Tisch) müssen in einer Transaktion liegen. Das bricht mit dem Single-Aggregate-Prinzip im Event Sourcing (ein Command → ein Aggregate).

**Zu klären:**

- Saga-Pattern (zwei Commands mit Kompensation) vs. atomare Multi-Aggregate-Transaktion?
- Neuer Event-Typ `tisch.umbuchung:v1` oder Komposition bestehender Events?
- Konsistenzgarantie: Was passiert, wenn das zweite Event fehlschlägt?

---

### [ ] ADR: Tagesabschluss (Q2)

**Offene Frage:** Wie wird der Tagesabschluss mit offenen Tischen behandelt?
**Kontext:** Am Ende einer Veranstaltung haben Tische möglicherweise offenen Saldo. Manuelles Schließen vs. automatische Stornierung vs. Übertrag.

**Zu klären:**

- Neuer Event-Typ `tisch.abschluss:v1`?
- Wie werden Berichte/Auswertungen generiert? (Read Model nötig?)
- Berechtigungskonzept: Nur Admin?

---

### [ ] ADR: Bon-Druck-Integration (Q3)

**Offene Frage:** Wie wird Bon-Druck integriert?
**Kontext:** Thermaldrucker als Side-Effect nach Bestell-Event. Touch-Point mit physischer Hardware — einzige Stelle im System.

**Zu klären:**

- Browser-Print-API vs. direkter ESC/POS-Treiber vs. Cloud-Print-Service?
- Synchron (blockiert UI) oder asynchron (Fire-and-Forget)?
- Fehlerbehandlung: Was wenn Drucker offline? Retry? Queue?

---

### [ ] ADR: Offline-Fähigkeit (Q4)

**Offene Frage:** Wann und wie wird Offline-Fähigkeit implementiert?
**Kontext:** Service Worker + lokale Queue + Sync bei Reconnect. Kollidiert mit „Backend als Single Source of Truth" (D12 / §12.2) — Offline erfordert Client-seitige Logik.

**Zu klären:**

- Welche Operationen offline verfügbar? (Nur Lesen? Oder auch Bestellen?)
- Konfliktauflösung bei Reconnect (Optimistic Concurrency?)
- Revisionsbedarf für D12 (Backend als SoT) — oder explizite Offline-Ausnahme?

---

### [ ] ADR: Freibon / Freie Preiseingabe (Q5)

**Offene Frage:** Soll der Freibon ein eigener Event-Typ werden?
**Kontext:** Freie Preiseingabe ohne Produkt-Zuordnung. Identifiziert als Hotspot im Event Storming.

**Zu klären:**

- Eigener Event-Typ `tisch.freibon:v1` vs. spezielle Position in `bestellung-aufgegeben`?
- Validierung: Min/Max-Betrag? Berechtigung (nur Admin/Serviceleitung)?
- Auswirkung auf Auswertungen (Freibons separat ausweisbar?)

---

## Hinweise zur Ausarbeitung

Jedes ADR sollte dem bestehenden Format folgen (siehe `event-sourcing.md`, `orm.md`):

```
# ADR: <Titel>
## Status        — Entschieden | Vorgeschlagen | Abgelöst
## Kontext       — Warum steht diese Entscheidung an?
## Entscheidung  — Was wurde entschieden?
## Alternativen  — Was wurde bewertet und verworfen?
## Konsequenzen  — Positive und negative Auswirkungen
```

Beim Schreiben darauf achten:

- **Verweis auf design.md** — Wenn die Entscheidung dort bereits dokumentiert ist (D1–D15), im ADR darauf verweisen statt duplizieren
- **Code-Referenzen** mit Dateipfaden angeben (nicht nur beschreiben)
- **Alternativen ehrlich bewerten** — nicht nur die gewählte Lösung loben
- **Konsequenzen explizit benennen** — insbesondere akzeptierte Nachteile
- **Trigger für Revision** definieren — unter welchen Umständen sollte die Entscheidung revidiert werden?
