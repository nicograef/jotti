# Plan: Praxis-Feedback der Vereine umsetzen

> Source PRD: n/a (Praxis-Feedback aus E-Mails von Vereinen, April bis September 2026;
> die Auswertung selbst liegt bewusst außerhalb des Repos)

## Goal

Die Rückmeldungen aus den ersten produktiven Einsätzen vollständig in Doku, Website und
Software einarbeiten und danach v1.0.0 veröffentlichen: bestätigte Bedienprobleme beheben,
Lücken in den Anleitungen schließen, Erwartungen an den Nutzungsprozess auf der Website
richtigstellen, die zugrunde liegenden Anforderungen hinter den Wünschen mit eigenen
Lösungen abdecken und festhalten, was jotti bewusst nicht tut.

## Architectural decisions

- **Kategorien bleiben ein festes Enum** (`essen`, `getraenk`, `sonstiges`). Sie steuern die
  Druckstationen-Zuordnung; freie Kategorien würden Druck-Konfiguration und DSFinV-K nach
  sich ziehen.
- **Bonmodus** bekommt den dritten Wert `pro_stueck`, zulässig nur für die Druckstation
  `abholbon`. Arbeitsbons der Produktstationen kennen ihn nicht.
- **Schema-Änderungen nur additiv** als neue Migration (Freeze-Disziplin); die
  CHECK-Bedingung auf `druckstationen.bonmodus` wird in einer neuen Migration ersetzt,
  `01_initial.up.sql` bleibt unverändert.
- **TSE bleibt Cloud-TSE (fiskaly)**, bis die ADR aus Phase 8 etwas anderes entscheidet.
- **Belegausgabe bleibt Papier** über die Druckstation `kassenbeleg`. Die Befreiung nach
  § 146a Abs. 2 Satz 2 AO erspart nur das ungefragte Aushändigen, nicht den Drucker;
  ein elektronischer Beleg ist Nicht-Ziel.
- **Leitfaden ist eine Quelle, Landing-Copy nicht**: die in
  `website/src/lib/published-docs.ts` gelisteten Seiten werden aus `docs/` gelesen
  (`website/src/content.config.ts`), Änderungen passieren nur in `docs/`. Die
  Landing-Komponenten (`FaqAccordion.tsx`) sind handgeschriebene Copy und werden bei
  Faktenänderungen im selben Schritt nachgezogen.
- **Abhängigkeiten zuerst**: Phase 0 hebt alle Ökosysteme auf den Stand der offenen
  Dependabot-PRs oder neuer; Code-Phasen bauen darauf auf. Ein Versionssprung von Go,
  TypeScript oder Node wird repo-weit nachgezogen (`go.work`, alle `go.mod`, Dockerfiles,
  CI-Workflows, `AGENTS.md`-Tech-Stack, `docs/`), sonst gilt er als nicht erledigt.
- **Release-Reihenfolge**: erst alle Phasen, dann v1.0.0. Vereine mit Einsatz im September
  arbeiten mit v0.17.3, das produktionsreif ist.

## Inventory

- `docs/leitfaden/installation.md — ## Bondruck einrichten (optional)` — nennt „Ethernet"
  als einzige Anschlussart, keine Modelle, kein USB-Hinweis
