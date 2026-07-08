# Plan: Praxistest-Fixes Runde 1 (Drawer-Layout, Fehler-Referenz, Druck-Backoff)

> Quell-PRD: [prd-praxistest-fixes.md](../prds/prd-praxistest-fixes.md).
> Ausgeliefert als vorgezogenes Release von main (Arbeitstitel v0.14.1;
> Migration und neuer Endpoint sprechen beim Schnitt eher für ein Minor).
> Arbeitsdokument, nach Abarbeitung aus `docs/plans/` entfernen.

## Ziel

Die vier verifizierten Praxistest-Befunde vom 07.07.2026 beheben:

1. Drawer-Listen scrollen nativ, Kommentarfeld und Buttons bleiben immer
   bedienbar, lange Namen werden abgeschnitten (Phasen 1 bis 3, Abnahme in
   Phase 7).
2. Serverfehler werden supportbar: Der generische 500er-Toast nennt die
   Correlation-ID als Fehler-Referenz (Phase 4). Panic-Recovery und
   idempotenter Kassenabschluss-Wiederanlauf liegen bereits auf main und
   kommen mit dem Release automatisch mit (User Stories 9, 10).
3. Druckaufträge überbrücken einen Rollenwechsel per Backoff-Nachdruck
   statt nach 6 Sekunden endgültig zu scheitern (Phase 5); für lange
   unbemerkte Störungen gibt es einen Sammel-Retry (Phase 6). Der
   Dashboard-Hinweis auf fehlgeschlagene Aufträge (User Story 18) liegt
   bereits auf main.

## Architekturentscheidungen

- **Routen:** Ein neuer Admin-Endpoint `admin/druckauftraege-erneut-versuchen`
  (Sammel-Retry, Plural analog zum bestehenden Singular-Endpoint
  `admin/druckauftrag-erneut-versuchen`). Alle übrigen Endpoints bleiben
  unverändert; das Relay-Protokoll (`relay/poll`, `relay/ergebnis`) ändert
  sich nicht.
- **Schema:** Migration `02_druckauftrag_backoff.up.sql` (erste echte
  Upgrade-Migration seit dem De-facto-Freeze): additiv, transaktional,
  fügt `druckauftraege.naechster_versuch_ab TIMESTAMPTZ NULL` hinzu
  (NULL = sofort fällig) und aktualisiert die Spalten-Kommentare
  (Endzustand nach 6 statt 3 Fehlversuchen). Kein neuer Index; der
  bestehende `idx_druckauftraege_status_id` reicht bei diesen Volumina.
- **Schlüssel-Komponenten:** `PositionAuswahlListe` als gemeinsame
  Service-Komponente (`frontend/src/service/components/`), controlled:
  bekommt Positionen (minimale Form: ID, Name, Einzelpreis, Maximalmenge),
  `mengen` und `onAdd`/`onRemove` von außen; die Mengenlogik bleibt in den
  Drawern (Mengen-Hook bzw. lokaler State der Umbuchung mit
  Voll-Vorauswahl). Sie kapselt den Scrollbereich (natives
  `overflow-y-auto`, dvh-basierte Maximalhöhe) und die Zeile
  (Name mit truncate, Einzelpreis, Minus/Anzahl/Plus).
  `Receipt` behält seine Schnittstelle und stellt intern auf natives
  Scrollen um. Die ScrollArea-UI-Komponente wird danach entfernt.
- **Backoff:** Lebt vollständig im Backend (`druckauftrag_repo`); das Relay
  bleibt unverändert (pollt alle 2 s, druckt, meldet). Reine Funktion
  Fehlversuchsnummer zu Wartezeit (5s, 15s, 30s, 60s, 180s);
  `MaxDruckversuche` steigt von 3 auf 6. Die Fälligkeit wird DB-seitig aus
  `NOW()` plus übergebener Wartezeit gesetzt (konsistent mit
  `erstellt_am`, keine App-Uhr).
- **Fehler-Referenz:** Transport über den bereits auf jeder Antwort
  gesetzten `X-Correlation-ID`-Header; kein neues Feld im Fehler-Body,
  keine Änderung an den 61 `SendServerError`-Aufrufstellen. Das Frontend
  liest den Header im Fehlerpfad und hängt ihn als Referenz an die
  generische Serverfehler-Meldung an.

