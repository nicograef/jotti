# Plan: Website-Redesign (jotti.rocks)

> Source PRD: docs/prds/prd-website-redesign.md

## Goal

Die Marketing-Website nach dem Design-Handoff neu aufbauen: vier echte Routen
(Landing, „Für Vereine", Impressum, Datenschutz), neue Token-/Font-Welt (Space
Grotesk + Inter, Spektral-Einsatz voll) konsistent über Landing und Doku, per
Schalter wechselbares Hell/Dunkel-Theme, sechs React-Islands für die
interaktiven Teile, echte per Skript reproduzierbare App-Screenshots und ein
mailto-basiertes Anfrage-Formular. Die Produktiv-CSP bleibt unverändert und ist
hartes Abnahmekriterium. Alles innerhalb des bestehenden Astro/Starlight-Pakets
`website/`; App, Backend, Proxy und Doku-Inhalte bleiben unberührt.

## Architectural decisions

Durable decisions, die für alle Phasen gelten:

- **Routen.** `/` (Landing), `/fuer-vereine` (Anfrage), `/impressum`,
  `/datenschutz`. Die Doku bleibt unverändert unter `/docs/...`. Anker-IDs der
  Landing übernehmen die Handoff-IDs: `#features`, `#demo`, `#ablauf`,
  `#fuerwen`, `#screenshots`, `#preis`, `#sicherheit`, `#faq`, `#download`.
- **React-Islands.** `@astrojs/react` mit React 19 als genehmigte neue
  Dependencies des Website-Pakets; dazu Lucide über die offiziellen Pakete
  `@lucide/astro` (statische Sektionen) und `lucide-react` (FeatureExplorer),
  beide auf dieselbe Lucide-Version gepinnt — kein Sprite-Nachbau, die
  ISC-Lizenz reist im Paket mit. Genau sechs Islands: ThemeToggle, MobileNav,
  FeatureExplorer, LiveDemo, FaqAccordion, AnfrageFormular. Alle übrigen
  Sektionen sind statische Astro-Komponenten ohne `client:`-Direktive. Der
  index.astro-Monolith wird dabei in Sektionskomponenten unter
  `website/src/components/` zerlegt.
- **CSP-Externalisierung.** Die Produktiv-CSP
  (`reverse-proxy/nginx.rocks.conf`, jotti.rocks-Block: `script-src 'self'
  'wasm-unsafe-eval'`, kein `unsafe-inline` für Skripte) bleibt unverändert.
  Damit darf kein einziges Inline-Skript ausgeliefert werden:
  `vite.build.assetsInlineLimit: 0` in `website/astro.config.mjs` erzwingt
  externe Skript-Dateien; Pre-Hydration-Skripte (Theme-Init) folgen dem
  bestehenden Muster „echte Datei in `public/` + `<script is:inline src=...>`".
  Astros eingebautes `security.csp`-Feature wird nicht verwendet (es hasht
  Inline-Skripte für einen Astro-verwalteten Header bzw. Meta-Tag; der echte
  Header ist fix und ein Meta-Tag kann ihn nicht lockern). Verifikation je
  Phase mit Client-JS: gebautes `dist/` hinter einem lokalen Static-Server, der
  den CSP-Header wortgleich aus `nginx.rocks.conf` setzt, Konsole ohne
  Verstöße. Invarianten dafür: kein `ViewTransitions`/`ClientRouter`, kein
  `define:vars`, keine `is:inline`-Skripte mit Inline-Body. *Umsetzungsbefund
  Phase 1:* Starlights Doku-Shell liefert `is:inline`-Skripte aus vier
  Komponenten (ThemeProvider, ThemeSelect, Search/Pagefind, SidebarPersister);
  statt vier driftanfälliger Component-Override-Kopien externalisiert eine
  Build-Integration (`website/src/lib/externalize-inline-scripts.ts`,
  kommentar-sicher, nur ausführbare Skripttypen) diese Shell-Skripte nach
  `_astro/` — die automatisierte Form von Fallback-Stufe (1), CSP und
  Semantik unverändert. *Gate-Befund Phase 2:* Astro 6 injiziert seine
  Hydration-Runtime (`getPrescripts()`) hart als Inline-Skripte außerhalb
  der Vite-Asset-Pipeline; `assetsInlineLimit: 0` greift dort nicht. Der
  Externalizer deckt auch diese zwei (deduplizierte) Runtime-Skripte ab und
  ist damit tragende Säule des Islands-Ansatzes — er darf nicht entfernt
  werden. Phase 2 ist das
  Machbarkeits-Gate des Islands-Ansatzes; scheitert der Beweis, gilt die
  Fallback-Leiter: (1) das betroffene Skript als externe Datei nach dem
  `public/`-Muster, (2) gezielter `sha256`-Hash für ein einzelnes stabiles
  Skript in `nginx.rocks.conf` (Präzedenz `'wasm-unsafe-eval'`; bewusste, vom
  Betreiber freizugebende PRD-Abweichung), (3) `'unsafe-inline'` nur als
  ausdrücklich letzte Option.
- **Theme-Mechanismus.** `data-theme="light"|"dark"` auf `<html>`, identisch zu
  Starlight. Persistenz über denselben Speicher wie der Doku-Theme-Schalter
  (Starlight-Default: localStorage-Key `starlight-theme` mit den Werten
  `light`/`dark`/`auto`; Key und Wertformat in Phase 1 an der installierten
  Version verifizieren). Initialisierung vor dem ersten Paint über ein
  externes, klassisch-synchrones Head-Skript: `<script is:inline
  src="/theme-init.js">` ohne `type="module"`, `defer` oder `async`, im
  `<head>` vor allen sichtbaren Elementen — das Modul-Muster von
  `mobile-nav.js` liefe erst nach dem ersten Paint und würde flackern. Der
  Starlight-Wert `auto` und ein fehlender Key bedeuten Systempräferenz.
  Starlights Default-`ThemeProvider` ist
  ein Inline-Skript; falls die Produktiv-CSP es blockiert, wird er per
  Starlight-Component-Override durch eine CSP-konforme externe Variante mit
  gleichem Key und gleicher Semantik ersetzt.
- **Token-Set.** Die Handoff-Tokens ersetzen die bisherigen Markenwerte in
  `website/src/styles/brand.css`: Light-Palette auf `:root`, Dark-Palette auf
  `[data-theme='dark']` (statt `prefers-color-scheme`-Media-Query), dazu die
  acht Spektral-Tokens (`--sp-*`), `--spectral`-Verlauf, Radien, Schatten und
  Typo-Skala aus dem Handoff-README. `website/src/styles/landing.css` mappt die
  Tokens via Tailwind-4-`@theme` auf Utilities; handgeschriebenes CSS nur für
  Keyframes und Spezialfälle (Spektral-Verläufe, Scroll-Reveal, Phone-Rahmen).
  `website/src/styles/starlight.css` bridged dieselben Tokens in die
  `--sl-*`-Variablen der Doku und wechselt dabei auf dieselbe Polarität
  (Light auf `:root`, Dark auf `[data-theme='dark']`) statt der heutigen
  umgekehrten.
