# Handoff: jotti – Marketing-Website (Redesign)

## Overview
jotti ist ein kostenloses, self-hosted **Kassensystem für Vereinsfeste**. Helfer:innen
kassieren am eigenen Smartphone (BYOD) im Browser — Bestellung, Zahlung, Storno, alles pro
Tisch, mit integrierten fiskalischen Bausteinen (Cloud-TSE, Belegausgabe, DSFinV-K).

Dieses Bundle enthält die **Marketing-/Produkt-Website** (eine Single-Page-Site plus einen
„Für Vereine"-Unterview mit Anfrageformular). Ziel der Seite: Vereine informieren, Vertrauen
in die Rechtskonformität schaffen und zur kostenlosen Nutzungsanfrage bewegen.

## About the Design Files
Die Dateien in diesem Bundle sind **Design-Referenzen, umgesetzt in HTML** — ein interaktiver
Prototyp, der Aussehen und Verhalten zeigt. Es ist **kein produktiver Code zum 1:1-Kopieren**.

Die Aufgabe ist, dieses Design in der **bestehenden Umgebung des Zielprojekts** nachzubauen
(z. B. React/Next.js, Vue, Astro o. ä.) mit dessen etablierten Mustern, Komponenten und
Utility-/Token-Systemen. Falls noch keine Umgebung existiert: das für ein Marketing-Frontend
sinnvollste Framework wählen (z. B. Astro oder Next.js) und dort umsetzen.

> Technischer Hinweis: Der Prototyp ist als „Design Component" (`.dc.html`) gebaut und nutzt
> ein hauseigenes Template-/Logik-Runtime (`support.js`). **Dieses Runtime NICHT übernehmen.**
> Es dient nur dazu, den Prototyp lauffähig zu machen. Nachbauen: das Markup (Struktur,
> Styles, Copy) und die im Prototyp gezeigte Logik (State-Toggles, Interaktionen) in den
> Idiomen des Zielframeworks.

## Fidelity
**High-fidelity (hifi).** Finale Farben, Typografie, Abstände, Radien, Schatten und
Interaktionen sind ausgearbeitet und sollen pixelgenau mit den Mitteln des Ziel-Codebase
nachgebaut werden. Feste Pixelwerte im Prototyp sind bewusst gewählt; die Tokens unten sind
die Quelle der Wahrheit.

---

## Global / Layout-Grundlagen
- **Content-Breite:** `max-width: 1200px`, zentriert, seitliches Padding `24px`.
- **Vertikaler Sektionsrhythmus:** i. d. R. `padding: 92px 24px` pro Sektion.
- **Font-Loading:** Google Fonts — `Space Grotesk` (400/500/600/700) und `Inter`
  (400/500/600/700/800).
- **Zwei Views (clientseitig umgeschaltet, keine echte Route im Prototyp):**
  - `home` — die komplette Landing Page.
  - `vereine` — „Für Vereine"-Seite mit Anfrageformular + Erfolgs-State.
  Umschaltung über einen `page`-State. In der echten App als Routen umsetzen
  (`/` und `/fuer-vereine`).
- **Theme:** Hell + Dunkel, umschaltbar über Button in der Nav. Persistenz in
  `localStorage` unter Key `jotti-theme`. Initial: gespeicherter Wert → sonst
  `prefers-color-scheme`. Theme wird als Attribut `data-theme="light|dark"` am Wurzel-Container
  gesetzt; alle Farben laufen über CSS-Variablen (s. Design Tokens).
- **Reduced motion:** Bei `prefers-reduced-motion: reduce` alle Animationen praktisch
  deaktivieren.
- **Scroll-Reveal:** Sektionen faden per CSS `animation-timeline: view()` beim Scrollen ein
  (Klasse `.reveal`: 0.8s, `translateY(28px)` → 0, `cubic-bezier(.2,.7,.2,1)`).

### Responsive-Verhalten
Der Prototyp schaltet mehrspaltige Grids bei Breakpoints auf 1–2 Spalten. Breakpoints:
- **≤ 860px** (`--mqN`): Hero, Feature-Grid, Demo, Preis, Download, Vereine-Grid → 1 Spalte;
  Nav-Links ausblenden (Burger/Fallback); Ablauf-Steps → 2 Spalten; Compliance → 2 Spalten;
  Footer → 4→2/… Spalten.
- **≤ 560px** (`--mqXs`): Ablauf-Steps → 1 Spalte; Compliance → 1 Spalte; Footer → 2 Spalten.
- Die horizontale Verbindungslinie der Ablauf-Steps wird unter 860px ausgeblendet.

---

## Screens / Views

### View: Home

#### 1. Header / Nav (sticky)
- **Layout:** `position: sticky; top: 0; z-index: 50`. Hintergrund
  `color-mix(in srgb, var(--bg) 82%, transparent)` mit `backdrop-filter: saturate(1.4) blur(12px)`,
  unten `1px solid var(--border)`. Innen Flex-Row, `max-width:1200px`, `padding:14px 24px`, `gap:24px`.
- **Komponenten:**
  - **Logo (links):** Original-Bildmarke `assets/jotti-symbol.png` (Höhe ~41px, transparent,
    theme-neutral) + Wortmarke „jotti" als Text (Space Grotesk 700, 22px, `letter-spacing:-.02em`).
    Klick → Home + Scroll-to-top.
  - **Nav-Links (mittig, ab 860px sichtbar):** Funktionen, Ablauf, Für wen, Sicherheit, FAQ.
    14.5px, weight 500, Farbe `var(--muted-fg)`, Anker-Scroll zu Sektions-IDs.
  - **Theme-Toggle:** 40×40, `border-radius:11px`, `1px solid var(--border)`, `background:var(--card)`.
    Icon ☾ (light) / ☀ (dark).
  - **CTA „Für Vereine":** Höhe 40px, `padding:0 18px`, `border-radius:11px`,
    `background:var(--primary)`, `color:var(--primary-fg)`, weight 600. → View `vereine`.

#### 2. Hero
- **Layout:** Grid `1.05fr .95fr`, `gap:40px`, `padding:76px 24px 40px`, `align-items:center`.
  Dahinter (z-index 0) weiche Radial-Gradient-Blobs in sp-teal/sp-orange, `filter:blur(24px)`.
- **Linke Spalte:**
  - **Badge:** Pill mit spektralem Punkt + Text „Beta 1.0 · Kostenlos für Vereine". 12.5px/600,
    `border:1px solid var(--border)`, `background:var(--card)`, Schatten.
  - **H1:** Space Grotesk 700, `clamp(40px,5.4vw,68px)`, `line-height:1.02`, `letter-spacing:-.03em`.
    Text: „Das Kassensystem fürs **Vereinsfest.**" — das Wort **„Vereinsfest."** hat einen
    **animierten Spektral-Farbverlauf** (Text-Gradient `var(--spectral)`, `background-size:200% auto`,
    Keyframe `sheen` 6s linear infinite, Hintergrundposition -140%→240%).
  - **Sub-Copy:** `clamp(16px,1.5vw,19px)`, `line-height:1.6`, `color:var(--muted-fg)`, `max-width:30em`.
  - **Buttons:** Primär „Für Verein anfragen →" (52px hoch, `border-radius:13px`, primary).
    Sekundär „▶ Live-Demo ansehen" (Outline, `background:var(--card)`, Anker → `#demo`).
  - **Trust-Row:** drei ✓-Punkte (0 € Softwarekosten / Keine Hardware-Bindung / BYOD), Häkchen in
    sp-green/sp-teal/sp-blue.
- **Rechte Spalte — iPhone-Mockup (UI-Nachbau, KEIN Screenshot):**
  - Dahinter animierte Blobs (sp-orange, sp-violet, Keyframe `orb`) + dünne spektrale Bodenlinie.
  - **Telefon:** 300×610, `border-radius:46px`, `background:#0a0a08`, `padding:12px`,
    Schatten `0 40px 90px -26px rgba(0,0,0,.55)` + `0 0 0 2px color-mix(in srgb,var(--sp-teal) 22%,transparent)`.
    Notch 96×24. Schwebt (Keyframe `jfloat` 7s).
  - **Screen-Inhalt (Bestellansicht „Tisch 4 / Biergarten"):** App-Header mit jotti-Symbol,
    Tischname + „Offen 14,50 €"; Kategorie-Pills (Getränke aktiv = `var(--fg)`/`var(--bg)`, Speisen outline);
    Produktzeilen mit Mengensteppern (Minus-Kreis outline, Menge, Plus-Kreis primary); aktive Zeile
    hat primären Rahmen + `color-mix(...primary 6%...)`; unten grüner „Kassieren 14,50 €"-Balken (50px,
    `border-radius:13px`, primary, Euro-Icon).
  - **Wichtig:** Dieses Mockup nutzt dieselben Theme-Tokens wie die Seite → passt sich Hell/Dunkel an.

#### 3. Funktionen (interaktiver Feature-Explorer) — `#features`
- **Kopf:** Eyebrow „Funktionen" (spektraler Strich + 12.5px/700 uppercase, `letter-spacing:.14em`,
  `color:var(--sp-teal)`), H2 „Ein echtes Kassensystem. Kein Spielzeug." (Space Grotesk 700,
  `clamp(30px,3.7vw,46px)`), Intro-Absatz.
