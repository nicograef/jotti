# Plan: Modulschnitt nach Kontextgrenzen

> Source PRD: [docs/prds/prd-modulschnitt-kontextgrenzen.md](../prds/prd-modulschnitt-kontextgrenzen.md)

## Goal

Backend und Frontend werden am Subdomänen-Zielbild ausgerichtet: Kontext-Ordner in `api/`
(`kasse`, `fiskal`, `druck`, `stammdaten`, `reporting`), Auflösung des God-Moduls `api/table`,
reporting-freier Kassenabschluss über eine reine Tagesabschluss-Domänenfunktion, Aufspaltung
von settings, deutsche Namen für Fachmodule, Frontend-Feinschnitt und Doku-Angleich.
Reiner Strukturumbau: API-Pfade, JSON-Formate, DB-Schema, Event-Formate und UI bleiben
identisch. Bestehende Tests ziehen mit ihren Modulen um und bleiben unverändert grün;
`make verify` sichert jede Phase.

## Architectural decisions

Ziel-Landkarte `api/` (Module behalten den Schnitt `{application,http}`):

| Kontext | Module | Herkunft |
| --- | --- | --- |
| `kasse/` (Core) | `tischgeschaeft`, `direktverkauf`, `kassenfuehrung` | api/table (Kern), api/direktverkauf, api/kasse |
| `fiskal/` | `signatur` (inkl. Worker + Watchdog), `setup`, `export`, `dsfinvk` (reines Paket, ohne application/http) | api/tse + app/, api/settings (TSE-Teil), api/export, domain/dsfinvk |
| `druck/` | `bondruck` (inkl. escpos), `beleg`, `auftrag`, `station`, `relay` | api/bondruck, api/table (Kassenbeleg), api/druckauftrag, api/druckstation, api/relay |
| `stammdaten/` | `produkt`, `tisch` (CRUD + Favoriten), `user`, `betreiber` | api/product, api/table (CRUD), api/user, api/settings (Betreiber-Teil) |
| `reporting/` | unverändert flach (`api/reporting/{application,http}`) | — |

- **Top-Level `api/`**: `auth/`, `middleware/`, `helper/`, `health/` und die Wiring-Dateien
  (`admin.go`, `service.go`, `serviceleitung.go`, `auth.go`, `relay.go`) bleiben, ergänzt um den
  Deps-Bausatz (Phase 8). Paketnamen der Modul-Unterordner bleiben `application`/`http`.
- **Domain-Schicht**: `domain/kasse` erhält die Tagesabschluss-Aggregation als reine Funktion.
  `domain/table` → `domain/tisch`, `domain/product` → `domain/produkt`. `domain/settings` entfällt:
  Betreiber → neues `domain/betreiber`; TSE-Konfiguration, TSE-Stammdaten, Kassenidentität →
  `domain/tse`. `domain/dsfinvk` zieht als Paket nach `api/fiskal/dsfinvk`; der Typ `EventSignatur`
  zieht vorab nach `domain/tse` (sonst importierte `repository/kassenjournal_repo` die api-Schicht).
- **Repository-Schicht** bleibt flach: `settings_repo` wird aufgeteilt in neues `betreiber_repo`
  (Betreiber) und `tse_repo` (Kassenidentität, TSE-Konfiguration, TSE-Stammdaten, Einrichtung);
  `product_repo` → `produkt_repo`, `table_repo` → `tisch_repo`. sqlc: `tables.sql` → `tische.sql`,
  danach `make sqlc` (Query-Dateien sind bereits pro Tabelle geschnitten, für den settings-Split
  ändern sich keine Queries).
- **Endpunkt-Zuordnung nach dem settings-Split** (Pfade unverändert): `fiskal/setup` bedient
  `get-tse-konfiguration`, `update-tse-konfiguration`, `tse-einrichten`, `tse-uebernehmen`,
  `tse-setup-pruefen`, `test-tse-verbindung`, `get-tse-status`, `get-kassenidentitaet`;
  `stammdaten/betreiber` bedient `get-betreiber`, `update-betreiber`.
- **Consumer-Interface statt Shim**: Die Konsumenten (tischgeschaeft, direktverkauf, beleg)
  definieren `GetKonfigurierteDruckstationen` mit `map[string]druckstation.Druckstation`
  (Domain-Typ); `druckstation_repo` erfüllt das direkt. Der Typ
  `bondruck/application.Druckstation` entfällt, die Arbeitsbon-Policy nimmt den Domain-Typ.
- **Tagesabschluss-Aggregation**: Eingabe alle Events der Kassensitzung (Journal-Events plus
  die im selben Command erzeugten, noch nicht gelesenen Kassensturz-/Differenz-Events in-memory
  angehängt), Ausgabe die drei Summen des `tagesabschluss-erstellt:v1`-Events. Der Soll-Bestand
  kommt weiterhin aus `kassenjournal_repo.GetKassenbestand` (SQL, kein Reporting-Baustein).
- **Worker als Lifecycle**: Signatur-Worker und Rückstand-Watchdog leben in `fiskal/signatur`;
  `app/` konstruiert und startet sie nur noch (exportierte Konstruktoren, konfigurationsfreie
  Signaturen — application-Pakete importieren `config` auch künftig nicht).
