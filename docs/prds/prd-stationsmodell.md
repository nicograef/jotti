# PRD: Ausgabestationen unabhängig von Produktkategorien

## Problem Statement

Bei einem Fest gibt es oft mehrere Ausgabebereiche, und jeder braucht seinen eigenen Bondrucker. jotti leitet einen Arbeitsbon allein über die Produktkategorie an einen Drucker. Es gibt drei feste Kategorien, also höchstens drei Drucker für Arbeitsbons.

Ein Verein mit drei räumlich getrennten Ausgabebereichen stieß im Echtbetrieb an diese Grenze:

- Ein Gericht stand unter „Getränke“, weil es neben dem Getränkedrucker zubereitet wurde.
- Ein Bereich für Kaffee und Kuchen bekam keinen eigenen Drucker und führte eine Strichliste.
- Zwei Essensausgaben mit je eigenem Drucker lassen sich nicht abbilden.

Die falsche Kategorie kostet an drei Stellen. Die Servicekraft sucht das Gericht unter dem Getränke-Reiter. Die Produktstatistik zählt es als Getränk, und der DSFinV-K-Export meldet die Kategorie als Warengruppe.

Zwei weitere Punkte des Berichts baut #146 getrennt: Der Piepton entfällt, und der Abholbon bekommt den Bonmodus „Kein Bon“. Mit ihm druckt der Direktverkauf gar nichts. Dieses PRD behandelt nur die Bindung des Druckers an die Kategorie.

## Solution

Eine Ausgabestation ist ein eigener Datensatz mit frei wählbarem Namen, Drucker-IP und Bonmodus. Der Admin legt so viele Stationen an, wie das Fest Ausgabebereiche hat.

Jede Kategorie zeigt auf eine Station. Ein einzelnes Produkt kann davon abweichen und auf eine andere Station zeigen. Jede Position druckt an genau einer Station: an der ihres Produkts, sonst an der ihrer Kategorie.

| Zuordnung                     | Ausgabestation | Drucker        |
| ----------------------------- | -------------- | -------------- |
| Kategorie Essen               | Ausgabe        | 192.168.50.100 |
| Produkt Zwiebelkuchen (Essen) | Backhaus       | 192.168.50.101 |
| Kategorie Getränke            | Theke          | 192.168.50.102 |
| Produkt Kaffee (Getränke)     | Kaffeeecke     | druckt nicht   |

Eine Station ohne Drucker-IP druckt nicht. So kann der Admin Stationen vor dem Aufbau anlegen und die IPs erst vor Ort eintragen.

Nach dem Update druckt jede Installation unverändert weiter. Das Update legt je Kategorie eine gleichnamige Station mit derselben IP und demselben Bonmodus an. Niemand muss etwas einstellen.

Kassenbeleg und Abholbon bleiben Sonderdrucker mit eigener Einstellung. Am fiskalischen Kassenbeleg ändert sich nichts.

## User Stories

1. Als Admin möchte ich Ausgabestationen mit eigenem Namen anlegen, damit jeder Ausgabebereich einen eigenen Drucker bekommt.
2. Als Admin möchte ich je Kategorie eine Standardstation wählen, damit ich nicht jedes Produkt einzeln zuordnen muss.
3. Als Admin möchte ich ein einzelnes Produkt an eine andere Station leiten, damit seine Kategorie fachlich richtig bleibt.
4. Als Admin möchte ich je Station sehen, was dort druckt, damit ich den Aufbau vor Ort prüfen kann.
5. Als Admin möchte ich an jede Station einen Testbon schicken, damit ich Drucker und Netzwerk vor dem Fest prüfe.
6. Als Admin möchte ich eine Station ohne Drucker-IP anlegen, damit ich Namen und Zuordnung vor dem Aufbau vorbereite.
7. Als Admin möchte ich eine Station löschen, damit ihre Produkte wieder über ihre Kategorie drucken.
8. Als Helfer an einer Ausgabe möchte ich nur die Bons meines Bereichs bekommen, damit ich nichts aussortieren muss.
9. Als Servicekraft möchte ich jedes Gericht unter „Essen“ finden, damit ich unter Stress nicht suchen muss.
10. Als Vorstand möchte ich eine Produktstatistik mit richtigen Kategorien, damit die Auswertung des Fests stimmt.
11. Als Betreiber möchte ich nach dem Update nichts umstellen, damit das nächste Fest wie gewohnt druckt.

## Implementation Decisions

