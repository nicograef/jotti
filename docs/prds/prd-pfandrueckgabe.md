# PRD: Pfandrückgabe und Verrechnung

## Problem Statement

Drei Vereine geben Getränke in Pfandgläsern aus. Das Pfand berechnen sie mit einem Pfand-Produkt zum positiven Preis, das funktioniert.

Gibt ein Gast das Glas zurück, kennt jotti keinen Vorgang dafür. Zahlt die Servicekraft Pfand aus, bleibt die Auszahlung unaufgezeichnet.

Der Soll-Kassenbestand liegt am Abend um alle ausgezahlten Pfandbeträge zu hoch. Der Kassensturz zeigt eine Differenz, die keine ist.

Steuerlich fehlt eine Entgeltminderung. Der Umsatz je Steuersatz enthält das Pfand, die Rückzahlung mindert ihn nirgends.

Ein Gast gibt oft zwei Gläser zurück und bestellt zwei Bier. Die Servicekraft will die Differenz kassieren, nicht erst auszahlen und dann einnehmen.

Ein Negativpreis hilft nicht. Kein Preis darf negativ sein, weder im Admin-Formular noch in gespeicherten Events.

Ein Verein fragt zusätzlich nach Rabatt und Gutschrift am Tisch. Diesen Wunsch klärt das Kapitel „Out of Scope".

## Solution

Der Admin kennzeichnet ein Produkt als Pfand-Produkt. Es verkauft sich wie jedes Produkt und berechnet das Pfand als eigene Position.

Der Kassieren-Drawer am Tisch und der Abschluss im Direktverkauf bekommen einen Abschnitt „Pfand zurück". Er listet die Varianten aller Pfand-Produkte mit Mengen-Stepper.

Die Servicekraft tippt die zurückgegebenen Gläser an. Der Drawer zeigt den Betrag netto: Der Gast zahlt die Differenz oder erhält sie bar.

Ohne Verkauf geht dasselbe: Der Abschnitt „Pfand zurück" allein ergibt eine reine Pfandrückgabe mit Barauszahlung.

Das gilt auch am Tisch, an dem in der Kassensitzung noch niemand bestellt hat. Die Rückgabe bucht dann in einen eigenen Stream wie im Direktverkauf.

Jede Pfandrückgabe ist ein eigener, TSE-signierter Kassenvorgang. Sie mindert den Soll-Kassenbestand und den Umsatz am Steuersatz des Pfand-Produkts.

Die Historie zeigt jede Pfandrückgabe. Die Serviceleitung storniert eine irrtümliche Rückgabe dort, mit Kommentar.

Der Pfandbeleg ist auf Verlangen druckbar, wie jeder Kassenbeleg. Der DSFinV-K-Export führt die Rückgabe als Geschäftsvorfall „PfandRueckzahlung".

## User Stories

