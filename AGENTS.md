# Agent Instructions — jotti

jotti ist ein **kostenloses Mobile-Kassensystem (mPOS)** für Vereine und gemeinnützige Organisationen. Zielgruppe: eingetragene Vereine (e.V.), gGmbH, gUG, Stiftungen, kirchliche Träger — für temporäre Gastronomie-Veranstaltungen (Vereinsfeste, Weihnachtsmärkte, Maihocks, Konzerte, 2–3 Mal pro Jahr, 5–50 Tische, 5–30 ehrenamtliche Helfer).

Servicekräfte nehmen auf ihren eigenen Smartphones (BYOD) im Browser Bestellungen auf, bestätigen die Ausgabe, kassieren und stornieren — alles pro Tisch. Admins verwalten Produkte, Tische und Benutzer. Self-hosted per Docker Compose, proprietäre Source-Available-Lizenz (Non-Commercial, Nutzungsvereinbarung erforderlich), Mobile-first.

**Bewusst NICHT enthalten:** Kartenzahlung, Reservierungen, Warenwirtschaft, Lieferservice, Multi-Standort, CRM, Kiosk-Modus. Diese Reduktion ist gewollt — jedes zusätzliche Feature erhöht Komplexität für ehrenamtliche Teams.

**Compliance-Roadmap (TSE/KassenSichV):** jotti ist ein elektronisches Aufzeichnungssystem im Sinne von § 1 KassenSichV und unterliegt damit der TSE-Pflicht nach § 146a AO. Die TSE-Integration (fiskaly Cloud-TSE, DSFinV-K-Export, ELSTER-Meldepflicht) wird schrittweise implementiert — siehe `docs/anforderungen.md` und `docs/compliance.md`.

## Instruktionshierarchie

- `AGENTS.md` ist die kanonische repo-weite Quelle für Produktkontext, Arbeitsregeln, Qualitätsprinzipien und Agenten-Workflow.
- `.github/copilot-instructions.md` ist bewusst kurz und spiegelt nur wenige harte Guardrails als Sicherheitskopie; bei Konflikten gilt `AGENTS.md`.
- `.github/instructions/*.instructions.md` ergänzen ausschließlich bereichsspezifische Konventionen für ihre `applyTo`-Bereiche und wiederholen keine generischen Repo-Regeln.

## Referenzdokumente

Die folgenden Dokumente beschreiben jotti vollständig. Sie werden **nicht automatisch geladen** (zu groß). Lies gezielt den relevanten Abschnitt per `read_file`, nicht das ganze Dokument.

| Dokument                      | Inhalt                                                                                                                                 | Lesen bei                                                   |
| ----------------------------- | -------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------- |
| `docs/anforderungen.md`       | Funktionale und querschnittliche Anforderungen, Akzeptanzkriterien, Priorisierung, Status (K-01–K-24, F-01–F-09, S-01–S-03, A-01–A-03, Q-01–Q-08, R-01–R-07) | neue Features, Akzeptanzkriterien, Rollen                   |
| `docs/handbuch.md`            | Architektur, Bounded Contexts, Invarianten, Event-Sourcing, Schichtenarchitektur, Read Models, Tagesabschluss, Bondruck                | Architekturentscheidungen, Invarianten, Endpunkte entwerfen |
| `docs/language.md`            | Verbindliche Fachbegriffe, Namenskonventionen pro Schicht (Go, TS, JSON, DB), Ist/Soll-Abweichungen                                    | Benennungen klären, neue Felder/Typen benennen              |
| `docs/produktbeschreibung.md` | Produktidentität, Positionierung, Personas, Abgrenzung                                                                                 | Zielgruppe verstehen, Positionierung                        |
| `docs/steuerrecht.md`         | Umsatzsteuerrecht Gastronomie ab 2026: Steuersätze, Ausnahmen, Kombi-Splitting, Gutscheine, Belegpflichtangaben                        | Steuerregeln verstehen, Steuersatz-Zuordnung, F-07-Arbeit   |
| `docs/compliance.md`          | KassenSichV, GoBD, DSFinV-K, ELSTER; Betreiberpflichten, TSE-Adapter-Interface                                                         | Compliance-Features implementieren, Betreiberdokumentation  |

