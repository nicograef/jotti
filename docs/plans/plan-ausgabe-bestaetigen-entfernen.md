# Plan: Ausgabe-Bestätigung entfernen

> Source PRD: `docs/prds/prd-ausgabe-bestaetigen-entfernen.md`
> Entscheidung: [ADR 01](../adrs/01_ausgabe-bestaetigen.md)

## Goal

Die Ausgabe-Bestätigung (K-03, Event `ausgabe-bestaetigt:v1`) wird
vollständig aus jotti entfernt — UI, API, Event-Typ, Projektion und
Altdatenbestand. „Offene Arbeit" bedeutet danach „noch nicht kassiert".
Bestehende Installationen werden per Release-Migration automatisch
bereinigt; ein manueller Eingriff ist nicht vorgesehen.

## Architectural decisions

Durable Entscheidungen, die für alle Phasen gelten:

- **Migration**: `database/migrations/03_ausgabe_entfernen.up.sql`,
  forward-only, in `BEGIN;`/`COMMIT;` geklammert (Konvention laut
  `database/migrations/README.md`). Reihenfolge innerhalb der
  Transaktion: `ALTER TABLE kassenjournal DISABLE TRIGGER
  kassenjournal_no_delete` → `DELETE FROM tisch_sessions` (FK auf
  Events; Projektion wird beim Start neu aufgebaut) → `DELETE FROM
  kassenjournal WHERE type = 'ausgabe-bestaetigt:v1'` → `ALTER TABLE
  kassenjournal ENABLE TRIGGER kassenjournal_no_delete` → `ALTER TABLE
  tisch_sessions DROP COLUMN ausstehende_positionen`. Läuft als
  Schema-Owner im Migrate-Container; der Append-only-Schutz besteht
  danach unverändert.
- **Domain-Modell**: `TischSession` verliert das Feld
  `AusstehendePositionen`. `EigeneArbeitAnTisch` verliert
  `AnzahlAusstehend`; `Erledigt` ist definiert als
  `AnzahlUnbezahlt == 0`. Die Event-Switches bleiben exklusiv
  (unbekannter Event-Typ → Fehler) — das Sicherheitsnetz wird nicht
  aufgeweicht.
- **API-Contract**: Die Route `POST /service/ausgabe-bestaetigen` und
  der Fehlercode `position_nicht_ausgebbar` entfallen ersatzlos. In den
  JSON-Antworten entfallen `ausstehendePositionen` (Tisch-State),
  `anzahlAusstehend` (Reporting) und die Historie-Eintragsart
  `ausgabe`. Keine API-Versionierung (Frontend wird im selben Release
  mitgezogen).
- **UI**: Das Badge im Tisch-Header ist zahlungsbasiert: destruktive
  Variante mit der Anzahl unbezahlter Positionen, solange welche
  existieren, sonst „Alles bezahlt". Der offene Saldo steht bereits
  daneben im Header und wird nicht dupliziert. Die Tischlisten-Karte
  zählt „offen" nur noch über unbezahlte Positionen.

## Inventory

Backend-Domäne:

- `backend/domain/kasse/tisch_session_events.go` —
  `EventTypeAusgabeBestaetigtV1`, `AusgabeBestaetigtV1Data`,
  `NewAusgabeBestaetigtEvent()`, `buildAusgabeFromEvent()`
- `backend/domain/kasse/ausgabe.go` — `Ausgabe` (komplette Datei)
- `backend/domain/kasse/tisch_session.go` —
  `TischSession.AusstehendePositionen`, `ApplyEvent()` (Ausgabe-Case),
  `ComputeNichtStorniertePositionen()` (Skip-Case)
- `backend/domain/kasse/historie.go` — `HistorieEintragAusgabe`,
  `HistorieEintrag.Ausgabe`, `GetHistorieFromEvents()` (Ausgabe-Case)
- `backend/domain/kasse/fiskalische_projektion.go` —
  `FiskalischeProjektion()` (nicht-signaturpflichtig-Case)
- `backend/domain/kasse/offene_arbeit.go` —
  `ComputeEigeneArbeitAnTisch()`, `ComputeOffeneArbeitRollup()`,
  `ComputeOffeneArbeitProServicekraft()`,
  `EigeneArbeitAnTisch.AnzahlAusstehend`
- `backend/domain/kasse/event_json_contract_test.go` — `allEventTypes`,
  `contractedTypes`, `TestEventContract_AusgabeBestaetigtV1`
- `backend/domain/kasse/replay_fuzz_test.go` — Event-Generatorliste

