# Plan: Website-Neubau mit Astro und Starlight

> Source PRD: [docs/prds/prd-website-neubau.md](../prds/prd-website-neubau.md)

## Goal

Die `jotti.rocks`-Website wird von handgeschriebenem HTML/CSS mit SSI-Partials
auf **Astro mit Starlight** umgestellt. Ergebnis: eine modern neugebaute
Marketing-Landing plus eine durchsuchbare, navigierbare Dokumentation im
Doku-Look. Die zentrale Garantie: **`docs/` bleibt die einzige Quelle der
Wahrheit.** Starlight liest die veröffentlichten Markdown-Dateien direkt aus
`docs/`, ohne Kopie und ohne Sync-Skript. Autoren und Agenten schreiben weiter
normale repo-relative Markdown-Links; ein remark-Plugin im Build wandelt sie in
Website-Routen (oder GitHub-URLs für Privates) um. Das alte SSI-/CSS-Mini-
Design-System-/`check-website.sh`-Setup wird vollständig zurückgebaut.

## Architectural decisions

Durchgängig gültige Entscheidungen:

- **Projektstruktur:** Neues, eigenständiges pnpm-Paket `website/` neben
  [frontend/](../../frontend/). **Eigener Lockfile** (`website/pnpm-lock.yaml`),
  getrennt vom Frontend (siehe Resolved decisions: bewusste Abweichung von der
  PRD-Vorgabe „ein Lockfile"). Das Frontend bleibt unangetastet.
- **Stack:** Astro + Starlight, Tailwind CSS 4 für die Landing, Starlight-Theme
  für die Doku. Keine Client-Framework-Islands (YAGNI). Suche über Pagefind
  (Starlight-eingebaut, vollständig statisch).
- **Sprache:** Fest deutsch. Starlight einsprachig mit einem Locale `de` als
  Root, keine i18n-Routen, kein Sprachumschalter. Die Landing sind reine
  deutsche Astro-Seiten ohne i18n-Routing.
- **Routen:**
  - Landing auf `/` als eigenständige Tailwind-Astro-Seite(n).
  - Doku unter `/docs/…`, z. B. `/docs/compliance/`,
    `/docs/produktbeschreibung/`, `/docs/leitfaden/<schritt>/`. Starlight
    rendert ausschließlich unterhalb von `/docs/`.
- **Quelle der Wahrheit:** Astro-Content-Collection mit Glob-Loader, dessen
  Basis auf das top-level [docs/](../../docs/) zeigt. Veröffentlichungs-Auswahl
  an genau einer Stelle (Glob-Muster plus Sidebar-Definition in der
  Website-Konfiguration). Fallback bei Loader-Problemen: im Build erzeugter
  Symlink vom Content-Verzeichnis auf `docs/`.
- **Veröffentlichte Dateien (Phase 1):** der restrukturierte Leitfaden
  (`docs/leitfaden/`), [compliance.md](../compliance.md),
  [steuerrecht.md](../steuerrecht.md),
  [verfahrensdokumentation.md](../verfahrensdokumentation.md),
  [produktbeschreibung.md](../produktbeschreibung.md),
  [lizenzmodell.md](../lizenzmodell.md).
- **Privat (nie veröffentlicht):** `handbuch.md`, `language.md`,
  `anforderungen.md`, `jotti-rocks-infra.md`, `docs/plans/`, `docs/prds/`.
- **Link-Auflösung:** Reine Funktion `(Link-Ziel, Quellpfad, Map
  veröffentlichter Dokumente, Repo-Basis-URL) -> href`. Veröffentlichtes Ziel
  wird Website-Route (Anker als Slug der Zielüberschrift); privates Ziel oder
  Datei außerhalb `docs/` wird absolute GitHub-URL; externe Links unverändert.
- **Frontmatter:** Veröffentlichte Dokumente erhalten minimales Frontmatter
  (`title`, `description`, optional Sidebar-Label/Reihenfolge); die bisherige H1
  entfällt zugunsten des Frontmatter-Titels. Private Dokumente bleiben ohne
  Frontmatter.
- **Build als Gate:** `astro check` plus Build (plus Vitest für den
  Link-Rewriter) ersetzen `scripts/check-website.sh`. Eine automatische
  Link-/Anker-Validierung als Build-Gate ist bewusst entfallen (User-Entscheidung
  in Phase 3: KISS, kein Link-Validator — `starlight-links-validator` ist mit dem
  externen Glob-Loader inkompatibel, siehe Phase 3).
- **Auslieferung:** Eigener `website`-Container analog zu
  [frontend/Dockerfile](../../frontend/Dockerfile) (Astro-Build → nginx-Runner).
  Der Reverse-Proxy proxyt `jotti.rocks` → `website:80`; die Security-Header/CSP
  bleiben im Reverse-Proxy. Deployment manuell wie bisher.

## Inventory

Bestehendes Setup, das ersetzt oder angepasst wird:

- [website/](../../website/) — handgeschriebene Seiten: `index.html`,
  `leitfaden-fuer-vereine/index.html` (handgebaute Spiegelung von
  `docs/leitfaden.md`), `partials/*.html` (SSI), `css/base.css`
  (Mini-Design-System), `404.html`, `robots.txt`, `sitemap.xml`,
  `fonts/Montserrat-*.woff2`, `icons/`, `img/`, `js/main.js`.
- [docs/leitfaden.md](../leitfaden.md) — eine lange Datei; Überschriften decken
  Standardweg, TSE, Pflichten, Checkliste, Experten-Weg, TSE-Sonderfälle,
  Fehlersuche, Häufige Fragen ab (Basis für die Aufteilung).
- [reverse-proxy/nginx.rocks.conf](../../reverse-proxy/nginx.rocks.conf) —
  Landing-Server-Block mit `ssi on`, `location ^~ /partials/ { internal; }`,
  `/jotti-selbst-betreiben/`-Redirect, `try_files`; mountet
  `/usr/share/nginx/html/landing`. Demo-Block proxyt bereits `frontend:80` (Muster).
- [reverse-proxy/nginx.website-dev.conf](../../reverse-proxy/nginx.website-dev.conf)
  — lokaler SSI-Preview, wird entfernt.
- [docker-compose.rocks.yml](../../docker-compose.rocks.yml) — `reverse-proxy`
  mountet `./website:/usr/share/nginx/html/landing:ro`; `frontend`-Service ist
  das Vorbild für einen eigenen `website`-Container.
- [frontend/Dockerfile](../../frontend/Dockerfile) — Multi-Stage-Muster
  (pnpm-Build → nginx:1.27-alpine-Runner) zum Nachbauen.
- [scripts/check-website.sh](../../scripts/check-website.sh) — SSI-/Link-/Asset-/
  CSS-Klassen-Prüfung, wird durch `astro check` + Build ersetzt.
- [Makefile](../../Makefile) (Zeilen 293–306) — `website`, `website-check`,
  `website-fmt`; werden ersetzt.
- [AGENTS.md](../../AGENTS.md) (Zeilen 23–28) — Doku-Referenztabelle; Verweise auf
  `docs/leitfaden.md` werden auf die neue Struktur aktualisiert.
- [frontend/package.json](../../frontend/package.json) — Vorbild für
  Scripts/Vitest; Vitest ist hier bereits etabliert.

## Resolved decisions

- **Workspace: zwei getrennte Lockfiles.** `website/` ist ein eigenständiges
  pnpm-Paket mit eigenem Lockfile, das Frontend bleibt unverändert. Bewusste
  Abweichung von der PRD-Vorgabe „derselbe pnpm-Workspace, ein Lockfile"
  (User-Entscheidung): kein Root-Workspace-Migrationsaufwand, dafür zwei
  Toolchains.
- **Routing: Landing auf `/`, Doku unter `/docs/…`.** Starlight rendert
  ausschließlich unter `/docs/`. Klare Trennung der zwei Welten; der
  Link-Rewriter mappt `x.md -> /docs/x/`.
- **Deployment: eigener `website`-Container wie `frontend`.** Reverse-Proxy
  proxyt `jotti.rocks` → `website:80`; CSP/Security-Header bleiben im
  Reverse-Proxy.
- **Alte URLs:** Redirects von `/leitfaden-fuer-vereine/` und
  `/jotti-selbst-betreiben/` auf die neue Leitfaden-Einstiegsseite bleiben
  erhalten (SEO-Kontinuität).
- **Sequenzierung:** Die Auslieferung (nginx/Compose) und der Rückbau des alten
  Setups erfolgen gemeinsam in der letzten Phase. Phasen 1–5 werden über den
  lokalen Astro-Dev-Server und das `astro build`-Gate validiert; die aktuelle
  Landing bleibt live, bis die neue Seite inhaltlich vollständig ist.

## Open questions / Risks

- **Glob-Loader auf externes Verzeichnis (Hauptrisiko).** ✅ In Phase 1
  aufgelöst: Astros generischer `glob()`-Loader (`astro/loaders`) mit
  `base: '../docs'` liest die veröffentlichten Dateien direkt aus dem top-level
  `docs/`. `astro check` und Build laufen grün; Datei-Watching im Dev-Modus
  greift (Änderung in `docs/` ohne Neustart sichtbar). Der Symlink-Fallback wird
  nicht gebraucht. Hinweis: `docsLoader()` selbst kann nicht auf ein externes
  Verzeichnis zeigen (feste Basis `src/content/docs/`), daher der `glob()`-Loader
  mit Starlights `docsSchema()`.
- **Starlight unter Subpfad `/docs/`.** ✅ In Phase 1 aufgelöst: Der
  `generateId`-Callback des Loaders prefixt jeden Slug mit `docs/`
  (`lizenzmodell.md` → `docs/lizenzmodell` → `/docs/lizenzmodell/`). Die Landing
  liegt als eigene Astro-Seite unter `src/pages/index.astro` auf `/`; Starlight
  beansprucht den Root nicht, weil kein Eintrag den leeren Slug hat.
- **Prettier ohne Frontend-Installation.** Die PRD nannte „Prettier bleibt über
  die Frontend-Installation". Mit getrennten Paketen bekommt `website/` eine
  eigene Prettier-Einbindung (oder teilt die Konfiguration). In Phase 1 festlegen.

---

## Phase 1: Walking Skeleton — website-Paket und ein Doku-Dokument end-to-end

**User stories**: 24, 25, 22 (sowie kostenlose Teilabdeckung von 1, 2, 3, 4, 5
über Starlight: Sidebar-Navigation, Volltextsuche, Mobile-Lesbarkeit,
Dunkelmodus, „Auf dieser Seite")

### Context

- [frontend/package.json](../../frontend/package.json) — Vorbild für Scripts,
  pnpm-Engines, Vitest.
- [frontend/pnpm-workspace.yaml](../../frontend/pnpm-workspace.yaml) — zeigt das
  bestehende `allowBuilds`-Muster.
- [website/fonts/](../../website/fonts/) — vorhandene Montserrat-`woff2`-Dateien
  zum Self-Hosting übernehmen.
- [Makefile](../../Makefile) (Zeilen 293–306) — die zu ersetzenden
  Website-Targets.
- PRD-Abschnitt „Risiko Glob-Loader auf externes Verzeichnis".

### What to build

Ein eigenständiges pnpm-Paket `website/` mit Astro, Starlight und Tailwind 4.
Marke einmal als gemeinsames Token-Set (Markengrün, Montserrat self-hosted)
definiert, das Landing und Starlight-Theme (über CSS-Custom-Properties) speist.
Starlight einsprachig (`de` als Root) konfiguriert, Doku-Routen unter `/docs/`.

Eine Content-Collection liest per Glob-Loader direkt aus dem top-level `docs/`.
**Genau ein** Dokument wird veröffentlicht (z. B. `lizenzmodell.md`) mit
minimalem Frontmatter und erscheint unter `/docs/lizenzmodell/`. Eine
Platzhalter-Landing liegt auf `/`. Der Dev-Modus muss Änderungen an der Datei in
`docs/` live übernehmen (Datei-Watching); falls der Loader auf das externe
Verzeichnis unzuverlässig ist, greift der Symlink-Fallback.

Neue `make`-Targets: lokaler Dev-Server (Astro dev), Build und Qualitäts-Check
(`astro check` plus Build).

### Acceptance criteria

- [x] `website/` ist ein eigenständiges pnpm-Paket (eigener Lockfile); ein
      Dev-Server startet mit einem `make`-Target und ist lokal erreichbar.
- [x] Die Platzhalter-Landing liegt auf `/`, das eine veröffentlichte Dokument
      auf `/docs/<slug>/`; das Frontend-Paket bleibt unverändert.
- [x] Das Doku-Dokument wird direkt aus `docs/` gelesen (keine Kopie); eine
      Änderung der Datei in `docs/` ist im Dev-Server ohne Neustart sichtbar.
- [x] Sidebar-Navigation, Volltextsuche (Pagefind), Dunkelmodus (folgt
      System-Einstellung) und „Auf dieser Seite" funktionieren auf der
      Doku-Seite; die Framework-Texte sind deutsch.
- [x] Marke ist sichtbar gebrandet (Markengrün, Montserrat) und über ein
      gemeinsames Token-Set für Landing und Doku konsistent.
- [x] `astro check` und der Build laufen grün durch.

---

## Phase 2: Eine Quelle der Wahrheit — die flachen Doku-Dokumente veröffentlichen

**User stories**: 13, 14, 10, 22, 29

### Context

- [docs/compliance.md](../compliance.md),
  [docs/steuerrecht.md](../steuerrecht.md),
  [docs/verfahrensdokumentation.md](../verfahrensdokumentation.md),
  [docs/produktbeschreibung.md](../produktbeschreibung.md),
  [docs/lizenzmodell.md](../lizenzmodell.md) — die fünf flachen, in Phase 1
  bereits teils erschlossenen Dokumente.
- Glob-Loader und Veröffentlichungs-Auswahl aus Phase 1.
- PRD-Abschnitt „Vorgeschlagene Sidebar-Struktur (Doku)".

### What to build

Die fünf flachen veröffentlichten Dokumente erhalten minimales Frontmatter
(`title`, `description`, optional Sidebar-Label und Reihenfolge); die bisherige
H1 entfällt. Die vollständige Starlight-Sidebar-Struktur wird angelegt (alle
Gruppen außer dem Leitfaden, der in Phase 4 folgt): „Recht und Steuern" und
„Über jotti" wie in der PRD vorgeschlagen.

Die Veröffentlichungs-Auswahl steht an genau einer Stelle (Glob-Muster plus
Sidebar-Definition). Private Dokumente (`handbuch.md`, `language.md`,
`anforderungen.md`, `jotti-rocks-infra.md`, `docs/plans/`, `docs/prds/`) bleiben
unverändert und sind über die Website nicht erreichbar.

### Acceptance criteria

- [x] Alle fünf Dokumente rendern unter `/docs/<slug>/` mit dem aus Frontmatter
      gerenderten Titel; keine doppelte Überschrift (H1 entfernt).
- [x] Die Sidebar zeigt die Gruppen aus der PRD (ohne Leitfaden) in der
      definierten Reihenfolge.
- [x] Jede veröffentlichte Seite hat `title` und `description` für SEO/Open Graph.
- [x] Private Dokumente sind nicht als Website-Route erreichbar und unverändert
      (kein Frontmatter).
- [x] Welche Dateien veröffentlicht werden, ist an einer Stelle in der
      Website-Konfiguration ablesbar.
- [x] `astro check` und Build laufen grün; die Seiten sind auf dem Smartphone
      gut lesbar.

---

## Phase 3: remark-Link-Rewriter und Build-Link-Validierung

**User stories**: 18, 19, 20, 21, 23

### Context

- [frontend/package.json](../../frontend/package.json) — Vitest als Vorbild für
  isolierte Unit-Tests.
- [scripts/check-website.sh](../../scripts/check-website.sh) — die bisherigen
  Link-/Anker-Prüfungen, die hier abgelöst werden.
- PRD-Abschnitt „Querverweise: remark-Link-Rewriter".

### What to build

Die deklarierte Kernlogik als reine, isoliert testbare Funktion: Aus (Link-Ziel,
Pfad des Quelldokuments, Map veröffentlichter Dokumente, Repo-Basis-URL) wird
das korrekte `href` berechnet. Verhalten:

- Link auf veröffentlichtes Dokument → Website-Route (`/docs/<slug>/`), Anker als
  Slug der Zielüberschrift.
- Link auf privates Dokument oder Repo-Datei außerhalb `docs/` (`../TERMS.md`,
  `../LICENSE`, `../README.md`, `handbuch.md`) → absolute GitHub-URL.
- Externe Links (`http`, `https`, `mailto`) → unverändert.

Diese Funktion wird mit Vitest über repräsentative Fälle getestet und als
remark-Plugin in den Astro-Build eingehängt. Zusätzlich wird die
Build-Link-Validierung aktiviert: ein Link auf ein veröffentlichtes Ziel oder
einen Anker, der nicht existiert, bricht den Build ab.

### Acceptance criteria

- [x] Repo-relative Links in veröffentlichten Dokumenten lösen auf `/docs`-Routen
      auf; Anker zeigen auf den Slug der Zielüberschrift.
- [x] Links auf private Dokumente und Dateien außerhalb `docs/` werden zu
      absoluten GitHub-URLs; externe Links bleiben unverändert.
- [x] Die reine `href`-Funktion ist mit Vitest abgedeckt (veröffentlicht mit/ohne
      Anker, privat → GitHub, außerhalb `docs/` → GitHub, extern unverändert).
- [~] ~~Ein toter Link oder fehlender Anker auf ein veröffentlichtes Ziel lässt den
      Build fehlschlagen.~~ **Bewusst entfallen (User-Entscheidung: KISS, kein
      Link-Validator).** `starlight-links-validator` ist mit dem externen
      Glob-Loader + `/docs/`-`generateId`-Präfix inkompatibel: es schlüsselt
      Heading-Indizes über den Dateipfad relativ zu `src/content/docs` und würde
      jeden Link auf ein veröffentlichtes Dokument verwerfen. Die einzige Brücke
      wäre ein `slug: docs/<name>`-Frontmatter pro kanonischer Datei — das würde
      die Website-Route in die Quelle der Wahrheit einbetten und `generateId`
      duplizieren. Statt dessen kein Build-Gate für Links.
- [x] Die GitHub-Vorschau und die Editor-Vorschau der Quelldateien bleiben gültig
      (Autoren schreiben weiter repo-relative Links).

---

## Phase 4: Leitfaden-Restrukturierung

**User stories**: 6, 7, 8, 9, 11, 12

### Context

- [docs/leitfaden.md](../leitfaden.md) — die aufzuteilende Datei; bestehende
  Überschriften bilden die Grundlage der Gliederung.
- [AGENTS.md](../../AGENTS.md) (Zeilen 23–28) — Referenztabelle mit Verweisen auf
  den Leitfaden.
- Link-Rewriter aus Phase 3 (Querverweise auf die neue Struktur müssen auflösen).
- PRD-Abschnitt „Leitfaden-Restrukturierung".

### What to build

`docs/leitfaden.md` wird in einen Ordner `docs/leitfaden/` aus einzelnen
Schritt-Seiten aufgeteilt, gegliedert nach Zielgruppe und Thema entlang der
PRD-Sidebar:

- Erste Schritte: Was ist jotti?, Schnellstart (Standardweg)
- Vereinsbetrieb (Standardweg): Installation und Start, TSE einrichten
  (fiskaly), Täglicher Betrieb, Checkliste
- Recht und Steuern: Pflichten im Überblick, Kasse beim Finanzamt anmelden,
  Belege und Steuersätze, Datenaufbewahrung, Verfahrensdokumentation (Verweis),
  Compliance-Grundlagen (Verweis)
- Self-Hosting (Experten-Weg): Eigener Server (Ersteinrichtung), Aktualisieren
  und Backups, TSE-Sonderfälle
- Hilfe: Fehlersuche, Häufige Fragen

Der Inhalt wird nicht neu erfunden, sondern umgegliedert und in
Schritt-für-Schritt-Form gebracht; die Trennung Standardweg/Experten-Weg und
Technik/Recht entsteht über die Gruppierung. Jede Schritt-Seite erhält
Frontmatter. Die `AGENTS.md`-Referenztabelle und alle Markdown-Querverweise auf
`docs/leitfaden.md` werden auf die neue Struktur aktualisiert. Deep-Links aus
Nicht-Markdown-Konsumenten (Frontend, Backend, Compose) sieht der Astro-Build
nicht; sie werden in Phase 6 vereinheitlicht.

### Acceptance criteria

- [ ] `docs/leitfaden.md` ist in `docs/leitfaden/`-Schritt-Seiten aufgeteilt; der
      Inhalt ist erhalten, nur umgegliedert.
- [ ] Die Leitfaden-Sidebar trennt Standardweg vom Experten-Weg und Technik vom
      Recht (Gruppierung wie oben).
- [ ] Eine klare Schritt-für-Schritt-Führung für den Standardweg existiert; der
      Experten-Weg (Self-Hosting) ist ein eigener Bereich.
- [ ] Ein Fehlersuche-/FAQ-Bereich ist vorhanden.
- [ ] `AGENTS.md` und alle Markdown-Querverweise zeigen auf die neue
      `docs/leitfaden/`-Struktur; keine toten Verweise (Build grün).
      Nicht-Markdown-Deep-Links bleiben Phase 6 vorbehalten.

---

## Phase 5: Marketing-Landing neu bauen

**User stories**: 15, 16, 17, 29, 30

### Context

- [website/index.html](../../website/index.html) — aktuelle Landing (Hero,
  Vertrauensmerkmale, Features, Vergleich, CTA) als inhaltliche Vorlage.
- [website/css/base.css](../../website/css/base.css) — bestehendes Mini-Design-
  System als Referenz für Markenwerte (wird in Phase 6 entfernt).
- Brand-Token-Set und Tailwind-Setup aus Phase 1.
- Leitfaden-/Doku-Routen aus Phasen 2 und 4 (CTA-Ziele).

### What to build

Die Landing auf `/` wird mit Astro und Tailwind modern neu aufgebaut. Marke
(Grün, Montserrat, Logo) und Botschaft der heutigen Seite (Hero,
Vertrauensmerkmale, Features, Vergleich, CTA) bleiben inhaltlich erhalten; keine
inhaltliche Neukonzeption. Bilder laufen über Astros Bild-Optimierung (moderne
Formate, responsive Größen) statt manuell gepflegter Varianten. Die CTAs führen
klar in den Leitfaden bzw. die Doku.

Jede Seite (Landing wie Doku) trägt Titel und Beschreibung für SEO und Open
Graph. `sitemap.xml` und `robots.txt` werden automatisch erzeugt bzw. gepflegt,
sodass die neuen Doku-Seiten indexiert werden.

### Acceptance criteria

- [ ] Die Landing ist visuell auf Marke (Grün, Montserrat, Logo), responsiv und
      im Dunkelmodus konsistent mit der Doku.
- [ ] Hero, Vertrauensmerkmale, Features, Vergleich und CTA sind inhaltlich wie
      heute vorhanden; die Value-Proposition ist auf einen Blick erkennbar.
- [ ] Die CTAs führen in Leitfaden und Doku.
- [ ] Bilder werden über Astros Bild-Optimierung ausgeliefert.
- [ ] Jede veröffentlichte Seite hat Titel/Beschreibung (SEO/OG); `sitemap.xml`
      und `robots.txt` werden ausgeliefert und decken die Doku-Seiten ab.

---

## Phase 6: Auslieferung und Rückbau des alten Setups

**User stories**: 26, 27, 28

### Context

- [frontend/Dockerfile](../../frontend/Dockerfile) — Multi-Stage-Vorbild (Build →
  nginx-Runner).
- [docker-compose.rocks.yml](../../docker-compose.rocks.yml) — `frontend`-Service
  als Vorbild; `reverse-proxy` mountet heute `./website` als Volume.
- [reverse-proxy/nginx.rocks.conf](../../reverse-proxy/nginx.rocks.conf) —
  Landing-Server-Block (SSI, Partials, `/jotti-selbst-betreiben/`-Redirect),
  wird auf `proxy_pass website:80` umgestellt; CSP/Header bleiben.
- [reverse-proxy/nginx.website-dev.conf](../../reverse-proxy/nginx.website-dev.conf),
  [scripts/check-website.sh](../../scripts/check-website.sh),
  [website/partials/](../../website/partials/),
  [website/css/base.css](../../website/css/base.css),
  [website/leitfaden-fuer-vereine/](../../website/leitfaden-fuer-vereine/) — zu
  entfernen.
- [Makefile](../../Makefile) (Zeilen 293–306) — alte Targets ersetzen.
- Externe Deep-Link-Konsumenten — zeigen heute auf GitHub-`blob`-URLs von
  `docs/leitfaden.md` und brechen nach der Aufteilung in Phase 4:
  [DokumenteUndPflichtenSection.tsx](../../frontend/src/admin/finanzamt/DokumenteUndPflichtenSection.tsx)
  (`docs/leitfaden.md`, `docs/compliance.md`),
  [KassenidentitaetSection.tsx](../../frontend/src/admin/finanzamt/KassenidentitaetSection.tsx)
  (`docs/leitfaden.md#kasse-beim-finanzamt-anmelden`),
  [reverse-proxy/statuspage.go](../../reverse-proxy/statuspage.go)
  (`docs/leitfaden.md#fehlersuche`); dazu Kommentar-Verweise in
  `docker-compose.prod.yml` und `docker-compose.local.yml`.

### What to build

Ein eigener `website`-Container analog zum Frontend: Multi-Stage-Dockerfile
(Astro-Build → nginx:1.27-alpine-Runner), als `website`-Service in
`docker-compose.rocks.yml`. Der Reverse-Proxy-Landing-Block wird von statischem
Serving mit SSI auf `proxy_pass http://website:80` umgestellt; Security-Header
und CSP bleiben im Reverse-Proxy. Redirects von `/leitfaden-fuer-vereine/` und
`/jotti-selbst-betreiben/` auf die neue Leitfaden-Einstiegsseite bleiben für
SEO-Kontinuität erhalten. Das `./website`-Volume entfällt.

Externe Doku-Links werden vereinheitlicht: Frontend und Backend zeigen nicht mehr
auf GitHub-`blob`-Deep-Links, sondern auf kanonische
`https://jotti.rocks/docs/<slug>`-URLs. Betroffen sind
`DokumenteUndPflichtenSection.tsx`, `KassenidentitaetSection.tsx` und
`reverse-proxy/statuspage.go`; die Kommentar-Verweise in den Compose-Dateien
werden auf die neue `docs/leitfaden/`-Struktur nachgezogen. Weil die Slugs stabil
gehalten werden, braucht eine spätere Doku-Restrukturierung keinen Eingriff in
TS/Go mehr, sondern nur einen Eintrag in der Redirect-Map. Diese Redirect-Map
liegt an einer Stelle im Reverse-Proxy (Muster der bestehenden
`/leitfaden-fuer-vereine/`- und `/jotti-selbst-betreiben/`-Redirects) und deckt
umbenannte oder verschobene Doku-Routen ab, sodass alte
`jotti.rocks/docs/<slug>`-URLs weiter auflösen. Vor dem Rückbau wird das Repo
einmal nach verbliebenen Deep-Links auf `docs/leitfaden.md` durchsucht (Frontend,
Backend, Konfiguration), damit kein toter Verweis zurückbleibt, den der
Astro-Build nicht sieht.

Das alte Setup wird vollständig zurückgebaut: SSI-Partials und ihre
nginx-Konfiguration, `css/base.css`, `scripts/check-website.sh`, die handgebaute
Leitfaden-HTML (`website/leitfaden-fuer-vereine/`) und `nginx.website-dev.conf`.
Die alten Makefile-Targets werden durch die neuen aus Phase 1 ersetzt. Die
Projektdoku zum Website-Workflow (Hinweise auf „statische Seite, kein
Build-Schritt") wird angepasst.

### Acceptance criteria

- [ ] Der Rocks-Stack baut das `website`-Image in einer Build-Stage und serviert
      Landing und Doku über `jotti.rocks` via Reverse-Proxy; kein `dist/` im Repo,
      kein Host-seitiger Build.
- [ ] Alte URLs (`/leitfaden-fuer-vereine/`, `/jotti-selbst-betreiben/`) leiten
      auf die neue Leitfaden-Einstiegsseite um.
- [ ] SSI-Partials, `css/base.css`, `scripts/check-website.sh`, die
      Leitfaden-HTML und `nginx.website-dev.conf` sind entfernt; das
      `./website`-Volume ist aus dem Compose-Setup raus.
- [ ] Die Makefile-Targets sind ersetzt (Dev-Server, Build, Check); kein
      Parallelsystem bleibt.
- [ ] Die Projektdoku zum Website-Workflow ist auf das neue Build-Setup
      aktualisiert.
- [ ] Frontend und Backend verlinken Doku über kanonische
      `jotti.rocks/docs/<slug>`-URLs statt GitHub-`blob`-Deep-Links
      (`DokumenteUndPflichtenSection.tsx`, `KassenidentitaetSection.tsx`,
      `statuspage.go`); die Compose-Kommentare zeigen auf `docs/leitfaden/`.
- [ ] Eine repo-weite Suche nach Deep-Links auf `docs/leitfaden.md` außerhalb
      veröffentlichter Markdown ist leer; kein toter Anker-Verweis bleibt.
- [ ] Umbenannte oder verschobene Doku-Routen sind über eine Redirect-Map an
      einer Stelle im Reverse-Proxy abgedeckt; alte `jotti.rocks/docs/<slug>`-URLs
      lösen weiter auf.
- [ ] Deployment funktioniert manuell wie bisher (kein automatisiertes Deployment).