- `README.md — ### Print-Relay` und `website/src/components/FaqAccordion.tsx` (Antwort
  „Brauchen wir spezielle Hardware?") — dieselbe Ethernet-Angabe; die FAQ-Copy ist
  handgeschrieben und steht nicht in `website/src/lib/published-docs.ts`
- `packaging/windows/KURZANLEITUNG.md` — Absatz „Danach läuft jotti auch ohne Internet"
  widerspricht der TSE-Internetpflicht in `docs/leitfaden/haeufige-fragen.md`
- `docs/leitfaden/fehlersuche.md — ## Router-Hinweise` — FRITZ!Box-Rezept ohne den
  Hinweis, dass die Ausnahme nach Router-Neustart einmal Internet braucht
- `docs/leitfaden/haeufige-fragen.md` — keine Antwort zu Stromausfall, Geräteanzahl,
  Helfer-Verzehr, Drucker-Notwendigkeit, Größenordnung der TSE-Kosten
- `docs/leitfaden/belege-steuersaetze.md` — Befreiung von der Aushändigung beschrieben
- `website/src/pages/fuer-vereine.astro` + `website/src/components/AnfrageFormular.tsx`
  — Formular öffnet `mailto:` per JS-Navigation; Erfolgs-State ohne Installationslink;
  kein Kopier-Fallback, wenn kein Mailprogramm reagiert
- `website/src/lib/anfrage-mailto.ts — buildMailtoUrl()` mit Tests in
  `website/src/lib/anfrage-mailto.test.ts`
- `TERMS.md` — deutsches Recht, kein Hinweis auf Österreich/Schweiz
- `frontend/src/components/common/VariantNamePreis.tsx — VariantNamePreis()` —
  einzeiliges `truncate`; ähnliche Variantennamen werden auf dem Handy identisch
- `frontend/src/service/components/table/ProductList.tsx — ProductList()` —
  Kategorie-Pills (seit vor v0.17.2 vorhanden), danach eine lange Liste ohne Sortierung
- `frontend/src/service/components/ServiceSplitLayout.tsx` — ab 1024 px feste Spalte
  für den Abschluss
- `backend/api/druck/bondruck/application/arbeitsbon_policy.go —
createAbholbonAuftraege()` — ein Auftrag je Position oder je Bestellung
- `backend/api/druck/bondruck/application/escpos/formatter.go —
FormatDirektverkaufAbholbon()` — delegiert an `FormatSammelBon()`; druckt je Position
  `Nx Bezeichnung`
- `backend/domain/druckstation/druckstation.go` — Bonmodus-Typ, `HatBonmodus()` und
  `Validate()`; die Doc-Kommentare behaupten dort, `abholbon` trage keinen Bonmodus
- `backend/api/druck/station/http/handler.go` — Druckstationen-Konfiguration
- `frontend/src/admin/settings/DruckstationBackend.ts — BonmodusSchema` und
  `frontend/src/admin/settings/DruckstationConfigPage.tsx` — Auswahl der Bonmodi
- `database/migrations/01_initial.up.sql` — CHECK auf `druckstationen.bonmodus`
- `docs/language.md — #### Abholbon` und `#### Bonmodus`; `docs/handbuch.md —
  ### 4.6 Bondruck: Arbeitsbon und Kassenbeleg (K-12)`
- `docs/compliance.md — ### 3.5 TSE-Varianten und Anbieter-Entscheidung` — Hardware-TSE
  als ausgeschlossen begründet
- `docs/anforderungen.md` — Funktionsumfang-Tabelle und Nicht-Ziele-Tabelle
- `docs/plans/plan-bondruck-ursachenklaerung.md` — offen
- `docs/plans/guide-manuelle-qa-v1.0.0.md` — offen; verweist auf die nicht existierende
  Datei `plan-v1.0-release-blockers.md`
- Offene externe PRs: #109 (Reihenfolge für Produkte/Varianten), #110 (Variantenname auf
  eigener Zeile), #111 (Produktebene über der Variantenliste)
- `e2e/playwright.config.ts` — Projekt mit `devices['Pixel 7']` für Handy-Viewport
- `.github/dependabot.yml` — monatliche Gruppen je Ökosystem (github-actions, sechs
  Go-Module, npm für `frontend/`, `website/`, `e2e/`, Docker für `backend/` und
  `reverse-proxy/`)
- Offene Dependabot-PRs: #118 (frontend-npm, 39 Updates, darunter TypeScript 6 → 7,
  jsdom 29 → 30, jest-dom 6 → 7, eslint-plugin-simple-import-sort 13 → 14), #117
  (website-npm, 15 Updates, darunter Astro 6 → 7, @astrojs/react 5 → 6, TypeScript 7),
  #116 (e2e-npm, 3), #115 (resolver: miekg/dns), #114 (backend: x/crypto, x/text), #113
  und #112 (golang 1.26.5-alpine → 1.27.0-alpine in `reverse-proxy/` und `backend/`),
  #106 (actions/setup-go 6 → 7, actions/setup-node 6 → 7)
- `backend/go.mod` — `go 1.26.5`; `AGENTS.md` Tech-Stack-Tabelle nennt Go 1.26 und
  TypeScript 6.0

## Resolved decisions

- **Vollständig umsetzen, dann releasen.** Alle Phasen landen vor v1.0.0; kein Datum
  gegenüber Vereinen.
