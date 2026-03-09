# Überarbeitungsplan: `docs/theory/` — Von Projekt-Doku zu allgemeinem Architektur-Nachschlagewerk

## Ziel

Die fünf Dokumente in `docs/theory/` (sechs nach der Aufspaltung von ES/CQRS) sind aktuell stark auf jotti zugeschnitten. Jede Erklärung wird mit jotti-Beispielen illustriert, jotti-spezifische Implementierungsdetails sind direkt in die Theorie eingewoben, und die Themenbreite orientiert sich an dem, was jotti konkret nutzt — nicht an dem, was man als Full-Stack-Architekt über die jeweiligen Themen wissen sollte.

**Zielzustand:** Jedes Dokument wird zu einem **eigenständigen, allgemeinen Architektur-Guide**, der ohne Vorkenntnisse über jotti lesbar ist. Projekt-spezifische Anwendungsbeispiele werden klar abgetrennt (eigener Abschnitt am Ende oder separate Datei). Die Dokumente sollen breiter, tiefgründiger und mit mehr externen Quellen untermauert sein.

---

## Analyse des Ist-Zustands

### Gemeinsame Probleme aller Dokumente

| Problem                                         | Beschreibung                                                                                                |
| ----------------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| **Jotti-zentrierte Beispiele**                  | Theorie wird fast ausschließlich mit jotti-Code illustriert. Allgemeine Beispiele fehlen.                   |
| **Vermischung von Theorie und Implementierung** | Abschnitte wie „Ist-Zustand in jotti" stehen direkt neben theoretischen Erklärungen.                        |
| **Eingeschränkte Themenbreite**                 | Nur Patterns aufgeführt, die jotti nutzt. Alternativen, die _nicht_ gewählt wurden, werden kaum beleuchtet. |
| **Wenige externe Quellen**                      | Referenz-Abschnitte listen 3-5 Links. Ein allgemeines Nachschlagewerk braucht deutlich mehr.                |
| **Fehlende Entscheidungsframeworks**            | Wann nutzt man Pattern A vs. B? Entscheidungsmatrizen sind dünn oder fehlen ganz.                           |
| **Kein High-Level-Überblick**                   | Die Dokumente starten direkt mit Details, ohne den breiteren architektonischen Kontext zu setzen.           |

### Datei-spezifische Analyse

#### 1. `ddd.md` (413 Zeilen)

**Gut:** Strategisches & taktisches Design, Bounded Contexts, Aggregates, Context Mapping.
**Lücken:**

- **Event Storming** als Modellierungstechnik fehlt komplett
- **Sub-Domain-Klassifikation** (Core, Supporting, Generic) nur angedeutet
- **DDD in Microservices vs. Monolith** — Abgrenzung fehlt
- **Factories, Specifications, Domain Services** als eigene Patterns unterbelichtet
- **Strategic Patterns** (Published Language, Open Host Service, Separate Ways, Partnership) nur als Tabelle
- **DDD-Lifecycle**: Wie entwickelt sich ein DDD-Modell über die Zeit?
- Alle Beispiele sind jotti-Tisch/Position/Bestellung

#### 2. `event-sourcing-cqrs.md` (623 Zeilen) → **Aufspaltung in `event-sourcing.md` + `cqrs.md`**

**Entscheidung:** Event-Sourcing und CQRS sind zwei **unabhängige Patterns**, die auch einzeln funktionieren. CQRS geht ohne ES (z.B. mit Read Replicas), ES geht ohne CQRS (einfaches Replay). Die Vermischung ist ein häufiges Missverständnis (Fowler, Young, Dahan betonen das). Die Aufspaltung ermöglicht gezieltere Lektüre und klarere Entscheidungsgrundlagen.

**Gut (bisher):** Grundlagen ES + CQRS, Projektionsstrategien, Entscheidungsmatrix ES vs. CRUD, Anti-Patterns.

**Lücken `event-sourcing.md` (neu):**

- **Event Store-Technologien** (EventStoreDB, Marten, Axon, PostgreSQL, DynamoDB) — Vergleich fehlt
- **Event-Schema-Evolution** nur angerissen (Upcasting). Fehlt: Lazy Migration, Versioned Serializer, Schema Registry
- **Saga/Process Manager Pattern** als fortgeschrittenes ES-Pattern
- **Outbox Pattern** für reliable Event Publishing
- **Event Design**: Thin vs. Fat Events, Event Granularity, Domain Events vs. Integration Events
- **Reale Beispiele** aus Industrie (Banken, E-Commerce, Logistics) fehlen
- ~40% des Dokuments sind jotti-Implementierungsdetails (Abschnitte 5-6)

**Lücken `cqrs.md` (neu):**

- **CQRS ohne Event-Sourcing** — eigenständiger Wert von CQRS wird unterschätzt
- **Read Model Design**: Denormalisierung, Materialized Views, Search Indexes
- **Projektionsstrategien** im Detail: synchron, asynchron, hybrid, Change Data Capture
- **Eventual Consistency** — Strategien zur Handhabung (Compensation, Read-Your-Writes)
- **CQRS-Ausbaustufen** als eigenständiges Kapitel (Stufe 0-3)
- **Entscheidungsmatrix**: Wann CQRS einsetzen, wann nicht?

#### 3. `go-backend.md` (703 Zeilen)

**Gut:** Schichtenarchitektur, DI, Middleware, Handler-Pattern, Fehlerbehandlung, Testing.
**Lücken:**

- **Hexagonal Architecture / Ports & Adapters** als alternatives Architekturmuster
- **Clean Architecture (Uncle Bob)** — Vergleich mit der jotti-Schichtenarchitektur
- **Go-spezifische Concurrency-Patterns** (goroutines, channels, worker pools)
- **Error Handling Patterns** über „explizite Fehler" hinaus (sentinel errors, error wrapping, custom error types, pkg/errors vs. Go 1.13+ errors)
- **Middleware-Patterns** im Allgemeinen (Chain of Responsibility, Decorator)
- **Alternative Router** (Chi, Gorilla, Echo) — Stärken/Schwächen auch wenn jotti stdlib nutzt
- **gRPC, GraphQL** als API-Alternativen zu REST/POST-only
- **Configuration Patterns** (12-Factor App, Viper, envconfig)
- **Graceful Degradation & Circuit Breaker** als Resilienz-Patterns
- Fast gesamtes Dokument ist jotti-Architektur, wenig allgemeine Go-Backend-Theorie

#### 4. `postgresql-sqlc.md` (581 Zeilen) → Aufspaltung in `postgresql.md` + sqlc-Integration in `go-backend.md`

**Begründung der Aufspaltung:** PostgreSQL ist ein eigenständiges, umfangreiches Thema (Architektur, Indexing, MVCC, Advanced Features), das unabhängig von der Wahl des SQL-Toolings steht. sqlc hingegen ist ein Go-spezifisches Tool für den Datenbankzugriff und gehört thematisch in den Go-Backend-Architektur-Guide, wo es im Kontext anderer Go-Tooling-Alternativen (GORM, sqlx, ent, Jet) besser eingeordnet werden kann.

**`postgresql.md` (neu) — Lücken:**

- **PostgreSQL Advanced Features**: Partitioning, Materialized Views, LISTEN/NOTIFY, Foreign Data Wrappers, Row-Level Security
- **Indexing Deep Dive**: B-Tree vs. GIN vs. GiST vs. BRIN, Partial Indexes, Covering Indexes, Index-Only Scans
- **Query Optimization**: EXPLAIN ANALYZE lesen, Query Planner verstehen, Common Table Expressions (CTEs)
- **Connection Pooling**: PgBouncer vs. pgxpool vs. Odyssey — wann welches?
- **MVCC und Vacuum**: Wie PostgreSQL Concurrency handhabt
- **Backup und Recovery**: pg_dump, WAL, Point-in-Time Recovery
- **PostgreSQL für Event-Sourcing**: Vergleich mit spezialisierten Event Stores
- **Schema-Migration-Strategien**: Zero-Downtime Migrations, expand/contract pattern

