# Roadmap — jotti

> Geordnete Aufgabenliste für die Fertigstellung von jotti.
> Jeder Task ist ein abgeschlossener, einzeln abarbeitbarer Schritt.
> Referenz: [anforderungen.md](anforderungen.md), [handbuch.md](handbuch.md), [compliance.md](compliance.md)

---

## Abgeschlossen

| Ehem. § | Titel                       | Referenz | Anmerkung                                                                                                      |
| ------- | --------------------------- | -------- | -------------------------------------------------------------------------------------------------------------- |
| 1       | Codebase-Bereinigung        | —        | Frontend UL-Rename (TS-Typen, Zod-Schemas, Routen), `language.md` aktualisiert                                 |
| 6       | Kasse — DB-Schema           | —        | `events`→`kassenjournal`, `table_state`→`tisch_sessions`, `kassensitzungen` CRUD-Entität, `kassenjournal_repo` |
| 7       | Kassensitzung eröffnen      | K-16     | Domain-Modell `domain/kasse/`, Kassensitzung-Sperre (HTTP 409), Admin-UI, Service-Hinweis                      |
| 8       | Anfangsbestand setzen       | K-17     | `anfangsbestand-gesetzt:v1`-Event, Admin-UI, Anfangsbestand-Invariante                                         |
| 9       | Kassenbewegungen            | K-19     | `kassenbewegung-gebucht:v1`-Event mit `art`-Feld, Admin-UI                                                     |
| 10      | Kassenbestand               | K-18     | SQL-Aggregation über Kassenjournal, Aufschlüsselung nach Komponenten                                           |
| 11      | Kassensturz                 | K-21     | Zwei-Event-Muster (`kassensturz-durchgefuehrt:v1` + `differenz-soll-ist-gebucht:v1`), Admin-UI                 |
| 12      | Tagesabschluss / Z-Bon      | K-22     | `tagesabschluss-erstellt:v1`-Event, Tisch-Saldo-Sperre, Kassensturz-Reihenfolge-Invariante                     |
| 13      | Admin-Dashboard: Kasse      | —        | Kassensitzung-Übersicht, Navigation, Routing                                                                   |
| —       | Tisch-Favoriten             | K-14     | DB `tisch_favoriten`, `favorit_repo`, API-Endpunkte, Frontend Star-Toggle                                      |
| —       | Abrechnung pro Tisch        | R-03     | Als Bestandteil von R-01/`GetAbrechnung` (`UmsatzProTisch[]`)                                                  |
| —       | Abrechnung pro Servicekraft | R-04     | Als Bestandteil von R-01/`GetAbrechnung` (`UmsatzProServicekraft[]`)                                           |
| —       | Eigene Übersicht            | R-06     | `EigeneUebersicht` Domain-Modell, `/service/get-eigene-uebersicht`                                             |

---

## 1 · Seriennummer (F-01)

- [ ] DB: `system_config`-Tabelle anlegen
- [ ] Backend: UUID-v4 beim ersten Start generieren und persistieren, API-Endpunkt `/admin/kasse/seriennummer`
- [ ] Frontend: Seriennummer im Admin-Dashboard anzeigen (mit Kopierbutton)

---

## 2 · Steuersätze (F-07)

- [ ] DB: Spalte `steuersatz` auf `produkt_varianten` ergänzen (`standard` / `ermaessigt` / `befreit`)
- [ ] Backend: Steuersatz-Enum, Validierung, Produkt-CRUD anpassen
- [ ] Frontend: Steuersatz-Auswahl bei Produktanlage und -bearbeitung
- [ ] Backend: Steuersatz in `BestellungAufgenommen`-Fat-Event einfrieren
- [ ] Backend: Tagesabrechnung nach Steuersatz aufschlüsseln
- [ ] Frontend: Steuersatz-Aufschlüsselung im Reporting anzeigen

---

## 3 · Betreiber-Stammdaten (K-20)

- [ ] DB: `betreiber_stammdaten`-Tabelle anlegen (Name, Adresse, Steuernummer)
- [ ] Backend: CRUD — Repository, Service, Handler
- [ ] Frontend: Admin-Seite Betreiber-Stammdaten pflegen

---

