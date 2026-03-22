# Roadmap — jotti

> Geordnete Aufgabenliste für die Fertigstellung von jotti.
> Jeder Task ist ein abgeschlossener, einzeln abarbeitbarer Schritt.
> Referenz: [anforderungen.md](anforderungen.md), [handbuch.md](handbuch.md), [compliance.md](compliance.md)

---

## 1 · Codebase-Bereinigung

- [x] Frontend: TS-Typen und Zod-Schemas auf Ubiquitous Language umbenennen (`Product`→`Produkt`, `Variant`→`Variante`, `Category`→`Kategorie`, `PriceCents`→`PreisCents`, `Comment`→`Kommentar`, `Quantity`→`Menge`)
- [x] Frontend: Routen auf Ubiquitous Language umgestellt (`/admin/products`→`/admin/produkte`, `/admin/tables`→`/admin/tische`, `/admin/users`→`/admin/benutzer`, `/admin/reporting`→`/admin/auswertung`, `/service/tables`→`/service/tische`)
- [x] `language.md` aktualisieren — Ist/Soll-Abweichungen als erledigt markieren, Routen-Konvention auf Deutsch aktualisiert

---

## 2 · Seriennummer (F-01)

- [ ] DB: `system_config`-Tabelle anlegen
- [ ] Backend: UUID-v4 beim ersten Start generieren und persistieren, API-Endpunkt `/admin/kasse/seriennummer`
- [ ] Frontend: Seriennummer im Admin-Dashboard anzeigen (mit Kopierbutton)

---

## 3 · Steuersätze (F-07)

- [ ] DB: Spalte `steuersatz` auf `produkt_varianten` ergänzen (`standard` / `ermaessigt` / `befreit`)
- [ ] Backend: Steuersatz-Enum, Validierung, Produkt-CRUD anpassen
- [ ] Frontend: Steuersatz-Auswahl bei Produktanlage und -bearbeitung
- [ ] Backend: Steuersatz in `BestellungAufgenommen`-Fat-Event einfrieren
- [ ] Backend: Tagesabrechnung nach Steuersatz aufschlüsseln
- [ ] Frontend: Steuersatz-Aufschlüsselung im Reporting anzeigen

---

## 4 · Betreiber-Stammdaten (K-20)

- [ ] DB: `betreiber_stammdaten`-Tabelle anlegen (Name, Adresse, Steuernummer)
- [ ] Backend: CRUD — Repository, Service, Handler
- [ ] Frontend: Admin-Seite Betreiber-Stammdaten pflegen

---

## 5 · Belegausgabe — Pflichtfelder (F-03)

- [ ] Backend: Beleg um Pflichtfelder erweitern (Steuersatz, Steuerbetrag pro Position, Seriennummer der Kasse, Betreiberadresse, Zahlungsart)
- [ ] ESC/POS-Bonformat aktualisieren
- [ ] Dokumentation: Bonformat-Beschreibung in `docs/bondruck.md` anpassen

---

## 6 · Kasse — DB-Schema (Kassenjournal + Projektionen)

- [ ] DB: `events` → `kassenjournal` in `01_initial.up.sql` umbenennen (+ `kassensitzung_nr`-Spalte, Trigger, Indizes)
- [ ] DB: `table_state` → `tisch_session_state` (session-scoped: PK = `subject`, mit `tisch_id`, `kassensitzung_nr`)
- [ ] DB: Projektionstabelle `kassensitzung_state` anlegen (PK: `subject`, `z_nr` UNIQUE, `datum`, `status`)
- [ ] DB: Alte Tabellen `abrechnungskreis`, `kassenbewegungen`, `kassensturz`, `z_bons` entfernen (falls geplant)
- [ ] Repository: `kassenjournal_repo/` (ersetzt `event_repo/`) mit explizitem `StreamType`-Routing
- [ ] `make sqlc` — Queries generieren

---

## 7 · Kassensitzung eröffnen (K-16)