## Tech-Stack

| Komponente    | Technologie                                                                     |
| ------------- | ------------------------------------------------------------------------------- |
| Backend       | Go 1.26, stdlib `net/http`, `pgx/v5`, `sqlc`, `zerolog`, `zog`, `golang-jwt/v5` |
| Frontend      | React 19, Vite 8, TypeScript 5.9 (strict), Tailwind CSS 4, shadcn/ui, Zod 4     |
| Datenbank     | PostgreSQL 17, `golang-migrate`                                                 |
| Runtime       | Node 24+, pnpm 10+                                                              |
| Infrastruktur | Docker Compose, nginx Reverse Proxy, Let's Encrypt                              |

## Befehle

Alle Befehle werden über das **Makefile** ausgeführt (`make help` für die vollständige Liste):

| Befehl        | Beschreibung                                 |
| ------------- | -------------------------------------------- |
| `make check`  | Schnelle Repo-Prüfung ohne DB-Integration    |
| `make verify` | Vollständige Prüfung inkl. Integrationstests |
| `make test`   | Backend Unit-Tests                           |
| `make lint`   | Backend + Frontend Linting                   |
| `make fmt`    | Backend + Frontend Formatierung              |
| `make build`  | Backend + Frontend kompilieren               |
| `make sqlc`   | sqlc Code generieren (nach Query-Änderungen) |
| `make dev`    | Dev-Stack starten (Docker Compose)           |

## Aktive Entwicklungsphase

jotti befindet sich in aktiver Entwicklung (Pre-Release). **Breaking Changes sind ausdrücklich erwünscht** — es gibt keine produktiven Instanzen und keine Nutzer, auf die Rücksicht genommen werden muss.

- **DB-Schema:** Änderungen direkt in `database/migrations/01_initial.up.sql` vornehmen. Keine neuen Migrationsdateien anlegen. Die vorhandene `01_initial.down.sql` dient ausschließlich dem lokalen Dev-Reset und ist kein produktiver Migrations-Pfad. Dev-DB neu aufsetzen: `make down && make dev`.
- **Backend-API:** Endpunkte, Request-/Response-Formate und JSON-Keys direkt ändern. Keine API-Versionierung, keine Migrations-Strategien.
- **Event-Formate:** Event-Data-Strukturen und JSON-Keys direkt ändern. Kein Dual-Read, kein Custom `UnmarshalJSON` für alte Daten. Alte Events werden nicht migriert.
- **Frontend:** Typen, Schemas und Komponenten direkt an geänderte Backend-Datenformate anpassen.

## Wichtige Regeln

