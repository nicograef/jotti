---
description: "Use when working on database migrations, SQL queries, sqlc configuration, schema changes, or data model design."
applyTo: "database/**,backend/sqlc/**,backend/sqlc.yaml"
---

> **Referenz:** Für Stammdaten-Schema (Produkte, Tische, Benutzer) → `docs/handbuch.md` §4. Für Event-Store-Schema → `docs/handbuch.md` §3.4. Für DB-Spalten-Konventionen → `docs/language.md`.

# Datenbank-Konventionen

## Befehle

- **sqlc generieren:** `make sqlc` (nach Query-Änderungen)
- **Dev-DB starten:** `make dev` (startet PostgreSQL im Docker-Stack)
- **DB-Shell öffnen:** `make db-shell`

## Schema

Domain-Tabellen: `tische`, `produkte`, `produkt_varianten` (deutsch, snake_case).
Infrastruktur-Tabellen: `users`, `events` (englisch, snake_case).

Domain-Spalten sind deutsch (`kategorie`, `preis_cents`, `produkt_id`), technische Spalten englisch (`created_at`, `updated_at`, `status`, `id`).

Aktuelles Schema: siehe SQL-Migrationen in `database/migrations/` (alle `*.up.sql`-Dateien in Reihenfolge).

## Schema-Änderungen

jotti befindet sich in aktiver Entwicklung — Breaking Changes sind erlaubt und erwünscht. Schema-Änderungen direkt in `database/migrations/01_initial.up.sql` vornehmen. Keine neuen Migrationsdateien erstellen, keine Down-Migrationen pflegen. Dev-Datenbank bei Bedarf neu aufsetzen (`make down && make dev`).

## sqlc

- Konfiguration: `backend/sqlc.yaml`
- Queries: `backend/sqlc/queries/<domain>.sql`
- Generierter Code: `backend/sqlc/dbgen/` (NICHT EDITIEREN)
- Nach Query-Änderungen: `sqlc generate` ausführen
