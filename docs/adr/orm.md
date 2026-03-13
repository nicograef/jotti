# ADR: Datenbankzugriff — ORM-Evaluation

## Status

**Überprüft und bestätigt** — 13. März 2026

## 1. Kontext

### 1.1 Systemüberblick

jotti ist ein Mobile-Point-of-Sale-System für temporäre Gastronomie-Veranstaltungen gemeinnütziger Organisationen. Das Backend ist in Go implementiert und nutzt PostgreSQL 17 als einzige Datenbank.

### 1.2 Hybrides Persistenzmodell

jotti kombiniert bewusst **zwei unterschiedliche Persistenzstrategien** in einer PostgreSQL-Instanz:

| Bounded Context       | Persistenz         | Tabellen                                           | Zugriffsmuster                               |
| --------------------- | ------------------ | -------------------------------------------------- | -------------------------------------------- |
| **Kassenbetrieb**     | Event-Sourcing     | `events` (append-only)                             | INSERT + SELECT, JSONB-Payloads, OCC         |
| **Stammdaten**        | CRUD + Soft-Delete | `users`, `tische`, `produkte`, `produkt_varianten` | Standard-CRUD, Enum-Filter, JSON-Aggregation |
| **Ausgabe** (MVP2)    | Event-getrieben    | Projektionen über `events`                         | LISTEN/NOTIFY, Read-only                     |
| **Abrechnung** (MVP2) | Read-only          | Materialized Views                                 | SQL-Aggregationen, Window Functions          |

Diese Hybridität ist die **zentrale Constraint** für die Toolwahl: Es gibt keine Gesamtlösung „ein ORM für alles".

### 1.3 Architektur-Constraints

Die folgenden Eigenschaften sind im Entwickler-Handbuch (§3, §4, §6) festgelegt und nicht verhandelbar:

1. **Fat Events (JSONB):** Produktdaten (Name, Preis, Kategorie) werden zum Bestellzeitpunkt im Event eingefroren. Das Event enthält alles, was Consumer brauchen — keine JOINs gegen Stammdaten nötig. Spätere Preisänderungen haben keinen Einfluss auf historische Bestellungen (Anti-Corruption Layer).
2. **Append-Only Events:** DB-Trigger verhindern UPDATE, DELETE und TRUNCATE auf der `events`-Tabelle. Events sind immutable.
3. **OCC per INSERT:** `UNIQUE(subject, version)` — Optimistic Concurrency Control erfolgt über eine aufsteigende Versionsnummer pro Tisch-Aggregat. Bei Konflikt schlägt der INSERT fehl, die Applikation löst einen Retry aus.
4. **State-Rekonstruktion per Replay:** Der Tisch-Zustand wird nicht aus der Datenbank gelesen, sondern aus dem Event Stream berechnet: `state = fold(apply, events)`. Snapshots optimieren die Ladezeit.
5. **Soft-Deletes:** CRUD-Tabellen nutzen `status = 'deleted'` statt physischer Löschung — historische Events bleiben referenziell konsistent.
6. **PostgreSQL-spezifische Features:** Custom Enums (`UserRole`, `EntityStatus`, `ProduktKategorie`), JSONB mit `json_agg()` und `json_build_object()`, CTEs, Trigger.
7. **Geldbeträge in Cent (int):** Niemals Floats für Geldbeträge.
8. **Pre-Release-Phase:** Schema ändert sich direkt in `database/migrations/01_initial.up.sql`. Keine inkrementellen Migrationsdateien, keine Down-Migrationen. Schema-Drift-Schutz ist essentiell.

### 1.4 Ist-Zustand (gemessen am Code)

| Metrik                          | Wert                                           |
| ------------------------------- | ---------------------------------------------- |
| Generierter Code (sqlc)         | ~1.200 LOC (`backend/sqlc/dbgen/`)             |
| Repository-Wrapper              | ~450 LOC (`backend/repository/`)               |
| SQL-Queries                     | ~215 LOC, 27 Queries (`backend/sqlc/queries/`) |
| Mapping-Boilerplate (db→domain) | ~150 LOC                                       |
| JSON-Unpacking (Varianten)      | ~50 LOC (`product_repo/types.go`)              |
| Mock-Implementierungen          | 4 Dateien                                      |
| Domain-Modelle                  | ~400 LOC (`backend/domain/`)                   |
| Fehler-Mapping                  | ~50 LOC (`backend/db/db.go`)                   |

