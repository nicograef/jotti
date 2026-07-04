# Corporate Identity & Design-Dokumentation: jotti

## 1. Einführung & Markenbeschreibung

**Markenname:** jotti  
**Slogan:** Das kostenlose Kassensystem für Vereinsfeste.

„jotti" ist ein kostenloses, quelloffenes, selbst gehostetes Gastronomie-Kassensystem (mPOS), das speziell für Vereine und gemeinnützige Organisationen entwickelt wurde. Es ist die ideale Lösung für Vereinsfeste, Weihnachtsmärkte, Konzerte, Maihocks und Sommerfeste.

### Kernwerte & Markenpersönlichkeit

- **Frei & Zugänglich:** Kostenlos, Open Source, self-hosted, keine Hardware-Bindung, keine laufenden Kosten.
- **Gemeinschaftsorientiert:** Von einem Entwickler für Vereine gebaut.
- **Modern & Effizient:** Mobile-first, Bestellungen auf eigenen Smartphones (BYOD), Echtzeit-KDS.
- **Sicher & Transparent:** Rollenmodell, Event-Sourcing für lückenlose Historie (GoBD-Grundsätze), JWT-Auth.
- **Professionell & Zuverlässig:** Ein echtes, funktionierendes Kassensystem, kein Spielzeug.

---

## 2. Design-Philosophie & Visueller Ton

Der visuelle Stil von jotti muss die Kernwerte widerspiegeln.

- **Klarheit & Lesbarkeit:** Da Servicekräfte oft in hektischen Umgebungen auf eigenen Smartphones arbeiten, müssen Icons, Text und UI-Elemente sofort verständlich und kontrastreich sein.
- **Modernität:** Der „faltbare" Stil des `J` steht für Dynamik, Papier (Bons), Fluss und Fortschritt. Es ist modern und vertrauenswürdig.
- **Minimalismus:** Ein klares, unaufdringliches Design, das sich auf die Aufgabe konzentriert.
- **Flexibilität:** Der Einsatz im Browser auf unterschiedlichsten Geräten (Smartphone, Tablet, Desktop) erfordert ein responsives Design, das sich im Logo widerspiegelt (Favicons, isoliertes Symbol).

---

## 3. Das Logo-System

Das jotti Logo-System basiert auf dem stilisierten, gefalteten `J` und dem Wortmarken-Schriftzug. Die korrekte Anwendung der verschiedenen Varianten ist entscheidend für ein konsistentes Markenbild.

### 3.1. Das Kernsymbol (Das gefaltete `J`)

- **Beschreibung:** Ein stilisiertes `J`, das durch Faltungen räumliche Tiefe und Dynamik erhält.
- **Design-Logik:** Die obere Hälfte hat einen Emerald-Verlauf, die untere Basis einen dunkleren Grün-Verlauf.

### 3.2. Der Schriftzug (Wortmarke)

- **Beschreibung:** Das Wort „jotti" in Kleinschreibung, sans-serif, modern und geometrisch.
- **Schriftart (Empfehlung):** Die Logos nutzen eine saubere, serifenlose Schrift wie „Inter" oder „Montserrat".

### 3.3. Die Logo-Varianten und ihr Einsatz

