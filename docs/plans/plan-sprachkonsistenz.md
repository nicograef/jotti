# Plan: Sprachkonsistenz über die gesamte jotti-Codebase

> Source PRD: ../../PROMPT-sprachkonsistenz.md (Wegwerf-Prompt, nach Umsetzung löschen)

## Goal

Die Bezeichner-Sprache über Backend und Frontend an das Prinzip angleichen, das
der Kasse-Core bereits lebt: deutsche Fachverben für zustandsändernde
Domänen-Commands, englische Verben für Queries, Persistence und Infrastruktur,
deutsche Domänen-Nomen überall. Reiner Bezeichner- und Kommentar-Umbau, null
Verhaltensänderung. `make verify` bleibt nach jedem Subsystem grün.

Kein stumpfes Ersetzen einer Verb-Liste: Jede Umbenennung wird gegen das
Leitprinzip und die Nachbarschaft im Code geprüft. Dasselbe Verb ist je nach
Bounded Context und Schicht mal Fachbegriff (deutsch), mal Plumbing (englisch).

## Leitprinzip (Command/Query + Schicht)

Empirisch aus dem Kasse-Core abgeleitet, nicht neu erfunden. Reihenfolge der
Prüfung pro Bezeichner:

1. **Domänen-Nomen bleiben immer deutsch.** Nur der Verb- oder Muster-Teil steht
   je zur Debatte (`GetAdminPUK`, nicht `GetAdminPuk`).
2. **Zustandsändernder Domänen-Command → deutsches Fachverb.** Ein Vorgang, den
   Kassenwart, Servicekraft oder das Fiskalrecht benennen würde. Gilt in der
   application- und in der domain-nahen Worker-/Seed-Schicht.
   Vorbild im Code: `BestellungAufnehmen`, `ZahlungKassieren`,
   `StornierungErteilen`, `KasseAbschliessen`, `KassensitzungEroeffnen`,
   `GeldtransitBuchen`, `TischErstellen`/`TischAktualisieren`/`TischLoeschen`,
   `beginneStoerung`, `beendeStoerung`, `beschaffeSignatur`.
3. **Query / Derivation / Projektion → englisches Verb + deutsches Nomen.** Etwas
   berechnen, ableiten, lesen, bauen, ohne Zustandsänderung.
   Vorbild: `ComputeStornoAufteilung`, `ComputeNichtStorniertePositionen`,
   `BuildKassenbelegProcessData`, `GetHistorieFromEvents`, `ApplyEvent`,
   `checkSignaturGate` (bereits in `d0e3602` von `pruefeSignaturGate` übersetzt).
4. **Persistence (sqlc + Repository-Adapter) → englisches Verb**, auch bei
   Statuswechseln. Vorbild: `SetKassensitzungOffen`, `MarkDruckauftragGedruckt`,
   `Upsert…`, `Read…`, `Write…`, `Insert…`.
5. **Adapter / Integration / Infra / Tooling → englisches Verb.** Auth, Config,
   Advisory Lock, HTTP-Transport, Relay (reiner Transport), Seed-Builder/Writer,
   Test-Konstruktoren. Vorbild: `ensureLock`, `releaseLock`, `processOnce`.

Faustregel bei Zweifel: „Benennt das Fiskalrecht / DSFinV-K / ein Kassenwart
diesen Vorgang?" → deutsch. „Macht das jedes Programm unabhängig davon, dass es
eine Kasse ist (get, build, map, parse, persist, poll, retry)?" → englisch.

**Neue Kombinationen sind der Normalfall, nicht die Ausnahme:** englisches Verb
plus deutsches Domänen-Nomen (`ComputeAbschlussSummen`, `DetermineSignaturstatus`,
`ensureKeineOffeneKassensitzung`), oder deutsches Nomen plus englisches
Muster-Suffix (`sitzungsBuilder`).

## Domänen-Landkarte (Sprach-Haltung je Bounded Context)

- **Kasse (Core):** deutsche Fachsprache. Commands deutsch, Read-Model/Projektion
  englisch. Bereits konsistent, kaum Handlungsbedarf.
