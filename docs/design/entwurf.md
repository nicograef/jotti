# Software-Entwurf — jotti

Dieses Dokument beschreibt den high-level Software-Entwurf für jotti — ein kostenloses, quelloffenes Mobile-Kassensystem (mPOS) für Vereine und gemeinnützige Organisationen. Der Entwurf ist unvoreingenommen und basiert ausschließlich auf den Ergebnissen des [Event Stormings](event-storming.md), den [Anforderungen](../anforderungen.md) und der [Produktbeschreibung](../produktbeschreibung.md).

Der Entwurf beschreibt _was_ und _warum_ — keine Implementierungsdetails, keine SQL-DDL, keine vollständigen Endpunkt-Listen, keine Framework-Versionen.

---

## Inhaltsverzeichnis

1. [Systemvision und Designziele](#1-systemvision-und-designziele)
   - 1.1 [Systemvision](#11-systemvision)
   - 1.2 [Designziele](#12-designziele)
   - 1.3 [Bewusste Abgrenzung](#13-bewusste-abgrenzung)
2. [Bounded Contexts und Domain Map](#2-bounded-contexts-und-domain-map)
   - 2.1 [Kontextübersicht](#21-kontextübersicht)
   - 2.2 [Klassifikation](#22-klassifikation)
   - 2.3 [Context Map](#23-context-map)
   - 2.4 [Beziehungen zwischen Kontexten](#24-beziehungen-zwischen-kontexten)
3. [Kassenbetrieb (Core Domain)](#3-kassenbetrieb-core-domain)
   - 3.1 [Tisch-Aggregat](#31-tisch-aggregat)
   - 3.2 [Invarianten](#32-invarianten)
   - 3.3 [Domain Events](#33-domain-events)
   - 3.4 [Zustandsberechnung (Event Replay)](#34-zustandsberechnung-event-replay)
   - 3.5 [Snapshot-Strategie](#35-snapshot-strategie)
   - 3.6 [Policies](#36-policies)
4. [Stammdaten (Supporting Sub-Domain)](#4-stammdaten-supporting-sub-domain)
   - 4.1 [Produkt-Aggregat](#41-produkt-aggregat)
   - 4.2 [Tisch-Stammdaten-Aggregat](#42-tisch-stammdaten-aggregat)
   - 4.3 [Benutzer-Aggregat](#43-benutzer-aggregat)
   - 4.4 [Persistenzstrategie (CRUD mit Soft-Delete)](#44-persistenzstrategie-crud-mit-soft-delete)
5. [Ausgabe (Supporting Sub-Domain)](#5-ausgabe-supporting-sub-domain)
   - 5.1 [Bondruck](#51-bondruck)
   - 5.2 [Küchendisplay (KDS)](#52-küchendisplay-kds)
   - 5.3 [Zubereitungsstatus](#53-zubereitungsstatus)
   - 5.4 [Offene Architekturentscheidungen](#54-offene-architekturentscheidungen)
6. [Abrechnung (Supporting Sub-Domain)](#6-abrechnung-supporting-sub-domain)
   - 6.1 [Read Models und Projektionen](#61-read-models-und-projektionen)
   - 6.2 [Aggregationsstrategie](#62-aggregationsstrategie)
   - 6.3 [Tagesabschluss](#63-tagesabschluss)
   - 6.4 [Datenexport](#64-datenexport)
   - 6.5 [Offene Architekturentscheidungen](#65-offene-architekturentscheidungen)
7. [Auth (Generic Sub-Domain)](#7-auth-generic-sub-domain)
   - 7.1 [Authentifizierung](#71-authentifizierung)
   - 7.2 [Autorisierung (Rollenmodell)](#72-autorisierung-rollenmodell)
   - 7.3 [Sicheres Onboarding](#73-sicheres-onboarding)
8. [Read Models](#8-read-models)
   - 8.1 [Service-Ansichten](#81-service-ansichten)
   - 8.2 [Admin-Ansichten (Reporting)](#82-admin-ansichten-reporting)
   - 8.3 [Ausgabe-Ansichten](#83-ausgabe-ansichten)
9. [Persistenzstrategie](#9-persistenzstrategie)
   - 9.1 [Zwei Strategien, eine Datenbank](#91-zwei-strategien-eine-datenbank)
   - 9.2 [Event Store](#92-event-store)
   - 9.3 [Stammdaten (CRUD-Prinzipien)](#93-stammdaten-crud-prinzipien)
   - 9.4 [Optimistic Concurrency Control](#94-optimistic-concurrency-control)
10. [Architekturprinzipien](#10-architekturprinzipien)
    - 10.1 [Schichtenarchitektur](#101-schichtenarchitektur)
    - 10.2 [API-Design-Prinzipien](#102-api-design-prinzipien)
    - 10.3 [Frontend-Architektur](#103-frontend-architektur)
    - 10.4 [Validierung](#104-validierung)
    - 10.5 [Geldbeträge](#105-geldbeträge)
    - 10.6 [Mehrbenutzerfähigkeit](#106-mehrbenutzerfähigkeit)
    - 10.7 [Mobile-first](#107-mobile-first)
    - 10.8 [Sicherheit](#108-sicherheit)
11. [Infrastruktur und Deployment](#11-infrastruktur-und-deployment)
    - 11.1 [Architekturüberblick](#111-architekturüberblick)
    - 11.2 [Deployment-Modell](#112-deployment-modell)
12. [Ubiquitous Language](#12-ubiquitous-language)
13. [Priorisierung und Ausbaustufen](#13-priorisierung-und-ausbaustufen)
    - 13.1 [Stufe 1 — Must-have (MVP)](#131-stufe-1--must-have-mvp)
    - 13.2 [Stufe 2 — Should-have](#132-stufe-2--should-have)
    - 13.3 [Stufe 3 — Nice-to-have](#133-stufe-3--nice-to-have)
14. [Offene Entwurfsfragen](#14-offene-entwurfsfragen)

---

## 1. Systemvision und Designziele

### 1.1 Systemvision

jotti ist ein Mobile-Point-of-Sale-System für temporäre Gastronomie-Veranstaltungen gemeinnütziger Organisationen. Ehrenamtliche Servicekräfte nehmen auf ihren eigenen Smartphones (BYOD) im Browser Bestellungen auf, liefern aus, kassieren und stornieren — alles pro Tisch. Admins verwalten Produkte, Tische und Benutzer. Das System ist self-hosted per Docker Compose.

### 1.2 Designziele

| Ziel                        | Bedeutung                                                                                                     |
| --------------------------- | ------------------------------------------------------------------------------------------------------------- |
| **Radikale Einfachheit**    | Minimaler Funktionsumfang, der genau das abdeckt, was ein Vereinsfest braucht — nicht mehr.                   |
| **Mobile-first**            | Alle Interaktionen sind für Smartphone-Browser und Touch-Bedienung optimiert.                                 |
| **Lückenlose Transparenz**  | Jede Transaktion ist unveränderlich protokolliert. Kein Datenverlust, keine Manipulation.                     |
| **Null Kosten**             | Keine Hardware, keine Abo-Gebühren, keine externe Abhängigkeit.                                               |
| **Volle Datenhoheit**       | Self-hosted, alle Daten auf dem eigenen Server.                                                               |
| **Niedrige Einstiegshürde** | Keine Schulung, keine App-Installation. Browser öffnen, einloggen, loslegen.                                  |
| **Nachvollziehbarkeit**     | Event-Sourcing für den Kassenbetrieb: Jede Bestellung, Zahlung und Stornierung ist jederzeit nachvollziehbar. |

### 1.3 Bewusste Abgrenzung

Folgende Features sind **bewusst nicht enthalten** — jedes zusätzliche Feature erhöht Komplexität für ehrenamtliche Teams:

- Kartenzahlung / Zahlungsgateway
- Zertifizierte TSE (KassenSichV)
- Reservierungssystem
- Warenwirtschaft / Inventory
- Lieferservice-Integration
- Multi-Standort-Verwaltung
- Kundenverwaltung / CRM
- Selbstbedienungs-Kiosk
- Trinkgeld-Tracking

---

## 2. Bounded Contexts und Domain Map

### 2.1 Kontextübersicht

Aus dem [Event Storming](event-storming.md) ergeben sich fünf klar abgegrenzte Bereiche mit eigener Sprache, eigenen Regeln und eigener Persistenzstrategie.

| Context           | Typ                   | Beschreibung                                                          | Persistenz      |
| ----------------- | --------------------- | --------------------------------------------------------------------- | --------------- |
| **Kassenbetrieb** | Core Domain           | Tisch-basierte Vorgänge: Bestellen, Liefern, Bezahlen, Stornieren     | Event-Sourcing  |
| **Stammdaten**    | Supporting Sub-Domain | Verwaltung von Produkten, Tischen, Benutzern (CRUD)                   | CRUD            |
| **Ausgabe**       | Supporting Sub-Domain | Bondruck, Küchendisplay (KDS), Zubereitungsstatus                     | Event-getrieben |
| **Abrechnung**    | Supporting Sub-Domain | Tagesabrechnung, Umsatzberichte, Datenexport (Read-only-Projektionen) | Read-only       |
| **Auth**          | Generic Sub-Domain    | Login, Logout, Passwort-Management, Token-Verwaltung                  | Infrastruktur   |

### 2.2 Klassifikation

- **Core Domain — Kassenbetrieb:** Das Alleinstellungsmerkmal von jotti. Hier liegt die zentrale Geschäftslogik: Event-Sourcing auf dem Tisch-Aggregat, Saldo-Berechnung, Berechtigungsprüfungen für Stornierungen. Dieser Bereich verdient den meisten Entwurfs- und Testaufwand.
- **Supporting Sub-Domains — Stammdaten, Ausgabe, Abrechnung:** Notwendig für den Betrieb, aber standardisierbar. Stammdaten sind klassisches CRUD. Ausgabe und Abrechnung konsumieren Events aus dem Kassenbetrieb und projizieren sie in spezialisierte Ansichten.
- **Generic Sub-Domain — Auth:** Standard-Authentifizierung (JWT, Argon2id). Keine fachliche Tiefe, aber sicherheitskritisch.

### 2.3 Context Map

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                           jotti — Context Map                                │
│                                                                              │
│  ┌──────────────────────────────────────────┐                                │
│  │  KASSENBETRIEB (Core Domain)             │                                │
│  │                                          │                                │
│  │  Sprache: Bestellung, Position,          │                                │
│  │  Lieferung, Zahlung, Stornierung,        │                                │
│  │  Saldo, Kassenjournal                    │                                │
│  │                                          │                                │
│  │  Aggregat: Tisch (Event-Sourcing)        │                                │
│  │  Events: BestellungAufgegeben,           │                                │
│  │  ProdukteGeliefert, ZahlungRegistriert,  │                                │
│  │  ProdukteStorniert, FreibonAusgestellt   │                                │
│  └──────┬──────────────┬────────────────────┘                                │
│         │              │                                                     │
│         │ Published    │ Published                                            │
│         │ Language     │ Language                                             │
│         ▼              ▼                                                     │
│  ┌──────────────┐  ┌───────────────────┐                                     │
│  │  AUSGABE     │  │  ABRECHNUNG       │                                     │
│  │  (Supporting)│  │  (Supporting)     │                                     │
│  │              │  │                   │                                     │
│  │  KDS, Bons,  │  │  Tagesabrechnung, │                                     │
│  │  Zuberei-    │  │  Umsatzberichte,  │                                     │
│  │  tungsstatus │  │  Datenexport      │                                     │
│  └──────────────┘  └───────────────────┘                                     │
│                                                                              │
│  ┌──────────────────────────────────────────┐                                │
│  │  STAMMDATEN (Supporting Sub-Domain)      │                                │
│  │                                          │                                │
│  │  Sprache: Produkt, Variante, Kategorie,  │                                │
│  │  Preis, Tisch (Stamm), Benutzer, Rolle   │                                │
│  │                                          │                                │
│  │  Aggregate: Produkt, Tisch, Benutzer     │                                │
│  │  Persistenz: CRUD                        │                                │
│  └──────────────┬───────────────────────────┘                                │
│                 │ Customer/Supplier + ACL                                     │
│                 ▼                                                             │
│         KASSENBETRIEB (liest Produkte/Tische, friert Daten in Events ein)    │
│                                                                              │
│  ┌──────────────────────────────────────────┐                                │
│  │  AUTH (Generic — Infrastruktur)          │                                │
│  │                                          │                                │
│  │  Login, Logout, Passwort, Token          │                                │
│  │  → Open Host Service für alle Contexts   │                                │
│  └──────────────────────────────────────────┘                                │
└──────────────────────────────────────────────────────────────────────────────┘
```

### 2.4 Beziehungen zwischen Kontexten

| Upstream      | Downstream    | Beziehungstyp                     | Beschreibung                                                                             |
| ------------- | ------------- | --------------------------------- | ---------------------------------------------------------------------------------------- |
| Stammdaten    | Kassenbetrieb | Customer/Supplier + ACL           | Kassenbetrieb liest Produkte/Tische, friert Daten zum Bestellzeitpunkt in Fat Events ein |
| Kassenbetrieb | Ausgabe       | Published Language (Event-driven) | Bestellungs-Events triggern KDS-Anzeige und Bon-Druck                                    |
| Kassenbetrieb | Abrechnung    | Published Language (Event-driven) | Tisch-Events werden zu Auswertungen projiziert                                           |
| Auth          | Kassenbetrieb | Open Host Service                 | Token mit Benutzer-ID und Rolle                                                          |
| Auth          | Stammdaten    | Open Host Service                 | Token mit Benutzer-ID und Rolle                                                          |
| Auth          | Ausgabe       | Open Host Service                 | Token mit Benutzer-ID und Rolle                                                          |
| Auth          | Abrechnung    | Open Host Service                 | Token mit Benutzer-ID und Rolle                                                          |

**Anti-Corruption Layer (ACL):** Der Kassenbetrieb schützt sich vor Stammdaten-Änderungen, indem Bestellungs-Events alle relevanten Produktdaten zum Zeitpunkt der Bestellung enthalten (Fat Events). Spätere Preisänderungen haben keinen Einfluss auf historische Bestellungen.

---

## 3. Kassenbetrieb (Core Domain)

### 3.1 Tisch-Aggregat

Das Tisch-Aggregat ist die zentrale transaktionale Grenze im Kassenbetrieb. Jede Bestellung, Lieferung, Zahlung, Stornierung und jeder Freibon verändert den Zustand genau eines Tisches. Der Zustand wird nicht direkt gespeichert, sondern aus dem Event Stream berechnet (→ [3.4 Zustandsberechnung](#34-zustandsberechnung-event-replay)).

Der Tisch-Zustand folgt einer **zweistufigen Modellierung**: Ein Tisch enthält Bestellungen, und jede Bestellung enthält Positionen.

```
Tisch
├── tisch_id              (UUID)
├── saldo                 (int, Cent — berechnet)
├── event_version         (int — letzte Event-Version)
└── bestellungen[]
    ├── bestellung_id     (UUID)
    ├── bezeichnung?      (string, optional — K-07, z. B. „Familie Müller")
    ├── kommentar?        (string, optional — max. 100 Zeichen)
    ├── zeitstempel       (datetime)
    ├── benutzer_id       (UUID)
    ├── benutzer_name     (string)
    ├── ist_freibon       (bool)
    └── positionen[]
        ├── position_id   (UUID)
        ├── variante_id   (UUID)
        ├── produkt_name  (string — Fat Event)
        ├── variante_name (string — Fat Event)
        ├── kategorie     (food | beverage | other — Fat Event)
        ├── einzelpreis   (int, Cent — Fat Event)
        ├── menge         (int, ≥ 1)
        ├── geliefert     (bool)
        ├── bezahlt       (bool)
        └── storniert     (bool)
```

**Zweistufige Modellierung:** Die Gruppierung von Positionen in Bestellungen bildet den fachlichen Vorgang ab: Eine Servicekraft nimmt an einem Tisch eine Bestellung mit mehreren Positionen auf. Die optionale Bezeichnung (K-07) ermöglicht es, mehrere Gruppen an einem Tisch zu unterscheiden (z. B. „Familie Müller", „Gruppe links"). Der optionale Kommentar dient Sonderwünschen (z. B. „ohne Zwiebeln").

**Fat Events:** Die Produktdaten (Name, Variantenname, Kategorie, Einzelpreis) werden zum Zeitpunkt der Bestellung im Event eingefroren. Spätere Änderungen an den Stammdaten haben keinen Einfluss auf historische Bestellungen. Dadurch schützt sich der Kassenbetrieb per Anti-Corruption Layer vor Änderungen im Stammdaten-Context (→ [2.4 Beziehungen zwischen Kontexten](#24-beziehungen-zwischen-kontexten)).

**Freibon:** Ein Freibon ist eine Sonder-Bestellung mit freier Bezeichnung und freiem Preis, die nicht aus dem Produktkatalog stammt (→ [3.3 Domain Events — FreibonAusgestellt](#33-domain-events)). Im Zustand wird er als Bestellung mit `ist_freibon = true` und genau einer Position dargestellt.

**Abgrenzung zu Tisch-Stammdaten:** Das Tisch-Aggregat im Kassenbetrieb modelliert den laufenden Betrieb (Bestellungen, Zahlungen, Saldo). Die Tisch-Stammdaten (Name, Status active/inactive/deleted) gehören zum Stammdaten-Context (→ [4.2 Tisch-Stammdaten-Aggregat](#42-tisch-stammdaten-aggregat)). Beide teilen sich die Tisch-ID.

### 3.2 Invarianten

Das Tisch-Aggregat schützt die folgenden Geschäftsregeln. Jede Operation (Command) wird gegen diese Invarianten geprüft, bevor ein Event erzeugt wird.

**Saldo-Formel:**

$$\text{Saldo} = \sum \text{Bestellungen} - \sum \text{Zahlungen} - \sum \text{Stornierungen}$$

Alle Beträge in Cent (Integer). Der Saldo gibt den offenen Betrag des Tisches an. Ein Saldo von 0 bedeutet: alle Positionen bezahlt oder storniert.

**Liefer-Invariante:** Nur bestellte, nicht-stornierte Positionen können als geliefert markiert werden. Bereits gelieferte Positionen können nicht erneut geliefert werden. Teillieferungen sind zulässig — pro Tisch können beliebig viele Liefer-Events entstehen.

**Bezahl-Invariante:** Nur bestellte, nicht-stornierte, nicht-bezahlte Positionen können bezahlt werden. Der Zahlungsbetrag ergibt sich aus der Summe der gewählten Positionen — eine Überzahlung ist nicht möglich. Teilzahlungen sind zulässig (K-02).

**Stornierungsinvariante:** Nur bestellte, nicht-stornierte Positionen können storniert werden — **unabhängig vom Liefer- und Bezahlstatus**. Auch bereits bezahlte oder gelieferte Positionen sind stornierbar (z. B. bei Reklamationen nach Zahlung). Bei Stornierung bereits bezahlter Positionen reduziert sich der Saldo — er kann dabei temporär negativ werden. Dies ist ein bewusstes Design: Der negative Saldo wird durch nachfolgende Bestellungen verrechnet oder beim Tagesabschluss manuell ausgeglichen.

**Rolleninvariante:** Stornierungen dürfen nur durch die Rollen `senior_service` und `admin` durchgeführt werden (K-04). Die Rolle `service` hat keinen Zugriff auf die Stornierungsfunktion. Alle anderen Tischoperationen (Bestellen, Liefern, Bezahlen) stehen allen drei Rollen zur Verfügung.

**Mindestmengen-Invariante:** Jede Operation erfordert mindestens eine Position. Eine Bestellung ohne Positionen, eine Lieferung ohne gelieferte Positionen, eine Zahlung ohne bezahlte Positionen und eine Stornierung ohne stornierte Positionen sind ungültig.

### 3.3 Domain Events

Das Tisch-Aggregat kennt fünf fachliche Domain Events. Jedes Event ist unveränderlich (append-only) und wird im Event Stream des jeweiligen Tisches gespeichert. Die Event-Typen folgen der Konvention aus dem [Event Storming](event-storming.md): deutsche Sprache, Vergangenheitsform, PascalCase (Substantiv + Partizip).

#### Gemeinsame Event-Metadaten

Alle Events tragen dieselben Metadaten:

```
event_id          (UUID — eindeutige Event-ID)
tisch_id          (UUID — Aggregat-ID)
benutzer_id       (UUID — wer hat die Aktion ausgeführt)
benutzer_name     (string — Fat Event: Name zum Zeitpunkt der Aktion)
zeitstempel       (datetime — Zeitpunkt der Erzeugung)
version           (int — aufsteigende Versionsnummer pro Tisch, für Optimistic Concurrency)
```

#### BestellungAufgegeben (K-01, K-07)

Entsteht, wenn eine Servicekraft eine Bestellung am Tisch aufgibt. Enthält alle Positionen mit den Produktdaten zum Bestellzeitpunkt (Fat Event), eine optionale Bezeichnung zur Gruppenunterscheidung (K-07) und einen optionalen Kommentar.

```
BestellungAufgegeben
├── [Event-Metadaten]
├── bestellung_id     (UUID)
├── bezeichnung?      (string, optional — z. B. „Familie Müller")
├── kommentar?        (string, optional — max. 100 Zeichen)
└── positionen[]
    ├── position_id   (UUID)
    ├── variante_id   (UUID — Referenz auf Produktvariante)
    ├── produkt_name  (string — Fat Event)
    ├── variante_name (string — Fat Event)
    ├── kategorie     (food | beverage | other — Fat Event)
    ├── einzelpreis   (int, Cent — Fat Event)
    └── menge         (int, ≥ 1)
```

Die Fat-Event-Daten (`produkt_name`, `variante_name`, `kategorie`, `einzelpreis`) machen das Event selbsterklärend und unabhängig von späteren Stammdatenänderungen.

#### ProdukteGeliefert (K-03)

Entsteht, wenn eine Servicekraft bestellte Positionen als geliefert markiert. Teillieferungen sind möglich — es können beliebig viele Liefer-Events pro Tisch entstehen.

```
ProdukteGeliefert
├── [Event-Metadaten]
├── positionen[]      (Referenzen auf bestellte Positionen)
│   └── position_id   (UUID — Referenz auf eine Position aus BestellungAufgegeben)
└── kommentar?        (string, optional — max. 100 Zeichen)
```

#### ZahlungRegistriert (K-02)

Entsteht, wenn eine Servicekraft eine Barzahlung registriert. Der Betrag ergibt sich aus der Summe der gewählten Positionen. Teilzahlungen sind möglich.

```
ZahlungRegistriert
├── [Event-Metadaten]
├── positionen[]      (bezahlte Positionen)
│   └── position_id   (UUID — Referenz auf eine unbezahlte Position)
├── betrag            (int, Cent — Summe der gewählten Positionen)
└── kommentar?        (string, optional — max. 100 Zeichen)
```

Der „Gegeben"-Betrag und die Rückgeldberechnung (K-09) sind reine Frontend-Logik und werden nicht im Event gespeichert.

#### ProdukteStorniert (K-04)

Entsteht, wenn Serviceleitung oder Admin Positionen stornieren. Der Stornobetrag ergibt sich aus der Summe der stornierten Positionen. Stornierbar sind alle bestellten, nicht-stornierten Positionen — unabhängig vom Liefer- und Bezahlstatus (→ [3.2 Invarianten](#32-invarianten)).

```
ProdukteStorniert
├── [Event-Metadaten]
├── positionen[]      (stornierte Positionen)
│   └── position_id   (UUID — Referenz auf eine nicht-stornierte Position)
├── stornobetrag      (int, Cent — Summe der stornierten Positionen)
└── kommentar?        (string, optional — max. 100 Zeichen)
```

#### FreibonAusgestellt (K-11)

Entsteht, wenn eine Servicekraft einen Freibon erstellt — eine Sonderposition mit freier Bezeichnung und freiem Preis, die nicht aus dem Produktkatalog stammt. Typische Anwendungsfälle: Kinderteller mit halber Portion, Sonderwunsch, Ehrenpreis.

```
FreibonAusgestellt
├── [Event-Metadaten]
├── bestellung_id     (UUID)
├── position_id       (UUID)
├── bezeichnung       (string — freie Bezeichnung, z. B. „Kinderteller")
├── einzelpreis       (int, Cent — freier Preis)
└── kommentar?        (string, optional — max. 100 Zeichen)
```

Der Freibon wird im Tisch-Zustand als Bestellung mit `ist_freibon = true` und einer einzelnen Position dargestellt. Im Saldo wird er wie eine reguläre Bestellung behandelt — er kann geliefert, bezahlt und storniert werden. In der Abrechnung wird er separat ausgewiesen.

> **Entwurfsentscheidung:** `FreibonAusgestellt` ist ein eigener Event-Typ (statt einer Markierung im `BestellungAufgegeben`-Event), weil die Semantik sich grundlegend unterscheidet: Ein Freibon hat keine Varianten-ID, keinen Produktbezug und keine Kategorie. Ein eigener Typ macht die Abgrenzung im Code und in der Abrechnung explizit. Die Alternative (Freibon als Sonderfall in `BestellungAufgegeben`) wurde verworfen, weil sie die Event-Struktur unnötig verkompliziert (→ Hotspot H2 im [Event Storming](event-storming.md)).

### 3.4 Zustandsberechnung (Event Replay)

Der aktuelle Zustand eines Tisches wird nicht direkt gespeichert, sondern bei jedem Zugriff aus dem Event Stream berechnet. Der Replay-Algorithmus folgt drei Schritten:

1. **Snapshot laden** (falls vorhanden): Den letzten Snapshot für den Tisch laden (→ [3.5 Snapshot-Strategie](#35-snapshot-strategie)). Der Snapshot enthält den vollständigen Tisch-Zustand zu einem bestimmten Zeitpunkt und die zugehörige Event-Version.
2. **Events ab Snapshot-Version laden**: Alle Events des Tisches mit einer Version größer als die Snapshot-Version laden — chronologisch sortiert. Falls kein Snapshot existiert, alle Events des Tisches laden.
3. **Events sequentiell anwenden (Apply)**: Jedes Event wird auf den aktuellen Zustand angewendet:
   - `BestellungAufgegeben` → neue Bestellung mit Positionen anlegen, Saldo erhöhen
   - `FreibonAusgestellt` → neue Freibon-Bestellung mit einer Position anlegen, Saldo erhöhen
   - `ProdukteGeliefert` → referenzierte Positionen als `geliefert = true` markieren
   - `ZahlungRegistriert` → referenzierte Positionen als `bezahlt = true` markieren, Saldo reduzieren
   - `ProdukteStorniert` → referenzierte Positionen als `storniert = true` markieren, Saldo reduzieren

```
Replay-Algorithmus:

1. snapshot ← lade_snapshot(tisch_id)
2. if snapshot existiert:
       zustand ← snapshot.zustand
       ab_version ← snapshot.version
   else:
       zustand ← leerer Tisch-Zustand
       ab_version ← 0
3. events ← lade_events(tisch_id, ab_version)
4. for event in events:
       zustand ← apply(zustand, event)
5. return zustand
```

**Konsistenzgarantie:** Durch Optimistic Concurrency Control (→ [9.4 Optimistic Concurrency Control](#94-optimistic-concurrency-control)) wird sichergestellt, dass bei parallelen Zugriffen auf denselben Tisch keine Events verloren gehen. Jedes Event trägt eine aufsteigende Versionsnummer; ein Versionskonflikt beim Schreiben führt zu einem Retry.

### 3.5 Snapshot-Strategie

Snapshots sind eine rein technische Performance-Optimierung. Sie verkürzen den Event Replay, indem sie einen vorberechneten Zustand als Ausgangspunkt bereitstellen.

**Entwurfsentscheidung: Separate Speicherung (nicht im Event Stream).**

Snapshots werden **nicht** als Events im Event Stream gespeichert, sondern separat (z. B. in einer eigenen Tabelle). Gründe:

- **Fachliche Reinheit:** Der Event Stream enthält ausschließlich fachliche Domain Events. Snapshots sind technische Artefakte ohne Geschäftsbedeutung — sie beschreiben keinen Vorfall, sondern einen berechneten Zwischenstand.
- **Replay-Korrektheit:** Ein Event-Replay muss alle Events sequentiell anwenden können, ohne technische Artefakte herausfiltern zu müssen.
- **Unabhängige Lebensdauer:** Snapshots können jederzeit gelöscht und neu erstellt werden, ohne den Event Stream zu verändern. Bei einem Bug in der Snapshot-Logik kann der Snapshot verworfen und aus dem vollständigen Event Stream neu berechnet werden.

**Verworfene Alternative:** Snapshot als Event im Stream (`SnapshotErstellt`). Diese Variante wurde verworfen, weil sie fachliche und technische Concerns im selben Stream vermischt und das Event-Replay-Verfahren verkompliziert: Der Replay müsste Snapshot-Events als Sprungmarken erkennen und reguläre Events vor dem Snapshot überspringen.

**Erzeugung:** Snapshots können automatisch nach einer konfigurierbaren Anzahl von Events erstellt werden oder auf Admin-Anfrage. Für die erwartete Größenordnung (< 200 Events pro Tisch pro Veranstaltung) ist die Notwendigkeit gering — der vollständige Replay ist performant genug.

### 3.6 Policies

Policies sind automatische Reaktionen auf Domain Events oder Geschäftsregeln, die über die reine Invariantenprüfung hinausgehen.

**Stornierungsberechtigung (K-04):** Stornierungen dürfen nur durch die Rollen `senior_service` und `admin` durchgeführt werden. Diese Einschränkung ist eine bewusste organisatorische Entscheidung: Beim letzten Vereinsfest hat eine unerfahrene Servicekraft versehentlich einen ganzen Tisch storniert. Die Berechtigung wird in der Anwendungsschicht geprüft, bevor der Command an das Aggregat weitergegeben wird.

**Automatischer Bon-Druck nach Kategorie (K-11):** Wenn eine `BestellungAufgegeben` oder ein `FreibonAusgestellt` entsteht, wird automatisch ein Bon pro Kategorie an die zugeordnete Ausgabestation gesendet. Essenspositionen erzeugen einen Küchenbon, Getränkepositionen einen Thekenbon. Die Zuordnung (Kategorie → Drucker) wird in den Stammdaten konfiguriert. Details zur Ausgabe-Architektur → [5.1 Bondruck](#51-bondruck).

**Umbuchung (K-08) — Cross-Aggregat-Transaktion:** Eine Umbuchung verschiebt eine Bestellung von einem Quell-Tisch auf einen Ziel-Tisch. Fachlich setzt sie sich aus einer Stornierung am Quell-Tisch und einer neuen Bestellung am Ziel-Tisch zusammen. Da dies zwei verschiedene Aggregate betrifft, muss die Atomarität auf Anwendungsebene sichergestellt werden (z. B. über eine Datenbank-Transaktion, die beide Events in einem Schritt persistiert). Die Umbuchung ist nur für `senior_service` und `admin` zugänglich.

> **Offene Fragen (→ Hotspot H1):** Soll die Umbuchung einen eigenen Event-Typ erhalten, oder wird sie als Kombination aus `ProdukteStorniert` + `BestellungAufgegeben` mit einem Umbuchungs-Verweis modelliert? Wird der Kommentar „umgebucht von/auf Tisch X" automatisch gesetzt? → [14. Offene Entwurfsfragen](#14-offene-entwurfsfragen)

---

## 4. Stammdaten (Supporting Sub-Domain)

### 4.1 Produkt-Aggregat

Das Produkt-Aggregat verwaltet den Produktkatalog der Veranstaltung. Jedes Produkt gehört zu einer Kategorie und kann beliebig viele Varianten besitzen — jede Variante mit eigenem Namen und Preis.

```
Produkt
├── produkt_id       (UUID)
├── name             (string — nicht leer)
├── kategorie        (food | beverage | other)
├── status           (active | deleted)
└── varianten[]
    ├── variante_id  (UUID)
    ├── name         (string — nicht leer)
    ├── preis        (int, Cent — > 0)
    └── status       (active | inactive | deleted)
```

**Invarianten:**

- Produktname darf nicht leer sein.
- Kategorie muss ein gültiger Wert sein (`food`, `beverage`, `other`).
- Jede Variante benötigt einen nicht-leeren Namen und einen Preis > 0 (in Cent).
- Soft-Delete: Produkte und Varianten werden durch Status-Änderung auf `deleted` entfernt, nicht physisch gelöscht. Entfernte Produkte erscheinen nicht im Service-Produktkatalog, bleiben aber in der Datenbank erhalten — historische Bestellungen bleiben valide, weil die Events die Produktdaten zum Bestellzeitpunkt enthalten (Fat Events).
- Varianten können unabhängig vom Produkt deaktiviert werden (`inactive`). Inaktive Varianten erscheinen nicht im Service-Katalog, sind aber weiterhin in historischen Events referenziert.

### 4.2 Tisch-Stammdaten-Aggregat

Das Tisch-Stammdaten-Aggregat verwaltet die Basisdaten eines Tisches: seinen Namen und seinen Status. Es ist strikt vom Tisch-Aggregat im Kassenbetrieb (→ [3.1 Tisch-Aggregat](#31-tisch-aggregat)) zu unterscheiden.

```
Tisch (Stammdaten)
├── tisch_id    (UUID)
├── name        (string — nicht leer, z. B. „Tisch 1", „Stehtisch Eingang")
└── status      (active | inactive | deleted)
```

**Invarianten:**

- Name darf nicht leer sein.
- Soft-Delete: Tische werden durch Status-Änderung auf `deleted` entfernt. Der Datensatz bleibt in der Datenbank erhalten, damit historische Events valide bleiben.
- Nur aktive Tische (`active`) erscheinen in der Tischübersicht der Servicekräfte.

**Abgrenzung zum Kassenbetrieb:** Im Kassenbetrieb ist der Tisch ein Event-Sourced-Aggregat mit Bestellungen, Zahlungen und Saldo. In den Stammdaten ist er eine einfache CRUD-Entität mit Name und Status. Beide teilen sich die `tisch_id`, haben aber unterschiedliche Verantwortlichkeiten und unterschiedliche Persistenzstrategien.

### 4.3 Benutzer-Aggregat

Das Benutzer-Aggregat verwaltet die Zugangsdaten und Rollen der Helfer und Admins.

```
Benutzer
├── benutzer_id           (UUID)
├── name                  (string — Anzeigename)
├── benutzername          (string — eindeutig, Login-Name)
├── passwort_hash         (string — Argon2id)
├── rolle                 (admin | senior_service | service)
├── muss_passwort_setzen  (bool — true nach Erstanlage oder Passwort-Reset)
└── status                (active | inactive | deleted)
```

**Invarianten:**

- Benutzername muss systemweit eindeutig sein.
- Rolle muss ein gültiger Wert sein (`admin`, `senior_service`, `service`).
- Passwort wird mit Argon2id gehasht gespeichert — Klartext-Passwörter werden nie persistiert.
- Soft-Delete: Benutzer werden durch Status-Änderung auf `deleted` entfernt. Deaktivierte (`inactive`) und entfernte (`deleted`) Benutzer können sich nicht anmelden.
- Bei Neuanlage oder Passwort-Reset wird ein 6-stelliges Einmalpasswort generiert und `muss_passwort_setzen` auf `true` gesetzt. Bei der nächsten Anmeldung wird der Benutzer zur Passwort-Vergabe weitergeleitet (→ [7.3 Sicheres Onboarding](#73-sicheres-onboarding)).

### 4.4 Persistenzstrategie (CRUD mit Soft-Delete)

Stammdaten (Produkte, Tische, Benutzer) werden mit klassischem CRUD verwaltet — Anlegen, Lesen, Aktualisieren, Löschen. Event-Sourcing ist hier nicht nötig.

**Begründung:** Die Historie von Stammdaten-Änderungen ist fachlich irrelevant. Wenn ein Produktpreis von 350 auf 400 Cent geändert wird, interessiert nur der aktuelle Preis. Die historischen Preise stecken bereits in den Fat Events des Kassenbetrieb-Context — jede Bestellung enthält den Preis zum Bestellzeitpunkt. Dadurch ist die historische Treue gewährleistet, ohne dass Stammdaten eine Event-basierte Änderungshistorie benötigen.

**Prinzipien:**

- **Soft-Delete statt physischem Löschen:** Datensätze werden durch Status-Änderung auf `deleted` entfernt. Physisches Löschen ist nicht vorgesehen, damit referenzielle Integrität und historische Nachvollziehbarkeit erhalten bleiben.
- **Timestamps:** Alle Stammdaten tragen `erstellt_am` und `aktualisiert_am` Zeitstempel.
- **Referenzielle Integrität:** Produkte und Varianten werden nie physisch gelöscht, damit Fremdschlüssel-Referenzen aus dem Event Store valide bleiben.

---

## 5. Ausgabe (Supporting Sub-Domain)

### 5.1 Bondruck

Bons informieren Ausgabestationen (Küche, Getränketheke) über eingehende Bestellungen. Der Druck wird automatisch durch eine Policy ausgelöst.

**Policy: Automatischer Bon-Druck nach Kategorie (K-11).** Wenn ein `BestellungAufgegeben`- oder `FreibonAusgestellt`-Event entsteht, werden automatisch Bons pro Kategorie an die zugeordnete Ausgabestation gesendet:

- Essenspositionen (`food`) → Küchenbon
- Getränkepositionen (`beverage`) → Thekenbon
- Sonstige Positionen (`other`) → konfigurierbar

Die Zuordnung (Kategorie → Drucker) wird in den Stammdaten durch den Admin konfiguriert.

**Bon-Inhalt:** Tischname, Servicekraft, Positionen mit Mengen, Zeitstempel und optionaler Kommentar. Bei Freibons wird die freie Bezeichnung und der Preis angezeigt.

**Freibon als Sonderfall:** Ein Freibon hat keine Kategorie aus dem Produktkatalog. Der Bon wird an die vom Admin zugewiesene Standard-Ausgabestation gesendet, oder die Kategorie-Zuordnung wird im Freibon-Erstellungsprozess gewählt.

**Fehlerbehandlung — Fire-and-Forget mit Retry:** Der Bon-Druck darf den Bestellvorgang **nie** blockieren. Eine Bestellung wird immer gespeichert — unabhängig davon, ob der Drucker erreichbar ist. Bei Druckerausfall wird ein Retry-Mechanismus angestoßen, und die Servicekraft sieht eine Hinweismeldung.

**Nachdruck:** Einzelne Positionen oder ganze Bons können jederzeit nachgedruckt werden. Der Nachdruck ist eine reine Darstellung der bestehenden Bestelldaten — es entsteht kein neues Event.

**Offene Architekturentscheidungen:** Die konkrete Drucker-Integration (Protokoll, Hardware-Anbindung) ist ein offener Hotspot (→ [5.4 Offene Architekturentscheidungen](#54-offene-architekturentscheidungen), Hotspot H3).

### 5.2 Küchendisplay (KDS)

Das Küchendisplay (K-12) ist ein Read Model, das in Echtzeit die offenen (ungelieferten) Positionen an Ausgabestationen anzeigt — gefiltert nach Kategorie und gruppiert nach Tisch.

**Datenquelle:** `BestellungAufgegeben`- und `ProdukteGeliefert`-Events aus dem Kassenbetrieb.

**Inhalt pro Ausgabestation:**

- Offene (ungelieferte) Positionen einer Kategorie (Essen oder Getränke)
- Gruppiert nach Tisch
- Pro Position: Produkt, Variante, Menge, Zeitpunkt der Bestellung, optionaler Kommentar

**Anzeige:** Große Schrift, optimiert für Monitore in Küche und an der Getränkeausgabe. Die Filterung nach Kategorie erfolgt per URL-Parameter oder Einstellung — kein eigener Server, keine eigene App. Jede Ausgabestation sieht nur ihre relevanten Positionen.

**Akteure:** Ausgabe-Mitarbeiter (primär), Servicekräfte (Einsicht).

**Echtzeit-Aktualisierung:** Neue Bestellungen sollen mit minimaler Verzögerung auf dem Display erscheinen. Der genaue Echtzeit-Mechanismus (Polling, Server-Sent Events oder WebSockets) ist eine offene Architekturentscheidung (→ [5.4 Offene Architekturentscheidungen](#54-offene-architekturentscheidungen), Hotspot H4).

**Rückfallebene:** Das KDS dient als Ergänzung und Rückfallebene zum Bondruck — bei Bon-Verlust oder Druckerausfall können Ausgabe-Mitarbeiter die offenen Bestellungen jederzeit auf dem Display nachvollziehen.

### 5.3 Zubereitungsstatus

Aufbauend auf dem KDS (K-12) können Mitarbeiter an Ausgabestationen den Zubereitungsstatus einzelner Positionen verwalten (K-13). Servicekräfte sehen den Status und wissen, wann Positionen abholbereit sind.

**Workflow:**

```
offen → in Zubereitung → fertig
```

- **offen:** Position wurde bestellt, aber noch nicht von der Ausgabestation bearbeitet.
- **in Zubereitung:** Ausgabe-Mitarbeiter hat die Zubereitung begonnen.
- **fertig:** Position ist abholbereit. Servicekraft kann sie zum Tisch bringen.

Der Zubereitungsstatus hat **keinen Einfluss auf den Saldo** und ist für die Abrechnung irrelevant. Er ist ein rein operatives Hilfsmittel für den Ablauf zwischen Ausgabestation und Servicekraft.

**Akteure:** Ausgabe-Mitarbeiter (Status ändern), Servicekräfte (Status einsehen). Alle angemeldeten Benutzer dürfen den Status ändern — die Berechtigung ist hier nicht sicherheitskritisch, da es kein finanzieller Vorgang ist.

**Offene Modellierungsentscheidung:** Wie der Zubereitungsstatus technisch abgebildet wird, ist ein offener Hotspot (→ [5.4 Offene Architekturentscheidungen](#54-offene-architekturentscheidungen), Hotspot H5). Die drei Optionen:

- **Domain Events im Tisch-Aggregat:** Persistenter, nachvollziehbar (z. B. Analyse der Zubereitungsdauer), aber bläht den Event Stream auf.
- **Eigenes Aggregat:** Saubere Trennung vom Kassenbetrieb, eigene Invarianten möglich, aber zusätzliche Komplexität.
- **Transienter State:** Einfachste Lösung, aber der Status geht bei Seitenrefresh verloren — was den fachlichen Nutzen stark einschränkt.

### 5.4 Offene Architekturentscheidungen

Der Ausgabe-Context enthält drei offene Architekturentscheidungen, die im [Event Storming](event-storming.md) als Hotspots identifiziert wurden.

**Hotspot H3 — Drucker-Integration (K-11):**

Die konkrete Hardware-Anbindung für den Bondruck ist offen. Optionen:

| Option                | Pro                                          | Contra                                          |
| --------------------- | -------------------------------------------- | ----------------------------------------------- |
| ESC/POS über Netzwerk | Standard-Protokoll, viele kompatible Drucker | Backend muss Netzwerkdrucker direkt ansprechen  |
| WebUSB                | Direkt aus dem Browser, keine Middleware     | Eingeschränkte Browser-Unterstützung, nur lokal |
| Lokaler Print-Agent   | Flexibel, entkoppelt                         | Zusätzliche Software-Installation auf dem Gerät |

Unabhängig von der Lösung gilt: Bon-Druck darf den Bestellvorgang nie blockieren (Fire-and-Forget).

**Hotspot H4 — KDS-Echtzeit (K-12):**

Wie werden neue Bestellungen in Echtzeit an das Küchendisplay übertragen? Optionen:

| Option             | Pro                                          | Contra                                            |
| ------------------ | -------------------------------------------- | ------------------------------------------------- |
| Polling            | Einfachste Implementierung, keine neue Infra | Ineffizient — viele leere Anfragen, Latenz        |
| Server-Sent Events | Nur Server → Client, läuft über HTTP         | Keine bidirektionale Kommunikation (hier unnötig) |
| WebSockets         | Bidirektional, Echtzeit                      | Aufwändiger, bidirektional unnötig für KDS        |

Server-Sent Events sind der vielversprechendste Kompromiss: unidirektionale Push-Updates über bestehende HTTP-Infrastruktur.

**Hotspot H5 — Zubereitungsstatus-Modellierung (K-13):**

Wie wird der Zubereitungsstatus technisch modelliert? Die drei Optionen (Domain Events im Tisch-Aggregat, eigenes Aggregat, transienter State) sind in [5.3 Zubereitungsstatus](#53-zubereitungsstatus) beschrieben. Die Entscheidung hängt davon ab, ob Persistenz und Nachvollziehbarkeit oder Einfachheit priorisiert werden.

Alle drei Hotspots sind in → [14. Offene Entwurfsfragen](#14-offene-entwurfsfragen) zusammengefasst.

---

## 6. Abrechnung (Supporting Sub-Domain)

### 6.1 Read Models und Projektionen

Die Abrechnung konsumiert Events aus dem Kassenbetrieb und projiziert sie in spezialisierte Auswertungsansichten. Alle Reporting-Ansichten sind reine Read Models — sie erzeugen keine Events und verändern keinen Domänenzustand.

**Tagesabrechnung (R-01):**

- **Datenquelle:** Tisch-Events (tischübergreifend)
- **Inhalt:** Gesamtumsatz (Summe aller Zahlungen), Umsatz pro Servicekraft (Übersichtswerte), Übersicht aller Stornierungen (Zeitpunkt, Tisch, Positionen, Betrag), offene Beträge (noch nicht bezahlte Positionen)
- **Akteure:** Admin
- **Verfügbarkeit:** Jederzeit abrufbar — nicht erst beim Tagesabschluss

**Abrechnung pro Tisch (R-03):**

- **Datenquelle:** Tisch-Events (einzelner Tisch)
- **Inhalt:** Alle Bestellungen, Zahlungen, Lieferungen und Stornierungen des Tisches in chronologischer Reihenfolge, Gesamt-Saldo (bestellt, bezahlt, offen, storniert)
- **Akteure:** Admin

**Abrechnung pro Servicekraft (R-04):**

- **Datenquelle:** Tisch-Events (tischübergreifend, gruppiert nach Servicekraft)
- **Inhalt:** Umsatz pro Servicekraft (Summe registrierter Zahlungen), Anzahl aufgegebener Bestellungen, Anzahl und Betrag der Stornierungen
- **Akteure:** Admin — personenbezogene Auswertungen sind ausschließlich für den Admin zugänglich

**Produktumsatz (R-05):**

- **Datenquelle:** Tisch-Events (tischübergreifend, gruppiert nach Produkt/Variante)
- **Inhalt:** Verkaufte Menge pro Produkt und Variante (abzüglich Stornierungen), Ranking der meistverkauften Varianten, Gesamteinnahmen pro Produkt/Variante
- **Akteure:** Admin

### 6.2 Aggregationsstrategie

Die Abrechnung muss entscheiden, wie die Reporting-Daten aus den Rohdaten (Events) berechnet werden. Zwei grundsätzliche Strategien stehen zur Wahl:

**On-the-fly-Aggregation:** Bei jedem Abruf werden alle relevanten Events durchlaufen und aggregiert. Einfach zu implementieren, immer aktuell, kein zusätzlicher Speicher.

**Vorberechnete Projektionen:** Die Auswertungen werden bei jedem neuen Event inkrementell aktualisiert und stehen beim Abruf sofort bereit. Schneller beim Lesen, aber aufwändiger im Aufbau und bei Schema-Änderungen.

**Empfehlung für jotti:** On-the-fly-Aggregation ist für die erwartete Größenordnung ausreichend. Ein Vereinsfest mit 30 Tischen, je ca. 10 Bestellungen plus Zahlungen, Lieferungen und Stornierungen erzeugt ca. 500–2000 Events pro Veranstaltung. PostgreSQL kann diese Datenmenge in Millisekunden aggregieren.

Sollte sich zeigen, dass die Performance nicht ausreicht, erlaubt die Event-Sourcing-Architektur jederzeit den Wechsel zu vorberechneten Projektionen — die Events als Datenquelle bleiben unverändert.

Die endgültige Entscheidung ist als offener Hotspot dokumentiert (→ [6.5 Offene Architekturentscheidungen](#65-offene-architekturentscheidungen), Hotspot H8).

### 6.3 Tagesabschluss

Der Tagesabschluss (R-06) ist ein administrativer Prozess am Ende einer Veranstaltung, der den Betrieb formal abschließt.

**Prozess:**

1. **Offene Tische prüfen:** Das System zeigt alle Tische mit einem Saldo ≠ 0 an. Der Admin entscheidet: offene Tische erst klären, oder trotzdem abschließen.
2. **Abschlussbericht generieren:** Im Wesentlichen die Tagesabrechnung (R-01) — aber mit formalem Abschluss-Charakter.
3. **System zurücksetzen (optional):** Das System wird für die nächste Veranstaltung vorbereitet. Wie dieser Reset konkret aussieht, ist eine offene Frage.

**Umgang mit offenen Saldi:** Tische, an denen Gäste ohne Zahlung gegangen sind, müssen manuell behandelt werden — z. B. durch eine Stornierung des offenen Betrags mit Kommentar „Gast ohne Zahlung". Der Admin kann entscheiden, ob offene Beträge als Verlust verbucht werden.

**Veranstaltungskonzept:** Das Event Storming wirft die Frage auf, ob ein übergreifendes Konzept „Veranstaltung" als logischer Rahmen nötig ist — um Event-Streams einer Veranstaltung zu kapseln und den Reset zu ermöglichen, ohne Events zu löschen (Append-only-Prinzip). Alternativen: Events archivieren, einen Veranstaltungs-Marker setzen, oder pro Veranstaltung eine frische Datenbankinstanz aufsetzen.

Diese Fragen sind als offener Hotspot dokumentiert (→ [6.5 Offene Architekturentscheidungen](#65-offene-architekturentscheidungen), Hotspot H7).

### 6.4 Datenexport

Der Admin kann Veranstaltungsdaten als CSV exportieren (R-02), um sie extern weiterzuverarbeiten — insbesondere für die Vereinsbuchhaltung.

**Exportierbare Daten:**

- Umsätze (Zahlungen pro Tisch, pro Servicekraft)
- Bestellungen (alle Events mit Positionen, Mengen, Preisen, Zeitstempeln)
- Artikeldaten (verkaufte Mengen und Einnahmen pro Produkt/Variante)

**Format:** CSV — einfach, universell importierbar in Tabellenkalkulationen und Buchhaltungssoftware.

**Berechtigung:** Nur der Admin darf Daten exportieren. Der Export ist jederzeit auslösbar, nicht nur beim Tagesabschluss.

**Datenquelle:** Die exportierten Daten sind eine alternative Darstellung der bestehenden Event-Daten und Reporting-Read-Models — es entsteht kein neues Event.

### 6.5 Offene Architekturentscheidungen

Der Abrechnung-Context enthält zwei offene Architekturentscheidungen aus dem [Event Storming](event-storming.md).

**Hotspot H7 — Tagesabschluss und Veranstaltungskonzept (R-06):**

- Wie werden offene Saldi beim Tagesabschluss behandelt? (Manuelle Stornierung, Differenz-Event, Verlust-Buchung)
- Wie wird das System für die nächste Veranstaltung zurückgesetzt, ohne Events zu löschen (Append-only)?
- Braucht es ein übergreifendes Konzept „Veranstaltung" als logischen Rahmen?
- Voraussetzungen: Müssen alle Tische ausgeglichen sein, bevor ein Tagesabschluss möglich ist?

**Hotspot H8 — Reporting-Aggregation (R-01 bis R-05):**

- On-the-fly-Aggregation oder vorberechnete Projektionen?
- Reicht SQL-Aggregation über die Event-Tabelle, oder braucht es separate Reporting-Tabellen?
- Wie wird der Zeitraum für Auswertungen definiert? (pro Tag, pro Veranstaltung, frei wählbar)
- Performance-Schwellenwerte: Ab wann lohnt sich der Wechsel zu Projektionen?

Beide Hotspots sind in → [14. Offene Entwurfsfragen](#14-offene-entwurfsfragen) zusammengefasst.

---

## 7. Auth (Generic Sub-Domain)

### 7.1 Authentifizierung

Alle Benutzer melden sich mit Benutzername und Passwort an (A-01). Nach erfolgreicher Authentifizierung wird ein JWT ausgestellt, das bei jedem API-Aufruf mitgesendet und serverseitig geprüft wird.

**Passwort-Hashing:** Passwörter werden mit Argon2id gehasht — dem aktuellen Standard für sichere Passwort-Hashes. Klartext-Passwörter werden nie gespeichert oder übertragen (außer bei der Eingabe durch den Benutzer über HTTPS).

**Token:** Das JWT enthält die Benutzer-ID und die Rolle. Die Gültigkeit beträgt 12 Stunden — ein Vereinsfest dauert typischerweise einen Abend, und eine erneute Anmeldung soll innerhalb dieses Zeitraums nicht nötig sein. Abgelaufene oder ungültige Tokens führen zur automatischen Weiterleitung auf die Login-Seite.

**Generische Fehlermeldungen:** Bei ungültigen Zugangsdaten wird eine generische Fehlermeldung angezeigt — ohne Hinweis, ob Benutzername oder Passwort falsch ist. Dies erschwert das gezielte Erraten von Benutzernamen.

**Deaktivierte Benutzer:** Benutzer mit Status `inactive` oder `deleted` können sich nicht anmelden, auch wenn Benutzername und Passwort korrekt sind.

### 7.2 Autorisierung (Rollenmodell)

jotti kennt drei Rollen mit abgestuften Berechtigungen. Die Rollenprüfung erfolgt serverseitig anhand des JWT.

| Rolle              | Code-Bezeichnung | Beschreibung                                                                 |
| ------------------ | ---------------- | ---------------------------------------------------------------------------- |
| **Admin**          | `admin`          | Voller Zugriff auf Stammdaten (Produkte, Tische, Benutzer) und Kassenbetrieb |
| **Serviceleitung** | `senior_service` | Kassenbetrieb einschließlich Stornierung                                     |
| **Servicekraft**   | `service`        | Kassenbetrieb ohne Stornierung                                               |

**Berechtigungsmatrix:**

| Aktion                   | Admin | Serviceleitung | Servicekraft |
| ------------------------ | :---: | :------------: | :----------: |
| Produkte verwalten       |   ✔   |                |              |
| Tische verwalten         |   ✔   |                |              |
| Benutzer verwalten       |   ✔   |                |              |
| Passwort zurücksetzen    |   ✔   |                |              |
| Bestellung aufgeben      |   ✔   |       ✔        |      ✔       |
| Lieferung bestätigen     |   ✔   |       ✔        |      ✔       |
| Zahlung registrieren     |   ✔   |       ✔        |      ✔       |
| Stornierung durchführen  |   ✔   |       ✔        |              |
| Tischübersicht einsehen  |   ✔   |       ✔        |      ✔       |
| Kassenjournal einsehen   |   ✔   |       ✔        |      ✔       |
| Tagesabrechnung einsehen |   ✔   |                |              |
| Datenexport              |   ✔   |                |              |
| Tagesabschluss einleiten |   ✔   |                |              |
| Abmelden                 |   ✔   |       ✔        |      ✔       |

Die Rollenhierarchie ist inklusiv: Admin kann alles, was Serviceleitung kann. Serviceleitung kann alles, was Servicekraft kann — plus Stornierung.

### 7.3 Sicheres Onboarding

Neue Benutzer durchlaufen einen zweistufigen Onboarding-Prozess (A-02), der sicherstellt, dass nur der Benutzer sein eigenes Passwort kennt.

**Ablauf:**

1. **Benutzer anlegen:** Der Admin erstellt einen Benutzer mit Name, Benutzername und Rolle. Das System generiert ein 6-stelliges Einmalpasswort, das der Admin dem Benutzer mitteilt (z. B. mündlich oder auf einem Zettel).
2. **Erstanmeldung:** Der Benutzer meldet sich mit Benutzername und Einmalpasswort an. Das System erkennt `muss_passwort_setzen = true` und leitet automatisch zur Seite „Passwort setzen" weiter.
3. **Eigenes Passwort setzen:** Der Benutzer vergibt ein eigenes Passwort (min. 8 Zeichen). Das neue Passwort wird mit Argon2id gehasht gespeichert. `muss_passwort_setzen` wird auf `false` gesetzt.
4. **Normale Anmeldung:** Ab jetzt meldet sich der Benutzer mit seinem selbst gewählten Passwort an.

**Passwort-Reset:** Wenn der Admin ein Passwort zurücksetzt, wird ein neues 6-stelliges Einmalpasswort generiert und `muss_passwort_setzen` wieder auf `true` gesetzt. Der Benutzer durchläuft beim nächsten Login erneut Schritt 2 und 3.

Das Einmalpasswort ist bewusst kurz (6 Zeichen) — es dient nur der einmaligen Identifikation und wird sofort durch ein sicheres Passwort ersetzt.

---

## 8. Read Models

Read Models sind aufbereitete Lese-Ansichten, die aus den Domain Events oder den Stammdaten zusammengebaut werden. Sie enthalten genau die Informationen, die ein bestimmter Akteur in einer bestimmten Situation braucht. Read Models werden nicht geschrieben — sie sind reine Projektionen über vorhandene Daten.

### 8.1 Service-Ansichten

#### Tischübersicht (K-05)

| Eigenschaft  | Beschreibung                                                      |
| ------------ | ----------------------------------------------------------------- |
| **Quelle**   | Tisch-Events + Tisch-Stammdaten                                   |
| **Inhalt**   | Pro aktivem Tisch: Name, aktueller Saldo, Anzahl unbezahlter Positionen, Anzahl ungelieferter Positionen |
| **Anzeige**  | Karten-Layout, alle aktiven Tische; Schnellsuche nach Tischname (K-10) |
| **Akteure**  | Servicekraft, Serviceleitung, Admin                               |

Die Tischübersicht ist die Startseite des Service-Bereichs. Auf einen Blick sieht die Servicekraft, welche Tische aktiv sind und wo noch offene Bestellungen oder ausstehende Zahlungen vorliegen.

#### Tischdetails (K-05)

| Eigenschaft  | Beschreibung                                                      |
| ------------ | ----------------------------------------------------------------- |
| **Quelle**   | Tisch-Events (Event-Stream des jeweiligen Tisches)                |
| **Inhalt**   | Alle Positionen mit Status (bestellt / geliefert / bezahlt / storniert), gruppiert nach Bestellung (mit opt. Bezeichnung aus K-07), aktueller Saldo, unbezahlte Positionen, ungelieferte Positionen |
| **Anzeige**  | Tabs: Übersicht, Bestellen, Liefern, Bezahlen, Stornieren, Historie |
| **Akteure**  | Servicekraft, Serviceleitung, Admin                               |

Die Tischdetail-Ansicht ist der zentrale Arbeitsplatz der Servicekraft. Alle Tischoperationen (Bestellung aufgeben, Lieferung bestätigen, Zahlung registrieren, Stornierung) werden als Drawer von dieser Ansicht aus geöffnet.

#### Produktkatalog

| Eigenschaft  | Beschreibung                                                      |
| ------------ | ----------------------------------------------------------------- |
| **Quelle**   | Produkt-Stammdaten                                                |
| **Inhalt**   | Alle aktiven Produkte mit ihren aktiven Varianten, nach Kategorie gruppiert (Speisen, Getränke, Sonstiges), jeweils mit Name und Preis |
| **Anzeige**  | Kategorie-Tabs, Produkte als auswählbare Karten, Plus/Minus für Mengenauswahl |
| **Akteure**  | Servicekraft, Serviceleitung, Admin (beim Bestellen)              |

Der Produktkatalog ist kein eigenständiges Navigations-Ziel, sondern wird im Kontext des Bestellvorgangs (Tab „Bestellen" im Tischdetail) geladen. Er zeigt immer den aktuellen Stand der Stammdaten.

#### Kassenjournal (K-06)

| Eigenschaft  | Beschreibung                                                      |
| ------------ | ----------------------------------------------------------------- |
| **Quelle**   | Tisch-Events (Event-Stream des jeweiligen Tisches)                |
| **Inhalt**   | Chronologische Liste aller Vorgänge am Tisch: Zeitstempel, Typ (Bestellung / Lieferung / Zahlung / Stornierung / Freibon), Positionen, Betrag, Servicekraft, Kommentar |
| **Anzeige**  | Timeline / Liste im Tab „Historie" der Tischdetail-Ansicht        |
| **Akteure**  | Servicekraft, Serviceleitung, Admin                               |

Das Kassenjournal ist die menschenlesbare Darstellung des Event-Streams. Es ist unveränderlich — jeder Vorgang am Tisch erscheint hier in der Reihenfolge, in der er eingetreten ist.

### 8.2 Admin-Ansichten (Reporting)

Alle Reporting-Ansichten sind Read Models, die aus den Tisch-Events über alle Tische hinweg aggregiert werden. Sie sind nur für Admins zugänglich.

#### Tagesabrechnung (R-01)

| Eigenschaft  | Beschreibung                                                      |
| ------------ | ----------------------------------------------------------------- |
| **Quelle**   | Tisch-Events (tischübergreifend)                                  |
| **Inhalt**   | Gesamtumsatz (Summe aller Zahlungen), Umsatz pro Servicekraft (Übersichtswerte), Übersicht aller Stornierungen (Zeitpunkt, Tisch, Positionen, Betrag), offene Beträge |
| **Akteure**  | Admin                                                             |

#### Abrechnung pro Tisch (R-03)

| Eigenschaft  | Beschreibung                                                      |
| ------------ | ----------------------------------------------------------------- |
| **Quelle**   | Tisch-Events (einzelner Tisch)                                    |
| **Inhalt**   | Alle Bestellungen, Zahlungen, Lieferungen, Stornierungen in chronologischer Reihenfolge; Gesamt-Saldo (bestellt, bezahlt, offen, storniert) |
| **Akteure**  | Admin                                                             |

#### Abrechnung pro Servicekraft (R-04)

| Eigenschaft  | Beschreibung                                                      |
| ------------ | ----------------------------------------------------------------- |
| **Quelle**   | Tisch-Events (tischübergreifend, gruppiert nach Servicekraft)     |
| **Inhalt**   | Umsatz pro Servicekraft (Summe registrierter Zahlungen), Anzahl aufgegebener Bestellungen, Anzahl und Betrag der Stornierungen |
| **Akteure**  | Admin                                                             |

#### Produktumsatz (R-05)

| Eigenschaft  | Beschreibung                                                      |
| ------------ | ----------------------------------------------------------------- |
| **Quelle**   | Tisch-Events (tischübergreifend, gruppiert nach Produkt/Variante) |
| **Inhalt**   | Verkaufte Menge pro Produkt und Variante (abzüglich Stornierungen), Ranking der meistverkauften Varianten, Gesamteinnahmen pro Produkt/Variante |
| **Akteure**  | Admin                                                             |

### 8.3 Ausgabe-Ansichten

#### KDS-Ansicht (K-12)

| Eigenschaft  | Beschreibung                                                      |
| ------------ | ----------------------------------------------------------------- |
| **Quelle**   | Tisch-Events (`BestellungAufgegeben`), gefiltert nach Kategorie   |
| **Inhalt**   | Offene (ungelieferte) Positionen einer Kategorie (Speisen oder Getränke), gruppiert nach Tisch; pro Position: Produkt, Variante, Menge, Zeitpunkt der Bestellung |
| **Anzeige**  | Echtzeit-Updates, große Schrift (für Monitore in Küche/Ausgabe)   |
| **Akteure**  | Ausgabe-Mitarbeiter, Servicekraft (lesend)                        |

Jede Ausgabestation sieht nur die Positionen ihrer eigenen Kategorie. Die technische Umsetzung der Echtzeit-Aktualisierung ist offen (→ H4).

#### Zubereitungsstatus (K-13)

| Eigenschaft  | Beschreibung                                                      |
| ------------ | ----------------------------------------------------------------- |
| **Quelle**   | Tisch-Events (`BestellungAufgegeben`) + Zubereitungsstatus-Daten  |
| **Inhalt**   | Offene Positionen mit Status: „offen" → „in Zubereitung" → „fertig", gruppiert nach Tisch, Zeitpunkt des letzten Statuswechsels |
| **Anzeige**  | Farbcodierung nach Status; auf KDS und Servicekraft-Ansicht       |
| **Akteure**  | Ausgabe-Mitarbeiter (Status ändern), Servicekraft (Status einsehen) |

Die Modellierung des Zubereitungsstatus ist offen: Domain Events im Tisch-Aggregat, eigenes Aggregat oder transienter State sind alle denkbare Optionen (→ H5).

---

## 9. Persistenzstrategie

### 9.1 Zwei Strategien, eine Datenbank

jotti verwendet zwei Persistenzstrategien in einer einzigen PostgreSQL-Datenbank:

| Bereich                    | Strategie       | Begründung                                                  |
| -------------------------- | --------------- | ----------------------------------------------------------- |
| Kassenbetrieb (Tisch)      | Event-Sourcing  | Geschichte ist fachlich relevant (Kassenjournal, Buchhaltung); lückenlose Nachvollziehbarkeit ist ein Kernziel |
| Stammdaten (Produkt, Tisch, Benutzer) | CRUD | Nur aktueller Zustand benötigt; historische Stammdaten-Änderungen irrelevant (Fat Events decken das ab) |
| Auth                       | CRUD            | Infrastruktur ohne fachliche Event-Semantik                 |

Die Trennung ist klar: Überall dort, wo die Geschichte eines Objekts fachlich relevant ist, wird Event-Sourcing eingesetzt. Für reine Konfigurationsdaten reicht klassisches CRUD.

### 9.2 Event Store

Der Event Store ist das Herzstück des Kassenbetriebs. Er speichert alle Domain Events als unveränderliche, append-only Einträge.

**Prinzipien:**

- **Append-only:** Events werden niemals geändert oder gelöscht. Jeder Geschäftsvorfall hinterlässt einen dauerhaften Eintrag.
- **Subjekt-Format:** Events sind einem Tisch zugeordnet, Schlüssel `tisch:<id>`.
- **Optimistic Concurrency Control:** Jedes Event trägt eine `version`-Nummer. Die Kombination `(tisch_id, version)` ist eindeutig (UNIQUE Constraint in der Datenbank). Ein Schreibversuch mit einer bereits vergebenen Version schlägt fehl — so werden parallele Schreibkonflikte erkannt.
- **Keine Lösch-Operation:** Selbst fehlerhafte Events bleiben erhalten. Korrekturen erfolgen durch Kompensations-Events (z. B. `ProdukteStorniert` als Reaktion auf eine fehlerhafte `BestellungAufgegeben`).

**Snapshot-Speicherung:**

Snapshots sind rein technische Optimierungen ohne fachliche Bedeutung. Sie werden **separat** vom Event Stream gespeichert — in einer eigenen Tabelle oder als markierter Datensatz — und nicht als Event in den Stream geschrieben. Snapshots beschleunigen das Laden langer Event-Streams: Statt den gesamten Stream von Anfang an zu replizieren, wird der letzte Snapshot geladen und nur die darauf folgenden Events angewendet.

Snapshots werden automatisch nach einer konfigurierbaren Anzahl neuer Events oder auf Admin-Anfrage erstellt. Die Wahrheit bleibt immer der Event-Stream — ein Snapshot ist jederzeit aus dem Stream reproduzierbar.

_Verworfene Alternative: `SnapshotErstellt` als Event im Stream zu speichern würde fachliche und technische Concerns vermischen und den Stream unübersichtlicher machen._

### 9.3 Stammdaten (CRUD-Prinzipien)

Produkte, Tische (Stammdaten) und Benutzer werden mit klassischem CRUD verwaltet. Dabei gelten folgende Prinzipien:

- **Soft-Delete:** Entitäten werden nie physisch gelöscht, sondern auf `status = 'deleted'` gesetzt. Dadurch bleiben historische Referenzen in den Tisch-Events gültig. Deaktivierte Produkte erscheinen nicht mehr im Bestellvorgang, aber ihre Daten sind noch nachvollziehbar.
- **Timestamps:** Jede Entität trägt Erstellungs- und Änderungszeitpunkte.
- **Referenzielle Integrität:** Varianten gehören zu einem Produkt. Wenn ein Produkt gelöscht (Soft-Delete) wird, werden seine Varianten ebenfalls deaktiviert.
- **Fat Events decken Preishistorie ab:** Da die `BestellungAufgegeben`-Events den Produktnamen, den Variantennamen und den Einzelpreis zum Zeitpunkt der Bestellung einbetten, ist eine separate Preishistorie für Stammdaten nicht nötig.

### 9.4 Optimistic Concurrency Control

Mehrere Servicekräfte können gleichzeitig an verschiedenen Tischen arbeiten — dabei entstehen keine Konflikte. Arbeiten jedoch zwei Servicekräfte gleichzeitig am **selben** Tisch (z. B. zwei Zahlungsvorgänge parallel), kann ein Schreibkonflikt entstehen.

**Mechanismus:**

1. Beim Laden eines Tisches wird die aktuelle `event_version` mitgegeben.
2. Beim Schreiben eines neuen Events wird die erwartete Version mitgeschickt.
3. Die Datenbank prüft via UNIQUE Constraint `(tisch_id, version)`, ob die Version noch frei ist.
4. Ist die Version bereits vergeben (ein anderer Schreibvorgang war schneller), schlägt die Operation mit einem Konflikt-Fehler fehl.
5. Die Anwendungsschicht führt einen **Retry** durch: Tischzustand neu laden, Operation erneut anwenden, neuen Schreibversuch starten.

Dieser Mechanismus stellt sicher, dass der Tisch-Zustand immer konsistent ist, ohne Datenbankzeilen sperren zu müssen. Der Retry ist für die meisten Konfliktfälle bei einer Veranstaltung ausreichend — echte Konkurrenz am selben Tisch ist die Ausnahme.

---

## 10. Architekturprinzipien

### 10.1 Schichtenarchitektur

<!-- Abschnitt 5 -->

### 10.2 API-Design-Prinzipien

<!-- Abschnitt 5 -->

### 10.3 Frontend-Architektur

<!-- Abschnitt 5 -->

### 10.4 Validierung

<!-- Abschnitt 5 -->

### 10.5 Geldbeträge

<!-- Abschnitt 5 -->

### 10.6 Mehrbenutzerfähigkeit

<!-- Abschnitt 5 -->

### 10.7 Mobile-first

<!-- Abschnitt 5 -->

### 10.8 Sicherheit

<!-- Abschnitt 5 -->

---

## 11. Infrastruktur und Deployment

### 11.1 Architekturüberblick

<!-- Abschnitt 5 -->

### 11.2 Deployment-Modell

<!-- Abschnitt 5 -->

---

## 12. Ubiquitous Language

<!-- Abschnitt 6 -->

---

## 13. Priorisierung und Ausbaustufen

### 13.1 Stufe 1 — Must-have (MVP)

<!-- Abschnitt 6 -->

### 13.2 Stufe 2 — Should-have

<!-- Abschnitt 6 -->

### 13.3 Stufe 3 — Nice-to-have

<!-- Abschnitt 6 -->

---

## 14. Offene Entwurfsfragen

<!-- Abschnitt 6 -->
