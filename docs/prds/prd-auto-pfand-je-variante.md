# PRD: Auto-Pfand je Variante

## Problem Statement

Vereine mit Glas- oder Tellerpfand verkaufen das Pfand als eigene Position eines Pfand-Produkts.
Die Servicekraft tippt es heute zu jedem Getränk und jeder Speise von Hand dazu.

- Das kostet unter Zeitdruck einen zusätzlichen Tipp je Stück.
- Liegt das Pfand-Produkt in einer anderen Kategorie, kommt ein Kategoriewechsel hinzu.
- Vergisst die Servicekraft das Pfand, fehlt es im Verkauf.
- Die Pfandrückgabe zahlt das Glas später trotzdem aus.
- Je vergessenem Pfand fehlt der Kasse damit der Pfandbetrag.
- Kein Verein hat diesen Wunsch gemeldet; der Gläserpfand-Wunsch aus dem Feld betraf Rückgabe und Verrechnung.

## Solution

Der Admin verknüpft eine Variante mit einer Pfand-Variante, etwa „Bier 0,5 l" mit „Glaspfand 2,00 €".
Tippt die Servicekraft diese Variante, legt der Warenkorb je Stück ein Pfand dazu.

- Das Pfand bleibt im Warenkorb sichtbar und lässt sich wie jede Variante zurücknehmen.
- So bleibt der Gast mit mitgebrachtem Glas ohne Pfand.
- Die Zeile der verknüpften Variante zeigt den Pfandbetrag unter ihrem Preis.
- Nach dem Absenden ist das Pfand eine gewöhnliche Position.
- Zahlung, Stornierung, Umbuchung, Direktverkauf und Export behandeln es wie heute.
- Ohne Verknüpfung bestellt die Servicekraft mit genau denselben Tipps wie heute.
- Das Pfand von Hand zu tippen bleibt möglich, etwa für ein unverknüpftes Produkt.
- Das Backend lässt nur Verknüpfungen mit demselben Steuersatz zu, weil Pfand den Satz der Hauptleistung teilt.

## User Stories

1. Als Servicekraft möchte ich, dass ein Getränk sein Glaspfand selbst mitbringt, damit ich es unter Zeitdruck nicht vergesse.
2. Als Servicekraft möchte ich das automatisch hinzugefügte Pfand sehen, bevor ich absende, damit ich dem Gast den Betrag nennen kann.
3. Als Servicekraft möchte ich das Pfand für einen Gast mit eigenem Glas zurücknehmen, damit er nichts doppelt zahlt.
4. Als Servicekraft möchte ich, dass beim Zurücknehmen eines Getränks sein Pfand mitgeht, damit kein verwaistes Pfand bleibt.
5. Als Servicekraft möchte ich Pfand weiter von Hand tippen, damit unverknüpfte Produkte und Sonderfälle gehen.
6. Als Servicekraft möchte ich ohne Verknüpfung genauso bestellen wie heute, damit sich für Vereine ohne Pfand nichts ändert.
7. Als Servicekraft im Direktverkauf möchte ich dasselbe Verhalten, damit Theke und Tisch gleich funktionieren.
8. Als Admin möchte ich je Variante ein Pfand wählen, damit das Glas Wein Pfand trägt, die Flasche nicht.
9. Als Admin möchte ich nur Pfand-Varianten mit passendem Steuersatz angeboten bekommen, damit kein Pfand zum falschen Satz verkauft wird.
10. Als Admin möchte ich in der Produktliste sehen, welche Varianten Pfand mitbringen, damit ich ohne Dialoge prüfe.
11. Als Admin möchte ich ein Pfand vor dem Fest deaktivieren können, ohne die Verknüpfungen zu verlieren.
12. Als Kassenwart möchte ich das automatisch gebuchte Pfand im Export unter Pfand sehen, damit die Steueraufteilung stimmt.
13. Als Serviceleitung möchte ich Pfand-Positionen wie heute stornieren und umbuchen, damit diese Abläufe gleich bleiben.

