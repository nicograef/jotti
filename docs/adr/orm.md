# ADR: Datenbankzugriff — sqlc als Persistenz-Werkzeug

## Status

**Entschieden** — sqlc wurde als Persistenz-Werkzeug für jotti übernommen.

## Kontext

jotti benötigt eine Persistenzschicht für zwei Strategien:

1. **CRUD für Stammdaten** (Benutzer, Tische, Produkte) — relationale Tabellen mit Soft-Deletes
2. **Event-Sourcing für Tisch-Operationen** (Bestellungen, Zahlungen, Lieferungen, Stornierungen) — Append-Only-Events

Vor sqlc: handgeschriebene SQL-Queries mit `database/sql` + `pgx/v5`. Funktional, aber manuelles `row.Scan()`-Boilerplate und keine Compile-Time-Validierung gegen das Schema.

## Bewertete Alternativen

| Kriterium                | Status Quo (`pgx/v5`) | GORM             | sqlx        | **sqlc**        |
| ------------------------ | --------------------- | ---------------- | ----------- | --------------- |
| Event-Sourcing           | ✅                    | ❌ Inkompatibel  | ✅          | ✅              |
| PostgreSQL-natives SQL   | ✅                    | ❌ `db.Raw()`    | ✅          | ✅              |
| Data-Mapper-Architektur  | ✅                    | ❌ Active Record | ✅          | ✅              |
| Compile-Time-Validierung | ❌                    | ❌               | ❌          | ✅              |
| Schema-Drift-Schutz      | ❌                    | ⚠️               | ❌          | ✅              |
| Boilerplate-Reduktion    | —                     | ~170 Zeilen      | ~100 Zeilen | **~300 Zeilen** |
| Laufzeit-Abhängigkeiten  | 0                     | +2               | +1          | **0**           |

**GORM** scheitert an Event-Sourcing (Lifecycle-Modell setzt veränderliche Datensätze voraus), erzwingt Active Record und kann CTEs/`json_agg()` nicht ausdrücken.

**sqlx** bietet nur moderate Einsparungen (~100 Zeilen) ohne Compile-Time-Validierung; `pgx` bietet bereits ähnliche Features.

## Entscheidung

**sqlc** — ein SQL-Compiler, der annotierte `.sql`-Dateien gegen das Migrationsschema parst und typsichere Go-Funktionen generiert.

**Begründung:**

- SQL-first, PostgreSQL-nativ, zero Runtime-Dependency
- Eliminiert `Scan()`-Boilerplate bei Compile-Time-Schema-Validierung
- Voll kompatibel mit Event-Sourcing, Data-Mapper-Pattern und bestehender Mock-Strategie

## Konsequenzen

**Positiv:** Compile-Time-Validierung, ~300 Zeilen weniger Boilerplate, typsichere Queries, keine neue Laufzeitabhängigkeit.

**Negativ:** Build-Schritt `make sqlc` bei Schema-/Query-Änderungen nötig, Custom-Type-Konfiguration in `sqlc.yaml`, generierte Dateien (`sqlc/dbgen/`) nicht editierbar.

### Dateien

- `backend/sqlc.yaml` — Konfiguration
- `backend/sqlc/queries/` — SQL-Query-Definitionen
- `backend/sqlc/dbgen/` — Generierter Code (NICHT EDITIEREN)
