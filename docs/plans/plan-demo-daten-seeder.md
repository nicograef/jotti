# Plan: Demo-Daten-Seeder

> Source PRD: docs/prds/prd-demo-daten-seeder.md

## Goal

Die handgepflegte `database/seed.sql` (3289 Zeilen, ~993 Events) wird durch einen
deterministischen Go-Seeder ersetzt, der als Subkommando `seed` des Backend-Binaries läuft.
Er erzeugt das 3-Tage-Sommerfest des „TSV Musterstadt e.V." vollständig über die
Domain-Event-Konstruktoren — inklusive Fake-TSE-Daten mit Ausfallfenster, Direktverkauf,
Umbuchung, kompletter Kassenführung, Bondruck-Historie und Tisch-Favoriten — und stößt
abschließend den Projektions-Rebuild an. `make seed` und `scripts/prod-reset-and-seed.sh`
rufen das Subkommando auf; die seed.sql wird gelöscht.

## Architectural decisions

Durable decisions that apply across all phases:

- **CLI**: Subkommando `seed` im Backend-Binary, Dispatch in `backend/main.go` analog zu
  `rebuild-projections` (dev: `go run ./main.go seed`, prod: `jotti seed`). `make seed`
  ruft nur noch das Subkommando auf — der Rebuild-Schritt entfällt aus dem Make-Target,
  weil der Seeder ihn selbst anstößt.
- **Paketlayout**: Ein Paket `backend/seed` mit einer Datei pro Modul (Szenario-Definition,
  Szenario-Engine, Fake-TSE-Signierer, Seed-Writer) plus zugehörige `_test.go`-Dateien.
  Keine Unterpakete; interne Typen bleiben unexportiert.
- **Datenfluss**: Drehbuch (reine Daten, kein I/O) → Engine (Events mit Subjects,
  lückenlosen Versionen ab 1 und Zeitstempeln relativ zu „jetzt"; berechnet
  Tagesabschluss-Summen aus den Tages-Events) → Fake-TSE-Signierer (TSE-Daten,
  Nachsignier-Aufträge, Signatur-Zeilen) → Writer (eine Transaktion: Stammdaten, Events,
  Seitentabellen; Sequenzkorrektur per `setval`; danach `RebuildAllProjections`).
- **Guard**: Enthält `kassenjournal` bereits Events, bricht der Seeder ohne Schreibzugriff
  ab und nennt den Reset-Weg (`make clean && make dev && make seed` bzw.
  prod-reset-and-seed.sh). Kein `--force`.
- **Events**: Ausschließlich über die bestehenden Domain-Konstruktoren
  (`kasse.New…Event`) plus `kasse.EmbedTSEIn…`. Die Engine überschreibt nach der
  Konstruktion nur `Event.Time` (Konstruktoren setzen `time.Now()`) und vergibt
  `Event.Version` selbst — Payload-Aufbau und Validierung bleiben in der Domain-Schicht.
- **Subjects**: Bestehende Konventionen aus `domain/kasse/subject.go`
  (`kassensitzung-<zNr>`, `…/tisch-<id>`, `…/direktverkauf-<uuid>`). Drei
  Kassensitzungen mit z_nr 1–3 (Freitag ≈ heute−2, Samstag ≈ heute−1, Sonntag = heute).
- **DB-Zugriff**: Writer nutzt sqlc-Queries in einer Transaktion; wo bestehende Queries
  nicht passen (explizite IDs für Stammdaten, historische Status/Zeitstempel für
  Druckaufträge und Nachsignier-Aufträge), kommen seed-spezifische Queries in eine eigene
  Query-Datei. Kein Schema-Change.
- **TSE-Konfiguration bleibt leer** (Migration befüllt sie mit Leerstrings) — der
  Nachsignier-Worker bleibt dadurch inaktiv.

## Inventory

