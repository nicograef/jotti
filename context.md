# Ubiquitous Language Migration — Vollständiger Kontext & Plan

> **Zweck:** Dieses Dokument enthält den kompletten Kontext und Ausführungsplan, um die Domänensprache von Englisch auf Deutsch umzustellen. Gedacht als Übergabedokument an einen KI-Coding-Agenten in einer neuen Session.

---

## 1. Ziel

Die gesamte Domänensprache des Projekts von Englisch auf Deutsch umstellen (Ubiquitous Language im DDD-Sinne). Englische Wörter bleiben im Code (Funktionsnamen wie `Get`, `Create`, Variablen wie `err`, `ctx`) — nur die **Fachbegriffe der Domäne** werden deutsch.

**Motivation:** Die Software ist ausschließlich für den deutschen Markt. Der Entwickler möchte die gleiche Fachsprache sprechen wie die Vereine/Kunden.

---

## 2. Zentrale Design-Entscheidung: Varianten vs. Produkte

### Problem

Die Servicekräfte im Verein sprechen von „Produkten", nicht von „Varianten". Im Admin gibt es ein Produkt-Varianten-Modell (Produkt = Cola, Varianten = 0.3L, 0.5L), aber im Service-Kontext bestellt man einfach „Produkte".

### Lösung — Zwei Bounded Contexts, eine Sprache

- **Admin-Kontext (Stammdaten):** `Produkt` hat `Varianten` — das Modell bleibt so
- **Service-Kontext (Kassenbetrieb):** Was bisher `LineItem`/`Variant` heißt, wird zu `Position`
  - Im Service sieht man nur die Varianten als flache Liste von Positionen
  - `LineItem` (struct) → `Position`
  - `variants` (JSON-Feld in Events/Requests) → `positionen`
  - `unpaidVariants` → `unbezahltePositionen`
  - `undeliveredVariants` → `ungeliefertePositionen`
  - Event-Typen: `variants-canceled` → `produkte-storniert`, `variants-delivered` → `produkte-geliefert`

---

## 3. Vollständige Begriffstabelle

### Domain-Structs / Types

| Englisch (alt)    | Deutsch (neu) | Kontext                                                     |
| ----------------- | ------------- | ----------------------------------------------------------- |
| `Order`           | `Bestellung`  | Struct, Type, Event-Daten                                   |
| `Payment`         | `Zahlung`     | Struct, Type, Event-Daten                                   |
| `Delivery`        | `Lieferung`   | Struct, Type, Event-Daten                                   |
| `Cancelation`     | `Stornierung` | Struct, Type, Event-Daten                                   |
| `LineItem`        | `Position`    | Struct, die eine Zeile in Bestellung/Zahlung/etc. darstellt |
| `Snapshot`        | `Snapshot`    | Bleibt (technisch, kein Fachbegriff)                        |
| `Table` (Domain)  | `Tisch`       | Domain-Struct. DB-Tabelle bleibt `tables`                   |
| `Variant` (Admin) | `Variante`    | Bleibt im Admin-Kontext                                     |

### Event-Types

| Alt                           | Neu                              |
| ----------------------------- | -------------------------------- |
| `table.order-placed:v1`       | `tisch.bestellung-aufgegeben:v1` |
| `table.payment-registered:v1` | `tisch.zahlung-registriert:v1`   |
| `table.variants-canceled:v1`  | `tisch.produkte-storniert:v1`    |
| `table.variants-delivered:v1` | `tisch.produkte-geliefert:v1`    |
| `table.snapshot:v1`           | `tisch.snapshot:v1`              |

### Event Subject

| Alt          | Neu          |
| ------------ | ------------ |
| `table:<id>` | `tisch:<id>` |

### JSON-Felder (API-Kontrakt zwischen Frontend & Backend)