## 4 · Belegausgabe — Pflichtfelder (F-03)

- [ ] Backend: Beleg um Pflichtfelder erweitern (Steuersatz, Steuerbetrag pro Position, Seriennummer der Kasse, Betreiberadresse, Zahlungsart)
- [ ] ESC/POS-Bonformat aktualisieren
- [ ] Dokumentation: Bonformat-Beschreibung in `docs/bondruck.md` anpassen

---

## 5 · ELSTER-Dokumentation (F-05)

- [ ] `docs/betrieb/elster-meldung.md` — Schritt-für-Schritt-Anleitung für manuelle ELSTER-Meldung
- [ ] Frontend: Meldepflicht-Hinweis im Admin-Dashboard mit Link zur Anleitung und Seriennummer
- [ ] Frontend: Manuell setzbarer Meldestatus (`ausstehend` / `gemeldet am TT.MM.JJJJ`)

---

## 6 · Verfahrensdokumentation

- [ ] `docs/betrieb/verfahrensdokumentation.md` — Vorlage für Betriebsprüfer erstellen (Systemarchitektur, Datenbankschutz, Zugriffskontrollen)

---

## 7 · Dokumentation synchronisieren

- [x] `anforderungen.md` — Status aller Kasse- und Compliance-Features aktualisieren
- [x] `roadmap.md` — Komplett überarbeiten (Abgeschlossen-Tabelle + offene Items)
- [x] `language.md` — Abweichungstabelle aktualisieren
- [x] `compliance.md` — F-06 Status auf ✅ Umgesetzt
- [x] `handbuch.md` — EigeneUebersicht als Read-Model ergänzen
- [x] `diagrams.md` — Alle Diagramme gegen den Code prüfen und korrigieren
- [x] `AGENTS.md` und `README.md` — Gegen aktuellen Stand prüfen

---

## 8 · Reporting: Produktumsatz (R-05)

- [ ] Backend: Produktumsatz-Aggregation (Mengen pro Variante, Ranking, Einnahmen)
- [ ] Frontend: Produktumsatz-Ansicht im Admin-Reporting
- [ ] Tests

---

## 9 · Küchendisplay / KDS (K-13)

- [ ] Backend: Endpunkte für offene Bestellungen nach Kategorie (gruppiert nach Tisch)
- [ ] Frontend: KDS-Ansicht als eigene Route (Echtzeit-Polling oder SSE)
- [ ] Dokumentation: KDS-Setup-Anleitung
- [ ] Tests

---

## 10 · TSE-Integration — Interface & Adapter (F-02)

- [ ] Backend: `TSEClient` Go-Interface definieren (`StartTransaction`, `UpdateTransaction`, `FinishTransaction`)
- [ ] Backend: `FiskalyTSEClient` implementieren (REST-Adapter für fiskaly Cloud-TSE)
- [ ] Backend: Betriebsmodus `strict` / `bypass` konfigurierbar
- [ ] Backend: TSE-Mock für Unit- und Integrationstests
- [ ] Konfiguration: Umgebungsvariablen `TSE_PROVIDER`, `FISKALY_API_KEY`, `FISKALY_API_SECRET`

---

## 11 · TSE-Hooks auf Zahlungsfluss

- [ ] Backend: TSE-Transaktion bei `BestellungAufnehmen`
- [ ] Backend: TSE-Transaktion bei `ZahlungKassieren` und `StornierungErteilen`
- [ ] Backend: TSE-Transaktion bei Tagesabschluss und Kassenbewegungen
- [ ] Backend: Event-Daten um TSE-Felder erweitern (Signatur, Transaktionsnummer, Signaturzähler, TSE-Seriennummer)
- [ ] Tests

---

## 12 · TSE-Felder auf Beleg

- [ ] Beleg um TSE-Pflichtfelder erweitern (Transaktionsnummer, Signaturzähler, TSE-Seriennummer, Zeitpunkt)
- [ ] QR-Code auf Beleg generieren (DSFinV-K Anhang I)

---

## 13 · DSFinV-K-Export (F-04)

