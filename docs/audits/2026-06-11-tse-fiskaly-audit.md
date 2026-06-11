# Audit: TSE- und fiskaly-Integration

> **Datum:** 11.06.2026 · **Scope:** `backend/domain/tse`, `backend/repository/tse_repo`, `backend/api/*/application/tse_signing.go`, `backend/app/tse_nachsignier_worker.go`, TSE-Settings (Backend + Frontend), Kassenbeleg-Druck, Doku (`docs/compliance.md`, `docs/handbuch.md` §3.13, `docs/anforderungen.md` F-02)

## Methodik

1. **Primärquellen-Recherche:** DSFinV-K (Anhang I im Volltext aus offizieller PDF extrahiert), AEAO zu § 146a AO (BMF-Schreiben 30.06.2023, Volltext), fiskaly SIGN DE API-Spec 2.2.2 (`temp/fiskaly_sign_de_api_spec.json`), offizielle Postman-Collection.
2. **Live-Validierung:** 10 Testszenarien gegen die fiskaly **TEST**-Umgebung mit den Dev-Credentials (TSS `728e3cda-12c4-4390-b9ea-7e6f8389059e`, Client `90977ec5-c94d-4bd6-b04b-61b4428a71cb`, Client-Serial `jotti-audit-test-kasse-1` — bleibt für künftige Integrationstests bestehen).
3. **Code-Audit:** Cross-Layer-Konsistenz (Frontend ↔ HTTP ↔ Application ↔ Repository ↔ DB), Vereinfachungs-Review, `make check` (Build + Lint + Unit-Tests: **grün**).

Severity: 🔴 Kritisch (Integration funktionsunfähig oder klarer Rechtsverstoß) · 🟠 Hoch (Compliance-Lücke / Datenverlust-Risiko) · 🟡 Mittel · ⚪ Niedrig.

---

## Executive Summary

Die TSE-Integration ist architektonisch solide (atomares Festzelt-Muster, Outbox, deterministische tx-IDs, Secret-Maskierung), aber **funktional defekt**: `process_data` wird nicht Base64-codiert übertragen, wodurch fiskaly **jeden** Signier-Request mit `400 E_PARSER` ablehnt (live verifiziert). In der Praxis wird derzeit **keine einzige Transaktion signiert**; alle Vorgänge laufen in die Nachsignier-Queue und schlagen dort endlos fehl. Die Unit-Tests bestätigen das fehlerhafte Verhalten, weil der HTTP-Fake kein Base64 prüft.

Daneben bestehen mehrere DSFinV-K-/AEAO-Konformitätsverstöße (processData bei Start nicht leer, falsches `Bestellung-V1`-Format, nicht existenter processType `SonstigerVorgang-V1`, falsche Vorgangsart für Geldtransit/Kassendifferenz) sowie eine Beleg-Lücke (§ 6 KassenSichV) beim Direktverkauf. Die Doku enthält vier sachliche Fehler gegenüber den Primärquellen.

---

## A. Findings: Implementierung

### 🔴 I-01 — `process_data` nicht Base64-codiert → Integration vollständig funktionsunfähig

