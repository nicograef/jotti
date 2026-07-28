# Anforderungen: jotti

Übersicht und Roadmap aller Anforderungen an jotti: was umgesetzt ist und was geplant ist. Die fachliche Spezifikation einzelner Features entsteht bei Bedarf vor der Umsetzung als PRD (`docs/prds/`).

Weiterführend: Recht und Fiskal in [compliance.md](compliance.md), Architektur und Domäne in [handbuch.md](handbuch.md), Begriffe in [language.md](language.md), Betrieb in [Leitfaden](leitfaden/was-ist-jotti.md).

Aufbau: Die Roadmap listet die offenen Anforderungen, der Funktionsumfang die umgesetzten je Domäne. Jede Anforderung steht an genau einer Stelle und wandert beim Umsetzen aus der Roadmap in ihre Domäne.

## Konventionen

Priorität (MoSCoW): Must, Should, Nice.
IDs sind stabil und werden nicht wiederverwendet; Lücken in der Nummerierung sind normal.
Rollen: `admin` (Stammdaten und Kasse), `serviceleitung` (Kasse inkl. Storno), `service` (Kasse ohne Storno). Vollständige Matrix in [handbuch.md](handbuch.md) §5.1.

## Roadmap

Aktuell keine offenen Roadmap-Anforderungen — alle spezifizierten Anforderungen sind umgesetzt (siehe Funktionsumfang).

## Nicht-Ziele

Bewusst nicht geplant. Zurückgezogene IDs werden nicht wiederverwendet (siehe Konventionen).

