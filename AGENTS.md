# Agent Instructions — jotti

jotti ist ein **kostenloses Mobile-Kassensystem (mPOS)** für Vereine und gemeinnützige Organisationen. Zielgruppe: eingetragene Vereine (e.V.), gGmbH, gUG, Stiftungen, kirchliche Träger — für temporäre Gastronomie-Veranstaltungen (Vereinsfeste, Weihnachtsmärkte, Maihocks, Konzerte, 2–3 Mal pro Jahr, 5–50 Tische, 5–30 ehrenamtliche Helfer).

Servicekräfte nehmen auf ihren eigenen Smartphones (BYOD) im Browser Bestellungen auf, kassieren und stornieren — alles pro Tisch. Admins verwalten Produkte, Tische und Benutzer. Self-hosted per Docker Compose, proprietäre Source-Available-Lizenz (Non-Commercial, Nutzungsvereinbarung erforderlich), Mobile-first.

**Bewusst NICHT enthalten:** Kartenzahlung, Reservierungen, Warenwirtschaft, Lieferservice, Multi-Standort, CRM, Kiosk-Modus. Diese Reduktion ist gewollt — jedes zusätzliche Feature erhöht Komplexität für ehrenamtliche Teams.

**Compliance (TSE/KassenSichV):** jotti ist ein elektronisches Aufzeichnungssystem im Sinne von § 1 KassenSichV und unterliegt damit der TSE-Pflicht nach § 146a AO. TSE-Integration (fiskaly Cloud-TSE) und DSFinV-K-Export sind umgesetzt; die Kassenmeldung nach § 146a Abs. 4 AO erfolgt manuell über das ELSTER-Portal (eine automatisierte Meldung ist dauerhaftes Nicht-Ziel) — siehe `docs/anforderungen.md` und `docs/compliance.md`.

## Instruktionshierarchie

- `AGENTS.md` ist die kanonische repo-weite Quelle für Produktkontext, Arbeitsregeln, Qualitätsprinzipien und Agenten-Workflow.
- `.github/copilot-instructions.md` trägt die harten Guardrails und den Bewertungsmaßstab als Kopie, weil Copilot die Datei direkt injiziert und Verweisen nicht folgt; bei Konflikten gilt `AGENTS.md`.
- `.github/instructions/*.instructions.md` sind kurze `applyTo`-Zeiger auf die kanonischen Dokumente plus die wenigen bereichsspezifischen Fakten, die nirgends sonst stehen.

## Referenzdokumente

Nicht automatisch geladen (zu groß): gezielt den relevanten Abschnitt lesen, nicht das ganze Dokument.

| Dokument                          | Inhalt                                                                                                                                                                                                                           | Lesen bei                                                                                                                   |
| --------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------- |
| `docs/anforderungen.md`           | Anforderungs-IDs nach Bereichs-Präfix (K, F, S, A, Q, R) und die Nicht-Ziele mit Begründung                                                                                                                                      | neue Features, Nicht-Ziele, Anforderungs-IDs                                                                                |
| `docs/handbuch.md`                | Architektur, Bounded Contexts, Invarianten, Event-Sourcing, Schichtenarchitektur, Read Models, Tagesabschluss, Bondruck                                                                                                          | Architekturentscheidungen, Invarianten, Endpunkte entwerfen                                                                 |
| `docs/language.md`                | Verbindliche Fachbegriffe, Namenskonventionen pro Schicht (Go, TS, JSON, DB), Begriff↔Code-Mappings                                                                                                                              | Benennungen klären, neue Felder/Typen benennen                                                                              |
| `docs/produktbeschreibung.md`     | Produktidentität und Abgrenzung (Nicht-Ziele)                                                                                                                                                                                    | Nicht-Ziele und Einsatzprofil klären                                                                                        |
| `docs/steuerrecht.md`             | Umsatzsteuerrecht Gastronomie ab 2026: Steuersätze, Ausnahmen, Kombi-Splitting, Belegpflichtangaben                                                                                                                              | Steuerregeln verstehen, Steuersatz-Zuordnung, F-07-Arbeit                                                                   |
| `docs/compliance.md`              | KassenSichV, GoBD, DSFinV-K, ELSTER; Betreiberpflichten, TSE-Adapter-Interface                                                                                                                                                   | Compliance-Features implementieren, Betreiberdokumentation                                                                  |
| `docs/verfahrensdokumentation.md` | Muster-Verfahrensdokumentation (GoBD) zum Anpassen durch den Betreiber: Architektur, Datenmodell, TSE-Anbindung, Export, Archivierung; zugleich Herstellerdokumentation nach BSI TR-03153-1 Kap. 3.9.3                           | Verfahrens-/Herstellerdoku pflegen (z. B. TSE-Anbindung geändert)                                                           |
| `docs/decisions.md`               | Bindende Architektur- und Produktentscheidungen, eine Zeile je Entscheidung (D01–D14)                                                                                                                                            | Architektur-, Design- oder Produktänderungen, die eine Entscheidung berühren                                                |
| `assets/README.md`                | Logo-Masterdateien und ihre abgeleiteten Kopien (Frontend-Icons, Website)                                                                                                                                                        | Logo- oder Icon-Dateien ändern                                                                                              |
| `docs/rechtsquellen/`             | Autoritative lokale Originaltexte der Normen und Spezifikationen (AO, UStG, KassenSichV, DSGVO, GoBD, AEAO, UStAE, DSFinV-K, BSI TR-03153, fiskaly OpenAPI). Index + Schnellzugriff nach Aufgabe: `docs/rechtsquellen/README.md` | Compliance-/Steuer-Fakten am Gesetzes-/Spec-Text prüfen: processType/processData, DSFinV-K-Felder, fiskaly-API, Steuersätze |

