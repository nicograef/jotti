---
description: "Use when working on database migrations, SQL queries, sqlc configuration, schema changes, or data model design."
applyTo: "database/**,backend/sqlc/**,backend/sqlc.yaml"
---

> **Referenz:** Tabellen-Schemata stehen kanonisch in `database/migrations/01_initial.up.sql`. Für Architektur und Invarianten → `docs/handbuch.md` §3.2 (Kassenjournal) und §4 (Stammdaten). Für DB-Spalten-Konventionen → `docs/language.md`.

Repo-weite Regeln und Guardrails stehen kanonisch in `AGENTS.md`. Diese Datei ergänzt nur datenbankspezifische Konventionen für `database/**`, `backend/sqlc/**` und `backend/sqlc.yaml`.

# Datenbank-Konventionen

## Befehle

- **sqlc generieren:** `make sqlc` (nach Query-Änderungen)
- **Dev-DB starten:** `make dev` (startet PostgreSQL im Docker-Stack)
- **DB-Shell öffnen:** `make db-shell`

## Schema

Domain-Tabellen: `tische`, `produkte`, `produkt_varianten` (deutsch, snake_case).
Infrastruktur-Tabellen: `users` (englisch, snake_case).
Kasse-Tabellen: `kassenjournal` (append-only Event Store), `tisch_sessions` (Tisch-Projektion), `kassensitzungen` (Kassensitzung-Entität, CRUD).

Domain-Spalten sind deutsch (`kategorie`, `preis_cents`, `produkt_id`), technische Spalten englisch (`created_at`, `updated_at`, `status`, `id`).

Aktuelles Schema: siehe SQL-Migrationen in `database/migrations/` (alle `*.up.sql`-Dateien in Reihenfolge).

## Schema-Änderungen

Die repo-weite Schema-Policy ist in `AGENTS.md` unter „Freeze-Disziplin" kanonisch beschrieben (neue Migrationen: `database/migrations/README.md`). Diese Datei ergänzt nur die DB-spezifischen Arbeitsschritte: Dev-DB bei Bedarf mit `make down && make dev` neu aufsetzen und nach Query-Änderungen `make sqlc` ausführen.

## sqlc

- Konfiguration: `backend/sqlc.yaml`
- Queries: `backend/sqlc/queries/<domain>.sql`
- Generierter Code: `backend/sqlc/dbgen/`
- Nach Query-Änderungen: `sqlc generate` ausführen
