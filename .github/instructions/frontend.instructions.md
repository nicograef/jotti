---
description: "Use when working on React frontend code, components, pages, hooks, styling, or TypeScript types."
applyTo: "frontend/**"
---

> **Referenz:** Für Ubiquitous Language, Namenskonventionen und Ist/Soll-Abweichungen (Rename-Status) → `docs/language.md`. Für Frontend-Architektur → `docs/handbuch.md` §6.3.

Repo-weite Regeln und Guardrails stehen kanonisch in `AGENTS.md`. Diese Datei ergänzt nur frontend-spezifische Konventionen für `frontend/**`.

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
  src/lib/                      # Auth, Backend-Client, Utilities
  src/admin/                    # Admin-Bereich (Produkte, Tische, Benutzer, Kasse/Kassensitzung)
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

- **Namenskonventionen (UI-Labels deutsch, Code-Mappings, Ist vs. Soll):** [docs/language.md](../../docs/language.md)
- **Architektur, Frontend-Patterns, Validierung:** [docs/handbuch.md](../../docs/handbuch.md) Kap. 6

## Code-Beispiele

### Backend-Client-Klasse

Pattern: Zod-Schema für Request definieren → `BackendClient.post()` aufrufen → Response mit Zod validieren.

```typescript
import { z } from "zod";
import type { BackendClient } from "@/lib/Backend";
import { type Produkt, ProduktIdSchema, ProduktSchema } from "./Produkt";

export const CreateProduktSchema = ProduktSchema.pick({
  name: true,
  kategorie: true,
});

export class ProduktBackend {
  constructor(private readonly backend: BackendClient) {}

  async createProdukt(
    newProdukt: z.infer<typeof CreateProduktSchema>,
  ): Promise<number> {
    const body = CreateProduktSchema.parse(newProdukt);
    const { id } = await this.backend.post(
      "admin/create-produkt",
      body,
      z.object({ id: ProduktIdSchema }),
    );
    return id;
  }

  async getAllProdukte(): Promise<Produkt[]> {
    const { produkte } = await this.backend.post(
      "admin/get-all-produkte",
      {},
      z.object({ produkte: z.array(ProduktSchema) }),
    );
    return produkte;
  }
}
```

### Custom Hook mit react-query

Datenbeschaffung läuft über `@tanstack/react-query` (`QueryClient` in `src/main.tsx`).
Lesezugriffe nutzen `useQuery`, Schreibzugriffe `useMutation` bzw.
`queryClient.invalidateQueries(...)`. Query-Keys als Konstante exportieren, damit
Mutationen gezielt invalidieren können. Das Lade-Flag heißt einheitlich `isPending`.

```typescript
import { useQuery } from "@tanstack/react-query";
import { BackendSingleton } from "@/lib/Backend";
import type { Produkt } from "./Produkt";
import { ProduktBackend } from "./ProduktBackend";

const produktBackend = new ProduktBackend(BackendSingleton);

export const ALLE_PRODUKTE_KEY = "alle-produkte";

export function useAllProdukte() {
  const { data: produkte = [] as Produkt[], isPending } = useQuery({
    queryKey: [ALLE_PRODUKTE_KEY],
    queryFn: () => produktBackend.getAllProdukte(),
  });
  return { produkte, isPending };
}
```

Schreibzugriffe invalidieren den betroffenen Query-Key, damit Reads neu laden:

```typescript
import { useMutation, useQueryClient } from "@tanstack/react-query";

const queryClient = useQueryClient();

const loeschenMutation = useMutation({
  mutationFn: (produktId: number) => produktBackend.deleteProdukt(produktId),
  onSuccess: () =>
    queryClient.invalidateQueries({ queryKey: [ALLE_PRODUKTE_KEY] }),
  onError: () => toast.error("Produkt konnte nicht gelöscht werden."),
});
```

### Zod-Schema

```typescript
import { z } from "zod";

export const ProduktIdSchema = z.number().int().min(1);

const NameSchema = z
  .string()
  .min(3, { message: "Das sieht nicht nach einem echten Namen aus." })
  .max(100, { message: "Der Name ist zu lang." });

const PreisCentsSchema = z
  .number()
  .int()
  .min(0, { message: "Preis muss mindestens 0 Cent sein." });

export const ProduktSchema = z.object({
  id: ProduktIdSchema,
  name: NameSchema,
  kategorie: z.enum(["essen", "getraenk", "sonstiges"]),
  varianten: z.array(VarianteSchema),
  createdAt: DateStringSchema,
});
export type Produkt = z.infer<typeof ProduktSchema>;
```
