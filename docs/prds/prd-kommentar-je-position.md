# PRD: Kommentar je Position

## Problem Statement

Die Servicekraft erfasst genau einen Kommentar je Bestellung, nicht je Position.
Der Arbeitsbon trägt ihn deshalb an jeder Station, die eine Position dieser Bestellung druckt.

- Bei getrennten Stationen steht „ohne Kümmel" zum Zwiebelkuchen auch auf dem Getränkebon.
- Im Bonmodus „pro Position" steht derselbe Kommentar auf jedem Einzelbon der Bestellung.
- Ein Verein zählte an einem zweitägigen Fest 171 Bestellungen mit Kommentar unter 1603 Bestellungen.
- Bei 76 davon (44 %) ging der Kommentar an mindestens zwei Drucker.
- Fast jeder häufige Kommentar gehört zu einer Position: „2 Gläser" zur Weinflasche, „alkoholfrei" zum Bier.
- Nur Hinweise wie „Stehtisch" betreffen die ganze Bestellung.
- Die Ausgabe musste deshalb nachfragen, zu welcher Position ein Kommentar gehört.

## Solution

Die Servicekraft kann im Bestellabschluss zu jeder Position einen eigenen Kommentar erfassen.
Dieser Positionskommentar erscheint nur auf dem Arbeitsbon der Station, die die Position zubereitet.
Der Bestellkommentar bleibt der Standard für Hinweise zur ganzen Bestellung.

- Beide Kommentare sind optional und höchstens 100 Zeichen lang.
- Voraussetzung ist der Backlog-Punkt „Zeichen gegen Bytes": Frontend und Backend zählen dann dieselbe Einheit.
- Ohne Kommentar bestellt die Servicekraft mit genau so vielen Tipps wie heute.
- Der Direktverkauf bekommt denselben Positionskommentar, auch auf dem Abholbon.
- Die Historie zeigt den Positionskommentar unter seiner Position.
- Kassenbeleg, Zahlung, Stornierung, Korrektur und Umbuchung bleiben, wie sie sind.
- Betrifft ein Kommentar nur einen Teil der Menge, schreibt die Servicekraft das hinein: „1 davon ohne Kümmel".

## User Stories

1. Als Servicekraft möchte ich einen Kommentar an eine einzelne Position hängen, damit nur die zuständige Station ihn bekommt.
2. Als Servicekraft möchte ich einen Kommentar an die ganze Bestellung hängen, damit „Stehtisch" alle Stationen erreicht.
3. Als Servicekraft möchte ich ohne Kommentar ohne Zusatzschritt bestellen, damit der Normalfall unter Zeitdruck schnell bleibt.
4. Als Servicekraft möchte ich Positions- und Bestellkommentar in einer Bestellung kombinieren, damit beide Arten von Hinweisen ankommen.
5. Als Servicekraft möchte ich, dass der Kommentar mit seiner Position aus dem Warenkorb verschwindet, damit kein verwaister Hinweis druckt.
6. Als Servicekraft im Direktverkauf möchte ich Positionskommentare erfassen, damit Abholbon und Stationsbon sie tragen.
7. Als Servicekraft möchte ich den Positionskommentar in der Historie sehen, damit ich Rückfragen der Station beantworten kann.
8. Als Helfer an einer Ausgabestation möchte ich nur die Kommentare meiner Positionen lesen, damit ich nicht nachfragen muss.
9. Als Helfer an einer Ausgabestation möchte ich den Kommentar direkt unter seiner Position lesen, damit die Zuordnung eindeutig ist.
10. Als Helfer an einer Station mit Bonmodus „pro Position" möchte ich den Kommentar nur auf seinem Einzelbon sehen.
11. Als Serviceleitung möchte ich Positionen mit Kommentar wie jede andere stornieren und umbuchen, damit diese Abläufe gleich bleiben.
12. Als Admin möchte ich TSE-Daten und DSFinV-K-Export unverändert behalten, damit die Kasse ohne erneute Prüfung konform bleibt.

## Implementation Decisions

### Abgewogene Ansätze