| Alt                             | Neu                      | Kontext                                     |
| ------------------------------- | ------------------------ | ------------------------------------------- |
| `orderId`                       | `bestellungId`           | Event-Daten                                 |
| `paymentId`                     | `zahlungId`              | Event-Daten                                 |
| `deliveryId`                    | `lieferungId`            | Event-Daten                                 |
| `cancelationId`                 | `stornierungId`          | Event-Daten                                 |
| `totalPriceCents`               | `gesamtPreisCents`       | Bestellung                                  |
| `totalPaymentCents`             | `gesamtZahlungCents`     | Zahlung                                     |
| `totalCancelationCents`         | `gesamtStornierungCents` | Stornierung                                 |
| `variants` (in events/requests) | `positionen`             | Events und API-Requests im Service          |
| `placedAt`                      | `aufgegebenAm`           | Bestellung-Zeitstempel                      |
| `registeredAt`                  | `registriertAm`          | Zahlung-Zeitstempel                         |
| `deliveredAt`                   | `geliefertAm`            | Lieferung-Zeitstempel                       |
| `canceledAt`                    | `storniertAm`            | Stornierung-Zeitstempel                     |
| `balanceCents`                  | `saldoCents`             | Saldo-Query-Response                        |
| `unpaidVariants`                | `unbezahltePositionen`   | Snapshot-Daten / Query                      |
| `undeliveredVariants`           | `ungeliefertePositionen` | Snapshot-Daten / Query                      |
| `totalPaymentsCents`            | `gesamtZahlungenCents`   | Snapshot                                    |
| `priceCents`                    | `preisCents`             | Position (LineItem)                         |
| `tableId`                       | `tischId`                | Alle Requests die einen Tisch referenzieren |

### API-Endpunkte

| Alt                                            | Neu                                        |
| ---------------------------------------------- | ------------------------------------------ |
| `POST /service/place-table-order`              | `POST /service/bestellung-aufgeben`        |
| `POST /service/register-table-payment`         | `POST /service/zahlung-registrieren`       |
| `POST /service/deliver-table-variants`         | `POST /service/produkte-liefern`           |
| `POST /senior_service/cancel-table-variants`   | `POST /senior_service/produkte-stornieren` |
| `POST /service/get-table`                      | `POST /service/get-tisch`                  |
| `POST /service/get-active-tables`              | `POST /service/get-aktive-tische`          |
| `POST /service/get-table-history`              | `POST /service/get-tisch-historie`         |
| `POST /service/get-table-balance`              | `POST /service/get-tisch-saldo`            |
| `POST /service/get-table-unpaid-variants`      | `POST /service/get-tisch-unbezahlt`        |
| `POST /service/get-table-undelivered-variants` | `POST /service/get-tisch-ungeliefert`      |
| `POST /service/get-active-products`            | `POST /service/get-aktive-produkte`        |
| `POST /admin/create-product`                   | `POST /admin/create-produkt`               |
| `POST /admin/update-product`                   | `POST /admin/update-produkt`               |
| `POST /admin/create-variant`                   | `POST /admin/create-variante`              |
| `POST /admin/update-variant`                   | `POST /admin/update-variante`              |
| `POST /admin/activate-variant`                 | `POST /admin/activate-variante`            |
| `POST /admin/deactivate-variant`               | `POST /admin/deactivate-variante`          |
| `POST /admin/get-all-products`                 | `POST /admin/get-all-produkte`             |
| `POST /admin/create-table`                     | `POST /admin/create-tisch`                 |
| `POST /admin/update-table`                     | `POST /admin/update-tisch`                 |
| `POST /admin/activate-table`                   | `POST /admin/activate-tisch`               |
| `POST /admin/deactivate-table`                 | `POST /admin/deactivate-tisch`             |
| `POST /admin/get-all-tables`                   | `POST /admin/get-all-tische`               |

---

## 4. Breaking Change Strategie

**Gewählt: Breaking Change (Option B).**

Das Projekt ist pre-production. Eine DB-Migration schreibt alle bestehenden Event-Typen und Subjects in der `events`-Tabelle um. Keine Rückwärtskompatibilität nötig.

---

## 5. Betroffene Dateien — Detailierte Aufstellung

### 5.1 Database Migration (NEU ERSTELLEN)

**`database/migrations/03_german_ubiquitous_language.up.sql`** — Erstellen:

```sql
BEGIN;
-- Event-Typen umbenennen
UPDATE events SET type = 'tisch.bestellung-aufgegeben:v1' WHERE type = 'table.order-placed:v1';
UPDATE events SET type = 'tisch.zahlung-registriert:v1' WHERE type = 'table.payment-registered:v1';
UPDATE events SET type = 'tisch.produkte-storniert:v1' WHERE type = 'table.variants-canceled:v1';
UPDATE events SET type = 'tisch.produkte-geliefert:v1' WHERE type = 'table.variants-delivered:v1';
UPDATE events SET type = 'tisch.snapshot:v1' WHERE type = 'table.snapshot:v1';
-- Subject-Format umbenennen
UPDATE events SET subject = REPLACE(subject, 'table:', 'tisch:');
COMMIT;
```