Backend-API und Persistenz:

- `backend/api/kasse/tischgeschaeft/application/command.go` —
  `AusgabeBestaetigen()`; die geteilten Helfer `validatePositionRefs()`
  und `resolvePositions()` bleiben (Kassieren/Umbuchen nutzen sie)
- `backend/api/kasse/tischgeschaeft/application/errors.go` —
  `ErrPositionNichtAusgebbar`
- `backend/api/kasse/tischgeschaeft/http/command_handler.go` —
  `AusgabeBestaetigenHandler()` + Interface-Methode
- `backend/api/kasse/tischgeschaeft/http/query_handler.go` —
  `ausgabe`-Struct, `toAusgabe()`, Tisch-State-Response
- `backend/api/kasse/tischgeschaeft/application/query.go` —
  `TischDetail.AusstehendePositionen`, `TischDetail.FuerMichErledigt`
- `backend/api/service.go` — Route `POST /ausgabe-bestaetigen`
- `backend/api/reporting/application/query.go` — Rollup-Mapping
- `backend/api/reporting/http/query_handler.go` —
  `offeneArbeitTischResponse.AnzahlAusstehend`
- `backend/repository/kassenjournal_repo/repo.go` —
  `upsertTischSessionState()`, `toTischSession()`,
  `RebuildAllProjections()` (läuft bei jedem Backend-Start via
  `backend/main.go`)
- `backend/repository/kassenjournal_repo/repo_test.go` —
  Integrationstest „Ausgabe ist nicht signaturpflichtig"
- `backend/sqlc/queries/tisch_sessions.sql` — Spalte
  `ausstehende_positionen` (danach `make sqlc`)
- `backend/seed/engine.go`, `backend/seed/szenario.go` —
  `ausgeben`-Aktionen; `backend/seed/seed_integration_test.go`
- `database/migrations/01_initial.up.sql` — Append-only-Trigger
  `kassenjournal_no_delete`, Spalten-Definition (bleibt unverändert;
  nur als Referenz)

Frontend und E2E:

- `frontend/src/service/components/table/Ausgabe.tsx`,
  `AusgabeDrawer.tsx`, `Ausgabe.test.tsx` — komplett löschen
- `frontend/src/service/table/Ausgabe.ts` — komplett löschen
- `frontend/src/service/table/TischBackend.ts` — `ausgabeBestaetigen()`
- `frontend/src/service/table/Tisch.ts` —
  `TischSessionSchema.ausstehendePositionen`, `fuerMichErledigt`
- `frontend/src/service/TablePage.tsx` — Header-Badge, Ausgabe-Card im
  Bestellen-Tab; `TablePage.test.tsx`
- `frontend/src/service/components/MeinTischCard.tsx` —
  `countOffenePositionen()` (Vereinigung ausstehend + unbezahlt)
- `frontend/src/service/components/table/TischHistorie.tsx` —
  Eintragsart `ausgabe` in Liste und Detailansicht
- `frontend/src/lib/errorMessages.ts` — `position_nicht_ausgebbar`
- `frontend/src/admin/reporting/types.ts` —
  `OffeneArbeitTischSchema.anzahlAusstehend`;
  `frontend/src/admin/reporting/LiveReportingSection.tsx`
- `e2e/support/servicekraft.ts` — Klickpfad „Ausgabe bestätigen"

Dokumentation:

- `docs/anforderungen.md` — K-03 (Funktionsumfang), K-13/K-15 (Roadmap)
- `docs/handbuch.md` — Event-Tabelle, Ausgabe-Invariante,
  Rollen-/Mindestmengen-Invariante, Berechtigungsmatrix,
  UI-Beschreibung des Bestellen-Tabs
- `docs/language.md` — Glossareintrag „Ausgabe"
- `docs/produktbeschreibung.md`, `README.md` — Feature-Listen

## Resolved decisions

- Vollständige Entfernung inklusive Lesepfade; Alt-Events werden per
  Release-Migration gelöscht (kein manueller Eingriff, keine
  Legacy-Lesepfade).
- „Offene Arbeit" = unbezahlt; Kennzahl und Signale bleiben erhalten,
  nur die Datenbasis wechselt.
- K-13 (Küchendisplay) und K-15 (Zubereitungsstatus) werden von der
  Roadmap gestrichen (ADR 01).
- Testumfang: bestehende Tests anpassen plus neue Unit-Tests für die
  Offene-Arbeit-Neudefinition und ein Integrationstest für die
  Migration.
