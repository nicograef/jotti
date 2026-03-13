---
description: "Use when working on React frontend code, components, pages, hooks, styling, or TypeScript types."
applyTo: "frontend/**"
---

> **Referenz:** Für Ubiquitous Language, Namenskonventionen und Ist/Soll-Abweichungen (Rename-Status) → `docs/design/language.md`. Für Frontend-Architektur → `docs/design/handbuch.md` §6.3.

# Frontend-Konventionen

## Befehle

- **Dev-Server:** `make dev` (ganzer Stack) oder `cd frontend && pnpm dev`
- **Build:** `make build-frontend`
- **Tests:** `make test-frontend`
- **Lint:** `make lint-frontend`
- **Format:** `make fmt-frontend`

## Verzeichnisstruktur

```
frontend/
  src/routes.ts                 # Alle Routen + Guards
  src/App.tsx                   # Root-Komponente
  src/lib/                      # Auth, Backend-Client, useFetch-Hook, Utilities
  src/admin/                    # Admin-Bereich (Produkte, Tische, Benutzer)
  src/service/                  # Service-Bereich (Tisch-Workflow)
  src/pages/                    # Login, Passwort setzen
  src/components/ui/            # shadcn/ui-Komponenten
  src/components/common/        # Gemeinsame Komponenten
```

## UI-Bibliotheken

- **shadcn/ui** (Stil: `new-york`, Radix-basiert)
- **Lucide React** (Icons)
- **Sonner** (Toasts) — alle mutativen Aktionen zeigen `toast.error(...)` bei Fehlern
- **Vaul** (Drawers)

## Patterns

- **401-Interceptor**: `Backend.post()` erkennt 401, loggt aus und leitet zu `/login` weiter — kein manuelles 401-Handling nötig
- **Drawer-Pattern**: Bestellen, Bezahlen, Stornieren, Liefern öffnen Bottom-Sheet-Drawer mit Zusammenfassung. Hilfsfunktionen (`selectPositionen`, `calculateTotalPrice`) in `src/service/components/table/drawerUtils.ts`
- **Geldbeträge anzeigen**: `formatCents()` aus `src/lib/utils.ts` — nie inline formatieren
- **API-Vertrag durch Backend-DTOs definiert**: Die JSON-Struktur der API ist durch Response-DTOs in der Backend HTTP-Schicht festgelegt, nicht durch Domain-Modelle. Frontend Zod-Schemas sollten gegen die API-Dokumentation bzw. tatsächliche Responses validiert werden.

## Styling

- Tailwind CSS 4 via `@tailwindcss/vite` (keine `tailwind.config.js`)
- CSS-Variablen in `src/index.css` (Violet/Indigo-Schema, Dark Mode via `.dark`-Klasse)
- `cn()` Utility aus `src/lib/utils.ts` (`clsx` + `tailwind-merge`)
- Path-Alias: `@/` → `./src/`

## Weiterführende Dokumentation

- **Namenskonventionen (UI-Labels deutsch, Code-Mappings, Ist vs. Soll):** [docs/design/language.md](../../docs/design/language.md)
- **Architektur, Frontend-Patterns, Validierung:** [docs/design/handbuch.md](../../docs/design/handbuch.md) Kap. 6

## Code-Beispiele

### Backend-Client-Klasse

Pattern: Zod-Schema für Request definieren → `BackendClient.post()` aufrufen → Response mit Zod validieren.

```typescript
import { z } from "zod";
import type { BackendClient } from "@/lib/Backend";
import { type Product, ProductIdSchema, ProductSchema } from "./Product";

export const CreateProductSchema = ProductSchema.pick({
  name: true,
  category: true,
});

export class ProductBackend {
  constructor(private readonly backend: BackendClient) {}

  async createProduct(
    newProduct: z.infer<typeof CreateProductSchema>,
  ): Promise<number> {
    const body = CreateProductSchema.parse(newProduct);
    const { id } = await this.backend.post(
      "admin/create-produkt",
      body,
      z.object({ id: ProductIdSchema }),
    );
    return id;
  }

  async getAllProducts(): Promise<Product[]> {
    const { produkte } = await this.backend.post(
      "admin/get-all-produkte",
      {},
      z.object({ produkte: z.array(ProductSchema) }),
    );
    return produkte;
  }
}
```

### Custom Hook mit useFetch

```typescript
import { BackendSingleton } from "@/lib/Backend";
import { useFetch } from "@/lib/useFetch";
import type { Product } from "./Product";
import { ProductBackend } from "./ProductBackend";

const productBackend = new ProductBackend(BackendSingleton);

export function useAllProducts() {
  const {
    data: products,
    setData: setProducts,
    ...rest
  } = useFetch(() => productBackend.getAllProducts(), [] as Product[]);
  return { ...rest, products, setProducts };
}
```

### Zod-Schema

```typescript
import { z } from "zod";

export const ProductIdSchema = z.number().int().min(1);

const NameSchema = z
  .string()
  .min(3, { message: "Das sieht nicht nach einem echten Namen aus." })
  .max(50, { message: "Der Name ist zu lang." });

const PreisCentsSchema = z
  .number()
  .int()
  .min(0, { message: "Preis muss mindestens 0 Cent sein." });

export const ProductSchema = z.object({
  id: ProductIdSchema,
  name: NameSchema,
  category: z.enum(["food", "beverage", "other"]),
  variants: z.array(VariantSchema),
  createdAt: DateStringSchema,
});
export type Product = z.infer<typeof ProductSchema>;
```
