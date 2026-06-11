# PRD: Website Design System & Build/Deploy-Verbesserung

> Kontext: Ergebnis des Code-Audits der Projekt-Website (jotti.rocks) vom
> 11. Juni 2026. Betrifft ausschließlich die statische Website, nicht die
> jotti-App (Frontend/Backend).

## Problem Statement

Die jotti-Website (Landing Page, zwei Leitfaden-Unterseiten, 404-Seite) wird
über eine einzige, ~1200 Zeilen lange CSS-Datei gestylt. Diese Datei ist
organisch gewachsen und schwer zu pflegen:

- Rund 70 Komponentenklassen, viele davon genau einmal benutzt
  (z. B. Hero-Tags, Guide-Eyebrow, CTA-Icon, Compliance-Actions). Jede neue
  Sektion erzeugt neue Einmal-Klassen statt vorhandene Bausteine zu nutzen.
- Layout-Muster (Grid-Spalten, Abstände, Zentrierung, Schriftgrößen)
  wiederholen sich in leicht abweichenden Varianten; die Breakpoints 640px
  und 960px sind ~23-mal als Magic Numbers dupliziert — desktop-first per
  max-width, während die jotti-App mobile-first (Tailwind) arbeitet.
- Es gibt tote Klassen (Showcase-Grid) und ungenutzte Assets
  (Admin-Screenshots, 64px-Icons), die niemand bemerkt, weil nichts sie prüft.
- Wer die jotti-App kennt (Tailwind-Utilities), findet auf der Website ein
  komplett anderes Styling-Modell vor.

Auch Auslieferung und Qualitätssicherung haben Lücken:

- Der Landing-Page-nginx liefert CSS, Fonts und Bilder **ohne gzip und ohne
  Cache-Control-Header** aus — jeder Besuch lädt alles unkomprimiert neu.
- Die Website ist von Prettier, Make-Targets und CI vollständig ausgenommen.
  Kaputte interne Links, fehlende Assets oder nicht auflösbare SSI-Includes
  würden erst in Produktion auffallen.

## Solution

Ein kleines, jotti-eigenes **utility-first CSS Design System** ersetzt die
gewachsene CSS-Datei — handgeschrieben, in plain CSS, ohne Build-Schritt und
ohne Dependencies:

- **Design-Tokens** (Farben, Spacing-Skala, Radius, Typografie) als CSS Custom
  Properties im jotti-Theme. Die Werte werden bewusst manuell mit der
  Frontend-App synchron gehalten (dokumentiert per Kommentar) — keine
  technische Kopplung.
- **Kuratiertes Utility-Set** nach Tailwind-Vorbild (Namen wie `flex`,
  `grid-cols-3`, `gap-4`, `text-muted`, `font-bold`), mobile-first mit zwei
  Breakpoint-Präfixen (`md:` ≈ 640px, `lg:` ≈ 960px). Es entsteht nur, was
  die Seiten tatsächlich benutzen — kein generiertes Vollsortiment.
- **Wenige echte Komponentenklassen** (~8–12) für mehrfach genutzte,
  zustandsbehaftete oder verschachtelte Muster: Buttons, Header/Navigation
  (inkl. Mobile-Menü), Vergleichstabelle, Hinweis-Box, Checklisten, Tags,
  Code-Box, Prosa-Layer für die Leitfaden-Seiten.
- Alle Seiten und Partials werden auf Utilities + Komponenten migriert;
  Einmal-Klassen, tote Styles, ungenutzte Assets und obsolete Meta-Tags
  fliegen raus.

Build und Deploy werden gezielt, minimal verbessert:

- **nginx**: gzip-Kompression und Cache-Control-Header für statische Assets
  (Fonts/Bilder lange + immutable, CSS/JS kurz, HTML nicht gecacht) — in der
  Produktions- und der Dev-Konfiguration.
- **Qualitätssicherung**: Ein Website-Check als Make-Target und CI-Job prüft
  Formatierung (Prettier), Link-/Asset-/SSI-Integrität und
  CSS-Klassen-Konsistenz in beide Richtungen (keine benutzte Klasse fehlt,
  keine definierte Klasse ist ungenutzt). Das Audit, das diese PRD ausgelöst
  hat, wird damit zum dauerhaften, automatisierten Check.

## User Stories

1. As a Entwickler, I want eine dokumentierte Spacing-/Farb-/Typo-Skala als
   CSS-Tokens, so that neue Sektionen ohne neue Magic Numbers entstehen.
2. As a Entwickler, I want Tailwind-ähnlich benannte Utilities, so that ich
   das Mental Model der jotti-App auf der Website wiederverwenden kann.
3. As a Entwickler, I want mobile-first Breakpoint-Präfixe (`md:`, `lg:`),
   so that responsives Verhalten im HTML sichtbar ist statt in verstreuten
   max-width-Queries.
4. As a Entwickler, I want neue Seitenabschnitte ausschließlich aus
   vorhandenen Utilities zusammensetzen können, so that die CSS-Datei beim
   Wachsen der Website nicht mitwächst.
