# Plan: Release v1.0.0

**Ziel:** jotti 1.0.0 veröffentlichen — manuelle QA abnehmen, Version heben, Tag setzen, danach
die Nachlaufarbeiten erledigen.

**Betroffene Dateien:** `CHANGELOG.md`, `docs/leitfaden/self-hosting.md`,
`docs/leitfaden/aktualisieren.md`, `.github/workflows/ci.yml`, `docs/plans/`.

**Reihenfolge = Abhängigkeit.** Die Abschnitte 4 und 5 sind kein Release-Blocker.

## 1. Manuelle QA

Nur das Nicht-Automatisierbare: physische Hardware, der echte Windows-Rechner, das fiskaly-Konto
samt TEST→LIVE-Umschaltung und PUK/PIN-Verwahrung, destruktive Ops-Schritte, TLS-Abnahme,
Usability mit echten Vereinshelfern und die Abnahme-Entscheidungen selbst.

### Vorbereitung

- [ ] fiskaly-TEST-Konto angelegt (Zugangsdaten griffbereit, `.env.fiskaly-test` nach
      `.env.fiskaly-test.example` befüllt — ohne diese Datei bricht `make test-tse-live` ab)
- [ ] 80-mm-Bondrucker (ESC/POS) angeschlossen
- [ ] Zwei Endgeräte (Handys/Tablets) für den Zwei-Geräte-Test in echt
- [ ] Ein echter Windows-Rechner für den Starter-Smoke-Test
- [ ] Frischer Server/VM für den destruktiven Restore-Test und die TLS-Abnahme (nicht der
      Build-Host, nicht der Dev-Rechner)

### Block A: Hardware und Beleg

- [ ] QR-Code auf echtem 80-mm-Drucker gedruckt und mit dem Handy scannbar — die dynamische
      Modulgröße ist gegen echte fiskaly-Payload (~350–470 Byte) zu prüfen. Byteform und
      Pflichtangaben deckt bereits die ESC/POS-Formatter-Testsuite ab; hier zählt nur das
      physische Druckbild.
- [ ] Bondrucker-Druckbild insgesamt sichtprüfen (Lesbarkeit, Schnitt, Papiervorschub).

### Block B: Windows-Rechner

- [ ] `make release-windows VERSION=…` baut `jotti-start.exe` + `jotti-relay.exe` + Release-ZIP
      unter `dist/`; ZIP auf dem echten Windows-Rechner entpacken und bis zum ersten Login
      smoke-testen. Der API-Roundtrip auf Linux-Hosts läuft bereits über
      `scripts/ops-smoke.sh install`; hier zählt nur der reale Windows-Start.

### Block C: fiskaly-Konto und TSE-Inbetriebnahme

- [ ] Setup-Wizard im Admin-Bereich real durchlaufen: TSS und Client von jotti anlegen lassen,
      nicht im fiskaly-HUB.
- [ ] TEST→LIVE-Umschaltung im Wizard geprüft; PUK/PIN-Verwahrung dokumentiert
      (Betreiber-Leitfaden). PUK/PIN existieren nur im fiskaly-Konto, nicht in einer Suite
      reproduzierbar.
- [ ] Signaturbetrieb, Ausfall/Nachsignierung und Latenzmessung sind durch `make test-tse-live`
      abgedeckt (alle Geschäftsvorfälle, Störungsprotokoll, Nachsignierung, Abschluss-Gate,
      p95-Messung). Hier nur gegenlesen, ob die Verfahrensdokumentation die gemessene Latenz
      korrekt wiedergibt.

**Vorsicht bei `make test-tse-live-setup` bzw. dem Test `TestFiskalySetup_LiveVollerDurchlauf`:**
legt eine unlöschbare TSS im fiskaly-TEST-Konto an (begrenztes Kontolimit). Nur bewusst ausführen,
wenn eine neue TSS gebraucht wird — nicht Teil des normalen QA-Durchlaufs und kein Ersatz für den
Setup-Wizard-Durchlauf oben.

### Block D: DSFinV-K-Gegenlesen

- [ ] Export möglichst mit IDEA oder fiskaly-Prüftooling gegenlesen; mindestens gegen die
      DSFinV-K-2.4-Beispiele plausibilisieren. Struktur- und Inhaltsregeln (Dateinamen, CSV-Form,
      index.xml/DTD, Storno-Referenzen, Kombi-Aufteilung, Bediener-Felder, TSE-Stammdaten,
      Tagesabschluss-Zeile, Steuersätze) prüft automatisch der DSFinV-K-Validator in
      `make verify`; hier zählt nur der externe Tooling-Abgleich als zusätzliche Absicherung.

### Block E: Ops — destruktiv und TLS

- [ ] `make prod-restore` destruktiv einmal geprüft (mit Bestätigung). Bewusst nicht in
      `scripts/ops-smoke.sh`, weil destruktiv.
- [ ] TLS/Let's Encrypt live grün auf dem produktiv genutzten Host (lokale LAN-Infra bereits E2E
      verifiziert, hier nur Regressionscheck gegen den echten Domain-Namen).
      `prod-init`/`prod-update`/`prod-backup`/`prod-backup-verify` inkl. Security-Header und
      Rate-Limiting deckt `scripts/ops-smoke.sh install|ops` auf einem Testhost ab; beide Läufe
      mit demselben `ADMIN_PASSWORD`, denn `ops` meldet sich an und bucht vor dem Backup einen
      Verkauf.

### Block F: Zwei-Geräte-Test in echt

- [ ] Zwei echte Servicekraft-Handys am selben Tisch parallel bedienen, ohne Dateninkonsistenz.
      Die Datenkonsistenz-Zusage selbst prüft der Parallelzugriffstest in `make verify`; hier
      zählt nur der reale Zwei-Geräte-Eindruck (Latenz, UI-Reaktion, echtes WLAN).