1. Als Servicekraft im Direktverkauf möchte ich zurückgegebene Gläser erfassen und das Pfand bar auszahlen, damit die Kassenlade zur Aufzeichnung passt.
2. Als Servicekraft im Direktverkauf möchte ich in einem Abschluss verkaufen und Gläser zurücknehmen, damit der Tausch ein Schritt bleibt.
3. Als Servicekraft am Tisch möchte ich beim Kassieren zurückgegebene Gläser verrechnen, damit der Gast nur die Differenz zahlt.
4. Als Servicekraft am Tisch möchte ich Pfand auch ohne Zahlung auszahlen, damit ein Gast ohne weitere Bestellung sein Geld bekommt.
5. Als Servicekraft möchte ich Pfand an einem Tisch ohne Bestellung auszahlen, damit kein Gast von der Theke abgewiesen wird.
6. Als Servicekraft möchte ich Nettobetrag und Rückgeld fertig berechnet sehen, damit ich unter Stress nicht im Kopf rechne.
7. Als Servicekraft möchte ich mehr Pfand auszahlen dürfen, als der Gast gerade kauft, damit ein Rückgabestapel nicht scheitert.
8. Als Servicekraft möchte ich nur Pfand-Varianten in der Rückgabe sehen, damit ich kein Bier statt eines Glases zurückbuche.
9. Als Servicekraft möchte ich einen Pfandbeleg auf Verlangen drucken, damit der Verein die Belegpflicht erfüllt.
10. Als Admin möchte ich ein Produkt als Pfand-Produkt kennzeichnen, damit die Rückgabe genau dessen Varianten anbietet.
11. Als Admin möchte ich dem Pfand-Produkt den Steuersatz der Hauptleistung geben, damit Pfand und Rückgabe am richtigen Satz landen.
12. Als Serviceleitung möchte ich eine irrtümliche Pfandrückgabe mit Kommentar stornieren, damit Kassenbestand und Aufzeichnung wieder stimmen.
13. Als Serviceleitung möchte ich stornierte Pfandrückgaben in der Historie erkennen, damit niemand dieselbe Rückgabe zweimal korrigiert.
14. Als Kassenwart möchte ich die Pfandrückgaben im Soll-Kassenbestand sehen, damit der Kassensturz ohne Scheindifferenz aufgeht.
15. Als Kassenwart möchte ich den Umsatz je Steuersatz abzüglich der Pfandrückgaben sehen, damit die Umsatzsteuer stimmt.
16. Als Kassenwart möchte ich in der Abrechnung pro Servicekraft das ausgezahlte Pfand abgezogen sehen, damit ich weiß, was jede abgibt.
17. Als Kassenwart möchte ich Anzahl und Summe der Pfandrückgaben im Kassenabschluss sehen, damit ich den Abend nachvollziehe.
18. Als Kassenwart möchte ich einen DSFinV-K-Export mit Pfandrückgaben als „PfandRueckzahlung", damit eine Kassen-Nachschau ohne Rückfragen läuft.
19. Als Servicekraft möchte ich, dass ein doppelt gesendeter Abschluss nur einmal bucht, damit ein Funkloch kein Pfand verdoppelt.

## Implementation Decisions

### Weg: Rückgabe-Vorgang

| Weg                      | Aufwand | Risiko                                                                                                                                                                                                          | Umkehrbarkeit                                                                                  |
| ------------------------ | ------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| Negativ-Variante         | gering  | hoch: Die Preisgrenze prüft auch gespeicherte Events, die DB verbietet negative Preise und Salden, die Steueraufteilung lehnt negative Beträge ab, die Rückgabe läuft als Bestellung über Arbeitsbons und Saldo | schlecht: Negative Preise in v1-Events bleiben für immer, jeder künftige Leser muss sie tragen |
| Preisfindung je Position | hoch    | mittel: Folgeversionen von Bestellung, Zahlung, Direktverkauf und Stornierung; eine Rückgabe ohne Kauf hat keine Position zum Mindern; die Preisfindung kennt laut DSFinV-K nur Grundpreis, Rabatt und Zuschlag | mittel: Folgeversionen bleiben im Journal                                                      |
| Rückgabe-Vorgang         | mittel  | gering: Bestehende Events, Schemas und die Positiv-Regel der Geldsummen bleiben unberührt; das Vorzeichen trägt der Event-Typ wie bei der Warenrücknahme                                                        | gut: rein additive Event-Typen, eine additive Spalte und zwei Indexe                           |

Gewählt ist der Rückgabe-Vorgang. Er deckt Rückgabe mit und ohne Kauf ab und bricht keinen eingefrorenen Vertrag.

### Stammdaten

- Das Produkt bekommt ein Pfand-Kennzeichen, Standard „nein". Eine additive Migration fügt die Spalte hinzu.
- Das Kennzeichen steuert nur, welche Varianten die Rückgabe anbietet. Bestellung, Arbeitsbon und Kategorie bleiben unverändert.
- Der Steuersatz des Pfand-Produkts ist der Satz der Rückgabe. Die fette Position friert ihn mit Name und Preis ein.
- Admin-Formular und Produkt-Antwort tragen das Kennzeichen; zog und Zod prüfen es auf beiden Seiten.

### Event-Verträge

