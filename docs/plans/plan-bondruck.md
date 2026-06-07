# Plan: Bondruck — Neuordnung (Arbeitsbon, Kassenbeleg, Outbox)

> Source PRD: `docs/prds/prd-bondruck.md`
> Verwandt: `docs/prds/prd-direktverkauf.md` (Direktverkauf-Bons konsumieren diese Infrastruktur)

## Goal

Die Bon-Domäne wird entlang **zweier Familien** neu geordnet — operativer
**Arbeitsbon** (nicht-fiskalisch, automatisch) und gesetzlicher **Kassenbeleg**
(fiskalisch, auf Anforderung) — auf einer gemeinsamen **Druckauftrags-Outbox**. Der
Relay wird zum reinen Transport. `kategorie_drucker` wird zu `druckstationen`. Die
falsche F-03-Behauptung (Auto-Beleg nach Zahlung) wird korrigiert.

Dieser Plan ist **self-contained** (bedienter Tisch + Druck-Infrastruktur). Die
Direktverkauf-Bons (Abholbon, Stations-Routing, Direktverkauf-Kassenbeleg) sind **nicht**
Teil dieses Plans — sie konsumieren diese Infrastruktur und werden mit dem
Direktverkauf-Feature (`prd-direktverkauf.md`) umgesetzt.

**Empfohlene Reihenfolge:** 0 → 1 → 2 → 3 → 4.

## Architectural decisions

Durable-Entscheidungen, die für alle Phasen gelten:

- **Zwei Bon-Familien, eine Outbox.** Arbeitsbon und Kassenbeleg teilen keinen Trigger,
  Inhalt, Formatter oder Rechtsstatus — nur die `druckauftraege`-Outbox als Transport.
- **ESC/POS-Formatierung lebt im Backend.** Sie wird aus `backend/api/relay/application/escpos`
  in den neuen Bondruck-Bereich verschoben. Der Relay formatiert nichts mehr.
- **Outbox ist Single Source of Truth für Druckjobs.** Tabelle `druckauftraege`
  (`id`, `ziel_ip`, `payload` Base64-ESC/POS, `status` `offen|gedruckt`,
  `bon_art` `arbeitsbon|kassenbeleg`, `referenz`, `erstellt_am`, `gedruckt_am`). Einziger
  Übergang: `offen → gedruckt`. Die Outbox ist **kein** fiskalisches Journal — sie ist
  eine technische Warteschlange.
- **Relay = Transport.** `POST /relay/poll` liefert offene Jobs; neuer
  `POST /relay/quittieren` bestätigt gedruckte IDs. Kein Cursor, keine lokale
  State-Datei, keine Fachlogik. Drucker-Erreichbarkeit/Retry bleibt im Relay.
- **Schema-Änderungen direkt in `database/migrations/01_initial.up.sql`** (aktive
  Entwicklungsphase, keine neuen Migrationsdateien). Dev-Reset: `make down && make dev`.
  `make sqlc` nach jeder Query-/Schema-Änderung.
- **`ProduktKategorie`-Enum bleibt unverändert.** Nur `kategorie_drucker` → `druckstationen`.
- **POST-only, Geld in Cent, beidseitige Validierung (zog + Zod).** Response-DTOs in
  der HTTP-Schicht; Domain-Modelle nie direkt serialisiert.
- **Konfiguration getrennt:** `druckstationen` (3 Kategorie-Zeilen, Arbeitsbon-Stationen)
  vs. `bondruck_einstellungen` (Singleton: Kassenbeleg-Drucker). Den Direktverkauf-Modus
  - Abholbon-Drucker ergänzt später der Direktverkauf (`prd-direktverkauf.md`).

## Inventory

Verifizierte, betroffene Stellen:

**Relay (wird zum Transport umgebaut)**

- `backend/sqlc/queries/relay.sql:1-7` — `GetBestellungEventsSinceCursor` (entfällt;
  ersetzt durch Outbox-Queries)
