# Plan: QA-Fixes Juni 2026

> Source PRD: [docs/prds/prd-qa-fixes-2026-06.md](../prds/prd-qa-fixes-2026-06.md)

## Goal

Die sieben Befunde aus der QA-Session vom 2026-06-12 beheben (Befund 5 ist out of scope,
eigenes PRD): fehlende Betreiber-Stammdaten als Client-Fehler statt 500 (Befund 1),
semantische Toast-Farben (Befund 2), stabile Sortierung von Produktvarianten (Befund 3) und
Druckstationen (Befund 4), sofortige Favoriten-Aktualisierung unter „Meine Tische"
(Befund 6), Sticky-Aktionsleiste auf der Tischseite (Befund 7) und 0-Cent-Beträge als
gültige Eingabe (Befund 8). Jeder Befund ist eine eigene, unabhängige Phase; die
Reihenfolge folgt der PRD-Empfehlung „Backend-Fixes zuerst".

## Architectural decisions

Durable Entscheidungen für alle Phasen:

- **Fehler-Semantik:** Erwartbare Konfigurationszustände sind Client-Fehler mit stabilem
  Fehlercode (hier: `betreiber_nicht_konfiguriert`); nur echte Datenbankfehler bleiben 500
  und werden als Error geloggt.
- **Sortierung im SQL, nicht im Frontend:** Deterministische Reihenfolgen liefert das
  Backend (Regel 9, AGENTS.md). Produktvarianten sortieren nach Varianten-ID
  (Anlage-Reihenfolge), Druckstationen nach Kategorie (fachlicher Unique-Schlüssel).
- **Pointer-Muster für 0-erlaubende Betragsfelder:** Request-DTO-Felder mit Regel „≥ 0"
  werden als `*int` modelliert und mit `z.Ptr(z.Int().GTE(0, …)).NotNil(…)` validiert
  (zog v0.22.0, `pointers.go:31,126`). Das ist die Referenz für künftige Betragsfelder,
  bei denen 0 gültig ist. Felder mit Regel „≥ 1" bleiben `int` + `Required()`.
- **Query-Key-Konstanten für Cross-Component-Invalidierung:** Query-Keys, die von anderen
  Komponenten invalidiert werden, sind exportierte Konstanten in
  `frontend/src/service/table/hooks.ts` (Muster: `AKTIVE_TISCHE_MIT_FAVORITEN_KEY`).
- **Sticky-Aktionsleiste als Muster der Tischseite:** Primäraktionen der Tabs „Bestellen"
  und „Kassieren" sitzen in einer Leiste am unteren Viewport-Rand, auf Mobilgeräten
  oberhalb der bereits fixierten Tab-Leiste (`TablePage.tsx`), auf Desktop dasselbe Muster.
  Die Review-Drawer bleiben unverändert.

## Inventory

- `backend/api/kasse/application/command.go:104-112` — `KassensitzungEroeffnen`: mappt
  jeden `GetBetreiber`-Fehler auf `ErrDatabase` (Befund-1-Ursache); `:109-112` behandelt
  unvollständige Stammdaten bereits korrekt als `ErrBetreiberNichtKonfiguriert`
- `backend/repository/settings_repo/repo.go:13-23` — `GetBetreiber` liefert
  `db.ErrNotFound`, wenn nie gespeichert; `db.ErrDatabase` nur bei echten DB-Fehlern
- `backend/api/kasse/http/command_handler.go:82-85` — Handler mappt
  `ErrBetreiberNichtKonfiguriert` bereits auf `betreiber_nicht_konfiguriert`; keine Änderung
- `backend/api/kasse/http/command_handler.go:27-35` — `kassensitzungEroeffnenRequest`:
  `BetragCents int` mit `GTE(0).Required()` (Befund-8-Ursache); `:56-62` dasselbe für
  `IstBestandCents`; `:47-54` Geldtransit hat `GTE(1)` und bleibt unverändert
- `backend/api/kasse/application/command_test.go:34-44,64-75` — Command-Test-Muster mit
  `settingsMock` (Prior Art Befund 1)
- `backend/api/kasse/http/command_handler_test.go` — Handler-Test-Muster (Prior Art
  Befunde 1, 8)
