# Implementierungsplan — Phase 1 & 2

Dieser Plan beschreibt die nächsten sechs Features in der Reihenfolge ihrer Implementierung. Jeder Abschnitt enthält genug Detail, damit ein Coding Agent das Feature eigenständig umsetzen kann.

---

## Übersicht

| #   | Feature                                         | Typ               | Komplexität | Branch                       |
| --- | ----------------------------------------------- | ------------------ | ----------- | ---------------------------- |
| 21  | Produktkategorien in Service-UI gruppieren       | Frontend-only      | Niedrig     | `feature/21-product-categories` |
| 22  | Neue Rolle `senior_service` + Stornierung einschränken | Full-Stack + DB | Mittel      | `feature/22-senior-service-role` |
| 23  | Tisch-Schnellsuche                               | Frontend-only      | Niedrig     | `feature/23-table-search`    |
| 37  | Rückgeldberechnung                               | Frontend-only      | Niedrig     | `feature/37-change-calculation` |
| 24  | Übersicht eigene Bestellungen                    | Full-Stack         | Mittel      | `feature/24-user-orders`     |
| 26  | Umsatz pro Bediener (Tagesabrechnung)            | Full-Stack         | Mittel      | `feature/26-revenue-by-user` |

---

## #21 — Produktkategorien in Service-UI gruppieren

### Ziel

Die flache Produktliste in der Bestellansicht soll nach Produktkategorie (`food`, `beverage`, `other`) gruppiert werden, damit Servicekräfte schneller das richtige Produkt finden.

### Ist-Zustand

- `ProductList.tsx` (`frontend/src/service/components/table/ProductList.tsx`) rendert `props.products.map(...)` als flache Liste.
- Jedes `Product`-Objekt hat bereits ein `category`-Feld (`food | beverage | other`), das vom Backend geliefert wird.
- Die Daten kommen über `useActiveProducts()` → `ProductBackend.getActiveProducts()`.

### Implementierung

**Nur Frontend-Änderungen. Kein Backend, kein DB-Schema.**

#### 1. Kategorie-Labels definieren

In `frontend/src/service/product/Product.ts` ein Mapping ergänzen:

```typescript
export const ProductCategoryLabels: Record<ProductCategory, string> = {
  food: 'Essen',
  beverage: 'Getränke',
  other: 'Sonstiges',
}

/** Sortierreihenfolge der Kategorien in der UI */
export const ProductCategoryOrder: ProductCategory[] = ['food', 'beverage', 'other']
```

#### 2. Produkte gruppieren

In `ProductList.tsx`:

1. Produkte nach Kategorie gruppieren (z. B. mit `Object.groupBy` oder manuellem Reduce).
2. Pro Kategorie eine Überschrift rendern, dann die Produkte dieser Kategorie.
3. Die Reihenfolge der Kategorien fest definieren: Essen → Getränke → Sonstiges.

```tsx
// Pseudocode-Struktur in ProductList.tsx
{ProductCategoryOrder.map((category) => {
  const categoryProducts = props.products.filter(p => p.category === category)
  if (categoryProducts.length === 0) return null
  return (
    <div key={category}>
      <h2 className="text-lg font-semibold mt-4 mb-2">
        {ProductCategoryLabels[category]}
      </h2>
      <ItemGroup className="grid gap-2 lg:grid-cols-2 2xl:grid-cols-3">
        {categoryProducts.map((product) => (
          // ... bestehender Product-Render-Code ...
        ))}
      </ItemGroup>
    </div>
  )
})}
```

#### 3. Skeleton anpassen

`ProductListSkeleton` sollte ebenfalls Kategorie-Überschriften andeuten (3 Gruppen mit je 2 Skeleton-Items).

### Akzeptanzkriterien

- [ ] Produkte sind nach Kategorie gruppiert mit deutscher Überschrift
- [ ] Reihenfolge: Essen → Getränke → Sonstiges
- [ ] Leere Kategorien werden nicht angezeigt
- [ ] +/−-Buttons und Mengenanzeige funktionieren wie zuvor
- [ ] `pnpm lint` ohne Fehler

---

## #22 — Neue Rolle `senior_service` + Stornierung einschränken

### Ziel

Einführung einer neuen Rolle **`senior_service`** (Serviceleitung), die alles darf was `service` darf, plus zusätzlich stornieren. Normale Servicekräfte (`service`) dürfen **nicht** stornieren. Nur `admin` und `senior_service` dürfen stornieren.