## Implementation Decisions

### Abgewogene Ansätze

| Ansatz                                                                       | Aufwand                                                                                        | Risiko                                                                                           | Umkehrbarkeit                                          |
| ---------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------ | ------------------------------------------------------ |
| A: Verknüpfung in den Stammdaten, der Warenkorb legt das Pfand sichtbar dazu | klein: eine Spalte, Variantendialog, Kopplung von Plus und Minus im Warenkorb                  | gering: Events, Bestell-API, Zahlung und Export bleiben unverändert                              | Feature entfernbar, die Spalte bleibt dann wirkungslos |
| B: Verknüpfung in den Stammdaten, das Backend hängt das Pfand beim Buchen an | mittel: Anreicherung von Bestellung und Direktverkauf, dazu ein Abwahlfeld für das eigene Glas | mittel: gebuchte Bestellung weicht von der Vorschau ab; die Servicekraft sieht das Pfand zu spät | wie A, aber die Bestell-API ändert ihre Bedeutung      |
| C: Pfandbetrag als Aufschlag am Preis der Variante, ohne eigene Position     | groß: neue Versionen der Verkaufs-Events, eigene Pfand-Zeilen im Export                        | hoch: bricht das Pfand-Kennzeichen je Position; die Rückgabe braucht ein nie verkauftes Pfand    | jede neue Event-Version bleibt für immer lesepflichtig |

Gewählt ist A.

- A erfüllt alle Stories und fasst kein Event, keine Projektion und keinen Kassenvorgang an.
- Das Pfand ist eine gewöhnliche Position; jeder nachgelagerte Vorgang behandelt es wie heute das getippte Pfand.
- Die Servicekraft sieht und korrigiert das Pfand vor dem Absenden, auch beim eigenen Glas.
- B versteckt das Pfand bis nach dem Buchen und braucht für das eigene Glas trotzdem ein neues Eingabefeld.
- C widerspricht D20 und kostet ein Vielfaches.

Die Verknüpfung hängt an der Variante, nicht am Produkt.

- Varianten eines Produkts brauchen oft verschiedenes Pfand: Glas Wein mit Glaspfand, Flasche Wein ohne.
- Die Einrichtung kostet je Variante eine Auswahl, einmal vor der Saison.
- Eine Produktebene darüber wäre eine zweite Regel mit Vorrangfrage.

### Stammdaten und Schema

- Eine Variante trägt optional einen Verweis auf eine Pfand-Variante; leer heißt kein Pfand.
- Der Verweis ist CRUD-Stammdatum und nie Teil eines Events.
- Eine additive Migration mit der nächsten freien Nummer ergänzt die Varianten um eine nullable Verweisspalte.
- Die Spalte ist ein Fremdschlüssel auf die Varianten; Varianten werden nur weich gelöscht.
- Bestehende Varianten bleiben ohne Pfand; die Migration schreibt keine Daten.
- Dieselbe Migration ersetzt die JSON-Darstellung der Variante und ergänzt sie um den Verweis.
- Mehrere Varianten dürfen dieselbe Pfand-Variante verknüpfen.

### Regeln der Verknüpfung