| Ansatz                                                                         | Aufwand                                                                                                                                | Risiko                                                                         | Umkehrbarkeit                                                        |
| ------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------ | -------------------------------------------------------------------- |
| A: optionales Feld `kommentar` in den Positionen der zwei v1-Erfassungs-Events | klein: ein Event-Format, Arbeitsbon, Historie, API, zwei Abschlüsse                                                                    | gering: Alt-Events byte-identisch, SQL unberührt                               | Feature entfernbar, geschriebene Felder bleiben wirkungslos          |
| B: Version v2 von `bestellung-aufgenommen` und `direktverkauf-getaetigt`       | groß: 15 Go-Dateien, 24 SQL-Literale (19 in den Queries, 5 in `01_initial`), zwei SQL-Funktionen, zwei Idempotenz-Indexe per Migration | hoch: ein übersehenes Literal fehlt still im Soll-Bestand und Tagesabschluss   | v2 bleibt für immer lesepflichtig, jede Abfrage kennt zwei Versionen |
| C: Bestellkommentar einer Kategorie zuordnen                                   | mittel: ebenfalls ein Event-Feld plus Kategorie-Auswahl                                                                                | Fehlzuordnung bleibt bei Bonmodus „pro Position" und innerhalb einer Kategorie | wie A                                                                |

Gewählt ist A.

- A löst das Problem vollständig und fasst weder Schema noch Reporting noch Kassensturz an.
- B kostet ein Vielfaches und gefährdet kassenwirksame Summen für ein rein operatives Feld.
- C verlangt eine Entscheidung unter Zeitdruck und trennt „2 Gläser" (Wein) nicht von „alkoholfrei" (Bier).
- A folgt dem Muster des optionalen Benutzerkommentars der Umbuchung: additiv, leer weggelassen, Alt-Events unverändert.
- A ist eine ausdrückliche Ausnahme vom Freeze der Event-Formate und braucht die Freigabe des Owners.
- Verweigert der Owner die Ausnahme, ist B der Rückfall; C scheidet in jedem Fall aus.

### Event-Format

- Die Positionen von `bestellung-aufgenommen:v1` und `direktverkauf-getaetigt:v1` erhalten das optionale Feld `kommentar`.
- Ist es leer, fehlt der Schlüssel im JSON; solche Events sind byte-identisch zum Bestandsformat.
- Ein Alt-Event liest sich mit leerem Positionskommentar; nichts wird migriert oder umgedeutet.
- Das Feld steht in einer eigenen Event-Form der Bestellposition, nur in diesen zwei Events.
- Die geteilte Positionsform von Zahlung, Stornierung, Korrektur und Umbuchung bleibt unberührt.
- So kann der Positionskommentar nie in kassenwirksame Events wandern.
- Das Event-Schema prüft das Feld optional mit höchstens 100 Zeichen, wie den Bestellkommentar.

### API

- Die Positionen von „Bestellung aufnehmen" und „Direktverkauf tätigen" erhalten das optionale Feld `kommentar`.
- zog im Backend und Zod im Frontend prüfen beide höchstens 100 Zeichen.
- Heute zählt Zod Code-Units, zog Bytes; ein Umlaut-Kommentar mit 100 Zeichen bekommt deshalb 400.
- Der Positionskommentar wird erst gebaut, wenn der Backlog-Punkt „Zeichen gegen Bytes" erledigt ist.
- Das Backend nimmt dieselbe Variante mehrfach je Bestellung an; jede Eingabe wird eine eigene Position.
- Die Historien-DTOs von Bestellung und Direktverkauf tragen den Positionskommentar je Position, leer als Leerstring.
- Die Positionen der Tisch-Ansicht und der Zahlungsauswahl tragen ihn nicht.

### Domäne, Projektion und Replay

- Die Domänen-Position trägt den Positionskommentar nur beim Lesen der zwei Erfassungs-Events.
- Die Projektion `tisch_sessions` übernimmt ihn nicht, weil Zahlung, Stornierung und Umbuchung ihn nicht brauchen.
- Replay und Projektionsaufbau ändern sich deshalb nicht.
- Es gibt keine Migration; SQL-Funktionen, Idempotenz-Indexe und Reporting-Abfragen bleiben gleich.

