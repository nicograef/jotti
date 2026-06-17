# Anforderungen: jotti

Alle funktionalen und querschnittlichen Anforderungen an jotti. Umgesetzte Features sind als kompakte Tabelle zusammengefasst; offene Features enthalten vollständige Akzeptanzkriterien.

---

## Legende

| Kürzel       | Bedeutung                                                |
| ------------ | -------------------------------------------------------- |
| Must-have    | Kernfunktion, ohne die das System nicht nutzbar ist      |
| Should-have  | Wichtig für einen runden Betrieb, aber nicht blockierend |
| Nice-to-have | Komfortfunktion, die den Alltag erleichtert              |

| Symbol | Bedeutung  |
| ------ | ---------- |
| ✅     | Umgesetzt  |
| 🔲     | Offen      |
| 🚫     | Won't-have |

---

## Rollen und Berechtigungen

| Code-Rolle       | Beschreibung                                                         |
| ---------------- | -------------------------------------------------------------------- |
| `admin`          | Voller Zugriff auf Stammdaten (Produkte, Tische, Benutzer) und Kasse |
| `serviceleitung` | Kasse einschließlich Stornierung und Auszahlung                      |
| `service`        | Kasse ohne Stornierung                                               |

→ Vollständige Berechtigungsmatrix: handbuch.md §5.1

---

## 1 · Kasse (Core Domain)

### Umgesetzt

| ID   | Titel                   | Beschreibung                                                                                                                                                                                    |
| ---- | ----------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| K-01 | Bestellung aufnehmen    | Servicekraft wählt Tisch, stellt aus Produktkatalog (nach Kategorien) eine Bestellung zusammen und gibt ab.                                                                                     |
| K-02 | Zahlung registrieren    | Barzahlung mit Positions-Auswahl (Teilzahlung möglich). Saldo wird reduziert.                                                                                                                   |
| K-03 | Ausgabe bestätigen      | Bestellte Positionen als ausgegeben markieren. Nachverfolgung ausstehender Positionen.                                                                                                          |
| K-04 | Stornierung erteilen    | Serviceleitung/Admin stornieren Positionen, unabhängig von Ausgabe-/Bezahlstatus. Saldo kann negativ werden.                                                                                    |
| K-05 | Auszahlung leisten      | Serviceleitung/Admin gleichen negativen Tischsaldo durch positionsunabhängige Auszahlung aus.                                                                                                   |
| K-06 | Tischübersicht          | Dashboard mit „Meine Tische" (Favoriten), Alle-Tische-Drawer, Tisch-Detail mit Tabs (Bestellen, Kassieren, Historie).                                                                           |
| K-07 | Kassenjournal           | Unveränderliche Event-Tabelle (append-only) als Single Source of Truth. Synchrone Projektion + CRUD-Entität.                                                                                    |
| K-09 | Bestellungen umbuchen   | Serviceleitung/Admin buchen unbezahlte Bestellungen atomar zwischen Tischen um (Quell-Storno + Ziel-Bestellung in einer Transaktion).                                                           |
| K-10 | Rückgeldberechnung      | Optionale Eingabe von erhaltenem Betrag und Zielbetrag (inkl. Trinkgeld) beim Kassieren. Rückgeld und Trinkgeld werden rein clientseitig berechnet und angezeigt.                               |
| K-11 | Tisch-Schnellsuche      | Suchfeld im Alle-Tische-Drawer. Clientseitige Echtzeit-Filterung nach Tischname (case-insensitive).                                                                                             |
| K-12 | Arbeitsbon              | Automatischer Arbeitsbon (ohne Preise) bei Bestellaufnahme an konfigurierte Druckstationen (Küche, Theke). Operativ, nicht-fiskalisch, kein Kassenbeleg (→ F-03).                               |
| K-14 | Tisch-Favoriten         | Serverseitig gespeicherte Favoriten pro Benutzer. Stern-Toggle im Alle-Tische-Drawer.                                                                                                           |
| K-16 | Kassensitzung eröffnen  | Global nummerierter Betriebstag. Kassensitzung-Sperre blockiert Betrieb ohne offene Sitzung.                                                                                                    |
| K-17 | Anfangsbestand setzen   | Wechselgeld als Basis für Kassenbestandsführung. Genau einmal pro Kassensitzung.                                                                                                                |
| K-18 | Kassenbestand einsehen  | Soll-Bestand als SQL-Aggregation über Kassenjournal mit Aufschlüsselung nach Komponenten.                                                                                                       |
| K-19 | Geldtransit buchen      | Geldtransit (Einlage oder Entnahme) als Event im Kassenjournal.                                                                                                                                 |
| K-20 | Betreiber-Stammdaten    | Admin pflegt Vereinsname, Adresse, Steuernummer und USt-ID. Erscheint auf Beleg (F-03) und DSFinV-K-Export (F-04).                                                                              |
| K-21 | Kassensturz durchführen | Gezählter Ist-Bestand vs. Soll-Bestand. Differenz wird automatisch gebucht.                                                                                                                     |
| K-22 | Tagesabschluss / Z-Bon  | Formaler Tagesabschluss. Schließt Kassensitzung ab. Voraussetzung: Kassensturz + alle Tische auf Saldo 0.                                                                                       |
| K-24 | Direktverkauf           | Barverkauf an der Theke: bestellen + zahlen + ausgeben in einem Schritt. Eigener Event-Stream pro Verkauf (`direktverkauf-getaetigt:v1`), sofort kassenwirksam, ohne Tisch und ohne Projektion. |

