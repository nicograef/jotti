# Plan: TSE-Admin auf Monitoring vereinfachen

> Source PRD: [docs/prds/prd-tse-admin-vereinfachung.md](../prds/prd-tse-admin-vereinfachung.md)

## Goal

Die Signaturauftrags-Verwaltung wird ersatzlos zurückgebaut: kein Zurücksetzen, kein Verwerfen, keine
Einzelauftrags-Liste, kein Status `verworfen`. Übrig bleibt Monitoring mit minimaler Diagnose (grobe Warnung
plus letzter Fehlertext), dessen fehlgeschlagen-Warnung mit dem Kassenabschluss endet. Reparatur nach einem
Bugfix ist ein dokumentiertes SQL-Runbook. Die Audit-Befunde B1, B2 und B3 entfallen damit strukturell.

## Architectural decisions

- **Schema**: `tse_signaturauftraege.status` CHECK ohne `verworfen`; die Spalten `verworfen_grund`,
  `verworfen_von`, `verworfen_am` entfallen (Änderung direkt in `01_initial.up.sql`, Dev-Reset). Endstatus sind
  `fehlgeschlagen` und `tse_nicht_konfiguriert`. Der Druckauftrags-Status `verworfen`
  (`database/migrations/01_initial.up.sql:283`) ist ein eigener Begriff und bleibt unberührt.
- **Routen**: Die `signaturauftrag`-Familie schrumpft auf zwei Lesewege: `/admin/get-tse-signatur-queue` und
  `/admin/get-tse-stoerungen`. `get-tse-signaturauftraege` und die drei Mutations-Endpunkte entfallen ersatzlos.
- **Queue-Zustand (Antwortform)**: `offeneAuftraege`, `rueckstandSekunden`, `signaturenProMinute`,
  `signierdauerP95Sekunden` bleiben global wie bisher. `fehlgeschlageneAuftraege` zählt nur noch Aufträge, deren
  Event zur aktiven Kassensitzung gehört (`offen` oder `wird_abgeschlossen`); ohne aktive Sitzung 0 — der
  Kassenabschluss quittiert die Warnung. Neu: `letzterFehler` (Fehlertext des jüngsten fehlgeschlagenen
  Auftrags der aktiven Sitzung, sonst leer).
- **Ausfallbegriff**: `BestimmeSignaturstatus` kennt als Endstatus nur noch `fehlgeschlagen` und
  `tse_nicht_konfiguriert`; Gate- und Beleg-Semantik sind sonst unverändert.
- **Reparaturpfad**: dokumentiertes SQL im Handbuch (Status zurück auf `offen`, Versuche nullen), keine UI.

## Inventory

Schema und Domain:

- `database/migrations/01_initial.up.sql:485` — Status-CHECK mit `verworfen`; `:491-493` Verwerfen-Spalten;
  `:512,517-519` zugehörige Kommentare.
- `backend/domain/tse/signaturauftrag.go:11` — `StatusVerworfen`.
- `backend/domain/tse/signaturstatus.go:65` — `verworfen` als Ausfall-Endstatus; Tests
  `signaturstatus_test.go`.

Backend Verwaltung (entfällt):

- `backend/sqlc/queries/tse_signaturauftraege.sql:84-120` — `GetTSESignaturauftraege`,
  `TSESignaturauftragZuruecksetzen`, `TSESignaturauftraegeZuruecksetzenGesamt`, `TSESignaturauftragVerwerfen`;
  nach dem Entfernen `make sqlc`.
- `backend/repository/tse_repo/repo.go:33-48` — Struct `Signaturauftrag`; `:144-209` Admin-Methoden;
  Tests `repo_test.go:294-356,427-527` (Zurücksetzen/Verwerfen/Status-Guards) und `:358-425`
  (BleibtZuruecksetzbar-Anteil).
- `backend/api/tse/application/command.go` — komplett (nur die drei Admin-Commands).
- `backend/api/tse/http/handler.go:17-73` — Auftrags-Liste im Query-Teil; `:138-223` Command-Teil;
  `handler_test.go` entsprechend.
- `backend/api/admin.go:155,158-160` — Routen-Registrierung.

