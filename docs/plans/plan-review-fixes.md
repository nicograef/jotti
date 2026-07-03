# Plan: Behebung der Findings aus dem Architektur-Review (2026-07)

> Quelle: Architektur- und Compliance-Review der Gesamtimplementierung
> (Event-Sourcing/Journal, TSE-Integration, DSFinV-K-Export, Security,
> Geld-/Steuerlogik). Die Phasen entsprechen den priorisierten
> Empfehlungen 1–8 des Reviews.

## Goal

Die bestätigten Fehler und Compliance-Lücken des Reviews werden behoben:
zwei Rechenfehler (Vorzeichen der Kassensturz-Differenz, USt-Aufschlüsselung
ohne Warenrücknahmen), eine Race Condition im OCC-Schreibpfad
(Doppel-Kassieren), die unvollständige TSE-Ausfalldokumentation im
DSFinV-K-Export, zwei Security-Schwächen im Auth-Pfad, die nicht-atomare
Kassensitzungs-Eröffnung, formale DSFinV-K-Abweichungen sowie die
TSE-processData-Vorzeichen bei Korrektur/Umbuchung. Nach Abschluss stimmen
Kassenbestand, Reporting, Belege und Export unter allen Fehler- und
Parallelitätsszenarien überein.

## Inventory (Findings → Code)

| #   | Finding                                                                  | Ort                                                                                      | Schwere  |
| --- | ------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------- | -------- |
| 1   | Differenzbuchung mit invertiertem Vorzeichen im Soll-Bestand; Abschluss-Retry eskaliert | `backend/sqlc/queries/kassensitzungen.sql` (GetKassenbestand), `backend/seed/engine.go`  | kritisch |
| 2   | `GetUmsatzProSteuersatz` ignoriert `stornierung-erteilt:v1`              | `backend/sqlc/queries/reporting.sql`                                                      | hoch     |
| 3   | OCC: `GetMaxVersion` erst beim Write → veraltete Validierung schreibt durch | `backend/api/table/application/command.go` (`writeEventOCC`), analog direktverkauf/kasse | kritisch |
| 4   | TSE-Ausfall nur bei Zahlung/Direktverkauf dokumentiert (Flag + Export)   | `backend/domain/kasse/tse_embedding.go`, `backend/domain/dsfinvk/mapper.go`               | hoch     |
| 5   | Rate-Limiter-Key aus Client-`X-Forwarded-For`; Einmalpasswort 6 Ziffern ohne Versuchszähler; Relay-Token `==`-Vergleich | `backend/api/middleware/middleware.go`, `backend/domain/user/password.go`, `backend/api/relay/http/handler.go` | hoch |
| 6   | Kassensitzungs-Eröffnung nicht atomar; kein Status-Guard beim Event-Write | `backend/api/kasse/application/command.go`, `backend/repository/kassenjournal_repo/repo.go` | mittel |
| 7   | DSFinV-K-Formalia: eigene `index.xml`, fehlende Leerdateien, 5 statt 2 Zertifikatsspalten, `TSE_ZEITFORMAT` hartkodiert `utcTime` | `backend/domain/dsfinvk/` (index.go, dsfinvk.go, mapper.go)                              | mittel   |
| 8   | Korrektur/Umbuchung mit positiven Mengen signiert; nachsignierte Belege ohne Kennzeichnung; Betriebstag-Datum UTC-truncated; EuroInput verwirft `.` | `backend/api/table/application/tse_signing.go`, `backend/api/tse/application/processdata.go`, `backend/api/table/application/kassenbeleg_command.go`, `backend/api/kasse/application/command.go`, `frontend/src/components/common/EuroInput.tsx` | mittel |

## Architectural decisions

Durable decisions, die über alle Phasen gelten:

- **Vorzeichen-Konvention Differenz.** `differenzCents` bleibt überall
  Soll − Ist (Event-Payload unverändert). Die Bargeldwirkung ist Ist − Soll;
  jede Stelle, die den Kassenbestand fortschreibt, **subtrahiert** deshalb
  `differenzCents`. TSE-processData (`tse_signing.go`) und DSFinV-K-Mapper
  machen das bereits richtig und bleiben unverändert — nur die
  Bestandsrechnung wird angeglichen.
