# Offene Punkte

Sammelstelle für offene Folgearbeiten aus abgeschlossenen und nach dem Merge
gelöschten Plänen und Audits (Docs-Cleanup `30c32de`; die Quell-Dokumente
liegen in der Git-Historie unter `30c32de^`). Erledigte Punkte werden
abgehakt; ist alles erledigt, wird die Datei gelöscht.

## Audit-Folgen v0.15.0

Quelle: `plan-audit-v0.15.0.md`, Phase-5-Befunde vom 09.07.2026.

- [ ] **G3: Prettier-Hook auf die Repo-Wurzel begrenzen** (`.claude/settings.json`;
  nur der Maintainer kann die Hook-Config ändern, kein Agent). Die
  `node_modules`-Suche des PostToolUse-Hooks läuft per `dirname` bis `/` hoch;
  ein außerhalb des Repos liegendes `node_modules/.bin/prettier` würde
  ausgeführt (aktuell No-op, weil dort keines existiert). Patch: in der
  while-Schleife vor dem Binary-Test einfügen:
  `if [ -n "$CLAUDE_PROJECT_DIR" ]; then case "$d" in "$CLAUDE_PROJECT_DIR"|"$CLAUDE_PROJECT_DIR"/*) ;; *) break;; esac; fi;`
- [x] **G10: Dark-Mode-Kontrast der Löschen-Buttons** (Design-Entscheidung
  offen). Solide Buttons mit `bg-destructive text-white` erreichen im
  Dark-Mode nur 2,89:1 (WCAG AA verlangt 4,5:1; `--destructive` ist dort
  red-400, `frontend/src/index.css`). Betroffen: DruckstationConfigPage,
  ProductItem, VariantItem, UserItem, TischItem. Optionen: dunkleres
  Dark-`--destructive` (einfach, senkt aber den `text-destructive`-Kontrast
  auf Flächen) oder eigenes `--destructive-foreground`-Token (präziser, fünf
  Call-Sites). Alle übrigen Kern-Token-Paare bestehen AA in Light und Dark;
  a11y-relevant wegen BYOD-Nutzung im Freien.
  **Erledigt (2026-07-12, Admin-Redesign Phase 12):** Zweifach adressiert —
  (1) das Dark-`--destructive` wurde bereits in `ea3ade0` angehoben; (2) das
  Admin-Redesign hat die Befund-Grundlage aufgelöst: `VariantItem` und
  `UserItem` sind gelöscht, und jede destruktive Löschen-Aktion sitzt jetzt in
  einem „···"-DropdownMenu + AlertDialog. Die einzigen verbleibenden soliden
  `bg-destructive`-Buttons sind AlertDialog-Bestätigungen (ein Button pro
  Modal), keine über die Listenzeilen gestreuten Volltonflächen mehr
  (Compliance-Audit `acbafa6..HEAD`, alle Call-Sites geprüft).
- [ ] **G7: `reset-and-seed.sh --yes` umgeht den Bestands-Guard**
  (Produktentscheidung offen). `jotti seed` (CLI) und der E2E-Reset-Endpoint
  brechen bei vorhandenen Kassenjournal-Events ab; `scripts/reset-and-seed.sh
  --yes` (`make local-reset-and-seed`) löscht dagegen das Postgres-Volume
  komplett und umgeht damit jeden Guard — auch auf dem echten `local`-Stack
  mit aufbewahrungspflichtigen Daten. Empfehlung: Bestands-Check
  (Kassenjournal-Count) vor `docker volume rm`, auch bei `--yes`.
- [ ] **zog-Restmuster `GTE(0).Required()` auf Event-Summenfeldern prüfen.**
  zog behandelt den Go-Zero-Value `0` als „fehlend"; ein legitimer 0-Betrag
  würde mit HTTP 500 abgelehnt (Muster und Fix: Commit `531a17f`,
  Kassenführungs-Fluss). Elf Summenfelder in `backend/domain/kasse/` tragen
  das Muster noch: `bestellung.go:136`, `zahlung.go:28`, `stornierung.go:34`,
  `umbuchung.go:41`, `direktverkauf_events.go:27,46`,
  `tisch_session_events.go:34,48,69,87,110`. Pro Feld klären, ob 0 über die
  Domänenpfade erreichbar ist; wo ja, `Required()` streichen (`GTE(0)` fängt
  Negative weiterhin).

## Release-Vorbereitung v1.0.0

Quelle: `plan-v1.0.0-release.md` (Gates) und `plan-v1.0.0-nacharbeit.md`
Block 6. Die manuelle QA-Handarbeit steht im
[Rest-Guide](guide-manuelle-qa-v1.0.0.md); hier der Rest der
Release-Vorbereitung.

- [ ] **`CHANGELOG.md` anlegen** (git-cliff; `cliff.toml` existiert,
  `release.yml` nutzt es bisher nur für flüchtige Release-Notes) mit
  Einträgen ab v0.14.0; ab 1.0.0 gepflegt.
- [ ] **Beispiel- und Pin-Versionen anheben** (`docs/leitfaden/self-hosting.md`,
  `.env.example`, `docker-compose.release.yml`, Verfahrensdokumentation);
  `frontend/package.json` als tot dokumentieren und aus dem Bump-Set
  streichen (C13).
- [ ] **TODO/FIXME/XXX-Grep** in fiskalisch relevanten Modulen (Befunde
  beheben oder bewusst dokumentieren); voller `make verify` plus
  `make lint-backend-full` auf dem Release-Commit.
- [ ] **Release-Mechanik:** Version-Bump auf 1.0.0 (Image-Tags/`JOTTI_VERSION`,
  `VERSION` für den Windows-Build, ldflags-Version), Tag `v1.0.0` pushen —
  `release.yml` baut Images, Windows-ZIP und Release-Notes. Danach
  `scripts/ops-smoke.sh release VERSION=1.0.0` auf frischem Host
  (Rest-Guide Block H) und Abnahme nach dem Rest-Guide.
- [ ] **Roadmap-Lücke klären:** Der gelöschte Release-Guide hielt fest, dass
  kein offenes Roadmap-Feature v1.0-Blocker ist und K-23 (manuelle
  Tischfreigabe), R-05 (Produktumsatz-Reporting), F-08
  (GoBD-Integritätsnachweis) und F-09 (eBeleg) als v1.1 nachziehen sollten.
  Commit `186aea9` hat K-23, R-02 (CSV-Export), F-08 und F-09 jedoch ohne
  Begründung aus `anforderungen.md` entfernt (weder Roadmap noch
  Nicht-Ziele). R-05 ist inzwischen umgesetzt und in den Funktionsumfang
  verschoben; die Roadmap ist damit leer. Entscheiden, ob K-23, R-02, F-08
  und F-09 zurück in die Roadmap kommen oder als Nicht-Ziele dokumentiert
  werden.
