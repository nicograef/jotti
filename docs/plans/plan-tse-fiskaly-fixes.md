# Plan: TSE-/fiskaly-Integration — Fixes und Verbesserungen

> Source PRD: n/a — Quelle ist das Audit [docs/audits/2026-06-11-tse-fiskaly-audit.md](../audits/2026-06-11-tse-fiskaly-audit.md) (Findings I-01–I-15, D-01–D-07)

## Goal

Die TSE-Integration funktionsfähig (I-01/I-02: derzeit schlägt **jeder** Signier-Request mit `400 E_PARSER` fehl), DSFinV-K-/AEAO-konform (Formate, Vorgangsarten, Belegpflichtfelder) und betriebsrobust (Worker, Ausfalldokumentation) machen. Zusätzlich: TSS-/Client-Einrichtung aus jotti heraus automatisieren, Doku-Fehler korrigieren.

## Architectural decisions

Durable Entscheidungen für alle Phasen:

- **Base64-Grenze:** Base64-Encoding/-Decoding von `process_data` ist ausschließlich Sache des `FiskalyTSEClient`. Domain- und Application-Schicht arbeiten mit Klartext-processData.
- **Start-Request ohne Schema:** `StartTransaction` sendet nur `state: ACTIVE` + `client_id` (DSFinV-K: processType/processData bei Start immer leer).
- **processType-Konstanten zentral** in `backend/domain/tse`: `Kassenbeleg-V1`, `Bestellung-V1`, `SonstigerVorgang` (ohne `-V1`).
- **Vorgangsarten-Mapping:** Geldtransit, Kassendifferenz, Auszahlung → `Kassenbeleg-V1` (Eigenbelege, AEAO 2.2.3.6.1); Tagesabschluss → `SonstigerVorgang`; Bestellung → `Bestellung-V1`.
- **tx_id:** UUIDv4 (`uuid.New()`), einmal erzeugt beim Signierversuch und als `tseTxId` in den Event-Daten persistiert (Breaking Change erlaubt, kein Dual-Read). Bis zur Umstellung in Phase 6 bleibt die deterministische v5-Ableitung in Betrieb.
- **Worker-Statusmodell** analog Druckaufträge (handbuch.md §4.6): `offen → erledigt`; nach N Fehlversuchen `fehlgeschlagen`; von dort `verworfen` oder zurück auf `offen`. Spalten: `versuche`, `letzter_fehler`, `naechster_versuch_am` (exponentielles Backoff).
- **Ausfalldokumentation** (AEAO 1.14.1) wird aus den Nachsignier-Aufträgen abgeleitet (erstellt_am = Beginn, erledigt_am = Ende, letzter_fehler = Grund) und im Admin sichtbar gemacht — keine separate Tabelle.
- **Routes (POST-only):** `admin/tse-einrichten` (Setup-Wizard); `admin/get-tse-nachsignier-auftraege`, `admin/tse-nachsignier-auftrag-zuruecksetzen`, `admin/tse-nachsignier-auftrag-verwerfen` (analog Druckauftrags-Verwaltung).
- **Schema:** `tse_konfiguration` erhält `admin_puk`, `admin_pin` (für Setup-Automatisierung; Klartext wie `api_secret` — Self-hosted-Prämisse); `tse_nachsignier_auftraege` erhält die Worker-Spalten; Status-CHECK erweitert. Änderungen direkt in `database/migrations/01_initial.up.sql`.
- **fiskaly-Client-`serial_number` = jotti-Kassen-Seriennummer** (Kassenidentitäts-UUID). Der QR-Code trägt diese serial_number; Beleg und QR müssen übereinstimmen.

## Inventory

