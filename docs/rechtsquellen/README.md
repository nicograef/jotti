# Rechtsquellen für jotti

Dieser Ordner sammelt die autoritativen Originaltexte der zentralen Normen und Spezifikationen, auf die sich [compliance.md](../compliance.md) und [steuerrecht.md](../steuerrecht.md) berufen; der GoBD-Basistext und die nur als HTML verfügbaren Quellen (BMF-FAQ, ELSTER) liegen nicht hier. Er dient zwei Zwecken: der Überprüfung und Betriebsprüfung von jottis Fiskal-Implementierung gegen die jeweils maßgebliche Quelle, und als autoritative lokale Nachschlagequelle für Agenten bei Compliance- und Steuer-Implementierungen (siehe [Nutzung durch Agenten](#nutzung-durch-agenten)).

Alle Dateien stammen aus offiziellen Quellen (gesetze-im-internet.de des BMJ, Bundesfinanzministerium, Bundeszentralamt für Steuern, BSI, EUR-Lex); die OpenAPI-Spezifikation stammt von fiskaly selbst. Abrufdatum: 24. Juni 2026. Prüfsummen: [SHA256SUMS.txt](SHA256SUMS.txt).

## Nutzung durch Agenten

Bei fiskal- und steuerrechtlichen Fragen zuerst hier nachsehen, statt im Web zu suchen oder aus dem Gedächtnis zu antworten (AGENTS.md, Regel 13). Den Domänen- und Implementierungsbezug liefern [compliance.md](../compliance.md) und [steuerrecht.md](../steuerrecht.md); dieser Ordner liefert den dahinterstehenden Originaltext.

Lesen:

- `.dtd` und `.xml` direkt mit dem Datei-Lesewerkzeug ganz öffnen; sie sind klein genug. Die fiskaly-OpenAPI-`.json` (rund 250 KB in einer einzigen Zeile) nicht ganz lesen, sondern gezielt per Suche nach dem jeweiligen Pfad oder Schema.
- PDFs seitenweise lesen (Seitenbereich angeben). Die großen Gesetzes- und Spezifikations-PDFs (AO 209 S., DSFinV-K-Spezifikation 130 S., AEAO 41 S., UStAE 7,9 MB) nie ganz lesen, sondern gezielt die relevante Norm, den Anhang oder Abschnitt.
- Einstiegspunkt ist die jeweilige Stelle in compliance.md/steuerrecht.md. Die Spalte „compliance.md" in der folgenden Tabelle verweist darauf.

### Schnellzugriff nach Aufgabe

| Aufgabe in jotti | Quelldatei(en), relativ zu diesem Ordner | Gezielt lesen | compliance.md |
| ---------------- | ---------------------------------------- | ------------- | ------------- |
| TSE-Transaktion absichern (processType, processData, Start/Finish) | `verwaltung-bmf/AEAO-zu-146a-2023-06-30.pdf`; `technik-spezifikationen/DSFinV-K-2.4/20231215_DSFinV_K_2_4.pdf` | AEAO Nr. 2.2.3.5–2.2.3.6 (Art/Daten des Vorgangs); DSFinV-K Anhang I (processData-Formate) | §3 |
| fiskaly Cloud-TSE anbinden (TSS, Client, Transaktion, Export, Auth) | `fiskaly/fiskaly-SIGN-DE-API-v2-openapi.json` | Pfade `/tss`, `/tss/{tss_id}/client`, `/tss/{tss_id}/tx`, `/tss/{tss_id}/export` | §3.5 |
| DSFinV-K-Export (CSV-Dateinamen, Felder, index.xml, DTD) | `technik-spezifikationen/DSFinV-K-2.4/20231215_DSFinV_K_2_4.pdf`; `technik-spezifikationen/DSFinV-K-2.4/02_index.xml/index.xml`; `technik-spezifikationen/gdpdu-01-09-2004.dtd` | Modul- und Feldtabellen; DTD und XML ganz lesbar | §6 |
| Steuersätze, Kombi 70/30, USt-Schlüssel | `gesetze/UStG.pdf`; `verwaltung-bmf/UStAE-konsolidiert-2026-06-02.pdf`; `technik-spezifikationen/DSFinV-K-2.4/20241205_DSFinV_K_Anlage2_UStSchluessel.pdf` | UStG §§ 12, 24; UStAE Abschn. 10.1 Abs. 12 | §6.7; steuerrecht.md |
| Beleg- und Bondruck-Pflichtangaben | `gesetze/KassenSichV.pdf`; `gesetze/UStG.pdf`; `verwaltung-bmf/AEAO-zu-146a-2023-06-30.pdf` | KassenSichV § 6 (6 S., ganz); UStG § 14 | §5 |
| GoBD: Unveränderbarkeit, Aufbewahrung, Verfahrensdoku | `verwaltung-bmf/GoBD-Aenderung-2025-07-14.pdf`; `gesetze/AO.pdf` | AO §§ 146, 147; GoBD-Basistext extern (siehe Provenienz) | §4 |
| ELSTER-Meldepflicht | `verwaltung-bmf/Kassenmeldepflicht-BMF-2024-06-28.pdf`; `gesetze/AO.pdf` | AO § 146a Abs. 4; Beginn des Mitteilungsverfahrens und Fristen | §7 |
| Datenschutz (BEDIENER_NAME, Datenminimierung) | `gesetze/DSGVO-VO-EU-2016-679.pdf` | Art. 5 Abs. 1 lit. c | §6.4 |
| TSE: technische Anforderungen, Zertifizierung | `technik-spezifikationen/BSI-TR-03153-1.pdf` (+ Anhänge A/B); `technik-spezifikationen/BSI-TR-03153-2.pdf` | Teil 1 Kap. 9.3 (Seriennummer) und Anhang A (Zertifizierung); Teil 2: übergangsweise Aufrechterhaltung der Zertifizierung | §3.1 |
| Gemeinnützigkeit, Zweckbetrieb, Kleinunternehmer | `gesetze/AO.pdf`; `gesetze/UStG.pdf` | AO §§ 14, 64, 67a; UStG § 19 | §2; steuerrecht.md |

## Struktur

| Ordner                     | Inhalt                                                                                                      |
| -------------------------- | ----------------------------------------------------------------------------------------------------------- |
| `gesetze/`                 | Formelle Gesetze und Verordnungen (AO, UStG, KassenSichV, DSGVO)                                            |
| `verwaltung-bmf/`          | BMF-Schreiben und Anwendungserlasse (GoBD, AEAO, UStAE, Kassenmeldepflicht)                                 |
| `technik-spezifikationen/` | Technische Normen (BSI TR-03153, DSFinV-K samt DTD und USt-Schlüssel)                                       |
| `fiskaly/`                 | OpenAPI-Spezifikation der fiskaly Cloud-TSE (Schnittstelle, gegen die jottis `TSEClient` implementiert ist) |

## Inhalt und Bezug zu jotti

### gesetze/

| Datei                      | Norm                                          | Relevanz für jotti                                                                                                                                  |
| -------------------------- | --------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| `AO.pdf`                   | Abgabenordnung (Volltext, 209 S.)             | Enthält §§ 14, 64, 67a (Gemeinnützigkeit/Zweckbetrieb), 146, 146a, 147 (TSE, Belegausgabe, Meldepflicht, Aufbewahrung), 379 (Steuergefährdung)      |
| `UStG.pdf`                 | Umsatzsteuergesetz (Volltext, 98 S.)          | §§ 12 Abs. 2 Nr. 15 (7 % auf Speisen ab 2026), 14 (Belegpflichtangaben), 19 (Kleinunternehmer), 24 (Durchschnittssätze, processData-Positionen 3/4) |
| `KassenSichV.pdf`          | Kassensicherungsverordnung (Volltext, 6 S.)   | §§ 1, 4, 6: Definition Aufzeichnungssystem, digitale Schnittstelle, Belegausgabe und TSE-Pflichtdaten                                               |
| `DSGVO-VO-EU-2016-679.pdf` | Datenschutz-Grundverordnung (Volltext, 88 S.) | Art. 5 Abs. 1 lit. c (Datenminimierung): Begründung für `BEDIENER_NAME` = interner Benutzername statt Klarname (§6.4)                               |

### verwaltung-bmf/

| Datei                                           | Dokument                                                               | Relevanz für jotti                                                                                   |
| ----------------------------------------------- | ---------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------- |
| `AEAO-zu-146a-2023-06-30.pdf`                   | Anwendungserlass zur AO zu § 146a (Neufassung, 41 S.)                  | Offizielle processType-Definitionen (Nr. 2.2.3.5–2.2.3.6), Eingabegeräte (BYOD), Seriennummer, Belegpflicht |
| `AEAO-zu-146a-Aenderung-2024-06-28.pdf`         | AEAO zu § 146a, Änderung 28.06.2024                                    | Fortschreibung des Anwendungserlasses                                                                |
| `AEAO-zu-146-146a-147-Aenderung-2025-09-01.pdf` | AEAO zu §§ 146, 146a, 147, Änderung 01.09.2025                         | Jüngste Fortschreibung; maßgeblich für den aktuellen Stand                                           |
| `GoBD-Aenderung-2024-03-11.pdf`                 | GoBD, Änderung 11.03.2024                                              | Fortschreibung der GoBD-Grundsätze (Unveränderbarkeit, Nachvollziehbarkeit, Verfahrensdokumentation) |
| `GoBD-Aenderung-2025-07-14.pdf`                 | GoBD, Änderung 14.07.2025                                              | Jüngste GoBD-Fortschreibung; maßgeblich für den aktuellen Stand                                      |
| `Kassenmeldepflicht-BMF-2024-06-28.pdf`         | BMF-Schreiben Mitteilungsverpflichtung § 146a Abs. 4 AO                | ELSTER-Meldepflicht: Beginn des Mitteilungsverfahrens und Fristen                                    |
| `UStAE-konsolidiert-2026-06-02.pdf`             | Umsatzsteuer-Anwendungserlass (konsolidiert, Stand 02.06.2026, 7,9 MB) | Abschn. 10.1 Abs. 12: 70/30-Pauschalierung für Kombinationsangebote (Kombi-Steuersatz)               |

### technik-spezifikationen/

| Datei                                                      | Dokument                                           | Relevanz für jotti                                                                                                            |
| ---------------------------------------------------------- | -------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| `BSI-TR-03153-1.pdf`                                       | BSI TR-03153 Teil 1, Hauptdokument v1.1.1          | Anforderungen an die TSE; Grundlage der fiskaly-Zertifizierung. v1.1.1 ist laut BSI die aktuelle Fassung, v1.0.1 ist Legacy   |
| `BSI-TR-03153-1-Anhang-A.pdf`                              | TR-03153 Teil 1, Anhang A                          | Zertifizierungsanforderungen                                                                                                  |
| `BSI-TR-03153-1-Anhang-B.pdf`                              | TR-03153 Teil 1, Anhang B                          | Betriebsanforderungen                                                                                                         |
| `BSI-TR-03153-2.pdf`                                       | BSI TR-03153 Teil 2                                | Regelung zur übergangsweisen Aufrechterhaltung der TSE-Zertifizierung in begründeten Ausnahmefällen                            |
| `DSFinV-K-2.4.zip`                                         | DSFinV-K v2.4 Gesamtpaket (BZSt, Stand 15.12.2023) | Spezifikation, Anhänge, Beispiele, DTD und USt-Schlüssel; entpackt unter `DSFinV-K-2.4/`. v2.4 ist die jüngste auf bzst.de veröffentlichte Fassung, eine 2.5 existiert dort nicht |
| `DSFinV-K-2.4/20231215_DSFinV_K_2_4.pdf`                   | DSFinV-K Spezifikation (Volltext)                  | Tabellenstruktur, Dateinamen, Nr. 2.7 und Anhang H (Durchbedienen), Anhang I (processData)                                    |
| `DSFinV-K-2.4/20241205_DSFinV_K_Anlage2_UStSchluessel.pdf` | Anlage 2: USt-Schlüssel                            | Maßgebliche USt-Schlüssel für `transactions_vat.csv` / `lines_vat.csv`                                                        |
| `gdpdu-01-09-2004.dtd`                                     | GDPdU-Beschreibungsstandard (DTD)                  | Pflicht-DTD im DSFinV-K-Export; auch unter `DSFinV-K-2.4/02_index.xml/` mit Beispiel-`index.xml`                              |

Das DSFinV-K-Paket enthält zusätzlich BSI-TR-03153-Fassungen, das Einführungsschreiben zu § 146a AO und die GoBD-Änderung 03/2024 (im Unterordner `01_rechtliche Grundlagen/`) sowie XLSX-Beispiele zur Vorgangsbehandlung.

### fiskaly/

| Datei                                 | Dokument                                               | Relevanz für jotti                                                                                                                                                                                                                                                                          |
| ------------------------------------- | ------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `fiskaly-SIGN-DE-API-v2-openapi.json` | OpenAPI 3.0.3, fiskaly SIGN DE API, Spec-Version 2.2.2 | Maschinenlesbarer Vertrag der Cloud-TSE: `/tss` (TSS anlegen), `/tss/{id}/client` (Client registrieren), `/tss/{id}/tx` (Start/Finish einer Transaktion), `/tss/{id}/export` (TAR/DSFinV-K-Export), `/auth`. Referenz für die Konformität des `TSEClient`-Adapters (compliance.md §3.5) |

Quelle: `https://kassensichv.fiskaly.com/api/v2/_spec.json` (Spec hinter der Redoc-Doku unter `https://kassensichv.fiskaly.com/api/v2/_docs/`). Die fließtextliche Entwicklerdokumentation liegt unter `https://developer.fiskaly.com/api/kassensichv/v2` und ist nicht als geschlossene Datei abrufbar; maßgeblich für ein Audit ist die OpenAPI-Spezifikation.

## Provenienz und Quellen-URLs

| Datei                                                          | Quelle (Abruf 24.06.2026)                                                                               |
| -------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------- |
| `gesetze/AO.pdf`                                               | https://www.gesetze-im-internet.de/ao_1977/AO.pdf                                                       |
| `gesetze/UStG.pdf`                                             | https://www.gesetze-im-internet.de/ustg_1980/UStG.pdf                                                   |
| `gesetze/KassenSichV.pdf`                                      | https://www.gesetze-im-internet.de/kassensichv/KassenSichV.pdf                                          |
| `gesetze/DSGVO-VO-EU-2016-679.pdf`                             | https://eur-lex.europa.eu/legal-content/DE/TXT/PDF/?uri=CELEX:32016R0679                                |
| `verwaltung-bmf/AEAO-zu-146a-2023-06-30.pdf`                   | bundesfinanzministerium.de, AO-Anwendungserlass, 2023-06-30-AEAO-Par-146-AO.pdf                         |
| `verwaltung-bmf/AEAO-zu-146a-Aenderung-2024-06-28.pdf`         | bundesfinanzministerium.de, AO-Anwendungserlass, 2024-06-28-aenderung-aeao-146a.pdf                     |
| `verwaltung-bmf/AEAO-zu-146-146a-147-Aenderung-2025-09-01.pdf` | bundesfinanzministerium.de, AO-Anwendungserlass, 2025-09-01-aenderung-aeao-146-usw.pdf                  |
| `verwaltung-bmf/GoBD-Aenderung-2024-03-11.pdf`                 | bundesfinanzministerium.de, AO-Anwendungserlass, 2024-03-11-aenderung-gobd.pdf (auch im DSFinV-K-Paket) |
| `verwaltung-bmf/GoBD-Aenderung-2025-07-14.pdf`                 | bundesfinanzministerium.de, 2025-07-14-GoBD-2-aenderung.pdf                                             |
| `verwaltung-bmf/Kassenmeldepflicht-BMF-2024-06-28.pdf`         | bundesfinanzministerium.de, 2024-06-28-mitteilungsverpflichtung-nach-AO.pdf                             |
| `verwaltung-bmf/UStAE-konsolidiert-2026-06-02.pdf`             | bundesfinanzministerium.de, Umsatzsteuer-Anwendungserlass-aktuell.pdf (Stand 02.06.2026)                |
| `technik-spezifikationen/BSI-TR-03153-*.pdf`                   | bsi.bund.de, TR03153 (Teil 1 v1.1.1, Anhang A v1.1.1, Anhang B v1.1.0, Teil 2 v1.0.0)                   |
| `technik-spezifikationen/DSFinV-K-2.4.zip`                     | bzst.de, dsfinv_k_v_2_4.zip                                                                             |
| `fiskaly/fiskaly-SIGN-DE-API-v2-openapi.json`                  | kassensichv.fiskaly.com/api/v2/\_spec.json                                                              |
| BMF-FAQ zu § 146a AO (§9 [11], nur HTML)                      | https://www.bundesfinanzministerium.de/Content/DE/FAQ/FAQ-steuergerechtigkeit-belegpflicht.html         |
| ELSTER für Entwickler (§9 [6], nur HTML)                      | https://www.elster.de/elsterweb/infoseite/entwickler                                                    |

Hinweis zum erneuten Abruf: Die BMF-, BSI- und BZSt-Downloads benötigen einen Browser-User-Agent; die Versionsparameter (`v=`) in den BMF-URLs veralten und ändern sich bei jeder Aktualisierung. Integrität der abgelegten Dateien über `SHA256SUMS.txt` prüfbar (`sha256sum -c SHA256SUMS.txt`).
