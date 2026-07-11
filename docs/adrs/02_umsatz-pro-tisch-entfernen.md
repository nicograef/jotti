# ADR 02: Umsatz pro Tisch entfernen

- **Status:** akzeptiert (2026-07-11)
- **Kontext-Dokumente:**
  UX-Umbau-Plan `plan-admin-dashboard-ux.md` (nach Merge gelöscht,
  siehe Git-Historie), `docs/anforderungen.md`, ADR 01 (Präzedenz)

## Kontext

Das Admin-Reporting kannte einen Breakdown „Umsatz pro Tisch" (R-03):
kassierte Zahlungen gruppiert nach Tisch, angezeigt als eigener Tab im
Live-Dashboard und in der historischen Auswertung. Der UX-Review des
Admin-Dashboards (11.07.2026), gestützt auf das Praxistest-Feedback vom
09.07.2026, hat gezeigt, dass diese Kennzahl keine Entscheidung des
Kassenwarts stützt:

- Der Kassenwart beantwortet drei Fragen: Wie viel ist in der Kasse?
  Steht noch Geld an Tischen aus? Gibt es auffällige Stornierungen? Der
  kassierte Umsatz je Tisch beantwortet keine davon — welcher Tisch am
  meisten umgesetzt hat, ist bei einem Vereinsfest ohne Belang.
- Für die offene Frage „an welchem Tisch steht noch Geld aus?" existiert
  bereits „Offene Tische" (Salden je Tisch). Der abgeschlossene
  Tisch-Umsatz konkurriert mit dieser Ansicht um Aufmerksamkeit auf dem
  375-px-Handy, ohne eigenen Nutzen.
- Kein anderer Teil des Produkts hängt von der Kennzahl ab; sie ist ein
  reines Anzeige-Aggregat.

jotti ist bewusst reduziert: nur notwendige Funktionen, diese aber
hochwertig. Eine Kennzahl, die der reale Einsatz als Ballast entlarvt,
ist ein Entfernungs-Kandidat (Präzedenz: ADR 01, Ausgabe-Bestätigung).

Erwogene Alternativen:

1. **Nur die Tabs ausblenden, Backend behalten** — hinterlässt eine
   Query, DTO-Felder und einen Domain-Typ ohne Konsumenten; die
   Komplexität bleibt in Tests und Doku.
2. **Live entfernen, historisch behalten** — zwei divergierende
   Reporting-Sichten für dieselbe fachlich verworfene Kennzahl.
3. **Vollständige Entfernung über alle Schichten** — sauberster
   Schnitt, keine toten Pfade.

## Entscheidung

„Umsatz pro Tisch" wird **ersatzlos und über alle Schichten entfernt**
(Alternative 3): beide Frontend-Tabs (live + historisch), die
Zod-Schemas, die Response-DTOs, der Domain-Typ `reporting.UmsatzTisch`
samt Breakdown-Feld, die Repository-Methode und die SQL-Query
`GetUmsatzProTisch`. Anforderung R-03 entfällt.

Diese Entfernung berührt **keine persistierten Daten**: „Umsatz pro
Tisch" war ein on-demand aus dem Kassenjournal berechnetes Read Model,
kein eigener Event-Typ und keine Schema-Spalte. Es gibt weder eine
Migration noch eine Event-Contract-Änderung — die Freeze-Disziplin
bleibt unberührt.

„Offene Tische" (Salden je Tisch aus der `tisch_sessions`-Projektion)
bleibt unverändert erhalten und wird im Zuge des UX-Umbaus aufgewertet.

## Konsequenzen

- Das Admin-Dashboard zeigt keinen Tisch-Umsatz mehr; die Reporting-Tabs
  reduzieren sich auf Übersicht, Servicekräfte und Stornierungen.
- Der Reporting-Breakdown trägt nur noch `UmsatzProServicekraft`
  (R-04 bleibt bestehen).
- Sollte ein späterer Praxisbedarf den Tisch-Umsatz doch rechtfertigen,
  ist die Kennzahl aus dem Kassenjournal jederzeit neu ableitbar; eine
  Wiederaufnahme bliebe über ein neues ADR möglich.
