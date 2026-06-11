# Plan: Bondruck-Verbesserungen — Druckstationen vereinheitlichen, Outbox robust machen

> Source PRD: docs/prds/prd-bondruck-verbesserungen.md

## Goal

Alle Drucker (drei Produktkategorien plus Kassenbeleg und Abholbon) werden als Druckstationen auf einer einzigen Admin-Seite konfiguriert; die Singleton-Tabelle `bondruck_einstellungen` und der Direktverkauf-Modus entfallen ersatzlos (Ableitungsregel: Abholbon-Station konfiguriert → Abholbon(s) gemäß ihrem Bonmodus, sonst Produktstationen, sonst nichts). Das Print-Relay verarbeitet Drucker unabhängig voneinander mit genau einem Zustellversuch pro Auftrag und Zyklus; das Backend zählt Fehlversuche und markiert Aufträge nach drei Versuchen als `fehlgeschlagen`. Fehlgeschlagene Aufträge sind auf der Druckstationen-Seite sichtbar und können erneut eingereiht oder verworfen werden — kein manueller DB-Eingriff mehr.

## Architectural decisions

Durable Entscheidungen, die für alle Phasen gelten:

- **Routes**:
  - `POST /relay/poll` bleibt unverändert (offene Aufträge, älteste zuerst).
  - `POST /relay/ergebnis` ersetzt `POST /relay/quittieren`. Body: `{token, gedruckteIds: [int], fehlversuche: [{id: int, fehler: string}]}` — ein Request pro Zyklus.
  - Neue admin-only Endpoints (POST-only, Konvention `verb-objekt`): `/admin/get-fehlgeschlagene-druckauftraege`, `/admin/druckauftrag-erneut-versuchen`, `/admin/druckauftrag-verwerfen`.
  - `/admin/get-druckstationen` und `/admin/update-druckstationen` bleiben, erweitert auf fünf Kategorien.
  - `/admin/get-bondruck-einstellungen` und `/admin/update-bondruck-einstellungen` entfallen ersatzlos.
- **Schema** (alle Änderungen direkt in `database/migrations/01_initial.up.sql`, Down-Skript mitziehen — Pre-Release-Regel):
  - Neuer DB-Enum `DruckstationKategorie` (`essen`, `getraenk`, `sonstiges`, `kassenbeleg`, `abholbon`) als PK-Typ von `druckstationen`; `ProduktKategorie` bleibt unverändert für Produkte. Fünf Seed-Zeilen.
  - `druckstationen.bonmodus` wird NULLable; CHECK-Constraint erzwingt: `essen`/`getraenk`/`sonstiges`/`abholbon` `IN ('pro_position','pro_bestellung')`, `kassenbeleg` NULL.
  - `druckauftraege`: Status-CHECK erweitert auf (`offen`, `gedruckt`, `fehlgeschlagen`, `verworfen`); neue Spalten `versuche INT NOT NULL DEFAULT 0` und `letzter_fehler TEXT NULL`.
  - Tabelle `bondruck_einstellungen` entfällt ersatzlos.
- **Key models**:
  - Domain `druckstation.Druckstation` bekommt typisierte Kategorie- und Bonmodus-Werte mit Konstruktor-Validierung (Muster des bisherigen `settings.BondruckEinstellungen`); zog/Zod-Schemas bleiben an den HTTP-Rändern. `settings.BondruckEinstellungen` und `settings.DirektverkaufModus` entfallen.
  - Policy-Schnittstelle: `CreateArbeitsbonAuftraegeFromEvent(evt, druckstationen)` — die `DirektverkaufBondruckKonfiguration` entfällt; die Ableitungsregel lebt in der Policy.
  - Ein gemeinsamer Druckauftrag-Typ zwischen Policy und Repository ersetzt die doppelten 1:1-Mappings.
  - Statusübergänge der Outbox: `offen → gedruckt` (Quittierung), `offen → fehlgeschlagen` (drittes Fehlversuchs-Inkrement), `fehlgeschlagen → offen` (erneut versuchen, `versuche = 0`), `fehlgeschlagen → verworfen`. Einträge werden nie gelöscht. Fehlversuchs-Limit = 3 als Backend-Konstante.
  - Die neuen Admin-Endpoints für Druckaufträge bilden ein eigenes Modul `backend/api/druckauftrag/{application,http}` nach dem bestehenden Modulmuster (ein Modul pro Ressource).