### Block G: Usability mit Vereinshelfern

- [ ] Mindestens eine Servicekraft und ein Admin aus einem echten Verein die Kernflows ohne
      Anleitung bedienen lassen; Stolperstellen notieren. Nicht durch die E2E-Suite oder eine
      heuristische UX-Review ersetzbar, weil es um echte Erstnutzer-Reaktionen geht.

### Abnahme

- [ ] Alle Suiten-Läufe grün: `make verify` (DSFinV-K-Validator, Berechtigungs-Matrix,
      Parallelzugriffstest, Fuzz-Seed-Korpus, Repo-Gates), `make test-e2e`, `make test-tse-live`,
      `scripts/ops-smoke.sh install|ops`, Schwachstellen-Scans in
      `.github/workflows/security-scans.yml`.
- [ ] Blöcke A–G durchgespielt und abgenommen.
- [ ] Go/No-Go-Entscheidung für den v1.0.0-Tag getroffen.

## 2. Release schneiden (Block H, nur nach Go)

- [ ] Beispielversion in `docs/leitfaden/self-hosting.md` (`JOTTI_VERSION=v0.18.0`) auf `v1.0.0`
      heben. Nur diese Datei ist betroffen: `docker-compose.release.yml` ist ein
      `:RELEASE_VERSION`-Template (der Release-Workflow ersetzt den Platzhalter), `.env.example`
      hält `JOTTI_VERSION=` bewusst leer, die Verfahrensdokumentation trägt an dieser Stelle einen
      Betreiber-Platzhalter (`«z. B. v1.0.0»`) und `frontend/package.json` steht auf `0.0.0`, das
      nirgends im Build gelesen wird.
- [ ] `docs/leitfaden/aktualisieren.md` prüfen: die Aussage, dass das Print-Relay in 0.17.1 und allen
      neueren Versionen funktional dasselbe ist, gegen den 1.0.0-Stand halten
- [ ] Release-Datum im Abschnitt `[1.0.0]` der `CHANGELOG.md` eintragen
- [ ] CI auf dem Release-Commit in `main` grün: die Jobs `backend-ci`, `backend-golangci`,
      `repo-checks`, `frontend-ci`, `website-ci`, `shellcheck-ci`, `resolver-ci`, `local-proxy-ci`,
      `windows-ci`, `backend-integration-tests`, `e2e` und `upgrade-path` decken `make verify` und
      `make lint-backend-full` ab.
- [ ] Version-Bump auf 1.0.0 (Image-Tags/`JOTTI_VERSION`, `VERSION` für den Windows-Build,
      ldflags-Version) und Tag `v1.0.0` pushen — `release.yml` baut Images, Windows-ZIP und
      Release-Notes (git-cliff) und veröffentlicht das GitHub-Release.
- [ ] Release-Text (Abschnitt „Release-Text“ unten) vor die git-cliff-Notes stellen; die
      Laufzeit-Versionen (Go, Node, pnpm) müssen darin stehen
- [ ] Block I: `bash scripts/ops-smoke.sh release v1.0.0` auf frischem Server (nicht der
      Build-Host) mit den gepinnten 1.0.0-Images: prüft `prod-init`, ersten Login, einen Verkauf,
      einen Beleg, einen Export automatisiert. Hier nur das Ergebnis abnehmen.

## 3. Nach dem Tag (eigener Commit)

- [ ] `PREVIOUS_VERSION` in `.github/workflows/ci.yml` von `v0.18.0` auf `v1.0.0` heben;
      Job `upgrade-path` grün
- [ ] Diesen Plan löschen, sobald alle Boxen abgehakt sind

## Release-Text

Text des GitHub-Releases vor den git-cliff-Notes:

jotti 1.0.0 ist die erste stabile Version des kostenlosen Mobile-Kassensystems für Vereinsfeste:
Servicekräfte nehmen auf ihren eigenen Smartphones Bestellungen auf, kassieren und stornieren,
Admins verwalten Produkte, Tische und Benutzer. Alle Vorgänge werden von einer BSI-zertifizierten
Cloud-TSE signiert, das Kassenjournal ist append-only und der DSFinV-K-Export liegt bereit.

Alle Funktionen und Änderungen im Detail: Abschnitt `[1.0.0]` in `CHANGELOG.md`.

### Enthaltene Laufzeit-Versionen

| Laufzeit   | Version     | Quelle                                                          |
| ---------- | ----------- | --------------------------------------------------------------- |
| Go         | 1.27.1      | `backend/go.mod`, `backend/Dockerfile` (`golang:1.27.1-alpine`) |
| Node       | 24 (Alpine) | `frontend/Dockerfile` (`node:24-alpine`)                        |
| pnpm       | 11.27.1     | `frontend/package.json` (`packageManager`)                      |
| PostgreSQL | 17.11       | `docker-compose.prod.yml` (`postgres:17.11`)                    |
| Caddy      | 2.11.4      | `reverse-proxy/Dockerfile` (`caddy:2.11.4`)                     |

### Aktualisieren

- **Windows-Rechner (Standardweg):** `jotti-stop.cmd`, neues Release-ZIP entpacken,
  `jotti-start.exe` starten. Der Starter sichert die Datenbank automatisch, bevor er
  aktualisiert. Ablauf, Rauchtest und der Weg zurück: `docs/leitfaden/aktualisieren.md`.
- **Eigener Server (Self-Hosting):** `JOTTI_VERSION=v1.0.0` in `.env` setzen und `make prod-update`
  ausführen. Das Skript zieht vor der Aktualisierung ein Backup. Ablauf und Backup-Strategie:
  `docs/leitfaden/aktualisieren-backups.md`.