- **Deps-Bausatz**: ein Struct im `api`-Paket, einmal in `app.SetupRoutes` konstruiert und an die
  Bereichs-Konstruktoren gereicht; jede Repository-Konstruktion existiert genau einmal.

## Inventory

God-Modul und Wiring:

- `backend/api/table/application/command.go:87-97` — Command-Struct mit neun Dependencies;
  `:20-77` Consumer-Interfaces; `:253-353` Favoriten und Tisch-CRUD; `:355-767` Tischgeschäft
  (Bestellen, Umbuchen, Kassieren, Storno, Ausgabe) samt geteilter Helfer
  (`writeEventOCC:123`, `loadTischState:197`, `validatePositionRefs:229`, `resolvePositions:460`).
- `backend/api/table/application/query.go` — Tisch-Queries; `kassenbeleg_command.go` (433 Zeilen)
  — Kassenbeleg-Command inkl. Signaturstatus-Auflösung (`tseAbschnittFuerBeleg:144`).
- `backend/api/table_bondruck_adapters.go` — der Adapter-Shim (entfällt).
- Partielle Injektion: `backend/api/service.go:46-57` (9 Deps), `admin.go:79-85` (4 Deps),
  `serviceleitung.go:29-37` (6 Deps). Duplizierte Repo-Konstruktion in allen fünf Wiring-Dateien.
- `backend/api/direktverkauf/application/command.go:35-36` — dasselbe
  `bondruckApp.Druckstation`-Interface (vom Shim-Entfall mitbetroffen);
  `backend/seed/bondruck.go:9-10` — dritter Nutzer der Policy.
- `backend/api/bondruck/application/arbeitsbon_policy.go:15-18` — Typ `Druckstation`;
  `:37-40` Policy-Signatur. `backend/repository/druckstation_repo/repo.go:40-54` — liefert bereits
  `map[string]druckstation.Druckstation` (`domain/druckstation/druckstation.go:51-55`).
- `backend/api/relay.go` — `druckauftragRepoRelayAdapter` ist ein bewusster Typ-Mapper der
  HTTP-Schicht, nicht der PRD-Shim; bleibt.

Kassenabschluss und Reporting:

- `backend/api/kasse/application/command.go:38-40,54` — `reportingRepo`-Interface und -Feld;
  `:251-394` `KasseAbschliessen`; `:362-368` der zu ersetzende `GetReporting`-Aufruf;
  `:317` Soll-Bestand aus `GetKassenbestand` (bleibt). `kassenabschluss_gate.go` zieht mit um.
- `backend/sqlc/queries/reporting.sql:10-43` — `GetReportingStats`: Referenzsemantik der drei
  Summen; `differenz-soll-ist-gebucht:v1` ist nicht im Typ-Filter.
- `backend/sqlc/queries/kassensitzungen.sql:43-68` — `GetKassenbestand` inkl. Differenz (bleibt).
- `backend/repository/kassenjournal_repo/repo.go:526` — `ReadEventsByKassensitzung` (liefert
  Events + Signaturen für den Export); `:473` importiert `dsfinvk.EventSignatur`.
- `backend/domain/kasse/kassensitzung_events.go:173-190` — `NewTagesabschlussErstelltEvent`
  (drei Summen). Vorbilder für tabellengetriebene Tests: `domain/kasse/storno_aufteilung_test.go`,
  `offene_arbeit_test.go`; Konsistenztest-Vorbild:
  `backend/api/reporting/application/query_export_konsistenz_test.go` (unit) und
  `backend/seed/seed_integration_test.go` (integration).

settings und Fiskal:

- `backend/domain/settings/` — `betreiber.go`, `kassenidentitaet.go`, `tse_konfiguration.go`
  (+ Test), `tse_stammdaten.go`. Importiert von 12 Nicht-Test-Dateien (api/export, api/kasse,
  api/settings, api/table, app/Worker, domain/dsfinvk, settings_repo, seed).
- `backend/repository/settings_repo/repo.go:15-42` — Betreiber-Zugriffe (→ betreiber_repo);
  `:44-160` — Kassenidentität, TSE-Konfiguration, Einrichtung, TSE-Stammdaten (→ tse_repo).
- `backend/api/settings/application/` — `query.go` (255 Z.), `command.go` (80 Z.),
  `setup.go` (473 Z., TSE-Einrichtung/-Übernahme); HTTP-Handler je in einer Datei für beide
  Zielmodule. Routen: `backend/api/admin.go:156-184`.
- `backend/api/tse/application/query.go` — Signatur-Monitoring (Queue-Zustand,
  Störungsprotokoll), hängt an `tse_repo`-Typen.
- `backend/app/app.go:86-90` — Worker-Start; `tse_signatur_worker.go:104-110` — Konstruktor
  nutzt `cfg.FiskalyBaseURL`; `tse_rueckstand_watchdog.go`; Tests inkl.
  `tse_signatur_worker_integration_test.go` ziehen mit.
- `backend/api/export/application/export.go:32` — Interface auf `dsfinvk.EventSignatur`;
  `:40-44` settingsRepo-Interface (Betreiber, Kassenidentität, TSE-Stammdaten — wird auf
  betreiber_repo + tse_repo aufgeteilt).
- `backend/domain/dsfinvk/` — `mapper.go:10-13` importiert event, kasse, steuer, tse;
  `dsfinvk.go:14-15` settings, steuer; `signaturen.go:10-13` Typ `EventSignatur`;
  `mapper.go:31,407-415,1055` GV_TYP `DifferenzSollIst`.