- Die einmalige Append-only-Ausnahme ist durch ADR 01 gedeckt
  (pre-v1.0, Testbetrieb ohne TSE, Event nicht signaturpflichtig).

> **Assumption:** Der 4-Phasen-Schnitt wurde als empfohlener Default
> übernommen; die strukturierte Rückfrage zur Granularität konnte nicht
> beantwortet werden (Tool-Abbruch). Bei Bedarf Phasen 1+2 bzw. 3+4
> zusammenlegen — die Schnittstellen zwischen den Phasen bleiben gleich.

## Open questions / Risks

- Die Migration nimmt durch `ALTER TABLE … DISABLE TRIGGER` ein
  `ACCESS EXCLUSIVE`-Lock auf `kassenjournal`. Im normalen
  Update-Ablauf (Migrate-Container läuft vor dem App-Start) ist das
  unkritisch; bei laufendem Backend würde die Migration auf offene
  Transaktionen warten.
- Versionslücken je Subject nach der Event-Löschung sind unschädlich:
  Die optimistische Nebenläufigkeitskontrolle setzt nur auf der
  Maximal-Version auf (`GetMaxVersion()`), und der Startup-Rebuild
  erzeugt die Projektion aus den verbleibenden Events.
- Reihenfolge der Phasen ist risikoarm gewählt: Erst entstehen keine
  neuen Events mehr (Phase 1), dann hängt fachlich nichts mehr am
  Ausgabe-Status (Phase 2), erst danach folgt der harte Schnitt am
  Datenmodell (Phase 3).

---

## Phase 1: Schreibpfad und Service-UI entfernen

**User stories**: 1, 4

### Context

- `backend/api/service.go` — Routen-Registrierung des Service-Bereichs
- `backend/api/kasse/tischgeschaeft/application/command.go` —
  `AusgabeBestaetigen()` (einziger Schreiber des Events)
- `backend/api/kasse/tischgeschaeft/http/command_handler.go` —
  `AusgabeBestaetigenHandler()` + Interface-Methode
- `frontend/src/service/TablePage.tsx` — Bestellen-Tab mit Ausgabe-Card
  und Header-Badge
- `backend/seed/szenario.go` — Demo-Szenario mit `ausgeben`-Aktionen
- `e2e/support/servicekraft.ts` — E2E-Schritte „Ausgabe bestätigen"

### What to build

