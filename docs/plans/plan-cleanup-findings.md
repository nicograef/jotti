# Plan: Cleanup-Befunde aus dem Multi-Experten-Review

> Source PRD: n/a (Befundliste aus dem repo-weiten Cleanup-Review vom 2026-07-16, 48 verifizierte Befunde)

## Goal

Alle bestätigten Befunde des repo-weiten Cleanup-Reviews beheben: veraltete Doku und Agenten-Instruktionen korrigieren, totes und dupliziertes Frontend/Backend/Website-Code entfernen, Cross-Layer-Verträge verschlanken und zwei Struktur-Refactorings umsetzen. Kein Verhalten ändert sich; einzige bewusste Ausnahmen sind zwei sichtbare Textkorrekturen (Rollen-Label im Helfer-Dialog, Preis-Validierungstext), die falsche Fachaussagen richtigstellen.

## Architectural decisions

Durable Entscheidungen, die für alle Phasen gelten:

- **`db.WithTx`**: Die vier identischen `withTx`-Methoden werden durch eine gemeinsame Funktion im Paket `backend/db` ersetzt (Begin/Rollback/Commit um einen `func(*dbgen.Queries) error`-Callback). Die Repos rufen `db.WithTx` direkt auf; die per-Repo-Methoden entfallen.
- **Read-Model-Typen gehören der Domain**: `SignaturQueueZustand` und `Stoerungszeitraum` ziehen nach `backend/domain/tse`, `FehlgeschlagenerDruckauftrag` nach `backend/domain/druckstation`. Repos liefern Domain-Typen; HTTP/Application importieren keine Repository-Structs mehr (Präzedenz: `reporting.ReportingData`).
- **Ein Zähler statt zwei**: `AnzahlOffen` bleibt (passt zu `OffenCents` und zur ADR-01-Semantik „offen = unbezahlt"), `AnzahlUnbezahlt` entfällt in `kasse`, `reporting` und dem Mapping. Kein HTTP-Response serialisiert einen der beiden Zähler (verifiziert), der Umbau ist rein intern.
- **Gemeinsames Status-Schema im Frontend**: `EntityStatusSchema` (+ Typ `EntityStatus`) einmalig in `frontend/src/lib/entityStatus.ts`, gespiegelt am Backend-`produkt.Status` / DB-Enum `entitystatus`. Ersetzt `VarianteStatusSchema` (admin) und `ProduktStatusSchema` (service).
- **`formatEuro` ist die einzige Geldbetrag-Darstellung mit Euro-Suffix**: Alle `formatCents(x)` + `€`-Inline-Stellen werden ersetzt; `formatCents` bleibt nur für suffixlose Verwendungen.
- **E2E-Helfer-Heimat**: Der Overflow-Helfer zieht nach `e2e/support/viewport.ts` (Repo-Konvention ist `support/`, nicht `helpers/`). `launchBrowser` zieht in ein neues Modul `e2e/website/browser.mjs`, weil `csp-server.mjs` bewusst dependency-frei bleibt.
- **Eyebrow als CSS-Klasse**: Die bestehende `.eyebrow`-Klasse in `website/src/styles/landing.css` wird auf die Werte der Inline-Variante korrigiert (Spektral-Balken als `::before`) und ersetzt alle Inline-Kopien.
- **Werkzeug-Pinning**: `setup-dev-tools.sh` pinnt goimports auf die in `ci.yml` verwendete Version (aktuell v0.40.0) und sqlc auf die Version, die das eingecheckte `dbgen/` unverändert reproduziert.

## Inventory

- `backend/domain/kasse/offene_arbeit.go — EigeneArbeitAnTisch, ComputeEigeneArbeitAnTisch, OffeneArbeitTisch` und `backend/domain/reporting/reporting.go — OffeneArbeitTisch`, Mapping in `backend/api/reporting/application/query.go`
- `backend/repository/{kassenjournal_repo,druckauftrag_repo,user_repo}/repo.go` und `backend/repository/tse_repo/einrichtung.go — withTx` (vier identische Kopien); `backend/db/db.go — Error, ResultError` (Ziel-Paket, dort auch der ungenutzte `Close`-Helper)
- `backend/api/fiskal/signatur/http/handler.go` und `backend/api/druck/auftrag/http/handler.go` (Repository-Typen an der HTTP-Grenze)
- `backend/api/druck/beleg/application/kassenbeleg_command.go — KassenbelegDrucken` (~225 Zeilen, vier Quell-Auflösungszweige)
- `backend/api/kasse/kassenfuehrung/http/command_handler.go — signaturenAusstehendDetails` und `frontend/src/admin/kasse/KasseBackend.ts — SignaturenAusstehendDetailsSchema` (ungenutztes `alterSekunden`)
- `backend/api/kasse/tischgeschaeft/application/query.go — tischState-Read-Model, GetTischHistorie` und `backend/api/kasse/tischgeschaeft/http/query_handler.go` (DTO mit `gesamtZahlungenCents`); Frontend-Gegenstücke `frontend/src/service/table/Tisch.ts — TischSessionSchema`, `frontend/src/service/table/hooks.ts`
- `frontend/src/lib/utils.ts — formatEuro, formatCents`; 16 Service-Dateien mit Inline-Suffix (u. a. `Receipt.tsx`, `TablePage.tsx`, `Zahlung.tsx`)
- `frontend/src/components/common/FormFields.tsx — NameField, RoleField, LockedField, DescriptionField, CategoryField, SteuersatzField` (hartkodierte `form-*`-IDs; `UsernameField`/`EuroField` nutzen bereits dynamische IDs), `PriceField` (JSDoc „net price")
- `frontend/src/admin/users/UserRolle.tsx — rolleLabel` (bisher nicht exportiert), `frontend/src/admin/users/UserCreatedDialog.tsx` (rohe Enum-Interpolation)
- `website/src/styles/landing.css — .eyebrow, .btn-outline, .btn-lg`, `website/src/styles/brand.css — --brand-font`; Inline-Eyebrow in `Features`, `FuerWen`, `Sicherheit`, `Preis`, `Ablauf`, `Download`, `Faq`, `Screenshots`, `fuer-vereine`, `LiveDemo`
- `e2e/tests/tischservice-viewport-ueberlauf.mobile.spec.ts` und `e2e/tests/admin-finanzamt-einrichtung.spec.ts — erwarteKeinenHorizontalenUeberlauf`; `e2e/website/{csp-check.mjs,screenshots.mjs} — launchBrowser`
- `scripts/prod-init.sh`, `scripts/prod-update.sh` (Versions-Kommentar), `scripts/setup-dev-tools.sh`, `backend/Dockerfile`
- Verifikation: `make check`, `make verify`, `make test-e2e`, `make website-screenshots` (Screenshot-Parität)

## Resolved decisions

- Dockerfile-Änderung (`RUN apk update` löschen) von Nico freigegeben (2026-07-16).
- Read-Model-Umzug in die Domain-Pakete von Nico freigegeben (2026-07-16); das Muster „Repository-Struct an der HTTP-Grenze" gilt damit als Drift, nicht als Konvention.
- `TischAuswahlDrawer`: nur das redundante Disjunkt im `disabled`-Check entfernen (`favoritLoading` umschließt `favoritMutation.isPending` nachweislich); beide Abstraktionen bleiben, da sie getrennte Aufgaben haben (Invalidierung bzw. Fehler-Toast).
- `gesamtZahlungenCents` wird Ende-zu-Ende entfernt (Application-Read-Model, HTTP-DTO, Zod-Schema, Hook-Default, alle TS-Fixtures). Das Domain-Projektionsfeld `TischSession.GesamtZahlungenCents` bleibt (persistierte Projektion, Replay-Drift-Integrationstest).
- Preis-Validierungstext wird zu „Der Preis darf nicht negativ sein." (deckt `min(0)` korrekt ab und entspricht der Backend-zog-Meldung); „Nettopreis" verschwindet, weil `preisCents` der Bruttopreis ist.
- Die 12 im Review widerlegten Befunde bleiben unbearbeitet; ebenso der `HinweisKarte.title`-Befund (bewusste API-Symmetrie zu `WarnKarte`).
- Kommentar-only-Befunde (Relay-Druckerstatus, Phase-Labels, Manifest-Link, Prod-Skripte) werden als Kommentartext-Korrekturen umgesetzt, kein Codepfad ändert sich.

## Open questions / Risks

- `formatEuro`-Konsolidierung: Nur Stellen ersetzen, an denen auf `formatCents(...)` ein Euro-Suffix folgt (auch `&nbsp;€`); suffixlose Verwendungen behalten `formatCents`.
- Eyebrow-Konsolidierung: visuelle Parität per Screenshot-Vergleich sichern (`make website-screenshots`, md5-Vergleich vorher/nachher); die korrigierten `.eyebrow`-Werte müssen exakt der bisherigen Inline-Variante entsprechen, nicht den bisherigen (veralteten) Klassenwerten.
- `KassenbelegDrucken`-Extraktion ist das größte Einzelstück; bestehende Beleg-Tests und `make verify` sind das Sicherheitsnetz, es wird ausschließlich Code verschoben, nicht umgeschrieben.

---

## Phase 1: Doku und Instruktionen aktualisieren

### Context

- `.github/instructions/event-sourcing.instructions.md` — dokumentiert das per ADR 01 entfernte Event `ausgabe-bestaetigt` als lebendig
- `.github/instructions/frontend.instructions.md` — nennt den entfernten „Liefern"-Drawer und ein falsches Farbschema
- `docs/produktbeschreibung.md`, `docs/handbuch.md`, `docs/anforderungen.md`, `docs/language.md`, `database/migrations/README.md` — faktisch veraltete Stellen
- `docs/plans/` — zwei vollständig abgehakte Pläne (Konvention: nach Merge löschen)

### What to build

Alle dokumentierten Falschaussagen korrigieren, sodass Instruktions- und Referenzdoku wieder mit dem Code übereinstimmen, und die erledigten Pläne löschen. Im Einzelnen: die drei `ausgabe-bestaetigt`-Reste (Event-Liste, `AusstehendePositionen`-Projektionszeile, `AusgabeBestaetigtV1Data`-Beispiel) entfernen; die Drawer-Liste auf den ADR-03-Kanon (Bestellen, Kassieren, Stornieren, Umbuchen, Direktverkauf) und die Palette auf olive/emerald/zinc korrigieren; die zwei „(in Entwicklung)"-Marker in der Produktbeschreibung entfernen; in Handbuch und Anforderungen (jeweils Q-06) Caddy als Produktions-Reverse-Proxy nennen (nginx nur im jotti.rocks-Demo-Stack); den Watchdog-Pfad in `language.md` auf `backend/api/fiskal/signatur/tse_rueckstand_watchdog.go` korrigieren; in `database/migrations/README.md` den konkreten „zuletzt"-Dateinamen aus Regel 1 streichen, damit die Regel nicht bei jeder Migration veraltet; `docs/plans/plan-ui-refinements.md` und `docs/plans/plan-service-split-screen.md` löschen.

### Acceptance criteria

- [x] `grep -r "ausgabe-bestaetigt\|AusgabeBestaetigt" .github/instructions/` liefert keine Treffer mehr
- [x] `.github/instructions/frontend.instructions.md` nennt weder „Liefern" noch „Violet/Indigo"
- [x] `grep -rn "in Entwicklung" docs/produktbeschreibung.md` leer; Q-06 in `docs/handbuch.md` und `docs/anforderungen.md` nennt Caddy; der Watchdog-Pfad in `docs/language.md` existiert im Repo
- [x] `database/migrations/README.md` referenziert keinen konkreten „zuletzt"-Migrationsnamen mehr
- [x] In `docs/plans/` liegen nur noch Dateien mit offenen Checkboxen bzw. der QA-Guide

---

## Phase 2: Backend-Schnellkorrekturen

### Context

- `backend/api/kasse/tischgeschaeft/application/query.go — GetTischHistorie()` — einziges Fehler-Log der Datei ohne `.Err(err)`
- `backend/api/service.go — NewServiceApi()` — einzeiliges `tq.Query`-Struct-Literal zwischen mehrzeiligen Geschwistern
- `backend/api/auth/http/command_handler.go — LoginHandler()` — `||`-Form statt Komma-Cases wie im `SetPasswordHandler` derselben Datei
- `backend/config/config.go — parseEnvString()` — redundantes else nach return
- `backend/main.go — main()` — doppelter `len(os.Args) > 1 && os.Args[1] == …`-Guard
- `backend/db/db.go — Close()` — exportierter Helper ohne einen einzigen Aufrufer
- `backend/api/reporting/http/query_handler.go — liveSummaryResponse` — byte-identisch zu `summaryResponse`

### What to build

Sieben kleine, verhaltensneutrale Go-Korrekturen: `.Err(err)` im Historie-Log ergänzen; das `tq.Query`-Literal mehrzeilig formatieren; den Zwei-Fehler-Case im `LoginHandler` auf Komma-Cases vereinheitlichen; das else in `parseEnvString` in zwei Guards auflösen; das Subkommando in `main()` einmal in eine lokale Variable lesen; den toten `Close`-Helper löschen; `liveSummaryResponse` löschen und `summaryResponse` als Feldtyp in `liveReportingResponse` verwenden (JSON bleibt identisch, analog zum bereits geteilten `stornierungServicekraft`).

### Acceptance criteria

- [x] `make check` grün
- [x] `grep -rn "liveSummaryResponse\|func Close" backend/` liefert keine Treffer mehr
- [x] Kein JSON-Feld eines Endpunkts hat sich geändert (Reporting-Handler-Tests unverändert grün)

---

## Phase 3: Backend-Struktur

### Context

- `backend/repository/{kassenjournal_repo,druckauftrag_repo,user_repo}/repo.go`, `backend/repository/tse_repo/einrichtung.go — withTx()` — vier identische Kopien
- `backend/db/db.go` — Ziel-Paket für `WithTx` (hostet bereits `Error`/`ResultError`)
- `backend/api/fiskal/signatur/http/handler.go`, `backend/api/druck/auftrag/http/handler.go` — Repository-Structs als Boundary-Typen
- `backend/domain/tse`, `backend/domain/druckstation` — Zielpakete für die Read-Model-Typen
- `backend/domain/kasse/offene_arbeit.go`, `backend/domain/reporting/reporting.go`, `backend/api/reporting/application/query.go` — Zähler-Kollaps
- `backend/api/druck/beleg/application/kassenbeleg_command.go — KassenbelegDrucken()` — God Function

### What to build

Vier Struktur-Refactorings ohne Verhaltensänderung. Erstens `db.WithTx(ctx, db, fn)` im `db`-Paket einführen und die vier `withTx`-Kopien darauf zurückbauen. Zweitens die Read-Model-Structs in ihre Domain-Pakete verschieben (`tse.SignaturQueueZustand`, `tse.Stoerungszeitraum`, `druckstation.FehlgeschlagenerDruckauftrag`); Repos geben Domain-Typen zurück, Application- und HTTP-Schicht importieren die Repository-Pakete an diesen Stellen nicht mehr. Drittens `AnzahlUnbezahlt` ersatzlos streichen: `EigeneArbeitAnTisch`, `kasse.OffeneArbeitTisch`, `reporting.OffeneArbeitTisch` und das Mapping behalten nur `AnzahlOffen`; `ComputeEigeneArbeitAnTisch` zählt direkt (die ID-Deduplikation über eine Map entfällt, da `UnbezahltePositionen` je `PositionID` höchstens einen Eintrag trägt), `Erledigt` leitet sich aus demselben Zähler ab. Viertens `KassenbelegDrucken` entflechten: je Beleg-Form eine `resolve…`-Funktion, die ein kleines `belegQuelle`-Struct (Quell-Event, Positionen, Gesamtbetrag, Referenz, Storno-Felder) liefert; die Hauptfunktion wird zur Kette resolve, TSE, Formatierung, Enqueue. Reines Verschieben, keine Logikänderung.

### Acceptance criteria

- [x] `make verify` grün (inkl. Integrationstests)
- [x] `grep -rn "func (r Repository) withTx" backend/repository/` leer; genau eine `WithTx`-Definition im `db`-Paket
- [x] Kein Import von `repository/…`-Paketen mehr in `backend/api/fiskal/signatur/http`, `backend/api/fiskal/signatur/application`, `backend/api/druck/auftrag/http`, `backend/api/druck/auftrag/application`
- [x] `grep -rn "AnzahlUnbezahlt" backend/` leer
- [x] `KassenbelegDrucken` enthält keine Quell-Auflösungslogik mehr, nur noch den Ablauf resolve, TSE, Formatierung, Enqueue; Beleg-Tests unverändert grün

---

## Phase 4: Cross-Layer-Verträge verschlanken

### Context

- `backend/api/kasse/kassenfuehrung/http/command_handler.go — signaturenAusstehendDetails` und `frontend/src/admin/kasse/KasseBackend.ts — SignaturenAusstehendDetailsSchema` — `alterSekunden` wird berechnet, übertragen, geparst und nie gelesen
- `backend/api/kasse/tischgeschaeft/application/query.go — tischState`, `backend/api/kasse/tischgeschaeft/http/query_handler.go` — `gesamtZahlungenCents` im Read-Model/DTO ohne UI-Konsument
- `frontend/src/service/table/Tisch.ts — TischSessionSchema`, `frontend/src/service/table/hooks.ts`, `frontend/src/service/TableSelectionPage.test.tsx` — Frontend-Gegenstücke und Fixtures
- `frontend/src/components/common/FormFields.tsx — PriceField()` und `frontend/src/service/product/Produkt.ts — PreisCentsSchema` — falsche Netto-Bezeichnung für den Bruttopreis
- `frontend/src/admin/products/Produkt.ts — VarianteStatusSchema` und `frontend/src/service/product/Produkt.ts — ProduktStatusSchema` — zwei Namen für ein Backend-Konzept

### What to build

Die vier von den Flow-Tracern gefundenen Vertragsbefunde beheben, Backend und Frontend im selben Schnitt. `alterSekunden` aus dem 409-Detail-Struct samt Berechnung und aus dem Zod-Schema entfernen (die UI liest nur `anzahl`). `gesamtZahlungenCents` aus dem Application-Read-Model, dem HTTP-DTO, dem `TischSessionSchema`, dem Hook-Default und sämtlichen TS-Fixtures entfernen; das Domain-Projektionsfeld bleibt unangetastet. Die Preis-Texte korrigieren: JSDoc von `PriceField` auf „gross price (Brutto)", Validierungsmeldung auf „Der Preis darf nicht negativ sein.". `EntityStatusSchema`/`EntityStatus` in `frontend/src/lib/entityStatus.ts` anlegen und in beiden Produkt-Schemas (admin: Produkt und Variante, service: Produkt) verwenden; die beiden alten Schema-Namen entfallen.

### Acceptance criteria

- [x] `make check` grün; `make test-e2e` grün
- [x] `grep -rn "alterSekunden\|gesamtZahlungenCents" backend/api frontend/src e2e/` leer; das Read-Model `tischState` und das HTTP-DTO tragen kein `GesamtZahlungenCents`-Feld mehr (das Domain-Feld in `backend/domain/kasse` und seine Projektions-/Integrationstests bleiben)
- [x] `grep -rn "Nettopreis\|net price" frontend/src` leer
- [x] `grep -rn "VarianteStatusSchema\|ProduktStatusSchema" frontend/src` leer; genau eine `EntityStatusSchema`-Definition

---

## Phase 5: Frontend-Aufräumarbeiten

### Context

- `frontend/src/service/components/direktverkauf/{DirektverkaufAbschluss,DirektverkaufDrawer}.tsx` — lokale Re-Deklarationen von `VerkaufPositionInput`
- `frontend/src/service/direktverkauf/Direktverkauf.ts` und `frontend/src/service/table/Bestellung.ts — PositionRefSchema` — doppelt deklariert; Ziel: `frontend/src/service/schemas.ts`
- `frontend/src/service/DirektverkaufPage.tsx` und `frontend/src/service/TablePage.tsx — dockFreiraum` — duplizierte Dock-Höhen-Konstante; Ziel: Export neben `frontend/src/service/components/ServiceDock.tsx`
- `frontend/src/admin/kasse/Kassensitzung.ts — kassensitzungStatusLabel` — nur vom eigenen Test referenziert
- `frontend/src/admin/kasse/LaufenderBetriebSection.tsx — formatBewegungZeit()` — dupliziert `formatLocalTime` aus `frontend/src/admin/reporting/utils.ts`
- `frontend/src/admin/users/UserCreatedDialog.tsx` — rohe Rollen-Enum-Interpolation; `frontend/src/admin/users/UserRolle.tsx — rolleLabel`
- `frontend/src/lib/errorMessages.ts — commonErrorMessages` — zwei unerreichbare Einträge
- `frontend/src/components/theme-provider.tsx` — immer leerer `...props`-Spread
- `frontend/src/service/components/TischAuswahlDrawer.tsx` — redundantes Disjunkt im `disabled`-Check
- `frontend/src/components/common/FormFields.tsx` — sechs Felder mit hartkodierten `form-*`-IDs
- `frontend/src/lib/utils.ts — formatEuro()` — Inline-`€`-Suffix in ~16 Service-Dateien

### What to build

Die Frontend-Befunde in einem Schnitt: die lokalen `VerkaufPositionInput`-Interfaces löschen und den exportierten Typ importieren; `PositionRefSchema` einmal in `service/schemas.ts` definieren und in beiden Modulen importieren; die `dockFreiraum`-Konstante neben `ServiceDock` exportieren und in beiden Seiten importieren; `kassensitzungStatusLabel` samt Testblock löschen; `formatBewegungZeit` durch `formatLocalTime` ersetzen; im `UserCreatedDialog` `rolleLabel` (aus `UserRolle.tsx` exportieren) statt des rohen Enums rendern; die zwei unerreichbaren `commonErrorMessages`-Einträge (`internal_server_error`, `unknown`) löschen; den leeren `...props`-Spread im Theme-Provider entfernen; im `TischAuswahlDrawer` das redundante `favoritMutation.isPending ||` im `disabled`-Check streichen. Dazu die zwei größeren Posten: die sechs `form-*`-Felder auf `useId()` umstellen (ID jeweils für `FieldLabel htmlFor` und Input/Trigger wiederverwenden, wie `UsernameField`); alle Inline-`formatCents(x)` + `€`-Stellen der Service-Area auf `formatEuro(x)` konsolidieren.

### Acceptance criteria

- [x] `make check` grün (Frontend-Lint, -Tests, -Typecheck)
- [x] `grep -rn "interface VerkaufPositionInput" frontend/src` leer; genau eine `PositionRefSchema`- und eine `dockFreiraum`-Definition
- [x] `grep -rn "kassensitzungStatusLabel\|formatBewegungZeit" frontend/src` leer; der Dialog rendert „Service-Helfer" statt „service-Helfer"
- [x] `grep -rn 'id="form-' frontend/src/components/common/FormFields.tsx` leer
- [x] In `frontend/src/service/` folgt auf keinen `formatCents(...)`-Aufruf mehr ein Euro-Zeichen (inkl. `&nbsp;€`)

---

## Phase 6: Website

### Context

- `website/src/styles/landing.css — .eyebrow, .btn-outline, .btn-lg` — Eyebrow-Klasse veraltet und ungenutzt, Button-Varianten tot
- `website/src/styles/brand.css — --brand-font` — Alias ohne Konsument
- Inline-Eyebrow-Kopien in `Features`, `FuerWen`, `Sicherheit`, `Preis`, `Ablauf`, `Download`, `Faq`, `Screenshots`, `fuer-vereine`, `LiveDemo` (Astro und React)

### What to build

Die `.eyebrow`-Klasse auf die Werte der Inline-Variante korrigieren (Schriftgröße, Tracking, Farbe; Spektral-Balken als `::before`) und alle zehn Inline-Kopien durch `<p class="eyebrow">…</p>` ersetzen, sodass das Eyebrow eine einzige Quelle hat. Die ungenutzten Klassen `.btn-outline` und `.btn-lg` sowie den toten Alias `--brand-font` löschen.

### Acceptance criteria

- [x] Website-Build grün; `grep -rn "btn-outline\|btn-lg\|--brand-font" website/src` leer
- [x] Der Eyebrow-Markup-Block (Spektral-Balken-Span + Label) existiert in keiner Komponente mehr inline
- [x] Screenshot-Parität: `make website-screenshots` vor und nach der Änderung erzeugt pixel-identische Bilder (md5-Vergleich) für die betroffenen Sektionen

---

## Phase 7: E2E und Windows

### Context

- `e2e/tests/tischservice-viewport-ueberlauf.mobile.spec.ts` und `e2e/tests/admin-finanzamt-einrichtung.spec.ts — erwarteKeinenHorizontalenUeberlauf()` — byte-identische Kopien
- `e2e/tests/admin-finanzamt-einrichtung.spec.ts` und `e2e/tests/admin-kontrast-axe.spec.ts` — lokale Variable `zugang` statt suite-weitem `zugangsdaten`
- `e2e/website/csp-check.mjs` und `e2e/website/screenshots.mjs — launchBrowser()` — dupliziert inkl. hartkodiertem Chromium-Fallback-Pfad
- `windows/relay/main.go — checkPrinter()` — undokumentiertes Verschlucken des Status-Timeouts
- `windows/starter/jotti-start.manifest` — Link auf gelöschten Plan
- `windows/starter/main.go`, `windows/starter/core/{diagnose,adminmarker,backup}.go`, `windows/starter/backup.go` — „Phase 2/3"-Labels aus gelöschten Plänen

### What to build

Test- und Windows-Hygiene: den Overflow-Helfer nach `e2e/support/viewport.ts` extrahieren und in beiden Specs importieren; die zwei `zugang`-Locals in `zugangsdaten` umbenennen; `launchBrowser` nach `e2e/website/browser.mjs` extrahieren und in beiden Skripten importieren. In den Windows-Modulen nur Kommentare: am `return nil` in `checkPrinter` erklären, dass nicht jeder Drucker die DLE-EOT-Statusabfrage beantwortet und eine fehlende Antwort als erreichbar-und-OK gilt; den toten Plan-Link im Manifest streichen; die sechs „Phase 2/3"-Labels durch die inhaltliche Beschreibung ersetzen (z. B. „einer Version vor Einführung des Pre-Update-Backups").

### Acceptance criteria

- [x] `make test-e2e` grün; `erwarteKeinenHorizontalenUeberlauf` und `launchBrowser` existieren je genau einmal
- [x] `grep -rn "const zugang " e2e/tests/` leer
- [x] `go build ./...` in `windows/starter` und `windows/relay` grün; `grep -rn "Phase 2\|Phase 3\|Phase-3\|plan-windows-verpackung" windows/` leer

---

## Phase 8: Infra und Ops

### Context

- `scripts/prod-init.sh` und `scripts/prod-update.sh` — Kommentar behauptet einen `${JOTTI_VERSION:-latest}`-Fallback, den `docker-compose.prod.yml` nicht hat
- `scripts/setup-dev-tools.sh` — goimports `@latest` trotz CI-Pin v0.40.0; sqlc `@latest`
- `backend/Dockerfile` — `RUN apk update` ohne nachfolgendes `apk add` (Löschung von Nico freigegeben)

### What to build

Die beiden Versions-Kommentare auf das reale Verhalten korrigieren (bare `${JOTTI_VERSION}`, kein Default, Compose bricht bei leerem Wert ab). In `setup-dev-tools.sh` goimports über eine Versions-Konstante auf die ci.yml-Version pinnen (analog zu golangci-lint/migrate) und sqlc auf die Version pinnen, die das eingecheckte `dbgen/` unverändert reproduziert (nach dem Pin `make sqlc` ausführen und leeren Diff verifizieren). Die `RUN apk update`-Zeile im Backend-Dockerfile löschen.

### Acceptance criteria

- [ ] `bash -n` für alle drei Skripte fehlerfrei; kein Skript-Kommentar erwähnt mehr einen `latest`-Fallback
- [ ] goimports- und sqlc-Version in `setup-dev-tools.sh` sind gepinnt; goimports-Version identisch mit `ci.yml`
- [ ] `make sqlc` erzeugt nach dem Pin einen leeren `git diff`
- [ ] Backend-Docker-Image baut erfolgreich (lokal oder über den bestehenden CI-Build)