### Offen

---

#### K-13 · Küchendisplay (KDS)

> **Rolle:** Servicekraft · Serviceleitung · Admin · **Prio:** Should-have

Mitarbeiter an den Ausgabestationen (Getränketheke, Küche) sehen auf einem eigenen Bildschirm in Echtzeit die eingehenden Bestellungen ihrer Kategorie. Das Display ist eine passive Anzeige, es zeigt offene Bestellungen gruppiert nach Tisch, sodass Ausgabestationen auch bei Bon-Verlust die Bestellungen nachvollziehen können.

**Akzeptanzkriterien:**

- [ ] Echtzeit-Anzeige offener Bestellungen nach Kategorie (Essen, Getränke), gruppiert nach Tisch
- [ ] Getränkeausgabe sieht offene Getränkebestellungen, Essensausgabe sieht offene Essensbestellungen
- [ ] Letzte Bestellungen sind einsehbar (bei Bon-Verlust)

---

#### K-15 · Ausgabestationen mit Zubereitungsstatus

> **Rolle:** Servicekraft · Serviceleitung · Admin · **Prio:** Nice-to-have

Aufbauend auf dem Küchendisplay (K-13) können Mitarbeiter an Ausgabestationen den Zubereitungsstatus einzelner Positionen verwalten. Servicekräfte sehen den Zubereitungsstatus und wissen, wann Positionen abholbereit sind.

**Akzeptanzkriterien:**

- [ ] Positionen können als „in Zubereitung" und „fertig" markiert werden
- [ ] Servicekraft kann den Zubereitungsstatus ihrer Bestellungen einsehen

---

#### K-23 · Manuelle Tischfreigabe

> **Rolle:** Servicekraft · Serviceleitung · Admin · **Prio:** Nice-to-have

Für jottis Festzelt-Betrieb (Vereinsfest, Maihock) die korrektere Abbildung der Realität: mehrere Gästegruppen an einem Tisch erhalten jeweils einen eigenen Abrechnungskreis (Tisch-Session). Nach Abschluss einer Gästegruppe kann die Servicekraft den Tisch freigeben, sodass eine neue Tisch-Session mit Suffix gestartet wird (z. B. `kassensitzung-1/tisch-42-b`).

**Akzeptanzkriterien:**

- [ ] Domain-Aktion `TischFreigegeben` startet eine neue Tisch-Session mit Suffix
- [ ] UI-Aktion „Tisch freimachen" in der Servicekraft-Ansicht
- [ ] Voraussetzung: Tisch-Saldo = 0

---

## 2 · Stammdaten (Supporting Domain)

### Umgesetzt

| ID   | Titel              | Beschreibung                                                                                    |
| ---- | ------------------ | ----------------------------------------------------------------------------------------------- |
| S-01 | Produktverwaltung  | Admin verwaltet Produktkatalog mit Kategorien und Varianten. Soft-Delete, Preise in Cent.       |
| S-02 | Tischverwaltung    | Admin verwaltet Tische (Name, Status). Nur aktive Tische in der Service-Übersicht. Soft-Delete. |
| S-03 | Benutzerverwaltung | Admin verwaltet Benutzerkonten mit Rollen. Passwort-Reset generiert 6-stelliges Einmalpasswort. |

---

## 3 · Auth (Infrastruktur)

### Umgesetzt

| ID   | Titel           | Beschreibung                                                                                   |
| ---- | --------------- | ---------------------------------------------------------------------------------------------- |
| A-01 | Login           | Benutzername + Passwort → JWT (12h, Argon2id). Generische Fehlermeldung bei Fehlversuch.       |
| A-02 | Passwort setzen | 6-stelliges Einmalpasswort → automatische Weiterleitung zu „Passwort setzen" (min. 6 Zeichen). |
| A-03 | Logout          | JWT aus Speicher entfernen, Weiterleitung auf Login-Seite.                                     |

---

## 4 · Querschnittsanforderungen

### Umgesetzt