- **Scroll-Problem** wird über Reihenfolge (#109) gelöst. Die Produktebene (#111) wird
  ohne Feldtest entschieden: per Code-Review und Design-Check gegen das aktuelle
  Service-Layout, nachdem #109 gelandet ist. Die Kategorie-Pills existieren bereits, der
  PR darf sie nicht duplizieren.
- **Mobile-Abschneiden** wird über Umbruch statt Kürzung gelöst; #110 ist der Kandidat.
- **Bon pro Stück** wird nicht 1:1 übernommen. Die Anforderung (mehrere Einheiten auf
  einmal kaufen, einzeln an der Theke einlösen) deckt der Abholbon-Modus `pro_stueck`.
- **Bon per E-Mail** wird nicht gebaut. Die Anforderung („Beleg ohne eigenen Drucker")
  ist nicht erfüllbar: jotti erzeugt den Kassenbeleg nur über die Druckstation
  `kassenbeleg`, und auf Verlangen ist der Beleg auch mit Befreiung Pflicht. Die FAQ
  sagt das klar; elektronischer Beleg bleibt Nicht-Ziel, weil das Gast-Handy nicht im
  Vereins-LAN ist.
- **Vereinslogo auf dem Bon** und **Helferdeckel** sind Nicht-Ziele.
- **USB-/Hardware-TSE** bekommt einen Kosten- und Anbieter-Spike mit ADR, keine
  Implementierung. `docs/compliance.md — ### 3.5` (Cloud-TSE gesetzt) bleibt bis zur ADR
  unverändert; kein Eintrag in der Nicht-Ziele-Tabelle.
- **Druckerliste** nennt im Einsatz bestätigte Modelle (Epson TM-T20IV per Ethernet,
  Sam4s H-Cube per WLAN) plus eine recherchierte Kaufempfehlung mit Preisklasse; USB-Drucker
  sind nicht unterstützt und werden so benannt.
- **Österreich/Schweiz**: Nutzung erlaubt, Klartext auf Website und in TERMS, dass nur die
  deutsche KassenSichV abgebildet ist.
- **Formular-Fallback** bleibt serverlos: der Mailtext wird auf der Seite gezeigt und ist
  kopierbar.
- **TERMS bekommt ein neues Fassungsdatum**, weil der Absatz zu Österreich/Schweiz eine
  inhaltliche Änderung ist. Bestehende Annahmen unter der Fassung vom 14. Juli 2026
  bleiben gültig. Das neue Datum ist der Tag, an dem Phase 2 landet.
- **Alle Dependabot-Updates werden übernommen**, auch die Major-Sprünge (TypeScript 7,
  Astro 7, Go 1.27). Ein Update, das sich nicht grün bekommen lässt, wird gepinnt und mit
  Begründung unter „Open questions / Risks" eingetragen, nicht still übersprungen.
- **Dependabot-PRs werden nicht manuell gemerged.** Die Updates landen als eigene Commits
  je Ökosystem auf `main`; Dependabot schließt seine PRs danach selbst.

## Open questions / Risks

- Zwei Vereine setzen jotti in der zweiten Septemberhälfte produktiv ein, einer davon
  zweitägig mit zwei Druckern. Phase 6 sollte vorher abgeschlossen sein.
- Die externen PRs sind nach dem Repo-Stand vom 04.08. entstanden; Konflikte mit
  späteren Änderungen sind beim Review zu prüfen.
- Die Phasen 1/8 (`haeufige-fragen.md`), 5/7 (`language.md`, `handbuch.md`), 7/9
  (`anforderungen.md`) und 4/5 (`ProductList.tsx`) schreiben je dieselbe Datei. Bei
  paralleler Ausführung wird vor dem Merge rebased und die Datei erneut gelesen.
- Ein Verein setzt jotti auch für die laufende Bewirtung im Vereinsheim ein. Das ist nicht
  die Zielgruppe (2–3 Feste pro Jahr), wird aber nicht verhindert. Ob das ein Nicht-Ziel
  wird, ist offen.
- Die Major-Updates in Phase 0 (TypeScript 7, Astro 7, Go 1.27) können Build- oder
  Typfehler auslösen, die vor jeder Code-Phase behoben sein müssen.

---

## Phase 0: Abhängigkeiten aktualisieren

**Depends on**: none

### Context

- `.github/dependabot.yml` — Gruppen und Verzeichnisse
- Offene Dependabot-PRs #106, #112–#118 (siehe Inventory) als Vorgabe für den Zielstand
- `backend/go.mod`, `resolver/go.mod`, `reverse-proxy/go.mod`, `windows/relay/go.mod`,
  `windows/starter/go.mod`, `go.work` — Go-Version und Module
- `backend/Dockerfile`, `reverse-proxy/Dockerfile` — Basis-Image `golang:…-alpine`
- `frontend/package.json`, `website/package.json`, `e2e/package.json` — npm-Stände,
  `packageManager`-Pin
- `.github/workflows/ci.yml`, `release.yml`, `security-scans.yml`, `fuzz.yml` —
  Action-Versionen, Go- und Node-Setup
- `AGENTS.md` — Tech-Stack-Tabelle; `docs/` — Versionsnennungen (`grep -rn '1\.26\|6\.0'`)

### What to build

Jedes Ökosystem wird auf den Stand der offenen Dependabot-PRs oder neuer gehoben, je
Ökosystem ein Commit: GitHub Actions, Go-Module (alle sechs, auch die drei ohne offenen
PR), Docker-Basis-Images, npm in `frontend/`, `website/` und `e2e/`. Versionssprünge von
Go, TypeScript und Node werden repo-weit nachgezogen (Versionskonsistenz-Regel). Brüche
durch Major-Updates (TypeScript 7: Typfehler; Astro 7: Integrations- und Starlight-API;
Go 1.27: Vet- und Lint-Regeln) werden im selben Commit behoben. Lockfiles entstehen nur
über die Tooling-Befehle (`pnpm install`, `go mod tidy`), nie von Hand. Nach dem Landen
auf `main` schließt Dependabot die acht PRs selbst; die drei Feature-PRs #109–#111 werden
danach gegen den neuen Stand reviewt.

### Acceptance criteria

- [ ] Alle in #106 und #112–#118 genannten Pakete stehen mindestens auf der dort genannten
      Version; `resolver/`, `reverse-proxy/`, `windows/relay/`, `windows/starter/` sind auf
      demselben Go-Stand
- [ ] `go.work`, alle `go.mod`, beide Dockerfiles, CI-Workflows, `AGENTS.md` und `docs/`
      nennen dieselbe Go-Version; `frontend/` und `website/` dieselbe TypeScript-Version
- [ ] Ein Commit je Ökosystem mit Conventional-Commit-Betreff `chore(deps): …`
- [ ] `make verify`, `make website-check` und `make test-e2e` grün; CI grün inklusive
      `security-scans` (govulncheck, pnpm audit)
- [ ] Die acht Dependabot-PRs sind nach dem Landen geschlossen (durch Dependabot) oder,
      falls nicht, mit Verweis auf den Commit manuell geschlossen
- [ ] Nicht übernehmbare Updates stehen mit Begründung unter „Open questions / Risks"

---

## Phase 1: Leitfaden-Korrekturen

**Depends on**: none

### Context

- `packaging/windows/KURZANLEITUNG.md` — falscher Offline-Satz
- `docs/leitfaden/fehlersuche.md — ## Router-Hinweise` — Router-Neustart
- `docs/leitfaden/installation.md — ## Bondruck einrichten (optional)` — Drucker
- `docs/leitfaden/haeufige-fragen.md` — neue Fragen
- `docs/leitfaden/belege-steuersaetze.md` — Befreiung von der Aushändigung
- `README.md — ### Print-Relay` — Anschluss-Angabe zum Bondrucker
- `website/src/components/FaqAccordion.tsx` — Antwort „Brauchen wir spezielle Hardware?"

### What to build

Doku-Änderungen, die Fragen aus dem Feld ohne Rückfrage beantworten. Kurzanleitung:
„Zertifikat und Handy-Zugang laufen danach ohne Internet; die TSE braucht beim Fest
Internet." Fehlersuche: FRITZ!Box wendet die Rebind-Ausnahme nach Neustart erst mit
Internet an; ohne Internet die Fallback-Adresse nutzen. Installation: Abschnitt
„Bondrucker" mit den zwei bestätigten Modellen und Anschlussart, einer recherchierten
Kaufempfehlung (Modell, Anschluss, Preisklasse, Stand der Recherche) und dem Satz, dass
USB-Drucker nicht unterstützt werden. Weil ein bestätigtes Modell per WLAN angebunden
ist, wird die Anschluss-Angabe vereinheitlicht — „im Netzwerk erreichbar (Ethernet oder
WLAN), TCP-Port 9100, feste IP-Adresse empfohlen" — in `docs/leitfaden/installation.md`,
`README.md` und `website/src/components/FaqAccordion.tsx`. FAQ: Stromausfall (Server aus,
Kasse aus, Daten bleiben; danach den Start wie am Festtag wiederholen, die Status-Seite
zeigt die dann gültige Adresse; Server und Router an eine USV oder Powerbank), wie viele
Handys (bis 30 Helfer, ein Handy pro Servicekraft), Helfer-Verzehr (ein Tisch pro Helfer,
am Abend kassieren oder stornieren), „Braucht ihr einen Drucker?" (ja: die Befreiung
erspart nur das ungefragte Aushändigen, auf Verlangen muss der Beleg gedruckt werden;
Bon per E-Mail gibt es nicht).

### Acceptance criteria

- [ ] Kurzanleitung und FAQ widersprechen sich nicht mehr zur Internetfrage
- [ ] Fehlersuche nennt den Neustart-Fall und verweist auf die Fallback-Adresse
- [ ] Installation listet die zwei bestätigten Modelle, eine Kaufempfehlung mit
      Preisklasse und Recherchedatum, und schließt USB aus
- [ ] FAQ beantwortet Stromausfall (inkl. Neustart und Status-Seite), Geräteanzahl,
      Helfer-Verzehr, Drucker-Notwendigkeit
- [ ] FAQ nennt den Drucker als nötig; die Befreiung erspart nur das ungefragte
      Aushändigen
- [ ] `docs/leitfaden/installation.md`, `README.md` und
      `website/src/components/FaqAccordion.tsx` nennen dieselbe Anschluss-Angabe
- [ ] `make check` und `make website-check` grün; die geänderten Leitfaden-Seiten rendern

---

## Phase 2: Website — Prozess und Ausland klarstellen

**Depends on**: 0 (Astro-7-Update vor Website-Änderungen)

### Context

- `website/src/pages/fuer-vereine.astro` — Seite mit Formular
- `website/src/components/AnfrageFormular.tsx` — Erfolgs-State
- `TERMS.md` — Geltungsbereich; Kopfzeile `Stand:` und E-Mail-Vorlage nennen beide das
  Fassungsdatum
- `website/src/lib/anfrage-mailto.ts — buildMailtoUrl()` — Annahmesatz mit Fassungsbezug,
  auch im Kommentar darüber
- `website/src/lib/anfrage-mailto.test.ts` — prüft den Annahmesatz wörtlich

### What to build

Auf `/fuer-vereine` und im Erfolgs-State: „Es gibt keine Freigabe und keine Zugangsdaten.
Mit dem Absenden könnt ihr installieren" mit direktem Link auf die Installationsseite und
dem Hinweis, den Spam-Ordner zu prüfen. Auf `/fuer-vereine` und in `TERMS.md` ein Absatz:
jotti setzt die deutsche KassenSichV um; für Österreich (RKSV) und die Schweiz ist nichts
vorgesehen, die Nutzung ist erlaubt, die Konformität Sache des Vereins. Die
Rechtsform-Auswahl bekommt keinen Länder-Eintrag; der Absatz reicht.

### Acceptance criteria

- [ ] Erfolgs-State enthält Installationslink und Spam-Hinweis
- [ ] `/fuer-vereine` und `TERMS.md` tragen den Absatz zu Österreich/Schweiz
- [ ] `TERMS.md` trägt ein neues Fassungsdatum (Tag des Landens) in der Kopfzeile
      `Stand:` und in der E-Mail-Vorlage; `buildMailtoUrl()` (Kommentar und Annahmesatz)
      und die Erwartung in `anfrage-mailto.test.ts` nennen dasselbe Datum — geprüft per
      `grep -rn 'Fassung vom' TERMS.md website/src`; `14. Juli 2026` kommt dort nicht
      mehr vor
- [ ] `make website-check` grün (`make check` deckt `website/` nicht ab)

---

## Phase 3: Website — Formular-Fallback ohne Mailprogramm

**Depends on**: 2 (beide Phasen schreiben denselben Erfolgs-State in
`AnfrageFormular.tsx` und berühren `anfrage-mailto.ts`)

### Context

- `website/src/lib/anfrage-mailto.ts — buildMailtoUrl()` — erzeugt Betreff und Text
- `website/src/components/AnfrageFormular.tsx` — öffnet `mailto:` per Navigation
- `website/src/lib/anfrage-mailto.test.ts` — bestehende Tests

### What to build

Nach dem Absenden zeigt die Seite zusätzlich Empfänger, Betreff und den vollständigen
Mailtext in einem Textfeld mit „Kopieren"-Button. Geräte ohne Mailprogramm (iOS ohne
Mail-App, Browser ohne Handler) können den Text so in jedes Postfach einfügen. Ein
reines Logik-Modul liefert die drei Teile getrennt, damit der Test sie ohne DOM prüft.

