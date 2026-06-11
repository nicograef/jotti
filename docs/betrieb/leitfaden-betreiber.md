# Betreiber-Leitfaden: jotti rechtssicher betreiben

> ⚖️ **Dieser Leitfaden ist keine Rechts- oder Steuerberatung.** Er fasst die Rechtslage
> nach bestem Wissen zusammen. Im Zweifel klärt Details mit eurem **Steuerberater** oder
> dem zuständigen **Finanzamt**.

---

## 1. Das Wichtigste in 60 Sekunden

- jotti ist eine **Kasse** im Sinne des Gesetzes (ein „elektronisches Aufzeichnungssystem").
  Damit gelten dieselben Regeln wie für jede Registrierkasse.
- **Drei Dinge** müsst ihr als Verein selbst erledigen:
  1. Eine **TSE** (manipulationssicheres Signaturmodul) bei einem Anbieter wie **fiskaly**
     buchen und in jotti eintragen.
  2. Eure Kasse **beim Finanzamt anmelden** (online über ELSTER) — dafür braucht ihr die
     **Seriennummer**, die jotti euch im Admin-Bereich anzeigt.
  3. Alle Kassendaten **10 Jahre aufbewahren** (regelmäßige Backups).
- **Der Entwickler von jotti ist dafür nicht verantwortlich.** Er stellt nur den Programmcode
  bereit. Die Pflichten gegenüber dem Finanzamt liegen **bei euch als Verein**.
- **Was kostet uns das?** jotti selbst ist für berechtigte Vereine **kostenlos**. Kosten
  entstehen nur für den **Server** und die **Cloud-TSE** (kleine laufende Gebühr beim Anbieter).
- ⏳ **Aktueller Stand:** TSE-Anbindung, Belegausgabe und DSFinV-K-Export sind noch **in
  Entwicklung**. Solange diese Funktionen nicht fertig sind, ist jotti **noch nicht** für den
  fiskalpflichtigen Echtbetrieb geeignet. Dieser Leitfaden beschreibt, was zu tun ist,
  **sobald** die Funktionen verfügbar sind.

---

## 2. Wer ist verantwortlich? Entwickler vs. Verein

jotti ist **keine gemietete Cloud-Software** (kein „SaaS"), sondern **quelloffener Code**, den
ihr **selbst auf eurem eigenen Server** installiert und betreibt. Daraus ergibt sich eine klare
Rollentrennung:

| Aufgabe                                                        | Entwickler (jotti) | Euer Verein (Betreiber) |
| -------------------------------------------------------------- | :----------------: | :---------------------: |
| Programmcode mit TSE- und DSFinV-K-Schnittstelle bereitstellen |         ✅         |            —            |
| Vertrag mit TSE-Anbieter (z. B. fiskaly) abschließen           |         —          |           ✅            |
| TSE-Schlüssel in jotti eintragen                               |         —          |           ✅            |
| Kasse beim Finanzamt anmelden (ELSTER)                         |         —          |           ✅            |
| Server betreiben, Backups, Datenschutz                         |         —          |           ✅            |
| Daten 10 Jahre aufbewahren                                     |         —          |           ✅            |
| Belege ausgeben (bzw. Befreiung beantragen)                    |         —          |           ✅            |

> **Warum ist das so?** Wer eine Kasse **tatsächlich verwendet**, ist gegenüber dem Finanzamt
> verantwortlich — das seid **ihr als Verein**, nicht der Programmierer. Der Entwickler hat
> seine gesetzliche Pflicht erfüllt, indem er Software bereitstellt, die eine zertifizierte TSE
> anbinden **kann** (das verlangt § 146a Abs. 1 Satz 5 AO). Ob ihr die TSE dann auch wirklich
> einrichtet, liegt in eurer Verantwortung.
>
> _Rechtsgrundlage: § 146a Abs. 1 Satz 5 AO (Pflicht des Herstellers), § 146a Abs. 4 AO
> (Meldepflicht des Betreibers), § 379 AO (Bußgeld bei Verstoß)._

---

## 3. Was heißt „fiskalkonform"? Einfach erklärt

„Fiskalkonform" bedeutet: Die Kasse erfüllt alle Anforderungen, die das Finanzamt an
elektronische Kassen stellt, damit **niemand Umsätze heimlich löschen oder verändern kann**.
Konkret braucht eine konforme Kasse vier Bausteine:

| Baustein                        | Was es bedeutet                                                | Wie jotti es löst                                                                                                    | Status            |
| ------------------------------- | -------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------- | ----------------- |
| **Unveränderbare Aufzeichnung** | Einmal gebuchte Vorgänge dürfen nie heimlich geändert werden   | „Event-Sourcing": Jeder Vorgang wird als unveränderlicher Eintrag gespeichert; Korrekturen nur als neue Gegenbuchung | ✅ vorhanden      |
| **TSE-Signatur**                | Ein Sicherheitsmodul „versiegelt" jeden Vorgang kryptografisch | Anbindung an eine Cloud-TSE (z. B. fiskaly)                                                                          | ⏳ in Entwicklung |
| **Beleg**                       | Für jeden Kassiervorgang kann ein gültiger Bon erstellt werden | Kassenbeleg auf Knopfdruck, mit allen Pflichtangaben                                                                 | ⏳ in Entwicklung |
| **DSFinV-K-Export**             | Standard-Datenpaket, das ein Prüfer einlesen kann              | ZIP mit genormten CSV-Dateien (`transactions.csv`, `lines.csv` …)                                                    | ⏳ in Entwicklung |

> **Wichtig:** jotti ist **so gebaut**, dass es nach Fertigstellung dieser Bausteine konform
> ist. Heute ist es das noch nicht vollständig; den genauen Stand seht ihr in
> [anforderungen.md, Abschnitt 6 „Fiskalkonformität"](../anforderungen.md). Ein TSE-freier
> Betrieb ist nur für **Tests/Entwicklung** gedacht (für echte Verkäufe nicht konform, § 379 AO).

---

## 4. Die Gesetze in einfacher Sprache

Ihr müsst diese Gesetze **nicht auswendig kennen** — aber es hilft zu wissen, woher die
Pflichten kommen. Hier die wichtigsten, in Alltagssprache:

| Gesetz / Regel                                | Was es einfach gesagt verlangt                                                                                                                                                                                                                                     | Betrifft euch als            |
| --------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ---------------------------- |
| **§ 146a AO** (Abgabenordnung)                | „Jede Kasse muss jeden Vorgang einzeln und manipulationssicher (per TSE) aufzeichnen." Gilt **auch für gemeinnützige Vereine** und auch bei kurzen Festen.                                                                                                         | Pflicht zur TSE + Anmeldung  |
| **KassenSichV** (Kassensicherungs­verordnung) | Die „Bedienungsanleitung" zu § 146a AO: Sie sagt genau, was die TSE können muss und was auf den Beleg gehört.                                                                                                                                                      | technische Pflichten         |
| **§ 146a Abs. 2 AO + § 6 KassenSichV**        | „Belegausgabepflicht": Für jeden Kassiervorgang muss ein Beleg **erstellt** werden.                                                                                                                                                                                | Belegausgabe                 |
| **§ 146a Abs. 2 Satz 2 AO**                   | **Befreiung möglich** beim „Verkauf an eine Vielzahl nicht bekannter Personen" (= typisches Vereinsfest). Dann müsst ihr den Bon nicht aktiv aushändigen — **aber nur, wenn das Finanzamt es auf Antrag genehmigt**, und der Bon muss trotzdem erstellbar bleiben. | Belegausgabe (Erleichterung) |
| **§ 146a Abs. 4 AO**                          | „Meldepflicht": Jede Kasse muss dem Finanzamt online gemeldet werden (seit 1.1.2025).                                                                                                                                                                              | Anmeldung der Kasse          |
| **GoBD** (BMF-Schreiben)                      | Daten müssen vollständig, nachvollziehbar, unveränderbar und **10 Jahre** aufbewahrt werden.                                                                                                                                                                       | Aufbewahrung & Backups       |
| **DSFinV-K**                                  | Das genormte Datenformat, in dem ihr bei einer Prüfung eure Daten herausgebt.                                                                                                                                                                                      | Datenexport bei Prüfung      |
| **§ 379 AO**                                  | „Steuergefährdung": Wer ohne TSE kassiert, riskiert ein **Bußgeld** (bis 25.000 €).                                                                                                                                                                                | Warnung                      |

> **Gilt das auch für unseren gemeinnützigen Verein?** Ja. Sobald ihr bei einem Fest Speisen
> oder Getränke gegen Geld verkauft, betreibt ihr einen „wirtschaftlichen Geschäftsbetrieb"
> (§ 14, § 64 AO). Die Kassenpflichten gelten **unabhängig von der Gemeinnützigkeit**. Ob auf
> eure Umsätze Steuern anfallen, ist eine **andere** Frage (Freibeträge, Zweckbetrieb nach
> § 67a AO) — das klärt euer Steuerberater.

---

## 5. Eure Pflichten als Verein — Schritt für Schritt

### Schritt 1: TSE einrichten (BYO TSE mit fiskaly)

**Was ist eine TSE?** Die „Technische Sicherheitseinrichtung" ist ein Sicherheitsmodul, das
jeden Kassenvorgang fälschungssicher signiert (wie ein digitales Siegel). Das Gesetz schreibt
sie zwingend vor.

**Warum „BYO TSE" (Bring Your Own TSE)?** jotti läuft auf einem normalen Server (z. B. einem
gemieteten Cloud-Server). Dort kann man **keinen USB-Stick einstecken**. Deshalb nutzt jotti
eine **Cloud-TSE**: Ihr bucht sie als Online-Dienst und gebt jotti die Zugangsschlüssel.
„Bring Your Own" heißt: **Ihr** schließt den Vertrag ab, nicht der Entwickler.

**So geht ihr vor:**

1. **Anbieter wählen.** Empfohlen für den Start: **fiskaly** (Cloud-TSE, einfache Einrichtung
   über das Internet). Andere zertifizierte Anbieter sind ebenfalls möglich.
2. **Vertrag abschließen** und ein TSE-/Kassen-Konto („Client") anlegen. Ihr erhaltet
   **Zugangsdaten** (typischerweise ein API-Key und ein Secret).
3. **Schlüssel in jotti eintragen.** Tragt die Zugangsdaten in die `.env`-Konfigurationsdatei
   eures Servers ein (die genauen Variablennamen stehen in der `.env.example` bzw. der
   technischen Doku). Damit verbindet sich jotti automatisch mit der TSE.
4. **Testen.** Macht einen Testverkauf und prüft, ob auf dem Beleg eine TSE-Signatur,
   eine Transaktionsnummer und ein Zeitstempel erscheinen.

> 🔒 **Schlüssel sind geheim.** Die TSE-Zugangsdaten gehören **nicht** in den Programmcode und
> nicht in ein öffentliches Repository — nur in die `.env`-Datei auf eurem Server.

> _Rechtsgrundlage: § 146a Abs. 1 AO i. V. m. KassenSichV; technischer Standard BSI TR-03153._

### Schritt 2: Kasse beim Finanzamt anmelden (mit jotti-Seriennummer)

Seit dem **1. Januar 2025** muss jede elektronische Kasse dem Finanzamt **online** gemeldet
werden. Das ist eine **eigene** Pflicht — sie ersetzt nicht die TSE und wird nicht durch sie
ersetzt.

**Was ihr braucht:**

- Die **Seriennummer eurer jotti-Kasse.** Da es keine Hardware mit aufgedruckter Nummer gibt,
  erzeugt jotti beim ersten Start automatisch eine eindeutige Nummer (eine „UUID") und zeigt
  sie im **Admin-Bereich** an. Beispiel: `Kassen-ID: 7f3a9d12-84e1-4b2c-9f6a-1234567890ab`.
- Die **Zertifizierungs-ID und Seriennummer eurer TSE** (bekommt ihr von fiskaly).
- **Steuernummer** des Vereins, **Anschrift** der Betriebsstätte, **Anschaffungs-/
  Inbetriebnahmedatum**, Softwarename „jotti".

**So geht ihr vor:**

1. Im **Mein-ELSTER-Portal** ([elster.de](https://www.elster.de)) anmelden.
2. Das Formular „Mitteilung über elektronische Aufzeichnungssysteme" ausfüllen — die oben
   genannten Daten findet ihr **gebündelt im jotti-Admin-Bereich**.
3. Absenden und die **Bestätigung aufbewahren**.

**Fristen:**

| Situation                                     | Frist                                               |
| --------------------------------------------- | --------------------------------------------------- |
| Kasse **ab 1. Juli 2025** in Betrieb genommen | **innerhalb 1 Monat** nach Inbetriebnahme           |
| Kasse **vor 1. Juli 2025** angeschafft        | war bis **31. Juli 2025** zu melden                 |
| Kasse **stillgelegt**                         | **innerhalb 1 Monat** nach Außerbetriebnahme melden |

> **Gute Nachricht:** Die privaten **Smartphones eurer Servicekräfte müsst ihr NICHT melden.**
> Sie gelten nur als „Eingabegeräte" (wie eine Tastatur). Gemeldet wird **nur die eine
> jotti-Instanz** auf eurem Server.
>
> _Rechtsgrundlage: § 146a Abs. 4 AO; AEAO zu § 146a AO (Eingabegeräte)._

> ⚠️ **Seriennummer sichern!** Die Kassen-Seriennummer steckt in eurer Datenbank. Geht die
> Datenbank ohne Backup verloren, müsst ihr die alte Kasse beim Finanzamt **abmelden** und eine
> neue mit neuer Nummer **anmelden**. Deshalb sind Backups (Schritt 4) Pflicht.

### Schritt 3: Belege & Steuersätze

- **Belegausgabe:** Für jeden Kassiervorgang muss ein Beleg **erstellbar** sein. Beim
  Vereinsfest greift meist die **Befreiung von der Aushändigung** („Verkauf an eine Vielzahl
  nicht bekannter Personen", § 146a Abs. 2 Satz 2 AO). **Aber:** Diese Befreiung müsst ihr
  beim Finanzamt **schriftlich beantragen** — sie gilt nicht automatisch. Verlangt ein Gast
  einen Bon, müsst ihr ihn aushändigen.
- **Steuersätze:** Ordnet jedem Produkt im Admin-Bereich den richtigen Steuersatz zu:
  **19 %** (z. B. Getränke), **7 %** (z. B. Speisen) oder **0 % / steuerbefreit**
  (Zweckbetrieb). Der Steuersatz erscheint auf dem Beleg und im Datenexport. Welcher Satz für
  euch gilt, klärt euer Steuerberater.

### Schritt 4: Daten 10 Jahre aufbewahren

- Alle Kassendaten (das Kassenjournal und spätere DSFinV-K-Exporte) müssen **10 Jahre**
  aufbewahrt werden — vollständig, lesbar und unveränderbar.
- Macht **regelmäßige (täglich empfohlene) Backups** eurer Datenbank und bewahrt sie sicher
  auf. Das schützt zugleich eure Kassen-Seriennummer (siehe Schritt 2).
- Sorgt dafür, dass nur berechtigte Personen Zugriff auf den Server und die Daten haben.

> _Rechtsgrundlage: §§ 146, 147 AO; GoBD._

---

## 6. Laufende Pflichten (Checkliste)

**Einmalig, vor dem ersten Einsatz:**

- [ ] Nutzungsvereinbarung mit dem Autor abgeschlossen (siehe [Lizenzmodell](../lizenzmodell.md))
- [ ] TSE bei fiskaly (o. a.) gebucht und Schlüssel in `.env` eingetragen
- [ ] Betreiber-Stammdaten (Vereinsname, Adresse, Steuernummer) im Admin-Bereich gepflegt
- [ ] Produkte mit korrekten Steuersätzen angelegt
- [ ] Kasse über ELSTER beim Finanzamt angemeldet (Seriennummer aus dem Admin-Bereich)
- [ ] Ggf. Antrag auf Befreiung von der Belegausgabe gestellt
- [ ] Backup-Routine eingerichtet

**Laufend / regelmäßig:**

- [ ] Tägliche Datenbank-Backups laufen
- [ ] Nach jedem Veranstaltungstag: Tagesabschluss (Z-Bon) erstellt
- [ ] Daten werden 10 Jahre archiviert
- [ ] Bei Stilllegung der Kasse: Abmeldung beim Finanzamt innerhalb 1 Monat

---

## 7. Glossar

| Begriff                      | Einfache Erklärung                                                                     |
| ---------------------------- | -------------------------------------------------------------------------------------- |
| **AO**                       | Abgabenordnung — das „Grundgesetz" des Steuerrechts.                                   |
| **TSE**                      | Technische Sicherheitseinrichtung — das digitale Siegel-Modul für jeden Kassenvorgang. |
| **Cloud-TSE / BYO TSE**      | Eine TSE als Online-Dienst, die ihr selbst bucht und in jotti eintragt.                |
| **fiskaly**                  | Ein Anbieter solcher Cloud-TSEs.                                                       |
| **KassenSichV**              | Verordnung mit den technischen Detailregeln zur Kasse.                                 |
| **GoBD**                     | Regeln zu Aufbewahrung und Unveränderbarkeit von Daten.                                |
| **DSFinV-K**                 | Genormtes Datenpaket (ZIP mit CSV-Dateien) für die Finanzprüfung.                      |
| **ELSTER**                   | Das Online-Portal der Finanzverwaltung — hier meldet ihr die Kasse an.                 |
| **Z-Bon / Tagesabschluss**   | Der förmliche Abschluss eines Kassen-Tages.                                            |
| **Seriennummer / Kassen-ID** | Die eindeutige Nummer eurer Kasse; bei jotti automatisch erzeugt.                      |
| **Betreiber**                | Diejenigen, die die Kasse tatsächlich verwenden — euer Verein.                         |