- **OCC gegen den gelesenen Zustand.** Die erwartete Version kommt aus dem
  Zustand, gegen den validiert wurde (Projektion `TischSession.LastEventVersion`
  bzw. max. Version des Replays), nie aus einem frischen `GetMaxVersion` zum
  Schreibzeitpunkt. Der `UNIQUE(subject, version)`-Constraint bleibt der
  Konfliktdetektor; ein Konflikt bedeutet dann tatsächlich „Zustand hat sich
  seit dem Lesen geändert“ → 409, Client lädt neu. Multi-Event-Writes
  (Storno-Aufteilung, Umbuchung) vergeben fortlaufende Versionen ab derselben
  gelesenen Basis.
- **Ausfall-Ableitung statt Typ-Flags.** Der DSFinV-K-Export leitet den
  TSE-Ausfall generisch ab: Jeder fiskalische Bonkopf erhält genau eine Zeile
  in `transactions_tse.csv` — mit Signatur (aus Event oder Nachsignier-
  Seitentabelle) oder als `TSE_TA_FEHLER`-Zeile. Die per-Event-Flags
  (`TSEAusfall`) bleiben für die Beleg-Darstellung erhalten und werden auf
  alle signierpflichtigen Event-Typen ausgeweitet.
- **Kassensitzungs-Status als Sperrpunkt.** Jeder Event-Write prüft den
  Status der Kassensitzung **innerhalb** der Event-Transaktion mit
  Zeilensperre (`FOR SHARE` auf `kassensitzungen`); der Tagesabschluss
  wechselt den Status in derselben Transaktion wie sein Event unter
  `FOR UPDATE`. Damit sind „Write in geschlossene Sitzung“ und
  „Abschluss während laufendem Write“ serialisiert.
- **DSFinV-K-Formalia folgen der amtlichen Vorlage.** Es wird die amtliche,
  unveränderte `index.xml` der DSFinV-K v2.4 ausgeliefert (alle Tabellen
  deklariert); nicht befüllte Tabellen als Header-only-CSV. Eigene
  Erweiterungen (5 Zertifikatsspalten) entfallen zugunsten des amtlichen
  Schemas. Autoritative Quelle: `docs/rechtsquellen/` (DSFinV-K v2.4).
- **Aktive Entwicklungsphase (AGENTS.md).** Event-Strukturen und DB-Schema
  werden direkt geändert (`database/migrations/01_initial.up.sql`), keine
  neuen Migrationsdateien, kein Dual-Read. Nach Query-Änderungen `make sqlc`.

## Resolved decisions

- **Reihenfolge/Abhängigkeiten:** Phasen 1, 2, 5, 8 sind unabhängig
  voneinander. Phase 4 vor Phase 7 (beide ändern den Mapper). Phase 3 vor
  Phase 6 (beide ändern den Write-Pfad).
- **Rate-Limiter hinter Proxy:** jotti läuft produktiv immer hinter dem
  eigenen Caddy (docker-compose). Caddy hängt die echte Client-IP als
  **letzten** Eintrag an `X-Forwarded-For` an → der Limiter nimmt den
  letzten (rechtesten) Eintrag, Fallback `RemoteAddr`. Kein
  Trusted-Proxy-Konfigurationsschalter (YAGNI für das Deployment-Modell).
- **Einmalpasswort:** 8 Zeichen alphanumerisch ohne verwechselbare Zeichen
  (kein `0/O`, `1/l/I`), Erzeugung per Rejection-Sampling (kein Modulo-Bias).
  Harter Fehlversuchszähler pro Benutzer: nach 5 Fehlversuchen wird das
  Einmalpasswort ungültig, der Admin muss ein neues erzeugen.
- **BON-Vergabe und Nachsignieren bleiben wie designed:** BON_NR wird
  weiterhin deterministisch beim Export vergeben; der Nachsignier-Worker
  (echte Zeitstempel, Auftragstabelle als Ausfalldoku) bleibt — es kommt nur
  die Kennzeichnung auf dem Beleg hinzu (Phase 8).

## Open questions / Risks

- **JWT-Revocation** (deaktivierte Benutzer behalten bis 12 h Zugriff) ist
  ein Review-Finding, aber **nicht Teil der Empfehlungen 1–8** → bewusst out
  of scope; separat entscheiden (Token-Version-Claim vs. kurze Expiry+Refresh).
