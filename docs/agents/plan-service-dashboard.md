# Plan: Service-Dashboard mit Meine Tische, Schnellsuche & Eigene Übersicht

## TL;DR

Das Service-Dashboard wird von einer flachen Tischliste zu einem personalisierten Dashboard umgebaut. Servicekräfte markieren sich Tische als **Favoriten** (serverseitig, DB), sehen auf dem Dashboard ihre Tische mit vollem State (Saldo, ausstehende Lieferungen, offene Bezahlungen, Auszahlungsbedarf) als **Rich Cards**, erhalten eine kompakte **Eigene Übersicht** (R-06) mit KPIs und können über einen **Drawer mit Schnellsuche** (K-11) alle Tische erreichen und Favoriten verwalten.

## Betroffene Anforderungen

| ID | Name | Ist-Status | Änderung |
|----|------|-----------|----------|
| K-06 | Tischübersicht & Navigation | ✅ Umgesetzt | Erweitern: Dashboard zeigt primär Favoriten mit vollem State |
| K-11 | Tisch-Schnellsuche | 🔲 Offen | Implementieren: Integriert in Alle-Tische-Drawer |
| R-06 | Eigene Übersicht | 🔲 Offen | Implementieren: KPI-Sektion auf dem Dashboard |
| NEU | Tisch-Favoriten (K-14) | – | Neue Anforderung: Servicekraft kann Tische als Favoriten markieren |

## Entscheidungen

- **Zuordnungsmodell:** Servicekraft wählt selbst (Favoriten), kein Admin-Assignment
- **Persistenz:** Serverseitig (DB-Tabelle `tisch_favoriten`)
- **Leerer Zustand:** Wenn keine Favoriten → Hinweis "Keine Tische markiert" + Button zur Tischauswahl
- **Alle Tische:** Erreichbar über Drawer/Sheet mit Suchfeld (K-11) und Stern-Toggle zum Favorisieren
- **R-06 Platzierung:** Kompakte KPI-Sektion direkt auf dem Service-Dashboard
- **Favoriten markieren:** In der Alle-Tische-Ansicht (Drawer) per Stern-Toggle

## Ubiquitous Language — Neue Begriffe

| Begriff | Go-Struct | TS-Typ | JSON-Key | DB-Tabelle | API-Pfad | UI-Label |
|---------|-----------|--------|----------|------------|----------|----------|
| Favorit | `Favorit` | `Favorit` | – | `tisch_favoriten` | `/favorit-*` | „Meine Tische" |
| Eigene Übersicht | `EigeneUebersicht` | `EigeneUebersicht` | camelCase | – (Query auf events) | `/get-eigene-uebersicht` | „Meine Übersicht" |

---

## Phase 0 · Anforderungen & Design aktualisieren

**Ziel:** Dokumentation ist konsistent, bevor Code geschrieben wird.

1. **anforderungen.md** aktualisieren:
   - K-06 erweitern: Dashboard zeigt primär "Meine Tische" (Favoriten) mit vollem State (Rich Cards: Saldo, ausstehende Lieferungen, unbezahlte Positionen, Auszahlungsbedarf). Alle-Tische-Zugang über Drawer.
   - K-11 präzisieren: Schnellsuche als Suchfeld im Alle-Tische-Drawer, clientseitige Filterung
   - R-06 präzisieren: KPI-Sektion auf dem Dashboard (eigene Bestellungen Anzahl+Summe, eigene Zahlungen Anzahl+Summe)
   - Neue Anforderung K-14 "Tisch-Favoriten" mit Akzeptanzkriterien
2. **language.md** aktualisieren: Neue Begriffe (Favorit, Eigene Übersicht) mit Schicht-Repräsentationen
3. **handbuch.md** §4 aktualisieren: Favoriten als einfache CRUD-Relation im Stammdaten-Kontext beschreiben

**Relevante Dateien:**
- `docs/anforderungen.md` — K-06 (Zeile ~146), K-11 (Zeile ~214), R-06 (Zeile ~574), neue K-14
- `docs/design/language.md` — Neue Begriffsdefinitionen
- `docs/design/handbuch.md` — §4 Stammdaten, §7 Read Models

---

## Phase 1 · Database & Backend — Tisch-Favoriten (Kern)

**Ziel:** Favoriten CRUD-Operationen funktionieren End-to-End.

### 1.1 · Database-Schema (*depends on Phase 0*)

Neue Tabelle in `database/migrations/01_initial.up.sql`:

```sql
CREATE TABLE IF NOT EXISTS tisch_favoriten (
    user_id INT REFERENCES users(id) NOT NULL,
    tisch_id INT REFERENCES tische(id) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, tisch_id)
);

CREATE INDEX IF NOT EXISTS idx_tisch_favoriten_user_id ON tisch_favoriten(user_id);

COMMENT ON TABLE tisch_favoriten IS 'Servicekraft-eigene Tisch-Favoriten (Meine Tische)';
```

