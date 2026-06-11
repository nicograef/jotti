# Compliance-Anforderungen: Fiskalische Grundlagen für jotti

> **Betrifft:** KassenSichV, TSE, GoBD, Belegausgabepflicht, DSFinV-K, ERiC/ELSTER

## 1. Einleitung

jotti ist ein **elektronisches Aufzeichnungssystem** (§ 1 KassenSichV) und unterliegt nach § 146a AO der TSE-Pflicht — unabhängig von Rechtsform, Gemeinnützigkeit oder Veranstaltungsdauer. Dieses Dokument beschreibt die Rechtsnormen, Entwickler- und Betreiberpflichten sowie die Compliance-Architektur. Technische Umsetzung phasenweise — siehe [anforderungen.md](anforderungen.md); Architektur-Entscheidungen — siehe [handbuch.md §3.13](handbuch.md#313-tse-architektur).

---

## 2. Rechtliche Grundlagen

### 2.1 Abgabenordnung (AO) — § 146a

§ 146a Abs. 1 AO verpflichtet jeden Betreiber eines elektronischen Aufzeichnungssystems, jeden Geschäftsvorfall einzeln, vollständig, richtig, zeitgerecht und geordnet zu erfassen und durch eine **zertifizierte TSE** zu schützen — unabhängig von Rechtsform, Gemeinnützigkeit, Betriebsdauer oder Gewinnerzielungsabsicht. [1]

Für jottis Direktverkauf (K-24) gilt die Einzelaufzeichnung strikt pro Event: `direktverkauf-getaetigt:v1` und `direktverkauf-storniert:v1` sind eigenständige Geschäftsvorfälle — der Storno korrigiert durch Gegeneintrag, nie durch Änderung des Verkaufsdatensatzes. Der fiskalische Beleg entsteht wie am Tisch on-demand: `POST /service/beleg-drucken` mit `verkaufId` erzeugt genau einen Kassenbeleg-Druckauftrag auf Basis des referenzierten Events — der Geschäftsvorfall bleibt 1:1 referenzierbar.

### 2.2 Kassensicherungsverordnung (KassenSichV)

Die KassenSichV konkretisiert § 146a AO; § 1 KassenSichV definiert als Aufzeichnungssysteme „elektronische oder computergestützte Kassensysteme oder Registrierkassen" — jotti fällt als browserbasiertes Kassensystem eindeutig darunter.

Für **mobile Geräte** gilt: Kann ein Gerät Zahlungsvorgänge eigenständig und offline erfassen, braucht es eine eigene TSE-Anbindung. Fungiert es nur als Eingabeterminal mit sofortiger Weiterleitung an ein TSE-gesichertes Backend, genügt die Backend-TSE. In jottis BYOD-Modell sind die Smartphones der Servicekräfte **reine Eingabegeräte** — rechtlich nicht mehr als eine Tastatur. Konsequenzen:

1. Die Smartphones benötigen **keine eigene TSE**; TSE-Absicherung, Protokollierung und DSFinV-K-Speicherung erfolgen zentral im Backend.
2. **Architektonische Pflicht:** Die Webapp muss bei Verbindungsverlust **sofort blockieren** — jede Offline-Erfassung von Barzahlungen muss technisch verhindert sein. Nur so ist die Einordnung als reines Eingabegerät rechtlich haltbar. [2, 11, 15]

### 2.3 GoBD

Die GoBD (BMF-Schreiben 28.11.2019) fordern Nachvollziehbarkeit, Vollständigkeit, Richtigkeit, Zeitgerechtigkeit, Ordnung und Unveränderbarkeit. Sie gelten für alle Steuerpflichtigen — einschließlich gemeinnütziger Vereine im wirtschaftlichen Geschäftsbetrieb. [4]

### 2.4 Belegausgabepflicht (§ 146a Abs. 2 AO, § 6 KassenSichV)

§ 146a Abs. 2 AO und § 6 KassenSichV verpflichten zur Belegausgabe in **unmittelbarem zeitlichen Zusammenhang** nach jedem Kassiervorgang — in Papierform oder (mit Zustimmung des Kunden) elektronisch. [1] Details und Festzelt-Befreiung: Abschnitt 5.

### 2.5 DSFinV-K

§ 4 KassenSichV verlangt eine **einheitliche digitale Schnittstelle** für den Datenexport an die Finanzverwaltung. Die DSFinV-K (Version 2.4, Stand Januar 2024) definiert das Format: CSV-Dateien mit fest vorgeschriebenen (englischen, kleingeschriebenen) Dateinamen, Semikolon-Trennung und `index.xml`, verpackt als ZIP. [5] Details: Abschnitt 6.

### 2.6 Elektronische Kassenmeldepflicht (§ 146a Abs. 4 AO)

Seit dem 1. Januar 2025 müssen elektronische Aufzeichnungssysteme dem zuständigen Finanzamt **elektronisch** gemeldet werden — über ELSTER, manuell im Portal oder programmatisch via ERiC. Fristen und Umsetzung: Abschnitt 7. [1, 11]

### 2.7 Sonderfall: Source-Available Self-hosted — Pflichten des Entwicklers

jotti ist kein SaaS: Der Code ist öffentlich auf GitHub, die Vereine betreiben das System eigenverantwortlich (VPS, Docker). Das trennt die Rechtspflichten von **Entwickler (Hersteller)** und **Betreibern (Vereinen)**.

**Pflichten des Entwicklers:** Nach **§ 146a Abs. 1 Satz 5 AO** i.V.m. **§ 379 AO** ist es verboten, Kassensoftware in Verkehr zu bringen, die nicht über die **Möglichkeit** verfügt, eine zertifizierte TSE anzubinden. Auch das kostenlose Bereitstellen von Code auf GitHub gilt als „In-Verkehr-Bringen". Daraus folgt:

- **TSE-Schnittstelle** (`TSEClient`-Interface) und **DSFinV-K-Export** müssen im Code vorhanden und nutzbar sein.
- Entscheidet ein Verein, keinen TSE-API-Key einzutragen, liegt das rechtliche Risiko **ausschließlich beim Verein**.
- Der Entwickler hat **keine Meldepflicht** für fremdbetriebene Installationen — mitteilungspflichtig ist nur, wer das System tatsächlich verwendet.
- Eine **Muster-Verfahrensdokumentation** im Repository ist empfehlenswert (Architektur, Datenschutz, TSE-Anbindung für Betriebsprüfer).

**Pflichten der Betreiber:** → Abschnitt 8.

---

## 3. TSE-Integration (Technische Sicherheitseinrichtung)

### 3.1 Hintergrund

Die TSE ist das kryptografische Herzstück eines konformen Kassensystems. Sie besteht aus drei Modulen: **Sicherheitsmodul** (SMAERS-Datenaufbereitung + CSP-Signatur), **Speichermedium** (lokale Sicherung der signierten Daten) und **Einheitliche Digitale Schnittstelle** (EDS, Export). [3]

### 3.2 Protokollierungs-Ablauf (Transaktions-Lebenszyklus)

Jeder abzusichernde Vorgang kommuniziert mit der TSE über drei Operationen:

| Operation | Wann | Request | Response |
| --- | --- | --- | --- |
| **`StartTransaction`** | **Unmittelbar** bei Vorgangsbeginn (Ordnungsanforderung; keine maximale Transaktionsdauer in BSI TR-03153) | Kassen-ID — `processType` und `processData` sind bei Start **immer leer** (DSFinV-K Anhang I) | Fortlaufende Transaktionsnummer, `logTime`, TSE-Seriennummer, Signaturzähler |
| **`UpdateTransaction`** | Optional — **nur** für `Bestellung-V1` und `SonstigerVorgang`; für `Kassenbeleg-V1` **verboten** (processData erst bei Abschluss bekannt) [11] | Kassen-ID, Transaktionsnummer, aktualisierte `processData` | — |
| **`FinishTransaction`** | Bei Abschluss oder Abbruch des Vorgangs | Transaktionsnummer, finale `processData` (Summen nach Steuersätzen und Zahlungsarten) | Signatur, `logTime`, finaler Signaturzähler |

### 3.3 Offizielle processType-Werte

Die `processType`-Werte sind im AEAO zu § 146a AO, Anhang I, festgelegt. Die **-V1-Endung ist nur bei `Kassenbeleg-V1` und `Bestellung-V1` Bestandteil des offiziellen Strings** — der dritte Typ heißt `SonstigerVorgang` (ohne Suffix).

| processType        | Verwendung                                                                                                                 |
| ------------------ | --------------------------------------------------------------------------------------------------------------------------- |
| `Kassenbeleg-V1`   | Zahlungsbeleg (Rechnung), der dem Kunden ausgehändigt wird; auch Eigenbelege über Ein-/Auszahlungen wie Geldtransit, Kassendifferenz und Auszahlung (AEAO 2.2.3.6.1) |
| `Bestellung-V1`    | Zwischenabsicherung einer Bestellung ohne sofortige Zahlung (Gastronomie)                                                  |
| `SonstigerVorgang` | Alle anderen abzusichernden Vorgänge (Tagesabschluss, TSE-Selbsttest, ...)                                                 |

**Mapping auf jotti-Events:**

| jotti-Event                  | processType           | Anmerkung                                                                                     |
| ---------------------------- | --------------------- | --------------------------------------------------------------------------------------------- |
| `bestellung-aufgenommen:v1`  | `Bestellung-V1`       | Sofort geschlossen (Festzelt-Muster, §3.6)                                                    |
| `zahlung-kassiert:v1`        | `Kassenbeleg-V1`      | Tisch-Teilzahlung oder Vollzahlung                                                            |
| `direktverkauf-getaetigt:v1` | `Kassenbeleg-V1`      | Atomar — keine vorgelagerte `Bestellung-V1`, da Bestellung und Zahlung zeitgleich stattfinden |
| `direktverkauf-storniert:v1` | `Kassenbeleg-V1`      | Stornobeleg mit negativem Betrag, `REF_BON_ID` auf den Ursprungsverkauf                       |
| `stornierung-erteilt:v1`     | `Kassenbeleg-V1`      | Positions-Storno am Tisch mit negativen Beträgen                                              |
| Tagesabschluss (Z-Bon)       | `SonstigerVorgang`    | Aggregierte Tagessummen                                                                       |

Das vollständige Mapping aller jotti-Vorgänge (inkl. Auszahlung, Geldtransit, Kassendifferenz): [handbuch.md §3.13](handbuch.md#313-tse-architektur).

### 3.4 Datenformat-Vorgaben (`processData`)

- **Encoding:** UTF-8 oder ASCII, kein BOM
- **Dezimaltrennzeichen:** ausschließlich Punkt (`.`); keine Tausendertrennzeichen, keine Exponentialschreibweise, kein `+`; mindestens eine Stelle vor dem Punkt (`0.5`, nicht `.5`)
- **Bei `StartTransaction`:** `processType` und `processData` sind **immer leer** — beide werden erst bei `FinishTransaction` übergeben (DSFinV-K Anhang I)
- **Format `Kassenbeleg-V1`:** `Beleg^<Betraege>^<Zahlungen>` — die fünf Brutto-Steuerbeträge in fester Reihenfolge: **1.** Allgemeiner Steuersatz (19 %) · **2.** Ermäßigter Steuersatz (7 %) · **3.** Durchschnittssatz § 24 Abs. 1 Nr. 3 UStG (10,7 %) · **4.** Durchschnittssatz § 24 Abs. 1 Nr. 1 UStG (5,5 %) · **5.** 0 %. Zahlungen als `<Betrag>:<Zahlungsart>`; **Zahlungen von `0.00` entfallen**.
  > **Kombi-Positionen (70/30):** Speisen-Anteil (70 %) → Position 2 (ermäßigt), Getränke-Anteil (30 %) → Position 1 (allgemein); steuerbefreite Positionen → Position 5 (0 %). Die fünf Betragstellen bilden alle jotti-Steuersätze ab.
- **Format `Bestellung-V1`:** CSV-Darstellung, pro Position eine Zeile `<Menge>;"<Bezeichnung>";<Brutto-Einzelpreis>` — Zeilentrenner `\r` (U+000D), Bezeichnung in Anführungszeichen (innere `"` werden verdoppelt), Preis mit exakt 2 Nachkommastellen. Beispiel aus DSFinV-K Anhang I: `2;"Eisbecher ""Himbeere""";3.99`

### 3.5 TSE-Varianten und Anbieter-Entscheidung

| Variante         | Beschreibung                                       | Beispiel-Anbieter                |
| ---------------- | -------------------------------------------------- | -------------------------------- |
| **Hardware-TSE** | Physisches Gerät (USB-Stick, SD-Karte, Smartcard)  | Swissbit, Epson, Diebold Nixdorf |
| **Cloud-TSE**    | TSE als Cloud-Service, Kommunikation via HTTPS-API | fiskaly, Deutsche Fiskal         |

Für jotti als Self-hosted-System ist die **Cloud-TSE** gesetzt — eine Hardware-TSE scheidet für BYOD-Setups auf gemieteten Servern praktisch aus. **Gewählter erster Zielanbieter: fiskaly** (API-first, BSI-zertifiziert nach TR-03153, unterstützt alle drei processTypes, bietet optional eine Submission-API für die Kassenmeldung). Das Backend-Interface `TSEClient` bleibt **anbieter-agnostisch** (Adapter-Pattern) — ein Anbieterwechsel erfordert keine Änderung am Domain-Code.

### 3.6 Das Festzelt-Muster: Atomare TSE-Transaktionen

Im Festzelt-Betrieb laufen an einem Tisch über Stunden Bestellrunden und Teilzahlungen auf. Eine einzige TSE-Transaktion so lange offenzuhalten wäre technisch riskant und ist laut BMF-FAQ nicht erforderlich. Die DSFinV-K dokumentiert in Nr. 2.7 und Anhang H eine **Vereinfachungsregelung für langanhaltende Bestellvorgänge** („Erleichterungsregelung beim Durchbedienen"): Jede atomare Aktion ist eine **eigene, sofort geschlossene** TSE-Transaktion. [13]

- **Schicht 1 — Bestellabsicherung:** jede Bestellrunde als `StartTransaction(processType="Bestellung-V1")` → sofort `FinishTransaction`; `processData` enthält die Positionen.
- **Schicht 2 — Zahlungsabsicherung:** jede Teil- oder Vollzahlung als `StartTransaction(processType="Kassenbeleg-V1")` → sofort `FinishTransaction`; `processData` enthält Gesamtbetrag, Steuersätze, Zahlungsart.
- **Verknüpfung:** Alle Transaktionen eines Tisches teilen im DSFinV-K-Export denselben `ABRECHNUNGSKREIS` (→ §6.5).

**Szenario Maihock, Tisch 42** — jeder Vorgang eine eigene, sofort geschlossene Transaktion, alle mit `ABRECHNUNGSKREIS = "Tisch 42"`:

| Zeit  | Vorgang                           | TSE-Transaktion             |
| ----- | --------------------------------- | --------------------------- |
| 18:01 | Bestellung 4× Bier, 2× Weißwurst  | `Bestellung-V1` (Nr. 1001)  |
| 19:30 | Nachbestellung 4× Bier            | `Bestellung-V1` (Nr. 1002)  |
| 20:00 | Teilzahlung 14,00 € bar (2 Gäste) | `Kassenbeleg-V1` (Nr. 1003) |
| 21:00 | Restzahlung 18,00 € bar (2 Gäste) | `Kassenbeleg-V1` (Nr. 1004) |

Jeder Zahlungsbeleg trägt zusätzlich den Startzeitpunkt der ersten Bestellung in Klarschrift (Durchbedienen-Pflicht → §5.3). Setzt sich eine neue Gästegruppe an denselben Tisch, beginnt ein neuer Abrechnungskreis (z. B. `Tisch 42-B`). Der Betriebsprüfer rekonstruiert den vollständigen Tischverlauf über den gemeinsamen `ABRECHNUNGSKREIS`.

### 3.7 Seriennummer-Generierung bei Self-hosted Docker-Instanzen

Ohne physische Kassenhardware gibt es kein Typenschild — die gesetzlich geforderte eindeutige Seriennummer (§ 146a AO, § 6 KassenSichV, DSFinV-K) wird softwareseitig erfüllt:

- **Beim allerersten Start** (Datenbank-Initialisierung) generiert jotti eine UUID als Kassen-Seriennummer (herstellerunabhängig, weltweit eindeutig, keine zentrale Registrierung nötig), speichert sie dauerhaft und überschreibt sie **nie** — auch nicht bei Updates oder Neustarts.
- Die Seriennummer wird im **Admin-Dashboard** angezeigt (für ELSTER-Meldung und DSFinV-K-Export).
- **Disaster Recovery:** Geht die Datenbank ohne Backup verloren, muss die alte Seriennummer beim Finanzamt abgemeldet und die neue Instanz neu angemeldet werden — Datenbank-Backups sichern die Seriennummern-Kontinuität.

| Verwendungsort                               | Feld / Kontext                                         |
| -------------------------------------------- | ------------------------------------------------------ |
| ELSTER-Meldung                               | „Seriennummer des elektronischen Aufzeichnungssystems" |
| DSFinV-K Export (`cashregister.csv`)         | Feld `KASSE_SERIENNR`                                  |
| Kassenbon (Pflichtfeld nach § 6 KassenSichV) | Angedruckter String, z. B. `Kassen-ID: 7f3a9d12-...`   |
| TSE-Kommunikation                            | `serial_number` des fiskaly-Clients (Client-Registrierung) |

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
| Kryptografische Verkettung | ✅ Erfüllt          | TSE-Signatur (fiskaly Cloud-TSE) für alle Geschäftsvorfälle; Signaturdaten im Event persistiert, Ausfälle werden nachsigniert |
| 10-Jahres-Aufbewahrung     | ⚠️ Nicht adressiert | Keine Archivierungsstrategie implementiert            |

### 4.2 Anforderungen gemäß §§ 146, 147 AO und GoBD

- **Aufbewahrungspflicht:** Alle steuerlich relevanten Daten **10 Jahre** — jederzeit verfügbar, unverzüglich lesbar, vollständig, unveränderbar.
- **Elektronisches Radierverbot:** Kein `UPDATE` oder `DELETE` nach der Erfassung.
- **Stornierungen:** Immer als neue Buchungssätze (neuer Zeitstempel, neue TSE-Signatur), die den alten Wert ausgleichen.
- **Verfahrensdokumentation:** Wie das System Daten erzeugt, verarbeitet und archiviert, muss dokumentiert sein. [4]

### 4.3 Handlungsbedarf

- Archivierungsstrategie für die 10-Jahres-Aufbewahrung definieren
- Verfahrensdokumentation erstellen
- Soft-Delete-Praxis bei Stammdaten prüfen

---

## 5. Belegausgabepflicht

### 5.1 Gesetzliche Grundlage

Gemäß § 146a Abs. 2 AO und § 6 KassenSichV muss für **jeden Kassiervorgang** ein Beleg erzeugt und dem Kunden zur Verfügung gestellt werden — gedruckt oder (mit Zustimmung) elektronisch. [1, 2]

> **Befreiung für Vereinsfeste (§ 146a Abs. 2 Satz 2 AO):** Beim „Verkauf von Waren an eine **Vielzahl nicht bekannter Personen**" kann das Finanzamt von der Pflicht zur **Aushändigung** befreien — der typische Festzelt-Tatbestand. Zu beachten:
>
> - Die Befreiung betrifft **nur die Aushändigung** — TSE-Absicherung, Erfassung und Belegerzeugung bleiben Pflicht; der Beleg muss jederzeit **erstellbar** sein.
> - Sie wird **nicht automatisch** gewährt: schriftlicher Antrag beim Finanzamt, widerruflich.
> - Verlangt ein Gast einen Beleg, ist er auszuhändigen.
>
> jotti erstellt den Kassenbeleg deshalb **auf Anforderung** (→ [anforderungen.md F-03](anforderungen.md)) statt nach jeder Zahlung automatisch zu drucken.

> **Arbeitsbon ≠ Kassenbeleg:** Der automatische **Arbeitsbon** (Küche/Theke, ohne Preise) ist rein operativ — kein Beleg i. S. v. § 146a AO, keine TSE-Transaktion. Nur der **Kassenbeleg** (auf Anforderung pro Kassiervorgang) ist der fiskalische Beleg mit den Pflichtangaben aus §5.2. Details: [handbuch.md §4.6](handbuch.md#46-bondruck-arbeitsbon-und-kassenbeleg-k-12).

### 5.2 Pflichtangaben auf dem Beleg

**Standard-Kassendaten:**

- Vollständiger Name und Anschrift des leistenden Unternehmens (Verein)
- Datum der Belegausgabe
- Menge und Art der gelieferten Gegenstände / Umfang der Dienstleistung
- Entgelt und darauf entfallender Steuerbetrag, oder Hinweis auf Steuerbefreiung
- Transaktionsnummer (Bonnummer)
- Pro Position: Steuerkennzeichen (z. B. `A` für 19 %, `B` für 7 %). Bei `kombi`-Positionen (70/30): zwei Teilzeilen oder gemeinsamer Positionstext mit Verweis auf die Steuermatrix.
- Im Belegfuß — Steuermatrix: Brutto, Netto und Steuerbetrag je Steuersatz; `kombi`-Anteile fließen anteilig in die 7-%- und 19-%-Zeilen ein.

**TSE-Pflichtdaten (§ 6 KassenSichV):**

- Zeitpunkt des Vorgangsbeginns und der Vorgangsbeendigung (TSE-`logTime` aus Start/Finish)
- Seriennummer des Aufzeichnungssystems (Kassen-ID) und der TSE
- TSE-Transaktionsnummer, Signaturzähler, kryptografischer Prüfwert (Signatur)

### 5.3 Besondere Anforderung beim Festzelt-Muster (Durchbedienen)

Wurden Bestellungen mit `Bestellung-V1` abgesichert und erst später bezahlt (→ §3.6), gilt laut BMF-FAQ: „Zusätzlich ist auf den Bon der **Startzeitpunkt der ersten Bestellung in Klarschrift aufzudrucken**." [11, 13] Der Zahlungsbeleg trägt also **zwei Zeitstempel**: den TSE-`logTime` der aktuellen `Kassenbeleg-V1`-Transaktion und den `logTime` der **ersten `Bestellung-V1`** der Tisch-Session in Klarschrift.

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

Die TSE-Daten können platzsparend als **QR-Code** auf den Beleg — das Format muss der DSFinV-K (Anhang I) entsprechen.

### 5.5 Architektonische Anforderungen an jotti

1. **Beleg-Generator:** Bereits implementiert — `POST /service/beleg-drucken` akzeptiert `verkaufId` (Direktverkauf) oder `tischId` + `zahlungId` (Tisch-Zahlung) und erzeugt einen ESC/POS-Druckauftrag an den Kassenbeleg-Drucker.
2. **TSE-Daten auf dem Beleg** (nach TSE-Integration, F-02).
3. **Erste-Bestellung-Zeitstempel:** `logTime` der ersten `Bestellung-V1` einer Tisch-Session persistieren und beim Beleg-Druck abrufen (nur Tisch-Belege — Direktverkäufe haben keine vorgelagerte Bestellung).
4. **QR-Code-Generierung** im DSFinV-K-Format.
5. **Beleg-Archivierung** für DSFinV-K-Export und GoBD.

### 5.6 Umsetzung der Belegausgabe im BYOD-Setup

Die Servicekräfte nutzen private Smartphones ohne mobile Bondrucker — daraus folgen zwei Ausgabewege:

**Variante A — Zentraler Bondrucker an der Theke (Primärlösung, Phase 1):** Die Servicekraft kassiert auf dem Smartphone; erst nach erfolgreichem TSE-Abschluss (`FinishTransaction`) sendet das Backend den Druckbefehl an den stationären Bondrucker — die Reihenfolge TSE-Abschluss **vor** Druck ist rechtlich zwingend, da erst dann die Prüfwerte feststehen. Die Servicekraft bietet dem Gast den Bon an. Lehnt der Gast ab, ist die Pflicht dennoch erfüllt — § 146a Abs. 2 AO verlangt das „Ausstellen und Zur-Verfügung-Stellen", nicht die Annahme.

**Variante B — Digitaler eBeleg via QR-Code (optional, Phase 2):** Die bloße Bildschirmanzeige auf dem Kellner-Smartphone ist **unzulässig** — der Gast muss den Beleg elektronisch entgegennehmen können. jotti zeigt nach dem Kassiervorgang einen QR-Code (Download-Link zu PDF/JPG/PNG auf dem Backend); der Gast scannt ihn mit dem eigenen Gerät. Die Zustimmung gilt mit dem Scan als konkludent erteilt. Generierte eBelege müssen für DSFinV-K und GoBD archiviert werden. [1, 2]

---

## 6. DSFinV-K Export-Schnittstelle

### 6.1 Übersicht

Bei Kassen-Nachschau oder Betriebsprüfung verlangt die Finanzverwaltung einen genormten Export nach **DSFinV-K v2.4** (Stand Januar 2024), lesbar durch die Prüfsoftware IDEA. [5]

### 6.2 Dateiformat und Grundregeln

- **Gesamtstruktur:** ZIP-Archiv mit CSV-Dateien und einer `index.xml` (Metadaten für das Prüftool)
- **CSV-Regeln:** Header-Zeile zwingend; Semikolon-Trennung; CRLF; Punkt als Dezimaltrennzeichen, keine Tausendertrennzeichen, mindestens eine Stelle vor dem Punkt, keine führenden Nullen; Spaltenreihenfolge exakt nach Spezifikation
- **Dateinamen:** **englisch und kleingeschrieben** (`transactions.csv`, `lines.csv`, `cashregister.csv`, `tse.csv`, …) — nicht abänderbar. Die deutschen Begriffe der Spezifikation („Bonkopf", „Bonpos") sind **logische** Bezeichnungen, keine Dateinamen: Wer `Bonkopf.csv` exportiert, erzeugt eine **nicht** konforme Datei.
- **Custom-Felder:** Zusätzliche Spalten am Ende erlaubt, müssen in `index.xml` definiert sein

### 6.3 Modul-Struktur und offizielle Dateinamen

Drei Module; jeweils offizieller Dateiname (englisch) und logische DSFinV-K-Bezeichnung (deutsch):

#### A. Stammdatenmodul

| Dateiname (offiziell)  | Logische Bezeichnung | Inhalt                                                                                                                       |
| ---------------------- | -------------------- | ----------------------------------------------------------------------------------------------------------------------------- |
| `cashpointclosing.csv` | Stamm_Abschluss      | Metadaten zum Z-Bon: Unternehmensname, Steuernummer, Start-/End-Zeitpunkt                                                    |
| `location.csv`         | Stamm_Orte           | Standortdaten der Betriebsstätte                                                                                             |
| `cashregister.csv`     | Stamm_Kassen         | Kassendaten: Hersteller, Seriennummer, Software-Typ und -Version                                                             |
| `tse.csv`              | Stamm_TSE            | TSE-Daten: Zertifikats-ID, Signaturalgorithmus, TSE-Seriennummer (64-stelliger Hexadezimalstring), Public Key (Base64)       |
| `vat.csv`              | Stamm_USt            | Stammdaten der verwendeten Steuersätze                                                                                       |

#### B. Einzelaufzeichnungsmodul (Bonmodul)

| Dateiname (offiziell)   | Logische Bezeichnung | Inhalt                                                                                                                           |
| ----------------------- | -------------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| `transactions.csv`      | Bonkopf              | Ein Datensatz pro Kassenbon: `BON_ID`, `BON_NR`, `BON_TYP`, `BON_START`, `BON_ENDE` (ISO 8601), Gesamtbruttoumsatz, `BON_STORNO` |
| `transactions_vat.csv`  | Bonkopf_USt          | USt-Aufschlüsselung pro Bon nach Steuerschlüsseln (Brutto, Netto, USt)                                                           |
| `allocation_groups.csv` | Bonkopf_AbrKreis     | Zuordnung eines Bons zu einem `ABRECHNUNGSKREIS` (Tisch-Verknüpfung im Festzelt)                                                 |
| `datapayment.csv`       | Bonkopf_Zahlarten    | Zahlungsarten pro Bon (Bar, EC-Karte, Kreditkarte)                                                                               |
| `references.csv`        | Bon_Referenzen       | Referenzen auf andere Bons, u. a. `REF_BON_ID` bei Stornos                                                                       |
| `lines.csv`             | Bonpos               | Einzelne Artikel: `POS_ZEILE`, `ART_NR`, `MENGE`, `EINHEIT`, `STK_BR` (Stückpreis brutto)                                        |
| `lines_vat.csv`         | Bonpos_USt           | USt-Aufschlüsselung pro Artikelzeile                                                                                             |
| `transactions_tse.csv`  | TSE_Transaktionen    | **Kritisch:** TSE-Transaktionsnummer (`TSE_TANR`), Signaturzähler (`TSE_TA_SIGZ`), Krypto-Signatur (`TSE_TA_SIG`)                |

#### C. Kassenabschlussmodul (Z-Bon)

| Dateiname (offiziell)   | Logische Bezeichnung | Inhalt                                                         |
| ----------------------- | -------------------- | -------------------------------------------------------------- |
| `businesscases.csv`     | Z_GV_Typ             | Aggregierte Beträge pro Geschäftsvorfall-Typ nach Steuersätzen |
| `payment.csv`           | Z_Zahlart            | Aggregierte Summen der Zahlungsarten (Bar vs. unbar)           |
| `cash_per_currency.csv` | Z_Waehrungen         | Bargeldbestand je Währung zum Abschluss                        |

### 6.4 Schlüsselfelder (Relationale Verknüpfung)

Fast jede CSV-Datei führt in den ersten Spalten: **`Z_KASSE_ID`** (Kassen-ID), **`Z_ERSTELLUNG`** (Zeitstempel des Kassenabschlusses), **`Z_NR`** (fortlaufende Z-Bon-Nummer), **`BON_ID`** (Vorgangs-ID).

### 6.5 Abrechnungskreis — Tisch-Verknüpfung für Festzelt

Das Feld `ABRECHNUNGSKREIS` (in `allocation_groups.csv`, je `BON_ID`) verknüpft Bestellungen und Zahlungen eines Tisches zu einer logischen Einheit — Beispieltabelle und Ablauf: → [§3.6](#36-das-festzelt-muster-atomare-tse-transaktionen).

- **Vergabe (✅ umgesetzt):** pro Tisch und Kassensitzung, Wert = Tischname (z. B. `Tisch 42`; das Format erlaubt beliebige Strings bis 40 Zeichen — `Tisch {Name}` ist eine jotti-interne Konvention); intern Subject `kassensitzung-{nr}/tisch-{id}`. Jeder Tisch erhält seinen **eigenen** Abrechnungskreis — ein Gesamt-Schlüssel für alle Tische verstieße gegen die GoBD-Nachvollziehbarkeit. Der Tagesabschluss schließt die Kassensitzung und damit alle Tisch-Sessions.
- **Geplant — manuelle Tischfreigabe** (→ [anforderungen.md K-23](anforderungen.md)): neue Gästegruppe am selben Tisch erhält ein Suffix (`Tisch 42-B`) — korrektere Abbildung im Festzelt, weniger Rückfragen bei Prüfungen. Bis dahin teilen sich Gästegruppen am selben Tisch einen Abrechnungskreis — zulässig, solange alle Bons korrekt verknüpft sind.
- **Direktverkauf:** sofort geschlossene Transaktion ohne Tisch — im Export **ohne** `ABRECHNUNGSKREIS` (Feld ist optional) oder per Konvention `Theke`; Entscheidung bei Implementierung des Exporters. Keine `Bestellung-V1` nötig — direkt als `Kassenbeleg-V1` abgesichert.

### 6.6 Storno-Handling in DSFinV-K

Stornierungen erzeugen **immer neue Datensätze** (GoBD-Radierverbot), nie Änderungen am Ursprungsbon:

- **Positions-Storno (vor Zahlung):** neuer Bon mit negativer Menge, `BON_STORNO = 1` in `transactions.csv`, `REF_BON_ID` in `references.csv` auf den Ursprungsbon, eigene TSE-Signatur, gleicher `ABRECHNUNGSKREIS`.
- **Bon-Storno (nach Zahlung):** neuer Bon mit negativem Gesamtbetrag, `BON_STORNO = 1`, `REF_BON_ID` auf den Original-Zahlungsbeleg, eigene `Kassenbeleg-V1`-Transaktion.
- **Direktverkauf:** `direktverkauf-getaetigt:v1` (positiver Geschäftsvorfall) und `direktverkauf-storniert:v1` (negativer Geschäftsvorfall) sind je eigene Belegvorgänge; Zielbild ist ein 1:1-Mapping je Event mit `REF_BON_ID`-Referenz.

### 6.7 Architektonische Anforderungen an jotti

1. **CSV-Generator** aus Event-Store- und Stammdaten (offizielle englische Dateinamen)
2. **`index.xml`-Generator**
3. **Z-Bon-Logik** (Tagessummen aggregieren)
4. **Abrechnungskreis-Verwaltung** (Tisch-Session-ID in allen zugehörigen Bons)
5. **Admin-Endpunkt** zum Auslösen des Exports
6. **ZIP-Generierung**
7. **Steuersatz-Verwaltung:** USt-Sätze als Stammdaten — **19 %** (Regelsteuersatz, z. B. Getränke), **7 %** (ermäßigt, z. B. Speisen), **0 % / steuerbefreit** (Zweckbetrieb nach § 67a AO), **Kombi 70/30** (Kombinationsangebote nach Abschn. 10.1 Abs. 12 UStAE). Produkte erhalten einen konfigurierbaren Steuersatz-Schlüssel. Bei `kombi`-Positionen entfaltet der Export den Pauschalpreis in zwei Steueranteile (70 % → 7 %, 30 % → 19 %): zwei VAT-Einträge in `lines_vat.csv`, beide Anteile in `transactions_vat.csv`. Steuerregeln: [steuerrecht.md](steuerrecht.md).

---

## 7. Elektronische Meldepflicht (ERiC / ELSTER)

### 7.1 Gesetzliche Grundlage und Fristen

Nach § 146a Abs. 4 AO müssen elektronische Aufzeichnungssysteme beim zuständigen Finanzamt gemeldet werden; das Mitteilungsverfahren ist seit dem 1. Januar 2025 aktiv. [1, 12]

| Datum                                    | Ereignis                                                                      |
| ---------------------------------------- | ----------------------------------------------------------------------------- |
| 1. Januar 2025                           | Meldeportal öffnet; Mitteilungsverfahren ist aktiv                            |
| 31. Juli 2025                            | Abgabefrist für alle Systeme, die **vor dem 1. Juli 2025** angeschafft wurden |
| Innerhalb 1 Monat nach Anschaffung       | Abgabefrist für Systeme, die **ab dem 1. Juli 2025** neu angeschafft werden   |
| Innerhalb 1 Monat nach Außerbetriebnahme | Abgabefrist bei Stilllegung eines Systems                                     |

### 7.2 Übermittlungswege und gewählter Ansatz

Drei Wege: Direkteingabe im ELSTER-Portal (manuell), XML-Dateiupload (semi-automatisch), programmatische Übermittlung über **ERiC** (ELSTER Rich Client — offizielle Komponente der Finanzverwaltung mit lokaler Vorab-Validierung und Bestätigungsprotokoll).

**Gewählter Ansatz für jotti** (Self-hosted, ehrenamtliche Betreiber):

- **Phase 1 (manuell):** Der Admin-Bereich stellt alle meldepflichtigen Daten strukturiert bereit; der Vorstand meldet über das ELSTER-Webportal.
- **Phase 2 (optional):** ERiC (native C-Library, aufwendig, kein Vendor-Lock-in) **oder** fiskaly-Submission-API (einfacher, aber Cloud-Abhängigkeit und keine staatliche Vorab-Validierung). Entscheidung bei Implementierungsbeginn nach Aufwands-/Kosten-Abwägung. [6]

### 7.3 Meldepflichtige Daten (Payload)

- Name und Steuernummer des Steuerpflichtigen
- Art des Kassensystems (Softwaretyp, Versionsnummer)
- Seriennummer des Kassensystems
- Zertifizierungs-ID der TSE (Format `BSI-K-TR-nnnn-yyyy`)
- Seriennummer der TSE (64-stelliger Hexadezimalstring, 0–9/A–F — nicht zu verwechseln mit dem Base64-Public-Key in `tse.csv`)
- Anschaffungs- bzw. Inbetriebnahmedatum
- Betriebsstättenadresse

### 7.4 Architektonische Anforderungen an jotti

1. **Konfiguration:** Vereinsdaten, Betriebsstätte, Steuernummer als Stammdaten
2. **Admin-Datenanzeige (Phase 1):** alle meldepflichtigen Felder strukturiert anzeigen — inkl. der generierten Kassen-Seriennummer (→ §3.7); kein API-Aufruf
3. **Meldestatus (Phase 1):** manuell setzbarer Status („ausstehend / gemeldet am TT.MM.JJJJ"), persistiert in den Stammdaten
4. **Automatisierte Übermittlung (Phase 2, optional):** ERiC oder fiskaly-Submission-API

### 7.5 BYOD-Smartphones: keine Meldepflicht

Die Smartphones der Servicekräfte müssen **nicht** gemeldet werden. Der AEAO zu § 146a AO stellt klar: Bei Systemen ohne eigene Kassenfunktion (Eingabegeräte), die mit einem Aufzeichnungssystem mit Kassenfunktion verbunden sind, ist nur das **Hauptsystem** mitteilungspflichtig — Vereine melden ausschließlich ihre jotti-Instanz, nicht die Handys der Helfer. [15] Die vollständige Pflichtenverteilung Entwickler/Betreiber: → Abschnitt 8.

---

## 8. Betreiberpflichten

Die Vereine tragen als Betreiber die volle operative und rechtliche Verantwortung für ihre jotti-Instanz. Laienverständliche Anleitung: [betrieb/leitfaden-betreiber.md](betrieb/leitfaden-betreiber.md).

**Pflichtenverteilung Entwickler / Betreiber:**

| Pflicht                                      | Entwickler (jotti) | Verein (Betreiber)          |
| -------------------------------------------- | ------------------ | --------------------------- |
| TSE-Schnittstelle im Code implementieren     | ✅ Pflicht         | —                           |
| DSFinV-K-Export im Code implementieren       | ✅ Pflicht         | —                           |
| Muster-Verfahrensdokumentation bereitstellen | Empfohlen          | —                           |
| Cloud-TSE-Vertrag abschließen                | —                  | ✅ Pflicht                  |
| TSE-API-Keys konfigurieren (Admin-UI)        | —                  | ✅ Pflicht                  |
| ELSTER-Meldung (§ 146a Abs. 4 AO)            | ❌ Keine Pflicht   | ✅ Pflicht (Frist: 1 Monat) |
| BYOD-Smartphones melden                      | —                  | ❌ Nicht erforderlich       |
| 10-Jahres-Archivierung (GoBD)                | —                  | ✅ Pflicht                  |
| Server-Betrieb und Datensicherung            | —                  | ✅ Pflicht                  |

**Vor dem ersten Einsatz:**

1. **Cloud-TSE-Vertrag (BYOT):** Vertrag mit einem Cloud-TSE-Anbieter (z. B. fiskaly oder D-Trust) abschließen; API-Schlüssel in den TSE-Einstellungen des Admin-Bereichs hinterlegen (gespeichert in der Datenbank).
2. **ELSTER-Meldung:** Innerhalb von **einem Monat** nach Inbetriebnahme die Instanz über [ELSTER](https://www.elster.de) anmelden. Benötigt: Kassen-Seriennummer (Admin-Dashboard), Softwarename „jotti", Inbetriebnahmedatum (→ §7.3).
3. **Seriennummer sichern:** Die Kassen-UUID ist die rechtliche Identität der Kasse (→ §3.7) — das Datenbank-Backup muss sie enthalten. Bei Verlust: alte Nummer abmelden, neue Instanz anmelden.

**Laufend:**

- **Kassensturz und Z-Bon:** Beim Kassensturz den gezählten Ist-Bestand in einem Zählprotokoll festhalten; nach jedem Veranstaltungstag einen Z-Bon erstellen (ein X-Bon/Zwischenbericht ersetzt ihn rechtlich nicht). Differenzen ehrlich buchen — eine Kasse, die „immer auf den Cent genau stimmt", gilt der Finanzverwaltung als manipulationsverdächtig (IDEA-Prüfung). Z-Bons und Zählprotokolle fallen unter die 10-Jahres-Aufbewahrung.
- **10-Jahres-Aufbewahrung:** Kassenjournal und DSFinV-K-Exporte GoBD-konform und jederzeit lesbar aufbewahren (§§ 146, 147 AO).
- **Tägliche Backups** — für die Aufbewahrung und die Seriennummern-Kontinuität.
- **Server-Betrieb:** Verfügbarkeit, Zugriffsschutz, Datensicherung.
- **Außerbetriebnahme** innerhalb eines Monats bei ELSTER melden.

---

## 9. Architekturprinzipien

→ **[handbuch.md §3.13 — TSE-Architektur](handbuch.md#313-tse-architektur)**: vollständiges Transaktions-Mapping (Atomares Modell), TSE-Datenpersistenz im Event Store, DSFinV-K-Exporter-Übersicht. Interface-Definition: `backend/domain/tse/client.go`. Anbieter- und Meldeweg-Entscheidungen: §3.5 und §7 in diesem Dokument.

---

## 10. Quellenverzeichnis

Verweise im Text (z. B. [1]) beziehen sich auf die Nummern dieser Liste.

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
