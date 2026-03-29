# jotti — Die vollständige User Journey und Architekturentscheidungen

> Dieser Text beschreibt die gesamte User Journey von jotti — von der ersten Einrichtung über den laufenden Betrieb bis zum Tagesabschluss — und erklärt die technischen Entscheidungen hinter der Architektur: warum Event-Sourcing eingesetzt wird, wie es funktioniert, warum synchrone Projektionen notwendig sind und warum das Datenmodell so aufgebaut ist, wie es ist.

---

## 1. Was ist jotti?

jotti ist ein kostenloses, quelloffenes mobiles Kassensystem für Vereine. Es richtet sich an eingetragene Vereine, gemeinnützige Gesellschaften und kirchliche Träger, die ein- bis dreimal im Jahr Veranstaltungen mit temporärer Gastronomie organisieren — Vereinsfeste, Weihnachtsmärkte, Konzerte, Maihocks. Das Besondere an jotti ist, dass es ohne spezielle Hardware auskommt: Die Servicekräfte nutzen ihre eigenen Smartphones im Browser. Es gibt keine App, die installiert werden muss, keine teuren Kassenhardware, keine monatlichen Abo-Gebühren. Das System läuft vollständig selbst gehostet per Docker auf einem einfachen virtuellen Server.

Das Ziel ist radikale Einfachheit: ein Kassensystem, das ein ehrenamtlicher Vorstand in wenigen Minuten einrichten kann und das eine Servicekraft ohne Schulung sofort versteht.

---

## 2. Erste Einrichtung

Die Einrichtung beginnt damit, dass jemand — typischerweise der Schatzmeister oder ein technisch versiertes Vorstandsmitglied — jotti per Docker Compose auf einem Server startet. Das System erzeugt beim ersten Start automatisch eine unveränderliche Seriennummer (UUID), die die Instanz eindeutig identifiziert. Diese Seriennummer wird später für die Kassenmeldung beim Finanzamt benötigt.

Nach dem Start ist genau ein Benutzer vorhanden: der initiale Admin. Dieser Admin hat eine Einmal-PIN bekommen, die er beim ersten Login eingeben muss. Nach dem Login leitet das System ihn automatisch auf die Seite zur Passwortänderung weiter. Erst nachdem er ein eigenes Passwort gesetzt hat, kann er das System vollständig nutzen. Dieses Muster — Einmal-PIN bei der Erstellung, Passwortänderung beim ersten Login — gilt für jeden neu erstellten Benutzer in jotti.

Das Einmal-Passwort ist sechsstellig, wird bei der Benutzererstellung generiert und mit Argon2id gehasht gespeichert. Argon2id ist der Stand der Technik für Passwort-Hashing: speicherintensiv (64 MB), zeitintensiv (2 Iterationen) und resistent gegen Brute-Force- sowie Side-Channel-Angriffe. Das fertige Passwort des Benutzers wird ebenfalls mit Argon2id gehasht. In der Datenbank liegen also immer nur Hashes — niemals Klartextpasswörter.

---

## 3. Admin-Konfiguration: Stammdaten

Bevor die erste Veranstaltung beginnen kann, richtet der Admin die Stammdaten ein. Es gibt drei Kategorien: Produkte, Tische und Benutzer.

### 3.1 Produkte

Ein Produkt ist ein Verkaufsartikel wie „Bier", „Bratwurst" oder „Cappuccino". Jedes Produkt gehört einer Kategorie an: Essen, Getränk oder Sonstiges. Die Kategorie hat später eine praktische Bedeutung: Im Bondruck-System ist pro Kategorie ein eigener Drucker konfigurierbar, sodass Essensbons direkt in der Küche ankommen und Getränkebons an der Bar.

Produkte haben Varianten. Eine Variante ist eine konkrete Ausprägung des Produkts mit eigenem Namen und Preis. Zum Beispiel hat das Produkt „Bier" die Varianten „0,3 l — 2,50 €" und „0,5 l — 3,80 €". Preise werden immer in Cent gespeichert: 250 und 380. Das ist keine Eigenheit, sondern eine fundamentale technische Entscheidung — dazu später mehr.

Produkte und Varianten können aktiviert und deaktiviert werden. Nur aktive Varianten erscheinen im Service-Bestellformular. Ein gelöschtes Produkt wird mit `status = 'deleted'` markiert und bleibt in der Datenbank erhalten — es ist ein Soft-Delete. Das ist wichtig für die historische Nachvollziehbarkeit: Alte Bestellungen referenzieren Varianten, die möglicherweise inzwischen gelöscht wurden.

### 3.2 Tische

Ein Tisch ist eine benannte Einheit, an der Gäste Platz nehmen: „Tisch 1", „Tisch 2", „Biergarten links", „VIP-Tisch". Der Name ist frei wählbar. Tische haben ebenfalls einen Status — aktiv, inaktiv, gelöscht. Nur aktive Tische erscheinen im Service.