- [ ] Backend: Domain-Modell `domain/kasse/` mit Kassensitzung-Events und -State
- [ ] Backend: Application Service, HTTP Handler für `KassensitzungEroeffnen`
- [ ] Backend: Kassensitzung-Sperre — schreibende Tischoperationen mit HTTP 409 ablehnen, wenn keine `offen`e Kassensitzung existiert
- [ ] Frontend: Admin-Seite — Kassensitzung mit Bezeichnung eröffnen (Tageseröffnung)
- [ ] Frontend: Hinweis „Kasse ist noch nicht geöffnet" für Servicekräfte bei fehlender offener Kassensitzung
- [ ] Tests

---

## 8 · Anfangsbestand setzen (K-17)

- [ ] Backend: Application Service, Handler (Betrag pro Kassensitzung, einmalig via `AnfangsbestandGesetzt`-Event)
- [ ] Frontend: Anfangsbestand-Eingabe im Admin
- [ ] Tests

---

## 9 · Kassenbewegungen (K-19)

- [ ] Backend: `KassenbewegungGebucht`-Event mit `art`-Feld (geldtransit | privatentnahme | privateinlage)
- [ ] Backend: Application Service, Handler für einheitliches `kassenbewegung-buchen`
- [ ] Frontend: Kassenbewegungen-UI im Admin (einheitliches Formular mit Bewegungsart-Auswahl)
- [ ] Tests

---

## 10 · Kassenbestand (K-18)

- [ ] Backend: Kassenbestand als SQL-Aggregation über das Kassenjournal (`WHERE kassensitzung_nr = $1`)
- [ ] Frontend: Kassenbestand-Anzeige mit Aufschlüsselung nach Komponenten
- [ ] Tests

---

## 11 · Kassensturz (K-21)

- [ ] Backend: Zwei-Event-Muster — `KassensturzDurchgefuehrt` + `DifferenzSollIstGebucht` (wenn Differenz ≠ 0) in einer Transaktion
- [ ] Frontend: Kassensturz-Dialog im Admin (Soll vs. Ist, Differenz-Anzeige)
- [ ] Tests

---

## 12 · Tagesabschluss / Z-Bon (K-22)

- [ ] Backend: `TagesabschlussErstellt`-Event, Kassensitzung abschließen (Status → `abgeschlossen`)
- [ ] Backend: Z-Bon als Aggregation über Kassenjournal (keine Stammdaten-Snapshot-Pflicht, Änderungssperre garantiert Konsistenz)
- [ ] Backend: Tisch-Saldo-Sperre prüfen (alle Tische müssen Saldo = 0 haben)
- [ ] Frontend: Tagesabschluss-Workflow im Admin (offene Tische anzeigen, Kassensturz-Voraussetzung prüfen, Bestätigung)
- [ ] Tests

---

## 13 · Admin-Dashboard: Kasse

- [ ] Frontend: Kasse-Übersichtsseite im Admin (aktive Kassensitzung, Kassenbestand, letzte Bewegungen, Z-Bon-Historie)
- [ ] Frontend: Navigation und Routing für alle Kasse-Funktionen

---

## 14 · ELSTER-Dokumentation (F-05)

- [ ] `docs/betrieb/elster-meldung.md` — Schritt-für-Schritt-Anleitung für manuelle ELSTER-Meldung
- [ ] Frontend: Meldepflicht-Hinweis im Admin-Dashboard mit Link zur Anleitung und Seriennummer
- [ ] Frontend: Manuell setzbarer Meldestatus (`ausstehend` / `gemeldet am TT.MM.JJJJ`)

---

## 15 · Verfahrensdokumentation

- [ ] `docs/betrieb/verfahrensdokumentation.md` — Vorlage für Betriebsprüfer erstellen (Systemarchitektur, Datenbankschutz, Zugriffskontrollen)

---

## 16 · Dokumentation nach Phase 1