- Das Ziel ist eine nicht gelöschte Variante eines Pfand-Produkts.
- Die verknüpfende Variante gehört zu keinem Pfand-Produkt; Pfand verkettet sich nicht.
- Beide Produkte tragen denselben Steuersatz.
- Grund: Pfand für eine Warenumschließung teilt den Satz der Hauptleistung (BMF 20.10.2014).
- Glaspfand zu 19 % passt damit nur zu Getränken zu 19 %.
- Tellerpfand zu 7 % passt nur zu Speisen zu 7 %.
- Ein Kombi-Produkt verknüpft nur ein Kombi-Pfand; das folgt aus derselben Regel ohne Sonderfall.
- Eine inaktive Pfand-Variante darf verknüpft werden, damit der Admin vor dem Fest einrichtet.
- Die Domäne prüft die Regeln als reine Funktion über Variante, ihr Produkt und das Zielprodukt.
- Das Backend prüft bei jedem Schreiben, das eine Regel brechen kann.
- Das sind: Variante anlegen oder ändern mit Verweis sowie Produkt ändern.
- Beim Produkt prüft es jede Verknüpfung, die von seinen Varianten ausgeht oder auf sie zeigt.
- So scheitert ein Steuersatzwechsel, der eine bestehende Verknüpfung bräche.
- Ebenso scheitert das Abschalten des Pfand-Kennzeichens, solange nicht gelöschte Varianten auf das Produkt zeigen.
- Ebenso scheitert das Einschalten, solange Varianten des Produkts selbst ein Pfand verknüpfen.
- Der Katalog eines Vereins umfasst Dutzende Varianten; die Prüfung liest ihn deshalb vollständig.
- Eine Verletzung beantwortet das Backend mit 400 und einem eigenen Fehlercode je Regel.
- Das Frontend zeigt dazu einen deutschen Meldungstext, etwa „Pfand und Variante brauchen denselben Steuersatz".

### Löschen und Deaktivieren

- Löscht der Admin eine Pfand-Variante, entfernt dieselbe Transaktion alle Verweise auf sie.
- Löscht er ein Pfand-Produkt, entfernt dieselbe Transaktion alle Verweise auf dessen Varianten.
- Zeigen Varianten auf das Gelöschte, nennt der Lösch-Dialog zusätzlich: „Verknüpfte Varianten verlieren ihr Pfand."
- Eine Verweigerung des Löschens schiede aus: Der Admin müsste jede verknüpfte Variante einzeln öffnen.
- Deaktiviert der Admin eine Pfand-Variante, bleibt jeder Verweis erhalten.
- Nach dem Reaktivieren wirkt die Verknüpfung wieder ohne neue Einrichtung.
- Löscht der Admin die verknüpfende Variante, bleibt ihr Verweis am gelöschten Datensatz wirkungslos.

### API

- Variante anlegen nimmt optional die ID der Pfand-Variante an.
- Variante ändern überträgt den Verweis immer mit; leer entfernt ihn.
- Beide Produkt-Lesepfade liefern je Variante die ID der Pfand-Variante oder null.
- Der Admin-Lesepfad liefert den gespeicherten Verweis, auch auf ein inaktives Ziel.
- Der Service-Lesepfad liefert ihn nur, wenn das Ziel in derselben Antwort aktiv erscheint.
- So filtert das Backend, und das Frontend legt nur bestellbares Pfand dazu.
- Bestellung, Direktverkauf, Zahlung, Stornierung, Korrektur, Umbuchung und Pfandrückgabe bleiben unverändert.
- Frontend und Backend ändern die Stammdaten-API im selben Release.

### Bestellen und Direktverkauf

- Der Warenkorb bleibt je Variante geschlüsselt.
- Eine reine Funktion koppelt Plus und Minus der Produktliste an die verknüpfte Pfand-Variante.
- Plus auf eine verknüpfte Variante erhöht sie und ihr Pfand um je eins.
- Minus senkt sie um eins und ihr Pfand um eins, solange das Pfand über null liegt.
- Plus und Minus am Pfand selbst ändern nur das Pfand.
- Damit bleiben das Tippen von Hand und das Abziehen für ein eigenes Glas erhalten.
- Mehrere Varianten mit demselben Pfand summieren sich in dessen einer Warenkorbzeile.
- Die Zeile einer verknüpften Variante zeigt unter dem Preis den Pfandbetrag, etwa „+ 2,00 € Pfand".
- Abschluss und Vorschau zeigen das Pfand als eigene Position, wie heute das getippte.
- Tisch-Bestellung und Direktverkauf nutzen dieselbe Kopplung.
- Die Auswahl der Pfandrückgabe nutzt sie nicht; Pfand-Varianten verknüpfen ohnehin nichts.
- Der Zustand lebt im bestehenden Vorgangs-Hook; es gibt keinen globalen Store.
- Bekannte Unschärfe: Nach dem Abzug für ein eigenes Glas zieht ein weiteres Minus am Getränk noch ein Pfand ab.
- Welches Getränk zum eigenen Glas gehört, weiß die Kasse nicht; die Vorschau zeigt das Ergebnis vor dem Absenden.

