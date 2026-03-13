# Softwarearchitektur — Theorie

Dieses Dokument ist ein theoretisches Nachschlagewerk für Softwarearchitektur. Es erklärt grundlegende Architekturprinzipien, verbreitete Architekturstile (Layered, Hexagonal, Onion, Clean Architecture), Frontend-Architektur-Patterns sowie System-Design-Konzepte wie Monolith und Microservices. Das Dokument gibt außerdem Orientierung, wann welches Muster sinnvoll ist.

---

## Inhaltsverzeichnis

1. [Was ist Softwarearchitektur?](#1-was-ist-softwarearchitektur)
2. [Architektur-Denker: Fowler, Evans, Uncle Bob](#2-architektur-denker-fowler-evans-uncle-bob)
3. [SOLID-Prinzipien](#3-solid-prinzipien)
4. [Schichtenarchitektur (Layered Architecture)](#4-schichtenarchitektur-layered-architecture)
5. [Hexagonale Architektur (Ports & Adapters)](#5-hexagonale-architektur-ports--adapters)
6. [Onion Architecture](#6-onion-architecture)
7. [Clean Architecture (Uncle Bob)](#7-clean-architecture-uncle-bob)
8. [Vergleich der Architekturstile](#8-vergleich-der-architekturstile)
9. [Frontend-Architektur-Patterns](#9-frontend-architektur-patterns)
10. [System Design: Monolith vs. Microservices](#10-system-design-monolith-vs-microservices)
11. [Best Practices und Leitlinien](#11-best-practices-und-leitlinien)
12. [Anti-Patterns](#12-anti-patterns)
13. [Entscheidungsmatrix](#13-entscheidungsmatrix)
14. [Referenzen](#14-referenzen)

---

## 1. Was ist Softwarearchitektur?

### 1.1 Definition

Softwarearchitektur ist die hochrangige Struktur eines Softwaresystems: die Summe aller bedeutsamen Entscheidungen darüber, wie ein System aus Komponenten aufgebaut ist, wie diese Komponenten miteinander interagieren und welche Prinzipien ihr Design lenken.

> **Martin Fowler:** _"Architecture is about the important stuff. Whatever that is."_

> **Ralph Johnson:** _"Architecture is the decisions that are hard to change."_

Architekturentscheidungen zeichnen sich dadurch aus, dass sie **schwer rückgängig zu machen** sind. Sie betreffen:

- **Systemgrenzen** — Welche Verantwortlichkeiten gehören in welche Komponente?
- **Kommunikationswege** — Wie tauschen Komponenten Daten aus?
- **Abhängigkeiten** — Was darf von was abhängen?
- **Persistenz und State** — Wo und wie werden Daten gespeichert?
- **Querschnittsthemen** — Wie werden Logging, Auth, Fehlerbehandlung und Monitoring umgesetzt?

### 1.2 Architektur vs. Design

| Aspekt              | Architektur                              | Design                                  |
| ------------------- | ---------------------------------------- | --------------------------------------- |
| **Ebene**           | Systemweit, strategisch                  | Lokal, taktisch                         |
| **Veränderlichkeit**| Schwer rückgängig zu machen              | Leicht refaktorierbar                   |
| **Beispiele**       | Schichtenmodell, Datenbankwahl, API-Stil | Klassenstruktur, Methodendesign, Patterns |
| **Wirkungsbereich** | Alle Teams, alle Komponenten             | Einzelne Datei, Modul, Feature          |

**Faustregel:** Architektur = was auf dem Whiteboard erklärt wird; Design = was in der konkreten Implementierung entschieden wird.

### 1.3 Warum Architektur wichtig ist

Schlechte Architektur erzeugt **technische Schulden**, die sich exponentiell häufen:

- **Big Ball of Mud** — Unstrukturiertes Geflecht ohne klare Grenzen; Änderungen an einer Stelle brechen unvorhersehbar andere
- **Tight Coupling** — Komponenten kennen die internen Details anderer Komponenten; Austausch wird unmöglich
- **Anemic Domain Model** — Geschäftslogik ist in Services verstreut; kein zentraler Ort für Geschäftsregeln
- **Leaky Abstractions** — Implementierungsdetails sickern durch Schichtgrenzen; Austausch einer Komponente reißt andere mit

Gute Architektur hingegen schafft **accidental complexity** (= von uns erzeugte Komplexität) zu minimieren und erlaubt es dem Team, sich auf die **essential complexity** (= inhärente Fachkomplexität) zu konzentrieren.

---

## 2. Architektur-Denker: Fowler, Evans, Uncle Bob

Drei Persönlichkeiten haben die moderne Softwarearchitektur maßgeblich geprägt:

### 2.1 Martin Fowler — Patterns, Refactoring, Enterprise Architecture

Martin Fowler ist Chefwissenschaftler bei ThoughtWorks und Autor einflussreicher Standardwerke. Seine zentralen Beiträge:

- **Enterprise Application Architecture Patterns (2002):** Muster für Datenzugriff (Active Record, Repository, Data Mapper), Domänenlogik (Transaction Script, Domain Model, Table Module) und Präsentation (MVC, Page Controller, Front Controller)
- **Refactoring (1999):** Systematische Verbesserung von Code ohne Verhaltensänderung; Grundlage für kontinuierliche Verbesserung
- **Patterns of Enterprise Application Architecture (PoEAA):** Standardreferenz für Enterprise-Patterns; definierte viele Begriffe, die heute selbstverständlich sind

**Fowlers Architekturphilosophie:**

- Architektur ist der Teil, der schwer zu ändern ist — deshalb ist Reversibilität ein wichtiges Ziel
- Gute Architektur maximiert die Anzahl der offenen Entscheidungen (YAGNI, Defer Commitments)
- Pragmatismus vor Purismus: Das beste Muster für eine Situation ist das, das das Problem löst — nicht das akademisch korrekteste

### 2.2 Eric Evans — Domain-Driven Design

Eric Evans veröffentlichte 2003 _Domain-Driven Design: Tackling Complexity in the Heart of Software_ und legte damit den Grundstein für modernes Domänenmodellieren (→ siehe [DDD Theorie](ddd.md)).

Kernbeiträge:

- **Ubiquitous Language** — Gemeinsame Sprache zwischen Domänenexperten und Entwicklern, direkt im Code
- **Bounded Context** — Klare Grenzen für Teildomänen; jedes Modell ist nur in seinem Kontext gültig
- **Aggregates** — Konsistenzgrenzen für Domänenobjekte; Transaktionen überqueren keine Aggregate-Grenzen
- **Domain Events** — Ereignisse als First-Class Citizens; explizite Sprache für Zustandsänderungen

**Evans' Beitrag zur Architektur:** DDD löst das strukturelle Problem vieler Enterprise-Systeme: die Kopplung von Fachlogik an Datenbankstrukturen (Impedance Mismatch). Bounded Contexts geben Microservices ihr konzeptuelles Fundament.

### 2.3 Uncle Bob (Robert C. Martin) — Clean Architecture, SOLID

Robert C. Martin, bekannt als „Uncle Bob", ist Autor von _Clean Code_ (2008), _Clean Architecture_ (2017) und Mitunterzeichner des Agile Manifesto. Seine Kernbeiträge:

- **SOLID-Prinzipien** — Fünf Prinzipien für objektorientiertes Design; Grundlage wartbaren Codes
- **Clean Architecture** — Konzentrisches Modell mit strikter Abhängigkeitsregel: Abhängigkeiten zeigen nur nach innen
- **Dependency Inversion Principle** — Hochrangige Module sollen nicht von niedrigrangigen abhängen; beide sollen von Abstraktionen abhängen
- **Screaming Architecture** — Eine Architektur soll von sich selbst erzählen; die Verzeichnisstruktur soll die Domäne widerspiegeln, nicht das Framework

**Uncle Bobs Kernthese:**

> _"The goal of software architecture is to minimize the human resources required to build and maintain the required system."_

---

## 3. SOLID-Prinzipien

SOLID ist ein Akronym für fünf Designprinzipien objektorientierter Programmierung, die Robert C. Martin systematisiert hat. Sie gelten nicht nur für OOP, sondern sind auf jede Art von Modularchitektur anwendbar.

### 3.1 Single Responsibility Principle (SRP)

> **Ein Modul soll genau einen Grund haben, sich zu ändern.**

Ein Modul gehört einem _Akteur_ (einem Benutzer oder Stakeholder). Wenn zwei verschiedene Akteure unterschiedliche Anforderungen an dasselbe Modul stellen, führt das zu unbeabsichtigten Kopplungen.

```
❌ God-Klasse:
UserService
  ├── createUser()          ← Auth-Bereich
  ├── getUserProfile()      ← Profil-Bereich
  ├── sendWelcomeEmail()    ← Benachrichtigungs-Bereich
  └── generateInvoice()     ← Abrechnungs-Bereich

✅ Getrennte Verantwortlichkeiten:
UserRepository        ← Persistenz
UserProfileService    ← Profil-Logik
NotificationService   ← Benachrichtigungen
BillingService        ← Abrechnung
```

### 3.2 Open/Closed Principle (OCP)

> **Ein Modul soll offen für Erweiterung, aber geschlossen für Änderung sein.**

Neue Verhaltensweisen werden durch neue Implementierungen (Erweiterungen) hinzugefügt, nicht durch Änderung bestehender. Erreicht wird dies durch Abstraktion (Interfaces, Polymorphismus).

```go
// Geschlossen für Änderung: Interface bleibt stabil
type PaymentProvider interface {
    Charge(amount int, currency string) error
}

// Offen für Erweiterung: Neue Provider ohne Änderung des Kernmodells
type StripeProvider struct{}
type PayPalProvider struct{}
type MockProvider struct{}  // für Tests
```

### 3.3 Liskov Substitution Principle (LSP)

> **Instanzen einer Unterklasse müssen die Verträge der Oberklasse erfüllen.**

Subtypen müssen überall dort einsetzbar sein, wo ihr Basistyp erwartet wird — ohne das Verhalten des Programms zu verändern. Verstöße führen zu „if instanceof"-Checks, die OCP verletzen.

**Konkreter:** Ein `ReadOnlyCollection`, das `Collection` implementiert aber bei `add()` eine Exception wirft, verletzt LSP, weil der Aufrufer von `Collection` nicht damit rechnen muss.

### 3.4 Interface Segregation Principle (ISP)

> **Clients sollen nicht von Interfaces abhängen, die sie nicht nutzen.**

Große, fette Interfaces werden in kleine, spezifische aufgeteilt. Jedes Interface bedient genau die Bedürfnisse seiner Nutzer.

```go
// ❌ Fettes Interface — jeder Nutzer muss alle Methoden implementieren
type Repository interface {
    Create(item Item) error
    Read(id int) (Item, error)
    Update(item Item) error
    Delete(id int) error
    Search(query string) ([]Item, error)
    Export(format string) ([]byte, error)
}

// ✅ Segregierte Interfaces
type ItemReader interface {
    Read(id int) (Item, error)
}

type ItemWriter interface {
    Create(item Item) error
    Update(item Item) error
}

type ItemSearcher interface {
    Search(query string) ([]Item, error)
}
```

### 3.5 Dependency Inversion Principle (DIP)

> **Hochrangige Module sollen nicht von niedrigrangigen abhängen. Beide sollen von Abstraktionen abhängen. Abstraktionen sollen nicht von Details abhängen. Details sollen von Abstraktionen abhängen.**

Dies ist das Fundament der meisten modernen Architekturstile (Hexagonal, Clean, Onion). Konkret bedeutet es: Die Geschäftslogik definiert die Interfaces (Ports), die Infrastruktur implementiert sie (Adapters).

```
❌ Direkte Abhängigkeit (Hochrangig → Niedrigrangig):
OrderService → PostgresOrderRepository

✅ Invertierte Abhängigkeit (beide → Abstraktion):
OrderService → OrderRepository (Interface)
                       ↑
              PostgresOrderRepository
```

### 3.6 SOLID im Frontend

SOLID ist nicht nur für Backend-Code. Für React gilt:

| Prinzip | React-Anwendung                                                                   |
| ------- | --------------------------------------------------------------------------------- |
| **SRP** | Eine Komponente macht eine Sache: Anzeigen oder Logik — nicht beides (Container vs. Presenter) |
| **OCP** | Komponenten per Props erweiterbar, ohne Kerncode zu ändern (Render Props, Slots) |
| **LSP** | Komponenten-APIs konsistent halten; keine überraschenden Verhaltensabweichungen   |
| **ISP** | Props-Interfaces minimal halten; keine ungenutzten Props übergeben               |
| **DIP** | Backend-Aufrufe über injizierbare Clients statt direkte `fetch()`-Aufrufe        |

---

## 4. Schichtenarchitektur (Layered Architecture)

### 4.1 Grundmodell

Die Schichtenarchitektur (auch: N-Tier Architecture) ist das klassischste und verbreitetste Architekturmuster. Das System wird in horizontale Schichten aufgeteilt, die jeweils eine klar definierte Verantwortlichkeit haben.

```
┌─────────────────────────────────────────────────────────────────┐
│                        Presentation-Schicht                     │
│  HTTP-Handler, REST-Controller, GraphQL-Resolver, CLI           │
│  Verantwortung: Request parsen, validieren, Response senden     │
├─────────────────────────────────────────────────────────────────┤
│                     Application-Schicht                         │
│  Command Services, Query Services, Use Cases                    │
│  Verantwortung: Orchestrierung, Workflow-Steuerung              │
├─────────────────────────────────────────────────────────────────┤
│                       Domain-Schicht                            │
│  Entities, Value Objects, Domain Events, Domain Services        │
│  Verantwortung: Geschäftsregeln und Invarianten                 │
├─────────────────────────────────────────────────────────────────┤
│                     Infrastruktur-Schicht                       │
│  Datenbankzugriff, External APIs, Message Brokers, Cache        │
│  Verantwortung: Persistenz, externe Kommunikation               │
└─────────────────────────────────────────────────────────────────┘
```

**Abhängigkeitsregel:** Jede Schicht darf nur von der direkt darunter liegenden abhängen:

```
Presentation → Application → Domain ← Infrastruktur
```

### 4.2 Strikte vs. Relaxed Layering

| Variante           | Regel                                             | Anwendungsfall                        |
| ------------------ | ------------------------------------------------- | ------------------------------------- |
| **Strict Layering** | Jede Schicht darf nur die direkt darunterliegende verwenden | Maximale Isolation, höhere Disziplin |
| **Relaxed Layering** | Jede Schicht darf alle darunterliegenden verwenden | Weniger Overhead, pragmatisch        |

Die meisten modernen Architekturen verwenden Relaxed Layering: Die Presentation-Schicht darf direkt auf die Infrastruktur für einfache Abfragen zugreifen, ohne zwingend durch Application und Domain zu gehen.

### 4.3 Klassisches 3-Tier vs. Modernes Layered

| Aspekt            | Klassisch (3-Tier)             | Modern (Layered + DDD)                      |
| ----------------- | ------------------------------ | ------------------------------------------- |
| **Schichten**     | Presentation, Business, Data   | HTTP, Application, Domain, Repository, Infra |
| **Domäne**        | Anemic (nur Daten)             | Reiches Domänenmodell mit Verhalten         |
| **Persistenz**    | Direkte DB-Abhängigkeit        | Repository-Interface (invertiert)           |
| **Testbarkeit**   | Schwierig (alles gekoppelt)    | Gut (Domain isoliert testbar)               |

### 4.4 Stärken und Schwächen

| Stärken                                    | Schwächen                                               |
| ------------------------------------------ | ------------------------------------------------------- |
| Einfach verständlich und kommunizierbar    | Domäne kann von Infrastruktur abhängig werden (DIP-Verletzung) |
| Gute Testbarkeit bei sauberer Implementierung | „Durchbluten" von Domain-Objekten durch Schichtgrenzen |
| Verbreitet: viele Tools, Frameworks, Wissen | Tendenz zu anämischen Domänenmodellen                   |
| Klar definierte Verantwortlichkeiten        | Technologiewechsel kann mehrere Schichten betreffen     |

---

## 5. Hexagonale Architektur (Ports & Adapters)

### 5.1 Ursprung und Kernidee

Alistair Cockburn beschrieb 2005 die **Hexagonale Architektur** (auch: Ports & Adapters Architecture). Die Kernidee: Der Anwendungskern ist von seiner Umgebung vollständig entkoppelt. Die Kommunikation erfolgt ausschließlich über definierte **Ports** (Interfaces) und **Adapters** (Implementierungen).

Der Name „Hexagonal" ist metaphorisch: Die sechs Seiten eines Hexagons symbolisieren viele mögliche Eingangs- und Ausgangspunkte — keine strukturelle Vorgabe.

```
                    ┌─────────────────────────────┐
  HTTP Adapter ────►│                             │◄──── DB Adapter
  CLI Adapter  ────►│      Anwendungskern         │◄──── Cache Adapter
  Test Adapter ────►│   (Domain + App Logic)      │────► Email Adapter
                    │                             │────► Queue Adapter
                    └─────────────────────────────┘
         Primäre Ports                    Sekundäre Ports
         (Driving Side)                   (Driven Side)
```

### 5.2 Primäre und Sekundäre Ports

| Kategorie            | Richtung       | Beschreibung                                        | Beispiele                          |
| -------------------- | -------------- | --------------------------------------------------- | ---------------------------------- |
| **Primäre Ports**    | Von außen nach innen | Der Aufrufer treibt das System (driving side)  | HTTP-Request, CLI-Command, Test    |
| **Sekundäre Ports**  | Von innen nach außen | Das System treibt externe Dienste (driven side) | DB, E-Mail, Messaging, Cache       |

**Ports** sind Interfaces, die der Kern definiert. **Adapters** sind Implementierungen außerhalb des Kerns:

```go
// Port: Vom Kern definiertes Interface (sekundärer Port)
type BestellungRepository interface {
    Speichern(ctx context.Context, bestellung Bestellung) error
    LadenNachID(ctx context.Context, id int) (Bestellung, error)
}

// Adapter: Konkrete Implementierung außerhalb des Kerns
type PostgresBestellungRepository struct {
    db *pgxpool.Pool
}

func (r *PostgresBestellungRepository) Speichern(ctx context.Context, b Bestellung) error {
    _, err := r.db.Exec(ctx, "INSERT INTO bestellungen ...")
    return err
}

// Primärer Adapter: HTTP-Handler ruft den Kern über primären Port auf
func (h *Handler) BestellungAufgebenHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // h.bestellungService ist der primäre Port (Interface)
        err := h.bestellungService.BestellungAufgeben(r.Context(), cmd)
        // ...
    }
}
```

### 5.3 Stärken und Schwächen

| Stärken                                              | Schwächen                                               |
| ---------------------------------------------------- | ------------------------------------------------------- |
| Vollständige Isolation des Kerns von Infrastruktur   | Mehr Interfaces und Boilerplate als Layered             |
| Einfacher Austausch von Adaptern (DB, Messaging)     | Höhere initiale Komplexität                             |
| Exzellente Testbarkeit: Kern mit Mock-Adaptern testen | Bei kleinen Systemen Overkill                           |
| Kein Framework-Coupling im Kern                      | Konzeptueller Lernaufwand                               |
| Symmetrisch: Alle Eingaben/Ausgaben gleichbehandelt  | Port/Adapter-Proliferation bei vielen Integrationen     |

---

## 6. Onion Architecture

### 6.1 Konzept

Jeffrey Palermo beschrieb 2008 die **Onion Architecture** als Weiterentwicklung des Layered-Modells. Im Unterschied zur klassischen Schichtenarchitektur — wo die Abhängigkeiten nach unten (zur Infrastruktur) zeigen — kehrt die Onion Architecture die Abhängigkeiten um: **Alle Abhängigkeiten zeigen zur Mitte (zur Domain)**.

```
┌─────────────────────────────────────────────────────────────┐
│                     Infrastruktur                           │
│   (DB, HTTP, External APIs, UI)                             │
│  ┌───────────────────────────────────────────────────────┐  │
│  │              Applikation (Use Cases)                   │  │
│  │  ┌─────────────────────────────────────────────────┐  │  │
│  │  │          Domain Services                         │  │  │
│  │  │  ┌───────────────────────────────────────────┐  │  │  │
│  │  │  │      Domain Model (Kern)                   │  │  │  │
│  │  │  │  Entities, Value Objects, Domain Events    │  │  │  │
│  │  │  └───────────────────────────────────────────┘  │  │  │
│  │  └─────────────────────────────────────────────────┘  │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘

Abhängigkeitsrichtung: Außen → Innen (nie umgekehrt)
```

### 6.2 Schichten der Onion Architecture

| Schicht (von innen)   | Inhalt                                      | Abhängigkeiten           |
| --------------------- | ------------------------------------------- | ------------------------ |
| **Domain Model**      | Entities, Value Objects, Aggregates         | Keine (isolierter Kern)  |
| **Domain Services**   | Domänenlogik, die mehrere Entities verbindet | Nur Domain Model         |
| **Application**       | Use Cases, Command/Query Handler, DTOs      | Domain Model + Services  |
| **Infrastruktur**     | DB, HTTP, Messaging, UI                     | Alle inneren Schichten   |

**Kernunterschied zu Layered:** In der Onion Architecture definiert die **Applikationsschicht** die Repository-Interfaces — nicht die Infrastrukturschicht. Die Infrastruktur **implementiert** diese Interfaces. Das ist das Dependency Inversion Principle in Reinform.

### 6.3 Stärken und Schwächen

| Stärken                                              | Schwächen                                             |
| ---------------------------------------------------- | ----------------------------------------------------- |
| Domain-Kern vollständig unabhängig von Infrastruktur | Mehr Schichten und Indirektion als Layered            |
| Repository-Interfaces in Application-Schicht         | Initialer Designaufwand höher                        |
| Sehr testbare Kernschichten                          | Klare Abgrenzung zwischen den Schichten erfordert Disziplin |
| Eignet sich gut für DDD                              | Bei einfachen Systemen überdimensioniert              |

---

## 7. Clean Architecture (Uncle Bob)

### 7.1 Konzept

Robert C. Martin beschrieb 2012 (Artikel) und 2017 (Buch) die **Clean Architecture** als Synthese verschiedener Architekturansätze (Hexagonal, Onion, BCE). Das bekannteste Symbol ist das konzentrische Kreisdiagramm.

```
┌─────────────────────────────────────────────────────────────────┐
│  Frameworks & Drivers (blau)                                    │
│  Web, DB, UI, Devices, External Interfaces                      │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │  Interface Adapters (grün)                                 │  │
│  │  Controllers, Gateways, Presenters                         │  │
│  │  ┌─────────────────────────────────────────────────────┐  │  │
│  │  │  Application Business Rules (rot)                    │  │  │
│  │  │  Use Cases                                           │  │  │
│  │  │  ┌───────────────────────────────────────────────┐  │  │  │
│  │  │  │  Enterprise Business Rules (gelb)              │  │  │  │
│  │  │  │  Entities                                      │  │  │  │
│  │  │  └───────────────────────────────────────────────┘  │  │  │
│  │  └─────────────────────────────────────────────────────┘  │  │
│  └───────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘

→ Die Dependency Rule: Abhängigkeiten zeigen nur nach innen
```

### 7.2 Die Dependency Rule

> **„Source code dependencies must point only inward, toward higher-level policies."**

Das bedeutet konkret:

- **Entities** (innerster Ring) kennen nichts von Use Cases, Adaptern oder Frameworks
- **Use Cases** kennen Entities, aber nichts von Adaptern oder Frameworks
- **Interface Adapters** kennen Use Cases und Entities, aber keine Framework-Details
- **Frameworks & Drivers** kennen alle inneren Ringe, sind aber leicht austauschbar

### 7.3 Die vier Ringe

| Ring                          | Inhalt                                    | Charakteristika                                  |
| ----------------------------- | ----------------------------------------- | ------------------------------------------------ |
| **Entities**                  | Enterprise Business Rules, Domain-Objekte | Stabilste Schicht; ändert sich kaum              |
| **Use Cases**                 | Application Business Rules, Interactors   | Spezifisch für die Anwendung                     |
| **Interface Adapters**        | Controller, Presenter, Gateway            | Übersetzen zwischen Use Cases und Frameworks     |
| **Frameworks & Drivers**      | DB, Web, UI, Devices                      | Austauschbar; kein Wissen über Business Rules    |

### 7.4 The Humble Object Pattern

Ein zentrales Hilfsmuster in Clean Architecture: Das **Humble Object Pattern** trennt schwer testbare Logik (z.B. UI-Rendering, DB-Treiber) von der testbaren Logik durch ein Interface.

```
┌──────────────────┐     Interface     ┌─────────────────────────┐
│  Humble Object   │ ◄──────────────── │  Testable Logic         │
│  (schwer testbar)│                   │  (Presenter, Use Case)  │
│  z.B. HTTP-Resp. │                   │  einfach unit-testbar   │
└──────────────────┘                   └─────────────────────────┘
```

### 7.5 Stärken und Schwächen

| Stärken                                              | Schwächen                                             |
| ---------------------------------------------------- | ----------------------------------------------------- |
| Sehr hohe Testbarkeit aller inneren Ringe            | Viel Boilerplate (DTOs, Mapper, Interfaces)           |
| Framework- und DB-unabhängiger Kern                  | „Over-engineering" für einfache Anwendungen           |
| Maximale Flexibilität beim Technologiewechsel        | Konzeptueller Lernaufwand hoch                        |
| „Screaming Architecture": Verzeichnisstruktur zeigt Domäne | Ausführliches Mapping zwischen Schichten nötig  |

---

## 8. Vergleich der Architekturstile

### 8.1 Gegenüberstellung

| Merkmal                   | Layered        | Hexagonal      | Onion          | Clean          |
| ------------------------- | -------------- | -------------- | -------------- | -------------- |
| **Abhängigkeitsrichtung** | Oben → Unten   | Außen → Kern   | Außen → Innen  | Außen → Innen  |
| **Kernkonzept**           | Schichten      | Ports/Adapters | Zwiebel        | Konzentrische Ringe |
| **DIP angewendet**        | Optional       | Ja             | Ja             | Ja             |
| **Framework-Kopplung**    | Möglich        | Gering         | Gering         | Sehr gering    |
| **Testbarkeit**           | Mittel         | Hoch           | Hoch           | Sehr hoch      |
| **Boilerplate**           | Gering         | Mittel         | Mittel         | Hoch           |
| **Lernkurve**             | Flach          | Moderat        | Moderat        | Steil          |
| **Eignet sich für**       | Einfache bis mittlere Systeme | Systeme mit vielen Integrationen | DDD-lastige Systeme | Komplexe Enterprise-Systeme |

### 8.2 Gemeinsamkeiten aller modernen Stile

Trotz unterschiedlicher Terminologie verfolgen Hexagonal, Onion und Clean Architecture dasselbe Ziel:

1. **Isolierter Domänenkern** — Geschäftslogik hat keine Abhängigkeiten zu Infrastruktur
2. **Invertierte Abhängigkeiten** — Infrastruktur implementiert Interfaces, die der Kern definiert
3. **Testbarkeit** — Kern mit Mock-Implementierungen testbar, ohne echte DB oder HTTP-Server
4. **Austauschbarkeit** — DB, HTTP-Framework, Messaging-System ohne Kernänderungen austauschbar

```
Martin Fowler:  "The details are the UI and the database"
Uncle Bob:      "The database is a detail. The web is a detail."
Cockburn:       "The application should work the same whether driven by users, tests, or scripts"
```

---

## 9. Frontend-Architektur-Patterns

### 9.1 Warum Frontend-Architektur?

Frontend-Anwendungen wachsen in Komplexität erheblich — moderne SPAs (Single Page Applications) verwalten Routing, State, API-Kommunikation, Validierung, Authentifizierung und komplexe UI-Logik. Ohne klare Architektur entsteht schnell ein „Big Ball of Mud" aus verschachtelten Komponenten und verstreutem State.

### 9.2 Layered Frontend Architecture

Auch im Frontend lässt sich ein Schichtenmodell anwenden:

```
┌─────────────────────────────────────────────────────────────────┐
│                       UI-Schicht (Pages)                        │
│  React-Seiten, Routing, Layout                                  │
├─────────────────────────────────────────────────────────────────┤
│                   Feature-Schicht (Organisms)                   │
│  Feature-Komponenten, Feature-Hooks, lokaler State              │
├─────────────────────────────────────────────────────────────────┤
│                  Shared-Schicht (Molecules/Atoms)               │
│  Gemeinsame Komponenten, UI-Primitives (shadcn/ui)              │
├─────────────────────────────────────────────────────────────────┤
│                    Service-Schicht (Backend)                     │
│  Backend-Klassen, API-Kommunikation, Daten-Schemas (Zod)        │
└─────────────────────────────────────────────────────────────────┘
```

### 9.3 Atomic Design (Brad Frost)

Brad Frost beschrieb 2013 **Atomic Design** als Methodik für die Entwicklung von Design-Systemen. Die Analogie zur Chemie: Atome bilden Moleküle, Moleküle bilden Organismen.

```
Atoms → Molecules → Organisms → Templates → Pages
```

| Ebene         | Beschreibung                          | Beispiele                          |
| ------------- | ------------------------------------- | ---------------------------------- |
| **Atoms**     | Kleinste UI-Bausteine ohne Abhängigkeiten | Button, Input, Badge, Label    |
| **Molecules** | Kombinationen von Atoms               | SearchForm, UserAvatar, MenuItem   |
| **Organisms** | Komplexe, eigenständige UI-Sektionen  | Header, OrderCard, NavigationMenu  |
| **Templates** | Seitenstruktur ohne echte Inhalte     | PageLayout, DashboardLayout        |
| **Pages**     | Templates mit echten Inhalten         | OrderOverviewPage, AdminDashboard  |

**Praktischer Vorteil:** Komponenten auf niedrigeren Ebenen sind hochgradig wiederverwendbar und testbar. Design-Systeme (shadcn/ui, Material UI) operieren auf Atom- und Molekül-Ebene.

### 9.4 Feature-Sliced Design (FSD)

Feature-Sliced Design ist eine moderne Methodik für skalierbare Frontend-Architekturen. Sie organisiert Code vertikal nach Features statt horizontal nach technischer Rolle.

```
src/
├── app/         ← Globale Konfiguration, Provider, Routing
├── pages/       ← Seiten-Kompositionen (Routing-Endpunkte)
├── widgets/     ← Zusammengesetzte, eigenständige UI-Blöcke
├── features/    ← Business-Features (Bestellen, Bezahlen, Stornieren)
├── entities/    ← Business-Entities (Tisch, Bestellung, Produkt)
└── shared/      ← Gemeinsame Utilities, UI-Primitives, API-Client
```

**Abhängigkeitsregel (FSD):** Jede Schicht darf nur von Schichten unterhalb importieren:

```
app → pages → widgets → features → entities → shared
```

**Stärken von FSD:**
- Features sind isoliert: Änderungen an einem Feature betreffen keine anderen
- Einfaches Onboarding: Code-Organisation folgt der Fachdomäne
- Skalierbar: Neue Features werden als neue Slice hinzugefügt

### 9.5 Microfrontends

Microfrontends übertragen das Microservices-Konzept auf das Frontend: Verschiedene Teams entwickeln und deployen ihre Feature-Bereiche unabhängig.

```
┌─────────────────────────────────────────────┐
│              Shell Application               │
│  (Routing, Authentication, Navigation)       │
│                                              │
│  ┌────────────┐  ┌────────────┐  ┌────────┐ │
│  │  Checkout  │  │  Catalog   │  │ Profile│ │
│  │   MFE      │  │   MFE      │  │  MFE   │ │
│  │ (Team A)   │  │ (Team B)   │  │(Team C)│ │
│  └────────────┘  └────────────┘  └────────┘ │
└─────────────────────────────────────────────┘
```

**Technologien:** Module Federation (Webpack/Vite), Single-SPA, Iframe-basierte Integration.

**Wann sinnvoll:** Mehrere Teams, unterschiedliche Deployment-Zyklen, sehr große Frontend-Codebasis (>50k LOC). Für kleine bis mittlere Projekte ist der Overhead nicht gerechtfertigt.

### 9.6 Container/Presenter Pattern (Smart/Dumb Components)

Das **Container/Presenter Pattern** (auch Smart/Dumb Components oder Separation of Concerns) trennt Daten-Logik von Darstellungs-Logik:

```typescript
// Container (Smart): kennt die Backend-Logik, kein UI
function BestellungContainer() {
  const { bestellung, isLoading } = useBestellung(tischId);
  const { aufgeben } = useBestellungAufgeben();

  if (isLoading) return <Spinner />;
  return <BestellungCard bestellung={bestellung} onAufgeben={aufgeben} />;
}

// Presenter (Dumb): nur UI, kein Business-Wissen
function BestellungCard({
  bestellung,
  onAufgeben,
}: {
  bestellung: Bestellung;
  onAufgeben: () => void;
}) {
  return (
    <div>
      <h2>{bestellung.tischName}</h2>
      <button onClick={onAufgeben}>Bestellung aufgeben</button>
    </div>
  );
}
```

**Vorteile:** Presenter sind einfach testbar und wiederverwendbar. Mit React Hooks ist die Trennung heute oft über Custom Hooks statt Container-Komponenten realisiert.

---

## 10. System Design: Monolith vs. Microservices

### 10.1 Monolithische Architektur

Ein **Monolith** ist eine Anwendung, bei der alle Komponenten in einem einzigen Deployment-Artefakt gebündelt sind. Es gibt drei Varianten:

```
┌──────────────────────────────────────────────────────┐
│                    Monolith                          │
│                                                      │
│  ┌──────────┐  ┌──────────┐  ┌──────────────────┐   │
│  │  Modul A  │  │  Modul B  │  │     Modul C      │   │
│  │(Bestellen)│  │(Bezahlen)│  │(Stammdaten)      │   │
│  └──────────┘  └──────────┘  └──────────────────┘   │
│                                                      │
│              ┌───────────────┐                       │
│              │   Datenbank   │                       │
│              └───────────────┘                       │
└──────────────────────────────────────────────────────┘
```

| Variante               | Beschreibung                                               |
| ---------------------- | ---------------------------------------------------------- |
| **Big Ball of Mud**    | Keine Modulgrenzen; alles ist miteinander verschränkt      |
| **Structured Monolith** | Interne Modulgrenzen; getrenntes Deployment                |
| **Modular Monolith**   | Strenge Modul-Isolation mit definierten Schnittstellen; Single Deployment |

**Empfehlung:** Ein **Modular Monolith** ist für die meisten Teams der beste Startpunkt. Er bietet die Einfachheit des Monolithen mit der Modularität, die später eine Migration zu Microservices ermöglicht.

### 10.2 Microservices

**Microservices** sind ein Architekturstil, bei dem ein System aus kleinen, eigenständig deploybaren Services besteht — jeder mit eigener Datenbank und eigenem Deployment-Prozess.

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  Bestell-    │     │  Zahlungs-   │     │  Stammdaten- │
│  Service     │────►│  Service     │     │  Service     │
│  (Go + PG)   │     │  (Go + PG)   │     │  (Go + PG)   │
└──────┬───────┘     └──────┬───────┘     └──────┬───────┘
       │                    │                     │
       ▼                    ▼                     ▼
  ┌─────────┐          ┌─────────┐          ┌─────────┐
  │  DB A   │          │  DB B   │          │  DB C   │
  └─────────┘          └─────────┘          └─────────┘
```

**Kommunikation zwischen Services:**

| Stil                   | Technologie              | Geeignet für                                |
| ---------------------- | ------------------------ | ------------------------------------------- |
| **Synchron (Request/Response)** | REST, gRPC       | Echtzeit-Abfragen, simple Request-Reply     |
| **Asynchron (Event-Driven)** | Kafka, RabbitMQ, NATS | Entkopplung, hoher Durchsatz, Resilienz |

### 10.3 Monolith-first Strategy

Martin Fowler empfiehlt die **Monolith-first Strategy**: Beginne mit einem Monolithen und extrahiere Services erst dann, wenn klare Grenzen sichtbar sind.

```
Phase 1: Monolith           Phase 2: Modular Monolith    Phase 3: Microservices
                                                         (nur wenn nötig)
┌───────────────┐           ┌───────────────────────┐    ┌───┐ ┌───┐ ┌───┐
│  Alles in     │  ──────►  │  Modul A | Modul B    │    │ A │ │ B │ │ C │
│  einem        │           │  Modul C | Modul D    │    └───┘ └───┘ └───┘
└───────────────┘           └───────────────────────┘
```

**Begründung:** Modulgrenzen in einem Microservices-System falsch zu ziehen, ist viel teurer als in einem Monolithen. Erst wenn die Grenzen klar sind (durch DDD, Event Storming), zahlt sich die Extraktion aus.

### 10.4 Entscheidungsmatrix: Monolith vs. Microservices

| Kriterium                  | Monolith              | Microservices                        |
| -------------------------- | --------------------- | ------------------------------------ |
| **Teamgröße**              | 1–10 Entwickler       | 10+ Entwickler, mehrere Teams        |
| **Deployment-Frequenz**    | Gemeinsame Releases   | Unabhängige Releases pro Service     |
| **Skalierung**             | Gesamt skalieren      | Einzelne Services skalieren          |
| **Technologie-Heterogenität** | Eine Technologie   | Verschiedene Technologien möglich    |
| **Komplexität**            | Geringer Betriebsaufwand | Hohe operationale Komplexität      |
| **Datenkonsistenz**        | ACID-Transaktionen    | Eventual Consistency, Sagas          |
| **Debugging**              | Einfach (ein Prozess) | Komplex (Distributed Tracing nötig)  |
| **Startpunkt**             | ✅ Empfohlen           | ❌ Zu früh für die meisten Teams     |

### 10.5 Skalierbarkeitspatterns

Unabhängig von Monolith oder Microservices gibt es universelle Skalierungsstrategien:

| Pattern                  | Beschreibung                                              | Wann sinnvoll                           |
| ------------------------ | --------------------------------------------------------- | --------------------------------------- |
| **Horizontal Scaling**   | Mehrere Instanzen eines Services hinter Load Balancer     | Zustandslose Services                   |
| **Vertical Scaling**     | Mehr CPU/RAM für eine Instanz                             | Einfache, kurzfristige Lösung           |
| **Caching**              | Read-Ergebnisse im Cache (Redis, CDN) halten              | Hohe Leserate, niedrige Schreibrate     |
| **Database Read Replicas** | Leseabfragen auf Read Replicas, Schreibabfragen auf Primary | Leselastige Systeme               |
| **CQRS**                 | Getrennte Lese- und Schreibmodelle                        | Komplexe Query-Anforderungen            |
| **Event-Sourcing**       | Append-only Event Log statt mutable State                 | Audit-Anforderungen, hohe Schreiblast   |
| **Queue-basiertes Load Leveling** | Aufgaben über Queue verteilen              | Burst-Traffic, asynchrone Verarbeitung  |

---

## 11. Best Practices und Leitlinien

### 11.1 Dependency Rule einhalten

> **Die wichtigste Architekturentscheidung:** Abhängigkeiten zeigen immer von Infrastruktur → Applikation → Domäne. Nie umgekehrt.

```
❌ Domain importiert Repository (DB-abhängig):
import "github.com/project/repository/postgres"

✅ Domain definiert Interface (unabhängig):
type BestellungRepository interface {
    Speichern(ctx context.Context, b Bestellung) error
}
```

### 11.2 Screaming Architecture

> **Eine Architektur soll von sich selbst erzählen.** Die Verzeichnisstruktur soll den Zweck des Systems kommunizieren, nicht die technischen Details.

```
❌ Technisch organisiert:          ✅ Fachlich organisiert:
controllers/                      kassenbetrieb/
  order.go                          bestellung/
  payment.go                        zahlung/
models/                             stornierung/
  order.go                        stammdaten/
  payment.go                        produkte/
repositories/                       tische/
  order.go                        auth/
  payment.go
```

### 11.3 Ports vor Adapters designen

Beginne mit dem Interface (Port), nicht der Implementierung (Adapter). Das zwingt dazu, die benötigte Abstraktion klar zu definieren, bevor Implementierungsdetails das Design beeinflussen.

### 11.4 Fail Fast und Validierung an den Grenzen

Validiere Eingaben an **Systemgrenzen** (HTTP-Handler, Queue-Consumer). Je früher ein ungültiger Input abgefangen wird, desto weniger Schaden richtet er an. Validierungslogik gehört nicht in die Domäne — Domänen-Invarianten sind Constraints, keine Eingabe-Validierungen.

```
Eingabe → Validierung (Boundary) → Domain-Objekt → Business Logic
         ↑ Hier abfangen          ↑ Hier gilt Invariante
```

### 11.5 Defer Commitments (YAGNI)

> _"You Aren't Gonna Need It."_ — XP-Prinzip

Architekturentscheidungen so lange wie möglich offen lassen. Eine gute Architektur maximiert die Anzahl der _nicht_ getroffenen Entscheidungen. Microservices, Event Sourcing und CQRS sind keine Standards — sie sind Werkzeuge für spezifische Probleme.

### 11.6 Zweimal bauen: ADR nutzen

Architektur-Entscheidungen dokumentieren mit **Architecture Decision Records (ADR)**. Format:

```markdown
# ADR-001: PostgreSQL als Event Store

## Status: Akzeptiert

## Kontext
...

## Entscheidung
...

## Konsequenzen
...
```

ADRs schaffen Kontext für zukünftige Entwickler und verhindern, dass Entscheidungen ohne Verständnis ihrer Hintergründe revertiert werden.

---

## 12. Anti-Patterns

### 12.1 Big Ball of Mud

**Problem:** Kein erkennbares Architekturmuster; alle Teile sind mit allen anderen verbunden. Änderungen an einer Stelle brechen unvorhersehbar andere.

**Ursachen:** Fehlende Planung, organisch gewachsener Code, keine Code-Reviews, zu hoher Zeitdruck.

**Abhilfe:** Refactoring in Richtung Modular Monolith; iterative Einführung von Modulgrenzen; kein „Rewrite from scratch" ohne Architekturplan.

### 12.2 Golden Hammer

**Problem:** Ein bekanntes Werkzeug (z.B. Microservices, Event Sourcing) wird für alle Probleme eingesetzt, unabhängig davon, ob es passt.

> _„If all you have is a hammer, everything looks like a nail."_

**Abhilfe:** Muster kennen, aber situationsabhängig einsetzen. Entscheidungsmatrizen nutzen. Einfache Lösungen bevorzugen.

### 12.3 Distributed Monolith

**Problem:** Ein vermeintliches Microservices-System, bei dem alle Services eng gekoppelt sind — synchrone Aufrufe zwischen allen Services, geteilte Datenbanken, gemeinsame Deployments.

**Resultat:** Alle Nachteile eines Monolithen (Kopplung, aufwändige Koordination) plus alle Nachteile von Microservices (Netzwerklatenz, Komplexität).

**Abhilfe:** Services entlang echter Bounded Contexts schneiden; asynchrone Kommunikation bevorzugen; Datenbanken nie teilen.

### 12.4 Anemic Domain Model

**Problem:** Domain-Objekte sind reine Datencontainer ohne Verhalten. Geschäftslogik ist in Service-Klassen verstreut.

```
// Anämisch:
type Bestellung struct { Positionen []Position; Status string }
func (s *BestellungService) BestellungAufgeben(b *Bestellung) { ... }

// Reich:
type Bestellung struct { ... }
func (b *Bestellung) PositionHinzufuegen(p Position) error { ... }
```

**Folge:** Services werden zu Prozedur-Sammlungen; Geschäftsregeln sind schwer zu finden und zu testen.

### 12.5 Architecture by Implication

**Problem:** Architekturentscheidungen werden implizit getroffen, nie dokumentiert. Neues Teammitglied hat keine Orientierung; Entscheidungen werden unbewusst revertiert.

**Abhilfe:** ADRs schreiben. Architektur-Diagramme aktuell halten. Entscheidungen im Team kommunizieren.

### 12.6 Premature Architecture

**Problem:** Komplexe Architekturmuster (Clean Architecture, CQRS, Event Sourcing, Microservices) werden eingeführt, bevor das System groß genug ist, um davon zu profitieren.

**Kosten:** Enormer Overhead, verlangsamt initiale Entwicklung, erhöht Onboarding-Kosten.

**Abhilfe:** Einfach starten (Monolith + Layered), Komplexität erst einführen wenn das Problem real sichtbar ist. „Make it work, make it right, make it fast."

---

## 13. Entscheidungsmatrix

### 13.1 Architekturstil wählen

| Kriterium                        | Layered     | Hexagonal | Onion   | Clean   |
| -------------------------------- | ----------- | --------- | ------- | ------- |
| **Systemgröße**                  | Klein–Mittel | Mittel–Groß | Mittel–Groß | Groß  |
| **Domänenkomplexität**           | Niedrig–Mittel | Mittel | Mittel–Hoch | Hoch |
| **Externe Integrationen**        | Wenige      | Viele     | Mittel  | Viele   |
| **Testbarkeit Priorität**        | Mittel      | Hoch      | Hoch    | Sehr hoch |
| **Team DDD-Erfahrung**           | Keine nötig | Hilfreich | Hilfreich | Nötig  |
| **Zeit bis zum ersten Release**  | Kurz        | Mittel    | Mittel  | Lang    |

### 13.2 Wann Clean/Hexagonal/Onion lohnt sich

```
Ist die Domäne komplex (viele Geschäftsregeln)?
├── Nein → Layered Architecture reicht vollständig aus
└── Ja
    Gibt es viele externe Integrationen (DB, Messaging, APIs)?
    ├── Nein → Onion Architecture
    └── Ja → Hexagonal Architecture oder Clean Architecture
        Ist maximale Testbarkeit und Framework-Unabhängigkeit kritisch?
        ├── Nein → Hexagonal Architecture
        └── Ja → Clean Architecture
```

### 13.3 Frontend-Architektur wählen

| Kriterium                   | Einfach (Hooks + Seiten) | Feature-Sliced Design | Microfrontends |
| --------------------------- | ------------------------ | --------------------- | -------------- |
| **Teamgröße**               | 1–3                      | 3–10                  | 10+            |
| **Feature-Anzahl**          | < 10                     | 10–50                 | 50+            |
| **Deployment-Autonomie**    | Gemeinsam                | Gemeinsam             | Unabhängig     |
| **State-Komplexität**       | Gering                   | Mittel                | Hoch           |
| **Onboarding-Aufwand**      | Niedrig                  | Mittel                | Hoch           |

---

## 14. Referenzen

### Bücher

- **Robert C. Martin** (2017): _Clean Architecture: A Craftsman's Guide to Software Structure and Design_ — Das Hauptwerk zu Clean Architecture; SOLID-Prinzipien, Dependency Rule, Humble Object
- **Robert C. Martin** (2008): _Clean Code: A Handbook of Agile Software Craftsmanship_ — Grundlagen für lesbaren, wartbaren Code; SOLID, Namensgebung, Funktionsdesign
- **Martin Fowler, Kent Beck et al.** (1999): _Refactoring: Improving the Design of Existing Code_ — Systematisches Refactoring; Grundlage für evolutionäre Architektur
- **Martin Fowler** (2002): _Patterns of Enterprise Application Architecture_ — Klassiker; Enterprise Patterns wie Repository, Active Record, Domain Model, Unit of Work
- **Eric Evans** (2003): _Domain-Driven Design: Tackling Complexity in the Heart of Software_ — Ubiquitous Language, Bounded Context, Aggregate, Domain Events
- **Sam Newman** (2021): _Building Microservices, 2nd Edition_ — Microservices in der Praxis; Service-Schnitte, Kommunikation, Datenkonsistenz, Deployment
- **Sam Newman** (2019): _Monolith to Microservices_ — Migration-Strategien vom Monolithen zu Microservices; Strangler Fig Pattern

### Online-Quellen: Grundlagen

- [Martin Fowler: Software Architecture Guide](https://martinfowler.com/architecture/) — Überblick über Enterprise Patterns, Architekturentscheidungen, Einordnung moderner Ansätze
- [Martin Fowler: Who Needs an Architect?](https://martinfowler.com/ieeeSoftware/whoNeedsArchitect.pdf) — Grundsatzartikel über die Rolle von Architektur; Unterschied zwischen „Architektur als Entscheidung" und „Architektur als Infrastruktur"
- [Martin Fowler: Monolith First](https://martinfowler.com/bliki/MonolithFirst.html) — Begründung der Monolith-first Strategy; wann Microservices sinnvoll sind
- [Uncle Bob: The Clean Architecture (2012)](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html) — Originalartikel zu Clean Architecture; Dependency Rule, vier Ringe

### Online-Quellen: Architekturstile

- [Alistair Cockburn: Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/) — Originalartikel (2005) zur Hexagonalen Architektur; Ports, Adapters, symmetrische Sicht auf Eingaben und Ausgaben
- [Jeffrey Palermo: Onion Architecture (2008)](https://jeffreypalermo.com/2008/07/the-onion-architecture-part-1/) — Originalartikel zur Onion Architecture; Abhängigkeiten nach innen, Domain im Kern
- [Kristoffer Godari: Hexagonal, Onion, and Clean Architecture](https://kristogodari.com/software-architecture/hexagonal-onion-clean-architecture/) — Praxisorientierter Vergleich der drei Architekturen mit Code-Beispielen
- [Rupesh Singh: Layered vs Clean vs Onion vs Hexagonal](https://medium.com/@rup.singh88/stop-confusing-clean-onion-hexagonal-architecture-heres-when-to-use-each-692079e56267) — Entscheidungshilfe für den praktischen Einsatz
- [42coffeecups: Software Architecture Best Practices](https://www.42coffeecups.com/blog/software-architecture-best-practices) — Übersicht über Best Practices für 2025; SOLID, Clean Architecture, Layered

### Online-Quellen: System Design

- [Ajit Singh: System Design Cheat Sheet](https://singhajit.com/system-design-cheat-sheet/) — Kompakte Übersicht über System-Design-Konzepte: Load Balancing, Caching, DB-Skalierung, CAP-Theorem
- [Chris Richardson: Microservice Patterns](https://microservices.io/patterns/) — Umfangreiche Sammlung von Microservice-Patterns; Saga, CQRS, API Gateway, Service Discovery
- [Microsoft: Azure Architecture Center](https://learn.microsoft.com/en-us/azure/architecture/) — Cloud-native Architekturpatterns; Design-Prinzipien für skalierbare Systeme

### Online-Quellen: Frontend-Architektur

- [Brad Frost: Atomic Design](https://atomicdesign.bradfrost.com/) — Das Grundlagenwerk zu Atomic Design; Atoms, Molecules, Organisms, Templates, Pages
- [LogRocket: Guide to Modern Frontend Architecture Patterns](https://blog.logrocket.com/guide-modern-frontend-architecture-patterns/) — Überblick über SPA-Architekturmuster, Microfrontends, Feature-basierte Strukturen
- [CodeWithSeb: Atomic Design + Feature Slices](https://www.codewithseb.com/blog/from-components-to-systems-scalable-frontend-with-atomiec-design) — Kombination von Atomic Design und Feature-Sliced Design für skalierbare Frontends
- [Feature-Sliced Design: Dokumentation](https://feature-sliced.design/) — Offizielle Dokumentation zu FSD; Schichten, Slices, Segmente
- [Dennis Persson: 21 React Design Patterns](https://www.perssondennis.com/articles/21-fantastic-react-design-patterns-and-when-to-use-them) — Praxiskatalog mit 21 React-Patterns; Container/Presenter, Compound Components, HOC, Render Props

### Online-Quellen: Perspektiven

- [Stackademic: Fowler, Evans und Uncle Bob im Vergleich](https://blog.stackademic.com/three-perspectives-on-software-architecture-fowler-evans-and-uncle-bob-1-1-5a51dce545d1) — Drei Perspektiven auf Softwarearchitektur; Enterprise Patterns, DDD und Clean Architecture im Vergleich
- [Reddit r/softwarearchitecture: Hexagonal vs Clean vs Onion](https://www.reddit.com/r/softwarearchitecture/comments/1otdz3g/hexagonal_vs_clean_vs_onion_architecture_which_is/) — Community-Diskussion über Trade-offs der drei Architekturstile; praktische Erfahrungen

### Verwandte Theorie-Dokumente

- [DDD Theorie](ddd.md) — Domain-Driven Design: Ubiquitous Language, Bounded Context, Aggregate
- [CQRS Theorie](cqrs.md) — Command Query Responsibility Segregation
- [Event-Sourcing Theorie](event-sourcing.md) — Event-Sourcing Grundlagen und Patterns
- [Go Backend Architektur](go-backend.md) — Architektur-Patterns in Go-Backends
- [React Frontend Architektur](react-frontend.md) — React-spezifische Architekturpatterns und Design-Entscheidungen
