# PRD: Tischservice pro Servicekraft (Zuordnung, Filter, Schichtende)

## Problem Statement

Im Praxistest war der Tisch-Modus für die einzelne Servicekraft schwer zu
lesen. Eine Servicekraft sah an ihrem Favoriten-Tisch offene Positionen und war
verwirrt, weil sie dort selbst gar nichts aufgenommen hatte. Tatsächlich hatte
eine andere Servicekraft an diesem Tisch bestellt. Der Tisch-Zustand wird heute
nur als Ganzes angezeigt (ausstehende, unbezahlte Positionen, Saldo), ohne zu
unterscheiden, wer welche Position aufgenommen hat.

Daraus folgen drei konkrete Lücken:

- **Keine persönliche Erledigt-Sicht.** Eine Servicekraft kann nicht erkennen,
  ob sie persönlich noch offene Aufgaben hat, getrennt vom Gesamtzustand des
  Tisches.
- **Vermischung beim Kassieren und Stornieren.** Beim Kassieren und Stornieren
  sieht jede Servicekraft alle Positionen des Tisches. Mehrere Bedienungen
  kommen durcheinander, und Umsätze landen leicht bei der falschen Servicekraft.
- **Keine Schichtende-Prüfung pro Person.** Service-Schichten enden individuell.
  Der Admin kann nicht prüfen, ob von einer bestimmten Servicekraft noch Tische
  oder Positionen offen sind, bevor diese Person geht.

## Solution

Jede offene Position wird der Servicekraft zugeordnet, die die zugehörige
Bestellung aufgenommen hat (im Folgenden "Besteller"). Diese Zuordnung ist
bereits eindeutig vorhanden: Jedes Bestell-Event trägt die handelnde
Servicekraft, und Positions-IDs sind je Bestellung eindeutig. Die Zuordnung
fließt in die Tisch-Projektion ein und macht damit überall ableitbar, welche
offenen Positionen von wem stammen.

Darauf aufbauend:

- **Persönliche Erledigt-Sicht.** Die Servicekraft sieht eine eigene
  "Alles erledigt!"-Aussage über alle Tische, an denen sie bestellt hat. Sie
  gilt als erledigt, wenn keine ihrer bestellten Positionen mehr ausstehend
  (nicht ausgegeben) und keine mehr unbezahlt ist, unabhängig davon, wer
  ausgibt oder kassiert. So funktioniert auch die Schichtübergabe.
- **Eigene Positionen zuerst.** Beim Kassieren, Stornieren, Ausgabe-Bestätigen
  und Umbuchen stehen die eigenen Positionen oben und separat. Fremde Positionen
  sind eingeklappt und über "Alle anzeigen" erreichbar, damit eine Servicekraft
  bei Bedarf für eine abwesende Kollegin einspringen kann.
- **Schichtende-Prüfung im Live-Dashboard.** Der bestehende
  "Servicekräfte"-Tab zeigt pro Servicekraft zusätzlich die noch offene eigene
  Arbeit: an welchen Tischen noch eigene Positionen ausstehend oder unbezahlt
  sind. Der Admin kann so vor Schichtende prüfen, ob eine Person fertig ist.

Die Zuordnung des Umsatzes bleibt unverändert an die kassierende Servicekraft
gebunden. Indem das Kassieren der eigenen Positionen erleichtert wird, landet
der Umsatz in der Praxis bei der richtigen Person. Der Tisch-Gesamtzustand
(Gesamtsaldo, alle ausstehenden/unbezahlten Positionen) bleibt zusätzlich
sichtbar.

## User Stories

### Servicekraft: persönliche Sicht

1. Als Servicekraft möchte ich auf der Tisch-Übersicht eine eigene
   "Alles erledigt!"-Aussage sehen, damit ich vor Schichtende sicher bin, dass
   ich nichts Offenes hinterlasse.
2. Als Servicekraft möchte ich, dass diese Erledigt-Aussage alle Tische
   einbezieht, an denen ich bestellt habe, nicht nur meine Favoriten-Tische.
3. Als Servicekraft möchte ich, dass "für mich erledigt" gilt, sobald alle von
   mir bestellten Positionen ausgegeben und bezahlt sind, auch wenn eine
   Kollegin die Ausgabe oder Zahlung übernommen hat.
4. Als Servicekraft möchte ich sehen, an welchen konkreten Tischen ich noch
   offene eigene Positionen habe, damit ich gezielt dorthin gehe.
5. Als Servicekraft möchte ich auf einer Tischkarte erkennen, wie viele der
   offenen Positionen von mir stammen und wie viele von anderen (zum Beispiel
   "3 offen, davon 2 von dir"), damit ich nicht über fremde Bestellungen an
   meinem Favoriten-Tisch stolpere.