## Inventar

Drawer/ScrollArea:

- `frontend/src/service/components/table/HistorieStornierungDrawer.tsx:98-152` — ScrollArea `max-h-72` plus duplizierter Zeilenblock; Kommentar Pflicht (Zeile 58)
- `frontend/src/service/components/table/HistorieUmbuchungDrawer.tsx:57-98,143-197` — dritter Duplikat-Block; eigener Mengen-State mit Voll-Vorauswahl statt `useMengen`
- `frontend/src/service/components/direktverkauf/DirektverkaufStornoDrawer.tsx:100-155` — Duplikat-Block mit aria-Labels an Plus/Minus (Vorbild für die gemeinsame Komponente)
- `frontend/src/service/components/table/Receipt.tsx:19` — ScrollArea `max-h-96`; Nutzer: `ZahlungDrawer.tsx:111`, `AusgabeDrawer.tsx:91`, `BestellungDrawer.tsx:98`, `TischHistorie.tsx:458`
- `frontend/src/service/components/TischAuswahlDrawer.tsx:81` — Vorbild natives Scrollen (`overflow-y-auto max-h-[60vh]`), bleibt unverändert
- `frontend/src/components/ui/scroll-area.tsx` — Radix-Wrapper, wird am Ende entfernt
- `frontend/src/hooks/use-mengen.ts` — Mengen-Hook (add/remove/Grenzen), bleibt
- `frontend/src/service/components/table/drawerUtils.ts` — `selectPositionen`, `calculateTotalPrice`, `toPositionRefs`, `toReceiptItems`
- `frontend/src/service/components/direktverkauf/DirektverkaufHistorie.test.tsx:17-34` — jsdom-Mock für ScrollArea, wird mit der Umstellung obsolet

Fehler-Referenz:

- `backend/api/middleware/middleware.go:46-58` — CorrelationIDMiddleware setzt `X-Correlation-ID` auf jede Antwort; Tests `middleware_test.go:51-69`
- `backend/api/helper/http.go:95-97` — `SendServerError` (Body bleibt unverändert)
- `frontend/src/lib/Backend.ts:110-153` — `throwIfNotOk` baut `BackendError(status, code, details)`
- `frontend/src/lib/errorMessages.ts:3-4,104-106` — generische Serverfehler-Meldung, 500er-Pfad; Tests `errorMessages.test.ts`

Druck:

- `database/migrations/01_initial.up.sql:284-304` — Tabelle `druckauftraege` (Status-Kette, `versuche`, Kommentare)
- `backend/sqlc/queries/relay.sql:1-38` — alle Druckauftrags-Queries (offen, gedruckt, Fehlversuch, fehlgeschlagen, Retry, Discard)
- `backend/repository/druckauftrag_repo/repo.go:14,104-151,177-189` — `MaxDruckversuche = 3`, Poll-Abfrage, Ergebnis-Transaktion, Einzel-Retry/Discard; Integrationstests `repo_test.go:15-31` (Build-Tag `integration`)
- `backend/api/druck/auftrag/application/command.go` und `.../http/handler.go` — Einzel-Retry/Discard-Command samt Handler-Tests (Vorbild für Sammel-Retry)
- `backend/api/admin.go:121-127` — Registrierung der Druckauftrags-Endpoints
- `windows/relay/main.go:38,213-264` — Poll-Takt 2 s, Gruppierung nach Ziel-IP, ID-Reihenfolge je Gruppe (deckt User Story 14 ab)
- `frontend/src/admin/settings/DruckstationBackend.ts:97-103`, `hooks.ts:42-56`, `DruckstationConfigPage.tsx:148-250` — Einzel-Retry/Verwerfen-UI; Hinweistext "nach drei Versuchen" (Zeile 248) muss mitziehen
- `frontend/src/admin/users/UserItem.tsx:4` — AlertDialog-Bestätigungsmuster im Admin

Migrations-/Release-Harness:

- `database/migrations/README.md` — Forward-only-Regeln (additiv, transaktional, Nummerierung)
- `.github/workflows/ci.yml:398-470` — Upgrade-Path-Job (Vorversion v0.14.0 gepinnt); wird mit der ersten `02_`-Migration zum Pflicht-Gate
- `Makefile:47-71,146,295` — `test`, `test-frontend`, `test-integration`, `lint`, `sqlc`, `verify`

## Entschieden (Klärungsrunde 08.07.2026)

- Fehler-Referenz über den `X-Correlation-ID`-Response-Header, nicht über
  den 500-Body; der Backend-Beleg ist ein Test, dass 500-Antworten den
  Header tragen.
- Playwright-Abnahme als eigene Schlussphase (ein Seed-Setup, volle Matrix
  über alle fünf Stellen); die Drawer-Phasen sichern sich über
  Komponententests ab.
- Phasenschnitt mit 7 Phasen bestätigt.

> **Assumption** (aus dem PRD übernommen): Der konkrete Auslöser des
> Kassenabschluss-500ers bleibt ungeklärt; die Maßnahme ist Härtung plus
> Supportbarkeit, kein Root-Cause-Fix.

## Offene Punkte / Risiken

- Bildschirmtastatur auf iOS (User Story 8): dvh-Maximalhöhe und
  vaul-Drawer müssen mit geöffneter Tastatur zusammenspielen. Playwright
  kann die Tastatur nicht simulieren; Phase 7 nähert das über reduzierte
  Viewport-Höhe an, die endgültige Bestätigung ist ein manueller
  Geräte-Check beim nächsten Praxiseinsatz.
- Bestandsdaten: Bereits fehlgeschlagene Aufträge bleiben nach der
  Migration fehlgeschlagen (kein Auto-Requeue); offene Aufträge mit
  `versuche < 3` profitieren automatisch vom neuen Limit. Gewollt.
- Der Release-Schnitt selbst (Tag, Versionsnummer, `PREVIOUS_VERSION`-Bump
  im Upgrade-Path-Job) folgt der bestehenden Release-Mechanik und ist
  nicht Teil dieses Plans.

---

## Phase 1: PositionAuswahlListe + Stornierungs-Drawer

**User Stories:** 1, 2, 3 (8 anteilig über die dvh-Maximalhöhe)

### Kontext

- `frontend/src/service/components/table/HistorieStornierungDrawer.tsx:98-152` — abzulösender ScrollArea-Block
- `frontend/src/service/components/TischAuswahlDrawer.tsx:81` — Vorbild natives Scrollen
- `frontend/src/service/components/direktverkauf/DirektverkaufStornoDrawer.tsx:126,143` — aria-Labels, die die gemeinsame Komponente übernimmt
- `frontend/src/hooks/use-mengen.ts`, `frontend/src/service/components/table/drawerUtils.ts` — bleiben unverändert
- Vorhandene Drawer-/StickyActionBar-Tests im Service-Bereich als Test-Vorbild

### Was zu bauen ist

Neue Komponente `PositionAuswahlListe` unter
`frontend/src/service/components/`: rendert die Positionszeilen (Name mit
truncate, Einzelpreis und Bestellmenge, Minus/Anzahl/Plus mit
aria-Labels) in einem nativen Scrollcontainer (`overflow-y-auto`,
Maximalhöhe dvh-basiert statt `max-h-72`), controlled über `mengen` und
`onAdd`/`onRemove`. Der Stornierungs-Drawer stellt auf die Komponente um;
Header, Summe, Kommentarfeld und Footer-Buttons liegen außerhalb des
Scrollbereichs. Damit ist der Praxistest-Blocker (Stornierung großer
Bestellungen) behoben.

### Akzeptanzkriterien

- [ ] Komponententests (Vitest + Testing Library): Zeilen rendern,
      Plus/Minus melden sich controlled, Anzahl wird angezeigt,
      aria-Labels vorhanden
- [ ] Stornierungs-Drawer nutzt `PositionAuswahlListe`; kein
      ScrollArea-Import mehr in dieser Datei; Mengen-Grenzen (0 bis
      Bestellmenge) wirken weiter über `useMengen`
