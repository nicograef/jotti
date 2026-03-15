Details siehe docs/agents/plan-service-dashboard.md

## Agent Anweisungen

> **Lies diese Anweisungen vollständig, bevor du mit der Arbeit beginnst.**

### Kontext laden (vor jedem Abschnitt)

Bevor du einen Abschnitt beanspruchst, lies **genau die Dateien, die im `Kontext:`-Block des Abschnitts aufgelistet sind** — nicht mehr, nicht weniger. Zusätzlich:

- Bereits erstellte/geänderte Dateien aus vorherigen Abschnitten lesen (um nahtlos anzuknüpfen)

Diese Dateien werden in jeder neuen Session erneut gelesen — die Kontext-Beschaffung ist kein eigener Abschnitt, sondern Pflicht vor jeder Arbeit.

### Abschnitt beanspruchen

1. **Lies die gesamte plan.md** — insbesondere den Parallelisierungs-Abschnitt und alle Abschnitts-Überschriften.
2. **Finde den nächsten verfügbaren Abschnitt.** Ein Abschnitt ist verfügbar, wenn:
   - Er offene Tasks hat (`- [ ]`)
   - Er **nicht** mit 🔒 oder ✅ markiert ist
   - Seine Abhängigkeiten erfüllt sind (alle Vorgänger-Abschnitte sind ✅)
3. **Beanspruche den Abschnitt sofort**, indem du `🔒` an die Überschrift anhängst (`## Abschnitt N: Titel` → `## Abschnitt N: Titel 🔒`). Erst danach mit der Arbeit beginnen.
4. **Falls kein verfügbarer Abschnitt existiert: Stoppe sofort, ohne Änderungen vorzunehmen.** Erkläre dem User: welche Abschnitte noch offen sind, warum sie nicht bearbeitet werden können (🔒 = anderer Agent arbeitet daran, oder Abhängigkeiten noch nicht ✅), und welche Vorgänger-Abschnitte zuerst abgeschlossen werden müssen. **Führe keine Änderungen an Dateien durch.**

### Abschnitt abarbeiten

1. **Ein Task nach dem anderen.** Arbeite Tasks innerhalb des Abschnitts sequentiell ab — von oben nach unten.
2. **Sofort abhaken.** Ändere `- [ ]` zu `- [x]` in dieser Datei **unmittelbar** nachdem ein Task erfolgreich erledigt ist. Nicht erst am Ende des Abschnitts, nicht gebündelt — **nach jedem einzelnen Task**. Verwende beim Abhaken immer die **Abschnitts-Überschrift + den vollständigen Task-Text** als Kontext, damit die Ersetzung eindeutig ist.
3. **Abschnitt abschließen.** Wenn du an Code gearbeitet hast: Wenn alle Tasks eines Abschnitts `[x]` sind, führe die wichtigsten Dev-Scripte und CI-Steps lokal aus: compilation, build, linting, formatting, testing. Stelle sicher, dass es keine Fehler oder Warnings gibt. Erst dann ist der Abschnitt fertig. Wenn du an Dokumentation gearbeitet hast: Lese Korrektur, stelle sicher, dass alle Links funktionieren, und dass die Formatierung korrekt ist.
4. **✅ setzen.** Ersetze `🔒` durch `✅` in der Abschnitts-Überschrift (`## Abschnitt N: Titel 🔒` → `## Abschnitt N: Titel ✅`).
5. **Stoppen.** Nach Abschluss eines Abschnitts: **stopp**. Beginne nicht den nächsten Abschnitt, sondern melde, dass der Abschnitt abgeschlossen ist.
6. **Conventional Commit Message schreiben.** Wenn du an Code gearbeitet hast: Schreibe zu deinen Änderungen bzw. dem Abschnitt eine Conventional Commit Message. Führe kein Commit selbst durch, schreibe nur die Message in den Chat, sodass diese kopiert werden kann. Wenn du an Dokumentation gearbeitet hast: Schreibe eine passende Commit Message für die Dokumentationsänderungen.

---

## Parallelisierung

Die folgenden Abschnitte können **parallel** in separaten Chat-Sessions bearbeitet werden:

- Abschnitte 2 und 3 (keine gemeinsamen Dateien — Abschnitt 2 arbeitet an Favoriten-DB/SQL/Repo, Abschnitt 3 an Reporting-SQL/Repo/Domain)

Die folgenden Abschnitte haben **Abhängigkeiten**:

- Abschnitt 1 → muss als erstes abgeschlossen sein (Dokumentation definiert Begriffe und Akzeptanzkriterien für alle folgenden Abschnitte)
- Abschnitt 2 → muss nach Abschnitt 1 abgeschlossen sein
- Abschnitt 3 → muss nach Abschnitt 1 abgeschlossen sein
- Abschnitt 4 → muss nach Abschnitt 2 abgeschlossen sein (Favoriten-Repo wird im Application-Layer verwendet)
- Abschnitt 5 → muss nach Abschnitt 3 abgeschlossen sein (Reporting-Repo wird im Application-Layer verwendet)
- Abschnitt 6 → muss nach Abschnitt 4 abgeschlossen sein (Favoriten-Application-Layer wird im HTTP-Handler aufgerufen)
- Abschnitt 7 → muss nach Abschnitt 5 und 6 abgeschlossen sein (alle Backend-Endpoints müssen registriert werden)
- Abschnitt 8 → muss nach Abschnitt 7 abgeschlossen sein (Backend-Endpoints müssen stehen, bevor Frontend-Typen definiert werden)
- Abschnitt 9 → muss nach Abschnitt 8 abgeschlossen sein (Typen und Backend-Klassen werden in Hooks verwendet)
- Abschnitt 10 → muss nach Abschnitt 9 abgeschlossen sein (Hooks werden in Komponenten und Pages verwendet)

---

## Abschnitt 1: Dokumentation aktualisieren ✅

Kontext:

- `docs/agents/plan-service-dashboard.md` — Gesamtplan mit neuen Begriffen und Anforderungen
- `docs/anforderungen.md` — Bestehende Anforderungen, K-06 und R-06 finden und erweitern
- `docs/design/language.md` — Ubiquitous Language, neue Begriffe eintragen
- `docs/design/handbuch.md` — §4 Stammdaten, Favoriten als CRUD-Relation beschreiben

- [x] In `docs/anforderungen.md`: K-06 erweitern — Dashboard zeigt primär "Meine Tische" (Favoriten) mit Rich Cards (Saldo, ausstehende Lieferungen, unbezahlte Positionen, Auszahlungsbedarf). Alle-Tische-Zugang über Drawer.
- [x] In `docs/anforderungen.md`: K-11 präzisieren — Schnellsuche als Suchfeld im Alle-Tische-Drawer, clientseitige Filterung nach Tischname/-nummer
- [x] In `docs/anforderungen.md`: R-06 präzisieren — KPI-Sektion auf dem Service-Dashboard (eigene Bestellungen Anzahl+Summe, eigene Zahlungen Anzahl+Summe)
- [x] In `docs/anforderungen.md`: Neue Anforderung K-14 "Tisch-Favoriten" hinzufügen mit Akzeptanzkriterien (Servicekraft kann Tische als Favoriten markieren/entfernen, serverseitig in DB, unabhängig pro User)
- [x] In `docs/design/language.md`: Neue Begriffe eintragen — `Favorit` (Go: `Favorit`, TS: `Favorit`, DB: `tisch_favoriten`, API: `/favorit-*`, UI: „Meine Tische") und `EigeneUebersicht` (Go: `EigeneUebersicht`, TS: `EigeneUebersicht`, API: `/get-eigene-uebersicht`, UI: „Meine Übersicht")
- [x] In `docs/design/handbuch.md` §4: Favoriten als einfache CRUD-Relation im Stammdaten-Kontext beschreiben (DB-Tabelle `tisch_favoriten`, user_id + tisch_id Composite PK, kein Aggregat, keine Events)

---

## Abschnitt 2: Database-Schema & sqlc-Queries — Favoriten

Kontext:

- `database/migrations/01_initial.up.sql:1-200` — Bestehendes Schema, insbesondere `users`-Tabelle (Zeile 8-34) und `tische`-Tabelle (Zeile 40-55) für FK-Referenzen
- `backend/sqlc/queries/tables.sql:1-20` — Bestehende Tisch-Queries als Pattern-Vorlage
- `backend/sqlc.yaml` — sqlc-Konfiguration
- `backend/sqlc/queries/table_state.sql:1-19` — Pattern für sqlc-Queries mit ON CONFLICT