- **Layout:** Grid `1.15fr 1fr`, `gap:28px`. Links 2×3 Feature-Kacheln, rechts eine sticky
  Detail-Karte (`top:92px`).
- **Feature-Kacheln (6):** Button, `padding:18px`, `border-radius:16px`, `border:1px solid var(--border)`,
  `background:var(--card)`. Jede Kachel hat eine **Akzentfarbe**; aktive Kachel: farbiger Rahmen +
  `box-shadow: 0 0 0 1px <color>, var(--shadow)`, `translateY(-3px)`, `background: color-mix(<color> 7%, card)`,
  Icon-Box `color-mix(<color> 15%)`, Icon in `<color>`. Inhalt: Icon-Box 42×42 (`border-radius:12px`),
  Titel (Space Grotesk 600, 16.5px), Kurztext 13px/muted.
- **Detail-Karte:** `border-radius:20px`, `padding:30px`, oben 5px spektraler Streifen. Enthält
  großes Icon (58×58), H3 (25px), Beschreibung, 3 ✓-Punkte.
- **Die 6 Features (Titel · Akzent · Icon · Kurz · Detail · Punkte):**
  1. **Bestellungen** · sp-red · Icon `order` (Beleg/Bon) · „Pro Tisch aufnehmen" ·
     „Bestellungen auf Tische buchen — mit Produkten, Varianten, Steuersätzen und Kommentaren.
     Umbuchen auf andere Tische inklusive." · Produkte & Varianten / Steuersätze pro Position /
     Kommentare & Umbuchung
  2. **Zahlung & Storno** · sp-orange · Icon `pay` (Geldbörse — jotti ist **bargeldbasiert**,
     KEIN Karten-Terminal) · „Kassieren mit Rückgeld" · „Zahlungen kassieren, auch Teilzahlungen mit
     automatischer Rückgeldberechnung. Stornos nur für Admin & Serviceleitung — mit Pflichtkommentar."
     · Teilzahlung & Rückgeld / Rollenbasierter Storno / Pflichtkommentar
  3. **Direktverkauf** · sp-green · Icon `direkt` (Einkaufstasche) · „Theke ohne Tisch" ·
     „An der Theke bestellen, kassieren und ausgeben in einem Schritt — ohne Tisch, mit Historie und
     Storno." · Ein-Schritt-Verkauf / Volle Historie / Storno möglich
  4. **Küche & Bon-Druck** · sp-teal · Icon `kueche` (Drucker) · „Bons automatisch" ·
     „Bestell- und Küchenbons gehen automatisch an die zugeordneten ESC/POS-Bondrucker — pro Kategorie
     konfigurierbar." · ESC/POS-Drucker / Pro Kategorie / Automatischer Versand
  5. **Kasse & Abschluss** · sp-blue · Icon `kasse` (Registrierkasse) · „Sitzung bis Z-Bon" ·
     „Kassensitzungen eröffnen und schließen, Anfangsbestand erfassen, Kassensturz mit Differenz und
     formaler Tagesabschluss (Z-Bon)." · Kassensitzung / Kassensturz / Tagesabschluss (Z-Bon)
  6. **Reporting & Export** · sp-violet · Icon `report` (Balkendiagramm) · „Bis DSFinV-K" ·
     „Tagesabrechnung nach Steuersatz, Abrechnung pro Tisch und pro Servicekraft — plus
     maschinenlesbarer DSFinV-K-Export (v2.4)." · Umsätze nach Steuersatz / Pro Tisch & Person / DSFinV-K v2.4

