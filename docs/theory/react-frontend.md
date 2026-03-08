# React Frontend Architektur — Theorie

Dieses Dokument beschreibt allgemeine Architekturprinzipien für React-Frontends: Komponentenstruktur, State-Management, Backend-Integration, UI-Patterns und Design-Entscheidungen für wartbare, mobile-first Single-Page-Applications.

> **Verwandte Dokumente:**
>
> - [Go Backend Architektur](go-backend.md) — Backend-Architektur, API-Format
> - [DDD Theorie](ddd.md) — Domain-Driven Design (Bounded Contexts)
> - [Architektur-Übersicht](README.md) — Index aller Theorie-Dokumente

---

## Inhaltsverzeichnis

1. [Architekturprinzipien](#1-architekturprinzipien)
2. [Komponentenarchitektur und Atomic Design](#2-komponentenarchitektur-und-atomic-design)
3. [State-Management ohne Redux](#3-state-management-ohne-redux)
4. [Backend-Integration](#4-backend-integration)
5. [Routing und Guards](#5-routing-und-guards)
6. [Validierung mit Zod](#6-validierung-mit-zod)
7. [UI-Patterns und Bibliotheken](#7-ui-patterns-und-bibliotheken)
8. [Styling mit Tailwind CSS](#8-styling-mit-tailwind-css)
9. [Fehlerbehandlung im Frontend](#9-fehlerbehandlung-im-frontend)
10. [Empfohlene Design Patterns](#10-empfohlene-design-patterns)
11. [Anti-Patterns](#11-anti-patterns)
12. [Appendix: Anwendungsbeispiel (jotti)](#appendix-anwendungsbeispiel-jotti)
13. [Referenzen](#13-referenzen)

---

## 1. Architekturprinzipien

### Mobile-first

Mobile Web-Apps (z. B. POS-Systeme, Field-Service-Apps) werden primär auf Smartphones bedient. Daraus folgt:

- **Touch-first UI** — Große Touch-Targets, Bottom-Sheet-Drawers, keine Hover-Effekte
- **Responsive Minimal** — Primär Mobile, Desktop als Bonus (z. B. Admin-Bereich)
- **Offline-Toleranz** — Klare Fehlermeldungen bei Netzwerkausfall
- **Performance** — Schnelle Ladezeiten, minimaler Bundle-Size

### Kein globaler State-Store

Für kleinere Anwendungen mit wenigen Seiten kann bewusst auf Redux, Zustand, MobX oder ähnliche State-Store-Libraries verzichtet werden:

```
❌ Redux/Zustand/MobX     → Overkill für ~10 Seiten
✅ React Hooks + Singletons → Einfach, explizit, ausreichend
```

**Begründung:**

- Wenige Seiten mit überschaubarem State
- Server-State wird bei jeder Navigation frisch geladen
- Client-State (Auth-Token, UI-Zustand) passt in Singletons + Hooks
- Weniger Abhängigkeiten = weniger Wartung

### Backend ist Single Source of Truth

**Filterung, Aggregation und Aufbereitung gehören ins Backend.** Das Frontend zeigt an, was das Backend liefert.

```
❌ Frontend: fetch('/events') → filter → sort → aggregate → display
✅ Frontend: fetch('/api/get-order') → display
```

---

## 2. Komponentenarchitektur und Atomic Design

### Atomic Design (Brad Frost)

Atomic Design strukturiert UI-Komponenten in fünf Ebenen, von klein (Atom) nach groß (Seite):

```
┌───────────────────────────────────────────────────────────┐
│  Seiten (Pages)                                           │
│  src/features/*/pages/                                    │
│  ┌─────────────────────────────────────────────────────┐  │
│  │  Features (Organisms)                                │  │
│  │  src/features/*/components/                          │  │
│  │  ┌───────────────────────────────────────────────┐   │  │
│  │  │  Gemeinsame Komponenten (Molecules)            │   │  │
│  │  │  src/components/common/                        │   │  │
│  │  │  ┌─────────────────────────────────────────┐   │   │  │
│  │  │  │  UI-Primitives (Atoms)                   │   │   │  │
│  │  │  │  src/components/ui/ (shadcn/ui)          │   │   │  │
│  │  │  └─────────────────────────────────────────┘   │   │  │
│  │  └───────────────────────────────────────────────┘   │  │
│  └─────────────────────────────────────────────────────┘  │
└───────────────────────────────────────────────────────────┘
```

### Ebenen im Atomic Design

| Ebene         | Atomic Design          | Typischer Pfad             | Beispiele                     |
| ------------- | ---------------------- | -------------------------- | ----------------------------- |
| **Atoms**     | UI-Primitives          | `src/components/ui/`       | Button, Input, Badge, Card    |
| **Molecules** | Gemeinsame Komponenten | `src/components/common/`   | LoadingSpinner, ErrorDisplay  |
| **Organisms** | Feature-Komponenten    | `src/features/*/components/` | OrderCard, OrderDrawer      |
| **Pages**     | Seiten                 | `src/features/*/pages/`    | OrderOverview, OrderDetail    |
| **Templates** | Layouts                | `src/App.tsx`              | Root-Layout mit ThemeProvider |

### Komponenten-Prinzipien

**Single Responsibility:** Jede Komponente hat eine klare Aufgabe.

```tsx
// RICHTIG: Klare Verantwortlichkeit
<OrderCard order={order} onSelect={handleSelect} />
<OrderDrawer items={items} onSubmit={handleSubmit} />

// FALSCH: God Component
<OrderManager /* macht alles: laden, anzeigen, bestellen, bezahlen */ />
```

**Komposition über Vererbung:** React-Komponenten werden zusammengesetzt, nicht vererbt.

```tsx
// RICHTIG: Komposition
<Drawer>
  <DrawerContent>
    <ItemList items={items} />
    <TotalPriceDisplay cents={total} />
    <ConfirmButton onClick={onConfirm} />
  </DrawerContent>
</Drawer>
```

**Props Down, Events Up:** Daten fließen von oben nach unten (Props), Aktionen von unten nach oben (Callbacks).

---

## 3. State-Management ohne Redux

### State-Kategorien

| Kategorie         | Lösung                     | Beispiele                         |
| ----------------- | -------------------------- | --------------------------------- |
| **Server-State**  | `useFetch` Hook            | Bestell-Daten, Produkte, Benutzer |
| **Auth-State**    | Singleton (`Auth.ts`)      | JWT-Token, Rolle, UserID          |
| **UI-State**      | `useState`                 | Drawer offen/zu, Formular-Werte   |
| **Derived State** | Berechnung aus Props/State | Gesamtpreis, gefilterte Listen    |

### useFetch Hook

Zentraler Hook für alle Backend-Anfragen:

```tsx
// src/lib/useFetch.ts
function useFetch<T>(fetcher: () => Promise<T>) {
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const result = await fetcher();
      setData(result);
    } catch (e) {
      setError(e as Error);
    } finally {
      setLoading(false);
    }
  }, [fetcher]);

  useEffect(() => {
    load();
  }, [load]);

  return { data, loading, error, reload: load };
}

// Nutzung
const {
  data: order,
  loading,
  error,
  reload,
} = useFetch(() => orderBackend.getOrder(orderId));
```

### Auth Singleton

```tsx
// src/lib/Auth.ts
class Auth {
    getToken(): string | null { ... }
    getRole(): string | null { ... }
    getUserID(): number | null { ... }
    isAuthenticated(): boolean { ... }
    logout(): void { ... }
}
```

Kein Context/Provider nötig — Auth ist ein Singleton, weil es **einen** authentifizierten Benutzer pro Browser-Tab gibt.

### Wann Context, wann nicht?

| Szenario                   | Lösung                 | Begründung                             |
| -------------------------- | ---------------------- | -------------------------------------- |
| Globales Theme (Dark Mode) | Context/Provider       | Muss den gesamten Baum durchdringen    |
| Auth-Token                 | Singleton              | Kein Re-Render bei Token-Refresh nötig |
| Formular-State             | `useState` im Formular | Lokal, nicht geteilt                   |
| Server-Daten               | `useFetch` pro Seite   | Bei Navigation frisch laden            |
| Drawer-Zustand             | `useState` in Parent   | Lokal, 1-2 Ebenen tief                 |

---

## 4. Backend-Integration

### BackendClient-Interface

Alle Backend-Aufrufe laufen über das `BackendClient`-Interface — niemals direkt `fetch()`:

```tsx
// src/lib/Backend.ts
interface BackendClient {
  post<T>(path: string, body?: unknown): Promise<T>;
}

class Backend implements BackendClient {
  async post<T>(path: string, body?: unknown): Promise<T> {
    const response = await fetch(path, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${Auth.getToken()}`,
      },
      body: body ? JSON.stringify(body) : undefined,
    });

    // 401 → Automatisch ausloggen und zu /login weiterleiten
    if (response.status === 401) {
      Auth.logout();
      window.location.href = "/login";
      throw new Error("Unauthorized");
    }

    if (!response.ok) {
      const error = await response.json();
      throw new BackendError(error.code, error.details);
    }

    return response.json();
  }
}
```

### Domain-Backend-Klassen

Für jeden Fachbereich eine Backend-Klasse:

```tsx
// src/features/orders/OrderBackend.ts
class OrderBackend {
  constructor(private client: BackendClient) {}

  getOrder(id: number) {
    return this.client.post<Order>("/api/get-order", { orderId: id });
  }

  submitOrder(orderId: number, items: OrderItem[], comment?: string) {
    return this.client.post("/api/submit-order", {
      orderId,
      items,
      comment,
    });
  }
}
```

### 401-Interceptor

Der `Backend.post()`-Interceptor fängt 401-Responses ab und leitet automatisch zum Login weiter. **Kein manuelles 401-Handling in Komponenten nötig.**

---

## 5. Routing und Guards

### Route-Struktur

```tsx
// src/routes.ts
const routes = [
  // Auth (öffentlich)
  { path: "/login", element: <LoginPage /> },

  // Admin (nur admin)
  {
    path: "/admin/*",
    loader: AdminGuard,
    children: [
      { path: "products", element: <ProductsPage /> },
      { path: "users", element: <UsersPage /> },
    ],
  },

  // Hauptbereich (authentifiziert)
  {
    path: "/app/*",
    loader: AppGuard,
    children: [
      { path: "", element: <Overview /> },
      { path: ":id", element: <DetailPage /> },
    ],
  },
];
```

### Guards (React Router Loaders)

Guards prüfen vor dem Rendern, ob der Benutzer zugriffsberechtigt ist:

```tsx
// AdminGuard: Nur admin darf zugreifen
function AdminGuard() {
  if (!Auth.isAuthenticated()) redirect("/login");
  if (Auth.getRole() !== "admin") redirect("/app");
  return null;
}

// AppGuard: Alle authentifizierten Rollen
function AppGuard() {
  if (!Auth.isAuthenticated()) redirect("/login");
  return null;
}
```

### Rollenbasiertes Routing

| Rolle      | Zugriff              | Redirect nach Login  |
| ---------- | -------------------- | -------------------- |
| `admin`    | Admin + Hauptbereich | `/admin` oder `/app` |
| `user`     | Hauptbereich         | `/app`               |

---

## 6. Validierung mit Zod

### Client-seitige Validierung

Zod validiert Benutzer-Eingaben **vor** dem Backend-Request:

```tsx
import { z } from "zod";

const OrderSchema = z.object({
  orderId: z.number().min(1),
  items: z
    .array(
      z.object({
        id: z.number(),
        name: z.string(),
        priceCents: z.number().min(0),
        quantity: z.number().min(1),
      }),
    )
    .min(1, "Mindestens ein Eintrag"),
  comment: z.string().max(500).optional(),
});

type Order = z.infer<typeof OrderSchema>;
```

### Zwei-Schichten-Validierung

```
┌─────────────────────────────────────────────────────┐
│  Frontend (Zod)                                     │
│  • Schnelles Feedback ohne Netzwerk-Roundtrip       │
│  • UI-spezifische Fehlermeldungen (deutsch)         │
│  • Nie allein als Sicherheit — Backend validiert    │
└──────────────────────────┬──────────────────────────┘
                           │ POST (JSON)
┌──────────────────────────▼──────────────────────────┐
│  Backend (zog)                                      │
│  • Single Source of Truth für Validierung            │
│  • Schützt vor manipulierten Requests               │
│  • Gleiche Regeln wie Frontend                       │
└─────────────────────────────────────────────────────┘
```

---

## 7. UI-Patterns und Bibliotheken

### shadcn/ui

shadcn/ui liefert **kopierte, nicht importierte** UI-Komponenten (Radix-basiert):

```tsx
// Komponenten liegen in src/components/ui/ und gehören zum Projekt
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
```

**Stil:** `new-york` (schärfere Ecken, kompakteres Design)

**Vorteile:**

- Volle Kontrolle über den Code (keine Black Box)
- Radix-Primitives für Accessibility
- Tailwind-basiertes Styling
- Keine Runtime-Dependency (Code ist Teil des Projekts)

### Lucide React (Icons)

```tsx
import { Plus, Minus, Trash2, Check } from "lucide-react";
<Button>
  <Plus className="h-4 w-4" /> Hinzufügen
</Button>;
```

### Sonner (Toasts)

Alle mutativen Aktionen zeigen bei Fehler einen Toast:

```tsx
import { toast } from "sonner";

try {
  await backend.submitOrder(orderId, items);
  toast.success("Bestellung aufgegeben");
} catch (error) {
  toast.error("Bestellung fehlgeschlagen");
}
```

### Vaul (Drawers)

Bottom-Sheet-Drawers für mobile Interaktionen:

```tsx
import { Drawer, DrawerContent, DrawerTrigger } from "vaul";

<Drawer>
  <DrawerTrigger>Bestellen</DrawerTrigger>
  <DrawerContent>
    <Summary items={items} />
    <Button onClick={handleConfirm}>Bestätigen</Button>
  </DrawerContent>
</Drawer>;
```

### Drawer-Pattern

Aktionen wie Bestellen, Bezahlen oder Stornieren öffnen Bottom-Sheet-Drawers mit Zusammenfassung:

```
1. Benutzer wählt Positionen/Mengen
2. Drawer öffnet sich mit Zusammenfassung
3. Benutzer bestätigt oder bricht ab
4. Backend-Request + Toast-Feedback
```

---

## 8. Styling mit Tailwind CSS

### Setup

Tailwind CSS 4 via `@tailwindcss/vite` (keine `tailwind.config.js` nötig):

```ts
// vite.config.ts
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  plugins: [react(), tailwindcss()],
});
```

### CSS-Variablen

Design-Tokens als CSS-Variablen in `src/index.css`:

```css
:root {
  --background: 0 0% 100%;
  --foreground: 224 71.4% 4.1%;
  --primary: 262.1 83.3% 57.8%; /* Violet */
  --primary-foreground: 210 20% 98%;
  /* ... */
}

.dark {
  --background: 224 71.4% 4.1%;
  --foreground: 210 20% 98%;
  /* ... */
}
```

### cn() Utility

Kombination aus `clsx` und `tailwind-merge`:

```tsx
import { cn } from "@/lib/utils";

<div
  className={cn(
    "rounded-lg border p-4",
    isActive && "border-primary bg-primary/10",
    isDisabled && "opacity-50 cursor-not-allowed",
  )}
/>;
```

### Geldbeträge anzeigen

```tsx
import { formatCents } from "@/lib/utils";

// formatCents(350) → "3,50 €"
// formatCents(0)   → "0,00 €"

<span>{formatCents(saldoCents)}</span>;
```

**Nie inline formatieren** — immer `formatCents()` verwenden.

---

## 9. Fehlerbehandlung im Frontend

### Fehler-Strategien

| Fehlerart              | Strategie                 | UI-Feedback                                |
| ---------------------- | ------------------------- | ------------------------------------------ |
| **401 Unauthorized**   | Automatisch (Interceptor) | Redirect zu `/login`                       |
| **Validierungsfehler** | Inline-Anzeige            | Feld-basierte Fehlermeldungen              |
| **Netzwerkfehler**     | Toast + Retry             | `toast.error('Netzwerkfehler')`            |
| **Backend-Fehler**     | Toast                     | `toast.error('Aktion fehlgeschlagen')`     |
| **Unerwartete Fehler** | Error Boundary            | Fallback-UI                                |

### Error Boundary (React)

```tsx
class ErrorBoundary extends React.Component {
  state = { hasError: false };

  static getDerivedStateFromError() {
    return { hasError: true };
  }

  render() {
    if (this.state.hasError) {
      return <FallbackUI />;
    }
    return this.props.children;
  }
}
```

---

## 10. Empfohlene Design Patterns

### 10.1 Container/Presentational Pattern

Trennung von Datenlogik und UI:

```tsx
// Container: Holt Daten, verwaltet State
function OrderDetailContainer({ orderId }: { orderId: number }) {
  const { data: order, loading } = useFetch(() => backend.getOrder(orderId));
  if (loading) return <LoadingSpinner />;
  return <OrderDetailView order={order!} />;
}

// Presentational: Reines UI, keine Side Effects
function OrderDetailView({ order }: { order: Order }) {
  return (
    <Card>
      <CardHeader>{order.name}</CardHeader>
      <CardContent>Saldo: {formatCents(order.totalCents)}</CardContent>
    </Card>
  );
}
```

### 10.2 Custom Hook Pattern

Logik in wiederverwendbare Hooks extrahieren:

```tsx
// Hook kapselt Lade- und Mutationslogik
function useOrder(orderId: number) {
    const [items, setItems] = useState<OrderItem[]>([]);

    const addItem = (variant: Variant, quantity: number) => { ... };
    const removeItem = (variantId: number) => { ... };
    const submit = async () => {
        await backend.submitOrder(orderId, items);
    };
    const totalCents = items.reduce(
        (sum, p) => sum + p.priceCents * p.quantity, 0
    );

    return { items, addItem, removeItem, submit, totalCents };
}
```

### 10.3 Render Props / Children Pattern

Flexible Komposition:

```tsx
<DataLoader fetcher={() => backend.getOrders()}>
  {(orders) => <OrderGrid orders={orders} onSelect={handleSelect} />}
</DataLoader>
```

### 10.4 Compound Components

Zusammengehörige Komponenten als Gruppe:

```tsx
// shadcn/ui nutzt dieses Pattern extensiv
<Card>
  <CardHeader>
    <CardTitle>Order #1</CardTitle>
  </CardHeader>
  <CardContent>
    <ItemList items={items} />
  </CardContent>
  <CardFooter>
    <Button>Bestellen</Button>
  </CardFooter>
</Card>
```

---

## 11. Anti-Patterns

### 11.1 Direktes fetch()

```tsx
// FALSCH: Direkt fetch() aufrufen
const response = await fetch('/api/get-order', { ... });

// RICHTIG: Über Backend-Klasse
const order = await orderBackend.getOrder(orderId);
```

### 11.2 Frontend-Filterung

```tsx
// FALSCH: Im Frontend filtern
const unpaid = allItems.filter((p) => !p.paid);

// RICHTIG: Backend liefert bereits gefiltert
const unpaid = await backend.getUnpaidItems(orderId);
```

### 11.3 Prop Drilling über viele Ebenen

```tsx
// FALSCH: Props durch 5 Ebenen durchreichen
<App user={user}>
  <Layout user={user}>
    <Page user={user}>
      <Section user={user}>
        <UserName user={user} />

// RICHTIG: Context oder Singleton für globale Daten
const role = Auth.getRole();
```

### 11.4 Inline Geldbeträge formatieren

```tsx
// FALSCH
<span>{(cents / 100).toFixed(2)} €</span>
<span>{cents / 100} €</span>

// RICHTIG
<span>{formatCents(cents)}</span>
```

### 11.5 Globaler State-Store

```tsx
// FALSCH: Redux/Zustand für wenige Seiten
const store = createStore({ orders: [], products: [], user: null, ... });

// RICHTIG: Lokale Hooks + useFetch + Singletons
const { data: orders } = useFetch(() => backend.getOrders());
```

---

## Appendix: Anwendungsbeispiel (jotti)

Dieser Anhang zeigt, wie die oben beschriebenen Prinzipien im **jotti**-Projekt (POS-System für Vereinsfeste) konkret umgesetzt werden.

### A.1 Architekturprinzipien in jotti

**Mobile-first:** jotti ist ein POS-System für Smartphones. Servicekräfte nehmen Bestellungen **auf dem Handy** auf.

**Kein globaler State-Store:** jotti verzichtet bewusst auf Redux/Zustand/MobX — die App hat ~10 Seiten mit überschaubarem State. Server-State (Tisch-Daten) wird bei jeder Navigation frisch geladen.

**Backend ist Single Source of Truth:** Regel #10 aus AGENTS.md: Filterung, Aggregation und Aufbereitung gehören ins Backend.

```
❌ Frontend: fetch('/events') → filter → sort → aggregate → display
✅ Frontend: fetch('/get-tisch-unbezahlt') → display
```

### A.2 Komponentenarchitektur in jotti

#### Verzeichnisstruktur

```
┌───────────────────────────────────────────────────────────┐
│  Seiten (Pages)                                           │
│  src/service/pages/, src/admin/pages/                     │
│  ┌─────────────────────────────────────────────────────┐  │
│  │  Features (Organisms)                                │  │
│  │  src/service/components/, src/admin/components/      │  │
│  │  ┌───────────────────────────────────────────────┐   │  │
│  │  │  Gemeinsame Komponenten (Molecules)            │   │  │
│  │  │  src/components/common/                        │   │  │
│  │  │  ┌─────────────────────────────────────────┐   │   │  │
│  │  │  │  UI-Primitives (Atoms)                   │   │   │  │
│  │  │  │  src/components/ui/ (shadcn/ui)          │   │   │  │
│  │  │  └─────────────────────────────────────────┘   │   │  │
│  │  └───────────────────────────────────────────────┘   │  │
│  └─────────────────────────────────────────────────────┘  │
└───────────────────────────────────────────────────────────┘
```

#### Ebenen in jotti

| Ebene         | Atomic Design          | jotti-Pfad                | Beispiele                     |
| ------------- | ---------------------- | ------------------------- | ----------------------------- |
| **Atoms**     | UI-Primitives          | `src/components/ui/`      | Button, Input, Badge, Card    |
| **Molecules** | Gemeinsame Komponenten | `src/components/common/`  | LoadingSpinner, ErrorDisplay  |
| **Organisms** | Feature-Komponenten    | `src/service/components/` | TischKarte, BestellungDrawer  |
| **Pages**     | Seiten                 | `src/service/pages/`      | TischÜbersicht, TischDetail   |
| **Templates** | Layouts                | `src/App.tsx`             | Root-Layout mit ThemeProvider |

#### Komponenten-Beispiele

```tsx
<TischKarte tisch={tisch} onSelect={handleSelect} />
<BestellungDrawer positionen={positionen} onSubmit={handleSubmit} />

<Drawer>
  <DrawerContent>
    <PositionenListe positionen={positionen} />
    <GesamtPreisAnzeige cents={total} />
    <BestätigenButton onClick={onConfirm} />
  </DrawerContent>
</Drawer>
```

### A.3 State-Management in jotti

| Kategorie         | Lösung                     | jotti-Beispiele                 |
| ----------------- | -------------------------- | ------------------------------- |
| **Server-State**  | `useFetch` Hook            | Tisch-Daten, Produkte, Benutzer |
| **Auth-State**    | Singleton (`Auth.ts`)      | JWT-Token, Rolle, UserID        |
| **UI-State**      | `useState`                 | Drawer offen/zu, Formular-Werte |
| **Derived State** | Berechnung aus Props/State | Gesamtpreis, gefilterte Listen  |

```tsx
const {
  data: tisch,
  loading,
  error,
  reload,
} = useFetch(() => tischBackend.getTisch(tischId));
```

### A.4 Backend-Integration in jotti

Domain-Backend-Klasse für den Tisch-Workflow:

```tsx
// src/service/TischBackend.ts
class TischBackend {
  constructor(private client: BackendClient) {}

  getTisch(id: number) {
    return this.client.post<Tisch>("/service/get-tisch", { tischId: id });
  }

  bestellungAufgeben(
    tischId: number,
    positionen: Position[],
    comment?: string,
  ) {
    return this.client.post("/service/bestellung-aufgeben", {
      tischId,
      positionen,
      comment,
    });
  }
}
```

### A.5 Routing in jotti

#### Route-Struktur

```tsx
// src/routes.ts
const routes = [
  // Auth (öffentlich)
  { path: "/login", element: <LoginPage /> },
  { path: "/set-password", element: <SetPasswordPage /> },

  // Admin (nur admin)
  {
    path: "/admin/*",
    loader: AdminGuard,
    children: [
      { path: "produkte", element: <ProduktePage /> },
      { path: "tische", element: <TischePage /> },
      { path: "benutzer", element: <BenutzerPage /> },
    ],
  },

  // Service (admin, senior_service, service)
  {
    path: "/service/*",
    loader: ServiceGuard,
    children: [
      { path: "", element: <TischÜbersicht /> },
      { path: ":tischId", element: <TischDetail /> },
    ],
  },
];
```

#### Guards

```tsx
function AdminGuard() {
  if (!Auth.isAuthenticated()) redirect("/login");
  if (Auth.getRole() !== "admin") redirect("/service");
  return null;
}

function ServiceGuard() {
  if (!Auth.isAuthenticated()) redirect("/login");
  return null;
}
```

#### Rollenbasiertes Routing

| Rolle            | Zugriff                     | Redirect nach Login      |
| ---------------- | --------------------------- | ------------------------ |
| `admin`          | Admin + Service             | `/admin` oder `/service` |
| `senior_service` | Service (inkl. Stornierung) | `/service`               |
| `service`        | Service (ohne Stornierung)  | `/service`               |

### A.6 Validierung in jotti

```tsx
const BestellungSchema = z.object({
  tischId: z.number().min(1),
  positionen: z
    .array(
      z.object({
        id: z.number(),
        name: z.string(),
        preisCents: z.number().min(0),
        quantity: z.number().min(1),
      }),
    )
    .min(1, "Mindestens eine Position"),
  comment: z.string().max(500).optional(),
});

type Bestellung = z.infer<typeof BestellungSchema>;
```

### A.7 Drawer-Pattern in jotti

Bestellen, Bezahlen, Stornieren und Liefern öffnen Bottom-Sheet-Drawers mit Zusammenfassung.

Hilfsfunktionen in `src/service/components/table/drawerUtils.ts`:

- `selectPositionen()` — Positionen auswählen/abwählen
- `calculateTotalPrice()` — Gesamtpreis berechnen

### A.8 Design Patterns in jotti

#### Container/Presentational

```tsx
function TischDetailContainer({ tischId }: { tischId: number }) {
  const { data: tisch, loading } = useFetch(() => backend.getTisch(tischId));
  if (loading) return <LoadingSpinner />;
  return <TischDetailView tisch={tisch!} />;
}

function TischDetailView({ tisch }: { tisch: Tisch }) {
  return (
    <Card>
      <CardHeader>{tisch.name}</CardHeader>
      <CardContent>Saldo: {formatCents(tisch.saldoCents)}</CardContent>
    </Card>
  );
}
```

#### Custom Hook

```tsx
function useBestellung(tischId: number) {
    const [positionen, setPositionen] = useState<Position[]>([]);

    const addPosition = (variante: Variante, quantity: number) => { ... };
    const removePosition = (variantId: number) => { ... };
    const submit = async () => {
        await backend.bestellungAufgeben(tischId, positionen);
    };
    const totalCents = positionen.reduce(
        (sum, p) => sum + p.preisCents * p.quantity, 0
    );

    return { positionen, addPosition, removePosition, submit, totalCents };
}
```

#### Weitere Beispiele

```tsx
// Render Props
<DataLoader fetcher={() => backend.getTische()}>
  {(tische) => <TischGrid tische={tische} onSelect={handleSelect} />}
</DataLoader>

// Compound Components
<Card>
  <CardHeader>
    <CardTitle>Tisch 1</CardTitle>
  </CardHeader>
  <CardContent>
    <PositionenListe positionen={positionen} />
  </CardContent>
  <CardFooter>
    <Button>Bestellen</Button>
  </CardFooter>
</Card>
```

### A.9 Anti-Patterns (jotti-Beispiele)

```tsx
// FALSCH: Direkt fetch()
const response = await fetch('/service/get-tisch', { ... });
// RICHTIG: Über Backend-Klasse
const tisch = await tischBackend.getTisch(tischId);

// FALSCH: Im Frontend filtern
const unbezahlt = allePositionen.filter((p) => !p.bezahlt);
// RICHTIG: Backend liefert bereits gefiltert
const unbezahlt = await backend.getTischUnbezahlt(tischId);

// FALSCH: Redux/Zustand für wenige Seiten
const store = createStore({ tische: [], produkte: [], user: null, ... });
// RICHTIG: Lokale Hooks + useFetch + Singletons
const { data: tische } = useFetch(() => backend.getTische());
```

```tsx
// Sonner-Toasts in jotti
try {
  await backend.bestellungAufgeben(tischId, positionen);
  toast.success("Bestellung aufgegeben");
} catch (error) {
  toast.error("Bestellung fehlgeschlagen");
}
```

---

## 13. Referenzen

### React-Architektur

- [21 Fantastic React Design Patterns](https://www.perssondennis.com/articles/21-fantastic-react-design-patterns-and-when-to-use-them) — Pattern-Katalog
- [Guide to Modern Frontend Architecture Patterns](https://blog.logrocket.com/guide-modern-frontend-architecture-patterns/) — Feature-Sliced Architecture
- [Mastering Atomic Design in React](https://javascript.plainenglish.io/mastering-atomic-design-a-step-by-step-guide-to-building-scalable-ui-components-60b0d2a94cc3) — Atomic Design Praxis

### Libraries

- [shadcn/ui](https://ui.shadcn.com/) — UI-Komponenten (Radix + Tailwind)
- [Zod](https://zod.dev/) — TypeScript-first Schema-Validierung
- [React Router](https://reactrouter.com/) — Client-Side Routing
- [Sonner](https://sonner.emilkowal.dev/) — Toast-Notifications
- [Vaul](https://vaul.emilkowal.dev/) — Drawer-Komponente
- [Lucide](https://lucide.dev/) — Icon-Library
- [Tailwind CSS](https://tailwindcss.com/) — Utility-first CSS

### Projekt-intern

- `src/lib/Backend.ts` — BackendClient-Interface und 401-Interceptor
- `src/lib/useFetch.ts` — Generischer Daten-Hook
- `src/lib/utils.ts` — `cn()` und `formatCents()`
- `src/lib/Auth.ts` — Auth-Singleton
