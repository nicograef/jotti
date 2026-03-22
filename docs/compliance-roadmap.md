# Compliance-Roadmap — jotti

> **Status:** Aktiv — Phasenweise Umsetzung der KassenSichV/TSE-Anforderungen
> **Zuletzt aktualisiert:** 2026-03-21
> **Basis:** [docs/compliance.md](compliance.md), [docs/anforderungen.md §7](anforderungen.md#7--fiskalkonformität)

jotti ist ein elektronisches Aufzeichnungssystem nach § 1 KassenSichV und unterliegt der TSE-Pflicht nach § 146a AO. Die Compliance-Anforderungen werden phasenweise in vier Phasen umgesetzt.

---

## Phase 0 — Baseline (aktuell implementiert)

Diese Phase beschreibt den **aktuellen Stand** des Systems. Grundlegende Compliance-Bausteine sind bereits vorhanden.

| #   | Feature                          | Anforderung                                 | Status                     |
| --- | -------------------------------- | ------------------------------------------- | -------------------------- |
| 0.1 | **Event-Sourcing (Append-Only)** | GoBD — Unveränderbarkeit                    | ✅                         |
| 0.2 | **Kassenjournal**                | GoBD — Nachvollziehbarkeit, Vollständigkeit | ✅                         |
| 0.3 | **Belegausgabe via Bondrucker**  | F-03, § 146a Abs. 2 AO                      | ✅ Basis (ohne TSE-Felder) |
| 0.4 | **Tagesabrechnung / Reporting**  | § 5 Anforderungen                           | ✅                         |
| 0.5 | **Rollenbasierter Zugriff**      | Security, GoBD                              | ✅                         |

**Offene Phase-0-Punkte:**

| #   | Feature                       | Anforderung                  | Status                   |
| --- | ----------------------------- | ---------------------------- | ------------------------ |
| 0.6 | **Dokumentation korrigieren** | (Compliance, Positionierung) | ✅ Erledigt (2026-03-21) |

---

## Phase 1 — Compliance-Grundlage

Diese Phase legt die **rechtlich notwendige Grundlage** für den Betrieb: Seriennummer, Steuersätze, Abrechnungskreis und ELSTER-Anleitung. Phase 1 muss abgeschlossen sein, bevor Phase 2 beginnt.

### Features

| ID   | Feature                                        | Anforderung                       | Priorität   | Aufwand |
| ---- | ---------------------------------------------- | --------------------------------- | ----------- | ------- |
| F-01 | **Seriennummer der Kasse**                     | § 146a AO, DSFinV-K, Kassenbeleg  | Must        | Gering  |
| F-07 | **Steuersätze (19 % / 7 % / 0 %)**             | KassenSichV, DSFinV-K             | Must        | Mittel  |
| F-06 | **Abrechnungskreis (tagesbasiert, pro Tisch)** | DSFinV-K Nr. 2.7                  | Should      | Mittel  |
| F-03 | **Belegausgabe — Pflichtfelder**               | § 146a Abs. 2 AO, § 6 KassenSichV | Must        | Mittel  |
| F-05 | **ELSTER-Anleitung**                           | § 146a Abs. 4 AO                  | Must (Doku) | Gering  |
| —    | **Verfahrensdokumentation**                    | GoBD                              | Empfohlen   | Gering  |

### Details

#### F-01 Seriennummer der Kasse

- UUID-v4 beim ersten Containerstart generieren
- Dauerhaft in `system_config` (Schlüssel `kassen_id`) speichern — nie überschreiben
- Im Admin-Dashboard prominent anzeigen (Kopierbutton)
- API-Endpunkt `/admin/kasse/seriennummer` für programmatischen Abruf
- Auf Kassenbeleg und im DSFinV-K-Export (`Stamm_Kassen.csv`, Feld `KASSE_SERIENNR`) verwenden

#### F-07 Steuersätze

- Pflichtfeld `steuersatz` für Produkte: Enum `standard` (19 %), `ermaessigt` (7 %), `befreit` (0 %)
- Im Admin-Bereich bei Produktanlage/-bearbeitung auswählbar
- Als Fat Event in `BestellungAufgenommen` einfrieren (für historische Korrektheit)
- Tagesabrechnung weist Umsätze nach Steuersatz aufgeschlüsselt aus

#### F-06 Abrechnungskreis (Phase 1: tagesbasiert, pro Tisch)

- DB-Tabelle `abrechnungskreis` mit fortlaufender Nummer, Start- und Endzeitpunkt
- **Tageseröffnung durch Admin** (bewusste Admin-Aktion, morgens/vormittags) eröffnet den Abrechnungskreis — kein automatisches Eröffnen
- Format der Session-ID: `Tisch-{Nr}-{YYYYMMDD}` — das Datum stammt aus `DATE(abrechnungskreis.beginn)`, nicht aus `NOW()` des Ereignisses (relevant bei Betrieb über Mitternacht)
- Beim Tagesabschluss werden alle laufenden Sessions des Tages geschlossen — neue Sessions entstehen erst wieder durch die nächste Tagesöffnung
- **Kassenbetrieb-Sperre:** Bestellungen, Zahlungen und sonstige Tischoperationen sind nur möglich, wenn ein Abrechnungskreis mit Status `offen` existiert
- Alle Events eines Tisches sind einem `ABRECHNUNGSKREIS` zugeordnet
- Hinweis: Mehrere Gästegruppen an einem Tisch teilen denselben `ABRECHNUNGSKREIS` — zulässig, aber Phase 2 (manuelle Tischfreigabe) ist für jottis Festzelt-Betrieb die korrektere Lösung

#### F-03 Belegausgabe — Pflichtfelder

- Beleg enthält alle Pflichtfelder nach § 6 KassenSichV:
  - Datum/Uhrzeit des Vorgangs
  - Betrag, Steuersatz, Steuerbetrag pro Position
  - Zahlungsart (Bar)
  - Seriennummer der Kasse (F-01)
- TSE-Pflichtfelder (Signatur, Zähler) werden in Phase 2 ergänzt

#### F-05 ELSTER-Anleitung

- Datei `docs/betrieb/elster-meldung.md` mit Schritt-für-Schritt-Anleitung für manuelle ELSTER-Meldung
- Admin-Dashboard zeigt Meldepflicht-Hinweis mit Link zur Anleitung und der Seriennummer an
- Manuell setzbarer Meldestatus in der Admin-UI: `ausstehend` / `gemeldet am TT.MM.JJJJ`

#### Verfahrensdokumentation

- Datei `docs/betrieb/verfahrensdokumentation.md`
- Beschreibt für Betriebsprüfer: Systemarchitektur, Datenbankschutz, TSE-Anbindung, Zugriffskontrollen
- Vorlage, die Betreiber mit ihren Vereinsdaten ergänzen

---

## Phase 2 — TSE-Integration

Diese Phase integriert eine **zertifizierte Cloud-TSE** (fiskaly) und schafft den DSFinV-K-Export. Nach Abschluss ist jotti vollständig KassenSichV-konform.

### Features

| ID   | Feature                                          | Anforderung             | Priorität | Aufwand |
| ---- | ------------------------------------------------ | ----------------------- | --------- | ------- |
| F-02 | **TSEClient-Interface + fiskaly-Adapter**        | § 146a AO, BSI TR-03153 | Should    | Hoch    |
| —    | **TSE-Hooks auf Zahlungsfluss**                  | § 146a AO               | Should    | Hoch    |
| —    | **Event-Daten um TSE-Felder erweitern**          | DSFinV-K                | Should    | Mittel  |
| F-04 | **DSFinV-K-Export**                              | § 4 KassenSichV         | Should    | Hoch    |
| —    | **Z-Bon-Logik (Kassenführung KF-07)**            | DSFinV-K                | Must      | Mittel  |
| —    | **TSE-Felder auf Beleg (ergänzt F-03)**          | § 6 KassenSichV         | Should    | Gering  |
| —    | **QR-Code auf Beleg**                            | DSFinV-K Anhang I       | Nice      | Gering  |
| —    | **Abrechnungskreis Phase 2** (manuelle Freigabe) | DSFinV-K                | Should    | Mittel  |

### Details

#### F-02 TSEClient-Interface + fiskaly-Adapter

- Go-Interface `TSEClient` mit Methoden: `StartTransaction`, `UpdateTransaction`, `FinishTransaction`
- `FiskalyTSEClient` als erster Anbieter (REST-API, Cloud-TSE, BSI-zertifiziert)
- Konfiguration über Umgebungsvariablen: `TSE_PROVIDER`, `FISKALY_API_KEY`, `FISKALY_API_SECRET`
- Betriebsmodus `strict` (Fehler bei fehlender TSE) / `bypass` (nur für Entwicklung/Tests)

#### TSE-Hooks auf Zahlungsfluss (Atomares Transaktionsmodell)

Jeder jotti-Vorgang ist eine eigenständige, sofort geschlossene TSE-Transaktion:

| jotti-Vorgang        | TSE-Operation             | processType                |
| -------------------- | ------------------------- | -------------------------- |
| Bestellung aufnehmen | `Start` + sofort `Finish` | `Bestellung-V1`            |
| Zahlung kassieren    | `Start` + sofort `Finish` | `Kassenbeleg-V1`           |
| Stornierung          | `Start` + sofort `Finish` | `Kassenbeleg-V1` (negativ) |
| Tagesabschluss       | `Start` + sofort `Finish` | `SonstigerVorgang-V1`      |

#### F-04 DSFinV-K-Export

- Adminendpunkt `POST /admin/export/dsfinvk` — ZIP-Archiv mit:
  - `Stamm_Abschluss.csv`, `Stamm_Kassen.csv`, `Stamm_TSE.csv`, `Stamm_Orte.csv`
  - `Bonkopf.csv`, `Bonkopf_USt.csv`, `Bonkopf_Zahlarten.csv`
  - `Bonpos.csv`, `Bonpos_USt.csv`, `TSE_Transaktionen.csv`
  - `Z_GV_Typ.csv`, `Z_Zahlart.csv`
  - `index.xml`
- Alle Dateinamen nach DSFinV-K-Spezifikation v2.4 (deutsch, exakt)
- Steuersatz-Aufschlüsselung korrekt pro Position und Bon

#### Abrechnungskreis Phase 2 — Manuelle Tischfreigabe

Für jottis Festzelt-Betrieb (Vereinsfest, Maihock) sitzt häufig mehr als eine Gästegruppe pro Tag an einem Tisch. Phase 2 ist deshalb für diesen Primär-Anwendungsfall die korrekte Abbildung und sollte zeitnah nach Phase 2 umgesetzt werden:

- Servicekräfte können einen Tisch nach dem Bezahlen explizit „freigeben" (UI-Aktion „Tisch freimachen")
- Eine neue Tisch-Session erhält ein laufendes Buchstaben-Suffix: `Tisch-42-20260501-A`, `Tisch-42-20260501-B`, etc.
- Der erste `ABRECHNUNGSKREIS` des Tages wird ohne Suffix vergeben (`Tisch-{Nr}-{YYYYMMDD}`) und beim ersten Gästewechsel in `-A`/`-B` umgestellt
- Neue Domain-Aktion `TischFreigegeben` im Tisch-Aggregat; neue UI-Aktion in der Servicekraft-Ansicht

