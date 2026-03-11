# Domain-Driven Design (DDD) — Theorie

Dieses Dokument ist ein theoretisches Nachschlagewerk für Domain-Driven Design. Es erklärt die zentralen DDD-Konzepte, zeigt typische Muster und Anti-Patterns und gibt eine Entscheidungshilfe, wann DDD sinnvoll ist.

---

## Inhaltsverzeichnis

1. [Was ist Domain-Driven Design?](#1-was-ist-domain-driven-design)
2. [Strategisches Design](#2-strategisches-design)
3. [Taktisches Design](#3-taktisches-design)
4. [DDD und Architekturstile](#4-ddd-und-architekturstile)
5. [DDD-Lifecycle](#5-ddd-lifecycle)
6. [Anti-Patterns und Fallstricke](#6-anti-patterns-und-fallstricke)
7. [Entscheidungshilfe: Wann lohnt sich DDD?](#7-entscheidungshilfe-wann-lohnt-sich-ddd)
8. [Referenzen](#8-referenzen)

---

## 1. Was ist Domain-Driven Design?

Domain-Driven Design ist ein Ansatz zur Softwareentwicklung, der die **Fachdomäne** in den Mittelpunkt stellt. Eric Evans formulierte DDD 2003 in seinem Buch _"Domain-Driven Design: Tackling Complexity in the Heart of Software"_. Die Kernthese:

> **Software sollte die Sprache und Struktur der Fachdomäne widerspiegeln, nicht die der Datenbank oder des Frameworks.**

DDD umfasst zwei Ebenen:

- **Strategisches Design** — Wie wird das System in Fachbereiche (Bounded Contexts) aufgeteilt?
- **Taktisches Design** — Wie werden Domain-Objekte innerhalb eines Bounded Context modelliert?

### 1.1 Historischer Kontext

**Eric Evans (2003):** Das Grundlagenwerk _Domain-Driven Design_ entstand aus Evans' Beobachtung, dass komplexe Softwareprojekte nicht an technischen Problemen scheitern, sondern an der Lücke zwischen Domänenwissen und Code. Evans entwickelte eine gemeinsame Sprache für Praktiker: Begriffe wie _Ubiquitous Language_, _Bounded Context_, _Aggregate_ und _Domain Event_ wurden durch dieses Buch zu Industriestandards.

**Vaughn Vernon (2013):** _Implementing Domain-Driven Design_ (IDDD) übersetzte Evans' theoretische Konzepte in konkrete Implementierungsanleitungen. Vernon betonte besonders das strategische Design und popularisierte den Begriff „Context Mapping". Sein zweites Buch _Domain-Driven Design Distilled_ (2016) machte DDD einem breiteren Publikum zugänglich.

**Die DDD-Community (2010er):** Praktiker wie Greg Young (CQRS, Event-Sourcing), Udi Dahan (Domain Events, NServiceBus) und Alberto Brandolini (Event Storming) erweiterten das DDD-Ökosystem erheblich. CQRS und Event-Sourcing entstanden als natürliche Erweiterungen des DDD-Denkens, sind aber eigenständige Patterns.

### 1.2 Das Problem: Accidental Complexity

DDD adressiert primär **accidental complexity** — Komplexität, die nicht aus der Fachdomäne selbst stammt, sondern aus schlechter Modellierung:

- **Impedance Mismatch:** Das Domänenmodell ist direkt an das Datenbankschema gekoppelt. Änderungen an der Fachlogik erfordern Datenbankmigrationen.
- **Anemic Domain Model:** Domänenobjekte sind nur Datencontainer. Geschäftslogik ist in Services verstreut und schwer zu lokalisieren.
- **Vocabulary Gap:** Entwickler und Domänenexperten sprechen unterschiedliche Sprachen. Im Code heißt es `getRecordById`, im Business „Bestellung abrufen".
- **Big Ball of Mud:** Ohne klare Grenzen wächst ein System zu einem unkontrollierbaren Geflecht von Abhängigkeiten.

### 1.3 Abgrenzung zu anderen Ansätzen

| Ansatz                  | Fokus                                        | Wenn sinnvoll                                      |
| ----------------------- | -------------------------------------------- | -------------------------------------------------- |
| **Transaction Script**  | Prozeduraler Ablauf, direkt auf DB           | Einfache CRUD-Operationen ohne Geschäftsregeln     |
| **Table Module**        | Eine Klasse pro DB-Tabelle                   | Datenbankzentrierte Anwendungen                    |
| **Active Record**       | Domain-Objekte kennen ihre Persistenz        | Web-Frameworks (Rails, Django) mit einfacher Logik |
| **Domain Model (DDD)**  | Reiches Modell, persistenzunabhängig         | Komplexe Geschäftslogik, die sich weiterentwickelt |
| **Anemic Domain Model** | Domain-Objekte ohne Verhalten (Anti-Pattern) | Nicht empfohlen — führt zu verstreuter Logik       |

**Faustregel:** DDD lohnt sich, wenn die **Geschäftslogik** die **Persistenzlogik** an Komplexität übersteigt. Für einfache CRUD-Anwendungen ist DDD Overkill.

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

### 2.3 Sub-Domain-Klassifikation

Nicht alle Teile einer Fachdomäne sind gleich wichtig. DDD unterscheidet drei Typen von Sub-Domains:

| Sub-Domain-Typ            | Beschreibung                                                         | Beispiele                                               | Empfehlung                                                |
| ------------------------- | -------------------------------------------------------------------- | ------------------------------------------------------- | --------------------------------------------------------- |
| **Core Domain**           | Alleinstellungsmerkmal des Unternehmens; höchster strategischer Wert | Preisalgorithmus (Amazon), Empfehlungs-Engine (Netflix) | Eigene Entwicklung, bestes Team, DDD vollständig anwenden |
| **Supporting Sub-Domain** | Notwendig, aber kein Wettbewerbsvorteil                              | Bestellverwaltung, Inventar, Reporting                  | Eigene Entwicklung mit einfacherem Ansatz                 |
| **Generic Sub-Domain**    | Standardfunktionalität, die jedes System braucht                     | Authentifizierung, E-Mail-Versand, PDF-Generierung      | Standardsoftware oder Open-Source einsetzen               |

**Entscheidungshilfe:**

```
Ist diese Funktionalität der Kern unseres Wettbewerbsvorteils?
├── Ja  → Core Domain: Vollständiges DDD, eigene Entwicklung
└── Nein
    Ist sie kaufbar/lizenzierbar?
    ├── Ja  → Generic Sub-Domain: Standardlösung verwenden
    └── Nein → Supporting Sub-Domain: Einfachere Patterns, ggf. eigene Entwicklung
```

**Warum die Klassifikation wichtig ist:**

Teams verschwenden oft Ressourcen damit, Generic Sub-Domains (wie Auth oder E-Mail) perfekt zu modellieren, während die Core Domain vernachlässigt wird. Die Sub-Domain-Klassifikation hilft, Investitionen auf die wichtigsten Bereiche zu fokussieren.

### 2.4 Context Mapping

Context Mapping beschreibt, wie Bounded Contexts miteinander interagieren. DDD definiert neun Beziehungstypen (Context Map Patterns) und drei Team-Beziehungen:

**Team-Beziehungen:**

| Beziehung               | Beschreibung                                                                         |
| ----------------------- | ------------------------------------------------------------------------------------ |
| **Mutually Dependent**  | Beide Teams müssen gemeinsam liefern; enge Koordination erforderlich (→ Partnership) |
| **Upstream/Downstream** | Das Upstream-Team beeinflusst das Downstream-Team; umgekehrt nicht wesentlich        |
| **Free**                | Unabhängige Teams ohne organisatorische oder technische Verknüpfung                  |

**Context Map Patterns:**

| Pattern                   | Richtung              | Beschreibung                                                                                  | Typisches Beispiel                                             |
| ------------------------- | --------------------- | --------------------------------------------------------------------------------------------- | -------------------------------------------------------------- |
| **Partnership**           | Bidirektional         | Zwei Teams koordinieren sich eng, um gemeinsam zu liefern                                     | Frontend- und Backend-Team einer Feature-Gruppe                |
| **Shared Kernel**         | Bidirektional         | Explizit geteilter Code/Modell zwischen Kontexten; Änderungen nur mit Zustimmung beider Teams | `Money`-Typ in Bestellwesen und Buchhaltung                    |
| **Customer/Supplier**     | Upstream → Downstream | Downstream-Prioritäten fließen in Upstream-Planung ein; Upstream liefert nach Absprache       | Produktkatalog (Supplier) → Bestellwesen (Customer)            |
| **Conformist**            | Upstream → Downstream | Downstream übernimmt exakt das Modell des Upstream — keine Verhandlung                        | Microservice übernimmt das Modell eines Legacy-Systems         |
| **Anti-Corruption Layer** | Upstream → Downstream | Downstream isoliert sich durch eine Übersetzungsschicht                                       | Übersetzung zwischen internem Modell und externer Zahlungs-API |
| **Open Host Service**     | Upstream → Downstream | Upstream veröffentlicht ein stabiles, gut dokumentiertes Protokoll für alle Consumer          | REST-API mit versionierten Endpunkten                          |
| **Published Language**    | Upstream → Downstream | Standardisiertes, dokumentiertes Format für den Datenaustausch                                | Event-Schema mit Versionierung (CloudEvents, Avro)             |
| **Separate Ways**         | Keine Verbindung      | Kontexte haben keine sinnvolle Integration; Eigenentwicklung in beiden                        | Buchhaltung und Marketing ohne gemeinsame Daten                |
| **Big Ball of Mud**       | Unstrukturiert        | Bestehende chaotische Systeme — abgrenzen, nicht imitieren                                    | Legacy-Monolith ohne klare Grenzen                             |

**Praktischer Tipp:** Nicht alle Muster müssen in einer einzigen Context Map abgebildet werden. Für komplexe Systeme empfehlen sich mehrere spezialisierte Maps mit je einem Fokus (z.B. eine für Datenpropagation, eine für Team-Verantwortlichkeiten).

### 2.5 Event Storming

Event Storming ist eine kollaborative Modellierungsmethode, die von **Alberto Brandolini** entwickelt wurde. Teams visualisieren den Domänenprozess gemeinsam auf einem großen Whiteboard (oder digitalen Board) mit farbcodierten Klebezetteln — kein Code, keine UML-Diagramme.

**Drei Styles:**

| Style                            | Ziel                                                                 | Teilnehmer                   | Dauer       |
| -------------------------------- | -------------------------------------------------------------------- | ---------------------------- | ----------- |
| **Big Picture Event Storming**   | Gesamten Geschäftsprozess verstehen; Domänengrenzen identifizieren   | Alle Stakeholder             | 1–2 Tage    |
| **Process Level Event Storming** | Einzelnen Prozess im Detail modellieren; Bounded Contexts verfeinern | Domänenexperten + Entwickler | 4–8 Stunden |
| **Design Level Event Storming**  | Software-Design ableiten; Aggregate und Commands definieren          | Entwicklerteam               | 2–4 Stunden |

**Klebezettel-Legende (Standard):**

| Farbe     | Element             | Beschreibung                                           |
| --------- | ------------------- | ------------------------------------------------------ |
| 🟧 Orange | **Domain Event**    | Etwas ist passiert: `OrderPlaced`, `PaymentReceived`   |
| 🟦 Blau   | **Command**         | Auslöser eines Events: `PlaceOrder`, `RegisterPayment` |
| 🟨 Gelb   | **Actor**           | Wer führt den Command aus: `Kunde`, `System`           |
| 🟪 Lila   | **Policy**          | Reaktion auf ein Event: „Wenn X dann Y"                |
| 🟩 Grün   | **Read Model**      | Welche Daten der Actor zum Entscheiden braucht         |
| 🩷 Pink   | **External System** | Außenstehende Systeme: Zahlungsanbieter, E-Mail        |
| 🔴 Rot    | **Problem/Frage**   | Offene Fragen oder Konflikte                           |

**Ablauf (Big Picture):**

1. **Chaotische Exploration:** Alle schreiben Domain Events auf, ohne Reihenfolge
2. **Zeitliche Sortierung:** Events werden in chronologische Reihenfolge gebracht
3. **Pivotal Events markieren:** Wichtige Wendepunkte im Prozess hervorheben
4. **Commands hinzufügen:** Welche Aktionen lösen die Events aus?
5. **Actors hinzufügen:** Wer führt die Commands aus?
6. **Policies identifizieren:** Automatische Reaktionen auf Events
7. **Grenzen einzeichnen:** Potenzielle Bounded Contexts entstehen um zusammengehörige Events

**Warum Event Storming?**

- Kein Vorwissen über DDD notwendig — alle können mitmachen
- Missverständnisse zwischen Business und Entwicklung werden sofort sichtbar
- Bounded Contexts entstehen natürlich aus dem Prozess
- Viel schneller als traditionelle Requirements-Meetings

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

| Value Object | Attribute                     | Kontext                         |
| ------------ | ----------------------------- | ------------------------------- |
| `OrderItem`  | ProduktID, Name, Preis, Menge | Eine Zeile in einer Bestellung  |
| `Money`      | Betrag, Währung               | Geldbeträge als Cent-Integer    |
| `Address`    | Straße, PLZ, Ort, Land        | Liefer- oder Rechnungsadresse   |
| `DateRange`  | Von, Bis                      | Zeitraum, z.B. Gültigkeitsdauer |

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

### 3.5 Domain Services vs. Application Services

Ein häufiges Missverständnis ist die Unterscheidung zwischen Domain Services und Application Services:

| Aspekt             | Domain Service                                             | Application Service                                         |
| ------------------ | ---------------------------------------------------------- | ----------------------------------------------------------- |
| **Enthält**        | Geschäftslogik (Was ist erlaubt? Was gilt?)                | Orchestrierungslogik (Was passiert in welcher Reihenfolge?) |
| **Zustand**        | Zustandslos                                                | Zustandslos                                                 |
| **Abhängigkeiten** | Nur Domain-Objekte und Domain-Interfaces                   | Repositories, Domain Services, externe Services             |
| **Fachsprache**    | Verwendet Ubiquitous Language                              | Verwendet Ubiquitous Language                               |
| **Beispiele**      | `CalculateOrderTotal`, `ApplyTaxRules`, `ValidateDiscount` | `PlaceOrderUseCase`, `ProcessPaymentUseCase`                |
| **Schicht**        | Domain                                                     | Application                                                 |

**Entscheidungsregel:** Enthält die Logik **Geschäftsregeln** (Was darf passieren? Was ist der korrekte Wert?)? → Domain Service. Orchestriert sie nur den **Ablauf** (Lade → Verarbeite → Speichere)? → Application Service.

### 3.6 Domain Services

Ein Domain Service enthält Geschäftslogik, die **keiner einzelnen Entity zugeordnet** werden kann. Domain Services sind zustandslos und operieren auf Aggregaten oder Events.

**Beispiel:**

```go
// Domain Service: Preisberechnung
// Logik, die keiner einzelnen Entity zugeordnet werden kann
func CalculateOrderTotal(items []OrderItem, discounts []Discount) Money { ... }
func ApplyTaxRules(subtotal Money, region Region) Money { ... }
```

Diese Funktionen sind **reine Funktionen** (keine Seiteneffekte, deterministisch) und gehören zur Domain-Schicht, nicht zur Application-Schicht.

### 3.7 Application Services

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

### 3.8 Repositories

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

### 3.9 Factories

Factories kapseln die **komplexe Erstellung** von Aggregaten und Domain-Objekten. Sie gehören zur Domain-Schicht, wenn die Erstellung Geschäftsregeln beinhaltet.

**Wann Factories?**

- Die Erstellung eines Aggregates erfordert mehrere Schritte oder Validierungen
- Objekte müssen aus einem persistierten Zustand rekonstruiert werden (z.B. Event-Sourcing Replay)
- Die Erstellungslogik soll nicht im Application Service liegen

**DDD-Fabrikmuster im Vergleich:**

| Muster             | Beschreibung                                                             | Beispiel                                           |
| ------------------ | ------------------------------------------------------------------------ | -------------------------------------------------- |
| **Factory Method** | Statische Methode im Aggregat oder Value Object                          | `Order.NewDraftOrder(customerID, items)`           |
| **Factory Class**  | Eigene Klasse für komplexe Erstellung                                    | `OrderFactory.CreateFromCheckout(basket, address)` |
| **Builder**        | Schrittweise Konfiguration, `Build()` gibt fertiges Objekt zurück        | `NewOrderBuilder().WithItems(...).Build()`         |
| **Reconstitution** | Aggregate aus gespeichertem Zustand wiederherstellen (keine Validierung) | `Order.Reconstitute(events)`                       |

**Beispiel (Factory Method in Go):**

```go
// Factory Method: Neue Bestellung erstellen
func NewOrder(customerID CustomerID, items []OrderItem) (*Order, error) {
    if len(items) == 0 {
        return nil, errors.New("order must have at least one item")
    }
    return &Order{
        ID:         NewOrderID(),
        CustomerID: customerID,
        Items:      items,
        Status:     OrderStatusDraft,
        CreatedAt:  time.Now(),
    }, nil
}

// Reconstitution: Bestehende Bestellung aus DB laden (keine Validierung)
func ReconstituteOrder(id OrderID, customerID CustomerID, items []OrderItem, status OrderStatus) *Order {
    return &Order{ID: id, CustomerID: customerID, Items: items, Status: status}
}
```

### 3.10 Specifications

Das Specification-Pattern kapselt **komplexe Geschäftsregeln** in eigenständige, kombinierbare Objekte. Es macht Businessregeln explizit benennbar und wiederverwendbar.

**Wann Specifications?**

- Komplexe Auswahlkriterien (z.B. „Welche Bestellungen sind stornierbar?")
- Gleiche Regel an mehreren Stellen (Validierung + Abfrage + UI)
- Businessregeln sollen testbar und benennbar sein

**Beispiel:**

```go
type Specification[T any] interface {
    IsSatisfiedBy(candidate T) bool
}

// Konkrete Specification
type CancellableOrderSpec struct{}

func (s CancellableOrderSpec) IsSatisfiedBy(order Order) bool {
    return order.Status == OrderStatusPlaced && order.TotalPaidCents == 0
}

// Kombination: AND, OR, NOT
type AndSpec[T any] struct{ Left, Right Specification[T] }

func (s AndSpec[T]) IsSatisfiedBy(candidate T) bool {
    return s.Left.IsSatisfiedBy(candidate) && s.Right.IsSatisfiedBy(candidate)
}

// Verwendung
spec := AndSpec[Order]{
    Left:  CancellableOrderSpec{},
    Right: OwnedByCustomerSpec{CustomerID: customerID},
}
cancellable := filter(orders, spec.IsSatisfiedBy)
```

**Vorteile:**

- Geschäftsregeln sind explizit benannt und dokumentiert
- Kombinierbar (AND, OR, NOT) für komplexe Auswahllogik
- Selbe Specification für Validierung und Datenbankabfragen verwendbar

**Nachteile:**

- Für einfache Regeln Overkill (1-2 `if`-Bedingungen brauchen keine Specification)
- Kann bei zu vielen Specifications unübersichtlich werden

### 3.11 Module und Package Design

Module (oder Packages) strukturieren Bounded Contexts im Code. Das Ziel ist **hohe Kohäsion** (Zusammengehöriges liegt beisammen) und **lose Kopplung** (zwischen Modulen).

**Zwei Ansätze im Vergleich:**

| Ansatz                   | Struktur                                     | Stärken                                     | Schwächen                           |
| ------------------------ | -------------------------------------------- | ------------------------------------------- | ----------------------------------- |
| **Package by Layer**     | `controllers/`, `services/`, `repositories/` | Technische Trennung klar                    | Features über viele Ordner verteilt |
| **Package by Feature**   | `order/`, `customer/`, `payment/`            | Bounded Context = Ordner; Features isoliert | Kann zu doppeltem Code führen       |
| **Package by Component** | Feature-Ordner mit internen Schichten        | Balance aus beiden Ansätzen                 | Etwas mehr Struktur erforderlich    |

**Empfehlung für DDD:** Package by Feature oder Package by Component. Jeder Bounded Context bekommt einen eigenen Top-Level-Ordner:

```
src/
├── order/                  ← Bestellwesen (Bounded Context)
│   ├── domain/             ← Entities, Value Objects, Domain Services
│   ├── application/        ← Use Cases, Application Services
│   ├── infrastructure/     ← Repository-Implementierungen, DB-Code
│   └── api/                ← HTTP-Handler, DTOs
├── customer/               ← Kundenverwaltung (Bounded Context)
│   ├── domain/
│   ├── application/
│   └── ...
└── shared/                 ← Shared Kernel (minimal halten!)
    └── money/              ← Money-Typ, Currency
```

**Regeln:**

- Code innerhalb eines Bounded Context darf frei aufeinander zugreifen
- Zugriff **zwischen** Bounded Contexts nur über definierte Interfaces oder Events
- Der Shared Kernel enthält nur stabilen, selten geänderten Code
- Infrastruktur-Details (SQL, HTTP) gehören nicht in die Domain-Schicht

---

## 4. DDD und Architekturstile

DDD ist **architekturunabhängig** — die Konzepte funktionieren in Monolithen, Microservices und verschiedenen Schichtenarchitekturen. Entscheidend ist die Trennung der Domain-Logik von Infrastruktur-Details.

### 4.1 DDD in Monolithen vs. Microservices

**Modularer Monolith mit DDD:**

Der Monolith ist häufig der richtige Einstieg. Bounded Contexts werden als **Module** innerhalb desselben Deployments implementiert. Der Overhead von verteilten Systemen entfällt.

```
Modularer Monolith
├── order-module/       ← Bounded Context als Package
├── customer-module/    ← Bounded Context als Package
└── payment-module/     ← Bounded Context als Package
     Shared Database (mit schema per context)
```

**Microservices mit DDD:**

Jeder Bounded Context wird zu einem eigenständig deployten Microservice. DDD liefert die fachliche Grundlage für den Schnitt der Services:

> _"Each microservice should correspond to a single bounded context."_ — Sam Newman

```
Microservices-Landschaft
├── order-service/      ← Eigene DB, eigenes Deployment
├── customer-service/   ← Eigene DB, eigenes Deployment
└── payment-service/    ← Eigene DB, eigenes Deployment
     Kommunikation via Events / API
```

**Vergleich:**

| Aspekt                  | Monolith                           | Microservices                             |
| ----------------------- | ---------------------------------- | ----------------------------------------- |
| **DDD-Eignung**         | ✅ Gut — einfachere Umsetzung      | ✅ Gut — Bounded Contexts als Services    |
| **Betriebskomplexität** | Niedrig                            | Hoch (Deployment, Monitoring, Networking) |
| **Datenkonsistenz**     | Einfach (shared DB, Transaktionen) | Schwierig (Eventual Consistency)          |
| **Team-Skalierung**     | Begrenzt (gemeinsamer Codebase)    | Gut (Teams pro Service)                   |
| **Empfehlung (Start)**  | ✅ Modularer Monolith              | Nur wenn Teams und Last es erfordern      |

### 4.2 DDD mit Event-Sourcing

Event-Sourcing ist ein natürlicher Partner für DDD, aber kein Pflichtbestandteil. Domain Events aus dem taktischen Design werden direkt als persistierter Zustand genutzt:

- **Aggregates** erzeugen Domain Events statt Zustand zu speichern
- **State Reconstruction** erfolgt durch Replay aller Events
- Vollständiger **Audit Trail** ist inhärent im Modell
- **Temporal Queries**: Zustand zu beliebigem Zeitpunkt rekonstruierbar

```go
// Aggregat mit Event-Sourcing
type Order struct {
    ID     OrderID
    events []DomainEvent
}

func (o *Order) Place(items []OrderItem) {
    o.apply(OrderPlaced{Items: items, PlacedAt: time.Now()})
}

func (o *Order) apply(event DomainEvent) {
    switch e := event.(type) {
    case OrderPlaced:
        o.Items = e.Items
        o.Status = OrderStatusPlaced
    }
    o.events = append(o.events, event)
}
```

**Wann Event-Sourcing mit DDD?** Wenn Audit-Trail, Zeitreise-Abfragen oder komplexe State-Rekonstruktion benötigt werden. Für einfache CRUD-Aggregates ist Event-Sourcing Overkill.

→ Vertiefung: [Event-Sourcing Theorie](event-sourcing.md)

### 4.3 DDD mit CQRS

CQRS (Command Query Responsibility Segregation) trennt Schreib- und Lesemodelle. In Kombination mit DDD:

- **Commands** korrespondieren mit Methoden der Aggregate Root
- **Queries** nutzen optimierte Read Models (denormalisierte Views)
- Domain Events lösen Aktualisierungen der Read Models aus

```
Command Side (DDD)          Query Side (optimiert)
─────────────────           ──────────────────────
Order Aggregate         →   OrderSummaryView
  ├── PlaceOrder            CustomerOrderHistoryView
  ├── CancelOrder           ActiveOrdersView
  └── ShipOrder
```

CQRS ist auch ohne Event-Sourcing und ohne DDD verwendbar. Die Kombination aller drei (DDD + ES + CQRS) ist mächtig, aber nur für wirklich komplexe Domänen gerechtfertigt.

→ Vertiefung: [CQRS Theorie](cqrs.md)

### 4.4 DDD mit Hexagonal Architecture (Ports & Adapters)

Die Hexagonale Architektur (Alistair Cockburn, 2005) ergänzt DDD ideal: Die **Domain** ist der Kern, alle externen Details (HTTP, DB, Messaging) sind **Adapter** an **Ports** (Interfaces).

```
        ┌─────────────────────────────────────┐
        │           Hexagon (Domain)            │
        │                                       │
HTTP ───┤ Port (API)   Domain Logic  Port (DB) ├─── PostgreSQL
gRPC ───┤              - Entities              ├─── MongoDB
CLI ────┤              - Aggregates            ├─── In-Memory
        │              - Domain Services       │
        └─────────────────────────────────────┘
              Primäre Ports         Sekundäre Ports
              (Input/Driving)       (Output/Driven)
```

**Vorteile für DDD:**

- Domain-Code hat **keine Abhängigkeit** zu Frameworks oder Datenbanken
- Austauschbarkeit der Infrastruktur ohne Domain-Änderung
- Einfaches Testen der Domain-Logik ohne Datenbankverbindung

---

## 5. DDD-Lifecycle

Domänenmodelle sind keine statischen Konstrukte — sie **entwickeln sich** mit dem Verständnis der Domäne weiter. Evans nennt diesen Prozess „Refactoring Toward Deeper Insight".

### 5.1 Refactoring Toward Deeper Insight

Ein Domain-Modell beginnt oft mit unvollständigem Verständnis. Im Dialog mit Domänenexperten entstehen nach und nach tiefere Einsichten:

**Phasen der Modell-Evolution:**

```
Erste Version          Tieferes Verständnis       Reifes Modell
─────────────          ────────────────────       ─────────────
Order (Status)    →    Order + OrderLine     →    Order + OrderLine +
                       + Payment                  Payment + Invoice +
                                                  ShippingLabel
```

**Signale, dass das Modell vertieft werden muss:**

- Entwickler und Domänenexperten verwenden unterschiedliche Begriffe
- Viele `if`-Bedingungen für Sonderfälle häufen sich
- Ein Konzept hat viele verschiedene Bedeutungen in verschiedenen Kontexten
- Tests sind schwer zu schreiben, weil die Intention des Codes unklar ist

### 5.2 Breakthrough-Momente

Ein **Breakthrough** tritt auf, wenn ein tiefes Domänenverständnis das Modell fundamental vereinfacht oder klärt. Typisch: Ein einzelner Begriff wird gefunden, der ein ganzes Bündel komplexer Regeln elegant beschreibt.

**Beispiel:** Ein E-Commerce-System hat ein komplexes Netz aus `Order`, `Cart`, `Wishlist`, `SavedOrder`. Nach einem Workshop mit Domänenexperten wird klar: Sie sind alle eine einzige Idee — ein **ShoppingContext** mit verschiedenen **Stages** (Draft → Active → Placed → Fulfilled). Das neue Modell ist simpler und reflektiert, wie das Business denkt.

### 5.3 Modell-Evolution im Team

**Praktische Empfehlungen:**

| Empfehlung                       | Beschreibung                                                                                  |
| -------------------------------- | --------------------------------------------------------------------------------------------- |
| **Regelmäßige Model-Reviews**    | Monatliche Workshops mit Domänenexperten, um das Modell zu validieren                         |
| **Living Documentation**         | Das Modell in Code (Klassen, Tests) spiegelt die aktuelle Realität wider                      |
| **Ubiquitous Language pflegen**  | Neue Begriffe sofort dokumentieren; veraltete Begriffe konsequent umbenennen                  |
| **Event Storming wiederholen**   | Bei größeren Änderungen der Geschäftsprozesse erneutes Event Storming durchführen             |
| **Anti-Corruption Layer nutzen** | Bei Legacy-Systemintegration verhindern, dass schlechte Modelle die Core Domain verunreinigen |

### 5.4 Bounded Context Evolution

Mit wachsender Systemgröße können sich Bounded Contexts aufteilen, zusammenwachsen oder neu schneiden:

- **Context-Split:** Ein Bounded Context wird zu groß; Teams einigen sich auf eine Aufspaltung entlang fachlicher Grenzen
- **Context-Merge:** Zwei kleinere Kontexte teilen so viel, dass die Trennung mehr schadet als nützt
- **Context-Extract:** Aus einem Monolithen wird ein Microservice ausgelagert; der Bounded Context wird zum Service

---

## 6. Anti-Patterns und Fallstricke

### 6.1 Anemic Domain Model

**Problem:** Domain-Objekte sind nur Datencontainer ohne Verhalten. Alle Logik steckt in Services.

**Symptome:**

- Domain-Structs haben nur Felder, keine Methoden
- Application Services enthalten Geschäftslogik
- Die Domain-Schicht könnte durch DTOs ersetzt werden

**Vermeidung:** Geschäftslogik gehört in die Domain-Objekte oder Domain Services. Bei Event-Sourcing ist eine externe Zustandsrekonstruktion akzeptabel, weil Events als Value Objects keine Methoden benötigen.

### 6.2 Big Ball of Mud

**Problem:** Kein erkennbares Modell, alles ist mit allem verbunden.

**Vermeidung:** Klare Verzeichnisstruktur, Bounded Contexts als separate Module, Repository-Pattern als Abstraktion zwischen Domain und Persistenz.

### 6.3 Premature Abstraction

**Problem:** Zu viele Schichten und Interfaces für ein einfaches Problem.

**Vermeidung:** DDD-Patterns nur dort einsetzen, wo die Fachlogik es rechtfertigt. Einfache CRUD-Entities brauchen keine Domain Events, keine Aggregates und kein Event-Sourcing.

### 6.4 Shared Kernel Creep

**Problem:** Gemeinsam genutzte Typen wachsen unkontrolliert und koppeln Bounded Contexts.

**Vermeidung:** Den Shared Kernel minimal halten. Nur stabile, selten geänderte Typen teilen.

---

## 7. Entscheidungshilfe: Wann lohnt sich DDD?

### Entscheidungsmatrix

| Kriterium                 | Einfaches CRUD         | DDD sinnvoll                 |
| ------------------------- | ---------------------- | ---------------------------- |
| Geschäftsregeln           | Wenige, offensichtlich | Komplex, sich entwickelnd    |
| Datenmodell               | 1:1 mit DB-Schema      | Abweichend von DB-Schema     |
| Team-Kommunikation        | Technisch geprägt      | Fachlich geprägt             |
| Änderungsrate Fachlogik   | Niedrig                | Hoch                         |
| Audit/Compliance          | Nicht relevant         | Kritisch                     |
| Persistenzmuster          | Standard CRUD          | Event-Sourcing, CQRS         |
| Domain-Experten verfügbar | Nein / nicht nötig     | Ja und involviert            |
| Teamgröße                 | 1–3 Entwickler         | 5+ Entwickler, mehrere Teams |
| Systemlebensdauer         | Kurz (Wegwerfprojekt)  | Lang (strategisches System)  |

**Entscheidungsfluss:**

```
Ist die Domäne komplex (viele Geschäftsregeln)?
├── Nein → Einfaches CRUD reicht
└── Ja
    Werden Domänenexperten eingebunden?
    ├── Nein → Taktisches DDD allein (Entities, Value Objects)
    └── Ja
        Mehrere Teams / große Codebasis?
        ├── Nein → Taktisches DDD + Ubiquitous Language
        └── Ja → Vollständiges DDD (Strategisch + Taktisch + Event Storming)
```

**Faustregel:** DDD dort einsetzen, wo die Geschäftslogik die Persistenzlogik an Komplexität übersteigt.

### Wann DDD _nicht_ sinnvoll ist

- **Admin-Panels und CRUD-Backends:** Einfache Datenverwaltung ohne Businessregeln
- **Reporting-Services:** Primär Read-Only, keine Domain-Logik
- **Technische Services:** Auth-Service, Logging-Service, File-Upload — Generic Sub-Domains
- **Prototypen und MVPs:** Schnelle Iteration wichtiger als saubere Architektur
- **Kleines Team, klare Domäne:** 1–2 Entwickler in einer gut verstandenen Domäne brauchen keine Bounded Contexts

---

## 8. Referenzen

### Bücher

- **Eric Evans** (2003): _Domain-Driven Design: Tackling Complexity in the Heart of Software_ — Das Grundlagenwerk; führte Begriffe wie Ubiquitous Language, Bounded Context, Aggregate ein
- **Vaughn Vernon** (2013): _Implementing Domain-Driven Design_ — Praxisorientierter Nachfolger mit Fokus auf strategisches Design und Context Mapping
- **Vaughn Vernon** (2016): _Domain-Driven Design Distilled_ — Kurzfassung für Einsteiger und Manager
- **Scott Millett & Nick Tune** (2015): _Patterns, Principles, and Practices of Domain-Driven Design_ — Umfangreicher Muster-Katalog
- **Martin Kleppmann** (2017): _Designing Data-Intensive Applications_ — Vertiefung zu Event-Sourcing, Immutability und Stream Processing
- **Vlad Khononov** (2021): _Learning Domain-Driven Design_ — Moderner Einstieg: Strategic und Tactical Design mit Praxisbeispielen, Event Storming, Bounded Context Integration

### Online-Quellen

- [Martin Fowler: Domain-Driven Design](https://martinfowler.com/bliki/DomainDrivenDesign.html) — Übersicht und Einordnung von DDD
- [Martin Fowler: Bounded Context](https://martinfowler.com/bliki/BoundedContext.html) — Bounded Context im Detail
- [Martin Fowler: Anemic Domain Model](https://martinfowler.com/bliki/AnemicDomainModel.html) — Das Anti-Pattern erläutert
- [Martin Fowler: DDD Artikel-Sammlung](https://martinfowler.com/tags/domain%20driven%20design.html) — Alle DDD-bezogenen Artikel
- [DDD Foundational Guide (Spartner)](https://spartner.software/kennisbank/domain-driven-design-ddd) — Strategisches & taktisches Design kompakt, FAQ zu DDD in Legacy und ohne Buy-in
- [DDD Crew: Context Mapping](https://github.com/ddd-crew/context-mapping) — Alle neun Context Map Patterns mit visuellen Templates
- [DDD Crew: Bounded Context Canvas](https://github.com/ddd-crew/bounded-context-canvas) — Vorlage für das Definieren von Bounded Contexts
- [Alberto Brandolini: Event Storming](https://www.eventstorming.com/) — Event Storming als Modellierungsmethode
- [Event-Driven Architecture in Golang (PacktPublishing)](https://github.com/PacktPublishing/Event-Driven-Architecture-in-Golang) — DDD + ES + CQRS in Go, Aggregate-Design, Domain Event Patterns
- [Nick Tune: Domain-Driven Architecture Blog](https://medium.com/nick-tune-tech-strategy-blog) — Strategic DDD und Context Mapping in der Praxis

### Verwandte Theorie-Dokumente

- [Event-Sourcing Theorie](event-sourcing.md) — Event-Sourcing Grundlagen
- [CQRS Theorie](cqrs.md) — Command Query Responsibility Segregation