- **Relay-Zyklus**: Aufträge pro Zyklus nach Ziel-IP gruppiert, Gruppen parallel; innerhalb einer IP bleibt die ID-Reihenfolge erhalten. Genau ein Zustellversuch pro Auftrag und Zyklus mit kurzem TCP-Timeout, kein Sleep-Retry. Erstfehler einer IP überspringt die übrigen Aufträge dieser IP im Zyklus (übersprungene zählen nicht als Fehlversuch). At-least-once bleibt: erst drucken, dann melden.

## Inventory

**Schema & Queries**

- `database/migrations/01_initial.up.sql:56` — `ProduktKategorie`-Enum (bleibt für Produkte)
- `database/migrations/01_initial.up.sql:222-237` — `druckstationen` (PK `ProduktKategorie`, drei Seed-Zeilen, `bonmodus NOT NULL`)
- `database/migrations/01_initial.up.sql:242-258` — `druckauftraege` (Status nur `offen|gedruckt`)
- `database/migrations/01_initial.up.sql:391-407` — `bondruck_einstellungen` (Singleton, entfällt)
- `database/migrations/01_initial.down.sql:4-5` — Drop-Reihenfolge für Dev-Reset
- `backend/sqlc/queries/druckstation.sql:1-16` — Get/GetKonfigurierte/Upsert
- `backend/sqlc/queries/relay.sql:1-15` — Insert/GetOffene/MarkGedruckt (`AND status = 'offen'`-Guard vorhanden)
- `backend/sqlc/queries/bondruck_einstellungen.sql` — entfällt

**Backend**

- `backend/domain/druckstation/druckstation.go:1-7` — untypisiertes Struct ohne Validierung
- `backend/domain/settings/bondruck_einstellungen.go` — Domain-Typ mit Konstruktor-Validierung (Muster-Vorlage; entfällt inkl. Test)
- `backend/api/bondruck/application/arbeitsbon_policy.go:15-30` — Policy-Typen inkl. `DirektverkaufBondruckKonfiguration`
- `backend/api/bondruck/application/arbeitsbon_policy.go:43-92` — Event-Switch + Direktverkauf-Modus-Routing
- `backend/api/bondruck/application/arbeitsbon_policy_test.go` — bestehende Policy-Unit-Tests (werden erweitert)
- `backend/api/direktverkauf/application/command.go:100-108` — Policy-Aufruf beim Direktverkauf
- `backend/api/direktverkauf/application/command.go:145-177` — `loadBondruckKonfiguration` (lädt Modus aus Settings) + `toNeuerDruckauftraege` (Mapping-Duplikat 1)
- `backend/api/table/application/command.go:480` — zweiter Policy-Aufruf (Bestellung) mit leerer Direktverkauf-Konfiguration
- `backend/api/table/application/kassenbeleg_command.go:220-227` — Kassenbeleg-IP aus `bondruck_einstellungen`, Fehler bei leerer IP
- `backend/api/table/application/errors.go:61-62` — `ErrKassenbelegDruckerNichtKonfiguriert` (Fehlersemantik bleibt)
- `backend/api/table/http/command_handler.go:440,470` — Error-Mapping `kassenbeleg_drucker_nicht_konfiguriert`
- `backend/api/table_bondruck_adapters.go:31-47` — Mapping-Duplikat 2 (Policy-Druckauftrag → Repo-Insert)
- `backend/api/druckstation/http/handler.go:70-80` — zog-Schema (drei Kategorien, Bonmodus required)
- `backend/api/druckstation/application/{command,query}.go` — Upsert/GetAlle
- `backend/repository/druckstation_repo/repo.go:22-63` — GetAlle/GetKonfigurierte/Upsert
- `backend/repository/druckauftrag_repo/repo.go:33-117` — Enqueue/GetOffene/Quittiere
- `backend/repository/druckauftrag_repo/repo_test.go` — Integrationstest-Muster (`//go:build integration`, echte DB)
- `backend/repository/settings_repo/repo.go:60-77`, `types.go:35-42` — Bondruck-Methoden (entfallen)
- `backend/api/relay.go:13-56` — Forwarding-Adapter + Routen `/poll`, `/quittieren`
- `backend/api/relay/application/{command,query}.go` — reine 1:1-Durchreich-Schichten (werden gestrafft)
- `backend/api/relay/http/handler.go:50-106` — Poll- und Quittieren-Handler
- `backend/api/relay/relay_integration_test.go` — bestehender Integrationstest des Relay-API
- `backend/api/admin.go:118-124,136,144` — Admin-Routen-Verdrahtung
- `backend/api/settings/http/{query_handler.go:104-118,command_handler.go:40-115}` — Bondruck-Endpoints (entfallen)

