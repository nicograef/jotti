# Architektur-Theorie — Übersicht

Dieses Verzeichnis enthält **allgemeine theoretische Grundlagen** zu Architekturmustern und Technologien, die in modernen Web-Anwendungen eingesetzt werden. Die Dokumente sind bewusst **technologie- und projekt-unabhängig** gehalten — jotti-spezifische Implementierungsdetails finden sich jeweils im Appendix der zugehörigen Dokumente.

Die Dokumente dienen als Nachschlagewerk für:

- **Architektur-Entscheidungen** — Fundierte Grundlage für ADRs
- **Neue Features** — Welche Patterns passen? Wie integrieren sie sich?
- **Onboarding** — Warum wurde was wie gebaut?
- **Refactoring** — Was ist der Zielzustand, wo stehen wir?

---

## Dateistruktur

```
docs/theory/
├── README.md                # Übersicht und Leseempfehlungen (diese Datei)
├── ddd.md                   # Domain-Driven Design
├── event-sourcing.md        # Event-Sourcing
├── cqrs.md                  # CQRS (Command Query Responsibility Segregation)
├── go-backend.md            # Go Backend Architektur + SQL-Tooling + Testing
├── postgresql.md            # PostgreSQL — Architektur, Indexing, Optimierung
├── react-frontend.md        # React Frontend Architektur, Patterns, Testing
├── security.md              # Security & Authentifizierung
├── devops.md                # DevOps, Deployment & Infrastruktur (geplant)
└── pos.md                   # POS-Systeme & Gastronomie-Domäne (geplant)
```

---

## Dokumente

### [Domain-Driven Design (DDD)](ddd.md)

Strategisches und taktisches Design: Bounded Contexts, Ubiquitous Language, Context Mapping (9 Patterns inkl. ACL, OHS, CF), Sub-Domain-Klassifikation (Core/Supporting/Generic), Aggregates, Entities, Value Objects, Domain Events, Domain Services, Repositories, Factories. Event Storming als Modellierungstechnik. DDD in Kombination mit Hexagonaler Architektur, CQRS und Microservices. Im Appendix: Anwendungsbeispiel jotti.

**Lesenswert wenn:** Domain-Konzepte modelliert werden, Aggregate-Grenzen unklar sind, die Ubiquitous Language erweitert wird, oder Architekturentscheidungen zur Systemaufteilung anstehen.

### [Event-Sourcing](event-sourcing.md)

Event-Sourcing-Grundlagen: Event Store, Replay, Snapshots, State-Rekonstruktion. Event Design (Granularität, Versionierung, Schema-Evolution). Event Store-Technologien (EventStoreDB, PostgreSQL, Kafka). Saga-Pattern und Distributed Transactions. Outbox/Inbox-Pattern für zuverlässige Event-Delivery. Entscheidungsmatrix Event-Sourcing vs. CRUD. Anti-Patterns. Im Appendix: Anwendungsbeispiel jotti.

**Lesenswert wenn:** Event-Typen hinzugefügt oder überarbeitet werden, die Balance zwischen Event-Sourcing und CRUD hinterfragt wird, oder Fragen zur Schema-Evolution aufkommen.

### [CQRS](cqrs.md)

Command Query Responsibility Segregation auf System-Ebene: Trennung von Write- und Read-Modellen. Ausbaustufen (Stufe 0–3, von einfacher CQS bis zu separaten Datenbanken). Read Model Design (Projektionen, Denormalisierung). Projektionsstrategien (synchron, asynchron, hybrid). Eventual Consistency und ihre Implikationen. CQRS ohne Event-Sourcing. Im Appendix: Anwendungsbeispiel jotti.

**Lesenswert wenn:** Projektionen erweitert werden, Read Models optimiert werden, oder Fragen zur Eventual Consistency aufkommen.

### [Go Backend Architektur](go-backend.md)

Schichtenarchitektur (Hexagonal/Clean Architecture): HTTP → Application → Domain → Repository. Middleware-Stack (Auth, Rate Limiting, Logging). Dependency Injection. Go-spezifische Patterns: Error Handling, Concurrency (Goroutines, Channels, sync), Context-Propagation. API-Design-Patterns (REST-Alternativen, Versioning). Resilienz-Patterns (Retry, Circuit Breaker, Timeout). Observability (structured Logging, Tracing, Metriken). Datenbankzugriff: sqlc Deep Dive, sqlc vs. GORM vs. sqlx vs. ent, Migration-Tooling, Repository Pattern. Testing: Table-Driven Tests, Testcontainers, httptest, Mocking. Im Appendix: Anwendungsbeispiel jotti.

