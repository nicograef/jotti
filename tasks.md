# Ubiquitous Language Migration — Tasks

> Jeder Task ist so geschnitten, dass er in **einer einzelnen Coding-Agent-Session** ausführbar ist.
> Kontext für alle Tasks: `context.md` (Begriffstabelle, Design-Entscheidungen, Struct-Definitionen).
>
> **Wichtig:** Vor Task 1 muss die offene Entscheidung zur Event-Data-Migration geklärt werden (Option A: JSONB migrieren, Option C: DB resetten). Siehe `context.md` Abschnitt 8.

---

## Vorbedingung: Entscheidung Event-Data-Migration

Die bestehenden Events haben in ihrer `data`-JSONB-Spalte englische Feldnamen (`orderId`, `variants`, `totalPriceCents` etc.). Wenn die Go-Structs auf deutsche JSON-Tags umgestellt werden, können alte Events nicht mehr deserialisiert werden.

**Optionen:**

- **A: JSONB in DB migrieren** — SQL-UPDATE mit `jsonb_set`/`jsonb_build_object` für alle betroffenen Felder. Aufwändig aber sauber.
- **C: DB resetten** — Alle Events löschen und frisch starten (nur wenn pre-production).

→ **Entscheide dich vor Task 1 und teile dem Agent die Entscheidung mit.**

---

## Task 1: Database Migration ✅

**Status:** Abgeschlossen — obsolet durch Konsolidierung aller Migrationen in eine einzige `01_initial.up.sql`. Die `senior_service`-Rolle wurde in die initiale Enum übernommen. Event-Daten-Migration entfällt, da keine bestehenden Events vorhanden.

**Scope:** `database/migrations/`

**Was tun:**

- `03_german_ubiquitous_language.up.sql` erstellen:
  - Event-Typen umbenennen (`table.order-placed:v1` → `tisch.bestellung-aufgegeben:v1` etc.)
  - Subject-Format umbenennen (`table:` → `tisch:`)
  - Je nach Entscheidung: JSONB-`data`-Spalte migrieren (Option A) oder Events löschen (Option C)
- `03_german_ubiquitous_language.down.sql` erstellen (Rollback)

**Referenz:** `context.md` Abschnitt 5.1 + Abschnitt 8 (offene Entscheidung)

**Verifizierung:** Migration-Dateien syntaktisch korrekt, up+down vorhanden.

---

## Task 2: Backend Domain Layer — Structs & Events ✅

**Status:** Abgeschlossen

**Scope:** `backend/domain/table/` (12 Dateien)

**Was tun:**

- Dateien umbenennen (per `git mv`):
  - `order.go` → `bestellung.go`
  - `payment.go` → `zahlung.go`
  - `delivery.go` → `lieferung.go`
  - `cancelation.go` → `stornierung.go`
  - `table.go` → `tisch.go`
  - `orderPlacedEvent.go` → `bestellungAufgegebenEvent.go`
  - `paymentRegisteredEvent.go` → `zahlungRegistriertEvent.go`
  - `productsCanceledEvent.go` → `produkteStorniertEvent.go`
  - `productsDeliveredEvent.go` → `produkteGeliefertEvent.go`
- Structs umbenennen: `Order`→`Bestellung`, `Payment`→`Zahlung`, `Delivery`→`Lieferung`, `Cancelation`→`Stornierung`, `LineItem`→`Position`, `Table`→`Tisch`
- JSON-Tags anpassen (siehe Begriffstabelle in `context.md` Abschnitt 3)
- Event-Type-Konstanten anpassen (`events.go`)
- Event-Constructor-Funktionen umbenennen (`events.go` + einzelne Event-Dateien)
- State-Rekonstruktions-Funktionen umbenennen (`events.go`)
- Hilfsfunktionen umbenennen (`accumulateVariants`→`accumulatePositionen` etc.)
- `events_test.go` anpassen

**Referenz:** `context.md` Abschnitte 5.2, die Struct-Definitionen und Funktionsnamen-Tabellen

**Verifizierung:** `cd backend && go build ./domain/table/`

---

## Task 3: Backend Repository Layer ✅

**Status:** Abgeschlossen

**Scope:** `backend/repository/table_repo/`, `backend/repository/product_repo/`, `backend/repository/event_repo/`

**Was tun:**

- `table_repo/`: Domain-Mapping-Funktionen anpassen (`Table`→`Tisch` in allen Stellen, die Domain-Structs verwenden)
- `product_repo/`: Nur anpassen falls Structs sich geändert haben (minimal)
- `event_repo/`: Prüfen ob Event-Typen oder Subjects referenziert werden → anpassen
- Mock-Dateien in `event_repo/mock.go` anpassen
- Tests anpassen: `table_repo/repo_test.go`, `event_repo/repo_test.go`, `product_repo/repo_test.go`

**Referenz:** `context.md` Abschnitt 5.7

**Verifizierung:** `cd backend && go build ./repository/...`

---