### Rollenname

`senior_service` — im UI als **„Serviceleitung"** angezeigt. Kurz, selbsterklärend, passt zum bestehenden `service`-Namensschema.

### Betroffene Dateien

#### Backend

1. **DB-Migration** (`database/migrations/02_add_senior_service_role.up.sql` + `.down.sql`)
2. **Domain** (`backend/domain/user/user.go`)
3. **JWT** (`backend/domain/jwt/jwt.go` — keine Änderung nötig, Role ist generisch als `string`)
4. **Middleware** (`backend/api/middleware/middleware.go` — keine Änderung nötig, arbeitet mit `allowedRoles []string`)
5. **Service-Router** (`backend/api/service.go`) — Route-Split oder Rollenprüfung für Cancel
6. **App-Wiring** (`backend/app/app.go`) — `senior_service` zu Service-Rollen hinzufügen

#### Frontend

7. **Auth** (`frontend/src/lib/Auth.ts`) — Rolle im Token erkennen
8. **Admin User-Model** (`frontend/src/admin/users/User.ts`) — Rolle als Option
9. **Route Guards** (`frontend/src/routes.ts`) — `senior_service` darf Service-Bereich nutzen
10. **Payment/Stornierung** (`frontend/src/service/components/table/Payment.tsx`) — CancelationDrawer nur für `admin` + `senior_service`

### Implementierung

#### 1. DB-Migration

```sql
-- 02_add_senior_service_role.up.sql
ALTER TYPE UserRole ADD VALUE 'senior_service';
```

```sql
-- 02_add_senior_service_role.down.sql
-- PostgreSQL erlaubt kein einfaches Entfernen von ENUM-Werten.
-- Für Rollback: Typ neu erstellen und Spalte migrieren.
-- In der Praxis: Diese Migration ist vorwärts-only.
-- Sicherheitshalber: UPDATE users SET role = 'service' WHERE role = 'senior_service';
-- Dann Typ-Umbau (CREATE TYPE ... AS ENUM, ALTER TABLE, DROP TYPE, RENAME TYPE).
```

#### 2. Domain-Modell (`backend/domain/user/user.go`)

```go
const (
    AdminRole         Role = "admin"
    SeniorServiceRole Role = "senior_service"
    ServiceRole       Role = "service"
)

var RoleSchema = z.StringLike[Role]().OneOf(
    []Role{AdminRole, SeniorServiceRole, ServiceRole},
    z.Message("Invalid role"),
)
```

#### 3. App-Wiring (`backend/app/app.go`)

Die Service-Middleware muss `senior_service` als erlaubte Rolle akzeptieren:

```go
service := middleware.NewJwtMiddleware(cfg.JWTSecret, []string{"admin", "senior_service", "service"})
```

#### 4. Stornierung nur für `admin` und `senior_service`

Zwei Optionen (Option A bevorzugt):

**Option A: Separate Middleware für Cancel-Route**

In `backend/api/service.go` den `cancel-table-variants`-Handler aus dem allgemeinen Service-Router herausnehmen und separat mit eingeschränkter Middleware registrieren.

Alternativ: In `backend/app/app.go` einen separaten Router für stornierungsberechtigte Endpunkte erstellen:

```go
// Normaler Service-Bereich (alle Service-Rollen)
serviceApi := api.NewServiceApi(db)
service := middleware.NewJwtMiddleware(cfg.JWTSecret, []string{"admin", "senior_service", "service"})
r.Handle("/service/", service(http.StripPrefix("/service", serviceApi)))

// Stornierung nur für admin + senior_service
cancelApi := api.NewCancelApi(db) // Neuer Mini-Router nur für cancel-table-variants
cancelMiddleware := middleware.NewJwtMiddleware(cfg.JWTSecret, []string{"admin", "senior_service"})
r.Handle("/service/cancel-table-variants", cancelMiddleware(http.StripPrefix("/service", cancelApi)))
```

**Alternativ Option B: Rollenprüfung im Handler**

Im `CancelTableVariantsHandler` die Rolle aus dem JWT-Context prüfen. Dafür müsste die Middleware die Rolle zusätzlich zum UserID in den Context schreiben. Aktuell wird nur `UserIDKey` gesetzt — man müsste einen `RoleKey` ergänzen.

