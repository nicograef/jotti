# Release-Guide: jotti v1.0.0

Übersicht, wann `v1.0.0` getaggt werden darf. Die Arbeit selbst liegt in drei Dokumenten: Breaking-Arbeit vor der Erstinstallation in [plan-v0.14.0-breaking.md](plan-v0.14.0-breaking.md), automatisierbare Nacharbeit in [plan-v1.0.0-nacharbeit.md](plan-v1.0.0-nacharbeit.md), verbleibende Handarbeit im [Rest-Guide](guide-manuelle-qa-v1.0.0.md) (der Großteil der ursprünglichen manuellen QA läuft inzwischen über automatisierte Suiten, siehe [PRD QA-Automatisierung](../prds/prd-qa-automatisierung.md)). Befund-Details im [Audit-Bericht](audit-v1.0.0.md) (Register, Status wird in den Plänen geführt). Arbeitsdokument; nach dem Release aus `docs/plans/` entfernen.

---

## 0. Was v1.0.0 bedeutet (Grundsatz-Entscheidungen)

Ab 1.0.0 gilt SemVer im vollen Sinn. Diese Zusagen müssen vor dem Tag bewusst getroffen und dokumentiert sein, weil sie danach nur noch mit Major-Bump gebrochen werden dürfen:

- [ ] **Stabiles DB-Schema.** `01_initial.up.sql` ist eingefroren. Jede spätere Änderung ist eine neue, additive Migration (`02_*.up.sql`, forward-only, kein `.down.sql` — siehe Gate 4). De facto gilt der Freeze bereits seit der Erstinstallation des ersten Vereins (2026-07-07), nicht erst ab dem Tag.
- [ ] **Stabile Event-Contracts.** Die JSON-Form der Kassenjournal-Events (`bestellung-aufgenommen:v1`, `zahlung-kassiert:v1`, …) ist eingefroren; Guard ist `backend/domain/kasse/event_json_contract_test.go`. Änderungen an einem Event nur als neue Version (`:v2`), nie in-place. Grund: Events sind 10 Jahre aufzubewahren und müssen replaybar bleiben. Auch hier: de facto seit der Erstinstallation.
- [ ] **Stabile HTTP-API.** Endpunkte und Payloads, auf die das Print-Relay und das Frontend bauen, bleiben rückwärtskompatibel. Letzte Änderung vor dem Freeze: Nacharbeit Block 1 (B4/B6).
- [x] **Scope-Entscheidung offene Features: eingefroren.** v1.0.0 friert den heutigen Funktionsumfang ein. Kein offenes Roadmap-Feature ist Pflicht (K-13, K-15, K-23, R-02, R-05, F-08, F-09 sind alle „Should"/„Nice", keins gesetzlich erforderlich); die günstigen Wins (K-23, R-05, F-08) und F-09 ziehen als **v1.1** nach. Fokus von 1.0.0: Stabilität, Korrektheit, Qualität, Robustheit, Usability, Compliance — keine neuen Features.

> Die Memory-Notiz „jotti v0 Beta, Breaking Changes ok" (Schema direkt in `01_initial.up.sql` editieren) ist mit der Erstinstallation **obsolet**. Es gilt die Migrations-Disziplin aus Gate 4.

---

## 1. Gate: Code- und Qualitätssicherung

- [ ] `make verify` grün (= `check-tools` + `check-full`, inkl. Backend, Relay, Starter, Resolver, Local-Proxy, Frontend, **und** Integrationstests gegen echte DB) auf dem Release-Commit.
- [ ] `make lint-backend-full` (golangci-lint) ohne Befund; Frontend-Lint mit `--max-warnings=0`.
- [x] Code-Audit: erledigt durch das Multi-Experten-Audit vom 2026-07-06 (`audit-v1.0.0.md`); die Befunde sind auf die Pläne verteilt.
- [ ] Finaler Cleanup-/Review-Pass und TODO/FIXME-Grep in fiskalischen Modulen (Nacharbeit Block 6).

---

## 2. Gate: Fiskal-End-to-End-Validierung (der eigentliche v1.0-Kern)

Der „compliant"-Anspruch ruht nicht auf der Doku, sondern auf einem real durchgespielten Durchlauf mit einer fiskaly-**TEST**-TSE. Erst danach optional gegen LIVE.

