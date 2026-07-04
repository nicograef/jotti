# Plan: Regenbogen-Branding für Logo und Website

> Source PRD: ../prds/prd-regenbogen-branding.md

## Goal

Das gefaltete J bekommt einen kontinuierlichen Spektral-Verlauf (Rot bis Violett) und ersetzt als neues
kanonisches Logo die grünen Master. Die Website übernimmt alle neuen Assets, bekommt dekorative
Spektrum-Elemente (Hero, Haarlinie) bei unverändert grüner Bedienoberfläche und wird strukturell wie
sprachlich für nicht-technische Vereinshelfer gestrafft. Die App bleibt unangetastet.

## Architectural decisions

- **Spektral-Generator**: eigenständiges, versioniertes Python-3-Skript (Pillow, keine weiteren
  Abhängigkeiten) im Verzeichnis `scripts/`. Ortsabhängige Hue-Zuordnung im OKLCH-Farbraum entlang der
  Verlaufsachse des J; Helligkeit und Schattierung stammen aus den Mastern, Chroma wird ans sRGB-Gamut
  geklemmt, Neutraltöne und Alpha bleiben unangetastet. Checks sind in das Skript eingebaut.
- **Kanonische Master**: die Spektral-Varianten ersetzen die grünen Dateien in `assets/` unter
  unveränderten Dateinamen. Website-Kopien bleiben byte-identisch zu den Mastern.
  `frontend/public/icons/` bleibt grün, bis die App in einem eigenen Vorhaben nachzieht.
- **Dekorative Tokens**: Spektrum-Verlaufstoken kommen ins Marken-Token-Set (`brand.css`) mit Light- und
  Dark-Ausprägung. Alle semantischen UI-Tokens der Landing und des Doku-Themes bleiben unverändert grün.
- **Ziel-Sektionsstruktur der Landing** (11 auf 8 Sektionen): Hero, Für wen, Nutzen (aus Versprechen und
  Warum zusammengelegt), Features, Einblicke, So funktioniert's, Service, Technik und Compliance
  (zusammengelegt), CTA.
- **Anker-Politik**: bestehende IDs `versprechen`, `features`, `compliance`, `technik` bleiben
  funktionsfähig (Navigation und externe Links). Die zusammengelegte Technik-und-Compliance-Sektion trägt
  `id="compliance"`, `technik` bleibt als zusätzlicher Anker erhalten. `warum` entfällt, die Sektion
  "Für wen" bekommt eine eigene ID.
- **Abnahme-Gates**: kein Master und keine Landing-Änderung geht ohne visuelle Freigabe durch Nico live.

## Inventory

- `assets/jotti-*.png` — 12 grüne Master (Full-Logo hell/dunkel/transparent, Icons hell/dunkel, Favicons
  16/32/64, Symbol); Stand nach der Grün-Umfärbung vom 2026-07-04
- `assets/assets-and-design.md:38` — Design-Logik des J; `:47-60` Varianten-Tabelle; `:89` Hinweis auf die
  Umfärbung (wird zum Übergangszustand-Hinweis)
- `website/src/styles/brand.css:37-45` — Marken-Token-Set (Markengrün, Schrift); Ort der neuen
  Spektrum-Token
- `website/src/styles/landing.css:12-45` — semantische Farb-Token Light/Dark (bleiben grün); `:76-79`
  Selection; `:83-133` Buttons, Eyebrow, Icon-Box, Tag (bleiben grün); `:140-148` `.hero-gradient`
  (wird mehrfarbig)
- `website/src/styles/starlight.css:13-22` — Doku-Akzentfarben (bleiben grün)
- `website/src/layouts/Landing.astro:49-70` — Favicon- und apple-touch-icon-Links; `:5,124,263`
  Symbol-Einbindung Header/Footer; `:133-148` Desktop-Nav-Anker; `:216-231` Mobilmenü-Anker
- `website/src/pages/index.astro` — Sektionen: Hero `71-204`, Versprechen `205-332`, Warum `333-551`,
  Features `552-1038`, Einblicke `1039-1147`, Compliance `1148-1352`, So funktioniert's `1353-1436`,
  Für wen `1437-1528`, Technik `1529-1596`, Service `1597-1747`, CTA `1748-1801` (Symbol in `:1752`)
- `website/astro.config.mjs:22` — Starlight-Konfiguration ohne `favicon`-Option; der Starlight-Default
  `/favicon.svg` existiert nicht in `website/public/` (bestehende Lücke)
- `website/public/icons/` — 5 Icon-Kopien; `website/src/assets/jotti-symbol.png` — Symbol-Kopie
- `frontend/public/icons/` — 7 grüne Kopien, bleiben in diesem Vorhaben unverändert
- `website/package.json:12-16` — `pnpm dev/build/preview/check/test`
- `scripts/` — Repo-Tooling (bisher Shell), Zielort für den Generator
- Prior Art: konstante OKLCH-Hue-Rotation der Grün-Umfärbung (2026-07-04, dokumentiert in
  `assets/assets-and-design.md:89`); Headless-Firefox-Screenshots aus der Design-System-Migration für die
  visuelle Abnahme