- **Fonts.** Space Grotesk (Headings/Wortmarke, 400–700) und Inter (Body/UI,
  400–800) als selbst gehostete Variable-woff2 in `website/public/fonts/`
  (OFL-Lizenzdateien daneben), `@font-face` in brand.css, Preload im Head wie
  bisher. Montserrat (Dateien, Deklarationen, Preloads) entfällt vollständig.
  Die Doku läuft bewusst vollständig in Inter (`--sl-font` ist eine Familie);
  Space Grotesk bleibt der Landing vorbehalten — keine Heading-Brücke in der
  Doku.
- **Reine Logik-Module.** `website/src/lib/live-demo.ts` (Warenkorb,
  Summenbildung, Auto-Demo-Skriptablauf, Stopp bei manueller Interaktion,
  Reset) und `website/src/lib/anfrage-mailto.ts` (Feldwerte zu encodierter
  mailto-URL, Pflichtfeld-Validierung), beide UI-frei und vitest-getestet
  (Prior Art: `website/src/lib/link-rewriter.test.ts`).
- **Screenshot-Skript.** Playwright-Skript im e2e-Paket (nutzt die vorhandene
  Playwright-Installation und `e2e/support/seed.ts — resetAndSeed()`), neues
  Make-Target `website-screenshots`. Läuft gegen den
  `docker-compose.e2e.yml`-Stack (`JOTTI_ENABLE_TEST_API=1`), erzeugt jedes
  Motiv deterministisch benannt in Hell und Dunkel und schreibt nach
  `website/src/assets/screenshots/`. Der zweite Modus (OG-Bild) braucht
  keinen e2e-Stack, sondern das gebaute Website-Artefakt: `make
  website-build`, `dist/` hinter demselben lokalen Static-Server wie bei der
  CSP-Verifikation, Capture 1200×630. Das neue Target kommt in die
  `.PHONY`-Liste des Makefile.
- **ADR.** Die Ablösung von „Spektrum nur dekorativ" wird als
  `docs/adrs/05_spektral-branding-website.md` festgehalten. Nummer 05, weil 04
  bereits durch `docs/adrs/04_warn-bestaetigung.md` belegt ist.
- **Übergangsregel.** Der Umbau läuft sektionsweise auf `main`: alte Sektionen
  bleiben stehen, bis ihr Nachfolger in der jeweiligen Phase landet, und werden
  dann mit entfernt. Phase 1 tauscht dafür nur Token-Werte und Fonts: alte
  Token-Namen (`--brand-spectrum*`) und Komponentenklassen (`.btn`, `.card`,
  `.icon-box`, `.carousel-*`) bleiben — nötigenfalls als Alias auf die neuen
  Tokens — bis zum Rebuild ihrer konsumierenden Sektion erhalten und werden
  mit ihr entfernt. Nav- und CTA-Anker werden erst eingehängt, wenn ihre
  Zielsektion existiert (`#faq` ab Phase 7, `#demo` ab Phase 5); kein
  Zwischenstand hat tote Links. Zwischenstände werden nicht deployt (Deploy
  auf jotti.rocks erfolgt manuell per `make rocks-up` erst nach der
  Gesamtabnahme in Phase 10).

## Inventory

Bestehende Dateien, Muster und Infrastruktur (Pfad + Symbol):

Website-Paket:

- `website/package.json` — Scripts `dev`/`build`/`check` (= `astro check &&
  astro build`)/`test` (= `vitest run`); Deps Astro 6, Starlight 0.40,
  Tailwind 4 (Vite-Plugin), vitest. Kein React.
- `website/astro.config.mjs` — `site`, Starlight-Integration (customCss
  `./src/styles/starlight.css`, Sidebar), Remark-Plugin `remarkDocLinks`,
  externes Docs-Glob über `../docs`. Hier kommen React-Integration und
  `assetsInlineLimit` dazu.
- `website/src/pages/index.astro` — heutige Landing als Monolith (~1580
  Zeilen), alle Sektionen inline: Hero, `#fuer-wen`, `#versprechen` (Nutzen +
  Vergleichstabelle), `#features` (4 Gruppen), `#einblicke` (Carousel),
  `#so-funktionierts` (3 Schritte), `#service`, `#compliance`/`#technik`,
  `#bereit`.
- `website/src/layouts/Landing.astro` — Head (Title/Canonical/OG/Twitter,
  Favicons, Font-Preloads), Header mit Beta-Banner und Nav, Footer, lädt
  `public/mobile-nav.js`.
- `website/src/styles/brand.css` — Montserrat-`@font-face`, `--brand-green`,
  `--brand-spectrum` (heute nur Haarlinien/Wash, „nur dekorativ").
- `website/src/styles/landing.css` — Tailwind-Entry, semantische Tokens
  (Dark via Media-Query), `@theme`-Mapping, Button-/Card-/Carousel-Klassen.
- `website/src/styles/starlight.css` — Bridge `--brand-*` zu `--sl-*`
  (`--sl-font`, Akzentfarben, `[data-theme='light']`-Selektor).
- `website/public/mobile-nav.js`, `website/public/carousel.js` — bewusst
  externe Skripte (Astro inlinet kleine Skripte, die CSP blockt das); Muster
  für das Theme-Init-Skript.
- `website/src/lib/link-rewriter.ts` + `link-rewriter.test.ts` — Prior Art
  für reine Module mit vitest. `website/src/lib/links.ts` — `leitfadenUrl`,
  `githubUrl`.
- `website/src/assets/screenshots/` — 9 veraltete Motive (tisch-bestellen,
  tisch-zahlung, tisch-stornierung, direktverkauf, service-tische-uebersicht,
  produkte-verwalten, benutzer-verwalten, kasse-geldtransit,
  auswertung-historisch). `website/src/assets/jotti-bon.webp` — heutiges
  OG-Bild. `website/src/assets/jotti-symbol.png` — Logo-Bildmarke.
- `website/nginx.conf`, `website/Dockerfile` — Container-Serve (Build-Kontext
  Repo-Root); setzt keine CSP, die kommt vom Reverse-Proxy.

Design-Handoff (`docs/prds/design_handoff_jotti_website/`):

- `README.md` — Tokens (Light/Dark, `--sp-*`, `--spectral`), Typo-Skala,
  Radien, Schatten, Sektionsliste, Copy- und Asset-Vorgaben.
- `jotti Website.dc.html` — Prototyp: Sektionsfolge und -inhalte, Icon-Sprite
  (Lucide-nah), CSS-Phone-Rahmen (kein Bild-Asset), Demo-Daten und -Skript
  (Menü Bier 0,5l 400 / 0,3l 300, Weinschorle 350, Bratwurst 350, Pommes 300
  Cent; Ablauf Bratwurst, 2× Bier 0,5l, Pommes, Kassieren, Reset; Endsumme
  14,50 €), FAQ (7 Items, single-open), Formularfelder (`verein`, `name`,
  `email`, `art`-Select, `message`). Der clientseitige View-Wechsel und
  `support.js` werden nicht übernommen; Google-Fonts-Links ebenso nicht.
  Achtung: Der Prototyp hat keine mobile Navigation (Links unter 860px nur
  ausgeblendet), die PRD verlangt ein Burger-Menü.
- `assets/` — Original-Logos (`jotti-symbol.png`, Full-/Icon-Varianten
  hell/dunkel, Favicon-Größe); keine Nachbauten erlaubt.
- `screenshots/` — 14 Referenz-Screenshots (10 hell, 4 dunkel) als
  Pixel-Referenz.

Infrastruktur:

- `reverse-proxy/nginx.rocks.conf` — Produktiv-CSP des jotti.rocks-Blocks
  (`'wasm-unsafe-eval'`-Ausnahme für die Pagefind-Suche der Doku). Bleibt
  unverändert, dient nur der Verifikation.
- `e2e/playwright.config.ts`, `e2e/support/seed.ts — resetAndSeed()` —
  Playwright-Suite mit Reset-und-Seed gegen `POST /api/test/reset-and-seed`;
  `e2e/support/anmelden.ts` — Login-Helper.
- `docker-compose.e2e.yml` — Stack aus echten Images mit
  `JOTTI_ENABLE_TEST_API=1`, HTTP-only, Port über `E2E_HTTP_PORT`.
- `backend/seed/szenario.go — demoSzenario()` — „3-Tage-Sommerfest TSV
  Musterstadt e.V.": 19 Produkte / 49 Varianten, 22 Tische, 10 Benutzer;
  realistische Daten für alle Screenshot-Motive.
- `Makefile` — `website-check` (= `pnpm test && pnpm check`), `website-dev`,
  `website-build`; nicht Teil von `make check`/`verify`. CI-Job `website-ci`
  läuft bei Änderungen unter `website/**`.
- `screenshots/` (Repo-Root) — 18 veraltete Roh-PNGs + README, von nichts
  referenziert, zur Löschung freigegeben.
- `docs/adrs/` — 01–03 vergeben, 04 reserviert (siehe Architectural
  decisions).

Mapping alte zu neuen Sektionen:

| Heute (`index.astro`) | Neu | Phase |
| --- | --- | --- |
| Hero | Hero (Handoff, echter Phone-Screenshot) | 3 |
| `#fuer-wen` | `#fuerwen` | 6 |
| `#versprechen` (Nutzen + Vergleichstabelle) | „Kostenlos"-Aussagen: Hero, `#preis`, FAQ; Alternativen-Vergleich wird FAQ-Item („Warum nicht Excel oder eine Profi-Kasse?") | 6/7 |
| `#features` (4 Gruppen) | `#features` (Explorer, 6 Bereiche) | 4 |
| `#einblicke` (Carousel) | Galerie in `#screenshots` | 9 |
| `#so-funktionierts` (3 Schritte) | `#ablauf` (4 Schritte) | 6 |
| `#service` | FAQ-Item + Erwähnung auf `/fuer-vereine` | 7/8 |
| `#compliance` + `#technik` | `#sicherheit` + `#preis` | 6 |
| `#bereit` | `#download` | 7 |
| Beta-Banner im Header | Hero-Badge + Hinweis-Pill im Download-Bereich | 2/3/7 |