---

## 2. Bewertete Alternativen — Fünf Abstraktionsstufen

### 2.1 Stufe 1: Raw SQL (`database/sql` + `pgx/v5`)

**Ansatz:** Handgeschriebene SQL-Strings als Go-Konstanten, manuelles `row.Scan()` für jede Query.

**So sieht das aus:**

```go
const findUserByID = `SELECT id, name, username, role, status FROM users WHERE id = $1`

func (r *repo) FindByID(ctx context.Context, id int) (User, error) {
    row := r.db.QueryRowContext(ctx, findUserByID, id)
    var u User
    err := row.Scan(&u.ID, &u.Name, &u.Username, &u.Role, &u.Status)
    if err != nil {
        return User{}, mapError(err)
    }
    return u, nil
}
```

| Dimension            | Bewertung | Begründung                                                         |
| -------------------- | --------- | ------------------------------------------------------------------ |
| Event-Sourcing       | ✅ 5/5    | Volle Kontrolle über INSERT/SELECT, JSONB, OCC                     |
| CRUD Stammdaten      | ⚠️ 2/5    | Funktional, aber ~30 LOC pro Query (SQL + Scan + Error)            |
| Compile-Time-Safety  | ❌ 1/5    | Keine Validierung gegen Schema — Spalte umbenannt → Runtime-Fehler |
| Schema-Drift-Schutz  | ❌ 1/5    | Inkompatibilitäten erst zur Laufzeit oder in Tests sichtbar        |
| PostgreSQL-Features  | ✅ 5/5    | Alles direkt nutzbar                                               |
| Testbarkeit          | ✅ 5/5    | Interface-basiert, gleiche Mock-Strategie                          |
| Runtime-Dependencies | ✅ 5/5    | Nur pgx (bereits vorhanden)                                        |
| Lernkurve            | ✅ 5/5    | Jeder Go-Entwickler kennt `database/sql`                           |

**Stärken:** Maximale Kontrolle, keine Abstraktions-Lecks, null zusätzliche Dependencies.

**Schwächen:** ~30% mehr Boilerplate als sqlc, kein Schema-Drift-Schutz. Bei jeder Schema-Änderung müssen alle betroffenen `Scan()`-Aufrufe manuell angepasst werden — vergessene Stellen führen zu Runtime-Fehlern.

**Fazit für jotti:** War der Status quo vor sqlc. Für ein Projekt in aktiver Schema-Entwicklung (Pre-Release, Schema ändert sich in `01_initial.up.sql`) ist das fehlende Compile-Time-Feedback ein reales Risiko.

---

### 2.2 Stufe 2: SQL + Runtime-Reflection (`sqlx`)

**Ansatz:** SQL bleibt handgeschrieben, aber `sqlx.StructScan()` eliminiert manuelles `row.Scan()` via Struct-Tags.

**So sieht das aus:**

```go
type User struct {
    ID       int    `db:"id"`
    Name     string `db:"name"`
    Username string `db:"username"`
    Role     string `db:"role"`
    Status   string `db:"status"`
}

func (r *repo) FindByID(ctx context.Context, id int) (User, error) {
    var u User
    err := sqlx.GetContext(ctx, r.db, &u,
        `SELECT id, name, username, role, status FROM users WHERE id = $1`, id)
    return u, mapError(err)
}
```

