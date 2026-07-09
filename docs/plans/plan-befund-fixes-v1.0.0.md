# Plan: Befund-Fixes QA v1.0.0

> Source: docs/plans/befund-report-qa-v1.0.0.md (Befund-Report der einmaligen QA-Durchführung, inkl. Abschnitt 5 „Entscheidungen aus der Klärung")

## Goal

Alle Befunde des QA-Befund-Reports v1.0.0 abarbeiten (H1–H2, M1–M5, N1–N7) und die Toolchain auf go1.26.5 vereinheitlichen, sodass beide GitHub-Workflows (CI und Security Scans) grün sind. Umsetzung phasenweise, jede Phase einzeln verifizierbar und committbar. Nach Abschluss wird der Befund-Report gelöscht.

## Architectural decisions

- **Toolchain**: einheitlich go1.26.5 in allen 5 `go.mod`, allen `go-version`-Pins der Workflows und allen `golang:*-alpine`-Dockerfiles. Das Image-Tag `golang:1.26.5-alpine` existiert auf Docker Hub (geprüft 2026-07-09).
- **npm-Audit-Policy**: `pnpm audit --prod --audit-level high` bricht den CI-Job (blockierend); das volle Audit inkl. Dev-Deps läuft zusätzlich als informativer Schritt, der nie fehlschlägt. Der bisherige „bewusst rot"-Kommentar im Workflow wird entsprechend ersetzt.
- **404/Fehler-Routing**: gestaltete deutsche 404-Seite über `errorElement` an der Root-Route plus Catch-all-Route `path: '*'` in `frontend/src/routes.ts`; dazu ein `HydrateFallback`, der die Konsolenwarnung beim Erstladen beseitigt.
- **Druckauftrags-Referenz**: Mapping von technischer Referenz zu fachlichem Text ausschließlich im Frontend. Die Backend-API liefert weiterhin den rohen Wert (`bestellung-aufgenommen:86` usw.); es gibt genau 5 Formate (siehe Inventory).
- **Tabs-Scroll-Affordance**: horizontale Scrollbarkeit mit Rand-Fade und Chevron als wiederverwendbare Erweiterung an `frontend/src/components/ui/tabs.tsx`, genutzt von beiden Auswertungs-Ansichten.
- **Kontrast-Politik**: UI bleibt grün (Branding-Entscheidung). WCAG-AA-Schwellen: 4.5:1 für Normaltext, 3:1 für großen Text und UI-Komponenten. Prüfung rein automatisiert, einmalig, kein dauerhafter Guard.
- **UI-Verifikation**: alle Layout-Phasen werden per Playwright (gebündeltes Chromium, headless) automatisiert verifiziert; Viewports 393x851 (mobil) und 1280x800 (Desktop). Keine manuelle Abnahme.

## Inventory

- `backend/go.mod`, `resolver/go.mod`, `reverse-proxy/go.mod`, `windows/starter/go.mod`, `windows/relay/go.mod` — alle `go 1.26.0`
- `.github/workflows/ci.yml` (8x `go-version: 1.26.0`), `.github/workflows/security-scans.yml` (2x), `.github/workflows/release.yml` (1x)
- `backend/Dockerfile`, `resolver/Dockerfile`, `reverse-proxy/Dockerfile` — `golang:1.26.4-alpine`
- `.github/workflows/security-scans.yml — vuln-scan-npm` — `pnpm audit --audit-level high`, laut Kommentar bewusst rot (4 bekannte Dev-Dep-Advisories: hono via shadcn-CLI, undici via jsdom/vitest)
- `frontend/src/admin/products/NewProductDialog.tsx`, `frontend/src/admin/tables/NewTischDialog.tsx`, `frontend/src/admin/users/NewUserDialog.tsx` — FAB als `fixed bottom-[calc(1rem+env(safe-area-inset-bottom,0px))] right-4 … z-50`
- `frontend/src/admin/products/AdminProductsPage.tsx`, `frontend/src/admin/tables/AdminTablesPage.tsx`, `frontend/src/admin/users/AdminUsersPage.tsx` — Listen unter dem FAB
- `frontend/src/admin/products/ProductItem.tsx`, `frontend/src/admin/tables/TischItem.tsx`, `frontend/src/admin/users/UserItem.tsx` — Edit-/Löschen-Icons; Lösch-`AlertDialog` mit Objektname existiert bereits (M4-Verifikation)
- `frontend/src/service/TablePage.tsx` — `TabsContent` mit `pb-40 md:pb-28` (reicht laut Befund nicht) und fixierte Leisten
- `frontend/src/service/components/table/StickyActionBar.tsx` — Leiste „Bestellung überprüfen", `fixed inset-x-0 bottom-[calc(4rem+…)]`
- `frontend/src/admin/reporting/LiveReportingSection.tsx — TabsList` und `frontend/src/admin/reporting/ReportingResults.tsx — TabsList` — die abgeschnittenen Tab-Leisten
- `frontend/src/components/ui/tabs.tsx` — shadcn-Tabs-Basis
- `frontend/src/routes.ts — createBrowserRouter` — kein `errorElement`, kein `HydrateFallback`, kein Catch-all
- `frontend/src/pages/LoginPage.tsx`, `frontend/src/components/common/LoginForm.tsx` — Login-Karte (N3), Primär-CTA (N2)
- `frontend/src/index.css` — Farb-Token (`:root` und Dark-Variante)
- `frontend/src/admin/settings/DruckstationConfigPage.tsx` — rendert `Referenz: {auftrag.referenz}` roh
- Referenz-Produzenten (alle 5 Formate): `backend/api/druck/bondruck/application/arbeitsbon_policy.go` (`bestellung-aufgenommen:%d`, `direktverkauf-getaetigt:%d`) und `backend/api/druck/beleg/application/kassenbeleg_command.go` (`zahlung-kassiert:%d`, `direktverkauf-getaetigt:%d`, `direktverkauf-storniert:%d`, `stornierung-erteilt:%d`)
- `backend/seed/seed_integration_test.go — TestSeedRun_ErstlaufUndGuard(), TestResetAndSeed_LeertUndSeedetNeu()` — M5; leert `tse_konfiguration` ohne Default-Row
- `backend/.golangci.yml`, `Makefile — lint-backend-full, check-backend` — Lint läuft ohne `--build-tags=integration`
- `e2e/tests/admin-benutzer-verwalten.spec.ts` — OTP-Assertion über `p.text-3xl.tracking-widest`
- `e2e/support/servicekraft.ts — zeileMit(), settleAlleOffenenTische()` plus lokale `zeileMit`-Kopien in `e2e/tests/bestellen-kassieren.spec.ts` und `e2e/tests/kassieren-fehlerpfade.spec.ts`
- `backend/domain/kasse/replay_fuzz_test.go — FuzzApplyEvent` und `backend/api/druck/bondruck/application/escpos/formatter_fuzz_test.go — assertQRCommandLengths`
- `docs/leitfaden/datenaufbewahrung.md` — beschreibt den DSFinV-K-Export-Ort bereits korrekt (Dashboard, Historische Auswertung); Drift-Kandidaten: `docs/verfahrensdokumentation.md`, `docs/compliance.md`, `docs/leitfaden/checkliste.md`
- `frontend/src/admin/finanzamt/FinanzamtPage.tsx` — existiert, enthält Betreiber-/TSE-/Signatur-Infos, keinen Export; bleibt unverändert

## Resolved decisions

Aus der Klärung vom 2026-07-09 (vollständig in Abschnitt 5 des Befund-Reports):

- Scope: alle Befunde. M4 im Kern erledigt (Dialog existiert), Rest fällt mit H1 zusammen; N1 fällt mit M2 ab.
- TSE-Burst (p95=7192ms bei Burst-24): abgenommen wie dokumentiert, kein Fix.
- M1: Tabs scrollbar mit Fade/Chevron (kein Umbruch, kein Dropdown).
- N2: systematischer Token-Audit, automatisiert, einmalig, kein Regressions-Guard, keine manuelle Abnahme.
- N5: Doku an Ist-Ort angleichen; keine Verschiebung des Exports und keine Änderung an `/admin/finanzamt`.
- npm-Audit: prod-blockierend plus informatives Vollaudit (siehe Architectural decisions).
- Human-Gate-Punkte (Ops-Smoke, prod-restore, TLS, Hardware/Windows/Zwei-Geräte): nicht Teil dieses Plans, bleiben im Rest-Guide.
- Toolchain-Bump auf go1.26.5: entschieden; lokal ist go1.26.5 bereits installiert.

## Open questions / Risks

- H2: `pb-40` existiert bereits und reicht nicht. Der Implementer soll die tatsächliche Gesamthöhe der übereinanderliegenden Leisten im Browser messen statt eine neue Konstante zu raten; ggf. ist die Ursache eine dynamische Leistenhöhe.
- N6: Umstellung auf Testattribute (`data-testid`) berührt Produktions-JSX; die Änderungen müssen rein additiv bleiben.
- Security-Scans-Workflow läuft mit Matrix-fail-fast: im letzten Lauf wurden 3 Jobs abgebrochen. Erst der Lauf nach Phase 1 zeigt alle Module wirklich grün.

---

## Phase 1: Toolchain-Konsistenz und Security-Scans grün

### Context

- `backend/go.mod`, `resolver/go.mod`, `reverse-proxy/go.mod`, `windows/starter/go.mod`, `windows/relay/go.mod` — `go 1.26.0`
- `.github/workflows/ci.yml`, `.github/workflows/security-scans.yml`, `.github/workflows/release.yml` — `go-version: 1.26.0`-Pins
- `backend/Dockerfile`, `resolver/Dockerfile`, `reverse-proxy/Dockerfile` — `golang:1.26.4-alpine`
- `.github/workflows/security-scans.yml — vuln-scan-npm` — Audit-Schritt und „bewusst rot"-Kommentar

### What to build

Alle Go-Versionsangaben auf 1.26.5 vereinheitlichen: die 5 `go.mod`-Direktiven, alle 11 Workflow-Pins, die 3 Builder-Images. Den `vuln-scan-npm`-Job auf zwei Schritte umbauen: `pnpm audit --prod --audit-level high` blockierend, danach das volle Audit als informativer Schritt (Exit-Code neutralisiert, Ausgabe im Log); den Workflow-Kommentar an die neue Policy anpassen. Nach Commit und Push den Security-Scans-Lauf beobachten.

### Acceptance criteria

- [x] Alle 5 `go.mod` stehen auf `go 1.26.5`, `go mod tidy -diff` ist in allen Modulen sauber
- [x] Kein `go-version: 1.26.0`-Pin und kein `golang:1.26.4`-Image mehr im Repo (`grep -rn "1.26.0\|1.26.4" .github backend/Dockerfile resolver/Dockerfile reverse-proxy/Dockerfile` ohne Treffer)
- [x] `govulncheck` (v1.5.0) lokal in allen 5 Modulen ohne Stdlib-Findings
- [x] `make verify` grün
- [ ] Nach Push: Workflow „Security Scans" komplett grün (alle `vuln-scan-go`-Jobs, `vuln-scan-go-windows`, `vuln-scan-npm`), Workflow „CI" grün

---

## Phase 2: H1 — FAB-Overlap in den Admin-Listen (inkl. M4-Rest)

### Context

- `frontend/src/admin/products/NewProductDialog.tsx`, `frontend/src/admin/tables/NewTischDialog.tsx`, `frontend/src/admin/users/NewUserDialog.tsx` — der fixed FAB
- `frontend/src/admin/products/AdminProductsPage.tsx`, `frontend/src/admin/tables/AdminTablesPage.tsx`, `frontend/src/admin/users/AdminUsersPage.tsx` — Listen ohne Unterkanten-Freiraum
- `frontend/src/admin/products/ProductItem.tsx`, `frontend/src/admin/tables/TischItem.tsx`, `frontend/src/admin/users/UserItem.tsx` — Edit-/Löschen-Icons (Touch-Ziele)

### What to build

Die drei Admin-Listen erhalten unten so viel Freiraum (Padding in FAB-Höhe plus Safe-Area-Inset), dass die letzte Karte beim Scrollen vollständig über dem FAB sichtbar und bedienbar ist; die Lösung gehört an eine gemeinsame Stelle (Listen-Container bzw. geteiltes Layout), nicht dreimal kopiert. Edit- und Löschen-Icons bekommen Touch-Ziele von mindestens 44px mit sichtbarem Abstand zueinander (M4-Rest).

### Acceptance criteria

- [ ] Playwright (393x851 und 1280x800): auf allen drei Listen ans Ende gescrollt überschneiden sich die Bounding-Boxen von FAB und Edit-/Löschen-Icons der letzten Karte nicht
- [ ] Playwright: Löschen-Icon der letzten Karte ist ohne `force: true` klickbar (der bestehende `AlertDialog` öffnet sich; Dialog abbrechen)
- [ ] Edit-/Löschen-Buttons haben eine Hit-Area von mindestens 44x44px
- [ ] `make verify` grün (inkl. Frontend-Tests und Lint)

---

## Phase 3: H2 — Leisten-Overlap im Tisch-Detail

### Context

- `frontend/src/service/TablePage.tsx` — `TabsContent` mit `pb-40 md:pb-28`, fixierte Bestell-Leiste und Tab-Leiste
- `frontend/src/service/components/table/StickyActionBar.tsx` — Leiste „Bestellung überprüfen"

### What to build

Im Tisch-Detail (Tabs Bestellen und Kassieren) endet die Produktliste heute hinter den beiden fixierten Leisten. Die tatsächliche Gesamthöhe der übereinanderliegenden Leisten im Browser messen und den unteren Freiraum der Liste daran ausrichten (inkl. Safe-Area-Inset und Sicherheitsabstand), sodass die letzte Zeile vollständig sichtbar und antippbar ist. Falls die Leistenhöhe variiert, den Freiraum von der realen Höhe ableiten statt eine feste Klasse zu setzen.

### Acceptance criteria

- [ ] Playwright (393x851, als Servicekraft im Tisch-Detail): ans Listenende gescrollt überschneiden sich die Bounding-Boxen der letzten Produktzeile und der Leisten nicht (Tabs Bestellen und Kassieren)
- [ ] Playwright: die letzte Produktzeile ist ohne `force: true` antippbar (Position erhöht sich)
- [ ] Bestehende E2E-Suite grün (`make test-e2e E2E_BASE_URL=http://localhost:8093`)
- [ ] `make verify` grün

---

## Phase 4: M1 — Auswertungs-Tabs scrollbar mit Fade/Chevron

### Context

- `frontend/src/admin/reporting/LiveReportingSection.tsx — TabsList` — Live-Dashboard-Tabs
- `frontend/src/admin/reporting/ReportingResults.tsx — TabsList` — Tabs der Historischen Auswertung
- `frontend/src/components/ui/tabs.tsx` — gemeinsame Tabs-Basis

### What to build

Die Tab-Leisten (Übersicht/Servicekräfte/Tische/Stornierungen) werden auf schmalen Viewports horizontal scrollbar mit deutlicher Affordance: Rand-Fade auf der abgeschnittenen Seite und Chevron als Scroll-Hinweis, der verschwindet, wenn das Ende erreicht ist. Die Erweiterung lebt in der gemeinsamen Tabs-Komponente (oder einem Wrapper daneben) und wird von beiden Auswertungs-Ansichten genutzt. Badges (z. B. bei „Tische") werden nicht mehr abgeschnitten. Desktop-Darstellung bleibt unverändert.

### Acceptance criteria

- [ ] Playwright (393x851, `/admin/auswertung`): Tab „Stornierungen" ist per horizontalem Scroll erreichbar und anklickbar, in beiden Tab-Leisten
- [ ] Playwright (393x851): Fade/Chevron-Affordance ist im DOM vorhanden, solange rechts Tabs verborgen sind, und verschwindet am Scroll-Ende
- [ ] Playwright (1280x800): alle 4 Tabs ohne Scroll sichtbar, keine Affordance
- [ ] `make verify` grün

---

## Phase 5: M2 + N1 — Deutsche 404-Seite und HydrateFallback

### Context

- `frontend/src/routes.ts — createBrowserRouter` — kein `errorElement`, kein Catch-all, kein `HydrateFallback`
- `frontend/src/pages/LoginPage.tsx` — Referenz für Seiten-Layout/Branding

### What to build

Eine gestaltete deutsche Fehlerseite im App-Branding: Überschrift und Erklärtext auf Deutsch, Button „Zurück zur Startseite" (navigiert auf `/`). Sie dient als `errorElement` der Root-Route (fängt Render-/Loader-Fehler) und als Komponente einer Catch-all-Route `path: '*'` (unbekannte Pfade wie `/gibtsnicht`). Zusätzlich ein `HydrateFallback` (schlichter gebrandeter Ladezustand), damit die React-Router-Konsolenwarnung beim Erstladen verschwindet.

### Acceptance criteria

- [ ] Playwright ausgeloggt auf `/gibtsnicht`: deutsche Fehlerseite mit Button „Zurück zur Startseite", Klick landet auf `/login` (via Redirect von `/`)
- [ ] Kein „Unexpected Application Error!"-Rohbildschirm mehr erreichbar
- [ ] Playwright-Konsole beim Erstladen ohne `HydrateFallback`-Warnung (mobil und Desktop)
- [ ] `make verify` und E2E-Suite grün

---

## Phase 6: M3 — Druckauftrags-Referenz fachlich rendern

### Context

- `frontend/src/admin/settings/DruckstationConfigPage.tsx` — rendert `Referenz: {auftrag.referenz}` roh
- Referenz-Formate (Backend, unverändert lassen): `bestellung-aufgenommen:%d`, `direktverkauf-getaetigt:%d` (`backend/api/druck/bondruck/application/arbeitsbon_policy.go`), `zahlung-kassiert:%d`, `direktverkauf-storniert:%d`, `stornierung-erteilt:%d` (`backend/api/druck/beleg/application/kassenbeleg_command.go`)

### What to build

Eine kleine Frontend-Mapping-Funktion übersetzt die 5 bekannten Referenz-Formate in fachliche Texte: „Bestellung Nr. 86", „Zahlung Nr. …", „Direktverkauf Nr. …", „Direktverkauf-Storno Nr. …", „Stornierung Nr. …". Unbekannte Formate fallen auf den Rohwert zurück. Der technische Rohwert bleibt als `title`-Attribut erreichbar. Die Funktion bekommt einen Vitest-Test über alle 5 Formate plus Fallback.

### Acceptance criteria

- [ ] Vitest-Test: alle 5 Formate werden korrekt übersetzt, unbekanntes Format fällt auf den Rohwert zurück
- [ ] In „Fehlgeschlagene Druckaufträge" erscheint für bekannte Formate kein roher kebab-case-Bezeichner mehr
- [ ] Rohwert per `title`-Attribut einsehbar
- [ ] `make verify` grün

---

## Phase 7: N2 + N3 — Kontrast-Token-Audit und Login-Karte

### Context

- `frontend/src/index.css` — Farb-Token (`:root` und Dark-Variante), u. a. `--primary`, `--muted-foreground`
- `frontend/src/pages/LoginPage.tsx`, `frontend/src/components/common/LoginForm.tsx` — Login-Layout und Primär-CTA

### What to build

Ein einmaliges Audit-Skript (Scratch, wird nicht committet) liest die Token aus `index.css`, berechnet die WCAG-Kontrast-Ratios aller in der App verwendeten Vordergrund/Hintergrund-Paare (light und dark) und listet Verstöße gegen AA (4.5:1 Text, 3:1 großer Text/UI-Komponenten). Verstöße werden an den Token-Werten korrigiert: Grünton bleibt (Branding), wird aber dunkler/gesättigter, Sekundärtexte dunkler. Danach Re-Run des Skripts als Verifikation. Zusätzlich die Login-Karte auf Mobile höher und kompakter platzieren (weniger Leerraum oben, kürzerer Daumenweg). Kein dauerhafter Guard, keine manuelle Abnahme.

### Acceptance criteria

- [ ] Audit-Re-Run: alle geprüften Token-Paare erfüllen AA; die geprüften Paare mit Ratios stehen als Kurzliste in der Commit-Message der Phase
- [ ] Primär-CTA („Anmelden") erreicht mindestens 4.5:1 gegen seinen Hintergrund und bleibt grün
- [ ] Playwright (393x851, `/login`): Oberkante der Login-Karte liegt messbar höher als vorher (Bounding-Box-Vergleich)
- [ ] `make verify` grün (Frontend-Snapshot-/Unit-Tests angepasst, falls Token-Änderungen sie berühren)

---

## Phase 8: M5 + N4 — Seed-Reset deterministisch und Integration-Lint sauber

### Context

- `backend/seed/seed_integration_test.go — TestSeedRun_ErstlaufUndGuard(), TestResetAndSeed_LeertUndSeedetNeu()` — leert `tse_konfiguration` ohne Default-Row
- `backend/api/test` — Reset-Endpoint der E2E-Suite
- `backend/.golangci.yml`, `Makefile — lint-backend-full, check-backend` — Lint ohne Integration-Tag; ~24 Altlasten (errcheck/gosec/errorlint) in `*_integration_test.go`

### What to build

Der Reset (Testendpoint und Integrationstests) stellt einen deterministischen Ausgangszustand her, der die Default-Row in `tse_konfiguration` einschließt, sodass beliebige Wiederholungen der Seed-Integrationstests grün bleiben. Die Lint-Altlasten in den Integration-Testdateien beheben (unbehandelte Fehler wie `db.Close` explizit behandeln oder begründet verwerfen) und die Backend-Lint-Aufrufe in `Makefile` um `--build-tags=integration` erweitern, damit die Dateien dauerhaft mitgeprüft werden.

### Acceptance criteria

- [ ] Seed-Integrationstests zweimal direkt hintereinander grün (gleiche DB, kein manuelles Aufräumen)
- [ ] `golangci-lint run --build-tags=integration` im Backend ohne Findings
- [ ] Backend-Lint in `check-backend` und `lint-backend-full` läuft mit `--build-tags=integration`
- [ ] `make verify` und `make check-integration` grün

---

## Phase 9: N6 + N7 — E2E- und Fuzz-Härtung

### Context

- `e2e/tests/admin-benutzer-verwalten.spec.ts` — OTP-Assertion über `p.text-3xl.tracking-widest`, prüft nur nicht-leer
- `e2e/support/servicekraft.ts — zeileMit()` — breiter `div`-`hasText`-Filter mit `.last()`; lokale Kopien in `e2e/tests/bestellen-kassieren.spec.ts` und `e2e/tests/kassieren-fehlerpfade.spec.ts`
- `backend/domain/kasse/replay_fuzz_test.go — FuzzApplyEvent` — assertet nur Panic-Freiheit
- `backend/api/druck/bondruck/application/escpos/formatter_fuzz_test.go — assertQRCommandLengths` — durch `GS(k`-Präfix in der QR-Payload fehlleitbar

### What to build

E2E: Die OTP-Assertion prüft das 6-stellige Ziffernformat (`/^\d{6}$/`). Die an Tailwind-Klassen gekoppelten Selektoren (OTP-Absatz, `div.space-y-4` u. ä.) wechseln auf `data-testid`-Attribute; die Attribute werden rein additiv im Produktions-JSX ergänzt. Die lokalen `zeileMit`-Kopien werden auf die Version in `e2e/support/servicekraft.ts` zusammengeführt.
Fuzz: `assertQRCommandLengths` parst die ESC/POS-Kommandostruktur, statt auf ein Payload-kollisionsanfälliges Präfix zu matchen. `FuzzApplyEvent` prüft zusätzlich semantische Invarianten des Replays (mindestens: Saldo nie negativ, Positionsmengen konsistent zu den angewendeten Events).

### Acceptance criteria

- [ ] OTP-Assertion schlägt bei nicht-6-stelligem Wert fehl (Format-Regex statt nicht-leer)
- [ ] Die im Befund genannten Tailwind-Klassen-Selektoren sind durch `data-testid` ersetzt; keine `zeileMit`-Dubletten mehr in den Spec-Dateien
- [ ] `assertQRCommandLengths` erkennt eine QR-Payload, die selbst mit `GS(k` beginnt, korrekt (Regressionstest vorhanden)
- [ ] `FuzzApplyEvent` prüft die genannten Invarianten; `make fuzz` grün (4 Targets je 90s, kein Crasher)
- [ ] E2E-Suite und `make verify` grün

---

## Phase 10: N5 + Abschluss — Doku-Angleich und Report-Abbau

### Context

- `docs/leitfaden/datenaufbewahrung.md` — beschreibt den Export-Ort bereits korrekt (nicht anfassen, außer als Referenzformulierung)
- `docs/verfahrensdokumentation.md`, `docs/compliance.md`, `docs/leitfaden/checkliste.md` — Drift-Kandidaten für „Export in der Finanzamt-Ansicht"
- `docs/plans/befund-report-qa-v1.0.0.md` — laut Kopfzeile nach Abarbeitung zu löschen

### What to build

Doku-Sweep über `docs/` nach Aussagen, der DSFinV-K-Export liege in der Finanzamt-Ansicht (`/admin/finanzamt`), und Angleich an den Ist-Ort: Export im Auswertungs-Dashboard (`/admin/auswertung`, Abschnitt Historische Auswertung). Die Finanzamt-Ansicht selbst und ihre Beschreibung bleiben unverändert. Abschließend den Befund-Report löschen und die Checkboxen dieses Plans final abgleichen; danach Push und Kontrolle, dass CI und Security Scans grün sind.

### Acceptance criteria

- [ ] `grep -rni "finanzamt" docs/` liefert keine Stelle mehr, die den DSFinV-K-Export in der Finanzamt-Ansicht verortet
- [ ] Doku-Änderungen folgen dem Doku-Stil (minimal, keine Slop-Syntax)
- [ ] `docs/plans/befund-report-qa-v1.0.0.md` ist gelöscht
- [ ] Nach Push: Workflows „CI" und „Security Scans" grün
