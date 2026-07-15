# PRD: Produkt- und Varianten-Statistik im Report

## Problem Statement

Der Report zeigt heute Gesamt-Kennzahlen einer Kassensitzung: kassierter Umsatz,
Umsatz pro Servicekraft, Umsatz pro Steuersatz, Stornierungen. Was jedoch fehlt,
ist die einfachste operative Frage eines Kassenwarts nach dem Fest:

> „Wieviele Pommes mit Ketchup haben wir rausgegeben? Wieviele Tagesessen?
> Wieviel Umsatz haben wir mit welcher Bratwurst gemacht?"

Ohne diese Aufschlüsselung lässt sich weder der Einkauf fürs nächste Fest planen
(„von welcher Ware brauchen wir wieviel") noch erkennen, welches Produkt sich
finanziell lohnt. Die Daten liegen vollständig im Kassenjournal — jede Bestell-,
Zahlungs- und Storno-Position trägt Produktname, Variantenname, Einzelpreis und
Menge — sie werden nur nirgends pro Produkt/Variante zusammengefasst.

Das entspricht der offenen Roadmap-Anforderung **R-05 (Produktumsatz-Reporting,
Prio: Nice)**.

## Solution

Der Report bekommt einen neuen Abschnitt **„Verkäufe pro Produkt"**, sowohl in
der Tagesabrechnung einer (i. d. R. abgeschlossenen) Kassensitzung als auch im
Live-Dashboard der offenen Kassensitzung. Beide zeigen die Statistik jeweils
**pro Kassensitzung** — passend zum restlichen Report, ohne neuen Filter.

Die Liste ist **nach Kategorie in getrennte Abschnitte gegliedert** (Essen,
Getränke, Sonstiges — in dieser Reihenfolge) und innerhalb jedes Abschnitts
**pro Variante aufgeschlüsselt und auf Produktebene gruppiert**: jedes Produkt
bildet eine Gruppe mit einer Zwischensumme, darunter stehen seine Varianten.
**Produkte mit nur einer Variante werden zu einer einzigen Zeile
zusammengefasst** (z. B. „Tagesessen" statt „Tagesessen → Tagesessen"; „Cola
0,5 l" statt Produkt- plus identischer Variantenzeile). Jede Zeile (Variante wie
Produkt-Zwischensumme) trägt zwei bewusst getrennte, klar benannte Zahlen:

- **Ausgegebene Menge** — wieviele Portionen zubereitet bzw. rausgegeben wurden.
  Basis: aufgenommene Bestellungen minus geldneutrale Korrekturen, inklusive
  Direktverkäufe. Beantwortet „wieviele Pommes/Tagesessen gingen über die Theke".
- **Umsatz** — die tatsächlich erzielten Einnahmen. Basis: kassierte Zahlungen
  und Direktverkäufe minus Warenrücknahmen und Direktverkauf-Stornos.
  Beantwortet „wieviel Geld haben wir mit diesem Produkt eingenommen".

Die beiden Zahlen ruhen bewusst auf unterschiedlichen Grundlagen (Produktion vs.
Einnahmen) und sind nicht ineinander umrechenbar — eine zurückgenommene Portion
zählt als ausgegeben, mindert aber den Umsatz. Ein kurzer erklärender Hinweis im
Abschnitt macht das transparent, damit niemand aus Umsatz ÷ Menge einen
Stückpreis ableitet.

Sortierung: innerhalb jeder Kategorie werden die Produkte **nach ausgegebener
Menge absteigend** gelistet (Produktname als stabiler Tiebreaker); Varianten
innerhalb eines Produkts ebenso nach ausgegebener Menge absteigend. Eine
Kassensitzung ohne Verkäufe zeigt einen leeren Zustand; ein Kategorie-Abschnitt
ohne Verkäufe entfällt.

## User Stories

1. Als **Admin/Kassenwart** möchte ich pro Produkt und Variante die ausgegebene
   Menge sehen, damit ich Einkauf und Produktion für das nächste Fest planen kann.
2. Als **Admin/Kassenwart** möchte ich pro Produkt und Variante den erzielten
   Umsatz sehen, damit ich erkenne, welche Produkte sich finanziell lohnen.
3. Als **Admin/Kassenwart** möchte ich die Varianten unter ihrem Produkt gruppiert
   und mit Produkt-Zwischensumme sehen, damit ich schnell das große Bild und
   zugleich die Detailtiefe habe.
4. Als **Admin/Kassenwart** möchte ich die Produkte nach Kategorie (Essen,
   Getränke, Sonstiges) getrennt aufgelistet sehen, damit ich verwandte Ware
   zusammen im Blick habe und Essen und Getränke nicht vermischt werden.
5. Als **Admin/Kassenwart** möchte ich die Statistik bereits während des
   laufenden Fests im Live-Dashboard sehen, damit ich kurzfristig nachsteuern kann
   (z. B. Ware nachordern, wenn ein Produkt schnell weggeht).
6. Als **Admin/Kassenwart** möchte ich die Statistik je Kassensitzung erhalten,
   passend zum übrigen Report, damit sich Zahlen einer Sitzung eindeutig zuordnen
   lassen.

## Implementation Decisions

### Datenquelle und Architektur

- **Reine additive Leseauswertung** über das `kassenjournal`, gefiltert nach
  `kassensitzung_nr` — kein neues Event, kein Schema-Bestand-Eingriff, keine
  Umdeutung bestehender Daten. Das Feature liest ausschließlich eingefrorene
  Event-Positionen und folgt exakt dem bestehenden Muster der
  Umsatz-pro-Steuersatz-Auswertung (`kj_extract_umsatz_pro_steuersatz` /
  `GetUmsatzPositionszeilen`).
- **Keine Cross-Context-Kopplung.** Das Reporting liest weiterhin nur das
  Kassenjournal, nicht die Stammdaten (ACL). Es gibt daher keinen Join auf die
  aktuellen Produkt-/Varianten-Stammdaten.

### Gruppierung und Sortierung

- **Variantenschlüssel:** `varianteId` (stabil, global eindeutige Identity-PK).
  Anzeigename ist der **eingefrorene** `varianteName` aus dem Event.
- **Produktschlüssel:** `produktName` (eingefroren im Event). Die Position trägt
  keine `produktId`, nur `varianteId` plus die Fat-Event-Namen; innerhalb einer
  Kassensitzung ist der Produktname stabil und eindeutig (aktiver Name ist unique).
  Eine Produkt-ID-basierte Gruppierung ist bewusst nicht vorgesehen (siehe
  Out of Scope).
- **Kategorie-Abschnitte:** die Position trägt ein eingefrorenes `kategorie`-Feld
  (`essen`/`getraenk`/`sonstiges`). Produkte werden nach Kategorie in getrennte
  Abschnitte gegliedert, feste Reihenfolge Essen → Getränke → Sonstiges. (Ein
  Produkt liegt je Kassensitzung in genau einer Kategorie; ein theoretischer
  Kategoriewechsel eines Produkts innerhalb einer Sitzung würde als getrennte
  Gruppen erscheinen — vernachlässigbarer Randfall.)
- **Sortierung:** innerhalb jedes Kategorie-Abschnitts Produkte nach
  `AusgegebeneMenge` **absteigend** (`ProduktName` als Tiebreaker); Varianten je
  Produkt ebenso nach `AusgegebeneMenge` absteigend.
- **Ein-Varianten-Produkte** werden in der Anzeige zu einer einzigen Zeile
  zusammengefasst; Beschriftung dann `produktName` + `varianteName` (analog
  `Position.Bezeichnung()`). Das Datenmodell bleibt unverändert (Produkt mit einer
  Varianten-Zeile); nur die Darstellung fasst zusammen.

### Die zwei Kennzahlen (Event-Mapping)

Die Aufteilung spiegelt exakt die bestehende Domänen-Unterscheidung
geldneutral vs. kassenwirksam:

**Ausgegebene Menge** = Summe der Positions-`menge` über:

| Event                        | Vorzeichen | Begründung                                        |
| ---------------------------- | ---------- | ------------------------------------------------- |
| `bestellung-aufgenommen:v1`  | `+`        | bestellt → wird zubereitet/rausgegeben            |
| `bestellung-korrigiert:v1`   | `−`        | geldneutrale Korrektur unbezahlter Positionen: doch nicht rausgegeben |
| `direktverkauf-getaetigt:v1` | `+`        | Direktverkauf gibt die Ware sofort aus            |

**Umsatz** = Summe von `einzelpreisCents × menge` über:

| Event                         | Vorzeichen | Begründung                                    |
| ----------------------------- | ---------- | --------------------------------------------- |
| `zahlung-kassiert:v1`         | `+`        | kassierte Einnahme                            |
| `direktverkauf-getaetigt:v1`  | `+`        | kassierte Einnahme                            |
| `stornierung-erteilt:v1`      | `−`        | kassenwirksame Warenrücknahme (negativer Umsatz) |
| `direktverkauf-storniert:v1`  | `−`        | kassenwirksame Rückgabe (negativer Umsatz)    |

- **`bestellung-umgebucht:v1` wird in keiner Zahl gezählt** — die Positionen
  wurden bereits bei der Bestellaufnahme erfasst; das Umbuchen zwischen Tischen
  ändert produkt-/varianten-seitig nichts (sonst Doppelzählung).
- **Warenrücknahme/Direktverkauf-Storno mindern nur den Umsatz, nicht die Menge:**
  die Ware wurde ausgegeben und danach zurückgenommen — die Produktionszahl bleibt.
- **`bestellung-korrigiert:v1` mindert nur die Menge, nicht den Umsatz:**
  geldneutral, betrifft nur noch nicht rausgegebene/bezahlte Ware.

### Backend-Module

- **Domäne (`domain/reporting`):** zwei neue Read-Model-Typen ohne `json`-Tags,
  analog zu `UmsatzServicekraft`/`StornierungPosition`:
  - `VarianteStatistik` — `VarianteID`, `VarianteName`, `AusgegebeneMenge`,
    `UmsatzCents`.
  - `ProduktStatistik` — `Kategorie`, `ProduktName`, `AusgegebeneMenge`
    (Zwischensumme), `UmsatzCents` (Zwischensumme), `Varianten []VarianteStatistik`.
  - Neues Feld `ProduktStatistik []ProduktStatistik` auf `ReportingData` **und**
    `LiveReportingData`.
- **Query (`sqlc/queries/reporting.sql`):** eine neue Query
  `GetProduktStatistik(@kassensitzung_nr)`, die `data->'positionen'` per
  `jsonb_array_elements` entfaltet und **flache Zeilen pro Variante**
  `(kategorie, varianteId, produktName, varianteName, ausgegebeneMenge, umsatzCents)`
  mit den obigen Vorzeichen liefert
  (`GROUP BY kategorie, varianteId, produktName, varianteName`). Anschließend
  `make sqlc`. Die Aggregation der beiden Vorzeichen-Regeln bleibt in der Query;
  die Kategorie-/Produkt-Gruppierung, Zwischensummen und Sortierung erledigt die
  Anwendungsschicht (wie `computeUmsatzProSteuersatz`).
- **Repository (`repository/reporting_repo`):** die neue Query in die bestehenden
  `errgroup`-Blöcke von `GetReporting` und `GetLiveReporting` einreihen; Zeilen →
  flache Varianten-Zeilen mappen.
- **Anwendungsschicht (`api/reporting/application`):** eine reine Funktion baut
  aus den flachen Varianten-Zeilen die Produkt-Hierarchie, berechnet die
  Zwischensummen und sortiert: Kategorien in fester Reihenfolge (Essen → Getränke
  → Sonstiges), Produkte je Kategorie nach `AusgegebeneMenge` absteigend,
  Varianten je Produkt nach `AusgegebeneMenge` absteigend (`ProduktName` bzw.
  `VarianteName` als stabiler Tiebreaker). Die Liste ist damit bereits fertig
  sortiert; das Frontend rendert nur. Deep Module, isoliert testbar über flache
  Eingabe → gruppierte, sortierte Ausgabe.
- **HTTP (`api/reporting/http`):** neue `json`-getaggte Response-DTOs
  (`produktStatistik` mit verschachtelten `varianten`) plus `to*`-Mapper. Keine
  neuen Endpunkte — die bestehenden `admin/get-abrechnung` und
  `admin/get-live-reporting` tragen das neue Feld.

### Frontend-Module

- **Zod-Schemas (`admin/reporting/types.ts`):** Spiegel der neuen DTOs
  (`ProduktStatistikSchema` mit `kategorie` und `varianten`), eingehängt in
  `ReportingDataSchema` und `LiveReportingDataSchema`.
- **Backend-Klasse:** unverändert im Aufruf; validiert das erweiterte Schema.
- **Anzeige:** neuer Abschnitt „Verkäufe pro Produkt" in `ReportingResults.tsx`
  (Tagesabrechnung, Teil des druckoptimierten Z-Bons) und in
  `LiveReportingSection.tsx` (Live-Dashboard). Gegliedert in Kategorie-Abschnitte
  (Essen, Getränke, Sonstiges) mit Abschnittsüberschrift; je Produkt eine
  Produktzeile mit Zwischensumme und darunter die Variantenzeilen, **außer bei
  einer einzigen Variante — dann eine zusammengefasste Zeile**. Zwei Spalten
  **„Ausgegeben"** (Menge) und **„Umsatz"** (€, über `formatEuro`), plus der
  erklärende Ein-Satz-Hinweis zu den beiden Grundlagen. Leerer Zustand bei einer
  Sitzung ohne Verkäufe; leere Kategorie-Abschnitte werden weggelassen. Die
  Reihenfolge kommt fertig vom Backend — das Frontend sortiert nicht um.

### Dokumentation

- Nach Umsetzung **R-05 aus der Roadmap** in `docs/anforderungen.md` in den
  Reporting-Funktionsumfang verschieben; die neuen Read-Model-Typen in der
  Read-Model-Übersicht von `docs/handbuch.md` und `docs/language.md` ergänzen.

## Testing Decisions

Getestet wird externes Verhalten (welche Zahlen kommen bei welchen Events heraus),
nicht die interne Zerlegung.

- **Aggregations-/Vorzeichen-Verhalten (Repository, Integrationstest gegen echtes
  Postgres):** Prior Art `repository/reporting_repo/repo_test.go`. Ein Szenario mit
  Bestellungen, Korrekturen, Zahlungen, Warenrücknahmen, Direktverkäufen und
  Direktverkauf-Stornos über mehrere Produkte/Varianten prüft: korrekte
  ausgegebene Menge (Bestellung − Korrektur + Direktverkauf), korrekter Umsatz
  (Kassiert + Direktverkauf − Storno), Umbuchungen zählen nicht, und die Trennung
  (Warenrücknahme mindert nur Umsatz, Korrektur nur Menge).
- **Gruppierung und Sortierung (Anwendungsschicht, Unit-Test):** Prior Art
  `api/reporting/application/query_test.go` (analog `computeUmsatzProSteuersatz`).
  Flache Varianten-Zeilen → korrekte Kategorie-Abschnitte in fester Reihenfolge
  (Essen → Getränke → Sonstiges), korrekte Produkt-Gruppen, korrekte
  Zwischensummen, Sortierung nach Menge absteigend, stabile Tiebreaker.
- **Konsistenz mit den Gesamtzahlen:** Prior Art
  `api/reporting/application/query_export_konsistenz_test.go`. Die Summe aller
  Produkt-`UmsatzCents` muss mit dem bestehenden kassierten Gesamtumsatz (bzw. der
  Summe der `UmsatzProSteuersatz`-Brutto­werte) übereinstimmen — dieselbe
  Positions-Basis.
- **Frontend-Darstellung:** Prior Art `ReportingResults.test.tsx` /
  `LiveReportingSection.test.tsx`. Aus einer Fixture werden Kategorie-Abschnitte,
  Produkte, Varianten, Zwischensummen, die Ein-Zeilen-Zusammenfassung bei
  Ein-Varianten-Produkten und der leere Zustand gerendert; das Schema wird
  validiert.

## Out of Scope

- **Auswertung über mehrere Kassensitzungen** (Saison-/Jahressumme). Ausschließlich
  pro Kassensitzung, wie der übrige Report.
- **Produkt-Rollup über die Stammdaten (`produktId`).** Gruppiert wird über den
  eingefrorenen `produktName`; ein Join auf die aktuellen Stammdaten würde die
  Reporting-ACL verletzen.
- **Kategorie-Zwischensummen.** Die Kategorien Essen/Getränke/Sonstiges dienen nur
  der Gliederung in Abschnitte (siehe Solution), nicht als eigene aggregierte
  Summenzeile pro Kategorie.
- **Zeitliche Aufschlüsselung** (Stunden-/Tageszeit-Verlauf), Trends, Diagramme.
- **Separate „kassierte Menge"-Spalte.** Es gibt genau die zwei vereinbarten
  Zahlen (ausgegebene Menge, Umsatz).
- **Umsatz pro Produkt je Servicekraft** oder andere Kreuzauswertungen.
- **Export-Änderungen.** Der DSFinV-K-Export bleibt unberührt; dies ist eine reine
  Bildschirm-/Report-Auswertung.
- **Filtern/Suchen/CSV-Download** innerhalb der Produktliste.

## Further Notes

- Setzt die Roadmap-Anforderung **R-05 (Produktumsatz-Reporting, Nice)** um.
- **Direktverkauf zählt in beide Zahlen** (er gibt aus *und* nimmt ein) — als
  konsequente Anwendung des „rausgegeben/eingenommen"-Prinzips bestätigt.
- **Sortierung nach Menge absteigend** (größte Menge zuerst, Meistverkaufte oben)
  je Kategorie — wie festgelegt.
- Die bewusste Trennung der Grundlagen ist ein Feature, kein Bug: eine
  zurückgenommene Portion bleibt „ausgegeben", mindert aber den Umsatz. Der
  erklärende Hinweis im Report verhindert Fehlinterpretationen.
- Produkt-Konservatismus: Das Feature bleibt bewusst flach (zwei Zahlen, eine
  Ranking-Liste, pro Sitzung). Keine Konfigurierbarkeit, keine Diagramme, keine
  Vorratsfelder — jede Erweiterung müsste sich am realen Bedarf ehrenamtlicher
  Teams neu rechtfertigen.
