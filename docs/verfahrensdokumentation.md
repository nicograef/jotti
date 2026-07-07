---
title: Muster-Verfahrensdokumentation
description: 'Muster-Verfahrensdokumentation für die Kassenführung mit jotti: Vorlage, die der Betreiber für seine GoBD-Pflicht an die eigene Instanz anpasst.'
---

> ⚠️ **Dies ist eine Vorlage.** jotti stellt diese Muster-Verfahrensdokumentation als Hersteller bereit ([compliance.md §2.7](compliance.md#27-sonderfall-source-available-self-hosted-pflichten-des-entwicklers)). Der Betreiber (Verein) passt sie an seine Instanz an, ergänzt die betrieblichen Angaben und legt sie bei einer Kassen-Nachschau oder Betriebsprüfung vor. Die Verfahrensdokumentation selbst ist eine Betreiberpflicht nach GoBD ([compliance.md §4.2](compliance.md#42-anforderungen-gemäß--146-147-ao-und-gobd)).
>
> **So benutzt ihr die Vorlage:** Alle vom Betreiber auszufüllenden Stellen sind mit Guillemets markiert, zum Beispiel «Vereinsname». Ersetzt jede solche Stelle durch eure konkreten Angaben und entfernt anschließend diesen Hinweiskasten. Die technischen Abschnitt 2 bis 8 beschreiben jotti und gelten für jede Instanz; prüft sie auf Aktualität gegen eure jotti-Version und passt nur die mit «…» markierten Stellen an.

## Was eine Verfahrensdokumentation leistet

Die GoBD (BMF-Schreiben vom 28.11.2019) verlangen vom Betreiber eines elektronischen Kassensystems eine Verfahrensdokumentation: eine nachvollziehbare Beschreibung, wie das System steuerlich relevante Daten erzeugt, verarbeitet, sichert und aufbewahrt. Sie versetzt einen sachverständigen Dritten (Betriebsprüfer) in die Lage, den Kassenprozess in angemessener Zeit zu verstehen und zu prüfen. Dieses Dokument erfüllt diese Pflicht für die Kassenführung mit jotti.

Rechtliche Grundlagen im Detail: [compliance.md](compliance.md). Technische Architektur: [handbuch.md](handbuch.md). Betreiber-Anleitung in laienverständlicher Form: [Leitfaden](leitfaden/was-ist-jotti.md).

---

## 1. Stammdaten der Instanz (vom Betreiber auszufüllen)

| Angabe | Wert |
| --- | --- |
| Betreiber (Verein) | «Vereinsname e.V.» |
| Anschrift der Betriebsstätte | «Straße, PLZ, Ort» |
| Steuernummer | «Steuernummer» |
| USt-IdNr. (falls vorhanden) | «USt-IdNr. oder „nicht vorhanden"» |
| Kassen-Seriennummer (jotti-Kassen-UUID) | «UUID aus dem Admin-Bereich → Finanzamt» |
| Softwarename / Hersteller | jotti |
| Eingesetzte jotti-Version | «z. B. v0.2.0» |
| Inbetriebnahmedatum | «TT.MM.JJJJ» |
| ELSTER-Meldung erfolgt am | «TT.MM.JJJJ» |
| TSE-Anbieter | «z. B. fiskaly (SIGN DE), Cloud-TSE» |
| TSE-Zertifizierungs-ID | «Format BSI-K-TR-nnnn-yyyy» |
| TSE-Seriennummer | «64-stelliger Hexadezimalstring» |
| Betriebsumgebung | «z. B. Windows-Rechner im Vereinsheim (lokales WLAN) / eigener VPS mit Domain» |
| Verantwortlich für die Kasse | «Name, Rolle im Verein, z. B. Schatzmeister» |
| Stand dieser Dokumentation | «TT.MM.JJJJ» |

---

## 2. Systemüberblick und Architektur

jotti ist ein self-hosted Kassensystem (mobile Point of Sale) für Vereinsfeste. Das Backend ist in Go geschrieben, das Frontend in React, als Datenbank dient PostgreSQL, der Betrieb erfolgt über Docker Compose. Die Servicekräfte bedienen jotti im Browser auf ihren eigenen Smartphones (BYOD); ein gedrucktes Kassensystem oder eine installierte App gibt es nicht.

**Rolle der Smartphones:** Die Geräte der Servicekräfte sind reine Eingabegeräte mit sofortiger Weiterleitung an das Backend. Sie erfassen keine Zahlungen eigenständig und offline; jeder Vorgang ist ein synchroner Backend-Request, ohne Verbindung ist keine Erfassung möglich (kein Service Worker, kein Offline-Speicher). TSE-Absicherung, Protokollierung und DSFinV-K-Speicherung erfolgen ausschließlich zentral im Backend. Die Smartphones sind deshalb nicht meldepflichtig ([compliance.md §7.5](compliance.md#75-byod-smartphones-keine-meldepflicht)).

**Bounded Contexts:** Das System ist in drei fachliche Bereiche gegliedert.

| Bereich | Aufgabe | Persistenz |
| --- | --- | --- |
| Kasse | Alle finanziellen Geschäftsvorfälle: Bestellen, Ausgeben, Kassieren, Stornieren, Umbuchen, Kassenbewegungen, Kassensturz, Tagesabschluss | Event-Sourcing (Kassenjournal) |
| Stammdaten | Produkte, Tische, Benutzer, Betreiber-Stammdaten | CRUD mit Soft-Delete |
| Auth | Login, Logout, Passwort, Token | Infrastruktur |

**Betriebsumgebung dieser Instanz:** «Beschreibt hier, wo jotti läuft. Beispiel Standardweg: ein Windows-Rechner im Vereinsheim wird zum Kassenrechner, die Helfer bedienen jotti im selben WLAN. Beispiel Experten-Weg: jotti läuft auf einem VPS bei «Anbieter» unter der Domain «kasse-musterverein.de» mit HTTPS.»

Vollständige Architektur-Referenz: [handbuch.md §1 und §2](handbuch.md#1-überblick).

---

## 3. Datenmodell und Event-Sourcing

**Kassenjournal als unveränderliche Aufzeichnung:** Alle finanziellen Geschäftsvorfälle werden in der Tabelle `kassenjournal` festgehalten, einer chronologischen, vollständigen und append-only geführten Aufzeichnung im Sinne von § 146 AO. Ein Datenbank-Trigger verhindert nachträgliches Ändern und Löschen (`UPDATE` und `DELETE` sind technisch gesperrt). Das Kassenjournal ist die alleinige Quelle der Wahrheit; alle Auswertungen und Exporte werden aus ihm abgeleitet.

**Events statt Datensatz-Änderungen:** Jeder Vorgang wird als typisiertes, unveränderliches Ereignis (Event) gespeichert, zum Beispiel `bestellung-aufgenommen:v1`, `zahlung-kassiert:v1` oder `stornierung-erteilt:v1`. Ein bestehendes Ereignis wird nie verändert; Korrekturen entstehen ausschließlich durch neue Gegen-Ereignisse (siehe [Abschnitt 8](#8-nachvollziehbarkeit-von-änderungen)).

**Einfrieren der Bezugsdaten (Fat Events):** Eine Bestellung speichert Produktname, Variante, Kategorie, Steuersatz und Einzelpreis so, wie sie im Moment der Bestellung galten. Spätere Preis- oder Stammdaten-Änderungen wirken nur auf künftige Bestellungen und verändern historische Buchungen nicht.

**Strukturierung (Subjects):** Die Ereignisse sind nach Betriebstag (Kassensitzung) und Abrechnungseinheit geordnet: jeder Tisch innerhalb einer Kassensitzung sowie jeder Theken-Direktverkauf bildet einen eigenen, lückenlos versionierten Ereignis-Strom. Das ergibt pro Tisch und pro Betriebstag einen geschlossenen Abrechnungskreis.

**Geldbeträge:** Alle Beträge sind ganzzahlige Cent-Werte. Es werden keine Fließkommazahlen verwendet.

**Stammdaten:** Produkte, Tische und Benutzer werden klassisch verwaltet (CRUD). Gelöscht wird nie physisch, sondern per Soft-Delete (Status `deleted`); die referenzielle Integrität und die historische Nachvollziehbarkeit bleiben dadurch erhalten.

Datenmodell-Referenz: [handbuch.md §3](handbuch.md#3-kasse-core-domain); Schema: `database/migrations/01_initial.up.sql`.

---

## 4. TSE-Anbindung

jotti unterliegt nach § 146a AO der Pflicht, jeden Geschäftsvorfall durch eine zertifizierte Technische Sicherheitseinrichtung (TSE) abzusichern. Diese Instanz nutzt die Cloud-TSE von «TSE-Anbieter», angebunden über eine HTTPS-API.

**Absicherung jedes Vorgangs:** Jeder relevante Vorgang wird über das anbieter-agnostische `TSEClient`-Interface signiert. jotti folgt dem atomaren Festzelt-Muster: Jeder Vorgang ist eine eigene, sofort geschlossene TSE-Transaktion. Die Zuordnung der TSE-Vorgangsarten:

| jotti-Vorgang | TSE-Vorgangsart (processType) |
| --- | --- |
| Bestellung aufnehmen, geldneutrale Korrektur, Umbuchung | `Bestellung-V1` |
| Zahlung, Warenrücknahme (kassenwirksamer Storno), Geldtransit, Kassendifferenz, Direktverkauf (inkl. Storno) | `Kassenbeleg-V1` |
| Tagesabschluss (Z-Bon) | `SonstigerVorgang` |

**Signaturablauf und Persistenz:** Die Signierung ist vom Kassiervorgang entkoppelt, das Buchen wartet nie auf die TSE. Jeder signaturpflichtige Vorgang schreibt im selben Datenbank-Commit wie das Ereignis genau einen Signaturauftrag (transaktionale Outbox `tse_signaturauftraege`); ein Signatur-Worker signiert asynchron über das `TSEClient`-Interface und speichert die von der TSE gelieferten Signaturdaten (Transaktionsnummer, Signaturzähler, Signatur, Zeitstempel, Seriennummer) direkt am Auftrag. Beleg und DSFinV-K-Export lesen genau diese eine Quelle. Im Regelbetrieb liegt die Signatur binnen Sekunden vor (angestrebte Latenz: p95 unter fünf Sekunden).

**Ausfallpfad und mögliche Verzögerungen:** Weil das Buchen nie auf die TSE wartet, kann sich die Signatur verzögern, ohne den Vorgang aufzuhalten; Ursachen sind ein TSE-Ausfall oder eine Zeitüberschreitung, ein Signatur-Rückstand unter Last oder eine fehlende TSE-Konfiguration. Es gibt keine still unsignierten Geschäftsvorfälle: Jeder Vorgang ist ein offener oder ein endgültig markierter Auftrag, nie ein verlorener. Das Störungsprotokoll (`tse_stoerungen`) dokumentiert jede TSE-weite Störung als Zeitraum mit Grund und dient zusammen mit der Auftragstabelle als Ausfalldokumentation (AEAO zu § 146a, 1.14.1). Ein nach der Störung nachsignierter Kassenbeleg weist die nachträgliche Signierung mit dem Vermerk „Nachsigniert am …" aus, da die TSE-Zeitpunkte dann vom Belegdatum abweichen; ein Vorgang ohne konfigurierte TSE trägt den Vermerk „keine TSE konfiguriert"; im DSFinV-K-Export erhalten noch unsignierte Vorgänge aller Vorgangsarten eine `TSE_TA_FEHLER`-Zeile.

**TSE-Stammdaten dieser Instanz:** Anbieter «TSE-Anbieter», Zertifizierungs-ID «BSI-K-TR-nnnn-yyyy», TSE-Seriennummer «Hexadezimalstring». Die Einrichtung (fiskaly-Konto, geführter Assistent, Verwahrung von Admin-PUK und Admin-PIN, Wechsel von TEST zu LIVE) ist im [Leitfaden, Abschnitt „TSE einrichten"](leitfaden/tse-einrichten.md) dokumentiert.

Rechtliche und technische Details: [compliance.md §3](compliance.md#3-tse-integration-technische-sicherheitseinrichtung) und [handbuch.md §3.13](handbuch.md#313-tse-architektur).

---

## 5. DSFinV-K-Export

§ 4 KassenSichV verlangt eine genormte digitale Schnittstelle für den Datenexport an die Finanzverwaltung. jotti erzeugt den Export nach DSFinV-K (Digitale Schnittstelle der Finanzverwaltung für Kassensysteme).

**Form:** ein ZIP-Archiv pro Kassensitzung, bestehend aus CSV-Dateien mit fest vorgeschriebenen englischen Dateinamen, einer `index.xml` und der zugehörigen `gdpdu-01-09-2004.dtd`. Die Daten werden aus dem Kassenjournal und den Stammdaten erzeugt.

**Inhalt:** Stammdatenmodul (Kassen-, Standort-, Steuersatz- und TSE-Daten), Einzelaufzeichnungsmodul (jeder einzelne Bon mit Positionen, Zahlarten, Referenzen und TSE-Signaturen) und Kassenabschlussmodul (aggregierter Z-Bon je Betriebstag). Bestellungen und Zahlungen eines Tisches sind über einen gemeinsamen `ABRECHNUNGSKREIS` verknüpft.

**Auslösung:** Der Admin erzeugt den Export im Admin-Bereich unter „Auswertungen" für die gewählte (abgeschlossene) Kassensitzung. Er ist mit jeder Tabellenkalkulation und mit der Prüfsoftware IDEA lesbar.

Format- und Felddetails: [compliance.md §6](compliance.md#6-dsfinv-k-export-schnittstelle).

---

## 6. Rollen- und Zugriffskonzept

**Authentifizierung:** Jeder Benutzer meldet sich mit Benutzername und Passwort an und erhält ein zeitlich begrenztes Token (JWT, 12 Stunden gültig). Passwörter werden ausschließlich als Argon2id-Hash gespeichert, niemals im Klartext. Die Rechteprüfung erfolgt serverseitig: Das Token weist den Benutzer aus, Rolle und Status werden bei jedem Request gegen die Benutzerdatenbank geprüft. Rollenänderungen und Deaktivierungen wirken damit sofort, nicht erst beim Ablauf des Tokens.

**Login-Fehlermeldungen (bewusste Abwägung):** Der Login unterscheidet in seinen Fehlermeldungen die Fälle „kein Passwort gesetzt" und „Benutzer inaktiv" vom allgemeinen „Benutzername oder Passwort falsch". Ein Angreifer kann daraus im Einzelfall ablesen, dass ein Benutzerkonto existiert (Enumerationsrisiko). Diese Abwägung ist bewusst getroffen: Die Zielgruppe sind nicht-technische, ehrenamtliche Helfer, die mit einer klaren Fehlermeldung ihr Anmeldeproblem selbst lösen können; eine serverseitige Anmeldedrosselung begrenzt automatisiertes Durchprobieren.

**Drei Rollen mit abgestuften Rechten:**

| Rolle | Berechtigung |
| --- | --- |
| `admin` | Voller Zugriff: Stammdaten, Kasse, Kassensitzung, Auswertungen, Export, Finanzamt-Daten |
| `serviceleitung` | Kasse einschließlich Stornierung |
| `service` | Kasse ohne Stornierung |

Stornierungen sind ausschließlich `serviceleitung` und `admin` vorbehalten; die Kassensitzung (Eröffnen, Kassensturz, Tagesabschluss) und alle Stammdaten- und Auswertungsfunktionen sind dem `admin` vorbehalten. Die vollständige Berechtigungsmatrix steht in [handbuch.md §5.1](handbuch.md#51-rollen-und-berechtigungsmatrix).

**Rollenvergabe in diesem Verein (auszufüllen):**

| Person | jotti-Benutzer | Rolle |
| --- | --- | --- |
| «Name» | «Benutzername» | «admin / serviceleitung / service» |
| «Name» | «Benutzername» | «…» |

«Beschreibt hier kurz, wer die Admin-Rolle hält (üblicherweise Vorstand oder Schatzmeister) und nach welchem Verfahren neue Servicekräfte angelegt und nach dem Fest wieder deaktiviert werden.»

**Bedienerzuordnung:** Jede Kassenaktion hält im Kassenjournal und im DSFinV-K-Export die stabile Benutzer-ID (`BEDIENER_ID`) und den zum Zeitpunkt der Aktion eingefrorenen Benutzernamen (`BEDIENER_NAME`) fest, nicht den bürgerlichen Klarnamen. Die Benutzerverwaltung (`users.name`) hält die Zuordnung Benutzername → Person und dient damit als Bedienerliste für die Betriebsprüfung; die obige Tabelle bildet dieselbe Zuordnung dokumentiert ab. Benutzernamen sind dauerhaft eindeutig und werden nie neu vergeben, sodass ein eingefrorener Name über die gesamte Aufbewahrungsfrist genau einer Person zugeordnet bleibt. Für gute Lesbarkeit empfiehlt sich eine Konvention wie Vorname plus Initial des Nachnamens (z. B. `annak`); jotti erzwingt sie nicht.

---

## 7. Archivierung und Aufbewahrung

Alle steuerlich relevanten Daten sind 10 Jahre vollständig, jederzeit verfügbar, lesbar und unveränderbar aufzubewahren (§§ 146, 147 AO). jotti stellt die Daten in offenen, ohne Spezialsoftware lesbaren Formaten bereit; die sichere Aufbewahrung selbst ist Betreiberpflicht.

| Artefakt | Quelle | Aufbewahrung |
| --- | --- | --- |
| DSFinV-K-Export je Kassensitzung | Admin-Bereich → Auswertungen | 10 Jahre, an mindestens zwei getrennten Orten |
| Datenbank-Backup (rohes Kassenjournal samt TSE-Signaturen und Stammdaten) | automatisches Backup; `make prod-backup` auf dem Server | 10 Jahre |
| Z-Bons (Tagesabschlüsse) | im Kassenjournal und DSFinV-K-Export enthalten | 10 Jahre |
| Zählprotokolle (Kassensturz) | manuell beim Kassensturz | 10 Jahre |
| Kassen-Seriennummer | Admin-Bereich; im DB-Backup enthalten | dauerhaft |

**Aufbewahrung in diesem Verein (auszufüllen):** «Beschreibt, wo und wie ihr archiviert. Beispiel: Nach jedem Veranstaltungstag exportieren wir die DSFinV-K-ZIP und legen sie auf einem USB-Stick im Vereinssafe sowie zusätzlich in «Cloud-Speicher» ab. Tägliche Datenbank-Backups laufen «automatisch über … / manuell durch …». Verantwortlich: «Name».»

Strategie im Detail: [compliance.md §4.4](compliance.md#44-aufbewahrungsstrategie-f-10). Laienverständliche Anleitung: [Leitfaden, Datenaufbewahrung](leitfaden/datenaufbewahrung.md).

---

## 8. Nachvollziehbarkeit von Änderungen

Die Unveränderbarkeit und Nachvollziehbarkeit nach GoBD wird auf drei Ebenen sichergestellt.

**Geschäftsvorfälle (Radierverbot):** Das Kassenjournal ist append-only; einmal erfasste Ereignisse werden nie geändert oder gelöscht. Jede Korrektur ist ein neuer Geschäftsvorfall mit eigenem Zeitstempel und eigener TSE-Signatur. Eine Stornierung gleicht den ursprünglichen Wert durch eine Gegenbuchung aus und verweist auf den Ursprungsvorgang, statt ihn zu überschreiben. Die fortlaufende Versionierung je Abrechnungskreis macht Lücken oder Dopplungen erkennbar.

**Stammdaten:** Produkte, Tische und Benutzer werden nur per Soft-Delete entfernt (Status `deleted`), nie physisch gelöscht. Verkaufspreise und Steuersätze sind zum Zeitpunkt jeder Buchung in den Ereignissen eingefroren, sodass spätere Stammdaten-Änderungen historische Buchungen nicht verändern. Vor einer Stammdaten-Änderung mit fiskaler Wirkung (Preise, Steuersätze) ist der Tagesabschluss durchzuführen.

**Software- und Konfigurationsstände:** Die eingesetzte jotti-Version ist im Admin-Bereich (Fußzeile der Seitenleiste) ersichtlich und im DSFinV-K-Export hinterlegt. Updates verändern die Datenbank nur vorwärts (kein Downgrade); vor jedem Update zieht jotti automatisch ein Backup. Wesentliche betriebliche Änderungen (Versions-Updates, Wechsel des TSE-Anbieters, Änderung der Betriebsumgebung) trägt der Betreiber in die Änderungshistorie am Ende dieses Dokuments ein.

---

## 9. Betrieb und Verantwortlichkeiten (vom Betreiber auszufüllen)

**Ablauf eines Veranstaltungstags:** «Beschreibt euren typischen Ablauf. Beispiel: Vor Beginn eröffnet «Name (admin)» die Kassensitzung und setzt den Anfangsbestand (Wechselgeld). Während des Fests nehmen die Servicekräfte Bestellungen auf und kassieren bar. Nach Veranstaltungsende zählt «Name» die Kasse (Kassensturz), bucht eine etwaige Differenz und erstellt den Tagesabschluss (Z-Bon). Anschließend wird der DSFinV-K-Export erzeugt und archiviert.»

**Internes Kontrollsystem:** Beim Kassensturz wird der gezählte Ist-Bestand in einem Zählprotokoll festgehalten und mit dem von jotti errechneten Soll-Bestand abgeglichen. Differenzen werden ehrlich gebucht und nicht geglättet. Verantwortlich für den Kassensturz: «Name».

**Verantwortlichkeiten:**

| Aufgabe | Verantwortliche Person |
| --- | --- |
| Kassensitzung eröffnen und abschließen | «Name» |
| Kassensturz und Differenzbuchung | «Name» |
| DSFinV-K-Export und Archivierung | «Name» |
| Datenbank-Backups | «Name» |
| TSE-Verwaltung (PUK/PIN-Verwahrung) | «Name» |
| ELSTER-Meldung und Stammdatenpflege | «Name» |

Betreiberpflichten vollständig: [compliance.md §8](compliance.md#8-betreiberpflichten); Praxis-Checkliste: [Leitfaden, Checkliste](leitfaden/checkliste.md).

---

## 10. Änderungshistorie dieser Dokumentation (vom Betreiber zu führen)

| Version | Datum | Änderung | Bearbeiter |
| --- | --- | --- | --- |
| 1.0 | «TT.MM.JJJJ» | Erstfassung auf Basis der jotti-Muster-Verfahrensdokumentation | «Name» |
| «…» | «…» | «z. B. jotti-Update auf v…, Wechsel des TSE-Anbieters» | «…» |

---

> Diese Vorlage basiert auf den jotti-Compliance-Grundlagen ([compliance.md](compliance.md)) und der Architektur-Referenz ([handbuch.md](handbuch.md)). Bei einem jotti-Update prüft der Betreiber, ob die technischen Abschnitte 2 bis 8 noch zur eingesetzten Version passen, und vermerkt Änderungen in Abschnitt 10.