- [ ] TSE-Live-Suite (`make test-tse-live`) grün: alle Geschäftsvorfälle, Ausfall und Nachsignierung, p95-Latenzmessung. DSFinV-K-Validator (`make verify`) grün: Struktur, Storno, Steuersätze, Bediener- und TSE-Felder.
- [ ] Komplett durchgespielt nach [Rest-Guide](guide-manuelle-qa-v1.0.0.md), Blöcke A und C (QR-Scan/Druckbild auf echter Hardware, fiskaly-Konto samt TEST→LIVE-Umschaltung und PUK/PIN-Verwahrung) und Block D (DSFinV-K-Gegenlesen mit IDEA/fiskaly-Prüftooling).

---

## 3. Gate: Betrieb / Produktionsreife (Ops)

- [ ] Ops-Härtung (PRD Runde 1 / PR #64) verifiziert: Version-Pinning-Hard-Fail, non-root-Container, Log-Rotation, Health-Check, Ping-URL.
- [ ] Ops-Smoke (`scripts/ops-smoke.sh install` und `ops`) grün: frische Installation, prod-backup/-backup-verify/-update, Security-Header, Rate-Limiting. Parallelzugriffstest (`make verify`) grün.
- [ ] Komplett durchgespielt nach [Rest-Guide](guide-manuelle-qa-v1.0.0.md), Block B (Windows-Rechner), Block E (destruktives `prod-restore`, TLS-Abnahme auf echtem Domain-Namen) und Block F (Zwei-Geräte-Test in echt).

---

## 4. Gate: Schema- und Migrations-Disziplin

Der Bruch mit der v0-Praxis. **Entschieden: forward-only, keine Down-Migrationen.** Prod-Rollback = Backup-Restore (bereits der `prod-update`-Weg), nie `migrate down`. Grund: fiskalisch append-only (Radierverbot, 10 Jahre) — ein destruktives `down` auf Prod wäre ein Footgun, und Backups sind der ehrliche Rollback-Pfad.

- [x] **Forward-only umgesetzt:** `01_initial.down.sql` entfernt; die up→down→up-Roundtrip-Prüfung aus `scripts/test-integration.sh` und `.github/workflows/ci.yml` entfernt. golang-migrate läuft nur noch `up`.
- [ ] **Freeze bestätigt:** `01_initial.up.sql` wird nicht mehr editiert; letzter Edit ist Phase 2 des [Breaking-Plans](plan-v0.14.0-breaking.md), vor der Erstinstallation.
- [ ] **Migrations-Konvention dokumentiert** (`database/migrations/README.md` + `AGENTS.md`): neue Änderungen nur als `02_<name>.up.sql`, fortlaufend nummeriert, additiv, vorwärtskompatibel, **kein** `.down.sql`. Jede Migration in einer Transaktion (Postgres-DDL ist transaktional), damit ein Fehlschlag sauber zurückrollt und keinen `dirty`-Zustand hinterlässt. golang-migrate `up` läuft beim Deploy über das `jotti-migrate`-Image. (AGENTS.md-Umschrieb: Nacharbeit Block 5, C20.)
- [ ] **Migrations-CI-Tests:** (a) `up` auf leerer DB → Frischinstallation bootet (existiert im Integrationstest); (b) `up` auf befüllter Vorversions-DB → App bootet und `make rebuild-projections` läuft fehlerfrei (Nacharbeit Block 4, C10).
- [ ] **Projektions-Rebuild** (`make rebuild-projections`) läuft nach dem eingefrorenen Schema fehlerfrei; bleibt nach jeder künftigen Migration Pflichtprüfung.
- [ ] **Breaking-Change-Prozess** notiert: DB-Schema, Event-JSON und HTTP-API brechen nur mit Major-Bump; Event-Änderungen additiv als neue `:vN`.

---

## 5. Gate: Dokumentation und Recht

Die Arbeit liegt in Nacharbeit Block 5 und 6; hier nur die Abnahme:

- [ ] **Verfahrensdokumentation** (`docs/verfahrensdokumentation.md`, F-11) gegen den real ausgelieferten Stand geprüft: Softwarename/-version, TSE-Latenzzusage (p95 < 5 s, ggf. nach Messung angepasst), Datenmodell, Aufbewahrung.
- [ ] `compliance.md`, `handbuch.md`, `anforderungen.md`, `language.md` konsistent mit dem ausgelieferten Funktionsumfang.
- [ ] Betreiber-Leitfaden aktuell: TSE einrichten, Datenaufbewahrung, Backups/Aktualisieren, Kassenmeldepflicht-Frist.
- [ ] **Software-Version konsistent** an allen fiskalisch sichtbaren Stellen: DSFinV-K `cashregister.csv` meldet die Build-Version (C11-Fix), Admin-Footer zeigt sie (C12-Fix); zentral gepflegt über ldflags, driftet nicht.
- [ ] **CHANGELOG.md** existiert mit 1.0.0-Eintrag (Funktionsumfang-Stand).
- [ ] `lizenzmodell.md` / README: 1.0-Aussage und Source-Available-Bedingungen final.
- [ ] Muster-Verfahrensdokumentation als anpassbare Vorlage im Repo vorhanden (Betreiberpflicht bleibt beim Verein).

---

## 6. Gate: Release-Mechanik

- [ ] **Version-Bump auf 1.0.0** an allen Quellen: Image-Tags/`JOTTI_VERSION`, `VERSION` für Windows-Build, die per ldflags eingebrannte Software-Version (fließt in `tse.csv`/`cashregister.csv`). `frontend/package.json` ist tot und gehört nicht zum Bump-Set (C13-Entscheidung).
- [ ] Images bauen und pushen: `ghcr.io/nicograef/jotti-backend`, `jotti-frontend`, `jotti-migrate`, `jotti-reverse-proxy`, alle mit Tag `1.0.0` (Migrationen sind ins Migrate-Image gebacken — alle vier müssen zusammenpassen).
- [ ] `make release-windows VERSION=1.0.0` → Release-ZIP.
- [ ] **Smoke-Test der gepinnten 1.0.0-Images**: `scripts/ops-smoke.sh release VERSION=1.0.0` (`prod-init`, erster Login, ein Verkauf, ein Beleg, ein Export), Abnahme nach [Rest-Guide](guide-manuelle-qa-v1.0.0.md), Block H.
- [ ] `git tag v1.0.0` (annotated) + GitHub Release mit Notes (Funktionsumfang, Betreiberpflichten-Hinweis, Upgrade-Hinweis ab v0.14.0).
- [ ] Nach dem Release: Plan-Dateien und Audit aus `docs/plans/` entfernen.

---

## 7. Go-Live-Checkliste (kompakt)

- [ ] `make verify` grün, Lint sauber, Cleanup-Pass durch (Gate 1)
- [ ] Fiskal-E2E mit fiskaly-TEST-TSE komplett durchgespielt (Gate 2: TSE-Live-Suite, DSFinV-K-Validator, Rest-Guide Blöcke A, C, D)
- [ ] Ops-Roundtrips und frische Installation getestet (Gate 3: Ops-Smoke, Parallelzugriffstest, Rest-Guide Blöcke B, E, F)
- [ ] Schema eingefroren, Migrations-Konvention dokumentiert (Gate 4)
- [ ] Doku konsistent, Verfahrensdoku final, CHANGELOG angelegt, Version überall 1.0.0 (Gate 5)
- [ ] Images gebaut/gepusht, Smoke-Test auf frischem Server, `v1.0.0` getaggt + Release veröffentlicht (Gate 6)

---

## Bewusst nicht Teil von v1.0.0

- Automatisierte ELSTER-Meldung (ERiC/API) — dauerhaftes Nicht-Ziel (siehe [anforderungen.md, Nicht-Ziele](../anforderungen.md#nicht-ziele)).
- Offene Roadmap-Features K-13, K-15, K-23, R-02, R-05, F-08, F-09 — kein v1.0-Blocker; als 1.x geplant.
- Audit-Abschnitt D (soweit nicht in den Plänen): nach 1.0 ohne Bruch nachrüstbar, Liste am Ende von [plan-v1.0.0-nacharbeit.md](plan-v1.0.0-nacharbeit.md).