- [ ] `anforderungen.md` — Status aller Kasse- und Compliance-Features aktualisieren
- [ ] `compliance.md` — Compliance-Status nach Phase 1 aktualisieren
- [ ] `handbuch.md` — Kasse-Code-Referenzen ergänzen

---

## 17 · Reporting: Abrechnung pro Tisch (R-03)

- [ ] Backend: Detaillierte Tisch-Abrechnung (alle Events chronologisch mit Saldo)
- [ ] Frontend: Tisch-Detail-Abrechnung im Admin-Reporting
- [ ] Tests

---

## 18 · Reporting: Produktumsatz (R-05)

- [ ] Backend: Produktumsatz-Aggregation (Mengen pro Variante, Ranking, Einnahmen)
- [ ] Frontend: Produktumsatz-Ansicht im Admin-Reporting
- [ ] Tests

---

## 19 · Küchendisplay / KDS (K-13)

- [ ] Backend: Endpunkte für offene Bestellungen nach Kategorie (gruppiert nach Tisch)
- [ ] Frontend: KDS-Ansicht als eigene Route (Echtzeit-Polling oder SSE)
- [ ] Dokumentation: KDS-Setup-Anleitung
- [ ] Tests

---

## 20 · TSE-Integration — Interface & Adapter (F-02)

- [ ] Backend: `TSEClient` Go-Interface definieren (`StartTransaction`, `UpdateTransaction`, `FinishTransaction`)
- [ ] Backend: `FiskalyTSEClient` implementieren (REST-Adapter für fiskaly Cloud-TSE)
- [ ] Backend: Betriebsmodus `strict` / `bypass` konfigurierbar
- [ ] Backend: TSE-Mock für Unit- und Integrationstests
- [ ] Konfiguration: Umgebungsvariablen `TSE_PROVIDER`, `FISKALY_API_KEY`, `FISKALY_API_SECRET`

---

## 21 · TSE-Hooks auf Zahlungsfluss

- [ ] Backend: TSE-Transaktion bei `BestellungAufnehmen`
- [ ] Backend: TSE-Transaktion bei `ZahlungKassieren` und `StornierungErteilen`
- [ ] Backend: TSE-Transaktion bei Tagesabschluss und Kassenbewegungen
- [ ] Backend: Event-Daten um TSE-Felder erweitern (Signatur, Transaktionsnummer, Signaturzähler, TSE-Seriennummer)
- [ ] Tests

---

## 22 · TSE-Felder auf Beleg

- [ ] Beleg um TSE-Pflichtfelder erweitern (Transaktionsnummer, Signaturzähler, TSE-Seriennummer, Zeitpunkt)
- [ ] QR-Code auf Beleg generieren (DSFinV-K Anhang I)

---

## 23 · DSFinV-K-Export (F-04)

- [ ] Backend: `POST /admin/export/dsfinvk` — ZIP-Archiv mit allen Pflicht-CSV-Dateien
- [ ] `Stamm_Abschluss.csv`, `Stamm_Kassen.csv`, `Stamm_TSE.csv`, `Stamm_Orte.csv`
- [ ] `Bonkopf.csv`, `Bonkopf_USt.csv`, `Bonkopf_Zahlarten.csv`, `Bonpos.csv`, `Bonpos_USt.csv`
- [ ] `TSE_Transaktionen.csv`, `Z_GV_Typ.csv`, `Z_Zahlart.csv`, `index.xml`
- [ ] Frontend: Export-Button im Admin-Bereich
- [ ] Tests

---

## 24 · Dokumentation nach Phase 2

- [ ] `compliance.md` — Compliance-Status nach Phase 2 aktualisieren
- [ ] `anforderungen.md` — F-02, F-04 Status aktualisieren
- [ ] `handbuch.md` — TSE-Architektur und DSFinV-K-Export dokumentieren

---

## 25 · Rückgeldberechnung (K-10)

