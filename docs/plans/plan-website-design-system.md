# Plan: Website Design System & Auslieferungs-Verbesserung

> Source PRD: `docs/prds/prd-website-design-system.md`

## Goal

Die gewachsene ~1200-Zeilen-CSS-Datei der Website durch ein kleines,
jotti-eigenes utility-first Design System ersetzen (plain CSS, null
Dependencies, kein Build-Schritt) und die Auslieferung gezielt verbessern
(gzip, Cache-Header, Format-/Integritäts-Checks in Make und CI). Das
Erscheinungsbild bleibt unverändert — es ändert sich nur, _wie_ es
implementiert ist.

## Architectural decisions

- **Ein Stylesheet, ein Request**: `website/css/base.css` bleibt die einzige,
  direkt verlinkte CSS-Datei (Name unverändert — Cache-Busting unnötig, da
  CSS kurze max-age bekommt).
- **Cascade Layers**: Interne Struktur über native `@layer tokens, base,
components, utilities` — Utilities überschreiben Komponenten ohne
  `!important`, gleiche Semantik wie Tailwind 4 in der App.
- **Tokens**: CSS Custom Properties in `:root`; Farbwerte identisch zur App
  (oklch, manuell synchron gehalten, als solche kommentiert),
  Spacing-Skala in rem nach Tailwind-Viertel-Schritten, Radius, Fonts,
  Container-Breite.
- **Utilities**: Tailwind-Namenskonvention (`mx-auto`, `grid-cols-2`,
  `gap-6`, `text-muted`, …), kuratiert — nur tatsächlich benutzte Utilities
  existieren; der CSS-Klassen-Check erzwingt das.
- **Responsive**: mobile-first, zwei min-width-Breakpoints mit
  Präfix-Utilities — `md:` = 640px, `lg:` = 960px (escaped Klassennamen,
  z. B. `.md\:grid-cols-2`). Bestehende max-width-Queries werden invertiert.
- **Komponenten (~8–12)**: `btn` (+ Varianten/Größen), Site-Header +
  Navigation inkl. Mobile-Menü/`nav-open`, `dev-banner`,
  `comparison-table`, `compliance-note`, Check-/To-do-Listen
  (Pseudo-Elemente), `tag`, `quickstart-box`, `guide-prose`
  (Fließtext-Typografie inkl. `guide-steps`-Counter und `guide-faq`).
  `cta-section` darf als Komponente bleiben (auf Startseite und 404
  genutzt) oder in Utilities aufgehen — Entscheidung bei der Migration.
- **Checks**: Bash-Script `scripts/check-website.sh` (dependency-frei,
  grep/comm-basiert) prüft Links/Anker, Asset-Existenz (beidseitig),
  SSI-Includes und CSS-Klassen-Konsistenz (benutzt ⇆ definiert). Prettier
  läuft über die vorhandene Frontend-Installation (`pnpm exec prettier`),
  Konfig-Auflösung über `.editorconfig` (kein neues Config-File).
- **Make/CI**: Targets `website-check` (Script + Prettier-Check) und
  `website-fmt`; CI bekommt einen `website`-Pfadfilter + Job nach dem
  Muster der bestehenden Bereichs-Jobs.
- **Deploy unverändert**: `git pull` + `make jotti-rocks-up` auf dem VPS.

## Inventory

- `website/css/base.css:1-1192` — die abzulösende Monolith-Datei; Tokens
  bereits in `:root` (`32-58`), tote Showcase-Styles (`701-721`),
  `btn-outline` dupliziert `btn-sm`-Größe (`208-224`), `code`-Styling
  dreifach (`543-549`, `865-871`, `1140-1146`), Breakpoints 640px ×13 /
  960px ×10 als max-width.
- `website/index.html` — einzige Seite mit relativen Pfaden: Bilder
  (`61`, `441`, `685`), interne Links (`487`, `674`); obsolete Meta-Tags
  (`9-13`).
- `website/leitfaden-fuer-vereine/index.html:9-13`,
  `website/jotti-selbst-betreiben/index.html:9-13`, `website/404.html:9` —
  obsolete Meta-Tags (`keywords`, `X-UA-Compatible`).