- [ ] Kommentarfeld und "Stornierung erteilen" liegen außerhalb des
      Scrollbereichs
- [ ] `make test-frontend` und `make lint` grün

---

## Phase 2: Umbuchungs- und Direktverkauf-Storno-Drawer umstellen

**User Stories:** 4, 5

### Kontext

- `frontend/src/service/components/table/HistorieUmbuchungDrawer.tsx:57-98,143-197` — eigener Mengen-State mit Voll-Vorauswahl (Default: alles ausgewählt) und `onAdd` mit Obergrenze; bleibt im Drawer, die Komponente ist controlled
- `frontend/src/service/components/direktverkauf/DirektverkaufStornoDrawer.tsx:100-155` — dritter Duplikat-Block
- `frontend/src/service/components/direktverkauf/DirektverkaufHistorie.test.tsx:17-34` — ScrollArea-jsdom-Mock, nach der Umstellung entfernen

### Was zu bauen ist

Beide Drawer stellen auf `PositionAuswahlListe` um; der dreifach
duplizierte Zeilenblock ist damit vollständig konsolidiert. Die
Umbuchung behält ihre Voll-Vorauswahl (State bleibt lokal), der
Direktverkauf-Storno seinen `useMengen`-Hook. Ziel-Tisch-Auswahl,
Summen, Kommentar und Buttons bleiben außerhalb des Scrollbereichs.

### Akzeptanzkriterien

- [ ] Keine ScrollArea-Nutzung mehr in den drei Auswahl-Drawern
- [ ] Umbuchung: Voll-Vorauswahl und Mengen-Obergrenze verhalten sich wie
      bisher (bestehende bzw. angepasste Tests belegen es)
- [ ] Direktverkauf-Storno: Pflicht-Kommentar und Buttons außerhalb des
      Scrollbereichs; ScrollArea-Mock aus `DirektverkaufHistorie.test.tsx`
      entfernt
- [ ] `make test-frontend` und `make lint` grün

---

## Phase 3: Receipt nativ scrollen, ScrollArea entfernen

**User Stories:** 6, 7

### Kontext

- `frontend/src/service/components/table/Receipt.tsx:19` — ScrollArea `max-h-96`
- Nutzer: `ZahlungDrawer.tsx:111` (Trinkgeld/Erhalten/Kommentar/Kassieren unterhalb), `AusgabeDrawer.tsx:91`, `BestellungDrawer.tsx:98`, `TischHistorie.tsx:458`
- `frontend/src/components/ui/scroll-area.tsx` — letzter Nutzer fällt in dieser Phase

### Was zu bauen ist

`Receipt` behält seine Schnittstelle (Positionen, Gesamtsumme) und
ersetzt intern die ScrollArea durch natives `overflow-y-auto` mit
dvh-basierter Maximalhöhe; die Maximalhöhe ist so bemessen, dass im
Zahlungs-Drawer Trinkgeld-/Erhalten-Feld, Kommentar und "Kassieren"
sichtbar bleiben. Damit sind Zahlungs- und Ausgabe-Drawer abgedeckt
(Bestellungs-Drawer und Tischhistorie profitieren mit). Anschließend
wird `scroll-area.tsx` gelöscht (kein Nutzer mehr im Repo).

### Akzeptanzkriterien

- [ ] Receipt scrollt nativ; Schnittstelle unverändert (bestehende Tests
      der Nutzer bleiben grün)
- [ ] `rg -l ScrollArea frontend/src` liefert keine Treffer mehr;
      `scroll-area.tsx` gelöscht
- [ ] `make test-frontend` und `make lint` grün

---

## Phase 4: Fehler-Referenz im Serverfehler-Toast

**User Stories:** 11, 12

### Kontext

- `backend/api/middleware/middleware.go:46-58` — Header wird vor jedem Handler gesetzt, steht damit auch auf 500-Antworten (inkl. RecoveryMiddleware-Pfad, Zeilen 85-105)
- `backend/api/middleware/middleware_test.go:51-69` — bestehende Header-Tests (Vorbild)
- `frontend/src/lib/Backend.ts:110-153` — `throwIfNotOk`, einziger Fehler-Konstruktionspfad
- `frontend/src/lib/errorMessages.ts:3-4,92-116` — `serverErrorMessage` und 500er-Zweig
- `frontend/src/hooks/use-action-submit.ts` — Toast-Pfad, bleibt unverändert