- **Wo:** [fiskaly_client.go:66-69](../../backend/repository/tse_repo/fiskaly_client.go#L66-L69) (`rawSchema`), Aufrufer in allen drei `tse_signing.go`
- **Beleg:** API-Spec `ProcessDataRaw.process_data`: *„Identifies the data of the transaction as a **base64 encoded string**"*. Live-Test 1 (jotti-Verhalten 1:1 nachgestellt): `HTTP 400 {"code":"E_PARSER","message":"error parsing request body: illegal base64 data at input byte 5"}`.
- **Impact:** Jede Signierung schlägt fehl → jeder Vorgang landet als Nachsignier-Auftrag in der Queue → der Worker wiederholt denselben fehlerhaften Request alle 5 s endlos. Es entsteht **kein einziges TSE-Log**, kein Beleg erhält TSE-Daten.
- **Empfehlung:** Base64-Encoding (Decoding für Anzeige) zentral im `FiskalyTSEClient` kapseln — Domain-Schicht bleibt bei Klartext-processData. Kontrakt-Test mit Base64-Assertion ergänzen.
- **Aufwand:** S

### 🔴 I-02 — processType/processData werden bei `StartTransaction` mitgesendet

- **Wo:** [fiskaly_client.go:140-170](../../backend/repository/tse_repo/fiskaly_client.go#L140-L170)
- **Beleg:** DSFinV-K Anhang I (S. 107): *„Für alle Vorgangstypen gilt, dass processType und processData für die StartTransaction-Operation **immer leer** sind."* Ebenso API-Spec: *„When you start a transaction, the type and data of the transaction must be empty. This is required by the DSFinV-K."* Live-Test 4: fiskaly akzeptiert gefüllte Schemas bei Start zwar technisch (200), die Daten landen dann aber im signierten Start-Log — nicht konform.
- **Impact:** DSFinV-K-Verstoß in jedem TSE-Log; bei Kassenbeleg-V1 zusätzlich Widerspruch zur eigenen Doku (processData steht erst bei Finish fest).
- **Empfehlung:** Start-Request nur mit `state: ACTIVE` + `client_id` senden (wie offizielle Postman-Collection); Schema ausschließlich bei Finish.
- **Aufwand:** S

### 🔴 I-03 — `Bestellung-V1`-processData hat falsches Format

- **Wo:** [table/application/tse_signing.go:370-398](../../backend/api/table/application/tse_signing.go#L370-L398) (`buildBestellungProcessData` → `4x Maß Bier_2x Weißwurst`)
- **Beleg:** DSFinV-K Anhang I (S. 110): CSV-Darstellung `<Menge>;"<Bezeichnung>";<Preis>` — Zeilentrenner `\r` (U+000D), Spaltentrenner `;`, Bezeichnung **in Anführungszeichen** (innere `"` verdoppeln), Preis = **Brutto-Einzelpreis** mit exakt 2 Nachkommastellen, Menge mit minimalen Nachkommastellen. Beispiel der Spec: `2;"Eisbecher ""Himbeere""";3.99`.
- **Impact:** Alle Bestellungs-Signaturen wären inhaltlich nicht DSFinV-K-konform; das falsche Format landet auch im QR-Code-Feld `<processData>`. Der Preis fehlt komplett.
- **Empfehlung:** Formatter neu nach Spec implementieren (inkl. Quoting-Regel und Preis aus `pos.Einzelpreis`).
- **Aufwand:** S

### 🔴 I-04 — processType `SonstigerVorgang-V1` existiert nicht

- **Wo:** [kasse/application/tse_signing.go:20](../../backend/api/kasse/application/tse_signing.go#L20)
- **Beleg:** DSFinV-K Anhang I (S. 111): `processType: SonstigerVorgang` — **ohne** `-V1`-Suffix (anders als `Kassenbeleg-V1`/`Bestellung-V1`). Live-Test 7: fiskaly validiert den Wert nicht und schreibt den falschen String ungeprüft ins TSE-Log und in `qr_code_data`.
- **Impact:** Nicht-konformer processType in TSE-Log, DSFinV-K-Export (`transactions_tse.csv`) und Beleg-QR.
- **Empfehlung:** Konstante auf `SonstigerVorgang` korrigieren (Code + Doku, siehe D-03).
- **Aufwand:** S

### 🔴 I-05 — Geldtransit und Kassendifferenz mit falscher Vorgangsart abgesichert

- **Wo:** [kasse/application/tse_signing.go:61-95](../../backend/api/kasse/application/tse_signing.go#L61-L95) (Geldtransit, DifferenzSollIst als `SonstigerVorgang-V1` mit Freitext)
- **Beleg:** AEAO zu § 146a, Nr. 2.2.3.6.1: Die Art „Kassenbeleg" *„gilt auch für abgeschlossene Vorgänge, die Geschäftsvorfälle abbilden, an denen nur der Unternehmer selbst beteiligt ist (z. B. **Eigenbelege über Ein- oder Auszahlungen**)."* Geldtransit/Einlage/Entnahme und Kassendifferenz sind solche Geschäftsvorfälle (DSFinV-K GV_TYP `Geldtransit`, `DifferenzSollIst`).
- **Impact:** Geschäftsvorfälle ohne Kassenbeleg-Absicherung → bei Prüfung fehlt die Kassensturzfähigkeit über die TSE-Daten (Zahlbeträge sind nicht im Kassenbeleg-Format signiert). Die `Auszahlung` ist dagegen bereits **korrekt** als `Kassenbeleg-V1` umgesetzt.
- **Empfehlung:** Geldtransit und DifferenzSollIst als `Kassenbeleg-V1` mit `Beleg^0.00_0.00_0.00_0.00_0.00^<±Betrag>:Bar` absichern (analog Auszahlung); Tagesabschluss bleibt zulässig als `SonstigerVorgang`.
- **Aufwand:** M

### 🔴 I-06 — Direktverkauf-Kassenbeleg ohne TSE-Pflichtangaben (§ 6 KassenSichV)

- **Wo:** [table/application/kassenbeleg_command.go:139-160](../../backend/api/table/application/kassenbeleg_command.go#L139-L160) — der Direktverkauf-Zweig lädt weder `TSEData` noch den Signatur-Fallback (`GetTSESignaturByTxID`) noch einen Ausfallvermerk; die lokale `direktverkaufGetaetigtV1Data` (Z. 68-73) hat nicht einmal ein `tseData`-Feld.
- **Impact:** Direktverkaufs-Belege werden **immer** ohne TSE-Transaktionsnummer, Signaturzähler, Signatur, TSE-Zeiten und QR-Code gedruckt — Verstoß gegen § 6 KassenSichV, sobald die TSE konfiguriert ist. Außerdem kein Ausfallvermerk nach AEAO 1.14.2.
- **Empfehlung:** Direktverkauf-Zweig analog zum Tisch-Zweig implementieren (TSEData aus Event → Fallback Seitentabelle → Ausfallvermerk).
- **Aufwand:** M

### 🟠 I-07 — Nachsignier-Worker: Poison-Pill + Head-of-Line-Blocking

- **Wo:** [tse_nachsignier_worker.go:112-142](../../backend/app/tse_nachsignier_worker.go#L112-L142), [tse_nachsignier.sql](../../backend/sqlc/queries/tse_nachsignier.sql)
- **Beleg:** Live-Test 8: War die fiskaly-Transaktion bereits FINISHED (z. B. Finish erfolgreich, aber Quittierung/DB-Write fehlgeschlagen), liefert das erneute `StartTransaction` (rev=1, ohne Schema nach I-02-Fix) `409 E_TX_NO_TYPE_DEFINED`. 409 ist nicht retry-fähig → `processAuftrag` schlägt dauerhaft fehl. Die Queue hat **keinen Fehlerzähler, kein Backoff, keinen Dead-Letter-Status**, und `GetOffeneTSENachsignierAuftraege` (`ORDER BY id ASC LIMIT 20`) bedeutet: 20 dauerhaft fehlschlagende Aufträge blockieren alle neueren **für immer**.
- **Empfehlung:** (a) Worker prüft vor Start per `GET /tss/{id}/tx/{tx_id}` den Ist-Zustand und quittiert bereits abgeschlossene Transaktionen direkt; (b) Spalten `versuche`, `letzter_fehler`, `naechster_versuch_am` + exponentielles Backoff + Status `fehlgeschlagen` nach N Versuchen mit Admin-Sichtbarkeit.
- **Aufwand:** M

### 🟠 I-08 — TSE-Ausfalldokumentation fehlt; Ausfallvermerk nur bei Tisch-Zahlungen

- **Wo:** [table/application/tse_signing.go:302-322](../../backend/api/table/application/tse_signing.go#L302-L322) (`withZahlungEventTSEAusfall` existiert nur für `zahlung-kassiert`); Direktverkauf/Bestellung/Storno/Auszahlung haben kein `TSEAusfall`-Flag
- **Beleg:** AEAO zu § 146a, Nr. 1.14.1: *„Ausfallzeiten und ‑grund einer TSE sind zu dokumentieren … kann auch automatisiert … erfolgen."* Nr. 1.14.2: Ausfall muss *„auf einem eventuellen Beleg ersichtlich sein."*
- **Impact:** Es gibt keine persistierte, abrufbare Ausfallzeitraum-Dokumentation (nur Log-Warnungen + Queue-Einträge); Direktverkaufs-Belege können den Ausfall nicht ausweisen. Die Nachsignierung ersetzt die Ausfalldoku nicht — die nachgeholte Signatur trägt eine spätere logTime und ist rechtlich keine Heilung des Ausfalls.
- **Empfehlung:** `TSEAusfall`-Flag für alle belegrelevanten Events; Ausfallfenster (von/bis, Grund) als eigene Tabelle oder ableitbar aus der Queue persistieren und im Admin anzeigen.
- **Aufwand:** M

### 🟠 I-09 — Beleg-QR und Beleg-Seriennummer können auseinanderfallen (fiskaly-Client-`serial_number`)

- **Wo:** Konzeptlücke — jotti verwaltet die fiskaly-Client-Registrierung nicht; [kassenbeleg_command.go:248](../../backend/api/table/application/kassenbeleg_command.go#L248) druckt die jotti-Kassenidentitäts-UUID
- **Beleg:** DSFinV-K Anhang I QR-Code-Feld `<kassen-seriennummer>` = *„Seriennummer (Client-Id) der Kasse"*. Live-Test 3: fiskalys `qr_code_data` enthält die bei der **Client-Registrierung** vergebene `serial_number` (`jotti-audit-test-kasse-1`), nicht jottis Kassen-UUID.
- **Impact:** Registriert der Verein den fiskaly-Client mit beliebiger Seriennummer, widersprechen sich der gedruckte Beleg („Kassen-ID: <UUID>") und der QR-Code → Belegprüfung (Amtsträger-App) ordnet den Bon keiner gemeldeten Kasse zu. Zudem verbietet DSFinV-K ≥ 2.3 `/` und `_` in der `serial_number`.
- **Empfehlung:** Anforderung dokumentieren (Betreiber-Leitfaden + Settings-Hilfetext): fiskaly-Client **muss** mit jottis Kassen-Seriennummer als `serial_number` registriert werden. Optional: Client-Registrierung aus jotti heraus automatisieren (Admin-PIN-Flow) oder beim Verbindungstest `client.serial_number` gegen die Kassenidentität prüfen.
- **Aufwand:** S (Doku) / M (Prüfung im Verbindungstest)

### 🟡 I-10 — Durchbedienen-Pflichtaufdruck entfällt nach TSE-Ausfall dauerhaft

- **Wo:** [tisch_session.go:39-41](../../backend/domain/kasse/tisch_session.go#L39-L41) — `ErsteBestellungLogTime` wird nur aus `TSEData.LogTimeStart` gesetzt
- **Beleg:** AEAO 2.2.3.6.2: Bei Inanspruchnahme der Erleichterung *„muss der Start-Zeitpunkt der ersten Bestellung zusätzlich auf dem Beleg abgedruckt werden"*; AEAO 1.14.3: bei TSE-Ausfall stellt das Aufzeichnungssystem Datum/Uhrzeit bereit.
- **Impact:** Wurde die erste Bestellung eines Tisches während eines TSE-Ausfalls erfasst, fehlt der Pflichtaufdruck auf **allen** späteren Zahlungsbelegen dieses Tisches.
- **Empfehlung:** Fallback auf die Event-Zeit (`evt.Time`) der ersten Bestellung, wenn keine TSE-logTime vorliegt.
- **Aufwand:** S

### 🟡 I-11 — tx_id ist UUIDv5, fiskaly fordert UUIDv4

- **Wo:** alle `tseTransactionIDFor*`-Funktionen (`uuid.NewSHA1` → Version 5)
- **Beleg:** API-Spec: *„WARNING: Future major versions of SIGN DE API will strictly enforce UUIDv4 format requirements."* Live-Test 5: v5 wird derzeit akzeptiert.
- **Impact:** Funktioniert heute; Bruchrisiko bei API-Major-Update. Die Determinismus-Eigenschaft (Idempotenz bei Retries) ist fachlich wertvoll und sollte nicht aufgegeben werden.
- **Empfehlung:** Risiko bewusst dokumentieren; alternativ v4 generieren und die Zuordnung Event↔tx_id persistieren (größerer Umbau, derzeit nicht nötig).
- **Aufwand:** S (Doku) / L (Umbau)

### 🟡 I-12 — Dreifach duplizierter TSE-Signier-Code mit beginnender Drift

- **Wo:** [table/application/tse_signing.go](../../backend/api/table/application/tse_signing.go), [direktverkauf/application/tse_signing.go](../../backend/api/direktverkauf/application/tse_signing.go), [kasse/application/tse_signing.go](../../backend/api/kasse/application/tse_signing.go)
- **Was:** `signEventWithTSE`, `handleSignierAusfall`, `tseBetragString`, `tseTimeString`, `nonZeroTime`, `tseNachsignierAuftrag`, `eventSignierungErgebnis` existieren 3×, `buildKassenbelegProcessData(WithFaktor)` 2×. Die table-Variante hat bereits einen zusätzlichen `withTSEAusfall`-Parameter — die Drift, die I-08 verursacht, ist also schon eingetreten.
- **Empfehlung:** Gemeinsames Paket (z. B. `backend/api/helper/tsesign` oder `domain/tse/processdata`) für Signier-Orchestrierung + processData-Formatter; Event-spezifische Teile als Parameter.
- **Aufwand:** M

### 🟡 I-13 — Synchrones Signieren kann den Kassier-Request > 40 s blockieren

- **Wo:** [fiskaly_client.go:20-23](../../backend/repository/tse_repo/fiskaly_client.go#L20-L23) (10 s HTTP-Timeout, 3 Retries, Backoff bis 5 s) × Start **und** Finish im Request-Pfad
- **Impact:** Bei fiskaly-Störung wartet die Servicekraft im schlimmsten Fall ~80 s auf die Kassieren-Antwort, bevor der Ausfallpfad greift. Im Festzelt inakzeptabel; konterkariert das „DON'T BLOCK THE TILL"-Ziel der Outbox.
- **Empfehlung:** Gesamt-Deadline für den Signierversuch im Request-Pfad (z. B. `context.WithTimeout` 3–5 s, max. 1 Versuch); volle Retry-Strategie nur im Worker.
- **Aufwand:** S

### 🟡 I-14 — Unit-Tests validieren das falsche API-Verhalten

- **Wo:** [fiskaly_client_test.go:86-97](../../backend/repository/tse_repo/fiskaly_client_test.go#L86-L97) u. a. — Fake-Server prüft weder Base64 noch Schema-leer-bei-Start; `"kasse-1"` als tx_id (kein UUID)
- **Impact:** Genau deshalb blieb I-01 unentdeckt: `make check` ist grün, obwohl kein einziger Live-Request funktioniert.
- **Empfehlung:** Kontrakt-Tests gegen die Spec-Beispiele (Base64-Assert, Start-Body ohne `schema`, UUID-Pfad, Revision-Sequenz 1→2); optional ein per Env-Flag aktivierbarer Integrationstest gegen die TEST-TSS aus diesem Audit.
- **Aufwand:** M

### ⚪ I-15 — Kleinere Punkte

1. **Inkonsistentes Path-Escaping:** `StartTransaction` escapt `txID` nicht ([fiskaly_client.go:150](../../backend/repository/tse_repo/fiskaly_client.go#L150)), `FinishTransaction` schon (Z. 186). Bei UUIDs harmlos, trotzdem angleichen.
2. **Interface-Semantik:** `TSEClient.StartTransaction(kassenID …)` — der Parameter ist tatsächlich die **Transaktions-ID**; `transactionNumber` bei `FinishTransaction` wird ignoriert; `UpdateTransaction` liefert immer einen Fehler. Das Interface beschreibt BSI-Semantik, implementiert wird fiskaly-Upsert-Semantik — Parameter umbenennen, `UpdateTransaction` aus dem Interface entfernen (Atomares Muster braucht es nicht).
3. **`0.00:Bar`-Edge:** DSFinV-K: *„Zahlungen von 0.00 müssen entfallen."* `buildKassenbelegProcessData` würde bei Zahlbetrag 0 `0.00:Bar` erzeugen (Live-Test 10: fiskaly validiert das nicht). Guard ergänzen.
4. **Verbindungstest prüft Client nicht:** `TestConnection` prüft nur TSS-State; ein `DEREGISTERED`-Client fällt erst beim Signieren auf. `GET /tss/{id}/client/{client_id}` ergänzen.
5. **Direktverkauf-Storno ohne Beleg-Druck:** `KassenbelegDrucken` akzeptiert nur `verkaufId` des positiven Verkaufs; für `direktverkauf-storniert` ist kein Stornobeleg druckbar.

---

## B. Findings: Dokumentation

### 🔴 D-01 — `compliance.md` §3.4: Reihenfolge der fünf Steuerbeträge falsch beschriftet

Dokumentiert: `<Betrag_Normal>_<Betrag_Ermaessigt>_<Betrag_Null>_<Betrag_Besonderer_Satz>_<Betrag_Befreit>`.
Offiziell (DSFinV-K Anhang I, S. 109): **1.** Allgemeiner Steuersatz (19 %) · **2.** Ermäßigter Steuersatz (7 %) · **3.** Durchschnittssatz § 24 (1) Nr. 3 UStG (10,7 %) · **4.** Durchschnittssatz § 24 (1) Nr. 1 UStG (5,5 %) · **5.** 0 %.
Der **Code** ([buildKassenbelegProcessDataWithFaktor](../../backend/api/table/application/tse_signing.go#L328-L368)) mappt korrekt (befreit → Position 5); nur die Doku-Labels sind falsch.

### 🔴 D-02 — `compliance.md` §3.4: `Bestellung-V1`-Format falsch

Dokumentiert: *„Positionen als strukturierter Text (z. B. `4x Maß Bier_2x Weißwurst`) gemäß AEAO § 146a Anhang I"* — das Beispiel ist frei erfunden und widerspricht der zitierten Quelle (korrekt: CSV `Menge;"Bezeichnung";Preis`, siehe I-03). Betrifft auch das Szenario-Beispiel in §3.6 und `temp/tse_ressources.md`.

### 🔴 D-03 — `compliance.md` §3.3 / `handbuch.md` §3.13: `SonstigerVorgang-V1` falsch

*„Die -V1-Endung ist Bestandteil des offiziellen Strings"* stimmt nur für `Kassenbeleg-V1` und `Bestellung-V1`. Der dritte Typ heißt `SonstigerVorgang` (DSFinV-K Anhang I, S. 111). Betroffen: processType-Tabelle, Event-Mapping-Tabellen in beiden Dokumenten.

### 🟠 D-04 — `compliance.md` §3.2: StartTransaction-Request falsch beschrieben

Tabelle nennt als Request *„Kassen-ID, processType, initiale processData"* — laut DSFinV-K sind processType/processData bei Start **immer leer**. (Die UpdateTransaction-Zeile „nur für Bestellung-V1 und SonstigerVorgang" ist sinngemäß korrekt.)

### 🟠 D-05 — `handbuch.md` §3.13: Mapping Geldtransit/Kassendifferenz → `SonstigerVorgang-V1` widerspricht AEAO 2.2.3.6.1

Eigenbelege über Ein-/Auszahlungen sind als Art „Kassenbeleg" abzusichern (siehe I-05). Mapping-Tabelle und compliance.md §3.3 anpassen.

### 🟡 D-06 — Veraltete/zu optimistische Statusangaben

1. `compliance.md` §4.1: „Kryptografische Verkettung ❌ Fehlt — Keine TSE-Signatur" — die TSE-Signierung existiert inzwischen (wenn auch defekt, I-01). Tabelle aktualisieren.
2. `anforderungen.md` F-02: Status „✅" — angesichts I-01 (kein Request funktioniert gegen die echte API) nicht haltbar, bis der Fix verifiziert ist.
3. `compliance.md` §8 + Pflichtenmatrix: „TSE-API-Keys konfigurieren (`.env`)" — tatsächlich werden die Credentials über das Admin-UI in der DB (`tse_konfiguration`) gepflegt, nicht in der `.env`.

### 🟡 D-07 — Fehlende Betreiber-Doku: fiskaly-Client-Registrierung

Weder `compliance.md` noch der Betreiber-Leitfaden beschreiben den TSS-Lifecycle (CREATED → UNINITIALIZED → Admin-PIN → INITIALIZED → Client-Registrierung) oder die Pflicht, den fiskaly-Client mit jottis Kassen-Seriennummer als `serial_number` zu registrieren (I-09, DSFinV-K ≥ 2.3: keine `/` oder `_`).

### ⚪ D-08 — Korrekt verifizierte Doku-Aussagen (kein Handlungsbedarf)

Folgende zentrale Aussagen wurden gegen Primärquellen geprüft und sind **korrekt**: Festzelt-Muster/Erleichterung (DSFinV-K Tz. 2.7, Anhang H; AEAO 2.2.3.6.2), Durchbedienen-Pflichtaufdruck, BYOD-Eingabegeräte ohne eigene TSE/Meldepflicht (AEAO 2.7/2.9), Belegausgabe-Befreiung § 146a Abs. 2 Satz 2 AO, Meldepflicht-Fristen (1.1.2025 / 31.7.2025 / 1 Monat), DSFinV-K-Dateinamen englisch/kleingeschrieben, QR-Code-Aufbau Anhang I, „Arbeitsbon ≠ Kassenbeleg", DSFinV-K 2.4 gültig ab 1.1.2024.

---

## C. Positivbefunde (beibehalten)

- **Atomares Festzelt-Muster** (Start + sofort Finish je Vorgang) entspricht exakt der DSFinV-K-Erleichterung — Architekturentscheidung bestätigt.
- **Deterministische tx-IDs** je Event → idempotente Retries; Live-Test 6/9 bestätigt: fiskaly behandelt wiederholte PUTs mit gleicher Revision idempotent (200).
- **Outbox transaktional:** Event + Nachsignier-Auftrag in einer DB-Transaktion ([repo.go:89-130](../../backend/repository/kassenjournal_repo/repo.go#L89-L130)) — „DON'T BLOCK THE TILL" konzeptionell richtig.
- **`parseFlexibleInt` ist berechtigt:** fiskaly liefert `signature.counter` mal als Zahl, mal als String (live beobachtet).
- **Unix-Timestamps korrekt geparst**; QR-Code aus fiskalys `qr_code_data` ist DSFinV-K-Anhang-I-konform; ESC/POS-Beleg druckt alle TSE-Pflichtfelder (Tisch-Pfad).
- **Settings-Schicht konsistent** über Frontend ↔ HTTP ↔ DB; `api_secret` wird nie an das Frontend zurückgegeben (`apiSecretGesetzt`-Flag); Token-Caching mit 401-Refresh.
- **`make check` grün** (Build, Lint, Unit-Tests, `go vet`, Race-Detector).

---

## D. Live-Test-Protokoll (fiskaly TEST, 11.06.2026)

| # | Szenario | Ergebnis |
|---|----------|----------|
| 1 | Start mit Klartext-Schema (jotti-Ist-Verhalten) | **400 E_PARSER** „illegal base64 data" |
| 2 | Start nur `state`+`client_id` | 200, `signature.counter` als Zahl |
| 3 | Finish mit Base64-`process_data` | 200, `qr_code_data` im V0-Format mit Client-`serial_number` |
| 4 | Start **mit** Base64-Schema | 200 (toleriert, aber DSFinV-K-widrig) |
| 5 | tx_id als UUIDv5 | 200 (Warnung: künftig strikt UUIDv4) |
| 6 | Retry Start, gleiche Revision | 200 idempotent, `counter` als **String** |
| 7 | `process_type: SonstigerVorgang-V1` | 200 — keine Validierung durch fiskaly |
| 8 | Start rev=1 auf FINISHED-Transaktion | **409 E_TX_NO_TYPE_DEFINED** (Worker-Poison-Szenario) |
| 9 | Finish rev=2 wiederholt (gleicher Payload) | 200 idempotent |
| 10 | `0.00:Bar` im Zahlungsteil | 200 — keine Validierung durch fiskaly |

---

## E. Priorisierte Empfehlungen

1. **Sofort (Funktionsfähigkeit):** I-01 Base64 + I-02 leerer Start + Kontrakt-Tests (I-14) — ohne sie ist F-02 faktisch nicht vorhanden.
2. **Vor erstem Echtbetrieb (Compliance):** I-03 Bestellformat, I-04 processType, I-05 Vorgangsarten, I-06 Direktverkauf-Beleg, D-01–D-05 Doku-Korrekturen.
3. **Robustheit:** I-07 Worker-Hardening, I-08 Ausfalldokumentation, I-10 Durchbedienen-Fallback, I-13 Timeout-Budget.
4. **Aufräumen:** I-12 Dedup der Signier-Schicht, I-15 Kleinpunkte, I-09/D-07 Betreiber-Doku, I-11 UUID-Hinweis.