6. Als Servicekraft möchte ich auf der Tisch-Detailseite sehen, ob an diesem
   Tisch für mich alles erledigt ist, getrennt vom Gesamtzustand des Tisches.

### Servicekraft: Kassieren und Ausgabe

7. Als Servicekraft möchte ich beim Kassieren zuerst nur meine eigenen
   unbezahlten Positionen sehen, damit ich schnell die richtigen kassiere.
8. Als Servicekraft möchte ich fremde unbezahlte Positionen bei Bedarf über
   "Alle anzeigen" einblenden, damit ich für eine abwesende Kollegin kassieren
   kann.
9. Als Servicekraft möchte ich, dass mein eigener Umsatz korrekt mir zugeordnet
   wird, weil ich primär meine eigenen Positionen kassiere.
10. Als Servicekraft möchte ich beim Bestätigen der Ausgabe meine eigenen
    ausstehenden Positionen zuerst sehen, damit ich meine Bestellungen zuerst
    bediene.
11. Als Servicekraft möchte ich erkennen, von wem eine fremde Position stammt
    (Name der bestellenden Servicekraft), damit Rückfragen einfacher sind.

### Serviceleitung: Stornieren und Umbuchen

12. Als Serviceleitung möchte ich beim Stornieren zuerst die Bestellungen sehen,
    die ich selbst aufgenommen habe, damit ich nicht versehentlich fremde
    Positionen storniere.
13. Als Serviceleitung möchte ich fremde Bestellungen beim Stornieren bei Bedarf
    einblenden, damit ich auch fremde Positionen stornieren kann, wenn nötig.
14. Als Serviceleitung möchte ich beim Umbuchen zuerst meine eigenen
    Bestellungen sehen.

### Admin: Schichtende und Reporting

15. Als Admin möchte ich im Live-Dashboard pro Servicekraft sehen, ob noch
    Tische oder Positionen von dieser Person offen sind, damit ich am
    Schichtende prüfen kann, ob sie fertig ist.
16. Als Admin möchte ich pro Servicekraft die konkreten offenen Tische mit der
    Anzahl ihrer noch ausstehenden und unbezahlten eigenen Positionen sehen.
17. Als Admin möchte ich auch eine Servicekraft sehen, die offene eigene Arbeit
    hat, aber noch keine Zahlung kassiert hat, damit niemand übersehen wird.
18. Als Admin möchte ich erkennen, wenn eine Servicekraft keine offene eigene
    Arbeit mehr hat ("fertig"), damit ich ihre Schicht freigeben kann.
19. Als Admin möchte ich den bestehenden Umsatz pro Servicekraft (kassiert)
    weiterhin sehen, ergänzt um die offene eigene Arbeit derselben Person.

### Querschnitt

20. Als Nutzer des Systems möchte ich, dass die Zuordnung einer Position zum
    Besteller stabil bleibt, auch nach Teilzahlungen, Teilstornos und Umbuchung,
    damit die Sichten konsistent sind.
21. Als Nutzer des Systems möchte ich, dass der Tisch-Gesamtzustand
    (Gesamtsaldo, alle offenen Positionen, ausstehende Auszahlung) weiterhin
    sichtbar bleibt, zusätzlich zur persönlichen Sicht.
22. Als Betreiber möchte ich, dass der Direktverkauf von dieser Änderung
    unberührt bleibt, weil er keine Tische und keinen offenen Positionszustand
    kennt.

## Implementation Decisions

### Zuordnung der Positionen zum Besteller

- Die Domänen-/Projektions-Struktur `Position` erhält die Identität der
  bestellenden Servicekraft (`BestellerUserID`, `BestellerName`). Diese Felder
  sind reine Projektions-/Anzeige-Information.
- Die Event-Daten ändern sich nicht. Der Besteller wird beim Anwenden eines
  `bestellung-aufgenommen`-Events aus dem Event-Umschlag (`UserID`/`UserName`)
  übernommen. Die immutable Event-Form (`PositionEventData`) bleibt ohne
  Besteller-Feld, weil der Besteller dort bereits über den Event-Umschlag
  eindeutig ist.
- Beim Anwenden der Events tagt die Projektion jede neu bestellte Position mit
  ihrem Besteller. Reduktionen (Zahlung, Stornierung, Ausgabe) ordnen weiterhin
  über die Positions-ID zu und erhalten das Besteller-Tag. Positions-IDs sind je
  Bestellung eindeutig, daher ist die Zuordnung pro Position genau ein
  Besteller (gleiche Invariante wie bei den Restmengen in der Historie).
- Konsequenz: Die Umwandlung zwischen `Position` und `PositionEventData` ist
  keine reine Struktur-Konvertierung mehr, sondern bildet die Felder explizit
  ab. Die Projektion (`tisch_sessions`) speichert die Positionen inklusive
  Besteller im JSONB. Da es keine produktiven Instanzen gibt, wird die
  Projektion ohne Migration neu aufgebaut.