| ID   | Titel             | Beschreibung                                                                                                                                                                                                                              |
| ---- | ----------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Q-01 | Mobile-first      | Alle Seiten auf Smartphones ab 360 px bedienbar. Touch-optimiert, Drawer-Overlays.                                                                                                                                                        |
| Q-02 | Mehrbenutzerfähig | Parallele Zugriffe ohne Datenverlust. Optimistic Concurrency Control.                                                                                                                                                                     |
| Q-03 | Validierung       | Zod (Frontend) + zog (Backend). Doppelte Validierung, deutsche Fehlermeldungen.                                                                                                                                                           |
| Q-04 | Datenintegrität   | Transaktionssicher, append-only Kassenjournal, Cent-Werte, Soft-Deletes.                                                                                                                                                                  |
| Q-06 | HTTPS / TLS       | Lokal: Caddy terminiert TLS, vertrauenswürdiges Let's-Encrypt-Zertifikat via `lokal.jotti.rocks` (DNS-01), selbstsignierter Fallback auf `https://<LAN-IP>` (einmalige Browserwarnung), HTTP→HTTPS-Redirect. Prod: nginx + Let's Encrypt. |
| Q-07 | Rate Limiting     | Login-Endpunkt geschützt (HTTP 429 bei Überschreitung).                                                                                                                                                                                   |
| Q-08 | Security Headers  | CSP, X-Content-Type-Options, X-Frame-Options, HSTS.                                                                                                                                                                                       |

### Offen

---

#### Q-05 · Offline-Fähigkeit

> **Rolle:** Servicekraft · Serviceleitung · Admin · **Prio:** Nice-to-have

Bei einem Internetausfall während der Veranstaltung soll die Bestellaufnahme weiterhin möglich sein. Bestellungen werden lokal zwischengespeichert und bei Wiederherstellung der Verbindung automatisch synchronisiert.

**Akzeptanzkriterien:**

- [ ] Die App bleibt bei Verbindungsverlust bedienbar (App-Shell wird aus Cache geladen)
- [ ] Bestellungen können offline aufgenommen und lokal zwischengespeichert werden
- [ ] Bei Wiederherstellung der Verbindung werden lokale Bestellungen automatisch synchronisiert
- [ ] Der Produktkatalog ist lokal gecacht und offline verfügbar
- [ ] Noch nicht abgesendete Bestellungen überleben einen App-Neustart (lokale Persistierung)
- [ ] Der Benutzer wird sichtbar über den Offline-Zustand informiert

---

## 5 · Reporting und Auswertung

> **Kassensitzung als Abrechnungszeitraum:** Alle zeitraumbezogenen Auswertungen beziehen sich auf eine Kassensitzung (`kassensitzung_nr`). Der Admin wählt die Kassensitzung aus, standardmäßig die aktuelle (offene).

### Umgesetzt

| ID   | Titel                       | Beschreibung                                                                           |
| ---- | --------------------------- | -------------------------------------------------------------------------------------- |
| R-01 | Tagesabrechnung             | KPIs, Breakdown-Sektionen, Stornierungsdetails für eine wählbare Kassensitzung.        |
| R-03 | Abrechnung pro Tisch        | Umsatzübersicht pro Tisch als Bestandteil von R-01 (`UmsatzProTisch[]`).               |
| R-04 | Abrechnung pro Servicekraft | Umsatzübersicht pro Servicekraft als Bestandteil von R-01 (`UmsatzProServicekraft[]`). |
| R-06 | Eigene Übersicht            | KPI-Sektion auf dem Service-Dashboard: eigene Bestellungen und kassierte Zahlungen.    |

### Offen

---

#### R-05 · Produktumsatz-Reporting

> **Rolle:** Admin · **Prio:** Nice-to-have

Der Admin kann Auswertungen über Produktumsätze in der gewählten Kassensitzung einsehen: verkaufte Mengen pro Produkt und Variante, ein Ranking der meistverkauften Varianten sowie Gesamteinnahmen pro Produkt.

**Akzeptanzkriterien:**

- [ ] Übersicht über verkaufte Mengen pro Produkt und Variante in der Kassensitzung
- [ ] Ranking der meistverkauften Varianten
- [ ] Gesamteinnahmen pro Produkt/Variante
- [ ] Nur durch Admin einsehbar

---

#### R-02 · Datenexport CSV

> **Rolle:** Admin · **Prio:** Nice-to-have

Der Admin kann Umsätze, Bestellungen und Artikeldaten als CSV exportieren, um sie extern weiterverarbeiten zu können (z. B. für die Vereinsbuchhaltung). Der Export bezieht sich auf die gewählte Kassensitzung.

**Akzeptanzkriterien:**

- [ ] Export von Umsätzen, Bestellungen und Artikeldaten als CSV
- [ ] Export bezieht sich auf die gewählte Kassensitzung (`kassensitzung_nr`)
- [ ] Export jederzeit durch den Admin auslösbar

