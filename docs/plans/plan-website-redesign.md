# Plan: Website-Redesign — Clean Design & responsive Bugfixes

> Quell-PRD: n/a (abgeleitet aus der Design-Aufgabe „Website neu gestalten,
> vereinfachen, cleaner; UI/UX auf Mobile & Desktop prüfen")

## Ziel

Die jotti-Website (`website/`) optisch entschlacken: **weniger visuelle
Komplexität, weniger Boxen/Icons/Bold, mehr Ruhe und Lesbarkeit** — der Inhalt
steht im Vordergrund. Dabei:

1. **Design-Sprache vereinheitlichen** statt vieler unterschiedlicher UI-Muster
   (Card-Varianten, Verläufe, Glow-Effekte, Emoji-Icons, Badges).
2. **Inhalte konsolidieren** — redundante Abschnitte zusammenführen,
   Wiederholungen kürzen; alle wesentlichen Informationen bleiben erhalten.
3. **Farbwelt auf die kanonische grüne CI ausrichten** (konsistent mit der
   App), Violett/Indigo aus dem CSS heraushalten.
4. **Zwei konkrete Responsive-Bugs beheben:** verzerrtes Logo im Header und
   überlaufende Navbar — anschließend UI/UX auf Mobile **und** Desktop prüfen.

Die statische Architektur (Static HTML + nginx-SSI) bleibt unverändert; es gibt
keinen Build-Schritt und kein Re-Platforming. Das Redesign ist rein
HTML/CSS/Inhalt.

## Architekturentscheidungen

Durchgängige, phasenübergreifende Festlegungen:

- **Architektur/Stack bleibt:** statisches HTML + nginx-SSI-Partials, kein
  Build-Schritt, **nur Light-Theme** (kein Dark Mode — bewusst außerhalb des
  Scopes, da es Komplexität erhöhen würde).
- **Farbe — kanonisches Grün:** Die App ist die Quelle der Wahrheit
  (`--primary: oklch(0.508 0.118 165.612)`, Grün-Hue ≈165 in
  [frontend/src/index.css](../../frontend/src/index.css#L66)). Das bereits grüne
  Token-System der Website (`--primary: #4a7c59` u. a.) wird damit
  abgeglichen/konsolidiert; **kein Violett/Indigo im CSS**.
- **Design-Sprache reduzieren:** eine konsistente, leichte Behandlung statt
  vieler Varianten — wenig Rahmen/Boxen, **keine dekorativen Emoji-Icons**,
  weniger Bold, weniger Verläufe/Schatten/Glow-Pseudoelemente, großzügiger
  Weißraum, content-first Typografie.
- **Geteilte Partials einmal sauber:** Header, Footer, Mobile-Nav, Dev-Banner
  werden zentral überarbeitet und gelten dadurch auf allen Seiten.
- **Dichte Datenblöcke bleiben — aber leichter:** echte tabellarische Daten
  (Server-Specs, Fristen, Glossar, Vergleich) bleiben als **leichte Tabelle**
  (lesbarste Form); rein dekorative Boxen/Cards werden aufgelöst.
- **Schriftart bleibt Montserrat** (kein Wechsel auf Inter — bewusste
  Scope-Grenze).
- **Verifikationspfad:** `make website` → `http://localhost:8080` (nginx+SSI),
  geprüft auf Mobile (≤640 px), Tablet (~768/960 px) und Desktop (≥1120 px).

## Inventar

Bestehender Code, Muster und Referenzen — jeweils mit Pfad:Zeilen:

- [website/index.html](../../website/index.html) — Landingpage, **12 Abschnitte**:
  HERO [L29](../../website/index.html#L29), PROBLEM
  [L72](../../website/index.html#L72), SOLUTION
  [L112](../../website/index.html#L112), FEATURES
  [L158](../../website/index.html#L158) (4 Gruppen, ~22 Emoji-Cards),
  FISKALKONFORMITÄT [L341](../../website/index.html#L341), SO FUNKTIONIERT'S
  [L442](../../website/index.html#L442), VERGLEICH
  [L481](../../website/index.html#L481) (große Tabelle), FÜR WEN
  [L582](../../website/index.html#L582), SICHERHEIT
  [L625](../../website/index.html#L625), KOSTEN
  [L673](../../website/index.html#L673), TECHNIK
  [L731](../../website/index.html#L731), CTA
  [L842](../../website/index.html#L842).
- [website/jotti-selbst-betreiben/index.html](../../website/jotti-selbst-betreiben/index.html) —
  Hosting-Leitfaden (673 Z.), `.guide-prose`-Struktur, viele
  `.compliance-note`-Callouts mit Emoji, `.card-grid`-Intros, mehrere
  `.comparison-table`-Datentabellen (Specs, IP-Befehle, Glossar).
- [website/leitfaden-fuer-vereine/index.html](../../website/leitfaden-fuer-vereine/index.html) —
  Compliance-Leitfaden (562 Z.), gleiche `.guide-*`-Muster, Emoji-Cards
  („Das Wichtigste"), Verantwortungs-/Fristen-Tabellen, `.guide-faq`.
- [website/css/base.css](../../website/css/base.css) — **1426 Z.**, alle Stile.
  Relevante Gruppen: Tokens `:root`
  [L30](../../website/css/base.css#L30), Buttons
  [L161](../../website/css/base.css#L161), Header
  [L232](../../website/css/base.css#L232) (`.logo`
  [L256](../../website/css/base.css#L256), `.header-byline`
  [L268](../../website/css/base.css#L268), `.main-nav`
  [L293](../../website/css/base.css#L293)), Hero
  [L415](../../website/css/base.css#L415), Card-Grid + Glow-`::before`
  [L598](../../website/css/base.css#L598), Compliance-Verlauf
  [L556](../../website/css/base.css#L556), Comparison-Table
  [L853](../../website/css/base.css#L853), Pricing
  [L939](../../website/css/base.css#L939), Tech
  [L1014](../../website/css/base.css#L1014), CTA-Verlauf
  [L1099](../../website/css/base.css#L1099), Badge-WIP ≈L1245,
  Guide-Stile ≈L1280–1426.
- [website/partials/header.html](../../website/partials/header.html) — Logo-`<img>`
  **ohne `width`/`height`-Attribute** (L4–9), Byline (L11–16), `.main-nav` mit
  **8 Links + GitHub-Button** (L17–35), Hamburger (L36–42).
- [website/partials/mobile-nav.html](../../website/partials/mobile-nav.html) —
  9 Links (Overlay).
- [website/partials/footer.html](../../website/partials/footer.html),
  [website/partials/dev-banner.html](../../website/partials/dev-banner.html),
  [website/partials/head-assets.html](../../website/partials/head-assets.html).
- [website/js/main.js](../../website/js/main.js) — 20 Z., nur Mobile-Menü-Toggle.
- `website/img/` — Logo `jotti-logo-full-light-transparent.png` **697×316**
  (≈2,2:1, **violett**), `jotti-symbol.png` 537×825 (violett), Screenshots
  `jotti-admin-products.webp` 1000×611, `jotti-service-dashboard.webp` 600×1216,
  `jotti-admin-reporting.png` 800×579 (zeigen **alte violette App-UI**).
- [frontend/src/index.css](../../frontend/src/index.css#L52-L100) — kanonisches
  Grün, aber **veraltete `/* violet-600 */`-Kommentare** (Werte sind grün);
  `--font-sans: 'Inter Variable'` [L44](../../frontend/src/index.css#L44).
- [assets/assets-and-design.md](../../assets/assets-and-design.md) — Abschnitt 5
  „Das Theme (Farben)" dokumentiert **noch eine violette CI** (veraltet).
- [Makefile](../../Makefile#L214) — `make website` startet nginx+SSI unter
  `http://localhost:8080` (Verifikationspfad aller Phasen).
- [docs/plans/plan-website-seo-404.md](./plan-website-seo-404.md) — paralleler
  Website-Plan (SEO/404); berührt v. a. `<head>` und nginx, **kein Konflikt**
  mit diesem visuellen Redesign (auf Überschneidung im Head achten).

## Aufgelöste Entscheidungen

- **Grün ist kanonisch**, auf den App-Wert ausgerichtet; kein Violett/Indigo im
  CSS. (Die Website ist bereits grün — Phase 1 konsolidiert die Tokens, führt
  Grün also nicht neu ein.)
- **Aufräumen + Inhalte konsolidieren:** redundante Abschnitte zusammenführen,
  Wiederholungen kürzen, alle wesentlichen Infos behalten.
- **Dichte Blöcke behalten, aber leichter:** Tabellendaten bleiben als leichte
  Tabellen; dekorative Boxen/Cards werden aufgelöst.
- **Montserrat behalten** (kein Wechsel auf Inter).
- **Violette Bild-Assets zurückgestellt:** Logo/Favicons/Symbol/Screenshots
  bleiben vorerst; Neugestaltung in Grün ist separates Design-Follow-up (Bilder
  nicht per Code erzeugbar). Die **Logo-Verzerrung** wird dennoch gefixt (CSS).
- **Mini-Cleanup aufgenommen:** veraltete Violett-Kommentare in
  `frontend/src/index.css` und die Farbsektion in `assets/assets-and-design.md`
  auf Grün korrigieren (Phase 5).
- **Granularität:** 4 Website-Phasen + 1 kleine Cleanup-Phase; beide Leitfäden
  in einer Phase.

## Offene Fragen / Risiken

- **Violette Bild-Assets bleiben ein Rest-Inkonsistenz** zum grünen Theme
  (Logo, Favicons, J-Symbol, Screenshots zeigen alte violette App-UI). Bewusst
  zurückgestellt → eigenständiges Design-/Asset-Follow-up.
- **Logo-Verzerrung — Ursache in-Browser bestätigen.** Wahrscheinlich fehlen die
  intrinsischen `width`/`height`-Attribute am `<img class="logo">` (697×316), in
  Kombination mit globalem `img { height: auto }`. Fix: Attribute setzen, im CSS
  nur **eine** Dimension vorgeben (`height` + `width:auto`), Seitenverhältnis
  absichern, rendern lassen.
- **`make website` nutzt Docker:** ggf. Dev-Stack vorher stoppen
  (`make down`), um Port-Konflikte zu vermeiden (siehe Repo-Notizen).
- **Überschneidung mit dem SEO/404-Plan** im `<head>`/Partials beachten, falls
  beide parallel umgesetzt werden.

---

## Phase 1: Saubere CSS-Grundlage + globales Chrome

### Kontext

- [website/css/base.css](../../website/css/base.css#L30) — `:root`-Tokens; Header
  [L232](../../website/css/base.css#L232), `.logo`
  [L256](../../website/css/base.css#L256), `.main-nav`
  [L293](../../website/css/base.css#L293); dekorative Muster: Card-Glow-`::before`
  ≈L690, Verläufe in Compliance/How-it-works/Pricing/CTA
  ([L556](../../website/css/base.css#L556),
  [L1099](../../website/css/base.css#L1099)).
- [website/partials/header.html](../../website/partials/header.html) — Logo ohne
  Maße, 8 Nav-Links + Byline.
- [website/partials/mobile-nav.html](../../website/partials/mobile-nav.html),
  [website/partials/footer.html](../../website/partials/footer.html),
  [website/partials/dev-banner.html](../../website/partials/dev-banner.html).

### Was umgesetzt wird

Das gemeinsame Fundament: Design-Tokens auf das kanonische Grün konsolidieren
und das visuelle Vokabular reduzieren (dekorative Verläufe, schwere Schatten,
Glow-Pseudoelemente und Emoji-Icon-Styling entfernen; eine ruhige, konsistente
Typo-/Spacing-Skala). Gleichzeitig das geteilte Chrome aufräumen und die beiden
Responsive-Bugs beheben:

- **Header-Logo entzerren:** intrinsische Maße am `<img>` setzen, im CSS nur eine
  Dimension vorgeben, korrektes Seitenverhältnis auf allen Viewports.
- **Navbar entlasten:** von 8 auf wenige zentrale Einträge reduzieren
  (Simplifizierung), Breakpoint/Spacing so anpassen, dass kein horizontaler
  Overflow entsteht; die „Entwickelt von"-Byline aus dem Header entfernen
  (Clutter; steht weiterhin im Footer). Mobile-Nav entsprechend angleichen.

Da die Partials geteilt sind, wirkt das sofort auf allen drei Seiten.

### Akzeptanzkriterien

- [x] Tokens enthalten **kein Violett/Indigo**; Primär-/Akzentgrün entspricht der
      App-CI (Grün-Hue ≈165).
- [x] Header-Logo wird auf Mobile **und** Desktop unverzerrt im korrekten
      Seitenverhältnis dargestellt (kein Stauchen/Strecken, kein Layout-Shift).
- [x] Die Navigation läuft auf keinem Viewport zwischen 320 px und 1440 px über
      (kein horizontaler Overflow); Anzahl der Nav-Einträge reduziert.
- [x] Dekorative Verläufe, Glow-`::before` und schwere Schatten im globalen CSS
      entfernt oder deutlich reduziert; Header/Footer/Dev-Banner wirken ruhig.
- [x] Alle drei Seiten laden weiterhin korrekt (SSI-Partials intakt), `make website`
      rendert ohne Konsolenfehler.

---

## Phase 2: Startseite (`index.html`) entschlacken & konsolidieren

### Kontext

- [website/index.html](../../website/index.html#L29) — die 12 Abschnitte (siehe
  Inventar). Stärkste Komplexität: FEATURES
  [L158](../../website/index.html#L158) (4 Gruppen, ~22 Emoji-Cards), VERGLEICH
  [L481](../../website/index.html#L481), TECHNIK
  [L731](../../website/index.html#L731).
- Überschneidungen: PROBLEM [L72](../../website/index.html#L72) ↔ VERGLEICH
  [L481](../../website/index.html#L481); SOLUTION
  [L112](../../website/index.html#L112) ↔ SICHERHEIT
  [L625](../../website/index.html#L625) (beide Vorteils-Check-Listen); KOSTEN
  [L673](../../website/index.html#L673) ↔ TECHNIK-Hosting
  [L731](../../website/index.html#L731).

### Was umgesetzt wird

Die Landingpage von 12 auf ~7–8 ruhige Abschnitte verdichten und das Design
vereinheitlichen. **Empfohlene Ziel-Informationsarchitektur** (Feinschnitt der
Merges während der Umsetzung):

1. **Hero** (verschlankt).
2. **Warum jotti** — Problem + Vergleich zusammenführen (Schmerzpunkte +
   leichte Vergleichstabelle).
3. **Funktionen** — Features straffen: gruppierte, ruhige Textlisten statt
   Emoji-Card-Raster.
4. **Fiskalkonform** — inkl. Sicherheit/Unveränderbarkeit (Security hier
   einbetten), Bon-Bild + ein Hinweis-Callout, ein CTA zum Leitfaden.
5. **So funktioniert's** — kompakte 3 Schritte (ohne schwere Karten).
6. **Für wen** — Zielgruppen + Veranstaltungstypen (Tag-Cloud leicht).
7. **Selbst betreiben & Kosten** — Technik + Kosten zusammenführen (Specs als
   leichte Tabelle, Schnellstart-Codeblock, „kostenlos"-Aussage), CTA zum
   Hosting-Leitfaden.
8. **Abschluss-CTA** (eine starke CTA statt mehrerer Wiederholungen).

Quer durch: **dekorative Emoji-Icons entfernen**, Cards entkasteln, Bold auf das
Nötige reduzieren, eine konsistente Abschnitts-/Tabellen-/Callout-Behandlung. Die
inhaltliche Substanz (Funktionsumfang, Fiskal-Aussagen, „In Entwicklung"-Hinweise,
Specs) bleibt vollständig erhalten.

### Akzeptanzkriterien

- [x] Abschnittszahl spürbar reduziert (~12 → ~7–8); keine wesentliche Information
      verloren (Funktionen, Fiskalkonformität inkl. WIP-Stand, Specs, Kosten,
      Zielgruppe weiterhin vorhanden).
- [x] **Keine dekorativen Emoji-Icons** mehr in Karten/Listen der Startseite.
- [x] Deutlich weniger Box-/Card-Rahmen und Bold; Vergleichs- und Spec-Daten als
      **leichte** Tabellen erhalten.
- [x] Genau **ein** Token-Set/CI durchgängig (Grün), keine Sonderverläufe pro
      Abschnitt.
- [x] Lesefluss auf Mobile und Desktop geprüft; kein Overflow, ruhiger
      Weißraum-Rhythmus.

---

## Phase 3: Leitfäden (beide Unterseiten) entschlacken

### Kontext

- [website/jotti-selbst-betreiben/index.html](../../website/jotti-selbst-betreiben/index.html) —
  Emoji-Callouts (💻🖥️🔒💡), `.card-grid`-Intros, Daten-/Glossar-Tabellen.
- [website/leitfaden-fuer-vereine/index.html](../../website/leitfaden-fuer-vereine/index.html) —
  „Das Wichtigste"-Emoji-Cards (🔐🏛️📦), Verantwortungs-/Fristen-Tabellen,
  `.guide-faq`.
- Gemeinsame Stile `.guide-prose`/`.guide-chapter`/`.compliance-note`/
  `.comparison-table` in [base.css](../../website/css/base.css#L1280) (≈L1280–1426)
  — einmal anpassen, beide Seiten profitieren.

### Was umgesetzt wird

Beide Leitfäden auf ruhigen, gut lesbaren Fließtext umstellen — sie sind
content-lastig, und das ist erwünscht. Dekorative Emoji in Callouts und den
„Das Wichtigste"-Karten entfernen, die einleitenden `.card-grid`-Kästen zu
schlichtem Text/kurzen Listen auflösen, Bold-Dichte reduzieren. **Echte
Datentabellen** (Server-Specs, ELSTER-Fristen, Verantwortungsmatrix, Glossar)
bleiben als **leichte** Tabellen. `.compliance-note`-Callouts bleiben als ruhiges
Hinweis-Element (leichter: ein Akzentrand statt schwerer Fläche). Inhalt
(rechtliche/technische Aussagen, Schritt-für-Schritt-Anleitungen) bleibt
vollständig.

### Akzeptanzkriterien

- [ ] Keine dekorativen Emoji mehr in Callouts/Karten beider Leitfäden.
- [ ] Einleitende Deko-`card-grid`-Kästen zu Text/leichten Listen aufgelöst;
      echte Datentabellen als leichte Tabellen erhalten.
- [ ] Callouts (`.compliance-note`) einheitlich und leicht gestaltet; Bold
      reduziert.
- [ ] Beide Seiten teilen dieselbe ruhige Typo-/Abstands-Behandlung wie die
      Startseite (Konsistenz).
- [ ] Mobile und Desktop geprüft; lange Tabellen scrollen horizontal ohne
      Layout-Bruch.

---

## Phase 4: Responsive- & Konsistenz-QA (alle Seiten)

### Kontext

- Alle drei Seiten über [Makefile](../../Makefile#L214) `make website` →
  `http://localhost:8080`.
- Breakpoints im CSS u. a. bei 960 px und 640 px (siehe Header/Hero/Card-Grid).

### Was umgesetzt wird

Abschließende Querprüfung der gesamten Website auf Mobile (≤640 px), Tablet
(~768/960 px) und Desktop (≥1120 px). Geprüft werden: kein horizontaler
Overflow, korrektes Logo-Seitenverhältnis, bedienbare Navigation (Desktop + Mobile
Overlay), Kontrast/Lesbarkeit (AA), einheitliche Abstände/Typo zwischen den
Seiten, konsistente Button-/Link-/Tabellen-/Callout-Darstellung. Verbleibende
Abweichungen werden behoben.

### Akzeptanzkriterien

- [ ] Auf Mobile, Tablet und Desktop kein horizontaler Overflow auf allen drei
      Seiten.
- [ ] Logo, Navigation (inkl. Mobile-Overlay) und Footer auf allen Viewports
      korrekt und bedienbar.
- [ ] Text-Kontrast erfüllt WCAG AA; Fließtext angenehm lesbar.
- [ ] Abstände, Typografie, Buttons, Tabellen und Callouts seitenübergreifend
      konsistent.
- [ ] Visuelle Endkontrolle: deutlich „cleaner" als vorher (weniger Boxen,
      Icons, Bold, Verläufe).

---

## Phase 5: Violett-Rest-Cleanup (Code-Kommentare + Marken-Doku)

### Kontext

- [frontend/src/index.css](../../frontend/src/index.css#L52-L100) — Werte grün,
  aber Kommentare sagen `/* violet-600 */`, `/* indigo-600 */` etc.
- [assets/assets-and-design.md](../../assets/assets-and-design.md) — Abschnitt 5
  „Das Theme (Farben)" dokumentiert die veraltete violette CI.

### Was umgesetzt wird

Kleiner, risikoarmer Aufräumschritt außerhalb des `website/`-Ordners (vom Nutzer
ausdrücklich gewünscht: „Violett kann überall entfernt werden"). Die irreführenden
Violett-Kommentare in `frontend/src/index.css` auf die tatsächlichen grünen
Farbwerte korrigieren und die Farbsektion in `assets/assets-and-design.md` auf die
grüne CI aktualisieren. **Kein Verhaltens- oder Wert-Change im Frontend** — nur
Kommentare/Doku. Bild-Assets bleiben hier ausdrücklich unberührt (separates
Follow-up).

### Akzeptanzkriterien

- [ ] Keine `violet`/`indigo`-Kommentare mehr in `frontend/src/index.css`, die
      grünen Werten widersprechen.
- [ ] `assets/assets-and-design.md` beschreibt die grüne CI (keine violetten
      Hex-/Token-Angaben als „aktuell").
- [ ] Keine funktionalen Änderungen am Frontend (nur Kommentare); `make lint`
      bleibt grün.
