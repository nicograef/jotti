# Plan: Release v1.0.0

**Ziel:** jotti 1.0.0 veröffentlichen — manuelle QA abnehmen, Version heben, Tag setzen, danach
die Nachlaufarbeiten erledigen.

**Betroffene Dateien:** `CHANGELOG.md`, `docs/leitfaden/self-hosting.md`,
`docs/leitfaden/aktualisieren.md`, `.github/workflows/ci.yml`, `database/migrations/README.md`,
`docs/plans/`.

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
      nicht im fiskaly-Dashboard.
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
      Rate-Limiting deckt `scripts/ops-smoke.sh install|ops` ab.

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

- [ ] Beispielversion in `docs/leitfaden/self-hosting.md` (`JOTTI_VERSION=v0.14.0`) auf `v1.0.0`
      heben. Nur diese Datei ist betroffen: `docker-compose.release.yml` ist ein
      `:RELEASE_VERSION`-Template (der Release-Workflow ersetzt den Platzhalter), `.env.example`
      hält `JOTTI_VERSION=` bewusst leer, die Verfahrensdokumentation trägt an dieser Stelle einen
      Betreiber-Platzhalter (`«z. B. v0.2.0»`) und `frontend/package.json` steht auf `0.0.0`, das
      nirgends im Build gelesen wird.
- [ ] `docs/leitfaden/aktualisieren.md` prüfen: die Aussage zur Print-Relay-Version 0.17.3 gegen den
      1.0.0-Stand halten
- [ ] Release-Datum im Abschnitt `[1.0.0]` der `CHANGELOG.md` eintragen
- [ ] CI auf dem Release-Commit in `main` grün: die Jobs `backend-ci`, `backend-golangci`,
      `repo-checks`, `frontend-ci`, `backend-integration-tests`, `e2e` und `upgrade-path` decken
      `make verify` und `make lint-backend-full` ab.
- [ ] Version-Bump auf 1.0.0 (Image-Tags/`JOTTI_VERSION`, `VERSION` für den Windows-Build,
      ldflags-Version) und Tag `v1.0.0` pushen — `release.yml` baut Images, Windows-ZIP und
      Release-Notes (git-cliff) und veröffentlicht das GitHub-Release.
- [ ] Release-Text (Abschnitt „Release-Text“ unten) vor die git-cliff-Notes stellen; die
      Laufzeit-Versionen (Go, Node, pnpm) müssen darin stehen
- [ ] Block I: `bash scripts/ops-smoke.sh release v1.0.0` auf frischem Server (nicht der
      Build-Host) mit den gepinnten 1.0.0-Images: prüft `prod-init`, ersten Login, einen Verkauf,
      einen Beleg, einen Export automatisiert. Hier nur das Ergebnis abnehmen.

## 3. Nach dem Tag (eigener Commit)

- [ ] `PREVIOUS_VERSION` in `.github/workflows/ci.yml` und `database/migrations/README.md` von
      `v0.17.1` auf `v1.0.0` heben; Job `upgrade-path` grün
- [ ] Diesen Plan löschen, sobald alle Boxen abgehakt sind

## 4. Produktentscheidungen

Jeder Punkt braucht eine Entscheidung des Betreibers, keiner blockiert das Release.

- [ ] Kassenabschluss-Wiederanlauf nach einer Geldbewegung: Steht ein Kassensturz und bewegt sich
      danach Geld (Tischzahlung, Direktverkauf, Storno mit Rückgabe, Geldtransit), bricht jeder
      Wiederanlauf mit `ErrBuchungenNachKassensturz` ab; kein Befehl hebt einen Kassensturz auf,
      die Kassensitzung bleibt offen, das Frontend sagt „Bitte den Administrator kontaktieren".
      Reparaturweg (z. B. zweiter Kassensturz beim Wiederanlauf) bauen oder Admin-Ablauf
      dokumentieren.
- [ ] `GetEigeneUebersicht` im Barrierestatus: liest weiter `GetOffeneKassensitzungNr`; die eigene
      Übersicht der Servicekraft zeigt Nullen, solange die Kassensitzung `wird_abgeschlossen` ist.
      Auf `GetAktiveKassensitzung` umstellen oder die Nullen hinnehmen.
- [ ] Zeichen gegen Bytes an Kommentar- und Tischnamen-Feldern: Zod zählt Code-Units, die
      eingefrorenen Event-Schemas Bytes — ein Umlaut-Kommentar mit 100 Zeichen passiert das
      Frontend und bekommt 400. Betreiber-Felder zählen Zeichen. Frontend zählt Bytes
      (`TextEncoder`) oder die UI nennt die Einheit.
- [ ] Login-Formular: `frontend/src/lib/AuthBackend.ts` nutzt für den Login das volle
      `PasswordSchema` (min 6/max 72) und verrät damit die Passwort-Policy; der Endpunkt tut es
      nicht mehr. Eigenes Login-Schema (trim + min 1).