### Abwägung

| Variante                            | Umfang                                                                                   | Aufwand | Risiko                                            | Umkehrbarkeit                               |
| ----------------------------------- | ---------------------------------------------------------------------------------------- | ------- | ------------------------------------------------- | ------------------------------------------- |
| A: Stationsmodell (Minimalvariante) | Stationen als Datensätze, Standardstation je Kategorie, Ausnahme je Produkt              | mittel  | gering, Events und Kassenbeleg bleiben unberührt  | Ausnahmen per Klick, Schema nur per Backup  |
| B: Vollmodell                       | wie A, Zuordnung je Variante, Kassenbeleg und Abholbon als Rollen in der Stationstabelle | hoch    | mittel, der fiskalische Kassenbeleg-Pfad zieht um | gering, zwei Druckarten teilen eine Tabelle |
| C: Druckkategorie am Produkt        | ein Produkt druckt am Drucker einer anderen Kategorie                                    | gering  | gering                                            | Ausnahmen per Klick, eine optionale Spalte  |
| 0: nicht bauen                      | Piepton-Wegfall und „Kein Bon“ genügen                                                   | keiner  | Kategorie bleibt Steuerungsmittel                 | –                                           |

Empfehlung, falls gebaut wird: Variante A. Sie löst alle drei Fälle des Berichts mit drei kleinen Tabellen. Variante C löst nur das falsch einsortierte Gericht und bleibt bei drei Druckern. Variante B verlangt eine Zuordnung je Produktvariante, löst aber keinen Fall, den A nicht löst.

Die Ausnahme hängt am Produkt, nicht an der Variante. Braucht eine Variante einen anderen Ort, wird sie ein eigenes Produkt. Der Bestell-Bildschirm zeigt ohnehin eine flache Variantenliste (D10), die Servicekraft merkt davon nichts.

### Modell

- Ausgabestation: Name, Drucker-IP (IPv4 oder leer), Bonmodus (pro Position oder pro Bestellung), Status aktiv oder gelöscht.
- Der Name ist Pflicht und unter aktiven Stationen eindeutig. Er erscheint auf dem Testbon und in der Verwaltung.
- Löschen ist ein Soft Delete wie bei allen Stammdaten.
- Kategorie-Zuordnung: Jede der drei Kategorien zeigt auf genau eine aktive Station. Eine Lücke gibt es nicht.
- Produkt-Ausnahme: optional, höchstens eine je Produkt.
- Auflösung: Produkt-Ausnahme vor Kategorie-Zuordnung. Diese Regel lebt an genau einer Stelle im Druck-Modul.
- Eine Station löschen entfernt ihre Produkt-Ausnahmen in derselben Transaktion. Ist sie Standardstation einer Kategorie, lehnt das Backend das Löschen ab.

### Schema

Eine additive Migration mit der nächsten freien Nummer nach Migration 10 aus #146, in einer Transaktion:

- Tabelle `ausgabestationen` mit Name, Drucker-IP, Bonmodus und Status. Bonmodus und Status sind TEXT mit CHECK.
- Tabelle `kategorie_ausgabestationen`: Kategorie als Primärschlüssel, Fremdschlüssel auf die Station, nie NULL.
- Tabelle `produkt_ausgabestationen`: Produkt als Primärschlüssel, Fremdschlüssel auf die Station.
- Eigene Tabelle statt Spalte am Produkt: Das Produkt-Aggregat bleibt frei von Druckbelangen. Die Zuordnung gehört dem Druck-Modul.
- Datenübernahme: Je Zeile `essen`, `getraenk` und `sonstiges` in `druckstationen` entsteht eine Station mit IP und Bonmodus.
- Die Stationen heißen wie die Kategorien in der Oberfläche: „Essen“, „Getränke“ und „Sonstiges“. Jede Kategorie zeigt auf ihre Station.
- Die drei Kategoriezeilen in `druckstationen` bleiben unverändert stehen; der Code liest dort nur noch die Sonderdrucker (offene Entscheidung 2).
- So bleibt die Migration additiv, wie die Migrationsregeln verlangen. Den CHECK aus Migration 10 mit `'kein_bon'` fasst sie nicht an.
- Produkt-Ausnahmen übernimmt die Migration keine. Die Kategorie-Zuordnung deckt jede bestehende Installation ab.

Kein Event ändert sich, und es entsteht kein Event-Typ. Das Kassenjournal bleibt unberührt. Die Druckkonfiguration ist CRUD wie die heutige Druckstationen-Einstellung.