- `backend/repository/tse_repo/fiskaly_client.go:56-69, 140-206` — Upsert-Request mit `rawSchema` (Klartext, kein Base64); Start sendet Schema mit; `:150` txID nicht escaped
- `backend/repository/tse_repo/fiskaly_client_test.go:86-97` — Fake-Server ohne Base64-/Schema-Assertions
- `backend/api/table/application/tse_signing.go` — Signierung Bestellung/Zahlung/Storno/Auszahlung; `:370-398` falsches Bestellformat; `:302-322` TSEAusfall nur für Zahlung; `:409-475` v5-tx-IDs
- `backend/api/direktverkauf/application/tse_signing.go`, `backend/api/kasse/application/tse_signing.go` — Duplikate von `signEventWithTSE` & Co.; `kasse/...:20` `SonstigerVorgang-V1`
- `backend/api/table/application/kassenbeleg_command.go:68-73, 139-160, 191-218` — Direktverkauf-Zweig ohne TSE-Daten/Fallback/Ausfallvermerk
- `backend/domain/kasse/tisch_session.go:39-41` — `ErsteBestellungLogTime` nur aus TSEData
- `backend/app/tse_nachsignier_worker.go:18-21, 112-142` — 5-s-Poll, kein Backoff/Fehlerzähler; Poison-Pill bei 409
- `backend/sqlc/queries/tse_nachsignier.sql`, `tse_signaturen.sql`, `tse_konfiguration.sql` — Queries
- `database/migrations/01_initial.up.sql:405-459` — TSE-Tabellen
- `backend/domain/tse/client.go` — Interface (kassenID-Param ist tx-ID; UpdateTransaction toter Code)
- `backend/domain/settings/tse_konfiguration.go`, `backend/api/settings/http/*` — Konfiguration + Verbindungstest
- `backend/api/bondruck/application/escpos/formatter.go:234-251` — TSE-Abschnitt + QR auf dem Beleg
- `frontend/src/lib/EinstellungenBackend.ts:23-131`, `frontend/src/admin/settings/EinstellungenPage.tsx:189-260` — TSE-Settings-UI
- `temp/fiskaly_sign_de_api_spec.json` — API-Spec 2.2.2 (Referenz für Kontrakt-Tests)
- Verwaltung fehlgeschlagener Druckaufträge (Commit `c8a8d9c`-Muster, handbuch.md §4.6) — Vorbild für Nachsignier-Verwaltung
- Live-Test-TSS in fiskaly TEST: TSS `728e3cda-…`, Client `90977ec5-…` (für Integrationstests nutzbar)

## Resolved decisions

- **TSS-Setup-Vollautomatisierung** (User-Entscheidung): jotti legt die TSS an, führt PUK→PIN→INITIALIZED durch und registriert den Client mit der Kassen-Seriennummer — eigene Phase 5. Manuelle Eingabe von TSS-ID/Client-ID bleibt als Fallback für bestehende TSS erhalten.
- **Worker wie Druckaufträge** (User-Entscheidung): Fehlerzähler + Backoff + `fehlgeschlagen`-Status + Admin-UI.
- **tx_id auf UUIDv4 umstellen** (User-Entscheidung nach Klärung): v4 einmal erzeugen, als `tseTxId` in Event-Daten persistieren; Beleg-Druck liest das Feld statt v5-Neuableitung. Umsetzung in Phase 6.
- **Phasenzuschnitt bestätigt**; durch die Vollautomatisierung kommt eine sechste Phase hinzu.
- Pre-Release-Regeln (AGENTS.md): Breaking Changes an Events/Schema/API direkt, keine Migrationspfade.

## Open questions / Risks

- **fiskaly-Kosten/Limits:** Der Setup-Wizard legt TSS an — in LIVE kostenpflichtig. Wizard muss vorhandene TSS erkennen und Doppel-Anlage verhindern; Umgebung (TEST/LIVE) deutlich anzeigen.
- **PUK/PIN-Ablage:** Klartext in DB (wie api_secret). Akzeptiert für Self-hosted; im Betreiber-Leitfaden ausweisen.
- **UUIDv4-Enforcement-Zeitpunkt** bei fiskaly unbekannt — bis Phase 6 läuft v5 (live verifiziert funktionsfähig).

---

## Phase 1: fiskaly-Client API-konform machen (I-01, I-02, I-14, I-15.1)

### Context