#### 4. Live-Demo (interaktives Telefon) — `#demo`
- **Layout:** `background:var(--card2)`, Grid `1fr 1fr`, `gap:56px`.
- **Links:** Eyebrow „Live-Demo", H2 „Bestellen, kassieren, fertig.", 3 nummerierte Schritte
  (farbige Zahlen-Badges), Button „Demo neu abspielen".
- **Rechts — Telefon (328×660):** Voll interaktiver Nachbau der Bestellansicht. Verhalten:
  - **Auto-Demo:** Startet, wenn das Telefon in den Viewport scrollt (IntersectionObserver,
    threshold 0.25). Skript: Bratwurst +, Bier 0,5l ×2, Pommes +, dann „Kassieren", dann Reset
    (Loop-artig; Delays ~700–1900ms).
    Sobald der Nutzer manuell tippt (+/−/Kassieren), stoppt die Auto-Demo.
  - **Produkte antippen** erhöht Menge; aktive Zeilen bekommen primären Rahmen; Summe wächst live
    im Kopf und im Kassieren-Balken mit.
  - **Kassieren** blendet ein „Zahlung erfolgreich"-Overlay ein (Häkchen-Kreis mit `pop`-Animation,
    „Beleg signiert · TSE bestätigt").
  - **Demo-Menü/Preise:** Bier 0,5l 4,00 € / 0,3l 3,00 €; Weinschorle 0,25l 3,50 €; Bratwurst 3,50 €;
    Pommes 3,00 €.