- **Zertifikate > 2000 Zeichen:** `TSE_ZERTIFIKAT_I/II` fassen amtlich je
  1000 Zeichen. Falls das fiskaly-Zertifikat länger ist, gegen
  `docs/rechtsquellen/` prüfen, wie die DSFinV-K Überlänge vorsieht
  (Verweis auf TSE-Export vs. Kürzung); Entscheidung in Phase 7 am Spec-Text.
- **Negative Mengen in `Bestellung-V1`:** DSFinV-K Anhang I sieht sie vor;
  ob die fiskaly-API sie anstandslos signiert, ist vor dem Merge gegen die
  fiskaly-TEST-Umgebung zu verifizieren (Phase 8). Fallback: Kennzeichnung
  über Positionstext ist **keine** Option — dann Rücksprache.
- **OCC-Umstellung** berührt alle Command-Pfade (Tisch, Direktverkauf,
  Kassensitzung). Risiko von Regressionen → `make verify` (inkl.
  Integrationstests) ist Pflicht, plus neue Nebenläufigkeitstests.
- **Globale statt per-IP-Drosselung** darf das Onboarding (30 Helfer setzen
  vor dem Fest gleichzeitig Passwörter) nicht blockieren → Limit nach dem
  Umbau mit realistischem Szenario nachmessen.

---

## Phase 1: Vorzeichen der Kassensturz-Differenz (Empfehlung 1)

### Context

`GetKassenbestand` addiert `kj_extract_differenz_cents` (= Soll − Ist) auf
den Soll-Bestand (`backend/sqlc/queries/kassensitzungen.sql`), die eigene
Doku im Code sagt das Gegenteil (`backend/api/kasse/application/tse_signing.go`:
„Die tatsächliche Bargeldbewegung ist Ist − Soll“). Nach einer
Differenzbuchung entfernt sich das Buch-Soll vom Ist; der explizit
wiederholbare `KasseAbschliessen`-Pfad bucht die Differenz beim Retry
verdoppelt erneut — als TSE-signiertes Event im unveränderbaren Journal.
`backend/seed/engine.go` repliziert die falsche Rechnung.

### What to build

- In `GetKassenbestand` das Vorzeichen drehen: `- kj_extract_differenz_cents`
  statt `+`. Kommentar der Query um die Konvention (Soll − Ist, Bargeldwirkung
  negativ) ergänzen. `make sqlc`.
- Dieselbe Korrektur in der Bestandsrechnung von `backend/seed/engine.go`.
- Extract-Funktion `kj_extract_differenz_cents` bleibt unverändert (liefert
  rohes `betragCents`); TSE-Signierung und DSFinV-K-Mapper bleiben unverändert.

### Acceptance criteria

- Unit-/Integrationstest: Soll 100 €, Ist 90 € → nach Kassensturz +
  Differenzbuchung ist der Soll-Bestand 90 € (= Ist).
- Idempotenz-Test: zweiter `KasseAbschliessen`-Durchlauf nach Teilfehler
  berechnet Differenz 0 und bucht **kein** zweites Differenz-Event.
- Test prüft das Vorzeichen des Bestands (nicht nur die Event-Reihenfolge).
- Bestand, DSFinV-K-`cash_per_currency.csv` und TSE-processData der
  Differenz sind für denselben Fall konsistent (ein gemeinsamer Testfall).

## Phase 2: Warenrücknahmen in der USt-Aufschlüsselung (Empfehlung 2)

### Context

`GetUmsatzProSteuersatz` (`backend/sqlc/queries/reporting.sql`) filtert auf
drei Event-Typen; `stornierung-erteilt:v1` fehlt, obwohl
`kj_extract_umsatz_pro_steuersatz` (`database/migrations/01_initial.up.sql`)
den Typ ausdrücklich mit Faktor −1 unterstützt. Nach einer Warenrücknahme
ist die Brutto/Netto/USt-Aufschlüsselung der Tagesabrechnung zu hoch und
divergiert von `gesamt_umsatz_cents` und vom DSFinV-K-Export.

### What to build