- `backend/sqlc/queries/produkte.sql:11,32,64` — drei `json_agg`-Aggregationen ohne
  `ORDER BY` (Befund-3-Ursache): `GetProdukt`, `GetAlleProdukte`, `GetAktiveProdukte`
- `backend/sqlc/queries/druckstation.sql:1-8` — `GetDruckstationen` und
  `GetKonfigurierteDruckstationen` ohne `ORDER BY` (Befund-4-Ursache)
- `backend/repository/product_repo/repo_test.go`,
  `backend/repository/druckstation_repo/repo_test.go` — Integrationstest-Muster
  (`//go:build integration`, `dbpkg.OpenTestDatabase()`, Reset im Setup)
- `frontend/src/components/ui/sonner.tsx` — Toaster-Wrapper: semantische Icons vorhanden,
  Farben neutral über `--normal-bg`/`--normal-text` (Befund-2-Ursache); eingebunden in
  `frontend/src/App.tsx:12`
- `frontend/src/service/components/TischAuswahlDrawer.tsx:46-55` — Favoriten-Mutation
  invalidiert nur `AKTIVE_TISCHE_MIT_FAVORITEN_KEY` (Befund-6-Ursache)
- `frontend/src/service/table/hooks.ts:55-70` — `AKTIVE_TISCHE_MIT_FAVORITEN_KEY`
  (exportierte Konstante, Muster) und `useMeineTischeState` mit Inline-Key
  `'meine-tische-state'`; konsumiert von `frontend/src/service/TableSelectionPage.tsx:16`
- `frontend/src/service/TablePage.tsx:90-114` — mobile fixierte Tab-Leiste
  (`fixed bottom-…`), `:115,139,152` — `pb-24`-Padding der Tab-Inhalte (Muster für den
  Abstand unter der Sticky-Leiste)
- `frontend/src/service/components/table/BestellungDrawer.tsx:74-84` — bisheriger
  Primär-Button „Bestellung überprüfen" als `DrawerTrigger` oberhalb der Produktliste
- `frontend/src/service/components/table/Zahlung.tsx:74-101` — Aktionszeile des
  Kassieren-Tabs: `AuszahlungDrawer` (sekundär) + `ZahlungDrawer`-Trigger „Kassieren"
  (`ZahlungDrawer.tsx:93-100`)
- `frontend/src/service/components/table/drawerUtils.ts:10-29` — `selectPositionen` /
  `calculateTotalPrice` (liefern Positionsanzahl und Summe für die Sticky-Leiste);
  `BestellungDrawer.tsx:122-150` — `toBestellungData` für den Bestellen-Tab
- `frontend/src/lib/errorMessages.ts:9-10,86-118` — `commonErrorMessages` enthält bereits
  einen generischen Text für `betreiber_nicht_konfiguriert`; `byCode` überschreibt ihn
  seitenspezifisch
- `frontend/src/admin/kasse/KassensitzungPage.tsx:108-122` — Eröffnen-Submit über
  `useFormActionSubmit` ohne `byCode` (Einhängepunkt Befund 1)
- `frontend/src/hooks/use-action-submit.ts`, `use-form-action-submit.ts` —
  `byCode`-Mechanismus der Submit-Hooks
- `frontend/src/service/components/direktverkauf/Direktverkauf.test.tsx` —
  Komponententest-Muster im Service-Bereich (Prior Art Befunde 6, 7)

## Resolved decisions

Aus dem PRD (Entscheidungen vom 2026-06-12) und der Codebasis-Recherche:

- **Phasenzuschnitt:** je Befund eine Phase, Backend-Fixes (1, 3, 4, 8) zuerst — vom PRD
  vorgegeben.
- **Befund 1 braucht nur eine Backend-Zeile + Frontend-Text:** Handler-Mapping und
  Fehlercode existieren bereits; zu ändern ist nur das `db.ErrNotFound`-Mapping im Command
  und die `byCode`-Meldung auf der Kassensitzungs-Seite.
- **Befund 8 nutzt `z.Ptr(...).NotNil(...)`:** im zog-v0.22.0-Quellcode verifiziert
  (Modul-Cache, `pointers.go`); `NotNil` prüft Anwesenheit statt Nicht-Zero.