## Resolved decisions

- Volles Spektrum im Logo als fließender Verlauf, keine Streifen, keine Sechs-Farben-Anordnung, Grün
  bleibt in der Mitte des Verlaufs (PRD).
- Website nur dekorativ bunt, UI-Tokens und CTAs bleiben grün (PRD).
- Spektral-Varianten werden kanonische Master; App/PWA bleibt vorerst grün (PRD).
- Straffung darf konsolidieren: Für wen nach vorn, Versprechen+Warum zusammen, Technik+Compliance
  zusammen, Features stark kürzen (Klärungsrunde zum Plan).
- Absicherung über Generator-Checks plus visuelle Abnahme; bewusst kein Link-/Anker-Validator im Build
  (PRD). `pnpm check` bleibt als normale Vollständigkeitsprüfung Teil der Phasen, die Astro-Dateien
  anfassen.
- Werte bleiben implizit, kein Werte-Text (PRD).

## Open questions / Risks

- Verlaufsachse und Richtung des Spektrums (z. B. Rot oben nach Violett unten oder entlang der
  Faltdiagonale) werden in Phase 1 am Ergebnis iteriert; das Risiko mehrerer Abnahmerunden ist
  eingeplant.
- Die Master sind weich schattierte Rasterbilder; eine ortsabhängige Hue-Zuordnung kann an Facettenkanten
  sichtbare Übergänge erzeugen. Gegenmittel: Achse an der Faltgeometrie ausrichten, in der Abnahme prüfen.
- 16-Pixel-Favicon: ein volles Spektrum auf wenigen Pixeln kann matschig wirken. Falls die Abnahme das
  zeigt, bekommen die kleinsten Größen eine reduzierte Tonzahl (Entscheidung in Phase 1).
- Druck (US 17): Verhalten des Spektral-Logos im Graustufen-Druck in der Abnahme von Phase 1 mitprüfen.

---

## Phase 1: Spektral-Generator und neue Logo-Master

**User stories**: 1, 2, 3, 4, 6, 14, 16, 17, 18

### Context

- `assets/jotti-*.png` — die 12 zu ersetzenden Master
- `assets/assets-and-design.md:38,47-60,89` — Farbbeschreibungen und Umfärbungs-Hinweis
- `scripts/` — Zielort des Generators

### What to build

Ein versioniertes Generator-Skript liest die grünen Master und erzeugt den kompletten Spektral-Satz in
einen Staging-Ordner (nicht direkt in `assets/`). Der Farbton wandert kontinuierlich von Rot nach Violett
entlang der Verlaufsachse des J; Schattierung, Neutraltöne und Alpha bleiben erhalten. Eingebaute Checks
brechen bei Verletzung ab. Achse, Richtung und Tonverteilung werden am Staging-Ergebnis iteriert, bis die
visuelle Abnahme (alle Varianten, hell/dunkel/transparent, 16-Pixel-Ansicht, Graustufen-Probe) erfolgt
ist. Erst nach Freigabe ersetzen die optimierten Dateien die Master, und die Asset-Doku wird nachgezogen
(Farbbeschreibungen, Übergangszustand App bleibt grün, Hinweis auf das weiterhin alte
Dokumentationsbild).

### Acceptance criteria

- [x] Generator läuft mit System-Python und Pillow, ohne weitere Abhängigkeiten, und ist im Repo
      versioniert
- [x] Eingebaute Checks: Hue-Spannweite deckt das Spektrum ab, Neutraltöne bleiben wertstabil, Alpha
      unverändert, alle 12 Varianten erzeugt, Dateigrößen in der Größenordnung der bisherigen Assets
- [x] Staging-Ergebnis von Nico visuell freigegeben (inkl. 16-Pixel- und Graustufen-Check), erst danach
      Master ersetzt
- [x] `assets/assets-and-design.md` beschreibt den Spektral-Stand und den Übergangszustand
- [x] `frontend/public/icons/` zeigt keinen Diff

---

## Phase 2: Website übernimmt das neue Logo

**User stories**: 2, 5, 6, 12, 15

### Context

- `website/public/icons/`, `website/src/assets/jotti-symbol.png` — zu ersetzende Kopien
- `website/src/layouts/Landing.astro:49-70,124,263` — Favicons und Symbol-Einbindung
- `website/src/pages/index.astro:1752` — Symbol im CTA
- `website/astro.config.mjs:22` — Starlight ohne `favicon`-Option

### What to build

Die Website-Kopien werden byte-identisch durch die neuen Master ersetzt (Favicons, apple-touch-icon,
Symbol in Header, Footer und CTA). Die Starlight-Doku bekommt eine explizite `favicon`-Konfiguration auf
ein vorhandenes Icon, damit die bestehende `/favicon.svg`-Lücke geschlossen ist und Landing und Doku
dieselbe Marke zeigen. Light- und Dark-Zustände werden per Screenshot geprüft und abgenommen.