- `pfand-zurueckgegeben:v1` trägt eine client-erzeugte Rückgabe-ID, die Tisch-ID, fette Positionen, die Gesamtsumme und einen optionalen Kommentar.
- Die Tisch-ID ist im Direktverkauf 0, wie beim Direktverkauf-Storno im Reporting.
- Positionen und Summe sind positiv wie bei jeder Warenrücknahme. Die Bedeutung „Auszahlung" folgt aus dem Event-Typ.
- `pfandrueckgabe-storniert:v1` trägt Stornierungs-ID, Rückgabe-ID, die stornierten Positionen, die Summe und einen Pflichtkommentar.
- Die Stornierung darf je Position höchstens die zurückgegebene Menge zurücknehmen, wie beim Direktverkauf-Storno.
- Beide Typen bekommen eingefrorene Vertragstests. Bestehende Event-Typen ändern sich nicht.

### Streams und Projektion

- Am Tisch landet die Rückgabe im Stream der Tisch-Session. Die Projektion übergeht sie, Saldo und Positionen bleiben gleich.
- Im Direktverkauf bekommt jede Rückgabe einen eigenen Stream unter der Kassensitzung, wie ein Direktverkauf.
- Am Tisch ohne Tisch-Session bekommt die Rückgabe denselben eigenen Stream, mit der Tisch-ID im Event.
- Grund: Die Tisch-Session entsteht mit der ersten Bestellung. Eine Rückgabe als erstes Event hinterließe einen Stream ohne Projektionszeile.
- Ablehnen scheidet aus: Die Servicekraft müsste in den Direktverkauf wechseln, dessen Auszahlung der Theke statt ihr zugerechnet würde.
- Diese Rückgabe erscheint in der Direktverkauf-Historie mit Tischname. Dort storniert die Serviceleitung sie.
- Eine Rückgabe erzeugt kein Guthaben am Tisch. Verrechnet wird nur im selben Abschluss, so bleibt der Saldo nie negativ.
- Die Stornierung hängt sich als Folge-Version an den Stream der Rückgabe.

### Commands und API

- „Zahlung kassieren" und „Direktverkauf tätigen" nehmen optional Pfand-Positionen an. Beide Events entstehen in einer Transaktion.
- Ein eigener Command „Pfand zurückgeben" deckt die Rückgabe ohne Kauf ab, am Tisch wie im Direktverkauf.
- Die Rückgabe-ID macht jeden Command mit Pfand idempotent. Ein zweiter Aufruf mit derselben ID bucht nichts.
- Ein partieller Unique-Index auf die Rückgabe-ID für `pfand-zurueckgegeben:v1` erzwingt das, nach dem Vorbild `idx_kassenjournal_verkauf_id`.
- Ein zweiter partieller Unique-Index auf die Stornierungs-ID für `pfandrueckgabe-storniert:v1` macht die Stornierung idempotent.
- Beide Indexe legt die additive Migration der Pfand-Spalte an. Verletzt ein Wiederholversuch den Index im verrechneten Abschluss, rollt auch die Zahlung zurück.
- Jede Servicekraft darf Pfand zurückgeben. Stornieren dürfen nur Serviceleitung und Admin.
- Alle Endpunkte sind POST. Die Abschluss-Barriere der Kassensitzung gilt wie für jeden Kassenvorgang.
- Das Frontend ruft die Endpunkte über die bestehenden Backend-Klassen. Der Pfand-Abschnitt lebt in der einen Abschluss-Komponente je Fläche.

### TSE

- Jede Pfandrückgabe ist ein eigener Geschäftsvorfall und bekommt eine eigene TSE-Transaktion.
- Grund: Die Auszahlung ändert die Vermögenszusammensetzung (AEAO zu § 146a Nr. 1.10.1). § 2 Satz 1 KassenSichV verlangt dafür eine Transaktion.
- Prozesstyp ist `Kassenbeleg-V1` mit negativen Bruttobeträgen am Satz und negativer Barzahlung, wie bei der Warenrücknahme.
- Die Stornierung einer Rückgabe ist ebenfalls `Kassenbeleg-V1`, mit positiven Beträgen.
- Ein verrechneter Abschluss erzeugt zwei Transaktionen und zwei Belege: Zahlung oder Verkauf und Pfandrückgabe.
- Die Signatur läuft über den bestehenden Signaturauftrag, samt Ausfall-Pfad.

