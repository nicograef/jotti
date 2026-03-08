# ADR: Datenbankzugriff im Backend — Bewertung von ORM-Alternativen und Entscheidung für sqlc

## Status

**Entschieden** — sqlc wurde als Persistenz-Werkzeug für jotti übernommen.

## Kontext

jotti benötigt eine Persistenzschicht für zwei grundlegend verschiedene Strategien:

1. **CRUD für Stammdaten** — Benutzer, Tische, Produkte und Produktvarianten werden in relationalen Tabellen gespeichert (Create/Read/Update mit Soft-Deletes via `status = 'deleted'`).
2. **Event-Sourcing für Tisch-Operationen** — Bestellungen, Zahlungen, Lieferungen und Stornierungen werden als unveränderliche Events in einer Append-Only-Tabelle (`events`) gespeichert. Der aktuelle Zustand wird durch Replay aller Events rekonstruiert.

Die Persistenzschicht folgt dem **Repository-Pattern** mit klarer Schichtentrennung:

```
HTTP-Handler → Application-Service → Repository → PostgreSQL
                                         ↕
                                    Domain-Modell
```

Vor der Einführung von sqlc verwendete jotti handgeschriebene SQL-Queries mit der `database/sql`-Standardbibliothek und dem [`pgx/v5`](https://github.com/jackc/pgx)-PostgreSQL-Treiber. Jedes Repository-Paket enthielt:

- Ein `Repository`-Struct mit `*sql.DB`-Feld
- Private Datenbank-Typen (z. B. `dbuser`, `dbtable`) zur Abbildung auf SQL-Spalten
- `toDomain()`-Methoden zur Konvertierung von Datenbank- in Domain-Modelle
- Manuelles SQL via `QueryRowContext`, `QueryContext`, `ExecContext`
- Fehler-Mapping über das gemeinsame `db`-Paket

Dieser Ansatz gab volle Kontrolle über die Queries, war aber mit manuellem Boilerplate für `row.Scan()`-Aufrufe und Adapter-Structs verbunden — ohne Compile-Time-Validierung der SQL-Queries gegen das Schema.

---

## Datenbankschema

Das Schema wird durch SQL-Migrationsdateien unter `database/migrations/` definiert:

- `01_initial.up.sql` — Erstellt alle Tabellen, Enums, Indizes, Trigger und Seed-Daten
- `02_add_senior_service_role.up.sql` — Fügt `senior_service` zum `UserRole`-Enum hinzu

### Enums

| Enum              | Werte                                | Verwendet in                                               |
| ----------------- | ------------------------------------ | ---------------------------------------------------------- |
| `UserRole`        | `admin`, `senior_service`, `service` | `users.role`                                               |
| `EntityStatus`    | `active`, `inactive`, `deleted`      | `users.status`, `tables.status`, `product_variants.status` |
| `ProductCategory` | `food`, `beverage`, `other`          | `products.category`                                        |

Enums sind PostgreSQL-Custom-Types (`CREATE TYPE ... AS ENUM`). Die Go-Domain-Modelle spiegeln diese als typisierte String-Konstanten wider (z. B. `user.AdminRole = "admin"`).

### Tabellen

#### `users`

System-Benutzer, die Aktionen in jotti ausführen.

| Spalte                  | Typ              | Nullable | Beschreibung                                      |
| ----------------------- | ---------------- | -------- | ------------------------------------------------- |
| `id`                    | `INT` (Identity) | Nein     | Primärschlüssel, automatisch generiert            |
| `name`                  | `TEXT`           | Nein     | Vollständiger Name                                |
| `username`              | `TEXT` (Unique)  | Nein     | Eindeutiger Login-Name                            |
| `password_hash`         | `TEXT`           | Ja       | Argon2id-Hash; NULL bis Passwort gesetzt          |
| `onetime_password_hash` | `TEXT`           | Ja       | Einmalpasswort für Onboarding/Reset               |
| `role`                  | `UserRole`       | Nein     | Zugriffsrolle: admin, senior_service oder service |
| `status`                | `EntityStatus`   | Nein     | active, inactive oder deleted (Soft-Delete)       |
| `created_at`            | `TIMESTAMPTZ`    | Nein     | Erstellungszeitpunkt (UTC)                        |

Indizes: `idx_users_username` (username), `idx_users_status` (status).

#### `tables`

Kunden sitzen an Tischen und bestellen von dort.

| Spalte       | Typ              | Nullable | Beschreibung                                |
| ------------ | ---------------- | -------- | ------------------------------------------- |
| `id`         | `INT` (Identity) | Nein     | Primärschlüssel, automatisch generiert      |
| `name`       | `TEXT` (Unique)  | Nein     | Name oder Nummer (z. B. „Tisch 1")          |
| `status`     | `EntityStatus`   | Nein     | active, inactive oder deleted (Soft-Delete) |
| `created_at` | `TIMESTAMPTZ`    | Nein     | Erstellungszeitpunkt (UTC)                  |

Index: `idx_tables_status` (status).

#### `products`

Produkte, die von Kunden bestellt werden können.

| Spalte       | Typ               | Nullable | Beschreibung                           |
| ------------ | ----------------- | -------- | -------------------------------------- |
| `id`         | `INT` (Identity)  | Nein     | Primärschlüssel, automatisch generiert |
| `name`       | `TEXT`            | Nein     | Produktname                            |
| `category`   | `ProductCategory` | Nein     | food, beverage oder other              |
| `created_at` | `TIMESTAMPTZ`     | Nein     | Erstellungszeitpunkt (UTC)             |

Produkte haben keine Status-Spalte — ihre Sichtbarkeit wird durch den Status ihrer Varianten gesteuert.

#### `product_variants`

Varianten von Produkten mit individuellen Preisen (z. B. „Cola 0,3L" und „Cola 0,5L").

| Spalte        | Typ              | Nullable | Beschreibung                                |
| ------------- | ---------------- | -------- | ------------------------------------------- |
| `id`          | `INT` (Identity) | Nein     | Primärschlüssel, automatisch generiert      |
| `product_id`  | `INT` (FK)       | Nein     | Referenz auf `products(id)`                 |
| `name`        | `TEXT`           | Nein     | Variantenname (z. B. „0,5L")                |
| `price_cents` | `INT`            | Nein     | Preis in Cent (z. B. 299 für 2,99 €)        |
| `status`      | `EntityStatus`   | Nein     | active, inactive oder deleted (Soft-Delete) |
| `created_at`  | `TIMESTAMPTZ`    | Nein     | Erstellungszeitpunkt (UTC)                  |

Index: `idx_product_variants_status` (status).

### Events-Tabelle (Event-Sourcing)

Die `events`-Tabelle ist der Kern des Event-Sourcing-Systems für Tisch-Operationen. Sie ist **append-only** — Zeilen werden nur eingefügt und gelesen, nie aktualisiert oder gelöscht.

| Spalte      | Typ              | Nullable | Beschreibung                                       |
| ----------- | ---------------- | -------- | -------------------------------------------------- |
| `id`        | `INT` (Identity) | Nein     | Primärschlüssel, für Sortierung verwendet          |
| `user_id`   | `INT` (FK)       | Nein     | Referenz auf `users(id)` — der Akteur              |
| `type`      | `TEXT`           | Nein     | Event-Typ (z. B. `tisch.bestellung-aufgegeben:v1`) |
| `subject`   | `TEXT`           | Nein     | Aggregat-Schlüssel (z. B. `tisch:42`)              |
| `timestamp` | `TIMESTAMPTZ`    | Nein     | Event-Zeitpunkt (UTC)                              |
| `data`      | `JSONB`          | Nein     | Event-Payload, versioniert nach Typ                |

**Indizes:** `idx_events_user_id`, `idx_events_subject`, `idx_events_type`, `idx_events_subject_type` (Komposit).

**Unveränderlichkeits-Schutz:**

Die Events-Tabelle verfügt über mehrere Schutzebenen gegen Mutation:

1. **Privilegien-Einschränkung:** `REVOKE ALL ON TABLE events FROM PUBLIC; GRANT SELECT, INSERT ON TABLE events TO PUBLIC;` — nur SELECT und INSERT für Nicht-Owner-Rollen.
2. **Trigger-basierte Durchsetzung** (für ALLE Rollen einschließlich Table-Owner):
   - `events_no_update` — `BEFORE UPDATE`-Trigger löst eine Exception aus
   - `events_no_delete` — `BEFORE DELETE`-Trigger löst eine Exception aus
   - `events_no_truncate` — `BEFORE TRUNCATE`-Trigger löst eine Exception aus

Alle drei Trigger rufen `prevent_event_mutation()` auf, die eine PostgreSQL-Exception mit der Nachricht `events table is append-only: <Operation> not allowed` auslöst.

**Event-Typen** (definiert in `domain/table/events.go`):

| Event-Typ                        | Beschreibung                     |
| -------------------------------- | -------------------------------- |
| `tisch.bestellung-aufgegeben:v1` | Eine Bestellung wurde aufgegeben |
| `tisch.zahlung-registriert:v1`   | Eine Zahlung wurde registriert   |
| `tisch.produkte-storniert:v1`    | Produkte wurden storniert        |
| `tisch.produkte-geliefert:v1`    | Produkte wurden geliefert        |
| `tisch.snapshot:v1`              | Ein Zustands-Snapshot            |

### Indizes

| Index                         | Tabelle            | Spalte(n)         | Zweck                                        |
| ----------------------------- | ------------------ | ----------------- | -------------------------------------------- |
| `idx_users_username`          | `users`            | `username`        | Schnelle Suche nach Benutzername (Login)     |
| `idx_users_status`            | `users`            | `status`          | Filterung aktiver/inaktiver Benutzer         |
| `idx_tables_status`           | `tables`           | `status`          | Filterung aktiver/inaktiver Tische           |
| `idx_product_variants_status` | `product_variants` | `status`          | Filterung aktiver/inaktiver Varianten        |
| `idx_events_user_id`          | `events`           | `user_id`         | Events nach Akteur finden                    |
| `idx_events_subject`          | `events`           | `subject`         | Events für ein bestimmtes Aggregat finden    |
| `idx_events_type`             | `events`           | `type`            | Events nach Typ finden                       |
| `idx_events_subject_type`     | `events`           | `subject`, `type` | Snapshots für ein bestimmtes Aggregat finden |

### Migrationen

Migrationen werden mit [`golang-migrate`](https://github.com/golang-migrate/migrate) verwaltet und unter `database/migrations/` gespeichert:

```
database/migrations/
  01_initial.up.sql
  01_initial.down.sql
  02_add_senior_service_role.up.sql
  02_add_senior_service_role.down.sql
```

Jede Migration hat eine `up`- (anwenden) und `down`-Datei (zurückrollen). Die initiale Migration umschließt alle DDL-Anweisungen in einem `BEGIN`/`COMMIT`-Transaktionsblock.

---

## Datenbankverbindung (`db`-Paket)

Das `db`-Paket (`backend/db/`) stellt gemeinsame Hilfsfunktionen für den Datenbankzugriff bereit. Es verwaltet **nicht** die Verbindung selbst — der `*sql.DB`-Pool wird in `backend/app/app.go` beim Anwendungsstart erstellt und in alle Repositories injiziert.

### Fehler-Mapping

Die Funktion `db.Error(err)` übersetzt Low-Level-Datenbankfehler in domänenspezifische Sentinel-Errors:

| Datenbankfehler                       | Abgebildet auf        |
| ------------------------------------- | --------------------- |
| PostgreSQL Unique Violation (`23505`) | `db.ErrAlreadyExists` |
| `sql.ErrNoRows`                       | `db.ErrNotFound`      |
| Alle anderen Fehler                   | `db.ErrDatabase`      |

Die Funktion `db.ResultError(result)` prüft `RowsAffected()` und gibt `db.ErrNotFound` zurück, wenn keine Zeilen betroffen waren (verwendet für UPDATE-Operationen).

Diese Sentinel-Errors ermöglichen es Application-Services und HTTP-Handlern, zwischen „nicht gefunden" (→ 404), „existiert bereits" (→ 409) und „interner Fehler" (→ 500) zu unterscheiden, ohne Datenbankdetails preiszugeben.

### NullTime-Helfer

Der `db.NullTime`-Typ ist ein benutzerdefinierter Nullable-Time-Wrapper, der Folgendes implementiert:

- `sql.Scanner` — zum Lesen von `TIMESTAMPTZ`-Spalten (behandelt sowohl `NULL`- als auch `time.Time`-Werte)
- `driver.Valuer` — zum Schreiben von Zeitwerten in die Datenbank
- `json.Unmarshaler` / `json.Marshaler` — zum Parsen und Kodieren von Zeitwerten in JSON

Dies ist notwendig, weil `sql.NullTime` kein `json.Unmarshaler` implementiert, was Probleme beim Scannen von JSON-Ergebnissen aus PostgreSQLs `json_agg()`-Funktion verursacht. Das `product_repo`-Paket verwendet `json_agg()`, um Produktvarianten in JSON-Arrays zu aggregieren.

### Ressourcen-Bereinigung

Die Funktion `db.Close(closer, name)` schließt sicher einen `io.Closer` (typischerweise `*sql.Rows`) und protokolliert Fehler mit `zerolog`. Dies wird in Repositories mit `defer db.Close(rows, "Ressourcenname")` verwendet.

### Test-Datenbank

Die Funktion `db.OpenTestDatabase()` öffnet eine Verbindung zu einer lokalen PostgreSQL-Instanz (`localhost:5432`, Datenbank `jotti`, Benutzer/Passwort `admin`) für Integrationstests. Sie beendet das Programm, wenn die Verbindung fehlschlägt.

---

## Repository-Schicht (vor sqlc)

Jedes Domain-Paket hatte sein eigenes Repository-Paket unter `backend/repository/`:

```
backend/repository/
  event_repo/       # Event-Sourcing (append-only)
  product_repo/     # Produkte + Varianten (CRUD)
  table_repo/       # Tische (CRUD)
  user_repo/        # Benutzer (CRUD)
```

Jedes Repository-Paket enthielt:

| Datei          | Zweck                                        |
| -------------- | -------------------------------------------- |
| `types.go`     | Privater DB-Struct + `toDomain()`-Konverter  |
| `repo.go`      | `Repository`-Struct + SQL-Query-Methoden     |
| `mock.go`      | In-Memory-Mock für Unit-Tests                |
| `repo_test.go` | Integrationstests (`//go:build integration`) |

### DB-zu-Domain-Mapping (vor sqlc)

Jedes CRUD-Repository definierte einen privaten Struct, der das Datenbank-Zeilenlayout spiegelte, mit `sql.NullString`, `sql.NullTime` oder `db.NullTime` für nullable Spalten. Diese Structs hatten eine `toDomain()`-Methode zur Konvertierung in Domain-Modelltypen:

```go
// user_repo/types.go (vor sqlc)
type dbuser struct {
    ID                  int            `db:"id"`
    Name                string         `db:"name"`
    Username            string         `db:"username"`
    Role                string         `db:"role"`
    Status              string         `db:"status"`
    PasswordHash        sql.NullString `db:"password_hash"`
    OnetimePasswordHash sql.NullString `db:"onetime_password_hash"`
    CreatedAt           sql.NullTime   `db:"created_at"`
}

func (dp *dbuser) toDomain() user.User {
    return user.User{
        ID:                  dp.ID,
        Name:                dp.Name,
        Username:            dp.Username,
        Role:                user.Role(dp.Role),
        Status:              user.Status(dp.Status),
        PasswordHash:        dp.PasswordHash.String,
        OnetimePasswordHash: dp.OnetimePasswordHash.String,
        CreatedAt:           dp.CreatedAt.Time,
    }
}
```

Die `db:"..."`-Struct-Tags waren rein dokumentarisch — sie wurden nicht von einem Reflection-basierten Mapper verwendet. Alle Spalten-Zuordnungen erfolgten über explizite `rows.Scan()`-Aufrufe mit positionalen Argumenten.

**Richtung des Mappings:**

- **Lesen (DB → Domain):** `Scan()` in DB-Struct → `toDomain()` → Domain-Modell zurückgeben
- **Schreiben (Domain → DB):** Felder aus Domain-Modell extrahieren → direkt als SQL-Parameter übergeben

### CRUD-Repositories

Die drei CRUD-Repositories (`user_repo`, `table_repo`, `product_repo`) teilten die gleichen Muster:

- **`user_repo`** — Verwaltung der `users`-Tabelle (GetUser, GetUserByUsername, GetAllUsers, CreateUser, UpdateUser)
- **`table_repo`** — Verwaltung der `tables`-Tabelle (GetTable, GetAllTables, GetActiveTables, CreateTable, UpdateTable)
- **`product_repo`** — Verwaltung der `products`- und `product_variants`-Tabellen, einschließlich komplexer JSON-Aggregation mit `json_agg()` + CTEs

#### JSON-Aggregationsmuster (product_repo)

Produkte werden immer mit ihren Varianten über PostgreSQLs `json_agg()` und `json_build_object()` abgerufen. Dies vermeidet das N+1-Query-Problem:

```sql
WITH variant_json AS (
    SELECT product_id,
           json_agg(json_build_object(
               'id', id, 'name', name,
               'price_cents', price_cents,
               'status', status, 'created_at', created_at
           )) AS variants
    FROM product_variants
    WHERE status != 'deleted'
    GROUP BY product_id
)
SELECT p.id, p.name, p.category, p.created_at,
       COALESCE(vj.variants, '[]') AS variants
FROM products p
LEFT JOIN variant_json vj ON vj.product_id = p.id
ORDER BY p.id ASC
```

### Event-Repository (`event_repo`)

Das Event-Repository behandelt die Append-Only-`events`-Tabelle. Im Gegensatz zu den CRUD-Repositories hatte es kein `types.go` — das `event.Event`-Domain-Modell wurde direkt gescannt, weil die `data`-Spalte (`JSONB`) als `json.RawMessage` gespeichert wird.

**Snapshot-Optimierung:**

Die `ReadEventsWithSnapshot`-Methode verwendet ein CTE, um das letzte Snapshot-Event zu finden und nur Events ab diesem Punkt zu lesen:

```sql
WITH last_snapshot AS (
    SELECT COALESCE(MAX(id), 0) AS id
    FROM events
    WHERE subject = $1 AND type = $2
)
SELECT e.id, e.user_id, e.type, e.subject, e.data, e.timestamp
FROM events e, last_snapshot ls
WHERE e.subject = $1 AND e.id >= ls.id
ORDER BY e.id ASC
```

### Mock-Repositories

Jedes Repository-Paket bietet einen `NewMock(items, err)`-Konstruktor, der eine In-Memory-Implementierung des Repository-Interfaces zurückgibt. Mocks werden in **Unit-Tests** für Application-Services und HTTP-Handler verwendet.

### Integrationstests

Jedes Repository hat eine Integrationstestdatei mit dem `//go:build integration`-Build-Tag. Diese Tests laufen gegen eine echte PostgreSQL-Instanz.

---

## Duale Persistenzstrategie

jotti verwendet zwei grundlegend verschiedene Persistenzstrategien:

### 1. CRUD für Stammdaten

**Benutzer**, **Tische**, **Produkte** und **Produktvarianten** werden in traditionellen relationalen Tabellen mit Standard-CRUD-Operationen gespeichert. Löschungen sind Soft-Deletes via `status = 'deleted'` — Datensätze werden nie physisch entfernt.

### 2. Event-Sourcing für Tisch-Operationen

**Bestellungen**, **Zahlungen**, **Lieferungen** und **Stornierungen** werden als unveränderliche Events gespeichert. Der aktuelle Zustand eines Tisches wird durch Replay seiner Events rekonstruiert:

- **Saldo** = Σ(Bestellsummen) − Σ(Bezahlsummen) − Σ(Stornierungssummen)
- **Unbezahlt** = bestellt − bezahlt − storniert
- **Ungeliefert** = bestellt − geliefert − storniert

Snapshots (`tisch.snapshot:v1`-Events) erfassen periodisch den vollständigen Zustand, sodass `ReadEventsWithSnapshot` das Replay älterer Events überspringen kann.

---

## Bewertete Alternativen

### GORM — Vollständiges ORM

**Ergebnis: Nicht geeignet für jotti.**

GORM (Go's populärstes ORM) wurde umfassend evaluiert. Die Kernprobleme:

1. **Event-Sourcing widerspricht GORMs Grundannahmen.** GORMs gesamtes Lebenszyklusmodell (Create → Update → Delete) setzt veränderliche Datensätze voraus. Ein Append-Only-Event-Store widerspricht dieser Annahme fundamental.

2. **Jottis komplexestes SQL kann nicht in GORM ausgedrückt werden.** Das `product_repo` verwendet PostgreSQLs `json_agg()` mit CTEs. Das Event-Repo verwendet CTE-basierte Snapshot-Queries. Beides erfordert `db.Raw()` — rohes SQL innerhalb eines ORMs, was die Abstraktion aufgibt, aber die Abhängigkeitskosten behält.

3. **Data-Mapper vs. Active-Record-Konflikt.** jotti trennt sauber `domain/` (Business-Logik) von `repository/` (SQL). GORM drängt zum Active-Record-Pattern mit `gorm:"..."`-Tags auf Domain-Structs.

4. **Datenbankportabilität ist irrelevant.** jotti ist auf PostgreSQL festgelegt (Custom-Enums, JSONB, Trigger, CTEs, Privilegien-Einschränkungen).

5. **Kosten-Nutzen-Verhältnis ungünstig.** GORM würde ~170 Zeilen CRUD-Boilerplate reduzieren. Im Gegenzug: große Abhängigkeit (~20k Codezeilen), geteilte Architektur (ORM für CRUD, rohes SQL für alles andere), komplexeres Mocking.

**Bewertungsmatrix (GORM):**

| Kriterium                  | GORM-Bewertung                   | jotti-Passung                                |
| -------------------------- | -------------------------------- | -------------------------------------------- |
| CRUD-Boilerplate-Reduktion | ✅ Moderat                       | ⚠️ Einsparungen real aber gering             |
| JSON-Aggregation           | ❌ Nicht ersetzbar               | ❌ Erfordert `db.Raw()`                      |
| Event-Sourcing             | ❌ Fundamental inkompatibel      | ❌ Widerspricht Lifecycle-Modell             |
| Domain-Trennung            | ❌ Active Record vs. Data Mapper | ❌ Architekturkonflikt                       |
| PostgreSQL-Features        | ❌ Eingeschränkte Unterstützung  | ❌ Rohes SQL für Enums, Trigger, JSONB       |
| Testing & Mocking          | ❌ Komplexer                     | ❌ Interface-basiertes Mocking ist einfacher |
| Abhängigkeiten             | ❌ Große Abhängigkeit            | ❌ jotti ist bewusst abhängigkeitsarm        |

### sqlx — `database/sql`-Erweiterungsbibliothek

**Ergebnis: Valide aber bescheidene Verbesserung.**

sqlx ist ein leichtgewichtiger Wrapper um `database/sql`, der automatisches Struct-Scanning via `db:"..."`-Tags bietet. Es ist kein ORM und kein Code-Generator.

**Vorteile:**

- Minimale Lernkurve (`database/sql`-Superset)
- Drop-in-kompatibel, inkrementelle Migration möglich
- Eliminiert `row.Scan()`-Boilerplate (~100 Zeilen Einsparung)
- Voll kompatibel mit Data-Mapper-Pattern

**Nachteile:**

- Keine Compile-Time-Validierung (Fehler erst zur Laufzeit erkannt)
- Keine Schema-Validierung
- Runtime-Reflection für Struct-Scanning
- pgx bietet bereits ähnliche Features (`pgx.CollectRows()`, `pgx.RowToStructByName()`)
- Einsparung gering (~100 Zeilen) für eine neue Abhängigkeit

**Bewertungsmatrix (sqlx):**

| Kriterium                  | sqlx-Bewertung              | jotti-Passung                           |
| -------------------------- | --------------------------- | --------------------------------------- |
| CRUD-Boilerplate-Reduktion | ✅ Moderat                  | ⚠️ ~100 Zeilen; Adapter-Structs bleiben |
| Event-Sourcing             | ✅ Keine Lifecycle-Annahmen | ✅ Funktioniert natürlich               |
| Domain-Trennung            | ✅ Voll kompatibel          | ✅ Data-Mapper-Pattern erhalten         |
| Compile-Time-Validierung   | ❌ Keine Schema-Erkennung   | ❌ Fehler nur zur Laufzeit              |
| Abhängigkeiten             | ⚠️ Kleine Abhängigkeit      | ⚠️ +1 Abhängigkeit für ~100 Zeilen      |

### sqlc — SQL-Compiler (Code-Generator)

**Ergebnis: Starker Kandidat → Übernommen.**

sqlc ist **kein** ORM. Es ist ein **Code-Generator**, der einen fundamental anderen Ansatz für Datenbankzugriff verfolgt: Statt Objekte zur Laufzeit auf Tabellen abzubilden, **parst es SQL-Queries zur Compile-Time** und generiert typsichere Go-Funktionen, Parameter-Structs und Ergebnis-Structs.

#### Wie sqlc funktioniert

1. **Schema definieren** — sqlc liest Migrationsdateien, um ein internes Modell des Datenbankschemas zu erstellen
2. **Annotierte SQL-Queries schreiben** — Standard-SQL in `.sql`-Dateien mit spezieller Kommentar-Annotation:
   ```sql
   -- name: GetUserByUsername :one
   SELECT id, name, username, role, status, created_at
   FROM users
   WHERE username = $1 AND status != 'deleted';
   ```
3. **Go-Code generieren** — `sqlc generate` erzeugt `models.go` (Structs), Query-Funktionen und ein `Queries`-Struct mit `DBTX`-Interface

#### Dimensions-Analyse (sqlc vs. jotti)

**CRUD-Repositories:** sqlc bietet einen klaren Vorteil. Es eliminiert manuelles Scanning-Boilerplate bei Beibehaltung der exakt gleichen SQL-Queries. Compile-Time-Validierung fängt Schema-Drift ab.

**Produkt-Repository (JSON-Aggregation):** sqlc unterstützt CTEs, `json_agg()` und `json_build_object()` vollständig, weil es PostgreSQLs eigenen Parser verwendet. Deutlich besser als GORM.

**Event-Repository (Event-Sourcing):** sqlc ist eine exzellente Passung. Im Gegensatz zu GORM generiert sqlc genau die Funktionen, die definiert werden — keine Lifecycle-Hooks, keine Update/Delete-Methoden, es sei denn, man schreibt diese SQL-Queries. Ein Append-Only-Store ist vollkommen natürlich.

**Domain-Trennung:** Kompatibel mit Data-Mapper-Pattern. sqlc generiert reine Daten-Structs ohne Tags, Methoden oder Framework-Kopplung. Die `domain/`-Schicht bleibt unberührt.

**PostgreSQL-Features:** Exzellente Unterstützung. sqlc verwendet PostgreSQLs eigenen Parser (`pg_query_go`) und versteht praktisch alle PostgreSQL-spezifischen Features.

**Fehlerbehandlung:** Voll kompatibel. sqlc führt keine eigene Fehlerabstraktion ein. Das bestehende `db.Error()`-Mapping funktioniert unverändert.

**Testing & Mocking:** Kompatibel. Die bestehende Interface-basierte Mock-Strategie bleibt erhalten.

**Abhängigkeiten:** Zero Runtime-Dependency. sqlc ist ein Build-Tool, keine Laufzeitabhängigkeit. Der generierte Code verwendet `database/sql` oder `pgx/v5` — dieselben Interfaces, die jotti bereits nutzt.

**Bewertungsmatrix (sqlc):**

| Kriterium                  | sqlc-Bewertung                          | jotti-Passung                          |
| -------------------------- | --------------------------------------- | -------------------------------------- |
| CRUD-Boilerplate-Reduktion | ✅ Eliminiert `Scan()`-Boilerplate      | ✅ ~300 Zeilen Einsparung              |
| JSON-Aggregation           | ✅ Voller PostgreSQL-Parser             | ✅ CTE + `json_agg()` nativ            |
| Event-Sourcing             | ✅ Keine Lifecycle-Annahmen             | ✅ Nur definierte Queries generiert    |
| Domain-Trennung            | ✅ Generierte Models sind plain Structs | ✅ Data-Mapper kompatibel              |
| PostgreSQL-Features        | ✅ PostgreSQL-eigener Parser            | ✅ Enums, JSONB, CTEs, Trigger         |
| Compile-Time-Validierung   | ✅ Schema-Drift und Typos erkannt       | ✅ Verhindert Laufzeitfehler           |
| Testing & Mocking          | ✅ Standard-Go-Interfaces               | ✅ Bestehende Strategie erhalten       |
| Abhängigkeiten             | ✅ Zero Runtime-Dependency              | ✅ Nur Build-Tool                      |
| Go-Idiom-Alignment         | ✅ Idiomatischer Go-Code                | ✅ Passt zu Gos expliziter Philosophie |
| Build-Workflow             | ⚠️ `sqlc generate`-Schritt nötig        | ⚠️ Muss in CI/CD integriert werden     |

### Status Quo — Bare `database/sql` + `pgx/v5`

**Ergebnis: Lauffähig, aber mit Schwächen.**

Der ursprüngliche Ansatz war sauber, explizit und funktional. Hauptschwächen:

- Manuelles `Scan()`-Boilerplate mit positionalen Argumenten
- Keine Compile-Time-Validierung der SQL-Queries
- Kein Schutz vor Schema-Drift

---

## Gesamtvergleich

| Feature                     | Status Quo           | GORM                        | sqlc                     | sqlx                   |
| --------------------------- | -------------------- | --------------------------- | ------------------------ | ---------------------- |
| **Query-Sprache**           | Rohes SQL            | Go-API + `db.Raw()`         | Rohes SQL (annotiert)    | Rohes SQL              |
| **Typsicherheit**           | Laufzeit             | Laufzeit (Reflection)       | Compile-Time             | Laufzeit (Reflection)  |
| **Schema-Validierung**      | Keine                | Teilweise                   | Vollständig              | Keine                  |
| **CTE-Unterstützung**       | ✅ Natives SQL       | ❌ Nur `db.Raw()`           | ✅ Natives SQL           | ✅ Natives SQL         |
| **`json_agg()`**            | ✅ Natives SQL       | ❌ Nur `db.Raw()`           | ✅ Natives SQL           | ✅ Natives SQL         |
| **Event-Sourcing**          | ✅ Keine Annahmen    | ❌ Veränderliches Lifecycle | ✅ Keine Annahmen        | ✅ Keine Annahmen      |
| **Data-Mapper**             | ✅ Aktuelles Pattern | ❌ Active Record            | ✅ Generierte Models     | ✅ Tag-basiert         |
| **Laufzeit-Abhängigkeiten** | 0 neue               | +2                          | 0 neue (Build-Tool)      | +1                     |
| **Boilerplate-Reduktion**   | Basis                | ~170 Zeilen (nur CRUD)      | ~300 Zeilen (alle Repos) | ~100 Zeilen (Scanning) |
| **Framework-Lock-in**       | Keiner               | Hoch                        | Gering                   | Gering                 |

### Passung für jotti

| jotti-Anforderung                 | Status Quo     | GORM                 | sqlc           | sqlx                 |
| --------------------------------- | -------------- | -------------------- | -------------- | -------------------- |
| Event-Sourcing (Append-Only-Kern) | ✅             | ❌                   | ✅             | ✅                   |
| PostgreSQL-spezifisches SQL       | ✅             | ❌                   | ✅             | ✅                   |
| Data-Mapper-Architektur           | ✅             | ❌                   | ✅             | ✅                   |
| Drei-Stufen-Soft-Deletes          | ✅             | ⚠️                   | ✅             | ✅                   |
| Abhängigkeitsarme Philosophie     | ✅             | ❌                   | ✅             | ⚠️                   |
| Compile-Time-Fehlererkennung      | ❌             | ❌                   | ✅             | ❌                   |
| Schema-Drift-Schutz               | ❌             | ⚠️                   | ✅             | ❌                   |
| Boilerplate-Reduktion             | ❌             | ⚠️                   | ✅             | ⚠️                   |
| Go-Idiom-Alignment                | ✅             | ❌                   | ✅             | ✅                   |
| **Bewertung**                     | **8 ✅, 2 ❌** | **1 ✅, 4 ❌, 5 ⚠️** | **9 ✅, 1 ⚠️** | **7 ✅, 2 ❌, 1 ⚠️** |

---

## Entscheidung

**sqlc wurde als Persistenz-Werkzeug für jotti übernommen.**

### Begründung

1. **sqlc passt zur Projektphilosophie:** SQL-first, PostgreSQL-nativ, explizit und abhängigkeitsarm. Es adressiert die Hauptschwächen des Status Quo — manuelles `Scan()`-Boilerplate und fehlende Compile-Time-Validierung — ohne die Architektur zu kompromittieren.

2. **Keine Architekturkonflikte:** Im Gegensatz zu GORM ist sqlc mit Event-Sourcing, Data-Mapper-Pattern und PostgreSQL-spezifischen Features voll kompatibel.

3. **Höchste Bewertung:** Mit 9 von 10 ✅-Bewertungen bei den jotti-Anforderungen schneidet sqlc deutlich besser ab als alle Alternativen.

4. **Kein Trade-off bei der Laufzeit:** sqlc ist ein Build-Tool ohne Laufzeitabhängigkeit. Der generierte Code ist identisch mit handgeschriebenem `database/sql`-Code.

### Umsetzung

Die Migration wurde wie folgt durchgeführt:

1. `sqlc.yaml`-Konfiguration erstellt, die auf die bestehenden Migrationsdateien verweist
2. `.sql`-Query-Dateien für jedes Repository erstellt, basierend auf den bestehenden SQL-Queries
3. `sqlc generate` ausgeführt, um den initialen Go-Code zu erzeugen
4. Jedes Repository refaktorisiert, um sqlcs generiertes `Queries`-Struct zu wrappen und Ergebnisse auf Domain-Typen abzubilden
5. Handgeschriebene `db*`-Adapter-Structs und `Scan()`-Aufrufe durch sqlc-generierte Models und Funktionen ersetzt

Die Migration erfolgte inkrementell — ein Repository nach dem anderen — ohne den Rest der Codebasis zu beeinträchtigen.

---

## Konsequenzen

### Positive Konsequenzen

- Compile-Time-Validierung aller SQL-Queries gegen das aktuelle Datenbankschema
- Eliminierung von manuellem `Scan()`-Boilerplate (~300 Zeilen eingespart)
- Typsichere Query-Parameter und Ergebnisse
- Keine neue Laufzeitabhängigkeit
- Bestehende Mock- und Teststrategie vollständig erhalten

### Negative Konsequenzen

- Zusätzlicher Build-Schritt (`sqlc generate`) bei Schema- oder Query-Änderungen
- SQL-Queries leben in separaten `.sql`-Dateien statt inline im Go-Code
- Konfiguration der Custom-Types (Enums, JSONB) in `sqlc.yaml` erforderlich
- Generierte Dateien (`sqlc/dbgen/`) dürfen nicht manuell editiert werden

### Dateien und Verzeichnisse

Die sqlc-Integration ist in folgenden Dateien und Verzeichnissen zu finden:

- `backend/sqlc.yaml` — sqlc-Konfiguration
- `backend/sqlc/queries/` — SQL-Query-Definitionen
- `backend/sqlc/dbgen/` — Generierter Code (NICHT EDITIEREN)
- `backend/repository/` — Repositories, die sqlc-generierte Funktionen wrappen