**Empfehlung:** Option A ist sauberer, da die Middleware bereits die Rollenprüfung macht und kein Code im Handler geändert werden muss.

**Konkrete Umsetzung mit Option A:**

1. In `api/service.go` den `cancel-table-variants`-Handler entfernen.
2. Neue Funktion `NewCancelApi(db) http.Handler` in `api/service.go` (oder eigene Datei), die nur den Cancel-Handler registriert.
3. In `app/app.go` den Cancel-Endpunkt mit eingeschränkter Middleware registrieren (s. oben).

#### 5. Frontend: Auth erweitern (`frontend/src/lib/Auth.ts`)

```typescript
const JottiTokenSchema = z.object({
  iss: z.literal('jotti'),
  exp: z.int().min(0),
  iat: z.int().min(0),
  sub: z.number().int().min(1),
  role: z.enum(['admin', 'senior_service', 'service']),
})

// Neue Getter:
public get isSeniorService(): boolean {
  return this.token?.role === 'senior_service'
}

public get canCancel(): boolean {
  return this.isAdmin || this.isSeniorService
}
```

#### 6. Frontend: Route Guards (`frontend/src/routes.ts`)

```typescript
export function ServiceGuard() {
  const canAccessService =
    AuthSingleton.isAuthenticated &&
    (AuthSingleton.isService || AuthSingleton.isSeniorService || AuthSingleton.isAdmin)

  if (!canAccessService) {
    return redirect('/')
  }
}
```

#### 7. Frontend: User-Model (`frontend/src/admin/users/User.ts`)

```typescript
export const UserRole = {
  ADMIN: 'admin',
  SENIOR_SERVICE: 'senior_service',
  SERVICE: 'service',
} as const
```

Im `NewUserDialog` und `EditUserDialog`: Die Auswahl der Rolle um „Serviceleitung" ergänzen. Das Label für die Rolle im UI als Map:

```typescript
export const UserRoleLabels: Record<UserRole, string> = {
  admin: 'Admin',
  senior_service: 'Serviceleitung',
  service: 'Service',
}
```

#### 8. Frontend: CancelationDrawer ausblenden (`frontend/src/service/components/table/Payment.tsx`)

```tsx
import { AuthSingleton } from '@/lib/Auth'

// Im JSX: CancelationDrawer nur rendern wenn canCancel
{AuthSingleton.canCancel && (
  <div className="flex-1">
    <CancelationDrawer ... />
  </div>
)}
```

### DB-Migration testen

```bash
# Container neu starten, damit Migration 02 angewendet wird
docker compose -f docker-compose.dev.yml up --build -d
```

### Akzeptanzkriterien

- [ ] Neue Rolle `senior_service` in DB, Backend und Frontend
- [ ] `senior_service` kann Service-Bereich nutzen (Bestellen, Bezahlen, Liefern)
- [ ] `senior_service` kann stornieren
- [ ] `service` kann **nicht** stornieren (kein Button, Backend lehnt ab)
- [ ] `admin` kann weiterhin stornieren
- [ ] Admin kann Benutzer mit Rolle `senior_service` anlegen/bearbeiten
- [ ] Rolle wird im Admin-User-Management als „Serviceleitung" angezeigt
- [ ] Unit-Tests für Rollenvalidierung angepasst
- [ ] `go test -tags=unit -race ./...` und `pnpm lint` ohne Fehler

### Dokumentation aktualisieren

Nach Implementierung folgende Dateien anpassen:

- `AGENTS.md` — Rollendokumentation, DB-Schema-Enums
- `.github/copilot-instructions.md` — Auth-Rollen, Middleware-Beschreibung
- `README.md` — Rollen-Liste
- `ANFORDERUNGEN.md` — #22 als ✅ markieren, neue Rolle dokumentieren

---

## #23 — Tisch-Schnellsuche

### Ziel

Auf der Tischübersicht (`TableSelectionPage.tsx`) soll ein Suchfeld die Tischliste filtern, damit Servicekräfte bei vielen Tischen schnell den richtigen finden.

### Ist-Zustand

- `TableSelectionPage.tsx` zeigt alle aktiven Tische als Karten-Grid.
- Daten kommen von `useActiveTables()` → Tischliste.
- Keine Filterung vorhanden.

### Implementierung

**Nur Frontend-Änderungen.**

#### 1. Suchfeld hinzufügen

