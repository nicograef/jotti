# Plan: Vereinfachter Kassenabschluss

> Source PRD: docs/prds/prd-kasse-abschluss-vereinfachung.md

## Goal

Kassensturz und Tagesabschluss werden aus Sicht des Kassenwarts zu einem
geführten Schritt "Kasse abschließen" zusammengelegt: Ist-Bestand eingeben,
Bestätigungsdialog mit Soll/Ist/Differenz und Z-Bon-Vorschau, dann ein Vorgang,
der Differenzbuchung, Z-Bon und Schließen ausführt. Das
`tagesabschluss-erstellt:v1`-Event trägt die echten Tagessummen statt Nullen.
Begleitend werden Hilfetexte, fingerfreundliche Auswahlflächen beim Geldtransit
und eine Komponenten-Umbenennung ergänzt. Fiskalische Pflichten und Invarianten
bleiben unverändert.

## Architectural decisions

Durchgängige Entscheidungen über alle Phasen:

- Routes: Neuer Endpunkt `POST admin/kasse-abschliessen` ersetzt
  `admin/kassensturz-durchfuehren` und `admin/tagesabschluss-erstellen`. Die
  übrigen Kassen-Endpunkte (`kassensitzung-eroeffnen`, `geldtransit-buchen`,
  `get-kassenbestand`, `get-offene-kassensitzung`, `get-live-reporting`)
  bleiben unverändert.
- Events: Keine neuen Event-Typen. Feste Schreibreihenfolge
  `kassensturz-durchgefuehrt:v1`, bei Differenz ungleich Null zusätzlich
  `differenz-soll-ist-gebucht:v1`, abschließend `tagesabschluss-erstellt:v1`.
  Das `tagesabschluss-erstellt:v1`-Event behält seine Felder, wird aber mit
  echten Tagessummen befüllt.
- Datenmodell: Die Reporting-`Summary` erhält ein Feld `GeldtransitCents`
  (Frontend `geldtransitCents`). Es ist die einzige Quelle für die
  Geldtransit-Summe und versorgt sowohl den Abschluss-Command als auch den
  Bestätigungsdialog (über das Live-Reporting).
- Command: Ein einziger Anwendungs-Command
  `KasseAbschliessen(ctx, userID, userName, istBestandCents)`. Der kasse-`Command`
  bekommt eine `reportingRepo`-Abhängigkeit, um die Tagessummen aus
  `GetReporting` zu lesen.
- Auth: Unverändert über die bestehende Admin-Middleware.
- Fehlercodes: `kassensturz_erforderlich` entfällt samt
  `ErrKassensturzErforderlich`. `tische_saldo_offen` und `kasse_nicht_geoeffnet`
  bleiben.
- TSE: Signierung der signierungspflichtigen Events (Differenzbuchung,
  Tagesabschluss) inklusive Nachsignierungs-Fallback bleibt unverändert.
- Konsistenz: Kein Umbau auf eine einzige umschließende DB-Transaktion. Das
  bestehende sequentielle Schreibmuster mit optimistischer
  Nebenläufigkeitskontrolle bleibt; das Verhalten bei Teilfehlern wird
  dokumentiert.

## Inventory

Backend:

- `backend/api/kasse/application/command.go:34-39` Command-Struct mit
  KassenjournalRepo, KassensitzungenRepo, SettingsRepo, TSESignierer.
- `backend/api/kasse/application/command.go:200-251`
  `KassensturzDurchfuehren` (Soll/Ist, Zwei-Event-Muster).
- `backend/api/kasse/application/command.go:257-329`
  `TagesabschlussErstellen` mit Kassensturz-Invariante, Tisch-Saldo-Sperre und
  den hartkodierten Nullen `0, 0, 0, 0` (Zeile 308).
- `backend/api/kasse/application/errors.go:23-27`
  `ErrKassensturzErforderlich`, `ErrTischeSaldoOffen`.
- `backend/api/kasse/application/tse_signing.go:29-32`
  `signTagesabschlussErstelltEvent`.
- `backend/api/kasse/http/command_handler.go:14-19` command-Interface;
  `:122-178` Kassensturz- und Tagesabschluss-Handler samt Fehler-Mapping.
- `backend/api/admin.go:105-124` Command-Verdrahtung und Routen-Registrierung;
  `:94-97` `reportingRepo := reporting_repo.NewRepository(db)` ist bereits vorhanden.
- `backend/domain/kasse/kassensitzung_events.go:80-102` Event-Daten und Schema
  `TagesabschlussErstelltV1Data`; `:189-212` Erstellungsfunktion.
- `backend/domain/reporting/reporting.go:50-59` `Summary`-Struct.
- `backend/repository/reporting_repo/repo.go:38-77` `GetReporting`;
  `:202-213` `toSummary`.