---

## Phase 3 — Erweiterungen

Optional / Nice-to-have — kein harter Compliance-Bedarf, aber sinnvolle Verbesserungen.

| ID   | Feature                                       | Anforderung                          | Priorität | Aufwand |
| ---- | --------------------------------------------- | ------------------------------------ | --------- | ------- |
| —    | **Digitaler eBeleg (QR-Code / Downloadlink)** | § 146a Abs. 2 AO                     | Nice      | Mittel  |
| —    | **Automatisierte ELSTER-Meldung**             | § 146a Abs. 4 AO                     | Nice      | Hoch    |
| F-08 | **GoBD-Hash-Chain auf Events**                | GoBD — Unveränderbarkeit (erweitert) | Nice      | Mittel  |
| —    | **10-Jahres-Archivierungsstrategie**          | GoBD, §§ 146/147 AO                  | Nice      | Mittel  |

### Details

#### Digitaler eBeleg

- Nach TSE-`FinishTransaction` QR-Code anzeigen (Download-Link zu PDF/PNG)
- Beleg abrufbar auf Backend-Server für konfigurierbaren Zeitraum
- Gilt als Belegausgabe mit konkludenter Einwilligung des Gastes (BMF-FAQ)
- Primär: Bondrucker bleibt Standardweg; eBeleg als Alternative/Ergänzung
- Speicherung für DSFinV-K-Archiv

