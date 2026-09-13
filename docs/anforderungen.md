# Anforderungen: jotti

IDs für alle umgesetzten Funktionen und für die bewusst ausgeschlossenen. Sie sind stabil und werden nicht wiederverwendet; Lücken in der Nummerierung sind normal.

## Nicht-Ziele

Bewusst nicht geplant.

| Ex-ID      | Titel                                      | Begründung                                                                                                                                                                                                                                                                                                                                                                  |
| ---------- | ------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| K-03       | Ausgabe bestätigen                         | Der Ausgabe-Status ist rein informativ, keine andere Funktion hängt davon ab; die Ausgabe koordinieren die Teams über Arbeitsbons (K-12). Siehe [D01](decisions.md).                                                                                                                                                                                                        |
| K-13, K-15 | Küchendisplay (KDS), Zubereitungsstatus    | Bauen auf dem im Praxistest verworfenen Ausgabe-Tracking (K-03) auf. Siehe [D01](decisions.md).                                                                                                                                                                                                                                                                             |
| F-12       | Automatisierte ELSTER-Meldung (ERiC/API)   | Die Kassenmeldung nach § 146a Abs. 4 AO fällt pro Instanz nur einmal an (Inbetriebnahme, Außerbetriebnahme). Die manuelle Meldung über das ELSTER-Portal (F-05) deckt sie vollständig ab; eine native ERiC-C-Library oder die fiskaly-Submission-API lohnt für einen einmaligen Vorgang nicht. Siehe [compliance.md §7](compliance.md#7-elektronische-meldepflicht-elster). |
| R-03       | Abrechnung pro Tisch                       | Der kassierte Umsatz je Tisch beantwortet keine Frage des Kassenwarts, und die offenen Salden deckt „Offene Tische" bereits ab. Siehe [D02](decisions.md).                                                                                                                                                                                                                  |
| —          | Vereinslogo auf dem Bon                    | Kosmetisch: Raster-Druck und Logo-Upload wären Aufwand ohne Kernnutzen; der Vereinsname steht bereits im Bonkopf.                                                                                                                                                                                                                                                           |
| —          | Elektronischer Beleg per E-Mail oder Link  | Das Gast-Handy ist nicht im Vereins-WLAN; Mailversand vom Vereins-Server und Adress-Erfassung am Tisch kämen hinzu, während der Kassenbeleg-Drucker den Bedarf bereits deckt.                                                                                                                                                                                               |
| —          | Helferdeckel                               | Ein Tisch pro Helfer deckt den Bedarf bereits ab.                                                                                                                                                                                                                                                                                                                           |
| —          | Autostart / Windows-Dienst für den Starter | Eintägige Feste mit langen Pausen zwischen den Einsätzen; der tägliche manuelle Start ist gewollt.                                                                                                                                                                                                                                                                          |

## Funktionsumfang

### Kasse (Core Domain)

| ID   | Titel                   | Beschreibung                                                                                                                            |
| ---- | ----------------------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| K-01 | Bestellung aufnehmen    | Tisch wählen, aus dem Produktkatalog eine Bestellung zusammenstellen und abgeben.                                                       |
| K-02 | Zahlung registrieren    | Barzahlung mit Positionsauswahl (Teilzahlung), reduziert den Tischsaldo.                                                                |
| K-04 | Stornierung erteilen    | Serviceleitung/Admin storniert Positionen; Aufteilung nach Bezahlstatus → [handbuch.md §3.7](handbuch.md#37-invarianten).               |
| K-06 | Tischübersicht          | Dashboard mit Favoriten, Alle-Tische-Drawer und Tisch-Detail.                                                                           |
| K-07 | Kassenjournal           | Append-only Event-Tabelle als Single Source of Truth.                                                                                   |
| K-09 | Bestellungen umbuchen   | Unbezahlte Bestellungen atomar zwischen Tischen umbuchen.                                                                               |
| K-10 | Rückgeldberechnung      | Rückgeld und Trinkgeld clientseitig beim Kassieren.                                                                                     |
| K-11 | Tisch-Schnellsuche      | Echtzeit-Filterung nach Tischname im Drawer.                                                                                            |
| K-12 | Arbeitsbon              | Automatischer Bon ohne Preise an Druckstationen (nicht-fiskalisch).                                                                     |
| K-14 | Tisch-Favoriten         | Serverseitige Favoriten pro Benutzer, Stern-Toggle.                                                                                     |
| K-16 | Kassensitzung eröffnen  | Global nummerierter Betriebstag; Sperre ohne offene Sitzung.                                                                            |
| K-17 | Anfangsbestand setzen   | Wechselgeld als Basis; wird beim Eröffnen (K-16) gesetzt, kein eigener Schritt.                                                         |
| K-18 | Kassenbestand einsehen  | Soll-Bestand als Aggregation über das Kassenjournal.                                                                                    |
| K-19 | Geldtransit buchen      | Einlage oder Entnahme als Journal-Event.                                                                                                |
| K-20 | Betreiber-Stammdaten    | Vereinsdaten für Beleg (F-03) und DSFinV-K-Export (F-04).                                                                               |
| K-21 | Kassensturz durchführen | Gezählter Ist- gegen Soll-Bestand, Differenz wird gebucht; Teil von K-22.                                                               |
| K-22 | Kassenabschluss / Z-Bon | Kassensturz (K-21) und Tagesabschluss in einem Schritt; alle Tische auf Saldo 0.                                                        |
| K-24 | Direktverkauf           | Bestellen, zahlen und ausgeben in einem Schritt, ohne Tisch; mit Historie/Storno.                                                       |
| K-25 | Druckstationen          | Konfiguration der Ausgabestationen; Zuordnung von Produktkategorien; Bonmodus pro Position, pro Bestellung, am Abholbon auch pro Stück. |
| K-26 | Druckauftrag-Verwaltung | Druckaufträge per Relay abrufen; fehlgeschlagene erneut versuchen/verwerfen.                                                            |

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

| ID   | Titel                        | Beschreibung                                                                                                                                                                                                    |
| ---- | ---------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| R-01 | Tagesabrechnung              | KPIs und Breakdowns je Kassensitzung.                                                                                                                                                                           |
| R-04 | Abrechnung pro Servicekraft  | Bargeld-Abrechnung des Tischservice je Servicekraft (Teil von R-01 und R-07); Direktverkäufe bleiben außen vor (eigene Kasse). Storno-Zuordnung → [handbuch.md §7.2](handbuch.md#72-admin-ansichten-reporting). |
| R-05 | Produkt-/Varianten-Statistik | Ausgegebene Menge und Umsatz pro Produkt und Variante je Kassensitzung — in Tagesabrechnung und Live-Dashboard.                                                                                                 |
| R-06 | Eigene Übersicht             | KPI-Sektion auf dem Service-Dashboard; bei zugeordneter Warenrücknahme zusätzlich ein Hinweis mit dem abzugebenden Betrag.                                                                                      |
| R-07 | Live-Dashboard               | Echtzeit-KPIs der offenen Kassensitzung.                                                                                                                                                                        |

### Querschnitt (Qualitätsmerkmale)

| ID   | Titel             | Beschreibung                                                |
| ---- | ----------------- | ----------------------------------------------------------- |
| Q-01 | Mobile-first      | Bedienbar ab 360 px, touch-optimiert.                       |
| Q-02 | Mehrbenutzerfähig | Parallele Zugriffe, Optimistic Concurrency Control.         |
| Q-03 | Validierung       | Zod (Frontend) und zog (Backend), deutsche Fehlermeldungen. |
| Q-04 | Datenintegrität   | Transaktionssicher, append-only, Cent-Werte, Soft-Deletes.  |
| Q-06 | HTTPS / TLS       | Caddy mit Let's Encrypt, lokal/LAN und in Produktion.       |
| Q-07 | Rate Limiting     | Login-Endpunkt geschützt (HTTP 429).                        |
| Q-08 | Security Headers  | CSP, X-Content-Type-Options, X-Frame-Options, HSTS.         |

### Fiskalkonformität

Rechtliche Grundlagen und Compliance-Entscheidungen: [compliance.md](compliance.md).

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