- [ ] Backend: `POST /admin/export/dsfinvk` — ZIP-Archiv mit allen Pflicht-CSV-Dateien
- [ ] `Stamm_Abschluss.csv`, `Stamm_Kassen.csv`, `Stamm_TSE.csv`, `Stamm_Orte.csv`
- [ ] `Bonkopf.csv`, `Bonkopf_USt.csv`, `Bonkopf_Zahlarten.csv`, `Bonpos.csv`, `Bonpos_USt.csv`
- [ ] `TSE_Transaktionen.csv`, `Z_GV_Typ.csv`, `Z_Zahlart.csv`, `index.xml`
- [ ] Frontend: Export-Button im Admin-Bereich
- [ ] Tests

---

## 14 · Dokumentation nach TSE-Integration

- [ ] `compliance.md` — Compliance-Status nach TSE-Integration aktualisieren
- [ ] `anforderungen.md` — F-02, F-04 Status aktualisieren
- [ ] `handbuch.md` — TSE-Architektur und DSFinV-K-Export dokumentieren

---

## 15 · Rückgeldberechnung (K-10)

- [ ] Frontend: Eingabefeld für erhaltenen Bargeldbetrag bei Zahlung, clientseitige Rückgeld-Anzeige

---

## 16 · Tisch-Schnellsuche (K-11)

- [ ] Frontend: Suchfeld im Alle-Tische-Drawer (clientseitige Filterung nach Name/Nummer)

---

## 17 · Datenexport CSV (R-02)

- [ ] Backend: `POST /admin/export/csv` — Umsätze, Bestellungen, Artikeldaten als CSV
- [ ] Frontend: Export-Button im Reporting (Kassensitzung wählbar)

---

## 18 · Bestellungen umbuchen (K-09)

- [ ] Backend: Umbuchungs-Command (atomare Stornierung am Quell-Tisch + Neubestellung am Ziel-Tisch)
- [ ] Frontend: Umbuchung-UI (Ziel-Tisch auswählen)
- [ ] Tests

---

## 19 · Ausgabestationen / Zubereitungsstatus (K-15)

- [ ] Backend: Zubereitungsstatus-Endpunkte (in Zubereitung → fertig)
- [ ] Frontend: Status-Verwaltung auf KDS-Ansicht
- [ ] Frontend: Zubereitungsstatus-Anzeige für Servicekräfte
- [ ] Tests

---

## 20 · Tisch-Session Phase 2 — Manuelle Tischfreigabe

> Für jottis Festzelt-Betrieb (Vereinsfest, Maihock) die korrektere Abbildung der Realität: mehrere Gästegruppen an einem Tisch erhalten jeweils einen eigenen Abrechnungskreis (Tisch-Session). Empfohlen zeitnah nach Abschluss der TSE-Integration umzusetzen.

- [ ] Backend: Domain-Aktion `TischFreigegeben` — neue Tisch-Session mit Suffix starten (z. B. `kassensitzung-1/tisch-42-b`)
- [ ] Frontend: UI-Aktion „Tisch freimachen" in der Servicekraft-Ansicht
- [ ] Tests

---

## 21 · Compliance Phase 3 — eBeleg

- [ ] Backend: Digitaler Beleg als Download-Link (PDF oder HTML)
- [ ] Backend: QR-Code mit Download-URL generieren
- [ ] Frontend: QR-Code-Anzeige nach Zahlung

---

## 22 · Compliance Phase 3 — Automatisierte ELSTER-Meldung

- [ ] Backend: Programmatische Meldung über ERiC-Schnittstelle oder fiskaly Submission API
- [ ] Frontend: ELSTER-Meldung direkt aus Admin-Dashboard auslösen

---

## 23 · GoBD Hash-Chain (F-08)

- [ ] Backend: SHA-256-Verkettung aller Events (`previous_hash` pro Event)
- [ ] Backend: Genesis-Hash pro Abrechnungskreis
- [ ] Backend: Integritätsprüfungs-Endpunkt `POST /admin/integrity/check`
- [ ] Tests

---

## 24 · Offline-Fähigkeit (Q-05)

- [ ] Frontend: Service Worker für App-Shell-Caching
- [ ] Frontend: Lokale Bestellungs-Speicherung bei Verbindungsverlust
- [ ] Frontend: Automatische Synchronisierung bei Reconnect
- [ ] Frontend: Offline-Indikator in der UI
