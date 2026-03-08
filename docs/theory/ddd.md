# Domain-Driven Design (DDD) — Theorie

Dieses Dokument ist ein theoretisches Nachschlagewerk für Domain-Driven Design. Es erklärt die zentralen DDD-Konzepte, zeigt typische Muster und Anti-Patterns und gibt eine Entscheidungshilfe, wann DDD sinnvoll ist. Ein konkretes Anwendungsbeispiel findet sich im [Appendix](#appendix-anwendungsbeispiel-jotti).

> **Verwandte Dokumente:**
>
> - [Ubiquitous Language](../language.md) — Kanonische Fachbegriffe
> - [Event-Sourcing Theorie](event-sourcing.md) — Event-Sourcing Grundlagen
> - [CQRS Theorie](cqrs.md) — Command Query Responsibility Segregation
> - [Go Backend Architektur](go-backend.md) — Schichtenarchitektur im Backend
> - [Architektur-Übersicht](README.md) — Index aller Theorie-Dokumente

---

## Inhaltsverzeichnis

1. [Was ist Domain-Driven Design?](#1-was-ist-domain-driven-design)
2. [Strategisches Design](#2-strategisches-design)
3. [Taktisches Design](#3-taktisches-design)
4. [Anti-Patterns und Fallstricke](#4-anti-patterns-und-fallstricke)
5. [Entscheidungshilfe: Wann lohnt sich DDD?](#5-entscheidungshilfe-wann-lohnt-sich-ddd)
6. [Appendix: Anwendungsbeispiel (jotti)](#appendix-anwendungsbeispiel-jotti)
7. [Referenzen](#7-referenzen)

---

## 1. Was ist Domain-Driven Design?

Domain-Driven Design ist ein Ansatz zur Softwareentwicklung, der die **Fachdomäne** in den Mittelpunkt stellt. Eric Evans formulierte DDD 2003 in seinem Buch _"Domain-Driven Design: Tackling Complexity in the Heart of Software"_. Die Kernthese:

> **Software sollte die Sprache und Struktur der Fachdomäne widerspiegeln, nicht die der Datenbank oder des Frameworks.**

DDD umfasst zwei Ebenen:

- **Strategisches Design** — Wie wird das System in Fachbereiche (Bounded Contexts) aufgeteilt?
- **Taktisches Design** — Wie werden Domain-Objekte innerhalb eines Bounded Context modelliert?

---

## 2. Strategisches Design

### 2.1 Ubiquitous Language

Die Ubiquitous Language ist das Fundament von DDD. Alle Beteiligten — Entwickler, Domänenexperten, Dokumentation — verwenden **dieselben Begriffe**. Es gibt keine Übersetzung zwischen „was der Fachbereich sagt" und „was der Code tut".

**Prinzipien:**

- **Ein Begriff = eine Bedeutung.** „Bestellung" heißt im Code `Bestellung`, nicht `Order`, `Request` oder `Transaction`.
- **Begriffe sind kontextgebunden.** „Kunde" im Bestellwesen ist ein Käufer mit Lieferadresse; „Kunde" in der Kundenverwaltung ist ein Account mit Profil und Präferenzen.
- **Die Sprache entwickelt sich weiter.** Wenn ein neues Konzept entsteht, wird es benannt und dokumentiert.

### 2.2 Bounded Contexts

Ein Bounded Context ist ein klar abgegrenzter Fachbereich mit eigener Ubiquitous Language und eigenem Modell. Innerhalb eines Bounded Context sind Begriffe eindeutig; über Kontextgrenzen hinweg können sie unterschiedliche Bedeutungen haben.

**Warum Bounded Contexts?**

In einem monolithischen Modell entsteht schnell ein „Big Ball of Mud" — ein einzelnes Modell, das alle Aspekte abdecken soll und dabei inkonsistent wird. Bounded Contexts schaffen klare Grenzen:

```
┌──────────────────────────────────────────────────────────────────────┐
│                      E-Commerce (System)                             │
│                                                                      │
│  ┌─────────────────────────┐     ┌──────────────────────────────┐   │
│  │   Bestellwesen           │     │   Kundenverwaltung            │   │
│  │   (Bounded Context)      │     │   (Bounded Context)           │   │
│  │                          │     │                               │   │
│  │  Kunde = Käufer mit      │     │  Kunde = Account mit          │   │
│  │         Lieferadresse    │     │         Profil + Präferenzen  │   │
│  │  Bestellung              │     │  Adresse                      │   │
│  │  Zahlung                 │     │  Zahlungsmethode              │   │
│  │  Versand                 │     │                               │   │
│  │                          │     │                               │   │
│  │  Persistenz: Events      │     │  Persistenz: CRUD             │   │
│  └─────────────────────────┘     └──────────────────────────────┘   │
│                                                                      │
│  ┌─────────────────────────┐                                        │
│  │   Auth (Infrastruktur)   │  ← Kein eigenständiger                │
│  │   Login, JWT, Rollen     │    Bounded Context                    │
│  └─────────────────────────┘                                        │
└──────────────────────────────────────────────────────────────────────┘
```

### 2.3 Context Mapping

Context Mapping beschreibt, wie Bounded Contexts miteinander interagieren. DDD definiert verschiedene Beziehungstypen:

| Beziehungstyp             | Beschreibung                                      | Beispiel                                                       |
| ------------------------- | ------------------------------------------------- | -------------------------------------------------------------- |
| **Shared Kernel**         | Gemeinsamer Code zwischen Kontexten               | `Money`-Typ wird in Bestellwesen und Buchhaltung geteilt       |
| **Customer/Supplier**     | Ein Kontext liefert Daten, der andere konsumiert  | Produktkatalog (Supplier) → Bestellwesen (Customer)            |
| **Conformist**            | Consumer übernimmt exakt das Modell des Suppliers | Bestellwesen übernimmt Produkt-IDs und Preise aus dem Katalog  |
| **Anti-Corruption Layer** | Übersetzungsschicht zwischen Kontexten            | Übersetzung zwischen internem Modell und externer Zahlungs-API |
| **Published Language**    | Standardisiertes Format für Daten                 | Event-Format mit Versionierung für asynchrone Kommunikation    |

---

## 3. Taktisches Design

### 3.1 Entities

Eine Entity ist ein Objekt mit **eigener Identität**, die über die Zeit bestehen bleibt. Zwei Entities mit denselben Attributen sind nicht gleich, wenn ihre IDs unterschiedlich sind.

**Merkmale:**

- Eindeutige ID (z.B. Datenbank-PK, UUID)
- Veränderlicher Zustand (Attribute können sich ändern)
- Lebenszyklus (Erstellung, Änderung, Löschung)

**Beispiele:**

| Entity     | ID     | Veränderlicher Zustand          |
| ---------- | ------ | ------------------------------- |
| `Customer` | `UUID` | Name, E-Mail, Status            |
| `Order`    | `UUID` | Status, Positionen, Gesamtsumme |
| `Product`  | `UUID` | Name, Preis, Kategorie          |
| `Account`  | `UUID` | Kontostand, Inhaberdaten        |

### 3.2 Value Objects

Ein Value Object hat **keine eigene Identität**. Es wird vollständig durch seine Attribute definiert. Zwei Value Objects mit denselben Attributen sind gleich.

**Merkmale:**

- Keine ID
- Immutable (unveränderlich nach Erstellung)
- Gleichheit über Attribute, nicht über Identität
- Austauschbar (wird ersetzt, nicht geändert)

**Beispiele:**

| Value Object | Attribute                      | Kontext                                |
| ------------ | ------------------------------ | -------------------------------------- |
| `OrderItem`  | ProduktID, Name, Preis, Menge  | Eine Zeile in einer Bestellung         |
| `Money`      | Betrag, Währung                | Geldbeträge als Cent-Integer           |
| `Address`    | Straße, PLZ, Ort, Land         | Liefer- oder Rechnungsadresse          |
| `DateRange`  | Von, Bis                       | Zeitraum, z.B. Gültigkeitsdauer        |

**Warum sind Geldbeträge Value Objects?**

Geld ist das klassische Beispiel für Value Objects. Ein 10€-Schein ist austauschbar — es zählt nur der Wert, nicht welcher konkrete Schein es ist. Geldbeträge sollten als Integer (Cents) modelliert werden, um Floating-Point-Rundungsfehler zu verhindern:

```go
// RICHTIG: Geldbetrag als Cent-Integer (Value Object)
type OrderItem struct {
    PriceCents int  // 350 = 3,50€
    Quantity   int
}

// FALSCH: Geldbetrag als Float
type OrderItem struct {
    Price    float64  // 3.50 — Rundungsfehler möglich!
    Quantity int
}
```

### 3.3 Aggregates

Ein Aggregate ist ein Cluster von Entities und Value Objects, die **als eine Einheit** behandelt werden. Das Aggregate definiert eine **Konsistenzgrenze**: Alle Invarianten innerhalb des Aggregates müssen bei jeder Transaktion erfüllt sein.

**Bestandteile:**

- **Aggregate Root** — Die einzige Entity, über die von außen auf das Aggregate zugegriffen wird
- **Interne Entities/Value Objects** — Nur über die Aggregate Root erreichbar
- **Invarianten** — Geschäftsregeln, die immer gelten müssen

**Beispiel: Order-Aggregat (E-Commerce)**

```
┌─────────────────────────────────────────────────────┐
│              Order-Aggregat                           │
│              (Aggregate Root: Order)                  │
│                                                     │
│  Interne Objekte (Value Objects):                   │
│  ┌──────────────────────────────────────────────┐   │
│  │ OrderItem          ← Produkt, Menge, Preis    │   │
│  │ ShippingAddress    ← Straße, PLZ, Ort         │   │
│  │ PaymentInfo        ← Methode, Status          │   │
│  └──────────────────────────────────────────────┘   │
│                                                     │
│  Zustand:                                           │
│  ┌──────────────────────────────────────────────┐   │
│  │ TotalAmount        (berechnet)                │   │
│  │ Status             (Placed → Paid → Shipped)  │   │
│  └──────────────────────────────────────────────┘   │
│                                                     │
│  Invarianten:                                       │
│  • Gesamtbetrag = Σ (Preis × Menge) aller Items    │
│  • Nur bezahlte Bestellungen können versendet werden│
│  • Stornierung nur vor Versand möglich              │
└─────────────────────────────────────────────────────┘
```

### 3.4 Domain Events

Ein Domain Event beschreibt etwas, das in der Domäne **passiert ist** — in der Vergangenheitsform. Domain Events sind immutable und bilden ein fachliches Journal.

**Merkmale:**

- Vergangenheitsform: `OrderPlaced`, nicht `PlaceOrder`
- Immutable (unveränderlich nach Erstellung)
- Tragen alle relevanten Daten (self-contained)
- Versioniert für Schema-Evolution

**Beispiele:**

| Domain Event      | Bedeutung                              |
| ----------------- | -------------------------------------- |
| `OrderPlaced`     | Kunde hat eine Bestellung aufgegeben   |
| `PaymentReceived` | Zahlung für die Bestellung eingegangen |
| `OrderShipped`    | Bestellung wurde versendet             |
| `OrderCancelled`  | Bestellung wurde storniert             |

### 3.5 Domain Services

Ein Domain Service enthält Geschäftslogik, die **keiner einzelnen Entity zugeordnet** werden kann. Domain Services sind zustandslos und operieren auf Aggregaten oder Events.

**Beispiel:**

```go
// Domain Service: Preisberechnung
// Logik, die keiner einzelnen Entity zugeordnet werden kann
func CalculateOrderTotal(items []OrderItem, discounts []Discount) Money { ... }
func ApplyTaxRules(subtotal Money, region Region) Money { ... }
```

Diese Funktionen sind **reine Funktionen** (keine Seiteneffekte, deterministisch) und gehören zur Domain-Schicht, nicht zur Application-Schicht.

### 3.6 Application Services

Application Services orchestrieren den Ablauf einer Anwendungsoperation. Sie koordinieren Domain-Objekte, Repositories und Infrastruktur — enthalten aber **keine Geschäftslogik** selbst.

**Beispiel:**

```go
// Application Service: Bestellung aufgeben
func (s *OrderService) PlaceOrder(ctx, customerID, items, shippingAddress) {
    // 1. Kunde laden (Repository)
    // 2. Bestellung erstellen (Domain)
    // 3. Bestellung speichern (Repository)
    // 4. Event publizieren (Domain Event)
}
```

### 3.7 Repositories

Ein Repository abstrahiert den Datenzugriff. Es bietet eine **Collection-artige Schnittstelle** für das Laden und Speichern von Aggregaten.

**Prinzipien:**

- Repository-Interface gehört zur Domain-Schicht
- Implementierung gehört zur Infrastruktur-Schicht
- Ein Repository pro Aggregate Root
- Keine SQL-Details in der Domain

**Beispiele:**

| Repository     | Aggregat | Muster                                   |
| -------------- | -------- | ---------------------------------------- |
| `OrderRepo`    | Order    | Create, FindByID, Update, FindByCustomer |
| `CustomerRepo` | Customer | CRUD + Suche nach E-Mail                 |
| `ProductRepo`  | Product  | CRUD + Katalogabfragen                   |

---

## 4. Anti-Patterns und Fallstricke

### 4.1 Anemic Domain Model

**Problem:** Domain-Objekte sind nur Datencontainer ohne Verhalten. Alle Logik steckt in Services.

**Symptome:**

- Domain-Structs haben nur Felder, keine Methoden
- Application Services enthalten Geschäftslogik
- Die Domain-Schicht könnte durch DTOs ersetzt werden

**Vermeidung:** Geschäftslogik gehört in die Domain-Objekte oder Domain Services. Bei Event-Sourcing ist eine externe Zustandsrekonstruktion akzeptabel, weil Events als Value Objects keine Methoden benötigen.

### 4.2 Big Ball of Mud

**Problem:** Kein erkennbares Modell, alles ist mit allem verbunden.

**Vermeidung:** Klare Verzeichnisstruktur, Bounded Contexts als separate Module, Repository-Pattern als Abstraktion zwischen Domain und Persistenz.

### 4.3 Premature Abstraction

**Problem:** Zu viele Schichten und Interfaces für ein einfaches Problem.

**Vermeidung:** DDD-Patterns nur dort einsetzen, wo die Fachlogik es rechtfertigt. Einfache CRUD-Entities brauchen keine Domain Events, keine Aggregates und kein Event-Sourcing.

### 4.4 Shared Kernel Creep

**Problem:** Gemeinsam genutzte Typen wachsen unkontrolliert und koppeln Bounded Contexts.

**Vermeidung:** Den Shared Kernel minimal halten. Nur stabile, selten geänderte Typen teilen.

---

## 5. Entscheidungshilfe: Wann lohnt sich DDD?

### Entscheidungsmatrix

| Kriterium               | Einfaches CRUD         | DDD sinnvoll              |
| ----------------------- | ---------------------- | ------------------------- |
| Geschäftsregeln         | Wenige, offensichtlich | Komplex, sich entwickelnd |
| Datenmodell             | 1:1 mit DB-Schema      | Abweichend von DB-Schema  |
| Team-Kommunikation      | Technisch geprägt      | Fachlich geprägt          |
| Änderungsrate Fachlogik | Niedrig                | Hoch                      |
| Audit/Compliance        | Nicht relevant         | Kritisch                  |
| Persistenzmuster        | Standard CRUD          | Event-Sourcing, CQRS      |

**Faustregel:** DDD dort einsetzen, wo die Geschäftslogik die Persistenzlogik an Komplexität übersteigt.

---

## Appendix: Anwendungsbeispiel (jotti)

Die folgenden Abschnitte zeigen, wie die oben beschriebenen DDD-Konzepte konkret im jotti-Projekt (einem Non-Profit-POS-System für Vereinsfeste) umgesetzt werden.

### Warum DDD für jotti?

jotti ist kein triviales CRUD-System. Die Tisch-Operationen (Bestellungen, Zahlungen, Stornierungen, Lieferungen) haben fachliche Regeln, die über einfaches Speichern hinausgehen:

- Saldo-Berechnung als Invariante
- Nur bestellte Positionen können bezahlt, geliefert oder storniert werden
- Stornierung erfordert eine erhöhte Berechtigung (Serviceleitung/Admin)
- Jede Aktion erzeugt ein unveränderliches Event (Audit Trail)

Diese Komplexität rechtfertigt ein bewusstes Domain-Modell — aber kein Enterprise-Grade DDD mit Dutzenden Aggregaten. jotti nutzt DDD **pragmatisch und selektiv**.

### Strategisches Design in jotti

#### Ubiquitous Language

Fachbegriffe der Domäne sind deutsch (Bestellung, Zahlung, Tisch, Stornierung). Infrastruktur-Code bleibt englisch (Auth, Config, DB). Die kanonischen Begriffe sind in [docs/language.md](../language.md) dokumentiert.

#### Bounded Contexts

```
┌──────────────────────────────────────────────────────────────────────┐
│                         jotti (System)                               │
│                                                                      │
│  ┌─────────────────────────┐     ┌──────────────────────────────┐   │
│  │   Kassenbetrieb          │     │   Stammdaten                  │   │
│  │   (Bounded Context)      │     │   (Bounded Context)           │   │
│  │                          │     │                               │   │
│  │  Tisch = Abrechnungs-    │     │  Tisch = Physischer Tisch     │   │
│  │         einheit           │     │         mit Name + Status     │   │
│  │  Bestellung              │     │  Produkt + Variante           │   │
│  │  Zahlung                 │     │  Benutzer                     │   │
│  │  Lieferung               │     │                               │   │
│  │  Stornierung             │     │                               │   │
│  │                          │     │                               │   │
│  │  Persistenz: Events      │     │  Persistenz: CRUD             │   │
│  └─────────────────────────┘     └──────────────────────────────┘   │
│                                                                      │
│  ┌─────────────────────────┐                                        │
│  │   Auth (Infrastruktur)   │  ← Kein eigenständiger                │
│  │   Login, JWT, Rollen     │    Bounded Context                    │
│  └─────────────────────────┘                                        │
└──────────────────────────────────────────────────────────────────────┘
```

| Bounded Context   | Backend-Pfad                                     | API-Prefix   | Persistenz     |
| ----------------- | ------------------------------------------------ | ------------ | -------------- |
| **Kassenbetrieb** | `api/table/`                                     | `/service/*` | Event-Sourcing |
| **Stammdaten**    | `api/product/`, `api/user/`, `api/table/` (CRUD) | `/admin/*`   | CRUD           |
| **Auth**          | `api/auth/`                                      | `/auth/*`    | CRUD (Users)   |

#### Context Mapping

| Beziehungstyp             | Relevanz für jotti                                                                     |
| ------------------------- | -------------------------------------------------------------------------------------- |
| **Shared Kernel**         | `Position`-Struct wird in Kassenbetrieb und Events geteilt                             |
| **Customer/Supplier**     | Stammdaten (Supplier) → Kassenbetrieb (Customer): Produkt-IDs, Varianten-Namen, Preise |
| **Conformist**            | Kassenbetrieb übernimmt Varianten-IDs und Preise aus Stammdaten                        |
| **Anti-Corruption Layer** | Nicht nötig — jotti ist ein Monolith mit geteilter DB                                  |
| **Published Language**    | Event-Format (CloudEvents-inspiriert) mit Versionierung                                |

Die Beziehung zwischen Kassenbetrieb und Stammdaten ist ein **Customer/Supplier**-Verhältnis. Der Kassenbetrieb referenziert Produkt-IDs und kopiert Name + Preis **zum Zeitpunkt der Bestellung** in das Event. Dadurch sind Events historisch korrekt, selbst wenn der Produktpreis später geändert wird.

### Taktisches Design in jotti

#### Entities

| Entity     | ID                    | Veränderlicher Zustand       |
| ---------- | --------------------- | ---------------------------- |
| `User`     | `users.id`            | Name, Username, Role, Status |
| `Tisch`    | `tables.id`           | Name, Status                 |
| `Produkt`  | `products.id`         | Name, Kategorie, Status      |
| `Variante` | `product_variants.id` | Name, Preis, Status          |

#### Value Objects

| Value Object  | Attribute                      | Kontext                                |
| ------------- | ------------------------------ | -------------------------------------- |
| `Position`    | ID, Name, PreisCents, Quantity | Eine Zeile in einer Bestellung/Zahlung |
| `Geldbeträge` | Wert in Cents (int)            | Immer als Cent-Integer, nie Float      |
| `Rolle`       | admin, senior_service, service | Enum, keine eigene Identität           |
| `Kategorie`   | food, beverage, other          | Enum, keine eigene Identität           |

#### Das Tisch-Aggregat

```
┌─────────────────────────────────────────────────────┐
│              Tisch-Aggregat                          │
│              (Aggregate Root: Tisch)                 │
│                                                     │
│  Events (Value Objects, immutable):                 │
│  ┌──────────────────────────────────────────────┐   │
│  │ BestellungAufgegeben  ← Positionen           │   │
│  │ ZahlungRegistriert    ← Positionen           │   │
│  │ ProdukteGeliefert     ← Positionen           │   │
│  │ ProdukteStorniert     ← Positionen           │   │
│  └──────────────────────────────────────────────┘   │
│                                                     │
│  Rekonstruierter Zustand:                           │
│  ┌──────────────────────────────────────────────┐   │
│  │ SaldoCents            (Invariante: berechnet) │   │
│  │ UnbezahltePositionen  (abgeleitet)           │   │
│  │ UngeliefertePositionen (abgeleitet)          │   │
│  └──────────────────────────────────────────────┘   │
│                                                     │
│  Invarianten:                                       │
│  • Saldo = Σ Bestellungen − Σ Zahlungen − Σ Storno │
│  • Nur bestellte Positionen können bezahlt werden   │
│  • Nur bestellte Positionen können storniert werden │
│  • Event-Stream ist append-only (immutable)         │
└─────────────────────────────────────────────────────┘
```

**Warum ist der Tisch die Aggregate Root?**

- Alle Kassenbetrieb-Operationen beziehen sich auf **einen Tisch**
- Der Event-Stream wird über das Subject `"tisch:<id>"` identifiziert
- Alle Invarianten (Saldo, unbezahlte/ungelieferte Positionen) sind tisch-bezogen
- Konkurrierende Zugriffe auf denselben Tisch müssen konsistent aufgelöst werden

#### Domain Events

| Domain Event           | Typ-String                       | Bedeutung                          |
| ---------------------- | -------------------------------- | ---------------------------------- |
| `BestellungAufgegeben` | `tisch.bestellung-aufgegeben:v1` | Gast hat Positionen bestellt       |
| `ZahlungRegistriert`   | `tisch.zahlung-registriert:v1`   | Zahlung für Positionen eingegangen |
| `ProdukteStorniert`    | `tisch.produkte-storniert:v1`    | Positionen wurden storniert        |
| `ProdukteGeliefert`    | `tisch.produkte-geliefert:v1`    | Positionen wurden ausgeliefert     |

**Zusätzlich (technisch, kein Domain Event):**

| System Event | Typ-String          | Bedeutung                                      |
| ------------ | ------------------- | ---------------------------------------------- |
| `Snapshot`   | `tisch.snapshot:v1` | Technische Optimierung — komprimierter Zustand |

#### Domain Services

```go
// Domain Service: Zustandsrekonstruktion aus Events
// In: domain/table/events.go
func GetSaldoFromEvents(events []event.Event) int { ... }
func GetUnbezahltePositionenFromEvents(events []event.Event) []Position { ... }
func GetUngeliefertePositionenFromEvents(events []event.Event) []Position { ... }
```

Diese Funktionen sind **reine Funktionen** (keine Seiteneffekte, deterministisch) und gehören zur Domain-Schicht, nicht zur Application-Schicht.

#### Application Services

```go
// Application Service: Bestellung aufgeben
// In: api/table/application/command.go
func (s *CommandService) BestellungAufgeben(ctx, userID, tischID, positionen, comment) {
    // 1. Tisch aus DB laden (Repository)
    // 2. Event erstellen (Domain)
    // 3. Event speichern (Repository)
    // 4. Optional: Snapshot erstellen
}
```

#### Repositories

| Repository     | Aggregat             | Muster                              |
| -------------- | -------------------- | ----------------------------------- |
| `event_repo`   | Events (Event Store) | Append-only: WriteEvent, ReadEvents |
| `table_repo`   | Tisch (CRUD)         | CRUD: Get, GetAll, Create, Update   |
| `product_repo` | Produkt + Varianten  | CRUD mit Nested Entities            |
| `user_repo`    | Benutzer             | CRUD + Password-Hash                |

### DDD-Mapping: Was nutzt jotti von DDD?

| DDD-Konzept           | Genutzt?     | Umsetzung in jotti                                   |
| --------------------- | ------------ | ---------------------------------------------------- |
| Ubiquitous Language   | ✅ Ja        | Deutsche Fachbegriffe, dokumentiert in `language.md` |
| Bounded Contexts      | ✅ Ja        | Kassenbetrieb vs. Stammdaten vs. Auth                |
| Context Mapping       | ✅ Implizit  | Customer/Supplier (Stammdaten → Kassenbetrieb)       |
| Entities              | ✅ Ja        | User, Tisch, Produkt, Variante                       |
| Value Objects         | ✅ Ja        | Position, Geldbeträge (Cents), Rollen, Kategorien    |
| Aggregates            | ✅ Ja        | Tisch-Aggregat mit Event Stream                      |
| Domain Events         | ✅ Ja        | 4 Event-Typen + Snapshot                             |
| Domain Services       | ✅ Ja        | Zustandsrekonstruktion (reine Funktionen)            |
| Application Services  | ✅ Ja        | Command/Query Handler                                |
| Repositories          | ✅ Ja        | Repository-Pattern mit sqlc                          |
| Factories             | ⚠️ Teilweise | `New*`-Funktionen für Handler/Services               |
| Specifications        | ❌ Nein      | Nicht kompliziert genug                              |
| Sagas/Process Manager | ❌ Nein      | Keine verteilten Transaktionen                       |
| Module Mapping        | ✅ Ja        | Verzeichnisstruktur = Bounded Context                |

### Was jotti bewusst weglässt

**Specifications:** Das Specification-Pattern (komplexe Geschäftsregeln als Objekte) ist für jottis Regeln überdimensioniert. Die Validierung erfolgt direkt in `zog`-Schemas.

**Sagas:** Sagas koordinieren verteilte Transaktionen über mehrere Aggregaten hinweg. jotti ist ein Monolith — alle Operationen laufen in derselben Datenbank.

**CQRS mit separaten Datenbanken:** jotti nutzt eine einzelne PostgreSQL-Instanz für Write Store (Events) und Read Store (CRUD). Eine separate Read-DB ist für die aktuelle Last (Vereinsfeste) nicht nötig.

### Anti-Patterns in jotti

**Anemic Domain Model:** Teilweise vorhanden — die Zustandsrekonstruktion erfolgt in Domain Services (`events.go`), nicht in den Event-Structs selbst. Das ist akzeptabel, weil Event-Sourcing naturgemäß den Zustand extern rekonstruiert. Die Events selbst sind Value Objects und **sollten** keine Methoden haben.

**Big Ball of Mud:** Vermeidung durch klare Verzeichnisstruktur (`domain/`, `repository/`, `api/`), Bounded Contexts als separate Pfade, Repository-Pattern als Abstraktion.

**Premature Abstraction:** Bewusst vermieden. CRUD-Entities (User, Produkt) haben keine Domain Events, keine Aggregates, kein Event-Sourcing. Nur der Kassenbetrieb rechtfertigt die zusätzliche Komplexität.

**Shared Kernel Creep:** Der Shared Kernel ist minimal: `Position`-Struct und `event.Event`. Diese Typen sind stabil und ändern sich selten.

### Anwendung der Entscheidungsmatrix auf jotti

- **Stammdaten** → Einfaches CRUD. Keine DDD-Patterns nötig (außer Repository).
- **Kassenbetrieb** → DDD sinnvoll. Geschäftsregeln, Event-Sourcing, Audit Trail, fachliche Sprache.

---

## 7. Referenzen

### Bücher

- **Eric Evans** (2003): _Domain-Driven Design: Tackling Complexity in the Heart of Software_ — Das Grundlagenwerk
- **Vaughn Vernon** (2013): _Implementing Domain-Driven Design_ — Praxisorientierter Nachfolger
- **Scott Millett & Nick Tune** (2015): _Patterns, Principles, and Practices of Domain-Driven Design_ — Muster-Katalog

### Online-Quellen

- [DDD Foundational Guide](https://spartner.software/kennisbank/domain-driven-design-ddd) — Strategisches & taktisches Design kompakt erklärt
- [Martin Fowler: DDD](https://martinfowler.com/tags/domain%20driven%20design.html) — Artikel-Sammlung zu DDD-Patterns
- [Event-Driven Architecture in Golang](https://github.com/PacktPublishing/Event-Driven-Architecture-in-Golang) — DDD + Event-Sourcing + CQRS in Go

### Projekt-intern

- [Ubiquitous Language](../language.md) — Kanonische Fachbegriffe
- [ADR: Event-Sourcing](../adr/event-sourcing.md) — Entscheidung für Event-Sourcing
- [ADR: sqlc](../adr/orm.md) — Entscheidung für sqlc