5. As a Entwickler, I want eine kleine, benannte Menge an Komponentenklassen
   (Button, Tabelle, Hinweis-Box, …), so that wiederkehrende Muster nicht als
   lange Utility-Ketten dupliziert werden.
6. As a Entwickler, I want einen Kommentarkopf in der CSS-Datei, der Tokens,
   Skala und Namenskonvention erklärt, so that ich das System ohne externe
   Doku verstehe.
7. As a Entwickler, I want die jotti-Theme-Werte an einer markierten Stelle
   gepflegt, so that ein Farbabgleich mit der App ein bewusster, lokaler
   Eingriff bleibt.
8. As a Website-Besucher, I want unverändertes Erscheinungsbild und Verhalten
   nach der Migration, so that die Umstellung für mich unsichtbar ist.
9. As a Website-Besucher mit Smartphone, I want komprimierte und gecachte
   Assets, so that die Seite auch im Festzelt-Netz schnell lädt.
10. As a wiederkehrender Besucher, I want dass Fonts und Bilder aus dem
    Browser-Cache kommen, so that Folgebesuche praktisch sofort rendern.
11. As a Entwickler, I want `make`-Targets für Format-Check und
    Website-Check, so that ich vor dem Commit lokal dieselben Prüfungen
    fahren kann wie die CI.
12. As a Entwickler, I want einen CI-Job, der kaputte interne Links, fehlende
    Assets und nicht auflösbare SSI-Includes ablehnt, so that solche Fehler
    nie in Produktion gelangen.
13. As a Entwickler, I want einen beidseitigen CSS-Klassen-Check, so that
    tote Utilities und Tippfehler in Klassennamen automatisch auffallen.
14. As a Entwickler, I want Prettier-Formatierung für die Website-Dateien,
    so that HTML/CSS-Diffs klein und einheitlich bleiben.
15. As a Betreiber der jotti.rocks-Instanz, I want dass Deploy weiterhin
    `git pull` + ein Make-Target ist, so that keine neue Infrastruktur
    (Secrets, Pipelines) entsteht.
16. As a Contributor, I want dass die Website ohne Node/Build-Toolchain
    bearbeitbar bleibt (Editor + `make website`), so that die Einstiegshürde
    minimal bleibt.
17. As a Suchmaschine, I want unveränderte URLs, Canonicals, Sitemap und
    strukturierte Meta-Daten, so that das bestehende Ranking erhalten bleibt.
18. As a Entwickler, I want konsistent absolute Pfade für Assets und interne
    Links, so that Markup zwischen Seiten verschoben werden kann, ohne dass
    Links brechen.
19. As a Entwickler, I want dass ungenutzte Assets und obsolete Meta-Tags
    entfernt sind, so that das Deployment nur enthält, was die Website
    wirklich braucht.
20. As a Entwickler, I want dass die lokale Vorschau (`make website`) die
    Produktion inkl. gzip/Cache-Verhalten spiegelt, so that
    Auslieferungsfehler lokal reproduzierbar sind.

## Implementation Decisions

**Design System**

- Plain CSS, handgeschrieben, null Dependencies, kein Build-Schritt. Eine
  Stylesheet-Auslieferung wie bisher (eine Datei, ein Request).
- Interne Struktur über native CSS Cascade Layers (`@layer tokens, base,
  components, utilities`), damit Utilities Komponenten ohne `!important`
  überschreiben — gleiche Layer-Semantik wie Tailwind 4 in der App.
- Tokens als CSS Custom Properties: jotti-Farbpalette (identische
  oklch-Werte wie die App, manuell synchron gehalten und als solche
  kommentiert), Spacing-Skala in rem (Tailwind-Viertel-Schritte), Radius,
  Schriftfamilie/-größen, Container-Breite.
- Utility-Namenskonvention nach Tailwind (z. B. `mx-auto`, `grid-cols-2`,
  `gap-6`, `text-center`, `text-muted`, `rounded`, `max-w-prose`).
  Kuratiert: nur tatsächlich verwendete Utilities existieren; der
  CSS-Klassen-Check erzwingt das dauerhaft.
- Responsive: mobile-first, zwei min-width-Breakpoints mit Präfix-Utilities
  (`md:` ≈ 640px, `lg:` ≈ 960px, escaped Klassennamen). Bestehende
  desktop-first max-width-Queries werden bei der Migration invertiert.
- Hybrid-Schnitt: Layout, Abstände, Typografie und Farben im HTML über
  Utilities; Komponentenklassen nur für mehrfach genutzte oder
  interaktive/verschachtelte Muster (Buttons mit Varianten, Site-Header und
  Mobile-Navigation inkl. `nav-open`-Zustand, Vergleichstabelle,
  Hinweis-Box, Check-/To-do-Listen mit Pseudo-Elementen, Tag-Pille,
  Code-/Quickstart-Box, Guide-Prosa-Layer für Fließtext-Typografie).
- Die Button-Größenvarianten werden dedupliziert (Outline-Variante erbt die
  Small-Größe nicht mehr implizit).

**Migration**