---

## 6 · Fiskalkonformität

jotti unterliegt als elektronisches Aufzeichnungssystem der KassenSichV-Pflicht nach § 146a AO. Die folgenden Anforderungen sind verbindlich. Rechtliche Grundlagen und Compliance-Entscheidungen: [compliance.md](compliance.md).

| ID   | Titel                    | Status                     | Prio          |
| ---- | ------------------------ | -------------------------- | ------------- |
| F-01 | Seriennummer             | ✅                         | Must          |
| F-07 | Steuersätze              | ✅                         | Must          |
| F-03 | Belegausgabepflicht      | ✅                         | Must          |
| F-05 | ELSTER-Meldung           | ✅ Doku                    | Must (Doku)   |
| F-06 | Abrechnungskreis         | ✅ Pro Tisch/Kassensitzung | Should        |
| F-11 | Verfahrensdokumentation  | ✅ Doku                    | Should (Doku) |
| F-02 | TSE-Integration          | ✅                         | Should        |
| F-04 | DSFinV-K Export          | ✅                         | Should        |
| F-09 | eBeleg                   | 🔲                         | Nice          |
| F-10 | 10-Jahres-Archivierung   | ✅ Doku                    | Should (Doku) |
| F-08 | GoBD-Integritätsnachweis | 🔲                         | Nice          |
| F-12 | ELSTER-Meldung (API)     | 🔲                         | Nice          |

**Legende:** ✅ Umgesetzt · 🔲 Offen. **Phasen:** 0 = Baseline · 1 = Compliance-Grundlage · 2 = TSE-Integration · 3 = Erweiterungen

---

#### F-09 · eBeleg (Digitaler Beleg)

> **Prio:** Nice-to-have

Als papierloser Ersatz für den Bondruck kann dem Gast ein digitaler Beleg per QR-Code angeboten werden. Der Beleg wird als PDF oder HTML zum Download bereitgestellt.

**Akzeptanzkriterien:**

- [ ] Digitaler Beleg als Download-Link (PDF oder HTML)
- [ ] QR-Code mit Download-URL wird generiert
- [ ] QR-Code-Anzeige nach Zahlung im Frontend

---

#### F-08 · GoBD-Integritätsnachweis

> **Prio:** Nice-to-have

Die Unveränderbarkeit nach GoBD ist bereits erfüllt: Das Kassenjournal ist append-only (DB-Trigger plus REVOKE, K-07) und alle Geschäftsvorfälle tragen eine TSE-Signatur (F-02). Beides sind nach GoBD Rn. 110 zulässige softwareseitige Mittel zur Sicherung der Unveränderbarkeit (Sperren, Festschreibung, Versionierung). Eine zusätzliche kryptografische Hash-Chain ist dafür nicht erforderlich und ohne externen Anker auch nicht fälschungssicher: Wer die Append-only-Sperre aushebelt, kann eine selbst berechnete Kette neu berechnen. Den externen Anker liefert bereits die TSE.

F-08 ergänzt stattdessen einen read-only Selbsttest, der die bestehenden Garantien nachweisbar macht. Er nutzt ausschließlich vorhandene Daten und ändert das Schema nicht.

**Akzeptanzkriterien:**

- [ ] `/admin/integrity/check` prüft je `subject` die lückenlose Versionsfolge (keine Lücken oder Dopplungen in der OCC-Sequenz)
- [ ] Prüft, dass jeder signierpflichtige Event entweder eine TSE-Signatur trägt oder als TSE-Ausfall markiert und in der Nachsignier-Outbox vorhanden ist (keine still-unsignierten Geschäftsvorfälle)
- [ ] Liefert einen kompakten Report (geprüfte Subjects, gefundene Abweichungen) für Admin und Betriebsprüfer
- [ ] Kein Schema-Eingriff, keine neue Spalte, keine Hash-Chain

---

#### F-12 · Programmatische ELSTER-Meldung

> **Prio:** Nice-to-have

Statt der manuellen Meldung im Mein-ELSTER-Portal (F-05) meldet jotti die Kasse programmatisch beim Finanzamt an, entweder direkt über die ERiC-Bibliothek oder über die fiskaly-API, die die Meldung als Service kapselt. Für nicht-technische Vereinsbetreiber entfällt damit der manuelle Schritt im ELSTER-Portal.

**Akzeptanzkriterien:**

- [ ] Programmatische Meldung über ERiC oder fiskaly ist konfigurierbar
- [ ] Meldedaten (Seriennummer, TSE-Daten, Betreiber-Stammdaten) stammen aus den vorhandenen jotti-Stammdaten
- [ ] Übermittlungsbestätigung und Meldestatus werden im Admin-Bereich angezeigt
- [ ] Fehlgeschlagene Übermittlung ist erkennbar und wiederholbar

---