### Frontend

- Der Warenkorb bleibt je Variante geschlüsselt.
- Der Bestellabschluss hält die Positionskommentare als eigene Zuordnung von Variante zu Text, neben dem Bestellkommentar.
- Jede Zeile im Abschluss zeigt eine Schaltfläche „Kommentar"; ein Tipp öffnet ein einzeiliges Feld unter der Zeile.
- Fällt die Menge einer Variante auf null, verwirft der Abschluss ihren Kommentar.
- Startet ein Vorgang aus leerem Warenkorb, setzt der Abschluss alle Positionskommentare zurück wie den Bestellkommentar.
- Die Positionskommentare gehören zum Inhalt des Absendeversuchs.
- Ein Wiederholversuch mit geändertem Kommentar bekommt dadurch einen frischen Idempotenz-Schlüssel.
- Der Direktverkauf-Abschluss verhält sich gleich.
- Die Historie zeigt den Positionskommentar in Anführungszeichen unter seiner Positionszeile.
- Der State lebt in Hooks des Abschlusses; es gibt keinen globalen Store.

Verworfen ist ein Warenkorb je Paar aus Variante und Kommentar.
Er ändert Mengensteuerung, Produktliste und Abschlussliste und kostet unter Zeitdruck Tipps.
Die Felddaten zeigen Kommentare zur ganzen Zeile; Teilmengen deckt der Freitext ab.

### Arbeitsbon und Abholbon

- Die Arbeitsbon-Policy gruppiert die Positionen nach Kategorie.
- Der Positionskommentar reist mit seiner Position und erreicht nur deren Station.
- Sammelbon: Der Positionskommentar steht eingerückt direkt unter seiner Positionszeile.
- Der Bestellkommentar steht am Bonende, auf jedem Bon der Bestellung.
- Bonmodus „pro Position": Der Einzelbon trägt den Kommentar seiner Position und den Bestellkommentar.
- Abholbon „pro Stück": Jeder Stückbon trägt den Kommentar seiner Position.
- Ein Teilmengen-Kommentar wie „1 davon ohne Kümmel" erscheint dort auf jedem Stückbon dieser Position.
- Zeichensatz und Umbruch auf 48 Zeichen folgen dem Bestellkommentar.

### Kassenbeleg, TSE und DSFinV-K

- Der Kassenbeleg ändert sich nicht: Er listet die Positionen der Zahlung, die keinen Positionskommentar trägt.
- Der Kassenbeleg des Direktverkaufs druckt den Positionskommentar ebenfalls nicht.
- Die TSE-processData bleiben Menge, Bezeichnung und Preis; der Kommentar ist nicht Teil der Signatur.
- Der DSFinV-K-Export übernimmt den Positionskommentar nicht.
- Die Spezifikation kennt nur die Bonkopf-Notiz `BON_NOTIZ` mit 255 Zeichen, kein Notizfeld je Zeile.
- `BON_NOTIZ` trägt unverändert genau den Bestellkommentar, ohne Kürzung.
- Das Kassenjournal bewahrt den Positionskommentar vollständig auf.

### Stornierung, Korrektur und Umbuchung

- Stornierung und Korrektur referenzieren Positionen per ID; ihre Events tragen keinen Positionskommentar.
- Die Umbuchung legt Zielpositionen ohne Kommentar an, denn der Arbeitsbon ist zu diesem Zeitpunkt gedruckt.
- Die Historie des Quelltischs zeigt den Kommentar an der Ursprungsbestellung.

## Testing Decisions

Getestet wird nur äußeres Verhalten an den öffentlichen Schnittstellen; erwartete Werte sind Literale.