- Alle vier Seiten und alle Partials werden in einem Zug migriert; das
  Erscheinungsbild bleibt pixel-nah erhalten (kein Redesign).
- Aufräumarbeiten aus dem Audit fließen ein: tote Showcase-Styles entfernen,
  ungenutzte Bilder/Icons löschen, `meta keywords` und `X-UA-Compatible`
  entfernen, relative Pfade auf der Startseite durch absolute ersetzen,
  Sitemap-`lastmod` aktualisieren.

**Auslieferung (nginx)**

- gzip für textbasierte Typen (HTML, CSS, JS, SVG, XML) aktivieren.
- Cache-Control: Fonts und Bilder lang + `immutable` (Dateinamen ändern sich
  praktisch nie; bei Austausch wird der Dateiname versioniert), CSS/JS mit
  kurzer max-age, HTML ohne Caching (SSI-gerendert).
- Identische Regeln in der Produktions-Serverdefinition (Landing-Block) und
  der lokalen Dev-Konfiguration, damit `make website` die Produktion
  spiegelt. Die bewusste Duplikation der beiden Configs bleibt bestehen
  (dokumentiert); es entsteht keine dritte Kopie.

**Qualitätssicherung**

- Prettier (bereits im Frontend vorhanden, keine neue Dependency) formatiert
  die Website-Dateien; Make-Targets für Schreiben und Check.
- Ein dependency-freies Check-Script prüft statisch: interne Links zeigen
  auf existierende Seiten/Anker, referenzierte Assets (Bilder, Fonts, CSS,
  JS) existieren, SSI-Includes lösen auf, und CSS-Klassen sind beidseitig
  konsistent (benutzt ⇆ definiert).
- Beides läuft als Make-Target lokal und als CI-Job bei Änderungen unter dem
  Website-Pfad.
- Der Docker-basierte Rendering-Smoke-Test (nginx starten, Seiten curlen)
  bleibt ein manueller, lokaler Schritt über `make website` — kein CI-Job.

**Deploy**

- Der Deploy-Mechanismus bleibt unverändert: `git pull` auf dem Server +
  bestehendes Make-Target. Kein Auto-Deploy, keine neuen Secrets.

## Testing Decisions

- Ein guter Test prüft von außen sichtbares Verhalten des Artefakts — hier:
  die ausgelieferten Dateien als Ganzes (Links erreichbar, Assets vorhanden,
  Includes auflösbar, Klassen konsistent) — nicht die interne Struktur des
  CSS oder einzelne Selektoren.
- Getestet wird über das Check-Script (statische Integritätsprüfung) und den
  Prettier-Check; beide laufen in CI. Sie ersetzen genau die manuellen
  Prüfschritte des Audits.
- Der visuelle Abgleich nach der Migration erfolgt manuell über die lokale
  Vorschau (`make website`) auf den drei Referenz-Breakpoints (Mobil,
  Tablet, Desktop) — automatisierte visuelle Regressionstests sind den
  Aufwand für vier statische Seiten nicht wert.
- Prior Art: die bestehenden Check-/Lint-Targets im Makefile und die
  CI-Jobs für Backend/Frontend; das Integrations-Testscript im Repo-Root
  als Muster für ein eigenständiges Prüfscript.

## Out of Scope

- **Kein Redesign**: Inhalte, Layout, Farben und Typografie der Website
  bleiben visuell unverändert; es ändert sich nur, *wie* sie implementiert
  sind.
- **Kein Tailwind-Tooling** (npm, Standalone-CLI, PostCSS) und kein
  selbstgebauter Utility-Generator.
- **Keine geteilte Token-Datei** zwischen App und Website; Werte werden
  manuell synchron gehalten.
- **Keine Styleguide-/Preview-Seite**; die Doku lebt als Kommentarkopf in
  der CSS-Datei.
- **Kein Auto-Deploy** (GitHub Actions → VPS) und keine Änderungen an
  Zertifikats-/Domain-Setup oder den Demo-App-Serverblöcken.
- **Keine neuen Inhalte oder Seiten** (z. B. Blog, Pressebereich).
- **Kein JavaScript-Umbau**; das Mobile-Menü-Script bleibt wie es ist.

## Further Notes

- Die Website bleibt bewusst SSI-basiert (nginx rendert Partials zur
  Laufzeit) — das ist der existierende „Build-Schritt“ und hat sich im Audit
  als korrekt funktionierend erwiesen (alle Seiten 200, Partials gesperrt,
  404-Seite greift).
- Der CSS-Klassen-Konsistenz-Check ist das Herzstück gegen schleichenden
  Verfall: Er macht das kuratierte Utility-Set durchsetzbar (jede ungenutzte
  Utility bricht den Check) und ersetzt die Disziplin, die ein
  Tailwind-Purge-Schritt sonst automatisch erzwingen würde.
- Erwartete Nebeneffekte: kleinere CSS-Datei, ein konsistentes
  Styling-Modell über App und Website hinweg, und ein Deployment, das nur
  noch tatsächlich genutzte Assets enthält.