**Relay**

- `cmd/relay/go.mod` — eigenes Go-Modul (`jotti-relay`), nicht in `make check`
- `cmd/relay/main.go:34` — `maxRetries = 60`
- `cmd/relay/main.go:112-151` — sequenzielle Hauptschleife, Quittierung nur für Erfolge
- `cmd/relay/main.go:153-174` — `printAuftragWithRetry` (60 × 5 s Sleep-Retry, blockiert alle anderen Aufträge)
- `cmd/relay/main.go:223-247` — `checkPrinter`: DLE-EOT-4-Statusbits (`0x40` = leer, `0x20` = fast leer — vermutlich falsche Bits)
- `cmd/relay/main_test.go` — testet ausschließlich Config-Parsing (größte Testlücke)
- `Makefile:198-209` — `check-backend`/`check-frontend`/`check`, Relay fehlt

**Frontend**

- `frontend/src/admin/settings/DruckstationConfigPage.tsx:22-131` — drei Zeilen, Speichern pro Zeile, keine Feldvalidierung
- `frontend/src/lib/DruckstationBackend.ts:5-37` — Zod-Schemas; `parse()` in `updateDruckstation` wirft bei ungültiger IP (generischer Toast)
- `frontend/src/admin/settings/EinstellungenPage.tsx:202,385-520` — Bondruck-Sektion (Form, Modus-Select, konditionales IP-Feld; entfällt)
- `frontend/src/lib/EinstellungenBackend.ts:23-117` — `BondruckEinstellungenSchema` + zwei Methoden (entfallen)
- `frontend/src/admin/settings/hooks.ts:19-79` — `useDruckstationen`, `useBondruckEinstellungen`
- `frontend/src/routes.ts:104-109` — Route `/admin/druckstationen` (bleibt)
- `frontend/src/admin/AdminSidebar.tsx:61-62` — Navigationseintrag „Druckstationen“

**Referenzdokumente**

- `docs/handbuch.md:426,644-701,914` — Bondruck-Policy, §4.6 (Outbox-Statusmodell, `bondruck_einstellungen`, Direktverkauf-Modus), Relay-Beschreibung
- `docs/language.md:37,149-153,451-483` — Direktverkauf-Modus-Definition, Abholbon, Druckauftrag-Status (enthält veralteten Hinweis „transientes DTO“), Relay
- `docs/anforderungen.md:49,240` — K-12 Arbeitsbon, F-03 Kassenbeleg

## Resolved decisions

- **`/relay/ergebnis` ersetzt `/relay/quittieren`** (User-Entscheidung): Der Pfad wird umbenannt, da sich die Semantik ändert; das Relay wird in derselben Phase umgestellt. Pre-Release, keine Kompatibilitätsschicht.
- **Sieben Phasen** (User bestätigt): Direktverkauf-Ableitung (Phase 2) und Kassenbeleg-Umstellung inkl. Tabellen-Entfall (Phase 3) bleiben getrennt.
- **Neues Backend-Modul `backend/api/druckauftrag`** für die Verwaltung fehlgeschlagener Aufträge — folgt dem bestehenden Muster „ein Modul pro Ressource“ (abgeleitet aus der Codebasis, nicht im PRD spezifiziert).
- **Speichern pro Stationszeile bleibt** (wie heute, ein Upsert pro Kategorie) — das PRD ändert nur Felder, Texte und Validierung der Seite, nicht das Speichermodell.