## Resolved decisions

Aus der PRD (dort bereits mit dem Betreiber geklärt): React-Islands, echte
Screenshots überall außer Live-Demo, mailto-Formular ohne Backend, ehrliche
Download-Copy, Spektral-Einsatz voll (mit ADR), Starlight-Persistenz statt
eigenem Theme-Key, vier Routen, Rechtsseiten-Texte liefert der Betreiber,
keine E2E-/Visual-Tests für die Website.

Während der Planung entschieden (Annahmen, von der PRD gedeckt, aber dort
nicht festgelegt):

- **Assumption (ADR-Nummer):** 05 statt „nächste freie", weil 04 im aktiven
  UI-Politur-Plan reserviert ist.
- **Assumption (OG-Bild):** wird als 1200×630-Capture des neuen Hero (hell)
  vom Screenshot-Skript miterzeugt (zweiter Skript-Modus gegen das gebaute
  Website-Artefakt), statt eines manuell gestalteten Assets. Reproduzierbar
  wie alle anderen Bilder.
- **Assumption (Galerie-Motive):** Die Galerie übernimmt die neun heutigen
  Motive (PRD: „angelehnt an die heutige Auswahl"); jedes Motiv wird hell und
  dunkel aufgenommen und theme-passend angezeigt.
- **Assumption (Skript-Standort):** Screenshot-Skript im e2e-Paket, nicht im
  Website-Paket, um Playwright-Installation und Seed-/Login-Helper
  wiederzuverwenden.
- **Assumption (Rechtsseiten):** Seiten kommen mit fertiger Struktur und
  deutlich markierten Platzhaltertexten; die finalen Texte setzt der Betreiber
  vor der Abnahme ein.
- **Assumption (Header-CTA-Übergang):** Der CTA „Für Vereine" zeigt bis
  Phase 8 übergangsweise auf `#download`, danach auf `/fuer-vereine` (kein
  toter Link in Zwischenständen; der Anker existiert ab Phase 2 durch das
  Retagging von `#bereit` zu `#download`).
- Mobile Navigation wird als Burger-Menü neu gestaltet (Interaktionsmuster des
  heutigen `mobile-nav.js`: aria-expanded, Escape, Schließen bei Link-Klick),
  da der Prototyp hier nichts vorgibt.

Im Multi-Experten-Review (2026-07-13) mit dem Betreiber entschieden:

- **CSP-Fallback-Leiter** statt Lockerung von vornherein (siehe
  „CSP-Externalisierung"); gilt analog für den ThemeProvider-Override in
  Phase 1.
- **Alternativen-Vergleich wird FAQ-Item** („Warum nicht Excel oder eine
  Profi-Kasse?", Kernaussage der heutigen Vergleichstabelle) — neun
  FAQ-Items gesamt (Phase 7).
- **Alle sechs Islands bleiben React** (inkl. ThemeToggle/MobileNav;
  PRD-Entscheidung bestätigt).
- **Doku-Typografie vollständig Inter**, keine Space-Grotesk-Heading-Brücke
  (siehe „Fonts").
- **Beta-Hinweis auch auf `/fuer-vereine`** nahe dem Formular — dritter
  Träger der Beta-Kommunikation (Phase 8).
- **Icons über die offiziellen Lucide-Pakete** `@lucide/astro` +
  `lucide-react`, gleiche Version (siehe „React-Islands").
- **A11y-Ergänzungen:** Skip-Link (Phase 2), Alt-Texte für Screenshots
  (Phase 3/9), programmatische Formular-Fehler (Phase 8), Eyebrow-Kontrast
  namentlich in der Kontrastprüfung (Phase 10).
- **Datenschutz-Platzhalter-Gliederung erweitert** um Server-Logs/Hosting,
  E-Mail-Provider, Betroffenenrechte und Speicherdauer (Phase 7).
- **Explorer-Copy gegen Anforderungs-Status geprüft:** „Abrechnung pro
  Tisch" entfällt (R-03 per ADR 02 ersatzlos entfernt); kein „In
  Entwicklung" auf der Website (Phase 4).
- **Anker-Links erst mit Zielsektion:** `#faq`-Nav-Link ab Phase 7,
  `#demo`-CTA ab Phase 5, `#bereit` wird schon in Phase 2 zu `#download`
  retaggt (siehe „Übergangsregel").

## Open questions / Risks

- **Starlight-Theme unter CSP:** Starlights Default-`ThemeProvider` ist ein
  Inline-Skript und ist unter der Produktiv-CSP möglicherweise heute schon
  blockiert. Phase 1 prüft das am gebauten Artefakt und ersetzt ihn bei Bedarf
  per Component-Override (CSP-konform, gleicher Speicher-Key; falls auch
  `ThemeSelect` betroffen ist, behält der Override Key und Semantik bei).
  Greift auch der Override nicht, gilt die Fallback-Leiter aus
  „CSP-Externalisierung".
- **Island-Hydration unter CSP:** Mit `assetsInlineLimit: 0` liefert Astro
  alle Skripte extern aus; ob das für die komplette Hydration-Kette reicht,
  beweist Phase 2 am gebauten Artefakt mit echtem CSP-Header, bevor weitere
  Islands entstehen. Scheitert der Beweis, greift die Fallback-Leiter aus
  „CSP-Externalisierung".
- **Scroll-getriebene Animationen:** `animation-timeline: view()` ist nicht in
  allen Browsern verfügbar. Scroll-Reveal wird als Progressive Enhancement
  umgesetzt (`@supports`); ohne Unterstützung bleiben Sektionen schlicht
  sichtbar.
- **Dunkle App-Screenshots:** Erwartung ist, dass die App
  `prefers-color-scheme` folgt und Playwrights `colorScheme`-Option genügt.
  Phase 9 verifiziert das zuerst; falls die App einen eigenen
  Theme-Schalter-State braucht, steuert das Skript diesen an.
- **Farbkontraste:** Die Handoff-Palette (z. B. `--muted-fg` auf `--muted`,
  Spektral-Akzente auf Karten) wird bei der Umsetzung geprüft (PRD
  Barrierefreiheit); nötige Korrekturen erfolgen token-seitig, nicht als
  Einzelfall-Hacks. Namentlich zu prüfen: der Eyebrow (`--sp-teal` `#1f9b8a`
  auf Weiß bei 12,5px, überschlägig ~3,4:1 — unter AA 4,5:1).
  **Gelöst (Phase 10, token-seitig):** Drei AA-sichere Text-Varianten in
  `brand.css` für Light — `--sp-teal-text` `#177a6b` (Eyebrow),
  `--sp-red-text` `#c03f1c` (Formular-Fehler) und `--sp-orange-text` `#985b17`
  (Platzhalter-Label, bei der Messung ebenfalls unter AA gefunden). Im Dunkeln
  erben sie die vollen Töne (dort schon AA-sicher). Die vollen `--sp-*`-Werte
  bleiben für Verlauf, Streifen, Icons und Akzentflächen unverändert (Non-Text
  3:1), damit die Spektral-Identität und die `--spectral`-Verläufe erhalten
  bleiben. Jede Text-Verwendung jedes Spektral-Tokens wurde in beiden Themes im
  Browser gemessen: kleiner Text durchgehend ≥ 4,5:1, Icons/dekorative Marken
  ≥ 3:1.
- **Rechtsseiten-Texte** sind eine Bring-Schuld des Betreibers und blockieren
  die Gesamtabnahme (Phase 10), nicht die Umsetzung.

---

## Phase 1: Token-Fundament, Fonts und Theme-Mechanik (Landing + Doku)

**User stories**: 12 (Theme-Wahl gespeichert), 13 (eine Marke über Landing und
Doku)

### Context

- `website/src/styles/brand.css` — Token- und Font-Quelle, wird ersetzt.
- `website/src/styles/landing.css` — semantische Tokens + `@theme`-Mapping;
  Dark-Werte wandern von Media-Query auf `[data-theme='dark']`.
- `website/src/styles/starlight.css` — Bridge in `--sl-*`-Variablen.
- `website/src/layouts/Landing.astro` — Font-Preloads, Head; hier hängt sich
  das Theme-Init-Skript ein.
- `website/public/mobile-nav.js` — Muster für externe `public/`-Skripte.
- Starlight-`ThemeProvider`/`ThemeSelect` in `node_modules/@astrojs/starlight`
  — tatsächlichen Speicher-Key, Wertformat und Inline-Skript-Verhalten der
  installierten Version nachlesen.
- Handoff-`README.md` — Token-Werte, Typo-Skala, Radien, Schatten.

### What to build

Das gemeinsame Fundament beider Seitenteile: Handoff-Tokens ersetzen die
bisherigen Markenwerte in brand.css (Light auf `:root`, Dark auf
`[data-theme='dark']`, Spektral-Tokens und -Verläufe inklusive), Space Grotesk
und Inter ersetzen Montserrat als selbst gehostete Variable-Fonts mit Preload.
Die bestehenden Landing-Sektionen werden nur so weit angefasst, dass ihre
Dark-Styles am neuen `data-theme`-Attribut hängen (die optische Neugestaltung
folgt in späteren Phasen); alte Token-Namen und Komponentenklassen bleiben
gemäß Übergangsregel intakt. Ein externes Pre-Paint-Skript in `public/`
(klassisch-synchron im Head, siehe „Theme-Mechanismus"; `auto` und fehlender
Key bedeuten Systempräferenz) liest den Starlight-Speicher und setzt
`data-theme` auf `<html>`, bevor gerendert wird. Die Doku übernimmt Fonts und Farbwelt über die
starlight.css-Bridge. Dabei wird geprüft, ob Starlights Theme-Initialisierung
unter der Produktiv-CSP funktioniert; falls nicht, wird der `ThemeProvider`
per Starlight-Component-Override CSP-konform ersetzt (gleicher Key, gleiche
Semantik). Noch kein Schalter auf der Landing, die Theme-Wahl lässt sich über
den Doku-Schalter demonstrieren.

### Acceptance criteria

- [x] Landing und Doku laufen in Space Grotesk/Inter und der neuen Farbwelt,
      hell und dunkel; keine Montserrat-Reste (Dateien, `@font-face`,
      Preloads).
- [x] Die Theme-Wahl im Doku-Schalter wirkt nach Navigation auch auf die
      Landing (gemeinsamer Speicher, `data-theme` auf `<html>`), ohne
      Aufblitzen des falschen Themes beim Laden.
- [x] Spektral-Tokens (`--sp-*`, `--spectral`) stehen bereit und stimmen mit
      dem Handoff-README überein.
- [x] Alt-Sektionen sind optisch intakt: alte Token-Namen und
      Komponentenklassen wirken weiter (nötigenfalls als Alias).
- [x] Gebautes `dist/` hinter einem lokalen Static-Server mit dem
      CSP-Header aus `nginx.rocks.conf`: keine CSP-Verstöße in der Konsole auf
      Landing- und Doku-Seiten, Theme-Init eingeschlossen.
- [x] `make website-check` grün.

---

## Phase 2: React-Islands-Grundstein: neuer Header mit ThemeToggle und MobileNav

**User stories**: 12 (Umschalter), 14 (mobile Navigation)

### Context

- `website/astro.config.mjs` — React-Integration und
  `vite.build.assetsInlineLimit: 0` kommen hier hinein.
- `website/src/layouts/Landing.astro` — heutiger Header (Beta-Banner, Nav,
  Mobile-Toggle), wird ersetzt.
- `website/public/mobile-nav.js` — heutiges Interaktionsmuster
  (aria-expanded, Escape, Schließen bei Link-Klick); entfällt zugunsten der
  Island.
- Handoff-Prototyp — Header-Layout (sticky, Blur, Logo + Wortmarke,
  Anker-Links, Theme-Button, CTA); Breakpoint 860px.

### What to build

Die React-Integration wird eingerichtet und mit den zwei kleinsten Islands
bewiesen — diese Phase ist das Machbarkeits-Gate für den Islands-Ansatz
(Fallback-Leiter siehe „CSP-Externalisierung"). Der Header wird nach Handoff
neu gebaut (Bildmarke `jotti-symbol.png` plus Wortmarke in Space Grotesk,
Anker-Links `#features`, `#ablauf`, `#fuerwen`, `#sicherheit`,
Leitfaden-Link zur Doku, CTA „Für Vereine", vorerst auf `#download`; der
`#faq`-Link folgt erst in Phase 7 mit seiner Sektion). Ein Skip-Link „Zum
Inhalt springen" steht vor dem Header. Der Beta-Banner entfällt. ThemeToggle
(zugängliches Label, schreibt in den Starlight-Speicher aus Phase 1) und
MobileNav (Burger-Menü unter 860px, Tastatur- und Escape-Verhalten wie das
heutige `mobile-nav.js`; Datei und Script-Tag entfallen mit der Island) sind
React-Islands. Damit die Anker-Links sofort funktionieren, erhalten die
bestehenden alten Sektionen in dieser Phase die neuen Anker-IDs (`#fuer-wen`
wird `#fuerwen`, `#so-funktionierts` wird `#ablauf`, `#compliance` wird
`#sicherheit`, `#bereit` wird `#download`); die Sektionen selbst werden
erst in ihren jeweiligen Phasen neu gebaut.

### Acceptance criteria

- [x] Header entspricht dem Handoff (sticky, Blur, Anker-Nav ohne den erst in
      Phase 7 folgenden `#faq`-Link, Leitfaden-Link, CTA); Beta-Banner ist
      entfernt; Skip-Link vorhanden und bei Fokus sichtbar.
- [x] ThemeToggle wechselt Hell/Dunkel sofort, persistiert und bleibt mit dem
      Doku-Schalter konsistent (beide Richtungen).
- [x] Mobile Navigation als Burger-Menü unter 860px, per Tastatur bedienbar
      (Fokus, Escape, Schließen bei Link-Klick), mit korrektem
      `aria-expanded`.
- [x] Das gebaute `dist/`-HTML enthält keine Inline-Skripte; beide Islands
      hydrieren unter dem Produktiv-CSP-Header ohne Konsolen-Verstöße.
- [x] `public/mobile-nav.js` und sein Script-Tag in `Landing.astro` sind
      entfernt.
- [x] `make website-check` grün.

---

## Phase 3: Hero und Spektral-Grundelemente, ADR

**User stories**: 1 (Hero-Botschaft), 15 (reduzierte Bewegung)

### Context

- Handoff-Prototyp — Hero-Layout (Eyebrow, H1 mit animiertem Spektral-Verlauf,
  Sub-Copy, Badges/Bullets, zwei CTAs, CSS-Phone-Rahmen mit Notch), weiche
  Blobs, `.reveal`-Keyframes (`animation-timeline: view()`),
  `prefers-reduced-motion`-Neutralisierung.
- `website/src/pages/index.astro` — heutiger Hero, wird ersetzt.
- `website/src/assets/screenshots/tisch-bestellen.png` — Platzhalter für das
  Hero-Phone, bis Phase 9 die neuen Aufnahmen liefert.
- `docs/adrs/` — README-Index, Format der ADRs 01–03.

### What to build

Der neue Hero nach Handoff: Eyebrow, H1 mit animiertem Spektral-Textverlauf,
ehrliche Sub-Copy, Beta-Badge (übernimmt die Beta-Kommunikation vom entfernten
Banner), CTAs (der Demo-CTA auf `#demo` wird erst in Phase 5 mit seiner
Sektion eingehängt), daneben ein CSS-Telefon-Rahmen mit echtem Screenshot der
Bestellansicht mit beschreibendem Alt-Text (theme-abhängig hell/dunkel; bis
Phase 9 dient der heutige Screenshot als helle Übergangsvariante). Dazu die wiederverwendbaren
Spektral-Grundelemente für alle Folgephasen: weiche Hintergrund-Blobs und
Scroll-Reveal als Progressive Enhancement (`@supports
(animation-timeline: view())`, sonst schlicht sichtbar).
`prefers-reduced-motion` neutralisiert Verlaufs-Animation, Blobs und Reveal
global. Der ADR `docs/adrs/05_spektral-branding-website.md` hält die Ablösung
von „Spektrum nur dekorativ" fest (Kontext, Entscheidung, Konsequenzen,
Verweis auf die PRD; „Regenbogen" bleibt als Wort verboten) und wird im
ADR-Index verlinkt.

### Acceptance criteria

- [x] Hero entspricht dem Handoff in Hell und Dunkel (Referenz: Bundle
      `screenshots/01-light.png`/`01-dark.png`); Phone-Rahmen ist reines CSS.
      Der Bildinhalt des Phones bleibt bis Phase 9 die helle
      Übergangsvariante und wird erst dort dunkel abgenommen; der Screenshot
      trägt einen beschreibenden Alt-Text.
- [x] H1-Verlauf, Blobs und Scroll-Reveal laufen dort, wo der Browser sie
      unterstützt; ohne Unterstützung und unter `prefers-reduced-motion` sind
      alle Inhalte statisch und vollständig sichtbar.
- [x] Beta-Kommunikation über das Hero-Badge vorhanden.
- [x] ADR 05 liegt vor und ist im ADR-Index verlinkt.
- [x] `make website-check` grün; CSP-Durchlauf weiterhin ohne Verstöße.

---

## Phase 4: Feature-Explorer (`#features`)

**User stories**: 2 (interaktiver Explorer)

### Context

- Handoff-Prototyp — sechs Bereiche (Bestellung, Zahlung, Direktverkauf,
  Küche/Ausgabe, Kassenführung, Auswertung) mit je einer Spektral-Akzentfarbe,
  Icon, Beschreibung und drei Punkten; aktives Tile mit farbigem Rand und
  sticky Detail-Karte rechts.
- `website/src/pages/index.astro` — heutige `#features`-Sektion (4 Gruppen),
  entfällt mit dieser Phase.
- Handoff-README — Icon-Zuordnung (Lucide-nah; Bestellung = Beleg, Zahlung =
  Geldbörse und ausdrücklich kein Kartenterminal, Direktverkauf =
  Einkaufstasche).

### What to build

Die FeatureExplorer-Island ersetzt die heutige Funktionen-Sektion: sechs
antippbare Bereichs-Tiles mit Spektral-Akzent und Lucide-Icons (via
`lucide-react`) gemäß Handoff-Zuordnung, ein Bereich aktiv, Detail-Karte mit
Icon, Titel, Beschreibung und drei Punkten (Copy aus dem Prototyp). Die
Prototyp-Copy wird dabei gegen den Anforderungs-Status
(`docs/anforderungen.md`) geprüft: im Reporting-Tile entfällt „Abrechnung
pro Tisch"/„Pro Tisch & Person" (R-03 per ADR 02 ersatzlos entfernt)
zugunsten von „pro Servicekraft"; „In Entwicklung"-Markierungen gibt es auf
der neuen Website nicht (Beta-Kommunikation nur über Hero-Badge,
Download-Pill und den Hinweis auf `/fuer-vereine`). Tastaturbedienung und
ARIA-Semantik nach dem einschlägigen WAI-ARIA-Pattern (Tabs), sichtbarer
Fokus. Auf Mobil einspaltig mit Detail-Karte unterhalb der Tiles.

### Acceptance criteria

- [x] Sechs Bereiche mit korrekten Akzentfarben, Icons (Bedeutungen gemäß
      Handoff erhalten) und Copy; Referenz `02-light.png`/`02-dark.png`.
- [x] Explorer-Copy gegen `docs/anforderungen.md` geprüft: kein Verweis auf
      entfernte Features (insb. keine „Abrechnung pro Tisch"), kein „In
      Entwicklung".
- [x] Auswahl per Maus, Touch und Tastatur (Pfeiltasten/Tab gemäß
      ARIA-Pattern); Screenreader-Semantik korrekt.
- [x] Mobil (390px) einspaltig ohne horizontalen Überlauf.
- [x] Alte Funktionen-Sektion entfernt; Anker `#features` zeigt auf die neue.
- [x] `make website-check` grün; Island hydriert unter Produktiv-CSP.

---

## Phase 5: Live-Demo (`#demo`)

**User stories**: 3 (selbst ausprobieren oder zusehen), 15 (reduzierte
Bewegung)

### Context

- Handoff-Prototyp — Demo-Menü und -Preise (Bier 0,5l 400 / 0,3l 300,
  Weinschorle 0,25l 350, Bratwurst im Weck 350, Pommes Portion 300 Cent),
  Auto-Skript (Bratwurst, Bier 0,5l, Bier 0,5l, Pommes, Kassieren, Reset;
  Endsumme 14,50 €), Timings, IntersectionObserver-Start ab 25% Sichtbarkeit,
  Stopp bei manueller Interaktion, Kassieren-Overlay mit Erfolgsanimation,
  „Demo neu abspielen".
- `website/src/lib/link-rewriter.test.ts` — Prior Art für Modul-Tests.
- Phase 3 — CSS-Phone-Rahmen, wird hier wiederverwendet.

### What to build

Das UI-freie Modul `website/src/lib/live-demo.ts` kapselt Warenkorb (Menge
je Variante erhöhen/verringern), Summenbildung in Cent, den Auto-Demo-Ablauf
als deterministische Schrittfolge, den permanenten Stopp bei manueller
Interaktion und Reset; vitest-Tests decken diese Verhalten ab (Eingabe zu
Ergebnis, keine DOM-Details). Die LiveDemo-Island rendert das Modul im
Phone-Rahmen: Produktliste mit Mengen-Steppern, live mitwachsende Summe,
Kassieren-Overlay mit Erfolgsanimation, Reset-Button. Auto-Demo startet
einmalig beim Scrollen in den Viewport und läuft mit den Handoff-Timings;
unter `prefers-reduced-motion` startet keine Auto-Demo (manuelle Bedienung
bleibt voll funktionsfähig). Dies ist der einzige UI-Nachbau der Seite.

### Acceptance criteria

- [x] `live-demo.ts` ist UI-frei; vitest-Tests decken Mengenänderung,
      Summenbildung, Auto-Skriptablauf, Stopp bei manueller Interaktion und
      Reset ab.
- [x] Auto-Demo startet beim Hereinscrollen, durchläuft das Handoff-Skript
      (Endsumme 14,50 €) und stoppt dauerhaft bei manueller Interaktion;
      „Demo neu abspielen" startet neu.
- [x] Kassieren-Overlay mit Erfolgsanimation und Reset funktionieren; Beträge
      sind durchgehend in Cent gerechnet und korrekt formatiert.
- [x] Unter `prefers-reduced-motion` keine Auto-Demo, keine Animationen;
      manuelle Bedienung funktioniert.
- [x] `make website-check` grün; Island hydriert unter Produktiv-CSP.

---

## Phase 6: Statische Sektionen: Ablauf, Für wen, Preis, Sicherheit

**User stories**: 4 (Ablauf in vier Schritten), 5 (geeignet/nicht geeignet),
7 (kostenlos + Fremdkosten), 8 (fiskalische Bausteine)

### Context

- Handoff-Prototyp — `#ablauf` (vier nummerierte Schritte mit
  Spektral-Akzenten), `#fuerwen` (geeignet/nicht geeignet), `#preis`
  (Preis-Karte „Kostenlos. Punkt." mit Punkteliste und Fremdkosten-Hinweisen),
  `#sicherheit` (Compliance-Karten mit Spektral-Streifen).
- `website/src/pages/index.astro` — heutige Sektionen `#fuer-wen`,
  `#versprechen`, `#so-funktionierts`, `#compliance`/`#technik`; entfallen
  mit dieser Phase.
- PRD „Erhaltene Inhalte" — Preis-Karte erhält den Punkt „Keine Werbung, kein
  Tracking"; laufende Fremdkosten (Cloud-TSE, optional Server) bleiben ehrlich
  benannt (Zahlen aus der heutigen `#technik`-Kostenliste übernehmen).

### What to build

Vier statische Astro-Sektionskomponenten in Handoff-Layout und -Copy: der
Ablauf vom Einrichten bis zum Z-Bon in vier Schritten; „Für wen" mit ehrlicher
Geeignet/Nicht-geeignet-Gegenüberstellung; die Preis-Sektion mit der
„Kostenlos. Punkt."-Karte, ergänzt um „Keine Werbung, kein Tracking" und die
laufenden Fremdkosten; „Sicherheit und Compliance" mit den fiskalischen
Bausteinen (TSE, Belegausgabe, GoBD-Journal, DSFinV-K-Export, Rollen,
Onboarding) als Karten mit Spektral-Akzenten. Die abgelösten Alt-Sektionen
(inklusive Vergleichstabelle) entfallen; die Kernaussage des
Alternativen-Vergleichs lebt als FAQ-Item in Phase 7 weiter. `#service`
bleibt bis Phase 7 stehen, damit das Service-Angebot nie unauffindbar ist.

### Acceptance criteria

- [x] Vier Sektionen entsprechen dem Handoff (Referenzen `04-light.png`,
      `05-light.png`, `07-light.png`, `08-light.png`), hell und dunkel, mobil
      einspaltig.
- [x] Preis-Karte enthält „Keine Werbung, kein Tracking" und die ehrlichen
      Fremdkosten; keine Kostenaussage geht verloren.
- [x] Alt-Sektionen `#fuer-wen`, `#versprechen`, `#so-funktionierts`,
      `#compliance`/`#technik` sind entfernt; Nav-Anker zeigen auf die neuen
      Sektionen.
- [x] Scroll-Reveal auf den neuen Sektionen; unter `prefers-reduced-motion`
      alles statisch sichtbar.
- [x] `make website-check` grün.

---

## Phase 7: FAQ, Download-CTA, Footer und Rechtsseiten

**User stories**: 9 (FAQ), 11 (ehrlicher Download), 16 (Impressum/Datenschutz
im Footer)

### Context

- Handoff-Prototyp — FAQ (7 Items, single-open, Plus-Icon rotiert zu ×),
  Download-CTA (große Karte mit Hinweis-Pill), Footer (Spalten Produkt,
  Ressourcen, Rechtliches).
- PRD — zusätzliches FAQ-Item „Gibt es Unterstützung beim Einrichten?"
  (bezahlter Service auf Anfrage); Download-Copy: ZIP mit Starter, benötigt
  Docker Desktop, Leitfaden führt durch, Links auf Leitfaden und Quellcode,
  kein „.exe"/„5 Minuten"-Versprechen; Rechtsseiten-Texte liefert der
  Betreiber.
- `website/src/pages/index.astro` — `#service` und `#bereit`, entfallen mit
  dieser Phase. `website/src/layouts/Landing.astro` — heutiger Footer.
- `website/src/lib/links.ts` — `leitfadenUrl`, `githubUrl` für die
  Download- und Footer-Links.

### What to build

Die FaqAccordion-Island mit den sieben Handoff-Items plus dem Service-Item
und dem Alternativen-Vergleichs-Item („Warum nicht Excel oder eine
Profi-Kasse?", Kernaussage der heutigen Vergleichstabelle) — neun gesamt,
single-open, mit korrekter Accordion-Semantik (Button/`aria-expanded`,
Tastaturbedienung). Mit der neuen Sektion wird der `#faq`-Link in die
Header-Navigation aufgenommen (siehe Phase 2). Der Download-Bereich nach
Handoff-Layout mit der ehrlichen Copy und der Hinweis-Pill (zweiter Träger
der Beta-Kommunikation), Links auf Leitfaden und Quellcode. Der neue Footer
mit drei Spalten; unter Rechtliches Links auf `/impressum` und
`/datenschutz`, die heutigen Links auf Nutzungsbedingungen (`TERMS.md`) und
Lizenz (`LICENSE`) bleiben erhalten. Die beiden Rechtsseiten entstehen als
statische Astro-Seiten mit eigener Meta (Title, Description, Canonical,
noindex nicht nötig) auf Basis eines schlichten Prosa-Layouts: Impressum
(Anbieterkennzeichnung), Datenschutzerklärung (trackingfreie statische
Seite, funktionale localStorage-Theme-Speicherung, Verarbeitung von
E-Mail-Anfragen inklusive E-Mail-Provider des Betreibers,
Server-Logs/Hosting, Betroffenenrechte, Speicherdauer), beide mit deutlich
markierten Platzhaltern für die Betreiber-Texte. Beim Entfernen von
`#service` wird die Betreiber-Mailadresse als Konstante in
`website/src/lib/links.ts` gesichert (Empfänger für Phase 8). `#service`
und `#bereit` entfallen.

### Acceptance criteria

- [x] FAQ: neun Items inklusive Service- und Vergleichs-Item, single-open,
      per Tastatur bedienbar, korrekte ARIA-Semantik; Referenz
      `09-light.png`/`04-dark.png`; `#faq`-Link ist im Header eingehängt.
- [x] Download-Bereich: ehrliche Copy (ZIP, Starter, Docker Desktop,
      Leitfaden), Links auf Leitfaden und Quellcode, Hinweis-Pill; die Wörter
      „.exe" und „5 Minuten" kommen nicht vor.
- [x] Footer mit drei Spalten; `/impressum` und `/datenschutz` erreichbar,
      mit eigener Meta und markierten Platzhaltertexten;
      Nutzungsbedingungen- und Lizenz-Links erhalten. Die
      Datenschutz-Platzhalter decken auch Server-Logs/Hosting,
      E-Mail-Provider, Betroffenenrechte und Speicherdauer ab.
- [x] Die Betreiber-Mailadresse liegt als Konstante in
      `website/src/lib/links.ts`.
- [x] Alt-Sektionen `#service` und `#bereit` entfernt; das Service-Angebot
      ist im FAQ-Item abgedeckt.
- [x] `make website-check` grün; FaqAccordion hydriert unter Produktiv-CSP.

---

## Phase 8: „Für Vereine" (`/fuer-vereine`)

**User stories**: 10 (Formular öffnet vorbefüllten E-Mail-Entwurf)

### Context

- Handoff-Prototyp — Formular „Nutzungsvereinbarung anfragen": Felder
  `verein` (Pflicht), `name` (Pflicht), `email` (Pflicht, type=email), `art`
  (Select mit den vollen Prototyp-Labels: „Eingetragener Verein (e.V.)",
  „Gemeinnützige Stiftung", „NGO / NPO", „Sonstige gemeinnützige
  Organisation"), `message` (optional); Erfolgs-State mit Häkchen-Animation
  und zwei CTAs.
- PRD — Absenden validiert clientseitig und öffnet einen vorbefüllten
  mailto-Entwurf an die Betreiber-Adresse; ehrlicher Bestätigungs-State
  („Entwurf geöffnet, muss noch gesendet werden"); E-Mail-Adresse sichtbar als
  Fallback; Service-Angebot wird erwähnt.
- `website/src/lib/links.ts` — die Betreiber-Mailadresse liegt hier seit
  Phase 7 als Konstante (dort beim Entfernen der `#service`-Sektion
  gesichert).
- `website/src/lib/link-rewriter.test.ts` — Prior Art für Modul-Tests.

### What to build

Das UI-freie Modul `website/src/lib/anfrage-mailto.ts` baut aus den
Feldwerten eine korrekt encodierte mailto-URL (Empfänger ist die
Betreiber-Adress-Konstante aus `links.ts`, Betreff und strukturierter Body
aus den Feldern) und validiert die Pflichtfelder; vitest-Tests decken
Encoding (Umlaute, Zeilenumbrüche, Sonderzeichen) und Validierung ab. Die
Seite `/fuer-vereine` (eigene Meta) erklärt die Nutzungsanfrage, erwähnt das
bezahlte Service-Angebot, trägt nahe dem Formular einen knappen Beta-Hinweis
(dritter Träger der Beta-Kommunikation) und zeigt die E-Mail-Adresse
sichtbar als Fallback. Die AnfrageFormular-Island rendert die
Handoff-Felder, zeigt Validierungsfehler inline und programmatisch
verknüpft (`aria-invalid`, `aria-describedby`, Ankündigung per Live-Region),
öffnet bei gültigem Absenden den mailto-Entwurf per JS-Navigation (kein
natives `<form action="mailto:">` — das blockt `form-action 'self'` der
Produktiv-CSP) und wechselt in den ehrlichen Bestätigungs-State (Entwurf
geöffnet, noch abzusenden, Adresse nochmals sichtbar, zurück zum Formular
möglich). Der Header-CTA „Für Vereine" zielt ab jetzt auf `/fuer-vereine`.

### Acceptance criteria

- [x] `anfrage-mailto.ts` ist UI-frei; vitest-Tests decken Pflichtfelder und
      URL-Encoding ab (Umlaute, Zeilenumbrüche).
- [x] Formular validiert clientseitig; Fehler sind programmatisch verknüpft
      (`aria-invalid`/`aria-describedby`, Live-Region); gültiges Absenden
      öffnet den vorbefüllten E-Mail-Entwurf mit allen Feldwerten per
      JS-Navigation, ohne `form-action`-Verstoß.
- [x] Bestätigungs-State sagt ehrlich, dass der Entwurf geöffnet wurde und
      noch zu senden ist; E-Mail-Adresse steht sichtbar auf der Seite.
- [x] Service-Angebot und Beta-Hinweis stehen auf der Seite; Meta (Title,
      Description, Canonical, OG) je Route korrekt; Header-CTA führt auf
      `/fuer-vereine`.
- [x] Mobil einspaltig bedienbar; `make website-check` grün; Island hydriert
      unter Produktiv-CSP.

---

## Phase 9: Screenshot-Skript, echte Screenshots, „Jedes Gerät" und Galerie

**User stories**: 6 (echte Screenshots hell/dunkel + Galerie), 17
(reproduzierbares Skript)

### Context

- `docker-compose.e2e.yml` + `e2e/support/seed.ts — resetAndSeed()` +
  `e2e/support/anmelden.ts` — Stack, Seed („Sommerfest TSV Musterstadt") und
  Login für die Aufnahmen.
- `website/src/assets/screenshots/` — die neun heutigen Motive als
  Motiv-Vorlage; werden ersetzt.
- Handoff-Prototyp `#screenshots` — „Ein Design. Jedes Gerät.": Desktop-Admin
  im Browser-Rahmen plus zwei Phones, eines fest hell, eines fest dunkel.
- `website/src/layouts/Landing.astro` — OG-Bild-Verdrahtung
  (heute `jotti-bon.webp`).
- `screenshots/` (Repo-Root) — zur Löschung freigegeben.
- Phase 3 — Hero-Phone erwartet ab hier echte hell/dunkel-Varianten.

### What to build

Ein Playwright-Skript im e2e-Paket (Make-Target `website-screenshots`)
startet gegen den e2e-Stack, ruft `resetAndSeed()` auf, meldet sich mit den
Seed-Zugangsdaten an und nimmt alle Website-Motive deterministisch auf: die
Desktop-Admin-Ansicht (Produktverwaltung) und als Phone-Motive Bestellansicht
(Hero), Zahlung, Stornierung, Direktverkauf, Tischübersicht, Produkte,
Benutzer, Geldtransit und Auswertung, jedes Motiv hell und dunkel
(`colorScheme`; zuvor verifizieren, dass die App der Systempräferenz folgt).
Ein zweiter Skript-Modus nimmt das OG-Bild als 1200×630-Capture des neuen
Hero vom gebauten Website-Artefakt auf: dafür baut das Skript die Website
(`make website-build`) und serviert `dist/` über denselben lokalen
Static-Server wie die CSP-Verifikation. Für jedes Motiv wird ein
beschreibender Alt-Text festgelegt. Die Dateien ersetzen
`website/src/assets/screenshots/` und das OG-Bild. Die neue
`#screenshots`-Sektion kombiniert das „Ein Design. Jedes Gerät."-Statement
(Desktop-Rahmen plus Phone hell und Phone dunkel, fest im jeweiligen Theme)
mit einer kompakten Galerie der übrigen Motive, theme-passend angezeigt. Der
Hero zeigt ab jetzt die echte helle/dunkle Bestellansicht theme-abhängig. Die
alte `#einblicke`-Carousel-Sektion entfällt; der Repo-Root-Ordner
`screenshots/` wird gelöscht.

### Acceptance criteria

- [x] `make website-screenshots` erzeugt alle Motive hell und dunkel
      reproduzierbar gegen den geseedeten e2e-Stack; das Skript ist
      committed und im Makefile dokumentiert (inkl. `.PHONY`-Eintrag).
- [x] Hero und `#screenshots` zeigen echte, aktuelle App-Screenshots; das
      „Jedes Gerät"-Statement entspricht dem Handoff
      (`06-light.png`/`03-dark.png`: ein Phone bleibt fest hell, eines fest
      dunkel).
- [x] Galerie zeigt die weiteren Motive kompakt und theme-passend; die alte
      `#einblicke`-Sektion samt Carousel-Code ist entfernt, sofern die
      Galerie ihn nicht wiederverwendet.
- [x] Jedes eingesetzte Motiv (Hero, „Jedes Gerät", Galerie) trägt einen
      beschreibenden Alt-Text.
- [x] OG-Bild zeigt das neue Design (1200×630) und ist je Route korrekt
      verdrahtet.
- [x] Veraltete Screenshot-Assets und der Repo-Root-Ordner `screenshots/`
      sind gelöscht; kein Bild der Seite zeigt mehr den alten App-Stand.
- [x] `make website-check` grün.

---

## Phase 10: Abschluss: Handoff entfernen, Gesamtabnahme

**User stories**: Abnahme-Querschnitt über alle Stories

### Context

- `docs/prds/design_handoff_jotti_website/` — wird nach abgeschlossener
  Umsetzung entfernt (PRD Further Notes; Git-Historie bewahrt es).
- `reverse-proxy/nginx.rocks.conf` — CSP für die finale Verifikation.
- PRD Testing Decisions — Abnahme ist Sichtprüfung des Betreibers plus
  CSP-Verifikation; keine automatisierten E2E-/Visual-Tests.

### What to build

Der Abschluss-Durchgang: Handoff-Bundle aus dem Repo entfernen; ein Scan
stellt sicher, dass „Regenbogen" nirgends im Website-Quelltext oder in der
ausgelieferten Copy vorkommt und keine Referenzen auf das Bundle
zurückbleiben. Vollständige Verifikation des gebauten Artefakts unter der
Produktiv-CSP über alle vier Routen und die Doku (hell/dunkel, Desktop/Mobil,
`prefers-reduced-motion`), Kontrast-Stichprobe der Token-Palette auf den
finalen Flächen (namentlich der Eyebrow, siehe Open questions),
Meta/OG/Canonical je Route. Die von Starlight transitiv erzeugte Sitemap
(`dist/sitemap-*.xml`) wird nach dem Build geprüft und muss die vier Routen
enthalten.
Danach die Sichtabnahme durch den Betreiber (dafür müssen die
Rechtsseiten-Texte eingesetzt sein); Deploy per `make rocks-up` erst nach der
Abnahme.

### Acceptance criteria

- [x] Handoff-Bundle entfernt; keine Referenzen darauf im Repo (außer
      Git-Historie und PRD-Erwähnung).
- [x] „Regenbogen" kommt in Quelltext und Copy der Website nicht vor.
- [x] Alle vier Routen und die Doku laufen unter der Produktiv-CSP ohne
      Konsolen-Verstöße, in Hell und Dunkel, auf Desktop und 390px-Mobil,
      mit und ohne `prefers-reduced-motion`.
- [x] Meta (Title, Description, Canonical, OG, Twitter) je Route korrekt;
      Sitemap enthält `/`, `/fuer-vereine/`, `/impressum/`, `/datenschutz/`.
- [x] Rechtsseiten-Texte des Betreibers sind eingesetzt (Platzhalter weg).
      — Texte mit dem Betreiber geklärt und eingesetzt; `Platzhalter.astro`
      samt `--sp-orange-text`-Token entfernt. Flankierend: Access-Logs auf
      gekürzte IPs umgestellt (`reverse-proxy/nginx.rocks.conf`,
      `website/nginx.conf`), damit die Datenschutzerklärung der Realität
      entspricht.
- [ ] `make website-check` grün; Sichtabnahme durch den Betreiber erteilt.
      — `make website-check` ist grün; offen bleibt die Sichtabnahme des
      Betreibers.