- **Fiskal / Compliance (TSE-Signierung, Störung, Beleg):** deutsche
  Gesetzessprache (language.md §Fiskalkonformität). Worker-Domänenintent deutsch
  (`beginneStoerung`, `quittieren`), Persistence/Derivation englisch.
- **Stammdaten (Supporting):** Commands deutsch (`TischErstellen`), Persistence
  englisch (`UpsertTischSession`). Nicht im Umbau-Scope, dient nur als Vorbild.
- **Fiskal-Setup / Integration:** Anbindung an fiskaly, überwiegend Adapter/
  Integration → englisch.
- **Reporting (Read Model):** reine Aggregation → englisch.
- **Auth / Infra (Generic):** vollständig englisch. Nicht im Scope.
- **Seed (Dev-Tooling):** Domänen-Simulatoren spiegeln die deutschen
  Kasse-Commands (`bestellen`, `kassieren`, `stornieren`, `schliesseTagAb`);
  Builder/Writer/Helfer sind Plumbing → englisch.
- **Druck-Outbox / Relay:** Warteschlangen- und Transport-Infrastruktur →
  englisch. Die Bons selbst (Arbeitsbon, Kassenbeleg) sind deutsche Nomen.

## Architectural decisions

- **sqlc-Prozedur** je Query-Umbenennung: (1) Query-Namen in
  `backend/sqlc/queries/*.sql` ändern, (2) `make sqlc`, (3) Querier-Interface,
  Repo-Wrapper, Aufrufer, Mocks/Fakes nachziehen. `backend/sqlc/dbgen/` nie von
  Hand editieren.
- **Bit-identische externe Verträge.** Unangetastet: Domänen-Nomen, UI-Texte,
  DB-Schema (Tabellen/Spalten/Enums), Routen-Strings (`tse-setup-pruefen`,
  `zahlung-kassieren`), Event-Typen, JSON-Feldnamen, API-Fehlercodes, Modulname
  `user`.
