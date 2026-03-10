# System Design — jotti

Dieses Dokument beschreibt das vollständige System Design von jotti: ein Self-hosted, Open-Source, Mobile-first Gastronomie-Kassensystem für Vereine und Non-Profit-Veranstaltungen. Es ist die zentrale Referenz für Architektur-, Domain- und Implementierungsentscheidungen.

> **Verwandte Dokumente:**
>
> - [AGENTS.md](../AGENTS.md) — Regeln für KI-Coding-Agenten
> - [Ubiquitous Language](language.md) — Kanonische Fachbegriffe
> - [Event Storming](event-storming.md) — Domain Events, Aggregate, Bounded Contexts
> - [Anforderungskatalog](requirements.md) — Funktionale Anforderungen mit Status
> - [Implementierungsplan](implementation-plan.md) — Nächste Features
> - [ADR: Event-Sourcing](adr/event-sourcing.md) — Entscheidung für Event-Sourcing
> - [ADR: sqlc](adr/orm.md) — Entscheidung für sqlc
> - [CQRS in jotti](cqrs.md) — CQRS Ist-Zustand und Projektionsplan

---

## Inhaltsverzeichnis

1. [Vision und Systemkontext](#1-vision-und-systemkontext)
2. [Bounded Contexts und Domain Map](#2-bounded-contexts-und-domain-map)
3. [Ubiquitous Language](#3-ubiquitous-language)
4. [Domänenmodell](#4-domänenmodell)
5. [Software-Architektur](#5-software-architektur)
6. [Modular Monolith](#6-modular-monolith)
7. [Persistenzstrategie: Hybrides Event-Sourcing + CRUD](#7-persistenzstrategie-hybrides-event-sourcing--crud)
8. [CQRS-Architektur](#8-cqrs-architektur)
9. [Datenbankmodell](#9-datenbankmodell)
10. [API-Design](#10-api-design)
11. [Authentifizierung und Autorisierung](#11-authentifizierung-und-autorisierung)
12. [Frontend-Architektur](#12-frontend-architektur)
13. [Querschnittsthemen](#13-querschnittsthemen)
14. [Deployment-Architektur](#14-deployment-architektur)
15. [Entscheidungsprotokoll](#15-entscheidungsprotokoll)

---

## 1. Vision und Systemkontext

### 1.1 Vision

jotti ermöglicht Vereinen und Non-Profit-Organisationen, Veranstaltungen (Vereinsfeste, Weihnachtsmärkte, Konzerte, Maihocks) mit einem einfachen, selbst gehosteten Kassensystem zu betreiben. Servicekräfte nehmen auf ihren eigenen Smartphones Bestellungen auf, liefern aus, kassieren und stornieren — alles pro Tisch.

### 1.2 Abgrenzung

jotti ist **kein** kommerzielles Gastro-POS:

| Eigenschaft       | jotti                         | Kommerzielle POS            |
| ----------------- | ----------------------------- | --------------------------- |
| Hardware-Bindung  | Keine — BYOD (Smartphone)     | Dedizierte Terminals        |
| Hosting           | Self-hosted (Docker)          | Cloud-SaaS                  |
| Zahlungsgateway   | Keines — nur Bargeld-Tracking | Kartenleser, Online-Payment |
| Lizenz            | Open Source (AGPL-3.0)        | Kostenpflichtig             |
| Zielgruppe        | Ehrenamtliche, Vereinsfeste   | Gastronomie-Betriebe        |
| Fiskalkonformität | Nicht erforderlich            | TSE, GoBD etc.              |

### 1.3 Systemkontextdiagramm (C4 Level 1)

```
                    ┌──────────────────────┐
                    │    Servicekraft       │
                    │  (Smartphone/Browser) │
                    └──────────┬───────────┘
                               │ HTTPS
                               ▼
                    ┌──────────────────────┐
                    │       jotti          │
                    │  (Kassensystem)      │
                    └──────────┬───────────┘
                               │
              ┌────────────────┼────────────────┐
              ▼                ▼                ▼
     ┌──────────────┐ ┌──────────────┐ ┌──────────────────┐
     │  Admin        │ │  Servicekraft│ │  Serviceleitung  │
     │  (Browser)    │ │  (Smartphone)│ │  (Smartphone)    │
     └──────────────┘ └──────────────┘ └──────────────────┘
```

### 1.4 Akteure und Rollen

| Akteur             | Rolle            | Berechtigungen                                                    |
| ------------------ | ---------------- | ----------------------------------------------------------------- |
| **Admin**          | `admin`          | Voller Zugriff: Stammdaten + Kassenbetrieb + Stornierung          |
| **Serviceleitung** | `senior_service` | Kassenbetrieb + Stornierung                                       |
| **Servicekraft**   | `service`        | Kassenbetrieb (bestellen, liefern, kassieren) — keine Stornierung |

---

## 2. Bounded Contexts und Domain Map

### 2.1 Strategisches Design

jotti gliedert sich in zwei Bounded Contexts und eine technische Infrastrukturschicht:

```
┌──────────────────────────────────────────────────────────────────────────┐
│                            jotti (System)                                │
│                                                                          │
│  ┌───────────────────────────────┐   ┌────────────────────────────────┐  │
│  │  Kassenbetrieb (Core Domain)  │   │  Stammdaten (Supporting)       │  │
│  │  Bounded Context: service     │   │  Bounded Context: admin        │  │
│  │                               │   │                                │  │
│  │  Tisch (Aggregat)             │   │  Produkt + Varianten (CRUD)    │  │
│  │  Bestellung                   │   │  Tisch-Stammdaten (CRUD)       │  │
│  │  Zahlung                      │   │  Benutzer (CRUD)               │  │
│  │  Lieferung                    │   │                                │  │
│  │  Stornierung                  │   │  Persistenz: CRUD / SQL        │  │
│  │  Saldo                        │   │                                │  │
│  │                               │   │                                │  │
│  │  Persistenz: Event-Sourcing   │   │                                │  │
│  └───────────────┬───────────────┘   └──────────────┬─────────────────┘  │
│                  │                                   │                    │
│                  │ liest Produkte/Tische              │                    │
│                  │◄──────────────────────────────────┘                    │
│                                                                          │
│  ┌───────────────────────────────────────────────────────────────────┐   │
│  │  Auth (Infrastruktur — kein Bounded Context)                      │   │
│  │  Login, Passwort setzen, JWT-Ausstellung, Middleware               │   │
│  └───────────────────────────────────────────────────────────────────┘   │
└──────────────────────────────────────────────────────────────────────────┘
```

### 2.2 Sub-Domain-Klassifikation

| Sub-Domain        | Typ         | Begründung                                                         | Persistenz     |
| ----------------- | ----------- | ------------------------------------------------------------------ | -------------- |
| **Kassenbetrieb** | Core Domain | Differenzierendes Kerngeschäft: Event-Sourcing, Saldo, Audit-Trail | Event-Sourcing |
| **Stammdaten**    | Supporting  | Notwendig, aber standardisiertes CRUD ohne komplexe Logik          | CRUD           |
| **Auth**          | Generic     | Technische Infrastruktur, austauschbar                             | CRUD           |

### 2.3 Context Mapping

```
  Kassenbetrieb (service)        Stammdaten (admin)
  ┌────────────────────┐         ┌────────────────────┐
  │                    │ U ← D   │                    │
  │  Tisch-Aggregat    │◄────────│  Produkt-Katalog   │
  │  (Event-Sourced)   │         │  Tisch-Stammdaten  │
  │                    │         │  (CRUD)            │
  └────────────────────┘         └────────────────────┘
        Upstream (U)                  Downstream (D)
```

**Beziehungstyp: Conformist (vereinfacht)**

Der Kassenbetrieb liest Stammdaten (Produkte, Tische) direkt aus den CRUD-Tabellen. Es gibt keine Anti-Corruption-Layer — bei einem Monolith mit gemeinsamer Datenbank ist das pragmatisch vertretbar. Die Events im Kassenbetrieb enthalten **eingebettete Momentaufnahmen** der Stammdaten (Fat Events: Produktname, Preis zum Bestellzeitpunkt), sodass historische Korrektheit unabhängig von späteren Stammdaten-Änderungen gewährleistet ist.

---

## 3. Ubiquitous Language

### 3.1 Sprachregeln

| Regel                                    | Beispiel                                              |
| ---------------------------------------- | ----------------------------------------------------- |
| Fachbegriffe der Domäne sind **deutsch** | `Bestellung`, `Zahlung`, `Tisch`, `Stornierung`       |
| Infrastruktur-Code bleibt **englisch**   | `Auth`, `Config`, `middleware`, `repository`          |
| UI-Texte sind **deutsch**                | „Bestellung aufgeben", „Zahlung registrieren"         |
| Commits sind **englisch**                | `feat: add table search`, `fix: correct balance calc` |
| DB-Tabellennamen bleiben **englisch**    | `users`, `tables`, `products`, `events`               |

### 3.2 Kanonisches Glossar

#### Kassenbetrieb

| Fachbegriff       | Code-Name                   | Definition                                                                                     |
| ----------------- | --------------------------- | ---------------------------------------------------------------------------------------------- |
| **Tisch**         | `Tisch`                     | Abrechnungseinheit und Aggregatwurzel. Träger des Kassenjournals.                              |
| **Bestellung**    | `Bestellung`                | Auftrag mit Positionen, aufgegeben von einer Servicekraft an einem Tisch.                      |
| **Position**      | `Position`                  | Einzelposten: Varianten-ID + Name + Preis (in Cent) + Menge.                                   |
| **Lieferung**     | `Lieferung`                 | Bestätigung, dass Positionen physisch ausgeliefert wurden.                                     |
| **Zahlung**       | `Zahlung`                   | Registrierung eines Zahlungseingangs. Teilzahlung möglich.                                     |
| **Stornierung**   | `Stornierung`               | Rücknahme von Positionen. Nur Serviceleitung/Admin.                                            |
| **Saldo**         | `saldoCents`                | Offener Betrag: $\sum \text{Bestellungen} - \sum \text{Zahlungen} - \sum \text{Stornierungen}$ |
| **Kassenjournal** | Event Stream (`tisch:<id>`) | Unveränderliche Folge aller Tisch-Events.                                                      |

#### Stammdaten

| Fachbegriff   | Code-Name              | Definition                                                          |
| ------------- | ---------------------- | ------------------------------------------------------------------- |
| **Produkt**   | `Produkt` / `product`  | Angebotener Artikel (z. B. „Cola", „Pommes").                       |
| **Variante**  | `Variante` / `variant` | Ausprägung eines Produkts mit eigenem Preis (z. B. „0,3l", „0,5l"). |
| **Kategorie** | `category`             | Gruppierung: `food`, `beverage`, `other`.                           |
| **Benutzer**  | `user`                 | Person mit Zugangsdaten und Rolle.                                  |

#### Rollen

| Fachbegriff        | Code-Rolle       | Berechtigung                      |
| ------------------ | ---------------- | --------------------------------- |
| **Admin**          | `admin`          | Alles: Stammdaten + Kassenbetrieb |
| **Serviceleitung** | `senior_service` | Kassenbetrieb + Stornierung       |
| **Servicekraft**   | `service`        | Kassenbetrieb ohne Stornierung    |

---

## 4. Domänenmodell

### 4.1 Aggregate

#### Aggregat: Tisch (Core Domain — Event-Sourced)

Das `Tisch`-Aggregat ist die zentrale Geschäftseinheit von jotti. Es kapselt den gesamten Kassenbetrieb an einem physischen Tisch.

```
┌──────────────────────────────────────────────────────────────┐
│  Aggregat: Tisch                                              │
│  Aggregate Root: Tisch (identifiziert durch "tisch:<id>")     │
│                                                               │
│  ┌──────────────┐                                             │
│  │   Tisch      │                                             │
│  │   id: int    │                                             │
│  │   name       │                                             │
│  │   status     │                                             │
│  └──────┬───────┘                                             │
│         │ Event Stream (append-only)                          │
│         ▼                                                     │
│  ┌──────────────────────────────────────────┐                 │
│  │  Kassenjournal (Events)                   │                │
│  │                                           │                │
│  │  BestellungAufgegeben ──┐                 │                │
│  │  ZahlungRegistriert   ──┤ → Zustand       │                │
│  │  ProdukteGeliefert    ──┤   rekonstruiert  │                │
│  │  ProdukteStorniert    ──┘   aus Events     │                │
│  │                                           │                │
│  │  Rekonstruierter Zustand:                 │                │
│  │  ├─ Saldo (Cents)                         │                │
│  │  ├─ Unbezahlte Positionen                 │                │
│  │  ├─ Ungelieferte Positionen               │                │
│  │  └─ Historie (chronologisch)              │                │
│  └──────────────────────────────────────────┘                 │
└──────────────────────────────────────────────────────────────┘
```

**Invarianten:**

| #   | Invariante                                                                   |
| --- | ---------------------------------------------------------------------------- |
| I1  | Saldo darf nach Stornierung nicht negativ werden                             |
| I2  | Stornierung nur für Positionen, die bestellt aber nicht bereits bezahlt sind |
| I3  | Lieferung nur für Positionen, die bestellt wurden                            |
| I4  | Zahlung bezieht sich auf bestellte, unbezahlte Positionen                    |
| I5  | Stornierung nur durch Rollen `admin` oder `senior_service`                   |

**Domain Events:**

| Event-Typ                        | Auslöser               | Daten (Fat Event)                                        |
| -------------------------------- | ---------------------- | -------------------------------------------------------- |
| `tisch.bestellung-aufgegeben:v1` | Servicekraft           | Positionen[], Kommentar, UserID, TischID                 |
| `tisch.zahlung-registriert:v1`   | Servicekraft           | Positionen[], Kommentar, UserID, TischID                 |
| `tisch.produkte-geliefert:v1`    | Servicekraft           | Positionen[], Kommentar, UserID, TischID                 |
| `tisch.produkte-storniert:v1`    | Serviceleitung / Admin | Positionen[], Kommentar, UserID, TischID                 |
| `tisch.snapshot:v1`              | System                 | Saldo, UnbezahltePos., UngeliefertePos., GesamtZahlungen |

**State-Rekonstruktion:**

```
Saldo            = Σ(Bestellungen.GesamtPreis) − Σ(Zahlungen.GesamtZahlung) − Σ(Stornierungen.GesamtStornierung)
UnbezahltePosn.  = accumulate(Bestellungen.Positionen) − reduce(Zahlungen.Positionen) − reduce(Stornierungen.Positionen)
UngeliefertePosn.= accumulate(Bestellungen.Positionen) − reduce(Lieferungen.Positionen) − reduce(Stornierungen.Positionen)
```

### 4.2 Domain-Modelle (Go-Structs)

```go
// === Tisch-Aggregat (Event-Sourced) ===

type Position struct {
    ID         int    // Varianten-ID
    Name       string // Produktname + Variantenname (eingefroren)
    PreisCents int    // Preis zum Bestellzeitpunkt (eingefroren)
    Quantity   int
}

type Bestellung struct {
    ID               string
    UserID           int
    TischID          int
    Positionen       []Position
    GesamtPreisCents int
    Comment          string
    AufgegebenAm     time.Time
}

type Zahlung struct {
    ID                 string
    UserID             int
    TischID            int
    Positionen         []Position
    GesamtZahlungCents int
    Comment            string
    RegistriertAm      time.Time
}

type Lieferung struct {
    ID          string
    UserID      int
    TischID     int
    Positionen  []Position
    Comment     string
    GeliefertAm time.Time
}

type Stornierung struct {
    ID                     string
    UserID                 int
    TischID                int
    Positionen             []Position
    GesamtStornierungCents int
    Comment                string
    StorniertAm            time.Time
}

type Snapshot struct {
    TischID                int
    SaldoCents             int
    UnbezahltePositionen   []Position
    UngeliefertePositionen []Position
    GesamtZahlungenCents   int
    CreatedAt              time.Time
}
```

```go
// === Stammdaten (CRUD) ===

type Product struct {
    ID        int
    Name      string
    Category  Category  // "food" | "beverage" | "other"
    Variants  []Variant
    CreatedAt time.Time
}

type Variant struct {
    ID         int
    Name       string
    PriceCents int
    Status     Status  // "active" | "inactive"
    CreatedAt  time.Time
}

type Tisch struct {
    ID        int
    Name      string
    Status    Status  // "active" | "inactive"
    CreatedAt time.Time
}

type User struct {
    ID                  int
    Name                string
    Username            string
    Role                Role    // "admin" | "senior_service" | "service"
    Status              Status  // "active" | "inactive"
    PasswordHash        string  // Argon2id
    OnetimePasswordHash string
    CreatedAt           time.Time
}
```

### 4.3 Value Objects

| Value Object | Felder                                   | Invarianten                  |
| ------------ | ---------------------------------------- | ---------------------------- |
| `Position`   | ID, Name, PreisCents, Quantity           | PreisCents ≥ 0, Quantity > 0 |
| `Category`   | "food" \| "beverage" \| "other"          | Einer der drei Werte         |
| `Role`       | "admin" \| "senior_service" \| "service" | Einer der drei Werte         |
| `Status`     | "active" \| "inactive"                   | Einer der zwei Werte         |

### 4.4 Domain Services

Freie Funktionen in `domain/table/events.go`, die Zustand aus dem Event-Stream rekonstruieren:

| Funktion                              | Eingabe   | Ausgabe       | Beschreibung                     |
| ------------------------------------- | --------- | ------------- | -------------------------------- |
| `GetSaldoFromEvents`                  | `[]Event` | `int` (Cents) | Berechnet offenen Saldo          |
| `GetHistoryFromEvents`                | `[]Event` | `[]any`       | Chronologische Tisch-Historie    |
| `GetUnbezahltePositionenFromEvents`   | `[]Event` | `[]Position`  | Noch nicht bezahlte Positionen   |
| `GetUngeliefertePositionenFromEvents` | `[]Event` | `[]Position`  | Noch nicht gelieferte Positionen |

Hilfsfunktionen:

| Funktion               | Beschreibung                                                      |
| ---------------------- | ----------------------------------------------------------------- |
| `accumulatePositionen` | Addiert Positionen zu einer Liste (gleiche IDs → Quantity +)      |
| `reducePositionen`     | Subtrahiert Positionen von einer Liste (gleiche IDs → Quantity −) |

---

## 5. Software-Architektur

### 5.1 Architekturstil: Modular Monolith

jotti ist ein **Modular Monolith** — ein einziger Deployable mit klar getrennten Modulen. Die Module entsprechen den Bounded Contexts und sind durch definierte Schnittstellen entkoppelt.

**Warum kein Microservice?**

- Vereinsfest-Betrieb: geringe Last, kleines Team, einfaches Deployment
- Ein Docker-Compose-Stack genügt
- Keine operationale Komplexität durch verteilte Systeme
- Shared PostgreSQL-Instanz ohne Netzwerk-Overhead

### 5.2 Schichtenarchitektur (pro Modul)

Jedes Modul folgt einer einheitlichen 4-Schichten-Architektur:

```
┌─────────────────────────────────────────────────┐
│                HTTP-Handler                      │
│  api/<domain>/http/                              │
│  Parst HTTP-Requests, delegiert an Application   │
├─────────────────────────────────────────────────┤
│                Application Service               │
│  api/<domain>/application/                       │
│  Orchestriert Use Cases, Command/Query-Trennung  │
├─────────────────────────────────────────────────┤
│                Domain Model                      │
│  domain/<domain>/                                │
│  Geschäftslogik, Validierung, Event-Factories    │
├─────────────────────────────────────────────────┤
│                Repository                        │
│  repository/<domain>_repo/                       │
│  Datenbankzugriff via sqlc-generiertem Code      │
└─────────────────────────────────────────────────┘
```

**Abhängigkeitsregel (Dependency Rule):**

```
HTTP → Application → Domain ← Repository
                        ↑
                 Keine Abhängigkeit
                 zu äußeren Schichten
```

- `domain/` hat **keine** Imports von `api/`, `repository/` oder `db/`
- `repository/` importiert `domain/` (für Modell-Konvertierung) und `db/` (für Fehler-Mapping)
- `application/` importiert `domain/` und definiert Repository-Interfaces (Dependency Inversion)
- `http/` importiert `application/`

### 5.3 Dependency Injection

jotti verwendet **Constructor Injection** über das `app`-Package:

```go
// app/app.go — Dependency Wiring
type App struct {
    DB     *sql.DB
    Config config.Config
    // ... Repository- und Service-Instanzen
}
```

Repositories werden einmalig beim Start erzeugt und in Application Services injiziert. Application Services werden in HTTP-Handler injiziert. Kein DI-Framework — nur Go-Konstruktoren.

---

## 6. Modular Monolith

### 6.1 Modulstruktur

```
backend/
├── api/                          # API-Gateway: Routes + Module
│   ├── service.go                # Service-Routes (Kassenbetrieb)
│   ├── senior_service.go         # Stornierung (erhöhte Berechtigung)
│   ├── admin.go                  # Admin-Routes (Stammdaten)
│   ├── auth.go                   # Auth-Routes (Login, Passwort)
│   │
│   ├── table/                    # ─── Modul: Tisch (Kassenbetrieb) ───
│   │   ├── http/
│   │   │   ├── command_handler.go    # POST: Bestellung, Zahlung, Lieferung, Stornierung
│   │   │   └── query_handler.go      # POST: Saldo, Historie, Unbezahlt, Ungeliefert
│   │   └── application/
│   │       ├── command.go            # Command-Service (schreibt Events)
│   │       └── query.go             # Query-Service (liest Events → Zustand)
│   │
│   ├── product/                  # ─── Modul: Produkt (Stammdaten) ───
│   │   ├── http/
│   │   │   ├── command_handler.go
│   │   │   └── query_handler.go
│   │   └── application/
│   │       ├── command.go
│   │       └── query.go
│   │
│   ├── user/                     # ─── Modul: Benutzer (Stammdaten) ───
│   │   ├── http/
│   │   │   ├── command_handler.go
│   │   │   └── query_handler.go
│   │   └── application/
│   │       ├── command.go
│   │       └── query.go
│   │
│   ├── auth/                     # ─── Modul: Auth (Infrastruktur) ───
│   │   ├── http/handler.go
│   │   └── application/service.go
│   │
│   ├── health/                   # ─── Health Check ───
│   │   └── health.go
│   │
│   ├── middleware/                # ─── Querschnitt: Middleware ───
│   │   └── middleware.go             # JWT-Auth, Rate-Limiting, Logging, CORS
│   │
│   └── helper/                   # ─── Querschnitt: HTTP-Hilfsfunktionen ───
│       └── http.go                   # JSON-Parsing, Error-Responses
│
├── domain/                       # ─── Domain Layer (pure Geschäftslogik) ───
│   ├── event/event.go                # Generisches Event-Modell
│   ├── jwt/jwt.go                    # JWT-Erzeugung/Validierung
│   ├── product/                      # Produkt + Variante
│   │   ├── product.go
│   │   └── variant.go
│   ├── table/                        # Tisch-Aggregat + Events + State-Rekonstruktion
│   │   ├── tisch.go
│   │   ├── bestellung.go
│   │   ├── zahlung.go
│   │   ├── lieferung.go
│   │   ├── stornierung.go
│   │   ├── events.go                 # State-Rekonstruktion
│   │   ├── bestellungAufgegebenEvent.go
│   │   ├── zahlungRegistriertEvent.go
│   │   ├── produkteGeliefertEvent.go
│   │   ├── produkteStorniertEvent.go
│   │   └── snapshotEvent.go
│   └── user/
│       ├── user.go
│       └── password.go               # Argon2id-Hashing
│
├── repository/                   # ─── Persistence Layer ───
│   ├── event_repo/repo.go           # Append-only Event Store
│   ├── product_repo/repo.go         # CRUD für Produkte + Varianten
│   ├── table_repo/repo.go           # CRUD für Tisch-Stammdaten
│   └── user_repo/repo.go            # CRUD für Benutzer
│
├── sqlc/                         # ─── SQL-Queries + generierter Code ───
│   ├── queries/
│   │   ├── events.sql
│   │   ├── products.sql
│   │   ├── tables.sql
│   │   └── users.sql
│   └── dbgen/                        # Generiert (NICHT EDITIEREN)
│
├── config/config.go              # Umgebungsvariablen
├── db/                           # DB-Verbindung, Fehler-Mapping
├── app/app.go                    # Dependency Wiring
└── main.go                       # Einstiegspunkt
```

### 6.2 Modulabhängigkeiten

```
         ┌──────────────────────────────────────────────────┐
         │                    main.go                        │
         │              (Startet Server,                     │
         │               verdrahtet Module)                  │
         └──────────────────────┬───────────────────────────┘
                                │
                                ▼
         ┌──────────────────────────────────────────────────┐
         │              app/app.go                           │
         │         (Dependency Wiring)                       │
         └───┬────────────┬────────────┬───────────────────┘
             │            │            │
             ▼            ▼            ▼
    ┌────────────┐ ┌────────────┐ ┌────────────┐
    │  table/    │ │  product/  │ │  user/     │    ← API-Module
    │  (CQRS)   │ │  (CQRS)    │ │  (CQRS)    │
    └─────┬──────┘ └─────┬──────┘ └─────┬──────┘
          │              │              │
          ▼              ▼              ▼
    ┌────────────┐ ┌────────────┐ ┌────────────┐
    │ domain/    │ │ domain/    │ │ domain/    │    ← Domain Layer
    │ table/     │ │ product/   │ │ user/      │
    └─────┬──────┘ └────────────┘ └────────────┘
          │
          ▼
    ┌────────────┐
    │ domain/    │    ← Shared Domain
    │ event/     │
    └────────────┘
```

**Regel: Module kommunizieren nicht direkt miteinander.** Der `table/`-Modul liest keine Daten aus `product/`. Wenn das `table/`-Modul Produktdaten benötigt (z. B. zur Validierung), werden diese über den Application Service aus dem Repository geladen — nicht über den `product/`-Application-Service.

### 6.3 Inter-Modul-Kommunikation

| Von           | Nach                 | Mechanismus                                    |
| ------------- | -------------------- | ---------------------------------------------- |
| table-Command | event_repo           | Direkter Repository-Zugriff (Write Event)      |
| table-Query   | event_repo           | Direkter Repository-Zugriff (Read Events)      |
| table-Command | table_repo           | Direkter Repository-Zugriff (Tisch validieren) |
| service.go    | table, product       | Route-Registration (HTTP-Mux)                  |
| admin.go      | table, product, user | Route-Registration (HTTP-Mux)                  |

---

## 7. Persistenzstrategie: Hybrides Event-Sourcing + CRUD

### 7.1 Entscheidung

| Datenkategorie    | Strategie      | Begründung                            |
| ----------------- | -------------- | ------------------------------------- |
| Tisch-Operationen | Event-Sourcing | Audit-Trail, Immutabilität, Zeitreise |
| Stammdaten        | CRUD           | Einfachheit, referenzielle Integrität |
| Auth-Daten        | CRUD           | Standard-Pattern, kein Audit nötig    |

Vollständige Begründung: [ADR: Event-Sourcing](adr/event-sourcing.md)

### 7.2 Event Store

Die `events`-Tabelle ist ein generischer, Append-only Event Store in PostgreSQL:

```sql
CREATE TABLE events (
    id        INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    user_id   INT REFERENCES users(id) NOT NULL,
    type      TEXT NOT NULL,
    subject   TEXT NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    data      JSONB NOT NULL
);
```

**Immutabilitäts-Garantie** (Defense-in-Depth):

1. `REVOKE ALL ... GRANT SELECT, INSERT` — Privilege-Einschränkung
2. `BEFORE UPDATE` Trigger → Exception
3. `BEFORE DELETE` Trigger → Exception
4. `BEFORE TRUNCATE` Trigger → Exception

### 7.3 Event-Design-Prinzipien

| Prinzip            | Umsetzung                                                                    |
| ------------------ | ---------------------------------------------------------------------------- |
| **Fat Events**     | Alle relevanten Daten eingebettet (Name, Preis zum Zeitpunkt der Bestellung) |
| **Versioniert**    | Event-Typen mit `:v1`-Suffix für Schema-Evolution                            |
| **Idempotent**     | Jedes Event hat eine eindeutige ID (autoincrement)                           |
| **Self-contained** | Kein Nachladen von Stammdaten bei Replay nötig                               |

### 7.4 Event-Typen

| Event-Typ                        | Subject      | Datenstruktur                                            |
| -------------------------------- | ------------ | -------------------------------------------------------- |
| `tisch.bestellung-aufgegeben:v1` | `tisch:<id>` | `{positionen, gesamtPreisCents, comment, tischID}`       |
| `tisch.zahlung-registriert:v1`   | `tisch:<id>` | `{positionen, gesamtZahlungCents, comment, tischID}`     |
| `tisch.produkte-geliefert:v1`    | `tisch:<id>` | `{positionen, comment, tischID}`                         |
| `tisch.produkte-storniert:v1`    | `tisch:<id>` | `{positionen, gesamtStornierungCents, comment, tischID}` |
| `tisch.snapshot:v1`              | `tisch:<id>` | `{saldo, unbezahlt, ungeliefert, gesamtZahlungen}`       |

### 7.5 Snapshot-Mechanismus

Snapshots optimieren die Leseperformance, indem sie den rekonstruierten Zustand als Event persistieren. Beim Lesen wird ab dem letzten Snapshot replayed statt ab dem ersten Event.

```
Events:  [E1] [E2] [E3] [SNAP] [E4] [E5]
                          ↑
                     Replay startet hier
```

**Geplante Evolution:** Der Snapshot-Mechanismus soll durch synchrone CQRS-Projektionen ersetzt werden (siehe [Abschnitt 8](#8-cqrs-architektur)).

---

## 8. CQRS-Architektur

### 8.1 Aktueller Zustand: CQRS Stufe 1 (Logische Trennung)

jotti implementiert CQRS bereits auf der logischen Ebene:

```
┌─────────────────────────────────────────────────────────────────┐
│  Application Layer (api/table/application/)                      │
│                                                                  │
│  ┌──────────────────────┐    ┌───────────────────────────────┐  │
│  │  Command Struct       │    │  Query Struct                  │  │
│  │                       │    │                                │  │
│  │  BestellungAufgeben() │    │  GetTischSaldo()               │  │
│  │  ZahlungRegistrieren()│    │  GetTischHistorie()            │  │
│  │  ProdukteLiefern()    │    │  GetTischUnbezahlt()           │  │
│  │  ProdukteStornieren() │    │  GetTischUngeliefert()         │  │
│  │  SnapshotErstellen()  │    │                                │  │
│  └──────────┬───────────┘    └────────────┬───────────────────┘  │
│             │                             │                      │
│             ▼                             ▼                      │
│  ┌──────────────────────┐    ┌───────────────────────────────┐  │
│  │  eventRepoCommand    │    │  eventRepoQuery                │  │
│  │  (Interface)          │    │  (Interface)                   │  │
│  │  WriteEvent()         │    │  ReadEventsBySubject()         │  │
│  └──────────────────────┘    │  ReadEventsWithSnapshot()      │  │
│                               └───────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

**Was vorhanden ist:**

- ✅ Separate `Command`/`Query`-Structs im Application Layer
- ✅ Separate HTTP-Handler (`command_handler.go` / `query_handler.go`)
- ✅ Interface Segregation (getrennte Repository-Interfaces für Command und Query)

**Was fehlt:**

- ❌ Kein separates Read Model — Queries lesen weiterhin Events und rekonstruieren in Go
- ❌ Keine event-getriebenen Projektionen

### 8.2 Ziel: CQRS Stufe 2 (Synchrone Projektionen)

Der geplante Zielzustand führt eine `table_state`-Projektionstabelle ein, die synchron beim Schreiben jedes Events aktualisiert wird:

```
┌─────────────────────────────────────────────────────────────────┐
│  Command Side                         Query Side                 │
│                                                                  │
│  Command Handler                     Query Handler               │
│       │                                   │                      │
│       ▼                                   ▼                      │
│  ┌──────────┐                       ┌──────────────┐            │
│  │ events   │  ── synchrone ──►     │ table_state  │            │
│  │ (append  │     Projektion        │ (Read Model) │            │
│  │  only)   │     (in derselben     │              │            │
│  └──────────┘      Transaktion)     └──────────────┘            │
│                                                                  │
│  Write Store                         Read Store                  │
│  (Event-Sourcing)                    (vorberechneter Zustand)    │
└─────────────────────────────────────────────────────────────────┘
```

**Vorteile der synchronen Projektion:**

- Queries werden zu simplen `SELECT`-Statements
- Snapshot-Mechanismus wird obsolet
- Starke Konsistenz (kein Eventual Consistency)
- Typisierte Spalten statt JSONB-Parsing auf der Leseseite

Vollständige Analyse und Implementierungsplan: [CQRS in jotti](cqrs.md)

---

## 9. Datenbankmodell

### 9.1 Übersicht

```
┌─────────────────────────────────────────────────────────────┐
│                       PostgreSQL                             │
│                                                              │
│  ┌─────────────────────┐    ┌─────────────────────────────┐ │
│  │  CRUD-Tabellen       │    │  Event Store                │ │
│  │                      │    │                              │ │
│  │  users              │    │  events (append-only)        │ │
│  │  tables             │    │    ├── id (PK)               │ │
│  │  products           │    │    ├── user_id (FK → users)  │ │
│  │  product_variants   │    │    ├── type                  │ │
│  │                      │    │    ├── subject              │ │
│  │  (Soft-Deletes via   │    │    ├── timestamp            │ │
│  │   status = 'deleted')│    │    └── data (JSONB)         │ │
│  └─────────────────────┘    └─────────────────────────────┘ │
│                                                              │
│  Enums: UserRole, EntityStatus, ProductCategory              │
└─────────────────────────────────────────────────────────────┘
```

### 9.2 Entity-Relationship-Diagramm

```
┌──────────────┐       ┌──────────────────┐       ┌──────────────┐
│    users     │       │   product_variants│       │   products   │
├──────────────┤       ├──────────────────┤       ├──────────────┤
│ id       PK  │       │ id           PK  │       │ id       PK  │
│ name         │       │ product_id   FK──┼──────►│ name         │
│ username  UQ │       │ name             │       │ category     │
│ password_hash│       │ price_cents      │       │ created_at   │
│ onetime_pw   │       │ status           │       └──────────────┘
│ role         │       │ created_at       │
│ status       │       └──────────────────┘
│ created_at   │
└──────┬───────┘
       │
       │ FK (user_id)
       ▼
┌──────────────┐       ┌──────────────┐
│    events    │       │    tables    │
├──────────────┤       ├──────────────┤
│ id       PK  │       │ id       PK  │
│ user_id  FK  │       │ name     UQ  │
│ type         │       │ status       │
│ subject      │       │ created_at   │
│ timestamp    │       └──────────────┘
│ data  (JSONB)│
└──────────────┘
  (append-only)
```

### 9.3 Enums

| Enum              | Werte                                | Verwendet in                                               |
| ----------------- | ------------------------------------ | ---------------------------------------------------------- |
| `UserRole`        | `admin`, `senior_service`, `service` | `users.role`                                               |
| `EntityStatus`    | `active`, `inactive`, `deleted`      | `users.status`, `tables.status`, `product_variants.status` |
| `ProductCategory` | `food`, `beverage`, `other`          | `products.category`                                        |

### 9.4 Indizes

| Index                         | Tabelle            | Spalte(n)         | Zweck                       |
| ----------------------------- | ------------------ | ----------------- | --------------------------- |
| `idx_users_username`          | `users`            | `username`        | Login-Lookup                |
| `idx_users_status`            | `users`            | `status`          | Aktive Benutzer filtern     |
| `idx_tables_status`           | `tables`           | `status`          | Aktive Tische filtern       |
| `idx_product_variants_status` | `product_variants` | `status`          | Aktive Varianten filtern    |
| `idx_events_user_id`          | `events`           | `user_id`         | Events nach Akteur          |
| `idx_events_subject`          | `events`           | `subject`         | Events pro Aggregat (Tisch) |
| `idx_events_type`             | `events`           | `type`            | Events nach Typ             |
| `idx_events_subject_type`     | `events`           | `subject`, `type` | Snapshot-Lookup pro Tisch   |

### 9.5 Soft-Delete-Strategie

Stammdaten werden nicht physisch gelöscht, sondern via `status = 'deleted'` markiert. Alle Queries filtern standardmäßig `WHERE status != 'deleted'`.

**Warum:** Tisch-Events referenzieren Benutzer-IDs. Physisches Löschen würde Fremdschlüssel-Constraints verletzen und die Audit-Trail-Nachvollziehbarkeit zerstören.

### 9.6 Migrationen

Verwaltet mit `golang-migrate`. Dateien unter `database/migrations/`:

```
01_initial.up.sql       — Schema, Enums, Tabellen, Trigger, Seed-User
01_initial.down.sql     — Rollback
02_add_senior_service_role.up.sql   — senior_service Enum-Wert
02_add_senior_service_role.down.sql — Rollback
```

---

## 10. API-Design

### 10.1 Grundprinzipien

| Prinzip                        | Umsetzung                                                   |
| ------------------------------ | ----------------------------------------------------------- |
| **POST-only**                  | Alle Endpoints sind POST. Keine GET/PUT/DELETE.             |
| **JSON-Request/Response**      | Content-Type: `application/json`                            |
| **Geldbeträge in Cent**        | `int`, niemals Float. `199` = 1,99 €                        |
| **Einheitliches Fehlerformat** | `{"code": "<string>", "details": "<optional>"}`             |
| **JWT-Bearer-Token**           | `Authorization: Bearer <token>` in allen geschützten Routen |
| **Validierung beidseitig**     | Backend: `zog`-Schemas. Frontend: `Zod`-Schemas.            |

### 10.2 API-Übersicht nach Bounded Context

#### Auth (`/auth/*`) — Kein JWT erforderlich

| Endpoint                  | Command/Query | Beschreibung                             |
| ------------------------- | ------------- | ---------------------------------------- |
| `POST /auth/login`        | Command       | Login mit Username + Passwort → JWT      |
| `POST /auth/set-password` | Command       | Einmalpasswort → eigenes Passwort setzen |

#### Service (`/service/*`) — Rollen: `admin`, `senior_service`, `service`

| Endpoint                              | Command/Query | Beschreibung                        |
| ------------------------------------- | ------------- | ----------------------------------- |
| `POST /service/bestellung-aufgeben`   | Command       | Bestellung am Tisch aufgeben        |
| `POST /service/zahlung-registrieren`  | Command       | Zahlung am Tisch registrieren       |
| `POST /service/produkte-liefern`      | Command       | Positionen als geliefert markieren  |
| `POST /service/get-aktive-produkte`   | Query         | Alle aktiven Produkte mit Varianten |
| `POST /service/get-tisch`             | Query         | Einzelnen Tisch abrufen             |
| `POST /service/get-aktive-tische`     | Query         | Alle aktiven Tische                 |
| `POST /service/get-tisch-historie`    | Query         | Chronologische Tisch-Historie       |
| `POST /service/get-tisch-saldo`       | Query         | Aktueller Saldo in Cent             |
| `POST /service/get-tisch-unbezahlt`   | Query         | Unbezahlte Positionen               |
| `POST /service/get-tisch-ungeliefert` | Query         | Ungelieferte Positionen             |

#### Senior Service (`/senior-service/*`) — Rollen: `admin`, `senior_service`

| Endpoint                                   | Command/Query | Beschreibung          |
| ------------------------------------------ | ------------- | --------------------- |
| `POST /senior-service/produkte-stornieren` | Command       | Positionen stornieren |

#### Admin (`/admin/*`) — Rolle: `admin`

| Endpoint                          | Command/Query | Beschreibung                    |
| --------------------------------- | ------------- | ------------------------------- |
| `POST /admin/create-user`         | Command       | Benutzer anlegen                |
| `POST /admin/update-user`         | Command       | Benutzer bearbeiten             |
| `POST /admin/activate-user`       | Command       | Benutzer aktivieren             |
| `POST /admin/deactivate-user`     | Command       | Benutzer deaktivieren           |
| `POST /admin/reset-password`      | Command       | Passwort zurücksetzen           |
| `POST /admin/get-all-users`       | Query         | Alle Benutzer (nicht gelöschte) |
| `POST /admin/create-produkt`      | Command       | Produkt erstellen               |
| `POST /admin/update-produkt`      | Command       | Produkt bearbeiten              |
| `POST /admin/create-variante`     | Command       | Variante erstellen              |
| `POST /admin/update-variante`     | Command       | Variante bearbeiten             |
| `POST /admin/activate-variante`   | Command       | Variante aktivieren             |
| `POST /admin/deactivate-variante` | Command       | Variante deaktivieren           |
| `POST /admin/get-all-produkte`    | Query         | Alle Produkte mit Varianten     |
| `POST /admin/create-tisch`        | Command       | Tisch erstellen                 |
| `POST /admin/update-tisch`        | Command       | Tisch umbenennen                |
| `POST /admin/activate-tisch`      | Command       | Tisch aktivieren                |
| `POST /admin/deactivate-tisch`    | Command       | Tisch deaktivieren              |
| `POST /admin/get-all-tische`      | Query         | Alle Tische (nicht gelöschte)   |

### 10.3 Fehlerformat

Alle Fehler-Responses folgen einem einheitlichen Schema:

```json
{
  "code": "tisch-not-found",
  "details": "Tisch mit ID 42 nicht gefunden"
}
```

HTTP-Statuscodes:

| Status | Bedeutung                 | Code-Beispiel             |
| ------ | ------------------------- | ------------------------- |
| 400    | Ungültige Eingabe         | `invalid-request`         |
| 401    | Nicht authentifiziert     | `unauthorized`            |
| 403    | Keine Berechtigung        | `forbidden`               |
| 404    | Ressource nicht gefunden  | `tisch-not-found`         |
| 409    | Konflikt (z. B. Duplikat) | `username-already-exists` |
| 429    | Rate-Limit überschritten  | `rate-limit-exceeded`     |
| 500    | Interner Fehler           | `internal-error`          |

---

## 11. Authentifizierung und Autorisierung

### 11.1 Authentifizierung: JWT

| Eigenschaft | Wert                                                    |
| ----------- | ------------------------------------------------------- |
| Algorithmus | HS256 (HMAC-SHA256)                                     |
| Gültigkeit  | 12 Stunden                                              |
| Issuer      | `jotti`                                                 |
| Claims      | `sub` (userID), `role` (admin\|senior_service\|service) |
| Speicherort | Frontend: `localStorage`                                |
| Transport   | `Authorization: Bearer <token>`                         |

### 11.2 Passwort-Handling

```
Admin erstellt Benutzer → Einmalpasswort generiert → Argon2id-Hash gespeichert
          ↓
Servicekraft loggt sich ein → Einmalpasswort eingeben → Weiterleitung zu "Passwort setzen"
          ↓
Eigenes Passwort setzen → Argon2id-Hash des neuen Passworts → Einmalpasswort gelöscht
          ↓
Zukünftige Logins → Passwort gegen gespeicherten Hash prüfen → JWT ausstellen
```

### 11.3 Autorisierung: Rollen-basierte Middleware

```go
// Middleware-Kette:
Request → JWT extrahieren → Claims validieren → Rolle prüfen → Handler aufrufen

// Route-Gruppen mit Rollen-Beschränkung:
/auth/*           → Keine Middleware (öffentlich)
/service/*        → JWT + Rollen: admin, senior_service, service
/senior-service/* → JWT + Rollen: admin, senior_service
/admin/*          → JWT + Rolle: admin
```

### 11.4 401-Interceptor (Frontend)

Der `BackendClient` im Frontend erkennt 401-Responses automatisch, löscht das gespeicherte Token und leitet zu `/login` weiter. Kein manuelles 401-Handling in einzelnen Komponenten nötig.

---

## 12. Frontend-Architektur

### 12.1 Tech-Stack

| Komponente    | Technologie                       |
| ------------- | --------------------------------- |
| Framework     | React 19 + TypeScript (strict)    |
| Build         | Vite                              |
| Styling       | Tailwind CSS 4                    |
| UI-Bibliothek | shadcn/ui (New York Style, Radix) |
| Icons         | Lucide React                      |
| Toasts        | Sonner                            |
| Drawers       | Vaul                              |
| Routing       | React Router (Data-Router)        |
| Validierung   | Zod                               |

### 12.2 Architekturprinzipien

| Prinzip                         | Umsetzung                                                             |
| ------------------------------- | --------------------------------------------------------------------- |
| Kein globaler State-Store       | Nur React Hooks + Singletons (`Auth`, `Backend`)                      |
| Backend-Klassen statt `fetch()` | Alle API-Aufrufe über [BackendClient](../frontend/src/lib/Backend.ts) |
| Backend ist Source of Truth     | Keine Frontend-Filterung — Backend liefert korrekte Daten             |
| Mobile-first                    | Touch-optimierte UI, Bottom-Sheet-Drawers                             |
| Feature-basierte Ordnerstruktur | Pro Feature: Schema + Backend-Klasse + Hook + Komponenten             |

### 12.3 Routing und Guards

```
/                       → Redirect zu /login
/login                  → LoginPage (AuthRedirect: leitet ein, wenn schon eingeloggt)
/set-password           → PasswordPage (AuthRedirect)

/admin                  → AdminLayout (AdminGuard: nur admin)
  /admin/products       → AdminProductsPage
  /admin/tables         → AdminTablesPage
  /admin/users          → AdminUsersPage

/service                → ServiceLayout (ServiceGuard: admin, senior_service, service)
  /service/tables       → TableSelectionPage
  /service/tables/:id   → TablePage
```

### 12.4 Daten-Fluss-Pattern

```
                  ┌─────────────────┐
                  │  React-Komponente│
                  └────────┬────────┘
                           │ ruft auf
                           ▼
                  ┌─────────────────┐
                  │  Custom Hook     │
                  │  useFetch<T>()  │
                  └────────┬────────┘
                           │ nutzt
                           ▼
                  ┌─────────────────┐
                  │ Backend-Klasse   │
                  │ (z.B. TischBackend)│
                  └────────┬────────┘
                           │ ruft auf
                           ▼
                  ┌─────────────────┐
                  │ BackendClient    │
                  │ .post()          │
                  │ (JSON + Auth)    │
                  └────────┬────────┘
                           │ HTTPS
                           ▼
                  ┌─────────────────┐
                  │  Backend API     │
                  └─────────────────┘
```

### 12.5 Drawer-Pattern (Kassenbetrieb-UI)

Bestellen, Bezahlen, Stornieren und Liefern öffnen Bottom-Sheet-Drawers:

1. Positionen auswählen (Checkbox + Mengenauswahl)
2. Zusammenfassung mit Gesamtbetrag anzeigen
3. Bestätigen → API-Aufruf → Toast bei Erfolg/Fehler

Hilfsfunktionen in `drawerUtils.ts`: `selectPositionen()`, `calculateTotalPrice()`

### 12.6 Frontend-Verzeichnisstruktur

```
frontend/src/
├── App.tsx                     # Root-Komponente
├── routes.ts                   # Alle Routen + Guards
├── main.tsx                    # Entry Point
├── index.css                   # CSS-Variablen, Theme
│
├── lib/                        # Shared Infrastructure
│   ├── Auth.ts                 # Auth-Singleton (Token, Rollen)
│   ├── AuthBackend.ts          # Login/SetPassword API
│   ├── Backend.ts              # BackendClient (POST, 401-Interceptor)
│   ├── useFetch.ts             # Custom Fetch-Hook
│   └── utils.ts                # formatCents(), cn()
│
├── admin/                      # Admin-Bereich
│   ├── AdminLayout.tsx
│   ├── products/               # Produkt-Verwaltung
│   ├── tables/                 # Tisch-Verwaltung
│   └── users/                  # Benutzer-Verwaltung
│
├── service/                    # Service-Bereich (Kassenbetrieb)
│   ├── ServiceLayout.tsx
│   ├── TableSelectionPage.tsx  # Tischauswahl
│   ├── TablePage.tsx           # Einzelner Tisch (Bestellen, Zahlen, ...)
│   ├── components/table/       # Drawer-Komponenten
│   ├── product/                # Produkt-Daten (Schema, Backend, Hook)
│   └── table/                  # Tisch-Daten (Schema, Backend, Hook)
│
├── pages/                      # Auth-Seiten
│   ├── LoginPage.tsx
│   └── PasswordPage.tsx
│
└── components/
    ├── ui/                     # shadcn/ui-Basiskomponenten
    └── common/                 # Gemeinsame Komponenten
```

---

## 13. Querschnittsthemen

### 13.1 Validierung (duale Validierung)

```
Frontend (Zod)                    Backend (zog)
┌──────────────┐                  ┌──────────────┐
│ Zod-Schema   │  ── HTTP ──►    │ zog-Schema   │
│ .parse()     │                  │ .Parse()     │
│ Client-side  │                  │ Server-side  │
│ UX-Feedback  │                  │ Source of    │
│              │                  │ Truth        │
└──────────────┘                  └──────────────┘
```

Beide Seiten validieren dieselben Invarianten. Das Backend ist die Single Source of Truth.

### 13.2 Fehler-Mapping (Backend)

```
PostgreSQL Error                    Application Error          HTTP Status
───────────────                    ──────────────────         ───────────
Unique Violation (23505)   →       db.ErrAlreadyExists   →   409 Conflict
sql.ErrNoRows              →       db.ErrNotFound        →   404 Not Found
Sonstiger Fehler           →       db.ErrDatabase        →   500 Internal
```

### 13.3 Logging

Strukturiertes Logging via `zerolog`:

- Request-Logging in Middleware (Method, Path, Status, Duration)
- Error-Logging bei 500er-Responses
- Event-Writes werden geloggt (Event-Typ, Subject, UserID)

### 13.4 Rate-Limiting

Rate-Limiting Middleware auf sensiblen Endpoints (`/auth/login`). Verhindert Brute-Force-Angriffe.

### 13.5 Geldbeträge

**Invariante: Alle Geldbeträge sind Ganzzahlen in Cent.**

| Schicht   | Darstellung                                |
| --------- | ------------------------------------------ |
| Datenbank | `INT` (`price_cents`)                      |
| Backend   | `int` (`PreisCents`, `GesamtZahlungCents`) |
| API       | `int` (JSON: `"preisCents": 199`)          |
| Frontend  | `number` → `formatCents(199)` → „1,99 €"   |

### 13.6 Zeitstempel

- Datenbank: `TIMESTAMPTZ` (UTC)
- Backend: `time.Time` (UTC)
- API: ISO 8601 String
- Frontend: Lokale Zeitzone für Anzeige

---

## 14. Deployment-Architektur

### 14.1 Containerisierung

```
┌──────────────────────────────────────────────────────────────┐
│                     Docker Compose Stack                      │
│                                                              │
│  ┌────────────────┐  ┌────────────────┐  ┌───────────────┐  │
│  │  nginx          │  │  backend       │  │  postgresql   │  │
│  │  Reverse Proxy  │  │  Go Binary     │  │  Database     │  │
│  │  :443 → :8080  │──│  :8080         │──│  :5432        │  │
│  │  TLS (Let's E.) │  │                │  │               │  │
│  └────────────────┘  └────────────────┘  └───────────────┘  │
│                                                              │
│  ┌────────────────┐  ┌────────────────┐                     │
│  │  frontend       │  │  migrate       │                     │
│  │  nginx + SPA   │  │  DB-Migrationen│                     │
│  │  :80 (intern)  │  │  (init)        │                     │
│  └────────────────┘  └────────────────┘                     │
└──────────────────────────────────────────────────────────────┘
```

### 14.2 Umgebungen

| Umgebung    | Datei                        | Zweck              |
| ----------- | ---------------------------- | ------------------ |
| Development | `docker-compose.dev.yml`     | Lokale Entwicklung |
| Staging     | `docker-compose.staging.yml` | Test mit TLS       |
| Production  | `docker-compose.yml`         | Live-Betrieb       |

### 14.3 Reverse Proxy (nginx)

- TLS-Terminierung (Let's Encrypt Zertifikate)
- `/api/*` → Backend-Container
- `/*` → Frontend-Container (SPA mit Fallback auf `index.html`)
- WebSocket-Support (falls in Zukunft nötig)

---

## 15. Entscheidungsprotokoll

### 15.1 Architekturentscheidungen (ADRs)

| Entscheidung                 | Status      | Dokument                                     |
| ---------------------------- | ----------- | -------------------------------------------- |
| Event-Sourcing für Tisch-Ops | Entschieden | [ADR: Event-Sourcing](adr/event-sourcing.md) |
| sqlc für Datenbankzugriff    | Entschieden | [ADR: sqlc](adr/orm.md)                      |
| CQRS Stufe 2 (Projektionen)  | Geplant     | [CQRS in jotti](cqrs.md)                     |

### 15.2 Designentscheidungen (dieses Dokument)

| #   | Entscheidung                              | Begründung                                                       |
| --- | ----------------------------------------- | ---------------------------------------------------------------- |
| D1  | Modular Monolith statt Microservices      | Geringe Last, kleines Team, einfaches Deployment                 |
| D2  | Hybrides Event-Sourcing + CRUD            | ES nur für Core Domain (Kassenbetrieb), CRUD für Stammdaten      |
| D3  | POST-only API                             | Vereinfachung, konsistentes Verhalten, keine Cache-Probleme      |
| D4  | Fat Events mit eingebetteten Daten        | Self-contained Events, kein Nachladen bei Replay                 |
| D5  | Synchrone Projektion für CQRS             | Starke Konsistenz nötig (Saldo muss sofort stimmen)              |
| D6  | Deutsche Ubiquitous Language              | Domäne der Zielgruppe (Vereine) ist deutsch                      |
| D7  | PostgreSQL als Event Store                | Kein separates System nötig, ACID, JSONB, Team-Expertise         |
| D8  | JWT mit HS256, 12h Gültigkeit             | Simpel, passend für Session-Dauer einer Veranstaltung            |
| D9  | Argon2id für Passwort-Hashing             | State-of-the-Art, resistent gegen GPU/ASIC-Angriffe              |
| D10 | Soft-Deletes statt physischer Löschung    | Referenzielle Integrität, Audit-Trail                            |
| D11 | Kein globaler State-Store im Frontend     | React Hooks + Singletons genügen für die Komplexität             |
| D12 | Validierung auf beiden Seiten (Zod + zog) | Defense-in-Depth, Backend als Source of Truth                    |
| D13 | Feature-basierte Ordnerstruktur           | Co-Location von zusammengehörigem Code                           |
| D14 | Append-only mit DB-Triggern               | Defense-in-Depth: Immutabilität auf DDL- und DML-Ebene erzwungen |
| D15 | Snapshot als Event                        | Pragmatischer Ansatz, wird durch CQRS-Projektionen abgelöst      |

### 15.3 Offene Designfragen

| #   | Frage                                                      | Kontext                                          |
| --- | ---------------------------------------------------------- | ------------------------------------------------ |
| Q1  | Wie wird die Tischumbuchung atomar umgesetzt?              | Zwei Events (Storno + Neubestellung) in einer TX |
| Q2  | Wie wird der Tagesabschluss mit offenen Tischen behandelt? | Manuelles Schließen vs. automatische Stornierung |
| Q3  | Wie wird Bon-Druck integriert?                             | Thermaldrucker, Side-Effect nach Bestell-Event   |
| Q4  | Wann und wie wird Offline-Fähigkeit implementiert?         | Service Worker, lokale Queue, Sync bei Reconnect |
| Q5  | Soll der Freibon ein eigener Event-Typ werden?             | Freie Preiseingabe ohne Produkt-Zuordnung        |
