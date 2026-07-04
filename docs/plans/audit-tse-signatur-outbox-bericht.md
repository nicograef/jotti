# Audit-Bericht: TSE-Signatur-Outbox

Branch `feat/tse-signatur-outbox-phase1` (7 Commits, Phasen 1-7 des Plans
[plan-tse-signatur-outbox.md](plan-tse-signatur-outbox.md), 127 Dateien, +7861/-5950), auditiert am 2026-07-04
gegen den Plan, die Rechtsquellen unter `docs/rechtsquellen/` (DSFinV-K 2.4, KassenSichV, AEAO zu § 146a,
fiskaly SIGN DE API v2 OpenAPI) und die Codebasis über alle Schichten (DB, Backend, Frontend, Doku, Seed, Tests).

## Ergebnis in Kürze

Die Implementierung ist in Architektur und Detail bemerkenswert konsistent: Outbox im selben Commit, ein Worker
als einziger TSE-Sprecher, eine Signaturquelle, ein Ausfallbegriff (Signaturstatus-Funktion, von Beleg und Gate
gemeinsam genutzt), Störungsprotokoll nach AEAO 1.14, processData exakt nach DSFinV-K Anhang I, Belegangaben
vollständig nach § 6 KassenSichV, fiskaly-Aufrufe konform zur OpenAPI-Spec. Alle Plan-Akzeptanzkriterien sind
in Code und Tests nachvollziehbar. `make verify` war vor den Audit-Fixes grün.

Gefunden wurden sieben kleine, klar behebbare Punkte (A1-A7, direkt umgesetzt, siehe Umsetzungsstatus) und fünf
Punkte mit Entscheidungsbedarf (B1-B5, dokumentiert in
[audit-tse-signatur-outbox-entscheidungen.md](audit-tse-signatur-outbox-entscheidungen.md)). Kein Befund stellt
das Modell in Frage; die zwei gewichtigsten (A1, A2) betreffen Fehlerklassifizierung und einen nie endenden
Störungszeitraum in einem Konfigurations-Sonderpfad.

## 1. Cross-Layer-Konsistenz (Skill-Schritt 1)

Geprüfte Flüsse, jeweils Ende-zu-Ende (UI → Backend-Klasse → Handler → Application → Repository → SQL → Schema):

- Buchen → Outbox: `writeEventInTx` schreibt Event + Signaturauftrag transaktional; die fiskalische Projektion
  deckt alle 13 Event-Typen ab, unbekannte Typen sind ein harter Fehler. Status-Guard (FOR SHARE) gegen die
  Abschluss-Barriere. In Ordnung.
- Beleg-Abruf: `service/beleg-drucken` antwortet `eingereiht`/`ausstehend`; `beleg.ts` fasst 6x alle 1,5 s nach
  (rund 10 s, plankonform); Zod-Schema = Handler-DTO. TSE-Abschnitt aus den Signaturspalten; Ausfall-Vermerke in
  zwei Varianten. In Ordnung.
- Kassenabschluss: Gate vor der Barriere über die Signaturstatus-Funktion; 409 `signaturen_ausstehend` mit
  `anzahl`/`alterSekunden`, Erfolgsantwort mit `ausfallResteAnzahl`/`ohneKonfigurationAnzahl`; UI zeigt Meldung
  und lässt erneut anfordern. Frontend-Schemas decken sich mit den Handler-DTOs. In Ordnung.
- Admin Signaturaufträge/Queue/Störungen: Endpunkte, DTOs, Zod-Schemas, Status-Enums (inkl. CHECK-Constraint)
  und UI-Guards (Zurücksetzen nur fehlgeschlagen/tse_nicht_konfiguriert, Verwerfen nur offen/fehlgeschlagen mit
  Pflicht-Begründung 1-500 Zeichen) sind deckungsgleich mit den SQL-Guards. In Ordnung.
- Dashboard: warnt ab 60 s Rückstand und bei endgültig fehlgeschlagenen Aufträgen, permanenter
  Konfigurationsalarm ohne TSE; Schwellen konsistent mit den Backend-Konstanten. In Ordnung.