### Acceptance criteria

- [ ] Nach dem Absenden sind Empfänger, Betreff und Text sichtbar und kopierbar
- [ ] Test: die getrennten Teile entsprechen dem Inhalt der `mailto:`-URL
- [ ] Der bisherige `mailto:`-Weg bleibt unverändert
- [ ] `make website-check` grün (`make check` deckt `website/` nicht ab)

---

## Phase 4: Service-UI — Variantennamen auf dem Handy lesbar

**Depends on**: 0

### Context

- `frontend/src/components/common/VariantNamePreis.tsx — VariantNamePreis()` — `truncate`
- `frontend/src/service/components/table/ProductList.tsx — VariantRow()` — Zeile mit
  Stepper
- `frontend/src/admin/products/VariantChip.tsx — VariantChip()` — zweiter Konsument von
  `VariantNamePreis`, Chip mit Switch
- `frontend/src/components/common/VariantNamePreis.test.tsx` — Kommentar nennt `truncate`
- PR #110 — Kandidat
- `e2e/playwright.config.ts` — Projekt `mobile-service` (`devices['Pixel 7']`, ignoriert
  `admin-*.spec.ts`)

### What to build

Variantennamen dürfen nie so gekürzt werden, dass zwei Varianten desselben Produkts
gleich aussehen. Der Name bricht auf zwei Zeilen um; der Preis steht darunter oder
rechts, der Stepper behält seine Breite. Gilt für Tisch-Bestellung und Direktverkauf.
Beide Wege rendern `ProductList`. PR #110 wird dagegen reviewt: passt er, wird er
übernommen; sonst wird der Umbruch direkt in `VariantNamePreis` umgesetzt. Trifft es
`VariantNamePreis`, ändern sich die Admin-Variantenchips mit
(`frontend/src/admin/products/VariantChip.tsx`) — die Chip-Darstellung wird dann
mitgeprüft, und die Kontrakt-Kommentare in `VariantNamePreis.tsx` und
`VariantNamePreis.test.tsx` beschreiben danach den Umbruch statt der Kürzung.