Backend Monitoring (bleibt, wird angepasst):

- `backend/sqlc/queries/tse_signaturauftraege.sql:59-72` — `GetTSESignaturQueueZustand` (wird sitzungsbezogen
  für `fehlgeschlagen`, plus `letzter_fehler`).
- `backend/repository/tse_repo/repo.go:50-61,211-225` — `SignaturQueueZustand`; Test `repo_test.go:528-569`.
- `backend/api/tse/application/query.go`, `backend/api/tse/http/handler.go:75-100` — Queue-DTO.
- `backend/api/kasse/application/kassenabschluss_gate.go:27-35` — Kommentar nennt „verworfene" Aufträge
  (Text anpassen, Logik unverändert).

Seed und Tests:

- `backend/seed/faketse.go:32-35` (`fehlschlagJederNte`-Kommentar), `:198` (`verworfenVergeben`),
  `:306-327` (`dauerhaftGescheitert`) — Verworfen-Dramaturgie entfällt, dauerhaft gescheiterte bleiben
  `fehlgeschlagen`.
- `backend/seed/seed_integration_test.go:177-189` — erwartet vier Status und genau einen verworfenen
  (Achtung: `:236` prüft dasselbe für Druckaufträge, bleibt unberührt).
- `backend/seed/bondruck.go:169` — Druckauftrags-`verworfen`, nicht anfassen.

Frontend:

- `frontend/src/admin/finanzamt/SignaturauftraegeSection.tsx` — Liste, Aktionen, Verwerfen-Dialog entfallen;
  die Kennzahlen-Karte (`QueueZustand`, Z. 66-90) bleibt als Kern der neuen Monitoring-Karte.
- `frontend/src/lib/EinstellungenBackend.ts:148-167` — `TSESignaturauftragSchema`; `:299-345` — Methoden
  (`getTSESignaturauftraege`, Zurücksetzen x2, Verwerfen); `TSESignaturQueueSchema:172-179` erhält
  `letzterFehler`.
- `frontend/src/admin/settings/hooks.ts:154-194` — `useTSESignaturauftraege` entfällt.
- `frontend/src/admin/reporting/AdminDashboardPage.tsx:30-40` — Warnlogik konsumiert die neue
  sitzungsbezogene Semantik unverändert; Warntext um Fehlertext ergänzen.

Doku und PRD:

- `docs/language.md:418` — Signaturauftrag-Eintrag (Statusliste), Verwerfen-Bezüge in den TSE-Begriffen.
- `docs/handbuch.md:227-231` — §3.13 Signaturauftrag- und Störungsprotokoll-Absätze (Verwerfen, Statusliste);
  Ort des neuen Runbook-Absatzes.
- `docs/prds/prd-tse-signatur-outbox.md:94-121` — User Stories 6, 9, 12, 13 (Revisionsvermerk).
- `docs/plans/plan-tse-signatur-outbox.md` — Phase-6-Beschreibung bleibt als historisches Dokument unberührt
  (wird nach dem Merge ohnehin gelöscht).

## Resolved decisions

- Diagnose minimalistisch: nur grobe Warnung plus letzter Fehlertext, keine Einzelauftrags-Ansicht
  (Entwickler-Entscheid, 2026-07-04).
- fehlgeschlagen-Warnung endet mit dem Kassenabschluss (sitzungsbezogene Zählung); der Abschluss weist die
  Reste aus und quittiert damit.
- `tse_nicht_konfiguriert` ist endgültig, ohne Wiedereinreihung: keine TSE konfiguriert heißt keine TSE für
  diesen Zeitraum.
- Umsetzung auf diesem Branch (`feat/tse-signatur-outbox-phase1`), drei Phasen (bestätigt).
- Verwerfen wird nicht durch eine Quittier-Funktion ersetzt; das Quittieren übernimmt der Kassenabschluss.
- Reparatur nach Bugfix als SQL-Runbook im Handbuch, bewusst keine UI.
- Der Druckauftrags-Status `verworfen` bleibt unberührt (eigener Begriff der Druckverwaltung).

## Open questions / Risks

