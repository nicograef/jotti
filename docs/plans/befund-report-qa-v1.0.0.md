# Befund-Report QA-Durchführung v1.0.0

Stand: 2026-07-09

Arbeitsdokument. Es hält die Befunde der einmaligen QA-Durchführung fest und wird nach Abarbeitung der Fixes gelöscht.

## 1. Kurzfazit

Die automatisierten Suiten sind grün: E2E (23/23), `make verify` (alle 5 Go-Module inkl. Integrationstests, Frontend inkl. 222 Vitest-Tests) und Fuzz (4 Targets, je 90s, kein Crasher). Kein Blocker. Die Kernpfade Bestellen, Kassieren und Historie funktionieren.

Die Befunde konzentrieren sich auf mobile Layout-Probleme: ein schwebender Aktionsbutton (FAB) verdeckt in den Admin-Listen systematisch die Edit-/Löschen-Icons der Karten, die untere Bestell-/Tab-Leiste im Tisch-Detail überlappt die Produktliste, und die Auswertungs-Tabs sind auf schmalen Viewports abgeschnitten (Tab „Stornierungen" unsichtbar). Dazu kommen eine unstylte englische Fehlerseite bei unbekannten Routen, ein roher Event-Bezeichner als Druckauftrags-Referenz und kleinere Kontrast-/Konsolen-Themen.

Die Sicherheits-Scans sind formal rot, aber ohne projekteigenen Code mit Call-Pfad: nur Go-Stdlib-Lücken (Toolchain-Bump nötig) und Frontend-Dev-Dependencies. Kein Produktionscode betroffen.

Befunde nach Schwere: 0 Blocker, 2 Hoch, 5 Mittel, 7 Niedrig.

## 2. Suiten-Protokoll

| Suite | Kommando | Ergebnis | Quelle / Datum |
| --- | --- | --- | --- |
| E2E | `make test-e2e E2E_BASE_URL=http://localhost:8093` | grün, 23/23 (3.4 min) | Suiten-Lauf 2026-07-09 |
| Voll-Verify | `make verify` | grün (5 Go-Module, Integrationstests, Frontend) | Suiten-Lauf 2026-07-09 |
| Fuzz | `make fuzz` | grün, 4 Targets je 90s, kein Crasher | Suiten-Lauf 2026-07-09 |
| Vuln-Scan Go | `govulncheck` (v1.5.0) | rot, nur Stdlib (kein Call-Pfad in eigenem Code) | Scan 2026-07-09 |
| Vuln-Scan Frontend | `pnpm audit --audit-level high` | 13 Findings (2 low/7 mod/4 high), nur Dev-/Tooling-Deps | Scan 2026-07-09 |
| TSE-Live | Live-Suite | Regelbetrieb p95=326ms; Burst-24 p95=7192ms | Messung 2026-07-09 |
| UX-Review | heuristisch, 14 Screenshots | siehe Befunde | UX-Review 2026-07-09 |
| Route-Sweep | Playwright, mobile+desktop | siehe Befunde | Sweep 2026-07-09 |

## 3. Befunde nach Schwere

### Hoch

**H1 — FAB verdeckt Aktions-Icons in Admin-Listen**
Der schwebende Button „+ Neues Produkt / + Neuer Tisch / + Neuer Benutzer" ist fixed positioniert und überlappt beim Scrollen die Edit-/Löschen-Icons bzw. den Text der darunterliegenden Karte. Betroffene Icons sind verdeckt und vermutlich nicht antippbar.
Reproduktion: Als Admin `/admin/produkte`, `/admin/tische` bzw. `/admin/benutzer` öffnen, nach unten scrollen. Tritt auf Mobile (393x851) und Desktop (1280x800) auf; konkret gesichtet an den Karten Weizen, Grillplatte, Zelt A1, Tisch 8 und Anna Krause.
Quelle: UX-Review (hoch) und Route-Sweep (dort je Route als mittel/niedrig gesplittet; hier konsolidiert und hochgestuft, weil es dasselbe systematische, aktionsblockierende Muster über alle Listen ist).
Fix-Empfehlung: Liste unten mit `padding-bottom` in FAB-Höhe (plus Safe-Area-Inset) versehen, oder den Button in eine eigene, nicht überlappende Sticky-Footerleiste verschieben.

**H2 — Bestell-/Tab-Leiste überlappt Produktliste im Tisch-Detail (mobile)**
Im Tisch-Detail überlappen die untere Leiste „Bestellung überprüfen 0,00 EUR" und die Tab-Leiste (Bestellen/Kassieren/Historie) die Produktliste; die Zeile „Kuchen" und die Folgezeile werden durchschnitten und teils verdeckt. Genau diese Einträge hinter der Leiste sind schwer zu treffen.
Reproduktion: Als Servicekraft ein Tisch-Detail auf Mobile öffnen, zum unteren Ende der Produktliste scrollen.
Quelle: UX-Review (hoch).
Fix-Empfehlung: Produktliste unten `padding-bottom` in Höhe der fixierten Leisten plus Sicherheitsabstand (Safe-Area-Inset) geben.

### Mittel

**M1 — Auswertungs-Tabs auf Mobile abgeschnitten**
Die Tab-Leiste (Übersicht/Servicekräfte/Tische/Stornierungen) ist im Live-Dashboard und in der Historischen Auswertung breiter als der Viewport: das Badge bei „Tische" wird in der Zahl abgeschnitten, der Tab „Stornierungen" ist rechts komplett unsichtbar. Kein Scroll-Indiz (kein Fade/Chevron), sodass Admins den Tab nicht finden.
Reproduktion: Als Admin `/admin/auswertung` auf Mobile (393x851) öffnen, beide Tab-Leisten prüfen.
Quelle: Route-Sweep und UX-Review (beide mittel).
Fix-Empfehlung: Tab-Leiste horizontal scrollbar mit sichtbarem Rand-Fade/Chevron, oder bei Platzmangel umbrechen bzw. als Dropdown.

**M2 — Unbekannte Route zeigt unstylte englische Fehlerseite**
`/gibtsnicht` (und generell unbekannte Routen) zeigen den rohen React-Router-Bildschirm „Unexpected Application Error! 404 Not Found" statt einer gestalteten deutschen 404-Seite. Kein Zurück-Link, wirkt für Helfer wie ein Absturz.
Reproduktion: Ausgeloggt `http://localhost:8093/gibtsnicht` aufrufen.
Quelle: Route-Sweep und UX-Review (beide mittel).
Fix-Empfehlung: Gestaltete deutsche 404-/Fehler-Route mit Button „Zurück zur Startseite" (`errorElement`); löst zugleich die HydrateFallback-Warnung (siehe N1).

**M3 — Druckauftrags-Referenz im technischen Rohformat**
In „Fehlgeschlagene Druckaufträge" wird die Referenz als roher Event-Bezeichner angezeigt, z. B. „Referenz: bestellung-aufgenommen:86". Das ist ein kebab-case-Domain-Event plus laufende Nummer und für Vereins-Admins ohne Tech-Hintergrund kryptisch. `docs/language.md` verlangt benutzer-sichtbare Strings deutsch und fachlich verständlich.
Reproduktion: Als Admin `/admin/druckstationen` öffnen, Abschnitt „Fehlgeschlagene Druckaufträge".
Quelle: Route-Sweep (niedrig) und UX-Review (mittel); hier mittel, weil es eine dokumentierte Sprachregel verletzt.
Fix-Empfehlung: Fachlichen Anzeigetext rendern (z. B. „Bestellung Nr. 86"); den technischen Bezeichner höchstens als Detail/Tooltip.

**M4 — Lösch-Bestätigung in Admin-Listen zu verifizieren**
Das rote Löschen-Icon steht direkt neben Edit und ist unmittelbar antippbar; im Screenshot-Material war kein Bestätigungsschritt sichtbar. Auf Mobile mit eng stehenden Icons (teils vom FAB überlappt, siehe H1) ist versehentliches Löschen wahrscheinlich.
Reproduktion: Admin-Listen (`/admin/produkte`, `/admin/tische`, `/admin/benutzer`), Löschen-Icon antippen.
Quelle: UX-Review (mittel, ausdrücklich „zu verifizieren").
Fix-Empfehlung: Zunächst prüfen, ob ein Bestätigungsdialog mit Objektname existiert. Falls nicht: ergänzen; Edit/Delete deutlicher trennen, Mindest-Touchgröße ~44px mit Abstand.
Verifiziert (2026-07-09): `ProductItem.tsx`, `TischItem.tsx` und `UserItem.tsx` nutzen alle einen `AlertDialog` mit Objektname („Produkt löschen?"). Der Bestätigungsschritt existiert; von M4 bleibt nur der Touch-Abstand, der mit dem H1-Fix zusammenfällt.

**M5 — Seed-Integrationstests kontaminieren die DB über Läufe**
`TestResetAndSeed` leert `tse_konfiguration` ohne Default-Row; ein Re-Run von `TestSeedRun_Erstlauf` ist danach betroffen. Vorbestehend, nicht durch diese QA verursacht.
Reproduktion: Seed-Integrationstests zweimal hintereinander laufen lassen.
Quelle: Bekannter Befund aus den Implementierungs-Wellen.
Fix-Empfehlung: Reset auf einen deterministischen Ausgangszustand inkl. Default-Row bringen, oder betroffene Tests isolieren/idempotent machen.

### Niedrig

**N1 — HydrateFallback-Konsolenwarnung**
React-Router-Warnung „No `HydrateFallback` element provided to render during initial hydration" beim Erstladen auf allen geprüften Routen und Viewports. Kein sichtbarer Effekt; der Redirect ausgeloggter Routen (`/admin`, `/service`) auf `/login` funktioniert korrekt. Nur Konsolen-Rauschen.
Quelle: Route-Sweep, Konsole-Fehler.
Fix-Empfehlung: Mit M2 zusammen über eine gemeinsame `errorElement`/`HydrateFallback`-Route lösen.

**N2 — Kontrast bei Sonnenlicht / Primärbutton wirkt deaktiviert**
Sekundärtexte und der mintgrüne Primärbutton („Anmelden") haben niedrigen Kontrast; der aktive Button wirkt wie deaktiviert. Auf BYOD-Smartphones im Freien schlecht lesbar.
Quelle: UX-Review (niedrig).
Fix-Empfehlung: Kontraste gegen WCAG AA (>=4.5:1) prüfen, aktiven Primär-CTA kräftiger einfärben, Sekundärtext dunkler setzen.

**N3 — Login-Karte auf Mobile tief platziert**
Die Login-Karte sitzt vertikal zentriert relativ tief mit großem Leerraum oben; kostet bei Einhandbedienung Daumenweg. Der Falschpasswort-Zustand ist sauber (klare rote deutsche Meldung, Passwort-Toggle vorhanden).
Quelle: UX-Review (niedrig).
Fix-Empfehlung: Login-Karte auf Mobile höher/kompakter platzieren.

**N4 — golangci-lint mit `--build-tags=integration` meldet Altlasten**
Rund 24 Issues (errcheck/gosec/errorlint) in älteren `*_integration_test.go`, u. a. `db.Close` in `idempotenz_integration_test.go:48`. Im normalen Lint ohne integration-Tag unsichtbar. Vorbestehend.
Quelle: Bekannter Befund aus den Implementierungs-Wellen.
Fix-Empfehlung: Integration-Test-Lint separat sauberziehen oder bewusst per Linter-Konfiguration ausklammern.

**N5 — Begriffs-Drift: DSFinV-K-Export-Button-Ort**
Der DSFinV-K-Export-Button liegt auf `/admin/auswertung` (Dashboard), Plan/Doku sprachen von einer Finanzamt-Ansicht (`/admin/finanzamt`). Reine Doku-/Begriffs-Drift, keine Fehlfunktion.
Quelle: Bekannter Befund aus den Implementierungs-Wellen.
Fix-Empfehlung: Doku an den Ist-Ort angleichen oder Route/Bezeichnung vereinheitlichen.

**N6 — E2E-Test-Robustheit (Minor)**
Die Einmalpasswort-Assertion prüft nur nicht-leer statt 6-stelliges OTP-Format; einige Selektoren sind an Tailwind-Klassen gekoppelt (`p.text-3xl.tracking-widest`, `div.space-y-4`). `zeileMit()` nutzt einen breiten `div`-`hasText`-Filter mit `.last()`, und `settleAlleOffenenTische` koppelt die Kassenabschluss-Spec eng ans Seed-Drehbuch (`test.setTimeout` 120s).
Quelle: Bekannter Befund aus den Implementierungs-Wellen.
Fix-Empfehlung: OTP-Format-Assertion schärfen, Selektoren auf stabile Testattribute umstellen.

**N7 — Fuzz-Assertions oberflächlich (Minor)**
`assertQRCommandLengths` kann durch ein `GS(k`-Präfix in der QR-Payload fehlgeleitet werden; `FuzzApplyEvent` assertet nur Panic-Freiheit statt semantischer Invarianten.
Quelle: Bekannter Befund aus den Implementierungs-Wellen.
Fix-Empfehlung: QR-Längenprüfung robuster gegen Payload-Kollisionen, Fuzz-Ziele um semantische Checks ergänzen.

## 4. Offene Punkte mit Human-Gate

- **Sicherheits-Scans (entschieden 2026-07-09):** govulncheck ist in allen 5 Go-Modulen rot, ausschließlich wegen Go-Stdlib-CVEs (u. a. GO-2026-5856/5039/5037/4971/4918; reverse-proxy zusätzlich GO-2026-4982/4980 in html/template). Kein Modul hat einen projekteigenen Nicht-Stdlib-Fund mit Call-Pfad; Fix ist ein Toolchain-Bump auf go1.26.3+ (bzw. .4/.5 je CVE), kein Code-Fix. Frontend `pnpm audit` meldet 4 high (hono via shadcn-devDep, undici via jsdom/vitest), alle Dev-/Tooling-Deps, kein Produktionscode. Entscheidung: Bump auf go1.26.5. Lokal ist go1.26.5 bereits installiert; noch offen sind die 5 `go.mod` (stehen auf 1.26.0), die Workflow-Pins (`go-version: 1.26.0` in ci/security-scans/release) und die Dockerfiles (`golang:1.26.4-alpine`). Geht als Phase „Toolchain-Konsistenz" in den Fix-Plan.
- **TSE-Live-Burst (abgenommen 2026-07-09):** Regelbetrieb bestätigt die Zusage (p95=326ms < 5s). Burst mit 24 gleichzeitigen Signaturen erreicht p95=7192ms durch serielles Worker-Queueing (in der Verfahrensdoku dokumentiert). Das Nachsigniert-Kennzeichen (>1min) wurde im Ausfalltest nicht erzwungen, weil der Worker-Backoff unter der Schwelle blieb. Nico akzeptiert das Burst-Verhalten wie dokumentiert; kein Fix-Bedarf, keine Optimierung geplant.
- **Ops-Smoke:** `scripts/ops-smoke.sh` in allen drei Modi auf einem Wegwerf-Host; braucht einen von Nico gestellten Ubuntu-Host. Ebenso destruktives prod-restore, TLS-Abnahme sowie die Hardware-, Windows-, Zwei-Geräte- und Usability-Punkte laut Rest-Guide. Entschieden: bleibt im Rest-Guide, kommt nicht in den Fix-Plan.
- **CI-Erstlauf (erledigt 2026-07-09):** main ist gepusht, die neuen Jobs sind remote gelaufen. Workflow „CI" grün (6m14s, inkl. e2e), Workflow „Security Scans" rot: govulncheck meldet die o. g. Stdlib-CVEs, weil der Workflow `go-version: 1.26.0` pinnt. Wird mit der Phase „Toolchain-Konsistenz" grün.

## 5. Entscheidungen aus der Klärung (2026-07-09)

- Scope des Fix-Durchgangs: alle Befunde (H1–H2, M1–M5, N1–N7). M4 ist per Code-Check im Kern erledigt (siehe oben), N1 fällt mit M2 ab.
- Vorgehen: nur Plan-Datei erstellen; Umsetzung in einer späteren Session.
- M1: Tab-Leiste horizontal scrollbar mit sichtbarem Rand-Fade und Chevron.
- N2: systematischer Token-Audit gegen WCAG AA (UI bleibt grün), Prüfung rein automatisiert (berechnete Kontrast-Ratios), einmalig im Fix-Durchgang, kein dauerhafter Regressions-Guard, keine manuelle Abnahme.
- N5: Doku an den Ist-Ort angleichen (Export im Auswertungs-Dashboard), keine eigene Route `/admin/finanzamt`.
- Toolchain: Bump auf go1.26.5 in go.mod, Workflows und Dockerfiles vervollständigen; Ziel ist ein grüner Security-Scans-Lauf.
- TSE-Burst: abgenommen wie dokumentiert.
- Human-Gate-Punkte (Ops-Smoke, prod-restore, TLS, Hardware/Windows/Zwei-Geräte): nicht Teil des Fix-Plans, bleiben im Rest-Guide.