## Task 4: Backend Application Layer (Commands & Queries) ✅

**Status:** Abgeschlossen

**Scope:** `backend/api/table/application/`

**Was tun:**

- `command.go`: Interface-Methoden umbenennen (z.B. `PlaceTableOrder`→`BestellungAufgeben`, `RegisterTablePayment`→`ZahlungRegistrieren` etc.)
- `query.go`: Query-Methoden umbenennen (`GetTable`→`GetTisch`, `GetActiveTables`→`GetAktiveTische` etc.)
- `errors.go`: Error-Variablen umbenennen (`ErrTableNotFound`→`ErrTischNotFound` etc.)
- Parameter-Typen anpassen (`[]table.LineItem`→`[]table.Position`)
- Tests anpassen: `command_test.go`, `query_test.go`

**Referenz:** `context.md` Abschnitt 5.3

**Verifizierung:** `cd backend && go build ./api/table/...`

---

## Task 5: Backend HTTP Handler ✅

**Status:** Abgeschlossen

**Scope:** `backend/api/table/http/`

**Was tun:**

- `command_handler.go`: Handler-Methoden umbenennen (`PlaceTableOrderHandler`→`BestellungAufgebenHandler` etc.)
- Request-Struct JSON-Tags anpassen (`"tableId"`→`"tischId"`, `"variants"`→`"positionen"`)
- `query_handler.go`: Query-Handler umbenennen + Response JSON-Tags anpassen (`"table"`→`"tisch"`, `"tables"`→`"tische"`, `"balanceCents"`→`"saldoCents"` etc.)
- `factory.go`: Falls dort Handler referenziert werden → anpassen
- Tests anpassen: `command_handler_test.go`, `query_handler_test.go`

**Referenz:** `context.md` Abschnitt 5.4

**Verifizierung:** `cd backend && go build ./api/table/...`

---

## Task 6: Backend Routen & Product API

**Scope:** `backend/api/service.go`, `backend/api/senior_service.go`, `backend/api/admin.go`, `backend/api/product/`

**Was tun:**

- `service.go`: Alle Service-Routen umbenennen (siehe Endpunkt-Tabelle in `context.md`)
- `senior_service.go`: `/cancel-table-variants` → `/produkte-stornieren`
- `admin.go`: Tisch-, Produkt- und Variante-Routen umbenennen
- `api/product/http/`: API-Endpunkte in Handler anpassen, Error-Codes anpassen (`variant_not_found`→`variante_not_found`, `product_not_found`→`produkt_not_found`)
- `api/product/application/`: Error-Codes anpassen falls dort definiert
- Tests anpassen: `api/product/http/command_handler_test.go`

**Referenz:** `context.md` Abschnitte 5.5, 5.6

**Verifizierung:** `cd backend && go build ./... && go test -tags=unit -race ./...`

---

## Task 7: Frontend Service Types, Backend-Client & Hooks

**Scope:** `frontend/src/service/table/`, `frontend/src/service/product/`

**Was tun:**

- Dateien umbenennen:
  - `Order.ts` → `Bestellung.ts`
  - `Payment.ts` → `Zahlung.ts`
  - `Delivery.ts` → `Lieferung.ts`
  - `Cancelation.ts` → `Stornierung.ts`
  - `Table.ts` → `Tisch.ts`
  - `TableBackend.ts` → `TischBackend.ts`
- Zod-Schemas + TypeScript-Typen umbenennen (siehe `context.md` Abschnitt 5.9)
- `TischBackend.ts`: Methoden + API-Pfade anpassen (Abschnitt 5.10)
- `hooks.ts`: Hook-Namen anpassen (`useTable`→`useTisch` etc., Abschnitt 5.11)
- `service/product/`: API-Endpunkt anpassen (`get-active-products`→`get-aktive-produkte`), Response-Key `"products"`→`"produkte"` (Abschnitt 5.15)

**Referenz:** `context.md` Abschnitte 5.9, 5.10, 5.11, 5.15

**Verifizierung:** `cd frontend && pnpm lint` (nur Type-Checking, Komponenten werden in Task 8 angepasst — Lint-Fehler dort erwartet)

---

## Task 8: Frontend Service-Komponenten

**Scope:** `frontend/src/service/components/table/`, `frontend/src/service/TablePage.tsx`, `frontend/src/service/TableSelectionPage.tsx`

**Was tun:**

- Dateien umbenennen:
  - `Order.tsx` → `Bestellung.tsx`
  - `OrderDrawer.tsx` → `BestellungDrawer.tsx`
  - `Payment.tsx` → `Zahlung.tsx`
  - `PaymentDrawer.tsx` → `ZahlungDrawer.tsx`
  - `Delivery.tsx` → `Lieferung.tsx`
  - `DeliveryDrawer.tsx` → `LieferungDrawer.tsx`
  - `CancelationDrawer.tsx` → `StornierungDrawer.tsx`
  - `TableHistory.tsx` → `TischHistorie.tsx`