- `backend/sqlc/queries/reporting.sql:10-38` `GetReportingStats` (Summe-Spalten
  und `WHERE type IN (...)`, ohne `geldtransit-gebucht:v1`).
- `backend/sqlc/queries/kassensitzungen.sql:17-35` vorhandene
  `kj_extract_geldtransit_cents`-Extraktion (vom Soll-Kassenbestand genutzt).
- `backend/seed/engine.go:455-482` `schliesseTagAb` befüllt den Tagesabschluss
  bereits mit echten Summen (`UmsatzGesamtCents()`, `StornierungenCents`,
  `AuszahlungenCents`, `GeldtransitCents`); kein Eingriff nötig.
- `backend/api/kasse/application/command_test.go` bestehende Command-Tests für
  Kassensturz und Tagesabschluss, plus Mock-Repos in
  `backend/repository/kassenjournal_repo/mock.go` und
  `backend/repository/kassensitzungen_repo/mock.go`.

Frontend:

- `frontend/src/admin/kasse/KassensitzungPage.tsx:29-87` Seite und Info-Karte
  mit Soll-Anzeige (`:64-71`); `:186-308` `KassenbewegungSection` (Geldtransit,
  Radio-Buttons `:243-259`); `:310-379` `KassensturzSection`; `:381-414`
  `TagesabschlussSection` (destruktiver Button ohne Dialog).
- `frontend/src/admin/kasse/KasseBackend.ts:35-37,75-82` Schema und Methoden
  `kassensturzDurchfuehren`, `tagesabschlussErstellen`.
- `frontend/src/admin/kasse/hooks.ts:22-29` `useKassenbestand`.
- `frontend/src/admin/reporting/hooks.ts:29-36` `useLiveReporting`.
- `frontend/src/admin/reporting/types.ts:3-12` `SummarySchema`; `:71-85`
  `LiveReportingDataSchema`.
- `frontend/src/components/ui/alert-dialog.tsx` Bestätigungsdialog-Primitive;
  Prior Art `frontend/src/admin/tables/TischItem.tsx`,
  `frontend/src/admin/users/UserItem.tsx`.
- `frontend/src/admin/settings/EinstellungenPage.tsx:39,221` Stil der
  Inline-Hilfetexte (`<p className="text-muted-foreground text-sm mb-4">`).

## Resolved decisions

- Geldtransit-Summe wird in die Reporting-`Summary` aufgenommen (eine Quelle für
  Command und Dialog), nicht über getrennte Queries.
- Der Soll-Kassenbestand wird von der Seite entfernt und erst im
  Bestätigungsdialog gezeigt (Soll, Ist, Differenz), um eine unverzerrte Zählung
  zu fördern.
- Breaking Changes an API und Events sind in der Pre-Release-Phase erlaubt; alte
  Endpunkte und der Fehlercode `kassensturz_erforderlich` werden ersatzlos entfernt.

## Open questions / Risks

- `GeldtransitCents` in der gemeinsamen `Summary` erscheint auch in historischen
  Reports (`GetReporting`-Antwort). Das ist beabsichtigt und unkritisch, sollte
  aber beim Frontend-Reporting kurz gegengeprüft werden, damit kein Zod-Parsing
  bricht.
- Teilfehler nach dem ersten Event lassen die Kassensitzung offen; der Abschluss
  ist wiederholbar. Dieses Verhalten wird im Command dokumentiert, ein
  transaktionaler Umbau ist nicht Teil dieses Plans.

---

## Phase 1: Geldtransit in der Reporting-Summary

**User stories**: 4, 15, 16 (Datengrundlage)

### Context

- `backend/sqlc/queries/reporting.sql:10-38` `GetReportingStats`: Summe-Spalten
  und die `WHERE type IN (...)`-Liste, die `geldtransit-gebucht:v1` noch nicht
  enthält.
- `backend/sqlc/queries/kassensitzungen.sql:17-35` zeigt die vorhandene
  `kj_extract_geldtransit_cents`-Extraktion als Vorbild.
- `backend/domain/reporting/reporting.go:50-59` `Summary`-Struct.
- `backend/repository/reporting_repo/repo.go:202-213` `toSummary`-Mapping.
- `frontend/src/admin/reporting/types.ts:3-12` `SummarySchema` (wird von Report
  und Live-Reporting genutzt).

### What to build

Die Geldtransit-Tagessumme wird durchgängig als vierte Tagessumme verfügbar
gemacht. In `GetReportingStats` kommt eine Spalte `gesamt_geldtransit_cents` auf
Basis von `kj_extract_geldtransit_cents` hinzu, und `geldtransit-gebucht:v1`
wird in die `WHERE type IN (...)`-Liste aufgenommen. Nach sqlc-Generierung wird
das Feld durch `toSummary` in die Domain-`Summary` gemappt und über die
bestehenden Reporting-DTOs bis ins Frontend-`SummarySchema` durchgereicht. Damit
liefern sowohl `GetReporting` als auch das Live-Reporting die Geldtransit-Summe.