- **Testumfang:** Backend-Tests (Command, Handler, Repo-Integration) + gezielte
  Frontend-Komponententests (Befunde 6, 7). Toast-Farben und Sticky-Positionierung werden
  manuell geprüft (rein visuell) — gemäß Testing Decisions im PRD.

> **Assumption:** Im Kassieren-Tab wandert nur die Primäraktion („Kassieren",
> `ZahlungDrawer`-Trigger) in die Sticky-Leiste. Der `AuszahlungDrawer`-Button (sekundär,
> nur für Storno-Berechtigte bzw. bei negativem Saldo sichtbar) bleibt oberhalb der
> Positionsliste — das PRD spricht von „der primären Aktion" (Singular) und nennt die
> Auszahlung nicht.

## Open questions / Risks

- **Sonner-Theming (Befund 2):** Ob `richColors` allein in beiden Themes ausreichend
  kontrastiert oder zusätzlich CSS-Variablen (`--success-bg` etc.) nötig sind, zeigt erst
  die manuelle Prüfung — beide Wege sind von sonner vorgesehen, die Entscheidung fällt
  beim Implementieren (Doku konsultieren, Regel 13 AGENTS.md).
- **Sticky-Leiste vs. Drawer-Trigger (Befund 7):** Die bisherigen Buttons sind
  `DrawerTrigger asChild`-Kinder. Beim Umzug in die Sticky-Leiste muss der Drawer ggf.
  kontrolliert (per `open`-State) geöffnet werden, wenn Trigger und Drawer nicht mehr
  im selben Teilbaum liegen — beide Komponenten halten ihren `open`-State bereits selbst,
  das Muster ist vorhanden.

---

## Phase 1: Befund 1 — Betreiber nicht konfiguriert als Client-Fehler

**User stories**: 1, 2, 3, 4

### Context

- `backend/api/kasse/application/command.go:104-112` — `GetBetreiber`-Fehler wird
  undifferenziert zu `ErrDatabase` (500) mit Error-Log
- `backend/repository/settings_repo/repo.go:13-23` — liefert `db.ErrNotFound`, wenn
  Betreiber nie gespeichert
- `backend/api/kasse/http/command_handler.go:82-85` — Mapping auf
  `betreiber_nicht_konfiguriert` existiert; keine Änderung
- `backend/api/kasse/application/command_test.go:34-44` — `settingsMock` (Prior Art)
- `frontend/src/admin/kasse/KassensitzungPage.tsx:108-122` — Eröffnen-Submit ohne `byCode`
- `frontend/src/lib/errorMessages.ts:9-10` — generische Common-Meldung für den Fehlercode

### What to build

Im Kasse-Command wird `db.ErrNotFound` aus `GetBetreiber` als
`ErrBetreiberNichtKonfiguriert` interpretiert (Warn-Log statt Error-Log); nur andere
Fehler bleiben `ErrDatabase`. Der bestehende Handler liefert dann automatisch den
Client-Fehler `betreiber_nicht_konfiguriert`. Auf der Kassensitzungs-Seite bekommt der
Eröffnen-Submit eine `byCode`-Meldung mit konkreter Handlungsanweisung (Betreiber-Stammdaten
in den Einstellungen pflegen, dann Kassensitzung eröffnen).

### Acceptance criteria

- [x] Command-Test: `GetBetreiber` mit `db.ErrNotFound` → `ErrBetreiberNichtKonfiguriert`,
      nicht `ErrDatabase`
- [x] Command-Test: `GetBetreiber` mit `db.ErrDatabase` → weiterhin `ErrDatabase`
- [x] Handler-Test: Response ist Client-Fehler (4xx) mit Code `betreiber_nicht_konfiguriert`
- [x] Kein Error-Log-Eintrag für den Nicht-konfiguriert-Fall (nur Warn)
- [x] Kassensitzungs-Seite zeigt für den Fehlercode die Handlungsanweisung (deutsch,
      verweist auf die Einstellungen)
- [x] `make check` grün

---

## Phase 2: Befund 3 — Stabile Produktvarianten-Reihenfolge

**User stories**: 8, 9, 10

### Context