Frontend:

- `frontend/src/lib/EinstellungenBackend.ts:185-296` — 13-Methoden-Sammelklasse
  (`:192-206` Kassenidentität/Betreiber, `:209-296` TSE). Nutzer: `admin/settings/hooks.ts`,
  `admin/finanzamt/{BetreiberSection,SignaturauftraegeSection,TSEAusfalldokumentationSection}.tsx`,
  `admin/tse/{TSEEinrichtungWizard,TSEKonfigurationSection}.tsx`.
- `frontend/src/lib/DruckstationBackend.ts` (+ Test) — admin-exklusiv, Nutzer:
  `admin/settings/{DruckstationConfigPage.tsx,hooks.ts}`.
- `frontend/src/admin/kasse/KassensitzungPage.tsx` (613 Z.) — drei Sektionen in einer Datei:
  `EroeffnenSection:106`, `GeldtransitSection:240`, `KasseAbschliessenSection:415`; Test
  `KassensitzungPage.test.tsx`.
- `frontend/src/lib/` verbleibend (generisch): `Backend.ts`, `Auth.ts`/`AuthBackend.ts`,
  `arbeitsmodus.ts`, `download.ts`, `errorMessages.ts`, `identity.ts`, `utils.ts` (+ Tests).

Dokumentation:

- `docs/handbuch.md:31-53` — 2.1/2.2 Kontextübersicht (drei Kontexte) und Beziehungen;
  `:41` der zu streichende Bondruck-Policy-Satz; `:211-216` 3.12 Policies; `:217-252` 3.13
  TSE-Architektur mit Pfaden (`backend/app/tse_signatur_worker.go`, `backend/domain/dsfinvk`,
  `backend/api/export`); `:284-303` 4.6 Bondruck; `:361-371` 6.1 Schichtenarchitektur.
- `docs/language.md:17` — Namenskonventionen pro Schicht; `:103` Hinweis zu `domain/table/`;
  `:325` EntityStatus-Zeile nennt `domain/table`, `domain/product`.
- `.github/instructions/event-sourcing.instructions.md:3` — applyTo mit `backend/domain/table/**`;
  `backend.instructions.md:79,124,162,165` — Codebeispiele mit `product.*`.
- `AGENTS.md:103-107` — Bereichsbeschreibung mit Pfad-Referenzen.

## Resolved decisions

- **Differenzbuchung bleibt summen-neutral** (recherchiert statt gefragt, auf Nutzeranweisung):
  DSFinV-K v2.4 führt `DifferenzSollIst` als eigenen GV_TYP in der Gruppe „betreffen (direkt)
  ausschließlich den Kassenbestand" neben `Geldtransit` (Tz. 4.1.3, S. 33) und definiert ihn als
  Abweichung zwischen errechnetem und gezähltem Kassenbestand beim Kassensturz (Anhang C, S. 58).
  jotti ist heute spec-konform: Kassenbestand rechnet die Differenz ein
  (`kassensitzungen.sql:49-68`, handbuch 3.9), der Export führt sie als eigenen GV_TYP
  (`mapper.go:31,407-415`), die drei Z-Bon-Summen schließen sie aus (`reporting.sql:34-43`).
  Die neue Domänenfunktion behandelt sie deshalb summen-neutral (wie Korrektur und Umbuchung);
  die PRD-Notiz zum In-Memory-Anhängen wird als Input-Vollständigkeit umgesetzt (Vertrag „alle
  Events der Sitzung"; deckt zugleich den Wiederanlauf ab, bei dem Kassensturz/Differenz eines
  gescheiterten Vorversuchs bereits im Journal stehen).
- **Kassenidentität-Endpunkt** (`get-kassenidentitaet`) liegt in `fiskal/setup`: Die PRD ordnet
  die Kassenidentität explizit dem Fiskal-Kontext zu; das Frontend gruppiert sie in der
  Betreiber-Backend-Klasse (Finanzamt-Slice), was nur den Aufrufer betrifft, nicht den Pfad.
- **`EventSignatur` zieht nach `domain/tse`**: erzwungen durch die Schichtung —
  `kassenjournal_repo` liefert den Typ (`repo.go:473,526`) und darf nach dem dsfinvk-Umzug kein
  api-Paket importieren. Fachlich passt er zur Signaturstatus-Familie in `domain/tse`.
- **Soll-Bestand bleibt SQL** (`kassenjournal_repo.GetKassenbestand`): Die PRD entfernt nur die
  Reporting-Abhängigkeit; das Kassenjournal-Repository ist ein Kasse-Baustein, und die Query wird
  vom `get-kassenbestand`-Endpunkt ohnehin weiter gebraucht.
- **Paketpräfix-Stottern beim settings-Umzug bereinigen**: `settings.TSEKonfiguration` und
  `settings.TSEStammdaten` werden in `domain/tse` zu `tse.Konfiguration` und `tse.Stammdaten`
  (Go-Idiom, mechanisches Rename beim Umzug); `Kassenidentitaet` und `Betreiber` behalten ihre
  Namen.
- **Relay-Adapter bleibt**: `druckauftragRepoRelayAdapter` (`api/relay.go`) ist ein dokumentierter
  Typ-Mapper, damit die HTTP-Schicht das Repository nicht importiert — nicht der PRD-Shim.
