# Plan: Audit-Fixes jotti

> Source PRD: docs/plans/findings-jotti-audit.md (Stand 2026-09-08)

## Goal

Die bestätigten Befunde des Vollaudits beheben und jede Defektklasse mit einem Gate
absichern, das den Rückfall verhindert. Reihenfolge: Release-Blocker, dann Gates, dann
Major, dann Minor, zuletzt eine Entscheidungsphase für große Refactorings. Danach ist der
Baum für v1.0.0 bereit.

## Architectural decisions

Durchgängig für alle Phasen:

- **Freeze-Disziplin bleibt unangetastet.** Kein Eingriff in `01_initial.up.sql`, keine
  In-place-Änderung an Event-JSON, keine Handarbeit in `backend/sqlc/dbgen/`.
- **Schema-Änderungen** entstehen nur als neue additive Migration `NN_<name>.up.sql`; die
  Nummer wird beim Landen vergeben. Dieser Plan braucht keine.
- **Lesepfad-Schemas sind eingefroren.** Neue Grenzen kommen in ein Eingabe-Schema, nie in
  ein Event-Data-Schema: `event.ParseData` validiert beim Lesen jedes alte Event.
- **Cleanup ändert kein Verhalten.** Wo ein Fix das Verhalten ändert, sagt das Kriterium es
  ausdrücklich und nennt den Test, der die neue Zusage prüft.
- **Gates sind ausführbare Befehle**, keine Review-Regeln: Go-Tests, `golangci-lint`,
  Prettier, oder ein Shell-Skript unter `scripts/`. Jedes neue Gate landet grün.
- **Neue Gate-Skripte** liegen unter `scripts/` mit englischem Namen. `make check-repo`
  bündelt per Glob **alle** `scripts/check-*.sh`; der CI-Job `repo-checks` ruft es.
- **Gate-Skripte lesen nur `git ls-files`.** Sonst scannen sie `node_modules` und
  `frontend/dist` und werden unbrauchbar langsam und falsch rot.
- **Keine neuen Abhängigkeiten.** Prettier kommt aus `frontend/node_modules`, die
  AST-Gates aus der Go-Standardbibliothek (`go/ast`, `go/parser`), ESLint-Regeln aus dem
  Kernregelwerk.
- **Fehler-Codes** sind der Vertrag mit dem Frontend: Anwendungs-Sentinel →
  `helper.MapError`-Karte → `commonErrorMessages`. Alle drei Enden ändern sich gemeinsam.
- **Deutsche Ubiquitous Language** gilt für Domäne und Benutzer-Strings, Englisch für
  Infrastruktur. Windows-Konsolenausgaben bleiben ASCII; Quellkommentare nicht.

## Inventory

| Datei / Symbol                                                                             | Rolle im Plan                                                   |
| ------------------------------------------------------------------------------------------ | --------------------------------------------------------------- |
| `Makefile — check, check-frontend, lint-backend`                                           | Merge-Gate; nimmt `check-repo` und `check-format` auf           |
| `.github/workflows/ci.yml — changes, frontend-ci, shellcheck-ci`                           | CI-Gates; neuer Job `repo-checks`, neuer `docs`-Filter          |
| `backend/.golangci.yml`                                                                    | `run.build-tags` kennt nur `integration`; nimmt `forbidigo` auf |
| `scripts/setup-dev-tools.sh — GOLANGCI_LINT_VERSION`                                       | Werkzeug-Pins; Toolchain-Erkennung fehlt                        |
| `frontend/package.json — scripts.lint`                                                     | `eslint --fix` verhindert ein rotes Gate                        |
| `frontend/.prettierrc`, `website/.prettierrc`                                              | inhaltsgleich; werden zur Root-Konfiguration                    |
| `backend/api/helper/http.go — MapError, SendConflictDetails`                               | zentrale Fehler-Abbildung                                       |
| `backend/domain/kasse/bestellung.go — positionSchema`                                      | Event-Data-Schema, Lese- und Schreibpfad; bleibt unverändert    |
| `backend/domain/event/event.go — ParseData`                                                | validiert jedes gelesene Event gegen sein Schema                |
| `backend/domain/kasse/event_json_contract_test.go`                                         | eingefrorener Event-Vertrag; Vorbild für das Decode-Gate        |
| `backend/api/kasse/enrichment/enrichment.go — EnrichPositionen`                            | einzige Anreicherung für Fat Events                             |
| `backend/api/druck/bondruck/application/escpos/formatter.go`                               | alle gedruckten Zeitstempel                                     |
| `backend/repository/kassensitzungen_repo — GetOffeneKassensitzung, GetAktiveKassensitzung` | Statusgrenze `wird_abgeschlossen`                               |
| `backend/sqlc/queries/kassensitzungen.sql — GetKassenbestand`                              | zieht die gebuchte Differenz ab; Grundlage des Wiederanlaufs    |
| `frontend/src/lib/errorMessages.ts — commonErrorMessages`                                  | zentrale Fehlermeldungen                                        |
| `frontend/src/admin/AdminSidebar.tsx`                                                      | Quelle der zitierten Menüpunkte für das UI-Label-Gate           |
| `docs/leitfaden/**`                                                                        | Anwenderdoku; wird aus `docs/` in die Website gelesen           |

## Resolved decisions

- **Kollision Bereichsbefund gegen „Verworfene Befunde".** Der bestätigte Bereichsbefund
  gilt; Präzedenz ist die Lead-Entscheidung zu `packaging/windows/KURZANLEITUNG.md`.
  Betrifft unter anderem `hooks.ts`, `Makefile:78-87`, `docs/compliance.md:296`.
- **Rein verworfene Befunde bleiben unberührt** — mit einer benannten Ausnahme.
  `DirektverkaufAbschluss.tsx` rotiert seinen `verkaufId` weiterhin nur bei leer→gefüllt.
- **Ausnahme: `frontend/src/components/common/FormFields.tsx:14-19`.** Der Einzelbefund ist
  verworfen, die Defektklasse „Geteilte Schichten importieren aufwärts" ist bestätigt.
  Die Klasse gilt; die Datei ist die einzige heutige Verletzung und wird aufgelöst.
- **Mengen-Obergrenze 999 nur auf der Eingabeseite.** `positionSchema` ist Event-Data und
  wird beim Lesen validiert; eine neue Grenze dort macht alte Events unlesbar.
- **`helper.MapError` bekommt eine geordnete Liste.** Heute matcht in keiner Karte mehr als
  ein Sentinel; die Änderung ist verhaltensneutral und hält die Regel konstruktiv fest.
- **Prettier-Geltungsbereich.** Eine Root-Konfiguration deckt `ts, tsx, js, mjs, cjs, json,
css, md` im ganzen Repo ab. Grund: 29 Markdown-Dateien außerhalb `docs/plans/` und drei
  `e2e`/`website`-Dateien sind heute nicht formatiert, ohne dass ein Gate greift.
- **`docs/plans/**` bleibt von Prettier ausgenommen.** `implement-plan` reserviert die
  Plandatei dem Lead; ein `--write`-Lauf eines Workers kollidiert mit dessen Tick-Commit.
- **`TERMS.md`, `CLA.md` und `CHANGELOG.md` werden mitformatiert.** Prettier ändert nur
  Whitespace; der Eigentümer liest den Diff dieser drei Dateien vor dem Merge gegen.
- **`*.astro` bleibt ausgenommen.** `prettier-plugin-astro` wäre eine neue Abhängigkeit
  (Regel 16); `astro check` deckt diese Dateien ab.
- **`frontend/src/components/ui` bleibt von ESLint und Prettier ausgenommen.** Es ist
  übernommener shadcn-Code; die wenigen eigenen Änderungen laufen über Review.
- **Kein ESLint-Gate für `isError`.** Eine Regel „jeder `useQuery`-Konsument reicht
  `isError` durch" gibt es im Kernregelwerk nicht; ein Plugin wäre eine neue Abhängigkeit.
- **Kein ESLint in `e2e/`.** Das Paket hat weder Konfiguration noch Abhängigkeit; das Gate
  entsteht als Grep-Skript im vorhandenen Muster.
- **`tisch_repo` wird umbenannt statt dokumentiert.** `docs/language.md` Regel 1/5 fordert
  deutsche Domänen-Nomen; die Umbenennung ist compilergeprüft, eine Ausnahme schwächt die
  Regel.
- **`docs/language.md` §Geplant wird gestrichen.** `docs/anforderungen.md` führt die
  Roadmap als leer; neue Anforderungen aufzunehmen wäre Feature-Creep.
- **`01_initial.up.sql` ist nicht das kanonische Schema.** Alle Verweise zeigen auf die
  Summe der `*.up.sql` in Reihenfolge.
- **Kein Versions-Bump.** Dependabot, TypeScript 7 und vitest 5 übernimmt eine andere
  Session; kein Fix dieses Plans braucht einen Bump.
- **CLA.md § 2 b) wird klargestellt**, nicht inhaltlich geändert: Abschnitt 1 und
  `LICENSE:112-113` gewähren die Rechte dem Autor. Der Eigentümer liest den Wortlaut vor
  dem Merge gegen.
- **Raspberry Pi wird nicht mehr zugesagt** (Lead-Entscheidung). `docs/produktbeschreibung.md`
  nennt den Pi; `.github/workflows/release.yml` baut ohne `--platform`, die Images sind
  amd64-only. Kriterium 12.10 beschränkt die Aussage auf x86-64. Grund: eine ungetestete
  Plattform ohne Praxisbeleg widerspricht dem Produkt-Konservatismus. Multi-Arch-Builds
  brächten längere Builds und eine zweite Plattform ohne Testabdeckung.
- **`make test-e2e` läuft in dieser Session nur in der CI auf dem Draft-PR #126.** Die
  Sandbox kann die Compose-Images nicht bauen. Der Phasen-Worker führt `make check` bzw.
  `make verify` lokal aus. Die Lead-Session prüft den CI-Job `e2e` nach dem Landen jeder
  Phase, die `make test-e2e` als Gate nennt. Die Gate-Zeilen bleiben unverändert.
- **Plan-Dateien sind transient.** Befunde zu `docs/plans/**` werden nicht übernommen.

### Kritik (2026-09-08)

Eingearbeitet aus der Opus-Dreifachkritik am Entwurf:

- **Mengen-Grenze trifft den Lesepfad**: 1.4 begrenzt jetzt nur Eingabe-Schemas.
- **Kassensturz-Wiederanlauf**: 7.4 vergleicht ohne die abschluss-eigene Differenzbuchung.
- **Prosa-Gate gegen eingefrorene Migrationen**: 3.1 nimmt `database/migrations/**` aus.
- **Prosa-Gate gegen Regeltexte**: 3.1 nimmt `AGENTS.md` und `.github/instructions/**` aus.
- **Prosa-Gate ohne Wortgrenzen**: 3.1 prüft ganze Wörter und führt eine Allowlist.
- **Fehlende Defektklasse**: „Kommentare behaupten Verhalten" hat Zeile und Kriterium 13.10.
- **`nilerr` prüft `log.Error()` nicht**: 5.9 bringt ein AST-Gate und die 20 Fundstellen.
- **`errorlint` ist bereits aktiv**: die Gate-Zeile nennt nur noch 9.4.
- **`GetOffeneKassensitzung`-Gate**: 7.1 zählt jede Aufrufstelle auf und verankert das Muster.
- **`check-build-tags.sh` war unverdrahtet**: 2.3 hängt es in `make check`, 3.2 per Glob.
- **CI-Job ohne Bedingung**: 3.2 ergänzt den `docs`-Filter und die Auslösebedingung.
- **Doppelter Zugriff auf `jotti-restore.cmd`**: 1.8 ist Doku, 11.1 besitzt die `.cmd`.
- **ASCII-Gate zu breit**: 4.1 prüft nur Go-String-Literale und `.cmd`, keine Kommentare.
- **`schema_grenzen_test.go`**: 6.1 bekommt Ausnahmeliste und ein existierendes Paket.
- **Sentinel-Scan zu eng**: 5.1 liest alle `backend/api/**`, nicht nur `errors.go`.
- **`import/no-restricted-paths`**: 10.9 nutzt die Kernregel `no-restricted-imports`.
- **ESLint in `e2e/`**: 13.5 ersetzt es durch `scripts/check-e2e-assertions.sh`.
- **`PasswordSchema` am Login**: 8.1 normalisiert nur, statt die Policy zu verraten.
- **`JOTTI_DOMAIN` erzwingen**: als Regel-16-Frage dem Eigentümer vorgelegt, mit A entschieden.
- **Betreiber-Grenzen nur beim Schreiben**: 6.5 kürzt zusätzlich im DSFinV-K-Mapper.
- **`MapError`-Umbau war Scope**: 5.7 trägt ihn eigenständig, mit Begründung.
- **Zeitzone nur im Druckpfad**: 8.8 und 8.9 ergänzen ELSTER-Datum und Exportnamen.
- **Fehlende Zugehörigkeitsprüfung beim Löschen**: 5.8 deckt `DeleteVariante()` ab.
- **Caddyfile mit 0o644**: 11.10 schreibt 0600 und prüft den Modus.
- **Kein Pin-Gate**: 11.8 bringt `scripts/check-pins.sh` samt `packageManager`-Abgleich.
- **Kein Enum-Gate**: 13.2 bringt `scripts/check-domain-enums.sh`.
- **UI-Label-Gate zu klein gedacht**: 12.1 und 12.2 decken alle 26 offenen Zitate ab.
- **Depends-on-Graph**: jede Phase hängt an jeder Phase, die eine ihrer Dateien schreibt.
- **Fehlende Choke-Zeilen**: Phase 5, 6, 8, 9, 13 führen sie nun.
- **Prettier gegen die Plandatei**: 2.5 nimmt `docs/plans/**` aus.
- **Unbegrenzte Sweeps auf Sonnet**: 2.8 und 2.9 laufen auf Opus 5 mit Abbruchgrenze.
- **Sammelkriterien**: 3.4, 5.5, 5.6, 10.4, 13.1, 13.4, 13.6 sind aufgeteilt.
- **Zeilennummern im Context**: die driftenden Verweise nennen jetzt Pfad plus Symbol.
- **Abdeckungszahlen**: die Summe nennt die geprüften Zählungen und die Abweichung.
- **Restmengen-Zusage war falsch**: eine sechste Klasse führt die offenen Korrektheits-Minors.
- **`caddyfile.go`-Kollision**: Phase 3 lässt die Datei, Phase 11 schreibt sie.
- **Falsche Zahl „28 Markdown-Dateien"**: gemessen sind es 29 außerhalb `docs/plans/`.
- **Herkunft der TERMS-Frage**: als Drift ohne Audit-Befund ausgewiesen, vom Eigentümer entschieden.
- **TERMS-Frage ohne Rückfall**: 12.14 setzt Option A um.
- **`AlleKategorien()` existiert nicht**: 13.2 legt die Funktion an.
- **Tautologische `e2e`-Bedingung**: 2.6 sagt, welche Jobs ohne Bedingung laufen.
- **`isError`-Gate herabgestuft**: die Begründung steht in den Resolved decisions.
- **DRY-Gate-Zeile überzogen**: sie nennt nur noch Shell und E2E-Selektoren.
- **CSP-Gate-Zeile überzogen**: die übrigen Kopien stehen unter „Nicht übernommene Befunde".
- **Satzlängen**: die überarbeiteten Kriterien halten die 20-Wörter-Grenze.