- `backend/main.go:46-51` — Subkommando-Dispatch (`rebuild-projections`); Vorbild für `seed`.
- `backend/main.go:84-102` — `rebuildProjections` über `kassenjournal_repo.RebuildAllProjections`.
- `Makefile:169-176` — `seed`-Target (psql-Import + rebuild-projections) und Container-Variablen.
- `Makefile:128-129` — `prod-reset-and-seed`-Target.
- `scripts/prod-reset-and-seed.sh` — psql-Import von `database/seed.sql` + `jotti rebuild-projections` im Backend-Container; stellt auf `jotti seed` um.
- `database/seed.sql` — altes Drehbuch (Benutzer, Tische, Produkte, Tisch-Dramaturgie, Event-Verteilung); Vorlage für die Szenario-Definition, wird in Phase 5 gelöscht.
- `database/migrations/01_initial.up.sql:105-127` — `kassensitzungen` (z_nr-Identity, Unique-Index „eine offen").
- `database/migrations/01_initial.up.sql:130-146` — `kassenjournal` (append-only, `UNIQUE (subject, version)`, FK auf kassensitzungen, Schreibschutz-Trigger).
- `database/migrations/01_initial.up.sql:218-224` — `tisch_favoriten`.
- `database/migrations/01_initial.up.sql:234-256` — `druckstationen` (5 Kategorien vorbefüllt mit leerer IP).
- `database/migrations/01_initial.up.sql:259-272` — `druckauftraege` (Status offen/gedruckt/fehlgeschlagen/verworfen, bon_art, versuche, letzter_fehler).
- `database/migrations/01_initial.up.sql:389-407` — `betreiber` (Singleton id=1).
- `database/migrations/01_initial.up.sql:435-459` — `tse_nachsignier_auftraege` (Ausfalldokumentation: erstellt_am = Beginn, erledigt_am = Ende, letzter_fehler = Grund; tx_id UNIQUE).
- `database/migrations/01_initial.up.sql:462-475` — `tse_signaturen` (Seitentabelle für nachsignierte Vorgänge).
- `backend/domain/event/event.go:36-68` — `event.New` setzt `Time = time.Now()`; Version vergibt der Aufrufer.
- `backend/domain/kasse/subject.go:10-23` — Subject-Konventionen.
- `backend/domain/kasse/kassensitzung_events.go:106-212` — Konstruktoren Eröffnung, Geldtransit, Kassensturz, Differenz, Tagesabschluss.
- `backend/domain/kasse/tisch_session_events.go:101-213` — Konstruktoren Bestellung, Zahlung, Stornierung, Ausgabe, Auszahlung (Payload-IDs intern via `uuid.New()`).
- `backend/domain/kasse/direktverkauf_events.go:57-118` — Konstruktoren Direktverkauf getätigt/storniert.
- `backend/domain/kasse/tse_data.go:10-45` — `TSEData` + `Validate` (Pflichtfelder, RFC3339-LogTimes).
- `backend/domain/kasse/tse_embedding.go:46-93` — `EmbedTSEIn…` für genau die neun fiskalischen Event-Typen; `tseData == nil` markiert den Ausfall.
- `backend/api/tse/application/signing.go:101-139` — Vorbild für TSEData-Aufbau (txID als UUID, LogTimes, Validierung vor Embed).
- `backend/api/tse/application/processdata.go:16-136` — processData-Builder (Kassenbeleg, Bestellung, Geldtransit/Eigenbeleg, Tagesabschluss) — vom Fake-Signierer wiederzuverwenden.
- `backend/repository/kassenjournal_repo/repo.go:34-51,204-242` — `WriteEvent`/`writeEventInTx` (Event + synchrone Projektion).
- `backend/repository/kassenjournal_repo/repo.go:175-199` — `WriteUmbuchung` (Storno + Bestellung atomar).
- `backend/repository/kassenjournal_repo/repo.go:432-453` — `GetTSESignaturByTxID` (Belegansicht löst nachsignierte Signaturen auf).
- `backend/repository/kassenjournal_repo/repo.go:466-566` — `RebuildAllProjections`.
- `backend/api/table/application/command.go:557-585` — Umbuchungs-Semantik: Storno-/Bestellungs-Paar, Standard-Kommentare „Umbuchung auf/von Tisch …", OCC-Versionen.
- `backend/api/bondruck/application/escpos/formatter.go:256-280` — ESC/POS-Formatter; QR-Code wird 1:1 aus `TSE.QRCodeData` gerendert.
- `backend/app/tse_nachsignier_worker.go:85-92` — Worker prüft `IstKonfiguriert()`; leere TSE-Konfiguration hält ihn inaktiv.
- `backend/domain/user/password.go:41-75` — Argon2id mit zufälligem Salt (nicht deterministisch).
- `backend/sqlc/queries/` — bestehende Inserts: `users.sql:14`, `tables.sql:25`, `produkte.sql:93`, `kassensitzungen.sql:2`, `betreiber.sql:7`, `favoriten.sql:2`, `relay.sql:2` (druckauftraege), `tse_nachsignier.sql:2`, `tse_signaturen.sql:2`.
- `test-integration.sh` — Integrationstest-Setup (Postgres 17-Container, `-tags=integration`, `-p 1`).
- `backend/repository/kassenjournal_repo/repo_test.go:1163` — Prior Art: Events per Raw-Insert im Integrationstest.

## Resolved decisions

Geklärt am 2026-06-12 (Planungs-Runde, zusätzlich zu den PRD-Entscheidungen):

- **Payload-UUIDs bleiben zufällig.** Die Domain-Konstruktoren erzeugen IDs intern via
  `uuid.New()`; das wird akzeptiert. Determinismus gilt auf Zustandsebene (identische
  Salden, Summen, Status, Event-Folgen), nicht auf Byte-Ebene der Payload-IDs. Die
  PRD-Aussage „feste UUIDs nach erkennbarem Schema" gilt nur dort, wo der Seeder die IDs
  selbst vergibt: Direktverkauf-Subjects und TSE-txIDs stehen fest im Drehbuch.
- **Ein Paket `backend/seed`**, eine Datei pro Modul — keine Unterpakete.
- **Druckauftrag-Payloads** werden über den bestehenden ESC/POS-Formatter aus den echten
  Szenario-Daten erzeugt (Kassenbelege inklusive TSE-QR), keine Platzhalter.
- **Passwort-Hash**: feste Argon2id-Konstante für `jotti123` (wie bisher in seed.sql),
  weil `createArgon2idHash` salt-zufällig ist. Der Admin aus der Migration bleibt unberührt.
- **QR-Code-Daten**: KassenSichV-üblicher `V0;…`-String wie von fiskaly geliefert; der
  Belegdruck rendert das Feld unverändert, die Domain-Validierung verlangt nur Nicht-Leere.
- **Kein Zufallsgenerator in der Engine nötig**: Das Drehbuch ist vollständig deklarativ;
  der „feste Zufalls-Seed" aus der PRD reduziert sich darauf, dass nichts gewürfelt wird.

## Open questions / Risks

- **Zwischenstand nach Phase 1**: `make seed` liefert bis Phase 2 nur das minimale
  Drehbuch. Akzeptiert (Pre-Release); die alte seed.sql bleibt bis Phase 5 im Repo und
  der Prod-Weg unverändert, falls die volle Demo gebraucht wird.
- **Guard prüft nur das Kassenjournal** (PRD-Entscheidung). Eine DB mit Stammdaten, aber
  ohne Events, würde geseedet und könnte ID-Kollisionen erzeugen — auf dem dokumentierten
  Reset-Weg (frische DB) ausgeschlossen.
- **Engine-Komplexität**: Die Tagesabschluss-Summen-Berechnung dupliziert bewusst keine
  Reporting-Logik, sondern aggregiert die selbst erzeugten Events; Unit-Tests prüfen die
  Übereinstimmung mit einer unabhängigen Aggregation.

---

## Phase 1: Durchstich — Subkommando, Guard, Stammdaten, minimales Drehbuch

**User stories**: 4, 23, 27, 31

### Context

- `backend/main.go:46-51` — Subkommando-Dispatch erweitern.
- `Makefile:169-176` — `seed`-Target auf das Subkommando umstellen.
- `backend/repository/kassenjournal_repo/repo.go:466-566` — Rebuild, vom Seeder angestoßen.
- `backend/domain/event/event.go:36-68` — `Time`-Überschreiben und Versionsvergabe durch die Engine.
- `backend/domain/kasse/kassensitzung_events.go:106-125`, `tisch_session_events.go:101-213` — Konstruktoren für den Minimal-Ablauf.
- `backend/sqlc/queries/users.sql:14`, `tables.sql:25`, `produkte.sql:93`, `kassensitzungen.sql:2`, `betreiber.sql:7` — bestehende Inserts; seed-spezifische Varianten mit expliziten IDs ergänzen.
- `database/migrations/01_initial.up.sql:130-146` — Guard-Ziel `kassenjournal`.
- `test-integration.sh` — Testumgebung für den Integrationstest.

### What to build

Das Paket `backend/seed` mit allen vier Modulen im Miniatur-Zustand plus CLI-Verdrahtung:
Die Szenario-Definition enthält die aktiven Stammdaten (Benutzer mit bekanntem
`jotti123`-Hash, Tische, Produkte mit Varianten, Betreiber „TSV Musterstadt e.V.").
Die Engine erzeugt eine offene Kassensitzung „heute" (Sonntag) mit einigen wenigen
Bestell-/Ausgabe-/Zahlungs-Zyklen über Domain-Konstruktoren, vergibt Versionen und
historische Zeitstempel. Der Writer prüft den Guard, persistiert Stammdaten und Events
transaktional über sqlc, korrigiert die Identity-Sequenzen und stößt den Rebuild an.
`make seed` ruft `go run ./main.go seed` im Backend-Container auf. seed.sql und
Prod-Skript bleiben unangetastet.

### Acceptance criteria

- [x] `make seed` auf frischer Dev-DB (`make clean && make dev`) spielt Stammdaten und
      die offene Sonntags-Kassensitzung in einem Schritt ein — ohne separaten
      `rebuild-projections`-Aufruf, ohne psql-Import.
- [x] Demo-Logins funktionieren mit Passwort `jotti123`; Betreiber-Stammdaten sind in den
      Einstellungen sichtbar; mindestens ein Tisch zeigt Bestellung, Ausgabe und Zahlung.
- [x] Alle Events entstehen über Domain-Konstruktoren; Versionen pro Subject lückenlos
      ab 1; Zeitstempel liegen im Sitzungszeitraum.
- [x] Guard: Lauf gegen eine DB mit Kassenjournal-Events bricht ohne Schreibzugriff ab
      und nennt den Reset-Weg.
- [x] Integrationstest (`-tags=integration`): Erstlauf auf frischer DB erfolgreich,
      Zweitlauf bricht mit Guard-Fehler ab und schreibt nichts.
- [x] `database/seed.sql` und `prod-reset-and-seed.sh` bleiben unverändert (Abriss in Phase 5).

---

## Phase 2: Volles 3-Tage-Drehbuch (Kassenführung, Umbuchung, Direktverkauf, Favoriten)

**User stories**: 1, 5, 6, 7, 8, 9, 10, 15, 16, 17, 18, 19, 20, 21, 22, 26, 28

### Context

- `database/seed.sql` — bewährte Tisch-Dramaturgie, Produktliste und Event-Verteilung als Vorlage.
- `backend/api/table/application/command.go:557-585` — Umbuchungs-Kommentare und Event-Paar.
- `backend/domain/kasse/direktverkauf_events.go:57-118`, `subject.go:19-23` — Direktverkaufs-Strang mit festen Subject-UUIDs aus dem Drehbuch.
- `backend/domain/kasse/kassensitzung_events.go:127-212` — Geldtransit, Kassensturz, Differenz, Tagesabschluss.
- `backend/sqlc/queries/favoriten.sql:2` — Favoriten-Insert.

### What to build

Die Szenario-Definition wächst zum vollständigen Drehbuch: Freitag (~160 Events, ruhig)
und Samstag (~700 Events, voll) abgeschlossen, Sonntag (~130 Events) offen — mit
Stammtisch, Geburtstagsfeier und Stoßzeiten wie bisher, plus den neuen Erzählsträngen:
Direktverkaufsstand (Festbändchen/Kuchen, über alle drei Tage, ein Storno), eine
Umbuchung vom Stehtisch an einen freien Tisch, Geldtransit (Entnahme Samstag, Einlage
Sonntag), Kassensturz Samstagabend mit kleiner Soll/Ist-Differenz und Differenz-Buchung,
Auszahlungen, Stornierungen mit Begründung durch Serviceleitung. Die Engine berechnet
die Tagesabschluss-Summen für Freitag und Samstag aus den tatsächlich erzeugten
Tages-Events. Stammdaten ergänzt um Inaktiv-/Gelöscht-Beispiele (Benutzer, Tische,
Produkte) und Tisch-Favoriten für mehrere Service-Benutzer.

### Acceptance criteria

- [x] Drei Kassensitzungen in der bisherigen Größenordnung (~160/~700/~130 Events);
      Freitag und Samstag abgeschlossen, Sonntag offen.
- [x] Tagesabschluss-Events tragen aus den Tages-Events berechnete Summen; ein Unit-Test
      prüft sie gegen eine unabhängige Aggregation.
- [x] Umbuchung als atomares Storno-/Bestellungs-Paar mit Standard-Kommentaren
      („Umbuchung auf/von Tisch …") und identischen Positionen.
- [x] Direktverkäufe mit festen Subject-UUIDs aus dem Drehbuch, mindestens ein Storno.
- [x] Kassenführung komplett: Geldtransit beide Richtungen, Kassensturz mit Differenz und
      zugehöriger Differenz-Buchung.
- [x] Sonntags-Tische decken alle Zustände ab (leer, frisch bestellt, teilgeliefert,
      teilbezahlt, Guthaben/Auszahlung, abgeschlossen); für abgeschlossene Tische gilt
      Bestellungen − Stornierungen − Zahlungen (± Auszahlung) = 0.
- [x] Benutzer in allen Lebenszyklus-Zuständen und beiden Service-Rollen; Produkte und
      Tische mit Soft-Delete-/Inaktiv-Beispielen; Tisch-Favoriten für mehrere
      Service-Benutzer.
- [x] Engine-Unit-Tests (`-tags=unit`): lückenlose Versionen, monotone Zeitstempel im
      Sitzungszeitraum, Tagesabschluss-Summen, Umbuchungspaar-Konsistenz, Salden.

---

## Phase 3: Fake-TSE — Signaturen, Ausfallfenster, Nachsignier-Aufträge

**User stories**: 2, 13, 14 (und 3 teilweise)

### Context

- `backend/domain/kasse/tse_embedding.go:46-93` — Embed-Funktionen der neun fiskalischen Event-Typen.
- `backend/domain/kasse/tse_data.go:21-45` — Validierung, die die Fake-Daten bestehen müssen.
- `backend/api/tse/application/signing.go:101-139` — Vorbild für txID und TSEData-Aufbau.
- `backend/api/tse/application/processdata.go:16-136` — processType/processData-Builder (wiederverwenden).
- `database/migrations/01_initial.up.sql:435-475` — Nachsignier-Outbox und Signatur-Seitentabelle.
- `backend/repository/kassenjournal_repo/repo.go:432-453` — Auflösung nachsignierter Signaturen in der Belegansicht.
- `backend/app/tse_nachsignier_worker.go:85-92` — bleibt durch leere TSE-Konfiguration inaktiv.

### What to build

Der Fake-TSE-Signierer nimmt die Event-Folge der Engine und versieht genau die
fiskalischen Event-Typen über `kasse.EmbedTSEIn…` mit konsistenten TSE-Daten: global
streng monotone Transaktionsnummern und Signaturzähler, eine feste Fake-Seriennummer,
logTime-Paare aus den Event-Zeitstempeln, processType/processData über die bestehenden
Builder, Signatur-Strings und QR-Code-Daten im `V0;…`-Format. Im Ausfallfenster
(Samstagabend zur Stoßzeit, ca. 45–90 Minuten) erhalten die Events nur die txID ohne
TSE-Daten; stattdessen entsteht je Event ein Nachsignier-Auftrag mit Beginn, Ende und
Grund — überwiegend erledigt mit nachgetragener Signatur in `tse_signaturen`, einzelne
fehlgeschlagen mit Fehlertext, genau einer verworfen. Der Writer persistiert beide
Seitentabellen in derselben Transaktion.

### Acceptance criteria

- [x] Genau die fiskalischen Event-Typen (Bestellung, Zahlung, Stornierung, Auszahlung,
      Direktverkauf und -storno, Geldtransit, Differenz, Tagesabschluss) tragen TSE-Daten;
      Ausgabe, Kassensitzungs-Eröffnung und Kassensturz nicht.
- [x] Transaktionsnummern und Signaturzähler global streng monoton; eine feste
      Fake-Seriennummer; alle erzeugten TSE-Daten bestehen `kasse.TSEData.Validate`.
- [x] Ausfallfenster: Events ohne TSE-Daten (TSEAusfall-Flag wo vorhanden); je Event genau
      ein Nachsignier-Auftrag; erledigte Aufträge haben eine passende
      `tse_signaturen`-Zeile, die die Belegansicht auflöst; einzelne fehlgeschlagen mit
      verständlichem Fehlertext, genau einer verworfen.
- [x] In der Demo: Beleg eines bezahlten Tischs zeigt TSE-Signaturdaten und QR-Code; die
      TSE-Einstellungen zeigen Nachsignier-Aufträge in allen Status.
- [x] TSE-Konfiguration bleibt leer; der Nachsignier-Worker startet keine Signierversuche.
- [x] Unit-Tests (`-tags=unit`) gemäß PRD-Testentscheidungen (Monotonie, Typ-Abdeckung,
      Ausfallfenster-Paarigkeit, Domain-Validierung).

---

## Phase 4: Bondruck — Druckstationen und Druckauftrags-Historie

**User stories**: 11, 12, 32 (und 3 abschließend)

### Context

- `database/migrations/01_initial.up.sql:234-272` — druckstationen (vorbefüllt, leere IPs) und druckauftraege.
- `backend/sqlc/queries/relay.sql:2` — Druckauftrag-Insert (nur Status `offen`); seed-spezifische Variante für historische Status nötig.
- `backend/api/bondruck/application/escpos/formatter.go` — Payload-Erzeugung inkl. TSE-QR auf Kassenbelegen.
- `backend/sqlc/queries/druckstation.sql` — Update der Stations-Konfiguration.

### What to build

Das Drehbuch erhält Druckstations-Konfiguration (alle fünf Stationen mit realistischen
LAN-IPs, gemischte Bonmodi, Kassenbeleg ohne Bonmodus) und eine Druckauftrags-Historie,
die zu den Szenario-Events passt: überwiegend gedruckt (mit `gedruckt_am`), einige offen,
mehrere fehlgeschlagen mit verständlichem Fehlertext und `versuche ≥ 3`, einer verworfen
— beide Bon-Arten. Die Payloads baut die Engine über den bestehenden ESC/POS-Formatter
aus den echten Bestellungen und Zahlungen (Kassenbelege inklusive TSE-QR-Daten aus
Phase 3).

### Acceptance criteria

- [x] Alle fünf Druckstationen haben plausible LAN-IPs und gemischte Bonmodi
      (`pro_position`/`pro_bestellung`, Kassenbeleg NULL).
- [x] Druckaufträge existieren in allen vier Status mit beiden Bon-Arten; fehlgeschlagene
      tragen verständliche Fehlertexte (z. B. „Drucker nicht erreichbar").
- [x] Payloads sind echte ESC/POS-Bytes aus dem bestehenden Formatter und passen
      inhaltlich zu den Szenario-Events.
- [x] Die Bondruck-Fehlerliste der Demo zeigt die fehlgeschlagenen Aufträge; „Erneut
      versuchen" und „Verwerfen" sind bedienbar.
- [x] Integrationstest erweitert: `druckstationen` konfiguriert, `druckauftraege` in allen
      Status befüllt.

---

## Phase 5: Umstellung und Abriss — Prod-Skript, seed.sql löschen, E2E-Test

**User stories**: 24, 25, 29, 30

### Context

- `scripts/prod-reset-and-seed.sh` — `SEED_FILE`-Prüfung, psql-Import und Rebuild-Aufruf ersetzen.
- `Makefile:128-129,169-176` — finale Targets.
- `database/seed.sql` — löschen.
- `docs/plans/plan-bondruck-test-escpresso.md`, `backend/repository/kassenjournal_repo/repo_test.go:1163` — verbleibende Verweise prüfen/anpassen.

### What to build

`prod-reset-and-seed.sh` ruft statt psql-Import + Rebuild nur noch `jotti seed` im
Backend-Container auf (Guard und Rebuild übernimmt der Seeder). `database/seed.sql` wird
gelöscht und alle Verweise im Repo (Makefile, Skripte, Doku, Kommentare) werden
umgestellt. Der Ende-zu-Ende-Integrationstest wird komplettiert: Ein voller Lauf auf
frischer DB befüllt alle Feature-Tabellen, ein erneuter Projektions-Rebuild ändert
nichts, der zweite Seed-Lauf bricht ab, ohne zu schreiben.

### Acceptance criteria

- [ ] `prod-reset-and-seed.sh` nutzt das Subkommando; kein psql-Import, kein separater
      Rebuild-Schritt; `make prod-reset-and-seed` funktioniert unverändert.
- [ ] `database/seed.sql` ist gelöscht; `grep -r seed.sql` findet keine produktiven
      Verweise mehr (nur Git-Historie).
- [ ] Integrationstest komplett: Betreiber, Druckstationen, Druckaufträge,
      Nachsignier-Aufträge, Signaturen und Favoriten sind befüllt; erneuter
      `RebuildAllProjections` ändert keine Projektionen; Zweitlauf-Guard schreibt nichts.
- [ ] `make check` und `make verify` laufen grün.