### Arbeitsbon-Policy

- Die Policy bleibt eine reine Funktion. Der Aufrufer lädt je Variante der Bestellung die aufgelöste Station und übergibt sie.
- Die Policy gruppiert Positionen nach Station statt nach Kategorie. Zwei Kategorien an einer Station ergeben im Modus pro Bestellung einen Sammelbon.
- Eine Position ohne Station mit Drucker-IP erzeugt keinen Druckauftrag.
- Die Auflösung nutzt die Zuordnung im Moment der Bestellung. Ein Arbeitsbon entsteht genau dann und nie ein zweites Mal.
- Direktverkauf: Erst greift der Abholbon-Bonmodus „Kein Bon“, dann der Abholbon, sonst druckt jede Position an ihre aufgelöste Station.
- Testbon: je Ausgabestation, die Druckauftrags-Referenz trägt den Stationsnamen. Der Testbon der Sonderdrucker bleibt gleich.
- Das Print-Relay ändert sich nicht. Es sieht weiter nur Ziel-IP und Payload.

### API und Verträge

- Alle Endpunkte sind POST und prüfen mit zog im Backend und Zod im Frontend.
- Ausgabestationen auflisten, anlegen, ändern und löschen.
- Die Liste trägt je Station ihre Kategorien und Produkte, fertig vom Backend zusammengestellt.
- Standardstation einer Kategorie setzen, Produkt-Ausnahme setzen und entfernen.
- Testbon je Ausgabestation drucken.
- Der Sonderdrucker-Endpunkt nimmt allein Kassenbeleg und Abholbon an. Die alten Kategoriezeilen sind über die API nicht mehr erreichbar.
- Die Produktliste liefert je Produkt den Namen seiner wirksamen Station. Das Frontend schlägt keine Station selbst nach.
- Backend und Frontend ändern sich im selben Release, die API trägt keine Version.

### Admin-UX

Die Seite „Bondrucker“ hat drei Abschnitte:

1. Ausgabestationen: je Station eine Karte mit Name, Drucker-IP, Bonmodus, Testbon und Löschen. Darunter steht, was dort druckt, etwa „Kategorie Essen · Zwiebelkuchen“.
2. Zuordnung: je Kategorie ein Auswahlfeld für die Standardstation. Darunter eine Liste „Ausnahmen“ mit Produkt, Station und Entfernen.
3. Sonderdrucker: Kassenbeleg und Abholbon wie heute, der Abholbon samt Bonmodus „Kein Bon“.

Die Ausnahme steht auf der Stationsseite, nicht im Produktdialog. Die ganze Druckeinrichtung geschieht so vor Ort auf einer Seite. Die Produktliste zeigt die wirksame Station nur als Hinweis.

Gespeichert wird beim Verlassen eines Felds, wie auf der Seite heute. Der Knopf „Station hinzufügen“ legt eine Station ohne IP an.

## Testing Decisions

Tests prüfen nur äußeres Verhalten. Erweitert werden die bestehenden Testdateien.

- Arbeitsbon-Policy: Tabellentests wie die heutigen je Bonmodus. Fälle: Ausnahme schlägt Kategorie, zwei Kategorien an einer Station, Station ohne IP, Direktverkauf ohne Abholbon.
- Domäne: Die bestehenden Tabellentests der Druckstation prüfen Name, IP und Bonmodus der Ausgabestation.
- Repository: Integrationstest gegen echtes Postgres wie im Druckstationen-Repository. Er prüft Auflösung, Löschsperre und das Entfernen der Ausnahmen.
- Migration: Ein Integrationstest spielt die Migrationsdatei auf eine Datenbank mit konfigurierten Kategoriezeilen. Vorbild ist der Test der Favoriten-Bereinigung.
- Der Migrationstest prüft: Jede Kategorie druckt danach an dieselbe IP mit demselben Bonmodus. Die alten Kategoriezeilen stehen unverändert. Der CI-Job für den Upgrade-Pfad deckt Boot und Projektionen ab.
- HTTP: nur die Löschsperre als Fehlercode. Die Feldabbildung deckt der Integrationstest ab.
- Frontend: die Zod-Prüfung in der Backend-Klasse, wie im bestehenden Druckstationen-Backend-Test.
- E2E: Die bestehende Spezifikation „Druckstationen verwalten“ legt eine Station an, setzt eine Ausnahme und schickt einen Testbon.
- Events: keine, abgedeckt durch den Vertragstest der Event-JSON-Formate.
- Produktstatistik und DSFinV-K: keine, abgedeckt durch ihre bestehenden Tests. Das Modell berührt beide nicht.