## Open questions / Risks

- **Papierstatus-Bits (DLE EOT 4):** Die aktuelle Auswertung (`0x40` = Papier leer, `0x20` = fast leer) widerspricht vermutlich der ESC/POS-Spezifikation (Near-End-Bits `0x0C`, End-Bits `0x60`). In Phase 5 gegen die offizielle Epson-Dokumentation verifizieren (Websuche, AGENTS-Regel 13) und korrigieren.
- **Doppeldruck bleibt möglich** (bewusst, PRD): Schlägt die Ergebnis-Meldung nach erfolgreichem Druck fehl, wird der Bon im nächsten Zyklus erneut gedruckt. Für Arbeitsbons unkritisch, für Kassenbelege wie bisher akzeptiert.
- **Verlorene Fähigkeit** (bewusst, PRD): Mit konfigurierten Produktstationen lassen sich Direktverkauf-Bons nicht mehr unterdrücken. Rückfalloption wäre ein Modus-Feld an der Abholbon-Zeile.

---

## Phase 1: Fünf Druckstationen konfigurierbar

**User stories**: 1, 2, 5, 7, 8 (+ Konfigurations-Teil von 3 und 4)

### Context

- `database/migrations/01_initial.up.sql:222-237` — Tabelle bekommt neuen PK-Typ, NULLable Bonmodus, fünf Seeds
- `backend/domain/druckstation/druckstation.go:1-7` — bekommt typisierte Werte + Konstruktor-Validierung (Muster: `backend/domain/settings/bondruck_einstellungen.go`)
- `backend/api/druckstation/http/handler.go:70-80` — zog-Schema auf fünf Kategorien und kategorieabhängigen Bonmodus erweitern
- `backend/repository/druckstation_repo/repo.go:22-63` + `backend/sqlc/queries/druckstation.sql` — NULL-Bonmodus durchreichen
- `frontend/src/admin/settings/DruckstationConfigPage.tsx:22-131` — fünf Zeilen, Erklärtexte, Inline-Validierung
- `frontend/src/lib/DruckstationBackend.ts:5-37` — Schema erweitern, `parse()`-Throw beim Speichern ersetzen

### What to build

Die Druckstationen-Tabelle trägt fünf Kategorien (neuer DB-Enum `DruckstationKategorie`, fünf Seed-Zeilen, `bonmodus` NULL nur für `kassenbeleg` per CHECK). Domain, Repository, zog-Schema und Admin-Endpoints akzeptieren alle fünf Kategorien; der Bonmodus ist für `essen`/`getraenk`/`sonstiges`/`abholbon` verpflichtend und für `kassenbeleg` unzulässig (Konstruktor-Validierung im Domain-Modell). Die Druckstationen-Seite zeigt fünf Zeilen — Bonmodus-Select bei allen außer Kassenbeleg — mit kurzen Erklärtexten je Station (insbesondere die Abholbon-Ableitungsregel) und Inline-IPv4-Feldvalidierung statt Schema-Throw mit generischem Toast. Die neuen Stationen haben in dieser Phase noch keine Routing-Wirkung; das bestehende Arbeitsbon-Verhalten bleibt unverändert.

### Acceptance criteria

- [x] `druckstationen` hat fünf Seed-Zeilen mit neuem Kategorie-Typ; CHECK erzwingt Bonmodus-Kopplung (essen/getraenk/sonstiges/abholbon gesetzt, kassenbeleg NULL); `01_initial.down.sql` räumt den neuen Typ mit ab
- [x] `/admin/get-druckstationen` liefert fünf Stationen; `/admin/update-druckstationen` akzeptiert alle fünf Kategorien und lehnt Bonmodus für kassenbeleg ab (zog + Domain-Konstruktor, Unit-Tests)
- [x] Druckstationen-Seite zeigt fünf Zeilen mit Erklärtexten; Bonmodus-Select bei Essen/Getränk/Sonstiges/Abholbon (nicht Kassenbeleg); Deaktivieren per Leeren der IP funktioniert
- [x] Ungültige IP erzeugt eine Feldmeldung am Eingabefeld, keinen generischen Speicherfehler
- [x] Bestehendes Arbeitsbon-Routing unverändert (bestehende Policy- und Handler-Tests grün)
- [x] `make check` und `make check-integration` grün