### Was zu bauen ist

`throwIfNotOk` liest `X-Correlation-ID` aus den Response-Headern und gibt
sie als optionales Feld `referenz` an `BackendError` weiter.
`getActionErrorMessage` hängt bei Serverfehlern (Status >= 500 bzw. Code
`internal_server_error`/`unknown`) die Referenz an die generische Meldung
an ("… Referenz: a1b2c3d4"); ohne Referenz bleibt die Meldung unverändert.
Backend-seitig belegt ein Test, dass eine 500-Antwort (z. B. der
Recovery-Pfad) den `X-Correlation-ID`-Header trägt; am Backend-Code
ändert sich nichts.

### Akzeptanzkriterien

- [ ] Test des Fehlermeldungs-Moduls: 500er-Meldung zeigt die Referenz an;
      ohne Referenz unverändert; Nicht-500er-Meldungen unverändert
- [ ] Test: `BackendError` trägt die Referenz aus dem Response-Header
- [ ] Backend-Test: 500-Antwort enthält den `X-Correlation-ID`-Header
- [ ] `make test-all` und `make lint` grün

---

## Phase 5: Druckauftrag-Backoff

**User Stories:** 13, 14, 15, 19, 20

### Kontext

- `database/migrations/README.md` — Forward-only-Regeln (additiv, `BEGIN;`/`COMMIT;`, Nummer `02_`)
- `database/migrations/01_initial.up.sql:284-304` — Tabelle und Kommentare
- `backend/sqlc/queries/relay.sql:5-22` — Poll-Abfrage und Fehlversuch-Update
- `backend/repository/druckauftrag_repo/repo.go:14,104-151` — `MaxDruckversuche`, Poll, Ergebnis-Transaktion; Tests `repo_test.go`
- `backend/api/druck/auftrag/http/handler_test.go` — Handler-Test-Vorbild
- `frontend/src/admin/settings/DruckstationConfigPage.tsx:248` — Hinweistext "nach drei Versuchen"
- `.github/workflows/ci.yml:398-470` — Upgrade-Path-Job wird mit dieser Migration erstmals scharf
- `windows/relay/main.go` — bleibt unverändert (Poll-Takt, Gruppierung, ID-Reihenfolge)

### Was zu bauen ist

Migration `02_druckauftrag_backoff.up.sql` fügt
`naechster_versuch_ab TIMESTAMPTZ NULL` hinzu (NULL = sofort fällig) und
aktualisiert die Kommentare (Endzustand nach 6 Fehlversuchen). Im
Repository: reine Backoff-Funktion (Fehlversuch 1 bis 5 zu 5s, 15s, 30s,
60s, 180s), `MaxDruckversuche` auf 6. Die Poll-Abfrage liefert nur
fällige Aufträge (`naechster_versuch_ab IS NULL OR <= NOW()`), weiterhin
in ID-Reihenfolge (Nachdrucke kommen so in ursprünglicher Reihenfolge,
Gruppierung je Drucker macht das Relay). Der Fehlversuch-Update setzt die
nächste Fälligkeit DB-seitig (`NOW()` plus übergebene Wartezeit); beim
6. Fehlversuch kippt der Auftrag wie bisher auf fehlgeschlagen.
Einzel-Retry setzt zusätzlich `naechster_versuch_ab` zurück. sqlc-Code
regenerieren (`make sqlc`); den Hinweistext auf der Druckstationen-Seite
an die neue Semantik anpassen (rund 5 Minuten statt drei Versuche).
Kassenbelege laufen über dieselbe Tabelle und profitieren automatisch.

### Akzeptanzkriterien

- [ ] Unit-Tests der Backoff-Funktion (Versuchsnummer zu Wartezeit,
      Grenzen)