### 1.2 · sqlc-Queries (*depends on 1.1*)

Neue Datei `backend/sqlc/queries/favoriten.sql`:
- `AddFavorit :exec` — INSERT ... ON CONFLICT DO NOTHING
- `RemoveFavorit :exec` — DELETE WHERE user_id = $1 AND tisch_id = $2
- `GetFavoritenByUser :many` — SELECT tisch_id FROM tisch_favoriten WHERE user_id = $1

→ `make sqlc` nach Query-Erstellung

### 1.3 · Repository (*depends on 1.2*)

Neues Package `backend/repository/favorit_repo/`:
- `FavoritRepo` struct mit Methoden:
  - `Add(ctx, userID, tischID) error`
  - `Remove(ctx, userID, tischID) error`
  - `GetByUser(ctx, userID) ([]int, error)`
- Pattern: analog zu `table_repo` (sqlc-basiert, simple CRUD)

### 1.4 · Application Service (*depends on 1.3*)

In `backend/api/table/application/command.go` oder neues Favorit-Application-Package:
- `FavoritHinzufuegen(ctx, userID, tischID)` — prüft ob Tisch aktiv & existiert
- `FavoritEntfernen(ctx, userID, tischID)` — entfernt, idempotent

### 1.5 · HTTP-Handler (*depends on 1.4*)

Neue Endpoints (POST-only, per AGENTS.md):
- `POST /service/favorit-hinzufuegen` — Body: `{tischId}` — UserID aus JWT
- `POST /service/favorit-entfernen` — Body: `{tischId}` — UserID aus JWT

Registrierung in `backend/api/service.go`.

### 1.6 · Batch-State-Endpoint (*depends on 1.3, bestehende TischState-Logik*)

Neuer Query-Endpoint:
- `POST /service/get-meine-tische-state` → Liefert `TischState[]` für alle Favoriten des Users
- Application: `GetMeineTischeState(ctx, userID)` → lädt Favoriten-IDs, dann TischState pro Favorit (batch SQL-Query auf `table_state` + `tische`)
- Neuer sqlc-Query: `GetTableStatesByTischIDs :many` — WHERE tisch_id = ANY($1)

### 1.7 · Erweiterte Aktive-Tische-Antwort (*depends on 1.3*)

`GET-aktive-tische` Response erweitern um `isFavorit: boolean` pro Tisch:
- Application-Layer: Favoriten-IDs laden, mit aktiven Tischen joinen
- HTTP-DTO: `aktiverTischDTO` bekommt `IsFavorit bool json:"isFavorit"`
- Oder: eigener Endpoint `POST /service/get-aktive-tische-mit-favoriten`

**Relevante Dateien:**
- `database/migrations/01_initial.up.sql` — Schema
- `backend/sqlc/queries/favoriten.sql` — Neue Datei
- `backend/sqlc/queries/table_state.sql` — Batch-Query ergänzen
- `backend/repository/favorit_repo/repo.go` — Neues Package
- `backend/api/table/application/command.go` — Favorit-Commands
- `backend/api/table/application/query.go` — `GetMeineTischeState`, erweitertes `GetAktiveTische`
- `backend/api/table/http/command_handler.go` — Favorit-Endpoints
- `backend/api/table/http/query_handler.go` — Batch-State + erweitertes Aktive-Tische
- `backend/api/service.go` — Route-Registrierung
- `backend/app/app.go` — FavoritRepo Wiring

**Verification Phase 1:**
- Unit-Tests für Application-Service (Favorit hinzufügen/entfernen, Tisch nicht gefunden, etc.)
- Integrationstests: Favorit CRUD-Zyklus, Batch-State-Endpoint
- `make sqlc`, `make lint`, `make test`

---

## Phase 2 · Backend — Eigene Übersicht (R-06)

*Parallel mit Phase 1 möglich (unabhängig von Favoriten)*

**Ziel:** Service-User kann eigene Aktivitäts-KPIs abrufen.

### 2.1 · SQL-Query

Neue Query in `backend/sqlc/queries/reporting.sql` oder neue Datei `eigene_uebersicht.sql`:

```sql
-- name: GetEigeneUebersicht :one
SELECT
    COALESCE(COUNT(CASE WHEN type = 'tisch.bestellung-aufgenommen:v1' THEN 1 END), 0)::int
        AS anzahl_bestellungen,
    COALESCE(SUM(CASE WHEN type = 'tisch.bestellung-aufgenommen:v1'
        THEN (data->>'gesamtPreisCents')::int END), 0)::int
        AS bestellungen_cents,
    COALESCE(COUNT(CASE WHEN type = 'tisch.zahlung-kassiert:v1' THEN 1 END), 0)::int
        AS anzahl_zahlungen,
    COALESCE(SUM(CASE WHEN type = 'tisch.zahlung-kassiert:v1'
        THEN (data->>'gesamtZahlungCents')::int END), 0)::int
        AS zahlungen_cents
FROM events
WHERE user_id = @user_id
AND type IN ('tisch.bestellung-aufgenommen:v1', 'tisch.zahlung-kassiert:v1');
```

### 2.2 · Domain-Modell

In `backend/domain/reporting/reporting.go`:

```go
type EigeneUebersicht struct {
    AnzahlBestellungen int
    BestellungenCents  int
    AnzahlZahlungen    int
    ZahlungenCents     int
}
```

### 2.3 · Application + HTTP (*depends on 2.1, 2.2*)

- Application: `GetEigeneUebersicht(ctx, userID) → EigeneUebersicht`
- HTTP: `POST /service/get-eigene-uebersicht` — UserID aus JWT
- Response-DTO mit `json`-Tags in HTTP-Schicht
- Registrierung in `backend/api/service.go`

**Relevante Dateien:**
- `backend/sqlc/queries/reporting.sql` oder neue Datei
- `backend/domain/reporting/reporting.go` — EigeneUebersicht Struct
- `backend/api/reporting/application/` oder `backend/api/table/application/query.go`
- `backend/api/table/http/query_handler.go` oder `backend/api/reporting/http/`
- `backend/api/service.go` — Route

**Verification Phase 2:**
- Unit-Test: Application-Service mit Mock-Repo
- Integrationstest: Query mit Seed-Daten validieren
- `make sqlc`, `make lint`, `make test`

---

## Phase 3 · Frontend — Service-Dashboard Redesign

*Depends on Phase 1 + 2 (Backend-Endpoints müssen stehen)*

**Ziel:** Neues personalisiertes Dashboard mit Rich Cards und KPIs.

### 3.1 · Types & Backend-Klassen

- `frontend/src/service/table/Tisch.ts` — Erweitern:
  - `EigeneUebersichtSchema` Zod-Schema + `EigeneUebersicht` Type
  - `AktiverTischMitFavoritSchema` (erweitert `TischSchema` um `isFavorit: z.boolean()`)
- `frontend/src/service/table/TischBackend.ts` — Neue Methoden:
  - `getMeineTischeState(): Promise<TischState[]>`
  - `favoritHinzufuegen(tischId: number): Promise<void>`
  - `favoritEntfernen(tischId: number): Promise<void>`
  - `getAktiveTischeMitFavoriten(): Promise<AktiverTischMitFavorit[]>`
  - `getEigeneUebersicht(): Promise<EigeneUebersicht>`

### 3.2 · Hooks

- `frontend/src/service/table/hooks.ts` — Neue Hooks:
  - `useMeineTischeState()` → fetch batch state for favorites
  - `useEigeneUebersicht()` → fetch personal KPIs
  - `useAktiveTischeMitFavoriten()` → for Drawer (all tables + favorit flag)

### 3.3 · Dashboard-Seite Redesign (*depends on 3.1, 3.2*)

Umbau `frontend/src/service/TableSelectionPage.tsx` → `ServiceDashboard`:

**Layout:**

```
┌──────────────────────────────────┐
│ Header: "Meine Tische"          │
├──────────────────────────────────┤
│ KPI-Sektion (R-06):             │
│ ┌──────────┐ ┌──────────┐      │
│ │ X Bestell.│ │ X,XX € kas│     │
│ └──────────┘ └──────────┘      │
├──────────────────────────────────┤
│ Meine Tische (Rich Cards):      │
│ ┌──────────────────────────┐    │
│ │ Tisch 5      Saldo: 45€ │    │
│ │ 🔴 3 ausstehend          │    │
│ │ 💳 2 unbezahlt           │    │
│ └──────────────────────────┘    │
│ ┌──────────────────────────┐    │
│ │ Tisch 8      Saldo: -5€ │    │
│ │ ✅ Alles ausgegeben       │    │
│ │ ⚠️ Auszahlung: 5,00 €    │    │
│ └──────────────────────────┘    │
├──────────────────────────────────┤
│ [Alle Tische öffnen] FAB/Button │
└──────────────────────────────────┘
```

**Empty State** (keine Favoriten):

```
┌──────────────────────────────────┐
│ 🪑 Du hast noch keine Tische    │
│ markiert.                        │
│ [Tische auswählen →]            │
└──────────────────────────────────┘
```

### 3.4 · MeinTischCard Komponente (*parallel mit 3.3*)