### Acceptance criteria

- [x] `GetReportingStats` liefert `gesamt_geldtransit_cents` und schließt
      `geldtransit-gebucht:v1` ein; sqlc-Code ist neu generiert (`make sqlc`).
- [x] `reporting.Summary` enthält `GeldtransitCents`, `toSummary` befüllt es.
- [x] Die Reporting- und Live-Reporting-Antworten enthalten `geldtransitCents`;
      `SummarySchema` im Frontend kennt das Feld und parst ohne Fehler.
- [x] Backend- und Frontend-Build/Tests laufen grün.

---

## Phase 2: Backend KasseAbschliessen Command und Endpunkt

**User stories**: 1, 2, 3, 5, 7, 9, 15, 16, 17

### Context

- `backend/api/kasse/application/command.go:34-39,200-329` die beiden
  zusammenzulegenden Commands und das Command-Struct.
- `backend/api/kasse/application/tse_signing.go:24-32` Signierung von
  Differenzbuchung und Tagesabschluss.
- `backend/api/kasse/application/errors.go:23-27` zu entfernender
  `ErrKassensturzErforderlich`.
- `backend/api/kasse/http/command_handler.go:14-19,122-178` Handler und
  Fehler-Mapping.
- `backend/api/admin.go:105-124` Verdrahtung; `reportingRepo` liegt unter
  `:94-97` bereits vor.
- `backend/api/reporting/application/query.go:16-20` zeigt die vorhandene
  `reportingRepo`-Interface-Form (`GetReporting`).
- `backend/api/kasse/application/command_test.go` Prior Art und Mock-Repos.

### What to build

Ein neuer Command `KasseAbschliessen(ctx, userID, userName, istBestandCents)`
ersetzt `KassensturzDurchfuehren` und `TagesabschlussErstellen`. Er ermittelt die
offene Kassensitzung, liest den Soll-Bestand, berechnet die Differenz, prüft die
Tisch-Saldo-Sperre und schreibt in fester Reihenfolge
`kassensturz-durchgefuehrt:v1`, bei Differenz ungleich Null
`differenz-soll-ist-gebucht:v1`, und `tagesabschluss-erstellt:v1`. Die
Tagessummen für den Z-Bon kommen aus `GetReporting` (jetzt inklusive
Geldtransit) statt der bisherigen Nullen; der kasse-`Command` erhält dafür eine
`reportingRepo`-Abhängigkeit. Die Kassensturz-Invariante entfällt, die
Tisch-Saldo-Sperre bleibt. Der neue Endpunkt `admin/kasse-abschliessen` ersetzt
die beiden alten; `kassensturz_erforderlich` und `ErrKassensturzErforderlich`
werden entfernt. TSE-Signierung und Nachsignierungs-Fallback bleiben. Das
Teilfehler-Verhalten wird im Command-Kommentar dokumentiert. Die alten
Command-Tests werden durch neue für `KasseAbschliessen` ersetzt.

### Acceptance criteria

- [x] `KasseAbschliessen` schreibt bei Differenz Null genau ein Kassensturz- und
      ein Tagesabschluss-Event, keine Differenzbuchung; die Kassensitzung erhält
      Status abgeschlossen.
- [x] Bei Differenz ungleich Null wird zusätzlich genau eine Differenzbuchung
      geschrieben.
- [x] Das `tagesabschluss-erstellt:v1`-Event enthält die echten Tagessummen
      (Umsatz, Stornierungen, Auszahlungen, Geldtransit) statt Nullen.
- [x] Abschluss wird mit `tische_saldo_offen` abgelehnt, wenn ein Tisch einen
      Saldo ungleich Null hat.
- [x] Abschluss wird mit `kasse_nicht_geoeffnet` abgelehnt, wenn keine Sitzung
      offen ist.
- [x] `admin/kasse-abschliessen` existiert; die beiden alten Endpunkte, der
      Fehlercode `kassensturz_erforderlich` und `ErrKassensturzErforderlich` sind
      entfernt.
- [x] Command-Tests decken die fünf Fälle ab; Seed-Generator bleibt unverändert
      und erzeugt dieselbe Semantik wie die Produktion.

---

## Phase 3: Frontend Abschluss-Sektion mit Bestätigungsdialog

**User stories**: 1, 2, 3, 4, 5, 6, 8, 10

### Context

- `frontend/src/admin/kasse/KassensitzungPage.tsx:29-87` Seite und Soll-Anzeige
  (`:64-71`); `:310-379` `KassensturzSection`; `:381-414` `TagesabschlussSection`.
