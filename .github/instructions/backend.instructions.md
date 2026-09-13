---
description: 'Use when working on Go backend code, API handlers, middleware, repositories, domain models, or application services.'
applyTo: 'backend/**'
---

Repo-weite Regeln und Guardrails: `AGENTS.md`. Schichtenarchitektur, API-Design, Validierung, OCC: [docs/handbuch.md](../../docs/handbuch.md) Kap. 6. Namenskonventionen pro Schicht: [docs/language.md](../../docs/language.md).

Der Request-Context trägt `UserIDKey` und `UserNameKey` (Name ebenfalls aus dem Benutzer-Datensatz), aber keine Rolle. Die Autorisierung prüft die Rolle live aus dem Benutzer-Datensatz, nicht aus dem Token-Claim — eine Rollenänderung wirkt sofort.