- `'stornierung-erteilt:v1'` in die `WHERE type IN (…)`-Liste von
  `GetUmsatzProSteuersatz` aufnehmen. `make sqlc`.

### Acceptance criteria

- Test: Zahlung 19 % über 10 €, danach Warenrücknahme 4 € → Aufschlüsselung
  zeigt 6 € bei 19 %; Summe über alle Steuersätze == `gesamt_umsatz_cents`.
- Invarianten-Test im Reporting: Σ(Umsatz pro Steuersatz) ==
  Gesamtumsatz für jede Kassensitzung mit gemischten Storno-Arten
  (Warenrücknahme kassenwirksam, Korrektur geldneutral).

## Phase 3: OCC gegen den gelesenen Zustand (Empfehlung 3)

### Context

`writeEventOCC` (`backend/api/table/application/command.go`) holt
`GetMaxVersion` erst nach Validierung **und** TSE-Signierung (bis 5 s
Netzwerk-Latenz). Ein paralleler Commit im Zeitfenster erhöht die
Maximalversion → kein Unique-Konflikt, das zweite Event wird trotz
veralteter Validierung geschrieben: Doppel-Kassieren (Saldo negativ,
`reduceByPosition` kappt stumm), Doppel-Warenrücknahme (doppelte
Bar-Auszahlung), analog Umbuchung und Direktverkauf-Storno.

### What to build

- `writeEventOCC` nimmt eine erwartete Version als Parameter
  (`e.Version = expected + 1`) statt `GetMaxVersion` zu rufen.
- Tisch-Pfade: `loadTischState` liefert die Projektion bereits —
  `state.LastEventVersion` als erwartete Version durch alle Commands
  reichen (`BestellungAufnehmen`, `AusgabeBestaetigen`, `ZahlungKassieren`,
  `StornierungErteilen`/`persistStornoEvents`, `BestellungUmbuchen`).
  Multi-Event-Writes vergeben `expected+1 … expected+n`; die Umbuchung
  nutzt je Stream die jeweils gelesene Version.
- Direktverkauf-Storno: erwartete Version = max. Version des ohnehin
  durchgeführten Stream-Replays (`backend/api/direktverkauf/application/command.go`).
- Kassensitzungs-Commands: erwartete Version aus dem In-Memory-Replay der
  KS-Events, gegen das validiert wird.
- `GetMaxVersion` aus dem Schreibpfad entfernen (Repo-Funktion nur behalten,
  falls anderweitig genutzt).
- Defense in depth: `reduceByPosition`
  (`backend/domain/kasse/tisch_session.go`) meldet Überreduktion als Fehler
  statt still zu kappen.

### Acceptance criteria

- Nebenläufigkeitstest (Integrationstest): zwei parallele
  `ZahlungKassieren` für dieselben Positionen → genau eine Zahlung
  persistiert, der zweite Request erhält 409.
- Gleicher Test für Warenrücknahme und Umbuchung.
- Kein Pfad kann den Tisch-Saldo negativ machen; Test auf Saldo ≥ 0 nach
  parallelen Writes.
- `make verify` grün; bestehende OCC-Tests (409 bei Konflikt) unverändert grün.

## Phase 4: TSE-Ausfalldokumentation für alle Vorgangsarten (Empfehlung 4)

### Context

Das `TSEAusfall`-Flag existiert nur in `ZahlungKassiert` und
`DirektverkaufGetaetigt` (`backend/domain/kasse/tse_embedding.go`); der
Mapper schreibt `TSE_TA_FEHLER`-Zeilen nur dort
(`backend/domain/dsfinvk/mapper.go`, `buildTransactionsTSE`). Unsignierte
Stornierungen, Korrekturen, Umbuchungen, Bestellungen, Geldtransit- und
Differenzbuchungen erscheinen ohne jede TSE-Zeile im Export — die
Ausfalldokumentation (§ 146a-Umfeld, AEAO 1.14) ist für diese Typen nicht
erfüllt. Der Anfangsbestand (`kassensitzung-eroeffnet`) wird als Beleg
exportiert, aber nie signiert.

### What to build

- `TSEAusfall`-Flag und `TSETxID` auf alle signierpflichtigen Event-Typen
  ausweiten (`tse_embedding.go`): Bestellung, Korrektur, Umbuchung,
  Warenrücknahme, Direktverkauf-Storno, Geldtransit, Differenz,
  Tagesabschluss.