- **Phasenschnitt**: 10 Phasen, vom Nutzer bestätigt; echte Änderungen (1, 5, 8) strikt getrennt
  von mechanischen Umzügen (PRD-Gegenmittel).

## Open questions / Risks

- **Versehentliche Logikänderung bei Umzügen** ist das Hauptrisiko (PRD). Gegenmittel: Umzüge per
  `git mv` plus reiner Import-/Alias-Anpassung, unveränderte Tests als Regressionsnetz,
  `make verify` als Abschluss jeder Phase.
- **Große Import-Umschreibungen** (Phasen 4, 6, 7) sind breit, aber mechanisch; `goimports` und
  der Compiler erzwingen Vollständigkeit.
- Der Äquivalenztest koppelt Go-Aggregation und `kj_extract_*`-SQL dauerhaft: jeder künftige
  geldrelevante Event-Typ muss beide Stellen nachziehen (bewusster Preis laut PRD).

---

## Phase 1: Tagesabschluss-Aggregation als Domänenfunktion

**User stories**: 6, 7, 19

### Context

- `backend/api/kasse/application/command.go:38-40,50-56` — reportingRepo-Interface und
  Command-Deps; `:251-394` Abschluss-Ablauf; `:362-377` GetReporting-Aufruf und Event-Bau.
- `backend/sqlc/queries/reporting.sql:10-43` — Referenzsemantik der drei Summen:
  Umsatz = Zahlungen + Direktverkäufe − Direktverkauf-Storni − Warenrücknahmen;
  Stornierungen = Warenrücknahmen + Korrekturen + Direktverkauf-Storni;
  Geldtransit = Einlagen − Entnahmen.
- `backend/repository/kassenjournal_repo/repo.go:491-545` — bestehende Lesewege
  (`ReadEventsBySubject`, `ReadEventsByKassensitzung`).
- `backend/domain/kasse/storno_aufteilung_test.go`, `offene_arbeit_test.go` — Vorbild
  tabellengetriebener Tests; `backend/seed/seed_integration_test.go` — Vorbild Integrationstest.
- `backend/api/admin.go:117-124` — Wiring des Kasse-Commands (ReportingRepo entfällt).

### What to build

Eine reine Funktion in `domain/kasse` berechnet aus allen Events einer Kassensitzung die drei
Summen des Tagesabschluss-Events. Summen-wirksam sind Zahlung, Warenrücknahme, Korrektur,
Direktverkauf und Direktverkauf-Storno sowie Geldtransit; alle übrigen Event-Typen (Bestellung,
Umbuchung, Eröffnung, Kassensturz, Differenzbuchung, …) sind summen-neutral (siehe Resolved
decisions). Der Abschluss-Command verliert Interface und Feld des Reporting-Repositorys: Er liest
die Events der Sitzung über das Kassenjournal-Repository (events-only-Leseweg; die bestehende
Kassensitzungs-Query kann dafür einen Signatur-freien Zweig erhalten), hängt die im selben
Vorgang erzeugten Kassensturz- und Differenz-Events in-memory an und übergibt alles der
Domänenfunktion. Der Soll-Bestand kommt unverändert aus `GetKassenbestand`. Das Admin-Wiring
verdrahtet den Kasse-Command ohne `reporting_repo`; das Reporting-Modul selbst bleibt unberührt.

### Acceptance criteria

- [ ] `domain/kasse` enthält die Tagesabschluss-Aggregation als reine, exportierte Funktion ohne
      Repository- oder Kontext-Abhängigkeiten.
- [ ] Tabellengetriebene Unit-Tests decken die PRD-Fälle ab: leere Kassensitzung, Bestellungen
      ohne Zahlung, Zahlungen über mehrere Steuersätze, Warenrücknahme, geldneutrale Korrektur
      und Umbuchung (Summen unverändert), Direktverkauf samt Storno, Geldtransit in beide
      Richtungen, Differenzbuchung (summen-neutral).
- [ ] Ein Äquivalenztest belegt für dieselbe Kassensitzung Deckungsgleichheit der drei Summen mit
      dem Reporting-Ergebnis (Integrationstest über das Seed-Szenario oder gleichwertig gegen die
      echte SQL-Schicht).
- [ ] `api/kasse/application.Command` hat keine Reporting-Abhängigkeit mehr; der Command-Test
      verliert das Reporting-Mock und prüft, dass die Tagesabschluss-Event-Daten aus den
      Journal-Events (inkl. Same-Command-Events) berechnet werden.
- [ ] Kein Verhalten geändert: Endpunkt-Pfade, Antwortformate und Event-Daten identisch;
      `make verify` grün.

---

## Phase 2: Kasse-Kontext anlegen (kassenfuehrung, direktverkauf)

**User stories**: 1, 2

### Context

- `backend/api/kasse/` — wird zu `api/kasse/kassenfuehrung/` (inkl. `kassenabschluss_gate.go`
  und aller Tests).
- `backend/api/direktverkauf/` — wird zu `api/kasse/direktverkauf/` (unverändert).
- Wiring-Importe: `backend/api/admin.go:13-14`, `service.go:7-8`, `serviceleitung.go:7-10`.

### What to build