---

## Phase 2: Direktverkauf-Routing per Ableitung

**User stories**: 9, 10, 11 (+ 6 anteilig)

### Context

- `backend/api/bondruck/application/arbeitsbon_policy.go:27-92` — Modus-Switch wird durch Ableitungsregel ersetzt
- `backend/api/bondruck/application/arbeitsbon_policy_test.go` — Prior Art für die erweiterten Routing-Tests
- `backend/api/direktverkauf/application/command.go:100-108,145-177` — `loadBondruckKonfiguration` vereinfachen, Mapping-Duplikat 1
- `backend/api/table/application/command.go:480` — Aufruf ohne Direktverkauf-Konfiguration anpassen
- `backend/api/table_bondruck_adapters.go:31-47` — Mapping-Duplikat 2
- `backend/repository/druckauftrag_repo/repo.go:11-16` — `NeuerDruckauftrag` als Kandidat für den gemeinsamen Typ

### What to build

Die Policy verliert die `DirektverkaufBondruckKonfiguration`: Eingabe ist nur noch das Event plus die Map aller konfigurierten Druckstationen. Für `direktverkauf-getaetigt` gilt die Ableitungsregel — Abholbon-Station konfiguriert → Abholbon(s) an diese Station gemäß ihrem Bonmodus (`pro_position` = ein Abholbon je Position, `pro_bestellung` = ein Sammel-Abholbon); sonst Produktstationen je Kategorie (inkl. Bonmodus-Gruppierung); ohne konfigurierte Stationen entstehen keine Aufträge. `bestellung-aufgenommen` bleibt unverändert. Der Direktverkauf-Command lädt nur noch Druckstationen (kein Settings-Zugriff für den Modus mehr). Im selben Zug werden die doppelten 1:1-Mappings zwischen Policy-Druckauftrag und Repository-Insert-Typ auf einen gemeinsamen Typ bzw. einen Helfer reduziert. Die Policy-Unit-Tests decken alle Routing-Zweige ab.

### Acceptance criteria

- [x] Policy-Signatur ohne `DirektverkaufBondruckKonfiguration`; die Policy importiert kein `settings`-Paket mehr
- [x] Direktverkauf erzeugt Abholbon(s) gemäß Abholbon-Bonmodus, wenn Abholbon-Station konfiguriert; sonst Aufträge an Produktstationen; ohne Stationen keine Aufträge — Unit-Tests für alle Zweige inkl. Bonmodus-Gruppierung (auch Abholbon) und leerer Konfiguration
- [x] Direktverkauf schließt ohne einen einzigen konfigurierten Drucker erfolgreich ab
- [x] Die 1:1-Mappings (Direktverkauf-Command, Table-Adapter) sind auf einen gemeinsamen Typ/Helfer reduziert
- [x] `make check` und `make check-integration` grün

---

## Phase 3: Kassenbeleg über Druckstation, `bondruck_einstellungen` entfällt

**User stories**: 3, 4, 6, 21

### Context

- `backend/api/table/application/kassenbeleg_command.go:220-227` — bezieht IP künftig aus der Kassenbeleg-Druckstation
- `backend/api/table/application/errors.go:61-62` + `backend/api/table/http/command_handler.go:440,470` — Fehlersemantik bleibt
- `database/migrations/01_initial.up.sql:391-407` — Tabelle + Seed entfallen
- `backend/sqlc/queries/bondruck_einstellungen.sql`, `backend/repository/settings_repo/repo.go:60-77`, `types.go:35-42` — entfallen
- `backend/domain/settings/bondruck_einstellungen.go` + Test — entfallen
- `backend/api/settings/{http,application}` — beide Bondruck-Endpoints inkl. Handler-Tests entfallen
- `backend/api/admin.go:136,144` — Routen-Verdrahtung entfernen
- `frontend/src/admin/settings/EinstellungenPage.tsx:202,385-520`, `frontend/src/lib/EinstellungenBackend.ts:23-117`, `frontend/src/admin/settings/hooks.ts:60-79` — Frontend-Sektion, Schema, Methoden, Hook entfallen

