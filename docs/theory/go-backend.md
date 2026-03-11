# Go Backend Architektur — Theorie

Dieses Dokument ist ein allgemeiner Architektur-Guide für Go-Backends: Architektur-Patterns, HTTP-Ökosystem, API-Design, Concurrency, Fehlerbehandlung, Resilienz, Observability, Datenbankzugriff und Testing.

---

## Inhaltsverzeichnis

1. [Architektur-Patterns im Vergleich](#1-architektur-patterns-im-vergleich)
2. [Go-HTTP-Ökosystem](#2-go-http-ökosystem)
3. [API-Design-Patterns](#3-api-design-patterns)
4. [Dependency Injection und Wiring](#4-dependency-injection-und-wiring)
5. [HTTP-Schicht: Routing, Middleware, Handler](#5-http-schicht-routing-middleware-handler)
6. [Application Services: Commands und Queries](#6-application-services-commands-und-queries)
7. [Domain-Schicht](#7-domain-schicht)
8. [Fehlerbehandlung in Go](#8-fehlerbehandlung-in-go)
9. [Validierung](#9-validierung)
10. [Go Concurrency Patterns](#10-go-concurrency-patterns)
11. [Configuration & 12-Factor App](#11-configuration--12-factor-app)
12. [Resilienz-Patterns](#12-resilienz-patterns)
13. [Observability](#13-observability)
14. [Datenbankzugriff / SQL-Tooling](#14-datenbankzugriff--sql-tooling)
15. [Testing in Go](#15-testing-in-go)
16. [Referenzen](#16-referenzen)

---

## 1. Architektur-Patterns im Vergleich

### 1.1 Layered Architecture (Schichtenarchitektur)

Die Schichtenarchitektur ist das klassische Muster für Backend-Systeme. Jede Schicht hat eine klar definierte Verantwortung und kommuniziert nur mit der darunterliegenden Schicht.

```
┌─────────────────────────────────────────────────────────────────┐
│                        HTTP-Schicht                             │
│  Handler: Request parsen, validieren, Services aufrufen         │
│  Middleware: Auth, Rate-Limit, Logging, CORS                    │
├─────────────────────────────────────────────────────────────────┤
│                     Application-Schicht                         │
│  Command Services: Orchestrierung von Schreiboperationen        │
│  Query Services: Datenabruf und Aufbereitung                    │
├─────────────────────────────────────────────────────────────────┤
│                       Domain-Schicht                            │
│  Entities, Value Objects, Domain Events, Domain Services        │
│  Geschäftsregeln und Invarianten                                │
├─────────────────────────────────────────────────────────────────┤
│                     Repository-Schicht                          │
│  DB-Zugriff, Fehler-Mapping, Query-Ausführung                   │
├─────────────────────────────────────────────────────────────────┤
│                     Infrastruktur-Schicht                       │
│  DB-Verbindung, Konfiguration, Dependency Wiring                │
└─────────────────────────────────────────────────────────────────┘
```

**Abhängigkeitsregel:** Abhängigkeiten zeigen von außen nach innen:

```
HTTP → Application → Domain ← Repository
```

- **Domain** hat keine Abhängigkeiten zu anderen Schichten
- **Repository** importiert Domain (für Modelle) und DB-Tooling
- **Application** importiert Domain und Repository-Interfaces
- **HTTP** importiert Application und Infrastruktur-Helper

**Stärken:** Einfach verständlich, gute Testbarkeit, klare Verantwortlichkeiten.
**Schwächen:** Strenge Schichtgrenzen können zu unnötigen Abstraktionen führen; Domain-Objekte neigen dazu, Schichtgrenzen zu "durchbluten".

### 1.2 Hexagonal Architecture (Ports & Adapters)

Alistair Cockburn (2005): Das System hat einen **Kern** (Domain + Application Logic) und kommuniziert mit der Außenwelt ausschließlich über **Ports** (Interfaces) und **Adapters** (Implementierungen).

```
                    ┌─────────────────────┐
    HTTP Adapter ──►│                     │◄── DB Adapter
    gRPC Adapter ──►│    Core (Hexagon)   │◄── Cache Adapter
    CLI Adapter  ──►│  Domain + App Logic │◄── Message Queue Adapter
                    │                     │──► Email Adapter
                    └─────────────────────┘
         Primary Ports (Driving)    Secondary Ports (Driven)
```

- **Primary Ports (Left Side / Driving):** Eingaben von außen (HTTP, CLI, Tests). Der Aufrufer treibt das System.
- **Secondary Ports (Right Side / Driven):** Ausgaben nach außen (DB, Messaging, Email). Das System treibt Externe.
- **Ports:** Interfaces im Core (z.B. `OrderRepository`, `EmailSender`)
- **Adapters:** Konkrete Implementierungen (z.B. `PostgresOrderRepository`, `SMTPEmailSender`)

```go
// Port (Interface im Core)
type OrderRepository interface {
    Save(ctx context.Context, order Order) error
    FindByID(ctx context.Context, id int) (Order, error)
}

// Adapter (Implementierung außerhalb des Core)
type PostgresOrderRepository struct {
    db *pgxpool.Pool
}

func (r *PostgresOrderRepository) Save(ctx context.Context, order Order) error {
    _, err := r.db.Exec(ctx, "INSERT INTO orders ...", order.ID, order.Status)
    return err
}
```

**Vorteile:** Core vollständig testbar ohne externe Dependencies (Mock-Adapters); Adapters austauschbar (PostgreSQL → SQLite für Tests); keine Framework-Kopplung im Core.

### 1.3 Clean Architecture (Uncle Bob)

Robert C. Martin: Vier konzentrische Ringe, wobei die **Dependency Rule** gilt: Abhängigkeiten zeigen immer nach innen (zur stabileren Schicht).

```
         ┌──────────────────────────────────────────┐
         │  Frameworks & Drivers (HTTP, DB, UI)     │
         │  ┌──────────────────────────────────────┐│
         │  │  Interface Adapters (Controllers,    ││
         │  │  Gateways, Presenters)               ││
         │  │  ┌──────────────────────────────────┐││
         │  │  │  Application Business Rules      │││
         │  │  │  (Use Cases)                     │││
         │  │  │  ┌──────────────────────────────┐│││
         │  │  │  │  Enterprise Business Rules   ││││
         │  │  │  │  (Entities)                  ││││
         │  │  │  └──────────────────────────────┘│││
         │  │  └──────────────────────────────────┘││
         │  └──────────────────────────────────────┘│
         └──────────────────────────────────────────┘
```

| Ring                 | Inhalt                                     | Änderungsfrequenz |
| -------------------- | ------------------------------------------ | ----------------- |
| Entities             | Enterprise-Geschäftsregeln, Domain-Modelle | Sehr selten       |
| Use Cases            | Anwendungs-Geschäftsregeln                 | Selten            |
| Interface Adapters   | Controller, Gateway, Presenter             | Gelegentlich      |
| Frameworks & Drivers | HTTP, DB, UI                               | Häufig            |

**Konsequenz:** Entities kennen keine Use Cases; Use Cases kennen keine HTTP-Adapter; keine Framework-Imports im Domain-Code.

### 1.4 Onion Architecture

Jeffrey Palermo (2008): Ähnlich wie Clean Architecture, betont aber explizit den **Domain Model** als innersten Ring und **Application Services** als Vermittler.

```
         ┌─ Infrastructure ───────────────────────┐
         │  ┌─ Application Services ──────────────┤
         │  │  ┌─ Domain Services ────────────────┤
         │  │  │  ┌─ Domain Model ───────────────┐│
         │  │  │  │  (Entities, Value Objects)  ││
         │  │  │  └─────────────────────────────┘│
         │  │  └─────────────────────────────────┘
         │  └─────────────────────────────────────┘
         └─────────────────────────────────────────┘
```

### 1.5 Vergleichstabelle

| Pattern            | Hauptprinzip                    | Go-Umsetzung                          | Empfohlen wenn...                             |
| ------------------ | ------------------------------- | ------------------------------------- | --------------------------------------------- |
| Layered            | Schichten oben → unten          | Pakete je Schicht (`api/`, `domain/`) | Einfache APIs, gute Einsteiger-Eignung        |
| Hexagonal          | Core + Ports + Adapters         | Interfaces + Pakete je Adapter        | Viele externe Systeme, hohe Testanforderungen |
| Clean Architecture | Dependency Rule, konzentrisch   | Wie Hexagonal, strenger               | Komplexe Domains mit langer Lebensdauer       |
| Onion              | Domain im Zentrum, kein Adapter | Wie Hexagonal, Domain-first           | DDD-intensive Projekte                        |

> **Praxis-Empfehlung:** Für Go-Services beginnt man oft mit einer einfachen Layered Architecture und extrahiert bei Bedarf Ports (Interfaces) für testbare Boundaries — das entspricht einer pragmatischen Hexagonal Architecture ohne strikte Framework-Grenzen.

---

## 2. Go-HTTP-Ökosystem

### 2.1 stdlib `net/http`

Seit **Go 1.22** (Februar 2024) unterstützt der `http.ServeMux` Method Matching und Wildcards direkt:

```go
mux := http.NewServeMux()

// Go 1.22+: Method + Path Pattern
mux.HandleFunc("GET /products/{id}", handlers.GetProduct)
mux.HandleFunc("POST /orders", handlers.CreateOrder)
mux.HandleFunc("DELETE /orders/{id}", handlers.DeleteOrder)

// Path-Parameter lesen
func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")  // "id" aus {id}
    // ...
}
```

**Stärken:** Keine externe Dependency, maximale Kontrolle, Compile-Time-Stabilität.
**Schwächen:** Kein eingebautes Middleware-Chaining, kein Named-Routes-System, kein automatisches Request-Binding.

### 2.2 Chi

Leichtgewichtiger Router, der das `net/http`-Interface vollständig respektiert:

```go
import "github.com/go-chi/chi/v5"

r := chi.NewRouter()
r.Use(middleware.Logger)
r.Use(middleware.Recoverer)

r.Route("/orders", func(r chi.Router) {
    r.Get("/", handlers.ListOrders)
    r.Post("/", handlers.CreateOrder)
    r.Route("/{id}", func(r chi.Router) {
        r.Get("/", handlers.GetOrder)
        r.Delete("/", handlers.DeleteOrder)
    })
})
```

**Stärken:** Middleware-First-Design, Sub-Routing, Composability, 100% stdlib-kompatibel.

### 2.3 Echo, Gin, Fiber

| Framework | Stars (2024) | Philosophie                   | Besonderheiten                              |
| --------- | ------------ | ----------------------------- | ------------------------------------------- |
| **Echo**  | ~29k         | Minimalismus, gute Middleware | Binder für JSON/XML/Form, Validator-Hookups |
| **Gin**   | ~77k         | Performance, einfaches API    | Eigenes Context-Objekt, schnelles Routing   |
| **Fiber** | ~32k         | Express-Inspiration, Fasthttp | Nicht stdlib-kompatibel (Fasthttp-Context)  |

### 2.4 Entscheidungshilfe

| Situation                                               | Empfehlung                         |
| ------------------------------------------------------- | ---------------------------------- |
| Neues Projekt, wenige Routes, kein Framework-Lock       | stdlib `net/http` (Go 1.22+)       |
| Middleware-heavy, Sub-Router, `http.Handler`-kompatibel | Chi                                |
| Schneller Einstieg, viel Dokumentation, große Community | Gin oder Echo                      |
| Migration von Express (Node.js), Performance-kritisch   | Fiber (aber Vorsicht: kein stdlib) |

---

## 3. API-Design-Patterns

### 3.1 REST

**Representational State Transfer** nutzt HTTP-Methoden semantisch: `GET` für Lesen, `POST` für Erstellen, `PUT`/`PATCH` für Ändern, `DELETE` für Löschen. Ressourcen sind über URLs adressiert.

```
GET    /products          → Liste aller Produkte
GET    /products/42       → Produkt 42
POST   /products          → Neues Produkt erstellen
PATCH  /products/42       → Produkt 42 partiell aktualisieren
DELETE /products/42       → Produkt 42 löschen
```

**Stärken:** Weit verbreitet, gut dokumentiert, Browser-/Cache-freundlich.
**Schwächen:** HTTP-Methoden-Semantik kann bei komplexen Domänen-Aktionen künstlich wirken ("Wie heißt die Route für `cancelOrder`?").

### 3.2 POST-only API (RPC-Stil)

Alle Endpunkte verwenden HTTP POST. Namen sind Aktionen (Verben), nicht Ressourcen (Substantive):

```
POST /orders/create
POST /orders/cancel
POST /orders/confirm-delivery
POST /payments/register
```

**Stärken:** Einheitliches Request-Format (immer JSON-Body), keine URL-Parameter-Sanitierung, kein Caching-Problem.
**Schwächen:** Kein HTTP-Methoden-Caching, erfordert API-Dokumentation für Entdeckbarkeit.

### 3.3 gRPC

**gRPC** nutzt HTTP/2 und Protocol Buffers für stark typisierte, bi-direktionale Kommunikation:

```protobuf
service OrderService {
  rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse);
  rpc StreamOrders(StreamOrdersRequest) returns (stream OrderEvent);
}
```

**Stärken:** Starke Typisierung durch Protobuf, automatisches Code-Generierung für Client + Server, Streaming, HTTP/2 Multiplexing.
**Schwächen:** Nicht browser-nativ (benötigt gRPC-Web-Proxy), schwererer Setup, .proto-Schema-Management.

### 3.4 GraphQL

Clients definieren **genau**, welche Felder sie brauchen. Kein Over- oder Under-Fetching:

```graphql
query {
  order(id: 42) {
    id
    status
    items {
      name
      quantity
    }
    customer {
      name
      email
    }
  }
}
```

**Stärken:** Flexible Abfragen, reduziert Over-Fetching, ein Endpoint für alle Queries.
**Schwächen:** Komplexität bei Authorization (Field-Level), N+1-Problem bei naiven Implementierungen (DataLoader-Pattern nötig), kein HTTP-Caching.

### 3.5 API-Versionierung

| Strategie       | Beispiel                              | Bewertung                           |
| --------------- | ------------------------------------- | ----------------------------------- |
| URL-Pfad        | `/v1/orders`, `/v2/orders`            | Einfachste, am weitesten verbreitet |
| HTTP-Header     | `Accept: application/vnd.api.v2+json` | Sauber, aber unsichtbar für Browser |
| Query-Parameter | `/orders?version=2`                   | Einfach, aber URL-Verschmutzung     |
| Subdomain       | `v2.api.example.com`                  | DNS-Overhead, selten empfohlen      |

### 3.6 Paginierung

```go
// Offset-basiert (einfach, aber ineffizient bei großen Datasets)
GET /orders?page=3&limit=25

// Cursor-basiert (effizient für große Datasets)
GET /orders?cursor=eyJpZCI6NDJ9&limit=25

// Response-Format
{
    "data": [...],
    "pagination": {
        "cursor": "eyJpZCI6NjZ9",
        "has_more": true,
        "total": 1234
    }
}
```

### 3.7 Rate Limiting

| Algorithmus        | Charakteristik                                       | Geeignet für                    |
| ------------------ | ---------------------------------------------------- | ------------------------------- |
| **Token Bucket**   | Erlaubt Bursts bis zur Bucket-Kapazität              | APIs mit gelegentlichen Bursts  |
| **Leaky Bucket**   | Konstanter Outflow, glättet Bursts                   | Upstream-Schutz                 |
| **Fixed Window**   | Zählt Anfragen in fixen Zeitfenstern (z.B. 1 Minute) | Einfach, aber Grenzfall-Problem |
| **Sliding Window** | Gleitendes Zeitfenster, kein Grenzfall-Problem       | Präzises Rate Limiting          |

---

## 4. Dependency Injection und Wiring

### Manual Constructor Injection

Go verwendet keine DI-Frameworks. Dependencies werden über Konstruktoren injiziert und sind **zur Compile-Zeit** sichtbar:

```go
// Repository erstellen
productRepo := product_repo.NewRepository(dbPool)
orderRepo   := order_repo.NewRepository(dbPool)

// Service erstellen (mit Repository-Dependency)
commandService := order_application.NewCommandService(productRepo, orderRepo)
queryService   := order_application.NewQueryService(orderRepo)

// Handler erstellen (mit Service-Dependency)
commandHandler := order_http.NewCommandHandler(commandService)
queryHandler   := order_http.NewQueryHandler(queryService)
```

### Composition Root

Das Wiring aller Dependencies erfolgt zentral an einem einzigen Ort (_Composition Root_, oft `app/app.go` oder `main.go`):

```go
type App struct {
    DB     *pgxpool.Pool
    Config *config.Config
    Mux    *http.ServeMux
}

func New(cfg *config.Config) (*App, error) {
    pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
    if err != nil {
        return nil, fmt.Errorf("db connect: %w", err)
    }

    mux := http.NewServeMux()

    // Repositories
    productRepo := product_repo.NewRepository(pool)
    orderRepo   := order_repo.NewRepository(pool)

    // Services
    orderCmd := order_application.NewCommandService(productRepo, orderRepo)
    orderQry := order_application.NewQueryService(orderRepo)

    // Handler & Routes
    order_http.RegisterRoutes(mux, orderCmd, orderQry)

    return &App{DB: pool, Config: cfg, Mux: mux}, nil
}
```

### Vorteile Manual Wiring

| Vorteil                 | Beschreibung                           |
| ----------------------- | -------------------------------------- |
| **Compile-Time Safety** | Fehlende Dependencies → Compile-Fehler |
| **Keine Reflection**    | Kein Runtime-Overhead                  |
| **Explizit**            | Jede Abhängigkeit ist im Code sichtbar |
| **Testbar**             | Mocks über Interfaces injizierbar      |
| **Debuggbar**           | Stack Traces führen direkt zur Ursache |

---

## 5. HTTP-Schicht: Routing, Middleware, Handler

### Routing

Routen werden in dedizierten Dateien pro Domain registriert:

```go
func RegisterRoutes(mux *http.ServeMux, cmd *CommandService, qry *QueryService) {
    mux.Handle("POST /api/orders/create",
        middleware.Chain(
            NewCommandHandler(cmd).CreateOrder,
            middleware.JWT(jwtSecret, "admin", "manager"),
            middleware.RateLimit(10),
        ))
    mux.Handle("POST /api/orders/list",
        middleware.Chain(
            NewQueryHandler(qry).ListOrders,
            middleware.JWT(jwtSecret, "admin", "manager", "staff"),
        ))
}
```

### Middleware-Pattern (Chain of Responsibility)

Middleware ist eine Funktion, die einen `http.Handler` nimmt und einen `http.Handler` zurückgibt:

```go
type Middleware func(http.Handler) http.Handler

func Chain(h http.HandlerFunc, middlewares ...Middleware) http.Handler {
    result := http.Handler(h)
    for i := len(middlewares) - 1; i >= 0; i-- {
        result = middlewares[i](result)
    }
    return result
}
```

Ausführungsreihenfolge (von außen nach innen):

```
Request → RateLimit → CorrelationID → Logging → JWT → Handler → Response
```

| Middleware                | Funktion                                           |
| ------------------------- | -------------------------------------------------- |
| `RateLimitMiddleware`     | Token-Bucket per IP                                |
| `CorrelationIDMiddleware` | Setzt `X-Correlation-ID` Header (UUID)             |
| `LoggingMiddleware`       | Loggt Path, Status, Duration, Correlation-ID       |
| `JWTMiddleware`           | Validiert JWT, extrahiert UserID + Role in Context |
| `RecoveryMiddleware`      | Fängt Panics, gibt 500 zurück                      |

> Für Authentifizierungs-Middleware, JWT-Validierung und Security-Patterns siehe [Security & Authentifizierung](security.md).

### Handler-Pattern

```go
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
    // 1. Request parsen (JSON → Struct)
    var req CreateOrderRequest
    if err := helper.ParseJSON(r, &req); err != nil {
        helper.WriteError(w, http.StatusBadRequest, "invalid_json")
        return
    }

    // 2. UserID aus Context (gesetzt von JWT-Middleware)
    userID, _ := r.Context().Value(middleware.UserIDKey).(int)

    // 3. Application Service aufrufen
    if err := h.service.CreateOrder(r.Context(), userID, req); err != nil {
        helper.HandleDomainError(w, err)
        return
    }

    // 4. Erfolgsresponse
    helper.WriteJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
}
```

### Request/Response-Format

```json
// Erfolg
{"status": "created", "id": 42}

// Fehler
{"code": "order_not_found"}
{"code": "validation_error", "details": {"items": ["min 1 item required"]}}
```

---

## 6. Application Services: Commands und Queries

### Command Service

Verantwortlich für **schreibende Operationen**. Orchestriert Domain-Logik und Repositories:

```go
type CommandService struct {
    productRepo ProductRepository
    orderRepo   OrderRepository
}

func (s *CommandService) CreateOrder(ctx context.Context, userID int, req CreateOrderRequest) error {
    // 1. Produkte laden und validieren
    products, err := s.productRepo.FindByIDs(ctx, req.ProductIDs)
    if err != nil {
        return fmt.Errorf("load products: %w", err)
    }

    // 2. Domain-Logik: Order erstellen (Geschäftsregeln prüfen)
    order, err := domain.NewOrder(userID, products, req.Comment)
    if err != nil {
        return err
    }

    // 3. Persistieren
    return s.orderRepo.Save(ctx, order)
}
```

### Query Service

Verantwortlich für **lesende Operationen**. Liest aus dem Read Model oder rekonstruiert Zustand:

```go
type QueryService struct {
    orderRepo OrderReadRepository
}

func (s *QueryService) GetOrderSummary(ctx context.Context, orderID int) (OrderSummaryDTO, error) {
    return s.orderRepo.FindSummary(ctx, orderID)
}
```

### Factory-Pattern für Services

Services werden über Factory-Funktionen erstellt, die alle Dependencies injizieren:

```go
func NewCommandHandler(db *pgxpool.Pool) *CommandHandler {
    productRepo := product_repo.NewRepository(db)
    orderRepo   := order_repo.NewRepository(db)
    service     := NewCommandService(productRepo, orderRepo)
    return &CommandHandler{service: service}
}
```

---

## 7. Domain-Schicht

### Grundregeln

- **Keine DB-Abhängigkeiten** — kein `import` von `pgx`, `sqlc`, etc.
- **Keine HTTP-Abhängigkeiten** — kein `import` von `net/http`
- **Reine Geschäftslogik** — Validierung, Berechnung, Event-Erzeugung
- **Testbar ohne Infrastruktur** — Unit-Tests ohne DB oder Server

### Entities

```go
type Order struct {
    ID        int
    CustomerID int
    Status    OrderStatus
    Items     []OrderItem
    CreatedAt time.Time
}

type OrderStatus string

const (
    OrderStatusPending   OrderStatus = "pending"
    OrderStatusConfirmed OrderStatus = "confirmed"
    OrderStatusCancelled OrderStatus = "cancelled"
)
```

### Domain Services (reine Funktionen)

Zustandsrekonstruktion und komplexe Berechnungen als zustandslose Funktionen:

```go
// Berechnung aus Events rekonstruieren (Event-Sourcing)
func CalculateBalance(events []Event) int { ... }
func GetOpenItems(events []Event) []Item  { ... }
func GetHistory(events []Event) []Entry   { ... }

// Domänen-Validierung
func NewOrder(customerID int, items []Item) (Order, error) {
    if len(items) == 0 {
        return Order{}, ErrOrderMustHaveItems
    }
    return Order{CustomerID: customerID, Items: items, Status: StatusPending}, nil
}
```

**Vorteile reiner Funktionen:** Deterministisch, keine Seiteneffekte, einfach zu testen, parallelisierbar.

---

## 8. Fehlerbehandlung in Go

### 8.1 Go-Philosophie: Fehler als Werte

Go verwendet **explizite Fehlerbehandlung** (kein try/catch). Das `error`-Interface ist denkbar einfach:

```go
type error interface {
    Error() string
}
```

Jeder Fehler wird entweder behandelt, weitergereicht oder geloggt:

```go
result, err := doSomething()
if err != nil {
    // Option 1: Behandeln
    return defaultValue, nil
    // Option 2: Weiterreichen (mit Kontext)
    return nil, fmt.Errorf("doSomething failed: %w", err)
    // Option 3: Loggen + HTTP-Fehler
    log.Error().Err(err).Msg("unexpected error")
    http.Error(w, "internal server error", 500)
}
```

### 8.2 Sentinel Errors

Vordefnierte Fehlerwerte, die mit `==` verglichen werden können:

```go
// Definieren (Konvention: Err-Prefix, package-level)
var ErrNotFound      = errors.New("not found")
var ErrAlreadyExists = errors.New("already exists")
var ErrUnauthorized  = errors.New("unauthorized")

// Prüfen
if err == ErrNotFound {
    http.Error(w, "resource not found", 404)
}
```

**Problem:** Sentinel Errors können keine zusätzlichen Informationen tragen und können nicht geWrapped werden. Seit Go 1.13 bevorzugt: `errors.Is()`.

### 8.3 Custom Error Types

Structs, die das `error`-Interface implementieren und zusätzliche Felder tragen:

```go
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed on %s: %s", e.Field, e.Message)
}

// Erstellen
return &ValidationError{Field: "email", Message: "invalid format"}

// Prüfen mit errors.As
var valErr *ValidationError
if errors.As(err, &valErr) {
    fmt.Println("Field:", valErr.Field)
}
```

### 8.4 Error Wrapping (Go 1.13+)

`fmt.Errorf` mit `%w`-Verb wraps den Fehler und erhält die ursprüngliche Fehler-Identität:

```go
// Fehler wrapen (mit Kontext)
return fmt.Errorf("order service CreateOrder: %w", err)

// Fehler-Kette aufbauen
func LoadUser(id int) error {
    _, err := db.Query("SELECT ...")
    if err != nil {
        return fmt.Errorf("LoadUser id=%d: %w", id, err)
    }
    return nil
}

// Fehler prüfen (unwrapped die gesamte Kette)
if errors.Is(err, pgx.ErrNoRows) { ... }  // findet auch in wrapped errors

// Bestimmten Typ extrahieren
var pgErr *pgconn.PgError
if errors.As(err, &pgErr) {
    fmt.Println("PG Code:", pgErr.Code)
}
```

### 8.5 Domain Errors vs. Infrastructure Errors

```
PostgreSQL Error (pgconn.PgError)
     ↓ Repository: mapError()
Domain Error (ErrNotFound, ErrAlreadyExists)
     ↓ Application Service: return fmt.Errorf("...: %w", err)
     ↓ HTTP Handler: HandleDomainError()
HTTP Response: {"code": "order_not_found"}
```

```go
// Repository: mapError übersetzt Infra → Domain
func mapError(err error) error {
    if err == nil { return nil }
    if errors.Is(err, pgx.ErrNoRows) { return ErrNotFound }

    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) {
        switch pgErr.Code {
        case "23505": return ErrAlreadyExists   // Unique Violation
        case "23503": return ErrForeignKey       // FK Violation
        }
    }
    return ErrDatabase
}
```

### 8.6 Fehler-Mapping auf HTTP-Status

| Domain Error      | HTTP-Status | Code                    |
| ----------------- | ----------- | ----------------------- |
| Validierung       | 400         | `validation_error`      |
| Nicht gefunden    | 404         | `*_not_found`           |
| Bereits vorhanden | 409         | `*_already_exists`      |
| Nicht autorisiert | 401         | `unauthorized`          |
| Kein Zugriff      | 403         | `forbidden`             |
| Rate Limit        | 429         | `rate_limit_exceeded`   |
| Server-Fehler     | 500         | `internal_server_error` |

---

## 9. Validierung

### Zwei-Schichten-Validierung

```
Client (Zod/Yup) → HTTP (JSON Parsing) → Domain (Schema) → Database (Constraints)
```

| Schicht      | Tool                          | Prüft                                   |
| ------------ | ----------------------------- | --------------------------------------- |
| Frontend     | Zod / Yup / Valibot           | UI-Validierung, sofortiges Feedback     |
| HTTP-Handler | JSON-Decoder                  | JSON-Syntax, bekannte Felder            |
| Domain       | zog / go-playground/validator | Geschäftsregeln (Min/Max, OneOf)        |
| Database     | PostgreSQL Constraints        | Eindeutigkeit, Fremdschlüssel, NOT NULL |

### Striktes JSON-Parsing

```go
func ParseJSON(r *http.Request, dst any) error {
    dec := json.NewDecoder(r.Body)
    dec.DisallowUnknownFields()  // Unbekannte Felder → Fehler
    if err := dec.Decode(dst); err != nil {
        return fmt.Errorf("parse json: %w", err)
    }
    return nil
}
```

---

## 10. Go Concurrency Patterns

### 10.1 goroutines und channels

Go's Concurrency-Modell basiert auf goroutines (leichtgewichtige Threads, ~2KB Stack) und channels (typisierte Kommunikationskanäle):

```go
// Goroutine starten
go func() {
    result := processItem(item)
    resultChan <- result
}()

// Channel-Kommunikation
jobs := make(chan Job, 100)      // buffered channel
results := make(chan Result, 100)

go worker(jobs, results)         // Consumer
jobs <- Job{ID: 1}               // Producer
result := <-results              // Empfangen
```

### 10.2 Worker Pool

Kontrollierte Parallelverarbeitung mit fixem Pool von goroutines:

```go
func WorkerPool(numWorkers int, jobs <-chan Job, results chan<- Result) {
    var wg sync.WaitGroup
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for job := range jobs {
                results <- processJob(job)
            }
        }()
    }
    go func() {
        wg.Wait()
        close(results)
    }()
}

// Nutzung
jobs    := make(chan Job, 100)
results := make(chan Result, 100)
WorkerPool(5, jobs, results)   // 5 parallele Worker
```

### 10.3 Fan-out / Fan-in

**Fan-out:** Eine Quelle, viele Consumer.
**Fan-in:** Viele Quellen, eine Senke.

```go
// Fan-out: Job an mehrere Worker verteilen
func fanOut(input <-chan Item, workers int) []<-chan Result {
    outputs := make([]<-chan Result, workers)
    for i := 0; i < workers; i++ {
        outputs[i] = process(input)
    }
    return outputs
}

// Fan-in: Mehrere Channels in einen zusammenführen
func fanIn(channels ...<-chan Result) <-chan Result {
    merged := make(chan Result)
    var wg sync.WaitGroup
    for _, ch := range channels {
        wg.Add(1)
        go func(c <-chan Result) {
            defer wg.Done()
            for r := range c { merged <- r }
        }(ch)
    }
    go func() { wg.Wait(); close(merged) }()
    return merged
}
```

### 10.4 errgroup für parallele Operationen

`golang.org/x/sync/errgroup` führt parallele Operationen aus und sammelt Fehler:

```go
import "golang.org/x/sync/errgroup"

func loadOrderDetails(ctx context.Context, orderID int) (*OrderDetails, error) {
    g, ctx := errgroup.WithContext(ctx)

    var order Order
    var payments []Payment
    var items []Item

    g.Go(func() error {
        var err error
        order, err = orderRepo.FindByID(ctx, orderID)
        return err
    })
    g.Go(func() error {
        var err error
        payments, err = paymentRepo.FindByOrder(ctx, orderID)
        return err
    })
    g.Go(func() error {
        var err error
        items, err = itemRepo.FindByOrder(ctx, orderID)
        return err
    })

    if err := g.Wait(); err != nil {
        return nil, err
    }
    return &OrderDetails{Order: order, Payments: payments, Items: items}, nil
}
```

### 10.5 Graceful Shutdown

```go
func main() {
    srv := &http.Server{Addr: ":8080", Handler: mux}

    // Server in goroutine starten
    go func() {
        if err := srv.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatal().Err(err).Msg("server error")
        }
    }()

    // OS-Signal abwarten (SIGINT, SIGTERM)
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    // Graceful Shutdown: laufende Requests abschließen (max. 30s)
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    if err := srv.Shutdown(ctx); err != nil {
        log.Error().Err(err).Msg("shutdown error")
    }
    pool.Close()
    log.Info().Msg("server stopped")
}
```

---

## 11. Configuration & 12-Factor App

### 11.1 12-Factor App (Auszug für Go)

Die [12-Factor App](https://12factor.net/) ist eine Methodologie für portable, cloud-native Services. Relevante Faktoren für Go-Backends:

| Faktor                | Prinzip                                          | Go-Umsetzung                           |
| --------------------- | ------------------------------------------------ | -------------------------------------- |
| **III. Config**       | Config aus Umgebungsvariablen, nie im Code       | `os.Getenv`, envconfig, Viper          |
| **VI. Processes**     | Zustandslose, horizontalskalierende Prozesse     | Kein In-Memory-State zwischen Requests |
| **IX. Disposability** | Schneller Start, Graceful Shutdown               | `context.WithTimeout`, `signal.Notify` |
| **XI. Logs**          | Logs als Eventstream (stdout), keine Log-Dateien | `zerolog`, `slog`, Ausgabe auf stdout  |

### 11.2 Configuration Pattern

```go
type Config struct {
    // Pflichtfelder
    DatabaseURL string `env:"DATABASE_URL,required"`
    JWTSecret   string `env:"JWT_SECRET,required"`

    // Optionale Felder mit Defaults
    Port        int    `env:"PORT" envDefault:"8080"`
    LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`
    Environment string `env:"ENVIRONMENT" envDefault:"development"`
}

func Load() (*Config, error) {
    cfg := &Config{}
    if err := envconfig.Process("", cfg); err != nil {
        return nil, fmt.Errorf("load config: %w", err)
    }
    return cfg, nil
}
```

### 11.3 Config-Libraries im Vergleich

| Library       | Ansatz                           | Stärken                                   |
| ------------- | -------------------------------- | ----------------------------------------- |
| **envconfig** | Struct-Tags, nur Env-Vars        | Einfach, keine externe Dependency         |
| **Viper**     | Env-Vars + Config-Files + Flags  | Flexibel, Hierarchie, Hot Reload          |
| **koanf**     | Modular (Env, YAML, JSON, Vault) | Composable, Typsicher, kein `interface{}` |

---

## 12. Resilienz-Patterns

### 12.1 Circuit Breaker

Verhindert, dass ein fehlerhafter Downstream-Service das gesamte System destabilisiert:

```
Zustand: CLOSED → (Fehlerrate > Threshold) → OPEN
                                                ↓
                     (Timeout) → HALF-OPEN → (Testanfrage ok) → CLOSED
                                           → (Testanfrage fehlgeschlagen) → OPEN
```

```go
// Beispiel mit github.com/sony/gobreaker
cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
    Name:        "payment-service",
    MaxRequests: 3,               // Anfragen in HALF-OPEN
    Interval:    10 * time.Second,
    Timeout:     60 * time.Second,
    ReadyToTrip: func(counts gobreaker.Counts) bool {
        return counts.ConsecutiveFailures > 5
    },
})

result, err := cb.Execute(func() (interface{}, error) {
    return paymentClient.Charge(ctx, amount)
})
```

### 12.2 Retry mit Exponential Backoff

```go
func retryWithBackoff(ctx context.Context, maxRetries int, fn func() error) error {
    backoff := 100 * time.Millisecond
    for attempt := 0; attempt <= maxRetries; attempt++ {
        if err := fn(); err == nil {
            return nil
        } else if attempt == maxRetries {
            return fmt.Errorf("max retries exceeded: %w", err)
        }
        // Jitter hinzufügen, um Thundering Herd zu vermeiden
        jitter := time.Duration(rand.Intn(50)) * time.Millisecond
        select {
        case <-time.After(backoff + jitter):
            backoff *= 2
        case <-ctx.Done():
            return ctx.Err()
        }
    }
    return nil
}
```

### 12.3 Timeout

```go
// Context-basiertes Timeout
ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
defer cancel()

result, err := externalService.Call(ctx, request)
if errors.Is(err, context.DeadlineExceeded) {
    return nil, ErrServiceTimeout
}
```

### 12.4 Bulkhead

Isoliert Ressourcen (z.B. Connection Pools, goroutines) für verschiedene Workload-Typen, sodass ein überladener Teil das System nicht vollständig blockiert:

```go
// Separate Semaphoren pro Service-Typ
criticalSem := make(chan struct{}, 50)  // max 50 parallele critical Requests
reportingSem := make(chan struct{}, 5)  // max 5 parallele Report-Requests

func handleCritical(ctx context.Context, fn func() error) error {
    select {
    case criticalSem <- struct{}{}:
        defer func() { <-criticalSem }()
        return fn()
    case <-ctx.Done():
        return ErrServiceBusy
    }
}
```

### 12.5 Health Checks

```go
mux.HandleFunc("GET /health/live", func(w http.ResponseWriter, r *http.Request) {
    // Liveness: Ist der Prozess noch am Leben?
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
})

mux.HandleFunc("GET /health/ready", func(w http.ResponseWriter, r *http.Request) {
    // Readiness: Kann der Service Traffic verarbeiten?
    if err := db.Ping(r.Context()); err != nil {
        w.WriteHeader(http.StatusServiceUnavailable)
        json.NewEncoder(w).Encode(map[string]string{"status": "not ready", "reason": "db unreachable"})
        return
    }
    json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
})
```

---

## 13. Observability

### 13.1 Structured Logging

Structured Logging gibt Log-Einträge als maschinenlesbare Key-Value-Paare aus (JSON), nicht als freie Strings:

```go
// zerolog (performant, zero-allocation)
log.Info().
    Str("path", r.URL.Path).
    Int("status", statusCode).
    Dur("duration", duration).
    Str("correlation_id", correlationID).
    Int("user_id", userID).
    Msg("request completed")

// slog (Go 1.21 stdlib)
slog.Info("request completed",
    "path", r.URL.Path,
    "status", statusCode,
    "duration_ms", duration.Milliseconds(),
)
```

**Library-Vergleich:**

| Library     | Allokationen | API-Stil       | Empfehlen für                     |
| ----------- | ------------ | -------------- | --------------------------------- |
| **zerolog** | Null         | Fluent Chain   | Performance-kritische Dienste     |
| **zap**     | Minimal      | Strongly typed | Hochlastsysteme                   |
| **slog**    | Niedrig      | stdlib, stabil | Neue Projekte (kein extra Import) |
| **logrus**  | Moderat      | Classic        | Legacy-Projekte                   |

### 13.2 Distributed Tracing mit OpenTelemetry

OpenTelemetry ist der Standard für herstellerunabhängiges Distributed Tracing:

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("order-service")

func (s *CommandService) CreateOrder(ctx context.Context, req CreateOrderRequest) error {
    ctx, span := tracer.Start(ctx, "CreateOrder")
    defer span.End()

    span.SetAttributes(
        attribute.Int("user_id", req.UserID),
        attribute.Int("product_count", len(req.Items)),
    )

    if err := s.orderRepo.Save(ctx, order); err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return err
    }
    return nil
}
```

**Exporters:** Jaeger, Zipkin, Tempo (Grafana), AWS X-Ray, Google Cloud Trace.

### 13.3 Metrics mit Prometheus

```go
import "github.com/prometheus/client_golang/prometheus"

var (
    httpRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total HTTP requests",
    }, []string{"method", "path", "status"})

    httpRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "http_request_duration_seconds",
        Help:    "HTTP request duration",
        Buckets: prometheus.DefBuckets,
    }, []string{"method", "path"})
)

// In Middleware
httpRequestsTotal.WithLabelValues(method, path, strconv.Itoa(status)).Inc()
httpRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())

// Metrics-Endpoint
mux.Handle("GET /metrics", promhttp.Handler())
```

---

## 14. Datenbankzugriff / SQL-Tooling

### 14.1 SQL-Tooling Landschaft in Go

| Tool     | Ansatz                    | Code-Generierung | Typsicherheit | Lernkurve | Flexibilität |
| -------- | ------------------------- | ---------------- | ------------- | --------- | ------------ |
| **sqlc** | SQL-first, Code-Gen       | Ja (aus SQL)     | Sehr hoch     | Niedrig   | Sehr hoch    |
| **sqlx** | SQL-first, kein Code-Gen  | Nein             | Mittel        | Niedrig   | Sehr hoch    |
| **GORM** | ORM, Convention-over-Conf | Nein             | Niedrig       | Mittel    | Mittel       |
| **ent**  | Graph-Schema, Code-Gen    | Ja (aus Schema)  | Sehr hoch     | Hoch      | Hoch         |
| **Jet**  | SQL-Builder, Code-Gen     | Ja (aus DB)      | Sehr hoch     | Mittel    | Hoch         |

### 14.2 sqlc Deep Dive

**sqlc** generiert typsichere Go-Code aus SQL-Queries. Workflow:

```
SQL-Queries (*.sql) + Schema → sqlc generate → typsichere Go-Structs + Funktionen
```

```sql
-- sqlc/queries/orders.sql

-- name: CreateOrder :one
INSERT INTO orders (customer_id, status, created_at)
VALUES ($1, $2, NOW())
RETURNING *;

-- name: FindOrderByID :one
SELECT * FROM orders WHERE id = $1;

-- name: ListOrdersByCustomer :many
SELECT * FROM orders WHERE customer_id = $1 ORDER BY created_at DESC;

-- name: UpdateOrderStatus :exec
UPDATE orders SET status = $2 WHERE id = $1;
```

```go
// Generierter Code (NICHT manuell bearbeiten)
func (q *Queries) CreateOrder(ctx context.Context, arg CreateOrderParams) (Order, error)
func (q *Queries) FindOrderByID(ctx context.Context, id int64) (Order, error)
func (q *Queries) ListOrdersByCustomer(ctx context.Context, customerID int64) ([]Order, error)
```

**Batch Operations:**

```sql
-- name: BatchInsertItems :batchexec
INSERT INTO order_items (order_id, product_id, quantity, price_cents)
VALUES ($1, $2, $3, $4);
```

**Einschränkungen:** Keine dynamischen WHERE-Klauseln (Workaround: mehrere Queries oder sqlx), kein automatisches Schema-Inferencing für komplexe Joins.

### 14.3 ORM vs. SQL-first vs. Query Builder

| Ansatz                     | Wann verwenden                                    | Wann vermeiden                              |
| -------------------------- | ------------------------------------------------- | ------------------------------------------- |
| **ORM (GORM)**             | Rapid Prototyping, CRUD-lastige Apps              | Performance-kritische Queries, komplexe SQL |
| **SQL-first (sqlc, sqlx)** | Wenn SQL-Kontrolle wichtig ist, explizite Queries | Wenn Datenbankschema noch sehr fluide       |
| **Query Builder (Jet)**    | Dynamische Queries mit Typsicherheit              | Wenn einfache statische SQL reicht          |

> **Go-Empfehlung:** SQL-first-Ansatz (sqlc oder sqlx) ist idiomatischer Go — er vermeidet "N+1 durch ORM"-Überraschungen und ist besser inspizierbar.

### 14.4 Repository Pattern

Das Repository Pattern entkoppelt Domain-Logik von DB-Details über Interfaces:

```go
// Interface (Domain-Schicht oder Application-Schicht)
type OrderRepository interface {
    Save(ctx context.Context, order Order) error
    FindByID(ctx context.Context, id int) (Order, error)
    FindByCustomer(ctx context.Context, customerID int) ([]Order, error)
    UpdateStatus(ctx context.Context, id int, status OrderStatus) error
}

// Implementierung (Repository-Schicht, wraps sqlc)
type orderRepository struct {
    q *dbgen.Queries
}

func NewOrderRepository(db *pgxpool.Pool) OrderRepository {
    return &orderRepository{q: dbgen.New(db)}
}

func (r *orderRepository) FindByID(ctx context.Context, id int) (Order, error) {
    row, err := r.q.FindOrderByID(ctx, int64(id))
    if err != nil {
        return Order{}, mapError(err)
    }
    return toOrder(row), nil
}
```

**Mock für Tests:**

```go
type mockOrderRepository struct {
    orders map[int]Order
}

func (m *mockOrderRepository) FindByID(ctx context.Context, id int) (Order, error) {
    if order, ok := m.orders[id]; ok {
        return order, nil
    }
    return Order{}, ErrNotFound
}
```

### 14.5 Migration-Tooling

| Tool                   | Ansatz                  | Besonderheiten                                 |
| ---------------------- | ----------------------- | ---------------------------------------------- |
| **golang-migrate**     | SQL-Dateien (up/down)   | Einfach, viele Driver, kein Schema-Inferencing |
| **Atlas**              | Deklarativ (HCL/SQL)    | Schema-Diff, CI-Integration, Cloud-Service     |
| **goose**              | SQL oder Go-Migrations  | Flexibel, Go-basierte Migrationen möglich      |
| **Flyway / Liquibase** | SQL-Dateien (JVM-Tools) | Ausgereift, aber JVM-Dependency                |

**Zero-Downtime Migration (Expand/Contract Pattern):**

```sql
-- Phase 1: Expand — neues Feld hinzufügen (nullable)
ALTER TABLE orders ADD COLUMN shipping_address TEXT;

-- Phase 2: Backfill — alten Code deployen, der beide Felder schreibt
-- (Deployments dazwischen)

-- Phase 3: Contract — altes Feld entfernen, wenn kein Code mehr darauf zugreift
ALTER TABLE orders DROP COLUMN old_address;
```

---

## 15. Testing in Go

### 15.1 Test-Pyramide im Go-Kontext

```
         ┌─────────────┐
         │   E2E Tests  │  Wenige, langsam, teuer
         ├─────────────┤
         │Integration  │  Mittel (Testcontainers, httptest)
         ├─────────────┤
         │ Unit Tests  │  Viele, schnell, günstig
         └─────────────┘
```

Build-Tags trennen Test-Typen:

```go
//go:build unit        // Nur bei: go test -tags=unit
//go:build integration // Nur bei: go test -tags=integration
```

### 15.2 Table-Driven Tests

Das idiomatischste Go-Testing-Pattern — ein Test-Case-Slice mit benannten Feldern:

```go
func TestCalculateDiscount(t *testing.T) {
    tests := []struct {
        name            string
        orderTotalCents int
        customerTier    string
        wantDiscount    int
        wantErr         bool
    }{
        {"bronze no discount",   5000, "bronze",   0,   false},
        {"silver 5% discount",  10000, "silver",   500, false},
        {"gold 10% discount",   20000, "gold",    2000, false},
        {"unknown tier → error", 5000, "unknown",   0,   true},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            got, err := CalculateDiscount(tc.orderTotalCents, tc.customerTier)
            if tc.wantErr {
                if err == nil {
                    t.Fatal("expected error, got nil")
                }
                return
            }
            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }
            if got != tc.wantDiscount {
                t.Errorf("got %d, want %d", got, tc.wantDiscount)
            }
        })
    }
}
```

**Vorteile:** Alle Test-Cases auf einen Blick, leicht erweiterbar, parallele Ausführung (`t.Parallel()`).

### 15.3 httptest: HTTP-Handler testen

```go
func TestCreateOrderHandler(t *testing.T) {
    // Arrange
    mockService := &mockOrderService{
        createOrderFn: func(ctx context.Context, req CreateOrderRequest) error { return nil },
    }
    handler := NewCommandHandler(mockService)

    body := `{"customer_id": 1, "items": [{"product_id": 2, "quantity": 3}]}`
    req := httptest.NewRequest(http.MethodPost, "/orders/create", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    // Act
    handler.CreateOrder(w, req)

    // Assert
    resp := w.Result()
    if resp.StatusCode != http.StatusCreated {
        t.Errorf("expected 201, got %d", resp.StatusCode)
    }
}
```

### 15.4 Testcontainers: Integration Tests mit echter DB

```go
//go:build integration

func TestOrderRepository_Save(t *testing.T) {
    ctx := context.Background()

    // PostgreSQL-Container starten
    container, err := postgres.Run(ctx,
        "postgres:16",
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
        testcontainers.WithWaitStrategy(
            wait.ForLog("database system is ready to accept connections"),
        ),
    )
    require.NoError(t, err)
    defer container.Terminate(ctx)

    // Migrations anwenden
    connStr, _ := container.ConnectionString(ctx, "sslmode=disable")
    runMigrations(t, connStr)

    // Repository testen
    pool, _ := pgxpool.New(ctx, connStr)
    repo := NewOrderRepository(pool)

    order := Order{CustomerID: 1, Status: StatusPending}
    err = repo.Save(ctx, order)
    require.NoError(t, err)
    require.NotZero(t, order.ID)
}
```

### 15.5 Mocking-Strategien

**Interface-basierte Mocks (empfohlen in Go):**

```go
// Interface definieren
type EmailSender interface {
    SendConfirmation(ctx context.Context, to string, orderID int) error
}

// Mock-Implementierung
type mockEmailSender struct {
    sentEmails []sentEmail
    returnErr  error
}

func (m *mockEmailSender) SendConfirmation(_ context.Context, to string, orderID int) error {
    m.sentEmails = append(m.sentEmails, sentEmail{to: to, orderID: orderID})
    return m.returnErr
}

// In Test verwenden
emailer := &mockEmailSender{}
service := NewOrderService(repo, emailer)
```

**Mocking mit testify/mock:**

```go
type MockOrderRepo struct {
    mock.Mock
}

func (m *MockOrderRepo) FindByID(ctx context.Context, id int) (Order, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(Order), args.Error(1)
}

// In Test
mockRepo := new(MockOrderRepo)
mockRepo.On("FindByID", mock.Anything, 42).Return(Order{ID: 42}, nil)
```

### 15.6 Golden Files (Snapshot Testing)

Für komplexe Outputs (JSON-APIs, Report-Generierung) kann der erwartete Output als Datei abgelegt werden:

```go
func TestGenerateReport(t *testing.T) {
    result := GenerateReport(testData)
    got, _ := json.MarshalIndent(result, "", "  ")

    goldenFile := "testdata/report.golden.json"
    if *update {  // go test -update Flag
        os.WriteFile(goldenFile, got, 0644)
        return
    }
    want, _ := os.ReadFile(goldenFile)
    if !bytes.Equal(got, want) {
        t.Errorf("output mismatch. run with -update to regenerate")
    }
}
```

### 15.7 Coverage und Benchmarks

```bash
# Coverage-Report
go test -tags=unit -cover ./...
go test -tags=unit -coverprofile=coverage.out ./...
go tool cover -html=coverage.out  # HTML-Visualisierung

# Benchmark
go test -bench=. -benchmem ./...
```

```go
func BenchmarkCalculateDiscount(b *testing.B) {
    for i := 0; i < b.N; i++ {
        CalculateDiscount(10000, "gold")
    }
}
```

---

## 16. Referenzen

### Architektur-Patterns

- [Alistair Cockburn: Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/) — Ports & Adapters Originalquelle
- [Robert C. Martin: Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html) — Clean Architecture Blog Post
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout) — Community-Konventionen für Verzeichnisstruktur (kein offizieller Go-Standard; siehe auch [Organizing a Go module](https://go.dev/doc/modules/layout))
- [Event-Driven Architecture in Golang](https://github.com/PacktPublishing/Event-Driven-Architecture-in-Golang) — DDD + ES + CQRS in Go

### Go-Spezifisch

- [Effective Go](https://go.dev/doc/effective_go) — Offizieller Style Guide
- [Go Proverbs (Rob Pike)](https://go-proverbs.github.io/) — Go Design-Philosophie
- [Go Wiki: Code Review Comments](https://go.dev/wiki/CodeReviewComments) — Idiomatisches Go
- [Go Blog: Error Handling](https://go.dev/blog/error-handling-and-go) — Offizielle Error-Handling-Patterns
- [Go Blog: Errors Are Values](https://go.dev/blog/errors-are-values) — Fehlerbehandlung als Design-Prinzip
- [Alex Edwards: Let's Go Further](https://lets-go-further.alexedwards.net/) — Go Web Application Patterns, Middleware, APIs
- [GopherCon Talks](https://www.youtube.com/c/GopherAcademy) — Go Architecture Talks

### API-Design & Resilienz

- [12-Factor App](https://12factor.net/) — Configuration, Logging, Concurrency
- [Microsoft: Resiliency Patterns](https://learn.microsoft.com/en-us/azure/architecture/patterns/category/resiliency) — Circuit Breaker, Retry, Bulkhead

### Datenbankzugriff

- [sqlc Dokumentation](https://docs.sqlc.dev/) — Offizielle sqlc-Referenz, Query Patterns
- [pgx v5 Dokumentation](https://pkg.go.dev/github.com/jackc/pgx/v5) — Go PostgreSQL Driver, Connection Pooling
- [ent (Go ORM)](https://entgo.io/) — Graph-basiertes Go ORM, Schema-as-Code
- [Atlas (Schema Migration)](https://atlasgo.io/) — Deklarative Schema-Migrationen für Go
- [DB Performance 101](https://dev.to/ari-ghosh/db-performance-101-a-practical-deep-dive-into-backend-database-optimization-4cag) — N+1, Connection Pooling, Query Optimization

### Testing

- [Dave Cheney: Table-Driven Tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests) — Table-Driven Tests, Go-Testing-Philosophie
- [testcontainers-go](https://golang.testcontainers.org/) — Integrationstests mit echten Containern
- [Mitchell Hashimoto: Testing in Go](https://www.youtube.com/watch?v=8hQG7QlcLBk) — Advanced Go Testing Patterns (GopherCon)