- `frontend/src/admin/kasse/KasseBackend.ts:35-37,75-82` zu ersetzende Methoden.
- `frontend/src/admin/kasse/hooks.ts:22-29` `useKassenbestand` (Soll für den
  Dialog).
- `frontend/src/admin/reporting/hooks.ts:29-36` `useLiveReporting` (Tagessummen
  für die Z-Bon-Vorschau).
- `frontend/src/components/ui/alert-dialog.tsx` und Prior Art
  `frontend/src/admin/tables/TischItem.tsx`.
- `frontend/src/admin/settings/EinstellungenPage.tsx:39` Hilfetext-Stil.

### What to build

`KassensturzSection` und `TagesabschlussSection` werden zu einer Sektion "Kasse
abschließen" zusammengeführt: ein Ist-Bestand-Feld und ein Button. Der Klick
öffnet einen `alert-dialog`, der Soll, Ist und Differenz gegenüberstellt, die
vier Tagessummen als Z-Bon-Vorschau zeigt und auf die Unwiderruflichkeit
hinweist. Bestätigung ruft eine neue Backend-Methode
`kasseAbschliessen(istBestandCents)`, Abbruch verändert nichts. Soll kommt aus
`useKassenbestand`, die Tagessummen aus `useLiveReporting`, das die Seite dafür
neu lädt. Die Soll-Zeile in der Info-Karte oben entfällt. Die Sektion bekommt
einen kurzen Abschluss-Hilfetext (Bargeld zählen, gezählten Betrag eintragen,
kleine Differenzen sind normal). Bei Erfolg erscheint eine klare Rückmeldung.
Leichtgewichtige Vitest-/Testing-Library-Tests prüfen die Ist-Eingabe und die
korrekte Anzeige von Soll, Ist und Differenz im Dialog.

### Acceptance criteria

- [x] Eine einzige "Kasse abschließen"-Sektion ersetzt die getrennten
      Kassensturz- und Tagesabschluss-Sektionen.
- [x] Der Bestätigungsdialog zeigt Soll, Ist, Differenz und die vier Tagessummen
      und weist auf die Unwiderruflichkeit hin; Abbruch löst keine Aktion aus.
- [x] Bestätigung ruft `kasseBackend.kasseAbschliessen(istBestandCents)` gegen
      `admin/kasse-abschliessen`; bei Erfolg gibt es eine Rückmeldung und die
      Seite aktualisiert sich.
- [x] Die Soll-Kassenbestand-Zeile in der oberen Info-Karte ist entfernt; Soll
      erscheint nur noch im Dialog.
- [x] `KasseBackend` hat `kasseAbschliessen` statt `kassensturzDurchfuehren` und
      `tagesabschlussErstellen`.
- [x] FE-Tests für Ist-Eingabe und Soll/Ist/Differenz-Anzeige laufen grün;
      Lint/Typecheck ohne Warnungen.

---

## Phase 4: Begleitende UX-Verbesserungen

**User stories**: 11, 12, 13, 14, 18

### Context

- `frontend/src/admin/kasse/KassensitzungPage.tsx:89-184` `EroeffnenSection`;
  `:186-308` `KassenbewegungSection` mit den Radio-Buttons (`:243-259`).
- `frontend/src/admin/settings/EinstellungenPage.tsx:39,221` Hilfetext-Stil.
- `frontend/src/components/ui` vorhandene Bausteine für größere Auswahlflächen
  (z.B. Button/Toggle-Varianten als Vorbild).

### What to build

Die Geldtransit-Komponente `KassenbewegungSection` wird zu `GeldtransitSection`
umbenannt (reine Umbenennung ohne Verhaltensänderung). Die Einlage/Entnahme-Wahl
wird von kleinen Radio-Buttons auf große, fingerfreundliche Auswahlflächen
umgestellt, Funktion und abgesendeter Wert bleiben gleich. Für die Blöcke
Eröffnen und Geldtransit kommen kurze Inline-Hilfetexte im Stil der
Einstellungen-Seite hinzu (gedämpfter, kleiner Absatz unter dem Titel), und die
Fachbegriffe erhalten eine Alltagsübersetzung als Untertitel, ohne die
verbindlichen Begriffe zu ersetzen.

### Acceptance criteria

- [x] Die Komponente heißt `GeldtransitSection`; alle Verweise sind aktualisiert,
      Verhalten unverändert.
- [x] Einlage und Entnahme sind als große Flächen auf dem Smartphone sicher
      antippbar; gesendeter Richtungswert bleibt `einlage`/`entnahme`.
- [x] Eröffnen und Geldtransit haben je einen kurzen Hilfetext und eine
      Alltagsübersetzung des Fachbegriffs; die Titel bleiben die verbindlichen
      Fachbegriffe.
- [x] Lint/Typecheck ohne Warnungen.