- Imports, Prop-Types, Variablen-Namen, Callbacks in allen Dateien anpassen
- `drawerUtils.ts` + `drawerUtils.test.ts`: Funktionsnamen anpassen (`selectVariants`→analog, `calculateTotalPrice`→analog falls betroffen)
- UI-Texte anpassen: "Varianten liefern"→"Produkte liefern", "Varianten stornieren"→"Produkte stornieren" etc. (Abschnitt 5.12)
- `TablePage.tsx` und `TableSelectionPage.tsx`: Imports + Verwendung der neuen Hooks/Typen anpassen
- `routes.ts`: Import-Pfade aktualisieren (URL-Pfade bleiben gleich)

**Referenz:** `context.md` Abschnitte 5.12, 5.16

**Verifizierung:** `cd frontend && pnpm lint`

---

## Task 9: Frontend Admin-Bereich

**Scope:** `frontend/src/admin/tables/`, `frontend/src/admin/products/`

**Was tun:**

- `admin/tables/`:
  - API-Endpunkte anpassen (`admin/create-table`→`admin/create-tisch` etc.)
  - Response-Keys anpassen (`"tables"`→`"tische"`, `"table"`→`"tisch"`)
  - `TableBackend.ts`, `Table.ts`, `hooks.ts` und alle Komponenten anpassen
- `admin/products/`:
  - API-Endpunkte anpassen (`admin/create-variant`→`admin/create-variante`, `admin/create-product`→`admin/create-produkt` etc.)
  - Error-Codes anpassen (`variant_not_found`→`variante_not_found`, `product_not_found`→`produkt_not_found`)
  - `ProductBackend.ts`, `Product.ts`, `hooks.ts` und alle Komponenten anpassen

**Referenz:** `context.md` Abschnitte 5.13, 5.14

**Verifizierung:** `cd frontend && pnpm lint`

---

## Task 10: Build verifizieren & Tests ausführen

**Scope:** Gesamtprojekt

**Was tun:**

- Backend: `cd backend && go build ./...`
- Backend Unit-Tests: `cd backend && go test -tags=unit -race ./...`
- Frontend Lint: `cd frontend && pnpm lint`
- Integrationstests: `./test-integration.sh` (falls DB-Migration angewendet)
- Fehler analysieren und beheben

**Wichtig:** Dieser Task ist ein Checkpoint. Alle vorherigen Tasks müssen abgeschlossen sein.

---

## Task 11: Dokumentation aktualisieren

**Scope:** `AGENTS.md`, `docs/language.md`, `docs/requirements.md`, `docs/implementation-plan.md`, `docs/development.md`, `README.md`

**Was tun:**

- `AGENTS.md`: Alle Beispiele und Referenzen auf neue Endpunkte, Struct-Namen, Event-Typen aktualisieren
- `docs/language.md`: Ubiquitous Language Tabelle auf neue deutsche Begriffe aktualisieren
- `docs/requirements.md`: Referenzen auf alte Begriffe aktualisieren
- `docs/implementation-plan.md`: Falls Endpunkte/Strukturen referenziert → anpassen
- `docs/development.md`: Falls API-Beispiele vorhanden → anpassen
- `README.md`: Falls API-Beispiele vorhanden → anpassen
- `context.md`: Status-Checkboxen auf erledigt setzen

**Verifizierung:** Manuelles Review — alle Docs konsistent mit neuem Code.

---

## Zusammenfassung

| #   | Task                              | Scope                                                          | Abhängig von |
| --- | --------------------------------- | -------------------------------------------------------------- | ------------ |
| 0   | Entscheidung Event-Data-Migration | —                                                              | —            |
| 1   | Database Migration                | `database/migrations/`                                         | Task 0       |
| 2   | Backend Domain Layer              | `backend/domain/table/`                                        | —            |
| 3   | Backend Repository Layer          | `backend/repository/`                                          | Task 2       |
| 4   | Backend Application Layer         | `backend/api/table/application/`                               | Task 2, 3    |
| 5   | Backend HTTP Handler              | `backend/api/table/http/`                                      | Task 4       |
| 6   | Backend Routen & Product API      | `backend/api/*.go`, `backend/api/product/`                     | Task 5       |
| 7   | Frontend Service Types & Hooks    | `frontend/src/service/table/`, `frontend/src/service/product/` | Task 6       |
| 8   | Frontend Service-Komponenten      | `frontend/src/service/components/table/`, Pages                | Task 7       |
| 9   | Frontend Admin-Bereich            | `frontend/src/admin/`                                          | Task 6       |
| 10  | Build & Tests                     | Gesamtprojekt                                                  | Task 1–9     |
| 11  | Dokumentation                     | `docs/`, `AGENTS.md`, `README.md`                              | Task 10      |

> **Tasks 1 und 2 können parallel gestartet werden** (keine Abhängigkeit untereinander).
> **Tasks 7, 8, 9 können teilweise parallel laufen** (7 vor 8, aber 9 ist unabhängig von 7/8).