Nachprüfung (Opus, gleicher Tag):

- **Depends-on-Graph unvollständig**: `mapper.go`, `Makefile`, `frontend/package.json` und
  die 4.4-Kommentare in 59 Backend-Dateien serialisieren jetzt die Phasen 5 bis 13.
- **`check-pins.sh` ohne Allowlist**: 11.8 nennt Quelle, Vergleichsregel und Allowlist.
- **Unterlisten in 1.4, 7.1, 7.4, 13.5**: als echte Listen mit Leerzeile davor.
- **7.2 Name gegen Verhalten**: Methode, Handler, Tests, Route und Frontend-Aufruf heißen
  `…AktiveKassensitzung…`.
- **Raspberry-Pi-Frage**: vom Lead mit Option A entschieden, jetzt Resolved decision.
- **`make test-e2e` in der Sandbox**: Resolved decision zur CI-Prüfung auf PR #126.

## Open questions / Risks

Keine offenen Fragen. Der Eigentümer hat am 2026-09-08 entschieden:

- **TERMS.md § 8 Abs. 2 (Option A)**: die Zusage lautet „unterstützt die Anforderungen der
  deutschen Kassensicherungsverordnung (KassenSichV) technisch". Kriterium 12.14 setzt das
  um. Das Versionsdatum von `TERMS.md` bleibt unverändert; der Eigentümer prüft es beim
  Merge zusammen mit dem Diff.
- **`JOTTI_DOMAIN` (Option A)**: `docker-compose.prod.yml` erzwingt die Variable mit
  `${JOTTI_DOMAIN:?…}`, zusätzlich zum `loadConfig`-Guard. Die Regel-16-Rückfrage ist
  damit beantwortet. Kriterium 11.5 setzt beides um.

### Risiken

- **Phase 3 und Phase 12 fassen beide Doku an.** Phase 12 hängt an Phase 3, sonst
  entstehen Konflikte in `docs/handbuch.md`.
- **Phase 3 und Phase 11 fassen beide `reverse-proxy/caddyfile.go` an.** Phase 3 lässt die
  Datei aus; Phase 11 schreibt den Kommentar in Kriterium 11.4.
- **Der Prettier-Sweep in Phase 2 berührt 29 Doku-Dateien.** Er muss vor Phase 3 und 12
  landen, sonst wird jeder Doku-Commit doppelt formatiert.
- **Phase 1 Kriterium 2 verschiebt gedruckte Uhrzeiten um bis zu zwei Stunden.** Das ist
  die beabsichtigte Korrektur; alte Belege bleiben unverändert.
- **Kriterium 4.4 schreibt Kommentare in 59 Backend-Dateien.** Darum hängen Phase 5 und
  Phase 11 an Phase 4; alle späteren Backend-Phasen folgen transitiv.

---

## Phase 1: Release-Blocker

**Depends on**: none

**Modell**: Opus 5

**Review-Tier**: Gate + Sweep + Opus-Review + Skeptiker je Kriterium

**Gate**: `make verify` und `make test-e2e`

**Choke-Dateien**: `packaging/windows/KURZANLEITUNG.md`, `frontend/src/admin/users/UserRow.tsx`

### Context

- `docs/leitfaden/tse-einrichten.md` — Schritt 2 nennt „Finanzamt", „TSE-Anbindung",
  „Einrichten oder ändern"; keiner dieser Texte existiert im Frontend
- `frontend/src/admin/finanzamt/EinrichtungSection.tsx — EinrichtungSection()` — der Link
  auf `/admin/tse-einrichtung` steht nur im `!tseOk`-Zweig
- `frontend/src/admin/AdminSidebar.tsx` — Menüpunkt heißt „Finanzamt & TSE"
- `backend/api/druck/bondruck/application/escpos/formatter.go` — acht `.Format(`-Aufrufe,
  kein `time.LoadLocation`
- `backend/api/kasse/enrichment/enrichment.go — EnrichPositionen()` — prüft Aktiv-Status,
  nie die Paarung Produkt/Variante
- `backend/repository/produkt_repo/batch.go — GetVariantenByIDs()` — selektiert kein
  `produkt_id`
- `backend/domain/kasse/bestellung.go — positionSchema` — Event-Data-Schema; `ParseData`
  validiert damit auch jedes gelesene Event
- `backend/api/kasse/tischgeschaeft/http/command_handler.go — bestellPositionInputSchema,
positionRefRequestSchema` — Eingabe-Schemas des Tischgeschäfts
- `backend/api/kasse/direktverkauf/http/command_handler.go — verkaufPositionInputSchema,
positionRefRequestSchema` — Eingabe-Schemas des Direktverkaufs
- `backend/domain/kasse/positionen.go — ValidatePositionRefs()` — deckelt jede Ref-Menge
  auf die vorhandene Positionsmenge
- `backend/api/stammdaten/user/http/command_handler.go — DeleteUserHandler()` — einziger
  Handler mit Selbst-Guard
- `frontend/src/service/components/table/BestellungAbschluss.tsx — BestellungAbschluss()`
- `frontend/src/admin/kasse/GeldtransitDialog.tsx — GeldtransitDialog()`
- `scripts/prod-backup.sh` — `mkdir -p "$BACKUP_DIR"` ohne `umask`
- `packaging/windows/KURZANLEITUNG.md — Abschnitt „jotti aktualisieren"` — meldet den
  Restore-Erfolg zusammen mit dem Stackstart

### What to build

Die acht Befunde, die einen Release verhindern: ein gesetzlich zwingender
Einrichtungsschritt ohne Einstieg, falsche Belegzeiten, ein umgehbarer Steuersatz, ein
Überlauf im unveränderlichen Journal, die dauerhafte Selbstaussperrung, zwei still
verworfene Buchungen, weltlesbare Backups und ein Recovery-Pfad, der nicht heilt.

`DirektverkaufAbschluss.tsx` bleibt unverändert — der gleichlautende Befund wurde im Audit
verworfen. Die `.cmd`-Skripte gehören Phase 11.

### Acceptance criteria