| Dimension            | Bewertung | Begründung                                                    |
| -------------------- | --------- | ------------------------------------------------------------- |
| Event-Sourcing       | ✅ 5/5    | Gleich wie Raw SQL — volle SQL-Kontrolle                      |
| CRUD Stammdaten      | ⚠️ 3/5    | Scan-Boilerplate eliminiert, SQL-Strings bleiben manuell      |
| Compile-Time-Safety  | ❌ 1/5    | Struct-Tag-Mapping über Reflection — Fehler erst zur Laufzeit |
| Schema-Drift-Schutz  | ⚠️ 2/5    | Struct-Tags können veralten, kein Build-Fehler                |
| PostgreSQL-Features  | ✅ 5/5    | Alles direkt nutzbar                                          |
| Testbarkeit          | ✅ 5/5    | Gleich wie Raw SQL                                            |
| Runtime-Dependencies | ⚠️ 4/5    | +1 Dependency (`jmoiron/sqlx`)                                |
| Lernkurve            | ✅ 5/5    | Minimal — Struct-Tags plus SQL                                |

**Stärken:** Kleiner, pragmatischer Schritt von Raw SQL. Eliminiert `Scan()`-Boilerplate. Gut geeignet für Projekte, die bewusst SQL schreiben wollen.

**Schwächen:** Löst das Kern-Problem (Schema-Drift) nicht. Reflection-basiertes Mapping fängt Fehler nicht beim Build, sondern erst in Tests oder Produktion. Für ein Projekt mit häufigen Schema-Änderungen problematisch.

**Fazit für jotti:** Inkrementelle Verbesserung über Raw SQL, aber unzureichend für die Pre-Release-Phase. Wenn das Schema selten ändert, wäre sqlx eine solide Wahl — für jotti aktuell nicht.

---

### 2.3 Stufe 3: SQL-first + Code-Generierung (`sqlc`) ← aktueller Stand

**Ansatz:** SQL-Queries in `.sql`-Dateien mit Annotationen. sqlc parst die Queries gegen das Migrationsschema und generiert typsichere Go-Funktionen + Structs.

**So sieht das aus:**

```sql
-- backend/sqlc/queries/users.sql
-- name: GetUser :one
SELECT id, name, username, role, status, password_hash,
       onetime_password_hash, created_at, updated_at
FROM users WHERE id = $1 AND status != 'deleted';
```

```go
// Generiert in backend/sqlc/dbgen/users.sql.go (NICHT EDITIEREN)
func (q *Queries) GetUser(ctx context.Context, id int) (GetUserRow, error) { ... }
```

```go
// Repository-Wrapper in backend/repository/user_repo/repo.go
func (r Repository) GetUser(ctx context.Context, id int) (user.User, error) {
    row, err := r.q.GetUser(ctx, id)
    if err != nil {
        return user.User{}, db.Error(err)
    }
    return userRowToDomain(row), nil
}
```

| Dimension            | Bewertung | Begründung                                                      |
| -------------------- | --------- | --------------------------------------------------------------- |
| Event-Sourcing       | ✅ 5/5    | Volle SQL-Kontrolle, JSONB nativ, OCC trivial                   |
| CRUD Stammdaten      | ✅ 4/5    | Typsicher, kein Scan-Boilerplate, aber Mapping db→domain nötig  |
| Compile-Time-Safety  | ✅ 5/5    | `make sqlc` validiert Queries gegen Schema                      |
| Schema-Drift-Schutz  | ✅ 5/5    | Schema geändert → `make sqlc` schlägt fehl bei Inkompatibilität |
| PostgreSQL-Features  | ✅ 5/5    | CTEs, `json_agg()`, Window Functions direkt in SQL              |
| Testbarkeit          | ✅ 5/5    | Repository-Interface + Mock-Implementierungen                   |
| Runtime-Dependencies | ✅ 5/5    | 0 Runtime-Dependencies — sqlc ist reines Build-Tool             |
| Lernkurve            | ✅ 4/5    | SQL schreiben + `make sqlc` — Build-Step als einzige Hürde      |

**Konkrete Stärken im jotti-Code:**

- **Event-Store-Queries** funktionieren exzellent — `ReadEventsWithSnapshot` nutzt eine CTE, die automatisch den letzten Snapshot findet und alle folgenden Events lädt. Das wäre in keinem ORM oder Query Builder so elegant ausdrückbar.
- **Produkt-Queries** mit `json_agg(json_build_object(...))` aggregieren Varianten direkt in der DB-Query — ein einzelner Round-Trip statt N+1.
- **Schema-Drift-Schutz** ist das Killer-Feature für die Pre-Release-Phase: Schema in `01_initial.up.sql` ändern → `make sqlc` → Build-Fehler zeigen genau, welche Queries angepasst werden müssen.