- `backend/sqlc/queries/produkte.sql:11,32,64` — drei `json_agg` ohne `ORDER BY`
- `backend/repository/product_repo/repo_test.go` — Integrationstest-Muster
  (`//go:build integration`)

### What to build

Alle drei Varianten-Aggregationen (`GetProdukt`, `GetAlleProdukte`, `GetAktiveProdukte`)
sortieren innerhalb des `json_agg` nach Varianten-ID aufsteigend
(`json_agg(… ORDER BY id)`). Danach sqlc-Code neu generieren (`make sqlc`).

### Acceptance criteria

- [x] Repo-Integrationstest: Nach einem `UpdateVariante` auf eine mittlere Variante bleibt
      die Reihenfolge der Varianten (nach ID) in allen drei Queries stabil
- [x] `make sqlc` ausgeführt, generierter Code committed, `sqlc/dbgen/` nicht von Hand
      editiert

---

## Phase 3: Befund 4 — Stabile Druckstationen-Reihenfolge

**User stories**: 11, 12

### Context

- `backend/sqlc/queries/druckstation.sql:1-8` — beide Queries ohne `ORDER BY`
- `backend/repository/druckstation_repo/repo_test.go` — Integrationstest-Muster mit
  Reset-Setup

### What to build

`GetDruckstationen` und `GetKonfigurierteDruckstationen` sortieren mit
`ORDER BY kategorie`. Danach `make sqlc`.

### Acceptance criteria

- [x] Repo-Integrationstest: Nach `UpsertDruckstation` auf eine mittlere Kategorie bleibt
      die Reihenfolge (nach Kategorie, Enum-Deklaration) stabil
- [x] `make sqlc` ausgeführt, generierter Code committed
- [x] `make verify` grün

---

## Phase 4: Befund 8 — 0-Cent-Beträge gültig

**User stories**: 21, 22, 23, 24

### Context

- `backend/api/kasse/http/command_handler.go:27-35` — `BetragCents int` +
  `GTE(0).Required()` (Eröffnung); `:56-62` `IstBestandCents` (Kassensturz); `:80,138` —
  Übergabe an den Command (nach Umstellung dereferenzieren)
- `backend/api/kasse/http/command_handler_test.go` — Handler-Test-Muster
- zog v0.22.0 `pointers.go:31,126` — `z.Ptr(schema)` + `.NotNil(options…)`

### What to build

`BetragCents` (Kassensitzung eröffnen) und `IstBestandCents` (Kassensturz) werden in den
Request-DTOs zu `*int`; die Schemas nutzen `z.Ptr(z.Int().GTE(0, …)).NotNil(…)` mit den
bestehenden deutschen Fehlermeldungen. Die Handler dereferenzieren nach erfolgreicher
Validierung. Geldtransit (`GTE(1)`) bleibt unverändert. Command-Signaturen bleiben `int`.

### Acceptance criteria

- [ ] Handler-Test: Eröffnung mit `betragCents: 0` → Erfolg
- [ ] Handler-Test: Kassensturz mit `istBestandCents: 0` → Erfolg
- [ ] Handler-Test: fehlendes Betragsfeld → 400 mit Feld-Fehlermeldung
- [ ] Handler-Test: negativer Betrag → 400 mit Feld-Fehlermeldung
- [ ] Geldtransit-Validierung (`≥ 1`) unverändert
- [ ] `make check` grün

---

## Phase 5: Befund 2 — Semantische Toast-Farben

**User stories**: 5, 6, 7

### Context

- `frontend/src/components/ui/sonner.tsx` — zentraler Toaster-Wrapper (einzige
  Änderungsstelle); Icons bereits semantisch, Farben neutral
- `frontend/src/App.tsx:12` — Einbindung `<Toaster position="top-right" />`

### What to build

Der Toaster-Wrapper erhält semantische Farbgebung für success, info, warning und error über
die von sonner vorgesehene Theming-Möglichkeit (`richColors` bzw. CSS-Variablen — aktuelle
sonner-Doku konsultieren), abgestimmt auf Light- und Dark-Theme. Aufrufstellen bleiben
unverändert.

### Acceptance criteria

- [ ] Manuelle Prüfung: `toast.success` grün, `toast.error` rot, `toast.warning` und
      `toast.info` davon unterscheidbar — im Light- und im Dark-Theme lesbar
