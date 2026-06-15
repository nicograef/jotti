# Plan: Website-Audit – Findings & Empfehlungen umsetzen

> Source PRD: n/a (abgeleitet aus dem Code-Audit der Website, Konversation 2026-06-15)

## Goal

Die statische Website unter `website/` (jotti.rocks) auf den **realen
Funktionsstand** bringen und die im Audit gefundenen Konsistenz-, A11y- und
Perf-Mängel beheben. Kern ist ein Abgleich der „In Entwicklung"-Aussagen mit
der maßgeblichen [anforderungen.md §6](../anforderungen.md) und dem Code: TSE
und Beleg sind **live**, eine beworbene Hash-Chain wird **bewusst nicht
gebaut**. Die Website unterschätzt also Erreichtes und bewirbt zugleich ein
Won't-have.

Jede Phase ist eigenständig deploybar und über `make website-check`
(SSI, Links/Anker, Assets, CSS-Klassen, Prettier) verifizierbar.

## Architectural decisions

Durchgängige Festlegungen für alle Phasen:

- **Status-Quelle der Wahrheit**: [docs/anforderungen.md §6](../anforderungen.md)
  (Fiskal) bzw. §1/§5 (Features). Eine Website-Aussage „verfügbar/In
  Entwicklung" muss exakt dem dortigen Symbol (✅/🔲/🚫) entsprechen.
