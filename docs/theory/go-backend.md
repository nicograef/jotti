# Go Backend Architektur — Theorie

Dieses Dokument beschreibt allgemeine Architekturprinzipien für Go-Backends: Schichtenarchitektur, Dependency Injection, Fehlerbehandlung, Middleware und Patterns für ein typsicheres, wartbares Backend. Projektspezifische Anwendungsbeispiele finden sich im [Appendix](#13-appendix-anwendungsbeispiel-jotti).

> **Verwandte Dokumente:**
>
> - [DDD Theorie](ddd.md) — Domain-Driven Design Grundlagen
> - [Event-Sourcing Theorie](event-sourcing.md) — Event-Sourcing Grundlagen
> - [CQRS Theorie](cqrs.md) — Command Query Responsibility Segregation
> - [PostgreSQL](postgresql.md) — Datenbankzugriff
> - [Entwicklung & Deployment](../development.md) — Setup, Tests, CI/CD
> - [Architektur-Übersicht](README.md) — Index aller Theorie-Dokumente

---

## Inhaltsverzeichnis

1. [Architekturprinzipien](#1-architekturprinzipien)
2. [Schichtenarchitektur](#2-schichtenarchitektur)
3. [Dependency Injection und Wiring](#3-dependency-injection-und-wiring)
4. [HTTP-Schicht: Routing, Middleware, Handler](#4-http-schicht-routing-middleware-handler)
5. [Application Services: Commands und Queries](#5-application-services-commands-und-queries)
6. [Domain-Schicht](#6-domain-schicht)
7. [Repository-Schicht](#7-repository-schicht)
8. [Fehlerbehandlung](#8-fehlerbehandlung)
9. [Validierung](#9-validierung)
10. [Authentifizierung und Autorisierung](#10-authentifizierung-und-autorisierung)
11. [Testing-Strategie](#11-testing-strategie)
12. [Go-spezifische Best Practices](#12-go-spezifische-best-practices)
13. [Appendix: Anwendungsbeispiel (jotti)](#13-appendix-anwendungsbeispiel-jotti)
14. [Referenzen](#14-referenzen)

---

## 1. Architekturprinzipien

### Minimalismus

Go's Standardbibliothek (`net/http`, `encoding/json`, `context`) deckt den Großteil der Web-Infrastruktur ab. Kleine bis mittelgroße Go-Webanwendungen kommen oft ohne schwere Frameworks aus und setzen stattdessen auf:

- `net/http` — HTTP-Server und Routing
- `pgx/v5` — PostgreSQL-Driver (direkt, kein ORM)
- `sqlc` — Typsichere SQL-Codegenerierung
- `zerolog` — Structured Logging
- `zog` — Struct-Validierung
- `golang-jwt/v5` — JWT-Handling

**Kein Gin, Echo, Fiber, Chi.** Der Go-stdlib-Router (`http.ServeMux`) reicht für APIs mit einigen Dutzend Endpunkten oft aus.

### Explizitheit über Magie

Go bevorzugt expliziten Code gegenüber Reflection-basierter Magie:

- **Kein ORM** — SQL ist explizit sichtbar, sqlc generiert typsichere Go-Structs
- **Kein Dependency-Injection-Framework** — Constructor-basiertes Manual Wiring
- **Kein Struct-Tag-basiertes Routing** — Handler werden explizit registriert
- **Kein globaler State** — Dependencies werden als Argumente durchgereicht

### API-Design

RESTful und RPC-ähnliche Designs sind beides gängige Optionen. Eine Variante ist eine **POST-only API**, bei der alle Endpunkte ausschließlich HTTP POST verwenden (siehe [Appendix](#post-only-api) für ein konkretes Beispiel). Wichtig ist ein konsistentes Request/Response-Format über alle Endpunkte.

---

## 2. Schichtenarchitektur

### Überblick

```
┌─────────────────────────────────────────────────────────────────┐
│                        HTTP-Schicht                             │
│  api/<domain>/http/                                             │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Handler parsen JSON, validieren, rufen Services auf     │   │
│  │  Middleware: JWT, Rate-Limit, Logging, CORS              │   │
│  └──────────────────────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────────────────┤
│                     Application-Schicht                         │
│  api/<domain>/application/                                      │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Command Services: Orchestrierung, Validierung           │   │
│  │  Query Services: Daten lesen und aufbereiten             │   │
│  └──────────────────────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────────────────┤
│                       Domain-Schicht                            │
│  domain/<domain>/                                               │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Entities, Value Objects, Domain Events, Domain Services │   │
│  │  Geschäftsregeln, Invarianten, Validierungs-Schemas      │   │
│  └──────────────────────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────────────────┤
│                     Repository-Schicht                          │
│  repository/<domain>_repo/                                      │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  DB-Zugriff, sqlc-Wrapper, Fehler-Mapping                │   │
│  └──────────────────────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────────────────┤
│                     Infrastruktur-Schicht                       │
│  db/, config/, app/                                             │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  DB-Verbindung, Konfiguration, Dependency Wiring         │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

### Abhängigkeitsregel

Abhängigkeiten zeigen **von außen nach innen**:

```
HTTP → Application → Domain ← Repository
        ↓                        ↓
     (nutzt)                  (implementiert)
```

- **Domain** hat **keine** Abhängigkeiten (keine Imports aus anderen Schichten)
- **Repository** importiert Domain (für Modelle) und sqlc/dbgen
- **Application** importiert Domain und Repository
- **HTTP** importiert Application und Helper

Eine projektspezifische Verzeichnisstruktur, die diese Schichten abbildet, findet sich im [Appendix](#verzeichnisstruktur).

---

## 3. Dependency Injection und Wiring

### Manual Constructor Injection

Go verwendet keine DI-Frameworks. Dependencies werden über Konstruktoren injiziert:

```go
// Repository erstellen
eventRepo := event_repo.NewRepository(dbPool)
orderRepo := order_repo.NewRepository(dbPool)

// Service erstellen (mit Repository-Dependency)
commandService := order_application.NewCommandService(eventRepo, orderRepo)
queryService := order_application.NewQueryService(eventRepo, orderRepo)

// Handler erstellen (mit Service-Dependency)
commandHandler := order_http.NewCommandHandler(commandService)
queryHandler := order_http.NewQueryHandler(queryService)
```

### App-Struct (Composition Root)

Das Wiring aller Dependencies erfolgt zentral in `app/app.go`:

```go
type App struct {
    DB     *pgxpool.Pool
    Config *config.Config
    Mux    *http.ServeMux
}

func New(cfg *config.Config) (*App, error) {
    pool := db.Connect(cfg.DatabaseURL)
    mux := http.NewServeMux()

    // Alle Repositories, Services, Handler erstellen
    // Routen registrieren

    return &App{DB: pool, Config: cfg, Mux: mux}, nil
}
```

### Vorteile Manual Wiring

| Vorteil                 | Beschreibung                           |
| ----------------------- | -------------------------------------- |
| **Compile-Time Safety** | Fehlende Dependencies = Compile-Fehler |
| **Keine Reflection**    | Kein Runtime-Overhead                  |
| **Explizit**            | Jede Abhängigkeit ist im Code sichtbar |
| **Testbar**             | Mocks über Interfaces injizierbar      |

---

## 4. HTTP-Schicht: Routing, Middleware, Handler

### Routing

Alle Routen werden in dedizierten Dateien registriert:

```go
func RegisterRoutes(mux *http.ServeMux, ...) {
    mux.Handle("/api/orders/create",
        middleware.Chain(commandHandler.CreateOrder,
            middleware.JWT(jwtSecret, "admin", "manager", "staff"),
        ))
}
```

**Konvention:** Route = `/bereich/aktion` (z.B. `/api/orders/create`, `/admin/create-user`)

### Middleware-Stack

Middleware wird in einer **Chain** (von innen nach außen) angewendet:

```
Request → PostMethodOnly → RateLimit → CorrelationID → Logging → JWT → Handler
```

| Middleware                 | Funktion                                           |
| -------------------------- | -------------------------------------------------- |
| `PostMethodOnlyMiddleware` | Lehnt nicht-POST-Requests ab (405)                 |
| `RateLimitMiddleware`      | Token-Bucket per IP (10 req/s)                     |
| `CorrelationIDMiddleware`  | Setzt `X-Correlation-ID` Header (UUID)             |
| `LoggingMiddleware`        | Loggt Path, Status, Duration, Correlation-ID       |
| `JWTMiddleware`            | Validiert JWT, extrahiert UserID + Role in Context |

### Handler-Pattern

Jeder Handler folgt einem einheitlichen Muster:

```go
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
    // 1. Request parsen (JSON → Struct)
    var req CreateOrderRequest
    if err := helper.ParseJSON(r, &req); err != nil {
        helper.WriteError(w, http.StatusBadRequest, "invalid_json")
        return
    }

    // 2. UserID aus JWT-Context extrahieren
    userID, _ := r.Context().Value(middleware.UserIDKey).(int)

    // 3. Application Service aufrufen
    err := h.service.CreateOrder(r.Context(), userID, req.OrderID, ...)
    if err != nil {
        helper.HandleDomainError(w, err)
        return
    }

    // 4. Erfolgsresponse
    helper.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
```

### Request/Response-Format

**Requests:** Immer JSON Body mit strenger Validierung (unbekannte Felder → Fehler).

**Responses:** Einheitliches Format:

```json
// Erfolg
{"status": "ok"}
{"data": {...}}

// Fehler
{"code": "order_not_found"}
{"code": "validation_error", "details": {...}}
```

---

## 5. Application Services: Commands und Queries

### Command Service

Verantwortlich für **schreibende Operationen**. Orchestriert Domain-Logik und Repositories:

```go
type CommandService struct {
    eventRepo EventRepository
    orderRepo OrderRepository
}

func (s *CommandService) CreateOrder(ctx, userID, orderID, items, comment) error {
    // 1. Order laden (existiert? aktiv?)
    order, err := s.orderRepo.Get(ctx, orderID)
    // 2. Event erstellen (Domain-Logik)
    event := order.NewOrderCreatedEvent(userID, orderID, items, comment)
    // 3. Event persistieren
    _, err = s.eventRepo.WriteEvent(ctx, event)
    // 4. Snapshot aktualisieren
    return s.UpdateOrderSnapshot(ctx, userID, orderID)
}
```

### Query Service

Verantwortlich für **lesende Operationen**. Liest Events und rekonstruiert Zustand:

```go
type QueryService struct {
    eventRepo EventRepository
    orderRepo OrderRepository
}

func (s *QueryService) GetOrderBalance(ctx, orderID) (int, error) {
    // 1. Events ab letztem Snapshot laden
    events, err := s.eventRepo.ReadEventsWithSnapshot(ctx, subject, snapshotType)
    // 2. Zustand in Domain-Schicht rekonstruieren
    return order.GetBalanceFromEvents(events), nil
}
```

### Factory-Pattern für Services

Services werden über Factory-Funktionen erstellt, die alle Dependencies injizieren:

```go
func NewCommandHandler(db *pgxpool.Pool) *CommandHandler {
    eventRepo := event_repo.NewRepository(db)
    orderRepo := order_repo.NewRepository(db)
    service := NewCommandService(eventRepo, orderRepo)
    return &CommandHandler{service: service}
}
```

---

## 6. Domain-Schicht

### Grundregeln

- **Keine DB-Abhängigkeiten** — kein `import` von `pgx`, `sqlc`, etc.
- **Keine HTTP-Abhängigkeiten** — kein `import` von `net/http`
- **Reine Geschäftslogik** — Validierung, Berechnung, Event-Erstellung
- **Testbar ohne Infrastruktur** — Unit-Tests ohne DB oder Server

### Entities mit zog-Validierung

```go
type Order struct {
    ID     int    `json:"id"`
    Name   string `json:"name"`
    Status string `json:"status"`
}

var orderSchema = zog.Struct(zog.Shape{
    "name":   zog.String().Min(1).Max(100),
    "status": zog.String().OneOf("active", "inactive", "deleted"),
})
```

### Domain Services (reine Funktionen)

Zustandsrekonstruktion als zustandslose Funktionen:

```go
// domain/order/events.go
func GetBalanceFromEvents(events []event.Event) int { ... }
func GetUnpaidItemsFromEvents(events []event.Event) []Item { ... }
func GetUndeliveredItemsFromEvents(events []event.Event) []Item { ... }
func GetHistoryFromEvents(events []event.Event) []any { ... }
```

**Vorteile reiner Funktionen:**
- Deterministisch (gleiche Eingabe → gleiches Ergebnis)
- Keine Seiteneffekte
- Einfach zu testen
- Parallelisierbar

---

## 7. Repository-Schicht

### Interface und Implementierung

Jedes Repository hat ein Interface (implizit in Go) und eine Implementierung:

```go
// Nutzung in Application Service (Interface)
type EventRepository interface {
    WriteEvent(ctx context.Context, event event.Event) (int, error)
    ReadEventsWithSnapshot(ctx context.Context, subject, snapshotType string) ([]event.Event, error)
}

// Implementierung (wraps sqlc)
type Repository struct {
    db *pgxpool.Pool
    q  *dbgen.Queries
}

func (r *Repository) WriteEvent(ctx context.Context, e event.Event) (int, error) {
    id, err := r.q.WriteEvent(ctx, dbgen.WriteEventParams{...})
    return int(id), mapError(err)
}
```

### Fehler-Mapping

PostgreSQL-Fehler werden in Domain-Fehler übersetzt:

```go
func mapError(err error) error {
    if err == nil { return nil }
    if err == pgx.ErrNoRows { return domain.ErrNotFound }

    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) {
        switch pgErr.Code {
        case "23505": return domain.ErrAlreadyExists  // Unique Violation
        case "23503": return domain.ErrForeignKey      // FK Violation
        }
    }
    return domain.ErrDatabase
}
```

### Mock-Repositories für Tests

Jedes Repository hat eine Mock-Implementierung für Unit-Tests:

```go
// repository/event_repo/mock.go
type MockRepository struct {
    Events []event.Event
}

func (m *MockRepository) WriteEvent(ctx context.Context, e event.Event) (int, error) {
    m.Events = append(m.Events, e)
    return len(m.Events), nil
}
```

---

## 8. Fehlerbehandlung

### Fehler-Philosophie

Go verwendet **explizite Fehlerbehandlung** (kein try/catch). Jeder Fehler wird entweder:

1. **Behandelt** — Recovery-Logik, Default-Wert
2. **Weitergereicht** — `return err` oder `return fmt.Errorf("context: %w", err)`
3. **Geloggt und als HTTP-Fehler zurückgegeben**

### Fehler-Schichten

```
PostgreSQL Error
     ↓ mapError()
Domain Error (ErrNotFound, ErrAlreadyExists)
     ↓ return err
Application Service
     ↓ return err
HTTP Handler
     ↓ HandleDomainError()
HTTP Response: {"code": "order_not_found"}
```

### Fehler-Typen

| Fehler-Typ        | HTTP-Status | Code                    |
| ----------------- | ----------- | ----------------------- |
| Validierung       | 400         | `validation_error`      |
| JSON ungültig     | 400         | `invalid_json`          |
| Nicht gefunden    | 404         | `*_not_found`           |
| Bereits vorhanden | 409         | `*_already_exists`      |
| Nicht autorisiert | 401         | `unauthorized`          |
| Kein Zugriff      | 403         | `forbidden`             |
| Rate Limit        | 429         | `rate_limit_exceeded`   |
| Server-Fehler     | 500         | `internal_server_error` |

---

## 9. Validierung

### Zwei-Schichten-Validierung

Validierung erfolgt auf **beiden Seiten** (Frontend + Backend):

```
Client (Zod) → HTTP (JSON Parsing) → Domain (zog Schema) → Database (Constraints)
```

| Schicht      | Tool                   | Prüft                                      |
| ------------ | ---------------------- | ------------------------------------------ |
| Frontend     | Zod                    | UI-Validierung, sofortiges Feedback        |
| HTTP-Handler | `helper.ParseJSON`     | JSON-Syntax, bekannte Felder               |
| Domain       | zog                    | Geschäftsregeln (Min/Max, OneOf, Required) |
| Database     | PostgreSQL Constraints | Eindeutigkeit, Fremdschlüssel, NOT NULL    |

### zog im Backend

```go
var orderSchema = zog.Struct(zog.Shape{
    "orderId": zog.Int().Min(1),
    "items":   zog.Slice(itemSchema).Min(1),
    "comment": zog.String().Max(500).Optional(),
})

// Nutzung
errs := orderSchema.Parse(data, &result)
if errs != nil {
    return ValidationError(errs)
}
```

### Striktes JSON-Parsing

`helper.ParseJSON` verwendet `json.Decoder` mit `DisallowUnknownFields()`:

```go
func ParseJSON(r *http.Request, dst any) error {
    dec := json.NewDecoder(r.Body)
    dec.DisallowUnknownFields()  // Unbekannte Felder → Fehler
    return dec.Decode(dst)
}
```

---

## 10. Authentifizierung und Autorisierung

### JWT-basierte Auth

```
Login → JWT (HS256, 12h) → Bearer Token im Header → Middleware validiert → Context
```

| Claim  | Wert                             | Beschreibung            |
| ------ | -------------------------------- | ----------------------- |
| `sub`  | UserID (int)                     | Benutzer-Identifikation |
| `role` | admin / senior_service / service | Berechtigung            |
| `exp`  | Unix Timestamp                   | Ablaufzeit (12 Stunden) |

### Rollen-basierte Autorisierung

Middleware prüft erlaubte Rollen pro Route:

```go
// Nur bestimmte Rolle
middleware.JWT(secret, "admin")

// Mehrere Rollen erlaubt
middleware.JWT(secret, "admin", "senior_service", "service")

// Eingeschränkter Zugriff
middleware.JWT(secret, "admin", "senior_service")
```

### Passwort-Hashing

Argon2id (aktuell stärkster Passwort-Hash-Algorithmus):

```go
// domain/user/password.go
func HashPassword(password string) (string, error) { ... }
func VerifyPassword(hash, password string) bool { ... }
```

---

## 11. Testing-Strategie

### Unit-Tests (Build-Tag: `unit`)

```bash
cd backend && go test -tags=unit -race ./...
```

- Testen Domain-Logik und Application Services
- Mock-Repositories statt echte DB
- Build-Tag `//go:build unit` verhindert versehentliches Ausführen mit DB-Tests

### Integrationstests

```bash
./test-integration.sh
```

- Testen gegen echte PostgreSQL-Instanz (Docker)
- Prüfen SQL-Queries, Constraints, Trigger
- Nutzen Test-Utilities aus `db/testing.go`

### Testbare Architektur

Die Schichtenarchitektur unterstützt Testbarkeit:

| Schicht     | Test-Art    | Mocking                              |
| ----------- | ----------- | ------------------------------------ |
| Domain      | Unit        | Keine Mocks nötig (reine Funktionen) |
| Application | Unit        | Mock-Repositories                    |
| Repository  | Integration | Echte DB                             |
| HTTP        | Integration | Echte Server + Test-Client           |

---

## 12. Go-spezifische Best Practices

### Context-Propagierung

`context.Context` wird durch alle Schichten propagiert:

```go
func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    userID := ctx.Value(middleware.UserIDKey).(int)
    result, err := h.service.DoSomething(ctx, userID, ...)
}
```

### Structured Logging mit zerolog

```go
log.Info().
    Str("path", r.URL.Path).
    Int("status", status).
    Dur("duration", duration).
    Str("correlation_id", correlationID).
    Msg("request completed")
```

### Connection Pooling mit pgx

`pgxpool.Pool` verwaltet DB-Connections effizient:

```go
pool, err := pgxpool.New(ctx, databaseURL)
// Pool wird einmal erstellt und überall geteilt
// Automatische Connection-Verwaltung
```

### Graceful Shutdown

```go
srv := &http.Server{Addr: ":8080", Handler: mux}
go srv.ListenAndServe()

// Auf Signal warten
<-ctx.Done()
srv.Shutdown(context.Background())
pool.Close()
```

---

## 13. Appendix: Anwendungsbeispiel (jotti)

Die folgenden Abschnitte zeigen, wie die oben beschriebenen Prinzipien konkret im jotti-Projekt (Gastronomie-Kassensystem) angewendet werden.

### Verzeichnisstruktur

```
backend/
├── main.go                          # Einstiegspunkt: Server starten
├── app/app.go                       # Dependency Wiring (alle Handler/Services)
├── config/config.go                 # Konfiguration aus Umgebungsvariablen
├── db/db.go                         # DB-Verbindung und Connection Pool
│
├── domain/                          # Domain-Schicht (keine DB-Abhängigkeit)
│   ├── event/event.go               # Event-Modell (CloudEvents-inspiriert)
│   ├── table/                       # Tisch-Aggregat
│   │   ├── tisch.go                 # Entity + Validierung
│   │   ├── events.go                # Zustandsrekonstruktion (Domain Services)
│   │   ├── bestellung.go            # Value Objects
│   │   ├── zahlung.go
│   │   ├── stornierung.go
│   │   ├── lieferung.go
│   │   └── *Event.go                # Event-Structs + Konstruktoren
│   ├── product/product.go           # Entity + Varianten
│   └── user/user.go                 # Entity + Password-Hashing
│
├── repository/                      # Repository-Schicht (sqlc-Wrapper)
│   ├── event_repo/repo.go           # Append-only Event Store
│   ├── table_repo/repo.go           # Tisch-CRUD
│   ├── product_repo/repo.go         # Produkt+Varianten-CRUD
│   └── user_repo/repo.go            # Benutzer-CRUD
│
├── api/                             # HTTP + Application Schicht
│   ├── service.go                   # Service-Routen registrieren
│   ├── admin.go                     # Admin-Routen registrieren
│   ├── auth.go                      # Auth-Routen registrieren
│   ├── senior_service.go            # Senior-Service-Routen
│   ├── middleware/middleware.go      # JWT, Rate-Limit, Logging, etc.
│   ├── helper/http.go               # JSON-Parsing, Response-Helper
│   └── <domain>/
│       ├── http/command.go           # HTTP Handler (Commands)
│       ├── http/query.go            # HTTP Handler (Queries)
│       ├── application/command.go   # Command Service
│       └── application/query.go     # Query Service
│
└── sqlc/
    ├── queries/*.sql                # SQL-Queries (Eingabe für sqlc)
    └── dbgen/                       # Generierter Code (NICHT EDITIEREN)
```

### POST-only API

Alle API-Endpunkte in jotti verwenden HTTP POST. Keine GET/PUT/DELETE. Vorteile:

- **Einheitliches Request-Format:** Immer JSON Body
- **Keine URL-Parameter:** Keine Injection über URL-Encoding
- **Cache-Busting:** POST-Responses werden nicht gecacht (kein Browser-Cache-Problem)
- **Einfache Middleware:** Nur ein HTTP-Methoden-Check

### Service-Routen

```go
// api/service.go — Service-Routen
func RegisterServiceRoutes(mux *http.ServeMux, ...) {
    mux.Handle("/service/bestellung-aufgeben",
        middleware.Chain(commandHandler.BestellungAufgeben,
            middleware.JWT(jwtSecret, "admin", "senior_service", "service"),
        ))
}
```

### Handler-Pattern (BestellungAufgeben)

```go
func (h *Handler) BestellungAufgeben(w http.ResponseWriter, r *http.Request) {
    // 1. Request parsen (JSON → Struct)
    var req BestellungAufgebenRequest
    if err := helper.ParseJSON(r, &req); err != nil {
        helper.WriteError(w, http.StatusBadRequest, "invalid_json")
        return
    }

    // 2. UserID aus JWT-Context extrahieren
    userID, _ := r.Context().Value(middleware.UserIDKey).(int)

    // 3. Application Service aufrufen
    err := h.service.BestellungAufgeben(r.Context(), userID, req.TischID, ...)
    if err != nil {
        helper.HandleDomainError(w, err)
        return
    }

    // 4. Erfolgsresponse
    helper.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
```

### Command Service

```go
type CommandService struct {
    eventRepo EventRepository
    tableRepo TableRepository
}

func (s *CommandService) BestellungAufgeben(ctx, userID, tischID, positionen, comment) error {
    // 1. Tisch laden (existiert? aktiv?)
    tisch, err := s.tableRepo.Get(ctx, tischID)
    // 2. Event erstellen (Domain-Logik)
    event := table.NewBestellungAufgegebenEvent(userID, tischID, positionen, comment)
    // 3. Event persistieren
    _, err = s.eventRepo.WriteEvent(ctx, event)
    // 4. Snapshot aktualisieren
    return s.TischSnapshotErstellen(ctx, userID, tischID)
}
```

### Query Service

```go
type QueryService struct {
    eventRepo EventRepository
    tableRepo TableRepository
}

func (s *QueryService) GetTischSaldo(ctx, tischID) (int, error) {
    // 1. Events ab letztem Snapshot laden
    events, err := s.eventRepo.ReadEventsWithSnapshot(ctx, subject, snapshotType)
    // 2. Zustand in Domain-Schicht rekonstruieren
    return table.GetSaldoFromEvents(events), nil
}
```

### Factory-Pattern

```go
func NewCommandHandler(db *pgxpool.Pool) *CommandHandler {
    eventRepo := event_repo.NewRepository(db)
    tableRepo := table_repo.NewRepository(db)
    service := NewCommandService(eventRepo, tableRepo)
    return &CommandHandler{service: service}
}
```

### Domain-Modell (Tisch)

```go
type Tisch struct {
    ID     int    `json:"id"`
    Name   string `json:"name"`
    Status string `json:"status"`
}

var tischSchema = zog.Struct(zog.Shape{
    "name":   zog.String().Min(1).Max(100),
    "status": zog.String().OneOf("active", "inactive", "deleted"),
})
```

### Domain Services (Zustandsrekonstruktion)

```go
// domain/table/events.go
func GetSaldoFromEvents(events []event.Event) int { ... }
func GetUnbezahltePositionenFromEvents(events []event.Event) []Position { ... }
func GetUngeliefertePositionenFromEvents(events []event.Event) []Position { ... }
func GetHistoryFromEvents(events []event.Event) []any { ... }
```

### Validierungs-Schema (Bestellung)

```go
var bestellungSchema = zog.Struct(zog.Shape{
    "tischId":    zog.Int().Min(1),
    "positionen": zog.Slice(positionSchema).Min(1),
    "comment":    zog.String().Max(500).Optional(),
})
```

---

## 14. Referenzen

### Go-Architektur

- [Standard Go Project Layout](https://github.com/golang-standards/project-layout) — Verzeichnisstruktur-Konventionen
- [Effective Go](https://go.dev/doc/effective_go) — Offizieller Style Guide
- [Go Proverbs](https://go-proverbs.github.io/) — Design-Philosophie

### Event-Driven Go

- [Event-Driven Architecture in Golang](https://github.com/PacktPublishing/Event-Driven-Architecture-in-Golang) — DDD + ES + CQRS in Go

### Spezifische Tools

- [pgx v5 Dokumentation](https://pkg.go.dev/github.com/jackc/pgx/v5) — PostgreSQL-Driver
- [sqlc Dokumentation](https://docs.sqlc.dev/) — SQL → Go Code-Generator
- [zerolog](https://pkg.go.dev/github.com/rs/zerolog) — Structured Logging
- [golang-jwt/v5](https://pkg.go.dev/github.com/golang-jwt/jwt/v5) — JWT-Handling