**Lesenswert wenn:** Neue Endpunkte gebaut werden, die Schichtentrennung unklar ist, Fehlerbehandlung oder Middleware erweitert wird, oder Entscheidungen zur Datenbankanbindung anstehen.

### [PostgreSQL](postgresql.md)

PostgreSQL-Architektur (MVCC, WAL, Autovacuum). Features: JSONB, Trigger, Enums, IDENTITY Columns. Indexing Deep Dive: B-Tree, GIN, GiST, BRIN, Partial Indexes, Covering Indexes. Query Optimization: EXPLAIN ANALYZE, CTEs, Window Functions, Parallelisierung. Advanced Features: LISTEN/NOTIFY, Partitioning, Materialized Views, Row-Level Security. Connection Management: pgxpool-Konfiguration, PgBouncer. PostgreSQL als Event Store. Zero-Downtime Migration-Strategien. Vergleich mit MySQL, MongoDB und NewSQL (CockroachDB, TiDB).

**Lesenswert wenn:** SQL-Queries geschrieben werden, neue Tabellen oder Migrationen erstellt werden, Performance-Probleme untersucht werden, oder Infrastrukturentscheidungen zur Datenbankwahl anstehen.

### [React Frontend Architektur](react-frontend.md)

Frontend Architecture Patterns: Monolithic, Modular, Micro-Frontend, Flux. State-Management Landscape: TanStack Query (Server State), Zustand (Client State), React Hook Form (Form State), Context API. 15+ React Design Patterns: Component Composition, Custom Hook, Control Props, Provider, Container/Presentational, Compound Components, Headless Components, Render Props, Props Getters, Error Boundary, Portal, HOC, MVVM, Dependency Injection, SOLID-Prinzipien. Testing: Vitest, React Testing Library, MSW (Mock Service Worker), Playwright. Performance: memo, lazy/Suspense, Virtualisierung. Accessibility: ARIA, Keyboard Navigation, Focus Management. TypeScript-Patterns: Discriminated Unions, Generic Components, Branded Types, Template Literal Types. Im Appendix: Anwendungsbeispiel jotti.

**Lesenswert wenn:** Neue Seiten oder Komponenten gebaut werden, State-Management-Fragen aufkommen, UI-Patterns konsistent bleiben sollen, Testing-Strategien gesucht werden, oder Performance-Probleme auftreten.

### [Security & Authentifizierung](security.md)

Authentication-Patterns (Session, JWT, OAuth2/OIDC, Passkeys/WebAuthn). Authorization-Modelle: RBAC, ABAC, ReBAC (Google Zanzibar, OpenFGA). Passwort-Sicherheit: Argon2id, bcrypt, NIST-Empfehlungen. OWASP Top 10. API-Security: Input Validation, CORS, CSRF, Rate Limiting, Security Headers. Frontend-Security: XSS-Prävention, sicheres Token-Handling. Secrets Management. TLS & Certificate Management. Security Testing (SAST, DAST, Dependency Scanning). Entscheidungsmatrix: Auth-Pattern vs. Anwendungsfall. Im Appendix: Anwendungsbeispiel jotti.

**Lesenswert wenn:** Authentifizierungs- oder Autorisierungsmechanismen implementiert oder überarbeitet werden, Security-Reviews durchgeführt werden, oder Fragen zu OWASP-Risiken aufkommen.

### DevOps, Deployment & Infrastruktur _(geplant: devops.md)_

DevOps-Kultur & CALMS-Framework. Containerisierung: Docker (Multi-Stage Builds, Image-Optimierung, Container Security). Orchestrierung: Docker Compose vs. Kubernetes vs. K3s — Entscheidungsmatrix. Reverse Proxy & Load Balancing: nginx, Traefik, Caddy — Vergleich. CI/CD Pipeline Design: Lint → Test → Build → Publish → Deploy, GitHub Actions. Deployment-Strategien: Recreate, Rolling Update, Blue/Green, Canary. Zero-Downtime Deployment: Health Checks, Graceful Shutdown, DB-Migrations. Monitoring & Observability: Prometheus, Grafana, Loki (RED/USE Method). Backup & Disaster Recovery: pg_dump, Restic, RTO/RPO. Self-Hosting Patterns: VM-Provisionierung, DNS, Firewall, automatische Zertifikate.

**Lesenswert wenn:** Deployment-Infrastruktur aufgebaut oder geändert wird, CI/CD-Pipelines konfiguriert werden, Monitoring eingerichtet wird, oder Entscheidungen zur Orchestrierung anstehen.

### POS-Systeme & Gastronomie-Domäne _(geplant: pos.md)_