- `backend/repository/tse_repo/fiskaly_client.go:140-206` — Start/Finish senden Klartext-Schema
- `backend/repository/tse_repo/fiskaly_client_test.go` — Tests asserten das falsche Verhalten
- Audit-Live-Tests 1–4: Base64 Pflicht, Start-Schema muss leer sein, Timestamps = Unix-Sekunden

### What to build

Der `FiskalyTSEClient` sendet Requests, die die echte API akzeptiert: `StartTransaction` ohne Schema (nur `state` + `client_id`), `FinishTransaction` mit Base64-codiertem `process_data`. txID-Path-Escaping vereinheitlichen. Kontrakt-Tests bilden die Spec ab (Base64-Assert, leerer Start-Body, Revision-Sequenz 1→2, UUID-Pfade, Spec-Beispiel `Beleg^0.00_2.55_0.00_0.00_0.00^2.55:Bar`). Ein per Env-Variable aktivierbarer Integrationstest signiert eine echte Transaktion gegen die fiskaly-TEST-TSS.

### Acceptance criteria

- [ ] Start-Request enthält kein `schema`; Finish-Request enthält `process_data` ausschließlich Base64-codiert
- [ ] Kontrakt-Tests prüfen Base64, leeren Start und Revisionsfolge; `make check` grün
- [ ] Env-gated Integrationstest (TEST-Credentials) durchläuft Start→Finish erfolgreich und validiert `qr_code_data`-Prefix `V0;`
- [ ] Manuell verifiziert: Direktverkauf in Dev-Umgebung erzeugt Event mit gefülltem `tseData` (kein Nachsignier-Auftrag)

---

## Phase 2: processData-Formate und Vorgangsarten DSFinV-K-konform (I-03, I-04, I-05, I-15.3 + D-01–D-05)

### Context

- `backend/api/table/application/tse_signing.go:370-398` — Bestellformat `4x Maß Bier_…` statt CSV
- `backend/api/kasse/application/tse_signing.go:20, 61-95, 254-269` — `SonstigerVorgang-V1`, Geldtransit/Differenz als Freitext
- DSFinV-K Anhang I (S. 107–111), AEAO 2.2.3.6.1–2.2.3.6.3 — verbindliche Formate
- `docs/compliance.md` §3.2–§3.4, `docs/handbuch.md` §3.13 — falsche Format-/Mapping-Angaben

### What to build

`Bestellung-V1`-processData als CSV `<Menge>;"<Bezeichnung>";<Preis>` (Zeilentrenner `\r`, Anführungszeichen-Verdopplung, Brutto-Einzelpreis 2 Nachkommastellen). processType `SonstigerVorgang` ohne Suffix. Geldtransit und Kassendifferenz als `Kassenbeleg-V1` mit `Beleg^0.00_…_0.00^<±Betrag>:Bar` (analog Auszahlung); Tagesabschluss bleibt `SonstigerVorgang`. Guard: Zahlungen von `0.00` entfallen. Parallel die Doku korrigieren: Steuerbetrags-Reihenfolge (19/7/10,7/5,5/0), Bestellformat, processType-Namen, „processData bei Start leer", Vorgangsarten-Mapping.

### Acceptance criteria

- [ ] Bestellung-processData entspricht dem Spec-Beispiel inkl. Quoting (`2;"Eisbecher ""Himbeere""";3.99`)
- [ ] Kein Vorkommen von `SonstigerVorgang-V1` mehr in Code und Doku
- [ ] Geldtransit/Differenz/Auszahlung signieren als `Kassenbeleg-V1`; Unit-Tests decken positive und negative Beträge ab
- [ ] processData ohne `0.00:Bar`-Zahlungsteil, wenn Zahlbetrag 0
- [ ] compliance.md §3.2–§3.4 und handbuch.md §3.13 stimmen mit DSFinV-K Anhang I/AEAO überein (D-01–D-05 geschlossen)

---

## Phase 3: Beleg-Vollständigkeit (I-06, I-10, I-15.5, Teil von I-08)

### Context