### DSFinV-K

- Die Rückgabe ist ein Bon mit Typ „Beleg". Ihre Zeilen tragen den Geschäftsvorfalltyp „PfandRueckzahlung" (DSFinV-K Anhang C, S. 50).
- Die Zeilen tragen negative Menge und negative Beträge je USt-Schlüssel (Tz. 4.2.5). Das Storno-Kennzeichen bleibt 0.
- Eine Rückgabe verweist auf keinen Ursprungsbon. Pfandrückgaben gelten unabhängig vom Verkauf als Entgeltminderung (Abschn. 10.1 Abs. 8 Satz 3 UStAE).
- Die Stornierung einer Rückgabe verweist per Referenz auf den Rückgabe-Bon.
- Der Abrechnungskreis ist der Tisch, auch ohne Tisch-Session, im Direktverkauf leer.
- Die Geschäftsvorfall-Summen des Kassenabschlusses führen „PfandRueckzahlung" je USt-Schlüssel.
- Der Export setzt den Geschäftsvorfalltyp je Bon. Ein Bon trägt hier stets nur einen Typ.
- Preisfindung und Zusatzinfo bleiben reine Kopfzeilen. Die Preisfindung gilt nur für Rabatt und Zuschlag im Positionspreis (Tz. 3.1.1.2).
- Die Zusatzinfo dient laut Spezifikation nur der Erläuterung von Warenzusammenstellungen (Tz. 3.1.1.3, Kap. 18). Pfand nennt sie nicht.

### Umsatzsteuer

- Pfand für eine Warenumschließung teilt das Schicksal der Hauptleistung (Abschn. 3.10 Abs. 5a Satz 2 UStAE).
- Das Pfand ist Teil des Entgelts für die Lieferung (Abschn. 10.1 Abs. 8 Satz 1 UStAE).
- Gläserpfand zu Getränken trägt damit 19 %, Tellerpfand zu Speisen 7 % (§ 12 Abs. 2 Nr. 15 UStG).
- Die Rückzahlung ist eine Entgeltminderung (Abschn. 3.10 Abs. 5a Satz 3 und Abschn. 10.1 Abs. 8 Satz 2 UStAE).
- Den Steuerbetrag berichtigt der Verein im Zeitraum der Rückgabe (§ 17 Abs. 1 Satz 1 und 8 UStG).
- Elektronische Kassen ordnen Entgeltminderungen direkt dem Steuersatz zu (DSFinV-K Tz. 4.2.4). jotti nimmt dafür den Satz des Pfand-Produkts.
- Dieselbe Zuordnung verlangt Abschn. 10.1 Abs. 8 Satz 4 UStAE: sachgerecht nach den geltenden Steuersätzen.

### Kasse und Reporting

- Der Soll-Kassenbestand zieht Pfandrückgaben ab und addiert deren Stornierungen. Die Bareinnahmen der Vier-Komponenten-Aufschlüsselung tun dasselbe.
- Der Umsatz je Steuersatz zieht Pfandrückgaben am Satz ab. Eine additive Migration ersetzt dafür die SQL-Extraktionsfunktion und ergänzt eine für Pfand.
- Die Abrechnung pro Servicekraft zieht Rückgaben mit Tisch-ID beim Akteur ab, auch die ohne Tisch-Session. Rückgaben im Direktverkauf zählen zur Kasse des Direktverkaufs.
- Ein abzugebender Betrag darf dadurch negativ werden. Die Servicekraft hat dann mehr Pfand ausgezahlt als eingenommen.
- Eine Stornierung zählt für die Servicekraft der ursprünglichen Rückgabe, nach der bestehenden Storno-Zuordnung.
- Der abzugebende Betrag der eigenen Übersicht und des Live-Dashboards folgt derselben Rechnung.
- Kassenabschluss und Tagesabschluss führen Anzahl und Summe der Pfandrückgaben.
- Die Produktstatistik bleibt auf der Bestellbasis. Rückgaben ändern sie nicht, wie Warenrücknahmen.

