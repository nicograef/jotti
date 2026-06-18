# Plan: Bediener-Identität auf Username umstellen (Klarname minimieren)

> Source PRD: n/a (Task aus Datenschutz-/DSFinV-K-Analyse; korrigiert nebenbei
> [docs/prds/prd-tischservice-pro-servicekraft.md](../prds/prd-tischservice-pro-servicekraft.md))

## Goal

Überall den **Username** (statt des bürgerlichen Klarnamens) als angezeigte und
eingefrorene Akteur-Identität verwenden. Konkret:

- Jede Aktion friert den Username im Event ein (`kassenjournal.user_name`),
  fiskalisch wie operativ.
- Der DSFinV-K-Export liefert den Username als `BEDIENER_NAME` — laut Spec
  genügt der „unternehmensinterne Name" (DSFinV-K, Feld `BEDIENER_NAME`,
  Feldlänge 50: *„Unternehmensinterner Name der Person, die den Vorgang
  erfasst."*); der echte Vor-/Nachname ist nicht erforderlich.
- Der Klarname (`users.name`) bleibt **Pflicht**, wird aber nur lokal in der
  Benutzerverwaltung verwendet und **nie eingefroren**.
- Gegenmittel zu den Nachteilen sind integriert: Username global eindeutig
  (kein Recycling), Klarname-Lesbarkeit in Admin-Auswertungen per Live-Join,
  Dokumentation der Bediener-Zuordnung.

Datenschutz-Wirkung: Der bürgerliche Klarname wird nicht mehr 10 Jahre
unveränderlich im Journal und im Finanzamt-Export festgeschrieben
(Datenminimierung, Art. 5 Abs. 1 lit. c / Art. 25 DSGVO), ohne die
GoBD-/DSFinV-K-Konformität zu verlieren.

## Architectural decisions

Durchgängig gültige Entscheidungen:

- **Identitätsmodell:**
  - `user_id` (numerisch, FK, stabil) bleibt der **harte Schlüssel** →
    DSFinV-K `BEDIENER_ID`.
  - `kassenjournal.user_name` hält ab jetzt den **Username** (eingefroren zum
    Event-Zeitpunkt) → DSFinV-K `BEDIENER_NAME`, Reporting-/Service-Anzeige.
  - `users.name` (Klarname) bleibt **NOT NULL / Pflicht**, mutabel, nur in der
    Benutzerverwaltung sichtbar; nie eingefroren, **nie im Export**.
- **JWT:** Claim `name` → `username` umbenannt, Wert = `u.Username`. Das
  Frontend-Token-Schema liest diesen Claim ohnehin nicht
  ([frontend/src/lib/Auth.ts:4-9](../../frontend/src/lib/Auth.ts)), daher kein
  Frontend-Bruch beim Login.
- **Schema:**
  - `users.username` wird **global eindeutig** (Unique-Index ohne
    `WHERE status != 'deleted'`) → ein Username ist dauerhaft genau einer
    Person zugeordnet, kein Recycling.
  - `kassenjournal.user_name` wird **nicht umbenannt**; nur der Spalten-COMMENT
    wird angepasst.
- **Anzeige:** Admin-Auswertungen (Reporting, Live-Dashboard) zeigen
  `username (Klarname)`; der Klarname wird **live** via `LEFT JOIN users ON
  id = user_id` aufgelöst. Service-Sichten und DSFinV-K: nur Username.
- **Keine Datenmigration:** Es gibt keine produktiven Instanzen; die DB wird
  neu aufgebaut, der Seed wird angepasst.

## Inventory

Quelle der Identität (ein einziger Umschalt-Punkt, der Rest fließt automatisch):

- [backend/api/auth/application/command.go:54](../../backend/api/auth/application/command.go) —
  `GenerateJWTTokenForUser(u.ID, u.Name, …)`: hier wird der Klarname ins Token
  gelegt. **Umschalt-Punkt → `u.Username`.**
