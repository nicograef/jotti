# Rest-Guide: Manuelle QA vor v1.0.0

Handlungsleitfaden für Nico. Enthält nur noch, was keine Suite automatisiert: physische Hardware, der echte Windows-Rechner, das fiskaly-Konto samt TEST→LIVE-Umschaltung und PUK/PIN-Verwahrung, destruktive Ops-Schritte, TLS-Abnahme, Usability mit echten Vereinshelfern und die Abnahme-Entscheidungen selbst. Alles Automatisierbare liegt in einer dieser Suiten:

- **E2E-Suite** (`make test-e2e`, `e2e/`; CI-Job `e2e`): Service- und Admin-Kernflows, Fehlerpfade, Export-Download, Handy-Viewport.
- **DSFinV-K-Validator** (`backend/dsfinvkpruefung`, läuft in `make verify`): Struktur- und Inhaltsregeln des Exports.
- **Berechtigungs-Matrix** (`backend/app/matrix_integration_test.go`, läuft in `make verify`): jede Route gegen jede Rolle und fremde Objekt-IDs.
- **Parallelzugriffstest** (`backend/api/kasse/tischgeschaeft/application/parallelzugriff_integration_test.go`, läuft in `make verify`): Datenkonsistenz bei gleichzeitiger Bedienung.
- **Repo-Gates** (`make check-repo`, läuft in `make check`; CI-Job `repo-checks`): Build-Tags, Sprache und Zeichen, Prosa nach Regel 18, Verweise, Zeitzonen, Versions-Pins, UI-Label-Zitate, Domain-Enums, E2E-Assertions.
- **Schwachstellen-Scans** (`.github/workflows/security-scans.yml`): govulncheck, pnpm audit.
- **Fuzz-Targets** (`make fuzz`, Seed-Korpus läuft in `make test`): Event-Replay, DSFinV-K-CSV, ESC/POS.
- **TSE-Live-Suite** (`make test-tse-live`): alle Geschäftsvorfälle real gegen fiskaly-TEST signiert, Ausfall/Nachsignierung, p95-Latenzmessung.
- **Ops-Smoke** (`scripts/ops-smoke.sh install|ops|release`): Erstinstallation, Backup/-Verifikation, Update-Roundtrip, Security-Header, Rate-Limiting, Release-Smoke.

Die manuelle Release-Vorbereitung (Versions-Bump, Tag) steht unten als Block H, der Release-Smoke als Block I. Die Abnahme läuft über die Abnahme-Entscheidungen am Ende dieses Guides.

## Vorbereitung

- [ ] fiskaly-TEST-Konto angelegt (Zugangsdaten griffbereit, `.env.fiskaly-test` nach `.env.fiskaly-test.example` befüllt — ohne diese Datei bricht `make test-tse-live` ab)
- [ ] 80-mm-Bondrucker (ESC/POS) angeschlossen
- [ ] Zwei Endgeräte (Handys/Tablets) für den Zwei-Geräte-Test in echt
- [ ] Ein echter Windows-Rechner für den Starter-Smoke-Test
- [ ] Frischer Server/VM für den destruktiven Restore-Test und die TLS-Abnahme (nicht der Build-Host, nicht der Dev-Rechner)

## Block A: Hardware und Beleg

- [ ] QR-Code auf echtem 80-mm-Drucker gedruckt und mit dem Handy scannbar — die dynamische Modulgröße ist gegen echte fiskaly-Payload (~350–470 Byte) zu prüfen. Byteform und Pflichtangaben deckt bereits die ESC/POS-Formatter-Testsuite ab; hier zählt nur das physische Druckbild.
- [ ] Bondrucker-Druckbild insgesamt sichtprüfen (Lesbarkeit, Schnitt, Papiervorschub) — das ist mit realer Hardware nicht automatisierbar.

## Block B: Windows-Rechner

- [ ] `make release-windows VERSION=…` baut `jotti-start.exe` + `jotti-relay.exe` + Release-ZIP unter `dist/`; ZIP auf dem echten Windows-Rechner entpacken und bis zum ersten Login smoke-testen. Der API-Roundtrip auf Linux-Hosts läuft bereits über `scripts/ops-smoke.sh install`; hier zählt nur der reale Windows-Start.

## Block C: fiskaly-Konto und TSE-Inbetriebnahme

- [ ] Setup-Wizard im Admin-Bereich real durchlaufen: TSS und Client von jotti anlegen lassen, nicht im fiskaly-Dashboard. Der TSS-anlegende Testlauf `make test-tse-live-setup` ist dafür bewusst nicht geeignet, siehe unten.
- [ ] TEST→LIVE-Umschaltung im Wizard geprüft; PUK/PIN-Verwahrung dokumentiert (Betreiber-Leitfaden). PUK/PIN existieren nur im fiskaly-Konto, nicht in einer Suite reproduzierbar.
- [ ] Signaturbetrieb, Ausfall/Nachsignierung und Latenzmessung sind durch `make test-tse-live` abgedeckt (alle Geschäftsvorfälle, Störungsprotokoll, Nachsignierung, Abschluss-Gate, p95-Messung). Hier nur gegenlesen, ob die Verfahrensdokumentation die gemessene Latenz korrekt wiedergibt.
- [ ] Vorsicht bei `make test-tse-live-setup` bzw. dem Test `TestFiskalySetup_LiveVollerDurchlauf`: legt eine unlöschbare TSS im fiskaly-TEST-Konto an (begrenztes Kontolimit). Nur bewusst ausführen, wenn eine neue TSS gebraucht wird, nicht Teil des normalen QA-Durchlaufs.

