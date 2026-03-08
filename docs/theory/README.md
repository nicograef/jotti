# Architektur-Theorie — Übersicht

Dieses Verzeichnis enthält die **theoretischen Grundlagen** der Architekturmuster und Technologien, die jotti einsetzt. Die Dokumente dienen als Nachschlagewerk für:

- **Neue Features** — Welche Patterns passen? Wie integrieren sie sich?
- **Architecture Decision Records (ADRs)** — Fundierte Entscheidungsgrundlage
- **Onboarding** — Warum wurde was wie gebaut?
- **Refactoring** — Was ist der Zielzustand, wo stehen wir?

---

## Dokumente

### [Domain-Driven Design (DDD)](ddd.md)

Strategisches und taktisches Design: Bounded Contexts, Aggregates, Entities, Value Objects, Domain Events, Domain Services, Repositories. Erklärt, wie jotti DDD pragmatisch einsetzt — umfassend für den Kassenbetrieb, minimal für Stammdaten.

**Lesenswert wenn:** Neue Domain-Konzepte modelliert werden, Aggregate-Grenzen unklar sind, die Ubiquitous Language erweitert wird.

### [Event-Sourcing](event-sourcing.md)

Event-Sourcing-Theorie: Grundkonzepte (Event Store, Replay, Snapshots), Entscheidungsmatrix Event-Sourcing vs. CRUD, Patterns und Anti-Patterns. Im Appendix: Anwendungsbeispiel jotti.

**Lesenswert wenn:** Event-Typen hinzugefügt werden, die Balance zwischen Event-Sourcing und CRUD hinterfragt wird.

### [CQRS](cqrs.md)

CQRS-Theorie: Command/Query Separation auf System-Ebene, Ausbaustufen (Stufe 0–3), Projektionsstrategien (synchron/asynchron/hybrid), Kombination mit Event-Sourcing. Im Appendix: Anwendungsbeispiel jotti.

**Lesenswert wenn:** CQRS erweitert wird (Projektionen), Read Models optimiert werden.

### [Go Backend Architektur](go-backend.md)

Schichtenarchitektur (HTTP → Application → Domain → Repository), Middleware-Stack, Dependency Injection, Fehlerbehandlung, Validierung, Auth, Testing. Erklärt die Go-spezifischen Designentscheidungen.

**Lesenswert wenn:** Neue Endpunkte gebaut werden, die Schichtentrennung unklar ist, Fehlerbehandlung oder Middleware erweitert wird.

### [PostgreSQL](postgresql.md)

PostgreSQL-spezifische Features (JSONB, Trigger, Enums, IDENTITY), sqlc-Workflow (SQL → Go), hybride Persistenz (CRUD + Event Store), Performance-Optimierung, Migrations-Strategie.

**Lesenswert wenn:** SQL-Queries geschrieben werden, neue Tabellen/Migrationen erstellt werden, Performance-Probleme untersucht werden.

### [React Frontend Architektur](react-frontend.md)

Komponentenstruktur (Atomic Design), State-Management (Hooks + Singletons, kein Redux), Backend-Integration (BackendClient), Routing und Guards, Zod-Validierung, UI-Patterns (Drawers, Toasts), Tailwind Styling.

**Lesenswert wenn:** Neue Seiten/Komponenten gebaut werden, State-Management-Fragen aufkommen, UI-Patterns konsistent bleiben sollen.

---

## Abgrenzung zu operativen Dokumenten

| Theorie-Dokumente (dieses Verzeichnis)                              | Operative Dokumente (`docs/`)                            |
| ------------------------------------------------------------------- | -------------------------------------------------------- |
| **Warum** und **Was** — Prinzipien, Muster, Entscheidungsgrundlagen | **Wie** — Konkreter Ist-Zustand, Implementierungsdetails |
| Stabil — Ändert sich selten                                         | Dynamisch — Ändert sich mit dem Code                     |
| Basis für ADRs                                                      | Ergebnis von ADRs                                        |

### Querverweis zu operativen Dokumenten

| Operatives Dokument                                     | Beschreibung                                                 |
| ------------------------------------------------------- | ------------------------------------------------------------ |
| [CQRS in jotti](../cqrs.md)                             | Ist-Zustand, konkreter Implementierungsplan für Projektionen |
| [Event-Sourcing vs. CRUD](../event-sourcing-vs-crud.md) | Detaillierter Vergleich mit 8-Tabellen-CRUD-Alternative      |
| [Datenbank & Persistenz](../database.md)                | Operative sqlc-Integration, Repository-Pattern               |
| [Ubiquitous Language](../language.md)                   | Kanonische Fachbegriffe und Inkonsistenzen                   |
| [Anforderungskatalog](../requirements.md)               | 50 Anforderungen mit Status                                  |
| [Implementierungsplan](../implementation-plan.md)       | Nächste Features (Phase 1 & 2)                               |
| [Entwicklung & Deployment](../development.md)           | Setup, Tests, CI/CD                                          |
| [ADR: Event-Sourcing](../adr/event-sourcing.md)         | Entscheidung pro Event-Sourcing                              |
| [ADR: sqlc](../adr/orm.md)                              | Entscheidung pro sqlc                                        |

---

## Quellen

Die Theorie-Dokumente basieren auf folgenden externen Quellen:

1. [Event-Driven Architecture in Golang](https://github.com/PacktPublishing/Event-Driven-Architecture-in-Golang) — Event-Sourcing, CQRS, DDD in Go
2. [DDD Foundational Guide](https://spartner.software/kennisbank/domain-driven-design-ddd) — Strategisches & taktisches Design
3. [Event Sourcing Explained (2025)](https://www.baytechconsulting.com/blog/event-sourcing-explained-2025) — Write-Store als Single Source of Truth
4. [Event Sourcing vs. CRUD](https://dev.to/alex_aslam/event-sourcing-vs-crud-when-1000-database-writes-dont-matter-5bpj) — Entscheidungsmatrix
5. [AWS CQRS Pattern](https://docs.aws.amazon.com/prescriptive-guidance/latest/modernization-data-persistence/cqrs-pattern.html) — CQRS-Implementierung
6. [DB Performance 101](https://dev.to/ari-ghosh/db-performance-101-a-practical-deep-dive-into-backend-database-optimization-4cag) — PostgreSQL-Optimierung
7. [PostgreSQL vs MySQL](https://www.bytebase.com/blog/postgres-vs-mysql/) — Feature-Vergleich
8. [21 React Design Patterns](https://www.perssondennis.com/articles/21-fantastic-react-design-patterns-and-when-to-use-them) — React-Patterns
9. [Modern Frontend Architecture Patterns](https://blog.logrocket.com/guide-modern-frontend-architecture-patterns/) — UI-Architektur
10. [Mastering Atomic Design](https://javascript.plainenglish.io/mastering-atomic-design-a-step-by-step-guide-to-building-scalable-ui-components-60b0d2a94cc3) — Atomic Design

Sowie die Standardwerke:

- **Eric Evans** (2003): _Domain-Driven Design_ — Grundlagen DDD
- **Greg Young** (2010): _CQRS Documents_ — CQRS + Event-Sourcing
- **Bertrand Meyer** (1988): _Object-Oriented Software Construction_ — CQS-Prinzip
- **Martin Fowler**: Artikel zu DDD, Event-Sourcing, CQRS