**`database/migrations/03_german_ubiquitous_language.down.sql`** — Erstellen (Rollback).

### 5.2 Backend: `domain/table/` — Go-Structs und Events

Alle Dateien sind im Go-Package `table` unter `backend/domain/table/`. Das **Package** heißt weiterhin `table` (Go-Konvention: Verzeichnisname = Package).

**Dateien umbenennen (mv) + Inhalt ändern:**

| Alte Datei                  | Neue Datei                     | Was ändern                                                 |
| --------------------------- | ------------------------------ | ---------------------------------------------------------- |
| `order.go`                  | `bestellung.go`                | `Order` → `Bestellung`, `LineItem` → `Position`, JSON-Tags |
| `payment.go`                | `zahlung.go`                   | `Payment` → `Zahlung`, JSON-Tags                           |
| `delivery.go`               | `lieferung.go`                 | `Delivery` → `Lieferung`, JSON-Tags                        |
| `cancelation.go`            | `stornierung.go`               | `Cancelation` → `Stornierung`, JSON-Tags                   |
| `table.go`                  | `tisch.go`                     | `Table` → `Tisch`-Struct, Validierung                      |
| `orderPlacedEvent.go`       | `bestellungAufgegebenEvent.go` | Event-Constructor + Builder                                |
| `paymentRegisteredEvent.go` | `zahlungRegistriertEvent.go`   | Event-Constructor + Builder                                |
| `productsCanceledEvent.go`  | `produkteStorniertEvent.go`    | Event-Constructor + Builder                                |
| `productsDeliveredEvent.go` | `produkteGeliefertEvent.go`    | Event-Constructor + Builder                                |
| `snapshotEvent.go`          | `snapshotEvent.go` (bleibt)    | JSON-Tags für unbezahlt/ungeliefert                        |
| `events.go`                 | `events.go` (bleibt)           | EventType-Konstanten, alle Funktionsnamen                  |
| `events_test.go`            | `events_test.go` (bleibt)      | Test-Code anpassen                                         |

**Konkrete Struct-Umbenennungen im Domain-Layer:**

```go
// bestellung.go (war order.go)
type Position struct {           // war: LineItem
    ID         int    `json:"id"`
    Name       string `json:"name"`
    PreisCents int    `json:"preisCents"`    // war: priceCents → preisCents
    Quantity   int    `json:"quantity"`
}

type Bestellung struct {         // war: Order
    ID              string     `json:"id"`
    UserID          int        `json:"userId"`
    TischID         int        `json:"tischId"`          // war: tableId
    Positionen      []Position `json:"positionen"`       // war: variants
    GesamtPreisCents int       `json:"gesamtPreisCents"` // war: totalPriceCents
    Comment         string     `json:"comment"`
    AufgegebenAm    time.Time  `json:"aufgegebenAm"`     // war: placedAt
}

// zahlung.go (war payment.go)
type Zahlung struct {            // war: Payment
    ID                  string     `json:"id"`
    UserID              int        `json:"userId"`
    TischID             int        `json:"tischId"`
    Positionen          []Position `json:"positionen"`
    GesamtZahlungCents  int        `json:"gesamtZahlungCents"`  // war: totalPaymentCents
    Comment             string     `json:"comment"`
    RegistriertAm       time.Time  `json:"registriertAm"`       // war: registeredAt
}

// lieferung.go (war delivery.go)
type Lieferung struct {          // war: Delivery
    ID          string     `json:"id"`
    UserID      int        `json:"userId"`
    TischID     int        `json:"tischId"`
    Positionen  []Position `json:"positionen"`
    Comment     string     `json:"comment"`
    GeliefertAm time.Time  `json:"geliefertAm"`     // war: deliveredAt
}

// stornierung.go (war cancelation.go)
type Stornierung struct {        // war: Cancelation
    ID                      string     `json:"id"`
    UserID                  int        `json:"userId"`
    TischID                 int        `json:"tischId"`
    Positionen              []Position `json:"positionen"`
    GesamtStornierungCents  int        `json:"gesamtStornierungCents"` // war: totalCancelationCents
    Comment                 string     `json:"comment"`
    StorniertAm             time.Time  `json:"storniertAm"`           // war: canceledAt
}

// tisch.go (war table.go)
type Tisch struct {              // war: Table
    ID        int       `json:"id"`
    Name      string    `json:"name"`
    Status    Status    `json:"status"`
    CreatedAt time.Time `json:"createdAt"`
}

// snapshotEvent.go — Snapshot-Felder:
type snapshotV1Data struct {
    SaldoCents              int        `json:"saldoCents"`              // war: balanceCents
    UnbezahltePositionen    []Position `json:"unbezahltePositionen"`    // war: unpaidVariants
    UngeliefertePositionen  []Position `json:"ungeliefertePositionen"`  // war: undeliveredVariants
    GesamtZahlungenCents    int        `json:"gesamtZahlungenCents"`    // war: totalPaymentsCents
}

type Snapshot struct {
    TischID                 int
    SaldoCents              int
    UnbezahltePositionen    []Position
    UngeliefertePositionen  []Position
    GesamtZahlungenCents    int
    CreatedAt               time.Time
}
```