- Mapper: Ausfall generisch ableiten — jeder fiskalische Bonkopf erhält
  genau eine `transactions_tse.csv`-Zeile: Signatur aus Event, sonst aus
  der Nachsignier-Seitentabelle, sonst `TSE_TA_FEHLER = "TSE-Ausfall"`.
- Anfangsbestand signieren: `kassensitzung-eroeffnet:v1` erhält eine
  `Kassenbeleg-V1`-Signierung (Eigenbeleg, analog Geldtransit) samt
  Embed-Funktion und Nachsignier-Pfad; der Export übernimmt die TSE-Daten.
- Die Test-Invariante „jeder Bonkopf hat eine TSE-Zeile“
  (`mapper_test.go`) auf **alle** Bonkopf-Typen ausweiten.

### Acceptance criteria

- Export-Test: unsignierte Warenrücknahme/Korrektur/Umbuchung/Geldtransit
  → `transactions_tse.csv`-Zeile mit `TSE_TA_FEHLER`, keine Bonkopf-Zeile
  ohne TSE-Zeile (Invariante über alle Typen).
- Nach erfolgreicher Nachsignierung ersetzt die Signatur die Fehler-Zeile
  beim nächsten Export.
- `kassensitzung-eroeffnet` mit Anfangsbestand > 0 trägt TSE-Daten im Event
  und im Export; bei TSE-Ausfall entsteht ein Nachsignier-Auftrag.

## Phase 5: Security-Härtung Auth-Pfad (Empfehlung 5)

### Context

Der Rate-Limiter-Key stammt aus dem Client-kontrollierbaren
`X-Forwarded-For`-Header (`backend/api/middleware/middleware.go`) — Caddy
hängt die echte IP nur an, der Header ist frei variierbar → das
5-rps-Limit auf `/auth` ist wirkungslos. Das Einmalpasswort hat 6 Ziffern
(10⁶ Kombinationen, `backend/domain/user/password.go`), bleibt unbegrenzt
gültig und kennt keinen Fehlversuchszähler. Der Relay-Token wird mit `==`
verglichen (`backend/api/relay/http/handler.go`).

### What to build

- Limiter-Key: letzten (rechtesten) `X-Forwarded-For`-Eintrag verwenden,
  Fallback `RemoteAddr`; Header-Parsing mit Tests.
- Einmalpasswort: 8 Zeichen alphanumerisch ohne verwechselbare Zeichen,
  Rejection-Sampling statt Modulo. Fehlversuchszähler pro Benutzer
  (Spalte in `users`, Schema in `01_initial.up.sql`): nach 5 Fehlversuchen
  Einmalpasswort invalidieren; Admin-Reset erzeugt ein neues und setzt den
  Zähler zurück. Fehlermeldung unterscheidet „ungültig“ nicht von
  „gesperrt“ erst nach Sperre (deutsche Meldung: neues Einmalpasswort beim
  Admin anfordern).
- Relay-Token-Vergleich auf `subtle.ConstantTimeCompare` umstellen;
  `/relay/*` in das bestehende Rate-Limiting aufnehmen.

### Acceptance criteria

- Test: variierender `X-Forwarded-For` mit gleicher echter Quelle wird auf
  denselben Limiter-Key abgebildet.
- Test: 5 falsche Einmalpasswort-Versuche → 6. Versuch schlägt auch mit
  korrektem Einmalpasswort fehl, bis der Admin zurücksetzt.
- Onboarding-Szenario: 30 Benutzer setzen nacheinander Passwörter über
  dieselbe Proxy-IP, ohne ausgesperrt zu werden (Limit-Kalibrierung).
- Frontend zeigt die neuen 8-stelligen Einmalpasswörter korrekt an
  (Anzeige/Copy im Admin, `input-otp`-Länge).

## Phase 6: Kassensitzungs-Konsistenz (Empfehlung 6)

### Context

