# Plan: TSE-Integration (F-02)

> Source PRD: `docs/prds/prd-tse-integration.md`

## Goal

jotti erhält eine vollständige, anbieter-agnostische TSE-Integration nach dem
**Festzelt-Muster** (atomare, sofort geschlossene Transaktionen): ein
`TSEClient`-Port mit einer fiskaly-Implementierung (SIGN DE API v2), synchrones
**sign-then-persist** für alle neun fiskalisch relevanten Vorgänge, ein
**Ausfall-Pfad** („DON'T BLOCK THE TILL") mit idempotentem Nachsignier-Worker,
sowie ein um TSE-Pflichtfelder und DSFinV-K-QR-Code erweiterter Kassenbeleg.
TSE bleibt **optional** — ohne Konfiguration läuft jotti unverändert weiter und
der Beleg ist byte-identisch zum heutigen Zustand.

## Architectural decisions

Durable decisions that apply across all phases:

- **API-Routen** (POST-only, Admin-Bereich):
  - `POST /admin/get-tse-konfiguration` — Konfiguration lesen (Secrets maskiert)
  - `POST /admin/update-tse-konfiguration` — Konfiguration setzen/leeren
  - `POST /admin/test-tse-verbindung` — Auth + TSS-Status gegen fiskaly prüfen
  - `POST /admin/get-tse-status` — Umgebung (TEST/LIVE) + Anzahl offener
    Nachsignierungen
- **Schema** (direkt in `database/migrations/01_initial.up.sql`, keine neue
  Migrationsdatei):
  - `tse_konfiguration` — Singleton (`id = 1`), Felder `api_key`, `api_secret`,
    `tss_id`, `client_id`, `updated_at`; Muster wie `betreiber` /
    `bondruck_einstellungen`. Plaintext (Verschlüsselung-at-rest out of scope).
  - `tse_nachsignier_auftraege` — Outbox für ausgefallene Signierungen:
    `tx_id` (deterministisch, UNIQUE), `process_type`, `process_data`,
    `status` (`offen` | `erledigt`), `erstellt_am`, `erledigt_am`; Muster wie
    `druckauftraege` (Index auf `(status, id)`).
  - `tse_signaturen` — Signatur-Seitentabelle für **nachsignierte** Vorgänge,
    per `tx_id` (UNIQUE) verknüpft: Transaktionsnummer, Signaturzähler,
    TSE-Seriennummer, LogTime-Start/-Ende, Signatur, QR-Code-Daten.
  - `tisch_sessions` erhält `erste_bestellung_logtime TIMESTAMPTZ NULL` für die
    Durchbedienen-Klarschrift.
- **Key models / Ports**:
  - `TSEClient`-Interface (Domain-Port) nach `docs/handbuch.md` §3.13:
    `StartTransaction` / `UpdateTransaction` / `FinishTransaction` mit
    `StartResult` / `FinishResult`. `UpdateTransaction` ist Teil des Interfaces,
    wird im atomaren Muster aber nicht aufgerufen.
  - `FiskalyTSEClient` — Adapter gegen fiskaly SIGN DE API v2 (Auth + Token-Cache
    - `upsertTransaction`); dünner HTTP/JSON↔Domain-Übersetzer.
  - `TSEData`-Struct (JSON-Tags erlaubt — dokumentierte Event-Data-Ausnahme),
    eingebettet in die neun fiskalischen Event-Data-Structs.
  - Deterministische TSE-`tx_id` pro Vorgang, abgeleitet aus der Event-Identität
    (Idempotenz für Start/Finish und Nachsignieren).
- **Orchestrierung**: Ein TSE-Application-Service kapselt sign-then-persist,
  Ausfall-Pfad und Nachsignieren. Der Adapter implementiert ausschließlich den
  Port (tiefes, testbares Modul vs. dünner Adapter).
- **Worker**: In-Process-Goroutine im Backend (direkter DB- + fiskaly-Zugriff),
  gestartet beim App-Start. **Nicht** der externe `cmd/relay`-Prozess.
- **QR-Code**: Nativer ESC/POS-Befehl `GS ( k` im Bondruck-Formatter (Drucker
  rendert den QR selbst). Exakte Geräteunterstützung am MUNBYN-Gerät zu
  verifizieren — analog zum bestehenden Codepage-Vorbehalt.
- **fiskaly-Basis-URL**: Backend-Env/Config (`config.Config`) mit Default auf die
  SIGN-DE-v2-Middleware-URL; konsistent mit `JWT_SECRET` / `RELAY_AUTH_TOKEN`.
  TEST/LIVE ergibt sich aus den Credentials (Token-Claim), nicht aus der URL.

## Inventory

Bestehende Strukturen, die gespiegelt oder erweitert werden:

**Event-Sourcing-Write-Pfad (die neun Vorgänge)**

- `backend/domain/kasse/tisch_session_events.go:7-15` — Event-Typ-Konstanten;
  `:20-95` Event-Data-Structs + Schemas; `:91-152` Konstruktoren
  (`bestellung-aufgenommen`, `zahlung-kassiert`, `stornierung-erteilt`,
  `auszahlung-geleistet`).
- `backend/domain/kasse/direktverkauf_events.go:6-7` Konstanten; `:10-43`
  Structs; `:48-78` Konstruktoren (`direktverkauf-getaetigt`/`-storniert`).
- `backend/domain/kasse/kassensitzung_events.go:9-14` Konstanten; `:30-96`
  Structs; `:123-243` Konstruktoren (`geldtransit-gebucht`,
  `differenz-soll-ist-gebucht`, `tagesabschluss-erstellt`).
- `backend/domain/kasse/bestellung.go:10-18` — `Position` (mit `Steuersatz`);
  `:23-32` `positionEventData` (fat positions).
- `backend/repository/kassenjournal_repo/repo.go:33` `WriteEvent`; `:57`
  `WriteEventWithDruckauftraege`; `:117` `writeEventInTx` (Projektion-Routing).
- `backend/api/table/application/command.go:26-60` Repo-Interfaces; `:119-160`
  `writeEventOCC` / `writeEvent` / `writeEventWithDruckauftraege` (OCC-Wrapper).
- `backend/api/kasse/application/command.go:51-65` `writeKassensitzungEvent`;
  `:134` `GeldtransitBuchen`; `:181` `KassensturzDurchfuehren`; `:251`
  `TagesabschlussErstellen`.
- `backend/api/direktverkauf/application/command.go:81` `DirektverkaufTaetigen`;
  `:186` `DirektverkaufStornieren`; `:314-350` OCC-Wrapper.
- DI-Wiring: `backend/api/service.go:37-49` (table), `:54-65` (direktverkauf);
  `backend/api/admin.go:110-124` (kasse).

**Beleg / Bondruck**

- `backend/api/bondruck/application/escpos/formatter.go:17-31` `KassenbelegData`;
  `:33-40` `TSEAbschnitt` (Vorarbeit, bereits vorhanden); `:172`
  `FormatKassenbeleg`.
- `backend/api/bondruck/application/escpos/constants.go:1-30` — ESC/POS-Befehle
  (`Init`, `CutPaper`, `SetCodepageCP858` …; QR-Befehl `GS ( k` zu ergänzen).
- `backend/api/bondruck/application/escpos/formatter_test.go:297-360` — TSE-Block-
  Verhaltenstests (ohne/mit TSE) als Regressionsanker.
- `backend/api/table/application/kassenbeleg_command.go:90` `KassenbelegDrucken`
  (baut `KassenbelegData`, ruft `FormatKassenbeleg`, Enqueue Druckauftrag).

**Steuer-Aufteilungslogik (wiederverwenden, nicht duplizieren)**

- `backend/domain/steuer/steuer.go:7-11` Steuersätze; `:13-18` `Aufteilung`;
  `:42` `Aufteilen` (kombi 70/30); `:64` `Steuermatrix`.

**Singleton-Config-Muster (Phase 1 Vorlage)**

- `database/migrations/01_initial.up.sql:368-385` `betreiber`; `:388-407`
  `bondruck_einstellungen` (Singleton `id = 1 CHECK`).
- `backend/sqlc/queries/betreiber.sql`, `…/bondruck_einstellungen.sql` —
  `:one` + `UpsertX :exec` mit `ON CONFLICT (id) DO UPDATE`.
- `backend/repository/settings_repo/repo.go:13-42` Get/Upsert Betreiber; `:44-54`
  `GetKassenidentitaet`; `:59-81` Bondruck.
- `backend/domain/settings/betreiber.go`, `…/bondruck_einstellungen.go` —
  Domain-Modelle (keine JSON-Tags) + zog-Validierung.
- `backend/api/settings/application/{query,command}.go`,
  `backend/api/settings/http/{query_handler,command_handler}.go` — Query/Command +
  Response-DTOs.
- `backend/api/admin.go:121-128` — Settings-Endpunkt-Wiring.
- `frontend/src/lib/EinstellungenBackend.ts`,
  `frontend/src/admin/settings/EinstellungenPage.tsx`,
  `frontend/src/lib/Backend.ts` (`BackendClient`).

**Outbox + Worker-Muster (Phase 4 Vorlage)**

- `database/migrations/01_initial.up.sql:239-256` `druckauftraege`
  (Status-CHECK, Index `(status, id)`).
- `backend/repository/druckauftrag_repo/repo.go:35-117` — Enqueue / Insert (tx) /
  GetOffene / Quittiere.
- `cmd/relay/main.go:33-100+` — Poll-Loop-Vorbild (Struktur, nicht Ablageort).
- `backend/app/app.go:24-39` `NewApp`; `:42-75` `SetupRoutes`; `:78` `Run`
  (Einhängepunkt für die Worker-Goroutine).

**Config / Seriennummer**

- `backend/config/config.go:17-23` `Config`; `:26` `Load` (fiskaly-Basis-URL
  ergänzen).
- `database/migrations/01_initial.up.sql:328-365` `kassenidentitaet` (F-01,
  Seriennummer für fiskaly `serial_number`).

**Referenzdokumente**

- `docs/handbuch.md:429-560` (§3.13 TSE-Architektur, `TSEClient`, `TSEData`,
  Vorgang→processType-Mapping).
- `docs/compliance.md:139-238` (§3.3–§3.6 processType, processData-Format,
  Festzelt-Muster); `:341-390` (§5.3 Durchbedienen).
- `docs/steuerrecht.md:114-128` (§6 Belegausweis, Steuerkennzeichen).

## Resolved decisions

- **Nachsignier-Worker** läuft als **In-Process-Goroutine** im Backend (direkter
  DB- + fiskaly-Zugriff), nicht im externen `cmd/relay`-Prozess.
- **QR-Code** wird **nativ per ESC/POS `GS ( k`** gedruckt (Drucker rendert);
  keine Backend-Bitmap, keine neue Bild-Dependency.
- **fiskaly-Basis-URL** kommt aus **Backend-Env/Config** mit sinnvollem Default;
  kein zusätzliches DB-Konfig-Feld.
- TSE ist **optional ohne `strict`/`bypass`-Schalter**; `docs/anforderungen.md`
  F-02 wird auf „optional + Admin-Hinweis/Warnung" angeglichen (PRD-Vorgabe).
- `kassensitzung-eroeffnet:v1` (Anfangsbestand) wird **nicht** signiert
  (PRD-Annahme; Mapping in §3.13 führt nur Geldtransit als `SonstigerVorgang-V1`).

## Open questions / Risks

- **ESC/POS `GS ( k` am MUNBYN-Gerät**: native QR-Unterstützung und Modellgröße
  sind hardwareabhängig und am Zielgerät zu verifizieren (wie der bestehende
  Codepage-Index-Vorbehalt in `constants.go`).
- **fiskaly TEST-Umgebung** löscht TSS regelmäßig (sonntags / nach 14 Tagen
  Inaktivität, max. 5 aktive TSS). CI testet ausschließlich gegen einen
  Mock-HTTP-Server, nicht gegen fiskaly.
- **`signature.counter`** kann von fiskaly als Zahl **oder** String kommen — der
  Adapter muss beides akzeptieren.

---

## Phase 1: TSE-Konfiguration (BYOT) + Admin-Hinweis/Warnung + F-02-Doku

**User stories**: 1, 2, 3, 4, 5, 7, 8

### Context

- `backend/api/settings/**` + `backend/repository/settings_repo/repo.go:13-42` —
  Singleton-Config-Muster (Domain → Repo → Query/Command → HTTP-DTO).
- `database/migrations/01_initial.up.sql:368-407` — Vorlage `betreiber` /
  `bondruck_einstellungen` (Singleton `id = 1 CHECK`).
- `frontend/src/lib/EinstellungenBackend.ts`,
  `frontend/src/admin/settings/EinstellungenPage.tsx` — Admin-Settings-UI.
- `database/migrations/01_initial.up.sql:328-365` — Kassenseriennummer (F-01),
  die als fiskaly `serial_number` angezeigt wird.
- `docs/anforderungen.md` F-02 — Akzeptanzkriterien anzugleichen.

### What to build

Ein vollständiger CRUD-Slice für die TSE-Konfiguration **ohne** fiskaly-Aufrufe:
neue Singleton-Tabelle `tse_konfiguration`, Lese-/Schreib-/Leer-Endpunkte
(`get-tse-konfiguration`, `update-tse-konfiguration`), beidseitige Validierung
(zog + Zod), und eine Admin-Seite zum Pflegen von `api_key`, `api_secret`,
`tss_id`, `client_id`. Secrets werden in der Lese-Antwort maskiert (Präsenz statt
Klartext). Die Admin-Oberfläche zeigt prominent die Kassenseriennummer (F-01) zur
Übernahme als fiskaly `serial_number`. Das Admin-Dashboard zeigt einen deutlichen
**Hinweis + Warnung**, solange keine TSE konfiguriert ist (Betreiberverantwortung,
nur Test/Demo/Übung). Abschließend wird `docs/anforderungen.md` F-02 von
„strict/bypass" auf „TSE optional + Hinweis/Warnung" angeglichen.

### Acceptance criteria

- [x] `tse_konfiguration` existiert als Singleton in der initialen Migration;
      `make down && make dev` setzt die DB sauber neu auf.
- [x] `POST /admin/update-tse-konfiguration` setzt und leert die vier Felder;
      `POST /admin/get-tse-konfiguration` liefert sie mit maskierten Secrets.
- [x] Beide Seiten validieren (zog + Zod) mit deutschen Fehlermeldungen; Aufrufe
      laufen über eine `BackendClient`-basierte Backend-Klasse (kein `fetch`).
- [x] Die Admin-Seite zeigt die Kassenseriennummer und erlaubt das Pflegen/Leeren
      der TSE-Konfiguration.
- [x] Bei fehlender TSE erscheint im Admin-Dashboard ein sichtbarer Hinweis +
      Warnung; bei konfigurierter TSE verschwindet er.
- [x] `docs/anforderungen.md` F-02 spiegelt „optional + Hinweis/Warnung".
- [x] `make check` und Frontend-Lint sind grün; Response-DTOs liegen in der
      HTTP-Schicht (Domain-Modelle ohne JSON-Tags).

---

## Phase 2: fiskaly-Adapter + `TSEClient`-Port + Fake + „Verbindung testen"

**User stories**: 6, 20, 21, 25, 26

### Context

- `docs/handbuch.md:455-520` — `TSEClient`-Interface, `StartResult` /
  `FinishResult`.
- `docs/compliance.md:114-159` — Transaktions-Lebenszyklus, processType-Werte.
- `backend/config/config.go:17-26` — Env/Config (fiskaly-Basis-URL ergänzen).
- `backend/repository/settings_repo/repo.go:13-42` — Quelle der TSE-Credentials
  (aus Phase 1).
- `cmd/relay/main.go` + bestehende HTTP-Handler-Tests — Vorbild für
  Mock-HTTP-Server-Tests.

### What to build

Der anbieter-agnostische `TSEClient`-Port wird definiert und mit zwei
Implementierungen versehen: einem **Fake** (für Service-/Command-Tests) und dem
**`FiskalyTSEClient`** gegen fiskaly SIGN DE API v2. Der Adapter authentifiziert
per `api_key`/`api_secret`, **cacht** das Bearer-JWT und erneuert es bei
Ablauf/`401`; er signiert per `upsertTransaction` (Start `state=ACTIVE`,
`tx_revision=1`; Finish `state=FINISHED`, `tx_revision=2`) über das `raw`-Schema
(`process_type` + `process_data`) und mappt die Antwort (`number`,
`tss_serial_number`, `signature.counter` als Zahl **oder** String,
`signature.value`, `log.timestamp` / `time_start` / `time_end`, `qr_code_data`).
Wiederholbare Fehler (`5xx`, `499`, `429` mit `Retry-After`) werden mit
exponentiellem Backoff erneut versucht. Ein `POST /admin/test-tse-verbindung`
prüft Auth + TSS-Status und meldet die Umgebung; die Admin-UI zeigt **TEST** oder
**LIVE** an.

### Acceptance criteria

- [x] `TSEClient`-Port mit `StartResult`/`FinishResult` exakt nach §3.13; Fake
      deckt Erfolg, Fehler und Timeout ab.
- [x] `FiskalyTSEClient` erwirbt und cacht das Token und erneuert bei `401`.
- [x] Start/Finish senden korrektes `state` + `tx_revision`; Antwortfelder werden
      vollständig auf `StartResult`/`FinishResult` gemappt (`counter` als Zahl und
      String akzeptiert).
- [x] Retry mit Backoff bei `5xx`/`499`/`429`; nicht-wiederholbare Fehler werden
      durchgereicht.
- [x] `POST /admin/test-tse-verbindung` liefert ein Ergebnis inkl. TEST/LIVE; die
      Admin-UI zeigt die Umgebung an.
- [x] Adapter-Tests laufen gegen einen lokalen Mock-HTTP-Server (kein echter
      fiskaly-Zugriff); `make check` grün.

---

## Phase 3: Happy-Path-Signierung `zahlung-kassiert:v1` + processData + Beleg-TSE-Block

**User stories**: 10, 13, 15, 22, 23, 24, 27

### Context

- `backend/api/table/application/command.go:608-612` (Zahlung kassieren) +
  `:119-160` OCC-Wrapper — Einhängepunkt sign-then-persist.
- `backend/domain/kasse/tisch_session_events.go:36-49` — `zahlung-kassiert`
  Event-Data (um `TSEData` zu erweitern).
- `backend/domain/steuer/steuer.go:42-86` — `Aufteilen`/`Steuermatrix` als Quelle
  für die processData-Beträge (kombi 70/30 → ermäßigt/normal).
- `docs/compliance.md:160-172` — `processData`-Formatvorgaben
  (`Beleg^…`, Punkt-Dezimal).
- `backend/api/bondruck/application/escpos/formatter.go:17-40` — `KassenbelegData`
  - bestehender `TSEAbschnitt` (nur befüllen + rendern).
- `backend/api/table/application/kassenbeleg_command.go:90-200` — Beleg-Bau.

### What to build

Der erste fiskalische Tracer Bullet: ein **TSE-Application-Service** umschließt das
Kassieren einer Tischzahlung mit synchronem **sign-then-persist** — er leitet eine
deterministische `tx_id` aus der Event-Identität ab, ruft `StartTransaction` und
unmittelbar `FinishTransaction` (Festzelt-Muster) und bettet das Ergebnis als
`TSEData` in die Event-Daten ein, **bevor** das Event transaktional ins
Kassenjournal geschrieben wird. Eine **reine** processData-Formatter-Funktion
erzeugt den `Kassenbeleg-V1`-String (fünf Betragstellen, kombi 70/30 in
ermäßigt/normal, Punkt als Dezimaltrenner, keine Tausendertrenner) auf Basis der
**bestehenden** Steuer-Aufteilungslogik. Der Kassenbeleg rendert den bereits
vorhandenen optionalen `TSEAbschnitt` aus den Signaturdaten. Ist **keine** TSE
konfiguriert, wird nicht signiert und der Beleg bleibt byte-identisch.

### Acceptance criteria

- [x] Mit konfigurierter TSE trägt ein `zahlung-kassiert:v1`-Event die
      `TSEData`-Felder (Transaktionsnummer, Signaturzähler, TSE-Seriennummer,
      LogTime-Start/-Ende, Signatur) im Kassenjournal.
- [x] Die processData für `Kassenbeleg-V1` entspricht §3.4: korrekte Slot-Belegung,
      kombi 70/30, befreit (0 %), Punkt-Dezimal, keine Tausendertrenner — über
      table-driven Tests abgesichert.
- [x] Die `tx_id` ist deterministisch aus der Event-Identität abgeleitet.
- [x] Der gedruckte Kassenbeleg zeigt die TSE-Pflichtfelder, wenn Signaturdaten
      vorliegen.
- [x] Ohne konfigurierte TSE wird nicht signiert und der Beleg ist byte-identisch
      zum heutigen Zustand (bestehende Formatter-Tests bleiben grün).
- [x] Der Service wird gegen den Fake-`TSEClient` getestet (Happy Path);
      `make check` grün.

---

## Phase 4: Ausfall-Pfad + Nachsignier-Worker + Beleg-Ausfallvermerk + Admin-Status

**User stories**: 9, 11, 12, 14, 19, 28, 29, 30

### Context

- Phase-3-TSE-Service — erweitert um den Fehler-Zweig.
- `backend/repository/druckauftrag_repo/repo.go:35-117` — Outbox-Muster (Insert /
  GetOffene / Quittiere) als Vorlage.
- `cmd/relay/main.go:33-100+` — Poll-Loop-Struktur (Logik, nicht Ablageort).
- `backend/app/app.go:24-39,78` — `NewApp` / `Run` (Worker-Goroutine starten).
- `backend/api/bondruck/application/escpos/formatter.go:172` — Ausfallvermerk auf
  dem Beleg; Signaturbezug aus Event **oder** Seitentabelle.

### What to build

Der Ausfall-Pfad nach „DON'T BLOCK THE TILL": schlägt die TSE-Kommunikation fehl,
wird der Vorgang **ohne** Signatur persistiert, ein Eintrag in die neue Outbox
`tse_nachsignier_auftraege` (mit deterministischer `tx_id`) geschrieben und auf dem
Beleg ein **Ausfallvermerk** gedruckt — der Verkauf wird **nie** blockiert. Eine
**In-Process-Goroutine** drained die Outbox idempotent (`upsertTransaction` trifft
dank `tx_id` denselben fiskaly-Vorgang) und legt die erhaltene Signatur in der
Seitentabelle `tse_signaturen` ab; das immutable Event bleibt unverändert. Der
Beleg-Druck bezieht die TSE-Felder entweder aus dem Event (Happy Path) oder aus
`tse_signaturen` (nachsigniert). `POST /admin/get-tse-status` liefert die Anzahl
offener Nachsignierungen; die Admin-UI zeigt sie an.

### Acceptance criteria

- [ ] Bei TSE-Fehler wird das Event ohne Signatur persistiert, ein Outbox-Eintrag
      angelegt und der Verkauf erfolgreich abgeschlossen (nicht blockiert).
- [ ] Der Beleg trägt bei Ausfall einen klaren Ausfallvermerk.
- [ ] Die Worker-Goroutine signiert offene Aufträge nach, markiert sie als
      erledigt und schreibt die Signatur in `tse_signaturen`.
- [ ] Ein erneuter Lauf gegen dieselbe `tx_id` erzeugt **keine** Dublette
      (Idempotenz).
- [ ] Der Beleg-Druck bezieht die Signatur aus Event **oder** Seitentabelle.
- [ ] `POST /admin/get-tse-status` liefert die Anzahl offener Nachsignierungen;
      die Admin-UI zeigt sie an.
- [ ] Service- und Worker-Verhalten gegen den Fake getestet (Erfolg/Fehler/
      Idempotenz); `make check` grün.

---

## Phase 5: Signierung auf alle neun Vorgänge ausweiten

**User stories**: 11, 17, 22, 23, 27, 28

### Context

- `backend/api/table/application/command.go:650,712` — `stornierung-erteilt`,
  `auszahlung-geleistet`.
- `backend/api/direktverkauf/application/command.go:81,186` —
  `direktverkauf-getaetigt`/`-storniert`.
- `backend/api/kasse/application/command.go:134,181,251` — `geldtransit-gebucht`,
  `differenz-soll-ist-gebucht`, `tagesabschluss-erstellt`.
- `backend/domain/kasse/{tisch_session_events,direktverkauf_events,kassensitzung_events}.go`
  — die übrigen Event-Data-Structs (um `TSEData` erweitern).
- `docs/handbuch.md:484-505` — Vorgang→processType-Mapping.

### What to build

Der in Phase 3/4 etablierte sign-then-persist- + Ausfall-Pfad wird auf die übrigen
acht fiskalischen Vorgänge ausgeweitet, jeweils mit dem korrekten `processType`:
`Bestellung-V1` (Bestellung), `Kassenbeleg-V1` (Storno, Auszahlung,
Direktverkauf/-storno — Storni mit negativem Betrag) und `SonstigerVorgang-V1`
(Geldtransit, Kassendifferenz, Tagesabschluss). Der processData-Formatter erhält
die Zweige `Bestellung-V1` (Positionen als strukturierter Text) und
`SonstigerVorgang-V1` (vorgangsspezifischer Inhalt). Alle neun Event-Data-Structs
tragen das optionale `TSEData`-Feld und überstehen einen verlustfreien
Marshal/Unmarshal-Round-Trip (mit und ohne TSE).

### Acceptance criteria

- [ ] Alle neun Vorgänge signieren im Happy Path mit korrektem `processType` und
      betten `TSEData` ein; bei Ausfall greift der Outbox-/Nachsignier-Pfad.
- [ ] processData-Zweige für `Bestellung-V1` und `SonstigerVorgang-V1` existieren;
      Storni erzeugen negative Beträge.
- [ ] Event-Data-Round-Trip ist für alle neun Structs verlustfrei — mit TSE-Feldern
      und ohne (kein TSE konfiguriert bleibt gültig).
- [ ] `ausgabe-bestaetigt:v1` und `kassensitzung-eroeffnet:v1` werden **nicht**
      signiert.
- [ ] Bestehende Command-/Handler-Tests bleiben grün; `make check` grün.

---

## Phase 6: QR-Code (native ESC/POS) + Durchbedienen-Klarschrift

**User stories**: 16, 18, 31

### Context

- `backend/api/bondruck/application/escpos/constants.go:1-30` — ESC/POS-Befehle
  (QR-Befehl `GS ( k` ergänzen).
- `backend/api/bondruck/application/escpos/formatter.go:17-40,172` — `qr_code_data`
  rendern; Beleg-Kopf um Erste-Bestellung-Zeitstempel erweitern.
- `database/migrations/01_initial.up.sql` (`tisch_sessions`) — Spalte
  `erste_bestellung_logtime` ergänzen.
- `backend/repository/kassenjournal_repo/repo.go:117` — Projektion der
  Tisch-Session (erste `Bestellung-V1`-LogTime festhalten).
- `docs/compliance.md:341-390` — §5.3 Durchbedienen (Klarschrift-Pflichtfeld).

### What to build

Der Beleg wird um den **DSFinV-K-QR-Code** ergänzt: der Bondruck-Formatter erhält
eine native ESC/POS-QR-Fähigkeit (`GS ( k`) und druckt `qr_code_data` aus dem
`FinishTransaction`-Ergebnis. Für die **Durchbedienen-Klarschrift** hält die
Tisch-Session-Projektion den `logTime` der **ersten** `Bestellung-V1`-Transaktion
fest (neue Spalte `erste_bestellung_logtime`); der Zahlungsbeleg druckt diesen
Startzeitpunkt zusätzlich in Klarschrift (nur Tisch-Kassenbelege — Direktverkäufe
haben keine vorgelagerte Bestellung).

### Acceptance criteria

- [ ] Liegt `qr_code_data` vor, druckt der Beleg einen nativen ESC/POS-QR-Code;
      ohne TSE-Daten erscheint kein QR.
- [ ] `tisch_sessions.erste_bestellung_logtime` wird beim ersten
      `bestellung-aufgenommen:v1` der Session gesetzt.
- [ ] Der Tisch-Zahlungsbeleg zeigt den Startzeitpunkt der ersten Bestellung in
      Klarschrift; Direktverkauf-Belege nicht.
- [ ] Formatter-Verhaltenstests decken QR-Präsenz/-Abwesenheit und die Klarschrift
      ab; der Beleg endet weiterhin mit dem Schnittbefehl.
- [ ] `make check` grün; ohne TSE bleibt der Beleg byte-identisch (kein QR, keine
      TSE-Zeilen).
