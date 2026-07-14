# Plan: Cleanup-Audit — 30 bestätigte Findings umsetzen

> Source PRD: n/a (Multi-Experten-Cleanup-Audit vom 2026-07-14, 17 Reviewer + adversariale Verifikation; 30 bestätigte Findings)

## Goal

Alle 30 bestätigten Findings des repo-weiten Cleanup-Audits umsetzen — verhaltenserhaltend
(Boundary-Fixes ändern nur Validierung/Feldnamen im Rahmen der Freeze-Disziplin), inklusive
des Strukturvorschlags (Entdopplung der puren Positions-Helfer nach `domain/kasse`).
Kein Feature-Scope; ausschließlich Korrektheit der Docs, tote-Code-Entfernung, Konsistenz
und DRY.

## Architectural decisions

Durable Entscheidungen, auf die sich mehrere Phasen beziehen:

- **API-Feldname**: Der DSFinV-K-Export-Request nutzt `kassensitzungNr` (wie alle anderen
  Kassensitzungs-Endpunkte), nicht mehr `kassensitzung`. Frontend, Backend und
  `scripts/ops-smoke.sh` ändern sich im selben Commit (Freeze-Disziplin: API-Formate dürfen
  sich ändern, solange alles zusammen ausgeliefert wird).
- **Domain-Helfer**: Die duplizierten puren Positions-Helfer werden als exportierte
  Funktionen `kasse.ValidatePositionRefs(available []Position, requested []PositionRef) bool`
  und `kasse.ResolvePositionen(available []Position, requested []PositionRef) ([]Position, int)`
  in einer neuen Datei `backend/domain/kasse/positionen.go` zusammengeführt. Sie operieren
  ausschließlich auf Domain-Typen (`Position`, `PositionRef`) — keine neuen Abhängigkeiten
  der Domain-Schicht.
- **Frontend-Namensvalidierung**: `NameSchema` in `frontend/src/admin/products/Produkt.ts`
  trimmt vor der Längenprüfung (`z.string().trim().min(3).max(100)`), identisch zur
  Backend-Quelle der Wahrheit (`backend/domain/produkt/product.go — NameSchema`).

## Inventory

