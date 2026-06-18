# Anforderungen: jotti

Übersicht und Roadmap aller Anforderungen an jotti: was umgesetzt ist und was geplant ist. Die fachliche Spezifikation einzelner Features entsteht bei Bedarf vor der Umsetzung als PRD (`docs/prds/`).

Weiterführend: Recht und Fiskal in [compliance.md](compliance.md), Architektur und Domäne in [handbuch.md](handbuch.md), Begriffe in [language.md](language.md), Betrieb in [Leitfaden](leitfaden/was-ist-jotti.md).

Aufbau: Die Roadmap listet die offenen Anforderungen, der Funktionsumfang die umgesetzten je Domäne. Jede Anforderung steht an genau einer Stelle und wandert beim Umsetzen aus der Roadmap in ihre Domäne.

## Konventionen

Priorität (MoSCoW): Must, Should, Nice.
IDs sind stabil und werden nicht wiederverwendet; Lücken in der Nummerierung sind normal.
Rollen: `admin` (Stammdaten und Kasse), `serviceleitung` (Kasse inkl. Storno und Auszahlung), `service` (Kasse ohne Storno). Vollständige Matrix in [handbuch.md](handbuch.md) §5.1.

## Roadmap

| ID   | Titel                    | Beschreibung                                                                  | Bereich     | Prio   |
| ---- | ------------------------ | ----------------------------------------------------------------------------- | ----------- | ------ |
| K-13 | Küchendisplay (KDS)      | Passive Echtzeit-Anzeige offener Bestellungen je Ausgabestation.              | Kasse       | Should |
| K-15 | Zubereitungsstatus       | Positionen als „in Zubereitung" / „fertig" markieren (baut auf K-13).         | Kasse       | Nice   |
| K-23 | Manuelle Tischfreigabe   | Neue Tisch-Session mit Suffix bei Saldo 0.                                    | Kasse       | Nice   |
| Q-05 | Offline-Fähigkeit        | Bestellaufnahme offline, lokale Zwischenspeicherung und Auto-Sync.            | Querschnitt | Nice   |
| R-02 | Datenexport CSV          | Umsätze, Bestellungen und Artikel als CSV je Kassensitzung.                   | Reporting   | Nice   |
| R-05 | Produktumsatz-Reporting  | Mengen, Ranking und Einnahmen pro Produkt und Variante.                       | Reporting   | Nice   |
| F-08 | GoBD-Integritätsnachweis | Read-only Selbsttest (Versionsfolge und Signaturpflicht) ohne Schemaänderung. | Fiskal      | Nice   |
| F-09 | eBeleg                   | Digitaler Beleg per QR-Code (PDF oder HTML).                                  | Fiskal      | Nice   |
| F-12 | ELSTER-Meldung (API)     | Programmatische Meldung via ERiC oder fiskaly.                                | Fiskal      | Nice   |

## Funktionsumfang

### Kasse (Core Domain)

| ID   | Titel                   | Beschreibung                                                                      |
| ---- | ----------------------- | --------------------------------------------------------------------------------- |
| K-01 | Bestellung aufnehmen    | Tisch wählen, aus dem Produktkatalog eine Bestellung zusammenstellen und abgeben. |
| K-02 | Zahlung registrieren    | Barzahlung mit Positionsauswahl (Teilzahlung), reduziert den Tischsaldo.          |
| K-03 | Ausgabe bestätigen      | Positionen als ausgegeben markieren, offene nachverfolgen.                        |
| K-04 | Stornierung erteilen    | Serviceleitung/Admin storniert Positionen; Saldo kann negativ werden.             |
| K-05 | Auszahlung leisten      | Negativen Tischsaldo positionsunabhängig ausgleichen.                             |
| K-06 | Tischübersicht          | Dashboard mit Favoriten, Alle-Tische-Drawer und Tisch-Detail.                     |
| K-07 | Kassenjournal           | Append-only Event-Tabelle als Single Source of Truth.                             |
| K-09 | Bestellungen umbuchen   | Unbezahlte Bestellungen atomar zwischen Tischen umbuchen.                         |
| K-10 | Rückgeldberechnung      | Rückgeld und Trinkgeld clientseitig beim Kassieren.                               |
| K-11 | Tisch-Schnellsuche      | Echtzeit-Filterung nach Tischname im Drawer.                                      |
| K-12 | Arbeitsbon              | Automatischer Bon ohne Preise an Druckstationen (nicht-fiskalisch).               |
| K-14 | Tisch-Favoriten         | Serverseitige Favoriten pro Benutzer, Stern-Toggle.                               |
| K-16 | Kassensitzung eröffnen  | Global nummerierter Betriebstag; Sperre ohne offene Sitzung.                      |
| K-17 | Anfangsbestand setzen   | Wechselgeld als Basis, einmal pro Kassensitzung.                                  |
| K-18 | Kassenbestand einsehen  | Soll-Bestand als Aggregation über das Kassenjournal.                              |
| K-19 | Geldtransit buchen      | Einlage oder Entnahme als Journal-Event.                                          |
| K-20 | Betreiber-Stammdaten    | Vereinsdaten für Beleg (F-03) und DSFinV-K-Export (F-04).                         |
| K-21 | Kassensturz durchführen | Gezählter Ist- gegen Soll-Bestand, Differenz wird gebucht.                        |
| K-22 | Tagesabschluss / Z-Bon  | Schließt die Kassensitzung (Voraussetzung: Kassensturz, alle Tische auf 0).       |
| K-24 | Direktverkauf           | Bestellen, zahlen und ausgeben in einem Schritt, ohne Tisch.                      |

