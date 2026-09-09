# Plan: Release v1.0.0

**Ziel:** jotti 1.0.0 veröffentlichen — offene Dependabot-PRs mergen, manuelle QA abnehmen,
Version heben, Tag setzen, danach die Nachlaufarbeiten erledigen.

**Betroffene Dateien:** `CHANGELOG.md`, `docs/leitfaden/self-hosting.md`,
`docs/leitfaden/aktualisieren.md`, `.github/workflows/ci.yml`, `database/migrations/README.md`,
`docs/plans/`.

**Reihenfolge = Abhängigkeit.** Die Abschnitte 5 und 6 sind kein Release-Blocker.

## 1. Offene Pull Requests

- [ ] PR #129 (frontend-npm-Gruppe: vitest 4 → 5 Major, `eslint-plugin-react-x` und
      `eslint-plugin-react-dom` 5.19.0) — braucht einen eigenen Testlauf, da vitest 5 Node 22+
      und Vite 6.4+ verlangt und Mocks per Default vor jedem Test zurücksetzt

## 2. Manuelle QA

Nur das Nicht-Automatisierbare; Ablauf und Akzeptanz je Block:
[guide-manuelle-qa-v1.0.0.md](guide-manuelle-qa-v1.0.0.md).

- [ ] Vorbereitung: fiskaly-TEST-Konto mit `.env.fiskaly-test`, 80-mm-Bondrucker, zwei Endgeräte,
      echter Windows-Rechner, frische VM
- [ ] Block A: QR-Code mit echter fiskaly-Payload auf echtem Drucker scannbar, Druckbild sichtprüfen
- [ ] Block B: `make release-windows VERSION=…`, ZIP auf Windows entpacken, bis zum ersten Login
      smoke-testen
- [ ] Block C: Setup-Wizard real durchlaufen (TSS + Client aus jotti), TEST→LIVE-Umschaltung,
      PUK/PIN-Verwahrung dokumentiert, `make test-tse-live` gegenlesen; `make test-tse-live-setup`
      nur bewusst auslösen (die TSS ist unlöschbar)
- [ ] Block D: DSFinV-K-Export mit IDEA oder dem fiskaly-Prüftool prüfen, alternativ gegen die
      2.4-Beispiele gegenlesen
- [ ] Block E: `make prod-restore` destruktiv auf frischer VM, TLS/Let's Encrypt live auf dem
      echten Domain-Namen
- [ ] Block F: zwei echte Handys am selben Tisch parallel
- [ ] Block G: eine Servicekraft und ein Admin aus einem echten Verein ohne Anleitung,
      Stolperstellen notieren
- [ ] Abnahme: alle Suiten grün (`make verify`, `make test-e2e`, `make test-tse-live`,
      `scripts/ops-smoke.sh install|ops`), Blöcke A–G abgenommen, Go/No-Go entschieden

## 3. Release schneiden (Block H, nur nach Go)

- [ ] `JOTTI_VERSION=v0.14.0` in `docs/leitfaden/self-hosting.md` auf `v1.0.0` heben
- [ ] `docs/leitfaden/aktualisieren.md` prüfen: die Aussage zur Print-Relay-Version 0.17.3 gegen den
      1.0.0-Stand halten
- [ ] Release-Datum im Abschnitt `[1.0.0]` der `CHANGELOG.md` eintragen
- [ ] CI auf dem Release-Commit in `main` grün, inklusive der Jobs `e2e` und `upgrade-path`
- [ ] Version-Bump auf 1.0.0 (Image-Tags, `JOTTI_VERSION`, `VERSION` des Windows-Builds, ldflags)
      und Tag `v1.0.0` pushen; `release.yml` baut Images, ZIP und GitHub-Release
- [ ] Release-Text aus [release-notes-v1.0.0.md](release-notes-v1.0.0.md) vor die git-cliff-Notes
      stellen; die Laufzeit-Versionen (Go, Node, pnpm) müssen darin stehen
- [ ] Block I: `bash scripts/ops-smoke.sh release v1.0.0` auf frischem Server gegen die
      1.0.0-Images

## 4. Nach dem Tag (eigener Commit)

- [ ] `PREVIOUS_VERSION` in `.github/workflows/ci.yml` und `database/migrations/README.md` von
      `v0.17.1` auf `v1.0.0` heben; Job `upgrade-path` grün
- [ ] Löschen: `guide-manuelle-qa-v1.0.0.md`, `release-notes-v1.0.0.md`, `plan-praxis-feedback.md`,
      `plan-orchestrierung.md` und diesen Plan, sobald alle Boxen abgehakt sind

## 5. Produktentscheidungen

Offener Rest aus [findings-jotti-audit.md](findings-jotti-audit.md); jeder Punkt braucht eine
Entscheidung des Betreibers, keiner blockiert das Release.

- [ ] Kassenabschluss-Wiederanlauf nach einer Geldbewegung: Reparaturweg bauen oder Admin-Ablauf
      dokumentieren
- [ ] `GetEigeneUebersicht` im Barrierestatus: auf `GetAktiveKassensitzung` umstellen oder die
      Nullen hinnehmen
- [ ] Zeichen gegen Bytes an Kommentar- und Tischnamen-Feldern: Frontend zählt Bytes, oder die UI
      nennt die Einheit
- [ ] Login-Formular: eigenes Login-Schema (trim + min 1) statt `PasswordSchema`
- [ ] Windows: Restore-Weg für `manuell-*.sql` dokumentieren oder das Skript erweitern
- [ ] `JOTTI_DOMAIN` bei jedem Compose-Befehl des Public-Stacks: Option A (hinnehmen) bestätigen

## 6. Folgekandidaten

Ohne Release-Bezug, optional.

- [ ] `reverse-proxy/caddyfile.go`: Historien-Prosa umschreiben, den Ausschluss in
      `scripts/check-prose.sh` streichen
- [ ] `export.go`: `//nolint:forbidigo` auflösen und `GetOffeneKassensitzung` durch
      `GetAktiveKassensitzung` ersetzen
- [ ] `scripts/check-pins.sh` auf die Postgres-Pins in `scripts/test-integration.sh`,
      `scripts/test-tse-live.sh` und den CI-Services ausweiten
- [ ] `scripts/check-ui-labels.sh`: Treffer in Bezeichnern und Kommentaren ausschließen
- [ ] `dsfinvkpruefung`: `MaxLength` prüfen (heute nur Mapper-Tests)
- [ ] `kassenjournal_repo/mock.go`: Differenzbuchungen ungleich 0 abbildbar machen
- [ ] `tisch_sessions.unbezahlte_positionen`: die PascalCase-Projektion liegt außerhalb des
      Event-Vertragsgates — Guard ergänzen oder als ADR festhalten
- [ ] `api/auth/http/command_handler.go`: `PasswordSchema.Required()` mutiert das geteilte Schema —
      eine Kopie verwenden
- [ ] Status-Seite: eine `install.json` mit regelwidriger Subdomain führt zu einem wirkungslosen
      Neustart-Hinweis (theoretisch)