Relevante bestehende Muster und Dateien (Auszug; Details pro Phase unter „Context"):

- `backend/api/kasse/{tischgeschaeft,direktverkauf}/application/command.go` — die
  duplizierten Helfer `validatePositionRefs` (byte-identisch) und
  `resolvePositions`/`resolvePositionen` (identisch bis auf Namen/Parameternamen/Kommentar).
- `backend/domain/kasse/kassensitzung_events.go` — kanonische Konstanten
  `GeldtransitRichtungEinlage`/`GeldtransitRichtungEntnahme` („die einzigen erlaubten Werte").
- `frontend/src/admin/tse/tseAmpel.ts — tseAmpel()` — laut Kommentar in
  `frontend/src/admin/AdminSidebar.tsx` die Single Source of Truth für den TSE-Fehlerzustand.
- `frontend/src/service/components/table/drawerUtils.ts — toBestellungData(), calculateTotalPrice()`
  — kanonisches Produkte×Mengen-Mapping; `BestellungDrawer.tsx` nutzt es bereits.
- `scripts/ops-smoke.sh` — Smoke-Test sendet `{"kassensitzung":0}` an
  `/api/admin/export/dsfinvk`; Backend liest den Body mit `DisallowUnknownFields`.
- Vorbild-Muster für Status-Change-Helfer: `applyVarianteStatusChange` (produkt) und
  `applyTischStatusChange` (tisch) parametrisieren nur `successMsg`.

## Resolved decisions

- **Refactor-Scope (dokumentierte Annahme, Rückfrage war nicht zustellbar):** Der
  Strukturrefactor hebt **nur die puren Helfer** (`validatePositionRefs`,
  `resolvePositions`/`resolvePositionen`) nach `domain/kasse` — exakt der bestätigte
  Finding-Vorschlag mit geringstem Risiko. Die ebenfalls duplizierten
  Kassensitzungs-/OCC-Helfer (`getOffeneKassensitzungOderFehler`, OCC-Write-Helfer) bleiben
  unangetastet (siehe Open questions).
- Die 2 im Audit widerlegten Findings (`betreiber/http/command_handler.go`-Schema,
  `FormFields.tsx`-Imports) sind **nicht** Teil des Plans.
- Datei-Löschungen (`mode-toggle.tsx`, `plan-ui-audit-politur.md`, `prd-ui-audit-politur.md`)
  sind durch die Plan-Freigabe gedeckt; vor dem Löschen wird per `grep -r '<dateiname>' .`
  geprüft, dass keine Referenzen zurückbleiben (Repo-Regel „No dead links").
- Alle Änderungen sind verhaltenserhaltend; einzige beobachtbare Ausnahme laut
  Freeze-Disziplin: der API-Feldname `kassensitzungNr` und das Frontend-Trimming
  (beide gewollt, beidseitig im selben Release).

## Open questions / Risks

- **Volle Entdopplung der Kasse-Application-Pakete** (`getOffeneKassensitzungOderFehler` +
  OCC-Write-Helfer über tischgeschaeft/direktverkauf/kassenfuehrung konsolidieren): bewusst
  nicht in diesem Plan. Falls gewünscht, als separates Vorhaben mit eigenem Plan — der
  Zuschnitt (gemeinsames Application-Helferpaket vs. Domain) ist eine eigene
  Architekturentscheidung.
- `frontend/src/admin/reporting/ReportingBackend.ts — exportDsfinvk()` sendet bei
  `kassensitzungNr === null` ein leeres Objekt `{}`; dieses Verhalten bleibt unverändert —
  nur der Key im Nicht-Null-Fall wird umbenannt.

---

## Phase 1: Boundary-Fixes (Frontend/Backend-Konsistenz)

### Context

- `frontend/src/admin/products/Produkt.ts — NameSchema` — trimmt nicht, Backend trimmt.
- `backend/domain/produkt/product.go — NameSchema` — Quelle der Wahrheit (`Trim().Min(3).Max(100)`).
- `backend/api/fiskal/export/http/handler.go — exportRequest` — einziger Endpunkt mit
  Feldname `kassensitzung` statt `kassensitzungNr`.
- `frontend/src/admin/reporting/ReportingBackend.ts — exportDsfinvk()` — mappt
  `kassensitzungNr` auf den Ausreißer-Key.
- `scripts/ops-smoke.sh` — Export-Schritt sendet `{"kassensitzung":0}`.

### What to build

Die zwei bestätigten Cross-Layer-Mismatches beheben: (1) Frontend-`NameSchema` trimmt vor
der Längenprüfung, sodass Produkt- und Variantennamen mit umgebendem Whitespace auf beiden
Seiten identisch bewertet werden. (2) Der DSFinV-K-Export-Request nutzt beidseitig
`kassensitzungNr` — Go-Struct-Feld inkl. `json`-Tag und Verwendungsstelle im Handler,
Frontend-Key in `exportDsfinvk()`, und der Smoke-Test-Payload in `ops-smoke.sh`.

### Acceptance criteria

- [x] `NameSchema` in `Produkt.ts` enthält `.trim()` vor `min`/`max`; bestehende
      Fehlermeldungs-Strings bleiben erhalten.
- [x] `backend/api/fiskal/export/http/handler.go` verwendet Struct-Feld
      `KassensitzungNr int` mit Tag `json:"kassensitzungNr"`; keine Vorkommen von
      `json:"kassensitzung"` mehr im Repo (`grep -rn 'json:"kassensitzung"' backend/`).
- [x] `ReportingBackend.ts — exportDsfinvk()` sendet `{ kassensitzungNr }` im
      Nicht-Null-Fall und weiterhin `{}` bei `null`.
- [x] `scripts/ops-smoke.sh` sendet `{"kassensitzungNr":0}`.
- [x] `make check` läuft grün.

---

## Phase 2: Backend-Cleanup (toter Code, Konstanten, Konsistenz)

### Context

- `backend/api/kasse/tischgeschaeft/application/errors.go — ErrTischAlreadyExists, fromRepositoryError()` —
  unerreichbarer Copy-Paste-Code aus stammdaten/tisch.
- `backend/api/kasse/tischgeschaeft/http/command_handler.go — toBestellPositionInput(), toPositionRef()` —
  einmal benutzte Einzelitem-Konverter; Vorbild: Inline-Loops in
  `backend/api/kasse/direktverkauf/http` (`toVerkaufPositionInputs`, `toPositionRefs`).
- `backend/api/druck/bondruck/application/escpos/constants.go — QRCodeModuleSize6` — nirgends referenziert.
- `backend/repository/reporting_repo/repo.go — toStornierungPositionen()` — liefert nie nil;
  der nachgelagerte `if positionen == nil`-Guard ist unerreichbar.
- `backend/domain/kasse/tse_processdata.go — BuildGeldtransitProcessData()` und
  `backend/domain/kasse/tagesabschluss_summen.go` — Literale `"einlage"`/`"entnahme"` statt
  `GeldtransitRichtungEinlage`/`GeldtransitRichtungEntnahme` (definiert in
  `backend/domain/kasse/kassensitzung_events.go`).
- `backend/domain/event/event.go — Event` (Felder `Type`, `Subject`) — CloudEvents-Boilerplate-Beispiele
  und doppeltes „E.g.".
- `backend/api/stammdaten/user/application/command.go — applyUserStatusChange()` — vier
  positionsbasierte Log-Strings; Vorbilder `applyVarianteStatusChange`/`applyTischStatusChange`.
- `backend/app/app.go — NewApp()` — nie belegter `error`-Rückgabewert; Aufrufer:
  `backend/main.go` und `backend/app/app_test.go`.
- `backend/api/health/health_integration_test.go — TestHealthCheck_WithDatabase` — nur `t.Skip`.
- `backend/api/kasse/tischgeschaeft/application/command_test.go — newTestCommand(), newTestCommandWithEventMock()` —
  identisch bis auf den Event-Mock-Parameter.

### What to build

Alle bestätigten Backend-Findings außerhalb des Strukturrefactors: toten Code löschen
(unerreichbarer Error-Branch samt Sentinel, ungenutzte ESC/POS-Konstante, unerreichbarer
nil-Guard, leerer Skip-Test), Magic Strings durch die vorhandenen Konstanten ersetzen,
Kommentar-Beispiele in `event.go` auf jottis reale Formate korrigieren (z. B.
`bestellung-aufgenommen:v1`, `kassensitzung-1/tisch-42`), die Einzelitem-Konverter in ihre
Slice-Loops inlinen (Stil der direktverkauf-Geschwister), `applyUserStatusChange` auf das
Geschwister-Muster angleichen (nur `successMsg` parametrisiert, feste generische
Failure-Messages), `NewApp` ohne `error`-Rückgabe, und die Test-Helfer-Dopplung per
Delegation auflösen (`newTestCommand` ruft `newTestCommandWithEventMock` mit Inline-Mock).

### Acceptance criteria

- [x] `errors.go` (tischgeschaeft/application): `ErrTischAlreadyExists` und der
      `db.ErrAlreadyExists`-Branch sind entfernt; `grep -rn ErrTischAlreadyExists backend/api/kasse/`
      liefert nichts.
- [x] `QRCodeModuleSize6` existiert nicht mehr im Repo.
- [x] Der `if positionen == nil`-Guard in `reporting_repo/repo.go` ist entfernt.
- [x] `tse_processdata.go` und `tagesabschluss_summen.go` referenzieren ausschließlich
      `GeldtransitRichtungEinlage`/`GeldtransitRichtungEntnahme`; keine Literale
      `"einlage"`/`"entnahme"` mehr in Vergleichen/Switches dieser Dateien.
- [x] `event.go`-Kommentare zeigen jotti-Formate, kein `com.library.book.borrowed`, kein
      doppeltes „E.g." mehr.
- [x] `command_handler.go` (tischgeschaeft/http): `toBestellPositionInput` und
      `toPositionRef` sind inlined/entfernt; nur die Slice-Varianten bleiben.
- [x] `applyUserStatusChange` hat die Signatur-Form der Geschwister (nur `successMsg` als
      Message-Parameter); alle Aufrufer angepasst.
- [x] `NewApp` gibt nur `*App` zurück; `backend/main.go` und `app_test.go` ohne toten
      err-Check.
- [x] `TestHealthCheck_WithDatabase` ist gelöscht.
- [x] `newTestCommand` delegiert an `newTestCommandWithEventMock`.
- [x] `make check` läuft grün (inkl. `event_json_contract_test.go` — Event-JSON unverändert).

---

## Phase 3: Frontend-Cleanup (DRY, toter Code, Konsistenz)

### Context

- `frontend/src/admin/reporting/AdminDashboardPage.tsx` — leitet den TSE-Fehlerzustand
  inline her statt über `tseAmpel()`.
- `frontend/src/admin/tse/tseAmpel.ts — tseAmpel()` — Single Source of Truth; muss ggf. die
  Einzel-Flags (nichtKonfiguriert, rueckstand, signaturFehlgeschlagen) mit zurückgeben,
  damit `tseText` weiter gebaut werden kann.
- `frontend/src/service/components/direktverkauf/Direktverkauf.tsx — selectItems()` —
  dupliziert das Mapping aus `toBestellungData()` (drawerUtils, dort bereits importiert);
  Vorbild: `frontend/src/service/components/table/BestellungDrawer.tsx`.
- `frontend/src/service/components/table/Zahlung.tsx — loading-Prop, PositionItemSkeleton` —
  unerreichbar, da `TablePage.tsx` Zahlung nur unter `!stateLoading` rendert.
- `frontend/src/components/mode-toggle.tsx — ModeToggle` — nirgends importiert; echter
  Toggle lebt inline in `frontend/src/components/common/UserDropdown.tsx`.
- `frontend/src/routes.ts — AuthRedirect()` — evaluiert das seiteneffektbehaftete
  `AuthSingleton.isAuthenticated` bis zu dreimal; Vorbild: Guard-Clauses in
  `AdminGuard`/`ServiceGuard` derselben Datei.
- `frontend/src/service/components/table/HistorieStornierungDrawer.tsx`,
  `HistorieUmbuchungDrawer.tsx`, `DirektverkaufStornoDrawer.tsx` — `open={true}` statt
  Shorthand `open` (Vorbild: Detail-Drawer in `TischHistorie.tsx`, `DirektverkaufHistorie.tsx`).
- `frontend/src/components/common/EuroInput.tsx` — Doc-Kommentar mit Füllwort „seamlessly".

### What to build

Alle bestätigten Frontend-Findings: `AdminDashboardPage` bezieht den TSE-Fehlerzustand aus
`tseAmpel()` (Funktion bei Bedarf um die Einzel-Flags erweitern, damit die Regel genau
einmal definiert ist); `Direktverkauf.tsx` ersetzt `selectItems`/`SelectedItem` durch
`toBestellungData()` + `calculateTotalPrice()` (Gesamtsumme und Positionsanzahl aus dessen
Rückgabe); tote Pfade löschen (`loading`-Prop samt Skeleton-Branch und
`PositionItemSkeleton` in `Zahlung.tsx` inkl. Callsite in `TablePage.tsx`; Datei
`mode-toggle.tsx`); `AuthRedirect` mit früher Guard-Clause und genau einer
`isAuthenticated`-Auswertung; `open={true}` → `open`-Shorthand in den drei Drawern;
„seamlessly" aus dem `EuroInput`-Kommentar streichen.

### Acceptance criteria

- [x] `AdminDashboardPage.tsx` berechnet keinen eigenen TSE-Fehlerzustand mehr; die
      Schwellen-/Fehlerregel steht nur noch in `tseAmpel.ts`.
- [x] `Direktverkauf.tsx` enthält kein lokales `selectItems`/`SelectedItem` mehr und nutzt
      `toBestellungData` + `calculateTotalPrice`; Anzeige (Summe, Anzahl, Beleg-Items) und
      gesendete Positionen sind unverändert.
- [x] `Zahlung.tsx` hat keine `loading`-Prop und kein `PositionItemSkeleton`;
      `TablePage.tsx` übergibt kein `loading` mehr.
- [x] `frontend/src/components/mode-toggle.tsx` ist gelöscht;
      `grep -rn 'mode-toggle\|ModeToggle' frontend/src/` liefert nichts.
- [x] `AuthRedirect` beginnt mit `if (!AuthSingleton.isAuthenticated) return` und wertet
      den Getter genau einmal aus.
- [x] Die drei Storno-/Umbuchungs-Drawer nutzen das `open`-Shorthand.
- [x] `EuroInput.tsx`-Kommentar ohne „seamlessly".
- [x] `make check` läuft grün.

---

## Phase 4: Docs & Satelliten (Korrektheit, tote Links, Konventionen)

### Context

- `docs/language.md` — DSFinV-K-Glossareintrag: falsche Version („2.5") und falscher Status
  („Geplant"); Code pinnt `Version = "2.4"` (`backend/api/fiskal/dsfinvk/dsfinvk.go`),
  Zeile „DSFinV-K-Export sind umgesetzt" steht bereits in derselben Datei.
- `docs/handbuch.md` — vier Findings: toter Link `[betrieb/](betrieb/)` im Kopf; toter
  PRD-Link `prds/prd-tse-admin-vereinfachung.md`; stale „Die TSE-Integration wird
  phasenweise implementiert"; verwaiste Referenz „Compliance-Phase 1" (echte Regel steht in
  §3.11 des handbuch.md).
- `docs/plans/plan-ui-audit-politur.md` — 50/50 Checkboxen abgehakt, gemergt; ADR 04
  (`docs/adrs/04_warn-bestaetigung.md`) dokumentiert Plan und PRD bereits als „nach Merge
  gelöscht".
- `docs/prds/prd-ui-audit-politur.md` — dito.
- `reverse-proxy/nginx.rocks.conf` — zwei Kommentare verweisen auf ein nicht mehr
  existierendes `nginx.conf`; der CSP-Zwilling ist `reverse-proxy/caddyfile.go` (dessen
  Gegenkommentar bereits korrekt `nginx.rocks.conf` nennt).
- `scripts/prod-init.sh` — einzelne deutsche Zeile „Stack neustarten" im englischen
  Operator-Block; Vorbild: voll-englischer Block in `scripts/rocks-init.sh`.

### What to build

Die Dokumentations- und Satelliten-Findings: Glossareintrag auf „Version 2.4" und
„umgesetzt (→ F-04)" korrigieren; in `handbuch.md` den Kopf-Link auf
`[Leitfaden](leitfaden/was-ist-jotti.md)` zeigen (Konvention der Geschwister-Docs), den
PRD-Verweis ersatzlos streichen, die TSE-Aussage auf den Ist-Zustand umformulieren
(„unterliegt der TSE-Pflicht nach § 146a AO (umgesetzt)"), die
„Compliance-Phase 1"-Parenthese streichen bzw. durch die tatsächliche Regel aus §3.11
ersetzen; die zwei abgeschlossenen Audit-Artefakte (`plan-ui-audit-politur.md`,
`prd-ui-audit-politur.md`) löschen und Referenzen prüfen; die `nginx.conf`-Verweise in
`nginx.rocks.conf` auf das Caddy-Setup (`caddyfile.go`, `docker-compose.prod.yml`)
korrigieren; „Stack neustarten" → „Restart the stack".

### Acceptance criteria

- [x] `docs/language.md`: DSFinV-K-Eintrag nennt „Version 2.4" und einen
      Umgesetzt-Status; kein „Geplant (→ F-04)" mehr.
- [x] `docs/handbuch.md`: kein Link auf `betrieb/` oder
      `prds/prd-tse-admin-vereinfachung.md`; keine Formulierung „wird phasenweise
      implementiert"; kein Vorkommen von „Compliance-Phase 1" mehr im Repo.
- [x] `docs/plans/plan-ui-audit-politur.md` und `docs/prds/prd-ui-audit-politur.md` sind
      gelöscht; `grep -rn 'ui-audit-politur' . --include='*.md'` trifft nur noch die
      Git-Historie-Verweise in ADR 04 (und ggf. diesen Plan).
- [x] `nginx.rocks.conf` verweist nicht mehr auf ein unqualifiziertes `nginx.conf`;
      CSP-Pflegehinweis nennt nur noch `caddyfile.go`.
- [x] `scripts/prod-init.sh`: Operator-Block durchgängig englisch.
- [x] Link-Check: alle in dieser Phase angefassten relativen Links zeigen auf existierende
      Dateien.

---

## Phase 5: Struktur — Positions-Helfer nach domain/kasse

### Context

- `backend/api/kasse/tischgeschaeft/application/command.go — validatePositionRefs(), resolvePositions()`
- `backend/api/kasse/direktverkauf/application/command.go — validatePositionRefs(), resolvePositionen()`
- `backend/domain/kasse/` — Zielpaket; enthält bereits `Position` und `PositionRef`.

### What to build

Die vier duplizierten puren Funktionen zu zwei exportierten Domain-Funktionen
zusammenführen (siehe Architectural decisions): neue Datei
`backend/domain/kasse/positionen.go` mit `ValidatePositionRefs` und `ResolvePositionen`
(Doc-Kommentare aus beiden Quellen zusammengeführt: Duplikat-/Mengen-Prüfung bzw.
„fat Positions" für selbst-tragende Events). Beide Application-Pakete rufen die
Domain-Funktionen auf; die lokalen Kopien werden gelöscht. Bestehende Unit-Tests der
Application-Pakete bleiben als Verhaltens-Guard bestehen; falls die Helfer dort direkt
getestet werden, wandern diese Tests nach `backend/domain/kasse/positionen_test.go`.

### Acceptance criteria

- [x] `grep -rn 'func validatePositionRefs\|func resolvePosition' backend/api/` liefert
      nichts; die einzige Implementierung liegt in `backend/domain/kasse/positionen.go`.
- [x] Beide Application-Pakete nutzen `kasse.ValidatePositionRefs`/`kasse.ResolvePositionen`;
      Aufruf-Semantik (Reihenfolge, Rückgaben, Fehlerpfade) unverändert.
- [x] Direkt auf die Helfer zielende Tests liegen im Domain-Paket; alle bestehenden
      Storno-/Bestell-Tests laufen unverändert grün.
- [x] `make check` läuft grün; abschließend `make verify` für den ganzen Plan.