1. **Alle API-Endpunkte sind POST-only.** Keine GET/PUT/DELETE.
2. **Geldbeträge sind immer in Cent (int).** Niemals Floats für Geld verwenden.
3. **Event-Sourcing für Kasse-Operationen.** Das Kassenjournal (`kassenjournal`-Tabelle) ist immutable (append-only). Nie Einträge im Kassenjournal updaten oder löschen. Eine synchrone Projektion (`tisch_sessions`) und eine CRUD-Entität (`kassensitzungen`) werden in derselben Transaktion aktualisiert.
4. **CRUD für Stammdaten** (Benutzer, Produkte, Tische). Soft-Deletes via `status = 'deleted'`.
5. **Validierung mit Schemas.** Backend: `zog`. Frontend: `Zod`. Beide Seiten validieren.
6. **Deutsche Ubiquitous Language.** Fachbegriffe der Domäne sind deutsch (Bestellung, Zahlung, Ausgabe, Stornierung, Tisch, Position). Infrastruktur-Code (Auth, Config, DB) bleibt englisch. Alle Benutzer-sichtbaren Strings auf Deutsch. Commits auf Englisch.
7. **Kein globaler State-Store im Frontend.** Nur React Hooks + Singletons.
8. **Frontend API-Aufrufe nur über Backend-Klassen.** Nie direkt `fetch()` verwenden. Alle Domain-Backend-Klassen nutzen das `BackendClient`-Interface aus `src/lib/Backend.ts`.
9. **Backend ist die Single Source of Truth für Daten-Filterung.** Filterung, Aggregation und Aufbereitung gehören ins Backend. Das Frontend zeigt an, was das Backend liefert.
10. **Domain-Modelle tragen keine `json`-Tags.** Die Domain-Schicht (`domain/`) kennt kein HTTP und keine Serialisierung. `json`-Tags gehören ausschließlich in Response-DTOs der HTTP-Schicht (`api/<domain>/http/`) und in Event-Data-Structs (für Event-Store-Persistenz). Domain-Structs werden nie direkt als API-Response serialisiert.
11. **Verifizieren statt vermuten.** Vor jeder Aussage über bestehenden Code, Architektur, Benennung oder Verhalten muss die Codebasis durchsucht werden (grep, file search, semantic search, read file). Nie raten, was eine Datei enthält, was eine Funktion tut oder wie ein Feature funktioniert — immer den tatsächlichen Quellcode lesen. Domänenbezogene Behauptungen gegen `docs/` gegenprüfen.
12. **Fragen statt annehmen.** Bei Unsicherheit über Anforderungen, Design-Absicht oder Erwartungen des Nutzers muss der Clarify-Skill oder das AskQuestion-Tool verwendet werden, um Unklarheiten mit strukturierten Fragen zu klären. Mit dokumentierten Annahmen fortfahren ist nur akzeptabel, wenn der Nutzer ausdrücklich ablehnt zu antworten.
13. **Websuche für externes Wissen.** Bei Arbeit mit externen Bibliotheken, Sprachfeatures, APIs, Compliance-Vorgaben oder anderem Wissen außerhalb der Projekt-Codebasis sollen autoritative Quellen im Web gesucht werden (offizielle Dokumentation, RFCs, Spezifikationen) statt sich auf potenziell veraltetes Trainingswissen zu verlassen. Verifizierte Fakten immer gegenüber erinnerten Informationen bevorzugen.
14. 🚫 **`sqlc/dbgen/` niemals editieren** (generierter Code).
15. ✅ **`make sqlc`** nach Query-Änderungen ausführen; **`make lint`** nach Code-Änderungen.
16. ⚠️ **Erst fragen** vor neuen Dependencies oder Änderungen an Docker/Nginx-Konfiguration.
17. 🚫 **Keine Secrets oder Passwörter** in den Code committen.

## Qualitätsprinzipien

- **Bewertungsmetriken — der Maßstab für jede Änderung, in jedem Chat-Modus (Ask, Plan, Agent):**
  - **Maßgeblich, immer optimieren:** Korrektheit, Einfachheit, Codequalität, Konsistenz.
  - **Bewusst nachrangig, nie ein Gegenargument:** Aufwand, Zeit, Arbeitsumfang, Kosten, Breaking Changes.
  - Eine korrekte, einfache, saubere und konsistente Lösung wird nie zugunsten einer schnelleren, kleineren oder bequemeren Variante verworfen. Mehr Arbeit allein ist kein Grund, die richtige Lösung zu vermeiden.
  - **„Arbeitsumfang“ ist nicht Feature-Scope.** Aufwandsscheu wird ignoriert; der Scope Guard bleibt unberührt: keine ungefragten Features, kein Gold-Plating (siehe „Scope Guard“ unten).