- `backend/api/table/application/kassenbeleg_command.go:68-73, 139-160` — Direktverkauf-Zweig ohne TSE-Daten
- `backend/api/table/application/kassenbeleg_command.go:191-218` — Tisch-Zweig mit Fallback auf `tse_signaturen` (Vorbild)
- `backend/domain/kasse/tisch_session.go:39-41` — Durchbedienen-Zeitstempel nur aus TSEData
- AEAO 1.14.2/1.14.3 — Ausfallvermerk und Zeiten vom Aufzeichnungssystem

### What to build

Direktverkaufs-Kassenbelege erhalten denselben TSE-Pfad wie Tisch-Zahlungen: TSE-Daten aus dem Event, Fallback auf die Signatur-Seitentabelle, sonst Ausfallvermerk (dafür `TSEAusfall`-Flag auch für `direktverkauf-getaetigt`). Der Durchbedienen-Pflichtaufdruck fällt auf die Event-Zeit der ersten Bestellung zurück, wenn keine TSE-logTime vorliegt. Stornobeleg für `direktverkauf-storniert` wird druckbar (`beleg-drucken` akzeptiert die Stornierung als Quelle).

### Acceptance criteria

- [ ] Direktverkauf-Beleg druckt TSE-Transaktionsnummer, Signaturzähler, Seriennummer, Zeiten, Signatur und QR-Code, sobald signiert
- [ ] Bei TSE-Ausfall trägt der Direktverkauf-Beleg den Ausfallvermerk
- [ ] Beleg eines Tisches, dessen erste Bestellung während eines Ausfalls erfasst wurde, druckt den Startzeitpunkt aus der Event-Zeit
- [ ] Stornierter Direktverkauf ist als Stornobeleg druckbar

---

## Phase 4: Robustheit — Worker, Timeout-Budget, Ausfalldokumentation (I-07, I-13, I-08)

### Context

- `backend/app/tse_nachsignier_worker.go:112-142` — Poison-Pill (Live-Test 8: 409 `E_TX_NO_TYPE_DEFINED`), kein Backoff
- `backend/sqlc/queries/tse_nachsignier.sql` — `ORDER BY id ASC LIMIT 20` → Head-of-Line-Blocking
- `backend/repository/tse_repo/fiskaly_client.go:20-23` — 10-s-Timeout × 3 Retries im Request-Pfad
- Druckauftrags-Verwaltung (handbuch.md §4.6) — Statusmodell-Vorbild

### What to build

Worker-Hardening nach Druckauftrags-Muster: `versuche`/`letzter_fehler`/`naechster_versuch_am`, exponentielles Backoff, Status `fehlgeschlagen` nach N Versuchen, `verworfen`/Zurücksetzen über Admin-Endpunkte und Admin-UI. Vor dem Nachsignieren fragt der Worker den fiskaly-Ist-Zustand der Transaktion ab und quittiert bereits abgeschlossene direkt (heilt das 409-Szenario). Im synchronen Kassier-Pfad bekommt der Signierversuch eine Gesamt-Deadline (einstelliger Sekundenbereich, max. 1 Versuch); die volle Retry-Strategie lebt nur im Worker. Die Admin-Ansicht der Aufträge (offen/fehlgeschlagen, Zeiten, letzter Fehler) dient zugleich als automatisierte TSE-Ausfalldokumentation (AEAO 1.14.1).

### Acceptance criteria

- [ ] Ein dauerhaft fehlschlagender Auftrag erreicht `fehlgeschlagen` und blockiert keine neueren Aufträge mehr
- [ ] Worker quittiert eine bei fiskaly bereits FINISHED-Transaktion ohne neuen Signierversuch (409-Szenario getestet)
- [ ] Kassieren-Request wartet bei fiskaly-Störung maximal die definierte Deadline, dann greift der Ausfallpfad
- [ ] Admin sieht offene/fehlgeschlagene Nachsignierungen mit Zeitraum und Fehlergrund und kann zurücksetzen/verwerfen

---

## Phase 5: TSS-Setup-Automatisierung aus jotti (I-09, I-15.4, D-07)

### Context

