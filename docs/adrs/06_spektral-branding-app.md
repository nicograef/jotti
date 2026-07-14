# ADR 06: Spektral-Branding im App-Frontend (Ablösung der Frontend-Ausnahme aus ADR 05)

- **Status:** akzeptiert (2026-07-14)
- **Kontext-Dokumente:**
  PRD `docs/prds/prd-spektral-redesign-frontend.md` (mit dem verbindlichen
  Motion-Inventar), Plan `docs/plans/plan-spektral-redesign-frontend.md`,
  Design-Handoff `docs/prds/design_handoff_spektral_redesign/`
- **Löst ab:** die Frontend-Ausnahme aus
  [ADR 05](05_spektral-branding-website.md) (siehe dort den Ablöse-Vermerk)

## Kontext

[ADR 05](05_spektral-branding-website.md) hat den vollen Spektral-Einsatz für
die Marketing-Website abgenommen, das Produkt-Frontend aber ausdrücklich
ausgenommen: „Das Produkt-Frontend (`frontend/`) behält sein bestehendes,
zurückhaltendes Design; die Spektral-Erweiterung ist keine Aussage über die
App." Damit blieben Website und App visuell entkoppelt.

Der Design-Handoff und die PRD `docs/prds/prd-spektral-redesign-frontend.md`
für das App-Redesign kehren diese Ausnahme um — allerdings nicht als Kopie des
Website-Auftritts, sondern als bewusst dezente Übertragung der Markensprache
auf ein Bedien-Werkzeug, das ehrenamtliche Helfer unter Zeitdruck bedienen.
Die PRD hat diesen Umfang mit dem Betreiber als verbindlich abgenommen
(inklusive des tabellierten Motion-Inventars).

Das ist eine Entscheidung mit langfristiger Tragweite (sie prägt das gesamte
App-Design und bindet alle Redesign-Phasen) und zugleich eine bewusste Abkehr
von einer zuvor dokumentierten Festlegung — genau der Fall, für den
`docs/adrs/` gedacht ist.

Erwogene Alternativen:

1. **Frontend-Ausnahme aus ADR 05 beibehalten** — die App bliebe optisch
   markenlos und zur neu gestalteten Website inkonsistent; die bereits
   abgenommene PRD wäre nicht umsetzbar.
2. **Den vollen Website-Spektral-Einsatz 1:1 in die App übernehmen** —
   animierte Verläufe, Spektral-Flächen und -Streifen als tragende UI-Elemente
   wären für ein Kassen-Bedienwerkzeug zu laut, lenkten von Beträgen und
   Status ab und widersprächen dem Produkt-Konservatismus (AGENTS.md).
3. **Dezente Übertragung mit einer harten Leitplanke** (gewählt): das Spektrum
   erscheint an genau vier klar umrissenen Stellen, alles Übrige (Layout,
   Bedienablauf, Status-Semantik, Barrierefreiheit) bleibt unangetastet.

## Entscheidung

Die Spektral-Markensprache wird dezent in das App-Frontend (`frontend/`)
übertragen (Alternative 3); die Frontend-Ausnahme aus ADR 05 ist damit
abgelöst. Das Redesign ist ein reiner Restyling-Layer.

**Vier-Stellen-Regel (neue Leitplanke).** Der Spektral-Verlauf (`--spectral`)
und die `--sp-*`-Töne erscheinen in der App ausschließlich an vier Stellen:

1. **Wortmarke** — „jotti" als Spektral-Text-Füllung (Space Grotesk 700),
2. **Hintergrund-Glows** — weiche, geblurte Farbflächen am Login und hinter den
   Admin-Seitenköpfen,
3. **Ladezustände** — der Spektral-getönte Skeleton-Shimmer,
4. **Hairline-Akzente** — die 2-px-Kante der Hero-Kennzahlkarte und der 3-px-
   Marker am aktiven Sidebar-Eintrag.

Alles darüber hinaus ist ausgeschlossen. Insbesondere bleiben unangetastet:

- die **Status-Semantik** (Grün `--primary` / Rot `--destructive` / Amber): die
  Ampelfarben sind reserviert und werden nie durch Spektral-Töne ersetzt oder
  eingefärbt,
- **Layouts, Bedienabläufe und Barrierefreiheit** (ARIA-Rollen, `aria-label`,
  Fokusreihenfolge): das Redesign ändert nur die Optik, nie das Verhalten.

**Bewegung** folgt durchgehend dem verbindlichen Motion-Inventar der PRD
(Dauern und Easings dort tabelliert). Eine zentrale
`@media (prefers-reduced-motion: reduce)`-Regel neutralisiert alle Animationen
und Transitions global — inklusive der Loop-Animationen (Shimmer, Live-Puls,
Glow-Drift). Kein Zustand darf sich allein aus Bewegung erschließen.

## Konsequenzen

- Die **Vier-Stellen-Regel** bindet alle künftigen Design-Änderungen am
  Frontend: eine neue Spektral-Fläche außerhalb dieser vier Stellen — oder eine
  Einfärbung der Status-Ampel — braucht ein neues ADR.
- **ADR 05 bleibt akzeptiert** und für die Website vollständig gültig; nur seine
  Frontend-Ausnahme ist abgelöst. Der dortige Ablöse-Vermerk verweist hierher.
- Kontrast- und Barrierefreiheits-Prüfungen berücksichtigen die Glows und
  Hairlines als Hintergrund für Text; Korrekturen erfolgen token-/deckkraft-
  seitig innerhalb der Handoff-Bandbreite, nicht als Einzelfall-Hacks.
- Die verbindliche Motion-Tabelle der PRD ist die einzige Quelle für Dauern und
  Kurven; Einsatzstellen konsumieren nur die zentral definierten Utilities und
  legen keine Ad-hoc-Animationen an.
- Eine Rückkehr zum markenlosen Frontend bräuchte ein neues ADR.