- `website/partials/` — 6 SSI-Partials; `head-assets.html` verlinkt das
  Stylesheet; `header.html`/`mobile-nav.html` tragen die
  Navigations-Komponenten.
- `website/js/main.js` — Mobile-Menü, toggelt `nav-open` auf `body`;
  bleibt unverändert (Klassennamen-Kontrakt beachten).
- Ungenutzte Assets: `website/img/jotti-admin-products.webp`,
  `website/img/jotti-admin-reporting.png`,
  `website/img/jotti-admin-reporting.webp`,
  `website/icons/jotti-icon-dark-64.png`,
  `website/icons/jotti-icon-light-64.png`.
- `reverse-proxy/nginx.jotti-rocks.conf:38-76` — Landing-Block (prod):
  SSI, 404, Partials-Sperre, Security-Header; **kein gzip, keine
  Cache-Header**.
- `reverse-proxy/nginx.website-dev.conf` — lokaler Spiegel (`make
website`), bewusst duplizierte Logik (Kommentar `1-3`).
- `Makefile:220-224` — `website`-Target (nginx-Container, Port 8080);
  Help-Konvention `##`-Kommentare.
- `.github/workflows/ci.yml:10-28` — `changes`-Job mit
  `dorny/paths-filter`; `frontend-ci` (`92-126`) als Muster für
  pnpm-basierte Jobs.
- `frontend/.prettierrc`, `frontend/package.json:15-16` — vorhandene
  Prettier-Installation und Format-Scripts.
- `frontend/src/index.css:53-110` — kanonische jotti-Theme-Werte (Light
  Theme) als Referenz für die Token-Übernahme.
- `scripts/` + `test-integration.sh` — Prior Art für eigenständige
  Bash-Prüfscripts.

## Resolved decisions

- Plain CSS, handgeschrieben — kein Tailwind-Tooling, kein Generator,
  keine Dependencies (PRD-Klärung).
- Hybrid-Schnitt: Utilities für Layout/Spacing/Typo/Farben, wenige echte
  Komponentenklassen (PRD-Klärung).
- Mobile-first mit `md:`/`lg:`-Präfixen (PRD-Klärung).
- Tokens manuell synchron mit der App, keine geteilte Datei (PRD-Klärung).
- Build/Deploy-Scope: gzip + Cache-Header, Prettier + Checks in Make/CI;
  kein Auto-Deploy (PRD-Klärung).
- Check-Umfang: Prettier, Link-/Asset-/SSI-Integrität,
  CSS-Klassen-Konsistenz beidseitig; Docker-Smoke-Test bleibt manuell
  (PRD-Klärung).
- Keine Styleguide-Seite; Doku als Kommentarkopf in der CSS-Datei
  (PRD-Klärung).
- Phasenschnitt: 4 Phasen, Tooling vor der Migration (Plan-Klärung).
- 404-Seite wandert in Phase 3 (statt 4): Sie besteht ausschließlich aus
  Markup, das die Startseite und die Partials teilen (`cta-section`,
  Header/Footer) — eine Migration erst in Phase 4 würde den
  Klassen-Konsistenz-Check zwischen den Phasen brechen.

## Open questions / Risks

- **add_header-Vererbungsfalle (nginx)**: `add_header` in einem
  `location`-Block verwirft alle vom `server`-Block geerbten Header
  (HSTS, CSP, …). Cache-Steuerung deshalb über `expires` (per
  `map $sent_http_content_type`) oder Wiederholung der Security-Header —
  Akzeptanzkriterium prüft beides zusammen.
- **Prettier-Erstlauf erzeugt Format-Rauschen**: Der erste
  `website-fmt`-Lauf wird als eigener Commit von inhaltlichen Änderungen
  getrennt.
- **Pixel-Parität ist manuell**: Abgleich über `make website` auf drei
  Breakpoints pro Seite; kein automatisierter visueller Regressionstest
  (bewusst, PRD).

---

## Phase 1: Auslieferung — gzip + Cache-Header