- [ ] Windows: `jotti-restore.cmd` liest nur `jotti-*.sql` aus dem Volume; für `manuell-*.sql` auf
      dem Host gibt es keinen dokumentierten Weg. Restore-Weg dokumentieren oder das Skript
      erweitern.
- [ ] `JOTTI_DOMAIN` bei jedem Compose-Befehl des Public-Stacks: Folge der Entscheidung Option A —
      jeder Compose-Befehl mit `docker-compose.prod.yml` (auch `make prod-down`, `make prod-logs`)
      verlangt die Variable. Option A (hinnehmen) laut statt still bestätigen.
- [ ] Laufende Bewirtung im Vereinsheim: Ein Verein setzt jotti auch dafür ein. Das ist nicht die
      Zielgruppe (2–3 Feste pro Jahr), wird aber nicht verhindert. Ob das ein Nicht-Ziel wird, ist
      offen.
- [ ] Kontingent-Funktion für den Bondruck: Ein Verein mit Lehrgängen und einem Turnier wünscht sie
      an der Kasse (Teilnehmer erhalten ein festes Kontingent). Anforderung noch unklar, Rückfrage
      läuft. Warenwirtschaft ist Nicht-Ziel; ein Tisch pro Teilnehmer oder Abholbons könnten
      reichen.

## 5. Folgekandidaten

Ohne Release-Bezug, optional.

- [ ] `reverse-proxy/caddyfile.go`: die Zeilen 14, 51, 69 und 92 tragen Historien-Prosa;
      umschreiben, dann den Ausschluss in `scripts/check-prose.sh` streichen
- [ ] `export.go` liest die Kassensitzung mit `GetOffeneKassensitzung` und `//nolint:forbidigo`;
      wegen der Sortierung von `GetAllKassensitzungen` liefert `GetAktiveKassensitzung` dieselbe
      Sitzung — die Ausnahme kauft kein Verhalten. Auflösen und ersetzen.
- [ ] `scripts/check-pins.sh` prüft Compose, Dockerfiles und `packageManager`; die Postgres-Pins in
      `scripts/test-integration.sh`, `scripts/test-tse-live.sh` und den CI-Services liegen
      außerhalb. Prüfung darauf ausweiten.
- [ ] `scripts/check-ui-labels.sh` prüft das Vorkommen zitierter Bedienelemente in `frontend/src`;
      ein Zitat, das nur als Bezeichner oder Kommentar vorkommt, läuft durch. Treffer in
      Bezeichnern und Kommentaren ausschließen — schärfer ginge es nur gegen gerenderte Texte.
- [ ] `dsfinvkpruefung` prüft Typen und Dezimalformat, nicht `MaxLength`; die Feldlängen sichern
      allein die Mapper-Tests. `MaxLength` ergänzen.
- [ ] `backend/repository/kassenjournal_repo/mock.go` kann keine Summe der Differenzbuchungen ≠ 0
      abbilden; den unterscheidenden Fall deckt nur der Integrationstest. Abbildbar machen.
- [ ] `tisch_sessions.unbezahlte_positionen` persistiert `[]kasse.Position` (ohne JSON-Tags,
      PascalCase-Schlüssel) als Projektion; außerhalb des Event-Vertragsgates, ein Umbenennen der
      Felder wäre ein Bruch persistierter Daten. Guard ergänzen oder in `docs/decisions.md`
      festhalten.
- [ ] `api/auth/http/command_handler.go` ruft `user.PasswordSchema.Required()` auf dem geteilten
      Paket-Schema auf (zog mutiert in place) — latente Falle, heute ohne Auswirkung. Eine Kopie
      verwenden.
- [ ] Status-Seite: eine alte `install.json` mit einer Subdomain, die den Subdomain-Regex verletzt,
      lässt den Proxy nur mit der internen CA weiterlaufen; die Status-Seite rät dann zu einem
      Neustart, der nichts ändert (acme-dns vergibt Kleinbuchstaben-UUIDs, daher theoretisch).

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
| pnpm       | 11.6.0      | `frontend/package.json` (`packageManager`)                      |
| PostgreSQL | 17.8        | `docker-compose.prod.yml` (`postgres:17.8`)                     |
| Caddy      | 2.11.4      | `reverse-proxy/Dockerfile` (`caddy:2.11.4`)                     |

### Aktualisieren

- **Windows-Rechner (Standardweg):** `jotti-stop.cmd`, neues Release-ZIP entpacken,
  `jotti-start.exe` starten. Der Starter sichert die Datenbank automatisch, bevor er
  aktualisiert. Ablauf, Rauchtest und der Weg zurück: `docs/leitfaden/aktualisieren.md`.
- **Eigener Server (Self-Hosting):** `JOTTI_VERSION=v1.0.0` in `.env` setzen und `make prod-update`
  ausführen. Das Skript zieht vor der Aktualisierung ein Backup. Ablauf und Backup-Strategie:
  `docs/leitfaden/aktualisieren-backups.md`.
