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
- Ein Produkt kann ausverkauft sein und kann in diesem Fall nicht 
mehr bestellt werden.

**Tisch**
- Ein Tisch entspricht in Jotti einem Kunden.

## Aggregates

**Bestellung**
- Jede Bestellung wird auf einen Tisch gebucht.
- Eine Bestellung "kauft" eine Liste von Produkten für einen Tisch.

**Bezahlung**
- Jede Bezahlung wird auf einen Tisch gebucht.


## Events

**Bestellung aufgegeben**
 type: "jotti.bestellung.aufgegeben:v1"
 

**Bezahlung getätigt**