- [ ] Frontend: Eingabefeld für erhaltenen Bargeldbetrag bei Zahlung, clientseitige Rückgeld-Anzeige

---

## 26 · Tisch-Schnellsuche (K-11)

- [ ] Frontend: Suchfeld im Alle-Tische-Drawer (clientseitige Filterung nach Name/Nummer)

---

## 27 · Datenexport CSV (R-02)

- [ ] Backend: `POST /admin/export/csv` — Umsätze, Bestellungen, Artikeldaten als CSV
- [ ] Frontend: Export-Button im Reporting (Kassensitzung wählbar)

---

## 28 · Bestellungen umbuchen (K-09)

- [ ] Backend: Umbuchungs-Command (atomare Stornierung am Quell-Tisch + Neubestellung am Ziel-Tisch)
- [ ] Frontend: Umbuchung-UI (Ziel-Tisch auswählen)
- [ ] Tests

---

## 29 · Ausgabestationen / Zubereitungsstatus (K-15)

- [ ] Backend: Zubereitungsstatus-Endpunkte (in Zubereitung → fertig)
- [ ] Frontend: Status-Verwaltung auf KDS-Ansicht
- [ ] Frontend: Zubereitungsstatus-Anzeige für Servicekräfte
- [ ] Tests

---

## 30 · Tisch-Session Phase 2 — Manuelle Tischfreigabe

> Für jottis Festzelt-Betrieb (Vereinsfest, Maihock) die korrektere Abbildung der Realität: mehrere Gästegruppen an einem Tisch erhalten jeweils einen eigenen Abrechnungskreis (Tisch-Session). Empfohlen zeitnah nach Abschluss der Phase-2-TSE-Integration umzusetzen.

- [ ] Backend: Domain-Aktion `TischFreigegeben` — neue Tisch-Session mit Suffix starten (z. B. `kassensitzung-20260322-tisch-42-b`)
- [ ] Frontend: UI-Aktion „Tisch freimachen" in der Servicekraft-Ansicht
- [ ] Tests

---

## 31 · Compliance Phase 3 — eBeleg

- [ ] Backend: Digitaler Beleg als Download-Link (PDF oder HTML)
- [ ] Backend: QR-Code mit Download-URL generieren
- [ ] Frontend: QR-Code-Anzeige nach Zahlung

---

## 32 · Compliance Phase 3 — Automatisierte ELSTER-Meldung

- [ ] Backend: Programmatische Meldung über ERiC-Schnittstelle oder fiskaly Submission API
- [ ] Frontend: ELSTER-Meldung direkt aus Admin-Dashboard auslösen

---

## 33 · GoBD Hash-Chain (F-08)

- [ ] Backend: SHA-256-Verkettung aller Events (`previous_hash` pro Event)
- [ ] Backend: Genesis-Hash pro Abrechnungskreis
- [ ] Backend: Integritätsprüfungs-Endpunkt `POST /admin/integrity/check`
- [ ] Tests

---

## 34 · Offline-Fähigkeit (Q-05)

- [ ] Frontend: Service Worker für App-Shell-Caching
- [ ] Frontend: Lokale Bestellungs-Speicherung bei Verbindungsverlust
- [ ] Frontend: Automatische Synchronisierung bei Reconnect
- [ ] Frontend: Offline-Indikator in der UI

---

## 35 · Dokumentation & Release

- [ ] `README.md` aktualisieren (Kasse, Compliance-Status, aktuelle Features)
- [ ] Website (`website/`) aktualisieren (Featureliste, Screenshots)
- [ ] `docs/hosting.md` vervollständigen (Setup-Anleitung, Backup, Updates)
- [ ] Changelog / Release Notes erstellen
- [ ] `anforderungen.md` — alle Status finalisieren
- [ ] `compliance.md` — Gesamtstatus finalisieren
- [ ] `produktbeschreibung.md` — Marketingtexte an finalen Funktionsumfang anpassen