In `TableSelectionPage.tsx` ein `Input`-Feld oberhalb der Tischliste ergänzen:

```tsx
import { Input } from '@/components/ui/input'
import { Search } from 'lucide-react'
import { useState } from 'react'

export function TableSelectionPage() {
  const { loading, tables } = useActiveTables()
  const [search, setSearch] = useState('')

  const filteredTables = tables.filter((table) =>
    table.name.toLowerCase().includes(search.toLowerCase())
  )

  return (
    <>
      <h1 className="text-2xl font-bold">Tisch auswählen</h1>
      <div className="relative my-4">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
        <Input
          placeholder="Tisch suchen..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="pl-9"
          autoFocus
        />
      </div>
      {loading ? <TableListSkeleton /> : <TableList tables={filteredTables} />}
    </>
  )
}
```

#### 2. Leerer Zustand

Wenn keine Tische zum Suchbegriff passen, einen Hinweis anzeigen:

```tsx
{filteredTables.length === 0 && !loading && (
  <p className="text-muted-foreground text-center mt-8">
    Kein Tisch gefunden.
  </p>
)}
```

### Akzeptanzkriterien

- [ ] Suchfeld ist auf der Tischübersicht sichtbar
- [ ] Tischliste wird bei Eingabe gefiltert (case-insensitive)
- [ ] "Kein Tisch gefunden"-Hinweis bei leerem Ergebnis
- [ ] Leeres Suchfeld zeigt alle Tische
- [ ] `pnpm lint` ohne Fehler

---

## #37 — Rückgeldberechnung

### Ziel

Im Bezahl-Drawer soll die Servicekraft den vom Gast erhaltenen Betrag eingeben können. Das System berechnet und zeigt das Rückgeld an.

### Ist-Zustand

- `PaymentDrawer.tsx` zeigt Zusammenfassung und Gesamtbetrag.
- Kein Eingabefeld für den erhaltenen Betrag.

### Implementierung

**Nur Frontend-Änderungen. Kein Backend.**

#### 1. Eingabefeld im PaymentDrawer

In `PaymentDrawer.tsx` nach der `Receipt`-Komponente ein Eingabefeld für den erhaltenen Betrag ergänzen:

```tsx
import { useState } from 'react'
import { centsToPrice, priceToCents } from '@/components/common/FormFields'

// Im Drawer-Content, nach <Receipt>:
const [receivedAmount, setReceivedAmount] = useState('')
const receivedCents = priceToCents(receivedAmount)
const changeCents = receivedCents !== null && receivedCents >= totalPrice
  ? receivedCents - totalPrice
  : null

// JSX:
<div className="px-4 mt-4 space-y-2">
  <label className="text-sm font-medium">Erhalten (€)</label>
  <Input
    type="text"
    inputMode="decimal"
    placeholder="0,00"
    value={receivedAmount}
    onChange={(e) => setReceivedAmount(e.target.value)}
  />
  {changeCents !== null && (
    <p className="text-lg font-bold text-primary">
      Rückgeld: {formatCents(changeCents)} €
    </p>
  )}
</div>
```

**Hinweis:** Prüfen, ob `priceToCents` bzw. eine ähnliche Hilfsfunktion (`centsToPrice`) in `FormFields.tsx` existiert. Falls nicht, eine einfache Parsing-Funktion schreiben, die `"12,50"` → `1250` konvertiert.

#### 2. Kein Backend-Endpunkt

Die Berechnung ist rein clientseitig. Der erhaltene Betrag wird nicht gespeichert.

### Akzeptanzkriterien

- [ ] Eingabefeld "Erhalten" im PaymentDrawer nach der Bon-Vorschau
- [ ] Rückgeld wird berechnet und angezeigt, wenn der eingegebene Betrag ≥ Gesamtbetrag
- [ ] Keine Anzeige, wenn Feld leer oder Betrag zu niedrig
- [ ] Komma und Punkt als Dezimaltrenner akzeptiert
- [ ] `pnpm lint` ohne Fehler

---

## #24 — Übersicht eigene Bestellungen

### Ziel

Servicekräfte sollen ihre eigenen aufgegebenen Bestellungen, Bezahlungen und Lieferungen sehen können — chronologisch, mit Tisch, Zeitstempel und Status.

### Implementierung

#### Backend

##### 1. Repository: Neues Query-Interface

In `repository/event_repo/` eine neue Query-Methode:

```go
// GetEventsByUserID returns all events for a specific user, ordered by timestamp DESC.
func (r *Repo) GetEventsByUserID(ctx context.Context, userID int) ([]event.Event, error) {
    rows, err := r.db.QueryContext(ctx,
        `SELECT id, user_id, type, subject, timestamp, data
         FROM events
         WHERE user_id = $1
         ORDER BY timestamp DESC
         LIMIT 200`, userID)
    // ... scan rows into []event.Event
}
```

##### 2. Application Service

Neuer Application Service in `api/table/application/` oder neuer Service:

```go
type GetUserOrdersQuery struct {
    UserID int
}

type UserOrderEntry struct {
    ID        int       `json:"id"`
    Type      string    `json:"type"`      // "order", "payment", "delivery", "cancelation"
    TableName string    `json:"tableName"` // Denormalisiert oder per JOIN
    Timestamp time.Time `json:"timestamp"`
    Items     []LineItem `json:"items"`
    Comment   string    `json:"comment"`
    TotalCents int      `json:"totalCents"`
}
```

##### 3. HTTP-Handler

Neuer Handler `GetUserOrdersHandler` in `api/table/http/`:

```go
func (h *QueryHandler) GetUserOrdersHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        userID := r.Context().Value(middleware.UserIDKey).(int)
        // Fetch events for this user, aggregate into UserOrderEntry list
        // Return JSON response
    }
}
```

##### 4. Route registrieren

In `api/service.go`:

```go
r.HandleFunc("/get-user-orders", tq.GetUserOrdersHandler())
```

#### Frontend

##### 5. Zod-Schema und Backend-Client

Neues Schema für die Response. Backend-Client-Methode in `TableBackend.ts`:

```typescript
async getUserOrders(): Promise<UserOrderEntry[]> {
  return this.client.post('/service/get-user-orders', {}, UserOrdersResponseSchema)
}
```

##### 6. Hook

```typescript
export function useUserOrders() {
  const { data: orders, ...rest } = useFetch(
    () => tableBackend.getUserOrders(),
    [] as UserOrderEntry[],
  )
  return { ...rest, orders }
}
```

##### 7. Seite

Neue Seite `frontend/src/service/UserOrdersPage.tsx`:
- Chronologische Liste der eigenen Aktionen
- Pro Eintrag: Typ-Icon (Bestellung/Bezahlung/Lieferung/Storno), Tisch, Zeitstempel, Positionen, Gesamtbetrag
- Pull-to-Refresh oder Reload-Button

##### 8. Route und Navigation

In `routes.ts`:

```typescript
{ path: 'orders', Component: UserOrdersPage },
```

In `ServiceSidebar.tsx` (oder ServiceLayout) einen Link "Meine Bestellungen" ergänzen.

### Akzeptanzkriterien

- [ ] Neuer Endpunkt `POST /service/get-user-orders` liefert Events des aktuellen Users
- [ ] Neue Service-Seite unter `/service/orders`
- [ ] Chronologische Ansicht mit Typ, Tisch, Zeitstempel, Positionen
- [ ] Navigation von Service-Sidebar erreichbar
- [ ] `go test -tags=unit -race ./...` und `pnpm lint` ohne Fehler

---

## #26 — Umsatz pro Bediener (Tagesabrechnung)

### Ziel

Admins sollen den Umsatz pro Servicekraft für einen bestimmten Tag einsehen können, um die Tagesabrechnung zu erstellen.

### Implementierung

#### Backend

##### 1. Repository-Query

Neue Query in `repository/event_repo/`:

```go
type RevenueByUser struct {
    UserID        int    `json:"userId"`
    UserName      string `json:"userName"`
    TotalCents    int    `json:"totalCents"`
    PaymentCount  int    `json:"paymentCount"`
}

// GetRevenueByUser aggregates payment events by user for a given date.
func (r *Repo) GetRevenueByUser(ctx context.Context, date time.Time) ([]RevenueByUser, error) {
    rows, err := r.db.QueryContext(ctx,
        `SELECT e.user_id, u.name, COUNT(*) as payment_count,
                COALESCE(SUM((item->>'priceCents')::int * (item->>'quantity')::int), 0) as total_cents
         FROM events e
         JOIN users u ON u.id = e.user_id,
         jsonb_array_elements(e.data->'items') AS item
         WHERE e.type = 'table.payment-registered:v1'
           AND e.timestamp::date = $1
         GROUP BY e.user_id, u.name
         ORDER BY total_cents DESC`, date)
    // ... scan rows
}
```

