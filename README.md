# jotti

Jotti ist ein einfaches Bestellsystem für Veranstaltungen und kleine Gastronomiebetriebe.

## Entities

**Benutzer**

- Benutzer entsprechen in Jotti Mitarbeiter:innen der Gastroeinrichtung bzw. Veranstaltung.
- Benutzer können Servicekräfte, PoS-Personal oder auch Verwaltungspersonal sein.
- Alle Aktionen in Jotti werden von jeweils einem Benutzer durchgeführt.

**Produkt**

- Ein Produkt entspricht in Jotti einem Getränk oder einem Gericht.
- Ein Produkt hat einen Namen und einen Preis.
- Produkte werden über Bestellungen verkauft.
- Ein Produkt kann ausverkauft sein und kann in diesem Fall nicht mehr bestellt werden.
- Produkte werden von Administratoren verwaltet.

**Kunde**

- Kunden bestellen Produkte und müssen diese anschließend bezahlen.
- Kunden werden von Administratoren verwaltet.


## Aggregates

**Bestellung**

- Jede Bestellung wird auf einen Tisch gebucht.
- Eine Bestellung "kauft" eine Liste von Produkten für einen Tisch.

**Bezahlung**

- Jede Bezahlung wird auf einen Tisch gebucht.

## Events

**Bestellung aufgegeben**
type: "bestellung-aufgegeben:v1"

**Bezahlung getätigt**
type: "bezahlung-getaetigt:v1"

---

## Ideen für spätere Weiterentwicklungen

- Optionen für Produkte wie Beispielsweise "mit Soße".