- [ ] Repository-Integrationstests: nicht fällige Aufträge werden nicht
      ausgeliefert; Fehlversuch setzt die nächste Fälligkeit; 6.
      Fehlversuch führt in den Endzustand; Einzel-Retry setzt Versuche
      und Fälligkeit zurück
- [ ] Migration additiv und transaktional; Frischinstallation
      (`test-integration`) und Upgrade-Path-CI-Job grün
- [ ] Hinweistext auf der Druckstationen-Seite entspricht der neuen
      Semantik
- [ ] `make verify` grün

---

## Phase 6: Sammel-Retry für fehlgeschlagene Druckaufträge

**User Stories:** 16, 17

### Kontext

- `backend/sqlc/queries/relay.sql:30-33` — Einzel-Retry-Query (Status-Guard-Vorbild)
- `backend/api/druck/auftrag/application/command.go` und `.../http/handler.go` — Command/Handler-Vorbild inkl. Tests
- `backend/api/admin.go:121-127` — Endpoint-Registrierung
- `frontend/src/admin/settings/DruckstationBackend.ts:97-103`, `hooks.ts:42-56`, `DruckstationConfigPage.tsx:200-250` — bestehende Verwaltung der fehlgeschlagenen Aufträge
- `frontend/src/admin/users/UserItem.tsx:4` — AlertDialog-Bestätigungsmuster

### Was zu bauen ist

Neue Query "alle fehlgeschlagenen Aufträge erneut einreihen"
(Status-Guard `fehlgeschlagen`, setzt Versuche, letzten Fehler und
Fälligkeit zurück), dazu Command und Endpoint
`admin/druckauftraege-erneut-versuchen`. Auf der Druckstationen-Seite
ein Button "Alle erneut einreihen" oberhalb der Liste (nur sichtbar,
wenn fehlgeschlagene Aufträge existieren) mit Bestätigungsdialog; der
Einzel-Retry und das Verwerfen bleiben unverändert bestehen.

### Akzeptanzkriterien

- [ ] Repository-Test: Sammel-Retry reiht nur fehlgeschlagene Aufträge
      wieder ein; gedruckte/verworfene/offene bleiben unberührt
- [ ] Handler-Test für den neuen Endpoint (Vorbild Einzel-Retry)
- [ ] Frontend-Test: Button erscheint nur bei fehlgeschlagenen Aufträgen,
      löst nach Bestätigung den Sammel-Retry aus, Liste aktualisiert sich
- [ ] `make verify` grün

---

## Phase 7: Playwright-Abnahme am Mobil-Viewport

**User Stories:** 1 bis 8 (Verifikation)

### Kontext

- PRD Testing Decisions: Layout-/CSS-Verhalten ist in jsdom nicht sinnvoll testbar, Abnahme per Playwright
- Gebündeltes Chromium headless wie bei den bisherigen Screenshot-Abnahmen (`npx playwright screenshot --browser chromium` bzw. MCP mit `--browser chromium --headless`)
- `backend/seed/` — Seed-Engine für Testdaten

### Was zu bauen ist

Lokale App mit Testdaten, die die Praxistest-Situation nachstellen:
eine Bestellung mit 10 und mehr Positionen und überlangen Produkt- und
Variantennamen (Seed-Szenario erweitern oder Daten über die UI anlegen).
Am iPhone-Viewport alle fünf umgestellten Stellen abnehmen
(Stornierung, Umbuchung, Direktverkauf-Storno, Zahlung, Ausgabe):
Liste scrollt, Kommentarfeld und Buttons sichtbar, truncate greift,
Plus/Minus erreichbar, nichts läuft über. Die Bildschirmtastatur
(User Story 8) wird über eine reduzierte Viewport-Höhe angenähert;
der echte Geräte-Check bleibt Teil der manuellen QA.

### Akzeptanzkriterien

- [ ] Screenshots aller fünf Stellen am iPhone-Viewport mit vielen
      Positionen und überlangen Namen liegen vor und sind abgenommen
- [ ] Szenario reduzierte Viewport-Höhe (Tastatur-Näherung): Kommentarfeld
      und Primär-Button bleiben sichtbar
- [ ] Kein Befund offen bzw. Befunde sind behoben und erneut abgenommen