**Hinweis:** Die SQL-Abfrage muss den JSONB-Datenaufbau der `payment-registered`-Events matchen. Die genaue Struktur des `data`-Felds muss vorher geprüft werden (`items[].priceCents`, `items[].quantity`).

##### 2. Application Service

```go
type GetRevenueByUserQuery struct {
    Date string // Format: "2026-02-15"
}
```

##### 3. HTTP-Handler

Neuer Handler in `api/table/http/` (oder neuer Admin-spezifischer Handler):

```go
func (h *QueryHandler) GetRevenueByUserHandler() http.HandlerFunc {
    // Parse date from request body
    // Call repository
    // Return JSON response with []RevenueByUser + Summenzeile
}
```

##### 4. Route registrieren

In `api/admin.go`:

```go
r.HandleFunc("/get-revenue-by-user", tq.GetRevenueByUserHandler())
```

**Hinweis:** Da dieser Handler auf den `event_repo` zugreift und nicht auf den `table_repo`, muss ggf. ein neuer Query-Handler oder ein Reporting-Handler erstellt werden.

#### Frontend

##### 5. Admin-Seite

Neues Feature-Verzeichnis `frontend/src/admin/reporting/`:

- `ReportingBackend.ts` — Backend-Client
- `hooks.ts` — `useRevenueByUser(date)`
- `AdminReportingPage.tsx` — Tabelle mit Datumswähler

##### 6. Datumswähler

Ein einfaches Datums-Input (`<input type="date">`) mit dem aktuellen Tag als Default.

##### 7. Tabelle

| Bediener      | Umsatz    | Zahlungen |
| ------------- | --------- | --------- |
| Max Mustermann | 1.234,56 € | 42        |
| **Gesamt**     | **2.500,00 €** | **85** |

##### 8. Route und Navigation

In `routes.ts` unter admin:

```typescript
{ path: 'reporting', Component: AdminReportingPage },
```

In `AdminSidebar.tsx` einen Link "Tagesabrechnung" ergänzen.

### Akzeptanzkriterien

- [ ] Neuer Endpunkt `POST /admin/get-revenue-by-user` mit Datums-Parameter
- [ ] Aggregation der `payment-registered`-Events nach `user_id` und Datum
- [ ] Neue Admin-Seite unter `/admin/reporting`
- [ ] Tabelle mit Bediener, Umsatz, Anzahl Zahlungen und Gesamt-Summe
- [ ] Datumswähler mit Standard = heute
- [ ] Navigation von Admin-Sidebar erreichbar
- [ ] `go test -tags=unit -race ./...` und `pnpm lint` ohne Fehler

---

## Allgemeine Hinweise für Coding Agents

### Branch-Workflow

1. Jedes Feature auf eigenem Branch implementieren (Branch-Name aus Tabelle oben).
2. Von `main` branchen.
3. Nach Implementierung: Tests laufen lassen, Lint prüfen, committen.
4. Dokumentation aktualisieren (AGENTS.md, copilot-instructions.md, README.md, ANFORDERUNGEN.md).

### Reihenfolge der Implementierung

1. **#21** zuerst (kein Dependency)
2. **#22** danach (DB-Migration + Rolle nötig für spätere Features)
3. **#23** und **#37** parallel möglich (beide Frontend-only, keine Abhängigkeiten)
4. **#24** und **#26** danach (Backend + Frontend, unabhängig voneinander)

### Test-Befehle

```bash
# Backend Unit-Tests
cd backend && go test -tags=unit -race ./...

# Frontend Lint
cd frontend && pnpm lint

# Dev-Stack neu starten (nach DB-Migration)
docker compose -f docker-compose.dev.yml up --build -d
```

### Konventionen (Wiederholung)

- **POST-only API.** Keine GET/PUT/DELETE.
- **Geldbeträge in Cent (int).** Nie Floats.
- **Deutsche UI, englischer Code.**
- **Backend ist Single Source of Truth** für Datenfilterung — #23 (Tisch-Suche) ist eine der wenigen gerechtfertigten Frontend-Filterungen, da es sich um eine rein visuelle Filterung einer bereits vollständigen Liste handelt.
- **Dokumentation synchron halten** nach jeder Änderung.
