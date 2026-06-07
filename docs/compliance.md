# Compliance-Anforderungen: Fiskalische Grundlagen für jotti

> **Status:** Verbindliche Anforderungsdefinition — Grundsatzentscheidungen getroffen 2026-03-19
> **Betrifft:** KassenSichV, TSE, GoBD, Belegausgabepflicht, DSFinV-K, ERiC/ELSTER

---

## Inhaltsverzeichnis

1. [Einleitung](#1-einleitung)
2. [Rechtliche Grundlagen](#2-rechtliche-grundlagen)
3. [TSE-Integration (Technische Sicherheitseinrichtung)](#3-tse-integration-technische-sicherheitseinrichtung)
4. [GoBD-Konformität](#4-gobd-konformität)
5. [Belegausgabepflicht](#5-belegausgabepflicht)
6. [DSFinV-K Export-Schnittstelle](#6-dsfinv-k-export-schnittstelle)
7. [Elektronische Meldepflicht (ERiC / ELSTER)](#7-elektronische-meldepflicht-eric--elster)
8. [Betreiberpflichten](#8-betreiberpflichten)
9. [Architekturprinzipien → handbuch.md §3.13](#9-architekturprinzipien)
10. [Quellenverzeichnis](#10-quellenverzeichnis)

---

## 1. Einleitung

jotti ist ein **elektronisches Aufzeichnungssystem** (§ 1 KassenSichV) und unterliegt nach § 146a AO der TSE-Pflicht — unabhängig von Rechtsform, Gemeinnützigkeit oder Veranstaltungsdauer. Dieses Dokument beschreibt die Rechtsnormen, Entwickler- und Betreiberpflichten sowie die Compliance-Architektur. Technische Umsetzung phasenweise — siehe [anforderungen.md](anforderungen.md).

### Grundsatzentscheidungen (2026-03-19)

| Thema                                | Entscheidung                                                                                                                                                                                                                                                                               |
| ------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **Strategische Richtung**            | KassenSichV/TSE schrittweise implementieren (siehe [anforderungen.md](anforderungen.md))                                                                                                                                                                                                   |
| **TSE-Anbieter**                     | fiskaly (Cloud-TSE, API-first) als erster Zielanbieter; Adapter-Pattern für spätere Anbieter-Flexibilität                                                                                                                                                                                  |
| **Abrechnungskreis-Session**         | Pro Tisch pro Kassensitzung (= Tisch-Session). Im DSFinV-K-Export als `ABRECHNUNGSKREIS` mit dem Tischnamen abgebildet (z. B. `Tisch 42`). Intern: Subject `kassensitzung-{nr}/tisch-{id}`. Phase 2: manuelle Tischfreigabe durch Servicekraft bei Gästewechsel (neues Subject mit Suffix) |
| **Steuersätze**                      | 19 % (Standardsatz, z.B. Getränke), 7 % (ermäßigt, z.B. Speisen), 0 % / steuerbefreit (Zweckbetrieb)                                                                                                                                                                                       |
| **Kassenmeldung (§ 146a Abs. 4 AO)** | Phase 1: manuell über ELSTER-Webportal; Phase 2: ERiC oder fiskaly-Submission-API                                                                                                                                                                                                          |
| **Seriennummer**                     | UUID beim ersten Containerstart generieren, dauerhaft in DB speichern, im Admin-Dashboard anzeigen                                                                                                                                                                                         |
| **Belegausgabe BYOD**                | Phase 1: zentraler Bondrucker an der Theke (Backend steuert Drucker nach TSE-Abschluss); Phase 2 (optional): digitaler eBeleg via QR-Code als Download-Link                                                                                                                                |
| **Rechtliche Rollenverteilung**      | jotti ist Source-Available-Software (kein SaaS); Entwickler implementiert TSE-Schnittstellen; Betreiber (Verein) trägt Betriebspflichten und ELSTER-Meldepflicht                                                                                                                           |

---

## 2. Rechtliche Grundlagen

### 2.1 Abgabenordnung (AO) — § 146a

§ 146a Abs. 1 AO verpflichtet jeden Betreiber eines elektronischen Aufzeichnungssystems, jeden Geschäftsvorfall einzeln, vollständig, richtig, zeitgerecht und geordnet zu erfassen und durch eine **zertifizierte TSE** zu schützen — unabhängig von Rechtsform, Gemeinnützigkeit, Betriebsdauer oder Gewinnerzielungsabsicht.

_(Quelle: § 146a AO — https://www.gesetze-im-internet.de/ao_1977/__146a.html)_

### 2.2 Kassensicherungsverordnung (KassenSichV)

Die KassenSichV konkretisiert § 146a AO; § 1 KassenSichV definiert als Aufzeichnungssysteme „elektronische oder computergestützte Kassensysteme oder Registrierkassen" — jotti fällt als browserbasiertes Kassensystem eindeutig darunter.

Für **mobile Geräte** (Smartphones als Browser-Clients in jottis BYOD-Modell) gilt: Wenn ein Gerät technisch in der Lage ist, Zahlungsvorgänge eigenständig zu erfassen und offline zu betreiben, muss es selbst an eine TSE angebunden sein. Fungiert es ausschließlich als Eingabeterminal, das sofort an ein TSE-gesichertes Backend weiterleitet, genügt die Backend-seitige TSE-Anbindung. Entscheidend ist die technische Fähigkeit zur selbständigen Offline-Erfassung, nicht die tatsächliche Nutzung.

**Einordnung für jotti:** In jottis Architektur fungieren die Smartphones der Servicekräfte als **reine „Eingabegeräte"** — ihre Funktion geht rechtlich nicht über die einer einfachen Tastatur hinaus. Dies hat zwei unmittelbare Konsequenzen:

1. Die Smartphones benötigen **keine eigene TSE**. Die gesamte TSE-Absicherung, Protokollierung und Datenspeicherung (DSFinV-K) erfolgt zentral im Backend.
2. **Architektonische Pflicht:** Die Webapp muss bei einem Internetausfall **sofort blockieren** und jede Offline-Erfassung von Barzahlungen technisch verhindern. Sobald die Servicekraft keine direkte Verbindung zum Backend herstellen kann, darf keine Zahlung erfasst werden. Nur so ist die Einordnung als reines Eingabegerät rechtlich haltbar.

_(Quelle: KassenSichV — https://www.gesetze-im-internet.de/kassensichv/BJNR351500017.html)_
_(Quelle: BMF-FAQ zu § 146a AO, Frage zur Abgrenzung von Eingabegeräten und eigenständigen Aufzeichnungssystemen — https://www.bundesfinanzministerium.de/)_
_(Quelle: AEAO zu § 146a AO — Klarstellung zur Mitteilungspflicht für verbundene Eingabegeräte — https://www.bundesfinanzministerium.de/)_

### 2.3 GoBD

Die GoBD (BMF-Schreiben 28.11.2019) fordern Nachvollziehbarkeit, Vollständigkeit, Richtigkeit, Zeitgerechtigkeit, Ordnung und Unveränderbarkeit. Sie gelten für alle Steuerpflichtigen — einschließlich gemeinnütziger Vereine im wirtschaftlichen Geschäftsbetrieb (z.B. Vereinsfeste außerhalb § 67a AO).

_(Quelle: GoBD — https://www.bundesfinanzministerium.de/Content/DE/Downloads/BMF_Schreiben/Weitere_Steuerthemen/Abgabenordnung/2019-11-28-GoBD.html)_

### 2.4 Belegausgabepflicht (§ 146a Abs. 2 AO, § 6 KassenSichV)

§ 146a Abs. 2 AO und § 6 KassenSichV verpflichten zur Belegausgabe in **unmittelbarem zeitlichen Zusammenhang** nach jedem Kassiervorgang — in Papierform oder (mit Zustimmung des Kunden) elektronisch.

_(Quelle: § 146a Abs. 2 AO — https://www.gesetze-im-internet.de/ao_1977/__146a.html)_

### 2.5 DSFinV-K (Digitale Schnittstelle der Finanzverwaltung für Kassensysteme)

§ 4 KassenSichV verlangt eine **einheitliche digitale Schnittstelle**, über die die gespeicherten Daten für die Finanzverwaltung exportiert werden können. Die DSFinV-K (aktuell Version 2.4, Stand Januar 2024) definiert das genaue Format dieses Exports: eine Sammlung von CSV-Dateien mit fest vorgeschriebenen (englischen, kleingeschriebenen) Dateinamen, Spaltenreihenfolge und Semikolon-Trennung, verpackt in einem ZIP-Archiv mit `index.xml`.

_(Quelle: BZSt — DSFinV-K — https://www.bzst.de/DE/Unternehmen/Aussenpruefungen/DigitaleSchnittstelleFinV/digitaleschnittstellefinv_node.html)_

### 2.6 Elektronische Kassenmeldepflicht (§ 146a Abs. 4 AO)

Ab dem 1. Januar 2025 müssen elektronische Aufzeichnungssysteme dem zuständigen Finanzamt **elektronisch** gemeldet werden. Die Meldung erfolgt über das ELSTER-System, wahlweise direkt im Portal oder programmatisch über die ERiC-Schnittstelle (ELSTER Rich Client). Genauere Fristen: siehe Abschnitt 7.

_(Quelle: § 146a Abs. 4 AO — https://www.gesetze-im-internet.de/ao_1977/__146a.html)_
_(Quelle: BMF-FAQ — https://www.bundesfinanzministerium.de/)_

### 2.7 Sonderfall: Source-Available Self-hosted System — Pflichten des Entwicklers und der Betreiber

jotti unterscheidet sich von kommerziellen SaaS-Kassensystemen fundamental: Es wird als quelloffene Software auf GitHub bereitgestellt, und die Vereine betreiben das System eigenverantwortlich auf ihrem eigenen Server (VPS via Docker). Diese Konstellation schafft eine klare Rollentrennung zwischen dem **Entwickler (Hersteller)** und den **Betreibern (Vereine)** — mit jeweils unterschiedlichen Rechtspflichten.

#### a) Pflichten des Entwicklers (Hersteller)

Nach **§ 146a Abs. 1 Satz 5 AO** i.V.m. **§ 379 AO** ist es verboten, Kassensoftware in Verkehr zu bringen oder zu bewerben, die nicht über die **Möglichkeit** verfügt, eine zertifizierte TSE anzubinden. Das kostenlose Bereitstellen von Code auf GitHub gilt als „In-Verkehr-Bringen“ — der Source-Available-Charakter des Projekts ändert daran nichts.

**Kernpflichten des Entwicklers:**

- Die **TSE-Schnittstelle** (z.B. `TSEClient`-Interface) und der **DSFinV-K-Export** müssen im Code vorhanden und nutzbar sein.
- Wenn der Code diese Schnittstelle enthält, der Verein bei der Docker-Installation aber entscheidet, keinen TSE-API-Key einzutragen und die Kasse ohne TSE zu nutzen, liegt das rechtliche Risiko **ausschließlich beim Verein**.
- Der Entwickler hat **keine Meldepflicht** gegenüber dem Finanzamt für die durch Dritte betriebenen Installationen. Mitteilungspflichtig ist immer nur die juristische Person, die das System tatsächlich verwendet.
- Eine **Muster-Verfahrensdokumentation** im GitHub-Repository ist empfehlenswert: Sie beschreibt für Betriebsprüfer, wie die Architektur funktioniert, wie die Datenbank geschützt ist und wie die TSE-Anbindung technisch erfolgt.

#### b) Pflichten der Betreiber (Vereine)

Die Vereine tragen als Betreiber die volle operative und rechtliche Verantwortung:

- **TSE-Beschaffung (Bring Your Own TSE):** Da auf einem Cloud-VPS keine Hardware-TSE eingesteckt werden kann, müssen die Vereine selbst einen Vertrag mit einem Cloud-TSE-Anbieter (wie fiskaly oder D-Trust) abschließen. Sie erhalten API-Keys, die über Umgebungsvariablen (`.env`-Datei) in den Docker-Container injiziert werden.
- **ELSTER-Meldung:** Jede Kassen-Instanz, die auf dem VPS des Vereins läuft, muss innerhalb eines Monats nach Inbetriebnahme über das eigene ELSTER-Portal beim Finanzamt angemeldet werden (§ 146a Abs. 4 AO). Die hierfür benötigte Seriennummer wird von jotti bei der ersten Inbetriebnahme automatisch generiert und im Admin-Bereich angezeigt (siehe Abschnitt 3.7).
- **Server-Betrieb:** Der Verein ist für Verfügbarkeit, Datensicherung, Zugriffskontrolle und die 10-jährige GoBD-konforme Aufbewahrung der Daten verantwortlich.

_(Quelle: § 146a Abs. 1 Satz 5 AO — https://www.gesetze-im-internet.de/ao_1977/__146a.html)_
_(Quelle: § 379 AO — Steuergefährdung — https://www.gesetze-im-internet.de/ao_1977/__379.html)_

---

## 3. TSE-Integration (Technische Sicherheitseinrichtung)

### 3.1 Hintergrund

Die TSE ist das kryptografische Herzstück eines konformen Kassensystems. Sie besteht aus drei Modulen:

1. **Sicherheitsmodul:** Aufgeteilt in SMAERS (Datenaufbereitung) und CSP (kryptografische Signatur)
2. **Speichermedium:** Lokale Sicherung der signierten Transaktionsdaten
3. **Einheitliche Digitale Schnittstelle (EDS):** Standardisierte Exportschnittstelle

_(Quelle: BSI TR-03153 — https://www.bsi.bund.de/SharedDocs/Downloads/DE/BSI/Publikationen/TechnischeRichtlinien/TR03153/TR-03153.pdf?__blob=publicationFile)_

### 3.2 Protokollierungs-Ablauf (Transaktions-Lebenszyklus)

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

### 3.3 Offizielle processType-Werte

Die `processType`-Werte sind im AEAO zu § 146a AO, Anhang I, festgelegt und in DSFinV-K Anhang I referenziert. Die **-V1-Endung ist Bestandteil des offiziellen Strings** und muss exakt so an die TSE übergeben werden.

| processType           | Verwendung                                                                 |
| --------------------- | -------------------------------------------------------------------------- |
| `Kassenbeleg-V1`      | Zahlungsbeleg (Rechnung), der dem Kunden ausgehändigt wird                 |
| `Bestellung-V1`       | Zwischenabsicherung einer Bestellung ohne sofortige Zahlung (Gastronomie)  |
| `SonstigerVorgang-V1` | Alle anderen abzusichernden Vorgänge (Tagesabschluss, TSE-Selbsttest, ...) |

### 3.4 Datenformat-Vorgaben (`processData`)

Die Formatierung der `processData` ist streng reguliert:

- **Encoding:** UTF-8 oder ASCII, kein BOM (Byte-Order-Mark)
- **Dezimaltrennzeichen:** Ausschließlich Punkt (`.`)
- **Verboten:** Tausendertrennzeichen, Exponentialschreibweise, `+` vor positiven Werten
- **Mindestens eine Stelle vor dem Dezimaltrennzeichen:** `0.5` statt `.5`
- **Format-String für `Kassenbeleg-V1`:** `Beleg^<Betrag_Normal>_<Betrag_Ermaessigt>_<Betrag_Null>_<Betrag_Besonderer_Satz>_<Betrag_Befreit>^<Zahlbetrag>:<Zahlungsart>`
- **Format-String für `Bestellung-V1`:** Positionen als strukturierter Text (z.B. `4x Maß Bier_2x Weißwurst`) — genaues Format gemäß AEAO § 146a Anhang I

### 3.5 TSE-Varianten

Es gibt zwei grundlegende Integrationsansätze:

| Variante         | Beschreibung                                       | Beispiel-Anbieter                |
| ---------------- | -------------------------------------------------- | -------------------------------- |
| **Hardware-TSE** | Physisches Gerät (USB-Stick, SD-Karte, Smartcard)  | Swissbit, Epson, Diebold Nixdorf |
| **Cloud-TSE**    | TSE als Cloud-Service, Kommunikation via HTTPS-API | fiskaly, Deutsche Fiskal         |

Für jotti als Self-hosted-System ist eine **Cloud-TSE** der gewählte Ansatz, da sie keine zusätzliche Hardware erfordert und über HTTP-POST-Requests angesprochen wird. Eine Hardware-TSE scheidet für BYOD-Smartphone-Setups im Festzelt praktisch aus.

**Gewählter Anbieter: fiskaly**

fiskaly ist der initiale Zielanbieter für die TSE-Integration. Gründe:

- API-first, gut dokumentierte REST-Schnittstelle
- BSI-zertifizierte Cloud-TSE (nach BSI TR-03153)
- Unterstützt alle drei `processType`-Werte (`Kassenbeleg-V1`, `Bestellung-V1`, `SonstigerVorgang-V1`)
- Bietet optional eine Submission-API für die Kassenmeldung (§ 146a Abs. 4 AO)

Das Backend-Interface `TSEClient` ist **anbieter-agnostisch** (Adapter-Pattern), so dass ein späterer Wechsel zu einem anderen Cloud-TSE-Anbieter ohne Änderungen am Domain-Code möglich ist.

### 3.6 Das Festzelt-Muster: Atomare TSE-Transaktionen (Empfehlung)

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

**Verknüpfung**: Alle Bestellungen und Zahlungen eines Tisches werden im DSFinV-K-Export über das Feld `ABRECHNUNGSKREIS` (z.B. `"Tisch 42"`) zusammengeführt (siehe Abschnitt 6).

#### Konkretes Szenario: Maihock, Tisch 42

```
18:00 — Gruppe A setzt sich (4 Personen)
18:01 — Bestellung: 4x Maß Bier, 2x Weißwurst
        → StartTransaction(Bestellung-V1) + FinishTransaction
          → TSE-Signatur S1, transactionNumber=1001
        → ABRECHNUNGSKREIS = "Tisch 42"

19:30 — Bestellung: 4x Maß Bier (Nachbestellung)
        → StartTransaction(Bestellung-V1) + FinishTransaction
          → TSE-Signatur S2, transactionNumber=1002
        → ABRECHNUNGSKREIS = "Tisch 42"

20:00 — 2 Gäste zahlen: 2x Bier = 14,00 €, bar
        → StartTransaction(Kassenbeleg-V1) + FinishTransaction
          → TSE-Signatur S3, transactionNumber=1003
          → Bon: enthält zusätzlich "Erste Bestellung: 18:01 Uhr" (s. §7)
        → ABRECHNUNGSKREIS = "Tisch 42"

21:00 — Restliche 2 Gäste zahlen: 2x Bier + 2x Weißwurst = 18,00 €, bar
        → StartTransaction(Kassenbeleg-V1) + FinishTransaction
          → TSE-Signatur S4, transactionNumber=1004
          → Bon: enthält zusätzlich "Erste Bestellung: 18:01 Uhr"
        → ABRECHNUNGSKREIS = "Tisch 42"

22:00 — Gruppe B setzt sich (neue Gäste, neuer ABRECHNUNGSKREIS)
        → ABRECHNUNGSKREIS = "Tisch 42-B"
```

Der Betriebsprüfer sieht im DSFinV-K-Export alle vier Transaktionen mit demselben `ABRECHNUNGSKREIS` und kann den vollständigen Tischverlauf nachvollziehen, obwohl jede Transaktion sofort geschlossen wurde.

### 3.7 Seriennummer-Generierung bei Self-hosted Docker-Instanzen

Da jotti auf einem VPS ohne physische Kassenhardware betrieben wird, gibt es kein Typenschild mit aufgedruckter Seriennummer. Die gesetzliche Anforderung (§ 146a AO, DSFinV-K, § 6 KassenSichV) nach einer eindeutigen Seriennummer des elektronischen Aufzeichnungssystems muss softwareseitig erfüllt werden.

#### Anforderung

- **Beim allerersten Start** des Docker-Containers (Datenbank-Initialisierung) generiert jotti automatisch eine eindeutige UUID als Kassen-Seriennummer.
- Die UUID wird dauerhaft in der Datenbank gespeichert und **nie überschrieben** (auch nicht bei Updates oder Neustart des Containers).
- Die Seriennummer wird im **Admin-Dashboard** prominent angezeigt, damit der Betreiber (Verein) sie bei der ELSTER-Meldung eintragen und für den DSFinV-K-Export verwenden kann.
- **Disaster Recovery:** Bei einem Verlust der Datenbank (z.B. Server-Ausfall ohne Backup) muss die alte Seriennummer beim Finanzamt abgemeldet und eine neue Instanz mit neuer Seriennummer angemeldet werden. Das Backup der Datenbank ist daher zwingend notwendig, um die Seriennummerenkontinuität zu wahren.

#### Verwendung der Seriennummer

| Verwendungsort                               | Feld / Kontext                                         |
| -------------------------------------------- | ------------------------------------------------------ |
| ELSTER-Meldung                               | „Seriennummer des elektronischen Aufzeichnungssystems" |
| DSFinV-K Export (`cashregister.csv`)         | Feld `KASSE_SERIENNR`                                  |
| Kassenbon (Pflichtfeld nach § 6 KassenSichV) | Angedruckter String, z.B. `Kassen-ID: 7f3a9d12-...`    |
| TSE-Kommunikation (`StartTransaction`)       | Parameter `kassenID`                                   |

#### Beispiel-Format

```
Kassen-ID: 7f3a9d12-84e1-4b2c-9f6a-1234567890ab
```

Die UUID ist herstellerunabhängig, weltweit eindeutig und erfordert keine zentrale Registrierung — ideal für ein Self-hosted Source-Available-System, bei dem der Entwickler keine Kontrolle über die Anzahl der laufenden Instanzen hat.

---

## 4. GoBD-Konformität

### 4.1 Aktueller Stand

jotti erfüllt durch die Event-Sourcing-Architektur bereits mehrere GoBD-Grundsätze:

| GoBD-Grundsatz             | Aktueller Status    | Anmerkung                                             |
| -------------------------- | ------------------- | ----------------------------------------------------- |
| Unveränderbarkeit          | ✅ Erfüllt          | Kassenjournal ist append-only, kein UPDATE/DELETE     |
| Nachvollziehbarkeit        | ✅ Erfüllt          | Lückenloses Kassenjournal pro Tisch-Session           |
| Vollständigkeit            | ✅ Erfüllt          | Jeder Geschäftsvorfall wird als Event erfasst         |
| Zeitgerechte Buchung       | ✅ Erfüllt          | Events mit Echtzeit-Zeitstempel                       |
| Ordnungsmäßigkeit          | ✅ Erfüllt          | Strukturiertes Datenmodell, typisierte Events         |
| Kryptografische Verkettung | ❌ Fehlt            | Keine TSE-Signatur, keine kryptografische Absicherung |
| 10-Jahres-Aufbewahrung     | ⚠️ Nicht adressiert | Keine Archivierungsstrategie implementiert            |

### 4.2 Anforderungen gemäß §§ 146, 147 AO und GoBD

- **Aufbewahrungspflicht:** Alle steuerlich relevanten Daten müssen **10 Jahre** aufbewahrt werden, jederzeit verfügbar, unverzüglich lesbar, vollständig und absolut unveränderbar.
- **Elektronisches Radierverbot:** Datensätze dürfen nach der Erfassung nicht per `UPDATE` oder `DELETE` überschrieben oder gelöscht werden.
- **Stornierungen:** Müssen als neue Buchungssätze (mit neuem Zeitstempel und neuer TSE-Signatur) erzeugt werden, die den alten Wert ausgleichen — niemals als nachträgliche Änderung.
- **Verfahrensdokumentation:** Es muss dokumentiert sein, wie das System Daten erzeugt, verarbeitet und archiviert.

_(Quelle: GoBD — BMF-Schreiben vom 28.11.2019 — https://www.bundesfinanzministerium.de/Content/DE/Downloads/BMF_Schreiben/Weitere_Steuerthemen/Abgabenordnung/2019-11-28-GoBD.html)_

### 4.3 Handlungsbedarf

| Maßnahme                                     | Priorität | Aufwand |
| -------------------------------------------- | --------- | ------- |
| TSE-Signatur in Event-Daten integrieren      | Hoch      | Mittel  |
| Archivierungsstrategie definieren (10 Jahre) | Mittel    | Gering  |
| Verfahrensdokumentation erstellen            | Mittel    | Gering  |
| Soft-Delete-Praxis bei Stammdaten prüfen     | Niedrig   | Gering  |

---

## 5. Belegausgabepflicht

### 5.1 Gesetzliche Grundlage

Gemäß § 146a Abs. 2 AO und § 6 KassenSichV muss für **jeden Kassiervorgang** ein Beleg erzeugt und dem Kunden zur Verfügung gestellt werden. Der Beleg kann gedruckt oder — mit Zustimmung des Kunden — elektronisch (z.B. per QR-Code oder PDF) bereitgestellt werden.

> **Wichtige Befreiung für Vereinsfeste (§ 146a Abs. 2 Satz 2 AO):** Bei „Verkauf von Waren an eine **Vielzahl nicht bekannter Personen**" kann das Finanzamt aus Zumutbarkeitsgründen von der Pflicht zur **Aushändigung** des Belegs befreien. Genau dieser Tatbestand greift im typischen Festzelt-/Vereinsfest-Betrieb. Zu beachten:
>
> - Die Befreiung betrifft **nur die Aushändigung**, **nicht** die TSE-Absicherung, die Erfassung und die Belegerzeugung — diese bleiben vollumfänglich Pflicht. Der Beleg muss also jederzeit **erstellbar** sein, auch wenn er nicht aktiv ausgehändigt wird.
> - Die Befreiung wird **nicht automatisch** gewährt: Der Betreiber muss sie **schriftlich beim zuständigen Finanzamt beantragen**; sie kann widerrufen werden.
> - Verlangt ein Gast ausdrücklich einen Beleg, ist dieser auszuhändigen.
>
> jotti setzt darauf auf, indem es den Kassenbeleg **auf Anforderung** erstellt (→ [anforderungen.md F-03](anforderungen.md)), statt nach jeder Zahlung automatisch zu drucken.

_(Quelle: § 146a Abs. 2 Satz 2 AO — https://www.gesetze-im-internet.de/ao_1977/__146a.html)_

_(Quelle: KassenSichV § 6 — https://www.gesetze-im-internet.de/kassensichv/BJNR351500017.html)_

> **Abgrenzung in jotti — Arbeitsbon ≠ Kassenbeleg:** jotti kennt zwei Bon-Familien (→ [handbuch.md §4.6](handbuch.md)). Der **Arbeitsbon** (automatisch bei Bestellaufnahme, an Küche/Theke) ist **rein operativ und nicht-fiskalisch**: Er trägt keine Preise, löst keine TSE-Transaktion aus und ist **kein** Beleg i. S. v. § 146a AO. Der **Kassenbeleg** (auf Anforderung pro Kassiervorgang) ist der **fiskalische § 146a-Beleg** mit allen Pflichtangaben (siehe Abschnitt 5.2). Nur der Kassenbeleg unterliegt der Belegausgabepflicht und — nach Umsetzung der TSE-Integration (→ [Abschnitt 3](#3-tse-integration-technische-sicherheitseinrichtung)) — der TSE-Absicherung.

### 5.2 Pflichtangaben auf dem Beleg

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

### 5.3 Besondere Anforderung beim Festzelt-Muster (Durchbedienen)

Wenn die Tisch-Bestellungen mit `Bestellung-V1`-Transaktionen abgesichert wurden und erst später bezahlt wird (atomares Transaktionsmodell gemäß Abschnitt 3.6), gilt laut BMF-FAQ:

> „Zusätzlich ist auf den Bon der **Startzeitpunkt der ersten Bestellung in Klarschrift aufzudrucken**."
> _(Quelle: BMF-FAQ zu § 146a AO; DSFinV-K Nr. 2.7 sowie Anhang H)_

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

### 5.4 QR-Code-Format

Um Platz auf dem Beleg zu sparen, können die TSE-Daten in einen standardisierten **QR-Code** verpackt werden. Das Format muss den Vorgaben der DSFinV-K (Anhang I) entsprechen.

### 5.5 Architektonische Anforderungen an jotti

1. **Beleg-Generator:** Komponente, die aus einem abgeschlossenen Kassiervorgang einen konformen Beleg (PDF oder Druckformat) erzeugt
2. **TSE-Daten auf dem Beleg:** Alle TSE-Rückgabewerte müssen auf dem Beleg erscheinen
3. **Erste-Bestellung-Zeitstempel:** Das Backend muss den `logTime` der ersten `Bestellung-V1`-Transaktion einer Tisch-Session persistieren und beim Beleg-Druck abrufen
4. **QR-Code-Generierung:** TSE-Daten als QR-Code im DSFinV-K-Format
5. **Beleg-Ausgabekanal:** Primär über stationären Bondrucker an der Theke (Backend steuert Drucker nach TSE-Abschluss); optional als digitaler eBeleg via QR-Code (Download-Link) — Details siehe Abschnitt 5.6
6. **Beleg-Archivierung:** Belegdaten müssen für den DSFinV-K-Export persistiert werden

### 5.6 Umsetzung der Belegausgabepflicht im BYOD-Setup

Da die Servicekräfte private Smartphones ohne mobile Bondrucker verwenden, erfordert die Belegausgabe nach § 146a Abs. 2 AO in jottis BYOD-Setup besondere architektonische Überlegungen.

#### Variante A: Zentraler Bondrucker an der Theke (Primärlösung)

**Ablauf:**

1. Die Servicekraft kassiert auf ihrem Smartphone (Webapp).
2. Die `Kassenbeleg-V1`-Transaktion wird **vollständig abgeschlossen** (TSE `FinishTransaction` aufgerufen). Erst jetzt stehen alle kryptografischen Prüfwerte für den Beleg fest.
3. Erst nach erfolgreicher TSE-Signatur sendet die Webapp über das Backend einen Druckbefehl an den stationären Bondrucker an der Theke.
4. Die Servicekraft nimmt den gedruckten Bon von der Theke und bietet ihn dem Gast an.
5. **Der Gast ist gesetzlich nicht verpflichtet, den Bon anzunehmen.** Lehnt der Gast ab, ist die Belegausgabepflicht dennoch erfüllt — § 146a Abs. 2 AO verlangt das „Ausstellen und Zur-Verfügung-Stellen", nicht die Annahme; der Papierbon kann vernichtet werden.

> **Hinweis:** Die Reihenfolge TSE-Abschluss **vor** Drucken ist rechtlich zwingend. Der Bon darf erst gedruckt werden, wenn `FinishTransaction` erfolgreich zurückgekehrt ist.

#### Variante B: Digitaler eBeleg via QR-Code (Optionale Erweiterung)

Als ergänzende oder alternative Lösung (insbesondere wenn kein Bondrucker an der Theke vorhanden ist) können digitale eBelege ausgegeben werden. Dabei gelten folgende gesetzliche Anforderungen:

- **Bildschirmanzeige allein ist unzulässig:** Es reicht nicht aus, den Kassenbon lediglich als Bild auf dem Kellner-Smartphone anzuzeigen. Der Gast muss die Möglichkeit haben, den Beleg **elektronisch entgegenzunehmen**.
- **QR-Code als Download-Link:** jotti zeigt nach dem Kassiervorgang auf dem Kellner-Smartphone einen QR-Code an. Dieser Code ist ein Download-Link zu einem Beleg im standardisierten Format (PDF, JPG oder PNG), der auf dem Backend-Server abrufbar ist. Der Gast scannt diesen Code mit seinem eigenen Smartphone.
- **Zustimmung des Gastes:** Ein elektronischer Beleg darf nur mit Zustimmung des Empfängers ausgegeben werden. Diese gilt als konkludent (stillschweigend) erteilt, sobald der Gast den QR-Code scannt (§ 146a Abs. 2 AO i.V.m. dem Grundsatz der konkludenten Einwilligung; vgl. auch BMF-FAQ zu § 146a AO zum elektronischen Beleg).
- **Aufbewahrung:** Die generierten eBelege müssen für den DSFinV-K-Export und die GoBD-konforme Archivierung gespeichert werden.

**Umsetzungsreihenfolge:**

- **Phase 1:** Zentraler Bondrucker an der Theke (geringerer Entwicklungsaufwand, sofortige Rechtskonformität)
- **Phase 2 (optional):** Digitaler eBeleg via QR-Code als Ergänzung oder Alternative

_(Quelle: § 146a Abs. 2 AO — Belegausgabepflicht)_
_(Quelle: KassenSichV § 6 — Inhalt des Belegs)_

---

## 6. DSFinV-K Export-Schnittstelle

### 6.1 Übersicht

Die Finanzverwaltung verlangt bei einer Kassen-Nachschau oder Betriebsprüfung einen genormten, maschinenlesbaren Datenexport. Dieser Export folgt der **DSFinV-K-Spezifikation** (Version 2.4, Stand Januar 2024) und muss von der Prüfsoftware IDEA der Finanzämter gelesen werden können.

_(Quelle: BZSt — https://www.bzst.de/DE/Unternehmen/Aussenpruefungen/DigitaleSchnittstelleFinV/digitaleschnittstellefinv_node.html)_

### 6.2 Dateiformat und Grundregeln

- **Gesamtstruktur:** ZIP-Archiv mit CSV-Dateien und einer `index.xml`
- **`index.xml`:** Metadaten-Datei, die dem Prüftool mitteilt, wie die CSVs zu parsen sind
- **Kopfzeile:** Jede CSV-Datei beginnt zwingend mit einem Header-Datensatz
- **Trennzeichen:** Semikolon (`;`) als Feldtrennzeichen
- **Zeilenumbrüche:** CRLF (`\r\n`)
- **Zahlenformate:** Keine Tausendertrennzeichen, Punkt als Dezimaltrennzeichen, mindestens eine Stelle vor dem Dezimaltrennzeichen (`0.5`, nicht `.5`), keine führenden Nullen
- **Spaltenreihenfolge:** Exakt wie in der Spezifikation vorgegeben
- **Dateinamen:** Die physischen CSV-Dateinamen der DSFinV-K sind **englisch und kleingeschrieben** (z. B. `transactions.csv`, `lines.csv`, `cashregister.csv`, `tse.csv`, `businesscases.csv`) und dürfen nicht abgeändert werden. Die DSFinV-K-Spezifikation verwendet **zusätzlich** deutsche _logische_ Bezeichnungen für die Abschnittsüberschriften (z. B. „Bonkopf" für `transactions.csv`, „Bonpos" für `lines.csv`) — diese logischen Namen sind **nicht** die Dateinamen.
- **Custom-Felder:** Zusätzliche Spalten am Ende erlaubt, müssen aber in `index.xml` definiert werden

> **Achtung — verbreitetes Missverständnis:** Die offiziellen Dateinamen sind `transactions.csv`, `lines.csv`, `allocation_groups.csv`, `references.csv` usw. — **nicht** `Bonkopf.csv`/`Bonpos.csv`. Wer deutsche Dateinamen wie `Bonkopf.csv` exportiert, erzeugt eine **nicht** DSFinV-K-konforme Datei. Maßgeblich ist die offizielle DFKA-Taxonomie (DSFinV-K v2.4) bzw. das BMF-Datenschema.

### 6.3 Modul-Struktur und offizielle Dateinamen

Der Export gliedert sich in drei Module. Die Tabellen nennen jeweils den **offiziellen Dateinamen** (englisch) und in Klammern die **logische DSFinV-K-Bezeichnung** (deutsch).

#### A. Stammdatenmodul (Master Data)

| Dateiname (offiziell)  | Logische Bezeichnung | Inhalt                                                                                                                                                                   |
| ---------------------- | -------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `cashpointclosing.csv` | Stamm_Abschluss      | Metadaten zum Z-Bon (Kassenabschluss): Unternehmensname, Steuernummer, Start-/End-Zeitpunkt                                                                              |
| `location.csv`         | Stamm_Orte           | Standortdaten der Betriebsstätte                                                                                                                                         |
| `cashregister.csv`     | Stamm_Kassen         | Kassendaten: Hersteller, Seriennummer, Software-Typ und -Version                                                                                                         |
| `tse.csv`              | Stamm_TSE            | TSE-Daten: Zertifikats-ID, Signaturalgorithmus (z. B. `ecdsa-plain-SHA256`), TSE-Seriennummer (64-stelliger Hexadezimalstring, 0–9 und A–F), Public Key (Base64-kodiert) |
| `vat.csv`              | Stamm_USt            | Stammdaten der verwendeten Steuersätze                                                                                                                                   |

#### B. Einzelaufzeichnungsmodul (Bonmodul / Transactions)

| Dateiname (offiziell)   | Logische Bezeichnung | Inhalt                                                                                                                           |
| ----------------------- | -------------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| `transactions.csv`      | Bonkopf              | Ein Datensatz pro Kassenbon: `BON_ID`, `BON_NR`, `BON_TYP`, `BON_START`, `BON_ENDE` (ISO 8601), Gesamtbruttoumsatz, `BON_STORNO` |
| `transactions_vat.csv`  | Bonkopf_USt          | USt-Aufschlüsselung pro Bon nach Steuerschlüsseln (Brutto, Netto, USt)                                                           |
| `allocation_groups.csv` | Bonkopf_AbrKreis     | Zuordnung eines Bons zu einem `ABRECHNUNGSKREIS` (Tisch-Verknüpfung im Festzelt)                                                 |
| `datapayment.csv`       | Bonkopf_Zahlarten    | Zahlungsarten pro Bon (Bar, EC-Karte, Kreditkarte)                                                                               |
| `references.csv`        | Bon_Referenzen       | Referenzen auf andere Bons, u. a. `REF_BON_ID` bei Stornos                                                                       |
| `lines.csv`             | Bonpos               | Einzelne Artikel: `POS_ZEILE`, `ART_NR`, `MENGE` (Dezimal, 3 Nachstellen), `EINHEIT`, `STK_BR` (Stückpreis brutto)               |
| `lines_vat.csv`         | Bonpos_USt           | USt-Aufschlüsselung pro Artikelzeile                                                                                             |
| `transactions_tse.csv`  | TSE_Transaktionen    | **Kritisch:** TSE-Transaktionsnummer (`TSE_TANR`), Signaturzähler (`TSE_TA_SIGZ`), Krypto-Signatur (`TSE_TA_SIG`)                |

#### C. Kassenabschlussmodul (Z-Bon)

| Dateiname (offiziell)   | Logische Bezeichnung | Inhalt                                                         |
| ----------------------- | -------------------- | -------------------------------------------------------------- |
| `businesscases.csv`     | Z_GV_Typ             | Aggregierte Beträge pro Geschäftsvorfall-Typ nach Steuersätzen |
| `payment.csv`           | Z_Zahlart            | Aggregierte Summen der Zahlungsarten (Bar vs. unbar)           |
| `cash_per_currency.csv` | Z_Waehrungen         | Bargeldbestand je Währung zum Abschluss                        |

### 6.4 Schlüsselfelder in `transactions.csv` (Bonkopf, Relationale Verknüpfung)

Fast jede CSV-Datei muss folgende Schlüssel in den ersten Spalten mitführen:

1. **`Z_KASSE_ID`** — Eindeutige Kassen-ID
2. **`Z_ERSTELLUNG`** — Zeitstempel des zugehörigen Kassenabschlusses
3. **`Z_NR`** — Fortlaufende Z-Bon-Nummer
4. **`BON_ID`** — Eindeutige Vorgangs-ID des Bons

### 6.5 Abrechnungskreis — Tisch-Verknüpfung für Festzelt

Das Feld `ABRECHNUNGSKREIS` (in `allocation_groups.csv`, logisch _Bonkopf_AbrKreis_, je `BON_ID`) verknüpft mehrere Bons (Bestellungen + Zahlungen) zu einer logischen Einheit. Im Festzelt-Betrieb trägt jede Bestellung und jede Zahlung für denselben Tisch innerhalb einer Kassensitzung denselben `ABRECHNUNGSKREIS`-Wert (= Tischname):

```
BON_ID | BON_TYP      | BON_START            | ABRECHNUNGSKREIS
-------|--------------|----------------------|--------------------
1001   | Bestellung   | 2026-05-01T18:01:00  | Tisch 42
1002   | Bestellung   | 2026-05-01T19:30:00  | Tisch 42
1003   | Beleg        | 2026-05-01T20:00:12  | Tisch 42
1004   | Beleg        | 2026-05-01T21:00:05  | Tisch 42
```

Die Prüfsoftware IDEA der Finanzämter kann so den vollständigen Tischverlauf rekonstruieren, auch wenn jede Transaktion für sich als sofort geschlossen registriert ist.

#### Session-Grenzen

Der `ABRECHNUNGSKREIS` wird **pro Tisch und Kassensitzung** vergeben. Intern bildet das Subject `kassensitzung-{nr}/tisch-{id}` die Tisch-Session ab; im DSFinV-K-Export wird der **Tischname** als `ABRECHNUNGSKREIS` verwendet (z. B. `Tisch 42`).

> **Wichtiger Hinweis:** Jeder Tisch bekommt seinen eigenen `ABRECHNUNGSKREIS`. Ein Gesamt-Schlüssel für alle Tische wäre ein Verstoß gegen die GoBD-Anforderung der Nachvollziehbarkeit.

**Aktueller Stand (✅ Umgesetzt):**

- Der `ABRECHNUNGSKREIS` wird **einmal pro Tisch und Kassensitzung** vergeben: der Tischname (z. B. `Tisch 42`).
- Intern: Subject `kassensitzung-{nr}/tisch-{id}` im Kassenjournal.
- Beim **Tagesabschluss** (Z-Bon) wird die Kassensitzung geschlossen und alle zugehörigen Tisch-Sessions sind abgeschlossen.
- Neue Sessions entstehen erst wieder durch eine explizite Kassensitzungseröffnung (Admin-Aktion).
- Nachteil: Mehrere Gästegruppen am gleichen Tisch innerhalb einer Kassensitzung teilen denselben `ABRECHNUNGSKREIS` — für den Betriebsprüfer erkennbar, aber zulässig, solange alle Bons korrekt verknüpft sind.

**Geplant — Manuelle Tischfreigabe (empfohlene Erweiterung für Festzelt-Betrieb):**

- Servicekräfte können einen Tisch explizit für neue Gäste freigeben ("Tisch freimachen").
- Eine neue Session erhält dann ein Suffix: `Tisch 42-A`, `Tisch 42-B`, etc.
- In jottis Festzelt-Szenario, wo häufig mehrere Gästegruppen am selben Tisch sitzen (Maihock, Vereinsfest), ist dies die korrektere Abbildung der Realität und reduziert Rückfragen bei Betriebsprüfungen.
- Erfordert eine neue UI-Aktion in der Servicekraft-Ansicht und eine entsprechende Domain-Aktion in der Tisch-Session.

> **Hinweis:** Das DSFinV-K-Format erlaubt beliebige Strings als `ABRECHNUNGSKREIS` (max. 40 Zeichen). Das Format `Tisch {Name}` (aktuell) bzw. `Tisch {Name}-{Buchstabe}` (manuelle Tischfreigabe) ist eine jotti-interne Konvention.

### 6.6 Storno-Handling in DSFinV-K

#### Positions-Storno (Storno vor Zahlung)

Ein falsch gebuchtes Produkt (z.B. 5 statt 4 Maß Bier) wird als neuer Bon mit **negativer Menge** gebucht:

- Neuer Bon mit `BON_TYP = "AVBelegstorno"` oder negativen Positionsmengen
- **`BON_STORNO = 1`** in `transactions.csv`
- **`REF_BON_ID`** in `references.csv` = `BON_ID` des fehlerhaften Ursprungsbons
- Eigene TSE-Signatur (`Kassenbeleg-V1` oder `Bestellung-V1` mit negativen Werten)
- `ABRECHNUNGSKREIS` identisch mit dem stornierten Vorgang

#### Bon-Storno (Storno nach Zahlung)

Eine bereits bezahlte Rechnung wird mit einem neuen Beleg mit negativen Beträgen storniert:

- Neuer Bon mit negativem Gesamtbetrag
- **`BON_STORNO = 1`** in `transactions.csv`
- **`REF_BON_ID`** (in `references.csv`) = `BON_ID` des Original-Zahlungsbelegs
- Eigene TSE-Transaktion (`Kassenbeleg-V1` mit negativem Betrag)

> **GoBD-Grundsatz:** Datensätze dürfen niemals per `UPDATE` oder `DELETE` aus der Datenbank entfernt werden. Stornierungen erzeugen immer neue Datensätze, die den Ursprungswert ausgleichen (Append-Only-Prinzip).

### 6.7 Architektonische Anforderungen an jotti

1. **CSV-Generator:** Komponente, die aus den Event-Store-Daten und Stammdaten die DSFinV-K-CSV-Struktur erzeugt (mit den offiziellen, englischen Dateinamen, z. B. `transactions.csv`, `lines.csv`)
2. **`index.xml`-Generator:** Metadaten-Datei für die Prüfsoftware
3. **Z-Bon-Logik:** Kassenabschluss-Funktion, die Tagessummen aggregiert und einen Z-Bon erzeugt
4. **Abrechnungskreis-Verwaltung:** Tisch-Session-ID persistieren und in allen zugehörigen Bons mitführen (Phase 1: tagesbasiert; Phase 2: manuelle Freigabe)
5. **Admin-Endpunkt:** API-Endpunkt zum Auslösen des Exports (z.B. `POST /admin/dsfinvk-export`)
6. **ZIP-Generierung:** Alle CSVs + `index.xml` in ein ZIP-Archiv verpacken
7. **Steuersatz-Verwaltung:** USt-Sätze müssen als Stammdaten gepflegt werden — zu unterstützende Sätze: **19 %** (Standardsatz, z.B. Getränke), **7 %** (ermäßigt, z.B. Speisen), **0 % / steuerbefreit** (Zweckbetrieb nach § 67a AO). Produkte erhalten einen konfigurierbaren Steuersatz-Schlüssel.

---

## 7. Elektronische Meldepflicht (ERiC / ELSTER)

### 7.1 Gesetzliche Grundlage und Fristen

Nach § 146a Abs. 4 AO müssen elektronische Aufzeichnungssysteme beim zuständigen Finanzamt gemeldet werden. Das Mitteilungsverfahren ist seit dem 1. Januar 2025 aktiv.

| Datum                                    | Ereignis                                                                      |
| ---------------------------------------- | ----------------------------------------------------------------------------- |
| 1. Januar 2025                           | Meldeportal öffnet; Mitteilungsverfahren ist aktiv                            |
| 31. Juli 2025                            | Abgabefrist für alle Systeme, die **vor dem 1. Juli 2025** angeschafft wurden |
| Innerhalb 1 Monat nach Anschaffung       | Abgabefrist für Systeme, die **ab dem 1. Juli 2025** neu angeschafft werden   |
| Innerhalb 1 Monat nach Außerbetriebnahme | Abgabefrist bei Stilllegung eines Systems                                     |

Systeme, die bereits vor dem 1. Juli 2025 außer Betrieb genommen wurden, müssen nicht gemeldet werden.

_(Quelle: § 146a Abs. 4 AO — https://www.gesetze-im-internet.de/ao_1977/__146a.html)_
_(Quelle: BMF-Schreiben 28. Juni 2024 — Meldepflicht)_

### 7.2 Drei Übermittlungswege

1. **Direkteingabe im ELSTER-Web-Portal** (`www.elster.de`, „Mein ELSTER") — manuell, für Einzelfälle
2. **XML-Dateiupload** im ELSTER-Portal — semi-automatisch
3. **Programmatische Übermittlung über ERiC** (ELSTER Rich Client) — vollautomatisch aus der Kassensoftware heraus

ERiC ist die offizielle Softwarekomponente der Finanzverwaltung für die maschinelle Übermittlung. Sie validiert die Daten lokal, bevor sie an das Finanzamt übermittelt werden, und erzeugt bei Erfolg ein offizielles Bestätigungsprotokoll.

**Gewählter Ansatz für jotti:**

Da jotti ein **Self-hosted-Produkt** ist (kein SaaS), das von ehrenamtlichen Vereinen betrieben wird, wird die Kassenmeldung in zwei Phasen umgesetzt:

- **Phase 1 (manuell):** Der Admin-Bereich stellt alle meldepflichtigen Daten (Steuernummer, Kassen-ID, TSE-Seriennummer, Inbetriebnahmedatum etc.) strukturiert bereit. Der Vereinsvorstand meldet manuell über das ELSTER-Webportal. Das App nennt die benötigten Felder und deren Werte klar — kein Code-Aufwand.
- **Phase 2 (optional, automatisiert):** Integration entweder via ERiC (native C-Library, aufwendig, kein Vendor-Lock-in) oder fiskalys Submission-API (einfacher, Cloud-Abhängigkeit). Entscheidung fällt bei Implementierungsbeginn durch Aufwands-/Kosten-Abwägung.

_(Quelle: ELSTER für Entwickler — https://www.elster.de/elsterweb/infoseite/entwickler)_

### 7.3 Submission-API (kommerzielle Alternative)

Alternativ bieten TSE-Anbieter wie fiskaly eine **Submission-API** als Abstraktionsschicht an:

- **Vorteil:** Keine direkte ERiC-Integration nötig; Kommunikation über Cloud-API
- **Nachteil:** Keine staatliche Vorab-Validierung; Verantwortung für Datenqualität liegt bei der Software
- **Abhängigkeit:** Vendor-Lock-in zum TSE-Anbieter

### 7.4 Meldepflichtige Daten (Payload)

Die Kassenmeldung umfasst je Kassensystem:

- Name und Steuernummer des Steuerpflichtigen
- Art des Kassensystems (Softwaretyp, Versionsnummer)
- Seriennummer des Kassensystems
- Zertifizierungs-ID der TSE (Format: `BSI-K-TR-nnnn-yyyy`)
- Seriennummer der TSE (64-stelliger Hexadezimalstring, ausschließlich 0–9 und A–F; **nicht** Base64 — Hinweis: in `tse.csv` (logisch _Stamm_TSE_) für den DSFinV-K-Export wird zusätzlich der Public Key als Base64-Wert gespeichert, hierbei handelt es sich um ein anderes Feld)
- Anschaffungs- bzw. Inbetriebnahmedatum
- Betriebsstättenadresse

### 7.5 Architektonische Anforderungen an jotti

1. **Konfiguration:** Vereinsdaten, Betriebsstätte, Steuernummer als Stammdaten (bereits für Phase-1-Meldung nötig)
2. **Admin-Datenanzeige (Phase 1):** Admin-Bereich zeigt alle meldepflichtigen Felder strukturiert an, damit der Admin die manuelle ELSTER-Meldung ohne Suche durchführen kann; kein API-Aufruf, nur Datenanzeige — inkl. der automatisch generierten Kassen-Seriennummer (siehe Abschnitt 3.7)
3. **Meldestatus (Phase 1):** Manuell setzbarer Status in der Admin-UI (z.B. „ausstehend / gemeldet am TT.MM.JJJJ / Fehler") — persistiert in den Stammdaten
4. **Automatisierte Übermittlung (Phase 2, optional):** Entweder ERiC-Integration oder fiskaly-Submission-API; Entscheidung bei Implementierungsbeginn

### 7.6 Verantwortungsverteilung im Source-Available-Self-hosted-Modell

Die Meldepflicht nach § 146a Abs. 4 AO liegt **ausschließlich beim Betreiber (Verein)**, nicht beim Entwickler der Software. Dies hat folgende praktische Konsequenzen:

#### BYOD-Smartphones: Keine Meldepflicht

Die Smartphones der Servicekräfte müssen dem Finanzamt **nicht gemeldet** werden. Der Anwendungserlass zur Abgabenordnung (AEAO) zu § 146a AO stellt klar: Wenn Systeme ohne eigene Kassenfunktion (Smartphones als Eingabegeräte) mit einem elektronischen Aufzeichnungssystem mit Kassenfunktion (jotti Backend) verbunden sind, ist ausschließlich das **Hauptsystem** mitteilungspflichtig. Das bedeutet: Vereine müssen nicht hunderte wechselnde private Handys von Aushilfskellnern beim Finanzamt an- und abmelden.

#### Self-hosted Docker-Instanz: Verein meldet eigenständig

- Jeder Verein, der jotti in eigener Regie auf einem VPS betreibt, meldet **seine Docker-Instanz** (die Seriennummer der Kasse) beim Finanzamt an.
- Die für die Meldung benötigte Seriennummer wird von jotti automatisch generiert und im Admin-Bereich angezeigt (siehe Abschnitt 3.7).
- **Der Entwickler hat keine Meldepflicht** — er stellt nur das gesetzeskonforme Werkzeug bereit.

#### Zusammenfassung der Verantwortlichkeiten

| Pflicht                                      | Entwickler (jotti) | Verein (Betreiber)          |
| -------------------------------------------- | ------------------ | --------------------------- |
| TSE-Schnittstelle im Code implementieren     | ✅ Pflicht         | —                           |
| DSFinV-K-Export im Code implementieren       | ✅ Pflicht         | —                           |
| Muster-Verfahrensdokumentation bereitstellen | Empfohlen          | —                           |
| Cloud-TSE-Vertrag abschließen                | —                  | ✅ Pflicht                  |
| TSE-API-Keys konfigurieren (`.env`)          | —                  | ✅ Pflicht                  |
| ELSTER-Meldung (§ 146a Abs. 4 AO)            | ❌ Keine Pflicht   | ✅ Pflicht (Frist: 1 Monat) |
| BYOD-Smartphones melden                      | —                  | ❌ Nicht erforderlich       |
| 10-Jahres-Archivierung (GoBD)                | —                  | ✅ Pflicht                  |
| Server-Betrieb und Datensicherung            | —                  | ✅ Pflicht                  |

_(Quelle: AEAO zu § 146a AO — Klarstellung zur Mitteilungspflicht für verbundene Eingabegeräte — https://www.bundesfinanzministerium.de/)_
_(Quelle: § 146a Abs. 1 Satz 5 AO — Verbot des In-Verkehr-Bringens nicht-TSE-fähiger Kassensoftware — https://www.gesetze-im-internet.de/ao_1977/__146a.html)_

---

## 8. Betreiberpflichten

Die Vereine tragen als Betreiber die volle operative und rechtliche Verantwortung für ihre jotti-Instanz. Eine detaillierte Verantwortlichkeitstabelle findet sich in Abschnitt 7.6.

### Pflichten vor dem ersten Einsatz

1. **Cloud-TSE-Vertrag (BYOT):** Vertrag mit Cloud-TSE-Anbieter (z. B. fiskaly oder D-Trust) abschließen. API-Schlüssel als Umgebungsvariablen in die `.env`-Datei eintragen.
2. **ELSTER-Meldung:** Nach der ersten Inbetriebnahme innerhalb von **einem Monat** die jotti-Instanz beim zuständigen Finanzamt über [ELSTER](https://www.elster.de) anmelden (§ 146a Abs. 4 AO). Benötigte Daten: Seriennummer der Kasse (im Admin-Dashboard), Softwarename „jotti", Inbetriebnahmedatum.
3. **Seriennummer sichern:** Die Kassen-UUID in den System-Stammdaten ist die rechtliche Identität der Kasse. Das Datenbank-Backup muss diese enthalten. Bei Verlust: alte Seriennummer beim Finanzamt abmelden, neue Instanz mit neuer Seriennummer neu anmelden.

### Laufende Pflichten

- **10-Jahres-Aufbewahrung:** Alle Kassendaten (Kassenjournal, DSFinV-K-Exporte) sind 10 Jahre aufzubewahren (§§ 146, 147 AO, GoBD). Sicherstellen, dass Backups entsprechend archiviert und jederzeit lesbar sind.
- **Regelmäßige Backups:** Tägliche Datenbank-Backups sind Pflicht — nicht nur für die Compliance, sondern auch zur Seriennummern-Sicherung.
- **Server-Betrieb:** Verfügbarkeit, Datensicherung, Zugriffsschutz und 10-jährige GoBD-konforme Aufbewahrung der Daten liegen beim Verein.
- **Außerbetriebnahme melden:** Wenn eine jotti-Instanz dauerhaft stillgelegt wird, muss dies innerhalb von einem Monat bei ELSTER gemeldet werden.
- **BYOD-Smartphones:** Müssen dem Finanzamt **nicht** gemeldet werden (AEAO zu § 146a AO: Eingabegeräte ohne eigenständige Kassenfunktion).

---

## 9. Architekturprinzipien

→ **[handbuch.md §3.13 — TSE-Architektur](handbuch.md#313-tse-architektur)**

Enthält: TSEClient-Interface, Transaktions-Mapping (jotti-Vorgang → processType), Event-Store-Extension (TSE-Felder), DSFinV-K-Exporter-Übersicht, Entscheidungsmatrix Cloud-TSE vs. ERiC.

---

## 10. Quellenverzeichnis

| #   | Quelle                                                                                                                   | URL                                                                                                                                 |
| --- | ------------------------------------------------------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------- |
| 1   | § 146a AO — Ordnungsvorschrift für die Buchführung und für Aufzeichnungen mittels elektronischer Aufzeichnungssysteme    | https://www.gesetze-im-internet.de/ao_1977/__146a.html                                                                              |
| 2   | KassenSichV — Kassensicherungsverordnung                                                                                 | https://www.gesetze-im-internet.de/kassensichv/BJNR351500017.html                                                                   |
| 3   | BSI TR-03153 — Technische Richtlinie für Technische Sicherheitseinrichtungen                                             | https://www.bsi.bund.de/SharedDocs/Downloads/DE/BSI/Publikationen/TechnischeRichtlinien/TR03153/TR-03153.pdf?__blob=publicationFile |
| 4   | GoBD — BMF-Schreiben zur ordnungsmäßigen Führung elektronischer Bücher                                                   | https://www.bundesfinanzministerium.de/Content/DE/Downloads/BMF_Schreiben/Weitere_Steuerthemen/Abgabenordnung/2019-11-28-GoBD.html  |
| 5   | DSFinV-K — Digitale Schnittstelle der Finanzverwaltung für Kassensysteme (BZSt)                                          | https://www.bzst.de/DE/Unternehmen/Aussenpruefungen/DigitaleSchnittstelleFinV/digitaleschnittstellefinv_node.html                   |
| 6   | ELSTER für Entwickler — Offizielle Entwickler-Dokumentation                                                              | https://www.elster.de/elsterweb/infoseite/entwickler                                                                                |
| 7   | § 14 AO — Wirtschaftlicher Geschäftsbetrieb                                                                              | https://www.gesetze-im-internet.de/ao_1977/__14.html                                                                                |
| 8   | § 64 AO — Steuerpflicht wirtschaftlicher Geschäftsbetriebe                                                               | https://www.gesetze-im-internet.de/ao_1977/__64.html                                                                                |
| 9   | § 67a AO — Sportliche Veranstaltungen (Zweckbetrieb)                                                                     | https://www.gesetze-im-internet.de/ao_1977/__67a.html                                                                               |
| 10  | § 19 UStG — Kleinunternehmerregelung                                                                                     | https://www.gesetze-im-internet.de/ustg_1980/__19.html                                                                              |
| 11  | BMF-FAQ zu § 146a AO (Stand Januar 2026) — Meldepflicht und processType-Erläuterungen                                    | https://www.bundesfinanzministerium.de/                                                                                             |
| 12  | BMF-Schreiben 28. Juni 2024 — Elektronische Kassenmeldepflicht nach § 146a Abs. 4 AO                                     | https://www.bundesfinanzministerium.de/                                                                                             |
| 13  | DSFinV-K Nr. 2.7 und Anhang H — Vereinfachungen für langanhaltende Bestellvorgänge (Festzelt-/Durchbedienen-Muster)      | https://www.bzst.de/DE/Unternehmen/Aussenpruefungen/DigitaleSchnittstelleFinV/digitaleschnittstellefinv_node.html                   |
| 14  | § 379 AO — Steuergefährdung (bis 25.000 € Bußgeld)                                                                       | https://www.gesetze-im-internet.de/ao_1977/__379.html                                                                               |
| 15  | AEAO zu § 146a AO — Anwendungserlass zur Abgabenordnung, Klarstellungen zu Eingabegeräten, Meldepflicht und Seriennummer | https://www.bundesfinanzministerium.de/                                                                                             |