### Acceptance criteria

- [ ] E2E-Test im Projekt `mobile-service` (Pixel-7-Viewport): zwei Varianten mit langem
      gemeinsamem Präfix sind vollständig lesbar
- [ ] Hoch- und Querformat des Handy-Viewports geprüft (Playwright `setViewportSize`)
- [ ] Falls `VariantNamePreis` geändert wurde: Admin-Preisliste (Variantenchips) geprüft
      und die Kontrakt-Kommentare in Komponente und Test angepasst
- [ ] `make test-e2e` grün (Playwright läuft weder in `make check` noch in `make verify`)
- [ ] `make check` grün

---

## Phase 5: Service-UI — Reihenfolge von Produkten und Varianten

**Depends on**: 0

### Context

- PR #109 — Migration `reihenfolge`, Endpunkte zum Verschieben, Admin-UI
- `database/migrations/README.md` — Regeln für additive Migrationen
- `docs/language.md` — Begriff „Reihenfolge" aufnehmen
- `docs/handbuch.md` — Sortierregel der Produktliste

### What to build

Admins legen die Reihenfolge von Produkten und ihren Varianten fest; die Service-Liste
sortiert innerhalb der Kategorie danach. Häufig bestellte Varianten stehen oben, das
Scrollen sinkt ohne neue Interaktionsebene. PR #109 wird reviewt und, wenn Migration
(additiv, forward-only), POST-only-Endpunkte und Validierung passen, übernommen.