**User stories**: 9, 10, 20

### Context

- `reverse-proxy/nginx.jotti-rocks.conf:38-76` — Landing-Block mit
  Security-Headern, ohne gzip/Caching.
- `reverse-proxy/nginx.website-dev.conf` — muss identische Regeln
  bekommen, damit `make website` die Produktion spiegelt.

### What to build

Der Landing-Page-nginx komprimiert textbasierte Antworten (HTML, CSS, JS,
SVG, XML) per gzip und liefert Cache-Control aus: Fonts und Bilder lang +
`immutable`, CSS/JS mit kurzer max-age (~1 h), HTML ohne Caching
(SSI-gerendert). Dev- und Prod-Config erhalten dieselben Regeln; die
Security-Header der Produktion bleiben auf allen Antworten erhalten
(add_header-Vererbung beachten).

### Acceptance criteria

- [ ] `make website`: `curl -sI -H 'Accept-Encoding: gzip'` zeigt
      `Content-Encoding: gzip` für HTML und CSS.
- [ ] `curl -sI` zeigt Cache-Control: lang + immutable für
      `/fonts/*.woff2` und `/img/*`, kurze max-age für `/css/*` und
      `/js/*`, kein Caching für HTML-Seiten.
- [ ] In der Prod-Config sind auf Asset-Antworten weiterhin alle
      Security-Header gesetzt (Config-Review; nach Deploy per
      `curl -sI https://jotti.rocks/css/base.css` verifiziert).
- [ ] Dev- und Prod-Config enthalten identische gzip-/Cache-Regeln.

---

## Phase 2: Aufräumen + Checks scharf schalten

**User stories**: 11, 12, 13, 14, 18, 19

### Context

- `website/css/base.css:701-721` — tote Showcase-Styles.
- `website/index.html:9-13,61,441,487,674,685` — obsolete Meta-Tags,
  relative Pfade.
- `website/leitfaden-fuer-vereine/index.html:9-13`,
  `website/jotti-selbst-betreiben/index.html:9-13`,
  `website/404.html:9` — obsolete Meta-Tags.
- `website/sitemap.xml` — `lastmod` aktualisieren.
- `Makefile:220-224`, `.github/workflows/ci.yml:10-28`,
  `test-integration.sh` — Integrationspunkte und Prior Art für Script,
  Targets und CI-Job.

### What to build

Erst aufräumen: tote Showcase-Klassen und die fünf ungenutzten Assets
entfernen, `meta keywords` und `X-UA-Compatible` von allen vier Seiten
löschen, relative Pfade der Startseite auf absolute umstellen,
Sitemap-`lastmod` aktualisieren.

Dann absichern: `scripts/check-website.sh` prüft interne Links/Anker,
Asset-Referenzen (beidseitig: referenziert ⇒ existiert, existiert ⇒
referenziert), SSI-Includes und CSS-Klassen-Konsistenz (benutzt ⇆
definiert). Make-Targets `website-check` (Script + Prettier-Check) und
`website-fmt` (Prettier schreibend, via Frontend-Installation). CI erhält
einen `website`-Pfadfilter und einen Job, der `website-check` ausführt.
Der Prettier-Erstlauf wird als separater Format-Commit eingebracht.

### Acceptance criteria

- [ ] Tote Klassen, ungenutzte Assets und obsolete Meta-Tags sind
      entfernt; Startseite nutzt durchgehend absolute Pfade; Website
      sieht unverändert aus (Sichtprüfung über `make website`).
- [ ] `make website-check` läuft grün; ein absichtlich eingebauter
      Fehler (kaputter Link, fehlendes Asset, unbenutzte CSS-Klasse,
      Tippfehler im Klassennamen) lässt es jeweils fehlschlagen
      (Negativprobe).
- [ ] `make website-fmt` formatiert `website/`; danach ist
      `website-check` (inkl. Prettier-Check) grün.
- [ ] CI führt den Website-Job nur bei Änderungen unter `website/` aus
      und ist auf dem PR grün.

---

## Phase 3: Design-System-Fundament + Startseite, Partials, 404