Geschichte und Entwicklung von POS-Systemen (Registrierkasse → ECR → Cloud-POS → mPOS). POS-Architektur-Patterns: On-Premise, Cloud-POS (SaaS), Hybrid, Mobile POS. Gastronomie-POS: Order Lifecycle, tischbasierter Workflow, Kitchen Display Systems, Split Bills. Datenmodelle: tischbasierter offener Saldo, transaktionsbasiert, Event-Sourcing im POS-Kontext. Fiskalgesetzgebung: TSE/KassenSichV (Deutschland), RKSV (Österreich), Non-Profit-Ausnahmen. Payment Integration: Gateway-Architektur, SumUp/Square/Zettle, NFC/Contactless. Non-Profit vs. Commercial POS: Feature-Abgrenzung, TCO, Setup-Aufwand. POS-Marktlandschaft: kommerzielle Systeme vs. Open-Source-Alternativen. Im Appendix: Anwendungsbeispiel jotti.

**Lesenswert wenn:** Neue Gastro-Workflows modelliert werden, Fragen zur Fiskalkonformität aufkommen, oder POS-Architekturentscheidungen anstehen.

---

## Leseempfehlungen

Je nach Einstiegspunkt empfehlen sich unterschiedliche Lesereihenfolgen:

### Einstieg: Domain-Modellierung

1. [DDD](ddd.md) — Bounded Contexts, Aggregates, Events
2. [Event-Sourcing](event-sourcing.md) — Event Store, Replay, Snapshots
3. [CQRS](cqrs.md) — Projektion, Read Models, Eventual Consistency

### Einstieg: Backend-Implementierung

1. [Go Backend Architektur](go-backend.md) — Schichten, Patterns, Testing
2. [PostgreSQL](postgresql.md) — Queries, Indexing, Migrations
3. [DDD](ddd.md) — Domain-Modell, Repositories

### Einstieg: Frontend-Entwicklung

1. [React Frontend Architektur](react-frontend.md) — Patterns, State, Testing
2. [DDD](ddd.md) — Domain-Begriffe verstehen
3. [CQRS](cqrs.md) — Warum die API so aufgebaut ist (Read vs. Write)

### Einstieg: Infrastruktur & Betrieb _(sobald Dokumente verfügbar)_

1. DevOps & Deployment — Container, CI/CD, Monitoring
2. [Security](security.md) — Auth, OWASP, TLS
3. [PostgreSQL](postgresql.md) — Connection Pooling, Backup

### Vollständige Lesereihenfolge (empfohlen für Onboarding)

1. [DDD](ddd.md) — Domänen-Sprache und Grundkonzepte
2. [Event-Sourcing](event-sourcing.md) — Persistenz-Pattern
3. [CQRS](cqrs.md) — Lese-/Schreibtrennung
4. [Go Backend Architektur](go-backend.md) — Implementierungsschichten
5. [PostgreSQL](postgresql.md) — Datenbank-Grundlagen
6. [React Frontend Architektur](react-frontend.md) — UI-Architektur
7. [Security](security.md) — Querschnittsthema
8. DevOps _(geplant)_ — Betrieb und Deployment
9. POS-Systeme _(geplant)_ — Fachdomäne vertiefen

---

## Abgrenzung zu operativen Dokumenten

| Theorie-Dokumente (dieses Verzeichnis)                              | Operative Dokumente (`docs/`)                            |
| ------------------------------------------------------------------- | -------------------------------------------------------- |
| **Warum** und **Was** — Prinzipien, Muster, Entscheidungsgrundlagen | **Wie** — Konkreter Ist-Zustand, Implementierungsdetails |
| Stabil — Ändert sich selten                                         | Dynamisch — Ändert sich mit dem Code                     |
| Basis für ADRs                                                      | Ergebnis von ADRs                                        |

### Querverweis zu operativen Dokumenten

| Operatives Dokument                               | Beschreibung                                                 |
| ------------------------------------------------- | ------------------------------------------------------------ |
| [CQRS in jotti](../cqrs.md)                       | Ist-Zustand, konkreter Implementierungsplan für Projektionen |
| [Datenbank & Persistenz](../database.md)          | Operative sqlc-Integration, Repository-Pattern               |
| [Ubiquitous Language](../language.md)             | Kanonische Fachbegriffe und Inkonsistenzen                   |
| [Anforderungskatalog](../requirements.md)         | 50 Anforderungen mit Status                                  |
| [Implementierungsplan](../implementation-plan.md) | Nächste Features (Phase 1 & 2)                               |
| [Entwicklung & Deployment](../development.md)     | Setup, Tests, CI/CD                                          |
| [POS-Positionierung](../pos.md)                   | jotti-Positionierung im POS-Markt                            |
| [ADR: Event-Sourcing](../adr/event-sourcing.md)   | Entscheidung pro Event-Sourcing                              |
| [ADR: sqlc](../adr/orm.md)                        | Entscheidung pro sqlc                                        |

