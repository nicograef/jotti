# Release-Guide: jotti v1.0.0

Was geprüft, getestet, validiert, korrigiert und dokumentiert sein muss, bevor die erste stabile offizielle Version `v1.0.0` getaggt und veröffentlicht wird. Arbeitsdokument mit Abhak-Listen; nach dem Release aus `docs/plans/` entfernen.

Gliederung: erst die Grundsatz-Entscheidungen (was 1.0 verspricht), dann sechs Gates (Code, Fiskal, Betrieb, Schema-Disziplin, Doku, Release-Mechanik), am Ende die kompakte Go-Live-Checkliste.

---

## 0. Was v1.0.0 bedeutet (Grundsatz-Entscheidungen)

Ab 1.0.0 gilt SemVer im vollen Sinn. Diese Zusagen müssen vor dem Tag bewusst getroffen und dokumentiert sein, weil sie danach nur noch mit Major-Bump gebrochen werden dürfen:

- [ ] **Stabiles DB-Schema.** `01_initial.up.sql` ist ab 1.0.0 eingefroren. Jede spätere Änderung ist eine neue, additive Migration (`02_*.up.sql` + `.down.sql`). Siehe Gate 4.
- [ ] **Stabile Event-Contracts.** Die JSON-Form der Kassenjournal-Events (`bestellung-aufgenommen:v1`, `zahlung-kassiert:v1`, …) ist eingefroren; Guard ist `backend/domain/kasse/event_json_contract_test.go`. Änderungen an einem Event nur als neue Version (`:v2`), nie in-place. Grund: Events sind 10 Jahre aufzubewahren und müssen replaybar bleiben.
- [ ] **Stabile HTTP-API.** Endpunkte und Payloads, auf die das Print-Relay und das Frontend bauen, bleiben rückwärtskompatibel.
- [x] **Scope-Entscheidung offene Features: eingefroren.** v1.0.0 friert den heutigen Funktionsumfang ein. Kein offenes Roadmap-Feature ist Pflicht (K-13, K-15, K-23, R-02, R-05, F-08, F-09 sind alle „Should"/„Nice", keins gesetzlich erforderlich); die günstigen Wins (K-23, R-05, F-08) und F-09 ziehen als **v1.1** nach. Fokus von 1.0.0: Stabilität, Korrektheit, Qualität, Robustheit, Usability, Compliance — keine neuen Features.

> Die Memory-Notiz „jotti v0 Beta, Breaking Changes ok" (Schema direkt in `01_initial.up.sql` editieren) wird mit diesem Release **obsolet**. Ab 1.0.0 gilt die Migrations-Disziplin aus Gate 4.

---

## 1. Gate: Code- und Qualitätssicherung

- [ ] `make verify` grün (= `check-tools` + `check-full`, inkl. Backend, Relay, Starter, Resolver, Local-Proxy, Frontend, **und** Integrationstests gegen echte DB).
- [ ] `make lint-backend-full` (golangci-lint) ohne Befund; Frontend-Lint mit `--max-warnings=0`.
- [ ] `/code-audit` über das Repo laufen lassen, Befunde abarbeiten.
- [ ] `/cleanup` auf die zuletzt geänderten Bereiche (Slop, Boundaries, Readability).
- [ ] Fiskalisch kritische Pfade mit erhöhter Test-Sorgfalt bestätigt: `domain/steuer` (19/7/0/kombi), `domain/kasse` TSE-processData (`tse_processdata_test.go`), `api/fiskal/dsfinvk`, Storno (geldwirksam + geldneutral), Umbuchung, Kassenabschluss/Signatur-Gate.
- [ ] Keine offenen `TODO`/`FIXME`/`XXX` in fiskalisch relevanten Modulen (grep, bewusst abhaken).

---

## 2. Gate: Fiskal-End-to-End-Validierung (der eigentliche v1.0-Kern)

Ziel: der „compliant"-Anspruch ruht nicht auf der Doku, sondern auf einem real durchgespielten Durchlauf mit einer fiskaly-**TEST**-TSE. Erst danach optional gegen LIVE.