#### Automatisierte ELSTER-Meldung

- ERiC-Integration (native C-Library, kein Vendor-Lock-in) — oder —
- fiskaly-Submission-API (einfacher, aber Cloud-Abhängigkeit)
- Entscheidung bei Implementierungsbeginn: Aufwands-/Kosten-Abwägung

#### F-08 GoBD-Hash-Chain

- Jedes Event in `events`-Tabelle speichert SHA-256-Hash des vorherigen Events (`previous_hash`)
- Erstes Event jedes `ABRECHNUNGSKREIS` verwendet definierten Genesis-Hash
- Integritätsprüfungs-Endpunkt `POST /admin/integrity/check` validiert vollständige Hash-Chain
- Ergänzt TSE-Signatur — unabhängige Manipulationssicherheit

---

## Übersicht aller Compliance-Anforderungen

| ID   | Anforderung               | Phase | Priorität   | Status   |
| ---- | ------------------------- | ----- | ----------- | -------- |
| F-01 | Seriennummer der Kasse    | 1     | Must        | 🔲       |
| F-02 | TSE-Adapter-Schnittstelle | 2     | Should      | 🔲       |
| F-03 | Belegausgabepflicht       | 0/1/2 | Must        | ✅ Basis |
| F-04 | DSFinV-K-Export           | 2     | Should      | 🔲       |
| F-05 | ELSTER-Meldung            | 1     | Must (Doku) | 🔲       |
| F-06 | Abrechnungskreis          | 1/2   | Should      | 🔲       |
| F-07 | Steuersätze               | 1     | Must        | 🔲       |
| F-08 | GoBD-Hash-Chain           | 3     | Nice        | 🔲       |