`InsertKassensitzung` (CRUD, Status `offen`) und das
`kassensitzung-eroeffnet:v1`-Event laufen in getrennten Transaktionen
(`backend/api/kasse/application/command.go`) — schlägt der Event-Write
fehl, existiert eine offene Sitzung ohne Eröffnungs-Event (Anfangsbestand
fehlt dauerhaft, kein Repair-Pfad). Außerdem prüft der Event-Write-Pfad
den Sitzungsstatus nicht: ein paralleler Request kann nach dem
Tagesabschluss noch in die geschlossene Sitzung schreiben (anderes
Subject → kein OCC-Konflikt), der Z-Bon divergiert dann vom Journal.

### What to build

- Atomare Eröffnung: neue Repo-Methode, die `INSERT kassensitzungen` und
  den Event-Write in **einer** Transaktion ausführt (inkl. TSE-Signierung
  vorab wie in allen anderen Pfaden; Nachsignier-Auftrag in derselben Tx).
- Status-Guard: `writeEventInTx` (`backend/repository/kassenjournal_repo/repo.go`)
  liest innerhalb der Event-Transaktion
  `SELECT status FROM kassensitzungen WHERE z_nr = $1 FOR SHARE` und bricht
  mit einem eigenen Fehler ab, wenn der Status nicht `offen` ist
  (HTTP 409 im Handler).
- Tagesabschluss: Statuswechsel `offen → abgeschlossen` in derselben
  Transaktion wie das `tagesabschluss-erstellt:v1`-Event, Zeile mit
  `FOR UPDATE` gesperrt — parallele Writes warten und scheitern dann am
  Status-Guard.

### Acceptance criteria

- Integrationstest: erzwungener Event-Write-Fehler bei der Eröffnung
  hinterlässt **keine** offene Sitzung (Rollback beider Schritte).
- Integrationstest: `BestellungAufnehmen` parallel zum Tagesabschluss →
  entweder vor dem Abschluss committed (und vom Saldo-0-Check erfasst)
  oder 409; nie ein Event nach `tagesabschluss-erstellt:v1` in der Sitzung.
- Kein Event mit `kassensitzung_nr` einer abgeschlossenen Sitzung
  schreibbar (Test direkt auf Repo-Ebene).

## Phase 7: DSFinV-K-Formalia (Empfehlung 7)

### Context

Der Export nutzt eine selbstgebaute `index.xml`, lässt `slaves.csv`,
`pa.csv`, `itemamounts.csv`, `subitems.csv` ganz weg, erweitert `tse.csv`
auf fünf Zertifikatsspalten und deklariert `TSE_ZEITFORMAT` hartkodiert
als `utcTime`, obwohl fiskaly `unixTime` liefert
(`TSEStammdaten.LogTimeFormat`) und die Zeiten als RFC3339 ausgegeben
werden. Prüftools validieren gegen die amtliche `index.xml`. Kleinere
Punkte aus dem Review: leere `Numeric`-Felder (FAKTOR, TSE_TANR in
Fehlerzeilen), `vat.csv` nur mit verwendeten Schlüsseln, `GV_TYP = Umsatz`
auf geldneutralen `AVBestellung`-Positionen.

### What to build

- Amtliche `index.xml` der DSFinV-K v2.4 (aus `docs/rechtsquellen/`)
  unverändert als eingebettetes Asset ausliefern; den eigenen
  `index.xml`-Generator entfernen oder auf die amtliche Datei umstellen.
- Nicht befüllte Tabellen (`slaves.csv`, `pa.csv`, `itemamounts.csv`,
  `subitems.csv`) als Header-only-CSV erzeugen.
- `tse.csv` auf die amtlichen Spalten zurückführen
  (`TSE_ZERTIFIKAT_I/II`); Umgang mit Überlänge am Spec-Text in
  `docs/rechtsquellen/` entscheiden (→ Open questions).
- `TSE_ZEITFORMAT` aus `TSEStammdaten.LogTimeFormat` übernehmen;
  `TSE_TA_START`/`TSE_TA_ENDE` im deklarierten Format ausgeben
  (`unixTime` → Epochensekunden).
- `vat.csv` mit den amtlich vordefinierten Schlüsseln 1–7 ausgeben.
- `GV_TYP` auf `AVBestellung`-Positionszeilen leeren; leere
  `Numeric`-Felder gegen die DTD-Vorgaben prüfen und konsistent füllen
  (z. B. `FAKTOR` weglassen oder `1.000`).