**Event-Type Konstanten (`events.go`):**

```go
const (
    EventTypeBestellungAufgegebenV1 EventType = "tisch.bestellung-aufgegeben:v1"
    EventTypeZahlungRegistriertV1   EventType = "tisch.zahlung-registriert:v1"
    EventTypeProdukteStorniertV1    EventType = "tisch.produkte-storniert:v1"
    EventTypeProdukteGeliefertV1    EventType = "tisch.produkte-geliefert:v1"
    EventTypeSnapshotV1             EventType = "tisch.snapshot:v1"
)
```

**Event-Daten-Structs (intern in Event-Dateien):**

```go
// bestellungAufgegebenEvent.go
type bestellungAufgegebenV1Data struct {
    BestellungID    string     `json:"bestellungId"`
    Positionen      []Position `json:"positionen"`
    GesamtPreisCents int       `json:"gesamtPreisCents"`
    Comment         string     `json:"comment"`
}

// zahlungRegistriertEvent.go
type zahlungRegistriertV1Data struct {
    ZahlungID          string     `json:"zahlungId"`
    Positionen         []Position `json:"positionen"`
    GesamtZahlungCents int        `json:"gesamtZahlungCents"`
    Comment            string     `json:"comment"`
}

// produkteStorniertEvent.go
type produkteStorniertV1Data struct {
    StornierungID          string     `json:"stornierungId"`
    Positionen             []Position `json:"positionen"`
    GesamtStornierungCents int        `json:"gesamtStornierungCents"`
    Comment                string     `json:"comment"`
}

// produkteGeliefertEvent.go
type produkteGeliefertV1Data struct {
    LieferungID string     `json:"lieferungId"`
    Positionen  []Position `json:"positionen"`
    Comment     string     `json:"comment"`
}
```

**Funktionsnamen (`events.go`):**

```go
// Event-Constructors:
NewBestellungAufgegebenEvent()     // war: NewOrderPlacedEvent
NewZahlungRegistriertEvent()       // war: NewPaymentRegisteredEvent
NewProdukteStorniertEvent()        // war: NewVariantsCanceledEvent
NewProdukteGeliefertEvent()        // war: NewVariantsDeliveredEvent
NewSnapshotEvent()                 // bleibt

// Event-Builders (intern):
buildBestellungFromEvent()         // war: buildOrderFromEvent
buildZahlungFromEvent()            // war: buildPaymentFromEvent
buildStornierungFromEvent()        // war: buildCancelationFromEvent
buildLieferungFromEvent()          // war: buildDeliveryFromEvent
buildSnapshotFromEvent()           // bleibt

// State-Rekonstruktion:
GetSaldoFromEvents()               // war: GetBalanceFromEvents
GetHistoryFromEvents()             // bleibt (generisch)
GetUnbezahltePositionenFromEvents()  // war: GetUnpaidVariantsFromEvents
GetUngeliefertePositionenFromEvents() // war: GetUndeliveredVariantsFromEvents
GetGesamtZahlungenFromEvents()     // war: GetTotalPaymentsFromEvents

// Hilfsfunktionen:
accumulatePositionen()             // war: accumulateVariants
reducePositionen()                 // war: reduceVariants
parseTischIDFromSubject()          // war: parseTableIDFromSubject (prefix "tisch:")
```