### What to build

Der Kassenbeleg-Command bezieht die Ziel-IP aus der Kassenbeleg-Druckstation statt aus den Bondruck-Einstellungen; die Fehlersemantik `kassenbeleg_drucker_nicht_konfiguriert` bei leerer IP bleibt erhalten. Danach wird `bondruck_einstellungen` über alle Schichten entfernt: Tabelle und Seed, sqlc-Queries, Domain-Typ samt Tests, Repository-Methoden, beide Admin-Endpoints samt Handler-Tests, Frontend-Schema, -Methoden, -Hook und die Bondruck-Sektion der Einstellungen-Seite. Die Druckstationen-Seite ist danach der einzige Konfigurationsort für den Druck.

### Acceptance criteria

- [x] Kassenbeleg-Druck nutzt die IP der Kassenbeleg-Druckstation; ohne Konfiguration kommt weiterhin `kassenbeleg_drucker_nicht_konfiguriert` (Servicekraft sieht die bestehende klare Meldung)
- [x] `bondruck_einstellungen` existiert in keiner Schicht mehr: kein Treffer für `bondruck_einstellungen`, `BondruckEinstellungen` oder `DirektverkaufModus` in `backend/`, `frontend/src/`, `database/`
- [x] Einstellungen-Seite ohne Bondruck-Sektion; übrige Sektionen unverändert
- [x] `make check` und `make check-integration` grün

---

## Phase 4: Outbox-Statusmodell & Relay-Ergebnis-Endpoint

**User stories**: 14, 15 (Backend-Hälfte; Relay meldet Fehlversuche erst ab Phase 5)

### Context

- `database/migrations/01_initial.up.sql:242-258` — Status-CHECK, `versuche`, `letzter_fehler`
- `backend/sqlc/queries/relay.sql:1-15` — neue Queries für Fehlversuchs-Inkrement
- `backend/repository/druckauftrag_repo/repo.go:93-117` + `repo_test.go` — Quittieren + Integrationstest-Muster
- `backend/api/relay/http/handler.go:84-106` — Quittieren-Handler wird Ergebnis-Handler
- `backend/api/relay/application/{command,query}.go` + `backend/api/relay.go:13-56` — reine Durchreich-Schichten straffen
- `backend/api/relay/relay_integration_test.go` — auf neuen Endpoint umstellen
- `cmd/relay/main.go:202-221` — `quittieren()` ruft den neuen Endpoint

### What to build

Die Outbox bekommt das erweiterte Statusmodell (`fehlgeschlagen`, `verworfen`) plus `versuche` und `letzter_fehler`. `POST /relay/ergebnis` ersetzt `/relay/quittieren` und nimmt pro Zyklus erfolgreiche IDs und Fehlversuche (ID + Fehlertext) in einem Request entgegen. Das Backend besitzt die Fehlversuchs-Logik: pro gemeldetem Fehlversuch `versuche + 1` und `letzter_fehler` aktualisieren; beim dritten Versuch (Backend-Konstante) wechselt der Auftrag auf `fehlgeschlagen` und wird nicht mehr ausgeliefert. Quittieren bleibt idempotent (Status-Guard `offen`). Bei der Erweiterung werden die reinen Forwarding-Schichten der Relay-API gestrafft (keine 1:1-Durchreich-Ebenen zwischen Handler und Repository mehr). Das Relay wird minimal angepasst: Es meldet seine Erfolge an den neuen Endpoint (Fehlversuche meldet erst der neue Zyklus in Phase 5); der alte Pfad entfällt.

### Acceptance criteria