#### 5. Ablauf („So funktioniert's") — `#ablauf`
- H2 „Von der Vorbereitung bis zum Z-Bon." + 4 Schritte in einer Reihe, verbunden durch eine
  horizontale spektrale Linie (unter 860px ausgeblendet). Jede Karte: 62×62 Icon-Kachel, „SCHRITT n"
  in Akzentfarbe, Titel, Beschreibung.
  1. **Einrichten** (sp-red) — Produkte/Varianten/Steuersätze, Tische, Helfer anlegen.
  2. **Helfer starten** (sp-green) — Adresse im Browser, Einmalpasswort, App aufs Handy (PWA, kein App Store).
  3. **Kassieren** (sp-teal) — bestellen, kassieren, stornieren, Küchenbons — pro Tisch.
  4. **Abschluss** (sp-violet) — Kassensturz mit Differenz, Z-Bon, DSFinV-K-Export.

#### 6. Für wen — `#fuerwen`
- `background:var(--card2)`. H2 „Gemacht für Vereine. Ehrlich über die Grenzen." + zwei Karten:
  - **Geeignet für** (grüner ✓-Header): e.V./gemeinnützige Orgs/NPOs · Vereins-/Sommerfeste,
    Weihnachtsmärkte, Maihocks, Konzerte · Teams von 1–30 Helfer:innen · Bargeld-Betrieb mit
    fiskalischen Bausteinen.
  - **Nicht geeignet für** (roter ✗-Header, gedämpfte Farbe): Dauerbetrieb (Restaurants/Cafés) ·
    Kartenzahlung/NFC/Online-Payment · kommerzielle Gastro (ohne separate Lizenz) · Betrieb ohne
    eigene TSE-Vereinbarung mit fiskaly.