### Acceptance criteria

- [ ] Neue Migration ist additiv, `01_initial.up.sql` unverändert; die Nummer ist beim
      Anlegen und erneut beim Rebase die nächste freie (`database/migrations/README.md`
      Regel 1) — Phase 6 und Phase 7 können ebenfalls eine Migration mitbringen
- [ ] Service-Liste sortiert nach (Kategorie, Reihenfolge, ID)
- [ ] `docs/language.md` und `docs/handbuch.md` beschreiben die Reihenfolge
- [ ] `make rebuild-projections` läuft nach der Migration fehlerfrei durch
      (`database/migrations/README.md` Regel 5)
- [ ] CI-Job `upgrade-path` grün — Pflicht-Gate für Schema-Änderungen
- [ ] `make verify` grün

---

## Phase 6: Bondruck-Ursachenklärung abschließen

**Depends on**: none

### Context

- `docs/plans/plan-bondruck-ursachenklaerung.md` — bestehender Plan, alle Punkte offen

### What to build

Den bestehenden Plan abarbeiten oder bewusst mit „Ursache nicht bestimmbar" schließen.
Nichts wird hier dupliziert; diese Phase ist nur die Frist: vor den produktiven
Einsätzen Ende September.

### Acceptance criteria

- [ ] Bestehender Plan hat keine offene Checkbox mehr oder ist gelöscht

---

## Phase 7: Abholbon pro Stück im Direktverkauf

**Depends on**: 0

### Context

- `backend/api/druck/bondruck/application/arbeitsbon_policy.go —
createAbholbonAuftraege()` — Ableitung der Abholbon-Aufträge
- `backend/api/druck/bondruck/application/escpos/formatter.go —
FormatDirektverkaufAbholbon()` → `FormatSammelBon()` — Zeilenformat `Nx Bezeichnung`
- `backend/domain/druckstation/druckstation.go` — Bonmodus-Typ, `HatBonmodus()`,
  `Validate()`; die Doc-Kommentare dort behaupten, `abholbon` trage keinen Bonmodus
- `backend/api/druck/station/http/handler.go` — zog-Vorfilter und DTO-Kommentar
- `frontend/src/admin/settings/DruckstationBackend.ts — BonmodusSchema`, `hatBonmodus()`
- `frontend/src/admin/settings/DruckstationConfigPage.tsx` — Auswahl
- `database/migrations/01_initial.up.sql` — CHECK und `COMMENT ON COLUMN` auf `bonmodus`
- `database/migrations/README.md` — Regeln für additive Migrationen, Upgrade-Pfad-Gate
- `docs/language.md — #### Bonmodus`, `docs/handbuch.md — ### 4.6`,
  `docs/anforderungen.md` — Funktionsumfang

### What to build