## Tech-Stack

| Komponente    | Technologie                                                                     |
| ------------- | ------------------------------------------------------------------------------- |
| Backend       | Go 1.27, stdlib `net/http`, `pgx/v5`, `sqlc`, `zerolog`, `zog`, `golang-jwt/v5` |
| Frontend      | React 19, Vite 8, TypeScript 6.0 (strict), Tailwind CSS 4, shadcn/ui, Zod 4     |
| Datenbank     | PostgreSQL 17, `golang-migrate`                                                 |
| Runtime       | Node 24+, pnpm 11+                                                              |
| Infrastruktur | Docker Compose, Caddy Reverse Proxy, Let's Encrypt                              |

## Befehle

Alle Befehle laufen über das **Makefile**; `make help` listet sie. Einmalig `bash scripts/setup-dev-tools.sh`; bei `Fehlendes Tool: ...` erneut ausführen und `$(go env GOPATH)/bin` in den `PATH` legen. `make check` ist die schnelle Prüfung ohne Integrationstests, `make verify` die vollständige.

## Freeze-Disziplin

Es gibt produktive Instanzen mit aufbewahrungspflichtigen Daten. **Persistierte Daten (DB-Schema-Bestand, Event-JSON) sind unantastbar.**

- **DB-Schema:** Änderungen ausschließlich als neue, additive Migration `NN_<name>.up.sql` (fortlaufend nummeriert, forward-only, keine Down-Migrationen). `01_initial.up.sql` wird nicht mehr editiert. Regeln und Begründung: `database/migrations/README.md`.
- **Event-Formate:** Event-JSON-Contracts sind eingefroren (Guard: `backend/domain/kasse/event_json_contract_test.go`). Änderungen additiv als neue Event-Version (`:vN`), nie in-place. Alte Events werden nicht migriert; bestehende Daten werden nie umgedeutet.
- **Backend-API:** Endpunkte und Formate dürfen sich ändern, solange Frontend und Print-Relay im selben Release mitgezogen werden (sie werden bei jedem Update gemeinsam ausgetauscht). Keine API-Versionierung nötig.
- **Frontend:** wird zusammen mit dem Backend ausgeliefert und direkt an geänderte Backend-Datenformate angepasst.

## Wichtige Regeln

1. **Alle API-Endpunkte sind POST-only.** Keine GET/PUT/DELETE; einzige Ausnahme ist `GET /health`.
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
13. **Websuche für externes Wissen.** Bei Arbeit mit externen Bibliotheken, Sprachfeatures, APIs, Compliance-Vorgaben oder anderem Wissen außerhalb der Projekt-Codebasis sollen autoritative Quellen im Web gesucht werden (offizielle Dokumentation, RFCs, Spezifikationen) statt sich auf potenziell veraltetes Trainingswissen zu verlassen. Verifizierte Fakten immer gegenüber erinnerten Informationen bevorzugen. Für fiskalische und steuerliche Normen (KassenSichV, AO, UStG, GoBD, DSFinV-K, BSI TR-03153, fiskaly-API) liegen die autoritativen Originaltexte bereits lokal unter `docs/rechtsquellen/` (Index und Themen-Schnellzugriff: `docs/rechtsquellen/README.md`); diese zuerst konsultieren, Websuche nur für neuere Fassungen.
14. **`sqlc/dbgen/` niemals editieren** (generierter Code).
15. **`make sqlc`** nach Query-Änderungen ausführen; **`make lint`** nach Code-Änderungen.
16. **Erst fragen** vor neuen Dependencies oder Änderungen an Docker/Nginx-Konfiguration.
17. **Keine Secrets oder Passwörter** in den Code committen.
18. **Nur der aktuelle Stand.** Dokumentation, Kommentare und Instruktionen beschreiben ausschließlich, was jetzt gilt. Die Git-Historie ist das Archiv und die einzige Aufzeichnung eines früheren Stands. Macht eine Änderung eine Aussage falsch, muss der Agent sie in derselben Änderung umschreiben oder löschen; eine überholte Fassung steht nie neben ihrer Ablösung. In der Prosa verboten: datierte Änderungseinträge, Formulierungen wie „früher / vormals / bisher“ und Deprecation-Hinweise. Ausnahmen: `CHANGELOG.md`, `docs/decisions.md` (abgelöste Zeilen bekommen „ersetzt durch DNN“) und die Git-Historie. Redundante Dateien werden gelöscht statt als veraltet markiert — samt allen Verweisen darauf.