## Out of Scope

- Zuordnung je Variante. Eine Variante mit eigenem Ort wird ein eigenes Produkt.
- Piepton je Station. Der Piepton entfällt für alle Stationen.
- Der Abholbon-Bonmodus „Kein Bon“. Er wird getrennt gebaut (#146).
- Eine Position an zwei Stationen, etwa als Kopie für die Theke.
- Mehrere Kassenbeleg- oder Abholbon-Drucker.
- Frei definierbare Kategorien. Es bleiben die drei festen Werte.
- Umleiten offener Druckaufträge, wenn ein Drucker ausfällt oder seine IP wechselt.
- Eine Auswertung je Station.
- Kategorien vergangener Kassensitzungen umschreiben. Die Positionen im Kassenjournal bleiben, wie sie sind.
- Kommentar je Position (#145). Dafür gibt es ein eigenes PRD.

## Further Notes

**Bauentscheidung.** Den Bedarf meldet ein Verein. Die getrennten Arbeiten aus #146 lösen zwei seiner drei Punkte.

Für den Bau spricht: K-25 nennt die Konfiguration der Ausgabestationen und die Zuordnung der Kategorien als zwei Aufgaben. Die zweckentfremdete Kategorie verfälscht Produktstatistik und DSFinV-K-Warengruppe.

Gegen den Bau spricht: Ein dritter Ausgabeort geht heute schon über „Sonstiges“, solange die Kategorie frei ist. Die Verwaltung bekommt eine Seite mit drei Abschnitten statt einer Liste.

**Offene Entscheidungen für den Owner:**

1. Bauen oder nicht. Empfohlen ist Variante A, falls gebaut wird; sonst Variante 0.
2. Die alten Kategoriezeilen in `druckstationen`: behalten (empfohlen) oder löschen.

Zu Entscheidung 2:

- Behalten hält die Migration additiv. Der Code ignoriert die Zeilen.
- Löschen samt engerem CHECK wäre eine ausdrückliche Ausnahme des Owners von der additiven Migrationsregel.
- Für das Löschen spricht nur: Tote Zeilen mit veralteten IPs verwirren beim Blick in die Datenbank.
- Auch ein engerer CHECK behält `'kein_bon'` am Abholbon.

**Auswirkung auf die Produktstatistik.** Sie gruppiert nach der Kategorie, die jede Position im Kassenjournal trägt. Das Stationsmodell ändert an ihr keinen Code.

Der Gewinn ist mittelbar: Ein Verein kann jedes Produkt in seine richtige Kategorie legen. Ab der nächsten Bestellung zählt es dort, vergangene Kassensitzungen bleiben unverändert.

Wer ein Produkt während einer laufenden Kassensitzung umsortiert, sieht es in zwei Kategorie-Abschnitten. Das Handbuch sollte raten, Kategorien vor dem Eröffnen der Kassensitzung zu ordnen.

**Folgeänderungen bei Umsetzung.** Der Handbuch-Abschnitt zum Bondruck und die Druckstation-Einträge der Sprachtabelle werden angepasst. Die Anforderung K-25 bleibt wörtlich gültig.

**Vorgeschlagene Einträge für die Sprachtabelle:**

| Begriff          | Definition                                                                                              | Zu vermeiden                               |
| ---------------- | ------------------------------------------------------------------------------------------------------- | ------------------------------------------ |
| Ausgabestation   | Benannter Ausgabeort mit Bondrucker und Bonmodus, an den Arbeitsbons gehen.                             | Produktstation, Kategorie-Station, Drucker |
| Standardstation  | Die Ausgabestation, an die eine Kategorie ihre Positionen schickt, wenn das Produkt keine Ausnahme hat. | Kategorie-Drucker, Default-Station         |
| Produkt-Ausnahme | Zuordnung eines einzelnen Produkts zu einer anderen Ausgabestation als der seiner Kategorie.            | Override, Routing-Regel, Sonderzuordnung   |
| Sonderdrucker    | Kassenbeleg- und Abholbon-Drucker, die nicht nach Produkten drucken.                                    | Sonderstation, Spezialstation              |
| Druckstation     | Sammelbegriff der Verwaltungsseite für Ausgabestationen und Sonderdrucker.                              | Station allein im Code                     |