### Acceptance criteria

- [x] Alle Website-Kopien byte-identisch zu den neuen Mastern (md5-Vergleich)
- [x] Doku-Seiten liefern ein konfiguriertes Favicon ohne 404
- [ ] Screenshots Header/Footer/CTA und Browser-Tab in Light und Dark von Nico freigegeben
- [x] `pnpm check` läuft fehlerfrei

---

## Phase 3: Dekorative Spektrum-Tokens, Hero und Haarlinie

**User stories**: 3, 5, 11

### Context

- `website/src/styles/brand.css:37-45` — Marken-Token-Set
- `website/src/styles/landing.css:12-45,140-148` — semantische Token und `.hero-gradient`
- `website/src/styles/starlight.css:13-22` — Doku-Akzente (unverändert lassen)

### What to build

Das Marken-Token-Set bekommt Spektrum-Verlaufstoken mit Light- und Dark-Ausprägung (gedeckte Chroma- und
Deckkraftwerte, kein greller Regenbogen). Der Hero-Hintergrund wird von rein grünem Tint auf einen
subtilen Mehrfarbverlauf umgestellt, und eine feine Spektral-Haarlinie kommt an höchstens zwei
wiederkehrende Stellen (z. B. Header-Unterkante und Footer). Buttons, Eyebrows, Icon-Boxen, Tags,
Selection und alle Doku-Akzente bleiben unverändert grün.

### Acceptance criteria

- [x] Spektrum-Token existieren in Light- und Dark-Ausprägung; kein semantisches UI-Token geändert
- [ ] Hero-Verlauf und Haarlinie in Light und Dark subtil (visuelle Abnahme durch Nico, Mobil und
      Desktop)
- [x] Stichprobe bestätigt: CTAs, Buttons, Tags und Doku-Akzente unverändert grün
- [x] `pnpm check` läuft fehlerfrei

---

## Phase 4: Landing-Umbau, Konsolidierung und Reihenfolge

**User stories**: 7, 8, 9, 10

### Context

- `website/src/pages/index.astro:205-551` — Versprechen und Warum (werden eine Nutzen-Sektion)
- `website/src/pages/index.astro:1148-1352,1529-1596` — Compliance und Technik (werden eine Sektion)
- `website/src/pages/index.astro:1437-1528` — Für wen (wandert hinter den Hero, bekommt eine ID)
- `website/src/layouts/Landing.astro:133-148,216-231` — Nav-Anker Desktop und Mobil

### What to build

Die Sektionen werden auf die Zielstruktur gebracht: Für wen direkt hinter den Hero, Versprechen und Warum
zu einer Nutzen-Sektion unter `id="versprechen"`, Technik und Compliance zu einer kompakten Sektion unter
`id="compliance"` mit erhaltenem `technik`-Anker. Reihenfolge danach: Features, Einblicke, So
funktioniert's, Service, Technik und Compliance, CTA. Texte werden nur so weit angepasst, wie das
Zusammenlegen es erfordert (Redundanzen zwischen Versprechen und Warum entfallen). Navigation in Desktop-
und Mobilmenü wird nachgezogen.

### Acceptance criteria

- [ ] Zielstruktur mit 8 Sektionen umgesetzt; alle Nav-Links und die Anker `versprechen`, `features`,
      `compliance`, `technik` funktionieren
- [ ] Kein inhaltlicher Verlust außer bewusst entfernten Redundanzen (Wortstrom-Diff geprüft)
- [ ] Visuelle Abnahme der neuen Reihenfolge durch Nico (Mobil und Desktop)
- [ ] `pnpm check` läuft fehlerfrei

---

## Phase 5: Sprachliche Vereinfachung und Kürzung

**User stories**: 7, 8, 10

### Context

- `website/src/pages/index.astro` — alle Sektionen nach dem Umbau aus Phase 4; Features ist mit rund 490
  Zeilen die schwerste Sektion

### What to build

Alle Sektionstexte werden für nicht-technische Vereinshelfer vereinfacht: kurze Sätze, Fachbegriffe beim
ersten Auftreten erklärt oder gestrichen, konsistente Ansprache. Die Features-Sektion wird stark gekürzt
(Ziel: mindestens halbiert), die Landing insgesamt spürbar entlastet (Richtwert: rund ein Drittel weniger
Text). Es entsteht kein neuer Inhalt und kein Werte-Text. Abschließend erfolgt die Gesamtabnahme über
alle Zustände.

### Acceptance criteria

- [ ] Features-Sektion mindestens auf die Hälfte des Textumfangs reduziert; Landing gesamt messbar
      kürzer (Wortzählung vorher/nachher im Ergebnis dokumentiert)
- [ ] Keine unerklärten Fachbegriffe im Nutzen-Teil der Seite (Sichtprüfung)
- [ ] Kein neuer Inhalt, kein Werte-Text (Wortstrom-Diff geprüft)
- [ ] Finale Gesamtabnahme durch Nico: Light/Dark, Mobil/Desktop
- [ ] `pnpm check` läuft fehlerfrei