- [x] `docs/leitfaden/tse-einrichten.md` Schritt 2 beschreibt den realen Pfad
      („Finanzamt & TSE" → Schrittkarte „2 · TSE aktiv" → „TSE einrichten"), und
      `EinrichtungSection.tsx — EinrichtungSection()` bietet den Link auf
      `/admin/tse-einrichtung` auch im `tseOk`-Zweig an; `FinanzamtPage.test.tsx` prüft
      beide Zweige. Befund: docs/leitfaden/tse-einrichten.md:26-32
- [x] `escpos/formatter.go` lädt `Europe/Berlin` einmal als Paket-Variable und formatiert
      jeden Zeitpunkt als `zeitpunkt.In(berlin)`; ein Unit-Test mit
      `2026-07-01T23:30:00Z` erwartet auf Kassenbeleg und Arbeitsbon den
      02.07.2026, 01:30. Befund: backend/api/druck/bondruck/application/escpos/formatter.go:113-259,
      :113-305
- [x] `produkt_repo/batch.go — GetVariantenByIDs()` liefert `produkt_id` mit, und
      `enrichment.go — EnrichPositionen()` lehnt eine Variante, die nicht zum
      mitgesendeten Produkt gehört, vor dem Aktiv-Check mit `ErrProduktNotFound` ab;
      `produkt_repo/mock.go` und ein Integrationstest „fremde Variante wird abgelehnt"
      ziehen nach. Befund: backend/api/kasse/enrichment/enrichment.go:76-99, :72-99,
      backend/repository/produkt_repo/batch.go:16-63
- [x] Die Mengen-Obergrenze 999 gilt ausschließlich auf der Eingabeseite.
      Befund: backend/domain/kasse/bestellung.go:97-106

  - `domain/kasse` erhält `PositionEingabeSchema` mit `GTE(1).LTE(999)`.
  - `positionSchema` bleibt als Event-Data-Schema unverändert.
  - `bestellPositionInputSchema` (tischgeschaeft:70) und `verkaufPositionInputSchema`
    (direktverkauf:41) beziehen die Grenze daraus.
  - Die beiden `positionRefRequestSchema` (tischgeschaeft:75, direktverkauf:104)
    bleiben bei `GTE(1)`; `kasse.ValidatePositionRefs()` deckelt sie fachlich.
  - Zod spiegelt die Grenze in `service/table/Bestellung.ts:14,25` und
    `service/direktverkauf/Direktverkauf.ts:13,49`.
  - `service/schemas.ts — PositionRefSchema` bleibt unverändert.
  - Je ein Test lehnt 1000 auf beiden Eingabepfaden ab.
  - Ein Test liest ein persistiertes Event mit Menge 1000 und storniert es erfolgreich.

- [x] `DeactivateUserHandler()` und der Rollenwechsel in `UpdateUserHandler()` lehnen die
      eigene Benutzer-ID mit `cannot_deactivate_self` bzw. `cannot_demote_self` ab (wie
      `DeleteUserHandler()`), `UserRow.tsx — UserRow()` sperrt Status-Switch und Rollenfeld
      bei `isSelf`, und `commonErrorMessages` kennt beide Codes.
      Befund: backend/api/stammdaten/user/http/command_handler.go:166-219, :166-183,
      frontend/src/admin/users/UserRow.tsx:69-84
- [x] `BestellungAbschluss.tsx` erneuert `bestellungId`, sobald sich Positionen oder
      Kommentar gegenüber dem letzten Absendeversuch unterscheiden, und
      `GeldtransitDialog.tsx` erneuert `geldtransitId` im Öffnen-Effekt neben
      `form.reset`; je ein Vitest-Fall erzwingt den neuen Schlüssel.
      Befund: frontend/src/service/components/table/BestellungAbschluss.tsx:50-81,
      frontend/src/admin/kasse/GeldtransitDialog.tsx:50-70
- [x] `scripts/prod-backup.sh` setzt `umask 077` vor `mkdir -p "$BACKUP_DIR"`, erzwingt
      `chmod 700` auf dem Verzeichnis und `chmod 600` auf jedem Dump, und prüft den Modus
      der erzeugten Datei vor der Erfolgsmeldung. Befund: scripts/prod-backup.sh:66-94
- [x] `packaging/windows/KURZANLEITUNG.md` Abschnitt „jotti aktualisieren" meldet den
      Restore-Erfolg getrennt vom Stackstart und nennt das vorherige Release-ZIP als
      Rückweg. Die `.cmd`-Skripte selbst ändert Kriterium 11.1.
      Befund: packaging/windows/jotti-restore.cmd:31-37

---

## Phase 2: Gates, die rot werden können

**Depends on**: 1

**Modell**: Sonnet 5; die Sweep-Kriterien 2.8 und 2.9 laufen auf Opus 5

**Review-Tier**: Gate + Sweep + Opus-Review

**Gate**: `make check`, `make website-check`, `make check-format`,
`bash scripts/check-build-tags.sh`

**Hinweis**: Das Gate entsteht im ersten Kriterium und wird erst durch den Sweep dieser
Phase grün; gelandet wird nur der grüne Endstand.

**Choke-Dateien**: `Makefile`, `.github/workflows/ci.yml`, `frontend/package.json`,
`backend/.golangci.yml`

### Context

- `frontend/package.json — scripts.lint` — `eslint --fix --max-warnings=0 .`
- `Makefile — lint-backend` — `goimports -l .` endet immer mit 0; `check-backend` hat den
  Guard bereits
- `backend/.golangci.yml — run.build-tags` — nur `integration`, also wird keine
  `//go:build unit`-Datei gelintet
- `frontend/.prettierrc`, `frontend/.prettierignore`, `website/.prettierrc` — identische
  Konfiguration, Geltungsbereich endet am Paketordner
- `scripts/setup-dev-tools.sh — GOLANGCI_LINT_VERSION` — überspringt den Neubau, sobald die
  Version stimmt, auch nach einem Go-Toolchain-Wechsel
- `.github/workflows/ci.yml — changes, shellcheck-ci, e2e` — `shellcheck` ohne `-x`,
  tautologische `if`-Bedingung am `e2e`-Job

### What to build

Jedes Gate, das heute nicht rot werden kann, wird scharfgestellt, und der Baum wird in
derselben Phase grün gemacht. Danach bedeutet „make check grün" dasselbe wie „CI grün".

### Acceptance criteria

- [ ] `frontend/package.json` trennt `"lint": "eslint --max-warnings=0 ."` von
      `"lint:fix": "eslint --fix --max-warnings=0 ."`, `make fmt-frontend` ruft
      `pnpm format && pnpm lint:fix`, und `make lint-frontend` bleibt der Gate-Aufruf.
      Befund: frontend/package.json:14
- [ ] `Makefile — lint-backend` verwendet denselben Exit-Guard wie `check-backend`
      (`if [ "$$(goimports -l . | wc -l)" -gt 0 ]; then goimports -l .; exit 1; fi`).
      Befund: Makefile:78-79
- [ ] `scripts/check-build-tags.sh` fordert für jede `backend/**/*_test.go` genau eine
      `//go:build`-Zeile mit `unit` oder `integration`. `make check` ruft das Skript mit,
      bis `make check-repo` in Kriterium 3.2 entsteht. `escpos/formatter_test.go` erhält
      `//go:build unit`; `api/health/health_integration_test.go` wird zu einem `unit`-Test
      umbenannt.
- [ ] `backend/.golangci.yml`, `Makefile — lint-backend-full`/`check-backend` und der
      CI-Job `backend-golangci` linten zusätzlich mit `--build-tags=unit`.
- [ ] Eine Root-`.prettierrc` und `.prettierignore` gelten für `ts, tsx, js, mjs, cjs,
json, css, md` im ganzen Repo. Ausgenommen sind `node_modules`, `dist`, Lockfiles,
      `backend/sqlc/dbgen`, `docs/rechtsquellen`, `docs/plans`,
      `frontend/src/components/ui`, `frontend/src/hooks/use-mobile.ts` und `*.astro`.
      `make check-format` prüft, ein einmaliger `--write`-Lauf macht den Baum grün.
      `frontend/.prettierrc` und `website/.prettierrc` entfallen; `check-frontend` und der
      CI-Job `frontend-ci` rufen `make check-format`.
- [ ] `.github/workflows/ci.yml` führt `shellcheck -x scripts/*.sh` aus. Die tautologische
      `if`-Bedingung am `e2e`-Job ist gelöscht; `changes` und `e2e` laufen ohne Bedingung.
      Jede verbleibende Job-Bedingung enthält `|| needs.changes.outputs.ci == 'true'`.
- [ ] `scripts/setup-dev-tools.sh` vergleicht zusätzlich die Go-Version, mit der die
      vorhandene `golangci-lint`-Binary gebaut wurde (`go version -m`), gegen
      `backend/go.mod`, und baut bei Abweichung neu.
- [ ] (Opus 5) Die durch Kriterium 2.1 sichtbaren ESLint-Verstöße sind behoben. Bei mehr
      als 50 Verstößen bricht der Worker ab und meldet Anzahl und Regelverteilung.
- [ ] (Opus 5) Die durch Kriterium 2.4 sichtbaren `unit`-Tag-Verstöße sind behoben. Bei
      mehr als 50 Verstößen bricht der Worker ab und meldet Anzahl und Regelverteilung.

---

## Phase 3: Prosa- und Verweis-Gate

**Depends on**: 2

**Modell**: Sonnet 5

**Review-Tier**: Gate only, Review gebündelt am Phasenende

**Gate**: `bash scripts/check-prose.sh`, `bash scripts/check-links.sh`, `make check`,
`make website-check`

**Hinweis**: Das Gate entsteht im ersten Kriterium und wird erst durch den Sweep dieser
Phase grün; gelandet wird nur der grüne Endstand.

**Choke-Dateien**: `Makefile`, `.github/workflows/ci.yml`

### Context

- `AGENTS.md` Regel 18 — der Regeltext zitiert die verbotenen Wörter selbst
- `database/migrations/01_initial.up.sql`, `06_favoriten_cleanup.up.sql` — eingefroren und
  enthalten „Phase 3, B2" beziehungsweise „bisher"
- `.claude/workflows/*.js` — Prompt-Texte zitieren die Wortliste
- `frontend/src/lib/utils.test.ts — 'ergänzt an früheren Tagen das Datum'` — korrekte
  Fachprosa, kein Historienbezug
- `docs/leitfaden/aktualisieren.md` — „Seit Version 0.17.3 …" ist ein datierter
  Änderungseintrag nach Regel 18
- `website/src/styles/brand.css`, `website/src/layouts/Landing.astro` — rund 90
  Kommentarstellen zitieren gelöschte PRDs und Phasennummern
- `scripts/prod-update.sh`, `scripts/prod-backup.sh`, `scripts/ops-smoke.sh`, `Makefile`,
  `.env.example` — fünf Verweise auf das nicht existierende `docs/leitfaden.md`
- `e2e/website/screenshots.mjs` — Kopf behauptet ein `data-theme`-Attribut; der
  ThemeProvider schaltet die Klassen `light`/`dark`

### What to build

Ein Grep-Gate über die verbotene Historien-Prosa und die toten Repo-Verweise, dazu der
Sweep, der beide Klassen im ganzen Baum auflöst. Jede Fundstelle wird zur Ist-Aussage
umgeschrieben, nie kommentarlos gelöscht, wenn sie eine Begründung trägt.

`reverse-proxy/caddyfile.go` bleibt unberührt; Kriterium 11.4 schreibt diesen Kommentar.

### Acceptance criteria

- [ ] `scripts/check-prose.sh` liest `git ls-files` und schlägt fehl bei den ganzen Wörtern
      `früher`, `frueher`, `vormals`, `bisher`, `bislang`, `neuerdings` sowie bei
      `Phase <Ziffer>`, `NEU<zwei Ziffern>`, `Muster <Ziffer>`, `Befund #`,
      `Design-Handoff`, `design_handoff`, `Seit Version <Ziffer>`, `Ab Version <Ziffer>`.
      Ausgenommen sind `CHANGELOG.md`, `docs/adrs/**`, `docs/plans/**`,
      `docs/rechtsquellen/**`, `database/migrations/**`, `backend/sqlc/dbgen/**`,
      `AGENTS.md`, `.github/copilot-instructions.md`, `.github/instructions/**`,
      `.claude/**`, `reverse-proxy/caddyfile.go` und Lockfiles. Eine versionierte
      Allowlist-Datei nennt jede weitere Ausnahme mit Grund; `frontend/src/lib/utils.test.ts`
      steht darin.
- [ ] `scripts/check-links.sh` prüft jeden im Repo genannten relativen `*.md`-Pfad auf
      Existenz (gleiche Quelle und Ausnahmen). `make check-repo` bündelt per Glob alle
      `scripts/check-*.sh`, `make check` ruft `check-repo` mit. Der `changes`-Job in
      `ci.yml` erhält einen `docs`-Filter (`docs/**`, `**/*.md`, `AGENTS.md`,
      `README.md`), und der neue Job `repo-checks` läuft bei `docs`, `scripts` oder `ci`.
- [ ] Die Historien-Prosa im Go-Code ist auf Ist-Aussagen umgeschrieben — mindestens
      `backend/app/routes.go`, `backend/config/config.go`, `backend/config/config_test.go`,
      `backend/repository/druckauftrag_repo/repo.go`,
      `backend/repository/tse_repo/repo_test.go`, `backend/api/fiskal/dsfinvk/mapper.go`,
      `backend/api/kasse/enrichment/enrichment.go`, `backend/api/druck/beleg/application/command.go`,
      `backend/seed/bondruck.go` und die Test-Kommentare in `backend/domain/kasse/`.
- [ ] Die Historien- und Handoff-Klauseln im Service-Bereich des Frontends sind entfernt,
      die geltende Aussage bleibt stehen: `service/components/ServiceDock.tsx`,
      `TischAuswahlDrawer.tsx`, `table/Zahlung.tsx`, `table/Bestellung.tsx`,
      `direktverkauf/Direktverkauf.tsx`.
      Befund: frontend/src/service/components/ServiceDock.tsx:4-7
- [ ] Website und E2E-Tooling nennen nur noch den Ist-Zustand: die PRD- und Planzitate in
      `website/src/**` und `e2e/website/**` sind ersatzlos gestrichen, der
      Übergangsregel-Absatz in `brand.css` ist durch die geltende Token-Aussage ersetzt,
      und `screenshots.mjs` beschreibt `emulateMedia({ colorScheme })` plus die
      `light`/`dark`-Klasse statt eines `data-theme`-Attributs.
      Befund: website/src/styles/brand.css:2-3, :11-17, website/src/layouts/Landing.astro:23-24,
      e2e/website/csp-check.mjs:1-12, e2e/website/csp-server.mjs:7-9,
      e2e/website/screenshots.mjs:12-15
- [ ] Die fünf Verweise auf `docs/leitfaden.md` zeigen auf die konkrete Seite:
      `scripts/prod-update.sh` und `.env.example` auf `aktualisieren.md`,
      `scripts/prod-backup.sh` auf `datenaufbewahrung.md`, `Makefile` auf
      `betriebsarten.md`, `scripts/ops-smoke.sh` auf `self-hosting.md`; zusätzlich sind
      die Planverweise in `cliff.toml`, `packaging/cron/jotti-backup.cron`,
      `packaging/systemd/jotti-backup.service`, `docker-compose.release.yml` und
      `windows/starter/rsrc_windows_amd64.syso` (per `make starter-syso` neu erzeugt)
      aufgelöst. Befund: scripts/prod-update.sh:111; scripts/prod-backup.sh:132;
      scripts/ops-smoke.sh:34; Makefile:231; .env.example:16
- [ ] Die Historien- und Handoff-Klauseln im Admin- und Komponentenbereich sind entfernt:
      `admin/reporting/UebersichtStatusZeile.tsx`, `admin/kasse/GeldtransitDialog.tsx`,
      `admin/finanzamt/LaeuftAllesSection.tsx`, `admin/reporting/SitzungsListe.tsx`,
      `admin/components/AdminPageHeader.tsx`, `admin/users/UserRolle.tsx`,
      `admin/users/Users.tsx`, `components/HinweisKarte.tsx`, `components/WarnKarte.tsx`,
      `components/ui/button.tsx`, `components/common/AuthLayout.tsx`,
      `components/common/VersionsHinweis.tsx`.
      Befund: frontend/src/admin/reporting/UebersichtStatusZeile.tsx:8-10,
      frontend/src/admin/components/AdminPageHeader.tsx:5-10

---

## Phase 4: Sprach- und Zeichen-Gate

**Depends on**: 3

**Modell**: Sonnet 5

**Review-Tier**: Gate only, Review gebündelt am Phasenende

**Gate**: `bash scripts/check-language.sh`, `make check`

**Hinweis**: Das Gate entsteht im ersten Kriterium und wird erst durch den Sweep dieser
Phase grün; gelandet wird nur der grüne Endstand.

**Choke-Dateien**: `.gitattributes`, `Makefile`

### Context

- `windows/starter/core/diagnose.go — Diagnose*-Konstanten` — dokumentiert die
  ASCII-Konvention ausdrücklich für Konsolenausgaben, nicht für Kommentare
- `windows/**` — 95 Zeilen mit Nicht-ASCII, davon 82 in Go-Kommentaren; dazu
  `jotti-start.manifest` und die Binärdatei `rsrc_windows_amd64.syso`
- `packaging/windows/*.cmd` — LF-Zeilenenden, ein Em-Dash; keine `.gitattributes` im Repo
- `backend/**` — deutsche Kommentare mischen echte Umlaute und `ae/oe/ue` in denselben
  Dateien
- `frontend/src/components/ui/dialog.tsx`, `sheet.tsx`, `sidebar.tsx`, `spinner.tsx` —
  englische Screenreader-Texte

### What to build

Eine entscheidbare Sprachregel je Bereich, als Skript geprüft: Windows-Ausgaben bleiben
ASCII, Go-Kommentare im Backend tragen echte Umlaute, Benutzer-sichtbare Strings sind
deutsch. Bezeichner ändern sich nicht.

### Acceptance criteria

- [ ] `scripts/check-language.sh` liest `git ls-files` und schlägt fehl bei
      Nicht-ASCII-Bytes in Go-String-Literalen unter `windows/**` und in
      `packaging/**/*.cmd`. Kommentare, `*.manifest` und `*.syso` sind ausgenommen; sie
      tragen deutsche Prosa und stehen nie auf der Konsole. Zusätzlich schlägt es fehl bei
      den transliterierten Formen `fuer, ueber, koennen, muessen, waehrend, naechst,
auftraege, aenderung, gemaess, zurueck, moeglich, spaeter, aendern, pruefen, laeuft,
haelt, groesse, schliessen, genuegt, einfuehrung` als ganzes Wort in Kommentarzeilen
      unter `backend/`.
- [ ] `.gitattributes` im Repo-Root enthält `*.cmd text eol=crlf`, und die drei
      `packaging/windows/*.cmd` sind mit CRLF und ohne Em-Dash eingecheckt.
- [ ] Die gedruckten Strings in `windows/starter/backup.go`, `windows/starter/main.go`,
      `windows/starter/system.go`, `windows/starter/core/diagnose.go`,
      `windows/relay/env.go` und `windows/relay/main.go` sind reines ASCII („—" → „-",
      „→" → „->", „ü" → „ue"). Befund: windows/starter/backup.go:67,69;
      windows/starter/main.go:238; windows/starter/system.go:58,114,251;
      windows/relay/env.go:84; windows/relay/main.go:148
- [ ] Die deutsche Prosa in Go-Kommentaren unter `backend/` schreibt durchgängig echte
      Umlaute, und die vier englischen Screenreader-Texte lauten „Schließen",
      „Seitenleiste", „Zeigt die mobile Seitenleiste." und „Wird geladen".
      Befund: frontend/src/components/ui/dialog.tsx:93 (ebenso sheet.tsx:90,
      sidebar.tsx:195-197, spinner.tsx:6)

---

## Phase 5: Fehler-Kontrakt über alle Schichten

**Depends on**: 2, 4

**Modell**: Opus 5

**Review-Tier**: Gate + Sweep + Opus-Review

**Gate**: `make check`, `cd backend && go test -tags=unit ./api/...`,
`cd frontend && pnpm test`

**Choke-Dateien**: `backend/api/helper/http.go`,
`backend/api/stammdaten/produkt/application/command.go`,
`frontend/src/lib/errorMessages.ts`

### Context

- `backend/api/helper/http.go — MapError()` — wählt den Code über Map-Iteration; heute
  matcht in keiner Karte mehr als ein Sentinel
- `backend/api/helper/http.go — SendConflictDetails()` — der Doc-Kommentar verspricht im
  `signaturen_ausstehend`-Detail ein Alter, das keine Schicht führt
- `backend/api/kasse/enrichment/enrichment.go — ErrProduktNotFound, ErrVarianteNichtAktiv`,
  `backend/api/fiskal/export/application/export.go — ErrKassensitzungNichtGefunden`,
  `backend/api/fiskal/dsfinvk/mapper.go — ErrKeineVorgaenge`,
  `backend/api/reporting/application/query.go — ErrDatabase` — exportierte Sentinels
  außerhalb von `errors.go`
- `backend/api/kasse/tischgeschaeft/application/command.go — persistTischEvent(),
BestellungAufnehmen()` — verlieren `ErrKasseNichtGeoeffnet` im `ErrDatabase`-Fallback;
  `persistStornoEvents()` macht es richtig
- `backend/api/stammdaten/produkt/application/command.go — DeleteVariante()` — lädt das
  Produkt, verwirft es und löscht jede fremde `varianteId`
- `backend/api/stammdaten/user/application/command.go`,
  `backend/api/stammdaten/produkt/application/command.go`,
  `backend/api/stammdaten/produkt/application/query.go`,
  `backend/api/auth/application/command.go`,
  `backend/api/fiskal/setup/application/setup.go` — 20 `log.Error()`-Ketten in
  `if err != nil`-Zweigen ohne `.Err(err)`
- `frontend/src/lib/errorMessages.ts — commonErrorMessages` — kennt weder
  `login_throttled` noch `invalid_kassensitzung`, führt aber `kassensturz_erforderlich`

### What to build

Ein Test, der jeden Anwendungs-Sentinel gegen die HTTP-Karten prüft, plus die Stellen, an
denen heute ein 500 statt eines 400/409 herauskommt, plus die Frontend-Seite desselben
Vertrags. Dazu zwei Härtungen im selben Thema: eine deterministische `MapError`-Auswahl und
Fehler-Logs, die ihren Fehler mitführen.

### Acceptance criteria

- [ ] `backend/api/error_mapping_contract_test.go` (`//go:build unit`) parst per
      `go/parser` alle exportierten `Err*`-Sentinels unter `backend/api/**` — auch
      `enrichment.go`, `export.go`, `mapper.go` und `query.go` — und alle
      `helper.MapError`-Karten in `api/**/http/*.go`. Der Test schlägt fehl, sobald ein
      Sentinel in keiner Karte und in keiner im Test dokumentierten Ausnahmeliste steht.
- [ ] `TischAktualisierenHandler` mappt `application.ErrTischAlreadyExists` auf
      `tisch_already_exists`, und ein Handler-Test sichert 400 samt Code zu.
      Befund: backend/api/stammdaten/tisch/http/command_handler.go:77-84, :78-84
- [ ] `produkt/application/command.go — UpdateProdukt()` bildet `db.ErrAlreadyExists` auf
      `ErrProduktAlreadyExists` ab, der Handler mappt auf `produkt_already_exists`, und
      ein Test sichert den Weg bis zum Code zu.
      Befund: backend/api/stammdaten/produkt/application/command.go:76-80, :76-80
- [ ] `tischgeschaeft/application/command.go` reicht `ErrKasseNichtGeoeffnet` in
      `persistTischEvent()` und `BestellungAufnehmen()` vor dem `ErrDatabase`-Fallback
      durch, sodass der Handler 409 `kasse_nicht_geoeffnet` liefert; ein Test deckt beide
      Pfade. Befund: backend/api/kasse/tischgeschaeft/application/command.go:135-147
- [ ] Der 409-Doc-Kommentar in `helper/http.go — SendConflictDetails()` nennt nur noch die
      Anzahl ausstehender Signaturen, kein Alter.
      Befund: backend/api/helper/http.go:23-25
- [ ] `commonErrorMessages` enthält deutsche Texte für `login_throttled` und
      `invalid_kassensitzung`, der tote Eintrag `kassensturz_erforderlich` ist samt
      Testzeile gelöscht, und `errorMessages.test.ts` iteriert über
      `Object.entries(commonErrorMessages)` statt über eine zweite Liste.
- [ ] `helper.MapError()` nimmt eine geordnete `{error, code}`-Liste entgegen (erster
      Treffer gewinnt) und hat Unit-Tests für Treffer, Nicht-Treffer und gewrappte Fehler.
      Die Änderung ist verhaltensneutral: heute matcht in keiner der 34 Karten mehr als ein
      Sentinel. Die geordnete Liste hält das konstruktiv fest.
- [ ] `produkt/application/command.go — DeleteVariante()` prüft, dass die geladene Variante
      zum geladenen Produkt gehört, und lehnt sonst mit `ErrVarianteNotFound` ab; ein Test
      „fremde Variante wird nicht gelöscht" sichert das zu.
      Befund: backend/api/stammdaten/produkt/application/command.go:245-278
- [ ] `backend/api/log_error_contract_test.go` (`//go:build unit`) schlägt per `go/parser`
      fehl, sobald eine `log.Error()`-Kette in einem `if err != nil`-Block kein `.Err(`
      trägt. Die 20 heutigen Fundstellen in `user/application/command.go`,
      `produkt/application/command.go`, `produkt/application/query.go`,
      `auth/application/command.go` und `fiskal/setup/application/setup.go` sind ergänzt.
      Befund: backend/api/stammdaten/produkt/application/command.go:49
- [ ] Die `byCode`-Umformulierungen in `DirektverkaufAbschluss.tsx`,
      `DirektverkaufStornoDrawer.tsx` und `DirektverkaufHistorie.tsx` entfallen; ein
      `byCode`-Eintrag ergänzt nur noch Kontext zur zentralen Meldung.

---

## Phase 6: Schema-Grenzen und Trim

**Depends on**: 2, 3, 5

**Modell**: Opus 5

**Review-Tier**: Gate + Sweep + Opus-Review

**Gate**: `make check`, `cd backend && go test -tags=unit ./... `,
`cd frontend && pnpm test`

**Choke-Dateien**: `backend/api/fiskal/dsfinvk/mapper.go`

### Context

- `backend/domain/betreiber/betreiber.go` — Stammdaten ohne Maximallängen; DSFinV-K 2.4
  `index.xml` begrenzt Name und Straße auf 60, PLZ 10, Ort 62, StNr 20, UstID 15
- `database/migrations/01_initial.up.sql — Tabelle betreiber` — die Spalten sind `TEXT`
  ohne Längenbegrenzung; Bestandswerte können die amtliche MaxLength überschreiten
- `backend/api/fiskal/dsfinvk/mapper.go — stammdatenZeile()` — schreibt Vereinsname,
  Straße, PLZ und Ort ungekürzt in die DSFinV-K-Stammdaten
- `backend/api/stammdaten/betreiber/http/command_handler.go — updateBetreiberSchema` —
  zweite Wahrheitsquelle mit eigenen Meldungen; der `NewBetreiber`-Zweig antwortet 500
- `backend/domain/**` — 22 exportierte `*Schema`-Variablen, davon elf Enum- oder
  Struct-Schemas ohne Längengrenze; `backend/domain/` selbst hat keine Go-Dateien
- `frontend/src/admin/tables/Tisch.ts`, `admin/users/User.ts`, `lib/identity.ts`,
  `admin/finanzamt/BetreiberBackend.ts` — Namensschemas ohne `.trim()`/`.min()`

### What to build

Jedes persistierte Feld bekommt eine obere Schranke, und die HTTP-Schicht bezieht ihre
Schemas aus der Domäne statt sie zu kopieren. Die Zod-Seite spiegelt dieselben Grenzen, wie
Regel 5 es fordert. Der Exportrand kürzt, was Bestandsdaten mitbringen.

### Acceptance criteria

- [ ] `backend/api/schema_grenzen_test.go` (`//go:build unit`, Paket `api`) führt eine
      Tabelle aller persistierten Feld-Schemas mit erwarteter Unter- und Obergrenze und
      prüft je Eintrag Annahme und Ablehnung. Er ermittelt per `go/parser` alle
      exportierten `*Schema`-Variablen unter `backend/domain/**`; ein fehlender
      Tabelleneintrag macht ihn rot. Enum- und Struct-Schemas stehen in einer im Test
      dokumentierten Ausnahmeliste, ebenso der Alias `produkt.SteuersatzSchema`.
- [ ] `domain/betreiber/betreiber.go` begrenzt Vereinsname und Straße auf 60, PLZ auf 10,
      Ort auf 62, Steuernummer auf 20 und USt-IdNr. auf 15 Zeichen und trimmt jedes Feld;
      ein Test lehnt einen 61-Zeichen-Namen ab. Befund: backend/domain/betreiber/betreiber.go:24-32
- [ ] `betreiber/http/command_handler.go` baut `updateBetreiberSchema` aus den
      exportierten Feld-Schemas von `domain/betreiber` (wie user/tisch/produkt), und der
      `NewBetreiber`-Fehlerzweig liefert 400 statt 500.
      Befund: backend/api/stammdaten/betreiber/http/command_handler.go:31-51
- [ ] Die Zod-Schemas in `admin/tables/Tisch.ts`, `admin/users/User.ts`,
      `lib/identity.ts` und `admin/finanzamt/BetreiberBackend.ts` trimmen und begrenzen
      dieselben Felder wie ihre zog-Gegenstücke. Eine gemeinsame Namensschema-Quelle
      ersetzt die drei Kopien derselben Meldung; ein Vitest prüft deren Grenzen.
- [ ] `dsfinvk/mapper.go` kürzt Vereinsname, Straße, PLZ und Ort runensicher auf die
      amtlichen Längen, bevor sie in die Stammdatenzeile gehen; ein Mapper-Test prüft einen
      70-Zeichen-Vereinsnamen. Bestandsdaten bleiben in der Datenbank unverändert.

---

## Phase 7: Kassensitzungs-Status und Event-JSON-Verträge

**Depends on**: 1, 2, 5

**Modell**: Opus 5

**Review-Tier**: Gate + Sweep + Opus-Review + Skeptiker je Kriterium

**Gate**: `make verify`

**Choke-Dateien**: `backend/.golangci.yml`,
`backend/api/kasse/kassenfuehrung/application/command.go`

### Context

- `backend/repository/kassensitzungen_repo — GetOffeneKassensitzung()` — sieht den
  Barrierestatus `wird_abgeschlossen` nicht; `GetAktiveKassensitzung()` schon
- `backend/repository/kassensitzungen_repo — GetOffeneKassensitzungNr()` — ruft
  `GetOffeneKassensitzung` intern auf; das Gate-Muster darf sie nicht treffen
- `backend/api/kasse/tischgeschaeft/application/query.go` — fünf Aufrufe; bei `nil` wird
  `kassensitzungNr = 0` und damit ein anderes `TischSessionSubject` gelesen
- `backend/api/kasse/direktverkauf/application/query.go`,
  `backend/api/reporting/application/query.go`,
  `backend/api/fiskal/export/application/export.go` — je ein weiterer Aufruf
- `backend/api/kasse/kassenfuehrung/application/query.go — GetOffeneKassensitzung()` —
  Lesepfad der Kassentag-Seite
- `backend/api/fiskal/setup/application/command.go — kassensitzungReader` — Guard vor dem
  TSS-Wechsel
- `backend/sqlc/queries/kassensitzungen.sql — GetKassenbestand` — zieht die gebuchte
  Differenz ab; nach Schritt 2 entspricht der Soll-Bestand dem gezählten Ist-Bestand
- `backend/api/kasse/kassenfuehrung/application/command.go — findeVorhandenenKassensturz()`
  — erkennt den Wiederanlauf, prüft aber den Soll-Bestand nicht
- `backend/api/druck/bondruck/application/arbeitsbon_policy.go` — dekodiert Event-JSON in
  `kasse.Position` statt `kasse.PositionEventData`; `kassenbeleg_command.go` macht es
  richtig

### What to build

Die Statusgrenze `wird_abgeschlossen` wird an jedem Lesepfad und jedem Guard respektiert,
und Event-JSON wird ausschließlich über die Vertragstypen dekodiert. Beide Klassen
bekommen ein Gate, das den Rückfall meldet.

### Acceptance criteria

- [ ] `backend/.golangci.yml` verbietet per `forbidigo` das verankerte Muster
      `KassensitzungenRepo\.GetOffeneKassensitzung\(`. Es trifft weder
      `GetOffeneKassensitzungNr` noch den repository-internen Aufruf. Diese Aufrufe
      wechseln auf `GetAktiveKassensitzung`:

  - `tischgeschaeft/application/query.go:36,69,101,135,182` — sonst zeigt jeder Tisch
    im Barrierestatus leer und mit Saldo 0.
  - `direktverkauf/application/query.go:26` — sonst bleibt die Historie leer.
  - `reporting/application/query.go:307` — sonst meldet die Live-Sicht „keine Sitzung".

  Unverändert bleibt `fiskal/export/application/export.go:141` mit begründeter
  `//nolint:forbidigo`-Zeile: der Export will genau die offene Sitzung.
  `reporting/application/query.go:288` nutzt `GetOffeneKassensitzungNr` und bleibt
  vom Muster unberührt.

- [ ] `kassenfuehrung/application/query.go` liest über `GetAktiveKassensitzung`, und die
      Anwendungsmethode trägt denselben Namen. Umbenannt werden:

  - Interface und Handler in `kassenfuehrung/http/query_handler.go`
    (`GetAktiveKassensitzungHandler`).
  - Mock und die drei Tests in `kassenfuehrung/http/query_handler_test.go`.
  - Die Route in `api/admin.go` (`/get-aktive-kassensitzung`) und ihr Aufruf in
    `admin/kasse/KasseBackend.ts`.

  `frontend/src/admin/kasse/KassensitzungPage.tsx` zeigt bei `wird_abgeschlossen`
  Schritt 3 mit dem Hinweis „Abschluss unterbrochen — erneut abschließen" statt des
  Eröffnen-Formulars; ein Integrationstest und ein Vitest decken den Fall ab.
  Befund: backend/api/kasse/kassenfuehrung/application/query.go:14-24

- [ ] `fiskal/setup/application/command.go` prüft den TSE-Konfigurationsguard gegen
      `GetAktiveKassensitzung`, und der Kommentar nennt „offen oder wird_abgeschlossen";
      ein Test lehnt den TSS-Wechsel im Barrierestatus ab.
      Befund: backend/api/fiskal/setup/application/command.go:17-44
- [ ] `kassenfuehrung/application/command.go` erkennt Zwischenbuchungen nach dem
      protokollierten Kassensturz auch dann, wenn die Differenzbuchung bereits geschrieben
      ist. Verglichen wird der Soll-Bestand **ohne** die abschluss-eigene Differenzbuchung
      gegen `sturz.SollBestandCents`; bei Abweichung bricht der Abschluss mit
      `ErrBuchungenNachKassensturz` ab.
      Befund: backend/api/kasse/kassenfuehrung/application/command.go:364-376

  - Integrationstest: eine Tischzahlung nach dem Kassensturz bricht den Wiederanlauf ab.
  - Integrationstest: ein Wiederanlauf nach geschriebener Differenzbuchung läuft durch.

- [ ] `arbeitsbon_policy.go` dekodiert in `[]kasse.PositionEventData` und wandelt über
      `kasse.PositionFromEventData`, `arbeitsbon_policy_test.go` baut seine Events mit
      `kasse.NewBestellungAufgenommenEvent` bzw. `kasse.NewDirektverkaufGetaetigtEvent`,
      und `backend/api/event_decode_contract_test.go` (`//go:build unit`) schlägt per
      `go/parser` fehl, sobald ein Decode-Ziel unter `backend/api/**` ein Feld eines
      `domain/kasse`-Typs ohne `EventData`-Suffix trägt.
      Befund: backend/api/druck/bondruck/application/arbeitsbon_policy.go:15-21, :18-21,
      backend/api/druck/bondruck/application/arbeitsbon_policy_test.go:18-47

---

## Phase 8: Fiskal- und Kassen-Korrektheit

**Depends on**: 3, 6, 7

**Modell**: Opus 5

**Review-Tier**: Gate + Sweep + Opus-Review + Skeptiker je Kriterium

**Gate**: `make verify`, `bash scripts/check-timezone.sh`

**Choke-Dateien**: `backend/api/kasse/kassenfuehrung/application/command.go`,
`backend/api/fiskal/dsfinvk/mapper.go`

### Context

- `backend/domain/user/password.go — PasswordSchema` — `Trim().Min(6).Max(72)`
- `backend/api/auth/http/command_handler.go — loginSchema` — `z.String().Min(1)`, ohne
  `.Trim()`; ein Konto mit Rand-Leerzeichen bleibt unbenutzbar
- `backend/api/middleware/middleware.go — RequireAuth()` — legt den Benutzernamen aus dem
  JWT-Claim in den Context, obwohl der Datensatz bereits geladen ist
- `backend/domain/kasse/tisch_session_events.go — buildUmbuchungKommentar()` — kürzt auf
  Runen, das eingefrorene Schema prüft Bytes
- `backend/api/fiskal/dsfinvk/mapper.go — abrechnungskreis()` — DSFinV-K begrenzt
  `ABRECHNUNGSKREIS` auf 50 Zeichen, Tischnamen dürfen 100 haben
- `backend/api/fiskal/dsfinvk/table.go — column` — `typ` und `accuracy` werden gesetzt und
  nie gelesen
- `backend/repository/kassenjournal_repo/repo.go — Doc zu EroeffneKassensitzung` —
  beschreibt einen Ablauf, den der Code nicht hat
- `backend/sqlc/queries/betreiber.sql — SetElsterGemeldetAm` — `CURRENT_DATE` der
  DB-Sitzung; die Container laufen in UTC
- `backend/api/fiskal/export/application/export.go — dateiname()` — formatiert den
  ZIP-Namen ohne Zone

### What to build

Die verbleibenden bestätigten Korrektheitsfehler auf den geldführenden und fiskalischen
Pfaden, dazu die zwei Zeitzonen-Ränder außerhalb des Druckpfads. Jeder Fix kommt mit dem
Test, der ihn festhält.

### Acceptance criteria

- [ ] Das Login normalisiert das Passwort wie das Setzen: `loginSchema` verwendet
      `z.String().Trim().Min(1, …).Required()`. Die Längenregeln des `PasswordSchema`
      bleiben draußen, damit der Login-Endpunkt die Passwort-Policy nicht verrät. Ein Test
      setzt ein Passwort mit umgebenden Leerzeichen und meldet sich damit an.
      Befund: backend/domain/user/user.go:202-228, :202-228
- [ ] `middleware.go` legt `u.Username` aus dem geladenen Datensatz in den Context, und
      `middleware_test.go` sichert zu, dass ein veralteter Claim-Name nicht mehr ins
      Kassenjournal gelangt. Befund: backend/api/middleware/middleware.go:269-304
- [ ] `buildUmbuchungKommentar()` kürzt byteweise auf die Schemagrenze, das eingefrorene
      Event-Schema bleibt unverändert, und ein Test bucht einen Tisch mit Umlaut-Namen an
      der Grenze um. Befund: backend/domain/kasse/tisch_session_events.go:118-128
- [ ] `abrechnungskreis()` kürzt runensicher auf 50 Zeichen, und ein Mapper-Test prüft
      einen 100-Zeichen-Tischnamen gegen die amtliche MaxLength.
      Befund: backend/api/fiskal/dsfinvk/mapper.go:535-544
- [ ] `dsfinvk/table.go — column` trägt nur noch `name`, die Zuweisungen in den
      `build*`-Funktionen sind entfernt, und der Doc-Kommentar beschreibt die eingebettete
      amtliche `index.xml` statt einer Erzeugung.
      Befund: backend/api/fiskal/dsfinvk/table.go:13-34
- [ ] Der Doc-Kommentar in `kassenjournal_repo/repo.go` beschreibt den echten Ablauf
      (`EroeffneKassensitzung` schreibt in derselben Transaktion), die Routing-Zeile nennt
      `UPDATE`, und `repo_test.go` verliert den zusicherungsfreien
      `TestGetOffeneKassensitzung_NoneOpen`; die Zusicherung wandert nach
      `kassensitzungen_repo`. Befund: backend/repository/kassenjournal_repo/repo.go:264-267,
      backend/repository/kassenjournal_repo/repo_test.go:1583-1592
- [ ] Der Doc-Kommentar zu den Z-Bon-Summen in
      `kassenfuehrung/application/command.go` nennt `kasse.ComputeAbschlussSummen` und den
      Äquivalenz-Guard `reporting_repo/summen_abschluss_test.go`.
      Befund: backend/api/kasse/kassenfuehrung/application/command.go:260-261
- [ ] Das ELSTER-Meldedatum entsteht in `Europe/Berlin` statt aus `CURRENT_DATE` der
      DB-Sitzung: die Anwendungsschicht übergibt das Datum, die Query nimmt es als
      Parameter. Ein Test mit `2026-07-01T23:30:00Z` erwartet den 02.07.2026.
      Befund: backend/api/stammdaten/betreiber/application/command.go:31-42
- [ ] `scripts/check-timezone.sh` schlägt fehl, sobald in `backend/api/druck/**`,
      `backend/api/fiskal/**` oder `backend/api/reporting/**` ein `.Format(` steht, dessen
      Empfänger nicht durch `.In(` läuft. Eine versionierte Allowlist nennt jede Ausnahme
      mit Grund. `export.go — dateiname()` formatiert den ZIP-Namen als
      `zeitpunkt.In(berlin)`; ein Test prüft den Namen für einen UTC-Mitternachtszeitpunkt.

---

## Phase 9: Repository-Konsolidierung und Benennung

**Depends on**: 8

**Modell**: Sonnet 5

**Review-Tier**: Gate + Sweep + Opus-Review

**Gate**: `make verify`

**Choke-Dateien**: `backend/repository/tisch_repo/repo.go`

### Context

- `backend/repository/user_repo/types.go` — drei byte-identische Row-Mapper
- `backend/repository/produkt_repo/repo.go — GetAllProdukte(), GetActiveProdukte()` —
  zeilengleiche Varianten-JSON-Abbildung, dreifach
- `backend/repository/tisch_repo/repo.go` — öffentliche Methoden heißen `…Table`, DB,
  sqlc und Domäne sagen `Tisch`
- `backend/repository/reporting_repo/repo.go` — einziges Repository ohne `db.Error()`

### What to build

Vier mechanische, compilergeprüfte Konsolidierungen in der Persistenzschicht. Verhalten
bleibt gleich; die vorhandenen Integrationstests sind der Beweis.

### Acceptance criteria

- [ ] `user_repo/types.go` behält einen Row-Mapper, die Aufrufer konvertieren am
      Aufrufort, und kein Lesepfad verliert dadurch ein Feld.
      Befund: backend/repository/user_repo/types.go:19-62
- [ ] `produkt_repo` besitzt eine `produktRowToDomain`-Abbildung in `types.go`, aus der
      alle drei Lesepfade lesen. Befund: backend/repository/produkt_repo/repo.go:38-92
- [ ] Die öffentlichen Methoden von `tisch_repo` tragen deutsche Domänen-Nomen
      (`GetTisch`, `GetAlleTische`, `CreateTisch`, `UpdateTisch`, `DeleteTischMitFavoriten`
      …), und die vier Consumer-Interfaces sind mitgezogen.
      Befund: backend/repository/tisch_repo/repo.go:11-146
- [ ] `reporting_repo/repo.go` normalisiert jeden Rückgabefehler mit `db.Error(err)` wie
      die sechs Schwester-Repositories.

---

## Phase 10: Frontend — Korrektheit, Schemata, geteilte Bausteine

**Depends on**: 1, 2, 3, 5

**Modell**: Opus 5

**Review-Tier**: Gate + Sweep + Opus-Review

**Gate**: `make check`, `cd frontend && pnpm test`, `make test-e2e`

**Choke-Dateien**: `frontend/package.json`, `frontend/pnpm-lock.yaml`,
`frontend/eslint.config.js`

### Context

- `frontend/src/service/table/hooks.ts — useAktiveTischeMitFavoriten(),
useMeineTischeState(), useEigeneUebersicht()` — verwerfen `isError` und liefern Defaults
- `frontend/src/admin/reporting/ReportingResults.tsx` — druckt den Event-Rohwert
  (Soll − Ist); `KasseAbschliessenSection.tsx` zeigt Ist − Soll
- `frontend/src/admin/finanzamt/LaeuftAllesSection.tsx` — Klartext hängt allein an
  `offene === 0`, `letzterFehler` wird nie gezeigt
- `frontend/src/components/ui/sonner.tsx` — liest das Theme aus `next-themes` ohne Provider
- `frontend/package.json` — sieben Runtime-Dependencies ohne jeden Import
- `frontend/src/service/product/Produkt.ts` — zweite, bereits gedriftete Produkt-Definition
- `frontend/src/components/common/FormFields.tsx` — einzige Aufwärts-Import-Verletzung;
  `frontend/eslint.config.js` lädt kein `eslint-plugin-import`

### What to build

Die vier Frontend-Fehler, die falsche Tatsachen anzeigen, dazu die Konsolidierung der
doppelten Schemas und Bausteine und eine Import-Regel, die die Schichtrichtung festhält.

### Acceptance criteria

- [ ] Die Service-Hooks in `service/table/hooks.ts`, `service/direktverkauf/hooks.ts` und
      `service/product/hooks.ts` geben `isError` heraus, und `TableSelectionPage` sowie
      die weiteren Konsumenten rendern `LadefehlerAlert` statt Nullwerten; ein Vitest je
      Seite sichert zu, dass eine Fehlladung keine „0,00 €" zeigt.
      Befund: frontend/src/service/table/hooks.ts:56-90
- [ ] `ReportingResults.tsx` zeigt die Kassensturz-Differenz in derselben Perspektive wie
      der Abschluss-Bildschirm (Ist − Soll, `formatEuroMitVorzeichen`), und der Test in
      `ReportingResults.test.tsx` prüft das Vorzeichen eines Fehlbetrags.
      Befund: frontend/src/admin/reporting/ReportingResults.tsx:35-39
- [ ] `LaeuftAllesSection.tsx` meldet fehlgeschlagene Signaturen unabhängig von `offene`
      zuerst, führt `letzterFehler` und einen Fehler-Zähler in den Kennzahlen und
      beruhigt nur unterhalb von `RUECKSTAND_WARN_SEKUNDEN`; ein Vitest deckt den
      Fehlerfall ab. Befund: frontend/src/admin/finanzamt/LaeuftAllesSection.tsx:65-110
- [ ] `components/ui/sonner.tsx` bezieht das Theme aus `@/components/theme-provider`.
      Befund: frontend/src/components/ui/sonner.tsx:1-8
- [ ] `frontend/src/test/input-otp.ts` ist gelöscht, und die `afterEach`-Aufrufe in
      `OTPField.test.tsx` und `PasswordForm.test.tsx` sind entfernt.
      Befund: frontend/src/test/input-otp.ts:1-16
- [ ] Produkt, Variante, Kategorie, Steuersatz und `EntityStatus` haben je ein
      Response-Schema in einem geteilten Modul, das Admin- und Service-Bereich
      importieren; Formular- und Eingaberegeln bleiben im Admin-Formular.
      Befund: frontend/src/service/product/Produkt.ts:24-55
- [ ] `AbschlussContainer` und `BarzahlungFelder` existieren als geteilte Komponenten und
      werden aus `BestellungAbschluss.tsx`, `ZahlungAbschluss.tsx` und
      `DirektverkaufAbschluss.tsx` gerendert; das Markup bleibt unverändert.
      Befund: frontend/src/service/components/table/BestellungAbschluss.tsx:129-143,
      frontend/src/service/components/table/ZahlungAbschluss.tsx:123-170
- [ ] `TSEEinrichtungWizard.test.tsx` deckt die LIVE-Sperren ab: Button gesperrt bis
      `tippBestaetigung === 'LIVE'`, `tse_setup_pin_unbekannt`, `istEinsatzbereit` ohne
      PIN und `nurDisabledOderLeer`.
      Befund: frontend/src/admin/tse/TSEEinrichtungWizard.test.tsx:77-167
- [ ] `frontend/eslint.config.js` verbietet mit der Kernregel `no-restricted-imports`
      (`patterns: ['@/admin/*', '@/service/*']`) im `files`-Override für
      `src/components/**`, `src/lib/**` und `src/hooks/**` den Aufwärts-Import. Keine neue
      Abhängigkeit. `FormFields.tsx` ist die einzige heutige Verletzung und aufgelöst.
- [ ] `next-themes`, `@base-ui/react`, `cmdk`, `date-fns`, `embla-carousel-react`,
      `react-day-picker`, `react-resizable-panels` und `recharts` sind aus
      `frontend/package.json` und dem Lockfile entfernt.
      Befund: frontend/package.json:21-43

---

## Phase 11: Ops, Windows und Reverse-Proxy

**Depends on**: 1, 3, 4

**Modell**: Opus 5

**Review-Tier**: Gate + Sweep + Opus-Review + Skeptiker je Kriterium

**Gate**: `make check`, `shellcheck -x scripts/*.sh`, `bash scripts/check-pins.sh`

**Choke-Dateien**: `packaging/windows/KURZANLEITUNG.md`, `reverse-proxy/main.go`,
`reverse-proxy/caddyfile.go`, `e2e/package.json`

### Context

- `packaging/windows/jotti-restore.cmd`, `jotti-repair.cmd` — starten den Stack ohne
  `LAN_IP`; `docker-compose.release.yml` interpoliert die Variable
- `packaging/windows/KURZANLEITUNG.md — Abschnitt „Nur vorwärts, kein Downgrade"` —
  widerspricht dem Rückweg über das vorherige Release-ZIP nach einem Restore
- `packaging/windows/KURZANLEITUNG.md — manueller Backup-Befehl` — schreibt in einen
  Ordner, den erst `mirrorBackupToHost` anlegt
- `reverse-proxy/caddyfile.go — wildcardSite(), contentSecurityPolicy` — setzt
  `in.state.Subdomain` ungequotet in die Site-Adresse; der Kommentar zitiert das frühere
  nginx-Setup
- `reverse-proxy/main.go — main(), runLANMode(), ensureState()` — leeres `JOTTI_DOMAIN`
  fällt still in den LAN-Modus; die Caddyfile entsteht dreimal mit `0o644`
- `scripts/prod-restore.sh — decompress()` — spielt ein Archiv ohne Integritätsprüfung ein
- `scripts/test-integration.sh`, `scripts/test-tse-live.sh` — `postgres:17` statt `17.8`
- `e2e/package.json — packageManager` — `pnpm@11.6.0` ohne den sha512-Hash, den
  `frontend` und `website` tragen

### What to build

Die bestätigten Ops-Befunde: Recovery-Pfade, die tun was sie sagen, ein Proxy ohne
Injektionsfläche und stille Modus-Umschaltung, und Testharnesse mit denselben Pins wie die
Produktion.

### Acceptance criteria

- [ ] `jotti-restore.cmd` und `jotti-repair.cmd` starten den Stack nicht mehr selbst,
      sondern enden nach der Datenbankarbeit mit dem Verweis auf `jotti-start.exe`. Die
      `KURZANLEITUNG.md` beschreibt diesen Ablauf und löst den Widerspruch zu „Nur
      vorwärts, kein Downgrade" auf: nach einem Restore ist das vorherige Release der
      richtige Stand.
      Befund: packaging/windows/jotti-restore.cmd:36-37, packaging/windows/jotti-repair.cmd:15,46
- [ ] Der dokumentierte manuelle Backup-Befehl in `KURZANLEITUNG.md` legt
      `%PROGRAMDATA%\jotti\backups` mit `md … 2>nul` an, trägt einen Zeitstempel im
      Dateinamen und prüft auf eine nicht leere Datei.
      Befund: packaging/windows/KURZANLEITUNG.md:83-91, :85-91
- [ ] `InstallState.valid()` akzeptiert die acme-dns-Subdomain nur gegen
      `^[a-z0-9-]{1,63}$`, und ein Test lehnt eine Subdomain mit Leerzeichen oder
      geschweifter Klammer ab. Befund: reverse-proxy/caddyfile.go:201-221
- [ ] Ein Test in `package main` liest `reverse-proxy/nginx.rocks.conf` und sichert zu,
      dass sie die Konstante `contentSecurityPolicy` wörtlich enthält. Die beiden
      Prosakommentare in `caddyfile.go` und `nginx.rocks.conf` verweisen auf den Test und
      sind zugleich von der Historien-Prosa befreit.
      Befund: reverse-proxy/caddyfile.go:5-8; reverse-proxy/nginx.rocks.conf:142-143
- [ ] `loadConfig` verweigert den LAN-Modus ohne State-Verzeichnis, sodass ein leeres
      `JOTTI_DOMAIN` den Public-Stack nicht mehr still in den LAN-Modus fallen lässt, und
      `docker-compose.prod.yml` erzwingt die Variable am `reverse-proxy`-Service mit
      `${JOTTI_DOMAIN:?JOTTI_DOMAIN ist im Public-Stack Pflicht}` (Eigentümer-Entscheidung,
      Regel 16). Ein Test in `package main` deckt den `loadConfig`-Fall ab.
      Befund: reverse-proxy/main.go:78-86
- [ ] Die Status-Seite wiederholt `ensureState` im Hintergrund und rendert nach Erfolg neu
      — oder ihr Hinweistext nennt „jotti neu starten"; der Text und das Verhalten stimmen
      überein. Befund: reverse-proxy/main.go:134-158
- [ ] `scripts/prod-restore.sh` prüft ein `*.gz`-Archiv vor der Bestätigungsabfrage mit
      `gzip -t` und bricht bei Fehler ab, bevor Objekte gedroppt werden.
      Befund: scripts/prod-restore.sh:99-116
- [ ] `scripts/check-pins.sh` prüft die Versions-Pins und landet grün.
      Befund: scripts/test-integration.sh:34; scripts/test-tse-live.sh:51

  - Quelle sind nur `image:`-Zeilen in `docker-compose*.yml`, `FROM`-Zeilen in jedem
    `Dockerfile` und die `packageManager`-Felder der drei `package.json`.
  - `ghcr.io/nicograef/jotti-*` ist ausgenommen; die Tag-Variable ist dort Absicht.
  - Verglichen wird je Image-Name der Versionsanteil des Tags bis zum ersten `-`;
    `caddy:2.11.4-builder` und `caddy:2.11.4` gelten als gleich.
  - Zwei Versionen für einen Namen sind ein Fehler, außer die versionierte Allowlist
    nennt ihn mit Grund.
  - `nginx` steht darin (1.27 in Compose, 1.31 in den Dockerfiles); der Bump gehört der
    Dependency-Session.
  - `packageManager` muss in `frontend`, `website` und `e2e` wörtlich übereinstimmen.
  - `scripts/test-integration.sh` und `scripts/test-tse-live.sh` starten `postgres:17.8`.
  - `e2e/package.json` trägt denselben `packageManager`-Wert samt sha512-Hash.

- [ ] Der SECURITY-Absatz in `docker-compose.local.yml` widerspricht dem Dateikopf nicht
      mehr: Let's Encrypt bleibt der Primärpfad, die interne CA der Fallback, und die
      Warnung vor der Internet-Exposition bleibt stehen.
      Befund: docker-compose.local.yml:15-16
- [ ] `reverse-proxy/main.go` schreibt die Caddyfile an allen drei Stellen mit `0o600` wie
      `install.json`; ein Test prüft den Modus der erzeugten Datei. Sie trägt die
      acme-dns-Zugangsdaten. Befund: reverse-proxy/main.go:111,125,171-173

---

## Phase 12: Dokumentation gegen Code

**Depends on**: 3, 7, 8, 11

**Modell**: Sonnet 5

**Review-Tier**: Gate only, Review gebündelt am Phasenende

**Gate**: `bash scripts/check-ui-labels.sh`, `bash scripts/check-prose.sh`,
`bash scripts/check-links.sh`, `make website-check`

**Hinweis**: Das Gate entsteht im ersten Kriterium und wird erst durch den Sweep dieser
Phase grün; gelandet wird nur der grüne Endstand.

**Choke-Dateien**: `docs/handbuch.md`, `docs/language.md`, `docs/compliance.md`,
`README.md`, `TERMS.md`

### Context

- `frontend/src/admin/AdminSidebar.tsx` — die einzige Quelle der Menüpunkte („Kassentag",
  „Bondrucker", „Berichte & Export", „Finanzamt & TSE")
- `docs/leitfaden/**` — 65 in „…" zitierte Bedienelemente, davon 26 ohne Treffer in
  `frontend/src`; darunter Fremdzitate und über Zeilen umbrochene Zitate
- `docs/handbuch.md — §4.6 Bondruck-Retry`, `docs/language.md — Glossar Druckauftrag` —
  nennen drei Fehlversuche; `druckauftrag_repo/repo.go — MaxDruckversuche` steht auf 6,
  README nennt sechs
- `docs/handbuch.md — §5.4 Admin-Seiten` — nennt `DruckerConfigPage`;
  `frontend/src/routes.ts` registriert neun Admin-Routen mit `DruckstationConfigPage`
- `backend/api/druck/beleg/http/command_handler.go — KassenbelegHandler()` — vier gültige
  Body-Formen
- `.github/instructions/backend.instructions.md — Abschnitt JWT` — Rolle liegt nie im
  Context; `username` fehlt in der Claim-Liste
- `TERMS.md § 8 Abs. 2` — „Die Software setzt die deutsche Kassensicherungsverordnung
  (KassenSichV) um."

### What to build

Jede Doku-Aussage, die der Code widerlegt, wird korrigiert, und die riskanteste Klasse —
zitierte Bedienelemente — bekommt ein Gate. Alle Fakten stammen aus dem Stand nach den
Phasen 7, 8 und 11.

### Acceptance criteria

- [ ] `scripts/check-ui-labels.sh` sammelt die in `docs/leitfaden/**` in „…" zitierten
      Bedienelemente und schlägt fehl, sobald eines in `frontend/src` nicht vorkommt. Es
      zieht über Zeilenumbrüche verteilte Zitate vorher zusammen. Eine versionierte
      Allowlist-Datei nennt jedes Nicht-UI-Zitat mit seiner Quelle — mindestens Windows
      („Trotzdem ausführen", „Weitere Informationen"), Router („DNS-Rebind-Schutz",
      „Diese Domain(s) ausnehmen", „Automatisch", „Weitere Einstellungen"), GitHub
      („Source code (zip)"), ELSTER („Mitteilung über elektronische Aufzeichnungssysteme")
      und den UStAE-Wortlaut („Verkauf an eine Vielzahl nicht bekannter Personen").
- [ ] Die Bedienpfade der Anwenderdoku stimmen mit `AdminSidebar.tsx` überein:
      `tse-sonderfaelle.md` und `tse-einrichten.md` auf „Finanzamt & TSE" → „TSE
      einrichten", `datenaufbewahrung.md` auf „Berichte & Export" → „Archiv herunterladen
      (ZIP)", `veranstaltungstag.md` auf „Kassentag" und „Geld einlegen"/„Geld entnehmen",
      `installation.md` auf „Bondrucker", `betriebsarten.md` auf den realen
      Direktverkauf-Umschalter. Die übrigen offenen Zitate in `fehlersuche.md`,
      `haeufige-fragen.md`, `aktualisieren.md`, `belege-steuersaetze.md`,
      `self-hosting.md`, `finanzamt-anmelden.md` und `was-ist-jotti.md` sind korrigiert
      oder stehen mit Quelle in der Allowlist.
      Befund: docs/leitfaden/tse-sonderfaelle.md:47-53, docs/leitfaden/datenaufbewahrung.md:15-17,
      docs/leitfaden/veranstaltungstag.md:10-36, docs/leitfaden/installation.md:65-67
- [ ] `README.md` nennt für den Relay-Schnelltest `400` mit `{"code":"unauthorized"}` und
      den Menüpunkt „Bondrucker". Befund: README.md:78
- [ ] `docs/handbuch.md` beschreibt die neun registrierten Admin-Seiten mit
      `DruckstationConfigPage`, ergänzt die Service-Zeile um den Direktverkauf und kürzt
      die Tischdetail-Tabs auf „Bestellen, Kassieren, Historie".
      Befund: docs/handbuch.md:417-421, :481
- [ ] `docs/handbuch.md` §4.6 und `docs/compliance.md` nennen alle vier Body-Formen des
      Kassenbelegs (`tischId`+`zahlungId`, `tischId`+`stornierungId`, `verkaufId`,
      `verkaufId`+`stornierungId`) samt Stornobeleg-Familie.
      Befund: docs/handbuch.md:306, :306
- [ ] `docs/handbuch.md` und `docs/language.md` nennen sechs Fehlversuche bis
      `fehlgeschlagen` und das Backoff-Schema 5 s/15 s/30 s/60 s/180 s.
      Befund: backend/repository/druckauftrag_repo/repo.go:14-16, docs/handbuch.md:308
- [ ] `docs/handbuch.md` §3.11/§2.2 beschreibt den Ist-Stand ohne Stammdaten-Snapshot
      (Positions-Steuersätze eingefroren, Stammdaten beim Export gelesen), und §5.2
      beschreibt den realen Onboarding-Ablauf in vier Schritten (anlegen als `inactive`,
      „Neues Passwort festlegen", Admin aktiviert, regulärer Login).
      Befund: docs/handbuch.md:216, :364-367
- [ ] `docs/compliance.md` streicht die Abrechnungskreis-Zusage „Tisch 42-B" (§6.5 bleibt
      die einzige Aussage) und beschreibt den DSFinV-K-Versionsstring als Konstante
      `dsfinvk.Version`. Befund: docs/compliance.md:137, :296
- [ ] `.github/instructions/backend.instructions.md` nennt die echten Claims (`iss`, `iat`,
      `exp`, `sub`, `username`, `role`), die Context-Keys `UserIDKey`/`UserNameKey` und die
      Rolle aus dem Datensatz; `.github/instructions/database.instructions.md` nennt als
      kanonisches Schema alle `*.up.sql` in Reihenfolge.
      Befund: .github/instructions/backend.instructions.md:49-50,
      .github/instructions/database.instructions.md:6
- [ ] `docs/produktbeschreibung.md` beschränkt die Plattformaussage auf x86-64 (Resolved
      decisions) und beschreibt die Steuersätze auf Produktebene inklusive `kombi`
      (70/30). Befund: docs/produktbeschreibung.md:154,184, :138
- [ ] `docs/verfahrensdokumentation.md` listet die sechs Bounded Contexts aus
      `docs/handbuch.md` ohne den entfernten Vorgang „Ausgeben", und `docs/language.md`
      verliert den Abschnitt „Geplant"; der Geldtransit-Hinweis wandert an dessen
      Glossareintrag. Befund: docs/verfahrensdokumentation.md:46-52, docs/language.md:516-521
- [ ] `CLA.md` § 2 b) benennt die dem Autor gewährten Rechte als unwiderruflich, in
      Übereinstimmung mit Abschnitt 1 und `LICENSE:112-113`. Befund: CLA.md:36
- [ ] `docs/prds/prd-windows-nativ-ohne-docker.md` beschreibt den Docker-Weg im Präsens,
      ohne Verweise auf „Phase B" und eine Vorgänger-PRD, und die TLS-Aussage nennt Caddys
      interne CA und die DNS-01-Wildcard.
      Befund: docs/prds/prd-windows-nativ-ohne-docker.md:3-9,35,76-81
- [ ] `TERMS.md` § 8 Abs. 2 lautet „unterstützt die Anforderungen der deutschen
      Kassensicherungsverordnung (KassenSichV) technisch" (Eigentümer-Entscheidung, Option
      A). Das Versionsdatum von `TERMS.md` und `website/src/lib/anfrage-mailto.ts` bleibt
      unverändert.

---

## Phase 13: Minor-Sammelphase

**Depends on**: 3, 4, 9, 10, 11

**Modell**: Sonnet 5

**Review-Tier**: Gate only, Review gebündelt am Phasenende

**Gate**: `make verify`, `make test-e2e`, `bash scripts/check-e2e-assertions.sh`,
`bash scripts/check-domain-enums.sh`

**Choke-Dateien**: `Makefile`, `scripts/lib.sh`, `e2e/support/servicekraft.ts`

### Context

- `backend/api/kasse/kassenfuehrung/application/errors.go — ErrKasseAlreadyAbgeschlossen`
  — Sentinel ohne Erzeuger
- `backend/api/stammdaten/user/application/errors.go — ErrNoPassword,
ErrNoOnetimePassword` — zwei tote Sentinels
- `backend/domain/druckstation/druckstation.go` — trägt die Kategorien; eine Funktion
  `AlleKategorien()` gibt es noch nicht
- `backend/api/druck/station/http/handler.go — createSchema, updateSchema` — schreiben die
  Kategorie-Literale zweimal aus
- `backend/repository/produkt_repo/mock.go`, `kassenjournal_repo/mock.go` — Mocks
  antworten anders als die Produktion
- `e2e/support/viewport.ts — erwarteKeinenHorizontalenUeberlauf()` — `?? 0` macht eine
  fehlgeschlagene Messung grün
- `e2e/support/servicekraft.ts — Dateikommentar, waehleAlleVollAus()` — der Kommentar
  behauptet „ausschließlich zugängliche Selektoren", die Datei nutzt `data-slot`
- `e2e/tests/admin-kontrast-axe.spec.ts` — fünf `networkidle`-Wartezeiten
- `scripts/lib.sh` — `parse_semver`, Host-Preflight und `BACKUP_DIR`-Auflösung liegen in
  sechs Kopien

### What to build

Die verhaltensneutralen Minor-Befunde, gebündelt je Bereich: toter Code, Mocks mit
falschem Vertrag, weiche Test-Zusicherungen und doppelte Shell-Logik. Kein Kriterium ändert
Produktionsverhalten.

### Acceptance criteria

- [ ] Tote Sentinels und Konfigurationszweige sind entfernt:
      `ErrKasseAlreadyAbgeschlossen`, `ErrNoPassword` und `ErrNoOnetimePassword` in
      `stammdaten/user/application/errors.go`, sowie der unerreichbare `log.Fatalf`-Zweig
      in `config.go — parseEnvString()`.
- [ ] `domain/druckstation` erhält `AlleKategorien()` als einzige Quelle, aus der beide
      `OneOf`-Aufrufe in `api/druck/station/http/handler.go:73,117` lesen. Die
      Kategorie-Literale in `arbeitsbon_policy.go` sind durch die Domänenkonstanten
      ersetzt, und die Event-Konstruktoren in `domain/kasse` teilen einen
      `validateEventData`-Helfer. `scripts/check-domain-enums.sh` schlägt fehl, sobald ein
      Kategorie- oder Steuersatz-Literal außerhalb `backend/domain/**` steht.
- [ ] Die Repository-Mocks verhalten sich wie die Produktion: `produkt_repo`, `tisch_repo`
      und `user_repo` liefern `db.ErrNotFound` statt `nil`, `GetActiveProdukte` filtert wie
      der INNER JOIN, und `kassenjournal_repo/mock.go` vergibt neue IDs als
      `max(vorhandene)+1`.
- [ ] Toter Code in den Frontend-Hooks und im Reporting ist entfernt: `refetch` aus
      `useAktiveTische`, die `loading`-Prop samt Spinner-Zweig in `ReportingResults.tsx`,
      der unerreichbare Fallback in `TischHistorie.tsx — Details` und `unbezahlteMengen`
      in `Zahlung.tsx`.
- [ ] Die E2E-Suite scheitert bei fehlgeschlagener Messung.

  - `viewport.ts` gibt den Rohwert zurück und prüft `toBeGreaterThan(0)`.
  - `waehleAlleVollAus` sichert seine Nachbedingung mit dem Auswahl-Zähler zu.
  - `networkidle` ist durch Warten auf das gemessene Element ersetzt.
  - Der `tischSaldo`-Locator liegt einmal in `support/servicekraft.ts`.
  - `scripts/check-e2e-assertions.sh` verbietet `networkidle` und `?? 0` unter `e2e/`.
  - Das Skript läuft über `make check-repo`; kein ESLint-Setup, keine neue Abhängigkeit.

- [ ] Die Shell-Duplikate sind auf eine Quelle gezogen: `parse_semver`,
      `require_docker_stack`, `resolve_backup_dir`, `select_dump` und `decompress` liegen
      in `scripts/lib.sh`, und alle sechs Skripte lesen von dort.
- [ ] Toter Backend-Code außerhalb der Sentinels ist entfernt: `Area.Name` in
      `app/routes.go`, `zahlartReihenfolge` in `dsfinvk/mapper.go`, der Kombi-Zweig in
      `steuerMatrixLabel` und die zwei unerreichbaren Zweige im Umbuchungs-Kommentarbau.
- [ ] Doppelte Frontend-Konstanten sind entfernt: die Re-Exports in
      `KassensitzungPage.tsx`, die zweite `KATEGORIE_LABEL`-Definition und die dreifache
      Backend-Client-Instanz.
- [ ] `make check-tools` prüft zusätzlich `migrate` und `docker`, und die redundanten
      Makefile-Prerequisites (`clean`, `verify`) sind entfernt.
- [ ] Der Dateikommentar in `e2e/support/servicekraft.ts` beschreibt den Ist-Stand: die
      Datei nutzt überwiegend zugängliche Selektoren und für Zeile und Saldo die
      `data-slot`-Attribute. Befund: e2e/support/servicekraft.ts:4-6

---

## Phase 14: Entscheidungsphase für große Refactorings

**Depends on**: 13

**Modell**: Opus 5

**Review-Tier**: Gate + Sweep + Opus-Review

**Gate**: `bash scripts/check-prose.sh`, `bash scripts/check-links.sh`,
`make check-format`

**Choke-Dateien**: `docs/adrs/README.md`

### Context

- `frontend/src/admin/tse/TSEEinrichtungWizard.tsx` — 934 Zeilen, 14 Komponenten,
  Zugangsdaten über fünf Prop-Ebenen
- `backend/api/admin.go — RegisterAdminRoutes()`,
  `frontend/src/admin/products/Products.tsx`, `website/src/lib/live-demo.ts` — englische
  Bezeichner in deutscher Domäne
- `Makefile — go-Modul-Ziele` und `.github/workflows/ci.yml — Go-Modul-Jobs` — fünffach
  kopierte Pipeline; `docker-compose.release.yml` — vier Kopien des Service-Graphen
- `e2e/tests/kassenabschluss.mobile.spec.ts` — 120 s Klickarbeit für eine Zusicherung
- `docs/adrs/01_ausgabe-bestaetigen.md` — Vorbild für Format und Tonfall

### What to build

Je großem Refactoring eine Entscheidung mit Begründung: „jetzt", „v1.1" oder „nie". Ergebnis
ist ein ADR oder ein dokumentiertes No-Go, kein Code. Maßstab sind Korrektheit,
Einfachheit, Konsistenz und Produkt-Konservatismus.

### Acceptance criteria

- [ ] `docs/adrs/` enthält eine Entscheidung zur Aufteilung des TSE-Einrichtungs-Wizards
      (je Schritt eine Datei, Zugangsdaten über einen lokalen Context) mit Empfehlung und
      Begründung.
- [ ] `docs/adrs/` enthält eine Entscheidung zum Sprachschnitt der Bezeichner:
      Endpunkt-Verben in `api/admin.go`, `admin/products`, `Receipt`/`HistoryRow` und
      `live-demo.ts` — entweder Umbenennung als eigener Change oder die Ausnahme
      ausdrücklich in `docs/language.md`.
- [ ] `docs/adrs/` enthält eine Entscheidung zu den vervielfachten Pipelines: Go-Modul-Jobs
      in `Makefile` und CI als Matrix, und die vier Compose-Kopien als Basisdatei mit
      Overrides — inklusive der Risiken für Projektnamen und Volume-Identitäten.
- [ ] `docs/adrs/` enthält eine Entscheidung zur E2E-Seed-Variante mit ausgeglichenen
      Tischen für `kassenabschluss.mobile.spec.ts` statt der 120-s-Klickstrecke.
- [ ] `docs/adrs/README.md` listet die neuen ADRs, und keine der Entscheidungen hat Code
      geändert.

---

## Abdeckung

Alle 97 im Findings-Dokument als „bestätigt" geführten Befunde, in der Reihenfolge des
Dokuments. Mehrfach beschriebene Befunde zeigen auf dasselbe Kriterium.

| Befund (Datei:Zeilen)                                                       | Phase.Kriterium |
| --------------------------------------------------------------------------- | --------------- |
| backend/domain/kasse/bestellung.go:97-106                                   | 1.4             |
| backend/api/kasse/enrichment/enrichment.go:76-99                            | 1.3             |
| backend/repository/produkt_repo/batch.go:16-63                              | 1.3             |
| backend/api/stammdaten/user/http/command_handler.go:166-219                 | 1.5             |
| backend/api/kasse/kassenfuehrung/application/command.go:364-376             | 7.4             |
| backend/domain/user/user.go:202-228                                         | 8.1             |
| backend/domain/betreiber/betreiber.go:24-32                                 | 6.2             |
| backend/domain/kasse/tisch_session_events.go:118-128                        | 8.3             |
| backend/api/druck/bondruck/application/escpos/formatter.go:113-259          | 1.2             |
| backend/api/druck/bondruck/application/arbeitsbon_policy.go:15-21           | 7.5             |
| backend/api/druck/bondruck/application/arbeitsbon_policy_test.go:18-47      | 7.5             |
| backend/api/fiskal/dsfinvk/mapper.go:535-544                                | 8.4             |
| backend/api/fiskal/dsfinvk/table.go:13-34                                   | 8.5             |
| backend/api/middleware/middleware.go:269-304                                | 8.2             |
| backend/api/stammdaten/tisch/http/command_handler.go:77-84                  | 5.2             |
| backend/api/stammdaten/produkt/application/command.go:76-80                 | 5.3             |
| backend/api/stammdaten/betreiber/http/command_handler.go:31-51              | 6.3             |
| backend/repository/tisch_repo/repo.go:11-146                                | 9.3             |
| backend/repository/user_repo/types.go:19-62                                 | 9.1             |
| backend/repository/produkt_repo/repo.go:38-92                               | 9.2             |
| backend/repository/kassenjournal_repo/repo.go:264-267                       | 8.6             |
| backend/repository/kassenjournal_repo/repo_test.go:1583-1592                | 8.6             |
| backend/repository/druckauftrag_repo/repo.go:14-16                          | 12.6            |
| docs/handbuch.md:306 (Backend)                                              | 12.5            |
| .github/instructions/backend.instructions.md:49-50                          | 12.9            |
| frontend/package.json:14                                                    | 2.1             |
| frontend/package.json:21-43                                                 | 10.10           |
| frontend/src/components/ui/sonner.tsx:1-8                                   | 10.4            |
| frontend/src/components/ui/dialog.tsx:93 (+ sheet, sidebar, spinner)        | 4.4             |
| frontend/src/test/input-otp.ts:1-16                                         | 10.5            |
| frontend/src/service/table/hooks.ts:56-90                                   | 10.1            |
| frontend/src/service/components/table/BestellungAbschluss.tsx:50-81         | 1.6             |
| frontend/src/service/components/table/BestellungAbschluss.tsx:129-143       | 10.7            |
| frontend/src/service/components/table/ZahlungAbschluss.tsx:123-170          | 10.7            |
| frontend/src/service/components/ServiceDock.tsx:4-7 (+ vier Dateien)        | 3.4             |
| frontend/src/service/product/Produkt.ts:24-55                               | 10.6            |
| frontend/src/admin/reporting/ReportingResults.tsx:35-39                     | 10.2            |
| frontend/src/admin/kasse/GeldtransitDialog.tsx:50-70                        | 1.6             |
| frontend/src/admin/finanzamt/LaeuftAllesSection.tsx:65-110                  | 10.3            |
| frontend/src/admin/reporting/UebersichtStatusZeile.tsx:8-10 (+ elf Dateien) | 3.7             |
| frontend/src/admin/tse/TSEEinrichtungWizard.test.tsx:77-167                 | 10.8            |
| frontend/src/admin/users/UserRow.tsx:69-84                                  | 1.5             |
| frontend/src/admin/components/AdminPageHeader.tsx:5-10                      | 3.7             |
| website/src/styles/brand.css:2-3                                            | 3.5             |
| website/src/styles/brand.css:11-17                                          | 3.5             |
| website/src/layouts/Landing.astro:23-24                                     | 3.5             |
| e2e/website/csp-check.mjs:1-12                                              | 3.5             |
| e2e/website/csp-server.mjs:7-9                                              | 3.5             |
| e2e/website/screenshots.mjs:12-15                                           | 3.5             |
| packaging/windows/jotti-restore.cmd:31-37                                   | 1.8             |
| packaging/windows/jotti-restore.cmd:36-37, jotti-repair.cmd:15,46           | 11.1            |
| packaging/windows/KURZANLEITUNG.md:83-91                                    | 11.2            |
| windows/starter/backup.go:67,69 (+ fünf Dateien)                            | 4.3             |
| reverse-proxy/caddyfile.go:201-221                                          | 11.3            |
| reverse-proxy/caddyfile.go:5-8; nginx.rocks.conf:142-143                    | 11.4            |
| reverse-proxy/main.go:78-86                                                 | 11.5            |
| reverse-proxy/main.go:134-158                                               | 11.6            |
| Makefile:78-79                                                              | 2.2             |
| scripts/prod-update.sh:111 (+ vier Stellen)                                 | 3.6             |
| scripts/prod-backup.sh:66-94                                                | 1.7             |
| scripts/prod-restore.sh:99-116                                              | 11.7            |
| scripts/test-integration.sh:34; scripts/test-tse-live.sh:51                 | 11.8            |
| docker-compose.local.yml:15-16                                              | 11.9            |
| docs/leitfaden/tse-einrichten.md:26-32                                      | 1.1             |
| docs/leitfaden/tse-sonderfaelle.md:47-53                                    | 12.2            |
| docs/leitfaden/datenaufbewahrung.md:15-17                                   | 12.2            |
| docs/leitfaden/veranstaltungstag.md:10-36                                   | 12.2            |
| docs/leitfaden/installation.md:65-67                                        | 12.2            |
| README.md:78                                                                | 12.3            |
| CLA.md:36                                                                   | 12.12           |
| docs/handbuch.md:481                                                        | 12.4            |
| docs/handbuch.md:417-421                                                    | 12.4            |
| docs/handbuch.md:306 (Dokumentation)                                        | 12.5            |
| docs/compliance.md:137                                                      | 12.8            |
| docs/compliance.md:296                                                      | 12.8            |
| .github/instructions/database.instructions.md:6                             | 12.9            |
| docs/produktbeschreibung.md:154,184                                         | 12.10           |
| docs/produktbeschreibung.md:138                                             | 12.10           |
| docs/prds/prd-windows-nativ-ohne-docker.md:3-9,35,76-81                     | 12.13           |
| docs/verfahrensdokumentation.md:46-52                                       | 12.11           |
| docs/language.md:516-521                                                    | 12.11           |
| backend/api/druck/bondruck/application/escpos/formatter.go:113-305          | 1.2             |
| backend/api/druck/bondruck/application/arbeitsbon_policy.go:18-21           | 7.5             |
| backend/api/kasse/tischgeschaeft/application/command.go:135-147             | 5.4             |
| backend/api/kasse/kassenfuehrung/application/command.go:260-261             | 8.7             |
| backend/api/kasse/kassenfuehrung/application/query.go:14-24                 | 7.2             |
| backend/api/kasse/enrichment/enrichment.go:72-99                            | 1.3             |
| backend/api/fiskal/setup/application/command.go:17-44                       | 7.3             |
| backend/api/helper/http.go:23-25                                            | 5.5             |
| backend/api/stammdaten/produkt/application/command.go:76-80 (Cross-Layer)   | 5.3             |
| backend/api/stammdaten/tisch/http/command_handler.go:78-84                  | 5.2             |
| backend/api/stammdaten/user/http/command_handler.go:166-183                 | 1.5             |
| backend/domain/user/user.go:202-228 (Cross-Layer)                           | 8.1             |
| docs/handbuch.md:308                                                        | 12.6            |
| docs/handbuch.md:216                                                        | 12.7            |
| docs/handbuch.md:364-367                                                    | 12.7            |
| packaging/windows/KURZANLEITUNG.md:85-91                                    | 11.2            |

Maßgeblich sind die 506 Einträge im Fließtext des Findings-Dokuments: 97 tragen
`Status: bestätigt`, 409 tragen `Status: unverifiziert (minor)`. Alle 97 sind je genau
einem Kriterium zugeordnet, 0 bleiben offen.

Zur Severity: im Fließtext trägt kein Eintrag `blocker`, und 25 der 97 bestätigten tragen
`major`. Die Abschnittsköpfe summieren 1 Blocker und 106 Major; die Kopftabelle „Zahlen"
nennt 698 verbleibende Befunde mit 127 Major. Beide Zahlen sind mit dem Fließtext nicht
abgeglichen und dienen diesem Plan nicht als Maßstab.

Zusätzlich nennt der Plan einzelne unverifizierte Minor-Befunde dort, wo sie eine
Defektklasse offen ließen — unter anderem `DeleteVariante` (5.8), die `log.Error()`-Kette
(5.9), das ELSTER-Datum (8.8), die Caddyfile-Rechte (11.10) und der
Selektoren-Kommentar (13.10).

### Defektklassen und ihr Gate

| Defektklasse                                           | Gate                                                          | Phase     |
| ------------------------------------------------------ | ------------------------------------------------------------- | --------- |
| Regel-18-Prosa, tote Verweise, opake Audit-IDs         | `scripts/check-prose.sh`, `scripts/check-links.sh`            | 3         |
| Sprach- und Zeichenkonventionen                        | `scripts/check-language.sh`, `.gitattributes`                 | 4         |
| Gates, die nicht rot werden können                     | `make lint` = Merge-Gate, `make check-format`                 | 2         |
| Unit-Tests werden nie gelintet                         | `scripts/check-build-tags.sh`, `--build-tags=unit`            | 2         |
| Sentinels ohne HTTP-Mapping                            | `error_mapping_contract_test.go`                              | 5         |
| `byCode`-Overrides formulieren um                      | ungegated — Sweep 5.10; kein Lint ohne neue Abhängigkeit      | 5         |
| Schemas ohne obere Schranke (zog)                      | `schema_grenzen_test.go`                                      | 6         |
| Zod ohne Trim gegenüber zog                            | ungegated — Sweep 6.4 plus Vitest der geteilten Quelle        | 6         |
| Zwischenstatus `wird_abgeschlossen` ignoriert          | `forbidigo` auf `KassensitzungenRepo.GetOffeneKassensitzung(` | 7         |
| Event-JSON gegen ungetaggte Domain-Structs             | `event_decode_contract_test.go`                               | 7         |
| Zeitstempel ohne definierte Zone                       | `scripts/check-timezone.sh` plus Tests (1.2, 8.8, 8.9)        | 1, 8      |
| ID-Paare ohne Zugehörigkeitsprüfung                    | Integrationstests „fremde Variante" (1.3, 5.8)                | 1, 5      |
| Idempotenz-Schlüssel am UI-Zustand                     | Vitest je Dialog (1.6)                                        | 1         |
| React-Query verwirft `isError`                         | Vitest „Fehlladung zeigt keine Nulldaten" (10.1)              | 10        |
| Ein Vertrag in mehreren Frontend-Fassungen             | ein Schema-Modul je Begriff (10.6)                            | 10        |
| Geteilte Schichten importieren aufwärts                | `no-restricted-imports` (Kernregel)                           | 10        |
| Duplizierte Wahrheit ohne Kopplung (CSP)               | Gleichheitstest in `package main` (11.4)                      | 11        |
| Werkzeug- und Versionsdrift                            | `scripts/check-pins.sh`, `setup-dev-tools.sh`                 | 2, 11     |
| Doku-Drift gegen Code (zitierte Bedienelemente)        | `scripts/check-ui-labels.sh`                                  | 12        |
| Secret-tragende Dateien mit Standardrechten            | Modus-Prüfung in `prod-backup.sh` (1.7) und Proxy (11.10)     | 1, 11     |
| Weiche Test-Zusicherungen (`?? 0`, `networkidle`)      | `scripts/check-e2e-assertions.sh`                             | 13        |
| Self-referential Test-Fixtures                         | Fixtures über Domänen-Konstruktoren (7.5)                     | 7         |
| Repository-Fehler nicht normalisiert                   | `db.Error()` in `reporting_repo` (9.4)                        | 9         |
| Domänenwissen in Repository und HTTP dupliziert        | `scripts/check-domain-enums.sh`, Schemas aus `domain/`        | 6, 13     |
| `log.Error()` ohne `.Err(err)`                         | `log_error_contract_test.go`                                  | 5         |
| Kommentare behaupten Verhalten, das der Code nicht hat | je Zusage ein Test oder eine Ist-Aussage (3.5, 11.4, 13.10)   | 3, 11, 13 |
| DRY in Ops-Shell und E2E-Selektoren                    | `scripts/lib.sh` (13.6), `support/`-Selektoren (13.5)         | 13        |
| Große Refactorings (Aufwand L)                         | ADR oder dokumentiertes No-Go                                 | 14        |

## Nicht übernommene Befunde

### Einzelne Befunde

| Befund                                                                  | Grund                                                                                                  |
| ----------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------ |
| `docs/plans/review-externe-prs.md:40-49`                                | Datei ist in dieser Session gelöscht (Commit 88d047b)                                                  |
| `docs/plans/guide-manuelle-qa-v1.0.0.md:14,59,71`                       | Gehört zur Release-Phase E (Plan 2 Phase 11)                                                           |
| `docs/plans/plan-orchestrierung.md:17-26`                               | Plan-Datei ist transient; die Lead-Session fährt bereits auf Opus 5                                    |
| `.github/workflows/ci.yml:406-417`, `database/migrations/README.md:44`  | `PREVIOUS_VERSION` wird nach dem Tag gesetzt — Phase E                                                 |
| `e2e/package.json:17-22` (TypeScript 7.0.2)                             | Versions-Bump; die Dependency-Session übernimmt ihn                                                    |
| `website/pnpm-workspace.yaml:4-5` (`minimumReleaseAgeExclude`)          | Hängt an der Astro-Version; gehört zur Dependency-Session                                              |
| `frontend/eslint.config.js:17` (`src/components/ui` ausgenommen)        | shadcn-Fremdcode bleibt ausgenommen (Resolved decisions)                                               |
| `frontend/src/lib/Backend.ts:134-138` (Redirect im HTTP-Client)         | Verhaltensändernd; gehört in einen eigenen, geplanten Change                                           |
| `backend/api/middleware/middleware.go:199-203` (POST-only-Ausnahme)     | Regel 1 in `AGENTS.md` bleibt unangetastet; `/health` ist im Handbuch dokumentiert                     |
| `backend/sqlc/queries/produkte.sql:10-96` (View für die Varianten-JSON) | Bräuchte eine Migration ohne fachlichen Anlass — Freeze-Disziplin                                      |
| `docs/adrs/05_spektral-branding-website.md:4-6`                         | ADRs sind von Regel 18 ausgenommen und werden nie umgeschrieben                                        |
| `scripts/generate-spektral-logos.py`                                    | Einmal-Generator ohne Konsument; Löschen oder Behalten ist eine Eigentümerfrage ohne Release-Bezug     |
| `website/src/components/Hero.astro:42` („Beta 1.0" dreimal hartkodiert) | Release-Statustext; gehört zur Release-Phase E, nicht zu einem Fix-Plan                                |
| `website/astro.config.mjs:54-142` (Sidebar-Slugs vs. `publishedDocs`)   | Ein Vergleichstest bräuchte einen Astro-Testlauf; Nutzen deckt den Aufbau nicht                        |
| `website/src/lib/anfrage-mailto.ts:77-91` (TERMS-Datum hartkodiert)     | Option A ändert nur den Wortlaut von § 8; das Versionsdatum bleibt, der Eigentümer prüft es beim Merge |
| `website/nginx.conf:33-43` („content-hashed")                           | Kommentar-Zusage ohne Verhaltensfehler; Caching bleibt korrekt                                         |
| `e2e/helpers/fehlerpfade.ts:3-6` (POST-Filter-Zusage)                   | Der Helfer filtert absichtlich nach Pfad; der Kommentar wird beim nächsten Change gerade gezogen       |

### Ganze Klassen

| Klasse (Beispiel)                                                                                                                                                                                 | Anzahl       | Grund                                                                                                             |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------ | ----------------------------------------------------------------------------------------------------------------- |
| Reine Stilpräferenzen ohne Gate (`command_handler.go:44-193` Value-Receiver, `jwt.go:31-53` unbenannte Ergebnisse, `middleware.go:194-195` Rückgabeformen)                                        | ~45          | Ändern nichts an Korrektheit oder Lesbarkeit für einen Reviewer; erzeugen nur Diff                                |
| DRY-Refactorings in Testdateien (`fiskaly_client_test.go:40-124`, `handler_test.go:40-264`, `formatter_test.go:177-640`)                                                                          | ~40          | Kein Deckungsgewinn; Umbau riskiert Zusicherungen, die heute korrekt sind                                         |
| Neue Tests für ungetestete Helfer (`export.go:124-155`, `tisch.go:70-107`, `ProductList.tsx:24-52`)                                                                                               | ~25          | Testschulden ohne Befund; werden mit dem nächsten Change an der Stelle nachgezogen                                |
| Umbenennungen von Bezeichnern (`Receipt`→`Beleg`, `HistoryRow`, `settleAlleOffenenTische`, `live-demo.ts`)                                                                                        | ~20          | Sammeln sich in der Sprachschnitt-Entscheidung, Phase 14                                                          |
| Große Refactorings (Aufwand L: `TSEEinrichtungWizard.tsx`, `admin.go:34-71`, `Makefile:281-291`, `docker-compose.release.yml:31-172`, `kassenabschluss.mobile.spec.ts:32-53`)                     | 6            | Ergebnis ist eine Entscheidung, kein Code — Phase 14                                                              |
| Korrektheits-Minors ohne Release-Bezug (`externalize-inline-scripts.ts:43-48`, `probe.go:53-67`, `resolve.go:58-62`, `signaturstatus.go:59-64`, `UserDropdown.tsx:42-45`, `adminmarker.go:49-53`) | Rest der 409 | Ungeprüfte Einzelbefunde ohne Bezug zu einer bestätigten Defektklasse; sie werden nach v1.0.0 einzeln verifiziert |
| Verworfene Befunde des Audits                                                                                                                                                                     | 49           | Vom Audit geprüft und fallengelassen; werden nie erneut aufgeworfen                                               |

Die fünf oberen Klassen decken rund 136 der 409 unverifizierten Minor-Befunde ab. Die
übrigen fallen unter die Klasse „Korrektheits-Minors ohne Release-Bezug": ungeprüfte
Einzelbefunde, die dieser Plan bewusst offen lässt. Keiner davon berührt einen bestätigten
Befund oder eine gegatete Defektklasse.
