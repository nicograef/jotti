# Datenbank & Persistenz

Dieses Dokument beschreibt den Aufbau der Backend-Datenbankkommunikation in jotti.

> **Hintergrund:** Die Analyse und Bewertung der verschiedenen Ansätze (GORM, sqlc, sqlx, Status Quo) ist in der ADR unter [docs/adr/orm.md](adr/orm.md) dokumentiert.

---

## Überblick

jotti verwendet **[sqlc](https://sqlc.dev/)** — einen Compile-Time-SQL-Code-Generator — um typsichere Go-Funktionen aus handgeschriebenen SQL-Queries zu generieren. Der generierte Code nutzt die `database/sql`-Standardbibliothek mit dem [`pgx/v5`](https://github.com/jackc/pgx)-PostgreSQL-Treiber.

Die Persistenzschicht folgt dem **Repository-Pattern**:

```
HTTP-Handler → Application-Service → Repository → PostgreSQL
                                         ↕
                                    Domain-Modell
```

Es gibt zwei Persistenzstrategien:

- **CRUD für Stammdaten** (Benutzer, Tische, Produkte, Varianten) → relationale Tabellen mit Soft-Deletes
- **Event-Sourcing für Tisch-Operationen** (Bestellungen, Bezahlungen, Lieferungen, Stornierungen) → append-only `events`-Tabelle

---

## Verzeichnisstruktur

```
backend/
  sqlc.yaml                     # sqlc-Konfiguration
  sqlc/queries/                 # SQL-Query-Definitionen (.sql-Dateien)
    users.sql
    tables.sql
    products.sql
    events.sql
  sqlc/dbgen/                   # Generierter Code (NICHT EDITIEREN)
    db.go                       # DBTX-Interface + Queries-Struct
    models.go                   # Datenbank-Model-Typen + Enums
    *.sql.go                    # Generierte Query-Funktionen
  repository/
    user_repo/                  # Benutzer-Repository
    table_repo/                 # Tische-Repository
    product_repo/               # Produkte + Varianten
    event_repo/                 # Event-Sourcing (append-only)
  db/                           # Fehler-Mapping, NullTime-Helfer, Test-DB

database/
  migrations/                   # SQL-Migrationen (golang-migrate)
    01_initial.up.sql           # Hauptschema
    01_initial.down.sql
    02_add_senior_service_role.up.sql
    02_add_senior_service_role.down.sql
```

---

## Workflow: Neue Query hinzufügen

1. **SQL-Query schreiben** in `backend/sqlc/queries/<domain>.sql`:
   ```sql
   -- name: GetActiveUsers :many
   SELECT id, name, username, role, status, created_at
   FROM users WHERE status = 'active' ORDER BY name;
   ```

2. **Code generieren:**
   ```bash
   cd backend && sqlc generate
   ```

3. **Repository-Methode erstellen** in `backend/repository/<domain>_repo/repo.go`, die die sqlc-generierte Funktion aufruft und das Ergebnis auf Domain-Typen abbildet.

---

## Repository-Aufbau

Jedes Repository-Paket enthält:

| Datei           | Inhalt                                                              |
| --------------- | ------------------------------------------------------------------- |
| `types.go`      | `Repository`-Struct, `NewRepository()`-Konstruktor, Typ-Konverter  |
| `repo.go`       | Repository-Methoden (wrappen sqlc-generierte Query-Funktionen)     |
| `mock.go`       | In-Memory-Mock für Unit-Tests                                      |
| `repo_test.go`  | Integrationstests (`//go:build integration`)                       |

**Beispiel** — Repository-Aufbau (siehe `backend/repository/user_repo/types.go`):

```go
type Repository struct {
    DB *sql.DB
    q  *dbgen.Queries
}

func NewRepository(db *sql.DB) Repository {
    return Repository{DB: db, q: dbgen.New(db)}
}
```

---

## Fehlerbehandlung

Das `backend/db/`-Paket stellt gemeinsame Hilfsfunktionen bereit:

- `db.Error(err)` — Übersetzt PostgreSQL-Fehler in Domain-Sentinel-Errors (`ErrNotFound`, `ErrAlreadyExists`, `ErrDatabase`)
- `db.ResultError(result)` — Prüft `RowsAffected()` für UPDATE-Operationen
- `db.NullTime` — Nullable-Time-Typ für `json_agg()`-Ergebnisse
- `db.Close(closer, name)` — Sicheres Schließen von Ressourcen mit Fehler-Logging

---

## Datenbankschema

Das aktuelle Schema ist über die SQL-Migrationen definiert: `database/migrations/` (alle `*.up.sql`-Dateien in Reihenfolge).

**Tabellen:** `users`, `tables`, `products`, `product_variants`, `events` (append-only)

**Enums:** `UserRole` (admin, senior_service, service), `EntityStatus` (active, inactive, deleted), `ProductCategory` (food, beverage, other)

Neue Migration erstellen: `database/migrations/<nr>_<name>.up.sql` + `.down.sql`

---

## Tests

```bash
cd backend && go test -tags=unit -race ./...         # Unit-Tests (mit Mocks)
cd backend && go test -tags=integration -race ./...  # Integrationstests (gegen PostgreSQL)
```

Integrationstests benötigen eine laufende PostgreSQL-Instanz mit angewendeten Migrationen (siehe [development.md](development.md)).