- `backend/api/relay/application/query.go:34-66` — `GetDruckAuftraege` (compute-at-poll;
  wird durch „offene Jobs lesen" ersetzt)
- `backend/api/relay/application/print.go:30-92` — `createDruckAuftraegeFromEvent`,
  `parseTischName` (Formatier-/Gruppierlogik → wandert in die Arbeitsbon-Policy)
- `backend/api/relay/application/escpos/formatter.go`, `constants.go`,
  `formatter_test.go` — ESC/POS (→ Backend-Bondruck-Bereich; Relay nutzt es nicht mehr)
- `backend/api/relay/http/handler.go:33-71` — `PollHandler` (Semantik neu) + neuer
  Quittieren-Handler
- `backend/api/relay.go:37-54` — `NewRelayApi`, `druckerRepoRelayAdapter` (Adapter
  entfällt, da Relay keine Druckerkonfig mehr braucht)
- `cmd/relay/main.go:18-31` (`RelayState`), `:106-152` (`poll`, `printAuftragWithRetry`)
  — Cursor/State entfällt; poll-Antwort = Jobliste; nach Druck `quittieren`

**Druckerkonfiguration (Rename → Druckstation)**

- `database/migrations/01_initial.up.sql:214-231` — `CREATE TABLE kategorie_drucker`
  - Seed (3 Zeilen) → `druckstationen`
- `backend/sqlc/queries/` — Drucker-Queries (`GetKategorieDrucker`,
  `GetKonfigurierteKategorieDrucker`, `UpsertKategorieDrucker`)
- `backend/repository/drucker_repo/repo.go:11-70`, `mock.go`, `repo_test.go` →
  `druckstation_repo`
- `backend/api/drucker/application/{command,query,errors}.go`,
  `backend/api/drucker/http/handler.go:13-110` → `druckstation`
- `backend/api/admin.go:111-117` — Routen `/get-drucker-konfiguration`,
  `/update-drucker-konfiguration` → `…-druckstationen`
- `frontend/src/lib/DruckerBackend.ts:1-41` → `DruckstationBackend`
- `frontend/src/admin/settings/DruckerConfigPage.tsx`, `hooks` →
  `DruckstationConfigPage`

**Kassenbeleg-Datenquellen (vorhanden, werden gelesen)**

- `backend/domain/settings/betreiber.go:9-17` — Vereinsname/Adresse/Steuernummer (K-20)
- `backend/domain/settings/kassenidentitaet.go:9-12` — Seriennummer (F-01)
- `backend/domain/kasse/tisch_session_events.go:37-47` — `zahlungKassiertV1Data`
  (`Positionen`, `GesamtZahlungCents`, `Kommentar`) — Datenbasis des Belegs
- `backend/domain/kasse/bestellung.go:10-63` — `Position` (**kein** `Steuersatz` →
  Steuer-Aufschlüsselung blockiert auf F-07)
- `backend/api/service.go:42-54` — Service-Routen (neuer `/beleg-drucken`)

**Write-Through / Event-Write (für Enqueue-Kopplung)**

- `backend/api/table/application/command.go:25,80-88` — `WriteEvent(ctx, e, streamType,
kassensitzungNr)`, `writeEvent`-Helfer
- `docs/handbuch.md:327-336` — StreamType-Routing / Write-Through-Muster

**Domänen-Doku**

- `docs/handbuch.md:418-421` (§3.12 Policy), `:651-700` (§4.6 Bondruck) — Phase 4
- `docs/compliance.md:5.x, 7.x` (Belegausgabepflicht, processTypes) — Phase 4
- `docs/datenbankschema.md:258` (`kategorie_drucker` → `druckstationen`) — Phase 1
- `docs/language.md` (Bondruck-Vokabular + Ist/Soll) — **Phase 0 erledigt**
- `docs/anforderungen.md` (K-12 Arbeitsbon, F-03 korrigiert) — **Phase 0 erledigt**

## Resolved decisions

- **Outbox-Architektur** statt compute-at-poll (Nutzer-Entscheidung) — vereinheitlicht
  automatisch + auf Anforderung, Status in der DB, Payload eingefroren.
- **Kassenbeleg jetzt** als Basis-Beleg (Preise + Vereinsdaten + Kassen-ID); Steuer
  (F-07) und TSE (F-02) als ausgewiesene Folgeschritte.
- **Voller Rename** `kategorie_drucker` → `druckstationen` inkl. Paket/DTO/Route/Frontend.
- **Kassenbeleg pro Kassiervorgang** (eine `zahlung-kassiert`-Zahlung), nicht pro Tisch.
- **`bondruck_einstellungen` als Singleton** (Muster wie `betreiber`/`kassenidentitaet`).

## Open decisions — vor Umsetzung der jeweiligen Phase bestätigen

> **Enqueue-Transaktionalität (Phase 2, zentral):** Arbeitsbons werden vom
> Application-Command eingereiht.
> **(a) Post-Commit-Enqueue** _(empfohlen)_ — der Command schreibt das Event (bestehende
> `WriteEvent`-Transaktion) und reiht **danach** die Druckaufträge in einem separaten
> Insert ein. Vorteil: `WriteEvent`-Signatur und Repository bleiben unberührt, Bondruck
> bleibt ein sauber getrenntes, testbares Modul. Risiko: „Event committed, Enqueue
> schlägt fehl" → fehlender Arbeitsbon (operativ, nachdruckbar) — akzeptabel, da der
> Arbeitsbon nicht-fiskalisch ist.
> **(b) In-Transaktion-Enqueue** — der Command übergibt vorformatierte Druckaufträge an
> `WriteEvent`, das sie atomar mit dem Event-INSERT schreibt. Vorteil: „bestellt ⇒ wird
> gedruckt" garantiert. Nachteil: `WriteEvent`-Signatur ändert sich, Druckjobs werden
> durch jeden Write gefädelt.
> Der **Kassenbeleg** ist davon unberührt — er wird ohnehin durch einen **eigenen**
> On-Demand-Command in einem eigenen Insert eingereiht.

> **Reprint-Fenster im Relay (Phase 2):** DB-Status ist autoritativ. Bei Relay-Absturz
> zwischen Druck und Quittierung kann ein Job erneut gedruckt werden (gleiche
> Risikoklasse wie heute). Optionaler lokaler Idempotenz-Cache im Relay als „defense in
> depth" — Default: weglassen (einfacher), DB-Status genügt.

> **Bon-Nummer auf dem Kassenbeleg (Phase 3):** Event-ID des Kassiervorgangs als
> Bon-Nummer verwenden _(empfohlen, vorhanden, eindeutig)_ vs. separater Zähler. Bei
> TSE-Anbindung (F-02) wird ohnehin die TSE-Transaktionsnummer maßgeblich.

## Risiken

- **Phase 2** ist der invasivste Umbau (Relay-Protokoll + Outbox + Verschiebung der
  ESC/POS-Logik). Absicherung: Tisch-Arbeitsbon-Verhalten muss vor/nach identisch sein
  (gleicher ESC/POS-Output); Integrationstest poll→quittieren→poll.
- **Relay ist ein eigenes Go-Modul** (`cmd/relay/go.mod`). Protokolländerung Backend↔Relay
  muss synchron erfolgen; alter Relay ist nach dem Umbau inkompatibel (Breaking Change,
  laut `AGENTS.md` erwünscht).
- **F-03-Korrektur** berührt Compliance-Doku — sorgfältig formulieren (kein Auto-Beleg,
  aber On-Demand-Beleg + Pflichtfelder), um keine neue Fehlinterpretation zu erzeugen.
- **Direktverkauf-Bons** (Abholbon, Stations-Routing, Direktverkauf-Kassenbeleg) sind
  **nicht** Teil dieses Plans — sie werden mit dem Direktverkauf-Feature umgesetzt und
  konsumieren diese Infrastruktur.

---

## Phase 0: Dokumentation & Vokabular (erledigt)

**User stories**: 23, 24

### Context

- `docs/language.md` — Bondruck-Vokabular + Ist/Soll-Abweichungen
- `docs/anforderungen.md` — K-12 (Arbeitsbon), F-03 (Belegausgabepflicht)
- `docs/prds/prd-bondruck.md`, `docs/prds/prd-direktverkauf.md` — Scope-Schnitt

### What was built

Vor jeder Code-Zeile: die Wahrheits- und Vokabular-Korrekturen, die unabhängig von der
Implementierung gelten. `language.md` führt die kanonischen Begriffe ein (Arbeitsbon,
Kassenbeleg, Druckstation, Druckauftrag) und dokumentiert die offenen Soll-Abweichungen
(Rename/Outbox/Relay-Transport). `anforderungen.md`: K-12 → „Arbeitsbon" geschärft;
**F-03 korrigiert** (die falsche „Auto-Beleg nach Zahlung"-Behauptung entfernt, Status
auf „Offen", On-Demand-Kassenbeleg-Kriterien). Die Teil-2-Inhalte (Direktverkauf-Bons)
wurden aus `prd-bondruck.md` nach `prd-direktverkauf.md` migriert; beide PRDs sind jetzt
self-contained.

Die **Architektur-Beschreibungen** (handbuch §4.6/§3.12, compliance) werden **bewusst
nicht vorab** geschrieben — sie kommen mit dem jeweiligen Slice (Phase 4), damit die Doku
kein ungebautes System beschreibt (dieselbe Schein-Compliance, die F-03 hatte).

### Acceptance criteria

- [x] `language.md` enthält Arbeitsbon, Kassenbeleg, Druckstation, Druckauftrag + Ist/Soll
- [x] `anforderungen.md` F-03 ohne falsche Auto-Beleg-Behauptung; K-12 = Arbeitsbon
- [x] `prd-bondruck.md` self-contained (Teil 1); Direktverkauf-Bons in `prd-direktverkauf.md`

---

## Phase 1: Druckstation-Rename

**User stories**: 3, 4, 5, 20, 28

### Context

- `database/migrations/01_initial.up.sql:214-231` — `kategorie_drucker` + Seed
- `backend/repository/drucker_repo/repo.go:11-70` — Repo + `DruckerKonfig`
- `backend/api/drucker/{application,http}` — Command/Query/Handler + DTOs
- `backend/api/admin.go:111-117` — Drucker-Routen
- `frontend/src/lib/DruckerBackend.ts:1-41`, `frontend/src/admin/settings/DruckerConfigPage.tsx`
- `docs/language.md:428-444` — Begriffe

### What to build

Reiner Rename ohne Verhaltensänderung: `kategorie_drucker` → `druckstationen`
(Tabelle + Seed + Queries + `dbgen`), `drucker_repo` → `druckstation_repo`, das
`drucker`-API-Paket → `druckstation`, Admin-Routen `…-druckstationen`, Frontend-Klasse
`DruckstationBackend` und Seite `DruckstationConfigPage`. Das Vokabular ist in Phase 0
gesetzt; hier wird die `language.md`-Soll-Abweichung „Druckstation" geschlossen (Ist =
`druckstationen`). Die `ProduktKategorie`-Enum und der `bonmodus` bleiben.
Tisch-Bestellung druckt danach exakt wie zuvor (über den noch unveränderten Relay-Pfad).

### Acceptance criteria

- [ ] `druckstationen`-Tabelle ersetzt `kategorie_drucker` (3 Seed-Zeilen unverändert)
- [ ] `make sqlc` regeneriert ohne manuelle `dbgen/`-Eingriffe; Build grün
- [ ] Admin-Konfigseite lädt/speichert Druckstationen unverändert (Verhalten identisch)
- [ ] `language.md`-Soll-Abweichung „Druckstation" geschlossen (Ist = `druckstationen`)
- [ ] `make check` grün; keine verwaisten `drucker`-Bezeichner mehr

---

## Phase 2: Druckauftrags-Outbox + Relay = Transport

**User stories**: 1, 2, 6, 7, 15, 16, 17, 18, 19

### Context

- `backend/sqlc/queries/relay.sql:1-7` — alter compute-at-poll-Query (entfällt)
- `backend/api/relay/application/{query.go:34-66,print.go:30-92}` — Logik → Policy
- `backend/api/relay/application/escpos/*` — ESC/POS (→ Backend-Bondruck-Bereich)
- `backend/api/relay/http/handler.go:33-71` — poll (neu) + quittieren (neu)
- `cmd/relay/main.go:18-31,106-152` — Cursor/State entfällt; poll→drucken→quittieren
- `backend/api/table/application/command.go:25,80-88` — Event-Write (Enqueue-Kopplung)

### What to build

Die `druckauftraege`-Outbox als Single Source of Truth. Eine **Arbeitsbon-Policy** im
Backend reiht beim `bestellung-aufgenommen:v1` Druckaufträge ein (übernimmt die
Gruppier-/Formatierlogik aus `print.go` + `escpos`, jetzt im Backend). Der Relay holt
per `POST /relay/poll` offene Aufträge (`id`, `zielIp`, `payload`), druckt mit
bestehender Retry-Logik und bestätigt per `POST /relay/quittieren`; das Backend setzt
`status = 'gedruckt'`. Cursor und lokale State-Datei entfallen. Tisch-Arbeitsbons
drucken identisch zu Phase 1, nur jetzt über die Outbox. Enqueue-Variante (a/b) gemäß
Open Decision.

### Acceptance criteria

- [ ] `druckauftraege`-Tabelle existiert; Repo kann einreihen, offene lesen, quittieren
- [ ] Bestellung erzeugt korrekte Outbox-Zeilen je Kategorie/Modus; Kategorie ohne
      Druckstation → keine Zeile
- [ ] ESC/POS-Output der Arbeitsbons ist byte-identisch zum bisherigen Bon (Regression)
- [ ] `poll` liefert nur `offen`; nach `quittieren` liefert ein erneuter `poll` sie nicht
- [ ] Relay enthält keine ESC/POS-/Kategorie-/Cursor-Logik mehr (reiner Transport)
- [ ] Integrationstest: Bestellung → Outbox → poll → quittieren → poll (leer)
- [ ] `make verify` grün

---

## Phase 3: Kassenbeleg auf Anforderung (Basis)

**User stories**: 8, 9, 10, 11, 12, 13, 14, 21, 22, 26, 27

### Context

- `backend/domain/settings/betreiber.go:9-17`, `kassenidentitaet.go:9-12` — Beleginhalt
- `backend/domain/kasse/tisch_session_events.go:37-47` — Zahlungspositionen (Fat Event)
- `backend/domain/kasse/bestellung.go:10-63` — `Position` (kein `Steuersatz` → F-07)
- `backend/api/service.go:42-54` — Service-Routen (neuer `/beleg-drucken`)
- Outbox aus Phase 2 — Kassenbeleg wird als Druckauftrag eingereiht

### What to build

`bondruck_einstellungen` (Singleton) mit `kassenbeleg_drucker_ip` + Admin-Konfig
(get/update, beidseitig validiert). Ein **Kassenbeleg-ESC/POS-Formatter** (Vereinsname

- Adresse, Kassen-Seriennummer, Datum/Uhrzeit, Positionen mit Einzelpreis × Menge,
  Gesamtbetrag, „bar", Bon-Nummer). Ein Service-Command `KassenbelegDrucken` nimmt eine
  Referenz auf **einen Kassiervorgang** (`zahlung-kassiert`-Zahlung), lädt Positionen +
  Betreiber + Kassenidentität, formatiert und reiht **einen** Druckauftrag für den
  Kassenbeleg-Drucker ein. Endpunkt `POST /service/beleg-drucken`. Frontend: „Beleg
  drucken"-Aktion nach dem Kassieren bzw. in der Tisch-Historie pro Zahlung. Fehlender
  Drucker → klarer Fehler. Steuer/TSE als TODO im Formatter ausgewiesen (F-07/F-02).

### Acceptance criteria

- [ ] `bondruck_einstellungen`-Singleton + Admin-Konfig (Kassenbeleg-Drucker), validiert
- [ ] `POST /service/beleg-drucken` erzeugt genau einen Druckauftrag für einen echten
      Kassiervorgang
- [ ] Beleg enthält Vereinsdaten, Kassen-Seriennummer, Positionen mit Preisen,
      Gesamtbetrag, Zahlungsart „bar", Bon-Nummer
- [ ] Fehlender Kassenbeleg-Drucker → klare Fehlermeldung (kein stilles Scheitern)
- [ ] Erneuter Aufruf druckt erneut (Nachdruck), ohne den Kassiervorgang zu wiederholen
- [ ] Frontend „Beleg drucken" nur über `BackendClient`; löst genau einen Aufruf aus
- [ ] `make verify` grün

---

## Phase 4: Architektur-Dokumentation (Handbuch & Compliance)

**User stories**: 23, 25

### Context

- `docs/handbuch.md` — §3.12 (Policies), §4.6 (Bondruck)
- `docs/compliance.md` — Belegausgabepflicht (§ 3.4, § 5), processTypes
- (F-03, K-12, `language.md`, PRD-Scope: bereits in **Phase 0** erledigt)

### What to build

Die Architektur-Doku wird **nach** dem Bau (Phasen 1–3) an den Ist-Zustand angeglichen —
nicht vorab, damit sie kein ungebautes System beschreibt. `handbuch.md` §4.6 neu
schreiben (zwei Familien, Outbox, Relay = Transport); §3.12 Policy „Arbeitsbon-Druck";
Outbox-Tabelle dokumentieren; `kategorie_drucker`-Beschreibung → `druckstationen`.
`compliance.md`: Arbeitsbon ausdrücklich als nicht-fiskalisch kennzeichnen, Kassenbeleg =
§ 146a-Beleg, Belegausgabe-Befreiung am Fest (Betreiberpflicht). In `language.md` die
offenen Soll-Abweichungen schließen, sobald der jeweilige Rename/Outbox steht.

### Acceptance criteria

- [ ] §4.6 beschreibt zwei Familien + Outbox + Relay-Transport; §3.12 nennt die Policy
- [ ] handbuch nennt `druckstationen` (nicht mehr `kategorie_drucker`)
- [ ] `compliance.md` trennt Arbeitsbon (nicht-fiskalisch) und Kassenbeleg (§ 146a)
      explizit; Befreiungshinweis vorhanden
- [ ] `language.md` Soll-Abweichungen für umgesetzte Teile geschlossen
- [ ] `make check` grün