- **Realer Fiskal-Status** (verifiziert gegen anforderungen.md + Code):
  - **F-02 TSE-Integration ✅** — [backend/api/table/application/tse_signing.go](../../backend/api/table/application/tse_signing.go),
    [backend/api/direktverkauf/application/tse_signing.go](../../backend/api/direktverkauf/application/tse_signing.go),
    Tabelle `tse_signaturen` (`signatur`, `qr_code_data`, `signatur_zaehler`),
    [tse_nachsignier_worker.go](../../backend/app/tse_nachsignier_worker.go).
  - **F-03 Belegausgabe ✅** inkl. TSE-Signatur + QR — [processtype.go](../../backend/domain/tse/processtype.go) `Kassenbeleg-V1`.
  - **F-08 Hash-Chain 🚫 Won't-have** — [anforderungen.md:302](../anforderungen.md)
    („keine Hash-Chain"); Integrität = Append-only (K-07) + TSE-Signatur.
    Kein Hash-Chain-Code im Repo.
  - **F-04 DSFinV-K-Export 🔲 offen**, **R-02 CSV 🔲**, **K-13 KDS 🔲**,
    **K-15 Ausgabestationen 🔲** — Website-Badges hier korrekt, bleiben.
- **Fiskal-Headline-Framing** (Nutzerentscheid): Hero-Badge „In Entwicklung"
  **bleibt**, aber neu begründet — offen ist nur noch **DSFinV-K-Export**
  (+ Archivierung/ELSTER-Doku). TSE + Beleg werden als erledigt dargestellt.
- **Server-Specs-Quelle der Wahrheit**: [docs/betrieb/leitfaden-hosting.md:184-186](../betrieb/leitfaden-hosting.md)
  → Minimum **1 vCPU / 2 GB RAM / 20 GB SSD**. Die Website-Hosting-Seite stimmt
  damit überein; die Startseite wird angeglichen.
- **Kein Build-Schritt**: reines HTML/CSS/SSI. Bei jeder geänderten Seite die
  `lastmod` in [sitemap.xml](../../website/sitemap.xml) auf das Änderungsdatum
  setzen.

## Inventory

Website-Dateien mit den relevanten Stellen:

- [website/index.html:39-43](../../website/index.html#L39) — Hero „Fiskalkonform [In Entwicklung]".
- [website/index.html:10](../../website/index.html#L10) — Meta-Description nennt „DSFinV-K-Export" (noch offen).
- [website/index.html:194-200](../../website/index.html#L194) — Vergleichstabelle „Hash-Chain + Event-Sourcing".
- [website/index.html:399-406](../../website/index.html#L399) — Feature-Karte „TSE-Anbindung" mit `badge-wip`.
- [website/index.html:407-416](../../website/index.html#L407) — Feature-Karte „Kryptografische Hash-Chain" mit `badge-wip` (Won't-have).
- [website/index.html:417-425](../../website/index.html#L417) — Karte „Kassenbeleg": „TSE-Signatur, QR-Code … folgen mit der TSE-Phase".
- [website/index.html:433-441](../../website/index.html#L433) — Karte „DSFinV-K-Export" (bleibt WIP).
- [website/index.html:461-470](../../website/index.html#L461) — „kryptografisch verkettet und per TSE signiert".
- [website/index.html:208-385](../../website/index.html#L208) — Feature-Sektion: Kategorie-`h3` und Item-`h3` auf gleicher Ebene.
- [website/index.html:60-66](../../website/index.html#L60) — Hero-Bild (`width`, **kein** `height`, `loading="eager"`).
- [website/index.html:111-201](../../website/index.html#L111) — Vergleichstabelle ohne `scope`/Zeilen-`th`.
- [website/index.html:692-728](../../website/index.html#L692) — Anforderungen: Server „1 GB RAM, 10 GB SSD" (Zeile 698), „z.B." (Zeile 727).
- [website/partials/dev-banner.html:1-5](../../website/partials/dev-banner.html#L1) — nennt TSE-Anbindung + Belegausgabe + DSFinV-K als nicht verfügbar.
- [website/leitfaden-fuer-vereine/index.html:100-108](../../website/leitfaden-fuer-vereine/index.html#L100) — Callout „Aktueller Stand".
- [website/leitfaden-fuer-vereine/index.html:206-237](../../website/leitfaden-fuer-vereine/index.html#L206) — Fiskal-Status-Tabelle (TSE-Signatur/Beleg/DSFinV-K).
- [website/leitfaden-fuer-vereine/index.html:120-171](../../website/leitfaden-fuer-vereine/index.html#L120) / [Hosting-Tabellen](../../website/jotti-selbst-betreiben/index.html#L107) — weitere Tabellen ohne `scope`.
- [website/404.html:31](../../website/404.html#L31) — Seite ohne `<h1>` (nur `<h2>`).
- [website/partials/header.html:25-31](../../website/partials/header.html#L25) + [website/js/main.js](../../website/js/main.js) — Mobile-Toggle: statisches `aria-label`, kein `Esc`/Fokus.
- [website/sitemap.xml](../../website/sitemap.xml) — `lastmod` je Seite.
- [reverse-proxy/nginx.rocks.conf:72](../../reverse-proxy/nginx.rocks.conf#L72) (Landing-Block) + [nginx.website-dev.conf](../../reverse-proxy/nginx.website-dev.conf) — CSP `style-src 'self' 'unsafe-inline'`.
- [scripts/check-website.sh](../../scripts/check-website.sh) — Verifikationsskript (CI: [ci.yml:263-286](../../.github/workflows/ci.yml#L263)).

## Resolved decisions

- **Fiskal-Status statt Pauschal-Annahme**: Website-Aussagen werden 1:1 gegen
  [anforderungen.md §6](../anforderungen.md) + Code abgeglichen (Nutzervorgabe).
- **Hash-Chain**: Feature ersatzlos entfernen (Won't-have F-08), nicht durch ein
  anderes noch-offenes Feature ersetzen. Integritätsaussage stützt sich auf
  Event-Sourcing (Append-only) + TSE-Signatur.
- **TSE + Beleg**: als verfügbar darstellen (Badges/„folgen"-Wording entfernen).
- **Fiskal-Headline**: „In Entwicklung" beibehalten, neu begründet mit
  DSFinV-K-Export (+ Archivierung/ELSTER-Doku) als verbleibend offen.
- **Server-Specs**: Startseite auf 2 GB RAM / 20 GB SSD angleichen.
- **Phasenschnitt**: thematisch gruppiert (Nutzerentscheid).
- **Bewusst keine Aktion** (dokumentiert, kein Refactoring): Navigations-Links
  doppelt in [header.html](../../website/partials/header.html) + [mobile-nav.html](../../website/partials/mobile-nav.html);
  Duplikation [nginx.website-dev.conf](../../reverse-proxy/nginx.website-dev.conf) ↔
  Landing-Block [nginx.rocks.conf](../../reverse-proxy/nginx.rocks.conf) (bewusster Prod-Spiegel).

## Open questions / Risks

- **Doku-Parität (außerhalb `website/`)**: [docs/produktbeschreibung.md:102](../produktbeschreibung.md)
  und [:135](../produktbeschreibung.md) enthalten denselben veralteten
  Hash-Chain-Claim. Empfehlung: in derselben Iteration mitkorrigieren
  (Website↔Doku-Konsistenz), formal aber außerhalb des Website-Audits.
- **CSP-Risiko (Phase 4)**: [nginx.rocks.conf](../../reverse-proxy/nginx.rocks.conf)
  ist geteilte Deployment-Config. Tightening **nur** im Landing-`server`-Block
  (Zeilen 54-100), nicht im Demo-App-Block (die SPA braucht `'unsafe-inline'`).
  Verifiziert: Landingpage hat 0 Inline-Styles/-Scripts.

---

## Phase 1: Fiskalkonformitäts-Status & -Framing (Inhalts-Korrektheit)

### Context

- [anforderungen.md §6](../anforderungen.md) — maßgeblicher Status (F-02 ✅, F-03 ✅, F-08 🚫, F-04 🔲).
- [website/index.html:39-43,194-200,399-441,461-470](../../website/index.html#L399) — Hero, Vergleichstabelle, Fiskal-Karten, Integritätsabsatz.
- [website/index.html:10](../../website/index.html#L10) — Meta-Description.
- [website/partials/dev-banner.html:1-5](../../website/partials/dev-banner.html#L1) — globaler WIP-Banner.
- [website/leitfaden-fuer-vereine/index.html:100-108,206-237](../../website/leitfaden-fuer-vereine/index.html#L100) — Callout + Status-Tabelle.

### What to build

Die Website spiegelt den realen Fiskal-Status:

- **TSE-Anbindung** wird verfügbar dargestellt: `badge-wip` von der Feature-Karte
  entfernen; in der Leitfaden-Status-Tabelle die Zeile „TSE-Signatur" auf
  „✅ vorhanden" setzen.
- **Beleg**: In der Kassenbeleg-Karte die Formulierung „TSE-Signatur, QR-Code und
  Steuer-Aufschlüsselung folgen mit der TSE-Phase" streichen bzw. als enthalten
  darstellen; Leitfaden-Tabellenzeile „Beleg" auf „✅ vorhanden".
- **Hash-Chain (Won't-have)**: Feature-Karte „Kryptografische Hash-Chain"
  ersatzlos entfernen; in der Vergleichstabelle „Hash-Chain + Event-Sourcing"
  → „TSE-Signatur + Event-Sourcing"; im Absatz „Eure Daten. Manipulationssicher."
  „kryptografisch verkettet" durch eine korrekte Aussage ersetzen (unveränderliche
  Events / Append-only + TSE-Signatur).
- **Fiskal-Headline** (Entscheid „WIP behalten, neu begründen"): Hero-Badge
  bleibt; dev-banner und Leitfaden-Callout nennen nur noch DSFinV-K-Export
  (+ Archivierung/ELSTER-Doku) als offen, statt TSE/Beleg.
- **Meta-Description** ([index.html:10](../../website/index.html#L10)): DSFinV-K-Claim
  entschärfen (TSE bleibt), passend zum WIP-Framing.
- **DSFinV-K / KDS / Ausgabestationen / CSV**: unverändert (Badges korrekt).
- `lastmod` der geänderten Seiten in [sitemap.xml](../../website/sitemap.xml) aktualisieren.
- Optional (siehe Risks): denselben Hash-Chain-Fehler in
  [docs/produktbeschreibung.md:102,135](../produktbeschreibung.md) mitkorrigieren.

### Acceptance criteria

- [x] Keine Website-Stelle bewirbt eine „Hash-Chain"/„kryptografisch verkettet"
      (Grep nach „Hash-Chain"/„verkett" in `website/` ist leer).
- [x] „TSE-Anbindung" trägt keinen `badge-wip` mehr; Leitfaden-Tabelle weist
      TSE-Signatur und Beleg als vorhanden aus.
- [x] Kassenbeleg-Karte erweckt nicht mehr den Eindruck, TSE-Signatur/QR fehlten.
- [x] dev-banner und Leitfaden-Callout nennen als offen ausschließlich
      DSFinV-K-Export (+ Archivierung/ELSTER-Doku), nicht TSE/Beleg.
- [x] Hero-Badge „In Entwicklung" auf „Fiskalkonform" bleibt erhalten.
- [x] DSFinV-K-/KDS-/Ausgabestationen-/CSV-Badges unverändert vorhanden.
- [x] Jede inhaltlich geänderte Seite hat aktualisierte `lastmod` in der Sitemap.
- [x] `make website-check` ist grün (inkl. CSS-Klassen-Check nach Entfernen der Karte).

---

## Phase 2: Restliche Inhalts- & Daten-Konsistenz

### Context

- [website/index.html:698](../../website/index.html#L698) — Server „1 GB RAM, 10 GB SSD".
- [website/jotti-selbst-betreiben/index.html:343-353](../../website/jotti-selbst-betreiben/index.html#L343) — Minimum „2 GB RAM, 20 GB SSD".
- [docs/betrieb/leitfaden-hosting.md:184-186](../betrieb/leitfaden-hosting.md) — Quelle der Wahrheit.
- [website/index.html:727](../../website/index.html#L727) — „z.B." ohne `&nbsp;`.

### What to build

Eine thin slice „kleine Korrektheits-/Konsistenzfixes":

- Startseiten-Anforderungen an die Hosting-Seite + Hosting-Doku angleichen
  (1 vCPU / **2 GB RAM / 20 GB SSD**).
- „z.B." → „z.&nbsp;B." (einheitliche Abkürzungs-Typografie wie überall sonst).
- `lastmod` der geänderten Seite(n) in der Sitemap aktualisieren.

### Acceptance criteria

- [x] Startseite und Hosting-Seite nennen identische Server-Mindestanforderungen
      (2 GB RAM / 20 GB SSD), konsistent zu leitfaden-hosting.md.
- [x] Kein `z.B.` ohne geschütztes Leerzeichen mehr in `website/`.
- [x] Sitemap-`lastmod` aktualisiert; `make website-check` grün.

---

## Phase 3: Barrierefreiheit & Performance

### Context

- [website/index.html:111-201](../../website/index.html#L111) + Leitfaden/Hosting-Tabellen — `comparison-table` ohne Tabellen-Semantik.
- [website/index.html:208-385](../../website/index.html#L208) — doppelte `h3`-Ebene (Kategorie vs. Item).
- [website/404.html:31](../../website/404.html#L31) — kein `<h1>`.
- [website/partials/header.html:25-31](../../website/partials/header.html#L25) + [js/main.js](../../website/js/main.js) — Mobile-Menü.
- [website/index.html:60-66](../../website/index.html#L60) — Hero-Bild ohne `height` (LCP/CLS).

### What to build

Barrierefreiheits- und CLS-Politur ohne sichtbare Layout-Änderung:

- **Tabellen-Semantik**: Spaltenköpfe `scope="col"`; erste Zelle je Datenzeile
  als `<th scope="row">`. Gilt für alle `comparison-table` (Startseite, Leitfaden,
  Hosting). CSS ggf. minimal anpassen, falls `th`-Styling in der ersten Spalte
  abweicht (heute `td:first-child`).
- **Heading-Hierarchie**: In der Feature-Sektion Item-Titel von `h3` auf `h4`
  (Kategorie bleibt `h3`); auf der 404-Seite die Hauptüberschrift als `<h1>`.
- **Mobile-Menü**: `aria-label` bei geöffnetem Menü auf „Menü schließen" wechseln;
  `Esc` schließt das Overlay. Minimaler Zusatz in `main.js`.
- **Hero-Bild**: `height` passend zum WebP-Seitenverhältnis ergänzen (reserviert
  Platz, verhindert Layout-Shift beim eager-LCP-Bild).

### Acceptance criteria

- [x] Alle `comparison-table` haben `scope="col"` an Kopfzellen und `th scope="row"`
      in der ersten Spalte; Darstellung unverändert.
- [x] Dokument-Outline ist konsistent: Feature-Items sind `h4` unter `h3`-Kategorie;
      404-Seite hat genau ein `<h1>`.
- [x] Mobile-Toggle meldet den korrekten Zustand (`aria-label` + `aria-expanded`);
      `Esc` schließt das Menü und gibt den Body-Scroll frei.
- [x] Hero-Bild hat `width` **und** `height`; kein sichtbarer Sprung beim Laden.
- [x] `make website-check` grün.

---

## Phase 4: Optionales Hardening (CSP)

### Context

- [reverse-proxy/nginx.rocks.conf:54-100](../../reverse-proxy/nginx.rocks.conf#L54) — Landing-`server`-Block, CSP Zeile 72.
- [reverse-proxy/nginx.website-dev.conf](../../reverse-proxy/nginx.website-dev.conf) — Dev-Spiegel.

### What to build

Optionales Tightening der Content-Security-Policy für die Landingpage: Da die
Landingpage **keine** Inline-Styles enthält, kann `style-src 'self' 'unsafe-inline'`
im Landing-Block auf `style-src 'self'` reduziert werden. Ausschließlich der
Landing-`server`-Block wird angefasst (nicht der Demo-App-Block). Dev-Config zur
Parität analog.

> **Assumption:** Diese Phase betrifft `reverse-proxy/`, nicht `website/` — sie ist
> als „Empfehlung" aus dem Audit enthalten und kann entfallen, wenn nur `website/`
> angefasst werden soll.

### Acceptance criteria

- [ ] Landing-CSP enthält kein `'unsafe-inline'` mehr bei `style-src`; Demo-App-Block
      unverändert.
- [ ] Landingpage rendert unverändert (keine durch CSP blockierten Styles in der
      Browser-Konsole).
- [ ] `nginx -t` (bzw. Container-Start) erfolgreich; Security-Header-Vererbung intakt.