Der komplette Schreibpfad verschwindet end-to-end: kein API-Endpunkt,
kein Command, kein Fehlercode, keine Ausgabe-UI. Der Bestellen-Tab der
Tisch-Seite enthält nur noch die Bestellkomponente; das Header-Badge
(„X offen / Alles ausgegeben!") entfällt in dieser Phase ersatzlos
(zahlungsbasierter Ersatz folgt in Phase 2). Seed-Szenarien erzeugen
keine Ausgabe-Events mehr; der E2E-Service-Flow kommt ohne
Ausgabe-Schritt aus. Der Domain-Lesepfad (Event-Typ, Projektion,
Historie) bleibt in dieser Phase unangetastet — bestehende Events werden
weiterhin korrekt replayed.

### Acceptance criteria

- [ ] Die Route `POST /service/ausgabe-bestaetigen` ist entfernt;
      Command `AusgabeBestaetigen()`, Handler, Interface-Methode und
      `ErrPositionNichtAusgebbar` existieren nicht mehr. Die geteilten
      Helfer `validatePositionRefs()`/`resolvePositions()` bleiben und
      Kassieren/Umbuchen funktionieren unverändert.
- [ ] `Ausgabe.tsx`, `AusgabeDrawer.tsx`, `Ausgabe.test.tsx` und
      `Ausgabe.ts` (Service-Frontend) sind gelöscht;
      `TischBackend.ausgabeBestaetigen()` und die Fehlermeldung
      `position_nicht_ausgebbar` sind entfernt.
- [ ] Der Bestellen-Tab rendert nur noch die Bestellkomponente; das
      Header-Badge ist entfernt; `TablePage.test.tsx` ist angepasst.
- [ ] Seed-Engine und -Szenario enthalten keine `ausgeben`-Aktionen
      mehr; `seed_integration_test.go` erwartet als nicht-fiskalischen
      Event-Typ nur noch `kassensturz-durchgefuehrt:v1`.
- [ ] Der E2E-Service-Flow läuft ohne Ausgabe-Schritt durch.
- [ ] `make check` ist grün; die unveränderten Domain- und
      Contract-Tests bleiben grün (der Event-Typ existiert noch).

---

## Phase 2: Offene Arbeit = unbezahlt

**User stories**: 2, 3

### Context

- `backend/domain/kasse/offene_arbeit.go` — heutige Definition
  „erledigt = nichts ausstehend UND nichts unbezahlt"
- `backend/api/reporting/application/query.go` +
  `backend/api/reporting/http/query_handler.go` — Live-Reporting-Rollup
  mit `AnzahlAusstehend`
- `backend/api/kasse/tischgeschaeft/application/query.go` —
  `TischDetail.FuerMichErledigt`
- `frontend/src/service/components/MeinTischCard.tsx` —
  `countOffenePositionen()` (Vereinigung ausstehend + unbezahlt)
- `frontend/src/admin/reporting/types.ts`,
  `frontend/src/admin/reporting/LiveReportingSection.tsx` —
  Reporting-Anzeige
- `frontend/src/service/TablePage.tsx` — Header (Saldo-Anzeige
  vorhanden, Badge kommt zahlungsbasiert zurück)

### What to build

Die Kennzahl „offene Arbeit" wechselt vollständig auf die
Zahlungs-Datenbasis: `EigeneArbeitAnTisch` verliert `AnzahlAusstehend`,
`Erledigt` bedeutet `AnzahlUnbezahlt == 0`. Live-Reporting (Rollup
gesamt und je Servicekraft), das Signal `fuerMichErledigt` in der
Tischliste und der Tischlisten-Zähler basieren nur noch auf unbezahlten
Positionen. Der Tisch-Header bekommt das zahlungsbasierte Badge
(destruktiv mit Anzahl unbezahlter Positionen, sonst „Alles bezahlt").
Das Feld `AusstehendePositionen` existiert danach noch in Projektion
und API-Antworten, wird aber von keiner Fachlogik mehr gelesen.

### Acceptance criteria

- [ ] `EigeneArbeitAnTisch` hat kein Feld `AnzahlAusstehend` mehr;
      `Erledigt` ist genau dann true, wenn keine unbezahlten Positionen
      der Servicekraft am Tisch existieren. Neue table-driven
      Unit-Tests in `offene_arbeit_test.go` decken die Neudefinition ab
      (erledigt/nicht erledigt, Rollup gesamt, Rollup je Servicekraft).
- [ ] Die Live-Reporting-Antwort enthält kein `anzahlAusstehend` mehr;
      Frontend-Schema und Anzeige sind angepasst.
- [ ] `MeinTischCard` zählt „offen" nur über unbezahlte Positionen.
- [ ] Der Tisch-Header zeigt das zahlungsbasierte Badge; die
      Saldo-Anzeige daneben bleibt unverändert.
- [ ] `make check` ist grün.

---

## Phase 3: Event-Typ und Datenbestand ausbauen

**User stories**: 5

### Context

- `backend/domain/kasse/tisch_session_events.go`,
  `backend/domain/kasse/ausgabe.go`,
  `backend/domain/kasse/tisch_session.go`,
  `backend/domain/kasse/historie.go`,
  `backend/domain/kasse/fiskalische_projektion.go` — alle
  Ausgabe-Anteile des Event-Modells
- `backend/domain/kasse/event_json_contract_test.go` — erzwingt, dass
  jeder Event-Typ in `allEventTypes` gepflegt ist
- `backend/repository/kassenjournal_repo/repo.go` —
  `RebuildAllProjections()` (Startup-Replay),
  `upsertTischSessionState()`, `toTischSession()`
- `backend/sqlc/queries/tisch_sessions.sql` — Spalte
  `ausstehende_positionen`
- `database/migrations/README.md` — Migrations-Konventionen
  (Transaktions-Klammer, forward-only)
- `frontend/src/service/components/table/TischHistorie.tsx`,
  `frontend/src/service/table/Tisch.ts` — letzte Frontend-Reste

### What to build

Der harte Schnitt am Datenmodell. Die Migration
`03_ausgabe_entfernen.up.sql` bereinigt bestehende Installationen
gemäß den Architectural decisions (Trigger transaktional deaktivieren,
Projektion und Alt-Events löschen, Spalte droppen). Im Code verschwindet
der Event-Typ vollständig: Konstante, Data-Struct, Konstruktor,
`Ausgabe`-Struct, alle Switch-Cases, die Historie-Eintragsart und das
Projektionsfeld `AusstehendePositionen` samt DB-Spalte, sqlc-Queries und
Repository-Mapping. Die Event-Switches bleiben exklusiv. Im Frontend
entfallen der Historie-Case `ausgabe` und das Schema-Feld
`ausstehendePositionen`. Der Contract-Test pinnt den Typ nicht mehr;
der Replay-Fuzz generiert ihn nicht mehr; der Integrationstest „Ausgabe
ist nicht signaturpflichtig" entfällt ersatzlos.

### Acceptance criteria

- [ ] `database/migrations/03_ausgabe_entfernen.up.sql` existiert, ist
      in `BEGIN;`/`COMMIT;` geklammert und führt aus: Trigger
      `kassenjournal_no_delete` deaktivieren, `tisch_sessions` leeren,
      `ausgabe-bestaetigt:v1`-Events löschen, Trigger wieder
      aktivieren, Spalte `ausstehende_positionen` droppen. Ein
      Kommentar in der Migration verweist auf ADR 01.
- [ ] Neuer Integrationstest: Ein Kassenjournal mit
      Ausgabe-Alt-Events wird migriert; danach läuft
      `RebuildAllProjections()` fehlerfrei und die Tisch-Zustände
      (Saldo, unbezahlte Positionen) sind korrekt. Der Test belegt
      auch, dass der Append-only-Trigger nach der Migration wieder
      aktiv ist (DELETE auf `kassenjournal` schlägt fehl).
- [ ] Grep über `backend/` und `frontend/src/` nach
      `ausgabe-bestaetigt`, `AusgabeBestaetigt`, `AusstehendePositionen`
      bzw. `ausstehendePositionen` liefert keine Code-Treffer mehr
      (Migrationen und `docs/` ausgenommen).
- [ ] `ApplyEvent()`, `ComputeNichtStorniertePositionen()`,
      `GetHistorieFromEvents()` und `FiskalischeProjektion()` behandeln
      unbekannte Event-Typen weiterhin als Fehler (Contract-Test
      `allEventTypes` ohne den entfernten Typ, Fuzz angepasst).
- [ ] `make sqlc` ist gelaufen; `sqlc/dbgen/` enthält keine
      `ausstehende_positionen`-Referenzen mehr.
- [ ] Die Tisch-Historie im Frontend kennt nur noch Bestellung,
      Zahlung, Stornierung und Umbuchung.
- [ ] `make verify` (inkl. Integrationstests) ist grün.

---

## Phase 4: Dokumentation bereinigen und Abschlussverifikation

**User stories**: — (Doku-Konsistenz; stützt alle Stories)

### Context

- `docs/anforderungen.md` — K-03 im Funktionsumfang, K-13/K-15 in der
  Roadmap
- `docs/handbuch.md` — Event-Tabelle, Ausgabe-Invariante,
  Berechtigungsmatrix, UI-Beschreibung
- `docs/language.md` — Glossareintrag „Ausgabe" (der Eintrag
  „Arbeitsbon" bleibt)
- `docs/produktbeschreibung.md`, `README.md` — Feature-Listen

### What to build

Die Dokumentation spiegelt den neuen Ist-Zustand: K-03 verschwindet aus
dem Funktionsumfang, K-13/K-15 aus der Roadmap (jeweils mit Verweis auf
ADR 01). Das Handbuch verliert das Event in der Tisch-Session-Tabelle,
die Ausgabe-Invariante, die Ausgabe-Zeile der Berechtigungsmatrix und
die Ausgabe-Erwähnungen in der UI-Beschreibung; die Beschreibung der
Kennzahl „offene Arbeit" nennt die Zahlungs-Datenbasis. Glossar,
Produktbeschreibung und README werden bereinigt. Zum Abschluss läuft die
volle Verifikation.

### Acceptance criteria

- [ ] Grep über `docs/` (ohne `docs/prds/`, `docs/plans/`,
      `docs/adrs/`) und `README.md` nach „Ausgabe bestätigen",
      `ausgabe-bestaetigt` und K-03 liefert keine Treffer mehr, die den
      entfernten Ist-Zustand beschreiben.
- [ ] K-13 und K-15 sind aus der Roadmap entfernt; die Streichung
      verweist auf ADR 01.
- [ ] Die Berechtigungsmatrix und die Invarianten-Liste im Handbuch
      enthalten keine Ausgabe-Zeilen mehr; die
      Offene-Arbeit-Beschreibung nennt „unbezahlt" als Datenbasis.
- [ ] `make verify` und der E2E-Lauf sind grün; PRD, ADR und dieser
      Plan sind die einzigen verbleibenden Dokumente, die das entfernte
      Feature beschreiben.