- [ ] In `database/migrations/01_initial.up.sql`: Neue Tabelle `tisch_favoriten` vor dem `COMMIT;` einfügen — Spalten: `user_id INT REFERENCES users(id) NOT NULL`, `tisch_id INT REFERENCES tische(id) NOT NULL`, `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`, `PRIMARY KEY (user_id, tisch_id)`. Index auf `user_id`. Comment auf Tabelle.
- [ ] Neue Datei `backend/sqlc/queries/favoriten.sql` erstellen mit drei Queries: (1) `AddFavorit :exec` — `INSERT INTO tisch_favoriten (user_id, tisch_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, (2) `RemoveFavorit :exec` — `DELETE FROM tisch_favoriten WHERE user_id = $1 AND tisch_id = $2`, (3) `GetFavoritenByUser :many` — `SELECT tisch_id FROM tisch_favoriten WHERE user_id = $1 ORDER BY created_at ASC`
- [ ] In `backend/sqlc/queries/tables.sql`: Neuen Query `GetAktiveTischeMitFavoriten :many` hinzufügen — SELECT `t.id`, `t.name`, `COALESCE(ts.saldo_cents, 0)::integer AS saldo_cents`, `(f.user_id IS NOT NULL) AS ist_favorit` FROM `tische t` LEFT JOIN `table_state ts` ON `ts.tisch_id = t.id` LEFT JOIN `tisch_favoriten f` ON `f.tisch_id = t.id AND f.user_id = $1` WHERE `t.status = 'active'` ORDER BY `t.id ASC`
- [ ] In `backend/sqlc/queries/table_state.sql`: Neuen Query `GetTableStatesByTischIDs :many` hinzufügen — `SELECT tisch_id, saldo_cents, unbezahlte_positionen, ausstehende_positionen, gesamt_zahlungen_cents, last_event_id, last_event_version, updated_at FROM table_state WHERE tisch_id = ANY($1::int[])`
- [ ] `make sqlc` ausführen und sicherstellen, dass die Generierung fehlerfrei durchläuft

---

## Abschnitt 3: Database-Schema & sqlc-Queries — Eigene Übersicht

Kontext:

- `backend/sqlc/queries/reporting.sql:1-97` — Bestehende Reporting-Queries als Pattern-Vorlage (insbesondere `GetReportingStats` für Event-Aggregation mit CASE WHEN)
- `backend/domain/reporting/reporting.go:1-70` — Bestehende Reporting-Domain-Modelle
- `backend/sqlc.yaml` — sqlc-Konfiguration
- `.github/instructions/event-sourcing.instructions.md` — Event-Typen: `tisch.bestellung-aufgenommen:v1` und `tisch.zahlung-kassiert:v1` sind die relevanten Event-Typen (Hinweis: `tisch.bestellung-aufgegeben:v1` ist der korrekte Typ-Name laut event-sourcing.instructions.md — prüfen welcher in der DB tatsächlich verwendet wird)

- [ ] In `backend/sqlc/queries/reporting.sql`: Neuen Query `GetEigeneUebersicht :one` hinzufügen — SELECT Aggregation über `events`-Tabelle gefiltert auf `user_id = @user_id`, mit CASE WHEN für Event-Typen `tisch.bestellung-aufgenommen:v1` (Anzahl + Summe `gesamtPreisCents`) und `tisch.zahlung-kassiert:v1` (Anzahl + Summe `gesamtZahlungCents`). Pattern analog zu `GetReportingStats` (Zeile 7-30 in reporting.sql).
- [ ] In `backend/domain/reporting/reporting.go`: Neues Domain-Struct `EigeneUebersicht` hinzufügen — Felder: `AnzahlBestellungen int`, `BestellungenCents int`, `AnzahlZahlungen int`, `ZahlungenCents int`. Keine json-Tags (Domain-Schicht).
- [ ] `make sqlc` ausführen und sicherstellen, dass die Generierung fehlerfrei durchläuft

---

## Abschnitt 4: Backend — Favoriten Repository & Application-Layer

Kontext:

- `backend/repository/table_repo/types.go:1-30` — Repository-Pattern (struct, NewRepository, Konvertierungsfunktionen)
- `backend/repository/table_repo/repo.go:1-90` — Repository-Methoden (GetTable, GetAllTables, GetActiveTables)
- `backend/domain/table/tisch.go` — Bestehende Domain-Modelle `Tisch`, `AktiverTisch` — hier `AktiverTischMitFavorit` hinzufügen
- `backend/api/table/application/command.go:1-50` — Interface-Definitionen `tableRepo`, `eventRepo`, `productRepo`, Command-Struct
- `backend/api/table/application/query.go:1-95` — Query-Struct, bestehende Methoden `GetAktiveTische`, `GetTischState`
- `backend/api/table/application/errors.go:1-50` — Error-Definitionen
- `backend/sqlc/dbgen/` — Generierter Code (NICHT EDITIEREN, nur referenzieren für Typen)

- [ ] In `backend/domain/table/tisch.go`: Neues Struct `AktiverTischMitFavorit` hinzufügen — Felder: `ID int`, `Name string`, `SaldoCents int`, `IstFavorit bool`. Kein json-Tag (Domain-Schicht).
- [ ] Neue Datei `backend/repository/favorit_repo/types.go` erstellen — `Repository`-Struct mit `DB *sql.DB` und `q *dbgen.Queries`, `NewRepository(db *sql.DB) Repository`-Funktion. Pattern exakt wie `table_repo/types.go`.
- [ ] Neue Datei `backend/repository/favorit_repo/repo.go` erstellen — Drei Methoden: (1) `Add(ctx, userID, tischID int) error` → ruft `r.q.AddFavorit(ctx, ...)` auf, (2) `Remove(ctx, userID, tischID int) error` → ruft `r.q.RemoveFavorit(ctx, ...)` auf, (3) `GetByUser(ctx, userID int) ([]int, error)` → ruft `r.q.GetFavoritenByUser(ctx, userID)` auf und mappt Ergebnis auf `[]int`. Fehler-Mapping via `db.Error(err)`.
- [ ] In `backend/repository/table_repo/repo.go`: Neue Methode `GetActiveTablesWithFavorites(ctx, userID int) ([]table.AktiverTischMitFavorit, error)` hinzufügen — ruft `r.q.GetAktiveTischeMitFavoriten(ctx, userID)` auf, mappt Ergebnis auf `[]table.AktiverTischMitFavorit`.
- [ ] In `backend/repository/table_repo/repo.go`: Neue Methode `GetTableStatesByIDs(ctx, tischIDs []int) ([]table.TischState, error)` hinzufügen — ruft `r.q.GetTableStatesByTischIDs(ctx, tischIDs)` auf. JSON-Unmarshalling von `unbezahlte_positionen` und `ausstehende_positionen` analog zum Pattern in `event_repo` (JSONB → `[]table.Position`). Für jeden Ergebnis-Row: TischName aus separater GetTable-Abfrage auflösen oder als zweiten Query mitjoinen — hier TischName auf leer lassen und im Application-Layer ergänzen.
- [ ] Neues Interface `favoritRepo` in `backend/api/table/application/command.go` hinzufügen — Methoden: `Add(ctx context.Context, userID, tischID int) error`, `Remove(ctx context.Context, userID, tischID int) error`, `GetByUser(ctx context.Context, userID int) ([]int, error)`. Analog zum bestehenden `tableRepo`-Interface-Pattern.
- [ ] In `backend/api/table/application/command.go`: `Command`-Struct um Feld `FavoritRepo favoritRepo` erweitern.
- [ ] In `backend/api/table/application/command.go`: Neue Methode `FavoritHinzufuegen(ctx, userID, tischID int) error` — Prüft: (1) Tisch existiert und ist aktiv via `c.TableRepo.GetTable()`, (2) Ruft `c.FavoritRepo.Add()` auf. Fehler-Mapping: `ErrTischNotFound`, `ErrTischNotActive`, `ErrDatabase`.
- [ ] In `backend/api/table/application/command.go`: Neue Methode `FavoritEntfernen(ctx, userID, tischID int) error` — Ruft `c.FavoritRepo.Remove()` auf (idempotent, kein Fehler wenn nicht vorhanden). Fehler-Mapping: `ErrDatabase`.
- [ ] In `backend/api/table/application/query.go`: `Query`-Struct um Feld `FavoritRepo favoritRepo` erweitern (dazu muss das `favoritRepo`-Interface auch in query.go accessible sein — entweder im selben Package definiert oder importiert).
- [ ] In `backend/api/table/application/query.go`: `tableRepo`-Interface um Methode `GetActiveTablesWithFavorites(ctx context.Context, userID int) ([]table.AktiverTischMitFavorit, error)` erweitern.
- [ ] In `backend/api/table/application/query.go`: `tableRepo`-Interface um Methode `GetTableStatesByIDs(ctx context.Context, tischIDs []int) ([]table.TischState, error)` erweitern.
- [ ] In `backend/api/table/application/query.go`: Neue Methode `GetAktiveTischeMitFavoriten(ctx, userID int) ([]table.AktiverTischMitFavorit, error)` — ruft `q.TableRepo.GetActiveTablesWithFavorites(ctx, userID)` auf. Logging analog zu `GetAktiveTische`.
- [ ] In `backend/api/table/application/query.go`: Neue Methode `GetMeineTischeState(ctx, userID int) ([]table.TischState, error)` — (1) Favoriten-IDs laden via `q.FavoritRepo.GetByUser(ctx, userID)`, (2) Falls leer → leere Liste zurückgeben, (3) Batch-State via `q.TableRepo.GetTableStatesByIDs(ctx, favoritIDs)` laden, (4) Für jeden State: TischName via `q.TableRepo.GetTable(ctx, state.TischID)` auflösen und in State setzen. Logging analog zu bestehenden Query-Methoden.

---

## Abschnitt 5: Backend — Eigene Übersicht Application-Layer

Kontext:

- `backend/api/reporting/application/query.go:1-33` — Bestehendes Reporting-Application-Pattern (Query-Struct, reportingRepo-Interface, GetReporting-Methode)
- `backend/repository/reporting_repo/repo.go:1-80` — Reporting-Repo-Pattern (parallel errgroup-Queries, DB → Domain-Mapping)
- `backend/domain/reporting/reporting.go:1-70` — Domain-Modelle inkl. neues `EigeneUebersicht`-Struct (aus Abschnitt 3)

- [ ] In `backend/repository/reporting_repo/repo.go`: Neue Methode `GetEigeneUebersicht(ctx, userID int) (reporting.EigeneUebersicht, error)` — ruft `r.q.GetEigeneUebersicht(ctx, userID)` auf, mappt sqlc-Row auf `reporting.EigeneUebersicht`-Domain-Struct. Pattern analog zu `GetReporting`, aber einfacher (einzelner Query statt errgroup).
- [ ] In `backend/api/reporting/application/query.go`: Interface `reportingRepo` um Methode `GetEigeneUebersicht(ctx context.Context, userID int) (reporting.EigeneUebersicht, error)` erweitern.
- [ ] In `backend/api/reporting/application/query.go`: Neue Methode `GetEigeneUebersicht(ctx, userID int) (reporting.EigeneUebersicht, error)` — ruft `q.ReportingRepo.GetEigeneUebersicht(ctx, userID)` auf. Logging analog zu `GetReporting`.

---

## Abschnitt 6: Backend — HTTP-Handler für Favoriten & Batch-State

Kontext:

- `backend/api/table/http/command_handler.go:1-60` — Bestehendes Command-Handler-Pattern (`command`-Interface, `CommandHandler`-Struct, Request-DTOs, `helper.ReadBody`, `helper.MapError`, `middleware.UserIDKey`)
- `backend/api/table/http/query_handler.go:1-400` — Bestehendes Query-Handler-Pattern (`query`-Interface, `QueryHandler`-Struct, Response-DTOs mit json-Tags, `toPosition`/`toPositionen`-Converter, nil-safe JSON-Arrays)
- `backend/api/middleware/middleware.go:80-130` — UserIDKey/UserNameKey Extraktion aus JWT-Context

- [ ] In `backend/api/table/http/command_handler.go`: `command`-Interface um zwei Methoden erweitern: `FavoritHinzufuegen(ctx context.Context, userID, tischID int) error` und `FavoritEntfernen(ctx context.Context, userID, tischID int) error`.
- [ ] In `backend/api/table/http/command_handler.go`: Request-Struct `favoritRequest` hinzufügen — `TischID int json:"tischId"`.
- [ ] In `backend/api/table/http/command_handler.go`: Neuen Handler `FavoritHinzufuegenHandler() http.HandlerFunc` — ReadBody → UserID aus Context (`middleware.UserIDKey`) → `h.Command.FavoritHinzufuegen(ctx, userID, body.TischID)` → MapError (ErrTischNotFound → `"tisch_not_found"`, ErrTischNotActive → `"tisch_not_active"`) → `helper.SendEmptyResponse(w)`.
- [ ] In `backend/api/table/http/command_handler.go`: Neuen Handler `FavoritEntfernenHandler() http.HandlerFunc` — ReadBody → UserID aus Context → `h.Command.FavoritEntfernen(ctx, userID, body.TischID)` → MapError → `helper.SendEmptyResponse(w)`.
- [ ] In `backend/api/table/http/query_handler.go`: `query`-Interface um zwei Methoden erweitern: `GetAktiveTischeMitFavoriten(ctx context.Context, userID int) ([]t.AktiverTischMitFavorit, error)` und `GetMeineTischeState(ctx context.Context, userID int) ([]t.TischState, error)`.
- [ ] In `backend/api/table/http/query_handler.go`: Neues Response-DTO `aktiverTischMitFavorit` — Felder: `ID int json:"id"`, `Name string json:"name"`, `SaldoCents int json:"saldoCents"`, `IstFavorit bool json:"istFavorit"`.
- [ ] In `backend/api/table/http/query_handler.go`: Response-Struct `getAktiveTischeMitFavoritenResponse` — `Tische []aktiverTischMitFavorit json:"tische"`.
- [ ] In `backend/api/table/http/query_handler.go`: Neuen Handler `GetAktiveTischeMitFavoritenHandler() http.HandlerFunc` — UserID aus Context → `h.Query.GetAktiveTischeMitFavoriten(ctx, userID)` → Mapping auf `aktiverTischMitFavorit`-DTOs → SendResponse.
- [ ] In `backend/api/table/http/query_handler.go`: Neues Response-DTO `meinTischState` — Felder: `TischID int json:"tischId"`, `TischName string json:"tischName"`, `SaldoCents int json:"saldoCents"`, `UnbezahltePositionen []position json:"unbezahltePositionen"`, `AusstehendePositionen []position json:"ausstehendePositionen"`, `GesamtZahlungenCents int json:"gesamtZahlungenCents"`.
- [ ] In `backend/api/table/http/query_handler.go`: Response-Struct `getMeineTischeStateResponse` — `Tische []meinTischState json:"tische"`.
- [ ] In `backend/api/table/http/query_handler.go`: Neuen Handler `GetMeineTischeStateHandler() http.HandlerFunc` — UserID aus Context → `h.Query.GetMeineTischeState(ctx, userID)` → Mapping auf `meinTischState`-DTOs (mit nil-safe `toPositionen` wie in `GetTischStateHandler`) → SendResponse.

---

## Abschnitt 7: Backend — HTTP-Handler Eigene Übersicht & Route-Registrierung

Kontext:

- `backend/api/reporting/http/query_handler.go:1-80` — Bestehendes Reporting-HTTP-Handler-Pattern (query-Interface, QueryHandler-Struct, Request/Response-DTOs)
- `backend/api/service.go:1-43` — Service-Route-Registrierung (NewServiceApi, ProductRepo/TableRepo/EventRepo Instanziierung, Handler-Registrierung)
- `backend/api/table/http/command_handler.go` — Neue Favorit-Handler (aus Abschnitt 6)
- `backend/api/table/http/query_handler.go` — Neue Query-Handler (aus Abschnitt 6)

- [ ] In `backend/api/reporting/http/query_handler.go`: `query`-Interface um Methode `GetEigeneUebersicht(ctx context.Context, userID int) (reporting.EigeneUebersicht, error)` erweitern.
- [ ] In `backend/api/reporting/http/query_handler.go`: Response-DTO `eigeneUebersichtResponse` — `AnzahlBestellungen int json:"anzahlBestellungen"`, `BestellungenCents int json:"bestellungenCents"`, `AnzahlZahlungen int json:"anzahlZahlungen"`, `ZahlungenCents int json:"zahlungenCents"`.
- [ ] In `backend/api/reporting/http/query_handler.go`: Neuen Handler `GetEigeneUebersichtHandler() http.HandlerFunc` — UserID aus Context (`middleware.UserIDKey`) → `h.Query.GetEigeneUebersicht(ctx, userID)` → Mapping auf `eigeneUebersichtResponse` → `helper.SendResponse(w, ...)`.
- [ ] In `backend/api/service.go`: Import für `favorit_repo` und `reportingApp`/`reportingHTTP`/`reporting_repo` hinzufügen.
- [ ] In `backend/api/service.go`: `favoritRepo := favorit_repo.NewRepository(db)` instanziieren.
- [ ] In `backend/api/service.go`: `FavoritRepo: favoritRepo` zum bestehenden `tc.Command`-Struct hinzufügen (CommandHandler-Wiring).
- [ ] In `backend/api/service.go`: `FavoritRepo: favoritRepo` zum bestehenden `tq.Query`-Struct hinzufügen (QueryHandler-Wiring — neues Feld, muss analog zu `tq.Query = tableApp.Query{...}` erweitert werden).
- [ ] In `backend/api/service.go`: Drei neue Routen für Favoriten registrieren: `r.HandleFunc("/favorit-hinzufuegen", tc.FavoritHinzufuegenHandler())`, `r.HandleFunc("/favorit-entfernen", tc.FavoritEntfernenHandler())`, `r.HandleFunc("/get-aktive-tische-mit-favoriten", tq.GetAktiveTischeMitFavoritenHandler())`.
- [ ] In `backend/api/service.go`: Neue Route für Batch-State registrieren: `r.HandleFunc("/get-meine-tische-state", tq.GetMeineTischeStateHandler())`.
- [ ] In `backend/api/service.go`: Reporting-Query-Handler instanziieren und verdrahten: `reportingRepo := reporting_repo.NewRepository(db)`, `rq := reportingHTTP.QueryHandler{}`, `rq.Query = reportingApp.Query{ReportingRepo: reportingRepo}`. Route: `r.HandleFunc("/get-eigene-uebersicht", rq.GetEigeneUebersichtHandler())`.
- [ ] `make lint` und `make test` ausführen — sicherstellen, dass keine Fehler auftreten. Falls bestehende Tests angepasst werden müssen (z.B. Command/Query-Struct-Initialisierungen), diese anpassen.

---

## Abschnitt 8: Frontend — Typen, Schemas & Backend-Klassen

Kontext:

- `frontend/src/service/table/Tisch.ts:1-26` — Bestehende Zod-Schemas: `TischSchema` (id, name, saldoCents), `TischStateSchema` (tischId, tischName, saldoCents, unbezahltePositionen, ausstehendePositionen, gesamtZahlungenCents)
- `frontend/src/service/table/TischBackend.ts:1-100` — Bestehende Backend-Klasse: `getAktiveTische()` → TischSchema, `getTischState()` → TischStateSchema. Pattern: Zod-Schema definieren → `backend.post(endpoint, body, schema)` → typed result
- `frontend/src/lib/Backend.ts:1-116` — `BackendClient`-Interface: `post<TResponse>(endpoint, body, schema?)` → validated response

- [ ] In `frontend/src/service/table/Tisch.ts`: Neues Schema `AktiverTischMitFavoritSchema` hinzufügen — `z.object({ id: TischIdSchema, name: z.string(), saldoCents: z.number().int(), istFavorit: z.boolean() })`. Type exportieren: `export type AktiverTischMitFavorit = z.infer<typeof AktiverTischMitFavoritSchema>`.
- [ ] In `frontend/src/service/table/Tisch.ts`: Neues Schema `EigeneUebersichtSchema` hinzufügen — `z.object({ anzahlBestellungen: z.number().int(), bestellungenCents: z.number().int(), anzahlZahlungen: z.number().int(), zahlungenCents: z.number().int() })`. Type exportieren: `export type EigeneUebersicht = z.infer<typeof EigeneUebersichtSchema>`.
- [ ] In `frontend/src/service/table/TischBackend.ts`: Neue Methode `favoritHinzufuegen(tischId: number): Promise<void>` — `const body = z.object({ tischId: TischIdSchema }).parse({ tischId })`, `await this.backend.post('service/favorit-hinzufuegen', body)`.
- [ ] In `frontend/src/service/table/TischBackend.ts`: Neue Methode `favoritEntfernen(tischId: number): Promise<void>` — analog zu `favoritHinzufuegen`, Endpoint: `service/favorit-entfernen`.
- [ ] In `frontend/src/service/table/TischBackend.ts`: Neue Methode `getAktiveTischeMitFavoriten(): Promise<AktiverTischMitFavorit[]>` — `backend.post('service/get-aktive-tische-mit-favoriten', {}, z.object({ tische: z.array(AktiverTischMitFavoritSchema) }))` → return `tische`. Import `AktiverTischMitFavorit` und Schema.
- [ ] In `frontend/src/service/table/TischBackend.ts`: Neue Methode `getMeineTischeState(): Promise<TischState[]>` — `backend.post('service/get-meine-tische-state', {}, z.object({ tische: z.array(TischStateSchema) }))` → return `tische`.
- [ ] In `frontend/src/service/table/TischBackend.ts`: Neue Methode `getEigeneUebersicht(): Promise<EigeneUebersicht>` — `backend.post('service/get-eigene-uebersicht', {}, EigeneUebersichtSchema)`. Import `EigeneUebersicht` und Schema.

---

## Abschnitt 9: Frontend — Hooks

Kontext:

- `frontend/src/service/table/hooks.ts:1-48` — Bestehende Hooks: `useAktiveTische`, `useTischHistorie`, `useTischState`. Pattern: `const tischBackend = new TischBackend(BackendSingleton)` Singleton → `useFetch(() => tischBackend.method(), initialData, deps?)`.
- `frontend/src/lib/useFetch.ts:1-55` — `useFetch<T>(fetcher, initialData, deps?)` → `{ loading, data, error, reload, setData }`

- [ ] In `frontend/src/service/table/hooks.ts`: Neuen Hook `useAktiveTischeMitFavoriten()` — `useFetch(() => tischBackend.getAktiveTischeMitFavoriten(), [] as AktiverTischMitFavorit[])`. Return: `{ ...rest, tische }`. Import `AktiverTischMitFavorit`.
- [ ] In `frontend/src/service/table/hooks.ts`: Neuen Hook `useMeineTischeState()` — `useFetch(() => tischBackend.getMeineTischeState(), [] as TischState[])`. Return: `{ ...rest, tische }`.
- [ ] In `frontend/src/service/table/hooks.ts`: Neuen Hook `useEigeneUebersicht()` — `useFetch(() => tischBackend.getEigeneUebersicht(), { anzahlBestellungen: 0, bestellungenCents: 0, anzahlZahlungen: 0, zahlungenCents: 0 } as EigeneUebersicht)`. Return: `{ ...rest, uebersicht }`. Import `EigeneUebersicht`.

---

## Abschnitt 10: Frontend — Komponenten & Dashboard-Seite

Kontext:

- `frontend/src/service/TableSelectionPage.tsx:1-80` — Aktuelle Seite: `useAktiveTische()` → Grid mit `Item`-Komponenten pro Tisch
- `frontend/src/service/ServiceLayout.tsx:1-33` — Header mit "Tischauswahl"-Text, Back-Button auf Detail-Seite
- `frontend/src/service/components/` — Bestehende Drawer-Komponenten als Referenz
- `frontend/src/admin/reporting/ReportingResults.tsx:25-50` — `SummaryCard`-Pattern (Card + CardHeader + CardTitle + CardContent)
- `frontend/src/components/common/EmptyState.tsx:1-40` — EmptyState-Komponente (icon + title + description + action)
- `frontend/src/components/ui/card.tsx` — Card, CardHeader, CardTitle, CardContent
- `frontend/src/components/ui/badge.tsx` — Badge mit variants
- `frontend/src/components/ui/drawer.tsx` — Drawer (Vaul) Komponente
- `frontend/src/components/ui/input.tsx` — Input-Komponente
- `frontend/src/routes.ts:1-95` — Route-Konfiguration (ServiceLayout → TableSelectionPage Route)
- `frontend/src/lib/utils.ts:1-14` — `formatCents()` und `cn()`

- [ ] Neue Datei `frontend/src/service/components/EigeneUebersicht.tsx` erstellen — Komponente `EigeneUebersicht` mit Props `{ uebersicht: EigeneUebersicht, loading: boolean }`. Layout: 2-Spalten-Grid mit je einer `Card`-Komponente. Karte 1: "Bestellungen" — Anzahl als Hauptwert, Summe in Euro als Sub-Text (via `formatCents`). Karte 2: "Kassiert" — Anzahl als Hauptwert, Summe in Euro als Sub-Text. Pattern analog zu `SummaryCard` aus `ReportingResults.tsx`. Bei `loading`: Skeleton-Platzhalter.
- [ ] Neue Datei `frontend/src/service/components/MeinTischCard.tsx` erstellen — Komponente `MeinTischCard` mit Props `{ state: TischState }`. Auf Tap → `navigate('/service/tables/${state.tischId}')`. Verwendung von `Card` als Container. Anzeige: (1) Tischname + Saldo (rot via `text-destructive` wenn negativ). (2) Badge "X ausstehend" wenn `ausstehendePositionen.length > 0`. (3) Badge "X unbezahlt" wenn `unbezahltePositionen.length > 0`. (4) Badge variant="destructive" "Auszahlung: X €" wenn `saldoCents < 0`. (5) Wenn alles leer + saldo >= 0: grüner Status-Text "Alles erledigt". Geldbeträge via `formatCents()`.
- [ ] Neue Datei `frontend/src/service/components/TischAuswahlDrawer.tsx` erstellen — Komponente `TischAuswahlDrawer` mit Props `{ open: boolean, onOpenChange: (open: boolean) => void }`. Verwendet `Drawer` (Vaul) von unten. Inhalt: (1) Suchfeld (`Input`) oben im Drawer. (2) Liste aller aktiven Tische via `useAktiveTischeMitFavoriten()`-Hook. (3) Clientseitige Filterung: Tische filtern, deren Name den Suchtext enthält (case-insensitive). (4) Pro Tisch-Zeile: Stern-Toggle (★ gefüllt / ☆ leer) links, Tischname + Saldo rechts. Tap auf Stern → `tischBackend.favoritHinzufuegen()`/`.favoritEntfernen()` aufrufen + optimistisches UI-Update via `setData()` im Hook. Tap auf Tisch-Zeile → `navigate('/service/tables/${tisch.id}')` + Drawer schließen.
- [ ] `frontend/src/service/TableSelectionPage.tsx` komplett umbauen zu Service-Dashboard: (1) `useMeineTischeState()`-Hook für Favoriten-Cards. (2) `useEigeneUebersicht()`-Hook für KPI-Sektion. (3) State `drawerOpen` für Alle-Tische-Drawer. (4) Layout: Header "Meine Tische" (wird in ServiceLayout geändert), dann `EigeneUebersicht`-Komponente, dann Grid von `MeinTischCard`-Komponenten, dann "Alle Tische"-Button der Drawer öffnet. (5) Empty-State wenn keine Favoriten: `EmptyState`-Komponente mit Hinweis "Du hast noch keine Tische markiert" + Action-Button "Tische auswählen" der Drawer öffnet. (6) `TischAuswahlDrawer`-Komponente einbinden. (7) Loading-State: Skeleton analog zum bestehenden `TischListSkeleton`.
- [ ] In `frontend/src/service/ServiceLayout.tsx`: Header-Text von "Tischauswahl" auf "Meine Tische" ändern (sowohl im Link-Text als auch im `span`-Text für Non-Detail-Ansicht).
- [ ] `make lint-frontend` und `make test-frontend` ausführen — sicherstellen, dass keine Fehler oder Warnings auftreten. Bestehende Tests in `routes.test.ts` prüfen und ggf. anpassen falls sich Importe/Exports geändert haben.