Neue Komponente `frontend/src/service/components/MeinTischCard.tsx`:
- Props: `TischState` (voller State)
- Anzeige: Name, Saldo (rot wenn negativ), Badges/Chips:
  - Ausstehende Lieferungen: `Badge` mit count ("X ausstehend")
  - Unbezahlte Positionen: `Badge` mit count ("X unbezahlt")
  - Auszahlung erforderlich: `Badge variant="destructive"` ("Auszahlung: X €")
  - Alles ausgegeben + alles bezahlt: grüner Status
- Tap → Navigate zu `/service/tables/:id`
- Nutzt bestehende shadcn/ui Komponenten: `Card`, `Badge`, `Item`

### 3.5 · KPI-Sektion (R-06) (*parallel mit 3.3*)

Komponente `frontend/src/service/components/EigeneUebersicht.tsx`:
- Kompakte Darstellung: 2 oder 4 Cards in Grid
- "X Bestellungen (XX,XX €)" + "X Zahlungen (XX,XX €)"
- Pattern analog zu Admin-SummaryCard (`Card` + `CardHeader` + `CardTitle` + `CardContent`)

### 3.6 · Alle-Tische-Drawer mit Schnellsuche (K-11) (*depends on 3.1, 3.2*)

Neue Komponente `frontend/src/service/components/TischAuswahlDrawer.tsx`:
- Drawer (Vaul) öffnet von unten
- Oben: Suchfeld (`Input`) — clientseitige Filterung nach Tischname/-nummer
- Liste aller aktiven Tische:
  - Stern-Toggle (★/☆) zum Favorisieren/Entfavorisieren
  - Tischname + Saldo (kompakt)
  - Tap auf Tisch-Zeile → navigate zu `/service/tables/:id`
- Optimistic Update: Stern sofort umschalten, API-Call im Hintergrund
- K-11 Akzeptanzkriterien: Direkte Navigation per Eingabe der Tischnummer

### 3.7 · ServiceLayout anpassen

- Header-Text: "Tischauswahl" → "Meine Tische" (wenn auf Dashboard)
- Zurück-Navigation von Tisch-Detail weiterhin zu Dashboard

### 3.8 · Route-Anpassung

Minimal — `TableSelectionPage` wird zum `ServiceDashboard`, Route `/service/tables` bleibt.

**Relevante Dateien:**
- `frontend/src/service/TableSelectionPage.tsx` — Komplett umbauen
- `frontend/src/service/ServiceLayout.tsx` — Header anpassen
- `frontend/src/service/table/Tisch.ts` — Typen erweitern
- `frontend/src/service/table/TischBackend.ts` — API-Methoden
- `frontend/src/service/table/hooks.ts` — Neue Hooks
- `frontend/src/service/components/MeinTischCard.tsx` — Neue Komponente
- `frontend/src/service/components/EigeneUebersicht.tsx` — Neue Komponente
- `frontend/src/service/components/TischAuswahlDrawer.tsx` — Neue Komponente

**Verification Phase 3:**
- Vitest-Tests für neue Hooks (Mock-Backend)
- Manueller Test: Dashboard mit/ohne Favoriten, KPIs, Drawer-Suche, Stern-Toggle
- `make test-frontend`, `make lint`

---

## Phase 4 · Tests & End-to-End-Verification

1. `make check` — Schnelle Repo-Prüfung (Lint + Unit-Tests)
2. `make verify` — Vollständige Prüfung inkl. Integrationstests
3. Manueller Walkthrough:
   - Login als Servicekraft → leerer Dashboard-Zustand
   - Drawer öffnen → alle Tische sehen → 5 Tische favorisieren
   - Dashboard zeigt 5 Rich Cards mit vollem State
   - KPIs zeigen eigene Bestellungen/Zahlungen
   - Schnellsuche im Drawer → Tisch finden → navigieren
   - Favorit entfernen → Card verschwindet
   - Login als anderer User → eigene Favoriten sind unabhängig

---

## Scope-Grenzen

**Enthalten:**
- Tisch-Favoriten (CRUD, serverseitig)
- Dashboard mit Rich Cards (voller TischState pro Favorit)
- Eigene Übersicht KPIs auf Dashboard (R-06)
- Alle-Tische-Drawer mit Schnellsuche (K-11)
- Stern-Toggle zum Favorisieren im Drawer
- Dokumentations-Updates (Anforderungen, Language, Handbuch)

**Nicht enthalten:**
- Admin-Assignment von Tischen an Servicekräfte
- Auto-Refresh / Live-Updates auf dem Dashboard
- R-06 als eigene Seite mit Detailansicht
- Favoriten-Reihenfolge (Drag & Drop Sortierung)
- Favoriten für Admins auf dem Admin-Dashboard
