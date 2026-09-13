---
description: 'Use when working on database migrations, SQL queries, sqlc configuration, schema changes, or data model design.'
applyTo: 'database/**,backend/sqlc/**,backend/sqlc.yaml'
---

Repo-weite Regeln und Guardrails: `AGENTS.md` (Freeze-Disziplin). Das Schema steht kanonisch in `database/migrations/*.up.sql`. Architektur und Invarianten: [docs/handbuch.md](../../docs/handbuch.md) §3.2 (Kassenjournal) und §4 (Stammdaten).

Neue Migrationen folgen [database/migrations/README.md](../../database/migrations/README.md) (forward-only, additiv, `01_initial.up.sql` eingefroren). DB-Spalten-Konventionen: [docs/language.md](../../docs/language.md).