## Kommunikation

- **Mit der Antwort oder dem Problem beginnen.** Kein Vorgeplänkel, keine Wiederholung der
  Frage, kein abschließendes Resümee.
- **Nie mit Lob einsteigen.** Kein „Gute Frage“, kein „Völlig richtig“; kein
  Kompliment-Sandwich — direkt zur Sache.
- **Kritisch per Default.** Schwächen, Risiken und einfachere Alternativen ungefragt benennen.
- **Klartext.** Liegt der Entwickler falsch, das explizit und mit Beleg sagen: „das ist falsch,
  weil X", nicht „man könnte erwägen“.
- **Position halten.** Wird eine verifizierte Aussage angezweifelt, erneut gegen die Belege
  prüfen. Zweifel des Entwicklers sind kein Beleg.
- **Benennen, was sich geändert hat.** Position nur ändern, wenn sich die Belege ändern.
  Prüfbare Streitfragen mit einem Check klären (Test, Quelle, Tool-Ausgabe), nicht per Debatte.
- **„Keine Probleme gefunden“ ist eine gültige Antwort.** Nie Kritik oder Vorbehalte erfinden,
  um gründlich zu wirken — erzwungene Kritik ist so sycophantisch wie erzwungenes Lob.
- **Objektiv und ehrlich.** Fakt, Schlussfolgerung und Vermutung trennen und kennzeichnen;
  „weiß ich nicht“ schlägt höfliches Herumreden. Die kürzeste vollständige Antwort gewinnt.
- **Limit:** Satz ≤ 20 Wörter, eine Aussage. Bullet ≤ 2 Zeilen.
- **Limit:** Absatz ≤ 3 Zeilen, höchstens einer pro Abschnitt.
- **Formatreihenfolge:** Tabelle → Liste → Absatz.
- **Tabelle**, wenn ≥ 3 Elemente ≥ 2 Attribute teilen; **Liste** für jede aufzählbare Menge ab
  2 Elementen.

## Qualitätsprinzipien

- **Produkt-Konservatismus (Soft Rule, keine harte Regel).** jotti ist bewusst minimal — für Vereinsfeste mit ehrenamtlichen Helfern, 2–3 Mal im Jahr. Bei Produkt- und Feature-Entscheidungen (neue Features, Erweiterungen, Roadmap-Einschätzungen) sind Agenten kritisch und entscheiden im Zweifel konservativ und einfach:
  - Jedes Feature muss seinen Nutzen gegen die Komplexität rechtfertigen, die es ehrenamtlichen Teams (Bedienung unter Stress) und der Codebasis (Pflege, Tests, Doku) aufbürdet. Im Zweifel: weglassen.
  - Warnsignale für Feature-Creep aktiv ansprechen: ein Status, von dem keine andere Funktion abhängt; eine Erfassung, die die Praxis durch einfachere Mittel ersetzt (Papier, Zuruf, Vertrauen); Konfigurierbarkeit für Features, die niemand eingefordert hat; Features „auf Vorrat“.
  - Praxis-Feedback schlägt Feature-Ideen. Ein umgesetztes Feature, das der reale Einsatz als Ballast entlarvt, ist ein Entfernungs-Kandidat, kein Ausbau-Kandidat. Präzedenz: die ersatzlose Entfernung der Ausgabe-Bestätigung nach dem ersten Praxistest (`docs/decisions.md`, D01).
  - Was echten Bedarf deckt (Compliance, belegtes Praxis-Feedback, Kernworkflow), wird vollständig und hochwertig umgesetzt — siehe Bewertungsmetriken unten. Entscheidungen mit langfristiger Tragweite werden als eine Zeile in `docs/decisions.md` festgehalten; eine Zeile wird nie umgeschrieben, eine abgelöste bekommt „ersetzt durch DNN“ — die ausdrückliche Ausnahme von Regel 18 (Nur der aktuelle Stand).