| Ex-ID | Titel                                     | Begründung                                                                                                                                                                                                                                                                              |
| ----- | ----------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| K-03  | Ausgabe bestätigen                        | Nach dem ersten Praxiseinsatz ersatzlos entfernt: Der Ausgabe-Status war rein informativ, keine andere Funktion hing davon ab, und die Ausgabe koordinieren die Teams ohnehin über Arbeitsbons (K-12). Siehe [ADR 01](adrs/01_ausgabe-bestaetigen.md).                                   |
| K-13  | Küchendisplay (KDS)                       | Von der Roadmap gestrichen: baut auf dem im Praxistest verworfenen Ausgabe-Tracking (K-03) auf. Siehe [ADR 01](adrs/01_ausgabe-bestaetigen.md).                                                                                                                                          |
| K-15  | Zubereitungsstatus                        | Von der Roadmap gestrichen: baut auf K-13 und demselben verworfenen Ausgabe-Tracking auf. Siehe [ADR 01](adrs/01_ausgabe-bestaetigen.md).                                                                                                                                                |
| F-12  | Automatisierte ELSTER-Meldung (ERiC/API)  | Die Kassenmeldung nach § 146a Abs. 4 AO fällt pro Instanz nur einmal an (Inbetriebnahme, Außerbetriebnahme). Die manuelle Meldung über das ELSTER-Portal (F-05) deckt sie vollständig ab; eine native ERiC-C-Library oder die fiskaly-Submission-API lohnt für einen einmaligen Vorgang nicht. Siehe [compliance.md §7](compliance.md#7-elektronische-meldepflicht-elster). |
| R-03  | Abrechnung pro Tisch                      | Nach dem Praxistest- und UX-Review-Feedback ersatzlos entfernt: Der kassierte Umsatz je Tisch beantwortet keine Frage des Kassenwarts, und die offenen Salden deckt „Offene Tische" bereits ab. Siehe [ADR 02](adrs/02_umsatz-pro-tisch-entfernen.md). |

## Funktionsumfang

### Kasse (Core Domain)

| ID   | Titel                   | Beschreibung                                                                      |
| ---- | ----------------------- | --------------------------------------------------------------------------------- |
| K-01 | Bestellung aufnehmen    | Tisch wählen, aus dem Produktkatalog eine Bestellung zusammenstellen und abgeben. |
| K-02 | Zahlung registrieren    | Barzahlung mit Positionsauswahl (Teilzahlung), reduziert den Tischsaldo.          |
| K-04 | Stornierung erteilen    | Serviceleitung/Admin storniert Positionen; bezahlte als kassenwirksame Warenrücknahme, unbezahlte geldneutral; Saldo bleibt ≥ 0. |
| K-06 | Tischübersicht          | Dashboard mit Favoriten, Alle-Tische-Drawer und Tisch-Detail.                     |
| K-07 | Kassenjournal           | Append-only Event-Tabelle als Single Source of Truth.                             |
| K-09 | Bestellungen umbuchen   | Unbezahlte Bestellungen atomar zwischen Tischen umbuchen.                         |
| K-10 | Rückgeldberechnung      | Rückgeld und Trinkgeld clientseitig beim Kassieren.                               |
| K-11 | Tisch-Schnellsuche      | Echtzeit-Filterung nach Tischname im Drawer.                                      |
| K-12 | Arbeitsbon              | Automatischer Bon ohne Preise an Druckstationen (nicht-fiskalisch).               |
| K-14 | Tisch-Favoriten         | Serverseitige Favoriten pro Benutzer, Stern-Toggle.                               |
| K-16 | Kassensitzung eröffnen  | Global nummerierter Betriebstag; Sperre ohne offene Sitzung.                      |
| K-17 | Anfangsbestand setzen   | Wechselgeld als Basis; wird beim Eröffnen (K-16) gesetzt, kein eigener Schritt.   |
| K-18 | Kassenbestand einsehen  | Soll-Bestand als Aggregation über das Kassenjournal.                              |
| K-19 | Geldtransit buchen      | Einlage oder Entnahme als Journal-Event.                                          |
| K-20 | Betreiber-Stammdaten    | Vereinsdaten für Beleg (F-03) und DSFinV-K-Export (F-04).                         |
| K-21 | Kassensturz durchführen | Gezählter Ist- gegen Soll-Bestand, Differenz wird gebucht; Teil von K-22.         |
| K-22 | Kassenabschluss / Z-Bon | Kassensturz (K-21) und Tagesabschluss in einem Schritt; alle Tische auf Saldo 0.  |
| K-24 | Direktverkauf           | Bestellen, zahlen und ausgeben in einem Schritt, ohne Tisch; mit Historie/Storno. |
| K-25 | Druckstationen          | Konfiguration der Ausgabestationen; Zuordnung von Produktkategorien.              |
| K-26 | Druckauftrag-Verwaltung | Druckaufträge per Relay abrufen; fehlgeschlagene erneut versuchen/verwerfen.      |

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
| R-01 | Tagesabrechnung              | KPIs und Breakdowns je Kassensitzung.                                                                        |
| R-04 | Abrechnung pro Servicekraft  | Bargeld-Abrechnung des Tischservice je Servicekraft (Teil von R-01 und R-07): Kassiert − Rücknahmen = Abzugeben, dazu ein kombinierter Storno-Zähler. Ein Storno zählt bei der Servicekraft, deren Vorgang er rückgängig macht — Kassierer der zurückgenommenen Zahlung, Besteller der korrigierten Positionen —, nicht beim Akteur. Direktverkäufe bleiben außen vor (eigene Kasse). |
| R-05 | Produkt-/Varianten-Statistik | Ausgegebene Menge und Umsatz pro Produkt und Variante je Kassensitzung — in Tagesabrechnung und Live-Dashboard. |
| R-06 | Eigene Übersicht             | KPI-Sektion auf dem Service-Dashboard; bei zugeordneter Warenrücknahme zusätzlich ein Hinweis mit dem abzugebenden Betrag. |
| R-07 | Live-Dashboard               | Echtzeit-KPIs der offenen Kassensitzung.                                                                     |

### Querschnitt (Qualitätsmerkmale)

| ID   | Titel             | Beschreibung                                                |
| ---- | ----------------- | ----------------------------------------------------------- |
| Q-01 | Mobile-first      | Bedienbar ab 360 px, touch-optimiert.                       |
| Q-02 | Mehrbenutzerfähig | Parallele Zugriffe, Optimistic Concurrency Control.         |
| Q-03 | Validierung       | Zod (Frontend) und zog (Backend), deutsche Fehlermeldungen. |
| Q-04 | Datenintegrität   | Transaktionssicher, append-only, Cent-Werte, Soft-Deletes.  |
| Q-06 | HTTPS / TLS       | Caddy mit Let's Encrypt, lokal/LAN und in Produktion.        |
| Q-07 | Rate Limiting     | Login-Endpunkt geschützt (HTTP 429).                        |
| Q-08 | Security Headers  | CSP, X-Content-Type-Options, X-Frame-Options, HSTS.         |

### Fiskalkonformität

jotti unterliegt als elektronisches Aufzeichnungssystem der KassenSichV (§146a AO). Rechtliche Grundlagen und Compliance-Entscheidungen in [compliance.md](compliance.md).

| ID   | Titel                   | Beschreibung                                                      |
| ---- | ----------------------- | ----------------------------------------------------------------- |
| F-01 | Seriennummer            | Eindeutige Kassen- und Client-ID je Aufzeichnung.                 |
| F-02 | TSE-Integration         | Signatur jedes Geschäftsvorfalls (fiskaly Cloud-TSE).             |
| F-03 | Belegausgabepflicht     | Bondruck nach §146a AO.                                           |
| F-04 | DSFinV-K Export         | Prüfdatensatz im DSFinV-K-Format.                                 |
| F-05 | ELSTER-Meldung          | Manuelle Kassenmeldung im Mein-ELSTER-Portal (per Dokumentation). |
| F-06 | Abrechnungskreis        | Pro Tisch und Kassensitzung.                                      |
| F-07 | Steuersätze             | Korrekte USt-Sätze je Position.                                   |
| F-10 | 10-Jahres-Archivierung  | Aufbewahrungskonzept (per Dokumentation).                         |
| F-11 | Verfahrensdokumentation | Dokumentierte Kassenführung (per Dokumentation).                  |
| F-13 | TSE-Inbetriebnahme      | Geführte Ersteinrichtung der TSE: Konfiguration, Test, Status.    |
| F-14 | TSE-Ausfallsicherheit   | Nachsignierung bei TSE-Ausfall und Ausfalldokumentation.          |