| Asset                                         | Beschreibung                         | Erscheinungsbild                                                                                                        | Primärer Einsatzort                                                   | Wann & Wo einsetzen                                                                                         |
| --------------------------------------------- | ------------------------------------ | ----------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| 1 (`jotti-logo-examples-and-definitions.png`) | Beispiel und Definition aller Assets | Alle Varianten neben bzw. untereinander mit Angaben von Farbe usw.                                                      | Dokumentation                                                         | Nirgendwo.                                                                                                  |
| 2 (`jotti-logo-icon-dark.png`)                | Dark Icon                            | Quadrat, abgerundet. Hintergrund: Slate-950 (`#020617`). `J` Verlauf ca. Emerald-500 → Green-800.                           | App Store, Launcher Icon, Desktop Shortcut                            | Quadratisches Icon auf dunklem Hintergrund (z. B. PWA-Piktogramm auf einem dunklen Smartphone-Hintergrund). |
| 3 (`jotti-logo-icon-light.png`)               | Light Icon                           | Quadrat, abgerundet. Hintergrund: Slate-50 (`#f8fafc`). `J` Verlauf ca. Emerald-600 → Green-800 (reichhaltiger).            | App Store, Launcher Icon, Desktop Shortcut                            | Quadratisches Icon auf hellem Hintergrund.                                                                  |
| 4 (`jotti-logo-full-dark.png`)                | Dark Full Logo                       | Isoliert auf Slate-950 (`#020617`). `J` (wie #2) links. Text „jotti" (Slate-50, `#f8fafc`) rechts.                      | Webseite (Header), Admin-Bereich (Dark Mode), Marketing (Flyer)       | Primäre Markenplatzierung auf dunklem Hintergrund.                                                          |
| 5 (`jotti-logo-full-light.png`)               | Light Full Logo                      | Isoliert auf Slate-50 (`#f8fafc`). `J` (wie #3) links. Text „jotti" (Slate-900, `#0f172a`) rechts.                      | Webseite (Header), gedruckte Abrechnungen, Admin-Bereich (Light Mode) | Primäre Markenplatzierung auf hellem Hintergrund. Standardwahl für gedruckte Dokumente.                     |
| 6 (`jotti-symbol.png`)                        | J Symbol                             | Isoliert auf transparentem Hintergrund. Reines gefaltetes `J` (reichhaltiger Verlauf wie #3), ohne Hintergrund-Quadrat. | UI-Icon, Marketing, PWA-Titelzeile                                    | Für allgemeine UI-Symbole, in der Titelzeile der PWA, oder über Bildmaterial im Marketing.                  |
| 7 (`jotti-icon-light-16.png`)                 | Light Favicon 16×16                  | Transparenter Hintergrund. J-Symbol, Light-Verlauf (ca. Emerald-600 → Green-800).                                           | Browser-Tab-Favicon                                                   | Smallest Favicon für Browser-Tab.                                                                           |
| 8 (`jotti-icon-light-32.png`)                 | Light Favicon 32×32                  | Transparenter Hintergrund. J-Symbol, Light-Verlauf.                                                                     | Browser-Tab-Favicon, Lesezeichen                                      | Standard-Favicon für Browser-Tab und Lesezeichen-Ansicht.                                                   |
| 9 (`jotti-icon-light-64.png`)                 | Light Favicon 64×64                  | Transparenter Hintergrund. J-Symbol, Light-Verlauf.                                                                     | Desktop-Verknüpfung, Taskbar                                          | Größeres Favicon für Desktop-Shortcuts und Taskbar-Icons.                                                   |
| 10 (`jotti-icon-dark-16.png`)                 | Dark Favicon 16×16                   | Transparenter Hintergrund. J-Symbol, Dark-Verlauf (ca. Emerald-500 → Green-800).                                            | Browser-Tab-Favicon (Dark Mode)                                       | Kleinster Favicon für Browser-Tab im Dark Mode.                                                             |
| 11 (`jotti-icon-dark-32.png`)                 | Dark Favicon 32×32                   | Transparenter Hintergrund. J-Symbol, Dark-Verlauf.                                                                      | Browser-Tab-Favicon (Dark Mode), Lesezeichen                          | Standard-Favicon für Dark-Mode-Browser.                                                                     |
| 12 (`jotti-icon-dark-64.png`)                 | Dark Favicon 64×64                   | Transparenter Hintergrund. J-Symbol, Dark-Verlauf.                                                                      | Desktop-Verknüpfung (Dark Mode), Taskbar                              | Größeres Favicon für Desktop-Shortcuts im Dark Mode.                                                        |

---

## 4. Logo-Anwendungsrichtlinien

### 4.1. Schutzzone (Clear Space)

Um das Logo herum muss immer eine ausreichende Schutzzone eingehalten werden, in der keine anderen Elemente (Text, Grafiken, Ränder) platziert werden dürfen. Die Schutzzone sollte mindestens der Höhe des `J` im Logo entsprechen.

### 4.2. Minimale Größe

- **Full Logo:** Mindestbreite im Web `120px`, im Druck `25mm`.
- **Reines Icon (Asset #6):** Kann bis auf `16px` (Favicon-Größe) herunterskaliert werden.

### 4.3. Unzulässige Verwendung

- Das Logo darf nicht verzerrt, skaliert, rotiert oder verformt werden.
- Die Farben des Logos dürfen nicht verändert werden, außer für die definierten Light/Dark-Varianten.
- Es dürfen keine zusätzlichen Effekte wie Schatten oder Verläufe über das Logo gelegt werden.
- Das Logo darf nicht in einem engen, nicht-zentrierten Kreis platziert werden.
- Das isolierte `J` (Asset #6) sollte nicht als alleinige Wortmarke ohne Text verwendet werden.

---

## 5. Das Theme (Farben)

Das jotti Theme basiert auf dem TailwindCSS-Farbschema. Kanonische Quelle aller Token-Werte ist `frontend/src/index.css` (Light + Dark); die folgenden Tabellen sind ein Auszug. Die Palette: **olive** als neutrale Basis, **emerald** als primäre Markenfarbe (Grün), **zinc** als Sekundärfläche, **red** für destruktive Aktionen.

> **Hinweis:** Die Logo- und Icon-Bild-Assets (Abschnitt 3, 7) wurden im Juli 2026 per OKLCH-Hue-Rotation von der früheren violett/indigo Palette auf die grüne Markenfarbe umgefärbt (Violet-600 → Marken-Hue 165.6). Nur `jotti-logo-examples-and-definitions.png` zeigt noch die alte Palette.

**Border-Radius:** `--radius: 0.45rem` — wird konsequent auf UI-Elemente und die Logo-Icons angewendet.

### 5.1. Light Theme

_olive base, emerald primary, zinc secondary_

| Variable              | Wert                    | Beschreibung                    |
| --------------------- | ----------------------- | ------------------------------- |
| `--background`        | `#ffffff` (white)       |                                 |
| `--foreground`        | `#0c0c09` (olive-950)   |                                 |
| `--card`              | `#ffffff` (white)       |                                 |
| `--primary`           | `#007a55` (emerald-700) | Primärfarbe (Aktionen, Buttons) |
| `--secondary`         | `#f4f4f5` (zinc-100)    | Sekundärfläche (neutral)        |
| `--accent`            | `#f4f4f0` (olive-100)   |                                 |
| `--accent-foreground` | `#1d1d16` (olive-900)   |                                 |

### 5.2. Dark Theme

_olive base, emerald primary, zinc secondary_

| Variable              | Wert                    | Beschreibung                    |
| --------------------- | ----------------------- | ------------------------------- |
| `--background`        | `#0c0c09` (olive-950)   |                                 |
| `--foreground`        | `#fbfbf9` (olive-50)    |                                 |
| `--card`              | `#1d1d16` (olive-900)   |                                 |
| `--primary`           | `#006045` (emerald-800) | Primärfarbe (Aktionen, Buttons) |
| `--secondary`         | `#27272a` (zinc-800)    | Sekundärfläche (neutral)        |
| `--accent`            | `#2b2b22` (olive-800)   |                                 |
| `--accent-foreground` | `#fbfbf9` (olive-50)    |                                 |

---

## 6. Typografie

- **Primary Font (Web & UI):** Inter (oder standard System-Sans) — klar, modern und gut lesbar für UI und Fließtext.
- **Logo Font:** Der „jotti" Schriftzug im Full Logo verwendet eine geometrische Sans-Serif (wie Inter/Montserrat) und ist nicht für Fließtext zu verwenden.

---

## 7. Ikonografie

| Asset | Dateiname                         | Erscheinungsbild                                          | Anwendung                                                     |
| ----- | --------------------------------- | --------------------------------------------------------- | ------------------------------------------------------------- |
| 7–9   | `jotti-icon-light-{16,32,64}.png` | J-Symbol, transparenter Hintergrund, Light-Verlauf        | Browser-Favicons (Light Mode), Lesezeichen, Desktop-Shortcuts |
| 10–12 | `jotti-icon-dark-{16,32,64}.png`  | J-Symbol, transparenter Hintergrund, Dark-Verlauf         | Browser-Favicons (Dark Mode), Lesezeichen, Desktop-Shortcuts  |
| 2     | `jotti-logo-icon-dark.png`        | Quadratisches App-Icon, Slate-950-Hintergrund, abgerundet | PWA-Startbildschirm-Icon (Dark), Apple Touch Icon (Fallback)  |
| 3     | `jotti-logo-icon-light.png`       | Quadratisches App-Icon, Slate-50-Hintergrund, abgerundet  | PWA-Startbildschirm-Icon (Light), Apple Touch Icon (primär)   |

---

## 8. Frontend-Integration (jotti)

**Quelle der Wahrheit:** Dieses Verzeichnis (`assets/`) hält die Originale aller
Marken-Assets. `frontend/public/icons/` und `website/icons/` sind davon
abgeleitete Laufzeitkopien. Bei einer Änderung am Logo oder an den Icons werden
die Originale in `assets/` aktualisiert und die benötigten Größen in die beiden
Kopien übernommen.

### 8.1. Dateistruktur in `frontend/public/`

```
frontend/public/
  icons/
    jotti-icon-light-16.png     # Favicon 16×16 (Light)
    jotti-icon-light-32.png     # Favicon 32×32 (Light) — primärer Fallback
    jotti-icon-light-64.png     # Favicon 64×64 (Light)
    jotti-icon-dark-16.png      # Favicon 16×16 (Dark Mode)
    jotti-icon-dark-32.png      # Favicon 32×32 (Dark Mode)
    jotti-icon-dark-64.png      # Favicon 64×64 (Dark Mode)
    jotti-logo-icon-light.png   # PWA App-Icon / Apple Touch Icon (Light)
    jotti-logo-icon-dark.png    # PWA App-Icon (Dark)
  manifest.webmanifest          # PWA Manifest
```

### 8.2. PWA-Setup

jotti ist Mobile-first und für den Einsatz auf Smartphones (BYOD) konzipiert. Ein PWA-Setup macht Sinn:

- **Motivation:** Servicekräfte können die App beim Anlaufen des Vereinsfestes direkt als Icon auf den Startbildschirm hinzufügen — ohne App Store, ohne Installation.
- **Ergebnis:** `standalone`-Display (kein Browser-Chrome), grüne Theme-Color in der Statusleiste, sofort erkennbares Icon.

Das Manifest liegt unter `frontend/public/manifest.webmanifest`:

```json
{
  "name": "jotti",
  "short_name": "jotti",
  "description": "Das kostenlose Kassensystem für Vereinsfeste.",
  "start_url": "/",
  "display": "standalone",
  "background_color": "#ffffff",
  "theme_color": "#007a55",
  "lang": "de",
  "icons": [
    {
      "src": "/icons/jotti-icon-light-16.png",
      "sizes": "16x16",
      "type": "image/png"
    },
    {
      "src": "/icons/jotti-icon-light-32.png",
      "sizes": "32x32",
      "type": "image/png"
    },
    {
      "src": "/icons/jotti-icon-light-64.png",
      "sizes": "64x64",
      "type": "image/png"
    },
    {
      "src": "/icons/jotti-logo-icon-light.png",
      "sizes": "any",
      "type": "image/png",
      "purpose": "any"
    },
    {
      "src": "/icons/jotti-logo-icon-dark.png",
      "sizes": "any",
      "type": "image/png",
      "purpose": "any"
    }
  ]
}
```

**Hinweis Maskable Icons:** Für ein optimales Android-Erlebnis (adaptiver Icon-Hintergrund) wäre ein `maskable`-Icon mit Safe-Zone-Padding ideal. Da die aktuellen Icons bereits einen farbigen Quadrathintergrund mitbringen (`jotti-logo-icon-*.png`), sind sie im Normalfall ausreichend.

### 8.3. `index.html` — Favicon & Meta-Tags

```html
<!-- Favicons: light mode -->
<link
  rel="icon"
  type="image/png"
  sizes="16x16"
  href="/icons/jotti-icon-light-16.png"
  media="(prefers-color-scheme: light)"
/>
<link
  rel="icon"
  type="image/png"
  sizes="32x32"
  href="/icons/jotti-icon-light-32.png"
  media="(prefers-color-scheme: light)"
/>
<!-- Favicons: dark mode -->
<link
  rel="icon"
  type="image/png"
  sizes="16x16"
  href="/icons/jotti-icon-dark-16.png"
  media="(prefers-color-scheme: dark)"
/>
<link
  rel="icon"
  type="image/png"
  sizes="32x32"
  href="/icons/jotti-icon-dark-32.png"
  media="(prefers-color-scheme: dark)"
/>
<!-- Fallback (light, 32px) -->
<link
  rel="icon"
  type="image/png"
  sizes="32x32"
  href="/icons/jotti-icon-light-32.png"
/>

<!-- Apple Touch Icon für iOS-Startbildschirm -->
<link rel="apple-touch-icon" href="/icons/jotti-logo-icon-light.png" />

<!-- PWA Manifest -->
<link rel="manifest" href="/manifest.webmanifest" />

<!-- Browser-Chrome-Farbe passend zum Theme -->
<meta
  name="theme-color"
  content="#007a55"
  media="(prefers-color-scheme: light)"
/>
<meta
  name="theme-color"
  content="#0c0c09"
  media="(prefers-color-scheme: dark)"
/>
```

### 8.4. Logo-Nutzung im Frontend-Code

| Kontext                           | Empfohlenes Asset                         | Pfad (nach `public/`)              |
| --------------------------------- | ----------------------------------------- | ---------------------------------- |
| Admin-Sidebar Header (Light Mode) | `jotti-logo-full-light.png`               | Assets-Ordner → nicht im `public/` |
| Admin-Sidebar Header (Dark Mode)  | `jotti-logo-full-dark.png`                | Assets-Ordner → nicht im `public/` |
| Login-Seite Branding              | `jotti-logo-full-light.png` / `-dark.png` | je nach Theme                      |
| Browser-Favicon (Light)           | `jotti-icon-light-32.png`                 | `/icons/jotti-icon-light-32.png`   |
| Browser-Favicon (Dark)            | `jotti-icon-dark-32.png`                  | `/icons/jotti-icon-dark-32.png`    |
| PWA Startbildschirm-Icon          | `jotti-logo-icon-light.png`               | `/icons/jotti-logo-icon-light.png` |
| Apple Touch Icon                  | `jotti-logo-icon-light.png`               | `/icons/jotti-logo-icon-light.png` |

**Hinweis:** `jotti-logo-full-*.png` und `jotti-symbol.png` sind bisher nicht in `public/` kopiert. Sie werden erst benötigt, wenn die Logo-Komponente im Frontend implementiert wird (z. B. in `AdminSidebar.tsx` oder `LoginPage.tsx`).

---

## 9. Fazit

Diese Corporate Identity und Design-Dokumentation stellt sicher, dass „jotti" als professionelles, modernes und vertrauenswürdiges Kassensystem wahrgenommen wird, das seine Zielgruppe (Vereine) mit einem einheitlichen visuellen Auftritt unterstützt. Die korrekte Anwendung der generierten Assets ist entscheidend für den Aufbau einer starken Marke.
