# Rest-Guide: Manuelle QA vor v1.0.0

Handlungsleitfaden für Nico. Enthält nur noch, was keine Suite automatisiert: physische Hardware, der echte Windows-Rechner, das fiskaly-Konto samt TEST→LIVE-Umschaltung und PUK/PIN-Verwahrung, destruktive Ops-Schritte, TLS-Abnahme, Usability mit echten Vereinshelfern und die Abnahme-Entscheidungen selbst. Alles Automatisierbare ist in eine der folgenden Suiten gewandert (siehe [PRD QA-Automatisierung](../prds/prd-qa-automatisierung.md)):

- **E2E-Suite** (`make test-e2e`, `e2e/`): Service- und Admin-Kernflows, Fehlerpfade, Export-Download, Handy-Viewport.
- **DSFinV-K-Validator** (`backend/dsfinvkpruefung`, läuft in `make verify`): Struktur- und Inhaltsregeln des Exports.
- **Berechtigungs-Matrix** (`backend/app/matrix_integration_test.go`, läuft in `make verify`): jede Route gegen jede Rolle und fremde Objekt-IDs.
- **Parallelzugriffstest** (`backend/api/kasse/tischgeschaeft/application/parallelzugriff_integration_test.go`, läuft in `make verify`): Datenkonsistenz bei gleichzeitiger Bedienung.
- **Schwachstellen-Scans** (`.github/workflows/security-scans.yml`): govulncheck, pnpm audit.
- **Fuzz-Targets** (`make fuzz`, Seed-Korpus läuft in `make test`): Event-Replay, DSFinV-K-CSV, ESC/POS.
- **TSE-Live-Suite** (`make test-tse-live`): alle Geschäftsvorfälle real gegen fiskaly-TEST signiert, Ausfall/Nachsignierung, p95-Latenzmessung.
- **Ops-Smoke** (`scripts/ops-smoke.sh install|ops|release`): Erstinstallation, Backup/-Verifikation, Update-Roundtrip, Security-Header, Rate-Limiting, Release-Smoke.

Die Checkboxen der Gates im [Release-Guide](plan-v1.0.0-release.md) bleiben führend; hier steht nur die Handarbeit, die diese Gates noch brauchen.

## Vorbereitung

- [ ] fiskaly-TEST-Konto angelegt (Zugangsdaten griffbereit, `.env.fiskaly-test` für die TSE-Live-Suite befüllt)
- [ ] 80-mm-Bondrucker (ESC/POS) angeschlossen
- [ ] Zwei Endgeräte (Handys/Tablets) für den Zwei-Geräte-Test in echt
- [ ] Ein echter Windows-Rechner für den Starter-Smoke-Test
- [ ] Frischer Server/VM für den destruktiven Restore-Test und die TLS-Abnahme (nicht der Build-Host, nicht der Dev-Rechner)

## Block A: Hardware und Beleg

- [ ] QR-Code auf echtem 80-mm-Drucker gedruckt und mit dem Handy scannbar — die dynamische Modulgröße ist gegen echte fiskaly-Payload (~350–470 Byte) zu prüfen. Byteform und Pflichtangaben deckt bereits die ESC/POS-Formatter-Testsuite ab; hier zählt nur das physische Druckbild.
- [ ] Bondrucker-Druckbild insgesamt sichtprüfen (Lesbarkeit, Schnitt, Papiervorschub) — das ist mit realer Hardware nicht automatisierbar.

## Block B: Windows-Rechner

- [ ] `make release-windows VERSION=…` baut `jotti-start.exe` + `jotti-relay.exe` + Release-ZIP unter `dist/`; ZIP auf dem echten Windows-Rechner entpacken und bis zum ersten Login smoke-testen (Gate 3/6). Der API-Roundtrip auf Linux-Hosts läuft bereits über `scripts/ops-smoke.sh install`; hier zählt nur der reale Windows-Start.

## Block C: fiskaly-Konto und TSE-Inbetriebnahme

