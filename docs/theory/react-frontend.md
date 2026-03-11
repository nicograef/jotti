# React Frontend Architektur — Theorie

Dieses Dokument beschreibt allgemeine Architekturprinzipien für React-Frontends: Komponentenstruktur, State-Management, Backend-Integration, UI-Patterns und Design-Entscheidungen für wartbare, mobile-first Single-Page-Applications.

---

## Inhaltsverzeichnis

1. [Architekturprinzipien](#1-architekturprinzipien)
2. [Komponentenarchitektur und Atomic Design](#2-komponentenarchitektur-und-atomic-design)
3. [Frontend Architecture Patterns](#3-frontend-architecture-patterns)
4. [State-Management Landscape](#4-state-management-landscape)
5. [React Design Patterns Katalog](#5-react-design-patterns-katalog)
6. [Backend-Integration](#6-backend-integration)
7. [Routing und Guards](#7-routing-und-guards)
8. [Validierung mit Zod](#8-validierung-mit-zod)
9. [UI-Patterns und Bibliotheken](#9-ui-patterns-und-bibliotheken)
10. [Styling mit Tailwind CSS](#10-styling-mit-tailwind-css)
11. [Testing-Strategien](#11-testing-strategien)
12. [Performance-Patterns](#12-performance-patterns)
13. [Accessibility (a11y)](#13-accessibility-a11y)
14. [TypeScript-Patterns für React](#14-typescript-patterns-für-react)
15. [Fehlerbehandlung im Frontend](#15-fehlerbehandlung-im-frontend)
16. [Anti-Patterns](#16-anti-patterns)
17. [Referenzen](#17-referenzen)

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

| Ebene         | Atomic Design          | Typischer Pfad               | Beispiele                     |
| ------------- | ---------------------- | ---------------------------- | ----------------------------- |
| **Atoms**     | UI-Primitives          | `src/components/ui/`         | Button, Input, Badge, Card    |
| **Molecules** | Gemeinsame Komponenten | `src/components/common/`     | LoadingSpinner, ErrorDisplay  |
| **Organisms** | Feature-Komponenten    | `src/features/*/components/` | OrderCard, OrderDrawer        |
| **Pages**     | Seiten                 | `src/features/*/pages/`      | OrderOverview, OrderDetail    |
| **Templates** | Layouts                | `src/App.tsx`                | Root-Layout mit ThemeProvider |

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

## 3. Frontend Architecture Patterns

Für React-Frontends stehen mehrere Architektur-Muster zur Verfügung. Die Wahl hängt von Projektgröße, Teamgröße und Skalierungsanforderungen ab.

### 3.1 Monolithic Architecture (Single-Page Application)

Die gesamte Anwendung lebt in einem einzigen Code-Repository. Alle Seiten, Komponenten und Logik sind eng zusammen.

```
Stärken:
  + Einfach zu verstehen und zu deployen
  + Kein Overhead durch Sub-Projekte
  + Schnelle initiale Entwicklung
  + Einfaches Debugging (alles in einem Repo)

Schwächen:
  - Wächst mit der Codebasis an Komplexität
  - Merge-Konflikte bei großen Teams
  - Deployment startet die gesamte App neu
```

**Geeignet für:** Kleine bis mittelgroße Projekte mit 1–10 Entwicklern.

### 3.2 Modular Architecture (Feature-Sliced Design)

Die Codebasis wird nach Fachdomänen in unabhängige Module aufgeteilt. Jedes Modul hat eigene Komponenten, Hooks, Services und Tests.

```
src/
├── features/
│   ├── orders/          ← Modul: Bestellungen
│   │   ├── components/
│   │   ├── hooks/
│   │   ├── backend.ts
│   │   └── types.ts
│   ├── products/        ← Modul: Produkte
│   └── users/           ← Modul: Benutzer
├── shared/              ← geteilte Utilities
└── app/                 ← App-Konfiguration
```

Feature-Sliced Design (FSD) ist eine Variante: Module sind nach Schichten (app → pages → widgets → features → entities → shared) und Domänen strukturiert.

**Geeignet für:** Mittelgroße Projekte mit 5–20 Entwicklern und mehreren Fachdomänen.

### 3.3 Component-Based Architecture

Die fundamentale Architektur aller modernen Frontend-Frameworks (React, Vue, Angular). Die Anwendung wird als Baum von Komponenten modelliert, jede mit eigener Logik, UI und lokalem State.

```
App
├── Header (Molecule)
│   ├── Logo (Atom)
│   └── Navigation (Molecule)
├── Main
│   └── OrderPage (Page)
│       ├── OrderList (Organism)
│       │   └── OrderCard (Molecule) × n
│       └── OrderDrawer (Organism)
└── Footer (Molecule)
```

**Geeignet für:** Alle React-Projekte. Wird immer zusammen mit anderen Architekturmustern verwendet.

### 3.4 Micro-Frontend Architecture

Die Anwendung wird in voneinander unabhängige Frontend-Projekte (Micro-Frontends) aufgeteilt. Jedes Micro-Frontend wird von einem eigenen Team entwickelt, getestet und deployed.

```
Host App (Shell)
├── lädt Micro-Frontend A (via Module Federation)
├── lädt Micro-Frontend B (via Single-SPA)
└── stellt gemeinsames Theme/Routing bereit
```

**Technologien:** Webpack Module Federation, Single-SPA, Vite Federation.

```
Stärken:
  + Voneinander unabhängige Deployments
  + Teams können unterschiedliche Tech-Stacks nutzen
  + Skaliert für große Organisationen

Schwächen:
  - Hohe initiale Komplexität
  - UI/UX-Konsistenz schwerer sicherzustellen
  - Overhead bei Kommunikation zwischen Micro-Frontends
```

**Geeignet für:** Große Enterprise-Anwendungen mit mehreren unabhängigen Teams (20+ Entwickler).

### 3.5 Flux Architecture (und Redux)

Flux (entwickelt von Meta) zentralisiert den Application State in einem Store und erzwingt unidirektionalen Datenfluss:

```
View → Action → Dispatcher → Store → View
```

Redux ist eine vereinfachte Flux-Implementierung:

```
View → dispatch(action) → Reducer → Store → View
```

**Vorteile:** Vorhersagbarer State, gutes Debugging (Time-Travel), klare Trennung von Logik und UI.

**Nachteile:** Viel Boilerplate, Overkill für kleine Apps, steile Lernkurve.

**Modernes Redux:** Redux Toolkit (RTK) reduziert den Boilerplate erheblich. RTK Query ersetzt useFetch für Server-State.

### 3.6 Vergleichsmatrix

| Architektur        | Teamgröße | Komplexität | Skalierbarkeit | Deployment | Empfehlung                     |
| ------------------ | --------- | ----------- | -------------- | ---------- | ------------------------------ |
| **Monolithic**     | Klein     | Niedrig     | Mittel         | Einfach    | Startpunkt, kleine Apps        |
| **Modular (FSD)**  | Mittel    | Mittel      | Hoch           | Einfach    | Empfohlen für wachsende Teams  |
| **Micro-Frontend** | Groß      | Hoch        | Sehr hoch      | Unabhängig | Enterprise mit vielen Teams    |
| **Flux/Redux**     | Beliebig  | Mittel-hoch | Hoch           | Einfach    | Komplexer globaler State nötig |

---

## 4. State-Management Landscape

State-Management ist eines der komplexesten Themen in React. Die richtige Lösung hängt von der Art des States ab.

### 4.1 State-Kategorien

| Kategorie         | Definition                     | Lösung                            | Beispiele                         |
| ----------------- | ------------------------------ | --------------------------------- | --------------------------------- |
| **Server State**  | Remote-Daten, async, cacheable | TanStack Query, SWR, useFetch     | API-Daten, Listen, Einzel-Objekte |
| **Client State**  | Lokaler UI-State               | useState, useReducer, Zustand     | Drawer offen, Formular-Werte      |
| **Auth State**    | Nutzer-Session, Token, Rollen  | Singleton, Context                | JWT-Token, Rolle, UserID          |
| **Form State**    | Formular-Eingaben, Validierung | React Hook Form, Formik, useState | Eingabefelder, Fehler-Anzeige     |
| **URL State**     | State in der URL (navigierbar) | React Router, nuqs                | Filter, Tabs, Pagination          |
| **Derived State** | Berechnet aus anderem State    | Berechnung in Render/useMemo      | Gesamtpreis, gefilterte Listen    |

### 4.2 Server State: TanStack Query vs. SWR vs. useFetch

Für Remote-Daten ist spezialisiertes Tooling besser als rohes `useState` + `useEffect`:

**TanStack Query (React Query)** — der Standard für Server-State in React:

```tsx
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";

// Daten laden mit Caching, Refetch, Loading/Error-States
function useOrders() {
  return useQuery({
    queryKey: ["orders"],
    queryFn: () => orderBackend.getOrders(),
    staleTime: 5 * 60 * 1000, // 5 Minuten cached
  });
}

// Mutation mit automatischem Cache-Invalidierung
function useSubmitOrder() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: OrderInput) => orderBackend.submitOrder(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["orders"] });
    },
  });
}

// Nutzung in Komponente
function OrdersPage() {
  const { data: orders, isLoading, error } = useOrders();
  const { mutate: submitOrder } = useSubmitOrder();

  if (isLoading) return <LoadingSpinner />;
  if (error) return <ErrorDisplay error={error} />;

  return <OrderList orders={orders!} onSubmit={submitOrder} />;
}
```

**SWR** — leichtgewichtige Alternative von Vercel:

```tsx
import useSWR from "swr";

function useOrder(id: number) {
  return useSWR(["order", id], () => orderBackend.getOrder(id));
}
```

**Selbst gebauter useFetch** — einfachste Lösung ohne Caching, für kleine Apps ausreichend:

```tsx
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
```

### 4.3 Client State: useState vs. useReducer vs. Zustand vs. Jotai

**useState** — für einfachen, lokalen State:

```tsx
const [isOpen, setIsOpen] = useState(false);
const [count, setCount] = useState(0);
```

**useReducer** — für komplexen State mit mehreren Aktionen:

```tsx
type Action =
  | { type: "ADD_ITEM"; item: CartItem }
  | { type: "REMOVE_ITEM"; id: number }
  | { type: "CLEAR" };

function cartReducer(state: CartItem[], action: Action): CartItem[] {
  switch (action.type) {
    case "ADD_ITEM":
      return [...state, action.item];
    case "REMOVE_ITEM":
      return state.filter((i) => i.id !== action.id);
    case "CLEAR":
      return [];
  }
}

const [cart, dispatch] = useReducer(cartReducer, []);
```

**Zustand** — minimaler globaler State-Store ohne Provider:

```tsx
import { create } from "zustand";

interface CartStore {
  items: CartItem[];
  addItem: (item: CartItem) => void;
  clearCart: () => void;
}

const useCartStore = create<CartStore>((set) => ({
  items: [],
  addItem: (item) => set((state) => ({ items: [...state.items, item] })),
  clearCart: () => set({ items: [] }),
}));

// In jeder Komponente direkt nutzbar (kein Provider nötig)
function CartButton() {
  const { items, addItem } = useCartStore();
  return <button onClick={() => addItem(newItem)}>{items.length} items</button>;
}
```

**Jotai** — atomarer State nach Recoil-Muster:

```tsx
import { atom, useAtom } from "jotai";

const countAtom = atom(0);
const doubleCountAtom = atom((get) => get(countAtom) * 2);

function Counter() {
  const [count, setCount] = useAtom(countAtom);
  const [double] = useAtom(doubleCountAtom);
  return (
    <div>
      {count} × 2 = {double}
    </div>
  );
}
```

### 4.4 Form State: React Hook Form vs. Formik vs. native

**React Hook Form** — performant, wenig Re-Renders, uncontrolled by default:

```tsx
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";

function LoginForm() {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm({
    resolver: zodResolver(LoginSchema),
  });

  const onSubmit = (data: LoginInput) => authBackend.login(data);

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <input {...register("email")} />
      {errors.email && <span>{errors.email.message}</span>}
      <button type="submit">Login</button>
    </form>
  );
}
```

**Formik** — ältere, weit verbreitete Alternative, mehr Boilerplate.

**Native (useState + Zod)** — ausreichend für einfache Formulare:

```tsx
const [value, setValue] = useState("");
const result = Schema.safeParse({ value });
```

### 4.5 URL State und Auth State

**URL State** (navigierbar, shareable):

```tsx
// React Router v6
const [searchParams, setSearchParams] = useSearchParams();
const filter = searchParams.get("filter") ?? "all";

// nuqs (typsichere URL-Parameter)
const [page, setPage] = useQueryState("page", parseAsInteger.withDefault(1));
```

**Auth State** als Singleton (kein Re-Render nötig):

```tsx
// lib/auth.ts
class Auth {
    getToken(): string | null { ... }
    getRole(): string | null { ... }
    getUserID(): number | null { ... }
    isAuthenticated(): boolean { ... }
    logout(): void { ... }
}
```

### 4.6 Entscheidungsmatrix: Welches Tool für welchen State?

| State-Art        | Kleines Projekt     | Mittleres Projekt     | Großes Projekt        |
| ---------------- | ------------------- | --------------------- | --------------------- |
| **Server State** | useFetch (custom)   | TanStack Query        | TanStack Query + RTK  |
| **Client State** | useState            | useState / Zustand    | Zustand / Redux Tk.   |
| **Auth State**   | Singleton / Context | Singleton / Context   | Context + Provider    |
| **Form State**   | useState + Zod      | React Hook Form + Zod | React Hook Form + Zod |
| **URL State**    | React Router params | React Router / nuqs   | nuqs                  |

### 4.7 Wann Context, wann nicht?

| Szenario                   | Lösung                  | Begründung                             |
| -------------------------- | ----------------------- | -------------------------------------- |
| Globales Theme (Dark Mode) | Context/Provider        | Muss den gesamten Baum durchdringen    |
| Auth-Token                 | Singleton               | Kein Re-Render bei Token-Refresh nötig |
| Formular-State             | `useState` im Formular  | Lokal, nicht geteilt                   |
| Server-Daten               | TanStack Query/useFetch | Bei Navigation frisch laden            |
| Drawer-Zustand             | `useState` in Parent    | Lokal, 1-2 Ebenen tief                 |
| Globaler App-State         | Zustand / Jotai         | Kein Provider, kein Context-Hell       |

> **Wichtig:** React Context ist kein State-Management-Tool — es ist ein Dependency-Injection-Tool. Context-Wertänderungen lösen Re-Renders aller Consumer aus. Für häufig wechselnden State lieber Zustand oder Jotai nutzen.

---

## 5. React Design Patterns Katalog

React bietet durch Komponenten-Komposition und Hooks ideale Voraussetzungen für viele Design Patterns. Die folgenden 15+ Patterns sind in der Praxis am relevantesten.

### Kern-Patterns

#### 5.1 Component Composition Pattern

Das fundamentale React-Pattern. Anwendungen bestehen aus einem Baum von Komponenten, nicht aus einem monolithischen Block.

```tsx
// RICHTIG: Komposition
<Drawer>
  <DrawerContent>
    <ItemList items={items} />
    <TotalPriceDisplay cents={total} />
    <ConfirmButton onClick={onConfirm} />
  </DrawerContent>
</Drawer>

// FALSCH: God Component
<OrderManager /* macht alles: laden, anzeigen, bestellen, bezahlen */ />
```

**Props Down, Events Up:** Daten fließen von oben nach unten (Props), Aktionen von unten nach oben (Callbacks).

**Wann nutzen:** Immer. Dies ist das Fundament jeder React-Anwendung.

#### 5.2 Custom Hook Pattern

Logik wird in wiederverwendbare Hooks extrahiert, um Separation of Concerns zu erreichen:

```tsx
// ❌ Logik in der Komponente — schwer testbar, nicht wiederverwendbar
const PostsComponent = () => {
  const [posts, setPosts] = useState(null);
  const [loading, setLoading] = useState(true);
  useEffect(() => {
    fetch("/api/posts")
      .then((r) => r.json())
      .then(setPosts)
      .finally(() => setLoading(false));
  }, []);
  // ...
};

// ✅ Logik im Hook — testbar, wiederverwendbar
const usePosts = () => {
  const [posts, setPosts] = useState(null);
  const [loading, setLoading] = useState(true);
  useEffect(() => {
    fetch("/api/posts")
      .then((r) => r.json())
      .then(setPosts)
      .finally(() => setLoading(false));
  }, []);
  return { posts, loading };
};

const PostsComponent = () => {
  const { posts, loading } = usePosts();
  // Komponente ist jetzt sauber und lesbar
};
```

**Wann nutzen:** Immer wenn mehrere Komponenten die gleiche Logik brauchen, oder wenn eine Komponente zu viel Logik enthält.

#### 5.3 Control Props Pattern (Controlled vs. Uncontrolled)

Entscheidung, ob eine Komponente ihren eigenen State verwaltet (uncontrolled) oder den State vom Parent bekommt (controlled):

```tsx
// Uncontrolled — verwaltet eigenen State
const UncontrolledInput = () => {
  const [value, setValue] = useState("");
  return <input value={value} onChange={(e) => setValue(e.target.value)} />;
};

// Controlled — State kommt vom Parent
const ControlledInput = ({
  value,
  onChange,
}: {
  value: string;
  onChange: (value: string) => void;
}) => {
  return <input value={value} onChange={(e) => onChange(e.target.value)} />;
};

// Nutzung: Parent kontrolliert den State
const [name, setName] = useState("");
<ControlledInput value={name} onChange={setName} />;
```

Controlled Components sind besser für Wiederverwendbarkeit und das Open/Closed-Prinzip: Sie müssen nicht modifiziert werden, um von außen gesteuert zu werden.

**Wann nutzen:** Wenn der Parent den State kennen oder verändern muss (z.B. Formulare, Dropdown-Menüs die programmatisch geschlossen werden sollen).

#### 5.4 Provider Pattern

React Context als Dependency-Injection-System für Konfigurationen, die den gesamten Komponentenbaum durchdringen:

```tsx
const ThemeContext = createContext<{
  theme: string;
  setTheme: (t: string) => void;
} | null>(null);

const ThemeProvider = ({ children }: { children: ReactNode }) => {
  const [theme, setTheme] = useState("light");
  return (
    <ThemeContext.Provider value={{ theme, setTheme }}>
      {children}
    </ThemeContext.Provider>
  );
};

// Custom Hook für sicheren Context-Zugriff
const useTheme = () => {
  const context = useContext(ThemeContext);
  if (!context) throw new Error("useTheme must be used within ThemeProvider");
  return context;
};

const App = () => (
  <ThemeProvider>
    <AppContent />
  </ThemeProvider>
);
```

**Wichtig:** Context ist kein State-Management-Tool. Es eignet sich für selten ändernde Konfigurationswerte (Theme, i18n, Auth-Config). Für häufig ändernden State führt es zu unnötigen Re-Renders.

**Wann nutzen:** Theme, i18n, Feature-Flags, Auth-Konfiguration — Dinge die sich selten ändern und tief im Baum benötigt werden.

### Häufige Patterns

#### 5.5 Container/Presentational Pattern

Trennung von Datenlogik (Container) und reiner UI (Presentational):

```tsx
// Container: Holt Daten, verwaltet State, enthält Logik
function OrderDetailContainer({ orderId }: { orderId: number }) {
  const {
    data: order,
    loading,
    error,
  } = useFetch(() => backend.getOrder(orderId));
  const navigate = useNavigate();

  if (loading) return <LoadingSpinner />;
  if (error) return <ErrorDisplay error={error} />;

  return <OrderDetailView order={order!} onBack={() => navigate(-1)} />;
}

// Presentational: Reines UI, keine Side Effects, gut testbar
function OrderDetailView({
  order,
  onBack,
}: {
  order: Order;
  onBack: () => void;
}) {
  return (
    <Card>
      <CardHeader>
        <Button variant="ghost" onClick={onBack}>
          ← Zurück
        </Button>
        <CardTitle>{order.name}</CardTitle>
      </CardHeader>
      <CardContent>Saldo: {formatCents(order.totalCents)}</CardContent>
    </Card>
  );
}
```

**Wann nutzen:** Immer als Leitprinzip, auch wenn die Trennung nicht in separate Dateien geht. Presentational Components sind einfach zu testen und zu verwenden.

#### 5.6 Compound Components Pattern

Mehrere Komponenten, die zusammenarbeiten als wären sie eine einzige Einheit. Vermeidet Prop-Drilling und ermöglicht flexible Komposition:

```tsx
// Ohne Compound Components: Prop-Drilling
<Modal
  isOpen={isOpen}
  title="Löschen"
  body="Wirklich löschen?"
  onConfirm={handleDelete}
  onCancel={handleCancel}
  confirmText="Löschen"
  cancelText="Abbrechen"
/>

// Mit Compound Components: Flexibel, lesbar
<Modal isOpen={isOpen}>
  <Modal.Header>Löschen</Modal.Header>
  <Modal.Body>Wirklich löschen?</Modal.Body>
  <Modal.Footer>
    <Button onClick={handleCancel}>Abbrechen</Button>
    <Button variant="destructive" onClick={handleDelete}>Löschen</Button>
  </Modal.Footer>
</Modal>
```

shadcn/ui nutzt dieses Pattern extensiv (Card, Dialog, Select, etc.).

**Wann nutzen:** Komplexe UI-Komponenten, die intern Daten teilen müssen, aber nach außen flexibel konfigurierbar sein sollen.

#### 5.7 Headless Components Pattern

Komponentenlogik wird ohne Styling bereitgestellt — der Consumer ist für das visuelle Design zuständig:

```tsx
// Headless Hook: Nur Logik, kein Styling
const useDropdown = (items: string[]) => {
  const [isOpen, setIsOpen] = useState(false);
  const [selectedIndex, setSelectedIndex] = useState(-1);

  const toggle = () => setIsOpen(!isOpen);
  const select = (index: number) => {
    setSelectedIndex(index);
    setIsOpen(false);
  };
  const getItemProps = (index: number) => ({
    onClick: () => select(index),
    "aria-selected": selectedIndex === index,
  });

  return {
    isOpen,
    selectedIndex,
    toggle,
    getItemProps,
    selected: items[selectedIndex],
  };
};

// Consumer: Vollständige Kontrolle über Styling
function StyledDropdown({ items }: { items: string[] }) {
  const { isOpen, selectedIndex, toggle, getItemProps, selected } =
    useDropdown(items);

  return (
    <div className="relative">
      <button onClick={toggle} className="border rounded px-3 py-2">
        {selected ?? "Auswählen"} ▾
      </button>
      {isOpen && (
        <ul className="absolute border rounded shadow-lg bg-white">
          {items.map((item, i) => (
            <li
              key={i}
              {...getItemProps(i)}
              className={cn(
                "px-3 py-2 cursor-pointer hover:bg-gray-100",
                i === selectedIndex && "bg-blue-50 font-medium",
              )}
            >
              {item}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
```

Radix UI und Ark UI sind populäre Headless-Component-Libraries.

**Wann nutzen:** Wenn vollständige Styling-Kontrolle nötig ist, aber komplexe Logik (Keyboard-Navigation, ARIA, Accessibility) nicht selbst implementiert werden soll.

#### 5.8 Render Props Pattern

Eine Komponente erhält eine Funktion als Prop, die das Rendern übernimmt. Ermöglicht flexible Wiederverwendung von Logik:

```tsx
// Render Props Komponente
interface DataLoaderProps<T> {
  fetcher: () => Promise<T>;
  children: (data: T, reload: () => void) => ReactNode;
  fallback?: ReactNode;
}

function DataLoader<T>({ fetcher, children, fallback }: DataLoaderProps<T>) {
  const { data, loading, error, reload } = useFetch(fetcher);

  if (loading) return <LoadingSpinner />;
  if (error) return <ErrorDisplay error={error} />;
  if (!data) return fallback ?? null;

  return <>{children(data, reload)}</>;
}

// Nutzung
<DataLoader fetcher={() => backend.getOrders()}>
  {(orders, reload) => (
    <div>
      <OrderGrid orders={orders} />
      <Button onClick={reload}>Aktualisieren</Button>
    </div>
  )}
</DataLoader>;
```

**Hinweis:** Mit Hooks sind Render Props oft ersetzbar durch Custom Hooks. Render Props sind aber weiterhin nützlich für Logik, die direkt mit JSX verbunden ist.

**Wann nutzen:** Wenn eine Komponente Logik mit dem Consumer teilen muss und dabei das Rendering an den Consumer delegiert.

#### 5.9 Props Getters Pattern

Statt einzelner Props gibt eine Komponente/Hook eine `get*Props`-Funktion zurück, die alle nötigen Props bündelt und mergt:

```tsx
// Props Getter Hook
const useToggle = (initialValue = false) => {
  const [isOn, setIsOn] = useState(initialValue);

  const getToggleProps = (props = {}) => ({
    "aria-pressed": isOn,
    onClick: () => setIsOn(!isOn),
    ...props, // Consumer kann Props überschreiben/erweitern
  });

  return { isOn, getToggleProps };
};

// Consumer hat volle Kontrolle, bekommt aber alle nötigen Props
function ToggleButton() {
  const { isOn, getToggleProps } = useToggle();
  return (
    <button
      {...getToggleProps({
        className: cn("rounded", isOn ? "bg-blue-500" : "bg-gray-200"),
      })}
    >
      {isOn ? "An" : "Aus"}
    </button>
  );
}
```

**Wann nutzen:** Wenn ein Hook dem Consumer viele Props für ein Element geben muss, aber flexibel erweiterbar sein soll. Häufig in Headless-Libraries (Downshift, React Aria).

#### 5.10 Error Boundary Pattern

Fängt JavaScript-Fehler im Komponentenbaum ab und verhindert, dass die gesamte App abstürzt:

```tsx
class ErrorBoundary extends React.Component<
  { children: ReactNode; fallback?: ReactNode },
  { hasError: boolean; error?: Error }
> {
  state = { hasError: false };

  static getDerivedStateFromError(error: Error) {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error("Fehler abgefangen:", error, info);
    // Optional: Error-Reporting (Sentry etc.)
  }

  render() {
    if (this.state.hasError) {
      return (
        this.props.fallback ?? (
          <div role="alert">
            <h2>Etwas ist schiefgelaufen.</h2>
            <button onClick={() => this.setState({ hasError: false })}>
              Erneut versuchen
            </button>
          </div>
        )
      );
    }
    return this.props.children;
  }
}

// Strategische Platzierung — mehrere unabhängige Boundaries
const App = () => (
  <ErrorBoundary fallback={<AppCrashPage />}>
    <ErrorBoundary fallback={<DashboardFallback />}>
      <Dashboard />
    </ErrorBoundary>
    <ErrorBoundary fallback={<SidebarFallback />}>
      <Sidebar />
    </ErrorBoundary>
  </ErrorBoundary>
);
```

**Wichtig:** Error Boundaries fangen **nicht**: Event-Handler-Fehler, async Code, Server-Side Rendering, eigene Fehler. Für diese Fälle try-catch verwenden.

**Wann nutzen:** Strategisch um große, unabhängige Teile der App, damit lokale Fehler nicht die gesamte Anwendung zum Absturz bringen.

#### 5.11 Portal Pattern

Rendert Kinder in einem DOM-Knoten außerhalb der Komponenten-Hierarchie — löst z-Index- und overflow-Probleme:

```tsx
import { createPortal } from "react-dom";

const Modal = ({
  isOpen,
  onClose,
  children,
}: {
  isOpen: boolean;
  onClose: () => void;
  children: ReactNode;
}) => {
  if (!isOpen) return null;

  return createPortal(
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="fixed inset-0 bg-black/50" onClick={onClose} />
      <div className="relative bg-white rounded-lg p-6 z-10">
        <button onClick={onClose} className="absolute top-4 right-4">
          ✕
        </button>
        {children}
      </div>
    </div>,
    document.body, // Wird direkt in <body> gerendert
  );
};
```

**Wann nutzen:** Modals, Tooltips, Dropdowns, Notifications — alles was über anderen Elementen erscheinen muss, ohne von `overflow: hidden` oder z-Index des Parents beeinflusst zu werden.

### Allgemeine Software-Patterns in React

#### 5.12 Higher Order Components (HOC) — Legacy Pattern

HOCs sind Funktionen, die eine Komponente nehmen und eine erweiterte Komponente zurückgeben:

```tsx
// HOC: Fügt Auth-Check hinzu
function withAuth<P extends object>(Component: ComponentType<P>) {
  return function AuthenticatedComponent(props: P) {
    const isAuthenticated = Auth.isAuthenticated();
    if (!isAuthenticated) return <Navigate to="/login" />;
    return <Component {...props} />;
  };
}

const ProtectedDashboard = withAuth(Dashboard);
```

**Legacy-Pattern:** HOCs haben in modernem React meist bessere Alternativen:

- Auth-Guard → React Router Loader
- Data Fetching → Custom Hook
- Feature-Flag → Custom Hook oder Context

**Wann (noch) nutzen:** Wenn eine Code-Basis noch Class Components nutzt oder wenn eine Library HOCs erwartet.

#### 5.13 MVVM in React

Model-View-ViewModel: Hooks übernehmen die ViewModel-Rolle:

```tsx
// Model: Datenstruktur
interface Order {
  id: number;
  items: OrderItem[];
  totalCents: number;
}

// ViewModel (Custom Hook): Logik und State
function useOrderViewModel(orderId: number) {
  const { data: order, loading } = useFetch(() => backend.getOrder(orderId));
  const [selectedItems, setSelectedItems] = useState<number[]>([]);

  const toggleItem = (id: number) => {
    setSelectedItems((prev) =>
      prev.includes(id) ? prev.filter((i) => i !== id) : [...prev, id],
    );
  };

  const selectedTotal =
    order?.items
      .filter((i) => selectedItems.includes(i.id))
      .reduce((sum, i) => sum + i.priceCents, 0) ?? 0;

  return { order, loading, selectedItems, toggleItem, selectedTotal };
}

// View: Reines Rendering
function OrderView({ orderId }: { orderId: number }) {
  const { order, loading, selectedItems, toggleItem, selectedTotal } =
    useOrderViewModel(orderId);
  if (loading) return <LoadingSpinner />;
  return (
    <div>
      {order!.items.map((item) => (
        <ItemRow
          key={item.id}
          item={item}
          selected={selectedItems.includes(item.id)}
          onToggle={() => toggleItem(item.id)}
        />
      ))}
      <Total cents={selectedTotal} />
    </div>
  );
}
```

#### 5.14 Dependency Injection in React

Abhängigkeiten können über Props, Context oder Module injiziert werden:

```tsx
// DI via Interface und Props (testbar)
interface OrderBackend {
  getOrder(id: number): Promise<Order>;
  submitOrder(data: OrderInput): Promise<void>;
}

function OrderPage({ backend }: { backend: OrderBackend }) {
  const { data } = useFetch(() => backend.getOrder(1));
  return <div>{/* ... */}</div>;
}

// Produktion
<OrderPage backend={new RealOrderBackend()} />

// Test — Mock-Implementierung
<OrderPage backend={mockOrderBackend} />

// DI via Context (für tiefe Komponentenbäume)
const BackendContext = createContext<OrderBackend | null>(null);
const useBackend = () => {
  const ctx = useContext(BackendContext);
  if (!ctx) throw new Error('BackendProvider missing');
  return ctx;
};
```

#### 5.15 SOLID-Prinzipien in React

| Prinzip                   | React-Anwendung                                 | Beispiel                                         |
| ------------------------- | ----------------------------------------------- | ------------------------------------------------ |
| **S**ingle Responsibility | Eine Komponente = eine Aufgabe                  | `OrderCard` zeigt nur, `useOrder` lädt nur       |
| **O**pen/Closed           | Erweiterbar ohne Änderung                       | Controlled Components, Render Props              |
| **L**iskov Substitution   | Komponenten sind austauschbar wenn Props passen | `Button` vs. `IconButton`                        |
| **I**nterface Segregation | Kleine, spezifische Props-Interfaces            | Nicht alle Props in einen Typ packen             |
| **D**ependency Inversion  | Abhängig von Abstraktionen                      | `BackendClient`-Interface statt konkreter Klasse |

```tsx
// S — Single Responsibility
// FALSCH: Lädt und zeigt Daten in einer Komponente
function OrderComponent({ orderId }) { /* lädt + rendert */ }

// RICHTIG: Getrennte Verantwortlichkeiten
function useOrderData(orderId: number) { /* lädt */ }
function OrderCard({ order }: { order: Order }) { /* zeigt */ }

// O — Open/Closed
// FALSCH: Komponente muss für neue Varianten geändert werden
function Button({ type }: { type: 'primary' | 'danger' | 'ghost' }) { ... }

// RICHTIG: Erweiterbar über Props/Slots
function Button({ variant, className, children, ...props }: ButtonProps) {
  return <button className={cn(variantStyles[variant], className)} {...props}>{children}</button>;
}
```

---

## 6. Backend-Integration

### BackendClient-Interface

Alle Backend-Aufrufe laufen über das `BackendClient`-Interface — niemals direkt `fetch()`:

```tsx
// lib/backend.ts
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

## 7. Routing und Guards

> Für Token-Handling, XSS-Prävention und Frontend-Security-Patterns siehe [Security & Authentifizierung](security.md).

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

| Rolle   | Zugriff              | Redirect nach Login  |
| ------- | -------------------- | -------------------- |
| `admin` | Admin + Hauptbereich | `/admin` oder `/app` |
| `user`  | Hauptbereich         | `/app`               |

---

## 8. Validierung mit Zod

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

## 9. UI-Patterns und Bibliotheken

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
  await backend.saveChanges(payload);
  toast.success("Änderungen gespeichert");
} catch (error) {
  toast.error("Vorgang fehlgeschlagen");
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

## 10. Styling mit Tailwind CSS

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

## 11. Testing-Strategien

React-Frontends werden auf mehreren Ebenen getestet. Die Test-Pyramide gilt auch hier: Viele Unit-Tests, weniger Integration-Tests, wenige E2E-Tests.

### 11.1 Test-Pyramide für React

```
                    ╔═══════╗
                    ║  E2E  ║  Playwright, Cypress
                    ║ wenig ║  Langsam, teuer, realistisch
                    ╠═══════╣
                ╔═══════════════╗
                ║  Integration  ║  MSW + React Testing Library
                ║   mittel      ║  API-Mocking, User-Flows
                ╠═══════════════╣
          ╔═════════════════════════╗
          ║       Unit Tests        ║  Vitest, Jest
          ║          viele          ║  Schnell, isoliert
          ╚═════════════════════════╝
```

### 11.2 Unit Tests: Vitest und Jest

**Vitest** ist der Vite-native Test-Runner — schnell, ESM-native, Jest-kompatibel:

```tsx
// lib/utils.test.ts
import { describe, it, expect } from "vitest";
import { formatCents } from "./utils";

describe("formatCents", () => {
  it("formatiert 350 Cent korrekt", () => {
    expect(formatCents(350)).toBe("3,50 €");
  });

  it("formatiert 0 korrekt", () => {
    expect(formatCents(0)).toBe("0,00 €");
  });

  it("formatiert negative Beträge", () => {
    expect(formatCents(-100)).toBe("-1,00 €");
  });
});

// Custom Hook testen
import { renderHook, act } from "@testing-library/react";
import { useCounter } from "./useCounter";

describe("useCounter", () => {
  it("startet bei 0", () => {
    const { result } = renderHook(() => useCounter());
    expect(result.current.count).toBe(0);
  });

  it("erhöht den Wert", () => {
    const { result } = renderHook(() => useCounter());
    act(() => result.current.increment());
    expect(result.current.count).toBe(1);
  });
});
```

### 11.3 Component Tests: React Testing Library

React Testing Library (RTL) testet Komponenten aus Nutzer-Perspektive:

```tsx
// src/components/ProductCard.test.tsx
import { render, screen, fireEvent } from "@testing-library/react";
import { ProductCard } from "./ProductCard";

const mockProduct = { id: 1, name: "Espresso", priceCents: 350 };

describe("ProductCard", () => {
  it("zeigt den Produkt-Namen an", () => {
    render(<ProductCard product={mockProduct} onSelect={() => {}} />);
    expect(screen.getByText("Espresso")).toBeInTheDocument();
  });

  it("zeigt den formatierten Preis an", () => {
    render(<ProductCard product={mockProduct} onSelect={() => {}} />);
    expect(screen.getByText("3,50 €")).toBeInTheDocument();
  });

  it("ruft onSelect beim Klick auf", async () => {
    const handleSelect = vi.fn();
    render(<ProductCard product={mockProduct} onSelect={handleSelect} />);

    await userEvent.click(screen.getByRole("button", { name: /espresso/i }));
    expect(handleSelect).toHaveBeenCalledWith(mockProduct);
  });
});
```

**Prinzipien:**

- **Kein Test von Implementierungsdetails** — nicht `wrapper.find('.order-card')`, sondern `screen.getByRole('button')`
- **Queries in Priorität:** `getByRole` > `getByLabelText` > `getByText` > `getByTestId`
- **User Interactions:** `userEvent` statt `fireEvent` (simuliert echtes User-Verhalten)

### 11.4 Integration Tests: MSW (Mock Service Worker)

MSW interceptiert echte Netzwerk-Anfragen im Test-Environment:

```tsx
// src/mocks/handlers.ts
import { http, HttpResponse } from "msw";

export const handlers = [
  http.post("/api/get-product", ({ request }) => {
    return HttpResponse.json({ id: 1, name: "Espresso", priceCents: 350 });
  }),

  http.post("/api/create-order", async ({ request }) => {
    const body = await request.json();
    if (!body.items?.length) {
      return HttpResponse.json({ code: "EMPTY_CART" }, { status: 400 });
    }
    return HttpResponse.json({ success: true });
  }),
];

// src/mocks/server.ts (für Node/Jest/Vitest)
import { setupServer } from "msw/node";
export const server = setupServer(...handlers);

// Integrations-Test
import { render, screen, waitFor } from "@testing-library/react";
import { server } from "../mocks/server";

describe("ProductPage Integration", () => {
  it("lädt und zeigt Produkt an", async () => {
    render(<ProductPage productId={1} />);

    await waitFor(() => {
      expect(screen.getByText("Espresso")).toBeInTheDocument();
      expect(screen.getByText("3,50 €")).toBeInTheDocument();
    });
  });

  it("zeigt Fehler bei leerem Warenkorb", async () => {
    server.use(
      http.post("/api/create-order", () =>
        HttpResponse.json({ code: "EMPTY_CART" }, { status: 400 }),
      ),
    );

    render(<ProductPage productId={1} />);
    await userEvent.click(screen.getByText("Hinzufügen"));
    await waitFor(() => {
      expect(screen.getByText(/leerer warenkorb/i)).toBeInTheDocument();
    });
  });
});
```

### 11.5 E2E Tests: Playwright

Playwright testet die echte Anwendung im Browser:

```typescript
// e2e/checkout-flow.spec.ts
import { test, expect } from "@playwright/test";

test("Checkout-Workflow", async ({ page }) => {
  // Login
  await page.goto("/login");
  await page.fill('[name="username"]', "cashier");
  await page.fill('[name="password"]', "password");
  await page.click('button[type="submit"]');

  // Kategorie auswählen
  await expect(page).toHaveURL("/dashboard");
  await page.click("text=Getränke");

  // Produkt hinzufügen
  await page.click("text=Zur Kasse");
  await page.click('[data-testid="product-espresso"]');
  await page.click("text=Bestätigen");

  // Bestellung bestätigt
  await expect(page.locator("text=Änderungen gespeichert")).toBeVisible();
});
```

---

## 12. Performance-Patterns

### 12.1 React.memo — Rendering-Optimierung

Verhindert unnötige Re-Renders von funktionalen Komponenten:

```tsx
// Ohne memo: Re-rendert bei jedem Parent-Re-Render
function ExpensiveList({ items }: { items: Item[] }) {
  return (
    <ul>
      {items.map((item) => (
        <li key={item.id}>{item.name}</li>
      ))}
    </ul>
  );
}

// Mit memo: Re-rendert nur wenn items sich ändert
const ExpensiveList = React.memo(({ items }: { items: Item[] }) => {
  return (
    <ul>
      {items.map((item) => (
        <li key={item.id}>{item.name}</li>
      ))}
    </ul>
  );
});
```

**Wann memo verwenden:**

- ✅ Teure Render-Operationen (große Listen, komplexe Berechnungen)
- ✅ Komponenten die oft re-rendern obwohl Props sich nicht ändern
- ❌ Nicht für jede Komponente — Overhead durch Vergleich
- ❌ Nicht wenn Props sich bei fast jedem Re-Render ändern

### 12.2 useMemo und useCallback

```tsx
function OrderSummary({
  items,
  discount,
}: {
  items: OrderItem[];
  discount: number;
}) {
  // useMemo: Teure Berechnung nur bei Änderung der Dependencies
  const subtotal = useMemo(
    () => items.reduce((sum, item) => sum + item.priceCents * item.quantity, 0),
    [items],
  );

  const total = useMemo(
    () => Math.round(subtotal * (1 - discount / 100)),
    [subtotal, discount],
  );

  // useCallback: Stabile Funktionsreferenz für memo-Kinder
  const handleRemove = useCallback((itemId: number) => {
    setItems((prev) => prev.filter((i) => i.id !== itemId));
  }, []); // Keine Dependencies → immer stabile Referenz

  return (
    <div>
      <ItemList items={items} onRemove={handleRemove} />
      <Total subtotal={subtotal} total={total} />
    </div>
  );
}
```

**Faustregel:** `useMemo`/`useCallback` nur bei nachgewiesenem Performance-Problem einsetzen — sie fügen Komplexität hinzu und sind kein kostenloses Upgrade.

### 12.3 React.lazy und Suspense — Code Splitting

```tsx
// Lazy Loading: Komponente wird erst beim ersten Rendern geladen
const AdminPage = React.lazy(() => import("./admin/AdminPage"));
const ServicePage = React.lazy(() => import("./service/ServicePage"));

function App() {
  return (
    <Suspense fallback={<LoadingSpinner />}>
      <Routes>
        <Route path="/admin/*" element={<AdminPage />} />
        <Route path="/service/*" element={<ServicePage />} />
      </Routes>
    </Suspense>
  );
}
```

**Vorteil:** Bundle wird aufgeteilt — Nutzer laden nur den Code, den sie brauchen.

### 12.4 Virtualization — Große Listen

Für sehr lange Listen (100+ Einträge) werden nur die sichtbaren Elemente gerendert:

```tsx
import { useVirtualizer } from "@tanstack/react-virtual";

function VirtualList({ items }: { items: Product[] }) {
  const parentRef = useRef<HTMLDivElement>(null);

  const virtualizer = useVirtualizer({
    count: items.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 60, // Geschätzte Höhe eines Items
  });

  return (
    <div ref={parentRef} style={{ height: "400px", overflow: "auto" }}>
      <div
        style={{
          height: `${virtualizer.getTotalSize()}px`,
          position: "relative",
        }}
      >
        {virtualizer.getVirtualItems().map((virtualItem) => (
          <div
            key={virtualItem.key}
            style={{ position: "absolute", top: virtualItem.start }}
          >
            <ProductCard product={items[virtualItem.index]} />
          </div>
        ))}
      </div>
    </div>
  );
}
```

### 12.5 Bundle-Analyse

```bash
# Vite Bundle-Analyse
npx vite-bundle-visualizer

# Tree Shaking prüfen: Werden alle Importe genutzt?
import { onlyThisFunction } from 'large-library'; // ✅ Tree-shaking
import * as everything from 'large-library';       // ❌ Kein Tree-shaking
```

---

## 13. Accessibility (a11y)

Accessibility macht Anwendungen für alle Nutzer nutzbar — auch mit Tastatur, Screen Reader oder motorischen Einschränkungen.

### 13.1 ARIA-Rollen und Labels

ARIA (Accessible Rich Internet Applications) ergänzt HTML-Semantik für komplexe Widgets:

```tsx
// Semantisches HTML bevorzugen — ARIA ist ein Fallback
// ✅ Native Semantik
<button onClick={handleClick}>Bestellen</button>
<input type="text" id="name" />
<label htmlFor="name">Name</label>

// ARIA für Custom Components
<div
  role="button"
  tabIndex={0}
  aria-label="Eintrag löschen"
  onClick={handleDelete}
  onKeyDown={e => e.key === 'Enter' && handleDelete()}
>
  <TrashIcon />
</div>

// Dynamische Inhalte ankündigen
<div aria-live="polite" aria-atomic="true">
  {statusMessage}
</div>

// Ladezustand
<div role="status" aria-label="Wird geladen...">
  <LoadingSpinner />
</div>
```

### 13.2 Keyboard Navigation

Alle interaktiven Elemente müssen per Tastatur erreichbar sein:

```tsx
// Tab-Reihenfolge — logische DOM-Reihenfolge
// Kein positives tabIndex verwenden (tabIndex={1}, tabIndex={2} etc.)

// Focus Trap für Modals
function Modal({ isOpen, onClose, children }: ModalProps) {
  const firstFocusableRef = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    if (isOpen) {
      firstFocusableRef.current?.focus();
    }
  }, [isOpen]);

  const handleKeyDown = (e: KeyboardEvent) => {
    if (e.key === "Escape") onClose();
  };

  return isOpen ? (
    <div role="dialog" aria-modal="true" onKeyDown={handleKeyDown}>
      <button ref={firstFocusableRef} onClick={onClose} aria-label="Schließen">
        ✕
      </button>
      {children}
    </div>
  ) : null;
}
```

### 13.3 Focus Management

```tsx
// Focus nach Aktion wiederherstellen
function DeleteButton({ item, onDelete }: Props) {
  const buttonRef = useRef<HTMLButtonElement>(null);

  const handleDelete = async () => {
    await onDelete(item.id);
    // Nach dem Löschen: Focus auf sinnvolles Element
    document.querySelector<HTMLElement>("[data-first-item]")?.focus();
  };

  return (
    <button ref={buttonRef} onClick={handleDelete}>
      Löschen
    </button>
  );
}

// Focus sichtbar machen (nicht outline: none!)
// In Tailwind: focus-visible:ring-2 focus-visible:ring-primary
<button className="rounded px-4 py-2 focus-visible:ring-2 focus-visible:ring-primary focus-visible:outline-none">
  Bestellen
</button>;
```

### 13.4 Farbe und Kontrast

```tsx
// Mindest-Kontrast WCAG AA: 4.5:1 (Text), 3:1 (UI-Elemente)
// Tailwind-Design-Tokens sind i.d.R. WCAG-konform

// ❌ Farbe als einzige Information
<span className="text-red-500">Fehler aufgetreten</span>

// ✅ Farbe + Icon/Text
<span className="text-red-500" role="alert">
  <AlertCircle className="inline mr-1" aria-hidden="true" />
  Fehler aufgetreten
</span>
```

### 13.5 Screen Reader Testing

```bash
# macOS: VoiceOver (Cmd + F5)
# Windows: NVDA (kostenlos) oder Narrator
# Automatische Prüfung:
npm install --save-dev @axe-core/react
```

```tsx
// Automatische Accessibility-Checks in Tests
import { axe } from "jest-axe";

it("hat keine Accessibility-Fehler", async () => {
  const { container } = render(<OrderForm />);
  const results = await axe(container);
  expect(results).toHaveNoViolations();
});
```

### 13.6 Checkliste für neue Komponenten

- [ ] Alle Bilder haben `alt`-Text (oder `alt=""` bei dekorativen)
- [ ] Alle Formulare haben verknüpfte Labels (`htmlFor` / `aria-label`)
- [ ] Interaktive Elemente sind per Tastatur erreichbar
- [ ] Focus-Indikator ist sichtbar
- [ ] Fehlermeldungen sind mit `role="alert"` oder `aria-live` gekennzeichnet
- [ ] Farbkontrast erfüllt WCAG AA (4.5:1)
- [ ] Modals haben `role="dialog"` und `aria-modal="true"`

---

## 14. TypeScript-Patterns für React

TypeScript erhöht die Sicherheit und Wartbarkeit von React-Code erheblich.

### 14.1 Discriminated Unions für Props

```tsx
// Verschiedene Varianten einer Komponente typsicher modellieren
type ButtonProps =
  | { variant: "primary"; onClick: () => void; children: ReactNode }
  | { variant: "link"; href: string; children: ReactNode }
  | { variant: "icon"; icon: ReactNode; label: string; onClick: () => void };

function Button(props: ButtonProps) {
  switch (props.variant) {
    case "primary":
      return <button onClick={props.onClick}>{props.children}</button>;
    case "link":
      return <a href={props.href}>{props.children}</a>;
    case "icon":
      return (
        <button onClick={props.onClick} aria-label={props.label}>
          {props.icon}
        </button>
      );
  }
}

// State mit Discriminated Union
type FetchState<T> =
  | { status: "loading" }
  | { status: "error"; error: Error }
  | { status: "success"; data: T };

function useOrderState(id: number): FetchState<Order> {
  // TypeScript weiß: wenn status === 'success', gibt es data
}
```

### 14.2 Generic Components

```tsx
// Generische Liste — typsicher für beliebige Item-Typen
interface DataListProps<T> {
  items: T[];
  getKey: (item: T) => string | number;
  renderItem: (item: T) => ReactNode;
  emptyMessage?: string;
}

function DataList<T>({
  items,
  getKey,
  renderItem,
  emptyMessage,
}: DataListProps<T>) {
  if (items.length === 0) {
    return <p>{emptyMessage ?? "Keine Einträge"}</p>;
  }
  return (
    <ul>
      {items.map((item) => (
        <li key={getKey(item)}>{renderItem(item)}</li>
      ))}
    </ul>
  );
}

// Nutzung — vollständig typsicher
<DataList
  items={orders}
  getKey={(order) => order.id}
  renderItem={(order) => <OrderCard order={order} />}
/>;
```

### 14.3 Branded Types für IDs

```tsx
// Ohne Branded Types: IDs sind austauschbar — Fehler werden nicht erkannt
type OrderId = number;
type ProductId = number;
function getOrder(id: OrderId) { ... }
getOrder(productId); // ✅ TypeScript beschwert sich NICHT — Fehler!

// Mit Branded Types: IDs sind distinguishable
type OrderId = number & { readonly __brand: 'OrderId' };
type ProductId = number & { readonly __brand: 'ProductId' };

function createOrderId(id: number): OrderId {
  return id as OrderId;
}

function getOrder(id: OrderId) { ... }

const productId = createProductId(42);
getOrder(productId); // ❌ TypeScript-Fehler — korrekt!
```

### 14.4 Props-Extraktion und Wiederverwendung

```tsx
// ComponentProps: Props einer Komponente extrahieren und erweitern
import { ComponentProps } from "react";

type CustomButtonProps = ComponentProps<"button"> & {
  isLoading?: boolean;
};

function LoadingButton({
  isLoading,
  children,
  disabled,
  ...props
}: CustomButtonProps) {
  return (
    <button disabled={disabled || isLoading} {...props}>
      {isLoading ? <Spinner /> : children}
    </button>
  );
}

// Ref Forwarding mit korrekten Typen
const Input = forwardRef<HTMLInputElement, ComponentProps<"input">>(
  (props, ref) => <input ref={ref} {...props} />,
);
```

### 14.5 Template Literal Types für Event Handler

```tsx
// Typsichere Event-Handler-Namen
type EventName = "click" | "focus" | "blur" | "change";
type HandlerName = `on${Capitalize<EventName>}`; // 'onClick' | 'onFocus' | 'onBlur' | 'onChange'

// Typsichere Zustandsübergänge
type OrderStatus = "pending" | "confirmed" | "delivered" | "cancelled";
type StatusTransition = `${OrderStatus}To${Capitalize<OrderStatus>}`;
// 'pendingToConfirmed' | 'confirmedToDelivered' | ...
```

---

## 15. Fehlerbehandlung im Frontend

### Fehler-Strategien

| Fehlerart              | Strategie                 | UI-Feedback                            |
| ---------------------- | ------------------------- | -------------------------------------- |
| **401 Unauthorized**   | Automatisch (Interceptor) | Redirect zu `/login`                   |
| **Validierungsfehler** | Inline-Anzeige            | Feld-basierte Fehlermeldungen          |
| **Netzwerkfehler**     | Toast + Retry             | `toast.error('Netzwerkfehler')`        |
| **Backend-Fehler**     | Toast                     | `toast.error('Aktion fehlgeschlagen')` |
| **Unerwartete Fehler** | Error Boundary            | Fallback-UI                            |

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

Siehe auch [Section 5.10 — Error Boundary Pattern](#510-error-boundary-pattern) für das vollständige Pattern mit Wiederherstellungs-Button und componentDidCatch.

---

## 16. Anti-Patterns

### 16.1 Direktes fetch()

```tsx
// FALSCH: Direkt fetch() aufrufen
const response = await fetch('/api/get-order', { ... });

// RICHTIG: Über Backend-Klasse
const order = await orderBackend.getOrder(orderId);
```

### 16.2 Frontend-Filterung

```tsx
// FALSCH: Im Frontend filtern
const unpaid = allItems.filter((p) => !p.paid);

// RICHTIG: Backend liefert bereits gefiltert
const unpaid = await backend.getUnpaidItems(orderId);
```

### 16.3 Prop Drilling über viele Ebenen

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

### 16.4 Inline Geldbeträge formatieren

```tsx
// FALSCH
<span>{(cents / 100).toFixed(2)} €</span>
<span>{cents / 100} €</span>

// RICHTIG
<span>{formatCents(cents)}</span>
```

### 16.5 Globaler State-Store (unnötig)

```tsx
// FALSCH: Redux/Zustand für wenige Seiten
const store = createStore({ orders: [], products: [], user: null, ... });

// RICHTIG: Lokale Hooks + useFetch + Singletons
const { data: orders } = useFetch(() => backend.getOrders());
```

---

## 17. Referenzen

### React-Architektur

- [21 Fantastic React Design Patterns](https://www.perssondennis.com/articles/21-fantastic-react-design-patterns-and-when-to-use-them) — 21 Patterns mit Code-Beispielen (Dennis Persson)
- [Guide to Modern Frontend Architecture Patterns](https://blog.logrocket.com/guide-modern-frontend-architecture-patterns/) — Monolithic, Modular, Micro-Frontend, Flux (LogRocket)
- [Mastering Atomic Design in React](https://javascript.plainenglish.io/mastering-atomic-design-a-step-by-step-guide-to-building-scalable-ui-components-60b0d2a94cc3) — Atomic Design Step-by-Step
- [Brad Frost: Atomic Design](https://atomicdesign.bradfrost.com/) — Original Atomic Design Methodik
- [Kent C. Dodds: Application State Management](https://kentcdodds.com/blog/application-state-management-with-react) — State ohne Redux
- [Feature-Sliced Design](https://feature-sliced.design/) — Modular Frontend Architecture
- [Dan Abramov: Writing Resilient Components](https://overreacted.io/writing-resilient-components/) — React Best Practices

### State Management

- [TanStack Query Docs](https://tanstack.com/query/latest) — Server State Management
- [Zustand](https://zustand-demo.pmnd.rs/) — Minimaler globaler State-Store
- [Jotai](https://jotai.org/) — Atomarer State für React
- [React Hook Form](https://react-hook-form.com/) — Performante Formulare

### Testing

- [React Testing Library](https://testing-library.com/docs/react-testing-library/intro/) — Component Testing (user-centric)
- [Vitest](https://vitest.dev/) — Vite-nativer Test-Runner
- [MSW (Mock Service Worker)](https://mswjs.io/) — API Mocking
- [Playwright](https://playwright.dev/) — E2E Testing

### Performance und Accessibility

- [React Docs: Performance](https://react.dev/reference/react/memo) — Offizielle React-Performance-Dokumentation
- [WCAG 2.1 Richtlinien](https://www.w3.org/TR/WCAG21/) — Web Content Accessibility Guidelines
- [MDN: ARIA](https://developer.mozilla.org/en-US/docs/Web/Accessibility/ARIA) — ARIA Roles und Attributes
- [TanStack Virtual](https://tanstack.com/virtual/latest) — Virtualization für große Listen

### TypeScript

- [React TypeScript Cheatsheet](https://react-typescript-cheatsheet.netlify.app/) — TypeScript-Patterns für React
- [TypeScript Handbook: Template Literal Types](https://www.typescriptlang.org/docs/handbook/2/template-literal-types.html) — Typsichere String-Templates

### Libraries

- [shadcn/ui](https://ui.shadcn.com/) — UI-Komponenten (Radix + Tailwind)
- [Zod](https://zod.dev/) — TypeScript-first Schema-Validierung
- [React Router](https://reactrouter.com/) — Client-Side Routing
- [Sonner](https://github.com/emilkowalski/sonner) — Toast-Notifications
- [Vaul](https://github.com/emilkowalski/vaul) — Drawer-Komponente
- [Lucide](https://lucide.dev/) — Icon-Library
- [Tailwind CSS](https://tailwindcss.com/) — Utility-first CSS