- [backend/domain/jwt/jwt.go:18](../../backend/domain/jwt/jwt.go) — Claim
  `"name": userName` (Erzeugung); [jwt.go:47](../../backend/domain/jwt/jwt.go) —
  `claims["name"]` (Parsen). Claim umbenennen.
- [backend/api/middleware/middleware.go:24](../../backend/api/middleware/middleware.go) —
  `UserNameKey = "username"` (Kontext-Key heißt bereits „username");
  [middleware.go:189,205-206](../../backend/api/middleware/middleware.go) —
  füllt den Kontext aus dem JWT. **Keine Änderung nötig.**
- Alle Event-Schreiber lesen `userName` aus dem Kontext, daher automatisch
  abgedeckt:
  [kasse command_handler.go:72,98,128](../../backend/api/kasse/http/command_handler.go),
  [direktverkauf command_handler.go:66,128](../../backend/api/direktverkauf/http/command_handler.go),
  [table command_handler.go:311](../../backend/api/table/http/command_handler.go).
  Kein Pfad schreibt `user.Name` am JWT vorbei in Events (verifiziert).

Schema & Persistenz:

- [database/migrations/01_initial.up.sql:8-23](../../database/migrations/01_initial.up.sql) —
  `users` (name, username) + `idx_users_username_active` (erlaubt heute
  Recycling nach Soft-Delete).
- [database/migrations/01_initial.up.sql:139-150,184](../../database/migrations/01_initial.up.sql) —
  `kassenjournal.user_name` Spalte + COMMENT.
- [backend/domain/user/user.go:48,50,67,87](../../backend/domain/user/user.go) —
  `NameSchema` (min 3, max 50, Pflicht), `UsernameSchema` (`^[a-z0-9]{3,20}$`),
  `NewUser(name, username, role)`.

DSFinV-K-Export:

- [backend/domain/dsfinvk/mapper.go:216-217](../../backend/domain/dsfinvk/mapper.go)
  (und weitere Belegtypen) — `bedienerID: ev.UserID`, `bedienerName:
  ev.UserName`. Nach dem Umschalten ist `ev.UserName` der Username; **kein
  Code-Change**, aber Testfixtures anpassen.

Reporting / Admin-Auswertungen (zeigen heute `userName`):

- [backend/sqlc/queries/reporting.sql:53-66](../../backend/sqlc/queries/reporting.sql) —
  `GetUmsatzProServicekraft`: `MAX(user_name)` GROUP BY `user_id`.
- [backend/sqlc/queries/reporting.sql:70-83](../../backend/sqlc/queries/reporting.sql) —
  `GetStornierungen`: `e.user_name`.
- [backend/repository/reporting_repo/repo.go:220-221,260-261](../../backend/repository/reporting_repo/repo.go),
  [backend/api/reporting/http/query_handler.go:78,112,121,199](../../backend/api/reporting/http/query_handler.go) —
  DTO-Kette `UserName`.
- [frontend/src/admin/reporting/types.ts:17,37](../../frontend/src/admin/reporting/types.ts),
  [ReportingResults.tsx:184,286](../../frontend/src/admin/reporting/ReportingResults.tsx),
  [LiveReportingSection.tsx:188,288](../../frontend/src/admin/reporting/LiveReportingSection.tsx) —
  4 Admin-Anzeigestellen.

Seed:

- [backend/seed/engine.go:519-525](../../backend/seed/engine.go) — `benutzerName`
  liest aus `b.benutzer[userID]`; die Map muss Usernames liefern, damit
  geseedete Events Usernames einfrieren.

Frontend Benutzerverwaltung (Klarname bleibt, unverändert):

- [frontend/src/admin/users/NewUserDialog.tsx:59-60](../../frontend/src/admin/users/NewUserDialog.tsx),
  [UserItem.tsx:88,92](../../frontend/src/admin/users/UserItem.tsx),
  [EditUserDialog.tsx:97](../../frontend/src/admin/users/EditUserDialog.tsx) —
  zeigen name + username; **keine Änderung**.

PRD-Korrektur:

- [docs/prds/prd-tischservice-pro-servicekraft.md:134-137](../prds/prd-tischservice-pro-servicekraft.md) (US 23),
  [146-150](../prds/prd-tischservice-pro-servicekraft.md) (§Zuordnung),
  [206-207](../prds/prd-tischservice-pro-servicekraft.md) (§Stornieren),
  [288-291](../prds/prd-tischservice-pro-servicekraft.md) (Further Notes) —
  innerer Widerspruch „eingefroren vs. aus Stammdaten aufgelöst" + falsche
  Reporting-Aussage.

## Resolved decisions

- Scope = Username-Umstellung end-to-end **+** PRD-Textkorrektur; die
  Besteller-Funktion selbst bleibt ihrem eigenen Plan.
- Recycling aus: Username **global eindeutig** inkl. `deleted`.
- Admin-Auswertungen zeigen `username (Klarname)` (Live-Join); Service-Sichten
  und DSFinV-K nur Username.
- Klarname `users.name` bleibt **Pflicht**, nur Benutzerverwaltung, nie
  eingefroren.
- `kassenjournal.user_name` **nicht** umbenennen, nur COMMENT.
- `BEDIENER_ID` = `user_id`, `BEDIENER_NAME` = Username (kein Klarname im
  Export).
- Keine Datenmigration (keine Prod-Instanzen); Seed wird angepasst.
- **`UsernameSchema` bleibt unverändert** (`^[a-z0-9]{3,20}$`): kein Uppercase,
  keine Unterstriche/Separatoren, kein Unicode/Umlaute, keine Leerzeichen.
  Begründung: vermeidet Case-Folding beim Login und Homoglyph-/Confusable-
  Mehrdeutigkeit, die direkt gegen die globale Eindeutigkeit aus Phase 2
  arbeiten würde; Lesbarkeit wird stattdessen über den `(Klarname)`-Zusatz
  gelöst.

## Open questions / Risks

- **Username-Lesbarkeit** (`^[a-z0-9]{3,20}$`, keine Umlaute/Leerzeichen):
  `anna` ist gut, `hans12` weniger. Bewusst akzeptiert; gemildert durch den
  `(Klarname)`-Zusatz in Admin-Auswertungen (für Service-Sichten bleibt es beim
  Username) und eine empfohlene Username-Konvention (z. B. Vorname + Initial),
  die nicht erzwungen wird.
- **Globale Eindeutigkeit blockiert Usernames dauerhaft** (auch ausgeschiedener
  Helfer). Bewusst akzeptiert als Preis für eindeutige Bediener-Historie.

---

## Phase 1: Username an der Quelle einfrieren (Kern-Slice)

### Context

- [backend/api/auth/application/command.go:54](../../backend/api/auth/application/command.go) —
  Umschalt-Punkt `u.Name` → `u.Username`.
- [backend/domain/jwt/jwt.go:18,47](../../backend/domain/jwt/jwt.go) — Claim
  `name` → `username` (Erzeugung + Parsen).
- [database/migrations/01_initial.up.sql:184](../../database/migrations/01_initial.up.sql) —
  COMMENT von `kassenjournal.user_name`.
- [backend/seed/engine.go:519-525](../../backend/seed/engine.go) —
  `benutzerName`/`b.benutzer`-Map muss Usernames liefern.
- [backend/domain/dsfinvk/mapper.go:216-217](../../backend/domain/dsfinvk/mapper.go) —
  `bedienerName` ist danach der Username (kein Code-Change, Fixtures anpassen).

### What to build

Die eingefrorene Akteur-Identität wird vom Klarnamen auf den Username
umgestellt. Ein angemeldeter Benutzer löst eine Kassenaktion aus; das Event
friert ab jetzt den Username in `kassenjournal.user_name` ein. Der
DSFinV-K-Export zeigt den Username als `BEDIENER_NAME` (bei unverändertem
`BEDIENER_ID = user_id`), das bestehende Reporting zeigt den Username. Der
JWT-Claim wird auf `username` umbenannt und mit `u.Username` befüllt. Der
COMMENT der Journalspalte wird auf „Username des Akteurs zum Event-Zeitpunkt"
aktualisiert. Der Seed erzeugt Events mit Usernames.

### Acceptance criteria

- [x] Nach Login enthält das JWT den Claim `username` mit dem Username; kein
      Klarname mehr im Token.
- [x] Eine über die API ausgelöste Kassenaktion schreibt den Username in
      `kassenjournal.user_name`.
- [x] Der DSFinV-K-Export zeigt `BEDIENER_NAME` = Username und `BEDIENER_ID` =
      `user_id`; kein Klarname im Export.
- [x] Das bestehende Umsatz-/Storno-Reporting zeigt den Username.
- [x] Seed-Lauf erzeugt Events mit Usernames; alle Backend-Tests grün
      (jwt-, mapper-, reporting-, seed-Tests auf Username aktualisiert).

---

## Phase 2: Eindeutigkeit härten (kein Username-Recycling)

### Context

- [database/migrations/01_initial.up.sql:22-23](../../database/migrations/01_initial.up.sql) —
  `idx_users_username_active` (heute nur `WHERE status != 'deleted'`).
- [backend/api/user/application/command.go:32-41](../../backend/api/user/application/command.go) —
  `CreateUser` mappt `db.ErrAlreadyExists` → `ErrUsernameAlreadyExists`.
- [backend/repository/user_repo/repo_test.go](../../backend/repository/user_repo/repo_test.go) —
  Tests zur Username-Eindeutigkeit / Wiederverwendung nach Soft-Delete.

### What to build

Der Username wird global eindeutig — auch über soft-gelöschte Benutzer hinweg.
Der partielle Unique-Index wird durch einen vollständigen Unique-Index (bzw.
Constraint) ersetzt. Anlegen oder Umbenennen auf einen bereits (auch von einem
gelöschten Benutzer) belegten Username schlägt mit `ErrUsernameAlreadyExists`
fehl. Damit ist der eingefrorene Username in der Journal-Historie dauerhaft
genau einer `user_id` zugeordnet.

### Acceptance criteria

- [x] Ein Username, der von einem aktiven **oder gelöschten** Benutzer belegt
      ist, kann nicht erneut vergeben werden.
- [x] Der bestehende „Username nach Soft-Delete wiederverwendbar"-Test ist auf
      das neue Verhalten umgestellt und grün.
- [x] Kein eingefrorener `user_name` im Journal kann über die Zeit auf zwei
      verschiedene `user_id` zeigen.

---

## Phase 3: Klarname als Pflicht-Label + Lesbarkeit in Admin-Auswertungen

### Context

- [backend/domain/user/user.go:67,87-106](../../backend/domain/user/user.go) —
  `users.name` ist bereits Pflicht (`NameSchema.Required()`); verifizieren,
  nicht aufweichen.
- [backend/sqlc/queries/reporting.sql:53-83](../../backend/sqlc/queries/reporting.sql) —
  `GetUmsatzProServicekraft` / `GetStornierungen`: `LEFT JOIN users` für den
  live aufgelösten Klarnamen ergänzen.
- [backend/repository/reporting_repo/repo.go:220-261](../../backend/repository/reporting_repo/repo.go),
  [backend/api/reporting/http/query_handler.go:78-199](../../backend/api/reporting/http/query_handler.go) —
  DTOs um `name` (Klarname) ergänzen.
- [frontend/src/admin/reporting/types.ts:17,37](../../frontend/src/admin/reporting/types.ts),
  [ReportingResults.tsx:184,286](../../frontend/src/admin/reporting/ReportingResults.tsx),
  [LiveReportingSection.tsx:188,288](../../frontend/src/admin/reporting/LiveReportingSection.tsx) —
  Anzeige `username (Klarname)`.

### What to build

Der Klarname bleibt Pflichtfeld und ausschließlich in der Benutzerverwaltung
sichtbar. Für die Lesbarkeit lösen die **Admin-Auswertungen** (Umsatz pro
Servicekraft, Stornierungen, Live-Dashboard) den aktuellen Klarnamen live über
`user_id` aus `users` auf und zeigen `username (Klarname)`. Der eingefrorene
Username bleibt der maßgebliche Wert; der Klarname ist reine Anzeige-Ergänzung.
Service-Sichten und der DSFinV-K-Export bleiben Username-only.

### Acceptance criteria

- [x] `users.name` ist weiterhin `NOT NULL` und im Anlage-/Bearbeiten-Formular
      Pflicht.
- [x] Umsatz-pro-Servicekraft, Stornierungs-Liste und Live-Dashboard zeigen
      `username (Klarname)`, mit live aus `users` aufgelöstem Klarnamen.
- [x] Service-Sichten zeigen nur den Username; der DSFinV-K-Export enthält
      **keinen** Klarnamen.
- [x] Reporting-Tests prüfen die `username (Klarname)`-Auflösung inkl. eines
      soft-gelöschten Benutzers (Klarname bleibt auflösbar).

---

## Phase 4: Dokumentation (PRD-Korrektur + Compliance)

### Context

- [docs/prds/prd-tischservice-pro-servicekraft.md:134-137,146-150,206-207,288-291](../prds/prd-tischservice-pro-servicekraft.md) —
  Widerspruch + falsche Reporting-Aussage.
- [docs/compliance.md §2.5/§6](../compliance.md),
  [docs/verfahrensdokumentation.md](../verfahrensdokumentation.md),
  [docs/steuerrecht.md](../steuerrecht.md) — Bediener-/DSFinV-K-Bezüge.

### What to build

Die Tischservice-PRD wird widerspruchsfrei gemacht: Der Besteller ist der
Umschlag-Akteur, dessen **Username** ohnehin im Event eingefroren ist; kein
neues Namensfeld in `PositionEventData`. Die Sätze „Name wird aus den
Stammdaten aufgelöst" (US 23 / §Stornieren) und „wie bei den fetten Events …
festgehalten" (Further Notes) werden auf eine konsistente Aussage
zusammengeführt; die falsche Behauptung „Reporting nutzt den zuletzt bekannten
Namen" wird korrigiert (Reporting nutzt den eingefrorenen Username; der
Klarname wird in Admin-Auswertungen live aufgelöst).

Die Compliance-Doku hält fest: `BEDIENER_NAME` = Username erfüllt den
„unternehmensinternen Namen" der DSFinV-K; der Klarname wird bewusst nicht
exportiert (Minimierung). Die Verfahrensdokumentation vermerkt, dass die
Zuordnung Username → Klarname in der Benutzerverwaltung (`users.name`) als
Bedienerliste für die Betriebsprüfung dient.

### Acceptance criteria

- [x] Tischservice-PRD enthält keinen Widerspruch mehr zwischen „eingefroren"
      und „live aufgelöst"; US 23 ist auf „Username im Umschlag bereits
      eingefroren, kein neues Feld in `PositionEventData`" umformuliert.
- [x] Die falsche Reporting-Aussage ist korrigiert.
- [x] `compliance.md`/`steuerrecht.md` begründen `BEDIENER_NAME` = Username
      anhand der DSFinV-K-Felddefinition.
- [x] `verfahrensdokumentation.md` nennt `users.name` als Bedienerliste
      (Username → Person) und empfiehlt eine Username-Konvention.