**Tisch-Struct (`tisch.go`, war `table.go`):**

```go
type Tisch struct { ... }
func NewTisch(name string) (Tisch, error) { ... }
func (t *Tisch) Activate() { ... }
func (t *Tisch) Deactivate() { ... }
func (t *Tisch) Rename(newName string) error { ... }
// Schemas: TischIDSchema, TischNameSchema, TischStatusSchema, TischSchema
```

### 5.3 Backend: `api/table/application/` — Commands und Queries

**`command.go`:**

```go
// Interface-Methoden:
PlaceTableOrder → BestellungAufgeben
RegisterTablePayment → ZahlungRegistrieren
CancelTableVariants → ProdukteStornieren
DeliverTableVariants → ProdukteLiefern
CreateTableSnapshot → TischSnapshotErstellen
CreateTable → TischErstellen
UpdateTable → TischAktualisieren
ActivateTable → TischAktivieren
DeactivateTable → TischDeaktivieren

// Parameter-Typ:
variants []table.LineItem → positionen []table.Position
```

**`query.go`:**

```go
GetTable → GetTisch
GetAllTables → GetAllTische
GetActiveTables → GetAktiveTische
GetTableBalance → GetTischSaldo
GetTableHistory → GetTischHistorie
GetTableUnpaidVariants → GetTischUnbezahlt
GetTableUndeliveredVariants → GetTischUngeliefert
```

**`errors.go`:**

```go
ErrTableNotFound → ErrTischNotFound
ErrTableAlreadyExists → ErrTischAlreadyExists
ErrInvalidTableData → ErrInvalidTischData
```

### 5.4 Backend: `api/table/http/` — Handler

**`command_handler.go`:**

```go
// Handler-Methoden:
PlaceTableOrderHandler → BestellungAufgebenHandler
RegisterTablePaymentHandler → ZahlungRegistrierenHandler
CancelTableVariantsHandler → ProdukteStornierenHandler
DeliverTableVariantsHandler → ProdukteLiefernHandler
CreateTableHandler → TischErstellenHandler
UpdateTableHandler → TischAktualisierenHandler
ActivateTableHandler → TischAktivierenHandler
DeactivateTableHandler → TischDeaktivierenHandler

// Request-Struct JSON-Tags:
"tableId" → "tischId"
"variants" → "positionen"
```

**`query_handler.go`:**

```go
GetTableHandler → GetTischHandler
GetAllTablesHandler → GetAllTischeHandler
GetActiveTablesHandler → GetAktiveTischeHandler
GetTableHistoryHandler → GetTischHistorieHandler
GetTableBalanceHandler → GetTischSaldoHandler
GetTableUnpaidVariantsHandler → GetTischUnbezahltHandler
GetTableUndeliveredVariantsHandler → GetTischUngeliefertHandler

// Response JSON-Tags anpassen:
"table" → "tisch"
"tables" → "tische"
"history" → "historie"
"balanceCents" → "saldoCents"
"variants" → "positionen"
```

### 5.5 Backend: `api/` — Routen

**`service.go`:**

```go
r.HandleFunc("/bestellung-aufgeben", tc.BestellungAufgebenHandler())
r.HandleFunc("/zahlung-registrieren", tc.ZahlungRegistrierenHandler())
r.HandleFunc("/produkte-liefern", tc.ProdukteLiefernHandler())
r.HandleFunc("/get-tisch", tq.GetTischHandler())
r.HandleFunc("/get-aktive-tische", tq.GetAktiveTischeHandler())
r.HandleFunc("/get-tisch-historie", tq.GetTischHistorieHandler())
r.HandleFunc("/get-tisch-saldo", tq.GetTischSaldoHandler())
r.HandleFunc("/get-tisch-unbezahlt", tq.GetTischUnbezahltHandler())
r.HandleFunc("/get-tisch-ungeliefert", tq.GetTischUngeliefertHandler())
r.HandleFunc("/get-aktive-produkte", pq.GetActiveProductsHandler()) // Produkt-Query bleibt erstmal
```

**`senior_service.go`:**

```go
r.HandleFunc("/produkte-stornieren", tc.ProdukteStornierenHandler())
```

**`admin.go`:**

