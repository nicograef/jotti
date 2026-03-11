---
description: "Use when working on database migrations, SQL queries, sqlc configuration, schema changes, or data model design."
applyTo: "database/**"
---

# Datenbank-Konventionen

## Befehle

- **sqlc generieren:** `make sqlc` (nach Query-Änderungen)
- **Dev-DB starten:** `make dev` (startet PostgreSQL im Docker-Stack)
- **DB-Shell öffnen:** `make db-shell`

## Schema

Tabellen: `users`, `tables`, `products`, `product_variants`, `events` (append-only).

Aktuelles Schema: siehe SQL-Migrationen in `database/migrations/` (alle `*.up.sql`-Dateien in Reihenfolge).

## Migrationen

Neue Migration: `database/migrations/<nr>_<name>.up.sql` + `.down.sql`

Immer reversible Migrationen erstellen — Rollback testen vor dem Merge.

## sqlc

- Konfiguration: `backend/sqlc.yaml`
- Queries: `backend/sqlc/queries/<domain>.sql`
- Generierter Code: `backend/sqlc/dbgen/` (NICHT EDITIEREN)
- Nach Query-Änderungen: `sqlc generate` ausführen