- Zwischenstand nach Phase 1: Die Finanzamt-Karte zeigt die globale fehlgeschlagen-Zahl noch ohne
  Sitzungsbezug und ohne Fehlertext; Phase 2 zieht die Semantik nach. Unkritisch, da beide Phasen auf
  demselben Branch direkt aufeinander folgen.

---

## Phase 1: Verwaltung entfernen

**User stories**: PRD 1 (Grundstein), 3; revidiert US 6/9/13b des Alt-PRDs

### Context

- `database/migrations/01_initial.up.sql:485,491-493,512,517-519` — Schema-Rückbau.
- `backend/domain/tse/signaturauftrag.go:11`, `signaturstatus.go:65` — Statusmodell.
- `backend/sqlc/queries/tse_signaturauftraege.sql:84-120`, `backend/repository/tse_repo/repo.go:33-48,144-209`,
  `backend/api/tse/application/command.go`, `backend/api/tse/http/handler.go:17-73,138-223`,
  `backend/api/admin.go:155,158-160` — Verwaltungs-Stack.
- `frontend/src/admin/finanzamt/SignaturauftraegeSection.tsx`, `frontend/src/lib/EinstellungenBackend.ts`,
  `frontend/src/admin/settings/hooks.ts:154-194` — UI-Rückbau.
- `backend/seed/faketse.go:198,306-327`, `backend/seed/seed_integration_test.go:177-189` — Seed-Dramaturgie.

### What to build