- [x] `druckauftraege` trägt erweiterten Status-CHECK, `versuche` (Default 0) und `letzter_fehler` (NULL)
- [x] `/relay/ergebnis` verarbeitet Erfolge und Fehlversuche in einem Request; `/relay/quittieren` existiert nicht mehr (Backend und Relay)
- [x] Fehlversuchs-Zählung: 1. und 2. Meldung inkrementieren und aktualisieren `letzter_fehler`, die 3. setzt `fehlgeschlagen`; fehlgeschlagene Aufträge erscheinen nicht mehr im Poll — als Repository-Integrationstests gegen echte DB
- [x] Quittieren ist idempotent (doppelte Meldung einer ID ändert nichts) — Integrationstest
- [x] Handler-Tests (httptest, Mock-Repo) für den Ergebnis-Endpoint inkl. Token-Prüfung
- [x] Zwischen Relay-Handler und Repository liegt keine reine 1:1-Durchreich-Schicht mehr
- [x] `make check` und `make check-integration` grün

---

## Phase 5: Relay-Zyklus-Umbau

**User stories**: 12, 13, 22, 23 (vervollständigt 14, 15)

### Context

- `cmd/relay/main.go:34,112-174` — Sleep-Retry und sequenzielle Schleife entfallen
- `cmd/relay/main.go:223-247` — `checkPrinter` mit fraglichen DLE-EOT-4-Statusbits
- `cmd/relay/main_test.go` — bisher nur Config-Parsing getestet
- `Makefile:198-209` — neues `check-relay`-Target, in `check` einhängen

### What to build

Der Relay-Zyklus wird neu geschnitten: Pro Poll werden die Aufträge nach Ziel-IP gruppiert; die Gruppen laufen parallel (ein toter Drucker blockiert keinen anderen), innerhalb einer IP bleibt die ID-Reihenfolge erhalten. Pro Auftrag genau ein Zustellversuch mit kurzem TCP-Timeout — das 60×5s-Sleep-Retry entfällt vollständig. Schlägt der erste Verbindungsaufbau zu einer IP fehl, werden die übrigen Aufträge dieser IP im Zyklus übersprungen (kein Fehlversuch, keine Meldung). Am Zyklusende meldet das Relay Erfolge und Fehlversuche gesammelt an `/relay/ergebnis`. Die Zyklus-Logik wird testbar geschnitten (Druck- und Melde-Funktionen injizierbar) und mit Unit-Tests abgedeckt. Die Papierstatus-Auswertung wird gegen die ESC/POS-Dokumentation verifiziert und korrigiert (Near-End- vs. End-Bits). Das Relay-Modul wird in `make check` aufgenommen (Lint, Vet, Tests, Build).

### Acceptance criteria

- [x] Pro Auftrag genau ein Zustellversuch pro Zyklus; `maxRetries`/Sleep-Retry sind entfernt
- [x] Aufträge nach Ziel-IP gruppiert, Gruppen parallel; ein nicht erreichbarer Drucker verzögert andere Drucker nur um einen kurzen Verbindungsversuch
- [x] Nach Erstfehler einer IP werden deren übrige Aufträge im Zyklus übersprungen und nicht als Fehlversuch gemeldet
- [x] Erfolge und Fehlversuche gehen gesammelt in einem `/relay/ergebnis`-Request ans Backend
- [x] Unit-Tests mit injizierten Fake-Funktionen: Gruppierung, Reihenfolge innerhalb einer IP, Skip nach Erstfehler, korrekte Erfolg-/Fehlversuchs-Meldung
- [x] Papierstatus-Bits gegen die ESC/POS-Dokumentation verifiziert und ggf. korrigiert (mit Quellenangabe im Commit/Kommentar)
- [x] `make check` prüft das Relay-Modul mit (Lint, Vet, Tests, Build)

---

## Phase 6: Fehlgeschlagene Aufträge sichtbar & verwaltbar

**User stories**: 16, 17, 18, 19, 20

### Context