### 2.1 TSE-Inbetriebnahme (F-13)
- [ ] fiskaly-TEST-Konto anlegen; Setup-Wizard im Admin-Bereich real durchlaufen (TSS + Client von jotti anlegen lassen, nicht im Anbieter-Dashboard).
- [ ] TEST→LIVE-Umschaltung im Wizard geprüft; PUK/PIN-Verwahrung dokumentiert (Betreiber-Leitfaden).
- [ ] TSE-Stammdaten werden beim Setup persistiert (Signaturalgorithmus, Public Key, Zertifikat, Seriennummer) — Voraussetzung für `tse.csv` im Export (compliance §6.7.7).

### 2.2 Signatur-Outbox im Normalbetrieb (F-02, §3.8)
- [ ] Für jeden signaturpflichtigen Vorgangstyp je einen Fall erzeugen und die Signatur prüfen: Bestellung, Teil-/Vollzahlung, Direktverkauf, Direktverkauf-Storno, Storno bezahlter Positionen (Warenrücknahme), geldneutrale Korrektur, Umbuchung, Geldtransit, Kassendifferenz, Tagesabschluss.
- [ ] Jeder Signaturauftrag wird vom Worker abgeschlossen; Latenz beobachten (Zusage p95 < 5 s).
- [ ] processType-Mapping stichprobenhaft gegen compliance §3.3 verifiziert (Bestellung-V1 / Kassenbeleg-V1 / SonstigerVorgang).

### 2.3 Ausfallsicherheit und Nachsignierung (F-14, §3.8)
- [ ] TSE künstlich stören (falscher Key / Netzwerk blockiert): Vorgänge bleiben gebucht, Buchen wartet nicht auf die TSE.
- [ ] Störungsprotokoll (`tse_stoerungen`) erfasst den Zeitraum mit Grund.
- [ ] Nach Wiederherstellung: automatische Nachsignierung; Beleg trägt „Nachsigniert am …".
- [ ] Vorgang ohne konfigurierte TSE trägt „keine TSE konfiguriert"; im Export `TSE_TA_FEHLER`-Zeile.
- [ ] Kassenabschluss-Verhalten: 409 mit Anzahl/Alter bei **frisch** ausstehenden Signaturen; **erlaubt** bei dokumentiertem Ausfall / fehlender Konfiguration, mit Ausweisung in der Abschlussmeldung.

### 2.4 Belegausgabe (F-03, §5)
- [ ] Kassenbeleg enthält alle Pflichtangaben (§5.2): Vereinsname/Anschrift, Datum, Positionen mit Steuerkennzeichen, Steuermatrix je Satz, Bon-Nr.
- [ ] TSE-Block gedruckt: Start-/Endzeit, TSE- und Kassen-Seriennummer, Transaktionsnummer, Signaturzähler, Signatur.
- [ ] QR-Code aus fiskalys `qr_code_data` als nativer ESC/POS-QR gedruckt (DSFinV-K Anhang I).
- [ ] Durchbedienen: Tisch-Beleg druckt den Startzeitpunkt der ersten Bestellung in Klarschrift (§5.3).
- [ ] Kombi 70/30 korrekt auf dem Beleg ausgewiesen (7-%- und 19-%-Anteil).
- [ ] Reihenfolge zwingend: TSE-`FinishTransaction` vor Druckbefehl (§5.6 Variante A).

### 2.5 DSFinV-K-Export (F-04, §6)
- [ ] Export-ZIP je Kassensitzung wird erzeugt.
- [ ] Struktur gegen die offizielle Spezifikation geprüft — Referenzmaterial liegt im Repo: `docs/rechtsquellen/technik-spezifikationen/DSFinV-K-2.4/` (inkl. `02_index.xml/` und `gdpdu-01-09-2004.dtd`).
- [ ] Dateinamen englisch und kleingeschrieben (`transactions.csv`, `lines.csv`, `cashregister.csv`, `tse.csv`, …); keine deutschen Namen.
- [ ] CSV-Regeln: Semikolon, CRLF, Komma-Dezimaltrennung, Header-Zeile, Spaltenreihenfolge exakt.
- [ ] `index.xml` deklariert nur vorhandene Tabellen; DTD beiliegend.
- [ ] Storno-Abbildung: neue Datensätze mit Negativbetrag, `REF_BON_ID` in `references.csv`, `BON_STORNO = 0` (compliance §6.6).
- [ ] Abrechnungskreis pro Tisch/Kassensitzung in `allocation_groups.csv`.
- [ ] kombi-Aufteilung erscheint in `lines_vat.csv` und `transactions_vat.csv` (zwei Steueranteile).
- [ ] `tse.csv` gefüllt (Zertifikats-ID, Algorithmus, TSE-Seriennummer, Public Key).
- [ ] `BEDIENER_NAME` = eingefrorener Benutzername (nicht Klarname), `BEDIENER_ID` = user_id (§6.4).
- [ ] Möglichst gegenlesen mit IDEA oder dem fiskaly-Prüftooling; mindestens gegen die DSFinV-K-2.4-Beispiele plausibilisieren.

