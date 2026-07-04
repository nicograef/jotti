# Plan: TSE-Signatur über Outbox und Signatur-Worker

> Source PRD: [docs/prds/prd-tse-signatur-outbox.md](../prds/prd-tse-signatur-outbox.md)

## Goal

Die TSE-Signierung wird vollständig vom Kassier-Pfad entkoppelt: Jeder signaturpflichtige Vorgang schreibt im selben Commit einen Signaturauftrag, ein Worker ist der einzige Sprecher für Signaturtransaktionen, alle Signaturen liegen in der Seitentabelle. Der heutige Nachsignier-Pfad wird vom Ausnahme- zum Normalfall befördert. Buchen blockiert nie auf die TSE, Beleg und Export lesen genau eine Signaturquelle, ein Störungsprotokoll ersetzt das eingefrorene Ausfall-Flag, und der Kassenabschluss prüft ein Gate mit Ausfall-Reste-Regel.

## Architectural decisions

Dauerhafte Entscheidungen, die für alle Phasen gelten:

- **Schema**:
  - `tse_signaturauftraege` (umbenannt aus `tse_nachsignier_auftraege`): zusätzlich `event_id` (FK auf `kassenjournal`, NOT NULL, UNIQUE — genau ein Auftrag je Event), `tx_id` bleibt zufällige UUID (Schema-Kommentar „deterministisch" wird korrigiert), `process_type`/`process_data` als Snapshot beim Einreihen. Status-CHECK: `offen`, `erledigt`, `fehlgeschlagen`, `verworfen`, `tse_nicht_konfiguriert`. Verwerfen ist protokollierter Statuswechsel: Felder für Grund, Benutzer und Zeitpunkt. Kein DELETE (GoBD).
  - `tse_signaturen` bleibt unverändert der einzige Signatur-Store (Schlüssel `tx_id`); nur der Tabellen-Kommentar ändert sich (kein „Happy Path im Event" mehr).
  - `tse_stoerungen` (neu, Störungsprotokoll): je Zeitraum `beginn`, `ende` (NULL solange aktiv), Grund-Art (`tse_fehler`, `rueckstand`, `keine_konfiguration`) plus Fehlertext. Aufbewahrungspflichtig, kein DELETE.
  - Event-Payloads verlieren sämtliche TSE-Felder (`tseTxId`, TSEData-Block, `tseAusfall`).
- **Leseweg**: Event → Auftrag (über `event_id`) → Signatur (über `tx_id`). Beleg, Export und Signaturstatus kennen keine zweite Quelle.
- **Komponenten**:
  - *Fiskalische Projektion*: eine zentrale Funktion Event → (signaturpflichtig, processType, processData), auch datenabhängig (Sitzungseröffnung nur bei Anfangsbestand > 0). Ersetzt die drei `tse_signing.go`. Plausibilisiert processData beim Einreihen (Schema, Vorzeichen, Summen) und protokolliert Verstöße, ohne zu blockieren.
  - *Signatur-Worker* (in `backend/app`): einziger Sprecher für Start/Finish/Ist-Abfrage. Sofort-Trigger nach Commit über eine In-Process-Benachrichtigung (non-blocking, gepufferter Kanal), der Polling-Tick bleibt Fallback. FIFO als Soll-Eigenschaft. Advisory Lock beim Start; bekommt eine zweite Instanz den Lock nicht, wartet der Worker mit periodischem Neuversuch und deutlicher Error-Log-Warnung (kein Fail-Fast der App).
  - *Störungsprotokoll-Komponente*: idempotentes Öffnen/Schließen von Zeiträumen. Schreiber: Worker (TSE-weiter Fehler / erste erfolgreiche Signatur), Leser-Pfade (Rückstands-Schwelle über-/unterschritten — Leser materialisieren, damit auch ein hängender Worker dokumentiert wird), Settings (fehlende Konfiguration / Einrichtung).
  - *Signaturstatus-Funktion*: ohne Zeitverhalten, genau vier Ergebnisse (vorhanden, vorhanden+nachsigniert, Ausfall mit Grund, ausstehend).
- **Routen** (POST-only, deutsche Fachsprache):
  - `/service/beleg-drucken` bleibt; Antwort erhält ein Status-Feld (`eingereiht` | `ausstehend`). Bei `ausstehend` wird kein Druckauftrag angelegt; die UI ruft denselben Endpunkt erneut auf (alle 1–2 s für rund 10 s, danach auf Anforderung).
  - `/admin/kasse-abschliessen` bleibt; das Gate antwortet bei Blockade mit 409 und strukturierten Details (Anzahl offener Aufträge, Alter des ältesten, Ausfall-Reste). Kein Vorab-Prüf-Endpunkt.
  - Die `nachsignier`-Endpunktfamilie wird zur `signaturauftrag`-Familie (Liste, Zurücksetzen einzeln/gesamt, Verwerfen mit Begründung); Störungsprotokoll und Queue-Zustand erhalten eigene Admin-Lesewege (genaue Namen in Phase 6).
- **Statusmaschine**: Worker quittiert `offen → erledigt`, zählt auftragsspezifische Fehlversuche (Backoff ~5/15/45 s, drei Versuche) bis `fehlgeschlagen`, markiert `offen → tse_nicht_konfiguriert` (endgültig). Admin setzt `fehlgeschlagen` und `tse_nicht_konfiguriert` auf `offen` zurück (einzeln/gesamt), verwirft `offen` und `fehlgeschlagen` mit Begründung. TSE-weite Fehler zählen nie auf den Auftrag, sondern schalten den Worker in den Störungszustand (eigener Backoff mit Minuten-Deckel, Half-Open-Probe).
- **Deployment-Annahme**: Single-Prozess; Scale-out (LISTEN/NOTIFY, FOR UPDATE SKIP LOCKED) ist bewusst außen vor.

## Inventory

Heutiger Sync-Pfad und Nachsignier-Infrastruktur:

- `backend/api/tse/application/signing.go:51-159` — synchroner `Signierer` mit Deadline im Kassier-Request; entfällt.
- `backend/api/kasse/application/tse_signing.go:17-42`, `backend/api/table/application/tse_signing.go` (67 Z.), `backend/api/direktverkauf/application/tse_signing.go` (39 Z.) — die drei verteilten Signierhelfer; gehen in der fiskalischen Projektion auf.
- `backend/api/tse/application/processdata.go:16-154` — processData-Builder (Kassenbeleg, Bestellung, Eigenbeleg, Tagesabschluss, Faktor-Varianten); Grundlage der Projektion, tabellengetriebene Tests in `processdata_test.go`.
- `backend/domain/kasse/tse_embedding.go:46-103` und `backend/domain/kasse/tse_data.go` — TSE-Felder in Event-Payloads; entfallen.
- `backend/repository/kassenjournal_repo/repo.go:57-276` — Methoden-Kombinatorik der Event-Writes (mit/ohne Nachsignier-Auftrag, mit/ohne Druckaufträge, atomare Mehrfach-Writes, `EroeffneKassensitzung` mit TSE-Aufruf im DB-Callback); `repo.go:580-601` — `GetTSESignaturByTxID`.
- `backend/app/tse_nachsignier_worker.go` — Worker mit Polling (5 s), Healing (`beschaffeSignatur`, Z. 156–185) und atomarer Quittierung; Vorbild-Tests in `tse_nachsignier_worker_test.go` (Fake-TSE-Client). Start in `backend/app/app.go:86-87`.
- `backend/repository/tse_repo/repo.go` — Auftrags-Store: `MaxNachsignierVersuche = 10` mit Minuten-Backoff (Z. 14–19, entfällt), atomare Quittierung (Z. 99–122), Status-Guards für Zurücksetzen/Verwerfen.
- `backend/repository/tse_repo/fiskaly_client.go` — fiskaly-Client (Start/Finish/Retrieve); bekommt die Fehlertaxonomie.
- `database/migrations/01_initial.up.sql:477-529` — `tse_nachsignier_auftraege` (mit veraltetem „deterministisch"-Kommentar, Z. 495) und `tse_signaturen`.
- `backend/sqlc/queries/tse_nachsignier.sql`, `tse_signaturen.sql` — Queries, nach Umbau `make sqlc`.

Leser:

- `backend/api/table/application/kassenbeleg_command.go:138-174` — `tseAbschnittFuerBeleg`: heutige Zwei-Quellen-Merge-Logik; `KassenbelegDrucken` (Z. 200–435) mit `Nachsigniert`-Kennzeichen und `TSEAusfallvermerk` im ESC/POS-Formatter.
- `backend/domain/dsfinvk/mapper.go:21-27` (SignaturNachladen), `:42-49` (`TSE-Ausfall`-Fehlertext), `:476-481` (`tseVorgangsart`) — Export-Merge über zwei Quellen.
- `backend/api/service.go:68` — `/service/beleg-drucken`.

Abschluss, Admin, Settings:

- `backend/api/kasse/application/command.go:301-436` — `KasseAbschliessen` (Barriere, Kassensturz, Differenzbuchung, Tagesabschluss); Route `backend/api/admin.go:132`.
- `backend/api/admin.go:156-162` — Nachsignier-Admin-Routen; `backend/api/tse/application/command.go`, `query.go`, `backend/api/tse/http/handler.go`.
- `backend/api/settings/application/setup.go:310-331` — `speichereEinrichtung`: gemeinsamer Speicher-Schritt aller Einrichtungspfade (Ort des Einrichtungs-Sweeps); `backend/api/settings/application/command.go:33` — reiner Zugangsdaten-Wechsel.
- `backend/api/settings/http/query_handler.go:62` — `offeneNachsignierungen` im TSE-Status.

Frontend:

- `frontend/src/service/table/TischBackend.ts:68-72`, `frontend/src/service/direktverkauf/DirektverkaufBackend.ts:48-52` — Beleg-Abruf; UI in `TischHistorie.tsx` und `DirektverkaufHistorie.tsx`.
- `frontend/src/admin/finanzamt/TSEAusfalldokumentationSection.tsx` — heutige Nachsignier-Verwaltung; `frontend/src/admin/reporting/AdminDashboardPage.tsx:23-38` — Dashboard-Warnung; `frontend/src/admin/kasse/KassensitzungPage.tsx` — Kassenabschluss-UI; `frontend/src/lib/EinstellungenBackend.ts`.

Seed und Doku:

- `backend/seed/szenario.go:165-197` — `tseAusfall`-Fenster; `backend/seed/faketse.go`; `backend/seed/seed_integration_test.go`.
- `docs/handbuch.md:217` (§3.13 TSE-Architektur), `docs/compliance.md:62` (§3 TSE-Integration), `docs/verfahrensdokumentation.md:78` (§4 TSE-Anbindung), `docs/language.md:408-420` (TSE-Begriffe). Im Arbeitsbaum liegen bereits unkommittierte Doku-Anpassungen (Kassenabschluss-Begriff); Phase 7 baut darauf auf.

## Resolved decisions

- Tabellen werden umbenannt: `tse_signaturauftraege` und neu `tse_stoerungen`; `tse_signaturen` bleibt.
- Beleg-Abruf: derselbe Endpunkt antwortet mit Status (`eingereiht` | `ausstehend`); kein separater Status-Endpunkt.
- Advisory Lock: zweite Instanz wartet mit periodischem Neuversuch und Error-Log; die App läuft weiter (kein Fail-Fast).
- Kassenabschluss-Gate: 409 mit strukturierten Details im bestehenden Endpunkt; kein Vorab-Prüf-Endpunkt.
- Phasenschnitt: Phase 1 ist bewusst der große Cutover (Payload-Felder fallen nur einmal); sie behält übergangsweise die heutige einfache Fehlversuchs-Kurve, die Fehlertaxonomie ersetzt sie in Phase 2.
- Rückstands-Störungszeiträume materialisieren die Leser-Pfade idempotent über die Störungsprotokoll-Komponente; so wird auch ein hängender Worker dokumentiert.

## Open questions / Risks

- Zwischenstand nach Phase 1: Bei echtem TSE-Ausfall gibt es bis Phase 3 keinen Beleg mit Ausfallvermerk, nur die Ausstehend-Antwort. Pre-Release akzeptiert, aber Phasen 2–3 sollten zügig folgen.
- FIFO ist Soll, keine Garantie: Übersprungene Gift-Aufträge und Healing können die TSE-Log-Chronologie lokal durchbrechen (PRD-konform, im Nachsigniert-Vermerk erklärbar).
- Metrik-Fenster für Signaturen/Minute und Signierdauer-p95 (Phase 6) ist nicht festgelegt; Vorschlag: gleitendes Fenster über die letzten 15 Minuten, aus `erstellt_am` vs. `log_time_end` der Auftrags-/Signaturzeilen berechnet.

---

## Phase 1: Outbox-Kern — Buchen ohne TSE-Wartezeit

**User stories**: 1, 14, 15, 17, 18, 19, 20, 21; 3 teilweise (Ausstehend-Meldung mit Nachfassen, noch ohne Vermerke)

### Context

- `backend/api/tse/application/signing.go:51-159` — Sync-Signierer, entfällt ersatzlos.
- `backend/api/kasse/application/tse_signing.go`, `backend/api/table/application/tse_signing.go`, `backend/api/direktverkauf/application/tse_signing.go` — gehen in der fiskalischen Projektion auf.
- `backend/domain/kasse/tse_embedding.go`, `tse_data.go` — Payload-TSE-Felder, entfallen.
- `backend/repository/kassenjournal_repo/repo.go:57-276` — Write-Kombinatorik wird zu einem Event-Write mit optionalem Auftrag; `EroeffneKassensitzung` (Z. 229–269) verliert den TSE-Aufruf im DB-Callback.
- `backend/app/tse_nachsignier_worker.go` — wird zum Signatur-Worker (Healing und Quittierung übernehmen).
- `database/migrations/01_initial.up.sql:477-529`, `backend/sqlc/queries/tse_nachsignier.sql` — Schema und Queries.
- `backend/api/table/application/kassenbeleg_command.go:138-174`, `backend/domain/dsfinvk/mapper.go` — Leser auf Seitentabelle umstellen.
- `backend/seed/szenario.go:165-197`, `backend/seed/faketse.go`, `backend/seed/seed_integration_test.go` — Seed-Fluss.

### What to build

Der Cutover auf das Outbox-Modell: Kassen-Commands validieren, bauen das Event ohne TSE-Felder und schreiben es zusammen mit genau einem Signaturauftrag (aus der fiskalischen Projektion, mit processData-Snapshot und Plausibilisierung) in einer Transaktion — auch ohne TSE-Konfiguration, immer als offen. Der Worker wird nach jedem Commit sofort angestoßen (In-Process-Trigger, Polling als Fallback), arbeitet FIFO ab, heilt per Ist-Abfrage und quittiert atomar in die Seitentabelle; ein Advisory Lock sichert die Single-Prozess-Annahme (warten + Warnung). Die Fehlerbehandlung bleibt in dieser Phase die heutige einfache Fehlversuchs-Zählung.

Beleg-Abruf und DSFinV-K-Export lesen ausschließlich Auftrag + Seitentabelle: Der Beleg antwortet sofort mit `eingereiht` (Druckauftrag mit TSE-Abschnitt) oder `ausstehend` (kein Druckauftrag; die UI fasst 1–2-sekündlich für rund 10 Sekunden nach, danach auf Anforderung). Der Export füllt für unsignierte Vorgänge das Fehlerfeld. Die bestehende Admin-Seite und der TSE-Status laufen mit umbenannten Endpunkten minimal weiter (voller Umbau in Phase 6). Seed und Seed-Integrationstest werden auf Schema und Worker-Fluss umgestellt.

### Acceptance criteria

- [ ] Schema: `tse_signaturauftraege` mit `event_id`-Referenz (UNIQUE), Status inkl. `tse_nicht_konfiguriert`, Verwerfen-Feldern (Grund, Benutzer, Zeitpunkt); UUID-Kommentar korrigiert; `tse_signaturen`-Kommentar aktualisiert; Seed angepasst.
- [ ] Event-Payloads tragen keine TSE-Felder mehr; `tse_embedding.go` entfernt; Kassen-Command-Tests kommen ohne TSE-Verdrahtung aus.
- [ ] Fiskalische Projektion tabellengetrieben getestet: je Event-Typ signaturpflichtig ja/nein, processType, processData inkl. Vorzeichen-/Faktor-Fällen und datenabhängiger Sitzungseröffnung (mit/ohne Anfangsbestand); Plausibilisierung protokolliert Verstöße ohne zu blockieren.
- [ ] Journal-Repo: ein Event-Write mit optionalem Signaturauftrag; atomare Mehrfach-Writes (Storno, Umbuchung, Sitzungseröffnung) nehmen je Event ihren Auftrag; kein TSE-Aufruf in offenen DB-Transaktionen.
- [ ] Jeder signaturpflichtige Vorgang erzeugt im selben Commit genau einen offenen Auftrag, auch ohne TSE-Konfiguration.
- [ ] Worker: Sofort-Trigger nach Commit, Polling-Fallback fängt verlorene Trigger (Crash-Recovery-Test), FIFO im Regelbetrieb, Healing-Fälle (abgeschlossen/aktiv/unbekannt), atomare Quittierung, Client-Wiederverwendung, Advisory Lock mit Warten + Warnung.
- [ ] Beleg-Abruf: Sofortantwort mit Status `eingereiht`/`ausstehend`; UI-Nachfassen in Tisch- und Direktverkauf-Historie; TSE-Abschnitt kommt aus der Seitentabelle.
- [ ] DSFinV-K-Export liest nur die Seitentabelle (Join über Auftrag) und füllt das Fehlerfeld für unsignierte Vorgänge; Verprobung TSE-Export gegen Export ohne Waisen.
- [ ] Admin-Seite und TSE-Status funktionieren mit umbenannter Endpunktfamilie weiter.
- [ ] Seed-Integrationstest auf neues Schema und Worker-Fluss umgestellt; `make verify` grün.

---

## Phase 2: Fehlertaxonomie und Störungszustand des Workers

**User stories**: 2, 18; Grundstein für 8

### Context

- `backend/app/tse_nachsignier_worker.go:80-122` — heutige undifferenzierte Fehlerbehandlung.
- `backend/repository/tse_repo/repo.go:14-19,124-134` — Minuten-Backoff mit zehn Versuchen, entfällt.
- `backend/repository/tse_repo/fiskaly_client.go` — HTTP-Fehler entstehen hier; Ort der Klassifizierung.
- `backend/app/tse_nachsignier_worker_test.go` — Test-Vorbild mit Fake-TSE-Client.

### What to build

Eine explizite Fehlertaxonomie trennt auftragsspezifische Fehler (etwa 400/422: Fehlversuch am Auftrag, Backoff ~5/15/45 s mit Jitter, nach drei Versuchen endgültig fehlgeschlagen, Auftrag wird übersprungen und staut nie die Queue) von TSE-weiten Fehlern (Verbindung, 5xx, 429 samt Retry-After, 401/403, TSS-Zustandsfehler: Durchlauf-Abbruch ohne Auftrags-Fehlversuche). TSE-weite Fehler schalten den Worker in einen Störungszustand mit eigenem Backoff (Sekunden bis Minuten-Deckel, Jitter) und Half-Open-Wiedereinstieg über einen einzelnen Probe-Auftrag; jeder Durchlauf hat eine Deadline.

Dazu kommt das Störungsprotokoll: Tabelle `tse_stoerungen` plus Komponente mit idempotentem Öffnen/Schließen. Erste Quelle ist der Worker-Störungszustand (TSE-weiter Fehler öffnet, die erste erfolgreiche Signatur schließt).

### Acceptance criteria

- [ ] Fehlertaxonomie als explizite Typen im TSE-Client/Worker, kein String-Matching.
- [ ] Gift-Auftrag: drei Versuche mit Sekunden-Backoff, dann endgültig fehlgeschlagen; nachfolgende Aufträge werden im selben Durchlauf weiter signiert (Test).
- [ ] TSE-weiter Fehler bricht den Durchlauf ab, zählt keine Auftrags-Fehlversuche und öffnet einen Störungszeitraum (Test).
- [ ] Half-Open-Probe beendet den Störungszeitraum bei Erfolg und startet die volle Aufarbeitung; während der Störung wird fiskaly nicht mit dem Rückstand bombardiert (Test).
- [ ] Beide Backoffs tragen Jitter; jeder Durchlauf hat eine Deadline.
- [ ] `tse_stoerungen`-Zeiträume mit Beginn, Ende und Grund; kein Löschpfad.

---

## Phase 3: Signaturstatus-Funktion und Beleg-Vermerke

**User stories**: 3, 4, 5, 16

### Context

- `backend/api/table/application/kassenbeleg_command.go:138-174` — wird durch die Signaturstatus-Funktion ersetzt.
- `backend/api/bondruck/application/escpos/formatter.go` — `TSEAbschnitt.Nachsigniert` und `TSEAusfallvermerk` existieren bereits.
- `frontend/src/service/components/table/TischHistorie.tsx`, `frontend/src/service/components/direktverkauf/DirektverkaufHistorie.tsx` — Nachfass-UI aus Phase 1.

### What to build

Die Signaturstatus-Funktion kapselt die Beleg-Logik ohne Zeitverhalten und liefert genau eines von vier Ergebnissen: Signatur vorhanden; vorhanden mit Nachsigniert-Kennzeichen (mindestens ein Fehlversuch am Auftrag oder Signatur später als rund eine Minute nach Auftragserstellung); Ausfall mit belegbarem Grund (aktiver oder dem Auftrag zuzurechnender Störungszeitraum, oder auftragsspezifische Fehlversuche — der Gift-Fall); Signatur ausstehend. Der Rückstands-Ausfall wird zweite Störungsquelle: Ab zwei Minuten Alter des ältesten offenen Auftrags materialisieren die Leser-Pfade einen Störungszeitraum, Unterschreiten schließt ihn; das deckt hängenden Worker und App-Fehler ab. Belege in der Aufholphase nach dokumentiertem Ausfall dürfen den Ausfallvermerk tragen; bloße Queue-Latenz unterhalb der Schwelle liefert nie einen Ausfallvermerk, sondern Ausstehend.

Die Beleg-Erzeugung (Tisch, Direktverkauf, Storno) nutzt die Funktion für alle vier Ergebnisarten; der Nachdruck nachsignierter Belege trägt den Vermerk.

### Acceptance criteria

- [ ] Signaturstatus-Funktion mit den vier Ergebnisarten, getestet: vorhanden; dokumentierter Ausfall mit Grund; Rückstau ohne Störung → ausstehend; Schwellen-Überschreitung öffnet Zeitraum und kippt Ausstehend in Ausfall; Aufholphase; verspätete Signatur → Nachsigniert-Kennzeichen; keine falschen Ausfallvermerke bei bloßer Latenz.
- [ ] Rückstands-Zeitraum öffnet und schließt an der Zwei-Minuten-Schwelle, auch bei hängendem Worker (Leser materialisieren).
- [ ] Beleg ohne TSE-Daten entsteht nur bei dokumentiertem Ausfall oder in der Aufholphase; er weist den Ausfall aus.
- [ ] Nachsigniert-Vermerk erscheint nur in echten Ausfall- und Aufholszenarien (Kriterium Fehlversuch oder verspätet).
- [ ] Kassenbeleg-Erzeugung für alle vier Ergebnisarten getestet.

---

## Phase 4: TSE nicht konfiguriert

**User stories**: 13

### Context

- `backend/app/tse_nachsignier_worker.go:85-94` — heutiges stilles Aussetzen ohne Konfiguration; wird zum Markieren.
- `backend/api/settings/application/setup.go:310-331` — `speichereEinrichtung`: gemeinsamer Speicher-Schritt, Ort des Einrichtungs-Sweeps.
- `backend/api/settings/application/command.go:33` — `UpdateTSEKonfiguration` (reiner Zugangsdaten-Wechsel).
- `frontend/src/admin/kasse/KassensitzungPage.tsx` — Eröffnungs-Bestätigung ohne TSE (F6) existiert bereits.

### What to build

Der Worker markiert offene Aufträge ohne vorhandene TSE-Konfiguration endgültig als `tse_nicht_konfiguriert` (keine Fehlversuche, keine Rückstands-Warnung, keine automatische Wiederaufnahme) und unterscheidet dabei fehlende von nicht lesbarer Konfiguration (Fehler: nichts tun). Der Übergang von nicht konfiguriert zu konfiguriert markiert in derselben Transaktion wie das Speichern der Konfiguration alle noch offenen Aufträge endgültig (Einrichtungs-Sweep in allen Einrichtungspfaden); ein reiner Zugangsdaten-Wechsel bei durchgehend vorhandener Konfiguration markiert nichts. Änderungen der TSE-Konfiguration sind nur ohne offene Kassensitzung möglich. Belege solcher Vorgänge tragen den Vermerk, dass keine TSE konfiguriert ist. Der Dauerzustand ohne Konfiguration wird dritte Störungsquelle (Zeitraum endet mit der Einrichtung). Das Admin-Zurücksetzen (einzeln oder gesamt) reiht endgültig markierte Aufträge bewusst wieder ein; automatisch nachsigniert wird nichts.

### Acceptance criteria

- [ ] Aufträge entstehen ohne Konfiguration als offen und werden vom Worker endgültig markiert; nicht lesbare Konfiguration markiert nichts (Test).
- [ ] Einrichtungs-Sweep: Übergang zu konfiguriert markiert auch noch unmarkierte offene Aufträge in derselben Transaktion; reiner Zugangsdaten-Wechsel nicht (Test).
- [ ] Spätere Einrichtung fasst endgültig markierte Aufträge nicht an; Admin-Zurücksetzen reiht sie ein und der Worker signiert sie nach (Test).
- [ ] TSE-Konfigurationsänderungen werden bei offener Kassensitzung abgelehnt (klare Fehlermeldung).
- [ ] Beleg-Vermerk „keine TSE konfiguriert"; Störungszeitraum `keine_konfiguration` endet mit der Einrichtung.

---

## Phase 5: Kassenabschluss-Gate

**User stories**: 10, 11

### Context

- `backend/api/kasse/application/command.go:301-436` — `KasseAbschliessen`; das Gate wird erste Handlung vor der Barriere (Z. 314–324).
- `backend/api/admin.go:132`, `frontend/src/admin/kasse/KasseBackend.ts:77`, `frontend/src/admin/kasse/KassensitzungPage.tsx` — Endpunkt und UI.

### What to build

Das Gate schützt die gesamte Ein-Klick-Abschluss-Operation und prüft sofort, noch vor der wird-abgeschlossen-Barriere: Nur frische offene Aufträge ohne Ausfallbezug blockieren; die 409-Antwort nennt Anzahl und Alter, die UI zeigt die Meldung und bietet erneutes Anfordern. Ausfall-Reste lassen den Abschluss zu und werden in der Abschlussmeldung ausgewiesen: endgültig fehlgeschlagene und verworfene Aufträge stets, offene Aufträge mit Zurechnung zu einem dokumentierten (auch laufenden) Störungszeitraum oder mit auftragsspezifischen Fehlversuchen. `tse_nicht_konfiguriert` blockiert nie; schließt ein Tag vollständig ohne TSE, weist die Abschlussmeldung das deutlich aus. Die signaturpflichtigen Abschluss-Events laufen regulär über die Queue; nach dem Abschluss verbliebene offene Reste arbeitet der Worker bei Rückkehr der TSE regulär nach.

### Acceptance criteria

- [ ] Gate prüft sofort und wartet nie; leere Queue lässt durch (Test).
- [ ] Frischer offener Auftrag blockiert mit 409 samt Anzahl und Alter; UI zeigt Meldung mit erneutem Anfordern (Test).
- [ ] Ausfall-Reste (inkl. offener Aufträge im laufenden Störungszeitraum) lassen den Abschluss zu; die Abschlussmeldung weist sie aus (Test).
- [ ] `tse_nicht_konfiguriert` blockiert nicht; Tag ohne TSE wird deutlich ausgewiesen (Test).
- [ ] Differenzbuchung und Tagesabschluss werden über die Queue signiert; Reste nach dem Abschluss werden nach TSE-Rückkehr nachsigniert.

---

## Phase 6: Admin und Monitoring

**User stories**: 6, 7, 8, 9, 12

### Context

- `backend/api/admin.go:156-162`, `backend/api/tse/application/command.go`, `query.go`, `backend/api/tse/http/handler.go` — heutige Admin-Endpunkte.
- `frontend/src/admin/finanzamt/TSEAusfalldokumentationSection.tsx` — wird zur Signaturauftrags-Verwaltung plus Störungsprotokoll-Ansicht.
- `frontend/src/admin/reporting/AdminDashboardPage.tsx:23-38`, `backend/api/settings/http/query_handler.go:62` — Dashboard-Warnung und TSE-Status.

### What to build

Die Nachsignier-Verwaltung wird zur Signaturauftrags-Verwaltung: Statusliste mit Versuchen und letztem Fehler, Zurücksetzen einzeln oder gesamt (für `fehlgeschlagen` und `tse_nicht_konfiguriert`), Verwerfen mit Begründung (für offene und fehlgeschlagene Aufträge; protokolliert Grund, Benutzer, Zeitpunkt). Dazu der Queue-Zustand: Anzahl offener Aufträge, Alter des ältesten, Signaturen pro Minute und Signierdauer p95 (Latenz aus `erstellt_am` vs. TSE-logTime, jederzeit nachweisbar). Das Dashboard warnt ab rund einer Minute Rückstand oder bei endgültig fehlgeschlagenen Aufträgen; ohne TSE-Konfiguration zeigt es eine permanente, unübersehbare Konfigurationswarnung. Die Ausfalldokumentations-Ansicht zeigt die Störungszeiträume mit Beginn, Ende und Grund direkt aus `tse_stoerungen`.

### Acceptance criteria

- [ ] Signaturauftrags-Verwaltung: Liste, Zurücksetzen einzeln/gesamt, Verwerfen mit Pflicht-Begründung; Statuswechsel protokolliert Benutzer und Zeitpunkt.
- [ ] Queue-Zustand sichtbar: offene Aufträge, Alter des ältesten, Signaturen/Minute, Signierdauer p95 — wachsender von schrumpfendem Rückstand unterscheidbar.
- [ ] Dashboard warnt ab ~1 Minute Rückstand und bei endgültig fehlgeschlagenen Aufträgen; ab 2 Minuten existiert der Störungszeitraum (Phase 3).
- [ ] Ohne TSE-Konfiguration permanenter Konfigurationsalarm statt Queue-Alarm.
- [ ] Ausfalldokumentations-Ansicht zeigt Störungszeiträume aus dem Störungsprotokoll (Zeitraum mit Grund, nicht Einzelaufträge).

---

## Phase 7: Dokumentation und Sprache

**User stories**: flankiert 8 und 16 (Nachweisbarkeit); Konformitätsbedingung Herstellerdokumentation

### Context

- `docs/language.md:408-420` — TSE-Begriffe (u. a. veralteter `Nachsignierung`-Eintrag); Kassensturz/Z-Bon-Einträge sind im Arbeitsbaum bereits auf den Kassenabschluss-Begriff umgestellt.
- `docs/handbuch.md:217` — §3.13 TSE-Architektur (Sync-Pfad).
- `docs/compliance.md:62` — §3 TSE-Integration.
- `docs/verfahrensdokumentation.md:78` — §4 TSE-Anbindung (zugleich Herstellerdoku nach BSI TR-03153-1 Kap. 3.9.3).

### What to build

Die Ubiquitous Language erhält die neuen Begriffe (Signaturauftrag, Signatur-Worker, Signaturstatus, Störungsprotokoll, Störungszeitraum, Aufholphase, Rückstands-Ausfall, Signatur ausstehend, TSE nicht konfiguriert); der Nachsignierung-Eintrag wird auf das Outbox-Modell umgestellt. Handbuch (TSE-Architektur) und Compliance-Dokumentation (TSE-Integration) beschreiben den einen Signaturpfad, den einen Leseweg und das Störungsprotokoll. Die Muster-Verfahrensdokumentation (TSE-Anbindung) erläutert Mechanismus, typische Latenz (Ziel p95 unter fünf Sekunden) und mögliche Verzögerungen und erfüllt damit die Herstellerdokumentations-Pflicht. Endpunkt- und UI-Texte werden gegen die Begriffe geprüft.

### Acceptance criteria

- [ ] Alle neuen Begriffe in `docs/language.md`, konsistent mit Endpunkt- und UI-Benennung; veraltete Einträge aktualisiert.
- [ ] `docs/handbuch.md` §3.13 beschreibt das Outbox-Modell (Auftrag, Worker, Seitentabelle, Störungsprotokoll, Gate).
- [ ] `docs/compliance.md` §3 auf den asynchronen Pfad und die Konformitätsbedingungen aktualisiert.
- [ ] `docs/verfahrensdokumentation.md` §4 erfüllt die Herstellerdoku-Pflicht (Mechanismus, typische Latenz, Verzögerungen).