#### 7. Jedes Gerät („Ein Design. Jedes Gerät.") — `#screenshots`
- H2 „Ein Design. Jedes Gerät." + Intro „… dieselbe klare, ruhige Oberfläche. Hell und dunkel."
- Drei UI-Nachbauten nebeneinander (`flex-wrap`, `align-items:flex-end`):
  1. **Desktop-Admin (Browser-Fenster, ~540px):** Traffic-Lights + URL-Pill `jotti.helferverein.de/admin`,
     linke Sidebar (jotti-Logo + Menü: Produkte aktiv, Tische, Benutzer, Kasse, Berichte), rechts
     „Produkte" mit Kategorie-Gruppen (Getränke 19 % / Speisen 7 %) und Varianten-Pills.
  2. **Phone „Meine Tische" (246×500) — IMMER HELL:** Dieses Telefon ist bewusst auf
     `data-theme="light"` fixiert, damit auch im Dark-Mode der Seite der Hell/Dunkel-Kontrast
     erhalten bleibt. (Beim Nachbau: dieses eine Mockup fest im Light-Token-Scope rendern,
     unabhängig vom globalen Theme.) Inhalt: Kopf „Meine Tische" + Avatar, Statistik-Kacheln
     (Bestellungen/Kassiert), Suchfeld, Listen „Noch offen" / „Erledigt" mit Status-Punkten
     (rot/gelb/grün), Footer-Button „Alle Tische".
  3. **Phone „Kassieren" (246×500) — IMMER DUNKEL:** fest `data-theme="dark"`. Inhalt:
     Kopf „Tisch 4 / kassieren", Auswahl-Zeilen mit Stepper, „Kassieren 12,50 €"-Balken.

#### 8. Preis — `#preis`
- `background:var(--card2)`, zentrierter Kopf „Kostenlos. Punkt." Grid `1.4fr 1fr`:
  - **Große Karte:** „0 €" (Space Grotesk 700, 64px, Spektral-Text) „/ für immer", 2×3 ✓-Liste
    (Alle Kassenfunktionen / Unbegrenzt Helfer:innen / Fiskalische Bausteine / DSFinV-K-Export /
    Self-hosted, deine Daten / Quellcode einsehbar), CTA „Kostenlose Nutzungsvereinbarung anfragen".
    Oben 6px spektraler Streifen.
  - **Nebenkarte „Was läuft laufend?":** Cloud-TSE (fiskaly, gesetzlich, Vertrag direkt mit fiskaly),
    Server (optional; fürs Fest reicht Windows-Rechner per Doppelklick).

#### 9. Sicherheit & Compliance — `#sicherheit`
- H2 „Auf die KassenSichV ausgelegt." + 3×2 Karten-Grid (`border-radius:16px`, je farbige Icon-Kachel):
  BSI-zertifizierte Cloud-TSE (sp-blue) · Belegausgabe § 146a AO (sp-teal) · GoBD-Kassenjournal
  (sp-green) · DSFinV-K-Export v2.4 (sp-violet) · Rollenmodell (sp-red) · Sicheres Onboarding (sp-orange).
- Darunter kleiner Rechts-Hinweis (13px/muted) zur Betreiberverantwortung.