---

## Quellen

Die Theorie-Dokumente basieren auf folgenden externen Quellen:

### DDD, Event-Sourcing & CQRS

1. [Event-Driven Architecture in Golang](https://github.com/PacktPublishing/Event-Driven-Architecture-in-Golang) — Event-Sourcing, CQRS, DDD in Go
2. [DDD Foundational Guide](https://spartner.software/kennisbank/domain-driven-design-ddd) — Strategisches & taktisches Design
3. [Event Sourcing Explained (2025)](https://www.baytechconsulting.com/blog/event-sourcing-explained-2025) — Write-Store als Single Source of Truth
4. [Event Sourcing vs. CRUD](https://dev.to/alex_aslam/event-sourcing-vs-crud-when-1000-database-writes-dont-matter-5bpj) — Entscheidungsmatrix
5. [AWS CQRS Pattern](https://docs.aws.amazon.com/prescriptive-guidance/latest/modernization-data-persistence/cqrs-pattern.html) — CQRS-Implementierung

### Go Backend & PostgreSQL

6. [DB Performance 101](https://dev.to/ari-ghosh/db-performance-101-a-practical-deep-dive-into-backend-database-optimization-4cag) — PostgreSQL-Optimierung
7. [PostgreSQL vs MySQL](https://www.bytebase.com/blog/postgres-vs-mysql/) — Feature-Vergleich

### React Frontend

8. [21 React Design Patterns](https://www.perssondennis.com/articles/21-fantastic-react-design-patterns-and-when-to-use-them) — React-Patterns
9. [Modern Frontend Architecture Patterns](https://blog.logrocket.com/guide-modern-frontend-architecture-patterns/) — UI-Architektur
10. [Mastering Atomic Design](https://javascript.plainenglish.io/mastering-atomic-design-a-step-by-step-guide-to-building-scalable-ui-components-60b0d2a94cc3) — Atomic Design

### Security _(für security.md)_

11. [OWASP Top 10](https://owasp.org/www-project-top-ten/) — Die 10 kritischsten Web-Security-Risiken
12. [OWASP Cheat Sheet Series](https://cheatsheetseries.owasp.org/) — Praxis-Checklisten für Security-Aspekte
13. [NIST Digital Identity Guidelines](https://pages.nist.gov/800-63-3/) — Password Policies, MFA, Identity Proofing
14. [OWASP API Security Top 10](https://owasp.org/API-Security/) — API-spezifische Risiken
15. [MDN: Web Security](https://developer.mozilla.org/en-US/docs/Web/Security) — CORS, CSP, CSRF aus Browser-Perspektive
16. [OpenFGA](https://openfga.dev/) — Open-Source Zanzibar-Implementation für ReBAC
17. [Open Policy Agent](https://www.openpolicyagent.org/) — Policy-based Authorization

### DevOps & Deployment _(für devops.md)_

18. [Docker Best Practices](https://docs.docker.com/build/building/best-practices/) — Multi-Stage Builds, Layer Caching, Security
19. [12-Factor App](https://12factor.net/) — Config, Logging, Disposability, Dev/Prod Parity
20. [GitHub Actions Docs](https://docs.github.com/en/actions) — CI/CD Pipeline-Design, Workflow-Syntax
21. [Google SRE Book](https://sre.google/sre-book/table-of-contents/) — Monitoring, Alerting, Incident Response, SLOs
22. [Martin Fowler: Continuous Delivery](https://martinfowler.com/bliki/ContinuousDelivery.html) — CD-Prinzipien

### POS-Systeme _(für pos.md)_

23. [Wikipedia: Point of Sale](https://en.wikipedia.org/wiki/Point_of_sale) — POS-Geschichte, Terminologie
24. [Square Developer Docs](https://developer.squareup.com/docs) — Payment Integration, POS-API-Design
25. [Toast Developer Platform](https://doc.toasttab.com/) — Restaurant POS API, KDS Integration

### Standardwerke

- **Eric Evans** (2003): _Domain-Driven Design_ — Grundlagen DDD
- **Greg Young** (2010): _CQRS Documents_ — CQRS + Event-Sourcing
- **Bertrand Meyer** (1988): _Object-Oriented Software Construction_ — CQS-Prinzip
- **Martin Fowler**: Artikel zu DDD, Event-Sourcing, CQRS
- **Vaughn Vernon** (2013): _Implementing Domain-Driven Design_ — Taktisches DDD in der Praxis