- [ ] Keine Änderung an Aufrufstellen
- [ ] `make lint` grün

---

## Phase 6: Befund 6 — Favoriten sofort sichtbar

**User stories**: 13, 14, 15

### Context

- `frontend/src/service/components/TischAuswahlDrawer.tsx:46-55` — Favoriten-Mutation,
  invalidiert nur `AKTIVE_TISCHE_MIT_FAVORITEN_KEY`
- `frontend/src/service/table/hooks.ts:55-70` — Key-Konstanten-Muster und
  `useMeineTischeState` (Inline-Key `'meine-tische-state'`)
- `frontend/src/service/TableSelectionPage.tsx:16` — Konsument von `useMeineTischeState`
- `frontend/src/service/components/direktverkauf/Direktverkauf.test.tsx` —
  Komponententest-Muster

### What to build

Der Query-Key von „Meine Tische" wird als exportierte Konstante in `hooks.ts` definiert
(analog `AKTIVE_TISCHE_MIT_FAVORITEN_KEY`). Die Favoriten-Mutation invalidiert nach Erfolg
beide Queries — Tischauswahl-Liste und „Meine Tische".

### Acceptance criteria

- [ ] Komponententest: Nach dem Favoriten-Toggle werden beide Queries invalidiert bzw.
      „Meine Tische" zeigt den neuen Stand ohne Reload
- [ ] Stern in der Tischauswahl und „Meine Tische" zeigen nach dem Toggle denselben Stand
      (manuelle Prüfung)
- [ ] `make lint` und Frontend-Tests grün

---

## Phase 7: Befund 7 — Sticky-Aktionsleiste auf der Tischseite

**User stories**: 16, 17, 18, 19, 20

### Context

- `frontend/src/service/TablePage.tsx:90-114` — mobile fixierte Tab-Leiste (Positionierungs-
  und `z-index`-Referenz); `:115,139,152` — `pb-24`-Muster für Listen-Padding
- `frontend/src/service/components/table/BestellungDrawer.tsx:74-84` — bisheriger Trigger
  „Bestellung überprüfen" (wird ersetzt)
- `frontend/src/service/components/table/Zahlung.tsx:74-101` — Aktionszeile Kassieren-Tab;
  `ZahlungDrawer.tsx:93-100` — Trigger „Kassieren" (wird ersetzt), `AuszahlungDrawer`
  bleibt
- `frontend/src/service/components/table/drawerUtils.ts:10-29`,
  `BestellungDrawer.tsx:122-150` — Positionsanzahl und Summe der aktuellen Auswahl
- `frontend/src/service/components/direktverkauf/Direktverkauf.test.tsx` —
  Komponententest-Muster

### What to build

Eine Sticky-Aktionsleiste am unteren Viewport-Rand für die Tabs „Bestellen" und
„Kassieren" — auf Mobilgeräten oberhalb der fixierten Tab-Leiste, auf Desktop dasselbe
Muster. Die Leiste zeigt Positionsanzahl und Summe der aktuellen Auswahl, ist ohne Auswahl
deaktiviert und öffnet wie bisher den jeweiligen Review-Drawer (Bestellung bzw. Zahlung);
die bisherigen Buttons oberhalb der Listen entfallen. Der `AuszahlungDrawer` bleibt an
seiner Stelle. Die Listen erhalten ausreichend Padding nach unten (analog zum bestehenden
`pb-24`-Muster), damit keine Einträge hinter den Leisten verschwinden.

### Acceptance criteria

- [ ] Komponententest: Leiste zeigt Positionsanzahl und Summe der Auswahl
- [ ] Komponententest: Leiste ist ohne gewählte Position deaktiviert
- [ ] Klick auf die Leiste öffnet den Review-Drawer (Bestellung bzw. Zahlung); Drawer
      unverändert
- [ ] Manuelle Prüfung (mobiler Viewport): Leiste verdeckt die Tab-Leiste nicht, kein
      Listeneintrag verschwindet hinter den Leisten, Aktion ohne Scrollen erreichbar
- [ ] Auszahlung weiterhin im Kassieren-Tab erreichbar
- [ ] `make lint` und Frontend-Tests grün