### Deep Module: offene Arbeit pro Servicekraft

- Neue, DB-freie Funktionen in der `kasse`-Domäne berechnen die offene Arbeit
  aus Tisch-Sessions:
  - Pro Tisch und Servicekraft: eigene ausstehende und eigene unbezahlte
    Positionen sowie das Kennzeichen "für diese Person an diesem Tisch
    erledigt".
  - Über mehrere Tisch-Sessions: ein Rollup pro Servicekraft mit der Liste der
    Tische, an denen noch eigene Arbeit offen ist, den jeweiligen Anzahlen und
    dem Gesamt-Kennzeichen "alles erledigt".
- "Erledigt" ist definiert als: keine eigenen ausstehenden Positionen (nicht
  ausgegeben) und keine eigenen unbezahlten Positionen. Die tischweite,
  ausstehende Auszahlung (negativer Saldo) fließt nicht in die persönliche
  Erledigt-Aussage ein, weil sie ein tischweites Kassenkonzept ist.
- Dieses Modul wird zweifach genutzt: vom Service-Endpunkt (gefiltert auf die
  anfragende Servicekraft) und vom Admin-Live-Dashboard (alle Servicekräfte).

### Tisch-Sichten (Service)

- Die Tisch-State-Abfragen (`GetTischState`, `GetMeineTischeState`) erhalten die
  ID der anfragenden Servicekraft (aus dem JWT) und liefern die offenen
  Positionen weiterhin als Liste, je Position ergänzt um `bestellerUserId` und
  `bestellerName`, sowie ein Kennzeichen "für mich erledigt" je Tisch.
- Das Frontend gruppiert anhand von `bestellerUserId` in eigene und fremde
  Positionen. Das entspricht dem bestehenden Muster der Historie, die schon
  heute `isFromUser` rein im Frontend ableitet. Der Tisch-Gesamtzustand wird
  weiterhin aus der vollständigen Liste dargestellt.
- Die persönliche Erledigt-Übersicht über alle Tische liefert ein
  Service-Endpunkt auf Basis des Rollup-Moduls, gefiltert auf die anfragende
  Servicekraft. Naheliegend ist eine Erweiterung der bestehenden
  `EigeneUebersicht` um die offene eigene Arbeit (Anzahl offener Tische, Anzahl
  eigener ausstehender und unbezahlter Positionen, Erledigt-Kennzeichen).

### Kassieren, Ausgabe, Stornieren, Umbuchen (Frontend)

- Kassieren und Ausgabe-Bestätigen: Die Positionsliste wird in "Meine" (oben,
  ausgeklappt) und "Andere" (eingeklappt, per "Alle anzeigen" erreichbar)
  getrennt. Die eigentliche Buchung bleibt unverändert.
- Stornieren und Umbuchen erfolgen weiterhin über die Historie je Bestellung.
  Die Historie kennt die bestellende Servicekraft bereits. Die Bestell-Einträge
  werden so sortiert/gruppiert, dass eigene Bestellungen oben stehen und fremde
  einklappbar sind.
- Es gibt keine harte Sperre für fremde Positionen. Die Trennung ist rein
  visuell und über "Alle anzeigen" auflösbar.

### Schichtende-Prüfung (Admin Live-Dashboard)

- `LiveReportingData` erhält die offene Arbeit pro Servicekraft (aus dem
  Rollup-Modul über die Tisch-Sessions der offenen Kassensitzung).
- Der bestehende "Servicekräfte"-Tab im Live-Dashboard zeigt pro Servicekraft
  weiterhin den kassierten Umsatz und zusätzlich die offene eigene Arbeit
  (offene Tische mit Anzahl ausstehender/unbezahlter eigener Positionen, oder
  ein "fertig"-Hinweis). Beide Datensätze werden über die Benutzer-ID
  zusammengeführt, damit auch Personen erscheinen, die offene Arbeit, aber noch
  keinen Umsatz haben.
- Die Daten stammen aus der Projektion `tisch_sessions` der offenen
  Kassensitzung. Für kleine Veranstaltungen (5 bis 50 Tische) ist das Laden
  aller Sessions der Kassensitzung unkritisch. Die Aggregation erfolgt in Go
  über das Rollup-Modul, nicht über JSONB-Aggregation in SQL.

### Abgrenzung zur bestehenden Umsatzzuordnung

- Der Umsatz pro Servicekraft bleibt an die kassierende Person gebunden (wie
  heute). Es wird keine neue, besteller-basierte Umsatzzuordnung eingeführt.
  Die "eigene Positionen zuerst"-Filter sind das Mittel, mit dem der Umsatz in
  der Praxis bei der richtigen Person landet.

## Testing Decisions