**Legende:** ✅ Implementiert · 🔲 Offen · ⏳ In Arbeit

---

## Betreiber-Hinweise

> Diese Abschnitte richten sich an Vereinsvorstände und IT-Verantwortliche, die jotti betreiben.

### Pflichten vor dem ersten Einsatz

1. **Cloud-TSE-Vertrag:** Vertrag mit fiskaly (oder anderem BSI-zertifizierten Cloud-TSE-Anbieter) abschließen. API-Schlüssel als Umgebungsvariablen in die `.env`-Datei eintragen.
2. **ELSTER-Meldung:** Nach der ersten Inbetriebnahme innerhalb von **einem Monat** die jotti-Instanz beim zuständigen Finanzamt über [ELSTER](https://www.elster.de) anmelden. Benötigte Daten: Seriennummer der Kasse (im Admin-Dashboard), Softwarename „jotti", Inbetriebnahmedatum.
3. **Seriennummer sichern:** Die Kassen-UUID in den System-Stammdaten ist die rechtliche Identität der Kasse. Das Datenbank-Backup muss diese enthalten. Bei Verlust: alte Seriennummer abmelden, neue Instanz mit neuer Seriennummer neu anmelden.

### Laufende Pflichten

- **10-Jahres-Aufbewahrung:** Alle Kassendaten (Events, DSFinV-K-Exporte) sind 10 Jahre aufzubewahren (§§ 146, 147 AO, GoBD). Sicherstellen, dass Backups entsprechend archiviert und jederzeit lesbar sind.
- **Regelmäßige Backups:** Tägliche Datenbank-Backups sind Pflicht — nicht nur für die Compliance, sondern auch zur Seriennummern-Sicherung.
- **Außerbetriebnahme melden:** Wenn eine jotti-Instanz dauerhaft stillgelegt wird, muss dies innerhalb von einem Monat bei ELSTER gemeldet werden.

---

## Referenzen

- [docs/compliance.md](compliance.md) — Vollständige rechtliche Grundlagen und technische Anforderungsdetails
- [docs/anforderungen.md §7](anforderungen.md#7--fiskalkonformität) — Anforderungen F-01 bis F-08 mit Akzeptanzkriterien
- [docs/language.md — Fiskalkonformität](language.md#fiskalkonformität-compliance-sub-domain) — Terminologie-Definitionen