- [ ] Setup-Wizard im Admin-Bereich real durchlaufen: TSS und Client von jotti anlegen lassen, nicht im fiskaly-Dashboard. Der TSS-anlegende Testlauf `make test-tse-live-setup` ist dafür bewusst nicht geeignet, siehe unten.
- [ ] TEST→LIVE-Umschaltung im Wizard geprüft; PUK/PIN-Verwahrung dokumentiert (Betreiber-Leitfaden). PUK/PIN existieren nur im fiskaly-Konto, nicht in einer Suite reproduzierbar.
- [ ] Signaturbetrieb, Ausfall/Nachsignierung und Latenzmessung sind durch `make test-tse-live` abgedeckt (alle Geschäftsvorfälle, Störungsprotokoll, Nachsignierung, Abschluss-Gate, p95-Messung). Hier nur gegenlesen, ob die Verfahrensdokumentation die gemessene Latenz korrekt wiedergibt.
- [ ] Vorsicht bei `make test-tse-live-setup` bzw. dem Test `LiveVollerDurchlauf`: legt eine unlöschbare TSS im fiskaly-TEST-Konto an (begrenztes Kontolimit). Nur bewusst ausführen, wenn eine neue TSS gebraucht wird, nicht Teil des normalen QA-Durchlaufs.

## Block D: DSFinV-K-Gegenlesen

- [ ] Export möglichst mit IDEA oder fiskaly-Prüftooling gegenlesen; mindestens gegen die DSFinV-K-2.4-Beispiele plausibilisieren. Struktur- und Inhaltsregeln (Dateinamen, CSV-Form, index.xml/DTD, Storno-Referenzen, Kombi-Aufteilung, Bediener-Felder, TSE-Stammdaten, Tagesabschluss-Zeile, Steuersätze) prüft automatisch der DSFinV-K-Validator in `make verify`; hier zählt nur der externe Tooling-Abgleich als zusätzliche Absicherung.

## Block E: Ops — destruktiv und TLS

- [ ] `make prod-restore` destruktiv einmal geprüft (mit Bestätigung). Bewusst nicht in `scripts/ops-smoke.sh`, weil destruktiv.
- [ ] TLS/Let's Encrypt live grün auf dem produktiv genutzten Host (lokale LAN-Infra bereits E2E verifiziert, hier nur Regressionscheck gegen den echten Domain-Namen). `prod-init`/`prod-update`/`prod-backup`/`prod-backup-verify` inkl. Security-Header und Rate-Limiting deckt `scripts/ops-smoke.sh install|ops` ab.

## Block F: Zwei-Geräte-Test in echt

- [ ] Zwei echte Servicekraft-Handys am selben Tisch parallel bedienen, ohne Dateninkonsistenz. Die Datenkonsistenz-Zusage selbst prüft der Parallelzugriffstest in `make verify`; hier zählt nur der reale Zwei-Geräte-Eindruck (Latenz, UI-Reaktion, echtes WLAN).

## Block G: Usability mit Vereinshelfern

- [ ] Mindestens eine Servicekraft und ein Admin aus einem echten Verein die Kernflows ohne Anleitung bedienen lassen; Stolperstellen notieren. Nicht durch die E2E-Suite oder eine heuristische UX-Review ersetzbar, weil es um echte Erstnutzer-Reaktionen geht.

## Block H: Release-Smoke v1.0.0

- [ ] `scripts/ops-smoke.sh release VERSION=1.0.0` auf frischem Server (nicht der Build-Host) mit den gepinnten 1.0.0-Images: prüft `prod-init`, ersten Login, einen Verkauf, einen Beleg, einen Export automatisiert (Gate 6). Hier nur das Ergebnis abnehmen.

## Abnahme-Entscheidungen

- [ ] Alle Suiten-Läufe (E2E, DSFinV-K-Validator, Berechtigungs-Matrix, Schwachstellen-Scans, Fuzz-Korpus, Parallelzugriffstest, TSE-Live-Suite, Ops-Smoke) grün; Befund-Report der einmaligen QA-Durchführung durchgesehen und offene Befunde priorisiert.
- [ ] Blöcke A–G dieses Guides durchgespielt und abgenommen.
- [ ] Go/No-Go-Entscheidung für den v1.0.0-Tag getroffen.