- Settings/Einrichtung: Kassensitzungs-Guard in allen drei Änderungspfaden; Einrichtungs-Sweep transaktional im
  Repository. Befunde A2 und A5 (unten).
- Keine `nachsignier`-Namen mehr in Routen, Backend-Clients oder UI; nur der Fachbegriff „nachsigniert" bleibt
  (plankonform, Phase 6).

## 2. Rechtsquellen- und Fiskaly-Abgleich

- DSFinV-K 2.4 Anhang I: Kassenbeleg-V1 (Feldreihenfolge der fünf Brutto-Steuerumsätze, 0.00-Pflicht, exakt zwei
  Dezimalstellen, Vorzeichenregeln, Zahlungen `Betrag:Bar`, „Zahlungen von 0.00 müssen entfallen"),
  Bestellung-V1 (Zeilentrenner \r, Semikolon, Anführungszeichen-Verdopplung, Brutto-Einzelpreis), negative
  Mengen für Bestell-Storni (amtlich: „neuer Datensatz mit umgekehrtem Vorzeichen"), SonstigerVorgang frei,
  Start ohne processType/processData. Alles konform. Abweichung: Millisekunden-Format der TSE_TA-Zeiten (A3).
- § 6 KassenSichV: alle sieben Pflichtangaben auf dem Kassenbeleg vorhanden (Betreiber, Datum + Vorgangsbeginn/
  -ende, Menge/Art, Transaktionsnummer, Entgelt/Steuer, beide Seriennummern, Prüfwert + Signaturzähler), dazu
  QR-Code. Bei TSE-Ausfall fehlen TSE-Werte zulässigerweise (AEAO 1.14.3), der Ausfall ist auf dem Beleg
  ersichtlich (AEAO 1.14.2), Datum/Uhrzeit kommen vom System. Konform.
- AEAO zu § 146a Nr. 1.14: Ausfallzeiten und -grund automatisiert dokumentiert (`tse_stoerungen`),
  Weiterbetrieb, Belegausgabepflicht. Konform. Nr. 2.2.2 („unmittelbar mit Beginn ... starten") steht in
  bewusster Spannung zum asynchronen Modell; das ist die fixierte PRD-Entscheidung, die Doku beschreibt
  Mechanismus und Latenz ehrlich (compliance.md §3.8, verfahrensdokumentation.md §4). Siehe B4.
- fiskaly SIGN DE API v2: Pfade, PUT-Transaktions-Upsert (tx_revision 1/2, `schema.raw` mit Base64-processData),
  Antwortfelder (inkl. optionaler time_end/qr_code_data mit Fallbacks), Auth samt Token-Cache und einmaligem
  401-Refresh, Retry mit Retry-After (429/499/5xx). Fehlerklassifizierung je Status/Code weitgehend korrekt
  (400-Zustandscodes, 423, 503, 432 TSE-weit; 409 auftragsspezifisch, durch Healing gedeckt). Zwei Lücken:
  A1 und A4.
- Phase-7-Doku: handbuch.md §3.13, compliance.md §3.8, verfahrensdokumentation.md §4 (Herstellerdoku:
  Mechanismus, Ziel-Latenz p95 unter fünf Sekunden, Verzögerungsursachen), language.md (alle neuen Begriffe,
  Nachsignierung-Eintrag ersetzt). Deckt sich mit dem Code; eine Detailpräzisierung (A6).

## 3. Simplification (Skill-Schritt 2)

- `BuildKassenbelegProcessData` und `BuildBestellungProcessData` (faktorlose Wrapper) haben keine
  Produktionsnutzer mehr, nur Tests. Entfernt (A7).
- Nach A2 verliert `UpsertTSEKonfiguration` seinen letzten Produktionsnutzer in der Application-Schicht;
  Interface-Eintrag und öffentliche Repo-Methode entfernt (Teil von A2).
- Bewusst belassen: Der Worker fragt vor jedem Erstversuch den Ist-Zustand bei fiskaly ab (ein Retrieve auch im
  Happy Path). Kostet einen Request pro Signatur, hält aber genau einen Code-Pfad für Erst- und Heilungsversuch;
  bei Ziel-p95 unter fünf Sekunden unkritisch. Keine Änderung.
- Die dünnen Application-Wrapper (Logging + Fehler-Mapping) entsprechen dem Hausmuster; keine toten Endpunkte,
  keine Alt-Architektur-Reste gefunden. `GetTSESignaturByTxID` und die drei verteilten `tse_signing.go` sind
  ersatzlos weg, `tse_signaturen` existiert nicht mehr.

## 4. Repo-Verifikation (Skill-Schritt 3)

- Vor den Fixes: `make verify` grün (Backend-Unit, Frontend, Lint, Integrationstests inkl. DB).
- Nach den Fixes: siehe Umsetzungsstatus unten.
- Testabdeckung spiegelt die Plan-Akzeptanzkriterien: Sofort-Trigger und Polling-Fallback, Advisory-Lock
  (Integrationstest mit zweiter Session), Healing-Fälle, Client-Wiederverwendung, Durchlauf-Deadline,
  Störungs-Backoff samt Half-Open, Gift-Auftrag endet vor der Rückstandsschwelle, Watchdog öffnet auch bei
  hängendem Worker, Gate-Fälle (blockiert, Ausfall-Reste, Tag ohne TSE), Einrichtungs-Sweep, Statusguards,
  Signaturstatus tabellengetrieben, fiskalische Projektion tabellengetrieben, DSFinV-K-Mapper inkl.
  Nachsigniert- und Fehlerzeilen-Fall.

## Befunde Kategorie A (direkt umgesetzt)

- A1 Fehlende TSE-weite Klassifizierung von `E_CLIENT_NOT_FOUND`.
  Where: `backend/repository/tse_repo/fiskaly_client.go:180` (`tssZustandsCodes400`).
  Impact: fiskaly liefert bei nicht existenter client_id HTTP 400 mit Code `E_CLIENT_NOT_FOUND` (OpenAPI,
  PUT tx). Als AuftragsFehler klassifiziert sammelt jeder Auftrag drei Fehlversuche und wird endgültig
  `fehlgeschlagen`, statt dass der Worker in den Störungszustand geht; nach Korrektur der Konfiguration müsste
  der Admin alle Aufträge manuell zurücksetzen.
  Suggestion: Code in die Zustands-Map aufnehmen. Effort: S.
- A2 `UpdateTSEKonfiguration` umgeht Einrichtungs-Sweep und Störungsschluss.
  Where: `backend/api/settings/application/command.go:60` (nutzt `UpsertTSEKonfiguration` statt
  `SpeichereEinrichtung`).
  Impact: Der Endpunkt `update-tse-konfiguration` trägt auch den UI-Fluss „Konfiguration leeren" und „direkt
  speichern" (TSEKonfigurationSection). Führt er den Übergang unkonfiguriert → konfiguriert aus, bleibt ein
  offener `keine_konfiguration`-Störungszeitraum für immer aktiv (Worker schließt nur `tse_fehler`, Watchdog nur
  `rueckstand`): Jeder künftige offene Auftrag gälte im Beleg-Abruf bis zur Signatur als Ausfall „keine TSE
  konfiguriert" statt ausstehend, und das Störungsprotokoll zeigte dauerhaft eine aktive Störung.
  Suggestion: Auch dieser Pfad speichert über `SpeichereEinrichtung`; der Sweep+Schluss läuft dort nur beim
  Übergang zu einer vollständigen Konfiguration (neuer Guard `IstKonfiguriert`, damit das Leeren keinen
  Störungszeitraum fälschlich schließt). `UpsertTSEKonfiguration` aus Interface und Repo-Oberfläche entfernen.
  Effort: S.
- A3 `TSE_TA_START`/`TSE_TA_ENDE` ohne Millisekunden.
  Where: `backend/domain/dsfinvk/mapper.go:499` (`isoZeit`, RFC3339).
  Impact: Die amtliche Feldbeschreibung (DSFinV-K 2.4, Einzelaufzeichnungsmodul) verlangt für beide Felder
  explizit das Format `YYYY-MM-DDThh:mm:ss.fffZ`; jotti schrieb `...:04Z`. Formale Abweichung, die ein
  amtliches Prüftool bemängeln kann. BON_START/Z_ERSTELLUNG sind nicht betroffen (RFC3339 dort amtlich
  zulässig, Beispiele ohne Millisekunden).
  Suggestion: `isoZeit` mit `.000` formatieren (fiskaly liefert Sekundenauflösung); Mapper-Tests anpassen.
  Effort: S.
- A4 `RetrieveTransaction` deutet jedes 404 als „Transaktion nicht gefunden".
  Where: `backend/repository/tse_repo/fiskaly_client.go:292`.
  Impact: GET tx 404 trägt `E_TX_NOT_FOUND` oder `E_TSS_NOT_FOUND` (falsche TSS-ID). Letzteres führte zu einem
  unnötigen StartTransaction-Versuch pro Probe, bevor der TSE-weite Fehler greift; netto korrekt, aber
  semantisch falsch gemappt.
  Suggestion: Nur `E_TX_NOT_FOUND` (und leeren Code als defensiven Fallback) auf `ErrTransactionNichtGefunden`
  mappen; `E_TSS_NOT_FOUND` durchreichen. Effort: S.
- A5 Stale `tse-status`-Cache nach Konfigurations-Speichern/-Leeren.
  Where: `frontend/src/admin/settings/hooks.ts:99-107`.
  Impact: `saveTSEKonfiguration`/`clearTSEKonfiguration` invalidieren nur `tse-konfiguration`; der
  Dashboard-Konfigurationsalarm (`useTSEStatus`) bleibt bis zum Reload veraltet (die Einrichtungspfade
  invalidieren beide Keys bereits korrekt).
  Suggestion: `tse-status` mit invalidieren. Effort: S.
- A6 Doku-Präzisierung Störungsprotokoll-Schreiber.
  Where: `docs/handbuch.md:229`.
  Impact: Der Satz nennt „die Einrichtung (fehlende Konfiguration)" als Schreiber; im Code öffnet der Worker den
  `keine_konfiguration`-Zeitraum (beim endgültigen Markieren), die Einrichtung schließt ihn.
  Suggestion: Satz entsprechend präzisieren. Effort: S.
- A7 Tote Convenience-Wrapper.
  Where: `backend/domain/kasse/tse_processdata.go:15,71`.
  Impact: `BuildKassenbelegProcessData`/`BuildBestellungProcessData` (faktorlose Varianten) haben keine
  Produktionsnutzer; zwei Exporte ohne Mehrwert.
  Suggestion: Entfernen, Tests rufen die WithFaktor-Varianten mit Faktor 1. Effort: S.

## Befunde Kategorie B (Entscheidung nötig)

Details und Optionen in [audit-tse-signatur-outbox-entscheidungen.md](audit-tse-signatur-outbox-entscheidungen.md):

- B1 Verwerfen-Race: Signatur kann bei fiskaly entstehen, während der Admin den Auftrag verwirft (stiller
  Quittierungs-No-Op; TSE-TAR-Export und DSFinV-K weichen für diesen Vorgang ab).
- B2 Admin-Aktionen ohne Wirkungs-Rückmeldung (Status-Guards schlucken No-Ops, Antwort ist trotzdem 200).
- B3 Signaturauftrags-Liste hart auf 200 Einträge begrenzt, UI zeigt die Kappung nicht an.
- B4 AEAO 2.2.2 („unmittelbar mit Beginn"): bewusste Modell-Spannung; optional ein expliziter Satz in
  compliance.md §3.8 zur Transparenz gegenüber Prüfern.
- B5 Seed schreibt keine historischen Störungszeiträume: Demo zeigt nachsignierte Belege, aber ein leeres
  Störungsprotokoll.

## Priorisierte Empfehlungen

1. Korrektheit: A1 (falsch-endgültige Fehlschläge bei Konfigurationsfehler), A2 (ewiger Störungszeitraum,
   falsche Beleg-Vermerke) — beide umgesetzt.
2. Formale Konformität: A3 (amtliches Zeitformat) — umgesetzt.
3. Quick Wins: A4, A5, A6, A7 — umgesetzt.
4. Entscheidungen: B1 (empfohlen: RowsAffected loggen), B2 (empfohlen: 409 bei No-Op), B3 (empfohlen:
   Hinweistext), B4 (empfohlen: ein Satz in compliance.md), B5 (empfohlen: Störungszeiträume seeden) — je nach
   Gewichtung des Entwicklers.

## Umsetzungsstatus der A-Fixes

Alle sieben Fixes sind umgesetzt, `make verify` ist danach vollständig grün (Backend-Unit mit -race,
Frontend 172 Tests, Lint, DB-Integrationstests):

- A1 `E_CLIENT_NOT_FOUND` in `tssZustandsCodes400` aufgenommen; Klassifizierungstest erweitert.
- A2 `UpdateTSEKonfiguration` speichert über `SpeichereEinrichtung`; dort neuer Guard, der Sweep und
  Störungsschluss nur beim Übergang zu einer vollständigen Konfiguration ausführt (Leeren sweept nichts).
  `UpsertTSEKonfiguration` aus Application-Interface und Repo-Oberfläche entfernt. Neuer Repo-Test
  `TestSpeichereEinrichtung_LeereKonfigurationSweeptNicht`.
- A3 `isoZeit` schreibt TSE_TA_START/ENDE amtlich mit Millisekunden (`.000`); Mapper-Testerwartungen angepasst
  (BON_START/Z_ERSTELLUNG unverändert RFC3339).
- A4 `RetrieveTransaction` mappt nur noch `E_TX_NOT_FOUND` (bzw. codelose 404) auf
  `ErrTransactionNichtGefunden`; neuer Test für `E_TSS_NOT_FOUND`.
- A5 `saveTSEKonfiguration`/`clearTSEKonfiguration` invalidieren zusätzlich den `tse-status`-Query-Cache.
- A6 handbuch.md §3.13: Schreiber des `keine_konfiguration`-Zeitraums präzisiert (Worker öffnet, Einrichtung
  schließt).
- A7 Faktorlose processData-Wrapper entfernt; Tests rufen die WithFaktor-Varianten.

## Commit-Vorschlag

```
fix(tse): audit fixes for error taxonomy, config transitions and export format

Findings from the branch audit against DSFinV-K 2.4, KassenSichV, AEAO
zu 146a and the fiskaly SIGN DE v2 OpenAPI spec:

- classify fiskaly E_CLIENT_NOT_FOUND (400) as TSS-wide so a
  misconfigured client id triggers the worker's outage state instead of
  permanently failing every job
- route the credentials endpoint through SpeichereEinrichtung so an
  unconfigured-to-configured transition runs the setup sweep and closes
  the keine_konfiguration outage window; guard the sweep against saving
  an incomplete (cleared) configuration; drop the now-unused
  UpsertTSEKonfiguration from the command interface and repo surface
- format TSE_TA_START/ENDE with milliseconds as mandated by the
  DSFinV-K field specification (BON_START/Z_ERSTELLUNG stay RFC3339)
- map only E_TX_NOT_FOUND 404s to "transaction not found" in
  RetrieveTransaction; E_TSS_NOT_FOUND stays a TSS-wide error
- invalidate the tse-status query cache when saving or clearing the
  TSE configuration so the dashboard config alarm updates immediately
- document the actual keine_konfiguration writers in handbuch 3.13
- remove the unused factorless processData wrapper functions

Audit report: docs/plans/audit-tse-signatur-outbox-bericht.md
Open decisions: docs/plans/audit-tse-signatur-outbox-entscheidungen.md
```