### Admin-Oberfläche

- Der Variantendialog bekommt beim Anlegen und Bearbeiten ein Auswahlfeld „Pfand".
- Die Auswahl bietet „Kein Pfand" und jede Pfand-Variante mit gleichem Steuersatz an.
- Jede Option nennt Produkt, Variante und Preis, etwa „Glaspfand · 2,00 €".
- Inaktive Pfand-Varianten tragen in der Auswahl den Zusatz „(inaktiv)".
- Das Feld erscheint nur, wenn das Produkt kein Pfand-Produkt ist und eine passende Pfand-Variante existiert.
- Ohne Pfand-Produkte bleibt der Dialog damit wie heute.
- Der Varianten-Chip in der Produktliste zeigt eine verknüpfte Variante mit einem Pfand-Hinweis.
- Scheitert eine Produktänderung an einer Verknüpfung, zeigt der Produktdialog die Meldung des Backends.

### Nachgelagerte Vorgänge

- Die Zahlung liest das Pfand-Kennzeichen beim Kassieren aus dem Produkt, wie D20 es festlegt.
- Der Export bucht die Position deshalb unter dem Geschäftsvorfalltyp Pfand.
- Eine Stornierung des Getränks storniert sein Pfand nicht mit; die Serviceleitung wählt beide wie heute.
- Die Umbuchung verschiebt Pfand-Positionen wie jede Position.
- Die Zahlungsauswahl wählt das Pfand nicht mit; „Alle auswählen" deckt den häufigsten Fall ab.
- Der Arbeitsbon am Tisch druckt die Pfand-Position an der Station ihrer Kategorie, wie heute.
- Wer keinen Pfand-Arbeitsbon will, legt das Pfand-Produkt in eine Kategorie ohne Druckstation.
- Kein Event-Format, keine Event-Version, keine Projektion und kein Replay ändert sich.
- Die Produktstatistik zählt die Pfand-Variante wie heute.

## Testing Decisions

Getestet wird nur äußeres Verhalten an öffentlichen Schnittstellen; erwartete Werte sind Literale.

- Domäne: die Regel-Funktion als Tabelle. Fälle: gültig, Ziel ohne Pfand-Kennzeichen, Ziel gelöscht, Quelle ist Pfand-Produkt, Satz 19 % gegen 7 %, Kombi gegen Kombi gültig. Vorbild ist `variante_test.go`.
- Repository gegen Postgres: Löschen einer Pfand-Variante und eines Pfand-Produkts entfernt die Verweise. Vorbild ist `delete_integration_test.go`.
- Repository gegen Postgres: Der Service-Lesepfad liefert keinen Verweis auf eine inaktive Pfand-Variante, der Admin-Lesepfad schon.
- Anwendung: Steuersatzwechsel und Pfand-Kennzeichen aus scheitern bei bestehender Verknüpfung mit ihrem Fehlercode. Vorbild ist `command_test.go`.
- Fehlercodes: keine eigenen; der bestehende Vertragstest der Fehlerabbildung erzwingt einen Code je neuem Fehler.
- Migration: keine eigenen; Frischinstallation und Upgrade-Pfad decken die bestehenden Integrations- und CI-Läufe ab.
- Frontend: die Kopplungsfunktion als Tabelle. Fälle: Plus, Minus, Boden bei null, zwei Varianten mit einem Pfand, Pfand von Hand. Vorbild ist `use-mengen.test.ts`.
- Bestellabschluss: Plus auf eine verknüpfte Variante sendet Variante und Pfand im Request. Vorbild ist `BestellungTab.test.tsx`.
- Direktverkauf: derselbe Fall in `DirektverkaufTab.test.tsx`.
- Admin: Das Pfand-Feld fehlt ohne passende Pfand-Variante und bietet nur den gleichen Steuersatz an. Vorbild ist `Products.test.tsx`.
- Event-Verträge, Zahlung, Stornierung, Umbuchung, TSE und Export: keine, abgedeckt durch die bestehenden Tests, die unverändert grün bleiben.
- End-to-End: keine, Domänen-, Integrations- und Komponententests erreichen jeden Pfad.