## Block D: DSFinV-K-Gegenlesen

- [ ] Export möglichst mit IDEA oder fiskaly-Prüftooling gegenlesen; mindestens gegen die DSFinV-K-2.4-Beispiele plausibilisieren. Struktur- und Inhaltsregeln (Dateinamen, CSV-Form, index.xml/DTD, Storno-Referenzen, Kombi-Aufteilung, Bediener-Felder, TSE-Stammdaten, Tagesabschluss-Zeile, Steuersätze) prüft automatisch der DSFinV-K-Validator in `make verify`; hier zählt nur der externe Tooling-Abgleich als zusätzliche Absicherung.

## Block E: Ops — destruktiv und TLS

- [ ] `make prod-restore` destruktiv einmal geprüft (mit Bestätigung). Bewusst nicht in `scripts/ops-smoke.sh`, weil destruktiv.
- [ ] TLS/Let's Encrypt live grün auf dem produktiv genutzten Host (lokale LAN-Infra bereits E2E verifiziert, hier nur Regressionscheck gegen den echten Domain-Namen). `prod-init`/`prod-update`/`prod-backup`/`prod-backup-verify` inkl. Security-Header und Rate-Limiting deckt `scripts/ops-smoke.sh install|ops` ab.

## Block F: Zwei-Geräte-Test in echt

- [ ] Zwei echte Servicekraft-Handys am selben Tisch parallel bedienen, ohne Dateninkonsistenz. Die Datenkonsistenz-Zusage selbst prüft der Parallelzugriffstest in `make verify`; hier zählt nur der reale Zwei-Geräte-Eindruck (Latenz, UI-Reaktion, echtes WLAN).

## Block G: Usability mit Vereinshelfern

- [ ] Mindestens eine Servicekraft und ein Admin aus einem echten Verein die Kernflows ohne Anleitung bedienen lassen; Stolperstellen notieren. Nicht durch die E2E-Suite oder eine heuristische UX-Review ersetzbar, weil es um echte Erstnutzer-Reaktionen geht.

## Block H: Release schneiden

Voraussetzung: Die code-seitigen v1.0-Blocker sind gemergt (zog-Validierung der persistierten Feldgrenzen, repo-weites Prettier-Gate, entfernte lokale DB-Wipe-Fähigkeit, `CHANGELOG.md`), die Blöcke A–G sind abgenommen und die Go/No-Go-Entscheidung (siehe unten) ist getroffen.

- [ ] Beispielversion in `docs/leitfaden/self-hosting.md` (`JOTTI_VERSION=v0.14.0`) auf `v1.0.0` heben. Nur diese Datei ist betroffen: `docker-compose.release.yml` ist ein `:RELEASE_VERSION`-Template (der Release-Workflow ersetzt den Platzhalter), `.env.example` hält `JOTTI_VERSION=` bewusst leer, die Verfahrensdokumentation trägt an dieser Stelle einen Betreiber-Platzhalter (`«z. B. v0.2.0»`) und `frontend/package.json` steht auf `0.0.0`, das nirgends im Build gelesen wird.
- [ ] CI auf dem Release-Commit in `main` grün: die Jobs `backend-ci`, `backend-golangci`, `repo-checks`, `frontend-ci`, `backend-integration-tests`, `e2e` und `upgrade-path` decken `make verify` und `make lint-backend-full` ab.
- [ ] Version-Bump auf 1.0.0 (Image-Tags/`JOTTI_VERSION`, `VERSION` für den Windows-Build, ldflags-Version) und Tag `v1.0.0` pushen — `release.yml` baut Images, Windows-ZIP und Release-Notes (git-cliff) und veröffentlicht das GitHub-Release.

## Block I: Release-Smoke v1.0.0

- [ ] `bash scripts/ops-smoke.sh release v1.0.0` auf frischem Server (nicht der Build-Host) mit den gepinnten 1.0.0-Images: prüft `prod-init`, ersten Login, einen Verkauf, einen Beleg, einen Export automatisiert. Hier nur das Ergebnis abnehmen.

## Abnahme-Entscheidungen

- [ ] Alle Suiten-Läufe grün: `make verify` (DSFinV-K-Validator, Berechtigungs-Matrix, Parallelzugriffstest, Fuzz-Seed-Korpus, Repo-Gates), `make test-e2e`, `make test-tse-live`, `scripts/ops-smoke.sh install|ops`, Schwachstellen-Scans in `.github/workflows/security-scans.yml`.
- [ ] Blöcke A–G dieses Guides durchgespielt und abgenommen.
- [ ] Go/No-Go-Entscheidung für den v1.0.0-Tag getroffen — danach Block H (Release schneiden) und Block I (Release-Smoke) durchführen und abnehmen.