### Beleg

- Der Pfandbeleg ist auf Verlangen druckbar und nachdruckbar, über die Rückgabe-ID. Er trägt negative Beträge und die Steuermatrix.
- Die Belegausgabe-Befreiung des Vereins gilt auch für ihn. Erstellbar bleibt er jederzeit (§ 146a Abs. 2 AO).

## Testing Decisions

Getestet wird äußeres Verhalten an der billigsten Ebene, die den Fehler sieht. Typ- und Schema-Form prüft kein Test.

- Positive Geldsummen: Die bestehende Tabelle der Positiv-Regel bekommt die beiden Event-Typen. Sie ist eine Repo-Regel als Test.
- Event-Verträge: Je ein eingefrorener Vertragstest pro Typ. Der Vollständigkeitstest der Vertragsdatei erzwingt ihn ohnehin.
- Domäne: Die Stornierung lehnt Mengen über der Rückgabe ab. Vorbild ist der Validierungstest des Direktverkauf-Stornos.
- Anwendung mit echter Datenbank: Ein doppelter Aufruf mit derselben Rückgabe-ID bucht einmal, ebenso mit derselben Stornierungs-ID. Vorbild ist der Idempotenz-Test des Direktverkaufs.
- Anwendung mit echter Datenbank: Eine Rückgabe am Tisch ohne Tisch-Session legt keine Projektionszeile an.
- Anwendung: Eine ungültige Pfand-Position verhindert auch die Zahlung, beide Events fehlen dann.
- Anwendung: Eine Rückgabe erzeugt einen Signaturauftrag mit negativem `Kassenbeleg-V1`. Den Baustein deckt der Storno-Fall der processData-Tests.
- SQL mit echter Datenbank: Soll-Kassenbestand und Umsatz je Steuersatz ziehen eine Rückgabe ab. Vorbild sind die Reporting-Integrationstests.
- Abrechnung pro Servicekraft: Eine Rückgabe am Tisch mindert den Betrag des Akteurs, mit und ohne Tisch-Session; eine im Direktverkauf nicht.
- DSFinV-K: Ein Archiv mit Rückgabe und Stornierung hat die Zeilen „PfandRueckzahlung" und keinen Prüfbefund. Vorbild ist der befundfreie Archivtest der DSFinV-K-Prüfung.
- Migration: Der Upgrade-Pfad in der CI deckt die zusätzliche Spalte und die ersetzte Funktion ab.
- Frontend: Der Abschluss zeigt bei Verrechnung Nettobetrag und Rückgeld. Vorbild sind die Tests der Abschluss-Komponenten von Tisch und Direktverkauf.
- Ende zu Ende: Bier mit Pfand verkaufen, Glas zurückgeben, Soll-Kassenbestand prüfen. Vorbild ist die mobile Spec zum Direktverkauf-Storno.

## Out of Scope

- **Rabatt.** Dieses PRD baut ihn nicht; ein einziger Verein fragt danach. Jede Entgeltminderung muss elektronisch direkt dem Steuersatz zugeordnet werden (DSFinV-K Tz. 4.2.4).
- Ein Rabatt auf die Zwischensumme über 7 % und 19 % bräuchte eine Aufteilung je Satz und Folgeversionen der Zahlungs-Events.
- Freiwillige unter Stress, die Nachlässe gewähren, brauchen Rechte und Grenzen. Das widerspricht der Produkt-Konservativität.
- **Gutschrift.** Ein reklamiertes Getränk deckt die Stornierung ab: Korrektur vor der Zahlung, Warenrücknahme danach.
- Ein Freigetränk für Helfer hat kein Entgelt, das sich mindern ließe (Abschn. 10.3 Abs. 1 UStAE). Der Helferdeckel ist bereits Nicht-Ziel.
- Kommt Rabatt zurück, ist die gesonderte negative Positionszeile mit Typ „Rabatt" der Weg (Tz. 4.2.4), nicht die Preisfindung.
- Guthaben am Tisch: Eine Rückgabe ohne Abschluss bleibt nicht als Gutschrift am Tisch stehen.
- Automatisches Pfand zur Getränke-Variante. Die Servicekraft tippt das Pfand-Produkt wie jede Position.
- Pfandsaldo oder Glasbestand: Wie viele Gläser fehlen, zählt der Verein selbst.
- Pfandmarken oder Pfandbons als Rückgabe-Nachweis. Papier und Vertrauen decken das ab.
- Geschäftsvorfalltyp „Pfand" für den Pfand-Verkauf. Der Verkauf bleibt „Umsatz", siehe Further Notes.
- Ein gemeinsamer Bon für Verkauf und Rückgabe. Anhang I der DSFinV-K erlaubt ihn, er bräuchte aber Folgeversionen der Zahlungs-Events.