## Out of Scope

- Pfand automatisch in der Zahlungsauswahl, Stornierung oder Umbuchung mitwählen.
- Eine Zuordnung, welches Pfand zu welchem Stück gehört.
- Mehr als ein Pfand je Variante, etwa Glas und Teller zu einem Menü.
- Eine Verknüpfung auf Produktebene oder für alle Varianten auf einmal.
- Pfand-Positionen vom Arbeitsbon am Tisch fernhalten; die Bestellung trägt kein Pfand-Kennzeichen.
- Ein Steuersatz, den das Pfand von der verknüpften Variante übernimmt; es behält den seines Produkts.
- Änderungen an Events, Pfandrückgabe, Verrechnung oder Export.

## Further Notes

Vorgeschlagener Eintrag für den Abschnitt „Pfand" in `docs/language.md`:

| Begriff           | Bedeutung                                                                                                                    | Zu vermeiden                 |
| ----------------- | ---------------------------------------------------------------------------------------------------------------------------- | ---------------------------- |
| Pfand-Variante    | Variante eines Pfand-Produkts.                                                                                               | Pfandartikel                 |
| Pfand-Verknüpfung | Stammdatenverweis einer Variante auf eine Pfand-Variante gleichen Steuersatzes; die Bestellung legt je Stück ein Pfand dazu. | Pfandzuordnung, Pfandbindung |

Der Eintrag kommt mit der Umsetzung, zusammen mit seinem Code-Mapping.
Bei Umsetzung ändern sich drei Stellen:

- der Einleitungssatz zu Pfand im Glossar;
- das Produkt-Aggregat im Handbuch;
- der Pfand-Absatz im Leitfaden „Belege und Steuersätze".

Die Wahl von Ansatz A gehört nach der Freigabe als Zeile in `docs/decisions.md`.

Offene Entscheidungen für den Owner:

1. Bauen ohne Feldnachfrage: Kein Verein hat Auto-Pfand gemeldet. Empfohlen ist die Freigabe. Ohne Verknüpfung ändert sich nichts. Vereine mit Pfand sparen je Getränk einen Tipp, und kein Pfand geht mehr vergessen. Rückfall: warten, bis ein Verein danach fragt.
2. Abholbon im Direktverkauf: Pfand druckt heute eigene Abholbons, „pro Position" einen je Verkauf, „pro Stück" einen je Glas. Mit Auto-Pfand trifft das jeden Getränkeverkauf. Empfohlen: Der Abholbon lässt Pfand-Positionen in jedem Bonmodus weg, im selben Release. Der Direktverkauf trägt das Pfand-Kennzeichen je Position bereits im Event. Rückfall: Abholbon unverändert lassen.
3. Kombi-Produkte: Die Regel „gleicher Steuersatz" erlaubt einem Kombi-Produkt nur ein Kombi-Pfand. Das BMF-Schreiben vom 20.10.2014 regelt Pfand zu einem 70/30-pauschalierten Menü nicht. Empfohlen: die Regel ohne Sonderfall lassen und keinen Kombi-Rat im Leitfaden geben. Rückfall: Kombi-Produkte von der Verknüpfung ausschließen.
