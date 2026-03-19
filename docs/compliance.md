# Compliance-Analyse: Fiskalische Anforderungen an jotti

> **Status:** Analyse & Anforderungsbeschreibung (kein Feature implementiert)
> **Autor:** CTO-Analyse
> **Datum:** 2026-03-19
> **Betrifft:** KassenSichV, TSE, GoBD, Belegausgabepflicht, DSFinV-K, ERiC/ELSTER

---

## Inhaltsverzeichnis

1. [Zusammenfassung](#1-zusammenfassung)
2. [Aktueller Stand im Repository — Fehlerhafte Annahmen](#2-aktueller-stand-im-repository--fehlerhafte-annahmen)
3. [Rechtliche Grundlagen](#3-rechtliche-grundlagen)
4. [Analyse: Warum die Ausnahme nicht greift](#4-analyse-warum-die-ausnahme-nicht-greift)
5. [Anforderung 1: TSE-Integration (Technische Sicherheitseinrichtung)](#5-anforderung-1-tse-integration-technische-sicherheitseinrichtung)
6. [Anforderung 2: GoBD-Konformität](#6-anforderung-2-gobd-konformität)
7. [Anforderung 3: Belegausgabepflicht](#7-anforderung-3-belegausgabepflicht)
8. [Anforderung 4: DSFinV-K Export-Schnittstelle](#8-anforderung-4-dsfinv-k-export-schnittstelle)
9. [Anforderung 5: Elektronische Meldepflicht (ERiC / ELSTER)](#9-anforderung-5-elektronische-meldepflicht-eric--elster)
10. [Architektonische Lösungsansätze](#10-architektonische-lösungsansätze)
11. [Handlungsempfehlungen und Priorisierung](#11-handlungsempfehlungen-und-priorisierung)
12. [Quellenverzeichnis](#12-quellenverzeichnis)

---

## 1. Zusammenfassung

jotti wird aktuell im Repository als Kassensystem positioniert, das **keine fiskalischen Anforderungen** (KassenSichV, TSE) erfüllen muss, weil es für gemeinnützige Vereine und temporäre Veranstaltungen konzipiert ist. **Diese Annahme ist falsch.**

Die Pflicht zur Nutzung einer zertifizierten TSE (Technische Sicherheitseinrichtung) nach § 146a AO i.V.m. der KassenSichV knüpft **nicht** an die Rechtsform des Betreibers (e.V., gGmbH) oder die Gemeinnützigkeit an, sondern an die **Verwendung eines elektronischen Aufzeichnungssystems**. Sobald ein Verein eine Software wie jotti einsetzt, um Bareinnahmen aufzuzeichnen, unterliegt er denselben Pflichten wie ein kommerzieller Gastronomiebetrieb.

### Kernbefunde

| Bereich | Aktueller Stand | Bewertung |
| --- | --- | --- |
| **TSE-Integration** | ❌ Nicht vorhanden | **Pflicht** bei Einsatz eines elektronischen Aufzeichnungssystems |
| **GoBD-Konformität** | ✅ Teilweise erfüllt (Event-Sourcing) | Append-Only-Architektur gut, aber kryptografische Verkettung fehlt |
| **Belegausgabepflicht** | ❌ Nicht vorhanden | **Pflicht** — Beleg muss bei jedem Kassiervorgang erzeugt werden können |
| **DSFinV-K Export** | ❌ Nicht vorhanden | **Pflicht** — maschinenlesbarer Datenexport für Finanzverwaltung |
| **ERiC / ELSTER Meldung** | ❌ Nicht vorhanden | **Pflicht** ab 2025 — elektronische Kassenmeldung beim Finanzamt |
| **Dokumentation im Repo** | ❌ Fehlerhafte Aussagen | Mehrere Stellen behaupten fälschlich, Vereine seien von TSE/KassenSichV befreit |

---

## 2. Aktueller Stand im Repository — Fehlerhafte Annahmen

An folgenden Stellen im Repository werden **falsche oder irreführende Aussagen** zur Anwendbarkeit fiskalischer Vorschriften getroffen:

### 2.1 README.md

> ✅ Geeignet für: Bargeld-Betrieb **ohne Kassenpflicht nach KassenSichV**

> ❌ Nicht geeignet für: Kommerzielle Gastro-Betriebe **mit TSE-Pflicht**

**Problem:** Die Formulierung suggeriert, dass es Bargeld-Betriebe gibt, für die die KassenSichV nicht gilt, und dass die TSE-Pflicht nur für „kommerzielle" Betriebe relevant ist. Das ist falsch — die TSE-Pflicht knüpft an das verwendete Aufzeichnungssystem, nicht an die Rechtsform.

### 2.2 docs/anforderungen.md (§ 6 Bewusste Abgrenzung)

> 🚫 TSE / KassenSichV — Gemeinnützige Vereine unterliegen **in der Regel keiner TSE-Pflicht**. Event-Sourcing erfüllt die GoBD-Grundsätze bereits.

**Problem:** Die Aussage „in der Regel keiner TSE-Pflicht" ist falsch. Gemeinnützige Vereine, die ein elektronisches Aufzeichnungssystem einsetzen, unterliegen derselben TSE-Pflicht wie jeder andere Nutzer eines solchen Systems. Die einzige Möglichkeit, der TSE-Pflicht zu entgehen, wäre der vollständige Verzicht auf ein elektronisches Aufzeichnungssystem (z.B. Nutzung einer offenen Ladenkasse mit handschriftlicher Aufzeichnung).

### 2.3 docs/produktbeschreibung.md (§ 7.3 Fiskalkonformität im Detail)

> Für gemeinnützige Vereine, die **keine Kassenpflicht nach KassenSichV haben**, bietet jotti damit ein hohes Maß an Transparenz und Nachvollziehbarkeit — ohne den Overhead einer zertifizierten TSE.

**Problem:** Die Aussage impliziert, dass gemeinnützige Vereine von der KassenSichV befreit sind. Das trifft nicht zu, sobald ein elektronisches Aufzeichnungssystem verwendet wird.

### 2.4 docs/handbuch.md (§ 1.3 Bewusste Abgrenzung)

> Zertifizierte TSE (KassenSichV) — [bewusst nicht enthalten]

**Problem:** Die TSE wird als optionales Feature dargestellt, das „bewusst nicht enthalten" ist. In Wirklichkeit handelt es sich um eine gesetzliche Pflicht, deren Nichteinhaltung einen Rechtsverstoß darstellt.

### 2.5 AGENTS.md

> Bewusst NICHT enthalten: Kartenzahlung, **TSE/KassenSichV**, Reservierungen [...]

**Problem:** Dieselbe fehlerhafte Einordnung wie oben — TSE wird als optionale Funktionalität eingestuft.

### 2.6 Bewertung

Alle genannten Stellen gehen von einer **fehlerhaften Prämisse** aus: dass die Gemeinnützigkeit oder der temporäre Charakter einer Veranstaltung eine Befreiung von der TSE-Pflicht begründet. Das ist rechtlich nicht zutreffend. Die TSE-Pflicht entsteht durch die **Nutzung eines elektronischen Aufzeichnungssystems** — und jotti ist genau das.

---

## 3. Rechtliche Grundlagen

### 3.1 Abgabenordnung (AO) — § 146a

§ 146a AO verpflichtet jeden, der ein **elektronisches Aufzeichnungssystem** im Sinne des § 1 KassenSichV verwendet, dieses mit einer **zertifizierten technischen Sicherheitseinrichtung (TSE)** zu schützen. Die Norm differenziert nicht nach:

- Rechtsform (e.V., GmbH, Einzelunternehmen)
- Gemeinnützigkeit
- Dauer des Betriebs (temporär oder dauerhaft)
- Gewinnerzielungsabsicht

> **§ 146a Abs. 1 AO:** „Wer aufzeichnungspflichtige Geschäftsvorfälle oder andere Vorgänge mit Hilfe eines elektronischen Aufzeichnungssystems erfasst, hat ein elektronisches Aufzeichnungssystem zu verwenden, das jeden aufzeichnungspflichtigen Geschäftsvorfall und anderen Vorgang einzeln, vollständig, richtig, zeitgerecht und geordnet aufzeichnet. Das elektronische Aufzeichnungssystem und die digitalen Aufzeichnungen sind durch eine zertifizierte technische Sicherheitseinrichtung zu schützen."

*(Quelle: § 146a AO — https://www.gesetze-im-internet.de/ao_1977/__146a.html)*

### 3.2 Kassensicherungsverordnung (KassenSichV)

Die KassenSichV konkretisiert die Anforderungen aus § 146a AO. § 1 KassenSichV definiert den Anwendungsbereich:

> **§ 1 KassenSichV:** Elektronische Aufzeichnungssysteme im Sinne des § 146a Absatz 1 Satz 1 der Abgabenordnung sind elektronische oder computergestützte Kassensysteme oder Registrierkassen [...].

jotti ist ein „computergestütztes Kassensystem" im Sinne dieser Definition. Es erfasst Geschäftsvorfälle elektronisch (Bestellungen, Zahlungen, Stornierungen) und berechnet Zahlungsbeträge.

Für **mobile Geräte** (Smartphones als Browser-Clients in jottis BYOD-Modell) gilt: Wenn ein Gerät technisch in der Lage ist, Zahlungsvorgänge eigenständig zu erfassen und offline zu betreiben, muss es selbst an eine TSE angebunden sein. Fungiert es ausschließlich als Eingabeterminal, das sofort an ein TSE-gesichertes Backend weiterleitet, genügt die Backend-seitige TSE-Anbindung. Entscheidend ist die technische Fähigkeit zur selbständigen Offline-Erfassung, nicht die tatsächliche Nutzung.

*(Quelle: KassenSichV — https://www.gesetze-im-internet.de/kassensichv/BJNR351500017.html)*
*(Quelle: BMF-FAQ zu § 146a AO, Frage zur Abgrenzung von Eingabegeräten und eigenständigen Aufzeichnungssystemen — https://www.bundesfinanzministerium.de/)*

### 3.3 GoBD (Grundsätze zur ordnungsmäßigen Führung und Aufbewahrung von Büchern, Aufzeichnungen und Unterlagen in elektronischer Form)

Die GoBD gelten für **alle Steuerpflichtigen**, unabhängig von der Rechtsform. Sie regeln:

- Nachvollziehbarkeit und Nachprüfbarkeit
- Vollständigkeit
- Richtigkeit
- Zeitgerechte Buchungen und Aufzeichnungen
- Ordnung
- Unveränderbarkeit

Auch gemeinnützige Vereine sind steuerpflichtig (z.B. im wirtschaftlichen Geschäftsbetrieb bei Vereinsfesten, die nicht unter § 67a AO fallen) und müssen die GoBD einhalten.

*(Quelle: BMF-Schreiben vom 28.11.2019 — GoBD — https://www.bundesfinanzministerium.de/Content/DE/Downloads/BMF_Schreiben/Weitere_Steuerthemen/Abgabenordnung/2019-11-28-GoBD.html)*

### 3.4 Belegausgabepflicht (§ 146a Abs. 2 AO, § 6 KassenSichV)

> **§ 146a Abs. 2 AO:** „Wer [...] ein elektronisches Aufzeichnungssystem [...] verwendet, hat dem an diesem Geschäftsvorfall Beteiligten in unmittelbarem zeitlichen Zusammenhang mit dem Geschäftsvorfall unbeschadet anderer gesetzlicher Vorschriften einen Beleg über den Geschäftsvorfall auszustellen und zur Verfügung zu stellen."

Die Belegausgabepflicht gilt für **jeden** Nutzer eines elektronischen Aufzeichnungssystems. Der Beleg kann in Papierform oder — mit Zustimmung des Kunden — in elektronischer Form ausgegeben werden.

*(Quelle: § 146a Abs. 2 AO — https://www.gesetze-im-internet.de/ao_1977/__146a.html)*

### 3.5 DSFinV-K (Digitale Schnittstelle der Finanzverwaltung für Kassensysteme)

§ 4 KassenSichV verlangt eine **einheitliche digitale Schnittstelle**, über die die gespeicherten Daten für die Finanzverwaltung exportiert werden können. Die DSFinV-K (aktuell Version 2.4, Stand Januar 2024) definiert das genaue Format dieses Exports: eine Sammlung von CSV-Dateien mit fest vorgeschriebenen deutschen Dateinamen, Spaltenreihenfolge und Semikolon-Trennung, verpackt in einem ZIP-Archiv mit `index.xml`.

*(Quelle: BZSt — DSFinV-K — https://www.bzst.de/DE/Unternehmen/Aussenpruefungen/DigitaleSchnittstelleFinV/digitaleschnittstellefinv_node.html)*

### 3.6 Elektronische Kassenmeldepflicht (§ 146a Abs. 4 AO)

Ab dem 1. Januar 2025 müssen elektronische Aufzeichnungssysteme dem zuständigen Finanzamt **elektronisch** gemeldet werden. Die Meldung erfolgt über das ELSTER-System, wahlweise direkt im Portal oder programmatisch über die ERiC-Schnittstelle (ELSTER Rich Client). Genauere Fristen: siehe Abschnitt 9.

*(Quelle: § 146a Abs. 4 AO — https://www.gesetze-im-internet.de/ao_1977/__146a.html)*
*(Quelle: BMF-FAQ — https://www.bundesfinanzministerium.de/)*

---

## 4. Analyse: Warum die Ausnahme nicht greift

### 4.1 Der Irrtum: „Gemeinnützige Vereine sind befreit"

Die im Repository verbreitete Annahme lässt sich so zusammenfassen:

> „Gemeinnützige Vereine, die temporäre Veranstaltungen durchführen, unterliegen nicht der KassenSichV und brauchen keine TSE."

Diese Annahme ist aus folgenden Gründen **falsch**:

#### a) Die KassenSichV knüpft an das Aufzeichnungssystem, nicht an den Betreiber

Die Pflicht zur Verwendung einer TSE entsteht durch die **Nutzung eines elektronischen Aufzeichnungssystems** (§ 146a Abs. 1 AO). Es gibt keine Befreiung für:

- Gemeinnützige Organisationen
- Temporäre Veranstaltungen
- Vereinsfeste
- Nicht-kommerzielle Zwecke

Die **einzige** Möglichkeit, der TSE-Pflicht zu entgehen, besteht darin, **kein elektronisches Aufzeichnungssystem** zu verwenden — also z.B. eine offene Ladenkasse mit handschriftlichen Aufzeichnungen (Kassenbuch) zu führen.

#### b) Vereine haben wirtschaftliche Geschäftsbetriebe

Auch gemeinnützige Vereine können **wirtschaftliche Geschäftsbetriebe** unterhalten (§ 14 AO). Der Verkauf von Speisen und Getränken auf einem Vereinsfest ist in der Regel ein wirtschaftlicher Geschäftsbetrieb, sofern:

- Die Einnahmen die Besteuerungsgrenze überschreiten (§ 64 Abs. 3 AO: 45.000 € brutto pro Jahr), oder
- Die Veranstaltung nicht als Zweckbetrieb nach § 67a AO eingestuft wird (z.B. weil bezahlte Arbeitskräfte eingesetzt werden oder die Veranstaltung nicht der Satzung dient).

In vielen Fällen sind Vereinsfeste als **wirtschaftlicher Geschäftsbetrieb** einzuordnen, für den volle Aufzeichnungspflichten gelten.

#### c) Aufzeichnungspflicht besteht unabhängig von der Steuerpflicht

Selbst wenn ein Verein keine Umsatzsteuer abführen muss (Kleinunternehmerregelung, § 19 UStG), besteht die **Einzelaufzeichnungspflicht** nach § 146 Abs. 1 AO für alle Geschäftsvorfälle. Die KassenSichV gilt zusätzlich, sobald ein elektronisches Aufzeichnungssystem eingesetzt wird.

### 4.2 Die korrekte Aussage

> Wer jotti (oder jedes andere elektronische Kassensystem) einsetzt, um Bareinnahmen bei einer Vereinsveranstaltung aufzuzeichnen, unterliegt der Pflicht zur Verwendung einer zertifizierten TSE gemäß § 146a AO i.V.m. KassenSichV — unabhängig von Rechtsform, Gemeinnützigkeit oder Dauer der Veranstaltung.

### 4.3 Die einzige echte Ausnahme

Eine Befreiung gibt es nur, wenn **kein elektronisches Aufzeichnungssystem** verwendet wird. Vereine können also:

1. **Ohne Software:** Eine offene Ladenkasse mit handschriftlichen Aufzeichnungen (Kassenbuch) führen — dann gilt die KassenSichV nicht.
2. **Mit Software (wie jotti):** Dann gelten alle Pflichten der KassenSichV, einschließlich TSE, Belegausgabe und DSFinV-K.

---

## 5. Anforderung 1: TSE-Integration (Technische Sicherheitseinrichtung)

### 5.1 Hintergrund

Die TSE ist das kryptografische Herzstück eines konformen Kassensystems. Sie besteht aus drei Modulen:

1. **Sicherheitsmodul:** Aufgeteilt in SMAERS (Datenaufbereitung) und CSP (kryptografische Signatur)
2. **Speichermedium:** Lokale Sicherung der signierten Transaktionsdaten
3. **Einheitliche Digitale Schnittstelle (EDS):** Standardisierte Exportschnittstelle

*(Quelle: BSI TR-03153 — https://www.bsi.bund.de/SharedDocs/Downloads/DE/BSI/Publikationen/TechnischeRichtlinien/TR03153/TR-03153.pdf?__blob=publicationFile)*

### 5.2 Protokollierungs-Ablauf (Transaktions-Lebenszyklus)

Jeder abzusichernde Vorgang kommuniziert mit der TSE über drei Operationen:

#### Phase 1: `StartTransaction`

Sobald ein neuer Vorgang beginnt (z.B. eine Bestellung aufgenommen wird oder ein Kassenbon geöffnet wird), muss **unmittelbar** `StartTransaction` aufgerufen werden. „Unmittelbar" ist eine Ordnungsanforderung — es gibt keine maximale Transaktionsdauer in BSI TR-03153.

- **Request-Payload:** Seriennummer des Aufzeichnungssystems (Kassen-ID), Art des Vorgangs (`processType`), initiale Vorgangsdaten (`processData`)
- **Response:** Eindeutige fortlaufende Transaktionsnummer, Zeitpunkt des Vorgangsbeginns (`logTime`), Seriennummer der TSE, aktueller Signaturzähler

#### Phase 2: `UpdateTransaction` (optional, nur für bestimmte processTypes)

`UpdateTransaction` darf ausschließlich für die `processType`-Werte `Bestellung-V1` und `SonstigerVorgang-V1` verwendet werden. **Für `Kassenbeleg-V1` ist `UpdateTransaction` ausdrücklich verboten**, da die `processData` beim Kassenbeleg erst mit dem Abschluss des Vorgangs bekannt ist (Quelle: BMF-FAQ zu § 146a AO, Fragen zu Transaktionsabläufen beim processType „Kassenbeleg").

- **Request-Payload:** Kassen-ID, Transaktionsnummer aus Phase 1, aktualisierte `processData`
- **Verwendungsfall:** Stufenweises Hinzufügen von Bestellpositionen zu einem noch laufenden `Bestellung-V1`-Vorgang

#### Phase 3: `FinishTransaction`

Wird der Vorgang abgeschlossen (Zahlung, Ausgabe von Beleg) oder abgebrochen (Storno vor Abschluss), muss die Transaktion beendet werden.

- **Request-Payload:** Transaktionsnummer, finale `processData` (Summen aufgeteilt nach Steuersätzen und Zahlungsarten)
- **Response:** Kryptografische Signatur (Prüfwert), Zeitpunkt der Vorgangsbeendigung, finaler Signaturzähler

### 5.3 Offizielle processType-Werte

Die `processType`-Werte sind im AEAO zu § 146a AO, Anhang I, festgelegt und in DSFinV-K Anhang I referenziert. Die **-V1-Endung ist Bestandteil des offiziellen Strings** und muss exakt so an die TSE übergeben werden.

| processType | Verwendung |
| --- | --- |
| `Kassenbeleg-V1` | Zahlungsbeleg (Rechnung), der dem Kunden ausgehändigt wird |
| `Bestellung-V1` | Zwischenabsicherung einer Bestellung ohne sofortige Zahlung (Gastronomie) |
| `SonstigerVorgang-V1` | Alle anderen abzusichernden Vorgänge (Tagesabschluss, TSE-Selbsttest, ...) |

### 5.4 Datenformat-Vorgaben (`processData`)

Die Formatierung der `processData` ist streng reguliert:

- **Encoding:** UTF-8 oder ASCII, kein BOM (Byte-Order-Mark)
- **Dezimaltrennzeichen:** Ausschließlich Punkt (`.`)
- **Verboten:** Tausendertrennzeichen, Exponentialschreibweise, `+` vor positiven Werten
- **Mindestens eine Stelle vor dem Dezimaltrennzeichen:** `0.5` statt `.5`
- **Format-String für `Kassenbeleg-V1`:** `Beleg^<Betrag_Normal>_<Betrag_Ermaessigt>_<Betrag_Null>_<Betrag_Besonderer_Satz>_<Betrag_Befreit>^<Zahlbetrag>:<Zahlungsart>`
- **Format-String für `Bestellung-V1`:** Positionen als strukturierter Text (z.B. `4x Maß Bier_2x Weißwurst`) — genaues Format gemäß AEAO § 146a Anhang I

### 5.5 TSE-Varianten

Es gibt zwei grundlegende Integrationsansätze:

| Variante | Beschreibung | Beispiel-Anbieter |
| --- | --- | --- |
| **Hardware-TSE** | Physisches Gerät (USB-Stick, SD-Karte, Smartcard) | Swissbit, Epson, Diebold Nixdorf |
| **Cloud-TSE** | TSE als Cloud-Service, Kommunikation via HTTPS-API | fiskaly, Deutsche Fiskal |

Für jotti als Self-hosted-System wäre eine **Cloud-TSE** naheliegend, da sie keine zusätzliche Hardware erfordert und über HTTP-POST-Requests angesprochen wird. Eine Hardware-TSE scheidet für BYOD-Smartphone-Setups im Festzelt praktisch aus.

### 5.6 Das Festzelt-Muster: Atomare TSE-Transaktionen (Empfehlung)

#### Das Problem langer Tischvorgänge

Im Festzelt-Betrieb (Oktoberfest, Vereinsfest, Maihock) sitzen an Tisch 42 über den ganzen Tag wechselnde Gästegruppen. Bestellt wird in mehreren Runden, gezahlt wird in Teilbeträgen zu verschiedenen Zeitpunkten. Eine einzige TSE-Transaktion über den gesamten Tischvorgang offenzuhalten wäre technisch riskant (Timeouts, Systemfehler) und laut BMF-FAQ auch nicht erforderlich.

#### Die offizielle Lösung: Zwei-Schichten-Ansatz (DSFinV-K Nr. 2.7)

Die Finanzverwaltung hat in DSFinV-K Nr. 2.7 und Anhang H eine **Vereinfachungsregelung für langanhaltende Bestellvorgänge** dokumentiert (oft auch „Erleichterungsregelung beim Durchbedienen" genannt). Der Kern: Jede atomare Aktion (Bestellung, Teilzahlung, Storno) ist eine **eigene, sofort geschlossene** TSE-Transaktion.

**Schicht 1 — Bestellabsicherung** (für jede Bestellrunde):
```
StartTransaction(processType="Bestellung-V1") → sofort → FinishTransaction
```
Jede Bestellrunde wird sofort als vollständige `Bestellung-V1`-Transaktion abgesichert und geschlossen. Die `processData` enthält die bestellten Positionen.

**Schicht 2 — Zahlungsabsicherung** (bei jeder Teil- oder Vollzahlung):
```
StartTransaction(processType="Kassenbeleg-V1") → sofort → FinishTransaction
```
Die `processData` enthält Gesamtbetrag, Steuersätze und Zahlungsart. Diese Transaktion ist ebenfalls sofort geschlossen.

**Verknüpfung**: Alle Bestellungen und Zahlungen eines Tisches werden im DSFinV-K-Export über das Feld `ABRECHNUNGSKREIS` (z.B. `"Tisch-42"`) zusammengeführt (siehe Abschnitt 8).

#### Konkretes Szenario: Maihock, Tisch 42

```
18:00 — Gruppe A setzt sich (4 Personen)
18:01 — Bestellung: 4x Maß Bier, 2x Weißwurst
        → StartTransaction(Bestellung-V1) + FinishTransaction
          → TSE-Signatur S1, transactionNumber=1001
        → ABRECHNUNGSKREIS = "Tisch-42-20260501"

19:30 — Bestellung: 4x Maß Bier (Nachbestellung)
        → StartTransaction(Bestellung-V1) + FinishTransaction
          → TSE-Signatur S2, transactionNumber=1002
        → ABRECHNUNGSKREIS = "Tisch-42-20260501"

20:00 — 2 Gäste zahlen: 2x Bier = 14,00 €, bar
        → StartTransaction(Kassenbeleg-V1) + FinishTransaction
          → TSE-Signatur S3, transactionNumber=1003
          → Bon: enthält zusätzlich "Erste Bestellung: 18:01 Uhr" (s. §7)
        → ABRECHNUNGSKREIS = "Tisch-42-20260501"

21:00 — Restliche 2 Gäste zahlen: 2x Bier + 2x Weißwurst = 18,00 €, bar
        → StartTransaction(Kassenbeleg-V1) + FinishTransaction
          → TSE-Signatur S4, transactionNumber=1004
          → Bon: enthält zusätzlich "Erste Bestellung: 18:01 Uhr"
        → ABRECHNUNGSKREIS = "Tisch-42-20260501"

22:00 — Gruppe B setzt sich (neue Gäste, neuer ABRECHNUNGSKREIS)
        → ABRECHNUNGSKREIS = "Tisch-42-20260501-B"
```

Der Betriebsprüfer sieht im DSFinV-K-Export alle vier Transaktionen mit demselben `ABRECHNUNGSKREIS` und kann den vollständigen Tischverlauf nachvollziehen, obwohl jede Transaktion sofort geschlossen wurde.

### 5.7 Architektonische Anforderungen an jotti

1. **TSE-Abstraktionsschicht:** Ein Interface `TSEClient` im Backend, das die drei Phasen (`StartTransaction`, `UpdateTransaction`, `FinishTransaction`) abstrahiert
2. **Atomares Transaktionsmodell:** Für jeden jotti-Vorgang (Bestellung, Zahlung, Storno) wird eine eigenständige, sofort geschlossene TSE-Transaktion erstellt
3. **Signatur-Speicherung:** TSE-Rückgabewerte (Transaktionsnummer, Signaturzähler, Prüfwert, `logTime`) müssen als Event-Daten persistiert werden
4. **Fehlerbehandlung:** Verhalten bei TSE-Nicht-Erreichbarkeit (Cloud-TSE offline, Timeout) — z.B. Offline-Queue mit späterer Nachsignierung
5. **Konfiguration:** TSE-Anbieter, Kassen-ID, API-Credentials als Umgebungsvariablen
6. **ABRECHNUNGSKREIS-Verwaltung:** Eindeutige, persistente Tisch-Session-ID für die DSFinV-K-Verknüpfung

---

## 6. Anforderung 2: GoBD-Konformität

### 6.1 Aktueller Stand

jotti erfüllt durch die Event-Sourcing-Architektur bereits mehrere GoBD-Grundsätze:

| GoBD-Grundsatz | Aktueller Status | Anmerkung |
| --- | --- | --- |
| Unveränderbarkeit | ✅ Erfüllt | Events sind append-only, kein UPDATE/DELETE |
| Nachvollziehbarkeit | ✅ Erfüllt | Lückenloses Kassenjournal pro Tisch |
| Vollständigkeit | ✅ Erfüllt | Jeder Geschäftsvorfall wird als Event erfasst |
| Zeitgerechte Buchung | ✅ Erfüllt | Events mit Echtzeit-Zeitstempel |
| Ordnungsmäßigkeit | ✅ Erfüllt | Strukturiertes Datenmodell, typisierte Events |
| Kryptografische Verkettung | ❌ Fehlt | Keine TSE-Signatur, keine kryptografische Absicherung |
| 10-Jahres-Aufbewahrung | ⚠️ Nicht adressiert | Keine Archivierungsstrategie implementiert |

### 6.2 Anforderungen gemäß §§ 146, 147 AO und GoBD

- **Aufbewahrungspflicht:** Alle steuerlich relevanten Daten müssen **10 Jahre** aufbewahrt werden, jederzeit verfügbar, unverzüglich lesbar, vollständig und absolut unveränderbar.
- **Elektronisches Radierverbot:** Datensätze dürfen nach der Erfassung nicht per `UPDATE` oder `DELETE` überschrieben oder gelöscht werden.
- **Stornierungen:** Müssen als neue Buchungssätze (mit neuem Zeitstempel und neuer TSE-Signatur) erzeugt werden, die den alten Wert ausgleichen — niemals als nachträgliche Änderung.
- **Verfahrensdokumentation:** Es muss dokumentiert sein, wie das System Daten erzeugt, verarbeitet und archiviert.

*(Quelle: GoBD — BMF-Schreiben vom 28.11.2019 — https://www.bundesfinanzministerium.de/Content/DE/Downloads/BMF_Schreiben/Weitere_Steuerthemen/Abgabenordnung/2019-11-28-GoBD.html)*

### 6.3 Handlungsbedarf

| Maßnahme | Priorität | Aufwand |
| --- | --- | --- |
| TSE-Signatur in Event-Daten integrieren | Hoch | Mittel |
| Archivierungsstrategie definieren (10 Jahre) | Mittel | Gering |
| Verfahrensdokumentation erstellen | Mittel | Gering |
| Soft-Delete-Praxis bei Stammdaten prüfen | Niedrig | Gering |

---

## 7. Anforderung 3: Belegausgabepflicht

### 7.1 Gesetzliche Grundlage

Gemäß § 146a Abs. 2 AO und § 6 KassenSichV muss für **jeden Kassiervorgang** ein Beleg erzeugt und dem Kunden zur Verfügung gestellt werden. Der Beleg kann gedruckt oder — mit Zustimmung des Kunden — elektronisch (z.B. per QR-Code oder PDF) bereitgestellt werden.

*(Quelle: KassenSichV § 6 — https://www.gesetze-im-internet.de/kassensichv/BJNR351500017.html)*

### 7.2 Pflichtangaben auf dem Beleg

Ein konformer Beleg muss mindestens folgende Angaben enthalten:

**Standard-Kassendaten:**
- Vollständiger Name und Anschrift des leistenden Unternehmens (Verein)
- Datum der Belegausgabe
- Menge und Art der gelieferten Gegenstände / Umfang der Dienstleistung
- Entgelt und darauf entfallender Steuerbetrag, oder Hinweis auf Steuerbefreiung
- Transaktionsnummer (Bonnummer)

**TSE-Pflichtdaten (§ 6 KassenSichV):**
- Zeitpunkt des Vorgangsbeginns (von TSE `StartTransaction`, entspricht dem Start der `Kassenbeleg-V1`-Transaktion)
- Zeitpunkt der Vorgangsbeendigung (von TSE `FinishTransaction`)
- Seriennummer des elektronischen Aufzeichnungssystems (Kassen-ID)
- Seriennummer der TSE
- TSE-Transaktionsnummer
- Signaturzähler
- Kryptografischer Prüfwert (TSE-Signatur)

### 7.3 Besondere Anforderung beim Festzelt-Muster (Durchbedienen)

Wenn die Tisch-Bestellungen mit `Bestellung-V1`-Transaktionen abgesichert wurden und erst später bezahlt wird (atomares Transaktionsmodell gemäß Abschnitt 5.6), gilt laut BMF-FAQ:

> „Zusätzlich ist auf den Bon der **Startzeitpunkt der ersten Bestellung in Klarschrift aufzudrucken**."
> *(Quelle: BMF-FAQ zu § 146a AO; DSFinV-K Nr. 2.7 sowie Anhang H)*

Der Zahlungsbeleg muss also **zwei Zeitstempel** enthalten:
1. Den TSE-`logTime` der aktuellen `Kassenbeleg-V1`-Transaktion (§ 6 KassenSichV-Pflicht)
2. Den `logTime` der **allerersten `Bestellung-V1`-Transaktion** für diesen Tisch/Session in Klarschrift (DSFinV-K-Pflicht beim Durchbedienen)

**Beispiel** (Tisch 42, Maihock):
```
Volksverein Musterstadt e.V.
Vereinsfest Maihock 2026
Tisch: 42
Erste Bestellung: 01.05.2026, 18:01 Uhr   ← Pflichtfeld beim Durchbedienen
Bon-Nr.: 1003
---
2x Maß Bier        14,00 €
---
Gesamt:            14,00 €
Bar erhalten:      14,00 €
---
TSE-Start: 01.05.2026, 20:00:12 Uhr
TSE-Ende:  01.05.2026, 20:00:14 Uhr
TSE-ID: SW-TSE-SN-0042
TSE-Nr.: 1003, Signatur-Zähler: 5871
[QR-Code mit TSE-Daten]
```

### 7.4 QR-Code-Format

Um Platz auf dem Beleg zu sparen, können die TSE-Daten in einen standardisierten **QR-Code** verpackt werden. Das Format muss den Vorgaben der DSFinV-K (Anhang I) entsprechen.

### 7.5 Architektonische Anforderungen an jotti

1. **Beleg-Generator:** Komponente, die aus einem abgeschlossenen Kassiervorgang einen konformen Beleg (PDF oder Druckformat) erzeugt
2. **TSE-Daten auf dem Beleg:** Alle TSE-Rückgabewerte müssen auf dem Beleg erscheinen
3. **Erste-Bestellung-Zeitstempel:** Das Backend muss den `logTime` der ersten `Bestellung-V1`-Transaktion einer Tisch-Session persistieren und beim Beleg-Druck abrufen
4. **QR-Code-Generierung:** TSE-Daten als QR-Code im DSFinV-K-Format
5. **Beleg-Ausgabekanal:** Druck (über bestehende Bondrucker-Anbindung) und/oder elektronisch (QR-Code auf dem Smartphone der Servicekraft)
6. **Beleg-Archivierung:** Belegdaten müssen für den DSFinV-K-Export persistiert werden

---

## 8. Anforderung 4: DSFinV-K Export-Schnittstelle

### 8.1 Übersicht

Die Finanzverwaltung verlangt bei einer Kassen-Nachschau oder Betriebsprüfung einen genormten, maschinenlesbaren Datenexport. Dieser Export folgt der **DSFinV-K-Spezifikation** (Version 2.4, Stand Januar 2024) und muss von der Prüfsoftware IDEA der Finanzämter gelesen werden können.

*(Quelle: BZSt — https://www.bzst.de/DE/Unternehmen/Aussenpruefungen/DigitaleSchnittstelleFinV/digitaleschnittstellefinv_node.html)*

### 8.2 Dateiformat und Grundregeln

- **Gesamtstruktur:** ZIP-Archiv mit CSV-Dateien und einer `index.xml`
- **`index.xml`:** Metadaten-Datei, die dem Prüftool mitteilt, wie die CSVs zu parsen sind
- **Kopfzeile:** Jede CSV-Datei beginnt zwingend mit einem Header-Datensatz
- **Trennzeichen:** Semikolon (`;`) als Feldtrennzeichen
- **Zeilenumbrüche:** CRLF (`\r\n`)
- **Zahlenformate:** Keine Tausendertrennzeichen, Punkt als Dezimaltrennzeichen, mindestens eine Stelle vor dem Dezimaltrennzeichen (`0.5`, nicht `.5`), keine führenden Nullen
- **Spaltenreihenfolge:** Exakt wie in der Spezifikation vorgegeben
- **Dateinamen:** Die Dateinamen sind in der DSFinV-K-Spezifikation auf Deutsch vorgegeben (z.B. `Bonkopf.csv`, `Bonpos.csv`) und dürfen nicht abgeändert werden
- **Custom-Felder:** Zusätzliche Spalten am Ende erlaubt, müssen aber in `index.xml` definiert werden

> **Achtung:** Manche TSE-Anbieter (z.B. fiskaly/fiskaltrust) stellen interne Export-Zwischenschichten mit abweichenden Dateinamen bereit (z.B. `transactions.csv`, `allocation_groups.csv`, `references.csv`). Diese sind **kein offizieller DSFinV-K-Standard** und müssen vor der Abgabe an das Finanzamt in das offizielle Format konvertiert werden.

### 8.3 Modul-Struktur und offizielle Dateinamen

Der Export gliedert sich in drei Module:

#### A. Stammdatenmodul (Master Data)

| Datei | Inhalt |
| --- | --- |
| `Stamm_Abschluss.csv` | Metadaten zum Z-Bon (Kassenabschluss): Unternehmensname, Steuernummer, Start-/End-Zeitpunkt |
| `Stamm_Orte.csv` | Standortdaten der Betriebsstätte |
| `Stamm_Kassen.csv` | Kassen-Hardware: Hersteller, Seriennummer, Software-Typ und -Version |
| `Stamm_TSE.csv` | TSE-Daten: Zertifikats-ID, Signaturalgorithmus (z.B. `ecdsa-plain-SHA256`), TSE-Seriennummer (64-stelliger Hexadezimalstring, 0–9 und A–F), Public Key (Base64-kodiert) |
| `Z_GV_Typ.csv` | Aggregierte Beträge pro Geschäftsvorfall-Typ nach Steuersätzen |
| `Z_Zahlart.csv` | Aggregierte Summen der Zahlungsarten (Bar vs. unbar) |

#### B. Einzelaufzeichnungsmodul (Transactions / Lines)

| Datei | Inhalt |
| --- | --- |
| `Bonkopf.csv` | Ein Datensatz pro Kassenbon: `BON_ID`, `BON_NR`, `BON_TYP`, `BON_START`, `BON_ENDE` (ISO 8601), Gesamtbruttoumsatz, `ABRECHNUNGSKREIS`, `BON_STORNO`, `REF_BON_ID` |
| `Bonkopf_USt.csv` | USt-Aufschlüsselung pro Bon nach Steuerschlüsseln (Brutto, Netto, USt) |
| `Bonkopf_Zahlarten.csv` | Zahlungsarten pro Bon (Bar, EC-Karte, Kreditkarte) |
| `Bonpos.csv` | Einzelne Artikel: `POS_ZEILE`, `ART_NR`, `MENGE` (Dezimal, 3 Nachstellen), `EINHEIT`, `STK_BR` (Stückpreis brutto) |
| `Bonpos_USt.csv` | USt-Aufschlüsselung pro Artikelzeile |
| `TSE_Transaktionen.csv` | **Kritisch:** TSE-Transaktionsnummer (`TSE_TANR`), Signaturzähler (`TSE_TA_SIGZ`), Krypto-Signatur (`TSE_TA_SIG`) |

#### C. Kassenabschlussmodul (Z-Bon)

Die aggregierten Tages-/Schicht-Abrechnungen werden in `Z_GV_Typ.csv` und `Z_Zahlart.csv` des Stammdatenmoduls abgebildet.

### 8.4 Schlüsselfelder in Bonkopf.csv (Relationale Verknüpfung)

Fast jede CSV-Datei muss folgende Schlüssel in den ersten Spalten mitführen:

1. **`Z_KASSE_ID`** — Eindeutige Kassen-ID
2. **`Z_ERSTELLUNG`** — Zeitstempel des zugehörigen Kassenabschlusses
3. **`Z_NR`** — Fortlaufende Z-Bon-Nummer
4. **`BON_ID`** — Eindeutige Vorgangs-ID des Bons

### 8.5 ABRECHNUNGSKREIS — Tisch-Verknüpfung für Festzelt

Das Feld `ABRECHNUNGSKREIS` in `Bonkopf.csv` verknüpft mehrere Bons (Bestellungen + Zahlungen) zu einer logischen Einheit. Im Festzelt-Betrieb trägt jede Bestellung und jede Zahlung für denselben Tisch und dieselbe Gästegruppe denselben `ABRECHNUNGSKREIS`-Wert:

```
BON_ID | BON_TYP      | BON_START            | ABRECHNUNGSKREIS
-------|--------------|----------------------|--------------------
1001   | Bestellung   | 2026-05-01T18:01:00  | Tisch-42-20260501
1002   | Bestellung   | 2026-05-01T19:30:00  | Tisch-42-20260501
1003   | Beleg        | 2026-05-01T20:00:12  | Tisch-42-20260501
1004   | Beleg        | 2026-05-01T21:00:05  | Tisch-42-20260501
```

Die Prüfsoftware IDEA der Finanzämter kann so den vollständigen Tischverlauf rekonstruieren, auch wenn jede Transaktion für sich als sofort geschlossen registriert ist.

### 8.6 Storno-Handling in DSFinV-K

#### Positions-Storno (Storno vor Zahlung)

Ein falsch gebuchtes Produkt (z.B. 5 statt 4 Maß Bier) wird als neuer Bon mit **negativer Menge** gebucht:
- Neuer Bon mit `BON_TYP = "AVBelegstorno"` oder negativen Positionsmengen
- **`BON_STORNO = 1`** in `Bonkopf.csv`
- **`REF_BON_ID`** in `Bonkopf.csv` = `BON_ID` des fehlerhaften Ursprungsbons
- Eigene TSE-Signatur (`Kassenbeleg-V1` oder `Bestellung-V1` mit negativen Werten)
- `ABRECHNUNGSKREIS` identisch mit dem stornierten Vorgang

#### Bon-Storno (Storno nach Zahlung)

Eine bereits bezahlte Rechnung wird mit einem neuen Beleg mit negativen Beträgen storniert:
- Neuer Bon mit negativem Gesamtbetrag
- **`BON_STORNO = 1`** in `Bonkopf.csv`
- **`REF_BON_ID`** = `BON_ID` des Original-Zahlungsbelegs
- Eigene TSE-Transaktion (`Kassenbeleg-V1` mit negativem Betrag)

> **GoBD-Grundsatz:** Datensätze dürfen niemals per `UPDATE` oder `DELETE` aus der Datenbank entfernt werden. Stornierungen erzeugen immer neue Datensätze, die den Ursprungswert ausgleichen (Append-Only-Prinzip).

### 8.7 Architektonische Anforderungen an jotti

1. **CSV-Generator:** Komponente, die aus den Event-Store-Daten und Stammdaten die DSFinV-K-CSV-Struktur erzeugt (mit korrekten deutschen Dateinamen)
2. **`index.xml`-Generator:** Metadaten-Datei für die Prüfsoftware
3. **Z-Bon-Logik:** Kassenabschluss-Funktion, die Tagessummen aggregiert und einen Z-Bon erzeugt
4. **ABRECHNUNGSKREIS-Verwaltung:** Tisch-Session-ID persistieren und in allen zugehörigen Bons mitführen
5. **Admin-Endpunkt:** API-Endpunkt zum Auslösen des Exports (z.B. `POST /admin/dsfinvk-export`)
6. **ZIP-Generierung:** Alle CSVs + `index.xml` in ein ZIP-Archiv verpacken
7. **Steuersatz-Verwaltung:** USt-Sätze müssen als Stammdaten gepflegt werden (aktuell nicht vorhanden)

---

## 9. Anforderung 5: Elektronische Meldepflicht (ERiC / ELSTER)

### 9.1 Gesetzliche Grundlage und Fristen

Nach § 146a Abs. 4 AO müssen elektronische Aufzeichnungssysteme beim zuständigen Finanzamt gemeldet werden. Das Mitteilungsverfahren ist seit dem 1. Januar 2025 aktiv.

| Datum | Ereignis |
| --- | --- |
| 1. Januar 2025 | Meldeportal öffnet; Mitteilungsverfahren ist aktiv |
| 31. Juli 2025 | Abgabefrist für alle Systeme, die **vor dem 1. Juli 2025** angeschafft wurden |
| Innerhalb 1 Monat nach Anschaffung | Abgabefrist für Systeme, die **ab dem 1. Juli 2025** neu angeschafft werden |
| Innerhalb 1 Monat nach Außerbetriebnahme | Abgabefrist bei Stilllegung eines Systems |

Systeme, die bereits vor dem 1. Juli 2025 außer Betrieb genommen wurden, müssen nicht gemeldet werden.

*(Quelle: § 146a Abs. 4 AO — https://www.gesetze-im-internet.de/ao_1977/__146a.html)*
*(Quelle: BMF-Schreiben 28. Juni 2024 — Meldepflicht)*

### 9.2 Drei Übermittlungswege

1. **Direkteingabe im ELSTER-Web-Portal** (`www.elster.de`, „Mein ELSTER") — manuell, für Einzelfälle
2. **XML-Dateiupload** im ELSTER-Portal — semi-automatisch
3. **Programmatische Übermittlung über ERiC** (ELSTER Rich Client) — vollautomatisch aus der Kassensoftware heraus

ERiC ist die offizielle Softwarekomponente der Finanzverwaltung für die maschinelle Übermittlung. Sie validiert die Daten lokal, bevor sie an das Finanzamt übermittelt werden, und erzeugt bei Erfolg ein offizielles Bestätigungsprotokoll.

*(Quelle: ELSTER für Entwickler — https://www.elster.de/elsterweb/infoseite/entwickler)*

### 9.3 Submission-API (kommerzielle Alternative)

Alternativ bieten TSE-Anbieter wie fiskaly eine **Submission-API** als Abstraktionsschicht an:

- **Vorteil:** Keine direkte ERiC-Integration nötig; Kommunikation über Cloud-API
- **Nachteil:** Keine staatliche Vorab-Validierung; Verantwortung für Datenqualität liegt bei der Software
- **Abhängigkeit:** Vendor-Lock-in zum TSE-Anbieter

### 9.4 Meldepflichtige Daten (Payload)

Die Kassenmeldung umfasst je Kassensystem:

- Name und Steuernummer des Steuerpflichtigen
- Art des Kassensystems (Softwaretyp, Versionsnummer)
- Seriennummer des Kassensystems
- Zertifizierungs-ID der TSE (Format: `BSI-K-TR-nnnn-yyyy`)
- Seriennummer der TSE (64-stelliger Hexadezimalstring, ausschließlich 0–9 und A–F; **nicht** Base64 — Hinweis: in `Stamm_TSE.csv` für den DSFinV-K-Export wird zusätzlich der Public Key als Base64-Wert gespeichert, hierbei handelt es sich um ein anderes Feld)
- Anschaffungs- bzw. Inbetriebnahmedatum
- Betriebsstättenadresse

### 9.5 Architektonische Anforderungen an jotti

1. **Konfiguration:** Vereinsdaten, Betriebsstätte, Steuernummer als Stammdaten
2. **Entscheidung: ERiC vs. Submission-API:** Abwägung zwischen direkter Integration und Cloud-Abstraktion
3. **Admin-Funktion:** Kassenmeldung aus der Software heraus auslösen (`POST /admin/kassenmeldung`)
4. **Status-Tracking:** Meldestatus persistieren (gemeldet, ausstehend, Fehler)

---

## 10. Architektonische Lösungsansätze

### 10.1 TSE-Integration (empfohlener Ansatz)

```
┌─────────────────────────────────────────────┐
│                  jotti Backend               │
│                                              │
│  ┌──────────┐    ┌──────────────────────┐   │
│  │ Service-  │───▶│  TSE-Service          │   │
│  │ Handler   │    │  (Application Layer)  │   │
│  └──────────┘    └──────────┬───────────┘   │
│                              │               │
│                    ┌─────────▼─────────┐     │
│                    │  TSEClient        │     │
│                    │  (Interface)      │     │
│                    └─────────┬─────────┘     │
│                              │               │
└──────────────────────────────┼───────────────┘
                               │
                    ┌──────────▼──────────┐
                    │   Cloud-TSE API     │
                    │   (z.B. fiskaly)    │
                    └─────────────────────┘
```

**Interface-Design:**

```go
type TSEClient interface {
    StartTransaction(ctx context.Context, kassenID string, processType string, processData string) (StartResult, error)
    // UpdateTransaction nur für processType="Bestellung-V1" und "SonstigerVorgang-V1" zulässig.
    // Für "Kassenbeleg-V1" verboten (BMF-FAQ).
    UpdateTransaction(ctx context.Context, kassenID string, transactionNumber int, processData string) error
    FinishTransaction(ctx context.Context, kassenID string, transactionNumber int, processType string, processData string) (FinishResult, error)
}

type StartResult struct {
    TransactionNumber int
    LogTime           time.Time // TSE-interner Zeitstempel
    SerialNumberTSE   string
    SignatureCounter  int
}

type FinishResult struct {
    Signature        string
    LogTime          time.Time
    SignatureCounter int
}
```

### 10.2 Mapping: jotti-Vorgänge → TSE-Transaktionen (Atomares Modell)

Für das Festzelt-Muster (Abschnitt 5.6) gilt: Jeder Vorgang ist eine **eigenständige, sofort geschlossene** TSE-Transaktion.

| jotti-Vorgang | TSE-Operation | processType | Anmerkung |
| --- | --- | --- | --- |
| Bestellung aufnehmen | `Start` + sofort `Finish` | `Bestellung-V1` | Positionen in processData |
| Zahlung kassieren (Teilzahlung) | `Start` + sofort `Finish` | `Kassenbeleg-V1` | Betrag + Zahlungsart in processData; **kein** UpdateTransaction |
| Zahlung kassieren (Vollzahlung) | `Start` + sofort `Finish` | `Kassenbeleg-V1` | Wie oben |
| Positions-Storno | `Start` + sofort `Finish` | `Kassenbeleg-V1` | Negative Menge/Betrag; BON_STORNO=1 im DSFinV-K |
| Bon-Storno (nach Zahlung) | `Start` + sofort `Finish` | `Kassenbeleg-V1` | Negativer Gesamtbetrag; BON_STORNO=1, REF_BON_ID gesetzt |
| Tagesabschluss (Z-Bon) | `Start` + sofort `Finish` | `SonstigerVorgang-V1` | Tagesaggregat in processData |

**Alle Transaktionen eines Tisches** teilen denselben `ABRECHNUNGSKREIS`-Wert im DSFinV-K-Export.

### 10.3 Event-Store-Erweiterung

Die TSE-Rückgabewerte müssen in den bestehenden Event-Daten persistiert werden:

```go
// Erweiterung der Event-Data-Structs um TSE-Felder
type TSEData struct {
    TransactionNumber int    `json:"tseTransactionNumber"`
    LogTimeStart      string `json:"tseLogTimeStart"`  // logTime von StartTransaction
    LogTimeEnd        string `json:"tseLogTimeEnd"`    // logTime von FinishTransaction
    SignatureCounter  int    `json:"tseSignatureCounter"`
    Signature         string `json:"tseSignature"`
    SerialNumberTSE   string `json:"tseSerialNumber"`
    ProcessType       string `json:"tseProcessType"`
}

// Zusätzlich in TischSession-Daten:
type TischSession struct {
    AbrechnungskreisID    string    // z.B. "Tisch-42-20260501"
    ErsteBestellungLogTime time.Time // logTime der ersten Bestellung-V1 für den Bon-Aufdruck
}
```

### 10.4 DSFinV-K Export-Architektur

```
┌──────────────────────────────────────────────┐
│              DSFinV-K Exporter                │
│                                               │
│  ┌────────────┐  ┌────────────┐  ┌────────┐ │
│  │ Stammdaten- │  │ Einzelauf- │  │ Z-Bon- │ │
│  │ modul       │  │ zeichnungs-│  │ Modul  │ │
│  │             │  │ modul      │  │        │ │
│  └──────┬─────┘  └──────┬─────┘  └───┬────┘ │
│         │               │             │       │
│  ┌──────▼───────────────▼─────────────▼────┐ │
│  │   CSV-Generator (offizielle Dateinamen) │ │
│  │   Bonkopf.csv, Bonpos.csv, Stamm_*.csv  │ │
│  └──────────────────┬──────────────────────┘ │
│                     │                         │
│  ┌──────────────────▼──────────────────────┐ │
│  │     index.xml Generator + ZIP-Builder  │ │
│  └─────────────────────────────────────────┘ │
└──────────────────────────────────────────────┘
```

### 10.5 Cloud-TSE vs. ERiC — Entscheidungsmatrix für jotti

| Kriterium | ERiC (direkt, für Kassenmeldung) | Submission-API (fiskaly o.ä.) |
| --- | --- | --- |
| Vorab-Validierung | ✅ Ja (staatlich) | ❌ Nein |
| Implementierungsaufwand | Hoch (native C-Bibliothek) | Niedrig (REST-API) |
| Vendor-Lock-in | Keiner | Ja (TSE-Anbieter) |
| Kosten | Kostenlos | Kostenpflichtig |
| Offline-Fähigkeit | ✅ (lokale Lib) | ❌ (Cloud-abhängig) |
| Passt zu jotti-Philosophie | ✅ (Self-hosted, keine Cloud) | ⚠️ (externe Abhängigkeit) |

**Empfehlung:**
- Für die **TSE** (Transaktionssignierung): Cloud-TSE ist für BYOD-Festzelt-Szenarien pragmatisch, da Hardware-TSEs in Smartphones unpraktisch sind.
- Für die **Kassenmeldung** (§ 146a Abs. 4 AO): ERiC langfristig vorzuziehen — kostenlos, keine Cloud-Abhängigkeit, staatliche Validierung.

---

## 11. Handlungsempfehlungen und Priorisierung

### 11.1 Sofortmaßnahmen (Phase 0)

| # | Maßnahme | Aufwand | Beschreibung |
| --- | --- | --- | --- |
| 0.1 | Dokumentation korrigieren | Gering | Falsche Aussagen in README.md, docs/produktbeschreibung.md, docs/anforderungen.md, docs/handbuch.md und AGENTS.md korrigieren |
| 0.2 | Compliance-Roadmap veröffentlichen | Gering | Transparente Kommunikation über den aktuellen Compliance-Status und geplante Maßnahmen |

### 11.2 Kurzfristige Maßnahmen (Phase 1)

| # | Maßnahme | Aufwand | Beschreibung |
| --- | --- | --- | --- |
| 1.1 | TSE-Interface definieren | Mittel | `TSEClient`-Interface im Backend, Adapter-Pattern für verschiedene TSE-Anbieter |
| 1.2 | Cloud-TSE-Adapter implementieren | Hoch | Integration mit einem Cloud-TSE-Anbieter (z.B. fiskaly) |
| 1.3 | Event-Daten um TSE-Felder erweitern | Mittel | TSE-Rückgabewerte (`LogTimeStart`, `LogTimeEnd`, Signatur, Zähler) in Event-Data-Structs integrieren |
| 1.4 | ABRECHNUNGSKREIS in Tisch-Session | Mittel | Session-ID für DSFinV-K und erste-Bestellung-Zeitstempel persistieren |
| 1.5 | Belegausgabe implementieren | Mittel | Beleg-Generator mit allen Pflichtfeldern (inkl. TSE-Daten und erster Bestellzeitstempel) |

### 11.3 Mittelfristige Maßnahmen (Phase 2)

| # | Maßnahme | Aufwand | Beschreibung |
| --- | --- | --- | --- |
| 2.1 | DSFinV-K Export implementieren | Hoch | CSV-Generator für alle drei Module mit korrekten deutschen Dateinamen + `index.xml` + ZIP |
| 2.2 | Z-Bon-Logik (Kassenabschluss) | Mittel | Tagesabschluss mit aggregierten Summen |
| 2.3 | USt-Satz-Verwaltung | Mittel | Steuersätze als Stammdaten (7%, 19%, befreit, etc.) |
| 2.4 | QR-Code auf Beleg | Gering | TSE-Daten als QR-Code im DSFinV-K-Format |

### 11.4 Langfristige Maßnahmen (Phase 3)

| # | Maßnahme | Aufwand | Beschreibung |
| --- | --- | --- | --- |
| 3.1 | ERiC-Integration / Kassenmeldung | Hoch | Elektronische Kassenmeldung beim Finanzamt über ERiC-Schnittstelle |
| 3.2 | Archivierungsstrategie (10 Jahre) | Mittel | GoBD-konforme Langzeitarchivierung |
| 3.3 | Verfahrensdokumentation | Gering | Dokumentation des Kassensystems für Betriebsprüfung |

---

## 12. Quellenverzeichnis

| # | Quelle | URL |
| --- | --- | --- |
| 1 | § 146a AO — Ordnungsvorschrift für die Buchführung und für Aufzeichnungen mittels elektronischer Aufzeichnungssysteme | https://www.gesetze-im-internet.de/ao_1977/__146a.html |
| 2 | KassenSichV — Kassensicherungsverordnung | https://www.gesetze-im-internet.de/kassensichv/BJNR351500017.html |
| 3 | BSI TR-03153 — Technische Richtlinie für Technische Sicherheitseinrichtungen | https://www.bsi.bund.de/SharedDocs/Downloads/DE/BSI/Publikationen/TechnischeRichtlinien/TR03153/TR-03153.pdf?__blob=publicationFile |
| 4 | GoBD — BMF-Schreiben zur ordnungsmäßigen Führung elektronischer Bücher | https://www.bundesfinanzministerium.de/Content/DE/Downloads/BMF_Schreiben/Weitere_Steuerthemen/Abgabenordnung/2019-11-28-GoBD.html |
| 5 | DSFinV-K — Digitale Schnittstelle der Finanzverwaltung für Kassensysteme (BZSt) | https://www.bzst.de/DE/Unternehmen/Aussenpruefungen/DigitaleSchnittstelleFinV/digitaleschnittstellefinv_node.html |
| 6 | ELSTER für Entwickler — Offizielle Entwickler-Dokumentation | https://www.elster.de/elsterweb/infoseite/entwickler |
| 7 | § 14 AO — Wirtschaftlicher Geschäftsbetrieb | https://www.gesetze-im-internet.de/ao_1977/__14.html |
| 8 | § 64 AO — Steuerpflicht wirtschaftlicher Geschäftsbetriebe | https://www.gesetze-im-internet.de/ao_1977/__64.html |
| 9 | § 67a AO — Sportliche Veranstaltungen (Zweckbetrieb) | https://www.gesetze-im-internet.de/ao_1977/__67a.html |
| 10 | § 19 UStG — Kleinunternehmerregelung | https://www.gesetze-im-internet.de/ustg_1980/__19.html |
| 11 | BMF-FAQ zu § 146a AO (Stand Januar 2026) — Meldepflicht und processType-Erläuterungen | https://www.bundesfinanzministerium.de/ |
| 12 | BMF-Schreiben 28. Juni 2024 — Elektronische Kassenmeldepflicht nach § 146a Abs. 4 AO | https://www.bundesfinanzministerium.de/ |
| 13 | DSFinV-K Nr. 2.7 und Anhang H — Vereinfachungen für langanhaltende Bestellvorgänge (Festzelt-/Durchbedienen-Muster) | https://www.bzst.de/DE/Unternehmen/Aussenpruefungen/DigitaleSchnittstelleFinV/digitaleschnittstellefinv_node.html |
| 14 | § 379 AO — Steuergefährdung (bis 25.000 € Bußgeld) | https://www.gesetze-im-internet.de/ao_1977/__379.html |