- fiskaly-Lifecycle (Postman/Spec): TSS anlegen → `UNINITIALIZED` → Admin-PIN aus PUK → Admin-Auth → `INITIALIZED` → Client registrieren (Admin-Auth nötig)
- `backend/api/settings/http/command_handler.go:39-52`, `frontend/src/admin/settings/EinstellungenPage.tsx:189-260` — bestehende TSE-Konfiguration
- `backend/domain/settings/kassenidentitaet.go` — Kassen-Seriennummer (UUID)
- Audit I-09: QR-`<kassen-seriennummer>` = fiskaly-Client-`serial_number`; DSFinV-K ≥ 2.3: keine `/` und `_`

### What to build

Ein Admin-Wizard „TSE einrichten": Nach Eingabe von API-Key/-Secret legt jotti die TSS an, durchläuft PUK→PIN→`INITIALIZED`, registriert den Client mit der Kassen-Seriennummer als `serial_number` und speichert TSS-ID/Client-ID/PUK/PIN in `tse_konfiguration`. Vorhandene TSS werden erkannt (keine Doppel-Anlage); Umgebung (TEST/LIVE) wird vor der Anlage angezeigt und bestätigt. Die manuelle Eingabe bleibt für bestehende TSS erhalten. Der Verbindungstest prüft zusätzlich Client-State `REGISTERED` und die Übereinstimmung `serial_number` ↔ Kassen-Seriennummer. Betreiber-Leitfaden dokumentiert beide Wege und die PUK/PIN-Ablage.

### Acceptance criteria

- [ ] Wizard richtet aus leerem fiskaly-TEST-Konto eine signierfähige TSS samt Client ein (Ende-zu-Ende verifiziert)
- [ ] Client wird mit der Kassen-Seriennummer registriert; Verbindungstest meldet Abweichung als Fehler
- [ ] Wizard verhindert Doppel-Anlage und zeigt die Umgebung vor kostenwirksamen Aktionen an
- [ ] Betreiber-Leitfaden beschreibt Setup, Fallback und PUK/PIN-Aufbewahrung (D-07 geschlossen)

---

## Phase 6: Konsolidierung — Dedup, Interface, UUIDv4, Status-Doku (I-12, I-15.2, I-11, D-06)

### Context

- Drei nahezu identische `tse_signing.go` (table/direktverkauf/kasse) — Drift bereits eingetreten (TSEAusfall nur in table)
- `backend/domain/tse/client.go` — `kassenID`-Param ist tx-ID, `UpdateTransaction` toter Code, ignorierter `transactionNumber`-Param
- `backend/api/table/application/kassenbeleg_command.go:198` — v5-Neuableitung beim Beleg-Druck
- `docs/compliance.md` §4.1/§8, `docs/anforderungen.md` F-02 — veraltete Statusangaben

### What to build

Die Signier-Orchestrierung (`signEventWithTSE`, Ausfallpfad, Betrags-/Zeit-Helfer) und die processData-Formatter wandern in ein gemeinsames Paket; die Event-spezifischen Teile werden Parameter. Das `TSEClient`-Interface beschreibt die tatsächliche Semantik (txID statt kassenID, kein `UpdateTransaction`). tx_id-Erzeugung wird auf UUIDv4 umgestellt: einmal erzeugen, als `tseTxId` im Event persistieren, Beleg-Druck und Nachsignier-Pfad lesen das Feld. Doku-Status aktualisieren: GoBD-Tabelle (Verkettung), F-02-Status, `.env`-Angabe → Admin-UI/DB.

### Acceptance criteria

- [ ] `signEventWithTSE` und processData-Formatter existieren genau einmal; alle drei Kontexte nutzen das gemeinsame Paket
- [ ] `TSEClient`-Interface ohne `UpdateTransaction`, Parameter heißen nach ihrer tatsächlichen Bedeutung
- [ ] Neue Events tragen `tseTxId` (UUIDv4); Beleg-Fallback findet nachsignierte Signaturen über das persistierte Feld
- [ ] compliance.md §4.1/§8 und anforderungen.md F-02 spiegeln den realen Stand (D-06 geschlossen)
