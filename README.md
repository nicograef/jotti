# jotti

Jotti ist ein einfaches Bestellsystem für Veranstaltungen und kleine Gastronomiebetriebe.

## Entities

**Benutzer**

- Benutzer entsprechen in Jotti Mitarbeiter:innen der Gastroeinrichtung bzw. Veranstaltung.
- Benutzer können Servicekräfte, PoS-Personal oder auch Verwaltungspersonal sein.
- Alle Aktionen in Jotti werden von jeweils einem Benutzer durchgeführt.
- Jeder Benutzer ist einer Rolle zugeordnet. Z.B. "Service", "Admin"

**Produkt**

- Ein Produkt entspricht in Jotti einem Getränk oder einem Gericht.
- Ein Produkt hat einen Namen und einen Preis.
- Produkte werden über Bestellungen verkauft.
- Ein Produkt kann ausverkauft sein und kann in diesem Fall nicht mehr bestellt werden.
- Produkte werden von Administratoren verwaltet.
- Produkte können mehrfach bestellt werden.
- Beispiel: Kaffee, Bier, Pommes, Pizza

**Kunde**

- Kunden bestellen Produkte und müssen diese anschließend bezahlen.
- Kunden werden von Administratoren verwaltet.
- Beispiel: "Tisch 1", "Tisch 23", "Sonderposten"

## Aggregates

**Bestellung**

- Jede Bestellung wird auf einen Kunden gebucht.
- Eine Bestellung beinhaltet eine Liste von Produkten mit Mengenangabe.

**Bezahlung**

- Jede Bezahlung wird auf einen Kunden gebucht und beinhaltet eine Liste von Produkten (und Mengenangaben)
- Bezahlungen können nur getätigt werden, wenn die ausgewählten Produkte (inkl. der angegebenen Menge) bei diesem Kunden noch unbezahlt sind.
- Bezahlungen sind unabhängig von Bestellungen. D.h. es wird nicht eine Bestellung bezahlt, sondern eine Menge von Produkten.

## Events

**Bestellung aufgegeben**
Beschreibt, dass ein Kunde eine Bestellung aufgegeben hat. Das Subject gibt den Kunden an, die bestellten Produkte und ihre Menge sind in der Payload angegeben.

```json
{
  "id": "3f12e1fe-92f1-4d97-b9d3-0a8d4a3c2f10",
  "type": "bestellung-aufgegeben:v1",
  "subject": "customer:tisch17",
  "time": "2025-10-21T17:59:14.966Z",
  "payload": {
    "products": [
      { "id": 3, "amount": 2 },
      { "id": 14, "amount": 3 }
    ]
  }
}
```

**Bezahlung getätigt**
Beschreibt, dass ein Kunde eine Bezahlung getätigt hat.

```json
{
  "id": "3f12e1fe-92f1-4d97-b9d3-0a8d4a3c2f10",
  "type": "bezahlung-getaetigt:v1",
  "subject": "customer:tisch17",
  "time": "2025-10-21T18:59:14.966Z",
  "payload": {
    "products": [
      { "id": 3, "amount": 1 },
      { "id": 14, "amount": 1 }
    ]
  }
}
```

---

## Ideen für spätere Weiterentwicklungen

- Optionen für Produkte wie Beispielsweise "mit Soße".