Gäste und Gruppen kaufen im Direktverkauf mehrere Einheiten auf einmal und lösen sie
einzeln an der Theke ein. Die Druckstation `abholbon` bekommt den dritten Bonmodus
`pro_stueck`: je Einheit einer Position ein eigener Abholbon mit `1x Bezeichnung`. Die
Aufteilung passiert in `createAbholbonAuftraege()`, das je Einheit eine Positions-Kopie
mit `Menge = 1` an `FormatDirektverkaufAbholbon()` übergibt; die Formatter bleiben
unverändert, `FormatPositionBon()` (Produktstationen) wird nicht angefasst.

Für die Produktstationen (`essen`, `getraenk`, `sonstiges`) bleibt der Wert unzulässig.
Maßgeblich ist das Domain-Modell: `druckstation.Validate()` prüft künftig eine
kategorieabhängige Menge — `abholbon`: {`pro_position`, `pro_bestellung`, `pro_stueck`},
Produktstationen: {`pro_position`, `pro_bestellung`}. Vorgelagert filtern das zog-Schema
in `api/druck/station/http/handler.go` und im Frontend `BonmodusSchema`. Die
Doc-Kommentare in `druckstation.go`, `handler.go` und `DruckstationBackend.ts`, die
`abholbon` einen Bonmodus absprechen, werden dabei richtiggestellt.

Die neue Migration ersetzt die CHECK-Bedingung und setzt `COMMENT ON COLUMN
druckstationen.bonmodus` neu; sie ändert keine Zeile. Die bestehende Bedingung ist
anonym und referenziert zwei Spalten, der generierte Constraint-Name ist daher vor dem
`DROP CONSTRAINT` zu prüfen. Danach erneuert `make sqlc` den Doc-Kommentar in
`backend/sqlc/dbgen/models.go`. Kassenbeleg und TSE sind nicht betroffen; der Abholbon
bleibt nicht-fiskalisch.

### Acceptance criteria

- [ ] Neue Migration `NN_abholbon_pro_stueck.up.sql` (`NN` = nächste freie Nummer beim
      Anlegen, `database/migrations/README.md` Regel 1; Phase 5 und Phase 6 können
      dieselbe Nummer beanspruchen) erlaubt `pro_stueck` nur für `abholbon`
- [ ] Die Migration setzt `COMMENT ON COLUMN druckstationen.bonmodus` neu und ändert
      keine Zeile in `druckstationen`; alle bestehenden Bonmodus-Werte bleiben
- [ ] `make sqlc` ausgeführt, `backend/sqlc/dbgen/models.go` mitcommittet
- [ ] Direktverkauf mit `3x Bier` und Modus `pro_stueck` erzeugt drei Druckaufträge mit
      `1x Bier`
- [ ] Modus `pro_stueck` an einer Produktstation wird von `Validate()`, zog-Schema und
      Admin-UI abgelehnt
- [ ] Kein Kommentar behauptet mehr, `abholbon` trage keinen Bonmodus
      (`druckstation.go`, `handler.go`, `DruckstationBackend.ts`)
- [ ] `docs/language.md` (`#### Abholbon`, `#### Bonmodus` inkl. DB-Enum),
      `docs/handbuch.md` § 4.6 und `docs/anforderungen.md` beschreiben den Modus
- [ ] `make rebuild-projections` läuft nach der Migration fehlerfrei durch
      (`database/migrations/README.md` Regel 5)
- [ ] CI-Job `upgrade-path` grün — Pflicht-Gate für Schema-Änderungen
- [ ] `make verify` grün

---

## Phase 8: TSE-Kosten und Hardware-TSE als ADR

**Depends on**: none

### Context

- `docs/compliance.md — ### 3.5 TSE-Varianten und Anbieter-Entscheidung`
- `docs/leitfaden/haeufige-fragen.md` — Frage „Was kostet der Betrieb?"
- `docs/leitfaden/tse-einrichten.md` — Abschnitt zu Kosten

### What to build

Ein Verein berichtet für die fiskaly-Cloud-TSE eine Mindestabnahme von mehreren Kassen
und eine Mindestlaufzeit; für zwei Feste im Jahr ist das die eigentliche Hürde. Der
Spike verifiziert die aktuellen fiskaly-Konditionen für eine Kasse, prüft alternative
Cloud-TSE-Anbieter mit Kurzzeitlizenz und schätzt, was eine Hardware-TSE am
Windows-Starter über das `TSEClient`-Adapter-Interface bedeuten würde. Ergebnis ist eine
ADR und eine ehrliche Kostenaussage in der FAQ.

### Acceptance criteria

- [ ] ADR `09_tse-kosten-und-hardware-tse.md` mit Konditionen, Alternativen,
      Entscheidung, in der Tabelle in `docs/adrs/README.md` verlinkt