```go
// Tisch-Routen:
r.HandleFunc("/create-tisch", ...)
r.HandleFunc("/update-tisch", ...)
r.HandleFunc("/activate-tisch", ...)
r.HandleFunc("/deactivate-tisch", ...)
r.HandleFunc("/get-all-tische", ...)
// Produkt-Routen:
r.HandleFunc("/create-produkt", ...)
r.HandleFunc("/update-produkt", ...)
r.HandleFunc("/get-all-produkte", ...)
// Variante-Routen:
r.HandleFunc("/create-variante", ...)
r.HandleFunc("/update-variante", ...)
r.HandleFunc("/activate-variante", ...)
r.HandleFunc("/deactivate-variante", ...)
```

### 5.6 Backend: `api/product/` — Produkt/Variante

Die Produkt-Struktur (`domain/product/`) bleibt weitgehend gleich, da Produkt und Variante auch auf Deutsch Produkt und Variante heißen. Aber:

- `variant_not_found` → `variante_not_found` (Error Codes)
- `product_not_found` → `produkt_not_found`
- API-Routen ändern sich (siehe oben)

### 5.7 Backend: `repository/table_repo/` und `product_repo/`

- `table_repo`: `Table` → `Tisch` in Domain-Mapping-Funktionen
- `product_repo`: kaum Änderungen (DB-Spalten bleiben englisch), nur Domain-Mapping anpassen falls Structs sich ändern

### 5.8 Backend: `sqlc/queries/`

SQL-Queries selbst bleiben weitgehend gleich (SQL-Tabellen-/Spaltennamen ändern sich NICHT). Aber sqlc-Funktionsnamen könnten angepasst werden:

- `GetTable` → `GetTisch` etc. — OPTIONAL, da das nur Go-interne Namen sind
- **Empfehlung:** sqlc-Funktionsnamen bleiben englisch, da sie die DB-Schicht abbilden. Die Übersetzung passiert im Repository/Domain-Layer.

### 5.9 Frontend: `service/table/` — TypeScript-Typen

**Dateien umbenennen + Inhalt ändern:**

| Alt               | Neu               |
| ----------------- | ----------------- |
| `Order.ts`        | `Bestellung.ts`   |
| `Payment.ts`      | `Zahlung.ts`      |
| `Delivery.ts`     | `Lieferung.ts`    |
| `Cancelation.ts`  | `Stornierung.ts`  |
| `Table.ts`        | `Tisch.ts`        |
| `TableBackend.ts` | `TischBackend.ts` |

**Zod-Schema-Umbenennungen (Beispiel `Bestellung.ts`):**

```typescript
export const PositionSchema = z.object({
  id: z.number().int().min(1),
  name: z.string().min(1).max(100),
  preisCents: z.number().int().min(0),   // war: priceCents
  quantity: z.number().int().min(1),
})
export type Position = z.infer<typeof PositionSchema>

export const BestellungAufgebenSchema = z.object({
  tischId: z.number().int().min(1),      // war: tableId
  positionen: PositionSchema.array().min(1), // war: variants
  comment: z.string().max(100),
})

export const BestellungSchema = z.object({
  id: z.uuid(),
  userId: z.number().int().min(1),
  tischId: z.number().int().min(1),
  positionen: PositionSchema.array().min(1),
  gesamtPreisCents: z.number().int().min(0), // war: totalPriceCents
  comment: z.string().max(100),
  aufgegebenAm: z.string().refine(...),      // war: placedAt
})
export type Bestellung = z.infer<typeof BestellungSchema>
```

### 5.10 Frontend: `service/table/TischBackend.ts`

Methoden:

```typescript
getAktiveTische()                  // war: getActiveTables
getTisch(id)                       // war: getTable
bestellungAufgeben(...)            // war: placeTableOrder
zahlungRegistrieren(...)           // war: registerTablePayment
produkteStornieren(...)            // war: cancelTableVariants
produkteLiefern(...)               // war: deliverTableVariants
getTischHistorie(tischId)          // war: getTableHistory
getTischSaldo(tischId)             // war: getTableBalance
getTischUnbezahlt(tischId)         // war: getTableUnpaidVariants
getTischUngeliefert(tischId)       // war: getTableUndeliveredVariants
```

API-Pfade in den `backend.post()`-Aufrufen anpassen (siehe Endpunkt-Tabelle).

### 5.11 Frontend: `service/table/hooks.ts`