**Konkrete Schwächen im jotti-Code:**

1. **Drei verschiedene User-Mapping-Funktionen** (`userRowToDomain`, `userByUsernameRowToDomain`, `allUsersRowToDomain`) in `backend/repository/user_repo/`, weil sqlc für jede Query mit unterschiedlichen SELECT-Spalten einen eigenen Result-Typ generiert.
2. **JSON-Unpacking-Boilerplate** (~50 LOC in `backend/repository/product_repo/types.go`) — `json_agg()` liefert `json.RawMessage`, die manuell in Domain-Typen gemappt werden muss.
3. **Build-Step vergessbar** — `make sqlc` muss nach jeder Query- oder Schema-Änderung ausgeführt werden. Mitigation: CI-Pipeline oder pre-commit Hook.
4. **Keine dynamischen Queries** — flexible Filter (z.B. „alle Produkte mit Status X oder Kategorie Y") brauchen separate SQL-Queries. Für jotti aktuell kein Problem (wenige, statische Filter).
5. **`sql_package: "database/sql"`** — sqlc generiert gegen `database/sql` statt direkt gegen pgx. Funktioniert über den pgx-stdlib-Adapter, aber verliert pgx-spezifische Features (z.B. direkte `pgtype`-Unterstützung).

**Fazit für jotti:** Die richtige Balance für das hybride Modell. Schwächen sind real, aber handhabbar. Keine davon ist architekturrelevant — sie betreffen Komfort, nicht Korrektheit.

---

### 2.4 Stufe 4: Query Builder mit Code-Generierung (`go-jet`, `bob`)

**Ansatz:** Code-Generierung aus dem DB-Schema erzeugt eine typsichere Go-DSL für Query-Konstruktion. Statt SQL zu schreiben, komponiert man Queries in Go.

**So sieht das aus (go-jet):**

```go
import . "github.com/go-jet/jet/v2/postgres"

stmt := SELECT(
    Users.ID, Users.Name, Users.Username, Users.Role, Users.Status,
).FROM(
    Users,
).WHERE(
    Users.ID.EQ(Int(42)).AND(Users.Status.NOT_EQ(String("deleted"))),
)
```

| Dimension            | Bewertung | Begründung                                                                 |
| -------------------- | --------- | -------------------------------------------------------------------------- |
| Event-Sourcing       | ⚠️ 3/5    | Möglich, aber DSL-Overhead für simple `INSERT INTO events ... RETURNING *` |
| CRUD Stammdaten      | ✅ 5/5    | Sehr komfortabel, dynamische Filter möglich, kein Mapping                  |
| Compile-Time-Safety  | ✅ 5/5    | Aus Schema generiert — Spaltenänderungen brechen den Build                 |
| Schema-Drift-Schutz  | ✅ 5/5    | Schema-Änderung → Regenerieren → Build-Fehler                              |
| PostgreSQL-Features  | ⚠️ 3/5    | CTEs und JSONB möglich, aber verbose; `json_agg()` in DSL umständlich      |
| Testbarkeit          | ⚠️ 3/5    | Query-Builder-DSL schwerer zu mocken als Interfaces                        |
| Runtime-Dependencies | ⚠️ 3/5    | +1 Runtime-Dependency mit signifikantem Dependency-Tree                    |
| Lernkurve            | ⚠️ 3/5    | Tool-spezifische DSL-Syntax lernen                                         |

**Stärken:** Kein SQL schreiben, dynamische Query-Komposition, kein Mapping-Boilerplate. Exzellent für CRUD-lastige Anwendungen mit vielen dynamischen Filtern.

**Schwächen für jotti:**

- **Event-Store-Queries werden umständlicher:** Die CTE in `ReadEventsWithSnapshot` (die den letzten Snapshot findet und alle Events seit diesem lädt) ist in nativem SQL 15 Zeilen — in einer Query-Builder-DSL wäre sie deutlich länger und schwerer lesbar.
- **`json_agg(json_build_object(...))`** ist in der DSL nicht nativ ausdrückbar — man fällt auf Raw SQL zurück, was den Zweck des Query Builders unterminiert.
- **Löst ein Problem, das jotti kaum hat:** Dynamische Filter kommen in jottis statischen Queries (alle Tische laden, alle aktiven Produkte laden) nicht vor.

**Fazit für jotti:** Für CRUD-lastige Anwendungen mit vielen dynamischen Filtern eine starke Option. Für jottis Hybrid-Architektur (50% Event-Sourcing, 50% einfaches CRUD) Overkill — es verschlechtert die Event-Store-Queries und verbessert die CRUD-Queries nur marginal.

---

### 2.5 Stufe 5: Full ORM

#### 2.5.1 GORM (Active Record, Convention-over-Configuration)

**Ansatz:** Go-Structs mit Tags definieren das Schema. GORM generiert SQL, verwaltet Migrationen, bietet Hooks und Scopes.

**So sieht das aus:**

```go
type User struct {
    gorm.Model              // ID, CreatedAt, UpdatedAt, DeletedAt
    Name     string
    Username string `gorm:"uniqueIndex"`
    Role     string
    Status   string
}

// CRUD
db.Create(&user)
db.First(&user, id)
db.Where("status != ?", "deleted").Find(&users)
db.Model(&user).Update("status", "deleted")
```

| Dimension            | Bewertung | Begründung                                                  |
| -------------------- | --------- | ----------------------------------------------------------- |
| Event-Sourcing       | ❌ 1/5    | **Fundamental inkompatibel** (siehe unten)                  |
| CRUD Stammdaten      | ✅ 5/5    | Minimal-Boilerplate, Auto-Migration, Preloading             |
| Compile-Time-Safety  | ❌ 1/5    | `interface{}`-basiert, Fehler erst zur Laufzeit             |
| Schema-Drift-Schutz  | ⚠️ 2/5    | `AutoMigrate` kann Schema undeterministisch driften lassen  |
| PostgreSQL-Features  | ❌ 2/5    | CTEs nur über `db.Raw()`, `json_agg` nicht nativ            |
| Testbarkeit          | ❌ 2/5    | Konkreter Typ (`*gorm.DB`) statt Interface — schwer mockbar |
| Runtime-Dependencies | ❌ 2/5    | +2 Runtime-Dependencies, signifikanter Tree                 |
| Lernkurve            | ⚠️ 3/5    | Einstieg einfach, Debugging von generiertem SQL schwierig   |

**Fundamentale Inkompatibilität mit Event-Sourcing:**

1. **Lifecycle-Hooks setzen mutable Records voraus.** GORMs `BeforeCreate`, `AfterUpdate`, `BeforeSave` gehen davon aus, dass Datensätze sich im Laufe ihres Lebens ändern. Events sind immutable — dieses Modell ist ein Widerspruch.
2. **Soft-Delete-Impedance:** GORMs `gorm.Model` nutzt `DeletedAt` (Nullable Timestamp). jotti nutzt `status`-Enum (`active`/`inactive`/`deleted`). Die Konventionen sind inkompatibel — man müsste GORMs Soft-Delete deaktivieren und alles manuell machen, was den ORM-Vorteil aufhebt.
3. **OCC-Mismatch:** GORMs Optimistic Locking funktioniert über `UPDATE ... WHERE version = $1` (UPDATE-basiert). jottis OCC funktioniert über `INSERT ... UNIQUE(subject, version)` (INSERT-basiert) — ein fundamental anderes Pattern.
4. **Fat Events als JSONB:** GORM würde versuchen, den JSONB-Payload in relationale Structs zu normalisieren. Fat Events enthalten bewusst denormalisierte Daten — das ist ein Anti-Pattern für jeden ORM.
5. **Append-Only-Trigger:** GORMs `Delete()`-Scope kollidiert mit dem DB-Trigger, der DELETE auf `events` verhindert. Das ist keine Konfigurationsfrage — es ist architekturelle Inkompatibilität.

**Fazit für jotti:** GORM wäre exzellent für eine reine CRUD-Anwendung. jottis Core Domain (Kassenbetrieb) ist Event-Sourced — GORM kann das nicht bedienen. Den Kassenbetrieb mit Raw SQL und die Stammdaten mit GORM zu machen, erzeugt zwei verschiedene Persistenz-Patterns im selben Projekt — mehr Komplexität, nicht weniger.

#### 2.5.2 ent (Graph-Schema, Meta-Programming)

**Ansatz:** Schema wird als Go-Code definiert (Entities, Edges, Fields). ent generiert daraus typsicheres CRUD, Queries und Migrationen.

**So sieht das aus:**

```go
// ent/schema/user.go
func (User) Fields() []ent.Field {
    return []ent.Field{
        field.String("name"),
        field.String("username").Unique(),
        field.Enum("role").Values("admin", "serviceleitung", "service"),
        field.Enum("status").Values("active", "inactive", "deleted"),
    }
}
```

| Dimension            | Bewertung | Begründung                                            |
| -------------------- | --------- | ----------------------------------------------------- |
| Event-Sourcing       | ❌ 1/5    | Graph-Relationen setzen mutable Nodes voraus          |
| CRUD Stammdaten      | ✅ 5/5    | Exzellent — typsicher, Edge-basiert, Traversals       |
| Compile-Time-Safety  | ✅ 5/5    | Code-Gen aus Schema-Definition                        |
| Schema-Drift-Schutz  | ✅ 5/5    | Deklaratives Schema, automatische Diff-Migration      |
| PostgreSQL-Features  | ⚠️ 3/5    | JSONB über Mixin möglich, aber umständlich            |
| Testbarkeit          | ✅ 4/5    | In-Memory-Client für Tests verfügbar                  |
| Runtime-Dependencies | ⚠️ 3/5    | +1 mit signifikantem Dependency-Tree (`entgo.io/ent`) |
| Lernkurve            | ❌ 2/5    | Eigene Schema-DSL, Codegen-Pipeline, Mixin-System     |

**Spezifische Probleme für jotti:**

1. **Event Store passt nicht ins Graph-Modell.** ent modelliert Entities mit Edges (Relationen). Events sind keine Entities mit Beziehungen — sie sind eine geordnete, unveränderliche Sequenz von Zustandsänderungen. Das lässt sich nicht als Graph ausdrücken.
2. **Overengineered für den Umfang.** jotti hat 4 CRUD-Tabellen und 1 Event-Tabelle. ents Stärke liegt in komplexen Graph-Traversals über Dutzende Entities — das ist hier nicht gegeben.
3. **Contributor-Einstiegshürde.** jottis Zielgruppe sind Vereine mit ehrenamtlichen Helfern. Potenzielle Contributors sollen SQL + Go-Structs mitbringen — nicht eine proprietäre Schema-DSL lernen müssen.

**Fazit für jotti:** ent ist ein hervorragendes Werkzeug für Graph-lastige Anwendungen. Für jottis kleines CRUD-Schema und Event-Sourced Core Domain ist es zu komplex, löst die falschen Probleme und erhöht die Einstiegshürde für Contributors.

---

## 3. Bewertungsmatrix

Gewichtung reflektiert jottis Architektur: Event-Sourcing ist die Core Domain (Kassenbetrieb), Compile-Time-Safety ist essentiell für die Pre-Release-Phase mit häufigen Schema-Änderungen.

| Kriterium                      | Gewicht | Raw SQL | sqlx    | **sqlc** | go-jet  | GORM    | ent     |
| ------------------------------ | ------- | ------- | ------- | -------- | ------- | ------- | ------- |
| Event-Sourcing-Kompatibilität  | 30%     | 5       | 5       | **5**    | 3       | 1       | 1       |
| CRUD-Komfort                   | 15%     | 2       | 3       | **4**    | 5       | 5       | 5       |
| Compile-Time-Safety            | 20%     | 1       | 1       | **5**    | 5       | 1       | 5       |
| PostgreSQL-Features nativ      | 15%     | 5       | 5       | **5**    | 3       | 2       | 3       |
| Einfachheit / Lernkurve        | 10%     | 5       | 4       | **4**    | 3       | 3       | 2       |
| Runtime-Dependencies           | 10%     | 5       | 4       | **5**    | 3       | 2       | 3       |
| **Gewichtetes Gesamtergebnis** |         | **3.8** | **3.5** | **4.7**  | **3.6** | **2.0** | **2.8** |

### Ergebnis-Ranking

1. **sqlc — 4.7** (SQL-first + Code-Gen)
2. Raw SQL — 3.8
3. go-jet — 3.6
4. sqlx — 3.5
5. ent — 2.8
6. GORM — 2.0

---

## 4. Entscheidung

**sqlc** bleibt das Persistenz-Werkzeug für jotti — sowohl für den Event Store als auch für Stammdaten-CRUD.

### 4.1 Begründung

Die Entscheidung ergibt sich nicht aus einer einzelnen Eigenschaft von sqlc, sondern aus der Kombination von jottis Architektur-Constraints:

**1. Event-Sourcing eliminiert den ORM-Nutzen.**

Events sind keine mutable Objekte. Der Tisch-Zustand wird nicht aus DB-Rows rekonstruiert, sondern durch `state = fold(apply, events)`. Ein ORM löst das Object-Relational-Mapping-Problem — jottis Core Domain hat dieses Problem nicht.

**2. Hybrid-Modell braucht ein Tool für beide Welten.**

sqlc bedient Event-Store-Queries (`INSERT INTO events ... RETURNING *`, CTE-basierter Snapshot-Replay) und CRUD-Queries (`SELECT p.*, json_agg(...) FROM produkte p`) gleich gut. Kein Tool-Wechsel zwischen Bounded Contexts.

**3. Schema-Drift-Schutz in der Pre-Release-Phase.**

jotti ändert das Schema direkt in `01_initial.up.sql`. `make sqlc` validiert alle Queries gegen das aktuelle Schema beim Build. Bei Raw SQL oder sqlx wären Schema-Inkompatibilitäten erst zur Laufzeit sichtbar — ein reales Risiko bei häufigen Änderungen.

**4. Fat Events + JSONB nativ.**

Event-Payloads als JSONB mit `json_agg()` und `json_build_object()` — sqlc versteht diese Queries und generiert korrekte Go-Typen (`json.RawMessage`). Kein ORM oder Query Builder versucht, die Event-Daten zu normalisieren oder über Fremdschlüssel aufzulösen.

**5. Null Runtime-Dependencies.**

sqlc ist ein reines Build-Tool. Der generierte Code hat keine Laufzeitabhängigkeit zu sqlc selbst. Für ein self-hosted Projekt, das Vereine mit Docker Compose betreiben, ist jede eingesparte Dependency eine weniger, die brechen kann.

**6. Niedrige Contributor-Einstiegshürde.**

SQL + Go-Structs kennt jeder Go-Entwickler. Kein proprietäres Schema-DSL (ent), keine ORM-Magie (GORM). Das passt zu jottis Zielgruppe — ehrenamtliche Contributors, die nicht wöchentlich am Projekt arbeiten.

### 4.2 Differenzierung pro Bounded Context

| Bounded Context       | Tool | Begründung                                                                          |
| --------------------- | ---- | ----------------------------------------------------------------------------------- |
| **Kassenbetrieb**     | sqlc | Append-only INSERT, CTE-Replay, JSONB-Payloads, OCC-Version                         |
| **Stammdaten**        | sqlc | Konsistenz mit Event Store, Soft-Delete explizit in SQL, `json_agg()` für Varianten |
| **Ausgabe** (MVP2)    | sqlc | LISTEN/NOTIFY + Event-Projektion — gleiche Query-Patterns                           |
| **Abrechnung** (MVP2) | sqlc | Materialized Views, Window Functions, CTEs — PostgreSQL-nativ                       |
| **Validierung**       | zog  | Unabhängig von Persistenz — Backend validiert vor DB-Zugriff                        |

---

## 5. Bekannte Schwächen und Mitigationen

| Schwäche                              | Impact  | Mitigation                                                                                                                                                |
| ------------------------------------- | ------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Drei User-Mapping-Funktionen**      | Niedrig | Verschiedene Result-Sets erfordern verschiedene Mapper. Explizit statt magisch — dokumentieren, warum die Queries unterschiedliche Spalten zurückgeben.   |
| **JSON-Unpacking-Boilerplate**        | Niedrig | ~50 LOC für `parseVariantsJSON` in `product_repo/types.go`. Einmalig geschrieben, selten geändert. Shared Helper extrahierbar bei Bedarf.                 |
| **Build-Step vergessbar**             | Mittel  | `make sqlc` muss nach Query-/Schema-Änderungen ausgeführt werden. Mitigation: CI-Pipeline prüft `make sqlc` + `git diff --exit-code`.                     |
| **Keine dynamischen Queries**         | Niedrig | Braucht separate SQL-Queries für verschiedene Filter-Kombinationen. Für jottis statische Queries kein Problem. Bei Bedarf: sqlc + Raw SQL in einem Repo.  |
| **`database/sql` statt pgx nativ**    | Niedrig | sqlc generiert gegen `database/sql` (via `sql_package`-Config). pgx-spezifische Features nicht direkt verfügbar. Umstellung auf `pgx/v5`-Package möglich. |
| **Generierter Code nicht editierbar** | Niedrig | `backend/sqlc/dbgen/` darf nicht editiert werden. Custom-Logik gehört in Repository-Wrapper. Bewusster Trade-off für Compile-Time-Safety.                 |

---

## 6. Konsequenzen

### Was sich nicht ändert

- Repository-Pattern mit Interface + Mock bleibt (Testbarkeit)
- Domain-Modelle bleiben unabhängig von DB-Typen (DDD)
- Zentrales Fehler-Mapping über `db.Error()` bleibt
- SQL-Queries bleiben explizit und inspizierbar

### Workflow

```
1. Schema ändern     → database/migrations/01_initial.up.sql
2. Query schreiben   → backend/sqlc/queries/*.sql
3. Code generieren   → make sqlc
4. Repo anpassen     → backend/repository/*_repo/
5. Tests schreiben   → *_test.go
6. Validieren        → make lint && make test
```

### Wann diese Entscheidung überprüfen?

- Wenn jotti aus der Pre-Release-Phase herauswächst und inkrementelle Migrationen braucht
- Wenn dynamische Query-Anforderungen signifikant zunehmen (z.B. flexible Reporting-Filter)
- Wenn PostgreSQL-spezifische Features (pgx nativ) stärker benötigt werden
- Wenn die Event-Sourcing-Strategie sich fundamental ändert

---

## Appendix: Vergleich mit ursprünglicher ADR

Die ursprüngliche ADR (`docs/adr/orm.md`) hat drei Alternativen bewertet (GORM, sqlx, Status quo) und sqlc gewählt. Diese Neufassung unterscheidet sich in:

| Aspekt                          | Ursprüngliche ADR       | Diese Neufassung                             |
| ------------------------------- | ----------------------- | -------------------------------------------- |
| Bewertete Alternativen          | 3 (GORM, sqlx, pgx)     | 5 Stufen (Raw → Full ORM) + 6 Tools          |
| Event-Sourcing-Analyse          | Erwähnt, nicht vertieft | Zentrale Constraint mit konkreten Beispielen |
| CRUD-Differenzierung            | Pauschal                | Pro Bounded Context differenziert            |
| Schwächen von sqlc              | Nur „Build-Step nötig"  | 6 konkrete Schwächen mit Mitigationen        |
| Code-Metriken                   | ~300 LOC Einsparung     | Detaillierte Aufschlüsselung (1.850 LOC)     |
| Bewertungsmethodik              | Subjektive Tabelle      | Gewichtete Matrix mit Begründung             |
| Query-Builder-Kategorie         | Nicht bewertet          | go-jet/bob als eigene Stufe analysiert       |
| ent                             | Nicht bewertet          | Als Graph-Schema-ORM analysiert              |
| Kontextbezug (jotti-spezifisch) | Generisch               | Konkrete Code-Beispiele und Pain Points      |
