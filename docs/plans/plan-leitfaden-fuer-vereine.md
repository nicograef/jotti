# Plan: Betreiber-Leitfaden als Website-Unterseite „Leitfaden für Vereine“

> Source PRD: n/a (aus Task-Beschreibung + Quelldokument `docs/betrieb/leitfaden-betreiber.md`)

## Goal

Das bestehende Dokument [docs/betrieb/leitfaden-betreiber.md](../betrieb/leitfaden-betreiber.md)
(„Betreiber-Leitfaden: jotti rechtssicher betreiben“) als eigenständige, gestylte Unterseite der
jotti-Website ([website/](../../website/)) bereitstellen — erreichbar unter der sauberen URL
`/leitfaden-fuer-vereine`, vollständig originalgetreu (alle 8 Abschnitte inkl. Tabellen,
Checkliste, FAQ, Glossar), im bestehenden Website-Design, prominent von der Startseite verlinkt.

## Architectural decisions

Durable decisions that apply across all phases:

- **Routes / URL**: Saubere URL `/leitfaden-fuer-vereine` über ein **Verzeichnis mit
  `index.html`**: `website/leitfaden-fuer-vereine/index.html`. Die jotti.rocks-nginx liefert
  `website/` als Document-Root aus (`root /usr/share/nginx/html/landing`, `try_files $uri $uri/
/index.html`, siehe [reverse-proxy/nginx.jotti-rocks.conf](../../reverse-proxy/nginx.jotti-rocks.conf#L59-L65)).
  nginx leitet `/leitfaden-fuer-vereine` automatisch per 301 auf `/leitfaden-fuer-vereine/` um und
  liefert dann die `index.html`. **Keine nginx-Änderung nötig** (vermeidet die „erst fragen“-Regel
  für Infrastruktur).
- **Asset-Pfade**: Gemeinsame Assets werden **root-relativ absolut** referenziert
  (`/css/base.css`, `/js/main.js`, `/img/…`, `/icons/…`, `/fonts/…`), damit sie aus dem
  Unterverzeichnis korrekt auflösen. Die CSP des Landing-Servers erlaubt nur `'self'` für
  Skripte/Styles/Fonts/Bilder — **keine neuen Fremdressourcen, keine Inline-Skripte**.
- **Styling**: Wiederverwendung des bestehenden Designsystems in
  [website/css/base.css](../../website/css/base.css) (CSS-Variablen in `:root`, `.container`,
  `.section`, `.card`, `.comparison-table`, `.check-list`, `.compliance-note`, `.badge-wip`,
  Header/Footer, Dev-Banner). Neues CSS wird **am Ende von `base.css`** ergänzt, ausschließlich für
  Elemente ohne bestehendes Pendant, und folgt dem vorhandenen Stil (natives CSS-Nesting,
  Tokens aus `:root`, **keine neuen Farben außerhalb `:root`**).
- **Kein JS-Framework**: Statisches HTML + bestehendes [website/js/main.js](../../website/js/main.js)
  (Mobile-Menü). Falls für das Inhaltsverzeichnis Interaktivität nötig ist, kommt sie in eine
  **same-origin** `.js`-Datei (CSP-konform), kein Inline-Script. Bevorzugt reine CSS/Anker-Lösung.
- **Section-Anker (durable)**: Acht Abschnitte mit stabilen IDs für TOC und Deep-Links —
  `#das-wichtigste`, `#verantwortung`, `#fiskalkonform`, `#gesetze`, `#pflichten`, `#checkliste`,
  `#faq`, `#glossar`.
- **Navigation (bidirektional)**: Die Unterseite verlinkt im Header zurück auf die
  Startseiten-Abschnitte (`index.html#…`) und auf die Startseite. Die Startseite verlinkt prominent
  auf die Unterseite (Header-Desktop-Nav, Mobile-Nav, Footer, Fiskalkonform-Abschnitt).
- **Inhaltstreue**: 1:1-Port aller 8 Abschnitte. Repo-relative Doku-Verweise → **GitHub-Blob-URLs**
  (`https://github.com/nicograef/jotti/blob/main/docs/…`), konsistent mit den bestehenden
  externen Doku-Links in `index.html`.

## Inventory

Bestehende Dateien, Muster und Integrationspunkte:

- **Quelldokument**: [docs/betrieb/leitfaden-betreiber.md](../betrieb/leitfaden-betreiber.md) —
  8 Abschnitte: (1) Das Wichtigste in 60 Sekunden, (2) Verantwortung Entwickler vs. Verein,
  (3) Was heißt „fiskalkonform“, (4) Die Gesetze in einfacher Sprache, (5) Pflichten Schritt für
  Schritt (4 Schritte), (6) Laufende Pflichten (Checkliste), (7) FAQ, (8) Glossar. Enthält mehrere
  Tabellen, Callouts (ℹ️/⚖️/⚠️/🔒), Rechtsgrundlagen-Hinweise (kursiv) und Querverweise auf
  `compliance.md`, `anforderungen.md`, `handbuch.md`, `lizenz-und-nutzung.md`.
- **Startseite**: [website/index.html](../../website/index.html)
  - Dev-Banner: `index.html:119`
  - Header-Desktop-Nav `.main-nav`: `index.html:146-167`
  - Mobile-Nav-Overlay: `index.html:172-190`
  - Fiskalkonform-Abschnitt `#fiskalkonform` (mit `.compliance-note` bei `:585`): `index.html:502-596`
  - Footer `.site-footer` (Footer-Links `:1025`): `index.html:1019-1065`
- **Designsystem**: [website/css/base.css](../../website/css/base.css)
  - Tokens `:root`: `base.css:5-32` · Typografie h1/h2/h3: `base.css:92-115` · Buttons: `base.css:137-205`
  - `.section`: `base.css:482` · `.compliance-note` (Callout-Muster): `base.css:548`
  - `.card`: `base.css:599` · `.check-list`: `base.css:703` · `.steps`: `base.css:732`
  - `.table-wrapper`/`.comparison-table`: `base.css:829-834` · `.tag`: `base.css:901`
  - `.site-footer`: `base.css:1117` · `.dev-banner`: `base.css:1204` · `.badge-wip`: `base.css:1217`
- **Fonts**: [website/css/fonts.css](../../website/css/fonts.css) — Montserrat, per `@import` in
  `base.css` geladen; Font-Pfade relativ zur CSS (`../fonts/…`), lösen daher unabhängig von der
  HTML-Position korrekt auf, solange die CSS als `/css/base.css` eingebunden wird.
- **JS**: [website/js/main.js](../../website/js/main.js) — Mobile-Menü-Toggle (`:1-21`).
- **Deployment**: [reverse-proxy/nginx.jotti-rocks.conf](../../reverse-proxy/nginx.jotti-rocks.conf#L40-L66)
  (Landing-Serverblock, `root` + `try_files`) und
  [docker-compose.jotti-rocks.yml](../../docker-compose.jotti-rocks.yml#L13)
  (`./website:/usr/share/nginx/html/landing:ro`).

## Resolved decisions

- **URL**: Saubere URL `/leitfaden-fuer-vereine` über Verzeichnis `website/leitfaden-fuer-vereine/index.html`.
- **Verlinkung**: Prominent und überall sinnvoll — Header-Desktop-Nav, Mobile-Nav, Footer und
  Fiskalkonform-Abschnitt der Startseite.
- **Inhaltstreue**: Vollständige, originalgetreue Umsetzung aller 8 Abschnitte inkl. Tabellen,
  Checkliste, FAQ, Glossar.
- **Querverweise**: Repo-relative Doku-Links → GitHub-Blob-URLs.

## Open questions / Risks

- **Nav-Label**: Vorschlag „Leitfaden“ (Alternative „Für Vereine“) für den neuen Nav-Eintrag.
  Anpassbar — die Startseiten-Nav ist bereits gut gefüllt (Features, Fiskalkonform, So geht's,
  Vergleich, Technik, Bereit?); ein 7. Eintrag kann auf schmalen Laptop-Breiten umbrechen. Mitigation:
  kurzes Label.
- **Saubere URL ohne Trailing Slash**: Die URL `/leitfaden-fuer-vereine` funktioniert über die
  automatische nginx-301-Umleitung auf `/leitfaden-fuer-vereine/`. Eine URL ganz **ohne** Trailing
  Slash und ohne Verzeichnis würde eine nginx-`location`-Regel erfordern (außerhalb des aktuellen
  Scopes; bräuchte Freigabe).
- **Callout-Palette**: Das Quelldokument nutzt mehrere Callout-Typen (Hinweis/Warnung/Recht). Die
  Website-Palette ist bewusst grün-monochrom. Callouts werden auf das bestehende
  `.compliance-note`-Muster abgebildet (Emoji als visuelle Unterscheidung), **ohne** neue Farben
  einzuführen.

---

## Phase 1: Seiten-Grundgerüst, Routing & Einstiegspunkte (Tracer Bullet)

### Context

- `reverse-proxy/nginx.jotti-rocks.conf:59-65` — `root` + `try_files`; bestimmt, dass ein
  Verzeichnis `leitfaden-fuer-vereine/` unter `/leitfaden-fuer-vereine/` ausgeliefert wird.
- `docker-compose.jotti-rocks.yml:13` — `website/` ist die Document-Root; neue Dateien sind ohne
  Build sofort statisch verfügbar.
- `website/index.html:1-115` — Head-Muster (Meta, OG/Twitter, Favicons, Theme-Color, Font-Preload)
  als Vorlage für den `<head>` der Unterseite.
- `website/index.html:119` (Dev-Banner), `:129-167` (Header), `:1019-1065` (Footer) — Chrome zum
  Wiederverwenden.
- `website/index.html:146-167` (Desktop-Nav), `:172-190` (Mobile-Nav), `:585` (Fiskalkonform-
  Callout), `:1025` (Footer-Links) — Einstiegspunkte für die Verlinkung.
- `website/css/base.css:5-32, 482, 548, 1117, 1204` — Tokens, `.section`, `.compliance-note`,
  Footer, Dev-Banner für das konsistente Erscheinungsbild.

### What to build

Eine erreichbare, korrekt gestylte (Teil-)Seite als durchgehender vertikaler Schnitt: Datei
`website/leitfaden-fuer-vereine/index.html` mit vollständigem `<head>` (eigener `<title>`,
Meta-Description, Open-Graph/Twitter, Favicons, `theme-color`, **Canonical** auf
`https://jotti.rocks/leitfaden-fuer-vereine/`), root-relativ eingebundenen Assets
(`/css/base.css`, `/js/main.js`, `/fonts/…`-Preloads), wiederverwendetem Dev-Banner, Header und
Footer. Der Header der Unterseite verlinkt zurück auf die Startseite und ihre Abschnitte
(`index.html`, `index.html#…`).

Inhaltlich enthält die Seite in dieser Phase bereits echten Inhalt: die Intro (Zielgruppe-Callout
und rechtlicher Disclaimer-Callout) sowie Abschnitt 1 „Das Wichtigste in 60 Sekunden“ (`#das-wichtigste`).

Gleichzeitig werden die **prominenten Einstiegspunkte** in `website/index.html` gesetzt: ein neuer
Nav-Eintrag in Desktop- und Mobile-Navigation, ein Link in den Footer-Links und ein kontextueller
Link/Button im Fiskalkonform-Abschnitt („Betreiber-Leitfaden für Vereine lesen“).

### Acceptance criteria

- [x] Seite ist unter `/leitfaden-fuer-vereine/` (und via 301 unter `/leitfaden-fuer-vereine`)
      erreichbar und lädt CSS, Fonts und JS aus `/css`, `/fonts`, `/js` (200, korrektes Rendering).
- [x] Dev-Banner, Header und Footer erscheinen identisch zum Website-Stil; Header-Links der
      Unterseite führen korrekt zurück zur Startseite und ihren Abschnitten.
- [x] Intro (Zielgruppe + rechtlicher Disclaimer) und Abschnitt „Das Wichtigste in 60 Sekunden“
      sind originalgetreu vorhanden und gestylt.
- [x] Startseite verlinkt prominent auf die Unterseite: Desktop-Nav, Mobile-Nav, Footer und
      Fiskalkonform-Abschnitt; alle Links funktionieren.
- [x] Keine Inline-Skripte und keine Fremdressourcen (CSP `'self'` eingehalten).

---

## Phase 2: Vollständiger Inhalts-Port (Abschnitte 2–8)

### Context

- `docs/betrieb/leitfaden-betreiber.md` — Quelle für alle restlichen Abschnitte (Tabellen,
  Schritt-für-Schritt, Checkliste, FAQ, Glossar, Rechtsgrundlagen-Hinweise).
- `website/css/base.css:834` (`.comparison-table`) — Tabellen-Stil zum Wiederverwenden für die
  Verantwortungs-, Bausteine-, Gesetze-, Fristen- und Glossar-Tabellen.
- `website/css/base.css:703` (`.check-list`) — Muster für die Checkliste (Abschnitt 6).
- `website/css/base.css:548` (`.compliance-note`) — Muster für alle Callouts (Hinweis/Warnung/Recht).
- `website/css/base.css:732` (`.steps`) und `:1217` (`.badge-wip`) — wiederverwendbar für die
  4 Pflicht-Schritte bzw. „In Entwicklung“-Status.

### What to build

Originalgetreue Umsetzung der Abschnitte 2–8 im Website-Design, jeweils mit stabiler Anker-ID:
Verantwortungs-Tabelle (`#verantwortung`), Fiskalkonform-Bausteine-Tabelle (`#fiskalkonform`),
Gesetze-Tabelle (`#gesetze`), Pflichten Schritt-für-Schritt inkl. Sub-Tabellen und Callouts
(`#pflichten`), Checkliste mit zwei Checkbox-Listen (`#checkliste`), FAQ als Frage/Antwort-Liste
(`#faq`) und Glossar-Tabelle (`#glossar`). Der abschließende „Mehr Details“-Verweisblock wird mit
GitHub-Links übernommen.

Bestehende Komponenten werden wiederverwendet (`.comparison-table`, `.check-list`,
`.compliance-note`, `.badge-wip`, `.steps`). Alle repo-relativen Doku-Verweise werden auf
GitHub-Blob-URLs umgestellt. Neues CSS wird nur dort ergänzt, wo kein passendes Pendant existiert
— voraussichtlich: eine generische Callout-Variante (auf Basis von `.compliance-note`), eine
FAQ-/Definitionslisten-Darstellung und eine Checkbox-Checkliste — alles innerhalb der bestehenden
Tokens und ohne neue Farben.

### Acceptance criteria

- [ ] Alle Abschnitte 2–8 sind inhaltlich vollständig und originalgetreu umgesetzt, inkl. aller
      Tabellen, Callouts, Rechtsgrundlagen-Hinweise und der Checkliste.
- [ ] Tabellen, Listen und Callouts nutzen die bestehenden Komponenten/Token; neu ergänztes CSS
      steht am Ende von `base.css`, ist minimal und führt keine neuen Farben außerhalb `:root` ein.
- [ ] Sämtliche Querverweise (compliance.md, anforderungen.md, handbuch.md, lizenz-und-nutzung.md)
      zeigen auf gültige GitHub-Blob-URLs.
- [ ] „In Entwicklung“-Status (TSE, Beleg, DSFinV-K) ist über `.badge-wip` bzw. den Dev-Banner
      konsistent kenntlich gemacht.
- [ ] Jeder Abschnitt hat die festgelegte stabile Anker-ID.

---

## Phase 3: In-Page-Inhaltsverzeichnis, Politur & QA

### Context

- `docs/betrieb/leitfaden-betreiber.md` — die Quelle besitzt ein klickbares Inhaltsverzeichnis;
  dieses wird als In-Page-Navigation nachgebildet.
- `website/css/base.css:45-48` — globales `scroll-behavior: smooth` und `scroll-padding-top`
  (bereits vorhanden, kommt Ankersprüngen zugute).
- `website/js/main.js:1-21` — bestehendes Mobile-Menü-Muster; Referenz, falls minimale
  same-origin-Interaktivität für ein aktives TOC nötig wird.

### What to build

Ein In-Page-Inhaltsverzeichnis mit Ankerlinks zu allen 8 Abschnitten (sticky oder als
Karten-/Listenblock unter der Intro), passend zum Website-Stil. Anschließend Politur und
Qualitätssicherung: Mobile-/Responsive-Verhalten (Tabellen-Overflow via `.table-wrapper`,
Umbrüche, Touch-Ziele), Accessibility (sinnvolle Überschriftenhierarchie, `aria`-Labels,
Fokus-Sichtbarkeit, ausreichender Kontrast), Bestätigung der CSP-Konformität (keine Inline-Skripte,
keine Fremdressourcen) und Prüfung aller internen wie externen Links.

### Acceptance criteria

- [ ] In-Page-Inhaltsverzeichnis verlinkt korrekt auf alle 8 Abschnitte; Ankersprünge landen mit
      korrektem Scroll-Offset (kein vom Header verdeckter Titel).
- [ ] Seite ist auf Mobil (≤640px) und Desktop sauber lesbar; breite Tabellen scrollen horizontal
      statt das Layout zu sprengen.
- [ ] Grundlegende Accessibility erfüllt: korrekte Überschriftenstruktur, Nav-`aria`-Labels,
      sichtbarer Fokus.
- [ ] Keine CSP-Verstöße (nur `'self'`-Ressourcen, keine Inline-Skripte); alle Links (intern +
      GitHub) sind gültig.