Der komplette Rückbau der Verwaltung in einem Schnitt: Schema ohne `verworfen` (CHECK, drei Spalten,
Kommentare), Domain ohne `StatusVerworfen` (Signaturstatus-Funktion kennt zwei Endstatus), Backend ohne die
vier Verwaltungs-Endpunkte samt Application-Command, Handler-Command-Teil, Repo-Methoden, Queries und Tests.
Im Frontend wird die `SignaturauftraegeSection` zur reinen Monitoring-Karte (die bestehenden
Queue-Kennzahlen bleiben unverändert sichtbar); Auftrags-Liste, Zurücksetzen- und Verwerfen-Aktionen samt
Dialog, Backend-Methoden, Zod-Schema und Hook entfallen. Der Seed verliert die Verworfen-Dramaturgie
(dauerhaft gescheiterte Aufträge bleiben `fehlgeschlagen`); der Integrationstest erwartet drei statt vier
Status. Beleg-, Gate- und Worker-Logik bleiben unverändert (nur der Gate-Kommentar verliert das Wort
„verworfen").

### Acceptance criteria

- [x] Schema: CHECK ohne `verworfen`, Spalten `verworfen_grund/von/am` entfernt, Kommentare angepasst;
      `make sqlc` regeneriert; Druckauftrags-Status unberührt.
- [x] `StatusVerworfen` existiert nicht mehr; `BestimmeSignaturstatus` liefert Ausfall genau für
      `fehlgeschlagen`, `tse_nicht_konfiguriert` und offen bei aktiver Störung (Tests angepasst).
- [x] Die Endpunkte `get-tse-signaturauftraege`, `tse-signaturauftrag-zuruecksetzen`,
      `tse-signaturauftraege-zuruecksetzen`, `tse-signaturauftrag-verwerfen` sind samt Application-Command,
      Handler, Repo-Methoden, Queries und Tests entfernt; `admin.go` registriert nur noch
      `get-tse-signatur-queue` und `get-tse-stoerungen`.
- [x] Finanzamt-Seite zeigt eine reine Monitoring-Karte (Kennzahlen wie bisher); Liste, Aktionen, Dialog,
      `useTSESignaturauftraege`, Backend-Methoden und `TSESignaturauftragSchema` sind entfernt.
- [x] Seed erzeugt keine verworfenen Aufträge mehr; Seed-Integrationstest prüft die drei verbleibenden
      Problemlos-/Problem-Status.
- [x] `make verify` grün.

---

## Phase 2: Minimal-Diagnose mit Abschluss-Quittierung

**User stories**: PRD 1, 2; revidiert US 12 des Alt-PRDs

### Context

- `backend/sqlc/queries/tse_signaturauftraege.sql:59-72` — Queue-Zustand-Query.
- `backend/repository/tse_repo/repo.go:50-61,211-225`, `repo_test.go:528-569` — Repo-Sicht.
- `backend/api/tse/http/handler.go:75-100`, `backend/api/tse/application/query.go` — DTO.
- `frontend/src/lib/EinstellungenBackend.ts:172-179` — Zod-Schema.
- `frontend/src/admin/reporting/AdminDashboardPage.tsx:30-70` — Warnlogik und -texte.
- Monitoring-Karte aus Phase 1 — Anzeigeort des Fehlertexts.

### What to build

Der Queue-Zustand wird zur minimalen Diagnose: `fehlgeschlageneAuftraege` zählt nur noch Aufträge, deren Event
zur aktiven Kassensitzung gehört (Join über `kassenjournal.kassensitzung_nr`; ohne aktive Sitzung 0), und die
Antwort trägt neu `letzterFehler` (Fehlertext des jüngsten fehlgeschlagenen Auftrags der aktiven Sitzung,
sonst leer). Dashboard-Warnung und Monitoring-Karte zeigen damit genau die gewünschte grobe Information: es
stimmt etwas nicht, plus Fehlertext; nach dem Kassenabschluss (der die Ausfall-Reste ausweist) verschwindet
die Warnung von selbst. Alle übrigen Kennzahlen bleiben unverändert global.

### Acceptance criteria

- [x] Queue-Query zählt `fehlgeschlagen` sitzungsbezogen und liefert `letzter_fehler` des jüngsten
      fehlgeschlagenen Auftrags der aktiven Sitzung (Repo-Test: mit aktiver Sitzung, ohne aktive Sitzung,
      Vorfall aus alter Sitzung zählt nicht).
- [x] Nach dem Kassenabschluss verschwindet die fehlgeschlagen-Warnung ohne weitere Aktion (Test).
- [x] Antwort-DTO und Zod-Schema um `letzterFehler` erweitert; Dashboard-Warntext und Monitoring-Karte zeigen
      Anzahl und Fehlertext.
- [x] Rückstands- und Konfigurationswarnung unverändert; `make verify` grün.

---

## Phase 3: Dokumentation und PRD-Revision

**User stories**: PRD 4; flankiert die Nachweisbarkeit

### Context

- `docs/language.md:418` — Signaturauftrag-Statusliste; Verwerfen-Bezüge.
- `docs/handbuch.md:227-231` — §3.13 (Signaturauftrag, Störungsprotokoll, Admin-Bezüge).
- `docs/prds/prd-tse-signatur-outbox.md:94-121` — User Stories 6, 9, 12, 13.
- `docs/compliance.md` §3.8 — erwähnt keine Verwaltungsfunktionen; nur gegenlesen.

### What to build

Die Doku zieht nach: language.md führt den Signaturauftrag ohne `verworfen` und ohne Verwaltungs-Begriffe;
handbuch.md §3.13 beschreibt Monitoring statt Verwaltung und erhält den Runbook-Absatz „Reparatur nach
Bugfix" mit dem dokumentierten SQL (fehlgeschlagene Aufträge zurück auf `offen`, Versuche nullen; Hinweis,
dass der Worker sie danach regulär signiert). Das Alt-PRD erhält bei den User Stories 6, 9, 12 und 13 einen
kurzen Revisionsvermerk mit Verweis auf das Mini-PRD. Endpunkt- und UI-Texte werden gegen die Begriffe
geprüft.

### Acceptance criteria

- [ ] `docs/language.md`: Statusliste ohne `verworfen`; keine Verwaltungs-Begriffe (Zurücksetzen/Verwerfen)
      mehr in den TSE-Einträgen.
- [ ] `docs/handbuch.md` §3.13: Monitoring statt Verwaltung beschrieben; Runbook-Absatz mit SQL vorhanden.
- [ ] `docs/prds/prd-tse-signatur-outbox.md`: Revisionsvermerk bei US 6, 9, 12, 13 mit Verweis auf
      [prd-tse-admin-vereinfachung.md](../prds/prd-tse-admin-vereinfachung.md).
- [ ] Keine `verworfen`-/Verwaltungs-Reste in UI-Texten und Endpunkt-Namen (Grep-Sweep, Druckaufträge
      ausgenommen).
