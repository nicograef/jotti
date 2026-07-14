# ADR 05: Spektral-Branding auf der Website (Ablösung von „Spektrum nur dekorativ")

- **Status:** akzeptiert (2026-07-13)
- **Kontext-Dokumente:**
  PRD `docs/prds/prd-website-redesign.md` (fasst den Design-Handoff
  zusammen), Plan `docs/plans/plan-website-redesign.md`

## Kontext

Die bisherige Marketing-Website setzte die Spektralfarben der Marke bewusst
nur zurückhaltend ein: als feine Haarlinie an Header- und Footer-Kante und als
leiser Hintergrund-Wash im Hero. Diese Festlegung — „Spektrum nur dekorativ,
keine UI-Fläche" — hielt den Verlauf aus Text, Flächen und interaktiven
Elementen heraus; der einzige tragende Akzent war das Markengrün.

Der Design-Handoff für das Website-Redesign
(im PRD `docs/prds/prd-website-redesign.md` zusammengefasst) macht den Spektral-Verlauf dagegen
zum zentralen Markenelement der Website: ein animierter Text-Verlauf im
H1-Akzentwort, sechs Spektral-Akzentfarben für Feature-Explorer,
Ablauf-Schritte und Compliance-Karten, weiche Spektral-Blobs im Hintergrund,
Spektral-Streifen auf Karten und ein Scroll-Reveal beim Hereinscrollen. Die
PRD hat diesen vollen Spektral-Einsatz mit dem Betreiber bereits als
verbindlich abgenommen (`docs/prds/prd-website-redesign.md`, Abschnitt zum
Spektral-Einsatz).

Damit steht der Handoff in direktem Widerspruch zur alten Festlegung. Eine
Entscheidung mit langfristiger Tragweite (sie prägt das visuelle Auftreten der
gesamten Website und bindet alle Redesign-Phasen), die zugleich eine bewusste
Abkehr von einer zuvor dokumentierten Festlegung ist — genau der Fall, für den
`docs/adrs/` gedacht ist.

Erwogene Alternativen:

1. **„Spektrum nur dekorativ" beibehalten** — der Handoff und die bereits
   abgenommene PRD wären nicht umsetzbar; das Redesign verlöre sein zentrales
   Unterscheidungsmerkmal gegenüber der austauschbaren Vorgängerseite.
2. **Spektral-Einsatz nur teilweise übernehmen** (z. B. Akzentfarben, aber kein
   animierter H1-Verlauf) — uneinheitliches Ergebnis, das weder dem Handoff
   noch der alten Festlegung folgt und die Konsistenz zwischen Sektionen
   aufweicht.
3. **Spektral-Einsatz voll übernehmen** wie im Handoff, die alte Festlegung
   ausdrücklich ablösen und in diesem ADR festhalten.

## Entscheidung

Der volle Spektral-Einsatz des Handoffs wird übernommen (Alternative 3) und
löst die Festlegung „Spektrum nur dekorativ" ausdrücklich ab. Der
Spektral-Verlauf ist fortan ein tragendes Marken- und UI-Element der Website:

- animierter Spektral-Textverlauf im H1-Akzentwort,
- sechs Spektral-Akzentfarben (`--sp-*`) für Feature-Explorer, Ablauf-Schritte
  und Compliance-Karten,
- weiche Spektral-Blobs als Hintergrund-Grundelement,
- Scroll-Reveal als Progressive Enhancement.

Diese Ablösung gilt ausschließlich für die Website (`website/`). Das
Produkt-Frontend (`frontend/`) behält sein bestehendes, zurückhaltendes Design;
die Spektral-Erweiterung ist keine Aussage über die App.

Bewegung ist durchgehend Progressive Enhancement: `prefers-reduced-motion`
neutralisiert H1-Verlaufs-Animation, Blobs und Reveal global; das
Scroll-Reveal setzt `animation-timeline: view()` voraus und zeigt ohne
Unterstützung schlicht die statischen, vollständig sichtbaren Inhalte.

Sprachliche Nebenbedingung (aus Handoff und PRD): Das Wort „Regenbogen" darf
weder im Quelltext noch in der ausgelieferten Copy der Website vorkommen. Die
spektralen Farben sind das Markenelement und bleiben — nur ihre textliche
Benennung als „Regenbogen" ist untersagt.

## Konsequenzen

- Alle Redesign-Phasen setzen den Spektral-Verlauf als tragendes Element ein;
  die früheren, rein dekorativen Muster (`--brand-spectrum-wash`,
  `.spectrum-edge-*`) entfallen mit den sie konsumierenden Alt-Sektionen im
  Zuge der Übergangsregel des Plans.
- Kontrast- und Barrierefreiheits-Prüfungen berücksichtigen die Spektral-Töne
  auf Flächen und als Text (namentlich der Eyebrow in `--sp-teal`); nötige
  Korrekturen erfolgen token-seitig, nicht als Einzelfall-Hacks.
- Ein „Regenbogen"-Scan gehört zur Gesamtabnahme (Plan Phase 10).
- Eine Rückkehr zu „Spektrum nur dekorativ" bräuchte ein neues ADR.