### 3.3 Benutzer

Der Admin legt alle Mitarbeiter als Benutzer an: Name, Benutzername und Rolle. Es gibt drei Rollen:

- **Service:** Darf Bestellungen aufnehmen, Ausgabe bestätigen und Zahlungen kassieren. Keine Stornierungen, keine Auszahlungen.
- **Serviceleitung:** Wie Service, aber zusätzlich Stornierungen erteilen und Auszahlungen leisten.
- **Admin:** Voller Zugriff auf Stammdaten und alle Kassenfunktionen.

Für jeden neuen Benutzer erzeugt das System ein sechsstelliges Einmal-Passwort. Der Admin teilt dieses Passwort dem Mitarbeiter mit — per Zettel, SMS, was auch immer. Beim ersten Login gibt der Mitarbeiter seinen Benutzernamen und die PIN ein, das System leitet ihn zur Passwortänderung weiter.

---

## 4. Kasseneröffnung

Vor jeder Veranstaltung muss eine Kassensitzung eröffnet werden. Eine Kassensitzung ist der administrative Rahmen für einen Betriebstag. Sie bekommt eine fortlaufende Nummer — die sogenannte Z-Nummer oder Kassenabschlussnummer. Diese Nummer ist das Herzstück der gesetzlichen Anforderungen: Im deutschen Kassenrecht (DSFinV-K, KassenSichV) identifiziert die Z-Nummer eindeutig jeden Tagesabschluss.

Der Admin eröffnet die Kassensitzung: Er gibt das Datum und optional eine Bezeichnung ein, beispielsweise „Sommerfest 2026". Das System erzeugt einen Event vom Typ `kassensitzung-eroeffnet:v1` und trägt die neue Kassensitzung in die Datenbank ein. Ab diesem Moment ist die Kasse „offen" und der Betrieb kann beginnen.

Danach setzt der Admin den Anfangsbestand — das Wechselgeld, das zu Beginn in der Kasse liegt. Das ist zum einen eine Betriebspflicht, zum anderen die Grundlage für den späteren Kassensturz: Der Soll-Bestand errechnet sich aus Anfangsbestand plus alle eingegangenen Zahlungen minus alle Auszahlungen minus alle Stornierungen plus alle Kassenbewegungen. Der Anfangsbestand wird als eigener Event `anfangsbestand-gesetzt:v1` protokolliert.

Eine wichtige Invariante: Es darf immer nur eine offene Kassensitzung geben. Wenn bereits eine Kassensitzung offen ist, blockiert das System die Eröffnung einer weiteren. Jede Kassensitzung muss mit einem Tagesabschluss formal abgeschlossen werden, bevor eine neue beginnen kann.

---

## 5. Service-Betrieb: Die Tisch-Session

Nun sind die Servicekräfte an der Reihe. Sie öffnen den Browser auf ihrem Smartphone, geben ihre Zugangsdaten ein und werden in den Service-Bereich weitergeleitet.

### 5.1 Tischauswahl

Die Service-Startseite zeigt zwei Sichten: „Meine Tische" (persönliche Favoriten) und alle aktiven Tische. Jeder Tisch zeigt seinen aktuellen Saldo — der offene Betrag, den die Gäste noch bezahlen müssen. Die Servicekraft tippt auf einen Tisch und kommt zur Tischdetailansicht.

Jeder Benutzer kann Tische als Favoriten markieren. Das erleichtert den Betrieb bei größeren Veranstaltungen: Jede Servicekraft betreut ihre eigenen Tische und sieht sie direkt auf der Startseite, ohne durch die gesamte Tischliste scrollen zu müssen.

### 5.2 Bestellung aufnehmen

In der Tischdetailansicht öffnet die Servicekraft den Bestellungs-Drawer — ein Schiebepanel von unten, das die Produktliste zeigt, sortiert nach Kategorie. Sie tippt auf ein Produkt, wählt die Variante aus, gibt die Menge ein. Mehrere verschiedene Positionen können gleichzeitig bestellt werden. Optional kann sie einen Kommentar hinterlassen.

Beim Absenden sendet das Frontend die Bestellung an das Backend. Das Backend validiert die Eingabe, lädt die aktuellen Produktdaten aus der Datenbank und schreibt einen Event `bestellung-aufgenommen:v1` in das Kassenjournal. Dieser Event enthält alle bestellten Positionen mit Produktname, Variantenname, Kategorie, Einzelpreis und Menge — ein sogenannter Fat Event. Das bedeutet: Alle relevanten Produktinformationen werden zum Zeitpunkt der Bestellung in den Event eingefroren. Wenn der Admin später den Preis einer Variante ändert oder ein Produkt umbenennt, beeinflusst das keine historischen Bestellungen.