- **Menschlich reviewbare Änderungen.** Jede Änderung muss sauber, lesbar und wartbar genug sein, damit ein Senior-Entwickler sie langfristig reviewen, verstehen und pflegen kann. Keinen cleveren Code, keine unnötigen Abstraktionen, keine Änderungen, die tiefen Kontext erfordern um verstanden zu werden.
- **Self-Review-Checkliste** (vor dem Präsentieren der Änderungen still durchlaufen, nur gefundene Probleme im Chat melden):
  1. Sind die Änderungen **korrekt** — lösen sie tatsächlich das genannte Problem?
  2. Sind die Änderungen **sauber** — kein toter Code, keine Debug-Artefakte, konsistenter Stil?
  3. Sind die Änderungen **lesbar** — würde ein menschlicher Reviewer sie ohne zusätzliche Erklärung verstehen?
  4. Sind die Änderungen **wartbar** — kein Over-Engineering, keine unnötigen Abstraktionen?
  5. Sind die Änderungen **im Scope** — nichts über das Geforderte oder klar Notwendige hinaus?
  6. Sind die Änderungen **vollständig** — Tests, Validierung, beide Seiten aktualisiert wo nötig?
- **Scope Guard.** Wenn der Agent bemerkt, dass er Änderungen außerhalb des Aufgabenumfangs macht oder machen will, muss er stoppen, die Out-of-Scope-Änderungen benennen und den Nutzer fragen, bevor er fortfährt.

## Bereiche

- **Admin** (`admin`): Routen `/admin/*` (`api/admin.go`), Frontend `src/admin/`, `AdminGuard`. Produkte, Tische, Benutzer verwalten. Kassensitzung eröffnen/verwalten, Kassenbestand, Kassensturz, Tagesabschluss. Reporting.
- **Service** (`admin` + `serviceleitung` + `service`): Routen `/service/*` (`api/service.go`), Stornierung über `api/serviceleitung.go`. Frontend `src/service/`, `ServiceGuard`. Bestellen, Ausgabe bestätigen, Kassieren, Stornieren.
- **Auth**: Routen `/auth/*` (`api/auth.go`). Login, Passwort setzen. JWT-Token (Benutzer-ID + Rolle).

## Git-Workflow

- **Commit-Messages:** Conventional Commits auf Englisch (`feat:`, `fix:`, `refactor:`, `docs:`, `test:`)
- **Kein auto-commit.** Agent schlägt Commit-Message vor, User führt Commit durch.
- **Kein `--force` push oder `--no-verify`.**
- **Abgeschlossene Pläne werden nach dem Merge aus `docs/plans/` gelöscht** (die Git-Historie bewahrt sie); im Arbeitsbaum bleiben nur Pläne mit offenen Checkboxen.
- **Nach jeder abgeschlossenen Aufgabe** postet der Agent eine vollständige, kopierfähige Conventional-Commit-Message in den Chat. Format:

  ```
  <type>(<scope>): <kurze Beschreibung>

  <Body: Was wurde geändert und warum. Mehrzeilig bei Bedarf.>

  <Footer: Breaking Changes, Issue-Referenzen etc. nur wenn zutreffend.>
  ```

  Die Message muss alle geänderten Dateien/Bereiche abdecken, den Kontext der Änderung erklären und direkt per Copy-Paste als `git commit -m` verwendbar sein.

- **Zusammenfassung für den Reviewer.** Nach jeder abgeschlossenen Aufgabe postet der Agent — zusätzlich zur Commit-Message — einen narrativen Absatz (in Konversationssprache), der erklärt: was wurde geändert, warum, und worauf der Reviewer achten sollte. Für einen Senior-Entwickler, der Intent und Impact schnell verstehen will, ohne jede Diff-Zeile zu lesen.

## Lokale Qualitaetspruefung

Fuer reproduzierbare lokale Checks (CI-nah):

```bash
bash scripts/setup-dev-tools.sh  # Tools installieren (einmalig)
make check                        # Schneller Check ohne Integrationstests
make verify                       # Voller Check inkl. Integrationstests
```

Bei `Fehlendes Tool: ...`: `bash scripts/setup-dev-tools.sh` erneut ausfuehren und sicherstellen, dass `$(go env GOPATH)/bin` im `PATH` liegt.