**User stories**: 1, 2, 3, 4, 5, 6, 7, 8, 16, 17

### Context

- `website/css/base.css:32-58` — bestehende Tokens als Ausgangspunkt.
- `frontend/src/index.css:53-110` — kanonische Theme-Werte für die
  Token-Übernahme.
- `website/index.html`, `website/404.html`, `website/partials/` — zu
  migrierendes Markup; `header.html`/`mobile-nav.html` tragen die
  Navigations-Komponenten, deren Klassen `js/main.js:2-3` referenziert.

### What to build

`base.css` wird auf `@layer tokens, base, components, utilities`
umgestellt: Kommentarkopf (Doku des Systems), Token-Block (Farben aus der
App, Spacing-Skala, Radius, Fonts, Container), schlanker Base-Layer
(Reset, Typografie-Grundlagen), dann die Migration: Startseite, alle
sechs Partials und die 404-Seite ersetzen ihre Einmal-Klassen durch
Utilities (mobile-first, `md:`/`lg:`-Präfixe statt invertierter
max-width-Queries); mehrfach genutzte Muster werden zu den definierten
Komponenten. Verwaiste Komponentenklassen werden gelöscht, die
`btn`-Größen-Duplikation aufgelöst. Es entstehen nur Utilities, die das
migrierte Markup tatsächlich nutzt. Die Leitfaden-Klassen
(`guide-*`-Styles) bleiben in dieser Phase unangetastet bestehen.

### Acceptance criteria

- [ ] `base.css` ist in die vier Layer gegliedert und beginnt mit einem
      Kommentarkopf, der Tokens, Skala und Namenskonvention erklärt.
- [ ] Startseite, 404 und Partials sind migriert; `make website-check`
      ist grün (keine verwaisten Klassen in beide Richtungen).
- [ ] Sichtprüfung auf drei Breakpoints (Mobil ~390px, Tablet ~768px,
      Desktop ≥1120px): Startseite und 404 sehen unverändert aus;
      Mobile-Menü öffnet/schließt wie bisher.
- [ ] Die Token-Werte stimmen mit der App überein und sind als manuell
      synchronisiert kommentiert.

---

## Phase 4: Leitfaden-Seiten migrieren — Endzustand kuratiertes Set

**User stories**: 4, 5, 8

### Context

- `website/leitfaden-fuer-vereine/index.html`,
  `website/jotti-selbst-betreiben/index.html` — zu migrierende Seiten.
- `website/css/base.css:1033-1192` — Guide-Styles: `guide-prose`
  (Fließtext), `guide-steps` (Counter), `guide-checklist`/`guide-faq`
  (Pseudo-Elemente) bleiben Komponenten; `guide-hero`, `guide-eyebrow`,
  `guide-lead`, `guide-chapter` werden auf Utilities geprüft.

### What to build

Beide Leitfaden-Seiten werden migriert: Einmal-Klassen
(Hero-/Eyebrow-/Lead-Styling) gehen in Utilities auf;
Fließtext-Typografie, nummerierte Schritte, To-do-Checklisten und FAQ
bleiben als `guide-prose`-Komponentenschicht (Pseudo-Elemente und Counter
sind als Utilities nicht sinnvoll). Danach werden alle nicht mehr
referenzierten Styles gelöscht — der Endzustand ist das kuratierte Set:
Tokens + Base + ~8–12 Komponenten + genau die benutzten Utilities.

### Acceptance criteria

- [ ] Beide Leitfaden-Seiten sind migriert; `make website-check` ist
      grün.
- [ ] Sichtprüfung beider Seiten auf drei Breakpoints: unverändertes
      Erscheinungsbild (inkl. Schritt-Nummerierung, Checklisten-Kästchen,
      FAQ, Tabellen).
- [ ] `base.css` enthält keine Einmal-Klassen mehr außerhalb des
      definierten Komponenten-Sets; die Datei ist messbar kleiner als
      die heutigen 1192 Zeilen.
- [ ] Voller Durchlauf: `make website` + Klick durch alle Seiten,
      `make website-check` grün, CI grün.