Nach dem Schreiben des Events löst das System — falls konfiguriert — den Bondruck aus: Die Bestellung wird an den zuständigen Drucker geschickt (für Essen an den Küchendrucker, für Getränke an den Thekendrucker).

### 5.3 Ausgabe bestätigen

Wenn die bestellten Speisen und Getränke fertig sind und ausgegeben werden, bestätigt die Servicekraft die Ausgabe. Sie sieht eine Liste aller ausstehenden Positionen — also alles, was bestellt aber noch nicht als ausgegeben markiert wurde. Sie kann einzelne Positionen oder alle auf einmal bestätigen. Das System schreibt einen `ausgabe-bestaetigt:v1` Event.

Die Ausgabebestätigung hat keine direkte finanzielle Auswirkung auf den Saldo. Sie ist ein operativer Status: Sie hilft der Servicekraft zu wissen, was noch aussteht, und sie ist die Grundlage für den späteren Tagesabschluss-Check, der sicherstellt, dass keine offenen Positionen vergessen wurden.

### 5.4 Zahlung kassieren

Wenn die Gäste bezahlen möchten, öffnet die Servicekraft den Zahlungs-Drawer. Sie sieht alle unbezahlten Positionen des Tisches — also alle bestellten, noch nicht bezahlten Positionen. Sie kann einzelne Positionen auswählen (Teilzahlung) oder alle auf einmal. Das System zeigt den Gesamtbetrag an.