**sqlc → `go-backend.md` — Lücken (dort als neuer Abschnitt „Datenbankzugriff / SQL-Tooling"):**

- **SQL-Tooling Landschaft in Go**: sqlc, sqlx, GORM, ent, Jet — erweiterter Vergleich mit Entscheidungsmatrix
- **sqlc-Workflow vertieft**: Query-Design, Custom Types, Batch Operations, Prepared Statements
- **Migration-Tooling**: golang-migrate, Atlas, goose — Vergleich und Best Practices

#### 5. `react-frontend.md` (732 Zeilen)

**Gut:** Atomic Design, State-Management, Backend-Integration, UI-Patterns, Styling.
**Lücken:**

- **React Design Patterns Katalog** (21+ Patterns): Nur Container/Presentational, Custom Hook, Compound Components abgedeckt. Fehlen: Control Props, Provider Pattern, Headless Components, Render Props (nur erwähnt), Props Getters, HOC (Legacy), MVVM
- **Frontend Architecture Patterns**: Monolithic, Modular, Micro-Frontends, Feature-Sliced Design — nur am Rande
- **State-Management Landscape**: Vergleich React Query/TanStack Query, SWR, Zustand, Jotai, Redux Toolkit — warum _nicht_ Redux?
- **Testing-Strategien**: Unit Tests (Vitest), Component Tests (Testing Library), E2E (Playwright/Cypress)
- **Performance-Patterns**: React.memo, useMemo, useCallback, Suspense, lazy loading, Code Splitting
- **Accessibility (a11y)**: ARIA, Keyboard Navigation, Screen Reader Support
- **Formulare**: React Hook Form, Formik, native Forms — Patterns und Trade-offs
- **Error Handling**: Error Boundaries im Detail, Suspense for Error Handling
- **TypeScript-Patterns**: Discriminated Unions, Template Literal Types, Branded Types, Generics

#### 6. Fehlendes Dokument: `security.md` (neu)

**Begründung:** Kein bestehendes Theory-Dokument behandelt Security als eigenständiges Thema. jotti nutzt JWT (HS256), Argon2id-Hashing und RBAC — die architektonische Theorie dahinter ist nirgends dokumentiert. Security ist querschnittlich (Backend, Frontend, Infrastruktur) und verdient ein eigenes Dokument.

**Fehlende Themen:**

- **Authentication-Patterns**: JWT vs. Session vs. OAuth2/OIDC vs. API Keys — Entscheidungsmatrix
- **Authorization-Modelle**: RBAC, ABAC, ReBAC, Policy-based (OPA/Cedar)
- **Passwort-Hashing**: Argon2id vs. bcrypt vs. scrypt — Algorithmen-Vergleich, Parameterwahl
- **OWASP Top 10**: Injection (SQL, XSS, Command), Broken Access Control, Cryptographic Failures, SSRF
- **API-Security**: CORS, CSRF, Rate Limiting, Input Validation, Content Security Policy
- **Secrets Management**: Environment Variables, Vault, Sealed Secrets
- **TLS/HTTPS**: Certificate Management, Let's Encrypt, HSTS
- **Frontend Security**: XSS-Prävention (React-inherent), Sanitization, Secure Storage (Token-Handling)

#### 7. Fehlendes Dokument: `devops.md` (neu)

**Begründung:** jotti nutzt Docker Compose, nginx Reverse Proxy, Let's Encrypt und CI/CD — die Theorie hinter Containerisierung, Deployment-Strategien und Infrastructure as Code fehlt komplett in `docs/theory/`. Ein eigenständiges Dokument vermeidet, dass DevOps-Themen in `go-backend.md` oder `postgresql.md` fragmentiert werden.

**Fehlende Themen:**

- **Containerisierung**: Docker Basics, Multi-Stage Builds, Image-Optimierung, Docker Compose vs. Kubernetes
- **Deployment-Strategien**: Blue/Green, Canary, Rolling Update, Recreate — Entscheidungsmatrix
- **Reverse Proxy**: nginx, Traefik, Caddy — Vergleich, TLS Termination, Load Balancing
- **CI/CD**: GitHub Actions, GitLab CI, Drone — Pipeline-Design, Test → Build → Deploy
- **Infrastructure as Code**: Terraform, Ansible, Docker Compose als IaC-Light
- **Monitoring & Alerting**: Prometheus + Grafana, Uptime Monitoring, Log Aggregation
- **Backup-Strategien**: Datenbank-Backups, Volume-Backups, Disaster Recovery
- **Zero-Downtime Deployment**: Health Checks, Graceful Shutdown, Connection Draining
- **Self-Hosting Patterns**: VM-Setup, DNS, Firewall, automatische Zertifikate

#### 8. Fehlendes Dokument: `pos.md` (neu — Theory-Version)

**Begründung:** `docs/pos.md` existiert bereits als jotti-spezifische Positionierung. Für `docs/theory/` fehlt eine **allgemeine POS-Theorie**, die das Domänenwissen unabhängig von jotti aufbereitet: Was ist ein POS-System architektonisch? Welche Patterns gibt es? Wie unterscheiden sich Retail- vs. Gastro-POS?

**Gut (in `docs/pos.md`):** POS-Definition, Gastro-spezifische Anforderungen, jotti-Positionierung, Produktvergleich.
**Lücken für ein Theory-Dokument:**

- **POS-Architektur-Patterns**: Cloud-POS vs. On-Premise, Offline-First, Hybrid
- **Gastro-POS-Workflows**: Order-Lifecycle, Kitchen Display Systems, Bon-Druck, Split Bills, Tab Management
- **Datenmodelle**: Tischbasiert vs. Kassenvorgangsbasiert, offener Saldo vs. sofortige Zahlung
- **Fiskalgesetzgebung**: TSE, GoBD, Finanzamtkonformität — warum relevant (auch wenn jotti es nicht braucht)
- **Event-Sourcing im POS-Kontext**: Kassenjournal, Audit Trail, Manipulationssicherheit
- **Non-Profit vs. Commercial POS**: Feature-Abgrenzung, Self-Hosting vs. SaaS, Total Cost of Ownership
- **Mobile POS (mPOS)**: Smartphone-basierte Systeme, PWA vs. Native, Offline-Sync
- **Payment Integration**: Kartenlesegeräte, NFC, Payment Gateways — Architektur-Überblick (auch wenn jotti es nicht implementiert)

---

## Überarbeitungsplan — Iterative Instruktionen

### Instruktion 1: Struktur-Refactoring — Theorie von Anwendung trennen

**Ziel:** In jedem Dokument die jotti-spezifischen Abschnitte identifizieren und in einen klar abgetrennten Appendix am Ende verschieben oder in die operativen Docs (`docs/`) auslagern.

**Schritte:**

1. `event-sourcing-cqrs.md` in zwei Dateien aufteilen: `event-sourcing.md` und `cqrs.md`. Abschnitte 1, 4, 5.x (ES), 7, 8 → `event-sourcing.md`. Abschnitte 2, 3, 6.x (CQRS) → `cqrs.md`. Jedes Dokument bekommt einen Abschnitt „Kombination mit CQRS/ES" mit Cross-Referenz.
2. `postgresql-sqlc.md` umbenennen in `postgresql.md`. sqlc-Abschnitte extrahieren und als Notizen für Integration in `go-backend.md` (Instruktion 4) bereithalten.
3. In jedem der 6 bestehenden Dokumente alle Abschnitte markieren, die `jotti` im Titel tragen oder jotti-spezifische Implementierungsdetails enthalten
4. Für jedes Dokument einen Abschnitt `## Appendix: Anwendungsbeispiel (jotti)` am Ende erstellen
5. Jotti-spezifische Inhalte dorthin verschieben
6. Im Theorie-Teil stattdessen **generische Beispiele** verwenden (E-Commerce-Shop, Banking, Ticket-System, etc.)
7. README.md aktualisieren: Beschreibungen anpassen (nicht mehr „in jotti" sondern „allgemein")

**Quellen für diesen Schritt:**

- Alle 6 Dateien in `docs/theory/` (nach Aufspaltung)
- Operative Dokmente in `docs/` (zum Prüfen, was bereits dort dokumentiert ist)

---

### Instruktion 2: `ddd.md` — Erweitern zu einem vollständigen DDD-Guide

**Ziel:** Von 413 Zeilen jotti-DDD zu einem ~800-1000 Zeilen allgemeinen DDD-Nachschlagewerk.

**Neue/erweiterte Abschnitte:**

1. **Einleitung: Warum DDD?** — Historischer Kontext (Eric Evans 2003, Vaughn Vernon 2013), Problemstellung (Complexity), Abgrenzung zu anderen Architekturansätzen
2. **Strategisches Design erweitern:**
   - Sub-Domain-Klassifikation: Core Domain, Supporting Sub-Domain, Generic Sub-Domain — mit Entscheidungshilfe
   - Context Mapping Patterns vollständig: Partnership, Shared Kernel, Customer/Supplier, Conformist, Anti-Corruption Layer, Open Host Service, Published Language, Separate Ways
   - **Event Storming** als Modellierungsmethode: Big Picture, Process Level, Design Level
3. **Taktisches Design erweitern:**
   - **Factories**: Wann und wie (Factory Method vs. Abstract Factory vs. Builder in DDD)
   - **Specifications**: Encapsulation komplexer Geschäftsregeln
   - **Domain Services vs. Application Services**: Klare Abgrenzung mit allgemeinen Beispielen
   - **Module / Package Design**: Wie strukturiert man Bounded Contexts im Code?
4. **DDD und Architekturstile:**
   - DDD in Monolithen vs. Microservices
   - DDD mit Event-Sourcing
   - DDD mit CQRS
   - DDD mit Hexagonal Architecture
5. **DDD-Lifecycle:** Wie evoliert ein Domain-Modell? Refactoring Toward Deeper Insight.
6. **Entscheidungsframework:** Wann lohnt DDD? Flowchart/Matrix erweitern.

**Quellen für diesen Schritt:**

| Quelle                                               | URL                                                                    | Fokus                                                                                        |
| ---------------------------------------------------- | ---------------------------------------------------------------------- | -------------------------------------------------------------------------------------------- |
| Spartner DDD Guide                                   | https://spartner.software/kennisbank/domain-driven-design-ddd          | Strategisches + taktisches Design kompakt, FAQ-Perspektiven (DDD in Legacy, DDD ohne Buy-in) |
| Event-Driven Architecture in Golang (Buch-Repo)      | https://github.com/PacktPublishing/Event-Driven-Architecture-in-Golang | DDD + ES + CQRS in Go, Aggregate-Design, Domain Event Patterns                               |
| Eric Evans: Domain-Driven Design (2003)              | Buch-Referenz                                                          | Das Grundlagenwerk — Begriffe, Muster, Philosophie                                           |
| Vaughn Vernon: Implementing DDD (2013)               | Buch-Referenz                                                          | Praxisnahe Umsetzung mit Code-Beispielen                                                     |
| Vaughn Vernon: Domain-Driven Design Distilled (2016) | Buch-Referenz                                                          | Kurzfassung für Einsteiger und Manager                                                       |
| Martin Fowler: DDD Tag                               | https://martinfowler.com/tags/domain%20driven%20design.html            | Artikel-Sammlung: Aggregate, Bounded Context, Ubiquitous Language                            |
| Alberto Brandolini: Event Storming                   | https://www.eventstorming.com/                                         | Event Storming als Modellierungsmethode                                                      |
| Nick Tune: Domain-Driven Architecture                | https://medium.com/nick-tune-tech-strategy-blog                        | Strategic DDD, Context Mapping in der Praxis                                                 |
| DDD Crew: Context Mapping                            | https://github.com/ddd-crew/context-mapping                            | Visuelle Templates für Context Maps                                                          |
| DDD Crew: Bounded Context Canvas                     | https://github.com/ddd-crew/bounded-context-canvas                     | Tool für das Definieren von Bounded Contexts                                                 |

---

### Instruktion 3a: `event-sourcing.md` (neu) — Eigenständiger Event-Sourcing-Guide

**Ziel:** Aus dem ES-Teil von `event-sourcing-cqrs.md` ein eigenständiges ~600-800 Zeilen Dokument.

**Neue/erweiterte Abschnitte:**

1. **Grundidee & Paradigmenwechsel:** Snapshot-basierte vs. Event-basierte Persistenz, Accounting-Ledger-Analogie
2. **Event Store:**
   - Technologien im Vergleich: EventStoreDB, Marten (.NET), Axon (Java), PostgreSQL (generisch), DynamoDB, Kafka als Event Store
   - Append-only, Immutability, Ordering Guarantees
3. **Event Design:**
   - Thin vs. Fat Events, Event Granularity
   - Domain Events vs. Integration Events
   - Self-contained, Past Tense, Versioniert
4. **State Reconstruction (Replay):** Apply-Funktion, Fold/Reduce-Pattern
5. **Snapshots:** Strategien (N-Events, zeitbasiert, bei jedem Write), Snapshot als Event
6. **Event-Schema-Evolution:**
   - Upcasting, Lazy Migration, Schema Registry
   - Versioned Serializers, Avro/Protobuf
   - Backward/Forward Compatibility
7. **Fortgeschrittene Patterns:**
   - **Saga Pattern / Process Manager**: Koordination verteilter Workflows
   - **Outbox Pattern**: Reliable Event Publishing aus der DB
   - **Inbox Pattern**: Idempotente Event-Verarbeitung
   - **Idempotenz & Concurrency Control**: Optimistic Concurrency, Idempotency Keys
8. **Event-Sourcing vs. CRUD: Entscheidungsmatrix** (erweitert, mehr Kriterien)
9. **Reale Fallstudien:** Banking, E-Commerce, Logistics
10. **Kombination mit CQRS:** Warum sie zusammenpassen, Cross-Referenz zu `cqrs.md`
11. **Anti-Patterns:** Event als Command, Zu große Events, Events mutieren, CRUD als ES
12. **Referenzen**

**Quellen für diesen Schritt:**

| Quelle                                   | URL                                                                                         | Fokus                                                                            |
| ---------------------------------------- | ------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------- |
| Event Sourcing Explained (BayTech 2025)  | https://www.baytechconsulting.com/blog/event-sourcing-explained-2025                        | Paradigm Shift, Core Mechanics, Event Store Anatomy, Snapshots, Temporal Queries |
| Event Sourcing vs. CRUD (dev.to)         | https://dev.to/alex_aslam/event-sourcing-vs-crud-when-1000-database-writes-dont-matter-5bpj | Entscheidungsmatrix, Ruby/Rails-Perspektive, Practical Examples                  |
| Event-Driven Architecture in Golang      | https://github.com/PacktPublishing/Event-Driven-Architecture-in-Golang                      | ES in Go-Microservices, NATS Messaging                                           |
| Greg Young: CQRS Documents               | https://cqrs.wordpress.com/                                                                 | Event-Sourcing Grundkonzepte                                                     |
| Martin Fowler: Event Sourcing            | https://martinfowler.com/eaaDev/EventSourcing.html                                          | Grundlagen und Konzepte                                                          |
| Oskar Dudycz: Event Sourcing in .NET     | https://event-driven.io/en/                                                                 | Praxis-Blog mit Patterns (Outbox, Snapshots, Schema Evolution)                   |
| EventStoreDB Docs                        | https://developers.eventstore.com/                                                          | Spezialisierter Event Store, Subscriptions                                       |
| Marten Docs                              | https://martendb.io/                                                                        | PostgreSQL als Event Store in .NET                                               |
| Chris Richardson: Microservices Patterns | https://microservices.io/patterns/                                                          | Saga, Outbox, Event Sourcing Patterns                                            |
| Designing Data-Intensive Applications    | Buch-Referenz (Kleppmann 2017)                                                              | Event Sourcing, Immutability, Duality of Streams                                 |

---

### Instruktion 3b: `cqrs.md` (neu) — Eigenständiger CQRS-Guide

**Ziel:** Aus dem CQRS-Teil von `event-sourcing-cqrs.md` ein eigenständiges ~500-700 Zeilen Dokument.

**Neue/erweiterte Abschnitte:**

1. **Grundidee:** CQS (Bertrand Meyer) → CQRS (Greg Young), Command Side vs. Query Side
2. **Command Side:** Commands als Absichtserklärung, Validierung, Idempotenz
3. **Query Side:** Read Models, DTOs, optimierte Datenstrukturen
4. **CQRS-Ausbaustufen:**
   - Stufe 0: Kein CQRS (klassisches CRUD)
   - Stufe 1: Logische Trennung (separate Handler)
   - Stufe 2: Getrennte Modelle (Projektionen)
   - Stufe 3: Getrennte Datenbanken
5. **Read Model Design:**
   - Denormalisierung
   - Materialized Views
   - Search Indexes (Elasticsearch)
   - Multiple Read Models für verschiedene Use-Cases
6. **Projektionsstrategien:**
   - Synchrone Projektion
   - Asynchrone Projektion (Polling, LISTEN/NOTIFY, Message Queue)
   - Hybride Projektion
   - Change Data Capture (Debezium)
   - Transactional Outbox
7. **Eventual Consistency Strategien:**
   - Read-Your-Own-Writes, Causal Consistency, Session Consistency
   - Compensation / Corrective Events
   - UI-Strategien: Optimistic Updates, Stale-While-Revalidate
8. **CQRS ohne Event-Sourcing:** Eigenständiger Wert, Beispiele mit klassischer DB
9. **Kombination mit Event-Sourcing:** Warum sie zusammenpassen, Cross-Referenz zu `event-sourcing.md`
10. **Entscheidungsmatrix:** Wann CQRS einsetzen? Wann reicht CRUD?
11. **Anti-Patterns:** Over-Engineering, Synchrone Projektionen überall, CQRS für CRUD
12. **Referenzen**

**Quellen für diesen Schritt:**

| Quelle                                   | URL                                                                                                       | Fokus                                                  |
| ---------------------------------------- | --------------------------------------------------------------------------------------------------------- | ------------------------------------------------------ |
| AWS CQRS Pattern                         | https://docs.aws.amazon.com/prescriptive-guidance/latest/modernization-data-persistence/cqrs-pattern.html | Cloud-native CQRS, DynamoDB+Aurora, Read Replicas      |
| Greg Young: CQRS Documents               | https://cqrs.wordpress.com/                                                                               | Ursprung von CQRS                                      |
| Martin Fowler: CQRS                      | https://martinfowler.com/bliki/CQRS.html                                                                  | Wann CQRS sinnvoll ist, Warnungen vor Over-Engineering |
| Udi Dahan: Clarified CQRS                | https://udidahan.com/2009/12/09/clarified-cqrs/                                                           | CQRS + DDD, eigenständiger Wert von CQRS               |
| Event-Driven Architecture in Golang      | https://github.com/PacktPublishing/Event-Driven-Architecture-in-Golang                                    | CQRS in Go-Microservices                               |
| Kamil Grzybek: Modular Monolith with DDD | https://github.com/kgrzybek/modular-monolith-with-ddd                                                     | Outbox, Inbox, CQRS in einem Monolithen                |
| Oskar Dudycz: Event-Driven.io            | https://event-driven.io/en/                                                                               | Projektionen, Read Models, CQRS-Praxis                 |
| Chris Richardson: Microservices Patterns | https://microservices.io/patterns/                                                                        | CQRS als Microservice-Pattern                          |
| Designing Data-Intensive Applications    | Buch-Referenz (Kleppmann 2017)                                                                            | Consistency, Derived Data, Stream Processing           |

---

### Instruktion 4: `go-backend.md` — Erweitern zu einem allgemeinen Go-Backend-Architektur-Guide

> _Hinweis: Durch die Aufspaltung von ES/CQRS gibt es nun 9 Schritte statt 8._

**Ziel:** Von 703 Zeilen jotti-Architektur zu ~900-1100 Zeilen allgemeiner Go-Backend-Theorie.

**Neue/erweiterte Abschnitte:**

1. **Architektur-Patterns im Vergleich:**
   - **Layered Architecture** (aktuell beschrieben) — Stärken/Schwächen
   - **Hexagonal Architecture (Ports & Adapters)** — Alistair Cockburn
   - **Clean Architecture** — Robert C. Martin
   - **Onion Architecture** — Jeffrey Palermo
   - Vergleichstabelle: Wann welches Pattern?
2. **Go-HTTP-Ökosystem:**
   - stdlib `net/http` (Go 1.22+ Enhanced Router mit Method Matching und Wildcards)
   - Chi, Gorilla Mux, Echo, Gin, Fiber — Feature-Vergleich
   - Wann reicht stdlib, wann lohnt sich ein Framework?
3. **API-Design-Patterns:**
   - REST, POST-only, RPC-style, GraphQL, gRPC
   - API-Versionierung (URL, Header, Content Negotiation)
   - Pagination, Filtering, Sorting
   - Rate Limiting Patterns (Token Bucket, Sliding Window, Fixed Window)
4. **Go Concurrency Patterns:**
   - Worker Pools, Fan-out/Fan-in, Pipeline
   - Graceful Shutdown mit `context.Context`
   - `errgroup` für parallele Operationen
5. **Error Handling in Go (vertieft):**
   - Sentinel Errors, Custom Error Types, Error Wrapping (Go 1.13+)
   - `errors.Is()`, `errors.As()`, `fmt.Errorf("%w", err)`
   - Domain Errors vs. Infrastructure Errors vs. Application Errors
6. **Configuration & 12-Factor App:**
   - Environment Variables, Config Files, Feature Flags
   - Viper, envconfig, koanf
7. **Resilienz-Patterns:**
   - Circuit Breaker, Retry with Backoff, Timeout, Bulkhead
   - Health Checks, Liveness/Readiness Probes
8. **Observability:**
   - Structured Logging (zerolog, slog, zap)
   - Distributed Tracing (OpenTelemetry)
   - Metrics (Prometheus)
9. **Datenbankzugriff / SQL-Tooling (aus `postgresql-sqlc.md` übernommen):**
   - **SQL-Tooling Landschaft:** sqlc, sqlx, GORM, ent, Jet — Feature-Vergleich und Entscheidungsmatrix
   - **sqlc Deep Dive:** Query-Design-Patterns, Custom Types, Batch Operations, Prepared Statements, Einschränkungen
   - **ORM vs. SQL-first vs. Query Builder:** Wann welcher Ansatz? Trade-offs (Typsicherheit, Flexibilität, Lernkurve)
   - **Migration-Tooling:** golang-migrate, Atlas, goose — Vergleich, Zero-Downtime Migrationen
   - **Repository Pattern in Go:** Interface-Design, Testbarkeit, Mocking mit sqlc
10. **Testing in Go:**
    - **Table-Driven Tests**: Idiomatisches Go-Testing-Pattern
    - **Testcontainers**: Integrationstests mit echten Datenbanken (PostgreSQL, Redis)
    - **httptest**: HTTP-Handler testen ohne Server
    - **Golden Files**: Snapshot-basiertes Testing für komplexe Outputs
    - **Mocking-Strategien**: Interface-basierte Mocks, `go generate`, testify/mock
    - **Test-Pyramide im Go-Kontext**: Unit → Integration → E2E, Build Tags (`//go:build unit`)
    - **Test Fixtures & Factories**: Setup/Teardown, testhelper-Packages
    - **Coverage & Benchmarks**: `go test -cover`, `go test -bench`

**Quellen für diesen Schritt:**

| Quelle                                    | URL                                                                                                       | Fokus                                            |
| ----------------------------------------- | --------------------------------------------------------------------------------------------------------- | ------------------------------------------------ |
| Event-Driven Architecture in Golang       | https://github.com/PacktPublishing/Event-Driven-Architecture-in-Golang                                    | Go-Microservices, DDD-Schichtung, Event Handling |
| Standard Go Project Layout                | https://github.com/golang-standards/project-layout                                                        | Verzeichnisstruktur-Konventionen                 |
| Effective Go                              | https://go.dev/doc/effective_go                                                                           | Offizieller Style Guide                          |
| Go Proverbs (Rob Pike)                    | https://go-proverbs.github.io/                                                                            | Go Design-Philosophie                            |
| Go Blog: Error Handling                   | https://go.dev/blog/error-handling-and-go                                                                 | Offizielle Error-Handling-Patterns               |
| Go Blog: Errors Are Values                | https://go.dev/blog/errors-are-values                                                                     | Fehlerbehandlung als Design-Prinzip              |
| Alistair Cockburn: Hexagonal Architecture | https://alistair.cockburn.us/hexagonal-architecture/                                                      | Ports & Adapters Originalquelle                  |
| Robert C. Martin: Clean Architecture      | https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html                              | Clean Architecture Blog Post                     |
| GopherCon Talks                           | https://www.youtube.com/c/GopherAcademy                                                                   | Go Architecture Talks                            |
| Alex Edwards: Let's Go (Further)          | https://lets-go-further.alexedwards.net/                                                                  | Go Web Application Patterns, Middleware, APIs    |
| Go Wiki: Go Code Review Comments          | https://go.dev/wiki/CodeReviewComments                                                                    | Idiomatisches Go                                 |
| 12-Factor App                             | https://12factor.net/                                                                                     | Configuration, Logging, Concurrency              |
| Microsoft: Resiliency Patterns            | https://learn.microsoft.com/en-us/azure/architecture/patterns/category/resiliency                         | Circuit Breaker, Retry, Bulkhead                 |
| sqlc Dokumentation                        | https://docs.sqlc.dev/                                                                                    | Offizielle sqlc-Referenz, Query Patterns         |
| pgx v5 Dokumentation                      | https://pkg.go.dev/github.com/jackc/pgx/v5                                                                | Go PostgreSQL Driver, Connection Pooling         |
| ent (Go ORM)                              | https://entgo.io/                                                                                         | Graph-basiertes Go ORM, Schema-as-Code           |
| Atlas (Schema Migration)                  | https://atlasgo.io/                                                                                       | Deklarative Schema-Migrationen für Go            |
| DB Performance 101 (dev.to)               | https://dev.to/ari-ghosh/db-performance-101-a-practical-deep-dive-into-backend-database-optimization-4cag | N+1, Connection Pooling, Query Optimization      |
| Go Testing (Dave Cheney)                  | https://dave.cheney.net/2019/05/07/prefer-table-driven-tests                                              | Table-Driven Tests, Go-Testing-Philosophie       |
| testcontainers-go                         | https://golang.testcontainers.org/                                                                        | Integrationstests mit echten Containern          |
| Mitchell Hashimoto: Testing in Go         | https://www.youtube.com/watch?v=8hQG7QlcLBk                                                               | Advanced Go Testing Patterns (GopherCon)         |

---

### Instruktion 5: `postgresql.md` (ehemals `postgresql-sqlc.md`) — Eigenständiger PostgreSQL-Guide

> _Hinweis: sqlc und SQL-Tooling-Vergleich wurden in Instruktion 4 (`go-backend.md`) integriert. Dieses Dokument fokussiert ausschließlich auf PostgreSQL als Datenbank._

**Ziel:** Von 581 Zeilen (nach Entfernung der sqlc-Abschnitte) zu ~600-800 Zeilen allgemeiner PostgreSQL-Theorie.

**Neue/erweiterte Abschnitte:**

1. **PostgreSQL Architecture Basics:**
   - MVCC (Multi-Version Concurrency Control) erklärt
   - WAL (Write-Ahead Logging) und wie es Recovery ermöglicht
   - Autovacuum und Bloat-Management
   - Shared Buffers, Work Memory, Effective Cache Size
2. **Indexing Deep Dive:**
   - B-Tree (Default), GIN (JSONB, Arrays), GiST (Geometrie, Volltext), BRIN (zeitliche Daten)
   - Partial Indexes (`WHERE active = true`)
   - Covering Indexes (`INCLUDE`)
   - Index-Only Scans
   - Composite vs. Single-Column Indexes — Entscheidungshilfe
3. **Query Optimization:**
   - `EXPLAIN ANALYZE` lesen und interpretieren
   - Common Table Expressions (CTEs) — materialized vs. non-materialized
   - Window Functions
   - Query Planner Statistiken (`pg_stat_statements`)
4. **Advanced Features:**
   - **LISTEN/NOTIFY**: Asynchrone Benachrichtigungen (ideal für CQRS-Projektionen)
   - **Partitioning**: Range, List, Hash — wann welches?
   - **Materialized Views**: Pre-computed Read Models
   - **Row-Level Security (RLS)**: Multi-Tenant-Sicherheit
   - **Generated Columns**: Computed Columns in der DB
   - **Foreign Data Wrappers**: Zugriff auf externe Datenquellen
5. **Connection Management:**
   - pgxpool (Go-intern), PgBouncer (externer Pool), Odyssey
   - Connection Limits, Pool Sizing, Idle Timeout
6. **PostgreSQL als Event Store:**
   - Vergleich mit EventStoreDB, Kafka
   - Trigger-basierte Immutability
   - LISTEN/NOTIFY für Projektionen
   - Partitioning für große Event-Tabellen
7. **Schema-Migration-Strategien:**
   - golang-migrate, Atlas, Flyway, Liquibase
   - Zero-Downtime Migration (expand/contract)
   - Backward-compatible Schema Changes
8. **PostgreSQL vs. Alternativen:**
   - PostgreSQL vs. MySQL/MariaDB (erweitert)
   - PostgreSQL vs. NoSQL (MongoDB, DynamoDB) — Wann welches?
   - NewSQL (CockroachDB, YugabyteDB) als Ausblick

**Quellen für diesen Schritt:**

| Quelle                                    | URL                                                                                                       | Fokus                                                 |
| ----------------------------------------- | --------------------------------------------------------------------------------------------------------- | ----------------------------------------------------- |
| PostgreSQL vs MySQL (Bytebase)            | https://www.bytebase.com/blog/postgres-vs-mysql/                                                          | Feature-Vergleich, JSONB, Indexierung, Replication    |
| DB Performance 101 (dev.to)               | https://dev.to/ari-ghosh/db-performance-101-a-practical-deep-dive-into-backend-database-optimization-4cag | Connection Pooling, N+1, Indexing, Query Optimization |
| PostgreSQL Docs                           | https://www.postgresql.org/docs/current/                                                                  | Offizielle Referenz für alle Features                 |
| PostgreSQL Wiki: Performance Optimization | https://wiki.postgresql.org/wiki/Performance_Optimization                                                 | Tuning-Checkliste                                     |
| Use The Index, Luke                       | https://use-the-index-luke.com/                                                                           | Indexing-Bibel (DB-agnostisch mit SQL-Fokus)          |
| Crunchy Data Blog                         | https://www.crunchydata.com/blog                                                                          | PostgreSQL-Praxistipps, Performance                   |
| Postgres.fm Podcast                       | https://postgres.fm/                                                                                      | PostgreSQL Deep Dives (Audio/Transkripte)             |
| Cybertec PostgreSQL Blog                  | https://www.cybertec-postgresql.com/en/blog/                                                              | MVCC, Vacuum, Performance, Partitioning               |
| Brandur: Postgres as Event Store          | https://brandur.org/postgres-atomicity                                                                    | PostgreSQL Internals, Event Patterns                  |
| Atlas (Schema Migration)                  | https://atlasgo.io/                                                                                       | Moderne deklarative Schema-Migrationen                |

---

### Instruktion 6: `react-frontend.md` — Erweitern zu einem vollständigen React-Frontend-Architektur-Guide

**Ziel:** Von 732 Zeilen zu ~1000-1300 Zeilen allgemeiner React-Frontend-Architektur.

**Neue/erweiterte Abschnitte:**

1. **React Design Patterns Katalog (erweitern auf 15+):**
   - **Component Composition** (vorhanden, erweitern)
   - **Custom Hook Pattern** (vorhanden, erweitern)
   - **Container/Presentational** (vorhanden)
   - **Compound Components** (vorhanden, erweitern)
   - **Control Props Pattern**: Controlled vs. Uncontrolled Components
   - **Provider Pattern**: Context als Dependency Injection
   - **Headless Components**: Logik ohne UI (Headless UI, Radix Primitives)
   - **Render Props** (vorhanden als Erwähnung, vollständig erklären)
   - **Props Getters Pattern**: Flexible API für Komponent-Konsumenten
   - **Error Boundary Pattern**: Fehlerbehandlung in React-Bäumen
   - **Portal Pattern**: Rendering außerhalb des DOM-Baums
   - **HOC (Higher Order Components)**: Legacy Pattern, warum Hooks besser sind
   - **MVVM in React**: Model-View-ViewModel mit Hooks als ViewModel
   - **Dependency Injection in React**: via Context, Props, oder Module
   - **SOLID in React**: Wie die SOLID-Prinzipien in React aussehen
2. **Frontend Architecture Patterns:**
   - Monolithic (Single-Page) Architecture
   - Modular Architecture (Feature-Sliced Design)
   - Component-Based Architecture
   - Micro-Frontend Architecture (Module Federation, Single-SPA)
   - Flux/Redux Architecture (und warum man es vermeiden kann)
   - Vergleichstabelle mit Empfehlungen
3. **State-Management Landscape:**
   - Server State: TanStack Query (React Query), SWR, RTK Query
   - Client State: useState, useReducer, Zustand, Jotai, Recoil, Redux Toolkit
   - Form State: React Hook Form, Formik, native
   - URL State: React Router, nuqs
   - Entscheidungsmatrix: Welches Tool für welchen State?
4. **Testing-Strategien:**
   - Unit Tests: Vitest, Jest
   - Component Tests: React Testing Library
   - Integration Tests: MSW (Mock Service Worker)
   - E2E Tests: Playwright, Cypress
   - Test-Pyramide für React
5. **Performance-Patterns:**
   - `React.memo`, `useMemo`, `useCallback` — wann wirklich nötig?
   - `React.lazy` + `Suspense` für Code Splitting
   - Virtualization (TanStack Virtual, react-window)
   - Bundle-Analyse und Tree Shaking
6. **Accessibility (a11y):**
   - ARIA-Rollen und Labels
   - Keyboard Navigation
   - Focus Management
   - Screen Reader Testing
7. **TypeScript-Patterns für React:**
   - Discriminated Unions für Props
   - Generic Components (z.B. `DataList<T>`)
   - Template Literal Types für Event Handler
   - Branded Types für IDs (User-ID ≠ Product-ID)
8. **Atomic Design (erweitern):**
   - Brad Frost's vollständiges Modell (Atoms, Molecules, Organisms, Templates, Pages)
   - Storybook Integration
   - Design System Aufbau

**Quellen für diesen Schritt:**

| Quelle                                      | URL                                                                                                                             | Fokus                                                                                                                                                                                                                              |
| ------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 21 React Design Patterns (Persson Dennis)   | https://www.perssondennis.com/articles/21-fantastic-react-design-patterns-and-when-to-use-them                                  | 21 Patterns mit Code-Beispielen: Composition, Hooks, Control Props, Provider, Container/Presentational, Compound, Headless, Atomic, Error Boundary, Portal, Render Props, Props Getters, HOC, DRY, SOLID, DI, SoC, MVVM, SDP, KISS |
| Frontend Architecture Patterns (LogRocket)  | https://blog.logrocket.com/guide-modern-frontend-architecture-patterns/                                                         | Monolithic, Modular, Component-Based, Microfrontend, Flux/Redux, Hybrid — Vergleichstabelle                                                                                                                                        |
| Atomic Design (Plain English)               | https://javascript.plainenglish.io/mastering-atomic-design-a-step-by-step-guide-to-building-scalable-ui-components-60b0d2a94cc3 | Step-by-Step Atomic Design in React                                                                                                                                                                                                |
| Brad Frost: Atomic Design                   | https://bradfrost.com/blog/post/atomic-web-design/                                                                              | Original Atomic Design Blogpost                                                                                                                                                                                                    |
| Kent C. Dodds: Application State Management | https://kentcdodds.com/blog/application-state-management-with-react                                                             | State-Management ohne Redux                                                                                                                                                                                                        |
| TanStack Query Docs                         | https://tanstack.com/query/latest                                                                                               | Server State Management                                                                                                                                                                                                            |
| React Docs                                  | https://react.dev/                                                                                                              | Offizielle React-Dokumentation (Hooks, Patterns)                                                                                                                                                                                   |
| shadcn/ui                                   | https://ui.shadcn.com/                                                                                                          | Radix + Tailwind Komponenten                                                                                                                                                                                                       |
| Testing Library Docs                        | https://testing-library.com/docs/react-testing-library/intro/                                                                   | React Component Testing                                                                                                                                                                                                            |
| Josh Comeau: Joy of React                   | https://www.joyofreact.com/                                                                                                     | React Mental Models, Performance                                                                                                                                                                                                   |
| Dan Abramov: Writing Resilient Components   | https://overreacted.io/writing-resilient-components/                                                                            | React Best Practices                                                                                                                                                                                                               |
| Feature-Sliced Design                       | https://feature-sliced.design/                                                                                                  | Architektur-Methodology für Frontend                                                                                                                                                                                               |

---

### Instruktion 7: `README.md` — Aktualisieren

**Ziel:** README.md an die neuen, allgemeineren Dokumente anpassen. Neue Ziel-Dateistruktur:

```
docs/theory/
├── README.md                # Übersicht und Leseempfehlungen
├── ddd.md                   # Domain-Driven Design
├── event-sourcing.md        # Event-Sourcing (eigenständig)
├── cqrs.md                  # CQRS (eigenständig)
├── go-backend.md            # Go Backend Architektur + SQL-Tooling (sqlc) + Testing
├── postgresql.md            # PostgreSQL (eigenständig, ohne sqlc)
├── react-frontend.md        # React Frontend Architektur
├── security.md              # Security & Auth (neu)
├── devops.md                # DevOps, Deployment & Infrastruktur (neu)
└── pos.md                   # POS-Systeme & Gastronomie-Domäne (neu)
```

**Schritte:**

1. Neue Dokumente aufnehmen: `cqrs.md`, `event-sourcing.md`, `security.md`, `devops.md`, `pos.md`
2. Alten Eintrag `event-sourcing-cqrs.md` entfernen
3. Beschreibungen aktualisieren (nicht mehr „jotti" im Fokus)
4. Neue Themenübersicht pro Dokument (jetzt 9 Theorie-Dokumente)
5. Leseempfehlungen erweitern (Lesereihenfolge für neue Dokumente)
6. Abgrenzungstabelle aktualisieren

---

### Instruktion 8: Review & Konsistenz (Zwischenstand nach Kern-Dokumenten)

**Ziel:** Zwischenreview über die 7 Kern-Dateien (README + 6 bestehende Theorie-Dokumente) bevor die neuen Dokumente geschrieben werden.

**Schritte:**

1. Konsistente Struktur prüfen (gleiche Abschnitts-Tiefe, gleiche Tabellen-Formate)
2. Cross-Referenzen zwischen Dokumenten aktualisieren (insb. ES↔CQRS)
3. Alle externen Links prüfen (erreichbar? aktuell?)
4. Duplikate eliminieren (gleiche Erklärung in zwei Dokumenten?)
5. Sprachkonsistenz: Theorie auf Deutsch, Code-Beispiele in der Originalsprache

---

### Instruktion 9: `security.md` (neu) — Security & Authentication Guide

**Ziel:** Neues Dokument, ~600-800 Zeilen allgemeine Web-Application-Security-Theorie.

**Abschnitte:**

1. **Einleitung: Warum Security-Architektur?**
   - Security als Architektur-Querschnittsthema, nicht als Afterthought
   - Threat Modeling Basics (STRIDE)
2. **Authentication-Patterns:**
   - **Session-basiert**: Server-Side Sessions, Cookie-Management, Sticky Sessions
   - **Token-basiert (JWT)**: HS256 vs. RS256, Access + Refresh Tokens, Token Rotation
   - **OAuth2 / OIDC**: Authorization Code Flow, PKCE, ID Tokens
   - **API Keys**: Wann sinnvoll? Scoping, Rate Limiting
   - **Passkeys / WebAuthn**: Passwordless als Zukunft
   - Entscheidungsmatrix: Wann welches Pattern?
3. **Authorization-Modelle:**
   - **RBAC** (Role-Based Access Control): Rollen, Berechtigungen, Rollenhierarchien
   - **ABAC** (Attribute-Based Access Control): Policies, Attribute, Kontext
   - **ReBAC** (Relationship-Based): Google Zanzibar, SpiceDB, OpenFGA
   - **Policy Engines**: OPA (Open Policy Agent), Cedar (AWS)
   - Vergleichstabelle mit Empfehlungen nach Anwendungsfall
4. **Passwort-Sicherheit:**
   - Hashing-Algorithmen: Argon2id, bcrypt, scrypt — Vergleich, Parameter-Empfehlungen
   - Salting, Peppering, Key Stretching
   - Password Policies: NIST-Empfehlungen (2024), Zxcvbn
5. **OWASP Top 10 (2021+):**
   - Broken Access Control, Cryptographic Failures, Injection, Insecure Design
   - Security Misconfiguration, Vulnerable Components, Auth Failures
   - Software Integrity Failures, Logging Failures, SSRF
   - Pro Kategorie: Erklärung, Beispiel, Mitigation
6. **API-Security:**
   - Input Validation (Whitelist > Blacklist)
   - CORS (Cross-Origin Resource Sharing) — Konfiguration und Fallstricke
   - CSRF-Schutz: SameSite Cookies, Double Submit, CSRF Tokens
   - Rate Limiting & Throttling
   - Content Security Policy (CSP), Helmet/Security Headers
7. **Frontend-Security:**
   - XSS-Prävention: React-inhärenter Schutz (`dangerouslySetInnerHTML`)
   - Sicheres Token-Handling: httpOnly Cookies vs. localStorage vs. Memory
   - Subresource Integrity (SRI)
   - Secure Communication (HTTPS Only, HSTS)
8. **Secrets Management:**
   - Environment Variables (12-Factor), .env-Dateien
   - HashiCorp Vault, SOPS, Sealed Secrets
   - Key Rotation, Secret Scanning (CI/CD)
9. **TLS & Certificate Management:**
   - TLS 1.3, Certificate Chains, Let's Encrypt / ACME
   - Certificate Renewal Automation, HSTS Preloading
10. **Security Testing:**
    - SAST (Static Analysis), DAST (Dynamic Analysis)
    - Dependency Scanning (Dependabot, Snyk, Trivy)
    - Penetration Testing Basics
11. **Entscheidungsmatrix:** Auth-Pattern vs. Anwendungsfall (SPA, API, Mobile, M2M)
12. **Referenzen**

**Quellen für diesen Schritt:**

| Quelle                           | URL                                                                                   | Fokus                                                    |
| -------------------------------- | ------------------------------------------------------------------------------------- | -------------------------------------------------------- |
| OWASP Top 10                     | https://owasp.org/www-project-top-ten/                                                | Die 10 kritischsten Web-Security-Risiken                 |
| OWASP Cheat Sheet Series         | https://cheatsheetseries.owasp.org/                                                   | Praxis-Checklisten für jeden Security-Aspekt             |
| Auth0 Blog: JWT Handbook         | https://auth0.com/resources/ebooks/jwt-handbook                                       | JWT Deep Dive, Signierung, Claims                        |
| NIST Digital Identity Guidelines | https://pages.nist.gov/800-63-3/                                                      | Password Policies, MFA, Identity Proofing                |
| Google Zanzibar Paper            | https://research.google/pubs/zanzibar-googles-consistent-global-authorization-system/ | ReBAC-Grundlage, Relationship Tuples                     |
| OWASP API Security Top 10        | https://owasp.org/API-Security/                                                       | API-spezifische Risiken                                  |
| MDN: Web Security                | https://developer.mozilla.org/en-US/docs/Web/Security                                 | CORS, CSP, CSRF — Browser-Perspektive                    |
| Let's Encrypt Docs               | https://letsencrypt.org/docs/                                                         | ACME-Protokoll, Certificate Management                   |
| Hashing in Context (OWASP)       | https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html      | Argon2id, bcrypt — Parameter-Empfehlungen                |
| Open Policy Agent                | https://www.openpolicyagent.org/                                                      | Policy-based Authorization                               |
| OpenFGA                          | https://openfga.dev/                                                                  | Open-Source Zanzibar-Implementation                      |
| Filippo Valsorda: Age/TLS        | https://words.filippo.io/                                                             | Go Crypto, TLS Best Practices                            |
| Veracode: DevSecOps & AppSec     | https://www.veracode.com/blog/what-is-devops-and-devsecops/                           | Sicherheitstests in CI/CD-Pipelines, Shift-Left Security |
| Duende Software: JWT Security    | https://duendesoftware.com/learn/best-practices-using-jwts-with-web-and-mobile-apps   | JWT Best Practices, Token-Lebensdauern, XSS/CSRF-Schutz  |
| Tecktol: Zod Schema Validation   | https://tecktol.com/zod-schema-validation-the-complete-guide/                         | Laufzeit-Validierung, Schutz vor Injection-Angriffen     |

---

### Instruktion 10: `devops.md` (neu) — DevOps, Deployment & Infrastruktur Guide

**Ziel:** Neues Dokument, ~600-800 Zeilen allgemeine DevOps- und Deployment-Theorie für Web-Anwendungen.

**Abschnitte:**

1. **Einleitung: DevOps-Kultur & Prinzipien**
   - DevOps als Kultur (nicht nur Tooling), CALMS-Framework
   - Continuous Delivery vs. Continuous Deployment
   - Infrastructure as Code (IaC) Philosophie
2. **Containerisierung:**
   - **Docker Basics**: Images, Container, Layers, Multi-Stage Builds
   - **Image-Optimierung**: Alpine vs. Distroless vs. Scratch, Layer Caching
   - **Docker Compose**: Service-Orchestrierung, Networking, Volumes, Profiles
   - **Container Security**: Non-root User, Read-only Filesystems, Image Scanning (Trivy)
3. **Orchestrierung:**
   - **Docker Compose** (für Self-Hosting, Single-Server)
   - **Kubernetes Basics**: Pods, Deployments, Services, Ingress
   - **K3s / K0s**: Leichtgewichtige K8s-Distributionen für Self-Hosting
   - Entscheidungsmatrix: Docker Compose vs. Kubernetes vs. Managed Cloud
4. **Reverse Proxy & Load Balancing:**
   - **nginx**: Konfiguration, Location Blocks, Proxy Pass, Caching
   - **Traefik**: Automatic Service Discovery, Let's Encrypt Integration
   - **Caddy**: Automatic HTTPS, Caddyfile-Einfachheit
   - TLS Termination, HTTP/2, WebSocket Proxying
   - Vergleichstabelle
5. **CI/CD Pipeline Design:**
   - **Pipeline-Stufen**: Lint → Test → Build → Publish → Deploy
   - **GitHub Actions**: Workflow-Syntax, Matrix Builds, Caching, Secrets
   - **GitLab CI / Drone / Woodpecker**: Alternativen
   - **Artefakt-Management**: Container Registry, Release Tags
   - **Deployment-Trigger**: Git Push, Tag, Manual Approval
6. **Deployment-Strategien:**
   - **Recreate**: Simpel, Downtime akzeptabel
   - **Rolling Update**: Zero-Downtime, schrittweise
   - **Blue/Green**: Sofortiger Switch, einfacher Rollback
   - **Canary**: Prozentuale Verteilung, Feature Flags
   - Vergleichstabelle nach Komplexität, Rollback-Fähigkeit, Ressourcenbedarf
7. **Zero-Downtime Deployment:**
   - Health Checks (Liveness, Readiness, Startup Probes)
   - Graceful Shutdown (Drain Connections, SIGTERM-Handling)
   - Database Migrations und Backward Compatibility
   - Connection Draining am Reverse Proxy
8. **Monitoring & Observability:**
   - **Metriken**: Prometheus, Grafana — RED Method, USE Method
   - **Logging**: Structured Logging, Log Aggregation (Loki, ELK)
   - **Alerting**: PagerDuty, Alertmanager, Uptime Monitoring
   - **Dashboards**: Wichtige Metriken für Web-Anwendungen
9. **Backup & Disaster Recovery:**
   - Datenbank-Backups (pg_dump, WAL Archiving, pg_basebackup)
   - Volume-Backups, Restic, Borg
   - RTO (Recovery Time Objective) vs. RPO (Recovery Point Objective)
   - Disaster Recovery Plan
10. **Self-Hosting Patterns:**
    - VM-Provisionierung (Hetzner, DigitalOcean, Netcup)
    - DNS-Setup, Firewall (ufw, iptables)
    - Automatische Zertifikate (Certbot, Traefik ACME)
    - Unattended Upgrades, Security Patching
11. **Referenzen**

**Quellen für diesen Schritt:**

| Quelle                                | URL                                                                                                     | Fokus                                                 |
| ------------------------------------- | ------------------------------------------------------------------------------------------------------- | ----------------------------------------------------- |
| Docker Docs                           | https://docs.docker.com/                                                                                | Offizielle Docker-Referenz                            |
| Docker Best Practices                 | https://docs.docker.com/build/building/best-practices/                                                  | Multi-Stage, Layer Caching, Security                  |
| 12-Factor App                         | https://12factor.net/                                                                                   | Config, Logging, Disposability, Dev/Prod Parity       |
| nginx Docs                            | https://nginx.org/en/docs/                                                                              | Reverse Proxy, Load Balancing, TLS                    |
| Traefik Docs                          | https://doc.traefik.io/traefik/                                                                         | Automatic HTTPS, Docker Integration                   |
| Caddy Docs                            | https://caddyserver.com/docs/                                                                           | Automatic TLS, Caddyfile                              |
| GitHub Actions Docs                   | https://docs.github.com/en/actions                                                                      | CI/CD Pipeline-Design, Workflow-Syntax                |
| Kubernetes Docs                       | https://kubernetes.io/docs/                                                                             | K8s Concepts, Deployments, Services                   |
| Google SRE Book                       | https://sre.google/sre-book/table-of-contents/                                                          | Monitoring, Alerting, Incident Response, SLOs         |
| Prometheus Docs                       | https://prometheus.io/docs/                                                                             | Metriken, Alerting, Grafana-Integration               |
| Let's Encrypt Docs                    | https://letsencrypt.org/docs/                                                                           | ACME, Certificate Lifecycle                           |
| Grafana Loki                          | https://grafana.com/oss/loki/                                                                           | Log Aggregation                                       |
| Martin Fowler: Continuous Delivery    | https://martinfowler.com/bliki/ContinuousDelivery.html                                                  | CD-Prinzipien                                         |
| Charity Majors: Observability         | https://charity.wtf/                                                                                    | Observability vs. Monitoring                          |
| Docker on VPS (hosting.international) | https://hosting.international/blog/the-ultimate-guide-to-containerization-how-to-use-docker-on-vps/     | Containerisierung auf VPS, Docker-Setup, Praxis-Guide |
| CI/CD on Linux Server (YouStable)     | https://www.youstable.com/blog/create-ci-cd-on-linux-server                                             | CI/CD-Pipeline mit Docker und Git auf Linux-Servern   |
| Professional CI/CD Pipeline (lilstex) | https://medium.com/@lilstex4good/building-a-professional-ci-cd-pipeline-my-vps-setup-guide-daeca53223cc | VPS-Setup für CI/CD-Pipelines, Deployment-Praxis      |

---

### Instruktion 11: `pos.md` (neu) — POS-Systeme & Gastronomie-Domäne

**Ziel:** `docs/pos.md` als Grundlage nehmen und zu einem ~500-700 Zeilen allgemeinen POS-Theory-Guide erweitern. jotti-spezifische Inhalte in einen Appendix verschieben.

**Ausgangsmaterial:** `docs/pos.md` (bestehend) — enthält POS-Definition, Gastro-Anforderungen, jotti-Positionierung, Produktvergleich.

**Abschnitte:**

1. **Was ist ein POS-System? (erweitern)**
   - Historische Entwicklung: Registrierkasse → ECR → POS → Cloud-POS → mPOS
   - Kernkomponenten: Transaction Engine, Product Catalog, User Management, Reporting
   - POS-Kategorien: Retail, Gastro, Mobile, Self-Service (Kiosk)
2. **POS-Architektur-Patterns:**
   - **On-Premise POS**: Lokaler Server, proprietäre Hardware, Offline-Fähigkeit
   - **Cloud-POS (SaaS)**: Zentrales Backend, Thin Clients, Always-Online
   - **Hybrid POS**: Cloud + lokaler Cache, Offline-First mit Sync
   - **Mobile POS (mPOS)**: Smartphone/Tablet-basiert, PWA vs. Native App
   - Vergleichstabelle: Kosten, Offline-Fähigkeit, Setup-Aufwand, Skalierbarkeit
3. **Gastronomie-POS im Detail:**
   - **Order Lifecycle**: Bestellung → Küche/Ausgabe → Lieferung → Zahlung → Abschluss
   - **Tischbasierter Workflow** vs. Counter-Service vs. Quick-Service
   - **Kitchen Display Systems (KDS)**: Bon-Druck vs. Bildschirm, Routing nach Station
   - **Split Bills & Tab Management**: Teilzahlungen, Tisch wechseln, Zusammenlegen
   - **Mehrere Servicekräfte**: Gleichzeitiger Zugriff, Konfliktvermeidung
4. **Datenmodelle für POS:**
   - **Tischbasiert** (offener Saldo): Gastro-Modell, mehrere Bestellrunden
   - **Transaktionsbasiert** (sofortige Zahlung): Retail-Modell, Scan & Pay
   - **Event-Sourcing im POS-Kontext**: Kassenjournal, Audit Trail, Manipulationssicherheit
   - **Produkt-Modellierung**: Varianten, Modifier, Combos, Dynamic Pricing
5. **Fiskalgesetzgebung & Compliance:**
   - **Deutschland**: TSE (Technische Sicherheitseinrichtung), KassenSichV, GoBD
   - **Österreich**: RKSV, Registrierkassenpflicht
   - **Allgemein**: Warum Fiskalgesetze existieren, digitale Signaturen, Audit-Anforderungen
   - **Non-Profit-Ausnahmen**: Wann gelten Vereinfachungen?
6. **Payment Integration:**
   - **Architektur-Überblick**: Payment Gateway, Acquirer, Processor, Card Network
   - **Kartenleser**: SumUp, Square, Zettle — API-Integration
   - **NFC / Contactless**: Apple Pay, Google Pay
   - **Nicht-Implementierung als bewusste Entscheidung**: Wann Bar-only genügt
7. **Non-Profit vs. Commercial POS:**
   - Feature-Abgrenzung: Was braucht ein Vereinsfest wirklich?
   - Total Cost of Ownership: SaaS-Abo vs. Self-Hosted
   - Setup-Aufwand: 2 Stunden vs. 2 Wochen
   - Hardware-Anforderungen: BYOD (Bring Your Own Device) vs. proprietäre Terminals
8. **POS-Marktlandschaft:**
   - Kommerzielle Systeme: Lightspeed, Orderbird, Toast, Square, Clover
   - Open-Source-Alternativen: UniCenta, Floreant, NORD POS, Apache OFBiz
   - Vergleichstabelle nach Zielgruppe, Preismodell, Offline-Fähigkeit
9. **Appendix: Anwendungsbeispiel (jotti)** — jotti-spezifische Positionierung aus `docs/pos.md`
10. **Referenzen**

**Quellen für diesen Schritt:**

| Quelle                           | URL                                         | Fokus                                     |
| -------------------------------- | ------------------------------------------- | ----------------------------------------- |
| `docs/pos.md` (bestehend)        | Lokal                                       | jotti-Positionierung als Ausgangsmaterial |
| Wikipedia: Point of Sale         | https://en.wikipedia.org/wiki/Point_of_sale | POS-Geschichte, Terminologie              |
| NRF (National Retail Federation) | https://nrf.com/                            | Retail POS Trends, Technologie-Reports    |
| Lightspeed Blog                  | https://www.lightspeedhq.com/blog/          | Gastro-POS Best Practices, Feature-Trends |
| Square Developer Docs            | https://developer.squareup.com/docs         | Payment Integration, POS-API-Design       |
| Orderbird Blog                   | https://www.orderbird.com/blog/             | Gastro-POS in Deutschland, Kassengesetz   |
| BMF: KassenSichV                 | https://www.bundesfinanzministerium.de/     | Fiskalgesetzgebung Deutschland            |
| UniCenta Docs                    | https://unicenta.com/                       | Open-Source POS Architektur               |
| Toast Developer Platform         | https://doc.toasttab.com/                   | Restaurant POS API, KDS Integration       |
| Hospitality Technology Magazine  | https://hospitalitytech.com/                | Gastro-Tech Trends, POS Market Reports    |

---

### Instruktion 12: Abschluss-Review & Konsistenz

**Ziel:** Abschluss-Review über alle 10 Dateien (README + 9 Theorie-Dokumente).

**Schritte:**

1. Konsistente Struktur prüfen (gleiche Abschnitts-Tiefe, gleiche Tabellen-Formate)
2. Cross-Referenzen zwischen Dokumenten aktualisieren (insb. ES↔CQRS, Security↔Go-Backend, DevOps↔PostgreSQL, POS↔ES)
3. Alle externen Links prüfen (erreichbar? aktuell?)
4. Duplikate eliminieren (gleiche Erklärung in zwei Dokumenten?)
5. Sprachkonsistenz: Theorie auf Deutsch, Code-Beispiele in der Originalsprache
6. Querschnittsthemen konsistent: Security-Verweise in go-backend.md und react-frontend.md, DevOps-Verweise in postgresql.md, POS-Verweise in event-sourcing.md

---

## Zusätzliche Quellen und Referenzen (Gesamtpool)

### Bücher

| Buch                                                               | Autor(en)                 | Jahr | Fokus                                           |
| ------------------------------------------------------------------ | ------------------------- | ---- | ----------------------------------------------- |
| Domain-Driven Design: Tackling Complexity in the Heart of Software | Eric Evans                | 2003 | DDD Grundlagenwerk                              |
| Implementing Domain-Driven Design                                  | Vaughn Vernon             | 2013 | Praxisnahe DDD-Umsetzung                        |
| Domain-Driven Design Distilled                                     | Vaughn Vernon             | 2016 | DDD-Kurzfassung                                 |
| Patterns, Principles, and Practices of DDD                         | Scott Millett & Nick Tune | 2015 | DDD-Musterkatalog                               |
| Event-Driven Architecture in Golang                                | Michael Stack             | 2022 | ES + CQRS + DDD in Go                           |
| Designing Data-Intensive Applications                              | Martin Kleppmann          | 2017 | Event Sourcing, CQRS, Consistency, Partitioning |
| Building Microservices                                             | Sam Newman                | 2021 | Microservice Patterns, CQRS                     |
| Clean Architecture                                                 | Robert C. Martin          | 2017 | Architektur-Prinzipien                          |
| Microservices Patterns                                             | Chris Richardson          | 2018 | Saga, Outbox, CQRS, Event Sourcing              |
| Let's Go Further                                                   | Alex Edwards              | 2022 | Go Web APIs, Middleware, Auth                   |
| Learning React (O'Reilly)                                          | Alex Banks & Eve Porcello | 2020 | React Hooks, Patterns                           |
| The Phoenix Project                                                | Gene Kim et al.           | 2013 | DevOps-Kultur, Value Stream, Three Ways         |
| Web Application Security                                           | Andrew Hoffman            | 2020 | OWASP, Auth, XSS, CSRF, API Security            |

### Online-Artikel & Blogs

| Quelle                              | URL                                                                                                                             | Thema                                                     |
| ----------------------------------- | ------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------- |
| Spartner DDD Guide                  | https://spartner.software/kennisbank/domain-driven-design-ddd                                                                   | DDD Foundational: Strategisches & taktisches Design, FAQs |
| Event Sourcing Explained 2025       | https://www.baytechconsulting.com/blog/event-sourcing-explained-2025                                                            | ES Paradigm Shift, Event Store Anatomy, Snapshots         |
| ES vs. CRUD (dev.to)                | https://dev.to/alex_aslam/event-sourcing-vs-crud-when-1000-database-writes-dont-matter-5bpj                                     | Entscheidungsmatrix ES vs. CRUD                           |
| AWS CQRS Pattern                    | https://docs.aws.amazon.com/prescriptive-guidance/latest/modernization-data-persistence/cqrs-pattern.html                       | Cloud-CQRS, DynamoDB + Aurora                             |
| DB Performance 101                  | https://dev.to/ari-ghosh/db-performance-101-a-practical-deep-dive-into-backend-database-optimization-4cag                       | Connection Pooling, N+1, Query Optimization               |
| PostgreSQL vs MySQL                 | https://www.bytebase.com/blog/postgres-vs-mysql/                                                                                | Feature-Vergleich, JSONB, Indizierung                     |
| 21 React Design Patterns            | https://www.perssondennis.com/articles/21-fantastic-react-design-patterns-and-when-to-use-them                                  | 21 Patterns mit React-Code                                |
| Frontend Architecture Patterns      | https://blog.logrocket.com/guide-modern-frontend-architecture-patterns/                                                         | Monolithic, Modular, Micro-Frontends, Flux                |
| Atomic Design (Plain English)       | https://javascript.plainenglish.io/mastering-atomic-design-a-step-by-step-guide-to-building-scalable-ui-components-60b0d2a94cc3 | Atomic Design Step-by-Step                                |
| Martin Fowler: Event Sourcing       | https://martinfowler.com/eaaDev/EventSourcing.html                                                                              | ES-Grundlagen                                             |
| Martin Fowler: CQRS                 | https://martinfowler.com/bliki/CQRS.html                                                                                        | CQRS-Grundlagen                                           |
| Martin Fowler: DDD                  | https://martinfowler.com/tags/domain%20driven%20design.html                                                                     | DDD Artikelsammlung                                       |
| Greg Young: CQRS Documents          | https://cqrs.wordpress.com/                                                                                                     | CQRS + ES Originaldokumente                               |
| Udi Dahan: Clarified CQRS           | https://udidahan.com/2009/12/09/clarified-cqrs/                                                                                 | CQRS + DDD                                                |
| Oskar Dudycz: Event-Driven.io       | https://event-driven.io/en/                                                                                                     | ES/CQRS Praxis-Blog                                       |
| Microservices.io Patterns           | https://microservices.io/patterns/                                                                                              | Saga, Outbox, CQRS                                        |
| Use The Index, Luke                 | https://use-the-index-luke.com/                                                                                                 | SQL Indexing Bibel                                        |
| 12-Factor App                       | https://12factor.net/                                                                                                           | Config, Logging, Concurrency                              |
| Hexagonal Architecture (Cockburn)   | https://alistair.cockburn.us/hexagonal-architecture/                                                                            | Ports & Adapters                                          |
| Clean Architecture (Uncle Bob Blog) | https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html                                                    | Clean Architecture                                        |
| Feature-Sliced Design               | https://feature-sliced.design/                                                                                                  | Frontend Architecture Methodology                         |
| Event Storming                      | https://www.eventstorming.com/                                                                                                  | DDD-Modellierungsmethode                                  |
| OWASP Top 10                        | https://owasp.org/www-project-top-ten/                                                                                          | Web-Security-Risiken Top 10                               |
| OWASP Cheat Sheets                  | https://cheatsheetseries.owasp.org/                                                                                             | Security Best Practices Checklisten                       |
| Auth0: JWT Handbook                 | https://auth0.com/resources/ebooks/jwt-handbook                                                                                 | JWT Deep Dive                                             |
| Martin Fowler: Continuous Delivery  | https://martinfowler.com/bliki/ContinuousDelivery.html                                                                          | CD-Prinzipien                                             |
| Charity Majors: Observability       | https://charity.wtf/                                                                                                            | Observability vs. Monitoring                              |
| Dave Cheney: Table-Driven Tests     | https://dave.cheney.net/2019/05/07/prefer-table-driven-tests                                                                    | Go Testing Patterns                                       |

### GitHub-Repositories

| Repository                          | URL                                                                    | Relevanz                                      |
| ----------------------------------- | ---------------------------------------------------------------------- | --------------------------------------------- |
| Event-Driven Architecture in Golang | https://github.com/PacktPublishing/Event-Driven-Architecture-in-Golang | ES + CQRS + DDD Referenz-Implementation in Go |
| DDD Crew: Context Mapping           | https://github.com/ddd-crew/context-mapping                            | Context Mapping Templates                     |
| DDD Crew: Bounded Context Canvas    | https://github.com/ddd-crew/bounded-context-canvas                     | BC-Definition Tool                            |
| Kamil Grzybek: Modular Monolith     | https://github.com/kgrzybek/modular-monolith-with-ddd                  | Outbox, CQRS, DDD in C#                       |
| Go Standard Project Layout          | https://github.com/golang-standards/project-layout                     | Go-Verzeichnisstruktur                        |
| testcontainers-go                   | https://github.com/testcontainers/testcontainers-go                    | Go Integrationstests mit Containern           |

### Offizielle Dokumentation

| Dokumentation     | URL                                        | Relevanz                         |
| ----------------- | ------------------------------------------ | -------------------------------- |
| PostgreSQL Docs   | https://www.postgresql.org/docs/current/   | Vollständige PostgreSQL-Referenz |
| sqlc Docs         | https://docs.sqlc.dev/                     | sqlc-Referenz                    |
| pgx v5 Docs       | https://pkg.go.dev/github.com/jackc/pgx/v5 | Go PostgreSQL Driver             |
| React Docs        | https://react.dev/                         | Offizielle React-Dok             |
| Tailwind CSS Docs | https://tailwindcss.com/docs               | Tailwind-Referenz                |
| Zod Docs          | https://zod.dev/                           | TypeScript Schema Validation     |
| TanStack Query    | https://tanstack.com/query/latest          | Server State Management          |
| EventStoreDB      | https://developers.eventstore.com/         | Spezialisierter Event Store      |
| Effective Go      | https://go.dev/doc/effective_go            | Go Style Guide                   |
| OWASP Docs        | https://owasp.org/                         | Web Application Security         |
| Docker Docs       | https://docs.docker.com/                   | Container-Referenz               |
| nginx Docs        | https://nginx.org/en/docs/                 | Reverse Proxy Referenz           |
| GitHub Actions    | https://docs.github.com/en/actions         | CI/CD Pipeline-Design            |
| Kubernetes Docs   | https://kubernetes.io/docs/                | Container-Orchestrierung         |
| Prometheus Docs   | https://prometheus.io/docs/                | Monitoring & Alerting            |
| Let's Encrypt     | https://letsencrypt.org/docs/              | TLS Certificate Management       |

---

## Step-by-Step Todo-Liste

Die Überarbeitung folgt einer iterativen Reihenfolge. Jeder Schritt baut auf dem vorherigen auf.

- [x] **Step 1 — Struktur-Refactoring:** `event-sourcing-cqrs.md` in `event-sourcing.md` + `cqrs.md` aufteilen. `postgresql-sqlc.md` → `postgresql.md` umbenennen. In allen Dokumenten jotti-spezifische Abschnitte in Appendix verschieben. Generische Beispiele als Platzhalter einfügen.
- [x] **Step 2 — `ddd.md` überarbeiten:** Event Storming, vollständige Context Mapping Patterns, Sub-Domain-Klassifikation, DDD + Architekturstile, DDD-Lifecycle ergänzen. Externe Quellen einarbeiten.
- [x] **Step 3a — `event-sourcing.md` überarbeiten:** Event Store-Technologien, Event Design, Schema-Evolution, Saga/Outbox/Inbox, Fallstudien, Entscheidungsmatrix ES vs. CRUD. Cross-Referenz zu `cqrs.md`.
- [x] **Step 3b — `cqrs.md` überarbeiten:** Ausbaustufen, Read Model Design, Projektionsstrategien, Eventual Consistency, CQRS ohne ES. Cross-Referenz zu `event-sourcing.md`.
- [x] **Step 4 — `go-backend.md` überarbeiten:** Hexagonal/Clean Architecture, Go Concurrency, Error Handling vertieft, API-Design-Patterns, Resilienz, Observability. **Neuer Abschnitt „Datenbankzugriff / SQL-Tooling":** sqlc Deep Dive, sqlc vs. GORM vs. sqlx vs. ent, Migration-Tooling, Repository Pattern. **Neuer Abschnitt „Testing":** Table-Driven Tests, Testcontainers, httptest, Mocking. Jotti-Code durch allgemeine Beispiele ersetzen.
- [x] **Step 5 — `postgresql.md` überarbeiten:** sqlc-Abschnitte entfernen (nach `go-backend.md` verschoben). MVCC, Indexing Deep Dive, Query Optimization, Advanced Features (LISTEN/NOTIFY, Partitioning), Connection Management, Migration-Strategien, PostgreSQL vs. Alternativen ergänzen.
- [x] **Step 6 — `react-frontend.md` überarbeiten:** 15+ Design Patterns, Frontend Architecture Patterns, State-Management Landscape, Testing, Performance, Accessibility, TypeScript-Patterns.
- [x] **Step 7 — `README.md` aktualisieren:** Neue Dateistruktur (9 Theorie-Dokumente statt 5), Beschreibungen, Themenübersicht, Leseempfehlungen anpassen.
- [x] **Step 8 — Zwischenreview:** Cross-Referenzen (insb. ES↔CQRS), Link-Check, Duplikat-Elimination, Sprachkonsistenz der 7 Kern-Dokumente.
- [ ] **Step 9 — `security.md` erstellen:** Authentication-Patterns (JWT, Session, OAuth2), Authorization (RBAC, ABAC, ReBAC), OWASP Top 10, API-Security, Frontend-Security, Secrets Management, TLS, Security Testing.
- [ ] **Step 10 — `devops.md` erstellen:** Containerisierung (Docker, Compose), Orchestrierung, Reverse Proxy (nginx, Traefik, Caddy), CI/CD Pipeline Design, Deployment-Strategien, Zero-Downtime, Monitoring, Backup, Self-Hosting.
- [ ] **Step 11 — `pos.md` erstellen:** `docs/pos.md` als Grundlage. POS-Geschichte, Architektur-Patterns, Gastro-Workflows, Datenmodelle, Fiskalgesetzgebung, Payment Integration, Non-Profit vs. Commercial, Marktlandschaft.
- [ ] **Step 12 — Abschluss-Review:** Alle 10 Dateien. Cross-Referenzen (Security↔Go-Backend, DevOps↔PostgreSQL, POS↔ES), Link-Check, Duplikat-Elimination, Querschnittsthemen konsistent.