- `backend/api/admin.go:118-124` — Verdrahtungsmuster für neue admin-only Endpoints
- `backend/api/druckstation/{http,application}` — Modulmuster für das neue Modul `backend/api/druckauftrag`
- `backend/repository/druckauftrag_repo/repo.go` + `repo_test.go` — neue Methoden + Integrationstests
- `frontend/src/admin/settings/DruckstationConfigPage.tsx` — Seite bekommt den Abschnitt „Fehlgeschlagene Druckaufträge“
- `frontend/src/lib/DruckstationBackend.ts` + `frontend/src/admin/settings/hooks.ts` — Backend-Klasse und Hook erweitern

### What to build

Drei neue admin-only POST-Endpoints in einem neuen Modul `backend/api/druckauftrag`: fehlgeschlagene Aufträge auflisten (Bon-Art, Ziel-IP, Referenz, Zeitpunkt, Versuche, letzter Fehler), Auftrag erneut versuchen (`fehlgeschlagen → offen`, `versuche = 0`) und Auftrag verwerfen (`fehlgeschlagen → verworfen`). Die Übergänge sind status-geschützt (nur aus `fehlgeschlagen`); Einträge werden nie gelöscht. Auf der Druckstationen-Seite erscheint darunter der Abschnitt „Fehlgeschlagene Druckaufträge“ mit Liste und den Aktionen „Erneut versuchen“ und „Verwerfen“ sowie einem leeren Zustand.

### Acceptance criteria

- [ ] `/admin/get-fehlgeschlagene-druckauftraege` listet ausschließlich fehlgeschlagene Aufträge mit Bon-Art, Ziel-IP, Referenz, Zeitpunkt, Versuchen und letztem Fehler
- [ ] „Erneut versuchen“ setzt den Auftrag auf `offen` mit `versuche = 0`; er erscheint wieder im Relay-Poll — Repository-Integrationstest
- [ ] „Verwerfen“ setzt `verworfen`; der Auftrag bleibt in der Datenbank erhalten — Repository-Integrationstest
- [ ] Beide Übergänge wirken nur auf Aufträge im Status `fehlgeschlagen` (Status-Guard, Integrationstest)
- [ ] Handler-Tests (httptest, Mock-Repo) für alle drei Endpoints; Endpoints nur für Admins erreichbar
- [ ] Druckstationen-Seite zeigt die Liste mit beiden Aktionen und einem leeren Zustand; Aufrufe laufen über die erweiterte `DruckstationBackend`-Klasse
- [ ] `make check` und `make check-integration` grün

---

## Phase 7: Referenzdokumente angleichen

**User stories**: — (Aufräumpunkt aus dem PRD)

### Context (veraltet)

- `docs/handbuch.md — Policy-Beschreibung, §4.6 Bondruck, Relay-Poll
- `docs/language.md` — Direktverkauf-Modus, Abholbon, Druckauftrag (inkl. veraltetem Hinweis „transientes DTO“), Relay
- `docs/anforderungen.md` — K-12, F-03

### What to build

Die Referenzdokumente werden an das neue Modell angepasst: fünf Druckstationen als einziger Konfigurationsort, Abholbon-Ableitungsregel statt Direktverkauf-Modus, Outbox-Statusmodell mit `fehlgeschlagen`/`verworfen` und Fehlversuchs-Zählung, `/relay/ergebnis` statt `/relay/quittieren`, Wegfall von `bondruck_einstellungen`. Veraltete Aussagen (z. B. „druckauftraege geplant; aktuell transientes DTO“ in language.md) werden dabei korrigiert.

### Acceptance criteria

- [ ] `docs/handbuch.md` §4.6 beschreibt fünf Stationen, Ableitungsregel, Statusmodell und Ergebnis-Endpoint; keine Erwähnung von `bondruck_einstellungen` oder `direktverkauf_modus` mehr
- [ ] `docs/language.md`: Direktverkauf-Modus-Begriff entfernt bzw. durch Ableitungsregel ersetzt, Abholbon- und Druckauftrag-Definitionen aktualisiert, veraltete Relay-/Outbox-Hinweise korrigiert
- [ ] `docs/anforderungen.md` K-12/F-03 gegengeprüft und wo nötig angepasst
- [ ] Außerhalb von `docs/prds/` und `docs/plans/` (historische Artefakte) verweist kein Dokument mehr auf das alte Modell