### 2.6 Steuersätze (F-07)
- [ ] Jedes Produkt trägt einen korrekten Steuerschlüssel (19 % Getränke, 7 % Speisen, 0 % Zweckbetrieb §67a, kombi 70/30).
- [ ] Steuerbeträge auf Beleg und im Export stimmen mit `domain/steuer` überein; Rundung geprüft (Cent).

---

## 3. Gate: Betrieb / Produktionsreife (Ops)

- [ ] Ops-Härtung (PRD Runde 1 / PR #64) verifiziert: Version-Pinning-Hard-Fail, non-root-Container, Log-Rotation, Health-Check, Ping-URL.
- [ ] `make prod-update` Roundtrip auf einem Test-Server: Pre-Update-Backup → Images ziehen → Migration → Health-Check → Rollback-Anleitung greift.
- [ ] `make prod-backup` + `make prod-backup-verify`: Backup lässt sich in Wegwerf-Postgres wieder einspielen. **Kritisch:** das Backup enthält die Kassen-Seriennummer (Seriennummern-Kontinuität, §3.7).
- [ ] `make prod-restore` destruktiv einmal geprüft (mit Bestätigung).
- [ ] TLS/Let's Encrypt live grün (lokale LAN-Infra bereits E2E verifiziert — Regressionscheck genügt).
- [ ] Rate Limiting am Login (Q-07, HTTP 429) und Security Headers (Q-08: CSP, HSTS, X-Frame-Options, X-Content-Type-Options) geprüft.
- [ ] Windows-Pfad: `make release-windows` (VERSION=1.0.0) baut `jotti-start.exe` + `jotti-relay.exe` + Release-ZIP unter `dist/`; auf einem Windows-Rechner smoke-getestet.
- [ ] Frische Installation: `make prod-init` auf einem leeren Server läuft bis zum ersten Admin-Login (Initial-Admin-Einmalpasswort-Pfad).
- [ ] Mehrbenutzer/OCC (Q-02): ein grober Parallelzugriffstest (zwei Servicekräfte, gleicher Tisch) ohne Dateninkonsistenz.

---

## 4. Gate: Schema- und Migrations-Disziplin ab v1.0.0

Der Bruch mit der v0-Praxis. **Entschieden: forward-only, keine Down-Migrationen.** Prod-Rollback = Backup-Restore (bereits der `prod-update`-Weg), nie `migrate down`. Grund: fiskalisch append-only (Radierverbot, 10 Jahre) — ein destruktives `down` auf Prod wäre ein Footgun, und Backups sind der ehrliche Rollback-Pfad.

- [x] **Forward-only umgesetzt:** `01_initial.down.sql` entfernt; die up→down→up-Roundtrip-Prüfung aus `scripts/test-integration.sh` und `.github/workflows/ci.yml` entfernt. golang-migrate läuft nur noch `up`.
- [ ] **Freeze bestätigt:** `01_initial.up.sql` wird ab dem Tag nicht mehr editiert. Der Stand, mit dem v1.0.0 ausgeliefert wird, ist der eingefrorene Ausgangspunkt.
- [ ] **Migrations-Konvention dokumentiert** (`database/migrations/README.md` + `AGENTS.md`): neue Änderungen nur als `02_<name>.up.sql`, fortlaufend nummeriert, additiv, vorwärtskompatibel, **kein** `.down.sql`. Jede Migration in einer Transaktion (Postgres-DDL ist transaktional), damit ein Fehlschlag sauber zurückrollt und keinen `dirty`-Zustand hinterlässt. golang-migrate `up` läuft beim Deploy über das `jotti-migrate`-Image.
- [ ] **Migrations-CI-Tests** (ersetzen den wegfallenden Roundtrip, wertvoller): (a) `up` auf **leerer** DB → Frischinstallation bootet; (b) `up` auf einer mit Vorversions-Daten befüllten DB → App bootet **und** `make rebuild-projections` läuft fehlerfrei (der echte Upgrade-Pfad). (a) existiert faktisch schon im Integrationstest; (b) als neuen CI-Job anlegen.
- [ ] **Projektions-Rebuild** (`make rebuild-projections`) läuft nach dem eingefrorenen Schema fehlerfrei; bleibt nach jeder künftigen Migration Pflichtprüfung.
- [ ] **Breaking-Change-Prozess** notiert: DB-Schema, Event-JSON und HTTP-API brechen nur mit Major-Bump; Event-Änderungen additiv als neue `:vN`.

---

## 5. Gate: Dokumentation und Recht

- [ ] **Verfahrensdokumentation** (`docs/verfahrensdokumentation.md`, F-11) gegen den real ausgelieferten Stand geprüft: Softwarename/-version, TSE-Latenzzusage (p95 < 5 s), Datenmodell, Aufbewahrung.
- [ ] `compliance.md`, `handbuch.md`, `anforderungen.md`, `language.md` konsistent mit dem ausgelieferten Funktionsumfang (nach der F-12-Streichung erneut querlesen).
- [ ] Betreiber-Leitfaden aktuell: TSE einrichten, Datenaufbewahrung, Backups/Aktualisieren.
- [ ] **Software-Version konsistent** an allen fiskalisch sichtbaren Stellen: DSFinV-K `cashregister.csv` (Softwaretyp/-version) und ELSTER-Payload melden „jotti 1.0.0". Sicherstellen, dass die Version zentral gepflegt ist und nicht dreifach driftet.
- [ ] **CHANGELOG.md anlegen** (existiert noch nicht) und ab 1.0.0 pflegen; 1.0.0-Eintrag mit dem Funktionsumfang-Stand.
- [ ] `lizenzmodell.md` / README: 1.0-Aussage und Source-Available-Bedingungen final.
- [ ] Muster-Verfahrensdokumentation als anpassbare Vorlage im Repo vorhanden (Betreiberpflicht bleibt beim Verein).

---

## 6. Gate: Release-Mechanik

- [ ] **Version-Bump auf 1.0.0** an allen Quellen: Image-Tags/`JOTTI_VERSION`, `VERSION` für Windows-Build, Frontend `package.json`, die im Backend eingebrannte Software-Version (die in `tse.csv`/`cashregister.csv`/ELSTER fließt).
- [ ] Images bauen und pushen: `ghcr.io/nicograef/jotti` **und** `ghcr.io/nicograef/jotti-migrate` mit Tag `1.0.0` (die Migrationen werden ins Migrate-Image gebacken — beide müssen zusammenpassen).
- [ ] `make release-windows VERSION=1.0.0` → Release-ZIP.
- [ ] **Smoke-Test der gepinnten 1.0.0-Images** auf frischem Server (nicht der Build-Host): `prod-init`, erster Login, ein Verkauf, ein Beleg, ein Export.
- [ ] `git tag v1.0.0` (annotated) + GitHub Release mit Notes (Funktionsumfang, Betreiberpflichten-Hinweis, Upgrade-Hinweis).
- [ ] Nach dem Merge: diese Plan-Datei aus `docs/plans/` entfernen.

---

## 7. Go-Live-Checkliste (kompakt)

- [ ] `make verify` grün, Lint sauber, Audit/Cleanup durch
- [ ] Fiskal-E2E mit fiskaly-TEST-TSE komplett durchgespielt (Signieren, Ausfall+Nachsignieren, Beleg+QR, DSFinV-K-Export validiert)
- [ ] Ops: prod-update/-backup/-restore/-backup-verify getestet, TLS + Security Headers + Rate Limiting grün
- [ ] Schema eingefroren, Migrations-Konvention + Down-Politik dokumentiert
- [ ] Doku konsistent, Verfahrensdoku final, CHANGELOG angelegt, Software-Version überall 1.0.0
- [ ] Images (app + migrate) gebaut/gepusht, Smoke-Test auf frischem Server, `v1.0.0` getaggt + Release veröffentlicht

---

## Bewusst nicht Teil von v1.0.0

- Automatisierte ELSTER-Meldung (ERiC/API) — dauerhaftes Nicht-Ziel (siehe [anforderungen.md, Nicht-Ziele](../anforderungen.md#nicht-ziele)).
- Offene Roadmap-Features K-13, K-15, K-23, R-02, R-05, F-08, F-09 — kein v1.0-Blocker; als 1.x geplant.