```typescript
useTisch(id); // war: useTable
useAktiveTische(); // war: useActiveTables
useTischHistorie(tischId); // war: useTableHistory
useTischSaldo(tischId); // war: useTableBalance
useTischUnbezahlt(tischId); // war: useTableUnpaidVariants
useTischUngeliefert(tischId); // war: useTableUndeliveredVariants
```

### 5.12 Frontend: `service/components/table/` — React-Komponenten

**Dateien umbenennen:**

| Alt                     | Neu                     |
| ----------------------- | ----------------------- |
| `Order.tsx`             | `Bestellung.tsx`        |
| `OrderDrawer.tsx`       | `BestellungDrawer.tsx`  |
| `Payment.tsx`           | `Zahlung.tsx`           |
| `PaymentDrawer.tsx`     | `ZahlungDrawer.tsx`     |
| `Delivery.tsx`          | `Lieferung.tsx`         |
| `DeliveryDrawer.tsx`    | `LieferungDrawer.tsx`   |
| `CancelationDrawer.tsx` | `StornierungDrawer.tsx` |
| `TableHistory.tsx`      | `TischHistorie.tsx`     |

In allen Dateien die Imports, Prop-Types, Variablen-Namen und Callbacks anpassen.

**UI-Texte die bereits Deutsch sind bleiben** (z.B. "Bestellung überprüfen", "Zahlung registrieren"). Texte mit "Variante" → "Produkt" wo es den Service-Kontext betrifft:

- "Varianten liefern" → "Produkte liefern"
- "Varianten stornieren" → "Produkte stornieren"
- "Variante hinzufügen/entfernen" (aria-labels) → "Produkt hinzufügen/entfernen"
- "Wurden diese Varianten an den Tisch ausgeliefert?" → "Wurden diese Produkte an den Tisch ausgeliefert?"
- "Sollen diese Varianten wirklich storniert werden?" → "Sollen diese Produkte wirklich storniert werden?"

### 5.13 Frontend: `admin/products/`

- `Variant`/`variant` in Prop-Namen und Callbacks → bleibt `Variante`/`variante` (Admin-Kontext)
- API-Endpunkt-Strings anpassen (z.B. `admin/create-variant` → `admin/create-variante`)
- Error-Codes von Backend anpassen (`variant_not_found` → `variante_not_found`)

### 5.14 Frontend: `admin/tables/`

- API-Endpunkt-Strings anpassen (`admin/create-table` → `admin/create-tisch` etc.)
- Response-Keys: `"tables"` → `"tische"`, `"table"` → `"tisch"`

### 5.15 Frontend: `service/product/`

- API-Endpunkt: `service/get-active-products` → `service/get-aktive-produkte`
- Response-Key: `"products"` → `"produkte"`

### 5.16 Frontend: `routes.ts`

- Keine Pfadänderungen nötig (Browser-URLs bleiben `/service/tables/:tableId` etc.)
- Import-Pfade ändern sich wenn Dateien umbenannt werden

### 5.17 Backend: Test-Dateien

Diese Test-Dateien müssen inhaltlich angepasst werden:

- `backend/domain/table/events_test.go`
- `backend/api/table/http/command_handler_test.go`
- `backend/api/table/http/query_handler_test.go`
- `backend/api/table/application/command_test.go`
- `backend/api/table/application/query_test.go`
- `backend/repository/table_repo/repo_test.go`
- `backend/repository/product_repo/repo_test.go`
- `backend/repository/event_repo/repo_test.go`
- `backend/api/product/http/command_handler_test.go`

---

## 6. Was NICHT geändert wird

- **DB-Tabellennamen** (`users`, `tables`, `products`, `product_variants`, `events`) — bleiben englisch
- **DB-Spaltennamen** (`price_cents`, `product_id`, etc.) — bleiben englisch
- **Go-Package-Namen** (`table`, `product`, `user`, `event`) — bleiben englisch (Go-Konvention)
- **sqlc-generierter Code** (`sqlc/dbgen/`) — wird nur regeneriert nach Query-Änderungen
- **Generische Infrastruktur** (`Auth`, `Backend`, `middleware`, `config`, `db`) — bleibt englisch
- **User/Benutzer** — bleibt englisch im Code (`user`, `UserID`) da es kein Kassenbetrieb-Fachbegriff ist
- **Browser-URL-Pfade** (`/service/tables/:tableId`) — bleiben englisch
- **Snapshot** — bleibt (technischer Begriff, kein Fachbegriff)

