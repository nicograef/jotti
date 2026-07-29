# Plan: Release v0.17.2 — reduzierter, regressionsfreier Stand für das laufende Fest

> Quell-PRD: n/a (aus Praxis-Meldung, Multi-Experten-Review und Commit-Einzelprüfung)
> Grundlage: `docs/plans/review-v0.17.2.md` (26 bestätigte Befunde) und fünf
> Einzelprüfungen der Grenzfall-Commits mit echten Probe-Cherry-Picks.

## Ziel

`main` wird ab Tag `v0.17.1` neu aufgebaut und enthält nur noch Änderungen, die ein
**belegtes** Feldproblem beheben und deren Sicherheit an der Zielbasis nachgewiesen ist.
Der Stand wird als `v0.17.2` getaggt und mitten im laufenden Vereinsfest auf die
Produktivinstanz eingespielt.

Maßstab ist Regressionsfreiheit, nicht Vollständigkeit der Verbesserung. v0.17.1 läuft
seit mehreren Tagen stabil mit 5–8 Servicekräften; jede Änderung muss sich gegen die
Frage verteidigen „was ging vorher und geht jetzt schlechter".

## Ausgangslage

Zwischen `v0.17.1` (`3d334de2`) und `4040cd21` liegen 33 Commits in drei fachlichen
Blöcken. Drei Feldprobleme vom ersten Festtag waren der Anlass:

- **(a)** Die Tischübersicht blieb blockiert, weil gelöschte Tische in den Favoriten standen.
- **(b)** 1–3 Bons wurden nicht gedruckt.
- **(c)** Der Kassenabrechner („Rechner") konnte Stornierungen keiner Servicekraft zuordnen.

Zwei Erkenntnisse haben den Zuschnitt gedreht:

1. **Tag 2 verlief ohne jede Druckerstörung.** Der Nutzen des Relay-Umbaus ist damit
   unbelegt, während das Review für genau diesen Umbau Verschlechterungen bestätigt hat.
2. **Der Robustheits-Block war nur zum kleinsten Teil durch (a) motiviert.** Sein eigener
   Plan (`docs/plans/plan-client-server-robustheit.md`) nennt als Auslöser ausschließlich
   die blockierte Tischübersicht und beschreibt alles Weitere als Härtung „der
   darunterliegenden Schwächen" — also Arbeit auf Vorrat.

## Architekturentscheidungen

Durchgängig gültig für alle Phasen:

- **Historie**: `main` wird per Force-Push durch den neu aufgebauten Branch ersetzt. Das
  überstimmt bewusst die Regel „kein `--force` push" aus `AGENTS.md`. Zulässig **nur**,
  weil vorher ein Archiv-Ref auf den bisherigen Stand gepusht wird — danach ist nichts
  unerreichbar und der Push benennt lediglich um, was `main` heißt.
- **Archiv-Ref**: `archiv/main-vor-v0.17.2` auf dem bisherigen `main`. Er ist die
  dauerhafte Heimat des Relay-Umbaus, der Idempotenz-Maschinerie und der
  Transporthärtung. (Umgesetzt auf `e0960b68` statt auf dem beim Schreiben dieses
  Plans aktuellen `4040cd21` — `e0960b68` kam danach hinzu und hat `4040cd21` als
  Vorfahr.)
- **Aufbau ausschließlich per `git cherry-pick`**, nie `git apply`. Zwei der gewählten
  Commits scheitern an `git apply --check`, während der Drei-Wege-Merge sauber durchläuft.
- **Keine handkomponierten Auszüge aus großen Commits.** Aufgenommen wird ein Commit
  ganz oder gar nicht. Einzige Ausnahmen sind zwei benannte Einzeiler (siehe Phase 2),
  die wörtlich aus einem verworfenen Commit übernommen und nicht neu erfunden werden.
- **Schema**: Der neue Stand bringt genau **eine** Migration mit,
  `database/migrations/06_favoriten_cleanup.up.sql`. Die Produktivinstanz steht auf
  Version 5, die Migration ist additiv und räumt ausschließlich verwaiste Zeilen in
  `tisch_favoriten` ab. Kein Event-Format, kein Eingriff ins Kassenjournal.
- **Ubiquitous Language, POST-only, Cent-Beträge, Freeze-Disziplin** gelten unverändert
  (`AGENTS.md`).

## Inventar

### Aufgenommen — 14 Commits

**Feldproblem (c), Storno-Zuordnung — kompletter Block, 10 Commits**

`ec371c44` · `0b0f3083` · `a937a5fe` · `6a2d7ca7` · `f5a3d0b6` · `0b6bc45b` · `7ad23253`
· `c60156f8` · `e9a0bbf6` · `a00f91e5`

Kernstellen: `backend/sqlc/queries/reporting.sql`, `backend/domain/reporting/reporting.go`,
`backend/api/reporting/application/query.go — LiveReporting()`,
`backend/repository/reporting_repo`, `frontend/src/admin/reporting/`,
`frontend/src/service/components/EigeneUebersicht.tsx`.

Nachgewiesen: Probe-Cherry-Pick aller 10 Commits auf `v0.17.1` konfliktfrei; Tree-Vergleich
gegen `a00f91e5` zeigt alle 28 vom Block besessenen Go-/TS-Dateien byte-identisch. Reiner
Lesepfad — keine Migration, kein Schema, keine Event-Struct berührt.

**Feldproblem (a), gelöschte Tische in Favoriten — 2 Commits**

`266c1ac8` · `631f4eeb`

Kernstellen: `backend/repository/tisch_repo/repo.go — DeleteTableMitFavoriten()`,
`backend/api/stammdaten/tisch/application/command.go`,
`backend/api/kasse/tischgeschaeft/application/query.go — GetMeineTischeState()`,
`backend/repository/favorit_repo`, `backend/sqlc/queries/favoriten.sql`,
`database/migrations/06_favoriten_cleanup.up.sql`.

**DSFinV-K-Export und Fehlerformat — 2 Commits**

`81f0b6d3` · `e020a467`

Kernstellen: `backend/api/middleware/middleware.go — responseWriter.Unwrap()`,
`backend/api/fiskal/export/http/handler.go`.

Nachgewiesen mit echtem `http.Server`: Ohne `responseWriter.Unwrap()` liefert jedes
`SetWriteDeadline` „feature not supported", der Export bricht mit `gelesen=0 bytes, err=EOF`
ab. Mit Unwrap kommen alle 65536 Bytes an. `Server.WriteTimeout` läuft ab dem Lesen des
Request-Headers und deckt den Archivbau mit ab — ein Export über 10 s scheitert heute also
**vollständig**, nicht abgeschnitten. Aufbewahrungsrechtlich relevant.

### Verworfen — 19 Commits

| Block | Commits | Grund |
|---|---|---|
| Print-Relay | `f75bc9ab` `7c2c37eb` `80a6a819` `b500915c` `95324ef9` `19a542d6` `481b6169` | Tag 2 störungsfrei → Nutzen unbelegt. Review-belegte Verschlechterungen: ein stumm sterbender Drucker bekommt bis zu sechs Bons quittiert statt einem; Wiederanlauf nach leerer Papierrolle ~172 s statt ~2 s |
| Idempotenz | `c6d9f840` `187fbfea` `73aee693` `9ef41ab6` | Kein beobachtetes Problem. Macht `vorgangId` zum Pflichtfeld und bricht jedes nicht neu geladene Handy beim Kassieren; legt mitten im Fest eine neue Tabelle an |
| Ladefehler-Anzeige | `1be7cd29` | Isolierter Pick ergibt den **fehlerhaften** Stand: die Korrektur `isError` → `isLoadingError` steckt nur in `9ef41ab6`. Folge auf v0.17.1-Basis: ein gescheiterter Hintergrund-Refetch reißt die ganze Tischauswahl weg, inklusive aller Einstiege in einen Tisch |
| Transporthärtung | `d85cf2ae` `c2d0b3c6` | Das globale `staleTime: 30_000` erzeugt ohne Post-Buchungs-Invalidierung eine neue Doppelkassier-Einladung: ein bezahlter Tisch steht bis zu 30 s weiter unter „Noch offen" mit altem Betrag. Genau das beschreibt Commit `18cf21d5`, und `9ef41ab6` hat die Schwelle deshalb zurückgebaut — beide sind nicht dabei |
| Übersichts-Invalidierung | `18cf21d5` | Auf v0.17.1-Basis nachweislich ein No-Op: `refetchType: 'active'` feuert nicht auf ausgehängte Queries, und die drei Übersichts-Queries sind beim Buchen immer ausgehängt. Dazu harte TS-Abhängigkeit auf `1be7cd29` |
| Doku und Mechanik der verworfenen Blöcke | `40aac5f0` `5f144868` `bcbe4919` `4040cd21` | Beschreiben Code, der nicht mehr im Baum steht. Sie leben im Archiv-Ref weiter |

### Bewusst nicht umgesetzt

- Die **18 verhaltenserhaltenden Cleanup-Punkte** aus `docs/plans/review-v0.17.2.md`.
  Umbau-Risiko ohne Nutzen für den Festbetrieb.
- Die **Relay-Korrekturen** (Quittungs-Fallback, gedeckelte Backoff-Flanke). Der Block ist
  draußen; zuerst wird die Ursache geklärt — `docs/plans/plan-bondruck-ursachenklaerung.md`.
- Der **Langläufer-Schutz für die TSE-Einrichtung** aus `9ef41ab6` (`context.WithoutCancel`,
  `einrichtungLaeuft`-Schloss, 2-Minuten-Schreibfristen). Fachlich richtig, aber er schützt
  ausschließlich die TSE-Ersteinrichtung — beim Verein ist die TSE eingerichtet. Er wäre
  zudem ein handkomponierter Commit über rund zehn Dateien aus einem 122-Dateien-Commit.
- Die zwei brauchbaren Teile aus `d85cf2ae` (Retry-Politik auf Lese-Queries,
  Korrelations-ID im Fehler-Toast). Klein und harmlos, aber nicht nötig — und nur per
  Handauszug zu haben.

## Resolved decisions

- **Schnitt**: Neuaufbau ab `v0.17.1`, Force-Push auf `main`, vorher Archiv-Ref.
- **Storno-Block**: aufgenommen. Belegtes Feldproblem, reiner Lesepfad, Pickbarkeit empirisch bewiesen.
- **Grenzfälle**: einzeln geprüft statt pauschal entschieden. Ergebnis: 2 von 6 aufgenommen.
- **Relay nach dem Fest**: erst Ursachenklärung, keine Umsetzungszusage.
- **Client-Reload**: wird für dieses Update von Hand angesagt und durchgesetzt. Der
  automatische Versions-Handshake kommt nach dem Fest —
  `docs/plans/plan-v0.17.3.md`.
- **Nachgelagerte Arbeit steht in eigenen Plänen**, nicht in diesem. Dieser Plan endet
  mit dem Release; siehe „Nach dem Fest" am Schluss.

## Risiken

- **Der Force-Push ist der einzige Schritt ohne Rückweg im Repo.** Er ist erst zulässig,
  wenn `git ls-remote --tags origin` den Archiv-Ref bestätigt.
- **PR #101** (dependabot, 27 Frontend-npm-Bumps) zielt auf `main` und muss nach dem
  Rewrite neu bewertet werden. Nicht vor dem Fest mergen.
- **Der gepickte Stand ist eine Kombination, die so nie gelaufen ist.** Deshalb hat die
  Verifikation eine eigene Phase und läuft vollständig, nicht stichprobenhaft.
- **Rollback nach dem Einspielen ist unsauber**: `jotti-start.exe` verweigert Downgrades,
  und die Migration ist forward-only. Der Rückweg ist `jotti-restore.cmd` aus dem
  automatischen Backup, nicht das Zurückkopieren des alten ZIPs.
- **Nummernkollision bei späterer Rückholung**: `06_druckauftrag_backoff_warteschlange.up.sql`
  aus dem Relay-Block belegt ebenfalls die 06. Wird der Relay-Block je zurückgeholt, muss
  seine Migration auf 07 rücken — golang-migrate lehnt sonst die ganze Quelle mit
  „duplicate migration file" ab, und Git meldet dabei nichts, weil die Dateinamen
  verschieden sind.

---

## Phase 1: Neuen Stand aufbauen

### Context

- `v0.17.1` (`3d334de2`) — Basis, entspricht dem produktiv laufenden Stand.
- `4040cd21` — bisheriger `main`, wird archiviert.
- `backend/api/fiskal/tse_live/tse_live_suite_test.go` — trägt in `631f4eeb` sieben
  `uuid.NewString()`-Argumente, deren Signaturen aus dem verworfenen `c6d9f840` stammen.
- `frontend/src/service/components/direktverkauf/DirektverkaufHistorie.test.tsx` — trägt in
  `631f4eeb` eine `vorgangId`-Assertion aus dem verworfenen `187fbfea`.

### What to build

Archiv-Ref setzen und pushen, danach einen Branch `release/v0.17.2` von `v0.17.1` abzweigen
und die 14 gewählten Commits in dieser Reihenfolge cherry-picken:

1. Storno-Block, 10 Commits in Historienreihenfolge
2. `266c1ac8`, dann `631f4eeb`
3. `81f0b6d3`, dann `e020a467`

Die Reihenfolge ist nicht beliebig. `e020a467` muss **nach** `631f4eeb` laufen: sein dritter
Hunk löscht `favorit_repo.Repository.RemoveByTisch`, das durch `631f4eeb` seinen letzten
Produktionsaufrufer verliert. In dieser Reihenfolge räumt sich der tote Code selbst ab.
`e020a467` muss **nach** `81f0b6d3` laufen, weil er dessen Testdatei erweitert.

Beim Pick von `631f4eeb` die zwei genannten Dateien auf den Stand von `v0.17.1`
zurücksetzen — sie sind im Commit-Text selbst als sachfremde Mitnahme deklariert und ziehen
sonst den verworfenen Block herein.

Die Mock-Methode in `backend/repository/favorit_repo/mock.go` bleibt bestehen; sie ist
Cleanup-Funktion in den Tisch-Command-Tests und hat mit der gelöschten Repository-Methode
nichts zu tun.

### Acceptance criteria

- [x] `git ls-remote --tags origin` listet `archiv/main-vor-v0.17.2` auf `4040cd21`
      — **abweichend umgesetzt:** der Tag zeigt auf `e0960b68`, den tatsächlichen
      `main` zum Zeitpunkt des Umbaus. Der Plan wurde geschrieben, bevor
      `e0960b68` (`docs: add plans for fix releases`) existierte; `4040cd21` ist
      dessen Vorfahr und bleibt damit erreichbar. Das Archiv ist so vollständiger,
      nicht lückenhafter.
- [x] Alle 14 Cherry-Picks laufen ohne ungelöste Konflikte durch — dazu `e0960b68`
      als 15. Pick, weil er reine Doku ist und diesen Plan selbst mitbringt
- [x] `database/migrations/` enthält genau `01`–`06`, mit `06_favoriten_cleanup.up.sql`
- [x] Weder `vorgang_idempotenz` noch `windows/relay`-Änderungen noch `OfflineBanner`
      sind im Baum (`git diff v0.17.1..HEAD --stat` prüfen)
- [x] `favorit_repo.Repository.RemoveByTisch` existiert nicht mehr, die gleichnamige
      Mock-Methode schon
- [x] `go build ./...` und `tsc -b` laufen fehlerfrei

---

## Phase 2: Nacharbeit auf dem neuen Stand

### Context

- `frontend/src/lib/errorMessages.ts` — Fehlercode-Zuordnung; `rate_limited` ist nach
  Phase 1 unzugeordnet und fällt auf eine generische Meldung zurück.
- `.github/workflows/ci.yml` — Job `upgrade-path`, Variable `PREVIOUS_VERSION`, steht auf
  `v0.14.0`.
- `database/migrations/README.md` — Regel 3 behauptet, eine in `BEGIN/COMMIT` geklammerte
  Migration hinterlasse bei Fehlschlag keinen `dirty`-Zustand.
- `docs/plans/plan-storno-zuordnung-servicekraft.md`,
  `docs/prds/prd-storno-zuordnung-servicekraft.md` — kommen mit dem Storno-Pick mit.

### What to build

Vier kleine, klar abgegrenzte Nachträge:

1. **`rate_limited` zuordnen.** Der Map-Eintrag wird **wörtlich** aus dem verworfenen
   `d85cf2ae` übernommen, nicht neu formuliert. Ohne ihn liefert `81f0b6d3` nicht die
   Meldung, die seine Commit-Message verspricht.
2. **`PREVIOUS_VERSION` auf `v0.17.1` heben**, an beiden Stellen (Workflow und
   Migrations-README). Der `upgrade-path`-Job muss genau den Pfad testen, der beim Verein
   gefahren wird.
3. **Regel 3 im Migrations-README korrigieren.** golang-migrate schreibt
   `SetVersion(target, dirty=true)` in einer eigenen Transaktion **vor** dem Lauf und
   committet. Scheitert eine Migration, rollt zwar das Schema sauber zurück,
   `schema_migrations` steht aber committet auf `dirty=true`, jeder weitere `migrate up`
   bricht mit `ErrDirty` ab und der Backend-Container startet gar nicht. Die Regel muss den
   tatsächlichen Wiederanlaufweg nennen (`migrate force` auf die Vorversion oder
   Backup-Restore) — genau die Doku, die beim Einspielen gebraucht wird.
4. **Abgeschlossene Plandokumente entfernen** gemäß Git-Workflow: der Storno-Plan und sein
   PRD. `docs/plans/review-v0.17.2.md` bleibt und bekommt einen Kopfhinweis, dass es den
   archivierten Stand `archiv/main-vor-v0.17.2` bewertet und die Begründung für die
   Rückstellungen dieses Plans liefert.

### Acceptance criteria

- [x] Ein 429 vom Rate-Limiter zeigt im Frontend eine Drosselungs-Meldung, nicht die
      generische Serverfehler-Meldung
- [x] `PREVIOUS_VERSION` steht in Workflow und README auf `v0.17.1`
- [x] Regel 3 im Migrations-README beschreibt den `dirty`-Fall korrekt und nennt den
      Wiederanlaufweg — belegt am Quelltext von golang-migrate v4.19.1
      (`SetVersion(ziel, dirty=true)` in eigener, sofort committeter Transaktion
      **vor** dem Lauf; `Up()` bricht danach mit `ErrDirty` ab)
- [x] Storno-Plan und Storno-PRD sind gelöscht, `review-v0.17.2.md` trägt den Kopfhinweis
      — die drei offenen Checkboxen des Storno-Plans waren zwei ausstehende
      `make verify`-Läufe (mit diesem Release nachgeholt) und ein Kriterium, dessen
      PRD-Annahme widerlegt ist; dessen Ergebnis steht dauerhaft in
      `docs/language.md` (Umbuchung → Rückfall auf den Akteur) und im Test
      `TestGetStornierungen_KorrekturUmgebuchterPositionFaelltAufAkteurZurueck`
- [x] `make lint` und `make fmt` sind sauber (in `make check` enthalten)

---

## Phase 3: Verifikation des neuen Stands

### Context

- `Makefile` — `make verify` (= `check-tools check-full`, inklusive Integrationstests),
  `make test-frontend`, `make test-e2e`.
- `.github/workflows/ci.yml` — Job `upgrade-path` migriert von `PREVIOUS_VERSION` auf HEAD.
- `e2e/tests/admin-live-reporting.spec.ts` — einziger vom Storno-Block berührter e2e-Spec.

### What to build

Vollständige Prüfung des neuen Stands, nicht stichprobenhaft — die Kombination ist neu.
Zusätzlich zwei gezielte Nachweise, die kein bestehender Test abdeckt:

- **Migration gegen einen realistischen Datenbestand.** `06_favoriten_cleanup.up.sql` hat
  keinen Test, und die Mutation `WHERE NOT EXISTS` → `WHERE EXISTS` bliebe in CI grün,
  löschte auf der Produktivinstanz aber **alle** Favoriten aller Servicekräfte. Ein
  Integrationstest muss diskriminieren: Favorit auf gelöschtem Tisch verschwindet, Favorit
  auf aktivem Tisch bleibt.
- **Antwortformat-Bruch eingrenzen.** Der Storno-Block ändert Antwortfelder des Reportings
  (`zahlungenCents` → `kassiertCents`, Wegfall der separaten Storno-Aufschlüsselung) und
  fasst mit `EigeneUebersicht` auch eine Ansicht im **Service**-Pfad an. Festzustellen ist,
  ob ein nicht neu geladenes Service-Handy nur diese eine Ansicht verliert oder ob ein
  Kernablauf betroffen ist. Das Ergebnis entscheidet, ob in Phase 4 alle Geräte oder nur
  der Rechner-Tab neu geladen werden müssen.

### Acceptance criteria

- [x] `make verify` läuft grün — in zwei Teilen gefahren, weil auf dem Entwicklungs-
      rechner ein fremder Container Port 5432 belegt: `make check` grün, dazu die
      vollständige `-tags=integration -race -p 1`-Suite gegen eine frisch migrierte
      PostgreSQL 17 auf Port 55432 (identisch zu `scripts/test-integration.sh`)
- [x] `make test-frontend` läuft grün (in `make check` über `check-frontend` enthalten)
- [x] `make test-e2e` läuft grün gegen den Dev-Stack — 37 von 37 Tests, Stack aus
      `docker-compose.e2e.yml` auf Port 8080
- [x] Der Job `upgrade-path` migriert `v0.17.1` → neuer Stand fehlerfrei — lokal
      nachgestellt: Migrationen und Seed aus `ghcr.io/nicograef/jotti-{migrate,backend}:v0.17.1`,
      dann `migrate up` (5 → 6, `dirty=false`), Backend startet auf den Altdaten
      (`/health` 200), `rebuild-projections` läuft über 40 Subjekte durch
- [x] Ein Integrationstest deckt `06_favoriten_cleanup.up.sql` diskriminierend ab und
      schlägt bei invertierter `WHERE`-Bedingung fehl —
      `backend/repository/tisch_repo/migration_favoriten_cleanup_test.go`; er liest die
      Migrationsdatei zur Laufzeit, statt ihre Anweisung zu kopieren, und deckt alle drei
      Tisch-Status ab (`deleted` verliert die Markierung, `active` und `inactive` behalten
      sie). Mutationsprobe beobachtet: `NOT EXISTS` → `EXISTS` und
      `status != 'deleted'` → `status = 'deleted'` brechen ihn beide
- [x] Dokumentiert ist, welche Clients einen Reload brauchen: **nur der Admin-Browser.**
      Empirisch bestimmt, nicht geschlossen: die v0.17.1-Zod-Schemata wurden gegen die
      JSON-Antworten des neuen Backends laufen gelassen. `LiveReportingDataSchema` und
      `ReportingDataSchema` werfen (fehlende `zahlungenCents`, `stornierungenProServicekraft`,
      `userId`/`userName`/`name`), `EigeneUebersichtSchema` nicht — die Service-Antwort ist
      rein additiv. Betroffen sind `/admin/auswertung`, `/admin/kassenberichte` und —
      unauffällig — `/admin/kasse`, wo die Warnung „N Tische sind noch offen" still
      verschwindet. Kein Service-Handy verliert einen Kernablauf
- [x] Bestellen, Kassieren, Stornieren, Umbuchen und Direktverkauf sind am laufenden Stack
      durchgespielt — durch die e2e-Suite gegen den echten Stack statt von Hand:
      `bestellen-kassieren`, `tischservice-teilzahlung`, `stornierung-serviceleitung`,
      `umbuchung`, `direktverkauf-storno`, dazu `admin-live-reporting` für die
      Storno-Zuordnung

---

## Phase 4: Einspiel-Dokumentation

### Context

- `docs/leitfaden/aktualisieren.md` — Standardweg für den Betreiber, drei Schritte.
- `docs/leitfaden/installation.md` — beschreibt das Print-Relay als separaten Prozess.
- `frontend/nginx.conf` — liefert `index.html` mit `Cache-Control: no-cache`, gehashte
  Assets `immutable`.

### What to build

Den Aktualisierungs-Leitfaden um den Schritt ergänzen, den dieses Update braucht: **jedes
betroffene Gerät einmal neu laden, bevor wieder gearbeitet wird.** Der Umfang folgt aus
Phase 3. Der Reload holt zuverlässig den neuen Stand, weil `index.html` nicht gecacht wird
— ausgelöst wird er von nichts, deshalb muss er angesagt werden.

Ausdrücklich **nicht** nötig: ein Austausch des Print-Relays. Der Relay-Block ist nicht im
Release, `windows/relay/` ist gegenüber v0.17.1 unverändert, und die Endpunkte `/relay/poll`
und `/relay/ergebnis` sind es ebenfalls. Das laufende Relay darf weiterlaufen. Dieser Satz
gehört in den Leitfaden, weil er sonst beim nächsten Update falsch geraten wird.

Dazu eine knappe Reihenfolge für den Betreiber, die den Rückweg über `jotti-restore.cmd`
und das automatische Backup benennt, sowie einen Rauchtest: an einem Tisch bestellen,
kassieren, stornieren, einen Direktverkauf tätigen, das Live-Dashboard öffnen und dort die
Storno-Zuordnung je Servicekraft sehen.

### Acceptance criteria

- [x] `docs/leitfaden/aktualisieren.md` nennt den Reload-Schritt mit dem in Phase 3
      ermittelten Umfang — samt Abschnitt „Bei Version 0.17.2: nur der Rechner des Admins"
      und der Reload-Geste für die installierte PWA
- [x] Der Leitfaden stellt fest, dass das Print-Relay bei diesem Update nicht getauscht wird
- [x] Der Rückweg (`jotti-restore.cmd`, automatisches Backup) ist benannt — mit dem
      tatsächlichen Ablauf aus `packaging/windows/jotti-restore.cmd` und der Warnung,
      dass mitten im Fest alles seit dem Backup verloren geht
- [x] Ein Rauchtest über die fünf Kernabläufe steht im Leitfaden

---

## Phase 5: Release

### Context

- `origin/main` steht auf `4040cd21`, keine Divergenz zum lokalen Stand.
- Offener PR #101 (dependabot) zielt auf `main`.
- `.github/workflows/release.yml` leitet die Version ausschließlich aus dem Tag-Namen ab
  und reicht sie als Build-Arg in `backend/Dockerfile` sowie an `make release-windows`
  weiter. Es gibt **keine** Versionsangabe im Repo, die von Hand zu heben wäre; das Tag
  ist die einzige Quelle.
- Das Release-ZIP enthält `jotti-relay.exe`. Da `windows/relay/` gegenüber `v0.17.1`
  unverändert bleibt, ist das mitgelieferte Relay funktional identisch mit dem laufenden —
  der Betreiber kann es tauschen oder das laufende weiterlaufen lassen.

### What to build

Archiv-Ref verifizieren, `release/v0.17.2` als neuen `main` force-pushen, `v0.17.2` taggen
und den Release-Build auslösen. Danach den dependabot-PR neu bewerten — nicht vor dem Fest
mergen.

### Acceptance criteria

- [ ] Der Archiv-Ref ist auf dem Remote nachweisbar, **bevor** der Force-Push läuft
- [ ] `origin/main` entspricht dem verifizierten Stand aus Phase 3
- [ ] Tag `v0.17.2` zeigt auf diesen Stand, CI ist grün
- [ ] Die Release-Artefakte sind gebaut und der Betreiber hat das ZIP
- [ ] PR #101 ist als „nach dem Fest" markiert

---

## Nach dem Fest

Dieser Plan endet mit dem Release. Die drei ursprünglich hier angehängten
Nach-dem-Fest-Phasen sind in eigene Pläne gezogen, damit dieser Plan vollständig
abgeschlossen werden kann und die zurückgestellte Arbeit dort steht, wo sie
umgesetzt wird:

| Ehemals | Jetzt |
|---|---|
| Phase 6 — Ursache der fehlenden Bons klären | `docs/plans/plan-bondruck-ursachenklaerung.md` |
| Phase 7 — Versions-Handshake für Clients | `docs/plans/plan-v0.17.3.md`, Phasen 4–7 |
| Phase 8 — Rückstand aufarbeiten | `docs/plans/plan-v0.17.3.md`, Phasen 1–3 |

`docs/plans/review-v0.17.2.md` bleibt als Begründung der Zurückstellungen bestehen,
bis der Rückstand abgearbeitet ist; es bewertet den archivierten Stand
`archiv/main-vor-v0.17.2`, nicht das ausgelieferte v0.17.2.