- **Dokumentierte Fiskal-Ausnahme:** `QuittiereTSESignaturauftrag` bleibt deutsch,
  obwohl Persistence-Layer. „Quittieren" ist dokumentierter Fiskalbegriff
  (language.md: „quittiert die Signatur direkt am Auftrag"); kein sauberes
  englisches Ein-Wort-Äquivalent.
- **Commit-Modus:** ein Commit pro Subsystem (`refactor(...)`), Message
  vorschlagen, erst nach Freigabe committen (No-Auto-Commit).
- **Verifikation:** Ground Truth ist `go build ./backend/...` (Exit-Code) + volles
  `make verify`. Editor-Diagnostics sind nach Renames und `make sqlc` oft stale.

## Resolved decisions

Klärungsrunde 1 (Scope): voller Repo-Sweep, alle deutschen Verb-sqlc-Queries,
Commit pro Subsystem mit Message-Vorschlag.

Klärungsrunde 2 (Domänen-Sprache, nach Code-Studium):

- **Leitprinzip Command/Query + Schicht** bestätigt (siehe oben).
- **`QuittiereTSESignaturauftrag` bleibt deutsch** (dokumentierter Fiskalbegriff).
- **Grenzbereich-Helfer werden englisch:** `waehlePositionen`→`selectPositionen`,
  `validiereProfil`→`validateProfil`, `verarbeite`→`process`,
  `signalisierePruefung`→`signalCheck`, `clientFuer`→`clientFor`.

### Bleibt deutsch (Command / Domänenintent) — kein Rename

Diese standen teils im ersten Planentwurf fälschlich auf der Übersetzungsliste
oder waren übersehen. Sie sind bereits korrekt:

- Worker-Domänenintent (`api/fiskal/signatur`): `beginneStoerung`,
  `beendeStoerung`, `beschaffeSignatur`, `markiereNichtKonfiguriert`,
  `markiereFehlgeschlagen`.
- Seed-Domänen-Simulatoren (`backend/seed/engine.go`): `bestellen`, `ausgeben`,
  `kassieren`, `stornieren`, `umbuchen`, `geldtransit`, `kassensturz`,
  `schliesseTagAb` (Tagesabschluss-Command, Geschwister der obigen).
- `QuittiereTSESignaturauftrag` (+ Repo-Wrapper) — dokumentierte Ausnahme.
- Alle bestehenden Kasse-/Stammdaten-Command-Handler (nie im Scope).

### Wird englisch — Entscheidung je Regel

| Bezeichner | Ziel | Regel |
| --- | --- | --- |
| `BestimmeSignaturstatus` | `DetermineSignaturstatus` | Derivation (wie `ComputeStornoAufteilung`) |
| `BerechneAbschlussSummen` | `ComputeAbschlussSummen` | Aggregation, Geschwister sind `Compute*` |
| `berechneUmsatzProSteuersatz` | `computeUmsatzProSteuersatz` | Reporting-Aggregation |
| `OeffneTSEStoerung` | `OpenTSEStoerung` | sqlc/Persistence (wie `SetKassensitzungOffen`) |
| `SchliesseTSEStoerung` | `CloseTSEStoerung` | sqlc/Persistence |
| `MarkiereOffeneTSESignaturauftraegeNichtKonfiguriert` | `MarkOffeneTSESignaturauftraegeNichtKonfiguriert` | sqlc (wie `MarkDruckauftragGedruckt`) |
| `MarkiereOffeneAlsNichtKonfiguriert` | `MarkOffeneAlsNichtKonfiguriert` | Repo-Adapter |
| `DruckauftragErneutVersuchen` | `RetryDruckauftrag` | Outbox-Mechanik |
| `DruckauftragVerwerfen` | `DiscardDruckauftrag` | Outbox-Mechanik |
| `MeldeDruckergebnis` / `meldeErgebnis` | `ReportDruckergebnis` / `reportErgebnis` | Relay-Transport |
| `HoleAdminPUK` | `GetAdminPUK` | Integration/Adapter |
| `SetzeAdminPIN` | `SetAdminPIN` | Integration/Adapter |
| `SpeichereEinrichtung` / `speichereEinrichtung` | `SaveEinrichtung` / `saveEinrichtung` | Persistence |
| `erzeugeAdminPIN` | `generateAdminPIN` | Plumbing (Zufalls-PIN) |
| `findeTSS` | `findTSS` | Integration-Lookup |
| `zieheTSEStammdaten` | `fetchTSEStammdaten` | Integration-Fetch |
| `pruefeStammdaten` | `checkStammdaten` | Validierung/Query |
| `PruefeTSESetup` (+Handler/Request/Schema) | `CheckTSESetup…` | Query/Validierung (wie `checkSignaturGate`); Route bleibt `tse-setup-pruefen` |
| `pruefeKeineOffeneKassensitzung` | `ensureKeineOffeneKassensitzung` | Guard/Query |
| `pruefeRueckstand` | `checkRueckstand` | Watchdog-Check (wie `checkSignaturGate`) |
| `signalisierePruefung` | `signalCheck` | Test-Helfer |
| `neuerTestWatchdog` / `neuerWorkerClient` | `newTestWatchdog` / `newWorkerClient` | Test-Konstruktor (Go-Idiom) |
| `clientFuer` | `clientFor` | Integration-Helfer |
| `schliesseKassensitzung` (Testfixture) | `closeKassensitzung` | Test-Persistence-Fixture |
| `sitzungsBauer` / `bondruckBauer` | `sitzungsBuilder` / `bondruckBuilder` | Muster-Suffix |
| `baueEvents` / `baueDruckauftraege` / `baueSignaturauftraege` | `buildEvents` / `buildDruckauftraege` / `buildSignaturauftraege` | Builder-Plumbing |
| `schreibe…` (7 Stück) | `write…` | Writer-Plumbing |
| `setzeStatus` | `setStatus` | Setter-Plumbing |
| `findeKassenbeleg` | `findKassenbeleg` | Test-Lookup |
| `waehlePositionen` | `selectPositionen` | Selektions-Query-Helfer |
| `validiereProfil` | `validateProfil` | Seed-Validierung |
| `verarbeite` | `process` | Dispatch-Plumbing |
| `pruefeTSESetup` (Frontend) | `checkTSESetup` | Query-Hook |
| `zaehleOffenePositionen` (Frontend) | `countOffenePositionen` | Derivation |

Bleibt unverändert (kein deutsches Verb): `IncrementDruckauftragFehlversuch`,
`TSESignaturauftragFehlversuch`.

## Open questions / Risks

- **Command/Query-Grenze im Worker:** `markiereNichtKonfiguriert` (deutsch, Intent)
  ruft `MarkOffeneAlsNichtKonfiguriert` (englisch, Repo) auf. Der Sprachwechsel an
  der Adapter-Grenze ist gewollt (Hexagonal), sieht aber auf den ersten Blick
  gemischt aus. Im Doc-Kommentar bleibt der Zusammenhang erklärt.
- **Cross-Subsystem-Aufrufer** bei geteilten Symbolen (`DetermineSignaturstatus`,
  `CloseTSEStoerung`) im selben Commit mitziehen, sonst rot zwischen den Phasen.
- Wird `make verify` rot und nicht sauber lösbar: anhalten und berichten, nicht
  schummeln.

---

## Phase 1: TSE-Setup / Einrichtung

### Context

- `backend/domain/tse/setup.go` — `SetupClient`-Interface (`HoleAdminPUK`,
  `SetzeAdminPIN`); `domain/tse/fake_setup_client.go` — Fake.
- `backend/repository/tse_repo/fiskaly_setup.go`, `einrichtung.go`.
- `backend/api/fiskal/setup/application/{setup,command,query}.go`,
  `http/query_handler.go`.

### What to build

Der Fiskal-Setup-Pfad ist Integration/Adapter (fiskaly-Anbindung) und
Query/Validierung, keine zustandsändernden Domänen-Commands. Alle Verben werden
englisch, Domänen-Nomen (AdminPUK, AdminPIN, Einrichtung, Stammdaten, TSS)
bleiben deutsch. Interface, Fake, reale Implementierung und Aufrufer atomar.

Umbenennungen: `HoleAdminPUK`→`GetAdminPUK`, `SetzeAdminPIN`→`SetAdminPIN`,
`SpeichereEinrichtung`→`SaveEinrichtung` (+ `speichereEinrichtung`→`saveEinrichtung`),
`PruefeTSESetup`→`CheckTSESetup` (+ Handler/`pruefeTSESetupRequest`/`pruefeTSESetupSchema`),
`pruefeKeineOffeneKassensitzung`→`ensureKeineOffeneKassensitzung`,
`pruefeStammdaten`→`checkStammdaten`, `erzeugeAdminPIN`→`generateAdminPIN`,
`findeTSS`→`findTSS`, `zieheTSEStammdaten`→`fetchTSEStammdaten`.

### Acceptance criteria

- [ ] Setup-Pfad frei von deutschen Verben (`grep` leer), Domänen-Nomen erhalten.
- [ ] `SetupClient`, Fake, fiskaly-Impl mit identischen englischen Methodennamen.
- [ ] Route `/admin/tse-setup-pruefen`, JSON-Felder, fiskaly-Feldnamen unverändert.
- [ ] `go build ./backend/...` und `make verify` grün.

---

## Phase 2: TSE-Störung & Signatur-Worker

### Context

- `backend/sqlc/queries/tse_stoerungen.sql`, `tse_signaturauftraege.sql`.
- `backend/repository/tse_repo/repo.go`, `einrichtung.go`.
- `backend/api/fiskal/signatur/tse_signatur_worker.go`,
  `tse_rueckstand_watchdog.go` und Tests.
- `backend/domain/tse/signaturstatus.go`, `stoerung.go`; externe Aufrufer
  `api/kasse/kassenfuehrung/application/kassenabschluss_gate.go`,
  `api/druck/beleg/application/kassenbeleg_command.go`.

### What to build

Persistence, Derivation und Checks englisch machen; die deutschen
Worker-Domänenintent-Verben (`beginneStoerung`, `beendeStoerung`,
`beschaffeSignatur`, `markiereNichtKonfiguriert`, `markiereFehlgeschlagen`) und
`QuittiereTSESignaturauftrag` bleiben deutsch. Zuerst sqlc (`make sqlc`), dann
Repo, Worker/Watchdog, Derivation samt subsystemfremder Aufrufer im selben Commit.

sqlc: `OeffneTSEStoerung`→`OpenTSEStoerung`, `SchliesseTSEStoerung`→`CloseTSEStoerung`,
`MarkiereOffeneTSESignaturauftraegeNichtKonfiguriert`→`MarkOffeneTSESignaturauftraegeNichtKonfiguriert`
(`QuittiereTSESignaturauftrag` bleibt). Repo:
`MarkiereOffeneAlsNichtKonfiguriert`→`MarkOffeneAlsNichtKonfiguriert`. Domain:
`BestimmeSignaturstatus`→`DetermineSignaturstatus`. Watchdog/Test:
`pruefeRueckstand`→`checkRueckstand`, `signalisierePruefung`→`signalCheck`,
`neuerTestWatchdog`→`newTestWatchdog`, `neuerWorkerClient`→`newWorkerClient`,
`clientFuer`→`clientFor`, `schliesseKassensitzung`(Fixture)→`closeKassensitzung`.

### Acceptance criteria

- [ ] `make sqlc` neu generiert; `dbgen/` nur maschinell geändert.
- [ ] Deutsche Intent-Verben und `QuittiereTSESignaturauftrag` unverändert; ihr
      Doc-Kommentar erklärt weiterhin die Adapter-Grenze zum englischen Repo.
- [ ] `DetermineSignaturstatus`-Aufrufer in kassenfuehrung-Gate und druck/beleg
      mitgezogen; Verhalten identisch.
- [ ] Tabellen/Spalten/Status-Enums (`offen`, `erledigt`, …) unverändert.
- [ ] `go build ./backend/...` und `make verify` grün.

---

## Phase 3: Druck (Relay + Auftrag)

### Context

- `backend/sqlc/queries/druckauftraege.sql`.
- `backend/repository/druckauftrag_repo/repo.go` und Tests.
- `backend/api/druck/relay/http/handler.go` und Tests, `backend/api/relay.go`,
  `api/druck/relay/relay_integration_test.go`.

### What to build

Die Druck-Outbox und der Relay-Rückkanal sind Warteschlangen- und
Transport-Infrastruktur → englische Verben; die Bon-Nomen (Druckauftrag,
Druckergebnis) bleiben deutsch. sqlc mit `make sqlc`, dann Repo/Handler.

sqlc: `DruckauftragErneutVersuchen`→`RetryDruckauftrag`,
`DruckauftragVerwerfen`→`DiscardDruckauftrag`. Go:
`MeldeDruckergebnis`→`ReportDruckergebnis`, `meldeErgebnis`→`reportErgebnis`.

### Acceptance criteria

- [ ] `make sqlc` neu generiert; Querier-Interface und Repo-Wrapper konsistent.
- [ ] Relay-Routen (`/relay/poll`, `/relay/ergebnis`), JSON und
      `druckauftraege`-Statuswerte (`offen`, `gedruckt`, `fehlgeschlagen`,
      `verworfen`) unverändert.
- [ ] `go build ./backend/...` und `make verify` grün.

---

## Phase 4: Kasse- & Reporting-Summen

### Context

- `backend/domain/kasse/tagesabschluss_summen.go` (`BerechneAbschlussSummen`);
  Aufrufer `api/kasse/kassenfuehrung/application/command.go`,
  `repository/reporting_repo/summen_abschluss_test.go`, Domain-Test.
- `backend/api/reporting/application/query.go` (`berechneUmsatzProSteuersatz`).

### What to build

Beide Funktionen sind Aggregationen ohne Zustandsänderung und stehen neben
`Compute*`-Geschwistern im selben Paket. `BerechneAbschlussSummen`→
`ComputeAbschlussSummen` (mit Aufrufern), `berechneUmsatzProSteuersatz`→
`computeUmsatzProSteuersatz`. Werte, DSFinV-K-Felder, JSON-Keys unverändert.

### Acceptance criteria

- [ ] Beide englisch benannt, alle Aufrufer/Tests nachgezogen.
- [ ] Summen-/Umsatzberechnung und `tagesabschluss-erstellt:v1`-Felder unverändert.
- [ ] `go build ./backend/...` und `make verify` grün.

---

## Phase 5: Seed

### Context

- `backend/seed/writer.go` (`schreibe…`, `baue…`), `engine.go` (`sitzungsBauer`,
  `baueEvents`, `waehlePositionen`, `validiereProfil`, `verarbeite`;
  Domänen-Simulatoren `bestellen`/`kassieren`/`schliesseTagAb` etc. **bleiben**),
  `bondruck.go` (`baueDruckauftraege`, `setzeStatus`, `bondruckBauer`),
  `faketse.go` (`baueSignaturauftraege`), `bondruck_test.go` (`findeKassenbeleg`).

### What to build

Nur Builder, Writer, Setter und Query-Helfer werden englisch; die
deutsch-verbigen Domänen-Simulatoren (`bestellen`, `ausgeben`, `kassieren`,
`stornieren`, `umbuchen`, `geldtransit`, `kassensturz`, `schliesseTagAb`)
bleiben unangetastet.

Umbenennungen: `sitzungsBauer`/`bondruckBauer`→`sitzungsBuilder`/`bondruckBuilder`,
alle `baue…`→`build…`, alle `schreibe…`→`write…`, `setzeStatus`→`setStatus`,
`findeKassenbeleg`→`findKassenbeleg`, `waehlePositionen`→`selectPositionen`,
`validiereProfil`→`validateProfil`, `verarbeite`→`process`.

### Acceptance criteria

- [ ] Builder/Writer/Helfer englisch; Domänen-Simulatoren bleiben deutsch.
- [ ] Domänen-Nomen in Bezeichnern unverändert (`writeSitzungen`, nicht
      `writeSessions`).
- [ ] Seed-Ausgabe unverändert; `make verify` (inkl. Seed-Tests) grün.

---

## Phase 6: Frontend

### Context

- `frontend/src/admin/tse/TSEBackend.ts:208`, `hooks.ts:52,55`,
  `TSEEinrichtungWizard.tsx:24,88` (`pruefeTSESetup`).
- `frontend/src/service/components/MeinTischCard.tsx:21,54,57`
  (`zaehleOffenePositionen`).

### What to build

Beide Bezeichner sind Query/Derivation (kein Domänen-Command) → englisch:
`pruefeTSESetup`→`checkTSESetup`, `zaehleOffenePositionen`→`countOffenePositionen`
(inkl. Kommentar). Kritisch: kein angezeigter String, kein `aria`-Text, kein
Domänen-Nomen, kein JSON-Feldname wird angefasst.

### Acceptance criteria

- [ ] Beide Bezeichner englisch, alle Vorkommen (inkl. Kommentar) nachgezogen.
- [ ] Kein UI-Text, `aria`-Text oder Domänen-Nomen verändert.
- [ ] Frontend-Check grün (pnpm lint `--max-warnings=0`, typecheck, test, build)
      bzw. volles `make verify`.

---

## Phase 7: docs/language.md

### Context

- `docs/language.md:7-16` — Abschnitt „Sprachkonventionen".
- `docs/language.md:431` — Referenz `BestimmeSignaturstatus`.

### What to build

Das Leitprinzip explizit dokumentieren, denn es ist das eigentliche Ergebnis
dieses Umbaus und war bisher nur implizit im Kasse-Core kodiert. Als Regel im
Abschnitt „Sprachkonventionen" festhalten: zustandsändernde Domänen-Commands
tragen deutsche Fachverben; Queries, Derivationen, Persistence und Infrastruktur
englische Verben; Domänen-Nomen bleiben deutsch. Die dokumentierte
`quittieren`-Ausnahme erwähnen. Danach die Referenz `BestimmeSignaturstatus`→
`DetermineSignaturstatus` (Zeile 431) nachziehen. Stil minimal, keine em-dashes,
kein liberales Bold. Historische `docs/plans/*`-Artefakte bleiben unberührt.

### Acceptance criteria

- [ ] Command/Query + Schicht-Prinzip in `docs/language.md` dokumentiert, inkl.
      `quittieren`-Ausnahme.
- [ ] `BestimmeSignaturstatus`-Referenz auf `DetermineSignaturstatus` aktualisiert.
- [ ] Keine sonstigen stale Identifier-Referenzen im Living Document (`grep` leer).
- [ ] Doc-Stil ohne Slop-Syntax.