- Event-Vertrag: beide Events mit Positionskommentar und ohne, ohne Schlüssel byte-identisch. Vorbild ist der Umbuchungsfall in `event_json_contract_test.go`.
- Ein Alt-Event ohne Positionskommentar validiert gegen das Schema und liest sich mit Leerstring.
- Arbeitsbon-Routing, die Regression aus #145: Essen mit Kommentar und Getränk ohne ergibt einen Getränkebon ohne Kommentar.
- Bonmodus „pro Position": Der Kommentar steht nur auf dem Einzelbon seiner Position. Vorbild ist `arbeitsbon_policy_test.go`.
- Bonlayout: Der Positionskommentar steht unter seiner Positionszeile, der Bestellkommentar am Ende. Vorbild ist `formatter_test.go`.
- Schemagrenzen im Backend: 100 Zeichen gültig, 101 abgelehnt. Erweitert werden `optionalKommentarTests` in `kommentar_test.go` und `schema_grenzen_test.go`.
- Schemagrenzen im Frontend: derselbe Fall für `BestellPositionInputSchema` und `VerkaufPositionInputSchema`.
- Dafür entstehen `Bestellung.test.ts` und `Direktverkauf.test.ts` neben ihren Schemas; heute testet keine Datei diese Schemas. Vorbild ist `Stornierung.test.ts`.
- Historie: Der Positionskommentar erscheint an seiner Position im Historien-Eintrag. Vorbild ist `historie_test.go`.
- Bestellabschluss: Der Request trägt den Kommentar nur an seiner Position. Vorbild ist `BestellungAbschluss.test.tsx`.
- Im Bestellabschluss verwirft eine auf null gesenkte Menge den Kommentar.
- Ein Wiederholversuch mit geändertem Positionskommentar sendet einen frischen Schlüssel.
- Direktverkauf-Abschluss: nur die Request-Prüfung, in `DirektverkaufAbschluss.test.tsx`.
- Projektion, TSE-processData, Kassenbeleg und DSFinV-K: keine, abgedeckt durch die bestehenden Tests, die unverändert grün bleiben.
- End-to-End: keine, Handler-, Policy- und Komponententests erreichen jeden Pfad.

## Out of Scope

- Verschiedene Kommentare für Teilmengen derselben Variante in einer Bestellung.
- Vordefinierte Kommentar-Chips oder Textbausteine je Produkt.
- Zuordnung des Bestellkommentars zu einer Kategorie.
- Positionskommentar bei Zahlung, Stornierung, Korrektur, Umbuchung und auf dem Kassenbeleg.
- Positionskommentar im Reporting und im DSFinV-K-Export.
- Nachträgliches Ändern eines Kommentars; Events sind unveränderlich.
- Migration oder Umdeutung von Alt-Events.

## Further Notes

Vorgeschlagener Eintrag für `docs/language.md`:

| Begriff            | Bedeutung                                                                                                                                                                                                        | Zu vermeiden                                 |
| ------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------- |
| Positionskommentar | Optionaler Freitext der Servicekraft zu einer Position einer Bestellung oder eines Direktverkaufs, gedruckt nur an deren Station. Go `Kommentar` an der Bestellposition · JSON/TS `kommentar` im Positionsobjekt | Positionsnotiz, Notiz, Hinweis, Sonderwunsch |
| Bestellkommentar   | Der Kommentar einer Bestellung, gedruckt auf jedem ihrer Arbeitsbons. Go `Kommentar` · JSON/TS `kommentar` auf Event-Ebene                                                                                       | Bestellnotiz, Bonkommentar                   |

Offene Entscheidungen für den Owner:

1. Freigabe des additiven Felds in den zwei v1-Events als Ausnahme vom Event-Freeze. Empfohlen; Rückfall ist Ansatz B.
2. DSFinV-K: Positionskommentar nicht exportieren (empfohlen) oder an `BON_NOTIZ` anhängen, mit Kürzung auf 255 Zeichen.
3. Abholbon „pro Stück": Kommentar auf jedem Stückbon (empfohlen, einfach) oder nur auf dem ersten.

Bei Umsetzung ändern sich die Beschreibung des Arbeitsbon-Inhalts und die Event-Tabelle im Handbuch.
Die Freigabe aus Entscheidung 1 gehört als Zeile in `docs/decisions.md`.
Der Melder aus #145 bietet an, das Feature zu bauen und zu testen.