Der Kontext-Ordner `api/kasse/` entsteht; das heutige Modul `api/kasse` zieht als
`kasse/kassenfuehrung` hinein, `api/direktverkauf` als `kasse/direktverkauf`. Reine
`git mv`-Umzüge plus Import-Pfad-/Alias-Anpassung in den Wiring-Dateien; Paketnamen
(`application`, `http`) und alle Dateiinhalte bleiben sonst unverändert.

### Acceptance criteria

- [ ] `api/kasse/kassenfuehrung/{application,http}` und `api/kasse/direktverkauf/{application,http}`
      existieren; die alten Pfade sind weg.
- [ ] Alle Routen-Registrierungen unverändert (gleiche Pfad-Strings in den Wiring-Dateien).
- [ ] Tests sind mitgezogen und inhaltlich unverändert; `make verify` grün.

---

## Phase 3: God-Modul `api/table` auflösen

**User stories**: 2, 3, 4, 5, 13

### Context

- `backend/api/table/application/command.go` — Aufteilung: `:355-767` Tischgeschäft,
  `:253-285` Favoriten, `:287-353` Tisch-CRUD; geteilte Helfer (`:99-251`) gehen zum
  Tischgeschäft.
- `backend/api/table/application/query.go` — `GetAllTische:34` → stammdaten; übrige Queries
  (aktive Tische, State, Historie, Meine Tische, mit Favoriten) → Tischgeschäft.
- `backend/api/table/application/kassenbeleg_command.go` + `errors.go`-Anteile → `druck/beleg`.
- `backend/api/table/http/{command_handler,query_handler}.go` (+ Tests) — Handler analog
  aufteilen.
- `backend/api/table_bondruck_adapters.go` — entfällt;
  `backend/api/direktverkauf/application/command.go:35-36` und
  `backend/api/bondruck/application/arbeitsbon_policy.go:15-18,37-40` — Typumstellung auf
  `domain/druckstation`; `backend/seed/bondruck.go` zieht nach.
- Wiring: `backend/api/service.go:46-88`, `serviceleitung.go:29-38`, `admin.go:74-94`.

### What to build

`api/table` wird entlang der Kontexte aufgelöst, direkt an die Zielorte:

1. `kasse/tischgeschaeft`: Bestellen, Umbuchen, Kassieren, Storno, Ausgabe samt Tisch-Queries
   und geteilter Event-Write-Helfer. Dependencies: Tisch-Lesezugriff, Kassenjournal,
   Produkte, Favoriten (lesend für „Meine Tische" — bewusster Cross-Context-Read laut PRD),
   Kassensitzungen, Druckstationen, Druckaufträge.
2. `stammdaten/tisch`: Tisch-CRUD (Erstellen, Aktualisieren, Aktivieren, Deaktivieren, Löschen,
   Alle-Tische-Query) plus Favoriten-Commands. Dependencies: Tisch-Repository, Favoriten.
3. `druck/beleg`: der Kassenbeleg-Command mit Signaturstatus-Auflösung. Dependencies:
   Kassenjournal (lesend), Kassensitzungen, Druckstationen, Druckaufträge, Betreiber/
   Kassenidentität (bis Phase 5 noch settings_repo), Signaturauftrags-Leseweg.

Der Adapter-Shim entfällt: Die Consumer-Interfaces aller drei Nutzer (tischgeschaeft,
direktverkauf, beleg) verlangen `map[string]druckstation.Druckstation`; das Repository erfüllt
sie direkt. Die Arbeitsbon-Policy und der Kassenbeleg lesen `DruckerIP`/`Bonmodus` vom
Domain-Typ; der Typ `bondruck/application.Druckstation` wird gelöscht. Die Wiring-Dateien
registrieren dieselben Pfade auf die neuen Handler; jedes Command-Struct hat an jedem
Einsatzort einen vollständigen, konstanten Dependency-Satz (keine partielle Injektion mehr,
keine nil-Toleranzen wie `konfigurierteDruckstationen`/`KassenbelegDrucken`-Guard nötig).

### Acceptance criteria

- [ ] `api/table/` und `api/table_bondruck_adapters.go` existieren nicht mehr; die drei
      Zielmodule tragen die Logik unverändert.
- [ ] Routen identisch: `bestellung-aufnehmen`, `bestellung-umbuchen`, `zahlung-kassieren`,
      `ausgabe-bestaetigen`, `stornierung-erteilen`, `beleg-drucken`, `favorit-hinzufuegen`,
      `favorit-entfernen`, `create/update/activate/deactivate/delete-tisch`, `get-all-tische`,
      `get-aktive-tische`, `get-tisch-historie`, `get-tisch-state`,
      `get-aktive-tische-mit-favoriten`, `get-meine-tische-state`.
- [ ] Kein Command-Struct wird mehr partiell injiziert; die nil-Guards für fehlende Deps
      (`command.go:186-191`, `kassenbeleg_command.go:212-215`) sind entfernt — Tests, die bisher
      auf den Guard bauten, verdrahten stattdessen einen leeren Stub (einzige zulässige
      Test-Anpassung neben Import-Pfaden).
- [ ] Kein Modul außerhalb von `druck/` importiert mehr einen bondruck-Typ für Druckstationen;
      die Interfaces nutzen `domain/druckstation`.
- [ ] Bestehende Tests laufen an den neuen Orten inhaltlich unverändert; `make verify` grün.

---

## Phase 4: Druck-Kontext komplettieren

**User stories**: 1, 5

### Context

- `backend/api/bondruck/` (inkl. `application/escpos/`) → `api/druck/bondruck/`.
- `backend/api/druckauftrag/` → `api/druck/auftrag/`; `backend/api/druckstation/` →
  `api/druck/station/`; `backend/api/relay/` → `api/druck/relay/` (inkl.
  `relay_integration_test.go`).
- Import-Stellen: Wiring-Dateien, `druck/beleg` (escpos), `kasse/tischgeschaeft` und
  `kasse/direktverkauf` (Arbeitsbon-Policy), `backend/seed/bondruck.go`.

### What to build

Die vier verbliebenen Druck-Module ziehen unter `api/druck/`; `druckauftrag` und `druckstation`
heißen dabei `auftrag` und `station` (PRD-Modulnamen). Reine Umzüge mit
Import-/Alias-Anpassung; die Repository-Namen (`druckauftrag_repo`, `druckstation_repo`)
bleiben unverändert.

### Acceptance criteria

- [ ] `api/druck/{bondruck,beleg,auftrag,station,relay}` ist vollständig; keine Druck-Module
      mehr auf api-Top-Level.
- [ ] Alle Druck- und Relay-Routen unverändert (`get-druckstationen`, `update-druckstationen`,
      `get-fehlgeschlagene-druckauftraege`, `druckauftrag-erneut-versuchen`,
      `druckauftrag-verwerfen`, `/relay/poll`, `/relay/ergebnis`).
- [ ] Tests mitgezogen, `make verify` grün.

---

## Phase 5: settings aufspalten

**User stories**: 8, 11

### Context

- `backend/domain/settings/` — vier Dateien plus Test (Aufteilung siehe Architectural decisions).
- `backend/repository/settings_repo/{repo.go,types.go,repo_test.go}` — `:15-42` Betreiber,
  `:44-160` TSE/Kassenidentität.
- `backend/api/settings/application/{query.go,command.go,setup.go}` und `http/` — Aufteilung in
  `fiskal/setup` und `stammdaten/betreiber`; Routen `backend/api/admin.go:156-184`.
- Konsumenten der settings-Typen/-Repos: `api/kasse/kassenfuehrung` (Betreiber,
  TSE-Konfiguration), `api/druck/beleg` (Betreiber, Kassenidentität), `api/export` (Betreiber,
  Kassenidentität, TSE-Stammdaten), `app/tse_signatur_worker.go` (TSE-Konfiguration),
  `domain/dsfinvk/dsfinvk.go:14`, `backend/seed/{szenario,bondruck}.go`.

### What to build

`domain/settings` wird aufgelöst: `Betreiber` zieht in ein neues `domain/betreiber`;
Kassenidentität, TSE-Konfiguration und TSE-Stammdaten ziehen nach `domain/tse` (Typen dort ohne
TSE-Präfix-Stottern, siehe Resolved decisions). `settings_repo` teilt sich in ein neues
`betreiber_repo` und die bestehenden `tse_repo`-Erweiterungen; die sqlc-Queries bleiben
unverändert. Das api-Modul `settings` teilt sich in `fiskal/setup` (TSE-Einrichtung, -Übernahme,
-Status, Verbindungstest, TSE-Konfiguration, Kassenidentität; behält die
Kassensitzungen-Abhängigkeit und die Fiskaly-Client-Factories) und `stammdaten/betreiber`
(Betreiber lesen/schreiben). Alle Konsumenten-Interfaces werden auf die neuen Repos und
Domain-Pakete umgestellt; der Export erhält dafür zwei schmale Deps statt einer. Die
HTTP-Handler und Tests ziehen entsprechend aufgeteilt mit.

### Acceptance criteria

- [ ] `domain/settings`, `repository/settings_repo` und `api/settings` existieren nicht mehr;
      `domain/betreiber`, `betreiber_repo`, `fiskal/setup` und `stammdaten/betreiber` tragen die
      Logik unverändert.
- [ ] Alle zehn settings-Routen unverändert erreichbar (Zuordnung siehe Architectural decisions).
- [ ] `tse_repo` bündelt Kassenidentität, TSE-Konfiguration, TSE-Stammdaten und Einrichtung;
      `betreiber_repo` ausschließlich Betreiber.
- [ ] Keine Verhaltensänderung an Setup-Flow, Kassensitzungs-Eröffnungs-Warnung oder Beleg/Export;
      `make verify` grün.

---

## Phase 6: Fiskal-Kontext komplettieren

**User stories**: 8, 9

### Context

- `backend/api/tse/` → `api/fiskal/signatur/` (Monitoring-Query, Handler, Tests).
- `backend/app/{tse_signatur_worker.go,tse_rueckstand_watchdog.go}` (+ drei Testdateien) →
  `fiskal/signatur`; `backend/app/app.go:86-90` — Startstellen; Konstruktor-Signatur
  `tse_signatur_worker.go:104-110` (nutzt `cfg.FiskalyBaseURL`).
- `backend/api/export/` → `api/fiskal/export/`.
- `backend/domain/dsfinvk/` → `api/fiskal/dsfinvk/`; vorab `EventSignatur`
  (`signaturen.go:10-13`) → `domain/tse`; Nutzer: `kassenjournal_repo/repo.go:473,526`,
  `api/fiskal/export` (Interface `:32`), Mapper.

### What to build

Der Fiskal-Kontext wird vollständig: Das Monitoring-Modul zieht als `fiskal/signatur` um, und
Signatur-Worker plus Rückstand-Watchdog ziehen aus `app/` in dieses Modul (sie tragen
Fiskal-Logik: Fehlertaxonomie, Störungsprotokoll-Führung). `app/` konstruiert und startet beide
nur noch über exportierte Konstruktoren; die Fiskaly-Basis-URL bzw. Client-Factory wird als
Parameter gereicht, damit das application-Paket `config` nicht importiert. Der Export zieht als
`fiskal/export` um. `dsfinvk` verlässt die Domain-Schicht und wird ein reines Mapper-Paket
`fiskal/dsfinvk`; zuvor zieht `EventSignatur` nach `domain/tse`, sodass Kassenjournal-Repository
und Export-Interface auf den Domain-Typ zeigen und die Schichtung sauber bleibt
(Repository importiert kein api-Paket).

### Acceptance criteria

- [ ] `api/fiskal/{signatur,setup,export,dsfinvk}` ist vollständig; `api/tse`, `api/export` und
      `domain/dsfinvk` existieren nicht mehr; `app/` enthält keine TSE-Logik mehr, nur noch
      Lifecycle.
- [ ] `repository/kassenjournal_repo` importiert kein dsfinvk-Paket mehr; `EventSignatur` lebt in
      `domain/tse`.
- [ ] Kein application-Paket importiert `config`.
- [ ] Monitoring-, Export- und Worker-Verhalten unverändert (Routen `get-tse-signatur-queue`,
      `get-tse-stoerungen`, `export/dsfinvk`; Worker-Trigger und Advisory Lock unangetastet);
      Tests inkl. Worker-Integrationstest mitgezogen; `make verify` grün.

---

## Phase 7: Deutsche Namen und Stammdaten-Kontext komplettieren

**User stories**: 1, 4, 10

### Context

- `backend/domain/product/` → `domain/produkt/`; `backend/domain/table/` → `domain/tisch/`;
  Importer laut Inventar (u. a. `domain/kasse/bestellung.go`, Reporting-Query, Seed).
- `backend/repository/product_repo/` → `produkt_repo/`; `backend/repository/table_repo/` →
  `tisch_repo/` (Paketnamen entsprechend).
- `backend/api/product/` → `api/stammdaten/produkt/`; `backend/api/user/` →
  `api/stammdaten/user/`.
- `backend/sqlc/queries/tables.sql` → `tische.sql` (Query-Namen bereits deutsch), danach
  `make sqlc` (`sqlc/dbgen/` regeneriert sich, niemals von Hand editieren).

### What to build

Die Sprachkonsistenz-Renames: Domain-Pakete, Repositories und api-Module für Produkt und Tisch
heißen deutsch; `user` behält den englischen Namen (dokumentierte Ausnahme) und zieht nur unter
`stammdaten/`. Damit ist der Stammdaten-Kontext (`produkt`, `tisch`, `user`, `betreiber`)
vollständig und `api/` besteht top-level nur noch aus den fünf Kontext-Ordnern, der generischen
Infrastruktur und den Wiring-Dateien. Keine DB-, Routen- oder Verhaltensänderungen
(`tisch_sessions` bleibt bewusst, siehe PRD Out of Scope).

### Acceptance criteria

- [ ] Keine Go-Pakete `product`/`table` mehr im Backend; `domain/produkt`, `domain/tisch`,
      `produkt_repo`, `tisch_repo`, `stammdaten/{produkt,tisch,user,betreiber}` stehen.
- [ ] `tables.sql` heißt `tische.sql`; `make sqlc` gelaufen, generierter Code konsistent.
- [ ] Alle Produkt-/User-Routen unverändert; `make verify` grün.

---

## Phase 8: Deps-Bausatz für die Wiring-Dateien

**User stories**: 2, 12

### Context

- `backend/api/{admin,service,serviceleitung,auth,relay}.go` — je Datei eigene
  Repo-Konstruktionen (z. B. `service.go:31-44`); `backend/app/app.go:45-72` —
  `SetupRoutes` ruft die Bereichs-Konstruktoren, `:56` konstruiert `user_repo` zusätzlich für
  die JWT-Middleware.

### What to build

Ein gemeinsamer Abhängigkeiten-Bausatz (Struct im `api`-Paket) bündelt alle Repositories und
geteilten Bausteine. Er wird genau einmal konstruiert (in `app.SetupRoutes` bzw. einem
api-Konstruktor, den `SetupRoutes` aufruft) und an die Bereichs-Konstruktoren gereicht; deren
Signaturen ändern sich entsprechend. Die Bereichs-Wiring-Dateien bleiben als Komposition
erhalten (welche Handler mit welchen Deps auf welche Routen), konstruieren aber nichts mehr
selbst. Die Fiskaly-Client-Factories (Closures über die Basis-URL) gehören zum Bausatz oder
bleiben als cfg-abgeleitete Parameter — einmalig definiert.

### Acceptance criteria

- [ ] Jede `New…Repository`-Konstruktion existiert im Produktions-Wiring genau einmal.
- [ ] Die fünf Bereichs-Konstruktoren beziehen alle Deps aus dem Bausatz; Routen und Verhalten
      unverändert.
- [ ] `make verify` grün.

---

## Phase 9: Frontend-Feinschnitt

**User stories**: 14, 15, 16

### Context

- `frontend/src/lib/EinstellungenBackend.ts` — Aufteilung laut Inventar (`:192-206` Betreiber/
  Kassenidentität, `:209-296` TSE inkl. Schemas/Typen im Dateikopf).
- `frontend/src/lib/DruckstationBackend.ts` (+ `.test.ts`) → `admin/settings/`.
- `frontend/src/admin/settings/hooks.ts` — gemischte Hooks (Druckstationen, TSE, Betreiber);
  TSE-/Betreiber-Hooks ziehen zu ihren Slices.
- Importstellen: `admin/tse/{TSEEinrichtungWizard,TSEKonfigurationSection}.tsx`,
  `admin/finanzamt/{BetreiberSection,SignaturauftraegeSection,TSEAusfalldokumentationSection}.tsx`.
- `frontend/src/admin/kasse/KassensitzungPage.tsx:106,240,415` — drei Sektions-Funktionen;
  `KassensitzungPage.test.tsx`.

### What to build

lib enthält danach nur Generisches (BackendClient, Auth, identity, download, utils,
errorMessages, arbeitsmodus). Die Einstellungen-Sammelklasse wird aufgeteilt in eine
TSE-Backend-Klasse (Konfiguration, Setup, Status, Signatur-Queue, Störungsprotokoll) im
tse-Slice und eine Betreiber-Backend-Klasse (Betreiber, Kassenidentität) im Finanzamt-Slice;
die Schemas/Typen ziehen mit. Die Finanzamt-Sektionen importieren die TSE-Klasse cross-slice
(gewollt laut PRD, die Seite bleibt unverändert). DruckstationBackend zieht zu den
Druck-Einstellungen; die Hooks folgen ihren Klassen. Die Kassensitzungs-Seite wird in ihre drei
Sektions-Dateien (Eröffnung, Geldtransit, Abschluss) zerlegt; Verhalten, Markup und Texte
bleiben identisch. Keine Änderungen an Routen, API-Aufrufen oder dem Typen-Schnitt
admin/service.

### Acceptance criteria

- [ ] `lib/` enthält keine admin-exklusive Backend-Klasse mehr; `EinstellungenBackend.ts`
      existiert nicht mehr.
- [ ] TSE- und Betreiber-Backend-Klasse leben in ihren Slices; alle Sektionen funktionieren
      unverändert (gleiche URLs, gleiche Oberfläche).
- [ ] `KassensitzungPage.tsx` importiert drei Sektions-Dateien; Seiten-Test bleibt inhaltlich
      unverändert grün.
- [ ] Frontend-Checks grün (`make check` deckt Lint, Typecheck, Tests ab).

---

## Phase 10: Dokumentation angleichen

**User stories**: 17, 18

### Context

- `docs/handbuch.md:31-53` (2.1/2.2), `:41` (Bondruck-Policy-Satz), `:211-216` (3.12),
  `:217-252` (3.13 TSE-Pfade), `:284-303` (4.6), `:361-371` (6.1 Schichten).
- `docs/language.md:17` (Namenskonventionen), `:85-103` (Kasse-Begriffe, domain/table-Hinweis),
  `:325` (EntityStatus-Paketliste).
- `.github/instructions/event-sourcing.instructions.md:3` (applyTo),
  `backend.instructions.md:79,124,162,165` (product-Beispiele).
- `AGENTS.md:103-107` (Bereiche).

### What to build

Die Dokumentation wird dem Code angeglichen: Kontextübersicht mit sechs Subdomänen (Kasse als
Core; Fiskalisierung, Druck/Ausgabe, Stammdaten, Reporting als Supporting; Auth als Generic)
samt Beziehungen und Klassifikations-Begründung; der Bondruck-Policy-Satz wird durch die
Druck-Kontext-Beschreibung ersetzt; Schichten- und Modulbeschreibung nennen die Kontext-Ordner;
die TSE-Abschnitte verweisen auf `fiskal/`-Pfade. language.md erhält die
Go-Paketnamens-Konvention (deutsch für Fachmodule, englisch für Infrastruktur, `user` als
Ausnahme), die Modulbegriffe Tischgeschäft und Kassenführung und aktualisierte Paket-Hinweise.
Die Instructions-applyTo-Muster zeigen auf die neuen Pfade (`domain/tisch`), die Codebeispiele
nutzen `produkt.*`. Die AGENTS.md-Bereichsbeschreibung wird auf die Kontext-Ordner geprüft und
angepasst. Stil laut Doku-Konvention schlicht halten (kein Format-Overhead).

### Acceptance criteria

- [ ] handbuch.md beschreibt sechs Subdomänen; die Policy-Aussage zu Bondruck ist ersetzt; alle
      Pfad-Referenzen (Worker, dsfinvk, export, settings) zeigen auf existierende Orte.
- [ ] language.md dokumentiert die Paketnamens-Konvention und die neuen Modulbegriffe; keine
      Referenz auf `domain/table`/`domain/product` mehr.
- [ ] Instructions-applyTo-Muster matchen die neuen Pfade; kein Beispiel nutzt alte Paketnamen.
- [ ] Kein Doku-Verweis im Repo zeigt mehr auf gelöschte Pfade (Stichproben: `api/table`,
      `api/settings`, `domain/settings`, `domain/dsfinvk`, `app/tse_signatur_worker`).