Die Servicekraft kassiert das Geld, gibt optional einen Kommentar ein (z.B. „Runde 2") und bestätigt die Zahlung. Das Backend schreibt einen `zahlung-kassiert:v1` Event. Der Saldo des Tisches sinkt entsprechend.

Zahlungen sind immer in bar — jotti unterstützt keine Kartenzahlung. Das ist eine bewusste Designentscheidung: Kartenzahlung erfordert teure POS-Terminals, externe Dienstleister und erheblich mehr Compliance-Aufwand. Für Vereinsfeste, wo ohnehin fast ausschließlich bar bezahlt wird, wäre das ein unverhältnismäßiger Overhead.

### 5.5 Stornierung erteilen

Wenn etwas falsch bestellt wurde, etwas nicht mehr verfügbar ist oder ein Gast seinen Platz verlässt, ohne zu bezahlen, kann eine Serviceleitung oder ein Admin eine Stornierung erteilen. Der Saldo des Tisches sinkt um den stornierten Betrag — er kann dabei auch negativ werden, wenn mehr storniert wird als noch bezahlt werden muss.

Stornierungen erfordern einen Kommentar (mindestens drei Zeichen). Das ist keine willkürliche Einschränkung, sondern eine gesetzliche Anforderung: Jede Stornierung muss begründet und dokumentiert werden. Der Kommentar wird im Event `stornierung-erteilt:v1` gespeichert.

Die Einschränkung auf Serviceleitungen und Admins bei Stornierungen ist ebenfalls keine willkürliche Restriktion — sie folgt dem Vier-Augen-Prinzip. Eine normale Servicekraft soll keine Bestellung einfach wegmachen können. Das schützt den Verein vor Missbrauch und entspricht guten Kassenpraktiken.

### 5.6 Auszahlung leisten

Wenn ein Tisch nach einer Stornierung einen negativen Saldo hat, muss das Geld zurückgegeben werden. Die Serviceleitung leistet eine Auszahlung: Sie gibt den Betrag ein, gibt Wechselgeld zurück und schreibt einen `auszahlung-geleistet:v1` Event. Der Saldo des Tisches steigt wieder.

---

## 6. Kassenführung

Parallel zum Service-Betrieb gibt es kassenführende Aufgaben, die nur Admins und Serviceeitungen ausführen.

### 6.1 Kassenbewegungen buchen

Manchmal muss Geld in die Kasse eingelegt oder entnommen werden, ohne dass es mit einem Tischvorgang zusammenhängt. Drei Arten von Kassenbewegungen gibt es:

- **Geldtransit:** Geld, das zu einer übergeordneten Kasse oder Bank weitergegeben wird.
- **Privatentnahme:** Der Kassenverantwortliche entnimmt privates Geld aus der Kasse (z.B. weil er Münzgeld hineingelegt hat).
- **Privateinlage:** Der Kassenverantwortliche legt privates Geld in die Kasse ein.

Jede Kassenbewegung erfordert einen Kommentar und wird als `kassenbewegung-gebucht:v1` Event protokolliert. Diese Bewegungen fließen in den Soll-Bestand ein.

### 6.2 Kassenbestand einsehen

Der Admin kann jederzeit den aktuellen Kassenbestand einsehen. Das System berechnet den Soll-Bestand in Echtzeit aus dem Kassenjournal:

```
Soll-Bestand = Anfangsbestand
             + alle Zahlungen (vom Gast kassiert)
             − alle Stornierungen (Geld zurückgegeben)
             − alle Auszahlungen (Geld zurückgegeben)
             + alle Geldtransit-Einlagen
             − alle Geldtransit-Entnahmen
             + alle Privateinlagen
             − alle Privatentnahmen
             − alle Differenzbuchungen aus Kassenstürzen
```

Das ist der Betrag, der rechnerisch in der Kasse liegen sollte. Zählt man das tatsächliche Bargeld und kommt auf denselben Betrag, ist alles in Ordnung.

---

## 7. Tagesabschluss

Am Ende der Veranstaltung folgt der formale Tagesabschluss — ein mehrstufiger Prozess, der durch klare Vorbedingungen abgesichert ist.

### 7.1 Kassensturz

Der Kassensturz ist das manuelle Zählen des Kassenbestands. Der Admin zählt das Bargeld und gibt den gezählten Betrag ein. Das System vergleicht ihn mit dem Soll-Bestand und zeigt die Differenz. Eine Differenz ist nicht ungewöhnlich — Rechenfehler, vergessenes Wechselgeld, falsch herausgegebene Münzen.

Die Differenz wird automatisch als `kassensturz-durchgefuehrt:v1` Event und als separater `differenz-soll-ist-gebucht:v1` Event protokolliert. Damit ist die Differenz transparent im Journal festgehalten — nicht versteckt, nicht ignoriert.

### 7.2 Tagesabschluss erstellen (Z-Bon)

Bevor der Tagesabschluss erstellt werden kann, prüft das System zwei Vorbedingungen:

1. Ein Kassensturz wurde durchgeführt.
2. Alle Tische haben einen Saldo von null — es gibt keine offenen Rechnungen mehr.

Sind beide Bedingungen erfüllt, erstellt der Admin den Tagesabschluss. Das System schreibt einen `tagesabschluss-erstellt:v1` Event mit allen relevanten Summenwerten: Gesamtumsatz, Stornierungen, Auszahlungen, Geldtransit, Z-Nummer, Zeitraum. Die Kassensitzung wird auf `abgeschlossen` gesetzt.

Der Tagesabschluss ist der Z-Bon — im deutschen Kassenrecht das zentrale Dokument, das jeden Geschäftstag abschließt. Der Z-Bon muss aufbewahrt und auf Verlangen der Finanzverwaltung vorgelegt werden.

---

## 8. Fiskalkonformität

jotti ist ein elektronisches Aufzeichnungssystem im Sinne von § 1 KassenSichV. Als solches unterliegt es der TSE-Pflicht nach § 146a AO — unabhängig davon, ob der Betreiber ein gemeinnütziger Verein ist, ob die Veranstaltung nur dreimal im Jahr stattfindet oder ob keine Gewinnerzielungsabsicht besteht. Das Gesetz differenziert in diesen Punkten nicht.

### 8.1 Warum gilt das für Vereine?

§ 146a AO gilt für jeden, der aufzeichnungspflichtige Geschäftsvorfälle mit einem elektronischen System erfasst. Ein Vereinsfest mit Getränke- und Essensverkauf ist ein wirtschaftlicher Geschäftsbetrieb. Sofern er nicht unter die Sonderregelung des § 67a AO fällt (steuerbefreiter Zweckbetrieb bei sportlichen Veranstaltungen bis 45.000 EUR Umsatz), ist er steuerpflichtig und kassenrechtlich relevant.

Selbst bei steuerbefreiten Zweckbetrieben empfehlen Steuerberater zunehmend die Nutzung TSE-gesicherter Kassen, da die Abgrenzung im Prüfungsfall problematisch sein kann.

### 8.2 BYOD und die Einordnung der Smartphones

In jottis Modell sind die Smartphones der Servicekräfte rechtlich gesehen reine **Eingabegeräte** — ihr Zweck ist nicht unähnlich einer einfachen Tastatur. Sie erfassen keine Daten autonom, sie können nicht offline Zahlungen aufnehmen, und alle relevanten Berechnungen und Protokollierungen finden ausschließlich im Backend statt.

Das ist wichtig, weil die Smartphones dadurch **keine eigene TSE benötigen**. Die gesamte TSE-Absicherung erfolgt zentral im Backend. Diese Einordnung ist aber an eine architektonische Pflicht geknüpft: Die Webapp muss bei einem Internetausfall sofort blockieren und jede Offline-Erfassung von Barzahlungen technisch verhindern. Sobald keine Verbindung zum Backend besteht, darf keine Zahlung erfasst werden. Nur so ist die Einordnung als reines Eingabegerät rechtlich haltbar.

### 8.3 TSE-Integration (Roadmap)

Die TSE-Integration ist in jottis Compliance-Roadmap als Phase 2 geplant. Als Anbieter ist fiskaly (Cloud-TSE, API-first) vorgesehen. Der Code wird ein `TSEClient`-Interface vorsehen, sodass andere TSE-Anbieter als Adapter angebunden werden können.

Eine TSE signiert jeden Kassiervorgang kryptographisch. Die Signatur und eine fortlaufende Transaktionsnummer werden im Kassenjournal gespeichert und erscheinen auf dem Beleg. Damit ist jeder Kassiervorgang manipulationssicher nachweisbar.

### 8.4 DSFinV-K Export

§ 4 KassenSichV verlangt eine standardisierte digitale Schnittstelle für die Finanzverwaltung. Das DSFinV-K-Format (Digitale Schnittstelle der Finanzverwaltung für Kassensysteme, Version 2.4) definiert ein ZIP-Archiv mit mehreren CSV-Dateien: Transaktionen, Positionen, Zahlungen, Kassenbewegungen, Tagesabschlüsse — alles mit festgeschriebenen Spaltenreihenfolgen und deutschen Dateinamen.

Für jotti ist dieser Export in Phase 2 geplant. Da alle Daten im Kassenjournal in unveränderlicher Form vorliegen, ist der Export konzeptionell eine reine Aggregation und Umformatierung der vorhandenen Events.

### 8.5 ELSTER-Meldepflicht

Ab dem 1. Januar 2025 müssen elektronische Aufzeichnungssysteme dem Finanzamt gemeldet werden. In Phase 1 gibt es dafür eine Anleitung für Betreiber, wie sie die Meldung manuell über das ELSTER-Webportal durchführen. In Phase 2 ist eine programmatische Meldung über die ERiC-Schnittstelle oder fiskaly geplant.

### 8.6 Steuersätze

In Phase 1 werden Steuersätze in das System eingeführt: 19 % (Normalsatz, z.B. Getränke), 7 % (ermäßigter Satz, z.B. Speisen), 0 % bzw. steuerbefreit (Zweckbetrieb). Diese Information wird auf den Varianten hinterlegt und in Fat Events eingefroren.

### 8.7 GoBD-Konformität

Die GoBD (Grundsätze zur ordnungsmäßigen Führung und Aufbewahrung von Büchern) verlangen Nachvollziehbarkeit, Vollständigkeit, Richtigkeit, zeitgerechte Buchungen und Unveränderbarkeit. Das Event-Sourcing-Modell von jotti erfüllt diese Anforderungen strukturell:

- **Nachvollziehbarkeit:** Jeder Event enthält Zeitstempel, Benutzer-ID und Benutzername, Tisch und alle relevanten Daten.
- **Vollständigkeit:** Jede Transaktion ist ein eigener Event — nichts wird überschrieben.
- **Richtigkeit:** Validierung auf beiden Seiten (Backend mit zog, Frontend mit Zod).
- **Zeitgerechte Buchung:** Events werden im Moment der Aktion geschrieben.
- **Unveränderbarkeit:** PostgreSQL-Trigger verhindern UPDATE, DELETE und TRUNCATE auf der kassenjournal-Tabelle auf Datenbankebene.

---

## 9. Warum Event-Sourcing?

Die zentrale Architekturentscheidung von jotti ist der Einsatz von Event-Sourcing für alle Kassenvorgänge. Das ist keine modische Wahl — es gibt konkrete, gewichtige Gründe.

### 9.1 Das Problem mit CRUD-Kassensystemen

Ein klassisches CRUD-Kassensystem würde so funktionieren: Es gibt eine Tabelle `bestellungen`, eine Tabelle `zahlungen`, eine Tabelle `stornierungen`. Wenn eine Bestellung storniert wird, wird der Datensatz in `bestellungen` aktualisiert oder gelöscht und in `stornierungen` ein neuer Datensatz angelegt.

Das Problem: Bei diesem Modell lässt sich im Nachhinein nicht mehr lückenlos rekonstruieren, wann was von wem geändert wurde. Die Stornierung verändert den Zustand der Bestellung, aber wenn der ursprüngliche Datensatz überschrieben oder gelöscht wird, ist die historische Information weg. Eine Betriebsprüfung, die alle Stornierungen eines Abends sehen will, stößt auf ein System, das die Antwort nicht vollständig geben kann.

Noch schlimmer: Ein technischer Fehler, ein versehentliches UPDATE ohne WHERE-Klausel, eine Manipulation durch einen Mitarbeiter mit Datenbankzugriff — all das kann in einem CRUD-System unbemerkt die Buchhaltung verfälschen. Im schlimmsten Fall ist das ein steuerrechtliches Problem.

### 9.2 Event-Sourcing als unveränderliches Protokoll

Event-Sourcing löst das fundamental anders: Statt den Zustand zu speichern, werden alle Zustandsübergänge als unveränderliche Events gespeichert. Das Kassenjournal ist ein chronologisches, append-only Protokoll aller Geschäftsvorfälle.

Wenn eine Bestellung storniert wird, wird nicht die ursprüngliche Bestellung geändert. Stattdessen wird ein neuer Event `stornierung-erteilt:v1` hinzugefügt, der die stornierten Positionen, den Betrag, den Zeitpunkt, den ausführenden Benutzer und den Pflichtkommentar enthält. Die ursprüngliche Bestellung ist weiterhin im Journal — unveränderlich.

Das hat mehrere Konsequenzen:

- **Vollständiger Audit-Trail:** Jede Aktion jedes Benutzers zu jedem Zeitpunkt ist nachvollziehbar. Wer hat wann was bestellt, kassiert, storniert? Das Journal gibt immer eine vollständige Antwort.
- **Manipulationssicherheit:** PostgreSQL-Trigger auf der kassenjournal-Tabelle verhindern auf Datenbankebene, dass Events geändert oder gelöscht werden. Selbst ein Datenbankadministrator kann den Verlauf nicht im Stillen revidieren.
- **GoBD-Konformität:** Die Unveränderlichkeit ist kein nice-to-have, sondern eine gesetzliche Anforderung. Das Event-Sourcing-Modell erfüllt sie strukturell, ohne dass man die Unveränderlichkeit erst nachträglich durch Audit-Log-Tabellen nachbauen muss.
- **TSE-Vorbereitung:** Die TSE (Technische Sicherheitseinrichtung) signiert jeden Kassiervorgang. In jottis Modell ist jeder Zahlungsvorgang ein Event — die TSE-Signatur wird direkt in den Event-Daten gespeichert. Bei einem CRUD-System müsste man extra ein TSE-Journal neben den normalen Tabellen führen.
- **Zeitreise und Replay:** Da der gesamte Zustand aus Events rekonstruierbar ist, kann man den Zustand eines Tisches zu jedem beliebigen Zeitpunkt in der Vergangenheit wiederherstellen. Das ist wertvoll für Fehleranalysen und für die Entwicklung neuer Projektionen.

### 9.3 Fat Events: Schutz vor Stammdaten-Veränderungen

Ein spezifisches Problem bei Kassensystemen: Produktpreise und -namen ändern sich. Wenn eine Bestellung auf eine Variante verweist und der Preis dieser Variante später geändert wird, stimmt die historische Bestellung nicht mehr mit dem damaligen Preis überein.

jotti löst das durch Fat Events: Jeder Bestellungs-Event enthält nicht nur die Varianten-ID, sondern auch eine Kopie aller relevanten Produktdaten zum Zeitpunkt der Bestellung — Produktname, Variantenname, Kategorie und Einzelpreis. Diese Daten werden in den Event eingefroren. Spätere Stammdatenänderungen haben keinen Einfluss auf historische Bestellungen.

Das ist eine Anti-Corruption Layer: Der Kasse-Kontext schützt sich vor Änderungen im Stammdaten-Kontext.

---

## 10. Warum synchrone Projektionen?

Event-Sourcing hat einen Trade-off: Um den aktuellen Zustand zu kennen, müsste man theoretisch alle Events eines Streams von Anfang an replayen. Bei einer aktiven Tisch-Session mit vielen Bestellungen wäre das ineffizient.

### 10.1 Das Problem mit reinem Replay

Stell dir vor, ein Tisch hat im Lauf des Abends 20 Bestellungen, 15 Zahlungen, 3 Stornierungen und 2 Ausgaben — insgesamt 40 Events. Bei jedem API-Aufruf alle 40 Events aus der Datenbank zu lesen, zu deserialisieren und zu replayen, wäre kein Performanceproblem für eine moderne Datenbank, aber es ist unnötig — der Zustand ändert sich selten zwischen zwei Aufrufen.

Bei 30 Servicekräften, die alle gleichzeitig Tischzustände abfragen, summiert sich das: Jede Anfrage liest mehrere Dutzend Zeilen, deserialisiert JSON, aggregiert Saldi. Das ist unnötige Last.

### 10.2 Die synchrone Projektion

jotti löst das mit einer synchronen Projektion: Es gibt eine Tabelle `tisch_sessions`, die den aktuellen Zustand jeder Tisch-Session materialisiert enthält — Saldo, unbezahlte Positionen, ausstehende Positionen, Gesamtzahlungen. Diese Tabelle wird in **derselben Datenbanktransaktion** aktualisiert, in der der Event ins Kassenjournal geschrieben wird.

Das ist der entscheidende Punkt: synchron bedeutet, dass die Projektion nie veraltet sein kann. Es gibt keine Verzögerung, keine asynchrone Verarbeitung, kein „eventual consistency"-Problem. Wenn ein Schreibvorgang erfolgreich war, ist die Projektion garantiert aktuell.

Der Ablauf für eine Bestellung:

1. Backend empfängt die Bestellanfrage.
2. Backend beginnt eine Datenbanktransaktion.
3. Backend schreibt den `bestellung-aufgenommen:v1` Event ins Kassenjournal.
4. Backend aktualisiert die `tisch_sessions`-Zeile: Saldo erhöhen, neue Positionen zu den Arrays hinzufügen.
5. Backend schreibt den Drucker-Job in die Relay-Queue (falls konfiguriert).
6. Backend committed die Transaktion.
7. Response an das Frontend.

Wenn in Schritt 4 oder 5 etwas schiefgeht, wird die gesamte Transaktion zurückgerollt — kein Event ohne aktualisierte Projektion, keine Projektion ohne Event. Atomarität.

### 10.3 Warum nicht asynchron?

Asynchrone Projektionen (z.B. über einen Message-Bus oder Polling) hätten folgende Nachteile in jottis Kontext:

- **Komplexität:** Ein Message-Bus, eine separate Projektions-Service, Fehlerbehandlung bei verpassten Events — das ist erhebliche Infrastruktur für ein System, das möglichst einfach und wartbar sein soll.
- **Consistency-Fenster:** Wenn das Frontend direkt nach einer Bestellung den Tischzustand abfragt und die Projektion noch nicht aktualisiert ist, sieht die Servicekraft eine veraltete Ansicht. Das führt zu Verwirrung und potenziell zu Fehlern (z.B. doppelten Bestellungen).
- **Kein echter Bedarf:** Event-Sourcing-Systeme verwenden asynchrone Projektionen typischerweise für Skalierbarkeit (z.B. wenn die Ereignisrate die synchrone Verarbeitung überlasten würde) oder für Projektionen in andere Datenspeicher. Bei jottis Lastprofil — ein Vereinsfest mit 50 Tischen und 30 Servicekräften — ist das kein Problem.

Die synchrone Projektion ist damit die einfachste Lösung, die korrekt funktioniert.

### 10.4 Optimistic Concurrency Control (OCC)

Was passiert, wenn zwei Servicekräfte gleichzeitig eine Bestellung am selben Tisch aufnehmen? Das Backend könnte beide Schreibversuche akzeptieren und inkonsistente Zustände erzeugen.

jotti löst das mit Optimistic Concurrency Control über einen Datenbankconstraint: `UNIQUE(subject, version)` auf der kassenjournal-Tabelle. Jeder Event für einen bestimmten Stream hat eine aufsteigende Versionsnummer. Wenn zwei Schreibversuche gleichzeitig versuchen, Version 5 zu schreiben, schlägt einer von ihnen mit einem eindeutigen Constraint-Fehler fehl. Das Backend kann den fehlgeschlagenen Versuch dann mit der aktuellen Version erneut versuchen (Retry) oder einen Fehler an das Frontend zurückgeben.

OCC ist hier sinnvoll, weil Konflikte selten sind — es ist ungewöhnlich, dass zwei Servicekräfte gleichzeitig exakt denselben Tisch bedienen. Wenn doch, ist ein gelegentlicher Retry akzeptabel. Sperren (Pessimistic Locking) wären unnötige Komplexität für einen seltenen Fall.

---

## 11. Warum dieses Datenmodell?

### 11.1 Kassenjournal als Single Source of Truth

Das Kassenjournal ist die zentrale Wahrheit aller finanziellen Vorgänge. Alle anderen Tabellen sind entweder Projektionen des Kassenjournals oder Stammdaten, die mit dem Kassenjournal interagieren.

Diese Architektur hat einen wichtigen Vorteil: Es gibt keine Inkonsistenzprobleme zwischen verschiedenen Tabellen. Wenn jotti eine Zusammenfassung aller Bestellungen des Abends braucht, fragt es das Kassenjournal — die unveränderliche Quelle. Es gibt keine separate `berichte`-Tabelle, die veraltet sein könnte.

### 11.2 Subjects und die hierarchische Struktur

Die Subject-Konvention (`kassensitzung-1/tisch-42`) spiegelt die fachliche Hierarchie direkt in den Daten wider. Eine Kassensitzung hat viele Tisch-Sessions. Diese Hierarchie ist im Subject sichtbar.

Das hat praktische Vorteile: Mit `WHERE subject = 'kassensitzung-1/tisch-42'` werden alle Events eines bestimmten Tisches einer bestimmten Kassensitzung abgefragt. Mit `WHERE kassensitzung_nr = 1` werden alle Events aller Tische einer Kassensitzung abgefragt — nützlich für Reporting und Kassenbestand.

Die Spalte `kassensitzung_nr` in der kassenjournal-Tabelle ist eine Denormalisierung: Die Information steckt auch im Subject (die Nummer nach `kassensitzung-`), aber als explizite Integer-Spalte ist sie indizierbar und effizient abfragbar, ohne String-Operationen.

### 11.3 Geldbeträge in Cent

Alle Geldbeträge werden als Integer in Cent gespeichert. Niemals als Dezimalzahl (Float oder Decimal). Das ist kein stilistisches Merkmal — es ist eine technische Notwendigkeit.

Fließkommazahlen (Float/Double) können viele Dezimalwerte nicht exakt darstellen. Der Wert 2,50 EUR ist in IEEE 754 Float nicht exakt repräsentierbar. Bei Berechnungen mit Fließkommazahlen akkumulieren sich Rundungsfehler. Ein System, das 1.000 Bestellungen à 2,50 EUR aufsummiert, könnte statt 2.500,00 EUR einen Wert wie 2499,9999... oder 2500,0001... ausgeben. In einem Kassensystem sind solche Fehler inakzeptabel — der Soll-Bestand muss auf den Cent genau stimmen.

Integer-Arithmetik in Cent ist exakt: 250 + 380 = 630. Kein Rundungsfehler.

### 11.4 Soft-Deletes für Stammdaten

Produkte, Varianten, Tische und Benutzer werden niemals hart gelöscht. Stattdessen bekommen sie den Status `deleted`. Das ist aus zwei Gründen notwendig:

Erstens referenzieren Events im Kassenjournal auf Entitäten (z.B. `variante_id`). Wenn eine Variante hart gelöscht würde, wäre diese Referenz invalid — historische Events würden auf nicht-existente Datenbankzeilen zeigen. Fat Events frieren zwar die Produktdaten ein, aber die Varianten-ID bleibt als Referenz erhalten.

Zweitens sind gelöschte Entitäten für Audit-Zwecke und Fehleranalysen wertvoll. Wenn ein Admin versehentlich eine Variante löscht und danach merkt, dass sie noch gebraucht wird, kann sie durch Reaktivierung wiederhergestellt werden — statt neu erstellt werden zu müssen.

### 11.5 CRUD für Stammdaten — kein Event-Sourcing

Während das Kassenjournal auf Event-Sourcing basiert, sind Stammdaten (Produkte, Tische, Benutzer) klassisches CRUD. Das ist eine bewusste Entscheidung:

Stammdaten ändern sich selten und haben keine komplexen Zustandsübergänge, die einen vollständigen Verlauf erfordern. Ein Produkt wird erstellt, vielleicht umbenannt, vielleicht deaktiviert. Die Geschichte dieser Änderungen ist für das operative Kassensystem irrelevant — was zählt, ist der aktuelle Zustand.

Event-Sourcing für Stammdaten wäre Over-Engineering: mehr Komplexität, mehr Infrastruktur, mehr Abstraktionsebenen — ohne messbaren Nutzen.

Die klare Trennung — Event-Sourcing für die Core Domain (Kasse), CRUD für die Supporting Sub-Domain (Stammdaten) — ist eine Anwendung der Domain-Driven Design-Prinzipien: Die wertvollste Komplexität (das Kassenjournal mit seinen Anforderungen an Unveränderlichkeit, Audit-Trail und TSE-Kompatibilität) bekommt die angemessene Architektur. Der Rest bleibt einfach.

---

## 12. Zusammenfassung

jotti verbindet eine einfache Benutzeroberfläche mit einer durchdachten technischen Architektur. Die Simplizität, die ein ehrenamtlicher Vereinshelfer am Smartphone erlebt — Tisch antippen, Getränke auswählen, bestätigen — ist das Ergebnis bewusster Designentscheidungen, die jeweils einen konkreten Grund haben.

Event-Sourcing ist nicht Selbstzweck, sondern die direkte Antwort auf die Anforderungen des deutschen Kassenrechts: Unveränderlichkeit, Vollständigkeit, Nachvollziehbarkeit. Synchrone Projektionen sind nicht eine beliebige Optimierung, sondern die einfachste Lösung, die starke Konsistenz garantiert ohne zusätzliche Infrastruktur. Geldbeträge in Cent sind nicht eine merkwürdige Eigenheit, sondern der einzig korrekte Umgang mit Dezimalwährungen in Software. Und Fat Events sind nicht redundante Datenduplizierung, sondern der entscheidende Schutz vor dem Problem, das jedes Kassensystem hat: Preise ändern sich, aber historische Buchungen müssen unveränderlich bleiben.

Die Architektur ist so gestaltet, dass sie — wenn in Phase 2 die TSE-Integration (fiskaly) kommt — nahtlos erweiterbar ist: Jeder Zahlungsvorgang ist bereits ein eigenständiger, identifizierbarer Event. Die TSE-Signatur wird einfach in den Event-Daten mitgespeichert. Der DSFinV-K-Export aggregiert das Kassenjournal in das gesetzlich vorgeschriebene Format. Keine Umstrukturierung der Daten nötig — die Architektur war von Anfang an auf diesen Weg vorbereitet.