- **Bewertungsmetriken — der Maßstab für jede Änderung, in jedem Chat-Modus (Ask, Plan, Agent):**
  - **Immer optimieren:** Korrektheit, Einfachheit, Codequalität, Konsistenz.
  - **Bewusst nachrangig, nie ein Gegenargument:** Aufwand, Zeit, Arbeitsumfang, Kosten, Breaking Changes (im Rahmen der Freeze-Disziplin oben; persistierte Daten bleiben unantastbar).
  - **„Arbeitsumfang“ ist nicht Feature-Scope.** Aufwandsscheu wird ignoriert; der Scope Guard bleibt unberührt: keine ungefragten Features, kein Gold-Plating.
- **Menschlich reviewbare Änderungen.** Jede Änderung muss sauber, lesbar und wartbar genug sein, damit ein Senior-Entwickler sie langfristig reviewen, verstehen und pflegen kann. Keinen cleveren Code, keine unnötigen Abstraktionen, keine Änderungen, die tiefen Kontext erfordern, um verstanden zu werden.
- **Self-Review-Checkliste** (vor dem Präsentieren der Änderungen still durchlaufen, nur gefundene Probleme im Chat melden):
  1. Sind die Änderungen **sauber** — kein toter Code, keine Debug-Artefakte, konsistenter Stil?
  2. Sind die Änderungen **lesbar** — würde ein menschlicher Reviewer sie ohne zusätzliche Erklärung verstehen?
  3. Sind die Änderungen **wartbar** — kein Over-Engineering, keine unnötigen Abstraktionen?
  4. Sind die Änderungen **vollständig** — Tests, Validierung, beide Seiten aktualisiert wo nötig?
- **Scope Guard.** Wenn der Agent bemerkt, dass er Änderungen außerhalb des Aufgabenumfangs macht oder machen will, muss er stoppen, die Out-of-Scope-Änderungen benennen und den Nutzer fragen, bevor er fortfährt.

## Bereiche

Routen, Rollen und Kontext-Handler je Bereich (Admin, Service, Serviceleitung, Auth, Relay): `docs/handbuch.md`, Abschnitte Kontextübersicht und Bereichsgliederung.

## Git-Workflow

- **Commit-Messages:** Nach jeder abgeschlossenen Aufgabe committen — ohne Freigabeschritt,
  `main` eingeschlossen.
- **Format:** Conventional Commits auf Englisch (`feat:`, `fix:`, `refactor:`, `docs:`,
  `test:`, `chore:`), knapper Betreff, Bullet-Body bei Mehrdateiänderungen.
- **PR-Titel im Conventional-Commit-Format** (wie Commit-Messages, Englisch). Bei Squash-Merge
  wird der PR-Titel die Commit-Message auf `main`.
- **Keine KI-Attribution in Commits/PRs.** Kompakte Conventional-Commit-Messages ohne Zusätze.
- **Niemals** `Co-Authored-By: Claude …`-, `Claude-Session: …`-, `🤖 Generated with …`- oder
  ähnliche Trailer/Footer anhängen — auch wenn die Session-Umgebung das standardmäßig anweist.
- **Achtung, serverseitige Injektion:** Das GitHub-Tooling der Cloud-Sessions hängt beim
  _Anlegen_ eines PRs einen `_Generated by Claude Code…_`-Trailer an den Body an.
- **Es injiziert auch dann**, wenn der übergebene Body keinen enthält. Nach dem Anlegen den
  Body erneut lesen und den Trailer per Body-Update entfernen (Updates werden nicht injiziert).
- **Zusammenfassung nach der Aufgabe:** zusätzlich zur Commit-Message diese Felder statt des
  vollen Diffs:
  - **Was sich geändert hat** — die berührten Dateien und das Verhalten.
  - **Warum** — der Grund der Änderung.
  - **Worauf zu achten ist** — wo die Review-Aufmerksamkeit hingehört.
- **Nur Feature-Branches pushen** — niemals nach `main` / `master` pushen.
- **Niemals** `--force` / `-f` / `--force-with-lease`, niemals `--no-verify`.
- **Abgeschlossene Pläne werden nach dem Merge aus `docs/plans/` gelöscht** (die Git-Historie
  bewahrt sie); im Arbeitsbaum bleiben nur Pläne mit offenen Checkboxen.