---

## 7. Ausführungsreihenfolge

1. **Database Migration** erstellen (`03_german_ubiquitous_language.up.sql` + `.down.sql`)
2. **Backend `domain/table/`** — Structs, Events, Funktionen (Dateien umbenennen + Inhalt)
3. **Backend `api/table/`** — Application-Layer (command.go, query.go, errors.go)
4. **Backend `api/table/http/`** — HTTP-Handler (command_handler.go, query_handler.go)
5. **Backend `api/`** — Routen (service.go, admin.go, senior_service.go)
6. **Backend `api/product/`** — Variante-Anpassungen in Routen/Errors
7. **Backend `repository/`** — Domain-Mapping anpassen
8. **Backend Tests** — Alle Test-Dateien anpassen
9. **sqlc Queries anpassen** (sofern nötig) + `sqlc generate`
10. **Frontend `service/table/`** — TypeScript-Typen + Backend-Client + Hooks
11. **Frontend `service/components/table/`** — React-Komponenten
12. **Frontend `admin/`** — Endpunkte + Response-Keys
13. **Frontend `service/product/`** — Endpunkte + Response-Keys
14. **Build verifizieren** (`cd backend && go build ./...`)
15. **Tests ausführen** (`cd backend && go test -tags=unit -race ./...` + `cd frontend && pnpm lint`)
16. **Dokumentation aktualisieren** (AGENTS.md, docs/language.md, etc.)

---

## 8. Wichtige Hinweise für die Umsetzung

- **Go-Package bleibt `table`** — Verzeichnis `backend/domain/table/` und Package-Name ändern sich NICHT. Nur die Exports (Structs, Funktionen) bekommen deutsche Namen.
- **JSON-Tags sind kritisch** — Frontend und Backend müssen exakt dieselben JSON-Feld-Namen verwenden. Die Tabelle in Abschnitt 3 ist die Single Source of Truth.
- **Event-Daten in der DB** — Die Event-Data (JSONB) enthält die alten JSON-Feldnamen. Die DB-Migration ändert NUR `type` und `subject`, NICHT die `data`-Spalte. Das bedeutet: **Alte Events mit alten JSON-Keys werden beim Deserialisieren brechen**, es sei denn wir migrieren auch die `data`-Spalte. **→ ENTSCHEIDUNG TREFFEN: Entweder auch `data`-JSONB migrieren ODER v1-Daten-Schemas als Fallback unterstützen.**
- **`sqlc generate`** muss nach jeder Änderung an `sqlc/queries/*.sql` ausgeführt werden.
- **`pnpm lint`** prüft im Frontend auf TypeScript-Fehler + ESLint.

### Offene Entscheidung: Event-Data-Migration

Die bestehenden Events haben in ihrer `data`-JSONB-Spalte Felder wie `orderId`, `variants`, `totalPriceCents` etc. Wenn wir die Go-Structs auf `bestellungId`, `positionen`, `gesamtPreisCents` umstellen, können alte Events nicht mehr deserialisiert werden.

**Optionen:**

- **A: JSONB in DB migrieren** — SQL-UPDATE mit `jsonb_set` für alle betroffenen Felder. Aufwädig aber sauber.
- **B: Alte JSON-Keys als Fallback** — Nicht empfohlen, verkompliziert den Code dauerhaft.
- **C: DB resetten** — Wenn das Projekt pre-production ist, einfach alle Events löschen und frisch starten.

**Empfehlung: Option A oder C** — Frage den Benutzer.

---

## 9. Status

Tracking analog zu `tasks.md`:

- [x] Task 1: Database Migration
- [x] Task 2: Backend Domain Layer
- [x] Task 3: Backend Repository Layer
- [x] Task 4: Backend Application Layer (Commands & Queries)
- [x] Task 5: Backend HTTP Handler
- [x] Task 6: Backend Routen & Product API
- [x] Task 7: Frontend Service Types, Backend-Client & Hooks
- [x] Task 8: Frontend Service-Komponenten
- [x] Task 9: Frontend Admin-Bereich
- [ ] Task 10: Build verifizieren & Tests ausführen
- [ ] Task 11: Dokumentation aktualisieren