### Acceptance criteria

- ZIP enthält alle amtlich deklarierten Dateien; `index.xml` ist
  byte-identisch mit der amtlichen Vorlage.
- Golden-File-Tests aktualisiert; Invariante „jede in `index.xml`
  deklarierte Tabelle liegt als Datei bei“ als Test.
- Manuelle Validierung eines Beispiel-Exports mit einem externen
  DSFinV-K-Prüftool (z. B. DFKA-Tool) ohne Struktur-Fehler; Ergebnis im PR
  dokumentiert.
- `TSE_TA_START/ENDE` einer fiskaly-Signatur sind als Epochensekunden
  exportiert und `TSE_ZEITFORMAT` = `unixTime` (Test mit Seed-Daten).

## Phase 8: TSE-Vorzeichen, Beleg-Kennzeichnung, Datums- und Eingabe-Fixes (Empfehlung 8)

### Context

Geldneutrale Korrekturen und Umbuchungen werden mit **positiven** Mengen
als `Bestellung-V1` signiert (`backend/api/table/application/tse_signing.go`)
— TSE-seitig nicht von Neubestellungen unterscheidbar und im Widerspruch
zum eigenen Export (negative Darstellung); der processData-Builder
verwirft Mengen ≤ 0 stillschweigend
(`backend/api/tse/application/processdata.go`). Nach erfolgreicher
Nachsignierung druckt der Beleg TSE-Daten ohne Hinweis
(`kassenbeleg_command.go`). `time.Now().In(berlin).Truncate(24h)`
schneidet auf UTC-Mitternacht → falsches Betriebstagsdatum zwischen 00:00
und ~02:00 Uhr. `EuroInput` verwirft `.` als Dezimaltrenner
(`"4.5"` → 45,00 €).

### What to build

- **Vorzeichen:** `BuildBestellungProcessData` akzeptiert negative Mengen;
  Korrektur signiert alle zurückgenommenen Positionen mit negativer Menge,
  Umbuchung signiert den Quell-Abgang negativ und den Ziel-Zugang positiv.
  Verifikation gegen die fiskaly-TEST-Umgebung (→ Open questions).
- **Beleg-Kennzeichnung:** Stammt die Signatur aus der
  Nachsignier-Seitentabelle, druckt der Kassenbeleg eine Zusatzzeile
  („TSE nachsigniert am <logTime Ende>“); `docs/verfahrensdokumentation.md`
  um den Nachsignier-Ablauf ergänzen.
- **Betriebstag:** Wandkalenderdatum in Europe/Berlin bestimmen
  (`y, m, d := now.In(berlin).Date()`), kein `Truncate` auf UTC-Grenzen.
- **EuroInput:** `.` als Dezimaltrenner akzeptieren (letztes `.`/`,` im
  String ist der Trenner); mehrdeutige Eingaben (`1,2,3`) als ungültig
  markieren statt still zu parsen; Tests für `4.5`, `4,5`, `1.234,56`,
  `1,2,3`.
- **Doku-Drift (Beifang):** `auszahlung-geleistet:v1` aus
  `.github/instructions/event-sourcing.instructions.md` entfernen,
  `bestellung-korrigiert/umgebucht` und den StreamType `direktverkauf`
  ergänzen; nginx-Kommentar in `backend/app/app.go` auf Caddy korrigieren.

### Acceptance criteria

- processData-Tests: Korrektur erzeugt Zeilen mit negativer Menge im
  Anhang-I-Format; Umbuchung Quelle negativ / Ziel positiv; Signierung
  gegen fiskaly-TEST erfolgreich (dokumentierter Live-Test, analog
  `fiskaly_client_live_test.go`).
- Beleg-Test: nachsignierter Vorgang druckt die Kennzeichnungszeile;
  regulär signierter Vorgang unverändert.
- Datums-Test: Eröffnung 00:30 Europe/Berlin (Sommer- und Winterzeit)
  ergibt das Kalenderdatum des laufenden Tages.
- `EuroInput.test.tsx` deckt Punkt-, Komma- und Mischeingaben ab;
  `4.5` ergibt 450 Cent.
- `make check` grün; Doku-Referenzen (`event-sourcing.instructions.md`)
  stimmen mit dem Code überein.