## Further Notes

### Vorgeschlagene Einträge für die Ubiquitous Language

| Begriff                   | Definition                                                                                                                | Zu vermeiden                             |
| ------------------------- | ------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------- |
| Pfand-Produkt             | Produkt mit Pfand-Kennzeichen, dessen Varianten die Pfandrückgabe anbietet und das sich sonst wie jedes Produkt verkauft. | Pfandartikel, Leergut                    |
| Pfandrückgabe             | Kassenwirksamer Vorgang, in dem ein Gast Pfandgut zurückgibt und den Pfandbetrag bar oder verrechnet erhält.              | Pfandrücknahme, Pfand-Storno, Gutschrift |
| Verrechnung               | Pfandrückgabe im selben Abschluss wie eine Zahlung oder ein Direktverkauf; bar fließt nur die Differenz.                  | Guthaben, Gutschrift, Pfandabzug         |
| Pfandrückgabe-Stornierung | Positionsgenaue Korrektur einer irrtümlichen Pfandrückgabe durch Serviceleitung oder Admin, mit Pflichtkommentar.         | Rückgabe-Storno, Pfand-Korrektur         |

„Gutschrift" bleibt aus Code und Oberfläche fern. Im Umsatzsteuerrecht ist sie eine Rechnung, die der Leistungsempfänger ausstellt (§ 14 Abs. 2 Satz 2 UStG).

### Pfand-Verkauf bleibt „Umsatz"

Den Pfand-Verkauf als „Pfand" zu führen, bräuchte ein Pfand-Merkmal in der Position. Das ginge nur über Folgeversionen aller Verkaufs-Events.

Ein Merkmal aus den Stammdaten zur Exportzeit würde alte Events umdeuten. Die Freeze-Regeln verbieten das.

Steuerlich ist „Umsatz" korrekt: gleicher Satz, gleiches Entgelt (Abschn. 10.1 Abs. 8 Satz 1 UStAE). Asymmetrisch bleibt nur die Buchungsgliederung.

### Prüfpunkte vor der Umsetzung

- Ob Festgläser Warenumschließungen oder Transporthilfsmittel sind, regelt das BMF-Schreiben vom 20.10.2014 (I. Nr. 1).
- Das Schreiben nennt Flaschen, keine Gläser oder Teller. Die Einordnung als Warenumschließung ist eine Folgerung aus seinen Kriterien, kein Zitat.
- Die DSFinV-K zitiert § 17 Abs. 1 Satz 7 UStG. Nach dem vorliegenden Gesetzestext steht die Zeitraumregel in Satz 8.
- Die Dateitabelle des DSFinV-K-Exports in der Compliance-Doku nennt Pfand als Beispiel für die Zusatzinfo. Die Spezifikation stützt das nicht.

### Bei der Umsetzung nachzuziehen

- Handbuch: Event-Tabelle, Subject-Design, Soll-Formel und die Invariante zum abzugebenden Betrag.
- Compliance-Doku: Prozesstyp-Tabelle, Geschäftsvorfalltypen, Storno-Handling und die Zeile zur Zusatzinfo.
- Leitfaden „Belege und Steuersätze": welcher Satz für Gläser- und Tellerpfand gilt.
- Anforderungen: eine Kennung für die Pfandrückgabe. Glossar: die vier Einträge oben.