#### 10. FAQ (Accordion) — `#faq`
- `background:var(--card2)`, `max-width:820px`. 7 Fragen als aufklappbare Items (`border-radius:15px`).
  Offenes Item: primärer Rahmen + `color-mix(primary 4%, card)`, Plus-Icon rotiert 45° (→×) in primary.
  Nur die geöffnete Frage zeigt die Antwort (Default: erste offen). Fragen: Kosten · TSE-Konformität ·
  Hardware · eigene Handys/PWA · Installation · Zielgruppe · Quellcode/Source-available.

#### 11. Download-CTA — `#download`
- Große Karte (`border-radius:26px`) mit weichen Radial-Gradients. Links H2 „In 5 Minuten startklar."
  + Beta-Hinweis-Pill. Rechts drei Link-Karten:
  - **Windows-Release laden** (primary) → `https://github.com/nicograef/jotti/releases`
  - **Leitfaden für Vereine** → `https://jotti.rocks`
  - **Quellcode auf GitHub** → `https://github.com/nicograef/jotti`

#### 12. Footer
- `background:var(--card2)`, Grid `1.6fr 1fr 1fr 1fr`. Spalte 1: Logo (Symbol + „jotti") +
  Beschreibung + 180×4px spektraler Balken. Spalten: **Produkt** (Funktionen/Live-Demo/Ablauf/Preis/FAQ),
  **Ressourcen** (Leitfaden/Installation/Compliance/Releases/GitHub), **Rechtliches**
  (Lizenzmodell/Nutzungsbedingungen/Für Vereine).
- Untere Leiste: „© 2025–2026 Nico Gräf. Alle Rechte vorbehalten. Source-available." +
  „Gebaut in Deutschland" (mit kleinem spektralem Punkt).
  > Hinweis: Es darf **nirgends** das Wort „Regenbogen" o. Ä. vorkommen. Die spektralen Farben
  > (Verlauf) sind das Markenelement und bleiben — nur die textliche Benennung ist verboten.

### View: Für Vereine (`/fuer-vereine`)
- „Zurück"-Button (→ Home). Kopf „Kostenlos für deinen Verein." + Erklärabsatz.
- Grid `1fr 1.05fr`: links „Das ist drin" (4 ✓-Punkte) + „So läuft's" (3 nummerierte Timeline-Schritte:
  Anfrage senden → Vereinbarung in Textform → Loslegen). Rechts **Formular** „Nutzungsvereinbarung
  anfragen" (oben 5px spektraler Streifen): Felder Verein/Organisation, Ansprechpartner:in, E-Mail,
  Rechtsform (Select: e.V. / Gemeinnützige Stiftung / NGO-NPO / Sonstige), Nachricht (optional),
  Submit „Anfrage senden".
- **Nach Absenden:** Erfolgs-State mit `pop`-animiertem Häkchen, „Danke für deine Anfrage!" +
  Download-/Startseite-Buttons. (Im Prototyp kein echter Versand — Backend/Formhandler in der App ergänzen.)

---

## Interactions & Behavior
- **Theme-Toggle:** wechselt light/dark, schreibt `localStorage['jotti-theme']`, setzt
  `data-theme` am Root. Initial aus localStorage oder `prefers-color-scheme`.
- **View-Wechsel Home ↔ Vereine:** in der echten App als Routen; scrollt bei Wechsel nach oben.
- **Feature-Explorer:** Klick auf Kachel setzt `feature`-Index → Detail-Karte + aktive Kachel-Styles.
- **FAQ-Accordion:** Klick toggelt offenes Item (nur eines offen; erneuter Klick schließt).
- **Live-Demo:**
  - IntersectionObserver startet Auto-Sequenz beim ersten Sichtbarwerden.
  - Manuelle Interaktion stoppt Auto-Demo dauerhaft (bis „Demo neu abspielen").
  - +/− ändern Mengen; Summe live; „Kassieren" → Erfolgs-Overlay; „Demo neu abspielen" resettet.
- **Formular:** `submit` → Erfolgs-State (Frontend). In Produktion: Validierung (required-Felder,
  E-Mail-Format) + echter Versand.
- **Animationen (Keyframes):** `sheen` (Text-Gradient, 6s), `jfloat` (Hero-Phone schweben, 7s),
  `orb` (Blobs, 9–11s), `pop` (Erfolgs-Häkchen, ~0.5s), `reveal` (Scroll-Einblenden, 0.8s),
  Accordion-Chevron-Rotation (0.25s). Alle unter `prefers-reduced-motion` neutralisieren.

## State Management
- `theme`: 'light' | 'dark' (persistiert).
- `page`: 'home' | 'vereine' (→ Route).
- `feature`: number (0–5), aktive Feature-Kachel.
- `faq`: number, Index des offenen FAQ-Items (-1 = alle zu).
- `demoCart`: `{ [variantId]: qty }`; abgeleitet: Summe, Menü mit Mengen/aktiv.
- `demoPaid`: boolean (Erfolgs-Overlay).
- `demoAuto`: boolean (läuft Auto-Demo noch?).
- `formSent`: boolean; `formName`: string (nur Anzeige).

## Design Tokens

### Farben — Light (`:root`)
| Token | Wert |
|---|---|
| `--bg` | `#ffffff` |
| `--fg` | `#0c0c09` |
| `--card` | `#ffffff` |
| `--card2` | `#fbfbf9` |
| `--muted` | `#f4f4f0` |
| `--muted-fg` | `#5b5b4f` |
| `--border` | `#e7e7df` |
| `--primary` | `#007a55` |
| `--primary-fg` | `#ffffff` |
| `--ring` | `rgba(0,122,85,.35)` |

### Farben — Dark (`[data-theme="dark"]`)
| Token | Wert |
|---|---|
| `--bg` | `#0c0c09` |
| `--fg` | `#fbfbf9` |
| `--card` | `#1d1d16` |
| `--card2` | `#161610` |
| `--muted` | `#2b2b22` |
| `--muted-fg` | `#a7a79a` |
| `--border` | `rgba(255,255,255,.10)` |
| `--primary` | `#10b981` |
| `--primary-fg` | `#04231a` |
| `--ring` | `rgba(16,185,129,.4)` |

### Spektralpalette (Markenakzent — Verlauf & Einzelfarben)
Light: sp-red `#d24a2a` · sp-orange `#c8781e` · sp-yellow `#8f9a2c` · sp-green `#4f9636` ·
sp-teal `#1f9b8a` · sp-blue `#2f6fc4` · sp-indigo `#4a4fc0` · sp-violet `#8b3fc0`.
Dark-Overrides: sp-red `#e2603f` · sp-green `#5fb045` · sp-teal `#28b8a3` · sp-blue `#4b86dd` ·
sp-violet `#a457dd`.
- `--spectral` = `linear-gradient(100deg,#d24a2a,#c8781e,#4f9636,#1f9b8a,#2f6fc4,#8b3fc0)`
- `--spectral-v` = vertikale Variante.

### Typografie
- Headings / Brand: **Space Grotesk** (700 meist; 600 für kleinere Headings).
- Body / UI: **Inter** (400/500/600; 800 verfügbar).
- H1 `clamp(40px,5.4vw,68px)` / `line-height:1.02` / `letter-spacing:-.03em`.
- H2 `clamp(30px,3.7vw,46px)` / `line-height:1.06` / `letter-spacing:-.025em`.
- Eyebrow: 12.5px / 700 / uppercase / `letter-spacing:.14em` / `color:var(--sp-teal)`.
- Body groß 17px/1.6; UI-Text 13–15px.
- Zahlen: `font-variant-numeric: tabular-nums`.

### Radius
Pills/Buttons klein 11–13px · Kacheln 16px · Karten 20–22px · große CTA-Karte 26px ·
Telefon-Bezel 38–46px · runde Stepper 50%.

### Schatten
`--shadow` light: `0 1px 2px rgba(12,12,9,.06), 0 12px 32px -12px rgba(12,12,9,.14)`.
`--shadow` dark: `0 1px 2px rgba(0,0,0,.4), 0 18px 40px -14px rgba(0,0,0,.6)`.

### Spacing
Sektion `92px` vertikal · Content-Padding `24px` · Grid-Gaps 12–56px · Kacheln-Padding 18–34px.

## Assets
Alle im Ordner `assets/` (Originale, transparent wo sinnvoll):
- `jotti-symbol.png` (537×825, transparent) — **Bildmarke „J"**, theme-neutral. Für Header, Footer,
  App-Header in den Mockups.
- `jotti-logo-full-light.png` / `jotti-logo-full-dark.png` (1024×512) — vollständiges Logo
  (Symbol + Wortmarke) für helle/dunkle Flächen (falls einteiliges Logo gewünscht).
- `jotti-logo-icon-light.png` / `jotti-logo-icon-dark.png` (1024×1024) — App-Icon-Variante
  (Symbol auf abgerundetem Hintergrund).
- `jotti-icon-light-32.png` — Favicon-Größe.
> Wichtig (Review-Vorgabe): **keine eigenen Logo-Nachbauten** (SVG etc.) mehr verwenden — immer
> das Original aus `assets/`.
- **Icons (UI):** Inline-SVG-Symbolset (Stroke-Icons, `stroke-width:2`, `currentColor`) im Prototyp
  definiert (order/pay/direkt/kueche/kasse/report/shield/lock/doc/phone/check/x/arrow/euro/users/
  download/plus/minus/github). Im Zielprojekt durch die dortige Icon-Library ersetzen (z. B. Lucide;
  die Namen entsprechen weitgehend Lucide-Icons). Bei den Feature-Icons unbedingt die Bedeutung
  wahren: Bestellung = Beleg/Bon, Zahlung = Geldbörse/Bargeld (NICHT Kartenzahlung), Direktverkauf =
  Einkaufstasche.

## Files
- `jotti Website.dc.html` — der komplette Prototyp (Markup + Styles + Logik). Referenz für Struktur,
  exakte Copy, Styles und Verhalten.
- `support.js` — Runtime des Prototyps. **Nicht übernehmen**, nur damit `jotti Website.dc.html` lokal
  im Browser läuft.
- `assets/` — Original-Logos/Icons (s. o.).

## Screenshots (`screenshots/`)
Referenzaufnahmen des gerenderten Designs (obere Bildschirmhöhe je Sektion; hohe Sektionen sind
ggf. abgeschnitten — Details stehen im HTML-Prototyp).

**Hell:**
- `01-light.png` — Hero (Header + iPhone-Mockup, „Vereinsfest." in Spektralfarben)
- `02-light.png` — Funktionen (Feature-Explorer, korrekte Icons)
- `03-light.png` — Live-Demo (interaktives Telefon)
- `04-light.png` — Ablauf (4 Schritte)
- `05-light.png` — Für wen (Geeignet / Nicht geeignet)
- `06-light.png` — Ein Design. Jedes Gerät. (Desktop-Admin + 2 Phones)
- `07-light.png` — Preis
- `08-light.png` — Sicherheit & Compliance
- `09-light.png` — FAQ
- `10-light.png` — Download-CTA

**Dunkel:**
- `01-dark.png` — Hero
- `02-dark.png` — Funktionen
- `03-dark.png` — Jedes Gerät (belegt: mittleres „Meine Tische"-Phone bleibt hell)
- `04-dark.png` — FAQ

### So lässt sich der Prototyp lokal ansehen
`jotti Website.dc.html` in einem lokalen Webserver (nicht `file://`) öffnen, z. B.:
```
cd design_handoff_jotti_website
python3 -m http.server 8000
# → http://localhost:8000/jotti%20Website.dc.html
```