Ein guter Test prüft beobachtbares Verhalten, nicht die interne Umsetzung. Für
die reinen Funktionen heißt das: Eingabe sind Events bzw. Tisch-Sessions,
Ausgabe sind die berechneten Sichten und Kennzeichen. Es wird nicht geprüft,
wie akkumuliert/reduziert wird, sondern was am Ende für eine Servicekraft offen
oder erledigt ist.

- **Besteller-Tagging in der Projektion.** Erweiterung der bestehenden
  Tisch-Session-Tests: Positionen tragen nach einer Bestellung den korrekten
  Besteller; Zahlung, Stornierung und Ausgabe erhalten das Besteller-Tag;
  mehrere Bestellungen verschiedener Servicekräfte an einem Tisch bleiben
  korrekt zugeordnet. Prior Art: `domain/kasse/tisch_session_test.go`.
- **Offene Arbeit pro Servicekraft (Deep Module).** Eigene Unit-Tests für die
  reinen Funktionen. Fälle: nur eigene offene Positionen, nur fremde, gemischt,
  alles erledigt, ausstehend ohne unbezahlt und umgekehrt, eigene Arbeit ohne
  Zahlung, Schichtübergabe (fremde Kassierung schließt eigene Arbeit ab),
  mehrere Tische und mehrere Servicekräfte im Rollup.
- **Tisch-State-Abfragen pro Servicekraft.** Tests, dass die Abfragen die
  Positionen mit Besteller liefern und das Erledigt-Kennzeichen je Tisch korrekt
  setzen. Prior Art: `api/table/application/query_test.go`.
- **Live-Reporting pro Servicekraft.** Tests, dass die offene Arbeit pro
  Servicekraft korrekt aus den Sessions zusammengeführt wird und mit dem
  kassierten Umsatz über die Benutzer-ID zusammenfällt. Prior Art:
  `api/reporting/application/query_test.go`.
- **Frontend: Kassieren.** Eigene vs. fremde Positionen werden getrennt
  dargestellt, fremde sind erst nach "Alle anzeigen" sichtbar. Prior Art:
  `service/components/table/Zahlung.test.tsx`.
- **Frontend: Tischkarte.** Die Karte zeigt die Anzahl eigener offener
  Positionen und die persönliche Erledigt-Aussage korrekt an.

## Out of Scope

- **Übernahme/Neuzuordnung von Positionen.** Eine explizite "Tisch übernehmen"-
  oder Umzuordnungs-Funktion wird nicht gebaut. Eine Schichtübergabe ist über
  die Erledigt-Definition abgedeckt: Sobald eine Kollegin die eigenen Positionen
  ausgibt oder kassiert, gilt für den Besteller "erledigt".
- **Neue Umsatzzuordnung nach Besteller.** Der Umsatz pro Servicekraft bleibt
  kassiert-basiert.
- **Direktverkauf.** Keine Tische, kein offener Positionszustand, daher nicht
  betroffen.
- **Historische Tagesabrechnung.** Die offene Arbeit ist ein Live-Konzept der
  offenen Kassensitzung. Die abgeschlossene Tagesabrechnung wird nicht um offene
  Arbeit erweitert, weil dort alles abgeschlossen sein soll.
- **Berechtigungen.** Die Rollenrechte (wer storniert, wer auszahlt) bleiben
  unverändert. Die Filter sind rein visuell und sperren nichts.
- **Push/Benachrichtigung.** Es gibt keine aktive Benachrichtigung an Admin oder
  Servicekraft über offene Arbeit; die Sichten sind Pull/Anzeige.

## Further Notes

- Die Erledigt-Definition (ausgegeben und bezahlt, unabhängig vom Ausführenden)
  wurde bewusst gewählt, damit Schichtübergaben funktionieren. Sie ist
  positionsbasiert und ignoriert die tischweite Auszahlung.
- Der Besteller-Name wird wie bei den fetten Events zum Zeitpunkt der Bestellung
  festgehalten. Spätere Umbenennungen eines Benutzers ändern angezeigte
  Besteller-Namen alter Positionen nicht. Das ist konsistent mit dem
  bestehenden Verhalten (Reporting nutzt den zuletzt bekannten Namen).
- Die persönliche Erledigt-Übersicht berücksichtigt alle Tische der offenen
  Kassensitzung, an denen die Servicekraft bestellt hat, nicht nur ihre
  Favoriten. Favoriten ("Meine Tische") bleiben eine getrennte, frei wählbare
  Markierung und sind nicht gleichbedeutend mit Verantwortung.
- Begriffe: "Besteller" oder "bestellende Servicekraft" bezeichnet die
  Servicekraft, die die Bestellung aufgenommen hat. Falls dieser Begriff dauerhaft
  verwendet wird, sollte er in `docs/language.md` aufgenommen werden.