### Stammdaten (Supporting Domain)

| ID   | Titel              | Beschreibung                                                       |
| ---- | ------------------ | ------------------------------------------------------------------ |
| S-01 | Produktverwaltung  | Katalog mit Kategorien und Varianten, Soft-Delete, Preise in Cent. |
| S-02 | Tischverwaltung    | Tische mit Name und Status, Soft-Delete.                           |
| S-03 | Benutzerverwaltung | Konten mit Rollen, Passwort-Reset per Einmalpasswort.              |

### Auth (Infrastruktur)

| ID   | Titel           | Beschreibung                                                |
| ---- | --------------- | ----------------------------------------------------------- |
| A-01 | Login           | Benutzername und Passwort gegen JWT (12 h, Argon2id).       |
| A-02 | Passwort setzen | Einmalpasswort führt zu „Passwort setzen" (min. 6 Zeichen). |
| A-03 | Logout          | JWT verwerfen, zurück zum Login.                            |

### Reporting

Zeitraumbezogene Auswertungen beziehen sich je auf eine Kassensitzung (`kassensitzung_nr`); Standard ist die aktuelle.

| ID   | Titel                       | Beschreibung                             |
| ---- | --------------------------- | ---------------------------------------- |
| R-01 | Tagesabrechnung             | KPIs und Breakdowns je Kassensitzung.    |
| R-03 | Abrechnung pro Tisch        | Umsatz pro Tisch (Teil von R-01).        |
| R-04 | Abrechnung pro Servicekraft | Umsatz pro Servicekraft (Teil von R-01). |
| R-06 | Eigene Übersicht            | KPI-Sektion auf dem Service-Dashboard.   |

### Querschnitt (Qualitätsmerkmale)

| ID   | Titel             | Beschreibung                                                |
| ---- | ----------------- | ----------------------------------------------------------- |
| Q-01 | Mobile-first      | Bedienbar ab 360 px, touch-optimiert.                       |
| Q-02 | Mehrbenutzerfähig | Parallele Zugriffe, Optimistic Concurrency Control.         |
| Q-03 | Validierung       | Zod (Frontend) und zog (Backend), deutsche Fehlermeldungen. |
| Q-04 | Datenintegrität   | Transaktionssicher, append-only, Cent-Werte, Soft-Deletes.  |
| Q-06 | HTTPS / TLS       | Caddy lokal, nginx und Let's Encrypt in Produktion.         |
| Q-07 | Rate Limiting     | Login-Endpunkt geschützt (HTTP 429).                        |
| Q-08 | Security Headers  | CSP, X-Content-Type-Options, X-Frame-Options, HSTS.         |

### Fiskalkonformität

jotti unterliegt als elektronisches Aufzeichnungssystem der KassenSichV (§146a AO). Rechtliche Grundlagen und Compliance-Entscheidungen in [compliance.md](compliance.md).

| ID   | Titel                   | Beschreibung                                                      |
| ---- | ----------------------- | ----------------------------------------------------------------- |
| F-01 | Seriennummer            | Eindeutige Kassen- und Client-ID je Aufzeichnung.                 |
| F-02 | TSE-Integration         | Signatur jedes Geschäftsvorfalls (fiskaly Cloud-TSE).             |
| F-03 | Belegausgabepflicht     | Bondruck oder eBeleg nach §146a AO.                               |
| F-04 | DSFinV-K Export         | Prüfdatensatz im DSFinV-K-Format.                                 |
| F-05 | ELSTER-Meldung          | Manuelle Kassenmeldung im Mein-ELSTER-Portal (per Dokumentation). |
| F-06 | Abrechnungskreis        | Pro Tisch und Kassensitzung.                                      |
| F-07 | Steuersätze             | Korrekte USt-Sätze je Position.                                   |
| F-10 | 10-Jahres-Archivierung  | Aufbewahrungskonzept (per Dokumentation).                         |
| F-11 | Verfahrensdokumentation | Dokumentierte Kassenführung (per Dokumentation).                  |