- [ ] FAQ nennt Größenordnung und Vertragsbindung der TSE mit Datum der Recherche
- [ ] `docs/compliance.md` Abschnitt 3.5 stimmt mit der ADR überein

---

## Phase 9: Nicht-Ziele aus dem Feedback festhalten

**Depends on**: none

### Context

- `docs/anforderungen.md` — Nicht-Ziele-Tabelle
- `docs/produktbeschreibung.md` — Abgrenzung

### What to build

Drei Punkte werden als Nicht-Ziele mit Begründung eingetragen: Vereinslogo auf dem Bon
(kosmetisch, Raster-Druck und Logo-Upload ohne Kernnutzen; der Vereinsname steht im
Kopf), elektronischer Beleg per E-Mail oder Link (Gast-Handy ist nicht im Vereins-LAN,
Mailversand vom Vereins-Server und Adress-Erfassung am Tisch; der Kassenbeleg-Drucker
deckt den Bedarf), Helferdeckel (ein Tisch pro Helfer deckt es ab). Die
Tabelle bleibt die einzige Stelle; die FAQ aus Phase 1 verweist darauf.

### Acceptance criteria

- [ ] Drei Zeilen in der Nicht-Ziele-Tabelle mit je einer Begründung; die Spalte `Ex-ID`
      bleibt „—", weil keiner der drei Punkte je eine Anforderungs-ID trug
- [ ] `docs/produktbeschreibung.md` Abgrenzung stimmt damit überein

---

## Phase 10: Produktebene bewerten

**Depends on**: 5

### Context

- PR #111 — Produkt-Kacheln vor der Variantenliste
- `frontend/src/service/components/table/ProductList.tsx — ProductList()` — vorhandene
  Kategorie-Pills
- `docs/adrs/08_service-split-screen.md` — bestehende Entscheidung zum Service-Layout

### What to build

Zuerst prüfen, auf welchem Stand der PR entstanden ist und ob er die vorhandenen
Kategorie-Pills berücksichtigt oder ersetzt. Dann ohne Feldtest entscheiden, per
Code-Review und Design-Check gegen das Service-Layout mit sortierter Liste (Phase 5): Wie
viele Varianten bleiben je Kategorie nach der Sortierung sichtbar, und rechtfertigt das
einen zusätzlichen Tap pro Bestellung unter Stress? Ergebnis ist eine ADR: angenommen
(dann #111 auf den aktuellen Stand bringen und übernehmen) oder abgelehnt (dann #111 mit
Begründung schließen).

### Acceptance criteria

- [ ] Prüfergebnis zum PR-Stand gegenüber den Kategorie-Pills liegt in der ADR
- [ ] Die ADR nennt die Variantenzahl je Kategorie aus den Praxis-Setups (7 Produkte,
      ~50 Varianten) als Entscheidungsgrundlage
- [ ] ADR `10_produktebene-service.md` mit Status und Begründung, in der Tabelle in
      `docs/adrs/README.md` verlinkt
- [ ] PR #111 gemerged oder mit Verweis auf die ADR geschlossen

---

## Phase 11: Release v1.0.0

**Depends on**: 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10

### Context

- `docs/plans/guide-manuelle-qa-v1.0.0.md` — Rest-Guide, toter Verweis auf
  `plan-v1.0-release-blockers.md`
- `CHANGELOG.md` — Abschnitt `[1.0.0]` vorhanden

### What to build

Den QA-Guide durchlaufen, den toten Verweis entfernen, den Abholbon-Modus `pro_stueck` und
die Reihenfolge in den `[1.0.0]`-Abschnitt aufnehmen, Version und Tag `v1.0.0` setzen.
Vereine mit Einsatz vor dem Release arbeiten mit v0.17.3 und erhalten nach dem Release
eine Nachricht; ein Update ist ein ZIP-Tausch mit automatischem Backup. Nach dem Tag
wird `PREVIOUS_VERSION` an beiden Stellen auf `v1.0.0` gehoben —
`.github/workflows/ci.yml` (Job `upgrade-path`) und `database/migrations/README.md`
(Abschnitt „Vorversions-Pinning"). Der Bump ist ein eigener Commit nach dem Tag: vorher
gibt es die `v1.0.0`-Images noch nicht.

### Acceptance criteria

- [ ] QA-Guide ohne offene Checkbox, Datei gelöscht
- [ ] Tag `v1.0.0` und GitHub-Release vorhanden
- [ ] `CHANGELOG.md` `[1.0.0]` trägt Release-Datum, Reihenfolge und `pro_stueck`
- [ ] Release-Notes nennen die aktualisierten Laufzeit-Versionen (Go, Node, pnpm) aus
      Phase 0
- [ ] `PREVIOUS_VERSION` steht nach dem Tag in `.github/workflows/ci.yml` und
      `database/migrations/README.md` auf `v1.0.0`, Job `upgrade-path` grün
